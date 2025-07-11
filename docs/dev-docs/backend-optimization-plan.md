# 后端节点测速优化方案

基于对 clash-speedtest-main 项目的分析，制定以下优化方案来改进我们的后端测速实现。

## 1. 测速方法优化

### 1.1 延迟测试增强
**现状问题**：
- 当前只进行单次延迟测试，结果可能不够准确
- 缺少抖动（Jitter）和丢包率统计

**优化方案**：
```go
// 新增延迟测试结构体
type LatencyResult struct {
    Average  float64 // 平均延迟
    Jitter   float64 // 抖动（标准差）
    LossRate float64 // 丢包率
    Min      float64 // 最小延迟
    Max      float64 // 最大延迟
}

// 实施多次测试（建议6次）
// 每次间隔100ms，避免突发
// 计算统计数据：平均值、标准差、丢包率
```

### 1.2 下载/上传测试优化
**现状问题**：
- 可能存在内存使用效率问题
- 缺少并发chunk下载支持

**优化方案**：
```go
// 实现ZeroReader优化
type ZeroReader struct {
    buf     []byte    // 固定1MB缓冲区
    read    int64     // 已读取字节数
    maxSize int64     // 最大读取大小
}

// 支持并发chunk下载
// 将总下载大小分成多个部分
// 使用多个goroutine并发下载
```

### 1.3 快速模式支持
**新增功能**：
- 添加 `-fast` 参数，只测试延迟
- 适用于快速检查大量节点可用性

## 2. 协议支持增强

### 2.1 扩展协议支持
**现状**：已支持大部分主流协议

**优化方案**：
- 确保支持所有Mihomo支持的协议
- 添加新协议：WireGuard, Tuic, SSH等
- 改进协议兼容性检查

### 2.2 改进Stash兼容性检查
```go
func isStashCompatible(proxy adapter.Proxy) bool {
    // 更详细的协议兼容性检查
    // 检查加密算法、传输层等
    switch proxy.Type() {
    case "shadowsocks":
        // 检查cipher是否被Stash支持
    case "vmess":
        // 检查加密和传输协议
    case "vless":
        // 检查flow类型
    // ... 其他协议
    }
}
```

## 3. 性能和用户体验优化

### 3.1 节点信息增强
**新增功能**：
- IP地理位置查询
- 自动添加国旗emoji
- 节点重命名支持

```go
type NodeInfo struct {
    Name      string
    Type      string
    Server    string
    Port      int
    Country   string  // 新增
    City      string  // 新增
    ISP       string  // 新增
    FlagEmoji string  // 新增
}
```

### 3.2 测试进度实时反馈
**优化方案**：
- 通过WebSocket实时推送每个节点的测试进度
- 包含当前测试阶段（延迟/下载/上传）
- 显示预计剩余时间

### 3.3 结果导出增强
**新增功能**：
- 支持更多导出格式（JSON、CSV）
- 支持导出详细测试数据
- 自动生成最优节点配置

## 4. 架构改进

### 4.1 模块化重构
```
backend/
├── speedtester/
│   ├── speedtester.go      // 主测试逻辑
│   ├── latency.go          // 延迟测试模块
│   ├── bandwidth.go        // 带宽测试模块
│   ├── protocols.go        // 协议处理
│   └── stats.go            // 统计计算
├── models/
│   ├── node.go             // 节点模型
│   └── result.go           // 结果模型
└── utils/
    ├── geo.go              // 地理位置
    └── export.go           // 导出功能
```

### 4.2 错误处理改进
- 实现重试机制
- 更细粒度的错误分类
- 优雅的降级处理

### 4.3 配置灵活性
- 支持测速服务器配置
- 支持自定义测试参数
- 支持批量配置源

## 5. API接口优化

### 5.1 RESTful API改进
```go
// 新增接口
GET  /api/test/:taskId/progress  // 获取测试进度
POST /api/test/batch            // 批量测试
GET  /api/nodes/geo/:nodeId     // 获取节点地理信息
```

### 5.2 WebSocket增强
```go
// 新增消息类型
type ProgressMessage struct {
    Type      string  // "latency", "download", "upload"
    NodeName  string
    Progress  float64
    ETA       int     // 预计剩余秒数
}
```

## 6. 实施优先级

1. **高优先级**（第一阶段）
   - 延迟测试增强（多次测试、统计数据）
   - ZeroReader内存优化
   - WebSocket进度细化

2. **中优先级**（第二阶段）
   - 节点地理位置支持
   - 快速模式
   - 并发chunk下载

3. **低优先级**（第三阶段）
   - 新协议支持
   - 批量测试API
   - 高级导出功能

## 7. 测试和验证

- 单元测试覆盖核心功能
- 性能基准测试
- 内存使用监控
- 兼容性测试

## 8. 向后兼容

- 保持现有API接口不变
- 新功能通过参数开关控制
- 配置文件格式兼容

这个优化方案将显著提升我们的测速准确性、性能和用户体验，同时保持良好的代码结构和可维护性。