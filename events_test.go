package grout

import (
	"encoding/binary"
	"testing"
)

func TestEventSubscribe_Encode(t *testing.T) {
	payload := make([]byte, 8)
	binary.LittleEndian.PutUint32(payload[0:4], 42)
	binary.LittleEndian.PutUint32(payload[4:8], 1)

	evType := binary.LittleEndian.Uint32(payload[0:4])
	if evType != 42 {
		t.Errorf("evType = %d, want 42", evType)
	}
	suppress := binary.LittleEndian.Uint32(payload[4:8])
	if suppress != 1 {
		t.Errorf("suppressSelf = %d, want 1", suppress)
	}
}

func TestEvent_Fields(t *testing.T) {
	ev := Event{
		EvType:     100,
		PayloadLen: 16,
		Payload:    []byte("test event data!"),
	}
	if ev.EvType != 100 {
		t.Errorf("EvType = %d, want 100", ev.EvType)
	}
	if ev.PayloadLen != 16 {
		t.Errorf("PayloadLen = %d, want 16", ev.PayloadLen)
	}
	if string(ev.Payload) != "test event data!" {
		t.Errorf("Payload = %q", ev.Payload)
	}
}
