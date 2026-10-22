<template>
  <div ref="pageRef" class="tb-page" v-loading="loading">

    <!-- 顶端栏：返回（左侧）。sticky 吸附，侧栏 sticky 偏移让出它的高度 -->
    <header class="tb-topbar">
      <button class="tb-back" type="button" @click="goBack">← 返回</button>

      <!-- 右上角：正文字体切换（本机可见的偏好，存 localStorage） -->
      <el-dropdown class="tb-font" trigger="click" @command="applyPostFont">
        <button class="tb-font-btn" type="button" title="切换正文字体">
          <span class="tb-font-label">字体</span>
          <span class="tb-font-cur">{{ fontLabel }}</span>
          <span class="tb-font-caret">▾</span>
        </button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item v-for="f in FONT_OPTIONS" :key="f.label" :command="f.label">
              <span class="tb-font-item" :class="{ on: f.label === fontLabel }">
                {{ f.label }}<span class="tb-font-demo" :style="{ fontFamily: f.stack }">永和九年</span>
              </span>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </header>

    <div v-if="experienceShare" class="tb-wrap" :class="{ 'rail-mode': railCollapsed }">
      <!-- 左栏：文章目录（可折叠成图标条） -->
      <aside class="tb-left" :class="{ collapsed: railCollapsed }">
        <!-- 收起态：图标条 -->
        <div v-if="railCollapsed" class="rail-bar">
          <button class="rail-btn" type="button" title="展开侧栏" @click="toggleRail">»</button>
        </div>

        <template v-else>
        <div class="nav-card toc-card">
          <div class="toc-head">
            <span class="toc-head-t">☰ 文章目录</span>
            <button class="toc-collapse" type="button" title="收起侧栏" @click="toggleRail">◧</button>
          </div>
          <div class="toc-mini">
            <a
              v-for="item in tocItems"
              :key="item.id"
              :class="{ lv2: item.level === 2, on: activeTocId === item.id }"
              :href="`#${item.id}`"
              @click.prevent="scrollToHeading(item.id)"
            >{{ item.num ? `${item.num}、` : '' }}{{ item.text }}</a>
            <a v-if="experienceShare.pdf_file_url" class="lv2" href="#att" @click.prevent="scrollToId('att')">附件</a>
            <a v-if="recommendedShares.length" class="lv2" href="#rel" @click.prevent="scrollToId('rel')">相关条目</a>
            <a class="lv2" href="#bbs-reply-box" @click.prevent="scrollToId('bbs-reply-box')">讨论区</a>
          </div>
          <div class="toc-foot">{{ tocChapterCount }} 个章节</div>
        </div>
        </template>
      </aside>

      <!-- 中栏：主帖 + 楼层 -->
      <main class="tb-mid">
        <!-- 主帖卡 -->
        <div class="post-card">

          <!-- 题图（非 PDF 导入时展示封面） -->
          <figure v-if="hasCover && !shouldShowPdfBody" class="cover-fig">
            <img :src="proxyImgSrc(experienceShare.cover_url)" :alt="experienceShare.title" />
          </figure>

          <h1 class="post-title">{{ experienceShare.title }}</h1>

          <!-- 作者信息 / 时间 / 点赞数（参考版式：标题之下、正文之上） -->
          <div class="post-meta">
            <div class="pm-left">
              <div class="pm-avatar" :style="avatarStyle(experienceShare.author_name || '星芒用户')">
                {{ (experienceShare.author_name || '星')[0] }}
              </div>
              <span class="pm-name">{{ experienceShare.author_name || '星芒用户' }}</span>
              <span class="pm-tag" :class="{ pdf: experienceShare.create_type === 'pdf_import' }">
                {{ experienceShare.create_type === 'pdf_import' ? 'PDF 导入' : '手动创建' }}
              </span>
              <span v-if="experienceShare.version" class="pm-tag">v{{ experienceShare.version }}</span>
              <span v-if="experienceShare.is_featured === 1" class="pm-tag feat">精华</span>
            </div>

            <div class="pm-right">
              <span class="pm-time">{{ formatDateTime(experienceShare.created_at) }}</span>
              <button
                class="pm-like"
                :class="{ liked: mainLiked }"
                type="button"
                :title="mainLiked ? '取消点赞' : '点赞'"
                @click="toggleMainLike"
              >{{ mainLiked ? '♥' : '♡' }} {{ mainLikeCount }} 点赞</button>
            </div>
          </div>

          <p v-if="experienceShare.summary" class="post-summary">{{ experienceShare.summary }}</p>


          <div ref="articleRef">
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

            <!-- 补充说明（PDF 导入时） -->
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
                <span class="ico-file"></span>
                <span class="nm">{{ experienceShare.pdf_file_name || 'PDF附件' }}</span>
                <span class="sz">{{ formatFileSize(experienceShare.pdf_file_size) }}</span>
                <span class="btns">
                  <a @click="openPdfPreview(experienceShare.pdf_file_url, experienceShare.pdf_file_name || 'PDF附件')">预览</a>
                  <a @click="downloadPdf(experienceShare.pdf_file_url, experienceShare.pdf_file_name || 'experience-share.pdf')">下载</a>
                </span>
              </div>
            </div>
          </div>

          <!-- 正文下方的右下角：保存状态 + 编辑信息 -->
          <div class="post-foot">
            <span v-if="saveStatus" class="pa-meta">{{ saveStatus }}</span>
            <button class="pf-edit" type="button" @click="openMetaComposer()">编辑信息</button>
          </div>
        </div>

        <!-- 回复区（含历史 / 日志） -->
        <section class="reply-card">
          <div class="rc-head">
            <h3>全部回复</h3>
            <span class="cnt">{{ bbsReplies.length }}</span>
            <div class="sort-tabs">
              <button :class="{ on: activeTab === 'article' }" @click="switchTab('article')">回复</button>
              <button :class="{ on: activeTab === 'history' }" @click="switchTab('history')">历史版本</button>
              <button :class="{ on: activeTab === 'log' }" @click="switchTab('log')">操作日志</button>
            </div>
          </div>

          <!-- 楼层 -->
          <div v-if="activeTab === 'article'">
            <div v-if="!bbsReplies.length" class="bbs-empty">
              <span>暂无回复，快来抢沙发</span>
            </div>

            <div v-for="(reply, index) in topReplies" :key="reply.id" class="floor">
              <div class="floor-av" :style="avatarStyle(reply.nickname)">{{ reply.nickname.charAt(0) }}</div>

              <div class="floor-main">
                <div class="floor-head">
                  <span class="floor-name">{{ reply.nickname }}</span>
                  <span v-if="reply.mine" class="floor-badge">我</span>
                  <span class="floor-when">{{ reply.createdAt }}　#{{ index + 1 }}</span>
                </div>

                <template v-if="editingId === reply.id">
                  <textarea v-model="editingText" class="edit-textarea" rows="3" placeholder="编辑回复内容…"></textarea>
                </template>
                <div v-else class="floor-content" v-html="renderContent(reply.content)"></div>

                <!-- 楼中楼：二级封顶，回复子楼只挂 @昵称、不再往下缩进 -->
                <div v-if="childrenOf(reply.id).length" class="sub-list">
                  <div v-for="child in childrenOf(reply.id)" :key="child.id" class="sub-item">
                    <div class="sub-head">
                      <span class="sub-name">{{ child.nickname }}</span>
                      <span v-if="child.replyToNickname" class="sub-at">回复 @{{ child.replyToNickname }}</span>
                      <span class="sub-when">{{ child.createdAt }}</span>
                    </div>

                    <template v-if="editingId === child.id">
                      <textarea v-model="editingText" class="edit-textarea" rows="2" placeholder="编辑回复内容…"></textarea>
                      <div class="sub-acts">
                        <button class="save" @click="saveEdit(child)">保存</button>
                        <button @click="cancelEdit">取消</button>
                      </div>
                    </template>
                    <template v-else>
                      <div class="sub-content" v-html="renderContent(child.content)"></div>
                      <div class="sub-acts">
                        <button @click="replyToFloor(reply.id, child.nickname)">回复</button>
                        <button :class="{ liked: child.liked }" @click="toggleLike(child)">♡ {{ child.likes }}</button>
                        <button v-if="child.mine" @click="startEdit(child)">编辑</button>
                        <button v-if="child.mine" class="del" @click="deleteReply(child)">删除</button>
                      </div>
                    </template>
                  </div>
                </div>

                <div class="floor-actions">
                  <template v-if="editingId === reply.id">
                    <button class="save" @click="saveEdit(reply)">保存</button>
                    <button @click="cancelEdit">取消</button>
                  </template>
                  <template v-else>
                    <button @click="replyToFloor(reply.id, reply.nickname)">回复</button>
                    <button :class="{ liked: reply.liked }" @click="toggleLike(reply)">♡ {{ reply.likes }}</button>
                    <button v-if="reply.mine" @click="startEdit(reply)">编辑</button>
                    <button v-if="reply.mine" class="del" @click="deleteReply(reply)">删除</button>
                  </template>
                </div>
              </div>
            </div>

            <div id="bbs-reply-box" class="reply-box">
              <div v-if="replyTarget" class="reply-quote">
                <span>回复 @{{ replyTarget.nickname }}（会挂在该楼中楼里）</span>
                <button class="x" @click="clearReplyTarget">×</button>
              </div>
              <textarea
                v-model="replyText"
                placeholder="发表你的看法…（Ctrl+Enter 发送）"
                @keydown.ctrl.enter.prevent="submitReply"
              ></textarea>
              <div class="rb-bar">
                <span class="nick-label">昵称：<input v-model="nickname" maxlength="12" placeholder="输入昵称" /></span>
                <div class="emoticons">
                  <button v-for="e in emoticons" :key="e.code" :title="e.code" @click="insertEmoticon(e.code)">{{ e.emoji }}</button>
                </div>
                <span class="hint">本地存储 · 仅本机可见</span>
                <button class="submit" @click="submitReply">发表回复</button>
              </div>
            </div>
          </div>

          <!-- 历史：版本历史 -->
          <div v-else-if="activeTab === 'history'" class="tab-panel">
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
          <div v-else class="tab-panel">
            <h2>操作日志</h2>
            <div class="log-list">
              <div v-for="(item, i) in logItems" :key="i" class="log-item">
                <span class="log-time">{{ item.time }}</span>
                <span class="log-action">{{ item.action }}</span>
                <span class="log-user">{{ item.user }}</span>
              </div>
            </div>
          </div>
        </section>
      </main>

      <!-- 右栏：相关推荐 -->
      <aside class="tb-right">
        <div v-if="recommendedShares.length" id="rel" class="side-card">
          <div class="sc-h">相关推荐 <span class="more">共 {{ recommendedShares.length }} 条</span></div>
          <a
            v-for="item in recommendedShares"
            :key="item.id"
            class="rel-item"
            @click="goToRecommended(item.id)"
          >
            <div class="ri-t">{{ item.title || '未命名' }}</div>
            <div v-if="relExcerpt(item)" class="ri-x">{{ relExcerpt(item) }}</div>
            <div v-if="proxyImgSrc(item.cover_url)" class="ri-img">
              <img :src="proxyImgSrc(item.cover_url)" :alt="item.title" loading="lazy" />
            </div>
            <div class="ri-foot">
              <span class="ri-chip">
                <i :style="avatarStyle(item.author_name || '星芒用户')"></i>
                <span>{{ sectionNameOf(item.section_id) || '未归类' }}</span>
              </span>
              <span class="ri-nums">{{ relReplyCount(item.id) }}评论 {{ relLikeTotal(item) }}点赞</span>
            </div>
          </a>
        </div>
      </aside>
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

    <!-- 全屏编辑弹窗（与列表页「发帖」同款形态：大一点的全屏弹窗，去弹窗壳） -->
    <teleport to="body">
      <div v-if="metaEditing" class="composer">
        <header class="cp-topbar">
          <button class="cp-back" type="button" @click="cancelMetaEditing">← 返回</button>
          <div class="cp-title">编辑帖子 · {{ boardConf.name }}</div>
          <button class="cp-publish" type="button" :disabled="metaSaving" @click="saveMetaComposer">
            {{ metaSaving ? '提交中…' : '保存修改' }}
          </button>
        </header>

        <div class="cp-steps">
          <span class="cp-step" :class="{ on: metaStep === 1 }">① 基本信息</span>
          <span class="cp-step-line"></span>
          <span class="cp-step" :class="{ on: metaStep === 2 }">② 补充信息（可选）</span>
        </div>

        <main class="cp-body">
          <template v-if="metaStep === 1">
            <div class="cp-row">
              <span class="cp-label">发布到吧</span>
              <span class="cp-bar"><i class="cp-bar-av">{{ boardConf.short }}</i>{{ boardConf.name }}</span>
            </div>
            <div class="cp-row">
              <span class="cp-label">版块分区</span>
              <div class="sec-chips">
                <button type="button" class="sec-chip" :class="{ on: !metaForm.section_id }" @click="metaForm.section_id = null">未归类</button>
                <button
                  v-for="s in sections"
                  :key="s.id"
                  type="button"
                  class="sec-chip"
                  :class="{ on: metaForm.section_id === s.id }"
                  @click="metaForm.section_id = metaForm.section_id === s.id ? null : s.id"
                >{{ s.name }}</button>
              </div>
            </div>
            <el-input
              v-model="metaForm.title"
              maxlength="120"
              show-word-limit
              placeholder="必填，请输入帖子标题（5-31 个字）"
              class="cp-title-input"
            />
            <TipTapEditor
              v-model="metaForm.content"
              placeholder="请输入正文…"
              class="cp-editor"
            />
          </template>

          <template v-else>
            <div class="cp-row top">
              <span class="cp-label">简介</span>
              <el-input
                v-model="metaForm.summary"
                type="textarea"
                :rows="3"
                maxlength="200"
                show-word-limit
                placeholder="选填，建议一句话概括核心经验，列表页会优先展示这里的内容"
                class="cp-summary-input"
              />
            </div>

            <div class="cp-row top">
              <span class="cp-label">封面图</span>
              <div class="cover-field">
                <div v-if="metaCoverSrc" class="cover-preview-card">
                  <img :src="metaCoverSrc" alt="封面预览" class="cover-preview-image" />
                  <div class="cover-preview-actions">
                    <el-button type="primary" plain size="small" @click="openMetaCoverPicker">替换图片</el-button>
                    <el-button type="danger" plain size="small" @click="removeMetaCover">移除</el-button>
                  </div>
                </div>
                <div v-else class="cover-upload-area" @click="openMetaCoverPicker">
                  <el-icon class="cover-upload-icon"><Plus /></el-icon>
                  <span class="cover-upload-text">点击上传封面图</span>
                  <span class="cover-upload-hint">支持 jpg、png、gif 格式，建议尺寸 16:9</span>
                </div>
                <div class="cover-url-input">
                  <span class="cover-url-label">或输入图片URL：</span>
                  <el-input v-model="metaForm.cover_url" placeholder="https://example.com/cover.jpg" clearable @input="handleMetaCoverUrlInput" />
                </div>
              </div>
            </div>

            <div class="cp-row">
              <span class="cp-label">作者</span>
              <el-input v-model="metaForm.author_name" maxlength="50" placeholder="作者名称" class="cp-inline-input" />
            </div>

            <div class="cp-row">
              <span class="cp-label">版本号</span>
              <el-input v-model="metaForm.version" maxlength="30" placeholder="如 v1.0.0" class="cp-inline-input" />
            </div>

            <div class="cp-row">
              <span class="cp-label">内容声明</span>
              <div class="declare-row">
                <el-checkbox v-model="metaForm.is_featured">设为精华</el-checkbox>
                <span class="declare-hint">精华与分区互不冲突，精华帖会在「精华」入口下出现</span>
              </div>
            </div>
          </template>
        </main>

        <footer class="cp-footer">
          <template v-if="metaStep === 1">
            <button class="cp-skip" type="button" :disabled="metaSaving" @click="saveMetaComposer">直接保存</button>
            <button class="cp-next" type="button" @click="metaStep = 2">下一步：补充信息（可选）→</button>
          </template>
          <template v-else>
            <button class="cp-skip" type="button" @click="metaStep = 1">← 上一步</button>
            <button class="cp-next" type="button" :disabled="metaSaving" @click="saveMetaComposer">
              {{ metaSaving ? '提交中…' : '保存修改' }}
            </button>
          </template>
        </footer>

        <input
          ref="metaCoverInputRef"
          type="file"
          accept="image/*"
          style="display: none"
          @change="handleMetaCoverChange"
        />
      </div>
    </teleport>

  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { experienceShareAPI, experienceSectionAPI } from '@/api'
