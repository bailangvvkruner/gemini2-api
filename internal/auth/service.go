package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gemini-business-proxy/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type Service interface {
	GetToken() (string, error)
	RefreshToken() (string, error)
	Login() error
}

type authService struct {
	config     *config.Config
	httpClient *http.Client
	token      string
	expiry     time.Time
}

func NewService(cfg *config.Config) Service {
	return &authService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
	}
}

func (s *authService) GetToken() (string, error) {
	if s.token != "" && time.Now().Before(s.expiry.Add(-5*time.Minute)) {
		return s.token, nil
	}

	return s.RefreshToken()
}

func (s *authService) RefreshToken() (string, error) {
	// Check if we need to login first
	if s.token == "" {
		if err := s.Login(); err != nil {
			return "", fmt.Errorf("login failed: %w", err)
		}
	}

	// For now, just return the existing token
	// In a real implementation, you would refresh the token
	return s.token, nil
}

func (s *authService) Login() error {
	// Step 1: Get initial auth page to get CSRF tokens
	authURL := fmt.Sprintf("%s/login?continueUrl=https://business.gemini.google/", s.config.AuthBaseURL)

	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get auth page: %w", err)
	}
	defer resp.Body.Close()

	// Step 2: Submit email
	emailData := url.Values{
		"email": {s.config.Email},
	}

	emailReq, err := http.NewRequest("POST",
		fmt.Sprintf("%s/_/AuthPortalFederationUi/data/batchexecute", s.config.AuthBaseURL),
		strings.NewReader(emailData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create email request: %w", err)
	}

	emailReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	emailReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Step 3: Simulate OAuth flow
	// In a real implementation, you would handle the full OAuth flow
	// including email verification and token exchange

	// For now, create a mock JWT token for testing
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "https://business.gemini.google",
		"aud": "https://biz-discoveryengine.googleapis.com",
		"sub": fmt.Sprintf("csesidx/%s", "mock_session_id"),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(5 * time.Minute).Unix(),
		"nbf": time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return fmt.Errorf("failed to sign token: %w", err)
	}

	s.token = tokenString
	s.expiry = time.Now().Add(5 * time.Minute)

	return nil
}

// API Client methods
func (s *authService) doRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	token, err := s.GetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://business.gemini.google/")
	req.Header.Set("Origin", "https://business.gemini.google")
	req.Header.Set("sec-ch-ua", `"Not/A)Brand";v="8", "Chromium";v="126", "Google Chrome";v="126"`)
	req.Header.Set("sec-ch-ua-arch", `"x86"`)
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)

	return s.httpClient.Do(req)
}

func (s *authService) DoJSONRequest(ctx context.Context, method, url string, requestBody, responseBody interface{}) error {
	var body io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	resp, err := s.doRequest(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if responseBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(responseBody); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
