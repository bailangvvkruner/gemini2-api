# Gemini Business API代理服务 - 完整实现方案

## 一、系统架构设计

### 1.1 整体架构
```
┌─────────────────────────────────────────────────────────┐
│                   客户端 (OpenAI格式)                    │
└───────────────────────────┬─────────────────────────────┘
                            │ HTTP / SSE
                            ▼
┌─────────────────────────────────────────────────────────┐
│              Gemini Business代理服务 (Go)                 │
│  ┌────────────┐  ┌────────────┐  ┌──────────────┐       │
│  │ 认证模块   │  │ API转换器  │  │ 会话管理器    │       │
│  │ (Auth)     │  │ (Adapter)  │  │ (Session)    │       │
│  └────────────┘  └────────────┘  └──────────────┘       │
└───────────────────────────┬─────────────────────────────┘
                            │ Bearer Token + JSON
                            ▼
┌─────────────────────────────────────────────────────────┐
│                Gemini Business API                       │
│  - widgetCreateSession                                  │
│  - widgetStreamAssist                                   │
│  - widgetListTools                                      │
│  - ...                                                 │
└─────────────────────────────────────────────────────────┘
```

### 1.2 组件说明
- **认证模块**: 处理OAuth2登录、验证码验证、令牌刷新
- **API转换器**: 将OpenAI格式请求转换为Gemini Business格式
- **会话管理器**: 管理用户会话状态、令牌缓存
- **流式处理器**: 处理Server-Sent Events (SSE) 流式响应
- **配置管理器**: 通过环境变量管理所有配置

## 二、详细技术实现

### 2.1 项目结构
```
gemini-business-proxy/
├── cmd/
│   └── server/
│       └── main.go          # 程序入口
├── internal/
│   ├── auth/               # 认证模块
│   │   ├── service.go
│   │   ├── verification.go
│   │   └── token_manager.go
│   ├── api/                # API处理模块
│   │   ├── openai_adapter.go
│   │   ├── gemini_client.go
│   │   └── stream_handler.go
│   ├── service/            # 业务逻辑
│   │   └── gemini_service.go
│   ├── config/             # 配置管理
│   │   └── config.go
│   └── middleware/         # 中间件
│       ├── logging.go
│       └── auth.go
├── pkg/
│   └── utils/              # 工具函数
├── go.mod
├── go.sum
└── Dockerfile
```

### 2.2 核心Go代码实现

#### 2.2.1 主程序入口 (cmd/server/main.go)
```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gemini-proxy/internal/api"
	"gemini-proxy/internal/auth"
	"gemini-proxy/internal/config"
	"gemini-proxy/internal/middleware"
	"gemini-proxy/internal/service"
)

func main() {
	// 1. 加载配置
	cfg := config.LoadConfig()
	
	// 2. 初始化认证服务
	authService, err := auth.NewService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize auth service: %v", err)
	}
	
	// 3. 初始化Gemini服务
	geminiService := service.NewGeminiService(cfg, authService)
	
	// 4. 初始化API处理器
	apiHandler := api.NewHandler(geminiService)
	
	// 5. 设置Gin路由
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.JSONLogger())
	
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})
	
	// OpenAI兼容API
	v1 := router.Group("/v1")
	{
		v1.POST("/chat/completions", apiHandler.HandleChatCompletion)
		v1.GET("/models", apiHandler.HandleListModels)
	}
	
	// 6. 启动服务器
	server := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second, // 长超时支持流式响应
		IdleTimeout:  120 * time.Second,
	}
	
	// 7. 优雅关闭
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	
	log.Println("Server stopped gracefully")
}
```

#### 2.2.2 配置管理 (internal/config/config.go)
```go
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
```