import type { ExperienceShare, ExperienceSection, BoardKind } from '@/types'
import { BOARD_CONF, normalizeBoard, tutorialBoardPath, tutorialPostPath } from '@/constants/boards'
import { Plus } from '@element-plus/icons-vue'
import TipTapEditor from '@/components/TipTapEditor.vue'

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

/* 详情页归属哪个吧：以帖子自己带的 board 为准（不依赖 referer / 路由），
   这样从分享链接直接打开也能显示正确的吧名、返回按钮也回对的吧。 */
const detailBoard = computed<BoardKind>(() => normalizeBoard(experienceShare.value?.board))
const boardConf = computed(() => BOARD_CONF[detailBoard.value])
const pdfPreviewVisible = ref(false)
const pdfPreviewUrl = ref('')
const pdfPreviewFileName = ref('')
const pdfPreviewDownloadName = ref('')
const showBackToTop = ref(false)
const saveStatus = ref('')
const savingContent = ref(false)
let saveStatusTimer: ReturnType<typeof setTimeout> | null = null

/* ---------- 沉浸式编辑页（与发帖页同款：整页覆盖，去弹窗） ---------- */
const metaEditing = ref(false)
const metaSaving = ref(false)
const metaStep = ref<1 | 2>(1)
const metaForm = ref({
  title: '',
  summary: '',
  content: '',
  author_name: '',
  version: '',
  cover_url: '',
  section_id: null as number | null,
  is_featured: false
})

/* 封面本地上传（与列表页发帖同一套：base64 走 cover_image 落盘） */
const META_MAX_COVER_SIZE = 5 * 1024 * 1024
const metaCoverInputRef = ref<HTMLInputElement | null>(null)
const metaLocalCover = ref<{ name: string; size: number; data_base64: string; objectUrl: string } | null>(null)
const metaRemoveCover = ref(false)

const revokeMetaLocalCover = () => {
  if (metaLocalCover.value?.objectUrl) URL.revokeObjectURL(metaLocalCover.value.objectUrl)
  metaLocalCover.value = null
}

const openMetaCoverPicker = () => metaCoverInputRef.value?.click()

const handleMetaCoverChange = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  try {
    if (!file) return
    if (!file.type.startsWith('image/')) throw new Error('请选择图片文件')
    if (file.size > META_MAX_COVER_SIZE) throw new Error('图片大小不能超过 5MB')

    revokeMetaLocalCover()
    const data_base64 = await new Promise<string>((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => resolve((reader.result as string).split(',')[1])
      reader.onerror = () => reject(new Error('读取图片失败'))
      reader.readAsDataURL(file)
    })
    metaLocalCover.value = {
      name: file.name,
      size: file.size,
      data_base64,
      objectUrl: URL.createObjectURL(file)
    }
    metaForm.value.cover_url = ''
    metaRemoveCover.value = false
  } catch (error: any) {
    ElMessage.error(error.message || '上传图片失败')
  } finally {
    input.value = ''
  }
}

const handleMetaCoverUrlInput = () => {
  if (metaForm.value.cover_url && metaLocalCover.value) revokeMetaLocalCover()
}

const removeMetaCover = () => {
  revokeMetaLocalCover()
  if (metaForm.value.cover_url) metaRemoveCover.value = true
  metaForm.value.cover_url = ''
}

/* 封面预览走同源代理（外链图床防盗链 403）；本地盘路径判不可用。
   ⚠ 只影响预览显示，metaForm.cover_url 保存的仍是原值。 */
