.PHONY: build run test clean tidy

# 构建
build:
	go build -o bin/eino-rag-demo ./cmd

# 运行
run:
	go run ./cmd serve

# 测试
test:
	go test -v ./...

# 测试覆盖率
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 清理
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# 依赖整理
tidy:
	go mod tidy

# 下载依赖
deps:
	go mod download

# 启动 Qdrant (Docker)
qdrant-start:
	docker run -d -p 6333:6333 -p 6334:6334 \
		-v $(PWD)/qdrant_storage:/qdrant/storage \
		qdrant/qdrant

# 停止 Qdrant
qdrant-stop:
	docker stop $$(docker ps -q --filter ancestor=qdrant/qdrant)

# 开发服务器 (热重载，需要 air)
dev:
	air

# lint
lint:
	golangci-lint run ./...

# 格式化代码
fmt:
	go fmt ./...

# 显示帮助
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  run         - Run the server"
	@echo "  test        - Run tests"
	@echo "  cover       - Run tests with coverage"
	@echo "  clean       - Clean build artifacts"
	@echo "  tidy        - Tidy go modules"
	@echo "  deps        - Download dependencies"
	@echo "  qdrant-start - Start Qdrant container"
	@echo "  qdrant-stop  - Stop Qdrant container"
	@echo "  dev         - Run with hot reload"
	@echo "  lint        - Run linter"
	@echo "  fmt         - Format code"
