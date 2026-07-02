package grout

import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
)

// AddrFamily identifies the address family.
type AddrFamily uint8

const (
	AddrFamilyUnspec AddrFamily = 0
	AddrFamilyIPv4   AddrFamily = 2  // AF_INET
	AddrFamilyIPv6   AddrFamily = 10 // AF_INET6
)

// IP4Addr represents an IPv4 address in network byte order.
type IP4Addr [4]byte

// ParseIP4 parses a dotted-decimal IPv4 string.
func ParseIP4(s string) (IP4Addr, error) {
	addr, err := netip.ParseAddr(s)
	if err != nil || !addr.Is4() {
		return IP4Addr{}, fmt.Errorf("invalid IPv4 address: %q", s)
	}
	return IP4Addr(addr.As4()), nil
}

// MustParseIP4 is like ParseIP4 but panics on error.
func MustParseIP4(s string) IP4Addr {
	a, err := ParseIP4(s)
	if err != nil {
		panic(err)
	}
	return a
}

func (a IP4Addr) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", a[0], a[1], a[2], a[3])
}

// ToUint32 returns the address as a uint32 in network byte order.
func (a IP4Addr) ToUint32() uint32 {
	return binary.BigEndian.Uint32(a[:])
}

// IP4AddrFromUint32 creates an IP4Addr from a uint32 in network byte order.
func IP4AddrFromUint32(v uint32) IP4Addr {
	var a IP4Addr
	binary.BigEndian.PutUint32(a[:], v)
	return a
}

// IP4Net represents an IPv4 network (address + prefix length).
type IP4Net struct {
	Addr      IP4Addr
	PrefixLen uint8
}

// ParseIP4Net parses a CIDR string like "192.168.1.0/24".
func ParseIP4Net(s string) (IP4Net, error) {
	prefix, err := netip.ParsePrefix(s)
	if err != nil || !prefix.Addr().Is4() {
		return IP4Net{}, fmt.Errorf("invalid IPv4 network: %q", s)
	}
	return IP4Net{
		Addr:      IP4Addr(prefix.Addr().As4()),
		PrefixLen: uint8(prefix.Bits()),
	}, nil
}

// MustParseIP4Net is like ParseIP4Net but panics on error.
func MustParseIP4Net(s string) IP4Net {
	n, err := ParseIP4Net(s)
	if err != nil {
		panic(err)
	}
	return n
}

func (n IP4Net) String() string {
	return fmt.Sprintf("%s/%d", n.Addr, n.PrefixLen)
}

// IP6Addr represents an IPv6 address.
type IP6Addr [16]byte

// ParseIP6 parses an IPv6 address string.
func ParseIP6(s string) (IP6Addr, error) {
	addr, err := netip.ParseAddr(s)
	if err != nil || !addr.Is6() {
		return IP6Addr{}, fmt.Errorf("invalid IPv6 address: %q", s)
	}
	return IP6Addr(addr.As16()), nil
}

// MustParseIP6 is like ParseIP6 but panics on error.
func MustParseIP6(s string) IP6Addr {
	a, err := ParseIP6(s)
	if err != nil {
		panic(err)
	}
	return a
}

func (a IP6Addr) String() string {
	addr := netip.AddrFrom16(a)
	return addr.String()
}

// IP6Net represents an IPv6 network (address + prefix length).
type IP6Net struct {
	Addr      IP6Addr
	PrefixLen uint8
}

// ParseIP6Net parses a CIDR string like "fd00::/64".
func ParseIP6Net(s string) (IP6Net, error) {
	prefix, err := netip.ParsePrefix(s)
	if err != nil || !prefix.Addr().Is6() {
		return IP6Net{}, fmt.Errorf("invalid IPv6 network: %q", s)
	}
	return IP6Net{
		Addr:      IP6Addr(prefix.Addr().As16()),
		PrefixLen: uint8(prefix.Bits()),
	}, nil
}

// MustParseIP6Net is like ParseIP6Net but panics on error.
func MustParseIP6Net(s string) IP6Net {
	n, err := ParseIP6Net(s)
	if err != nil {
		panic(err)
	}
	return n
}

func (n IP6Net) String() string {
	return fmt.Sprintf("%s/%d", n.Addr, n.PrefixLen)
}

// EtherAddr represents a 48-bit Ethernet MAC address.
type EtherAddr [6]byte

// ParseEtherAddr parses a colon-separated MAC address string like "aa:bb:cc:dd:ee:ff".
func ParseEtherAddr(s string) (EtherAddr, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 6 {
		return EtherAddr{}, fmt.Errorf("invalid MAC address: %q", s)
	}
	var addr EtherAddr
	for i, p := range parts {
		v, err := strconv.ParseUint(p, 16, 8)
		if err != nil {
			return EtherAddr{}, fmt.Errorf("invalid MAC address: %q", s)
		}
		addr[i] = byte(v)
	}
	return addr, nil
}

// MustParseEtherAddr is like ParseEtherAddr but panics on error.
func MustParseEtherAddr(s string) EtherAddr {
	a, err := ParseEtherAddr(s)
	if err != nil {
		panic(err)
	}
	return a
}

func (a EtherAddr) String() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		a[0], a[1], a[2], a[3], a[4], a[5])
}

// L3Addr represents a layer-3 address (IPv4 or IPv6).
type L3Addr struct {
	Family AddrFamily
	IP4    IP4Addr
	IP6    IP6Addr
}

// L3AddrFromIP4 creates an L3Addr from an IPv4 address.
func L3AddrFromIP4(addr IP4Addr) L3Addr {
	return L3Addr{Family: AddrFamilyIPv4, IP4: addr}
}

// L3AddrFromIP6 creates an L3Addr from an IPv6 address.
func L3AddrFromIP6(addr IP6Addr) L3Addr {
	return L3Addr{Family: AddrFamilyIPv6, IP6: addr}
}

func (a L3Addr) String() string {
	switch a.Family {
	case AddrFamilyIPv4:
		return a.IP4.String()
	case AddrFamilyIPv6:
		return a.IP6.String()
	default:
		return "<unspec>"
	}
}

// ClockNS represents a timestamp in nanoseconds.
type ClockNS int64

// l3AddrWireSize is the size of L3Addr on the wire.
// Family (1 byte) + 3 bytes padding + 16 bytes address = 20 bytes.
const l3AddrWireSize = 20

// encodeL3Addr serializes an L3Addr into wire format.
func encodeL3Addr(buf []byte, a L3Addr) {
	buf[0] = byte(a.Family)
	buf[1] = 0
	buf[2] = 0
	buf[3] = 0
	// Clear the 16-byte address area.
	for i := 4; i < l3AddrWireSize; i++ {
		buf[i] = 0
	}
	switch a.Family {
	case AddrFamilyIPv4:
		copy(buf[4:8], a.IP4[:])
	case AddrFamilyIPv6:
		copy(buf[4:20], a.IP6[:])
	}
}

// decodeL3Addr deserializes an L3Addr from wire format.
func decodeL3Addr(buf []byte) L3Addr {
	a := L3Addr{
		Family: AddrFamily(buf[0]),
	}
	switch a.Family {
	case AddrFamilyIPv4:
		copy(a.IP4[:], buf[4:8])
	case AddrFamilyIPv6:
		copy(a.IP6[:], buf[4:20])
	}
	return a
}
