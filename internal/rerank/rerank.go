/*
 * RAG Rerank 组件
 *
 * 使用 eino-ext/components/document/transformer/reranker/score 实现
 * 基于文档 score 进行重排序，优化 LLM 上下文处理
 */

package rerank

import (
	"context"

	"github.com/cloudwego/eino/components/document"
	score "github.com/cloudwego/eino-ext/components/document/transformer/reranker/score"
)

// Config reranker 配置
type Config struct {
	// ScoreFieldKey 指定 metadata 中存储 score 的 key
	// 如果为 nil，则使用文档的 Score() 方法
	ScoreFieldKey *string
}

// NewReranker 创建 reranker
// 直接使用 eino-ext/components/document/transformer/reranker/score 实现
//
// reranker 会基于文档 score 进行重排序：
// - 高分文档放在数组开头和结尾
// - 低分文档放在中间
// 这基于 "primacy and recency effect" 研究
func NewReranker(ctx context.Context, cfg *Config) (document.Transformer, error) {
	var scoreConfig *score.Config
	if cfg != nil && cfg.ScoreFieldKey != nil {
		scoreConfig = &score.Config{
			ScoreFieldKey: cfg.ScoreFieldKey,
		}
	}

	return score.NewReranker(ctx, scoreConfig)
}