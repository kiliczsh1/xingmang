<template>
  <div class="agent-page">
    <section class="agent-main">
      <div class="main-head">
        <div class="head-info">
          <span class="head-title">{{ currentSession ? currentSession.title : '星芒 Agent' }}</span>
          <el-tag v-if="currentSession" size="small" :type="currentSession.scope_kind === 'book' ? 'warning' : 'info'" effect="plain">
            {{ scopeLabel(currentSession) }}
          </el-tag>
        </div>
        <div class="head-actions">
          <span v-if="currentSession" class="cot-switch" :class="{ on: showCot }" :title="showCot ? '隐藏思考链' : '显示思考链'" @click="toggleCotVisible">
            <el-icon><View v-if="!showCot" /><Hide v-else /></el-icon>
            <span>思考链</span>
          </span>
        <el-popover v-model:visible="sessionMenuOpen" placement="bottom-end" :width="330" trigger="click" popper-class="session-menu-pop">
          <template #reference>
            <button class="head-btn" title="历史对话"><el-icon><Clock /></el-icon></button>
          </template>
          <div class="sm-head">
            <span class="sm-head-title">历史对话<span class="sm-count">{{ sessions.length }}</span></span>
            <el-button type="primary" size="small" :icon="Plus" @click="openCreate">新会话</el-button>
          </div>
          <div class="sm-list">
            <div
              v-for="s in sessions"
              :key="s.id"
              class="sm-item"
              :class="{ active: s.id === currentSession?.id }"
              @click="pickSession(s)"
            >
              <i class="sm-dot" :class="s.scope_kind"></i>
              <span class="sm-title">{{ s.title }}</span>
              <span class="sm-scope">{{ s.scope_kind === 'book' ? '书' : '全局' }}</span>
              <span class="sm-del" title="删除会话" @click.stop="removeSession(s)"><el-icon><Delete /></el-icon></span>
            </div>
            <div v-if="!sessions.length" class="sm-empty">还没有会话，点「新会话」开始</div>
          </div>
        </el-popover>
        </div>
      </div>

      <template v-if="currentSession">

        <div class="msg-list" ref="listRef">
          <div
            v-for="(group, gi) in runGroups"
            :key="group.runId || `g${gi}`"
            class="run-group"
            :class="{ selected: selectMode && selectedRunIds.has(group.runId), 'in-select': selectMode }"
            @click="selectMode && toggleSelect(group.runId)"
          >
            <div v-if="selectMode && group.runId" class="group-check" @click.stop>
              <el-checkbox
                :model-value="selectedRunIds.has(group.runId)"
                @change="toggleSelect(group.runId)"
              />
            </div>

            <!-- 用户消息（A 侧：右） -->
            <div v-for="(item, idx) in group.users" :key="'u' + idx" class="row row-user">
              <div class="msg-wrap">
                <div class="bubble-user">{{ item.text }}</div>
                <div v-if="!selectMode" class="msg-actions user-actions">
                  <span class="act" @click="copyText(item.text)"><el-icon><CopyDocument /></el-icon>复制</span>
                  <span class="act" @click="startEdit(item)"><el-icon><EditPen /></el-icon>编辑</span>
                  <span class="act" @click="resendRun(group)"><el-icon><RefreshRight /></el-icon>重新生成</span>
                  <span class="act danger" @click="deleteRunConfirm(group)"><el-icon><Delete /></el-icon>删除</span>
                </div>
              </div>
              <div class="avatar user-avatar"><el-icon><User /></el-icon></div>
            </div>

            <!-- 助手侧：一次 run 一张任务容器（过程 → 工具明细 → 回答） -->
            <div v-if="group.hasAssistantSide" class="row row-assistant">
              <div class="avatar ai-avatar"><el-icon><MagicStick /></el-icon></div>
              <div class="msg-wrap">
                <div class="task-card">
                  <!-- 过程区（由「思考链」开关控制显示） -->
                  <template v-if="showCot && cotInfo(group).show">
                    <div class="cot-head" @click="toggleCot(group.runId)">
                      <span v-if="cotInfo(group).running" class="cot-spinner"></span>
                      <span v-else class="cot-mark done">✓</span>
                      <span class="cot-title">{{ cotInfo(group).running ? '深度思考中' : '已深度思考' }}</span>
                      <span v-if="cotInfo(group).meta" class="cot-meta">{{ cotInfo(group).meta }}</span>
                      <span class="cot-toggle">
                        {{ cotInfo(group).expanded ? '收起' : '展开' }}
                        <el-icon><ArrowUp v-if="cotInfo(group).expanded" /><ArrowDown v-else /></el-icon>
                      </span>
                    </div>
                    <div v-if="cotInfo(group).expanded" class="cot-body">
                      <div v-if="cotInfo(group).think" class="cot-think">{{ cotInfo(group).think }}</div>
                      <div v-for="(ln, li) in cotInfo(group).lines" :key="li" class="cot-line">
                        <span v-if="ln.state === 'done'" class="cot-mark done">✓</span>
                        <span v-else-if="ln.state === 'error'" class="cot-mark err">✕</span>
                        <span v-else-if="ln.state === 'running'" class="cot-mark spin"></span>
                        <span class="cot-text" :class="ln.state">{{ ln.text }}</span>
                        <span v-if="ln.time" class="cot-time">{{ ln.time }}</span>
                      </div>
                      <!-- 工具明细（参数 / 结果 / 错误） -->
                      <div v-for="t in group.tools" :key="t.callId" class="tool-card" :class="{ error: t.error }">
                        <div class="tool-head" @click="t.expanded = !t.expanded">
                          <el-icon class="tool-icon"><Tools /></el-icon>
                          <span class="tool-name">{{ toolLabel(t.name) }}</span>
                          <span class="tool-brief">{{ t.brief }}</span>
                          <span class="tool-state">
                            <el-icon v-if="!t.done" class="is-loading"><Loading /></el-icon>
                            <el-icon v-else-if="t.error" class="err"><CircleCloseFilled /></el-icon>
                            <el-icon v-else class="ok"><CircleCheckFilled /></el-icon>
                          </span>
                        </div>
                        <div v-if="t.expanded" class="tool-body">
                          <div class="tool-section">
                            <div class="tool-label">参数</div>
                            <pre class="tool-pre">{{ t.args }}</pre>
                          </div>
                          <div v-if="t.result" class="tool-section">
                            <div class="tool-label">结果</div>
                            <pre class="tool-pre">{{ t.result }}</pre>
                          </div>
                          <div v-if="t.error" class="tool-section">
                            <div class="tool-label">错误</div>
                            <pre class="tool-pre err">{{ t.error }}</pre>
                          </div>
                        </div>
                      </div>
                    </div>
                  </template>

                  <!-- 失败 / 取消 -->
                  <div v-for="(st, si) in group.statuses" :key="'s' + si" class="task-alert">
                    <el-alert
                      :title="st.text"
                      :type="st.status === 'failed' ? 'error' : 'warning'"
                      :closable="false"
                      show-icon
                    />
                  </div>

                  <!-- 回答正文（可编辑：改后的版本进下一轮上下文） -->
                  <div v-if="editingAnswerRunId === group.runId" class="task-answer answer-edit">
                    <el-input
                      v-model="answerDraft"
                      type="textarea"
                      :rows="6"
                      resize="vertical"
                      placeholder="修改这一段的助手回答…"
                    />
                    <div class="answer-edit-acts">
                      <span class="answer-edit-hint">保存后该段回答被替换，后续对话按修改后的内容作为上下文记忆</span>
                      <el-button size="small" @click="cancelEditAnswer">取消</el-button>
                      <el-button size="small" type="primary" :loading="savingAnswer" @click="saveEditAnswer(group)">保存</el-button>
                    </div>
                  </div>
                  <template v-else>
                    <div v-for="(a, ai) in group.answers" :key="'a' + ai" class="task-answer">
                      <MarkdownRenderer :content="cleanAnswer(a.text)" />
                      <span v-if="!a.done" class="typing cursor">▍</span>
                    </div>

                    <!-- 执行中且还没有回答：思考动画 -->
                    <div v-if="group.thinking" class="task-answer thinking-line">
                      <span class="thinking-dots"><i></i><i></i><i></i></span>
                      <span class="thinking-text">星芒正在思考</span>
                      <span class="thinking-time">{{ (cotElapsed[group.runId] || 0).toFixed(1) }}s</span>
                    </div>
                  </template>
                </div>

                <!-- 写入提案（独立卡片，挂在任务容器下） -->
                <div v-for="(pp, pi) in group.proposals" :key="'p' + pi" class="proposal-card" :class="pp.status">
                  <div class="proposal-head">
                    <span class="proposal-tag" :class="pp.status">{{ proposalTag(pp) }}</span>
                    <span class="proposal-summary">{{ pp.summary }}</span>
                  </div>
                  <div v-if="proposalDiff(pp).length" class="proposal-diff">
                    <div v-for="(d, di) in proposalDiff(pp)" :key="di" class="diff-line" :class="d.kind">{{ d.kind === 'del' ? '− ' : '+ ' }}{{ d.text }}</div>
                  </div>
                  <div v-if="pp.status === 'pending'" class="proposal-actions">
                    <span class="proposal-note">{{ riskNote(pp) }}</span>
                    <el-button size="small" @click="decideProposal(pp, false)">拒绝</el-button>
                    <el-button size="small" type="primary" @click="decideProposal(pp, true)">批准写入</el-button>
                  </div>
                  <div v-else-if="pp.status === 'applied'" class="proposal-actions">
                    <span class="proposal-note ok">已写入，可在业务页面查看</span>
                    <el-button v-if="rollbackable(pp)" size="small" plain @click="rollback(pp)">回滚</el-button>
                  </div>
                  <div v-else-if="pp.status === 'apply_failed'" class="proposal-actions">
                    <span class="proposal-note err">写入失败，数据未改动</span>
                  </div>
                </div>

                <!-- 一次 run 只有一条操作行（重新生成为整段对话的动作，与用户侧并联等价） -->
                <div v-if="!selectMode && group.answersDone" class="msg-actions">
                  <span class="act" @click="copyText(answersText(group))"><el-icon><CopyDocument /></el-icon>复制</span>
                  <span class="act" @click="startEditAnswer(group)"><el-icon><EditPen /></el-icon>编辑</span>
                  <span class="act" @click="resendRun(group)"><el-icon><RefreshRight /></el-icon>重新生成</span>
                  <span class="act danger" @click="deleteRunConfirm(group)"><el-icon><Delete /></el-icon>删除</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 多选操作条 -->
        <div v-if="selectMode" class="select-bar">
          <span class="select-count">已选 {{ selectedRunIds.size }} 条</span>
          <el-button size="small" type="danger" plain :disabled="!selectedRunIds.size" @click="deleteSelected">
            删除所选
          </el-button>
          <el-button size="small" @click="exitSelect">取消</el-button>
        </div>

        <!-- 卡片式输入区：居中悬浮（两侧留白，中间抬高） -->
        <div class="composer-zone">
          <div class="composer-col">
            <div v-if="quickChips.length && runningRunIds.size === 0 && !selectMode && !editingRunId" class="quick-chips">
              <span v-for="c in quickChips" :key="c" class="chip" @click="sendQuick(c)">{{ c }}</span>
            </div>
            <div v-if="editingRunId" class="edit-hint">
              正在编辑历史消息，发送后将替换原条目
              <el-button link type="primary" size="small" @click="cancelEdit">取消编辑</el-button>
            </div>
            <div class="composer-card">
            <el-input
              v-model="input"
              type="textarea"
              :rows="2"
              resize="none"
              :placeholder="inputPlaceholder"
              class="composer-input"
              @keydown.enter.exact.prevent="send"
            />
            <div class="composer-controls">
              <div class="ctl-left">
                <el-popover placement="top-start" :width="170" trigger="click">
                  <template #reference>
                    <button class="ctl-btn" title="会话操作">+</button>
                  </template>
                  <div class="menu-item" @click="openCreate">新建会话</div>
                  <div class="menu-item" @click="toggleSelectMode">{{ selectMode ? '退出多选' : '多选模式' }}</div>
                  <div class="menu-item danger" @click="removeCurrentSession">删除当前会话</div>
                </el-popover>
                <el-select
                  v-model="overrideModelId"
                  placeholder="默认模型"
                  size="small"
                  clearable
                  class="model-inline"
                >
                  <el-option
                    v-for="m in models"
                    :key="m.id"
                    :value="m.id"
                    :label="agentModelLabel(m)"
                    :disabled="!m.agent_usable"
                  />
                </el-select>
                <span class="scope-pill" :class="currentSession.scope_kind" :title="currentSession.scope_kind === 'book' ? '会话创建时绑定的书，不可更改' : '全局只读作用域'">
                  <i class="dot"></i>{{ scopeLabel(currentSession) }}
                </span>
                <el-popover placement="top-start" :width="210" trigger="click">
                  <template #reference>
                    <span class="perm-pill" :class="currentSession.permission">
                      <i class="dot"></i>{{ PERM_LABELS[currentSession.permission] || '每次确认' }}
                    </span>
                  </template>
                  <div class="perm-menu-title">写入权限</div>
                  <div
                    class="perm-menu-item"
                    :class="{ active: currentSession.permission === 'observe' }"
                    @click="switchPermission('observe')"
                  >
                    <span>试运行</span><span class="perm-desc">只出提案，永不写入</span>
                  </div>
                  <div
                    class="perm-menu-item"
                    :class="{ active: currentSession.permission === 'confirm' }"
                    @click="switchPermission('confirm')"
                  >
                    <span>每次确认</span><span class="perm-desc">写入逐条批准</span>
                  </div>
                  <div
                    class="perm-menu-item"
                    :class="{ active: currentSession.permission === 'auto' }"
                    @click="switchPermission('auto')"
                  >
                    <span>自动执行</span><span class="perm-desc">新建类直写，改正文仍需批准</span>
                  </div>
                </el-popover>
              </div>
              <div class="ctl-right">
                <button
                  v-if="runningRunIds.size > 0"
                  class="send-btn stop"
                  :disabled="cancelling"
                  title="停止执行"
                  @click="cancelRunning"
                >■</button>
                <button
                  v-else
                  class="send-btn"
                  :disabled="!input.trim() || sending"
                  title="发送 (Enter)"
                  @click="send"
                >↑</button>
              </div>
            </div>
          </div>
          </div>
        </div>
      </template>

      <div v-else class="agent-empty">
        <img class="empty-icon" src="/brand/mark-128.png" width="56" height="56" alt="">
        <h3>星芒 Agent</h3>
        <p>直接说你要做什么，发送后会自动建一个会话；也可以先新建会话再绑定某本书。</p>
        <div class="composer-col empty-composer">
          <div class="composer-card">
            <div class="composer-input">
              <el-input
                v-model="input"
                type="textarea"
                :autosize="{ minRows: 2, maxRows: 8 }"
                resize="none"
                placeholder="给星芒一个任务…（发送即新建会话）"
                @keydown.enter.exact.prevent="send"
              />
            </div>
            <div class="composer-ctl">
              <span class="empty-hint">将按「全局作用域 · 每次确认」创建</span>
              <button class="send-btn" :disabled="!input.trim() || sending" title="发送 (Enter)" @click="send">↑</button>
            </div>
          </div>
        </div>
        <el-button link type="primary" size="small" class="empty-new" @click="openCreate">新建会话（可绑定书 / 改写入权限）</el-button>
      </div>
    </section>

    <!-- 新建会话弹窗 -->
    <el-dialog v-model="dialogVisible" title="新建 Agent 会话" width="460px" class="agent-dialog">
      <el-form label-width="80px">
        <el-form-item label="作用域">
          <el-radio-group v-model="form.scope_kind">
            <el-radio-button value="global">全局（跨书检索）</el-radio-button>
            <el-radio-button value="book">绑定单本书</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.scope_kind === 'book'" label="选择书籍">
          <el-select v-model="form.scope_id" placeholder="选择要绑定的书" style="width: 100%">
            <el-option v-for="b in books" :key="b.id" :value="b.id" :label="b.title" />
          </el-select>
        </el-form-item>
        <el-form-item label="模型">
          <el-select v-model="form.model_id" placeholder="使用默认模型" style="width: 100%" clearable>
            <el-option
              v-for="m in models"
              :key="m.id"
              :value="m.id"
              :label="agentModelLabel(m)"
              :disabled="!m.agent_usable"
            />
          </el-select>
          <div class="form-hint">Agent 需要 function calling 支持，仅 OpenAI 兼容格式可用</div>
        </el-form-item>
        <el-form-item label="写入权限">
          <el-radio-group v-model="form.permission">
            <el-radio-button value="observe">试运行</el-radio-button>
            <el-radio-button value="confirm">每次确认</el-radio-button>
            <el-radio-button value="auto">自动执行</el-radio-button>
          </el-radio-group>
          <div class="form-hint">自动执行 = 新建类直接写入；改正文任何档位都要批准。所有档位都不能删除数据。</div>
        </el-form-item>
        <el-form-item label="标题">
          <el-input v-model="form.title" placeholder="留空则用首条输入命名" maxlength="30" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!canCreate" @click="createSession">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onUnmounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, Delete, Tools, Loading, CircleCheckFilled, CircleCloseFilled, MagicStick,
  CopyDocument, EditPen, RefreshRight, User, View, Hide, ArrowUp, ArrowDown, Clock
} from '@element-plus/icons-vue'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import { agentAPI } from '@/api'

