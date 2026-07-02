package grout

import (
	"encoding/binary"
	"fmt"
)

// Wire sizes for interface-related structs.
const (
	// gr_iface base: id(2) + type(1) + mode(1) + flags(2) + state(2)
	//   + mtu(2) + vrf_id(2) + domain_id(2) + speed(4) + name[64] + desc[64]
	//   = 146, then + info[] flexible array
	ifaceBaseSize = 146
	ifaceNameLen  = 64
	ifaceDescLen  = 64
)

// InterfaceAddRequest describes parameters for creating a new interface.
type InterfaceAddRequest struct {
	Type    IfaceType
	Name    string
	VRFID   uint16
	Mode    IfaceMode
	Flags   IfaceFlags
	MTU     uint16
	DevArgs string // for port interfaces

	// Port-specific fields.
	NRxQ    uint16
	RxQSize uint16

	// VLAN-specific fields.
	ParentID uint16
	VLANID   uint16

	// IPIP-specific fields.
	Local  IP4Addr
	Remote IP4Addr

	// VXLAN-specific fields.
	VNI        uint32
	EncapVRFID uint16
	DstPort    uint16
	VXLANLocal L3Addr
}

// InterfaceSetRequest describes parameters for modifying an interface.
type InterfaceSetRequest struct {
	IfaceID     uint16
	SetAttrs    IfaceSetAttr
	Flags       IfaceFlags
	MTU         uint16
	VRFID       uint16
	Description string
}

type interfaceListConfig struct {
	ifaceType    IfaceType
	hasTypeFilter bool
}

// InterfaceListOption configures an InterfaceList query.
type InterfaceListOption func(*interfaceListConfig)

// WithIfaceType filters the list to interfaces of the given type.
func WithIfaceType(t IfaceType) InterfaceListOption {
	return func(c *interfaceListConfig) {
		c.ifaceType = t
		c.hasTypeFilter = true
	}
}

// InterfaceAdd creates a new interface and returns its ID.
func (c *Client) InterfaceAdd(req InterfaceAddRequest) (uint16, error) {
	if err := c.checkVersion(msgTypeIfaceAdd); err != nil {
		return 0, err
	}

	payload := encodeIfaceAdd(req)
	resp, err := c.request(msgTypeIfaceAdd, payload)
	if err != nil {
		return 0, err
	}

	iface, err := decodeIface(resp)
	if err != nil {
		return 0, err
	}
	return iface.ID, nil
}

// InterfaceDel deletes an interface by ID.
func (c *Client) InterfaceDel(ifaceID uint16) error {
	if err := c.checkVersion(msgTypeIfaceDel); err != nil {
		return err
	}

	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, ifaceID)
	_, err := c.request(msgTypeIfaceDel, payload)
	return err
}

// InterfaceGet retrieves an interface by ID.
func (c *Client) InterfaceGet(ifaceID uint16) (*Iface, error) {
	if err := c.checkVersion(msgTypeIfaceGet); err != nil {
		return nil, err
	}

	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, ifaceID)
	resp, err := c.request(msgTypeIfaceGet, payload)
	if err != nil {
		return nil, err
	}

	iface, err := decodeIface(resp)
	if err != nil {
		return nil, err
	}
	return &iface, nil
}

// InterfaceGetByName retrieves an interface by name.
func (c *Client) InterfaceGetByName(name string) (*Iface, error) {
	ifaces, err := c.InterfaceList()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Name == name {
			return &iface, nil
		}
	}
	return nil, fmt.Errorf("%w: interface %q", ErrNotFound, name)
}

// InterfaceList returns all interfaces, optionally filtered by type.
func (c *Client) InterfaceList(opts ...InterfaceListOption) ([]Iface, error) {
	if err := c.checkVersion(msgTypeIfaceList); err != nil {
		return nil, err
	}

	cfg := &interfaceListConfig{}
	for _, o := range opts {
		o(cfg)
	}

	var payload []byte
	if cfg.hasTypeFilter {
		payload = []byte{byte(cfg.ifaceType)}
	}

	results, err := c.requestStream(msgTypeIfaceList, payload)
	if err != nil {
		return nil, err
	}

	ifaces := make([]Iface, 0, len(results))
	for _, data := range results {
		iface, err := decodeIface(data)
		if err != nil {
			return nil, err
		}
		ifaces = append(ifaces, iface)
	}
	return ifaces, nil
}

