# 🚀 快速开始指南

## 一行命令启动（最简单）

```bash
docker run -d --name gemini-proxy --restart unless-stopped -p 8080:8080 -e BEARER_TOKEN="YOUR_TOKEN" -e CONFIG_ID="YOUR_CONFIG" ghcr.io/yourusername/gemini-proxy:latest
```

## 完整步骤

### 1️⃣ 获取配置

**Bearer Token:**
1. 打开 [Gemini Business](https://business.gemini.google)
2. 按 F12 → Network 标签
3. 发送消息，找到 `widgetStreamAssist` 请求
4. 复制 Request Headers 中的 `Authorization: Bearer eyJhbGci...`

**Config ID:**
从URL复制：`https://business.gemini.google/home/cid/CONFIG_ID/...`

### 2️⃣ 运行容器

```bash
# 替换下面两个值
export BEARER_TOKEN="你的BearerToken"
export CONFIG_ID="你的ConfigID"

# 运行
docker run -d \
  --name gemini-proxy \
  --restart unless-stopped \
  -p 8080:8080 \
  -e BEARER_TOKEN="$BEARER_TOKEN" \
  -e CONFIG_ID="$CONFIG_ID" \
  -e TZ=Asia/Shanghai \
  ghcr.io/yourusername/gemini-proxy:latest
```

### 3️⃣ 测试使用

```bash
# 流式响应
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"你好"}],"stream":true}'
```

## 🐳 Docker Compose 方式

```bash
# 创建 docker-compose.yml
cat > docker-compose.yml << 'EOF'
version: '3.8'
services:
  gemini-proxy:
    image: ghcr.io/yourusername/gemini-proxy:latest
    container_name: gemini-proxy
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - BEARER_TOKEN=${BEARER_TOKEN}
      - CONFIG_ID=${CONFIG_ID}
      - TZ=Asia/Shanghai
EOF

# 创建 .env 文件
cat > .env << EOF
BEARER_TOKEN=你的BearerToken
CONFIG_ID=你的ConfigID
EOF

# 启动
docker-compose up -d
```

## 📱 Python 客户端

```python
import openai

openai.api_base = "http://localhost:8080/v1"
openai.api_key = "dummy"

response = openai.ChatCompletion.create(
    model="gemini-2.5-flash",
    messages=[{"role": "user", "content": "你好"}],
    stream=True
)

for chunk in response:
    if hasattr(chunk.choices[0].delta, 'content'):
        print(chunk.choices[0].delta.content, end="", flush=True)
```

## 🔧 常用命令

```bash
# 查看日志
docker logs -f gemini-proxy

# 重启服务
docker restart gemini-proxy

# 停止服务
docker stop gemini-proxy

# 删除容器
docker stop gemini-proxy && docker rm gemini-proxy
```

## ⚠️ 常见问题

**Q: 401 错误？**
A: Bearer Token 过期或错误，重新获取

**Q: 404 错误？**
A: Config ID 错误，从URL重新获取

**Q: 连接超时？**
A: 检查网络，或设置 PROXY_URL 环境变量

---

**完成！** 现在你可以使用 OpenAI 格式的 API 访问 Gemini Business 了。