package grout

import (
	"encoding/binary"
	"fmt"
)

// DNAT44Policy represents a static DNAT44 rule.
type DNAT44Policy struct {
	IfaceID uint16
	Match   IP4Addr
	Replace IP4Addr
}

// SNAT44Policy represents a dynamic SNAT44 rule.
type SNAT44Policy struct {
	IfaceID uint16
	Net     IP4Net
	Replace IP4Addr
}

const (
	dnat44Size = 12 // iface_id(2) + pad(2) + match(4) + replace(4)
	snat44Size = 16 // iface_id(2) + pad(1) + prefix_len(1) + net(4) + replace(4) + pad(4)
)

// DNAT44Add adds a static DNAT44 rule.
func (c *Client) DNAT44Add(policy DNAT44Policy, existOK bool) error {
	if err := c.checkVersion(msgTypeDNAT44Add); err != nil {
		return err
	}
	payload := encodeDNAT44(policy, existOK)
	_, err := c.request(msgTypeDNAT44Add, payload)
	return err
}

// DNAT44Del deletes a DNAT44 rule.
func (c *Client) DNAT44Del(ifaceID uint16, match IP4Addr, missingOK bool) error {
	if err := c.checkVersion(msgTypeDNAT44Del); err != nil {
		return err
	}
	payload := make([]byte, 8)
	binary.LittleEndian.PutUint16(payload[0:2], ifaceID)
	copy(payload[4:8], match[:])
	_, err := c.request(msgTypeDNAT44Del, payload)
	return err
}

// DNAT44List returns DNAT44 rules for a VRF.
func (c *Client) DNAT44List(vrfID uint16) ([]DNAT44Policy, error) {
	if err := c.checkVersion(msgTypeDNAT44List); err != nil {
		return nil, err
	}
	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, vrfID)

	results, err := c.requestStream(msgTypeDNAT44List, payload)
	if err != nil {
		return nil, err
	}

	policies := make([]DNAT44Policy, 0, len(results))
	for _, data := range results {
		p, err := decodeDNAT44(data)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, nil
}

// SNAT44Add adds a dynamic SNAT44 rule.
func (c *Client) SNAT44Add(policy SNAT44Policy, existOK bool) error {
	if err := c.checkVersion(msgTypeSNAT44Add); err != nil {
		return err
	}
	payload := encodeSNAT44(policy, existOK)
	_, err := c.request(msgTypeSNAT44Add, payload)
	return err
}

// SNAT44Del deletes a SNAT44 rule.
func (c *Client) SNAT44Del(policy SNAT44Policy, missingOK bool) error {
	if err := c.checkVersion(msgTypeSNAT44Del); err != nil {
		return err
	}
	payload := encodeSNAT44(policy, missingOK)
	_, err := c.request(msgTypeSNAT44Del, payload)
	return err
}

// SNAT44List returns all SNAT44 rules.
func (c *Client) SNAT44List() ([]SNAT44Policy, error) {
	if err := c.checkVersion(msgTypeSNAT44List); err != nil {
		return nil, err
	}
	results, err := c.requestStreamNoPayload(msgTypeSNAT44List)
	if err != nil {
		return nil, err
	}

	policies := make([]SNAT44Policy, 0, len(results))
	for _, data := range results {
		p, err := decodeSNAT44(data)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, nil
}

func encodeDNAT44(p DNAT44Policy, flag bool) []byte {
	buf := make([]byte, dnat44Size)
	binary.LittleEndian.PutUint16(buf[0:2], p.IfaceID)
	if flag {
		buf[2] = 1
	}
	copy(buf[4:8], p.Match[:])
	copy(buf[8:12], p.Replace[:])
	return buf
}

func decodeDNAT44(data []byte) (DNAT44Policy, error) {
	if len(data) < dnat44Size {
		return DNAT44Policy{}, fmt.Errorf("dnat44 data too short: %d bytes", len(data))
	}
	p := DNAT44Policy{
		IfaceID: binary.LittleEndian.Uint16(data[0:2]),
	}
	copy(p.Match[:], data[4:8])
	copy(p.Replace[:], data[8:12])
	return p, nil
}

func encodeSNAT44(p SNAT44Policy, flag bool) []byte {
	buf := make([]byte, snat44Size)
	binary.LittleEndian.PutUint16(buf[0:2], p.IfaceID)
	if flag {
		buf[2] = 1
	}
	buf[3] = p.Net.PrefixLen
	copy(buf[4:8], p.Net.Addr[:])
	copy(buf[8:12], p.Replace[:])
	return buf
}

func decodeSNAT44(data []byte) (SNAT44Policy, error) {
	if len(data) < 12 {
		return SNAT44Policy{}, fmt.Errorf("snat44 data too short: %d bytes", len(data))
	}
	p := SNAT44Policy{
		IfaceID: binary.LittleEndian.Uint16(data[0:2]),
	}
	p.Net.PrefixLen = data[3]
	copy(p.Net.Addr[:], data[4:8])
	copy(p.Replace[:], data[8:12])
	return p, nil
}