// InterfaceSet modifies interface attributes.
func (c *Client) InterfaceSet(req InterfaceSetRequest) error {
	if err := c.checkVersion(msgTypeIfaceSet); err != nil {
		return err
	}

	payload := encodeIfaceSet(req)
	_, err := c.request(msgTypeIfaceSet, payload)
	return err
}

// InterfaceStatsGet returns statistics for all interfaces.
func (c *Client) InterfaceStatsGet() ([]IfaceStats, error) {
	if err := c.checkVersion(msgTypeIfaceStatsGet); err != nil {
		return nil, err
	}

	results, err := c.requestStreamNoPayload(msgTypeIfaceStatsGet)
	if err != nil {
		return nil, err
	}

	stats := make([]IfaceStats, 0, len(results))
	for _, data := range results {
		s, err := decodeIfaceStats(data)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// InterfaceMACAdd adds a MAC address to an interface.
func (c *Client) InterfaceMACAdd(ifaceID uint16, mac EtherAddr) error {
	if err := c.checkVersion(msgTypeIfaceMACAdd); err != nil {
		return err
	}

	payload := make([]byte, 10) // iface_id(2) + pad(2) + mac(6)
	binary.LittleEndian.PutUint16(payload[0:2], ifaceID)
	copy(payload[4:10], mac[:])
	_, err := c.request(msgTypeIfaceMACAdd, payload)
	return err
}

// InterfaceMACDel removes a MAC address from an interface.
func (c *Client) InterfaceMACDel(ifaceID uint16, mac EtherAddr) error {
	if err := c.checkVersion(msgTypeIfaceMACDel); err != nil {
		return err
	}

	payload := make([]byte, 10) // iface_id(2) + pad(2) + mac(6)
	binary.LittleEndian.PutUint16(payload[0:2], ifaceID)
	copy(payload[4:10], mac[:])
	_, err := c.request(msgTypeIfaceMACDel, payload)
	return err
}

// InterfaceMACList returns MAC addresses for an interface.
func (c *Client) InterfaceMACList(ifaceID uint16) ([]IfaceMAC, error) {
	if err := c.checkVersion(msgTypeIfaceMACList); err != nil {
		return nil, err
	}

	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, ifaceID)

	results, err := c.requestStream(msgTypeIfaceMACList, payload)
	if err != nil {
		return nil, err
	}

	macs := make([]IfaceMAC, 0, len(results))
	for _, data := range results {
		m, err := decodeIfaceMAC(data)
		if err != nil {
			return nil, err
		}
		macs = append(macs, m)
	}
	return macs, nil
}

// InterfaceMACSet sets the primary MAC address for an interface.
func (c *Client) InterfaceMACSet(ifaceID uint16, mac EtherAddr) error {
	if err := c.checkVersion(msgTypeIfaceMACSet); err != nil {
		return err
	}

	payload := make([]byte, 10) // iface_id(2) + pad(2) + mac(6)
	binary.LittleEndian.PutUint16(payload[0:2], ifaceID)
	copy(payload[4:10], mac[:])
	_, err := c.request(msgTypeIfaceMACSet, payload)
	return err
}

// Encoding/decoding helpers.

func encodeIfaceAdd(req InterfaceAddRequest) []byte {
	buf := make([]byte, ifaceBaseSize)
	// id is 0 for add (server assigns)
	buf[2] = byte(req.Type)
	buf[3] = byte(req.Mode)
	binary.LittleEndian.PutUint16(buf[4:6], uint16(req.Flags))
	// state is read-only
	binary.LittleEndian.PutUint16(buf[8:10], req.MTU)
	binary.LittleEndian.PutUint16(buf[10:12], req.VRFID)
	// domain_id at 12:14
	// speed at 14:18 (read-only)

	nameOff := 18
	copy(buf[nameOff:nameOff+ifaceNameLen], req.Name)
	descOff := nameOff + ifaceNameLen
	_ = descOff

	// Append type-specific info as needed.
	switch req.Type {
	case IfaceTypePort:
		info := make([]byte, 8) // n_rxq(2) + n_txq(2) + rxq_size(2) + txq_size(2)
		binary.LittleEndian.PutUint16(info[0:2], req.NRxQ)
		binary.LittleEndian.PutUint16(info[4:6], req.RxQSize)
		// Append devargs as null-terminated string.
		if req.DevArgs != "" {
			info = append(info, make([]byte, 6)...) // MAC placeholder
			info = append(info, []byte(req.DevArgs)...)
			info = append(info, 0)
		}
		buf = append(buf, info...)
	case IfaceTypeVLAN:
		info := make([]byte, 4) // parent_id(2) + vlan_id(2)
		binary.LittleEndian.PutUint16(info[0:2], req.ParentID)
		binary.LittleEndian.PutUint16(info[2:4], req.VLANID)
		buf = append(buf, info...)
	case IfaceTypeIPIP:
		info := make([]byte, 8) // local(4) + remote(4)
		copy(info[0:4], req.Local[:])
		copy(info[4:8], req.Remote[:])
		buf = append(buf, info...)
	case IfaceTypeVXLAN:
		info := make([]byte, 4+2+2+l3AddrWireSize) // vni(4) + encap_vrf(2) + dst_port(2) + local
		binary.LittleEndian.PutUint32(info[0:4], req.VNI)
		binary.LittleEndian.PutUint16(info[4:6], req.EncapVRFID)
		binary.LittleEndian.PutUint16(info[6:8], req.DstPort)
		encodeL3Addr(info[8:8+l3AddrWireSize], req.VXLANLocal)
		buf = append(buf, info...)
	}

	return buf
}

func encodeIfaceSet(req InterfaceSetRequest) []byte {
	// set_attrs(4) + iface_id(2) + flags(2) + mtu(2) + vrf_id(2) + desc[64]
	buf := make([]byte, 4+2+2+2+2+ifaceDescLen)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(req.SetAttrs))
	binary.LittleEndian.PutUint16(buf[4:6], req.IfaceID)
	binary.LittleEndian.PutUint16(buf[6:8], uint16(req.Flags))
	binary.LittleEndian.PutUint16(buf[8:10], req.MTU)
	binary.LittleEndian.PutUint16(buf[10:12], req.VRFID)
	copy(buf[12:12+ifaceDescLen], req.Description)
	return buf
}

