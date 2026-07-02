package grout

import (
	"encoding/binary"
	"fmt"
)

// IP6IfAddr represents an IPv6 interface address.
type IP6IfAddr struct {
	IfaceID uint16
	Addr    IP6Net
}

// IP6Route represents an IPv6 route with its nexthop.
type IP6Route struct {
	Dest   IP6Net
	VRFID  uint16
	Origin NHOrigin
	NH     Nexthop
}

// IP6RouteAddRequest describes parameters for adding an IPv6 route.
type IP6RouteAddRequest struct {
	VRFID  uint16
	Dest   IP6Net
	NH     IP6Addr
	Origin NHOrigin
}

// IP6PingRequest describes parameters for an IPv6 ICMP ping.
type IP6PingRequest struct {
	Addr   IP6Addr
	VRFID  uint16
	SeqNum uint16
	TTL    uint8
}

// ICMP6RecvResp holds the response from an IPv6 ping.
type ICMP6RecvResp struct {
	Type         uint8
	Code         uint8
	TTL          uint8
	Ident        uint16
	SeqNum       uint16
	SrcAddr      IP6Addr
	ResponseTime ClockNS
}

// FIB6Info holds IPv6 FIB table information.
type FIB6Info struct {
	VRFID      uint16
	MaxRoutes  uint32
	UsedRoutes uint32
	NumTbl8    uint32
	UsedTbl8   uint32
}

// RAConf holds router advertisement configuration.
type RAConf struct {
	Enabled  bool
	IfaceID  uint16
	Interval uint16
	Lifetime uint16
}

// Wire sizes.
const (
	ip6IfAddrSize   = 20 // iface_id(2) + pad(1) + prefix_len(1) + addr(16)
	fib6InfoSize    = 20
	raConfSize      = 8  // enabled(1) + pad(1) + iface_id(2) + interval(2) + lifetime(2)
)

// IP6AddrAdd adds an IPv6 address to an interface.
func (c *Client) IP6AddrAdd(addr IP6IfAddr, existOK bool) error {
	if err := c.checkVersion(msgTypeIP6AddrAdd); err != nil {
		return err
	}
	payload := encodeIP6IfAddr(addr, existOK)
	_, err := c.request(msgTypeIP6AddrAdd, payload)
	return err
}

// IP6AddrDel removes an IPv6 address from an interface.
func (c *Client) IP6AddrDel(addr IP6IfAddr, missingOK bool) error {
	if err := c.checkVersion(msgTypeIP6AddrDel); err != nil {
		return err
	}
	payload := encodeIP6IfAddr(addr, missingOK)
	_, err := c.request(msgTypeIP6AddrDel, payload)
	return err
}

// IP6AddrList returns IPv6 addresses, optionally filtered by VRF and interface.
func (c *Client) IP6AddrList(vrfID, ifaceID uint16) ([]IP6IfAddr, error) {
	if err := c.checkVersion(msgTypeIP6AddrList); err != nil {
		return nil, err
	}
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint16(payload[0:2], vrfID)
	binary.LittleEndian.PutUint16(payload[2:4], ifaceID)

	results, err := c.requestStream(msgTypeIP6AddrList, payload)
	if err != nil {
		return nil, err
	}

	addrs := make([]IP6IfAddr, 0, len(results))
	for _, data := range results {
		a, err := decodeIP6IfAddr(data)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, a)
	}
	return addrs, nil
}

// IP6AddrFlush removes all IPv6 addresses from an interface.
func (c *Client) IP6AddrFlush(ifaceID uint16) error {
	if err := c.checkVersion(msgTypeIP6AddrFlush); err != nil {
		return err
	}
	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, ifaceID)
	_, err := c.request(msgTypeIP6AddrFlush, payload)
	return err
}

// IP6RouteAdd adds an IPv6 route.
func (c *Client) IP6RouteAdd(req IP6RouteAddRequest) error {
	if err := c.checkVersion(msgTypeIP6RouteAdd); err != nil {
		return err
	}
	payload := encodeIP6RouteAdd(req)
	_, err := c.request(msgTypeIP6RouteAdd, payload)
	return err
}

// IP6RouteDel deletes an IPv6 route.
func (c *Client) IP6RouteDel(vrfID uint16, dest IP6Net, missingOK bool) error {
	if err := c.checkVersion(msgTypeIP6RouteDel); err != nil {
		return err
	}
	payload := make([]byte, 20)
	binary.LittleEndian.PutUint16(payload[0:2], vrfID)
	payload[3] = dest.PrefixLen
	copy(payload[4:20], dest.Addr[:])
	_, err := c.request(msgTypeIP6RouteDel, payload)
	return err
}

