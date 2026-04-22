﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿<template>
  <div class="write-container">
    <!-- 顶部工具栏 -->
    <div class="toolbar" :style="{ height: toolbarHeight + 'px' }">
      <div class="toolbar-leading">
        <button type="button" class="toolbar-round-btn" @click="goBackToBooks" title="返回书架">
          <el-icon><ArrowLeft /></el-icon>
        </button>
        <div class="toolbar-book-chip">
          <span class="toolbar-book-name">{{ currentBook?.title || '创作' }}</span>
          <span class="toolbar-book-mode">{{ currentMemo ? '备忘录' : '正文写作' }}</span>
        </div>
      </div>
      <div class="toolbar-actions">
        <el-button class="ai-write-btn" @click="toggleChatPanel">
          <el-icon><ChatDotSquare /></el-icon>
          AI 对话
        </el-button>
        <el-button class="ai-write-btn ai-write-btn-2" @click="toggleChatPanel2">
          <el-icon><ChatDotSquare /></el-icon>
          AI 写作
        </el-button>
        <el-button class="character-btn" @click="goToCharacterLibrary">
          <el-icon><User /></el-icon>
          角色库
        </el-button>
        <el-button class="analysis-btn" @click="goToBookAnalysis">
          <el-icon><Notebook /></el-icon>
          拆书库
        </el-button>
        <el-button class="graph-btn" @click="toggleGraphPanel">
          <el-icon><Connection /></el-icon>
          知识图谱
        </el-button>
        <el-tooltip content="全局备忘录" placement="bottom">
          <el-button class="toolbar-icon-btn memo-btn" @click="showGlobalMemoDialog = true">
            <el-icon><EditPen /></el-icon>
          </el-button>
        </el-tooltip>
        <el-button class="history-btn" @click="showHistoryDialog = true" title="历史记录">
          <el-icon><Document /></el-icon>
          历史记录
        </el-button>
        <div class="toolbar-status" v-if="activeContentUpdatedAt">
           {{ formatTime(activeContentUpdatedAt) }}
        </div>
      </div>
    </div>

    <!-- 顶部工具栏拖拽手柄 -->
    <div 
      class="toolbar-resize-handle" 
      @mousedown="startToolbarResize"
      title="拖拽调整高度"
    >
      <div class="toolbar-resize-bar"></div>
    </div>

    <!-- 三栏布局 -->
    <div class="content-wrapper">
      <!-- 左侧：目录 -->
      <div class="left-panel" :style="{ width: leftPanelWidth + 'px' }">
        <div class="left-panel-header">
          <div class="left-panel-book-title">{{ currentBook?.title || '新建作品' }}</div>
        </div>
        <el-tabs v-model="catalogType">
          <el-tab-pane label="正文目录" name="chapters">
            <div class="catalog-header">
              <div class="catalog-controls">
                <el-button 
                  size="small" 
                  @click="handleCreateChapter" 
                  class="icon-btn"
                  title="新建章节"
                >
                  <el-icon><Plus /></el-icon>
                </el-button>
                <el-button 
                  size="small" 
                  @click="handleCreateVolume" 
                  class="icon-btn"
                  title="新建分卷"
                >
                  <el-icon><Folder /></el-icon>
                </el-button>
                <el-button 
                  size="small" 
                  @click="toggleChapterOrder"
                  class="icon-btn"
                  :title="isDescending ? '正序排列' : '倒序排列'"
                >
                  <el-icon>
                    <component :is="isDescending ? 'Switch' : 'Sort'" />
                  </el-icon>
                </el-button>
              </div>
            </div>
            <div class="catalog-list">
              <template v-if="volumesWithChapters.length === 0 && unchapteredChapters.length === 0">
                <div class="empty-catalog">暂无章节</div>
              </template>
              
              <!-- 未分卷的章节（放在最前面） -->
              <div v-if="unchapteredChapters.length > 0" class="volume-container">
                <div 
                  class="volume-item" 
                  @click="toggleVolume('unchaptered')"
                >
                  <el-icon class="volume-arrow" :class="{ expanded: expandedVolumeId === 'unchaptered' }">
                    <ArrowRight />
                  </el-icon>
                  <span class="volume-title" style="display: none;">未分卷章节</span>
                  <span class="volume-title">0 号分卷</span>
                </div>
                <div class="chapter-list" v-show="expandedVolumeId === 'unchaptered'">
                  <div
                    v-for="chapter in unchapteredChapters"
                    :key="chapter.id"
                    :class="['catalog-item', { active: currentChapter?.id === chapter.id }]"
                    @click="selectChapter(chapter)"
                  >
                    <div class="catalog-main">
                      <div class="catalog-title-row">
                        <span class="catalog-title">{{ chapter.title }}</span>
                        <span class="catalog-word-count">{{ getContentLength(chapter.content) }} 字</span>
                      </div>
                      <div class="catalog-meta-row">
                        <span class="catalog-meta">{{ formatTime(chapter.updated_at) }}</span>
                        <span class="catalog-meta">{{ chapter.summary?.trim() ? '有概要' : '未写概要' }}</span>
                      </div>
                    </div>
                    <div class="catalog-actions" @click.stop>
                      <el-tooltip :content="chapter.summary?.trim() ? '编辑概要' : '概要储存'" placement="top">
                        <button
                          type="button"
                          :class="['catalog-action-btn', 'summary-action-btn', { filled: !!chapter.summary?.trim() }]"
                          @click.stop="saveChapterSummary(chapter)"
                        >
                          <el-icon><Document /></el-icon>
                        </button>
                      </el-tooltip>
                      <el-tooltip content="删除章节" placement="top">
                        <button
                          type="button"
                          class="catalog-action-btn delete-action-btn"
                          @click.stop="deleteChapter(chapter)"
                        >
                          <el-icon><Delete /></el-icon>
                        </button>
                      </el-tooltip>
                    </div>
                  </div>
                </div>
              </div>
              
              <!-- 分卷列表 -->
              <div
                v-for="volume in volumesWithChapters"
                :key="volume.id"
                class="volume-container"
              >
                <div 
                  class="volume-item" 
                  @click="toggleVolume(volume.id)"
                >
                  <el-icon class="volume-arrow" :class="{ expanded: expandedVolumeId === volume.id }">
                    <ArrowRight />
                  </el-icon>
                  <span class="volume-title">{{ volume.title }}</span>
                  <div class="volume-actions" @click.stop>
                    <el-button 
                      size="small" 
                      text 
                      @click="editVolume(volume)"
                      v-if="expandedVolumeId === volume.id"
                    >
                      <el-icon><Edit /></el-icon>
                    </el-button>
                    <el-button 
                      size="small" 
                      text 
                      @click="deleteVolume(volume)"
                      v-if="expandedVolumeId === volume.id"
                    >
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                </div>
                <div class="chapter-list" v-show="expandedVolumeId === volume.id">
                  <div
                    v-for="chapter in volume.chapters"
                    :key="chapter.id"
                    :class="['catalog-item', { active: currentChapter?.id === chapter.id }]"
                    @click="selectChapter(chapter)"
                  >
                    <div class="catalog-main">
                      <div class="catalog-title-row">
                        <span class="catalog-title">{{ chapter.title }}</span>
                        <span class="catalog-word-count">{{ getContentLength(chapter.content) }} 字</span>
                      </div>
                      <div class="catalog-meta-row">
                        <span class="catalog-meta">{{ formatTime(chapter.updated_at) }}</span>
                        <span class="catalog-meta">{{ chapter.summary?.trim() ? '有概要' : '未写概要' }}</span>
                      </div>
                    </div>
                    <div class="catalog-actions" @click.stop>
                      <el-tooltip :content="chapter.summary?.trim() ? '编辑概要' : '概要储存'" placement="top">
                        <button
                          type="button"
                          :class="['catalog-action-btn', 'summary-action-btn', { filled: !!chapter.summary?.trim() }]"
                          @click.stop="saveChapterSummary(chapter)"
                        >
                          <el-icon><Document /></el-icon>
                        </button>
                      </el-tooltip>
                      <el-tooltip content="删除章节" placement="top">
                        <button
                          type="button"
                          class="catalog-action-btn delete-action-btn"
                          @click.stop="deleteChapter(chapter)"
                        >
                          <el-icon><Delete /></el-icon>
                        </button>
                      </el-tooltip>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </el-tab-pane>
          
        </el-tabs>
      </div>

      <!-- 左侧拖拽手柄 -->
      <div 
        class="resize-handle" 
        @mousedown="startResize('left', $event)"
        title="拖拽调整宽度"
      >
        <div class="resize-handle-bar"></div>
      </div>

      <!-- 中间：编辑器 -->
      <div class="center-panel" :style="{ width: centerPanelWidth + 'px' }">
        <div v-if="currentChapter" class="editor-container">
          <div class="editor-sheet">
            <div class="editor-header">
              <el-input
                v-model="currentChapter.title"
                placeholder="请输入章节标题"
                class="chapter-title"
                @blur="saveChapter"
              />
              <span class="word-count-inline">
                {{ getContentLength(currentChapter.content) }} 字
                <span v-if="selectedTextLength > 0" class="selected-count">
                  / 已选 {{ selectedTextLength }}
                </span>
              </span>
            </div>
            <SplitRichTextEditor
              ref="chapterEditorRef"
              v-model="currentChapter.content"
              placeholder="开始创作..."
              class="chapter-content"
              @blur="saveChapter"
            />
          </div>
        </div>
        <div v-else-if="currentMemo" class="editor-container">
          <div class="editor-sheet">
            <div class="editor-header">
              <el-input
                v-model="currentMemo.title"
                placeholder="请输入备忘录标题"
                class="chapter-title"
                @blur="saveMemo"
              />
              <span class="word-count-inline">
                {{ getContentLength(currentMemo.content) }} 字
                <span v-if="selectedTextLength > 0" class="selected-count">
                  / 已选 {{ selectedTextLength }}
                </span>
              </span>
            </div>
            <el-input
              ref="memoEditorRef"
              v-model="currentMemo.content"
              type="textarea"
              placeholder="记录你的想法..."
              class="chapter-content"
              :style="{ fontSize: fontSize + 'px' }"
              @blur="saveMemo"
              @select="handleTextSelect"
              @mouseup="handleMouseUp"
              @keyup="handleKeyUp"
            />
          </div>
        </div>
        <div v-else class="empty-state">
          <el-empty description="请选择或创建章节/备忘录" />
        </div>
      </div>

      <!-- 右侧拖拽手柄 -->
      <div 
        class="resize-handle" 
        @mousedown="startResize('right', $event)"
        title="拖拽调整宽度"
        v-if="showChatPanel"
      >
        <div class="resize-handle-bar"></div>
      </div>

      <!-- 右侧：AI 对话 -->
      <div class="right-panel" v-if="showChatPanel" :style="{ width: rightPanelWidth + 'px' }">
        <div class="chat-main">
          <div class="chat-header">
            <el-button type="primary" size="small" @click="createConversation">
              <el-icon><Plus /></el-icon>
              新建对话
            </el-button>
            <span class="chat-title">{{ currentConversation?.title || 'AI 助手' }}</span>
            <el-select
              v-model="selectedConfigId"
              placeholder="API 配置"
              size="small"
              style="width: 150px;"
            >
              <el-option
                v-for="config in apiConfigs"
                :key="config.id"
                :label="config.name"
                :value="config.id"
              />
            </el-select>
            <el-button 
              size="small" 
              @click="closeChatPanel"
              circle
              title="收起"
            >
              <el-icon><Close /></el-icon>
            </el-button>
          </div>

          <div v-if="currentConversation" class="chat-content">
            <div class="chat-messages" ref="chatMessagesRef">
              <div
                v-for="(msg, index) in chatMessages"
                :key="index"
                :class="['message', msg.role]"
              >
                <div class="message-avatar">
                  <el-icon v-if="msg.role === 'user'"><User /></el-icon>
                  <el-icon v-else><ChatDotRound /></el-icon>
                </div>
                <div class="message-wrapper">
                  <div v-if="editingMessageIndex !== index" class="message-content">
                    <MarkdownRenderer :content="msg.content" />
                  </div>
                  <div v-else class="edit-wrapper">
                    <el-input
                      v-model="editingContent"
                      type="textarea"
                      :rows="10"
                      :autosize="{ minRows: 10, maxRows: 20 }"
                    />
                    <div class="edit-actions">
                      <el-button size="small" @click="cancelEdit">取消</el-button>
                      <el-button size="small" type="primary" @click="saveEdit(index)">
                        保存 (Ctrl+Enter)
                      </el-button>
                    </div>
                  </div>
                  <div class="message-actions" v-if="editingMessageIndex !== index">
                    <el-button size="small" text @click="copyMessage(msg.content)">
                      <el-icon><DocumentCopy /></el-icon>
                    </el-button>
                    <el-button size="small" text @click="startEdit(index, msg.content)">
                      <el-icon><Edit /></el-icon>
                    </el-button>
                    <el-button size="small" text type="danger" @click="deleteMessage(index)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                </div>
              </div>
              <div v-if="sending" class="message assistant">
                <div class="message-avatar">
                  <el-icon><ChatDotRound /></el-icon>
                </div>
                <div class="message-content typing">思考中...</div>
              </div>
            </div>

            <div class="chat-input-area">
              <div class="chat-tools">
                <el-button size="small" @click="showRelateDialog = true">
                  <el-icon><Link /></el-icon>
                  关联内容
                  <el-badge 
                    v-if="relatedContent.length > 0" 
                    :value="relatedContent.length" 
                    class="item"
                    type="primary"
                  />
                </el-button>
                <el-button size="small" @click="promptSelectDialogVisible = true">
                  <el-icon><Document /></el-icon>
                  选择提示词
                  <el-badge 
                    v-if="selectedPrompts.length > 0" 
                    :value="selectedPrompts.length" 
                    class="item"
                    type="primary"
                  />
                </el-button>
              </div>
              <div class="input-wrapper">
                <el-input
                  v-model="userInput"
                  type="textarea"
                  :rows="3"
                  placeholder="输入消息... (Ctrl+Enter 发送)"
                  @keydown.ctrl.enter="sendMessage"
                  :disabled="sending"
                />
                <el-button 
                  type="primary" 
                  :loading="sending" 
                  @click="sendMessage"
                  :disabled="!userInput.trim()"
                >
                  <el-icon><Promotion /></el-icon>
                </el-button>
              </div>
            </div>
          </div>
          <div v-else class="empty-chat">
            <el-empty description="开始新对话" />
          </div>
        </div>
      </div>

      <!-- 右侧拖拽手柄2 -->
      <div 
        class="resize-handle" 
        @mousedown="startResize('right2', $event)"
        title="拖拽调整宽度"
        v-if="showChatPanel2"
      >
        <div class="resize-handle-bar"></div>
      </div>

      <!-- 右侧：AI 写作 2 - 创意区 -->
      <div class="right-panel right-panel-2" v-if="showChatPanel2" :style="{ width: rightPanel2Width + 'px' }">
        <div class="creative-panel">
          <div class="creative-header">
            <span class="creative-title">AI 写作</span>
            <div class="creative-header-actions">
              <el-switch
                v-model="creative2AdvancedMode"
                size="small"
                inline-prompt
                active-text="高级"
                inactive-text="简洁"
                style="--el-switch-on-color: #00c9a7; --el-switch-off-color: #909399"
              />
              <el-button 
                size="small" 
                @click="closeChatPanel2"
                circle
                title="收起"
              >
                <el-icon><Close /></el-icon>
              </el-button>
            </div>
          </div>

          <div class="creative-content">
            <!-- 1. AI模型选择 -->
            <div class="creative-section">
              <div class="section-label">
                <el-icon><Monitor /></el-icon>
                AI 模型
              </div>
              <el-select
                v-model="creative2ConfigId"
                placeholder="选择API配置"
                style="width: 100%;"
              >
                <el-option
                  v-for="config in apiConfigs"
                  :key="config.id"
                  :label="config.name"
                  :value="config.id"
                />
              </el-select>
            </div>

            <!-- 高级功能区域 -->
            <template v-if="creative2AdvancedMode">
              <!-- 2. 故事背景 -->
              <div class="creative-section">
                <div class="section-label">
                  <el-icon><EditPen /></el-icon>
                  故事背景
                </div>
                <el-input
                  v-model="creative2StoryBackground"
                  maxlength="500"
                  show-word-limit
                  placeholder="输入故事背景设定..."
                />
              </div>

              <!-- 3. 关联角色卡 -->
              <div class="creative-section">
                <div class="section-label">
                  <el-icon><User /></el-icon>
                  关联角色卡（可选）
                </div>
                <el-select
                  v-model="creative2SelectedCharacterIds"
                  multiple
                  collapse-tags
                  collapse-tags-tooltip
                  clearable
                  filterable
                  placeholder="选择角色卡作为参考"
                  style="width: 100%;"
                >
                  <el-option
                    v-for="character in availableCharacters"
                    :key="character.id"
                    :label="character.name"
                    :value="character.id"
                  />
                </el-select>
                <div class="context-tip">
                  可以选择关联角色卡以提供上下文参考，不选择也可以直接生成内容。
                </div>
              </div>

              <!-- 4. 角色关系 -->
              <div class="creative-section">
                <div class="section-label">
                  <el-icon><Link /></el-icon>
                  角色关系
                </div>
                <el-input
                  v-model="creative2CharacterRelations"
                  maxlength="500"
                  show-word-limit
                  placeholder="输入本章涉及的角色及其关系..."
                />
              </div>

              <!-- 5. 本章剧情 -->
              <div class="creative-section">
                <div class="section-label">
                  <el-icon><ChatLineSquare /></el-icon>
                  本章剧情
                </div>
                <el-input
                  v-model="creative2ChapterPlot"
                  maxlength="3000"
                  show-word-limit
                  placeholder="输入本章剧情概要..."
                />
              </div>
            </template>

            <!-- 6. 写作风格 -->
            <div class="creative-section fixed-height-section">
              <div class="section-label">
                <el-icon><Document /></el-icon>
                写作风格
              </div>
              <div class="prompt-picker-row">
                <div class="prompt-picker-selected" @click="creative2PromptDialogVisible = true">
                  <span class="prompt-picker-name">{{ creative2PromptInfo?.name || '选择提示词' }}</span>
                  <el-tag v-if="creative2PromptInfo" size="small" type="info">{{ creative2PromptInfo.category }}</el-tag>
                  <el-icon class="prompt-picker-arrow"><ArrowRight /></el-icon>
                </div>
                <el-button
                  v-if="creative2PromptInfo"
                  type="primary"
                  circle
                  size="small"
                  class="view-intro-icon-btn"
                  @click="showCreative2PromptIntro"
                  title="查看介绍"
                >
                  <el-icon><View /></el-icon>
                </el-button>
              </div>
              <div v-if="creative2PromptInfo" class="prompt-picker-brief" v-html="creative2PromptInfo.description || '暂无简介'"></div>
            </div>

            <!-- 7. 写作要求（备用提示词） -->
            <div class="creative-section fixed-height-section">
              <div class="section-label">
                <el-icon><Document /></el-icon>
                写作要求
              </div>
              <div class="prompt-picker-row">
                <div class="prompt-picker-selected" @click="creative2SecondPromptDialogVisible = true">
                  <span class="prompt-picker-name">{{ creative2SecondPromptInfo?.name || '选择提示词' }}</span>
                  <el-tag v-if="creative2SecondPromptInfo" size="small" type="info">{{ creative2SecondPromptInfo.category }}</el-tag>
                  <el-icon class="prompt-picker-arrow"><ArrowRight /></el-icon>
                </div>
                <el-button
                  v-if="creative2SecondPromptInfo"
                  type="primary"
                  circle
                  size="small"
                  class="view-intro-icon-btn"
                  @click="showCreative2SecondPromptIntro"
                  title="查看介绍"
                >
                  <el-icon><View /></el-icon>
                </el-button>
              </div>
              <div v-if="creative2SecondPromptInfo" class="prompt-picker-brief" v-html="creative2SecondPromptInfo.description || '暂无简介'"></div>
            </div>

            <!-- 高级功能区域 -->
            <template v-if="creative2AdvancedMode">
              <!-- 8. 补充信息 -->
              <div class="creative-section">
                <div class="section-label">
                  <el-icon><EditPen /></el-icon>
                  补充信息
                </div>
                <el-input
                  v-model="creative2AdditionalInfo"
                  maxlength="500"
                  show-word-limit
                  placeholder="输入本次生成的额外要求、风格偏好或限制条件..."
                />
              </div>

              <!-- 9. 关联章节 -->
              <div class="creative-section">
                <div class="section-label">
                  <el-icon><Document /></el-icon>
                  关联章节（可选）
                </div>
                <el-select
                  v-model="creative2SelectedChapterId"
                  clearable
                  filterable
                  placeholder="选择章节内容作为参考"
                  style="width: 100%;"
                >
                  <el-option
                    v-for="chapter in chaptersList"
                    :key="chapter.id"
                    :label="chapter.title"
                    :value="chapter.id"
                  />
                </el-select>
                <div class="context-tip">
                  可以选择关联章节以提供上下文参考，不选择也可以直接生成内容。
                </div>
              </div>

              <!-- 10. 关联备忘录 -->
              <div class="creative-section">
                <div class="section-label">
                  <el-icon><Document /></el-icon>
                  关联备忘录（可选）
                </div>
                <el-select
                  v-model="creative2SelectedMemoId"
                  clearable
                  filterable
                  placeholder="选择备忘录内容作为参考"
                  style="width: 100%;"
                >
                  <el-option
                    v-for="memo in memos"
                    :key="memo.id"
                    :label="memo.title"
                    :value="memo.id"
                  />
                </el-select>
                <div class="context-tip">
                  可以选择关联备忘录以提供上下文参考，不选择也可以直接生成内容。
                </div>
              </div>
            </template>

            <!-- 开始生成按钮 -->
            <div class="creative-section">
              <el-button 
                type="primary" 
                :loading="creative2Generating"
                @click="handleCreative2Generate"
                :disabled="!creative2ConfigId"
                style="width: 100%;"
              >
                <el-icon><Lightning /></el-icon>
                开始生成
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 知识图谱弹窗 -->
    <el-dialog
      v-model="showGraphPanel"
      title="知识图谱"
      width="1200px"
      top="5vh"
      append-to-body
      destroy-on-close
      class="knowledge-graph-dialog"
    >
      <KnowledgeGraph
        ref="knowledgeGraphRef"
        :bookId="bookId"
        :apiConfigs="apiConfigs"
        :chapters="chaptersList"
        @analyze-start="onGraphAnalyzeStart"
        @analyze-end="onGraphAnalyzeEnd"
      />
    </el-dialog>

    <!-- AI 写作生成结果弹窗 -->
    <el-dialog
      v-model="creative2ResultDialogVisible"
      title="AI 写作生成结果"
      width="700px"
      align-center
      append-to-body
      class="creative2-result-dialog"
      :close-on-click-modal="false"
      :modal="false"
    >
      <div class="creative2-result-content">
        <div class="result-header">
          <div class="result-info">
            <el-icon class="result-icon"><Lightning /></el-icon>
            <span class="result-title">生成完成</span>
          </div>
          <el-button 
            size="small" 
            text 
            @click="copyCreative2Result"
            title="复制内容"
          >
            <el-icon><DocumentCopy /></el-icon>
            复制
          </el-button>
        </div>
        <div class="result-body">
          <el-scrollbar max-height="50vh">
            <div class="result-text">{{ creative2Result || '暂无生成内容' }}</div>
          </el-scrollbar>
        </div>
      </div>
      <template #footer>
        <div class="creative2-result-footer">
          <el-button @click="creative2ResultDialogVisible = false">关闭</el-button>
          <el-button 
            type="primary" 
            @click="applyCreative2Result"
            :disabled="!currentChapter || !creative2Result"
          >
            <el-icon><Download /></el-icon>
            应用到当前章节
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 关联内容对话框 -->
    <el-dialog
      v-model="showRelateDialog"
      title="关联内容"
      width="500px"
      align-center
      append-to-body
      class="relate-dialog"
    >
      <el-tabs v-model="relateType" class="relate-tabs">
        <el-tab-pane label="章节" name="chapters">
          <el-scrollbar class="relate-scrollbar" max-height="360px" always>
            <el-checkbox-group v-model="selectedChapters">
              <div v-for="chapter in chaptersList" :key="chapter.id" class="relate-option-item">
                <el-checkbox :label="chapter.id">{{ chapter.title }}</el-checkbox>
              </div>
            </el-checkbox-group>
          </el-scrollbar>
        </el-tab-pane>
        <el-tab-pane label="备忘录" name="memos">
          <el-scrollbar class="relate-scrollbar" max-height="360px" always>
            <el-checkbox-group v-model="selectedMemos">
              <div v-for="memo in memos" :key="memo.id" class="relate-option-item">
                <el-checkbox :label="memo.id">{{ memo.title }}</el-checkbox>
              </div>
            </el-checkbox-group>
          </el-scrollbar>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="showRelateDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmRelate">确定</el-button>
      </template>
    </el-dialog>

    <!-- 移动文件夹对话框 -->
    <el-dialog
      v-model="moveFolderDialogVisible"
      title="移动到文件夹"
      width="400px"
      class="move-folder-dialog"
    >
      <div class="move-folder-content">
        <p class="move-folder-hint">选择要移动到的目标文件夹：</p>
        <div class="folder-list">
          <button
            v-for="folder in globalMemoFolders"
            :key="folder"
            type="button"
            class="folder-option"
            :class="{ active: currentGlobalMemoFolder === folder }"
            @click="moveToFolder(folder)"
          >
            <el-icon><Folder /></el-icon>
            <span>{{ folder }}</span>
            <span class="folder-option-count">{{ getFolderMemoCount(folder) }}</span>
          </button>
        </div>
      </div>
      <template #footer>
        <el-button @click="moveFolderDialogVisible = false">取消</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="showGlobalMemoDialog"
      title="全局备忘录"
      width="1200px"
      top="5vh"
      draggable
      overflow
      append-to-body
      :modal="false"
      class="global-memo-dialog"
      destroy-on-close
    >
      <div class="global-memo-container">
        <div class="global-memo-sidebar" :class="{ collapsed: memoSidebarCollapsed }">
          <div class="memo-workspace-header">
            <div class="memo-workspace-title">
              <el-icon><Notebook /></el-icon>
              <span>备忘录</span>
            </div>
          </div>
          <div v-if="!memoSidebarCollapsed" class="sidebar-toolbar">
            <button
              type="button"
              class="sidebar-action sidebar-action-primary"
              @click="handleCreateGlobalMemoFolder"
            >
              <el-icon><Folder /></el-icon>
              <span>新建文件夹</span>
            </button>
            <button
              type="button"
              class="sidebar-action sidebar-action-accent"
              @click="handleCreateMemoInDialog('global')"
            >
              <el-icon><Plus /></el-icon>
              <span>新建</span>
            </button>
            <button
              type="button"
              class="sidebar-icon-btn"
              :class="{ active: memoBatchMode }"
              @click="toggleBatchMode"
            >
              <span>批量</span>
            </button>
            <button type="button" class="sidebar-icon-btn" @click="fetchMemos">
              <el-icon><Refresh /></el-icon>
            </button>
            <button type="button" class="sidebar-icon-btn" @click="memoSidebarCollapsed = true">
              <el-icon><ArrowLeft /></el-icon>
            </button>
          </div>
          <div v-else class="sidebar-collapsed-rail">
            <button type="button" class="sidebar-icon-btn" @click="memoSidebarCollapsed = false">
              <el-icon><ArrowRight /></el-icon>
            </button>
          </div>
          <div v-if="!memoSidebarCollapsed && memoBatchMode" class="batch-toolbar">
            <el-checkbox
              :model-value="hasSelectedMemos && filteredMemos.length > 0 && selectedMemoIds.length === filteredMemos.length"
              :indeterminate="hasSelectedMemos && selectedMemoIds.length < filteredMemos.length"
              @change="toggleSelectAll"
            >
              全选
            </el-checkbox>
            <el-button size="small" type="danger" :disabled="!hasSelectedMemos" @click="batchDelete">
              删除 ({{ selectedMemoIds.length }})
            </el-button>
            <el-button size="small" :disabled="!hasSelectedMemos" @click="batchTogglePin">
              置顶 ({{ selectedMemoIds.length }})
            </el-button>
          </div>
          <div v-if="!memoSidebarCollapsed" class="memo-search-box">
            <el-input
              v-model="memoSearchKeyword"
              placeholder="搜索备忘录..."
              clearable
              prefix-icon="Search"
              size="small"
            />
          </div>
          <div v-if="!memoSidebarCollapsed" class="sidebar-list">
            <div v-if="filteredMemos.length === 0" class="sidebar-empty">
              <el-icon><Document /></el-icon>
              <span>暂无备忘录</span>
            </div>
            <template v-else>
              <div class="memo-section-card">
                <div class="memo-tree-root">
                  <div
                    v-for="folder in globalMemoFolders"
                    :key="folder"
                    class="memo-tree-folder"
                  >
                    <button
                      type="button"
                      class="memo-tree-folder-header"
                      :class="{ active: currentGlobalMemoFolder === folder }"
                      @click="toggleFolder(folder)"
                    >
                      <div class="folder-header-content">
                        <el-icon class="folder-arrow">
                          <ArrowRight v-if="!folderExpandedStates[folder]" />
                          <ArrowDown v-else />
                        </el-icon>
                        <el-icon class="folder-icon"><Folder /></el-icon>
                        <span class="folder-name">{{ folder }}</span>
                        <span class="folder-count">{{ getFolderMemoCount(folder) }}</span>
                        <div class="folder-actions" @click.stop>
                          <button
                            type="button"
                            class="folder-action-btn"
                            title="新建备忘录"
                            @click="createMemoInFolder(folder)"
                          >
                            <el-icon><Plus /></el-icon>
                          </button>
                          <button
                            v-if="folder !== '默认'"
                            type="button"
                            class="folder-action-btn danger"
                            title="删除文件夹"
                            @click="deleteFolder(folder)"
                          >
                            <el-icon><Delete /></el-icon>
                          </button>
                        </div>
                      </div>
                    </button>
                    <div
                      v-show="folderExpandedStates[folder]"
                      class="memo-tree-children"
                    >
                      <div
                        v-for="memo in getMemosByFolder(folder)"
                        :key="memo.id"
                        :class="['memo-tree-item', { active: dialogSelectedMemo?.id === memo.id, 'batch-selected': selectedMemoIds.includes(memo.id) }]"
                        @click="handleMemoClick(memo)"
                      >
                        <div v-if="memoBatchMode" class="batch-checkbox">
                          <el-checkbox
                            :model-value="selectedMemoIds.includes(memo.id)"
                            @click.stop
                            @change="() => toggleMemoSelection(memo.id)"
                          />
                        </div>
                        <div class="item-content">
                          <div class="item-title-row">
                            <span class="item-title">{{ memo.title || '无标题' }}</span>
                            <div class="item-title-actions">
                              <button
                                type="button"
                                class="item-title-btn"
                                title="移动到文件夹"
                                @click.stop="openMoveFolderDialog(memo)"
                              >
                                <el-icon><Folder /></el-icon>
                              </button>
                              <button
                                type="button"
                                class="item-title-btn danger"
                                title="删除备忘录"
                                @click.stop="deleteMemoInList(memo)"
                              >
                                <el-icon><Delete /></el-icon>
                              </button>
                            </div>
                          </div>
                          <div class="item-tags-row" v-if="memo.tags">
                            <el-tag
                              v-for="tag in getMemoTags(memo.tags)"
                              :key="tag"
                              size="small"
                              class="item-tag"
                            >
                              {{ tag }}
                            </el-tag>
                          </div>
                          <div class="item-meta-row">
                            <span class="item-meta"></span>
                            <span class="item-meta-time"></span>
                          </div>
                        </div>
                        <div class="item-quick-actions">
                          <button
                            v-if="memo.is_pinned && !memoBatchMode"
                            type="button"
                            class="item-icon-btn pin"
                            @click.stop="toggleMemoPin(memo)"
                          >
                            <el-icon><Star /></el-icon>
                          </button>
                          <button
                            v-if="!memoBatchMode"
                            type="button"
                            class="item-icon-btn danger"
                            @click.stop="deleteMemo(memo)"
                          >
                            <el-icon><Delete /></el-icon>
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div class="memo-section-card">
                <div class="memo-tree-root">
                  <div class="memo-tree-folder">
                    <button
                      type="button"
                      class="memo-tree-folder-header"
                      :class="{ active: bookMemoExpanded }"
                      @click="bookMemoExpanded = !bookMemoExpanded"
                    >
                      <div class="folder-header-content">
                        <el-icon class="folder-arrow">
                          <ArrowRight v-if="!bookMemoExpanded" />
                          <ArrowDown v-else />
                        </el-icon>
                        <el-icon class="folder-icon"><Folder /></el-icon>
                        <span class="folder-name">本书备忘录</span>
                        <span class="folder-count">{{ bookMemoList.length }}</span>
                      </div>
                    </button>
                    <div
                      v-show="bookMemoExpanded"
                      class="memo-tree-children"
                    >
                      <div
                        v-for="memo in bookMemoList"
                        :key="memo.id"
                        :class="['memo-tree-item', { active: dialogSelectedMemo?.id === memo.id, 'batch-selected': selectedMemoIds.includes(memo.id) }]"
                        @click="handleMemoClick(memo)"
                      >
                        <div v-if="memoBatchMode" class="batch-checkbox">
                          <el-checkbox
                            :model-value="selectedMemoIds.includes(memo.id)"
                            @click.stop
                            @change="() => toggleMemoSelection(memo.id)"
                          />
                        </div>
                        <div class="item-content">
                          <div class="item-title-row">
                            <span class="item-title">{{ memo.title || '无标题' }}</span>
                            <button
                              type="button"
                              class="item-move-btn"
                              title="移动到文件夹"
                              @click.stop="openMoveFolderDialog(memo)"
                            >
                              <el-icon><Folder /></el-icon>
                            </button>
                          </div>
                          <div class="item-tags-row" v-if="memo.tags">
                            <el-tag
                              v-for="tag in getMemoTags(memo.tags)"
                              :key="tag"
                              size="small"
                              class="item-tag"
                            >
                              {{ tag }}
                            </el-tag>
                          </div>
                          <div class="item-meta-row">
                            <span class="item-meta">{{ getContentLength(memo.content) }} 字</span>
                            <span class="item-meta-time">{{ formatTime(memo.updated_at) }}</span>
                          </div>
                        </div>
                        <div class="item-quick-actions">
                          <button
                            v-if="memo.is_pinned && !memoBatchMode"
                            type="button"
                            class="item-icon-btn pin"
                            @click.stop="toggleMemoPin(memo)"
                          >
                            <el-icon><Star /></el-icon>
                          </button>
                          <button
                            v-if="!memoBatchMode"
                            type="button"
                            class="item-icon-btn danger"
                            @click.stop="deleteMemo(memo)"
                          >
                            <el-icon><Delete /></el-icon>
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </template>
          </div>
        </div>
        <div class="global-memo-content">
          <div v-if="dialogSelectedMemo" class="content-area">
            <div class="memo-content-header">
              <div class="memo-content-header-main">
                <el-input
                  v-model="dialogSelectedMemo.title"
                  placeholder="请输入备忘录标题"
                  class="memo-title-input"
                />
                <div class="memo-content-meta">
                  <span>创建 {{ formatExactTime(dialogSelectedMemo.created_at) }}</span>
                  <span>更新 {{ formatExactTime(dialogSelectedMemo.updated_at) }}</span>
                  <span>{{ getContentLength(dialogSelectedMemo.content) }} 字</span>
                </div>
              </div>
              <div class="memo-content-header-side">
                <span
                  class="memo-scope-badge"
                  :class="isBookMemo(dialogSelectedMemo) ? 'book' : 'global'"
                >
                  {{ isBookMemo(dialogSelectedMemo) ? '本书备忘录' : '全局备忘录' }}
                </span>
              </div>
            </div>
            <el-input
              v-model="dialogSelectedMemo.content"
              type="textarea"
              placeholder="在这里输入内容..."
              class="memo-content-input memo-content-input-large"
            />
            <div class="content-footer minimalist">
              <div class="footer-actions">
                <el-button @click="handleJumpToMemo(dialogSelectedMemo)">跳转编辑</el-button>
                <el-button type="primary" @click="saveDialogMemo">保存</el-button>
              </div>
            </div>
          </div>
          <div v-else class="content-empty">
            <el-icon><EditPen /></el-icon>
            <span>请选择或新建备忘录</span>
          </div>
        </div>
        <div
          class="dialog-resize-handle"
          @mousedown="startDialogResize"
        ></div>
      </div>
    </el-dialog>

    <!-- 重命名对话框 -->
    <el-dialog v-model="showRenameDialog" title="重命名对话" width="400px">
      <el-input v-model="renameTitle" placeholder="请输入新名称" />
      <template #footer>
        <el-button @click="showRenameDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmRename">确定</el-button>
      </template>
    </el-dialog>

    <!-- 提示词选择弹窗 -->
    <el-dialog
      v-model="promptSelectDialogVisible"
      title="选择提示词"
      width="600px"
      destroy-on-close
      class="prompt-select-dialog"
      align-center
      append-to-body
      center
      :modal="false"
    >
      <div class="prompt-select-content">
        <div class="prompt-select-toolbar">
          <div class="prompt-select-search-row">
            <el-input
              v-model="promptSearchKeyword"
              placeholder="搜索提示词..."
              prefix-icon="Search"
              clearable
              class="search-input"
            />
            <div class="prompt-select-count">共 {{ filteredPrompts.length }} 条</div>
          </div>
          <div class="category-tabs-wrapper">
            <el-button
              class="category-scroll-btn"
              size="small"
              circle
              @click="scrollCategory('left')"
              :disabled="!canScrollLeft"
            >
              <el-icon><ArrowLeft /></el-icon>
            </el-button>
            <div class="category-tabs" ref="categoryTabsRef" @scroll="handleCategoryScroll">
              <div class="category-tabs-inner">
                <div
                  v-for="category in promptCategories"
                  :key="category"
                  class="category-tab"
                  :class="{ active: selectedPromptCategory === category || (category === '全部' && selectedPromptCategory === 'all') }"
                  @click="selectedPromptCategory = category === '全部' ? 'all' : category"
                >
                  {{ category }}
                </div>
              </div>
            </div>
            <el-button
              class="category-scroll-btn"
              size="small"
              circle
              @click="scrollCategory('right')"
              :disabled="!canScrollRight"
            >
              <el-icon><ArrowRight /></el-icon>
            </el-button>
          </div>
        </div>
        <div class="prompt-select-list">
          <div
            v-for="prompt in filteredPrompts"
            :key="prompt.id"
            class="prompt-select-item"
            :class="{ selected: isSelectedPrompt(prompt.id) }"
            @click="togglePromptSelect(prompt)"
          >
            <el-checkbox
              :model-value="isSelectedPrompt(prompt.id)"
              @click.stop
              @change="togglePromptSelect(prompt)"
            />
            <div class="prompt-select-info">
              <div class="prompt-select-name">{{ prompt.name }}</div>
              <div class="prompt-select-desc">{{ prompt.content?.slice(0, 80) }}{{ prompt.content?.length > 80 ? '...' : '' }}</div>
            </div>
            <el-tag size="small" type="info" class="prompt-category-tag">{{ prompt.category }}</el-tag>
          </div>
          <div v-if="filteredPrompts.length === 0" class="empty-prompts">
            <el-icon :size="32"><Document /></el-icon>
            <p>暂无匹配的提示词</p>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 创意区提示词选择弹窗 -->
    <el-dialog
      v-model="creative2PromptDialogVisible"
      title="写作风格"
      width="600px"
      destroy-on-close
      class="prompt-select-dialog"
      align-center
      append-to-body
      :modal="false"
    >
      <div class="prompt-select-content">
        <div class="prompt-select-toolbar">
          <div class="prompt-select-search-row">
            <el-input
              v-model="creative2PromptSearchKeyword"
              placeholder="搜索提示词..."
              prefix-icon="Search"
              clearable
              class="search-input"
            />
            <div class="prompt-select-count">共 {{ creative2FilteredPrompts.length }} 条</div>
          </div>
        </div>
        <div class="prompt-select-list">
          <div
            v-for="prompt in creative2FilteredPrompts"
            :key="prompt.id"
            class="prompt-select-item"
            :class="{ selected: isCreative2PromptSelected(prompt.id) }"
            @click="toggleCreative2Prompt(prompt)"
          >
            <el-checkbox
              :model-value="isCreative2PromptSelected(prompt.id)"
              @click.stop
              @change="toggleCreative2Prompt(prompt)"
            />
            <div class="prompt-select-info">
              <div class="prompt-select-name">{{ prompt.name }}</div>
              <div class="prompt-select-desc">{{ prompt.content?.slice(0, 80) }}{{ prompt.content?.length > 80 ? '...' : '' }}</div>
            </div>
            <el-tag size="small" type="info" class="prompt-category-tag">{{ prompt.category }}</el-tag>
          </div>
          <div v-if="creative2FilteredPrompts.length === 0" class="empty-prompts">
            <el-icon :size="32"><Document /></el-icon>
            <p>暂无匹配的提示词</p>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="creative2PromptDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="creative2PromptDialogVisible = false">确定</el-button>
      </template>
    </el-dialog>

    <!-- 创意区第二个提示词选择弹窗 -->
    <el-dialog
      v-model="creative2SecondPromptDialogVisible"
      title="写作要求"
      width="600px"
      destroy-on-close
      class="prompt-select-dialog"
      align-center
      append-to-body
      :modal="false"
    >
      <div class="prompt-select-content">
        <div class="prompt-select-toolbar">
          <div class="prompt-select-search-row">
            <el-input
              v-model="creative2PromptSearchKeyword"
              placeholder="搜索提示词..."
              prefix-icon="Search"
              clearable
              class="search-input"
            />
            <div class="prompt-select-count">共 {{ creative2SecondFilteredPrompts.length }} 条</div>
          </div>
        </div>
        <div class="prompt-select-list">
          <div
            v-for="prompt in creative2SecondFilteredPrompts"
            :key="prompt.id"
            class="prompt-select-item"
            :class="{ selected: prompt.id === creative2SecondPromptId }"
            @click="selectSecondPrompt(prompt)"
          >
            <el-radio
              :model-value="prompt.id === creative2SecondPromptId"
              @click.stop
            />
            <div class="prompt-select-info">
              <div class="prompt-select-name">{{ prompt.name }}</div>
              <div class="prompt-select-desc">{{ prompt.content?.slice(0, 80) }}{{ prompt.content?.length > 80 ? '...' : '' }}</div>
            </div>
            <el-tag size="small" type="info" class="prompt-category-tag">{{ prompt.category }}</el-tag>
          </div>
          <div v-if="creative2SecondFilteredPrompts.length === 0" class="empty-prompts">
            <el-icon :size="32"><Document /></el-icon>
            <p>暂无匹配的提示词</p>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="creative2SecondPromptId = 0; creative2SecondPromptDialogVisible = false">清除选择</el-button>
        <el-button @click="creative2SecondPromptDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="creative2SecondPromptDialogVisible = false">确定</el-button>
      </template>
    </el-dialog>

    <!-- 选择提示词1介绍弹窗 -->
    <el-dialog
      v-model="creative2PromptIntroDialogVisible"
      title="提示词介绍"
      width="600px"
      destroy-on-close
      class="prompt-intro-dialog"
    >
      <div class="prompt-intro-content" v-if="creative2PromptInfo">
        <div class="prompt-intro-header">
          <h3 class="prompt-intro-name">{{ creative2PromptInfo.name }}</h3>
          <el-tag size="small" type="info">{{ creative2PromptInfo.category }}</el-tag>
        </div>
        <div class="prompt-intro-section">
          <div class="prompt-intro-label">简介</div>
          <div class="prompt-intro-description" v-html="creative2PromptInfo.description || '暂无简介'"></div>
        </div>
        <div class="prompt-intro-section">
          <div class="prompt-intro-label" @click="togglePromptContentCollapse">
            <span>提示词内容</span>
            <el-icon :class="['collapse-icon', { collapsed: isPromptContentCollapsed }]"><ArrowRight /></el-icon>
          </div>
          <el-collapse-transition>
            <div v-show="!isPromptContentCollapsed" class="prompt-intro-content-text">
              {{ creative2PromptInfo.content }}
            </div>
          </el-collapse-transition>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="creative2PromptIntroDialogVisible = false">确定</el-button>
      </template>
    </el-dialog>

    <!-- 选择提示词2介绍弹窗 -->
    <el-dialog
      v-model="creative2SecondPromptIntroDialogVisible"
      title="提示词介绍"
      width="600px"
      destroy-on-close
      class="prompt-intro-dialog"
    >
      <div class="prompt-intro-content" v-if="creative2SecondPromptInfo">
        <div class="prompt-intro-header">
          <h3 class="prompt-intro-name">{{ creative2SecondPromptInfo.name }}</h3>
          <el-tag size="small" type="info">{{ creative2SecondPromptInfo.category }}</el-tag>
        </div>
        <div class="prompt-intro-section">
          <div class="prompt-intro-label">简介</div>
          <div class="prompt-intro-description" v-html="creative2SecondPromptInfo.description || '暂无简介'"></div>
        </div>
        <div class="prompt-intro-section">
          <div class="prompt-intro-label" @click="togglePromptContentCollapse">
            <span>提示词内容</span>
            <el-icon :class="['collapse-icon', { collapsed: isPromptContentCollapsed }]"><ArrowRight /></el-icon>
          </div>
          <el-collapse-transition>
            <div v-show="!isPromptContentCollapsed" class="prompt-intro-content-text">
              {{ creative2SecondPromptInfo.content }}
            </div>
          </el-collapse-transition>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="creative2SecondPromptIntroDialogVisible = false">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分卷对话框 -->
    <el-dialog v-model="showVolumeDialog" :title="isEditVolume ? '编辑分卷' : '新建分卷'" width="400px">
      <el-input 
        v-model="volumeForm.title" 
        placeholder="请输入分卷名称" 
        maxlength="50"
        show-word-limit
      />
      <template #footer>
        <el-button @click="showVolumeDialog = false">取消</el-button>
        <el-button type="primary" @click="handleVolumeSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 历史记录对话框 -->
    <el-dialog v-model="showHistoryDialog" title="正文 AI 历史记录" width="800px" class="history-dialog-modal">
      <div class="history-dialog-content">
        <div v-if="loadingHistory" class="loading-state">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>加载历史记录中...</span>
        </div>
        <template v-else>
          <div class="history-header">
            <span class="history-count">共 {{ historyRecords.length }} 条记录</span>
            <div class="history-actions">
              <el-button
                type="danger"
                size="small"
                @click="clearHistoryRecords"
                :disabled="historyRecords.length === 0"
              >
                <el-icon><Delete /></el-icon>
                清空历史
              </el-button>
            </div>
          </div>

          <div v-if="historyRecords.length === 0" class="empty-state">
            <el-icon :size="48"><Document /></el-icon>
            <p>暂无历史记录</p>
            <p class="hint">每次 AI API 调用都会在这里留下一条连续对话记录</p>
          </div>

          <div v-else class="history-list">
            <div
              v-for="record in historyRecords"
              :key="record.id"
              class="history-item"
              @click="viewHistoryDetail(record)"
            >
              <div class="history-item-header">
                <div class="history-main">
                  <div class="history-icon-wrapper">
                    <el-icon class="history-icon"><ChatDotSquare /></el-icon>
                  </div>
                  <div class="history-info">
                    <div class="history-title-row">
                      <span class="history-title">{{ record.title }}</span>
                      <el-tag size="small" :type="getHistoryStatusType(record.status)" class="status-tag">
                        {{ getHistoryStatusLabel(record.status) }}
                      </el-tag>
                    </div>
                    <div class="history-meta">
                      <span class="meta-item">{{ record.sourceLabel }}</span>
                      <span class="meta-divider">·</span>
                      <span class="meta-item">{{ record.promptCount }} 条提示词</span>
                      <span class="meta-divider">·</span>
                      <span class="meta-item time">{{ formatTimestamp(record.timestamp) }}</span>
                    </div>
                  </div>
                </div>
                <div class="history-actions-mini">
                  <el-button
                    size="small"
                    text
                    @click.stop="viewHistoryDetail(record)"
                    title="查看详情"
                  >
                    <el-icon><View /></el-icon>
                  </el-button>
                  <el-button
                    size="small"
                    text
                    type="danger"
                    @click.stop="deleteHistoryRecord(record.id)"
                    title="删除"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
              <div class="history-preview-line">{{ record.previewContent || '无预览内容' }}</div>
            </div>
          </div>
        </template>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showHistoryDialog = false">关闭</el-button>
          <el-button type="primary" @click="loadAllHistory">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      v-model="historyDialogVisible"
      :title="selectedHistoryConversation?.title || '历史记录详情'"
      width="900px"
      class="history-detail-dialog"
      destroy-on-close
    >
      <div v-if="selectedHistoryConversation" class="history-detail-content">
        <div class="history-detail-meta">
          <el-tag size="small" type="info">{{ selectedHistoryConversation.sourceLabel }}</el-tag>
          <el-tag size="small" :type="getHistoryStatusType(selectedHistoryConversation.status)">
            {{ getHistoryStatusLabel(selectedHistoryConversation.status) }}
          </el-tag>
          <span class="history-detail-time">{{ formatTimestamp(selectedHistoryConversation.timestamp) }}</span>
        </div>

        <div class="conversation-messages">
          <div class="messages-container">
            <div
              v-for="(msg, msgIndex) in selectedHistoryConversation.messages"
              :key="msgIndex"
              class="message-bubble"
              :class="getHistoryMessageClass(msg.role)"
            >
              <div class="bubble-avatar">
                <el-icon v-if="msg.role === 'user'"><User /></el-icon>
                <el-icon v-else-if="msg.role === 'assistant'"><ChatDotRound /></el-icon>
                <el-icon v-else-if="msg.role === 'prompt'"><Star /></el-icon>
                <el-icon v-else><Monitor /></el-icon>
              </div>
              <div class="bubble-content">
                <div class="bubble-header">
                  <span class="bubble-role">{{ getHistoryRoleLabel(msg.role) }}</span>
                  <span class="bubble-time">{{ formatTimestamp(msg.timestamp) }}</span>
                </div>
                <div class="bubble-text">
                  <MarkdownRenderer :content="msg.content" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="historyDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, computed, watch } from 'vue'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import SplitRichTextEditor from '@/components/SplitRichTextEditor.vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBookStore } from '@/stores/book'
