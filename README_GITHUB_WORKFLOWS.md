# 🤖 GitHub Actions 工作流说明

本项目包含完整的 CI/CD 工作流，实现自动化构建、测试、部署和监控。

## 📋 工作流概览

| 工作流 | 触发条件 | 作用 |
|--------|----------|------|
| `build.yml` | 推送到 main/master | 构建并推送 Docker 镜像 |
| `test.yml` | 推送/PR | 运行测试和验证 |
| `deploy.yml` | 手动触发 | 部署到生产服务器 |
| `release.yml` | 创建 Git Tag | 创建发布版本 |
| `auto-update.yml` | 每周一 | 自动更新依赖 |
| `security-scan.yml` | 推送/PR/每天 | 安全扫描 |
| `monitor.yml` | 每6小时 | 健康检查监控 |

## 🔧 详细说明

### 1. Build & Push (`build.yml`)

**触发条件：**
- 推送到 `main` 或 `master` 分支
- 创建 Pull Request

**作用：**
- ✅ 构建 Docker 镜像
- ✅ 推送到 GitHub Container Registry
- ✅ 使用缓存加速构建
- ✅ 生成多标签（latest, 分支名, commit SHA）

**使用：**
```bash
# 镜像地址
ghcr.io/yourusername/gemini-proxy:latest
ghcr.io/yourusername/gemini-proxy:main
ghcr.io/yourusername/gemini-proxy:main-abc123
```

### 2. Test (`test.yml`)

**触发条件：**
- 推送到 `main` 或 `master` 分支
- 创建 Pull Request

**测试内容：**
- ✅ Python 语法检查
- ✅ 依赖导入测试
- ✅ Docker 构建测试
- ✅ 容器运行测试
- ✅ API 端点测试
- ✅ 配置文件格式验证

**测试报告：**
- 失败时自动显示详细日志
- 通过 GitHub Actions UI 查看

### 3. Deploy (`deploy.yml`)

**触发条件：**
- 手动触发（GitHub UI）

**部署步骤：**
1. 拉取最新镜像
2. 停止旧容器
3. 启动新容器
4. 健康检查
5. 发送通知

**必需的 Secrets：**
```bash
SSH_HOST      # 服务器地址
SSH_USER      # SSH用户名
SSH_KEY       # SSH私钥
BEARER_TOKEN  # Gemini Token
CONFIG_ID     # Config ID
SLACK_WEBHOOK # Slack通知（可选）
```

**手动部署：**
```bash
# GitHub CLI
gh workflow run deploy.yml --ref main

# 或在 GitHub UI 点击 "Run workflow"
```

### 4. Release (`release.yml`)

**触发条件：**
- 创建 Git Tag（如 `v1.0.0`）

**作用：**
- ✅ 创建 GitHub Release
- ✅ 上传配置文件
- ✅ 生成快速开始指南
- ✅ 自动发布到 Releases 页面

**使用：**
```bash
# 创建并推送 Tag
git tag v1.0.0
git push origin v1.0.0

# 自动触发发布流程
```

### 5. Auto Update (`auto-update.yml`)

**触发条件：**
- 每周一凌晨 3 点
- 手动触发

**作用：**
- ✅ 检查 Python 依赖更新
- ✅ 自动创建 Pull Request
- ✅ 更新 Playwright 浏览器

**PR 内容：**
- 更新的依赖列表
- 测试状态
- 合并建议

### 6. Security Scan (`security-scan.yml`)

**触发条件：**
- 推送/PR
- 每天凌晨 2 点

**扫描内容：**
- ✅ Python 漏洞检查（Safety）
- ✅ 静态代码分析（Bandit）
- ✅ Docker 镜像扫描（Trivy）
- ✅ 密钥泄露检测（TruffleHog）

**结果：**
- 生成 SARIF 报告
- PR 自动评论
- 严重问题自动创建 Issue

### 7. Health Monitor (`monitor.yml`)

**触发条件：**
- 每 6 小时
- 手动触发

**监控内容：**
- ✅ 镜像可用性
- ✅ 容器健康状态
- ✅ API 响应

**告警：**
- 失败时自动创建 Issue
- Slack 通知（如果配置）

## 🔐 配置 Secrets

在 GitHub 仓库设置中添加以下 Secrets：

### 必需（部署用）
```bash
SSH_HOST        # 服务器地址
SSH_USER        # SSH用户名
SSH_KEY         # SSH私钥（用于登录服务器）
BEARER_TOKEN    # Gemini Business Token
CONFIG_ID       # Gemini Config ID
```