// ===== 类型 =====
interface AgentSession {
  id: string
  title: string
  scope_kind: 'global' | 'book'
  scope_id: number | null
  model_id: number | null
  permission: 'observe' | 'confirm' | 'auto'
}
interface AgentModel {
  id: number
  name: string
  model: string
  is_default: number
  provider_name: string
  api_format: string
  agent_usable: boolean
}
interface LedgerEvent {
  seq: number
  type: string
  run_id?: string
  payload: any
  created_at?: string
}
type ChatItem =
  | { kind: 'user'; runId: string; text: string }
  | { kind: 'assistant'; runId: string; text: string; done: boolean }
  | { kind: 'tool'; runId: string; callId: string; name: string; args: string; brief: string; result: string; error: string; done: boolean; expanded: boolean; duration?: number }
  | { kind: 'status'; runId: string; status: string; text: string }
  | { kind: 'proposal'; runId: string; proposalId: string; toolName: string; target: string; risk: string; summary: string; status: string; args: any }

// ===== 状态 =====
const sessions = ref<AgentSession[]>([])
const currentSession = ref<AgentSession | null>(null)
const models = ref<AgentModel[]>([])
const books = ref<{ id: number; title: string }[]>([])
const overrideModelId = ref<number | null>(null)
const items = ref<ChatItem[]>([])
const input = ref('')
const sending = ref(false)
const cancelling = ref(false)
const dialogVisible = ref(false)
// 历史对话下拉（右上角时钟）
const sessionMenuOpen = ref(false)
const listRef = ref<HTMLElement | null>(null)
const runningRunIds = reactive(new Set<string>())
const form = reactive<{ title: string; scope_kind: 'global' | 'book'; scope_id: number | null; model_id: number | null; permission: 'observe' | 'confirm' | 'auto' }>({
  title: '', scope_kind: 'global', scope_id: null, model_id: null, permission: 'confirm'
})

