# Docker Usage Guide

## Quick Start

### Production
```bash
# Build and run all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Development
```bash
# Run with hot reload
docker-compose -f docker-compose.dev.yml up

# Run only backend
docker-compose -f docker-compose.dev.yml up backend-dev

# Run only frontend
docker-compose -f docker-compose.dev.yml up frontend-dev
```

## Service Endpoints

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Download Server (optional): http://localhost:8081

## Running Specific Services

### Backend as Speed Tester
```bash
# Run speed test with config file
docker run --rm -v $(pwd)/config.yaml:/app/config.yaml clash-speedtest-backend -c /app/config.yaml

# Run with custom flags
docker run --rm clash-speedtest-backend -c https://example.com/config.yaml -concurrent 10
```

### Download Server Only
```bash
# Using docker-compose profile
docker-compose --profile download-server up download-server

# Or modify the backend service in docker-compose.yml
# Uncomment the entrypoint line for download-server
```

## Building Images

```bash
# Build backend
docker build -t clash-speedtest-backend ./backend

# Build frontend
docker build -t clash-speedtest-frontend ./frontend

# Build all with docker-compose
docker-compose build
```

## Environment Variables

### Backend
- `TZ`: Timezone (default: Asia/Shanghai)

### Frontend
- `VITE_API_URL`: Backend API URL (default: http://backend:8080)
- `TZ`: Timezone (default: Asia/Shanghai)

## Volume Mounts

### Backend
- `./backend/configs:/app/configs:ro` - Config files (read-only)
- `./backend/output:/app/output` - Output files

### Frontend (Development)
- `./frontend/src:/app/src:ro` - Source files
- `./frontend/public:/app/public:ro` - Public assets

## Networking

All services are connected via the `clash-speedtest-net` bridge network with subnet `172.20.0.0/16`.

## Resource Limits

### Backend
- CPU: 2 cores (limit), 0.5 cores (reservation)
- Memory: 512MB (limit), 128MB (reservation)

### Download Server
- CPU: 1 core (limit)
- Memory: 256MB (limit)

## Health Checks

The frontend service includes a health check that:
- Runs every 30 seconds
- Times out after 10 seconds
- Retries 3 times
- Waits 40 seconds before starting checks

## Security

- All services run as non-root users
- Minimal base images (Alpine Linux)
- Read-only volume mounts where appropriate
- No unnecessary packages installed

## Troubleshooting

### Permission Issues
If you encounter permission issues with volumes:
```bash
# Fix ownership
docker-compose exec backend chown -R appuser:appuser /app/output
```

### Build Cache
To rebuild without cache:
```bash
docker-compose build --no-cache
```

### Network Issues
If services can't communicate:
```bash
# Recreate network
docker-compose down
docker network prune
docker-compose up -d
```