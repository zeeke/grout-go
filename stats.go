package grout

import (
	"encoding/binary"
	"fmt"
)

// StatsFlags controls which statistics to return.
type StatsFlags uint32

const (
	StatsFlagSW   StatsFlags = 1 << 0
	StatsFlagHW   StatsFlags = 1 << 1
	StatsFlagZero StatsFlags = 1 << 2
)

// Stat represents a graph node statistics entry.
type Stat struct {
	Name      string
	TopoOrder uint64
	Packets   uint64
	Batches   uint64
	Cycles    uint64
}

// StatsGetRequest describes parameters for getting statistics.
type StatsGetRequest struct {
	Flags StatsFlags
}

// GraphConf holds graph runtime configuration.
type GraphConf struct {
	RxBurstMax uint16
	VectorMax  uint16
}

// GraphDumpRequest describes parameters for dumping the graph.
type GraphDumpRequest struct {
	Format uint32
}

const (
	statNameLen  = 64
	statSize     = statNameLen + 32 // name(64) + topo_order(8) + packets(8) + batches(8) + cycles(8)
	graphConfSize = 4               // rx_burst_max(2) + vector_max(2)
)

// StatsGet returns graph node statistics.
func (c *Client) StatsGet(req StatsGetRequest) ([]Stat, error) {
	if err := c.checkVersion(msgTypeStatsGet); err != nil {
		return nil, err
	}
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, uint32(req.Flags))

	results, err := c.requestStream(msgTypeStatsGet, payload)
	if err != nil {
		return nil, err
	}

	stats := make([]Stat, 0, len(results))
	for _, data := range results {
		s, err := decodeStat(data)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// StatsReset resets all statistics counters.
func (c *Client) StatsReset() error {
	if err := c.checkVersion(msgTypeStatsReset); err != nil {
		return err
	}
	_, err := c.requestNoPayload(msgTypeStatsReset)
	return err
}

// GraphDump returns a text dump of the packet processing graph.
func (c *Client) GraphDump(req GraphDumpRequest) (string, error) {
	if err := c.checkVersion(msgTypeGraphDump); err != nil {
		return "", err
	}
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, req.Format)

	resp, err := c.request(msgTypeGraphDump, payload)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// GraphConfigGet retrieves graph runtime configuration.
func (c *Client) GraphConfigGet() (*GraphConf, error) {
	if err := c.checkVersion(msgTypeGraphConfGet); err != nil {
		return nil, err
	}
	resp, err := c.requestNoPayload(msgTypeGraphConfGet)
	if err != nil {
		return nil, err
	}
	if len(resp) < graphConfSize {
		return nil, fmt.Errorf("graph conf data too short: %d bytes", len(resp))
	}
	return &GraphConf{
		RxBurstMax: binary.LittleEndian.Uint16(resp[0:2]),
		VectorMax:  binary.LittleEndian.Uint16(resp[2:4]),
	}, nil
}

// GraphConfigSet updates graph runtime configuration.
func (c *Client) GraphConfigSet(conf GraphConf) error {
	if err := c.checkVersion(msgTypeGraphConfSet); err != nil {
		return err
	}
	payload := make([]byte, graphConfSize)
	binary.LittleEndian.PutUint16(payload[0:2], conf.RxBurstMax)
	binary.LittleEndian.PutUint16(payload[2:4], conf.VectorMax)
	_, err := c.request(msgTypeGraphConfSet, payload)
	return err
}

func decodeStat(data []byte) (Stat, error) {
	if len(data) < statSize {
		return Stat{}, fmt.Errorf("stat data too short: %d bytes", len(data))
	}
	return Stat{
		Name:      cstring(data[0:statNameLen]),
		TopoOrder: binary.LittleEndian.Uint64(data[statNameLen : statNameLen+8]),
		Packets:   binary.LittleEndian.Uint64(data[statNameLen+8 : statNameLen+16]),
		Batches:   binary.LittleEndian.Uint64(data[statNameLen+16 : statNameLen+24]),
		Cycles:    binary.LittleEndian.Uint64(data[statNameLen+24 : statNameLen+32]),
	}, nil
}