let lastSeq = 0
let streamCtrl: AbortController | null = null

// 消息操作状态
const selectMode = ref(false)
const selectedRunIds = reactive(new Set<string>())
const editingRunId = ref<string | null>(null)
// 助手回答编辑态（一次只编辑一段对话的回答）
const editingAnswerRunId = ref<string | null>(null)
const answerDraft = ref('')
const savingAnswer = ref(false)

// ===== 思考卡（CoT）状态 =====
// run 开始时间 / 已用秒数 / 展开状态；工具调用映射为卡内步骤
const runStartAt: Record<string, number> = {}
const callStartAt: Record<string, number> = {}
const cotElapsed = reactive<Record<string, number>>({})
const cotExpanded = reactive<Record<string, boolean>>({})
// 模型思考过程（thinking.delta 累积；只渲染在思考卡里，绝不进回答气泡）
const thinkingBuf = reactive<Record<string, string>>({})

// 思考链全局开关（记 localStorage；关掉后只在回答区显示思考动画，不显示步骤卡）
const showCot = ref(localStorage.getItem('xingmang.agent.showCot') !== '0')
function toggleCotVisible() {
  showCot.value = !showCot.value
  localStorage.setItem('xingmang.agent.showCot', showCot.value ? '1' : '0')
}

// run 结束（完成/失败/取消）时把还空着的思考占位收掉，避免一直闪
function finishEmptyAssistant(runId: string) {
  const last = lastAssistantOf(runId) as any
  if (last && !last.done && !last.text) last.done = true
}

const cotTicker = window.setInterval(() => {
  for (const id of runningRunIds) {
    const st = runStartAt[id]
    if (st) cotElapsed[id] = (Date.now() - st) / 1000
  }
}, 250)

// ⚠ SQLite 的 CURRENT_TIMESTAMP 是 UTC（"2026-10-03 13:09:30"），必须补 Z 按 UTC 解析，
// 否则按本地时区解析会差一个时区偏移（曾把耗时显示成 28810s = 多 8 小时）。
function parseServerTime(s?: string): number {
  if (!s) return 0
  const t = new Date(s.replace(' ', 'T') + 'Z').getTime()
  return isNaN(t) ? 0 : t
}

function truncErr(e: string): string {
  const r = [] as any
  const chars = Array.from(e || '')
  return chars.length > 60 ? chars.slice(0, 60).join('') + '…' : e
}

// 防御性剥离回答里的推理残留：
// 后端已把 <think> 块分流到思考卡，但旧账本里（本次修复前）存下的回答仍带标签，
// 历史消息重放时按显示剥离，避免英文思维链又冒到气泡里。
function cleanAnswer(text: string): string {
  if (!text) return ''
  let out = text.replace(/<think(?:ing)?>[\s\S]*?<\/think(?:ing)?>/gi, '')
  out = out.replace(/<think(?:ing)?>[\s\S]*$/i, '') // 未闭合的截断块（流式中的中间态）
  return out.replace(/^\s+/, '')
}

// 思考卡信息：是否有卡 / 是否进行中 / 摘要 meta / 步骤行
function cotInfo(group: { runId: string; items: ChatItem[] }) {
  const runId = group.runId
  const tools = group.items.filter(i => i.kind === 'tool') as any[]
  const thinkText = (thinkingBuf[runId] || '').trim()
  if ((!tools.length && !thinkText) || !runId) return { show: false }
  const assistantDone = group.items.some(i => i.kind === 'assistant' && i.done && !!i.text)
  const textStarted = group.items.some(i => i.kind === 'assistant' && i.text)
  const running = runningRunIds.has(runId)
  const expanded = cotExpanded[runId] ?? !(textStarted || assistantDone)
  const doneCount = tools.filter(t => t.done && !t.error).length
  const secs = cotElapsed[runId]
  const meta = [
    secs != null ? `${running && !assistantDone ? '已思考' : '用时'} ${secs.toFixed(1)}s` : '',
    doneCount ? `查阅 ${doneCount} 处` : ''
  ].filter(Boolean).join(' · ')

  const lines: { state: 'done' | 'error' | 'running'; text: string; time: string }[] = []
  if (tools.length) lines.push({ state: 'done', text: '分析任务', time: '' })
  for (const t of tools) {
    const label = toolLabel(t.name) + (t.brief ? ` · ${t.brief}` : '')
    const verb = WRITE_TOOLS.has(t.name) ? (t.name.startsWith('chapter') ? '写入' : '新建') : '查阅'
    if (!t.done) lines.push({ state: 'running', text: verb + ' ' + label, time: '' })
    else if (t.error) lines.push({ state: 'error', text: verb + ' ' + label + ' 失败：' + truncErr(t.error), time: '' })
    else lines.push({ state: 'done', text: verb + ' ' + label, time: t.duration ? t.duration.toFixed(1) + 's' : '' })
  }
  if (tools.length) lines.push({ state: assistantDone ? 'done' : 'running', text: '组织回答', time: '' })
  return { show: true, running: running && !assistantDone, expanded, lines, meta, think: thinkText }
}

function toggleCot(runId: string) {
  cotExpanded[runId] = !(cotExpanded[runId] ?? true)
}

// 按 run 分组：一条对话 = 用户消息 + 助手任务容器（过程 / 工具明细 / 回答 / 状态）
// 派生字段供模板直接用，避免在模板里反复过滤（run 内条目不多，成本可忽略）。
const runGroups = computed(() => {
  const groups: any[] = []
  for (const it of items.value) {
    const runId = (it as any).runId || ''
    const last = groups[groups.length - 1]
    if (!last || last.runId !== runId) {
      groups.push({
        runId,
        items: [it],
        users: it.kind === 'user' ? [it] : [],
        answers: it.kind === 'assistant' && (it as any).text ? [it] : [],
        tools: it.kind === 'tool' ? [it] : [],
        proposals: it.kind === 'proposal' ? [it] : [],
        statuses: it.kind === 'status' ? [it] : []
      })
    } else {
      last.items.push(it)
      if (it.kind === 'user') last.users.push(it)
      else if (it.kind === 'assistant' && (it as any).text) last.answers.push(it)
      else if (it.kind === 'tool') last.tools.push(it)
      else if (it.kind === 'proposal') last.proposals.push(it)
      else if (it.kind === 'status') last.statuses.push(it)
    }
  }
  const running = (id: string) => runningRunIds.has(id)
  for (const g of groups) {
    g.hasAssistantSide =
      !!g.answers.length || !!g.tools.length || !!g.proposals.length || !!g.statuses.length || running(g.runId)
    g.thinking = running(g.runId) && !g.answers.length
    g.answersDone = !!g.answers.length && g.answers.every((a: any) => a.done) && !running(g.runId)
  }
  return groups
})

// 复制整条回答（拼接该 run 的全部正文）
function answersText(group: { answers: any[] }): string {
  return group.answers.map((a: any) => cleanAnswer(a.text)).filter(Boolean).join('\n\n')
}

// ===== 编辑助手回答 =====
// 把编辑后的文本写回该段对话的回答（本地即时生效；事件流也会推 answer.edited，幂等）
function applyEditedAnswer(runId: string, text: string) {
  const last = lastAssistantOf(runId) as any
  if (last) {
    last.text = text
    last.done = true
  } else {
    items.value.push({ kind: 'assistant', runId, text, done: true })
  }
}

function startEditAnswer(group: { runId: string; answers: any[] }) {
  if (busyOrNoSession()) return
  cancelEdit() // 退出「编辑用户输入」态，两个编辑态不并存
  editingAnswerRunId.value = group.runId
  answerDraft.value = answersText(group)
}

function cancelEditAnswer() {
  editingAnswerRunId.value = null
  answerDraft.value = ''
}

