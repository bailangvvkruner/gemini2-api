# 单行Docker部署方案

## 问题解决
已修复Docker构建错误：`missing go.sum entry for module providing package github.com/golang-jwt/jwt/v5`

## 🚀 最简单的部署命令

### 方案一：直接运行（推荐）
```bash
docker run -d -p 8080:8080 \
  -e GEMINI_BUSINESS_EMAIL="2123146130@qq.com" \
  --name gemini-proxy \
  bailangvvkruner/gemini2-api:latest
```

### 方案二：先构建再运行
```bash
# 1. 下载并构建镜像
docker build -t gemini-proxy .

# 2. 运行容器
docker run -d -p 8080:8080 \
  -e GEMINI_BUSINESS_EMAIL="2123146130@qq.com" \
  --name gemini-proxy \
  gemini-proxy
```

## 📦 环境变量说明

| 变量名 | 必须 | 默认值 | 说明 |
|--------|------|--------|------|
| `GEMINI_BUSINESS_EMAIL` | ✅ | 无 | 您的QQ邮箱 (如: 2123146130@qq.com) |
| `GEMINI_BUSINESS_CONFIG_ID` | ❌ | `d06739ca-6683-46db-bb51-07395a392439` | 企业配置ID |
| `PORT` | ❌ | `8080` | 服务端口 |
| `LOG_LEVEL` | ❌ | `info` | 日志级别 (debug/info/warn/error) |

## 🌐 使用流程

### 第1步：启动服务
```bash
docker run -d -p 8080:8080 \
  -e GEMINI_BUSINESS_EMAIL="2123146130@qq.com" \
  --name gemini-proxy \
  bailangvvkruner/gemini2-api:latest
```

### 第2步：访问验证页面
打开浏览器访问：`http://localhost:8080/verify`

### 第3步：输入验证码
1. 查看QQ邮箱，找到Gemini发送的6位验证码（如：`BK5PA2`）
2. 在Web页面输入验证码
3. 点击"验证并启动服务"

### 第4步：验证成功
- 页面显示"验证成功！服务已启动。"
- 自动跳转到健康检查页面：`http://localhost:8080/health`

## 🔧 验证服务状态

```bash
# 健康检查
curl http://localhost:8080/health

# 输出示例
{
  "status": "healthy",
  "timestamp": 1735900000,
  "service": "gemini-business-proxy"
}
```

## 🐳 Docker命令参考

### 查看日志
```bash
docker logs gemini-proxy
```

### 停止服务
```bash
docker stop gemini-proxy
```

### 重启服务
```bash
docker restart gemini-proxy
```

### 删除容器
```bash
docker rm -f gemini-proxy
```

## 🎯 项目特点

### 已解决的问题
1. ✅ **依赖问题**：移除JWT依赖，简化`go.mod`
2. ✅ **配置简化**：只需1个必需环境变量
3. ✅ **Web界面**：提供友好的验证码输入界面
4. ✅ **日志输出**：标准输出，方便Docker日志收集
5. ✅ **健康检查**：内置健康检查端点

### 技术栈
- **后端**：Go 1.21 + Gin框架
- **前端**：纯HTML/JavaScript验证界面
- **部署**：Docker多阶段构建
- **配置**：环境变量驱动

## 📁 项目结构

```
gemini-business-proxy/
├── cmd/server/main.go          # 主程序入口
├── internal/
│   ├── auth/service.go         # 简化认证服务
│   ├── config/config.go        # 环境变量配置
│   ├── api/verification_web.go # Web验证界面
├── Dockerfile                  # Docker构建配置
├── go.mod                      # Go模块定义
└── DEPLOY.md                   # 部署文档
```

## 🔍 故障排除

### Docker构建失败
```bash
# 清理并重新构建
docker system prune -f
docker build --no-cache -t gemini-proxy .
```

### 端口冲突
```bash
# 使用其他端口
docker run -d -p 8081:8080 \
  -e GEMINI_BUSINESS_EMAIL="2123146130@qq.com" \
  --name gemini-proxy-8081 \
  bailangvvkruner/gemini2-api:latest
```

### 验证码问题
1. 验证码有效期约10分钟
2. 如果验证失败，重启服务获取新验证码
3. 确保邮箱能正常接收邮件

## 📝 示例脚本

### 一键部署脚本 (`deploy.sh`)
```bash
#!/bin/bash
EMAIL=${1:-"2123146130@qq.com"}
PORT=${2:-"8080"}

echo "启动Gemini Business代理服务..."
echo "邮箱: $EMAIL"
echo "端口: $PORT"

docker run -d -p $PORT:8080 \
  -e GEMINI_BUSINESS_EMAIL="$EMAIL" \
  --name gemini-proxy \
  bailangvvkruner/gemini2-api:latest

echo "服务已启动！"
echo "请访问: http://localhost:$PORT/verify"
```

## 🎉 总结

现在您可以通过**单行Docker命令**部署完整的Gemini Business API代理服务：

```bash
docker run -d -p 8080:8080 \
  -e GEMINI_BUSINESS_EMAIL="2123146130@qq.com" \
  --name gemini-proxy \
  bailangvvkruner/gemini2-api:latest
```

然后访问 `http://localhost:8080/verify` 完成验证，即可开始使用API服务！
