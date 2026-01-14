#!/usr/bin/env python3
"""
轻量级API代理服务（Python版本）
提供OpenAI格式的API接口，直接调用Gemini Business API
"""

import asyncio
import json
import os
import time
import aiohttp
from aiohttp import web
from datetime import datetime
from typing import Dict, List, Optional
import logging

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)

logger = logging.getLogger(__name__)

# 环境变量配置
class Config:
    BEARER_TOKEN = os.getenv('BEARER_TOKEN', '')
    CONFIG_ID = os.getenv('CONFIG_ID', '')
    PORT = int(os.getenv('PORT', '8080'))
    DEBUG = os.getenv('DEBUG', 'false').lower() == 'true'
    PROXY_URL = os.getenv('PROXY_URL', '')

config = Config()

# OpenAI格式请求模型
class OpenAIRequest:
    def __init__(self, data: Dict):
        self.model = data.get('model', 'gemini-2.5-flash')
        self.messages = data.get('messages', [])
        self.stream = data.get('stream', False)
        self.temperature = data.get('temperature', 0.7)
        self.max_tokens = data.get('max_tokens')
        self.user = data.get('user')

# Gemini API请求模型
class GeminiRequest:
    def __init__(self, openai_req: OpenAIRequest):
        self.model = openai_req.model
        self.messages = openai_req.messages
        self.temperature = openai_req.temperature
        self.max_tokens = openai_req.max_tokens