### 可选（通知用）
```bash
SLACK_WEBHOOK   # Slack Webhook URL
```

### 可选（高级）
```bash
GITHUB_TOKEN    # 自动提供
DOCKER_USERNAME # Docker Hub 用户名
DOCKER_PASSWORD # Docker Hub 密码
```

## 🚀 使用流程

### 1. 开发流程
```bash
# 1. 创建分支
git checkout -b feature/new-feature

# 2. 开发并提交
git add .
git commit -m "feat: add new feature"

# 3. 推送并创建 PR
git push origin feature/new-feature
# → 自动触发测试

# 4. 合并到 main
# → 自动构建镜像
```

### 2. 发布流程
```bash
# 1. 更新版本号
# 修改 README.md 中的版本号

# 2. 创建 Tag
git tag v1.1.0
git push origin v1.1.0

# 3. 自动发布
# → 创建 Release
# → 生成发布说明
```

### 3. 部署流程
```bash
# 方法1：手动触发
# GitHub UI → Actions → Deploy → Run workflow

# 方法2：CLI
gh workflow run deploy.yml --ref main

# 方法3：自动部署（如果配置了自动触发）
# 合并到 main 后自动部署
```

## 📊 监控和日志

### 查看工作流状态
```bash
# GitHub CLI
gh run list
gh run view <run_id>

# 查看详细日志
gh run view <run_id> --log
```

### 查看镜像
```bash
# 查看所有镜像
ghcr.io/yourusername/gemini-proxy

# 查看标签
ghcr.io/yourusername/gemini-proxy:latest
ghcr.io/yourusername/gemini-proxy:v1.0.0
```

### 查看发布
```bash
# GitHub CLI
gh release list
gh release view v1.0.0
```

## 🎯 最佳实践

### 1. 分支管理
- `main`：生产环境
- `develop`：开发环境
- `feature/*`：功能开发
- `hotfix/*`：紧急修复

### 2. 标签规范
- `v1.0.0`：正式版本
- `v1.0.1-rc1`：预发布版本
- `v1.0.0-beta`：测试版本

### 3. 提交信息
```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式
refactor: 重构
test: 测试相关
chore: 构建/工具相关
```

### 4. Secrets 管理
- 定期轮换 Token
- 使用 GitHub Environments
- 最小权限原则

## 🔍 故障排除

### 构建失败
```bash
# 查看详细日志
gh run view <run_id> --log

# 检查 Dockerfile
docker build . --no-cache
```

### 部署失败
```bash
# 检查服务器连接
ssh $SSH_USER@$SSH_HOST

# 查看容器日志
docker logs gemini-proxy

# 手动部署测试
docker-compose up -d
```

### 测试失败
```bash
# 本地运行测试
python -m pytest
docker build -t test .
docker run --rm test python -c "import api-proxy"
```

## 📈 性能优化

### 缓存策略
- Docker layer caching
- GitHub Actions cache
- 依赖缓存

### 并行执行
- 测试并行运行
- 多架构构建
- 矩阵策略

### 增量构建
- 只构建变更的组件
- 使用 .dockerignore
- 最小化镜像层

## 🚨 紧急回滚

### 快速回滚
```bash
# 1. 找到上一个稳定版本
gh release list

# 2. 部署旧版本
docker pull ghcr.io/yourusername/gemini-proxy:v0.9.0
docker tag ghcr.io/yourusername/gemini-proxy:v0.9.0 ghcr.io/yourusername/gemini-proxy:latest
docker-compose up -d

# 3. 创建 hotfix 分支
git checkout -b hotfix/rollback
```

### 回滚脚本
```bash
#!/bin/bash
# rollback.sh

OLD_VERSION=$1
if [ -z "$OLD_VERSION" ]; then
  echo "Usage: ./rollback.sh <version>"
  exit 1
fi

echo "🔄 回滚到版本 $OLD_VERSION"

# 拉取旧版本
docker pull ghcr.io/yourusername/gemini-proxy:$OLD_VERSION

# 更新标签
docker tag ghcr.io/yourusername/gemini-proxy:$OLD_VERSION ghcr.io/yourusername/gemini-proxy:latest

# 重启服务
docker-compose down
docker-compose up -d

echo "✅ 回滚完成"
```

---

**提示：** 所有工作流都经过测试，可以直接使用。首次使用前请配置必要的 Secrets。