import { onUnmounted } from 'vue'
import { chapterAPI, memoAPI, promptAPI, configAPI, conversationAPI, volumeAPI, characterAPI } from '@/api'
import type { Chapter, Memo, Prompt, ApiConfig, ChatMessage, RelatedContent, Volume, Character, KnowledgeGraphData } from '@/types'
import KnowledgeGraph from '@/components/KnowledgeGraph.vue'
import { 
  ChatDotSquare, User, Notebook, ChatLineSquare, 
  Plus, MoreFilled, Delete, Folder, Switch, Sort, 
  ArrowRight, ArrowLeft, ArrowDown, Edit, EditPen, RefreshLeft, RefreshRight,
  Star, Grid, Refresh, Search, CopyDocument, Checked, Loading, View,
  Monitor, MagicStick, DocumentCopy, Document, Close, Promotion, ChatDotRound, Link, Lightning, Setting, Download, Connection
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const bookStore = useBookStore()

const bookId = parseInt(route.params.bookId as string)
const currentBook = computed(() => bookStore.currentBook)

// 目录相关
const catalogType = ref('chapters')
const chaptersList = ref<Chapter[]>([])
const currentChapter = ref<Chapter | null>(null)
const memos = ref<Memo[]>([])
const currentMemo = ref<Memo | null>(null)
const activeContentType = computed<'chapters' | 'memos'>(() => {
  return currentMemo.value && !currentChapter.value ? 'memos' : 'chapters'
})

// 分卷相关
const volumes = ref<Volume[]>([])
const expandedVolumeId = ref<number | string | null>('unchaptered')
const isDescending = ref(false)
const volumesWithChapters = computed(() => {
  return volumes.value.map(volume => {
    let chapters = chaptersList.value.filter(chapter => chapter.volume_id === volume.id)
    if (isDescending.value) {
      chapters.sort((a, b) => b.id - a.id)
    } else {
      chapters.sort((a, b) => a.id - b.id)
    }
    return {
      ...volume,
      chapters
    }
  })
})
const unchapteredChapters = computed(() => {
  let chapters = chaptersList.value.filter(chapter => !chapter.volume_id || chapter.volume_id === 0)
  if (isDescending.value) {
    chapters.sort((a, b) => b.id - a.id)
  } else {
    chapters.sort((a, b) => a.id - b.id)
  }
  return chapters
})

// API配置
const apiConfigs = ref<ApiConfig[]>([])
const selectedConfigId = ref<number>()

// 提示词
const prompts = ref<Prompt[]>([])
const selectedPrompts = ref<number[]>([])

// 提示词选择弹窗
const promptSelectDialogVisible = ref(false)
const promptSearchKeyword = ref('')
const promptCategories = ref(['全部'])
const selectedPromptCategory = ref('all')
const categoryTabsRef = ref<HTMLElement>()
const canScrollLeft = ref(false)
const canScrollRight = ref(false)

// 过滤后的提示词
const filteredPrompts = computed(() => {
  let result = prompts.value
  
  // 按分类过滤
  if (selectedPromptCategory.value !== 'all') {
    result = result.filter(prompt => prompt.category === selectedPromptCategory.value)
  }
  
  // 按关键词搜索
  if (promptSearchKeyword.value) {
    const keyword = promptSearchKeyword.value.toLowerCase()
    result = result.filter(prompt => 
      prompt.name.toLowerCase().includes(keyword) || 
      prompt.content.toLowerCase().includes(keyword)
    )
  }
  
  return result
})

// 对话相关
const conversations = ref<any[]>([])
const currentConversation = ref<any>(null)
const showChatPanel = ref(false)
const showChatPanel2 = ref(false)
const showGraphPanel = ref(false)
const showGlobalMemoDialog = ref(false)
const dialogSelectedMemo = ref<Memo | null>(null)
const showHistoryDialog = ref(false)
const memoSearchKeyword = ref('')
const memoBatchMode = ref(false)
const selectedMemoIds = ref<number[]>([])
const dialogSelectedMemoTags = ref<string[]>([])
const memoSidebarCollapsed = ref(false)
const bookMemoExpanded = ref(true)
const memoCreateScope = ref<'global' | 'book'>('book')
const currentGlobalMemoFolder = ref('默认')
const globalMemoCustomFolders = ref<string[]>([])
const folderExpandedStates = ref<Record<string, boolean>>({
  '默认': true
})
const moveFolderDialogVisible = ref(false)
const memoToMove = ref<Memo | null>(null)
const allMemoTags = computed(() => {
  const tagSet = new Set<string>()
  memos.value.forEach(memo => {
    if (memo.tags) {
      try {
        const tags = JSON.parse(memo.tags)
        if (Array.isArray(tags)) {
          tags.forEach((tag: string) => tagSet.add(tag))
        }
      } catch (e) {
        // 忽略解析错误
      }
    }
  })
  return Array.from(tagSet)
})

const filteredMemos = computed(() => {
  if (!memoSearchKeyword.value) {
    return memos.value
  }
  const keyword = memoSearchKeyword.value.toLowerCase()
  return memos.value.filter(memo => 
    memo.title.toLowerCase().includes(keyword) || 
    memo.content.toLowerCase().includes(keyword)
  )
})

const isBookMemo = (memo: Memo) => {
  return (memo.category || '').includes('本书')
}

const getGlobalMemoFolder = (memo: Memo) => {
  if (isBookMemo(memo)) return ''
  const category = (memo.category || '').trim()
  if (!category || category === '全局') return '默认'
  return category
}

const globalMemoList = computed(() => {
  return filteredMemos.value.filter(memo => !isBookMemo(memo))
})

const bookMemoList = computed(() => {
  return filteredMemos.value.filter(memo => isBookMemo(memo))
})

const activeContentUpdatedAt = computed(() => {
  if (currentChapter.value?.updated_at) return currentChapter.value.updated_at
  if (currentMemo.value?.updated_at) return currentMemo.value.updated_at
  return ''
})

const globalMemoFolders = computed(() => {
  const folderSet = new Set<string>(['默认'])
  globalMemoCustomFolders.value.forEach(folder => folderSet.add(folder))
  globalMemoList.value.forEach(memo => folderSet.add(getGlobalMemoFolder(memo)))
  return Array.from(folderSet)
})

const goBackToBooks = () => {
  router.push('/books')
}

const toggleFolder = (folder: string) => {
  if (currentGlobalMemoFolder.value !== folder) {
    currentGlobalMemoFolder.value = folder
    folderExpandedStates.value[folder] = true
  } else {
    folderExpandedStates.value[folder] = !folderExpandedStates.value[folder]
  }
}

const getMemosByFolder = (folder: string) => {
  return globalMemoList.value.filter(memo => getGlobalMemoFolder(memo) === folder)
}

const getFolderMemoCount = (folder: string) => {
  return getMemosByFolder(folder).length
}

const hasSelectedMemos = computed(() => selectedMemoIds.value.length > 0)
const historyDialogVisible = ref(false)

// 历史记录相关
type WriteHistoryRole = 'system' | 'user' | 'prompt' | 'assistant'
type WriteHistoryStatus = 'completed' | 'cancelled' | 'failed'

interface WriteHistoryMessage {
  role: WriteHistoryRole
  content: string
  timestamp: number
}

interface WriteHistoryRecord {
  id: string
  bookId: number
  title: string
  source: 'chat' | 'continue' | 'creative2'
  sourceLabel: string
  promptName: string
  promptCount: number
  status: WriteHistoryStatus
  previewContent: string
  timestamp: number
  messages: WriteHistoryMessage[]
}

const WRITE_HISTORY_STORAGE_KEY = 'write-ai-api-history'
const conversationsList = ref<any[]>([])
const historyRecords = ref<WriteHistoryRecord[]>([])
const loadingHistory = ref(false)
const selectedHistoryConversation = ref<WriteHistoryRecord | null>(null)
const expandedConversationId = ref<number | null>(null)

// AI 写作 2 - 创意区专用
const creative2ConfigId = ref<number>()
const creative2SelectedPrompts = ref<number[]>([])
const creative2Generating = ref(false)
const creative2Result = ref('')
const creative2ResultDialogVisible = ref(false)
const creative2AdvancedMode = ref(false)
const creative2PromptDialogVisible = ref(false)
const creative2SecondPromptDialogVisible = ref(false)
const creative2SecondPromptId = ref<number>(0)
const creative2PromptSearchKeyword = ref('')
const creative2PromptCategories = ref(['全部'])
const creative2SelectedPromptCategory = ref('all')
const creative2CategoryTabsRef = ref<HTMLElement>()
const creative2CanScrollLeft = ref(false)
const creative2CanScrollRight = ref(false)
const creative2UseFixedPrompt = ref(false)
const creative2FixedPromptId = ref<number>(0)
const creative2StoryBackground = ref('')
const creative2CharacterRelations = ref('')
const creative2ChapterPlot = ref('')
const creative2AdditionalInfo = ref('')
const creative2SelectedChapterId = ref<number>()
const creative2SelectedMemoId = ref<number>()
const creative2SelectedCharacterIds = ref<number[]>([])
const availableCharacters = ref<Character[]>([])

// 创意区字段配置
const creative2FieldValues = ref<Record<string, string>>({})

const creative2AllFields = computed(() => {
  const fields: Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; description: string; required: boolean }> = []
  
  const promptsToUse = creative2UseFixedPrompt.value && creative2FixedPromptId.value !== 0
    ? [creative2FixedPromptId.value]
    : creative2SelectedPrompts.value
  
  promptsToUse.forEach(promptId => {
    const prompt = prompts.value.find(p => p.id === promptId)
    if (prompt && prompt.fields && Array.isArray(prompt.fields)) {
      prompt.fields.forEach((field: any) => {
        if (!fields.find(f => f.name === field.name)) {
          fields.push({
            name: field.name,
            label: field.label,
            type: field.type || 'text',
            options: field.options || [],
            description: field.description || '',
            required: field.required !== false
          })
        }
      })
    }
  })
  
  return fields
})