func decodeIface(data []byte) (Iface, error) {
	if len(data) < ifaceBaseSize {
		return Iface{}, fmt.Errorf("iface data too short: %d bytes", len(data))
	}

	nameOff := 18
	descOff := nameOff + ifaceNameLen

	iface := Iface{
		ID:          binary.LittleEndian.Uint16(data[0:2]),
		Type:        IfaceType(data[2]),
		Mode:        IfaceMode(data[3]),
		Flags:       IfaceFlags(binary.LittleEndian.Uint16(data[4:6])),
		State:       IfaceState(binary.LittleEndian.Uint16(data[6:8])),
		MTU:         binary.LittleEndian.Uint16(data[8:10]),
		VRFID:       binary.LittleEndian.Uint16(data[10:12]),
		DomainID:    binary.LittleEndian.Uint16(data[12:14]),
		Speed:       binary.LittleEndian.Uint32(data[14:18]),
		Name:        cstring(data[nameOff : nameOff+ifaceNameLen]),
		Description: cstring(data[descOff : descOff+ifaceDescLen]),
	}

	info := data[ifaceBaseSize:]
	if len(info) == 0 {
		return iface, nil
	}

	switch iface.Type {
	case IfaceTypePort:
		if len(info) >= 8 {
			p := &PortInfo{
				NRxQ:    binary.LittleEndian.Uint16(info[0:2]),
				NTxQ:    binary.LittleEndian.Uint16(info[2:4]),
				RxQSize: binary.LittleEndian.Uint16(info[4:6]),
				TxQSize: binary.LittleEndian.Uint16(info[6:8]),
			}
			if len(info) >= 14 {
				copy(p.MAC[:], info[8:14])
			}
			if len(info) > 14 {
				p.DevArgs = cstring(info[14:])
			}
			iface.Port = p
		}
	case IfaceTypeVRF:
		if len(info) >= 22 {
			iface.VRF = &VRFInfo{
				IPv4FIB: FIBConfig{
					MaxRoutes: binary.LittleEndian.Uint32(info[0:4]),
					NumTbl8:   binary.LittleEndian.Uint32(info[4:8]),
				},
				IPv6FIB: FIBConfig{
					MaxRoutes: binary.LittleEndian.Uint32(info[8:12]),
					NumTbl8:   binary.LittleEndian.Uint32(info[12:16]),
				},
			}
			copy(iface.VRF.MAC[:], info[16:22])
		}
	case IfaceTypeVLAN:
		if len(info) >= 10 {
			iface.VLAN = &VLANInfo{
				ParentID: binary.LittleEndian.Uint16(info[0:2]),
				VLANID:   binary.LittleEndian.Uint16(info[2:4]),
			}
			copy(iface.VLAN.MAC[:], info[4:10])
		}
	case IfaceTypeIPIP:
		if len(info) >= 8 {
			iface.IPIP = &IPIPInfo{}
			copy(iface.IPIP.Local[:], info[0:4])
			copy(iface.IPIP.Remote[:], info[4:8])
		}
	case IfaceTypeVXLAN:
		if len(info) >= 8+l3AddrWireSize+6 {
			iface.VXLAN = &VXLANInfo{
				VNI:        binary.LittleEndian.Uint32(info[0:4]),
				EncapVRFID: binary.LittleEndian.Uint16(info[4:6]),
				DstPort:    binary.LittleEndian.Uint16(info[6:8]),
				Local:      decodeL3Addr(info[8 : 8+l3AddrWireSize]),
			}
			copy(iface.VXLAN.MAC[:], info[8+l3AddrWireSize:8+l3AddrWireSize+6])
		}
	}

	return iface, nil
}

