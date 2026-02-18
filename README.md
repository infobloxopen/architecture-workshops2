# Architecture Workshops — Resilience Patterns

A 60-minute hands-on workshop teaching four resilience patterns through
a **run → observe failure → fix → rerun → compare** loop.

All services run locally in **k3d** (k3s-in-Docker). One Go binary,
one container image, fast iteration.

## Lab Cases

| # | Pattern | Fix Location |
|---|---------|-------------|
| 1 | Timeouts & Deadlines | `pkg/depclient/client.go` |
| 2 | DB Transaction Scope | `pkg/cases/tx_case.go` |
| 3 | Bulkheads | `pkg/worker/dispatcher.go` |
| 4 | Autoscaling (HPA) | `deploy/k8s/api-deploy.yaml` + `api-hpa.yaml` |

## Prerequisites

- Docker Desktop (4+ CPU, 6+ GB RAM)
- Go 1.24+
- kubectl
- k3d
- make

## Quick Start (Pre-Work)

```bash
git clone https://github.com/infobloxopen/architecture-workshops2.git
cd architecture-workshops2
make preflight   # check tools
make prefetch    # pull container images
make up          # create k3d cluster + deploy stack
make smoke       # verify services respond
```

## During the Workshop

```bash
# Run a scenario (generates HTML report)
go run ./cmd/driver run timeouts
go run ./cmd/driver run tx
go run ./cmd/driver run bulkheads
go run ./cmd/driver run autoscale

# After fixing code, rebuild + redeploy
make dev

# Rerun the scenario and compare reports
go run ./cmd/driver run timeouts
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make preflight` | Check required tools are installed |
| `make prefetch` | Pre-pull container images |
| `make up` | Create k3d cluster and deploy all services |
| `make down` | Destroy k3d cluster |
| `make dev` | Build image → load into k3d → restart pods |
| `make smoke` | Health-check all services |
| `make reset` | Wipe DB/worker state (keep cluster) |
| `make demo` | Full end-to-end demo run |

## Cleanup

```bash
make down
```
