# Go 后端调试指南

本文档介绍如何在 VSCode 中对 clash-speedtest 后端进行断点调试。

## 前提条件

1. 确保已安装 Go 扩展 (Go for Visual Studio Code)
2. 安装 delve 调试器：
   ```bash
   go install github.com/go-delve/delve/cmd/dlv@latest
   ```

## 调试配置

项目已配置了多种调试场景，位于 `.vscode/launch.json`：

### 1. Debug Backend (Development) - 推荐
- **用途**: 使用开发配置进行调试
- **配置文件**: `config-debug.yaml`
- **快捷键**: F5 或点击调试按钮

### 2. Debug Backend (Production Config)
- **用途**: 使用生产配置进行调试
- **配置文件**: `config.yaml`
- **适用场景**: 测试生产环境配置

### 3. Debug Backend (Custom Config)
- **用途**: 使用自定义配置文件
- **配置文件**: 运行时输入
- **适用场景**: 测试特定配置

### 4. Debug Download Server
- **用途**: 调试下载服务器组件
- **程序**: `download-server/download-server.go`

### 5. Attach to Backend Process
- **用途**: 附加到正在运行的后端进程
- **适用场景**: 调试已启动的服务

## 使用方法

### 方法一：VSCode 内置调试器（推荐）

1. 在 VSCode 中打开项目
2. 在代码中设置断点（点击行号左侧）
3. 按 F5 或点击调试面板中的"开始调试"
4. 选择调试配置（默认为 "Debug Backend (Development)"）
5. 程序将在断点处暂停，可以查看变量、调用栈等

### 方法二：命令行调试

```bash
# 在项目根目录
npm run debug:backend

# 或者直接在 backend 目录
cd backend
npm run debug
```

### 方法三：Turbo 集成调试

```bash
# 使用 turbo 运行调试
turbo run debug --filter=backend
```

## 调试技巧

### 1. 设置断点
- **行断点**: 点击行号左侧
- **条件断点**: 右键点击断点，选择"编辑断点"
- **日志断点**: 在断点处输出日志而不暂停程序

### 2. 查看变量
- **变量面板**: 左侧调试面板中的"变量"
- **监视表达式**: 添加自定义表达式进行监视
- **悬停查看**: 鼠标悬停在变量上查看值

### 3. 调用栈
- **调用栈面板**: 查看函数调用链
- **跳转到调用位置**: 点击调用栈中的项目

### 4. 调试控制
- **继续 (F5)**: 继续执行到下一个断点
- **单步执行 (F10)**: 逐行执行，不进入函数
- **单步进入 (F11)**: 进入函数内部
- **单步跳出 (Shift+F11)**: 跳出当前函数

## 常见调试场景

### 1. 调试 API 请求处理
```go
// 在 handleTestWithWebSocket 函数中设置断点
func handleTestWithWebSocket(w http.ResponseWriter, r *http.Request) {
    // 在这里设置断点，调试请求处理逻辑
    logger.Logger.Info("Speed test request received (WebSocket enabled)")
    // ...
}
```

### 2. 调试速度测试逻辑
```go
// 在 speedtester 包中设置断点
func (st *SpeedTester) TestProxies(proxies map[string]adapter.Proxy, callback func(*Result)) {
    // 在这里设置断点，调试测试逻辑
    // ...
}
```

### 3. 调试解锁检测
```go
// 在 unlock 包中设置断点
func DetectUnlock(proxy adapter.Proxy, config *UnlockTestConfig) []UnlockResult {
    // 在这里设置断点，调试解锁检测逻辑
    // ...
}
```

## 配置文件说明

### config-debug.yaml
开发环境配置，包含：
- 调试级别日志
- 详细错误信息
- 开发环境特定设置

### config.yaml
生产环境配置，包含：
- 优化的性能设置
- 生产级别日志
- 安全配置

## 故障排除

### 1. 调试器无法启动
- 确保 Go 扩展已正确安装
- 检查 delve 是否正确安装：`dlv version`
- 确保工作目录正确

### 2. 断点不生效
- 确保代码已保存
- 重新编译程序
- 检查断点是否在可执行代码行上

### 3. 变量无法查看
- 确保程序编译时包含调试信息
- 检查变量是否在当前作用域内
- 尝试使用监视表达式

## 远程调试

如果需要远程调试：

1. 启动远程调试服务器：
   ```bash
   dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient main.go -- -config=config-debug.yaml
   ```

2. 在 VSCode 中连接远程调试器（需要额外配置）

## 性能调试

对于性能相关的调试：

1. 使用 Go 的内置性能分析工具
2. 设置性能断点
3. 监控内存使用情况
4. 分析 goroutine 状态

---

## 快速参考

| 操作 | 快捷键 | 说明 |
|------|--------|------|
| 开始调试 | F5 | 启动调试会话 |
| 继续 | F5 | 继续执行 |
| 单步执行 | F10 | 逐行执行 |
| 单步进入 | F11 | 进入函数 |
| 单步跳出 | Shift+F11 | 跳出函数 |
| 停止调试 | Shift+F5 | 停止调试会话 |
| 重启调试 | Ctrl+Shift+F5 | 重新启动调试 |

通过这些配置和说明，您应该能够有效地调试 clash-speedtest 后端代码。