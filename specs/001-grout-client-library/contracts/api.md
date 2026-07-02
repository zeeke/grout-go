# Public API Contract: grout-go

**Package**: `github.com/zeeke/grout-go`

This document defines the public Go API surface. All exported
symbols listed here constitute the library's public contract
per the constitution's backward compatibility principle.

## Client Lifecycle

```go
// Option configures a Client connection.
type Option func(*clientConfig)

// WithSocketPath sets the UNIX socket path.
// Defaults to /run/grout.sock if not specified.
func WithSocketPath(path string) Option

// Connect establishes a connection to a grout daemon and performs
// the version handshake. Without options, connects to the default
// socket at /run/grout.sock.
// Returns ErrConnection if the daemon is unreachable, or
// ErrVersionMismatch if API versions are incompatible.
func Connect(opts ...Option) (*Client, error)

// Close closes the connection and releases resources.
// Safe to call multiple times.
func (c *Client) Close() error

// APIVersion returns the API version negotiated during handshake.
func (c *Client) APIVersion() uint32

// ServerVersion returns the grout daemon version string from handshake.
func (c *Client) ServerVersion() string
```

## Interface Operations

```go
// InterfaceListOption configures an InterfaceList query.
type InterfaceListOption func(*interfaceListConfig)

// WithIfaceType filters the list to interfaces of the given type.
// Without this option, all interface types are returned.
func WithIfaceType(t IfaceType) InterfaceListOption

func (c *Client) InterfaceAdd(req InterfaceAddRequest) (uint16, error)
func (c *Client) InterfaceDel(ifaceID uint16) error
func (c *Client) InterfaceGet(ifaceID uint16) (*Iface, error)
func (c *Client) InterfaceGetByName(name string) (*Iface, error)
func (c *Client) InterfaceList(opts ...InterfaceListOption) ([]Iface, error)
func (c *Client) InterfaceSet(req InterfaceSetRequest) error
func (c *Client) InterfaceStatsGet() ([]IfaceStats, error)
```

## MAC Management

```go
func (c *Client) InterfaceMACAdd(ifaceID uint16, mac EtherAddr) error
func (c *Client) InterfaceMACDel(ifaceID uint16, mac EtherAddr) error
func (c *Client) InterfaceMACList(ifaceID uint16) ([]IfaceMAC, error)
func (c *Client) InterfaceMACSet(ifaceID uint16, mac EtherAddr) error
```

## IPv4 Address Operations

```go
func (c *Client) IP4AddrAdd(addr IP4IfAddr, existOK bool) error
func (c *Client) IP4AddrDel(addr IP4IfAddr, missingOK bool) error
func (c *Client) IP4AddrList(vrfID, ifaceID uint16) ([]IP4IfAddr, error)
func (c *Client) IP4AddrFlush(ifaceID uint16) error
```

## IPv4 Route Operations

```go
func (c *Client) IP4RouteAdd(req IP4RouteAddRequest) error
func (c *Client) IP4RouteDel(vrfID uint16, dest IP4Net, missingOK bool) error
func (c *Client) IP4RouteGet(vrfID uint16, dest IP4Addr) (*Nexthop, error)
func (c *Client) IP4RouteList(vrfID uint16, maxCount uint16) ([]IP4Route, error)
```

## IPv4 FIB

```go
func (c *Client) IP4FIBDefaultSet(maxRoutes uint32) error
func (c *Client) IP4FIBInfoList(vrfID uint16) ([]FIB4Info, error)
```

## IPv4 Ping

```go
func (c *Client) IP4Ping(req IP4PingRequest) (*ICMPRecvResp, error)
```

## IPv6 Address Operations

```go
func (c *Client) IP6AddrAdd(addr IP6IfAddr, existOK bool) error
func (c *Client) IP6AddrDel(addr IP6IfAddr, missingOK bool) error
func (c *Client) IP6AddrList(vrfID, ifaceID uint16) ([]IP6IfAddr, error)
func (c *Client) IP6AddrFlush(ifaceID uint16) error
```

## IPv6 Route Operations

```go
func (c *Client) IP6RouteAdd(req IP6RouteAddRequest) error
func (c *Client) IP6RouteDel(vrfID uint16, dest IP6Net, missingOK bool) error
func (c *Client) IP6RouteGet(vrfID uint16, dest IP6Addr) (*Nexthop, error)
func (c *Client) IP6RouteList(vrfID uint16, maxCount uint16) ([]IP6Route, error)
```

## IPv6 FIB

```go
func (c *Client) IP6FIBDefaultSet(maxRoutes uint32) error
func (c *Client) IP6FIBInfoList(vrfID uint16) ([]FIB6Info, error)
```

## IPv6 Ping

```go
func (c *Client) IP6Ping(req IP6PingRequest) (*ICMP6RecvResp, error)
```

## IPv6 Router Advertisements

```go
func (c *Client) IP6RASet(req IP6RASetRequest) error
func (c *Client) IP6RAClear(ifaceID uint16) error
func (c *Client) IP6RAShow(ifaceID uint16) ([]RAConf, error)
```