async function saveEditAnswer(group: { runId: string }) {
  const text = answerDraft.value.trim()
  if (!text) {
    ElMessage.warning('回答内容不能为空')
    return
  }
  savingAnswer.value = true
  try {
    await agentAPI.editRunAnswer(group.runId, text)
    applyEditedAnswer(group.runId, text)
    editingAnswerRunId.value = null
    answerDraft.value = ''
    ElMessage.success('已保存，后续对话按修改后的内容作为上下文')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '保存失败')
  } finally {
    savingAnswer.value = false
  }
}

const TOOL_LABELS: Record<string, string> = {
  book_list: '列出书籍',
  book_read: '读取书籍结构',
  chapter_list: '章节列表',
  chapter_read: '读取章节正文',
  character_list: '角色卡',
  graph_query: '知识图谱查询'
}

const canCreate = computed(() =>
  form.scope_kind === 'global' || (form.scope_id != null && form.scope_id > 0)
)

const inputPlaceholder = computed(() => {
  if (runningRunIds.size > 0) return 'Agent 正在执行，可点「停止」中断…'
  if (currentSession.value?.scope_kind === 'book') return '问问这本书的情节、设定、伏笔…'
  return '全局模式：跨书检索、通用写作问答…'
})

// 快捷任务 chips（方案 B：绑定书会话给书内任务，全局会给库级任务）
const quickChips = computed(() => {
  if (!currentSession.value) return []
  return currentSession.value.scope_kind === 'book'
    ? ['查一下这本书的设定', '捋一遍时间线', '找前文伏笔']
    : ['列出我的书', '哪本书章节数最多']
})

function toolLabel(name: string) {
  return TOOL_LABELS[name] || name
}

function agentModelLabel(m: AgentModel) {
  return `${m.name}（${m.provider_name}）${m.agent_usable ? '' : ' · 不支持工具'}`
}

function scopeLabel(s: AgentSession) {
  if (s.scope_kind === 'book') {
    const b = books.value.find(x => x.id === s.scope_id)
    return `书：${b?.title || ('#' + s.scope_id)}`
  }
  return '全局'
}

function briefOf(name: string, args: any): string {
  try {
    if (name === 'chapter_read' && args?.chapter_id != null) return `章节 #${args.chapter_id}`
    if (name === 'graph_query' && args?.keyword) return `关键词「${args.keyword}」`
    if ((name === 'book_read' || name === 'chapter_list' || name === 'character_list') && args?.book_id != null) return `书 #${args.book_id}`
  } catch { /* ignore */ }
  return ''
}

// ===== 数据加载 =====
async function loadMeta() {
  const [ms, bs] = await Promise.all([agentAPI.models(), agentAPI.books()])
  models.value = (ms as any).data || []
  books.value = (bs as any).data || []
}

async function loadSessions() {
  const res: any = await agentAPI.sessions()
  sessions.value = res.data || []
}

// 上次打开的会话记在本地；回来时优先恢复它（找不到就退到最近更新的一条）
const LAST_SESSION_KEY = 'xingmang.agent.lastSession'

function rememberSession(id: string) {
  try { localStorage.setItem(LAST_SESSION_KEY, id) } catch { /* 隐私模式忽略 */ }
}

function forgetSession(id?: string) {
  try {
    if (!id || localStorage.getItem(LAST_SESSION_KEY) === id) localStorage.removeItem(LAST_SESSION_KEY)
  } catch { /* 忽略 */ }
}

// 初始化：拉列表 → 恢复上次会话 → 没有则退到最近更新的一条 → 都没有就留空状态（首屏即可输入）
async function initSessions() {
  try {
    await loadSessions()
  } catch { return }
  const list = sessions.value as AgentSession[]
  if (!list.length) return
  let lastID = ''
  try { lastID = localStorage.getItem(LAST_SESSION_KEY) || '' } catch { /* 忽略 */ }
  const target = (lastID && list.find((x: AgentSession) => x.id === lastID)) || list[0]
  if (target) selectSession(target)
}

// ===== 会话与事件流 =====
async function selectSession(s: AgentSession) {
  if (currentSession.value?.id === s.id) return
  currentSession.value = s
  rememberSession(s.id)
  items.value = []
  lastSeq = 0
  runningRunIds.clear()
  overrideModelId.value = null
  cancelEdit()
  cancelEditAnswer()
  await connectStream()
}

async function connectStream() {
  const sid = currentSession.value?.id
  if (!sid) return
  streamCtrl?.abort()
  const ctrl = new AbortController()
  streamCtrl = ctrl

  try {
    const res = await fetch(`/api/agent/sessions/${sid}/stream?after_seq=${lastSeq}`, { signal: ctrl.signal })
    if (!res.ok || !res.body) throw new Error(`连接失败(${res.status})`)
    const reader = res.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })
      let idx: number
      while ((idx = buffer.indexOf('\n\n')) >= 0) {
        const block = buffer.slice(0, idx)
        buffer = buffer.slice(idx + 2)
        for (const line of block.split('\n')) {
          if (!line.startsWith('data:')) continue
          try {
            handleEvent(JSON.parse(line.slice(5).trim()))
          } catch { /* 忽略无法解析的行 */ }
        }
      }
    }
    // 服务端断开（如重启）：自动重连续传
    if (!ctrl.signal.aborted && currentSession.value?.id === sid) {
      await new Promise(r => setTimeout(r, 1000))
      if (currentSession.value?.id === sid) await connectStream()
    }
  } catch (e: any) {
    if (e?.name !== 'AbortError') {
      // 连接失败静默重试，不打断使用
      setTimeout(() => { if (currentSession.value?.id === sid) connectStream() }, 2000)
    }
  }
}

function handleEvent(ev: LedgerEvent) {
  // seq 严格连续性校验：出现空洞说明漏事件，全量重拉
  if (ev.seq !== lastSeq + 1) {
    fullReload()
    return
  }
  lastSeq = ev.seq
  const p = ev.payload || {}
  const runId = ev.run_id || p.run_id || ''

  switch (ev.type) {
    case 'run.created':
      runningRunIds.add(runId)
      runStartAt[runId] = parseServerTime(ev.created_at) || Date.now()
      cotElapsed[runId] = 0
      cotExpanded[runId] = true
      items.value.push({ kind: 'user', runId, text: p.input || '' })
      // 立即挂一个空的助手占位：不等第一个 token 就能显示思考动画
      items.value.push({ kind: 'assistant', runId, text: '', done: false })
      break
    case 'model.delta': {
      // 回答开始输出 → 思考卡自动收起（DeepSeek 式）
      if ((p.text || '') !== '' && cotExpanded[runId] !== false) cotExpanded[runId] = false
      const last = lastAssistantOf(runId)
      if (last && !last.done) last.text += p.text || ''
      else items.value.push({ kind: 'assistant', runId, text: p.text || '', done: false })
      break
    }
    case 'thinking.delta': {
      // 思考过程单独累积，只进思考卡
      thinkingBuf[runId] = (thinkingBuf[runId] || '') + (p.text || '')
      break
    }
    case 'answer.edited': {
      // 用户改了助手回答：显示与上下文都按新版本
      applyEditedAnswer(runId, p.text || '')
      break
    }    case 'model.completed': {
      const last = lastAssistantOf(runId)
      if (last && !last.done) { last.text = p.text || last.text; last.done = true }
      else if (p.text) items.value.push({ kind: 'assistant', runId, text: p.text, done: true })
      break
    }
    case 'tool.started':
      closeAssistantOf(runId)
      callStartAt[p.call_id || ''] = parseServerTime(ev.created_at) || Date.now()
      items.value.push({
        kind: 'tool', runId, callId: p.call_id || '', name: p.name || '',
        args: prettyJSON(p.arguments), brief: briefOf(p.name, typeof p.arguments === 'string' ? safeParse(p.arguments) : p.arguments),
        result: '', error: '', done: false, expanded: false
      })
      break
    case 'tool.completed': {
      const t = items.value.find(i => i.kind === 'tool' && i.callId === p.call_id) as any
      if (t) {
        t.done = true
        if (p.error) t.error = p.error
        else t.result = prettyJSON(p.result)
        const st = callStartAt[p.call_id] || 0
        const en = parseServerTime(ev.created_at) || Date.now()
        if (st && en > st) t.duration = (en - st) / 1000
      }
      break
    }
    case 'run.status':
      finishEmptyAssistant(runId)
      if (p.status === 'failed' || p.status === 'cancelled') {
        items.value.push({
          kind: 'status', runId, status: p.status,
          text: p.status === 'failed' ? `执行失败：${p.error || '未知错误'}` : '已取消'
        })
      }
      break
    case 'proposal.created':
      items.value.push({
        kind: 'proposal', runId, proposalId: p.proposal_id, toolName: p.tool || '',
        target: p.target || '', risk: p.risk || 'low', summary: p.summary || '',
        status: p.status || 'pending', args: typeof p.arguments === 'string' ? safeParse(p.arguments) : (p.arguments || {})
      })
      break
    case 'proposal.decided': {
      const it = items.value.find(i => i.kind === 'proposal' && i.proposalId === p.proposal_id) as any
      if (it) {
        it.status = p.status
        if (p.result) it.summary = p.result
      }
      break
    }
    case 'run.completed': {
      runningRunIds.delete(runId)
      cotExpanded[runId] = false
      finishEmptyAssistant(runId)
      const st = runStartAt[runId]
      const en = parseServerTime(ev.created_at) || Date.now()
      if (st && en > st) cotElapsed[runId] = (en - st) / 1000
      break
    }
  }
  scrollToBottom()
}

