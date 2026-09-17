<template>
  <div class="creative-page">
    <div class="page-header">
      <div class="header-content">
        <h1>抽卡区</h1>
        <p class="subtitle">选择生成器，开启你的创作之旅</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" class="create-generator-btn" @click="openGeneratorManager">
          <el-icon><Setting /></el-icon>
          生成器管理
        </el-button>
      </div>
    </div>

    <div class="generators-grid">
      <div
        v-for="generator in generators"
        :key="generator.id"
        class="generator-card"
        @click="openGenerator(generator)"
      >
        <div class="gc-head">
          <span class="gc-icon">
            <el-icon :size="18"><component :is="getIconComponent(generator.icon)" /></el-icon>
          </span>
          <span class="gc-name">{{ generator.name }}</span>
          <span class="gc-cat">{{ getGeneratorCategory(generator) }}</span>
          <el-tag v-if="generator.isNew" size="small" type="danger" class="new-tag">NEW</el-tag>
        </div>
        <div class="gc-sample">
          <span class="gc-sample-label">输出示例</span>
          <span class="gc-sample-text">{{ getGeneratorExample(generator) }}</span>
        </div>
        <div class="gc-foot">
          <span class="gc-last">{{ getLastUsedText(generator) }}</span>
          <span class="gc-arrow" aria-hidden="true">
            <el-icon :size="14"><ArrowRight /></el-icon>
          </span>
        </div>
      </div>
    </div>

    <!-- 生成器弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      width="calc(100vw - 128px)"
      align-center
      destroy-on-close
      class="generator-dialog"
      modal-class="generator-dialog-modal"
      :show-close="false"
      append-to-body
    >
      <!-- 自定义标题栏 -->
      <template #header>
        <div class="gen-dialog-header">
          <div class="gen-dialog-title">
            <el-icon class="gen-title-icon"><MagicStick /></el-icon>
            <span>{{ currentGenerator?.name || '创意生成' }}</span>
          </div>
          <div class="gen-dialog-header-right">
            <a
              v-if="currentGenerator?.tutorial_url"
              :href="currentGenerator.tutorial_url"
              target="_blank"
              rel="noopener noreferrer"
              class="gen-tutorial-btn"
            >教程</a>
            <el-button
              type="primary"
              link
              size="small"
              class="gen-close-btn"
              @click="dialogVisible = false"
            >
              <el-icon :size="18"><Close /></el-icon>
            </el-button>
          </div>
        </div>
      </template>

      <!-- 第一步：配置视图（左模板 + 右表单） -->
      <div v-if="dialogStep === 'config'" class="dialog-content">
        <!-- 左侧：提示词模板展示区 -->
        <div class="prompt-template-area">
          <div class="template-header">
            <el-select
              v-model="selectedGeneratorCardPacks"
              multiple
              size="small"
              placeholder="选择卡包"
              class="card-pack-select"
              popper-class="gen-select-popper"
              :max-collapse-tags="1"
              @change="onCardPackSelectionChange"
            >
              <el-option
                v-for="pack in availableCardPacks"
                :key="pack"
                :label="pack"
                :value="pack"
              />
            </el-select>
          </div>
          <div class="template-list">
            <div
              v-for="prompt in pinnedPrompts"
              :key="prompt.id"
              class="template-item"
              :class="{ active: prompt.enabled }"
              @click="selectQuickPrompt(prompt.id)"
            >
              <el-button
                class="template-item-detail"
                size="small"
                @click.stop="promptSelectDialogVisible = true"
              >详情</el-button>
              <span class="template-item-name">{{ prompt.name }}</span>
            </div>
            <div v-if="pinnedPrompts.length === 0" class="empty-template">
              <el-icon :size="32"><DocumentRemove /></el-icon>
              <p>点击"详情"按钮选择提示词模板</p>
            </div>
          </div>
          <div class="template-footer">
            <el-button
              size="large"
              class="more-prompt-btn"
              @click="promptSelectDialogVisible = true"
            >
              更多提示词
            </el-button>
          </div>
        </div>

        <!-- 右侧：表单区 -->
        <div class="form-area">
          <!-- AI模型 -->
          <div class="form-block">
            <div class="form-block-label">
              <span>AI模型</span>
              <el-button
                type="primary"
                link
                size="small"
                class="form-block-action"
                @click="showFixedPrompt"
              >
                <el-icon><View /></el-icon>
                <span>查看完整提示词</span>
              </el-button>
            </div>
            <el-select
              v-model="selectedModel"
              placeholder="思考者"
              size="large"
              class="full-width"
              popper-class="form-area-select-popper"
            >
              <el-option
                v-for="model in models"
                :key="model.id"
                :label="model.name"
                :value="model.id"
              >
                <div class="model-option">
                  <span class="model-name">{{ model.name }}</span>
                  <el-tag v-if="model.provider_name" size="small" type="info">{{ model.provider_name }}</el-tag>
                </div>
              </el-option>
            </el-select>
          </div>

          <!-- 提示词模板选择 -->
          <div class="form-block">
            <div class="form-block-label">提示词</div>
            <el-radio-group v-model="promptPanelMode" class="prompt-radio-group">
              <el-radio-button value="quick">快捷选项</el-radio-button>
              <el-radio-button value="custom">自定义</el-radio-button>
              <el-radio-button value="more">更多</el-radio-button>
            </el-radio-group>
            <!-- 快捷选项模式：显示选中模板名 -->
            <template v-if="promptPanelMode === 'quick'">
              <el-input
                v-if="selectedTemplateName"
                :model-value="selectedTemplateName"
                readonly
                size="large"
                class="prompt-selected-display"
              >
                <template #suffix>
                  <el-icon class="is-loading" v-if="isLoadingPinnedPrompts"><Loading /></el-icon>
                </template>
              </el-input>
              <div v-else class="quick-prompt-empty">
                <span>暂无已选模板，点击「更多」从提示词库选择</span>
              </div>
            </template>
            <!-- 自定义模式 -->
            <div v-else-if="promptPanelMode === 'custom'" class="custom-prompt-container">
              <el-input
                v-model="customPromptContent"
                type="textarea"
                :autosize="{ minRows: 4, maxRows: 10 }"
                placeholder="请输入自定义提示词内容..."
                maxlength="2000"
                show-word-limit
                resize="none"
              />
            </div>
            <div v-if="currentTemplateDescription && promptPanelMode === 'quick'" class="form-tip">
              <span class="form-tip-text">使用方法：</span>
              <span class="form-tip-content">{{ currentTemplateDescription }}</span>
              <a class="form-tip-link form-tip-link-strong" @click="showFixedPrompt">查看介绍</a>
            </div>
          </div>

          <!-- 任务参数：动态字段 -->
          <template v-if="hasAnyFields">
            <div
              v-for="field in allFields"
              :key="field.name"
              class="form-block"
            >
              <div class="form-block-label">
                {{ field.label }}
                <span
                  v-if="field.required !== false"
                  class="required-mark"
                  title="必填"
                >*</span>
              </div>
              <div v-if="field.description" class="field-description">{{ field.description }}</div>

              <div v-if="field.type === 'text' || field.type === 'textarea'" class="input-container">
                <el-input
                  v-if="field.type === 'text'"
                  v-model="fieldValues[field.name]"
                  type="textarea"
                  :autosize="{ minRows: 1, maxRows: 6 }"
                  :placeholder="`请输入${field.label}`"
                  maxlength="500"
                  show-word-limit
                  resize="none"
                />

                <el-input
                  v-else-if="field.type === 'textarea'"
                  v-model="fieldValues[field.name]"
                  type="textarea"
                  :autosize="{ minRows: 2, maxRows: 8 }"
                  :placeholder="`请输入${field.label}`"
                  maxlength="500"
                  show-word-limit
                  resize="none"
                />

                <el-button
                  class="fullscreen-btn"
                  circle
                  size="small"
                  @click="openFieldFullscreenEditor(field.name, field.label)"
                  title="全屏编辑"
                >
                  <el-icon><FullScreen /></el-icon>
                </el-button>
              </div>

              <el-select
                v-else-if="field.type === 'select'"
                v-model="fieldValues[field.name]"
                :placeholder="`请选择${field.label}`"
                class="full-width"
                size="large"
                popper-class="gen-select-popper"
              >
                <el-option
                  v-for="(option, optionIndex) in field.options"
                  :key="option"
                  :label="field.optionLabels?.[optionIndex] || option"
                  :value="option"
                />
              </el-select>
            </div>
          </template>

          <!-- 补充信息 -->
          <div class="form-block">
            <div class="form-block-label">补充信息</div>
            <el-input
              v-model="additionalInfo"
              type="textarea"
              :rows="3"
              maxlength="500"
              show-word-limit
              placeholder=""
              resize="none"
            />
          </div>

          <!-- 关联作品 -->
          <div class="form-block">
            <div class="form-block-label">关联作品(可选)</div>
            <el-select
              v-model="selectedBookId"
              clearable
              filterable
              placeholder="选择一个作品以关联该作品的章节或备忘录"
              class="full-width"
              size="large"
              popper-class="gen-select-popper"
            >
              <el-option
                v-for="book in books"
                :key="book.id"
                :label="book.title"
                :value="book.id"
              />
            </el-select>
          </div>

          <!-- 关联章节 -->
          <div v-if="selectedBookId" class="form-block">
            <div class="form-block-label">关联章节</div>
            <el-select
              v-model="selectedChapterIds"
              multiple
              collapse-tags
              collapse-tags-tooltip
              placeholder="选择要关联的章节"
              class="full-width"
              size="large"
              popper-class="gen-select-popper"
            >
              <el-option
                v-for="chapter in availableChapters"
                :key="chapter.id"
                :label="chapter.title"
                :value="chapter.id"
              />
            </el-select>
          </div>

          <!-- 关联备忘录 -->
          <div v-if="selectedBookId" class="form-block">
            <div class="form-block-label">关联备忘录</div>
            <el-select
              v-model="selectedMemoId"
              clearable
              filterable
              placeholder="选择一个备忘录"
              class="full-width"
              size="large"
              popper-class="gen-select-popper"
            >
              <el-option
                v-for="memo in memos"
                :key="memo.id"
                :label="memo.title || memo.content?.slice(0, 30) + '...'"
                :value="memo.id"
              />
            </el-select>
          </div>

          <!-- 关联角色卡 -->
          <div v-if="selectedBookId" class="form-block">
            <div class="form-block-label">关联角色卡</div>
            <el-select
              v-model="selectedCharacterIds"
              multiple
              collapse-tags
              collapse-tags-tooltip
              placeholder="选择要关联的角色卡"
              class="full-width"
              size="large"
              popper-class="gen-select-popper"
            >
              <el-option
                v-for="character in availableCharacters"
                :key="character.id"
                :label="character.name"
                :value="character.id"
              />
            </el-select>
          </div>

          <!-- 关联内容（已关联列表） -->
          <div v-if="hasRelatedContext" class="form-block">
            <div class="form-block-label">已关联内容</div>
            <div class="context-preview-list">
              <div
                v-for="item in relatedContextItems"
                :key="`${item.type}-${item.id}`"
                class="context-preview-item"
              >
                <el-tag size="small" type="info">{{ getRelatedTypeLabel(item.type) }}</el-tag>
                <span>{{ item.title }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 第二步：结果视图（生成结果 + 内联追问，替代独立预览/追问弹窗） -->
      <div v-else class="dialog-content result-step">
        <div class="result-toolbar">
          <div class="result-title">
            <el-icon><MagicStick /></el-icon>
            <span>生成结果</span>
          </div>
          <div class="result-actions">
            <el-button
              v-if="generating"
              type="danger"
              link
              size="small"
              @click="stopGenerate"
            >
              <el-icon><Close /></el-icon>
              停止
            </el-button>
            <el-button
              v-if="!generating && !isEditingPreview"
              type="primary"
              link
              size="small"
              @click="isEditingPreview = true"
            >
              <el-icon><EditPen /></el-icon>
              编辑
            </el-button>
            <el-button
              v-if="isEditingPreview"
              type="success"
              link
              size="small"
              @click="isEditingPreview = false"
            >
              <el-icon><Check /></el-icon>
              完成
            </el-button>
            <el-button
              type="primary"
              link
              size="small"
              @click="copyPreviewContent"
            >
              <el-icon><CopyDocument /></el-icon>
              复制
            </el-button>
            <el-button
              v-if="!generating && previewContent"
              type="success"
              link
              size="small"
              @click="importToMemo"
            >
              <el-icon><Document /></el-icon>
              一键导入
            </el-button>
            <el-button
              v-if="!generating && previewContent"
              type="warning"
              link
              size="small"
              @click="showFollowUpPanel = !showFollowUpPanel"
            >
              <el-icon><ChatDotRound /></el-icon>
              {{ showFollowUpPanel ? '收起追问' : '追问' }}
            </el-button>
          </div>
        </div>

        <el-scrollbar class="result-scrollbar">
          <el-input
            v-if="isEditingPreview"
            v-model="previewContent"
            type="textarea"
            :autosize="{ minRows: 20 }"
            placeholder="编辑内容..."
          />
          <template v-else>
            <div v-if="generating && !previewContent" class="generating-hint">
              <el-icon class="is-loading"><Loading /></el-icon>
              <span>正在生成中...</span>
            </div>
            <MarkdownRenderer v-else :content="previewContent || '暂无生成内容'" />
          </template>
        </el-scrollbar>

        <!-- 内联追问面板 -->
        <div v-if="showFollowUpPanel" class="followup-inline">
          <div class="followup-inline-header">
            <el-icon><ChatDotRound /></el-icon>
            <span>追问对话</span>
            <el-select
              v-model="followUpSelectedModelId"
              placeholder="选择模型"
              class="followup-inline-model"
              popper-class="gen-select-popper"
              size="small"
            >
              <el-option
                v-for="model in models"
                :key="model.id"
                :label="model.name"
                :value="model.id"
              />
            </el-select>
          </div>
          <div class="followup-messages" ref="followUpMessagesContainer">
            <div v-if="followUpMessages.length > 2 && !followUpHistoryExpanded" class="history-collapsed-hint">
              <el-button
                text
                type="primary"
                @click="followUpHistoryExpanded = true"
                size="small"
                class="expand-btn"
              >
                <el-icon><ArrowDown /></el-icon>
                展开 {{ followUpMessages.length - 2 }} 条历史消息
              </el-button>
            </div>

            <template v-if="followUpHistoryExpanded">
              <transition-group name="message-slide">
                <div
                  v-for="(msg, msgIndex) in followUpMessages"
                  :key="msgIndex"
                  :class="['followup-message-bubble', msg.role, { 'history-message': msgIndex < followUpMessages.length - 2 }]"
                >
                  <div class="followup-message-sender">
                    <el-icon class="sender-icon">
                      <User v-if="msg.role === 'user'" />
                      <ChatDotRound v-if="msg.role === 'assistant'" />
                      <Setting v-if="msg.role === 'system'" />
                    </el-icon>
                    <span class="sender-name">
                      {{ msg.role === 'user' ? '用户' : msg.role === 'assistant' ? 'AI' : '系统' }}
                    </span>
                  </div>
                  <div class="followup-message-content">
                    <MarkdownRenderer :content="msg.content" />
                  </div>
                </div>
              </transition-group>
            </template>

            <div v-if="followUpGenerating" class="followup-message-bubble assistant">
              <div class="followup-message-sender">
                <el-icon class="sender-icon"><ChatDotRound /></el-icon>
                <span class="sender-name">AI</span>
              </div>
              <div class="followup-message-content">
                <div class="typing-indicator">
                  <span></span>
                  <span></span>
                  <span></span>
                </div>
              </div>
            </div>
          </div>

          <div class="followup-input-area">
            <el-input
              v-model="followUpUserInput"
              type="textarea"
              :rows="2"
              placeholder="输入您的问题，按 Ctrl+Enter 发送..."
              @keydown.enter.ctrl="sendFollowUpMessage"
              :disabled="followUpGenerating"
              resize="none"
              class="followup-input"
            />
            <div class="followup-input-actions">
              <el-tooltip content="发送 (Ctrl+Enter)" placement="top">
                <el-button
                  type="primary"
                  @click="sendFollowUpMessage"
                  :loading="followUpGenerating"
                  :disabled="!followUpUserInput.trim()"
                  :icon="ChatDotRound"
                  circle
                  class="send-btn"
                />
              </el-tooltip>
            </div>
          </div>
        </div>
      </div>

      <!-- 自定义底部 -->
      <template #footer>
        <div class="gen-dialog-footer">
          <div class="gen-footer-disclaimer">以上内容均由AI生成,仅供参考和借鉴</div>
          <div class="gen-footer-actions">
            <!-- 配置步骤：生成 / 停止 -->
            <template v-if="dialogStep === 'config'">
              <el-button
                v-if="!generating"
                type="success"
                size="large"
                round
                :disabled="!canGenerate"
                class="gen-generate-btn"
                @click="generateContent"
              >
                <el-icon><Promotion /></el-icon>
                <span>生成</span>
              </el-button>
              <el-button
                v-else
                type="danger"
                size="large"
                round
                class="gen-generate-btn"
                @click="stopGenerate"
              >
                <el-icon><Close /></el-icon>
                <span>停止生成</span>
              </el-button>
            </template>
            <!-- 结果步骤：返回配置 / 重新生成 -->
            <template v-else>
              <el-button size="large" round @click="dialogStep = 'config'">
                返回配置
              </el-button>
              <el-button
                v-if="!generating"
                type="success"
                size="large"
                round
                class="gen-generate-btn"
                @click="generateContent"
              >
                <el-icon><Promotion /></el-icon>
                <span>重新生成</span>
              </el-button>
            </template>
          </div>
        </div>
      </template>
    </el-dialog>

    <FieldFullscreenDialog
      v-model:visible="fieldFullscreenDialogVisible"
      :field-label="fieldFullscreenFieldLabel"
      :initial-content="fieldFullscreenContent"
      @save="onFieldFullscreenSave"
    />

    <PromptSelectDialog
      v-model:visible="promptSelectDialogVisible"
      :prompts="prompts"
      :pinned-prompts="pinnedPrompts"
      :generator-name="currentGenerator?.name || ''"
      :card-packs="selectedGeneratorCardPacks"
      @toggle-pin="togglePromptPin"
      @remove-pin="removePinnedPrompt"
      @toggle-enabled="togglePinnedPrompt"
      @create="onPromptCreated"
    />

    <FixedPromptDialog
      v-model:visible="promptEditDialogVisible"
      :initial-prompt="editingPrompt"
      @save="onPromptSaved"
    />

    <FixedPromptPreviewDialog
      v-model:visible="fixedPromptDialogVisible"
      :content="fixedPromptContent"
    />

    <GeneratorManagerDialog
      v-model:visible="generatorManagerVisible"
      :generators="allGenerators"
      @save="onManagerSave"
      @delete="onManagerDelete"
      @move="onManagerMove"
    />

    <HistoryDialog
      v-model:visible="showHistoryDialog"
      :records="historyRecords"
      @clear="clearHistory"
      @delete="deleteHistoryRecord"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowRight,
  ArrowDown,
  ArrowUp,
  Document,
  Cpu,
  EditPen,
  View,
  CopyDocument,
  Setting,
  Check,
  Plus,
  Delete,
  Lightning,
  Close,
  User,
  ChatDotRound,
  MagicStick,
  Reading,
  Tickets,
  FullScreen,
  CollectionTag,
  OfficeBuilding,
  Star,
  Sunny,
  Moon,
  Grid,
  Picture,
  Van,
  Brush,
  TrophyBase,
  Medal,
  DataLine,
  TrendCharts,
  Promotion,
  Clock,
  Timer,
  Switch,
  Key,
  Coin,
  Histogram,
  PieChart,
  DocumentRemove,
  InfoFilled,
  Loading
} from '@element-plus/icons-vue'
import { promptAPI, configAPI, generatorAPI, bookAPI, chapterAPI, memoAPI, characterAPI } from '@/api'
import type { Generator, Book, Chapter, Memo, Character, RelatedContent } from '@/types'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import FieldFullscreenDialog from './creative/FieldFullscreenDialog.vue'
import HistoryDialog from './creative/HistoryDialog.vue'
import FixedPromptDialog from './creative/FixedPromptDialog.vue'
import FixedPromptPreviewDialog from './creative/FixedPromptPreviewDialog.vue'
import GeneratorManagerDialog from './creative/GeneratorManagerDialog.vue'
import PromptSelectDialog from './creative/PromptSelectDialog.vue'
import './creative/creative.css'
import './creative/generator-dialog.css'
import './creative/followup.css'
import './creative/prompt-edit.css'
import './creative/generator-manager.css'
import './creative/prompt-select.css'
import './creative/element-plus.css'
import './creative/history.css'
import './creative/dark-theme.css'

