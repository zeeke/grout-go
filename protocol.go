package grout

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Module IDs from grout C headers.
const (
	moduleMain     = 0xcafe
	moduleInfra    = 0xacdc
	moduleIPv4     = 0xf00d
	moduleIPv6     = 0xfeed
	moduleL2       = 0xbabe
	moduleDHCP     = 0xd4c9
	moduleConntrack = 0xc0c0
	moduleNAT      = 0x0bad
	moduleSRv6     = 0xfeef
)

// msgType encodes a module ID and message ID into a 32-bit message type.
func msgType(moduleID, msgID uint32) uint32 {
	return (moduleID << 16) | msgID
}

// Message types: Main module.
const (
	msgTypeHello            = 0xcafe<<16 | 0x1981
	msgTypeLogPacketsSet    = 0xcafe<<16 | 0x0001
	msgTypeLogLevelList     = 0xcafe<<16 | 0x0002
	msgTypeLogLevelSet      = 0xcafe<<16 | 0x0003
	msgTypeEventSubscribe   = 0xcafe<<16 | 0x0004
	msgTypeEventUnsubscribe = 0xcafe<<16 | 0x0005
)

// Message types: Infra module.
const (
	msgTypeIfaceAdd       = 0xacdc<<16 | 0x0001
	msgTypeIfaceDel       = 0xacdc<<16 | 0x0002
	msgTypeIfaceGet       = 0xacdc<<16 | 0x0003
	msgTypeIfaceList      = 0xacdc<<16 | 0x0004
	msgTypeIfaceSet       = 0xacdc<<16 | 0x0005
	msgTypeIfaceStatsGet  = 0xacdc<<16 | 0x0006
	msgTypeIfaceMACAdd    = 0xacdc<<16 | 0x0007
	msgTypeIfaceMACDel    = 0xacdc<<16 | 0x0008
	msgTypeIfaceMACList   = 0xacdc<<16 | 0x0009
	msgTypeIfaceMACSet    = 0xacdc<<16 | 0x000a
	msgTypeAffinityRxQList = 0xacdc<<16 | 0x0010
	msgTypeAffinityRxQSet  = 0xacdc<<16 | 0x0011
	msgTypeAffinityCPUGet  = 0xacdc<<16 | 0x0012
	msgTypeAffinityCPUSet  = 0xacdc<<16 | 0x0013
	msgTypeStatsGet       = 0xacdc<<16 | 0x0020
	msgTypeStatsReset     = 0xacdc<<16 | 0x0021
	msgTypeGraphDump      = 0xacdc<<16 | 0x0022
	msgTypeGraphConfGet   = 0xacdc<<16 | 0x0023
	msgTypeGraphConfSet   = 0xacdc<<16 | 0x0024
	msgTypePktTraceSet    = 0xacdc<<16 | 0x0030
	msgTypePktTraceClear  = 0xacdc<<16 | 0x0031
	msgTypePktTraceDump   = 0xacdc<<16 | 0x0032
	msgTypeNHAdd          = 0xacdc<<16 | 0x0040
	msgTypeNHDel          = 0xacdc<<16 | 0x0041
	msgTypeNHGet          = 0xacdc<<16 | 0x0042
	msgTypeNHList         = 0xacdc<<16 | 0x0043
	msgTypeNHConfigGet    = 0xacdc<<16 | 0x0044
	msgTypeNHConfigSet    = 0xacdc<<16 | 0x0045
)

// Message types: IPv4 module.
const (
	msgTypeIP4RouteAdd     = 0xf00d<<16 | 0x0001
	msgTypeIP4RouteDel     = 0xf00d<<16 | 0x0002
	msgTypeIP4RouteGet     = 0xf00d<<16 | 0x0003
	msgTypeIP4RouteList    = 0xf00d<<16 | 0x0004
	msgTypeIP4AddrAdd      = 0xf00d<<16 | 0x0010
	msgTypeIP4AddrDel      = 0xf00d<<16 | 0x0011
	msgTypeIP4AddrList     = 0xf00d<<16 | 0x0012
	msgTypeIP4AddrFlush    = 0xf00d<<16 | 0x0013
	msgTypeIP4ICMPSend     = 0xf00d<<16 | 0x0020
	msgTypeIP4ICMPRecv     = 0xf00d<<16 | 0x0021
	msgTypeIP4FIBDefaultSet = 0xf00d<<16 | 0x0030
	msgTypeIP4FIBInfoList  = 0xf00d<<16 | 0x0031
)

