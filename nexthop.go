package grout

import (
	"encoding/binary"
	"fmt"
)

// NHType identifies the nexthop type.
type NHType uint8

const (
	NHTypeL3        NHType = 0
	NHTypeSR6Output NHType = 1
	NHTypeSR6Local  NHType = 2
	NHTypeDNAT      NHType = 3
	NHTypeBlackhole NHType = 4
	NHTypeReject    NHType = 5
	NHTypeGroup     NHType = 6
)

// NHOrigin identifies how a nexthop was created.
type NHOrigin uint8

const (
	NHOriginUnspec NHOrigin = 0
	NHOriginStatic NHOrigin = 4
	NHOriginZebra  NHOrigin = 11
	NHOriginDHCP   NHOrigin = 16
)

// NHState represents the nexthop resolution state.
type NHState uint8

const (
	NHStateNew       NHState = 0
	NHStatePending   NHState = 1
	NHStateReachable NHState = 2
	NHStateStale     NHState = 3
	NHStateFailed    NHState = 4
)

// NHFlags represents nexthop flags as a bitmask.
type NHFlags uint8

const (
	NHFlagStatic  NHFlags = 1 << 0
	NHFlagLocal   NHFlags = 1 << 1
	NHFlagGateway NHFlags = 1 << 2
	NHFlagLink    NHFlags = 1 << 3
	NHFlagMcast   NHFlags = 1 << 4
	NHFlagRemote  NHFlags = 1 << 5
)

// Nexthop represents a nexthop entry.
type Nexthop struct {
	Type    NHType
	Origin  NHOrigin
	IfaceID uint16
	VRFID   uint16
	NHID    uint32

	L3      *NHInfoL3
	Group   *NHInfoGroup
	SRv6    *NHInfoSRv6
	SRv6Loc *NHInfoSRv6Local
	DNAT    *NHInfoDNAT
}

// NHInfoL3 contains L3 nexthop-specific information.
type NHInfoL3 struct {
	State     NHState
	Flags     NHFlags
	Family    AddrFamily
	PrefixLen uint8
	Addr      L3Addr
	MAC       EtherAddr
}

// NHGroupMember represents a member in a nexthop group.
type NHGroupMember struct {
	NHID   uint32
	Weight uint32
}

// NHInfoGroup contains nexthop group information.
type NHInfoGroup struct {
	Members []NHGroupMember
}

// NHInfoSRv6 contains SRv6 encapsulation nexthop info.
type NHInfoSRv6 struct {
	EncapBehavior SRv6EncapBehavior
	SegList       []IP6Addr
}

// SRv6EncapBehavior defines the SRv6 encapsulation behavior.
type SRv6EncapBehavior uint8

const (
	SRv6EncapBehaviorEncaps    SRv6EncapBehavior = 0
	SRv6EncapBehaviorEncapsRed SRv6EncapBehavior = 1
)

// SRv6Behavior defines SRv6 local segment behavior.
type SRv6Behavior uint16

const (
	SRv6BehaviorEnd    SRv6Behavior = 0x0001
	SRv6BehaviorEndT   SRv6Behavior = 0x0009
	SRv6BehaviorEndDT6 SRv6Behavior = 0x0012
	SRv6BehaviorEndDT4 SRv6Behavior = 0x0013
	SRv6BehaviorEndDT46 SRv6Behavior = 0x0014
)

// SRv6Flags represents SRv6 flags.
type SRv6Flags uint8

const (
	SRv6FlagPSP     SRv6Flags = 1 << 0
	SRv6FlagUSD     SRv6Flags = 1 << 1
	SRv6FlagNextCSID SRv6Flags = 1 << 2
)

// NHInfoSRv6Local contains SRv6 local segment nexthop info.
type NHInfoSRv6Local struct {
	OutVRFID  uint16
	Behavior  SRv6Behavior
	Flags     SRv6Flags
	BlockBits uint8
	CSIDBits  uint8
}

// NHInfoDNAT contains DNAT nexthop info.
type NHInfoDNAT struct {
	Addr IP4Addr
}

// NHConfig holds nexthop table configuration.
type NHConfig struct {
	MaxCount             uint32
	LifetimeReachable    uint32
	LifetimeUnreachable  uint32
	MaxHeldPkts          uint32
	MaxUcastProbes       uint32
	MaxBcastProbes       uint32
}

