package grout

// IfaceType identifies the type of a network interface.
type IfaceType uint8

const (
	IfaceTypeUndef  IfaceType = 0
	IfaceTypeVRF    IfaceType = 1
	IfaceTypePort   IfaceType = 2
	IfaceTypeVLAN   IfaceType = 3
	IfaceTypeIPIP   IfaceType = 4
	IfaceTypeBond   IfaceType = 5
	IfaceTypeBridge IfaceType = 6
	IfaceTypeVXLAN  IfaceType = 7
)

func (t IfaceType) String() string {
	switch t {
	case IfaceTypeVRF:
		return "vrf"
	case IfaceTypePort:
		return "port"
	case IfaceTypeVLAN:
		return "vlan"
	case IfaceTypeIPIP:
		return "ipip"
	case IfaceTypeBond:
		return "bond"
	case IfaceTypeBridge:
		return "bridge"
	case IfaceTypeVXLAN:
		return "vxlan"
	default:
		return "undef"
	}
}

// IfaceFlags represents interface flags as a bitmask.
type IfaceFlags uint16

const (
	IfaceFlagUp         IfaceFlags = 1 << 0
	IfaceFlagPromisc    IfaceFlags = 1 << 1
	IfaceFlagPacketTrace IfaceFlags = 1 << 2
	IfaceFlagSNATStatic IfaceFlags = 1 << 3
	IfaceFlagSNATDynamic IfaceFlags = 1 << 4
)

// IfaceState represents interface operational state as a bitmask.
type IfaceState uint16

const (
	IfaceStateRunning     IfaceState = 1 << 0
	IfaceStatePromiscFixed IfaceState = 1 << 1
	IfaceStateAllMulti    IfaceState = 1 << 2
)

// IfaceMode represents the interface mode.
type IfaceMode uint8

const (
	IfaceModeVRF    IfaceMode = 0
	IfaceModeXC     IfaceMode = 1
	IfaceModeBond   IfaceMode = 2
	IfaceModeBridge IfaceMode = 3
)

// Iface represents a network interface in grout.
type Iface struct {
	ID          uint16
	Type        IfaceType
	Mode        IfaceMode
	Flags       IfaceFlags
	State       IfaceState
	MTU         uint16
	VRFID       uint16
	DomainID    uint16
	Speed       uint32
	Name        string
	Description string

	// Type-specific info (only one will be populated based on Type).
	Port   *PortInfo
	VRF    *VRFInfo
	VLAN   *VLANInfo
	Bond   *BondInfo
	Bridge *BridgeInfo
	VXLAN  *VXLANInfo
	IPIP   *IPIPInfo
}

// PortInfo contains port-specific interface information.
type PortInfo struct {
	NRxQ       uint16
	NTxQ       uint16
	RxQSize    uint16
	TxQSize    uint16
	MAC        EtherAddr
	DevArgs    string
	DriverName string
}

// FIBConfig holds FIB table configuration for a VRF.
type FIBConfig struct {
	MaxRoutes uint32
	NumTbl8   uint32
}

// VRFInfo contains VRF-specific interface information.
type VRFInfo struct {
	IPv4FIB FIBConfig
	IPv6FIB FIBConfig
	MAC     EtherAddr
}

// VLANInfo contains VLAN-specific interface information.
type VLANInfo struct {
	ParentID uint16
	VLANID   uint16
	MAC      EtherAddr
}

// BondMode represents the bonding mode.
type BondMode uint8

const (
	BondModeActiveBackup BondMode = 1
	BondModeLACP         BondMode = 2
)

// BondAlgo represents the bonding hash algorithm.
type BondAlgo uint8

const (
	BondAlgoRSS  BondAlgo = 1
	BondAlgoL2   BondAlgo = 2
	BondAlgoL3L4 BondAlgo = 3
)

// BondMember represents a bond member interface.
type BondMember struct {
	IfaceID uint16
	Active  bool
}

// BondInfo contains bond-specific interface information.
type BondInfo struct {
	Mode          BondMode
	Algo          BondAlgo
	MAC           EtherAddr
	PrimaryMember uint8
	Members       []BondMember
}

// BridgeFlags represents bridge flags.
type BridgeFlags uint16

const (
	BridgeFlagNoFlood BridgeFlags = 1 << 0
	BridgeFlagNoLearn BridgeFlags = 1 << 1
)

// BridgeInfo contains bridge-specific interface information.
type BridgeInfo struct {
	AgeingTime uint32
	Flags      BridgeFlags
	MAC        EtherAddr
	Members    []uint16
}

// VXLANInfo contains VXLAN-specific interface information.
type VXLANInfo struct {
	VNI        uint32
	EncapVRFID uint16
	DstPort    uint16
	Local      L3Addr
	MAC        EtherAddr
}

// IPIPInfo contains IPIP tunnel-specific interface information.
type IPIPInfo struct {
	Local  IP4Addr
	Remote IP4Addr
}

// IfaceStats holds per-interface traffic counters.
type IfaceStats struct {
	IfaceID     uint16
	RxPackets   uint64
	RxBytes     uint64
	RxDrops     uint64
	TxPackets   uint64
	TxBytes     uint64
	TxErrors    uint64
	CPRxPackets uint64
	CPRxBytes   uint64
	CPTxPackets uint64
	CPTxBytes   uint64
}

// IfaceMAC represents a MAC address entry for an interface.
type IfaceMAC struct {
	IfaceID  uint16
	RefCount uint16
	Primary  bool
	MAC      EtherAddr
}

// IfaceSetAttr defines which attributes to set on an interface.
type IfaceSetAttr uint32

const (
	IfaceSetFlags       IfaceSetAttr = 1 << 0
	IfaceSetMTU         IfaceSetAttr = 1 << 1
	IfaceSetVRF         IfaceSetAttr = 1 << 2
	IfaceSetDescription IfaceSetAttr = 1 << 3
)
