package grout

import (
	"encoding/binary"
	"testing"
)

func TestConntrackEntry_Decode(t *testing.T) {
	data := make([]byte, conntrackEntrySize)
	binary.LittleEndian.PutUint16(data[0:2], 3) // iface_id
	data[2] = byte(AddrFamilyIPv4)               // family
	data[3] = 6                                   // proto (TCP)

	// Forward flow.
	copy(data[4:8], []byte{10, 0, 0, 1})    // src
	copy(data[8:12], []byte{10, 0, 0, 2})   // dst
	binary.LittleEndian.PutUint16(data[12:14], 12345) // src_id
	binary.LittleEndian.PutUint16(data[14:16], 80)    // dst_id

	// Reverse flow.
	copy(data[16:20], []byte{10, 0, 0, 2})
	copy(data[20:24], []byte{10, 0, 0, 1})
	binary.LittleEndian.PutUint16(data[24:26], 80)
	binary.LittleEndian.PutUint16(data[26:28], 12345)

	binary.LittleEndian.PutUint64(data[28:36], 1000000) // last_update
	binary.LittleEndian.PutUint32(data[36:40], 42)      // id
	data[40] = byte(ConnStateEstablished)

	e, err := decodeConntrackEntry(data)
	if err != nil {
		t.Fatalf("decodeConntrackEntry error: %v", err)
	}
	if e.IfaceID != 3 {
		t.Errorf("IfaceID = %d, want 3", e.IfaceID)
	}
	if e.Proto != 6 {
		t.Errorf("Proto = %d, want 6", e.Proto)
	}
	if e.State != ConnStateEstablished {
		t.Errorf("State = %d, want Established", e.State)
	}
	if e.FwdFlow.Src.String() != "10.0.0.1" {
		t.Errorf("FwdFlow.Src = %s, want 10.0.0.1", e.FwdFlow.Src)
	}
	if e.FwdFlow.DstID != 80 {
		t.Errorf("FwdFlow.DstID = %d, want 80", e.FwdFlow.DstID)
	}
}

func TestConntrackConfig_RoundTrip(t *testing.T) {
	cfg := ConntrackConfig{
		MaxCount:              65536,
		TimeoutClosed:         10,
		TimeoutNew:            30,
		TimeoutUDPEstablished: 180,
		TimeoutTCPEstablished: 7200,
		TimeoutHalfClose:      120,
		TimeoutTimeWait:       60,
	}

	encoded := encodeConntrackConfig(cfg)
	decoded, err := decodeConntrackConfig(encoded)
	if err != nil {
		t.Fatalf("decodeConntrackConfig error: %v", err)
	}
	if decoded != cfg {
		t.Errorf("round-trip failed: got %+v", decoded)
	}
}