async function fullReload() {
  // 全量重拉：清空重建，从头订阅
  items.value = []
  lastSeq = 0
  runningRunIds.clear()
  await connectStream()
}

function lastAssistantOf(runId: string) {
  for (let i = items.value.length - 1; i >= 0; i--) {
    const it = items.value[i] as any
    if (it.kind === 'assistant' && it.runId === runId) return it
    if (it.kind === 'user' && it.runId === runId) return null
  }
  return null
}

function closeAssistantOf(runId: string) {
  const last = lastAssistantOf(runId) as any
  if (last && !last.done) last.done = true
}

function prettyJSON(v: any): string {
  if (v == null) return ''
  if (typeof v === 'string') {
    try { return JSON.stringify(JSON.parse(v), null, 2) } catch { return v }
  }
  try { return JSON.stringify(v, null, 2) } catch { return String(v) }
}

function safeParse(v: any): any {
  if (typeof v === 'string') { try { return JSON.parse(v) } catch { return {} } }
  return v || {}
}

let scrollTimer: number | undefined
function scrollToBottom() {
  if (scrollTimer) return
  scrollTimer = window.setTimeout(() => {
    scrollTimer = undefined
    const el = listRef.value
    if (el) el.scrollTop = el.scrollHeight
  }, 80)
}

// ===== 操作 =====
// 懒创建：没有会话时用首条输入当标题自动建一个（全局作用域 + 默认权限档）
async function ensureSession(firstInput: string): Promise<AgentSession | null> {
  if (currentSession.value) return currentSession.value
  try {
    const res: any = await agentAPI.createSession({
      title: Array.from(firstInput.trim()).slice(0, 24).join('') || '新会话',
      scope_kind: 'global',
      scope_id: null,
      model_id: null,
      permission: 'confirm'
    })
    const created = res.data as AgentSession
    await loadSessions()
    if (created) selectSession(created)
    return created || null
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '创建会话失败')
    return null
  }
}

async function send() {
  const text = input.value.trim()
  if (!text || runningRunIds.size > 0 || sending.value) return
  sending.value = true
  input.value = ''
  try {
    if (!currentSession.value) {
      const created = await ensureSession(text)
      if (!created) {
        input.value = text
        return
      }
    }
    // 编辑态：先删掉原 run（含其在上下文中的历史），再以新文本发起 run
    if (editingRunId.value) {
      const rid = editingRunId.value
      editingRunId.value = null
      try { await agentAPI.deleteRun(rid) } catch { /* 原条目可能已被删 */ }
      await fullReload()
    }
    const session = currentSession.value
    if (!session) {
      input.value = text
      return
    }
    await agentAPI.startRun(session.id, {
      input: text,
      model_id: overrideModelId.value ?? undefined
    })
  } catch (e: any) {
    input.value = text // 发送失败把内容还回去
    ElMessage.error(e?.response?.data?.message || e?.message || '发送失败')
  } finally {
    sending.value = false
  }
}

function busyOrNoSession(): boolean {
  if (!currentSession.value) return false
  if (runningRunIds.size > 0) {
    ElMessage.warning('Agent 正在执行中，请先停止或等待完成')
    return true
  }
  return false
}

// 点击快捷 chip：直接以该任务发起 run
async function sendQuick(text: string) {
  if (busyOrNoSession() || sending.value) return
  try {
    await agentAPI.startRun(currentSession.value!.id, {
      input: text,
      model_id: overrideModelId.value ?? undefined
    })
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '发送失败')
  }
}

function pickSession(s: AgentSession) {
  selectSession(s)
  sessionMenuOpen.value = false
}

function removeCurrentSession() {
  if (currentSession.value) removeSession(currentSession.value)
}

// ===== 写入提案 =====
const WRITE_TOOLS = new Set(['chapter_append', 'chapter_replace', 'book_create', 'character_create', 'graph_entity_create', 'graph_relation_create'])

function proposalTag(item: any): string {
  switch (item.status) {
    case 'applied': return '已写入'
    case 'rejected': return '已拒绝'
    case 'rolled_back': return '已回滚'
    case 'apply_failed': return '写入失败'
    default: return item.risk === 'high' ? '高风险 · 等待批准' : '写入提案'
  }
}

function riskNote(item: any): string {
  return item.risk === 'high'
    ? '修改正文，批准前请核对 diff；写入后原句保留可回滚'
    : '低风险创建类操作，写入后可在对应页面查看'
}

// 从提案参数合成 diff 行（append: +内容；replace: -find +replace_with）
function proposalDiff(item: any): { kind: 'del' | 'add'; text: string }[] {
  const a = item.args || {}
  if (item.toolName === 'chapter_append') {
    const c = String(a.content || '')
    return c ? [{ kind: 'add', text: truncLines(c, 160) }] : []
  }
  if (item.toolName === 'chapter_replace') {
    const out: { kind: 'del' | 'add'; text: string }[] = []
    const f = String(a.find || '')
    const r = String(a.replace_with || '')
    if (f) out.push({ kind: 'del', text: truncLines(f, 120) })
    if (r) out.push({ kind: 'add', text: truncLines(r, 120) })
    return out
  }
  return []
}

function truncLines(s: string, max: number): string {
  const chars = Array.from(s.replace(/\s+/g, ' ').trim())
  return chars.length > max ? chars.slice(0, max).join('') + '…' : chars.join('')
}

function rollbackable(item: any): boolean {
  return item.status === 'applied' && (item.toolName === 'chapter_append' || item.toolName === 'chapter_replace')
}

async function decideProposal(item: any, approve: boolean) {
  if (approve) {
    try {
      const res: any = await agentAPI.approveProposal(item.proposalId)
      item.status = res.data?.status || 'applied'
      if (res.data?.result) item.summary = res.data.result
      ElMessage.success('已写入')
    } catch (e: any) {
      ElMessage.error(e?.response?.data?.message || '写入失败')
    }
  } else {
    try {
      await agentAPI.rejectProposal(item.proposalId)
      item.status = 'rejected'
    } catch (e: any) {
      ElMessage.error(e?.response?.data?.message || '操作失败')
    }
  }
}

async function rollback(item: any) {
  try {
    await ElMessageBox.confirm('用提案前的原文快照还原该章节？此提案之后的其他改动不会撤销。', '回滚', { type: 'warning' })
  } catch { return }
  try {
    await agentAPI.rollbackProposal(item.proposalId)
    item.status = 'rolled_back'
    ElMessage.success('已回滚')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '回滚失败')
  }
}

// ===== 写入权限 =====
const PERM_LABELS: Record<string, string> = { observe: '试运行', confirm: '每次确认', auto: '自动执行' }

function switchPermission(perm: 'observe' | 'confirm' | 'auto') {
  if (!currentSession.value || currentSession.value.permission === perm) return
  agentAPI.patchSession(currentSession.value.id, { permission: perm }).then((res: any) => {
    if (currentSession.value && res.data) currentSession.value.permission = res.data.permission
    ElMessage.success(`写入权限已切换为「${PERM_LABELS[perm]}」`)
  }).catch((e: any) => {
    ElMessage.error(e?.response?.data?.message || '切换失败')
  })
}

async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制')
  } catch {
    ElMessage.error('复制失败：浏览器未授权剪贴板')
  }
}

function startEdit(item: any) {
  if (busyOrNoSession()) return
  input.value = item.text
  editingRunId.value = item.runId
}

function cancelEdit() {
  editingRunId.value = null
  input.value = ''
}

// 重新生成：删掉该条对话（旧回答 + 工具记录），保留原输入重新执行，让模型重新给一次结果
async function resendRun(group: { runId: string; items: ChatItem[] }) {
  if (busyOrNoSession()) return
  const u = group.items.find(i => i.kind === 'user') as any
  if (!u?.text || !currentSession.value) return
  try {
    await ElMessageBox.confirm(
      '丢弃当前的回答并重新生成？原回答与本次的工具调用记录会被删除。',
      '重新生成',
      { type: 'warning', confirmButtonText: '重新生成', cancelButtonText: '取消' }
    )
  } catch { return }
  try {
    await agentAPI.deleteRun(group.runId)
    await fullReload()
    await agentAPI.startRun(currentSession.value.id, {
      input: u.text,
      model_id: overrideModelId.value ?? undefined
    })
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '重新生成失败')
    await fullReload()
  }
}

function deleteRunConfirm(group: { runId: string; items: ChatItem[] }) {
  if (busyOrNoSession() || !group.runId) return
  ElMessageBox.confirm(
    '删除这条对话？用户消息、回复与工具调用记录一并移除，后续对话不再引用它。',
    '删除对话',
    { type: 'warning' }
  ).then(() => deleteRuns([group.runId])).catch(() => {})
}

