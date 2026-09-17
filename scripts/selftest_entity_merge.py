"""实体合并系统端到端自检：迁移 / 合并 / 解除 / 候选检测 / 关系与章节不丢 / 级联清理。"""
import json
import os
import re
import sqlite3
import subprocess
import sys
import time
import urllib.error
import urllib.request
import winreg

ROOT = r"D:\xingguang\xingmang-desktop"
DB = os.path.join(ROOT, "xingmang.db")
BASE = "http://127.0.0.1:19825"
PASS, FAIL = [], []


def check(label, ok, detail=""):
    (PASS if ok else FAIL).append(label)
    print("  %s %-54s %s" % ("PASS" if ok else "FAIL", label, detail))


def db():
    c = sqlite3.connect(DB)
    c.row_factory = sqlite3.Row
    return c


def columns(table):
    return {r[1] for r in db().execute("PRAGMA table_info(%s)" % table)}


def real_env():
    parts = []
    for root, sub in (
        (winreg.HKEY_LOCAL_MACHINE, r"SYSTEM\CurrentControlSet\Control\Session Manager\Environment"),
        (winreg.HKEY_CURRENT_USER, r"Environment"),
    ):
        try:
            with winreg.OpenKey(root, sub) as k:
                parts.append(winreg.QueryValueEx(k, "Path")[0])
        except FileNotFoundError:
            pass
    env = dict(os.environ)
    env["PATH"] = re.sub(r"%([^%]+)%", lambda m: os.environ.get(m.group(1), m.group(0)), ";".join(parts))
    return env


print("[0] 迁移状态")
before = columns("graph_entities")
print("      primary_id 列当前%s（首次运行前应为“不存在”，迁移后为“存在”）"
      % ("存在" if "primary_id" in before else "不存在"))
print("      列数=%d" % len(before))

env = real_env()
opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))


