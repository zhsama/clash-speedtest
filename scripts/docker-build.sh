#!/bin/bash

# Turborepo Docker Build Script
# This script is called by Turborepo to build Docker images

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Get the package name from the first argument
PACKAGE_NAME=$1
DOCKER_TAG=${2:-latest}
DOCKER_REGISTRY=${DOCKER_REGISTRY:-""}

echo -e "${BLUE}Building Docker image for $PACKAGE_NAME...${NC}"

# Function to build docker image
build_docker_image() {
    local package=$1
    local dockerfile_path=""
    local context_path=""
    local image_name=""
    
    case $package in
        "backend")
            dockerfile_path="backend/Dockerfile"
            context_path="backend"
            image_name="clash-backend"
            ;;
        "frontend")
            dockerfile_path="frontend/Dockerfile"
            context_path="frontend"
            image_name="clash-frontend"
            ;;
        *)
            echo -e "${YELLOW}Unknown package: $package${NC}"
            exit 1
            ;;
    esac
    
    # Full image name with registry if provided
    if [ -n "$DOCKER_REGISTRY" ]; then
        full_image_name="$DOCKER_REGISTRY/$image_name:$DOCKER_TAG"
    else
        full_image_name="$image_name:$DOCKER_TAG"
    fi
    
    echo -e "${GREEN}Building $full_image_name...${NC}"
    
    # Enable BuildKit for better performance
    export DOCKER_BUILDKIT=1
    
    # Build the image
    docker build \
        --build-arg BUILDKIT_INLINE_CACHE=1 \
        --cache-from "$full_image_name" \
        -f "$dockerfile_path" \
        -t "$full_image_name" \
        "$context_path"
    
    echo -e "${GREEN}✅ Successfully built $full_image_name${NC}"
    
    # Tag as latest if not already
    if [ "$DOCKER_TAG" != "latest" ]; then
        docker tag "$full_image_name" "$image_name:latest"
    fi
}

# Build the specified package
build_docker_image "$PACKAGE_NAME"

# If PUSH_TO_REGISTRY is set, push the image
if [ "$PUSH_TO_REGISTRY" = "true" ] && [ -n "$DOCKER_REGISTRY" ]; then
    echo -e "${BLUE}Pushing image to registry...${NC}"
    docker push "$DOCKER_REGISTRY/$image_name:$DOCKER_TAG"
    echo -e "${GREEN}✅ Successfully pushed to registry${NC}"
fi