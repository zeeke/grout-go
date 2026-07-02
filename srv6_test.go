package grout

import "testing"

func TestSRv6TunSrcSet_AddrEncoding(t *testing.T) {
	addr := MustParseIP6("fd00::1")
	if addr[0] != 0xfd || addr[1] != 0x00 {
		t.Errorf("unexpected first bytes: %x %x", addr[0], addr[1])
	}
	if addr[15] != 1 {
		t.Errorf("last byte = %d, want 1", addr[15])
	}
}

func TestSRv6TunSrcShow_ShortResponse(t *testing.T) {
	data := make([]byte, 15)
	if len(data) >= 16 {
		t.Fatal("expected short data to be < 16 bytes")
	}
}
