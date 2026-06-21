"""Prompt construction helpers for the novel cover generator."""

from __future__ import annotations

import re
from dataclasses import dataclass
from typing import Any, Dict, Iterable, List, Mapping


FIXED_NEGATIVE_PROMPT = (
    "不要生成真实出版社、平台 Logo、水印、二维码；"
    "不要出现错别字、乱码文字、多余署名；"
    "避免低清晰度、过度模糊、脸部崩坏、手部畸形。"
)

NEUTRAL_ELEMENT_DEFAULTS = {
    "characters": "主角或核心主体根据书名、项目资料或创作想法呈现，身份、动作和关系不添加未提供设定",
    "scene_background": "背景场景根据书名、项目资料或创作想法呈现，保持与内容一致，不引入额外题材",
    "symbols_props": "只使用书名、项目资料或创作想法中明确出现的道具、符号和核心意象",
    "atmosphere_color": "根据作品内容选择情绪、主色和光影，主体突出，氛围清晰",
}


@dataclass(frozen=True)
class CoverTemplate:
    id: str
    name: str
    genre: str
    description: str
    preview: str
    preview_image: str
    typography_prompt: str
    element_prompt: str
    variables: List[str]
    defaults: Dict[str, str]

    def to_dict(self) -> Dict[str, Any]:
        return {
            "id": self.id,
            "name": self.name,
            "genre": self.genre,
            "description": self.description,
            "preview": self.preview,
            "preview_image": self.preview_image,
            "typography_prompt": self.typography_prompt,
            "element_prompt": self.element_prompt,
            "variables": list(self.variables),
            "defaults": dict(self.defaults),
        }


def _cover_template(
    *,
    template_id: str,
    name: str,
    genre: str,
    description: str,
    preview: str,
    preview_image: str,
    typography_style: str,
    defaults: Dict[str, str],
) -> CoverTemplate:
    return CoverTemplate(
        id=template_id,
        name=name,
        genre=genre,
        description=description,
        preview=preview,
        preview_image=preview_image,
        typography_prompt=(
            "文字：“{title}”，小字：“{author}·著”，文字在画面下方三分之一处，标题区域稳定可读，"
            f"{typography_style}"
        ),
        element_prompt=(
            "{characters}，{scene_background}，{symbols_props}，{atmosphere_color}，"
            "商业小说封面构图，主体清晰，背景服务于作品氛围，二维精修插画质感，高清细节。"
        ),
        variables=["characters", "scene_background", "symbols_props", "atmosphere_color"],
        defaults=defaults,
    )