// 默认生成器基础配置（用于初始化）
const DEFAULT_GENERATORS_VERSION = '2026-03-22-v2'

const defaultGeneratorConfigs = [
  {
    id: 1,
    name: '脑洞生成器',
    description: '围绕题材、设定、冲突和反转，快速产出高概念创意点子。',
    icon: 'MagicStick',
    core_prompt: '你是一个专业的脑洞生成器助手，擅长围绕题材、设定、冲突、爽点和反转产出新颖且可落地的小说创意。根据用户输入，直接输出有吸引力的创意方案、卖点和延展方向。不要解释你的思考过程，不要输出多余客套话，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 2,
    name: '书名生成器',
    description: '根据题材、受众和卖点生成有传播力、有网文感的书名。',
    icon: 'Document',
    tutorial_url: 'https://www.bilibili.com/video/BV1CYKdzfE7c',
    core_prompt: '你是一个专业的书名生成器助手，擅长根据题材、关键词、目标读者和核心卖点生成有吸引力、易传播、符合网文风格的书名。可提供多个方向，但只输出书名结果及简短标签，不要解释推理过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 3,
    name: '简介生成器',
    description: '生成抓人眼球的作品简介、投放文案和一句话卖点。',
    icon: 'EditPen',
    core_prompt: '你是一个专业的简介生成器助手，擅长根据作品题材、人物关系、冲突主线和核心爽点，生成抓人眼球的小说简介。输出内容要强钩子、强冲突、强代入感，可包含一句话卖点或多版简介，但不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 4,
    name: '大纲生成器',
    description: '生成完整主线大纲，梳理故事结构、阶段目标和高潮节点。',
    icon: 'Reading',
    core_prompt: '你是一个专业的大纲生成器助手，擅长根据题材设定、角色关系和核心冲突生成完整的故事大纲。输出应包含故事主线、阶段推进、关键矛盾、高潮节点和结局方向，结构清晰、节奏明确。不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 5,
    name: '细纲生成器',
    description: '把大纲拆到章节或剧情段落，明确每段目标、冲突和悬念。',
    icon: 'Tickets',
    core_prompt: '你是一个专业的细纲生成器助手，擅长把已有题材或大纲进一步细化为章节级、情节级的执行方案。输出内容应明确每一段剧情的目标、冲突、推进、转折、情绪点和悬念钩子。不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 6,
    name: '黄金开篇生成器',
    description: '设计前三章或开篇剧情，强化钩子、冲突和读者留存。',
    icon: 'Lightning',
    core_prompt: '你是一个专业的黄金开篇生成器助手，擅长设计小说开篇的前三章或首个高能场景。输出要突出第一眼钩子、人物处境、核心冲突、情绪爆点和继续追读的悬念，确保开头强抓人。不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 7,
    name: '金手指生成器',
    description: '生成贴合题材的外挂、系统、能力或特殊资源设定。',
    icon: 'Cpu',
    core_prompt: '你是一个专业的金手指生成器助手，擅长根据题材、主角定位和成长路线设计独特且有升级空间的金手指设定。输出应包含能力机制、使用限制、成长方式、爽点来源和剧情联动性。不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 8,
    name: '名字生成器',
    description: '生成人名、地名、宗门名、势力名等，兼顾风格和辨识度。',
    icon: 'CollectionTag',
    core_prompt: '你是一个专业的名字生成器助手，擅长根据题材、时代感、地域风格和人物气质生成人名、地名、组织名、功法名等。输出名字时要注意辨识度、记忆点和风格统一，不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 9,
    name: '人设生成器',
    description: '生成立体角色设定，覆盖身份、性格、目标、秘密和成长线。',
    icon: 'User',
    core_prompt: '你是一个专业的人设生成器助手，擅长生成立体鲜明、适合剧情推进的角色设定。输出应包含身份背景、外在特征、性格标签、行为逻辑、欲望目标、弱点秘密和成长空间。不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 10,
    name: '世界观生成器',
    description: '构建世界规则、势力格局、修炼体系和社会运行逻辑。',
    icon: 'OfficeBuilding',
    core_prompt: '你是一个专业的世界观生成器助手，擅长构建完整且自洽的小说世界观。输出应包含时代背景、核心规则、力量体系、地理版图、势力结构、社会秩序和潜在矛盾，保证可持续支撑长篇剧情。不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  }
]

// 生成器类型
type GeneratorItem = {
  id: number
  name: string
  description: string
  icon: string
  core_prompt?: string
  remark?: string
  order_num?: number
  isDefault?: boolean
  isCustom?: boolean
  isNew?: boolean
  example?: string
  tutorial_url?: string
}

// 所有生成器列表（从localStorage加载默认生成器的修改）
const allGenerators = ref<GeneratorItem[]>([])

// 自定义生成器列表（来自数据库）
const customGenerators = ref<Generator[]>([])

// 合并后的生成器列表（用于显示）
const generators = computed(() => {
  return allGenerators.value.map(g => ({
    ...g,
    isCustom: !g.isDefault
  }))
})

// 历史记录相关
const showHistoryDialog = ref(false)

interface HistoryMessage {
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: number
}

interface HistoryRecord {
  id: number
  generatorName: string
  generatorIcon?: string
  fixedPrompt: string
  selectedPrompts: Array<{ name: string; content: string }>
  inputParams: string
  additionalInfo: string
  fieldValues: Record<string, string>
  relatedContext: Array<{ type: RelatedContent['type']; title: string }>
  messages: HistoryMessage[]
  previewContent: string
  timestamp: number
}

const historyRecords = ref<HistoryRecord[]>([])

type CreativeModelOption = {
  id: number
  name: string
  model: string
  provider_name?: string
  is_default?: number | boolean
}

// 模型列表
const models = ref<CreativeModelOption[]>([])
const books = ref<Book[]>([])
const memos = ref<Memo[]>([])
const availableChapters = ref<Chapter[]>([])
const availableCharacters = ref<Character[]>([])
const loadingBookContext = ref(false)

const getDefaultModelId = () => {
  if (models.value.length === 0) return ''
  const defaultModel = models.value.find(model => model.is_default === 1 || model.is_default === true)
  return defaultModel?.id ?? ''
}

// 加载模型列表
const fetchModels = async () => {
  try {
    const res = await configAPI.getAll()
    if (res.success && res.data) {
      models.value = res.data.filter(m => m.enabled !== 0)
      selectedModel.value = getDefaultModelId()
    }
  } catch (error) {
    // 使用默认数据
    models.value = [
      { id: 1, name: 'GPT-4', model: 'gpt-4' },
      { id: 2, name: 'GPT-3.5', model: 'gpt-3.5-turbo' },
      { id: 3, name: 'Claude', model: 'claude-3-opus-20240229' },
      { id: 4, name: 'Gemini', model: 'gemini-pro' }
    ]
    selectedModel.value = getDefaultModelId()
  }
}

// 状态
const dialogVisible = ref(false)
// 主生成弹窗两段式：'config' 配置表单 / 'result' 生成结果（含追问）
const dialogStep = ref<'config' | 'result'>('config')
// 结果页内联追问面板开关（替代独立的追问弹窗）
const showFollowUpPanel = ref(false)
const currentGenerator = ref<GeneratorItem | null>(null)
const selectedPrompt = ref<number>()
const selectedModel = ref<number | string>('')
const inputParams = ref('')
const additionalInfo = ref('')
// 'more' 是「更多」按钮的瞬时值：点击后由下方 watch 立即回退到原模式并打开提示词库弹窗
const promptPanelMode = ref<'quick' | 'custom' | 'more'>('quick')
const customPromptContent = ref('')
const selectedTemplateId = ref<number | null>(null)
const selectedBookId = ref<number | null>(null)
const selectedChapterIds = ref<number[]>([])
const selectedMemoId = ref<number | null>(null)
const selectedCharacterIds = ref<number[]>([])
const generating = ref(false)
const generatedContent = ref('')
const abortController = ref<AbortController | null>(null)
const isEditingPreview = ref(false)

type PromptItem = {
  id: number
  name: string
  category: string
  content: string
  description?: string
  author?: string
  version?: string
  fields?: Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; optionLabels?: string[]; description: string; required: boolean }>
  usage_count?: number
  created_at?: string
  is_featured?: boolean
}

