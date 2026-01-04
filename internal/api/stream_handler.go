package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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
	w.Header().Set("X-Accel-Buffering", "no") // 禁用Nginx缓冲
	
	// 发送初始空消息以确保连接建立
	fmt.Fprintf(w, "data: {}\n\n")
	flusher.Flush()
	
	var accumulatedContent string
	var hasContent bool
	
	for decoder.More() {
		var chunk map[string]interface{}
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode stream chunk: %w", err)
		}
		
		// 解析Gemini响应
		content, err := h.extractContentFromChunk(chunk)
		if err != nil {
			// 跳过无法解析的块
			continue
		}
		
		if content != "" {
			hasContent = true
			accumulatedContent += content
			
			// 发送SSE事件
			openAIChunk := CreateOpenAIStreamResponse(
				h.sessionID,
				h.model,
				content,
				false,
			)
			
			data, err := json.Marshal(openAIChunk)
			if err != nil {
				return fmt.Errorf("failed to marshal OpenAI chunk: %w", err)
			}
			
			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()
		}
	}
	
	// 如果没有任何内容，至少发送一个空响应
	if !hasContent {
		finalChunk := CreateOpenAIStreamResponse(h.sessionID, h.model, "", true)
		data, _ := json.Marshal(finalChunk)
		fmt.Fprintf(w, "data: %s\n\n", string(data))
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
		return nil
	}
	
	// 发送结束标记
	finalChunk := CreateOpenAIStreamResponse(h.sessionID, h.model, "", true)
	data, _ := json.Marshal(finalChunk)
	fmt.Fprintf(w, "data: %s\n\n", string(data))
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
	
	return nil
}

func (h *StreamHandler) extractContentFromChunk(chunk map[string]interface{}) (string, error) {
	// 尝试从不同格式中提取内容
	
	// 格式1: 直接包含content
	if content, ok := chunk["content"].(string); ok && content != "" {
		return content, nil
	}
	
	// 格式2: 包含streamAssistResponse
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
											return text, nil
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
	
	// 格式3: 包含advancedCompleteQueryResponse
	if advancedResp, ok := chunk["advancedCompleteQueryResponse"].(map[string]interface{}); ok {
		if answer, ok := advancedResp["answer"].(map[string]interface{}); ok {
			if replies, ok := answer["replies"].([]interface{}); ok && len(replies) > 0 {
				for _, reply := range replies {
					if replyMap, ok := reply.(map[string]interface{}); ok {
						if groundedContent, ok := replyMap["groundedContent"].(map[string]interface{}); ok {
							if content, ok := groundedContent["content"].(map[string]interface{}); ok {
								if text, ok := content["text"].(string); ok && text != "" {
									return text, nil
								}
							}
						}
					}
				}
			}
		}
	}
	
	return "", fmt.Errorf("no content found in chunk")
}
