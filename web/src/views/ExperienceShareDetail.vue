<template>
  <div ref="pageRef" class="wiki-page" v-loading="loading">
    <!-- 顶栏 -->
    <div class="wiki-topbar">
      <div class="wiki-logo">星芒<span>百科</span></div>
      <span class="wiki-tagline">自由的星芒经验库</span>
      <div class="wiki-search">
        <input :value="experienceShare?.title || ''" placeholder="搜索经验词条…" readonly />
        <button>搜索</button>
      </div>
      <div class="wiki-actions">
        <button class="wbtn" @click="goBack">← 返回</button>
        <button class="wbtn" @click="metaDialogVisible = true">编辑信息</button>
        <span v-if="saveStatus" class="save-status">{{ saveStatus }}</span>
        <button
          v-if="experienceShare?.pdf_file_url"
          class="wbtn primary"
          @click="openPdfPreview(experienceShare.pdf_file_url, experienceShare.pdf_file_name || 'PDF附件')"
        >
          查看附件
        </button>
      </div>
    </div>

    <div v-if="experienceShare" class="wiki-wrap" :class="{ 'no-toc': activeTab !== 'article' }">
      <!-- 左栏：目录框（仅条目页显示） -->
      <aside v-if="activeTab === 'article'" class="wiki-left">
        <div class="toc-box">
          <div class="toc-h">目录</div>
          <a
            v-for="item in tocItems"
            :key="item.id"
            class="toc-link"
            :class="{ lv2: item.level === 2 }"
            :href="`#${item.id}`"
            @click.prevent="scrollToHeading(item.id)"
          >
            <span v-if="item.num" class="n">{{ item.num }}</span>{{ item.text }}
          </a>
          <a class="toc-link lv2" href="#att" @click.prevent="goToArticleSection('att')">附件</a>
          <a v-if="recommendedShares.length" class="toc-link lv2" href="#rel" @click.prevent="goToArticleSection('rel')">相关条目</a>
          <a class="toc-link lv2" href="#discuss" @click.prevent="switchTab('discuss')">讨论区</a>
        </div>
        <div class="help-box">📖 词条由社区维护。如需引用，请注明「星芒百科 · 经验分享」。</div>
      </aside>

      <!-- 右栏：词条主体 -->
      <main class="wiki-main">
        <div class="page-head">
          <h1>{{ experienceShare.title }}</h1>
          <div class="tabs">
            <span :class="{ active: activeTab === 'article' }" @click="switchTab('article')">条目</span>
            <span :class="{ active: activeTab === 'discuss' }" @click="switchTab('discuss')">讨论</span>
            <span :class="{ active: activeTab === 'history' }" @click="switchTab('history')">历史</span>
            <span :class="{ active: activeTab === 'log' }" @click="switchTab('log')">日志</span>
          </div>
          <p v-if="experienceShare.summary" class="summary">{{ experienceShare.summary }}</p>
        </div>

        <div v-if="activeTab === 'article'" ref="articleRef" class="article">
          <!-- 信息框 -->
          <div class="infobox">
            <div class="ib-h">经验词条 · 基本信息</div>
            <div class="ib-img">
              <img v-if="experienceShare.cover_url" :src="experienceShare.cover_url" :alt="experienceShare.title" />
              <div v-else class="ib-img-ph">
                <div class="sun"></div>
                <div class="m"></div>
              </div>
            </div>
            <div class="ib-row"><span class="k">作者</span><span class="v">{{ experienceShare.author_name || '星芒用户' }}</span></div>
            <div v-if="experienceShare.version" class="ib-row"><span class="k">版本号</span><span class="v">{{ experienceShare.version }}</span></div>
            <div class="ib-row"><span class="k">创建时间</span><span class="v">{{ formatDateTime(experienceShare.created_at) }}</span></div>
            <div class="ib-row">
              <span class="k">创建方式</span>
              <span class="v">
                <span class="tag" :class="{ pdf: experienceShare.create_type === 'pdf_import' }">
                  {{ experienceShare.create_type === 'pdf_import' ? 'PDF 导入' : '手动创建' }}
                </span>
              </span>
            </div>
            <div class="ib-row"><span class="k">状态</span><span class="v">{{ experienceShare.status || 'published' }}</span></div>
            <div
              v-if="experienceShare.updated_at && experienceShare.updated_at !== experienceShare.created_at"
              class="ib-row"
            >
              <span class="k">更新时间</span><span class="v">{{ formatDateTime(experienceShare.updated_at) }}</span>
            </div>
            <div v-if="experienceShare.pdf_file_url" class="ib-row">
              <span class="k">附件</span>
              <span class="v">{{ experienceShare.pdf_file_name || 'PDF附件' }}</span>
            </div>
          </div>

          <!-- 题图（非 PDF 导入时展示封面） -->
          <figure v-if="experienceShare.cover_url && !shouldShowPdfBody" class="cover-fig">
            <img :src="experienceShare.cover_url" :alt="experienceShare.title" />
            <figcaption>▲ 词条题图</figcaption>
          </figure>

          <!-- 正文 -->
          <div v-if="shouldShowPdfBody" class="pdf-body">
            <PdfPageGallery v-if="experienceShare.pdf_file_url" :src="experienceShare.pdf_file_url" :scale="1.25" />
          </div>

          <div v-else class="article-body">
            <TipTapEditor
              v-model="experienceShare.content"
              placeholder="点击开始编辑正文…"
              ghost
              class="wiki-editor"
              @blur="handleContentBlur"
            />
          </div>

          <!-- 补充说明 -->
          <div v-if="shouldShowPdfBody" class="notice">
            <b>补充说明：</b>
            <TipTapEditor
              v-model="experienceShare.content"
              placeholder="补充说明…"
              ghost
              class="wiki-editor notice-editor"
              @blur="handleContentBlur"
            />
          </div>

          <!-- 附件框 -->
          <div v-if="experienceShare.pdf_file_url" id="att" class="att-box">
            <div class="ab-h">附件</div>
            <div class="att">
              <div class="ic">PDF</div>
              <div class="att-info">
                <div class="nm">{{ experienceShare.pdf_file_name || 'PDF附件' }}</div>
                <div class="sz">{{ formatFileSize(experienceShare.pdf_file_size) }}</div>
              </div>
              <div class="btns">
                <a @click="openPdfPreview(experienceShare.pdf_file_url, experienceShare.pdf_file_name || 'PDF附件')">预览</a>
                <a @click="downloadPdf(experienceShare.pdf_file_url, experienceShare.pdf_file_name || 'experience-share.pdf')">下载</a>
              </div>
            </div>
          </div>

          <!-- 相关条目 -->
          <div v-if="recommendedShares.length" id="rel" class="rel-section">
            <h2>相关条目</h2>
            <div class="related">
              <a
                v-for="item in recommendedShares"
                :key="item.id"
                class="rel-item"
                @click="goToRecommended(item.id)"
              >
                <div class="rc">
                  <img v-if="item.cover_url" :src="item.cover_url" :alt="item.title" />
                  <div v-else class="g ph"></div>
                </div>
                <div class="rt">{{ item.title }}</div>
              </a>
            </div>
          </div>

          <div class="catline">分类：写作方法论 · 知识管理 ｜ 本词条由星芒社区维护</div>
        </div>

        <!-- 讨论区（BBS 格式 · 本地存储） -->
        <section v-if="activeTab === 'discuss'" class="bbs-section" id="discuss">
          <div class="bbs-header">
            <div class="bbs-board">版块 · 经验分享</div>
            <div class="bbs-stats">主题：1 · 回复：{{ bbsReplies.length }} · 今日：{{ todayCount }}</div>
          </div>

          <!-- 空态：暂无回复 -->
          <div v-if="!bbsReplies.length" class="bbs-empty">
            <span>暂无回复，快来抢沙发 🛋️</span>
          </div>

          <!-- 回复楼层 -->
          <div v-for="(reply, index) in bbsReplies" :key="reply.id" class="bbs-post">
            <div class="post-side">
              <div class="avatar" :style="avatarStyle(reply.nickname)">{{ reply.nickname.charAt(0) }}</div>
              <div class="username">{{ reply.nickname }}</div>
              <span v-if="reply.mine" class="badge mine">我</span>
            </div>
            <div class="post-main">
              <div class="post-meta">
                <span class="floor">{{ index + 1 }}楼</span>
                <span class="time">{{ reply.createdAt }}</span>
              </div>
              <div v-if="reply.quote" class="quote-block">
                引用 @{{ reply.quote.nickname }} 的发言：{{ reply.quote.content }}
              </div>
              <div v-if="editingId === reply.id" class="post-content">
                <textarea v-model="editingText" class="edit-textarea" rows="3" placeholder="编辑回复内容…"></textarea>
              </div>
              <div v-else class="post-content" v-html="renderContent(reply.content)"></div>
              <div v-if="editingId === reply.id" class="post-actions">
                <button class="save" @click="saveEdit(reply)">保存</button>
                <button @click="cancelEdit">取消</button>
              </div>
              <div v-else class="post-actions">
                <button @click="quoteReply(reply)">引用</button>
                <button @click="focusReply">回复</button>
                <button :class="{ liked: reply.liked }" @click="toggleLike(reply)">👍 {{ reply.likes }}</button>
                <button class="edit" @click="startEdit(reply)">编辑</button>
                <button class="del" @click="deleteReply(reply)">删除</button>
              </div>
            </div>
          </div>

          <!-- 回复框 -->
          <div class="bbs-reply-box" id="bbs-reply-box">
            <div class="reply-toolbar">
              <span class="nick-label">昵称：<input v-model="nickname" maxlength="12" placeholder="输入昵称" /></span>
              <div class="emoticons">
                <button v-for="e in emoticons" :key="e.code" :title="e.code" @click="insertEmoticon(e.code)">{{ e.emoji }}</button>
              </div>
            </div>
            <div v-if="replyQuote" class="reply-quote">
              <span>引用 @{{ replyQuote.nickname }}：{{ replyQuote.content }}</span>
              <button class="x" @click="clearQuote">×</button>
            </div>
            <textarea v-model="replyText" placeholder="输入回复内容，支持 BBS 表情…"></textarea>
            <div class="reply-actions">
              <span class="hint">本地存储 · 仅本机可见</span>
              <button class="submit" @click="submitReply">发表回复</button>
            </div>
          </div>
        </section>

        <!-- 历史：版本历史 -->
        <div v-if="activeTab === 'history'" class="tab-panel">
          <h2>版本历史</h2>
          <div class="history-list">
            <div v-for="(item, i) in historyItems" :key="i" class="history-item">
              <div class="h-version">{{ item.label }}</div>
              <div class="h-info">
                <div class="h-time">{{ item.time }}</div>
                <div class="h-desc">{{ item.desc }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- 日志：操作日志 -->
        <div v-if="activeTab === 'log'" class="tab-panel">
          <h2>操作日志</h2>
          <div class="log-list">
            <div v-for="(item, i) in logItems" :key="i" class="log-item">
              <span class="log-time">{{ item.time }}</span>
              <span class="log-action">{{ item.action }}</span>
              <span class="log-user">{{ item.user }}</span>
            </div>
          </div>
        </div>
      </main>
    </div>

    <el-empty v-else-if="!loading" description="未找到这张经验卡片">
      <el-button type="primary" @click="goBack">返回列表</el-button>
    </el-empty>

    <button
      v-show="showBackToTop"
      type="button"
      class="backtop-float"
      aria-label="回到顶部"
      @click="scrollToTop"
    >
      <span class="backtop-icon">↑</span>
      <span class="backtop-text">回到顶部</span>
    </button>

    <el-dialog
      v-model="pdfPreviewVisible"
      title="PDF 在线预览"
      width="min(1180px, 96vw)"
      top="2vh"
      destroy-on-close
    >
      <div class="pdf-preview-shell">
        <div class="pdf-preview-toolbar">
          <div class="pdf-preview-file">
            <el-icon><Document /></el-icon>
            <span>{{ pdfPreviewFileName }}</span>
          </div>
          <div class="pdf-preview-actions">
            <el-button v-if="pdfPreviewUrl" @click="openPdfInNewTab">新窗口打开</el-button>
            <el-button v-if="pdfPreviewUrl" type="primary" @click="downloadCurrentPdf">下载</el-button>
          </div>
        </div>
        <div class="pdf-preview-body">
          <PdfPageGallery v-if="pdfPreviewUrl" :src="pdfPreviewUrl" :scale="1.45" />
        </div>
      </div>
    </el-dialog>

    <ExperienceShareMetaDialog
      v-model="metaDialogVisible"
      :share="experienceShare"
      @saved="handleMetaSaved"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { experienceShareAPI } from '@/api'
import type { ExperienceShare } from '@/types'
import TipTapEditor from '@/components/TipTapEditor.vue'
import ExperienceShareMetaDialog from '@/components/ExperienceShareMetaDialog.vue'
import PdfPageGallery from '@/components/PdfPageGallery.vue'

type TocItem = {
  id: string
  index: number
  text: string
  level: 1 | 2
  num: number | null
}

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const experienceShare = ref<ExperienceShare | null>(null)
const recommendedShares = ref<ExperienceShare[]>([])
const pdfPreviewVisible = ref(false)
const pdfPreviewUrl = ref('')
const pdfPreviewFileName = ref('')
const pdfPreviewDownloadName = ref('')
const showBackToTop = ref(false)
const metaDialogVisible = ref(false)
const saveStatus = ref('')
const savingContent = ref(false)
let saveStatusTimer: ReturnType<typeof setTimeout> | null = null

const pageRef = ref<HTMLElement | null>(null)
const articleRef = ref<HTMLElement | null>(null)
const tocItems = ref<TocItem[]>([])

let scrollContainer: HTMLElement | Window | null = null

const shouldShowPdfBody = computed(() => {
  if (!experienceShare.value) return false
  return experienceShare.value.create_type === 'pdf_import' && Boolean(experienceShare.value.pdf_file_url)
})

const formatDateTime = (value?: string | null) => {
  if (!value) return '未知时间'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatFileSize = (size?: number) => {
  const value = Number(size || 0)
  if (value <= 0) return '未知大小'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  return `${(value / 1024 / 1024).toFixed(2)} MB`
}

/* ---------- 目录：从渲染后的 markdown 标题生成 ---------- */
const buildToc = () => {
  tocItems.value = []
  const body = articleRef.value
  if (!body) return
  // 正文与补充说明都是 TipTap 编辑器，标题在 .tiptap 内
  const headings = body.querySelectorAll('.tiptap h2, .tiptap h3')
  let h2Count = 0
  headings.forEach((heading, index) => {
    const el = heading as HTMLElement
    const isH2 = el.tagName === 'H2'
    if (isH2) h2Count += 1
    tocItems.value.push({
      id: `wiki-sec-${index}`,
      index,
      text: el.textContent?.trim() || '',
      level: isH2 ? 1 : 2,
      num: isH2 ? h2Count : null
    })
  })
}

/* ---------- 目录：跟随正文 DOM 变化重建 ---------- */
let tocObserver: MutationObserver | null = null
let tocRaf = 0

const scheduleBuildToc = () => {
  if (tocRaf) cancelAnimationFrame(tocRaf)
  tocRaf = requestAnimationFrame(() => {
    tocRaf = 0
    buildToc()
  })
}

watch(
  articleRef,
  (el) => {
    tocObserver?.disconnect()
    tocObserver = null
    if (!el) {
      tocItems.value = []
      return
    }
    // TipTap 在挂载后才渲染出 .tiptap 正文（切换 Tab 会重建编辑器），
    // 仅靠 content 变化触发会漏掉首屏渲染与重新挂载的时机，导致目录一直为空。
    // 这里改为跟随正文 DOM 变化重建目录。
    buildToc()
    tocObserver = new MutationObserver(scheduleBuildToc)
    tocObserver.observe(el, { childList: true, subtree: true, characterData: true })
  },
  { immediate: true }
)

const scrollToHeading = (id: string) => {
  const item = tocItems.value.find(t => t.id === id)
  if (!item) return
  // 编辑器重渲染会丢 id，按索引定位更可靠
  const headings = articleRef.value?.querySelectorAll('.tiptap h2, .tiptap h3')
  const target = headings?.[item.index] as HTMLElement | undefined
  target?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

const scrollToId = (id: string) => {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

/* ---------- 正文自动保存（blur 触发，后端全量覆盖，必须带全字段） ---------- */
const buildPayload = () => {
  const s = experienceShare.value
  if (!s) return null
  return {
    title: s.title.trim(),
    summary: s.summary?.trim() || '',
    content: s.content || '',
    content_render_mode: 'html',
    cover_url: s.cover_url || '',
    create_type: s.create_type,
    author_id: s.author_id ?? null,
    author_name: s.author_name || '',
    version: s.version || '',
    status: s.status || 'published',
    pdf_file_url: s.pdf_file_url || null,
    pdf_file_name: s.pdf_file_name || null,
    pdf_file_size: s.pdf_file_size || 0,
    pdf_parse_status: s.pdf_parse_status || null,
    pdf_parse_result: s.pdf_parse_result || null,
    source_file_name: s.source_file_name || null
  }
}

const setSaveStatus = (text: string) => {
  saveStatus.value = text
  if (saveStatusTimer) clearTimeout(saveStatusTimer)
  saveStatusTimer = setTimeout(() => { saveStatus.value = '' }, 2000)
}

const handleContentBlur = async () => {
  const s = experienceShare.value
  if (!s || savingContent.value) return
  const payload = buildPayload()
  if (!payload) return
  savingContent.value = true
  setSaveStatus('保存中…')
  try {
    const res = await experienceShareAPI.update(s.id, payload)
    if (res.success && res.data) {
      experienceShare.value = res.data
      setSaveStatus('已保存')
    } else {
      setSaveStatus('保存失败')
    }
  } catch {
    setSaveStatus('保存失败')
  } finally {
    savingContent.value = false
  }
}

const handleMetaSaved = (data: ExperienceShare) => {
  experienceShare.value = data
}

/* ---------- 顶部 Tab 切换（条目 / 讨论 / 历史 / 日志） ---------- */
type WikiTab = 'article' | 'discuss' | 'history' | 'log'

const activeTab = ref<WikiTab>('article')

const switchTab = (tab: WikiTab) => {
  activeTab.value = tab
  nextTick(() => {
    const el = scrollContainer || window
    if (el === window) {
      window.scrollTo({ top: 0, behavior: 'smooth' })
    } else {
      (el as HTMLElement).scrollTo({ top: 0, behavior: 'smooth' })
    }
  })
}

const goToArticleSection = (id: string) => {
  if (activeTab.value !== 'article') {
    activeTab.value = 'article'
    nextTick(() => scrollToId(id))
  } else {
    scrollToId(id)
  }
}

const historyItems = computed(() => {
  const share = experienceShare.value
  if (!share) return []
  const items: { label: string; time: string; desc: string }[] = []
  if (share.updated_at && share.updated_at !== share.created_at) {
    items.push({ label: share.version || '当前版本', time: formatDateTime(share.updated_at), desc: '内容更新' })
  }
  items.push({
    label: '创建',
    time: formatDateTime(share.created_at),
    desc: share.create_type === 'pdf_import' ? 'PDF 导入创建' : '手动创建'
  })
  return items
})

const logItems = computed(() => {
  const share = experienceShare.value
  if (!share) return []
  const items: { time: string; action: string; user: string }[] = []
  if (share.updated_at && share.updated_at !== share.created_at) {
    items.push({ time: formatDateTime(share.updated_at), action: '更新词条内容', user: share.author_name || '星芒用户' })
  }
  items.push({
    time: formatDateTime(share.created_at),
    action: share.create_type === 'pdf_import' ? '导入 PDF 并创建词条' : '手动创建词条',
    user: share.author_name || '星芒用户'
  })
  return items
})

/* ---------- 讨论区（BBS 格式 · 本地存储） ---------- */
type BbsReply = {
  id: number
  nickname: string
  content: string
  quote: { nickname: string; content: string } | null
  likes: number
  liked: boolean
  mine: boolean
  createdAt: string
}

const emoticons = [
  { code: '[微笑]', emoji: '😊' },
  { code: '[大笑]', emoji: '😄' },
  { code: '[流泪]', emoji: '😭' },
  { code: '[生气]', emoji: '😠' },
  { code: '[惊讶]', emoji: '😲' },
  { code: '[赞]', emoji: '👍' },
  { code: '[汗]', emoji: '😓' },
  { code: '[酷]', emoji: '😎' },
  { code: '[爱心]', emoji: '❤️' },
  { code: '[疑问]', emoji: '🤔' }
]

const avatarColors = ['#3366cc', '#0f766e', '#b45309', '#7c3aed', '#b32424', '#0e7490', '#4d7c0f', '#be185d']

const bbsReplies = ref<BbsReply[]>([])
const replyText = ref('')
const replyQuote = ref<{ nickname: string; content: string } | null>(null)
const nickname = ref(localStorage.getItem('bbs-nickname') || '本地用户')
const editingId = ref<number | null>(null)
const editingText = ref('')

const bbsStorageKey = computed(() => `bbs-replies-${route.params.id}`)

const todayCount = computed(() => {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  const today = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
  return bbsReplies.value.filter(r => r.createdAt.startsWith(today)).length
})

const formatNow = () => {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const BBS_STORAGE_VERSION = '2'
const bbsVersionKey = 'bbs-storage-version'

/* 版本迁移：清掉旧版（含预置种子数据）的本地回复 */
const ensureStorageVersion = () => {
  if (localStorage.getItem(bbsVersionKey) !== BBS_STORAGE_VERSION) {
    Object.keys(localStorage)
      .filter(k => k.startsWith('bbs-replies-') || k.startsWith('bbs-op-likes-'))
      .forEach(k => localStorage.removeItem(k))
    localStorage.setItem(bbsVersionKey, BBS_STORAGE_VERSION)
  }
}

const loadReplies = () => {
  const id = Number(route.params.id)
  if (!Number.isFinite(id) || id <= 0) return
  ensureStorageVersion()
  const raw = localStorage.getItem(bbsStorageKey.value)
  if (raw) {
    try {
      bbsReplies.value = JSON.parse(raw)
    } catch {
      bbsReplies.value = []
    }
  } else {
    bbsReplies.value = []
  }
}

const saveReplies = () => {
  localStorage.setItem(bbsStorageKey.value, JSON.stringify(bbsReplies.value))
}

const insertEmoticon = (code: string) => {
  replyText.value += code
}

const focusReply = () => {
  document.getElementById('bbs-reply-box')?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  nextTick(() => {
    const ta = document.querySelector('#bbs-reply-box textarea') as HTMLTextAreaElement | null
    ta?.focus()
  })
}

const quoteReply = (reply: BbsReply) => {
  replyQuote.value = { nickname: reply.nickname, content: reply.content }
  focusReply()
}

const clearQuote = () => {
  replyQuote.value = null
}

const submitReply = () => {
  const content = replyText.value.trim()
  if (!content) return
  const name = nickname.value.trim() || '本地用户'
  localStorage.setItem('bbs-nickname', name)
  bbsReplies.value.push({
    id: Date.now(),
    nickname: name,
    content,
    quote: replyQuote.value,
    likes: 0,
    liked: false,
    mine: true,
    createdAt: formatNow()
  })
  replyText.value = ''
  replyQuote.value = null
  saveReplies()
}

const toggleLike = (reply: BbsReply) => {
  reply.liked = !reply.liked
  reply.likes += reply.liked ? 1 : -1
  saveReplies()
}

const deleteReply = (reply: BbsReply) => {
  bbsReplies.value = bbsReplies.value.filter(r => r.id !== reply.id)
  if (editingId.value === reply.id) cancelEdit()
  saveReplies()
}

const startEdit = (reply: BbsReply) => {
  editingId.value = reply.id
  editingText.value = reply.content
}

const saveEdit = (reply: BbsReply) => {
  const content = editingText.value.trim()
  if (!content) return
  reply.content = content
  editingId.value = null
  editingText.value = ''
  saveReplies()
}

const cancelEdit = () => {
  editingId.value = null
  editingText.value = ''
}

const renderContent = (content: string) => {
  let html = content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
  emoticons.forEach(e => {
    html = html.split(e.code).join(`<span class="emoji">${e.emoji}</span>`)
  })
  return html
}

const avatarStyle = (name: string) => {
  let sum = 0
  for (let i = 0; i < name.length; i += 1) sum += name.charCodeAt(i)
  return { background: avatarColors[sum % avatarColors.length] }
}

/* ---------- 数据加载 ---------- */
const loadDetail = async () => {
  const id = Number(route.params.id)
  if (!Number.isFinite(id) || id <= 0) {
    experienceShare.value = null
    return
  }

  loading.value = true
  try {
    const res = await experienceShareAPI.getOne(id)
    experienceShare.value = res.success && res.data ? res.data : null

    try {
      const listRes = await experienceShareAPI.getAll()
      if (listRes.success && listRes.data) {
        recommendedShares.value = listRes.data.filter((item: ExperienceShare) => item.id !== id).slice(0, 6)
      }
    } catch {
      recommendedShares.value = []
    }
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.push('/experience-shares')
}

const goToRecommended = (id: number) => {
  router.push(`/experience-shares/${id}`)
}

const openPdfPreview = (url: string, fileName: string) => {
  pdfPreviewUrl.value = url
  pdfPreviewFileName.value = fileName
  pdfPreviewDownloadName.value = fileName
  pdfPreviewVisible.value = true
}

const openPdfInNewTab = () => {
  if (!pdfPreviewUrl.value) return
  window.open(pdfPreviewUrl.value, '_blank', 'noopener')
}

const downloadPdf = (url: string, fileName: string) => {
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  link.target = '_blank'
  link.rel = 'noopener'
  link.click()
}

const downloadCurrentPdf = () => {
  if (!pdfPreviewUrl.value) return
  downloadPdf(pdfPreviewUrl.value, pdfPreviewDownloadName.value || 'experience-share.pdf')
}

/* ---------- 回到顶部（跟随布局滚动容器） ---------- */
const getScrollContainer = (): HTMLElement | Window => {
  return (pageRef.value?.closest('.content-frame') as HTMLElement | null) || window
}

const handlePageScroll = () => {
  const el = scrollContainer
  const top = el === window ? window.scrollY : (el as HTMLElement).scrollTop
  showBackToTop.value = top > 320
}

const scrollToTop = () => {
  const el = scrollContainer
  if (el === window) {
    window.scrollTo({ top: 0, behavior: 'smooth' })
  } else {
    (el as HTMLElement).scrollTo({ top: 0, behavior: 'smooth' })
  }
}

watch(() => route.params.id, () => {
  loadDetail()
  replyText.value = ''
  replyQuote.value = null
  editingId.value = null
  editingText.value = ''
  loadReplies()
})

onMounted(() => {
  loadDetail()
  loadReplies()
  scrollContainer = getScrollContainer()
  scrollContainer.addEventListener('scroll', handlePageScroll)
})

onBeforeUnmount(() => {
  scrollContainer?.removeEventListener('scroll', handlePageScroll)
  tocObserver?.disconnect()
  tocObserver = null
  if (tocRaf) cancelAnimationFrame(tocRaf)
})
</script>

<style scoped>
.wiki-page {
  min-height: 100vh;
  background: #f8f9fa;
  color: #202122;
  padding-bottom: 40px;
}

/* ---------- 顶栏 ---------- */
.wiki-topbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 26px;
  background: #fff;
  border-bottom: 1px solid #d8dde2;
  position: sticky;
  top: 0;
  z-index: 20;
}

.wiki-logo {
  font-family: 'Songti SC', 'SimSun', serif;
  font-size: 20px;
  font-weight: 900;
  color: #202122;
  letter-spacing: 1px;
  flex-shrink: 0;
}

.wiki-logo span {
  color: #3366cc;
}

.wiki-tagline {
  font-size: 11px;
  color: #6c757d;
  flex-shrink: 0;
}

.wiki-search {
  flex: 1;
  max-width: 400px;
  display: flex;
  min-width: 0;
}

.wiki-search input {
  flex: 1;
  min-width: 0;
  padding: 6px 12px;
  border: 1px solid #c8ccd1;
  border-radius: 3px 0 0 3px;
  font-size: 13px;
  outline: none;
  background: #fff;
  color: #202122;
  font-family: inherit;
}

.wiki-search button {
  padding: 6px 16px;
  border: 1px solid #c8ccd1;
  border-left: none;
  border-radius: 0 3px 3px 0;
  background: #f8f9fa;
  cursor: pointer;
  font-size: 13px;
  color: #202122;
  font-family: inherit;
}

.wiki-actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
  align-items: center;
  flex-shrink: 0;
}

.save-status {
  font-size: 12px;
  color: #6c757d;
  white-space: nowrap;
}

.wbtn {
  cursor: pointer;
  font-size: 12px;
  padding: 5px 12px;
  border-radius: 3px;
  background: #fff;
  color: #3366cc;
  border: 1px solid #3366cc;
  font-family: inherit;
  transition: background-color 0.15s ease;
}

.wbtn:hover {
  background: #eaf3ff;
}

.wbtn.on {
  background: #3366cc;
  color: #fff;
}

.wbtn.primary {
  background: #3366cc;
  color: #fff;
}

.wbtn.primary:hover {
  background: #2a56b0;
}

/* ---------- 布局 ---------- */
.wiki-wrap {
  max-width: 1060px;
  margin: 0 auto;
  padding: 18px 24px 0;
  display: grid;
  grid-template-columns: 190px 1fr;
  gap: 26px;
  align-items: start;
}

.wiki-wrap.no-toc {
  grid-template-columns: 1fr;
}

.wiki-left {
  position: sticky;
  top: 74px;
}

.toc-box {
  background: #f8f9fa;
  border: 1px solid #c8ccd1;
  font-size: 13px;
}

.toc-h {
  padding: 8px 12px;
  background: #eaf3ff;
  border-bottom: 1px solid #c8ccd1;
  font-weight: 700;
  color: #202122;
}

.toc-link {
  display: block;
  padding: 5px 12px;
  color: #202122;
  cursor: pointer;
  text-decoration: none;
}

.toc-link:hover {
  background: #eaf3ff;
  text-decoration: none;
}

.toc-link .n {
  color: #3366cc;
  margin-right: 4px;
}

.toc-link.lv2 {
  padding-left: 26px;
  font-size: 12px;
  color: #54595d;
}

.help-box {
  margin-top: 14px;
  border: 1px solid #c8ccd1;
  background: #fff;
  font-size: 12px;
  padding: 10px 12px;
  color: #54595d;
  line-height: 1.7;
}

/* ---------- 词条头部 ---------- */
.wiki-main {
  min-width: 0;
}

.page-head {
  border-bottom: 1px solid #a2a9b1;
  padding-bottom: 6px;
}

.page-head h1 {
  font-size: 28px;
  font-weight: 800;
  font-family: 'Songti SC', 'SimSun', serif;
  letter-spacing: 1px;
  color: #202122;
  line-height: 1.35;
  word-break: break-word;
}

.tabs {
  display: flex;
  gap: 4px;
  margin-top: 10px;
  font-size: 13px;
}

.tabs span {
  padding: 4px 12px;
  cursor: pointer;
  color: #3366cc;
  user-select: none;
  transition: background-color 0.15s ease;
}

.tabs span:hover {
  background: #eaf3ff;
}

.tabs .active {
  background: #fff;
  border: 1px solid #a2a9b1;
  border-bottom: none;
  border-radius: 3px 3px 0 0;
  color: #202122;
  font-weight: 600;
}

.summary {
  margin-top: 12px;
  font-size: 14px;
  line-height: 1.8;
  color: #3a3f44;
}

/* ---------- 信息框 ---------- */
.article {
  font-size: 14px;
  line-height: 1.9;
  color: #202122;
}

.infobox {
  float: right;
  margin: 8px 0 18px 24px;
  width: 264px;
  background: #f8f9fa;
  border: 1px solid #a2a9b1;
  font-size: 12px;
  font-family: 'Songti SC', 'SimSun', serif;
}

.ib-h {
  background: #3366cc;
  color: #fff;
  text-align: center;
  padding: 6px;
  font-weight: 700;
  letter-spacing: 1px;
}

.ib-img {
  height: 120px;
  background: linear-gradient(160deg, #dbe6f0, #c7d8e6);
  position: relative;
  overflow: hidden;
}

.ib-img img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.ib-img-ph {
  position: absolute;
  inset: 0;
}

.ib-img-ph .sun {
  position: absolute;
  right: 20%;
  top: 20%;
  width: 46px;
  height: 46px;
  border-radius: 50%;
  background: radial-gradient(circle, #f2c078, #d9974a);
}

.ib-img-ph .m {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 40px;
  background: linear-gradient(180deg, #7c9cb0, #5a7c92);
  clip-path: polygon(0 100%, 15% 30%, 30% 90%, 46% 12%, 62% 84%, 78% 24%, 92% 86%, 100% 16%, 100% 100%);
}

.ib-row {
  display: flex;
  padding: 5px 8px;
  border-top: 1px solid #d8dde2;
}

.ib-row .k {
  width: 78px;
  color: #54595d;
  flex-shrink: 0;
}

.ib-row .v {
  flex: 1;
  font-weight: 600;
  word-break: break-all;
}

.ib-row .v .tag {
  font-size: 11px;
  color: #3366cc;
  border: 1px solid #3366cc;
  border-radius: 3px;
  padding: 0 6px;
  font-weight: 600;
}

.ib-row .v .tag.pdf {
  color: #b45309;
  border-color: #b45309;
}

/* ---------- 题图 ---------- */
.cover-fig {
  margin: 18px 0;
}

.cover-fig img {
  display: block;
  width: 100%;
  max-height: 320px;
  object-fit: cover;
  border: 1px solid #d8dde2;
  border-radius: 4px;
}

.cover-fig figcaption {
  margin-top: 4px;
  font-size: 11px;
  color: #6c757d;
}

/* ---------- 正文 ---------- */
.article-body {
  min-width: 0;
}

.article-body :deep(.tiptap) {
  font-size: 14px;
  line-height: 1.9;
  color: #202122;
}

.article-body :deep(.tiptap p) {
  margin: 0 0 16px;
}

.article-body :deep(.tiptap ul),
.article-body :deep(.tiptap ol) {
  padding-left: 2em;
  margin: 0 0 16px;
}

.article-body :deep(.tiptap li) {
  margin: 0.25em 0;
}

.article-body :deep(.tiptap img) {
  display: block;
  max-width: 100%;
  height: auto;
  margin: 18px 0;
  border-radius: 14px;
}

.article-body :deep(.tiptap table) {
  display: block;
  max-width: 100%;
  overflow-x: auto;
  border-spacing: 0;
  border-collapse: collapse;
  margin-bottom: 16px;
}

.article-body :deep(.tiptap table td) {
  border: 1px solid #dfe2e5;
}

.article-body :deep(.tiptap h2) {
  font-family: 'Songti SC', 'SimSun', serif;
  font-size: 20px;
  border-bottom: 1px solid #a2a9b1;
  margin: 24px 0 10px;
  padding-bottom: 3px;
  scroll-margin-top: 90px;
}

.article-body :deep(.tiptap h3) {
  font-size: 15px;
  margin: 16px 0 8px;
  scroll-margin-top: 90px;
}

.article-body :deep(.tiptap blockquote) {
  margin: 12px 0;
  padding: 10px 14px;
  background: #f0f4ff;
  border-left: 4px solid #3366cc;
  color: #3a3f44;
  font-size: 13px;
}

.article-body :deep(.tiptap pre) {
  background: #f6f8fa;
  border: 1px solid #d8dde2;
  border-radius: 4px;
}

.article-body :deep(.tiptap code) {
  background-color: rgba(27, 31, 35, 0.05);
}

.article-body :deep(.tiptap table th) {
  background: #eaf3ff;
}

.article-body :deep(.tiptap a) {
  color: #3366cc;
}

/* PDF 正文 */
.pdf-body {
  margin: 18px 0;
}

/* 空内容 */
.empty-content {
  margin: 18px 0;
  padding: 22px;
  border: 1px dashed #c8ccd1;
  border-radius: 4px;
  color: #6c757d;
  text-align: center;
  font-size: 13px;
  background: #fff;
}

/* 补充说明 */
.notice {
  margin: 18px 0;
  padding: 12px 14px;
  background: #f8f9fa;
  border: 1px solid #a2a9b1;
  font-size: 13px;
  color: #54595d;
  scroll-margin-top: 90px;
}

.notice b {
  color: #202122;
}

.notice :deep(.tiptap) {
  margin-top: 8px;
  font-size: 13px;
  line-height: 1.8;
  color: #54595d;
}

.notice :deep(.tiptap p) {
  margin: 0 0 12px;
}

.notice :deep(.tiptap h1),
.notice :deep(.tiptap h2),
.notice :deep(.tiptap h3),
.notice :deep(.tiptap h4) {
  font-size: 15px;
  margin: 10px 0 6px;
  color: #202122;
}

/* ---------- 附件框 ---------- */
.att-box {
  margin: 18px 0;
  border: 1px solid #a2a9b1;
  background: #fff;
  scroll-margin-top: 90px;
}

.ab-h {
  background: #eaf3ff;
  padding: 6px 12px;
  font-size: 13px;
  font-weight: 700;
  border-bottom: 1px solid #a2a9b1;
  color: #202122;
}

.att {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
}

.att .ic {
  width: 32px;
  height: 38px;
  background: #b32424;
  color: #fff;
  font-size: 9px;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border-radius: 3px;
}

.att-info {
  flex: 1;
  min-width: 0;
}

.att .nm {
  font-size: 13px;
  font-weight: 700;
  color: #3366cc;
  word-break: break-all;
}

.att .sz {
  font-size: 11px;
  color: #6c757d;
  margin-top: 2px;
}

.att .btns {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.att .btns a {
  font-size: 12px;
  padding: 5px 11px;
  border: 1px solid #3366cc;
  border-radius: 3px;
  cursor: pointer;
  color: #3366cc;
  text-decoration: none;
}

.att .btns a:first-child {
  background: #3366cc;
  color: #fff;
}

/* ---------- 相关条目 ---------- */
.rel-section {
  margin-top: 24px;
  scroll-margin-top: 90px;
}

.rel-section h2 {
  font-family: 'Songti SC', 'SimSun', serif;
  font-size: 20px;
  border-bottom: 1px solid #a2a9b1;
  margin-bottom: 10px;
  padding-bottom: 3px;
  color: #202122;
}

.related {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-top: 12px;
}

.rel-item {
  border: 1px solid #c8ccd1;
  padding: 9px 12px;
  font-size: 13px;
  color: #3366cc;
  cursor: pointer;
  background: #fff;
  text-decoration: none;
  transition: background-color 0.15s ease;
}

.rel-item:hover {
  background: #eaf3ff;
}

.rel-item .rc {
  height: 56px;
  border-radius: 3px;
  margin-bottom: 7px;
  overflow: hidden;
  background: linear-gradient(135deg, #dbe6f0, #c7d8e6);
}

.rel-item .rc img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.rel-item .rc .g.ph {
  width: 100%;
  height: 100%;
}

.rel-item .rt {
  line-height: 1.5;
  word-break: break-word;
}

.catline {
  margin-top: 22px;
  padding-top: 10px;
  border-top: 1px solid #a2a9b1;
  font-size: 12px;
  color: #6c757d;
}

/* ---------- 回到顶部 ---------- */
.backtop-float {
  position: fixed;
  right: 22px;
  bottom: 24px;
  z-index: 30;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 40px;
  padding: 0 14px 0 10px;
  border: 1px solid #3366cc;
  border-radius: 4px;
  background: #fff;
  color: #3366cc;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.12);
  cursor: pointer;
  transition: transform 0.2s ease, background-color 0.2s ease;
}

.backtop-float:hover {
  background: #eaf3ff;
  transform: translateY(-2px);
}

.backtop-icon {
  font-size: 14px;
  font-weight: 700;
  line-height: 1;
}

.backtop-text {
  font-size: 12px;
}

/* ---------- PDF 预览弹窗 ---------- */
.pdf-preview-shell {
  display: flex;
  flex-direction: column;
  gap: 14px;
  max-height: 82vh;
}

.pdf-preview-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 14px;
  background: #f8fafc;
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.pdf-preview-file {
  color: #0f172a;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.pdf-preview-file span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pdf-preview-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.pdf-preview-body {
  overflow-y: auto;
  max-height: calc(82vh - 74px);
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
  padding-right: 2px;
}

/* ---------- 讨论区（BBS 格式） ---------- */
.bbs-section {
  margin-top: 30px;
  border: 1px solid #a2a9b1;
  background: #fff;
  clear: both;
  scroll-margin-top: 90px;
}

.bbs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 8px 14px;
  background: #3366cc;
  color: #fff;
  font-size: 13px;
  font-weight: 700;
}

.bbs-stats {
  font-size: 12px;
  font-weight: 400;
  opacity: 0.9;
}

.bbs-post {
  display: flex;
  border-bottom: 1px solid #e5e7eb;
}

.bbs-post:last-of-type {
  border-bottom: none;
}

.bbs-post:nth-child(even) {
  background: #f8f9fa;
}

.bbs-empty {
  padding: 26px 16px;
  text-align: center;
  font-size: 13px;
  color: #6c757d;
  border-bottom: 1px solid #e5e7eb;
  background: #f8f9fa;
}

.post-side {
  width: 120px;
  flex-shrink: 0;
  padding: 14px 10px;
  text-align: center;
  border-right: 1px solid #e5e7eb;
}

.avatar {
  width: 48px;
  height: 48px;
  margin: 0 auto;
  border-radius: 50%;
  background: #3366cc;
  color: #fff;
  font-size: 18px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.username {
  margin-top: 6px;
  font-size: 13px;
  font-weight: 700;
  color: #202122;
  word-break: break-all;
}

.badge {
  display: inline-block;
  margin-top: 4px;
  font-size: 10px;
  color: #b45309;
  border: 1px solid #b45309;
  border-radius: 3px;
  padding: 0 5px;
}

.badge.mine {
  color: #3366cc;
  border-color: #3366cc;
}

.post-main {
  flex: 1;
  min-width: 0;
  padding: 12px 16px;
}

.post-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  color: #6c757d;
}

.floor {
  font-weight: 700;
  color: #3366cc;
}

.post-content {
  margin-top: 8px;
  font-size: 14px;
  line-height: 1.8;
  color: #202122;
  word-break: break-word;
}

.post-content .emoji {
  font-size: 16px;
}

.quote-block {
  margin: 8px 0;
  padding: 8px 12px;
  background: #f0f4ff;
  border-left: 3px solid #3366cc;
  font-size: 12px;
  color: #54595d;
  line-height: 1.7;
  word-break: break-word;
}

.post-actions {
  margin-top: 10px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.post-actions button {
  font-size: 12px;
  padding: 4px 10px;
  border: 1px solid #c8ccd1;
  background: #fff;
  border-radius: 3px;
  cursor: pointer;
  color: #3366cc;
  font-family: inherit;
  transition: background-color 0.15s ease;
}

.post-actions button:hover {
  background: #eaf3ff;
}

.post-actions button.liked {
  background: #3366cc;
  color: #fff;
  border-color: #3366cc;
}

.post-actions button.save {
  background: #3366cc;
  color: #fff;
  border-color: #3366cc;
}

.post-actions button.save:hover {
  background: #2a56b0;
}

.post-actions button.edit {
  color: #0f766e;
  border-color: #0f766e;
}

.post-actions button.edit:hover {
  background: #ecfdf5;
}

.post-actions button.del {
  color: #b32424;
  border-color: #b32424;
}

.post-actions button.del:hover {
  background: #fef2f2;
}

.edit-textarea {
  width: 100%;
  min-height: 70px;
  padding: 8px 10px;
  border: 1px solid #c8ccd1;
  border-radius: 4px;
  font-size: 13px;
  font-family: inherit;
  line-height: 1.7;
  resize: vertical;
  outline: none;
  background: #fff;
  color: #202122;
}

.edit-textarea:focus {
  border-color: #3366cc;
}

/* 回复框 */
.bbs-reply-box {
  padding: 14px 16px;
  background: #f8f9fa;
  border-top: 1px solid #a2a9b1;
}

.reply-toolbar {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
}

.nick-label {
  font-size: 12px;
  color: #54595d;
}

.nick-label input {
  width: 110px;
  padding: 4px 8px;
  border: 1px solid #c8ccd1;
  border-radius: 3px;
  font-size: 12px;
  font-family: inherit;
  outline: none;
  background: #fff;
  color: #202122;
}

.emoticons {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.emoticons button {
  font-size: 14px;
  padding: 3px 7px;
  border: 1px solid #c8ccd1;
  background: #fff;
  border-radius: 3px;
  cursor: pointer;
  line-height: 1;
}

.emoticons button:hover {
  background: #eaf3ff;
}

.reply-quote {
  margin-top: 10px;
  padding: 8px 12px;
  background: #f0f4ff;
  border-left: 3px solid #3366cc;
  font-size: 12px;
  color: #54595d;
  display: flex;
  align-items: center;
  gap: 8px;
}

.reply-quote span {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reply-quote .x {
  border: none;
  background: none;
  cursor: pointer;
  font-size: 14px;
  color: #6c757d;
  flex-shrink: 0;
}

.bbs-reply-box textarea {
  width: 100%;
  min-height: 90px;
  margin-top: 10px;
  padding: 10px;
  border: 1px solid #c8ccd1;
  border-radius: 4px;
  font-size: 13px;
  font-family: inherit;
  line-height: 1.7;
  resize: vertical;
  outline: none;
  background: #fff;
  color: #202122;
}

.bbs-reply-box textarea:focus {
  border-color: #3366cc;
}

.reply-actions {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 10px;
  margin-top: 10px;
}

.reply-actions .hint {
  font-size: 11px;
  color: #6c757d;
}

.reply-actions .submit {
  font-size: 13px;
  padding: 7px 18px;
  border: none;
  border-radius: 3px;
  background: #3366cc;
  color: #fff;
  cursor: pointer;
  font-family: inherit;
  font-weight: 600;
}

.reply-actions .submit:hover {
  background: #2a56b0;
}

/* ---------- Tab 面板（历史 / 日志） ---------- */
.tab-panel {
  margin-top: 18px;
}

.tab-panel h2 {
  font-family: 'Songti SC', 'SimSun', serif;
  font-size: 20px;
  border-bottom: 1px solid #a2a9b1;
  margin-bottom: 14px;
  padding-bottom: 3px;
  color: #202122;
}

.history-list {
  display: flex;
  flex-direction: column;
}

.history-item {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 12px 0;
  border-bottom: 1px solid #e5e7eb;
}

.history-item:last-child {
  border-bottom: none;
}

.h-version {
  flex-shrink: 0;
  min-width: 72px;
  text-align: center;
  font-size: 12px;
  font-weight: 700;
  color: #3366cc;
  border: 1px solid #3366cc;
  border-radius: 3px;
  padding: 3px 8px;
}

.h-info .h-time {
  font-size: 12px;
  color: #6c757d;
}

.h-info .h-desc {
  margin-top: 3px;
  font-size: 14px;
  color: #202122;
}

.log-list {
  display: flex;
  flex-direction: column;
}

.log-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid #e5e7eb;
  font-size: 13px;
}

.log-item:last-child {
  border-bottom: none;
}

.log-time {
  flex-shrink: 0;
  font-family: Consolas, monospace;
  font-size: 12px;
  color: #6c757d;
}

.log-action {
  flex: 1;
  color: #202122;
}

.log-user {
  flex-shrink: 0;
  font-size: 12px;
  color: #3366cc;
}

/* ---------- 响应式 ---------- */
@media (max-width: 900px) {
  .wiki-wrap {
    grid-template-columns: 1fr;
  }

  .wiki-left {
    display: none;
  }

  .infobox {
    float: none;
    margin: 14px 0;
    width: 100%;
  }

  .related {
    grid-template-columns: 1fr;
  }

  .wiki-topbar {
    flex-wrap: wrap;
    padding: 8px 16px;
  }

  .wiki-search {
    order: 3;
    max-width: 100%;
    flex-basis: 100%;
  }

  .page-head h1 {
    font-size: 22px;
  }
}

@media (max-width: 600px) {
  .wiki-actions {
    flex-wrap: wrap;
  }

  .att {
    flex-wrap: wrap;
  }

  .att .btns {
    width: 100%;
    justify-content: flex-end;
  }

  .bbs-post {
    flex-direction: column;
  }

  .post-side {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid #e5e7eb;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 14px;
    text-align: left;
  }

  .avatar {
    margin: 0;
    width: 36px;
    height: 36px;
    font-size: 14px;
  }

  .username {
    margin-top: 0;
  }

  .badge {
    margin-top: 0;
  }
}

/* ---------- 暗色主题适配 ---------- */
:root[data-theme='dark'] .wiki-page {
  background: #16181d;
  color: #e6edf3;
}

:root[data-theme='dark'] .wiki-topbar {
  background: #1f2937;
  border-bottom-color: rgba(255, 255, 255, 0.08);
}

:root[data-theme='dark'] .wiki-logo {
  color: #e6edf3;
}

:root[data-theme='dark'] .wiki-logo span {
  color: #93c5fd;
}

:root[data-theme='dark'] .wiki-tagline {
  color: #9ca3af;
}

:root[data-theme='dark'] .wiki-search input {
  background: #111827;
  border-color: #374151;
  color: #e6edf3;
}

:root[data-theme='dark'] .wiki-search button {
  background: #1f2937;
  border-color: #374151;
  color: #e6edf3;
}

:root[data-theme='dark'] .wbtn {
  background: transparent;
  color: #93c5fd;
  border-color: #93c5fd;
}

:root[data-theme='dark'] .save-status {
  color: #9ca3af;
}

:root[data-theme='dark'] .wbtn:hover {
  background: rgba(59, 130, 246, 0.15);
}

:root[data-theme='dark'] .wbtn.on,
:root[data-theme='dark'] .wbtn.primary {
  background: #3b82f6;
  color: #fff;
  border-color: #3b82f6;
}

:root[data-theme='dark'] .toc-box {
  background: #1f2937;
  border-color: #374151;
}

:root[data-theme='dark'] .toc-h {
  background: rgba(59, 130, 246, 0.15);
  border-bottom-color: #374151;
  color: #e6edf3;
}

:root[data-theme='dark'] .toc-link {
  color: #e6edf3;
}

:root[data-theme='dark'] .toc-link:hover {
  background: rgba(59, 130, 246, 0.12);
}

:root[data-theme='dark'] .toc-link .n {
  color: #93c5fd;
}

:root[data-theme='dark'] .toc-link.lv2 {
  color: #9ca3af;
}

:root[data-theme='dark'] .help-box {
  background: #1f2937;
  border-color: #374151;
  color: #9ca3af;
}

:root[data-theme='dark'] .page-head {
  border-bottom-color: #374151;
}

:root[data-theme='dark'] .page-head h1 {
  color: #e6edf3;
}

:root[data-theme='dark'] .tabs span {
  color: #93c5fd;
}

:root[data-theme='dark'] .tabs .active {
  background: #16181d;
  border-color: #374151;
  color: #e6edf3;
}

:root[data-theme='dark'] .summary {
  color: #9ca3af;
}

:root[data-theme='dark'] .article {
  color: #e6edf3;
}

:root[data-theme='dark'] .infobox {
  background: #1f2937;
  border-color: #374151;
}

:root[data-theme='dark'] .ib-h {
  background: #3b82f6;
}

:root[data-theme='dark'] .ib-img {
  background: linear-gradient(160deg, #1e293b, #273449);
}

:root[data-theme='dark'] .ib-row {
  border-top-color: #374151;
}

:root[data-theme='dark'] .ib-row .k {
  color: #9ca3af;
}

:root[data-theme='dark'] .ib-row .v {
  color: #e6edf3;
}

:root[data-theme='dark'] .ib-row .v .tag {
  color: #93c5fd;
  border-color: #93c5fd;
}

:root[data-theme='dark'] .ib-row .v .tag.pdf {
  color: #fbbf24;
  border-color: #fbbf24;
}

:root[data-theme='dark'] .cover-fig img {
  border-color: #374151;
}

:root[data-theme='dark'] .cover-fig figcaption {
  color: #9ca3af;
}

:root[data-theme='dark'] .article-body :deep(.tiptap) {
  color: #e6edf3;
}

:root[data-theme='dark'] .article-body :deep(.tiptap h2) {
  color: #e6edf3;
  border-bottom-color: #374151;
}

:root[data-theme='dark'] .article-body :deep(.tiptap h3) {
  color: #e6edf3;
}

:root[data-theme='dark'] .article-body :deep(.tiptap blockquote) {
  background: rgba(59, 130, 246, 0.1);
  border-left-color: rgba(59, 130, 246, 0.5);
  color: #9ca3af;
}

:root[data-theme='dark'] .article-body :deep(.tiptap pre) {
  background: rgba(22, 27, 34, 0.95);
  border-color: rgba(42, 50, 66, 0.3);
}

:root[data-theme='dark'] .article-body :deep(.tiptap code) {
  background-color: rgba(28, 35, 51, 0.6);
  color: #e6edf3;
}

:root[data-theme='dark'] .article-body :deep(.tiptap table th) {
  background: rgba(28, 35, 51, 0.6);
  color: #e6edf3;
}

:root[data-theme='dark'] .article-body :deep(.tiptap table td) {
  color: #e6edf3;
}

:root[data-theme='dark'] .article-body :deep(.tiptap a) {
  color: #93c5fd;
}

:root[data-theme='dark'] .empty-content {
  background: #1f2937;
  border-color: #374151;
  color: #9ca3af;
}

:root[data-theme='dark'] .notice {
  background: #1f2937;
  border-color: #374151;
  color: #9ca3af;
}

:root[data-theme='dark'] .notice b {
  color: #e6edf3;
}

:root[data-theme='dark'] .notice :deep(.tiptap) {
  color: #9ca3af;
}

:root[data-theme='dark'] .notice :deep(.tiptap h1),
:root[data-theme='dark'] .notice :deep(.tiptap h2),
:root[data-theme='dark'] .notice :deep(.tiptap h3),
:root[data-theme='dark'] .notice :deep(.tiptap h4) {
  color: #e6edf3;
}

:root[data-theme='dark'] .att-box {
  background: #1f2937;
  border-color: #374151;
}

:root[data-theme='dark'] .ab-h {
  background: rgba(59, 130, 246, 0.15);
  border-bottom-color: #374151;
  color: #e6edf3;
}

:root[data-theme='dark'] .att .nm {
  color: #93c5fd;
}

:root[data-theme='dark'] .att .sz {
  color: #9ca3af;
}

:root[data-theme='dark'] .att .btns a {
  color: #93c5fd;
  border-color: #93c5fd;
}

:root[data-theme='dark'] .att .btns a:first-child {
  background: #3b82f6;
  color: #fff;
  border-color: #3b82f6;
}

:root[data-theme='dark'] .rel-section h2 {
  color: #e6edf3;
  border-bottom-color: #374151;
}

:root[data-theme='dark'] .rel-item {
  background: #1f2937;
  border-color: #374151;
  color: #93c5fd;
}

:root[data-theme='dark'] .rel-item:hover {
  background: rgba(59, 130, 246, 0.12);
}

:root[data-theme='dark'] .rel-item .rc {
  background: linear-gradient(135deg, #1e293b, #273449);
}

:root[data-theme='dark'] .catline {
  border-top-color: #374151;
  color: #9ca3af;
}

:root[data-theme='dark'] .backtop-float {
  background: #1f2937;
  border-color: #93c5fd;
  color: #93c5fd;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .backtop-float:hover {
  background: rgba(59, 130, 246, 0.15);
}

:root[data-theme='dark'] .pdf-preview-toolbar {
  background: #2a2a2a;
  border-color: rgba(255, 255, 255, 0.1);
}

:root[data-theme='dark'] .pdf-preview-file {
  color: #e6edf3;
}

:root[data-theme='dark'] .bbs-section {
  background: #1f2937;
  border-color: #374151;
}

:root[data-theme='dark'] .bbs-header {
  background: #3b82f6;
}

:root[data-theme='dark'] .bbs-post {
  border-bottom-color: #374151;
}

:root[data-theme='dark'] .bbs-post:nth-child(even) {
  background: #1a2230;
}

:root[data-theme='dark'] .bbs-empty {
  color: #9ca3af;
  border-bottom-color: #374151;
  background: #1a2230;
}

:root[data-theme='dark'] .post-side {
  border-right-color: #374151;
  border-bottom-color: #374151;
}

:root[data-theme='dark'] .username {
  color: #e6edf3;
}

:root[data-theme='dark'] .post-meta {
  color: #9ca3af;
}

:root[data-theme='dark'] .floor {
  color: #93c5fd;
}

:root[data-theme='dark'] .post-content {
  color: #e6edf3;
}

:root[data-theme='dark'] .quote-block {
  background: rgba(59, 130, 246, 0.1);
  border-left-color: rgba(59, 130, 246, 0.5);
  color: #9ca3af;
}

:root[data-theme='dark'] .post-actions button {
  background: #1f2937;
  border-color: #374151;
  color: #93c5fd;
}

:root[data-theme='dark'] .post-actions button:hover {
  background: rgba(59, 130, 246, 0.15);
}

:root[data-theme='dark'] .post-actions button.liked {
  background: #3b82f6;
  color: #fff;
  border-color: #3b82f6;
}

:root[data-theme='dark'] .post-actions button.del {
  color: #f87171;
  border-color: #f87171;
}

:root[data-theme='dark'] .post-actions button.save {
  background: #3b82f6;
  color: #fff;
  border-color: #3b82f6;
}

:root[data-theme='dark'] .post-actions button.edit {
  color: #5eead4;
  border-color: #5eead4;
}

:root[data-theme='dark'] .post-actions button.edit:hover {
  background: rgba(94, 234, 212, 0.12);
}

:root[data-theme='dark'] .post-actions button.del:hover {
  background: rgba(248, 113, 113, 0.12);
}

:root[data-theme='dark'] .edit-textarea {
  background: #111827;
  border-color: #374151;
  color: #e6edf3;
}

:root[data-theme='dark'] .edit-textarea:focus {
  border-color: #3b82f6;
}

:root[data-theme='dark'] .bbs-reply-box {
  background: #1a2230;
  border-top-color: #374151;
}

:root[data-theme='dark'] .nick-label {
  color: #9ca3af;
}

:root[data-theme='dark'] .nick-label input {
  background: #111827;
  border-color: #374151;
  color: #e6edf3;
}

:root[data-theme='dark'] .emoticons button {
  background: #1f2937;
  border-color: #374151;
}

:root[data-theme='dark'] .emoticons button:hover {
  background: rgba(59, 130, 246, 0.15);
}

:root[data-theme='dark'] .reply-quote {
  background: rgba(59, 130, 246, 0.1);
  border-left-color: rgba(59, 130, 246, 0.5);
  color: #9ca3af;
}

:root[data-theme='dark'] .reply-quote .x {
  color: #9ca3af;
}

:root[data-theme='dark'] .bbs-reply-box textarea {
  background: #111827;
  border-color: #374151;
  color: #e6edf3;
}

:root[data-theme='dark'] .bbs-reply-box textarea:focus {
  border-color: #3b82f6;
}

:root[data-theme='dark'] .reply-actions .hint {
  color: #9ca3af;
}

:root[data-theme='dark'] .tabs span:hover {
  background: rgba(59, 130, 246, 0.12);
}

:root[data-theme='dark'] .tab-panel h2 {
  color: #e6edf3;
  border-bottom-color: #374151;
}

:root[data-theme='dark'] .history-item {
  border-bottom-color: #374151;
}

:root[data-theme='dark'] .h-version {
  color: #93c5fd;
  border-color: #93c5fd;
}

:root[data-theme='dark'] .h-info .h-time {
  color: #9ca3af;
}

:root[data-theme='dark'] .h-info .h-desc {
  color: #e6edf3;
}

:root[data-theme='dark'] .log-item {
  border-bottom-color: #374151;
}

:root[data-theme='dark'] .log-time {
  color: #9ca3af;
}

:root[data-theme='dark'] .log-action {
  color: #e6edf3;
}

:root[data-theme='dark'] .log-user {
  color: #93c5fd;
}
</style>
