package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"gemini-business-proxy/internal/config"
)

type VerificationHandler struct {
	config     *config.Config
	httpClient *http.Client
}

func NewVerificationHandler(cfg *config.Config) *VerificationHandler {
	return &VerificationHandler{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
	}
}

// VerifyWithCode 使用验证码完成OAuth认证
func (h *VerificationHandler) VerifyWithCode(ctx context.Context, email, code string) (string, error) {
	// Step 1: 提交邮箱获取session
	sessionID, err := h.submitEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("提交邮箱失败: %w", err)
	}

	// Step 2: 提交验证码
	token, err := h.submitVerificationCode(ctx, sessionID, code)
	if err != nil {
		return "", fmt.Errorf("验证码验证失败: %w", err)
	}

	return token, nil
}

// submitEmail 提交邮箱地址
func (h *VerificationHandler) submitEmail(ctx context.Context, email string) (string, error) {
	// 构建请求数据
	formData := url.Values{
		"f.req": {fmt.Sprintf(`[[["IjXaFf","[\\"%s\\",null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,2]",null,"generic"]]]`, email)},
		"at":    {""}, // 需要从页面获取实际的at值
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/_/IdentityPlatformFrontendUI/data/batchexecute", h.config.AuthBaseURL),
		strings.NewReader(formData.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 这里简化处理，实际需要解析响应获取session
	// 返回模拟session ID
	return fmt.Sprintf("session_%d", time.Now().Unix()), nil
}

// submitVerificationCode 提交验证码
func (h *VerificationHandler) submitVerificationCode(ctx context.Context, sessionID, code string) (string, error) {
	// 验证码通常是6位字母数字
	if len(code) != 6 {
		return "", fmt.Errorf("验证码必须是6位字符")
	}

	// 构建验证码验证请求
	formData := url.Values{
		"f.req": {fmt.Sprintf(`[[["zZ3tQe","[\\"%s\\",\\"%s\\",null,null,null,null,null,null,2]",null,"generic"]]]`, sessionID, code)},
		"at":    {""},
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/_/IdentityPlatformFrontendUI/data/batchexecute", h.config.AuthBaseURL),
		strings.NewReader(formData.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("验证码验证失败，状态码: %d", resp.StatusCode)
	}

	// 实际实现中需要解析OAuth响应获取token
	// 这里返回模拟的JWT token
	return generateMockToken(email, sessionID), nil
}

func generateMockToken(email, sessionID string) string {
	// 实际应该从OAuth响应中获取
	// 这里生成一个模拟的5分钟有效token
	return fmt.Sprintf("mock_jwt_token_for_%s_%d", email, time.Now().Unix())
}

// GetVerificationCodeFromEnv 从环境变量获取验证码
func GetVerificationCodeFromEnv() string {
	// 在实际的config包中实现
	// 这里返回示例
	return os.Getenv("VERIFICATION_CODE")
}
