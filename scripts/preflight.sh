#!/usr/bin/env bash
# preflight.sh — Check that all required tools are installed.
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
FAIL=0

check() {
  local cmd=$1
  if command -v "$cmd" &>/dev/null; then
    printf "${GREEN}✓${NC} %s found: %s\n" "$cmd" "$(command -v "$cmd")"
  else
    printf "${RED}✗${NC} %s not found\n" "$cmd"
    FAIL=1
  fi
}

echo "==> Checking prerequisites..."
check docker
check go
check kubectl
check k3d
check make

# Check Docker daemon is running
if docker info &>/dev/null; then
  printf "${GREEN}✓${NC} Docker daemon is running\n"
else
  printf "${RED}✗${NC} Docker daemon is not running — start Docker Desktop\n"
  FAIL=1
fi

# Check Go version >= 1.24
if command -v go &>/dev/null; then
  GO_VER=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | head -1 | sed 's/go//')
  GO_MAJOR=$(echo "$GO_VER" | cut -d. -f1)
  GO_MINOR=$(echo "$GO_VER" | cut -d. -f2)
  if [[ "$GO_MAJOR" -ge 1 && "$GO_MINOR" -ge 24 ]]; then
    printf "${GREEN}✓${NC} Go version %s (>= 1.24)\n" "$GO_VER"
  else
    printf "${RED}✗${NC} Go version %s (need >= 1.24)\n" "$GO_VER"
    FAIL=1
  fi
fi

if [[ $FAIL -ne 0 ]]; then
  echo ""
  echo "Some prerequisites are missing. Please install them and retry."
  exit 1
fi

echo ""
echo "All prerequisites satisfied!"
