package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gemini-business-proxy/internal/config"
	"gemini-business-proxy/internal/service"
)

type Handler struct {
	config        *config.Config
	geminiService *service.GeminiService
}

func NewHandler(geminiService *service.GeminiService) *Handler {
	return &Handler{
		config:        geminiService.Config(),
		geminiService: geminiService,
	}
}

func (h *Handler) HandleChatCompletion(w http.ResponseWriter, r *http.Request) {
	// 验证请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 解析OpenAI请求
	var openAIReq OpenAIRequest
	if err := json.Unmarshal(body, &openAIReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// 转换消息为Gemini查询
	query, err := ConvertToGeminiQuery(openAIReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to convert query: %v", err), http.StatusBadRequest)
		return
	}

	// 获取认证令牌
	token, err := h.geminiService.GetAuthToken()
	if err != nil {
		http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
		return
	}

	// 创建Gemini客户端
	client := NewGeminiClient(h.config)
	client.SetAuthToken(token)

	// 创建会话
	sessionID, err := client.CreateSession(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
		return
	}

	// 处理流式或非流式请求
	if openAIReq.Stream {
		h.handleStreamResponse(w, r, client, sessionID, query, openAIReq.Model)
	} else {
		h.handleNonStreamResponse(w, r, client, sessionID, query, openAIReq.Model)
	}
}

func (h *Handler) handleStreamResponse(w http.ResponseWriter, r *http.Request, client *GeminiClient, sessionID, query, model string) {
	// 获取流式响应
	resp, err := client.StreamAssist(r.Context(), sessionID, query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get stream response: %v", err), http.StatusInternalServerError)
		return
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		http.Error(w, fmt.Sprintf("Gemini API error: %s", string(body)), resp.StatusCode)
		return
	}

	// 创建流处理器
	streamHandler := NewStreamHandler(sessionID, model)

	// 获取flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// 处理流式响应
	if err := streamHandler.ProcessStream(resp, w, flusher); err != nil {
		// 错误已处理，不需要再返回错误
		fmt.Printf("Stream processing error: %v\n", err)
	}
}

func (h *Handler) handleNonStreamResponse(w http.ResponseWriter, r *http.Request, client *GeminiClient, sessionID, query, model string) {
	// 获取流式响应但收集所有内容
	resp, err := client.StreamAssist(r.Context(), sessionID, query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get response: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("Gemini API error: %s", string(body)), resp.StatusCode)
		return
	}

	// 解析响应并收集内容
	decoder := json.NewDecoder(resp.Body)
	var fullContent strings.Builder

	for decoder.More() {
		var chunk map[string]interface{}
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			http.Error(w, fmt.Sprintf("Failed to decode response: %v", err), http.StatusInternalServerError)
			return
		}

		// 提取内容
		streamHandler := NewStreamHandler(sessionID, model)
		content, err := streamHandler.extractContentFromChunk(chunk)
		if err == nil && content != "" {
			fullContent.WriteString(content)
		}
	}

	// 创建OpenAI响应
	openAIResp := CreateOpenAIResponse(sessionID, model, fullContent.String())

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// 返回JSON响应
	if err := json.NewEncoder(w).Encode(openAIResp); err != nil {
		fmt.Printf("Failed to encode response: %v\n", err)
	}
}

func (h *Handler) HandleListModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	models := map[string]interface{}{
		"object": "list",
		"data": []map[string]interface{}{
			{
				"id":      "gemini-business",
				"object":  "model",
				"created": 1677664790,
				"owned_by": "gemini",
			},
			{
				"id":      "gemini-business-pro",
				"object":  "model",
				"created": 1677649963,
				"owned_by": "gemini",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

// 简单的API密钥验证中间件
func (h *Handler) APIKeyAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get(h.config.OpenAICompatible.APIKeyHeader)
		
		// 简单的API密钥验证
		if apiKey == "" || !isValidAPIKey(apiKey) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid API key",
			})
			return
		}
		
		next(w, r)
	}
}

func isValidAPIKey(apiKey string) bool {
	// 这里实现实际的API密钥验证逻辑
	// 可以从数据库、环境变量或配置文件读取有效密钥
	// 暂时接受所有非空密钥进行测试
	return apiKey != ""
}