const prompts = ref<PromptItem[]>([])

// 固定提示词相关状态
const pinnedPrompts = ref<Array<PromptItem & { enabled: boolean }>>([])
const promptSelectDialogVisible = ref(false)
const selectedGeneratorCardPacks = ref<string[]>([])

// 子组件触发新建提示词
const onPromptCreated = (newPrompt: PromptItem) => {
  prompts.value.unshift(newPrompt)
  savePromptsToStorage()
}

const savePromptsToStorage = () => {
  if (currentGenerator.value) {
    const key = `prompts_${currentGenerator.value.id}`
    localStorage.setItem(key, JSON.stringify(prompts.value))
  }
}


const isLoadingPinnedPrompts = ref(false)

// 预览弹窗状态
const previewContent = ref('')
const currentConversationMessages = ref<HistoryMessage[]>([])

// 追问对话状态
const followUpMessages = ref<HistoryMessage[]>([])
const followUpUserInput = ref('')
const followUpGenerating = ref(false)
const followUpMessagesContainer = ref<HTMLElement | null>(null)
const followUpHistoryExpanded = ref(false)
const followUpSelectedModelId = ref<number | string>('')

// 字段值存储
const fieldValues = ref<Record<string, string>>({})

// 动态字段全屏编辑相关状态
const fieldFullscreenDialogVisible = ref(false)
const fieldFullscreenContent = ref('')
const fieldFullscreenFieldName = ref('')
const fieldFullscreenFieldLabel = ref('')

