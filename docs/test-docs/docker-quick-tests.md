# Quick Docker Test Commands

## Essential Quick Tests

### 1. Basic Build Test
```bash
# Quick build validation (both services)
docker build -t test-backend ./backend && echo "✅ Backend builds"
docker build -t test-frontend ./frontend && echo "✅ Frontend builds"
```

### 2. Runtime Test
```bash
# Test basic functionality
docker run --rm test-backend --version
docker run --rm test-frontend:prod nginx -t
```

### 3. Security Test
```bash
# Verify non-root execution
[[ $(docker run --rm test-backend whoami) == "appuser" ]] && echo "✅ Non-root user"
```

### 4. Docker Compose Test
```bash
# Test the full stack
docker-compose up -d
sleep 5
docker-compose ps
docker-compose down
```

### 5. Network Connectivity Test
```bash
# Test external network access
docker run --rm test-backend sh -c "wget -O- https://api.github.com/rate_limit | head -5"
```

## Performance Validation Commands

### Check Image Sizes
```bash
docker images | grep -E "(test-backend|test-frontend)" | awk '{print $1":"$2"\t"$7}'
```

### Memory Usage Test
```bash
# Start containers and check memory
docker run -d --name mem-test test-backend sleep 60
docker stats --no-stream mem-test
docker rm -f mem-test
```

## Development Workflow Tests

### Hot Reload Test (Frontend)
```bash
# Test development server
docker run -d --name dev-test -p 3000:3000 test-frontend:dev
sleep 10
curl -I http://localhost:3000
docker logs dev-test
docker rm -f dev-test
```

### Volume Mount Test
```bash
# Test configuration mounting
echo "test: config" > test.yaml
docker run --rm -v $(pwd)/test.yaml:/config/test.yaml test-backend cat /config/test.yaml
rm test.yaml
```

## CI/CD Simulation

### Multi-platform Build Test
```bash
# Test cross-platform builds (requires buildx)
docker buildx build --platform linux/amd64,linux/arm64 --load -t multiarch-test ./backend
```

### Build Cache Test
```bash
# Test build caching efficiency
time docker build -t cache-test1 ./backend
time docker build -t cache-test2 ./backend  # Should be much faster
```

## Troubleshooting Commands

### Debug Container
```bash
# Interactive debugging
docker run --rm -it --entrypoint sh test-backend
```

### Inspect Image Layers
```bash
# Check layer efficiency
docker history test-backend
```

### Check Logs
```bash
# View container logs
docker run --name log-test test-backend --help
docker logs log-test
docker rm log-test
```

## Automated Test Script

Run all tests with the provided script:
```bash
./test-docker.sh
```

Or run specific test categories:
```bash
# Just build tests
./test-docker.sh | grep -A1 "Build Validation"

# Just security tests
./test-docker.sh | grep -A10 "Security Tests"
```

## Clean Up

```bash
# Remove all test images and containers
docker rm -f $(docker ps -aq --filter "label=clash-speedtest-test") 2>/dev/null || true
docker rmi $(docker images -q --filter "reference=*test*") 2>/dev/null || true
```