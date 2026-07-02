package grout

import (
	"encoding/binary"
	"testing"
)

func TestPacketTraceSet_Encode(t *testing.T) {
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint16(payload[0:2], 3)
	payload[2] = 1
	payload[3] = 1

	ifaceID := binary.LittleEndian.Uint16(payload[0:2])
	if ifaceID != 3 {
		t.Errorf("ifaceID = %d, want 3", ifaceID)
	}
	if payload[2] != 1 {
		t.Errorf("enabled = %d, want 1", payload[2])
	}
	if payload[3] != 1 {
		t.Errorf("all = %d, want 1", payload[3])
	}
}
