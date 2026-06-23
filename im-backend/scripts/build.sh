#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 打印带颜色的消息函数
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# 确保在项目根目录执行
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
    print_error "请在项目根目录执行此脚本"
    exit 1
fi

# 切换到项目根目录
cd "$PROJECT_ROOT"
print_info "工作目录: $(pwd)"

# 设置变量
APP_NAME="d-im"
API_SERVER_NAME="api-server"
WS_SERVER_NAME="ws-server"
UNIFIED_SERVER_NAME="unified-server"
BUILD_DIR="bin"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")

# 创建构建目录
print_step "创建构建目录..."
mkdir -p $BUILD_DIR

print_info "Building $APP_NAME version $VERSION..."

# 构建API服务器
print_step "构建 API 服务器..."
if go build -ldflags "-X main.version=$VERSION" -o $BUILD_DIR/$API_SERVER_NAME ./cmd/api-server; then
    print_success "API 服务器构建成功"
else
    print_error "API 服务器构建失败"
    exit 1
fi

# 构建WebSocket服务器
print_step "构建 WebSocket 服务器..."
if go build -ldflags "-X main.version=$VERSION" -o $BUILD_DIR/$WS_SERVER_NAME ./cmd/ws-server; then
    print_success "WebSocket 服务器构建成功"
else
    print_error "WebSocket 服务器构建失败"
    exit 1
fi

# 构建统一服务器（可选）
print_step "构建统一服务器..."
if go build -ldflags "-X main.version=$VERSION" -o $BUILD_DIR/$UNIFIED_SERVER_NAME ./cmd/server; then
    print_success "统一服务器构建成功"
else
    print_error "统一服务器构建失败"
    exit 1
fi

# 复制config目录
if [[ -d "config" ]]; then
    cp -r config/ $BUILD_DIR/config/ 2>/dev/null
    print_success "config 目录复制成功"
else
    print_warning "未找到 config 目录"
fi

# 复制部署文件
print_step "复制部署文件..."
if [[ -f "deploy/docker-compose.yml" ]]; then
    cp docker-compose.yml $BUILD_DIR/ 2>/dev/null
    print_success "docker-compose.yml 复制成功"
else
    print_warning "未找到 docker-compose.yml"
fi

if [[ -f "deploy/nginx.conf" ]]; then
    cp deploy/nginx.conf $BUILD_DIR/ 2>/dev/null
    print_success "nginx.conf 复制成功"
else
    print_warning "未找到 nginx.conf"
fi

print_success "构建完成！"
print_info "构建产物在 $BUILD_DIR 目录:"
echo -e "  ${CYAN}- $API_SERVER_NAME:${NC} API 服务器二进制"
echo -e "  ${CYAN}- $WS_SERVER_NAME:${NC} WebSocket 服务器二进制"
echo -e "  ${CYAN}- $UNIFIED_SERVER_NAME:${NC} 统一服务器二进制 (API + WebSocket)" 