COVER_TEMPLATES: List[CoverTemplate] = [
    _cover_template(
        template_id="wuxia_gold_blade",
        name="金锋书法字",
        genre="书法 / 金属 / 锋芒",
        description="剑刃飞白、金属高光和强对比描边组成的大气标题字。",
        preview="金色剑锋书法字、墨色山河、开阔气势",
        preview_image="/static/cover-examples/typography-01.jpg",
        typography_style=(
            "锋芒书法标题字，书法笔触飘逸凌厉，笔画带剑刃般的锋芒与飞白效果，"
            "金色渐变金属质感，边缘带亮白色高光与深灰色描边，外附柔和的白色辉光/光晕，"
            "字形开阔有力量感，标题层次清晰，超清细节，8K，高对比度"
        ),
        defaults={
            "characters": "主角或核心角色立于画面主体位置，轮廓坚毅，姿态有张力",
            "scene_background": "与故事设定匹配的辽阔场景，远景层次清晰",
            "symbols_props": "锋芒光痕、山河纹样、金色笔触装饰",
            "atmosphere_color": "金色高光与深色高反差，开阔有力量感",
        },
    ),
    _cover_template(
        template_id="luxury_gilded",
        name="轻奢鎏金浮雕字",
        genre="鎏金 / 浮雕 / 轻奢",
        description="烫金浮雕、铂金描边和柔和光影组成的高级标题字。",
        preview="鎏金浮雕字、铂金描边、柔和高级光影",
        preview_image="/static/cover-examples/typography-02.jpg",
        typography_style=(
            "高级轻奢鎏金标题字，金属烫金质感，细腻铂金描边，立体浮雕字体，"
            "字面层次简约清晰，光影柔和，高清质感，精致细腻"
        ),
        defaults={
            "characters": "核心角色轮廓利落，姿态克制高级，带精致杂志感",
            "scene_background": "极简深色空间、丝绒或大理石质感背景",
            "symbols_props": "合同、玫瑰、金色徽记或高定珠宝",
            "atmosphere_color": "柔和金色光影，干净克制，轻奢高级",
        },
    ),
    _cover_template(
        template_id="gothic_blood",
        name="哥特血滴字",
        genre="哥特 / 血滴 / 暗色",
        description="尖锐棱角、暗红血滴和白色外发光组成的冲击型标题字。",
        preview="暗红血滴、哥特尖角、白色外发光",
        preview_image="/static/cover-examples/typography-03.jpg",
        typography_style=(
            "哥特尖角标题字，尖锐棱角字体，红色血滴效果，液态滴落纹理，"
            "暗红色渐变，白色外发光描边，金属锐利质感，"
            "字形压迫感强，高对比度，强烈冲击力，超高清，细节拉满"
        ),
        defaults={
            "characters": "核心角色半身或剪影被暗色轮廓包裹，表情克制",
            "scene_background": "暗色空间、雨雾街角或高反差背景",
            "symbols_props": "血色信封、旧钥匙、裂纹镜面",
            "atmosphere_color": "暗红与黑色强对比，压迫感强",
        },
    ),
    _cover_template(
        template_id="chinese_horror_brush",
        name="血墨毛笔字",
        genre="血墨 / 毛笔 / 竖排",
        description="血墨质感、枯笔飞白和竖排版式组成的毛笔标题字。",
        preview="竖排血墨字、枯笔飞白、斑驳暗影",
        preview_image="/static/cover-examples/typography-04.jpg",
        typography_style=(
            "血墨毛笔书法标题字，凌厉锋利笔触，暗红血色水墨质感，猩红喷溅效果，"
            "破损斑驳做旧纹理，红白强对比光影，白色锐利外发光描边，"
            "竖版排版，高对比度氛围感，高清细节，字形张力强，"
            "血色晕染边缘，枯笔飞白笔触"
        ),
        defaults={
            "characters": "核心角色背影或侧影置于竖向构图中，轮廓清晰",
            "scene_background": "斑驳墙面、旧木纹或暗色空间，保留竖向留白",
            "symbols_props": "血墨晕染、铜色小物、斑驳木纹、红色纸片",
            "atmosphere_color": "暗红血墨与冷白光对撞，高反差压迫感",
        },
    ),
    _cover_template(
        template_id="sweet_campus_bluegreen",
        name="蓝绿柔光花体字",
        genre="蓝绿 / 花体 / 柔光",
        description="蓝绿色渐变、圆润手写花体和星光光斑组成的柔和标题字。",
        preview="蓝绿渐变花体、星光光斑、柔焦亮色",
        preview_image="/static/cover-examples/typography-05.jpg",
        typography_style=(
            "柔光花体标题字，梦幻渐变蓝绿色调，圆润流畅的手写花体字，"
            "带有飘逸装饰性曲线，高光与外发光效果，清新柔焦质感，点缀星光与光斑元素，"
            "字体边缘有柔和光晕，干净通透的色彩，细腻的光泽描边，"
            "整体轻盈明亮，高清细节"
        ),
        defaults={
            "characters": "角色或主体面向柔光，表情自然温柔，轮廓干净",
            "scene_background": "明亮户外、窗边或柔光空间，背景虚化通透",
            "symbols_props": "星光、光斑、课本、心动便签",
            "atmosphere_color": "蓝绿色梦幻渐变，柔焦清新，明亮通透",
        },
    ),
    _cover_template(
        template_id="fresh_xianxia_jade",
        name="青玉花体字",
        genre="青玉 / 花体 / 通透",
        description="青绿色玉石通透感、卷曲线条和星光装饰组成的花体标题字。",
        preview="青绿玉石花体、菱形星光、通透柔光",
        preview_image="/static/cover-examples/typography-06.jpg",
        typography_style=(
            "青玉花体标题字，青绿色渐变光泽字体，带有飘逸装饰性卷曲线条，"
            "金属玉石般通透质感，柔和外发光描边，点缀星光与菱形装饰元素，"
            "细腻高光描边，手写毛笔风曲线，高通透度，青绿色调渐变，边缘柔化光晕，超高清细节"
        ),
        defaults={
            "characters": "核心角色或主体被青绿色光效勾勒，姿态轻盈",
            "scene_background": "竹影、云雾、发光藤蔓与清泉等清透场景",
            "symbols_props": "青玉、草叶灵纹、星光菱形装饰",
            "atmosphere_color": "青绿色通透光泽，梦幻柔和，空气清澈",
        },
    ),
    _cover_template(
        template_id="cute_chibi_bold",
        name="Q版软萌粗描边字",
        genre="Q版 / 粗描边 / 涂鸦",
        description="白底黑边、胖乎乎字形和手绘涂鸦装饰组成的可爱标题字。",
        preview="Q版圆润字、黑白粗描边、星星涂鸦",
        preview_image="/static/cover-examples/typography-07.jpg",
        typography_style=(
            "Q版可爱圆润艺术字体，卡通漫画风加粗字体，白色填充+黑色粗描边，"
            "圆润饱满的胖乎乎笔画，搭配星星、圆点、小三角等可爱装饰元素，"
            "带有俏皮活泼的手绘涂鸦感，白底黑边醒目清晰，标题字稳定可读，"
            "高对比度，可爱软萌风格，细节干净利落"
        ),
        defaults={
            "characters": "Q版核心角色或主体，表情活泼讨喜，轮廓圆润",
            "scene_background": "暖色小空间、糖果色街景或轻松生活场景",
            "symbols_props": "星星、圆点、小三角、贴纸涂鸦",
            "atmosphere_color": "明亮高对比，软萌活泼，干净利落",
        },
    ),
    _cover_template(
        template_id="wuxia_blue_lightning",
        name="青蓝电光书法字",
        genre="青蓝 / 金属 / 电光",
        description="青蓝金属、书法飞白和火焰电光组成的强冲击标题字。",
        preview="青蓝金属书法、电光火焰、速度光效",
        preview_image="/static/cover-examples/typography-08.jpg",
        typography_style=(
            "青蓝金属书法标题字，凌厉锋利的毛笔书法字，青蓝色金属光泽渐变，"
            "带有发光外描边，水墨笔触质感，字体边缘带有火焰/电光特效，"
            "带有金属反光和光晕，高对比度，字形大气有速度感，"
            "超高清细节，强烈视觉冲击力"
        ),
        defaults={
            "characters": "动作感强的主角或主体，轮廓被青蓝电光切出",
            "scene_background": "高反差暗色空间、风暴光效或速度线背景",
            "symbols_props": "金属纹样、青蓝电光、火焰边缘、飞散光片",
            "atmosphere_color": "冷青蓝金属光与暗色背景强对比",
        },
    ),
    _cover_template(
        template_id="romance_blue_pink",
        name="蓝粉浪漫花体字",
        genre="蓝粉 / 花体 / 浪漫",
        description="蓝紫玫红渐变、优雅花体和爱心星光组成的华丽标题字。",
        preview="蓝紫玫红渐变、爱心星光、华丽曲线",
        preview_image="/static/cover-examples/typography-09.jpg",
        typography_style=(
            "浪漫优雅花体手写标题字，蓝紫到玫红渐变色彩，"
            "细腻高光描边，带有飘逸卷曲装饰线条，柔和外发光效果，点缀星光、爱心小装饰，"
            "字体边缘柔化光晕，通透干净的色彩质感，"
            "高清细节，精致华丽的曲线设计"
        ),
        defaults={
            "characters": "核心角色或主体靠近柔光，姿态优雅，关系感明确",
            "scene_background": "夜色光影、玻璃窗、玫瑰与柔光背景",
            "symbols_props": "爱心星光、玫瑰、戒指或高楼灯影",
            "atmosphere_color": "蓝紫到玫红渐变，通透干净，浪漫华丽",
        },
    ),
    _cover_template(
        template_id="ancient_romance_platinum",
        name="白金花瓣书法字",
        genre="白金 / 书法 / 花瓣",
        description="白金金属书法、深色描边和红色花瓣组成的清冷标题字。",
        preview="白金毛笔字、红色花瓣、清冷深情",
        preview_image="/static/cover-examples/typography-10.jpg",
        typography_style=(
            "白金书法标题字，凌厉飘逸的毛笔书法字，白金色渐变金属光泽，"
            "带有深色外描边和红色高光点缀，边缘带柔化光晕，搭配飘落的红色花瓣装饰，"
            "水墨风笔触，苍劲有力的笔画，带有飞白效果，清冷又深情的氛围感，高对比度，"
            "超高清细节，精致的金属反光质感"
        ),
        defaults={
            "characters": "核心角色或主体与红色花瓣形成前后景层次",
            "scene_background": "月光、雪色、深色墙面或台阶式背景",
            "symbols_props": "红色花瓣、玉簪、信笺、薄雾",
            "atmosphere_color": "白金清冷光泽配深色描边，深情克制",
        },
    ),
    _cover_template(
        template_id="cyber_mecha_neon",
        name="霓虹科技金属字",
        genre="科技 / 金属 / 霓虹",
        description="硬朗粗体、机械切割线和青蓝霓虹组成的科技标题字。",
        preview="青蓝金属粗体、机械切割线、霓虹描边",
        preview_image="/static/cover-examples/typography-11.jpg",
        typography_style=(
            "霓虹科技金属标题字，硬朗厚重的科技感粗体字，青蓝色金属渐变光泽，"
            "带有尖锐棱角与机械切割线条，发光霓虹描边，金属镀铬质感，"
            "带有几何装饰（三角/机械纹路），高对比度，冷色调，"
            "超高清细节，强烈视觉冲击力"
        ),
        defaults={
            "characters": "核心角色或主体被冷色轮廓光勾勒，姿态坚定",
            "scene_background": "金属空间、霓虹街景或几何光幕背景",
            "symbols_props": "金属装甲片、三角几何纹路、全息界面",
            "atmosphere_color": "青蓝冷光、高对比金属质感、未来科技感",
        },
    ),
    _cover_template(
        template_id="male_red_gold_seal",
        name="红金印章书法字",
        genre="红金 / 书法 / 印章",
        description="正红金属书法、鎏金高光和圆形印章框组成的厚重标题字。",
        preview="红金撞色、圆形印章、豪迈毛笔字",
        preview_image="/static/cover-examples/typography-12.jpg",
        typography_style=(
            "红金印章书法标题字，大气磅礴的毛笔书法字体，正红渐变金属光泽，"
            "带有鎏金高光与外发光描边，笔画苍劲飘逸、笔锋锐利有力，搭配圆形印章式文字框，"
            "点缀暖金色星光光斑，带有水墨飞白效果，红金撞色强烈对比，"
            "高清细节，厚重豪迈的氛围感"
        ),
        defaults={
            "characters": "核心角色或主体立于画面中央，姿态沉稳有压场感",
            "scene_background": "红金高反差背景、旗形光影或厚重纹样空间",
            "symbols_props": "圆形印章、赤色纹样、金色光斑、厚重边框",
            "atmosphere_color": "红金强对比，厚重豪迈，热血昂扬",
        },
    ),
    _cover_template(
        template_id="female_purple_vertical",
        name="紫底竖排白字",
        genre="白字 / 竖排 / 紫底",
        description="纯白竖排书法、深紫底色和红色小印章组成的柔美标题字。",
        preview="纯白竖排书法、深紫背景、红色小印章",
        preview_image="/static/cover-examples/typography-13.jpg",
        typography_style=(
            "竖排白色毛笔标题字，温婉飘逸的毛笔书法字体，纯白色笔触，"
            "带有柔和飞白效果，线条纤细灵动，竖排排版，搭配红色边框竖排小字印章，"
            "深紫色底色上保持高可读性，字体边缘干净利落，带有淡淡的朦胧质感，"
            "柔美又大气，高清细节，笔触流畅自然"
        ),
        defaults={
            "characters": "核心角色或主体置于竖向留白中，神情温婉却有韧性",
            "scene_background": "深紫底色、纱雾、花枝与柔和暗面背景",
            "symbols_props": "红色竖排小印章、花枝、玉佩、纱帘",
            "atmosphere_color": "深紫朦胧背景配纯白笔触，柔美大气",
        },
    ),
]