const creative2HasFields = computed(() => {
  return creative2AllFields.value.length > 0
})

const creative2FilteredPrompts = computed(() => {
  let result = prompts.value.filter(prompt => prompt.category === '写作风格')
  
  if (creative2PromptSearchKeyword.value) {
    const keyword = creative2PromptSearchKeyword.value.toLowerCase()
    result = result.filter(prompt => 
      prompt.name.toLowerCase().includes(keyword) || 
      prompt.content.toLowerCase().includes(keyword)
    )
  }
  return result
})

const creative2SecondFilteredPrompts = computed(() => {
  let result = prompts.value.filter(prompt => prompt.category === '写作要求')
  
  if (creative2PromptSearchKeyword.value) {
    const keyword = creative2PromptSearchKeyword.value.toLowerCase()
    result = result.filter(prompt => 
      prompt.name.toLowerCase().includes(keyword) || 
      prompt.content.toLowerCase().includes(keyword)
    )
  }
  return result
})
const chatMessages = ref<ChatMessage[]>([])
const userInput = ref('')
const sending = ref(false)
const chatMessagesRef = ref<HTMLElement>()
const editingMessageIndex = ref<number | null>(null)
const editingContent = ref('')
const selectedTextLength = ref(0)
const fontSize = ref(16) // 默认字体大小

// 面板宽度调节
const leftPanelWidth = ref(280)
const centerPanelWidth = ref(0) // 动态计算
const rightPanelWidth = ref(500)
const rightPanel2Width = ref(500)
const isResizing = ref(false)
const resizeSide = ref<'left' | 'right' | 'right2'>('left')

// 工具栏高度调节
const toolbarHeight = ref(64)
const isResizingToolbar = ref(false)

// 续写功能
const continuePromptId = ref(0)
const continueConfigId = ref<number>()
const continuWriting = ref(false)
const continueAbortController = ref<AbortController | null>(null)
const chapterEditorRef = ref()
const memoEditorRef = ref()
const knowledgeGraphRef = ref()
const cursorPosition = ref(0)

// 关联内容
const showRelateDialog = ref(false)
const relateType = ref('chapters')
const selectedChapters = ref<number[]>([])
const selectedMemos = ref<number[]>([])
const relatedContent = ref<RelatedContent[]>([])

// 重命名
const showRenameDialog = ref(false)
const renameTitle = ref('')
const renamingConversation = ref<any>(null)

// 分卷相关
const showVolumeDialog = ref(false)
const isEditVolume = ref(false)
const volumeForm = ref({
  id: 0,
  title: ''
})

onMounted(async () => {
  loadGlobalMemoFolders()
  await bookStore.fetchBook(bookId)
  await fetchChapters()
  await fetchMemos()
  await fetchCharacters()
  await fetchPrompts()
  await fetchConfigs()
  await fetchConversations()
  await fetchVolumes()
  
  // 默认不显示 AI 对话面板，只有点击按钮才显示
  // 如果有章节，默认选择最后一个章节
  if (chaptersList.value.length > 0) {
    const lastChapter = chaptersList.value[chaptersList.value.length - 1]
    await selectChapter(lastChapter)
  }
  
  // 加载字体大小设置
  loadFontSize()
  
  // 初始化面板宽度
  updateCenterWidth()
  
  // 监听窗口大小变化
  window.addEventListener('resize', handleResizeWindow)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResizeWindow)
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
  document.removeEventListener('mousemove', handleToolbarResize)
  document.removeEventListener('mouseup', stopToolbarResize)
})

const fetchChapters = async () => {
  const res = await chapterAPI.getByBook(bookId)
  if (res.success && res.data) {
    chaptersList.value = res.data
  }
}

const fetchMemos = async () => {
  const res = await memoAPI.getAll()
  if (res.success && res.data) {
    memos.value = res.data
  }
}

const fetchCharacters = async () => {
  const res = await characterAPI.getByBook(bookId)
  if (res.success && res.data) {
    availableCharacters.value = res.data
  } else {
    availableCharacters.value = []
  }
}

const fetchVolumes = async () => {
  const res = await volumeAPI.getByBook(bookId)
  if (res.success && res.data) {
    volumes.value = res.data
  }
}

const toggleChapterOrder = () => {
  isDescending.value = !isDescending.value
}

