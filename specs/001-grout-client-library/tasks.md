# Tasks: Grout Go Client Library

**Input**: Design documents from `specs/001-grout-client-library/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/api.md, quickstart.md

**Tests**: Tests are included — the constitution mandates unit tests for all exported functions and integration tests for I/O behavior.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single package library**: all `.go` files at repository root
- **Test files**: alongside source files per Go convention (`*_test.go`)

---

## Phase 1: Setup

**Purpose**: Project initialization and Go module structure

- [ ] T001 Initialize Go module with `go mod init github.com/zeeke/grout-go` in go.mod
- [ ] T002 [P] Create package documentation and Client type stub with Option pattern in grout.go
- [ ] T003 [P] Create error types (ErrConnection, ErrVersionMismatch, ErrUnsupported, ErrPayloadTooLarge, ErrNotFound, GrError) in errors.go
- [ ] T004 [P] Create shared network types (IP4Addr, IP4Net, IP6Addr, IP6Net, EtherAddr, L3Addr, AddrFamily, ClockNS) with parsing helpers (MustParseIP4, MustParseIP4Net, MustParseIP6, MustParseIP6Net) and String() methods in types.go
- [ ] T005 [P] Create test utilities: mock UNIX socket server, request/response helpers, assertion functions in testutil_test.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Wire protocol implementation that ALL user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [ ] T006 Implement protocol constants (module IDs, message type encoding macro, API version, max payload), request/response header types, and binary serialization (encodeHeader, decodeHeader) in protocol.go
- [ ] T007 Implement send (write request header + payload to socket with full-write loop) and recv (read response header, match by ID, buffer out-of-order responses, validate payload size) in protocol.go
- [ ] T008 Implement streaming response iterator (loop receiving responses until empty payload terminator, drain on early exit) in protocol.go
- [ ] T009 Implement event reception (read event header, validate type against registry, return header + payload) in protocol.go
- [ ] T010 Implement API version registry mapping message types to minimum required API version, with checkVersion helper that returns ErrUnsupported in version.go
- [ ] T011 [P] Write unit tests for protocol encoding/decoding: header serialization round-trip, message type encoding, payload size validation, out-of-order response matching, stream termination in protocol_test.go
- [ ] T012 [P] Write unit tests for shared types: IP4Addr/IP6Addr parsing and String(), IP4Net/IP6Net parsing, EtherAddr formatting, L3Addr union handling in types_test.go

**Checkpoint**: Wire protocol fully functional — all message types can be sent/received

---

## Phase 3: User Story 1 — Connect to a Grout Instance (Priority: P1) MVP

**Goal**: Establish connection to grout daemon with version handshake

**Independent Test**: Connect to a mock socket server, verify handshake, check version accessors

### Tests for User Story 1

- [ ] T013 [P] [US1] Write unit tests for Connect (successful handshake, connection refused, version mismatch, custom socket path via WithSocketPath option), Close (idempotent), APIVersion, ServerVersion in grout_test.go

### Implementation for User Story 1

- [ ] T014 [US1] Implement clientConfig, Option type, WithSocketPath option function in grout.go
- [ ] T015 [US1] Implement Connect: dial UNIX socket, send GR_HELLO with api_version and build version, receive response, store negotiated version, return Client in grout.go
- [ ] T016 [US1] Implement Close (close socket, safe for multiple calls), APIVersion(), ServerVersion() accessors in grout.go
- [ ] T017 [P] [US1] Write unit tests for error types: GrError.Error() formatting, errors.Is/As with sentinel errors in errors_test.go

**Checkpoint**: `grout.Connect()` works end-to-end against mock and real grout

---

## Phase 4: User Story 2 — Manage Interfaces (Priority: P1)

**Goal**: Create, list, configure, and delete network interfaces

**Independent Test**: Add a port interface, list it, set flags, delete it

### Tests for User Story 2

- [ ] T018 [P] [US2] Write unit tests for InterfaceAdd (port, VRF, VLAN types), InterfaceDel, InterfaceGet, InterfaceGetByName, InterfaceList (with/without WithIfaceType option), InterfaceSet in iface_test.go

### Implementation for User Story 2

- [ ] T019 [US2] Define interface enums (IfaceType, IfaceFlags, IfaceState, IfaceMode) and Iface struct with type-specific info deserialization in iface_types.go
- [ ] T020 [US2] Define all interface type-specific info structs (PortInfo, VRFInfo, VLANInfo, BondInfo, BondMember, BridgeInfo, VXLANInfo, IPIPInfo, FIBConfig) and their binary encoding/decoding in iface_types.go
- [ ] T021 [US2] Define request structs (InterfaceAddRequest, InterfaceSetRequest, InterfaceListOption, WithIfaceType) in iface.go
- [ ] T022 [US2] Implement InterfaceAdd, InterfaceDel, InterfaceGet, InterfaceGetByName in iface.go
- [ ] T023 [US2] Implement InterfaceList (streaming response, optional type filter via options pattern), InterfaceSet (set_attrs bitmask) in iface.go
- [ ] T024 [US2] Implement InterfaceStatsGet (streaming IfaceStats), InterfaceMACAdd, InterfaceMACDel, InterfaceMACList (streaming), InterfaceMACSet in iface.go
- [ ] T025 [P] [US2] Write unit tests for InterfaceStatsGet, InterfaceMAC* operations, and interface type-specific info serialization round-trips in iface_test.go

**Checkpoint**: Full interface lifecycle works — create port, configure, list, delete

---

## Phase 5: User Story 3 — Configure IP Addresses and Routes (Priority: P1)

**Goal**: IPv4/IPv6 address and route management within VRFs

**Independent Test**: Add IPv4 address to interface, add route, lookup route, list routes, flush addresses

### Tests for User Story 3

- [ ] T026 [P] [US3] Write unit tests for IP4AddrAdd, IP4AddrDel, IP4AddrList, IP4AddrFlush, IP4RouteAdd, IP4RouteDel, IP4RouteGet, IP4RouteList, IP4FIBDefaultSet, IP4FIBInfoList in ip4_test.go
- [ ] T027 [P] [US3] Write unit tests for IP6AddrAdd, IP6AddrDel, IP6AddrList, IP6AddrFlush, IP6RouteAdd, IP6RouteDel, IP6RouteGet, IP6RouteList, IP6FIBDefaultSet, IP6FIBInfoList in ip6_test.go

### Implementation for User Story 3

- [ ] T028 [US3] Define nexthop types (Nexthop, NHType, NHOrigin, NHState, NHFlags, NHInfoL3, NHInfoGroup, NHGroupMember, NHInfoSRv6, NHInfoSRv6Local, NHInfoDNAT, NHConfig) with binary encoding/decoding including flexible array deserialization in nexthop.go
- [ ] T029 [US3] Define IPv4 request/response structs (IP4IfAddr, IP4Route, IP4RouteAddRequest, IP4PingRequest, ICMPRecvResp, FIB4Info) in ip4.go
- [ ] T030 [US3] Implement IP4AddrAdd, IP4AddrDel, IP4AddrList (streaming), IP4AddrFlush in ip4.go
- [ ] T031 [US3] Implement IP4RouteAdd, IP4RouteDel, IP4RouteGet, IP4RouteList (streaming with nexthop deserialization) in ip4.go
- [ ] T032 [US3] Implement IP4FIBDefaultSet, IP4FIBInfoList (streaming) in ip4.go
- [ ] T033 [US3] Define IPv6 request/response structs (IP6IfAddr, IP6Route, IP6RouteAddRequest, IP6PingRequest, ICMP6RecvResp, FIB6Info, RAConf) in ip6.go
- [ ] T034 [US3] Implement IP6AddrAdd, IP6AddrDel, IP6AddrList (streaming), IP6AddrFlush in ip6.go
- [ ] T035 [US3] Implement IP6RouteAdd, IP6RouteDel, IP6RouteGet, IP6RouteList (streaming) in ip6.go
- [ ] T036 [US3] Implement IP6FIBDefaultSet, IP6FIBInfoList (streaming) in ip6.go
- [ ] T037 [P] [US3] Write unit tests for nexthop type serialization round-trips (L3, Group, SRv6, DNAT, Blackhole) in nexthop_test.go

**Checkpoint**: Full IPv4/IPv6 address and route lifecycle works within VRFs

---

## Phase 6: User Story 7 — Multi-Version Compatibility (Priority: P1)

**Goal**: Version-gated methods returning ErrUnsupported for unavailable features

**Independent Test**: Connect to mock servers advertising different API versions, verify supported ops succeed and unsupported ops return ErrUnsupported

### Tests for User Story 7

- [ ] T038 [P] [US7] Write unit tests for version checking: method succeeds at minimum version, returns ErrUnsupported below minimum, APIVersion accessor reflects negotiated version in version_test.go

### Implementation for User Story 7

- [ ] T039 [US7] Populate version registry with minimum API version for each message type based on grout changelog (all current messages require API v3) in version.go
- [ ] T040 [US7] Add version check calls at the top of every public Client method (InterfaceAdd, IP4RouteAdd, etc.) in all source files
- [ ] T041 [US7] Write integration test helper that starts grout containers for versions 0.15.0 and 0.16.0, runs the same test suite against both in integration_test.go

**Checkpoint**: Library gracefully handles version differences across grout releases

---

## Phase 7: User Story 4 — NAT and Connection Tracking (Priority: P2)

**Goal**: Configure DNAT44, SNAT44, and inspect conntrack entries

**Independent Test**: Add DNAT44 rule, list it, delete it; list conntrack entries; configure conntrack timeouts

### Tests for User Story 4

- [ ] T042 [P] [US4] Write unit tests for DNAT44Add, DNAT44Del, DNAT44List, SNAT44Add, SNAT44Del, SNAT44List in nat_test.go
- [ ] T043 [P] [US4] Write unit tests for ConntrackList, ConntrackFlush, ConntrackConfigGet, ConntrackConfigSet in conntrack_test.go

### Implementation for User Story 4

- [ ] T044 [US4] Define NAT types (DNAT44Policy, SNAT44Policy) and request structs with binary encoding in nat.go
- [ ] T045 [US4] Implement DNAT44Add, DNAT44Del, DNAT44List (streaming), SNAT44Add, SNAT44Del, SNAT44List (streaming) in nat.go
- [ ] T046 [US4] Define conntrack types (ConntrackEntry, ConntrackFlow, ConnState, ConntrackConfig) with binary encoding in conntrack.go
- [ ] T047 [US4] Implement ConntrackList (streaming), ConntrackFlush, ConntrackConfigGet, ConntrackConfigSet in conntrack.go

**Checkpoint**: Full NAT configuration and conntrack inspection works

---

## Phase 8: User Story 5 — Statistics, Diagnostics, and Events (Priority: P2)

**Goal**: Retrieve stats, graph info, packet traces, and subscribe to events

**Independent Test**: Get software stats, reset, enable packet trace, dump trace, subscribe to events

### Tests for User Story 5

- [ ] T048 [P] [US5] Write unit tests for StatsGet, StatsReset, GraphDump, GraphConfigGet, GraphConfigSet in stats_test.go
- [ ] T049 [P] [US5] Write unit tests for PacketTraceSet, PacketTraceClear, PacketTraceDump in trace_test.go
- [ ] T050 [P] [US5] Write unit tests for EventSubscribe, EventUnsubscribe, EventRecv (interface events, route events, address events) in events_test.go

### Implementation for User Story 5

- [ ] T051 [US5] Define stats types (Stat, StatsFlags, StatsGetRequest, GraphConf, GraphDumpRequest) with binary encoding in stats.go
- [ ] T052 [US5] Implement StatsGet (streaming), StatsReset, GraphDump, GraphConfigGet, GraphConfigSet in stats.go
- [ ] T053 [US5] Implement PacketTraceSet, PacketTraceClear, PacketTraceDump in trace.go
- [ ] T054 [US5] Define Event type with payload deserialization for all event types (infra, IPv4, IPv6, L2) in events.go
- [ ] T055 [US5] Implement EventSubscribe, EventUnsubscribe, EventRecv in events.go

**Checkpoint**: Full observability: stats, traces, and real-time events

---

## Phase 9: User Story 6 — Advanced Features (Priority: P3)

**Goal**: Ping, SRv6, router advertisements, affinity, logging, L2/FDB, DHCP, nexthop management

**Independent Test**: Send ping, set SRv6 tunnel source, set log levels, manage FDB entries, start DHCP client

### Tests for User Story 6

- [ ] T056 [P] [US6] Write unit tests for IP4Ping, IP6Ping in ip4_test.go and ip6_test.go
- [ ] T057 [P] [US6] Write unit tests for SRv6TunSrcSet, SRv6TunSrcClear, SRv6TunSrcShow in srv6_test.go
- [ ] T058 [P] [US6] Write unit tests for IP6RASet, IP6RAClear, IP6RAShow in ip6_test.go
- [ ] T059 [P] [US6] Write unit tests for AffinityRxQList, AffinityRxQSet, AffinityCPUGet, AffinityCPUSet in affinity_test.go
- [ ] T060 [P] [US6] Write unit tests for LogLevelList, LogLevelSet, LogPacketsSet in log_test.go
- [ ] T061 [P] [US6] Write unit tests for FDBAdd, FDBDel, FDBFlush, FDBList, FDBConfigGet, FDBConfigSet, FloodAdd, FloodDel, FloodList in l2_test.go
- [ ] T062 [P] [US6] Write unit tests for DHCPList, DHCPStart, DHCPStop in dhcp_test.go
- [ ] T063 [P] [US6] Write unit tests for NexthopAdd, NexthopDel, NexthopGet, NexthopList, NexthopConfigGet, NexthopConfigSet in nexthop_test.go

### Implementation for User Story 6

- [ ] T064 [US6] Implement IP4Ping (send ICMP request, receive reply with timing) in ip4.go
- [ ] T065 [US6] Implement IP6Ping (send ICMPv6 request, receive reply) in ip6.go
- [ ] T066 [US6] Implement IP6RASet, IP6RAClear, IP6RAShow (streaming) in ip6.go
- [ ] T067 [US6] Define SRv6 types (SRv6EncapBehavior, SRv6Behavior, SRv6Flags) in srv6.go
- [ ] T068 [US6] Implement SRv6TunSrcSet, SRv6TunSrcClear, SRv6TunSrcShow in srv6.go
- [ ] T069 [US6] Define affinity types (RxQueueMap, CPUAffinity, CPUSet) in affinity.go
- [ ] T070 [US6] Implement AffinityRxQList (streaming), AffinityRxQSet, AffinityCPUGet, AffinityCPUSet in affinity.go
- [ ] T071 [US6] Implement LogLevelList (streaming), LogLevelSet, LogPacketsSet in log.go
- [ ] T072 [US6] Define L2 types (FDBEntry, FDBFlags, FDBConfig, FloodEntry, FloodVTEP, FloodType, FDBFlushRequest, FDBListRequest) with binary encoding in l2.go
- [ ] T073 [US6] Implement FDBAdd, FDBDel, FDBFlush, FDBList (streaming), FDBConfigGet, FDBConfigSet in l2.go
- [ ] T074 [US6] Implement FloodAdd, FloodDel, FloodList (streaming) in l2.go
- [ ] T075 [US6] Define DHCP types (DHCPStatus, DHCPState) in dhcp.go
- [ ] T076 [US6] Implement DHCPList (streaming), DHCPStart, DHCPStop in dhcp.go
- [ ] T077 [US6] Implement NexthopAdd, NexthopDel, NexthopGet, NexthopList (streaming), NexthopConfigGet, NexthopConfigSet in nexthop.go

**Checkpoint**: All grcli command categories implemented — full feature parity

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Integration tests, documentation, quality gates

- [ ] T078 [P] Write integration tests: connect to grout 0.15.0 container, create port (net_tap), add IPv4 address, add route, verify route lookup, clean up in integration_test.go
- [ ] T079 [P] Write integration tests: connect to grout 0.16.0 container, repeat same test suite, verify identical behavior in integration_test.go
- [ ] T080 [P] Write integration test for event subscription: subscribe, trigger interface add, verify event received, unsubscribe in integration_test.go
- [ ] T081 [P] Add package-level doc comment explaining library purpose and typical usage in grout.go
- [ ] T082 [P] Add Example functions for Connect, InterfaceAdd, IP4RouteAdd, EventSubscribe in example_test.go
- [ ] T083 Run `go vet ./...` and `staticcheck ./...`, fix any warnings
- [ ] T084 Run `go test -race -count=1 -coverprofile=coverage.out ./...`, verify ≥80% line coverage
- [ ] T085 Validate all exported symbols have doc comments per constitution requirement
- [ ] T086 Run quickstart.md validation: verify all code snippets compile against the library API

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **US1 Connect (Phase 3)**: Depends on Foundational — BLOCKS US2, US3
- **US2 Interfaces (Phase 4)**: Depends on US1
- **US3 IP/Routes (Phase 5)**: Depends on US2 (needs interface types for nexthop deserialization)
- **US7 Multi-Version (Phase 6)**: Depends on US3 (needs methods to version-gate)
- **US4 NAT (Phase 7)**: Depends on US3 (needs IP types and nexthop)
- **US5 Stats/Events (Phase 8)**: Depends on US1 (only needs connection)
- **US6 Advanced (Phase 9)**: Depends on US3 (needs IP types, nexthop, interface types)
- **Polish (Phase 10)**: Depends on all user stories

### User Story Dependencies

- **US1 (P1)**: Foundational → standalone
- **US2 (P1)**: US1 → standalone after connection works
- **US3 (P1)**: US2 → needs interface types for nexthop
- **US7 (P1)**: US3 → needs methods to add version checks to
- **US4 (P2)**: US3 → needs IP4Addr, nexthop types
- **US5 (P2)**: US1 → can start after connection works (parallel with US2-US4)
- **US6 (P3)**: US3 → needs IP types, nexthop, and interface types

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Types/structs before methods
- Simple operations (add/del) before complex (list/streaming)
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks T002-T005 can run in parallel
- T011 and T012 can run in parallel (protocol vs type tests)
- US5 (Stats/Events) can run in parallel with US2-US4 (only needs US1)
- All Phase 9 test tasks (T056-T063) can run in parallel
- All Phase 10 tasks (T078-T082) can run in parallel

---

## Parallel Example: User Story 2

```bash
# Launch tests first (must fail):
Task: "Write unit tests for InterfaceAdd/Del/Get/List/Set in iface_test.go"

# Then define types (parallel):
Task: "Define interface enums and Iface struct in iface_types.go"
Task: "Define interface type-specific info structs in iface_types.go"

# Then implement methods (sequential):
Task: "Implement InterfaceAdd, InterfaceDel, InterfaceGet in iface.go"
Task: "Implement InterfaceList, InterfaceSet in iface.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1-3 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks everything)
3. Complete Phase 3: US1 Connect
4. Complete Phase 4: US2 Interfaces
5. Complete Phase 5: US3 IP/Routes
6. **STOP and VALIDATE**: Can connect, manage interfaces, configure IP and routes
7. This is a usable library for basic network configuration

### Incremental Delivery

1. Setup + Foundational → protocol layer ready
2. US1 Connect → `grout.Connect()` works (MVP seed)
3. US2 Interfaces → add/list/configure/delete interfaces
4. US3 IP/Routes → full address and routing management
5. US7 Multi-Version → version safety across grout releases
6. US4 NAT → NAT configuration support
7. US5 Stats/Events → observability and monitoring
8. US6 Advanced → complete grcli parity

### Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently testable once its dependencies are met
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
