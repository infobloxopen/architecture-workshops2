# Facilitator Guide — 60-Minute Resilience Patterns Workshop

## Pre-Workshop Checklist

- [ ] Participants have Docker, Go 1.24+, kubectl, k3d, make installed
- [ ] Participants ran `make preflight && make prefetch`
- [ ] Projector/screen sharing ready
- [ ] This guide open on your screen

## Timeline

| Time | Duration | Activity |
|------|----------|----------|
| 0:00 | 5 min | Welcome + Architecture Overview |
| 0:05 | 5 min | Setup: `make up && make smoke` |
| 0:10 | 10 min | Case 1: Timeouts |
| 0:20 | 10 min | Case 2: DB Transaction Scope |
| 0:30 | 10 min | Case 3: Bulkheads |
| 0:40 | 10 min | Case 4: Autoscaling |
| 0:50 | 5 min | Leaderboard + Score Review |
| 0:55 | 5 min | Wrap-up + Q&A |

---

## 0:00–0:05 Welcome (5 min)

### Talking Points

- "Today we'll learn 4 resilience patterns by breaking and fixing real services"
- "Everything runs locally in k3d — no cloud accounts needed"
- "Each case follows: observe failure → find the code → fix → verify improvement"
- Draw the architecture: API → Dep (dependency simulator), Worker (batch processing), PostgreSQL

### Architecture Diagram

```
                    ┌─────────────┐
  HTTP :8080 ──────►│   API       │──────► Dep (:8082)
                    │             │──────► PostgreSQL
                    └─────────────┘
  HTTP :8081 ──────►│   Worker    │
                    └─────────────┘
```

---

## 0:05–0:10 Setup (5 min)

### Instructions

```bash
make up       # Creates k3d cluster, builds image, deploys
make smoke    # Verifies all services respond
```

**If someone has issues**: Most common is Docker not running. Have them run `make preflight`.

### What to Say

- "This creates a single-node k3d cluster with all our services"
- "You should see 'All smoke tests passed!'"
- "If stuck, run `make down && make up` to start fresh"

---

## 0:10–0:20 Case 1: Timeouts (10 min)

### Flow

1. **Run the baseline** (2 min):
   ```bash
   go run ./cmd/driver run timeouts
   ```
   "Look at the p95 — it's ~3 seconds. Every request waits for the full dependency sleep."

2. **Find the TODOs** (2 min):
   "Search for `LAB: STEP1 TODO` in your editor. You'll find them in `pkg/depclient/client.go` and `pkg/cases/timeout_case.go`."

3. **Fix and verify** (5 min):
   - Add `Timeout: 2 * time.Second` to the HTTP client
   - Use `context.WithTimeout` in the handler
   - Use `http.NewRequestWithContext` in the client
   ```bash
   make dev
   go run ./cmd/driver run timeouts
   ```

4. **Discuss** (1 min):
   - "What happens if we set the timeout too low?"
   - "Should timeout be on the client or the server?"

### Key Insight

> "Never call an external service without a timeout. The default `http.Client` has no timeout — it will wait forever."

---

## 0:20–0:30 Case 2: DB Transaction Scope (10 min)

### Flow

1. **Run the baseline** (2 min):
   ```bash
   go run ./cmd/driver run tx
   curl http://localhost:8080/debug/dbstats
   ```
   "Notice the `waitCount` — requests are queuing for a DB connection."

2. **Find the TODOs** (2 min):
   "Open `pkg/cases/tx_case.go`. The network call to dep is INSIDE the transaction."

3. **Fix and verify** (5 min):
   - Move the dep call before `tx.Begin()`
   - Keep the transaction as short as possible
   ```bash
   make dev
   go run ./cmd/driver run tx
   ```

4. **Discuss** (1 min):
   - "How would you detect this in production? (monitoring pool metrics)"
   - "What if you need data from the DB to make the network call?"

### Key Insight

> "Keep transactions as short as possible. Never make network calls while holding a DB connection."

---

## 0:30–0:40 Case 3: Bulkheads (10 min)

### Flow

1. **Run the baseline** (2 min):
   ```bash
   go run ./cmd/driver run bulkheads
   ```
   "Fast jobs should complete in ~10ms, but their p95 is much higher. Slow jobs are starving them."

2. **Find the TODOs** (2 min):
   "Open `pkg/worker/dispatcher.go`. There's one shared semaphore for both fast and slow."

3. **Fix and verify** (5 min):
   - Create separate `fastSem` and `slowSem` channels
   - Route fast jobs to `fastSem`, slow jobs to `slowSem`
   ```bash
   make dev
   go run ./cmd/driver run bulkheads
   ```

4. **Discuss** (1 min):
   - "This is the bulkhead pattern from ship design"
   - "Real-world: separate thread pools for critical vs. non-critical work"

### Key Insight

> "Isolate slow/unreliable work from fast/critical work. Don't let one bad dependency bring down everything."

---

## 0:40–0:50 Case 4: Autoscaling (10 min)

### Flow

1. **Run the baseline** (2 min):
   ```bash
   go run ./cmd/driver run autoscale
   kubectl get hpa
   ```
   "Replicas stay at 1 even under CPU load."

2. **Find the TODOs** (2 min):
   "Look in `deploy/k8s/api-deploy.yaml` (no CPU requests) and `deploy/k8s/api-hpa.yaml` (wrong settings)."

3. **Fix and verify** (5 min):
   - Add CPU resource requests to the deployment
   - Set `maxReplicas: 5` and `averageUtilization: 50`
   ```bash
   kubectl apply -f deploy/k8s/
   go run ./cmd/driver run autoscale
   kubectl get hpa -w
   ```

4. **Discuss** (1 min):
   - "Why do we need resource requests for HPA to work?"
   - "What's the trade-off between min/max replicas?"

### Key Insight

> "HPA needs CPU resource requests to calculate utilization. Without them, it can't make scaling decisions."

---

## 0:50–0:55 Leaderboard (5 min)

Have participants share their scores (printed by the driver). Best combined score across all 4 cases wins.

---

## 0:55–1:00 Wrap-up (5 min)

### Summary

| Pattern | Problem | Solution |
|---------|---------|----------|
| Timeouts | Unbounded waits | Set context deadlines |
| TX Scope | Pool exhaustion | Minimize TX duration |
| Bulkheads | Resource starvation | Isolate workloads |
| Autoscaling | No scale response | Configure HPA + requests |

### Resources

- "All code is in this repo — try modifying parameters and re-running"
- "Check `docs/AGENT_PROMPTS.md` for AI-assisted exploration"
- "Run `make down` to clean up when done"
