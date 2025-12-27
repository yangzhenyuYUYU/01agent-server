.PHONY: build run test clean docker-up docker-down help

# 应用名称
APP_NAME := gin_web
# 构建版本
VERSION := $(shell git describe --tags --always --dirty)
# 构建时间
BUILD_TIME := $(shell date '+%Y-%m-%d_%H:%M:%S')
# Git提交号
GIT_COMMIT := $(shell git rev-parse HEAD)

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 默认目标
all: build

# 构建应用
build:
	@echo "Building $(APP_NAME)..."
	@go build $(LDFLAGS) -o bin/$(APP_NAME) .
	@echo "Build completed: bin/$(APP_NAME)"

# 运行应用
run:
	@echo "Running $(APP_NAME)..."
	@go run main.go

# 运行测试
test:
	@echo "Running tests..."
	@go test -v ./...

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 代码格式化
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# 代码检查
lint:
	@echo "Running linter..."
	@golangci-lint run

# 下载依赖
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# 清理构建文件
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean

# 启动Docker开发环境
docker-up:
	@echo "Starting Docker development environment..."
	@docker-compose up -d
	@echo "Docker services started"

# 停止Docker开发环境
docker-down:
	@echo "Stopping Docker development environment..."
	@docker-compose down
	@echo "Docker services stopped"

# 查看Docker服务状态
docker-status:
	@echo "Docker services status:"
	@docker-compose ps

# 查看Docker日志
docker-logs:
	@echo "Docker services logs:"
	@docker-compose logs -f

# 重启Docker服务
docker-restart:
	@echo "Restarting Docker services..."
	@docker-compose restart
	@echo "Docker services restarted"

# 生成API文档
docs:
	@echo "Generating API documentation..."
	@swag init -g main.go

# 安装开发工具
dev-setup:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Development tools installed"

# 数据库迁移
migrate:
	@echo "Running database migration..."
	@go run main.go migrate

# 创建发布包
release: clean build
	@echo "Creating release package..."
	@mkdir -p release
	@cp bin/$(APP_NAME) release/
	@cp -r configs release/
	@cp README.md release/
	@tar -czf release/$(APP_NAME)-$(VERSION).tar.gz -C release .
	@echo "Release package created: release/$(APP_NAME)-$(VERSION).tar.gz"

# 显示帮助信息
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  clean         - Clean build files"
	@echo "  docker-up     - Start Docker development environment"
	@echo "  docker-down   - Stop Docker development environment"
	@echo "  docker-status - Show Docker services status"
	@echo "  docker-logs   - Show Docker services logs"
	@echo "  docker-restart- Restart Docker services"
	@echo "  docs          - Generate API documentation"
	@echo "  dev-setup     - Install development tools"
	@echo "  migrate       - Run database migration"
	@echo "  release       - Create release package"
	@echo "  help          - Show this help message" 