#!/usr/bin/env bash
# smoke.sh — Verify all services are healthy after deployment.
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
FAIL=0
MAX_RETRIES=10
SLEEP=2

probe() {
  local name=$1
  local url=$2
  for i in $(seq 1 $MAX_RETRIES); do
    if curl -sf --max-time 2 "$url" &>/dev/null; then
      printf "${GREEN}✓${NC} %s is healthy (%s)\n" "$name" "$url"
      return 0
    fi
    sleep $SLEEP
  done
  printf "${RED}✗${NC} %s not reachable after %d attempts (%s)\n" "$name" "$MAX_RETRIES" "$url"
  FAIL=1
}

echo "==> Running smoke tests..."
probe "api"    "http://localhost:8080/healthz"
probe "worker" "http://localhost:8081/healthz"

if [[ $FAIL -ne 0 ]]; then
  echo ""
  echo "Smoke tests failed. Check pod logs with: kubectl logs -l app=api"
  exit 1
fi

echo ""
echo "All smoke tests passed!"
