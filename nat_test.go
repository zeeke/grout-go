package grout

import "testing"

func TestDNAT44_RoundTrip(t *testing.T) {
	p := DNAT44Policy{
		IfaceID: 3,
		Match:   MustParseIP4("10.0.0.1"),
		Replace: MustParseIP4("192.168.1.100"),
	}

	encoded := encodeDNAT44(p, false)
	decoded, err := decodeDNAT44(encoded)
	if err != nil {
		t.Fatalf("decodeDNAT44 error: %v", err)
	}
	if decoded.IfaceID != p.IfaceID {
		t.Errorf("IfaceID = %d, want %d", decoded.IfaceID, p.IfaceID)
	}
	if decoded.Match != p.Match {
		t.Errorf("Match = %v, want %v", decoded.Match, p.Match)
	}
	if decoded.Replace != p.Replace {
		t.Errorf("Replace = %v, want %v", decoded.Replace, p.Replace)
	}
}

func TestSNAT44_RoundTrip(t *testing.T) {
	p := SNAT44Policy{
		IfaceID: 5,
		Net:     MustParseIP4Net("10.0.0.0/8"),
		Replace: MustParseIP4("1.2.3.4"),
	}

	encoded := encodeSNAT44(p, true)
	decoded, err := decodeSNAT44(encoded)
	if err != nil {
		t.Fatalf("decodeSNAT44 error: %v", err)
	}
	if decoded.IfaceID != p.IfaceID {
		t.Errorf("IfaceID = %d, want %d", decoded.IfaceID, p.IfaceID)
	}
	if decoded.Net.PrefixLen != 8 {
		t.Errorf("PrefixLen = %d, want 8", decoded.Net.PrefixLen)
	}
	if decoded.Replace != p.Replace {
		t.Errorf("Replace = %v, want %v", decoded.Replace, p.Replace)
	}
}
