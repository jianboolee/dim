# d-im Makefile

# 变量定义
APP_NAME := d-im
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR := bin
DOCKER_REGISTRY := registry.cn-beijing.aliyuncs.com
DOCKER_NAMESPACE := xrv
DOCKER_IMAGE := api-im

# 环境变量文件路径
ENV_FILE := .env
ENV_CONFIG_FILE := config/.env

# Go 构建参数
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# 默认目标
.PHONY: help
help: ## 显示帮助信息
	@echo "d-im 构建和部署工具"
	@echo ""
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# 初始化项目
.PHONY: init
init: ## 初始化项目
	@echo "初始化项目..."
	@./scripts/init.sh

# 迁移目标
.PHONY: migrate
migrate: ## 迁移数据库
	@echo "迁移数据库..."
	@./scripts/init.sh --migrate

# 检查环境变量文件
.PHONY: env-check
env-check:
	@echo "🔍 检查环境变量文件..."
	@if [ -f $(ENV_FILE) ]; then \
		echo "✅ 找到根目录 .env 文件"; \
	elif [ -f $(ENV_CONFIG_FILE) ]; then \
		echo "✅ 找到 config/.env 文件"; \
	else \
		echo "⚠️  未找到 .env 文件，将使用默认配置"; \
		echo "💡 提示: 运行 'make init' 创建环境变量文件"; \
	fi


# 开发模式运行（热重载需要安装air）
.PHONY: dev
dev: env-check ## 开发模式运行（热重载）
	@echo "🔄 开发模式运行..."
	@which air > /dev/null || (echo "请先安装air: go install github.com/cosmtrek/air@latest" && exit 1)
	@if [ -f $(ENV_FILE) ]; then \
		echo "📝 加载根目录 .env 文件..."; \
		set -a; source $(ENV_FILE); set +a; air; \
	elif [ -f $(ENV_CONFIG_FILE) ]; then \
		echo "📝 加载 config/.env 文件..."; \
		set -a; source $(ENV_CONFIG_FILE); set +a; air; \
	else \
		echo "📝 使用默认配置运行..."; \
		air; \
	fi

# 开发模式运行 API 服务器
.PHONY: dev-api
dev-api: env-check ## 开发模式运行 API 服务器
	@echo "🔄 开发模式运行 API 服务器..."
	@which air > /dev/null || (echo "请先安装air: go install github.com/cosmtrek/air@latest" && exit 1)
	@if [ -f $(ENV_FILE) ]; then \
		echo "📝 加载根目录 .env 文件..."; \
		set -a; source $(ENV_FILE); set +a; air -c .air.api.toml; \
	elif [ -f $(ENV_CONFIG_FILE) ]; then \
		echo "📝 加载 config/.env 文件..."; \
		set -a; source $(ENV_CONFIG_FILE); set +a; air -c .air.api.toml; \
	else \
		echo "📝 使用默认配置运行..."; \
		air -c .air.api.toml; \
	fi

# 开发模式运行 WebSocket 服务器
.PHONY: dev-ws
dev-ws: env-check ## 开发模式运行 WebSocket 服务器
	@echo "🔄 开发模式运行 WebSocket 服务器..."
	@which air > /dev/null || (echo "请先安装air: go install github.com/cosmtrek/air@latest" && exit 1)
	@if [ -f $(ENV_FILE) ]; then \
		echo "📝 加载根目录 .env 文件..."; \
		set -a; source $(ENV_FILE); set +a; air -c .air.ws.toml; \
	elif [ -f $(ENV_CONFIG_FILE) ]; then \
		echo "📝 加载 config/.env 文件..."; \
		set -a; source $(ENV_CONFIG_FILE); set +a; air -c .air.ws.toml; \
	else \
		echo "📝 使用默认配置运行..."; \
		air -c .air.ws.toml; \
	fi


# 构建目标
.PHONY: build
build: ## 构建所有服务
	@echo "使用构建脚本..."
	@./scripts/build.sh

.PHONY: build-api
build-api: ## 仅构建 API 服务器
	@echo "构建 API 服务器..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/api-server ./cmd/api-server

.PHONY: build-ws
build-ws: ## 仅构建 WebSocket 服务器
	@echo "构建 WebSocket 服务器..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/ws-server ./cmd/ws-server

.PHONY: build-unified
build-unified: ## 仅构建统一服务器
	@echo "构建统一服务器..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/unified-server ./cmd/server

# 运行目标
.PHONY: run-api
run-api: ## 运行 API 服务器
	@echo "启动 API 服务器..."
	@go run ./cmd/api-server/main.go

.PHONY: run-ws
run-ws: ## 运行 WebSocket 服务器
	@echo "启动 WebSocket 服务器..."
	@go run ./cmd/ws-server/main.go

.PHONY: run-unified
run-unified: ## 运行统一服务器
	@echo "启动统一服务器..."
	@go run ./cmd/server/main.go

# 测试目标
.PHONY: test
test: ## 运行测试
	@echo "运行测试..."
	@go test -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率报告..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 清理目标
.PHONY: clean
clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean -cache

# Docker 目标
.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	@docker build -t $(APP_NAME):$(VERSION) .

.PHONY: docker-build-api
docker-build-api: ## 构建 API 服务器 Docker 镜像
	@echo "构建 API 服务器 Docker 镜像..."
	@docker build --target api-server-runtime -t $(APP_NAME)-api:$(VERSION) .

.PHONY: docker-build-ws
docker-build-ws: ## 构建 WebSocket 服务器 Docker 镜像
	@echo "构建 WebSocket 服务器 Docker 镜像..."
	@docker build --target ws-server-runtime -t $(APP_NAME)-ws:$(VERSION) .

.PHONY: docker-push
docker-push: ## 推送 Docker 镜像到远程仓库
	@echo "推送 Docker 镜像到远程仓库..."
	@docker tag $(APP_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(DOCKER_IMAGE):$(VERSION)
	@docker tag $(APP_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(DOCKER_IMAGE):latest
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(DOCKER_IMAGE):$(VERSION)
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(DOCKER_IMAGE):latest

# Docker Compose 目标
.PHONY: docker-compose-up
docker-compose-up: ## 启动 Docker Compose 服务
	@echo "启动 Docker Compose 服务..."
	@docker-compose -f deploy/docker-compose.yml up -d

.PHONY: docker-compose-down
docker-compose-down: ## 停止 Docker Compose 服务
	@echo "停止 Docker Compose 服务..."
	@docker-compose -f deploy/docker-compose.yml down

.PHONY: docker-compose-logs
docker-compose-logs: ## 查看 Docker Compose 日志
	@docker-compose -f deploy/docker-compose.yml logs -f

# 开发目标
.PHONY: dev-setup
dev-setup: ## 设置开发环境
	@echo "设置开发环境..."
	@go mod tidy
	@go mod download
	@echo "开发环境设置完成！"

.PHONY: lint
lint: ## 运行代码检查
	@echo "运行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过代码检查"; \
	fi

.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化代码..."
	@go fmt ./...

# 部署目标
.PHONY: deploy
deploy: build docker-build docker-push ## 完整部署流程
	@echo "部署完成！"

.PHONY: deploy-local
deploy-local: build docker-compose-up ## 本地部署
	@echo "本地部署完成！" 