const openFieldFullscreenEditor = (fieldName: string, fieldLabel: string) => {
  fieldFullscreenFieldName.value = fieldName
  fieldFullscreenFieldLabel.value = fieldLabel
  fieldFullscreenContent.value = fieldValues.value[fieldName] || ''
  fieldFullscreenDialogVisible.value = true
}

const onFieldFullscreenSave = (content: string) => {
  fieldValues.value[fieldFullscreenFieldName.value] = content
  fieldFullscreenDialogVisible.value = false
  fieldFullscreenFieldName.value = ''
  fieldFullscreenFieldLabel.value = ''
  fieldFullscreenContent.value = ''
}

// 提示词编辑相关状态
const promptEditDialogVisible = ref(false)
const editingPrompt = ref<{ id: number; name: string; content: string; category: string }>({
  id: 0,
  name: '',
  content: '',
  category: ''
})

// 固定提示词弹窗状态
const fixedPromptDialogVisible = ref(false)
const fixedPromptContent = ref('')

// 可编辑的固定提示词状态
const editableFixedPrompt = ref('')
const savedFixedPrompt = ref('')
const hasUserEditedFixedPrompt = ref(false)

// 生成器管理相关状态
const generatorManagerVisible = ref(false)

// 图标映射表：将图标名称映射到 Element Plus 图标组件
const iconMap: Record<string, any> = {
  MagicStick,
  Document,
  EditPen,
  Reading,
  Tickets,
  Lightning,
  Cpu,
  CollectionTag,
  User,
  OfficeBuilding,
  Star,
  Sunny,
  Moon,
  Picture,
  Timer,
  TrophyBase,
  Medal,
  Coin,
  DataLine,
  TrendCharts,
  Promotion,
  Clock,
  Brush,
  Setting,
  Grid
}

const getIconComponent = (iconName?: string) => {
  return (iconName && iconMap[iconName]) || MagicStick
}

// === 抽卡区卡片「识别优于回忆」：示例输出 / 分类 / 最近使用 ===
// 默认生成器示例输出（卡片直接展示"生成出来长什么样"）
const DEFAULT_GENERATOR_EXAMPLES: Record<number, string> = {
  1: '「一个每天醒来都会重置回同一天的渔村，只有主角记得。」',
  2: '《开局被雷劈，我觉醒创世能力》',
  3: '「当所有人都以为她是废物，她默默翻开了第 99 卷禁书……」',
  4: '卷一·觉醒 → 卷二·势成 → 卷三·逆命',
  5: '第3章 冲突升级 → 第4章 反转 → 第5章 高潮',
  6: '开篇 800 字即遇生死危机，钩子拉满',
  7: '「万象神瞳」— 看破一切伪象',
  8: '轩辕 · 白芷 · 天机阁 · 苍梧剑宗',
  9: '苏瑾 · 孤傲天才 · 秘密：身世成谜',
  10: '九州大陆 · 灵气复苏 · 三界格局'
}

// 默认生成器分类（用于标签语义）
const DEFAULT_GENERATOR_CATEGORIES: Record<number, string> = {
  1: '灵感', 2: '命名', 3: '灵感', 4: '结构', 5: '结构',
  6: '灵感', 7: '设定', 8: '命名', 9: '设定', 10: '设定'
}

const getGeneratorCategory = (generator: Partial<GeneratorItem>) => {
  const cat = generator.id != null ? DEFAULT_GENERATOR_CATEGORIES[generator.id] : undefined
  return generator.isDefault ? (cat || '创作') : '自定义'
}

// 输出示例的用户覆盖（localStorage 按生成器 id 存储，管理弹窗可编辑）
const getExampleOverrides = (): Record<string, string> => {
  try { return JSON.parse(localStorage.getItem('creative-generator-examples') || '{}') } catch { return {} }
}
const saveExampleOverride = (id: number | undefined, example: string | undefined) => {
  if (id == null) return
  const overrides = getExampleOverrides()
  if (example && example.trim()) overrides[String(id)] = example
  else delete overrides[String(id)]
  localStorage.setItem('creative-generator-examples', JSON.stringify(overrides))
}

const getGeneratorExample = (generator: Partial<GeneratorItem>) => {
  // 1) 用户覆盖的示例优先
  const override = generator.id != null ? getExampleOverrides()[String(generator.id)] : undefined
  if (override && override.trim()) return override
  // 2) 默认生成器的内置示例
  if (generator.isDefault && generator.id != null) {
    const example = DEFAULT_GENERATOR_EXAMPLES[generator.id]
    if (example) return example
  }
  // 3) 自定义生成器：退化为功能描述首句，保证卡片始终有可识别的内容
  return generator.description?.slice(0, 32) || '输入你的设定，开始创作'
}

// 最近使用时间（依据历史记录中的 generatorName 匹配）
const getLastUsedText = (generator: GeneratorItem) => {
  const records = historyRecords.value.filter(r => r.generatorName === generator.name)
  if (records.length === 0) return '尚未使用'
  const last = Math.max(...records.map(r => r.timestamp))
  const diffDays = Math.floor((Date.now() - last) / 86400000)
  if (diffDays < 1) return '今天用过'
  if (diffDays === 1) return '昨天用过'
  if (diffDays < 7) return `${diffDays} 天前用过`
  return '更早用过'
}

// 可选择的图标列表（用于图标选择器）
const availableIcons = computed(() => {
  return Object.keys(iconMap).map(name => ({
    name,
    component: iconMap[name]
  }))
})

// 计算属性
const selectedPromptInfo = computed(() => {
  return prompts.value.find(p => p.id === selectedPrompt.value)
})

const allFields = computed(() => {
  const fields: Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; optionLabels?: string[]; description: string; required: boolean }> = []
  pinnedPrompts.value.filter(p => p.enabled).forEach(prompt => {
    if (prompt.fields) {
      prompt.fields.forEach(field => {
        if (!fields.find(f => f.name === field.name)) {
          fields.push(field)
        }
      })
    }
  })
  return fields
})

