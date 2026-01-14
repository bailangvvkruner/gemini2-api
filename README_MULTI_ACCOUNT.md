# 🤖 Gemini Business 多账号自动管理器

支持**无头浏览器自动登录 + 验证码后台输入 + 多账号轮训**的完整解决方案。

## 🎯 核心功能

### ✅ 无头浏览器自动登录
- 自动打开浏览器（无头模式）
- 自动填写邮箱和密码
- 自动检测验证码输入框
- **后台等待用户输入验证码**

### ✅ 验证码后台输入
- 命令行输入验证码
- 可扩展HTTP API输入
- 支持多种验证码类型

### ✅ 多账号轮训
- 自动管理多个Gemini账号
- 按使用次数智能轮换
- Token过期自动切换
- Docker服务无缝切换

## 🚀 快速开始

### 1. 安装依赖

```bash
# 安装Python依赖
pip install -r requirements.txt

# 安装Playwright浏览器
playwright install chromium
```

### 2. 配置账号

```bash
# 复制示例配置
cp accounts.example.json accounts.json

# 编辑配置，填入账号信息
# 可以只填邮箱，密码和Token会自动获取
```

**accounts.json 示例：**
```json
{
  "accounts": [
    {
      "email": "user1@example.com",
      "password": "可选密码",
      "bearer_token": "",
      "config_id": "",
      "is_active": true
    },
    {
      "email": "user2@example.com", 
      "password": "",
      "bearer_token": "",
      "config_id": "",
      "is_active": true
    }
  ]
}
```

### 3. 运行管理器

```bash
# 启动多账号管理器
python multi-account-manager.py
```

## 🔄 工作流程

```
1. 读取 accounts.json
   ↓
2. 选择使用次数最少的账号
   ↓
3. 无头浏览器自动登录
   ↓
4. 检测到验证码 → 等待用户输入
   ↓
5. 登录成功 → 捕获 Token + Config ID
   ↓
6. 更新 accounts.json
   ↓
7. 部署到 Docker (自动重启)
   ↓
8. 后台监控 (5分钟检查一次)
   ↓
9. Token快过期 → 自动轮换下一个账号
```

## 📱 验证码输入方式

### 方式1：命令行（默认）
```
🚨 需要验证码: 请输入6位验证码
请输入验证码: 123456
✅ 验证成功
```

### 方式2：HTTP API（可选）
```bash
# 启动Web界面后
curl -X POST http://localhost:8081/captcha \
  -H "Content-Type: application/json" \
  -d '{"code": "123456"}'
```

## 🎲 账号轮训策略

### 智能轮换
- **优先级**: 使用次数最少的账号优先
- **过期检测**: 50分钟后检查Token状态
- **自动切换**: 无缝切换到下一个账号

### 轮换示例
```
时间 00:00: 使用账号 user1@example.com
时间 50:00: Token即将过期，切换到 user2@example.com
时间 50:01: Docker服务已更新，使用 user2@example.com
时间 100:00: 切换到 user3@example.com
时间 150:00: 切换回 user1@example.com (循环)
```

## 🛠️ 高级配置

### 修改检查间隔
```python
# 在 TokenMonitor 类中
self.check_interval = 300  # 5分钟 (默认)
self.check_interval = 600  # 10分钟
```

### 启用Web管理界面
```python
# 在 main() 函数中取消注释
await web_interface.start_server()
```

### 调试模式（显示浏览器）
```python
# 在 GeminiBrowser 类中
self.browser = await playwright.chromium.launch(
    headless=False,  # 显示浏览器窗口
    ...
)
```

## 📊 状态查看

### 查看账号状态
```bash
# 查看 accounts.json
cat accounts.json

# 查看Docker日志
docker logs -f gemini-proxy

# 查看当前使用的账号
docker exec gemini-proxy env | grep BEARER_TOKEN
```

### 监控运行状态
```bash
# 查看进程
ps aux | grep multi-account-manager

# 查看端口
netstat -tlnp | grep 8080
```

## 🔧 故障排除

### 1. 浏览器启动失败
```bash
# 重新安装Playwright
playwright install chromium
playwright install-deps
```

### 2. 验证码无法显示
```bash
# 切换到有头模式调试
# 修改 multi-account-manager.py
headless=False
```

### 3. Token获取失败
```bash
# 检查网络连接
curl https://business.gemini.google

# 检查浏览器版本
playwright --version
```

### 4. Docker部署失败
```bash
# 手动测试Docker命令
docker run --rm hello-world

# 检查端口占用
netstat -tlnp | grep 8080
```

## 📦 生产部署

### 使用Docker运行管理器
```dockerfile
FROM python:3.11-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
RUN playwright install chromium

COPY . .
CMD ["python", "multi-account-manager.py"]
```

### 使用systemd管理
```ini
[Unit]
Description=Gemini Multi Account Manager
After=network.target

[Service]
Type=simple
User=youruser
WorkingDirectory=/path/to/project
ExecStart=/usr/bin/python3 /path/to/project/multi-account-manager.py
Restart=always
RestartSec=60

[Install]
WantedBy=multi-user.target
```

## 🔒 安全建议

1. **账号隔离**: 不同账号使用不同邮箱
2. **Token保护**: accounts.json设置600权限
3. **日志清理**: 定期清理日志文件
4. **网络隔离**: 使用代理IP避免封禁

## 📈 性能优化

### 并发登录
```python
# 同时登录多个账号
async def batch_login(accounts):
    tasks = [login_account(acc) for acc in accounts]
    return await asyncio.gather(*tasks)
```

### Token缓存
```python
# 缓存有效Token
valid_tokens = [acc for acc in accounts if acc.bearer_token]
```

## 🎯 使用场景

### 场景1：个人多账号
- 2-3个账号轮换
- 自动维护Token有效性
- 24小时不间断服务

### 场景2：团队共享
- 多个团队成员账号
- 按使用量分配
- 自动负载均衡

### 场景3：生产环境
- 5+账号池
- 监控告警
- 自动故障转移

---

**提示**: 首次使用建议在有头模式下调试，确认流程正常后再切换到无头模式。