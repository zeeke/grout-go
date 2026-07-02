# Data Model: Grout Go Client Library

**Branch**: `001-grout-client-library` | **Date**: 2026-07-02

## Foundational Types

### Network Address Types

| Go Type        | C Equivalent        | Fields / Representation           |
|----------------|---------------------|------------------------------------|
| `IP4Addr`      | `ip4_addr_t`        | `uint32` (network byte order)      |
| `IP4Net`       | `struct ip4_net`    | `IP IP4Addr, PrefixLen uint8`      |
| `IP6Addr`      | `struct rte_ipv6_addr` | `[16]byte`                      |
| `IP6Net`       | `struct ip6_net`    | `IP IP6Addr, PrefixLen uint8`      |
| `EtherAddr`    | `struct rte_ether_addr` | `[6]byte`                     |
| `L3Addr`       | `struct l3_addr`    | `Family AddrFamily` + tagged union |
| `AddrFamily`   | `addr_family_t`     | `uint8` enum: Unspec=0, IP4=2, IP6=10 |

### Time and Clock

| Go Type        | C Equivalent      | Representation       |
|----------------|-------------------|-----------------------|
| `ClockNS`      | `gr_clock_ns_t`   | `int64` (nanoseconds) |

## Protocol Types

### Request/Response Headers

| Go Type          | C Equivalent          | Fields                              |
|------------------|-----------------------|--------------------------------------|
| `RequestHeader`  | `struct gr_api_request`  | `ID, Type, PayloadLen uint32`     |
| `ResponseHeader` | `struct gr_api_response` | `ForID, Status, PayloadLen uint32`|
| `EventHeader`    | `struct gr_api_event`    | `EvType uint32, PayloadLen uint64`|

### Hello Handshake

| Go Type       | C Equivalent          | Fields                        |
|---------------|-----------------------|--------------------------------|
| `HelloReq`    | `struct gr_hello_req` | `APIVersion uint32, Version [128]byte` |

## Infrastructure Entities

### Interface

| Go Type       | C Equivalent     | Key Fields |
|---------------|------------------|------------|
| `Iface`       | `struct gr_iface` | `ID uint16, Type IfaceType, Mode IfaceMode, Flags IfaceFlags, State IfaceState, MTU uint16, VRFID uint16, DomainID uint16, Speed uint32, Name string, Description string` |
| `IfaceType`   | `gr_iface_type_t` | `uint8` enum: Undef=0, VRF, Port, VLAN, IPIP, Bond, Bridge, VXLAN |
| `IfaceFlags`  | `gr_iface_flags_t` | `uint16` bitmask: Up, Promisc, PacketTrace, SNATStatic, SNATDynamic |
| `IfaceState`  | `gr_iface_state_t` | `uint16` bitmask: Running, PromiscFixed, AllMulti |
| `IfaceMode`   | `gr_iface_mode_t` | `uint8` enum: VRF=0, XC, Bond, Bridge |

### Interface Type-Specific Info

| Go Type          | C Equivalent                | Key Fields |
|------------------|-----------------------------|------------|
| `PortInfo`       | `gr_iface_info_port`        | `NRxQ, NTxQ, RxQSize, TxQSize uint16, MAC EtherAddr, DevArgs string, DriverName string` |
| `VRFInfo`        | `gr_iface_info_vrf`         | `IPv4FIB, IPv6FIB FIBConfig, MAC EtherAddr` |
| `VLANInfo`       | `gr_iface_info_vlan`        | `ParentID, VLANID uint16, MAC EtherAddr` |
| `BondInfo`       | `gr_iface_info_bond`        | `Mode BondMode, Algo BondAlgo, MAC EtherAddr, PrimaryMember uint8, Members []BondMember` |
| `BondMember`     | `gr_bond_member`            | `IfaceID uint16, Active bool` |
| `BridgeInfo`     | `gr_iface_info_bridge`      | `AgeingTime uint32, Flags BridgeFlags, MAC EtherAddr, Members []uint16` |
| `VXLANInfo`      | `gr_iface_info_vxlan`       | `VNI uint32, EncapVRFID uint16, DstPort uint16, Local L3Addr, MAC EtherAddr` |
| `IPIPInfo`       | `gr_iface_info_ipip`        | `Local, Remote IP4Addr` |
| `FIBConfig`      | `gr_iface_info_vrf_fib`     | `MaxRoutes, NumTbl8 uint32` |

### Bond/Bridge Enums

| Go Type       | C Equivalent     | Values |
|---------------|------------------|--------|
| `BondMode`    | `gr_bond_mode_t` | ActiveBackup=1, LACP |
| `BondAlgo`    | `gr_bond_algo_t` | RSS=1, L2, L3L4 |
| `BridgeFlags` | `gr_bridge_flags_t` | NoFlood, NoLearn |

