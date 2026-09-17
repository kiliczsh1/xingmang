<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="ai-chat-overlay"
      :class="theme === 'platinum' ? 'platinum' : ''"
      @click.self="close"
    >
      <!-- 左上角关闭（浮于遮罩上，独立于手机壳） -->
      <button class="ph-close" title="关闭 (Esc)" @click="close">✕</button>

      <div class="phone">
        <div class="status-bar"><span>9:41</span><span class="sig">📶</span></div>
        <div class="notch"></div>

        <div class="chat-head">
          <span class="back" @click="close">‹</span>
          <div class="ava">✦</div>
          <div class="head-info">
            <div class="nm">星芒</div>
            <div class="st">{{ currentConv ? currentConv.title : '在线 · 随时找你' }}</div>
          </div>
          <div class="head-acts">
            <div class="theme-switch" @click="toggleTheme" :title="'当前：' + (theme === 'black-gold' ? '黑金典藏' : '白金雅致')">
              <span class="knob">◐</span>
            </div>
            <span class="more" @click="openSheet('menu')">•••</span>
          </div>
        </div>

        <div class="msgs" ref="msgsRef">
          <!-- 空态：建议引导卡 -->
          <div v-if="showEmpty" class="empty">
            <div class="e-emoji">✍️</div>
            <h3>今天想写点什么？</h3>
            <p>点一张卡片开始，或者直接输入</p>
            <div class="sug-grid">
              <div
                v-for="(s, i) in SUGS"
                :key="i"
                class="sug-card"
                :data-c="i"
                @click="sendSuggestion(s.q)"
              >
                <div class="e">{{ s.e }}</div>
                <div class="t">{{ s.t }}</div>
                <div class="d">{{ s.d }}</div>
              </div>
            </div>
          </div>

          <!-- 消息列表 -->
          <template v-else>
            <div v-for="(m, i) in messages" :key="m.id ?? i" class="row" :class="m.role === 'user' ? 'me' : 'them'">
              <div v-if="m.role !== 'user'" class="av">✦</div>
              <div class="bub">
                <MarkdownRenderer v-if="m.role === 'assistant' && m.content" :content="m.content" class="md-body" />
                <span v-else class="plain">{{ m.content }}<i v-if="m.streaming && streamMode" class="crt"></i></span>
              </div>
            </div>
            <!-- 追问芯片 -->
            <div v-if="followChips.length" class="follow-chips">
              <button
                v-for="(c, i) in followChips"
                :key="i"
                class="fchip"
                :data-i="i"
                @click="sendChip(c)"
              >{{ c }}</button>
            </div>
          </template>
        </div>

        <div class="dock">
          <div class="inputrow">
            <textarea
              v-model="input"
              rows="1"
              placeholder="说点什么…"
              @keydown.enter.prevent="handleEnter"
              @input="autoGrow"
              :disabled="sending"
            ></textarea>
            <button class="send" @click="sendFromBtn" :disabled="sending || !input.trim()">➤</button>
          </div>
        </div>

        <!-- 遮罩 + 底部弹层 -->
        <div class="mask" :class="{ show: sheet !== null }" @click="closeSheets"></div>

        <div class="sheet" :class="{ show: sheet === 'menu' }">
          <div class="grab"></div>
          <h5>更多操作 <span class="close" @click="closeSheets">✕</span></h5>
          <button class="sitem" @click="createConvFromMenu"><span class="ic">＋</span>新建对话</button>
          <button class="sitem" @click="toggleTheme"><span class="ic">🎨</span>外观风格 <span class="sub">{{ theme === 'black-gold' ? '黑金' : '白金' }}</span></button>
          <button class="sitem" @click="openSheet('hist')"><span class="ic">🕘</span>历史切换</button>
          <button class="sitem" @click="openSheet('model')"><span class="ic">🧠</span>模型选择 <span class="sub">{{ currentModel ? currentModel.name : '未配置' }}</span></button>
          <button class="sitem" @click="toggleStreamMode"><span class="ic">⚡</span>输出模式 <span class="sub">{{ streamMode === 'stream' ? '流式' : '非流式' }}</span></button>
          <button class="sitem" @click="exportConv"><span class="ic">📤</span>导出对话 <span class="sub">.md</span></button>
          <button class="sitem danger" @click="openSheet('clear')"><span class="ic">🗑️</span>清空对话</button>
        </div>

        <div class="sheet" :class="{ show: sheet === 'hist' }">
          <div class="grab"></div>
          <h5>历史切换 <span class="close" @click="closeSheets">✕</span></h5>
          <div class="book-row">
            <span class="bl">作品</span>
            <select v-model="selectedBookId" class="book-select" @change="onBookChange">
              <option v-for="b in books" :key="b.id" :value="b.id">{{ b.title }}</option>
            </select>
            <button class="new-btn" @click="createConv">＋新建</button>
          </div>
          <div class="conv-list">
            <div v-if="!conversations.length" class="empty-tip">暂无对话，点「＋新建」开始</div>
            <div
              v-for="c in conversations"
              :key="c.id"
              class="conv-item"
              :class="{ cur: currentConv && c.id === currentConv.id }"
              @click="selectConv(c)"
            >
              <span class="dot"></span>
              <span class="ct">{{ c.title }}</span>
              <span class="cnt">{{ c.message_count || 0 }} 条</span>
            </div>
          </div>
        </div>

        <div class="sheet" :class="{ show: sheet === 'model' }">
          <div class="grab"></div>
          <h5>模型选择 <span class="close" @click="closeSheets">✕</span></h5>
          <div class="conv-list">
            <div v-if="!models.length" class="empty-tip">未配置可用模型，请先到「API 配置」添加</div>
            <div
              v-for="m in models"
              :key="m.id"
              class="conv-item"
              :class="{ cur: currentModel && m.id === currentModel.id }"
              @click="selectModel(m)"
            >
              <span class="dot" style="background:var(--gold)"></span>
              <span class="ct">{{ m.name }}<br><small class="md">{{ m.description || m.model }}</small></span>
              <span class="cnt">{{ currentModel && m.id === currentModel.id ? '✓ 当前' : '切换' }}</span>
            </div>
          </div>
        </div>

        <div class="sheet" :class="{ show: sheet === 'clear' }">
          <div class="grab"></div>
          <h5>清空对话 <span class="close" @click="closeSheets">✕</span></h5>
          <p class="cfm-txt">将删除当前对话的全部消息，此操作不可恢复。</p>
          <div class="cfm-row">
            <button class="cfm-btn no" @click="closeSheets">取消</button>
            <button class="cfm-btn yes" @click="clearConv">清空</button>
          </div>
        </div>

        <div class="home-bar"></div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import MarkdownRenderer from '../components/MarkdownRenderer.vue'
