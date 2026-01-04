# Gemini Business 浏览器 Docker 项目

这是一个Docker项目，使用无头浏览器访问 https://business.gemini.google/。

## 功能特性

- 提供Web界面控制无头浏览器
- 访问 Gemini Business 页面
- 获取页面截图
- 提取页面内容
- 健康检查接口

## 项目结构

```
browser-docker-project/
├── Dockerfile              # Docker构建文件
├── package.json           # Node.js依赖配置
├── server.js              # 主应用文件
├── README.md              # 项目说明
└── run.sh                 # 启动脚本（可选）
```

## 快速开始

### 1. 构建Docker镜像

```bash
docker build -t gemini-browser .
```

### 2. 运行Docker容器

```bash
docker run -p 3000:3000 --name gemini-browser-container gemini-browser
```

### 3. 访问Web界面

打开浏览器访问：http://localhost:3000

## API端点

- `GET /` - Web控制界面
- `GET /visit` - 访问Gemini Business页面
- `GET /screenshot` - 获取页面截图
- `GET /content` - 获取页面内容
- `GET /health` - 健康检查

## 环境变量

- `PORT` - 服务器端口（默认：3000）
- `PUPPETEER_EXECUTABLE_PATH` - Chromium可执行文件路径

## 技术栈

- **Node.js 18** - 运行时环境
- **Express** - Web服务器框架
- **Puppeteer** - 无头浏览器控制
- **Chromium** - 浏览器引擎

## 注意事项

1. 由于网络限制，访问Gemini Business页面可能需要合适的网络环境
2. 首次启动可能需要一些时间下载和安装依赖
3. Docker容器需要适当的权限运行Chromium

## 构建选项

### 使用Docker Compose

创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'
services:
  gemini-browser:
    build: .
    ports:
      - "3000:3000"
    environment:
      - PORT=3000
    restart: unless-stopped
```

然后运行：
```bash
docker-compose up -d
```

## 故障排除

### 常见问题

1. **Chromium启动失败**
   - 确保Docker容器有足够的权限
   - 检查Chromium是否正确安装

2. **网络访问问题**
   - 检查网络连接
   - 验证是否可以访问 https://business.gemini.google/

3. **内存不足**
   - 增加Docker容器的内存限制
   - 使用 `--memory` 参数限制内存使用

## 许可证

MIT License
