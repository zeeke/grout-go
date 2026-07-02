package grout

import (
	"encoding/binary"
	"testing"
)

func TestIP4IfAddr_RoundTrip(t *testing.T) {
	addr := IP4IfAddr{
		IfaceID: 5,
		Addr:    MustParseIP4Net("192.168.1.0/24"),
	}

	encoded := encodeIP4IfAddr(addr, false)
	decoded, err := decodeIP4IfAddr(encoded)
	if err != nil {
		t.Fatalf("decodeIP4IfAddr error: %v", err)
	}
	if decoded.IfaceID != addr.IfaceID {
		t.Errorf("IfaceID = %d, want %d", decoded.IfaceID, addr.IfaceID)
	}
	if decoded.Addr.PrefixLen != 24 {
		t.Errorf("PrefixLen = %d, want 24", decoded.Addr.PrefixLen)
	}
	if decoded.Addr.Addr != addr.Addr.Addr {
		t.Errorf("Addr = %v, want %v", decoded.Addr.Addr, addr.Addr.Addr)
	}
}

func TestIP4IfAddr_WithFlag(t *testing.T) {
	addr := IP4IfAddr{IfaceID: 1, Addr: MustParseIP4Net("10.0.0.0/8")}
	encoded := encodeIP4IfAddr(addr, true)
	if encoded[2] != 1 {
		t.Errorf("flag byte = %d, want 1", encoded[2])
	}
}

func TestIP4RouteAdd_Encode(t *testing.T) {
	req := IP4RouteAddRequest{
		VRFID:  VRFDefaultID,
		Dest:   MustParseIP4Net("0.0.0.0/0"),
		NH:     MustParseIP4("172.16.1.183"),
		Origin: NHOriginStatic,
	}

	encoded := encodeIP4RouteAdd(req)
	vrfID := binary.LittleEndian.Uint16(encoded[0:2])
	if vrfID != VRFDefaultID {
		t.Errorf("VRFID = %d, want %d", vrfID, VRFDefaultID)
	}
	if encoded[2] != byte(NHOriginStatic) {
		t.Errorf("Origin = %d, want Static", encoded[2])
	}
	if encoded[3] != 0 {
		t.Errorf("PrefixLen = %d, want 0", encoded[3])
	}
}

func TestIP4Route_Decode(t *testing.T) {
	data := make([]byte, 8+nexthopBaseSize)
	binary.LittleEndian.PutUint16(data[0:2], 1) // vrf_id
	data[2] = byte(NHOriginStatic)
	data[3] = 24 // prefix_len
	copy(data[4:8], []byte{10, 0, 0, 0})

	// Nexthop.
	nh := data[8:]
	nh[0] = byte(NHTypeBlackhole)

	route, err := decodeIP4Route(data)
	if err != nil {
		t.Fatalf("decodeIP4Route error: %v", err)
	}
	if route.VRFID != 1 {
		t.Errorf("VRFID = %d, want 1", route.VRFID)
	}
	if route.Dest.Addr.String() != "10.0.0.0" {
		t.Errorf("Dest = %s, want 10.0.0.0", route.Dest.Addr)
	}
	if route.NH.Type != NHTypeBlackhole {
		t.Errorf("NH.Type = %d, want Blackhole", route.NH.Type)
	}
}

func TestFIB4Info_Decode(t *testing.T) {
	data := make([]byte, fib4InfoSize)
	binary.LittleEndian.PutUint16(data[0:2], 1)
	binary.LittleEndian.PutUint32(data[4:8], 65536)
	binary.LittleEndian.PutUint32(data[8:12], 100)

	info, err := decodeFIB4Info(data)
	if err != nil {
		t.Fatalf("decodeFIB4Info error: %v", err)
	}
	if info.VRFID != 1 {
		t.Errorf("VRFID = %d, want 1", info.VRFID)
	}
	if info.MaxRoutes != 65536 {
		t.Errorf("MaxRoutes = %d, want 65536", info.MaxRoutes)
	}
}

func TestICMPRecvResp_Decode(t *testing.T) {
	data := make([]byte, 20)
	data[0] = 0  // type (echo reply)
	data[1] = 0  // code
	data[2] = 64 // ttl
	binary.LittleEndian.PutUint16(data[4:6], 1234) // ident
	binary.LittleEndian.PutUint16(data[6:8], 1)    // seq
	copy(data[8:12], []byte{10, 0, 0, 1})
	binary.LittleEndian.PutUint64(data[12:20], 500000) // 500us

	resp, err := decodeICMPRecvResp(data)
	if err != nil {
		t.Fatalf("decodeICMPRecvResp error: %v", err)
	}
	if resp.TTL != 64 {
		t.Errorf("TTL = %d, want 64", resp.TTL)
	}
	if resp.SeqNum != 1 {
		t.Errorf("SeqNum = %d, want 1", resp.SeqNum)
	}
	if resp.SrcAddr.String() != "10.0.0.1" {
		t.Errorf("SrcAddr = %s, want 10.0.0.1", resp.SrcAddr)
	}
	if resp.ResponseTime != 500000 {
		t.Errorf("ResponseTime = %d, want 500000", resp.ResponseTime)
	}
}
