# Eino RAG Demo

基于 [CloudWeGo Eino](https://github.com/cloudwego/eino) 框架实现的 RAG（Retrieval-Augmented Generation）知识库演示项目。

## 功能特性

- 📝 **文档向量化**: 使用 OpenAI-compatible Embedding API 将文本转换为向量
- 🔍 **向量检索**: 支持 Qdrant 向量数据库进行高效相似度搜索
- ✨ **重排序**: 使用 Rerank 模型优化检索结果相关性
- 🌐 **RESTful API**: 使用 Gin 框架提供简洁的 HTTP 接口
- ⚙️ **配置管理**: 支持 dotenv 配置和 urfave-cli 命令行参数

## 项目结构

```
eino-rag-demo/
├── cmd/
│   └── serve.go           # 命令行入口
├── internal/
│   ├── config/            # 配置管理
│   ├── embedding/         # Embedding 组件
│   ├── rerank/            # Rerank 组件
│   ├── vectorstore/       # 向量存储
│   ├── rag/               # RAG 核心逻辑
│   └── handler/           # HTTP 处理器
├── pkg/
│   └── models/            # 数据模型
├── .env.example           # 配置示例
└── README.md
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置环境变量

复制 `.env.example` 为 `.env` 并配置：

```bash
cp .env.example .env
```

编辑 `.env` 文件，填入你的 API Keys：

```env
# Server
SERVER_PORT=8080

# Embedding (OpenAI)
EMBEDDING_API_KEY=your-openai-api-key
EMBEDDING_MODEL=text-embedding-3-small
EMBEDDING_BASE_URL=https://api.openai.com/v1

# Rerank (Cohere)
RERANK_API_KEY=your-cohere-api-key
RERANK_MODEL=cohere-rerank-3.5
RERANK_BASE_URL=https://api.cohere.ai/v1

# Vector Store (Qdrant)
VECTOR_STORE_URL=http://localhost:6334
VECTOR_STORE_COLLECTION=knowledge_base
```

### 3. 启动 Qdrant（可选）

如果你使用 Qdrant 作为向量存储：

```bash
docker run -d -p 6333:6333 -p 6334:6334 \
    -v $(pwd)/qdrant_storage:/qdrant/storage \
    qdrant/qdrant
```

### 4. 运行服务

```bash
# 直接运行
go run cmd/serve.go serve

# 或使用 CLI 参数
go run cmd/serve.go serve --port 8080 --host 0.0.0.0
```

### 5. API 使用

#### 健康检查

```bash
curl http://localhost:8080/api/v1/health
```

#### 添加文档

```bash
curl -X POST http://localhost:8080/api/v1/documents \
  -H "Content-Type: application/json" \
  -d '{
    "id": "doc1",
    "content": "Go语言是一门由Google开发的编译型编程语言，具有高性能、高并发、简洁语法等特点。",
    "meta": {"source": "wiki"}
  }'
```

#### 批量添加文档

```bash
curl -X POST http://localhost:8080/api/v1/documents/batch \
  -H "Content-Type: application/json" \
  -d '{
    "docs": [
      {"id": "doc2", "content": "Rust是一门注重安全性、并发性和实用性的系统编程语言。"},
      {"id": "doc3", "content": "Python是一种广泛使用的高级编程语言，以其简洁易读的语法著称。"}
    ]
  }'
```

#### 查询

```bash
curl -X POST http://localhost:8080/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "Go语言有什么特点？"
  }'
```

#### 删除文档

```bash
curl -X DELETE http://localhost:8080/api/v1/documents/doc1
```

## API 文档

### POST /api/v1/query

RAG 查询接口

**请求体：**
```json
{
  "query": "你的问题"
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "documents": [
      {
        "id": "doc1_chunk_0",
        "content": "Go语言是一门由Google开发的编译型编程语言...",
        "meta": {"doc_id": "doc1", "chunk": 0}
      }
    ],
    "scores": [0.95]
  }
}
```

### POST /api/v1/documents

添加单个文档到知识库

**请求体：**
```json
{
  "id": "文档唯一ID",
  "content": "文档内容",
  "meta": {"source": "来源"}
}
```

### POST /api/v1/documents/batch

批量添加文档

**请求体：**
```json
{
  "docs": [
    {"id": "id1", "content": "内容1"},
    {"id": "id2", "content": "内容2"}
  ]
}
```

### DELETE /api/v1/documents/:id

删除文档

## 配置说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| SERVER_HOST | 服务监听地址 | 0.0.0.0 |
| SERVER_PORT | 服务监听端口 | 8080 |
| EMBEDDING_API_KEY | Embedding API Key | - |
| EMBEDDING_MODEL | Embedding 模型 | text-embedding-3-small |
| EMBEDDING_BASE_URL | Embedding API 地址 | https://api.openai.com/v1 |
| RERANK_API_KEY | Rerank API Key | - |
| RERANK_MODEL | Rerank 模型 | cohere-rerank-3.5 |
| VECTOR_STORE_URL | Qdrant 服务地址 | http://localhost:6334 |
| VECTOR_STORE_COLLECTION | Collection 名称 | knowledge_base |
| LLM_API_KEY | LLM API Key | - |
| LLM_MODEL | LLM 模型 | gpt-4o-mini |

## 技术栈

- **框架**: [CloudWeGo Eino](https://github.com/cloudwego/eino)
- **HTTP**: [Gin](https://github.com/gin-gonic/gin)
- **CLI**: [urfave/cli](https://github.com/urfave/cli)
- **向量数据库**: [Qdrant](https://qdrant.tech/)
- **配置**: [godotenv](https://github.com/joho/godotenv)

## License

MIT
# eino-rag-demo