// 面板拖拽调节
const startResize = (side: 'left' | 'right' | 'right2', event: MouseEvent) => {
  isResizing.value = true
  resizeSide.value = side
  document.addEventListener('mousemove', handleResize)
  document.addEventListener('mouseup', stopResize)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

const handleResize = (event: MouseEvent) => {
  if (!isResizing.value) return
  
  const container = document.querySelector('.content-wrapper') as HTMLElement
  if (!container) return
  
  const containerRect = container.getBoundingClientRect()
  const mouseX = event.clientX - containerRect.left
  
  if (resizeSide.value === 'left') {
    const newLeftWidth = mouseX
    if (newLeftWidth >= 200 && newLeftWidth <= 600) {
      leftPanelWidth.value = newLeftWidth
      updateCenterWidth()
    }
  } else if (resizeSide.value === 'right') {
    const newRightWidth = containerRect.width - mouseX
    if (newRightWidth >= 300 && newRightWidth <= 800) {
      rightPanelWidth.value = newRightWidth
      updateCenterWidth()
    }
  } else if (resizeSide.value === 'right2') {
    const newRight2Width = containerRect.width - mouseX
    if (newRight2Width >= 300 && newRight2Width <= 800) {
      rightPanel2Width.value = newRight2Width
      updateCenterWidth()
    }
  }
}

const stopResize = () => {
  isResizing.value = false
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}

const updateCenterWidth = () => {
  const container = document.querySelector('.content-wrapper') as HTMLElement
  if (!container) return
  
  const containerWidth = container.offsetWidth
  const handleWidth = 4
  
  let availableWidth = containerWidth - leftPanelWidth.value - handleWidth
  
  if (showChatPanel.value) {
    availableWidth -= rightPanelWidth.value + handleWidth
  }
  
  if (showChatPanel2.value) {
    availableWidth -= rightPanel2Width.value + handleWidth
  }
  
  if (availableWidth < 400) {
    availableWidth = 400
  }
  
  centerPanelWidth.value = availableWidth
}

// 监听窗口大小变化
const handleResizeWindow = () => {
  updateCenterWidth()
}

// 工具栏拖拽调节
const startToolbarResize = (event: MouseEvent) => {
  isResizingToolbar.value = true
  document.addEventListener('mousemove', handleToolbarResize)
  document.addEventListener('mouseup', stopToolbarResize)
  document.body.style.cursor = 'row-resize'
  document.body.style.userSelect = 'none'
}

const handleToolbarResize = (event: MouseEvent) => {
  if (!isResizingToolbar.value) return
  
  const mouseY = event.clientY
  const newHeight = mouseY
  
  // 限制工具栏高度范围：48px - 120px
  if (newHeight >= 48 && newHeight <= 120) {
    toolbarHeight.value = newHeight
  }
}

const stopToolbarResize = () => {
  isResizingToolbar.value = false
  document.removeEventListener('mousemove', handleToolbarResize)
  document.removeEventListener('mouseup', stopToolbarResize)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}

// 提示词选择相关函数
const isSelectedPrompt = (promptId: number) => {
  return selectedPrompts.value.includes(promptId)
}

const togglePromptSelect = (prompt: Prompt) => {
  const index = selectedPrompts.value.indexOf(prompt.id)
  if (index > -1) {
    selectedPrompts.value.splice(index, 1)
  } else {
    selectedPrompts.value.push(prompt.id)
  }
}

const scrollCategory = (direction: 'left' | 'right') => {
  if (categoryTabsRef.value) {
    const scrollAmount = 150
    if (direction === 'left') {
      categoryTabsRef.value.scrollLeft -= scrollAmount
    } else {
      categoryTabsRef.value.scrollLeft += scrollAmount
    }
  }
}

const scrollCreative2Category = (direction: 'left' | 'right') => {
  if (creative2CategoryTabsRef.value) {
    const scrollAmount = 150
    if (direction === 'left') {
      creative2CategoryTabsRef.value.scrollLeft -= scrollAmount
    } else {
      creative2CategoryTabsRef.value.scrollLeft += scrollAmount
    }
  }
}

// 检查分类标签是否可以滚动
const checkCategoryScroll = () => {
  if (categoryTabsRef.value) {
    canScrollLeft.value = categoryTabsRef.value.scrollLeft > 0
    canScrollRight.value = categoryTabsRef.value.scrollLeft < (categoryTabsRef.value.scrollWidth - categoryTabsRef.value.clientWidth)
  }
}

const checkCreative2CategoryScroll = () => {
  if (creative2CategoryTabsRef.value) {
    creative2CanScrollLeft.value = creative2CategoryTabsRef.value.scrollLeft > 0
    creative2CanScrollRight.value = creative2CategoryTabsRef.value.scrollLeft < (creative2CategoryTabsRef.value.scrollWidth - creative2CategoryTabsRef.value.clientWidth)
  }
}

// 监听分类标签滚动
const handleCategoryScroll = () => {
  checkCategoryScroll()
}

const handleCreative2CategoryScroll = () => {
  checkCreative2CategoryScroll()
}

// 从提示词中提取分类
const extractCategories = () => {
  const categories = new Set(['全部'])
  prompts.value.forEach(prompt => {
    if (prompt.category) {
      categories.add(prompt.category)
    }
  })
  promptCategories.value = Array.from(categories)
  creative2PromptCategories.value = Array.from(categories)
}

// 重写 fetchPrompts 函数，同时提取分类
const fetchPrompts = async () => {
  const res = await promptAPI.getAll()
  if (res.success && res.data) {
    prompts.value = res.data
    extractCategories()
  }
}

const fetchConfigs = async () => {
  const res = await configAPI.getAll()
  if (res.success && res.data) {
    apiConfigs.value = res.data
    
    // 如果已经有选中的配置，检查它是否还存在
    if (selectedConfigId.value) {
      const exists = res.data.find(c => c.id === selectedConfigId.value)
      if (!exists) {
        // 配置被删除了，使用默认配置
        const defaultConfig = res.data.find(c => c.is_default === 1)
        selectedConfigId.value = defaultConfig?.id
      }
    } else {
      // 首次加载，使用默认配置
      const defaultConfig = res.data.find(c => c.is_default === 1)
      if (defaultConfig) {
        selectedConfigId.value = defaultConfig.id
      }
    }
    
    // 设置续写功能的默认配置
    if (!continueConfigId.value && res.data.length > 0) {
      const defaultConfig = res.data.find((c: any) => c.is_default === 1)
      continueConfigId.value = defaultConfig?.id || res.data[0].id
    }
  }
}

const fetchConversations = async () => {
  const res = await conversationAPI.getByBook(bookId)
  if (res.success && res.data) {
    conversations.value = res.data
    if (res.data.length > 0 && !currentConversation.value) {
      await selectConversation(res.data[0])
    } else if (res.data.length === 0) {
      // 如果没有对话，自动创建一个
      await createConversation()
    }
  }
}

const createConversation = async () => {
  const res = await conversationAPI.create({
    book_id: bookId,
    title: `新对话 ${new Date().toLocaleTimeString()}`
  })
  if (res.success && res.data) {
    await fetchConversations()
    await selectConversation(res.data)
  }
}

const selectConversation = async (conv: any) => {
  currentConversation.value = conv
  const res = await conversationAPI.getMessages(conv.id)
  if (res.success && res.data) {
    chatMessages.value = res.data
    await nextTick()
    scrollToBottom()
  }
}

const readHistoryStorage = () => {
  try {
    const raw = localStorage.getItem(WRITE_HISTORY_STORAGE_KEY)
    const parsed = raw ? JSON.parse(raw) : []
    return Array.isArray(parsed) ? parsed as WriteHistoryRecord[] : []
  } catch (error) {
    console.error('读取正文历史记录失败:', error)
    return []
  }
}

const writeHistoryStorage = (records: WriteHistoryRecord[]) => {
  localStorage.setItem(WRITE_HISTORY_STORAGE_KEY, JSON.stringify(records))
}

const buildHistoryMessages = (options: {
  promptContents?: string[]
  systemContents?: string[]
  userContents?: string[]
  assistantContents?: string[]
}) => {
  const baseTimestamp = Date.now()
  const messages: WriteHistoryMessage[] = []

  options.systemContents?.filter(Boolean).forEach((content, index) => {
    messages.push({
      role: 'system',
      content,
      timestamp: baseTimestamp + index
    })
  })

  const promptStart = messages.length
  options.promptContents?.filter(Boolean).forEach((content, index) => {
    messages.push({
      role: 'prompt',
      content,
      timestamp: baseTimestamp + promptStart + index
    })
  })

  const userStart = messages.length
  options.userContents?.filter(Boolean).forEach((content, index) => {
    messages.push({
      role: 'user',
      content,
      timestamp: baseTimestamp + userStart + index
    })
  })

  const assistantStart = messages.length
  options.assistantContents?.filter(Boolean).forEach((content, index) => {
    messages.push({
      role: 'assistant',
      content,
      timestamp: baseTimestamp + assistantStart + index
    })
  })

  return messages
}

const saveHistoryRecord = (record: Omit<WriteHistoryRecord, 'id' | 'bookId' | 'timestamp'>) => {
  const historyRecord: WriteHistoryRecord = {
    ...record,
    id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    bookId,
    timestamp: Date.now()
  }

  const records = readHistoryStorage()
  records.unshift(historyRecord)
  writeHistoryStorage(records)
  historyRecords.value = records
    .filter(item => item.bookId === bookId)
    .sort((a, b) => b.timestamp - a.timestamp)

  return historyRecord.id
}

// 加载所有历史记录
const loadAllHistory = async () => {
  loadingHistory.value = true
  try {
    historyRecords.value = readHistoryStorage()
      .filter(item => item.bookId === bookId)
      .sort((a, b) => b.timestamp - a.timestamp)
  } catch (error) {
    console.error('加载历史记录失败:', error)
    ElMessage.error('加载历史记录失败')
  } finally {
    loadingHistory.value = false
  }
}

const viewHistoryDetail = (record: WriteHistoryRecord) => {
  selectedHistoryConversation.value = record
  historyDialogVisible.value = true
}

const deleteHistoryRecord = (recordId: string) => {
  const records = readHistoryStorage().filter(item => item.id !== recordId)
  writeHistoryStorage(records)
  historyRecords.value = records
    .filter(item => item.bookId === bookId)
    .sort((a, b) => b.timestamp - a.timestamp)

  if (selectedHistoryConversation.value?.id === recordId) {
    selectedHistoryConversation.value = null
    historyDialogVisible.value = false
  }
}

const clearHistoryRecords = () => {
  const records = readHistoryStorage().filter(item => item.bookId !== bookId)
  writeHistoryStorage(records)
  historyRecords.value = []
  selectedHistoryConversation.value = null
  historyDialogVisible.value = false
  ElMessage.success('历史记录已清空')
}

const getHistoryStatusLabel = (status: WriteHistoryStatus) => {
  if (status === 'completed') return '已完成'
  if (status === 'cancelled') return '已取消'
  return '已失败'
}

const getHistoryStatusType = (status: WriteHistoryStatus) => {
  if (status === 'completed') return 'success'
  if (status === 'cancelled') return 'warning'
  return 'danger'
}

const getHistoryRoleLabel = (role: WriteHistoryRole) => {
  if (role === 'user') return '用户'
  if (role === 'prompt') return '提示词'
  if (role === 'assistant') return 'AI'
  return '系统'
}

const getHistoryMessageClass = (role: WriteHistoryRole) => {
  if (role === 'user') return 'user'
  if (role === 'prompt') return 'prompt'
  if (role === 'assistant') return 'assistant'
  return 'system'
}

// 查看对话消息
const viewConversationMessages = (conv: any) => {
  if (expandedConversationId.value === conv.id) {
    expandedConversationId.value = null
  } else {
    expandedConversationId.value = conv.id
  }
}

// 复制整个对话
const copyConversation = async (conv: any) => {
  if (!conv.messages || conv.messages.length === 0) {
    ElMessage.warning('该对话没有消息')
    return
  }
  
  let text = `对话：${conv.title || '新对话'}\n`
  text += `创建时间：${conv.created_at}\n`
  text += `消息数量：${conv.messages.length}\n\n`
  text += '---\n\n'
  
  conv.messages.forEach((msg: any, index: number) => {
    const role = msg.role === 'user' ? '用户' : 'AI'
    text += `[${role}] ${formatTime(msg.created_at)}:\n${msg.content}\n\n`
    if (index < conv.messages.length - 1) {
      text += '---\n\n'
    }
  })
  
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('复制成功')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

// 处理对话操作命令
const handleConversationCommand = async (command: string, conv: any) => {
  if (command === 'rename') {
    const { value } = await ElMessageBox.prompt('请输入新的对话名称', '重命名对话', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputValue: conv.title || '新对话',
      inputPattern: /.+/,
      inputErrorMessage: '对话名称不能为空'
    })
    
    if (value) {
      const res = await conversationAPI.update(conv.id, { title: value })
      if (res.success) {
        ElMessage.success('重命名成功')
        // 更新本地列表
        const target = conversationsList.value.find(c => c.id === conv.id)
        if (target) {
          target.title = value
        }
        // 如果是当前选中的对话，也更新
        if (currentConversation.value && currentConversation.value.id === conv.id) {
          currentConversation.value.title = value
        }
      } else {
        ElMessage.error('重命名失败')
      }
    }
  } else if (command === 'clear') {
    try {
      await ElMessageBox.confirm('确定要清空此对话的所有消息吗？', '清空消息', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      
      // 删除所有消息
      for (const msg of conv.messages) {
        if (msg.id) {
          await conversationAPI.deleteMessage(conv.id, msg.id)
        }
      }
      
      // 更新本地列表
      const target = conversationsList.value.find(c => c.id === conv.id)
      if (target) {
        target.messages = []
        target.message_count = 0
      }
      
      // 如果是当前选中的对话，也清空
      if (currentConversation.value && currentConversation.value.id === conv.id) {
        chatMessages.value = []
      }
      
      ElMessage.success('消息已清空')
      // 刷新历史记录
      await loadAllHistory()
    } catch (error) {
      // 用户取消
    }
  } else if (command === 'delete') {
    try {
      await ElMessageBox.confirm('确定要删除此对话吗？此操作不可恢复。', '删除对话', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      
      const res = await conversationAPI.delete(conv.id)
      if (res.success) {
        ElMessage.success('删除成功')
        // 从本地列表移除
        conversationsList.value = conversationsList.value.filter(c => c.id !== conv.id)
        
        // 如果是当前选中的对话，重置当前对话
        if (currentConversation.value && currentConversation.value.id === conv.id) {
          currentConversation.value = null
          chatMessages.value = []
        }
      } else {
        ElMessage.error('删除失败')
      }
    } catch (error) {
      // 用户取消
    }
  }
}

// 格式化时间
const formatTime = (timeStr: string) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)
  
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`
  
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatTimestamp = (timestamp: number) => {
  return formatTime(new Date(timestamp).toISOString())
}

const formatExactTime = (timeStr: string) => {
  if (!timeStr) return '--'
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const GLOBAL_MEMO_FOLDER_STORAGE_KEY = 'write-global-memo-folders'

const loadGlobalMemoFolders = () => {
  try {
    const raw = localStorage.getItem(GLOBAL_MEMO_FOLDER_STORAGE_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) {
      globalMemoCustomFolders.value = parsed.filter(folder => typeof folder === 'string' && folder.trim())
    }
  } catch (error) {
    globalMemoCustomFolders.value = []
  }
}

const persistGlobalMemoFolders = () => {
  localStorage.setItem(GLOBAL_MEMO_FOLDER_STORAGE_KEY, JSON.stringify(globalMemoCustomFolders.value))
}

// 复制消息内容
const copyMessageContent = async (content: string) => {
  try {
    await navigator.clipboard.writeText(content)
    ElMessage.success('复制成功')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

// 监听历史记录弹窗打开
watch(showHistoryDialog, async (newVal) => {
  if (newVal) {
    await loadAllHistory()
  }
})

const toggleChatPanel = async () => {
  const nextState = !showChatPanel.value
  if (nextState) {
    showChatPanel2.value = false
    showGraphPanel.value = false
  }
  showChatPanel.value = nextState
  if (showChatPanel.value) {
    if (!currentConversation.value && conversations.value.length === 0) {
      await createConversation()
    } else if (!currentConversation.value && conversations.value.length > 0) {
      await selectConversation(conversations.value[0])
    }
  }
  setTimeout(() => {
    updateCenterWidth()
  }, 100)
}

const closeChatPanel = () => {
  showChatPanel.value = false
  setTimeout(() => {
    updateCenterWidth()
  }, 100)
}

const toggleChatPanel2 = async () => {
  const nextState = !showChatPanel2.value
  if (nextState) {
    showChatPanel.value = false
    showGraphPanel.value = false
  }
  showChatPanel2.value = nextState
  setTimeout(() => {
    updateCenterWidth()
  }, 100)
}

const closeChatPanel2 = () => {
  showChatPanel2.value = false
  setTimeout(() => {
    updateCenterWidth()
  }, 100)
}

const toggleGraphPanel = () => {
  showGraphPanel.value = !showGraphPanel.value
}

const closeGraphPanel = () => {
  showGraphPanel.value = false
  setTimeout(() => {
    updateCenterWidth()
  }, 100)
}

const onGraphAnalyzeStart = () => {}

const onGraphAnalyzeEnd = (_data: KnowledgeGraphData) => {}

const isCreative2PromptSelected = (promptId: number) => {
  return creative2SelectedPrompts.value.includes(promptId)
}

const toggleCreative2Prompt = (prompt: any) => {
  const index = creative2SelectedPrompts.value.indexOf(prompt.id)
  if (index > -1) {
    creative2SelectedPrompts.value.splice(index, 1)
  } else {
    creative2SelectedPrompts.value.push(prompt.id)
  }
}

const removeCreative2Prompt = (promptId: number) => {
  const index = creative2SelectedPrompts.value.indexOf(promptId)
  if (index > -1) {
    creative2SelectedPrompts.value.splice(index, 1)
  }
}

const selectSecondPrompt = (prompt: any) => {
  if (creative2SecondPromptId.value === prompt.id) {
    creative2SecondPromptId.value = 0
  } else {
    creative2SecondPromptId.value = prompt.id
  }
}

const getPromptName = (promptId: number) => {
  const prompt = prompts.value.find(p => p.id === promptId)
  return prompt?.name || '未知提示词'
}

const creative2PromptInfo = computed(() => {
  if (creative2SelectedPrompts.value.length > 0) {
    return prompts.value.find(p => p.id === creative2SelectedPrompts.value[0])
  }
  return null
})

const creative2SecondPromptInfo = computed(() => {
  if (creative2SecondPromptId.value) {
    return prompts.value.find(p => p.id === creative2SecondPromptId.value)
  }
  return null
})

const creative2PromptIntroDialogVisible = ref(false)
const creative2SecondPromptIntroDialogVisible = ref(false)
const isPromptContentCollapsed = ref(true)

const showCreative2PromptIntro = () => {
  creative2PromptIntroDialogVisible.value = true
  isPromptContentCollapsed.value = true
}

const showCreative2SecondPromptIntro = () => {
  creative2SecondPromptIntroDialogVisible.value = true
  isPromptContentCollapsed.value = true
}

const togglePromptContentCollapse = () => {
  isPromptContentCollapsed.value = !isPromptContentCollapsed.value
}

const handleCreative2FixedPromptToggle = (value: boolean) => {
  if (value) {
    if (creative2SelectedPrompts.value.length > 0) {
      creative2FixedPromptId.value = creative2SelectedPrompts.value[0]
    }
  } else {
    creative2FixedPromptId.value = 0
  }
}

// 监听提示词变化，清空字段值
watch([creative2SelectedPrompts, creative2UseFixedPrompt, creative2FixedPromptId], () => {
  creative2FieldValues.value = {}
}, { deep: true })

const handleCreative2Generate = async () => {
  if (!creative2ConfigId.value) {
    ElMessage.warning('请选择AI模型')
    return
  }

  let promptsToUse: number[] = []
  
  if (creative2UseFixedPrompt.value && creative2FixedPromptId.value !== 0) {
    promptsToUse = [creative2FixedPromptId.value]
  } else if (creative2SelectedPrompts.value.length > 0) {
    promptsToUse = creative2SelectedPrompts.value
  }

  creative2Generating.value = true
  creative2Result.value = ''
  creative2ResultDialogVisible.value = true

  try {
    const selectedPromptObjects = prompts.value.filter(p => 
      promptsToUse.includes(p.id)
    )
    
    const secondPromptObj = creative2SecondPromptId.value 
      ? prompts.value.find(p => p.id === creative2SecondPromptId.value)
      : null
    
    // 将字段值代入提示词模板
    let systemPrompts = selectedPromptObjects.map(p => {
      let content = p.content
      if (creative2AllFields.value.length > 0) {
        creative2AllFields.value.forEach(field => {
          const value = creative2FieldValues.value[field.name] || ''
          const regex = new RegExp(`\\$\\{${field.name}\\}`, 'g')
          content = content.replace(regex, value)
        })
      }
      return content
    })

    // 如果有第二个提示词，也添加到系统提示词
    if (secondPromptObj) {
      systemPrompts.push(secondPromptObj.content)
    }
    
    const config = apiConfigs.value.find(c => c.id === creative2ConfigId.value)
    if (!config) {
      ElMessage.error('未找到 API 配置')
      return
    }

    // 构建用户消息内容
    let userMessageContent = ''

    // 添加故事背景
    if (creative2StoryBackground.value.trim()) {
      userMessageContent += `【故事背景】\n${creative2StoryBackground.value.trim()}\n\n`
    }

    // 添加角色关系
    if (creative2CharacterRelations.value.trim()) {
      userMessageContent += `【角色关系】\n${creative2CharacterRelations.value.trim()}\n\n`
    }

    // 添加本章剧情
    if (creative2ChapterPlot.value.trim()) {
      userMessageContent += `【本章剧情】\n${creative2ChapterPlot.value.trim()}\n\n`
    }

    // 添加补充信息
    if (creative2AdditionalInfo.value.trim()) {
      userMessageContent += `【补充信息】\n${creative2AdditionalInfo.value.trim()}\n\n`
    }

    // 添加关联角色卡内容
    if (creative2SelectedCharacterIds.value.length > 0) {
      const selectedCharacters = availableCharacters.value.filter(c => 
        creative2SelectedCharacterIds.value.includes(c.id)
      )
      if (selectedCharacters.length > 0) {
        userMessageContent += `【关联角色卡】（仅供参考）\n`
        selectedCharacters.forEach(character => {
          userMessageContent += `【${character.name}】\n${character.info || '（无描述）'}\n`
        })
        userMessageContent += '\n'
      }
    }

    // 添加关联章节内容
    if (creative2SelectedChapterId.value) {
      const chapter = chaptersList.value.find(c => c.id === creative2SelectedChapterId.value)
      if (chapter) {
        userMessageContent += `【关联章节】（仅供参考）\n${chapter.title}\n${chapter.content || '（无内容）'}\n\n`
      }
    }

    // 添加关联备忘录内容
    if (creative2SelectedMemoId.value) {
      const memo = memos.value.find(m => m.id === creative2SelectedMemoId.value)
      if (memo) {
        userMessageContent += `【关联备忘录】（仅供参考）\n${memo.title}\n${memo.content || '（无内容）'}\n\n`
      }
    }

    userMessageContent += '请根据以上信息生成创意内容。'

    // 保存到对话记录
    const convRes = await conversationAPI.create({
      book_id: bookId,
      title: `创意生成 - ${selectedPromptObjects.map(p => p.name).join(', ')}${secondPromptObj ? ` + ${secondPromptObj.name}` : ''}`
    })
    
    if (convRes.success && convRes.data) {
      const convId = convRes.data.id
      
      // 保存系统提示词（如果有）
      if (systemPrompts.length > 0) {
        await conversationAPI.saveMessage(convId, {
          role: 'system',
          content: systemPrompts.join('\n\n---\n\n')
        })
      }
      
      // 保存用户消息
      await conversationAPI.saveMessage(convId, {
        role: 'user',
        content: userMessageContent
      })
    }

    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        configId: creative2ConfigId.value,
        systemPrompts: systemPrompts,
        messages: [{
          role: 'user',
          content: userMessageContent
        }]
      })
    })

    if (!response.ok) {
      throw new Error(`API请求失败: ${response.status}`)
    }

    const reader = response.body?.getReader()
    const decoder = new TextDecoder()

    if (reader) {
      while (true) {
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
              if (parsed.error) {
                throw new Error(parsed.error)
              }
              if (parsed.content) {
                creative2Result.value += parsed.content
              }
            } catch (e: any) {
              if (e.message && !e.message.includes('JSON')) {
                throw e
              }
            }
          }
        }
      }
    }

    ElMessage.success('生成完成')
    
    // 保存历史记录
    saveHistoryRecord({
      title: `创意生成 - ${selectedPromptObjects.map(p => p.name).join('、') || '未命名任务'}`,
      source: 'creative2',
      sourceLabel: '正文创意生成',
      promptName: selectedPromptObjects.map(p => p.name).join('、'),
      promptCount: systemPrompts.length,
      status: 'completed',
      previewContent: (creative2Result.value || userMessageContent).substring(0, 140),
      messages: buildHistoryMessages({
        promptContents: systemPrompts,
        userContents: [userMessageContent],
        assistantContents: creative2Result.value ? [creative2Result.value] : []
      })
    })
  } catch (error: any) {
    console.error('创意生成错误:', error)
    saveHistoryRecord({
      title: `创意生成 - ${prompts.value.filter(p => promptsToUse.includes(p.id)).map(p => p.name).join('、') || '未命名任务'}`,
      source: 'creative2',
      sourceLabel: '正文创意生成',
      promptName: prompts.value.filter(p => promptsToUse.includes(p.id)).map(p => p.name).join('、'),
      promptCount: prompts.value.filter(p => promptsToUse.includes(p.id)).length,
      status: 'failed',
      previewContent: (creative2Result.value || '生成失败').substring(0, 140),
      messages: buildHistoryMessages({
        promptContents: prompts.value
          .filter(p => promptsToUse.includes(p.id))
          .map(p => p.content),
        userContents: [''],
        assistantContents: creative2Result.value ? [creative2Result.value] : []
      })
    })
    ElMessage.error(error.message || '生成失败')
  } finally {
    creative2Generating.value = false
  }
}

const copyCreative2Result = async () => {
  try {
    await navigator.clipboard.writeText(creative2Result.value)
    ElMessage.success('复制成功')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const applyCreative2Result = () => {
  if (!currentChapter.value || !creative2Result.value) {
    ElMessage.warning('请选择章节或没有可应用的内容')
    return
  }
  
  // 将生成结果追加到当前章节内容末尾
  if (currentChapter.value.content) {
    currentChapter.value.content += '\n\n' + creative2Result.value
  } else {
    currentChapter.value.content = creative2Result.value
  }
  
  ElMessage.success('已应用到当前章节')
  creative2ResultDialogVisible.value = false
  
  // 自动保存章节
  saveChapter()
}

const copyMessage = async (content: string) => {
  try {
    await navigator.clipboard.writeText(content)
    ElMessage.success('复制成功')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const startEdit = (index: number, content: string) => {
  editingMessageIndex.value = index
  editingContent.value = content
}

const saveEdit = async (index: number) => {
  if (!editingContent.value.trim()) {
    ElMessage.warning('消息内容不能为空')
    return
  }
  
  const message = chatMessages.value[index]
  if (!message) {
    ElMessage.error('消息不存在')
    return
  }
  
  // 如果消息有 ID，说明已保存到数据库，需要调用 API 更新
  if (message.id && currentConversation.value) {
    const res = await conversationAPI.updateMessage(
      currentConversation.value.id, 
      message.id, 
      { content: editingContent.value }
    )
    if (!res.success) {
      ElMessage.error('保存失败')
      return
    }
  }
  
  // 更新本地内容
  message.content = editingContent.value
  editingMessageIndex.value = null
  editingContent.value = ''
  ElMessage.success('保存成功')
}

const cancelEdit = () => {
  editingMessageIndex.value = null
  editingContent.value = ''
}

const getContentLength = (content: string) => {
  return content ? content.length : 0
}

const handleTextSelect = (event: Event) => {
  const target = event.target as HTMLTextAreaElement
  const selectedText = target.value.substring(target.selectionStart, target.selectionEnd)
  // 只有选中文本时才显示字数，否则清零
  selectedTextLength.value = selectedText.length > 0 ? selectedText.length : 0
}

// 编辑功能
const handleBold = () => {
  insertTextAround('**', '**')
}

const handleItalic = () => {
  insertTextAround('*', '*')
}

const handleHighlight = () => {
  insertTextAround('==', '==')
}

const handleFormat = () => {
  const textarea = chapterEditorRef.value?.$el?.querySelector('textarea') as HTMLTextAreaElement
  if (!textarea || !currentChapter.value) return
  
  let text = textarea.value
  // 智能排版：段首空格、段落间距、标点符号优化
  text = text.replace(/[ \t]+/g, ' ') // 多个空格变一个
  text = text.replace(/\n{3,}/g, '\n\n') // 多个空行变两个
  text = text.replace(/([,.!?;:，。！？；：])\s*/g, '$1') // 标点后去空格
  text = text.replace(/\n/g, '\n    ') // 段首加空格
  text = '    ' + text.trimStart() // 第一段加空格
  
  textarea.value = text
  currentChapter.value.content = text
  ElMessage.success('排版完成')
}

const handleReplace = () => {
  const textarea = chapterEditorRef.value?.$el?.querySelector('textarea') as HTMLTextAreaElement
  if (!textarea) return
  
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const selectedText = textarea.value.substring(start, end)
  
  if (!selectedText) {
    ElMessage.warning('请先选中文本')
    return
  }
  
  ElMessageBox.prompt('请输入替换内容', '选中替换', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputValue: selectedText,
    inputPattern: /.*/,
    inputErrorMessage: '请输入内容'
  }).then(({ value }) => {
    const text = textarea.value
    textarea.value = text.substring(0, start) + value + text.substring(end)
    if (currentChapter.value) {
      currentChapter.value.content = textarea.value
    }
    ElMessage.success('替换成功')
  }).catch(() => {})
}

const handleSearch = () => {
  const textarea = chapterEditorRef.value?.$el?.querySelector('textarea') as HTMLTextAreaElement
  if (!textarea) return
  
  ElMessageBox.prompt('搜索内容', '搜索', {
    confirmButtonText: '查找',
    cancelButtonText: '取消',
    inputPattern: /.*/,
    inputErrorMessage: '请输入内容'
  }).then(({ value }) => {
    const text = textarea.value
    const index = text.indexOf(value)
    if (index !== -1) {
      textarea.focus()
      textarea.setSelectionRange(index, index + value.length)
      ElMessage.success('找到匹配内容')
    } else {
      ElMessage.warning('未找到匹配内容')
    }
  }).catch(() => {})
}

const handleCopy = async () => {
  const textarea = chapterEditorRef.value?.$el?.querySelector('textarea') as HTMLTextAreaElement
  if (!textarea) return
  
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const selectedText = textarea.value.substring(start, end)
  
  if (!selectedText) {
    ElMessage.warning('请先选中文本')
    return
  }
  
  try {
    await navigator.clipboard.writeText(selectedText)
    ElMessage.success('复制成功')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const handleSelectAll = () => {
  const textarea = chapterEditorRef.value?.$el?.querySelector('textarea') as HTMLTextAreaElement
  if (!textarea) return
  
  textarea.select()
  selectedTextLength.value = textarea.value.length
}

const handleUndo = () => {
  document.execCommand('undo')
}

const handleRedo = () => {
  document.execCommand('redo')
}

const insertTextAround = (before: string, after: string) => {
  const textarea = chapterEditorRef.value?.$el?.querySelector('textarea') as HTMLTextAreaElement
  if (!textarea) return
  
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const text = textarea.value
  const selectedText = text.substring(start, end)
  
  const newText = text.substring(0, start) + before + selectedText + after + text.substring(end)
  textarea.value = newText
  
  if (currentChapter.value) {
    currentChapter.value.content = newText
  }
  
  // 恢复光标位置
  setTimeout(() => {
    textarea.focus()
    textarea.setSelectionRange(start + before.length, end + before.length)
  }, 0)
}

// 监听鼠标抬起和键盘事件，实时更新选中状态
const handleMouseUp = (event: Event) => {
  handleTextSelect(event)
}

const handleKeyUp = (event: Event) => {
  handleTextSelect(event)
}

// 保存字体大小到localStorage
const saveFontSize = () => {
  localStorage.setItem('editorFontSize', fontSize.value.toString())
}

const goToBookAnalysis = () => {
  router.push({ 
    name: 'BookAnalysis', 
    params: { bookId: bookId } 
  })
}

const goToCharacterLibrary = () => {
  router.push({ 
    name: 'CharacterLibrary', 
    params: { bookId: bookId } 
  })
}

// 从localStorage加载字体大小
const loadFontSize = () => {
  const savedSize = localStorage.getItem('editorFontSize')
  if (savedSize) {
    fontSize.value = parseInt(savedSize)
  }
}

const getActiveContentTitle = () => {
  return activeContentType.value === 'chapters'
    ? (currentChapter.value?.title || '未命名章节')
    : (currentMemo.value?.title || '未命名备忘录')
}

const updateActiveContent = (content: string) => {
  if (activeContentType.value === 'chapters' && currentChapter.value) {
    currentChapter.value.content = content
  } else if (currentMemo.value) {
    currentMemo.value.content = content
  }
}

const saveActiveContent = async () => {
  if (activeContentType.value === 'chapters') {
    await saveChapter()
  } else {
    await saveMemo()
  }
}

// 续写功能
const continuWrite = async () => {
  if (!continueConfigId.value) {
    ElMessage.warning('请选择API配置')
    return
  }
  
  const content = activeContentType.value === 'chapters' ? currentChapter.value?.content : currentMemo.value?.content
  
  if (!content || !content.includes('user:')) {
    ElMessage.warning('请先输入 user: 格式的指令')
    return
  }
  
  // 解析内容中的对话 - 支持多行格式
  const lines = content!.split('\n')
  let beforeLines: string[] = []
  let afterLines: string[] = []
  const messages: any[] = []
  const systemPrompts: string[] = []
  let currentTargetTitle = ''
  let aiResponse = ''
  
  try {
    continuWriting.value = true
    
    // 创建AbortController用于取消请求
    continueAbortController.value = new AbortController()
    let inAiResponse = false
    let aiContent = ''
    let inUserMessage = false
    let userContent = ''
    
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      
      if (!line) continue
      
      // 检测 user: 开头
      if (line.trim().startsWith('user:')) {
        // 如果有之前的AI回复，保存它
        if (inAiResponse && aiContent.trim()) {
          messages.push({
            role: 'assistant',
            content: aiContent.trim()
          })
          aiContent = ''
          inAiResponse = false
        }
        
        // 如果之前有用户消息在收集中，先保存它
        if (inUserMessage && userContent.trim()) {
          messages.push({
            role: 'user',
            content: userContent.trim()
          })
        }
        
        // 开始收集新的用户消息（包含 user: 后面的内容和后续行）
        inUserMessage = true
        userContent = line.trim().substring(5).trim()
      }
      // 检测 -> 开头（AI回复开始）
      else if (line.trim() === '->') {
        // 保存之前的用户消息
        if (inUserMessage && userContent.trim()) {
          messages.push({
            role: 'user',
            content: userContent.trim()
          })
          userContent = ''
          inUserMessage = false
        }
        
        inAiResponse = true
        aiContent = ''
      }
      // 检测 <- 结尾（AI回复结束）
      else if (line.trim() === '<-') {
        if (inAiResponse && aiContent.trim()) {
          messages.push({
            role: 'assistant',
            content: aiContent.trim()
          })
        }
        inAiResponse = false
        aiContent = ''
      }
      // 如果在用户消息收集模式，继续收集内容
      else if (inUserMessage) {
        userContent += (userContent ? '\n' : '') + line
      }
      // AI回复内容
      else if (inAiResponse) {
        aiContent += (aiContent ? '\n' : '') + line
      }
    }
    
    // 处理最后一条可能未保存的用户消息
    if (inUserMessage && userContent.trim()) {
      messages.push({
        role: 'user',
        content: userContent.trim()
      })
    }
    
    if (messages.length === 0 || messages[messages.length - 1].role !== 'user') {
      ElMessage.warning('请确保最后一条是 user: 格式的指令')
      return
    }
    
    // 获取系统提示词
    if (continuePromptId.value && continuePromptId.value !== 0) {
      const prompt = prompts.value.find((p: any) => p.id == continuePromptId.value)
      if (prompt) {
        console.log('续写-选中的提示词:', prompt.name, '内容:', prompt.content)
        systemPrompts.push(prompt.content)
      } else {
        console.warn('续写-未找到提示词，ID:', continuePromptId.value, '所有提示词:', prompts.value)
      }
    }
    console.log('续写-最终systemPrompts:', systemPrompts)

    const lastUserMessage = [...messages].reverse().find(message => message.role === 'user')?.content || '续写请求'
    currentTargetTitle = getActiveContentTitle()
    
    // 找到最后一个user:的位置
    let lastUserLineIndex = -1
    for (let i = lines.length - 1; i >= 0; i--) {
      if (lines[i] && lines[i].trim().startsWith('user:')) {
        lastUserLineIndex = i
        break
      }
    }
    
    if (lastUserLineIndex === -1) {
      ElMessage.error('未找到 user: 指令')
      return
    }
    
    // 找到user:消息结束的位置（到 -> 或下一个user:之前）
    let userMessageEndIndex = lastUserLineIndex + 1
    let foundAiResponse = false
    
    for (let i = lastUserLineIndex + 1; i < lines.length; i++) {
      const trimmed = lines[i]?.trim() || ''
      
      // 遇到 -> 说明AI回复开始，user消息到这里结束
      if (trimmed === '->') {
        userMessageEndIndex = i
        foundAiResponse = true
        break
      }
      
      // 遇到下一个user:，说明user消息到这里结束
      if (trimmed.startsWith('user:')) {
        userMessageEndIndex = i
        break
      }
      
      // 否则继续扫描，这一行属于user消息的一部分
      userMessageEndIndex = i + 1
    }
    
    // 移除旧的AI回复（如果存在）
    let endIndex = userMessageEndIndex
    if (foundAiResponse) {
      // 从 -> 开始扫描，找到 <- 的位置
      for (let i = userMessageEndIndex + 1; i < lines.length; i++) {
        const trimmed = lines[i].trim()
        
        if (trimmed === '<-') {
          endIndex = i + 1
          break
        }
        
        // 遇到新的user:，说明AI回复没有正常结束
        if (trimmed.startsWith('user:')) {
          endIndex = i
          break
        }
      }
    }
    
    // 构建基础内容（保留 user: 及其后面的多行内容）
    beforeLines = lines.slice(0, userMessageEndIndex)
    afterLines = lines.slice(endIndex)
    
    // 先插入 -> 标记
    const initialLines = [
      ...beforeLines,
      '',
      '->'
    ]
    
    let initialContent = initialLines.join('\n')
    updateActiveContent(initialContent)
    
    await nextTick()
    
    // 调用AI
    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        messages,
        configId: continueConfigId.value,
        systemPrompts: systemPrompts
      }),
      signal: continueAbortController.value.signal
    })
    
    if (!response.ok) {
      throw new Error('续写失败')
    }
    
    const reader = response.body?.getReader()
    const decoder = new TextDecoder()
    
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
            
            // 处理错误信息
            if (json.error) {
              ElMessage.error(`AI调用失败: ${json.error}`)
              throw new Error(json.error)
            }
            
            // 处理正常内容
            if (json.content) {
              aiResponse += json.content
              
              // 实时更新内容
              const contentLines = [
                ...beforeLines,
                '',
                '->',
                aiResponse
              ]
              
              const newContent = contentLines.join('\n')
              
              updateActiveContent(newContent)
              
              await nextTick()
            }
          } catch (e: any) {
            console.error('解析响应失败:', e)
            // 如果是错误消息，重新抛出
            if (e.message && !e.message.includes('解析响应失败')) {
              throw e
            }
          }
        }
      }
    }
    
    // 添加结束标记
    const finalLines = [
      ...beforeLines,
      '',
      '->',
      aiResponse,
      '<-',
      ...afterLines
    ]
    
    const finalContent = finalLines.join('\n')
    
    updateActiveContent(finalContent)
    
    // 保存
    await saveActiveContent()
    
    ElMessage.success('续写完成')
    saveHistoryRecord({
      title: `正文续写 - ${currentTargetTitle}`,
      source: 'continue',
      sourceLabel: '正文续写',
      promptName: continuePromptId.value
        ? (prompts.value.find((p: any) => p.id == continuePromptId.value)?.name || '未命名提示词')
        : '',
      promptCount: systemPrompts.length,
      status: 'completed',
      previewContent: (aiResponse || lastUserMessage).substring(0, 140),
      messages: buildHistoryMessages({
        systemContents: [`目标内容：${currentTargetTitle}`],
        promptContents: systemPrompts,
        userContents: messages
          .filter(message => message.role === 'user')
          .map(message => message.content),
        assistantContents: aiResponse ? [aiResponse] : []
      })
    })
  } catch (error: any) {
    const lastUserMessage = [...messages].reverse().find(message => message.role === 'user')?.content || '续写请求'
    if (error.name === 'AbortError') {
      saveHistoryRecord({
        title: `正文续写 - ${currentTargetTitle || getActiveContentTitle()}`,
        source: 'continue',
        sourceLabel: '正文续写',
        promptName: continuePromptId.value
          ? (prompts.value.find((p: any) => p.id == continuePromptId.value)?.name || '未命名提示词')
          : '',
        promptCount: systemPrompts.length,
        status: 'cancelled',
        previewContent: (aiResponse || lastUserMessage).substring(0, 140),
        messages: buildHistoryMessages({
          systemContents: [`目标内容：${currentTargetTitle || getActiveContentTitle()}`],
          promptContents: systemPrompts,
          userContents: messages
            .filter(message => message.role === 'user')
            .map(message => message.content),
          assistantContents: aiResponse ? [aiResponse] : []
        })
      })
      ElMessage.info('续写已暂停')
      // 暂停时也要保存当前内容
      await saveActiveContent()
    } else {
      console.error('续写错误:', error)
      saveHistoryRecord({
        title: `正文续写 - ${currentTargetTitle || getActiveContentTitle()}`,
        source: 'continue',
        sourceLabel: '正文续写',
        promptName: continuePromptId.value
          ? (prompts.value.find((p: any) => p.id == continuePromptId.value)?.name || '未命名提示词')
          : '',
        promptCount: systemPrompts.length,
        status: 'failed',
        previewContent: (aiResponse || lastUserMessage).substring(0, 140),
        messages: buildHistoryMessages({
          systemContents: [`目标内容：${currentTargetTitle || getActiveContentTitle()}`],
          promptContents: systemPrompts,
          userContents: messages
            .filter(message => message.role === 'user')
            .map(message => message.content),
          assistantContents: aiResponse ? [aiResponse] : []
        })
      })
      
      // 发生错误时，恢复到原始内容（移除可能添加的 -> 标记）
      const restoreContent = beforeLines.join('\n')
      updateActiveContent(restoreContent)
      
      // 显示错误信息（已在流式响应中显示过了，这里只在控制台输出）
      if (!error.message.includes('AI调用失败')) {
        ElMessage.error(error.message || '续写失败')
      }
    }
  } finally {
    continuWriting.value = false
    continueAbortController.value = null
  }
}

// 停止续写
const stopContinueWrite = () => {
  if (continueAbortController.value) {
    continueAbortController.value.abort()
    continueAbortController.value = null
  }
  continuWriting.value = false
}

const deleteMessage = async (index: number) => {
  try {
    await ElMessageBox.confirm('确定删除这条消息吗？', '提示', {
      type: 'warning'
    })
    
    const message = chatMessages.value[index]
    if (!message) {
      ElMessage.error('娑堟伅涓嶅瓨鍦?')
      return
    }
    
    // 如果消息有ID，说明已保存到数据库，需要调用API删除
    if (message.id && currentConversation.value) {
      const res = await conversationAPI.deleteMessage(currentConversation.value.id, message.id)
      if (!res.success) {
        ElMessage.error('删除失败')
        return
      }
    }
    
    // 从本地数组中删除
    chatMessages.value.splice(index, 1)
    ElMessage.success('删除成功')
  } catch (error) {
    // 取消删除
  }
}

const handleConversationAction = async (command: string, conv: any) => {
  if (command === 'rename') {
    renameTitle.value = conv.title
    renamingConversation.value = conv
    showRenameDialog.value = true
  } else if (command === 'clear') {
    try {
      await ElMessageBox.confirm('确定清空该对话的所有消息吗？', '提示', {
        type: 'warning'
      })
      const res = await conversationAPI.clearMessages(conv.id)
      if (res.success) {
        if (currentConversation.value?.id === conv.id) {
          chatMessages.value = []
        }
        await fetchConversations()
        ElMessage.success('清空成功')
      }
    } catch (error) {
      // 取消操作
    }
  } else if (command === 'delete') {
    try {
      await ElMessageBox.confirm(`确定删除对话"${conv.title}"吗？`, '提示', {
        type: 'warning'
      })
      const res = await conversationAPI.delete(conv.id)
      if (res.success) {
        if (currentConversation.value?.id === conv.id) {
          currentConversation.value = null
          chatMessages.value = []
        }
        await fetchConversations()
        ElMessage.success('删除成功')
      }
    } catch (error) {
      // 取消操作
    }
  }
}

const confirmRename = async () => {
  if (!renameTitle.value.trim()) {
    ElMessage.warning('请输入标题')
    return
  }
  const res = await conversationAPI.update(renamingConversation.value.id, {
    title: renameTitle.value
  })
  if (res.success) {
    await fetchConversations()
    if (currentConversation.value?.id === renamingConversation.value.id) {
      currentConversation.value.title = renameTitle.value
    }
    showRenameDialog.value = false
    ElMessage.success('重命名成功')
  }
}

const selectChapter = (chapter: Chapter) => {
  catalogType.value = 'chapters'
  currentChapter.value = chapter
  currentMemo.value = null
}

const selectMemo = (memo: Memo) => {
  currentMemo.value = memo
  currentChapter.value = null
}

const handleGlobalMemoSelect = (memo: Memo) => {
  selectMemo(memo)
  showGlobalMemoDialog.value = false
}

const handleCreateChapter = async () => {
  const res = await chapterAPI.create({
    book_id: bookId,
    title: '新章节',
    content: '',
    summary: '',
    order_num: chaptersList.value.length,
    type: 'chapter',
    volume_id: expandedVolumeId.value && typeof expandedVolumeId.value === 'number' ? expandedVolumeId.value : undefined
  })
  if (res.success && res.data) {
    chaptersList.value.push(res.data)
    currentChapter.value = res.data
    currentMemo.value = null
  }
}

const handleCreateVolume = async () => {
  const title = `新分卷 ${volumes.value.length + 1}`
  const res = await volumeAPI.create({
    book_id: bookId,
    title,
    order_num: volumes.value.length
  })
  if (res.success && res.data) {
    volumes.value.push(res.data)
    expandedVolumeId.value = res.data.id
    ElMessage.success('创建成功')
  }
}

const toggleVolume = (volumeId: number | string) => {
  if (expandedVolumeId.value === volumeId) {
    expandedVolumeId.value = null
  } else {
    expandedVolumeId.value = volumeId
  }
}

const editVolume = (volume: Volume) => {
  isEditVolume.value = true
  volumeForm.value = {
    id: volume.id,
    title: volume.title
  }
  showVolumeDialog.value = true
}

const deleteVolume = async (volume: Volume) => {
  try {
    await ElMessageBox.confirm(`确定删除分卷"${volume.title}"吗？删除后章节将不受影响。`, '提示', {
      type: 'warning'
    })
    
    const res = await volumeAPI.delete(volume.id)
    if (res.success) {
      volumes.value = volumes.value.filter(v => v.id !== volume.id)
      if (expandedVolumeId.value === volume.id) {
        expandedVolumeId.value = null
      }
      ElMessage.success('删除成功')
    }
  } catch (error) {
    // 取消删除
  }
}

const handleVolumeSubmit = async () => {
  if (!volumeForm.value.title.trim()) {
    ElMessage.warning('请输入分卷名称')
    return
  }
  
  try {
    if (isEditVolume.value) {
      const res = await volumeAPI.update(volumeForm.value.id, {
        title: volumeForm.value.title
      })
      if (res.success && res.data) {
        const index = volumes.value.findIndex(v => v.id === volumeForm.value.id)
        if (index !== -1) {
          volumes.value[index] = res.data
        }
        ElMessage.success('更新成功')
      }
    }
    showVolumeDialog.value = false
  } catch (error) {
    ElMessage.error(isEditVolume.value ? '更新失败' : '创建失败')
  }
}

const handleCreateMemo = async () => {
  const res = await memoAPI.create({
    title: '新备忘录',
    content: '',
    category: '本书',
    order_num: memos.value.length
  })
  if (res.success && res.data) {
    memos.value.push(res.data)
    currentMemo.value = res.data
    currentChapter.value = null
    showGlobalMemoDialog.value = false
  }
}

const handleCreateMemoInDialog = async (scope?: 'global' | 'book') => {
  const targetScope = scope || memoCreateScope.value
  memoCreateScope.value = targetScope
  const res = await memoAPI.create({
    title: '新备忘录',
    content: '',
    category: targetScope === 'book' ? '本书' : currentGlobalMemoFolder.value,
    order_num: memos.value.length
  })
  if (res.success && res.data) {
    memos.value.push(res.data)
    selectDialogMemo(res.data)
  }
}

const handleCreateGlobalMemoFolder = async () => {
  try {
    const { value } = await ElMessageBox.prompt('请输入全局备忘录文件夹名称', '新建文件夹', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPlaceholder: '例如：灵感池',
      inputValidator: (input) => {
        const name = input.trim()
        if (!name) return '请输入文件夹名称'
        if (globalMemoFolders.value.includes(name)) return '文件夹已存在'
        return true
      }
    })

    const folderName = value.trim()
    globalMemoCustomFolders.value.push(folderName)
    persistGlobalMemoFolders()
    currentGlobalMemoFolder.value = folderName
    folderExpandedStates.value[folderName] = true
    memoCreateScope.value = 'global'
    ElMessage.success('文件夹创建成功')
  } catch (error) {
    // 用户取消
  }
}

const createMemoInFolder = async (folder: string) => {
  const res = await memoAPI.create({
    title: '新备忘录',
    content: '',
    category: folder === '默认' ? '全局' : folder,
    order_num: memos.value.length
  })
  if (res.success && res.data) {
    memos.value.push(res.data)
    currentGlobalMemoFolder.value = folder
    folderExpandedStates.value[folder] = true
    ElMessage.success('备忘录创建成功')
  }
}

const deleteFolder = async (folder: string) => {
  try {
    await ElMessageBox.confirm(`确定要删除文件夹"${folder}"吗？该文件夹下的所有备忘录将被移动到"默认"文件夹。`, '删除文件夹', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const memosInFolder = getMemosByFolder(folder)
    for (const memo of memosInFolder) {
      await memoAPI.update(memo.id, {
        category: '默认'
      })
    }

    const index = globalMemoCustomFolders.value.indexOf(folder)
    if (index !== -1) {
      globalMemoCustomFolders.value.splice(index, 1)
    }
    persistGlobalMemoFolders()
    
    if (currentGlobalMemoFolder.value === folder) {
      currentGlobalMemoFolder.value = '默认'
    }
    delete folderExpandedStates.value[folder]
    
    ElMessage.success('文件夹已删除')
  } catch (error) {
    // 用户取消
  }
}

const openMoveFolderDialog = (memo: Memo) => {
  memoToMove.value = memo
  moveFolderDialogVisible.value = true
}

const moveToFolder = async (folder: string) => {
  if (!memoToMove.value) return
  
  const res = await memoAPI.update(memoToMove.value.id, {
    title: memoToMove.value.title,
    content: memoToMove.value.content,
    category: folder === '默认' ? '全局' : folder,
    tags: memoToMove.value.tags
  })
  
  if (res.success) {
    const index = memos.value.findIndex(m => m.id === memoToMove.value!.id)
    if (index !== -1) {
      memos.value[index] = {
        ...memos.value[index],
        category: folder === '默认' ? '全局' : folder
      }
    }
    ElMessage.success('备忘录已移动')
    moveFolderDialogVisible.value = false
    memoToMove.value = null
  }
}

const deleteMemoInList = async (memo: Memo) => {
  try {
    await ElMessageBox.confirm(`确定要删除备忘录"${memo.title || '无标题'}"吗？`, '删除备忘录', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const res = await memoAPI.delete(memo.id)
    if (res.success) {
      const index = memos.value.findIndex(m => m.id === memo.id)
      if (index !== -1) {
        memos.value.splice(index, 1)
      }
      if (dialogSelectedMemo.value?.id === memo.id) {
        dialogSelectedMemo.value = null
      }
      ElMessage.success('备忘录已删除')
    }
  } catch (error) {
    // 用户取消
  }
}

const handleJumpToMemo = (memo: Memo) => {
  selectMemo(memo)
  showGlobalMemoDialog.value = false
}

const saveDialogMemo = async () => {
  if (!dialogSelectedMemo.value) return
  
  const res = await memoAPI.update(dialogSelectedMemo.value.id, {
    title: dialogSelectedMemo.value.title,
    content: dialogSelectedMemo.value.content,
    tags: JSON.stringify(dialogSelectedMemoTags.value)
  })
  
  if (res.success) {
    const index = memos.value.findIndex(m => m.id === dialogSelectedMemo.value!.id)
    if (index !== -1) {
      memos.value[index] = { 
        ...dialogSelectedMemo.value,
        tags: JSON.stringify(dialogSelectedMemoTags.value)
      }
    }
    ElMessage.success('保存成功')
  }
}

// 当在全局备忘录弹窗中选择备忘录时，同步加载标签
const selectDialogMemo = (memo: Memo) => {
  memoCreateScope.value = isBookMemo(memo) ? 'book' : 'global'
  if (!isBookMemo(memo)) {
    currentGlobalMemoFolder.value = getGlobalMemoFolder(memo)
  }
  dialogSelectedMemo.value = { ...memo }
  if (memo.tags) {
    try {
      const tags = JSON.parse(memo.tags)
      if (Array.isArray(tags)) {
        dialogSelectedMemoTags.value = [...tags]
      } else {
        dialogSelectedMemoTags.value = []
      }
    } catch (e) {
      dialogSelectedMemoTags.value = []
    }
  } else {
    dialogSelectedMemoTags.value = []
  }
}

let isDialogResizing = false
let dialogResizeStartX = 0
let dialogResizeStartY = 0
let dialogResizeStartWidth = 0
let dialogResizeStartHeight = 0
let dialogResizeStartLeft = 0
let dialogResizeStartTop = 0
let dialogResizeStartTranslateX = 0
let dialogResizeStartTranslateY = 0
let memoDialogElement: HTMLElement | null = null
let dialogResizeFrame = 0
let pendingDialogResizeX = 0
let pendingDialogResizeY = 0

const getElementTranslate = (element: HTMLElement) => {
  const transform = window.getComputedStyle(element).transform
  if (!transform || transform === 'none') {
    return { x: 0, y: 0 }
  }

  const matrix = new DOMMatrix(transform)
  return {
    x: matrix.m41,
    y: matrix.m42
  }
}

const stabilizeGlobalMemoDialogPosition = async () => {
  await nextTick()

  const dialog = document.querySelector('.global-memo-dialog') as HTMLElement | null
  if (!dialog) return

  const rect = dialog.getBoundingClientRect()
  dialog.style.position = 'fixed'
  dialog.style.margin = '0'
  dialog.style.left = `${rect.left}px`
  dialog.style.top = `${rect.top}px`
}

const startDialogResize = (e: MouseEvent) => {
  isDialogResizing = true
  dialogResizeStartX = e.clientX
  dialogResizeStartY = e.clientY
  pendingDialogResizeX = e.clientX
  pendingDialogResizeY = e.clientY
  
  memoDialogElement = (e.currentTarget as HTMLElement | null)?.closest('.el-dialog') as HTMLElement | null
    ?? document.querySelector('.global-memo-dialog')
  if (memoDialogElement) {
    const rect = memoDialogElement.getBoundingClientRect()
    const translate = getElementTranslate(memoDialogElement)
    memoDialogElement.style.position = 'fixed'
    memoDialogElement.style.margin = '0'
    memoDialogElement.style.left = `${rect.left}px`
    memoDialogElement.style.top = `${rect.top}px`
    dialogResizeStartWidth = memoDialogElement.offsetWidth
    dialogResizeStartHeight = memoDialogElement.offsetHeight
    dialogResizeStartLeft = rect.left
    dialogResizeStartTop = rect.top
    dialogResizeStartTranslateX = translate.x
    dialogResizeStartTranslateY = translate.y
    memoDialogElement.style.transition = 'none'
    memoDialogElement.style.willChange = 'width, transform'
  }
  
  document.addEventListener('mousemove', handleDialogResize)
  document.addEventListener('mouseup', stopDialogResize)
  document.body.style.userSelect = 'none'
  document.body.style.cursor = 'se-resize'
  e.preventDefault()
}

const applyDialogResize = () => {
  if (!isDialogResizing || !memoDialogElement) return

  dialogResizeFrame = 0

  const deltaX = pendingDialogResizeX - dialogResizeStartX
  const deltaY = pendingDialogResizeY - dialogResizeStartY
  
  const screenWidth = window.innerWidth
  const screenHeight = window.innerHeight
  
  const minWidth = Math.max(600, screenWidth * 0.1)
  const maxWidth = Math.max(minWidth, screenWidth - Math.max(dialogResizeStartLeft, 0) - 24)
  const minHeight = Math.max(400, screenHeight * 0.1)
  const maxHeight = Math.max(minHeight, screenHeight - Math.max(dialogResizeStartTop, 0) - 24)
  
  const newWidth = Math.min(maxWidth, Math.max(minWidth, dialogResizeStartWidth + deltaX))
  const newHeight = Math.min(maxHeight, Math.max(minHeight, dialogResizeStartHeight + deltaY))
  
  memoDialogElement.style.width = newWidth + 'px'
  memoDialogElement.style.height = 'auto'
  
  const container = memoDialogElement.querySelector('.global-memo-container')
  if (container) {
    (container as HTMLElement).style.height = (newHeight - 120) + 'px'
  }
}

const handleDialogResize = (e: MouseEvent) => {
  if (!isDialogResizing || !memoDialogElement) return

  pendingDialogResizeX = e.clientX
  pendingDialogResizeY = e.clientY

  if (!dialogResizeFrame) {
    dialogResizeFrame = requestAnimationFrame(applyDialogResize)
  }
}

const stopDialogResize = () => {
  isDialogResizing = false
  if (dialogResizeFrame) {
    cancelAnimationFrame(dialogResizeFrame)
    dialogResizeFrame = 0
  }
  if (memoDialogElement) {
    memoDialogElement.style.willChange = ''
    memoDialogElement.style.transition = ''
  }
  document.body.style.userSelect = ''
  document.body.style.cursor = ''
  document.removeEventListener('mousemove', handleDialogResize)
  document.removeEventListener('mouseup', stopDialogResize)
}

watch(showGlobalMemoDialog, (visible) => {
  if (visible) {
    void stabilizeGlobalMemoDialogPosition()
  }
})

watch(globalMemoCustomFolders, () => {
  persistGlobalMemoFolders()
}, { deep: true })

const syncChapterState = (updatedChapter: Chapter) => {
  const index = chaptersList.value.findIndex(chapter => chapter.id === updatedChapter.id)
  if (index === -1) {
    if (currentChapter.value?.id === updatedChapter.id) {
      currentChapter.value = updatedChapter
    }
    return
  }

  chaptersList.value[index] = {
    ...chaptersList.value[index],
    ...updatedChapter
  }

  if (currentChapter.value?.id === updatedChapter.id) {
    currentChapter.value = chaptersList.value[index]
  }
}

const saveChapter = async () => {
  if (!currentChapter.value) return
  const res = await chapterAPI.update(currentChapter.value.id, {
    title: currentChapter.value.title,
    content: currentChapter.value.content,
    summary: currentChapter.value.summary || '',
    order_num: currentChapter.value.order_num
  })

  if (res.success && res.data) {
    syncChapterState(res.data)
  }
}

const saveChapterSummary = async (chapter: Chapter) => {
  try {
    const { value } = await ElMessageBox.prompt(
      '请输入该章节的概要内容，后续可以继续修改或清空。',
      '概要储存',
      {
        confirmButtonText: '保存',
        cancelButtonText: '取消',
        inputType: 'textarea',
        inputValue: chapter.summary || '',
        inputPlaceholder: '例如：本章的关键事件、冲突推进和结果',
        inputValidator: (inputValue) => {
          if (inputValue.length > 1000) {
            return '概要请控制在 1000 字以内'
          }
          return true
        }
      }
    )

    const summary = value.trim()
    const res = await chapterAPI.update(chapter.id, {
      title: chapter.title,
      content: chapter.content,
      summary,
      order_num: chapter.order_num
    })

    if (res.success && res.data) {
      syncChapterState(res.data)
      ElMessage.success(summary ? '概要已保存' : '概要已清空')
    }
  } catch (error) {
    // 取消保存
  }
}

const saveMemo = async () => {
  if (!currentMemo.value) return
  await memoAPI.update(currentMemo.value.id, {
    title: currentMemo.value.title,
    content: currentMemo.value.content
  })
}

const deleteChapter = async (chapter: Chapter) => {
  try {
    await ElMessageBox.confirm(`确定删除章节"${chapter.title}"吗？`, '提示', {
      type: 'warning'
    })
    const res = await chapterAPI.delete(chapter.id)
    if (res.success) {
      chaptersList.value = chaptersList.value.filter(c => c.id !== chapter.id)
      if (currentChapter.value?.id === chapter.id) {
        currentChapter.value = null
      }
      ElMessage.success('删除成功')
    }
  } catch (error) {
    // 取消删除
  }
}

const deleteMemo = async (memo: Memo) => {
  try {
    await ElMessageBox.confirm(`确定删除备忘录"${memo.title}"吗？`, '提示', {
      type: 'warning'
    })
    const res = await memoAPI.delete(memo.id)
    if (res.success) {
      memos.value = memos.value.filter(m => m.id !== memo.id)
      if (currentMemo.value?.id === memo.id) {
        currentMemo.value = null
      }
      ElMessage.success('删除成功')
    }
  } catch (error) {
    // 取消删除
  }
}

// 批量操作相关方法
const toggleBatchMode = () => {
  memoBatchMode.value = !memoBatchMode.value
  if (!memoBatchMode.value) {
    selectedMemoIds.value = []
  }
}

const toggleMemoSelection = (memoId: number) => {
  const index = selectedMemoIds.value.indexOf(memoId)
  if (index > -1) {
    selectedMemoIds.value.splice(index, 1)
  } else {
    selectedMemoIds.value.push(memoId)
  }
}

const toggleSelectAll = (checked: boolean) => {
  if (checked) {
    selectedMemoIds.value = filteredMemos.value.map(m => m.id)
  } else {
    selectedMemoIds.value = []
  }
}

const batchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selectedMemoIds.value.length} 个备忘录吗？`, '提示', {
      type: 'warning'
    })
    const res = await memoAPI.batch({
      action: 'delete',
      ids: selectedMemoIds.value
    })
    if (res.success) {
      memos.value = memos.value.filter(m => !selectedMemoIds.value.includes(m.id))
      selectedMemoIds.value = []
      memoBatchMode.value = false
      ElMessage.success('批量删除成功')
    }
  } catch (error) {
    // 取消删除
  }
}