# API代理服务
class APIProxy:
    def __init__(self):
        self.base_url = "https://biz-discoveryengine.googleapis.com/v1alpha/locations/global/widgetStreamAssist"
        self.session = None
    
    async def get_session(self):
        """获取HTTP会话"""
        if not self.session:
            connector = aiohttp.TCPConnector(ssl=False)
            if config.PROXY_URL:
                connector = aiohttp.TCPConnector(ssl=False, proxy=config.PROXY_URL)
            self.session = aiohttp.ClientSession(connector=connector)
        return self.session
    
    def build_gemini_payload(self, gemini_req: GeminiRequest) -> Dict:
        """构建Gemini API请求体"""
        # 转换消息格式
        contents = []
        for msg in gemini_req.messages:
            role = msg.get('role', 'user')
            content = msg.get('content', '')
            
            if role == 'system':
                # 系统提示词作为配置
                continue
            elif role == 'user':
                contents.append({
                    "role": "user",
                    "parts": [{"text": content}]
                })
            elif role == 'assistant':
                contents.append({
                    "role": "model",
                    "parts": [{"text": content}]
                })
        
        payload = {
            "contents": contents,
            "generationConfig": {
                "temperature": gemini_req.temperature,
            }
        }
        
        if gemini_req.max_tokens:
            payload["generationConfig"]["maxOutputTokens"] = gemini_req.max_tokens
        
        return payload
    
    async def send_request(self, payload: Dict, stream: bool = False) -> aiohttp.ClientResponse:
        """发送请求到Gemini API"""
        session = await self.get_session()
        
        headers = {
            "Authorization": f"Bearer {config.BEARER_TOKEN}",
            "Content-Type": "application/json",
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
        }
        
        # 添加Config ID到URL
        url = f"{self.base_url}?configId={config.CONFIG_ID}"
        if stream:
            url += "&stream=true"
        
        if config.DEBUG:
            logger.info(f"发送请求到: {url}")
            logger.info(f"请求体: {json.dumps(payload, ensure_ascii=False)}")
        
        try:
            response = await session.post(
                url,
                json=payload,
                headers=headers,
                timeout=aiohttp.ClientTimeout(total=60)
            )
            return response
        except Exception as e:
            logger.error(f"请求失败: {e}")
            raise
    
    async def handle_normal_chat(self, openai_req: OpenAIRequest) -> Dict:
        """处理非流式响应"""
        gemini_req = GeminiRequest(openai_req)
        payload = self.build_gemini_payload(gemini_req)
        
        response = await self.send_request(payload, stream=False)
        
        if response.status != 200:
            error_text = await response.text()
            logger.error(f"Gemini API错误: {response.status} - {error_text}")
            raise web.HTTPException(status_code=response.status, text=error_text)
        
        data = await response.json()
        
        # 转换为OpenAI格式
        return self._convert_to_openai_format(data, openai_req.model)
    
    async def handle_stream_chat(self, request: web.Request, openai_req: OpenAIRequest):
        """处理流式响应"""
        gemini_req = GeminiRequest(openai_req)
        payload = self.build_gemini_payload(gemini_req)
        
        response = await self.send_request(payload, stream=True)
        
        if response.status != 200:
            error_text = await response.text()
            logger.error(f"Gemini API错误: {response.status} - {error_text}")
            raise web.HTTPException(status_code=response.status, text=error_text)
        
        # 设置响应头
        headers = {
            'Content-Type': 'text/event-stream',
            'Cache-Control': 'no-cache',
            'Connection': 'keep-alive',
        }
        
        # 创建流式响应
        async def generate():
            buffer = ""
            async for line in response.content:
                chunk = line.decode('utf-8')
                buffer += chunk
                
                # 按JSON块分割
                while '\n' in buffer:
                    line, buffer = buffer.split('\n', 1)
                    line = line.strip()
                    if not line:
                        continue
                    
                    try:
                        # 尝试解析JSON
                        data = json.loads(line)
                        
                        # 转换为OpenAI格式
                        openai_chunk = self._convert_to_stream_chunk(data, openai_req.model)
                        if openai_chunk:
                            yield f"data: {json.dumps(openai_chunk, ensure_ascii=False)}\n\n"
                    except json.JSONDecodeError:
                        continue
            
            yield "data: [DONE]\n\n"
        
        return web.Response(
            body=generate(),
            headers=headers,
            status=200
        )
    
    def _convert_to_openai_format(self, gemini_data: Dict, model: str) -> Dict:
        """转换Gemini响应为OpenAI格式"""
        if not gemini_data or 'candidates' not in gemini_data:
            return {
                "id": f"chatcmpl-{int(time.time())}",
                "object": "chat.completion",
                "created": int(time.time()),
                "model": model,
                "choices": [{
                    "index": 0,
                    "message": {
                        "role": "assistant",
                        "content": "No response from Gemini"
                    },
                    "finish_reason": "stop"
                }],
                "usage": {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}
            }
        
        candidate = gemini_data['candidates'][0]
        content = candidate.get('content', {})
        parts = content.get('parts', [])
        text = parts[0].get('text', '') if parts else ""
        
        return {
            "id": f"chatcmpl-{int(time.time())}",
            "object": "chat.completion",
            "created": int(time.time()),
            "model": model,
            "choices": [{
                "index": 0,
                "message": {
                    "role": "assistant",
                    "content": text
                },
                "finish_reason": candidate.get('finishReason', 'stop')
            }],
            "usage": {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}
        }
    
    def _convert_to_stream_chunk(self, gemini_data: Dict, model: str) -> Optional[Dict]:
        """转换Gemini流式块为OpenAI格式"""
        if not gemini_data or 'candidates' not in gemini_data:
            return None
        
        candidate = gemini_data['candidates'][0]
        content = candidate.get('content', {})
        parts = content.get('parts', [])
        
        if not parts:
            return None
        
        text = parts[0].get('text', '')
        
        return {
            "id": f"chatcmpl-{int(time.time())}",
            "object": "chat.completion.chunk",
            "created": int(time.time()),
            "model": model,
            "choices": [{
                "index": 0,
                "delta": {
                    "role": "assistant",
                    "content": text
                },
                "finish_reason": candidate.get('finishReason')
            }]
        }
    
    async def close(self):
        """关闭会话"""
        if self.session:
            await self.session.close()

