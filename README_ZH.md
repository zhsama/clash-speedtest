# Clash SpeedTest Pro

> 专业的代理节点性能测试工具 - Professional proxy speed testing tool

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.19-blue)](https://golang.org/)
[![Node.js Version](https://img.shields.io/badge/Node.js-%3E%3D18.0-green)](https://nodejs.org/)
[![License](https://img.shields.io/badge/license-GPL--3.0-green)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/zhsama/clash-speedtest)

**[English Documentation](README.md) | 中文文档**

基于 Clash/Mihomo 核心的专业测速工具，提供命令行和现代化 Web 界面，支持实时进度显示和流媒体解锁检测。

<img width="1332" alt="Clash SpeedTest Pro Web Interface" src="https://github.com/user-attachments/assets/fdc47ec5-b626-45a3-a38a-6d88c326c588">

## 🚀 核心特性

### 🎯 测试功能

- **直接测试**: 无需额外配置，直接读取 Clash/Mihomo 配置文件或订阅链接
- **高性能**: 支持并发测试，快速获取节点性能数据
- **双模式**: 同时支持速度测试和流媒体解锁检测
- **智能过滤**: 支持多种过滤条件（速度、延迟、协议类型、节点名称等）

### 🌐 用户界面

- **现代化界面**: React/TypeScript 构建的现代化 Web 界面
- **实时进度**: WebSocket 实时显示测试进度和结果
- **响应式设计**: 完美适配桌面和移动设备
- **导出功能**: 支持 Markdown 和 CSV 格式导出测试结果

### 🔓 解锁检测

支持 30+ 流媒体平台检测，包括：

- Netflix、YouTube、Disney+、ChatGPT
- Spotify、Bilibili、HBO Max、Hulu
- Amazon Prime Video、Paramount+、Peacock
- 及更多国际和地区性平台

### 🛡️ 安全可靠

- **开源代码**: 完全开源，可审计的代码
- **本地运行**: 保护节点隐私，数据不上传
- **跨平台**: 支持 Windows、macOS、Linux

## 📦 安装方法

### 方法一：一键开发环境 (推荐)

```bash
# 克隆仓库
git clone https://github.com/zhsama/clash-speedtest.git
cd clash-speedtest

# 安装依赖并启动完整开发环境
pnpm install
pnpm dev

# 访问 Web 界面: http://localhost:3000
# 后端 API: http://localhost:8080
```

### 方法二：Go Install (命令行版本)

```bash
go install github.com/zhsama/clash-speedtest@latest
```

### 方法三：预编译二进制文件

从 [Releases](https://github.com/zhsama/clash-speedtest/releases) 页面下载对应平台的二进制文件。

### 方法四：Docker 部署

```bash
# 构建并启动服务
docker-compose up -d

# 或使用优化版本
docker-compose -f docker-compose.optimized.yml up -d
```

## 🎯 使用方法

### Web 界面使用 (推荐)

#### 1. 启动服务

```bash
# 完整环境启动
pnpm dev

# 或分别启动
pnpm dev:backend  # 启动后端 API 服务
pnpm dev:frontend # 启动前端界面
```

#### 2. 使用 Web 界面

1. 打开浏览器访问 `http://localhost:3000`
2. 在"配置获取"部分输入配置文件路径或订阅链接
3. 点击"获取配置"加载节点列表
4. 配置测试参数：
   - **测试模式**: 全面测试（测速+解锁）/ 仅测速 / 仅解锁检测
   - **节点过滤**: 包含/排除特定节点，协议类型过滤
   - **速度过滤**: 设置最低速度和最大延迟阈值
   - **高级配置**: 并发数、超时时间、测试包大小等
5. 点击"开始测试"开始测试，实时查看进度和结果
6. 测试完成后可导出 Markdown 或 CSV 格式的测试报告

#### 3. Web 界面特性

- **实时进度**: 通过 WebSocket 实时显示测试进度
- **节点预览**: 测试前预览符合条件的节点列表
- **智能过滤**: 支持中英文逗号分隔的节点过滤
- **TUN 模式检测**: 自动检测并提醒 TUN 模式状态
- **结果导出**: 智能文件命名，包含配置来源信息

### 命令行使用

```bash
# 查看帮助
clash-speedtest -h

# 演示：

# 1. 测试全部节点，使用 HTTP 订阅地址
# 请在订阅地址后面带上 flag=meta 参数，否则无法识别出节点类型
clash-speedtest -c 'https://domain.com/api/v1/client/subscribe?token=secret&flag=meta'

# 2. 测试香港节点，使用正则表达式过滤，使用本地文件
clash-speedtest -c ~/.config/clash/config.yaml -f 'HK|港'

# 3. 混合使用多个配置源
clash-speedtest -c "https://domain.com/api/v1/client/subscribe?token=secret&flag=meta,/home/.config/clash/config.yaml"

# 4. 筛选出延迟低于 800ms 且下载速度大于 5MB/s 的节点，并输出到 filtered.yaml
clash-speedtest -c "https://domain.com/api/v1/client/subscribe?token=secret&flag=meta" -output filtered.yaml -max-latency 800ms -min-speed 5

# 5. 使用 -rename 选项按照 IP 地区和下载速度重命名节点
clash-speedtest -c config.yaml -output result.yaml -rename
# 重命名后的节点名称格式：🇺🇸 US | ⬇️ 15.67 MB/s
```

## 🏗️ 项目架构

### 技术栈

- **后端**: Go + Gin + WebSocket + Mihomo Core
- **前端**: React + TypeScript + Astro + Tailwind CSS
- **构建工具**: Turborepo + Vite + pnpm
- **容器化**: Docker + Multi-stage builds
- **部署**: GitHub Actions + 自动化发布

### 目录结构

```
clash-speedtest/
├── backend/                 # Go 后端服务
│   ├── main.go             # 主程序入口
│   ├── server/             # HTTP/WebSocket 服务
│   ├── speedtester/        # 核心测速逻辑
│   ├── unlock/             # 流媒体解锁检测
│   ├── detectors/          # 各平台检测器
│   ├── websocket/          # WebSocket 实时通信
│   ├── tasks/              # 异步任务管理
│   ├── utils/              # 工具函数
│   └── download-server/    # 可选的自托管测速服务器
├── frontend/               # React/TypeScript 前端
│   ├── src/
│   │   ├── components/     # React 组件
│   │   │   ├── SpeedTest.tsx           # 主测试组件
│   │   │   ├── RealTimeProgressTable.tsx  # 实时进度表格
│   │   │   ├── SpeedTestTable.tsx      # 速度测试表格
│   │   │   ├── UnlockTestTable.tsx     # 解锁测试表格
│   │   │   └── TUNWarning.tsx          # TUN模式检测
│   │   ├── hooks/          # 自定义 Hooks
│   │   │   └── useWebSocket.ts         # WebSocket 管理
│   │   ├── lib/            # 工具库
│   │   └── styles/         # 样式文件
│   ├── public/             # 静态资源
│   └── package.json
├── docs/                   # 项目文档
│   ├── dev-docs/           # 开发文档
│   ├── test-docs/          # 测试文档
│   └── docker/             # Docker 文档
├── scripts/                # 构建脚本
├── turbo.json              # Turborepo 配置
├── package.json            # 根目录配置
└── README.md
```

### 核心模块

#### 1. 后端架构 (Go)

- **SpeedTester**: 核心测速引擎，集成 Mihomo (Clash) 核心
- **Unlock Detector**: 30+ 平台的流媒体解锁检测
- **WebSocket Server**: 实时通信服务
- **Task Manager**: 异步任务调度和管理
- **Config Loader**: 支持本地文件和远程订阅
- **Export Utils**: 结果导出和格式化

#### 2. 前端架构 (React/TypeScript)

- **SpeedTest**: 主测试控制组件
- **RealTimeProgressTable**: 实时进度和结果展示
- **WebSocket Hook**: 实时通信状态管理
- **UI Components**: 基于 shadcn/ui 的组件库
- **Export System**: 智能文件导出功能

#### 3. 构建系统 (Turborepo)

- **并行构建**: 前后端同时构建优化
- **智能缓存**: 增量构建和任务缓存
- **Docker 集成**: 多阶段构建优化
- **CI/CD 集成**: GitHub Actions 自动化

## 📡 API 文档

### REST API 接口

```bash
# 获取节点列表
POST /api/nodes
Content-Type: application/json
{
  "configPaths": "config.yaml",
  "stashCompatible": false
}

# 开始异步测试
POST /api/test/async
Content-Type: application/json
{
  "configPaths": "config.yaml",
  "testMode": "both",           # both/speed_only/unlock_only
  "concurrent": 4,
  "timeout": 10,
  "unlockPlatforms": ["Netflix", "YouTube"],
  "unlockConcurrent": 5,
  "unlockTimeout": 10
}

# 获取解锁检测平台列表
GET /api/unlock/platforms

# 检查 TUN 模式状态
GET /api/tun-check

# 系统信息
GET /api/system/info
```

### WebSocket API

```bash
# 连接 WebSocket
ws://localhost:8080/ws

# 测试进度消息
{
  "type": "test_progress",
  "data": {
    "current_proxy": "节点名称",
    "completed_count": 5,
    "total_count": 20,
    "progress_percent": 25.0,
    "status": "testing",
    "current_stage": "speed_test",
    "unlock_platform": "Netflix"
  }
}

# 测试结果消息
{
  "type": "test_result",
  "data": {
    "proxy_name": "节点名称",
    "proxy_type": "vmess",
    "proxy_ip": "1.2.3.4",
    "download_speed_mbps": 15.67,
    "upload_speed_mbps": 8.32,
    "latency_ms": 120,
    "jitter_ms": 5.2,
    "packet_loss": 0.1,
    "unlock_results": [
      {
        "platform": "Netflix",
        "supported": true,
        "region": "US"
      }
    ],
    "status": "success"
  }
}

# 测试完成消息
{
  "type": "test_complete",
  "data": {
    "successful_tests": 18,
    "failed_tests": 2,
    "total_tested": 20,
    "total_duration": "2分30秒",
    "average_latency": 156.5,
    "average_download_mbps": 45.8,
    "average_upload_mbps": 18.3,
    "best_proxy": "最快节点名称",
    "best_download_speed_mbps": 78.9,
    "unlock_stats": {
      "successful_unlock_tests": 25,
      "total_unlock_tests": 40,
      "best_unlock_proxy": "解锁最多的节点",
      "best_unlock_platforms": ["Netflix", "YouTube", "Disney+"]
    }
  }
}
```

## 🔧 开发环境配置

### 前置要求

- **Go**: 1.19+ (后端开发)
- **Node.js**: 18.0+ (前端开发)
- **pnpm**: 8.0+ (包管理器)
- **Docker**: 20.0+ (可选，容器化部署)

### 快速开始

#### 1. 克隆项目

```bash
git clone https://github.com/zhsama/clash-speedtest.git
cd clash-speedtest
```

#### 2. 安装依赖

```bash
# 安装所有依赖 (前端+后端)
pnpm install
```

#### 3. 启动开发环境

```bash
# 方式一：同时启动前后端 (推荐)
pnpm dev

# 方式二：分别启动
pnpm dev:backend   # 启动后端 API 服务 (端口 8080)
pnpm dev:frontend  # 启动前端界面 (端口 3000)

# 方式三：仅启动后端 API
pnpm dev:api
```

#### 4. 访问应用

- **前端界面**: <http://localhost:3000>
- **后端 API**: <http://localhost:8080>
- **API 文档**: <http://localhost:8080/api/docs>

### 项目脚本

```bash
# 开发相关
pnpm dev              # 启动完整开发环境
pnpm dev:frontend     # 仅启动前端
pnpm dev:backend      # 仅启动后端
pnpm debug            # 启动调试模式

# 构建相关
pnpm build            # 构建前后端
pnpm build:frontend   # 仅构建前端
pnpm build:backend    # 仅构建后端
pnpm build:docker     # Docker 镜像构建

# 质量控制
pnpm test             # 运行所有测试
pnpm lint             # 代码检查
pnpm typecheck        # 类型检查
pnpm format           # 代码格式化

# 清理
pnpm clean            # 清理构建文件
pnpm clean:cache      # 清理 Turbo 缓存
```

### VS Code 调试配置

项目已配置完整的 VS Code 调试环境：

1. **调试后端**: 按 F5 选择 "Debug Backend" 配置
2. **调试前端**: 按 F5 选择 "Debug Frontend" 配置
3. **调试 Delve**: 使用 "Attach to Delve" 配置进行深度调试

### 环境变量配置

```bash
# 前端环境变量 (frontend/.env.local)
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080

# 后端环境变量
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
LOGGER_LEVEL=INFO
LOGGER_OUTPUT_TO_FILE=true
```

## 🐳 Docker 部署

### 快速启动

```bash
# 开发环境
docker-compose up -d

# 生产环境 (优化版本)
docker-compose -f docker-compose.optimized.yml up -d
```

### 构建镜像

```bash
# 构建所有镜像
pnpm build:docker

# 手动构建
docker build -t clash-speedtest-backend ./backend
docker build -t clash-speedtest-frontend ./frontend
```

### Docker 特性

- **多阶段构建**: 最小化镜像大小
- **UPX 压缩**: 二进制文件压缩减少 60%+ 体积
- **Distroless 基础镜像**: 提升安全性
- **健康检查**: 自动服务状态监控

## 📋 配置文件

### 后端配置 (backend/config.yaml)

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  cors:
    enabled: true
    allowed_origins: ["http://localhost:3000"]

logger:
  level: "INFO"                    # DEBUG/INFO/WARN/ERROR
  output_to_file: true
  log_dir: "logs"
  log_file_name: "clash-speedtest.log"
  max_size: 10485760              # 10MB
  max_files: 5
  rotate_on_start: true
  enable_console: true
  format: "text"                  # text/json

unlock:
  cache_enabled: true
  cache_duration: "1h"
  timeout: "10s"
  retry_count: 3
  concurrent: 5
```

### 前端配置 (frontend/astro.config.mjs)

```javascript
export default defineConfig({
  integrations: [
    react(),
    tailwind({ applyBaseStyles: false })
  ],
  server: {
    port: 3000,
    host: true
  },
  vite: {
    define: {
      'import.meta.env.VITE_API_URL': JSON.stringify(process.env.VITE_API_URL || 'http://localhost:8080'),
      'import.meta.env.VITE_WS_URL': JSON.stringify(process.env.VITE_WS_URL || 'ws://localhost:8080')
    }
  }
})
```

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
pnpm test

# 后端测试
cd backend && go test ./...

# 前端测试
cd frontend && pnpm test

# 端到端测试
pnpm test:e2e
```

### 性能测试

```bash
# 测速性能基准
go run main.go -c config.yaml -concurrent 16

# 内存使用监控
go run main.go -c config.yaml -memprofile mem.prof

# Docker 镜像大小测试
./scripts/analyze-docker-size.sh
```

### 测试策略

1. **单元测试**: 核心功能模块测试
2. **集成测试**: API 接口和 WebSocket 测试
3. **端到端测试**: 完整用户流程测试
4. **性能测试**: 并发和内存使用测试
5. **Docker 测试**: 容器化部署测试

## 🔍 故障排除

### 常见问题

**Q: 测试结果不准确怎么办？**
A: 建议关闭系统的 TUN 模式，使用 Stash 兼容模式，应用会自动检测并提醒

**Q: 订阅链接无法获取节点？**
A: 确保订阅链接包含 `&flag=meta` 参数，支持逗号分隔多个配置源

**Q: WebSocket 连接失败？**
A: 检查防火墙设置，确保 8080 端口未被占用，查看浏览器控制台错误信息

**Q: 前端无法连接后端？**
A: 检查后端是否正常启动，确认环境变量中的 API 地址配置正确

**Q: 编译失败？**
A: 确保 Go 版本 >= 1.19，Node.js >= 18.0，运行 `go mod tidy && pnpm install`

**Q: Docker 构建失败？**
A: 检查 Docker 版本，确保支持 multi-stage builds，查看构建日志

### 调试模式

```bash
# 启用详细日志
go run main.go -config=config-debug.yaml

# 使用 Delve 调试器
dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient main.go -- -config=config-debug.yaml

# 前端调试
cd frontend && pnpm dev --debug

# 查看构建缓存
pnpm turbo:info
```

### 日志分析

```bash
# 查看后端日志
tail -f backend/logs/clash-speedtest.log

# 查看 Docker 容器日志
docker-compose logs -f backend
docker-compose logs -f frontend

# 查看构建日志
pnpm build 2>&1 | tee build.log
```

## 📈 性能优化建议

### 测试参数优化

1. **并发数调整**: 根据网络条件调整 concurrent 参数 (推荐 4-8)
2. **超时设置**: 合理设置 timeout 跳过慢速节点 (推荐 10-30s)
3. **包大小**: 根据带宽调整 downloadSize (10-100MB)
4. **解锁并发**: 解锁检测并发数 (推荐 3-5)

### 系统优化

1. **内存管理**: 大量节点测试时适当降低并发数
2. **网络优化**: 使用有线网络，关闭其他网络应用
3. **系统配置**: 关闭 TUN 模式获得更准确结果
4. **代理设置**: 避免使用系统代理影响测试结果

### 构建优化

1. **Turbo 缓存**: 利用 Turborepo 增量构建
2. **Docker 优化**: 多阶段构建减少镜像大小
3. **并行构建**: 前后端并行构建提升效率
4. **依赖优化**: 定期清理和更新依赖

## 🧠 测速原理

### 测试机制

通过 HTTP GET/POST 请求测试节点性能，默认使用 <https://speed.cloudflare.com> 进行测试。

### 测试指标说明

1. **下载速度**: 下载指定大小文件的速度，反映节点的出口带宽
2. **上传速度**: 上传指定大小文件的速度，反映节点的上传带宽  
3. **延迟(Latency)**: HTTP GET 请求的 TTFB（Time To First Byte），反映网络延迟
4. **抖动(Jitter)**: 延迟的变化幅度，反映网络稳定性
5. **丢包率**: 数据包丢失的百分比，反映网络质量
6. **解锁状态**: 各流媒体平台的访问检测结果

### 解锁检测原理

通过访问各平台的特定检测端点，分析返回内容判断解锁状态：

- **Netflix**: 检测地区库可用性
- **YouTube**: 检测地区限制内容
- **Disney+**: 检测服务可用性和地区
- **ChatGPT**: 检测 API 访问限制
- **其他平台**: 根据各平台特性进行专门检测

### 重要说明

请注意带宽跟延迟是两个独立的指标：

1. **高带宽 + 高延迟**: 下载快但网页打开慢 (中转节点无 BGP 加速)
2. **低带宽 + 低延迟**: 网页打开快但下载慢 (IEPL/IPLC 带宽较小)

### 自建测速服务器

```bash
# 在测速服务器上安装和启动
go install github.com/zhsama/clash-speedtest/download-server@latest
download-server

# 使用自建服务器测试
clash-speedtest --server-url "http://your-server-ip:8080"
```

## 🤝 贡献指南

我们欢迎所有形式的贡献！

### 开发流程

1. Fork 项目到你的 GitHub 账户
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 进行开发并测试
4. 提交更改 (`git commit -m 'feat: add amazing feature'`)
5. 推送到分支 (`git push origin feature/amazing-feature`)
6. 创建 Pull Request

### 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```bash
feat: 新功能
fix: 修复
docs: 文档更新
style: 代码格式
refactor: 重构
test: 测试
chore: 构建工具、辅助工具变动
```

### 代码规范

```bash
# 后端代码检查
cd backend
go fmt ./...
go vet ./...
golangci-lint run

# 前端代码检查
cd frontend
pnpm lint
pnpm typecheck
pnpm format
```

### 开发建议

1. **单一职责**: 每个 PR 专注于单一功能或修复
2. **测试覆盖**: 为新功能添加相应测试
3. **文档更新**: 更新相关文档和 README
4. **向后兼容**: 避免破坏性变更
5. **性能考虑**: 注意新功能对性能的影响

## 📋 TODO List

### 🐳 Docker 优化计划

- [ ] **容器编排优化**
  - [ ] 添加 Kubernetes 部署配置
  - [ ] 优化 Docker Compose 健康检查
  - [ ] 集成 Docker Swarm 支持
  - [ ] 添加容器监控和日志聚合

- [ ] **镜像优化**
  - [ ] 进一步减少镜像大小 (目标 < 20MB)
  - [ ] 添加多架构支持 (ARM64/AMD64)
  - [ ] 实现镜像安全扫描
  - [ ] 优化层缓存策略

### 🔓 流媒体解锁检测完善

- [ ] **平台扩展**
  - [ ] 添加更多国际平台 (Crunchyroll, Funimation, VRV)
  - [ ] 支持中国大陆平台 (爱奇艺, 腾讯视频, 优酷)
  - [ ] 添加音乐平台检测 (Apple Music, Pandora)
  - [ ] 支持游戏平台检测 (Steam, Epic Games)

- [ ] **检测能力增强**
  - [ ] 实现地区精确检测 (具体到城市)
  - [ ] 添加解锁质量评估 (4K, HDR 支持)
  - [ ] 支持自定义检测规则
  - [ ] 实现批量平台检测优化

- [ ] **解锁结果改进**
  - [ ] 添加历史解锁记录对比
  - [ ] 实现解锁状态变化通知
  - [ ] 支持解锁结果导出和分享
  - [ ] 添加解锁稳定性评分

### 🎨 前端UI设计重构

- [ ] **设计系统升级**
  - [ ] 实现完整的 Design System
  - [ ] 添加深色/浅色主题切换
  - [ ] 优化移动端体验和手势操作
  - [ ] 实现无障碍访问 (WCAG 2.1 AA)

- [ ] **交互体验优化**
  - [ ] 重新设计测试进度展示
  - [ ] 添加数据可视化图表 (Chart.js/D3.js)
  - [ ] 实现拖拽排序和自定义面板
  - [ ] 优化加载状态和错误处理

- [ ] **功能界面完善**
  - [ ] 添加节点地图可视化
  - [ ] 实现测试历史记录管理
  - [ ] 支持多配置文件管理
  - [ ] 添加高级设置面板

- [ ] **性能和用户体验**
  - [ ] 实现虚拟滚动优化大量节点显示
  - [ ] 添加离线模式支持
  - [ ] 优化首屏加载速度
  - [ ] 实现渐进式 Web 应用 (PWA)

### 🚀 其他功能计划

- [ ] **核心功能增强**
  - [ ] 支持自定义测试脚本
  - [ ] 添加定时测试任务
  - [ ] 实现测试结果对比分析
  - [ ] 支持分布式测试架构

- [ ] **集成和扩展**
  - [ ] 添加 Webhook 通知支持
  - [ ] 集成主流代理管理工具
  - [ ] 支持 API 密钥认证
  - [ ] 实现插件系统架构

- [ ] **运维和监控**
  - [ ] 添加 Prometheus 指标导出
  - [ ] 实现 Grafana 监控面板
  - [ ] 添加日志分析和搜索
  - [ ] 支持性能基准测试

## 🌟 功能规划

### 短期计划 (1-3 个月)

- [ ] 完善流媒体解锁检测
- [ ] 优化 Docker 构建流程
- [ ] 重构前端 UI 设计
- [ ] 添加更多测试指标
- [ ] 实现测试结果历史记录

### 中期计划 (3-6 个月)

- [ ] 支持自定义测试规则
- [ ] 添加 API 认证和权限管理
- [ ] 实现分布式测试架构
- [ ] 集成更多代理协议
- [ ] 添加移动端原生应用

### 长期计划 (6-12 个月)

- [ ] 支持插件系统
- [ ] 实现 AI 智能推荐
- [ ] 添加社区功能
- [ ] 支持企业级部署
- [ ] 集成云服务提供商

## 📄 许可证

本项目基于 [GPL-3.0](LICENSE) 许可证开源。

### 许可证说明

- ✅ 商业使用: 允许
- ✅ 修改: 允许
- ✅ 分发: 允许
- ✅ 专利使用: 允许
- ✅ 私人使用: 允许
- ❗ 披露源码: 必须
- ❗ 许可证和版权声明: 必须
- ❗ 相同许可证: 必须

## 🙏 致谢

感谢以下开源项目和贡献者：

### 核心依赖

- [Mihomo](https://github.com/metacubex/mihomo) - Clash 核心实现
- [Gin](https://github.com/gin-gonic/gin) - Go Web 框架
- [React](https://reactjs.org/) - 前端框架
- [TypeScript](https://www.typescriptlang.org/) - 类型安全的 JavaScript
- [Astro](https://astro.build/) - 现代静态站点生成器

### 构建工具

- [Turborepo](https://turbo.build/) - 高性能构建系统
- [Vite](https://vitejs.dev/) - 现代前端构建工具
- [GoReleaser](https://goreleaser.com/) - 自动化发布工具
- [Docker](https://www.docker.com/) - 应用容器化平台

### UI 和样式

- [Tailwind CSS](https://tailwindcss.com/) - 实用优先的 CSS 框架
- [shadcn/ui](https://ui.shadcn.com/) - 现代化 React 组件库
- [Lucide React](https://lucide.dev/) - 优雅的图标库
- [Sonner](https://sonner.emilkowal.ski/) - 现代化 Toast 组件

### 特别感谢

- 所有贡献者和 Beta 测试用户
- 开源社区的支持和反馈
- Clash/Mihomo 开发团队
- 各流媒体平台的解锁检测参考

## 📞 支持与反馈

### 获取帮助

- 🐛 [问题反馈](https://github.com/zhsama/clash-speedtest/issues)
- 💬 [讨论区](https://github.com/zhsama/clash-speedtest/discussions)
- 📚 [文档中心](https://github.com/zhsama/clash-speedtest/tree/main/docs)
- 🔧 [开发指南](CLAUDE.md)

### 联系方式

- **GitHub Issues**: 技术问题和 Bug 报告
- **GitHub Discussions**: 功能建议和使用交流
- **Email**: 通过 GitHub Issues 联系维护者

### 反馈渠道

1. **Bug 报告**: 详细描述问题和复现步骤
2. **功能建议**: 说明需求和使用场景
3. **使用问题**: 查看文档或在讨论区提问
4. **贡献代码**: 参考贡献指南提交 PR

---

⭐ **如果这个项目对您有帮助，请给我们一个 Star！**

**Made with ❤️ by zhsama**
