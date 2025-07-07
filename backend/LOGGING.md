# Clash SpeedTest Logging Configuration

## Environment Variables

The logging system can be configured using environment variables:

### LOG_LEVEL
Controls the logging level. Valid values are:
- `DEBUG`: Most verbose, includes all log messages
- `INFO`: Default level, includes info, warn, and error messages  
- `WARN`: Only warnings and errors
- `ERROR`: Only error messages

Example:
```bash
export LOG_LEVEL=DEBUG
./speedtest-api-server
```

### Example Usage

```bash
# Run with debug logging
LOG_LEVEL=DEBUG ./speedtest-api-server

# Run with only error logging
LOG_LEVEL=ERROR ./speedtest-api-server

# Run with default (INFO) logging
./speedtest-api-server
```

## Log Format

Logs are output in JSON format for structured logging. Example log entry:

```json
{
  "time": "2024-01-01T12:00:00.000Z",
  "level": "INFO",
  "msg": "Speed test request received",
  "method": "POST",
  "path": "/api/test",
  "remote_addr": "127.0.0.1:12345"
}
```

## Log Levels by Component

### HTTP Server
- `INFO`: Server startup, shutdown, request completion
- `WARN`: Invalid requests, method not allowed
- `ERROR`: Server startup failures

### Speed Tester
- `INFO`: Proxy loading start/completion, test results summary
- `DEBUG`: Individual proxy tests, download/upload details
- `ERROR`: Configuration parsing errors, network failures

### Proxy Loading
- `INFO`: Config file processing, proxy counts
- `DEBUG`: HTTP requests, YAML parsing, filtering
- `ERROR`: File/network errors, parsing failures

## Performance Considerations

When using `DEBUG` level logging, expect increased log volume as each individual proxy test and network request will be logged. For production deployments, consider using `INFO` or `WARN` levels.