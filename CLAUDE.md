# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a monorepo containing:
- **backend/**: Clash-SpeedTest - a command-line tool for testing proxy speeds using the Clash/Mihomo core
- **frontend/**: Frontend application (to be developed)

The backend directly reads Clash configuration files or subscription URLs and tests proxy performance without requiring a separate Clash process.

## Build and Development Commands

### Backend Development

```bash
# Navigate to backend directory
cd backend

# Build the main executable
go build -o clash-speedtest .

# Build with optimizations (smaller binary)
go build -ldflags="-s -w" -o clash-speedtest .

# Install globally
go install .

# Build the download server (for self-hosted speed test servers)
go build -o download-server ./download-server
```

### Running and Testing

```bash
# Run directly without building (from backend directory)
cd backend && go run . -c config.yaml

# Format code
cd backend && go fmt ./...

# Check for issues
cd backend && go vet ./...

# Update dependencies
cd backend && go mod tidy

# Download dependencies
cd backend && go mod download
```

### Release Process

The backend uses GoReleaser for automated releases via GitHub Actions:
- Releases are triggered by pushing tags (e.g., `git tag v1.0.0 && git push origin v1.0.0`)
- GoReleaser config is in `backend/.goreleaser.yaml`
- Builds for Linux, Windows, and macOS with CGO disabled

## Architecture and Code Structure

### Monorepo Structure

```
.
├── backend/          # Go backend (clash-speedtest)
│   ├── main.go
│   ├── go.mod
│   ├── speedtester/
│   └── download-server/
├── frontend/         # Frontend application (TBD)
└── CLAUDE.md
```

### Backend Components

1. **backend/main.go** - Entry point and CLI interface
   - Handles command-line flags and arguments
   - Orchestrates the speed testing process
   - Manages output formatting (table display, YAML export)
   - Implements filtering logic (speed, latency thresholds)

2. **speedtester package** (`backend/speedtester/speedtester.go`)
   - Core proxy loading and testing logic
   - Integrates with Mihomo (Clash core) for proxy handling
   - Implements concurrent speed testing with configurable workers
   - Handles proxy compatibility checks (e.g., Stash mode)
   - IP location lookup for node renaming feature

3. **download-server** (`backend/download-server/download-server.go`)
   - Optional self-hosted speed test server
   - Provides endpoints for download/upload testing
   - Alternative to using Cloudflare's speed test servers

### Key Design Patterns

1. **Proxy Handling**: Uses Mihomo's adapter system to support all Clash proxy types
2. **Concurrent Testing**: Worker pool pattern for parallel proxy testing
3. **Configuration Loading**: Supports both local files and HTTP(S) URLs, with comma-separated multiple sources
4. **Result Filtering**: Pipeline pattern for applying multiple filters (speed, latency, compatibility)

### External Dependencies

- **github.com/metacubex/mihomo**: Core Clash implementation for proxy handling
- **github.com/olekukonko/tablewriter**: Terminal table formatting
- **github.com/schollz/progressbar**: Progress display during testing
- **gopkg.in/yaml.v3**: YAML parsing for config files and output

### Important Implementation Details

1. **Config Loading**: The `-c` flag accepts comma-separated values for multiple config sources
2. **URL Subscriptions**: Must include `&flag=meta` parameter for proper node type recognition
3. **Speed Calculation**: Downloads specified size file (default 50MB) and calculates bandwidth
4. **Latency Measurement**: HTTP GET TTFB (Time To First Byte) measurement
5. **Node Renaming**: Uses ip-api.com for IP geolocation when `-rename` flag is used

## Common Development Tasks

### Adding New Command-Line Flags
1. Define flag in `backend/main.go` using `flag.*` functions
2. Add corresponding logic in the main loop
3. Update README.md with new flag documentation

### Modifying Speed Test Logic
1. Core testing logic is in `speedtester/TestProxies()` method
2. Individual proxy testing in `speedtester/TestProxy()` method
3. Download/upload implementations use custom `ZeroReader` for efficient testing

### Extending Proxy Support
- Proxy support comes from Mihomo - any updates to proxy types should be done upstream
- Compatibility checks can be added in `speedtester/isStashCompatible()` function

### Output Format Changes
- Table output formatting in `backend/main.go` using tablewriter
- YAML output uses the `ExportProxies()` method to generate Clash-compatible config

## Testing Approach

The project currently has no unit tests. When adding tests:

```bash
# Run tests for backend (when available)
cd backend && go test ./...

# Run tests with coverage
cd backend && go test -cover ./...

# Run tests with verbose output
cd backend && go test -v ./...
```

## Debugging and Logging

- The project uses Mihomo's logging system (`github.com/metacubex/mihomo/log`)
- Interactive debug mode available with `-interactive` flag
- Log levels can be controlled through Mihomo's configuration

## Code Quality Standards

### Before Committing

```bash
# Format all Go files
cd backend && go fmt ./...

# Run static analysis
cd backend && go vet ./...

# Ensure dependencies are correct
cd backend && go mod tidy

# Verify the build
cd backend && go build .
```

### Error Handling

- HTTP requests include proper timeout handling
- Concurrent operations use sync.WaitGroup for coordination
- Context cancellation for graceful shutdown

## Important Notes

1. **No Test Files**: The project currently has no unit tests. Consider adding tests when implementing new features.

2. **Security Considerations**:
   - When using URL subscriptions, ensure HTTPS is used
   - The `-rename` feature makes external API calls to ip-api.com for geolocation

3. **Performance Tips**:
   - Adjust `-concurrent` flag based on system resources
   - Use `-timeout` to skip slow proxies
   - Filter results with `-min-download-speed` and `-max-latency` for better results

4. **Compatibility**:
   - Supports all proxy types that Mihomo supports
   - The `-stash-compatible` flag filters proxies for Stash app compatibility

5. **Color Output**:
   - The tool uses ANSI color codes for terminal output
   - Colors are defined as constants in backend/main.go (colorRed, colorGreen, etc.)