def get_cover_templates() -> List[Dict[str, Any]]:
    """Return public cover template definitions."""
    return [template.to_dict() for template in COVER_TEMPLATES]


def _template_by_id(template_id: str) -> CoverTemplate:
    wanted = str(template_id or "").strip()
    for template in COVER_TEMPLATES:
        if template.id == wanted:
            return template
    raise ValueError(f"未知封面模板：{template_id}")


def _compact_text(value: Any, limit: int = 220) -> str:
    text = re.sub(r"\s+", " ", str(value or "")).strip()
    if len(text) <= limit:
        return text
    return text[:limit].rstrip() + "..."


def _first_non_empty(*values: Any) -> str:
    for value in values:
        text = _compact_text(value)
        if text:
            return text
    return ""


def _join_non_empty(values: Iterable[Any], *, limit: int = 420) -> str:
    parts: List[str] = []
    total = ""
    for value in values:
        text = _compact_text(value, limit)
        if not text:
            continue
        candidate = "；".join([*parts, text])
        if len(candidate) > limit:
            break
        parts.append(text)
        total = candidate
    return total


def _contains_any(text: str, keywords: Iterable[str]) -> bool:
    return any(keyword in text for keyword in keywords)


def _infer_visual_elements_from_seed(seed_text: str, title: str, template: CoverTemplate) -> Dict[str, str]:
    """Build neutral cover constraints from user-provided names and ideas."""
    text = _compact_text(seed_text, 900)
    title_text = _compact_text(title, 80)
    if not text and not title_text:
        return {}

    subject = _first_non_empty(text, title_text)
    quoted = f"“{subject}”"
    inferred: Dict[str, str] = {
        "characters": f"围绕{quoted}呈现主角或核心主体，身份、动作和关系以已提供内容为准",
        "scene_background": f"围绕{quoted}建立封面背景，场景只来自已提供内容，不加入额外题材设定",
        "symbols_props": f"提取{quoted}中明确出现的道具、符号和核心意象；没有明确道具时保持简洁",
        "atmosphere_color": f"根据{quoted}判断情绪、主色和光影，保持商业封面清晰可读",
    }
    return {key: _compact_text(value, 260) for key, value in inferred.items() if _compact_text(value)}


