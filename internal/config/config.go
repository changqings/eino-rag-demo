package config

import (
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

// Config 应用配置
type Config struct {
	// Server
	ServerHost string
	ServerPort int

	// Embedding
	EmbeddingAPIKey     string
	EmbeddingModel      string
	EmbeddingBaseURL    string
	EmbeddingDimensions int
	EmbeddingBatchSize  int
	EmbeddingTimeout    int // 超时时间（秒）

	// Rerank
	RerankAPIKey  string
	RerankModel   string
	RerankBaseURL string
	RerankTopN    int

	// Vector Store (Qdrant)
	VectorStoreURL         string
	VectorStoreAPIKey      string
	VectorStoreCollection  string
	VectorStoreVectorSize   int
	VectorStoreDistance     string

	// Qdrant 专用配置
	QdrantHost      string // Qdrant 服务器地址
	QdrantGRPCPort  int    // Qdrant gRPC 端口
	QdrantAPIKey    string // Qdrant API Key
	QdrantCollection string // Qdrant 集合名称

	// LLM
	LLMAPIKey      string
	LLMModel       string
	LLMBaseURL     string
	LLMMaxTokens   int
	LLMTemperature float64

	// RAG
	RAGTopK         int
	RAGRerankTopK   int
	RAGChunkSize    int
	RAGChunkOverlap int
}

var (
	cfg  *Config
	once sync.Once
)

// Load 加载配置
func Load() *Config {
	once.Do(func() {
		// 尝试加载 .env 文件
		_ = godotenv.Load()

		cfg = &Config{
			// Server
			ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
			ServerPort: getEnvInt("SERVER_PORT", 8080),

			// Embedding
			EmbeddingAPIKey:     getEnv("EMBEDDING_API_KEY", ""),
			EmbeddingModel:      getEnv("EMBEDDING_MODEL", "text-embedding-3-small"),
			EmbeddingBaseURL:    getEnv("EMBEDDING_BASE_URL", "https://api.openai.com/v1"),
			EmbeddingDimensions: getEnvInt("EMBEDDING_DIMENSIONS", 1536),
			EmbeddingBatchSize:  getEnvInt("EMBEDDING_BATCH_SIZE", 100),
			EmbeddingTimeout:    getEnvInt("EMBEDDING_TIMEOUT", 30),

			// Rerank
			RerankAPIKey:  getEnv("RERANK_API_KEY", ""),
			RerankModel:   getEnv("RERANK_MODEL", "cohere-rerank"),
			RerankBaseURL: getEnv("RERANK_BASE_URL", "https://api.cohere.ai/v1"),
			RerankTopN:    getEnvInt("RERANK_TOP_N", 10),

			// Vector Store
			VectorStoreURL:        getEnv("VECTOR_STORE_URL", "http://localhost:6334"),
			VectorStoreAPIKey:     getEnv("VECTOR_STORE_API_KEY", ""),
			VectorStoreCollection: getEnv("VECTOR_STORE_COLLECTION", "knowledge_base"),
			VectorStoreVectorSize: getEnvInt("VECTOR_STORE_VECTOR_SIZE", 1536),
			VectorStoreDistance:   getEnv("VECTOR_STORE_DISTANCE", "Cosine"),

			// Qdrant 专用配置
			QdrantHost:      getEnv("QDRANT_HOST", ""),
			QdrantGRPCPort:  getEnvInt("QDRANT_GRPC_PORT", 6334),
			QdrantAPIKey:    getEnv("QDRANT_API_KEY", ""),
			QdrantCollection: getEnv("QDRANT_COLLECTION", "eino-rag"),

			// LLM
			LLMAPIKey:      getEnv("LLM_API_KEY", ""),
			LLMModel:       getEnv("LLM_MODEL", "gpt-4o-mini"),
			LLMBaseURL:     getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
			LLMMaxTokens:   getEnvInt("LLM_MAX_TOKENS", 2048),
			LLMTemperature: getEnvFloat("LLM_TEMPERATURE", 0.7),

			// RAG
			RAGTopK:         getEnvInt("RAG_TOP_K", 5),
			RAGRerankTopK:   getEnvInt("RAG_RERANK_TOP_K", 3),
			RAGChunkSize:    getEnvInt("RAG_CHUNK_SIZE", 512),
			RAGChunkOverlap: getEnvInt("RAG_CHUNK_OVERLAP", 50),
		}
	})
	return cfg
}

// Get 获取配置实例
func Get() *Config {
	if cfg == nil {
		return Load()
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}
