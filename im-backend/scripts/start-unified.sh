#!/bin/bash

# 统一服务器启动脚本（API + WebSocket）

# 设置环境变量
export API_SERVER_PORT=${API_SERVER_PORT:-8080}
export WS_SERVER_PORT=${WS_SERVER_PORT:-9000}
export MONGODB_URI=${MONGODB_URI:-"mongodb://localhost:27017/go_im"}
export REDIS_ADDR=${REDIS_ADDR:-"localhost:6379"}
export REDIS_PASSWORD=${REDIS_PASSWORD:-""}
export REDIS_DB=${REDIS_DB:-0}
export JWT_PUBLIC_KEY_PATH=${JWT_PUBLIC_KEY_PATH:-"./config/keys/public.pem"}

echo "Starting Unified Server (API + WebSocket) on port $API_SERVER_PORT..."

# 启动统一服务器
./unified-server 