package auth

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"gemini-business-proxy/internal/config"
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
	
	service := &AuthService{
		config: cfg,
		httpClient: &http.Client{
			Jar: jar,
			Timeout: 30 * time.Second,
		},
		jar: jar,
		token: "",
		lastLogin: time.Time{},
	}
	
	// 异步发送验证码请求，不阻塞服务启动
	go func() {
		if cfg.Gemini.AccountEmail != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			
			if err := service.SendVerificationRequest(ctx, cfg.Gemini.AccountEmail); err != nil {
				log.Printf("Failed to send verification request: %v", err)
				log.Printf("Note: You can manually trigger verification later via /auth/verify endpoint")
			} else {
				log.Printf("Verification code sent to %s", cfg.Gemini.AccountEmail)
				log.Printf("Please check your email and submit the verification code to /auth/verify endpoint")
			}
		}
	}()
	
	return service, nil
}

func (s *AuthService) SendVerificationRequest(ctx context.Context, email string) error {
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
	
	// 2. 提交邮箱请求发送验证码
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
	defer resp.Body.Close()
	
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send verification request: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

func (s *AuthService) Login(ctx context.Context, email, verificationCode string) error {
	// 验证验证码
	verifyURL := fmt.Sprintf("%s/verify-code", s.config.Gemini.AuthURL)
	verifyData := url.Values{
		"code": {verificationCode},
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", verifyURL, strings.NewReader(verifyData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create verification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to verify code: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to verify code: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	// 获取授权令牌
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
	if s.token != "" && time.Since(s.lastLogin) > time.Duration(s.config.Gemini.SessionTimeout)*time.Second {
		// 需要重新登录
		s.token = ""
		s.lastLogin = time.Time{}
		return "", fmt.Errorf("token expired")
	}
	
	// 如果令牌为空，返回错误而不是尝试自动登录
	if s.token == "" {
		return "", fmt.Errorf("not authenticated. Please send verification code to /auth/verify endpoint")
	}
	
	return s.token, nil
}

func (s *AuthService) InitializeLoginWithCode(verificationCode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	return s.Login(ctx, s.config.Gemini.AccountEmail, verificationCode)
}

func (s *AuthService) extractAuthToken() (string, error) {
	// 从cookie或响应中提取令牌
	// 这里需要根据实际响应实现
	// 暂时返回模拟令牌
	return "simulated-auth-token-12345", nil
}
