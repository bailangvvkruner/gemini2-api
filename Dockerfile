# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -o gemini-proxy ./cmd/server

# 运行阶段
FROM alpine:latest

# 安装CA证书
RUN apk --no-cache add ca-certificates

# 创建非root用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

# 从构建阶段复制二进制文件
COPY --from=builder /app/gemini-proxy .

# 更改文件所有权
RUN chown -R appuser:appgroup /home/appuser

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 运行应用程序
CMD ["./gemini-proxy"]