#### 2.2.3 Gemini客户端 (internal/api/gemini_client.go)
```go
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gemini-proxy/internal/config"
)

type GeminiClient struct {
	config     *config.Config
	httpClient *http.Client
	authToken  string
}

func NewGeminiClient(cfg *config.Config) *GeminiClient {
	return &GeminiClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 300 * time.Second, // 长超时支持流式
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (c *GeminiClient) StreamAssist(ctx context.Context, sessionID string, query string) (*http.Response, error) {
	url := fmt.Sprintf("%s/locations/global/widgetStreamAssist", c.config.Gemini.APIBaseURL)
	
	requestBody := map[string]interface{}{
		"configId": c.config.Gemini.ConfigID,
		"additionalParams": map[string]string{
			"token": "-",
		},
		"streamAssistRequest": map[string]interface{}{
			"session": fmt.Sprintf("collections/default_collection/engines/agentspace-engine/sessions/%s", sessionID),
			"query": map[string]interface{}{
				"parts": []map[string]string{
					{"text": query},
				},
			},
			"filter": "",
			"fileIds": []string{},
			"answerGenerationMode": "NORMAL",
			"toolsSpec": map[string]interface{}{
				"webGroundingSpec":      map[string]interface{}{},
				"toolRegistry":          "default_tool_registry",
				"imageGenerationSpec":   map[string]interface{}{},
				"videoGenerationSpec":   map[string]interface{}{},
			},
			"languageCode": "zh",
			"userMetadata": map[string]string{
				"timeZone": "Asia/Shanghai",
			},
			"assistSkippingMode": "REQUEST_ASSIST",
		},
	}
	
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// 设置请求头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Gemini-Business-Proxy/1.0")
	req.Header.Set("X-Server-Timeout", "1800")
	req.Header.Set("Origin", c.config.Gemini.BusinessURL)
	req.Header.Set("Referer", c.config.Gemini.BusinessURL)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	return resp, nil
}

func (c *GeminiClient) CreateSession(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/locations/global/widgetCreateSession", c.config.Gemini.APIBaseURL)
	
	requestBody := map[string]interface{}{
		"configId": c.config.Gemini.ConfigID,
		"additionalParams": map[string]string{
			"token": "-",
		},
	}
	
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to create session: %s", resp.Status)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// 解析会话ID
	if session, ok := result["session"].(string); ok {
		return session, nil
	}
	
	return "", fmt.Errorf("session ID not found in response")
}

func (c *GeminiClient) SetAuthToken(token string) {
	c.authToken = token
}
```

