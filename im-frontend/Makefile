.PHONY: help install dev build preview type-check format clean

help:
	@echo "MP 项目常用命令"
	@echo ""
	@echo "  make install     - 安装依赖"
	@echo "  make dev         - 启动开发服务器"
	@echo "  make build       - 构建生产包"
	@echo "  make preview     - 预览生产构建"
	@echo "  make type-check  - 运行 Vue/TypeScript 类型检查"
	@echo "  make format      - 格式化 src 目录"
	@echo "  make clean       - 清理构建和临时文件"

install:
	@echo "========================================="
	@echo "安装项目依赖"
	@echo "========================================="
	@npm install

dev:
	@echo "========================================="
	@echo "启动开发服务器"
	@echo "========================================="
	@npm run dev

build:
	@echo "========================================="
	@echo "构建生产包"
	@echo "========================================="
	@npm run build

preview:
	@echo "========================================="
	@echo "预览生产构建"
	@echo "========================================="
	@npm run preview

type-check:
	@echo "========================================="
	@echo "运行类型检查"
	@echo "========================================="
	@npm run type-check

format:
	@echo "========================================="
	@echo "格式化 src 目录"
	@echo "========================================="
	@npm run format

clean:
	@echo "========================================="
	@echo "清理构建和临时文件"
	@echo "========================================="
	@rm -rf dist node_modules/.tmp
