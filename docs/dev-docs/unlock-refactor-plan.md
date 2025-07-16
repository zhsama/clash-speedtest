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
│   ├── detector.go              # 解锁检测主逻辑
│   ├── types.go                 # 数据类型定义
│   ├── config.go                # 检测配置管理
│   ├── results.go               # 结果处理和格式化
│   ├── utils.go                 # 工具函数（已存在）
│   ├── services/                # 各平台检测实现
│   │   ├── netflix.go           # Netflix 检测（已存在）
│   │   ├── disney.go            # Disney+ 检测
│   │   ├── youtube.go           # YouTube Premium 检测
│   │   ├── openai.go            # ChatGPT/OpenAI 检测
│   │   ├── spotify.go           # Spotify 检测
│   │   ├── hbo.go               # HBO Max 检测
│   │   ├── hulu.go              # Hulu 检测
│   │   ├── prime_video.go       # Prime Video 检测
│   │   ├── bilibili.go          # Bilibili 检测
│   │   ├── steam.go             # Steam 货币检测
│   │   └── ...                  # 其他平台
│   └── cache/                   # 缓存模块
│       ├── cache.go             # 缓存实现
│       └── types.go             # 缓存数据类型
├── speedtester/
│   └── speedtester.go           # 集成解锁检测到测速流程
└── main.go                      # API 端点更新
```

#### 2.2 核心数据结构

```go
// UnlockResult 单个平台的解锁检测结果
type UnlockResult struct {
    Platform    string    `json:"platform"`      // 平台名称
    Status      UnlockStatus `json:"status"`     // 状态: unlocked/locked/failed/error
    Region      string    `json:"region"`        // 解锁地区
    Message     string    `json:"message"`       // 额外信息
    Latency     int64     `json:"latency_ms"`    // 检测延迟
    CheckedAt   time.Time `json:"checked_at"`    // 检测时间
}

// UnlockStatus 解锁状态枚举
type UnlockStatus string

const (
    StatusUnlocked UnlockStatus = "unlocked"
    StatusLocked   UnlockStatus = "locked"
    StatusFailed   UnlockStatus = "failed"
    StatusError    UnlockStatus = "error"
)

