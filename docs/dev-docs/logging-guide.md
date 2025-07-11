# 日志配置文档

## 概述

Clash SpeedTest 后端支持灵活的日志配置，包括文件输出、控制台输出、日志轮转等功能。

## 默认配置

```go
// 默认日志配置
LogConfig{
    Level:           INFO,
    OutputToFile:    true,
    LogDir:          "logs",
    LogFileName:     "clash-speedtest.log", 
    MaxSize:         10MB,
    MaxFiles:        5,
    RotateOnStart:   true,
    EnableConsole:   true,
}
```

## 环境变量配置

可以通过以下环境变量覆盖默认配置：

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `LOG_LEVEL` | 日志级别 | INFO | DEBUG, INFO, WARN, ERROR |
| `LOG_DIR` | 日志目录 | logs | /var/log/clash-speedtest |
| `LOG_FILE` | 日志文件名 | clash-speedtest.log | speedtest.log |
| `LOG_TO_FILE` | 是否输出到文件 | true | true, false |
| `LOG_TO_CONSOLE` | 是否输出到控制台 | true | true, false |

## 日志轮转

### 自动轮转

- **启动时轮转**: 程序启动时会自动轮转现有日志文件
- **大小轮转**: 当日志文件超过 10MB 时自动轮转
- **文件保留**: 最多保留 5 个历史日志文件

### 手动轮转

通过 API 手动触发日志轮转：

```bash
curl -X POST http://localhost:8080/api/logs \
  -H "Content-Type: application/json" \
  -d '{"action": "rotate"}'
```

### 轮转文件命名

轮转后的文件格式：`clash-speedtest_20240101_120000.log`

## 日志级别

支持四个日志级别，从低到高：

1. **DEBUG**: 详细的调试信息，包含源码位置
2. **INFO**: 一般信息，程序正常运行的关键事件
3. **WARN**: 警告信息，可能的问题但不影响运行
4. **ERROR**: 错误信息，程序运行中的错误

### 运行时修改日志级别

```bash
curl -X POST http://localhost:8080/api/logs \
  -H "Content-Type: application/json" \
  -d '{"action": "set_level", "level": "DEBUG"}'
```

## 日志格式

### 控制台输出格式（TEXT）

```
time=2024-01-01T12:00:00.000Z level=INFO msg="HTTP Request" method=GET path=/api/health remote_addr=127.0.0.1:12345 status_code=200 duration=1.2ms
```

### 文件输出格式（JSON）

```json
{
  "time": "2024-01-01T12:00:00.000Z",
  "level": "INFO", 
  "msg": "HTTP Request",
  "method": "GET",
  "path": "/api/health",
  "remote_addr": "127.0.0.1:12345",
  "status_code": 200,
  "duration": "1.2ms"
}
```

## 日志类型

### HTTP 请求日志

记录所有 HTTP 请求的详细信息：

```json
{
  "time": "2024-01-01T12:00:00.000Z",
  "level": "INFO",
  "msg": "HTTP Request",
  "method": "POST",
  "path": "/api/test/async",
  "remote_addr": "192.168.1.100:45678",
  "status_code": 200,
  "duration": "150ms"
}
```

### 测速日志

记录代理测试过程的详细信息：

```json
{
  "time": "2024-01-01T12:00:00.000Z",
  "level": "INFO",
  "msg": "Proxy test completed successfully",
  "proxy_name": "香港节点1",
  "proxy_type": "vmess",
  "latency_ms": 45,
  "download_speed_mbps": 85.6,
  "upload_speed_mbps": 42.3,
  "packet_loss": 0.0
}
```

### 错误日志

记录系统错误和异常：

```json
{
  "time": "2024-01-01T12:00:00.000Z",
  "level": "ERROR",
  "msg": "Failed to load proxies",
  "error": "invalid configuration format",
  "config_paths": "config.yaml"
}
```

### WebSocket 日志

记录 WebSocket 连接和消息：

```json
{
  "time": "2024-01-01T12:00:00.000Z", 
  "level": "INFO",
  "msg": "WebSocket client connected",
  "client_id": "20240101120000-001",
  "total_clients": 1
}
```

## API 管理

### 获取日志配置

```bash
GET /api/logs
```

响应：
```json
{
  "success": true,
  "config": {
    "level": true,
    "file_logging": true,
    "console_logging": true,
    "log_dir": "logs",
    "log_file": "clash-speedtest.log",
    "max_size_mb": 10,
    "max_files": 5
  }
}
```

### 日志操作

```bash
POST /api/logs
```

请求体：
```json
{
  "action": "rotate"  // 或 "set_level"
}
```

或

```json
{
  "action": "set_level",
  "level": "DEBUG"
}
```

## 使用示例

### 开发环境配置

```bash
export LOG_LEVEL=DEBUG
export LOG_TO_CONSOLE=true
export LOG_TO_FILE=false
./clash-speedtest
```

### 生产环境配置

```bash
export LOG_LEVEL=INFO
export LOG_DIR=/var/log/clash-speedtest
export LOG_TO_CONSOLE=false
export LOG_TO_FILE=true
./clash-speedtest
```

### Docker 环境配置

```dockerfile
ENV LOG_LEVEL=INFO
ENV LOG_DIR=/app/logs
ENV LOG_TO_CONSOLE=true
ENV LOG_TO_FILE=true
VOLUME ["/app/logs"]
```

## 监控和分析

### 日志查看命令

```bash
# 实时查看日志
tail -f logs/clash-speedtest.log

# 查看最近的错误
grep "ERROR" logs/clash-speedtest.log | tail -10

# 统计不同级别的日志数量
grep -c "level=INFO\|level=WARN\|level=ERROR\|level=DEBUG" logs/clash-speedtest.log
```

### 日志分析工具

推荐使用以下工具分析 JSON 格式的日志：

- **jq**: 命令行 JSON 处理器
- **ELK Stack**: Elasticsearch + Logstash + Kibana
- **Grafana + Loki**: 现代日志聚合和可视化

### jq 查询示例

```bash
# 查看所有错误日志
cat logs/clash-speedtest.log | jq 'select(.level=="ERROR")'

# 统计各种代理类型的测试数量
cat logs/clash-speedtest.log | jq -r 'select(.proxy_type) | .proxy_type' | sort | uniq -c

# 查找延迟超过100ms的测试
cat logs/clash-speedtest.log | jq 'select(.latency_ms > 100)'
```

## 故障排除

### 常见问题

1. **日志文件无法创建**
   - 检查目录权限：`chmod 755 logs`
   - 确保磁盘空间充足

2. **日志轮转失败**
   - 检查文件权限
   - 确保没有其他进程占用日志文件

3. **日志级别不生效**
   - 重启应用程序
   - 检查环境变量设置

### 调试模式

启用 DEBUG 级别可以看到更详细的信息：

```bash
export LOG_LEVEL=DEBUG
```

这会显示：
- 源码文件和行号
- 详细的函数调用信息
- 网络请求的完整参数
- 内部状态变化

## 性能考虑

### 文件 I/O 优化

- 使用缓冲写入减少磁盘 I/O
- 异步日志轮转不阻塞主程序
- 批量写入减少系统调用

### 内存使用

- 日志缓冲区大小合理控制
- 及时清理历史日志文件
- 避免在高频路径记录详细日志

## 安全考虑

### 敏感信息

确保不在日志中记录：
- 用户密码和密钥
- 完整的代理配置
- 个人身份信息

### 访问控制

- 设置适当的文件权限（644）
- 限制日志目录访问
- 定期清理敏感日志