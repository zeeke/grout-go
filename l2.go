package grout

import (
	"encoding/binary"
	"fmt"
)

// FDBFlags represents FDB entry flags.
type FDBFlags uint8

const (
	FDBFlagStatic FDBFlags = 1 << 0
	FDBFlagLearn  FDBFlags = 1 << 1
	FDBFlagExtern FDBFlags = 1 << 2
)

// FDBEntry represents a forwarding database entry.
type FDBEntry struct {
	BridgeID uint16
	MAC      EtherAddr
	VLANID   uint16
	IfaceID  uint16
	VTEP     L3Addr
	Flags    FDBFlags
	LastSeen ClockNS
}

// FDBConfig holds FDB table configuration.
type FDBConfig struct {
	MaxEntries  uint32
	UsedEntries uint32
}

// FDBFlushRequest describes parameters for flushing FDB entries.
type FDBFlushRequest struct {
	BridgeID uint16
	IfaceID  uint16
}

// FDBListRequest describes parameters for listing FDB entries.
type FDBListRequest struct {
	BridgeID uint16
}

// FloodType identifies the flood entry type.
type FloodType uint8

const (
	FloodTypeVTEP FloodType = 1
)

// FloodEntry represents a flood list entry.
type FloodEntry struct {
	Type  FloodType
	VRFID uint16
	VNI   uint32
	Addr  L3Addr
}

const (
	fdbEntrySize  = 48 // bridge_id(2) + mac(6) + vlan_id(2) + iface_id(2) + vtep(20) + flags(1) + pad(7) + last_seen(8)
	fdbConfigSize = 8  // max(4) + used(4)
)

// FDBAdd adds an FDB entry.
func (c *Client) FDBAdd(entry FDBEntry, existOK bool) error {
	if err := c.checkVersion(msgTypeFDBAdd); err != nil {
		return err
	}
	payload := encodeFDBEntry(entry, existOK)
	_, err := c.request(msgTypeFDBAdd, payload)
	return err
}

// FDBDel removes an FDB entry.
func (c *Client) FDBDel(bridgeID uint16, mac EtherAddr, vlanID uint16, missingOK bool) error {
	if err := c.checkVersion(msgTypeFDBDel); err != nil {
		return err
	}
	payload := make([]byte, 12) // bridge_id(2) + mac(6) + vlan_id(2) + flag(2)
	binary.LittleEndian.PutUint16(payload[0:2], bridgeID)
	copy(payload[2:8], mac[:])
	binary.LittleEndian.PutUint16(payload[8:10], vlanID)
	if missingOK {
		payload[10] = 1
	}
	_, err := c.request(msgTypeFDBDel, payload)
	return err
}

// FDBFlush removes FDB entries matching the given criteria.
func (c *Client) FDBFlush(req FDBFlushRequest) error {
	if err := c.checkVersion(msgTypeFDBFlush); err != nil {
		return err
	}
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint16(payload[0:2], req.BridgeID)
	binary.LittleEndian.PutUint16(payload[2:4], req.IfaceID)
	_, err := c.request(msgTypeFDBFlush, payload)
	return err
}

// FDBList returns FDB entries for a bridge.
func (c *Client) FDBList(req FDBListRequest) ([]FDBEntry, error) {
	if err := c.checkVersion(msgTypeFDBList); err != nil {
		return nil, err
	}
	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, req.BridgeID)

	results, err := c.requestStream(msgTypeFDBList, payload)
	if err != nil {
		return nil, err
	}

	entries := make([]FDBEntry, 0, len(results))
	for _, data := range results {
		e, err := decodeFDBEntry(data)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// FDBConfigGet retrieves FDB configuration.
func (c *Client) FDBConfigGet() (*FDBConfig, error) {
	if err := c.checkVersion(msgTypeFDBConfigGet); err != nil {
		return nil, err
	}
	resp, err := c.requestNoPayload(msgTypeFDBConfigGet)
	if err != nil {
		return nil, err
	}
	if len(resp) < fdbConfigSize {
		return nil, fmt.Errorf("fdb config data too short: %d bytes", len(resp))
	}
	cfg := &FDBConfig{
		MaxEntries:  binary.LittleEndian.Uint32(resp[0:4]),
		UsedEntries: binary.LittleEndian.Uint32(resp[4:8]),
	}
	return cfg, nil
}

// FDBConfigSet updates FDB configuration.
func (c *Client) FDBConfigSet(maxEntries uint32) error {
	if err := c.checkVersion(msgTypeFDBConfigSet); err != nil {
		return err
	}
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, maxEntries)
	_, err := c.request(msgTypeFDBConfigSet, payload)
	return err
}

