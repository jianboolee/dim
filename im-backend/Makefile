# d-im Makefile

APP_NAME := d-im
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR := bin
DOCKER_REGISTRY := registry.cn-beijing.aliyuncs.com
DOCKER_NAMESPACE := xrv
DOCKER_IMAGE := api-im
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

REQUIRE_AIR = @which air >/dev/null || (echo "请先安装 air: go install github.com/cosmtrek/air@latest" && exit 1)

.PHONY: help
help: ## 显示帮助信息
	@echo "d-im 构建和部署工具"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: init
init: ## 初始化项目
	@./scripts/init.sh

.PHONY: migrate
migrate: ## 迁移数据库
	@./scripts/init.sh --migrate

.PHONY: dev
dev: ## 开发模式运行（热重载）
	$(REQUIRE_AIR)
	@air

.PHONY: dev-api
dev-api: ## 开发模式运行 API 服务器
	$(REQUIRE_AIR)
	@air -c .air.api.toml

.PHONY: dev-ws
dev-ws: ## 开发模式运行 WebSocket 服务器
	$(REQUIRE_AIR)
	@air -c .air.ws.toml

.PHONY: build
build: ## 构建所有服务
	@./scripts/build.sh

.PHONY: build-api
build-api: ## 仅构建 API 服务器
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/api-server ./cmd/api-server

.PHONY: build-ws
build-ws: ## 仅构建 WebSocket 服务器
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/ws-server ./cmd/ws-server

.PHONY: build-unified
build-unified: ## 仅构建统一服务器
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/unified-server ./cmd/server

.PHONY: run-api
run-api: ## 运行 API 服务器
	@go run ./cmd/api-server

.PHONY: run-ws
run-ws: ## 运行 WebSocket 服务器
	@go run ./cmd/ws-server

.PHONY: run-unified
run-unified: ## 运行统一服务器
	@go run ./cmd/server

.PHONY: test
test: ## 运行测试
	@go test ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: clean
clean: ## 清理构建文件
	@rm -rf $(BUILD_DIR) coverage.out coverage.html
	@go clean -cache

.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	@docker build -t $(APP_NAME):$(VERSION) .

.PHONY: docker-build-api
docker-build-api: ## 构建 API 服务器 Docker 镜像
	@docker build --target api-server-runtime -t $(APP_NAME)-api:$(VERSION) .

.PHONY: docker-build-ws
docker-build-ws: ## 构建 WebSocket 服务器 Docker 镜像
	@docker build --target ws-server-runtime -t $(APP_NAME)-ws:$(VERSION) .

.PHONY: docker-push
docker-push: ## 推送 Docker 镜像到远程仓库
	@docker tag $(APP_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(DOCKER_IMAGE):$(VERSION)
	@docker tag $(APP_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(DOCKER_IMAGE):latest
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(DOCKER_IMAGE):$(VERSION)
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(DOCKER_IMAGE):latest

.PHONY: dev-setup
dev-setup: ## 设置开发环境
	@go mod tidy && go mod download

.PHONY: lint
lint: ## 运行代码检查
	@command -v golangci-lint >/dev/null && golangci-lint run || echo "golangci-lint 未安装，跳过"

.PHONY: fmt
fmt: ## 格式化代码
	@go fmt ./...

.PHONY: deploy
deploy: build docker-build docker-push ## 完整部署流程
