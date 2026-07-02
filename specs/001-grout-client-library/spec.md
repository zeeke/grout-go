# Feature Specification: Grout Go Client Library

**Feature Branch**: `001-grout-client-library`
**Created**: 2026-07-02
**Status**: Draft
**Input**: User description: "Build a Golang library to interact with a DPDK/grout instance. It has to support all the features grcli has, but no interactive terminal is needed at the moment. It has to deal with multiple grout versions."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Connect to a Grout Instance (Priority: P1)

A developer imports the grout-go library into their Go project and establishes a connection to a running grout daemon over its UNIX socket. The library handles the version handshake automatically and reports the negotiated API version.

**Why this priority**: Without a working connection and version negotiation, no other feature can function. This is the foundation for all operations.

**Independent Test**: Can be tested by connecting to a running grout daemon (or a mock socket) and verifying the handshake completes successfully, returning a usable client handle.

**Acceptance Scenarios**:

1. **Given** a running grout daemon at `/run/grout.sock`, **When** the developer calls `Connect("/run/grout.sock")`, **Then** a client handle is returned and the API version is negotiated successfully.
2. **Given** a custom socket path, **When** the developer calls `Connect("/tmp/my-grout.sock")`, **Then** the connection is established at the custom path.
3. **Given** no grout daemon is running, **When** the developer calls `Connect(...)`, **Then** a descriptive error is returned indicating the connection failed.
4. **Given** a grout daemon with an incompatible API version, **When** the developer calls `Connect(...)`, **Then** an error is returned indicating the version mismatch.

---

### User Story 2 - Manage Interfaces (Priority: P1)

A developer uses the library to create, list, configure, and delete network interfaces on a grout instance. This includes port interfaces, VLANs, bonds, bridges, VXLANs, IP-in-IP tunnels, and VRFs.

**Why this priority**: Interface management is the core building block for all network configuration. Routes, addresses, and NAT all depend on interfaces existing.

**Independent Test**: Can be tested by creating a port interface, listing interfaces, modifying interface flags (up/down), and deleting the interface — verifying each operation returns the expected result.

**Acceptance Scenarios**:

1. **Given** a connected client, **When** the developer adds a port interface with devargs and queue configuration, **Then** the interface is created and its ID is returned.
2. **Given** existing interfaces, **When** the developer lists interfaces (optionally filtered by type), **Then** all matching interfaces are returned with their full configuration.
3. **Given** an existing interface, **When** the developer sets its flags (e.g., UP), MTU, or VRF assignment, **Then** the change is applied and confirmed.
4. **Given** an existing interface, **When** the developer deletes it, **Then** the interface is removed and subsequent lookups return not-found.
5. **Given** an existing interface, **When** the developer manages its MAC addresses (add, delete, list, set primary), **Then** the MAC configuration is updated accordingly.

---

### User Story 3 - Configure IP Addresses and Routes (Priority: P1)

A developer uses the library to add/remove IPv4 and IPv6 addresses on interfaces and manage routing tables (add, delete, list, lookup routes) within VRFs.

**Why this priority**: IP addressing and routing are the primary use case for grout. Together with interfaces, they form the minimum viable network configuration.

**Independent Test**: Can be tested by adding an IP address to an interface, adding a route, performing a route lookup, and then cleaning up — verifying each step produces the expected state.

**Acceptance Scenarios**:

1. **Given** an interface, **When** the developer adds an IPv4/IPv6 address with prefix length, **Then** the address is assigned to the interface.
2. **Given** a VRF, **When** the developer adds a route with a destination prefix and nexthop, **Then** the route is installed in the routing table.
3. **Given** a routing table with entries, **When** the developer performs a longest-prefix-match lookup, **Then** the matching nexthop is returned.
4. **Given** a VRF, **When** the developer lists routes, **Then** all routes in that VRF are returned.
5. **Given** an interface with addresses, **When** the developer flushes all addresses, **Then** all addresses are removed from that interface.

---

### User Story 4 - NAT and Connection Tracking (Priority: P2)

A developer uses the library to configure static DNAT44 rules, dynamic SNAT44 with connection tracking, and inspect active connection tracking entries.

**Why this priority**: NAT is a critical network function but depends on interfaces and routing being functional first.

