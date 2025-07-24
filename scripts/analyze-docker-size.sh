#!/bin/bash

# Docker Image Size Comparison Script
# Compare different build strategies for size optimization

set -e

echo "🔍 Docker Image Size Optimization Analysis"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Build backend images with different strategies
echo -e "${BLUE}Building Backend Images...${NC}"
echo ""

# Strategy 1: Original Alpine-based image
echo -e "${YELLOW}1. Building original Alpine image...${NC}"
docker build -f backend/Dockerfile.original -t clash-backend:alpine backend/ 2>/dev/null || \
docker build -f backend/Dockerfile -t clash-backend:alpine backend/

# Strategy 2: Scratch-based with UPX compression
echo -e "${YELLOW}2. Building scratch + UPX image...${NC}"
docker build -f backend/Dockerfile -t clash-backend:scratch-upx backend/

# Strategy 3: Distroless image
echo -e "${YELLOW}3. Building distroless image...${NC}"
docker build -f backend/Dockerfile.distroless -t clash-backend:distroless backend/

# Build frontend images
echo ""
echo -e "${BLUE}Building Frontend Images...${NC}"
echo ""

# Strategy 1: Original nginx:alpine
echo -e "${YELLOW}1. Building original nginx:alpine image...${NC}"
docker build -f frontend/Dockerfile.original -t clash-frontend:nginx-alpine frontend/ 2>/dev/null || \
docker build -f frontend/Dockerfile -t clash-frontend:nginx-alpine frontend/

# Strategy 2: Optimized nginx:alpine-slim
echo -e "${YELLOW}2. Building optimized nginx:alpine-slim image...${NC}"
docker build -f frontend/Dockerfile -t clash-frontend:nginx-slim frontend/

echo ""
echo -e "${GREEN}Build Complete! Analyzing sizes...${NC}"
echo ""

# Function to get image size in MB
get_size_mb() {
    docker images --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}" | grep "$1" | awk '{print $2}'
}

# Function to get detailed layer info
get_layer_info() {
    docker history "$1" --format "table {{.CreatedBy}}\t{{.Size}}" --no-trunc | head -20
}

# Backend Analysis
echo "📦 Backend Image Sizes:"
echo "======================"
printf "%-30s %10s %15s\n" "Image" "Size" "Reduction"
printf "%-30s %10s %15s\n" "-----" "----" "---------"

# Get sizes
ALPINE_SIZE=$(docker images clash-backend:alpine --format "{{.Size}}")
SCRATCH_SIZE=$(docker images clash-backend:scratch-upx --format "{{.Size}}")
DISTROLESS_SIZE=$(docker images clash-backend:distroless --format "{{.Size}}")

# Calculate base size for comparison (using alpine as base)
ALPINE_MB=$(echo $ALPINE_SIZE | sed 's/MB//' | sed 's/GB/*1024/' | bc 2>/dev/null || echo "0")

printf "%-30s %10s %15s\n" "clash-backend:alpine" "$ALPINE_SIZE" "(baseline)"
printf "%-30s %10s %15s\n" "clash-backend:scratch-upx" "$SCRATCH_SIZE" "-"
printf "%-30s %10s %15s\n" "clash-backend:distroless" "$DISTROLESS_SIZE" "-"

echo ""
echo "📦 Frontend Image Sizes:"
echo "======================="
printf "%-30s %10s %15s\n" "Image" "Size" "Reduction"
printf "%-30s %10s %15s\n" "-----" "----" "---------"

# Get frontend sizes
NGINX_ALPINE_SIZE=$(docker images clash-frontend:nginx-alpine --format "{{.Size}}")
NGINX_SLIM_SIZE=$(docker images clash-frontend:nginx-slim --format "{{.Size}}")

printf "%-30s %10s %15s\n" "clash-frontend:nginx-alpine" "$NGINX_ALPINE_SIZE" "(baseline)"
printf "%-30s %10s %15s\n" "clash-frontend:nginx-slim" "$NGINX_SLIM_SIZE" "-"

# Size breakdown analysis
echo ""
echo "📊 Size Optimization Techniques Applied:"
echo "======================================="
echo ""
echo "Backend Optimizations:"
echo "- Multi-stage builds to minimize layers"
echo "- Static binary compilation (CGO_ENABLED=0)"
echo "- Strip debug symbols (-ldflags=\"-s -w\")"
echo "- UPX compression (--ultra-brute)"
echo "- Scratch/Distroless base images"
echo "- No package manager or shell"
echo ""
echo "Frontend Optimizations:"
echo "- Multi-stage builds"
echo "- Production dependencies only"
echo "- nginx:alpine-slim base"
echo "- Removed unnecessary nginx modules"
echo "- Optimized nginx configuration"
echo "- Gzip compression enabled"

# Security scan (optional)
echo ""
echo -e "${BLUE}Running security scans (requires trivy)...${NC}"
if command -v trivy &> /dev/null; then
    echo ""
    echo "🔒 Security Scan Results:"
    echo "========================"
    for image in clash-backend:scratch-upx clash-backend:distroless clash-frontend:nginx-slim; do
        echo ""
        echo "Scanning $image..."
        trivy image --quiet --severity HIGH,CRITICAL --format table "$image" 2>/dev/null || echo "Scan failed"
    done
else
    echo -e "${YELLOW}Trivy not found. Skip security scanning.${NC}"
fi

# Recommendations
echo ""
echo -e "${GREEN}💡 Recommendations:${NC}"
echo "=================="
echo ""
echo "1. For maximum size reduction:"
echo "   - Use scratch + UPX for backend (smallest but may have compatibility issues)"
echo "   - Use nginx:alpine-slim for frontend"
echo ""
echo "2. For better compatibility:"
echo "   - Use distroless for backend (small and secure)"
echo "   - Keep nginx:alpine-slim for frontend"
echo ""
echo "3. For development:"
echo "   - Use alpine-based images for easier debugging"
echo "   - Include development tools in dev-specific Dockerfiles"
echo ""
echo "4. Additional optimizations to consider:"
echo "   - Use BuildKit cache mounts for better caching"
echo "   - Implement multi-platform builds with --platform"
echo "   - Consider using Chainguard images for enhanced security"

# Cleanup (optional)
echo ""
read -p "Remove test images? (y/N) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker rmi clash-backend:alpine clash-backend:scratch-upx clash-backend:distroless \
               clash-frontend:nginx-alpine clash-frontend:nginx-slim 2>/dev/null || true
    echo "✅ Test images removed"
fi