import { conversationAPI, bookAPI, configAPI } from '@/api'
import type { Conversation, ApiModel } from '@/types'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'update:visible', v: boolean): void }>()
const close = () => emit('update:visible', false)

// ===== 主题：黑金 / 白金 =====
const theme = ref<'black-gold' | 'platinum'>(
  (localStorage.getItem('ai-chat-theme') as any) || 'black-gold'
)
const toggleTheme = () => {
  theme.value = theme.value === 'black-gold' ? 'platinum' : 'black-gold'
  localStorage.setItem('ai-chat-theme', theme.value)
}

// ===== 建议卡 / 追问芯片 =====
const SUGS = [
  { e: '🦸', t: '人物灵感', d: '给主角设计矛盾性格', q: '帮我设计一个表面冷酷但内心柔软的主角，给出性格矛盾点' },
  { e: '📖', t: '情节展开', d: '开头扩成三章大纲', q: '把「雨夜老宅重逢」扩写成三章大纲' },
  { e: '🏙️', t: '世界观', d: '构建蒸汽朋克城市', q: '帮我构建一座蒸汽朋克风格的城市设定' },
  { e: '💬', t: '对话润色', d: '让人物说话更真实', q: '教我怎么让角色对话更符合人物身份和情绪' },
  { e: '🔍', t: '书名脑暴', d: '悬疑小说起名', q: '为一本「失踪案」主题的悬疑小说想出 10 个书名' },
  { e: '❄️', t: '文风模仿', d: '简洁冷峻的笔调', q: '用简洁冷峻的笔调写一段「雪夜追凶」的开头' }
]
const FOLLOWUPS = [
  ['🔎 追问：具体说说', '🔄 换一个角度', '✍️ 扩写这段', '🎨 换个文风'],
  ['💡 再举个例子', '📋 总结要点', '➡️ 继续往下写', '💬 解释一下'],
  ['🗂️ 生成大纲', '🛠️ 给修改建议', '⚡ 更短的版本', '✨ 标记金句']
]
const followIdx = ref(0)
const followChips = ref<string[]>([])
const streamMode = ref<'stream' | 'batch'>(
  (localStorage.getItem('ai-chat-stream-mode') as any) === 'batch' ? 'batch' : 'stream'
)
const toggleStreamMode = () => {
  streamMode.value = streamMode.value === 'stream' ? 'batch' : 'stream'
  localStorage.setItem('ai-chat-stream-mode', streamMode.value)
  ElMessage.success('已切换为' + (streamMode.value === 'stream' ? '流式输出' : '非流式输出'))
}