const hasAnyFields = computed(() => {
  return allFields.value.length > 0
})

// 生成器与卡包的映射关系
const GENERATOR_TO_CATEGORY_MAP: Record<string, string> = {
  '脑洞生成器': '脑洞',
  '书名生成器': '书名',
  '简介生成器': '简介',
  '大纲生成器': '大纲',
  '细纲生成器': '细纲',
  '黄金开篇生成器': '黄金开篇',
  '金手指生成器': '金手指',
  '名字生成器': '名字',
  '人设生成器': '人设',
  '世界观生成器': '世界观'
}


const additionalInfoLength = computed(() => additionalInfo.value.length)

const currentTemplateDescription = computed(() => {
  if (!selectedTemplateId.value) return ''
  const target = pinnedPrompts.value.find(p => p.id === selectedTemplateId.value)
  return target?.description || ''
})

// 已启用（快捷选项）的提示词模板
const enabledPrompts = computed(() => {
  return pinnedPrompts.value.filter(p => p.enabled)
})

// 选中模板的名称（用于快捷选项模式展示）
const selectedTemplateName = computed(() => {
  if (!selectedTemplateId.value) return ''
  const target = pinnedPrompts.value.find(p => p.id === selectedTemplateId.value)
  return target?.name || ''
})

// 快捷选项：单选一个模板
const selectQuickPrompt = (promptId: number) => {
  pinnedPrompts.value.forEach(p => {
    p.enabled = p.id === promptId
  })
  selectedTemplateId.value = promptId
}

// 提示词面板 radio 切换：点击「更多」打开提示词库弹窗，并回退到原模式
watch(promptPanelMode, (mode, oldMode) => {
  if (mode === 'more') {
    promptPanelMode.value = oldMode || 'quick'
    promptSelectDialogVisible.value = true
  }
})

const selectedBook = computed(() => {
  return books.value.find(book => book.id === selectedBookId.value) || null
})

const selectedChapters = computed(() => {
  return availableChapters.value.filter(chapter => selectedChapterIds.value.includes(chapter.id))
})

const selectedMemo = computed(() => {
  return memos.value.find(memo => memo.id === selectedMemoId.value) || null
})

const selectedCharacters = computed(() => {
  return availableCharacters.value.filter(character => selectedCharacterIds.value.includes(character.id))
})

const relatedContextItems = computed<RelatedContent[]>(() => {
  const items: RelatedContent[] = []

  if (selectedBook.value) {
    const bookLines = [
      `作品名称：${selectedBook.value.title}`,
      selectedBook.value.author ? `作者：${selectedBook.value.author}` : '',
      selectedBook.value.category ? `分类：${selectedBook.value.category}` : '',
      selectedBook.value.tags?.length ? `标签：${selectedBook.value.tags.join('、')}` : '',
      selectedBook.value.description ? `简介：${selectedBook.value.description}` : ''
    ].filter(Boolean)

    items.push({
      type: 'book',
      id: selectedBook.value.id,
      title: selectedBook.value.title,
      content: bookLines.join('\n')
    })
  }

  selectedChapters.value.forEach(chapter => {
    items.push({
      type: 'chapter',
      id: chapter.id,
      title: chapter.title,
      content: chapter.content
    })
  })

  if (selectedMemo.value) {
    items.push({
      type: 'memo',
      id: selectedMemo.value.id,
      title: selectedMemo.value.title,
      content: selectedMemo.value.content
    })
  }

  selectedCharacters.value.forEach(character => {
    const characterLines = [
      character.gender ? `性别：${character.gender}` : '',
      character.folder ? `分组：${character.folder}` : '',
      character.personality ? `性格：${character.personality}` : '',
      character.info ? `设定：${character.info}` : ''
    ].filter(Boolean)

    items.push({
      type: 'character',
      id: character.id,
      title: character.name,
      content: characterLines.join('\n')
    })
  })

  return items
})

const hasRelatedContext = computed(() => relatedContextItems.value.length > 0)

const canGenerate = computed(() => {
  if (!selectedModel.value) return false

  // 自定义模式：检查自定义提示词内容
  if (promptPanelMode.value === 'custom') {
    const hasCustomPrompt = customPromptContent.value.trim() !== ''
    if (!hasCustomPrompt) return false
  } else {
    // 快捷选项模式：检查固定提示词或已启用的模板
    const hasPromptFoundation = Boolean(editableFixedPrompt.value.trim()) || pinnedPrompts.value.some(p => p.enabled)
    if (!hasPromptFoundation) return false
  }

  // 补充信息为非必填：只要有提示词或自定义输入即可生成
  // 动态字段中的必填字段仍需校验
  if (allFields.value.length > 0) {
    const requiredFieldsCompleted = allFields.value.every(field => {
      if (field.required === false) return true
      const value = fieldValues.value[field.name]
      return value && value.trim() !== ''
    })

    if (!requiredFieldsCompleted) return false
  }

  return true
})

// 检查固定提示词是否已修改
const hasFixedPromptChanged = computed(() => {
  return editableFixedPrompt.value !== savedFixedPrompt.value && editableFixedPrompt.value !== ''
})

// 方法
const togglePromptPin = (prompt: { id: number; name: string; category: string; content: string; fields?: Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; optionLabels?: string[]; description: string; required: boolean }> }) => {
  const index = pinnedPrompts.value.findIndex(p => p.id === prompt.id)
  if (index > -1) {
    pinnedPrompts.value.splice(index, 1)
  } else {
    pinnedPrompts.value.push({ ...prompt, enabled: false })
  }
}

const removePinnedPrompt = (promptId: number) => {
  const index = pinnedPrompts.value.findIndex(p => p.id === promptId)
  if (index > -1) {
    pinnedPrompts.value.splice(index, 1)
  }
}

const togglePinnedPrompt = (promptId: number) => {
  const target = pinnedPrompts.value.find(p => p.id === promptId)
  if (target) {
    target.enabled = !target.enabled
  }
}

const copyPreviewContent = () => {
  navigator.clipboard.writeText(previewContent.value)
  ElMessage.success('已复制到剪贴板')
}

const importToMemo = async () => {
  if (!previewContent.value) {
    ElMessage.warning('没有可导入的内容')
    return
  }
  
  try {
    await ElMessageBox.prompt('请输入备忘录标题', '导入到备忘录', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPlaceholder: '例如：创意灵感记录',
      inputValidator: (value) => {
        if (!value || value.trim() === '') {
          return '标题不能为空'
        }
        return true
      }
    }).then(async ({ value }) => {
      const title = value.trim()
      const memoData: Partial<Memo> = {
        title: title,
        content: previewContent.value,
        category: '创意灵感',
        order_num: 0
      }
      
      const res = await memoAPI.create(memoData)
      if (res.success) {
        ElMessage.success('已成功导入到备忘录')
        // 刷新备忘录列表
        await fetchMemos()
      } else {
        ElMessage.error('导入失败，请稍后重试')
      }
    })
  } catch (error) {
    // 用户取消操作
  }
}

const getPinnedPromptsKey = (generatorId: number) => {
  return `generator_pinned_prompts_${generatorId}`
}

const savePinnedPromptsToStorage = () => {
  if (currentGenerator.value) {
    const key = getPinnedPromptsKey(currentGenerator.value.id)
    localStorage.setItem(key, JSON.stringify(pinnedPrompts.value))
  }
}

const loadPinnedPromptsFromStorage = (generatorId: number) => {
  const key = getPinnedPromptsKey(generatorId)
  const saved = localStorage.getItem(key)
  if (saved) {
    try {
      return JSON.parse(saved)
    } catch {
      return []
    }
  }
  return []
}

const getGeneratorCardPacksKey = (generatorId: number) => {
  return `generator_card_packs_${generatorId}`
}

const saveGeneratorCardPacksToStorage = () => {
  if (currentGenerator.value) {
    const key = getGeneratorCardPacksKey(currentGenerator.value.id)
    localStorage.setItem(key, JSON.stringify(selectedGeneratorCardPacks.value))
  }
}

const loadGeneratorCardPacksFromStorage = (generatorId: number): string[] => {
  const key = getGeneratorCardPacksKey(generatorId)
  const saved = localStorage.getItem(key)
  if (saved) {
    try {
      return JSON.parse(saved)
    } catch {
      return []
    }
  }
  return []
}

const onCardPackSelectionChange = (value: string[]) => {
  if (value.length > 2) {
    selectedGeneratorCardPacks.value = value.slice(0, 2)
    return
  }
  saveGeneratorCardPacksToStorage()
}

const availableCardPacks = computed(() => {
  const packs = new Set<string>()
  prompts.value.forEach(p => {
    if (p.category) {
      packs.add(p.category)
    }
  })
  return Array.from(packs)
})

const getFormStateKey = (generatorId: number) => {
  return `generator_form_state_${generatorId}`
}

