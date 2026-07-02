// Package grout provides a Go client library for the grout DPDK router daemon.
//
// Connect to a grout instance and manage network interfaces, IP addresses,
// routes, NAT policies, and more. The library implements the grout binary
// wire protocol over UNIX domain sockets.
//
//	client, err := grout.Connect()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	id, err := client.InterfaceAdd(grout.InterfaceAddRequest{
//	    Type: grout.IfaceTypePort,
//	    Name: "p0",
//	})
package grout

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

const (
	DefaultSocketPath = "/run/grout.sock"
	MaxPayloadLen     = 128 * 1024
	CurrentAPIVersion = 3
)

// Client represents a connection to a grout daemon.
type Client struct {
	conn       net.Conn
	mu         sync.Mutex
	nextID     atomic.Uint32
	apiVersion uint32
	serverVer  string
	closed     bool

	// pending holds out-of-order responses keyed by request ID.
	pending map[uint32]*response
}

type response struct {
	status     uint32
	payloadLen uint32
	payload    []byte
}

type clientConfig struct {
	socketPath string
}

// Option configures a Client connection.
type Option func(*clientConfig)

// WithSocketPath sets the UNIX socket path.
// Defaults to /run/grout.sock if not specified.
func WithSocketPath(path string) Option {
	return func(c *clientConfig) {
		c.socketPath = path
	}
}

// Connect establishes a connection to a grout daemon and performs
// the version handshake. Without options, connects to the default
// socket at /run/grout.sock.
// Returns ErrConnection if the daemon is unreachable, or
// ErrVersionMismatch if API versions are incompatible.
func Connect(opts ...Option) (*Client, error) {
	cfg := &clientConfig{
		socketPath: DefaultSocketPath,
	}
	for _, o := range opts {
		o(cfg)
	}

	conn, err := net.Dial("unix", cfg.socketPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnection, err)
	}

	c := &Client{
		conn:    conn,
		pending: make(map[uint32]*response),
	}

	if err := c.hello(); err != nil {
		conn.Close()
		return nil, err
	}

	return c, nil
}

func (c *Client) hello() error {
	var req helloReq
	req.APIVersion = CurrentAPIVersion
	copy(req.Version[:], "grout-go")

	payload := make([]byte, helloReqSize)
	binary.LittleEndian.PutUint32(payload[0:4], req.APIVersion)
	copy(payload[4:], req.Version[:])

	resp, err := c.request(msgTypeHello, payload)
	if err != nil {
		return fmt.Errorf("%w: hello handshake failed: %v", ErrConnection, err)
	}

	if len(resp) < helloReqSize {
		return fmt.Errorf("%w: hello response too short (%d bytes)", ErrConnection, len(resp))
	}

	c.apiVersion = binary.LittleEndian.Uint32(resp[0:4])
	verBytes := resp[4:helloReqSize]
	c.serverVer = cstring(verBytes)

	return nil
}

// Close closes the connection and releases resources.
// Safe to call multiple times.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	return c.conn.Close()
}

// APIVersion returns the API version negotiated during handshake.
func (c *Client) APIVersion() uint32 {
	return c.apiVersion
}

// ServerVersion returns the grout daemon version string from handshake.
func (c *Client) ServerVersion() string {
	return c.serverVer
}

// cstring extracts a null-terminated string from a byte slice.
func cstring(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