## Nexthop Operations

```go
func (c *Client) NexthopAdd(req NexthopAddRequest) error
func (c *Client) NexthopDel(req NexthopDelRequest) error
func (c *Client) NexthopGet(nhID uint32) (*Nexthop, error)
func (c *Client) NexthopList(req NexthopListRequest) ([]Nexthop, error)
func (c *Client) NexthopConfigGet() (*NHConfig, error)
func (c *Client) NexthopConfigSet(config NHConfig) error
```

## NAT Operations

```go
func (c *Client) DNAT44Add(policy DNAT44Policy, existOK bool) error
func (c *Client) DNAT44Del(ifaceID uint16, match IP4Addr, missingOK bool) error
func (c *Client) DNAT44List(vrfID uint16) ([]DNAT44Policy, error)

func (c *Client) SNAT44Add(policy SNAT44Policy, existOK bool) error
func (c *Client) SNAT44Del(policy SNAT44Policy, missingOK bool) error
func (c *Client) SNAT44List() ([]SNAT44Policy, error)
```

## Connection Tracking

```go
func (c *Client) ConntrackList() ([]ConntrackEntry, error)
func (c *Client) ConntrackFlush() error
func (c *Client) ConntrackConfigGet() (*ConntrackConfig, error)
func (c *Client) ConntrackConfigSet(config ConntrackConfig) error
```

## L2 / FDB Operations

```go
func (c *Client) FDBAdd(entry FDBEntry, existOK bool) error
func (c *Client) FDBDel(bridgeID uint16, mac EtherAddr, vlanID uint16, missingOK bool) error
func (c *Client) FDBFlush(req FDBFlushRequest) error
func (c *Client) FDBList(req FDBListRequest) ([]FDBEntry, error)
func (c *Client) FDBConfigGet() (*FDBConfig, error)
func (c *Client) FDBConfigSet(maxEntries uint32) error
```

## L2 / Flood Operations

```go
func (c *Client) FloodAdd(entry FloodEntry, existOK bool) error
func (c *Client) FloodDel(entry FloodEntry, missingOK bool) error
func (c *Client) FloodList(floodType FloodType, vrfID uint16) ([]FloodEntry, error)
```

## DHCP Operations

```go
func (c *Client) DHCPList() ([]DHCPStatus, error)
func (c *Client) DHCPStart(ifaceID uint16) error
func (c *Client) DHCPStop(ifaceID uint16) error
```

## Statistics and Graph

```go
func (c *Client) StatsGet(req StatsGetRequest) ([]Stat, error)
func (c *Client) StatsReset() error
func (c *Client) GraphDump(req GraphDumpRequest) (string, error)
func (c *Client) GraphConfigGet() (*GraphConf, error)
func (c *Client) GraphConfigSet(conf GraphConf) error
```

## Packet Tracing

```go
func (c *Client) PacketTraceSet(ifaceID uint16, enabled, all bool) error
func (c *Client) PacketTraceClear() error
func (c *Client) PacketTraceDump(maxPackets uint16) ([]byte, error)
```

## Affinity

```go
func (c *Client) AffinityRxQList() ([]RxQueueMap, error)
func (c *Client) AffinityRxQSet(ifaceID, rxqID, cpuID uint16) error
func (c *Client) AffinityCPUGet() (*CPUAffinity, error)
func (c *Client) AffinityCPUSet(control, datapath CPUSet) error
```

## Logging

```go
func (c *Client) LogLevelList(showAll bool) ([]LogEntry, error)
func (c *Client) LogLevelSet(pattern string, level uint32) error
func (c *Client) LogPacketsSet(enabled bool) error
```

## SRv6

```go
func (c *Client) SRv6TunSrcSet(addr IP6Addr) error
func (c *Client) SRv6TunSrcClear() error
func (c *Client) SRv6TunSrcShow() (*IP6Addr, error)
```

## Events

```go
func (c *Client) EventSubscribe(evType uint32, suppressSelf bool) error
func (c *Client) EventUnsubscribe() error
func (c *Client) EventRecv() (*Event, error)
```

## Error Types

```go
var (
    ErrConnection      // UNIX socket connection failed
    ErrVersionMismatch // API version incompatible
    ErrUnsupported     // feature not available in connected version
    ErrPayloadTooLarge // payload exceeds 128 KiB limit
    ErrNotFound        // requested resource does not exist
)

// GrError wraps an errno returned by the grout daemon.
type GrError struct {
    Errno  int
    Method string
}
func (e *GrError) Error() string
```

## Constants

```go
const (
    DefaultSocketPath = "/run/grout.sock"
    MaxPayloadLen     = 128 * 1024
    APIVersion        = 3

    EventAll uint32 = 0xffffffff

    VRFDefaultID   uint16 = 1
    IfaceIDUndef   uint16 = 0
    DefaultVRFName        = "main"
)
```
