<template>
  <el-dialog
    :model-value="visible"
    title="敏感词检测"
    width="560px"
    append-to-body
    destroy-on-close
    class="xm-sw-dialog xm-dialog"
    @update:model-value="emit('update:visible', $event)"
  >
    <div class="xm-sw-body">
      <template v-if="!results.length">
        <div class="xm-sw-clean">
          <el-icon class="xm-sw-clean-icon"><CircleCheck /></el-icon>
          <p>{{ contentLength < 20 ? '正文内容太短' : '未检测到敏感词，内容安全' }}</p>
        </div>
      </template>
      <template v-else>
        <div class="xm-sw-summary">
          共检测到 <b>{{ totalCount }}</b> 处敏感词，涉及 <b>{{ results.length }}</b> 类：
        </div>
        <div v-for="group in results" :key="group.category" class="xm-sw-group">
          <div class="xm-sw-group-head">
            <span class="xm-sw-group-name" :style="{ color: group.color }">{{ group.category }}</span>
            <span class="xm-sw-group-count">{{ group.total }} 处</span>
          </div>
          <div class="xm-sw-words">
            <span
              v-for="w in group.words"
              :key="w.word"
              class="xm-sw-word"
              :title="`出现 ${w.count} 次，点击在正文中定位`"
              @click="emit('locate', w.word)"
            >{{ w.word }}<i>{{ w.count }}</i></span>
          </div>
        </div>
      </template>
      <p class="xm-sw-hint">点击词条可在正文中定位高亮 · 词库为常用网文审核参考，可自行扩充</p>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { CircleCheck } from '@element-plus/icons-vue'

defineOptions({ name: 'SensitiveWordPanel' })

const props = defineProps<{
  visible: boolean
  content: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'locate', word: string): void
}>()

// 内置参考词库（网文审核常用类别）
const LIBRARY: Array<{ category: string; color: string; words: string[] }> = [
  {
    category: '违禁敏感',
    color: '#e54d42',
    words: ['习近平', '毛泽东', '邓小平', '政府', '国家领导人', '公安局', '派出所', '拘留', '劳改', '死刑', '枪决', '游行', '示威', '颠覆', '分裂', '反动', '政变']
  },
  {
    category: '暴力血腥',
    color: '#d97706',
    words: ['砍杀', '虐杀', '肢解', '分尸', '爆头', '抽肠', '剥皮', '挖眼', '割舌', '血肉模糊', '肠子', '脑浆', '自残', '上吊', '吞枪']
  },
  {
    category: '色情低俗',
    color: '#c026d3',
    words: ['做爱', '性交', '淫水', '抽插', '肉棒', '阴道', '乳房', '高潮', '口交', '肛交', '强奸', '轮奸', '嫖娼', '卖淫', '春药', '淫荡']
  },
  {
    category: '广告导流',
    color: '#0ca678',
    words: ['加微信', '加QQ', '威信', 'vx', 'VX', '扣扣', '京东红包', '网赚', '兼职日结', '刷单', '私聊购买', '低价代充', '免费领取', '点击链接']
  },
  {
    category: '赌博毒品',
    color: '#4f6ef7',
    words: ['赌博', '赌场', '下注', '庄家', '冰毒', '摇头丸', '大麻', '海洛因', '吸毒', '贩毒', '罂粟', '快活林']
  }
]

const contentLength = computed(() => [...(props.content || '')].length)

const results = computed(() => {
  const text = props.content || ''
  if (!text) return []
  const out: Array<{ category: string; color: string; total: number; words: Array<{ word: string; count: number }> }> = []
  for (const group of LIBRARY) {
    const hits: Array<{ word: string; count: number }> = []
    let total = 0
    for (const w of group.words) {
      // 大小写不敏感统计
      const re = new RegExp(w.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi')
      const matches = text.match(re)
      if (matches && matches.length > 0) {
        hits.push({ word: w, count: matches.length })
        total += matches.length
      }
    }
    if (hits.length) {
      hits.sort((a, b) => b.count - a.count)
      out.push({ category: group.category, color: group.color, total, words: hits })
    }
  }
  return out
})

const totalCount = computed(() => results.value.reduce((s, g) => s + g.total, 0))
</script>

<style scoped>
.xm-sw-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 460px;
  overflow: auto;
}

.xm-sw-clean {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 36px 0;
}

.xm-sw-clean-icon {
  font-size: 42px;
  color: var(--el-color-success, #0ca678);
}

.xm-sw-clean p {
  margin: 0;
  font-size: 13px;
  color: var(--xm-ink-2, #5a6270);
}

.xm-sw-summary {
  font-size: 13px;
  color: var(--xm-ink, #374151);
}

.xm-sw-summary b {
  color: #e54d42;
}

.xm-sw-group {
  border: 1px solid var(--xm-line, #eef0f2);
  border-radius: 10px;
  padding: 10px 12px;
}

.xm-sw-group-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.xm-sw-group-name {
  font-size: 13px;
  font-weight: 600;
}

.xm-sw-group-count {
  font-size: 11.5px;
  color: var(--xm-ink-3, #a3a8b3);
}

.xm-sw-words {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.xm-sw-word {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--xm-ink, #374151);
  background: var(--xm-surface-2, #f3f5f7);
  border-radius: 6px;
  padding: 3px 8px;
  cursor: pointer;
  transition: all 0.15s var(--xm-ease, ease);
}

.xm-sw-word:hover {
  background: var(--xm-accent-soft, rgba(217, 119, 6, 0.12));
  color: var(--xm-accent, #d97706);
}

.xm-sw-word i {
  font-style: normal;
  font-size: 10px;
  color: var(--xm-ink-3, #a3a8b3);
}

.xm-sw-hint {
  margin: 0;
  font-size: 11px;
  color: var(--xm-ink-3, #a3a8b3);
}
</style>