// IP6RouteGet looks up a route for a destination address.
func (c *Client) IP6RouteGet(vrfID uint16, dest IP6Addr) (*Nexthop, error) {
	if err := c.checkVersion(msgTypeIP6RouteGet); err != nil {
		return nil, err
	}
	payload := make([]byte, 18)
	binary.LittleEndian.PutUint16(payload[0:2], vrfID)
	copy(payload[2:18], dest[:])
	resp, err := c.request(msgTypeIP6RouteGet, payload)
	if err != nil {
		return nil, err
	}
	nh, err := decodeNexthop(resp)
	if err != nil {
		return nil, err
	}
	return &nh, nil
}

// IP6RouteList returns IPv6 routes for a VRF.
func (c *Client) IP6RouteList(vrfID uint16, maxCount uint16) ([]IP6Route, error) {
	if err := c.checkVersion(msgTypeIP6RouteList); err != nil {
		return nil, err
	}
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint16(payload[0:2], vrfID)
	binary.LittleEndian.PutUint16(payload[2:4], maxCount)

	results, err := c.requestStream(msgTypeIP6RouteList, payload)
	if err != nil {
		return nil, err
	}

	routes := make([]IP6Route, 0, len(results))
	for _, data := range results {
		r, err := decodeIP6Route(data)
		if err != nil {
			return nil, err
		}
		routes = append(routes, r)
	}
	return routes, nil
}

// IP6FIBDefaultSet sets the default IPv6 FIB table size.
func (c *Client) IP6FIBDefaultSet(maxRoutes uint32) error {
	if err := c.checkVersion(msgTypeIP6FIBDefaultSet); err != nil {
		return err
	}
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, maxRoutes)
	_, err := c.request(msgTypeIP6FIBDefaultSet, payload)
	return err
}

// IP6FIBInfoList returns IPv6 FIB table information.
func (c *Client) IP6FIBInfoList(vrfID uint16) ([]FIB6Info, error) {
	if err := c.checkVersion(msgTypeIP6FIBInfoList); err != nil {
		return nil, err
	}
	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, vrfID)

	results, err := c.requestStream(msgTypeIP6FIBInfoList, payload)
	if err != nil {
		return nil, err
	}

	infos := make([]FIB6Info, 0, len(results))
	for _, data := range results {
		info, err := decodeFIB6Info(data)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// IP6Ping sends an ICMPv6 echo request and waits for the reply.
func (c *Client) IP6Ping(req IP6PingRequest) (*ICMP6RecvResp, error) {
	if err := c.checkVersion(msgTypeIP6ICMPSend); err != nil {
		return nil, err
	}
	payload := encodeIP6PingReq(req)
	_, err := c.request(msgTypeIP6ICMPSend, payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.requestNoPayload(msgTypeIP6ICMPRecv)
	if err != nil {
		return nil, err
	}

	icmp, err := decodeICMP6RecvResp(resp)
	if err != nil {
		return nil, err
	}
	return &icmp, nil
}

// IP6RASet configures router advertisements on an interface.
func (c *Client) IP6RASet(conf RAConf) error {
	if err := c.checkVersion(msgTypeIP6RASet); err != nil {
		return err
	}
	payload := encodeRAConf(conf)
	_, err := c.request(msgTypeIP6RASet, payload)
	return err
}

// IP6RAClear disables router advertisements on an interface.
func (c *Client) IP6RAClear(ifaceID uint16) error {
	if err := c.checkVersion(msgTypeIP6RAClear); err != nil {
		return err
	}
	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, ifaceID)
	_, err := c.request(msgTypeIP6RAClear, payload)
	return err
}

// IP6RAShow returns router advertisement configuration.
func (c *Client) IP6RAShow(ifaceID uint16) ([]RAConf, error) {
	if err := c.checkVersion(msgTypeIP6RAShow); err != nil {
		return nil, err
	}
	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, ifaceID)

	results, err := c.requestStream(msgTypeIP6RAShow, payload)
	if err != nil {
		return nil, err
	}

	confs := make([]RAConf, 0, len(results))
	for _, data := range results {
		conf, err := decodeRAConf(data)
		if err != nil {
			return nil, err
		}
		confs = append(confs, conf)
	}
	return confs, nil
}

// Encoding/decoding helpers.

func encodeIP6IfAddr(addr IP6IfAddr, flag bool) []byte {
	buf := make([]byte, ip6IfAddrSize)
	binary.LittleEndian.PutUint16(buf[0:2], addr.IfaceID)
	if flag {
		buf[2] = 1
	}
	buf[3] = addr.Addr.PrefixLen
	copy(buf[4:20], addr.Addr.Addr[:])
	return buf
}

