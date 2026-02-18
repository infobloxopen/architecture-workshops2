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
