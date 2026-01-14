# 📖 使用示例大全

## 🎯 基础使用

### 1. Docker Run 命令模板

```bash
# 最简命令（必须配置）
docker run -d \
  --name gemini-proxy \
  -p 8080:8080 \
  -e BEARER_TOKEN="YOUR_TOKEN" \
  -e CONFIG_ID="YOUR_CONFIG" \
  ghcr.io/yourusername/gemini-proxy:latest

# 完整配置（推荐）
docker run -d \
  --name gemini-proxy \
  --restart unless-stopped \
  -p 8080:8080 \
  -e BEARER_TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -e CONFIG_ID="d06739ca-6683-46db-bb51-07395a392439" \
  -e PORT="8080" \
  -e DEBUG="false" \
  -e PROXY_URL="" \
  -e TZ=Asia/Shanghai \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  ghcr.io/yourusername/gemini-proxy:latest
```

### 2. 环境变量配置

```bash
# 导出变量（推荐方式）
export BEARER_TOKEN="your_token_here"
export CONFIG_ID="your_config_here"
export PORT="8080"
export DEBUG="false"

# 运行时直接使用
docker run -d \
  -e BEARER_TOKEN \
  -e CONFIG_ID \
  -e PORT \
  -e DEBUG \
  ghcr.io/yourusername/gemini-proxy:latest
```

## 🌐 API 调用示例

### 1. cURL 流式响应

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dummy" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [
      {"role": "user", "content": "你好，请用中文介绍一下Python"}
    ],
    "stream": true,
    "temperature": 0.7,
    "max_tokens": 1000
  }'
```

### 2. cURL 非流式响应

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-pro",
    "messages": [
      {"role": "system", "content": "你是一个专业的技术助手"},
      {"role": "user", "content": "解释一下什么是Docker"}
    ],
    "stream": false,
    "temperature": 0.5
  }'
```

### 3. 多轮对话

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [
      {"role": "user", "content": "你好"},
      {"role": "assistant", "content": "你好！有什么我可以帮助你的吗？"},
      {"role": "user", "content": "推荐一个Python学习资源"}
    ],
    "stream": true
  }'
```

## 🐍 Python 客户端

### 1. OpenAI 库（推荐）

```python
import openai

# 配置
openai.api_base = "http://localhost:8080/v1"
openai.api_key = "dummy"  # 任意值

# 流式响应
def stream_chat():
    response = openai.ChatCompletion.create(
        model="gemini-2.5-flash",
        messages=[
            {"role": "user", "content": "你好，请介绍一下Go语言"}
        ],
        stream=True,
        temperature=0.7
    )
    
    for chunk in response:
        if hasattr(chunk.choices[0].delta, 'content'):
            print(chunk.choices[0].delta.content, end="", flush=True)

# 非流式响应
def normal_chat():
    response = openai.ChatCompletion.create(
        model="gemini-2.5-pro",
        messages=[
            {"role": "user", "content": "什么是容器化？"}
        ],
        stream=False
    )
    
    print(response.choices[0].message.content)

if __name__ == "__main__":
    stream_chat()
```

### 2. requests 库

```python
import requests
import json

def chat_stream():
    url = "http://localhost:8080/v1/chat/completions"
    headers = {"Content-Type": "application/json"}
    data = {
        "model": "gemini-2.5-flash",
        "messages": [{"role": "user", "content": "你好"}],
        "stream": True
    }
    
    response = requests.post(url, headers=headers, json=data, stream=True)
    
    for line in response.iter_lines():
        if line:
            print(line.decode('utf-8'))

def chat_normal():
    url = "http://localhost:8080/v1/chat/completions"
    headers = {"Content-Type": "application/json"}
    data = {
        "model": "gemini-2.5-flash",
        "messages": [{"role": "user", "content": "你好"}],
        "stream": False
    }
    
    response = requests.post(url, headers=headers, json=data)
    print(response.json())
```

## 📱 JavaScript/Node.js

### 1. OpenAI SDK

```javascript
const OpenAI = require('openai');

const openai = new OpenAI({
  baseURL: 'http://localhost:8080/v1',
  apiKey: 'dummy'
});