async function deleteRuns(ids: string[]) {
  try {
    for (const id of ids) {
      await agentAPI.deleteRun(id)
    }
    if (editingRunId.value && ids.includes(editingRunId.value)) {
      editingRunId.value = null
      input.value = ''
    }
    if (editingAnswerRunId.value && ids.includes(editingAnswerRunId.value)) cancelEditAnswer()
    if (editingAnswerRunId.value && ids.includes(editingAnswerRunId.value)) cancelEditAnswer()
    ElMessage.success(ids.length > 1 ? `已删除 ${ids.length} 条` : '已删除')
    await fullReload() // 删除后 seq 已重排，全量重拉对齐游标
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '删除失败')
    await fullReload()
  }
}

function toggleSelect(runId: string) {
  if (!runId) return
  if (selectedRunIds.has(runId)) selectedRunIds.delete(runId)
  else selectedRunIds.add(runId)
}

function toggleSelectMode() {
  selectMode.value = !selectMode.value
  if (!selectMode.value) selectedRunIds.clear()
}

function exitSelect() {
  selectMode.value = false
  selectedRunIds.clear()
}

function deleteSelected() {
  if (!selectedRunIds.size) return
  ElMessageBox.confirm(
    `删除所选 ${selectedRunIds.size} 条对话？不可恢复。`,
    '批量删除',
    { type: 'warning' }
  ).then(() => deleteRuns([...selectedRunIds])).catch(() => {})
}

async function cancelRunning() {
  if (!currentSession.value || runningRunIds.size === 0) return
  cancelling.value = true
  try {
      for (const runId of [...runningRunIds]) {
      await agentAPI.cancelRun(runId).catch(() => {})
    }
  } finally {
    cancelling.value = false
  }
}

function openCreate() {
  form.title = ''
  form.scope_kind = 'global'
  form.scope_id = null
  form.model_id = null
  form.permission = currentSession.value?.permission || 'confirm'
  dialogVisible.value = true
  if (!models.value.length) loadMeta()
}

async function createSession() {
  try {
    const api = agentAPI
    const res: any = await api.createSession({
      title: form.title,
      scope_kind: form.scope_kind,
      scope_id: form.scope_kind === 'book' ? form.scope_id : null,
      model_id: form.model_id,
      permission: form.permission
    })
    dialogVisible.value = false
    await loadSessions()
    const s = res.data as AgentSession
    if (s) selectSession(s)
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '创建失败')
  }
}

async function removeSession(s: AgentSession) {
  try {
    await ElMessageBox.confirm(`删除会话「${s.title}」？历史记录一并清除，不可恢复。`, '删除会话', { type: 'warning' })
  } catch { return }
  try {
      await agentAPI.deleteSession(s.id)
    if (currentSession.value?.id === s.id) {
      streamCtrl?.abort()
      streamCtrl = null
      currentSession.value = null
      items.value = []
      lastSeq = 0
    }
    forgetSession(s.id)
    await loadSessions()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '删除失败')
  }
}

// ===== 生命周期 =====
loadMeta()
initSessions()

watch(runningRunIds, () => { /* 触发输入框占位与按钮状态刷新 */ })

onUnmounted(() => {
  streamCtrl?.abort()
  streamCtrl = null
  clearInterval(cotTicker)
})
</script>

<style scoped>
.agent-page {
  height: 100%;
  min-height: 0;
  display: flex;
  gap: 12px;
  color: var(--el-text-color-primary);
}

/* ---------- 左栏 ---------- */

