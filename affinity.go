package grout

import (
	"encoding/binary"
	"fmt"
)

// RxQueueMap represents an RX queue to CPU mapping.
type RxQueueMap struct {
	IfaceID uint16
	RxQID   uint16
	CPUID   uint16
	Enabled uint16
}

// CPUAffinity holds CPU affinity configuration.
type CPUAffinity struct {
	Control  CPUSet
	Datapath CPUSet
}

// CPUSet represents a set of CPU IDs.
type CPUSet struct {
	CPUs []uint16
}

const rxQueueMapSize = 8 // iface_id(2) + rxq_id(2) + cpu_id(2) + enabled(2)

// AffinityRxQList returns RX queue to CPU mappings.
func (c *Client) AffinityRxQList() ([]RxQueueMap, error) {
	if err := c.checkVersion(msgTypeAffinityRxQList); err != nil {
		return nil, err
	}
	results, err := c.requestStreamNoPayload(msgTypeAffinityRxQList)
	if err != nil {
		return nil, err
	}

	maps := make([]RxQueueMap, 0, len(results))
	for _, data := range results {
		m, err := decodeRxQueueMap(data)
		if err != nil {
			return nil, err
		}
		maps = append(maps, m)
	}
	return maps, nil
}

// AffinityRxQSet sets the CPU affinity for an RX queue.
func (c *Client) AffinityRxQSet(ifaceID, rxqID, cpuID uint16) error {
	if err := c.checkVersion(msgTypeAffinityRxQSet); err != nil {
		return err
	}
	payload := make([]byte, 6) // iface_id(2) + rxq_id(2) + cpu_id(2)
	binary.LittleEndian.PutUint16(payload[0:2], ifaceID)
	binary.LittleEndian.PutUint16(payload[2:4], rxqID)
	binary.LittleEndian.PutUint16(payload[4:6], cpuID)
	_, err := c.request(msgTypeAffinityRxQSet, payload)
	return err
}

// AffinityCPUGet retrieves the CPU affinity configuration.
func (c *Client) AffinityCPUGet() (*CPUAffinity, error) {
	if err := c.checkVersion(msgTypeAffinityCPUGet); err != nil {
		return nil, err
	}
	resp, err := c.requestNoPayload(msgTypeAffinityCPUGet)
	if err != nil {
		return nil, err
	}
	aff, err := decodeCPUAffinity(resp)
	if err != nil {
		return nil, err
	}
	return &aff, nil
}

// AffinityCPUSet updates the CPU affinity configuration.
func (c *Client) AffinityCPUSet(control, datapath CPUSet) error {
	if err := c.checkVersion(msgTypeAffinityCPUSet); err != nil {
		return err
	}
	payload := encodeCPUAffinity(control, datapath)
	_, err := c.request(msgTypeAffinityCPUSet, payload)
	return err
}

func decodeRxQueueMap(data []byte) (RxQueueMap, error) {
	if len(data) < rxQueueMapSize {
		return RxQueueMap{}, fmt.Errorf("rx queue map data too short: %d bytes", len(data))
	}
	return RxQueueMap{
		IfaceID: binary.LittleEndian.Uint16(data[0:2]),
		RxQID:   binary.LittleEndian.Uint16(data[2:4]),
		CPUID:   binary.LittleEndian.Uint16(data[4:6]),
		Enabled: binary.LittleEndian.Uint16(data[6:8]),
	}, nil
}

func decodeCPUAffinity(data []byte) (CPUAffinity, error) {
	if len(data) < 4 {
		return CPUAffinity{}, fmt.Errorf("cpu affinity data too short: %d bytes", len(data))
	}
	ctrlCount := binary.LittleEndian.Uint16(data[0:2])
	dpCount := binary.LittleEndian.Uint16(data[2:4])

	needed := 4 + int(ctrlCount)*2 + int(dpCount)*2
	if len(data) < needed {
		return CPUAffinity{}, fmt.Errorf("cpu affinity data too short for counts: %d bytes", len(data))
	}

	aff := CPUAffinity{}
	off := 4
	aff.Control.CPUs = make([]uint16, ctrlCount)
	for i := range ctrlCount {
		aff.Control.CPUs[i] = binary.LittleEndian.Uint16(data[off : off+2])
		off += 2
	}
	aff.Datapath.CPUs = make([]uint16, dpCount)
	for i := range dpCount {
		aff.Datapath.CPUs[i] = binary.LittleEndian.Uint16(data[off : off+2])
		off += 2
	}
	return aff, nil
}

func encodeCPUAffinity(control, datapath CPUSet) []byte {
	buf := make([]byte, 4+len(control.CPUs)*2+len(datapath.CPUs)*2)
	binary.LittleEndian.PutUint16(buf[0:2], uint16(len(control.CPUs)))
	binary.LittleEndian.PutUint16(buf[2:4], uint16(len(datapath.CPUs)))
	off := 4
	for _, cpu := range control.CPUs {
		binary.LittleEndian.PutUint16(buf[off:off+2], cpu)
		off += 2
	}
	for _, cpu := range datapath.CPUs {
		binary.LittleEndian.PutUint16(buf[off:off+2], cpu)
		off += 2
	}
	return buf
}
