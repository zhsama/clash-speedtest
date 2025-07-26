# Clash SpeedTest Makefile
# Simplified commands for common tasks

.PHONY: help install dev build test clean docker release docker-build-frontend docker-build-backend

# Variables
DOCKER_COMPOSE := docker compose
DOCKER_BUILDKIT := DOCKER_BUILDKIT=1
IMAGE_PREFIX := clash-speedtest
VERSION := $(shell git describe --tags --always --dirty)

# Default target
help:
	@echo "Clash SpeedTest - Available Commands:"
	@echo ""
	@echo "  make install    - Install dependencies"
	@echo "  make dev        - Start development servers"
	@echo "  make build      - Build all packages"
	@echo "  make test       - Run all tests"
	@echo "  make docker     - Build Docker images"
	@echo "  make release    - Build for release (includes Docker)"
	@echo "  make clean      - Clean all build artifacts"
	@echo ""
	@echo "Docker commands:"
	@echo "  make docker-build     - Build Docker images with BuildKit"
	@echo "  make docker-up        - Start services"
	@echo "  make docker-down      - Stop all services"
	@echo "  make docker-logs      - View container logs"
	@echo "  make docker-push      - Push images to registry"
	@echo "  make docker-clean     - Clean Docker resources"
	@echo ""
	@echo "Advanced targets:"
	@echo "  make build-cross   - Cross-platform builds"
	@echo "  make perf-test     - Run performance tests"

# Install dependencies
install:
	pnpm install --frozen-lockfile

# Development
dev:
	pnpm run dev

dev-frontend:
	pnpm run dev:frontend

dev-backend:
	pnpm run dev:backend

# Build targets
build:
	pnpm run build

build-all:
	pnpm run build:all

build-cross:
	cd backend && pnpm run build:cross

# Testing
test:
	pnpm run test

lint:
	pnpm run lint

typecheck:
	pnpm run typecheck

# Docker targets
docker: docker-build

# Build Docker images with BuildKit
docker-build:
	@echo "Building Docker images with BuildKit..."
	$(DOCKER_BUILDKIT) $(DOCKER_COMPOSE) build --parallel

# Start services
docker-up:
	@echo "Starting services..."
	$(DOCKER_COMPOSE) up -d
	@echo "Services started! Frontend: http://localhost:3000, Backend: http://localhost:8080"

# Stop all services
docker-down:
	@echo "Stopping all services..."
	$(DOCKER_COMPOSE) down

# View container logs
docker-logs:
	$(DOCKER_COMPOSE) logs -f

# Push images to registry
docker-push:
	@echo "Pushing images to registry..."
	$(DOCKER_BUILDKIT) $(DOCKER_COMPOSE) build --push

# Clean Docker resources
docker-clean:
	@echo "Cleaning Docker resources..."
	$(DOCKER_COMPOSE) down -v --remove-orphans
	docker image prune -f
	docker builder prune -f

# Quick Docker commands
docker-restart: docker-down docker-up

docker-rebuild: docker-clean docker-build docker-up

# Build with specific options
docker-build-backend:
	$(DOCKER_BUILDKIT) docker build \
		--build-arg ENABLE_UPX=true \
		--build-arg VERSION=$(VERSION) \
		-f backend/Dockerfile \
		-t $(IMAGE_PREFIX)-backend:$(VERSION) \
		backend/

docker-build-frontend:
	@echo "🏗️ 构建前端 Docker 镜像..."
	$(DOCKER_BUILDKIT) docker build \
		--build-arg VITE_API_URL=http://backend:8080 \
		-f frontend/Dockerfile \
		-t $(IMAGE_PREFIX)-frontend:latest \
		-t $(IMAGE_PREFIX)-frontend:$(VERSION) \
		.

# Release
release: test lint build docker
	@echo "✅ Release build complete!"

# Clean
clean:
	pnpm run clean
	pnpm run clean:cache
	docker image prune -f

clean-all: clean
	rm -rf node_modules
	rm -rf .pnpm-store
	docker system prune -af

# Performance testing
perf-test:
	@echo "Running performance tests..."
	@echo ""
	@echo "=== Build Performance ==="
	@time make build
	@echo ""
	@echo "=== Docker Build Performance ==="
	@time make docker-build
	@echo ""
	@echo "=== Docker Image Sizes ==="
	@docker images | grep -E "(REPOSITORY|$(IMAGE_PREFIX))" | head -5
	@echo ""
	@echo "=== Container Resource Usage ==="
	@docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep -E "(CONTAINER|clash-speedtest)"

# Docker image analysis
docker-analyze:
	@echo "Analyzing Docker images..."
	@echo ""
	@echo "=== Backend Image Layers ==="
	@docker history $(IMAGE_PREFIX)-backend:latest
	@echo ""
	@echo "=== Frontend Image Layers ==="
	@docker history $(IMAGE_PREFIX)-frontend:latest
	@echo ""
	@echo "=== Image Security Scan ==="
	@docker scout quickview $(IMAGE_PREFIX)-backend:latest || echo "Docker Scout not available"

# Development shortcuts
up: docker-up
down: docker-down
logs: docker-logs
ps:
	@docker compose ps
