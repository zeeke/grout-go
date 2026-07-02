# Research: Grout Go Client Library

**Branch**: `001-grout-client-library` | **Date**: 2026-07-02

## R1: Grout Wire Protocol

**Decision**: Implement a pure-Go binary protocol client over UNIX domain sockets.

**Rationale**: The grout control plane uses a custom binary protocol
(`gr_api.h`) over `AF_UNIX`/`SOCK_STREAM`. The protocol is simple
(fixed 12-byte headers + variable payload) and well-documented in
the C headers. A pure-Go implementation avoids CGo overhead, build
complexity, and cross-compilation issues.

**Protocol details**:
- Request header: `{id uint32, type uint32, payload_len uint32}`
- Response header: `{for_id uint32, status uint32, payload_len uint32}`
- Event header: `{ev_type uint32, payload_len size_t}`
- Max payload: 128 KiB (`GR_API_MAX_MSG_LEN = 131072`)
- Message type encoding: `(module_id << 16) | msg_id`
- Streaming: terminated by zero-length response payload
- Out-of-order: responses matched by `for_id` to request `id`
- Handshake: `GR_HELLO` with `api_version` + `version` string
- Current API version: 3
- Default socket: `/run/grout.sock`
- Byte order: native (little-endian on x86/arm64) for headers
  and struct fields. IP addresses in network byte order.

**Alternatives considered**:
- CGo bindings to `libgrout`: Would tie to a specific grout
  version at build time, complicate cross-compilation, and
  prevent test-mode mocking. Rejected.
- gRPC/protobuf wrapper: Not applicable — grout uses a custom
  binary protocol, not gRPC.

## R2: Module ID Registry

**Decision**: Define module IDs as Go constants matching the C headers.

| Module     | ID       | Source Header           |
|------------|----------|-------------------------|
| Main       | `0xcafe` | `api/gr_api.h`          |
| Infra      | `0xacdc` | `modules/infra/api/gr_infra.h` |
| IPv4       | `0xf00d` | `modules/ip/api/gr_ip4.h` |
| IPv6       | `0xfeed` | `modules/ip6/api/gr_ip6.h` |
| L2         | `0xbabe` | `modules/l2/api/gr_l2.h` |
| DHCP       | `0xd4c9` | `modules/dhcp/api/gr_dhcp.h` |
| Conntrack  | `0xc0c0` | `modules/policy/api/gr_conntrack.h` |
| NAT        | `0x0bad` | `modules/policy/api/gr_nat.h` |
| SRv6       | `0xfeef` | `modules/srv6/api/gr_srv6.h` |

## R3: Complete Message Type Inventory

**Decision**: Implement all message types across all modules
for full grcli parity.

### Main Module (0xcafe)
- `GR_HELLO` (0xcafe<<16 | 0x1981) — handshake
- `GR_LOG_PACKETS_SET` — toggle packet logging
- `GR_LOG_LEVEL_LIST` — stream log entries
- `GR_LOG_LEVEL_SET` — set log level by pattern
- `GR_EVENT_SUBSCRIBE` — subscribe to events
- `GR_EVENT_UNSUBSCRIBE` — unsubscribe all

### Infra Module (0xacdc)
- `GR_IFACE_ADD/DEL/GET/LIST/SET` — interface CRUD
- `GR_IFACE_STATS_GET` — per-interface statistics
- `GR_IFACE_MAC_ADD/DEL/LIST/SET` — MAC management
- `GR_AFFINITY_RXQ_LIST/SET` — RX queue affinity
- `GR_AFFINITY_CPU_GET/SET` — CPU affinity
- `GR_STATS_GET/RESET` — graph node stats
- `GR_GRAPH_DUMP/CONF_GET/CONF_SET` — graph config
- `GR_PACKET_TRACE_SET/CLEAR/DUMP` — packet tracing
- `GR_NH_ADD/DEL/GET/LIST/CONFIG_GET/CONFIG_SET` — nexthops

### IPv4 Module (0xf00d)
- `GR_IP4_ROUTE_ADD/DEL/GET/LIST` — route management
- `GR_IP4_ADDR_ADD/DEL/LIST/FLUSH` — address management
- `GR_IP4_ICMP_SEND/RECV` — ping
- `GR_IP4_FIB_DEFAULT_SET/INFO_LIST` — FIB config

### IPv6 Module (0xfeed)
- `GR_IP6_ROUTE_ADD/DEL/GET/LIST` — route management
- `GR_IP6_ADDR_ADD/DEL/LIST/FLUSH` — address management
- `GR_IP6_ICMP6_SEND/RECV` — ping6
- `GR_IP6_FIB_DEFAULT_SET/INFO_LIST` — FIB config
- `GR_IP6_IFACE_RA_SET/CLEAR/SHOW` — router advertisements

### L2 Module (0xbabe)
- `GR_FDB_ADD/DEL/FLUSH/LIST` — FDB entries
- `GR_FDB_CONFIG_GET/SET` — FDB config
- `GR_FLOOD_ADD/DEL/LIST` — flood entries

### DHCP Module (0xd4c9)
- `GR_DHCP_LIST/START/STOP` — DHCP client

### NAT Module (0x0bad)
- `GR_DNAT44_ADD/DEL/LIST` — static DNAT
- `GR_SNAT44_ADD/DEL/LIST` — dynamic SNAT

### Conntrack Module (0xc0c0)
- `GR_CONNTRACK_LIST/FLUSH` — connection entries
- `GR_CONNTRACK_CONF_GET/SET` — timeouts/limits

