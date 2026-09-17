<!-- 手机端阅读预览 · 两个抽屉：目录 / 界面设置
     目录：按卷分组、当前章高亮并自动滚入视野；点击任意章直接跳转（nav 的 dir 实为偏移量）
     设置：排版 / 页面 / 主题 三组，每组 ≤7 项（对齐 DESIGN_GUIDELINES C1） -->
<template>
  <aside class="rx-sheet rx-sheet-toc" :class="{ 'is-on': open === 'toc' }">
    <div class="rx-sheet-head">
      <h4>目录</h4>
      <em>{{ tocMeta }}</em>
      <button class="x" type="button" title="收起" @click="emit('close')">✕</button>
    </div>
    <div ref="tocListRef" class="rx-toc-list">
      <template v-for="group in tocGroups" :key="group.title">
        <div v-if="group.title" class="rx-toc-vol">{{ group.title }}</div>
        <button
          v-for="row in group.rows"
          :key="row.index"
          type="button"
          class="rx-toc-row"
          :class="{ 'is-on': row.index === chapterIndex }"
          @click="emit('jump', row.index - chapterIndex)"
        >
          <span class="tno">{{ row.no }}</span>
          <span class="tt">{{ row.name }}</span>
          <span class="tw">{{ row.words ? row.words + ' 字' : '' }}</span>
        </button>
      </template>
      <div v-if="loading" class="rx-toc-vol">正在读取章节目录…</div>
      <div v-else-if="!tocGroups.length" class="rx-toc-vol">这本书还没有章节</div>
    </div>
  </aside>

  <aside class="rx-sheet rx-sheet-set" :class="{ 'is-on': open === 'set' }">
    <div class="rx-sheet-head">
      <h4>界面设置</h4>
      <button class="x" type="button" title="收起" @click="emit('close')">✕</button>
    </div>
    <div class="rx-set-tabs">
      <button
        v-for="tab in TABS"
        :key="tab.value"
        type="button"
        :class="{ 'is-on': activeTab === tab.value }"
        @click="activeTab = tab.value"
      >
        {{ tab.label }}
      </button>
    </div>

    <div class="rx-set-body">
      <!-- 排版 -->
      <template v-if="activeTab === 'type'">
        <div class="rx-srow">
          <span class="rx-slab">字号</span>
          <span class="grow">
            <div class="rx-step">
              <button type="button" @click="stepFont(-1)">−</button>
              <b>{{ settings.fontSize }}px</b>
              <button type="button" @click="stepFont(1)">+</button>
            </div>
          </span>
        </div>
        <div class="rx-srow">
          <span class="rx-slab">字体</span>
          <span class="grow">
            <div class="rx-seg">
              <button
                v-for="f in fontOptions"
                :key="f.value"
                type="button"
                :class="{ 'is-on': settings.font === f.value }"
                :style="{ fontFamily: f.stack }"
                @click="settings.font = f.value"
              >
                {{ f.label }}
              </button>
            </div>
          </span>
        </div>
        <div class="rx-srow">
          <span class="rx-slab">行距</span>
          <span class="grow">
            <div class="rx-seg">
              <button
                v-for="lv in LEVELS"
                :key="lv.value"
                type="button"
                :class="{ 'is-on': settings.lineHeight === lv.value }"
                @click="settings.lineHeight = lv.value"
              >
                {{ lv.label }}
              </button>
            </div>
          </span>
        </div>
        <div class="rx-srow">
          <span class="rx-slab">段距</span>
          <span class="grow">
            <div class="rx-seg">
              <button
                v-for="lv in LEVELS"
                :key="lv.value"
                type="button"
                :class="{ 'is-on': settings.paraGap === lv.value }"
                @click="settings.paraGap = lv.value"
              >
                {{ lv.label }}
              </button>
            </div>
          </span>
        </div>
        <div class="rx-srow">
          <span class="rx-slab">缩进</span>
          <span class="grow">
            <div class="rx-seg">
              <button type="button" :class="{ 'is-on': settings.indent }" @click="settings.indent = true">首行缩进</button>
              <button type="button" :class="{ 'is-on': !settings.indent }" @click="settings.indent = false">不缩进</button>
            </div>
          </span>
        </div>
        <div class="rx-srow">
          <span class="rx-slab">对齐</span>
          <span class="grow">
            <div class="rx-seg">
              <button type="button" :class="{ 'is-on': settings.align === 'justify' }" @click="settings.align = 'justify'">两端</button>
              <button type="button" :class="{ 'is-on': settings.align === 'left' }" @click="settings.align = 'left'">左对齐</button>
            </div>
          </span>
        </div>
        <div class="rx-srow">
          <span class="rx-slab">加粗</span>
          <span class="grow">
            <button type="button" class="rx-switch" :class="{ 'is-on': settings.bold }" @click="settings.bold = !settings.bold">
              <i />
            </button>
          </span>
        </div>
      </template>

      <!-- 页面 -->
      <template v-else-if="activeTab === 'page'">
        <div class="rx-srow">
          <span class="rx-slab">边距</span>
          <span class="grow">
            <div class="rx-seg">
              <button
                v-for="m in MARGINS"
                :key="m.value"
                type="button"
                :class="{ 'is-on': settings.margin === m.value }"
                @click="settings.margin = m.value"
              >
                {{ m.label }}
              </button>
            </div>
          </span>
        </div>
        <div class="rx-srow">
          <span class="rx-slab">翻页</span>
          <span class="grow">
            <div class="rx-seg">
              <button
                v-for="m in modeOptions"
                :key="m.value"
                type="button"
                :class="{ 'is-on': settings.mode === m.value }"
                @click="settings.mode = m.value"
              >
                {{ m.label }}
              </button>
            </div>
          </span>
        </div>
        <div class="rx-srow">
          <span class="rx-slab">点击区</span>
          <span class="grow">
            <div class="rx-seg">
              <button type="button" :class="{ 'is-on': settings.zones === 'page' }" @click="settings.zones = 'page'">翻页</button>
              <button type="button" :class="{ 'is-on': settings.zones === 'chapter' }" @click="settings.zones = 'chapter'">换章</button>
            </div>
          </span>
        </div>
        <div class="rx-srow">
          <span class="rx-slab">页眉脚</span>
          <span class="grow">
            <button
              type="button"
              class="rx-switch"
              :class="{ 'is-on': settings.chromeAlways }"
              @click="settings.chromeAlways = !settings.chromeAlways"
            >
              <i />
            </button>
          </span>
        </div>
        <div class="rx-srow">
          <span class="rx-slab">自动隐藏菜单</span>
          <span class="grow">
            <button
              type="button"
              class="rx-switch"
              :class="{ 'is-on': settings.autoHide }"
              @click="settings.autoHide = !settings.autoHide"
            >
              <i />
            </button>
          </span>
        </div>
        <div class="rx-srow">
          <span class="rx-slab">亮度</span>
          <span class="grow">
            <input v-model.number="settings.bright" class="rx-range" type="range" min="28" max="100" style="width: 130px" />
          </span>
        </div>
      </template>

      <!-- 主题 -->
      <template v-else>
        <div class="rx-srow">
          <span class="rx-slab">主题</span>
          <span class="grow">
            <div class="rx-swatches">
              <button
                v-for="(p, i) in paletteList"
                :key="p.name"
                type="button"
                :class="{ 'is-on': settings.palette === i, 'rx-swatch-night': p.night }"
                :style="{ background: p.bg, color: p.fg }"
                :title="p.name"
                @click="settings.palette = i"
              >
                文
              </button>
            </div>
          </span>
        </div>
        <div class="rx-srow">
          <span class="rx-slab">夜间</span>
          <span class="grow">
            <button type="button" class="rx-switch" :class="{ 'is-on': !!palette.night }" @click="toggleNight">
              <i />
            </button>
          </span>
        </div>
        <div class="rx-srow">
          <span class="rx-slab">跟随正文</span>
          <span class="grow">
            <button
              type="button"
              class="rx-switch"
              :class="{ 'is-on': settings.follow }"
              @click="settings.follow = !settings.follow"
            >
              <i />
            </button>
          </span>
        </div>
        <div class="rx-sgroup">当前主题：{{ palette.name }} · 翻页：{{ modeLabel }}</div>
        <div class="rx-srow">
          <span class="rx-slab">重置</span>
          <span class="grow">
            <button type="button" class="rx-reset" @click="reset">恢复默认设置</button>
          </span>
        </div>
      </template>
    </div>
  </aside>
