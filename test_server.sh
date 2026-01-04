#!/bin/bash

# 简单的测试脚本，用于验证项目构建和基本功能

echo "=== Gemini Business Proxy 测试 ==="
echo

# 1. 检查Go版本
echo "1. 检查Go版本:"
go version
echo

# 2. 检查项目构建
echo "2. 检查项目构建:"
if go build ./cmd/server; then
    echo "✓ 项目构建成功"
else
    echo "✗ 项目构建失败"
    exit 1
fi
echo

# 3. 检查依赖
echo "3. 检查依赖:"
go mod verify
echo "✓ 依赖检查完成"
echo

# 4. 检查代码格式
echo "4. 检查代码格式:"
go fmt ./...
echo "✓ 代码格式化完成"
echo

# 5. 静态分析
echo "5. 静态分析:"
if go vet ./...; then
    echo "✓ 静态分析通过"
else
    echo "⚠ 静态分析发现一些问题"
fi
echo

# 6. 创建测试配置
echo "6. 创建测试配置:"
cat > test_config.env << EOF
SERVER_ADDRESS=127.0.0.1
SERVER_PORT=18080
GEMINI_BUSINESS_URL=https://business.gemini.google
API_BASE_URL=https://biz-discoveryengine.googleapis.com/v1alpha
AUTH_URL=https://auth.business.gemini.google
ACCOUNT_EMAIL=test@example.com
SESSION_TIMEOUT=300
CONFIG_ID=d06739ca-6683-46db-bb51-07395a392439
OPENAI_COMPATIBLE=true
API_KEY_HEADER=Authorization
DEFAULT_MODEL=gemini-business
LOG_LEVEL=info
LOG_FORMAT=json
EOF
echo "✓ 测试配置文件已创建: test_config.env"
echo

# 7. 显示项目结构
echo "7. 项目文件结构:"
find . -type f -name "*.go" | head -20
echo "..."
echo

echo "=== 测试完成 ==="
echo
echo "下一步:"
echo "1. 设置真实的 ACCOUNT_EMAIL 环境变量"
echo "2. 运行: go run ./cmd/server"
echo "3. 或者使用Docker: docker build -t gemini-business-proxy . && docker run -p 8080:8080 --env-file .env gemini-business-proxy"
