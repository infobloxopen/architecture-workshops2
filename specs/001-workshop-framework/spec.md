# Feature Specification: 60-Minute Resilience Patterns Workshop Framework

**Feature Branch**: `001-workshop-framework`
**Created**: 2026-02-17
**Status**: Draft
**Input**: User description: "implement a 60 minute workshop framework"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Participant Completes Pre-Work Setup (Priority: P1)

A workshop participant receives setup instructions 1–2 days before the
session. They clone the repository, run a preflight check, pull required
container images, create a local Kubernetes cluster with the full lab
stack, and verify all services respond. By the end of pre-work the
participant has a working local environment and is ready for the live
session.

**Why this priority**: Without a working local environment, a participant
cannot complete any lab. This is the gate for everything else.

**Independent Test**: Run `make preflight`, `make prefetch`, `make up`,
and `make smoke` on a clean machine and verify all four succeed.

**Acceptance Scenarios**:

1. **Given** a Mac with Docker, Go, kubectl, k3d, and make installed,
   **When** the participant runs `make preflight`,
   **Then** all prerequisite checks pass and any missing tool is reported
   with install instructions.
2. **Given** preflight passes,
   **When** the participant runs `make prefetch`,
   **Then** all container images needed by the lab stack are pulled
   locally so no further downloads are needed during the live session.
3. **Given** images are prefetched,
   **When** the participant runs `make up`,
   **Then** a k3d cluster is created, all services and infrastructure
   are deployed, and the cluster reaches a ready state within 90 seconds.
4. **Given** the cluster is running,
   **When** the participant runs `make smoke`,
   **Then** HTTP probes confirm that the API service (port 8080) and the
   worker service (port 8081) respond successfully.

---

### User Story 2 — Participant Runs a Scenario and Views the Report (Priority: P1)

During the live session a participant runs a scenario driver command for
one of the four lab cases. The driver sends load to the lab services,
collects metrics, and generates an HTML report that opens automatically.
The report visualizes throughput, latency distribution, status codes,
and scenario-specific data so the participant can see the failure mode.

**Why this priority**: The run-then-observe loop is the core pedagogy of
the workshop. Without it there is no workshop.

**Independent Test**: With the cluster running, execute the driver for
any single case and verify a report is generated and contains meaningful
data.

**Acceptance Scenarios**:

1. **Given** the cluster is running with baseline (broken) code,
   **When** the participant runs the driver for the "timeouts" scenario,
   **Then** a report is generated within 90 seconds showing high tail
   latency and piled-up requests.
2. **Given** the cluster is running,
   **When** any scenario driver completes,
   **Then** an HTML report file is written under a timestamped
   directory, an index page listing all runs is updated, and the report
   auto-opens on Mac.
3. **Given** a generated report,
   **When** the participant views it,
   **Then** they see: RPS attempted vs succeeded, latency over time
   (p50/p95/p99), status code breakdown, and any scenario-specific
   panels.

---

### User Story 3 — Participant Fixes the Timeouts Case (Priority: P1)

The participant observes runaway latency caused by missing call
deadlines. Following the provided AI agent prompt (or manual
instructions), they add context deadlines and client timeouts to the
dependency-calling code. They rebuild, redeploy, rerun the scenario, and
compare the new report to the baseline.

**Why this priority**: This is the first and simplest case — it
establishes the "observe → fix → compare" cadence for the rest of the
workshop.

**Independent Test**: Apply the timeouts fix, rebuild, rerun the
scenario, and verify p95 drops below the defined threshold.

**Acceptance Scenarios**:

1. **Given** baseline code with no timeouts on dependency calls,
   **When** the participant runs the timeouts scenario,
   **Then** the report shows p95 latency well above acceptable limits
   and many requests hanging.
2. **Given** the participant applies the fix (context deadlines, request
   timeouts),
   **When** they rebuild/redeploy via the fast dev loop and rerun the
   scenario,
   **Then** p95 stays low and failed requests return quickly (no 30-
   second hangs).

---

### User Story 4 — Participant Fixes the DB Transaction Case (Priority: P2)

The participant observes throughput collapse caused by holding a database
transaction open while making a slow remote call. They refactor the
handler to commit the transaction before calling the dependency. Rebuild,
redeploy, rerun, compare.

**Why this priority**: Builds on the timeouts case but introduces a data
layer concern. Slightly more complex refactoring required.

**Independent Test**: Apply the transaction-scope fix, rebuild, rerun
the scenario, and verify DB pool wait count and p95 drop.

**Acceptance Scenarios**:

