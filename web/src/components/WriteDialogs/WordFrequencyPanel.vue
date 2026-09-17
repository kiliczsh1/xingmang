<template>
  <el-dialog
    :model-value="visible"
    title="高频词透镜"
    width="520px"
    append-to-body
    destroy-on-close
    class="xm-wf-dialog"
    @update:model-value="emit('update:visible', $event)"
  >
    <div class="xm-wf-body">
      <p v-if="!words.length" class="xm-wf-empty">正文太短，暂无可统计的高频词（至少需要 200 字）</p>
      <div v-for="(w, i) in words" :key="w.word" class="xm-wf-row" @click="emit('locate', w.word)">
        <span class="xm-wf-rank">{{ i + 1 }}</span>
        <span class="xm-wf-word">{{ w.word }}</span>
        <div class="xm-wf-bar-track">
          <div class="xm-wf-bar" :style="{ width: (w.count / maxCount * 100) + '%' }"></div>
        </div>
        <span class="xm-wf-count">{{ w.count }} 次</span>
      </div>
      <p class="xm-wf-hint">点击词条可在正文中定位高亮 · 统计范围：2-4 字中文词组（去除常见虚词）</p>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'

defineOptions({ name: 'WordFrequencyPanel' })

const props = defineProps<{
  visible: boolean
  content: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'locate', word: string): void
}>()

const STOPWORDS = new Set([
  '一个', '这个', '那个', '什么', '没有', '自己', '他们', '我们', '你们', '不是', '就是', '还是',
  '但是', '可是', '如果', '因为', '所以', '已经', '现在', '时候', '知道', '觉得', '出来', '起来',
  '过去', '过来', '出来', '一下', '一样', '一直', '所有', '可以', '这么', '那么', '如何', '不过',
  '不能', '不会', '东西', '地方', '没有', '有些', '于是', '只是', '之后', '最后', '这些', '那些',
  '成了', '出现', '看着', '说道', '的话', '的话', '一声', '之中', '一个', '清声', '一般', '的选择'
])

const words = computed(() => {
  const text = (props.content || '')
    .replace(/```[\s\S]*?```/g, '')
    .replace(/<[^>]*>/g, '')
    .replace(/[^\u4e00-\u9fa5]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
  if (text.length < 200) return []
  const tokens = text.split(' ').filter(Boolean)
  const counts = new Map<string, number>()
  const add = (w: string) => {
    if (w.length < 2 || STOPWORDS.has(w)) return
    counts.set(w, (counts.get(w) || 0) + 1)
  }
  for (const t of tokens) {
    for (let n = 2; n <= 4; n++) {
      if (t.length < n) continue
      // 单 token 内滑窗
      for (let i = 0; i + n <= t.length; i++) add(t.slice(i, i + n))
    }
  }
  // 过滤：出现 >= 3 次；去掉是其他高频词子串的冗余项
  const list = [...counts.entries()]
    .filter(([, c]) => c >= 3)
    .sort((a, b) => b[1] - a[1] || b[0].length - a[0].length)
    .slice(0, 60)
  const kept: Array<{ word: string; count: number }> = []
  for (const [word, count] of list) {
    // 若是已保留词条的子串且次数相同，视为冗余
    const redundant = kept.some(k => k.word.includes(word) && k.count === count)
    if (!redundant) kept.push({ word, count })
    if (kept.length >= 30) break
  }
  return kept
})

const maxCount = computed(() => words.value[0]?.count || 1)
</script>

<style scoped>
.xm-wf-body {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 460px;
  overflow: auto;
}

.xm-wf-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 8px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.1s ease;
}

.xm-wf-row:hover {
  background: rgba(217, 119, 6, 0.1);
}

.xm-wf-rank {
  width: 22px;
  font-size: 11px;
  color: #a3a8b3;
  text-align: right;
  font-variant-numeric: tabular-nums;
  flex-shrink: 0;
}

.xm-wf-word {
  width: 96px;
  font-size: 13px;
  font-weight: 500;
  color: #374151;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex-shrink: 0;
}

.xm-wf-bar-track {
  flex: 1;
  height: 8px;
  background: #f0f1f3;
  border-radius: 4px;
  overflow: hidden;
}

.xm-wf-bar {
  height: 100%;
  background: linear-gradient(90deg, #f59e0b, #f97316);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.xm-wf-count {
  width: 48px;
  font-size: 11.5px;
  color: #8a8f99;
  text-align: right;
  font-variant-numeric: tabular-nums;
  flex-shrink: 0;
}

.xm-wf-empty {
  font-size: 13px;
  color: #a3a8b3;
  text-align: center;
  padding: 40px 0;
}

.xm-wf-hint {
  margin: 8px 0 0;
  font-size: 11px;
  color: #a3a8b3;
}

:root[data-theme='dark'] .xm-wf-word {
  color: #e6edf3;
}

:root[data-theme='dark'] .xm-wf-bar-track {
  background: rgba(255, 255, 255, 0.08);
}

:root[data-theme='dark'] .xm-wf-row:hover {
  background: rgba(217, 119, 6, 0.18);
}

:root[data-theme='dark'] .xm-wf-hint,
:root[data-theme='dark'] .xm-wf-count,
:root[data-theme='dark'] .xm-wf-rank,
:root[data-theme='dark'] .xm-wf-empty {
  color: #6b7688;
}
</style>
