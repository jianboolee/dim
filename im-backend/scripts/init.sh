#!/bin/bash

# 项目初始化脚本
# 用于设置开发环境，创建配置文件，安装依赖等

set -e  # 遇到错误时退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT=$(cd $(dirname $0)/.. && pwd)

# 打印函数
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_step() {
    echo -e "${BLUE}🔧 $1${NC}"
}

# 环境变量文件路径
ENV_FILE=".env"

# 主要初始化函数
init_project() {
    print_step "初始化项目..."
    
    # 检查并创建环境变量文件
    setup_env_file
    
    # 安装依赖
    install_dependencies
    
    # 创建必要的目录
    create_directories
    
    # 检查并提示额外配置
    check_additional_setup
    
    print_success "项目环境初始化完成"
    
    # 显示后续步骤提示
    show_next_steps
}

# 设置环境变量文件
setup_env_file() {
    print_step "设置环境变量文件..."

    if [ ! -f "$ENV_FILE" ]; then
        print_info "创建环境变量文件..."
        
        if [ -f ".env.example" ]; then
            cp .env.example "$ENV_FILE"
            print_success "已创建 .env 文件（从 .env.example 复制）"
        else
            print_warning "未找到模板文件，创建空的 .env 文件"
            touch "$ENV_FILE"
            echo "# 项目环境变量配置文件" > "$ENV_FILE"
            echo "# 请根据需要添加配置项" >> "$ENV_FILE"
        fi
        
        print_warning "💡 请根据需要修改 .env 文件中的配置"
    else
        print_success ".env 文件已存在， 请根据需要修改 .env 文件中的配置"
    fi

    # 加载环境变量
    if [ -f "$ENV_FILE" ]; then
        set -a
        source "$ENV_FILE"
        set +a
        print_success "✅ 环境变量已加载"
    fi
}

# 安装依赖
install_dependencies() {
    print_info "安装Go模块依赖..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go 未安装，请先安装 Go"
        exit 1
    fi
    
    # 下载依赖
    go mod download
    
    # 整理依赖
    go mod tidy
    
    print_success "依赖安装完成"
}

# 创建必要的目录
create_directories() {
    print_info "创建必要的目录..."
    
    directories=(
        "bin"
        "logs"
        "tmp"
    )
    
    for dir in "${directories[@]}"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir"
            print_success "创建目录: $dir"
        fi
    done
}

# 检查额外配置
check_additional_setup() {
    print_info "检查额外配置..."
    
    # 检查是否需要生成密钥
    if [ ! -f "$JWT_PUBLIC_KEY_PATH" ]; then
        print_info "没有找到JWT公钥！"
    fi

    # 检查开发工具
    check_dev_tools
}

# 检查开发工具
check_dev_tools() {
    print_info "检查开发工具..."
    
    # 检查 air (热重载工具)
    if ! command -v air &> /dev/null; then
        print_warning "未安装 air 热重载工具"
        print_info "可运行: go install github.com/cosmtrek/air@latest"
    else
        print_success "air 热重载工具已安装"
    fi
    
    # # 检查 swag (API文档生成工具)
    # if ! command -v swag &> /dev/null; then
    #     print_warning "未安装 swag API文档生成工具"
    #     print_info "可运行: go install github.com/swaggo/swag/cmd/swag@latest"
    # else
    #     print_success "swag API文档生成工具已安装"
    # fi
}

# 运行数据库迁移
run_migrate() {
    print_step "运行数据库迁移..."
    
    # 检查migrate命令是否存在
    if [ ! -f "cmd/migrate/main.go" ]; then
        print_error "未找到数据库迁移文件 cmd/migrate/main.go"
        return 1
    fi
    
    print_info "开始执行数据库迁移..."
    
    # 运行迁移命令
    if go run cmd/migrate/main.go; then
        print_success "数据库迁移完成"
        return 0
    else
        print_error "数据库迁移失败, 请修改 .env 文件中的数据库连接信息后重试"
        return 1
    fi
}

# 显示帮助信息
show_help() {
    echo "项目初始化脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help        显示帮助信息"
    echo "  --dev-tools       安装开发工具"
    echo "  --migrate         运行数据库迁移"
    echo ""
    echo "功能:"
    echo "  - 创建环境变量文件 (.env)"
    echo "  - 安装Go模块依赖"
    echo "  - 创建必要的目录结构"
    echo "  - 检查开发工具安装状态"
    echo "  - 检查和配置数据库连接"
    echo "  - 执行数据库结构迁移"
}

# 安装开发工具
install_dev_tools() {
    print_step "安装开发工具..."
    
    # 安装 air
    if ! command -v air &> /dev/null; then
        print_info "安装 air 热重载工具..."
        go install github.com/cosmtrek/air@latest
        print_success "air 安装完成"
    fi
    
}

# 显示后续步骤
show_next_steps() {
    echo ""
    print_info "🎉 项目初始化完成！"
    echo ""
    print_info "后续步骤:"
    echo "  1. 检查并修改 .env 文件中的配置"
    echo "  2. 确保数据库服务正在运行"
    echo "  3. 运行以下命令启动项目:"
    echo "     • make run          - 普通运行"
    echo "     • make dev          - 开发模式（热重载）"
    echo "     • make build-start  - 构建并运行"
    echo ""
    print_info "数据库相关命令:"
    echo "     • scripts/init.sh --migrate     - 运行数据库迁移"
    echo ""
    print_info "其他有用命令:"
    echo "     • scripts/init.sh --dev-tools   - 安装开发工具"
    echo "     • make test                      - 运行测试"
    echo ""
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        --dev-tools)
            setup_env_file  # 确保环境变量加载
            install_dev_tools
            exit 0
            ;;
        --migrate)
            setup_env_file  # 确保环境变量加载
            run_migrate
            exit $?
            ;;
        *)
            print_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 执行主初始化流程
init_project 