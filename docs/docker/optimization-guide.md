# Docker 构建优化指南

本文档详细说明了如何通过多层构建和其他优化技术来减小 Docker 镜像体积。

## 优化成果对比

### Backend 镜像体积优化

| 构建策略 | 基础镜像 | 预期体积 | 特点 |
|---------|---------|---------|------|
| 原始 Alpine | alpine:latest | ~50-60MB | 包含包管理器和 shell |
| Scratch + UPX | scratch | ~5-8MB | 最小体积，UPX 压缩 |
| Distroless | gcr.io/distroless/static | ~10-15MB | 安全性好，兼容性佳 |
| Chainguard | cgr.dev/chainguard/static | ~8-12MB | 最安全，定期更新 |

### Frontend 镜像体积优化

| 构建策略 | 基础镜像 | 预期体积 | 特点 |
|---------|---------|---------|------|
| 原始 nginx | nginx:alpine | ~40-50MB | 完整 nginx 功能 |
| 优化 nginx | nginx:alpine-slim | ~20-30MB | 精简版 nginx |

## 优化技术详解

### 1. 多阶段构建优化

```dockerfile
# 阶段 1: 模块缓存层
FROM golang:1.23-alpine AS modules
WORKDIR /modules
COPY go.mod go.sum ./
RUN go mod download

# 阶段 2: 构建层
FROM golang:1.23-alpine AS builder
# 复用模块缓存
COPY --from=modules /go/pkg /go/pkg
```

**优势**：
- 分离依赖下载和代码构建
- 更好的层缓存利用
- 减少重复下载

### 2. 构建参数优化

```bash
# Go 构建优化参数
CGO_ENABLED=0      # 禁用 CGO，生成纯静态二进制
-ldflags="-s -w"   # 去除符号表和调试信息
-trimpath          # 去除文件路径信息
-tags netgo        # 使用纯 Go 网络实现
```

### 3. UPX 压缩（可选）

```dockerfile
RUN upx --ultra-brute -qq binary
```

**优势**：
- 可减少 50-70% 的二进制体积
- 对运行性能影响极小

**劣势**：
- 某些环境可能不兼容
- 增加构建时间

### 4. 基础镜像选择

#### Scratch（最小）
```dockerfile
FROM scratch
# 需要手动添加证书和用户信息
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
```

#### Distroless（推荐）
```dockerfile
FROM gcr.io/distroless/static:nonroot
# 已包含证书和非 root 用户
```

#### Chainguard（最安全）
```dockerfile
FROM cgr.dev/chainguard/static:latest
# 零 CVE，定期更新
```

### 5. BuildKit 高级特性

启用 BuildKit：
```bash
export DOCKER_BUILDKIT=1
```

使用缓存挂载：
```dockerfile
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build
```

### 6. 前端优化策略

- 分离生产和开发依赖
- 使用 nginx:alpine-slim
- 删除不必要的 nginx 模块
- 启用 gzip 压缩

## 使用方法

### 构建优化镜像

```bash
# 使用优化的 Dockerfile 构建
docker build -f backend/Dockerfile -t clash-backend:optimized backend/
docker build -f frontend/Dockerfile -t clash-frontend:optimized frontend/

# 使用 BuildKit 构建（推荐）
DOCKER_BUILDKIT=1 docker build -f backend/Dockerfile.buildkit -t clash-backend:buildkit backend/

# 使用 docker-compose 构建
docker-compose -f docker-compose.optimized.yml build
```

### 运行分析脚本

```bash
# 分析不同构建策略的体积差异
./analyze-docker-size.sh
```

### 验证镜像体积

```bash
# 查看镜像大小
docker images | grep clash

# 查看镜像层信息
docker history clash-backend:optimized

# 使用 dive 工具深入分析（需要安装 dive）
dive clash-backend:optimized
```

## 安全性考虑

1. **使用非 root 用户**：所有镜像都配置了非 root 用户运行
2. **只读文件系统**：生产环境启用 `read_only: true`
3. **最小权限原则**：使用 `no-new-privileges` 安全选项
4. **定期扫描**：集成 Trivy 或 Snyk 进行漏洞扫描

## 性能优化建议

1. **使用 .dockerignore**：排除不必要的文件
2. **合理排序 Dockerfile 指令**：将不常变化的指令放在前面
3. **使用特定版本标签**：避免使用 `latest` 标签
4. **启用 BuildKit**：获得更好的缓存和并行构建
5. **多平台构建**：支持 ARM64 等架构

```bash
# 多平台构建示例
docker buildx build --platform linux/amd64,linux/arm64 -t clash-backend:multi .
```

## 故障排除

### UPX 压缩后无法运行
- 某些环境不支持 UPX 压缩的二进制
- 解决方案：使用 Distroless 版本

### 证书错误
- Scratch 镜像需要手动复制 CA 证书
- 解决方案：确保复制 ca-certificates.crt

### 时区问题
- 精简镜像可能缺少时区数据
- 解决方案：从 alpine 复制 /usr/share/zoneinfo

## 最佳实践总结

1. **生产环境推荐**：
   - Backend: Distroless 或 Chainguard
   - Frontend: nginx:alpine-slim
   
2. **开发环境**：
   - 保留调试工具和 shell
   - 使用 Alpine 基础镜像

3. **CI/CD 集成**：
   - 使用 BuildKit 缓存
   - 实施多阶段构建
   - 自动化安全扫描

通过这些优化，我们可以将后端镜像从 50-60MB 减小到 8-15MB，前端镜像从 40-50MB 减小到 20-30MB，同时保持安全性和功能完整性。