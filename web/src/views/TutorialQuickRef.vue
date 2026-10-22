<template>
  <div class="qr-page">
    <header class="qr-head">
      <div class="qr-head-main">
        <h1>功能速查表</h1>
        <p>每个页面干什么、关键操作在哪。只写怎么用，不写原理。</p>
      </div>
      <div class="qr-search">
        <el-input v-model="keyword" placeholder="搜页面名或操作…" clearable>
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
    </header>

    <div v-if="filteredGroups.length === 0" class="qr-empty">
      <el-icon :size="28"><Search /></el-icon>
      <p>没找到「{{ keyword }}」相关的条目</p>
      <button class="qr-clear" type="button" @click="keyword = ''">清空搜索</button>
    </div>

    <section v-for="g in filteredGroups" :key="g.group" class="qr-group">
      <h2 class="qr-group-title">{{ g.group }}</h2>

      <div class="qr-cards">
        <article v-for="item in g.items" :key="item.name" class="qr-card">
          <div class="qr-card-head">
            <h3>{{ item.name }}</h3>
            <button class="qr-go" type="button" @click="goPage(item.path)">
              <span>去这个页面</span>
              <el-icon><TopRight /></el-icon>
            </button>
          </div>
          <p class="qr-what">{{ item.what }}</p>
          <ul class="qr-ops">
            <li v-for="(op, i) in item.ops" :key="i">{{ op }}</li>
          </ul>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Search, TopRight } from '@element-plus/icons-vue'
import { QUICKREF_GROUPS } from '@/data/quickref'

const keyword = ref('')

const filteredGroups = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  if (!q) return QUICKREF_GROUPS
  return QUICKREF_GROUPS
    .map(g => ({
      group: g.group,
      items: g.items.filter(item => {
        const hay = `${item.name} ${item.what} ${item.ops.join(' ')}`.toLowerCase()
        return hay.includes(q)
      })
    }))
    .filter(g => g.items.length > 0)
})

/**
 * ⚠ 用 window.open 而不是 router.push：教程中心是独立标签页，
 * 直接 router.push 会把教程页本身导航走，用户要看速查表就得后退回来。
 * 新开标签后两个窗口可以并排对照着操作。
 */
const goPage = (path: string) => {
  window.open(`${import.meta.env.BASE_URL.replace(/\/$/, '')}${path}`, '_blank', 'noopener')
}
</script>

<style scoped>
.qr-page {
  --qr-accent: var(--accent, #2563eb);
  --qr-accent-soft: var(--accent-soft, #eef4ff);
  --qr-card: var(--card, #ffffff);
  --qr-border: var(--content-border, #e6eaf0);
  --qr-bg: var(--content-background, #f4f6f8);
  --qr-text: var(--app-color, #1a2230);
  --qr-text-2: var(--ink2, #5a6a80);
  --qr-text-3: var(--ink3, #98a5b8);

  min-height: 100%;
  padding: 20px 24px 48px;
  background: var(--qr-bg);
  color: var(--qr-text);
}

.qr-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 20px;
  flex-wrap: wrap;
  margin-bottom: 22px;
}

.qr-head-main h1 {
  margin: 0 0 6px;
  font-size: 20px;
  font-weight: 600;
}

.qr-head-main p {
  margin: 0;
  font-size: 13px;
  color: var(--qr-text-3);
}

.qr-search {
  width: 260px;
  flex-shrink: 0;
}

.qr-group {
  margin-bottom: 26px;
}

.qr-group-title {
  margin: 0 0 12px;
  font-size: 13px;
  font-weight: 600;
  color: var(--qr-text-2);
  letter-spacing: .04em;
  padding-left: 8px;
  border-left: 3px solid var(--qr-accent);
  line-height: 1.2;
}

.qr-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 12px;
}

.qr-card {
  background: var(--qr-card);
  border: 1px solid var(--qr-border);
  border-radius: 12px;
  padding: 14px 16px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.qr-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.qr-card-head h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
}

.qr-go {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 26px;
  padding: 0 8px;
  border-radius: 6px;
  border: 1px solid var(--qr-border);
  background: transparent;
  color: var(--qr-text-3);
  font-size: 12px;
  cursor: pointer;
  white-space: nowrap;
  transition: background .15s, color .15s, border-color .15s;
}

.qr-go:hover {
  background: var(--qr-accent-soft);
  color: var(--qr-accent);
  border-color: var(--qr-accent);
}

.qr-what {
  margin: 0;
  font-size: 13px;
  color: var(--qr-text-2);
  line-height: 1.6;
}

.qr-ops {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.qr-ops li {
  position: relative;
  padding-left: 12px;
  font-size: 12.5px;
  line-height: 1.65;
  color: var(--qr-text-2);
}

.qr-ops li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 8px;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--qr-text-3);
}

.qr-empty {
  padding: 60px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: var(--qr-text-3);
  font-size: 13px;
}

.qr-clear {
  border: 1px solid var(--qr-border);
  background: transparent;
  color: var(--qr-accent);
  border-radius: 6px;
  padding: 4px 12px;
  font-size: 12px;
  cursor: pointer;
}
</style>

<style>
/* 暗色覆盖（非 scoped：data-theme 挂在 <html> 上，scoped 的 [data-v] 匹配不到） */
[data-theme='dark'] .qr-page {
  --qr-bg: #0f141a;
  --qr-card: #161b22;
  --qr-border: #21262d;
  --qr-text: #e6edf3;
  --qr-text-2: #9aa7b6;
  --qr-text-3: #6e7a89;
}
</style>
