package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"eino-rag-demo/internal/rag"
)

// Handler HTTP 处理器
type Handler struct {
	ragService *rag.RAGService
}

// NewHandler 创建 HTTP 处理器
func NewHandler(ragService *rag.RAGService) *Handler {
	return &Handler{
		ragService: ragService,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		// RAG 查询
		api.POST("/query", h.Query)
		
		// 文档管理
		api.POST("/documents", h.IndexDocument)
		api.POST("/documents/batch", h.IndexDocuments)
		api.DELETE("/documents/:id", h.DeleteDocument)
		
		// 健康检查
		api.GET("/health", h.Health)
	}
}

// QueryRequest 查询请求
type QueryRequest struct {
	Query string `json:"query" binding:"required"`
}

// QueryResponse 查询响应
type QueryResponse struct {
	Success bool              `json:"success"`
	Data    *rag.QueryResult `json:"data,omitempty"`
	Error   string           `json:"error,omitempty"`
}

// Query 处理 RAG 查询
func (h *Handler) Query(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, QueryResponse{
			Success: false,
			Error:   "invalid request: " + err.Error(),
		})
		return
	}

	result, err := h.ragService.Query(c.Request.Context(), req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, QueryResponse{
			Success: false,
			Error:   "query failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, QueryResponse{
		Success: true,
		Data:    result,
	})
}

// IndexDocumentRequest 索引文档请求
type IndexDocumentRequest struct {
	ID      string         `json:"id" binding:"required"`
	Content string         `json:"content" binding:"required"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// IndexDocument 处理文档索引
func (h *Handler) IndexDocument(c *gin.Context) {
	var req IndexDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request: " + err.Error(),
		})
		return
	}

	doc := &rag.Document{
		ID:      req.ID,
		Content: req.Content,
		Meta:    req.Meta,
	}

	if err := h.ragService.IndexDocument(c.Request.Context(), doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "index failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "document indexed successfully",
	})
}

// IndexDocumentsRequest 批量索引文档请求
type IndexDocumentsRequest struct {
	Docs []IndexDocumentRequest `json:"docs" binding:"required"`
}

// IndexDocuments 处理批量文档索引
func (h *Handler) IndexDocuments(c *gin.Context) {
	var req IndexDocumentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request: " + err.Error(),
		})
		return
	}

	docs := make([]*rag.Document, len(req.Docs))
	for i, d := range req.Docs {
		docs[i] = &rag.Document{
			ID:      d.ID,
			Content: d.Content,
			Meta:    d.Meta,
		}
	}

	if err := h.ragService.IndexDocuments(c.Request.Context(), docs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "batch index failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   len(docs),
		"message": "documents indexed successfully",
	})
}

// DeleteDocument 处理删除文档
func (h *Handler) DeleteDocument(c *gin.Context) {
	docID := c.Param("id")
	if docID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "document id is required",
		})
		return
	}

	if err := h.ragService.DeleteDocument(c.Request.Context(), docID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "delete failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "document deleted successfully",
	})
}

// Health 健康检查
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "eino-rag-demo",
	})
}
