package grout

import (
	"encoding/binary"
	"testing"
)

func TestLogEntry_Decode(t *testing.T) {
	data := make([]byte, logEntrySize)
	copy(data[0:logNameLen], "grout.main")
	binary.LittleEndian.PutUint32(data[logNameLen:logNameLen+4], 7)

	e, err := decodeLogEntry(data)
	if err != nil {
		t.Fatalf("decodeLogEntry error: %v", err)
	}
	if e.Name != "grout.main" {
		t.Errorf("Name = %q, want %q", e.Name, "grout.main")
	}
	if e.Level != 7 {
		t.Errorf("Level = %d, want 7", e.Level)
	}
}

func TestLogEntry_TooShort(t *testing.T) {
	data := make([]byte, logEntrySize-1)
	_, err := decodeLogEntry(data)
	if err == nil {
		t.Fatal("expected error for short data")
	}
}
