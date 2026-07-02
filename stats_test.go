package grout

import (
	"encoding/binary"
	"testing"
)

func TestStat_Decode(t *testing.T) {
	data := make([]byte, statSize)
	copy(data[0:statNameLen], "eth_rx")
	binary.LittleEndian.PutUint64(data[statNameLen:statNameLen+8], 5)
	binary.LittleEndian.PutUint64(data[statNameLen+8:statNameLen+16], 1000)
	binary.LittleEndian.PutUint64(data[statNameLen+16:statNameLen+24], 100)
	binary.LittleEndian.PutUint64(data[statNameLen+24:statNameLen+32], 500000)

	s, err := decodeStat(data)
	if err != nil {
		t.Fatalf("decodeStat error: %v", err)
	}
	if s.Name != "eth_rx" {
		t.Errorf("Name = %q, want %q", s.Name, "eth_rx")
	}
	if s.TopoOrder != 5 {
		t.Errorf("TopoOrder = %d, want 5", s.TopoOrder)
	}
	if s.Packets != 1000 {
		t.Errorf("Packets = %d, want 1000", s.Packets)
	}
	if s.Batches != 100 {
		t.Errorf("Batches = %d, want 100", s.Batches)
	}
	if s.Cycles != 500000 {
		t.Errorf("Cycles = %d, want 500000", s.Cycles)
	}
}

func TestStat_TooShort(t *testing.T) {
	data := make([]byte, statSize-1)
	_, err := decodeStat(data)
	if err == nil {
		t.Fatal("expected error for short data")
	}
}

func TestGraphConf_RoundTrip(t *testing.T) {
	conf := GraphConf{
		RxBurstMax: 256,
		VectorMax:  32,
	}

	buf := make([]byte, graphConfSize)
	binary.LittleEndian.PutUint16(buf[0:2], conf.RxBurstMax)
	binary.LittleEndian.PutUint16(buf[2:4], conf.VectorMax)

	got := GraphConf{
		RxBurstMax: binary.LittleEndian.Uint16(buf[0:2]),
		VectorMax:  binary.LittleEndian.Uint16(buf[2:4]),
	}
	if got != conf {
		t.Errorf("round-trip failed: got %+v, want %+v", got, conf)
	}
}
