package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"eino-rag-demo/internal/config"
)

// LLMClient LLM 客户端（用于生成 RAG 答案）
type LLMClient struct {
	config *config.Config
	client *http.Client
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message       ChatMessage `json:"message"`
		FinishReason  string      `json:"finish_reason"`
		Index         int         `json:"index"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewLLMClient 创建 LLM 客户端
func NewLLMClient(cfg *config.Config) *LLMClient {
	return &LLMClient{
		config: cfg,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Chat 使用 LLM 生成回复
func (l *LLMClient) Chat(ctx context.Context, systemPrompt string, userMessage string) (string, error) {
	reqBody := ChatRequest{
		Model: l.config.LLMModel,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		MaxTokens:   l.config.LLMMaxTokens,
		Temperature: l.config.LLMTemperature,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", l.config.LLMBaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", l.config.LLMAPIKey))

	resp, err := l.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return result.Choices[0].Message.Content, nil
}

// GenerateRAGAnswer 生成 RAG 答案
func (l *LLMClient) GenerateRAGAnswer(ctx context.Context, query string, contextDocs []string) (string, error) {
	// 构建上下文
	var contextBuilder string
	for i, doc := range contextDocs {
		contextBuilder += fmt.Sprintf("\n[%d] %s\n", i+1, doc)
	}

	systemPrompt := `You are a helpful assistant that answers questions based on the provided context.
When answering, you should:
1. Only use the information from the provided context
2. Cite the sources of your information using the [N] notation
3. If the context doesn't contain relevant information, say so
4. Be clear and concise in your answer`

	userMessage := fmt.Sprintf("Context:\n%s\n\nQuestion: %s\n\nPlease provide an answer based on the context above.", contextBuilder, query)

	return l.Chat(ctx, systemPrompt, userMessage)
}
