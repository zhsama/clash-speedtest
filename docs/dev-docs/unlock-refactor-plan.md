# 解锁检测功能重构综合方案

## 文档概述

本文档整合了解锁检测功能的技术架构设计和测试策略，为clash-speedtest项目从6个平台扩展到40+个平台的重构提供完整的实施指导。

## 第一部分：技术架构设计

### 1. 功能概述

#### 1.1 核心功能
在现有的代理测速功能基础上，新增流媒体解锁检测能力，支持检测节点对 40+ 流媒体平台的访问能力。

#### 1.2 测试模式
- **仅测速** (Speed Only)：只进行延迟、下载、上传速度测试
- **仅解锁** (Unlock Only)：只进行流媒体解锁检测
- **综合测试** (Both)：同时进行测速和解锁检测（默认）

#### 1.3 目标收益
- 用户可以快速了解节点的流媒体解锁能力
- 帮助用户选择适合观看特定流媒体平台的节点
- 提供详细的地区信息，指导用户选择最佳节点

### 2. 技术架构设计

#### 2.1 模块结构

```
backend/
├── unlock/                       # 流媒体解锁检测模块
│   ├── detector.go              # 解锁检测主逻辑与接口定义
│   ├── registration.go          # (新增) 自动化注册机制
│   ├── types.go                 # 数据类型定义
│   ├── config.go                # 检测配置管理
│   ├── cache/                   # 缓存模块
│   │   ├── cache.go             # 缓存实现
│   │   └── types.go             # 缓存数据类型
│   └── services/                # 各平台检测实现 (插件)
│       ├── netflix.go           # Netflix 检测实现
│       ├── disney.go            # Disney+ 检测实现
│       └── ...                  # 其他平台
├── speedtester/
│   └── speedtester.go           # 集成解锁检测到测速流程
└── main.go                      # API 端点更新
```

#### 2.2 核心数据结构 (优化)

```go
// UnlockResult 单个平台的解锁检测结果
type UnlockResult struct {
    Platform    string       `json:"platform"`      // 平台名称
    Status      UnlockStatus `json:"status"`        // 状态: unlocked/locked/failed/error
    Region      string       `json:"region"`        // 解锁地区
    Message     string       `json:"message"`       // (优化) 面向用户的友好信息
    Latency     int64        `json:"latency_ms"`    // 检测延迟
    CheckedAt   time.Time    `json:"checked_at"`    // 检测时间
    err         error        `json:"-"`             // (新增) 内部错误，不序列化，用于日志和调试
}

// UnlockStatus 解锁状态枚举
type UnlockStatus string

const (
    StatusUnlocked UnlockStatus = "unlocked"
    StatusLocked   UnlockStatus = "locked"
    StatusFailed   UnlockStatus = "failed" // 检测逻辑执行成功，但无法判断解锁状态
    StatusError    UnlockStatus = "error"   // 检测逻辑执行失败（如网络错误）
)

// UnlockDetector 解锁检测器接口 (优化)
type UnlockDetector interface {
    // (优化) 使用 context.Context 控制超时、取消和传递元数据
    Detect(ctx context.Context, proxy constant.Proxy) *UnlockResult
    GetPlatformName() string
    GetPriority() int  // 检测优先级
}

// UnlockTestConfig 解锁检测配置
type UnlockTestConfig struct {
    Enabled          bool     `json:"enabled"`           // 是否启用
    Platforms        []string `json:"platforms"`         // 要检测的平台列表
    Concurrent       int      `json:"concurrent"`        // 并发检测数
    Timeout          int      `json:"timeout"`           // 单个检测超时（秒）
    RetryOnError     bool     `json:"retry_on_error"`    // 错误时重试
    IncludeIPInfo    bool     `json:"include_ip_info"`   // 包含 IP 信息
    EnableCache      bool     `json:"enable_cache"`      // 启用缓存
    CacheTTL         int      `json:"cache_ttl"`         // 缓存时间（分钟）
}
```

#### 2.3 检测器自动化注册机制 (新增)

**目标**: 实现检测器的自动注册，避免在主检测器中硬编码，提升模块化和可维护性。

**实现方式**: 利用 Go 的 `init()` 函数特性。

```go
// unlock/registration.go
package unlock

import "sync"

var (
    detectors = make(map[string]UnlockDetector)
    mu        sync.RWMutex
)

// Register 用于在每个检测器实现文件中调用，完成自我注册
func Register(detector UnlockDetector) {
    mu.Lock()
    defer mu.Unlock()
    if detector == nil {
        panic("cannot register a nil detector")
    }
    name := detector.GetPlatformName()
    if _, ok := detectors[name]; ok {
        panic("detector already registered: " + name)
    }
    detectors[name] = detector
}

// GetDetectors 返回所有已注册检测器的副本
func GetDetectors() map[string]UnlockDetector {
    mu.RLock()
    defer mu.RUnlock()
    // 返回副本以保证并发安全
    d := make(map[string]UnlockDetector, len(detectors))
    for name, detector := range detectors {
        d[name] = detector
    }
    return d
}
```

**具体平台检测器实现示例**:

```go
// unlock/services/netflix.go
package services

import (
    "..."
    "path/to/unlock" // 导入 unlock 包
)

// 在文件级别执行，自动将检测器注册到全局映射中
func init() {
    unlock.Register(NewNetflixDetector())
}

type NetflixDetector struct { ... }

func NewNetflixDetector() unlock.UnlockDetector { ... }

func (d *NetflixDetector) Detect(ctx context.Context, proxy constant.Proxy) *unlock.UnlockResult {
    // (优化) 使用 ctx 控制请求超时
    client := createHTTPClient(proxy) // 超时应在 http.Request 级别设置

    req, err := http.NewRequestWithContext(ctx, "GET", "https://www.netflix.com/...", nil)
    // ...
}
```

