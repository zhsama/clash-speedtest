# Docker Setup Testing and Validation Strategy

## Overview

This document outlines a comprehensive testing strategy for validating the Docker setup of the clash-speedtest project. The tests cover build validation, runtime functionality, networking, performance, security, and CI/CD integration.

## 1. Build Validation Tests

### 1.1 Backend Build Tests

```bash
# Test 1: Clean build
docker build -t clash-speedtest-backend-test:latest ./backend
echo "✅ Backend build completed successfully"

# Test 2: Build cache effectiveness
time docker build -t clash-speedtest-backend-test:cache1 ./backend
time docker build -t clash-speedtest-backend-test:cache2 ./backend
# Second build should be significantly faster due to cache

# Test 3: Multi-architecture build
docker buildx build --platform linux/amd64,linux/arm64 -t clash-speedtest-backend-test:multiarch ./backend
echo "✅ Multi-architecture build successful"

# Test 4: Build args validation
docker build --build-arg VERSION=test-1.0.0 -t clash-speedtest-backend-test:buildargs ./backend
docker run --rm clash-speedtest-backend-test:buildargs --version | grep "test-1.0.0"
```

### 1.2 Frontend Build Tests

```bash
# Test 1: Clean build
docker build -t clash-speedtest-frontend-test:latest ./frontend
echo "✅ Frontend build completed successfully"

# Test 2: Production build validation
docker build --target production -t clash-speedtest-frontend-test:prod ./frontend
echo "✅ Production build completed successfully"

# Test 3: Development build validation
docker build --target development -t clash-speedtest-frontend-test:dev ./frontend
echo "✅ Development build completed successfully"

# Test 4: Build output validation
docker run --rm -v test-dist:/app/dist clash-speedtest-frontend-test:prod ls -la /app/dist
# Should show compiled frontend assets
```

## 2. Runtime Functionality Tests

### 2.1 Backend Runtime Tests

```bash
# Test 1: Basic execution
docker run --rm clash-speedtest-backend --help
echo "✅ Backend help command works"

# Test 2: Configuration file mounting
echo "proxies: []" > test-config.yaml
docker run --rm -v $(pwd)/test-config.yaml:/config/config.yaml clash-speedtest-backend -c /config/config.yaml
echo "✅ Configuration file mounting works"

# Test 3: Network connectivity test
docker run --rm --network bridge clash-speedtest-backend -c https://raw.githubusercontent.com/example/test-config.yaml
echo "✅ External network connectivity works"

# Test 4: Output directory mounting
docker run --rm -v $(pwd)/output:/output clash-speedtest-backend -c config.yaml -o /output/results.yaml
test -f output/results.yaml && echo "✅ Output file writing works"

# Test 5: Interactive mode
echo "q" | docker run -i --rm clash-speedtest-backend -c config.yaml -interactive
echo "✅ Interactive mode works"
```

### 2.2 Frontend Runtime Tests

```bash
# Test 1: Production server startup
docker run -d --name frontend-test -p 8080:80 clash-speedtest-frontend:prod
sleep 5
curl -I http://localhost:8080 | grep "200 OK"
docker stop frontend-test && docker rm frontend-test
echo "✅ Frontend production server works"

# Test 2: Static file serving
docker run -d --name frontend-test -p 8080:80 clash-speedtest-frontend:prod
sleep 5
curl http://localhost:8080/index.html | grep "<title>"
docker stop frontend-test && docker rm frontend-test
echo "✅ Static file serving works"

# Test 3: Development server with hot reload
docker run -d --name frontend-dev-test -p 3000:3000 -v $(pwd)/frontend/src:/app/src clash-speedtest-frontend:dev
sleep 10
curl -I http://localhost:3000 | grep "200"
# Modify a source file and verify hot reload
docker stop frontend-dev-test && docker rm frontend-dev-test
echo "✅ Development server with hot reload works"
```

## 3. Network Connectivity Tests

### 3.1 Docker Compose Network Tests

```bash
# Test 1: Service discovery
docker-compose up -d
docker-compose exec backend ping -c 1 frontend
docker-compose exec frontend wget -O- http://backend:8080/health
docker-compose down
echo "✅ Service discovery works"

# Test 2: Inter-service communication
cat > docker-compose.test.yml << EOF
version: '3.8'
services:
  backend:
    build: ./backend
    command: ["/app/download-server"]
    networks:
      - test-net
  frontend:
    build: ./frontend
    depends_on:
      - backend
    networks:
      - test-net
    environment:
      - API_URL=http://backend:8080
networks:
  test-net:
    driver: bridge
EOF

docker-compose -f docker-compose.test.yml up -d
sleep 5
docker-compose -f docker-compose.test.yml exec frontend curl http://backend:8080
docker-compose -f docker-compose.test.yml down
echo "✅ Inter-service communication works"
```

