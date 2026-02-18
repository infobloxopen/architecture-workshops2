# Implementation Plan: 60-Minute Resilience Patterns Workshop Framework

**Branch**: `001-workshop-framework` | **Date**: 2026-02-17 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-workshop-framework/spec.md`

## Summary

Build a self-contained workshop framework that teaches four resilience
patterns (timeouts, transaction scope, bulkheads, autoscaling) via a
repeatable "run scenario → observe failure → fix → rerun → compare"
loop. All services run locally in k3d. A single Go binary serves all
roles. A CLI driver generates HTML reports for each scenario.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: `net/http` (stdlib), `database/sql` + `lib/pq`, `k3d`
**Storage**: PostgreSQL (in-cluster, for the DB transaction case)
**Testing**: `go test` + make smoke targets
**Target Platform**: macOS (Docker Desktop + k3d), Linux secondary
**Project Type**: single monorepo
**Performance Goals**: dev loop < 30s; scenario run < 90s
**Constraints**: no cloud deps; no ingress; no Helm; single container image
**Scale/Scope**: 4 lab cases, ~10 Go source files, ~6 K8s manifests

## Constitution Check

| Principle | Status |
|-----------|--------|
| I. Simplicity First | ✅ Flat structure, single image, minimal deps |
| II. Local-First with k3d | ✅ k3d only, NodePort, offline after prefetch |
| III. Go Idiomatic | ✅ Go 1.24+, stdlib preferred, single `go build` |
| IV. Self-Contained Labs | ✅ Single `make demo` runs everything |

## Project Structure

### Documentation (this feature)

```text
specs/001-workshop-framework/
├── plan.md              # This file
├── spec.md              # Feature specification
├── checklists/
│   └── requirements.md  # Quality checklist
└── tasks.md             # Task list (generated next)
```

### Source Code (repository root)

```text
cmd/
├── lab/                 # Single binary entry point (api|worker|dep modes)
│   └── main.go
└── driver/              # Scenario driver CLI
    └── main.go

pkg/
├── api/                 # API service HTTP handlers
│   └── handler.go
├── worker/              # Worker service (batch processing)
│   └── dispatcher.go
├── dep/                 # Dependency simulator service
│   └── server.go
├── depclient/           # HTTP client for calling dep service
│   └── client.go
├── cases/               # Lab case implementations
│   ├── timeout_case.go  # Case 1: timeouts (LAB: STEP1 TODO)
│   ├── tx_case.go       # Case 2: DB transaction scope (LAB: STEP2 TODO)
│   ├── bulkhead_case.go # Case 3: bulkheads (LAB: STEP3 TODO)
│   └── autoscale_case.go# Case 4: autoscaling (API CPU endpoint)
├── driver/              # Driver load generator + metrics collector
│   ├── runner.go
│   ├── scenarios.go
│   └── scorer.go
└── report/              # HTML report generator
    ├── generator.go
    ├── templates/
    │   ├── report.html
    │   └── index.html
    └── data.go

deploy/
├── k3d-config.yaml      # k3d cluster config (port mappings)
└── k8s/
    ├── api-deploy.yaml   # API deployment + service (NodePort 30080→8080)
    ├── worker-deploy.yaml# Worker deployment + service (NodePort 30081→8081)
    ├── dep-deploy.yaml   # Dep deployment + service (ClusterIP)
    ├── postgres.yaml     # PostgreSQL StatefulSet + service
    ├── metrics-server.yaml# metrics-server for HPA
    └── api-hpa.yaml      # HPA manifest (LAB: STEP4 TODO)

scripts/
├── preflight.sh         # Tool checks
├── prefetch.sh          # Image pre-pull
├── smoke.sh             # Health check probes
└── reset.sh             # Wipe runtime state

docs/
├── LAB.md               # Step-by-step participant instructions
├── FACILITATOR.md       # Minute-by-minute facilitator guide
└── AGENT_PROMPTS.md     # AI agent prompts per case

reports/                 # Generated (gitignored)
├── index.html
└── <scenario>/<run-id>/
    ├── report.html
    └── data.json

Makefile                 # Top-level targets
Dockerfile               # Single multi-stage image
go.mod
go.sum
.gitignore
README.md
```

**Structure Decision**: Single-project layout. All Go services compile
into one binary (`cmd/lab/main.go`) that dispatches via subcommand
(`lab api`, `lab worker`, `lab dep`). The driver is a separate binary
(`cmd/driver/main.go`). Kubernetes manifests live in `deploy/k8s/`.

## Complexity Tracking

No constitution violations — no justification needed.