</template>

<script setup lang="ts">
// 抽屉：目录 + 界面设置。目录数据按需拉取，失败就退回「第 N 章」列表（C4 空状态必达）
import { computed, nextTick, ref, watch } from 'vue'
import { chapterAPI, volumeAPI } from '@/api'
import { useBookStore } from '@/stores/book'
import { useReaderSettings, type Level, type MarginKey } from '@/composables/useReaderSettings'
import { splitChapterTitle, countWords } from '@/utils/readerText'
import type { Chapter, Volume } from '@/types'

const props = defineProps<{
  open: 'toc' | 'set' | null
  /** 当前章下标（整本书） */
  chapterIndex: number
  totalChapters: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
  /** 跳到某章：带的是相对偏移量，父组件直接交给 nav 事件 */
  (e: 'jump', delta: number): void
}>()

const { settings, palette, paletteList, fontOptions, modeOptions, modeLabel, toggleNight, reset } =
  useReaderSettings()

const TABS = [
  { value: 'type', label: '排版' },
  { value: 'page', label: '页面' },
  { value: 'theme', label: '主题' }
] as const
const activeTab = ref<'type' | 'page' | 'theme'>('type')

const LEVELS: { value: Level; label: string }[] = [
  { value: 'compact', label: '紧凑' },
  { value: 'std', label: '标准' },
  { value: 'loose', label: '宽松' }
]
const MARGINS: { value: MarginKey; label: string }[] = [
  { value: 'narrow', label: '窄' },
  { value: 'std', label: '标准' },
  { value: 'wide', label: '宽' }
]

