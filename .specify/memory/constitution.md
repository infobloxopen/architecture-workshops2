<!--
Sync Impact Report
- Version change: N/A → 1.0.0
- Modified principles: N/A (initial creation)
- Added sections:
    - Core Principles (4 principles)
    - Technology Stack
    - Development Workflow
    - Governance
- Removed sections: N/A
- Templates requiring updates:
    - .specify/templates/plan-template.md — ✅ no updates needed (generic template)
    - .specify/templates/spec-template.md — ✅ no updates needed (generic template)
    - .specify/templates/tasks-template.md — ✅ no updates needed (generic template)
    - .specify/templates/checklist-template.md — ✅ no updates needed (generic template)
- Follow-up TODOs: none
-->

# Architecture Workshops Constitution

## Core Principles

### I. Simplicity First

Every lab MUST be understandable by a developer in under 10 minutes of
reading. No unnecessary abstractions, no over-engineered patterns.
Prefer flat directory structures and minimal file counts per lab.
If a concept requires more than one page of explanation, break it into
multiple labs instead of adding complexity to one.

**Rationale**: Workshops exist to teach, not to impress. Cognitive load
is the enemy of learning.

### II. Local-First with k3d

All labs MUST run locally using k3d as the Kubernetes runtime.
No cloud provider dependencies for core lab functionality.
Each lab MUST include a k3d cluster setup script that creates a
reproducible environment from scratch. Labs MUST work offline after
initial image pulls. Cluster creation and teardown MUST complete in
under 60 seconds on a modern laptop.

**Rationale**: Participants MUST NOT need cloud accounts, VPNs, or
external infrastructure to complete any lab. k3d provides a real
Kubernetes experience without the overhead.

### III. Go Idiomatic

All application code MUST use Go 1.24 or later. Prefer the standard
library over third-party dependencies. When external dependencies are
required, they MUST be well-maintained and widely adopted (e.g.,
`k8s.io/client-go`, `google.golang.org/grpc`). Code MUST pass
`go vet`, `golangci-lint`, and compile with zero warnings. Use Go
modules for dependency management. Every Go binary MUST build with a
simple `go build ./...` invocation.

**Rationale**: Go's standard library is rich enough for most workshop
scenarios. Minimal dependencies reduce setup friction and version
conflicts across participant machines.

### IV. Self-Contained Labs

Each lab MUST be independently runnable without completing prior labs.
A lab directory MUST contain everything needed: source code, manifests,
a `Makefile` (or `Taskfile`), and a `README.md` with step-by-step
instructions. Shared utilities MUST be vendored or referenced via Go
modules, never via relative path hacks. Each lab MUST have a
`make demo` target that runs the full happy-path demonstration
end-to-end.

**Rationale**: Participants join at different skill levels and may skip
labs. Instructors MUST be able to demonstrate any single lab in
isolation.

## Technology Stack

| Component       | Choice              | Constraint                        |
|-----------------|---------------------|-----------------------------------|
| Language        | Go 1.24+            | Standard library preferred        |
| Kubernetes      | k3d (k3s in Docker) | No cloud provider dependencies    |
| Container CLI   | Docker              | Podman acceptable as alternative  |
| Build Tool      | Make or Task        | One target per key action         |
| Testing         | `go test`           | Table-driven tests preferred      |
| Linting         | `golangci-lint`     | Default config unless justified   |
| Manifests       | Plain YAML or Kustomize | No Helm for lab simplicity   |
| Documentation   | Markdown            | One README.md per lab             |

## Development Workflow

1. **New lab creation**: Create a directory under `labs/` with a
   descriptive name (e.g., `labs/01-hello-k8s/`). Include `README.md`,
   `Makefile`, source code, and Kubernetes manifests.
2. **Testing**: Every lab MUST have at least one automated smoke test
   runnable via `make test`. The test MUST verify the lab's core
   learning outcome.
3. **Demo target**: Every lab MUST have `make demo` that sets up k3d,
   deploys the lab, runs the demonstration, and tears down cleanly.
4. **Review**: Lab PRs MUST include a screen recording or transcript
   of a successful `make demo` run.
5. **Naming**: Lab directories MUST be numbered and hyphen-separated
   (e.g., `01-topic-name`, `02-next-topic`).

## Governance

This constitution supersedes all other development practices for the
architecture-workshops2 repository. All pull requests MUST be verified
against these principles before merge. Amendments require:

1. A proposal describing the change and its rationale.
2. Review and approval by at least one project maintainer.
3. Version bump following semantic versioning:
   - **MAJOR**: Principle removal or backward-incompatible redefinition.
   - **MINOR**: New principle or materially expanded guidance.
   - **PATCH**: Clarifications, typo fixes, non-semantic refinements.
4. Update of `LAST_AMENDED_DATE` to the amendment date.

Complexity MUST be justified. When in doubt, simplify.

**Version**: 1.0.0 | **Ratified**: 2026-02-17 | **Last Amended**: 2026-02-17
