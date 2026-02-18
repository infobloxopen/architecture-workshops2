#!/usr/bin/env bash
# reset.sh — Wipe runtime state (DB data, worker batches) without destroying the cluster.
set -euo pipefail

echo "==> Resetting workshop state..."

# Reset PostgreSQL data if pod exists
if kubectl get pod -l app=postgres &>/dev/null 2>&1; then
  echo "  Resetting database..."
  kubectl exec -it deploy/postgres -- psql -U workshop -d workshop -c "DELETE FROM accounts;" 2>/dev/null || true
  echo "  Database reset."
fi

# Restart worker to clear in-memory batch state
echo "  Restarting worker..."
kubectl rollout restart deployment/worker
kubectl rollout status deployment/worker --timeout=60s

# Restart api to clear any cached state
echo "  Restarting api..."
kubectl rollout restart deployment/api
kubectl rollout status deployment/api --timeout=60s

echo ""
echo "==> Reset complete. All state cleared."