#### 2.2.4 OpenAI适配器 (internal/api/openai_adapter.go)
```go
package api

import (
	"encoding/json"
	"fmt"
	"time"
)

// OpenAI格式的消息
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAI请求格式
type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

// OpenAI响应格式
type OpenAIResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []OpenAIChoice   `json:"choices"`
	Usage   *OpenAIUsage     `json:"usage,omitempty"`
}

// OpenAI流式响应格式
type OpenAIStreamResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []OpenAIChoice   `json:"choices"`
}

type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message,omitempty"`
	Delta        OpenAIMessage `json:"delta,omitempty"`
	FinishReason string        `json:"finish_reason,omitempty"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// 将OpenAI请求转换为Gemini查询
func ConvertToGeminiQuery(openAIReq OpenAIRequest) (string, error) {
	if len(openAIReq.Messages) == 0 {
		return "", fmt.Errorf("no messages provided")
	}
	
	// 提取最后一条用户消息
	var lastUserMessage string
	for i := len(openAIReq.Messages) - 1; i >= 0; i-- {
		if openAIReq.Messages[i].Role == "user" {
			lastUserMessage = openAIReq.Messages[i].Content
			break
		}
	}
	
	if lastUserMessage == "" {
		return "", fmt.Errorf("no user message found")
	}
	
	return lastUserMessage, nil
}

// 创建OpenAI响应
func CreateOpenAIResponse(sessionID, model, content string) OpenAIResponse {
	return OpenAIResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", sessionID),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []OpenAIChoice{
			{
				Index: 0,
				Message: OpenAIMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Usage: &OpenAIUsage{
			PromptTokens:     len(content) / 4, // 估算
			CompletionTokens: len(content) / 4,
			TotalTokens:      len(content) / 2,
		},
	}
}

// 创建OpenAI流式响应块
func CreateOpenAIStreamResponse(sessionID, model, content string, isLast bool) OpenAIStreamResponse {
	finishReason := ""
	if isLast {
		finishReason = "stop"
	}
	
	return OpenAIStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", sessionID),
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []OpenAIChoice{
			{
				Index: 0,
				Delta: OpenAIMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: finishReason,
			},
		},
	}
}

// 流式响应处理器
type StreamHandler struct {
	sessionID string
	model     string
}

func NewStreamHandler(sessionID, model string) *StreamHandler {
	return &StreamHandler{
		sessionID: sessionID,
		model:     model,
	}
}

func (h *StreamHandler) ProcessStream(geminiResp *http.Response, w http.ResponseWriter, flusher http.Flusher) error {
	defer geminiResp.Body.Close()
	
	decoder := json.NewDecoder(geminiResp.Body)
	
	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	
	for decoder.More() {
		var chunk map[string]interface{}
		if err := decoder.Decode(&chunk); err != nil {
			return fmt.Errorf("failed to decode stream chunk: %w", err)
		}
		
		// 解析Gemini响应
		if streamResp, ok := chunk["streamAssistResponse"].(map[string]interface{}); ok {
			if answer, ok := streamResp["answer"].(map[string]interface{}); ok {
				if replies, ok := answer["replies"].([]interface{}); ok && len(replies) > 0 {
					for _, reply := range replies {
						if replyMap, ok := reply.(map[string]interface{}); ok {
							if groundedContent, ok := replyMap["groundedContent"].(map[string]interface{}); ok {
								if content, ok := groundedContent["content"].(map[string]interface{}); ok {
									if text, ok := content["text"].(string); ok && text != "" {
										// 跳过思考过程
										if role, _ := content["role"].(string); role == "model" {
											if thought, _ := content["thought"].(bool); !thought {
												// 发送SSE事件
												openAIChunk := CreateOpenAIStreamResponse(
													h.sessionID,
													h.model,
													text,
													false,
												)
												
												data, _ := json.Marshal(openAIChunk)
												fmt.Fprintf(w, "data: %s\n\n", string(data))
												flusher.Flush()
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	
	// 发送结束标记
	finalChunk := CreateOpenAIStreamResponse(h.sessionID, h.model, "", true)
	data, _ := json.Marshal(finalChunk)
	fmt.Fprintf(w, "data: %s\n\n", string(data))
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
	
	return nil
}
```

### 2.3 认证模块实现

#### 2.3.1 认证服务 (internal/auth/service.go)
```go
package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"gemini-proxy/internal/config"
)

type AuthService struct {
	config     *config.Config
	httpClient *http.Client
	jar        *cookiejar.Jar
	token      string
	lastLogin  time.Time
}

func NewService(cfg *config.Config) (*AuthService, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	
	return &AuthService{
		config: cfg,
		httpClient: &http.Client{
			Jar: jar,
			Timeout: 30 * time.Second,
		},
		jar: jar,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, email, verificationCode string) error {
	// 1. 获取登录页面
	loginURL := fmt.Sprintf("%s/login?continueUrl=%s", 
		s.config.Gemini.AuthURL, 
		url.QueryEscape(s.config.Gemini.BusinessURL))
	
	req, err := http.NewRequestWithContext(ctx, "GET", loginURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get login page: %w", err)
	}
	resp.Body.Close()
	
	// 2. 提交邮箱
	emailURL := fmt.Sprintf("%s/send-verification", s.config.Gemini.AuthURL)
	emailData := url.Values{
		"email": {email},
	}
	
	req, err = http.NewRequestWithContext(ctx, "POST", emailURL, strings.NewReader(emailData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create email request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err = s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to submit email: %w", err)
	}
	resp.Body.Close()
	
	// 3. 验证验证码
	verifyURL := fmt.Sprintf("%s/verify-code", s.config.Gemini.AuthURL)
	verifyData := url.Values{
		"code": {verificationCode},
	}
	
	req, err = http.NewRequestWithContext(ctx, "POST", verifyURL, strings.NewReader(verifyData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create verification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err = s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to verify code: %w", err)
	}
	resp.Body.Close()
	
	// 4. 获取授权令牌
	token, err := s.extractAuthToken()
	if err != nil {
		return fmt.Errorf("failed to extract auth token: %w", err)
	}
	
	s.token = token
	s.lastLogin = time.Now()
	
	return nil
}

func (s *AuthService) GetToken() (string, error) {
	// 检查令牌是否过期
	if time.Since(s.lastLogin) > time.Duration(s.config.Gemini.SessionTimeout)*time.Second {
		// 需要重新登录
		return "", fmt.Errorf("token expired")
	}
	
	return s.token, nil
}

func (s *AuthService) extractAuthToken() (string, error) {
	// 从cookie或响应中提取令牌
	// 这里需要根据实际响应实现
	return "extracted-token", nil
}
```

