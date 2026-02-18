# Tasks: 60-Minute Resilience Patterns Workshop Framework

**Input**: Design documents from `/specs/001-workshop-framework/`
**Prerequisites**: plan.md ✅, spec.md ✅

**Tests**: Not explicitly requested — test tasks omitted. Smoke tests are part of the framework itself (US1).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `cmd/`, `pkg/`, `deploy/`, `scripts/`, `docs/` at repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, Go module, single-binary skeleton

- [X] T001 Initialize Go module and create go.mod with `module github.com/infobloxopen/architecture-workshops2`
- [X] T002 Create single-binary entry point with subcommand dispatch (api|worker|dep) in cmd/lab/main.go
- [X] T003 [P] Create .gitignore with reports/, vendor/, and binary exclusions
- [X] T004 [P] Create README.md with project overview, pre-work instructions, and quick-start commands

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Implement dependency simulator service (configurable sleep/fail via query params) in pkg/dep/server.go
- [X] T006 Implement dep HTTP client (baseline: no timeouts, no context propagation) in pkg/depclient/client.go
- [X] T007 [P] Create k3d cluster config with port mappings (30080→8080, 30081→8081) in deploy/k3d-config.yaml
- [X] T008 [P] Create Dockerfile with multi-stage build (cached go mod download, distroless final) for single lab binary in Dockerfile
- [X] T009 Create API service HTTP handler skeleton with health endpoint in pkg/api/handler.go
- [X] T010 [P] Create worker service skeleton with health endpoint in pkg/worker/dispatcher.go
- [X] T011 Create Kubernetes Deployment + NodePort Service for api in deploy/k8s/api-deploy.yaml
- [X] T012 [P] Create Kubernetes Deployment + NodePort Service for worker in deploy/k8s/worker-deploy.yaml
- [X] T013 [P] Create Kubernetes Deployment + ClusterIP Service for dep in deploy/k8s/dep-deploy.yaml
- [X] T014 Create top-level Makefile with placeholder targets (preflight, prefetch, up, down, dev, smoke, reset, demo)

**Checkpoint**: Foundation ready — `go build ./cmd/lab/...` compiles, Dockerfile builds, k3d config valid

---

## Phase 3: US1 — Pre-Work Setup + US8 — Fast Dev Loop (Priority: P1) 🎯 MVP

**Goal**: Participants can create a cluster, deploy the stack, and iterate quickly

**Independent Test**: `make preflight && make prefetch && make up && make smoke` all succeed; `make dev` completes in < 30s

### Implementation

- [X] T015 [US1] Implement preflight tool-check script (docker, go, kubectl, k3d, make) in scripts/preflight.sh
- [X] T016 [US1] Implement prefetch script to pull k3d node, postgres, and metrics-server images in scripts/prefetch.sh
- [X] T017 [US1] Implement `make up` logic: k3d cluster create + kubectl apply all manifests + wait for ready in Makefile
- [X] T018 [US1] Implement `make smoke` health-check probes against localhost:8080 and localhost:8081 in scripts/smoke.sh
- [X] T019 [US1] Implement `make down` to delete k3d cluster and clean up in Makefile
- [X] T020 [P] [US1] Create PostgreSQL StatefulSet + Service manifest in deploy/k8s/postgres.yaml
- [X] T021 [P] [US1] Create metrics-server deployment manifest in deploy/k8s/metrics-server.yaml
- [X] T022 [US8] Implement `make dev` target: docker build → k3d image import → rollout restart → wait ready in Makefile
- [X] T023 [US1] Implement `make reset` script to wipe DB data and worker state without destroying cluster in scripts/reset.sh

**Checkpoint**: Full cluster lifecycle works — create, deploy, smoke, dev-loop rebuild, teardown

---

## Phase 4: US2 — Scenario Driver & Reports (Priority: P1)

**Goal**: Participants can run any scenario and view an HTML report showing failure patterns

**Independent Test**: `go run ./cmd/driver run timeouts` generates reports/timeouts/<run-id>/report.html

### Implementation

- [X] T024 [US2] Create driver CLI entry point with `run <scenario>` subcommand in cmd/driver/main.go
- [X] T025 [US2] Implement load generator (configurable RPS, duration, concurrency) in pkg/driver/runner.go
- [X] T026 [US2] Implement scenario registry mapping scenario names to target URLs and configs in pkg/driver/scenarios.go
- [X] T027 [US2] Implement metrics collection (latency histogram, status codes, throughput) in pkg/driver/runner.go
- [X] T028 [US2] Implement scoring function that produces single-line SCORE output in pkg/driver/scorer.go
- [X] T029 [US2] Create HTML report template with charts (RPS, latency p50/p95/p99, status codes) in pkg/report/templates/report.html
- [X] T030 [US2] Implement report generator that writes data.json + renders HTML from template in pkg/report/generator.go
- [X] T031 [US2] Create index.html template listing all runs per scenario with timestamps and scores in pkg/report/templates/index.html
- [X] T032 [US2] Implement index page updater and auto-open on Mac (`open` command) in pkg/report/generator.go
- [X] T033 [US2] Define report data structures (run metadata, latency buckets, status counts) in pkg/report/data.go

