# Vless 协议支持优化需求文档

## 项目概述

本文档详细分析了 clash-speedtest 后端对 vless 协议的支持现状，并提出针对性的优化方案，以解决当前 vless 节点测试全部失败的问题。

## 1. 现状分析

### 1.1 技术架构分析

**当前使用的核心依赖：**
- mihomo v1.19.10 - 主要的代理核心
- sing-vmess v0.2.2 - V2Ray 协议实现（包含 vless 支持）
- 其他相关依赖：quic-go, utls, wireguard-go 等

**代码实现现状：**
- ✅ vless 协议已在支持列表中（speedtester.go:262行）
- ✅ Stash 兼容性检查已实现（speedtester.go:375-382行）
- ✅ 基础的协议解析和代理创建功能完整
- ⚠️ 缺乏详细的错误诊断和日志记录

### 1.2 问题识别

**观察到的现象：**
- 所有 vless 节点测试显示为"失败"状态
- 从测试日志看到部分节点延迟测试失败（packet_loss: 100）
- 缺乏具体的失败原因说明

**可能的根本原因：**
1. **网络连接问题** - DNS 解析、TCP 连接建立失败
2. **协议配置问题** - UUID 格式、flow 参数、传输协议配置
3. **mihomo 实现细节** - 特定的 vless 实现要求
4. **测试逻辑问题** - 超时设置、错误处理机制

## 2. 需求分析

### 2.1 核心需求

**REQ-001: 增强错误诊断能力**
- 优先级：P0（必须）
- 描述：为 vless 节点测试失败提供详细的错误信息
- 验收标准：
  - 能够区分连接建立、协议握手、数据传输等不同阶段的失败
  - 提供具体的错误码和错误描述
  - 在 WebSocket 实时更新中包含错误详情

**REQ-002: 优化 vless 测试流程**
- 优先级：P0（必须）
- 描述：改进 vless 协议的测试逻辑和参数设置
- 验收标准：
  - 增加 vless 特定的超时和重试机制
  - 优化连接建立和数据传输参数
  - 支持不同的传输协议（TCP、WebSocket、HTTP/2、gRPC）

**REQ-003: 增强配置验证**
- 优先级：P1（重要）
- 描述：在测试前验证 vless 节点配置的有效性
- 验收标准：
  - UUID 格式验证
  - 服务器地址和端口有效性检查
  - 传输协议参数验证

**REQ-004: 改进日志记录**
- 优先级：P1（重要）
- 描述：增加详细的调试日志以便问题排查
- 验收标准：
  - 记录 vless 连接建立过程
  - 记录协议握手详情
  - 记录网络传输统计信息

## 3. 技术方案

### 3.1 错误诊断增强方案

**实现方式：**

1. **分阶段错误捕获**
   ```go
   type VlessTestError struct {
       Stage   string `json:"stage"`   // "dns", "connect", "handshake", "transfer"
       Code    string `json:"code"`    // 错误代码
       Message string `json:"message"` // 详细错误信息
   }
   ```

2. **详细错误分类**
   - DNS_RESOLUTION_FAILED - DNS 解析失败
   - CONNECTION_REFUSED - 连接被拒绝
   - HANDSHAKE_TIMEOUT - 握手超时
   - PROTOCOL_ERROR - 协议错误
   - AUTHENTICATION_FAILED - 认证失败

3. **WebSocket 错误推送**
   - 在现有 WebSocket 消息中增加错误详情字段
   - 实时推送测试过程中的错误信息

### 3.2 测试流程优化方案

**具体改进：**

1. **超时参数调优**
   ```go
   type VlessTestConfig struct {
       ConnectTimeout   time.Duration // 连接超时（默认10s）
       HandshakeTimeout time.Duration // 握手超时（默认15s）
       TransferTimeout  time.Duration // 传输超时（默认30s）
   }
   ```

2. **重试机制**
   - 连接失败重试：最多3次
   - 区分临时错误和永久错误
   - 指数退避重试间隔

3. **传输协议优化**
   - TCP：直接连接
   - WebSocket：正确处理 ws-path 和 headers
   - HTTP/2：支持 h2 多路复用
   - gRPC：正确配置 grpc-opts

### 3.3 配置验证方案

**验证规则：**

1. **UUID 验证**
   ```go
   func validateVlessUUID(uuid string) error {
       if _, err := uuid.Parse(uuid); err != nil {
           return fmt.Errorf("invalid UUID format: %v", err)
       }
       return nil
   }
   ```

2. **网络配置验证**
   - 服务器地址格式检查
   - 端口范围验证（1-65535）
   - TLS 配置一致性检查

3. **协议参数验证**
   - flow 参数有效性检查
   - 传输协议参数完整性验证

### 3.4 日志增强方案

**日志级别设计：**

1. **DEBUG 级别**
   - 连接建立过程
   - 协议握手详情
   - 数据包传输统计

2. **INFO 级别**
   - 测试开始/结束
   - 成功的测试结果
   - 重要的配置信息

3. **ERROR 级别**
   - 所有失败情况
   - 异常错误详情

## 4. 实施计划

### 4.1 开发阶段划分

**阶段一：诊断增强（1-2天）**
- 实现详细的错误分类和捕获
- 增强 WebSocket 错误推送
- 添加调试日志输出

**阶段二：测试优化（1-2天）**
- 优化超时和重试参数
- 改进连接建立逻辑
- 优化传输协议处理

**阶段三：验证和完善（1天）**
- 实现配置验证
- 完善日志记录
- 端到端测试验证

### 4.2 测试验证计划

**测试用例设计：**

1. **正常场景测试**
   - 有效的 vless 节点配置
   - 不同传输协议的节点
   - 各种 flow 参数组合

2. **异常场景测试**
   - 无效的 UUID 格式
   - 不可达的服务器地址
   - 错误的端口配置
   - 超时场景模拟

3. **性能测试**
   - 大量 vless 节点并发测试
   - 长时间稳定性测试
   - 内存和 CPU 使用率监控

### 4.3 成功标准

**定量指标：**
- vless 节点测试成功率提升至 >80%（对于有效节点）
- 错误诊断覆盖率达到 100%
- 测试响应时间 <30秒（单个节点）

**定性指标：**
- 用户能够快速识别失败原因
- 开发团队能够快速定位问题
- 系统稳定性显著提升

## 5. 风险评估

### 5.1 技术风险

**风险点：**
- mihomo 内核版本兼容性问题
- 第三方依赖库的 bug 或限制
- 网络环境差异导致的测试不稳定

**缓解措施：**
- 详细的兼容性测试
- 多环境验证
- 渐进式部署策略

### 5.2 业务风险

**风险点：**
- 优化过程中可能影响其他协议的测试
- 性能优化可能引入新的稳定性问题

**缓解措施：**
- 充分的回归测试
- 特性开关控制
- 监控和回滚机制

## 6. 依赖和约束

### 6.1 技术依赖

- mihomo v1.19.10+ 的稳定性和功能完整性
- Go 1.24+ 的并发和网络特性
- WebSocket 连接的稳定性

### 6.2 外部约束

- 代理服务器的质量和稳定性
- 网络环境的复杂性
- 不同地区的网络政策限制

## 7. 后续优化方向

1. **智能重试策略** - 基于错误类型的差异化重试
2. **性能优化** - 连接池复用、并发控制优化
3. **用户体验** - 更友好的错误提示和修复建议
4. **监控告警** - 实时的系统健康状态监控

---

**文档版本：** 1.0  
**创建日期：** 2025-07-07  
**负责人：** Claude Code  
**审核状态：** 待审核  