#!/usr/bin/env python3
"""
Gemini Business 多账号管理器
支持：无头浏览器自动登录 + 验证码后台输入 + 账号轮训
"""

import asyncio
import json
import time
import random
import os
from datetime import datetime
from playwright.async_api import async_playwright
from typing import List, Dict, Optional
import aiohttp
from dataclasses import dataclass
from pathlib import Path

@dataclass
class Account:
    """账号配置"""
    email: str
    password: str = ""
    bearer_token: str = ""
    config_id: str = ""
    last_used: int = 0
    usage_count: int = 0
    is_active: bool = True

class CaptchaSolver:
    """验证码输入接口"""
    
    def __init__(self):
        self.pending_captcha = None
        self.captcha_solved = asyncio.Event()
    
    async def wait_for_captcha(self, captcha_text: str) -> str:
        """等待用户输入验证码"""
        print(f"\n🚨 需要验证码: {captcha_text}")
        print(f"请查看邮箱并输入验证码...")
        
        # 重置事件
        self.captcha_solved.clear()
        self.pending_captcha = captcha_text
        
        # 等待用户输入（通过HTTP接口或命令行）
        code = await self._get_user_input()
        
        self.pending_captcha = None
        self.captcha_solved.set()
        return code
    
    async def _get_user_input(self) -> str:
        """获取用户输入（支持HTTP接口）"""
        # 方式1：命令行输入
        return input("请输入验证码: ")
        
        # 方式2：HTTP API输入（可选）
        # return await self._wait_http_input()

class GeminiBrowser:
    """无头浏览器管理"""
    
    def __init__(self, captcha_solver: CaptchaSolver):
        self.captcha_solver = captcha_solver
        self.browser = None
        self.context = None
    
    async def start(self):
        """启动浏览器"""
        playwright = await async_playwright().start()
        self.browser = await playwright.chromium.launch(
            headless=True,  # 可以改为False调试
            args=['--no-sandbox', '--disable-dev-shm-usage']
        )
        
        # 创建上下文，保存登录状态
        self.context = await self.browser.new_context(
            storage_state=None,  # 可以加载保存的登录状态
            viewport={'width': 1920, 'height': 1080}
        )
    
    async def login_and_get_token(self, account: Account) -> Optional[Dict]:
        """登录并获取Token"""
        page = await self.context.new_page()
        
        # 监听网络请求
        token_data = None
        
        async def intercept_request(route):
            nonlocal token_data
            request = route.request
            
            if 'widgetStreamAssist' in request.url:
                auth_header = request.headers.get('authorization', '')
                if auth_header.startswith('Bearer '):
                    token = auth_header.replace('Bearer ', '')
                    
                    # 获取Config ID从URL
                    config_id = request.url.split('/cid/')[1].split('/')[0] if '/cid/' in request.url else ""
                    
                    token_data = {
                        'bearer_token': token,
                        'config_id': config_id,
                        'email': account.email
                    }
                    print(f"✅ Token获取成功: {account.email}")
            
            await route.continue_()
        
        await page.route("**/*", intercept_request)
        
        try:
            # 打开登录页面
            print(f"🌐 打开登录页面: {account.email}")
            await page.goto('https://business.gemini.google', wait_until='networkidle')
            
            # 等待并填写邮箱
            email_input = await page.wait_for_selector('input[type="email"]', timeout=10000)
            await email_input.fill(account.email)
            await page.click('button:has-text("下一步")')
            
            # 等待密码输入（如果需要）
            try:
                password_input = await page.wait_for_selector('input[type="password"]', timeout=5000)
                if account.password:
                    await password_input.fill(account.password)
                    await page.click('button:has-text("下一步")')
            except:
                pass
            
            # 等待验证码
            while True:
                try:
                    # 检查是否有验证码
                    captcha_element = await page.query_selector('input[aria-label*="验证码"]')
                    if captcha_element:
                        # 获取验证码提示文本
                        captcha_text = await captcha_element.get_attribute('placeholder') or "请输入验证码"
                        
                        # 等待用户输入验证码
                        captcha_code = await self.captcha_solver.wait_for_captcha(captcha_text)
                        
                        await captcha_element.fill(captcha_code)
                        await page.click('button:has-text("下一步")')
                    else:
                        break
                except:
                    break
            
            # 等待登录完成，进入主页面
            await page.wait_for_url("**/home/cid/**", timeout=30000)
            
            # 触发一个API调用
            await page.wait_for_timeout(2000)
            
            # 发送测试消息
            try:
                textarea = await page.wait_for_selector('textarea', timeout=5000)
                await textarea.fill("test")
                await page.keyboard.press('Enter')
            except:
                pass
            
            # 等待Token被捕获
            await page.wait_for_timeout(5000)
            
            return token_data
            
        except Exception as e:
            print(f"❌ 登录失败 {account.email}: {e}")
            return None
        finally:
            await page.close()
    
    async def close(self):
        """关闭浏览器"""
        if self.browser:
            await self.browser.close()

