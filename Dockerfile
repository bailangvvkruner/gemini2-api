# Debian Slim - 最可靠的方案（推荐）
FROM python:3.11-slim

WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    chromium \
    chromium-driver \
    libnss3 \
    libfreetype6 \
    libharfbuzz0b \
    ca-certificates \
    fonts-freefont-ttf \
    && rm -rf /var/lib/apt/lists/*

# 复制依赖文件
COPY requirements.txt .

# 安装Python依赖（Playwright 官方支持 Debian）
RUN pip install --no-cache-dir -r requirements.txt

# 安装Playwright浏览器
RUN playwright install chromium

# 复制应用文件
COPY multi-account-manager.py .
COPY accounts.example.json .
COPY setup.sh .
COPY README_MULTI_ACCOUNT.md .

# 设置权限
RUN chmod +x setup.sh
RUN mkdir -p /app/config

# 环境变量
ENV PYTHONUNBUFFERED=1
ENV TZ=Asia/Shanghai

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import os; os.system('ps aux | grep multi-account-manager > /dev/null')"

CMD ["python", "multi-account-manager.py"]