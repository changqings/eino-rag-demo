package rag

import (
	"context"
	"fmt"
	"sync"

	"eino-rag-demo/internal/config"
	"eino-rag-demo/internal/vectorstore"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
)

// RAGService RAG 服务主类
type RAGService struct {
	cfg         *config.Config
	embedder    embedding.Embedder // eino embedding 接口
	reranker    document.Transformer // 使用 eino document.Transformer 接口
	vectorStore vectorstore.VectorStore
	chunkSize   int
	chunkOverlap int
	mu          sync.RWMutex
}

// NewRAGService 创建 RAG 服务实例
func NewRAGService(
	cfg *config.Config,
	embedder embedding.Embedder,
	reranker document.Transformer,
	vectorStore vectorstore.VectorStore,
) *RAGService {
	return &RAGService{
		cfg:          cfg,
		embedder:     embedder,
		reranker:     reranker,
		vectorStore:  vectorStore,
		chunkSize:    cfg.RAGChunkSize,
		chunkOverlap: cfg.RAGChunkOverlap,
	}
}

// Document 文档结构
type Document struct {
	ID      string         `json:"id"`
	Content string         `json:"content"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// QueryResult 查询结果
type QueryResult struct {
	Documents []Document `json:"documents"`
	Scores    []float64  `json:"scores"`
}

// Query 输入查询，返回相关文档
func (r *RAGService) Query(ctx context.Context, query string) (*QueryResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. 从向量库检索相关文档
	docs, err := r.vectorStore.Retrieve(ctx, query, r.cfg.RAGTopK)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve from vector store: %w", err)
	}

	if len(docs) == 0 {
		return &QueryResult{
			Documents: []Document{},
			Scores:    []float64{},
		}, nil
	}

	// 2. 使用 Reranker 进行重排序
	rerankedDocs, err := r.reranker.Transform(ctx, docs)
	if err != nil {
		// 如果 rerank 失败，返回原始检索结果
		result := &QueryResult{
			Documents: make([]Document, len(docs)),
			Scores:    make([]float64, len(docs)),
		}
		for i, doc := range docs {
			result.Documents[i] = Document{
				ID:      doc.ID,
				Content: doc.Content,
				Meta:    doc.MetaData,
			}
		}
		return result, nil
	}

	// 3. 构建最终结果
	result := &QueryResult{
		Documents: make([]Document, len(rerankedDocs)),
		Scores:    make([]float64, len(rerankedDocs)),
	}
	for i, doc := range rerankedDocs {
		result.Documents[i] = Document{
			ID:      doc.ID,
			Content: doc.Content,
			Meta:    doc.MetaData,
		}
		result.Scores[i] = doc.Score()
	}

	return result, nil
}

// IndexDocument 将文档添加到索引
func (r *RAGService) IndexDocument(ctx context.Context, doc *Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. 将文档分块
	chunks := r.chunkText(doc.Content)

	// 2. 转换为 schema.Document 格式
	einoDocs := make([]*schema.Document, len(chunks))
	for i, chunk := range chunks {
		einoDocs[i] = &schema.Document{
			ID:       fmt.Sprintf("%s_chunk_%d", doc.ID, i),
			Content:  chunk,
			MetaData: map[string]any{
				"doc_id": doc.ID,
				"chunk":  i,
			},
		}
	}

	// 3. 存储到向量库
	_, err := r.vectorStore.Store(ctx, einoDocs)
	return err
}

// IndexDocuments 批量索引文档
func (r *RAGService) IndexDocuments(ctx context.Context, docs []*Document) error {
	for _, doc := range docs {
		if err := r.IndexDocument(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

// DeleteDocument 从索引中删除文档 (not implemented for in-memory store)
func (r *RAGService) DeleteDocument(ctx context.Context, docID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return fmt.Errorf("delete not implemented for in-memory store")
}

// chunkText 将文本分块
func (r *RAGService) chunkText(text string) []string {
	if r.chunkSize <= 0 {
		r.chunkSize = 512
	}

	var chunks []string
	runes := []rune(text)
	length := len(runes)

	if length == 0 {
		return chunks
	}

	if length <= r.chunkSize {
		return []string{text}
	}

	start := 0
	for start < length {
		end := start + r.chunkSize
		if end > length {
			end = length
		}

		// 尝试在单词边界处分割
		if end < length {
			for end > start && runes[end] != ' ' && runes[end] != ',' && runes[end] != '.' && runes[end] != '\n' {
				end--
			}
			if end == start {
				end = start + r.chunkSize
				if end > length {
					end = length
				}
			}
		}

		chunks = append(chunks, string(runes[start:end]))
		start = end + r.chunkOverlap
		if start >= length {
			break
		}
		// 确保下一个块从上一个块的结尾之后开始
		if start <= end {
			start = end + 1
		}
	}

	return chunks
}