// nexthop base wire size: type(1) + origin(1) + iface_id(2) + vrf_id(2) + pad(2) + nh_id(4) = 12
const nexthopBaseSize = 12

// decodeNexthop deserializes a Nexthop from wire format.
func decodeNexthop(data []byte) (Nexthop, error) {
	if len(data) < nexthopBaseSize {
		return Nexthop{}, fmt.Errorf("nexthop data too short: %d bytes", len(data))
	}

	nh := Nexthop{
		Type:    NHType(data[0]),
		Origin:  NHOrigin(data[1]),
		IfaceID: binary.LittleEndian.Uint16(data[2:4]),
		VRFID:   binary.LittleEndian.Uint16(data[4:6]),
		NHID:    binary.LittleEndian.Uint32(data[8:12]),
	}

	info := data[nexthopBaseSize:]

	switch nh.Type {
	case NHTypeL3:
		if len(info) >= 4+l3AddrWireSize+6 {
			nh.L3 = &NHInfoL3{
				State:     NHState(info[0]),
				Flags:     NHFlags(info[1]),
				Family:    AddrFamily(info[2]),
				PrefixLen: info[3],
				Addr:      decodeL3Addr(info[4 : 4+l3AddrWireSize]),
			}
			copy(nh.L3.MAC[:], info[4+l3AddrWireSize:4+l3AddrWireSize+6])
		}
	case NHTypeGroup:
		memberSize := 8 // nhid(4) + weight(4)
		numMembers := len(info) / memberSize
		if numMembers > 0 {
			nh.Group = &NHInfoGroup{
				Members: make([]NHGroupMember, numMembers),
			}
			for i := 0; i < numMembers; i++ {
				off := i * memberSize
				nh.Group.Members[i] = NHGroupMember{
					NHID:   binary.LittleEndian.Uint32(info[off : off+4]),
					Weight: binary.LittleEndian.Uint32(info[off+4 : off+8]),
				}
			}
		}
	case NHTypeSR6Output:
		if len(info) >= 4 {
			nh.SRv6 = &NHInfoSRv6{
				EncapBehavior: SRv6EncapBehavior(info[0]),
			}
			// Segment list follows after 4 bytes of header.
			segData := info[4:]
			numSegs := len(segData) / 16
			nh.SRv6.SegList = make([]IP6Addr, numSegs)
			for i := 0; i < numSegs; i++ {
				copy(nh.SRv6.SegList[i][:], segData[i*16:(i+1)*16])
			}
		}
	case NHTypeSR6Local:
		if len(info) >= 8 {
			nh.SRv6Loc = &NHInfoSRv6Local{
				OutVRFID:  binary.LittleEndian.Uint16(info[0:2]),
				Behavior:  SRv6Behavior(binary.LittleEndian.Uint16(info[2:4])),
				Flags:     SRv6Flags(info[4]),
				BlockBits: info[5],
				CSIDBits:  info[6],
			}
		}
	case NHTypeDNAT:
		if len(info) >= 4 {
			nh.DNAT = &NHInfoDNAT{}
			copy(nh.DNAT.Addr[:], info[0:4])
		}
	}

	return nh, nil
}

const nhConfigSize = 24 // 6 x uint32

func decodeNHConfig(data []byte) (NHConfig, error) {
	if len(data) < nhConfigSize {
		return NHConfig{}, fmt.Errorf("nh config data too short: %d bytes", len(data))
	}
	return NHConfig{
		MaxCount:            binary.LittleEndian.Uint32(data[0:4]),
		LifetimeReachable:   binary.LittleEndian.Uint32(data[4:8]),
		LifetimeUnreachable: binary.LittleEndian.Uint32(data[8:12]),
		MaxHeldPkts:         binary.LittleEndian.Uint32(data[12:16]),
		MaxUcastProbes:      binary.LittleEndian.Uint32(data[16:20]),
		MaxBcastProbes:      binary.LittleEndian.Uint32(data[20:24]),
	}, nil
}