const saveFormStateToStorage = () => {
  if (currentGenerator.value) {
    const key = getFormStateKey(currentGenerator.value.id)
    const formState = {
      selectedModel: selectedModel.value,
      inputParams: inputParams.value,
      additionalInfo: additionalInfo.value,
      selectedBookId: selectedBookId.value,
      selectedChapterIds: selectedChapterIds.value,
      selectedMemoId: selectedMemoId.value,
      selectedCharacterIds: selectedCharacterIds.value,
      fieldValues: fieldValues.value,
      editableFixedPrompt: editableFixedPrompt.value,
      savedFixedPrompt: savedFixedPrompt.value,
      hasUserEditedFixedPrompt: hasUserEditedFixedPrompt.value
    }
    localStorage.setItem(key, JSON.stringify(formState))
  }
}

const loadFormStateFromStorage = (generatorId: number) => {
  const key = getFormStateKey(generatorId)
  const saved = localStorage.getItem(key)
  if (saved) {
    try {
      return JSON.parse(saved)
    } catch {
      return null
    }
  }
  return null
}

const syncGeneratorPromptState = (generator: Pick<GeneratorItem, 'id' | 'core_prompt'>) => {
  const latestPrompt = generator.core_prompt || ''
  const savedFormState = loadFormStateFromStorage(generator.id)

  if (savedFormState) {
    const key = getFormStateKey(generator.id)
    localStorage.setItem(key, JSON.stringify({
      ...savedFormState,
      editableFixedPrompt: latestPrompt,
      savedFixedPrompt: latestPrompt,
      hasUserEditedFixedPrompt: Boolean(latestPrompt)
    }))
  }

  if (currentGenerator.value?.id === generator.id) {
    currentGenerator.value = {
      ...currentGenerator.value,
      core_prompt: latestPrompt
    }
    editableFixedPrompt.value = latestPrompt
    savedFixedPrompt.value = latestPrompt
    hasUserEditedFixedPrompt.value = Boolean(latestPrompt)
    saveFormStateToStorage()
  }
}

const buildSelectedPromptContent = () => {
  let selectedContent = ''

  pinnedPrompts.value
    .filter(prompt => prompt.enabled)
    .forEach(prompt => {
      let content = prompt.content || ''
      if (prompt.fields) {
        prompt.fields.forEach(field => {
          const fieldValue = fieldValues.value[field.name] || ''
          content = content.replace(new RegExp(`\\$\\{${field.name}\\}`, 'g'), fieldValue)
        })
      }
      selectedContent += `${content}\n`
    })

  return selectedContent.trim()
}

const buildSystemPrompt = () => {
  const fixedContent = editableFixedPrompt.value.trim()

  // 自定义模式：使用用户输入的自定义提示词
  if (promptPanelMode.value === 'custom' && customPromptContent.value.trim()) {
    return [fixedContent, customPromptContent.value.trim()]
      .filter(Boolean)
      .join('\n\n')
  }

  // 模板选择模式：使用选中的模板
  const selectedContent = buildSelectedPromptContent()

  return [fixedContent, selectedContent ? `\n${selectedContent}` : '']
    .filter(Boolean)
    .join('\n\n')
}

const buildUserPrompt = () => {
  const generatorName = currentGenerator.value?.name || '创意生成'
  const sections: string[] = [``]

  if (additionalInfo.value.trim()) {
    sections.push(`补充信息：\n${additionalInfo.value.trim()}`)
  }

  if (selectedBook.value) {
    sections.push(`关联作品：${selectedBook.value.title}`)
  }

  return sections.join('\n\n').trim()
}

const buildFullPromptPreview = () => {
  const sections = [`${buildSystemPrompt() || ''}`]
  const userPrompt = buildUserPrompt()

  if (userPrompt) {
    sections.push(`${userPrompt}`)
  }

  if (relatedContextItems.value.length > 0) {
    const relatedText = relatedContextItems.value
      .map(item => `[${item.type}] ${item.title}\n${item.content}`)
      .join('\n\n')
    sections.push(`【关联参考内容】\n${relatedText}`)
  }

  return sections.join('\n\n')
}

const fetchBooks = async () => {
  try {
    const res = await bookAPI.getAll()
    books.value = res.success && res.data ? res.data : []
  } catch (error) {
    books.value = []
  }
}

const fetchMemos = async () => {
  try {
    const res = await memoAPI.getAll()
    memos.value = res.success && res.data ? res.data : []
  } catch (error) {
    memos.value = []
  }
}

const fetchBookContext = async (bookId: number) => {
  loadingBookContext.value = true

  try {
    const [chapterRes, characterRes] = await Promise.all([
      chapterAPI.getByBook(bookId),
      characterAPI.getByBook(bookId)
    ])

    availableChapters.value = chapterRes.success && chapterRes.data
      ? chapterRes.data.filter(chapter => chapter.type === 'chapter')
      : []
    availableCharacters.value = characterRes.success && characterRes.data ? characterRes.data : []
  } catch (error) {
    availableChapters.value = []
    availableCharacters.value = []
    ElMessage.warning('关联作品上下文加载失败，请稍后重试')
  } finally {
    loadingBookContext.value = false
  }
}

const resetRelatedSelections = () => {
  selectedBookId.value = null
  selectedChapterIds.value = []
  selectedMemoId.value = null
  selectedCharacterIds.value = []
  availableChapters.value = []
  availableCharacters.value = []
}

const getRelatedTypeLabel = (type: RelatedContent['type']) => {
  switch (type) {
    case 'book':
      return '作品'
    case 'chapter':
      return '章节'
    case 'memo':
      return '备忘录'
    case 'character':
      return '角色卡'
    default:
      return type
  }
}

const openGenerator = (generator: GeneratorItem) => {
  currentGenerator.value = generator
  dialogStep.value = 'config'
  showFollowUpPanel.value = false
  selectedPrompt.value = undefined
  generatedContent.value = ''
  resetRelatedSelections()

  isLoadingPinnedPrompts.value = true
  pinnedPrompts.value = loadPinnedPromptsFromStorage(generator.id)

  const savedCardPacks = loadGeneratorCardPacksFromStorage(generator.id)
  if (savedCardPacks.length > 0) {
    selectedGeneratorCardPacks.value = savedCardPacks
  } else {
    const mappedCategory = GENERATOR_TO_CATEGORY_MAP[generator.name]
    selectedGeneratorCardPacks.value = mappedCategory ? [mappedCategory] : []
  }

  const savedFormState = loadFormStateFromStorage(generator.id)
  if (savedFormState) {
    const savedModelId = savedFormState.selectedModel
    selectedModel.value = (savedModelId && models.value.some(m => m.id === savedModelId)) ? savedModelId : getDefaultModelId()
    inputParams.value = savedFormState.inputParams || ''
    additionalInfo.value = savedFormState.additionalInfo || ''
    selectedBookId.value = savedFormState.selectedBookId
    selectedChapterIds.value = savedFormState.selectedChapterIds || []
    selectedMemoId.value = savedFormState.selectedMemoId
    selectedCharacterIds.value = savedFormState.selectedCharacterIds || []
    fieldValues.value = savedFormState.fieldValues || {}
    editableFixedPrompt.value = savedFormState.editableFixedPrompt || ''
    savedFixedPrompt.value = savedFormState.savedFixedPrompt || ''
    hasUserEditedFixedPrompt.value = savedFormState.hasUserEditedFixedPrompt || false

    if (selectedBookId.value) {
      fetchBookContext(selectedBookId.value)
    }
  } else {
    selectedModel.value = getDefaultModelId()
    inputParams.value = ''
    additionalInfo.value = ''
    fieldValues.value = {}

    if (generator.core_prompt) {
      editableFixedPrompt.value = generator.core_prompt
      savedFixedPrompt.value = generator.core_prompt
      hasUserEditedFixedPrompt.value = true
    } else {
      const defaultPrompt = `你是一个专业的${generator.name}助手，请根据用户的输入生成相关内容。`
      editableFixedPrompt.value = defaultPrompt
      savedFixedPrompt.value = defaultPrompt
      hasUserEditedFixedPrompt.value = false
    }
  }

  setTimeout(() => {
    isLoadingPinnedPrompts.value = false
    dialogVisible.value = true
  }, 0)
}

// 打开提示词编辑弹窗
const openPromptEditDialog = () => {
  const prompt = selectedPromptInfo.value
  if (prompt) {
    editingPrompt.value = {
      id: prompt.id,
      name: prompt.name,
      content: prompt.content,
      category: prompt.category
    }
    promptEditDialogVisible.value = true
  }
}

// 显示固定提示词弹窗（方案 3 格式）
const showFixedPrompt = () => {
  fixedPromptContent.value = buildFullPromptPreview()
  fixedPromptDialogVisible.value = true
}

// 保存固定提示词
const saveFixedPrompt = () => {
  savedFixedPrompt.value = editableFixedPrompt.value
  hasUserEditedFixedPrompt.value = true
  ElMessage.success('固定提示词已保存')
}

watch(pinnedPrompts, (newVal, oldVal) => {
  if (!isLoadingPinnedPrompts.value) {
    savePinnedPromptsToStorage()
    if (newVal.length !== oldVal?.length) {
      fieldValues.value = {}
    }
  }
}, { deep: true })

watch(selectedGeneratorCardPacks, () => {
  saveGeneratorCardPacksToStorage()
}, { deep: true })

