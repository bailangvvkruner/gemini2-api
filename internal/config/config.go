package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server struct {
		Address string
		Port    int
	}
	
	Gemini struct {
		BusinessURL    string
		APIBaseURL     string
		AuthURL        string
		AccountEmail   string
		SessionTimeout int
		ConfigID       string
	}
	
	OpenAICompatible struct {
		Enabled       bool
		APIKeyHeader  string
		DefaultModel  string
	}
	
	Logging struct {
		Level  string
		Format string
	}
}

func LoadConfig() *Config {
	cfg := &Config{}
	
	// 服务器配置
	cfg.Server.Address = getEnv("SERVER_ADDRESS", "0.0.0.0")
	cfg.Server.Port = getEnvAsInt("SERVER_PORT", 8080)
	
	// Gemini配置
	cfg.Gemini.BusinessURL = getEnv("GEMINI_BUSINESS_URL", "https://business.gemini.google")
	cfg.Gemini.APIBaseURL = getEnv("API_BASE_URL", "https://biz-discoveryengine.googleapis.com/v1alpha")
	cfg.Gemini.AuthURL = getEnv("AUTH_URL", "https://auth.business.gemini.google")
	cfg.Gemini.AccountEmail = getEnv("ACCOUNT_EMAIL", "")
	cfg.Gemini.SessionTimeout = getEnvAsInt("SESSION_TIMEOUT", 1800)
	cfg.Gemini.ConfigID = getEnv("CONFIG_ID", "d06739ca-6683-46db-bb51-07395a392439")
	
	// OpenAI兼容配置
	cfg.OpenAICompatible.Enabled = getEnvAsBool("OPENAI_COMPATIBLE", true)
	cfg.OpenAICompatible.APIKeyHeader = getEnv("API_KEY_HEADER", "Authorization")
	cfg.OpenAICompatible.DefaultModel = getEnv("DEFAULT_MODEL", "gemini-business")
	
	// 日志配置
	cfg.Logging.Level = getEnv("LOG_LEVEL", "info")
	cfg.Logging.Format = getEnv("LOG_FORMAT", "json")
	
	return cfg
}

// 辅助函数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
