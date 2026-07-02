package grout

import (
	"encoding/binary"
	"testing"
)

func TestIP6IfAddr_RoundTrip(t *testing.T) {
	addr := IP6IfAddr{
		IfaceID: 3,
		Addr:    MustParseIP6Net("fd00::1/64"),
	}

	encoded := encodeIP6IfAddr(addr, false)
	decoded, err := decodeIP6IfAddr(encoded)
	if err != nil {
		t.Fatalf("decodeIP6IfAddr error: %v", err)
	}
	if decoded.IfaceID != 3 {
		t.Errorf("IfaceID = %d, want 3", decoded.IfaceID)
	}
	if decoded.Addr.PrefixLen != 64 {
		t.Errorf("PrefixLen = %d, want 64", decoded.Addr.PrefixLen)
	}
}

func TestIP6Route_Decode(t *testing.T) {
	data := make([]byte, 20+nexthopBaseSize)
	binary.LittleEndian.PutUint16(data[0:2], 1) // vrf_id
	data[2] = byte(NHOriginStatic)
	data[3] = 64
	dest := MustParseIP6("fd00::")
	copy(data[4:20], dest[:])

	nh := data[20:]
	nh[0] = byte(NHTypeL3)

	route, err := decodeIP6Route(data)
	if err != nil {
		t.Fatalf("decodeIP6Route error: %v", err)
	}
	if route.VRFID != 1 {
		t.Errorf("VRFID = %d, want 1", route.VRFID)
	}
	if route.Dest.PrefixLen != 64 {
		t.Errorf("PrefixLen = %d, want 64", route.Dest.PrefixLen)
	}
}

func TestFIB6Info_Decode(t *testing.T) {
	data := make([]byte, fib6InfoSize)
	binary.LittleEndian.PutUint16(data[0:2], 1)
	binary.LittleEndian.PutUint32(data[4:8], 65536)
	binary.LittleEndian.PutUint32(data[8:12], 50)

	info, err := decodeFIB6Info(data)
	if err != nil {
		t.Fatalf("decodeFIB6Info error: %v", err)
	}
	if info.MaxRoutes != 65536 {
		t.Errorf("MaxRoutes = %d, want 65536", info.MaxRoutes)
	}
	if info.UsedRoutes != 50 {
		t.Errorf("UsedRoutes = %d, want 50", info.UsedRoutes)
	}
}

func TestRAConf_RoundTrip(t *testing.T) {
	conf := RAConf{
		Enabled:  true,
		IfaceID:  7,
		Interval: 600,
		Lifetime: 1800,
	}

	encoded := encodeRAConf(conf)
	decoded, err := decodeRAConf(encoded)
	if err != nil {
		t.Fatalf("decodeRAConf error: %v", err)
	}
	if decoded != conf {
		t.Errorf("round-trip failed: got %+v, want %+v", decoded, conf)
	}
}

func TestICMP6RecvResp_Decode(t *testing.T) {
	data := make([]byte, 32)
	data[0] = 129 // echo reply
	data[2] = 64  // ttl
	binary.LittleEndian.PutUint16(data[4:6], 5678)
	binary.LittleEndian.PutUint16(data[6:8], 1)
	src := MustParseIP6("fe80::1")
	copy(data[8:24], src[:])
	binary.LittleEndian.PutUint64(data[24:32], 1000000) // 1ms

	resp, err := decodeICMP6RecvResp(data)
	if err != nil {
		t.Fatalf("decodeICMP6RecvResp error: %v", err)
	}
	if resp.Type != 129 {
		t.Errorf("Type = %d, want 129", resp.Type)
	}
	if resp.TTL != 64 {
		t.Errorf("TTL = %d, want 64", resp.TTL)
	}
	if resp.ResponseTime != 1000000 {
		t.Errorf("ResponseTime = %d, want 1000000", resp.ResponseTime)
	}
}
