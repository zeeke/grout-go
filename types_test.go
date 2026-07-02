package grout

import (
	"testing"
)

func TestIP4Addr_Parse(t *testing.T) {
	tests := []struct {
		input string
		want  IP4Addr
		err   bool
	}{
		{"192.168.1.1", IP4Addr{192, 168, 1, 1}, false},
		{"0.0.0.0", IP4Addr{0, 0, 0, 0}, false},
		{"255.255.255.255", IP4Addr{255, 255, 255, 255}, false},
		{"10.0.0.1", IP4Addr{10, 0, 0, 1}, false},
		{"invalid", IP4Addr{}, true},
		{"::1", IP4Addr{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseIP4(tt.input)
			if tt.err {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ParseIP4(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIP4Addr_String(t *testing.T) {
	a := IP4Addr{10, 20, 30, 40}
	if s := a.String(); s != "10.20.30.40" {
		t.Errorf("String() = %q, want %q", s, "10.20.30.40")
	}
}

func TestIP4Addr_Uint32RoundTrip(t *testing.T) {
	a := IP4Addr{192, 168, 1, 1}
	v := a.ToUint32()
	b := IP4AddrFromUint32(v)
	if a != b {
		t.Errorf("round-trip failed: %v -> %d -> %v", a, v, b)
	}
}

func TestIP4Net_Parse(t *testing.T) {
	tests := []struct {
		input     string
		wantAddr  IP4Addr
		wantPfx   uint8
		err       bool
	}{
		{"192.168.1.0/24", IP4Addr{192, 168, 1, 0}, 24, false},
		{"10.0.0.0/8", IP4Addr{10, 0, 0, 0}, 8, false},
		{"0.0.0.0/0", IP4Addr{0, 0, 0, 0}, 0, false},
		{"invalid", IP4Addr{}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseIP4Net(tt.input)
			if tt.err {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Addr != tt.wantAddr || got.PrefixLen != tt.wantPfx {
				t.Errorf("got %v, want %s/%d", got, tt.wantAddr, tt.wantPfx)
			}
		})
	}
}

func TestIP4Net_String(t *testing.T) {
	n := IP4Net{IP4Addr{10, 0, 0, 0}, 8}
	if s := n.String(); s != "10.0.0.0/8" {
		t.Errorf("String() = %q, want %q", s, "10.0.0.0/8")
	}
}

func TestIP6Addr_Parse(t *testing.T) {
	tests := []struct {
		input string
		err   bool
	}{
		{"::1", false},
		{"fe80::1", false},
		{"fd00::1", false},
		{"invalid", true},
		{"192.168.1.1", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseIP6(tt.input)
			if tt.err {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			s := got.String()
			reparsed, err := ParseIP6(s)
			if err != nil {
				t.Fatalf("reparse error: %v", err)
			}
			if got != reparsed {
				t.Errorf("round-trip failed: %v -> %q -> %v", got, s, reparsed)
			}
		})
	}
}

func TestIP6Net_Parse(t *testing.T) {
	n, err := ParseIP6Net("fd00::/64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.PrefixLen != 64 {
		t.Errorf("PrefixLen = %d, want 64", n.PrefixLen)
	}
	if s := n.String(); s != "fd00::/64" {
		t.Errorf("String() = %q, want %q", s, "fd00::/64")
	}
}

func TestEtherAddr_Parse(t *testing.T) {
	tests := []struct {
		input string
		want  EtherAddr
		err   bool
	}{
		{"aa:bb:cc:dd:ee:ff", EtherAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}, false},
		{"00:00:00:00:00:00", EtherAddr{0, 0, 0, 0, 0, 0}, false},
		{"invalid", EtherAddr{}, true},
		{"aa:bb:cc:dd:ee", EtherAddr{}, true},
		{"gg:bb:cc:dd:ee:ff", EtherAddr{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseEtherAddr(tt.input)
			if tt.err {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEtherAddr_String(t *testing.T) {
	a := EtherAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	if s := a.String(); s != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("String() = %q, want %q", s, "aa:bb:cc:dd:ee:ff")
	}
}

func TestL3Addr(t *testing.T) {
	t.Run("IPv4", func(t *testing.T) {
		a := L3AddrFromIP4(IP4Addr{10, 0, 0, 1})
		if a.Family != AddrFamilyIPv4 {
			t.Errorf("Family = %d, want %d", a.Family, AddrFamilyIPv4)
		}
		if s := a.String(); s != "10.0.0.1" {
			t.Errorf("String() = %q, want %q", s, "10.0.0.1")
		}
	})

	t.Run("IPv6", func(t *testing.T) {
		addr := MustParseIP6("::1")
		a := L3AddrFromIP6(addr)
		if a.Family != AddrFamilyIPv6 {
			t.Errorf("Family = %d, want %d", a.Family, AddrFamilyIPv6)
		}
		if s := a.String(); s != "::1" {
			t.Errorf("String() = %q, want %q", s, "::1")
		}
	})

	t.Run("Unspec", func(t *testing.T) {
		a := L3Addr{}
		if s := a.String(); s != "<unspec>" {
			t.Errorf("String() = %q, want %q", s, "<unspec>")
		}
	})
}

func TestL3Addr_WireRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		addr L3Addr
	}{
		{"IPv4", L3AddrFromIP4(IP4Addr{192, 168, 1, 1})},
		{"IPv6", L3AddrFromIP6(MustParseIP6("fe80::1"))},
		{"Unspec", L3Addr{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, l3AddrWireSize)
			encodeL3Addr(buf, tt.addr)
			got := decodeL3Addr(buf)
			if got.Family != tt.addr.Family {
				t.Errorf("Family = %d, want %d", got.Family, tt.addr.Family)
			}
			if got.String() != tt.addr.String() {
				t.Errorf("decoded = %q, want %q", got, tt.addr)
			}
		})
	}
}

func TestMustParseIP4_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	MustParseIP4("invalid")
}

func TestMustParseIP6_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	MustParseIP6("invalid")
}

func TestMustParseIP4Net_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	MustParseIP4Net("invalid")
}

func TestMustParseIP6Net_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	MustParseIP6Net("invalid")
}

func TestMustParseEtherAddr_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	MustParseEtherAddr("invalid")
}

func TestCstring(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{"null_terminated", []byte("hello\x00world"), "hello"},
		{"no_null", []byte("hello"), "hello"},
		{"empty", []byte{0}, ""},
		{"all_null", []byte{0, 0, 0}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cstring(tt.input)
			if got != tt.want {
				t.Errorf("cstring(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
