#!/bin/bash

# 🤖 Gemini 多账号管理器安装脚本

echo "=================================="
echo "🤖 Gemini 多账号管理器安装"
echo "=================================="

# 检查Python
if ! command -v python3 &> /dev/null; then
    echo "❌ 未找到Python3，请先安装Python 3.8+"
    exit 1
fi

# 检查Docker
if ! command -v docker &> /dev/null; then
    echo "❌ 未找到Docker，请先安装Docker"
    exit 1
fi

# 检查pip
if ! command -v pip3 &> /dev/null; then
    echo "❌ 未找到pip3，尝试安装..."
    apt-get update && apt-get install -y python3-pip
fi

echo "✅ 环境检查通过"

# 安装Python依赖
echo "📦 安装Python依赖..."
pip3 install -r requirements.txt

# 安装Playwright浏览器
echo "🌐 安装Playwright浏览器..."
playwright install chromium
playwright install-deps

# 创建账号配置
if [ ! -f "accounts.json" ]; then
    echo "📝 创建账号配置文件..."
    cp accounts.example.json accounts.json
    echo "⚠️  请编辑 accounts.json 填入您的Gemini账号信息"
else
    echo "✅ accounts.json 已存在，跳过创建"
fi

# 构建Golang代理镜像（如果还没有）
echo "🔨 构建Golang代理镜像..."
docker build -t gemini-proxy:latest .

echo ""
echo "=================================="
echo "✅ 安装完成！"
echo "=================================="
echo ""
echo "下一步："
echo "1. 编辑 accounts.json 填入账号信息"
echo "2. 运行: python3 multi-account-manager.py"
echo ""
echo "常用命令："
echo "  python3 multi-account-manager.py    # 启动管理器"
echo "  docker logs -f gemini-proxy         # 查看服务日志"
echo "  docker stop gemini-proxy            # 停止服务"
echo "  docker restart gemini-proxy         # 重启服务"
echo ""
echo "文档："
echo "  README_MULTI_ACCOUNT.md             # 多账号管理文档"
echo "  README.md                           # 基础使用文档"