const metaCoverSrc = computed(() => {
  if (metaLocalCover.value) return metaLocalCover.value.objectUrl
  const u = (metaForm.value.cover_url || '').trim()
  if (!u) return ''
  if (/^(\/|data:|blob:)/i.test(u)) return u
  if (/^https?:\/\//i.test(u)) return `/api/image-proxy?url=${encodeURIComponent(u)}`
  return ''
})

const openMetaComposer = () => {
  const s = experienceShare.value
  if (!s) return
  revokeMetaLocalCover()
  metaRemoveCover.value = false
  metaStep.value = 1
  metaForm.value = {
    title: s.title || '',
    summary: s.summary || '',
    content: s.content || '',
    author_name: s.author_name || '',
    version: s.version || '',
    cover_url: s.cover_url || '',
    section_id: s.section_id ?? null,
    is_featured: s.is_featured === 1
  }
  metaEditing.value = true
}

const cancelMetaEditing = () => {
  metaEditing.value = false
  revokeMetaLocalCover()
}

const saveMetaComposer = async () => {
  const s = experienceShare.value
  if (!s) return
  if (!metaForm.value.title.trim()) {
    ElMessage.warning('标题不能为空')
    return
  }

  metaSaving.value = true
  try {
    const payload: any = {
      title: metaForm.value.title.trim(),
      summary: metaForm.value.summary.trim(),
      content: metaForm.value.content,
      content_render_mode: 'html',
      cover_url: metaForm.value.cover_url.trim(),
      create_type: s.create_type,
      author_id: s.author_id ?? null,
      author_name: metaForm.value.author_name.trim(),
      version: metaForm.value.version.trim(),
      status: s.status || 'published',
      section_id: metaForm.value.section_id,
      is_featured: metaForm.value.is_featured ? 1 : 0,
      pdf_file_url: s.pdf_file_url || null,
      pdf_file_name: s.pdf_file_name || null,
      pdf_file_size: s.pdf_file_size || 0,
      pdf_parse_status: s.pdf_parse_status || null,
      pdf_parse_result: s.pdf_parse_result || null,
      source_file_name: s.source_file_name || null,
      remove_cover: metaRemoveCover.value
    }

    if (metaLocalCover.value) {
      payload.cover_image = {
        name: metaLocalCover.value.name,
        size: metaLocalCover.value.size,
        data_base64: metaLocalCover.value.data_base64
      }
    }

    const res = await experienceShareAPI.update(s.id, payload)
    if (res.success && res.data) {
      experienceShare.value = res.data
      if (experienceShare.value) {
        experienceShare.value.content = rewriteContentImgs(experienceShare.value.content || '')
      }
      revokeMetaLocalCover()
      ElMessage.success('已保存')
      metaEditing.value = false
    }
  } finally {
    metaSaving.value = false
  }
}

const pageRef = ref<HTMLElement | null>(null)
const articleRef = ref<HTMLElement | null>(null)

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

/* ---------- 目录：从正文 HTML 解析 ---------- */
// 目录条目 = 正文里的「小节标记」，有两种：
//   ① 标题元素 h1 / h2 / h3（工具栏下拉「标题 1 / 标题 2」产生的）
//   ② **整段加粗的段落** —— 把一行字整体加粗（工具栏 B 按钮）也算一个小节。
//      加粗只产生 <strong> 不产生标题结构，所以必须单独识别，否则用户靠加粗
//      标出的小标题在目录里永远看不到。
// 数据源直接用 content 而不是编辑器 DOM：不依赖渲染时机，编辑时随 v-model 实时同步。
const TOC_TAGS = 'h1, h2, h3, p'

// 整段加粗：段落里的文字必须全部落在 <strong> 内（strong 里还可以嵌 em/u 等）
const isBoldParagraph = (el: Element): boolean => {
  if (el.tagName !== 'P') return false
  const text = (el.textContent || '').trim()
  if (!text) return false
  const bolded = Array.from(el.querySelectorAll('strong'))
    .map(s => s.textContent || '')
    .join('')
    .trim()
  return bolded === text
}

const isTocTarget = (el: Element): boolean => {
  const tag = el.tagName
  return tag === 'H1' || tag === 'H2' || tag === 'H3' || isBoldParagraph(el)
}

// 编辑器 DOM 侧的候选元素：顺序与 content 解析一致（都是文档顺序）
const collectTocTargets = (root: ParentNode): HTMLElement[] =>
  Array.from(root.querySelectorAll(TOC_TAGS)).filter(isTocTarget) as HTMLElement[]

const tocItems = computed<TocItem[]>(() => {
  const html = experienceShare.value?.content
  if (!html) return []
  const doc = new DOMParser().parseFromString(html, 'text/html')
  let topCount = 0
  return Array.from(doc.querySelectorAll(TOC_TAGS))
    .filter(isTocTarget)
    .map((el, index): TocItem => {
      // h1 / h2 与「整段加粗」算一级小节（带序号），h3 作为二级
      const isTop = el.tagName !== 'H3'
      if (isTop) topCount += 1
      return {
        id: `wiki-sec-${index}`,
        index,
        text: el.textContent?.trim() || '',
        level: isTop ? 1 : 2,
        num: isTop ? topCount : null
      }
    })
})

const scrollToHeading = (id: string) => {
  const item = tocItems.value.find(t => t.id === id)
  if (!item) return
  const scroll = () => {
    const root = articleRef.value
    if (!root) return
    const targets = collectTocTargets(root)
    // 优先按索引定位；编辑器规范化导致数量对不上时，退化为按文字匹配
    let target = targets[item.index]
    if (!target || (target.textContent || '').trim() !== item.text) {
      target = targets.find(el => (el.textContent || '').trim() === item.text) as HTMLElement
    }
    target?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
  scroll()
  // 首次点击可能早于编辑器完成渲染，下一帧补一次
  nextTick(scroll)
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
    if (res.success) {
      // update 接口只回显请求体，created_at / updated_at 等字段是空的，
      // 直接拿它整体覆盖会把「创建时间」等展示信息清掉，所以保存后重拉一次完整记录
      const fresh = await experienceShareAPI.getOne(s.id)
      const data = fresh.success && fresh.data ? fresh.data : { ...s, ...res.data }
      // 防御：万一重拉回来的正文是空的（保存竞态 / 接口异常），不要用空值把编辑器里的正文清掉
      if (!data.content && s.content) data.content = s.content
      experienceShare.value = data
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

/* ---------- 回复区 Tab（回复 / 历史 / 日志）----------
   贴吧版式下主帖与楼层同页展示，不再有独立的「讨论」页，
   所以 tab 从「条目/讨论/历史/日志」收缩为回复区内的三个视图。 */
type WikiTab = 'article' | 'history' | 'log'

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
  likes: number
  liked: boolean
  mine: boolean
  createdAt: string
  /* 楼中楼：parentId 指向所属的顶层楼层；replyToNickname 是这一条在回谁（二级起只挂 @昵称，不再往下缩进）。
     老数据没有这两个字段 → 一律当顶层楼层，不会丢。 */
  parentId?: number | null
  replyToNickname?: string | null
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
/* 楼中楼回复目标：parentId = 顶层楼层 id，nickname = 被回复的人 */
const replyTarget = ref<{ parentId: number; nickname: string } | null>(null)
const nickname = ref(localStorage.getItem('bbs-nickname') || '本地用户')
const editingId = ref<number | null>(null)
const editingText = ref('')

const bbsStorageKey = computed(() => `bbs-replies-${route.params.id}`)

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

/* 顶层楼层 = 没有 parentId，或 parentId 指向的楼层已不存在（父被删后子回复自动上浮，不会丢） */
const topReplies = computed<BbsReply[]>(() => {
  const ids = new Set(bbsReplies.value.map(r => r.id))
  return bbsReplies.value.filter(r => !r.parentId || !ids.has(r.parentId))
})

const childrenOf = (id: number) => bbsReplies.value.filter(r => r.parentId === id)

const focusReply = () => {
  document.getElementById('bbs-reply-box')?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  nextTick(() => {
    const ta = document.querySelector('#bbs-reply-box textarea') as HTMLTextAreaElement | null
    ta?.focus()
  })
}

/* 「回复」= 楼中楼：parentId 挂到顶层楼层；nickname 是这一条在回谁 */
const replyToFloor = (parentId: number, nickname: string) => {
  replyTarget.value = { parentId, nickname }
  focusReply()
}

const clearReplyTarget = () => {
  replyTarget.value = null
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
    likes: 0,
    liked: false,
    mine: true,
    createdAt: formatNow(),
    parentId: replyTarget.value?.parentId ?? null,
    replyToNickname: replyTarget.value?.nickname ?? null
  })
  replyText.value = ''
  replyTarget.value = null
  saveReplies()
}

const toggleLike = (reply: BbsReply) => {
  reply.liked = !reply.liked
  reply.likes += reply.liked ? 1 : -1
  saveReplies()
}

const deleteReply = (reply: BbsReply) => {
  // 顶层楼层被删时，它的楼中楼一并删除（不留孤儿）
  bbsReplies.value = bbsReplies.value.filter(r => r.id !== reply.id && r.parentId !== reply.id)
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

/* ---------- 分区 / 精华 / 主楼赞 ---------- */
const sections = ref<ExperienceSection[]>([])

const fetchSections = async () => {
  try {
    const res = await experienceSectionAPI.getAll()
    if (res.success && res.data) sections.value = res.data
  } catch {
    sections.value = []
  }
}

/* ---------- 图片代理（与列表页同口径） ----------
   外链（http/https）走 /api/image-proxy：贴吧等图床有 Referer 防盗链，浏览器直连 403；
   站内/ data:/ blob: 直用；本地盘路径（D:\...）浏览器不可达，返回空串由调用方隐藏。 */
const proxyImgSrc = (url?: string | null): string => {
  const u = (url || '').trim()
  if (!u) return ''
  if (/^(\/|data:|blob:)/i.test(u)) return u
  if (/^https?:\/\//i.test(u)) return `/api/image-proxy?url=${encodeURIComponent(u)}`
  return ''
}
const hasCover = computed(() => !!proxyImgSrc(experienceShare.value?.cover_url))

/* 正文富文本里的外链 img 一并转代理；本地盘/空 src 的 img 直接删除（破图比溢出难看且可能撑爆排版）。
   ⚠ 重写后保存的也是代理地址：站内相对路径幂等可再代理，且后端有磁盘缓存兜底（源失效也能出图）。 */
const rewriteContentImgs = (html: string): string =>
  (html || '').replace(/<img[^>]*\ssrc=["']([^"']*)["'][^>]*>/gi, (tag, src: string) => {
    const proxied = proxyImgSrc(src)
    if (!proxied) return ''
    return tag.replace(src, proxied)
  })

/* ---------- 相关推荐卡片（与列表页同一套口径） ---------- */
const readBbsReplies = (id: number): any[] => {
  try {
    const raw = localStorage.getItem(`bbs-replies-${id}`)
    if (!raw) return []
    const arr = JSON.parse(raw)
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

/** 评论数：与详情页楼层共用同一份 localStorage */
const relReplyCount = (id: number) => readBbsReplies(id).length

/** 总赞数：主楼赞（后端 like_count）+ 楼层赞（localStorage），与列表页一致 */
const relLikeTotal = (item: ExperienceShare) =>
  (Number(item.like_count) || 0) +
  readBbsReplies(item.id).reduce((sum, r) => sum + (Number(r?.likes) || 0), 0)

const sectionNameOf = (sid?: number | null) =>
  sid ? sections.value.find(s => s.id === sid)?.name || '' : ''

/** 摘要：优先简介，其次从正文剥标签；超长截断 */
const relExcerpt = (item: ExperienceShare) => {
  const text =
    (item.summary || '').trim() ||
    (item.content || '')
      .replace(/<[^>]+>/g, ' ')
      .replace(/&nbsp;/gi, ' ')
      .replace(/\s+/g, ' ')
      .trim()
  return text.length > 60 ? `${text.slice(0, 60)}…` : text
}

/* 主楼赞：计数在后端（like_count），「我赞过」只在本地记一个标记 */
const mainLikeKey = computed(() => `bbs-main-like-${route.params.id}`)
const mainLiked = ref(false)
const mainLikeCount = computed(() => Number(experienceShare.value?.like_count) || 0)

const toggleMainLike = async () => {
  const s = experienceShare.value
  if (!s) return
  const nextLiked = !mainLiked.value
  try {
    const res = await experienceShareAPI.like(s.id, nextLiked ? 1 : -1)
    if (res.success) {
      s.like_count = res.data?.like_count ?? mainLikeCount.value + (nextLiked ? 1 : -1)
      mainLiked.value = nextLiked
      localStorage.setItem(mainLikeKey.value, nextLiked ? '1' : '0')
    }
  } catch {
    ElMessage.error('点赞失败，请确认后端已重启')
  }
}

/* ---------- 正文字体（右上角可切换，存 localStorage 本机偏好） ----------
   字体栈写在 :root 的 --post-font 上（inline style 覆盖），好处是 teleport 到 body 的
   发帖/编辑弹窗也能读到，做到「所见即所得」。 */
const FONT_OPTIONS = [
  { label: '微软雅黑', stack: "'Microsoft YaHei', '微软雅黑', system-ui, sans-serif" },
  { label: '华文宋体', stack: "'STSong', '华文宋体', 'Songti SC', SimSun, '宋体', serif" },
  { label: '华文中宋', stack: "'STZhongsong', '华文中宋', 'Songti SC', SimSun, '宋体', serif" },
  { label: '华文楷体', stack: "'STKaiti', '华文楷体', 'KaiTi', '楷体', serif" },
  { label: '华文仿宋', stack: "'STFangsong', '华文仿宋', 'FangSong', '仿宋', serif" },
  { label: '宋体', stack: "SimSun, '宋体', 'Songti SC', serif" },
  { label: '黑体', stack: "'SimHei', '黑体', 'Microsoft YaHei', sans-serif" },
  { label: '系统默认', stack: "system-ui, -apple-system, 'Segoe UI', 'PingFang SC', sans-serif" }
]

const POST_FONT_KEY = 'xingmang.post-font'
const fontLabel = ref(
  FONT_OPTIONS.some(f => f.label === localStorage.getItem(POST_FONT_KEY))
    ? (localStorage.getItem(POST_FONT_KEY) as string)
    : '微软雅黑'
)

const applyPostFont = (label: string) => {
  const opt = FONT_OPTIONS.find(f => f.label === label) || FONT_OPTIONS[0]
  fontLabel.value = opt.label
  localStorage.setItem(POST_FONT_KEY, opt.label)
  /* ① <html> 上挂一份：弹窗 teleport 到 body，只能从这里继承 */
  document.documentElement.style.setProperty('--post-font', opt.stack)
  /* ② .tb-page 自身在 CSS 里也有 --post-font 默认值，就近会盖掉 ①，所以再写一份 inline */
  pageRef.value?.style.setProperty('--post-font', opt.stack)
}

/* ---------- FullScreenLayout 直达时的主题兜底 ----------
   data-theme 由 MainLayout 的 applyTheme() 挂在 <html> 上，本页路由不在
   MainLayout 下，硬刷新 / 直达时没人挂。这里按本地存储的主题模式补挂一次，
   卸载时还原成进入前的状态，不影响返回 MainLayout 后的正常切换。 */
let themeSetByPage = false

const applyStandaloneTheme = () => {
  const root = document.documentElement
  if (root.getAttribute('data-theme')) return
  const saved = localStorage.getItem('main-theme-mode')
  const legacy = localStorage.getItem('main-theme')
  const mode = saved === 'light' || saved === 'dark' || saved === 'system'
    ? saved
    : (legacy === 'dark' ? 'dark' : 'light')
  const dark = mode === 'system'
    ? window.matchMedia('(prefers-color-scheme: dark)').matches
    : mode === 'dark'
  if (dark) {
    root.setAttribute('data-theme', 'dark')
    themeSetByPage = true
  }
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
    if (experienceShare.value) {
      experienceShare.value.content = rewriteContentImgs(experienceShare.value.content || '')
    }
    mainLiked.value = localStorage.getItem(mainLikeKey.value) === '1'

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
  router.push(tutorialBoardPath(detailBoard.value))
}

const goToRecommended = (id: number) => {
  router.push(tutorialPostPath(id))
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
  const el = pageRef.value
  if (!el) return window
  // MainLayout 下滚动的是 .content-frame
  const frame = el.closest('.content-frame') as HTMLElement | null
  if (frame) return frame
  // ⚠ 本页走的是 FullScreenLayout：真正的滚动容器是 .fullscreen-layout（自身 overflow:auto）。
  // 以前只判 .content-frame，落到 window 上 —— 而 window 在这条路由下根本不滚
  // （html/body 被锁成视口高），导致滚动监听永远收不到事件：回到顶部按钮不出现、
  // 目录高亮也不会跟随。这里补上。
  const fs = el.closest('.fullscreen-layout') as HTMLElement | null
  if (fs) return fs
  return window
}

/* ---------- 左栏折叠（rail）+ 目录当前章节高亮 ----------
   折叠状态记在 localStorage，下次进来保持上次的样子。 */
const RAIL_KEY = 'xingmang.detail.rail'
const railCollapsed = ref(false)

const toggleRail = () => {
  railCollapsed.value = !railCollapsed.value
  try {
    localStorage.setItem(RAIL_KEY, railCollapsed.value ? '1' : '0')
  } catch {
    /* 隐私模式下写不了，忽略 */
  }
}

const activeTocId = ref('')

/** 当前读到哪一节：取视口上方（留 140px 余量）最近的一个标题 */
const updateActiveToc = () => {
  const root = articleRef.value
  if (!root) return
  const targets = collectTocTargets(root)
  if (!targets.length) return
  const el = scrollContainer
  const baseTop = el === window || !el ? 0 : (el as HTMLElement).getBoundingClientRect().top
  const threshold = baseTop + 140
  let hit = -1
  targets.forEach((t, i) => {
    if (t.getBoundingClientRect().top <= threshold) hit = i
  })
  // 还没滚过任何标题时，默认视为「正在读第一节」（参考图的目录树就是这种口径：
  // 一进页面就有高亮，而不是等滚到第二个标题才亮）
  activeTocId.value = tocItems.value[hit >= 0 ? hit : 0]?.id || ''
}

/** 章节数＝一级小节（与参考图「N 个章节」口径一致） */
const tocChapterCount = computed(() => tocItems.value.filter(t => t.level === 1).length)

const handlePageScroll = () => {
  const el = scrollContainer
  const top = el === window ? window.scrollY : (el as HTMLElement).scrollTop
  showBackToTop.value = top > 320
  updateActiveToc()
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
  editingId.value = null
  editingText.value = ''
  loadReplies()
})

onMounted(() => {
  applyStandaloneTheme()
  applyPostFont(fontLabel.value)
  fetchSections()
  loadDetail()
  loadReplies()
  scrollContainer = getScrollContainer()
  scrollContainer.addEventListener('scroll', handlePageScroll)

  try {
    railCollapsed.value = localStorage.getItem(RAIL_KEY) === '1'
  } catch {
    railCollapsed.value = false
  }
  nextTick(updateActiveToc)
})

/* 正文是异步加载的：数据到了要先算一次高亮，另外编辑器（TipTap）是自己在 onMounted
   里挂载 DOM 的，正文节点出现得比数据更晚 —— 所以再挂一个 MutationObserver 盯着文章根，
   节点一落位就补算，否则首屏永远没有「当前章节」。 */
let tocObserver: MutationObserver | null = null

watch(
  () => tocItems.value.map(t => t.id).join(','),
  () => {
    nextTick(() => {
      updateActiveToc()
      const root = articleRef.value
      if (root && !tocObserver) {
        tocObserver = new MutationObserver(() => updateActiveToc())
        tocObserver.observe(root, { childList: true, subtree: true })
      }
    })
  }
)

onBeforeUnmount(() => {
  tocObserver?.disconnect()
  tocObserver = null
  scrollContainer?.removeEventListener('scroll', handlePageScroll)
  if (saveStatusTimer) clearTimeout(saveStatusTimer)
  if (themeSetByPage) {
    document.documentElement.removeAttribute('data-theme')
    themeSetByPage = false
  }
})
</script>

<style scoped>
/* ============================================================
   贴吧版式 · 完全复刻 docs/mockups/tieba-plan1-inplace.html
   亮/暗主题变量与 mockup 的 :root / :root[data-theme='dark'] 一一对应
   ============================================================ */
.tb-page {
  --xp-panel: #ffffff;
  --xp-panel-2: #f9fafb;
  --xp-border: #e5e7eb;
  --xp-border-2: #d0d7de;
  --xp-text: #1f2328;
  --xp-text-2: #57606a;
  --xp-text-3: #8b949e;
  --xp-chip: #f1f3f5;
  --xp-floor: #fafbfc;
  --xp-accent-line: #bfdbfe;
  --xp-radius: 8px;
  --xp-shadow: 0 1px 2px rgba(16, 24, 40, 0.04);
  --xp-bar-bg: #1f2937;
  --xp-bar-text: #e6edf3;
  /* 中栏基准宽：原 832px 卡片 → 按 1.2 倍放到 998px（832×1.2 = 998.4）。
     固定宽度，窗口变窄时正文不跟着变窄（先压左右栏，最后才横向滚动）。 */
  --mid-col: 998px;
  /* 帖子正文/标题默认字体：微软雅黑（右上角可切换，切换后由 inline style 覆盖这里）。 */
  --post-font: 'Microsoft YaHei', '微软雅黑', system-ui, sans-serif;
  /* 主题色走 MainLayout 挂在 <html> 上的 --accent（皮肤可换色）；
     本页在 FullScreenLayout 下直达时无人挂载，按 mockup 默认蓝兜底 */
  --tb-accent: var(--accent, #2563eb);
  --tb-accent-soft: var(--accent-soft, #eff6ff);
  min-height: 100vh;
  background: #f4f6f8;
  color: var(--xp-text);
  padding-bottom: 48px;
  font-size: 14px;
}

/* ---------- 顶端栏（返回在左侧） ---------- */
.tb-topbar {
  position: sticky;
  top: 0;
  z-index: 30;
  height: 48px;
  display: flex;
  align-items: center;
  padding: 0 18px;
  background: var(--xp-panel);
  border-bottom: 1px solid var(--xp-border);
}

.tb-back {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-family: inherit;
  font-size: 13.5px;
  line-height: 1;
  padding: 7px 14px;
  border: 1px solid var(--xp-border);
  border-radius: 999px;
  background: transparent;
  color: var(--xp-text-2);
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s, background 0.15s;
}

.tb-back:hover {
  border-color: var(--tb-accent);
  color: var(--tb-accent);
  background: var(--tb-accent-soft);
}

/* ---------- 顶栏右上角：字体选择 ---------- */
.tb-font {
  margin-left: auto;
}

.tb-font-btn {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  font-family: inherit;
  font-size: 13px;
  line-height: 1;
  padding: 7px 12px;
  border: 1px solid var(--xp-border);
  border-radius: 999px;
  background: transparent;
  color: var(--xp-text-2);
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s, background 0.15s;
}

.tb-font-btn:hover {
  border-color: var(--tb-accent);
  color: var(--tb-accent);
  background: var(--tb-accent-soft);
}

.tb-font-label {
  color: var(--xp-text-3);
}

/* 当前字体名用该字体自己渲染，选之前就能看出效果 */
.tb-font-cur {
  font-family: var(--post-font);
  font-size: 14px;
  color: var(--xp-text);
}

.tb-font-caret {
  font-size: 10px;
  color: var(--xp-text-3);
}

/* ⚠ 下拉菜单项由 EP 渲染且 teleport 到 body，只有插槽里的 .tb-font-item 带 data-v */
.tb-font-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  min-width: 148px;
  font-size: 13px;
}

.tb-font-item.on {
  color: var(--tb-accent);
  font-weight: 600;
}

.tb-font-item.on::after {
  content: '✓';
  font-size: 12px;
}

.tb-font-demo {
  font-size: 12.5px;
  color: var(--xp-text-3);
}

.tb-font-item.on .tb-font-demo {
  color: var(--tb-accent);
}

/* ---------- 三栏骨架 ---------- */
.tb-wrap {
  display: grid;
  /* 左右栏给「理想值 → 可压缩下限」区间；中栏固定基准阅读宽度。
     ⚠ 不要给容器加 max-width + margin:auto —— 那会把整块拉到窗口中间，
     左右栏就离两侧越来越远（宽屏下尤其明显）。这里改成占满窗口 + space-between：
     多出来的宽度全部进两条 gutter，左栏贴左、右栏贴右，中栏宽度不受窗口影响。 */
  grid-template-columns: minmax(150px, 200px) var(--mid-col) minmax(220px, 300px);
  gap: 16px;
  justify-content: space-between;
  padding: 18px;
  align-items: start;
  /* 折叠栏宽度变化做个短过渡，避免正文突兀跳动 */
  transition: grid-template-columns 0.18s ease;
}

/* 左栏折叠成 48px 图标条：正文与右栏位置只平移、不重排 */
.tb-wrap.rail-mode {
  grid-template-columns: 48px var(--mid-col) minmax(220px, 300px);
}

/* 顶栏 sticky 52px，侧栏吸附留出其高度 */
.tb-left {
  position: sticky;
  top: 66px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rail-bar {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 6px 0;
  background: var(--xp-panel);
  border: 1px solid var(--xp-border);
  border-radius: var(--xp-radius);
}

.rail-btn {
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  color: var(--xp-text-2);
  font-size: 15px;
  line-height: 1;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.rail-btn:hover {
  background: var(--tb-accent-soft);
  color: var(--tb-accent);
}

:root[data-theme='dark'] .rail-btn:hover {
  background: rgba(88, 166, 255, 0.14);
}

.tb-mid {
  display: flex;
  flex-direction: column;
  /* 主帖卡与回复区合并为一张连续纸面，所以不留缝（方案 A） */
  gap: 0;
  min-width: 0;
}

.tb-right {
  position: sticky;
  top: 66px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-width: 0;
  /* 右栏自己滚，滚动条无痕（firefox 用 scrollbar-width，chromium 用 ::-webkit-scrollbar） */
  max-height: calc(100vh - 84px);
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-width: none;
}

.tb-right::-webkit-scrollbar {
  display: none;
}

/* ---------- 左栏：入口 + 目录 ---------- */
















/* 暗色下两个彩色标签压深，避免刺眼 */
.nav-card {
  background: var(--xp-panel);
  border: 1px solid var(--xp-border);
  border-radius: var(--xp-radius);
  overflow: hidden;
}








/* 目录面板（参考贴吧「文章目录」：头 + 树 + 计数脚） */
.toc-card {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.toc-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 9px 8px 9px 12px;
  background: var(--xp-panel-2);
  border-bottom: 1px solid var(--xp-border);
}

.toc-head-t {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--xp-text-2);
}

.toc-collapse {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  color: var(--xp-text-3);
  font-size: 14px;
  line-height: 1;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.toc-collapse:hover {
  background: var(--tb-accent-soft);
  color: var(--tb-accent);
}

.toc-foot {
  padding: 8px 12px;
  font-size: 11.5px;
  color: var(--xp-text-3);
  border-top: 1px solid var(--xp-border);
  background: var(--xp-panel-2);
}

/* 目录（窄栏内允许换行，长标题不截断） */
.toc-mini {
  padding: 6px;
  max-height: 62vh;
  overflow-y: auto;
}

.toc-mini a {
  display: block;
  padding: 6px 9px;
  border-radius: 5px;
  font-size: 12.5px;
  line-height: 1.5;
  color: var(--xp-text-2);
  text-decoration: none;
}

.toc-mini a:hover {
  background: var(--tb-accent-soft);
  color: var(--tb-accent);
  text-decoration: none;
}

.toc-mini a.lv2 {
  padding-left: 21px;
  font-size: 12px;
  color: var(--xp-text-3);
}

/* 当前阅读章节：左侧 3px 蓝条 + 淡蓝底 + 主色文字 */
.toc-mini a.on,
.toc-mini a.lv2.on {
  position: relative;
  background: var(--tb-accent-soft);
  color: var(--tb-accent);
  font-weight: 600;
}

.toc-mini a.on::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 60%;
  border-radius: 0 2px 2px 0;
  background: var(--tb-accent);
}

:root[data-theme='dark'] .toc-mini a.on,
:root[data-theme='dark'] .toc-mini a.lv2.on {
  background: rgba(88, 166, 255, 0.14);
  color: var(--tb-accent);
}

/* ---------- 中栏：主帖卡 ---------- */
.post-card {
  background: var(--xp-panel);
  border: 1px solid var(--xp-border);
  /* 与下方回复区合并：只留上圆角、去掉下边框与投影 */
  border-radius: var(--xp-radius) var(--xp-radius) 0 0;
  border-bottom: none;
  box-shadow: none;
  padding: 18px 20px 20px;
}





.author {
  font-weight: 600;
  font-size: 14px;
  color: var(--xp-text);
}

.level {
  background: var(--tb-accent-soft);
  color: var(--tb-accent);
  border: 1px solid var(--xp-accent-line);
  font-size: 11px;
  padding: 0 5px;
  border-radius: 3px;
  line-height: 16px;
}

.level.pdf {
  background: #fff2e8;
  border-color: #ffd9bd;
  color: #c8560f;
}

/* 分区标签本身即控件：点开可把帖子移到别的分区 */
.level.sec-chip {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  cursor: pointer;
  user-select: none;
}

.level.sec-chip:hover {
  border-color: var(--tb-accent);
}


/* 精华徽标（mockup 的 badge-flag） */
.badge-flag {
  background: #fde68a;
  color: #92400e;
  font-size: 11px;
  padding: 0 5px;
  border-radius: 3px;
  line-height: 16px;
}




.post-title {
  font-family: var(--post-font);
  font-size: 19px;
  font-weight: 700;
  line-height: 1.4;
  margin: 0 0 14px;
  color: var(--xp-text);
  word-break: break-word;
}

/* ---------- 作者行：左=头像/昵称/标签，右=时间/点赞 ---------- */
.post-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin: 0 0 16px;
}

.pm-left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex-wrap: wrap;
}

.pm-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  flex: none;
  color: #fff;
  font-size: 15px;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pm-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--xp-text);
}

.pm-tag {
  font-size: 11px;
  line-height: 1;
  padding: 4px 8px;
  border-radius: 999px;
  background: var(--xp-chip);
  color: var(--xp-text-2);
  white-space: nowrap;
}

.pm-tag.pdf {
  background: #fff2e8;
  color: #c2410c;
}

.pm-tag.feat {
  background: #fef3c7;
  color: #b45309;
}

.pm-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 5px;
  flex: none;
  margin-left: auto;
}

.pm-time {
  font-size: 12px;
  color: var(--xp-text-3);
}

.pm-like {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-family: inherit;
  font-size: 12px;
  line-height: 1;
  padding: 5px 11px;
  border: 1px solid var(--xp-border);
  border-radius: 999px;
  background: transparent;
  color: var(--xp-text-2);
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s, background 0.15s;
}

.pm-like:hover {
  border-color: var(--tb-accent);
  color: var(--tb-accent);
}

.pm-like.liked {
  border-color: var(--tb-accent);
  color: var(--tb-accent);
  background: var(--tb-accent-soft);
  font-weight: 600;
}

/* 暗色：两个彩色标签压深 */
:root[data-theme='dark'] .pm-tag.pdf {
  background: rgba(234, 88, 12, 0.18);
  color: #fdba74;
}

:root[data-theme='dark'] .pm-tag.feat {
  background: #4a3a12;
  color: #fbbf24;
}

.post-summary {
  font-family: var(--post-font);
  background: var(--xp-panel-2);
  border-left: 3px solid var(--xp-accent-line);
  padding: 9px 12px;
  color: var(--xp-text-2);
  font-size: 13px;
  margin: 0 0 14px;
  border-radius: 0 6px 6px 0;
}

/* 封面（通栏大图） */
.cover-fig {
  margin: 0 0 18px;
}

/* 参考版式：封面通栏、16:9、大圆角、无边框无图注 */
.cover-fig img {
  display: block;
  width: 100%;
  aspect-ratio: 16 / 9;
  object-fit: cover;
  border-radius: 12px;
  background: var(--xp-chip);
}

/* ---------- 正文（编辑器） ---------- */
.article-body {
  min-width: 0;
}

.article-body :deep(.tiptap) {
  font-family: var(--post-font);
  /* 17px 是定稿尺寸（此前按 DPR 反推的 19.2px 偏大）。
     行距 1.9、字距 0.02em 保留：华文宋体笔画细，字距收紧会互相贴住。 */
  font-size: 17px;
  line-height: 1.9;
  letter-spacing: 0.02em;
  color: var(--xp-text);
}

.article-body :deep(.tiptap p) {
  margin: 0 0 12px;
}

/* ---------- 正文内联编辑器（ghost）的主题 ----------
   ⚠ 老问题：TipTapEditor 内部用 `@media (prefers-color-scheme: dark)` 跟**操作系统**，
   不认识应用的 data-theme。系统是暗色、应用切亮色时，工具栏/边框仍然是暗的 ——
   用户看到的就是「亮色页面里冒出一条黑菜单」。
   发帖弹窗 (.cp-editor) 早就强制过，只有正文这个 ghost 编辑器漏了。
   ⚠ 用 `.article-body .wiki-editor` 前缀提特异性，才赢得过组件里媒体查询那段
   （媒体查询不增加特异性，组件那条是 (0,2,0)，这里是 (0,3,0)）。 */
.article-body .wiki-editor {
  border-color: #e8ecf2;
  background: transparent;
}

/* 圆角四角 + 四边细框 + 下方留白：ghost 模式下正文没有外框包着，
   只给上圆角会像「从卡片里切出来的一条」，四角圆角才像独立的工具条浮在正文上。 */
.article-body .wiki-editor :deep(.editor-toolbar) {
  background: #f7f9fc;
  border: 1px solid #eaedf3;
  border-radius: 10px;
  padding: 6px 10px;
  gap: 2px;
  margin-bottom: 10px;
}

.article-body .wiki-editor :deep(.toolbar-btn) {
  min-width: 30px;
  height: 30px;
  border-radius: 6px;
  color: #5b6472;
}

.article-body .wiki-editor :deep(.toolbar-btn:hover:not(:disabled)) {
  background: #e9effc;
  color: var(--accent, #2563eb);
}

.article-body .wiki-editor :deep(.toolbar-btn.is-active) {
  background: var(--accent-soft, #e8f0fe);
  color: var(--accent, #2563eb);
}

.article-body .wiki-editor :deep(.toolbar-divider) { background: #e4e8ef; }

.article-body .wiki-editor :deep(.heading-btn),
.article-body .wiki-editor :deep(.color-btn) {
  border-radius: 6px;
  color: #5b6472;
}

.article-body .wiki-editor :deep(.editor-content) { background: transparent; }

:root[data-theme='dark'] .article-body .wiki-editor {
  border-color: #21262d;
}

:root[data-theme='dark'] .article-body .wiki-editor :deep(.editor-toolbar) {
  background: #1a1f29;
  border-color: #21262d;
}

:root[data-theme='dark'] .article-body .wiki-editor :deep(.toolbar-btn) { color: #9aa7b6; }

:root[data-theme='dark'] .article-body .wiki-editor :deep(.toolbar-btn:hover:not(:disabled)) {
  background: rgba(88, 166, 255, 0.12);
  color: #58a6ff;
}

:root[data-theme='dark'] .article-body .wiki-editor :deep(.toolbar-btn.is-active) {
  background: rgba(88, 166, 255, 0.16);
  color: #58a6ff;
}

:root[data-theme='dark'] .article-body .wiki-editor :deep(.toolbar-divider) { background: #2b3240; }

:root[data-theme='dark'] .article-body .wiki-editor :deep(.heading-btn),
:root[data-theme='dark'] .article-body .wiki-editor :deep(.color-btn) { color: #9aa7b6; }

.article-body :deep(.tiptap ul),
.article-body :deep(.tiptap ol) {
  margin: 0 0 12px;
  padding-left: 22px;
}

.article-body :deep(.tiptap li) {
  margin-bottom: 5px;
}

.article-body :deep(.tiptap h2) {
  /* 20px：比正文(17)明显重、又不压过帖子大标题(19) —— 大标题是页面的，h2 是文章内部的 */
  font-size: 20px;
  font-weight: 600;
  margin: 22px 0 10px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--xp-border);
  color: var(--xp-text);
  scroll-margin-top: 90px;
}

.article-body :deep(.tiptap h3) {
  font-size: 18px;
  font-weight: 600;
  margin: 16px 0 8px;
  color: var(--xp-text);
  scroll-margin-top: 90px;
}

.article-body :deep(.tiptap blockquote) {
  margin: 0 0 12px;
  padding: 9px 14px;
  background: var(--tb-accent-soft);
  border-left: 3px solid var(--tb-accent);
  color: var(--xp-text-2);
  border-radius: 0 6px 6px 0;
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
  border: 1px solid var(--xp-border-2);
}

.article-body :deep(.tiptap table th) {
  background: var(--tb-accent-soft);
}

.article-body :deep(.tiptap pre) {
  background: var(--xp-panel-2);
  border: 1px solid var(--xp-border-2);
  border-radius: 4px;
}

.article-body :deep(.tiptap code) {
  background-color: rgba(27, 31, 35, 0.05);
}

.article-body :deep(.tiptap a) {
  color: var(--tb-accent);
}

/* PDF 正文 */
.pdf-body {
  margin: 0 0 14px;
}

/* 补充说明 */
.notice {
  margin: 0 0 14px;
  padding: 12px 14px;
  background: var(--xp-floor);
  border: 1px solid var(--xp-border-2);
  border-radius: 6px;
  font-size: 13px;
  color: var(--xp-text-2);
  scroll-margin-top: 90px;
}

.notice b {
  color: var(--xp-text);
}

.notice :deep(.tiptap) {
  font-family: var(--post-font);
  margin-top: 8px;
  font-size: 16.8px;
  line-height: 1.9;
  letter-spacing: 0.02em;
  color: var(--xp-text-2);
}

.notice :deep(.tiptap p) {
  margin: 0 0 12px;
}

/* ---------- 附件框（mockup 的折角文件图标） ---------- */
.att-box {
  background: var(--xp-panel);
  border: 1px solid var(--xp-border);
  border-radius: 6px;
  overflow: hidden;
  margin: 16px 0 14px;
  scroll-margin-top: 90px;
}

.ab-h {
  background: var(--tb-accent-soft);
  border-bottom: 1px solid var(--xp-border);
  padding: 8px 12px;
  font-size: 13px;
  font-weight: 600;
  color: var(--xp-text);
}

.att {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 12px;
}

.att .ico-file {
  width: 30px;
  height: 38px;
  flex: none;
  border: 1.5px solid var(--tb-accent);
  border-radius: 3px;
  position: relative;
}

.att .ico-file::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  border-left: 9px solid var(--xp-panel);
  border-top: 9px solid var(--tb-accent);
}

.att .nm {
  flex: 1;
  min-width: 0;
  color: var(--tb-accent);
  font-size: 13.5px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.att .sz {
  color: var(--xp-text-3);
  font-size: 12px;
  flex: none;
}

.att .btns {
  display: flex;
  gap: 7px;
  flex: none;
}

.att .btns a {
  border: 1px solid var(--tb-accent);
  color: var(--tb-accent);
  border-radius: 4px;
  padding: 3px 11px;
  font-size: 12.5px;
  cursor: pointer;
  text-decoration: none;
}

.att .btns a:first-child {
  background: var(--tb-accent);
  color: #fff;
}

.att .btns a:hover {
  text-decoration: none;
}

/* ---------- 互动条 ---------- */





/* 精华状态按钮：黄色系与 badge-flag 呼应 */


.pa-meta {
  color: var(--xp-text-3);
  font-size: 12px;
}

/* ---------- 主帖尾部：正文下方右下角的「编辑信息」 ---------- */
.post-foot {
  margin-top: 12px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}












/* 靠右的「编辑信息」 */
.pf-edit {
  margin-left: auto;
  flex: none;
  font-family: inherit;
  font-size: 13px;
  line-height: 1;
  padding: 8px 16px;
  border: 1px solid var(--xp-border);
  border-radius: 999px;
  background: transparent;
  color: var(--xp-text-2);
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s, background 0.15s;
}

.pf-edit:hover {
  border-color: var(--tb-accent);
  color: var(--tb-accent);
  background: var(--tb-accent-soft);
}

/* 暗色：两个彩色标签压深，避免亮底刺眼 */
/* ---------- 回复区 ---------- */
.reply-card {
  background: var(--xp-panel);
  border: 1px solid var(--xp-border);
  /* 续在主帖卡下方，只留下圆角，不再有独立外壳的分节感 */
  border-top: none;
  border-radius: 0 0 var(--xp-radius) var(--xp-radius);
  box-shadow: var(--xp-shadow);
  overflow: hidden;
}

.rc-head {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 18px 20px 8px;
  border-bottom: none;
  background: transparent;
  flex-wrap: wrap;
}

.rc-head h3 {
  margin: 0;
  font-size: 13.5px;
  font-weight: 500;
  color: var(--xp-text-2);
}

.rc-head .cnt {
  color: var(--xp-text-3);
  font-size: 13px;
}

/* 计数前补一个中点，读起来是「全部回复 · 0」 */
.rc-head .cnt::before {
  content: '·';
  margin-right: 6px;
}

.sort-tabs {
  margin-left: auto;
  display: flex;
  gap: 2px;
}

.sort-tabs button {
  font-family: inherit;
  background: transparent;
  border: none;
  color: var(--xp-text-2);
  font-size: 13px;
  padding: 4px 10px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease;
}

.sort-tabs button:hover {
  background: var(--xp-border);
}

.sort-tabs button.on {
  background: var(--tb-accent-soft);
  color: var(--tb-accent);
  font-weight: 500;
}

.bbs-empty {
  padding: 26px 20px;
  text-align: center;
  font-size: 13px;
  color: var(--xp-text-3);
  border-bottom: none;
  background: transparent;
}

/* ---------- 楼层（参考图版式 · 紧凑密度） ---------- */
.floor {
  display: flex;
  gap: 10px;
  padding: 12px 20px;
  border-bottom: 1px solid var(--xp-border);
  transition: background-color 0.12s ease;
}

.floor:hover {
  background: var(--xp-floor);
}

.floor-av {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  flex: none;
  color: #fff;
  font-size: 13px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.floor-main {
  flex: 1;
  min-width: 0;
}

/* 昵称 + 徽标 + 右侧「时间 #楼号」压在同一行 —— 紧凑密度的关键，省掉一次换行 */
.floor-head {
  display: flex;
  align-items: center;
  gap: 7px;
  min-width: 0;
}

.floor-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--xp-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.floor-badge {
  flex: none;
  font-size: 11px;
  line-height: 1.45;
  padding: 0 5px;
  border-radius: 4px;
  color: var(--tb-accent);
  border: 1px solid var(--tb-accent);
}

.floor-when {
  margin-left: auto;
  flex: none;
  font-size: 11.5px;
  color: var(--xp-text-3);
}

.floor-content {
  margin-top: 7px;
  font-size: 13.5px;
  line-height: 1.6;
  color: var(--xp-text);
  word-break: break-word;
}

.floor-content .emoji {
  font-size: 16px;
}

/* 操作行右对齐，与参考图的「回复 ♡N」同位；自己发的才多出编辑/删除 */
.floor-actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.floor-actions button,
.sub-acts button {
  background: none;
  border: none;
  color: var(--xp-text-3);
  font-size: 12.5px;
  cursor: pointer;
  padding: 0;
  font-family: inherit;
  transition: color 0.15s ease;
}

.sub-acts button {
  font-size: 12px;
}

.floor-actions button:hover,
.sub-acts button:hover {
  color: var(--tb-accent);
}

.floor-actions button.liked,
.sub-acts button.liked,
.floor-actions button.save,
.sub-acts button.save {
  color: var(--tb-accent);
  font-weight: 600;
}

.floor-actions button.del:hover,
.sub-acts button.del:hover {
  color: #e54d42;
}

/* ---------- 楼中楼（二级封顶：回复子楼只挂 @昵称，不再往下缩进） ---------- */
.sub-list {
  margin-top: 8px;
  padding: 8px 10px;
  background: var(--xp-panel-2);
  border-radius: 6px;
}

.sub-item + .sub-item {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px dashed var(--xp-border);
}

.sub-head {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.sub-name {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--xp-text-2);
}

.sub-at {
  font-size: 11.5px;
  color: var(--tb-accent);
}

.sub-when {
  margin-left: auto;
  font-size: 11px;
  color: var(--xp-text-3);
}

.sub-content {
  margin-top: 5px;
  font-size: 12.5px;
  line-height: 1.6;
  color: var(--xp-text);
  word-break: break-word;
}

.sub-acts {
  display: flex;
  gap: 10px;
  margin-top: 5px;
  justify-content: flex-end;
}





/* 回复框 */
.reply-box {
  border-top: none;
  padding: 14px 20px 20px;
  background: transparent;
  scroll-margin-top: 90px;
}

/* 「回复 @某人」提示条 */
.reply-quote {
  margin-bottom: 10px;
  padding: 8px 12px;
  background: var(--xp-panel-2);
  border-left: 3px solid var(--xp-border-2);
  border-radius: 0 6px 6px 0;
  font-size: 12px;
  color: var(--xp-text-2);
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
  color: var(--xp-text-2);
  flex-shrink: 0;
}

.reply-box textarea {
  width: 100%;
  min-height: 62px;
  resize: vertical;
  background: var(--xp-panel);
  border: 1px solid var(--xp-border-2);
  border-radius: 6px;
  padding: 9px 11px;
  color: var(--xp-text);
  font: inherit;
  font-size: 13.5px;
  line-height: 1.7;
  outline: none;
}

.reply-box textarea:focus {
  border-color: var(--tb-accent);
}

.rb-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 9px;
  flex-wrap: wrap;
}

.nick-label {
  font-size: 12px;
  color: var(--xp-text-2);
  flex: none;
}

.nick-label input {
  width: 100px;
  padding: 3px 8px;
  border: 1px solid var(--xp-border-2);
  border-radius: 4px;
  font-size: 12px;
  font-family: inherit;
  outline: none;
  background: var(--xp-panel);
  color: var(--xp-text);
}

.nick-label input:focus {
  border-color: var(--tb-accent);
}

.emoticons {
  display: flex;
  gap: 3px;
  flex-wrap: wrap;
}

.emoticons button {
  font-size: 13px;
  padding: 2px 6px;
  border: none;
  background: transparent;
  border-radius: 4px;
  cursor: pointer;
  line-height: 1.4;
  font-family: inherit;
}

.emoticons button:hover {
  background: var(--tb-accent-soft);
}

.rb-bar .hint {
  color: var(--xp-text-3);
  font-size: 12px;
  flex: 1;
  text-align: right;
  min-width: 0;
}

.rb-bar .submit {
  background: var(--tb-accent);
  color: #fff;
  border: none;
  border-radius: 4px;
  padding: 6px 18px;
  font-size: 13px;
  cursor: pointer;
  font-family: inherit;
  font-weight: 500;
}

.rb-bar .submit:hover {
  opacity: 0.9;
}

.edit-textarea {
  width: 100%;
  min-height: 70px;
  padding: 8px 10px;
  border: 1px solid var(--xp-border-2);
  border-radius: 6px;
  font-size: 13px;
  font-family: inherit;
  line-height: 1.7;
  resize: vertical;
  outline: none;
  background: var(--xp-panel);
  color: var(--xp-text);
}

.edit-textarea:focus {
  border-color: var(--tb-accent);
}

/* ---------- Tab 面板（历史 / 日志） ---------- */
.tab-panel {
  padding: 16px 18px;
}

.tab-panel h2 {
  font-size: 15px;
  font-weight: 600;
  margin: 0 0 12px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--xp-border);
  color: var(--xp-text);
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
  border-bottom: 1px solid var(--xp-border);
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
  color: var(--tb-accent);
  border: 1px solid var(--tb-accent);
  border-radius: 3px;
  padding: 3px 8px;
}

.h-info .h-time {
  font-size: 12px;
  color: var(--xp-text-2);
}

.h-info .h-desc {
  margin-top: 3px;
  font-size: 14px;
  color: var(--xp-text);
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
  border-bottom: 1px solid var(--xp-border);
  font-size: 13px;
}

.log-item:last-child {
  border-bottom: none;
}

.log-time {
  flex-shrink: 0;
  font-family: Consolas, monospace;
  font-size: 12px;
  color: var(--xp-text-2);
}

.log-action {
  flex: 1;
  color: var(--xp-text);
}

.log-user {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--tb-accent);
}

/* ---------- 右栏：side-card（吧名片 + 相关推荐） ---------- */
.side-card {
  background: var(--xp-panel);
  border: 1px solid var(--xp-border);
  border-radius: var(--xp-radius);
  overflow: hidden;
}

.sc-h {
  padding: 10px 14px;
  font-size: 13.5px;
  font-weight: 600;
  color: var(--xp-text);
  border-bottom: 1px solid var(--xp-border);
  background: var(--xp-panel-2);
  display: flex;
  align-items: center;
  justify-content: space-between;
  scroll-margin-top: 90px;
}

.sc-h .more {
  font-size: 12px;
  color: var(--xp-text-3);
  font-weight: 400;
}














/* 相关推荐 */
/* 相关推荐卡片：与列表页「相关推荐」同款（标题 → 摘要 → 16:9 大图 → 底部信息） */
.rel-item {
  display: block;
  padding: 10px 14px 9px;
  border-bottom: 1px solid var(--xp-border);
  color: var(--xp-text);
  cursor: pointer;
  text-decoration: none;
}

.rel-item:last-child { border-bottom: none; }

.rel-item:hover {
  background: var(--tb-accent-soft);
  text-decoration: none;
}

.ri-t {
  font-size: 13px;
  font-weight: 500;
  line-height: 1.5;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  color: var(--xp-text);
}

.ri-x {
  margin-top: 5px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--xp-text-3);
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

/* 关键：宽度撑满 + 16:9 + 圆角 8，尺寸与列表页相关推荐一致 */
.ri-img {
  margin-top: 8px;
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: 8px;
  background: var(--xp-chip, #f1f3f5);
  overflow: hidden;
}

.ri-img img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.ri-foot {
  margin-top: 7px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.ri-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  min-width: 0;
  font-size: 11px;
  color: var(--xp-text-3);
}

.ri-chip i {
  width: 12px;
  height: 12px;
  border-radius: 4px;
  flex: none;
}

.ri-chip span {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ri-nums {
  flex: none;
  font-size: 11px;
  color: var(--xp-text-3);
  font-variant-numeric: tabular-nums;
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
  border: 1px solid var(--tb-accent);
  border-radius: 4px;
  background: var(--xp-panel);
  color: var(--tb-accent);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.12);
  cursor: pointer;
  font-family: inherit;
  transition: transform 0.2s ease, background-color 0.2s ease;
}

.backtop-float:hover {
  background: var(--tb-accent-soft);
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

/* ---------- 响应式（mockup 断点） ---------- */
/* ⚠ 断点跟着中栏宽度走：中栏 998 + 左右栏下限 150/220 + 间距 32 + 内边距 36 = 1436px。
   低于它就算把左右栏压到最小也不够，会出现横向滚动条 —— 宁可提前收起右栏，
   也不要在临界区间留一根滚动条。所以断点定在 1440。 */
@media (max-width: 1440px) {
  .tb-wrap {
    grid-template-columns: minmax(120px, 200px) var(--mid-col);
  }

  .tb-wrap.rail-mode {
    grid-template-columns: 48px var(--mid-col);
  }

  .tb-right {
    display: none;
  }
}

@media (max-width: 860px) {
  /* 手机级宽度：放弃 832px 下限，恢复流式，避免小屏横向滚动 */
  .tb-wrap {
    grid-template-columns: minmax(0, 1fr);
    padding: 14px;
  }

  .tb-left {
    display: none;
  }
}

@media (max-width: 600px) {
  .post-title {
    font-size: 18px;
  }

  .rc-head {
    flex-wrap: wrap;
  }

  .sort-tabs {
    margin-left: 0;
    width: 100%;
  }

  .floor {
    padding: 11px 14px;
  }

  .floor-when {
    font-size: 11px;
  }

  .rb-bar .hint {
    text-align: left;
  }
}

/* ---------- 暗色主题（mockup 的 :root[data-theme='dark'] 变量） ---------- */
:root[data-theme='dark'] .tb-page {
  --xp-panel: #1f2937;
  --xp-panel-2: #1a2230;
  --xp-border: #374151;
  --xp-border-2: #374151;
  --xp-text: #e6edf3;
  --xp-text-2: #9ca3af;
  --xp-text-3: #6e7681;
  --xp-chip: #232b3a;
  --xp-floor: #1a2230;
  --xp-accent-line: #2c4a6b;
  --xp-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
  --xp-bar-bg: #16181d;
  --xp-bar-text: #e6edf3;
  /* 暗色下 accent 用当文字的亮版（与 MainLayout 暗色行为一致） */
  --tb-accent: var(--accent, #58a6ff);
  --tb-accent-soft: var(--accent-soft, #16273d);
  background: #16181d;
  color: #e6edf3;
}



:root[data-theme='dark'] .badge-flag {
  background: #4a3a12;
  color: #fbbf24;
}

:root[data-theme='dark'] .level.pdf {
  background: rgba(234, 88, 12, 0.18);
  border-color: rgba(234, 88, 12, 0.45);
  color: #fdba74;
}

:root[data-theme='dark'] .article-body :deep(.tiptap pre) {
  background: rgba(22, 27, 34, 0.95);
  border-color: rgba(42, 50, 66, 0.3);
}

:root[data-theme='dark'] .article-body :deep(.tiptap code) {
  background-color: rgba(28, 35, 51, 0.6);
  color: #e6edf3;
}


/* 暗色下 --accent 是亮蓝（当文字用），做实心按钮底配白字对比度不足（约 2.5:1），
   实心按钮统一压回 EP 主蓝 */
:root[data-theme='dark'] .follow-btn:not(.ghost),
:root[data-theme='dark'] .rb-bar .submit,
:root[data-theme='dark'] .att .btns a:first-child {
  background: #3b82f6;
  border-color: #3b82f6;
  color: #fff;
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

:root[data-theme='dark'] .floor-actions button.del:hover {
  color: #f87171;
}

/* ---------- 全屏编辑弹窗（与列表页发帖同款） ---------- */
/* ---------- 沉浸式发帖页（C+D 合成：整页 + 两步） ----------
   teleport 到 body：--accent 可用（挂在 html），--xp-* 不可用 → 字面量 + 暗色覆盖。 */
.composer {
  position: fixed;
  inset: 0;
  z-index: 2600;
  background: #ffffff;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.cp-topbar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 28px;
  height: 60px;
  flex: none;
  border-bottom: 1px solid #eef0f2;
}
.cp-back {
  border: none;
  background: transparent;
  font-family: inherit;
  font-size: 14px;
  color: #5f6368;
  cursor: pointer;
  padding: 6px 10px;
  border-radius: 8px;
}
.cp-back:hover { background: #f5f6f7; color: #202124; }
.cp-title {
  flex: 1;
  text-align: center;
  font-size: 15px;
  font-weight: 500;
  color: #202124;
}
.cp-publish {
  border: none;
  background: var(--accent, #2563eb);
  color: #fff;
  font-family: inherit;
  font-size: 14px;
  font-weight: 500;
  padding: 9px 30px;
  border-radius: 999px;
  cursor: pointer;
}
.cp-publish:disabled { opacity: .6; cursor: default; }
.cp-steps {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14px;
  padding: 16px 0 4px;
  flex: none;
}
.cp-step { font-size: 13px; color: #9aa0a8; }
.cp-step.on { color: var(--accent, #2563eb); font-weight: 500; }
.cp-step-line { width: 64px; height: 1px; background: #e5e7eb; }
.cp-body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  width: min(860px, calc(100vw - 48px));
  margin: 0 auto;
  padding: 18px 0 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.cp-row {
  display: flex;
  align-items: center;
  gap: 14px;
}
.cp-row.top { align-items: flex-start; }
.cp-label {
  width: 76px;
  flex: none;
  font-size: 14px;
  color: #5f6368;
  text-align: right;
  line-height: 32px;
}
.cp-row.top .cp-label { line-height: 1.5; padding-top: 6px; }
.cp-bar {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #202124;
}
.cp-bar-av {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: linear-gradient(135deg, #60a5fa, #2563eb);
  color: #fff;
  font-size: 13px;
  font-weight: 500;
  font-style: normal;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.cp-title-input :deep(.el-input__inner) {
  height: 52px;
  font-size: 17px;
}
.cp-title-input :deep(.el-input__wrapper) { border-radius: 10px; }
.cp-editor { min-height: 46vh; display: flex; flex-direction: column; }
.cp-editor :deep(.editor-content) { flex: 1; min-height: 0; }
/* 帖子正文默认字体：华文宋体（与详情页正文一致；弹窗 teleport 到 body，用不了 --post-font） */
.cp-editor :deep(.editor-content .tiptap) {
  font-family: var(--post-font, 'Microsoft YaHei', '微软雅黑', system-ui, sans-serif);
  font-size: 19.2px;
  line-height: 1.95;
  letter-spacing: 0.02em;
}
.cp-summary-input :deep(.el-textarea__inner) { border-radius: 10px; }

/* 发帖页编辑器跟随应用主题（组件用 prefers-color-scheme 跟系统，需在页内强制）。
   ⚠ cp-editor 与 tiptap-editor 是同一个元素（class 合并在组件根），
   写 `.cp-editor :deep(.tiptap-editor…)` 会变成要求「祖先关系」而永不匹配，
   必须用同元素 compound 选择器 `.cp-editor.tiptap-editor` 才能赢过组件媒体查询。 */
.cp-editor.tiptap-editor {
  border-radius: 12px;
  border-color: #e5e7eb;
  background: #ffffff;
}
.cp-editor :deep(.editor-toolbar) {
  background: #fafafa;
  border-bottom-color: #e5e7eb;
  border-radius: 12px 12px 0 0;
}
.cp-editor :deep(.toolbar-btn) { color: #4b5563; }
.cp-editor :deep(.toolbar-btn:hover:not(:disabled)) { background: #eef2ff; color: var(--accent, #2563eb); }
.cp-editor :deep(.toolbar-btn.is-active) { background: var(--accent-soft, rgba(37, 99, 235, 0.1)); color: var(--accent, #2563eb); }
.cp-editor :deep(.toolbar-divider) { background: #e5e7eb; }
.cp-editor :deep(.heading-btn),
.cp-editor :deep(.color-btn) { color: #4b5563; }
.cp-editor.tiptap-editor:not(.is-ghost) :deep(.editor-content) { background: #ffffff; }
.cp-editor.tiptap-editor:not(.is-ghost) :deep(.editor-content .tiptap p) { color: #1f2937; }
.cp-editor.tiptap-editor:not(.is-ghost) :deep(.editor-content .tiptap h1),
.cp-editor.tiptap-editor:not(.is-ghost) :deep(.editor-content .tiptap h2) { color: #111827; }
:root[data-theme='dark'] .cp-editor.tiptap-editor {
  background: #0d1117;
  border-color: #30363d;
}
:root[data-theme='dark'] .cp-editor :deep(.editor-toolbar) {
  background: #10161f;
  border-bottom-color: #30363d;
}
:root[data-theme='dark'] .cp-editor :deep(.toolbar-btn) { color: #9ca3af; }
:root[data-theme='dark'] .cp-editor :deep(.toolbar-btn:hover:not(:disabled)) { background: rgba(88, 166, 255, 0.12); color: var(--accent, #58a6ff); }
:root[data-theme='dark'] .cp-editor :deep(.toolbar-btn.is-active) { background: rgba(88, 166, 255, 0.16); color: var(--accent, #58a6ff); }
:root[data-theme='dark'] .cp-editor :deep(.toolbar-divider) { background: #30363d; }
:root[data-theme='dark'] .cp-editor :deep(.heading-btn),
:root[data-theme='dark'] .cp-editor :deep(.color-btn) { color: #9ca3af; }
:root[data-theme='dark'] .cp-editor.tiptap-editor:not(.is-ghost) :deep(.editor-content) { background: #0d1117; }
:root[data-theme='dark'] .cp-editor.tiptap-editor:not(.is-ghost) :deep(.editor-content .tiptap p) { color: #e5e7eb; }
:root[data-theme='dark'] .cp-editor.tiptap-editor:not(.is-ghost) :deep(.editor-content .tiptap h1),
:root[data-theme='dark'] .cp-editor.tiptap-editor:not(.is-ghost) :deep(.editor-content .tiptap h2) { color: #f3f4f6; }

.cp-footer {
  flex: none;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 14px;
  padding: 14px 28px;
  border-top: 1px solid #eef0f2;
}
.cp-skip {
  border: none;
  background: transparent;
  font-family: inherit;
  font-size: 14px;
  color: #5f6368;
  cursor: pointer;
  padding: 9px 18px;
  border-radius: 999px;
}
.cp-skip:hover { background: #f5f6f7; color: #202124; }
.cp-skip:disabled { opacity: .6; cursor: default; }
.cp-next {
  border: none;
  background: var(--accent, #2563eb);
  color: #fff;
  font-family: inherit;
  font-size: 14px;
  font-weight: 500;
  padding: 10px 26px;
  border-radius: 999px;
  cursor: pointer;
}
.cp-next:hover { filter: brightness(1.05); }
.cp-next:disabled { opacity: .6; cursor: default; }


/* ---------- 沉浸式发帖页 · 暗色覆盖 ---------- */
:root[data-theme='dark'] .composer { background: #161b22; }
:root[data-theme='dark'] .cp-topbar { border-bottom-color: #21262d; }
:root[data-theme='dark'] .cp-back { color: #9ca3af; }
:root[data-theme='dark'] .cp-back:hover { background: #1f2630; color: #e5e7eb; }
:root[data-theme='dark'] .cp-title { color: #e5e7eb; }
:root[data-theme='dark'] .cp-step { color: #6b7280; }
:root[data-theme='dark'] .cp-step-line { background: #30363d; }
:root[data-theme='dark'] .cp-label { color: #9ca3af; }
:root[data-theme='dark'] .cp-bar { color: #e5e7eb; }
:root[data-theme='dark'] .cp-title-input :deep(.el-input__wrapper),
:root[data-theme='dark'] .cp-summary-input :deep(.el-textarea__inner) {
  background-color: #0d1117;
  box-shadow: 0 0 0 1px #30363d inset;
}
:root[data-theme='dark'] .cp-footer { border-top-color: #21262d; }
:root[data-theme='dark'] .cp-skip { color: #9ca3af; }
:root[data-theme='dark'] .cp-skip:hover { background: #1f2630; color: #e5e7eb; }
:root[data-theme='dark'] .composer .sec-chip { background: #262d36; color: #d1d5db; }
:root[data-theme='dark'] .composer .sec-chip:hover { background: #2d3540; }
:root[data-theme='dark'] .composer .sec-chip.on { background: rgba(88, 166, 255, 0.16); color: var(--accent, #58a6ff); }
:root[data-theme='dark'] .composer .cover-upload-area { background: #0d1117; border-color: #30363d; }
:root[data-theme='dark'] .composer .cover-upload-text { color: #d1d5db; }
:root[data-theme='dark'] .composer .cover-upload-hint { color: #6b7280; }
:root[data-theme='dark'] .composer .cover-preview-card { background: #0d1117; border-color: #30363d; }
:root[data-theme='dark'] .composer .declare-hint { color: #6b7280; }


/* 版块分区 chips（贴吧样式：灰底胶囊，选中亮蓝） */
.sec-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.sec-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.sec-chip {
  border: none;
  background: #f5f5f6;
  color: #374151;
  font-family: inherit;
  font-size: 13px;
  line-height: 1;
  padding: 9px 18px;
  border-radius: 999px;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}
.sec-chip:hover { background: #ececee; }
.sec-chip.on {
  background: var(--accent-soft, rgba(37, 99, 235, 0.1));
  color: var(--accent, #2563eb);
  font-weight: 600;
}

.declare-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.declare-hint {
  font-size: 12px;
  color: #9ca3af;
}

.cover-field {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.cover-preview-card {
  overflow: hidden;
  max-width: 420px;
  border-radius: 6px;
  border: 1px solid rgb(224, 224, 230);
  background: #ffffff;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.08);
}

.cover-preview-image {
  display: block;
  width: 100%;
  height: 220px;
  object-fit: cover;
}
</style>