// ===== 会话状态 =====
const books = ref<any[]>([])
const selectedBookId = ref<number | null>(null)
const conversations = ref<any[]>([])
const currentConv = ref<Conversation | null>(null)
const messages = ref<any[]>([])
const input = ref('')
const sending = ref(false)
const abortCtrl = ref<AbortController | null>(null)
const sheet = ref<'menu' | 'hist' | 'model' | 'clear' | null>(null)
const models = ref<ApiModel[]>([])
const currentModel = ref<ApiModel | null>(null)
const msgsRef = ref<HTMLElement>()

const showEmpty = computed(() => !currentConv.value || messages.value.length === 0)

// ===== 数据加载 =====
const loadBooks = async () => {
  const res = await bookAPI.getAll()
  if (res.success && res.data?.length) {
    books.value = res.data
    if (!selectedBookId.value) selectedBookId.value = res.data[0].id
  }
}
const loadConvos = async () => {
  const res = await conversationAPI.list()
  if (res.success && res.data) {
    conversations.value = res.data
  }
}
const loadModels = async () => {
  const res = await configAPI.getAll()
  if (res.success && res.data) {
    models.value = res.data.filter(m => m.enabled !== 0)
    if (!currentModel.value && models.value.length) currentModel.value = models.value[0]
  }
}

const onBookChange = async () => {
  closeSheets()
  currentConv.value = null
  messages.value = []
  followChips.value = []
  const res = await conversationAPI.getByBook(selectedBookId.value!)
  if (res.success && res.data) conversations.value = res.data
}

const createConvFromMenu = () => createConv()

const createConv = async () => {
  if (!selectedBookId.value) {
    ElMessage.warning('请先在「我的小说」创建一本书')
    return
  }
  const res = await conversationAPI.create({
    book_id: selectedBookId.value,
    title: '新对话 ' + new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  })
  if (res.success && res.data) {
    await loadConvos()
    currentConv.value = res.data
    messages.value = []
    followChips.value = []
    closeSheets()
  }
}

const selectConv = async (c: any) => {
  currentConv.value = c
  messages.value = []
  followChips.value = []
  closeSheets()
  const res = await conversationAPI.getMessages(c.id)
  if (res.success && res.data) {
    messages.value = res.data.map((m: any) => ({
      id: m.id,
      role: m.role,
      content: m.display_content || m.content,
      streaming: false
    }))
  }
  scrollBottom()
}

const selectModel = (m: ApiModel) => {
  currentModel.value = m
  closeSheets()
  ElMessage.success('已切换模型：' + m.name)
}

// ===== 清空 =====
const clearConv = async () => {
  if (!currentConv.value) { closeSheets(); return }
  await conversationAPI.clearMessages(currentConv.value.id)
  messages.value = []
  followChips.value = []
  closeSheets()
  ElMessage.success('已清空当前对话')
  await loadConvos()
}