### 3.2 External Network Tests

```bash
# Test 1: DNS resolution
docker run --rm clash-speedtest-backend sh -c "nslookup google.com"
echo "✅ DNS resolution works"

# Test 2: HTTPS connectivity
docker run --rm clash-speedtest-backend sh -c "wget -O- https://api.github.com/rate_limit | head -n 5"
echo "✅ HTTPS connectivity works"

# Test 3: Proxy testing through Docker network
docker run --rm --network host clash-speedtest-backend -c https://example.com/config.yaml
echo "✅ Host network mode works for proxy testing"
```

## 4. Performance Validation

### 4.1 Image Size Tests

```bash
# Backend image size check
docker images clash-speedtest-backend --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"
# Expected: < 50MB for Alpine-based image

# Frontend production image size check
docker images clash-speedtest-frontend:prod --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"
# Expected: < 100MB for nginx + static files

# Frontend dev image size check
docker images clash-speedtest-frontend:dev --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"
# Expected: < 500MB with node_modules
```

### 4.2 Startup Time Tests

```bash
# Backend startup time
time docker run --rm clash-speedtest-backend --version
# Expected: < 1 second

# Frontend production startup time
time docker run -d --name perf-test clash-speedtest-frontend:prod
docker logs perf-test
docker stop perf-test && docker rm perf-test
# Expected: < 2 seconds

# Memory usage test
docker stats --no-stream --format "table {{.Container}}\t{{.MemUsage}}"
```

### 4.3 Build Performance Tests

```bash
# Clean build time measurement
docker system prune -af
time docker build -t clash-speedtest-backend:perf ./backend
# Record time for baseline

# Cached build time measurement
time docker build -t clash-speedtest-backend:perf2 ./backend
# Should be significantly faster

# Layer caching effectiveness
docker history clash-speedtest-backend:latest
# Verify efficient layer usage
```

## 5. Security Checks

### 5.1 User Permission Tests

```bash
# Test 1: Non-root user verification (Backend)
docker run --rm clash-speedtest-backend whoami
# Expected output: appuser (not root)

# Test 2: Non-root user verification (Frontend)
docker run --rm clash-speedtest-frontend:prod sh -c "ps aux | head -n 2"
# nginx should be running as nginx user, not root

# Test 3: File permissions
docker run --rm clash-speedtest-backend ls -la /app/
# Binary should be owned by appuser

# Test 4: Write permission test
docker run --rm clash-speedtest-backend sh -c "touch /app/test 2>&1 || echo '✅ Cannot write to /app (expected)'"
```

### 5.2 Vulnerability Scanning

```bash
# Using Trivy for vulnerability scanning
# Install trivy first: https://github.com/aquasecurity/trivy

# Scan backend image
trivy image clash-speedtest-backend:latest
# Review and address any HIGH or CRITICAL vulnerabilities

# Scan frontend image
trivy image clash-speedtest-frontend:prod
# Review and address any HIGH or CRITICAL vulnerabilities

# Scan base images
trivy image alpine:3.18
trivy image node:18-alpine
trivy image nginx:alpine
```

### 5.3 Security Best Practices Validation

```bash
# Test 1: No sensitive data in images
docker run --rm clash-speedtest-backend sh -c "env | grep -E '(PASSWORD|SECRET|KEY)'"
# Should return nothing

# Test 2: Minimal attack surface
docker run --rm clash-speedtest-backend sh -c "apk list 2>/dev/null | wc -l"
# Should show minimal packages

# Test 3: Read-only root filesystem
docker run --rm --read-only clash-speedtest-backend --version
# Should work with read-only filesystem

# Test 4: No unnecessary capabilities
docker run --rm --cap-drop=ALL clash-speedtest-backend --version
# Should work without any capabilities
```

## 6. Development Workflow Validation

### 6.1 Hot Reload Testing

```bash
# Frontend hot reload test
cat > test-hot-reload.sh << 'EOF'
#!/bin/bash
# Start dev container
docker run -d --name hot-reload-test \
  -p 3000:3000 \
  -v $(pwd)/frontend/src:/app/src \
  clash-speedtest-frontend:dev

sleep 10

# Get initial content
INITIAL=$(curl -s http://localhost:3000)

# Modify a source file
echo "console.log('Hot reload test');" >> frontend/src/App.js

sleep 5

# Get updated content
UPDATED=$(curl -s http://localhost:3000)

# Cleanup
docker stop hot-reload-test && docker rm hot-reload-test
git checkout frontend/src/App.js

# Verify change
if [ "$INITIAL" != "$UPDATED" ]; then
  echo "✅ Hot reload works"
else
  echo "❌ Hot reload failed"
fi
EOF

bash test-hot-reload.sh
```

### 6.2 Volume Mount Testing