// Message types: IPv6 module.
const (
	msgTypeIP6RouteAdd      = 0xfeed<<16 | 0x0001
	msgTypeIP6RouteDel      = 0xfeed<<16 | 0x0002
	msgTypeIP6RouteGet      = 0xfeed<<16 | 0x0003
	msgTypeIP6RouteList     = 0xfeed<<16 | 0x0004
	msgTypeIP6AddrAdd       = 0xfeed<<16 | 0x0010
	msgTypeIP6AddrDel       = 0xfeed<<16 | 0x0011
	msgTypeIP6AddrList      = 0xfeed<<16 | 0x0012
	msgTypeIP6AddrFlush     = 0xfeed<<16 | 0x0013
	msgTypeIP6ICMPSend      = 0xfeed<<16 | 0x0020
	msgTypeIP6ICMPRecv      = 0xfeed<<16 | 0x0021
	msgTypeIP6FIBDefaultSet = 0xfeed<<16 | 0x0030
	msgTypeIP6FIBInfoList   = 0xfeed<<16 | 0x0031
	msgTypeIP6RASet         = 0xfeed<<16 | 0x0040
	msgTypeIP6RAClear       = 0xfeed<<16 | 0x0041
	msgTypeIP6RAShow        = 0xfeed<<16 | 0x0042
)

// Message types: L2 module.
const (
	msgTypeFDBAdd       = 0xbabe<<16 | 0x0001
	msgTypeFDBDel       = 0xbabe<<16 | 0x0002
	msgTypeFDBFlush     = 0xbabe<<16 | 0x0003
	msgTypeFDBList      = 0xbabe<<16 | 0x0004
	msgTypeFDBConfigGet = 0xbabe<<16 | 0x0005
	msgTypeFDBConfigSet = 0xbabe<<16 | 0x0006
	msgTypeFloodAdd     = 0xbabe<<16 | 0x0010
	msgTypeFloodDel     = 0xbabe<<16 | 0x0011
	msgTypeFloodList    = 0xbabe<<16 | 0x0012
)

// Message types: DHCP module.
const (
	msgTypeDHCPList  = 0xd4c9<<16 | 0x0001
	msgTypeDHCPStart = 0xd4c9<<16 | 0x0002
	msgTypeDHCPStop  = 0xd4c9<<16 | 0x0003
)

// Message types: Conntrack module.
const (
	msgTypeConntrackList    = 0xc0c0<<16 | 0x0001
	msgTypeConntrackFlush   = 0xc0c0<<16 | 0x0002
	msgTypeConntrackConfGet = 0xc0c0<<16 | 0x0003
	msgTypeConntrackConfSet = 0xc0c0<<16 | 0x0004
)

// Message types: NAT module.
const (
	msgTypeDNAT44Add  = 0x0bad<<16 | 0x0001
	msgTypeDNAT44Del  = 0x0bad<<16 | 0x0002
	msgTypeDNAT44List = 0x0bad<<16 | 0x0003
	msgTypeSNAT44Add  = 0x0bad<<16 | 0x0004
	msgTypeSNAT44Del  = 0x0bad<<16 | 0x0005
	msgTypeSNAT44List = 0x0bad<<16 | 0x0006
)

// Message types: SRv6 module.
const (
	msgTypeSRv6TunSrcSet   = 0xfeef<<16 | 0x0001
	msgTypeSRv6TunSrcClear = 0xfeef<<16 | 0x0002
	msgTypeSRv6TunSrcShow  = 0xfeef<<16 | 0x0003
)

// Header sizes in bytes.
const (
	requestHeaderSize  = 12
	responseHeaderSize = 12
	eventHeaderSize    = 12 // ev_type(4) + payload_len(8)
)

// Hello request struct size: api_version(4) + version[128](128) = 132.
const helloReqSize = 132

type helloReq struct {
	APIVersion uint32
	Version    [128]byte
}

// Event types.
const (
	EventAll uint32 = 0xffffffff
)

// Well-known IDs.
const (
	VRFDefaultID uint16 = 1
	IfaceIDUndef uint16 = 0
)

// request sends a single request and waits for its response.
// Returns the response payload. Returns an error if the response status is non-zero.
func (c *Client) request(msgT uint32, payload []byte) ([]byte, error) {
	if len(payload) > MaxPayloadLen {
		return nil, ErrPayloadTooLarge
	}

	id := c.nextID.Add(1)

	var hdr [requestHeaderSize]byte
	binary.LittleEndian.PutUint32(hdr[0:4], id)
	binary.LittleEndian.PutUint32(hdr[4:8], msgT)
	binary.LittleEndian.PutUint32(hdr[8:12], uint32(len(payload)))

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := c.conn.Write(hdr[:]); err != nil {
		return nil, fmt.Errorf("%w: write header: %v", ErrConnection, err)
	}
	if len(payload) > 0 {
		if _, err := c.conn.Write(payload); err != nil {
			return nil, fmt.Errorf("%w: write payload: %v", ErrConnection, err)
		}
	}

	return c.recvLocked(id)
}

// requestNoPayload sends a request with no payload body.
func (c *Client) requestNoPayload(msgT uint32) ([]byte, error) {
	return c.request(msgT, nil)
}