**Independent Test**: Can be tested by creating a DNAT44 rule, listing it, creating SNAT44 configuration, listing conntrack entries, and deleting the rules.

**Acceptance Scenarios**:

1. **Given** a configured interface, **When** the developer adds a static DNAT44 rule, **Then** the rule is installed and can be listed.
2. **Given** a configured interface, **When** the developer enables dynamic SNAT44, **Then** the SNAT configuration is active.
3. **Given** active NAT sessions, **When** the developer lists conntrack entries, **Then** the active connections are returned with their state.

---

### User Story 5 - Retrieve Statistics, Diagnostics, and Events (Priority: P2)

A developer uses the library to retrieve packet processing statistics (software and hardware counters), graph node information, packet traces, and subscribe to real-time events from the grout daemon.

**Why this priority**: Observability is essential for monitoring and debugging but is not required for basic network configuration.

**Independent Test**: Can be tested by retrieving software statistics, resetting counters, enabling packet tracing on an interface, dumping traces, and subscribing to interface events.

**Acceptance Scenarios**:

1. **Given** a connected client, **When** the developer requests software/hardware statistics, **Then** per-node packet counts, batches, and cycle stats are returned.
2. **Given** a connected client, **When** the developer resets statistics, **Then** all counters are zeroed.
3. **Given** an interface, **When** the developer enables packet tracing and later dumps the trace, **Then** captured packet details are returned.
4. **Given** a connected client, **When** the developer subscribes to events, **Then** interface state changes and route updates are delivered as they occur.
5. **Given** an active event subscription, **When** the developer unsubscribes, **Then** no further events are delivered.

---

### User Story 6 - Advanced Features: Ping, Traceroute, SRv6, Router Advertisements (Priority: P3)

A developer uses the library to send ICMP echo requests (ping), perform traceroutes, configure SRv6 tunnel source addresses, manage IPv6 router advertisement settings, configure CPU/queue affinity, and control logging.

**Why this priority**: These are specialized features that complete the grcli parity but are not required for core network configuration workflows.

**Independent Test**: Can be tested by sending a ping to a known address and verifying the response, configuring SRv6 tunnel source, and setting log levels.

**Acceptance Scenarios**:

1. **Given** a connected client and a reachable destination, **When** the developer sends an ICMP ping, **Then** the echo reply (or timeout) is returned with TTL and response time.
2. **Given** a connected client, **When** the developer sets the SRv6 tunnel source address, **Then** the configuration is applied.
3. **Given** a connected client, **When** the developer lists or sets log levels by pattern, **Then** the logging configuration is updated.
4. **Given** a connected client, **When** the developer configures CPU affinity for datapath and control plane, **Then** the affinity settings are applied.
5. **Given** a connected client, **When** the developer configures RX queue to CPU mappings, **Then** the queue affinity is updated.

---

### User Story 7 - Multi-Version Compatibility (Priority: P1)

A developer uses the library to interact with grout instances running different API versions. The library detects the remote version during handshake and adapts its behavior, clearly reporting when a requested feature is unavailable on the connected version.

**Why this priority**: The user explicitly requires multi-version support. Without this, the library would break when grout is upgraded or when managing heterogeneous deployments.

**Independent Test**: Can be tested by connecting to grout instances with different API versions and verifying that supported operations succeed while unsupported operations return clear "not supported in this version" errors.

**Acceptance Scenarios**:

1. **Given** a grout instance running API version N, **When** the developer calls a function available in version N, **Then** the operation succeeds.
2. **Given** a grout instance running API version N, **When** the developer calls a function introduced in version N+1, **Then** a descriptive error is returned indicating the feature requires a newer API version.
3. **Given** a connected client, **When** the developer queries the server API version, **Then** the negotiated version number is returned.

---

### Edge Cases