function stepFont(delta: number) {
  settings.fontSize = Math.min(24, Math.max(12, settings.fontSize + delta))
}

/* ---------- 目录数据：按需拉取，缓存 ---------- */
const tocListRef = ref<HTMLElement | null>(null)
const chapters = ref<Chapter[]>([])
const volumes = ref<Volume[]>([])
const loading = ref(false)
let loaded = false

/** 书库可能还没就绪（甚至 Pinia 未安装），拿不到就当目录不可用，不影响其它设置 */
function currentBookId(): number | undefined {
  try {
    return useBookStore().currentBook?.id
  } catch {
    return undefined
  }
}

async function ensureToc() {
  if (loaded || loading.value) return
  const bookId = currentBookId()
  if (!bookId) return
  loading.value = true
  try {
    const [chapterRes, volumeRes] = await Promise.all([
      chapterAPI.getByBook(bookId),
      volumeAPI.getByBook(bookId)
    ])
    if (chapterRes.success && chapterRes.data) {
      chapters.value = chapterRes.data.filter((c) => c.type !== 'memo')
      loaded = true
    }
    if (volumeRes.success && volumeRes.data) volumes.value = volumeRes.data
  } catch {
    /* 拿不到就退回章节号列表 */
  } finally {
    loading.value = false
  }
}

interface TocRow {
  index: number
  no: string
  name: string
  words: number
}

const tocGroups = computed<{ title: string; rows: TocRow[] }[]>(() => {
  if (!chapters.value.length) {
    // 退化：只知道总章数时给一份「第 N 章」列表
    return [
      {
        title: '',
        rows: Array.from({ length: Math.max(0, props.totalChapters) }, (_, i) => ({
          index: i,
          no: `第 ${i + 1} 章`,
          name: i === props.chapterIndex ? '当前章' : '',
          words: 0
        }))
      }
    ]
  }
  const groups = new Map<string, TocRow[]>()
  chapters.value.forEach((c, i) => {
    const vol = volumes.value.find((v) => v.id === c.volume_id)
    const head = splitChapterTitle(c.title)
    const key = vol?.title || '未分卷'
    if (!groups.has(key)) groups.set(key, [])
    groups.get(key)!.push({
      index: i,
      no: head.no || `第 ${i + 1} 章`,
      name: head.name,
      words: countWords(c.content || '')
    })
  })
  return [...groups.entries()].map(([title, rows]) => ({ title, rows }))
})

const tocMeta = computed(() => {
  if (chapters.value.length) return `共 ${chapters.value.length} 章`
  return props.totalChapters ? `共 ${props.totalChapters} 章` : '目录'
})

/** 打开目录时滚到当前章 */
watch(
  () => [props.open, props.chapterIndex],
  async ([open]) => {
    if (open === 'toc') await ensureToc()
    if (open !== 'toc' || !tocListRef.value) return
    await nextTick()
    const on = tocListRef.value.querySelector('.rx-toc-row.is-on')
    on?.scrollIntoView({ block: 'center' })
  },
  { immediate: true }
)
</script>
