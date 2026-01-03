package api

import (
	"fmt"
	"net/http"
	"sync"

	"gemini-business-proxy/internal/auth"
	"gemini-business-proxy/internal/config"

	"github.com/gin-gonic/gin"
)

type VerificationWebHandler struct {
	config            *config.Config
	verificationStore *VerificationStore
	authService       auth.Service
}

type VerificationStore struct {
	sync.RWMutex
	verificationCodes map[string]string      // email -> code
	pendingRequests   map[string]chan string // requestID -> channel for code
}

func NewVerificationWebHandler(cfg *config.Config, authSvc auth.Service) *VerificationWebHandler {
	return &VerificationWebHandler{
		config: cfg,
		verificationStore: &VerificationStore{
			verificationCodes: make(map[string]string),
			pendingRequests:   make(map[string]chan string),
		},
		authService: authSvc,
	}
}

func (h *VerificationWebHandler) SetupRoutes(router *gin.Engine) {
	// 验证码Web界面路由组
	verifyGroup := router.Group("/verify")
	{
		// 验证码输入页面
		verifyGroup.GET("", h.verificationPage)

		// 提交验证码
		verifyGroup.POST("/submit", h.submitVerification)

		// 获取验证状态
		verifyGroup.GET("/status", h.verificationStatus)

		// 触发发送验证码
		verifyGroup.POST("/send-code", h.sendVerificationCode)
	}
}

