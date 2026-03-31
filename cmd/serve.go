package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v3"

	"eino-rag-demo/internal/config"
	"eino-rag-demo/internal/embedding"
	"eino-rag-demo/internal/handler"
	"eino-rag-demo/internal/rag"
	"eino-rag-demo/internal/rerank"
	"eino-rag-demo/internal/vectorstore"
)

var (
	// Version 版本号
	Version = "v0.1.0"
	// BuildTime 构建时间
	BuildTime = "unknown"
)

func main() {
	app := &cli.Command{
		Name:    "eino-rag-demo",
		Version: Version,
		Usage:   "RAG knowledge base demo using eino framework",
		Commands: []*cli.Command{
			{
				Name:   "serve",
				Usage:  "Start the RAG API server",
				Action: serve,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "host",
						Usage:   "Server host",
						Value:   "0.0.0.0",
					},
					&cli.IntFlag{
						Name:    "port",
						Usage:   "Server port",
						Value:   8080,
					},
					&cli.StringFlag{
						Name:    "config",
						Usage:   "Config file path",
						Value:   ".env",
					},
				},
			},
			{
				Name:   "version",
				Usage:  "Show version information",
				Action: showVersion,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func showVersion(ctx context.Context, cmd *cli.Command) error {
	fmt.Printf("eino-rag-demo %s (built: %s)\n", Version, BuildTime)
	return nil
}

func serve(ctx context.Context, cmd *cli.Command) error {
	// 加载配置
	cfg := config.Load()

	// 获取命令行参数（覆盖配置文件）
	host := cmd.String("host")
	port := cmd.Int("port")
	if port == 0 {
		port = cfg.ServerPort
	}
	if host == "" {
		host = cfg.ServerHost
	}

	// 初始化 embedding (使用 eino-ext openai)
	embedder, err := embedding.NewEmbedder(ctx, &embedding.Config{
		APIKey:     cfg.EmbeddingAPIKey,
		Model:      cfg.EmbeddingModel,
		BaseURL:    cfg.EmbeddingBaseURL,
		Dimensions: cfg.EmbeddingDimensions,
		Timeout:    cfg.EmbeddingTimeout,
	})
	if err != nil {
		return fmt.Errorf("failed to create embedder: %w", err)
	}

	// 初始化 reranker (使用 eino-ext score reranker)
	reranker, err := rerank.NewReranker(ctx, &rerank.Config{})
	if err != nil {
		return fmt.Errorf("failed to create reranker: %w", err)
	}

	// 初始化向量存储 (使用 eino-ext qdrant 或内存模式)
	vectorStore, err := vectorstore.NewVectorStore(ctx, &vectorstore.Config{
		Embedding: embedder,
		VectorDim: cfg.EmbeddingDimensions,
		// Qdrant 配置 (如果未设置则使用内存模式)
		QdrantHost:     cfg.QdrantHost,
		QdrantGRPCPort: cfg.QdrantGRPCPort,
		QdrantAPIKey:   cfg.QdrantAPIKey,
		Collection:     cfg.QdrantCollection,
	})
	if err != nil {
		return fmt.Errorf("failed to create vector store: %w", err)
	}

	// 创建 RAG 服务
	ragService := rag.NewRAGService(cfg, embedder, reranker, vectorStore)

	// 创建 HTTP 处理器
	h := handler.NewHandler(ragService)

	// 设置 Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(loggingMiddleware())

	// 注册路由
	h.RegisterRoutes(r)

	// 主页
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "eino-rag-demo",
			"version": Version,
			"status":  "running",
		})
	})

	// 创建服务器
	addr := fmt.Sprintf("%s:%d", host, port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 启动服务器
	go func() {
		log.Printf("Starting eino-rag-demo server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
	return nil
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		log.Printf("[%s] %s %s %d %v", method, path, clientIP, status, latency)
	}
}