### Statistics

| Go Type       | C Equivalent         | Key Fields |
|---------------|----------------------|------------|
| `IfaceStats`  | `gr_iface_stats`     | `IfaceID uint16, RxPackets, RxBytes, RxDrops, TxPackets, TxBytes, TxErrors, CPRxPackets, CPRxBytes, CPTxPackets, CPTxBytes uint64` |
| `Stat`        | `struct gr_stat`     | `Name string, TopoOrder, Packets, Batches, Cycles uint64` |
| `StatsFlags`  | `gr_stats_flags_t`   | SW, HW, Zero |
| `GraphConf`   | `gr_graph_conf`      | `RxBurstMax, VectorMax uint16` |

### Affinity

| Go Type       | C Equivalent      | Key Fields |
|---------------|-------------------|------------|
| `RxQueueMap`  | `gr_port_rxq_map` | `IfaceID, RxQID, CPUID, Enabled uint16` |

### MAC Management

| Go Type       | C Equivalent      | Key Fields |
|---------------|-------------------|------------|
| `IfaceMAC`    | `gr_iface_mac`    | `IfaceID uint16, RefCount uint16, Primary bool, MAC EtherAddr` |

## Nexthop Entities

| Go Type            | C Equivalent               | Key Fields |
|--------------------|----------------------------|------------|
| `Nexthop`          | `struct gr_nexthop`        | `Type NHType, Origin NHOrigin, IfaceID, VRFID uint16, NHID uint32` + type-specific info |
| `NHType`           | `gr_nh_type_t`             | L3=0, SR6Output, SR6Local, DNAT, Blackhole, Reject, Group |
| `NHOrigin`         | `gr_nh_origin_t`           | Unspec=0, Static=4, DHCP=16, Zebra=11, ... (full enum) |
| `NHState`          | `gr_nh_state_t`            | New=0, Pending, Reachable, Stale, Failed |
| `NHFlags`          | `gr_nh_flags_t`            | Static, Local, Gateway, Link, Mcast, Remote |
| `NHInfoL3`         | `gr_nexthop_info_l3`       | `State NHState, Flags NHFlags, Family AddrFamily, PrefixLen uint8, Addr L3Addr, MAC EtherAddr` |
| `NHInfoGroup`      | `gr_nexthop_info_group`    | `Members []NHGroupMember` |
| `NHGroupMember`    | `gr_nexthop_group_member`  | `NHID uint32, Weight uint32` |
| `NHConfig`         | `gr_nexthop_config`        | `MaxCount, LifetimeReachable, LifetimeUnreachable, MaxHeldPkts, MaxUcastProbes, MaxBcastProbes uint32` |

## IPv4 Entities

| Go Type          | C Equivalent           | Key Fields |
|------------------|------------------------|------------|
| `IP4IfAddr`      | `gr_ip4_ifaddr`        | `IfaceID uint16, Addr IP4Net` |
| `IP4Route`       | `gr_ip4_route`         | `Dest IP4Net, VRFID uint16, Origin NHOrigin, NH Nexthop` |
| `FIB4Info`       | `gr_fib4_info`         | `VRFID uint16, MaxRoutes, UsedRoutes, NumTbl8, UsedTbl8 uint32` |
| `ICMPRecvResp`   | `gr_ip4_icmp_recv_resp`| `Type, Code, TTL uint8, Ident, SeqNum uint16, SrcAddr IP4Addr, ResponseTime ClockNS` |

## IPv6 Entities

| Go Type          | C Equivalent            | Key Fields |
|------------------|-------------------------|------------|
| `IP6IfAddr`      | `gr_ip6_ifaddr`         | `IfaceID uint16, Addr IP6Net` |
| `IP6Route`       | `gr_ip6_route`          | `Dest IP6Net, VRFID uint16, Origin NHOrigin, NH Nexthop` |
| `FIB6Info`       | `gr_fib6_info`          | `VRFID uint16, MaxRoutes, UsedRoutes, NumTbl8, UsedTbl8 uint32` |
| `ICMP6RecvResp`  | `gr_ip6_icmp_recv_resp` | `Type, Code, TTL uint8, Ident, SeqNum uint16, SrcAddr IP6Addr, ResponseTime ClockNS` |
| `RAConf`         | `gr_ip6_ra_conf`        | `Enabled bool, IfaceID uint16, Interval, Lifetime uint16` |

## L2 / Bridge Entities

