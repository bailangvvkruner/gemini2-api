package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gemini-business-proxy/internal/config"
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

func (c *GeminiClient) HealthCheck(ctx context.Context) error {
	// 简单的健康检查，尝试创建一个会话
	_, err := c.CreateSession(ctx)
	return err
}
