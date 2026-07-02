package grout

import (
	"encoding/binary"
	"fmt"
)

// IP4IfAddr represents an IPv4 interface address.
type IP4IfAddr struct {
	IfaceID uint16
	Addr    IP4Net
}

// IP4Route represents an IPv4 route with its nexthop.
type IP4Route struct {
	Dest   IP4Net
	VRFID  uint16
	Origin NHOrigin
	NH     Nexthop
}

// IP4RouteAddRequest describes parameters for adding an IPv4 route.
type IP4RouteAddRequest struct {
	VRFID  uint16
	Dest   IP4Net
	NH     IP4Addr
	Origin NHOrigin
}

// IP4PingRequest describes parameters for an IPv4 ICMP ping.
type IP4PingRequest struct {
	Addr   IP4Addr
	VRFID  uint16
	SeqNum uint16
	TTL    uint8
}

// ICMPRecvResp holds the response from an IPv4 ping.
type ICMPRecvResp struct {
	Type         uint8
	Code         uint8
	TTL          uint8
	Ident        uint16
	SeqNum       uint16
	SrcAddr      IP4Addr
	ResponseTime ClockNS
}

// FIB4Info holds IPv4 FIB table information.
type FIB4Info struct {
	VRFID      uint16
	MaxRoutes  uint32
	UsedRoutes uint32
	NumTbl8    uint32
	UsedTbl8   uint32
}

// Wire sizes.
const (
	ip4IfAddrSize = 8  // iface_id(2) + pad(1) + prefix_len(1) + addr(4)
	ip4RouteReqSize = 12 // vrf_id(2) + pad(1) + prefix_len(1) + dest(4) + nh(4)
	fib4InfoSize = 20  // vrf_id(2) + pad(2) + 4 x uint32
)

// IP4AddrAdd adds an IPv4 address to an interface.
func (c *Client) IP4AddrAdd(addr IP4IfAddr, existOK bool) error {
	if err := c.checkVersion(msgTypeIP4AddrAdd); err != nil {
		return err
	}
	payload := encodeIP4IfAddr(addr, existOK)
	_, err := c.request(msgTypeIP4AddrAdd, payload)
	return err
}

// IP4AddrDel removes an IPv4 address from an interface.
func (c *Client) IP4AddrDel(addr IP4IfAddr, missingOK bool) error {
	if err := c.checkVersion(msgTypeIP4AddrDel); err != nil {
		return err
	}
	payload := encodeIP4IfAddr(addr, missingOK)
	_, err := c.request(msgTypeIP4AddrDel, payload)
	return err
}

// IP4AddrList returns IPv4 addresses, optionally filtered by VRF and interface.
func (c *Client) IP4AddrList(vrfID, ifaceID uint16) ([]IP4IfAddr, error) {
	if err := c.checkVersion(msgTypeIP4AddrList); err != nil {
		return nil, err
	}
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint16(payload[0:2], vrfID)
	binary.LittleEndian.PutUint16(payload[2:4], ifaceID)

	results, err := c.requestStream(msgTypeIP4AddrList, payload)
	if err != nil {
		return nil, err
	}

	addrs := make([]IP4IfAddr, 0, len(results))
	for _, data := range results {
		a, err := decodeIP4IfAddr(data)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, a)
	}
	return addrs, nil
}

// IP4AddrFlush removes all IPv4 addresses from an interface.
func (c *Client) IP4AddrFlush(ifaceID uint16) error {
	if err := c.checkVersion(msgTypeIP4AddrFlush); err != nil {
		return err
	}
	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, ifaceID)
	_, err := c.request(msgTypeIP4AddrFlush, payload)
	return err
}

// IP4RouteAdd adds an IPv4 route.
func (c *Client) IP4RouteAdd(req IP4RouteAddRequest) error {
	if err := c.checkVersion(msgTypeIP4RouteAdd); err != nil {
		return err
	}
	payload := encodeIP4RouteAdd(req)
	_, err := c.request(msgTypeIP4RouteAdd, payload)
	return err
}

// IP4RouteDel deletes an IPv4 route.
func (c *Client) IP4RouteDel(vrfID uint16, dest IP4Net, missingOK bool) error {
	if err := c.checkVersion(msgTypeIP4RouteDel); err != nil {
		return err
	}
	payload := make([]byte, 8)
	binary.LittleEndian.PutUint16(payload[0:2], vrfID)
	payload[3] = dest.PrefixLen
	copy(payload[4:8], dest.Addr[:])
	_, err := c.request(msgTypeIP4RouteDel, payload)
	return err
}

// IP4RouteGet looks up a route for a destination address.
func (c *Client) IP4RouteGet(vrfID uint16, dest IP4Addr) (*Nexthop, error) {
	if err := c.checkVersion(msgTypeIP4RouteGet); err != nil {
		return nil, err
	}
	payload := make([]byte, 6)
	binary.LittleEndian.PutUint16(payload[0:2], vrfID)
	copy(payload[2:6], dest[:])
	resp, err := c.request(msgTypeIP4RouteGet, payload)
	if err != nil {
		return nil, err
	}
	nh, err := decodeNexthop(resp)
	if err != nil {
		return nil, err
	}
	return &nh, nil
}

// IP4RouteList returns IPv4 routes for a VRF.
func (c *Client) IP4RouteList(vrfID uint16, maxCount uint16) ([]IP4Route, error) {
	if err := c.checkVersion(msgTypeIP4RouteList); err != nil {
		return nil, err
	}
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint16(payload[0:2], vrfID)
	binary.LittleEndian.PutUint16(payload[2:4], maxCount)

	results, err := c.requestStream(msgTypeIP4RouteList, payload)
	if err != nil {
		return nil, err
	}

	routes := make([]IP4Route, 0, len(results))
	for _, data := range results {
		r, err := decodeIP4Route(data)
		if err != nil {
			return nil, err
		}
		routes = append(routes, r)
	}
	return routes, nil
}