# Web服务器
async def create_app():
    """创建Web应用"""
    app = web.Application()
    proxy = APIProxy()
    
    async def health_check(request):
        """健康检查"""
        return web.json_response({
            "status": "ok",
            "timestamp": datetime.now().isoformat(),
            "config": {
                "has_token": bool(config.BEARER_TOKEN),
                "has_config_id": bool(config.CONFIG_ID),
                "debug": config.DEBUG
            }
        })
    
    async def list_models(request):
        """列出模型"""
        models = [
            {"id": "gemini-2.5-flash", "object": "model", "created": 0, "owned_by": "google"},
            {"id": "gemini-2.5-pro", "object": "model", "created": 0, "owned_by": "google"},
            {"id": "gemini-3-flash-preview", "object": "model", "created": 0, "owned_by": "google"},
            {"id": "gemini-3-pro-preview", "object": "model", "created": 0, "owned_by": "google"}
        ]
        return web.json_response({"object": "list", "data": models})
    
    async def chat_completions(request):
        """聊天完成端点"""
        start_time = time.time()
        
        try:
            data = await request.json()
            
            if config.DEBUG:
                logger.info(f"收到请求: {json.dumps(data, ensure_ascii=False)}")
            
            openai_req = OpenAIRequest(data)
            
            # 验证配置
            if not config.BEARER_TOKEN:
                raise web.HTTPUnauthorized(text="Missing BEARER_TOKEN environment variable")
            if not config.CONFIG_ID:
                raise web.HTTPUnauthorized(text="Missing CONFIG_ID environment variable")
            
            # 验证模型
            valid_models = ["gemini-2.5-flash", "gemini-2.5-pro", "gemini-3-flash-preview", "gemini-3-pro-preview"]
            if openai_req.model not in valid_models:
                raise web.HTTPBadRequest(text=f"Invalid model. Valid models: {valid_models}")
            
            # 处理请求
            if openai_req.stream:
                return await proxy.handle_stream_chat(request, openai_req)
            else:
                result = await proxy.handle_normal_chat(openai_req)
                duration = time.time() - start_time
                logger.info(f"请求完成: model={openai_req.model}, duration={duration:.2f}s")
                return web.json_response(result)
        
        except json.JSONDecodeError:
            raise web.HTTPBadRequest(text="Invalid JSON")
        except Exception as e:
            logger.error(f"处理请求失败: {e}")
            raise web.HTTPException(status_code=500, text=str(e))
    
    # 注册路由
    app.router.add_get('/health', health_check)
    app.router.add_get('/v1/models', list_models)
    app.router.add_post('/v1/chat/completions', chat_completions)
    
    return app

async def main():
    """主函数"""
    logger.info("=" * 60)
    logger.info("🤖 Gemini Business API 代理服务 (Python)")
    logger.info("=" * 60)
    
    # 验证配置
    if not config.BEARER_TOKEN:
        logger.error("❌ 未设置 BEARER_TOKEN 环境变量")
        return
    
    if not config.CONFIG_ID:
        logger.error("❌ 未设置 CONFIG_ID 环境变量")
        return
    
    logger.info(f"✅ 配置验证通过")
    logger.info(f"   Config ID: {config.CONFIG_ID}")
    logger.info(f"   Debug: {config.DEBUG}")
    if config.PROXY_URL:
        logger.info(f"   Proxy: {config.PROXY_URL}")
    
    # 启动Web服务器
    app = await create_app()
    
    logger.info(f"🚀 服务启动: http://0.0.0.0:{config.PORT}")
    logger.info(f"   健康检查: http://localhost:{config.PORT}/health")
    logger.info(f"   API端点: http://localhost:{config.PORT}/v1/chat/completions")
    logger.info("")
    logger.info("等待请求...")
    
    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, '0.0.0.0', config.PORT)
    await site.start()
    
    # 保持运行
    try:
        await asyncio.Event().wait()
    except KeyboardInterrupt:
        logger.info("\n👋 服务正在关闭...")
    finally:
        await runner.cleanup()

if __name__ == "__main__":
    asyncio.run(main())