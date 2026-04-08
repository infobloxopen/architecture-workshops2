#!/usr/bin/env bash
# prefetch.sh — Pre-pull container images so workshop runs offline.
set -euo pipefail

echo "==> Pre-pulling images for offline workshop use..."

# k3d node image (matches k3d default)
K3S_IMAGE="docker.io/rancher/k3s:v1.31.4-k3s1"
POSTGRES_IMAGE="docker.io/library/postgres:16-alpine"
METRICS_IMAGE="registry.k8s.io/metrics-server/metrics-server:v0.7.2"
CNPG_IMAGE="ghcr.io/cloudnative-pg/cloudnative-pg:1.25.1"
CNPG_PG_IMAGE="ghcr.io/cloudnative-pg/postgresql:16-alpine"

images=("$K3S_IMAGE" "$POSTGRES_IMAGE" "$METRICS_IMAGE" "$CNPG_IMAGE" "$CNPG_PG_IMAGE")

for img in "${images[@]}"; do
  echo "  Pulling $img ..."
  docker pull "$img" || echo "  Warning: could not pull $img (may already be cached)"
done

echo ""
echo "==> All images cached. Workshop can run offline."
