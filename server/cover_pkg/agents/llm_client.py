"""Stub module: LLM 客户端，用于提示词润色。"""
import asyncio
import httpx


class LLMClient:
    def __init__(self, model_config=None, metrics_namespace=''):
        self._config = model_config

    async def call(self, messages=None, system_prompt='', temperature=0.7, max_tokens=900, stream=False):
        raise RuntimeError(
            'LLM 客户端未完全配置——提示词润色功能需要完整的 agents 模块。'
            '当前使用本地模板直接生成提示词。'
        )
