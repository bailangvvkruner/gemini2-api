# 使用Alpine Linux减小镜像大小
FROM node:18-alpine

# 安装必要的依赖，包括Chromium和必要的库
RUN apk update && apk add --no-cache \
    chromium \
    nss \
    freetype \
    freetype-dev \
    harfbuzz \
    ca-certificates \
    ttf-freefont \
    font-noto-emoji \
    && rm -rf /var/cache/apk/*

# 设置环境变量，让Puppeteer使用已安装的Chromium
ENV PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true \
    PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium

# 创建工作目录
WORKDIR /app

# 复制package.json文件
COPY package*.json ./

# 安装依赖
RUN npm install

# 复制应用代码
COPY . .

# 暴露端口（如果需要Web服务器）
EXPOSE 3000

# 启动应用
CMD ["node", "server.js"]
