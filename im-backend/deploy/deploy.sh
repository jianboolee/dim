#!/bin/bash

# 配置信息
REMOTE_USER="root"
REMOTE_HOST="182.92.156.241"
REMOTE_PORT="2222"
REMOTE_DIR="/mnt/www/x.wanfangche.com/api/im"
REMOTE_CONNECTION="${REMOTE_USER}@${REMOTE_HOST}"

# SSH 和 SCP 的通用参数
SSH_OPTS="-p ${REMOTE_PORT}"
SCP_OPTS="-P ${REMOTE_PORT}"

# 颜色输出函数
print_message() {
    echo -e "\033[1;36m[部署消息] $1\033[0m"
}

print_error() {
    echo -e "\033[1;31m[错误] $1\033[0m"
}

print_success() {
    echo -e "\033[1;32m[成功] $1\033[0m"
}

# 检查远程服务器的 Docker 版本
print_message "检查远程服务器 Docker 环境..."
ssh ${SSH_OPTS} ${REMOTE_CONNECTION} "echo 'Docker 版本：' && docker --version && echo 'Docker Compose 版本：' && docker-compose --version"

# 检查远程目录是否存在，不存在则创建
print_message "检查远程目录..."
ssh ${SSH_OPTS} ${REMOTE_CONNECTION} "mkdir -p ${REMOTE_DIR}"

# 复制配置文件到远程服务器
print_message "开始推送配置文件到远程服务器..."

# 复制 docker-compose.yml
scp ${SCP_OPTS} ./deploy/.env ${REMOTE_CONNECTION}:${REMOTE_DIR}/.env
scp ${SCP_OPTS} ./deploy/docker-compose.yml ${REMOTE_CONNECTION}:${REMOTE_DIR}/docker-compose.yml
if [ $? -ne 0 ]; then
    print_error ".env 和 docker-compose.yml 推送失败"
    exit 1
fi

print_success "配置文件推送完成"

# 在远程服务器上执行部署命令
print_message "开始在远程服务器上部署..."

ssh ${SSH_OPTS} ${REMOTE_CONNECTION} "cd ${REMOTE_DIR} && \
    docker-compose pull && \
    docker-compose down && \
    docker-compose up -d && \
    docker-compose ps"

if [ $? -ne 0 ]; then
    print_error "远程部署失败"
    exit 1
fi

print_success "部署完成！"

# 显示容器日志
print_message "显示容器日志（按 Ctrl+C 退出）..."
ssh ${SSH_OPTS} ${REMOTE_CONNECTION} "cd ${REMOTE_DIR} && docker-compose logs -f" 