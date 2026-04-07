# Architecture Workshops — Resilience Patterns

A 60-minute hands-on workshop teaching four resilience patterns through
a **run → observe failure → fix → rerun → compare** loop.

All services run locally in **k3d** (k3s-in-Docker). One Go binary,
one container image, fast iteration.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   k3d cluster                       │
│                                                     │
│  ┌──────────┐   ┌──────────┐   ┌──────────────────┐│
│  │   api    │──▶│   dep    │   │   PostgreSQL     ││
│  │ :30080   │   │ :8082    │   │   (StatefulSet)  ││
│  │          │──▶│          │   │                  ││
│  └──────────┘   └──────────┘   └──────────────────┘│
│  ┌──────────┐   ┌──────────┐                       │
│  │  worker  │   │ metrics- │                       │
│  │ :30081   │   │  server  │                       │
│  └──────────┘   └──────────┘                       │
└─────────────────────────────────────────────────────┘
         ▲              │
         │              │
  ┌──────┴──────┐       │
  │   driver    │       │    cmd/lab → single binary
  │ (load test) │       │    api | worker | dep modes
  └─────────────┘       │
```

**Single binary**: `cmd/lab/main.go` dispatches to `api`, `worker`, or `dep` mode.
**Driver**: `cmd/driver/main.go` generates load and produces HTML reports.

## Lab Cases

| # | Pattern | Fix Location |
|---|---------|-------------|
| 1 | Timeouts & Deadlines | `pkg/depclient/client.go` |
| 2 | DB Transaction Scope | `pkg/cases/tx_case.go` |
| 3 | Bulkheads | `pkg/worker/dispatcher.go` |
| 4 | Autoscaling (HPA) | `deploy/k8s/api-deploy.yaml` + `api-hpa.yaml` |
| 5 | PDB & CNPG Failover | `deploy/k8s/cnpg-cluster.yaml` |

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

## Project Structure

```
cmd/
  lab/main.go           # Single binary entry point (api|worker|dep)
  driver/main.go        # Load test driver with HTML reports
pkg/
  api/handler.go        # API server with case endpoints
  cases/
    timeout_case.go     # Case 1: Timeouts (LAB: STEP1)
    tx_case.go          # Case 2: DB TX scope (LAB: STEP2)
    autoscale_case.go   # Case 4: CPU-intensive work
  dep/server.go         # Dependency simulator
  depclient/client.go   # HTTP client (LAB: STEP1)
  worker/dispatcher.go  # Worker with batch processing (LAB: STEP3)
  driver/               # Load generator, scenarios, scorer
  report/               # HTML report generation
deploy/
  k3d-config.yaml       # k3d cluster configuration
  k8s/                  # Kubernetes manifests (LAB: STEP4)
scripts/                # preflight, prefetch, smoke, reset
docs/
  LAB.md                # Participant lab guide
  FACILITATOR.md        # Facilitator guide
  AGENT_PROMPTS.md      # AI assistant prompts for each case
```

## Solution Branches

Each case has a solution branch/tag you can reference:

| Tag | Branch | Description |
|-----|--------|-------------|
| `step-0` | `001-workshop-framework` | Baseline (all broken) |
| `step-1-solution` | `step-1-solution` | Timeouts fixed |
| `step-2-solution` | `step-2-solution` | DB TX scope fixed |
| `step-3-solution` | `step-3-solution` | Bulkheads fixed |
| `step-4-solution` | `step-4-solution` | Autoscaling fixed |

```bash
# View a solution
git diff step-0..step-1-solution

# Check out a solution
git checkout step-1-solution
```

## Troubleshooting

| Problem | Fix |
|---------|-----|
| `make preflight` fails on Docker | Start Docker Desktop, ensure 4+ CPU / 6+ GB RAM |
| `make up` hangs | Delete old cluster: `k3d cluster delete workshop` then retry |
| Services unreachable after `make up` | Run `make smoke` — wait for rollout. Check `kubectl get pods` |
| `make dev` — image not loading | Verify `k3d image import lab:latest -c workshop` succeeds |
| Driver shows 100% errors | Check `kubectl logs deploy/api` — may need `make reset` |
| HPA not scaling | Verify `kubectl get hpa` — needs CPU resources + metrics-server |
| DB connection errors | Check `kubectl get pods` for postgres — may need `kubectl rollout restart statefulset/postgres` |
| Port conflict on 8080/8081 | Stop other services or edit `deploy/k3d-config.yaml` port mappings |

## Documentation

- [Lab Guide](docs/LAB.md) — Step-by-step participant instructions
- [Facilitator Guide](docs/FACILITATOR.md) — Minute-by-minute workshop timeline
- [AI Prompts](docs/AGENT_PROMPTS.md) — Copy-paste prompts for AI assistants
- [Leaderboard](LEADERBOARD.md) — Scoring template