func decodeIP6IfAddr(data []byte) (IP6IfAddr, error) {
	if len(data) < ip6IfAddrSize {
		return IP6IfAddr{}, fmt.Errorf("ip6 ifaddr data too short: %d bytes", len(data))
	}
	a := IP6IfAddr{
		IfaceID: binary.LittleEndian.Uint16(data[0:2]),
	}
	a.Addr.PrefixLen = data[3]
	copy(a.Addr.Addr[:], data[4:20])
	return a, nil
}

func encodeIP6RouteAdd(req IP6RouteAddRequest) []byte {
	buf := make([]byte, 36) // vrf_id(2) + origin(1) + prefix_len(1) + dest(16) + nh(16)
	binary.LittleEndian.PutUint16(buf[0:2], req.VRFID)
	buf[2] = byte(req.Origin)
	buf[3] = req.Dest.PrefixLen
	copy(buf[4:20], req.Dest.Addr[:])
	copy(buf[20:36], req.NH[:])
	return buf
}

func decodeIP6Route(data []byte) (IP6Route, error) {
	if len(data) < 20 {
		return IP6Route{}, fmt.Errorf("ip6 route data too short: %d bytes", len(data))
	}
	r := IP6Route{
		VRFID:  binary.LittleEndian.Uint16(data[0:2]),
		Origin: NHOrigin(data[2]),
	}
	r.Dest.PrefixLen = data[3]
	copy(r.Dest.Addr[:], data[4:20])

	if len(data) > 20 {
		nh, err := decodeNexthop(data[20:])
		if err != nil {
			return IP6Route{}, err
		}
		r.NH = nh
	}
	return r, nil
}

func decodeFIB6Info(data []byte) (FIB6Info, error) {
	if len(data) < fib6InfoSize {
		return FIB6Info{}, fmt.Errorf("fib6 info data too short: %d bytes", len(data))
	}
	return FIB6Info{
		VRFID:      binary.LittleEndian.Uint16(data[0:2]),
		MaxRoutes:  binary.LittleEndian.Uint32(data[4:8]),
		UsedRoutes: binary.LittleEndian.Uint32(data[8:12]),
		NumTbl8:    binary.LittleEndian.Uint32(data[12:16]),
		UsedTbl8:   binary.LittleEndian.Uint32(data[16:20]),
	}, nil
}

func encodeIP6PingReq(req IP6PingRequest) []byte {
	buf := make([]byte, 24) // addr(16) + vrf_id(2) + seq_num(2) + ttl(1) + pad(3)
	copy(buf[0:16], req.Addr[:])
	binary.LittleEndian.PutUint16(buf[16:18], req.VRFID)
	binary.LittleEndian.PutUint16(buf[18:20], req.SeqNum)
	buf[20] = req.TTL
	return buf
}

func decodeICMP6RecvResp(data []byte) (ICMP6RecvResp, error) {
	if len(data) < 28 {
		return ICMP6RecvResp{}, fmt.Errorf("icmp6 recv data too short: %d bytes", len(data))
	}
	r := ICMP6RecvResp{
		Type:   data[0],
		Code:   data[1],
		TTL:    data[2],
		Ident:  binary.LittleEndian.Uint16(data[4:6]),
		SeqNum: binary.LittleEndian.Uint16(data[6:8]),
	}
	copy(r.SrcAddr[:], data[8:24])
	if len(data) >= 32 {
		r.ResponseTime = ClockNS(binary.LittleEndian.Uint64(data[24:32]))
	}
	return r, nil
}

func encodeRAConf(conf RAConf) []byte {
	buf := make([]byte, raConfSize)
	if conf.Enabled {
		buf[0] = 1
	}
	binary.LittleEndian.PutUint16(buf[2:4], conf.IfaceID)
	binary.LittleEndian.PutUint16(buf[4:6], conf.Interval)
	binary.LittleEndian.PutUint16(buf[6:8], conf.Lifetime)
	return buf
}

func decodeRAConf(data []byte) (RAConf, error) {
	if len(data) < raConfSize {
		return RAConf{}, fmt.Errorf("ra conf data too short: %d bytes", len(data))
	}
	return RAConf{
		Enabled:  data[0] != 0,
		IfaceID:  binary.LittleEndian.Uint16(data[2:4]),
		Interval: binary.LittleEndian.Uint16(data[4:6]),
		Lifetime: binary.LittleEndian.Uint16(data[6:8]),
	}, nil
}