1. **Given** baseline code that calls a remote service inside an open
   DB transaction,
   **When** the participant runs the tx scenario under concurrency,
   **Then** the report shows saturated DB connections, high lock
   contention, and collapsed throughput.
2. **Given** the participant moves the remote call outside the
   transaction,
   **When** they rebuild/redeploy and rerun,
   **Then** DB pool wait count drops significantly and p95 improves.

---

### User Story 5 — Participant Fixes the Bulkheads Case (Priority: P2)

The participant observes that slow message processing starves fast
messages in a shared worker pool. They add separate pools/queues for
fast and slow work with concurrency caps. Rebuild, redeploy, rerun,
compare.

**Why this priority**: Introduces the bulkhead pattern — a distinct
concept from the first two cases but same fix cadence.

**Independent Test**: Apply the bulkhead fix, rebuild, rerun, and verify
fast completion p95 stays low even with slow jobs present.

**Acceptance Scenarios**:

1. **Given** baseline code using a single shared worker pool,
   **When** a mixed batch of fast and slow messages is submitted,
   **Then** the report shows fast message completion time degraded by
   slow messages.
2. **Given** the participant adds separate queues and caps slow
   concurrency,
   **When** they rebuild/redeploy and rerun,
   **Then** fast completion p95 stays low while slow work still
   completes.

---

### User Story 6 — Participant Fixes Autoscaling (Priority: P3)

The participant observes that pods do not scale under CPU-heavy load due
to missing resource requests and misconfigured HPA. They fix Kubernetes
manifests (resource requests/limits and HPA settings). Apply manifests,
rerun, compare.

**Why this priority**: This case is manifest-only (no code change) and
is the quickest fix — good as a closing exercise if time permits.

**Independent Test**: Apply manifest fixes, rerun, and verify replicas
increase during the test and latency improves.

**Acceptance Scenarios**:

1. **Given** baseline manifests with missing CPU requests and no HPA,
   **When** the participant runs the autoscale scenario,
   **Then** the report shows rising latency with no replica changes.
2. **Given** the participant sets proper CPU requests/limits and
   configures HPA,
   **When** they apply manifests and rerun,
   **Then** replicas increase during the run and p95 improves
   compared to baseline.

---

### User Story 7 — Facilitator Runs the Workshop End-to-End (Priority: P2)

A facilitator (instructor) guides participants through the entire 60-
minute session using a structured timeline. For each case they direct
participants to run a command, review the report, apply the fix, and
share their score. Solution branches exist as escape hatches for
participants who fall behind.

**Why this priority**: The facilitator experience drives the workshop
quality. Without clear pacing, solution fallbacks, and a run script the
session degrades.

**Independent Test**: A facilitator can execute the full 60-minute
timeline on their own, including fall-back to solution branches, without
encountering blockers.

**Acceptance Scenarios**:

1. **Given** the facilitator starts the session,
   **When** they follow the documented timeline,
   **Then** each case segment fits within its allotted time window
   (12 / 15 / 13 / 10 minutes respectively).
2. **Given** a participant falls behind,
   **When** they check out a solution branch for the current step,
   **Then** they can immediately rebuild and rejoin the current case.
3. **Given** all four cases are completed,
   **When** participants share scores,
   **Then** a leaderboard captures scores per scenario per participant.

---

### User Story 8 — Fast Inner Dev Loop (Priority: P1)

After editing code, the participant runs a single command to rebuild,
reload the image into k3d, and restart deployments. The full cycle
completes quickly so the observe-fix-compare loop stays tight.

**Why this priority**: A slow rebuild cycle breaks the cadence of a
60-minute workshop. This must be fast or the timing falls apart.

**Independent Test**: Edit a Go file, run the dev command, and measure
time from invocation to all pods ready.

**Acceptance Scenarios**:

1. **Given** the participant has edited a Go source file,
   **When** they run the dev loop command,
   **Then** the image is rebuilt, loaded into k3d, deployments are
   restarted, and all pods reach ready state within 30 seconds.
2. **Given** all Go services share a single container image,
   **When** the image is rebuilt,
   **Then** only one build and one image load are needed regardless of
   how many services are deployed.

---

### Edge Cases

- What happens when Docker Desktop is not running or has insufficient
  resources (< 4 CPU / < 6 GB RAM)?
- What happens when a participant's port 8080 or 8081 is already in use?
- What happens when `make up` is run twice (idempotent or error)?
- What happens when a participant skips pre-work and tries to set up
  during the live session (network-dependent image pulls)?
- What happens if the k3d cluster crashes mid-lab?
- What happens when the driver is run against a cluster that is not
  fully ready?