/* ---------- 右栏 ---------- */
.agent-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  overflow: hidden;
}
.head-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  margin-left: auto;
}
.head-btn {
  width: 30px;
  height: 30px;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  background: transparent;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: color .15s, border-color .15s;
}
.head-btn:hover {
  color: var(--accent, var(--el-color-primary));
  border-color: var(--accent, var(--el-color-primary));
}
/* 历史对话下拉里的条目 */
.sm-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 8px;
  margin-bottom: 6px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.sm-head-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.sm-count {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color);
  border-radius: 999px;
  padding: 0 6px;
}
.sm-list {
  max-height: 52vh;
  overflow-y: auto;
  margin: 0 -4px;
  padding: 0 4px;
}
.sm-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 8px;
  border-radius: 6px;
  cursor: pointer;
}
.sm-item:hover {
  background: var(--el-fill-color-light);
}
.sm-item.active {
  background: var(--accent-soft, var(--el-fill-color));
}
.sm-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
  background: var(--el-color-success);
}
.sm-dot.book {
  background: var(--el-color-warning);
}
.sm-title {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.sm-scope {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}
.sm-del {
  display: none;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}
.sm-item:hover .sm-del {
  display: inline-flex;
}
.sm-del:hover {
  color: var(--el-color-danger);
}
.sm-empty {
  padding: 18px 8px;
  text-align: center;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.head-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  margin-left: auto;
}
.head-btn {
  width: 30px;
  height: 30px;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  background: transparent;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: color .15s, border-color .15s;
}
.head-btn:hover {
  color: var(--accent, var(--el-color-primary));
  border-color: var(--accent, var(--el-color-primary));
}
/* 历史对话下拉里的条目 */
.sm-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 8px;
  margin-bottom: 6px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.sm-head-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.sm-count {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color);
  border-radius: 999px;
  padding: 0 6px;
}
.sm-list {
  max-height: 52vh;
  overflow-y: auto;
  margin: 0 -4px;
  padding: 0 4px;
}
.sm-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 8px;
  border-radius: 6px;
  cursor: pointer;
}
.sm-item:hover {
  background: var(--el-fill-color-light);
}
.sm-item.active {
  background: var(--accent-soft, var(--el-fill-color));
}
.sm-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
  background: var(--el-color-success);
}
.sm-dot.book {
  background: var(--el-color-warning);
}
.sm-title {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.sm-scope {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}
.sm-del {
  display: none;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}
.sm-item:hover .sm-del {
  display: inline-flex;
}
.sm-del:hover {
  color: var(--el-color-danger);
}
.sm-empty {
  padding: 18px 8px;
  text-align: center;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.main-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}
.head-info {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.head-title {
  font-weight: 600;
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.msg-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 16px;
}

/* run 分组 + 多选 */
.run-group {
  position: relative;
  border-radius: 8px;
  transition: background .15s;
}
.run-group.selected {
  background: var(--accent-soft, var(--el-fill-color));
  outline: 1px solid var(--accent, var(--el-color-primary));
  outline-offset: -1px;
}
.run-group.in-select {
  padding-left: 28px;
  cursor: pointer;
}
.group-check {
  position: absolute;
  left: -2px;
  top: 8px;
  z-index: 2;
}
/* 任务容器：一次 run 一张卡（过程在上，虚线分隔，回答在下） */
.task-card {
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 2px 12px 12px 12px;
  overflow: hidden;
  min-width: 0;
}
.task-answer {
  padding: 10px 14px;
  font-size: 14px;
  line-height: 1.75;
  color: var(--el-text-color-primary);
}
.task-answer + .task-answer {
  padding-top: 0;
}
/* 回答编辑区 */
.answer-edit :deep(.el-textarea__inner) {
  background: var(--el-bg-color);
  color: var(--el-text-color-primary);
  font-size: 13px;
  line-height: 1.7;
}
.answer-edit-acts {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}
.answer-edit-hint {
  flex: 1;
  min-width: 0;
  font-size: 11px;
  color: var(--el-text-color-secondary);
}
/* 回答编辑区 */
.answer-edit :deep(.el-textarea__inner) {
  background: var(--el-bg-color);
  color: var(--el-text-color-primary);
  font-size: 13px;
  line-height: 1.7;
}
.answer-edit-acts {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}
.answer-edit-hint {
  flex: 1;
  min-width: 0;
  font-size: 11px;
  color: var(--el-text-color-secondary);
}
.task-alert {
  padding: 10px 14px 0;
}
.cot-head + .cot-body {
  border-top: 1px solid var(--el-border-color-lighter);
}
.task-card .cot-body {
  border-top: 1px solid var(--el-border-color-lighter);
}
/* 过程与回答之间用虚线分隔 */
.task-card .cot-body + .task-answer,
.task-card .cot-head + .task-answer {
  border-top: 1px dashed var(--el-border-color);
}
.task-card .proposal-card {
  border-radius: 8px;
  margin-top: 8px;
}

/* AB 对话结构：头像 + 气泡 + 气泡下方常显操作 */
.row {
  margin-bottom: 16px;
  display: flex;
  align-items: flex-start;
  gap: 10px;
}
.row-user {
  justify-content: flex-end;
}
.row-assistant {
  justify-content: flex-start;
}
.avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 16px;
}
.ai-avatar {
  background: var(--accent-soft, var(--el-fill-color));
  color: var(--accent, var(--el-color-primary));
}
.ai-avatar.ghost {
  width: 26px;
  height: 26px;
  margin-left: 4px;
  margin-right: 4px;
  font-size: 13px;
  background: transparent;
  color: var(--el-text-color-secondary);
  align-self: flex-start;
  margin-top: 4px;
}
.user-avatar {
  background: var(--el-fill-color);
  color: var(--el-text-color-secondary);
}
.msg-wrap {
  display: flex;
  flex-direction: column;
  min-width: 0;
  max-width: 88%;
  /* ⚠ 助手侧必须显式占满剩余宽度：容器化的任务卡没有天然内容宽度，
     只靠 shrink-to-fit 会与内部卡形成循环依赖，塌缩成 1px 宽（每字一行）。 */
  flex: 1 1 auto;
}
.row-user .msg-wrap {
  align-items: flex-end;
  flex: 0 1 auto;
}
.bubble-user {
  background: var(--accent-soft, var(--el-fill-color));
  border: 1px solid var(--accent, var(--el-color-primary));
  padding: 8px 12px;
  border-radius: 12px 12px 2px 12px;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 14px;
  line-height: 1.6;
  max-width: 100%;
  box-sizing: border-box;
}
/* 气泡下方操作条：常显 */
.msg-actions {
  display: flex;
  gap: 2px;
  margin-top: 5px;
}
.user-actions {
  justify-content: flex-end;
}
.msg-actions .act {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 3px 7px;
  border-radius: 4px;
  cursor: pointer;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  user-select: none;
  transition: background .15s, color .15s;
}
.msg-actions .act .el-icon {
  font-size: 13px;
}
.msg-actions .act:hover {
  background: var(--el-fill-color);
  color: var(--accent, var(--el-color-primary));
}
.msg-actions .act.danger:hover {
  color: var(--el-color-danger);
}
.typing {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

/* 思考动画（run 一开始就出现，不等首个 token / 工具调用） */
.thinking-line {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
.thinking-text {
  animation: thinking-pulse 1.6s ease-in-out infinite;
}
.thinking-dots {
  display: inline-flex;
  align-items: center;
  gap: 3px;
}
.thinking-dots i {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--accent, var(--el-color-primary));
  display: inline-block;
  animation: thinking-dot 1.2s ease-in-out infinite;
}
.thinking-dots i:nth-child(2) {
  animation-delay: .18s;
}
.thinking-dots i:nth-child(3) {
  animation-delay: .36s;
}
.thinking-time {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  opacity: .7;
}
@keyframes thinking-pulse {
  0%, 100% { opacity: .45; }
  50% { opacity: 1; }
}
@keyframes thinking-dot {
  0%, 100% { transform: translateY(0); opacity: .35; }
  50% { transform: translateY(-4px); opacity: 1; }
}

/* 思考链开关 */
.cot-switch {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  border: 1px solid var(--el-border-color);
  border-radius: 999px;
  padding: 3px 10px;
  cursor: pointer;
  user-select: none;
  flex-shrink: 0;
}
.cot-switch:hover {
  color: var(--accent, var(--el-color-primary));
  border-color: var(--accent, var(--el-color-primary));
}
.cot-switch.on {
  color: var(--accent, var(--el-color-primary));
  border-color: var(--accent, var(--el-color-primary));
  background: var(--accent-soft, transparent);
}

/* 思考卡（CoT） */
.cot-card {
  margin: 2px 0 12px 44px;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  overflow: hidden;
  max-width: 82%;
}
.cot-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  cursor: pointer;
  user-select: none;
}
.cot-title {
  font-size: 13px;
  color: var(--el-text-color-primary);
}
.cot-meta {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.cot-toggle {
  margin-left: auto;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  display: inline-flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
}
.cot-head:hover .cot-toggle {
  color: var(--accent, var(--el-color-primary));
}
.cot-body {
  border-top: 1px solid var(--el-border-color-lighter);
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
/* 模型原始思考过程（英文 CoT 等）：等宽、次级色、限高滚动 */
.cot-think {
  font-size: 12px;
  line-height: 1.7;
  color: var(--el-text-color-secondary);
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 240px;
  overflow-y: auto;
  padding-bottom: 8px;
  border-bottom: 1px dashed var(--el-border-color-lighter);
}
.cot-line {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}
.cot-mark {
  width: 15px;
  height: 15px;
  border-radius: 50%;
  font-size: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.cot-mark.done {
  background: var(--el-color-success-light-8, rgba(103, 194, 58, .15));
  color: var(--el-color-success);
}
.cot-mark.err {
  background: var(--el-color-danger-light-8, rgba(245, 108, 108, .15));
  color: var(--el-color-danger);
  font-weight: 500;
}
.cot-mark.spin {
  border: 2px solid var(--accent, var(--el-color-primary));
  border-top-color: transparent;
  animation: cot-spin .8s linear infinite;
}
@keyframes cot-spin {
  to { transform: rotate(360deg); }
}
.cot-text {
  color: var(--el-text-color-primary);
}
.cot-text.error {
  color: var(--el-color-danger);
}
.cot-text.running {
  color: var(--accent, var(--el-color-primary));
}
.cot-time {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

/* 多选操作条 */
.select-bar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  border-top: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-light);
}
.select-count {
  font-size: 13px;
  color: var(--el-text-color-regular);
}

/* 编辑提示条 */
.edit-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--el-color-warning);
  margin-bottom: 6px;
}

/* 输入区：居中悬浮卡片式 composer */
.composer-zone {
  flex-shrink: 0;
  padding: 4px 24px 16px;
}
/* 居中悬浮：输入列收窄居中，两侧留白撑起高级感 */
.composer-col {
  width: min(760px, 100%);
  margin: 0 auto;
}
.composer-card {
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 16px;
  padding: 12px 14px 9px;
  box-shadow: 0 10px 28px -20px rgba(0, 0, 0, 0.55);
  transition: border-color .15s, box-shadow .15s;
}
.composer-card:focus-within {
  border-color: var(--accent, var(--el-color-primary));
  box-shadow: 0 12px 30px -18px rgba(0, 0, 0, 0.6);
}
.composer-input :deep(.el-textarea__inner) {
  background: transparent;
  border: none;
  box-shadow: none !important;
  padding: 2px 4px;
  color: var(--el-text-color-primary);
  font-size: 13px;
  line-height: 1.7;
  min-height: 58px;
}
.composer-controls {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 6px;
}
.ctl-left,
.ctl-right {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.ctl-btn {
  width: 26px;
  height: 26px;
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  background: transparent;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  font-size: 15px;
  line-height: 1;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.ctl-btn:hover {
  color: var(--accent, var(--el-color-primary));
  border-color: var(--accent, var(--el-color-primary));
}
.model-inline {
  width: 180px;
}
.model-inline :deep(.el-select__wrapper),
.model-inline :deep(.el-input__wrapper) {
  background: transparent;
  box-shadow: 0 0 0 1px var(--el-border-color) !important;
  min-height: 26px;
}
.scope-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--el-border-color);
  border-radius: 999px;
  padding: 3px 10px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: default;
  flex-shrink: 1;
}
.scope-pill .dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}
.scope-pill.book .dot {
  background: var(--el-color-warning);
}
.scope-pill.book {
  color: var(--el-color-warning);
  border-color: var(--el-color-warning-light-5);
}
.scope-pill.global .dot {
  background: var(--el-color-success);
}
.send-btn {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  border: none;
  background: var(--el-color-primary);
  color: #fff;
  font-size: 15px;
  line-height: 1;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: opacity .15s, transform .15s;
}
.send-btn:hover:not(:disabled) {
  opacity: .85;
}
.send-btn:disabled {
  opacity: .35;
  cursor: not-allowed;
}
.send-btn.stop {
  background: var(--el-color-warning);
}
.menu-item {
  padding: 6px 10px;
  font-size: 13px;
  border-radius: 6px;
  cursor: pointer;
  color: var(--el-text-color-primary);
}
.menu-item:hover {
  background: var(--el-fill-color);
}
.menu-item.danger {
  color: var(--el-color-danger);
}

/* 写入提案卡片 */
.proposal-card {
  border: 1px solid var(--el-color-warning-light-5);
  border-radius: 10px;
  overflow: hidden;
  font-size: 13px;
}
.proposal-card.applied {
  border-color: var(--el-color-success-light-5);
}
.proposal-card.rejected,
.proposal-card.rolled_back {
  border-color: var(--el-border-color-lighter);
  opacity: .75;
}
.proposal-card.apply_failed {
  border-color: var(--el-color-danger-light-5);
}
.proposal-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--el-fill-color-light);
}
.proposal-tag {
  font-size: 11px;
  border-radius: 4px;
  padding: 2px 8px;
  flex-shrink: 0;
  background: var(--el-color-warning-light-8, rgba(230, 162, 60, .15));
  color: var(--el-color-warning);
}
.proposal-tag.applied {
  background: var(--el-color-success-light-8, rgba(103, 194, 58, .15));
  color: var(--el-color-success);
}
.proposal-tag.rejected,
.proposal-tag.rolled_back {
  background: var(--el-fill-color);
  color: var(--el-text-color-secondary);
}
.proposal-tag.apply_failed {
  background: var(--el-color-danger-light-8, rgba(245, 108, 108, .15));
  color: var(--el-color-danger);
}
.proposal-summary {
  color: var(--el-text-color-primary);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.proposal-diff {
  padding: 8px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.7;
  border-top: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-lighter, transparent);
}
.diff-line {
  padding: 1px 6px;
  border-radius: 3px;
  word-break: break-all;
  white-space: pre-wrap;
}
.diff-line.del {
  color: var(--el-color-danger);
  background: var(--el-color-danger-light-9, rgba(245, 108, 108, .08));
}
.diff-line.add {
  color: var(--el-color-success);
  background: var(--el-color-success-light-9, rgba(103, 194, 58, .08));
}
.proposal-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-top: 1px solid var(--el-border-color-lighter);
}
.proposal-note {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  flex: 1;
  min-width: 0;
}
.proposal-note.ok {
  color: var(--el-color-success);
}
.proposal-note.err {
  color: var(--el-color-danger);
}