const batchTogglePin = async () => {
  const res = await memoAPI.batch({
    action: 'toggle_pin',
    ids: selectedMemoIds.value
  })
  if (res.success) {
    selectedMemoIds.value.forEach(id => {
      const memo = memos.value.find(m => m.id === id)
      if (memo) {
        memo.is_pinned = 1 - (memo.is_pinned || 0)
      }
    })
    selectedMemoIds.value = []
    memoBatchMode.value = false
    ElMessage.success('批量操作成功')
  }
}

const handleMemoClick = (memo: Memo) => {
  if (memoBatchMode.value) {
    toggleMemoSelection(memo.id)
  } else {
    selectDialogMemo(memo)
  }
}

const toggleMemoPin = async (memo: Memo) => {
  const nextPinned = memo.is_pinned ? 0 : 1
  const res = await memoAPI.update(memo.id, { is_pinned: nextPinned })
  if (res.success) {
    const target = memos.value.find(item => item.id === memo.id)
    if (target) target.is_pinned = nextPinned
    if (dialogSelectedMemo.value?.id === memo.id) {
      dialogSelectedMemo.value.is_pinned = nextPinned
    }
    ElMessage.success(nextPinned ? '已置顶' : '已取消置顶')
  }
}

const getMemoTags = (tagsStr: string) => {
  try {
    const tags = JSON.parse(tagsStr)
    return Array.isArray(tags) ? tags.slice(0, 3) : [] // 最多显示 3 个标签
  } catch (e) {
    return []
  }
}