// UnlockDetector 解锁检测器接口
type UnlockDetector interface {
    Detect(proxy constant.Proxy, timeout time.Duration) *UnlockResult
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

#### 2.3 基于模板项目的检测器实现

参考模板项目的检测器设计，实现统一的检测器接口：

```go
// BaseDetector 基础检测器实现（已存在于utils.go）
type BaseDetector struct {
    platformName string
    priority     int
}

// 具体平台检测器实现示例
type NetflixDetector struct {
    *BaseDetector
}

func NewNetflixDetector() *NetflixDetector {
    return &NetflixDetector{
        BaseDetector: NewBaseDetector("Netflix", 1),
    }
}

func (d *NetflixDetector) Detect(proxy constant.Proxy, timeout time.Duration) *UnlockResult {
    d.logDetectionStart(proxy)
    
    client := createHTTPClient(proxy, timeout)
    
    // 访问Netflix进行检测
    resp, err := makeRequest(client, "GET", "https://www.netflix.com/title/81280792", nil)
    if err != nil {
        result := d.createErrorResult("Failed to connect to Netflix", err)
        d.logDetectionResult(proxy, result)
        return result
    }
    defer resp.Body.Close()
    
    // 解析响应并判断状态
    bodyStr, err := io.ReadAll(resp.Body)
    if err != nil {
        result := d.createErrorResult("Failed to read Netflix response", err)
        d.logDetectionResult(proxy, result)
        return result
    }
    
    // 根据响应内容判断解锁状态
    var result *UnlockResult
    if strings.Contains(string(bodyStr), "Not Available") {
        result = d.createResult(StatusLocked, "", "Netflix content not available")
    } else if strings.Contains(string(bodyStr), "requestCountry") {
        region := d.extractCountryCode(string(bodyStr))
        result = d.createResult(StatusUnlocked, region, "Netflix accessible")
    } else {
        result = d.createResult(StatusFailed, "", "Unable to determine Netflix status")
    }
    
    d.logDetectionResult(proxy, result)
    return result
}
```

#### 2.4 IP风险检测集成

```go
// IPRiskDetector IP风险检测器
type IPRiskDetector struct {
    client *http.Client
    cache  *sync.Map
}

type IPRiskInfo struct {
    IP        string  `json:"ip"`
    Country   string  `json:"country"`
    City      string  `json:"city"`
    ISP       string  `json:"isp"`
    RiskScore int     `json:"risk_score"`  // 风险评分 0-100
    Type      string  `json:"type"`        // 代理类型检测
}

func (d *IPRiskDetector) DetectIPRisk(proxy constant.Proxy) *IPRiskInfo {
    // 实现IP风险检测逻辑
    // 使用ip-api.com或类似服务
    return &IPRiskInfo{
        IP:        "x.x.x.x",
        Country:   "US",
        City:      "New York",
        ISP:       "Example ISP",
        RiskScore: 25,
        Type:      "datacenter",
    }
}
```

### 3. 主检测器实现

```go
// Detector 主检测器
type Detector struct {
    config     *UnlockTestConfig
    detectors  map[string]UnlockDetector
    cache      *UnlockCache
    ipDetector *IPRiskDetector
    semaphore  chan struct{}
}

func NewDetector(config *UnlockTestConfig) *Detector {
    detector := &Detector{
        config:     config,
        detectors:  make(map[string]UnlockDetector),
        cache:      NewUnlockCache(),
        ipDetector: NewIPRiskDetector(),
        semaphore:  make(chan struct{}, config.Concurrent),
    }
    
    // 注册所有检测器
    detector.registerDefaultDetectors()
    
    return detector
}

func (d *Detector) registerDefaultDetectors() {
    // 注册现有的6个平台
    d.Register("Netflix", NewNetflixDetector())
    d.Register("YouTube", NewYouTubeDetector())
    d.Register("Disney+", NewDisneyDetector())
    d.Register("ChatGPT", NewChatGPTDetector())
    d.Register("Spotify", NewSpotifyDetector())
    d.Register("Bilibili", NewBilibiliDetector())
    
    // 注册新增的34个平台
    d.Register("HBO Max", NewHBOMaxDetector())
    d.Register("Hulu", NewHuluDetector())
    d.Register("Prime Video", NewPrimeVideoDetector())
    d.Register("Apple TV+", NewAppleTVDetector())
    // ... 继续注册其他平台
}

func (d *Detector) DetectAll(proxy constant.Proxy, platforms []string) []UnlockResult {
    results := make([]UnlockResult, 0, len(platforms))
    resultsChan := make(chan UnlockResult, len(platforms))
    
    var wg sync.WaitGroup
    
    // 并发检测
    for _, platform := range platforms {
        wg.Add(1)
        go func(p string) {
            defer wg.Done()
            
            // 获取信号量
            d.semaphore <- struct{}{}
            defer func() { <-d.semaphore }()
            
            // 检查缓存
            if d.config.EnableCache {
                if cached := d.cache.Get(proxy.Name(), p); cached != nil {
                    resultsChan <- *cached
                    return
                }
            }
            
            // 执行检测
            detector := d.detectors[p]
            result := detector.Detect(proxy, time.Duration(d.config.Timeout)*time.Second)
            
            // 缓存结果
            if d.config.EnableCache {
                d.cache.Set(proxy.Name(), p, result, time.Duration(d.config.CacheTTL)*time.Minute)
            }
            
            resultsChan <- *result
        }(platform)
    }
    
    wg.Wait()
    close(resultsChan)
    
    // 收集结果
    for result := range resultsChan {
        results = append(results, result)
    }
    
    // 按优先级排序
    sort.Slice(results, func(i, j int) bool {
        return d.detectors[results[i].Platform].GetPriority() < d.detectors[results[j].Platform].GetPriority()
    })
    
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

func NewUnlockCache() *UnlockCache {
    return &UnlockCache{}
}

func (c *UnlockCache) Set(proxyName, platform string, result *UnlockResult, ttl time.Duration) {
    key := fmt.Sprintf("%s:%s", proxyName, platform)
    c.cache.Store(key, &CachedResult{
        Result:    result,
        ExpiresAt: time.Now().Add(ttl),
    })
}

func (c *UnlockCache) Get(proxyName, platform string) *UnlockResult {
    key := fmt.Sprintf("%s:%s", proxyName, platform)
    value, ok := c.cache.Load(key)
    if !ok {
        return nil
    }
    
    cached := value.(*CachedResult)
    if time.Now().After(cached.ExpiresAt) {
        c.cache.Delete(key)
        return nil
    }
    
    return cached.Result
}
```

## 第二部分：测试策略

### 1. 测试架构设计

#### 1.1 分层测试架构

```
┌─────────────────────────────────────────────────────────────┐
│                    E2E 测试层                                │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐│
│  │   前端界面测试   │  │   API集成测试   │  │   用户场景测试   ││
│  └─────────────────┘  └─────────────────┘  └─────────────────┘│
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   集成测试层                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐│
│  │   系统集成测试   │  │   服务集成测试   │  │   数据流测试     ││
│  └─────────────────┘  └─────────────────┘  └─────────────────┘│
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   单元测试层                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐│
│  │   检测器单元测试 │  │   工具函数测试   │  │   缓存机制测试   ││
│  └─────────────────┘  └─────────────────┘  └─────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

#### 1.2 测试工具选择

**Go后端测试工具**：
- **标准库testing**：基础单元测试
- **testify/suite**：测试套件组织
- **gomock**：接口Mock生成
- **httptest**：HTTP服务Mock
- **gock**：HTTP请求拦截和Mock

**前端测试工具**：
- **Vitest**：React组件单元测试
- **React Testing Library**：React组件测试
- **Playwright**：浏览器自动化测试

### 2. 测试用例设计

#### 2.1 平台检测器测试模板

```go
// 每个平台检测器需要实现的标准测试用例
func TestPlatformDetector(t *testing.T) {
    detector := NewPlatformDetector()
    
    testCases := []struct {
        name           string
        mockResponse   MockResponse
        expectedStatus UnlockStatus
        expectedRegion string
        expectedError  bool
    }{
        {
            name: "成功解锁_美国地区",
            mockResponse: MockResponse{
                Status: 200,
                Body:   getSuccessResponse("US"),
            },
            expectedStatus: StatusUnlocked,
            expectedRegion: "US",
            expectedError:  false,
        },
        {
            name: "地区封锁",
            mockResponse: MockResponse{
                Status: 200,
                Body:   getBlockedResponse(),
            },
            expectedStatus: StatusLocked,
            expectedRegion: "",
            expectedError:  false,
        },
        {
            name: "网络超时",
            mockResponse: MockResponse{
                Status: 0,
                Body:   "",
                Error:  errors.New("timeout"),
            },
            expectedStatus: StatusError,
            expectedRegion: "",
            expectedError:  true,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // 设置Mock
            // 执行测试
            // 验证结果
        })
    }
}
```

#### 2.2 集成测试用例

```go
// 完整检测流程测试
func TestUnlockDetectionFlow(t *testing.T) {
    // 设置测试环境
    // 模拟代理连接
    // 执行完整检测流程
    // 验证结果格式和内容
}

