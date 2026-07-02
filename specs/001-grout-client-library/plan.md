# Implementation Plan: Grout Go Client Library

**Branch**: `001-grout-client-library` | **Date**: 2026-07-02 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/001-grout-client-library/spec.md`

## Summary

Build a pure-Go client library for the grout DPDK router daemon,
implementing the full binary wire protocol over UNIX domain sockets.
The library covers all grcli command categories (interfaces, routing,
NAT, L2, DHCP, stats, events, ping, SRv6, etc.) with multi-version
support and idiomatic Go API design. Testing uses unit tests for
protocol correctness and integration tests against containerized
grout instances.

## Technical Context

**Language/Version**: Go 1.22+ (latest stable minus one)
**Primary Dependencies**: Standard library only (`net`, `encoding/binary`, `syscall`)
**Storage**: N/A (stateless client library)
**Testing**: `go test` with `-race`, table-driven unit tests, integration tests with build tag
**Target Platform**: Linux (UNIX domain sockets; grout is Linux-only)
**Project Type**: Library
**Performance Goals**: Same latency envelope as grcli (sub-millisecond per operation over UNIX socket)
**Constraints**: No CGo, no external dependencies, payload ≤128 KiB
**Scale/Scope**: ~70 message types across 9 modules, ~100 exported API methods

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Code Quality & Clarity

- [x] All exported functions have doc comments → enforced by API contract
- [x] Functions do one thing → each method maps to one protocol message
- [x] Cyclomatic complexity ≤10 → protocol encode/decode is linear;
  streaming uses a single loop
- [x] `go vet` + `staticcheck` → CI gate
- [x] Explicit error handling → every method returns error; GrError
  wraps errno from daemon
- [x] No package-level globals → Client struct holds all state via
  dependency injection (socket path, connection)

### II. Testing Standards (NON-NEGOTIABLE)

- [x] Unit tests for all exported functions → table-driven per type
- [x] Test names: `Test<Function>_<Scenario>` pattern
- [x] Deterministic tests → mock UNIX socket for unit tests;
  no wall-clock, no network dependency
- [x] Integration tests → grout container with `//go:build integration`
- [x] `go test -race` → CI gate
- [x] Coverage ≥80% for new packages

### III. User Experience Consistency

- [x] Uniform naming: `New*` constructors, `Err*` sentinels, `*Request`
  for input structs
- [x] Consistent nil/zero-value behavior documented per method
- [x] godoc reads as coherent documentation → package doc + examples
- [x] API mirrors grcli command groups for discoverability

### IV. Backward Compatibility

- [x] Semver versioning from v0.1.0 (initial development)
- [x] Public API defined in contracts/api.md
- [x] Multi-version support via version-gated methods
- [x] No breaking changes without major version bump

**Gate result**: PASS — no violations.

## Project Structure

### Documentation (this feature)

```text
specs/001-grout-client-library/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── api.md           # Public API contract
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```text
grout.go                 # Package doc, Client, Connect/Close
protocol.go              # Wire protocol: headers, send/recv, streaming
types.go                 # Shared types: IP4Addr, IP6Addr, EtherAddr, enums
errors.go                # Error types: GrError, sentinel errors
version.go               # API version registry, ErrUnsupported checks

iface.go                 # Interface CRUD operations
iface_types.go           # Interface type-specific info structs
ip4.go                   # IPv4 address, route, ping, FIB
ip6.go                   # IPv6 address, route, ping, FIB, RA
nexthop.go               # Nexthop CRUD and config
nat.go                   # DNAT44, SNAT44
conntrack.go             # Connection tracking
l2.go                    # FDB, flood, bridge/VXLAN
dhcp.go                  # DHCP client start/stop/list
stats.go                 # Statistics, graph dump/config
trace.go                 # Packet tracing
events.go                # Event subscription and reception
affinity.go              # CPU and RX queue affinity
log.go                   # Log level management
srv6.go                  # SRv6 tunnel source

grout_test.go            # Client lifecycle tests
protocol_test.go         # Wire protocol encoding tests
types_test.go            # Type conversion tests
iface_test.go            # Interface operation tests
ip4_test.go              # IPv4 tests
ip6_test.go              # IPv6 tests
nexthop_test.go          # Nexthop tests
nat_test.go              # NAT tests
conntrack_test.go        # Conntrack tests
l2_test.go               # L2/FDB tests
dhcp_test.go             # DHCP tests
stats_test.go            # Stats tests
trace_test.go            # Trace tests
events_test.go           # Event tests
affinity_test.go         # Affinity tests
log_test.go              # Log tests
srv6_test.go             # SRv6 tests

integration_test.go      # Integration tests (//go:build integration)
testutil_test.go         # Shared test helpers (mock socket, assertions)

go.mod                   # Module definition
go.sum                   # Dependency checksums
```

**Structure Decision**: Single-package library at the repository root.
All source files are in the `grout` package. This follows the Go
convention for focused libraries (similar to `database/sql`,
`net/http`) and avoids unnecessary import paths for consumers.
Test files sit alongside source files per Go convention.

## Complexity Tracking

> No violations to justify — Constitution Check passed cleanly.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| (none)    |            |                                     |
