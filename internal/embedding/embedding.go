/*
 * RAG Embedding 组件
 *
 * 使用 eino-ext/components/embedding/openai 实现
 * 支持 OpenAI 和 Azure OpenAI 兼容的 embedding API
 */

package embedding

import (
	"context"
	"time"

	"github.com/cloudwego/eino/components/embedding"
	openai "github.com/cloudwego/eino-ext/components/embedding/openai"
)

// Config embedding 配置
type Config struct {
	APIKey        string // OpenAI API Key
	Model         string // embedding 模型，如 text-embedding-3-small
	BaseURL       string // 可选：自定义 API 地址
	Timeout       int    // 可选：超时时间（秒）
	Dimensions    int    // 可选：向量维度（仅 text-embedding-3+ 支持）
	ByAzure       bool   // 是否使用 Azure OpenAI
	AzureBaseURL  string // Azure OpenAI 端点
	AzureAPIVersion string // Azure OpenAI API 版本
}

// NewEmbedder 创建 embedding embedder
// 直接使用 eino-ext/components/embedding/openai 实现
func NewEmbedder(ctx context.Context, cfg *Config) (embedding.Embedder, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	var timeout int64 = 30
	if cfg.Timeout > 0 {
		timeout = int64(cfg.Timeout)
	}

	config := &openai.EmbeddingConfig{
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Azure OpenAI 配置
	if cfg.ByAzure {
		config.ByAzure = true
		config.BaseURL = cfg.AzureBaseURL
		config.APIVersion = cfg.AzureAPIVersion
	} else if cfg.BaseURL != "" {
		config.BaseURL = cfg.BaseURL
	}

	// 设置维度
	if cfg.Dimensions > 0 {
		config.Dimensions = &cfg.Dimensions
	}

	return openai.NewEmbedder(ctx, config)
}