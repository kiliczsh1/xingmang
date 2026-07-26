"""Stub module: 为封面生成代码提供缺失的 agent_config 依赖。"""
import importlib
import sys


def normalize_image_api_format(fmt):
    if not fmt or fmt == 'auto':
        return 'auto'
    return str(fmt)


class _AgentModelConfig:
    def __init__(self, **kwargs):
        self.agent_name = kwargs.get('agent_name', 'CoverPrompt')
        self.api_config_id = kwargs.get('api_config_id', '')
        self.api_base = kwargs.get('api_base', '')
        self.api_key = kwargs.get('api_key', '')
        self.model = kwargs.get('model', '')
        self.temperature = kwargs.get('temperature', 0.7)
        self.max_tokens = kwargs.get('max_tokens', 900)
        self.api_type = kwargs.get('api_type', 'openai_chat')

    def __repr__(self):
        return f'<AgentModelConfig {self.model}>'


AgentModelConfig = _AgentModelConfig


class _FakeMultiConfig:
    configs = []

    def get_active_config(self):
        return None


class _FakeGlobalConfig:
    api_base = ''
    api_key = ''
    model = ''


class _ConfigManager:
    def get_multi_config(self):
        return _FakeMultiConfig()

    def get_global_config(self):
        return _FakeGlobalConfig()

    def get_active_config(self):
        return None


_config_manager_instance = _ConfigManager()


def get_config_manager():
    return _config_manager_instance


if not hasattr(sys.modules.get('utils', None), 'atomic_write'):
    _utils = type(sys)('utils')
    sys.modules.setdefault('utils', _utils)