**Checkpoint**: Driver can hit any URL, collect data, generate a viewable HTML report with charts

---

## Phase 5: US3 — Timeouts Case (Priority: P1)

**Goal**: Case 1 baseline shows runaway latency; after fix, p95 stays bounded

**Independent Test**: Run timeouts scenario before/after fix — report shows clear improvement

### Implementation

- [X] T034 [US3] Implement timeout case handler in api service: calls dep with no deadline (baseline broken) in pkg/cases/timeout_case.go
- [X] T035 [US3] Register timeout case endpoint on api handler at /cases/timeouts in pkg/api/handler.go
- [X] T036 [US3] Add LAB: STEP1 TODO markers in pkg/depclient/client.go with comments explaining what to fix
- [X] T037 [US3] Configure timeouts scenario in driver (target URL, load profile, success thresholds) in pkg/driver/scenarios.go
- [X] T038 [US3] Add dep simulator slow/hang mode (query param `?sleep=5s`) used by timeouts scenario in pkg/dep/server.go

**Checkpoint**: `go run ./cmd/driver run timeouts` shows high p95; applying fix + `make dev` + rerun shows bounded p95

---

## Phase 6: US4 — DB Transaction Case (Priority: P2)

**Goal**: Case 2 baseline shows pool saturation from holding TX across network calls; after fix, throughput recovers

**Independent Test**: Run tx scenario before/after fix — DB wait count and p95 drop

### Implementation

- [X] T039 [US4] Add database connection pool setup to api service (connect to postgres) in pkg/api/handler.go
- [X] T040 [US4] Implement tx case handler: begin TX → lock row → call dep → commit (baseline broken) in pkg/cases/tx_case.go
- [X] T041 [US4] Register tx case endpoint on api handler at /cases/tx in pkg/api/handler.go
- [X] T042 [US4] Add /debug/dbstats endpoint exposing pool stats (InUse, WaitCount) on api in pkg/api/handler.go
- [X] T043 [US4] Add LAB: STEP2 TODO markers in pkg/cases/tx_case.go with comments explaining the anti-pattern
- [X] T044 [US4] Configure tx scenario in driver (target URL, load profile, DB stats polling) in pkg/driver/scenarios.go
- [X] T045 [US4] Add DB-specific panels (pool in-use, wait count) to report template in pkg/report/templates/report.html
- [X] T046 [US4] Create DB schema init script (simple table for lock/update testing) applied during make up in deploy/k8s/postgres.yaml

**Checkpoint**: `go run ./cmd/driver run tx` shows pool saturation; applying fix + `make dev` + rerun shows healthy pool

---

## Phase 7: US5 — Bulkheads Case (Priority: P2)

**Goal**: Case 3 baseline shows fast-job starvation from shared pool; after fix, fast p95 stays low

**Independent Test**: Run bulkheads scenario before/after fix — fast completion p95 stays bounded

### Implementation

- [X] T047 [US5] Implement batch submission endpoint (POST /batches) accepting {fast, slow} counts in pkg/worker/dispatcher.go
- [X] T048 [US5] Implement batch status endpoint (GET /batches/:id) returning progress + p95 per type in pkg/worker/dispatcher.go
- [X] T049 [US5] Implement single shared worker pool processing (baseline broken: slow starves fast) in pkg/worker/dispatcher.go
- [X] T050 [US5] Add LAB: STEP3 TODO markers in pkg/worker/dispatcher.go with comments explaining bulkhead pattern
- [X] T051 [US5] Configure bulkheads scenario in driver (submit batch, poll status, measure fast vs slow p95) in pkg/driver/scenarios.go
- [X] T052 [US5] Add bulkhead-specific panels (fast vs slow completion p95) to report template in pkg/report/templates/report.html

**Checkpoint**: `go run ./cmd/driver run bulkheads` shows fast starvation; applying fix + `make dev` + rerun shows isolation

---

## Phase 8: US6 — Autoscaling Case (Priority: P3)

**Goal**: Case 4 baseline shows no scaling under CPU load; after manifest fix, replicas increase

**Independent Test**: Run autoscale scenario before/after fix — replicas increase and p95 improves

### Implementation

- [X] T053 [US6] Implement CPU-heavy endpoint on api (e.g., tight loop / crypto work) at /cases/autoscale in pkg/cases/autoscale_case.go
- [X] T054 [US6] Register autoscale case endpoint on api handler in pkg/api/handler.go
- [X] T055 [US6] Create baseline api-deploy.yaml without CPU requests/limits (LAB: STEP4 TODO) in deploy/k8s/api-deploy.yaml
- [X] T056 [US6] Create baseline api-hpa.yaml with wrong/missing settings (LAB: STEP4 TODO) in deploy/k8s/api-hpa.yaml
- [X] T057 [US6] Configure autoscale scenario in driver (target URL, load profile, HPA polling for replicas) in pkg/driver/scenarios.go
- [X] T058 [US6] Add autoscale-specific panels (replicas over time, HPA desired vs current) to report template in pkg/report/templates/report.html