const confirmRelate = () => {
  relatedContent.value = []
  
  selectedChapters.value.forEach(id => {
    const chapter = chaptersList.value.find(c => c.id === id)
    if (chapter) {
      relatedContent.value.push({
        type: 'chapter',
        id: chapter.id,
        title: chapter.title,
        content: chapter.content
      })
    }
  })
  
  selectedMemos.value.forEach(id => {
    const memo = memos.value.find(m => m.id === id)
    if (memo) {
      relatedContent.value.push({
        type: 'memo',
        id: memo.id,
        title: memo.title,
        content: memo.content
      })
    }
  })
  
  showRelateDialog.value = false
  ElMessage.success(`已关联${relatedContent.value.length}项内容`)
}

const scrollToBottom = () => {
  nextTick(() => {
    if (chatMessagesRef.value) {
      chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight
    }
  })
}

const sendMessage = async () => {
  if (!userInput.value.trim()) {
    ElMessage.warning('请输入消息')
    return
  }

  if (!currentConversation.value) {
    ElMessage.warning('请先选择或创建对话')
    return
  }

  const userMessage: ChatMessage = {
    role: 'user',
    content: userInput.value
  }
  const pendingInput = userInput.value
  chatMessages.value.push(userMessage)
  
  // 保存用户消息
  await conversationAPI.saveMessage(currentConversation.value.id, {
    role: 'user',
    content: pendingInput
  })
  
  userInput.value = ''

  // 创建一个空的assistant消息用于流式更新
  const assistantMessage: ChatMessage = {
    role: 'assistant',
    content: ''
  }
  chatMessages.value.push(assistantMessage)
  const messageIndex = chatMessages.value.length - 1
  const selectedPromptRecords = prompts.value.filter((p: any) => selectedPrompts.value.some(id => id == p.id))
  const promptNames = selectedPromptRecords.map((p: any) => p.name).join('、')
  const systemPrompts = selectedPromptRecords.map((p: any) => {
    console.log('对话-添加提示词内容:', p.content)
    return p.content
  })
  const relatedContentSummary = relatedContent.value.length > 0
    ? [
        `关联内容：\n${relatedContent.value
          .map(item => `${item.type === 'chapter' ? '章节' : '备忘录'}：${item.title}\n${item.content}`)
          .join('\n\n')}`
      ]
    : []

  try {
    sending.value = true
    scrollToBottom()
    
    // 获取选中的提示词内容
    console.log('对话-选中的提示词IDs:', selectedPrompts.value, '所有提示词:', prompts.value)
    selectedPromptRecords.forEach((p: any) => {
      console.log('对话-匹配到提示词:', p.name, 'ID:', p.id)
    })
    console.log('对话-最终systemPrompts:', systemPrompts)

    // 使用 fetch 进行流式请求
    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        messages: chatMessages.value.slice(0, -1),
        configId: selectedConfigId.value,
        systemPrompts,
        relatedContent: relatedContent.value
      })
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const reader = response.body?.getReader()
    const decoder = new TextDecoder()

    if (!reader) {
      throw new Error('无法获取响应流')
    }

    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.slice(6)
          
          if (data === '[DONE]') {
            break
          }

          try {
            const parsed = JSON.parse(data)
            const currentMessage = chatMessages.value[messageIndex]
            if (parsed.content && currentMessage) {
              currentMessage.content += parsed.content
              scrollToBottom()
            }
            if (parsed.error) {
              ElMessage.error(parsed.error)
            }
          } catch (e) {
            console.error('解析SSE数据错误:', e)
          }
        }
      }
    }

    // 保存助手消息
    const savedMessage = chatMessages.value[messageIndex]
    if (savedMessage?.content) {
      await conversationAPI.saveMessage(currentConversation.value.id, {
        role: 'assistant',
        content: savedMessage.content
      })
      await fetchConversations()
      saveHistoryRecord({
        title: `正文对话 - ${currentConversation.value?.title || '新对话'}`,
        source: 'chat',
        sourceLabel: '正文对话',
        promptName: promptNames,
        promptCount: systemPrompts.length,
        status: 'completed',
        previewContent: savedMessage.content.substring(0, 140),
        messages: buildHistoryMessages({
          systemContents: relatedContentSummary,
          promptContents: systemPrompts,
          userContents: chatMessages.value
            .slice(0, -1)
            .filter(message => message.role === 'user')
            .map(message => message.content),
          assistantContents: [savedMessage.content]
        })
      })
    } else {
      chatMessages.value.pop()
      ElMessage.error('AI未返回内容')
      saveHistoryRecord({
        title: `正文对话 - ${currentConversation.value?.title || '新对话'}`,
        source: 'chat',
        sourceLabel: '正文对话',
        promptName: promptNames,
        promptCount: systemPrompts.length,
        status: 'failed',
        previewContent: pendingInput.substring(0, 140),
        messages: buildHistoryMessages({
          systemContents: relatedContentSummary,
          promptContents: systemPrompts,
          userContents: chatMessages.value
            .filter(message => message.role === 'user')
            .map(message => message.content)
        })
      })
    }

  } catch (error: any) {
    const partialAssistantContent = chatMessages.value[messageIndex]?.content || ''
    chatMessages.value.pop()
    saveHistoryRecord({
      title: `正文对话 - ${currentConversation.value?.title || '新对话'}`,
      source: 'chat',
      sourceLabel: '正文对话',
      promptName: promptNames,
      promptCount: systemPrompts.length,
      status: 'failed',
      previewContent: (partialAssistantContent || pendingInput).substring(0, 140),
      messages: buildHistoryMessages({
        systemContents: relatedContentSummary,
        promptContents: systemPrompts,
        userContents: chatMessages.value
          .filter(message => message.role === 'user')
          .map(message => message.content),
        assistantContents: partialAssistantContent ? [partialAssistantContent] : []
      })
    })
    ElMessage.error(error.message || '发送失败')
  } finally {
    sending.value = false
  }
}
</script>

<style scoped>
/* ========== 碧绿色科技风格 - 作品正文板块 ========== */

.write-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #f0fdf9 0%, #e6fffa 100%);
  overflow: hidden;
  animation: fadeIn 0.4s ease;
}

/* ========== 顶部工具栏 - 碧绿渐变 ========== */
.toolbar {
  padding: 0 18px;
  height: 64px;
  background: linear-gradient(135deg, #00c9a7 0%, #00a896 50%, #00877a 100%);
  color: #fff;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 4px 24px rgba(0, 201, 167, 0.25);
  position: relative;
  overflow: hidden;
  flex-shrink: 0;
  transition: height 0.1s ease;
}

.toolbar::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255, 255, 255, 0.05) 25%,
    rgba(255, 255, 255, 0.1) 50%,
    rgba(255, 255, 255, 0.05) 75%,
    transparent 100%
  );
  animation: shimmer 3s infinite;
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

.toolbar :deep(.el-breadcrumb) {
  position: relative;
  z-index: 1;
}

.toolbar :deep(.el-breadcrumb__inner) {
  color: rgba(255, 255, 255, 0.95);
  font-weight: 600;
  font-size: 15px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.toolbar :deep(.el-breadcrumb__inner:hover) {
  color: #fff;
  text-shadow: 0 0 12px rgba(255, 255, 255, 0.6);
}

.toolbar :deep(.el-breadcrumb__separator) {
  color: rgba(255, 255, 255, 0.7);
}

.toolbar-leading {
  display: flex;
  align-items: center;
  gap: 12px;
  position: relative;
  z-index: 1;
  min-width: 0;
}

.toolbar-round-btn {
  width: 42px;
  height: 42px;
  border: 1px solid rgba(255, 255, 255, 0.82);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.94);
  color: #009d82;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 8px 18px rgba(0, 92, 74, 0.12);
  transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}

.toolbar-round-btn:hover {
  transform: translateY(-2px);
  background: #ffffff;
  box-shadow: 0 10px 24px rgba(0, 92, 74, 0.18);
}

.toolbar-book-chip {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 0;
  padding: 8px 16px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.16);
  border: 1px solid rgba(255, 255, 255, 0.26);
  backdrop-filter: blur(12px);
}

.toolbar-book-name {
  color: #ffffff;
  font-size: 15px;
  font-weight: 700;
  line-height: 1.1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.toolbar-book-mode {
  color: rgba(255, 255, 255, 0.78);
  font-size: 11px;
  line-height: 1.2;
  margin-top: 4px;
  letter-spacing: 0.06em;
}

.toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  position: relative;
  z-index: 1;
  align-items: center;
  justify-content: flex-end;
}

.toolbar-status {
  padding: 0 2px 0 8px;
  color: rgba(255, 255, 255, 0.78);
  font-size: 13px;
  white-space: nowrap;
}

.character-btn,
.analysis-btn,
.ai-write-btn,
.history-btn {
  background: rgba(255, 255, 255, 0.92) !important;
  border: 1px solid rgba(255, 255, 255, 0.78) !important;
  color: #00a187 !important;
  padding: 9px 20px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  gap: 7px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  backdrop-filter: blur(12px);
  font-size: 14px;
  font-weight: 600;
  box-shadow: 0 8px 18px rgba(0, 92, 74, 0.12);
}

.character-btn:hover,
.analysis-btn:hover,
.ai-write-btn:hover,
.history-btn:hover {
  background: #ffffff !important;
  border-color: rgba(255, 255, 255, 0.96) !important;
  transform: translateY(-2px);
  box-shadow: 0 10px 24px rgba(0, 92, 74, 0.18);
}

.character-btn .el-icon,
.analysis-btn .el-icon,
.ai-write-btn .el-icon,
.history-btn .el-icon {
  font-size: 18px;
}

.toolbar-icon-btn {
  width: 42px;
  height: 42px;
  padding: 0 !important;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.92) !important;
  border: 1px solid rgba(255, 255, 255, 0.45) !important;
  color: #009d82 !important;
  backdrop-filter: blur(12px);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 8px 18px rgba(0, 92, 74, 0.12);
}

.toolbar-icon-btn:hover {
  background: rgba(255, 255, 255, 1) !important;
  border-color: rgba(255, 255, 255, 0.88) !important;
  color: #008f77 !important;
  transform: translateY(-2px);
  box-shadow: 0 10px 24px rgba(0, 92, 74, 0.18);
}

.toolbar-icon-btn .el-icon {
  font-size: 18px;
}

/* ========== 主内容区域 ========== */
.content-wrapper {
  flex: 1;
  display: flex;
  overflow: hidden;
  position: relative;
}

/* ========== 拖拽手柄 ========== */
.resize-handle {
  width: 4px;
  cursor: col-resize;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s ease;
  position: relative;
  z-index: 10;
  flex-shrink: 0;
}

.resize-handle:hover {
  background: rgba(0, 201, 167, 0.08);
}

.resize-handle:hover .resize-handle-bar {
  opacity: 1;
  background: linear-gradient(180deg, #00c9a7 0%, #00a896 100%);
}

.resize-handle-bar {
  width: 1px;
  height: 32px;
  background: rgba(0, 201, 167, 0.26);
  border-radius: 2px;
  transition: all 0.2s ease;
  opacity: 0.42;
}

.resize-handle:active .resize-handle-bar {
  background: linear-gradient(180deg, #00c9a7 0%, #00a896 100%);
  opacity: 1;
  box-shadow: 0 0 8px rgba(0, 201, 167, 0.4);
}

/* ========== 顶部工具栏拖拽手柄 ========== */
.toolbar-resize-handle {
  height: 8px;
  cursor: row-resize;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s ease;
  position: relative;
  z-index: 10;
  flex-shrink: 0;
  background: rgba(0, 201, 167, 0.05);
}

.toolbar-resize-handle:hover {
  background: rgba(0, 201, 167, 0.15);
}

.toolbar-resize-handle:hover .toolbar-resize-bar {
  opacity: 1;
  background: linear-gradient(90deg, #00c9a7 0%, #00a896 100%);
}

.toolbar-resize-bar {
  width: 40px;
  height: 2px;
  background: rgba(0, 201, 167, 0.3);
  border-radius: 2px;
  transition: all 0.2s ease;
  opacity: 0.5;
}

.toolbar-resize-handle:active .toolbar-resize-bar {
  background: linear-gradient(90deg, #00c9a7 0%, #00a896 100%);
  opacity: 1;
  box-shadow: 0 0 8px rgba(0, 201, 167, 0.4);
}

/* ========== 左侧目录面板 - 科技卡片风 ========== */
.left-panel {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 8px 6px 8px;
  background: linear-gradient(180deg, rgba(237, 252, 248, 0.76) 0%, rgba(226, 248, 242, 0.94) 100%);
  backdrop-filter: blur(14px);
  overflow: hidden;
}

.left-panel-header {
  padding: 16px 18px 12px;
  border: 1px solid rgba(0, 201, 167, 0.16);
  border-bottom: none;
  border-radius: 28px 28px 0 0;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(248, 255, 252, 0.98) 100%);
  box-shadow:
    0 24px 48px rgba(0, 136, 110, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.72);
}

.left-panel-book-title {
  text-align: center;
  font-size: 17px;
  font-weight: 600;
  color: #1e4f47;
  letter-spacing: 0.04em;
}

.catalog-header {
  padding: 10px 12px;
  border-bottom: 1px solid rgba(0, 201, 167, 0.12);
  background: rgba(255, 255, 255, 0.72);
}

.catalog-controls {
  display: flex;
  gap: 8px;
  align-items: center;
  justify-content: space-between;
}

.icon-btn {
  width: 34px;
  height: 34px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background: linear-gradient(135deg, rgba(0, 201, 167, 0.08) 0%, rgba(0, 168, 150, 0.05) 100%);
  border: 1px solid rgba(0, 201, 167, 0.15);
}

.icon-btn:hover {
  background: linear-gradient(135deg, rgba(0, 201, 167, 0.15) 0%, rgba(0, 168, 150, 0.1) 100%);
  border-color: rgba(0, 201, 167, 0.3);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 201, 167, 0.15);
}

.icon-btn .el-icon {
  font-size: 15px;
  color: #00a896;
}

.volume-container {
  margin-bottom: 8px;
}

.volume-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  margin-bottom: 4px;
  cursor: pointer;
  border-radius: 14px;
  background: rgba(244, 252, 250, 0.9);
  border: 1px solid rgba(0, 201, 167, 0.12);
  transition: all 0.2s ease;
}

.volume-item:hover {
  background: rgba(234, 249, 245, 0.98);
  border-color: rgba(0, 201, 167, 0.35);
  transform: translateX(2px);
  box-shadow: 0 6px 18px rgba(0, 201, 167, 0.12);
}

.volume-arrow {
  font-size: 13px;
  color: #606266;
  margin-right: 6px;
  transition: transform 0.2s ease;
}

.volume-arrow.expanded {
  transform: rotate(90deg);
}

.volume-title {
  flex: 1;
  font-weight: 500;
  color: #303133;
  font-size: 13px;
}

.volume-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.3s;
}

.volume-item:hover .volume-actions {
  opacity: 1;
}

.chapter-list {
  padding-left: 10px;
}

.empty-catalog {
  text-align: center;
  color: #909399;
  padding: 20px 0;
  font-size: 13px;
}

:root[data-theme='dark'] .empty-catalog {
  color: #6b7280 !important;
}

.catalog-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 10px 10px 14px;
}

.catalog-list::-webkit-scrollbar {
  width: 6px;
}

.catalog-list::-webkit-scrollbar-track {
  background: #f5f7f9;
}

:root[data-theme='dark'] .catalog-list::-webkit-scrollbar-track {
  background: #1a2332 !important;
}

.catalog-list::-webkit-scrollbar-thumb {
  background: #c0c4cc;
  border-radius: 3px;
}

:root[data-theme='dark'] .catalog-list::-webkit-scrollbar-thumb {
  background: #3a4558 !important;
}

:root[data-theme='dark'] .catalog-list::-webkit-scrollbar-thumb:hover {
  background: #4a5568 !important;
}

.catalog-item {
  padding: 10px 12px;
  margin-bottom: 6px;
  cursor: pointer;
  border-radius: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(0, 201, 167, 0.12);
  box-sizing: border-box;
}

:root[data-theme='dark'] .catalog-item {
  background: rgba(30, 41, 59, 0.8) !important;
  border-color: rgba(71, 85, 105, 0.4) !important;
}

.catalog-item:hover {
  background: rgba(239, 252, 249, 0.98);
  border-color: rgba(0, 201, 167, 0.28);
  transform: translateX(3px);
  box-shadow: 0 8px 18px rgba(0, 201, 167, 0.12);
}

:root[data-theme='dark'] .catalog-item:hover {
  background: rgba(51, 65, 85, 0.9) !important;
  border-color: rgba(94, 234, 212, 0.3) !important;
  box-shadow: 0 8px 18px rgba(94, 234, 212, 0.15) !important;
}

.catalog-item.active {
  background: linear-gradient(135deg, #00c9a7 0%, #00a896 100%);
  color: #fff;
  box-shadow: 0 12px 24px rgba(0, 168, 150, 0.24);
  border-color: rgba(0, 168, 150, 0.9);
  width: 100%;
  margin-left: 0;
  padding: 12px 14px;
  border-radius: 16px;
}

:root[data-theme='dark'] .catalog-item.active {
  background: linear-gradient(135deg, #0f766e 0%, #115e59 100%) !important;
  border-color: rgba(94, 234, 212, 0.4) !important;
  box-shadow: 0 12px 24px rgba(15, 118, 110, 0.3) !important;
}

.catalog-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 6px;
}

.catalog-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.catalog-meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.catalog-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
  font-size: 14px;
  transition: font-weight 0.2s ease;
  color: #1f2937;
}

:root[data-theme='dark'] .catalog-title {
  color: #f4f7ff !important;
}

.catalog-word-count,
.catalog-meta {
  flex-shrink: 0;
  font-size: 11px;
  line-height: 1;
  padding: 5px 8px;
  border-radius: 999px;
  background: rgba(0, 201, 167, 0.08);
  color: #5a8d85;
}

:root[data-theme='dark'] .catalog-word-count,
:root[data-theme='dark'] .catalog-meta {
  background: rgba(94, 234, 212, 0.12) !important;
  color: #5eead4 !important;
}

.catalog-item.active .catalog-title {
  font-weight: 600;
}

.catalog-item.active .catalog-word-count,
.catalog-item.active .catalog-meta {
  background: rgba(255, 255, 255, 0.16);
  color: rgba(255, 255, 255, 0.92);
}

.catalog-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
  opacity: 0;
  max-width: 90px;
  overflow: hidden;
  transform: translateX(6px);
  transition: opacity 0.2s ease, transform 0.2s ease;
  pointer-events: none;
}

.catalog-item:hover .catalog-actions,
.catalog-item.active .catalog-actions {
  opacity: 1;
  transform: translateX(0);
  pointer-events: auto;
}

.catalog-action-btn {
  width: 26px;
  height: 26px;
  padding: 0;
  border: none;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 168, 150, 0.08);
  color: #5b7773;
  cursor: pointer;
  transition: all 0.2s ease;
}

:root[data-theme='dark'] .catalog-action-btn {
  background: rgba(94, 234, 212, 0.12) !important;
  color: #5eead4 !important;
}

.catalog-action-btn .el-icon {
  font-size: 14px;
}

.catalog-action-btn:hover {
  transform: translateY(-1px);
}

:root[data-theme='dark'] .catalog-action-btn:hover {
  background: rgba(94, 234, 212, 0.2) !important;
}

.summary-action-btn.filled {
  background: rgba(0, 201, 167, 0.16);
  color: #00a896;
}

.summary-action-btn:hover {
  background: rgba(0, 201, 167, 0.18);
  color: #00a896;
}

.delete-action-btn:hover {
  background: rgba(245, 108, 108, 0.12);
  color: #f56c6c;
}

.catalog-item.active .catalog-action-btn {
  background: rgba(255, 255, 255, 0.18);
  color: rgba(255, 255, 255, 0.96);
}

.catalog-item.active .summary-action-btn.filled {
  background: rgba(255, 255, 255, 0.28);
}

.catalog-item.active .summary-action-btn:hover,
.catalog-item.active .delete-action-btn:hover {
  background: rgba(255, 255, 255, 0.28);
  color: #fff;
}

/* ========== 中间编辑器面板 - 极简科技风 ========== */
.center-panel {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: linear-gradient(180deg, rgba(237, 252, 248, 0.76) 0%, rgba(226, 248, 242, 0.94) 100%);
  backdrop-filter: blur(14px);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.55);
}

.editor-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 8px 6px 8px;
  animation: slideIn 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.editor-sheet {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(248, 255, 252, 0.98) 100%);
  border: 1px solid rgba(0, 201, 167, 0.16);
  border-radius: 28px;
  box-shadow:
    0 24px 48px rgba(0, 136, 110, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.72);
  overflow: hidden;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.editor-header {
  display: flex;
  align-items: center;
  gap: 16px;
  justify-content: space-between;
  flex-shrink: 0;
  padding: 18px 22px 10px;
}

.chapter-title {
  flex: 1;
  margin-bottom: 0;
}

.word-count-inline {
  flex-shrink: 0;
  font-size: 13px;
  color: #7aa79f;
  white-space: nowrap;
  font-family: 'Georgia', 'Times New Roman', serif;
  letter-spacing: 0.06em;
}