// verificationPage 验证码输入页面
func (h *VerificationWebHandler) verificationPage(c *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Gemini Business 验证码输入</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 40px;
            max-width: 400px;
            width: 100%;
        }
        .logo {
            text-align: center;
            margin-bottom: 30px;
        }
        .logo h1 {
            color: #333;
            margin: 10px 0;
            font-size: 24px;
        }
        .logo p {
            color: #666;
            font-size: 14px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #555;
            font-weight: 500;
        }
        input {
            width: 100%;
            padding: 12px 16px;
            border: 2px solid #e1e5e9;
            border-radius: 10px;
            font-size: 16px;
            transition: border-color 0.3s;
            box-sizing: border-box;
        }
        input:focus {
            outline: none;
            border-color: #667eea;
        }
        .verification-input {
            letter-spacing: 10px;
            font-size: 24px;
            text-align: center;
            font-weight: bold;
        }
        .btn {
            width: 100%;
            padding: 14px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 10px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s;
        }
        .btn:hover {
            transform: translateY(-2px);
        }
        .btn:active {
            transform: translateY(0);
        }
        .status {
            margin-top: 20px;
            padding: 10px;
            border-radius: 8px;
            text-align: center;
            display: none;
        }
        .success {
            background: #d4edda;
            color: #155724;
            display: block;
        }
        .error {
            background: #f8d7da;
            color: #721c24;
            display: block;
        }
        .info {
            background: #d1ecf1;
            color: #0c5460;
            display: block;
        }
        .email-info {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 10px;
            margin-bottom: 20px;
            text-align: center;
        }
        .email-info strong {
            color: #667eea;
        }
        .steps {
            margin: 20px 0;
            padding: 0;
            list-style: none;
        }
        .steps li {
            padding: 8px 0;
            color: #666;
            display: flex;
            align-items: center;
        }
        .steps li:before {
            content: "✓";
            color: #28a745;
            margin-right: 10px;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">
            <h1>🔐 Gemini Business 验证</h1>
            <p>请输入邮箱收到的验证码</p>
        </div>
        
        <div class="email-info">
            邮箱：<strong>{{.Email}}</strong>
        </div>
        
        <ul class="steps">
            <li>验证码已发送到您的邮箱</li>
            <li>请在10分钟内输入验证码</li>
            <li>验证成功后即可使用API</li>
        </ul>
        
        <form id="verificationForm">
            <div class="form-group">
                <label for="code">6位验证码</label>
                <input type="text" 
                       id="code" 
                       name="code" 
                       class="verification-input" 
                       maxlength="6" 
                       placeholder="______"
                       required
                       pattern="[A-Z0-9]{6}"
                       title="请输入6位大写字母或数字">
            </div>
            
            <div class="form-group">
                <button type="submit" class="btn">验证并启动服务</button>
            </div>
        </form>
        
        <div id="status" class="status"></div>
    </div>

    <script>
        const form = document.getElementById('verificationForm');
        const statusDiv = document.getElementById('status');
        const codeInput = document.getElementById('code');
        
        // 自动聚焦到输入框
        codeInput.focus();
        
        // 输入时自动切换到下一个输入框（模拟6位输入）
        codeInput.addEventListener('input', function(e) {
            if (this.value.length === 6) {
                this.value = this.value.toUpperCase();
            }
        });
        
        form.addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const code = codeInput.value.trim().toUpperCase();
            
            if (code.length !== 6) {
                showError('请输入6位验证码');
                return;
            }
            
            // 显示加载状态
            const submitBtn = form.querySelector('button[type="submit"]');
            const originalText = submitBtn.textContent;
            submitBtn.textContent = '验证中...';
            submitBtn.disabled = true;
            
            try {
                const response = await fetch('/verify/submit', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ code: code })
                });
                
                const data = await response.json();
                
                if (data.success) {
                    showSuccess('验证成功！服务已启动。');
                    // 3秒后跳转到健康检查页面
                    setTimeout(() => {
                        window.location.href = '/health';
                    }, 3000);
                } else {
                    showError(data.message || '验证失败');
                }
            } catch (error) {
                showError('网络错误，请重试');
            } finally {
                submitBtn.textContent = originalText;
                submitBtn.disabled = false;
            }
        });
        
        // 监听键盘事件，支持回车提交
        codeInput.addEventListener('keydown', function(e) {
            if (e.key === 'Enter' && this.value.length === 6) {
                form.dispatchEvent(new Event('submit'));
            }
        });
        
        function showSuccess(message) {
            statusDiv.textContent = message;
            statusDiv.className = 'status success';
        }
        
        function showError(message) {
            statusDiv.textContent = message;
            statusDiv.className = 'status error';
            // 清空输入框
            codeInput.value = '';
            codeInput.focus();
        }
        
        // 自动检查验证状态
        async function checkVerificationStatus() {
            try {
                const response = await fetch('/verify/status');
                const data = await response.json();
                
                if (data.verified) {
                    showSuccess('已验证成功！正在跳转...');
                    setTimeout(() => {
                        window.location.href = '/health';
                    }, 2000);
                }
            } catch (error) {
                // 忽略检查错误
            }
        }
        
        // 每5秒检查一次验证状态
        setInterval(checkVerificationStatus, 5000);
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// submitVerification 提交验证码
func (h *VerificationWebHandler) submitVerification(c *gin.Context) {
	var request struct {
		Code string `json:"code" binding:"required,min=6,max=6"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "验证码必须是6位字符",
		})
		return
	}

	// 验证码标准化（大写）
	code := request.Code

	// 存储验证码
	h.verificationStore.Lock()
	h.verificationStore.verificationCodes[h.config.Email] = code
	h.verificationStore.Unlock()

	// 尝试使用验证码登录
	if err := h.authService.Login(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("验证失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "验证成功！服务已启动并可以处理API请求。",
	})
}

// verificationStatus 获取验证状态
func (h *VerificationWebHandler) verificationStatus(c *gin.Context) {
	h.verificationStore.RLock()
	_, hasCode := h.verificationStore.verificationCodes[h.config.Email]
	h.verificationStore.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"verified": hasCode,
		"email":    h.config.Email,
		"has_code": hasCode,
	})
}

// sendVerificationCode 触发发送验证码
func (h *VerificationWebHandler) sendVerificationCode(c *gin.Context) {
	// 这里应该调用实际的发送验证码API
	// 暂时返回成功
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "验证码已发送到您的邮箱，请查收。",
	})
}