**Checkpoint**: `go run ./cmd/driver run autoscale` shows flat replicas; applying fix + rerun shows scale-up

---

## Phase 9: US7 — Facilitator Docs & AI Prompts (Priority: P2)

**Goal**: Complete documentation for running the workshop

**Independent Test**: Facilitator can follow docs and complete all 4 cases end-to-end

### Implementation

- [ ] T059 [P] [US7] Write participant lab guide with step-by-step commands and expected output per case in docs/LAB.md
- [ ] T060 [P] [US7] Write facilitator guide with minute-by-minute timeline and talking points in docs/FACILITATOR.md
- [ ] T061 [P] [US7] Write AI agent prompts for all 4 cases (copy-paste ready) in docs/AGENT_PROMPTS.md
- [ ] T062 [US7] Create solution branches/tags for each case (step-0, step-1-solution, step-2-solution, step-3-solution, step-4-solution)

**Checkpoint**: A facilitator can run the full 60-minute session using only the docs

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Final quality pass

- [ ] T063 [P] Add LEADERBOARD.md template and scoring instructions in LEADERBOARD.md
- [ ] T064 [P] Update README.md with final architecture diagram, full command reference, and troubleshooting in README.md
- [ ] T065 Validate full end-to-end: make preflight → prefetch → up → smoke → all 4 scenarios → down
- [ ] T066 [P] Add `make demo` target that runs setup + all scenarios + generates comparison reports in Makefile

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **US1+US8 (Phase 3)**: Depends on Phase 2 — cluster lifecycle + dev loop
- **US2 (Phase 4)**: Depends on Phase 2 — driver needs compiled services
- **US3 (Phase 5)**: Depends on Phase 3 (cluster) + Phase 4 (driver)
- **US4 (Phase 6)**: Depends on Phase 3 (cluster + postgres) + Phase 4 (driver)
- **US5 (Phase 7)**: Depends on Phase 3 (cluster) + Phase 4 (driver)
- **US6 (Phase 8)**: Depends on Phase 3 (cluster + metrics-server) + Phase 4 (driver)
- **US7 (Phase 9)**: Depends on Phases 5–8 (all cases must exist to document)
- **Polish (Phase 10)**: Depends on all prior phases

### User Story Independence

- **US3, US4, US5, US6** can proceed in parallel once Phase 3 + Phase 4 are done
- **US7** is documentation-only and can be drafted in parallel but finalized after cases

### Within Each User Story

- Case handler → endpoint registration → LAB TODO markers → driver scenario config → report panels

### Parallel Opportunities

Within Phase 2: T007, T008, T010, T012, T013 can all run in parallel
Within Phase 3: T020, T021 can run in parallel
Within Phase 5–8: All four case phases can run in parallel (different files)
Within Phase 9: T059, T060, T061 can all run in parallel

---

## Parallel Examples

### Phase 2 parallel batch
```
T007: k3d-config.yaml
T008: Dockerfile
T010: worker skeleton
T012: worker-deploy.yaml
T013: dep-deploy.yaml
```

### Case implementation parallel batch (after Phase 4)
```
T034–T038: Timeouts case (pkg/cases/timeout_case.go, pkg/depclient/client.go)
T039–T046: TX case (pkg/cases/tx_case.go)
T047–T052: Bulkheads case (pkg/worker/dispatcher.go)
T053–T058: Autoscale case (pkg/cases/autoscale_case.go, deploy/k8s/)
```

---

## Implementation Strategy

### MVP First (US1 + US8 + US2 + US3)

1. Complete Phase 1: Setup (go.mod, main.go, .gitignore)
2. Complete Phase 2: Foundational (dep simulator, client, manifests, Makefile)
3. Complete Phase 3: Cluster lifecycle + dev loop (make up/dev/smoke/down)
4. Complete Phase 4: Driver + reports (run scenario, view HTML)
5. Complete Phase 5: Timeouts case (first "observe → fix → compare")
6. **STOP and VALIDATE**: Full loop works for one case
7. Proceed to remaining cases

### Incremental Delivery

1. Setup + Foundation → binary compiles, image builds
2. Add US1+US8 → cluster works, dev loop works (infra MVP)
3. Add US2 → driver generates reports (pedagogy MVP)
4. Add US3 → first case end-to-end (workshop MVP — can deliver a 15-min demo)
5. Add US4, US5 → two more cases (can deliver 45-min workshop)
6. Add US6 → all four cases (full 60-min workshop)
7. Add US7 → facilitator docs (production-ready workshop)
8. Polish → leaderboard, troubleshooting, final validation

---

## Notes

- [P] tasks = different files, no dependencies
- [USn] label maps task to specific user story
- Single binary means all Go changes require one `make dev` cycle
- Report template is shared — scenario-specific panels are conditionally rendered
- LAB TODO markers must be present in baseline code; solution branches remove them
- Commit after each phase checkpoint
