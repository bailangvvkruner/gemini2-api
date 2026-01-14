# Alpine 3.19 极简版 - 最小依赖
FROM alpine:3.19

WORKDIR /app

# 安装最小系统依赖
RUN apk add --no-cache \
    python3 \
    py3-pip \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ttf-freefont \
    ca-certificates \
    && rm -rf /var/cache/apk/*

# 创建软链接
RUN ln -sf python3 /usr/bin/python

# 复制并安装Python依赖
COPY requirements.txt .
RUN pip install --no-cache-dir --break-system-packages -r requirements.txt

# 安装Playwright浏览器
RUN playwright install chromium

# 复制应用
COPY multi-account-manager.py .
COPY accounts.example.json .
COPY setup.sh .
COPY README_MULTI_ACCOUNT.md .

RUN chmod +x setup.sh
RUN mkdir -p /app/config

# 环境变量
ENV PYTHONUNBUFFERED=1
ENV TZ=Asia/Shanghai

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import os; os.system('ps aux | grep multi-account-manager > /dev/null')"

CMD ["python", "multi-account-manager.py"]