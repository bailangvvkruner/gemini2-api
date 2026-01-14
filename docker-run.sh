#!/bin/bash

# Gemini Business API 代理服务 - Docker 运行脚本
# 使用说明: 
# 1. 替换 YOUR_BEARER_TOKEN 和 YOUR_CONFIG_ID 为实际值
# 2. 赋予执行权限: chmod +x docker-run.sh
# 3. 运行: ./docker-run.sh

# 必须配置的环境变量
export BEARER_TOKEN="YOUR_BEARER_TOKEN_HERE"
export CONFIG_ID="YOUR_CONFIG_ID_HERE"

# 可选配置的环境变量
export PORT="8080"
export DEBUG="false"
export PROXY_URL=""

# 检查必须的环境变量
if [ "$BEARER_TOKEN" = "YOUR_BEARER_TOKEN_HERE" ] || [ -z "$BEARER_TOKEN" ]; then
    echo "❌ 错误: 请设置 BEARER_TOKEN 环境变量"
    echo "从浏览器开发者工具获取 Authorization Bearer Token"
    exit 1
fi

if [ "$CONFIG_ID" = "YOUR_CONFIG_ID_HERE" ] || [ -z "$CONFIG_ID" ]; then
    echo "❌ 错误: 请设置 CONFIG_ID 环境变量"
    echo "从URL中获取，如: https://business.gemini.google/home/cid/CONFIG_ID/..."
    exit 1
fi

echo "🚀 启动 Gemini Business API 代理服务..."
echo "📊 端口: $PORT"
echo "🔧 调试模式: $DEBUG"
echo ""

# 停止并删除旧容器（如果存在）
docker stop gemini-proxy 2>/dev/null
docker rm gemini-proxy 2>/dev/null

# 拉取最新镜像（可选）
# docker pull ghcr.io/yourusername/gemini-proxy:latest

# 运行容器
docker run -d \
  --name gemini-proxy \
  --restart unless-stopped \
  -p $PORT:8080 \
  -e BEARER_TOKEN="$BEARER_TOKEN" \
  -e CONFIG_ID="$CONFIG_ID" \
  -e PORT="$PORT" \
  -e DEBUG="$DEBUG" \
  -e PROXY_URL="$PROXY_URL" \
  -e TZ=Asia/Shanghai \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  ghcr.io/yourusername/gemini-proxy:latest

if [ $? -eq 0 ]; then
    echo "✅ 容器启动成功！"
    echo ""
    echo "查看日志: docker logs -f gemini-proxy"
    echo "停止服务: docker stop gemini-proxy"
    echo "重启服务: docker restart gemini-proxy"
    echo ""
    echo "测试命令:"
    echo "curl -X POST http://localhost:$PORT/v1/chat/completions -H \"Content-Type: application/json\" -d '{\"model\":\"gemini-2.5-flash\",\"messages\":[{\"role\":\"user\",\"content\":\"你好\"}],\"stream\":true}'"
else
    echo "❌ 容器启动失败，请检查日志"
    docker logs gemini-proxy
fi