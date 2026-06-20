"""LLM-assisted prompt enrichment for novel cover drafts."""

from __future__ import annotations

import json
import logging
import re
from typing import Any, Dict, Iterable, Mapping, Tuple

from .agent_config import AgentModelConfig, get_config_manager
from .agents.llm_client import LLMClient
from .cover_prompt_builder import CoverPromptBuilder

logger = logging.getLogger(__name__)

_FIELD_KEYS = ("characters", "scene_background", "symbols_props", "atmosphere_color")
_IMAGE_MODEL_PATTERN = re.compile(
    r"(image|imagen|dall[-_ ]?e|gpt-image|codex-gpt-image|flux|stable[-_ ]?diffusion|"
    r"seedream|jimeng|midjourney|ideogram|recraft|hidream|playground|pixverse)",
    re.IGNORECASE,
)
_NON_TEXT_MODEL_PATTERN = re.compile(r"(embedding|embed|rerank|tts|whisper|audio|speech|moderation)", re.IGNORECASE)
_TEMPLATE_LEAKAGE_TERMS = (
    "长剑",
    "持剑",
    "剑刃",
    "仙山",
    "宗门",
    "古战场",
    "秘境",
    "玄幻",
    "战甲",
    "符纹",
    "符文",
    "灵光",
    "山河纹",
    "衣袂",
    "修仙",
    "仙侠",
    "神魔",
)


class CoverPromptEnhancer:
    """Use a selected text model to polish the four cover visual elements."""

    def __init__(self, builder: CoverPromptBuilder | None = None):
        self.builder = builder or CoverPromptBuilder()

    async def enhance(
        self,
        draft: Mapping[str, Any],
        *,
        api_config_id: str = "",
        model: str = "",
    ) -> Dict[str, Any]:
        config = self._resolve_model_config(api_config_id=api_config_id, model=model)
        client = LLMClient(model_config=config, metrics_namespace="CoverPrompt")

        payload = {
            "title": draft.get("title", ""),
            "author": draft.get("author", ""),
            "source_mode": draft.get("source_mode", ""),
            "elements": draft.get("elements", {}),
            "creative_idea": draft.get("creative_idea", ""),
            "inferred_fields": draft.get("inferred_fields", []),
            "fallback_fields": draft.get("fallback_fields", []),
            "project_context_empty": bool(draft.get("project_context_empty")),
            "custom_elements_empty": bool(draft.get("custom_elements_empty")),
            "font_template_rule": "字体模板只代表字形、材质、配色和光效，不代表题材、人物、背景或道具。",
        }

        system_prompt = (
            "你是小说商业封面提示词导演，只负责润色封面画面元素，不改书名、作者和字体提示词。"
            "严格输出 JSON，不要解释，不要 Markdown。JSON 只能包含四个键："
            "characters, scene_background, symbols_props, atmosphere_color。"
            "每个值用中文短句，视觉化、具体、适合封面，避免真实平台 Logo、水印、二维码。"
            "字体模板只代表文字字形、材质、配色和光效，不代表小说题材。"
            "禁止因为模板名、字体风格或颜色加入用户内容中没有出现的人物身份、背景、武器、门派、战甲或玄幻元素。"
            "如果项目资料为空，只能根据书名、创作想法和已有四项元素补全合理画面，不要假装知道不存在的剧情设定。"
        )
        user_prompt = (
            "请把以下封面草稿补全为更适合生图模型理解的四个画面元素。\n"
            "必须保留已有具体设定的核心信息；缺失字段只能根据书名、创作想法和已有元素补全。\n"
            "不要把字体模板当作题材参考；任何内容都不能因为字体风格被改成未提供的武器、宗门、战甲或仙山等元素。\n\n"
            f"{json.dumps(payload, ensure_ascii=False)}"
        )

        response = await client.call(
            messages=[{"role": "user", "content": user_prompt}],
            system_prompt=system_prompt,
            temperature=0.45,
            max_tokens=900,
            stream=False,
        )
        parsed = _parse_json_object(str(response or ""))
        enhanced = _normalize_elements(parsed)
        enhanced = _remove_unanchored_template_leakage(enhanced, draft)
        if not any(enhanced.values()):
            raise RuntimeError("文本模型没有返回可用的四项封面元素。")

        notice = "已使用所选文本模型润色角色、场景、道具和色彩元素。"
        if draft.get("prompt_generation_mode") == "template_defaults":
            notice = "项目资料和自定义元素为空，已先按中性封面约束补全，再由所选文本模型润色为可用封面描写。"
        return self.builder.apply_elements(
            draft,
            enhanced,
            prompt_generation_mode="text_model",
            prompt_api_config_id=config.api_config_id,
            prompt_model=config.model,
            completion_notice=notice,
        )

    @staticmethod
    def _resolve_model_config(*, api_config_id: str, model: str) -> AgentModelConfig:
        manager = get_config_manager()
        api_base = ""
        api_key = ""
        temperature = 0.7
        max_tokens = 900
        api_type = "openai_chat"
        resolved_model = str(model or "").strip()
        resolved_config_id = str(api_config_id or "").strip()

        if resolved_config_id:
            multi = manager.get_multi_config()
            for cfg in multi.configs:
                if cfg.id == resolved_config_id:
                    api_base = str(getattr(cfg, "api_base", "") or "").strip()
                    api_key = _get_primary_key(cfg)
                    temperature = float(getattr(cfg, "temperature", temperature) or temperature)
                    max_tokens = int(getattr(cfg, "max_tokens", max_tokens) or max_tokens)
                    api_type = str(getattr(cfg, "api_type", api_type) or api_type)
                    if not resolved_model:
                        resolved_model = _first_text_model(getattr(cfg, "models", []))
                    break

        if not api_base or not api_key:
            multi = manager.get_multi_config()
            active = multi.get_active_config()
            if active:
                api_base = api_base or str(getattr(active, "api_base", "") or "").strip()
                api_key = api_key or _get_primary_key(active)
                temperature = float(getattr(active, "temperature", temperature) or temperature)
                max_tokens = int(getattr(active, "max_tokens", max_tokens) or max_tokens)
                api_type = str(getattr(active, "api_type", api_type) or api_type)
                if not resolved_config_id:
                    resolved_config_id = str(getattr(active, "id", "") or "")
                if not resolved_model:
                    resolved_model = _first_text_model(getattr(active, "models", []))

        if not api_base or not api_key or not resolved_model:
            global_config = manager.get_global_config()
            api_base = api_base or str(getattr(global_config, "api_base", "") or "").strip()
            api_key = api_key or _get_primary_key(global_config)
            temperature = float(getattr(global_config, "temperature", temperature) or temperature)
            max_tokens = int(getattr(global_config, "max_tokens", max_tokens) or max_tokens)
            api_type = str(getattr(global_config, "api_type", api_type) or api_type)
            global_model = str(getattr(global_config, "model", "") or "").strip()
            if not resolved_model and _is_likely_text_model(global_model):
                resolved_model = global_model

        if not api_base or not api_key:
            raise ValueError("未配置可用的文本 API，请先在设置中完成 API 配置。")
        if not resolved_model:
            raise ValueError("未选择文本模型，请先选择用于生成描写提示词的语言模型。")
        if not _is_likely_text_model(resolved_model):
            raise ValueError(f"当前选择的模型“{resolved_model}”不像文本模型，请改选语言模型生成描写提示词。")

        return AgentModelConfig(
            agent_name="CoverPrompt",
            api_config_id=resolved_config_id,
            api_base=api_base,
            api_key=api_key,
            model=resolved_model,
            temperature=temperature,
            max_tokens=max_tokens,
            api_type=api_type,
        )


