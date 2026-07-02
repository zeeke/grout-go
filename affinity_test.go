package grout

import (
	"encoding/binary"
	"testing"
)

func TestRxQueueMap_Decode(t *testing.T) {
	data := make([]byte, rxQueueMapSize)
	binary.LittleEndian.PutUint16(data[0:2], 1)
	binary.LittleEndian.PutUint16(data[2:4], 0)
	binary.LittleEndian.PutUint16(data[4:6], 4)
	binary.LittleEndian.PutUint16(data[6:8], 1)

	m, err := decodeRxQueueMap(data)
	if err != nil {
		t.Fatalf("decodeRxQueueMap error: %v", err)
	}
	if m.IfaceID != 1 {
		t.Errorf("IfaceID = %d, want 1", m.IfaceID)
	}
	if m.CPUID != 4 {
		t.Errorf("CPUID = %d, want 4", m.CPUID)
	}
	if m.Enabled != 1 {
		t.Errorf("Enabled = %d, want 1", m.Enabled)
	}
}

func TestRxQueueMap_TooShort(t *testing.T) {
	data := make([]byte, rxQueueMapSize-1)
	_, err := decodeRxQueueMap(data)
	if err == nil {
		t.Fatal("expected error for short data")
	}
}

func TestCPUAffinity_RoundTrip(t *testing.T) {
	control := CPUSet{CPUs: []uint16{0, 1}}
	datapath := CPUSet{CPUs: []uint16{2, 3, 4, 5}}

	encoded := encodeCPUAffinity(control, datapath)
	decoded, err := decodeCPUAffinity(encoded)
	if err != nil {
		t.Fatalf("decodeCPUAffinity error: %v", err)
	}
	if len(decoded.Control.CPUs) != 2 {
		t.Fatalf("Control CPUs len = %d, want 2", len(decoded.Control.CPUs))
	}
	if decoded.Control.CPUs[0] != 0 || decoded.Control.CPUs[1] != 1 {
		t.Errorf("Control CPUs = %v, want [0 1]", decoded.Control.CPUs)
	}
	if len(decoded.Datapath.CPUs) != 4 {
		t.Fatalf("Datapath CPUs len = %d, want 4", len(decoded.Datapath.CPUs))
	}
	if decoded.Datapath.CPUs[3] != 5 {
		t.Errorf("Datapath CPUs[3] = %d, want 5", decoded.Datapath.CPUs[3])
	}
}

func TestCPUAffinity_Empty(t *testing.T) {
	control := CPUSet{CPUs: []uint16{}}
	datapath := CPUSet{CPUs: []uint16{}}

	encoded := encodeCPUAffinity(control, datapath)
	decoded, err := decodeCPUAffinity(encoded)
	if err != nil {
		t.Fatalf("decodeCPUAffinity error: %v", err)
	}
	if len(decoded.Control.CPUs) != 0 {
		t.Errorf("Control CPUs len = %d, want 0", len(decoded.Control.CPUs))
	}
	if len(decoded.Datapath.CPUs) != 0 {
		t.Errorf("Datapath CPUs len = %d, want 0", len(decoded.Datapath.CPUs))
	}
}
