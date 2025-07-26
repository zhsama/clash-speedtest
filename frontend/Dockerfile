ARG NODE_VERSION=22
ARG NGINX_VERSION=1.25-alpine-slim

# ============================================================================
# Stage 1: 基础环境
# ============================================================================
FROM --platform=$BUILDPLATFORM node:${NODE_VERSION}-alpine AS base
RUN corepack enable

# ============================================================================
# Stage 2: 依赖安装（workspace 模式）
# ============================================================================
FROM base AS deps
WORKDIR /workspace

# 复制 workspace 配置（变化频率低，缓存友好）
COPY pnpm-workspace.yaml package.json pnpm-lock.yaml ./

# 复制前端包配置
COPY frontend/package.json ./frontend/

# 安装依赖（使用缓存挂载）
RUN --mount=type=cache,id=pnpm,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile --filter=frontend

# ============================================================================
# Stage 3: 构建阶段
# ============================================================================
FROM base AS builder
WORKDIR /workspace

# 复制依赖
COPY --from=deps /workspace/node_modules ./node_modules
COPY --from=deps /workspace/frontend/node_modules ./frontend/node_modules

# 复制源代码
COPY frontend/ ./frontend/

# 构建应用
WORKDIR /workspace/frontend
RUN pnpm build

# ============================================================================
# Stage 4: 生产运行时
# ============================================================================
FROM nginx:${NGINX_VERSION} AS runtime

# 复制构建产物和配置
COPY --from=builder /workspace/frontend/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/nginx.conf

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/ || exit 1

EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]