#!/bin/bash

# 最简化的 Docker 运行命令（一行命令版本）
# 复制下面整段命令到终端执行，替换 YOUR_BEARER_TOKEN 和 YOUR_CONFIG_ID

docker run -d \
  --name gemini-proxy \
  --restart unless-stopped \
  -p 8080:8080 \
  -e BEARER_TOKEN="YOUR_BEARER_TOKEN_HERE" \
  -e CONFIG_ID="YOUR_CONFIG_ID_HERE" \
  -e PORT="8080" \
  -e DEBUG="false" \
  -e TZ=Asia/Shanghai \
  ghcr.io/yourusername/gemini-proxy:latest

# 使用说明：
# 1. 复制上面的命令
# 2. 替换 YOUR_BEARER_TOKEN 为你的 Bearer Token
# 3. 替换 YOUR_CONFIG_ID 为你的 Config ID
# 4. 粘贴到终端执行

# 查看日志：
# docker logs -f gemini-proxy

# 测试 API：
# curl -X POST http://localhost:8080/v1/chat/completions -H "Content-Type: application/json" -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"你好"}],"stream":true}'