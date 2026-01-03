package config

import (
	"os"
	"time"
)

type Config struct {
	// 必须的环境变量
	Email    string // GEMINI_BUSINESS_EMAIL
	ConfigID string // GEMINI_BUSINESS_CONFIG_ID

	// 可选的环境变量
	Port     string // PORT (默认: 8080)
	LogLevel string // LOG_LEVEL (默认: info)

	// API端点（固定值，无需配置）
	UpstreamBaseURL string
	AuthBaseURL     string

	// 超时配置
	HTTPTimeout time.Duration
}

func Load() (*Config, error) {
	// 从环境变量加载，没有默认值的必须提供
	email := os.Getenv("GEMINI_BUSINESS_EMAIL")
	if email == "" {
		return nil, &ConfigError{Field: "GEMINI_BUSINESS_EMAIL", Message: "必须设置Gemini Business邮箱"}
	}

	configID := os.Getenv("GEMINI_BUSINESS_CONFIG_ID")
	if configID == "" {
		configID = "d06739ca-6683-46db-bb51-07395a392439" // 提供默认值
	}

	return &Config{
		Email:           email,
		ConfigID:        configID,
		Port:            getEnv("PORT", "8080"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		UpstreamBaseURL: "https://biz-discoveryengine.googleapis.com/v1alpha",
		AuthBaseURL:     "https://accountverification.business.gemini.google",
		HTTPTimeout:     30 * time.Second,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Field + ": " + e.Message
}
