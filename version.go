package grout

import "fmt"

// minVersion maps message types to their minimum required API version.
var minVersion = map[uint32]uint32{
	// All current messages require API version 3.
	msgTypeHello:            3,
	msgTypeLogPacketsSet:    3,
	msgTypeLogLevelList:     3,
	msgTypeLogLevelSet:      3,
	msgTypeEventSubscribe:   3,
	msgTypeEventUnsubscribe: 3,
	msgTypeIfaceAdd:         3,
	msgTypeIfaceDel:         3,
	msgTypeIfaceGet:         3,
	msgTypeIfaceList:        3,
	msgTypeIfaceSet:         3,
	msgTypeIfaceStatsGet:    3,
	msgTypeIfaceMACAdd:      3,
	msgTypeIfaceMACDel:      3,
	msgTypeIfaceMACList:     3,
	msgTypeIfaceMACSet:      3,
	msgTypeAffinityRxQList:  3,
	msgTypeAffinityRxQSet:   3,
	msgTypeAffinityCPUGet:   3,
	msgTypeAffinityCPUSet:   3,
	msgTypeStatsGet:         3,
	msgTypeStatsReset:       3,
	msgTypeGraphDump:        3,
	msgTypeGraphConfGet:     3,
	msgTypeGraphConfSet:     3,
	msgTypePktTraceSet:      3,
	msgTypePktTraceClear:    3,
	msgTypePktTraceDump:     3,
	msgTypeNHAdd:            3,
	msgTypeNHDel:            3,
	msgTypeNHGet:            3,
	msgTypeNHList:           3,
	msgTypeNHConfigGet:      3,
	msgTypeNHConfigSet:      3,
	msgTypeIP4RouteAdd:      3,
	msgTypeIP4RouteDel:      3,
	msgTypeIP4RouteGet:      3,
	msgTypeIP4RouteList:     3,
	msgTypeIP4AddrAdd:       3,
	msgTypeIP4AddrDel:       3,
	msgTypeIP4AddrList:      3,
	msgTypeIP4AddrFlush:     3,
	msgTypeIP4ICMPSend:      3,
	msgTypeIP4ICMPRecv:      3,
	msgTypeIP4FIBDefaultSet: 3,
	msgTypeIP4FIBInfoList:   3,
	msgTypeIP6RouteAdd:      3,
	msgTypeIP6RouteDel:      3,
	msgTypeIP6RouteGet:      3,
	msgTypeIP6RouteList:     3,
	msgTypeIP6AddrAdd:       3,
	msgTypeIP6AddrDel:       3,
	msgTypeIP6AddrList:      3,
	msgTypeIP6AddrFlush:     3,
	msgTypeIP6ICMPSend:      3,
	msgTypeIP6ICMPRecv:      3,
	msgTypeIP6FIBDefaultSet: 3,
	msgTypeIP6FIBInfoList:   3,
	msgTypeIP6RASet:         3,
	msgTypeIP6RAClear:       3,
	msgTypeIP6RAShow:        3,
	msgTypeFDBAdd:           3,
	msgTypeFDBDel:           3,
	msgTypeFDBFlush:         3,
	msgTypeFDBList:          3,
	msgTypeFDBConfigGet:     3,
	msgTypeFDBConfigSet:     3,
	msgTypeFloodAdd:         3,
	msgTypeFloodDel:         3,
	msgTypeFloodList:        3,
	msgTypeDHCPList:         3,
	msgTypeDHCPStart:        3,
	msgTypeDHCPStop:         3,
	msgTypeConntrackList:    3,
	msgTypeConntrackFlush:   3,
	msgTypeConntrackConfGet: 3,
	msgTypeConntrackConfSet: 3,
	msgTypeDNAT44Add:        3,
	msgTypeDNAT44Del:        3,
	msgTypeDNAT44List:       3,
	msgTypeSNAT44Add:        3,
	msgTypeSNAT44Del:        3,
	msgTypeSNAT44List:       3,
	msgTypeSRv6TunSrcSet:    3,
	msgTypeSRv6TunSrcClear:  3,
	msgTypeSRv6TunSrcShow:   3,
}

// checkVersion returns ErrUnsupported if the connected API version
// does not support the given message type.
func (c *Client) checkVersion(msgT uint32) error {
	required, ok := minVersion[msgT]
	if !ok {
		return nil
	}
	if c.apiVersion < required {
		return fmt.Errorf("%w: requires API version %d, connected to %d",
			ErrUnsupported, required, c.apiVersion)
	}
	return nil
}