- What happens when a solution branch is checked out but the cluster
  still has the old image loaded?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a `make preflight` target that checks
  for required tools (Docker, Go, kubectl, k3d, make) and reports
  missing ones with install guidance.
- **FR-002**: System MUST provide a `make prefetch` target that pre-
  pulls all container images needed by the lab stack.
- **FR-003**: System MUST provide a `make up` target that creates a k3d
  cluster and deploys all services and infrastructure.
- **FR-004**: System MUST provide a `make smoke` target that verifies
  all services respond on their expected ports.
- **FR-005**: System MUST provide a `make dev` target that rebuilds the
  image, loads it into k3d, and restarts deployments.
- **FR-006**: System MUST provide a scenario driver that sends load to
  lab services and generates HTML reports with throughput, latency, and
  status code visualizations.
- **FR-007**: Each lab case MUST have clearly marked edit points in the
  source code (e.g., `LAB: STEP1 TODO`) indicating exactly where the
  participant should make changes.
- **FR-008**: Each lab case MUST have a corresponding solution available
  as a Git branch or tag (e.g., `step-1-solution`).
- **FR-009**: The driver MUST produce a single-line score output that
  participants can share for leaderboard purposes.
- **FR-010**: The report index page MUST list all runs for a scenario
  with timestamps and scores.
- **FR-011**: System MUST produce AI agent prompts for each lab case
  that participants can copy-paste into an AI assistant.
- **FR-012**: System MUST include a `make reset` target that wipes
  runtime state (DB data, worker queues) without destroying the cluster.
- **FR-013**: The lab stack MUST include: an API service, a worker
  service, a dependency simulator service, a database, and a metrics
  server.
- **FR-014**: All Go services MUST be compiled into a single binary
  that runs in different modes based on command arguments so only one
  container image is needed.
- **FR-015**: System MUST use k3d port mappings (NodePort) to expose
  services on localhost — no ingress controller required.
- **FR-016**: System MUST include facilitator documentation with a
  minute-by-minute timeline, talking points per case, and the commands
  to run at each step.
- **FR-017**: System MUST provide a `make down` target that destroys
  the k3d cluster and cleans up all resources.

### Key Entities

- **Lab Case**: A resilience failure scenario with a baseline (broken)
  state and a fix goal. Each case teaches one pattern (timeouts, tx
  scope, bulkheads, autoscaling).
- **Scenario Driver**: A load generator and metrics collector that runs
  against the lab stack, produces data, and generates reports.
- **Report**: An HTML visualization of a single scenario run showing
  throughput, latency, errors, and scenario-specific panels.
- **Score**: A single numeric value per scenario run that quantifies
  how well the fix performed, enabling comparison and leaderboarding.
- **Solution Branch**: A Git ref containing the completed fix for a
  specific lab case, usable as an escape hatch.
- **Lab Stack**: The set of services and infrastructure deployed in k3d
  that participants interact with during the workshop.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A participant with pre-work completed can run `make smoke`
  and see all services healthy within 2 minutes of starting the live
  session.
- **SC-002**: Each scenario driver run completes and produces a report
  in under 90 seconds.
- **SC-003**: The full dev loop (edit → build → load → restart → ready)
  completes in under 30 seconds.
- **SC-004**: Each lab case fix can be completed by a participant in
  under 5 minutes using the provided AI agent prompt.
- **SC-005**: The facilitator can deliver all four cases within a 60-
  minute session following the documented timeline.
- **SC-006**: 90% of participants who complete pre-work can finish at
  least 3 of 4 lab cases during the live session.
- **SC-007**: Every report clearly shows a measurable improvement (lower
  p95, fewer errors, higher throughput, or increased replicas) after
  the fix is applied vs the baseline run.
- **SC-008**: A participant who falls behind can check out a solution
  branch, rebuild, and rejoin the current case within 2 minutes.

## Assumptions

- Participants use macOS with Docker Desktop (Linux and Windows/WSL are
  secondary targets but not required for initial delivery).
- Participants have basic familiarity with Go, Docker, and Kubernetes
  concepts (this is not an introductory workshop).
- Network bandwidth during the live session is not relied upon — all
  images are pre-pulled during pre-work.
- The workshop is delivered remotely over video conferencing.
- k3d single-node clusters are sufficient for all lab scenarios.
- The AI agent prompts work with any major AI code assistant (GitHub
  Copilot, Cursor, etc.) but correctness does not depend on them —
  manual fixes are always documented as well.

## Dependencies

- k3d must be installed and functional on participant machines.
- Docker Desktop must have at least 4 CPU cores and 6 GB RAM allocated.
- Go 1.24+ toolchain must be installed.
- A metrics-server deployment must be available in k3d for the
  autoscaling case.