// IP4FIBDefaultSet sets the default FIB table size.
func (c *Client) IP4FIBDefaultSet(maxRoutes uint32) error {
	if err := c.checkVersion(msgTypeIP4FIBDefaultSet); err != nil {
		return err
	}
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, maxRoutes)
	_, err := c.request(msgTypeIP4FIBDefaultSet, payload)
	return err
}

// IP4FIBInfoList returns FIB table information for a VRF.
func (c *Client) IP4FIBInfoList(vrfID uint16) ([]FIB4Info, error) {
	if err := c.checkVersion(msgTypeIP4FIBInfoList); err != nil {
		return nil, err
	}
	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, vrfID)

	results, err := c.requestStream(msgTypeIP4FIBInfoList, payload)
	if err != nil {
		return nil, err
	}

	infos := make([]FIB4Info, 0, len(results))
	for _, data := range results {
		info, err := decodeFIB4Info(data)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// IP4Ping sends an ICMP echo request and waits for the reply.
func (c *Client) IP4Ping(req IP4PingRequest) (*ICMPRecvResp, error) {
	if err := c.checkVersion(msgTypeIP4ICMPSend); err != nil {
		return nil, err
	}
	payload := encodeIP4PingReq(req)
	_, err := c.request(msgTypeIP4ICMPSend, payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.requestNoPayload(msgTypeIP4ICMPRecv)
	if err != nil {
		return nil, err
	}

	icmp, err := decodeICMPRecvResp(resp)
	if err != nil {
		return nil, err
	}
	return &icmp, nil
}

// Encoding/decoding helpers.

func encodeIP4IfAddr(addr IP4IfAddr, flag bool) []byte {
	buf := make([]byte, ip4IfAddrSize)
	binary.LittleEndian.PutUint16(buf[0:2], addr.IfaceID)
	if flag {
		buf[2] = 1
	}
	buf[3] = addr.Addr.PrefixLen
	copy(buf[4:8], addr.Addr.Addr[:])
	return buf
}

func decodeIP4IfAddr(data []byte) (IP4IfAddr, error) {
	if len(data) < ip4IfAddrSize {
		return IP4IfAddr{}, fmt.Errorf("ip4 ifaddr data too short: %d bytes", len(data))
	}
	a := IP4IfAddr{
		IfaceID: binary.LittleEndian.Uint16(data[0:2]),
	}
	a.Addr.PrefixLen = data[3]
	copy(a.Addr.Addr[:], data[4:8])
	return a, nil
}

func encodeIP4RouteAdd(req IP4RouteAddRequest) []byte {
	buf := make([]byte, ip4RouteReqSize)
	binary.LittleEndian.PutUint16(buf[0:2], req.VRFID)
	buf[2] = byte(req.Origin)
	buf[3] = req.Dest.PrefixLen
	copy(buf[4:8], req.Dest.Addr[:])
	copy(buf[8:12], req.NH[:])
	return buf
}

func decodeIP4Route(data []byte) (IP4Route, error) {
	if len(data) < 8 {
		return IP4Route{}, fmt.Errorf("ip4 route data too short: %d bytes", len(data))
	}
	r := IP4Route{
		VRFID:  binary.LittleEndian.Uint16(data[0:2]),
		Origin: NHOrigin(data[2]),
	}
	r.Dest.PrefixLen = data[3]
	copy(r.Dest.Addr[:], data[4:8])

	if len(data) > 8 {
		nh, err := decodeNexthop(data[8:])
		if err != nil {
			return IP4Route{}, err
		}
		r.NH = nh
	}
	return r, nil
}

func decodeFIB4Info(data []byte) (FIB4Info, error) {
	if len(data) < fib4InfoSize {
		return FIB4Info{}, fmt.Errorf("fib4 info data too short: %d bytes", len(data))
	}
	return FIB4Info{
		VRFID:      binary.LittleEndian.Uint16(data[0:2]),
		MaxRoutes:  binary.LittleEndian.Uint32(data[4:8]),
		UsedRoutes: binary.LittleEndian.Uint32(data[8:12]),
		NumTbl8:    binary.LittleEndian.Uint32(data[12:16]),
		UsedTbl8:   binary.LittleEndian.Uint32(data[16:20]),
	}, nil
}

func encodeIP4PingReq(req IP4PingRequest) []byte {
	buf := make([]byte, 12) // addr(4) + vrf_id(2) + seq_num(2) + ttl(1) + pad(3)
	copy(buf[0:4], req.Addr[:])
	binary.LittleEndian.PutUint16(buf[4:6], req.VRFID)
	binary.LittleEndian.PutUint16(buf[6:8], req.SeqNum)
	buf[8] = req.TTL
	return buf
}

func decodeICMPRecvResp(data []byte) (ICMPRecvResp, error) {
	if len(data) < 16 {
		return ICMPRecvResp{}, fmt.Errorf("icmp recv data too short: %d bytes", len(data))
	}
	r := ICMPRecvResp{
		Type:   data[0],
		Code:   data[1],
		TTL:    data[2],
		Ident:  binary.LittleEndian.Uint16(data[4:6]),
		SeqNum: binary.LittleEndian.Uint16(data[6:8]),
	}
	copy(r.SrcAddr[:], data[8:12])
	if len(data) >= 20 {
		r.ResponseTime = ClockNS(binary.LittleEndian.Uint64(data[12:20]))
	}
	return r, nil
}
