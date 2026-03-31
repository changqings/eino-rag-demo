/*
 * RAG Vector Store 组件
 *
 * 使用 eino-ext/components/indexer/qdrant 实现文档存储
 * 使用 eino-ext/components/retriever/qdrant 实现文档检索
 *
 * 支持两种模式：
 * 1. Qdrant 模式：使用 Qdrant 向量数据库（需要配置 QDRANT_HOST 等）
 * 2. 内存模式：当未配置 Qdrant 时使用内存存储（适合开发测试）
 */

package vectorstore

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	qdrantIndexer "github.com/cloudwego/eino-ext/components/indexer/qdrant"
	qdrantRetriever "github.com/cloudwego/eino-ext/components/retriever/qdrant"
	qdrant "github.com/qdrant/go-client/qdrant"
)

// Config 向量存储配置
type Config struct {
	// Qdrant 配置
	QdrantHost     string // Qdrant 服务器地址，如 localhost:6334
	QdrantGRPCPort int    // Qdrant gRPC 端口，默认 6334
	QdrantAPIKey   string // Qdrant API Key（可选）
	Collection     string // 集合名称，默认 "eino-rag"

	// Embedding 配置
	Embedding embedding.Embedder // 必需的 embedding 组件
	VectorDim int                // 向量维度，默认 1536 (OpenAI)

	// 内存模式配置
	UseMemory bool // 强制使用内存模式
}

// VectorStore 向量存储接口
type VectorStore interface {
	// Store 存储文档
	Store(ctx context.Context, docs []*schema.Document) ([]string, error)

	// Retrieve 检索文档
	Retrieve(ctx context.Context, query string, topK int) ([]*schema.Document, error)

	// Close 关闭连接
	Close(ctx context.Context) error
}

// qdrantStore Qdrant 实现
type qdrantStore struct {
	indexer    *qdrantIndexer.Indexer
	retriever  *qdrantRetriever.Retriever
	indexerCli *qdrant.Client
}

// NewVectorStore 创建向量存储
// 使用 eino-ext qdrant indexer/retriever 实现
func NewVectorStore(ctx context.Context, cfg *Config) (VectorStore, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if cfg.Embedding == nil {
		return nil, fmt.Errorf("embedding is required")
	}

	// 内存模式
	if cfg.UseMemory || cfg.QdrantHost == "" {
		return newMemoryStore(ctx, cfg)
	}

	// Qdrant 模式
	return newQdrantStore(ctx, cfg)
}

// newQdrantStore 创建 Qdrant 存储
func newQdrantStore(ctx context.Context, cfg *Config) (*qdrantStore, error) {
	host := cfg.QdrantHost
	if host == "" {
		host = "localhost"
	}

	grpcPort := int(cfg.QdrantGRPCPort)
	if grpcPort == 0 {
		grpcPort = 6334
	}

	// 创建 Qdrant 客户端
	cli, err := qdrant.NewClient(&qdrant.Config{
		Host:   host,
		Port:   grpcPort,
		APIKey: cfg.QdrantAPIKey,
		UseTLS: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	collection := cfg.Collection
	if collection == "" {
		collection = "eino-rag"
	}

	vectorDim := cfg.VectorDim
	if vectorDim == 0 {
		vectorDim = 1536
	}

	// 创建 indexer
	indexer, err := qdrantIndexer.NewIndexer(ctx, &qdrantIndexer.Config{
		Client:     cli,
		Collection: collection,
		VectorDim:  vectorDim,
		Distance:   qdrant.Distance_Cosine,
		Embedding:  cfg.Embedding,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant indexer: %w", err)
	}

	// 创建 retriever
	retriever, err := qdrantRetriever.NewRetriever(ctx, &qdrantRetriever.Config{
		Client:     cli,
		Collection: collection,
		Embedding:  cfg.Embedding,
		TopK:       5,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant retriever: %w", err)
	}

	return &qdrantStore{
		indexer:    indexer,
		retriever:  retriever,
		indexerCli: cli,
	}, nil
}

// Store 存储文档
func (s *qdrantStore) Store(ctx context.Context, docs []*schema.Document) ([]string, error) {
	return s.indexer.Store(ctx, docs)
}

// Retrieve 检索文档
func (s *qdrantStore) Retrieve(ctx context.Context, query string, topK int) ([]*schema.Document, error) {
	opts := []retriever.Option{
		retriever.WithTopK(topK),
	}
	return s.retriever.Retrieve(ctx, query, opts...)
}

// Close 关闭连接
func (s *qdrantStore) Close(ctx context.Context) error {
	// Qdrant Go client 不需要显式关闭
	return nil
}

// ==================== 内存模式实现 ====================

// memoryStore 内存存储实现
type memoryStore struct {
	docs      map[string]*schema.Document
	vectors   map[string][]float64
	embedding embedding.Embedder
	mu        sync.RWMutex
}

// newMemoryStore 创建内存存储
func newMemoryStore(ctx context.Context, cfg *Config) (*memoryStore, error) {
	return &memoryStore{
		docs:      make(map[string]*schema.Document),
		vectors:   make(map[string][]float64),
		embedding: cfg.Embedding,
	}, nil
}

// Store 存储文档
func (s *memoryStore) Store(ctx context.Context, docs []*schema.Document) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ids := make([]string, 0, len(docs))
	texts := make([]string, 0, len(docs))

	for _, doc := range docs {
		ids = append(ids, doc.ID)
		s.docs[doc.ID] = doc
		texts = append(texts, doc.Content)
	}

	// 生成向量
	vectors, err := s.embedding.EmbedStrings(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("failed to embed documents: %w", err)
	}

	for i, id := range ids {
		if i < len(vectors) {
			s.vectors[id] = vectors[i]
		}
	}

	return ids, nil
}

// Retrieve 检索文档
func (s *memoryStore) Retrieve(ctx context.Context, query string, topK int) ([]*schema.Document, error) {
	// 生成查询向量
	queryVectors, err := s.embedding.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}
	if len(queryVectors) == 0 {
		return nil, fmt.Errorf("no query vector returned")
	}

	queryVec := queryVectors[0]

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 计算相似度并排序
	type scoredDoc struct {
		doc   *schema.Document
		score float64
	}

	scored := make([]scoredDoc, 0, len(s.docs))
	for _, doc := range s.docs {
		if vec, ok := s.vectors[doc.ID]; ok {
			score := cosineSimilarity(queryVec, vec)
			docCopy := *doc
			docCopy.WithScore(score)
			scored = append(scored, scoredDoc{doc: &docCopy, score: score})
		}
	}

	// 按分数排序
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// 取前 topK
	result := make([]*schema.Document, 0, topK)
	for i := 0; i < topK && i < len(scored); i++ {
		result = append(result, scored[i].doc)
	}

	return result, nil
}

// Close 关闭
func (s *memoryStore) Close(ctx context.Context) error {
	return nil
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}