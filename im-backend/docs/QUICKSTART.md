# d-im 快速开始指南

## 环境准备

### 1. 安装依赖

```bash
# 安装 Go 1.21+
# 下载地址: https://golang.org/dl/

# 安装 MongoDB 7.0+
# 下载地址: https://www.mongodb.com/try/download/community

# 安装 Redis 7.0+
# 下载地址: https://redis.io/download

# 安装 Docker (可选)
# 下载地址: https://www.docker.com/products/docker-desktop
```

### 2. 克隆项目

```bash
git clone <repository-url>
cd d-im
```

### 3. 安装 Go 依赖

```bash
go mod tidy
```

## 配置设置

### 1. 复制配置文件

```bash
cp config.yaml.example config.yaml
```

### 2. 生成 JWT 密钥

```bash
# 创建密钥目录
mkdir -p config/keys

# 生成私钥
openssl genrsa -out config/keys/private.pem 2048

# 生成公钥
openssl rsa -in config/keys/private.pem -pubout -out config/keys/public.pem
```

### 3. 配置环境变量

```bash
# 设置环境变量
export API_SERVER_PORT=8080
export WS_SERVER_PORT=9000
export MONGODB_URI="mongodb://localhost:27017/go_im"
export REDIS_ADDR="localhost:6379"
export REDIS_PASSWORD=""
export REDIS_DB=0
export JWT_PUBLIC_KEY_PATH="./config/keys/public.pem"
```

## 启动服务

### 方式一：分离部署（推荐）

#### 1. 启动 API 服务器

```bash
# 使用 Makefile
make run-api

# 或者直接运行
go run cmd/api-server/main.go
```

#### 2. 启动 WebSocket 服务器

```bash
# 新开一个终端
make run-ws

# 或者直接运行
go run cmd/ws-server/main.go
```

#### 3. 验证服务

```bash
# 检查 API 服务器
curl http://localhost:8080/health

# 检查 WebSocket 服务器
curl http://localhost:9000/health
```

### 方式二：统一部署

```bash
# 启动统一服务器（API + WebSocket）
make run-unified

# 或者直接运行
go run cmd/server/main.go
```

### 方式三：生产部署

生产环境使用仓库根目录 `deploy/` 下的 Portainer Stack 模板，详见 [`deploy/README.md`](../../deploy/README.md)。

## 测试连接

### 1. 测试 API 接口

```bash
# 健康检查
curl http://localhost:8080/health

# 获取消息列表（需要 JWT Token）
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8080/api/im/messages
```

### 2. 测试 WebSocket 连接

```javascript
// 使用 JavaScript 测试 WebSocket 连接
const ws = new WebSocket('ws://localhost:9000/im/ws?token=YOUR_JWT_TOKEN');

ws.onopen = function() {
    console.log('WebSocket 连接已建立');
    
    // 发送消息
    ws.send(JSON.stringify({
        receiver_id: "target_user_id",
        type: "text",
        content: "Hello, World!"
    }));
};

ws.onmessage = function(event) {
    console.log('收到消息:', event.data);
};

ws.onclose = function() {
    console.log('WebSocket 连接已关闭');
};

ws.onerror = function(error) {
    console.error('WebSocket 错误:', error);
};
```

## 开发工具

### 1. 使用 Makefile

```bash
# 查看所有可用命令
make help

# 构建所有服务
make build

# 运行测试
make test

# 格式化代码
make fmt

# 代码检查
make lint
```

### 2. 使用构建脚本

```bash
# 构建所有服务
./scripts/build.sh

# 查看构建产物
ls bin/
```

### 3. 使用启动脚本

```bash
# 启动 API 服务器
./scripts/start-api.sh

# 启动 WebSocket 服务器
./scripts/start-ws.sh

# 启动统一服务器
./scripts/start-unified.sh
```

## 常见问题

### 1. 端口被占用

```bash
# 查看端口占用
lsof -i :8080
lsof -i :9000

# 杀死进程
kill -9 <PID>
```

### 2. MongoDB 连接失败

```bash
# 检查 MongoDB 服务状态
brew services list | grep mongodb

# 启动 MongoDB 服务
brew services start mongodb-community
```

### 3. Redis 连接失败

```bash
# 检查 Redis 服务状态
brew services list | grep redis

# 启动 Redis 服务
brew services start redis
```

### 4. JWT 密钥问题

```bash
# 检查密钥文件权限
ls -la config/keys/

# 设置正确的权限
chmod 600 config/keys/private.pem
chmod 644 config/keys/public.pem
```

## 监控和调试

### 1. 查看日志

开发环境直接查看终端输出；生产环境通过 Portainer 或容器运行时查看各服务日志。

### 2. 健康检查

```bash
# API 服务器健康检查
curl http://localhost:8080/health

# WebSocket 服务器健康检查
curl http://localhost:9000/health
```

### 3. 性能监控

```bash
# 查看容器资源使用情况
docker stats

# 查看系统资源使用情况
htop
```

## 下一步

1. 阅读 [API 文档](api.md) 了解详细的 API 接口
2. 阅读 [架构文档](ARCHITECTURE.md) 了解系统架构
3. 查看 [SDK 文档](../sdk/) 了解客户端 SDK 使用方法
4. 参与项目开发和贡献

## 获取帮助

- 查看 [README.md](../README.md) 获取更多信息
- 提交 [Issue](https://github.com/your-repo/d-im/issues) 报告问题
- 参与 [Discussion](https://github.com/your-repo/d-im/discussions) 讨论 