#### 2.4 IP风险检测集成 (优化)

**目标**: 将 IP 风险检测抽象为接口，支持未来更换或聚合服务。

```go
// IPRiskInfo IP风险信息
type IPRiskInfo struct {
    IP        string  `json:"ip"`
    Country   string  `json:"country"`
    City      string  `json:"city"`
    ISP       string  `json:"isp"`
    RiskScore int     `json:"risk_score"`  // 风险评分 0-100
    Type      string  `json:"type"`        // 代理类型检测
}

// IPRiskDetector IP风险检测器接口
type IPRiskDetector interface {
    DetectIPRisk(ctx context.Context, proxy constant.Proxy) (*IPRiskInfo, error)
}

// ipApiComDetector ip-api.com 服务的实现
type ipApiComDetector struct { ... }

func (d *ipApiComDetector) DetectIPRisk(ctx context.Context, proxy constant.Proxy) (*IPRiskInfo, error) {
    // 实现具体的检测逻辑
}
```

### 3. 主检测器实现 (优化)

```go
// Detector 主检测器
type Detector struct {
    config     *UnlockTestConfig
    detectors  map[string]UnlockDetector // 从自动化注册机制中获取
    cache      *UnlockCache
    ipDetector IPRiskDetector // (优化) 使用接口类型
    semaphore  chan struct{}
}

func NewDetector(config *UnlockTestConfig) *Detector {
    detector := &Detector{
        config:     config,
        detectors:  GetDetectors(), // (优化) 从全局注册器获取
        cache:      NewUnlockCache(),
        ipDetector: NewIpApiComDetector(), // 注入具体的实现
        semaphore:  make(chan struct{}, config.Concurrent),
    }
    return detector
}

// (移除) registerDefaultDetectors 方法，因其已被自动化注册机制取代

func (d *Detector) DetectAll(ctx context.Context, proxy constant.Proxy, platforms []string) []UnlockResult {
    results := make([]UnlockResult, 0, len(platforms))
    resultsChan := make(chan UnlockResult, len(platforms))
    
    var wg sync.WaitGroup
    
    // (优化) 创建带总超时的 context
    overallCtx, cancel := context.WithTimeout(ctx, time.Duration(d.config.Timeout)*time.Second*time.Duration(len(platforms)))
    defer cancel()

    for _, platform := range platforms {
        wg.Add(1)
        go func(p string) {
            defer wg.Done()
            
            d.semaphore <- struct{}{}
            defer func() { <-d.semaphore }()
            
            // (优化) 为每个检测创建独立的、带超时和取消能力的 context
            detectCtx, detectCancel := context.WithTimeout(overallCtx, time.Duration(d.config.Timeout)*time.Second)
            defer detectCancel()

            // ... 检查缓存 ...
            
            detector, ok := d.detectors[p]
            if !ok {
                // 处理检测器不存在的情况
                return
            }
            
            // (优化) 传递 context
            result := detector.Detect(detectCtx, proxy)
            
            // ... 缓存结果 ...
            
            resultsChan <- *result
        }(platform)
    }
    
    wg.Wait()
    close(resultsChan)
    
    // ... 收集和排序结果 ...
    
    return results
}
```

### 4. 缓存机制实现

```go
// UnlockCache 解锁结果缓存
type UnlockCache struct {
    cache sync.Map
}

type CachedResult struct {
    Result    *UnlockResult
    ExpiresAt time.Time
}

// ... 实现保持不变 ...
```

## 第二部分：测试策略

### 1. 测试架构设计
... (保持不变) ...

### 2. 测试用例设计

#### 2.1 平台检测器测试模板

```go
// (优化) 测试用例需适配新的 Detect 接口签名
func TestPlatformDetector(t *testing.T) {
    detector := NewPlatformDetector()
    
    // ...
    
    t.Run(tc.name, func(t *testing.T) {
        // (优化) 创建带超时的 context 用于测试
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        // 设置Mock
        // 执行测试: detector.Detect(ctx, mockProxy)
        // 验证结果
    })
}
```
... (其余部分保持不变) ...

## 第三部分：实施计划

### 1. 重构阶段规划

#### 1.1 第一阶段：基础重构（2周）

**目标**：建立新的检测器架构，迁移现有平台

**任务清单**：
- [ ] **(优化)** 定义新的 `UnlockDetector` 接口，强制使用 `context.Context`。
- [ ] **(优化)** 实现检测器的自动化注册机制 (`unlock/registration.go`)。
- [ ] 重构现有的6个平台检测器，使其符合新接口并实现自动化注册。
- [ ] 实现主检测器(Detector)，使其从注册器动态加载检测器。
- [ ] 建立缓存机制。
- [ ] 创建基础测试框架，并更新测试用例以适应新接口。

**交付物**：
- 采用 `context` 并支持自动化注册的检测器架构。
- 现有6个平台的新实现。
- 基础测试套件。
- 缓存机制实现。

... (其余阶段保持不变，但都将受益于第一阶段的坚实基础) ...

## 总结

本综合方案通过整合技术架构设计和测试策略，为clash-speedtest项目的解锁检测功能重构提供了完整的指导。**经过优化，该方案进一步提升了系统的健壮性、模块化和可维护性**，确保了：

1.  **技术可行性与健壮性**：基于 `context.Context` 的现代化 Go 编程实践。
2.  **质量保障**：完整的测试体系和质量标准。
3.  **可扩展性**：**自动化注册的插件化架构**，支持未来轻松扩展。
4.  **维护性**：清晰、解耦的代码结构和文档。

通过按照本方案实施，项目能够顺利完成从6个平台到40+个平台的扩展，同时保持高质量的代码和卓越的系统性能。