// recvLocked reads responses until one matching the given ID is found.
// Must be called with c.mu held.
func (c *Client) recvLocked(id uint32) ([]byte, error) {
	// Check for a previously buffered out-of-order response.
	if r, ok := c.pending[id]; ok {
		delete(c.pending, id)
		if err := statusToError(r.status, ""); err != nil {
			return nil, err
		}
		return r.payload, nil
	}

	for {
		forID, status, payload, err := c.readResponse()
		if err != nil {
			return nil, err
		}

		if forID == id {
			if err := statusToError(status, ""); err != nil {
				return nil, err
			}
			return payload, nil
		}

		// Out-of-order response — buffer it.
		c.pending[forID] = &response{
			status:     status,
			payloadLen: uint32(len(payload)),
			payload:    payload,
		}
	}
}

// readResponse reads a single response header and payload from the connection.
func (c *Client) readResponse() (forID, status uint32, payload []byte, err error) {
	var hdr [responseHeaderSize]byte
	if _, err = io.ReadFull(c.conn, hdr[:]); err != nil {
		err = fmt.Errorf("%w: read response header: %v", ErrConnection, err)
		return
	}

	forID = binary.LittleEndian.Uint32(hdr[0:4])
	status = binary.LittleEndian.Uint32(hdr[4:8])
	payloadLen := binary.LittleEndian.Uint32(hdr[8:12])

	if payloadLen > MaxPayloadLen {
		err = fmt.Errorf("%w: response payload %d bytes", ErrPayloadTooLarge, payloadLen)
		return
	}

	if payloadLen > 0 {
		payload = make([]byte, payloadLen)
		if _, err = io.ReadFull(c.conn, payload); err != nil {
			err = fmt.Errorf("%w: read response payload: %v", ErrConnection, err)
			return
		}
	}

	return
}

// requestStream sends a request and returns all responses from a streaming reply.
// The stream ends when a zero-length response is received.
func (c *Client) requestStream(msgT uint32, payload []byte) ([][]byte, error) {
	if len(payload) > MaxPayloadLen {
		return nil, ErrPayloadTooLarge
	}

	id := c.nextID.Add(1)

	var hdr [requestHeaderSize]byte
	binary.LittleEndian.PutUint32(hdr[0:4], id)
	binary.LittleEndian.PutUint32(hdr[4:8], msgT)
	binary.LittleEndian.PutUint32(hdr[8:12], uint32(len(payload)))

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := c.conn.Write(hdr[:]); err != nil {
		return nil, fmt.Errorf("%w: write header: %v", ErrConnection, err)
	}
	if len(payload) > 0 {
		if _, err := c.conn.Write(payload); err != nil {
			return nil, fmt.Errorf("%w: write payload: %v", ErrConnection, err)
		}
	}

	return c.recvStreamLocked(id)
}

// requestStreamNoPayload sends a streaming request with no payload.
func (c *Client) requestStreamNoPayload(msgT uint32) ([][]byte, error) {
	return c.requestStream(msgT, nil)
}

// recvStreamLocked reads streaming responses until a zero-length terminator.
// Must be called with c.mu held.
func (c *Client) recvStreamLocked(id uint32) ([][]byte, error) {
	var results [][]byte

	for {
		forID, status, payload, err := c.readResponse()
		if err != nil {
			return results, err
		}

		if forID != id {
			// Out-of-order response — buffer it.
			c.pending[forID] = &response{
				status:     status,
				payloadLen: uint32(len(payload)),
				payload:    payload,
			}
			continue
		}

		if err := statusToError(status, ""); err != nil {
			return nil, err
		}

		// Zero-length payload = end of stream.
		if len(payload) == 0 {
			return results, nil
		}

		results = append(results, payload)
	}
}

// eventHeader represents a received event.
type eventHeader struct {
	EvType     uint32
	PayloadLen uint64
}

// readEvent reads a single event from the connection.
// Events are sent asynchronously after subscription.
func (c *Client) readEvent() (eventHeader, []byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var hdr [eventHeaderSize]byte
	if _, err := io.ReadFull(c.conn, hdr[:]); err != nil {
		return eventHeader{}, nil, fmt.Errorf("%w: read event header: %v", ErrConnection, err)
	}

	ev := eventHeader{
		EvType:     binary.LittleEndian.Uint32(hdr[0:4]),
		PayloadLen: binary.LittleEndian.Uint64(hdr[4:12]),
	}

	if ev.PayloadLen > MaxPayloadLen {
		return ev, nil, fmt.Errorf("%w: event payload %d bytes", ErrPayloadTooLarge, ev.PayloadLen)
	}

	var payload []byte
	if ev.PayloadLen > 0 {
		payload = make([]byte, ev.PayloadLen)
		if _, err := io.ReadFull(c.conn, payload); err != nil {
			return ev, nil, fmt.Errorf("%w: read event payload: %v", ErrConnection, err)
		}
	}

	return ev, payload, nil
}