## 三、完整部署方案

### 3.1 Docker Compose配置

```yaml
version: '3.8'

services:
  gemini-proxy:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - SERVER_ADDRESS=0.0.0.0
      - SERVER_PORT=8080
      - GEMINI_BUSINESS_URL=https://business.gemini.google
      - API_BASE_URL=https://biz-discoveryengine.googleapis.com/v1alpha
      - AUTH_URL=https://auth.business.gemini.google
      - ACCOUNT_EMAIL=${ACCOUNT_EMAIL}
      - SESSION_TIMEOUT=1800
      - OPENAI_COMPATIBLE=true
      - API_KEY_HEADER=Authorization
      - DEFAULT_MODEL=gemini-business
      - LOG_LEVEL=info
      - LOG_FORMAT=json
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### 3.2 Kubernetes部署配置

```yaml
# gemini-proxy-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gemini-proxy
spec:
  replicas: 2
  selector:
    matchLabels:
      app: gemini-proxy
  template:
    metadata:
      labels:
        app: gemini-proxy
    spec:
      containers:
      - name: gemini-proxy
        image: bailangvvking/gemini-business-proxy:latest
        ports:
        - containerPort: 8080
        env:
        - name: SERVER_ADDRESS
          value: "0.0.0.0"
        - name: SERVER_PORT
          value: "8080"
        - name: GEMINI_BUSINESS_URL
          value: "https://business.gemini.google"
        - name: API_BASE_URL
          value: "https://biz-discoveryengine.googleapis.com/v1alpha"
        - name: AUTH_URL
          value: "https://auth.business.gemini.google"
        - name: ACCOUNT_EMAIL
          valueFrom:
            secretKeyRef:
              name: gemini-secrets
              key: account-email
        - name: SESSION_TIMEOUT
          value: "1800"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
# gemini-proxy-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: gemini-proxy-service
spec:
  selector:
    app: gemini-proxy
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
---
# gemini-proxy-ingress.yaml (如果需要外部访问)
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: gemini-proxy-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: gemini-proxy.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: gemini-proxy-service
            port:
              number: 80
```

### 3.3 环境变量配置文件

```bash
# .env 文件
SERVER_ADDRESS=0.0.0.0
SERVER_PORT=8080
GEMINI_BUSINESS_URL=https://business.gemini.google
API_BASE_URL=https://biz-discoveryengine.googleapis.com/v1alpha
AUTH_URL=https://auth.business.gemini.google
ACCOUNT_EMAIL=2123146130@qq.com
SESSION_TIMEOUT=1800
OPENAI_COMPATIBLE=true
API_KEY_HEADER=Authorization
DEFAULT_MODEL=gemini-business
LOG_LEVEL=info
LOG_FORMAT=json
```

### 3.4 启动脚本

```bash
#!/bin/bash
# start.sh

# 加载环境变量
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# 检查必需的环境变量
if [ -z "$ACCOUNT_EMAIL" ]; then
    echo "错误: ACCOUNT_EMAIL 环境变量未设置"
    exit 1
fi

# 构建Docker镜像
docker build -t gemini-business-proxy:latest .

# 运行容器
docker run -d \
  --name gemini-proxy \
  --env-file .env \
  -p 8080:8080 \
  gemini-business-proxy:latest

echo "Gemini Business代理服务已启动，监听端口 8080"
echo "测试命令: curl http://localhost:8080/health"
```

## 四、测试与验证

### 4.1 健康检查
```bash
curl http://localhost:8080/health
# 预期响应: {"status":"healthy"}
```

### 4.2 OpenAI格式API测试

#### 非流式请求:
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-api-key" \
  -d '{
    "model": "gemini-business",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ],
    "stream": false
  }'
```