const ifaceStatsSize = 2 + 6 + 10*8 // iface_id(2) + pad(6) + 10 x uint64

func decodeIfaceStats(data []byte) (IfaceStats, error) {
	if len(data) < ifaceStatsSize {
		return IfaceStats{}, fmt.Errorf("iface stats data too short: %d bytes", len(data))
	}
	return IfaceStats{
		IfaceID:     binary.LittleEndian.Uint16(data[0:2]),
		RxPackets:   binary.LittleEndian.Uint64(data[8:16]),
		RxBytes:     binary.LittleEndian.Uint64(data[16:24]),
		RxDrops:     binary.LittleEndian.Uint64(data[24:32]),
		TxPackets:   binary.LittleEndian.Uint64(data[32:40]),
		TxBytes:     binary.LittleEndian.Uint64(data[40:48]),
		TxErrors:    binary.LittleEndian.Uint64(data[48:56]),
		CPRxPackets: binary.LittleEndian.Uint64(data[56:64]),
		CPRxBytes:   binary.LittleEndian.Uint64(data[64:72]),
		CPTxPackets: binary.LittleEndian.Uint64(data[72:80]),
		CPTxBytes:   binary.LittleEndian.Uint64(data[80:88]),
	}, nil
}

const ifaceMACSize = 12 // iface_id(2) + ref_count(2) + primary(1) + pad(1) + mac(6)

func decodeIfaceMAC(data []byte) (IfaceMAC, error) {
	if len(data) < ifaceMACSize {
		return IfaceMAC{}, fmt.Errorf("iface MAC data too short: %d bytes", len(data))
	}
	m := IfaceMAC{
		IfaceID:  binary.LittleEndian.Uint16(data[0:2]),
		RefCount: binary.LittleEndian.Uint16(data[2:4]),
		Primary:  data[4] != 0,
	}
	copy(m.MAC[:], data[6:12])
	return m, nil
}