class AccountManager:
    """多账号管理"""
    
    def __init__(self, config_file: str = "accounts.json"):
        self.config_file = config_file
        self.accounts: List[Account] = []
        self.current_index = 0
        self.load_accounts()
    
    def load_accounts(self):
        """加载账号配置"""
        if os.path.exists(self.config_file):
            with open(self.config_file, 'r') as f:
                data = json.load(f)
                self.accounts = [Account(**acc) for acc in data.get('accounts', [])]
        else:
            # 创建示例配置
            self.accounts = [
                Account(email="user1@example.com"),
                Account(email="user2@example.com"),
            ]
            self.save_accounts()
    
    def save_accounts(self):
        """保存账号配置"""
        data = {
            'accounts': [
                {
                    'email': acc.email,
                    'bearer_token': acc.bearer_token,
                    'config_id': acc.config_id,
                    'last_used': acc.last_used,
                    'usage_count': acc.usage_count,
                    'is_active': acc.is_active
                }
                for acc in self.accounts
            ]
        }
        with open(self.config_file, 'w') as f:
            json.dump(data, f, indent=2)
    
    def get_next_account(self) -> Optional[Account]:
        """获取下一个可用账号"""
        if not self.accounts:
            return None
        
        # 找到活跃账号
        active_accounts = [acc for acc in self.accounts if acc.is_active]
        if not active_accounts:
            return None
        
        # 按使用次数排序，选择使用最少的
        active_accounts.sort(key=lambda x: x.usage_count)
        return active_accounts[0]
    
    def update_account(self, email: str, bearer_token: str, config_id: str):
        """更新账号信息"""
        for acc in self.accounts:
            if acc.email == email:
                acc.bearer_token = bearer_token
                acc.config_id = config_id
                acc.last_used = int(time.time())
                acc.usage_count += 1
                break
        self.save_accounts()
    
    def get_healthy_accounts(self) -> List[Account]:
        """获取健康的账号（有有效Token）"""
        return [acc for acc in self.accounts if acc.bearer_token and acc.is_active]

class DockerManager:
    """Docker容器管理"""
    
    def __init__(self, container_name: str = "gemini-proxy"):
        self.container_name = container_name
    
    def deploy_account(self, account: Account):
        """部署账号到Docker"""
        # 停止旧容器
        os.system(f"docker stop {self.container_name} 2>/dev/null")
        os.system(f"docker rm {self.container_name} 2>/dev/null")
        
        # 启动新容器
        cmd = f"""
        docker run -d \
          --name {self.container_name} \
          --restart unless-stopped \
          -p 8080:8080 \
          -e BEARER_TOKEN="{account.bearer_token}" \
          -e CONFIG_ID="{account.config_id}" \
          -e TZ=Asia/Shanghai \
          ghcr.io/yourusername/gemini-proxy:latest
        """
        
        print(f"🚀 部署账号: {account.email}")
        print(f"   Config ID: {account.config_id}")
        print(f"   Token前20位: {account.bearer_token[:20]}...")
        
        os.system(cmd)
    
    def rotate_account(self, new_account: Account):
        """轮换账号"""
        print(f"\n🔄 轮换账号: {new_account.email}")
        self.deploy_account(new_account)

