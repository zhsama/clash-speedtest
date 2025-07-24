# Turborepo 构建系统使用指南

本项目使用 Turborepo 优化了整个构建流程，实现了前后端的并行构建和 Docker 镜像生成。

## 快速开始

### 一键构建所有内容

```bash
# 安装依赖
pnpm install

# 构建前后端 + Docker 镜像
pnpm run build:all

# 或使用 Make
make release
```

### 常用命令

| 命令 | 说明 |
|-----|------|
| `pnpm run build` | 构建前后端 |
| `pnpm run build:all` | 构建前后端 + Docker 镜像 |
| `pnpm run build:docker` | 仅构建 Docker 镜像 |
| `pnpm run docker:push` | 推送 Docker 镜像到仓库 |
| `pnpm run dev` | 启动前后端开发服务器 |
| `pnpm run test` | 运行所有测试 |
| `pnpm run lint` | 代码检查 |
| `pnpm run clean` | 清理构建产物 |

### 使用 Make 命令（推荐）

```bash
make help        # 查看所有可用命令
make build       # 构建项目
make docker      # 构建 Docker 镜像
make release     # 完整的发布构建
make clean       # 清理构建产物
```

## Turborepo 特性

### 1. 智能缓存

Turborepo 会自动缓存构建结果，当代码没有变化时会跳过构建：

```bash
# 第一次构建
pnpm run build  # 需要完整构建

# 第二次构建（无代码变化）
pnpm run build  # 直接使用缓存，秒级完成
```

### 2. 并行构建

前端和后端会并行构建，充分利用多核 CPU：

```bash
# 自动并行构建前后端
pnpm run build

# 查看构建日志
pnpm run build --verbose
```

### 3. 依赖感知

Turborepo 理解项目间的依赖关系，自动按正确顺序构建。

## Docker 集成

### 构建 Docker 镜像

```bash
# 构建所有镜像
pnpm run docker:build

# 构建特定镜像
cd backend && pnpm run docker:build
cd frontend && pnpm run docker:build
```

### 推送到镜像仓库

```bash
# 设置镜像仓库地址
export DOCKER_REGISTRY=ghcr.io/yourusername

# 推送镜像
pnpm run docker:push
```

## GitHub Actions 优化

### 1. 智能缓存策略

- pnpm 依赖缓存
- Turborepo 构建缓存
- Docker 层缓存
- Go 模块缓存

### 2. 并行作业

- 构建和测试并行执行
- 多平台 Docker 镜像并行构建

### 3. 条件执行

- PR 只运行测试
- 主分支构建 Docker 镜像
- Tag 触发完整发布流程

## 性能优化效果

通过 Turborepo 优化，我们实现了：

- **首次构建**: ~3-5 分钟
- **增量构建**: ~10-30 秒（有缓存）
- **并行度**: 前后端同时构建
- **Docker 构建**: 使用 BuildKit 缓存优化

## 开发工作流

### 1. 日常开发

```bash
# 启动开发服务器
make dev

# 或分别启动
make dev-frontend
make dev-backend
```

### 2. 提交前检查

```bash
# 运行所有检查
make test lint typecheck

# 或使用 pnpm
pnpm run ci:test
```

### 3. 发布流程

```bash
# 完整发布构建
make release

# 创建 tag 触发 CI 发布
git tag v1.0.0
git push origin v1.0.0
```

## 故障排除

### 清理缓存

```bash
# 清理 Turborepo 缓存
pnpm run clean:cache

# 完全清理
make clean-all
```

### 构建失败

```bash
# 查看详细日志
pnpm run build --verbose

# 禁用缓存重新构建
pnpm run build --force
```

### Docker 构建问题

```bash
# 启用 BuildKit 调试
DOCKER_BUILDKIT=1 BUILDKIT_PROGRESS=plain make docker
```

## 环境变量

| 变量 | 说明 | 默认值 |
|-----|------|--------|
| `TURBO_TOKEN` | Turborepo 远程缓存 token | - |
| `TURBO_TEAM` | Turborepo 团队名称 | - |
| `DOCKER_REGISTRY` | Docker 镜像仓库地址 | - |
| `NODE_ENV` | Node 环境 | development |
| `GOOS` | Go 目标系统 | linux |
| `GOARCH` | Go 目标架构 | amd64 |

## 最佳实践

1. **使用 Make 命令**: 简化常用操作
2. **保持依赖更新**: 定期运行 `pnpm update`
3. **利用缓存**: 不要随意清理缓存
4. **并行开发**: 前后端可以独立开发和测试
5. **CI/CD**: 充分利用 GitHub Actions 的并行能力

通过这些优化，整个构建流程变得更加高效和可靠。