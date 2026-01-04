#!/bin/bash

# Gemini Business浏览器Docker项目启动脚本

echo "=== Gemini Business 浏览器 Docker 项目 ==="
echo ""

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "错误: Docker未安装。请先安装Docker。"
    exit 1
fi

echo "1. 构建Docker镜像..."
docker build -t gemini-browser .

if [ $? -ne 0 ]; then
    echo "构建失败！"
    exit 1
fi

echo ""
echo "2. 运行Docker容器..."
echo "容器将在后台运行，端口映射: 3000:3000"
echo "使用 'docker logs gemini-browser-container' 查看日志"
echo "使用 'docker stop gemini-browser-container' 停止容器"
echo ""

docker run -d \
  -p 3000:3000 \
  --name gemini-browser-container \
  gemini-browser

if [ $? -ne 0 ]; then
    echo "容器启动失败！"
    exit 1
fi

echo ""
echo "3. 容器状态..."
sleep 2
docker ps | grep gemini-browser-container

echo ""
echo "4. 访问应用..."
echo "打开浏览器访问: http://localhost:3000"
echo ""
echo "5. 查看日志:"
echo "   docker logs -f gemini-browser-container"
echo ""
echo "6. 停止容器:"
echo "   docker stop gemini-browser-container"
echo ""
echo "7. 删除容器:"
echo "   docker rm gemini-browser-container"
echo ""
echo "8. 删除镜像:"
echo "   docker rmi gemini-browser"
echo ""
echo "启动完成！"