def _iter_dict_rows(rows: Any) -> Iterable[Mapping[str, Any]]:
    if isinstance(rows, dict):
        candidate = rows.get("items") or rows.get("chapters") or rows.get("data") or []
    else:
        candidate = rows
    if not isinstance(candidate, list):
        return []
    return [row for row in candidate if isinstance(row, Mapping)]


def _format_named_rows(rows: Any, *, limit: int = 3) -> str:
    parts: List[str] = []
    for row in list(_iter_dict_rows(rows))[:limit]:
        name = _first_non_empty(row.get("name"), row.get("title"), row.get("chapter_title"))
        role = _first_non_empty(row.get("role"), row.get("type"))
        desc = _first_non_empty(
            row.get("appearance"),
            row.get("description"),
            row.get("summary"),
            row.get("content"),
            row.get("details"),
        )
        if name and role and desc:
            parts.append(f"{name}（{role}）：{desc}")
        elif name and desc:
            parts.append(f"{name}：{desc}")
        elif name:
            parts.append(name)
        elif desc:
            parts.append(desc)
    return "；".join(parts)


def _query_terms(query: str) -> List[str]:
    text = _compact_text(query, 160)
    if not text:
        return []
    terms = [item.strip() for item in re.split(r"[\s,，、;；/|]+", text) if item.strip()]
    if text not in terms:
        terms.insert(0, text)
    return terms[:8]


