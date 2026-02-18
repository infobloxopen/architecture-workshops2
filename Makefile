CLUSTER_NAME := workshop
IMAGE_NAME   := lab:latest
K3D_CONFIG   := deploy/k3d-config.yaml

.PHONY: preflight prefetch up down dev smoke reset demo build

## ─── Pre-Work ───────────────────────────────────────────────

preflight:
	@bash scripts/preflight.sh

prefetch:
	@bash scripts/prefetch.sh

## ─── Cluster Lifecycle ──────────────────────────────────────

up: build
	@echo "==> Creating k3d cluster..."
	k3d cluster create --config $(K3D_CONFIG) 2>/dev/null || echo "Cluster already exists"
	@echo "==> Loading image into k3d..."
	k3d image import $(IMAGE_NAME) -c $(CLUSTER_NAME)
	@echo "==> Applying manifests..."
	kubectl apply -f deploy/k8s/
	@echo "==> Waiting for deployments..."
	kubectl rollout status deployment/api --timeout=60s
	kubectl rollout status deployment/worker --timeout=60s
	kubectl rollout status deployment/dep --timeout=60s
	@echo "==> Cluster ready!"

down:
	@echo "==> Deleting k3d cluster..."
	k3d cluster delete $(CLUSTER_NAME) 2>/dev/null || true
	@echo "==> Done."

## ─── Dev Loop ───────────────────────────────────────────────

build:
	@echo "==> Building Docker image..."
	docker build -t $(IMAGE_NAME) .

dev: build
	@echo "==> Loading image into k3d..."
	k3d image import $(IMAGE_NAME) -c $(CLUSTER_NAME)
	@echo "==> Restarting deployments..."
	kubectl rollout restart deployment/api deployment/worker deployment/dep
	@echo "==> Waiting for rollout..."
	kubectl rollout status deployment/api --timeout=60s
	kubectl rollout status deployment/worker --timeout=60s
	kubectl rollout status deployment/dep --timeout=60s
	@echo "==> Ready!"

## ─── Verification ───────────────────────────────────────────

smoke:
	@bash scripts/smoke.sh

## ─── Reset ──────────────────────────────────────────────────

reset:
	@bash scripts/reset.sh

## ─── Demo ───────────────────────────────────────────────────

demo: up smoke
	@echo "==> Running all scenarios..."
	go run ./cmd/driver run timeouts || true
	go run ./cmd/driver run tx || true
	go run ./cmd/driver run bulkheads || true
	go run ./cmd/driver run autoscale || true
	@echo "==> Demo complete. Check reports/ for results."
