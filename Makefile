.PHONY: help build run clean test install dev fmt lint all

# 默认目标
.DEFAULT_GOAL := help

APP_NAME = envkit
VERSION = 0.1.0
BUILD_DIR = dist
MAIN_FILE = cmd/envkit/main.go

# 帮助信息
help: ## 显示帮助信息
	@echo "EnvKit Makefile 命令:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""

# 构建
build: ## 构建当前平台二进制文件
	@echo "🔨 构建 $(APP_NAME)..."
	@go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(APP_NAME) $(MAIN_FILE)
	@echo "✅ 构建完成: ./$(APP_NAME)"

build-all: ## 构建所有平台二进制文件
	@./build.sh

# 运行
run: ## 运行程序
	@go run $(MAIN_FILE) $(ARGS)

dev: ## 开发模式运行（带自动重载）
	@echo "开发模式运行..."
	@go run $(MAIN_FILE)

# 测试
test: ## 运行测试
	@echo "🧪 运行测试..."
	@go test -v ./...

test-cover: ## 运行测试并生成覆盖率报告
	@echo "🧪 运行测试（带覆盖率）..."
	@go test -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆盖率报告已生成: coverage.html"

# 代码质量
fmt: ## 格式化代码
	@echo "🎨 格式化代码..."
	@go fmt ./...
	@echo "✅ 代码格式化完成"

lint: ## 代码检查
	@echo "🔍 代码检查..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint 未安装，使用 go vet"; \
		go vet ./...; \
	fi

# 清理
clean: ## 清理构建文件
	@echo "🧹 清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(APP_NAME)
	@rm -f coverage.out coverage.html
	@echo "✅ 清理完成"

# 安装
install: build ## 安装到系统
	@echo "📦 安装 $(APP_NAME)..."
	@sudo cp $(APP_NAME) /usr/local/bin/
	@echo "✅ 安装完成: /usr/local/bin/$(APP_NAME)"

uninstall: ## 卸载
	@echo "🗑️  卸载 $(APP_NAME)..."
	@sudo rm -f /usr/local/bin/$(APP_NAME)
	@echo "✅ 卸载完成"

# 依赖
deps: ## 安装依赖
	@echo "📥 安装依赖..."
	@go mod download
	@go mod tidy
	@echo "✅ 依赖安装完成"

# 更新依赖
update-deps: ## 更新依赖
	@echo "🔄 更新依赖..."
	@go get -u ./...
	@go mod tidy
	@echo "✅ 依赖更新完成"

# 检查
check: fmt lint test ## 运行所有检查（格式化、检查、测试）

# 发布准备
release-check: clean check build-all ## 发布前检查
	@echo "✅ 发布前检查完成"

# 开发辅助
init: ## 初始化开发环境
	@echo "🚀 初始化开发环境..."
	@go mod download
	@echo "✅ 开发环境初始化完成"

# 快速测试
quick-test: ## 快速测试（不带详细输出）
	@go test ./... > /dev/null 2>&1 && echo "✅ 测试通过" || echo "❌ 测试失败"

# 显示版本
version: ## 显示版本信息
	@echo "$(APP_NAME) version $(VERSION)"

# 全部
all: clean check build ## 清理、检查、构建