class ProjectCoverContextExtractor:
    """Extract cover-friendly elements from the current project."""

    def extract(self, project_manager: Any) -> Dict[str, str]:
        project = project_manager.get_current_project() if project_manager else None
        project_name = _first_non_empty(getattr(project, "name", ""))
        description = _first_non_empty(getattr(project, "description", ""))
        genre = _first_non_empty(getattr(project, "novel_type", ""))

        characters = _format_named_rows(project_manager.load_project_data("characters"), limit=3)
        worldbuilding = _format_named_rows(project_manager.load_project_data("worldbuilding"), limit=4)
        items = _format_named_rows(project_manager.load_project_data("items"), limit=4)
        outline = _format_named_rows(project_manager.load_project_data("outline"), limit=3)
        chapters = _format_named_rows(project_manager.load_project_data("chapters"), limit=2)
        scene_background = _join_non_empty([worldbuilding, outline, chapters, description], limit=520)

        return {
            "title": project_name,
            "genre": genre,
            "summary": description,
            "characters": characters,
            "scene_background": scene_background,
            "symbols_props": items,
            "atmosphere_color": _infer_atmosphere(genre, description, worldbuilding),
        }

    def find_character_context(self, project_manager: Any, query: str) -> str:
        terms = _query_terms(query)
        if not terms or not project_manager:
            return ""
        rows = list(_iter_dict_rows(project_manager.load_project_data("characters")))
        matched: List[Mapping[str, Any]] = []
        for row in rows:
            haystack = " ".join(
                _compact_text(row.get(key), 260)
                for key in ("name", "title", "role", "appearance", "description", "summary", "content", "details")
                if _compact_text(row.get(key))
            )
            if any(term and term in haystack for term in terms):
                matched.append(row)
        return _format_named_rows(matched, limit=2)


