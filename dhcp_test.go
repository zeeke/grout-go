package grout

import (
	"encoding/binary"
	"testing"
)

func TestDHCPStatus_Decode(t *testing.T) {
	data := make([]byte, dhcpStatusSize)
	binary.LittleEndian.PutUint16(data[0:2], 7)
	data[2] = byte(DHCPStateBound)
	copy(data[4:8], []byte{10, 0, 0, 1})
	copy(data[8:12], []byte{10, 0, 0, 100})
	binary.LittleEndian.PutUint32(data[12:16], 86400)
	binary.LittleEndian.PutUint32(data[16:20], 43200)
	binary.LittleEndian.PutUint32(data[20:24], 75600)

	s, err := decodeDHCPStatus(data)
	if err != nil {
		t.Fatalf("decodeDHCPStatus error: %v", err)
	}
	if s.IfaceID != 7 {
		t.Errorf("IfaceID = %d, want 7", s.IfaceID)
	}
	if s.State != DHCPStateBound {
		t.Errorf("State = %d, want Bound", s.State)
	}
	if s.ServerIP.String() != "10.0.0.1" {
		t.Errorf("ServerIP = %s, want 10.0.0.1", s.ServerIP)
	}
	if s.AssignedIP.String() != "10.0.0.100" {
		t.Errorf("AssignedIP = %s, want 10.0.0.100", s.AssignedIP)
	}
	if s.LeaseTime != 86400 {
		t.Errorf("LeaseTime = %d, want 86400", s.LeaseTime)
	}
}

func TestDHCPStatus_TooShort(t *testing.T) {
	data := make([]byte, dhcpStatusSize-1)
	_, err := decodeDHCPStatus(data)
	if err == nil {
		t.Fatal("expected error for short data")
	}
}
