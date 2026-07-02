package grout

import (
	"encoding/binary"
	"net"
	"testing"
)

func TestMsgType(t *testing.T) {
	tests := []struct {
		name     string
		module   uint32
		msg      uint32
		expected uint32
	}{
		{"hello", moduleMain, 0x1981, msgTypeHello},
		{"iface_add", moduleInfra, 0x0001, msgTypeIfaceAdd},
		{"ip4_route_add", moduleIPv4, 0x0001, msgTypeIP4RouteAdd},
		{"ip6_route_add", moduleIPv6, 0x0001, msgTypeIP6RouteAdd},
		{"fdb_add", moduleL2, 0x0001, msgTypeFDBAdd},
		{"dhcp_list", moduleDHCP, 0x0001, msgTypeDHCPList},
		{"conntrack_list", moduleConntrack, 0x0001, msgTypeConntrackList},
		{"dnat44_add", moduleNAT, 0x0001, msgTypeDNAT44Add},
		{"srv6_tunsrc_set", moduleSRv6, 0x0001, msgTypeSRv6TunSrcSet},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := msgType(tt.module, tt.msg)
			if got != tt.expected {
				t.Errorf("msgType(0x%x, 0x%x) = 0x%x, want 0x%x",
					tt.module, tt.msg, got, tt.expected)
			}
		})
	}
}

func TestHeaderSizes(t *testing.T) {
	if requestHeaderSize != 12 {
		t.Errorf("requestHeaderSize = %d, want 12", requestHeaderSize)
	}
	if responseHeaderSize != 12 {
		t.Errorf("responseHeaderSize = %d, want 12", responseHeaderSize)
	}
	if eventHeaderSize != 12 {
		t.Errorf("eventHeaderSize = %d, want 12", eventHeaderSize)
	}
	if helloReqSize != 132 {
		t.Errorf("helloReqSize = %d, want 132", helloReqSize)
	}
}

func TestRequestResponse_RoundTrip(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	c := &Client{
		conn:    client,
		pending: make(map[uint32]*response),
	}

	expectedPayload := []byte("hello world")
	go func() {
		// Read the request.
		var hdr [requestHeaderSize]byte
		if _, err := readFull(server, hdr[:]); err != nil {
			t.Errorf("read request header: %v", err)
			return
		}
		id := binary.LittleEndian.Uint32(hdr[0:4])
		msgT := binary.LittleEndian.Uint32(hdr[4:8])
		pLen := binary.LittleEndian.Uint32(hdr[8:12])

		if msgT != msgTypeIfaceList {
			t.Errorf("request type = 0x%x, want 0x%x", msgT, msgTypeIfaceList)
		}
		if pLen != 0 {
			t.Errorf("payload len = %d, want 0", pLen)
		}

		// Write the response.
		writeResponse(server, id, 0, expectedPayload)
	}()

	resp, err := c.request(msgTypeIfaceList, nil)
	if err != nil {
		t.Fatalf("request() error: %v", err)
	}
	if string(resp) != string(expectedPayload) {
		t.Errorf("response = %q, want %q", resp, expectedPayload)
	}
}

func TestRequestResponse_Error(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	c := &Client{
		conn:    client,
		pending: make(map[uint32]*response),
	}

	go func() {
		var hdr [requestHeaderSize]byte
		readFull(server, hdr[:])
		id := binary.LittleEndian.Uint32(hdr[0:4])
		// Return ENOENT (-2 as uint32).
		var status uint32 = 0xfffffffe // -2 in two's complement
		writeResponse(server, id, status, nil)
	}()

	_, err := c.request(msgTypeIfaceGet, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRequestStream(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	c := &Client{
		conn:    client,
		pending: make(map[uint32]*response),
	}

	items := []string{"item1", "item2", "item3"}
	go func() {
		var hdr [requestHeaderSize]byte
		readFull(server, hdr[:])
		id := binary.LittleEndian.Uint32(hdr[0:4])

		for _, item := range items {
			writeResponse(server, id, 0, []byte(item))
		}
		writeStreamEnd(server, id)
	}()

	results, err := c.requestStream(msgTypeIfaceList, nil)
	if err != nil {
		t.Fatalf("requestStream() error: %v", err)
	}
	if len(results) != len(items) {
		t.Fatalf("got %d results, want %d", len(results), len(items))
	}
	for i, r := range results {
		if string(r) != items[i] {
			t.Errorf("result[%d] = %q, want %q", i, r, items[i])
		}
	}
}

func TestOutOfOrderResponse(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	c := &Client{
		conn:    client,
		pending: make(map[uint32]*response),
	}

	go func() {
		// Read first request.
		var hdr1 [requestHeaderSize]byte
		readFull(server, hdr1[:])
		id1 := binary.LittleEndian.Uint32(hdr1[0:4])

		// Reply out of order: respond with a different ID first, then the correct one.
		writeResponse(server, id1+100, 0, []byte("wrong"))
		writeResponse(server, id1, 0, []byte("right"))
	}()

	resp, err := c.request(msgTypeIfaceGet, nil)
	if err != nil {
		t.Fatalf("request() error: %v", err)
	}
	if string(resp) != "right" {
		t.Errorf("response = %q, want %q", resp, "right")
	}

	// The out-of-order response should be buffered.
	if r, ok := c.pending[c.nextID.Load()+100]; !ok {
		t.Error("expected buffered out-of-order response")
	} else if string(r.payload) != "wrong" {
		t.Errorf("buffered payload = %q, want %q", r.payload, "wrong")
	}
}

func TestPayloadTooLarge(t *testing.T) {
	c := &Client{
		pending: make(map[uint32]*response),
	}
	bigPayload := make([]byte, MaxPayloadLen+1)

	_, err := c.request(msgTypeIfaceAdd, bigPayload)
	if err != ErrPayloadTooLarge {
		t.Errorf("expected ErrPayloadTooLarge, got: %v", err)
	}
}