async function streamChat() {
  const stream = await openai.chat.completions.create({
    model: 'gemini-2.5-flash',
    messages: [{role: 'user', content: '你好'}],
    stream: true
  });

  for await (const chunk of stream) {
    process.stdout.write(chunk.choices[0]?.delta?.content || '');
  }
}

async function normalChat() {
  const response = await openai.chat.completions.create({
    model: 'gemini-2.5-pro',
    messages: [{role: 'user', content: '什么是Kubernetes'}],
    stream: false
  });

  console.log(response.choices[0].message.content);
}

streamChat();
```

### 2. Fetch API

```javascript
async function chatStream() {
  const response = await fetch('http://localhost:8080/v1/chat/completions', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      model: 'gemini-2.5-flash',
      messages: [{role: 'user', content: '你好'}],
      stream: true
    })
  });

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  
  while (true) {
    const {done, value} = await reader.read();
    if (done) break;
    console.log(decoder.decode(value));
  }
}
```

## 🎨 高级用法

### 1. 系统提示词

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [
      {"role": "system", "content": "你是一个专业的Python编程助手，只用中文回答，代码要详细注释"},
      {"role": "user", "content": "写一个快速排序算法"}
    ],
    "stream": true
  }'
```

### 2. 多轮对话上下文

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [
      {"role": "user", "content": "Docker是什么？"},
      {"role": "assistant", "content": "Docker是一个容器化平台..."},
      {"role": "user", "content": "它和虚拟机有什么区别？"}
    ],
    "stream": true
  }'
```

### 3. 参数调优

```bash
# 低温度（更确定性）
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "你好"}],
    "temperature": 0.1,
    "stream": true
  }'

# 高温度（更有创意）
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "写一首诗"}],
    "temperature": 1.5,
    "stream": true
  }'
```

## 🔧 工具函数

### 1. 健康检查

```bash
curl http://localhost:8080/health
# 返回: {"status":"ok","timestamp":"2026-01-14T10:00:00Z"}
```

### 2. 获取模型列表

```bash
curl http://localhost:8080/v1/models
# 返回: {"object":"list","data":[...]}
```

### 3. 测试连接

```bash
# 快速测试
curl -s http://localhost:8080/health | jq .

# 完整测试
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"test"}],"stream":false}' \
  | jq .
```

## 📊 日志分析

### 1. 查看请求日志

```bash
# 实时日志
docker logs -f gemini-proxy 2>&1 | grep "Request"

# 统计请求次数
docker logs gemini-proxy 2>&1 | grep "Request" | wc -l

# 查看错误
docker logs gemini-proxy 2>&1 | grep "ERROR"
```

### 2. 性能监控

```bash
# 查看响应时间
docker logs -f gemini-proxy 2>&1 | grep "duration"

# 查看资源占用
docker stats gemini-proxy
```

## 🚨 故障排查

### 1. 认证失败

```bash
# 检查 Token
echo $BEARER_TOKEN | head -c 50

# 重新获取 Token
# 从浏览器 Network 标签复制
```

### 2. 配置错误

```bash
# 查看环境变量
docker exec gemini-proxy env

# 检查配置
docker logs gemini-proxy | grep "Config"
```

### 3. 网络问题

```bash
# 测试 API 连通性
curl -I https://biz-discoveryengine.googleapis.com

# 检查容器网络
docker exec gemini-proxy ping -c 3 8.8.8.8
```

## 💡 最佳实践

### 1. 生产环境部署

```bash
# 使用 docker-compose
docker-compose up -d

# 设置重启策略
docker run -d --restart unless-stopped ...

# 配置日志轮转
docker run -d \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  ...
```

### 2. 安全配置

```bash
# 不要在命令行暴露 Token
export BEARER_TOKEN="your_token"
docker run -e BEARER_TOKEN ...

# 使用 secrets（Docker Swarm）
echo "your_token" | docker secret create bearer_token -
docker run --secret bearer_token ...
```

### 3. 性能优化

```bash
# 限制资源
docker run -d \
  --memory=512m \
  --cpus=1.0 \
  ...
```

---

**提示**: 所有示例中的 `YOUR_TOKEN` 和 `YOUR_CONFIG` 需要替换为实际值。