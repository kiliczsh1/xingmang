"""Stub module: 原子写入 JSON 文件。"""
import json
import os


def atomic_write_json(path, data, old_content=None, ensure_ascii=False, indent=2):
    os.makedirs(os.path.dirname(os.path.abspath(str(path))), exist_ok=True)
    with open(str(path), 'w', encoding='utf-8') as f:
        json.dump(data, f, ensure_ascii=ensure_ascii, indent=indent)
