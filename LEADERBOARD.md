# Workshop Leaderboard

## Scoring

Each case is scored 0–100 by the driver's built-in scorer. The total score is the sum of all six cases (max **600**).

| Metric           | Weight | Deduction                              |
|------------------|--------|----------------------------------------|
| Error rate       | 40 pts | -40 × (error_rate) if > 5%             |
| p95 latency      | 40 pts | -40 × (p95 / threshold) if > threshold |
| p99 latency      | 10 pts | -10 if p99 > 2× threshold              |
| Clean completion | 10 pts | Free points if no crashes               |

### Thresholds per Case

| Case | Scenario | p95 Target | Duration |
|------|----------|------------|----------|
| 1 — Timeouts   | `timeouts`  | 2500 ms | 30s |
| 2 — DB TX      | `tx`        | 2500 ms | 30s |
| 3 — Bulkheads  | `bulkheads` | 100 ms (fast) | 10s |
| 4 — Autoscale  | `autoscale` | 500 ms  | 60s |
| 5 — PDB & CNPG | `pdb`       | 2000 ms | 60s |
| 6 — Circuit Breaker | `circuitbreaker` | 2000 ms | 60s |

## How to Run

```bash
# Run a single scenario
go run ./cmd/driver run timeouts

# Run all scenarios and view reports
go run ./cmd/driver run timeouts
go run ./cmd/driver run tx
go run ./cmd/driver run bulkheads
go run ./cmd/driver run autoscale
go run ./cmd/driver run pdb
go run ./cmd/driver run circuitbreaker
```

Reports are saved to `reports/<scenario>/<timestamp>/`.

## Results

| Team       | Timeouts | DB TX | Bulkheads | Autoscale | PDB & CNPG | Circuit Breaker | Total | Time |
|------------|----------|-------|-----------|-----------|------------|-----------------|-------|------|
| _Example_  | 85       | 90    | 95        | 80        | 100        | 77              | 527   | 65m  |
|            |          |       |           |           |            |                 |       |      |
|            |          |       |           |           |            |                 |       |      |
|            |          |       |           |           |            |                 |       |      |
|            |          |       |           |           |            |                 |       |      |
|            |          |       |           |           |            |                 |       |      |

## Notes

- Scores are printed by the driver at the end of each run (look for `SCORE:` in output)
- Teams can re-run scenarios to improve their score
- The leaderboard is honor-based — facilitator validates final scores
- Tie-breaker: faster completion time wins