// 多平台并发检测测试
func TestMultiPlatformConcurrentDetection(t *testing.T) {
    // 设置多个平台检测器
    // 模拟并发检测场景
    // 验证结果一致性
    // 验证性能指标
}
```

### 3. 性能测试

#### 3.1 并发性能测试

```go
func TestConcurrentDetectionPerformance(t *testing.T) {
    // 测试不同并发级别的性能
    // 测量检测延迟
    // 测量内存使用
    // 测量CPU使用率
}
```

#### 3.2 性能要求基准

```go
const (
    MaxDetectionTime = 30 * time.Second  // 最大检测时间
    MaxMemoryUsage   = 100 * 1024 * 1024 // 最大内存使用：100MB
    MaxCPUUsage      = 80                // 最大CPU使用率：80%
)
```

### 4. CI/CD配置

#### 4.1 GitHub Actions工作流

```yaml
name: Comprehensive Test Suite

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21, 1.22]
        test-type: [unit, integration, performance]
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Run unit tests
      if: matrix.test-type == 'unit'
      run: |
        cd backend
        go test -v -race -coverprofile=coverage.out ./...
    
    - name: Run integration tests
      if: matrix.test-type == 'integration'
      run: |
        cd backend
        go test -v -tags=integration ./...
    
    - name: Run performance tests
      if: matrix.test-type == 'performance'
      run: |
        cd backend
        go test -v -bench=. -benchmem ./...