| Go Type          | C Equivalent         | Key Fields |
|------------------|----------------------|------------|
| `FDBEntry`       | `gr_fdb_entry`       | `BridgeID uint16, MAC EtherAddr, VLANID, IfaceID uint16, VTEP L3Addr, Flags FDBFlags, LastSeen ClockNS` |
| `FDBFlags`       | `gr_fdb_flags_t`     | Static, Learn, Extern |
| `FDBConfig`      | `gr_fdb_config_*`    | `MaxEntries, UsedEntries uint32` |
| `FloodEntry`     | `gr_flood_entry`     | `Type FloodType, VRFID uint16` + type-specific |
| `FloodVTEP`      | `gr_flood_vtep`      | `VNI uint32, Addr L3Addr` |
| `FloodType`      | `gr_flood_type_t`    | VTEP=1 |

## DHCP Entities

| Go Type          | C Equivalent       | Key Fields |
|------------------|--------------------|------------|
| `DHCPStatus`     | `gr_dhcp_status`   | `IfaceID uint16, State DHCPState, ServerIP, AssignedIP IP4Addr, LeaseTime, RenewalTime, RebindTime uint32` |
| `DHCPState`      | `dhcp_state_t`     | Init=0, Selecting, Requesting, Bound, Renewing, Rebinding |

## NAT Entities

| Go Type          | C Equivalent         | Key Fields |
|------------------|----------------------|------------|
| `DNAT44Policy`   | `gr_dnat44_policy`   | `IfaceID uint16, Match, Replace IP4Addr` |
| `SNAT44Policy`   | `gr_snat44_policy`   | `IfaceID uint16, Net IP4Net, Replace IP4Addr` |

## Conntrack Entities

| Go Type           | C Equivalent          | Key Fields |
|-------------------|-----------------------|------------|
| `ConntrackEntry`  | `gr_conntrack`        | `IfaceID uint16, Family AddrFamily, Proto uint8, FwdFlow, RevFlow ConntrackFlow, LastUpdate ClockNS, ID uint32, State ConnState` |
| `ConntrackFlow`   | `gr_conntrack_flow`   | `Src, Dst IP4Addr, SrcID, DstID uint16` |
| `ConnState`       | `gr_conn_state_t`     | Closed=0, New, SimSynSent, SynReceived, Established, ... TimeWait |
| `ConntrackConfig` | `gr_conntrack_config` | `MaxCount, TimeoutClosed, TimeoutNew, TimeoutUDPEstablished, TimeoutTCPEstablished, TimeoutHalfClose, TimeoutTimeWait uint32` |

## SRv6 Entities

| Go Type               | C Equivalent                 | Key Fields |
|-----------------------|------------------------------|------------|
| `SRv6EncapBehavior`   | `gr_srv6_encap_behavior_t`   | Encaps=0, EncapsRed |
| `SRv6Behavior`        | `gr_srv6_behavior_t`         | End=0x0001, EndT=0x0009, EndDT6=0x0012, EndDT4=0x0013, EndDT46=0x0014 |
| `SRv6Flags`           | `gr_srv6_flags_t`            | PSP, USD, NextCSID |
| `NHInfoSRv6`          | `gr_nexthop_info_srv6`       | `EncapBehavior, SegList []IP6Addr` |
| `NHInfoSRv6Local`     | `gr_nexthop_info_srv6_local` | `OutVRFID uint16, Behavior SRv6Behavior, Flags SRv6Flags, BlockBits, CSIDBits uint8` |

## Logging Entities

| Go Type       | C Equivalent      | Key Fields |
|---------------|-------------------|------------|
| `LogEntry`    | `gr_log_entry`    | `Name string, Level uint32` |

## Entity Relationships

```
Client
 ├── connects to grout daemon via UNIX socket
 ├── negotiates API version via HelloReq
 └── sends/receives all message types

Iface (interface)
 ├── has type-specific info: PortInfo | VRFInfo | VLANInfo | ...
 ├── belongs to a VRF (vrf_id)
 ├── has zero or more IP4IfAddr (IPv4 addresses)
 ├── has zero or more IP6IfAddr (IPv6 addresses)
 ├── has zero or more IfaceMAC entries
 ├── has IfaceStats (counters)
 └── referenced by: Route.NH, FDBEntry, ConntrackEntry, NAT policies

IP4Route / IP6Route
 ├── belongs to a VRF
 ├── references a Nexthop
 └── has origin (static, zebra, dhcp, etc.)

Nexthop
 ├── has type: L3, SR6Output, SR6Local, DNAT, Blackhole, Reject, Group
 ├── references an Iface (iface_id)
 ├── has type-specific info: NHInfoL3 | NHInfoGroup | NHInfoSRv6 | ...
 └── belongs to a VRF

FDBEntry (L2)
 ├── belongs to a Bridge interface
 └── references a MAC address + VLAN + destination iface or VTEP

ConntrackEntry
 ├── references an Iface
 └── has forward and reverse flow (5-tuple)

DNAT44Policy / SNAT44Policy
 └── references an Iface

Event
 ├── carries a payload matching one of the above entities
 └── delivered asynchronously after subscription
```
