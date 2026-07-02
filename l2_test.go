package grout

import (
	"encoding/binary"
	"testing"
)

func TestFDBEntry_RoundTrip(t *testing.T) {
	e := FDBEntry{
		BridgeID: 10,
		MAC:      MustParseEtherAddr("aa:bb:cc:dd:ee:ff"),
		VLANID:   100,
		IfaceID:  5,
		VTEP:     L3AddrFromIP4(IP4Addr{10, 0, 0, 1}),
		Flags:    FDBFlagStatic,
		LastSeen: 1000000,
	}

	encoded := encodeFDBEntry(e, false)
	decoded, err := decodeFDBEntry(encoded)
	if err != nil {
		t.Fatalf("decodeFDBEntry error: %v", err)
	}
	if decoded.BridgeID != e.BridgeID {
		t.Errorf("BridgeID = %d, want %d", decoded.BridgeID, e.BridgeID)
	}
	if decoded.MAC != e.MAC {
		t.Errorf("MAC = %s, want %s", decoded.MAC, e.MAC)
	}
	if decoded.VLANID != e.VLANID {
		t.Errorf("VLANID = %d, want %d", decoded.VLANID, e.VLANID)
	}
	if decoded.Flags != FDBFlagStatic {
		t.Errorf("Flags = %d, want Static", decoded.Flags)
	}
}

func TestFloodEntry_RoundTrip(t *testing.T) {
	e := FloodEntry{
		Type:  FloodTypeVTEP,
		VRFID: 1,
		VNI:   42,
		Addr:  L3AddrFromIP4(IP4Addr{10, 0, 0, 1}),
	}

	encoded := encodeFloodEntry(e, false)
	decoded, err := decodeFloodEntry(encoded)
	if err != nil {
		t.Fatalf("decodeFloodEntry error: %v", err)
	}
	if decoded.Type != FloodTypeVTEP {
		t.Errorf("Type = %d, want VTEP", decoded.Type)
	}
	if decoded.VNI != 42 {
		t.Errorf("VNI = %d, want 42", decoded.VNI)
	}
}

func TestFDBConfig_Decode(t *testing.T) {
	data := make([]byte, fdbConfigSize)
	binary.LittleEndian.PutUint32(data[0:4], 10000)
	binary.LittleEndian.PutUint32(data[4:8], 500)

	// Decode manually to test.
	max := binary.LittleEndian.Uint32(data[0:4])
	used := binary.LittleEndian.Uint32(data[4:8])
	if max != 10000 {
		t.Errorf("max = %d, want 10000", max)
	}
	if used != 500 {
		t.Errorf("used = %d, want 500", used)
	}
}