- What happens when the UNIX socket connection is lost mid-operation? The library MUST return a connection error and the client handle MUST be safe to close.
- What happens when the server sends a response for a different request ID? The library MUST buffer out-of-order responses and match them correctly.
- What happens when a streaming response is interrupted? The library MUST drain remaining messages to keep the socket in a clean state.
- What happens when the payload exceeds the 128 KiB maximum? The library MUST reject the request before sending with a clear error.
- What happens with concurrent access to a single client? The library MUST document that client handles are NOT thread-safe (not safe for concurrent goroutine access), consistent with the upstream C implementation.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The library MUST connect to a grout daemon over a UNIX domain socket and perform the GR_HELLO version handshake.
- **FR-002**: The library MUST support all interface types: port, VLAN, bond, bridge, VXLAN, IP-in-IP, VRF.
- **FR-003**: The library MUST support CRUD operations on interfaces (add, delete, get, list, set attributes).
- **FR-004**: The library MUST support IPv4 and IPv6 address management (add, delete, list, flush).
- **FR-005**: The library MUST support IPv4 and IPv6 route management (add, delete, get/lookup, list).
- **FR-006**: The library MUST support NAT configuration (DNAT44 static rules, SNAT44 dynamic, conntrack inspection).
- **FR-007**: The library MUST support statistics retrieval (software counters, hardware counters, per-node stats) and reset.
- **FR-008**: The library MUST support packet tracing (enable, disable, dump, clear).
- **FR-009**: The library MUST support event subscription and reception (interface events, route events, address events).
- **FR-010**: The library MUST support ICMP ping (send request, receive reply with timing).
- **FR-011**: The library MUST support log level management (list, set by pattern).
- **FR-012**: The library MUST support CPU and RX queue affinity configuration.
- **FR-013**: The library MUST support graph dump and configuration (burst size, vector size).
- **FR-014**: The library MUST support SRv6 tunnel source address configuration.
- **FR-015**: The library MUST support IPv6 router advertisement configuration.
- **FR-016**: The library MUST support MAC address management on interfaces (add, delete, list, set).
- **FR-017**: The library MUST handle multiple grout API versions, detecting the remote version during handshake and reporting unsupported features clearly.
- **FR-018**: The library MUST correctly implement the binary wire protocol: fixed-size headers, request/response matching by ID, streaming responses terminated by empty payload.
- **FR-019**: The library MUST expose an idiomatic API with proper error types, context support, and clean resource management (Close method on client).

### Key Entities

- **Client**: Represents a connection to a grout daemon. Holds the UNIX socket, negotiated API version, and request ID counter.
- **Interface**: A network interface managed by grout (port, VLAN, bond, bridge, VXLAN, IPIP, VRF) with its type-specific configuration.
- **Route**: An IPv4 or IPv6 routing table entry with destination prefix, nexthop, VRF, and origin.
- **Address**: An IPv4 or IPv6 address assigned to an interface, with prefix length.
- **Nexthop**: A forwarding target for routes, with gateway address, interface, and origin type.
- **Event**: A notification from grout about state changes (interface add/remove/status, route/address changes).
- **Stats**: Packet processing statistics per graph node (packets, batches, cycles).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can establish a connection to a grout instance and perform any configuration operation supported by grcli within 5 minutes of importing the library (clear API, good documentation, no hidden setup).
- **SC-002**: All grcli command categories are covered by the library: interface, address, route, NAT, stats, trace, events, ping, logging, affinity, graph, SRv6, router-advert, conntrack, MAC management.
- **SC-003**: The library correctly operates against at least 2 different grout API versions without code changes by the consumer.
- **SC-004**: Operations complete within the same time envelope as grcli (no meaningful overhead introduced by the library layer).
- **SC-005**: Connection errors, version mismatches, and unsupported features produce actionable error messages that identify the problem without requiring protocol knowledge.

## Assumptions

- The grout daemon is already running and accessible via a UNIX socket. The library does not start or manage the daemon lifecycle.
- The binary wire protocol (fixed headers, request IDs, streaming termination) is stable across API versions. Only message types and payloads change between versions.
- The default socket path is `/run/grout.sock`, consistent with the upstream C client.
- The library targets programmatic use only. Interactive terminal features (tab completion, command editing, pager) are explicitly out of scope.
- Traceroute and ping are implemented as protocol operations (send request, receive reply), not as shell-level utilities with formatted output.
- The library will use standard library facilities for UNIX socket communication. No CGo bindings to the C client library are needed since the wire protocol is documented and straightforward.
- Thread safety follows the upstream model: one client handle per goroutine. The library documents this constraint but does not add mutex synchronization internally.