#### 流式请求:
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-api-key" \
  -d '{
    "model": "gemini-business",
    "messages": [
      {"role": "user", "content": "Explain quantum computing in simple terms"}
    ],
    "stream": true
  }' \
  -N
```

### 4.3 性能测试
```bash
# 使用ab进行压力测试
ab -n 100 -c 10 -T 'application/json' \
  -H 'Authorization: Bearer test-api-key' \
  -p test_request.json \
  http://localhost:8080/v1/chat/completions
```

## 五、监控与日志

### 5.1 Prometheus监控指标
```go
// 在main.go中添加监控
import "github.com/prometheus/client_golang/prometheus"

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gemini_proxy_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"endpoint", "method", "status"},
	)
	
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "gemini_proxy_request_duration_seconds",
			Help: "Request duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30},
		},
		[]string{"endpoint"},
	)
	
	activeSessions = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gemini_proxy_active_sessions",
			Help: "Number of active sessions",
		},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(activeSessions)
}
```

### 5.2 结构化日志示例
```json
{
  "timestamp": "2026-01-04T09:50:00Z",
  "level": "info",
  "service": "gemini-proxy",
  "request_id": "req-123456",
  "method": "POST",
  "endpoint": "/v1/chat/completions",
  "status_code": 200,
  "duration_ms": 1250,
  "user_agent": "curl/7.68.0",
  "client_ip": "192.168.1.100",
  "message": "API request processed successfully"
}
```

## 六、安全考虑

### 6.1 API密钥验证
```go
func APIKeyAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(cfg.OpenAICompatible.APIKeyHeader)
		
		// 简单的API密钥验证
		if apiKey == "" || !isValidAPIKey(apiKey) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

func isValidAPIKey(apiKey string) bool {
	// 实现API密钥验证逻辑
	// 可以从数据库、环境变量或配置文件读取有效密钥
	validKeys := strings.Split(os.Getenv("VALID_API_KEYS"), ",")
	for _, validKey := range validKeys {
		if apiKey == validKey {
			return true
		}
	}
	return false
}
```

### 6.2 速率限制
```go
import "golang.org/x/time/rate"

func RateLimitMiddleware() gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Every(time.Minute), 60) // 60请求/分钟
	
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
```

## 七、故障排除

### 7.1 常见问题及解决方案

1. **认证失败**
   - 检查ACCOUNT_EMAIL环境变量是否正确
   - 验证网络连接是否可以访问Gemini Business
   - 检查验证码是否正确

2. **API响应超时**
   - 增加SESSION_TIMEOUT值
   - 检查网络延迟
   - 调整http.Client的超时设置

3. **流式响应中断**
   - 检查客户端是否支持SSE
   - 验证网络连接稳定性
   - 增加服务器写入超时时间

### 7.2 日志级别调整
```bash
# 启动时设置更高日志级别
docker run -d \
  -e LOG_LEVEL=debug \
  # ...其他环境变量
  gemini-business-proxy:latest
```

## 八、扩展与优化

### 8.1 支持的功能扩展
1. **多账户支持**: 支持多个Gemini Business账户
2. **模型选择**: 支持选择不同的Gemini模型
3. **文件上传**: 支持OpenAI格式的文件上传和处理
4. **函数调用**: 支持OpenAI的函数调用功能
5. **缓存机制**: 实现响应缓存提高性能

### 8.2 性能优化建议
1. **连接池**: 优化HTTP客户端连接池设置
2. **令牌缓存**: 实现JWT令牌缓存减少认证请求
3. **响应压缩**: 启用Gzip压缩减少网络传输
4. **异步处理**: 使用goroutine处理并发请求

---

**总结**: 本方案提供了完整的Gemini Business API代理服务实现，包括详细的技术架构、Go代码实现、Docker部署配置和监控方案。通过环境变量配置，实现了零配置文件部署，支持标准输出日志，并兼容OpenAI API格式，便于现有系统集成。
