# Resilience Patterns Workshop — Lab Guide

## Prerequisites

Before the workshop, run:

```bash
make preflight    # Checks: docker, go, kubectl, k3d, make
make prefetch     # Pulls images for offline use
```

## Quick Start

```bash
make up           # Create cluster + deploy all services
make smoke        # Verify everything is running
```

---

## Case 1: Timeouts (LAB: STEP1)

**Problem**: The API calls a dependency service with no timeout. When the dependency is slow, requests pile up and latency explodes.

### Observe the Problem

```bash
go run ./cmd/driver run timeouts
```

Open the generated report — notice p95 latency is ~3000ms+ (the dep sleep time).

### Find the Code

Look for `LAB: STEP1 TODO` markers in:
- `pkg/depclient/client.go` — HTTP client with no timeout
- `pkg/cases/timeout_case.go` — Context with no deadline

### Fix It

1. **Add HTTP client timeout** in `pkg/depclient/client.go`:
   ```go
   HTTPClient: &http.Client{Timeout: 2 * time.Second},
   ```

2. **Use context with deadline** in `pkg/cases/timeout_case.go`:
   ```go
   ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
   defer cancel()
   ```

3. **Use context in HTTP request** in `pkg/depclient/client.go`:
   ```go
   req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
   resp, err := c.HTTPClient.Do(req)
   ```

### Verify

```bash
make dev
go run ./cmd/driver run timeouts
```

The report should show p95 latency bounded at ~2000ms.

---

## Case 2: DB Transaction Scope (LAB: STEP2)

**Problem**: A database transaction is held open while making a slow network call. This exhausts the connection pool under load.

### Observe the Problem

```bash
go run ./cmd/driver run tx
```

Check DB pool stats: `curl http://localhost:8080/debug/dbstats`

Notice high `waitCount` and all connections `inUse`.

### Find the Code

Look for `LAB: STEP2 TODO` markers in:
- `pkg/cases/tx_case.go` — Network call inside TX

### Fix It

1. **Move the dep call outside the transaction**:
   ```go
   // Call dep FIRST (outside any transaction)
   _, depErr := depclient.Call(r.Context(), tc.DepClient, "2s", "0.0")

   // THEN do the short DB transaction
   tx, err := tc.DB.Begin()
   // ... query + update + commit
   ```

### Verify

```bash
make dev
go run ./cmd/driver run tx
```

DB wait count should drop significantly.

---

## Case 3: Bulkheads (LAB: STEP3)

**Problem**: Fast and slow jobs share a single worker pool. Slow jobs (1s each) block fast jobs (10ms each), causing fast job latency to spike.

### Observe the Problem

```bash
go run ./cmd/driver run bulkheads
```

Notice fast p95 is very high — fast jobs are starved by slow jobs.

### Find the Code

Look for `LAB: STEP3 TODO` markers in:
- `pkg/worker/dispatcher.go` — Single shared pool

### Fix It

1. **Create separate pools** for fast and slow jobs:
   ```go
   fastSem := make(chan struct{}, 50)   // Large pool for fast
   slowSem := make(chan struct{}, 5)    // Capped pool for slow
   ```

2. **Route jobs** to the appropriate pool.

### Verify

```bash
make dev
go run ./cmd/driver run bulkheads
```

Fast p95 should drop to ~50ms while slow p95 stays at ~1000ms.

---

## Case 4: Autoscaling (LAB: STEP4)

**Problem**: The API deployment has no CPU resource requests and the HPA is misconfigured, so it never scales under CPU load.

### Observe the Problem

```bash
go run ./cmd/driver run autoscale
```

Check replicas: `kubectl get hpa`

Notice replicas stay at 1.

### Find the Code

Look for `LAB: STEP4 TODO` markers in:
- `deploy/k8s/api-deploy.yaml` — Missing CPU resource requests
- `deploy/k8s/api-hpa.yaml` — Wrong target utilization

### Fix It

1. **Add CPU resources** to `deploy/k8s/api-deploy.yaml`:
   ```yaml
   resources:
     requests:
       cpu: 100m
       memory: 64Mi
     limits:
       cpu: 500m
       memory: 128Mi
   ```

2. **Fix HPA** in `deploy/k8s/api-hpa.yaml`:
   ```yaml
   minReplicas: 1
   maxReplicas: 5
   # Set averageUtilization to 50
   ```

3. **Apply changes**:
   ```bash
   kubectl apply -f deploy/k8s/
   ```

### Verify

```bash
go run ./cmd/driver run autoscale
kubectl get hpa -w    # Watch replicas increase
```

---

## Case 5: PDB & CNPG Failover (LAB: STEP5)

**Problem**: A CloudNativePG PostgreSQL cluster runs with a single instance. When the node is drained (maintenance, spot eviction), the only instance is evicted and writes fail for 30+ seconds while it restarts.

### Observe the Problem

```bash
go run ./cmd/driver run pdb
```

