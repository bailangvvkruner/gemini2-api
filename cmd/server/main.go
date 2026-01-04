package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gemini-business-proxy/internal/api"
	"gemini-business-proxy/internal/auth"
	"gemini-business-proxy/internal/config"
	"gemini-business-proxy/internal/middleware"
	"gemini-business-proxy/internal/service"
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
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.New()
	router.Use(gin.Recovery())
	
	// 使用自定义日志中间件
	router.Use(func(c *gin.Context) {
		middleware.JSONLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	})
	
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	
	// 版本信息
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version": "1.0.0",
			"service": "gemini-business-proxy",
		})
	})
	
	// 验证码提交路由
	router.POST("/auth/verify", func(c *gin.Context) {
		var request struct {
			VerificationCode string `json:"verification_code" binding:"required"`
		}
		
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// 使用验证码登录
		ctx := context.Background()
		err := authService.Login(ctx, cfg.Gemini.AccountEmail, request.VerificationCode)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"status": "login_successful"})
	})
	
	// 登录状态检查
	router.GET("/auth/status", func(c *gin.Context) {
		token, err := authService.GetToken()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"authenticated": false,
				"error": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"authenticated": true,
			"has_token": token != "",
		})
	})
	
	// OpenAI兼容API
	v1 := router.Group("/v1")
	{
		// 应用API密钥验证中间件
		v1.Use(func(c *gin.Context) {
			apiHandler.APIKeyAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Next()
			})).ServeHTTP(c.Writer, c.Request)
		})
		
		v1.POST("/chat/completions", func(c *gin.Context) {
			apiHandler.HandleChatCompletion(c.Writer, c.Request)
		})
		
		v1.GET("/models", func(c *gin.Context) {
			apiHandler.HandleListModels(c.Writer, c.Request)
		})
	}
	
	// 6. 启动服务器
	serverAddr := cfg.Server.Address
	if cfg.Server.Port != 0 {
		serverAddr = fmt.Sprintf("%s:%d", serverAddr, cfg.Server.Port)
	}
	
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second, // 长超时支持流式响应
		IdleTimeout:  120 * time.Second,
	}
	
	// 7. 优雅关闭
	go func() {
		log.Printf("Starting server on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	
	log.Println("Server stopped gracefully")
}