def _infer_atmosphere(genre: str, description: str, worldbuilding: str) -> str:
    text = f"{genre} {description} {worldbuilding}"
    if any(keyword in text for keyword in ("恐怖", "诡异", "悬疑", "惊悚", "异常", "红月")):
        return "暗色基调，红黑对比，压抑氛围与强光效明暗对比"
    if any(keyword in text for keyword in ("科幻", "星舰", "未来", "机甲", "赛博")):
        return "冷色科技光效，远景宏大，金属质感清晰"
    if any(keyword in text for keyword in ("言情", "治愈", "甜宠", "校园")):
        return "柔和暖光，情绪细腻，画面干净通透"
    if any(keyword in text for keyword in ("玄幻", "修仙", "仙侠", "奇幻")):
        return "高幻想色彩，灵光流动，层次丰富"
    if any(keyword in text for keyword in ("三国", "历史", "穿越", "王朝")):
        return "厚重红金色调，史诗感强，光影豪迈"
    return ""


class CoverPromptBuilder:
    """Build typography, element and final prompts for cover generation."""

    def __init__(self, extractor: ProjectCoverContextExtractor | None = None):
        self.extractor = extractor or ProjectCoverContextExtractor()

    def build_prompt(
        self,
        *,
        project_manager: Any,
        template_id: str,
        source_mode: str = "project_plus_custom",
        title: str = "",
        author: str = "",
        custom_elements: Mapping[str, Any] | None = None,
    ) -> Dict[str, Any]:
        template = _template_by_id(template_id)
        custom = {
            str(key): _compact_text(value, 360)
            for key, value in dict(custom_elements or {}).items()
            if _compact_text(value)
        }
        creative_idea = _compact_text(custom.get("creative_idea"), 520)
        project_elements = self.extractor.extract(project_manager)
        source_mode = str(source_mode or "project_plus_custom").strip() or "project_plus_custom"

        if source_mode == "custom":
            merged: Dict[str, str] = {}
        else:
            merged = dict(project_elements)

        if source_mode in {"custom", "project_plus_custom"}:
            merged.update(custom)

        character_query = _compact_text(custom.get("characters"))
        if character_query and source_mode != "custom":
            matched_character = self.extractor.find_character_context(project_manager, character_query)
            if matched_character:
                merged["characters"] = _join_non_empty([character_query, matched_character], limit=420)

        variable_keys = [
            "characters",
            "scene_background",
            "symbols_props",
            "atmosphere_color",
        ]
        project_context_empty = not any(_compact_text(project_elements.get(key)) for key in variable_keys)
        custom_elements_empty = not any(_compact_text(custom.get(key)) for key in variable_keys)
        source_content_empty = not any(_compact_text(merged.get(key)) for key in variable_keys)
        idea_only_source = source_content_empty and bool(creative_idea)

        project_title = project_elements.get("title", "")
        explicit_title = _compact_text(title)
        merged["title"] = _first_non_empty(explicit_title, project_title, template.defaults.get("title"), "未命名小说")
        merged["author"] = _first_non_empty(author, merged.get("author"), "XXXX")

        seed_text = _join_non_empty(
            [
                creative_idea,
                merged.get("characters"),
                merged.get("scene_background"),
                merged.get("symbols_props"),
                project_elements.get("genre"),
                project_elements.get("summary"),
                merged.get("title"),
            ],
            limit=900,
        )
        inferred_elements = _infer_visual_elements_from_seed(seed_text, merged["title"], template)
        inferred_fields: List[str] = []
        for key in variable_keys:
            if not _compact_text(merged.get(key)) and _compact_text(inferred_elements.get(key)):
                merged[key] = inferred_elements[key]
                inferred_fields.append(key)

        fallback_fields: List[str] = []
        for key, value in {**template.defaults, **NEUTRAL_ELEMENT_DEFAULTS}.items():
            before_default = merged.get(key)
            merged[key] = _first_non_empty(before_default, value)
            if key in variable_keys and not _compact_text(before_default) and _compact_text(merged.get(key)):
                fallback_fields.append(key)

        typography_prompt = template.typography_prompt.format(**merged)
        element_prompt = template.element_prompt.format(**merged)
        final_prompt = "\n".join(
            [
                typography_prompt,
                element_prompt,
                "构图要求：小说商业封面，主体清晰，远看可识别，标题区域稳定可读，避免真实平台标识。",
                f"禁止项：{FIXED_NEGATIVE_PROMPT}",
            ]
        )

        title_source = "custom" if explicit_title else ("project" if project_title else "fallback")
        title_warning = ""
        if title_source == "project":
            title_warning = f"当前书名默认使用项目名“{project_title}”。如果项目名不是小说书名，请改成正确书名后再生成。"

        completion_notice = ""
        prompt_generation_mode = "project_or_custom"
        if inferred_fields:
            completion_notice = (
                "已根据书名、角色名或你的创作想法补全缺失画面元素；字体提示词保持所选示例不变。"
            )
            prompt_generation_mode = "idea_inferred" if idea_only_source else "local_inferred"
        if source_content_empty and not inferred_fields:
            completion_notice = (
                "当前选择的元素来源没有可用封面内容，已使用中性封面约束补全角色、场景、道具和色彩；"
                "如果项目为空，建议填写四个元素或“创作想法”。"
            )
            prompt_generation_mode = "template_defaults"

        return {
            "template_id": template.id,
            "template_name": template.name,
            "title": merged["title"],
            "title_source": title_source,
            "title_warning": title_warning,
            "author": merged["author"],
            "source_mode": source_mode,
            "elements": {
                key: merged.get(key, "")
                for key in [
                    "characters",
                    "scene_background",
                    "symbols_props",
                    "atmosphere_color",
                ]
            },
            "custom_elements": custom,
            "creative_idea": creative_idea,
            "inferred_fields": inferred_fields,
            "fallback_fields": fallback_fields,
            "project_context_empty": project_context_empty,
            "custom_elements_empty": custom_elements_empty,
            "completion_notice": completion_notice,
            "prompt_generation_mode": prompt_generation_mode,
            "prompt_api_config_id": "",
            "prompt_model": "",
            "typography_prompt": typography_prompt,
            "element_prompt": element_prompt,
            "final_prompt": final_prompt,
            "negative_prompt": FIXED_NEGATIVE_PROMPT,
        }

    def apply_elements(
        self,
        draft: Mapping[str, Any],
        elements: Mapping[str, Any],
        *,
        prompt_generation_mode: str,
        prompt_api_config_id: str = "",
        prompt_model: str = "",
        completion_notice: str = "",
        prompt_model_warning: str = "",
    ) -> Dict[str, Any]:
        """Return a prompt draft rebuilt with enhanced visual elements."""
        template = _template_by_id(str(draft.get("template_id") or ""))
        existing_elements = draft.get("elements") if isinstance(draft.get("elements"), Mapping) else {}
        merged: Dict[str, str] = {
            "title": _first_non_empty(draft.get("title"), "未命名小说"),
            "author": _first_non_empty(draft.get("author"), "XXXX"),
        }

        for key in template.variables:
            merged[key] = _first_non_empty(
                elements.get(key),
                existing_elements.get(key),
                NEUTRAL_ELEMENT_DEFAULTS.get(key),
            )

        typography_prompt = _first_non_empty(draft.get("typography_prompt")) or template.typography_prompt.format(**merged)
        element_prompt = template.element_prompt.format(**merged)
        final_prompt = "\n".join(
            [
                typography_prompt,
                element_prompt,
                "构图要求：小说商业封面，主体清晰，远看可识别，标题区域稳定可读，避免真实平台标识。",
                f"禁止项：{FIXED_NEGATIVE_PROMPT}",
            ]
        )

        result = dict(draft)
        result.update(
            {
                "elements": {key: merged.get(key, "") for key in template.variables},
                "typography_prompt": typography_prompt,
                "element_prompt": element_prompt,
                "final_prompt": final_prompt,
                "prompt_generation_mode": prompt_generation_mode,
                "prompt_api_config_id": prompt_api_config_id,
                "prompt_model": prompt_model,
            }
        )
        if completion_notice:
            result["completion_notice"] = completion_notice
        if prompt_model_warning:
            result["prompt_model_warning"] = prompt_model_warning
        return result