func encodeNHConfig(cfg NHConfig) []byte {
	buf := make([]byte, nhConfigSize)
	binary.LittleEndian.PutUint32(buf[0:4], cfg.MaxCount)
	binary.LittleEndian.PutUint32(buf[4:8], cfg.LifetimeReachable)
	binary.LittleEndian.PutUint32(buf[8:12], cfg.LifetimeUnreachable)
	binary.LittleEndian.PutUint32(buf[12:16], cfg.MaxHeldPkts)
	binary.LittleEndian.PutUint32(buf[16:20], cfg.MaxUcastProbes)
	binary.LittleEndian.PutUint32(buf[20:24], cfg.MaxBcastProbes)
	return buf
}

// NexthopAdd creates a new nexthop entry.
func (c *Client) NexthopAdd(nh Nexthop) error {
	if err := c.checkVersion(msgTypeNHAdd); err != nil {
		return err
	}
	// Encode base + type-specific info.
	payload := encodeNexthopBase(nh)
	_, err := c.request(msgTypeNHAdd, payload)
	return err
}

// NexthopDel deletes a nexthop entry.
func (c *Client) NexthopDel(nhID uint32) error {
	if err := c.checkVersion(msgTypeNHDel); err != nil {
		return err
	}
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, nhID)
	_, err := c.request(msgTypeNHDel, payload)
	return err
}

// NexthopGet retrieves a nexthop by ID.
func (c *Client) NexthopGet(nhID uint32) (*Nexthop, error) {
	if err := c.checkVersion(msgTypeNHGet); err != nil {
		return nil, err
	}
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, nhID)
	resp, err := c.request(msgTypeNHGet, payload)
	if err != nil {
		return nil, err
	}
	nh, err := decodeNexthop(resp)
	if err != nil {
		return nil, err
	}
	return &nh, nil
}

// NexthopList returns all nexthop entries.
func (c *Client) NexthopList() ([]Nexthop, error) {
	if err := c.checkVersion(msgTypeNHList); err != nil {
		return nil, err
	}
	results, err := c.requestStreamNoPayload(msgTypeNHList)
	if err != nil {
		return nil, err
	}
	nhs := make([]Nexthop, 0, len(results))
	for _, data := range results {
		nh, err := decodeNexthop(data)
		if err != nil {
			return nil, err
		}
		nhs = append(nhs, nh)
	}
	return nhs, nil
}

// NexthopConfigGet retrieves the nexthop table configuration.
func (c *Client) NexthopConfigGet() (*NHConfig, error) {
	if err := c.checkVersion(msgTypeNHConfigGet); err != nil {
		return nil, err
	}
	resp, err := c.requestNoPayload(msgTypeNHConfigGet)
	if err != nil {
		return nil, err
	}
	cfg, err := decodeNHConfig(resp)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// NexthopConfigSet updates the nexthop table configuration.
func (c *Client) NexthopConfigSet(config NHConfig) error {
	if err := c.checkVersion(msgTypeNHConfigSet); err != nil {
		return err
	}
	_, err := c.request(msgTypeNHConfigSet, encodeNHConfig(config))
	return err
}

func encodeNexthopBase(nh Nexthop) []byte {
	buf := make([]byte, nexthopBaseSize)
	buf[0] = byte(nh.Type)
	buf[1] = byte(nh.Origin)
	binary.LittleEndian.PutUint16(buf[2:4], nh.IfaceID)
	binary.LittleEndian.PutUint16(buf[4:6], nh.VRFID)
	binary.LittleEndian.PutUint32(buf[8:12], nh.NHID)

	switch nh.Type {
	case NHTypeL3:
		if nh.L3 != nil {
			info := make([]byte, 4+l3AddrWireSize+6)
			info[0] = byte(nh.L3.State)
			info[1] = byte(nh.L3.Flags)
			info[2] = byte(nh.L3.Family)
			info[3] = nh.L3.PrefixLen
			encodeL3Addr(info[4:4+l3AddrWireSize], nh.L3.Addr)
			copy(info[4+l3AddrWireSize:], nh.L3.MAC[:])
			buf = append(buf, info...)
		}
	case NHTypeDNAT:
		if nh.DNAT != nil {
			buf = append(buf, nh.DNAT.Addr[:]...)
		}
	}

	return buf
}