.chapter-title :deep(.el-input__wrapper) {
  font-size: 18px;
  font-weight: 500;
  border: none;
  box-shadow: none;
  padding: 4px 0;
  background: transparent;
  border-bottom: 1px solid transparent;
  transition: all 0.25s ease;
}

.chapter-title :deep(.el-input__wrapper):hover {
  box-shadow: none;
  border-bottom-color: rgba(0, 168, 150, 0.24);
}

.chapter-title :deep(.el-input__wrapper):focus-within {
  border-bottom-color: rgba(0, 168, 150, 0.86);
}

.chapter-title :deep(.el-input__inner) {
  color: #2f514b;
  font-weight: 500;
  font-family: 'Georgia', 'Times New Roman', serif;
  letter-spacing: 0.04em;
}

.editor-stats {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  gap: 12px;
  border-top: 1px solid rgba(0, 201, 167, 0.12);
  min-height: 32px;
}

.word-count {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: #6b9b97;
  white-space: nowrap;
  font-weight: 500;
}

.word-count .selected-count {
  color: #00c9a7;
  font-weight: 600;
}

.editor-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  flex: 1;
}

.editor-tools {
  display: flex;
  align-items: center;
  gap: 4px;
}

.editor-controls .el-button {
  width: 26px;
  height: 26px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgba(0, 201, 167, 0.08);
  border: 1px solid rgba(0, 201, 167, 0.15);
  color: #00a896;
  transition: all 0.3s;
}

.editor-controls .el-button:hover {
  background: rgba(0, 201, 167, 0.15);
  border-color: rgba(0, 201, 167, 0.3);
  color: #00c9a7;
  transform: scale(1.05);
}

.editor-controls .el-button .el-icon {
  font-size: 13px;
}

.font-size-control {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-left: auto;
}

.font-size-control .control-label {
  font-size: 13px;
  color: #6b9b97;
  white-space: nowrap;
  font-weight: 500;
}

.chapter-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.chapter-content :deep(.split-rich-text-editor) {
  flex: 1;
  display: flex;
  flex-direction: column;
  border: none;
  border-radius: 0;
  box-shadow: none;
  background: transparent;
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(247, 255, 252, 0.96) 0%, rgba(234, 250, 245, 0.92) 100%);
  border: 1px solid rgba(0, 201, 167, 0.18);
  border-radius: 28px;
}

/* ========== 右侧 AI 对话面板 - 悬浮卡片风 ========== */
.right-panel {
  flex-shrink: 0;
  border-left: 1px solid rgba(0, 201, 167, 0.15);
  display: flex;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(240, 253, 249, 0.92) 100%);
  backdrop-filter: blur(20px);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: -4px 0 24px rgba(0, 201, 167, 0.1);
  overflow: hidden;
}

.right-panel-2 {
  background: linear-gradient(180deg, rgba(255, 248, 240, 0.98) 0%, rgba(255, 240, 230, 0.92) 100%);
  box-shadow: -4px 0 24px rgba(255, 152, 0, 0.1);
  border-left-color: rgba(255, 152, 0, 0.15);
}

.knowledge-graph-dialog {
  :deep(.el-dialog__body) {
    padding: 0;
    height: 68vh;
    overflow: hidden;
  }
}

.graph-btn {
  background: rgba(255, 255, 255, 0.92) !important;
  border: 1px solid rgba(59, 130, 246, 0.3) !important;
  color: #3b82f6 !important;
  padding: 9px 20px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  gap: 7px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  backdrop-filter: blur(12px);
  font-size: 14px;
  font-weight: 600;
  box-shadow: 0 8px 18px rgba(59, 130, 246, 0.12);
}

.graph-btn:hover {
  background: rgba(59, 130, 246, 0.08) !important;
  border-color: rgba(59, 130, 246, 0.5) !important;
  box-shadow: 0 10px 24px rgba(59, 130, 246, 0.18);
  transform: translateY(-1px);
}

.ai-write-btn-2 {
  background: rgba(255, 255, 255, 0.92) !important;
  border: 1px solid rgba(255, 255, 255, 0.78) !important;
  color: #00a187 !important;
  padding: 9px 20px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  gap: 7px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  backdrop-filter: blur(12px);
  font-size: 14px;
  font-weight: 600;
  box-shadow: 0 8px 18px rgba(0, 92, 74, 0.12);
}

.ai-write-btn-2:hover {
  background: #ffffff !important;
  border-color: rgba(255, 255, 255, 0.96) !important;
  transform: translateY(-2px);
  box-shadow: 0 10px 24px rgba(0, 92, 74, 0.18);
}

.ai-write-btn-2 .el-icon {
  font-size: 18px;
}

/* ========== AI 创意区样式 ========== */
.creative-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
  background: linear-gradient(180deg, rgba(252, 255, 255, 0.7) 0%, rgba(244, 251, 250, 0.78) 100%);
}

.creative-header {
  padding: 12px 16px;
  border-bottom: 1px solid rgba(8, 198, 190, 0.16);
  display: flex;
  align-items: center;
  gap: 8px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.9) 0%, rgba(242, 251, 250, 0.88) 100%);
}

.creative-header-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.creative-title {
  flex: 1;
  font-weight: 700;
  font-size: 15px;
  color: #0f766e;
}

.creative-content {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.creative-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px 12px;
  background: rgba(255, 255, 255, 0.74);
  border: 1px solid rgba(8, 198, 190, 0.12);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(8, 198, 190, 0.08);
  backdrop-filter: blur(8px);
}

.fixed-height-section {
  height: 140px;
  min-height: 140px;
  max-height: 140px;
  overflow: hidden;
}

.fixed-height-section .prompt-picker-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.fixed-height-section .prompt-picker-selected {
  flex: 1;
  min-width: 0;
}

.fixed-height-section .prompt-picker-brief {
  flex-shrink: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  color: #4b5563;
  font-size: 13px;
  line-height: 1.5;
}

.fixed-height-section .prompt-picker-brief :deep(*) {
  color: #4b5563;
  font-size: 13px;
}

.fixed-height-section .prompt-picker-brief :deep(strong),
.fixed-height-section .prompt-picker-brief :deep(b) {
  font-weight: 600;
  color: #1f2937;
}

.fixed-height-section .prompt-picker-brief :deep(em),
.fixed-height-section .prompt-picker-brief :deep(i) {
  font-style: italic;
}

.fixed-height-section .prompt-picker-brief :deep(br) {
  line-height: 1.5;
}

.fixed-height-section .view-intro-icon-btn {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.section-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  color: #134e4a;
  font-size: 13px;
}

.section-label .el-icon {
  font-size: 15px;
  color: #08c6be;
}

.fixed-prompt-label {
  font-size: 12px;
  color: #999;
  margin-left: 4px;
}

.selected-prompts-preview {
  margin-top: 8px;
  padding: 8px;
  background: rgba(8, 198, 190, 0.08);
  border-radius: 6px;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.prompt-select-area {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid rgba(8, 198, 190, 0.15);
  border-radius: 8px;
  padding: 8px;
  background: rgba(227, 246, 243, 0.5);
}

.prompt-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 4px;
}

.prompt-item:hover {
  background: rgba(8, 198, 190, 0.1);
}

.prompt-item.selected {
  background: rgba(8, 198, 190, 0.15);
  border: 1px solid rgba(8, 198, 190, 0.3);
}

.prompt-info {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-width: 0;
}

.prompt-name {
  font-size: 13px;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-prompts-hint {
  text-align: center;
  color: #999;
  font-size: 13px;
  padding: 20px;
}

.result-area {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(8, 198, 190, 0.15);
  border-radius: 8px;
  padding: 12px;
  max-height: 300px;
  overflow-y: auto;
  font-size: 14px;
  line-height: 1.6;
}

.fields-input-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.field-input-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  color: #164e49;
  font-size: 14px;
}

.required-tag {
  margin-left: 4px;
  font-weight: 500;
  font-size: 11px;
}

.required-tag.el-tag--danger {
  background: rgba(248, 113, 113, 0.2);
  border-color: #f87171;
  color: #f87171;
}

.required-tag.el-tag--info {
  background: rgba(107, 114, 128, 0.2);
  border-color: #6b7280;
  color: #9ca3af;
}

.field-description {
  font-size: 12px;
  color: #62807c;
  line-height: 1.4;
}

.context-tip {
  font-size: 12px;
  color: #62807c;
  line-height: 1.5;
}

.creative-section .el-select .el-input__wrapper {
  background: rgba(255, 255, 255, 0.9);
  border-color: rgba(8, 198, 190, 0.2);
  padding: 6px 10px;
}

.creative-section .el-select .el-input__wrapper:hover {
  border-color: rgba(8, 198, 190, 0.4);
}

.creative-section .el-select .el-input__wrapper:focus-within {
  border-color: #08c6be;
  box-shadow: 0 0 0 2px rgba(8, 198, 190, 0.1);
}

.creative-section .el-textarea__inner {
  background: rgba(255, 255, 255, 0.9);
  border-color: rgba(8, 198, 190, 0.2);
  padding: 6px 10px;
  font-size: 13px;
}

.creative-section .el-textarea__inner:hover {
  border-color: rgba(8, 198, 190, 0.4);
}

.creative-section .el-textarea__inner:focus {
  border-color: #08c6be;
  box-shadow: 0 0 0 2px rgba(8, 198, 190, 0.1);
}

.creative-section .el-input__wrapper {
  background: rgba(255, 255, 255, 0.9);
  border-color: rgba(8, 198, 190, 0.2);
  padding: 6px 10px;
}

.creative-section .el-input__wrapper:hover {
  border-color: rgba(8, 198, 190, 0.4);
}

.creative-section .el-input__wrapper:focus-within {
  border-color: #08c6be;
  box-shadow: 0 0 0 2px rgba(8, 198, 190, 0.1);
}

.creative-section .el-button--primary {
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  border-color: transparent;
  box-shadow: 0 4px 12px rgba(8, 198, 190, 0.15);
  font-size: 13px;
  padding: 8px 16px;
}

.creative-section .el-button--primary:hover {
  background: linear-gradient(135deg, #09b4ad 0%, #047a75 100%);
  border-color: transparent;
}

.creative-section .el-button--primary.is-disabled {
  background: rgba(8, 198, 190, 0.5);
  border-color: transparent;
}

.view-intro-row {
  display: flex;
  justify-content: flex-start;
  margin-top: 8px;
}

.view-intro-btn {
  margin-top: 0;
}

.prompt-intro-content h3 {
  color: #1f2937;
  margin-bottom: 8px;
}

.prompt-intro-dialog :deep(.el-dialog) {
  max-height: 72vh;
  display: flex;
  flex-direction: column;
}

.prompt-intro-dialog :deep(.el-dialog__body) {
  flex: 1;
  min-height: 0;
}

.prompt-intro-content {
  max-height: 52vh;
  overflow-y: auto;
  padding-right: 6px;
}

.intro-text {
  color: #666;
  font-size: 14px;
  line-height: 1.8;
  padding: 16px;
  background: #f5f5f5;
  border-radius: 8px;
  white-space: pre-wrap;
}

.prompt-intro-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.prompt-intro-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.prompt-intro-name {
  margin: 0;
  color: #1f2937;
  font-size: 16px;
  font-weight: 600;
}

.prompt-intro-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.prompt-intro-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
  color: #374151;
  font-size: 13px;
  cursor: pointer;
  padding: 4px 0;
  transition: color 0.2s;
}

.prompt-intro-label:hover {
  color: #08c6be;
}

.collapse-icon {
  transition: transform 0.3s;
  font-size: 14px;
  color: #08c6be;
}

.collapse-icon.collapsed {
  transform: rotate(0deg);
}

.collapse-icon:not(.collapsed) {
  transform: rotate(90deg);
}

.prompt-intro-description {
  color: #4b5563;
  font-size: 13px;
  line-height: 1.5;
  padding: 8px 10px;
  background: rgba(8, 198, 190, 0.05);
  border-radius: 6px;
  border-left: 2px solid #08c6be;
}

.prompt-intro-description :deep(*) {
  color: #4b5563;
  font-size: 13px;
}

.prompt-intro-content-text {
  color: #374151;
  font-size: 13px;
  line-height: 1.6;
  padding: 10px 12px;
  background: #f9fafb;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
  white-space: pre-wrap;
  max-height: 300px;
  overflow-y: auto;
}

/* 收起时的展开按钮 */
.chat-toggle-btn {
  position: fixed;
  right: 24px;
  bottom: 24px;
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: linear-gradient(135deg, #00c9a7 0%, #00a896 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30px;
  cursor: pointer;
  box-shadow: 0 6px 24px rgba(0, 201, 167, 0.4);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 2px solid rgba(255, 255, 255, 0.3);
}

.chat-toggle-btn:hover {
  transform: scale(1.1) rotate(15deg);
  box-shadow: 0 8px 32px rgba(0, 201, 167, 0.5);
}

/* 对话历史对话框样式 */
.history-header {
  margin-bottom: 16px;
}

.history-list {
  max-height: 400px;
  overflow-y: auto;
}

.history-item {
  padding: 15px;
  margin-bottom: 10px;
  cursor: pointer;
  border-radius: 14px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.9) 0%, rgba(240, 253, 249, 0.7) 100%);
  border: 1px solid rgba(0, 201, 167, 0.1);
}

.history-item:hover {
  background: linear-gradient(135deg, rgba(0, 201, 167, 0.1) 0%, rgba(0, 168, 150, 0.06) 100%);
  border-color: rgba(0, 201, 167, 0.3);
  transform: translateX(4px);
  box-shadow: 0 4px 16px rgba(0, 201, 167, 0.15);
}

.history-item.active {
  background: linear-gradient(135deg, #00c9a7 0%, #00a896 100%);
  color: #fff;
  box-shadow: 0 6px 20px rgba(0, 201, 167, 0.35);
  border-color: rgba(0, 201, 167, 0.5);
}

.history-item.active .history-meta {
  color: rgba(255, 255, 255, 0.9);
}

.history-info {
  flex: 1;
  overflow: hidden;
}

.history-title {
  font-size: 14px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4px;
}

.history-meta {
  font-size: 12px;
  color: #88a8a5;
}

/* 对话主区域 */
.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: transparent;
}

.chat-header {
  padding: 12px 16px;
  border-bottom: 1px solid rgba(0, 201, 167, 0.12);
  display: flex;
  align-items: center;
  gap: 8px;
  background: linear-gradient(135deg, rgba(0, 201, 167, 0.04) 0%, rgba(0, 168, 150, 0.02) 100%);
}

.chat-title {
  flex: 1;
  font-weight: 700;
  font-size: 15px;
  color: #1a4a45;
}

.chat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  background: linear-gradient(180deg, rgba(240, 253, 249, 0.5) 0%, rgba(255, 255, 255, 0.3) 100%);
}

