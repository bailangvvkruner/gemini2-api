package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gemini-business-proxy/internal/auth"
	"gemini-business-proxy/internal/config"
)

type GeminiService interface {
	CreateSession(ctx context.Context) (string, error)
	StreamChat(ctx context.Context, sessionID string, message string, stream bool) (io.ReadCloser, error)
	CloseSession(ctx context.Context, sessionID string) error
}

type geminiService struct {
	config     *config.Config
	authService auth.Service
	httpClient *http.Client
}

func NewGeminiService(cfg *config.Config, authSvc auth.Service) GeminiService {
	return &geminiService{
		config:     cfg,
		authService: authSvc,
		httpClient: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
	}
}

// CreateSession 创建新的Gemini会话
func (s *geminiService) CreateSession(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/locations/global/widgetCreateSession", s.config.UpstreamBaseURL)
	
	requestBody := map[string]interface{}{
		"configId": s.config.ConfigID,
	}
	
	var response struct {
		Session string `json:"session"`
	}
	
	// 使用类型断言来调用DoJSONRequest
	authSvc, ok := s.authService.(interface {
		DoJSONRequest(ctx context.Context, method, url string, requestBody, responseBody interface{}) error
	})
	if !ok {
		return "", fmt.Errorf("auth service does not support DoJSONRequest")
	}
	
	if err := authSvc.DoJSONRequest(ctx, "POST", url, requestBody, &response); err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	
	return response.Session, nil
}

// StreamChat 发送消息并获取流式响应
func (s *geminiService) StreamChat(ctx context.Context, sessionID string, message string, stream bool) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/locations/global/widgetStreamAssist", s.config.UpstreamBaseURL)
	
	requestBody := map[string]interface{}{
		"configId": s.config.ConfigID,
		"additionalParams": map[string]interface{}{
			"token": "-",
		},
		"streamAssistRequest": map[string]interface{}{
			"session": sessionID,
			"query": map[string]interface{}{
				"parts": []map[string]interface{}{
					{"text": message},
				},
			},
			"filter": "",
			"fileIds": []interface{}{},
			"answerGenerationMode": "NORMAL",
			"toolsSpec": map[string]interface{}{
				"webGroundingSpec": map[string]interface{}{},
				"toolRegistry": "default_tool_registry",
				"imageGenerationSpec": map[string]interface{}{},
				"videoGenerationSpec": map[string]interface{}{},
			},
			"languageCode": "zh",
			"userMetadata": map[string]interface{}{
				"timeZone": "Asia/Shanghai",
			},
			"assistSkippingMode": "REQUEST_ASSIST",
		},
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	
	// 获取认证token
	token, err := s.authService.GetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://business.gemini.google/")
	req.Header.Set("Origin", "https://business.gemini.google")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	return resp.Body, nil
}

// CloseSession 关闭会话
func (s *geminiService) CloseSession(ctx context.Context, sessionID string) error {
	// Gemini Business API通常不需要显式关闭会话
	// 会话会在超时后自动关闭
	return nil
}

// bytes包需要导入
import "bytes"