// ===== 导出 =====
const exportConv = async () => {
  if (!currentConv.value) { ElMessage.info('还没有对话内容'); return }
  const title = currentConv.value.title || '星芒对话'
  const modelName = currentModel.value?.name || '未配置'
  const lines = [
    '# ' + title + '（星芒 AI 对话导出）',
    '> 模型：' + modelName + ' · 主题：' + (theme.value === 'black-gold' ? '黑金' : '白金') + ' · 导出时间：' + new Date().toLocaleString('zh-CN'),
    ''
  ]
  messages.value.forEach((m: any) => {
    lines.push((m.role === 'user' ? '**我**：' : '**星芒**：') + String(m.content || '').replace(/</g, '&lt;'))
    lines.push('')
  })
  const blob = new Blob([lines.join('\n')], { type: 'text/markdown;charset=utf-8' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = title + '-对话.md'
  a.click()
  URL.revokeObjectURL(a.href)
  closeSheets()
  ElMessage.success('已导出为 Markdown')
}

// ===== 底部弹层 =====
const openSheet = (s: 'menu' | 'hist' | 'model' | 'clear') => { sheet.value = s }
const closeSheets = () => { sheet.value = null }

// ===== 输入区 =====
const autoGrow = (e?: any) => {
  const el = e?.target as HTMLTextAreaElement | undefined
  if (!el) return
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 90) + 'px'
}
const scrollBottom = () => {
  nextTick(() => { if (msgsRef.value) msgsRef.value.scrollTop = msgsRef.value.scrollHeight })
}

const handleEnter = () => send()

// ===== 发送（SSE 流式） =====
const ensureConv = async (text: string): Promise<boolean> => {
  if (currentConv.value) return true
  if (!selectedBookId.value) {
    ElMessage.warning('请先在「我的小说」创建一本书，或在历史面板选择作品')
    return false
  }
  const res = await conversationAPI.create({
    book_id: selectedBookId.value,
    title: text.slice(0, 12) || '新对话'
  })
  if (res.success && res.data) {
    currentConv.value = res.data
    await loadConvos()
    return true
  }
  return false
}

const sendSuggestion = (q: string) => send(q)
const sendChip = (c: string) => send(c)
const sendFromBtn = () => send()

const send = async (text?: string) => {
  const content = (text ?? input.value).trim()
  if (!content || sending.value) return
  if (!(await ensureConv(content))) return
  if (!currentModel.value) {
    ElMessage.warning('请先到「API 配置」启用至少一个模型')
    return
  }
  const convId = currentConv.value!.id
  input.value = ''
  sending.value = true

  // 用户消息
  messages.value.push({ role: 'user', content })
  await conversationAPI.saveMessage(convId, { role: 'user', content })
  scrollBottom()

  // AI 占位
  const aiMsg = { role: 'assistant', content: '', streaming: true }
  messages.value.push(aiMsg)
  scrollBottom()

  const history = messages.value
    .slice(0, -1)
    .map(m => ({ role: m.role, content: m.content }))

  try {
    abortCtrl.value = new AbortController()
    const resp = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ messages: history, configId: currentModel.value.id }),
      signal: abortCtrl.value.signal
    })
    if (!resp.ok) throw new Error('HTTP ' + resp.status)

    const reader = resp.body?.getReader()
    const decoder = new TextDecoder()
    if (!reader) throw new Error('无法读取响应流')

    let buffer = ''
    let raw = ''
    let done = false
    while (true) {
      const { done: d, value } = await reader.read()
      if (d) break
      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''
      for (const line of lines) {
        if (!line.startsWith('data: ')) continue
        const data = line.slice(6)
        if (data === '[DONE]') { done = true; break }
        try {
          const parsed = JSON.parse(data)
          if (parsed.content) {
            raw += parsed.content
            // 流式模式：逐字渲染；非流式模式：收集完一次性显示
            if (streamMode.value === 'stream') {
              aiMsg.content = raw
              scrollBottom()
            }
          }
          if (parsed.error) ElMessage.error(parsed.error)
        } catch { /* skip */ }
      }
      if (done) break
    }
    if (streamMode.value !== 'stream' && raw) {
      aiMsg.content = raw
      scrollBottom()
    }

    if (raw.trim()) {
      aiMsg.streaming = false
      await conversationAPI.saveMessage(convId, { role: 'assistant', content: raw })
      followChips.value = FOLLOWUPS[followIdx.value % FOLLOWUPS.length]
      followIdx.value++
      await loadConvos()
    } else {
      messages.value.pop()
      ElMessage.error('AI 未返回内容')
    }
  } catch (err: any) {
    if (err.name !== 'AbortError') {
      messages.value.pop()
      ElMessage.error('发送失败：' + (err.message || '未知错误'))
    } else if (aiMsg.content) {
      aiMsg.streaming = false
      await conversationAPI.saveMessage(convId, { role: 'assistant', content: aiMsg.content })
      ElMessage.info('已停止生成，部分内容已保存')
    }
  } finally {
    sending.value = false
    abortCtrl.value = null
  }
}