// FloodAdd adds a flood entry.
func (c *Client) FloodAdd(entry FloodEntry, existOK bool) error {
	if err := c.checkVersion(msgTypeFloodAdd); err != nil {
		return err
	}
	payload := encodeFloodEntry(entry, existOK)
	_, err := c.request(msgTypeFloodAdd, payload)
	return err
}

// FloodDel removes a flood entry.
func (c *Client) FloodDel(entry FloodEntry, missingOK bool) error {
	if err := c.checkVersion(msgTypeFloodDel); err != nil {
		return err
	}
	payload := encodeFloodEntry(entry, missingOK)
	_, err := c.request(msgTypeFloodDel, payload)
	return err
}

// FloodList returns flood entries.
func (c *Client) FloodList(floodType FloodType, vrfID uint16) ([]FloodEntry, error) {
	if err := c.checkVersion(msgTypeFloodList); err != nil {
		return nil, err
	}
	payload := make([]byte, 4)
	payload[0] = byte(floodType)
	binary.LittleEndian.PutUint16(payload[2:4], vrfID)

	results, err := c.requestStream(msgTypeFloodList, payload)
	if err != nil {
		return nil, err
	}

	entries := make([]FloodEntry, 0, len(results))
	for _, data := range results {
		e, err := decodeFloodEntry(data)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func encodeFDBEntry(e FDBEntry, flag bool) []byte {
	buf := make([]byte, fdbEntrySize)
	binary.LittleEndian.PutUint16(buf[0:2], e.BridgeID)
	copy(buf[2:8], e.MAC[:])
	binary.LittleEndian.PutUint16(buf[8:10], e.VLANID)
	binary.LittleEndian.PutUint16(buf[10:12], e.IfaceID)
	encodeL3Addr(buf[12:12+l3AddrWireSize], e.VTEP)
	buf[12+l3AddrWireSize] = byte(e.Flags)
	if flag {
		buf[12+l3AddrWireSize+1] = 1
	}
	binary.LittleEndian.PutUint64(buf[40:48], uint64(e.LastSeen))
	return buf
}

func decodeFDBEntry(data []byte) (FDBEntry, error) {
	if len(data) < fdbEntrySize {
		return FDBEntry{}, fmt.Errorf("fdb entry data too short: %d bytes", len(data))
	}
	e := FDBEntry{
		BridgeID: binary.LittleEndian.Uint16(data[0:2]),
		VLANID:   binary.LittleEndian.Uint16(data[8:10]),
		IfaceID:  binary.LittleEndian.Uint16(data[10:12]),
		VTEP:     decodeL3Addr(data[12 : 12+l3AddrWireSize]),
		Flags:    FDBFlags(data[12+l3AddrWireSize]),
		LastSeen: ClockNS(binary.LittleEndian.Uint64(data[40:48])),
	}
	copy(e.MAC[:], data[2:8])
	return e, nil
}

func encodeFloodEntry(e FloodEntry, flag bool) []byte {
	buf := make([]byte, 4+4+l3AddrWireSize)
	buf[0] = byte(e.Type)
	if flag {
		buf[1] = 1
	}
	binary.LittleEndian.PutUint16(buf[2:4], e.VRFID)
	binary.LittleEndian.PutUint32(buf[4:8], e.VNI)
	encodeL3Addr(buf[8:8+l3AddrWireSize], e.Addr)
	return buf
}

func decodeFloodEntry(data []byte) (FloodEntry, error) {
	minSize := 8 + l3AddrWireSize
	if len(data) < minSize {
		return FloodEntry{}, fmt.Errorf("flood entry data too short: %d bytes", len(data))
	}
	return FloodEntry{
		Type:  FloodType(data[0]),
		VRFID: binary.LittleEndian.Uint16(data[2:4]),
		VNI:   binary.LittleEndian.Uint32(data[4:8]),
		Addr:  decodeL3Addr(data[8 : 8+l3AddrWireSize]),
	}, nil
}
