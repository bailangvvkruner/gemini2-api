package service

import (
	"context"
	"sync"
	"time"

	"gemini-business-proxy/internal/auth"
	"gemini-business-proxy/internal/config"
)

type GeminiService struct {
	config      *config.Config
	authService *auth.AuthService
	token       string
	tokenExpiry time.Time
	mutex       sync.RWMutex
}

func NewGeminiService(cfg *config.Config, authService *auth.AuthService) *GeminiService {
	return &GeminiService{
		config:      cfg,
		authService: authService,
	}
}

func (s *GeminiService) Config() *config.Config {
	return s.config
}

func (s *GeminiService) GetAuthToken() (string, error) {
	s.mutex.RLock()
	
	// 检查令牌是否有效且未过期
	if s.token != "" && time.Now().Before(s.tokenExpiry) {
		token := s.token
		s.mutex.RUnlock()
		return token, nil
	}
	
	s.mutex.RUnlock()
	
	// 需要获取新令牌
	return s.refreshAuthToken()
}

func (s *GeminiService) refreshAuthToken() (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// 双重检查，防止多个goroutine同时刷新
	if s.token != "" && time.Now().Before(s.tokenExpiry) {
		return s.token, nil
	}
	
	// 从认证服务获取令牌
	token, err := s.authService.GetToken()
	if err != nil {
		// 如果令牌过期，尝试重新登录
		if err.Error() == "token expired" {
			// 这里需要实现重新登录逻辑
			// 暂时返回错误
			return "", err
		}
		return "", err
	}
	
	// 更新令牌和过期时间
	s.token = token
	s.tokenExpiry = time.Now().Add(time.Duration(s.config.Gemini.SessionTimeout) * time.Second)
	
	return s.token, nil
}

func (s *GeminiService) HealthCheck(ctx context.Context) error {
	// 检查认证服务状态
	_, err := s.GetAuthToken()
	if err != nil {
		return err
	}
	
	return nil
}
