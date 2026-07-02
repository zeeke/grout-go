package grout

import (
	"encoding/binary"
	"testing"
)

func TestDecodeNexthop_L3(t *testing.T) {
	data := make([]byte, nexthopBaseSize+4+l3AddrWireSize+6)
	data[0] = byte(NHTypeL3)
	data[1] = byte(NHOriginStatic)
	binary.LittleEndian.PutUint16(data[2:4], 5)  // iface_id
	binary.LittleEndian.PutUint16(data[4:6], 1)  // vrf_id
	binary.LittleEndian.PutUint32(data[8:12], 42) // nh_id

	info := data[nexthopBaseSize:]
	info[0] = byte(NHStateReachable)
	info[1] = byte(NHFlagGateway)
	info[2] = byte(AddrFamilyIPv4)
	info[3] = 24

	encodeL3Addr(info[4:4+l3AddrWireSize], L3AddrFromIP4(IP4Addr{10, 0, 0, 1}))
	copy(info[4+l3AddrWireSize:], []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff})

	nh, err := decodeNexthop(data)
	if err != nil {
		t.Fatalf("decodeNexthop error: %v", err)
	}
	if nh.Type != NHTypeL3 {
		t.Errorf("Type = %d, want L3", nh.Type)
	}
	if nh.NHID != 42 {
		t.Errorf("NHID = %d, want 42", nh.NHID)
	}
	if nh.L3 == nil {
		t.Fatal("L3 info is nil")
	}
	if nh.L3.State != NHStateReachable {
		t.Errorf("State = %d, want Reachable", nh.L3.State)
	}
	if nh.L3.Addr.IP4.String() != "10.0.0.1" {
		t.Errorf("Addr = %s, want 10.0.0.1", nh.L3.Addr.IP4)
	}
}

func TestDecodeNexthop_Group(t *testing.T) {
	memberSize := 8
	data := make([]byte, nexthopBaseSize+3*memberSize)
	data[0] = byte(NHTypeGroup)
	binary.LittleEndian.PutUint32(data[8:12], 100)

	info := data[nexthopBaseSize:]
	for i := 0; i < 3; i++ {
		off := i * memberSize
		binary.LittleEndian.PutUint32(info[off:off+4], uint32(i+1))
		binary.LittleEndian.PutUint32(info[off+4:off+8], uint32(10*(i+1)))
	}

	nh, err := decodeNexthop(data)
	if err != nil {
		t.Fatalf("decodeNexthop error: %v", err)
	}
	if nh.Type != NHTypeGroup {
		t.Errorf("Type = %d, want Group", nh.Type)
	}
	if nh.Group == nil {
		t.Fatal("Group info is nil")
	}
	if len(nh.Group.Members) != 3 {
		t.Fatalf("got %d members, want 3", len(nh.Group.Members))
	}
	if nh.Group.Members[0].NHID != 1 || nh.Group.Members[0].Weight != 10 {
		t.Errorf("member[0] = %+v, want {1, 10}", nh.Group.Members[0])
	}
}

func TestDecodeNexthop_Blackhole(t *testing.T) {
	data := make([]byte, nexthopBaseSize)
	data[0] = byte(NHTypeBlackhole)
	binary.LittleEndian.PutUint32(data[8:12], 99)

	nh, err := decodeNexthop(data)
	if err != nil {
		t.Fatalf("decodeNexthop error: %v", err)
	}
	if nh.Type != NHTypeBlackhole {
		t.Errorf("Type = %d, want Blackhole", nh.Type)
	}
	if nh.L3 != nil || nh.Group != nil {
		t.Error("expected nil type-specific info for blackhole")
	}
}

func TestDecodeNexthop_TooShort(t *testing.T) {
	_, err := decodeNexthop(make([]byte, 5))
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestNHConfig_RoundTrip(t *testing.T) {
	cfg := NHConfig{
		MaxCount:            1000,
		LifetimeReachable:   30,
		LifetimeUnreachable: 60,
		MaxHeldPkts:         5,
		MaxUcastProbes:      3,
		MaxBcastProbes:      3,
	}

	encoded := encodeNHConfig(cfg)
	decoded, err := decodeNHConfig(encoded)
	if err != nil {
		t.Fatalf("decodeNHConfig error: %v", err)
	}
	if decoded != cfg {
		t.Errorf("round-trip failed: got %+v", decoded)
	}
}

func TestEncodeNexthopBase_L3(t *testing.T) {
	nh := Nexthop{
		Type:    NHTypeL3,
		Origin:  NHOriginStatic,
		IfaceID: 5,
		VRFID:   1,
		NHID:    42,
		L3: &NHInfoL3{
			State:  NHStateReachable,
			Flags:  NHFlagGateway,
			Family: AddrFamilyIPv4,
			Addr:   L3AddrFromIP4(IP4Addr{10, 0, 0, 1}),
		},
	}

	encoded := encodeNexthopBase(nh)
	decoded, err := decodeNexthop(encoded)
	if err != nil {
		t.Fatalf("decodeNexthop error: %v", err)
	}
	if decoded.Type != NHTypeL3 {
		t.Errorf("Type = %d, want L3", decoded.Type)
	}
	if decoded.NHID != 42 {
		t.Errorf("NHID = %d, want 42", decoded.NHID)
	}
}