class TokenMonitor:
    """Token状态监控"""
    
    def __init__(self, account_manager: AccountManager, docker_manager: DockerManager):
        self.account_manager = account_manager
        self.docker_manager = docker_manager
        self.check_interval = 300  # 5分钟检查一次
    
    async def monitor(self):
        """持续监控"""
        while True:
            try:
                await self._check_and_rotate()
                await asyncio.sleep(self.check_interval)
            except Exception as e:
                print(f"监控错误: {e}")
                await asyncio.sleep(60)
    
    async def _check_and_rotate(self):
        """检查Token状态并轮换"""
        healthy_accounts = self.account_manager.get_healthy_accounts()
        
        if not healthy_accounts:
            print("⚠️  没有可用的健康账号")
            return
        
        # 检查当前账号的Token是否快过期
        current_account = self.account_manager.get_next_account()
        if not current_account:
            return
        
        # 模拟检查Token过期时间（实际可以从JWT解析）
        # 这里简单检查是否使用超过50分钟
        if current_account.last_used:
            used_time = time.time() - current_account.last_used
            if used_time > 3000:  # 50分钟
                print(f"⚠️  Token可能过期，准备轮换: {current_account.email}")
                
                # 选择下一个账号
                next_account = self._get_next_in_rotation(healthy_accounts, current_account)
                if next_account:
                    self.docker_manager.rotate_account(next_account)
    
    def _get_next_in_rotation(self, accounts: List[Account], current: Account) -> Optional[Account]:
        """获取轮换列表中的下一个账号"""
        try:
            idx = accounts.index(current)
            return accounts[(idx + 1) % len(accounts)]
        except:
            return accounts[0] if accounts else None

class WebInterface:
    """Web管理界面（可选）"""
    
    def __init__(self, account_manager: AccountManager, captcha_solver: CaptchaSolver):
        self.account_manager = account_manager
        self.captcha_solver = captcha_solver
    
    async def start_server(self):
        """启动HTTP服务用于接收验证码"""
        from aiohttp import web
        
        async def handle_captcha(request):
            """接收验证码输入"""
            data = await request.json()
            code = data.get('code')
            if code and self.captcha_solver.pending_captcha:
                self.captcha_solver.captcha_input = code
                self.captcha_solver.captcha_solved.set()
                return web.json_response({'status': 'ok'})
            return web.json_response({'status': 'error'})
        
        async def handle_status(request):
            """获取状态"""
            return web.json_response({
                'accounts': [
                    {
                        'email': acc.email,
                        'active': acc.is_active,
                        'last_used': acc.last_used,
                        'usage_count': acc.usage_count
                    }
                    for acc in self.account_manager.accounts
                ],
                'pending_captcha': self.captcha_solver.pending_captcha
            })
        
        app = web.Application()
        app.router.add_post('/captcha', handle_captcha)
        app.router.add_get('/status', handle_status)
        
        runner = web.AppRunner(app)
        await runner.setup()
        site = web.TCPSite(runner, 'localhost', 8081)
        await site.start()
        print("🌐 Web管理界面: http://localhost:8081")

async def main():
    """主程序"""
    print("=" * 60)
    print("🤖 Gemini Business 多账号管理器")
    print("=" * 60)
    
    # 初始化组件
    captcha_solver = CaptchaSolver()
    browser = GeminiBrowser(captcha_solver)
    account_manager = AccountManager()
    docker_manager = DockerManager()
    token_monitor = TokenMonitor(account_manager, docker_manager)
    web_interface = WebInterface(account_manager, captcha_solver)
    
    try:
        # 启动浏览器
        await browser.start()
        
        # 启动Web界面（可选）
        # await web_interface.start_server()
        
        # 获取下一个账号
        account = account_manager.get_next_account()
        if not account:
            print("❌ 没有可用账号，请检查 accounts.json")
            return
        
        print(f"\n📋 准备登录账号: {account.email}")
        
        # 登录并获取Token
        token_data = await browser.login_and_get_token(account)
        
        if token_data:
            # 保存账号信息
            account_manager.update_account(
                token_data['email'],
                token_data['bearer_token'],
                token_data['config_id']
            )
            
            # 部署到Docker
            updated_account = next(acc for acc in account_manager.accounts if acc.email == token_data['email'])
            docker_manager.deploy_account(updated_account)
            
            print(f"\n✅ 完成！账号 {token_data['email']} 已部署")
            print(f"   服务地址: http://localhost:8080")
            print(f"   Token有效期: 约1小时")
            print(f"   下次轮换: 50分钟后自动检查")
            
            # 启动监控（后台）
            monitor_task = asyncio.create_task(token_monitor.monitor())
            print(f"\n🔄 后台监控已启动，将自动轮换账号")
            
            # 保持运行
            await asyncio.Event().wait()
        else:
            print("\n❌ 登录失败，请重试")
    
    finally:
        await browser.close()

if __name__ == "__main__":
    asyncio.run(main())