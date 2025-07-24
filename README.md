# Clash SpeedTest Pro

> 专业的代理节点性能测试工具 - Professional proxy speed testing tool

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.19-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-GPL--3.0-green)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/yourusername/clash-speedtest)

基于 Clash/Mihomo 核心的专业测速工具，提供命令行和现代化 Web 界面，支持实时进度显示和流媒体解锁检测。

## 🚀 核心特性

- **🎯 直接测试**: 无需额外配置，直接读取 Clash/Mihomo 配置文件或订阅链接
- **🌐 现代化界面**: 提供 React/TypeScript 构建的现代化 Web 界面
- **⚡ 高性能**: 支持并发测试，快速获取节点性能数据
- **🔓 解锁检测**: 内置流媒体解锁检测（Netflix、YouTube、Disney+、ChatGPT、Spotify、Bilibili）
- **📊 实时进度**: WebSocket 实时显示测试进度和结果
- **🔧 灵活过滤**: 支持多种过滤条件（速度、延迟、协议类型、节点名称等）
- **📱 响应式设计**: 完美适配桌面和移动设备
- **🖥️ 跨平台**: 支持 Windows、macOS、Linux
- **🛡️ 安全可靠**: 开源代码，本地运行，保护节点隐私

<img width="1332" alt="image" src="https://github.com/user-attachments/assets/fdc47ec5-b626-45a3-a38a-6d88c326c588">

## 📦 安装方法

### 方法一：Go Install (推荐)
```bash
go install github.com/zhsama/clash-speedtest@latest
```

### 方法二：预编译二进制文件
从 [Releases](https://github.com/zhsama/clash-speedtest/releases) 页面下载对应平台的二进制文件。

### 方法三：源码编译
```bash
git clone https://github.com/zhsama/clash-speedtest.git
cd clash-speedtest/backend
go build -o clash-speedtest .
```

### 方法四：开发环境安装
```bash
# 克隆仓库
git clone https://github.com/zhsama/clash-speedtest.git
cd clash-speedtest

# 安装依赖
pnpm install

# 启动完整开发环境（前端 + 后端）
pnpm dev
```

## 🎯 使用方法

### 命令行使用

```bash
# 查看帮助
clash-speedtest -h
Usage of clash-speedtest:
  -c string
        configuration file path, also support http(s) url
  -f string
        filter proxies by name, use regexp (default ".*")
  -server-url string
        server url for testing proxies (default "https://speed.cloudflare.com")
  -download-size int
        download size for testing proxies (default 50MB)
  -upload-size int
        upload size for testing proxies (default 20MB)
  -timeout duration
        timeout for testing proxies (default 5s)
  -concurrent int
        download concurrent size (default 4)
  -output string
        output config file path (default "")
  -stash-compatible
        enable stash compatible mode
  -max-latency duration
        filter latency greater than this value (default 800ms)
  -min-download-speed float
        filter speed less than this value(unit: MB/s) (default 5)
  -min-upload-speed float
        filter upload speed less than this value(unit: MB/s) (default 2)
  -rename
        rename nodes with IP location and speed

# 演示：

# 1. 测试全部节点，使用 HTTP 订阅地址
# 请在订阅地址后面带上 flag=meta 参数，否则无法识别出节点类型
> clash-speedtest -c 'https://domain.com/api/v1/client/subscribe?token=secret&flag=meta'

# 2. 测试香港节点，使用正则表达式过滤，使用本地文件
> clash-speedtest -c ~/.config/clash/config.yaml -f 'HK|港'
节点                                        	带宽          	延迟
Premium|广港|IEPL|01                        	484.80KB/s  	815.00ms
Premium|广港|IEPL|02                        	N/A         	N/A
Premium|广港|IEPL|03                        	2.62MB/s    	333.00ms
Premium|广港|IEPL|04                        	1.46MB/s    	272.00ms
Premium|广港|IEPL|05                        	3.87MB/s    	249.00ms

# 3. 当然你也可以混合使用
> clash-speedtest -c "https://domain.com/api/v1/client/subscribe?token=secret&flag=meta,/home/.config/clash/config.yaml"

# 4. 筛选出延迟低于 800ms 且下载速度大于 5MB/s 的节点，并输出到 filtered.yaml
> clash-speedtest -c "https://domain.com/api/v1/client/subscribe?token=secret&flag=meta" -output filtered.yaml -max-latency 800ms -min-speed 5
# 筛选后的配置文件可以直接粘贴到 Clash/Mihomo 中使用，或是贴到 Github\Gist 上通过 Proxy Provider 引用。

# 5. 使用 -rename 选项按照 IP 地区和下载速度重命名节点
clash-speedtest -c config.yaml -output result.yaml -rename
# 重命名后的节点名称格式：🇺🇸 US | ⬇️ 15.67 MB/s
# 包含国旗 emoji、国家代码和下载速度
```

### Web 界面使用

#### 启动 Web 服务
```bash
# 启动后端 API 服务
cd backend
go run main.go -config=config.yaml

# 或者使用调试模式
go run main.go -config=config-debug.yaml
```

#### 启动前端界面
```bash
# 启动前端开发服务器
cd frontend
pnpm dev

# 或者从根目录同时启动前端和后端
pnpm dev
```

#### 使用 Web 界面
1. 打开浏览器访问 `http://localhost:3000`
2. 在"配置获取"部分输入配置文件路径或订阅链接
3. 点击"获取配置"加载节点列表
4. 在右侧面板配置测试参数：
   - **测试模式**: 全面测试（测速+解锁）/ 仅测速 / 仅解锁检测
   - **节点过滤**: 包含/排除特定节点，协议类型过滤
   - **速度过滤**: 设置最低速度和最大延迟阈值
   - **高级配置**: 并发数、超时时间、测试包大小等
5. 点击"开始测试"开始测试，实时查看进度和结果
6. 测试完成后可以查看详细的测试报告

#### Web 界面特性
- **实时进度**: 通过 WebSocket 实时显示测试进度
- **节点预览**: 测试前预览符合条件的节点列表
- **智能过滤**: 支持按节点名称、协议类型、速度等多维度过滤
- **TUN 模式检测**: 自动检测并提醒 TUN 模式状态
- **响应式设计**: 完美适配桌面和移动设备

## 🏗️ 项目架构

### 目录结构
```
clash-speedtest/
├── backend/              # Go 后端
│   ├── main.go          # 主程序入口
│   ├── speedtester/     # 核心测速逻辑
│   ├── download-server/ # 可选的自托管测速服务器
│   ├── config*.yaml     # 配置文件
│   └── package.json     # 开发脚本
├── frontend/            # React/TypeScript 前端
│   ├── src/
│   │   ├── components/  # React 组件
│   │   ├── hooks/       # 自定义 Hooks
│   │   ├── lib/         # 工具库
│   │   └── styles/      # 样式文件
│   ├── public/          # 静态资源
│   └── package.json
├── docs/               # 文档
├── .vscode/            # VS Code 配置
├── turbo.json          # Turbo 配置
├── package.json        # 根目录配置
└── README.md
```

### 核心组件

1. **后端 (Go)**
   - **SpeedTester**: 核心测速引擎，集成 Mihomo (Clash) 核心
   - **Web API**: RESTful API 服务
   - **WebSocket**: 实时通信服务
   - **Download Server**: 可选的自托管测速服务器

2. **前端 (React/TypeScript)**
   - **SpeedTest**: 主测试组件
   - **RealTimeProgressTable**: 实时进度表格
   - **TUNWarning**: TUN 模式检测组件
   - **WebSocket Hook**: 实时通信管理

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

# 响应
{
  "success": true,
  "nodes": [
    {
      "name": "节点名称",
      "type": "vmess",
      "server": "server.com",
      "port": 443
    }
  ]
}

# 开始异步测试
POST /api/test/async
Content-Type: application/json

{
  "configPaths": "config.yaml",
  "testMode": "both",
  "concurrent": 4,
  "timeout": 10,
  "unlockPlatforms": ["Netflix", "YouTube"]
}

# 响应
{
  "success": true,
  "taskId": "uuid-task-id"
}

# 检查 TUN 模式状态
GET /api/tun-check

# 响应
{
  "success": true,
  "tun_enabled": false
}
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
    "current_stage": "speed_test"
  }
}