watch(editableFixedPrompt, (newVal, oldVal) => {
  if (newVal && newVal !== oldVal && !hasUserEditedFixedPrompt.value) {
    hasUserEditedFixedPrompt.value = true
  }
})

watch(selectedBookId, async (newBookId, oldBookId) => {
  if (newBookId === oldBookId) return

  selectedChapterIds.value = []
  selectedCharacterIds.value = []
  availableChapters.value = []
  availableCharacters.value = []

  if (newBookId) {
    await fetchBookContext(newBookId)
  }
})

watch(() => allFields.value.length, (newLen, oldLen) => {
  if (newLen !== oldLen && !isLoadingPinnedPrompts.value) {
    const newKeys = new Set(allFields.value.map(f => f.name))
    Object.keys(fieldValues.value).forEach(key => {
      if (!newKeys.has(key)) {
        delete fieldValues.value[key]
      }
    })
  }
})

watch(dialogVisible, (newVal) => {
  if (!newVal && currentGenerator.value) {
    saveFormStateToStorage()
  }
})

const generateContent = async () => {
  if (!canGenerate.value) return

  generating.value = true
  previewContent.value = ''
  isEditingPreview.value = false
  showFollowUpPanel.value = false
  // 直接在主弹窗内切换到结果视图，不再叠加独立预览弹窗
  dialogStep.value = 'result'

  abortController.value = new AbortController()
  
  try {
    const selectedModelData = models.value.find(m => m.id === selectedModel.value)
    
    if (!selectedModelData) {
      ElMessage.error('请选择模型')
      generating.value = false
      return
    }
    
    const generatorName = currentGenerator.value?.name || ''
    const systemPrompt = buildSystemPrompt()
    const userPrompt = buildUserPrompt()
    const relatedContent = relatedContextItems.value
    
    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        messages: [
          { role: 'user', content: userPrompt }
        ],
        configId: selectedModelData.id,
        systemPrompts: systemPrompt ? [systemPrompt] : [],
        relatedContent
      }),
      signal: abortController.value.signal
    })
    
    if (!response.ok) {
      throw new Error('AI调用失败')
    }
    
    const reader = response.body?.getReader()
    const decoder = new TextDecoder()
    let aiResponse = ''
    
    while (reader) {
      const { done, value } = await reader.read()
      if (done) break
      
      const text = decoder.decode(value)
      const textLines = text.split('\n')
      
      for (const line of textLines) {
        if (line.startsWith('data: ')) {
          const data = line.substring(6)
          if (data === '[DONE]') continue
          
          try {
            const json = JSON.parse(data)
            
            if (json.error) {
              ElMessage.error(`AI调用失败: ${json.error}`)
              throw new Error(json.error)
            }
            
            if (json.content) {
              aiResponse += json.content
              previewContent.value = aiResponse
              generatedContent.value = aiResponse
            }
          } catch (e) {
            console.error('解析流式数据失败:', e)
          }
        }
      }
    }
    
    if (aiResponse) {
      ElMessage.success('生成成功！')
      
      const historyRecord: HistoryRecord = {
        id: Date.now(),
        generatorName: generatorName,
        generatorIcon: currentGenerator.value?.icon,
        fixedPrompt: editableFixedPrompt.value,
        selectedPrompts: pinnedPrompts.value.filter(p => p.enabled).map(p => ({
          name: p.name,
          content: p.content || ''
        })),
        inputParams: inputParams.value,
        additionalInfo: additionalInfo.value,
        fieldValues: { ...fieldValues.value },
        relatedContext: relatedContent.map(item => ({
          type: item.type,
          title: item.title,
          content: item.content
        })),
        messages: [
          {
            role: 'system',
            content: systemPrompt,
            timestamp: Date.now()
          },
          {
            role: 'user',
            content: userPrompt,
            timestamp: Date.now()
          },
          {
            role: 'assistant',
            content: aiResponse,
            timestamp: Date.now()
          }
        ],
        previewContent: aiResponse,
        timestamp: Date.now()
      }
      
      saveHistoryRecord(historyRecord)
      currentConversationMessages.value = historyRecord.messages
    } else {
      ElMessage.error('生成内容为空')
    }
  } catch (error: any) {
    if (error.name === 'AbortError') {
      ElMessage.info('已停止生成')
    } else {
      console.error('生成内容失败:', error)
      ElMessage.error('生成失败，请重试')
    }
  } finally {
    generating.value = false
    abortController.value = null
  }
}

const stopGenerate = () => {
  if (abortController.value) {
    abortController.value.abort()
  }
}

const copyContent = () => {
  navigator.clipboard.writeText(generatedContent.value)
  ElMessage.success('已复制到剪贴板')
}

// 追问对话相关
const openFollowUp = () => {
  if (!currentConversationMessages.value.length) {
    ElMessage.warning('没有可追问的对话内容')
    return
  }
  followUpMessages.value = [...currentConversationMessages.value]
  followUpUserInput.value = ''
  followUpGenerating.value = false
  followUpHistoryExpanded.value = false
  followUpSelectedModelId.value = selectedModel.value
  // 追问改为主弹窗结果视图内的内联面板
  showFollowUpPanel.value = true

  nextTick(() => {
    scrollFollowUpToBottom()
  })
}

const scrollFollowUpToBottom = () => {
  if (followUpMessagesContainer.value) {
    followUpMessagesContainer.value.scrollTop = followUpMessagesContainer.value.scrollHeight
  }
}

const sendFollowUpMessage = async () => {
  if (!followUpUserInput.value.trim() || followUpGenerating.value) return

  const userMessage: HistoryMessage = {
    role: 'user',
    content: followUpUserInput.value.trim(),
    timestamp: Date.now()
  }

  followUpMessages.value.push(userMessage)
  followUpHistoryExpanded.value = true
  const currentInput = followUpUserInput.value
  followUpUserInput.value = ''
  followUpGenerating.value = true

  nextTick(() => {
    scrollFollowUpToBottom()
  })

  try {
    const selectedModelData = models.value.find(m => m.id === followUpSelectedModelId.value)
    if (!selectedModelData) {
      ElMessage.error('请选择模型')
      followUpGenerating.value = false
      return
    }

    const validRoles = ['system', 'user', 'assistant']
    const messages = followUpMessages.value
      .filter(msg => validRoles.includes(msg.role))
      .map(msg => ({
        role: msg.role,
        content: msg.content
      }))

    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        messages,
        configId: selectedModelData.id
      })
    })

    if (!response.ok) {
      throw new Error('请求失败')
    }

    const reader = response.body?.getReader()
    const decoder = new TextDecoder()
    let assistantMessage = ''

    const assistantMsg: HistoryMessage = {
      role: 'assistant',
      content: '',
      timestamp: Date.now()
    }
    followUpMessages.value.push(assistantMsg)
    const reactiveAssistantMsg = followUpMessages.value[followUpMessages.value.length - 1]

    while (reader) {
      const { done, value } = await reader.read()
      if (done) break

      const chunk = decoder.decode(value)
      const lines = chunk.split('\n')

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.slice(6)
          if (data === '[DONE]') continue

          try {
            const parsed = JSON.parse(data)
            if (parsed.content) {
              assistantMessage += parsed.content
              reactiveAssistantMsg.content = assistantMessage
              nextTick(() => {
                scrollFollowUpToBottom()
              })
            } else if (parsed.error) {
              throw new Error(parsed.error)
            }
          } catch (e) {
            if (e instanceof SyntaxError) continue
            throw e
          }
        }
      }
    }

    ElMessage.success('对话完成')
  } catch (error: any) {
    ElMessage.error(error.message || '发送失败')
    followUpMessages.value.pop()
    followUpMessages.value.pop()
  } finally {
    followUpGenerating.value = false
  }
}

// 保存提示词修改（由 FixedPromptDialog 触发）
const onPromptSaved = async (prompt: { id: number; name: string; content: string; category: string }) => {
  try {
    const response = await promptAPI.update(prompt.id, {
      name: prompt.name,
      content: prompt.content,
      category: prompt.category
    })

    if (response.success) {
      // 更新本地提示词列表
      const index = prompts.value.findIndex(p => p.id === prompt.id)
      if (index !== -1) {
        prompts.value[index] = {
          id: prompt.id,
          name: prompt.name,
          category: prompt.category,
          content: prompt.content
        }
      }

      ElMessage.success('提示词修改成功')
      promptEditDialogVisible.value = false
    }
  } catch (error) {
    console.error('保存提示词失败:', error)
    ElMessage.error('保存失败，请重试')
  }
}

// 生成器管理方法
const openGeneratorManager = async () => {
  generatorManagerVisible.value = true
  // 打开时刷新列表（子组件负责默认选中第一个生成器）
  await fetchGenerators()
  const defaults = loadDefaultGenerators()
  allGenerators.value = [...defaults, ...customGenerators.value.map(g => ({ ...g, isDefault: false }))]
}

const fetchGenerators = async () => {
  try {
    const res = await generatorAPI.getAll()
    if (res?.data) {
      customGenerators.value = res.data
    }
  } catch (error) {
    customGenerators.value = []
  }
}