```bash
# Test config file updates
docker run -d --name volume-test \
  -v $(pwd)/test-configs:/configs \
  clash-speedtest-backend sleep 3600

echo "proxies: [test1]" > test-configs/config1.yaml
docker exec volume-test cat /configs/config1.yaml | grep "test1"
echo "✅ Volume mounting works"

docker stop volume-test && docker rm volume-test
```

### 6.3 Development Tools Testing

```bash
# Test debugging capabilities
docker run --rm -it clash-speedtest-backend sh -c "which wget && which curl"
echo "✅ Basic debugging tools available"

# Test development dependencies (frontend)
docker run --rm clash-speedtest-frontend:dev npm list
echo "✅ Development dependencies installed"
```

## 7. CI/CD Integration Tests

### 7.1 GitHub Actions Simulation

```bash
# Simulate CI build
cat > .github/workflows/test-docker.yml << 'EOF'
name: Test Docker Build
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build Backend
        run: docker build -t test-backend ./backend
      - name: Build Frontend
        run: docker build -t test-frontend ./frontend
      - name: Run Backend Tests
        run: docker run --rm test-backend --version
      - name: Run Frontend Tests
        run: docker run --rm test-frontend:prod nginx -t
EOF

# Local simulation using act (https://github.com/nektos/act)
# act -j test
```

### 7.2 Multi-stage Build Validation

```bash
# Test build stages independently
# Backend stages
docker build --target builder -t test-builder ./backend
docker build --target final -t test-final ./backend

# Frontend stages
docker build --target dependencies -t test-deps ./frontend
docker build --target build -t test-build ./frontend
docker build --target production -t test-prod ./frontend
docker build --target development -t test-dev ./frontend
```

### 7.3 Registry Push Simulation

```bash
# Test image tagging and pushing (using local registry)
docker run -d -p 5000:5000 --name registry registry:2

# Tag and push
docker tag clash-speedtest-backend localhost:5000/clash-speedtest-backend:test
docker push localhost:5000/clash-speedtest-backend:test

# Pull and verify
docker pull localhost:5000/clash-speedtest-backend:test
docker run --rm localhost:5000/clash-speedtest-backend:test --version

# Cleanup
docker stop registry && docker rm registry
```

## 8. Integration Test Suite

Create an automated test script:

```bash
cat > run-all-tests.sh << 'EOF'
#!/bin/bash
set -e

echo "🧪 Starting Docker validation tests..."

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Test counter
PASSED=0
FAILED=0

# Test function
run_test() {
    echo -n "Testing: $1... "
    if eval "$2" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ PASSED${NC}"
        ((PASSED++))
    else
        echo -e "${RED}❌ FAILED${NC}"
        ((FAILED++))
    fi
}

# Run all tests
run_test "Backend Docker build" "docker build -t test-backend ./backend"
run_test "Frontend Docker build" "docker build -t test-frontend ./frontend"
run_test "Backend execution" "docker run --rm test-backend --help"
run_test "Frontend nginx config" "docker run --rm test-frontend:prod nginx -t"
run_test "Non-root user (backend)" "[[ $(docker run --rm test-backend whoami) == 'appuser' ]]"
run_test "Docker Compose" "docker-compose config"

# Summary
echo ""
echo "Test Summary:"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"

# Cleanup
docker rmi test-backend test-frontend 2>/dev/null || true

exit $FAILED
EOF

chmod +x run-all-tests.sh
```

## 9. Continuous Monitoring Tests

```bash
# Health check validation
docker run -d --name health-test \
  --health-cmd="wget -O- http://localhost:8080/health || exit 1" \
  --health-interval=30s \
  --health-timeout=3s \
  --health-retries=3 \
  clash-speedtest-backend /app/download-server

sleep 5
docker ps --filter name=health-test --format "table {{.Names}}\t{{.Status}}"
docker stop health-test && docker rm health-test
```

## 10. Troubleshooting Commands

```bash
# Debug build issues
docker build --no-cache --progress=plain -t debug-build ./backend

# Inspect image layers
docker history --no-trunc clash-speedtest-backend

# Debug running container
docker run --rm -it --entrypoint sh clash-speedtest-backend

# Check container logs
docker logs -f container-name

# Inspect container filesystem
docker run --rm -it clash-speedtest-backend find / -name "*.yaml" 2>/dev/null

# Network debugging
docker run --rm -it --cap-add=NET_ADMIN clash-speedtest-backend sh
# Inside container: ip addr, netstat -an, etc.
```

## Conclusion

This comprehensive testing strategy ensures:
- Build reliability across different environments
- Runtime functionality and performance
- Security compliance and best practices
- Development workflow efficiency
- CI/CD readiness

Run the automated test suite regularly and especially before releases to maintain Docker setup quality.