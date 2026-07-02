package grout

import (
	"encoding/binary"
	"fmt"
)

// ConnState represents a connection tracking state.
type ConnState uint8

const (
	ConnStateClosed      ConnState = 0
	ConnStateNew         ConnState = 1
	ConnStateSimSynSent  ConnState = 2
	ConnStateSynReceived ConnState = 3
	ConnStateEstablished ConnState = 4
	ConnStateCloseWait   ConnState = 5
	ConnStateLastAck     ConnState = 6
	ConnStateFinWait     ConnState = 7
	ConnStateClosing     ConnState = 8
	ConnStateTimeWait    ConnState = 9
)

// ConntrackFlow represents a flow 5-tuple.
type ConntrackFlow struct {
	Src   IP4Addr
	Dst   IP4Addr
	SrcID uint16
	DstID uint16
}

// ConntrackEntry represents a connection tracking entry.
type ConntrackEntry struct {
	IfaceID    uint16
	Family     AddrFamily
	Proto      uint8
	FwdFlow    ConntrackFlow
	RevFlow    ConntrackFlow
	LastUpdate ClockNS
	ID         uint32
	State      ConnState
}

// ConntrackConfig holds connection tracking configuration.
type ConntrackConfig struct {
	MaxCount               uint32
	TimeoutClosed          uint32
	TimeoutNew             uint32
	TimeoutUDPEstablished  uint32
	TimeoutTCPEstablished  uint32
	TimeoutHalfClose       uint32
	TimeoutTimeWait        uint32
}

const (
	conntrackFlowSize   = 12 // src(4) + dst(4) + src_id(2) + dst_id(2)
	conntrackEntrySize  = 44 // iface_id(2) + family(1) + proto(1) + 2*flow(24) + last_update(8) + id(4) + state(1) + pad(3)
	conntrackConfigSize = 28 // 7 x uint32
)

// ConntrackList returns all connection tracking entries.
func (c *Client) ConntrackList() ([]ConntrackEntry, error) {
	if err := c.checkVersion(msgTypeConntrackList); err != nil {
		return nil, err
	}
	results, err := c.requestStreamNoPayload(msgTypeConntrackList)
	if err != nil {
		return nil, err
	}

	entries := make([]ConntrackEntry, 0, len(results))
	for _, data := range results {
		e, err := decodeConntrackEntry(data)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// ConntrackFlush removes all connection tracking entries.
func (c *Client) ConntrackFlush() error {
	if err := c.checkVersion(msgTypeConntrackFlush); err != nil {
		return err
	}
	_, err := c.requestNoPayload(msgTypeConntrackFlush)
	return err
}

// ConntrackConfigGet retrieves the connection tracking configuration.
func (c *Client) ConntrackConfigGet() (*ConntrackConfig, error) {
	if err := c.checkVersion(msgTypeConntrackConfGet); err != nil {
		return nil, err
	}
	resp, err := c.requestNoPayload(msgTypeConntrackConfGet)
	if err != nil {
		return nil, err
	}
	cfg, err := decodeConntrackConfig(resp)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ConntrackConfigSet updates the connection tracking configuration.
func (c *Client) ConntrackConfigSet(config ConntrackConfig) error {
	if err := c.checkVersion(msgTypeConntrackConfSet); err != nil {
		return err
	}
	_, err := c.request(msgTypeConntrackConfSet, encodeConntrackConfig(config))
	return err
}

func decodeConntrackFlow(data []byte) ConntrackFlow {
	f := ConntrackFlow{}
	copy(f.Src[:], data[0:4])
	copy(f.Dst[:], data[4:8])
	f.SrcID = binary.LittleEndian.Uint16(data[8:10])
	f.DstID = binary.LittleEndian.Uint16(data[10:12])
	return f
}

func decodeConntrackEntry(data []byte) (ConntrackEntry, error) {
	if len(data) < conntrackEntrySize {
		return ConntrackEntry{}, fmt.Errorf("conntrack entry data too short: %d bytes", len(data))
	}
	e := ConntrackEntry{
		IfaceID: binary.LittleEndian.Uint16(data[0:2]),
		Family:  AddrFamily(data[2]),
		Proto:   data[3],
		FwdFlow: decodeConntrackFlow(data[4:16]),
		RevFlow: decodeConntrackFlow(data[16:28]),
	}
	e.LastUpdate = ClockNS(binary.LittleEndian.Uint64(data[28:36]))
	e.ID = binary.LittleEndian.Uint32(data[36:40])
	e.State = ConnState(data[40])
	return e, nil
}

func decodeConntrackConfig(data []byte) (ConntrackConfig, error) {
	if len(data) < conntrackConfigSize {
		return ConntrackConfig{}, fmt.Errorf("conntrack config data too short: %d bytes", len(data))
	}
	return ConntrackConfig{
		MaxCount:              binary.LittleEndian.Uint32(data[0:4]),
		TimeoutClosed:         binary.LittleEndian.Uint32(data[4:8]),
		TimeoutNew:            binary.LittleEndian.Uint32(data[8:12]),
		TimeoutUDPEstablished: binary.LittleEndian.Uint32(data[12:16]),
		TimeoutTCPEstablished: binary.LittleEndian.Uint32(data[16:20]),
		TimeoutHalfClose:      binary.LittleEndian.Uint32(data[20:24]),
		TimeoutTimeWait:       binary.LittleEndian.Uint32(data[24:28]),
	}, nil
}

func encodeConntrackConfig(cfg ConntrackConfig) []byte {
	buf := make([]byte, conntrackConfigSize)
	binary.LittleEndian.PutUint32(buf[0:4], cfg.MaxCount)
	binary.LittleEndian.PutUint32(buf[4:8], cfg.TimeoutClosed)
	binary.LittleEndian.PutUint32(buf[8:12], cfg.TimeoutNew)
	binary.LittleEndian.PutUint32(buf[12:16], cfg.TimeoutUDPEstablished)
	binary.LittleEndian.PutUint32(buf[16:20], cfg.TimeoutTCPEstablished)
	binary.LittleEndian.PutUint32(buf[20:24], cfg.TimeoutHalfClose)
	binary.LittleEndian.PutUint32(buf[24:28], cfg.TimeoutTimeWait)
	return buf
}
