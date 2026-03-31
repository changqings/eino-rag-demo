package models

// Document 文档模型
type Document struct {
	ID      string         `json:"id"`
	Content string         `json:"content"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// Chunk 文本块
type Chunk struct {
	ID       string `json:"id"`
	DocID    string `json:"doc_id"`
	Content  string `json:"content"`
	Index    int    `json:"index"`
	Vector   []float64 `json:"vector,omitempty"`
}

// Query 查询请求
type Query struct {
	Text   string `json:"text"`
	TopK   int    `json:"top_k,omitempty"`
}

// QueryResult 查询结果
type QueryResult struct {
	Query    string     `json:"query"`
	Chunks   []Chunk    `json:"chunks"`
	Scores   []float64  `json:"scores"`
	Answer   string     `json:"answer,omitempty"`
}