### SRv6 Module (0xfeef)
- `GR_SRV6_TUNSRC_SET/CLEAR/SHOW` — tunnel source

## R4: Event Types

**Decision**: Support all event subscriptions for real-time
monitoring use cases.

### Infra Events (0xacdc, offset 0x1001)
- `IFACE_ADD/POST_ADD/PRE_REMOVE/REMOVE` — lifecycle
- `IFACE_POST_RECONFIG` — configuration change
- `IFACE_STATUS_UP/DOWN` — link status
- `IFACE_MAC_CHANGE` — MAC address change

### IPv4 Events (0xf00d, offset 0x1001)
- `IP_ADDR_ADD/DEL` — address changes
- `IP_ROUTE_ADD/DEL` — route changes

### IPv6 Events (0xfeed, offset 0x1001)
- `IP6_ADDR_ADD/DEL` — address changes
- `IP6_ROUTE_ADD/DEL` — route changes

### L2 Events (0xbabe, offset 0x1001)
- `FDB_ADD/DEL/UPDATE` — FDB changes
- `FLOOD_ADD/DEL` — flood list changes

## R5: Go API Design

**Decision**: Use a single `Client` struct with domain-grouped
method receivers, mirroring grcli command groups.

**Rationale**: grcli organizes commands as `<group> <action>`
(e.g., `interface add`, `route list`). The Go API mirrors
this with `client.InterfaceAdd()`, `client.RouteList()`, etc.
This is simpler than sub-client objects and more discoverable
via godoc. The flat method approach avoids allocations for
sub-client construction and keeps the API surface grep-friendly.

**Pattern**:
```go
client, err := grout.Connect("/run/grout.sock")
defer client.Close()
id, err := client.InterfaceAdd(grout.InterfaceAddRequest{...})
ifaces, err := client.InterfaceList(grout.IfaceTypePort)
```

**Alternatives considered**:
- Sub-client pattern (`client.Interfaces().Add()`):
  More indirection, allocates intermediate objects, less
  idiomatic for protocol-level clients. Rejected.
- Separate packages per module: Over-segmented for a single
  protocol client. Users would need many imports. Rejected.

## R6: Multi-Version Strategy

**Decision**: Version-gate at the method level using the
negotiated API version from the handshake.

**Rationale**: The `GR_HELLO` handshake exchanges API versions.
Each public method checks whether the connected API version
supports the requested message type. If not, it returns
`ErrUnsupported` with a message indicating the minimum
required version.

**Implementation**:
- Store negotiated version in `Client` struct after handshake
- Maintain a compile-time map of message type → minimum API version
- Check before sending; return `ErrUnsupported{Feature, MinVersion}`
- The wire protocol format itself is stable across versions

**Versions to support**: API version 3 (current, grout ≥0.14).
Track older versions as needed via the grout changelog.

## R7: Testing Strategy

**Decision**: Three-tier testing approach.

### Unit tests (no grout daemon)
- Protocol encoding/decoding: serialize structs, verify bytes
- Request/response matching: out-of-order ID correlation
- Stream termination: empty payload detection
- Error mapping: errno → Go error types
- Type conversions: IP addresses, MAC, network byte order
- Table-driven tests per constitution requirement

### Integration tests (grout in container)
- Build tag: `//go:build integration`
- Container: `quay.io/grout/grout` with `-t` (test mode)
- Virtual ports via `net_tap` PMD (no real NIC needed)
- Test matrix: grout 0.15.x and 0.16.x (two versions)
- CI: Podman/Docker with `--privileged` or `--cap-add=NET_ADMIN`
- Each test creates interfaces, configures, validates, cleans up

### Race detection
- All tests run with `-race` per constitution requirement

## R8: Struct Serialization

**Decision**: Manual binary encoding using `encoding/binary`
with explicit field-by-field serialization.

**Rationale**: Grout structs use C native layout (with padding,
flexible arrays, unions, and the `BASE()` embedding pattern).
`encoding/binary.Read/Write` with explicit field handling gives
full control over padding, alignment, and variable-length tails.

**Key challenges**:
- `BASE()` pattern → Go struct embedding
- Flexible array members (`info[]`, `seglist[]`) → manual
  length-prefix parsing from `payload_len`
- Union types (`l3_addr`) → Go interface or tagged struct
- `ip4_addr_t` is `uint32` in network byte order
- Header fields are native endian (little-endian on x86/arm64)

**Alternatives considered**:
- `unsafe.Pointer` casting: Fragile, not portable, violates
  Go memory safety. Rejected.
- Code generation from C headers: Complex tooling, hard to
  maintain across versions. Rejected.

## R9: Container Testing Setup

**Decision**: Use `quay.io/grout/grout` images with test mode.

**Details**:
- Image: `quay.io/grout/grout:<version>`
- Available versions: 0.13, 0.14.x, 0.15.x, 0.16.x, edge
- Test mode flag: `-t` (no hugepages required)
- Virtual ports: `net_tap` PMD (`devargs "net_tap0,iface=x-name"`)
- Socket: mount a volume for the UNIX socket
- Capabilities: `NET_ADMIN` minimum for TAP device creation
- Multi-arch: amd64 and arm64 supported

**Test matrix**:
- grout 0.15.0 (API version 3, older stable)
- grout 0.16.0 (API version 3, latest stable)
- grout edge (rolling, for forward-compatibility checks)
