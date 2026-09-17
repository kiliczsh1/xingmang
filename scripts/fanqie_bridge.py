#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
番茄小说榜单 HTTP 桥 —— 供 xingmang Go 后端调用。

数据源：github.com/wengchengjian/fanqie-rank-mcp（a_bogus 签名 + PUA 字体解码）。
本脚本只复用其 api.py / a_bogus.py / pua_map.py，自行增强解析（额外提取字数/在读/名次变动）。

接口：
  GET /health                         健康检查
  GET /rankings                       榜单分组（tab + 题材列表）
  GET /ranking?gender=&rank_mold=&category_id=&offset=&limit=
                                       某榜书籍列表（含 wordNumber/readCount/rankDiff/creationStatus）

环境变量：
  FANQIE_REPO_DIR   仓库路径（默认 脚本同级 fanqie/ 目录，即打包产物内的 scripts/fanqie/）
  FANQIE_BRIDGE_PORT 监听端口（默认 19900）

依赖：
  pip install httpx

运行：python scripts/fanqie_bridge.py
"""
import asyncio
import json
import os
import sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

# 默认指向脚本同级的 fanqie/ 子目录（打包产物内为 scripts/fanqie/）
# 开发时若未打包，可设置 FANQIE_REPO_DIR 指向本机的 fanqie-rank-mcp 仓库
_SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
_DEFAULT_REPO_DIR = os.path.join(_SCRIPT_DIR, "fanqie")
REPO_DIR = os.environ.get("FANQIE_REPO_DIR", _DEFAULT_REPO_DIR)
if not os.path.isdir(REPO_DIR):
    # 打包产物里没有 fanqie/ 子目录时，回退到开发者本机路径（仅开发用）
    _DEV_FALLBACK = r"D:\xingguang\fanqie-rank-mcp"
    if os.path.isdir(_DEV_FALLBACK):
        REPO_DIR = _DEV_FALLBACK
PORT = int(os.environ.get("FANQIE_BRIDGE_PORT", "19900"))
sys.path.insert(0, REPO_DIR)

from api import generate_a_bogus, decrypt_pua, normalize_cover_url, http_get, UA_RANK, fetch_leaderboards

RANK_API = "https://fanqienovel.com/api/rank/category/list"
APP_ID = "2503"


def _to_int(v, default=0):
    try:
        return int(str(v).strip())
    except (TypeError, ValueError):
        return default


async def fetch_books(gender, rank_mold, category_id, offset, limit):
    """请求番茄榜单接口并增强解析 book_list。"""
    query = (
        f"app_id={APP_ID}&rank_list_type=3&offset={offset}&limit={limit}"
        f"&category_id={category_id}&rank_version=&gender={gender}&rankMold={rank_mold}"
    )
    a_bogus = generate_a_bogus(query, UA_RANK)
    url = f"{RANK_API}?{query}&a_bogus={a_bogus}"
    body = await http_get(url, UA_RANK)
    data = json.loads(body)
    if data.get("code") != 0:
        raise RuntimeError(f"番茄接口错误: {data.get('message')}")

    items = data.get("data", {}).get("book_list") or []
    books = []
    for idx, it in enumerate(items):
        rank = _to_int(it.get("currentPos"), idx + 1) or (idx + 1)
        books.append({
            "bookId": str(it.get("book_id") or it.get("bookId") or ""),
            "rank": rank,
            "title": decrypt_pua(str(it.get("bookName") or it.get("book_name") or "")),
            "author": decrypt_pua(str(it.get("author") or it.get("author_name") or "")),
            "synopsis": decrypt_pua(str(it.get("abstract") or it.get("description") or "")),
            "coverUrl": normalize_cover_url(str(it.get("thumbUri") or it.get("thumb_url") or "")),
            "wordNumber": _to_int(it.get("wordNumber") or it.get("word_number")),
            "readCount": _to_int(it.get("read_count") or it.get("readCount")),
            "rankDiff": _to_int(it.get("rankPosDiff")),
            "creationStatus": _to_int(it.get("creationStatus")),
            "categoryId": str(it.get("curent_category_id") or it.get("category_id") or ""),
        })
    return books


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        try:
            parsed = urlparse(self.path)
            path = parsed.path
            qs = parse_qs(parsed.query)
            if path == "/health":
                self._json(200, {"ok": True})
            elif path == "/rankings":
                groups = asyncio.run(fetch_leaderboards())
                self._json(200, {"groups": groups})
            elif path == "/ranking":
                gender = _to_int(qs.get("gender", ["1"])[0], 1)
                rank_mold = _to_int(qs.get("rank_mold", ["2"])[0], 2)
                category_id = qs.get("category_id", ["1141"])[0] or "1141"
                offset = max(_to_int(qs.get("offset", ["0"])[0]), 0)
                limit = min(max(_to_int(qs.get("limit", ["30"])[0], 30), 1), 30)
                books = asyncio.run(fetch_books(gender, rank_mold, category_id, offset, limit))
                self._json(200, {"books": books})
            else:
                self._json(404, {"error": "not found"})
        except Exception as e:  # noqa: BLE001
            self._json(502, {"error": str(e)})

    def _json(self, code, obj):
        body = json.dumps(obj, ensure_ascii=False).encode("utf-8")
        self.send_response(code)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *args):  # 静默访问日志
        pass


if __name__ == "__main__":
    server = ThreadingHTTPServer(("127.0.0.1", PORT), Handler)
    print(f"fanqie bridge listening on http://127.0.0.1:{PORT}", flush=True)
    server.serve_forever()