.message {
  margin-bottom: 12px;
  display: flex;
  gap: 8px;
  animation: fadeIn 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.message-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 16px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.message.user .message-avatar {
  background: linear-gradient(135deg, #00c9a7 0%, #00a896 100%);
  color: #fff;
  box-shadow: 0 4px 16px rgba(0, 201, 167, 0.35);
}

.message.assistant .message-avatar {
  background: linear-gradient(135deg, #2dd4bf 0%, #14b8a6 100%);
  color: #fff;
  box-shadow: 0 4px 16px rgba(45, 212, 191, 0.3);
}

.message-wrapper {
  flex: 1;
  max-width: 75%;
  position: relative;
}

.message-wrapper:hover .message-actions {
  opacity: 1;
}

.message-content {
  padding: 10px 14px;
  border-radius: 12px;
  word-wrap: break-word;
  line-height: 1.6;
  font-size: 14px;
}

.message.user .message-content {
  background: linear-gradient(135deg, #00c9a7 0%, #00a896 100%);
  color: #fff;
  border-bottom-right-radius: 4px;
  box-shadow: 0 6px 20px rgba(0, 201, 167, 0.35);
}

.message.user .message-content :deep(.markdown-body) {
  color: #fff;
}

.message.user .message-content :deep(.markdown-body code) {
  background-color: rgba(255, 255, 255, 0.2);
  color: #fff;
  border-radius: 4px;
  padding: 2px 6px;
}

.message.assistant .message-content {
  background: rgba(255, 255, 255, 0.95);
  color: #1a4a45;
  border-bottom-left-radius: 4px;
  box-shadow: 0 4px 16px rgba(0, 201, 167, 0.1);
  border: 1px solid rgba(0, 201, 167, 0.12);
}

.message-actions {
  display: flex;
  gap: 6px;
  margin-top: 8px;
  opacity: 0;
  transition: opacity 0.3s;
}

.message-actions .el-button {
  padding: 4px 10px;
  font-size: 13px;
}

.edit-wrapper {
  width: 100%;
}

.edit-wrapper .el-textarea {
  width: 100%;
}

.edit-wrapper :deep(.el-textarea__inner) {
  font-size: 15px;
  line-height: 1.8;
  border-radius: 10px;
  border: 1px solid rgba(0, 201, 167, 0.2);
}

.edit-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 12px;
}

.message-content.typing::after {
  content: '▊';
  animation: blink 1s infinite;
  margin-left: 2px;
  color: #00c9a7;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

.empty-chat {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(240, 253, 249, 0.5);
  border-radius: 16px;
}

.chat-input-area {
  border-top: 1px solid rgba(0, 201, 167, 0.12);
  padding: 12px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.95) 0%, rgba(240, 253, 249, 0.8) 100%);
  backdrop-filter: blur(10px);
}

.chat-tools {
  display: flex;
  margin-bottom: 8px;
  gap: 8px;
}

.chat-tools :deep(.el-select) {
  flex: 1;
}

.input-wrapper {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.input-wrapper .el-input {
  flex: 1;
}

.input-wrapper :deep(.el-textarea__inner) {
  resize: none;
  border-radius: 14px;
  padding: 14px 18px;
  font-size: 15px;
  line-height: 1.6;
  border: 1px solid rgba(0, 201, 167, 0.15);
  background: rgba(255, 255, 255, 0.9);
  transition: all 0.3s;
}

.input-wrapper :deep(.el-textarea__inner):focus {
  border-color: #00c9a7;
  box-shadow: 0 0 0 2px rgba(0, 201, 167, 0.1);
}

.input-wrapper :deep(.el-button) {
  height: 46px;
  padding: 0 24px;
  border-radius: 14px;
  background: linear-gradient(135deg, #00c9a7 0%, #00a896 100%);
  border: none;
  font-size: 15px;
  font-weight: 600;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.input-wrapper :deep(.el-button):hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 201, 167, 0.35);
}

/* ========== AI 写作生成结果弹窗样式 ========== */
.creative2-result-dialog :deep(.el-dialog) {
  border-radius: 12px;
  overflow: hidden;
}

.creative2-result-content {
  max-height: 60vh;
  display: flex;
  flex-direction: column;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: linear-gradient(135deg, #f0fdf9 0%, #e6f9f5 100%);
  border-bottom: 1px solid #eaeaea;
  margin: -16px -16px 16px -16px;
}

.result-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.result-icon {
  color: #00c9a7;
  font-size: 20px;
}

.result-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.result-header .el-button {
  color: #666;
}

.result-header .el-button:hover {
  color: #00c9a7;
}

.result-body {
  flex: 1;
  overflow: hidden;
}

.result-text {
  font-size: 14px;
  line-height: 1.8;
  color: #333;
  white-space: pre-wrap;
  word-break: break-word;
  padding: 0 4px;
}

.creative2-result-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 8px 0;
}

/* ========== 历史记录弹窗样式 ========== */
.history-dialog-modal :deep(.el-dialog) {
  border-radius: 12px;
  overflow: hidden;
}

.history-dialog-content {
  max-height: 70vh;
  overflow-y: auto;
}

.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 50px 20px;
  color: #999;
}

.loading-state .el-icon {
  font-size: 28px;
  margin-bottom: 12px;
  color: #00c9a7;
}

.empty-state .el-icon {
  margin-bottom: 12px;
  color: #ccc;
}

.empty-state p {
  margin-top: 10px;
  font-size: 13px;
}

.empty-state .hint {
  color: #999;
  font-size: 12px;
}

.history-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #eaeaea;
}

.history-count {
  font-size: 13px;
  font-weight: 600;
  color: #3a4a48;
}

.history-actions {
  display: flex;
  gap: 8px;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.history-item {
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 12px 14px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.history-item:hover {
  border-color: #00c9a7;
  background: #f9fffe;
  box-shadow: 0 2px 8px rgba(0, 201, 167, 0.1);
}

.history-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.history-main {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  flex: 1;
  min-width: 0;
}

.history-icon-wrapper {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: linear-gradient(135deg, #f0fdf9 0%, #e6f9f5 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.history-icon {
  color: #00c9a7;
  font-size: 16px;
}

.history-info {
  flex: 1;
  min-width: 0;
}

.history-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.history-title {
  font-size: 14px;
  font-weight: 600;
  color: #333;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-tag {
  flex-shrink: 0;
}

.history-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #999;
}

.meta-item {
  white-space: nowrap;
}

.meta-item.time {
  color: #aaa;
}

.meta-divider {
  color: #ddd;
  margin: 0 2px;
}

.history-actions-mini {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.history-actions-mini .el-button {
  padding: 4px 6px;
  font-size: 14px;
}

.history-preview-line {
  font-size: 12px;
  color: #666;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding-left: 42px;
}

.history-detail-content {
  max-height: 70vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.history-detail-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #eaeaea;
}

.history-detail-time {
  font-size: 12px;
  color: #999;
}

.history-conversations-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.history-conversation-item {
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.2s ease;
}

.history-conversation-item:hover {
  border-color: #00c9a7;
  box-shadow: 0 2px 8px rgba(0, 201, 167, 0.1);
}

.conversation-header {
  padding: 10px 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #f9f9f9;
  border-bottom: 1px solid #eaeaea;
}

.conversation-info {
  flex: 1;
}

.conversation-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.conversation-icon {
  color: #00c9a7;
  font-size: 16px;
  flex-shrink: 0;
}

.conversation-title {
  font-size: 14px;
  font-weight: 600;
  color: #333;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conversation-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #999;
  flex-wrap: wrap;
}

.conversation-time {
  display: flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
}

.conversation-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.conversation-actions .el-button {
  padding: 4px 8px;
  font-size: 12px;
}

.conversation-messages {
  padding: 0;
  background: #fafafa;
  max-height: 600px;
  overflow-y: auto;
  overflow-x: hidden;
  flex: 1;
  min-height: 0;
}

.messages-container {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-width: 0;
}

.message-bubble {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  max-width: 80%;
  min-width: 0;
}

.message-bubble.user {
  align-self: flex-end;
  flex-direction: row-reverse;
}

.message-bubble.assistant,
.message-bubble.prompt,
.message-bubble.system {
  align-self: flex-start;
}

.bubble-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: linear-gradient(135deg, #00c9a7 0%, #00a896 100%);
  color: #fff;
  font-size: 16px;
}

.message-bubble.user .bubble-avatar {
  background: linear-gradient(135deg, #4a90a4 0%, #3a7a94 100%);
}

.message-bubble.prompt .bubble-avatar {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
}

.message-bubble.system .bubble-avatar {
  background: linear-gradient(135deg, #7c8aa5 0%, #5f6b85 100%);
}

.bubble-content {
  flex: 1;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
}

.bubble-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 11px;
}

.bubble-role {
  font-weight: 600;
  color: #00c9a7;
}

.message-bubble.user .bubble-role {
  color: #4a90a4;
}

.message-bubble.prompt .bubble-role {
  color: #d97706;
}

.message-bubble.system .bubble-role {
  color: #5f6b85;
}

.bubble-time {
  color: #bbb;
  font-size: 10px;
}

.bubble-text {
  padding: 10px 12px;
  border-radius: 8px;
  font-size: 13px;
  line-height: 1.5;
  background: #fff;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  border: 1px solid #eaeaea;
  word-break: break-word;
  overflow-wrap: break-word;
  max-width: 100%;
  overflow-x: auto;
}

.message-bubble.user .bubble-text {
  background: #f0fdf9;
  border-color: #d4f4e8;
  border-bottom-right-radius: 4px;
}

.message-bubble.assistant .bubble-text,
.message-bubble.prompt .bubble-text,
.message-bubble.system .bubble-text {
  border-bottom-left-radius: 4px;
}

.message-bubble.prompt .bubble-text {
  background: linear-gradient(135deg, #fff7e6 0%, #fffbeb 100%);
  border-color: rgba(245, 158, 11, 0.2);
}

.message-bubble.system .bubble-text {
  background: linear-gradient(135deg, #f5f7fb 0%, #eef2f7 100%);
  border-color: rgba(95, 107, 133, 0.18);
}

.bubble-text :deep(.markdown-body) {
  font-size: 14px;
  line-height: 1.6;
}

.bubble-text :deep(.markdown-body p) {
  margin-bottom: 8px;
}

.bubble-text :deep(.markdown-body p:last-child) {
  margin-bottom: 0;
}

.bubble-text :deep(.markdown-body pre) {
  background: #f6f8fa;
  border-radius: 6px;
  padding: 12px;
  overflow-x: auto;
  margin: 8px 0;
  border: 1px solid rgba(0, 0, 0, 0.06);
}

.bubble-text :deep(.markdown-body code) {
  background: rgba(0, 201, 167, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
}

.message-bubble.user .bubble-text :deep(.markdown-body code) {
  background: rgba(0, 201, 167, 0.15);
}

.bubble-text :deep(.markdown-body pre code) {
  background: transparent;
  padding: 0;
}

.bubble-text :deep(.markdown-body ul),
.bubble-text :deep(.markdown-body ol) {
  padding-left: 20px;
  margin: 8px 0;
}

.bubble-text :deep(.markdown-body strong) {
  font-weight: 600;
  color: #333;
}

.message-bubble.user .bubble-text :deep(.markdown-body strong) {
  color: #1a4a45;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.input-wrapper :deep(.el-button):active {
  transform: translateY(0);
}

/* ========== Element Plus 组件覆盖 ========== */
.left-panel :deep(.el-tabs) {
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
  border: 1px solid rgba(0, 201, 167, 0.16);
  border-top: none;
  border-radius: 0 0 28px 28px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(248, 255, 252, 0.98) 100%);
  box-shadow:
    0 24px 48px rgba(0, 136, 110, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.72);
  overflow: hidden;
}

.left-panel :deep(.el-tabs__header) {
  display: none;
}

:deep(.el-tabs__content) {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

:deep(.el-tabs__header) {
  margin: 0;
  padding: 0 10px;
  background: transparent;
}

:deep(.el-tabs__item) {
  height: 44px;
  line-height: 44px;
  font-size: 14px;
  font-weight: 600;
  color: #6b9b97;
  border: none;
  transition: all 0.3s;
}

:deep(.el-tabs__item:hover) {
  color: #00c9a7;
}

:deep(.el-tabs__item.is-active) {
  color: #00c9a7;
  font-weight: 700;
}

:deep(.el-tabs__active-bar) {
  background: linear-gradient(90deg, #00c9a7 0%, #00a896 100%);
  height: 3px;
  border-radius: 3px;
}

:deep(.el-tab-pane) {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* 覆盖所有按钮样式 */
:deep(.el-button--primary),
.create-btn {
  background: linear-gradient(135deg, #00c9a7 0%, #00a896 100%) !important;
  border: none !important;
  color: #fff !important;
  font-weight: 600;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

:deep(.el-button--primary:hover),
.create-btn:hover {
  background: linear-gradient(135deg, #2dd4bf 0%, #14b8a6 100%) !important;
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 201, 167, 0.35);
}

:deep(.el-button--primary:active),
.create-btn:active {
  background: linear-gradient(135deg, #00a896 0%, #00877a 100%) !important;
  transform: translateY(0);
}

/* 对话框样式 */
:deep(.el-dialog) {
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 8px 40px rgba(0, 201, 167, 0.2);
}

:deep(.el-dialog__header) {
  background: linear-gradient(135deg, rgba(0, 201, 167, 0.05) 0%, rgba(0, 168, 150, 0.03) 100%);
  padding: 20px 24px;
  border-bottom: 1px solid rgba(0, 201, 167, 0.1);
}

:deep(.el-dialog__title) {
  color: #1a4a45;
  font-weight: 700;
  font-size: 18px;
}

:deep(.el-dialog__body) {
  padding: 24px;
}

:deep(.el-dialog__footer) {
  padding: 16px 24px;
  border-top: 1px solid rgba(0, 201, 167, 0.1);
  background: rgba(240, 253, 249, 0.3);
}

.global-memo-container {
  display: flex;
  height: 640px;
  min-height: 440px;
  border: 1px solid rgba(228, 213, 178, 0.45);
  border-radius: 14px;
  overflow: hidden;
  background: #fff;
}

.global-memo-dialog :deep(.el-dialog) {
  margin: 0 !important;
}

.global-memo-sidebar {
  width: 280px;
  border-right: 1px solid rgba(228, 232, 240, 0.95);
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #fffdfa 0%, #ffffff 100%);
  transition: width 0.25s ease;
}

.global-memo-sidebar.collapsed {
  width: 56px;
}

.memo-workspace-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px 8px;
}

.memo-workspace-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #243042;
  font-size: 13px;
  font-weight: 700;
}

.memo-workspace-title .el-icon {
  color: #4d8dff;
  font-size: 15px;
}

.sidebar-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px 10px;
}

.sidebar-action,
.sidebar-icon-btn {
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
}

.sidebar-action {
  height: 40px;
  padding: 0 12px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
}

.sidebar-action-primary {
  background: #fff4e6;
  color: #f28b24;
}

.sidebar-action-accent {
  background: #edf3ff;
  color: #3572ff;
}

.sidebar-icon-btn {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #f4f6fa;
  color: #556274;
  font-size: 12px;
  font-weight: 600;
}

.sidebar-icon-btn.active,
.sidebar-action:hover,
.sidebar-icon-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 18px rgba(31, 45, 61, 0.08);
}

.sidebar-collapsed-rail {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12px 0;
}

.batch-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 14px 10px;
  padding: 10px 12px;
  border-radius: 12px;
  background: #fff7ef;
  border: 1px solid #ffe0bf;
}

.memo-search-box {
  padding: 0 14px 10px;
}

.memo-section-card {
  margin-bottom: 0;
}

.memo-tree-root {
  padding: 4px;
}

.memo-tree-folder {
  margin-bottom: 2px;
}

.memo-tree-folder-header {
  width: 100%;
  padding: 8px 12px;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  border-radius: 6px;
  transition: all 0.2s ease;
  text-align: left;
}

.memo-tree-folder-header:hover {
  background: #f5f7fa;
}

.memo-tree-folder-header.active {
  background: #ecf5ff;
}

.memo-tree-folder-header.active .folder-name {
  color: #409eff;
  font-weight: 600;
}

.folder-header-content {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
}

.folder-arrow {
  color: #909399;
  font-size: 12px;
  transition: transform 0.2s ease;
  flex-shrink: 0;
}

.folder-icon {
  color: #e6a23c;
  font-size: 16px;
  flex-shrink: 0;
}

.folder-name {
  flex: 1;
  color: #606266;
  font-size: 13px;
  text-align: left;
}

.folder-count {
  color: #909399;
  font-size: 11px;
  background: #f4f4f5;
  padding: 2px 6px;
  border-radius: 8px;
  flex-shrink: 0;
}

.folder-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s ease;
  flex-shrink: 0;
}

.memo-tree-folder-header:hover .folder-actions {
  opacity: 1;
}

.folder-action-btn {
  width: 20px;
  height: 20px;
  padding: 0;
  border: none;
  background: transparent;
  color: #909399;
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.folder-action-btn:hover {
  background: #f5f7fa;
  color: #409eff;
}

.folder-action-btn.danger:hover {
  background: #fef0f0;
  color: #f56c6c;
}

.item-title-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.memo-tree-item:hover .item-title-actions {
  opacity: 1;
}

.item-title-btn {
  width: 20px;
  height: 20px;
  padding: 0;
  border: none;
  background: transparent;
  color: #909399;
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.item-title-btn:hover {
  background: #f5f7fa;
  color: #409eff;
}

.item-title-btn.danger:hover {
  background: #fef0f0;
  color: #f56c6c;
}

.move-folder-content {
  padding: 10px 0;
}

.move-folder-hint {
  margin: 0 0 16px;
  color: #606266;
  font-size: 13px;
}

.folder-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 400px;
  overflow-y: auto;
}

.folder-option {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 12px 16px;
  border: 1px solid #e4e7ed;
  background: #fff;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
}

.folder-option:hover {
  border-color: #409eff;
  background: #ecf5ff;
  transform: translateX(4px);
}

.folder-option.active {
  border-color: #409eff;
  background: #ecf5ff;
}

.folder-option .el-icon {
  color: #e6a23c;
  font-size: 18px;
}

.folder-option span {
  flex: 1;
  color: #606266;
  font-size: 14px;
}

.folder-option-count {
  color: #909399;
  font-size: 12px;
  background: #f4f4f5;
  padding: 4px 8px;
  border-radius: 10px;
  flex-shrink: 0;
}

.memo-tree-children {
  padding-left: 20px;
}

.memo-tree-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  margin: 2px 0;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.memo-tree-item:hover {
  background: #f5f7fa;
}

.memo-tree-item.active {
  background: #ecf5ff;
}

.memo-tree-item.active .item-title {
  color: #409eff;
  font-weight: 600;
}

.memo-tree-item.batch-selected {
  background: #fff7ef;
}

.batch-checkbox {
  margin-right: 12px;
}

.sidebar-item.batch-selected {
  background: #fff3e7;
}

.sidebar-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 14px 14px;
}

.sidebar-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 8px;
  color: #9bb8b5;
}

.sidebar-empty .el-icon {
  font-size: 32px;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  margin-bottom: 2px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.sidebar-item:hover {
  background: #f5f7fa;
}

.sidebar-item.active {
  background: #ecf5ff;
  color: #409eff;
}

.sidebar-item.active .item-title {
  color: #409eff;
  font-weight: 600;
}

.item-content {
  min-width: 0;
  flex: 1;
}

.item-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.item-title {
  display: block;
  font-size: 13px;
  font-weight: 400;
  color: #606266;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.memo-scope-tag {
  flex-shrink: 0;
  padding: 0 6px;
  border-radius: 4px;
  font-size: 11px;
  line-height: 18px;
  background: transparent;
  border: 1px solid;
}

.memo-scope-tag.global {
  color: #67c23a;
  border-color: #67c23a;
}

.memo-scope-tag.book {
  color: #409eff;
  border-color: #409eff;
}

.item-meta {
  display: block;
  font-size: 12px;
  color: #909399;
}

.item-meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 4px;
}

.item-tags-row {
  display: flex;
  gap: 4px;
  margin-top: 4px;
  flex-wrap: wrap;
}

.item-tag {
  height: 18px;
  line-height: 18px;
  font-size: 11px;
  background: #f4f4f5;
  color: #909399;
  border: none;
}

.item-meta-time {
  font-size: 10px;
  color: #b2bcc9;
}

.item-quick-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.sidebar-item:hover .item-quick-actions,
.sidebar-item.active .item-quick-actions {
  opacity: 1;
}

.item-icon-btn {
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 999px;
  background: #fff;
  color: #95a0af;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.item-icon-btn.pin {
  color: #f0a128;
}

.item-icon-btn.danger:hover {
  background: #fff1f2;
  color: #ef4444;
}

.global-memo-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #ffffff 0%, #fcfdff 100%);
}

.content-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 18px 22px;
}

.memo-content-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #eef1f5;
}

.memo-content-header-main {
  flex: 1;
  min-width: 0;
}

.memo-title-input :deep(.el-input__wrapper) {
  padding-left: 0;
  box-shadow: none !important;
  background: transparent;
  border: none;
  font-size: 17px;
  font-weight: 700;
}

.memo-content-meta {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-top: 8px;
  color: #9aa4b2;
  font-size: 12px;
  flex-wrap: wrap;
}

.memo-content-header-side {
  flex-shrink: 0;
}

.memo-scope-badge {
  display: inline-flex;
  align-items: center;
  padding: 5px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
}

.memo-scope-badge.global {
  color: #11a683;
  background: #e8faf4;
}

.memo-scope-badge.book {
  color: #4d86ff;
  background: #edf3ff;
}

.memo-content-input {
  flex: 1;
  margin-top: 14px;
}

.memo-content-input-large :deep(.el-textarea__inner) {
  height: 100%;
  min-height: 360px;
  padding: 0;
  resize: none;
  line-height: 1.8;
  border: none;
  box-shadow: none;
  font-size: 15px;
  color: #667085;
  background: transparent;
}

.content-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #eef1f5;
}

.content-footer.minimalist {
  justify-content: flex-end;
}

.footer-actions {
  display: flex;
  gap: 8px;
}

.content-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #9bb8b5;
}

.content-empty .el-icon {
  font-size: 48px;
}

/* 输入框样式 */
:deep(.el-input__wrapper) {
  border-radius: 10px;
  border: 1px solid rgba(203, 213, 225, 0.85);
  transition: all 0.3s;
}

:deep(.el-input__wrapper:hover) {
  border-color: rgba(77, 134, 255, 0.3);
}

:deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(77, 134, 255, 0.08);
  border-color: #4d86ff;
}

/* 复选框样式 */
:deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
  background-color: #4d86ff;
  border-color: #4d86ff;
}

:deep(.el-checkbox__inner) {
  border-radius: 4px;
  border: 1px solid rgba(77, 134, 255, 0.3);
}

/* 徽章样式 */
:deep(.el-badge__content.is-fixed) {
  background: linear-gradient(135deg, #2dd4bf 0%, #14b8a6 100%);
  border: 2px solid #fff;
}

/* 分类标签样式 */
.category-tabs-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px;
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.38);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.3) 0%, rgba(233, 255, 250, 0.18) 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.35);
}

.category-tabs {
  flex: 1;
  padding: 2px 2px 8px;
  overflow-x: auto;
  overflow-y: hidden;
  white-space: nowrap;
  scrollbar-width: thin;
  scrollbar-color: rgba(0, 201, 167, 0.35) transparent;
}

.category-tabs::-webkit-scrollbar {
  height: 6px;
}

.category-tabs::-webkit-scrollbar-track {
  background: transparent;
  border-radius: 999px;
}

.category-tabs::-webkit-scrollbar-thumb {
  background: rgba(0, 201, 167, 0.28);
  border-radius: 999px;
}

.category-tabs::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 201, 167, 0.42);
}

.category-tabs-inner {
  display: flex;
  gap: 8px;
  padding: 2px 0;
}

.category-tab {
  flex-shrink: 0;
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.42);
  border: 1px solid rgba(255, 255, 255, 0.55);
  border-radius: 999px;
  color: #2d5f5a;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.25s ease;
  white-space: nowrap;
  backdrop-filter: blur(12px);
  box-shadow: 0 8px 18px rgba(15, 118, 110, 0.08);
}

.category-tab:hover {
  background: rgba(255, 255, 255, 0.58);
  border-color: rgba(45, 212, 191, 0.32);
  color: #0f9f93;
  transform: translateY(-1px);
  box-shadow: 0 12px 24px rgba(15, 118, 110, 0.12);
}

.category-tab.active {
  background: linear-gradient(135deg, rgba(0, 201, 167, 0.92) 0%, rgba(13, 148, 136, 0.92) 100%);
  color: #fff;
  border-color: rgba(0, 201, 167, 0.3);
  box-shadow: 0 14px 28px rgba(0, 201, 167, 0.24);
}

.category-scroll-btn {
  flex-shrink: 0;
  border: 1px solid rgba(255, 255, 255, 0.52) !important;
  background: rgba(255, 255, 255, 0.42) !important;
  color: #0f766e !important;
  backdrop-filter: blur(12px);
  box-shadow: 0 10px 20px rgba(15, 118, 110, 0.1);
}

.category-scroll-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.62) !important;
  border-color: rgba(45, 212, 191, 0.34) !important;
  color: #0f9f93 !important;
  transform: translateY(-1px);
}

.category-scroll-btn:disabled {
  opacity: 0.45;
  box-shadow: none;
}

.prompt-select-dialog :deep(.el-dialog) {
  position: fixed;
  margin: 0 !important;
  max-height: calc(100vh - 48px);
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.48);
  border-radius: 28px;
  background:
    linear-gradient(145deg, rgba(247, 255, 253, 0.88) 0%, rgba(231, 252, 247, 0.82) 50%, rgba(224, 246, 244, 0.78) 100%);
  box-shadow:
    0 28px 60px rgba(15, 118, 110, 0.18),
    inset 0 1px 0 rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(24px);
}

:deep(.prompt-select-dialog-modal) {
  background: rgba(9, 30, 28, 0.24) !important;
  backdrop-filter: blur(14px);
}

.prompt-select-dialog :deep(.el-dialog__header) {
  margin-right: 0;
  padding: 22px 24px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.42);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.34) 0%, rgba(255, 255, 255, 0.14) 100%);
}

.prompt-select-dialog :deep(.el-dialog__title) {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: #154b45;
}

.prompt-select-dialog :deep(.el-dialog__headerbtn) {
  top: 18px;
  right: 18px;
}

.prompt-select-dialog :deep(.el-dialog__close) {
  color: #5b8d88;
}

.prompt-select-dialog :deep(.el-dialog__close:hover) {
  color: #0f9f93;
}

.prompt-select-dialog :deep(.el-dialog__body) {
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 140px);
  overflow: hidden;
  padding: 18px 24px 20px;
}

.prompt-select-dialog :deep(.el-dialog__footer) {
  padding: 14px 24px 22px;
  border-top: 1px solid rgba(255, 255, 255, 0.4);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0.26) 100%);
}

.prompt-select-dialog :deep(.el-dialog__footer .el-button) {
  min-width: 96px;
  border-radius: 12px;
}

.prompt-select-dialog :deep(.el-dialog__footer .el-button--default) {
  border-color: rgba(255, 255, 255, 0.52);
  background: rgba(255, 255, 255, 0.52);
  color: #285b56;
}

.prompt-select-dialog :deep(.el-dialog__footer .el-button--default:hover) {
  border-color: rgba(45, 212, 191, 0.34);
  background: rgba(255, 255, 255, 0.7);
  color: #0f9f93;
}

.prompt-select-dialog :deep(.el-dialog__footer .el-button--primary) {
  border-color: transparent;
  background: linear-gradient(135deg, #00c9a7 0%, #0f9f93 100%);
  box-shadow: 0 14px 24px rgba(0, 201, 167, 0.22);
}

.prompt-select-dialog :deep(.el-dialog__footer .el-button--primary:hover) {
  background: linear-gradient(135deg, #14d6b2 0%, #0c8f84 100%);
}

.prompt-select-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 0;
}

.prompt-select-toolbar {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px;
  border-radius: 22px;
  border: 1px solid rgba(255, 255, 255, 0.4);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.36) 0%, rgba(237, 255, 251, 0.2) 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.58),
    0 16px 30px rgba(15, 118, 110, 0.08);
  backdrop-filter: blur(18px);
}

.prompt-select-search-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.prompt-select-count {
  flex-shrink: 0;
  padding: 8px 12px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.52);
  background: rgba(255, 255, 255, 0.44);
  color: #4b7c77;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.02em;
  white-space: nowrap;
}

.search-input {
  flex: 1;
}

.search-input :deep(.el-input__wrapper) {
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.48);
  background: rgba(255, 255, 255, 0.55);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.58),
    0 10px 22px rgba(15, 118, 110, 0.08);
  transition: all 0.25s ease;
}

.search-input :deep(.el-input__wrapper:hover) {
  border-color: rgba(45, 212, 191, 0.28);
  background: rgba(255, 255, 255, 0.68);
}

.search-input :deep(.el-input__wrapper.is-focus) {
  border-color: rgba(0, 201, 167, 0.42);
  box-shadow:
    0 0 0 4px rgba(0, 201, 167, 0.12),
    0 14px 28px rgba(15, 118, 110, 0.12);
}

.search-input :deep(.el-input__inner) {
  color: #18453f;
}

.search-input :deep(.el-input__inner::placeholder) {
  color: #7ca19c;
}

.search-input :deep(.el-input__prefix-inner) {
  color: #59a59c;
}

.prompt-select-list {
  flex: 1;
  min-height: 0;
  max-height: calc(100vh - 280px);
  overflow-y: auto;
  padding: 4px 6px 4px 2px;
}

.prompt-select-list::-webkit-scrollbar {
  width: 6px;
}

.prompt-select-list::-webkit-scrollbar-track {
  background: rgba(0, 201, 167, 0.08);
  border-radius: 999px;
}

.prompt-select-list::-webkit-scrollbar-thumb {
  background: rgba(0, 201, 167, 0.28);
  border-radius: 999px;
}

.prompt-select-list::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 201, 167, 0.42);
}

.prompt-select-item {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  margin-bottom: 10px;
  padding: 14px 16px;
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.46);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.52) 0%, rgba(241, 255, 251, 0.34) 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.56),
    0 14px 26px rgba(15, 118, 110, 0.08);
  cursor: pointer;
  transition: all 0.25s ease;
  backdrop-filter: blur(14px);
}

.prompt-select-item:hover {
  transform: translateY(-2px);
  border-color: rgba(45, 212, 191, 0.28);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.62),
    0 18px 32px rgba(15, 118, 110, 0.12);
}

.prompt-select-item.selected {
  border-color: rgba(0, 201, 167, 0.34);
  background: linear-gradient(135deg, rgba(214, 252, 244, 0.82) 0%, rgba(232, 255, 251, 0.54) 100%);
  box-shadow:
    0 20px 34px rgba(0, 201, 167, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.68);
}

.prompt-select-item:last-child {
  margin-bottom: 0;
}

.prompt-select-info {
  flex: 1;
  min-width: 0;
}

.prompt-select-name {
  margin-bottom: 6px;
  color: #18453f;
  font-size: 15px;
  font-weight: 700;
  line-height: 1.4;
}

.prompt-select-desc {
  color: #6c948e;
  font-size: 13px;
  line-height: 1.6;
}

.prompt-category-tag {
  flex-shrink: 0;
  border-radius: 999px;
  border: 1px solid rgba(45, 212, 191, 0.2) !important;
  background: rgba(255, 255, 255, 0.5) !important;
  color: #0f766e !important;
}

.empty-prompts {
  min-height: 220px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 12px;
  border: 1px dashed rgba(45, 212, 191, 0.34);
  border-radius: 22px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.42) 0%, rgba(240, 255, 252, 0.24) 100%);
  color: #76a19c;
}

.empty-prompts .el-icon {
  color: #2db9ab;
}

.empty-prompts p {
  margin: 0;
  font-size: 14px;
}

@media (max-width: 768px) {
  .prompt-select-dialog :deep(.el-dialog) {
    width: calc(100vw - 24px) !important;
    max-height: calc(100vh - 24px);
    border-radius: 24px;
  }

  .prompt-select-dialog :deep(.el-dialog__header) {
    padding: 18px 18px 12px;
  }

  .prompt-select-dialog :deep(.el-dialog__body) {
    padding: 14px 18px 18px;
  }

  .prompt-select-dialog :deep(.el-dialog__footer) {
    padding: 12px 18px 18px;
  }

  .prompt-select-search-row {
    flex-direction: column;
    align-items: stretch;
  }

  .prompt-select-count {
    align-self: flex-start;
  }

  .prompt-select-item {
    padding: 12px 14px;
  }

  .prompt-category-tag {
    display: none;
  }
}

.relate-tabs {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

.relate-scrollbar {
  min-height: 0;
}

.relate-scrollbar :deep(.el-scrollbar__view) {
  padding-right: 8px;
}

.relate-option-item {
  margin: 5px 0;
}

.relate-dialog :deep(.el-dialog) {
  position: fixed;
  margin: 0 !important;
}

.relate-dialog :deep(.el-dialog__body) {
  display: flex;
  flex-direction: column;
  max-height: 60vh;
  overflow: hidden;
}

.relate-dialog :deep(.el-tabs) {
  flex: 1;
  min-height: 0;
}

.relate-dialog :deep(.el-tabs__content) {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.relate-dialog :deep(.el-tab-pane) {
  min-height: 0;
}

.prompt-picker-trigger {
  padding: 12px 14px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.55);
  border: 1px solid rgba(8, 198, 190, 0.18);
  cursor: pointer;
  transition: all 0.25s ease;
}

.prompt-picker-trigger:hover {
  background: rgba(255, 255, 255, 0.7);
  border-color: rgba(8, 198, 190, 0.28);
}

.prompt-picker-trigger-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  min-width: 0;
}

.prompt-picker-trigger-name {
  flex: 1;
  min-width: 0;
  color: #134e4a;
  font-size: 14px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.prompt-picker-trigger-placeholder {
  color: #64748b;
  font-size: 14px;
  font-weight: 500;
}

.prompt-picker-trigger-desc {
  margin-top: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #6b9b97;
  font-size: 12px;
  line-height: 1.5;
}

.prompt-picker-trigger-brief {
  margin-top: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #6b9b97;
  font-size: 12px;
  line-height: 1.4;
}

.prompt-picker-selected {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.55);
  border: 1px solid rgba(8, 198, 190, 0.18);
  cursor: pointer;
  transition: all 0.25s ease;
}

.prompt-picker-selected:hover {
  background: rgba(255, 255, 255, 0.7);
  border-color: rgba(8, 198, 190, 0.28);
}

.prompt-picker-name {
  flex: 1;
  min-width: 0;
  color: #134e4a;
  font-size: 14px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.prompt-picker-arrow {
  color: #08c6be;
  font-size: 16px;
}

.prompt-picker-brief {
  margin-top: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  color: #6b9b97;
  font-size: 12px;
  line-height: 1.4;
}

.view-intro-btn {
  margin-top: 8px;
}

/* 对话框缩放手柄 */
.global-memo-container {
  position: relative;
}

.dialog-resize-handle {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 24px;
  height: 24px;
  cursor: se-resize;
  z-index: 1000;
}

.dialog-resize-handle::after {
  content: '';
  position: absolute;
  bottom: 4px;
  right: 4px;
  width: 10px;
  height: 10px;
  border-right: 2px solid rgba(0, 201, 167, 0.5);
  border-bottom: 2px solid rgba(0, 201, 167, 0.5);
}

.dialog-resize-handle:hover::after {
  border-right-color: #00c9a7;
  border-bottom-color: #00c9a7;
}
</style>
