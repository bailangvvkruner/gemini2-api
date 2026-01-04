# Gemini Business API 代理服务

### Docker快速部署

```
docker run -d \
  --name gemini-proxy \
  -e SERVER_ADDRESS=0.0.0.0 \
  -e SERVER_PORT=8080 \
  -e GEMINI_BUSINESS_URL=https://business.gemini.google \
  -e API_BASE_URL=https://biz-discoveryengine.googleapis.com/v1alpha \
  -e AUTH_URL=https://auth.business.gemini.google \
  -e ACCOUNT_EMAIL=2123146130@qq.com \
  -e SESSION_TIMEOUT=1800 \
  -e OPENAI_COMPATIBLE=true \
  -e API_KEY_HEADER=Authorization \
  -e DEFAULT_MODEL=gemini-business \
  -e LOG_LEVEL=info \
  -e LOG_FORMAT=json \
  -p 8080:8080 \
  bailangvvking/gemini-business-proxy:latest
```



一个将OpenAI API格式转换为Gemini Business API的代理服务，支持Docker化部署。

## 功能特性

- ✅ **OpenAI API兼容** - 完全兼容OpenAI API格式
- ✅ **流式响应** - 支持Server-Sent Events流式传输
- ✅ **自动认证** - 自动处理Gemini Business的OAuth登录流程
- ✅ **Docker化** - 完整的Docker支持，多阶段构建
- ✅ **健康检查** - 内置健康检查端点
- ✅ **环境变量配置** - 无敏感信息硬编码

## 快速开始

### 1. 环境配置

```bash
# 复制环境配置文件
cp .env.example .env

# 编辑.env文件，填入您的信息
```

`.env`文件配置：
```env
# Gemini Business认证（无需密码！）
GEMINI_BUSINESS_EMAIL=your_email@qq.com
GEMINI_BUSINESS_CONFIG_ID=d06739ca-6683-46db-bb51-07395a392439

# 服务器配置
PORT=8080
LOG_LEVEL=info

# JWT配置（用于代理服务自身的认证）
JWT_SECRET=your_jwt_secret_key_here
```

**重要说明**：Gemini Business使用邮箱+验证码的OAuth认证，不需要传统密码！

### 2. Docker部署

```bash
# 构建并启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f gemini-proxy

# 健康检查
curl http://localhost:8080/health
```

### 3. API使用

#### OpenAI兼容接口

```bash
# 聊天补全（流式）
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "system", "content": "你是一个有用的助手"},
      {"role": "user", "content": "你好，请介绍一下自己"}
    ],
    "stream": true,
    "temperature": 0.7
  }'

# 非流式响应
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "你好"}],
    "stream": false
  }'
```

#### 原生Gemini接口（高级）

```bash
# 创建会话
curl -X POST http://localhost:8080/gemini/sessions \
  -H "Content-Type: application/json"

# 流式对话
curl -X POST http://localhost:8080/gemini/chat \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "your_session_id",
    "message": "你好",
    "stream": true
  }'
```

## 认证流程详解

### Gemini Business认证机制

通过MCP工具分析，Gemini Business的认证流程如下：

```
1. 邮箱提交 → POST /v1/verify-oob-code
   - 输入：邮箱地址
   - 响应：发送6位验证码到邮箱

2. 验证码验证 → POST /data/batchexecute
   - 输入：6位验证码（如BK5PA2）
   - 响应：OAuth授权码

3. Token获取 → OAuth 2.0流程
   - 交换授权码获取JWT Bearer Token
   - Token有效期：5分钟
```

### 代理服务认证实现

本项目实现了完整的认证流程：

1. **自动登录**：服务启动时自动完成OAuth登录
2. **Token刷新**：Token过期前自动刷新
3. **会话管理**：维护活跃的Gemini会话
4. **错误处理**：认证失败时自动重试

## 项目结构

```
gemini-business-proxy/
├── cmd/server/main.go          # 主程序入口
├── internal/
│   ├── auth/service.go         # 认证服务
│   ├── config/config.go        # 配置管理
│   ├── proxy/handler.go        # API代理处理器
│   └── service/gemini.go       # Gemini业务服务
├── Dockerfile                  # 多阶段Docker构建
├── docker-compose.yml          # Docker Compose配置
├── .env.example               # 环境变量示例
└── go.mod                     # Go模块定义
```

## 开发指南

### 本地开发

```bash
# 安装依赖
go mod download

# 运行测试
go test ./...

# 启动开发服务器
go run cmd/server/main.go
```

### 构建部署

```bash
# 本地构建
go build -o gemini-proxy ./cmd/server

# Docker构建
docker build -t gemini-business-proxy .

# Docker Compose
docker-compose up --build
```

## API参考

### OpenAI兼容端点

| 端点 | 方法 | 描述 |
|------|------|------|
| `/v1/chat/completions` | POST | 聊天补全（支持流式） |
| `/v1/models` | GET | 获取可用模型列表 |
| `/health` | GET | 健康检查 |

### Gemini原生端点

| 端点 | 方法 | 描述 |
|------|------|------|
| `/gemini/sessions` | POST | 创建新会话 |
| `/gemini/chat` | POST | 发送消息（支持流式） |
| `/gemini/sessions/{id}` | DELETE | 删除会话 |

## 故障排除

### 常见问题

1. **认证失败**
   - 检查邮箱是否正确
   - 确认邮箱能收到验证码
   - 查看服务日志：`docker-compose logs gemini-proxy`

2. **API调用失败**
   - 检查Config ID是否正确
   - 验证网络连接：`curl https://business.gemini.google`
   - 查看错误日志

3. **流式响应中断**
   - 检查客户端是否支持SSE
   - 增加超时时间设置
   - 查看网络代理配置

### 日志级别

通过`LOG_LEVEL`环境变量控制日志详细程度：
- `debug`: 最详细，包含所有请求/响应
- `info`: 一般信息（默认）
- `warn`: 警告信息
- `error`: 错误信息

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！

## 免责声明

本项目仅供学习和研究使用，请遵守Gemini Business的服务条款。
