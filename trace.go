package grout

import "encoding/binary"

// PacketTraceSet enables or disables packet tracing on an interface.
func (c *Client) PacketTraceSet(ifaceID uint16, enabled, all bool) error {
	if err := c.checkVersion(msgTypePktTraceSet); err != nil {
		return err
	}
	payload := make([]byte, 4) // iface_id(2) + enabled(1) + all(1)
	binary.LittleEndian.PutUint16(payload[0:2], ifaceID)
	if enabled {
		payload[2] = 1
	}
	if all {
		payload[3] = 1
	}
	_, err := c.request(msgTypePktTraceSet, payload)
	return err
}

// PacketTraceClear clears all packet trace data.
func (c *Client) PacketTraceClear() error {
	if err := c.checkVersion(msgTypePktTraceClear); err != nil {
		return err
	}
	_, err := c.requestNoPayload(msgTypePktTraceClear)
	return err
}

// PacketTraceDump returns captured packet trace data.
func (c *Client) PacketTraceDump(maxPackets uint16) ([]byte, error) {
	if err := c.checkVersion(msgTypePktTraceDump); err != nil {
		return nil, err
	}
	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, maxPackets)
	return c.request(msgTypePktTraceDump, payload)
}
