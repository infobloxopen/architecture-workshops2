# AI Agent Prompts for Resilience Patterns Workshop

Copy-paste these prompts into your AI coding assistant (GitHub Copilot, Claude, etc.) to get guided help with each lab case.

---

## Case 1: Timeouts

```
I'm working on a Go microservice workshop. The API service calls a dependency 
service using an HTTP client with no timeout configured. When the dependency 
is slow (3s+ response time), all requests pile up.

Look at these files for LAB: STEP1 TODO markers:
- pkg/depclient/client.go (HTTP client with no timeout)
- pkg/cases/timeout_case.go (context.Background() with no deadline)

Help me fix the timeout issue by:
1. Adding a timeout to the http.Client in depclient/client.go
2. Using context.WithTimeout in timeout_case.go
3. Using http.NewRequestWithContext to propagate the context

The goal is to bound p95 latency to ~2 seconds even when the dep service is slow.
```

---

## Case 2: DB Transaction Scope

```
I'm working on a Go microservice that has a database connection pool exhaustion 
problem. The handler opens a database transaction, locks a row with SELECT FOR 
UPDATE, then makes a slow HTTP call to an external service while holding the 
transaction open. Under load, all DB connections are consumed.

Look at this file for LAB: STEP2 TODO markers:
- pkg/cases/tx_case.go

The anti-pattern is: BEGIN TX → SELECT FOR UPDATE → HTTP call (2s) → UPDATE → COMMIT

Help me refactor so the slow network call happens OUTSIDE the transaction:
1. Make the HTTP call first
2. Then do a short transaction: BEGIN → SELECT FOR UPDATE → UPDATE → COMMIT
3. Keep transaction duration under 50ms

I can check pool stats at http://localhost:8080/debug/dbstats to verify the fix.
```

---

## Case 3: Bulkheads

```
I'm working on a Go worker service that processes batches of fast jobs (10ms) 
and slow jobs (1s). Currently they share a single goroutine pool with a 
semaphore of size 10. When slow jobs fill the pool, fast jobs are starved 
and their p95 spikes.

Look at this file for LAB: STEP3 TODO markers:
- pkg/worker/dispatcher.go (processBatch function)

Help me implement the bulkhead pattern:
1. Create separate semaphores: fastSem (size 50) and slowSem (size 5)
2. Route fast jobs through fastSem, slow jobs through slowSem
3. This ensures slow jobs can't consume more than 5 workers, leaving 
   plenty of capacity for fast jobs

The goal is fast p95 < 100ms while slow p95 stays at ~1000ms.
```

---

## Case 4: Autoscaling

```
I'm working on a Kubernetes deployment that should autoscale under CPU load, 
but the HPA never triggers. The /cases/autoscale endpoint does CPU-intensive 
work (SHA-256 hashing), but the pod never scales beyond 1 replica.

Look at these files for LAB: STEP4 TODO markers:
- deploy/k8s/api-deploy.yaml (no CPU resource requests)
- deploy/k8s/api-hpa.yaml (maxReplicas=1, averageUtilization=95)

Help me fix the autoscaling configuration:
1. Add CPU resource requests (100m) and limits (500m) to the api container
2. Set maxReplicas to 5 in the HPA
3. Lower averageUtilization to 50 so scaling triggers sooner

After changes: kubectl apply -f deploy/k8s/
Then verify: kubectl get hpa -w (should see replicas increase under load)
```

---

## Case 5: PDB & CNPG Failover

```
I'm working on a Kubernetes workshop using CloudNativePG (CNPG) for PostgreSQL.
The current setup has a single CNPG instance with no anti-affinity and no
automatic failover strategy. When the primary pod is killed (simulating a node
failure), writes are unavailable for 15+ seconds while the pod restarts.

Look at this file for LAB: STEP5 TODO markers:
- deploy/k8s/cnpg-cluster.yaml

The problems are:
1. instances: 1 — no replicas to fail over to
2. primaryUpdateStrategy is not set (defaults to supervised = manual intervention)
3. No anti-affinity — all pods could land on the same node

Help me fix the CNPG cluster configuration:
1. Set instances to 3 (1 primary + 2 replicas)
2. Add primaryUpdateStrategy: unsupervised for automatic failover
3. Add affinity with enablePodAntiAffinity: true and
   topologyKey: kubernetes.io/hostname to spread pods across nodes

After changes:
  kubectl apply -f deploy/k8s/cnpg-cluster.yaml
  kubectl wait --for=condition=Ready cluster/workshop-pg --timeout=180s
  kubectl get pods -l cnpg.io/cluster=workshop-pg -o wide

Verify pods are on different nodes, then re-run: go run ./cmd/driver run pdb
The goal is score 90+ with <5% error rate during fault injection.
```

---

## Case 6: Circuit Breaker

```
I'm working on a Kubernetes workshop using CloudNativePG (CNPG) for PostgreSQL.
The API has a /cases/circuitbreaker endpoint that writes to the CNPG database.
When the primary pod is killed (simulating a failure), every request blocks on a
30-second DB connection timeout, causing p95 latency to spike to 3+ seconds and
blocking all concurrency slots.

The project has a custom circuit breaker implementation in pkg/cases/breaker.go
with three states: Closed (normal), Open (failing fast), HalfOpen (probing).

Look at this file for LAB: STEP6 TODO markers:
- pkg/cases/circuitbreaker_case.go (breaker disabled with Threshold=0)

The breaker is created with Threshold: 0 and Timeout: 0, which disables it.
Fix: set Threshold to 5 (open after 5 consecutive failures) and Timeout to
5 * time.Second (probe for recovery after 5s).

After changes:
  make dev
  go run ./cmd/driver run circuitbreaker

The fix trades accuracy for responsiveness: error rate goes UP (breaker rejects
fast) but p95 drops from ~3374ms to ~8ms and throughput is maintained.
Target: score 70+ with p95 < 100ms.
```

---

## General Exploration

```
I'm exploring a Go microservice workshop that teaches 6 resilience patterns:
timeouts, DB transaction scope, bulkheads, autoscaling, PDB/CNPG failover,
and circuit breakers.

The project structure:
- cmd/lab/main.go — Single binary (api|worker|dep modes)
- cmd/driver/main.go — Load test driver with HTML reports
- pkg/cases/ — Lab case implementations with LAB TODO markers
- deploy/k8s/ — Kubernetes manifests (includes CNPG cluster)

Key commands:
- make up (create cluster + install CNPG operator) / make down (destroy)
- make dev (rebuild + redeploy < 30s)
- go run ./cmd/driver run <scenario> (run load test)
- go run ./cmd/driver list (list scenarios)

Help me understand the architecture and explore the codebase.
```