# 测试结果消息
{
  "type": "test_result",
  "data": {
    "proxy_name": "节点名称",
    "download_speed": 15.67,
    "upload_speed": 8.32,
    "latency": 120,
    "unlock_results": {
      "Netflix": "支持",
      "YouTube": "支持"
    }
  }
}

# 测试完成消息
{
  "type": "test_complete",
  "data": {
    "successful_tests": 18,
    "failed_tests": 2,
    "total_tests": 20
  }
}
```

## 🔧 开发环境配置

### 前置要求
- Go 1.19+
- Node.js 18+
- pnpm 8+

### 开发环境设置

#### 1. 克隆项目
```bash
git clone https://github.com/zhsama/clash-speedtest.git
cd clash-speedtest
```

#### 2. 安装依赖
```bash
# 安装前端依赖
pnpm install
```

#### 3. 配置环境变量
```bash
# 前端环境变量 (frontend/.env)
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

#### 4. 启动开发服务器
```bash
# 方式一：同时启动前端和后端
pnpm dev

# 方式二：分别启动
# 后端
cd backend && go run main.go -config=config-debug.yaml

# 前端
cd frontend && pnpm dev
```

### VS Code 调试配置

项目已配置 VS Code 调试环境：

1. **调试后端**：按 F5 选择 "Debug Backend" 配置
2. **调试前端**：按 F5 选择 "Debug Frontend" 配置
3. **调试 Delve**：使用 "Attach to Delve" 配置

### 代码规范

```bash
# 后端代码格式化
cd backend
go fmt ./...
go vet ./...

# 前端代码检查
cd frontend
pnpm lint
pnpm type-check
```

## 📋 配置文件

