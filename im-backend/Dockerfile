# 基础构建阶段
FROM golang:1.21-alpine AS base

# 安装构建依赖
RUN apk add --no-cache \
    gcc \
    musl-dev \
    vips-dev \
    pkgconfig

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# API 服务器构建阶段
FROM base AS api-server

RUN CGO_ENABLED=1 GOOS=linux go build -o api-server ./cmd/api-server

# WebSocket 服务器构建阶段
FROM base AS ws-server

RUN CGO_ENABLED=1 GOOS=linux go build -o ws-server ./cmd/ws-server

# 统一服务器构建阶段
FROM base AS unified-server

RUN CGO_ENABLED=1 GOOS=linux go build -o unified-server ./cmd/server

# API 服务器运行阶段
FROM alpine:latest AS api-server-runtime

# 安装运行时依赖
RUN apk add --no-cache \
    vips \
    vips-tools \
    ca-certificates

# 设置工作目录
WORKDIR /app

# 复制二进制文件和配置
COPY --from=api-server /app/api-server /app/

# 暴露端口
EXPOSE 8080

# 运行API服务器
CMD ["./api-server"]

# WebSocket 服务器运行阶段
FROM alpine:latest AS ws-server-runtime

# 安装运行时依赖
RUN apk add --no-cache \
    vips \
    vips-tools \
    ca-certificates

# 设置工作目录
WORKDIR /app

# 复制二进制文件和配置
COPY --from=ws-server /app/ws-server /app/

# 暴露端口
EXPOSE 9000

# 运行WebSocket服务器
CMD ["./ws-server"]

# 统一服务器运行阶段
FROM alpine:latest AS unified-server-runtime

# 安装运行时依赖
RUN apk add --no-cache \
    vips \
    vips-tools \
    ca-certificates

# 设置工作目录
WORKDIR /app

# 复制二进制文件和配置
COPY --from=unified-server /app/unified-server /app/

# 暴露端口
EXPOSE 8080

# 运行统一服务器
CMD ["./unified-server"] 