The driver sends steady writes for 60s and drains the primary's node at t=15s. With a single instance and no PDB, the report shows a long error window and high error rate.

### Find the Code

Look for `LAB: STEP5 TODO` markers in:
- `deploy/k8s/cnpg-cluster.yaml` — Single instance, no anti-affinity, supervised updates

### Fix It

1. **Scale to 3 instances** in `deploy/k8s/cnpg-cluster.yaml`:
   ```yaml
   instances: 3
   ```

2. **Enable unsupervised failover**:
   ```yaml
   primaryUpdateStrategy: unsupervised
   ```

3. **Enable pod anti-affinity** so instances spread across nodes:
   ```yaml
   affinity:
     enablePodAntiAffinity: true
     topologyKey: kubernetes.io/hostname
   ```

4. **Apply changes**:
   ```bash
   kubectl apply -f deploy/k8s/cnpg-cluster.yaml
   kubectl wait --for=condition=Ready cluster/workshop-pg --timeout=180s
   ```

### Verify

```bash
go run ./cmd/driver run pdb
```

With 3 instances + anti-affinity, draining the primary's node triggers automatic failover to a replica on another node. The report should show errors for only a few seconds (sub-5s recovery) and a much lower error rate.

---

## Case 6: Circuit Breaker (LAB: STEP6)

**Problem**: When the CNPG primary is killed, the application has no circuit breaker. Every request attempts a DB call that blocks on a 30-second connection timeout, causing p95 latency to spike to 3+ seconds and wasting all concurrency slots on doomed requests.

### Before You Start — Revert Case 5 Fixes

Case 6 demonstrates application-level resilience **without** infrastructure HA. Revert `cnpg-cluster.yaml` to a single instance so the breaker's value is clearly visible:

```yaml
instances: 1
# Remove or comment out primaryUpdateStrategy and affinity sections
```

```bash
kubectl apply -f deploy/k8s/cnpg-cluster.yaml
kubectl wait --for=condition=Ready cluster/workshop-pg --timeout=120s
kubectl rollout restart deployment/api
```

> **Why?** With 3 instances, CNPG failover is so fast the breaker barely activates. A single instance makes the outage window long enough to see the breaker in action.

### Observe the Problem

```bash
go run ./cmd/driver run circuitbreaker
```

The driver sends writes for 60s and kills the primary at t=15s (same fault as Case 5). Notice the p95 is ~3374ms — requests are blocking on connection timeouts instead of failing fast.

### Find the Code

Look for `LAB: STEP6 TODO` markers in:
- `pkg/cases/circuitbreaker_case.go` — Breaker created with `Threshold: 0` (disabled)
- `pkg/cases/breaker.go` — Circuit breaker state machine (Closed → Open → HalfOpen)

### Understand the Breaker

The circuit breaker has three states:
- **Closed** (normal): All requests pass through to the DB
- **Open** (failing fast): After N consecutive failures, requests are rejected immediately (<1ms)
- **HalfOpen** (probing): After a timeout, one request is allowed through to test recovery

### Fix It

1. **Enable the circuit breaker** in `pkg/cases/circuitbreaker_case.go`:
   ```go
   breaker: Breaker{
       Threshold: 5,              // Open after 5 consecutive failures
       Timeout:   5 * time.Second, // Probe for recovery after 5s
   },
   ```

### Verify

```bash
make dev
go run ./cmd/driver run circuitbreaker
```

The report should show:
- p95 drops from ~3374ms to ~8ms (fast rejection instead of blocking)
- Error rate goes UP (this is correct — the breaker rejects requests fast instead of letting them hang)
- Throughput is maintained at full capacity (600 requests vs 312 without breaker)
- Score improves to ~77/100

> **Note**: The scoring penalizes error rate, so the score won't be 100. The key insight is that the breaker *preserves throughput and latency* at the cost of accuracy — a fast "no" is better than a 30-second hang.

---

## Cleanup

```bash
make down    # Delete cluster
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `make up` fails | Run `make preflight` to check tools |
| Pods in CrashLoopBackOff | `kubectl logs -l app=api` |
| Stale image | `make dev` to rebuild and reload |
| DB connection errors | `make reset` to restart services |
| Port conflict on 8080 | `lsof -i :8080` and kill conflicting process |
| CNPG pod CrashLoopBackOff | WAL corrupted — `kubectl delete cluster workshop-pg && kubectl delete pvc -l cnpg.io/cluster=workshop-pg` then `kubectl apply -f deploy/k8s/cnpg-cluster.yaml` |
| CNPG pods stuck Pending | Need 3 nodes for anti-affinity — verify `kubectl get nodes` shows 3 nodes |
| API returns 503 on /cases/pdb | Stale DB connection — `kubectl rollout restart deployment/api` |
| Circuit breaker never opens | Verify `Threshold` is > 0 after the fix — `Threshold: 0` disables the breaker |
| CB score seems low with fix | Expected — breaker correctly rejects requests which count as errors. The p95 improvement is the key metric |
