<!--
Sync Impact Report
==================
- Version change: 0.0.0 → 1.0.0 (initial ratification)
- Added principles:
  - I. Code Quality & Clarity
  - II. Testing Standards (NON-NEGOTIABLE)
  - III. User Experience Consistency
  - IV. Backward Compatibility
- Added sections:
  - Development Standards
  - Quality Gates
  - Governance
- Removed sections: none (initial version)
- Templates requiring updates:
  - .specify/templates/plan-template.md ✅ compatible (Constitution Check section exists)
  - .specify/templates/spec-template.md ✅ compatible (Requirements section exists)
  - .specify/templates/tasks-template.md ✅ compatible (test-first guidance present)
- Follow-up TODOs: none
-->

# grout-go Constitution

## Core Principles

### I. Code Quality & Clarity

All code MUST be idiomatic Go following established conventions
(Effective Go, Go Code Review Comments).

- Every exported function, type, and method MUST have a doc comment.
- Functions MUST do one thing. If a function name contains "And",
  it MUST be split.
- Cyclomatic complexity per function MUST NOT exceed 10. Functions
  exceeding this limit MUST be decomposed.
- All code MUST pass `go vet`, `staticcheck`, and project linting
  with zero warnings before merge.
- Error handling MUST be explicit — no silently discarded errors.
  Every returned error MUST be checked or deliberately ignored with
  a comment explaining why.
- Package-level globals MUST be avoided. Prefer dependency injection
  via function parameters or struct fields.

### II. Testing Standards (NON-NEGOTIABLE)

Every behavioral change MUST be accompanied by tests that
verify the intended behavior.

- Unit tests MUST cover all exported functions and methods. Table-driven
  tests MUST be used when a function has more than two meaningful input
  variations.
- Test names MUST follow the pattern `Test<Function>_<Scenario>` and
  clearly describe the behavior under test.
- Tests MUST be deterministic — no reliance on wall-clock time, network,
  or filesystem ordering. Use dependency injection to replace external
  dependencies with test doubles.
- Integration tests MUST exist for cross-package interactions, API
  contract boundaries, and any behavior involving I/O.
- `go test -race ./...` MUST pass. Any race condition is a blocking
  defect.
- Code coverage MUST NOT decrease on any PR. New packages MUST start
  at ≥80% line coverage.

### III. User Experience Consistency

The library's public API MUST present a coherent, predictable
interface to consumers.

- Naming conventions MUST be uniform across all packages: same concept,
  same name. A "builder" in one package MUST NOT be called "factory"
  in another.
- Constructor functions MUST follow `New<Type>` naming. Functional
  options MUST follow `With<Option>` naming when the options pattern
  is used.
- Error types MUST implement the `error` interface and SHOULD implement
  `Unwrap()` for wrapping. Sentinel errors MUST use `Err` prefix
  (e.g., `ErrNotFound`).
- All public APIs MUST behave consistently with respect to nil inputs,
  zero-value structs, and context cancellation. Document the contract
  explicitly if nil or zero-value behavior differs from the obvious
  default.
- godoc output MUST read as coherent documentation — not as afterthought
  annotations. Package-level doc comments MUST explain the package's
  purpose and typical usage.

### IV. Backward Compatibility

Public API changes MUST NOT break existing consumers without an
explicit, versioned migration path.

- Exported function signatures, type definitions, and interface
  contracts are part of the public API. Removing or changing them
  is a MAJOR version bump under semver.
- New functionality MUST be additive — new functions, new optional
  fields, new interfaces. Existing behavior MUST NOT change unless
  a bug fix demands it, in which case the fix MUST be documented
  in release notes.
- Deprecation MUST follow a two-release cycle: mark with
  `// Deprecated:` doc comment in release N, eligible for removal
  in release N+2 at the earliest.
- Go module versioning MUST follow semver strictly. Breaking changes
  MUST increment the major version path (`v2/`, `v3/`, etc.).
- Configuration defaults MUST NOT change between minor/patch releases.
  If a default must change, provide an explicit migration note.

## Development Standards

- **Language**: Go (latest stable release minus one supported).
- **Dependencies**: Minimize external dependencies. Every new
  dependency MUST be justified in the PR description. Prefer
  standard library solutions.
- **Formatting**: All code MUST be formatted with `gofmt`.
  No exceptions.
- **Documentation**: Package-level doc, exported symbol docs,
  and a root README with quickstart and examples.
- **Commit messages**: Conventional Commits format
  (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`).

## Quality Gates

All of the following MUST pass before a PR is merged:

1. `go build ./...` succeeds with zero warnings.
2. `go test -race -count=1 ./...` passes.
3. `go vet ./...` reports no issues.
4. Linter suite (`staticcheck`, project-configured linters) passes.
5. Coverage does not regress; new packages meet ≥80% threshold.
6. All exported symbols have doc comments.
7. No backward-incompatible changes without major version bump.

## Governance

This constitution is the authoritative source of development
principles for grout-go. It supersedes informal conventions,
ad-hoc decisions, and conflicting documentation.

- **Amendments**: Any change to this constitution MUST be proposed
  as a PR with rationale. Changes to Core Principles require
  explicit approval. All amendments MUST include a migration plan
  for any code or process that becomes non-compliant.
- **Versioning**: This constitution follows semantic versioning.
  MAJOR for principle removals or redefinitions, MINOR for new
  principles or material expansions, PATCH for clarifications
  and wording fixes.
- **Compliance**: All PRs and code reviews MUST verify compliance
  with these principles. Non-compliance MUST be flagged and
  resolved before merge.
- **Complexity justification**: Any deviation from simplicity
  (additional abstractions, external dependencies, non-standard
  patterns) MUST be justified in writing and approved.

**Version**: 1.0.0 | **Ratified**: 2026-07-02 | **Last Amended**: 2026-07-02
