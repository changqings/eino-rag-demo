#!/bin/bash

# Eino RAG Demo 测试脚本

set -e

echo "Running eino-rag-demo tests..."

# 运行所有测试
echo "=== Running unit tests ==="
go test -v ./internal/config/...
go test -v ./internal/embedding/...
go test -v ./internal/rerank/...
go test -v ./internal/vectorstore/...
go test -v ./internal/rag/...

echo ""
echo "=== All tests passed ==="
