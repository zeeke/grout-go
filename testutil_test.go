package grout

import (
	"encoding/binary"
	"net"
	"os"
	"path/filepath"
	"testing"
)

// mockServer is a test helper that listens on a UNIX socket and handles
// grout protocol messages via a user-supplied handler function.
type mockServer struct {
	listener net.Listener
	sockPath string
	handler  func(net.Conn)
	done     chan struct{}
}

// newMockServer creates a mock grout server listening on a temporary UNIX socket.
// The handler is called for each accepted connection.
func newMockServer(t *testing.T, handler func(net.Conn)) *mockServer {
	t.Helper()
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "grout.sock")

	l, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("mock server listen: %v", err)
	}

	s := &mockServer{
		listener: l,
		sockPath: sockPath,
		handler:  handler,
		done:     make(chan struct{}),
	}

	go s.serve()
	return s
}

func (s *mockServer) serve() {
	defer close(s.done)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handler(conn)
	}
}

func (s *mockServer) close() {
	s.listener.Close()
	<-s.done
	os.Remove(s.sockPath)
}

// readRequestHeader reads a 12-byte request header from the connection.
func readRequestHeader(conn net.Conn) (id, msgType, payloadLen uint32, err error) {
	var hdr [requestHeaderSize]byte
	if _, err = readFull(conn, hdr[:]); err != nil {
		return
	}
	id = binary.LittleEndian.Uint32(hdr[0:4])
	msgType = binary.LittleEndian.Uint32(hdr[4:8])
	payloadLen = binary.LittleEndian.Uint32(hdr[8:12])
	return
}

// readPayload reads the specified number of payload bytes from the connection.
func readPayload(conn net.Conn, length uint32) ([]byte, error) {
	if length == 0 {
		return nil, nil
	}
	buf := make([]byte, length)
	_, err := readFull(conn, buf)
	return buf, err
}

// writeResponse writes a response header and payload to the connection.
func writeResponse(conn net.Conn, forID, status uint32, payload []byte) error {
	var hdr [responseHeaderSize]byte
	binary.LittleEndian.PutUint32(hdr[0:4], forID)
	binary.LittleEndian.PutUint32(hdr[4:8], status)
	binary.LittleEndian.PutUint32(hdr[8:12], uint32(len(payload)))
	if _, err := conn.Write(hdr[:]); err != nil {
		return err
	}
	if len(payload) > 0 {
		if _, err := conn.Write(payload); err != nil {
			return err
		}
	}
	return nil
}

// writeStreamEnd writes a zero-length response to terminate a stream.
func writeStreamEnd(conn net.Conn, forID uint32) error {
	return writeResponse(conn, forID, 0, nil)
}

// readFull reads exactly len(buf) bytes from r.
func readFull(r net.Conn, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := r.Read(buf[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// helloHandler returns a mock handler that responds to GR_HELLO with a
// successful handshake using the given server version string.
func helloHandler(serverVersion string) func(net.Conn) {
	return func(conn net.Conn) {
		defer conn.Close()

		id, _, _, err := readRequestHeader(conn)
		if err != nil {
			return
		}

		// Read the hello payload.
		payload, err := readPayload(conn, helloReqSize)
		if err != nil {
			return
		}
		_ = payload

		// Build hello response.
		resp := make([]byte, helloReqSize)
		binary.LittleEndian.PutUint32(resp[0:4], CurrentAPIVersion)
		copy(resp[4:], serverVersion)

		writeResponse(conn, id, 0, resp)
	}
}