### 后端配置 (config.yaml)
```yaml
server:
  port: 8080
  host: "0.0.0.0"

logger:
  level: "INFO"
  output_to_file: true
  log_dir: "logs"
  log_file_name: "clash-speedtest.log"
  max_size: 10485760
  max_files: 5
  rotate_on_start: true
  enable_console: true
  format: "text"
```

### 环境变量
```bash
# 服务器配置
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# 日志配置
LOGGER_LEVEL=INFO
LOGGER_OUTPUT_TO_FILE=true
LOGGER_LOG_DIR=logs
```

## 🧪 测试

### 运行测试
```bash
# 后端测试
cd backend
go test ./...

# 前端测试
cd frontend
pnpm test

# 端到端测试
pnpm test:e2e
```

### 性能测试
```bash
# 测速性能基准
go run main.go -c config.yaml -concurrent 16

# 内存使用监控
go run main.go -c config.yaml -memprofile mem.prof
```

## 🔍 故障排除

### 常见问题

**Q: 测试结果不准确怎么办？**
A: 建议关闭系统的 TUN 模式，使用 Stash 兼容模式

**Q: 订阅链接无法获取节点？**
A: 确保订阅链接包含 `&flag=meta` 参数

**Q: WebSocket 连接失败？**
A: 检查防火墙设置，确保 8080 端口未被占用

**Q: 前端无法连接后端？**
A: 检查后端是否正常启动，确认 API 地址配置正确

**Q: 编译失败？**
A: 确保 Go 版本 >= 1.19，运行 `go mod tidy` 更新依赖

### 调试模式
```bash
# 启用详细日志
go run main.go -config=config-debug.yaml

# 使用 Delve 调试器
dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient main.go -- -config=config-debug.yaml
```

## 📈 性能优化建议

1. **并发优化**: 根据网络条件调整并发数参数
2. **超时设置**: 合理设置超时时间跳过慢速节点
3. **结果过滤**: 使用速度和延迟过滤减少无效节点
4. **内存优化**: 大量节点测试时适当降低并发数

## 🧠 测速原理

通过 HTTP GET 请求下载指定大小的文件，默认使用 https://speed.cloudflare.com (50MB) 进行测试，计算下载时间得到下载速度。

### 测试指标说明
1. **下载速度**: 下载指定大小文件的速度，反映节点的出口带宽
2. **上传速度**: 上传指定大小文件的速度，反映节点的上传带宽  
3. **延迟**: HTTP GET 请求的 TTFB（Time To First Byte），反映网络延迟
4. **解锁状态**: 各流媒体平台的解锁检测结果

### 重要说明
请注意带宽跟延迟是两个独立的指标，两者并不关联：
1. **高带宽 + 高延迟**: 下载速度快但网页打开慢，可能是中转节点无 BGP 加速，但出海线路带宽充足
2. **低带宽 + 低延迟**: 网页打开快但下载慢，可能是中转节点有 BGP 加速，但出海线路的 IEPL/IPLC 带宽较小

### 自建测速服务器
Cloudflare 是全球知名的 CDN 服务商，一般情况下无需自建测速服务器。如有需要，可以自行搭建：

```bash
# 在测速服务器上安装和启动
go install github.com/zhsama/clash-speedtest/download-server@latest
download-server

# 使用自建服务器测试
clash-speedtest --server-url "http://your-server-ip:8080"
```

## 🤝 贡献指南

我们欢迎所有形式的贡献！请阅读以下指南：

### 开发流程
1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 提交规范
- 使用清晰的提交信息
- 遵循现有的代码风格
- 添加必要的测试
- 更新相关文档

### 代码审查
- 所有 PR 都需要经过代码审查
- 确保所有测试通过
- 保持代码覆盖率

## 🌟 功能规划

### 短期计划
- [ ] 支持更多流媒体平台检测
- [ ] 增加批量配置管理
- [ ] 优化测试算法和性能
- [ ] 添加移动端适配

### 长期计划
- [ ] 支持自定义测试规则
- [ ] 集成更多代理协议
- [ ] 添加历史记录功能
- [ ] 支持分布式测试

## 📄 许可证

本项目基于 [GPL-3.0](LICENSE) 许可证开源。

## 🙏 致谢

感谢以下开源项目和贡献者：

- [Mihomo](https://github.com/metacubex/mihomo) - Clash 核心实现
- [GoReleaser](https://goreleaser.com/) - 自动化发布工具
- [React](https://reactjs.org/) - 前端框架
- [TypeScript](https://www.typescriptlang.org/) - 类型安全的 JavaScript
- [Vite](https://vitejs.dev/) - 现代前端构建工具
- [Tailwind CSS](https://tailwindcss.com/) - 实用优先的CSS框架

## 📞 支持与反馈

- 🐛 [问题反馈](https://github.com/zhsama/clash-speedtest/issues)
- 💬 [讨论区](https://github.com/zhsama/clash-speedtest/discussions)
- 📧 邮件支持：请通过 GitHub Issues 联系

---

⭐ 如果这个项目对您有帮助，请给我们一个 Star！

**Made with ❤️ by the Clash SpeedTest Team**
