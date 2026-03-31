package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("EMBEDDING_API_KEY", "test-key")
	os.Setenv("RAG_TOP_K", "10")

	cfg := Load()

	if cfg.ServerPort != 9999 {
		t.Errorf("Expected ServerPort 9999, got %d", cfg.ServerPort)
	}

	if cfg.EmbeddingAPIKey != "test-key" {
		t.Errorf("Expected EmbeddingAPIKey test-key, got %s", cfg.EmbeddingAPIKey)
	}

	if cfg.RAGTopK != 10 {
		t.Errorf("Expected RAGTopK 10, got %d", cfg.RAGTopK)
	}
}

func TestGet(t *testing.T) {
	// 重置配置
	cfg = nil

	cfg = Get()
	if cfg == nil {
		t.Error("Expected non-nil config")
	}

	// 再次调用应该返回相同实例
	cfg2 := Get()
	if cfg != cfg2 {
		t.Error("Expected same config instance")
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		def      int
		expected int
	}{
		{"valid int", "TEST_INT", "123", 0, 123},
		{"invalid int", "TEST_INT2", "abc", 0, 0},
		{"empty env", "TEST_INT3", "", 42, 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
			} else {
				os.Unsetenv(tt.envKey)
			}

			result := getEnvInt(tt.envKey, tt.def)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetEnvFloat(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		def      float64
		expected float64
	}{
		{"valid float", "TEST_FLOAT", "1.5", 0.0, 1.5},
		{"invalid float", "TEST_FLOAT2", "abc", 0.0, 0.0},
		{"empty env", "TEST_FLOAT3", "", 0.7, 0.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
			} else {
				os.Unsetenv(tt.envKey)
			}

			result := getEnvFloat(tt.envKey, tt.def)
			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}
