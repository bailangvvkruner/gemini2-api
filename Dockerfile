# 使用官方Python 3.11镜像（Debian版本，更好的Playwright兼容性）
FROM python:3.11-slim as builder

# 设置工作目录
WORKDIR /app

# 安装系统依赖（构建阶段）
RUN apt-get update && apt-get install -y \
    chromium \
    chromium-driver \
    nss \
    libfreetype6 \
    libfreetype6-dev \
    libharfbuzz-dev \
    ca-certificates \
    fonts-freefont-ttf \
    curl \
    && rm -rf /var/lib/apt/lists/*

# 复制依赖文件（利用Docker缓存）
COPY requirements.txt .

# 安装Python依赖
RUN pip install --no-cache-dir -r requirements.txt

# 安装Playwright浏览器
RUN playwright install chromium

# 最终阶段
FROM python:3.11-slim

# 设置工作目录
WORKDIR /app

# 复制系统依赖（最小化安装）
RUN apt-get update && apt-get install -y \
    chromium \
    nss \
    libfreetype6 \
    libharfbuzz0 \
    ca-certificates \
    fonts-freefont-ttf \
    && rm -rf /var/lib/apt/lists/*

# 从构建阶段复制已安装的Python包
COPY --from=builder /usr/local/lib/python3.11/site-packages /usr/local/lib/python3.11/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin

# 复制应用文件
COPY multi-account-manager.py .
COPY accounts.example.json .
COPY setup.sh .
COPY README_MULTI_ACCOUNT.md .

# 设置环境变量
ENV PYTHONUNBUFFERED=1
ENV TZ=Asia/Shanghai

# 创建配置目录
RUN mkdir -p /app/config

# 设置权限
RUN chmod +x setup.sh

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import os; os.system('ps aux | grep multi-account-manager > /dev/null')"

# 默认命令
CMD ["python", "multi-account-manager.py"]