// 子组件触发保存：新建 / 编辑 统一入口
const onManagerSave = async (payload: { generator: Partial<GeneratorItem>; isNew: boolean }) => {
  const selectedGenerator = payload.generator
  const isNewGenerator = payload.isNew
  if (!selectedGenerator?.name?.trim()) {
    ElMessage.warning('请输入生成器名称')
    return
  }

  try {
    if (isNewGenerator) {
      const res = await generatorAPI.create({
        name: selectedGenerator.name,
        description: selectedGenerator.description,
        icon: selectedGenerator.icon || 'Lightning',
        core_prompt: selectedGenerator.core_prompt,
        remark: selectedGenerator.remark,
        order_num: selectedGenerator.order_num || 0
      })
      if (res.success) {
        const generatorName = selectedGenerator.name
        const generatorCorePrompt = selectedGenerator.core_prompt || ''
        
        try {
          const promptRes = await promptAPI.create({
            name: `${generatorName}默认提示词`,
            category: generatorName,
            content: generatorCorePrompt,
            description: `${generatorName}生成器的默认提示词模板`
          })
          if (promptRes?.data) {
            prompts.value.unshift({
              id: promptRes.data.id || Date.now(),
              name: `${generatorName}默认提示词`,
              category: generatorName,
              content: generatorCorePrompt,
              description: `${generatorName}生成器的默认提示词模板`,
              author: '系统',
              version: 'v1.0',
              created_at: new Date().toISOString(),
              usage_count: 0
            })
          }
        } catch (promptError) {
          console.warn('创建默认提示词失败:', promptError)
        }
        
        ElMessage.success('创建成功')
        saveExampleOverride((res.data as { id?: number } | null)?.id, selectedGenerator.example)
        await fetchGenerators()
        const defaults = loadDefaultGenerators()
        allGenerators.value = [...defaults, ...customGenerators.value.map(g => ({ ...g, isDefault: false }))]
      }
    } else if (selectedGenerator.isDefault) {
      // 保存默认生成器的修改到localStorage
      const defaults = loadDefaultGenerators()
      const index = defaults.findIndex((g: GeneratorItem) => g.id === selectedGenerator.id)
      if (index !== -1) {
        defaults[index] = { ...selectedGenerator } as GeneratorItem
        saveDefaultGenerators(defaults)
        syncGeneratorPromptState(defaults[index])
        saveExampleOverride(selectedGenerator.id, selectedGenerator.example)
        // 更新本地列表
        allGenerators.value = [...defaults, ...customGenerators.value.map(g => ({ ...g, isDefault: false }))]
        ElMessage.success('保存成功')
      }
    } else {
      // 保存自定义生成器到数据库
      const res = await generatorAPI.update(selectedGenerator.id!, {
        name: selectedGenerator.name,
        description: selectedGenerator.description,
        icon: selectedGenerator.icon,
        core_prompt: selectedGenerator.core_prompt,
        remark: selectedGenerator.remark,
        order_num: selectedGenerator.order_num
      })
      if (res.success) {
        ElMessage.success('保存成功')
        saveExampleOverride(selectedGenerator.id, selectedGenerator.example)
        await fetchGenerators()
        syncGeneratorPromptState({
          id: selectedGenerator.id!,
          core_prompt: selectedGenerator.core_prompt
        })
        // 更新本地列表
        const defaults = loadDefaultGenerators()
        allGenerators.value = [...defaults, ...customGenerators.value.map(g => ({ ...g, isDefault: false }))]
      }
    }
  } catch (error) {
    ElMessage.error('保存失败，请重试')
  }
}

// 子组件触发删除
const onManagerDelete = async (gen: Partial<GeneratorItem>) => {
  if (gen.id == null) return
  try {
    await ElMessageBox.confirm('确定要删除这个生成器吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    if (gen.isDefault) {
      // 删除默认生成器：从localStorage中移除
      const defaults = loadDefaultGenerators()
      const index = defaults.findIndex((g: GeneratorItem) => g.id === gen.id)
      if (index !== -1) {
        defaults.splice(index, 1)
        saveDefaultGenerators(defaults)
        allGenerators.value = [...defaults, ...customGenerators.value.map(g => ({ ...g, isDefault: false }))]
        ElMessage.success('删除成功')
      }
    } else {
      // 删除自定义生成器：从数据库中删除
      const res = await generatorAPI.delete(gen.id)
      if (res.success) {
        ElMessage.success('删除成功')
        await fetchGenerators()
        const defaults = loadDefaultGenerators()
        allGenerators.value = [...defaults, ...customGenerators.value.map(g => ({ ...g, isDefault: false }))]
      }
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败，请重试')
    }
  }
}

// 子组件触发上移/下移
const onManagerMove = (kind: 'up' | 'down', index: number) => {
  if (kind === 'up') {
    if (index <= 0) return
    const generators = [...allGenerators.value]
    const item = generators[index]
    generators[index] = generators[index - 1]
    generators[index - 1] = item
    generators.forEach((g, i) => { g.order_num = i })
    allGenerators.value = generators
  } else {
    if (index >= allGenerators.value.length - 1) return
    const generators = [...allGenerators.value]
    const item = generators[index]
    generators[index] = generators[index + 1]
    generators[index + 1] = item
    generators.forEach((g, i) => { g.order_num = i })
    allGenerators.value = generators
  }
  saveGeneratorsOrder()
}

const saveGeneratorsOrder = () => {
  const defaults = allGenerators.value.filter(g => g.isDefault)
  saveDefaultGenerators(defaults as GeneratorItem[])
}

const createDefaultGenerators = (): GeneratorItem[] =>
  defaultGeneratorConfigs.map((g, i) => ({ ...g, isDefault: true, order_num: i }))

// 从localStorage加载默认生成器配置
const loadDefaultGenerators = () => {
  const stored = localStorage.getItem('default_generators')
  const version = localStorage.getItem('default_generators_version')

  if (stored && version === DEFAULT_GENERATORS_VERSION) {
    try {
      return JSON.parse(stored)
    } catch (error) {
      console.error('Failed to parse default generators:', error)
    }
  }

  const defaults = createDefaultGenerators()
  saveDefaultGenerators(defaults)
  return defaults
}

// 保存默认生成器配置到localStorage
const saveDefaultGenerators = (generators: GeneratorItem[]) => {
  localStorage.setItem('default_generators', JSON.stringify(generators))
  localStorage.setItem('default_generators_version', DEFAULT_GENERATORS_VERSION)
}

// 加载提示词列表
onMounted(async () => {
  // 加载默认生成器（从localStorage）
  const defaults = loadDefaultGenerators()
  
  // 加载自定义生成器（从数据库）
  await fetchGenerators()
  
  // 合并所有生成器
  allGenerators.value = [...defaults, ...customGenerators.value.map(g => ({ ...g, isDefault: false }))]
  
  try {
    const res = await promptAPI.getAll()
    if (res?.data) {
      prompts.value = res.data.map((p: { id: number; name: string; category: string; content: string; fields?: any[] }) => ({
        id: p.id,
        name: p.name,
        category: p.category || '通用',
        content: p.content,
        fields: p.fields
      }))
    }
  } catch (error) {
    // 使用默认数据
    prompts.value = [
      { id: 1, name: '角色设定生成', category: '角色', content: '', author: '系统', version: 'v1.0', description: '用于生成角色设定的提示词' },
      { id: 2, name: '剧情大纲生成', category: '剧情', content: '', author: '系统', version: 'v1.0', description: '用于生成剧情大纲的提示词' },
      { id: 3, name: '对话润色', category: '对话', content: '', author: '系统', version: 'v1.0', description: '用于润色对话内容的提示词' },
      { id: 4, name: '场景描写', category: '场景', content: '', author: '系统', version: 'v1.0', description: '用于生成场景描写的提示词' },
      { id: 5, name: '情节转折', category: '剧情', content: '', author: '系统', version: 'v1.0', description: '用于生成情节转折的提示词' }
    ]
  }
  
  // 加载模型列表
  await fetchModels()
  await Promise.all([fetchBooks(), fetchMemos()])
  
  // 加载历史记录
  loadHistoryRecords()
})

// 历史记录相关函数
const loadHistoryRecords = () => {
  const stored = localStorage.getItem('creative-history')
  if (stored) {
    try {
      const records = JSON.parse(stored)
      historyRecords.value = records
    } catch (e) {
      console.error('Failed to parse history records:', e)
      historyRecords.value = []
    }
  }
}

const saveHistoryRecord = (record: HistoryRecord) => {
  const allRecords = JSON.parse(localStorage.getItem('creative-history') || '[]')
  allRecords.push(record)
  localStorage.setItem('creative-history', JSON.stringify(allRecords))
  loadHistoryRecords()
}

const deleteHistoryRecord = (index: number) => {
  historyRecords.value.splice(index, 1)
  const allRecords = JSON.parse(localStorage.getItem('creative-history') || '[]')
  allRecords.splice(index, 1)
  localStorage.setItem('creative-history', JSON.stringify(allRecords))
}

const clearHistory = () => {
  historyRecords.value = []
  localStorage.removeItem('creative-history')
  ElMessage.success('历史记录已清空')
}
</script>

