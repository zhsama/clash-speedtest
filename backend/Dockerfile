ARG GO_VERSION=1.24
ARG ENABLE_UPX=false
ARG VERSION=dev
ARG ALPINE_VERSION=3.19

# ============================================================================
# Stage 1: 基础构建环境
# ============================================================================
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS base
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /build

# ============================================================================
# Stage 2: 依赖下载（利用缓存优化）
# ============================================================================
FROM base AS modules
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

# ============================================================================
# Stage 3: 代码构建
# ============================================================================
FROM base AS builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION

# 复制源代码
COPY . .

# 构建主程序
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w -extldflags '-static' -X main.version=${VERSION} -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -trimpath \
    -tags "netgo osusergo" \
    -installsuffix netgo \
    -o clash-speedtest .

# ============================================================================
# Stage 4: 可选UPX压缩（生产环境可启用）
# ============================================================================
FROM builder AS compressor
ARG ENABLE_UPX
RUN if [ "${ENABLE_UPX}" = "true" ]; then \
        apk add --no-cache upx && \
        upx --ultra-brute -qq clash-speedtest && \
        upx -t clash-speedtest; \
    fi

# ============================================================================
# Stage 5: 生产运行时
# ============================================================================
FROM alpine:${ALPINE_VERSION} AS runtime

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    iproute2 \
    procps \
    curl \
    && adduser -D -s /bin/sh appuser

# 创建应用目录
WORKDIR /app

# 复制二进制文件（根据UPX设置选择源）
ARG ENABLE_UPX
COPY --from=compressor /build/clash-speedtest ./

# 修复权限
RUN chown -R appuser:appuser /app && \
    chmod +x /app/clash-speedtest

# 安全配置
USER appuser
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD ["/app/clash-speedtest", "-h"]

# 启动主程序
ENTRYPOINT ["/app/clash-speedtest"]

# ============================================================================
# 构建说明和使用示例
# ============================================================================
# 
# 基础构建：
# docker build -t clash-speedtest-backend .
#
# 启用UPX压缩（更小体积）：
# docker build --build-arg ENABLE_UPX=true -t clash-speedtest-backend .
#
# 自定义版本：
# docker build --build-arg VERSION=v1.0.0 -t clash-speedtest-backend:v1.0.0 .
#
# 镜像特点：
# - 基于 Alpine 镜像（支持TUN检测，包含系统工具）
# - 非root用户运行（安全性）
# - 支持跨平台构建（AMD64/ARM64）
# - BuildKit缓存优化（快速构建）
# - 可选UPX压缩（更小体积）
# - 内置健康检查
# - TUN模式完整支持（包含 ip, ps 等工具）
# ============================================================================