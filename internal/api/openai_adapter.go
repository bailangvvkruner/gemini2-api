package api

import (
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
