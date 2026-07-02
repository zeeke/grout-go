package grout

import (
	"encoding/binary"
	"fmt"
)

// DHCPState represents the DHCP client state.
type DHCPState uint8

const (
	DHCPStateInit       DHCPState = 0
	DHCPStateSelecting  DHCPState = 1
	DHCPStateRequesting DHCPState = 2
	DHCPStateBound      DHCPState = 3
	DHCPStateRenewing   DHCPState = 4
	DHCPStateRebinding  DHCPState = 5
)

// DHCPStatus represents the status of a DHCP client on an interface.
type DHCPStatus struct {
	IfaceID     uint16
	State       DHCPState
	ServerIP    IP4Addr
	AssignedIP  IP4Addr
	LeaseTime   uint32
	RenewalTime uint32
	RebindTime  uint32
}

const dhcpStatusSize = 24 // iface_id(2) + state(1) + pad(1) + server(4) + assigned(4) + lease(4) + renewal(4) + rebind(4)

// DHCPList returns the status of all DHCP clients.
func (c *Client) DHCPList() ([]DHCPStatus, error) {
	if err := c.checkVersion(msgTypeDHCPList); err != nil {
		return nil, err
	}
	results, err := c.requestStreamNoPayload(msgTypeDHCPList)
	if err != nil {
		return nil, err
	}

	statuses := make([]DHCPStatus, 0, len(results))
	for _, data := range results {
		s, err := decodeDHCPStatus(data)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, s)
	}
	return statuses, nil
}

// DHCPStart starts the DHCP client on an interface.
func (c *Client) DHCPStart(ifaceID uint16) error {
	if err := c.checkVersion(msgTypeDHCPStart); err != nil {
		return err
	}
	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, ifaceID)
	_, err := c.request(msgTypeDHCPStart, payload)
	return err
}

// DHCPStop stops the DHCP client on an interface.
func (c *Client) DHCPStop(ifaceID uint16) error {
	if err := c.checkVersion(msgTypeDHCPStop); err != nil {
		return err
	}
	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, ifaceID)
	_, err := c.request(msgTypeDHCPStop, payload)
	return err
}

func decodeDHCPStatus(data []byte) (DHCPStatus, error) {
	if len(data) < dhcpStatusSize {
		return DHCPStatus{}, fmt.Errorf("dhcp status data too short: %d bytes", len(data))
	}
	s := DHCPStatus{
		IfaceID:     binary.LittleEndian.Uint16(data[0:2]),
		State:       DHCPState(data[2]),
		LeaseTime:   binary.LittleEndian.Uint32(data[12:16]),
		RenewalTime: binary.LittleEndian.Uint32(data[16:20]),
		RebindTime:  binary.LittleEndian.Uint32(data[20:24]),
	}
	copy(s.ServerIP[:], data[4:8])
	copy(s.AssignedIP[:], data[8:12])
	return s, nil
}