```

## 第三部分：实施计划

### 1. 重构阶段规划

#### 1.1 第一阶段：基础重构（2周）

**目标**：建立新的检测器架构，迁移现有平台

**任务清单**：
- [ ] 创建统一的UnlockDetector接口
- [ ] 重构现有的6个平台检测器
- [ ] 实现主检测器(Detector)
- [ ] 建立缓存机制
- [ ] 创建基础测试框架

**交付物**：
- 重构后的检测器架构
- 现有6个平台的新实现
- 基础测试套件
- 缓存机制实现

#### 1.2 第二阶段：平台扩展（3周）

**目标**：实现34个新平台的检测器

**任务清单**：
- [ ] 实现流媒体平台检测器（Netflix系、Disney系、HBO系等）
- [ ] 实现音乐平台检测器（Spotify、Apple Music、Pandora等）
- [ ] 实现游戏平台检测器（Steam、Epic Games、PlayStation等）
- [ ] 实现AI平台检测器（ChatGPT、Claude、Midjourney等）
- [ ] 实现区域特定平台检测器（Bilibili、iQiyi、Tencent Video等）

**交付物**：
- 40+个平台检测器
- 完整的测试用例
- 性能基准测试
- 文档和示例

#### 1.3 第三阶段：IP风险检测集成（1周）

**目标**：集成IP风险检测功能

**任务清单**：
- [ ] 实现IP风险检测器
- [ ] 集成到主检测流程
- [ ] 添加风险评分逻辑
- [ ] 实现风险检测缓存

**交付物**：
- IP风险检测功能
- 风险评分算法
- 集成测试用例
- 性能优化

#### 1.4 第四阶段：测试和优化（2周）

**目标**：完善测试体系，优化性能

**任务清单**：
- [ ] 完善单元测试（覆盖率>85%）
- [ ] 实现集成测试
- [ ] 性能测试和优化
- [ ] 错误处理完善
- [ ] 文档完善

**交付物**：
- 完整的测试套件
- 性能优化报告
- 错误处理机制
- 完整的项目文档

#### 1.5 第五阶段：前端集成和部署（1周）

**目标**：前端集成和生产部署

**任务清单**：
- [ ] 前端UI更新
- [ ] WebSocket消息集成
- [ ] 实时进度展示
- [ ] 部署和监控

**交付物**：
- 更新的前端界面
- 实时检测展示
- 生产环境部署
- 监控和日志系统

### 2. 质量保障措施

#### 2.1 代码质量标准

- **测试覆盖率**：单元测试 >= 85%，集成测试 >= 70%
- **性能要求**：检测时间 < 30秒，内存使用 < 100MB
- **错误处理**：完善的错误处理和重试机制
- **代码审查**：所有代码都需要经过代码审查

#### 2.2 持续集成

- **每次提交**：单元测试、代码质量检查
- **每日构建**：完整测试套件、性能测试
- **每周构建**：压力测试、兼容性测试

### 3. 风险和挑战

#### 3.1 技术风险

1. **平台反检测**：部分平台可能检测并阻止自动化访问
2. **API变更**：流媒体平台可能更改检测端点
3. **性能影响**：大量并发检测可能影响整体性能
4. **准确性问题**：某些平台检测结果可能不准确

#### 3.2 缓解策略

1. **请求伪装**：使用真实浏览器头部，随机化请求
2. **监控机制**：实时监控检测成功率，及时调整
3. **性能优化**：智能缓存、连接池、并发控制
4. **持续验证**：定期验证检测准确性

## 第四部分：监控和维护

### 1. 监控指标

#### 1.1 核心指标

- **检测成功率**：各平台检测成功率
- **检测延迟**：平均检测时间
- **缓存命中率**：缓存效率
- **错误率**：各类错误的发生率

#### 1.2 业务指标

- **平台可用性**：各平台的可用性统计
- **用户使用情况**：用户对不同平台的使用偏好
- **性能表现**：系统整体性能表现

### 2. 运维策略

#### 2.1 自动化监控

- **健康检查**：定期检查各平台检测器状态
- **性能监控**：实时监控系统性能指标
- **错误告警**：异常情况自动告警

#### 2.2 维护策略

- **定期更新**：定期更新检测逻辑和端点
- **版本控制**：严格的版本管理和回滚机制
- **文档维护**：持续更新技术文档

## 总结

本综合方案通过整合技术架构设计和测试策略，为clash-speedtest项目的解锁检测功能重构提供了完整的指导。该方案确保了：

1. **技术可行性**：基于现有架构，渐进式重构
2. **质量保障**：完整的测试体系和质量标准
3. **可扩展性**：插件化架构支持未来扩展
4. **维护性**：清晰的代码结构和文档

通过按照本方案实施，项目能够顺利完成从6个平台到40+个平台的扩展，同时保持代码质量和系统性能。