def _get_primary_key(config: Any) -> str:
    if hasattr(config, "get_primary_key"):
        return str(config.get_primary_key() or "").strip()
    return str(getattr(config, "api_key", "") or "").strip()


def _first_text_model(models: Iterable[str]) -> str:
    for model in models or []:
        model_name = str(model or "").strip()
        if _is_likely_text_model(model_name):
            return model_name
    return ""


def _is_likely_text_model(model: str) -> bool:
    name = str(model or "").strip()
    return bool(name and not _IMAGE_MODEL_PATTERN.search(name) and not _NON_TEXT_MODEL_PATTERN.search(name))


def _parse_json_object(text: str) -> Dict[str, Any]:
    candidate = str(text or "").strip()
    for marker in ("```json", "```"):
        if marker in candidate:
            start = candidate.find(marker) + len(marker)
            end = candidate.rfind("```")
            if end > start:
                candidate = candidate[start:end].strip()
                break
    if "{" in candidate and "}" in candidate:
        candidate = candidate[candidate.find("{"): candidate.rfind("}") + 1]
    data = json.loads(candidate)
    return data if isinstance(data, dict) else {}


def _normalize_elements(data: Mapping[str, Any]) -> Dict[str, str]:
    aliases = {
        "characters": ("characters", "角色/人物", "角色", "人物"),
        "scene_background": ("scene_background", "背景场景", "场景", "背景"),
        "symbols_props": ("symbols_props", "关键道具/符号", "关键道具", "符号", "道具"),
        "atmosphere_color": ("atmosphere_color", "画面情绪/色彩", "情绪色彩", "色彩", "氛围"),
    }
    result: Dict[str, str] = {}
    for key in _FIELD_KEYS:
        value = ""
        for alias in aliases[key]:
            if alias in data:
                value = _clean_text(data.get(alias), limit=220)
                if value:
                    break
        result[key] = value
    return result


def _remove_unanchored_template_leakage(elements: Mapping[str, str], draft: Mapping[str, Any]) -> Dict[str, str]:
    existing = draft.get("elements") if isinstance(draft.get("elements"), Mapping) else {}
    source_text = _source_text_from_draft(draft)
    cleaned: Dict[str, str] = {}
    for key in _FIELD_KEYS:
        value = _clean_text(elements.get(key), limit=220)
        if value and _contains_unanchored_template_leakage(value, source_text):
            value = _clean_text(existing.get(key), limit=220)
        cleaned[key] = value
    return cleaned


def _source_text_from_draft(draft: Mapping[str, Any]) -> str:
    parts = [
        draft.get("title", ""),
        draft.get("creative_idea", ""),
    ]
    for container_name in ("elements", "custom_elements"):
        container = draft.get(container_name)
        if isinstance(container, Mapping):
            parts.extend(container.get(key, "") for key in _FIELD_KEYS)
    return _clean_text("；".join(str(part or "") for part in parts), limit=1800)


def _contains_unanchored_template_leakage(value: str, source_text: str) -> bool:
    text = str(value or "")
    source = str(source_text or "")
    return any(term in text and term not in source for term in _TEMPLATE_LEAKAGE_TERMS)


def _clean_text(value: Any, *, limit: int = 220) -> str:
    if isinstance(value, list):
        value = "；".join(str(item).strip() for item in value if str(item).strip())
    text = re.sub(r"\s+", " ", str(value or "")).strip()
    if len(text) <= limit:
        return text
    return text[:limit].rstrip() + "..."
