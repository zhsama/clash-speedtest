ARG GO_VERSION=1.24
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
# Stage 4: 生产运行时
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

# 复制二进制文件
COPY --from=builder /build/clash-speedtest ./

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