// ===== 打开时加载 / Esc 关闭 =====
watch(() => props.visible, (v) => {
  if (v) {
    Promise.all([loadBooks(), loadModels(), loadConvos()]).then(() => {
      if (!currentConv.value && conversations.value.length) {
        selectConv(conversations.value[0])
      }
    })
  }
})
const onKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && props.visible) close()
}
onMounted(() => document.addEventListener('keydown', onKeydown))
onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown)
  abortCtrl.value?.abort()
})
</script>

<style scoped>
/* ===== 主题变量：黑金（默认）/ 白金 ===== */
.ai-chat-overlay{
  --bg1:#0a0c10;--bg2:#171b24;
  --card:#13171d;--card2:#1a2029;--card3:#0e1218;
  --line:#262c36;--line-strong:#373e4b;
  --gold:#d4af37;--gold-d:#a8862a;--gold-hi:#f3e2a8;--gold-soft:rgba(212,175,55,.12);
  --ink:#efe9d8;--sub:#8f8874;--sub2:#6b6554;
  --danger:#e5484d;--danger-soft:rgba(229,72,77,.12);
  --chip1:linear-gradient(135deg,#d4af37,#b8932e);--chip2:linear-gradient(135deg,#c9a227,#e0c35a);
  --chip3:linear-gradient(135deg,#a8862a,#d4af37);--chip4:linear-gradient(135deg,#e0c35a,#f3e2a8);
  --bub-me:linear-gradient(135deg,#d4af37,#c39b2e);--bub-me-ink:#1a1405;
  --bub-ai:#13171d;--bub-ai-border:rgba(212,175,55,.32);
  --shell:#0d0d0f;--shell-glow:0 0 60px rgba(212,175,55,.18),0 0 0 1px rgba(212,175,55,.35) inset;
  --sheet-bg:#14181e;--sheet-line:rgba(212,175,55,.2);
  --serif:"Songti SC","Noto Serif SC",Georgia,"Times New Roman",serif;
}
.ai-chat-overlay.platinum{
  --bg1:#eef0f4;--bg2:#e2e5ec;
  --card:#ffffff;--card2:#f5f6f9;--card3:#f0f2f6;
  --line:#dfe3ea;--line-strong:#c7cdd8;
  --gold:#8f9aae;--gold-d:#74809a;--gold-hi:#d8dee9;--gold-soft:rgba(143,154,174,.14);
  --ink:#23262c;--sub:#7d8494;--sub2:#a3a9b6;
  --danger:#d92d3e;--danger-soft:rgba(217,45,62,.1);
  --chip1:linear-gradient(135deg,#9aa6bb,#7f8ba3);--chip2:linear-gradient(135deg,#b6bfd1,#8f9aae);
  --chip3:linear-gradient(135deg,#7f8ba3,#9aa6bb);--chip4:linear-gradient(135deg,#d8dee9,#b6bfd1);
  --bub-me:linear-gradient(135deg,#aab3c5,#8f9aae);--bub-me-ink:#ffffff;
  --bub-ai:#ffffff;--bub-ai-border:rgba(143,154,174,.4);
  --shell:#d5d9e0;--shell-glow:0 0 60px rgba(143,154,174,.25),0 0 0 1px rgba(255,255,255,.9) inset;
  --sheet-bg:#ffffff;--sheet-line:rgba(143,154,174,.22);
}
.ai-chat-overlay *{transition:background .3s,color .3s,border-color .3s,box-shadow .3s;box-sizing:border-box}
.ai-chat-overlay{position:fixed;inset:0;z-index:9000;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,.6);backdrop-filter:blur(6px);color:var(--ink);font-family:-apple-system,BlinkMacSystemFont,"PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif;animation:fade .25s ease}
@keyframes fade{from{opacity:0}to{opacity:1}}
.ph-close{position:fixed;top:16px;right:20px;z-index:9001;width:36px;height:36px;border-radius:50%;border:1px solid var(--line-strong);background:var(--card);color:var(--gold);font-size:15px;cursor:pointer;transition:.15s}
.ph-close:hover{background:var(--gold-soft);transform:rotate(90deg)}
/* 手机壳 */
.phone{display:flex;flex-direction:column;width:375px;height:min(780px, calc(100vh - 80px));min-height:560px;background:var(--card3);border-radius:46px;border:10px solid var(--shell);box-shadow:var(--shell-glow),0 40px 80px -28px rgba(0,0,0,.6);overflow:hidden;position:relative;animation:pop .3s cubic-bezier(.32,.72,.24,1)}
@keyframes pop{from{transform:translateY(14px) scale(.97);opacity:0}to{transform:none;opacity:1}}
.status-bar{flex-shrink:0;display:flex;align-items:center;justify-content:space-between;padding:10px 24px 6px;font-size:12px;font-weight:700;color:var(--ink)}
.status-bar .sig{letter-spacing:1px}
.notch{position:absolute;top:0;left:50%;transform:translateX(-50%);width:112px;height:26px;background:var(--shell);border-radius:0 0 16px 16px;z-index:12}
/* 对话头 */
.chat-head{flex-shrink:0;display:flex;align-items:center;gap:10px;padding:8px 16px 12px}
.chat-head .back{font-size:18px;color:var(--sub);cursor:pointer}
.chat-head .ava{width:36px;height:36px;border-radius:50%;background:linear-gradient(135deg,var(--gold),var(--gold-d));color:#1a1405;display:flex;align-items:center;justify-content:center;font-size:16px;border:1px solid var(--gold-hi);box-shadow:0 4px 12px -3px var(--gold-soft);flex-shrink:0}
.head-info{min-width:0}
.chat-head .nm{font-size:15px;font-weight:800;font-family:var(--serif);letter-spacing:.04em;color:var(--ink)}
.chat-head .st{font-size:11px;color:var(--gold);font-weight:700;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;max-width:150px}
.chat-head .st::before{content:'● ';font-size:8px}
.head-acts{margin-left:auto;display:flex;align-items:center;gap:8px}
.chat-head .more{font-size:18px;color:var(--sub);cursor:pointer;letter-spacing:1px;padding:2px 4px}
.theme-switch{width:34px;height:22px;border-radius:99px;border:1px solid var(--line-strong);background:var(--card2);position:relative;cursor:pointer;flex-shrink:0}
.theme-switch .knob{position:absolute;top:2px;left:2px;width:16px;height:16px;border-radius:50%;background:var(--gold);transition:.25s;display:flex;align-items:center;justify-content:center;font-size:9px}
.ai-chat-overlay.platinum .theme-switch .knob{left:14px;color:#fff}
/* 消息区 */
.msgs{flex:1;overflow-y:auto;padding:12px 14px 8px;display:flex;flex-direction:column;gap:10px}
.msgs::-webkit-scrollbar{width:3px}
.msgs::-webkit-scrollbar-thumb{background:var(--line-strong)}
.row{display:flex;align-items:flex-end;gap:8px}
.row .av{width:28px;height:28px;border-radius:50%;flex-shrink:0;display:flex;align-items:center;justify-content:center;font-size:13px;background:linear-gradient(135deg,var(--gold),var(--gold-d));color:#1a1405;border:1px solid var(--gold-hi)}
.row.me{justify-content:flex-end}
.row.me .av{display:none}
.bub{max-width:76%;padding:9px 14px;font-size:13.5px;line-height:1.7;border-radius:16px;word-break:break-word;border:1px solid}
.row.them .bub{background:var(--bub-ai);border-color:var(--bub-ai-border);color:var(--ink);border-bottom-left-radius:5px;box-shadow:0 4px 14px -8px rgba(0,0,0,.4)}
.row.me .bub{background:var(--bub-me);border-color:transparent;color:var(--bub-me-ink);border-bottom-right-radius:5px;box-shadow:0 5px 16px -6px var(--gold-soft)}
.plain{white-space:pre-wrap}
.crt{display:inline-block;width:2px;height:13px;background:var(--gold);vertical-align:-2px;margin-left:2px;animation:blink 1s steps(1) infinite}
@keyframes blink{50%{opacity:0}}
.md-body{font-size:13.5px}
/* 空态建议卡 */
.empty{flex:1;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:6px;padding:16px 14px}
.empty .e-emoji{font-size:38px;animation:bob 3s ease-in-out infinite;filter:grayscale(.15)}
@keyframes bob{50%{transform:translateY(-5px)}}
.empty h3{font-size:17px;letter-spacing:.03em;margin-top:4px;font-family:var(--serif);color:var(--ink)}
.empty p{font-size:12px;color:var(--sub)}
.sug-grid{display:grid;grid-template-columns:1fr 1fr;gap:9px;width:100%;margin-top:14px}
.sug-card{background:var(--card);border-radius:14px;border:1px solid var(--line);padding:10px 11px;cursor:pointer;transition:.15s;text-align:left;font-family:inherit;position:relative;overflow:hidden;color:var(--ink)}
.sug-card::before{content:'';position:absolute;left:0;top:0;bottom:0;width:4px;background:linear-gradient(180deg,var(--gold),var(--gold-d))}
.sug-card:hover{transform:translateY(-2px);border-color:var(--gold);box-shadow:0 8px 20px -8px var(--gold-soft)}
.sug-card .e{font-size:19px}
.sug-card .t{font-size:12.5px;font-weight:800;margin-top:4px}
.sug-card .d{font-size:10.5px;color:var(--sub);margin-top:2px;line-height:1.4}
/* 建议芯片 */
.follow-chips{display:flex;gap:7px;flex-wrap:wrap;padding:0 2px 4px}
.fchip{font-size:11.5px;font-weight:700;color:#fff;padding:6px 12px;border-radius:99px;cursor:pointer;transition:.15s;border:none}
.fchip[data-i="0"]{background:var(--chip1)}
.fchip[data-i="1"]{background:var(--chip2)}
.fchip[data-i="2"]{background:var(--chip3)}
.fchip[data-i="3"]{background:var(--chip4)}
.fchip:hover{transform:scale(1.05);filter:brightness(1.08)}
/* 输入区 */
.dock{flex-shrink:0;padding:8px 14px 12px}
.inputrow{display:flex;align-items:center;gap:8px;background:var(--card);border:1px solid var(--line);border-radius:22px;padding:7px 8px 7px 16px}
.inputrow:focus-within{border-color:var(--gold);box-shadow:0 0 0 2px var(--gold-soft)}
textarea{flex:1;border:none;outline:none;resize:none;background:transparent;font:inherit;font-size:13.5px;color:var(--ink);max-height:90px}
textarea::placeholder{color:var(--sub2)}
.send{width:34px;height:34px;border-radius:50%;border:none;background:var(--bub-me);color:var(--bub-me-ink);font-size:15px;cursor:pointer;flex-shrink:0;transition:.15s}
.send:hover{filter:brightness(1.08);transform:scale(1.06)}
.send:disabled{opacity:.35;transform:none;cursor:not-allowed}
.home-bar{flex-shrink:0;height:8px;width:120px;margin:0 auto 6px;background:var(--line-strong);border-radius:99px;align-self:center}
/* 底部弹层 */
.mask{position:absolute;inset:0;background:rgba(0,0,0,.5);opacity:0;pointer-events:none;transition:.25s;z-index:20}
.mask.show{opacity:1;pointer-events:auto}
.sheet{position:absolute;left:0;right:0;bottom:0;background:var(--sheet-bg);border-top:1px solid var(--sheet-line);border-radius:24px 24px 0 0;padding:12px 16px 24px;transform:translateY(105%);transition:transform .32s cubic-bezier(.32,.72,.24,1);z-index:21;box-shadow:0 -12px 50px -12px rgba(0,0,0,.6);max-height:70%;overflow-y:auto}
.sheet.show{transform:none}
.sheet .grab{width:44px;height:5px;border-radius:99px;background:var(--line-strong);margin:0 auto 10px;flex-shrink:0}
.sheet h5{font-size:13px;font-weight:800;margin-bottom:6px;display:flex;align-items:center;justify-content:space-between;color:var(--ink);font-family:var(--serif);letter-spacing:.05em}
.sheet h5 .close{color:var(--sub);cursor:pointer;font-size:15px;font-weight:400;font-family:-apple-system,"PingFang SC",sans-serif}
.sitem{display:flex;align-items:center;gap:11px;width:100%;padding:12px 8px;border:none;background:none;font:inherit;font-size:14px;color:var(--ink);cursor:pointer;border-radius:11px;text-align:left;transition:.12s}
.sitem:hover{background:var(--gold-soft)}
.sitem .ic{width:32px;height:32px;border-radius:10px;display:flex;align-items:center;justify-content:center;font-size:15px;flex-shrink:0;background:var(--card2);border:1px solid var(--line)}
.sitem .sub{margin-left:auto;font-size:11px;color:var(--sub)}
.sitem.danger{color:var(--danger)}
.sitem.danger .ic{background:var(--danger-soft)}
.sitem.danger:hover{background:var(--danger-soft)}
.book-row{display:flex;align-items:center;gap:8px;padding:6px 4px 10px;border-bottom:1px solid var(--line);margin-bottom:6px}
.book-row .bl{font-size:12px;color:var(--sub);font-weight:700;flex-shrink:0}
.book-select{flex:1;min-width:0;background:var(--card2);border:1px solid var(--line);color:var(--ink);border-radius:8px;padding:6px 8px;font:inherit;font-size:12.5px}
.new-btn{border:none;background:var(--bub-me);color:var(--bub-me-ink);font:inherit;font-size:12px;font-weight:700;padding:6px 12px;border-radius:99px;cursor:pointer;flex-shrink:0}
.conv-item{display:flex;align-items:center;gap:10px;width:100%;padding:11px 10px;border:none;background:none;font:inherit;font-size:13.5px;color:var(--ink);cursor:pointer;border-radius:10px;text-align:left;border-bottom:1px solid var(--line)}
.conv-item:hover{background:var(--gold-soft)}
.conv-item .dot{width:8px;height:8px;border-radius:50%;background:var(--gold);flex-shrink:0}
.conv-item .ct{flex:1;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.conv-item .cnt{margin-left:auto;font-size:11px;color:var(--sub)}
.conv-item.cur{font-weight:800}
.conv-item .md{font-size:10.5px;color:var(--sub);font-weight:400}
.empty-tip{padding:18px;text-align:center;font-size:12px;color:var(--sub)}
.cfm-txt{font-size:12.5px;color:var(--sub);line-height:1.6}
.cfm-row{display:flex;gap:10px;margin-top:14px}
.cfm-btn{flex:1;padding:11px;border-radius:12px;font:inherit;font-size:14px;font-weight:700;cursor:pointer;border:none}
.cfm-btn.no{background:var(--card2);color:var(--sub);border:1px solid var(--line)}
.cfm-btn.yes{background:var(--danger);color:#fff}
@media (max-width:520px){
  .phone{width:100%;border-width:6px;border-radius:26px;height:calc(100vh - 40px);min-height:520px}
  .ph-close{top:10px;right:12px}
}
</style>