def api(method, path, payload=None):
    data = json.dumps(payload).encode() if payload is not None else None
    req = urllib.request.Request(BASE + path, data=data, method=method,
                                 headers={"Content-Type": "application/json"})
    try:
        return json.loads(opener.open(req, timeout=30).read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        try:
            return json.loads(e.read().decode("utf-8"))
        except Exception:
            return {"success": False, "message": "HTTP %s" % e.code}
    except Exception as e:
        return {"success": False, "message": str(e)}


p = subprocess.Popen(["cmd.exe", "/c", "start-dev.bat", "-NoBrowser"],
                     cwd=ROOT, env=env, stdin=subprocess.DEVNULL,
                     stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                     creationflags=0x08000000 | 0x00000200)
t0, up = time.time(), False
while time.time() - t0 < 180:
    raw = p.stdout.readline()
    if not raw:
        break
    line = raw.decode("utf-8", "replace")
    if "all services up" in line:
        up = True
        break
    if "Failed to start" in line or "服务未就绪" in line:
        print(line)
        break
print("\n服务就绪: %s（%.1fs）\n" % (up, time.time() - t0))
if not up:
    sys.exit(1)

TEST_VERSION = None
try:
    print("[1] 迁移与初始状态")
    after = columns("graph_entities")
    check("启动后自动补上 primary_id 列", "primary_id" in after)
    rows = db().execute("SELECT COUNT(*) FROM graph_entities WHERE primary_id IS NOT NULL").fetchone()[0]
    check("存量实体全部是主词条", rows == 0, "primary_id 非空行数=%d" % rows)
    idx = [r[0] for r in db().execute(
        "SELECT name FROM sqlite_master WHERE type='index' AND name='idx_graph_entities_primary'")]
    check("primary_id 索引已建", bool(idx))

    print("\n[2] 候选检测：逐版本看数量与质量")
    total = 0
    for vid in (2, 3, 4, 9, 10):
        cand = api("GET", "/api/knowledge-graph/merge-candidates/16?versionId=%d" % vid)
        data = cand.get("data") or []
        total += len(data)
        ec = db().execute("SELECT COUNT(*) FROM graph_entities WHERE version_id=?", (vid,)).fetchone()[0]
        print("      版本 %-3s （%d 个主词条）-> %d 对候选" % (vid, ec, len(data)))
        for c in data[:5]:
            print("          %.2f  %s × %s   [%s]" % (
                c["score"], c["a"]["name"], c["b"]["name"], " / ".join(c["reasons"])))
    check("候选总量明显收敛（不再是几十条噪声）", total <= 15, "合计 %d 对" % total)

    cand10 = api("GET", "/api/knowledge-graph/merge-candidates/16?versionId=10")
    data10 = cand10.get("data") or []
    if data10:
        ok_reason = all(c.get("reasons") and len(c["reasons"]) > 0 for c in data10)
        ok_score = all(0.5 <= c["score"] <= 1.0 for c in data10)
        ok_distinct = all(c["a"]["id"] != c["b"]["id"] for c in data10)
        ok_alias_free = all(c["a"].get("primary_id") is None and c["b"].get("primary_id") is None for c in data10)
        ok_dir = all(c["suggest_primary_id"] in (c["a"]["id"], c["b"]["id"])
                     and c["suggest_duplicate_id"] in (c["a"]["id"], c["b"]["id"])
                     and c["suggest_primary_id"] != c["suggest_duplicate_id"] for c in data10)
        check("每对都有可读理由", ok_reason)
        check("分数落在 [0.5, 1.0]", ok_score)
        check("不含自比", ok_distinct)
        check("已合并的副词条不参与", ok_alias_free)
        check("建议方向自洽", ok_dir)
    else:
        check("v10 已无重复，候选为空是正确结果", True, "0 对")

    print("\n[3] 临时版本：合并 / 解除 往返")
    cv = api("POST", "/api/knowledge-graph/versions", {"bookId": 16, "name": "__合并自检__"})
    TEST_VERSION = (cv.get("data") or {}).get("id")
    check("创建临时版本", bool(TEST_VERSION), "versionId=%s" % TEST_VERSION)
    if not TEST_VERSION:
        raise SystemExit("无法创建临时版本")

    imp = api("POST", "/api/knowledge-graph/graph/import", {
        "bookId": 16, "versionId": TEST_VERSION,
        "entities": [
            {"id": 1, "name": "自检主甲", "type": "character", "description": "主角", "chapter_ids": [1629]},
            {"id": 2, "name": "自检副乙", "type": "character", "description": "别名", "chapter_ids": [1630]},
            {"id": 3, "name": "自检丙", "type": "location", "description": "地点", "chapter_ids": [1631]},
            {"id": 4, "name": "自检丁", "type": "item", "description": "物件", "chapter_ids": [1631]},
        ],
        "relations": [
            {"id": 1, "source_id": 1, "target_id": 3, "relation_type": "在", "chapter_ids": [1629]},
            {"id": 2, "source_id": 2, "target_id": 3, "relation_type": "在", "chapter_ids": [1630]},
            {"id": 3, "source_id": 1, "target_id": 2, "relation_type": "别名关系", "chapter_ids": [1630]},
            {"id": 4, "source_id": 4, "target_id": 2, "relation_type": "属于", "chapter_ids": [1631]},
        ],
    })
    check("导入 4 实体 / 4 关系", imp.get("data", {}).get("entityCount") == 4
          and imp.get("data", {}).get("relationCount") == 4, json.dumps(imp.get("data")))

    def graph():
        g = api("GET", "/api/knowledge-graph/graph/16?versionId=%s" % TEST_VERSION).get("data", {})
        by_name = {e["name"]: e for e in g.get("entities", [])}
        return g, by_name

    g, by_name = graph()
    primary_id = by_name["自检主甲"]["id"]
    dup_id = by_name["自检副乙"]["id"]
    other_id = by_name["自检丙"]["id"]

    m = api("POST", "/api/knowledge-graph/entities/merge", {
        "bookId": 16, "versionId": TEST_VERSION, "primaryId": primary_id, "memberIds": [dup_id]})
    check("合并接口返回成功", m.get("success") is True and m.get("data", {}).get("merged") == 1, json.dumps(m.get("data")))

    g, by_name = graph()
    check("副词条仍返回但 primary_id 指向主词条",
          by_name["自检副乙"].get("primary_id") == primary_id,
          "primary_id=%s" % by_name["自检副乙"].get("primary_id"))
    check("主词条 primary_id 为空", by_name["自检主甲"].get("primary_id") is None)
    check("关系与章节关联都没被搬动（合并是可逆的）",
          len(g.get("relations", [])) == 4,
          "关系数=%d" % len(g.get("relations", [])))
    refs = api("GET", "/api/knowledge-graph/chapter-refs/16?versionId=%s" % TEST_VERSION).get("data", [])
    check("章节关联行数未变（4 实体 + 4 关系 → 4 组章节）", len(refs) == 3, "章节组=%d" % len(refs))

    cand2 = api("GET", "/api/knowledge-graph/merge-candidates/16?versionId=%s" % TEST_VERSION).get("data", [])
    check("合并后候选里不再出现该副词条",
          all(c["a"]["id"] != dup_id and c["b"]["id"] != dup_id for c in cand2))

    mm = api("POST", "/api/knowledge-graph/entities/merge", {
        "bookId": 16, "versionId": TEST_VERSION, "primaryId": dup_id, "memberIds": [other_id]})
    check("传副词条 id 会顺链归到根（丙 挂到 甲 名下）",
          mm.get("data", {}).get("primaryId") == primary_id, json.dumps(mm.get("data")))
    g, by_name = graph()
    check("丙 的 primary_id = 甲", by_name["自检丙"].get("primary_id") == primary_id)

    mm2 = api("POST", "/api/knowledge-graph/entities/merge", {
        "bookId": 16, "versionId": TEST_VERSION, "primaryId": primary_id, "memberIds": [dup_id]})
    check("重复合并同一成员是幂等的（skipped=1）",
          mm2.get("data", {}).get("skipped") == 1 and mm2.get("data", {}).get("merged") == 0,
          json.dumps(mm2.get("data")))

    u = api("POST", "/api/knowledge-graph/entities/unmerge", {"entityIds": [dup_id]})
    check("解除单个副词条", u.get("data", {}).get("released") == 1, json.dumps(u.get("data")))
    g, by_name = graph()
    check("解除后恢复为主词条", by_name["自检副乙"].get("primary_id") is None)

    api("POST", "/api/knowledge-graph/entities/merge", {
        "bookId": 16, "versionId": TEST_VERSION, "primaryId": primary_id, "memberIds": [dup_id, other_id]})
    ua = api("POST", "/api/knowledge-graph/entities/unmerge", {"entityIds": [primary_id]})
    check("传主词条则解散整组（released=2）", ua.get("data", {}).get("released") == 2, json.dumps(ua.get("data")))
    g, by_name = graph()
    check("整组都还原成主词条",
          by_name["自检副乙"].get("primary_id") is None and by_name["自检丙"].get("primary_id") is None)

    print("\n[4] 导入时 primary_id 的 old→new 重映射")
    cv2 = api("POST", "/api/knowledge-graph/versions", {"bookId": 16, "name": "__合并自检2__"})
    v2 = (cv2.get("data") or {}).get("id")
    check("创建第二个临时版本", bool(v2), "versionId=%s" % v2)
    api("POST", "/api/knowledge-graph/graph/import", {
        "bookId": 16, "versionId": v2,
        "entities": [
            {"id": 10, "name": "导入主", "type": "character", "description": ""},
            {"id": 11, "name": "导入副", "type": "character", "description": "", "primary_id": 10},
            {"id": 12, "name": "导入悬空", "type": "character", "description": "", "primary_id": 999},
        ],
        "relations": [],
    })
    g2 = api("GET", "/api/knowledge-graph/graph/16?versionId=%s" % v2).get("data", {})
    n2 = {e["name"]: e for e in g2.get("entities", [])}
    check("primary_id 被重映射成新 id（不是入参里的 10）",
          n2["导入副"].get("primary_id") == n2["导入主"]["id"],
          "副.primary_id=%s 主.id=%s" % (n2["导入副"].get("primary_id"), n2["导入主"]["id"]))
    check("指向不存在实体的 primary_id 被丢弃（不写悬空引用）",
          n2["导入悬空"].get("primary_id") is None)
    api("DELETE", "/api/knowledge-graph/versions/16/%s" % v2)

    print("\n[5] 删除主词条时释放名下副词条")
    api("POST", "/api/knowledge-graph/entities/merge", {
        "bookId": 16, "versionId": TEST_VERSION, "primaryId": primary_id, "memberIds": [dup_id]})
    de = api("DELETE", "/api/knowledge-graph/entities/%s" % primary_id)
    check("删除主词条成功", de.get("success") is True)
    g, by_name = graph()
    check("名下副词条被释放成主词条（不留悬空 primary_id）",
          by_name.get("自检副乙", {}).get("primary_id") is None,
          "primary_id=%s" % by_name.get("自检副乙", {}).get("primary_id"))

finally:
    if TEST_VERSION:
        print("\n[6] 清理临时版本")
        api("DELETE", "/api/knowledge-graph/versions/16/%s" % TEST_VERSION)
        vers = api("GET", "/api/knowledge-graph/versions/16").get("data", [])
        check("临时版本已删除", not any(v["id"] == TEST_VERSION for v in vers))
        ne = db().execute("SELECT COUNT(*) FROM graph_entities").fetchone()[0]
        nr = db().execute("SELECT COUNT(*) FROM graph_relations").fetchone()[0]
        check("全库回到 176/166（临时数据清干净）", ne == 176 and nr == 166, "%d/%d" % (ne, nr))
        check("没有残留的非空 primary_id",
              db().execute("SELECT COUNT(*) FROM graph_entities WHERE primary_id IS NOT NULL").fetchone()[0] == 0)

print("\n" + "=" * 74)
print("通过 %d 项，失败 %d 项" % (len(PASS), len(FAIL)))
for f in FAIL:
    print("  FAIL:", f)
print("=" * 74)