/* 写入权限徽标 */
.perm-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--el-border-color);
  border-radius: 999px;
  padding: 3px 10px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  flex-shrink: 0;
}
.perm-pill:hover {
  color: var(--accent, var(--el-color-primary));
  border-color: var(--accent, var(--el-color-primary));
}
.perm-pill .dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--el-text-color-secondary);
}
.perm-pill.observe .dot {
  background: var(--el-color-info);
}
.perm-pill.confirm .dot {
  background: var(--el-color-warning);
}
.perm-pill.auto .dot {
  background: var(--el-color-success);
}
.perm-menu-title {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
}
.perm-menu-item {
  display: flex;
  flex-direction: column;
  gap: 1px;
  padding: 7px 10px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--el-text-color-primary);
}
.perm-menu-item:hover {
  background: var(--el-fill-color);
}
.perm-menu-item.active {
  background: var(--accent-soft, var(--el-fill-color));
  color: var(--accent, var(--el-color-primary));
}
.perm-menu-item .perm-desc {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

/* 快捷任务 chips */
.quick-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
  padding: 0 2px 10px;
}
.chip {
  border: 1px solid var(--el-border-color);
  border-radius: 999px;
  padding: 4px 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  user-select: none;
  transition: color .15s, border-color .15s;
}
.chip:hover {
  color: var(--accent, var(--el-color-primary));
  border-color: var(--accent, var(--el-color-primary));
}
@keyframes blink {
  50% { opacity: 0; }
}

/* 工具卡片 */
.tool-card {
  width: min(640px, 100%);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  overflow: hidden;
  font-size: 13px;
}
.tool-card.error {
  border-color: var(--el-color-danger-light-5);
}
.tool-head {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 10px;
  background: var(--el-fill-color-light);
  cursor: pointer;
  user-select: none;
}
.tool-icon {
  color: var(--el-text-color-secondary);
}
.tool-name {
  font-weight: 600;
  flex-shrink: 0;
}
.tool-brief {
  color: var(--el-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}
.tool-state .ok { color: var(--el-color-success); }
.tool-state .err { color: var(--el-color-danger); }
.tool-body {
  border-top: 1px solid var(--el-border-color-lighter);
  padding: 8px 10px;
}
.tool-section + .tool-section {
  margin-top: 8px;
}
.tool-label {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin-bottom: 4px;
}
.tool-pre {
  margin: 0;
  padding: 8px;
  background: var(--el-fill-color-light);
  border-radius: 6px;
  font-size: 12px;
  line-height: 1.5;
  max-height: 260px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'Consolas', 'Monaco', monospace;
}
.tool-pre.err {
  color: var(--el-color-danger);
}

/* 空态 */
.empty-composer {
  margin: 18px auto 10px;
}
.empty-hint {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}
.empty-new {
  margin-top: 2px;
}
.agent-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--el-text-color-secondary);
}
.agent-empty h3 {
  margin: 0;
  color: var(--el-text-color-primary);
}
.agent-empty p {
  margin: 0 0 12px;
  font-size: 13px;
}
.empty-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: block;
  margin-bottom: 4px;
}

.form-hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
  line-height: 1.4;
}
</style>

<style>
/* 暗色主题适配：dark-theme.css 没有重定义 EP 的 fill/bg/text/border 基础变量，
   Agent 页面全靠这批变量上色。在页面根上按主题重定义（CSS 变量随子树继承，
   EP 组件与自定义元素一起生效），并压住写死的 light-X 档位。 */
:root[data-theme='dark'] .agent-page {
  --el-bg-color: #161b22;
  --el-fill-color-blank: #1c2333;
  --el-fill-color-light: #1c2333;
  --el-fill-color: #192031;
  --el-fill-color-lighter: #161d2b;
  --el-border-color: #30363d;
  --el-border-color-light: #262e3c;
  --el-border-color-lighter: #222a38;
  --el-text-color-primary: #e6edf3;
  --el-text-color-regular: #c9d1d9;
  --el-text-color-secondary: #8b98a8;
  --el-text-color-placeholder: #4d5768;
  --el-color-warning-light-5: rgba(230, 162, 60, .35);
  --el-color-warning-light-8: rgba(230, 162, 60, .15);
  --el-color-warning-light-9: rgba(230, 162, 60, .08);
  --el-color-danger-light-5: rgba(245, 108, 108, .35);
  --el-color-danger-light-8: rgba(245, 108, 108, .15);
  --el-color-danger-light-9: rgba(245, 108, 108, .08);
  --el-color-success-light-5: rgba(103, 194, 58, .35);
  --el-color-success-light-8: rgba(103, 194, 58, .15);
  --el-color-success-light-9: rgba(103, 194, 58, .08);
  --el-color-info-light-8: rgba(144, 147, 153, .15);
}
/* EP 把按钮变量定义在 .el-button 自身作用域（--el-button-bg-color: var(--el-color-white)），
   从页面级改变量盖不住（EP 内部还有一层），直接落属性 + !important（与 dark-theme.css 同套路）。
   ⚠ 只能按颜色类型排除，不能用 :not([class*="el-button--"])——size 修饰符 el-button--small 也带这个前缀。 */
:root[data-theme='dark'] .agent-page .el-button:not(.el-button--primary):not(.el-button--danger):not(.el-button--success):not(.el-button--warning):not(.el-button--info) {
  background-color: #1c2333 !important;
  border-color: #30363d !important;
  color: #e6edf3 !important;
}
:root[data-theme='dark'] .agent-page .el-button:not(.el-button--primary):not(.el-button--danger):not(.el-button--success):not(.el-button--warning):not(.el-button--info):hover {
  background-color: #223047 !important;
  border-color: #3b82f6 !important;
  color: #58a6ff !important;
}

/* 暗色下实心主色按钮压回 #3b82f6（accent 亮版 #58a6ff 配白字只有 2.5:1） */
:root[data-theme='dark'] .agent-page .send-btn {
  background: #3b82f6;
}
:root[data-theme='dark'] .agent-page .send-btn.stop {
  background: #d78a12;
}

/* 历史对话下拉（teleport 到 body，不继承 .agent-page 变量） */
.session-menu-pop {
  padding: 10px 10px 8px !important;
}
:root[data-theme='dark'] .session-menu-pop {
  /* ⚠ dark-theme.css 有一条优先级更高的 !important：
     :root[data-theme=dark] .el-popper.is-light { background: var(--el-bg-color-overlay) !important }
     而 --el-bg-color-overlay 在暗色下没被重定义（仍是 #fff）→ 自定义弹层全变白底。
     直接在被覆盖元素上重定义它引用的变量，比打优先级战稳。 */
  --el-bg-color-overlay: #161b22;
  --el-popover-bg-color: #161b22;
  --el-text-color-regular: #e6edf3;
  --el-text-color-primary: #e6edf3;
  --el-text-color-secondary: #8b98a8;
  --el-border-color-lighter: #262e3c;
  /* dark-theme.css 的 !important 规则里 border 用的是 --el-border-color-light，
     它同样没被重定义（仍是浅色）→ 一并在被覆盖元素上重定义 */
  --el-border-color-light: #30363d;
  --el-popover-border-color: #30363d;
  --el-popper-border-color: #30363d;
  background: #161b22 !important;
  border: 1px solid #30363d !important;
}
:root[data-theme='dark'] .session-menu-pop .sm-head {
  border-bottom-color: #262e3c;
}
:root[data-theme='dark'] .session-menu-pop .sm-head-title,
:root[data-theme='dark'] .session-menu-pop .sm-title {
  color: #e6edf3;
}
:root[data-theme='dark'] .session-menu-pop .sm-count,
:root[data-theme='dark'] .session-menu-pop .sm-scope,
:root[data-theme='dark'] .session-menu-pop .sm-del,
:root[data-theme='dark'] .session-menu-pop .sm-empty {
  color: #8b98a8;
}:root[data-theme='dark'] .session-menu-pop .sm-count {
  background: #1c2333;
}
:root[data-theme='dark'] .session-menu-pop .sm-item:hover {
  background: #1c2333;
}
:root[data-theme='dark'] .session-menu-pop .sm-item.active {
  background: rgba(88, 166, 255, 0.12);
}

/* 传送到 body 的弹出层（+ 菜单 / 权限菜单）不继承 .agent-page 变量，单独覆盖 */
:root[data-theme='dark'] .menu-item,
:root[data-theme='dark'] .perm-menu-item {
  color: #e6edf3;
}
:root[data-theme='dark'] .perm-menu-item .perm-desc {
  color: #8b98a8;
}
:root[data-theme='dark'] .perm-menu-item.active {
  color: #58a6ff;
  background: rgba(88, 166, 255, 0.12);
}
:root[data-theme='dark'] .menu-item:hover,
:root[data-theme='dark'] .perm-menu-item:hover {
  background: #1c2333;
}
</style>
