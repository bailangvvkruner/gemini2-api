# Gemini Business API 代理服务

**纯Python方案** - 无头浏览器自动登录 + 多账号管理 + API代理

## 🚀 快速开始（3步搞定）

### 1️⃣ 安装依赖
```bash
chmod +x setup.sh
./setup.sh
```

### 2️⃣ 配置账号
```bash
# 编辑 accounts.json
nano accounts.json

# 填入您的Gemini邮箱
{
  "accounts": [
    {
      "email": "2123146130@qq.com",
      "password": "",
      "is_active": true
    }
  ]
}
```

### 3️⃣ 启动自动管理
```bash
python3 multi-account-manager.py
```

**等待浏览器自动登录 → 输入验证码 → 服务自动部署**

## 🎯 核心功能

| 功能 | 说明 |
|------|------|
| **无头浏览器** | 自动打开浏览器，填写邮箱 |
| **验证码输入** | 命令行输入验证码 |
| **自动捕获Token** | 自动获取 Bearer Token + Config ID |
| **多账号轮训** | 支持多个账号自动切换 |
| **Docker部署** | 自动部署和重启服务 |
| **后台监控** | 50分钟后自动轮换 |

## 📁 文件说明

```
.
├── Dockerfile                    # Python 3.11 Docker镜像
├── docker-compose.yml            # Docker编排配置
├── api-proxy.py                  # API代理服务（OpenAI格式）
├── multi-account-manager.py      # 多账号管理器（核心）
├── accounts.example.json         # 账号配置模板
├── requirements.txt              # Python依赖
├── setup.sh                      # 一键安装脚本
├── README.md                     # 主文档
├── README_DOCKER_PYTHON.md       # Docker部署详情
├── README_MULTI_ACCOUNT.md       # 多账号管理详情
└── USAGE_EXAMPLES.md             # 使用示例
```

## 🚀 三种使用方式

### 方式1：自动管理（推荐）
```bash
./setup.sh
python3 multi-account-manager.py
# 浏览器自动登录 → 输入验证码 → 自动部署
```

### 方式2：手动运行
```bash
export BEARER_TOKEN="你的Token"
export CONFIG_ID="你的ConfigID"
python3 api-proxy.py
```

### 方式3：Docker部署
```bash
docker-compose up -d
```

## 🎬 完整工作流程

```
运行 multi-account-manager.py
    ↓
无头浏览器自动打开
    ↓
自动填写邮箱: 2123146130@qq.com
    ↓
自动点击"下一步"
    ↓
检测到验证码输入框
    ↓
🚨 请输入验证码: [您输入]
    ↓
登录成功！自动捕获 Token
    ↓
自动部署到 Docker
    ↓
服务运行中: http://localhost:8080
    ↓
后台监控，50分钟后轮换
```

## 🧪 快速测试

```bash
# 1. 健康检查
curl http://localhost:8080/health

# 2. 流式对话
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"你好"}],"stream":true}'
```

## 🤖 Python客户端

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

## ⏰ Token过期说明

**Bearer Token有效期：1小时**

**自动管理方案：**
- ✅ 50分钟后自动检测
- ✅ 自动轮换下一个账号
- ✅ 无缝切换，零中断

**手动方案：**
- 重新获取Token → 重启服务

## 📚 详细文档

- **多账号管理**: [README_MULTI_ACCOUNT.md](README_MULTI_ACCOUNT.md)
- **Docker部署**: [README_DOCKER_PYTHON.md](README_DOCKER_PYTHON.md)
- **使用示例**: [USAGE_EXAMPLES.md](USAGE_EXAMPLES.md)

## 🔧 环境变量

```bash
# API代理服务
BEARER_TOKEN=你的BearerToken  # 必须
CONFIG_ID=你的ConfigID        # 必须
PORT=8080                     # 可选
DEBUG=false                   # 可选
PROXY_URL=                    # 可选
```

## ✅ 总结

**纯Python方案，无需Go，一键启动，自动管理！**

```bash
# 最简单命令
./setup.sh && python3 multi-account-manager.py
```

**只需输入验证码，其他全自动！**