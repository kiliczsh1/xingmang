<template>
  <div class="write-container">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
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
        <el-button class="character-btn" @click="openCharacterLibrary">
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

    <!-- 三栏布局 -->
    <div class="content-wrapper">
      <!-- 左侧：目录 -->
      <div class="left-panel" :style="{ width: leftPanelWidth + 'px' }">
        <el-tabs v-model="catalogType">
          <el-tab-pane label="正文目录" name="chapters">
            <div class="catalog-header">
              <div class="catalog-controls">
                <el-button 
                  size="small" 
                  @click="handleCreateChapter(null)" 
                  class="icon-btn"
                  title="新建文件"
                >
                  <el-icon><Plus /></el-icon>
                </el-button>
                <el-button 
                  size="small" 
                  @click="handleCreateFolder(null)" 
                  class="icon-btn"
                  title="新建文件夹"
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
                <el-button
                  size="small"
                  @click="openImportChapterDialog"
                  class="icon-btn"
                  title="导入新章节（txt/docx，不覆盖现有章节）"
                >
                  <el-icon><Upload /></el-icon>
                </el-button>
              </div>
            </div>
            <div class="catalog-list">
              <template v-if="rootFolders.length === 0 && rootFiles.length === 0">
                <div class="empty-catalog">暂无章节</div>
              </template>
              <div
                v-else
                class="catalog-root-drop"
                :class="{ 'is-drop-target': dragOverTarget === 'root' }"
                @dragover.prevent="handleFolderDragOver('root')"
                @dragleave="handleFolderDragLeave('root')"
                @drop.prevent="handleFolderDrop('root')"
              >
                <div class="catalog-root-hint"></div>
                <WriteCatalogTree
                  :folders="sortedFolders"
                  :files="sortedChapters"
                  :expanded-ids="expandedFolderIds"
                  :current-file-id="currentChapter?.id || null"
                  :drag-over-key="dragOverTarget"
                  @toggle-folder="toggleFolder"
                  @select-file="selectChapter"
                  @create-folder="handleCreateFolder"
                  @create-file="handleCreateChapter"
                  @edit-folder="editFolder"
                  @delete-folder="deleteFolder"
                  @move-file="openMoveChapterDialog"
                  @delete-file="deleteChapter"
                  @drag-file-start="handleChapterDragStart"
                  @drag-file-end="handleChapterDragEnd"
                  @drag-over-folder="handleFolderDragOver"
                  @drag-leave-folder="handleFolderDragLeave"
                  @drop-folder="handleFolderDrop"
                />
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
              <FontSelector v-model="fontFamily" @change="saveFontFamily" />
              <FontSizeSelector v-model="fontSize" @change="saveFontSize" />
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
              :font-family="fontFamily"
              :font-size="fontSize"
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
              <FontSelector v-model="fontFamily" @change="saveFontFamily" />
              <FontSizeSelector v-model="fontSize" @change="saveFontSize" />
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
              :style="{ fontSize: fontSize + 'px', fontFamily: fontFamily }"
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

      <!-- 右侧：AI 对话（新 ChatPanel 组件） -->
      <ChatPanel v-if="showChatPanel" :book-id="bookId" :panel-width="rightPanelWidth" @close="showChatPanel = false" />

      <!-- 右侧拖拽手柄2 -->
      <div 
        class="resize-handle" 
        @mousedown="startResize('right2', $event)"
        title="拖拽调整宽度"
        v-if="showChatPanel2"
      >
        <div class="resize-handle-bar"></div>
      </div>

      <!-- 右侧：AI 写作  -->
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
              <div class="model-select-trigger creative-model-trigger" @click="openCreative2ModelSelectDialog">
                <span class="model-select-name">{{ creative2ModelName || '选择模型' }}</span>
                <el-icon class="model-select-arrow"><ArrowDown /></el-icon>
              </div>
            </div>

            <!-- 高级功能区域 -->
            <template v-if="creative2AdvancedMode">
              <!-- 2. 故事背景 -->
              <div class="creative-section">
                <div class="section-label">
                  <el-icon><EditPen /></el-icon>
                  故事背景
                </div>
                <div class="input-container">
                  <el-input
                    v-model="creative2StoryBackground"
                    maxlength="500"
                    show-word-limit
                    placeholder="输入故事背景设定..."
                  />
                  <el-button 
                    class="fullscreen-btn"
                    circle
                    size="small"
                    @click="openCreativeFullscreenEditor('storyBackground', creative2StoryBackground)"
                    title="全屏编辑"
                  >
                    <el-icon><FullScreen /></el-icon>
                  </el-button>
                </div>
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
                <div class="input-container">
                  <el-input
                    v-model="creative2CharacterRelations"
                    maxlength="500"
                    show-word-limit
                    placeholder="输入本章涉及的角色及其关系..."
                  />
                  <el-button 
                    class="fullscreen-btn"
                    circle
                    size="small"
                    @click="openCreativeFullscreenEditor('characterRelations', creative2CharacterRelations)"
                    title="全屏编辑"
                  >
                    <el-icon><FullScreen /></el-icon>
                  </el-button>
                </div>
              </div>

              <!-- 5. 本章剧情 -->
              <div class="creative-section">
                <div class="section-label">
                  <el-icon><ChatLineSquare /></el-icon>
                  本章剧情
                </div>
                <div class="input-container">
                  <el-input
                    v-model="creative2ChapterPlot"
                    maxlength="3000"
                    show-word-limit
                    placeholder="输入本章剧情概要..."
                  />
                  <el-button 
                    class="fullscreen-btn"
                    circle
                    size="small"
                    @click="openCreativeFullscreenEditor('chapterPlot', creative2ChapterPlot)"
                    title="全屏编辑"
                  >
                    <el-icon><FullScreen /></el-icon>
                  </el-button>
                </div>
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
                <div class="input-container">
                  <el-input
                    v-model="creative2AdditionalInfo"
                    maxlength="500"
                    show-word-limit
                    placeholder="输入本次生成的额外要求、风格偏好或限制条件..."
                  />
                  <el-button 
                    class="fullscreen-btn"
                    circle
                    size="small"
                    @click="openCreativeFullscreenEditor('additionalInfo', creative2AdditionalInfo)"
                    title="全屏编辑"
                  >
                    <el-icon><FullScreen /></el-icon>
                  </el-button>
                </div>
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
          </div>
        </div>
        
        <!-- 开始生成按钮 - 固定在右下角 -->
        <div class="creative-generate-btn-wrapper">
          <el-button 
            type="primary" 
            :loading="creative2Generating"
            @click="handleCreative2Generate"
            class="creative-generate-btn"
          >
            <el-icon><Lightning /></el-icon>
            开始生成
          </el-button>
        </div>
      </div>
    </div>

    <!-- 知识图谱弹窗 -->
    <el-dialog
      v-model="showGraphPanel"
      title="知识图谱"
      width="1200px"
      top="0"
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
      :title="creative2WaitingForResponse ? 'AI 写作生成中...' : (creative2Generating ? 'AI 写作生成中' : 'AI 写作生成结果')"
      width="800px"
      align-center
      append-to-body
      class="creative2-result-dialog"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="!creative2Generating"
    >
      <div class="creative2-result-content">
        <!-- 等待AI响应时的过场动画 -->
        <div v-if="creative2WaitingForResponse" class="generating-animation">
          <div class="starburst-container">
            <div class="starburst-core"></div>
            <div class="starburst-ring ring-1"></div>
            <div class="starburst-ring ring-2"></div>
            <div class="starburst-ring ring-3"></div>
            <div class="starburst-ray ray-1"></div>
            <div class="starburst-ray ray-2"></div>
            <div class="starburst-ray ray-3"></div>
            <div class="starburst-ray ray-4"></div>
            <div class="starburst-ray ray-5"></div>
            <div class="starburst-ray ray-6"></div>
            <div class="starburst-ray ray-7"></div>
            <div class="starburst-ray ray-8"></div>
            <div class="starburst-sparkle sparkle-1"></div>
            <div class="starburst-sparkle sparkle-2"></div>
            <div class="starburst-sparkle sparkle-3"></div>
            <div class="starburst-sparkle sparkle-4"></div>
            <div class="starburst-sparkle sparkle-5"></div>
            <div class="starburst-sparkle sparkle-6"></div>
          </div>
          <div class="generating-text">正在生成中，请稍候...</div>
        </div>

        <!-- 流式输出内容区域 -->
        <div v-else-if="creative2Generating && creative2Result" class="streaming-content">
          <div class="result-body">
            <el-scrollbar ref="streamingScrollbarRef" max-height="50vh">
              <MarkdownRenderer :content="creative2Result" />
            </el-scrollbar>
          </div>
        </div>

        <!-- 生成完成内容区域 -->
        <div v-else-if="creative2Result" class="result-complete">
          <div class="result-body">
            <el-scrollbar max-height="50vh">
              <MarkdownRenderer :content="creative2Result" />
            </el-scrollbar>
          </div>
        </div>

        <!-- 空状态 -->
        <div v-else class="empty-result">
          <el-empty description="暂无生成内容" />
        </div>
      </div>
      <template #footer>
        <div class="creative2-result-footer">
          <el-button 
            v-if="!creative2Generating" 
            @click="creative2ResultDialogVisible = false"
          >
            关闭
          </el-button>
          <el-button 
            v-if="creative2Generating"
            type="danger" 
            @click="stopCreative2Generation"
          >
            <el-icon><CircleClose /></el-icon>
            停止生成
          </el-button>
          <el-button 
            v-if="!creative2Generating"
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

    <!-- 世界书管理对话框 -->
    <el-dialog
      v-model="showRelateDialog"
      :title="`世界书管理 - ${currentBook?.title || '当前书籍'}`"
      width="900px"
      top="3vh"
      align-center
      append-to-body
      class="worldbook-dialog"
    >
      <div class="worldbook-manager">
        <div class="worldbook-toolbar">
          <div class="toolbar-left">
            <el-input
              v-model="worldbookSearchKeyword"
              placeholder="搜索条目..."
              clearable
              :prefix-icon="Search"
              style="width: 300px"
            />
            <div class="worldbook-stats-inline">
              <span>共 {{ worldBookEntryCount }} 条</span>
              <span>已启用 {{ enabledWorldBookCount }} 条</span>
            </div>
          </div>
          <div class="toolbar-right">
            <el-button size="small" @click="openCreateGroupDialog">
              <el-icon><Folder /></el-icon>
              新建分组
            </el-button>
            <el-dropdown @command="handleImportCommand" trigger="click">
              <el-button size="small">
                <el-icon><Upload /></el-icon>
                导入
                <el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="current">
                    <el-icon><Plus /></el-icon>
                    导入到当前世界书
                  </el-dropdown-item>
                  <el-dropdown-item command="new">
                    <el-icon><FolderAdd /></el-icon>
                    导入为新世界书
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <el-dropdown @command="handleExportFormat" trigger="click">
              <el-button size="small">
                <el-icon><Download /></el-icon>
                导出
                <el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="default">
                    <el-icon><Document /></el-icon>
                    导出为本项目格式
                  </el-dropdown-item>
                  <el-dropdown-item command="sillytavern">
                    <el-icon><Connection /></el-icon>
                    导出为酒馆格式
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <el-button type="primary" size="small" @click="createWorldBookEntry">
              <el-icon><Plus /></el-icon>
              新建条目
            </el-button>
          </div>
        </div>

        <!-- 世界书选择器 -->
        <div class="worldbook-selector">
          <el-select
            v-model="currentWorldBookName"
            placeholder="选择世界书"
            size="small"
            style="width: 220px"
            @change="collapsedGroups.clear()"
          >
            <el-option
              v-for="name in worldBookList"
              :key="name"
              :label="name"
              :value="name"
            />
          </el-select>
          <el-button size="small" @click="openCreateWorldBookDialog">
            <el-icon><Plus /></el-icon>
            新建
          </el-button>
          <el-button size="small" @click="openRenameWorldBookDialog" :disabled="currentWorldBookName === '默认世界书'">
            <el-icon><Edit /></el-icon>
            重命名
          </el-button>
          <el-button size="small" type="danger" @click="handleDeleteWorldBook">
            <el-icon><Delete /></el-icon>
            删除
          </el-button>
        </div>

        <div class="worldbook-list-container">
          <div class="worldbook-grouped-list">
            <div v-if="worldBookGroups.length === 0" class="worldbook-empty-state">
              <el-empty description="暂无世界书条目">
                <el-button type="primary" @click="createWorldBookEntry">创建第一个条目</el-button>
              </el-empty>
            </div>
            <template v-else>
              <div
                v-for="[groupName, entries] in worldBookGroups"
                :key="groupName"
                class="wb-group-section"
              >
                <div class="wb-group-header" @click="toggleGroupCollapse(groupName)">
                  <el-icon class="group-collapse-icon">
                    <component :is="collapsedGroups.has(groupName) ? 'ArrowRight' : 'ArrowDown'" />
                  </el-icon>
                  <span class="wb-group-name">{{ groupName || '默认' }}</span>
                  <span class="wb-group-count">{{ entries.length }}条</span>
                  <div class="wb-group-actions" @click.stop>
                    <el-button
                      v-if="groupName"
                      size="small"
                      text
                      @click="openRenameGroupDialog(groupName)"
                    >
                      <el-icon><Edit /></el-icon>
                    </el-button>
                    <el-button
                      v-if="groupName"
                      size="small"
                      text
                      type="danger"
                      @click="deleteGroup(groupName)"
                    >
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                </div>
                <div v-show="!collapsedGroups.has(groupName)" class="wb-group-body">
                  <div class="wb-entry-grid">
                    <div
                      v-for="entry in entries"
                      :key="entry.uid"
                      :class="['wb-entry-card', { disabled: entry.disable }]"
                    >
                      <div class="wb-entry-top">
                        <el-switch
                          :model-value="!entry.disable"
                          size="small"
                          @click.stop
                          @change="(val: boolean) => toggleWorldBookEntry(entry, val)"
                        />
                        <div class="wb-entry-tags">
                          <el-tag v-if="entry.constant" size="small" type="success">常驻</el-tag>
                          <el-tag v-if="entry.disable" size="small" type="danger">禁用</el-tag>
                        </div>
                      </div>
                      <div class="wb-entry-title">{{ entry.comment || '未命名条目' }}</div>
                      <div class="wb-entry-actions">
                        <el-button size="small" @click="editWorldBookEntry(entry)" circle>
                          <el-icon><Edit /></el-icon>
                        </el-button>
                        <el-button size="small" type="danger" @click="deleteWorldBookEntry(entry)" circle>
                          <el-icon><Delete /></el-icon>
                        </el-button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </template>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 分组管理弹窗 -->
    <el-dialog
      v-model="showGroupDialog"
      :title="groupDialogMode === 'create' ? '新建分组' : '重命名分组'"
      width="400px"
      align-center
      append-to-body
      destroy-on-close
    >
      <el-input
        v-model="groupDialogName"
        placeholder="输入分组名称"
        @keydown.enter="confirmGroupDialog"
      />
      <template #footer>
        <el-button @click="showGroupDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmGroupDialog">确定</el-button>
      </template>
    </el-dialog>

    <!-- 世界书创建/重命名弹窗 -->
    <el-dialog
      v-model="showWorldBookDialog"
      :title="worldBookDialogMode === 'create' ? '新建世界书' : '重命名世界书'"
      width="400px"
      align-center
      append-to-body
      destroy-on-close
    >
      <el-input
        v-model="worldBookDialogName"
        placeholder="输入世界书名称"
        @keydown.enter="confirmWorldBookDialog"
      />
      <template #footer>
        <el-button @click="showWorldBookDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmWorldBookDialog">确定</el-button>
      </template>
    </el-dialog>

    <!-- 世界书条目编辑对话框 -->
    <el-dialog
      v-model="worldBookEditorVisible"
      :title="editingWorldBookEntry ? '编辑条目' : '新建条目'"
      width="800px"
      top="3vh"
      append-to-body
      destroy-on-close
      :close-on-click-modal="false"
    >
      <div class="worldbook-editor-form">
        <div class="editor-field">
          <label class="field-label">标题/备注</label>
          <el-input v-model="worldBookForm.comment" placeholder="输入条目标题或备注" />
        </div>

        <div class="editor-field">
          <label class="field-label">内容</label>
          <el-input
            v-model="worldBookForm.content"
            type="textarea"
            :rows="8"
            placeholder="触发后注入的内容，支持 Markdown 格式"
          />
        </div>

        <div class="editor-row">
          <div class="editor-field half">
            <label class="field-label">插入位置</label>
            <el-select v-model="worldBookForm.position" placeholder="选择插入位置">
              <el-option :value="0" label="在提示词前" />
              <el-option :value="1" label="在提示词后" />
              <el-option :value="2" label="在作者注释后" />
              <el-option :value="3" label="在角色描述后" />
              <el-option :value="7" label="世界书顶部" />
              <el-option :value="8" label="世界书底部" />
            </el-select>
          </div>

          <div class="editor-field half">
            <label class="field-label">优先级</label>
            <el-input-number v-model="worldBookForm.order" :min="0" :max="1000" />
          </div>
        </div>

        <div class="editor-row">
          <div class="editor-field half">
            <label class="field-label">扫描深度</label>
            <el-input-number v-model="worldBookForm.depth" :min="1" :max="999" />
          </div>
        </div>

        <div class="editor-field">
          <label class="field-label">选项</label>
          <div class="checkbox-group">
            <label class="field-label">状态</label>
            <el-radio-group v-model="worldBookEntryStatus" size="small">
              <el-radio value="normal">启用</el-radio>
              <el-radio value="constant">常驻激活</el-radio>
              <el-radio value="disable">禁用</el-radio>
            </el-radio-group>
          </div>
        </div>

        <div class="editor-field">
          <label class="field-label">分组</label>
          <el-select
            v-model="worldBookForm.group"
            filterable
            allow-create
            clearable
            placeholder="选择或输入分组名称"
            style="width: 100%;"
          >
            <el-option
              v-for="g in availableGroups"
              :key="g"
              :label="g || '默认'"
              :value="g"
            />
          </el-select>
        </div>

      </div>

      <template #footer>
        <el-button @click="worldBookEditorVisible = false">取消</el-button>
        <el-button type="primary" @click="saveWorldBookEntry">
          {{ editingWorldBookEntry ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <input
      ref="worldBookFileInput"
      type="file"
      accept=".json"
      style="display: none"
      @change="handleWorldBookFileSelect"
    />
    <input
      ref="importAsNewFileInput"
      type="file"
      accept=".json"
      style="display: none"
      @change="handleWorldBookImportAsNew"
    />

    <!-- 移动文件夹对话框 -->
    <el-dialog
      v-model="moveFolderDialogVisible"
      title="移动到文件夹"
      width="400px"
      class="move-folder-dialog"
      append-to-body
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
                      @click="toggleMemoFolder(folder)"
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
                              @click="deleteMemoFolder(folder)"
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
    <RenameDialog
      v-model:visible="showRenameDialog"
      :conversation="renamingConversation"
      :current-title="renamingConversation?.title || ''"
      @renamed="handleConversationRenamed"
    />

    <!-- 提示词选择弹窗 -->
    <PromptSelectDialog
      v-model:visible="promptSelectDialogVisible"
      v-model:model-value="selectedPrompts"
      :prompts="prompts"
      mode="multi"
      :show-categories="true"
      :show-footer="false"
    />

    <!-- 抽卡区提示词选择弹窗 -->
    <PromptSelectDialog
      v-model:visible="creative2PromptDialogVisible"
      v-model:model-value="creative2SelectedPrompts"
      :prompts="prompts"
      title="写作风格"
      mode="multi"
      category-filter="写作风格"
      :show-categories="false"
      :show-footer="true"
    />

    <!-- 抽卡区第二个提示词选择弹窗 -->
    <PromptSelectDialog
      v-model:visible="creative2SecondPromptDialogVisible"
      v-model:model-value="creative2SecondPromptId"
      :prompts="prompts"
      title="写作要求"
      mode="single"
      category-filter="写作要求"
      :show-categories="false"
      :show-footer="true"
    />

    <!-- 选择提示词1介绍弹窗 -->
    <el-dialog
      v-model="creative2PromptIntroDialogVisible"
      title="提示词介绍"
      width="600px"
      destroy-on-close
      class="prompt-intro-dialog"
      append-to-body
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
      append-to-body
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

    <!-- 文件夹对话框 -->
    <VolumeDialog
      v-model:visible="showVolumeDialog"
      :book-id="bookId"
      :is-edit="isEditVolume"
      :volume-id="volumeForm.id"
      :initial-title="volumeForm.title"
      :parent-id="editingFolderParentId"
      :volumes="volumes"
      :expanded-folder-ids="expandedFolderIds"
      @submitted="handleVolumeSubmitted"
    />

    <MoveChapterDialog
      v-model:visible="moveChapterDialogVisible"
      :folders="sortedFolders"
      :chapter-to-move="chapterToMove"
      @moved="handleChapterMoved"
    />

    <!-- 历史记录对话框 -->
    <HistoryDialog
      v-model:visible="showHistoryDialog"
      :book-id="bookId"
    />

    <ConversationListDialog
      v-model:visible="conversationListVisible"
      :conversations="conversations"
      :current-conversation-id="currentConversation?.id"
      @select="selectConversation"
      @rename="renameConversation"
      @delete="deleteConversation"
    />

    <!-- 模型选择弹窗 -->
    <ModelSelectDialog
      v-model:visible="modelSelectDialogVisible"
      v-model:temperature="modelTemperature"
      v-model:top-p="modelTopP"
      :api-configs="apiConfigs"
      :selected-config-id="modelSelectMode === 'creative2' ? creative2ConfigId : selectedConfigId"
      :model-select-mode="modelSelectMode"
      @confirm="handleModelSelectConfirm"
    />

    <!-- 模型配置弹窗 -->
    <el-dialog
      v-model="modelConfigDialogVisible"
      title="模型配置"
      width="780px"
      align-center
      append-to-body
      destroy-on-close
      class="model-config-dialog"
    >
      <div class="model-config-layout">
        <div class="model-list-panel">
          <div class="model-list-title">模型目录</div>
          <div
            v-for="config in apiConfigs"
            :key="config.id"
            :class="['model-list-item', { active: editingModelId === config.id }]"
            @click="selectModelForEdit(config)"
          >
            <div class="model-item-name">{{ config.name }}</div>
            <div class="model-item-provider">{{ config.provider_name }}</div>
          </div>
        </div>
        <div class="model-config-panel" v-if="editingModelData">
          <el-form label-width="100px" size="small">
            <el-form-item label="模型名称">
              <el-input v-model="editingModelData.name" placeholder="模型名称" />
            </el-form-item>
            <el-form-item label="模型描述" class="model-desc-form-item">
              <div class="model-desc-row" v-if="!editingDesc">
                <div class="model-desc-preview">
                  <div class="model-desc-content" v-if="editingModelData.description">
                    <MarkdownRenderer :content="editingModelData.description" />
                  </div>
                  <div class="model-desc-empty" v-else>
                    暂无描述
                  </div>
                </div>
                <el-tooltip content="编辑描述" placement="top">
                  <el-button size="small" circle class="model-desc-edit-btn" @click="editingDesc = true">
                    <el-icon><Edit /></el-icon>
                  </el-button>
                </el-tooltip>
              </div>
              <div class="model-desc-editing" v-else>
                <div class="model-desc-row">
                  <div class="model-desc-edit-box">
                    <el-input
                      v-model="editingModelData.description"
                      type="textarea"
                      :autosize="{ minRows: 1, maxRows: 8 }"
                      placeholder="使用 Markdown 编写模型简介..."
                    />
                    <div class="model-desc-editing-actions">
                      <el-button size="small" @click="editingDesc = false">取消</el-button>
                      <el-button size="small" type="primary" @click="saveDescription">完成</el-button>
                    </div>
                  </div>
                </div>
              </div>
            </el-form-item>
            <el-form-item label="最大输出">
              <el-input-number
                v-model="editingModelData.max_tokens"
                :min="100"
                :max="128000"
                :step="100"
                style="width: 100%"
              />
            </el-form-item>
            <el-form-item label="温度">
              <div class="slider-with-value">
                <el-slider v-model="editingModelData.temperature" :min="0" :max="2" :step="0.1" style="flex: 1" />
                <span class="slider-value">{{ editingModelData.temperature }}</span>
              </div>
            </el-form-item>
            <el-form-item label="推理能力">
              <div class="slider-with-value">
                <el-slider v-model="editingModelData.top_p" :min="0" :max="1" :step="0.05" style="flex: 1" />
                <span class="slider-value">{{ editingModelData.top_p }}</span>
              </div>
            </el-form-item>
            <el-form-item label="惩罚力度">
              <div class="slider-with-value">
                <el-slider v-model="editingModelData.frequency_penalty" :min="-2" :max="2" :step="0.1" style="flex: 1" />
                <span class="slider-value">{{ editingModelData.frequency_penalty }}</span>
              </div>
            </el-form-item>
          </el-form>
        </div>
        <div class="model-config-empty" v-else>
          <el-empty description="请从左侧选择一个模型" :image-size="80" />
        </div>
      </div>
      <template #footer>
        <el-button @click="modelConfigDialogVisible = false">取消</el-button>
        <el-button type="success" @click="useSelectedModel" :disabled="!editingModelData">使用</el-button>
        <el-button type="primary" @click="saveModelConfig" :disabled="!editingModelData">保存配置</el-button>
      </template>
    </el-dialog>

    <RelateContentDialog
      v-model:visible="showRelateContentDialog"
      :chapters="chaptersList"
      :memos="memos"
      :initial-selected-chapter-ids="relatedContent.filter(rc => rc.type === 'chapter').map(rc => Number(rc.id))"
      :initial-selected-memo-ids="relatedContent.filter(rc => rc.type === 'memo').map(rc => Number(rc.id))"
      @confirm="handleRelateContentConfirm"
    />

    <!-- 完整提示词预览 -->
    <FullPromptPreviewDialog
      v-model:visible="showFullPromptDialog"
      :content="fullPromptContent"
    />

    <WriteRegexDialog
      v-model:visible="writeRegexDialogVisible"
      v-model:write-display-regex-enabled="writeDisplayRegexEnabled"
      v-model:write-stream-guard-enabled="writeStreamGuardEnabled"
      v-model:write-copy-uses-filtered="writeCopyUsesFiltered"
      v-model:write-regex-rules="writeRegexRules"
      :compute-display-content="computeWriteAssistantDisplayContent"
      :create-default-rules="createDefaultWriteRegexRules"
    />

    <FullscreenEditorDialog
      v-model:visible="fullscreenEditorVisible"
      v-model="fullscreenEditorContent"
      title="编辑 - 写作剧情"
      @save="saveFullscreenEditor"
    />

    <MessageEditDialog
      v-model:visible="messageEditDialogVisible"
      v-model="messageEditContent"
      title="编辑消息"
      @save="saveMessageEdit"
      @cancel="cancelMessageEdit"
    />

    <CreativeFullscreenDialog
      v-model:visible="creativeFullscreenDialogVisible"
      v-model:content="creativeFullscreenContent"
      title="编辑 - 写作内容"
      @save="saveCreativeFullscreenEditor"
      @cancel="cancelCreativeFullscreenEditor"
    />

    <!-- 导入章节弹窗 -->
    <ImportChapterDialog
      v-model:visible="importChapterDialogVisible"
      :book-id="bookId"
      @imported="fetchChapters"
    />

    <!-- 角色库全屏弹窗 -->
    <Teleport to="body">
      <Transition name="modal-scale">
        <div
          v-if="showCharacterLibrary"
          class="character-library-overlay"
          @click.self="closeCharacterLibrary"
        >
          <!-- 右上角关闭按钮 -->
          <button class="modal-close-btn" @click="closeCharacterLibrary">
            <el-icon><Close /></el-icon>
          </button>

          <!-- 弹窗容器 -->
          <div class="modal-container">
            <CharacterLibraryModal
              :book-id="bookId"
              :is-modal="true"
              @close="closeCharacterLibrary"
              @select-character="handleCharacterSelected"
            />
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, onMounted, nextTick, computed, watch, defineAsyncComponent } from 'vue'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import SplitRichTextEditor from '@/components/SplitRichTextEditor.vue'
import WriteCatalogTree from '@/components/WriteCatalogTree.vue'
import VolumeDialog from '@/components/WriteDialogs/VolumeDialog.vue'
import MoveChapterDialog from '@/components/WriteDialogs/MoveChapterDialog.vue'
import HistoryDialog from '@/components/WriteDialogs/HistoryDialog.vue'
import ImportChapterDialog from '@/components/WriteDialogs/ImportChapterDialog.vue'
import RenameDialog from '@/components/WriteDialogs/RenameDialog.vue'
import FullscreenEditorDialog from '@/components/WriteDialogs/FullscreenEditorDialog.vue'
import ConversationListDialog from '@/components/WriteDialogs/ConversationListDialog.vue'
import CreativeFullscreenDialog from '@/components/WriteDialogs/CreativeFullscreenDialog.vue'
import MessageEditDialog from '@/components/WriteDialogs/MessageEditDialog.vue'
import FullPromptPreviewDialog from '@/components/WriteDialogs/FullPromptPreviewDialog.vue'
import PromptSelectDialog from '@/components/WriteDialogs/PromptSelectDialog.vue'
import RelateContentDialog from '@/components/WriteDialogs/RelateContentDialog.vue'
import ModelSelectDialog from '@/components/WriteDialogs/ModelSelectDialog.vue'
import WriteRegexDialog from '@/components/WriteDialogs/WriteRegexDialog.vue'
import FontSelector from '@/components/FontSelector.vue'
import FontSizeSelector from '@/components/FontSizeSelector.vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBookStore } from '@/stores/book'
import { useWorldBookStore } from '@/stores/worldbook'
import { onUnmounted } from 'vue'
import { chapterAPI, memoAPI, promptAPI, configAPI, conversationAPI, volumeAPI, characterAPI } from '@/api'
import type { Chapter, Memo, Prompt, ApiConfig, ChatMessage, RelatedContent, Volume, Character, KnowledgeGraphData } from '@/types'
import type { WorldBookEntry } from '@/types/worldbook'
import { createDefaultWorldBookEntry } from '@/types/worldbook'
import { matchWorldBookEntries, injectWorldBookContent } from '@/utils/worldbook'
import KnowledgeGraph from '@/components/KnowledgeGraph.vue'
import ChatPanel from '@/components/ChatPanel/index.vue'
import { 
  ChatDotSquare, User, Notebook, ChatLineSquare, 
  Plus, MoreFilled, Delete, Folder, Switch, Sort, 
  ArrowRight, ArrowLeft, ArrowDown, Edit, EditPen, RefreshLeft, RefreshRight,
  Star, Grid, Refresh, Search, CopyDocument, Checked, Loading, View,
  Monitor, MagicStick, DocumentCopy, Document, Close, Promotion, ChatDotRound, Link, Lightning, Setting, Download, Connection, VideoPause, Position, Reading, Upload, UploadFilled, FullScreen,
  Collection, InfoFilled
} from '@element-plus/icons-vue'

// 延迟加载角色库组件
const CharacterLibraryModal = defineAsyncComponent(() => import('@/views/CharacterLibrary.vue'))

const route = useRoute()
const router = useRouter()
const bookStore = useBookStore()

const bookId = parseInt(route.params.bookId as string) || 0
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

// 文件夹相关
const volumes = ref<Volume[]>([])
const expandedFolderIds = ref<Array<number | string>>([])
const isDescending = ref(false)
const draggingChapterId = ref<number | null>(null)
const dragOverTarget = ref<string | null>(null)
const sortedFolders = computed(() => {
  return [...volumes.value].sort((a, b) => {
    if (isDescending.value) {
      return (b.order_num - a.order_num) || (b.id - a.id)
    }
    return (a.order_num - b.order_num) || (a.id - b.id)
  })
})
const sortedChapters = computed(() => {
  let chapters = [...chaptersList.value]
  if (isDescending.value) {
    chapters.sort((a, b) => b.id - a.id)
  } else {
    chapters.sort((a, b) => a.id - b.id)
  }
  return chapters
})
const rootFolders = computed(() => sortedFolders.value.filter(folder => !folder.parent_id || folder.parent_id === 0))
const rootFiles = computed(() => sortedChapters.value.filter(chapter => !chapter.volume_id || chapter.volume_id === 0))

// API配置
const apiConfigs = ref<ApiConfig[]>([])
const selectedConfigId = ref<number>()

// 模型配置弹窗
const modelConfigDialogVisible = ref(false)
const editingModelId = ref<number | null>(null)
const editingModelData = ref<any>(null)
const editingDesc = ref(false)

// 当前选中的模型名称
const currentModelName = computed(() => {
  const config = apiConfigs.value.find(c => c.id === selectedConfigId.value)
  return config?.name || ''
})

// 模型选择弹窗
const modelSelectDialogVisible = ref(false)
const modelSelectMode = ref<'chat' | 'creative2'>('chat')
const modelTemperature = ref(0.7)
const modelTopP = ref(0.9)

const handleModelSelectConfirm = (payload: { configId: number; temperature: number; topP: number; mode: 'chat' | 'creative2' }) => {
  if (payload.mode === 'creative2') {
    creative2ConfigId.value = payload.configId
  } else {
    selectedConfigId.value = payload.configId
  }
}

const openModelSelectDialog = () => {
  modelSelectMode.value = 'chat'
  // 如果当前没有选中任何模型，自动选中默认模型或第一个可用模型
  if (!selectedConfigId.value) {
    const defaultConfig = apiConfigs.value.find(c => c.is_default)
    const firstConfig = apiConfigs.value[0]
    selectedConfigId.value = defaultConfig?.id ?? firstConfig?.id
  }
  modelSelectDialogVisible.value = true
}

const openModelConfigDialog = () => {
  editingModelId.value = null
  editingModelData.value = null
  editingDesc.value = false
  modelConfigDialogVisible.value = true
}

const selectModelForEdit = (config: ApiConfig) => {
  editingModelId.value = config.id
  editingDesc.value = false
  editingModelData.value = {
    id: config.id,
    provider_id: config.provider_id,
    name: config.name,
    model: config.model,
    temperature: config.temperature ?? 0.7,
    max_tokens: config.max_tokens ?? 2000,
    top_p: config.top_p ?? 0.9,
    frequency_penalty: config.frequency_penalty ?? 0.0,
    description: config.description ?? '',
    is_default: config.is_default
  }
}

const saveDescription = async () => {
  if (!editingModelData.value) return
  try {
    const d = editingModelData.value
    await configAPI.update(d.id, {
      ...d,
      is_default: d.is_default ? 1 : 0
    })
    editingDesc.value = false
    ElMessage.success('描述已保存')
    await fetchConfigs()
  } catch {
    // 错误已由拦截器显示
  }
}

const saveModelConfig = async () => {
  if (!editingModelData.value) return
  try {
    const res = await configAPI.update(editingModelData.value.id, {
      ...editingModelData.value,
      is_default: editingModelData.value.is_default ? 1 : 0
    })
    if (res.success) {
      ElMessage.success('模型配置已保存')
      if (selectedConfigId.value === editingModelData.value.id) {
        selectedConfigId.value = editingModelData.value.id
      }
      modelConfigDialogVisible.value = false
      await fetchConfigs()
    }
  } catch {
    // 错误已由拦截器显示
  }
}

const useSelectedModel = () => {
  if (!editingModelData.value) return
  selectedConfigId.value = editingModelData.value.id
  modelConfigDialogVisible.value = false
  ElMessage.success(`已切换到模型: ${editingModelData.value.name}`)
}

// 提示词
const prompts = ref<Prompt[]>([])
const selectedPrompts = ref<number[]>([])

// @ 引用功能
interface AttachedReference {
  id: number | string
  type: string
  label: string
  tagType: '' | 'success' | 'warning' | 'danger' | 'info'
}
const showAtMenu = ref(false)
const attachedReferences = ref<AttachedReference[]>([])
const atMenuOptions = [
  { type: 'chapter', label: '关联章节/备忘录', hint: '@chapter:', icon: 'Notebook', color: '#409eff' },
  { type: 'prompt', label: '提示词', hint: '@prompt:', icon: 'Document', color: '#e6a23c' },
  { type: 'worldbook', label: '世界书', hint: '', icon: 'Reading', color: '#67c23a' },
  { type: 'character', label: '角色', hint: '@character:', icon: 'User', color: '#f56c6c' },
  { type: 'regex', label: '正则过滤', hint: '', icon: 'Setting', color: '#909399' },
]

const handleAtSelect = (type: string) => {
  showAtMenu.value = false
  switch (type) {
    case 'chapter':
      openRelateContentDialog()
      break
    case 'prompt':
      promptSelectDialogVisible.value = true
      break
    case 'worldbook':
      openWorldBookDialog()
      break
    case 'character':
      // 角色功能可后续扩展
      ElMessage.info('角色引用功能开发中')
      break
    case 'regex':
      writeRegexDialogVisible.value = true
      break
  }
}

// 同步 attachedReferences 与 selectedPrompts / relatedContent
const syncAttachedReferences = () => {
  const list: AttachedReference[] = []
  selectedPrompts.value.forEach(id => {
    const p = prompts.value.find((item: any) => item.id === id)
    if (p) list.push({ id: p.id, type: 'prompt', label: p.name || `提示词${id}`, tagType: 'warning' })
  })
  relatedContent.value.forEach(rc => {
    if (rc.type === 'chapter') list.push({ id: rc.id, type: 'chapter', label: rc.title, tagType: '' })
    else if (rc.type === 'memo') list.push({ id: rc.id, type: 'memo', label: rc.title, tagType: 'info' })
    else if (rc.type === 'worldbook') list.push({ id: rc.id, type: 'worldbook', label: rc.title || '世界书条目', tagType: 'success' })
  })
  attachedReferences.value = list
}

const removeAttachedReference = (ref: AttachedReference) => {
  if (ref.type === 'prompt') {
    const idx = selectedPrompts.value.indexOf(ref.id as number)
    if (idx > -1) selectedPrompts.value.splice(idx, 1)
  } else {
    relatedContent.value = relatedContent.value.filter(rc => !(rc.id === ref.id && rc.type === ref.type))
  }
  syncAttachedReferences()
}

// 提示词选择弹窗
const promptSelectDialogVisible = ref(false)

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

const toggleMemoFolder = (folder: string) => {
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

// 历史记录相关 - 类型定义（写入逻辑仍在 Write.vue，UI 已迁移到 HistoryDialog 组件）
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
const expandedConversationId = ref<number | null>(null)

// AI 写作 2 - 抽卡区专用
const creative2ConfigId = ref<number>()
const creative2ModelName = computed(() => {
  const config = apiConfigs.value.find(c => c.id === creative2ConfigId.value)
  return config?.name || ''
})

const openCreative2ModelSelectDialog = () => {
  modelSelectMode.value = 'creative2'
  // 如果当前没有选中任何模型，自动选中默认模型或第一个可用模型
  if (!creative2ConfigId.value) {
    const defaultConfig = apiConfigs.value.find(c => c.is_default)
    const firstConfig = apiConfigs.value[0]
    creative2ConfigId.value = defaultConfig?.id ?? firstConfig?.id
  }
  modelSelectDialogVisible.value = true
}
const creative2SelectedPrompts = ref<number[]>([])
const creative2Generating = ref(false)
const creative2Result = ref('')
const creative2ResultDialogVisible = ref(false)
const creative2AbortController = ref<AbortController | null>(null)
const creative2WaitingForResponse = ref(false)
const streamingScrollbarRef = ref<any>(null)
const creative2AdvancedMode = ref(false)
const creative2PromptDialogVisible = ref(false)
const creative2SecondPromptDialogVisible = ref(false)
const creative2SecondPromptId = ref<number>(0)
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

// 抽卡区字段配置
const creative2FieldValues = ref<Record<string, string>>({})

const creative2AllFields = computed(() => {
  const fields: Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; optionLabels?: string[]; description: string; required: boolean }> = []
  
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
            optionLabels: field.optionLabels || [],
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

type RegexRule = {
  id: string
  name: string
  pattern: string
  flags: string
  replacement: string
  enabled: boolean
  affectsActual: boolean
  order: number
  lastError?: string
}

const WRITE_REGEX_SETTINGS_STORAGE_KEY = 'write-chat-regex-settings-v1'

const createRegexRuleId = () => {
  const g = globalThis as any
  if (g.crypto?.randomUUID) return g.crypto.randomUUID()
  return `write_regex_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 10)}`
}

const createDefaultWriteRegexRules = (): RegexRule[] => {
  const make = (partial: Omit<RegexRule, 'id' | 'order'>, order: number): RegexRule => ({
    id: createRegexRuleId(),
    order,
    ...partial
  })

  return [
    make({
      name: '移除 <think>…</think>',
      pattern: '<think>[\\s\\S]*?<\\/think>',
      flags: 'g',
      replacement: '',
      enabled: true,
      affectsActual: false
    }, 10),
    make({
      name: '移除 <analysis>…</analysis>',
      pattern: '<analysis>[\\s\\S]*?<\\/analysis>',
      flags: 'g',
      replacement: '',
      enabled: false,
      affectsActual: false
    }, 20)
  ]
}

const normalizeWriteRegexRulesOrder = () => {
  writeRegexRules.value.forEach((rule, index) => {
    rule.order = index + 1
  })
}

const writeRegexDialogVisible = ref(false)
const writeDisplayRegexEnabled = ref(true)
const writeStreamGuardEnabled = ref(true)
const writeCopyUsesFiltered = ref(false)
const writeRegexRules = ref<RegexRule[]>([])

const getActiveWriteRegexRules = (onlyAffectsActual = false) => {
  return writeRegexRules.value
    .filter(rule => rule.enabled && rule.pattern.trim() && (onlyAffectsActual ? rule.affectsActual : true))
    .slice()
    .sort((a, b) => a.order - b.order)
}

const applyWriteStreamingGuard = (text: string) => {
  if (!writeStreamGuardEnabled.value) return text

  const guards: Array<{ open: string; close: string }> = [
    { open: '<think>', close: '</think>' },
    { open: '<analysis>', close: '</analysis>' }
  ]

  const cutPositions: number[] = []
  for (const guard of guards) {
    const openIndex = text.lastIndexOf(guard.open)
    if (openIndex === -1) continue
    const closeIndex = text.indexOf(guard.close, openIndex + guard.open.length)
    if (closeIndex === -1) {
      cutPositions.push(openIndex)
    }
  }

  if (cutPositions.length === 0) return text
  return text.slice(0, Math.min(...cutPositions))
}

const applyWriteRegexPipeline = (text: string, onlyAffectsActual = false) => {
  if (!writeDisplayRegexEnabled.value && !onlyAffectsActual) return text

  let output = text
  const activeRules = getActiveWriteRegexRules(onlyAffectsActual)

  for (const rule of activeRules) {
    rule.lastError = undefined
    if (rule.pattern.length > 2000 || rule.replacement.length > 5000) {
      rule.lastError = '规则过长，已跳过'
      continue
    }

    try {
      const flags = rule.flags?.trim() || 'g'
      const regex = new RegExp(rule.pattern, flags)
      output = output.replace(regex, rule.replacement)
    } catch (error: any) {
      rule.lastError = error?.message || '无效正则'
    }
  }

  return output
}

const computeWriteAssistantDisplayContent = (rawContent: string) => {
  const guardedContent = applyWriteStreamingGuard(rawContent || '')
  return applyWriteRegexPipeline(guardedContent)
}

const getAssistantDisplayContent = (message: ChatMessage) => {
  return message.displayContent || computeWriteAssistantDisplayContent(message.content || '')
}

const getMessageCopyContent = (message: ChatMessage) => {
  if (message.role === 'assistant' && writeCopyUsesFiltered.value) {
    return getAssistantDisplayContent(message)
  }
  // user 消息优先用 displayContent（干净的），避免暴露注入的提示词/关联内容原文
  return message.displayContent || extractUserDisplayContent(message.content) || message.content
}

const refreshWriteChatDisplayContents = () => {
  chatMessages.value = chatMessages.value.map(message => {
    if (message.role === 'assistant') {
      return {
        ...message,
        displayContent: computeWriteAssistantDisplayContent(message.content)
      }
    }
    if (message.role === 'user') {
      return {
        ...message,
        displayContent: message.displayContent || extractUserDisplayContent(message.content)
      }
    }
    return message
  })
}

const loadWriteRegexSettings = () => {
  try {
    const raw = localStorage.getItem(WRITE_REGEX_SETTINGS_STORAGE_KEY)
    if (!raw) {
      writeRegexRules.value = createDefaultWriteRegexRules()
      normalizeWriteRegexRulesOrder()
      return
    }

    const parsed = JSON.parse(raw)
    writeDisplayRegexEnabled.value = parsed?.writeDisplayRegexEnabled !== false
    writeStreamGuardEnabled.value = parsed?.writeStreamGuardEnabled !== false
    writeCopyUsesFiltered.value = parsed?.writeCopyUsesFiltered === true

    if (Array.isArray(parsed?.writeRegexRules)) {
      writeRegexRules.value = parsed.writeRegexRules
        .map((item: any, index: number) => ({
          id: typeof item.id === 'string' && item.id ? item.id : createRegexRuleId(),
          name: typeof item.name === 'string' ? item.name : `规则 ${index + 1}`,
          pattern: typeof item.pattern === 'string' ? item.pattern : '',
          flags: typeof item.flags === 'string' ? item.flags : 'g',
          replacement: typeof item.replacement === 'string' ? item.replacement : '',
          enabled: typeof item.enabled === 'boolean' ? item.enabled : true,
          order: typeof item.order === 'number' ? item.order : index + 1
        }))
        .sort((a: RegexRule, b: RegexRule) => a.order - b.order)
      normalizeWriteRegexRulesOrder()
    } else {
      writeRegexRules.value = createDefaultWriteRegexRules()
      normalizeWriteRegexRulesOrder()
    }
  } catch {
    writeRegexRules.value = createDefaultWriteRegexRules()
    normalizeWriteRegexRulesOrder()
  }
}

const saveWriteRegexSettings = () => {
  const payload = {
    writeDisplayRegexEnabled: writeDisplayRegexEnabled.value,
    writeStreamGuardEnabled: writeStreamGuardEnabled.value,
    writeCopyUsesFiltered: writeCopyUsesFiltered.value,
    writeRegexRules: writeRegexRules.value.map(rule => ({
      id: rule.id,
      name: rule.name,
      pattern: rule.pattern,
      flags: rule.flags,
      replacement: rule.replacement,
      enabled: rule.enabled,
      order: rule.order
    }))
  }

  localStorage.setItem(WRITE_REGEX_SETTINGS_STORAGE_KEY, JSON.stringify(payload))
}

const chatMessages = ref<ChatMessage[]>([])
const userInput = ref('')
const sending = ref(false)
const collapsedUserMessages = ref(new Set<number>())
const chatAbortController = ref<AbortController | null>(null)
const showFullPromptDialog = ref(false)
const worldBookLinked = ref(false) // 世界书与对话是否关联
const fullscreenEditorVisible = ref(false)
const fullscreenEditorContent = ref('')
const messageEditDialogVisible = ref(false)
const messageEditContent = ref('')
const messageEditIndex = ref<number | null>(null)
const creativeFullscreenDialogVisible = ref(false)
const creativeFullscreenContent = ref('')
const creativeFullscreenField = ref('')

// 完整提示词预览
const fullPromptContent = computed(() => {
  const selectedPromptRecords = prompts.value.filter((p: any) => selectedPrompts.value.some(id => id == p.id))
  const promptContents = selectedPromptRecords.map((p: any) => p.content).join('\n\n')
  const relatedContentSummary = relatedContent.value.length > 0
    ? `关联内容：\n${relatedContent.value
        .map(item => `${item.type === 'chapter' ? '章节' : '备忘录'}：${item.title}\n${item.content}`)
        .join('\n\n')}`
    : ''

  const parts: string[] = []
  if (promptContents) parts.push(`【提示词】\n${promptContents}`)
  
  // 世界书注入预览
  if (worldBookLinked.value) {
    const bookEntries = currentBookWorldBook.value?.entries || []
    if (bookEntries.length > 0) {
      const matchResults = matchWorldBookEntries(userInput.value || '', bookEntries)
      if (matchResults.length > 0) {
        const injection = injectWorldBookContent('', matchResults).trim()
        if (injection) parts.push(injection)
      }
    }
  }
  
  if (relatedContentSummary) parts.push(relatedContentSummary)
  if (userInput.value.trim()) parts.push(`【用户输入】\n${userInput.value.trim()}`)

  if (parts.length === 0) return '（暂无内容，请选择提示词、关联内容或输入消息）'

  // 显示完整的消息历史
  const historyLines = chatMessages.value.map((msg, i) => {
    const role = msg.role === 'user' ? '用户' : msg.role === 'assistant' ? 'AI' : '系统'
    return `[${role}] ${msg.content}`
  }).join('\n\n---\n\n')

  return `=== 历史消息 ===\n${historyLines}\n\n=== 即将发送 ===\n${parts.join('\n\n')}`
})



const chatMessagesRef = ref<HTMLElement>()
const editingMessageIndex = ref<number | null>(null)
const editingContent = ref('')
const selectedTextLength = ref(0)
const fontSize = ref(16)
const fontFamily = ref('Microsoft YaHei')

// 面板宽度调节
const leftPanelWidth = ref(220)
const centerPanelWidth = ref(0) // 动态计算
const rightPanelWidth = ref(450)
const rightPanel2Width = ref(420)
const isResizing = ref(false)
const resizeSide = ref<'left' | 'right' | 'right2'>('left')

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
const conversationListVisible = ref(false)
const relateType = ref('chapters')
const selectedChapters = ref<number[]>([])
const selectedMemos = ref<number[]>([])
const relatedContent = ref<RelatedContent[]>([])

// 关联内容管理弹窗
const showRelateContentDialog = ref(false)

const handleRelateContentConfirm = ({ selectedChapterIds, selectedMemoIds }: { selectedChapterIds: number[]; selectedMemoIds: number[] }) => {
  // 清除旧的章节/备忘录关联，保留其他类型的关联（如世界书）
  relatedContent.value = relatedContent.value.filter(
    rc => rc.type !== 'chapter' && rc.type !== 'memo'
  )

  const stripHtml = (html: string) => html.replace(/<[^>]*>/g, '').replace(/&nbsp;/g, ' ').replace(/&amp;/g, '&').replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&quot;/g, '"').trim()

  // 添加选中的章节
  selectedChapterIds.forEach(id => {
    const chapter = chaptersList.value.find(c => c.id === id)
    if (chapter) {
      relatedContent.value.push({
        type: 'chapter',
        id: chapter.id,
        title: chapter.title || '无标题',
        content: stripHtml(chapter.content || '')
      })
    }
  })

  // 添加选中的备忘录
  selectedMemoIds.forEach(id => {
    const memo = memos.value.find(m => m.id === id)
    if (memo) {
      relatedContent.value.push({
        type: 'memo',
        id: memo.id,
        title: memo.title || '无标题',
        content: stripHtml(memo.content || '')
      })
    }
  })

  ElMessage.success(`已关联 ${selectedChapterIds.length + selectedMemoIds.length} 项内容`)
}

const openRelateContentDialog = () => {
  showRelateContentDialog.value = true
}

// 世界书关联
const worldBookStore = useWorldBookStore()
const worldbookSearchKeyword = ref('')
const selectedWorldBookEntries = ref<string[]>([])
const worldBookEditorVisible = ref(false)
const editingWorldBookEntry = ref<WorldBookEntry | null>(null)
const worldBookFileInput = ref<HTMLInputElement>()
const importAsNewFileInput = ref<HTMLInputElement>()

// 世界书选择器
const currentWorldBookName = computed({
  get: () => worldBookStore.getCurrentWorldBookName(String(bookId)),
  set: (name: string) => {
    worldBookStore.setCurrentWorldBookName(String(bookId), name)
  }
})
const worldBookList = computed(() => worldBookStore.getBookWorldBookList(String(bookId)))
const showWorldBookDialog = ref(false)
const worldBookDialogMode = ref<'create' | 'rename'>('create')
const worldBookDialogName = ref('')

const openCreateWorldBookDialog = () => {
  worldBookDialogMode.value = 'create'
  worldBookDialogName.value = ''
  showWorldBookDialog.value = true
}

const openRenameWorldBookDialog = () => {
  worldBookDialogMode.value = 'rename'
  worldBookDialogName.value = currentWorldBookName.value
  showWorldBookDialog.value = true
}

const confirmWorldBookDialog = () => {
  const name = worldBookDialogName.value.trim()
  if (!name) return ElMessage.warning('名称不能为空')
  if (worldBookDialogMode.value === 'create') {
    worldBookStore.createWorldBook(String(bookId), name)
    ElMessage.success('世界书已创建')
  } else {
    if (name === currentWorldBookName.value) {
      showWorldBookDialog.value = false
      return
    }
    if (!worldBookStore.renameWorldBook(String(bookId), currentWorldBookName.value, name)) {
      return ElMessage.warning('名称已存在或操作失败')
    }
    ElMessage.success('已重命名')
  }
  showWorldBookDialog.value = false
}

const handleDeleteWorldBook = async () => {
  const name = currentWorldBookName.value
  try {
    await ElMessageBox.confirm(`确定删除世界书「${name}」及其中所有条目吗？`, '删除确认', { type: 'warning' })
    worldBookStore.deleteWorldBook(String(bookId), name)
    ElMessage.success('已删除')
  } catch {}
}

const worldBookForm = ref(createDefaultWorldBookEntry())

// 词条状态：normal/constant/disable 三选一
const worldBookEntryStatus = computed({
  get: () => worldBookForm.value.disable ? 'disable' : worldBookForm.value.constant ? 'constant' : 'normal',
  set: (val: string) => {
    worldBookForm.value.constant = val === 'constant'
    worldBookForm.value.disable = val === 'disable'
  }
})

const currentBookWorldBook = computed(() => {
  return worldBookStore.getBookWorldBook(String(bookId), currentWorldBookName.value)
})

const filteredWorldBookEntries = computed(() => {
  const entries = currentBookWorldBook.value?.entries || []
  
  if (!worldbookSearchKeyword.value.trim()) {
    return entries
  }

  const keyword = worldbookSearchKeyword.value.toLowerCase()
  return entries.filter(entry => {
    return (
      entry.comment.toLowerCase().includes(keyword) ||
      entry.key.some(k => k.toLowerCase().includes(keyword)) ||
      entry.content.toLowerCase().includes(keyword)
    )
  })
})

// 分组
const collapsedGroups = ref<Set<string>>(new Set())

const worldBookGroups = computed(() => {
  const entries = filteredWorldBookEntries.value
  const groupMap = new Map<string, WorldBookEntry[]>()
  
  entries.forEach(entry => {
    const g = entry.group?.trim() || ''
    if (!groupMap.has(g)) groupMap.set(g, [])
    groupMap.get(g)!.push(entry)
  })
  
  // 有名称的分组在前 → 默认（空分组名）在最后
  return Array.from(groupMap.entries()).sort((a, b) => {
    if (!a[0] && !b[0]) return 0
    if (!a[0]) return 1
    if (!b[0]) return -1
    return a[0].localeCompare(b[0])
  })
})

// 编辑器分组下拉可选值
const availableGroups = computed(() => {
  const names = new Set<string>()
  currentBookWorldBook.value?.entries.forEach(e => {
    const g = e.group?.trim()
    if (g) names.add(g)
  })
  return Array.from(names).sort((a, b) => a.localeCompare(b))
})

const toggleGroupCollapse = (group: string) => {
  if (collapsedGroups.value.has(group)) {
    collapsedGroups.value.delete(group)
  } else {
    collapsedGroups.value.add(group)
  }
}

// 分组管理
const showGroupDialog = ref(false)
const groupDialogMode = ref<'create' | 'rename'>('create')
const groupDialogName = ref('')
const groupDialogOldName = ref('')

const openCreateGroupDialog = () => {
  groupDialogMode.value = 'create'
  groupDialogName.value = ''
  groupDialogOldName.value = ''
  showGroupDialog.value = true
}

const openRenameGroupDialog = (oldName: string) => {
  groupDialogMode.value = 'rename'
  groupDialogName.value = oldName
  groupDialogOldName.value = oldName
  showGroupDialog.value = true
}

const confirmGroupDialog = async () => {
  const name = groupDialogName.value.trim()
  if (!name) {
    ElMessage.warning('请输入分组名称')
    return
  }
  if (groupDialogMode.value === 'rename' && name === groupDialogOldName.value) {
    showGroupDialog.value = false
    return
  }
  
  if (groupDialogMode.value === 'create') {
    showGroupDialog.value = false
    ElMessage.success(`分组"${name}"已创建，新建条目时可选择该分组`)
  } else {
    const oldName = groupDialogOldName.value
    const entries = currentBookWorldBook.value?.entries.filter(e => (e.group?.trim() || '') === oldName) || []
    for (const entry of entries) {
      await worldBookStore.updateEntry(entry.uid, { group: name }, String(bookId), currentWorldBookName.value)
    }
    showGroupDialog.value = false
    ElMessage.success(`分组已重命名为"${name}"`)
  }
}

const deleteGroup = async (groupName: string) => {
  try {
    const entries = currentBookWorldBook.value?.entries.filter(e => (e.group?.trim() || '') === groupName) || []
    await ElMessageBox.confirm(
      `确定删除分组"${groupName}"吗？该分组下的 ${entries.length} 个条目将移至默认分组。`,
      '删除分组',
      { type: 'warning' }
    )
    for (const entry of entries) {
      await worldBookStore.updateEntry(entry.uid, { group: '' }, String(bookId))
    }
    ElMessage.success('分组已删除')
  } catch {
    // 取消
  }
}

const worldBookEntryCount = computed(() => {
  return currentBookWorldBook.value?.entries.length || 0
})

const enabledWorldBookCount = computed(() => {
  const entries = currentBookWorldBook.value?.entries || []
  return entries.filter(e => !e.disable).length
})

const openWorldBookDialog = () => {
  selectedWorldBookEntries.value = []
  worldbookSearchKeyword.value = ''
  // 所有分组默认折叠
  collapsedGroups.value = new Set(worldBookGroups.value.map(([name]) => name))
  showRelateDialog.value = true
}

const toggleWorldBookEntry = async (entry: WorldBookEntry, enabled: boolean) => {
  entry.disable = !enabled
  await worldBookStore.updateEntry(entry.uid, { disable: !enabled }, String(bookId))
}

const createWorldBookEntry = () => {
  editingWorldBookEntry.value = null
  worldBookForm.value = createDefaultWorldBookEntry()
  worldBookEditorVisible.value = true
}

const editWorldBookEntry = (entry: WorldBookEntry) => {
  editingWorldBookEntry.value = entry
  worldBookForm.value = { ...entry }
  worldBookEditorVisible.value = true
}

const saveWorldBookEntry = async () => {
  if (!worldBookForm.value.comment.trim()) {
    ElMessage.warning('请输入条目标题')
    return
  }

  if (editingWorldBookEntry.value) {
    await worldBookStore.updateEntry(editingWorldBookEntry.value.uid, worldBookForm.value, String(bookId))
    ElMessage.success('条目已更新')
  } else {
    await worldBookStore.addEntry(worldBookForm.value, String(bookId))
    ElMessage.success('条目已创建')
  }

  worldBookEditorVisible.value = false
}

const deleteWorldBookEntry = async (entry: WorldBookEntry) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除条目"${entry.comment || '未命名条目'}"吗？`,
      '删除确认',
      { type: 'warning' }
    )
    await worldBookStore.deleteEntry(entry.uid, String(bookId))
    ElMessage.success('条目已删除')
  } catch {
    // 取消
  }
}

const handleImportCommand = (command: string) => {
  if (command === 'current') {
    worldBookFileInput.value?.click()
  } else if (command === 'new') {
    importAsNewFileInput.value?.click()
  }
}

const handleWorldBookFileSelect = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  try {
    const text = await file.text()
    const data = JSON.parse(text)
    
    const success = worldBookStore.importWorldBookData(data, String(bookId))
    if (success) {
      ElMessage.success('世界书导入成功')
    } else {
      ElMessage.error('世界书格式不正确')
    }
  } catch (error) {
    ElMessage.error('导入失败：文件格式错误')
  }

  if (worldBookFileInput.value) {
    worldBookFileInput.value.value = ''
  }
}

const handleWorldBookImportAsNew = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  try {
    const text = await file.text()
    const data = JSON.parse(text)

    const fileName = file.name.replace(/\.(json|wb)$/i, '')

    const existingList = worldBookStore.getBookWorldBookList(String(bookId))
    if (existingList.includes(fileName)) {
      try {
        await ElMessageBox.confirm(
          `世界书「${fileName}」已存在，是否合并导入？`,
          '名称冲突',
          { type: 'warning' }
        )
      } catch {
        if (importAsNewFileInput.value) {
          importAsNewFileInput.value.value = ''
        }
        return
      }
    }

    data.name = fileName

    const success = worldBookStore.importWorldBookData(data, String(bookId))
    if (success) {
      ElMessage.success(`已导入为新世界书「${fileName}」`)
    } else {
      ElMessage.error('世界书格式不正确')
    }
  } catch (error) {
    ElMessage.error('导入失败：文件格式错误')
  }

  if (importAsNewFileInput.value) {
    importAsNewFileInput.value.value = ''
  }
}

const exportWorldBook = (format: 'default' | 'sillytavern' = 'default') => {
  const data = worldBookStore.exportWorldBookData(String(bookId), format)
  if (!data || !data.entries || (Array.isArray(data.entries) ? data.entries.length === 0 : Object.keys(data.entries).length === 0)) {
    ElMessage.warning('没有可导出的条目')
    return
  }

  const formatLabel = format === 'sillytavern' ? '酒馆' : '默认'
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `worldbook-${currentBook.value?.title || 'book'}-${formatLabel}-${Date.now()}.json`
  a.click()
  URL.revokeObjectURL(url)
  
  ElMessage.success(`世界书已导出（${formatLabel}格式）`)
}

const handleExportFormat = (format: string) => {
  exportWorldBook(format as 'default' | 'sillytavern')
}

// 重命名
const showRenameDialog = ref(false)
const renamingConversation = ref<any>(null)

// 文件夹相关
const showVolumeDialog = ref(false)
const isEditVolume = ref(false)
const volumeForm = ref({
  id: 0,
  title: ''
})
const editingFolderParentId = ref<number | null>(null)
const moveChapterDialogVisible = ref(false)
const chapterToMove = ref<Chapter | null>(null)

onMounted(async () => {
  loadWriteRegexSettings()
  loadGlobalMemoFolders()
  await bookStore.fetchBook(bookId)
  await fetchChapters()
  await fetchMemos()
  await fetchCharacters()
  await fetchPrompts()
  await fetchConfigs()
  await fetchConversations()
  await fetchVolumes()
  
  // 恢复上次位置：优先读取 localStorage 记录，否则默认最后一章
  if (chaptersList.value.length > 0) {
    let target: Chapter | null = null
    try {
      const savedChapterId = localStorage.getItem(`novel_last_chapter_${bookId}`)
      const savedMemoId = localStorage.getItem(`novel_last_memo_${bookId}`)
      if (savedMemoId) {
        const memo = memos.value.find((m: any) => m.id === Number(savedMemoId))
        if (memo) { selectMemo(memo); target = null }
      }
      if (!target && savedChapterId) {
        target = chaptersList.value.find(c => c.id === Number(savedChapterId)) || null
      }
    } catch {}
    if (!target && !currentMemo.value) {
      target = chaptersList.value[chaptersList.value.length - 1]
    }
    if (target) await selectChapter(target)
  }
  
  // 加载字体大小设置
  loadFontSize()
  loadFontFamily()
  
  // 初始化面板宽度
  updateCenterWidth()
  
  // 监听窗口大小变化
  window.addEventListener('resize', handleResizeWindow)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResizeWindow)
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
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
    expandedFolderIds.value = Array.from(new Set([...expandedFolderIds.value, ...res.data.map(folder => folder.id)]))
  }
}

const toggleChapterOrder = () => {
  isDescending.value = !isDescending.value
}

// ===== 导入章节弹窗 =====
const importChapterDialogVisible = ref(false)

/** 点击按钮 → 打开弹窗 */
const openImportChapterDialog = () => {
  importChapterDialogVisible.value = true
}

const handleChapterDragStart = (chapterId: number) => {
  draggingChapterId.value = chapterId
}

const handleChapterDragEnd = () => {
  draggingChapterId.value = null
  dragOverTarget.value = null
}

const handleFolderDragOver = (target: string) => {
  if (!draggingChapterId.value) return
  dragOverTarget.value = target
}

const handleFolderDragLeave = (target: string) => {
  if (dragOverTarget.value === target) {
    dragOverTarget.value = null
  }
}

const handleFolderDrop = async (target: string) => {
  const chapterId = draggingChapterId.value
  draggingChapterId.value = null
  dragOverTarget.value = null

  if (!chapterId) return
  const chapter = chaptersList.value.find(item => item.id === chapterId)
  if (!chapter) return

  const nextVolumeId = target === 'root' ? undefined : Number(target)
  const currentVolumeId = chapter.volume_id || undefined
  if (currentVolumeId === nextVolumeId) return

  const res = await chapterAPI.update(chapter.id, {
    title: chapter.title,
    content: chapter.content,
    summary: chapter.summary || '',
    order_num: chapter.order_num,
    volume_id: nextVolumeId
  } as any)

  if (res.success && res.data) {
    syncChapterState(res.data)
    if (nextVolumeId && !expandedFolderIds.value.includes(nextVolumeId)) {
      expandedFolderIds.value.push(nextVolumeId)
    }
    ElMessage.success('章节已移动')
  }
}

const toggleFolder = (folderId: number | string) => {
  const index = expandedFolderIds.value.indexOf(folderId)
  if (index >= 0) {
    expandedFolderIds.value.splice(index, 1)
  } else {
    expandedFolderIds.value.push(folderId)
  }
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

// 从提示词中提取分类（已迁移到 PromptSelectDialog 组件，此处仅保留 fetchPrompts）
const fetchPrompts = async () => {
  const res = await promptAPI.getAll()
  if (res.success && res.data) {
    prompts.value = res.data
  }
}

const fetchConfigs = async () => {
  const res = await configAPI.getAll()
  if (res.success && res.data) {
    // 确保 id 和 is_default 都是数字类型
    apiConfigs.value = res.data.filter(m => m.enabled !== 0).map(m => ({
      ...m,
      id: Number(m.id),
      is_default: Number(m.is_default) || 0
    }))
    
    // 如果已经有选中的配置，检查它是否还存在
    if (selectedConfigId.value) {
      const exists = apiConfigs.value.find(c => c.id === selectedConfigId.value)
      if (!exists) {
        // 配置被删除了，使用默认配置
        const defaultConfig = apiConfigs.value.find(c => c.is_default)
        selectedConfigId.value = defaultConfig?.id
      }
    } else {
      // 首次加载，使用默认配置
      const defaultConfig = apiConfigs.value.find(c => c.is_default)
      if (defaultConfig) {
        selectedConfigId.value = defaultConfig.id
      }
    }
    
    // 设置续写功能的默认配置
    if (!continueConfigId.value && apiConfigs.value.length > 0) {
      const defaultConfig = apiConfigs.value.find(c => c.is_default)
      continueConfigId.value = defaultConfig?.id || apiConfigs.value[0].id
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
    chatMessages.value = (res.data as ChatMessage[]).map(msg => {
      // 后端返回 display_content（蛇形），统一映射为 displayContent
      if (!msg.displayContent && (msg as any).display_content) {
        msg.displayContent = (msg as any).display_content
      }
      // 只有当服务端没有返回时才从 content 正向提取
      if (msg.role === 'user' && !msg.displayContent) {
        msg.displayContent = extractUserDisplayContent(msg.content)
      }
      if (msg.role === 'assistant' && !msg.displayContent) {
        msg.displayContent = computeWriteAssistantDisplayContent(msg.content)
      }
      return msg
    })
    await nextTick()
    scrollToBottom()
  }
}

// 从存储的完整用户消息中提取用户实际输入（前端显示用）
const extractUserDisplayContent = (fullContent: string): string => {
  // 存储格式: 提示词：\n...\n\n世界书：\n...\n\n关联内容：\n...\n\n{用户实际输入}
  // 循环剥离所有已知标签前缀块，剩余即为用户输入
  const labelRE = /^(提示词：|世界书：|关联内容：)[\s\S]*?\n\n(?=提示词：|世界书：|关联内容：|\S|$)/
  let result = fullContent
  let prev = ''
  while (result !== prev) {
    prev = result
    result = result.replace(labelRE, '')
  }
  const extracted = result.trim()
  // 剥离后为空说明本次只有注入内容无用户输入，用占位符避免暴露后端完整内容
  return extracted || '[已注入提示词/关联内容]'
}

// 获取消息的前端显示内容：system消息不显示，user消息只显示用户实际输入
const getUserDisplayContent = (msg: ChatMessage): string => {
  if (msg.role === 'user') return msg.displayContent || extractUserDisplayContent(msg.content)
  return msg.content
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

  return historyRecord.id
}

// 历史记录功能已迁移到 HistoryDialog 组件

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

const renameConversation = (conv: any) => {
  handleConversationCommand('rename', conv)
}

const deleteConversation = async (conv: any) => {
  handleConversationCommand('delete', conv)
}

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
        const target = conversations.value.find(c => c.id === conv.id)
        if (target) {
          target.title = value
        }
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
      
      const target = conversations.value.find(c => c.id === conv.id)
      if (target) {
        target.messages = []
        target.message_count = 0
      }
      
      if (currentConversation.value && currentConversation.value.id === conv.id) {
        chatMessages.value = []
      }
      
      ElMessage.success('消息已清空')
      await fetchConversations()
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
        conversations.value = conversations.value.filter(c => c.id !== conv.id)
        
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

// 监听引用变化，同步 Tags 显示
watch([selectedPrompts, relatedContent], () => {
  syncAttachedReferences()
}, { deep: true })

// 切换章节/备忘录时重置编辑器滚动位置到顶部
const resetEditorScroll = () => {
  // 延迟执行，等待 SplitRichTextEditor 内容渲染完成
  setTimeout(() => {
    const surface = document.querySelector('.editor-surface') as HTMLElement | null
    if (surface) surface.scrollTop = 0
    const shell = document.querySelector('.editor-shell') as HTMLElement | null
    if (shell) shell.scrollTop = 0
  }, 100)
}
watch(currentChapter, () => { resetEditorScroll() })
watch(currentMemo, () => { resetEditorScroll() })

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

watch(showGraphPanel, (val) => {
  if (val) {
    document.body.classList.add('knowledge-graph-open')
  } else {
    document.body.classList.remove('knowledge-graph-open')
  }
})

const onGraphAnalyzeStart = () => {}

const onGraphAnalyzeEnd = (_data: KnowledgeGraphData) => {}

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
  creative2WaitingForResponse.value = true
  creative2Result.value = ''
  creative2ResultDialogVisible.value = true
  creative2AbortController.value = new AbortController()

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
        }],
        temperature: modelTemperature.value,
        top_p: modelTopP.value
      }),
      signal: creative2AbortController.value.signal
    })

    if (!response.ok) {
      throw new Error(`API请求失败: ${response.status}`)
    }

    const reader = response.body?.getReader()
    const decoder = new TextDecoder()

    if (reader) {
      let isFirstChunk = true
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
                if (isFirstChunk) {
                  creative2WaitingForResponse.value = false
                  isFirstChunk = false
                }
                await new Promise(resolve => setTimeout(resolve, 0))
                if (streamingScrollbarRef.value) {
                  const scrollbar = streamingScrollbarRef.value
                  if (scrollbar.wrapRef) {
                    scrollbar.wrapRef.scrollTop = scrollbar.wrapRef.scrollHeight
                  }
                }
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
    if (error.name === 'AbortError') {
      ElMessage.info('已停止生成')
      saveHistoryRecord({
        title: `创意生成 - ${prompts.value.filter(p => promptsToUse.includes(p.id)).map(p => p.name).join('、') || '未命名任务'}`,
        source: 'creative2',
        sourceLabel: '正文创意生成',
        promptName: prompts.value.filter(p => promptsToUse.includes(p.id)).map(p => p.name).join('、'),
        promptCount: prompts.value.filter(p => promptsToUse.includes(p.id)).length,
        status: 'cancelled',
        previewContent: (creative2Result.value || '已停止生成').substring(0, 140),
        messages: buildHistoryMessages({
          promptContents: prompts.value
            .filter(p => promptsToUse.includes(p.id))
            .map(p => p.content),
          userContents: [''],
          assistantContents: creative2Result.value ? [creative2Result.value] : []
        })
      })
    } else {
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
    }
  } finally {
    creative2Generating.value = false
    creative2WaitingForResponse.value = false
    creative2AbortController.value = null
  }
}

const stopCreative2Generation = () => {
  if (creative2AbortController.value) {
    creative2AbortController.value.abort()
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

const applyToCursor = async (content: string) => {
  if (!currentChapter.value) {
    ElMessage.warning('请先选择章节')
    return
  }
  
  try {
    if (chapterEditorRef.value && typeof chapterEditorRef.value.insertPlainText === 'function') {
      await chapterEditorRef.value.insertPlainText(content)
      ElMessage.success('已应用到光标位置')
      saveChapter()
    } else {
      currentChapter.value.content += content
      ElMessage.success('已追加到章节末尾')
      saveChapter()
    }
  } catch (error) {
    ElMessage.error('应用失败')
  }
}

const toggleUserMessage = (index: number) => {
  const newSet = new Set(collapsedUserMessages.value)
  if (newSet.has(index)) {
    newSet.delete(index)
  } else {
    newSet.add(index)
  }
  collapsedUserMessages.value = newSet
}

const startEdit = (index: number, content: string) => {
  messageEditIndex.value = index
  messageEditContent.value = content
  messageEditDialogVisible.value = true
}

const saveMessageEdit = async () => {
  if (messageEditIndex.value === null) return
  
  if (!messageEditContent.value.trim()) {
    ElMessage.warning('消息内容不能为空')
    return
  }
  
  const message = chatMessages.value[messageEditIndex.value]
  if (!message) {
    ElMessage.error('消息不存在')
    return
  }
  
  // 如果消息有 ID，说明已保存到数据库，需要调用 API 更新
  if (message.id && currentConversation.value) {
    const res = await conversationAPI.updateMessage(
      currentConversation.value.id, 
      message.id, 
      { content: messageEditContent.value, displayContent: message.role === 'user' ? extractUserDisplayContent(messageEditContent.value) : undefined }
    )
    if (!res.success) {
      ElMessage.error('保存失败')
      return
    }
  }
  
  // 更新本地内容
  message.content = messageEditContent.value
  if (message.role === 'assistant') {
    message.displayContent = computeWriteAssistantDisplayContent(messageEditContent.value)
  }
  if (message.role === 'user') {
    message.displayContent = extractUserDisplayContent(messageEditContent.value)
  }
  
  messageEditDialogVisible.value = false
  messageEditIndex.value = null
  messageEditContent.value = ''
  ElMessage.success('保存成功')
}

const cancelMessageEdit = () => {
  messageEditDialogVisible.value = false
  messageEditIndex.value = null
  messageEditContent.value = ''
}

const openCreativeFullscreenEditor = (field: string, content: string) => {
  creativeFullscreenField.value = field
  creativeFullscreenContent.value = content
  creativeFullscreenDialogVisible.value = true
}

const saveCreativeFullscreenEditor = () => {
  switch (creativeFullscreenField.value) {
    case 'storyBackground':
      creative2StoryBackground.value = creativeFullscreenContent.value
      break
    case 'characterRelations':
      creative2CharacterRelations.value = creativeFullscreenContent.value
      break
    case 'chapterPlot':
      creative2ChapterPlot.value = creativeFullscreenContent.value
      break
    case 'additionalInfo':
      creative2AdditionalInfo.value = creativeFullscreenContent.value
      break
  }
  creativeFullscreenDialogVisible.value = false
  creativeFullscreenField.value = ''
  creativeFullscreenContent.value = ''
}

const cancelCreativeFullscreenEditor = () => {
  creativeFullscreenDialogVisible.value = false
  creativeFullscreenField.value = ''
  creativeFullscreenContent.value = ''
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
      { content: editingContent.value, displayContent: message.role === 'user' ? extractUserDisplayContent(editingContent.value) : undefined }
    )
    if (!res.success) {
      ElMessage.error('保存失败')
      return
    }
  }
  
  // 更新本地内容
  message.content = editingContent.value
  if (message.role === 'assistant') {
    message.displayContent = computeWriteAssistantDisplayContent(editingContent.value)
  }
  if (message.role === 'user') {
    message.displayContent = extractUserDisplayContent(editingContent.value)
  }
  editingMessageIndex.value = null
  editingContent.value = ''
  ElMessage.success('保存成功')
}

const cancelEdit = () => {
  editingMessageIndex.value = null
  editingContent.value = ''
}

const getContentLength = (content: string) => {
  if (!content) return 0
  const plainText = content
    .replace(/<[^>]*>/g, '')
    .replace(/```[\s\S]*?```/g, '')
    .replace(/`[^`]*`/g, '')
    .replace(/\s+/g, '')
  // 完全对齐 WPS「字符数(不计空格)」：每个非空白字符计 1
  return [...plainText].length
}

const handleTextSelect = (event: Event) => {
  const target = event.target as HTMLTextAreaElement
  const selectedText = target.value.substring(target.selectionStart, target.selectionEnd)
  selectedTextLength.value = selectedText.length > 0 ? getContentLength(selectedText) : 0
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
  selectedTextLength.value = getContentLength(textarea.value)
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

const saveFontFamily = () => {
  localStorage.setItem('editorFontFamily', fontFamily.value)
}

const goToBookAnalysis = () => {
  router.push({ 
    name: 'BookAnalysis', 
    params: { bookId: bookId } 
  })
}

// 角色库弹窗模式
const showCharacterLibrary = ref(false)

const openCharacterLibrary = () => {
  showCharacterLibrary.value = true
}

const closeCharacterLibrary = () => {
  showCharacterLibrary.value = false
}

const handleCharacterSelected = (character: Character) => {
  const insertText = `【角色：${character.name}】\n${character.info || character.personality || ''}`
  // 如果编辑器有插入方法，可在此调用
  closeCharacterLibrary()
}

// 从localStorage加载字体大小
const loadFontSize = () => {
  const savedSize = localStorage.getItem('editorFontSize')
  if (savedSize) {
    fontSize.value = parseInt(savedSize)
  }
}

const loadFontFamily = () => {
  const savedFont = localStorage.getItem('editorFontFamily')
  if (savedFont) {
    fontFamily.value = savedFont
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

const regenerateMessage = async (index: number) => {
  if (sending.value) return
  
  const message = chatMessages.value[index]
  if (!message || message.role !== 'assistant') return
  
  // 彻底删除旧消息（服务端 + 客户端），不留空壳
  if (message.id && currentConversation.value) {
    await conversationAPI.deleteMessage(currentConversation.value.id, message.id)
  }
  chatMessages.value.splice(index, 1)
  
  // 追加新的空 assistant 消息，与 sendMessage 流程一致
  const assistantMessage: ChatMessage = {
    role: 'assistant',
    content: '',
    displayContent: ''
  }
  chatMessages.value.push(assistantMessage)
  const messageIndex = chatMessages.value.length - 1
  
  try {
    sending.value = true
    chatAbortController.value = new AbortController()
    scrollToBottom()
    
    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        messages: chatMessages.value.slice(0, -1),
        configId: selectedConfigId.value,
        temperature: modelTemperature.value,
        top_p: modelTopP.value
      }),
      signal: chatAbortController.value.signal
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
    let rawContent = ''
    
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
              rawContent += parsed.content
              currentMessage.content = rawContent
              currentMessage.displayContent = computeWriteAssistantDisplayContent(rawContent)
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
    
    const savedMessage = chatMessages.value[messageIndex]
    if (savedMessage?.content && currentConversation.value) {
      const actualContent = applyWriteRegexPipeline(savedMessage.content, true)
      savedMessage.content = actualContent
      savedMessage.displayContent = computeWriteAssistantDisplayContent(actualContent)
      await conversationAPI.saveMessage(currentConversation.value.id, {
        role: 'assistant',
        content: actualContent,
        displayContent: savedMessage.displayContent
      })
      // 不调用 fetchConversations，避免把已删除的旧消息重新拉回来
    } else {
      chatMessages.value.splice(messageIndex, 1)
      ElMessage.error('AI未返回内容')
    }
  } catch (error: any) {
    if (error.name !== 'AbortError') {
      ElMessage.error('重新生成失败')
    }
  } finally {
    sending.value = false
    chatAbortController.value = null
  }
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

// RenameDialog 回调
const handleConversationRenamed = async (newTitle: string) => {
  if (renamingConversation.value) {
    await fetchConversations()
    if (currentConversation.value?.id === renamingConversation.value.id) {
      currentConversation.value.title = newTitle
    }
  }
  showRenameDialog.value = false
  ElMessage.success('重命名成功')
}

const selectChapter = (chapter: Chapter) => {
  catalogType.value = 'chapters'
  currentChapter.value = chapter
  currentMemo.value = null
  // 记住当前章节位置
  try { localStorage.setItem(`novel_last_chapter_${bookId}`, String(chapter.id)) } catch {}
}

const selectMemo = (memo: Memo) => {
  currentMemo.value = memo
  currentChapter.value = null
  // 记住当前备忘录位置
  try { localStorage.setItem(`novel_last_memo_${bookId}`, String(memo.id)) } catch {}
}

const handleGlobalMemoSelect = (memo: Memo) => {
  selectMemo(memo)
  showGlobalMemoDialog.value = false
}

const handleCreateChapter = async (folderId: number | null = null) => {
  const res = await chapterAPI.create({
    book_id: bookId,
    title: '新章节',
    content: '',
    summary: '',
    order_num: chaptersList.value.length,
    type: 'chapter',
    volume_id: folderId || undefined
  })
  if (res.success && res.data) {
    chaptersList.value.push(res.data)
    currentChapter.value = res.data
    currentMemo.value = null
    if (folderId && !expandedFolderIds.value.includes(folderId)) {
      expandedFolderIds.value.push(folderId)
    }
  }
}

const handleCreateFolder = (parentId: number | null = null) => {
  isEditVolume.value = false
  editingFolderParentId.value = parentId
  volumeForm.value = {
    id: 0,
    title: ''
  }
  showVolumeDialog.value = true
}

const editFolder = (volume: Volume) => {
  isEditVolume.value = true
  editingFolderParentId.value = volume.parent_id || null
  volumeForm.value = {
    id: volume.id,
    title: volume.title
  }
  showVolumeDialog.value = true
}

const deleteFolder = async (volume: Volume) => {
  try {
    await ElMessageBox.confirm(`确定删除文件夹"${volume.title}"吗？`, '提示', {
      type: 'warning'
    })
    
    const res = await volumeAPI.delete(volume.id)
    if (res.success) {
      volumes.value = volumes.value.filter(v => v.id !== volume.id)
      expandedFolderIds.value = expandedFolderIds.value.filter(id => id !== volume.id)
      ElMessage.success('删除成功')
    }
  } catch (error) {
    // 取消删除
  }
}

// VolumeDialog 已自包含，提交后无需额外操作
const handleVolumeSubmitted = () => {
  // VolumeDialog 组件内部已直接操作 volumes 和 expandedFolderIds
}

const openMoveChapterDialog = (chapter: Chapter) => {
  chapterToMove.value = chapter
  moveChapterDialogVisible.value = true
}

const handleChapterMoved = async (targetFolderId: number | null) => {
  if (!chapterToMove.value) return
  const chapter = chapterToMove.value
  const res = await chapterAPI.update(chapter.id, {
    title: chapter.title,
    content: chapter.content,
    summary: chapter.summary || '',
    order_num: chapter.order_num,
    volume_id: targetFolderId || undefined
  } as any)

  if (res.success && res.data) {
    syncChapterState(res.data)
    moveChapterDialogVisible.value = false
    chapterToMove.value = null
    ElMessage.success('文件已移动')
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

const deleteMemoFolder = async (folder: string) => {
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

watch(
  [writeDisplayRegexEnabled, writeStreamGuardEnabled, writeCopyUsesFiltered, writeRegexRules],
  () => {
    saveWriteRegexSettings()
    refreshWriteChatDisplayContents()
  },
  { deep: true }
)

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

const scrollToBottom = () => {
  nextTick(() => {
    if (chatMessagesRef.value) {
      chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight
    }
  })
}

const openFullscreenEditor = () => {
  fullscreenEditorContent.value = userInput.value
  fullscreenEditorVisible.value = true
}

const saveFullscreenEditor = () => {
  userInput.value = fullscreenEditorContent.value
}

const sendMessage = async () => {
  if (!currentConversation.value) {
    ElMessage.warning('请先选择或创建对话')
    return
  }

  const selectedPromptRecords = prompts.value.filter((p: any) => selectedPrompts.value.some(id => id == p.id))
  const promptNames = selectedPromptRecords.map((p: any) => p.name).join('、')
  const systemPrompts = selectedPromptRecords.map((p: any) => p.content)
  const relatedContentSummary = relatedContent.value.length > 0
    ? `关联内容：\n${relatedContent.value
        .map(item => `${item.type === 'chapter' ? '章节' : '备忘录'}：${item.title}\n${item.content}`)
        .join('\n\n')}`
    : ''

  const promptContents = selectedPromptRecords.map((p: any) => p.content).join('\n\n')

  // 世界书注入：根据用户输入匹配词条并注入内容（需开启关联）
  let worldBookInjection = ''
  if (worldBookLinked.value) {
    const bookEntries = currentBookWorldBook.value?.entries || []
    if (bookEntries.length > 0) {
      const matchResults = matchWorldBookEntries(userInput.value || '', bookEntries)
      if (matchResults.length > 0) {
        worldBookInjection = injectWorldBookContent('', matchResults).trim()
      }
    }
  }

  let fullUserContent = ''
  if (promptContents) {
    fullUserContent += `提示词：\n${promptContents}\n\n`
  }
  if (worldBookInjection) {
    fullUserContent += `世界书：\n${worldBookInjection}\n\n`
  }
  if (relatedContentSummary) {
    fullUserContent += `${relatedContentSummary}\n\n`
  }
  if (userInput.value.trim()) {
    fullUserContent += userInput.value.trim()
  }

  if (!fullUserContent.trim()) {
    ElMessage.warning('请输入消息或选择提示词/关联内容')
    return
  }

  const userActualInput = userInput.value.trim()
  
  const userMessage: ChatMessage = {
    role: 'user',
    content: fullUserContent,
    displayContent: userActualInput || '[已注入提示词/关联内容]'
  }
  chatMessages.value.push(userMessage)
  
  await conversationAPI.saveMessage(currentConversation.value.id, {
    role: 'user',
    content: fullUserContent,
    displayContent: userActualInput || '[已注入提示词/关联内容]'
  })
  
  userInput.value = ''

  const assistantMessage: ChatMessage = {
    role: 'assistant',
    content: '',
    displayContent: ''
  }
  chatMessages.value.push(assistantMessage)
  const messageIndex = chatMessages.value.length - 1

  try {
    sending.value = true
    chatAbortController.value = new AbortController()
    scrollToBottom()

    // 使用 fetch 进行流式请求
    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        messages: chatMessages.value.slice(0, -1),
        configId: selectedConfigId.value,
        temperature: modelTemperature.value,
        top_p: modelTopP.value
      }),
      signal: chatAbortController.value.signal
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
    let rawContent = ''

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
              rawContent += parsed.content
              currentMessage.content = rawContent
              currentMessage.displayContent = computeWriteAssistantDisplayContent(rawContent)
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

    // 保存助手消息（内容已在上方逐块过正则）
    const savedMessage = chatMessages.value[messageIndex]
    if (savedMessage?.content) {
      const actualContent = applyWriteRegexPipeline(savedMessage.content, true)
      savedMessage.content = actualContent
      savedMessage.displayContent = computeWriteAssistantDisplayContent(actualContent)
      await conversationAPI.saveMessage(currentConversation.value.id, {
        role: 'assistant',
        content: actualContent,
        displayContent: savedMessage.displayContent
      })
      await fetchConversations()
      saveHistoryRecord({
        title: `正文对话 - ${currentConversation.value?.title || '新对话'}`,
        source: 'chat',
        sourceLabel: '正文对话',
        promptName: promptNames,
        promptCount: systemPrompts.length,
        status: 'completed',
        previewContent: actualContent.substring(0, 140),
        messages: buildHistoryMessages({
          systemContents: relatedContentSummary ? [relatedContentSummary] : [],
          promptContents: systemPrompts,
          userContents: chatMessages.value
            .slice(0, -1)
            .filter(message => message.role === 'user')
            .map(message => message.content),
          assistantContents: [actualContent]
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
        previewContent: fullUserContent.substring(0, 140),
        messages: buildHistoryMessages({
          systemContents: relatedContentSummary ? [relatedContentSummary] : [],
          promptContents: systemPrompts,
          userContents: chatMessages.value
            .filter(message => message.role === 'user')
            .map(message => message.content)
        })
      })
    }

  } catch (error: any) {
    if (error.name === 'AbortError') {
      const partialContent = chatMessages.value[messageIndex]?.content || ''
      if (partialContent) {
        const actualContent = applyWriteRegexPipeline(partialContent, true)
        if (chatMessages.value[messageIndex]) {
          chatMessages.value[messageIndex].content = actualContent
          chatMessages.value[messageIndex].displayContent = computeWriteAssistantDisplayContent(actualContent)
        }
        await conversationAPI.saveMessage(currentConversation.value.id, {
          role: 'assistant',
          content: actualContent
        })
        await fetchConversations()
        saveHistoryRecord({
          title: `正文对话 - ${currentConversation.value?.title || '新对话'}`,
          source: 'chat',
          sourceLabel: '正文对话',
          promptName: promptNames,
          promptCount: systemPrompts.length,
        status: 'cancelled',
        previewContent: actualContent.substring(0, 140),
          messages: buildHistoryMessages({
            systemContents: relatedContentSummary ? [relatedContentSummary] : [],
            promptContents: systemPrompts,
            userContents: chatMessages.value
              .filter(message => message.role === 'user')
              .map(message => message.content),
            assistantContents: [actualContent]
          })
        })
        ElMessage.info('已停止生成，部分内容已保存')
      } else {
        chatMessages.value.pop()
        saveHistoryRecord({
          title: `正文对话 - ${currentConversation.value?.title || '新对话'}`,
          source: 'chat',
          sourceLabel: '正文对话',
          promptName: promptNames,
          promptCount: systemPrompts.length,
          status: 'cancelled',
          previewContent: fullUserContent.substring(0, 140),
          messages: buildHistoryMessages({
            systemContents: relatedContentSummary ? [relatedContentSummary] : [],
            promptContents: systemPrompts,
            userContents: chatMessages.value
              .filter(message => message.role === 'user')
              .map(message => message.content),
            assistantContents: []
          })
        })
        ElMessage.info('已停止生成')
      }
    } else {
      const partialAssistantContent = chatMessages.value[messageIndex]?.content || ''
      chatMessages.value.pop()
      saveHistoryRecord({
        title: `正文对话 - ${currentConversation.value?.title || '新对话'}`,
        source: 'chat',
        sourceLabel: '正文对话',
        promptName: promptNames,
        promptCount: systemPrompts.length,
        status: 'failed',
        previewContent: (partialAssistantContent || fullUserContent).substring(0, 140),
        messages: buildHistoryMessages({
          systemContents: relatedContentSummary ? [relatedContentSummary] : [],
          promptContents: systemPrompts,
          userContents: chatMessages.value
            .filter(message => message.role === 'user')
            .map(message => message.content),
          assistantContents: partialAssistantContent ? [partialAssistantContent] : []
        })
      })
      ElMessage.error(error.message || '发送失败')
    }
  } finally {
    sending.value = false
    chatAbortController.value = null
  }
}

const stopChatGeneration = () => {
  if (chatAbortController.value) {
    chatAbortController.value.abort()
    chatAbortController.value = null
    sending.value = false
    ElMessage.info('已停止生成')
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

/* ========== 顶部工具栏 ========== */
.toolbar {
  padding: 0 20px;
  height: 60px;
  background: transparent;
  color: #333;
  display: flex;
  justify-content: flex-start;
  align-items: center;
  border-bottom: none;
  position: relative;
  overflow: hidden;
  flex-shrink: 0;
  margin-bottom: 4px;
  gap: 10px;
  transition: height 0.1s ease;
}

.toolbar::before {
  content: none;
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
  color: #333;
  font-weight: 600;
  font-size: 14px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.toolbar :deep(.el-breadcrumb__inner:hover) {
  color: #00a896;
}

.toolbar :deep(.el-breadcrumb__separator) {
  color: #999;
}

.toolbar-leading {
  display: flex;
  align-items: center;
  gap: 10px;
  position: relative;
  z-index: 1;
  min-width: 0;
}

.toolbar-round-btn {
  width: 34px;
  height: 34px;
  border: 1px solid #e0e0e0;
  border-radius: 50%;
  background: #f5f5f5;
  color: #555;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
}

.toolbar-round-btn:hover {
  background: #eee;
  border-color: #ccc;
}

.toolbar-book-chip {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 0;
  padding: 4px 14px;
  border-radius: 17px;
  background: #f0f7f6;
  border: 1px solid #e0ece9;
}

.toolbar-book-name {
  color: #333;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.2;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.toolbar-book-mode {
  color: #888;
  font-size: 11px;
  line-height: 1.2;
  margin-top: 2px;
  letter-spacing: 0.04em;
}

.toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  position: relative;
  z-index: 1;
  align-items: center;
  justify-content: flex-start;
  flex: 1;
}

.toolbar-status {
  padding: 0 2px 0 8px;
  color: #999;
  font-size: 12px;
  white-space: nowrap;
  margin-left: auto;
}

.character-btn,
.analysis-btn,
.ai-write-btn,
.history-btn {
  background: #fff !important;
  border: 1px solid #e0e0e0 !important;
  color: #444 !important;
  padding: 5px 14px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  gap: 5px;
  transition: all 0.2s ease;
  font-size: 13px;
  font-weight: 500;
  box-shadow: none;
}

.character-btn:hover,
.analysis-btn:hover,
.ai-write-btn:hover,
.history-btn:hover {
  background: #f5f5f5 !important;
  border-color: #ccc !important;
  transform: none;
  box-shadow: none;
}

.character-btn .el-icon,
.analysis-btn .el-icon,
.ai-write-btn .el-icon,
.history-btn .el-icon {
  font-size: 15px;
}

.toolbar-icon-btn {
  width: 32px;
  height: 32px;
  padding: 0 !important;
  border-radius: 50%;
  background: #fff !important;
  border: 1px solid #e0e0e0 !important;
  color: #555 !important;
  transition: all 0.2s ease;
  box-shadow: none;
}

.toolbar-icon-btn:hover {
  background: #f5f5f5 !important;
  border-color: #ccc !important;
  color: #333 !important;
  transform: none;
  box-shadow: none;
}

.toolbar-icon-btn .el-icon {
  font-size: 15px;
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
  width: 2px;
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
  background: rgba(16, 185, 129, 0.06);
}

.resize-handle:hover .resize-handle-bar {
  opacity: 1;
  background: linear-gradient(180deg, #10b981 0%, #059669 100%);
}

.resize-handle-bar {
  width: 2px;
  height: 28px;
  background: #d1d5db;
  border-radius: 1px;
  transition: all 0.2s ease;
  opacity: 0.35;
}

.resize-handle:active .resize-handle-bar {
  background: linear-gradient(180deg, #00c9a7 0%, #00a896 100%);
  opacity: 1;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.35);
}

:root[data-theme='dark'] .resize-handle:hover {
  background: rgba(52, 211, 153, 0.08);
}

:root[data-theme='dark'] .resize-handle-bar {
  background: rgba(255, 255, 255, 0.12);
}

/* ========== 左侧目录面板 - 科技卡片风 ========== */
.left-panel {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 4px 5px 5px;
  background: transparent;
  overflow: hidden;
  border-radius: 12px;
}

.left-panel > .el-tabs,
.left-panel :deep(.el-tabs) {
  background: #f6f7f8;
  border-radius: 12px;
  overflow: hidden;
}

.catalog-header {
  padding: 10px 14px;
  border-bottom: 1px solid #ebedf0;
  background: transparent;
  border-radius: 12px 12px 0 0;
}

.catalog-controls {
  display: flex;
  gap: 4px;
  align-items: center;
}

.icon-btn {
  width: 28px;
  height: 28px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 5px;
  transition: background 0.1s ease;
  background: transparent;
  border: none;
}

.icon-btn:hover {
  background: #e2e5e9;
}

.icon-btn .el-icon {
  font-size: 14px;
  color: #707070;
}

.auto-split-btn[disabled] {
  opacity: 0.35;
  pointer-events: none;
  cursor: not-allowed;
}

/* ========== 文件树目录结构 ========== */
.file-tree {
  display: flex;
  flex-direction: column;
}

.tree-node {
  position: relative;
}

.tree-folder {
  display: flex;
  align-items: center;
  gap: 3px;
  padding: 4px 4px;
  cursor: pointer;
  border-radius: 5px;
  transition: background 0.12s ease;
  user-select: none;
  margin-bottom: 1px;
}

.tree-folder:hover {
  background: rgba(0, 201, 167, 0.07);
}

.tree-folder.is-drop-target {
  background: rgba(0, 201, 167, 0.14);
  outline: 1px dashed rgba(0, 168, 150, 0.45);
}

.tree-arrow {
  font-size: 10px;
  color: #909399;
  transition: transform 0.12s ease;
  flex-shrink: 0;
}

.tree-arrow.expanded {
  transform: rotate(90deg);
}

.tree-folder-icon {
  font-size: 14px;
  color: #00c9a7;
  flex-shrink: 0;
}

.tree-folder-name {
  flex: 1;
  font-size: 12.5px;
  font-weight: 500;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.tree-folder-count {
  font-size: 10.5px;
  color: #b0b8c1;
  flex-shrink: 0;
  margin-right: 2px;
  min-width: 16px;
  text-align: right;
}

.tree-folder-actions {
  display: flex;
  gap: 1px;
  flex-shrink: 0;
  opacity: 0;
  transition: opacity 0.12s;
}

.tree-folder:hover .tree-folder-actions {
  opacity: 1;
}

.tree-folder-act-btn {
  width: 20px;
  height: 20px;
  padding: 0;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: #909399;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.12s;
}

.tree-folder-act-btn .el-icon {
  font-size: 12px;
}

.tree-folder-act-btn:hover {
  background: rgba(0, 201, 167, 0.12);
  color: #00a896;
}

.tree-folder-act-danger:hover {
  background: rgba(245, 108, 108, 0.1);
  color: #f56c6c;
}

.tree-children {
  padding-left: 16px;
}

.tree-node-file {
  display: flex;
  align-items: stretch;
  padding: 1px 0;
}

.tree-node-file-body {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px 6px;
  cursor: pointer;
  border-radius: 5px;
  transition: background 0.12s ease;
  flex: 1;
  min-width: 0;
}

.tree-node-file-body:hover {
  background: rgba(0, 201, 167, 0.06);
}

.tree-node-file.active .tree-node-file-body {
  background: rgba(0, 201, 167, 0.12);
  border: 1px solid rgba(0, 201, 167, 0.25);
}

.tree-file-icon {
  font-size: 13px;
  color: #b0b8c1;
  flex-shrink: 0;
}

.tree-node-file.active .tree-file-icon {
  color: #00a896;
}

.tree-file-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.tree-file-top-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.tree-file-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12.5px;
  color: #303133;
  min-width: 0;
}

.tree-file-badge {
  font-size: 9.5px;
  color: #b0b8c1;
  padding: 1px 6px;
  border-radius: 3px;
  background: rgba(0, 201, 167, 0.07);
  flex-shrink: 0;
  line-height: 1.5;
}

.tree-file-sub-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-left: 18px;
}

.tree-file-meta {
  font-size: 9.5px;
  color: #c0c7d0;
}

.tree-file-actions {
  display: flex;
  align-items: center;
  gap: 1px;
  flex-shrink: 0;
  opacity: 0;
  padding: 4px 2px;
  transition: opacity 0.12s;
}

.tree-node-file:hover .tree-file-actions,
.tree-node-file.active .tree-file-actions {
  opacity: 1;
}

.tree-file-act-btn {
  width: 20px;
  height: 20px;
  padding: 0;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: #909399;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.12s;
}

.tree-file-act-btn .el-icon {
  font-size: 12px;
}

.tree-file-act-btn:hover {
  background: rgba(0, 201, 167, 0.1);
  color: #00a896;
}

.tree-summary-btn.filled {
  color: #00a896;
}

.tree-delete-btn:hover {
  background: rgba(245, 108, 108, 0.1);
  color: #f56c6c;
}

.catalog-root-drop {
  padding: 4px 8px;
  --t-fg: #555;
  --t-muted: #999;
  --t-hover: #e2e5e9;
  --t-fg-hover: #111;
  --t-active: #e2e5e9;
  --t-active-fg: #111;
  --t-drop: rgba(8, 109, 221, 0.08);
  --t-drop-border: rgba(8, 109, 221, 0.25);
  --t-btn-hover-fg: #333;
}

.catalog-root-hint {
  height: 4px;
}

.empty-catalog {
  text-align: center;
  color: #999;
  padding: 24px 16px;
  font-size: 12px;
}

/* ========== 左侧面板 - 暗色主题 ========== */
:root[data-theme='dark'] .left-panel {
  background: #181c20;
}

:root[data-theme='dark'] .catalog-header {
  border-bottom-color: #35393e;
  background: transparent;
}

:root[data-theme='dark'] .icon-btn {
  background: transparent;
  border: none;
}

:root[data-theme='dark'] .icon-btn:hover {
  background: #2c313c;
}

:root[data-theme='dark'] .icon-btn .el-icon {
  color: #999;
}

:root[data-theme='dark'] .catalog-root-drop {
  --t-fg: #bababa;
  --t-muted: #666;
  --t-hover: #2c313c;
  --t-fg-hover: #dadada;
  --t-active: #2c313c;
  --t-active-fg: #dadada;
  --t-drop: rgba(2, 122, 255, 0.1);
  --t-drop-border: rgba(2, 122, 255, 0.3);
  --t-btn-hover-fg: #dadada;
}

:root[data-theme='dark'] .empty-catalog {
  color: #666 !important;
}

/* ========== 目录列表滚动条 ========== */
.catalog-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 6px 8px 14px;
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

/* ========== 中间编辑器面板 ========== */
.center-panel {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: transparent;
}

.editor-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 4px 8px 7px;
  animation: slideIn 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.editor-sheet {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  border: 1px solid rgba(0, 201, 167, 0.15);
  border-radius: 20px;
  box-shadow:
    0 1px 3px rgba(0, 201, 167, 0.04),
    0 8px 24px rgba(0, 201, 167, 0.06);
  overflow: hidden;
  transition: border-color 0.3s ease, box-shadow 0.3s ease;
}

.editor-sheet:focus-within {
  border-color: rgba(0, 201, 167, 0.45);
  box-shadow:
    0 1px 3px rgba(0, 201, 167, 0.06),
    0 12px 32px rgba(0, 201, 167, 0.12);
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
  padding: 16px 28px 14px;
}

.chapter-title {
  flex: 1;
  margin-bottom: 0;
}

.word-count-inline {
  flex-shrink: 0;
  font-size: 12px;
  color: #5a8d85;
  white-space: nowrap;
  font-family: 'FangSong', 'STFangsong', '仿宋', serif;
  letter-spacing: 0.04em;
}

.chapter-title :deep(.el-input__wrapper) {
  font-size: 20px;
  font-weight: 600;
  border: none;
  box-shadow: none;
  padding: 4px 0;
  background: transparent;
  border-bottom: 1.5px solid transparent;
  transition: all 0.3s ease;
}

.chapter-title :deep(.el-input__wrapper):hover {
  box-shadow: none;
  border-bottom-color: rgba(0, 201, 167, 0.3);
}

.chapter-title :deep(.el-input__wrapper):focus-within {
  border-bottom-color: #00c9a7;
}

.chapter-title :deep(.el-input__inner) {
  color: #1a1c1e;
  font-weight: 600;
  font-family: 'Georgia', 'Times New Roman', serif;
  letter-spacing: 0.02em;
}

.editor-stats {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  gap: 12px;
  border-top: 1px solid rgba(0, 201, 167, 0.1);
  min-height: 32px;
}

.word-count {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #6c7278;
  white-space: nowrap;
  font-weight: 400;
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
  background: rgba(0, 201, 167, 0.06);
  border: 1px solid rgba(0, 201, 167, 0.12);
  color: #00a896;
  transition: all 0.2s ease;
}

.editor-controls .el-button:hover {
  background: rgba(0, 201, 167, 0.12);
  border-color: rgba(0, 201, 167, 0.25);
  color: #00877a;
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
  font-size: 12px;
  color: #6c7278;
  white-space: nowrap;
  font-weight: 400;
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
  background: #ffffff;
  border: 1px solid rgba(0, 201, 167, 0.15);
  border-radius: 16px;
}

/* 暗色主题适配 - 正文板块 */
:root[data-theme='dark'] .center-panel {
  background: #0f172a;
}

:root[data-theme='dark'] .editor-sheet {
  background: #1e293b;
  border-color: rgba(71, 85, 105, 0.35);
  box-shadow:
    0 1px 3px rgba(0, 0, 0, 0.2),
    0 8px 24px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .editor-sheet:focus-within {
  border-color: rgba(94, 234, 212, 0.35);
  box-shadow:
    0 1px 3px rgba(0, 0, 0, 0.25),
    0 12px 32px rgba(94, 234, 212, 0.08);
}

:root[data-theme='dark'] .word-count-inline {
  color: #5eead4;
}

:root[data-theme='dark'] .chapter-title :deep(.el-input__wrapper) {
  border-bottom-color: transparent;
}

:root[data-theme='dark'] .chapter-title :deep(.el-input__wrapper):hover {
  border-bottom-color: rgba(94, 234, 212, 0.25);
}

:root[data-theme='dark'] .chapter-title :deep(.el-input__wrapper):focus-within {
  border-bottom-color: #5eead4;
}

:root[data-theme='dark'] .chapter-title :deep(.el-input__inner) {
  color: #e5e7eb;
}

:root[data-theme='dark'] .editor-stats {
  border-top-color: rgba(71, 85, 105, 0.3);
}

:root[data-theme='dark'] .word-count {
  color: #9ca3af;
}

:root[data-theme='dark'] .word-count .selected-count {
  color: #5eead4;
}

:root[data-theme='dark'] .editor-controls .el-button {
  background: rgba(94, 234, 212, 0.08);
  border-color: rgba(94, 234, 212, 0.12);
  color: #5eead4;
}

:root[data-theme='dark'] .editor-controls .el-button:hover {
  background: rgba(94, 234, 212, 0.15);
  border-color: rgba(94, 234, 212, 0.25);
  color: #5eead4;
}

:root[data-theme='dark'] .font-size-control .control-label {
  color: #9ca3af;
}

:root[data-theme='dark'] .empty-state {
  background: #1e293b;
  border-color: rgba(71, 85, 105, 0.35);
}

/* ========== 右侧 AI 对话面板 ========== */
.right-panel {
  flex-shrink: 0;
  border-left: 1px solid rgba(16, 185, 129, 0.15);
  display: flex;
  background: linear-gradient(180deg, #f0fdfa 0%, #ffffff 50%, #f0fdfa 100%);
  backdrop-filter: blur(20px);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: -4px 0 32px rgba(16, 185, 129, 0.08);
  overflow: hidden;
}

.right-panel-2 {
  background: #ffffff;
  box-shadow: none;
  border-left-color: rgba(0, 0, 0, 0.06);
}

:root[data-theme='dark'] .right-panel {
  border-left-color: rgba(255, 255, 255, 0.06);
}

:root[data-theme='dark'] .right-panel-2 {
  border-left-color: rgba(255, 255, 255, 0.06);
}

.knowledge-graph-dialog {
  :deep(.el-dialog) {
    margin: 0;
    position: fixed !important;
    top: 0 !important;
    left: 50%;
    transform: translateX(-50%);
    max-height: 100vh;
    overflow: hidden;
  }
  :deep(.el-dialog__header) {
    padding: 12px 20px;
    margin: 0;
    flex-shrink: 0;
  }
  :deep(.el-dialog__body) {
    padding: 0;
    height: calc(100vh - 55px);
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }
}

body.knowledge-graph-open {
  overflow: hidden !important;
}

.graph-btn {
  background: #fff !important;
  border: 1px solid #e0e0e0 !important;
  color: #444 !important;
  padding: 5px 14px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  gap: 5px;
  transition: all 0.2s ease;
  font-size: 13px;
  font-weight: 500;
  box-shadow: none;
}

.graph-btn:hover {
  background: #f5f5f5 !important;
  border-color: #ccc !important;
  box-shadow: none;
  transform: none;
}

.ai-write-btn-2 {
  background: #fff !important;
  border: 1px solid #e0e0e0 !important;
  color: #444 !important;
  padding: 5px 14px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  gap: 5px;
  transition: all 0.2s ease;
  font-size: 13px;
  font-weight: 500;
  box-shadow: none;
}

.ai-write-btn-2:hover {
  background: #f5f5f5 !important;
  border-color: #ccc !important;
  transform: none;
  box-shadow: none;
}

.ai-write-btn-2 .el-icon {
  font-size: 15px;
}

/* ========== AI 抽卡区样式 ========== */
.creative-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
  background: linear-gradient(180deg, rgba(252, 255, 255, 0.7) 0%, rgba(244, 251, 250, 0.78) 100%);
}

.creative-header {
  padding: 10px 14px;
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
  padding: 12px 16px 50px 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.creative-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.fixed-height-section {
  height: 110px;
  min-height: 110px;
  max-height: 110px;
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
  -webkit-line-clamp: 1;
  line-clamp: 1;
  -webkit-box-orient: vertical;
  color: #4b5563;
  font-size: 12px;
  line-height: 1.4;
}

.fixed-height-section .prompt-picker-brief :deep(*) {
  color: #4b5563;
  font-size: 12px;
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
  font-weight: 500;
  color: #134e4a;
  font-size: 13px;
  margin-bottom: 1px;
}

.section-label .el-icon {
  font-size: 14px;
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

.creative-section .el-select .el-input__wrapper {
  border-radius: 6px;
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
  border-radius: 6px;
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
  border-radius: 6px;
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
  border-radius: 6px;
}

.creative-section .el-button--primary:hover {
  background: linear-gradient(135deg, #09b4ad 0%, #047a75 100%);
  border-color: transparent;
}

.creative-section .el-button--primary.is-disabled {
  background: rgba(8, 198, 190, 0.5);
  border-color: transparent;
}

.creative-content::-webkit-scrollbar {
  width: 6px;
}

.creative-content::-webkit-scrollbar-track {
  background: transparent;
}

.creative-content::-webkit-scrollbar-thumb {
  background: rgba(8, 198, 190, 0.2);
  border-radius: 3px;
}

.creative-content::-webkit-scrollbar-thumb:hover {
  background: rgba(8, 198, 190, 0.35);
}

/* 生成按钮固定定位 */
.creative-panel {
  position: relative;
}

.creative-generate-btn-wrapper {
  position: absolute;
  bottom: 16px;
  right: 16px;
  z-index: 10;
}

.creative-generate-btn {
  box-shadow: 0 8px 20px rgba(8, 198, 190, 0.25);
  padding: 10px 24px;
  font-size: 14px;
  font-weight: 600;
}

.creative-generate-btn:hover {
  box-shadow: 0 10px 24px rgba(8, 198, 190, 0.35);
}

/* creative-section 暗色主题 */
:root[data-theme='dark'] .creative-section {
  background: transparent;
  border-color: transparent;
  box-shadow: none;
}

:root[data-theme='dark'] .section-label {
  color: #e5e7eb;
}

:root[data-theme='dark'] .section-label .el-icon {
  color: #5eead4;
}

:root[data-theme='dark'] .fixed-height-section .prompt-picker-brief {
  color: #9ca3af;
}

:root[data-theme='dark'] .fixed-height-section .prompt-picker-brief :deep(*) {
  color: #9ca3af;
}

:root[data-theme='dark'] .fixed-height-section .prompt-picker-brief :deep(strong),
:root[data-theme='dark'] .fixed-height-section .prompt-picker-brief :deep(b) {
  color: #e5e7eb;
}

:root[data-theme='dark'] .creative-section .el-select .el-input__wrapper {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .creative-section .el-select .el-input__wrapper:hover {
  border-color: rgba(94, 234, 212, 0.4);
}

:root[data-theme='dark'] .creative-section .el-select .el-input__wrapper:focus-within {
  border-color: #00c9a7;
}

:root[data-theme='dark'] .creative-section .el-textarea__inner {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(71, 85, 105, 0.4);
  color: #f3f4f6;
}

:root[data-theme='dark'] .creative-section .el-textarea__inner:hover {
  border-color: rgba(94, 234, 212, 0.4);
}

:root[data-theme='dark'] .creative-section .el-textarea__inner:focus {
  border-color: #00c9a7;
}

:root[data-theme='dark'] .creative-section .el-input__wrapper {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .creative-section .el-input__wrapper:hover {
  border-color: rgba(94, 234, 212, 0.4);
}

:root[data-theme='dark'] .creative-section .el-input__wrapper:focus-within {
  border-color: #00c9a7;
}

:root[data-theme='dark'] .creative-section .el-input__inner {
  color: #f3f4f6;
}

:root[data-theme='dark'] .fixed-prompt-label {
  color: #6b7280;
}

:root[data-theme='dark'] .intro-text {
  color: #d1d5db;
  background: rgba(51, 65, 85, 0.6);
}

:root[data-theme='dark'] .prompt-intro-content h3 {
  color: #f3f4f6;
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
  min-height: 0;
  overflow: hidden;
}

.chat-header {
  padding: 8px 10px;
  border-bottom: 1px solid rgba(16, 185, 129, 0.15);
  display: flex;
  align-items: center;
  gap: 8px;
  background: linear-gradient(180deg, rgba(16, 185, 129, 0.06) 0%, rgba(6, 78, 59, 0.02) 100%);
}

.chat-header .el-button {
  width: 32px;
  height: 32px;
  padding: 0;
  background: rgba(255, 255, 255, 0.9) !important;
  border: 1px solid rgba(16, 185, 129, 0.25) !important;
  color: #059669 !important;
  border-radius: 50%;
  transition: all 0.25s ease;
}

.chat-header .el-button:hover {
  background: rgba(16, 185, 129, 0.12) !important;
  border-color: rgba(16, 185, 129, 0.45) !important;
  box-shadow: 0 2px 12px rgba(16, 185, 129, 0.2);
  transform: translateY(-1px);
}

.chat-header .el-button--primary {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%) !important;
  border: none !important;
  color: #fff !important;
  box-shadow: 0 3px 12px rgba(16, 185, 129, 0.35);
}

.chat-header .el-button--primary:hover {
  box-shadow: 0 5px 20px rgba(16, 185, 129, 0.45);
}

.chat-header .header-badge .el-badge__content {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  border: none;
  font-size: 10px;
  height: 16px;
  line-height: 16px;
  padding: 0 4px;
}

.chat-title {
  flex: 1;
  font-weight: 700;
  font-size: 15px;
  color: #065f46;
}

.chat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background: transparent;
  min-height: 0;
}

.chat-messages::-webkit-scrollbar {
  width: 6px;
}

.chat-messages::-webkit-scrollbar-track {
  background: rgba(16, 185, 129, 0.05);
  border-radius: 3px;
}

.chat-messages::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg, #10b981 0%, #059669 100%);
  border-radius: 3px;
}

.message {
  margin-bottom: 16px;
  display: flex;
  gap: 12px;
  animation: fadeIn 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
}

.message.user {
  flex-direction: row-reverse;
  justify-content: flex-start;
}

.message.assistant {
  flex-direction: row;
  justify-content: flex-start;
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
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 16px;
  transition: all 0.3s ease;
}

.message.user .message-avatar {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: #fff;
  box-shadow: 0 4px 16px rgba(16, 185, 129, 0.35);
}

.message.assistant .message-avatar {
  background: linear-gradient(135deg, #d1fae5 0%, #a7f3d0 100%);
  color: #059669;
  border: 1px solid rgba(16, 185, 129, 0.3);
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.15);
}

.message-wrapper {
  flex: 0 1 auto;
  max-width: 85%;
  min-width: 80px;
  position: relative;
}

.message.user .message-wrapper {
  align-items: flex-end;
}

.message.assistant .message-wrapper {
  align-items: flex-start;
}

.message-content {
  padding: 14px 18px;
  border-radius: 8px;
  word-wrap: break-word;
  line-height: 1.7;
  font-size: 14px;
  position: relative;
  transition: all 0.3s ease;
}

.message.user .message-content {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.1) 0%, rgba(6, 78, 59, 0.05) 100%);
  color: #065f46;
  border: 1px solid rgba(16, 185, 129, 0.25);
  box-shadow: 0 2px 12px rgba(16, 185, 129, 0.1);
}

.message.user .message-content:hover {
  border-color: rgba(16, 185, 129, 0.4);
  box-shadow: 0 4px 20px rgba(16, 185, 129, 0.15);
}

.message.user .message-content :deep(.markdown-body) {
  color: #065f46;
}

.message.user .message-content :deep(.markdown-body code) {
  background-color: rgba(16, 185, 129, 0.15);
  color: #047857;
  border-radius: 4px;
  padding: 2px 6px;
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.message.user .message-content :deep(.markdown-body pre) {
  background: rgba(6, 78, 59, 0.05);
  border: 1px solid rgba(16, 185, 129, 0.15);
  border-radius: 6px;
}

.message.assistant .message-content {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95) 0%, rgba(240, 253, 249, 0.9) 100%);
  color: #065f46;
  border: 1px solid rgba(16, 185, 129, 0.15);
  box-shadow: 0 2px 12px rgba(16, 185, 129, 0.08);
}

.message.assistant .message-content:hover {
  border-color: rgba(16, 185, 129, 0.3);
  box-shadow: 0 4px 20px rgba(16, 185, 129, 0.12);
}

.message-actions {
  display: flex;
  gap: 4px;
  margin-top: 8px;
  opacity: 1;
  transition: opacity 0.2s ease;
  padding-left: 4px;
}

.message-actions .el-button {
  padding: 4px 8px;
  font-size: 12px;
  color: #6b7280 !important;
  background: transparent !important;
  border: 1px solid transparent !important;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.message-actions .el-button:hover {
  color: #059669 !important;
  background: rgba(16, 185, 129, 0.1) !important;
  border-color: rgba(16, 185, 129, 0.2) !important;
}

.message-actions .el-button--danger:hover {
  color: #ef4444 !important;
  background: rgba(239, 68, 68, 0.1) !important;
  border-color: rgba(239, 68, 68, 0.2) !important;
}

.message-actions .el-icon {
  font-size: 14px;
}

.message.assistant .message-content :deep(.markdown-body) {
  color: #065f46;
}

.message.assistant .message-content :deep(.markdown-body code) {
  background-color: rgba(16, 185, 129, 0.12);
  color: #047857;
  border-radius: 4px;
  padding: 2px 6px;
  border: 1px solid rgba(16, 185, 129, 0.15);
}

.message.assistant .message-content :deep(.markdown-body pre) {
  background: rgba(6, 78, 59, 0.04);
  border: 1px solid rgba(16, 185, 129, 0.12);
  border-radius: 6px;
}

.message.assistant .message-content :deep(.markdown-body pre code) {
  background: transparent;
  border: none;
}

.message-content.typing::after {
  content: '▊';
  animation: blink 1s infinite;
  margin-left: 2px;
  color: #10b981;
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
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.03) 0%, rgba(6, 78, 59, 0.02) 100%);
  border-radius: 12px;
  min-height: 0;
  overflow: hidden;
  border: 1px dashed rgba(16, 185, 129, 0.2);
}

.empty-chat :deep(.el-empty__description) {
  color: rgba(6, 95, 70, 0.6);
}

.chat-input-area {
  border-top: 1px solid rgba(16, 185, 129, 0.12);
  padding: 12px 16px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.95) 0%, rgba(240, 253, 249, 0.9) 100%);
  backdrop-filter: blur(10px);
}

.config-row {
  margin-bottom: 10px;
}

.model-select-trigger {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.model-select-trigger:hover {
  background: rgba(255, 255, 255, 0.12);
  border-color: rgba(255, 255, 255, 0.25);
}

.creative-model-trigger {
  width: 100%;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
}

.creative-model-trigger:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.2);
}

.model-select-name {
  font-size: 13px;
  color: #e5e7eb;
}

.model-select-arrow {
  font-size: 12px;
  color: #9ca3af;
}

/* 模型配置弹窗 */
.model-config-dialog :deep(.el-dialog__body) {
  padding: 0;
}

.model-config-layout {
  display: flex;
  height: 460px;
}

.model-list-panel {
  width: 25%;
  border-right: 1px solid rgba(16, 185, 129, 0.15);
  overflow-y: auto;
  background: rgba(16, 185, 129, 0.03);
}

.model-list-title {
  padding: 12px 16px;
  font-size: 13px;
  font-weight: 600;
  color: #065f46;
  border-bottom: 1px solid rgba(16, 185, 129, 0.12);
}

.model-list-item {
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid rgba(16, 185, 129, 0.06);
  transition: all 0.2s;
}

.model-list-item:hover {
  background: rgba(16, 185, 129, 0.08);
}

.model-list-item.active {
  background: rgba(16, 185, 129, 0.15);
  border-left: 3px solid #10b981;
}

.model-item-name {
  font-size: 13px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 2px;
}

.model-item-provider {
  font-size: 11px;
  color: #9ca3af;
}

.model-config-panel {
  flex: 1;
  padding: 20px 24px;
  overflow-y: auto;
}

.model-config-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 模型描述预览/编辑 */
.model-desc-form-item :deep(.el-form-item__content) {
  flex-wrap: nowrap;
}

.model-desc-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  width: 100%;
}

.model-desc-preview {
  flex: 1;
  border: 1px solid rgba(16, 185, 129, 0.15);
  border-radius: 4px;
  padding: 6px 10px;
  min-height: 32px;
  max-height: 200px;
  overflow-y: auto;
  overflow-x: hidden;
  background: rgba(16, 185, 129, 0.02);
  min-width: 0;
}

.model-desc-content {
  font-size: 13px;
  line-height: 1.5;
  word-break: break-word;
}

.model-desc-content :deep(.markdown-body) {
  font-size: 13px;
}

.model-desc-content :deep(.markdown-body p) {
  margin: 0;
}

.model-desc-content :deep(.markdown-body img),
.model-desc-content :deep(.markdown-body table),
.model-desc-content :deep(.markdown-body pre) {
  max-width: 100%;
}

.model-desc-empty {
  display: flex;
  align-items: center;
  height: 100%;
  min-height: 20px;
  color: #9ca3af;
  font-size: 13px;
  font-style: italic;
}

.model-desc-edit-btn {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
}

.model-desc-edit-btn .el-icon {
  font-size: 13px;
}

.model-desc-editing {
  width: 100%;
}

.model-desc-edit-box {
  flex: 1;
  border: 1px solid rgba(16, 185, 129, 0.2);
  border-radius: 4px;
  padding: 4px;
  background: rgba(16, 185, 129, 0.02);
}

.model-desc-edit-box :deep(.el-textarea__inner) {
  border: none;
  box-shadow: none;
  background: transparent;
  resize: none;
  font-size: 13px;
  line-height: 1.5;
  padding: 4px 6px;
}

.model-desc-editing-actions {
  display: flex;
  justify-content: flex-end;
  gap: 6px;
  padding: 4px 6px 2px;
}

.slider-with-value {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.slider-value {
  font-size: 13px;
  font-weight: 600;
  color: #065f46;
  min-width: 36px;
  text-align: right;
}

.chat-header .model-config-btn {
  width: 32px;
  height: 32px;
  padding: 0;
  border-radius: 50%;
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.12) 0%, rgba(5, 150, 105, 0.08) 100%);
  border: 1px solid rgba(16, 185, 129, 0.25);
  color: #065f46;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.08);
  transition: all 0.25s ease;
}

.chat-header .model-config-btn .el-icon {
  font-size: 15px;
  color: #10b981;
}

.chat-header .model-config-btn:hover {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.22) 0%, rgba(5, 150, 105, 0.14) 100%);
  border-color: rgba(16, 185, 129, 0.5);
  color: #047857;
  box-shadow: 0 4px 14px rgba(16, 185, 129, 0.16);
  transform: translateY(-1px);
}

.chat-header .model-config-btn:active {
  transform: translateY(0);
}

.input-wrapper {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.input-container {
  flex: 1;
  position: relative;
}

.input-container .fullscreen-btn {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 24px;
  height: 24px;
  padding: 0;
  border: none;
  background: rgba(255, 255, 255, 0.95);
  color: #6b7280;
  z-index: 10;
  transition: all 0.2s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.input-container .fullscreen-btn:hover {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  box-shadow: 0 2px 6px rgba(16, 185, 129, 0.2);
}

/* @ 引用功能样式 */
.textarea-with-at {
  position: relative;
}

.at-trigger-btn {
  position: absolute;
  left: 6px;
  bottom: 6px;
  width: 26px !important;
  height: 26px !important;
  padding: 0 !important;
  border: none !important;
  background: rgba(255, 255, 255, 0.92) !important;
  color: #10b981 !important;
  font-size: 13px;
  z-index: 10;
  transition: all 0.2s ease;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
}

.at-trigger-btn:hover {
  background: rgba(16, 185, 129, 0.12) !important;
  color: #059669 !important;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.2);
}

.at-symbol {
  font-weight: 700;
  font-size: 13px;
  line-height: 1;
}

.textarea-with-at :deep(.el-textarea__inner) {
  padding-left: 36px !important;
}

/* 暗色主题：@按钮与全屏按钮 */
:root[data-theme='dark'] .at-trigger-btn {
  background: rgba(255, 255, 255, 0.08) !important;
  color: #34d399 !important;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .at-trigger-btn:hover {
  background: rgba(16, 185, 129, 0.15) !important;
  color: #6ee7b7 !important;
}

:root[data-theme='dark'] .fullscreen-btn {
  background: rgba(255, 255, 255, 0.08) !important;
  color: rgba(255, 255, 255, 0.5) !important;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .fullscreen-btn:hover {
  background: rgba(255, 255, 255, 0.12) !important;
  color: rgba(255, 255, 255, 0.8) !important;
}

.at-tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 8px;
  min-height: 0;
}

.at-tags-row :deep(.el-tag) {
  border-radius: 4px;
  font-size: 12px;
}

.at-menu-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.at-menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.at-menu-item:hover {
  background: rgba(16, 185, 129, 0.08);
}

.at-menu-icon {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 15px;
}

.at-menu-label {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.at-menu-hint {
  font-size: 12px;
  color: #999;
  margin-left: auto;
}

.creative-section .input-container {
  position: relative;
}

.creative-section .input-container :deep(.el-input__wrapper) {
  padding-right: 32px;
}

.creative-section .input-container :deep(.el-textarea__inner) {
  padding-right: 32px;
}

.input-wrapper .el-input {
  flex: 1;
}

.input-wrapper :deep(.el-textarea__inner) {
  resize: none;
  border-radius: 8px;
  padding: 12px 16px;
  font-size: 14px;
  line-height: 1.5;
  border: 1px solid rgba(16, 185, 129, 0.25);
  background: rgba(255, 255, 255, 0.95);
  color: #065f46;
  transition: all 0.25s;
}

.input-wrapper :deep(.el-textarea__inner::placeholder) {
  color: rgba(6, 95, 70, 0.45);
}

.input-wrapper :deep(.el-textarea__inner):focus {
  border-color: rgba(16, 185, 129, 0.5);
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
  background: #fff;
}

.input-actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.input-wrapper :deep(.el-button--primary) {
  width: 40px;
  height: 40px;
  padding: 0;
  border: none;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: #fff;
  box-shadow: 0 4px 16px rgba(16, 185, 129, 0.35);
}

.input-wrapper :deep(.el-button--primary:hover) {
  box-shadow: 0 6px 24px rgba(16, 185, 129, 0.45);
  transform: scale(1.05);
}

.input-wrapper :deep(.el-button--danger) {
  width: 40px;
  height: 40px;
  padding: 0;
  border: none;
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: #fff;
  box-shadow: 0 4px 16px rgba(239, 68, 68, 0.35);
}

.input-wrapper :deep(.el-button--danger:hover) {
  box-shadow: 0 6px 24px rgba(239, 68, 68, 0.45);
  transform: scale(1.05);
}

/* ========== AI 写作生成结果弹窗样式 ========== */
.creative2-result-dialog :deep(.el-dialog) {
  border-radius: 12px;
  overflow: hidden;
  min-height: 400px;
}

.creative2-result-content {
  min-height: 300px;
  display: flex;
  flex-direction: column;
}

.streaming-content {
  min-height: 300px;
  display: flex;
  flex-direction: column;
}

.generating-animation {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  min-height: 300px;
}

.starburst-container {
  position: relative;
  width: 120px;
  height: 120px;
  margin-bottom: 24px;
}

.starburst-core {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 20px;
  height: 20px;
  background: radial-gradient(circle, #00c9a7 0%, #00a896 50%, transparent 70%);
  border-radius: 50%;
  animation: corePulse 0.6s ease-in-out infinite alternate;
  z-index: 10;
}

.starburst-ring {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  border: 2px solid rgba(0, 201, 167, 0.3);
  border-radius: 50%;
  animation: ringExpand 2s ease-out infinite;
}

.ring-1 {
  width: 40px;
  height: 40px;
  animation-delay: 0s;
}

.ring-2 {
  width: 60px;
  height: 60px;
  animation-delay: 0.6s;
}

.ring-3 {
  width: 80px;
  height: 80px;
  animation-delay: 1.2s;
}

.starburst-ray {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 3px;
  height: 30px;
  background: linear-gradient(to top, transparent 0%, #00c9a7 50%, transparent 100%);
  transform-origin: center bottom;
  animation: rayFlash 1s ease-in-out infinite alternate;
}

.ray-1 { transform: translate(-50%, -100%) rotate(0deg); animation-delay: 0s; }
.ray-2 { transform: translate(-50%, -100%) rotate(45deg); animation-delay: 0.125s; }
.ray-3 { transform: translate(-50%, -100%) rotate(90deg); animation-delay: 0.25s; }
.ray-4 { transform: translate(-50%, -100%) rotate(135deg); animation-delay: 0.375s; }
.ray-5 { transform: translate(-50%, -100%) rotate(180deg); animation-delay: 0.5s; }
.ray-6 { transform: translate(-50%, -100%) rotate(225deg); animation-delay: 0.625s; }
.ray-7 { transform: translate(-50%, -100%) rotate(270deg); animation-delay: 0.75s; }
.ray-8 { transform: translate(-50%, -100%) rotate(315deg); animation-delay: 0.875s; }

.starburst-sparkle {
  position: absolute;
  width: 8px;
  height: 8px;
  animation: sparkleTwinkle 0.8s ease-in-out infinite alternate;
}

.starburst-sparkle::before,
.starburst-sparkle::after {
  content: '';
  position: absolute;
  background: #00c9a7;
}

.starburst-sparkle::before {
  width: 100%;
  height: 2px;
  top: 50%;
  left: 0;
  transform: translateY(-50%);
}

.starburst-sparkle::after {
  width: 2px;
  height: 100%;
  left: 50%;
  top: 0;
  transform: translateX(-50%);
}

.sparkle-1 { top: 10%; left: 50%; animation-delay: 0s; }
.sparkle-2 { top: 25%; right: 15%; animation-delay: 0.15s; }
.sparkle-3 { top: 50%; right: 5%; animation-delay: 0.3s; }
.sparkle-4 { bottom: 25%; right: 15%; animation-delay: 0.45s; }
.sparkle-5 { bottom: 10%; left: 50%; animation-delay: 0.6s; }
.sparkle-6 { bottom: 25%; left: 15%; animation-delay: 0.75s; }

@keyframes corePulse {
  0% {
    transform: translate(-50%, -50%) scale(1);
    box-shadow: 0 0 10px rgba(0, 201, 167, 0.5);
  }
  100% {
    transform: translate(-50%, -50%) scale(1.5);
    box-shadow: 0 0 30px rgba(0, 201, 167, 0.9), 0 0 50px rgba(0, 201, 167, 0.5);
  }
}

@keyframes ringExpand {
  0% {
    transform: translate(-50%, -50%) scale(0.5);
    opacity: 1;
    border-width: 3px;
  }
  100% {
    transform: translate(-50%, -50%) scale(2);
    opacity: 0;
    border-width: 1px;
  }
}

@keyframes rayFlash {
  0% {
    opacity: 0.3;
    height: 20px;
    filter: brightness(0.8);
  }
  50% {
    opacity: 1;
    height: 40px;
    filter: brightness(1.5);
  }
  100% {
    opacity: 0.5;
    height: 25px;
    filter: brightness(1);
  }
}

@keyframes sparkleTwinkle {
  0% {
    opacity: 0.2;
    transform: scale(0.5) rotate(0deg);
    filter: brightness(0.5);
  }
  100% {
    opacity: 1;
    transform: scale(1.2) rotate(45deg);
    filter: brightness(1.5);
  }
}

.generating-text {
  font-size: 16px;
  color: #666;
  font-weight: 500;
  animation: textBlink 1.5s ease-in-out infinite;
}

@keyframes textBlink {
  0%, 100% {
    opacity: 0.5;
  }
  50% {
    opacity: 1;
  }
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

.result-complete {
  min-height: 300px;
  display: flex;
  flex-direction: column;
}

.empty-result {
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
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
  gap: 8px;
}

.history-item {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 14px 16px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.history-item:hover {
  border-color: #10b981;
  background: linear-gradient(135deg, #f0fdf9 0%, #ecfdf5 100%);
  box-shadow: 0 2px 12px rgba(16, 185, 129, 0.12);
}

.history-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.history-main {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.history-icon-wrapper {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #d1fae5 0%, #a7f3d0 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.2);
}

.history-icon {
  color: #059669;
  font-size: 18px;
}

.history-info {
  flex: 1;
  min-width: 0;
}

.history-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 4px;
}

.history-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-tag {
  flex-shrink: 0;
  border-radius: 999px !important;
  font-size: 11px !important;
  padding: 2px 10px !important;
}

.history-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #9ca3af;
}

.meta-item {
  white-space: nowrap;
}

.meta-item.time {
  color: #d1d5db;
}

.meta-divider {
  color: #ddd;
  margin: 0 2px;
}

.history-actions-mini {
  display: flex;
  gap: 2px;
  flex-shrink: 0;
}

.history-actions-mini .el-button {
  width: 28px;
  height: 28px;
  padding: 0;
  border-radius: 50%;
  font-size: 14px;
  background: transparent !important;
  border: none !important;
  color: #9ca3af !important;
  transition: all 0.15s ease;
}

.history-actions-mini .el-button:hover {
  background: #f3f4f6 !important;
  color: #374151 !important;
}

.history-actions-mini .el-button--danger:hover {
  background: #fef2f2 !important;
  color: #dc2626 !important;
}

/* 暗色模式 */
::root[data-theme='dark'] .history-item {
  background: #1e293b;
  border-color: #334155;
}

::root[data-theme='dark'] .history-item:hover {
  border-color: #10b981;
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.12) 0%, rgba(16, 185, 129, 0.08) 100%);
  box-shadow: 0 2px 12px rgba(16, 185, 129, 0.15);
}

::root[data-theme='dark'] .history-icon-wrapper {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.3) 0%, rgba(16, 185, 129, 0.2) 100%);
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.25);
}

::root[data-theme='dark'] .history-icon {
  color: #34d399;
}

::root[data-theme='dark'] .history-title {
  color: #f1f5f9;
}

::root[data-theme='dark'] .history-meta {
  color: #64748b;
}

::root[data-theme='dark'] .meta-item.time {
  color: #475569;
}

::root[data-theme='dark'] .history-actions-mini .el-button {
  color: #64748b !important;
}

::root[data-theme='dark'] .history-actions-mini .el-button:hover {
  background: #334155 !important;
  color: #94a3b8 !important;
}

::root[data-theme='dark'] .history-actions-mini .el-button--danger:hover {
  background: rgba(220, 38, 38, 0.15) !important;
  color: #f87171 !important;
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

:root[data-theme='dark'] .left-panel :deep(.el-tabs) {
  border-color: rgba(71, 85, 105, 0.4);
  background: linear-gradient(180deg, rgba(30, 41, 59, 0.98) 0%, rgba(30, 41, 59, 0.95) 100%);
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.3), inset 0 1px 0 rgba(255, 255, 255, 0.05);
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

.relate-option-item.worldbook-item {
  padding: 8px 12px;
  border-radius: 6px;
  transition: background-color 0.2s;
}

.relate-option-item.worldbook-item:hover {
  background-color: var(--el-fill-color-light);
}

.worldbook-entry-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.worldbook-entry-title {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.worldbook-entry-keys {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.worldbook-relate-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.worldbook-search {
  margin-bottom: 8px;
}

.worldbook-stats {
  display: flex;
  justify-content: space-between;
  padding-top: 12px;
  border-top: 1px solid var(--el-border-color-lighter);
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.worldbook-empty {
  padding: 40px 0;
  text-align: center;
}

.worldbook-dialog :deep(.el-dialog__body) {
  padding: 0;
}

.worldbook-manager {
  display: flex;
  flex-direction: column;
  height: 520px;
  max-height: 70vh;
}

.worldbook-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.worldbook-list-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.worldbook-grouped-list {
  flex: 1;
  overflow-y: auto;
  padding: 1px 8px 0;
  min-height: 0;
}

.worldbook-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  gap: 8px;
  align-items: center;
}

.worldbook-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.worldbook-empty-state {
  padding: 60px 20px;
  text-align: center;
}

/* 分组网格布局 */
.worldbook-list-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.worldbook-grouped-list {
  flex: 1;
  overflow-y: auto;
  padding: 1px 8px 0;
  min-height: 0;
}

.wb-group-section {
  margin-bottom: 1px;
}

.wb-group-header {
  display: flex;
  align-items: center;
  gap: 3px;
  padding: 2px 6px;
  background: var(--el-fill-color-light);
  border-radius: 3px;
  cursor: pointer;
  user-select: none;
  margin-bottom: 1px;
}

.wb-group-header:hover {
  background: var(--el-fill-color);
}

.group-collapse-icon {
  flex-shrink: 0;
  font-size: 11px;
}

.wb-group-name {
  font-weight: 600;
  font-size: 11px;
  color: var(--el-text-color-primary);
}

.preset-tag {
  flex-shrink: 0;
}

.wb-group-count {
  font-size: 10px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
  margin-left: auto;
}

.wb-group-actions {
  display: flex;
  gap: 1px;
  flex-shrink: 0;
}

.wb-entry-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2px;
  min-width: 0;
}

.wb-entry-card {
  padding: 3px 5px;
  background: var(--el-fill-color-light);
  border-radius: 3px;
  border: 1px solid transparent;
  transition: all 0.15s;
  min-width: 0;
  overflow: hidden;
}

.wb-entry-card:hover {
  background: var(--el-fill-color);
}

.wb-entry-card.disabled {
  opacity: 0.5;
}

.wb-entry-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 2px;
}

.wb-entry-tags {
  display: flex;
  gap: 1px;
  flex-shrink: 0;
}

.wb-entry-title {
  font-weight: 600;
  font-size: 11px;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 2px;
}

.wb-entry-actions {
  display: flex;
  gap: 1px;
  justify-content: flex-end;
}

.wb-pagination {
  display: flex;
  justify-content: center;
  padding: 4px 0;
  border-top: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.worldbook-entry-card {
  padding: 12px;
  margin-bottom: 8px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: 2px solid transparent;
}

.worldbook-entry-card:hover {
  background: var(--el-fill-color);
}

.worldbook-entry-card.active {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}

.worldbook-entry-card.disabled {
  opacity: 0.5;
}

.entry-card-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.entry-card-title {
  flex: 1;
  font-weight: 600;
  font-size: 14px;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.entry-card-keys {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.entry-card-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.worldbook-detail-panel {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.worldbook-detail-panel.empty {
  display: flex;
  align-items: center;
  justify-content: center;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.detail-header h3 {
  margin: 0;
  font-size: 18px;
  color: var(--el-text-color-primary);
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-section.half {
  flex: 1;
}

.detail-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-regular);
}

.detail-value {
  font-size: 14px;
  color: var(--el-text-color-primary);
}

.detail-content-text {
  padding: 12px;
  background: var(--el-fill-color-light);
  border-radius: 6px;
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 13px;
  line-height: 1.6;
  max-height: 200px;
  overflow-y: auto;
}

.detail-row {
  display: flex;
  gap: 24px;
}

.detail-options {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 13px;
}

.detail-options span {
  padding: 4px 12px;
  background: var(--el-fill-color);
  border-radius: 4px;
  color: var(--el-text-color-regular);
}

.empty-text {
  color: var(--el-text-color-placeholder);
  font-size: 13px;
}

.worldbook-editor-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
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

/* memo 对话框暗色主题 */
:root[data-theme='dark'] .memo-content-input-large :deep(.el-textarea__inner) {
  color: #d1d5db;
  background: transparent;
}

:root[data-theme='dark'] .content-footer {
  border-top-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .content-empty {
  color: #6b7280;
}

:root[data-theme='dark'] .global-memo-container {
  border-color: rgba(71, 85, 105, 0.4);
  background: #1e293b;
}

:root[data-theme='dark'] .global-memo-sidebar {
  border-right-color: rgba(71, 85, 105, 0.4);
  background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
}

:root[data-theme='dark'] .memo-workspace-title {
  color: #f3f4f6;
}

:root[data-theme='dark'] .memo-workspace-title .el-icon {
  color: #5eead4;
}

:root[data-theme='dark'] .sidebar-action-primary {
  background: rgba(242, 139, 36, 0.15);
  color: #fbbf24;
}

:root[data-theme='dark'] .sidebar-action-accent {
  background: rgba(94, 234, 212, 0.15);
  color: #5eead4;
}

:root[data-theme='dark'] .sidebar-icon-btn {
  background: rgba(51, 65, 85, 0.6);
  color: #9ca3af;
}

:root[data-theme='dark'] .sidebar-icon-btn:hover {
  background: rgba(71, 85, 105, 0.6);
}

:root[data-theme='dark'] .batch-toolbar {
  background: rgba(242, 139, 36, 0.1);
  border-color: rgba(242, 139, 36, 0.3);
}

:root[data-theme='dark'] .memo-tree-folder-header:hover {
  background: rgba(51, 65, 85, 0.6);
}

:root[data-theme='dark'] .memo-tree-folder-header.active {
  background: rgba(94, 234, 212, 0.15);
}

:root[data-theme='dark'] .memo-tree-folder-header.active .folder-name {
  color: #5eead4;
}

:root[data-theme='dark'] .folder-arrow,
:root[data-theme='dark'] .folder-count {
  color: #9ca3af;
}

:root[data-theme='dark'] .folder-name {
  color: #d1d5db;
}

:root[data-theme='dark'] .folder-count {
  background: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .folder-action-btn:hover {
  background: rgba(71, 85, 105, 0.4);
  color: #5eead4;
}

:root[data-theme='dark'] .folder-action-btn.danger:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

:root[data-theme='dark'] .item-title-btn:hover {
  background: rgba(71, 85, 105, 0.4);
  color: #5eead4;
}

:root[data-theme='dark'] .item-title-btn.danger:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

:root[data-theme='dark'] .folder-option {
  border-color: rgba(71, 85, 105, 0.4);
  background: rgba(30, 41, 59, 0.6);
}

:root[data-theme='dark'] .folder-option:hover,
:root[data-theme='dark'] .folder-option.active {
  border-color: rgba(94, 234, 212, 0.4);
  background: rgba(94, 234, 212, 0.15);
}

:root[data-theme='dark'] .folder-option span {
  color: #d1d5db;
}

:root[data-theme='dark'] .folder-option-count {
  background: rgba(71, 85, 105, 0.4);
  color: #9ca3af;
}

:root[data-theme='dark'] .memo-tree-item:hover {
  background: rgba(51, 65, 85, 0.6);
}

:root[data-theme='dark'] .memo-tree-item.active {
  background: rgba(94, 234, 212, 0.15);
}

:root[data-theme='dark'] .memo-tree-item.active .item-title {
  color: #5eead4;
}

:root[data-theme='dark'] .memo-tree-item.batch-selected {
  background: rgba(242, 139, 36, 0.15);
}

:root[data-theme='dark'] .sidebar-empty {
  color: #6b7280;
}

:root[data-theme='dark'] .sidebar-item:hover {
  background: rgba(51, 65, 85, 0.6);
}

:root[data-theme='dark'] .sidebar-item.active {
  background: rgba(94, 234, 212, 0.15);
  color: #5eead4;
}

:root[data-theme='dark'] .sidebar-item.active .item-title {
  color: #5eead4;
}

:root[data-theme='dark'] .sidebar-item.batch-selected {
  background: rgba(242, 139, 36, 0.15);
}

:root[data-theme='dark'] .item-title {
  color: #d1d5db;
}

:root[data-theme='dark'] .item-meta {
  color: #9ca3af;
}

:root[data-theme='dark'] .item-tag {
  background: rgba(71, 85, 105, 0.4);
  color: #9ca3af;
}

:root[data-theme='dark'] .item-meta-time {
  color: #6b7280;
}

:root[data-theme='dark'] .item-icon-btn {
  background: rgba(51, 65, 85, 0.6);
  color: #9ca3af;
}

:root[data-theme='dark'] .item-icon-btn.pin {
  color: #fbbf24;
}

:root[data-theme='dark'] .item-icon-btn.danger:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

:root[data-theme='dark'] .global-memo-content {
  background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
}

:root[data-theme='dark'] .memo-content-header {
  border-bottom-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .memo-title-input :deep(.el-input__inner) {
  color: #f3f4f6;
}

:root[data-theme='dark'] .memo-content-meta {
  color: #9ca3af;
}

:root[data-theme='dark'] .memo-scope-badge.global {
  color: #34d399;
  background: rgba(52, 211, 153, 0.15);
}

:root[data-theme='dark'] .memo-scope-badge.book {
  color: #60a5fa;
  background: rgba(96, 165, 250, 0.15);
}

/* ========== 自动分卷对话框 ========== */
.auto-split-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.auto-split-info {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px 14px;
  background: rgba(0, 201, 167, 0.06);
  border-radius: 8px;
  border: 1px solid rgba(0, 201, 167, 0.12);
  font-size: 13px;
  color: #374151;
  line-height: 1.6;
}

.auto-split-info .el-icon {
  color: #00a896;
  font-size: 16px;
  flex-shrink: 0;
  margin-top: 1px;
}

.auto-split-info strong {
  color: #00a896;
  font-weight: 600;
}

.auto-split-presets {
  display: flex;
  align-items: center;
  gap: 10px;
}

.auto-split-label {
  font-size: 13px;
  color: #606266;
  white-space: nowrap;
  flex-shrink: 0;
  font-weight: 500;
}

.preset-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.preset-buttons .el-button {
  border-radius: 6px;
  font-size: 12px;
  padding: 5px 12px;
}

.auto-split-custom {
  display: flex;
  align-items: center;
  gap: 10px;
}

.auto-split-suffix {
  font-size: 13px;
  color: #909399;
}

.auto-split-preview {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: rgba(0, 201, 167, 0.04);
  border-radius: 6px;
  font-size: 12.5px;
  color: #606266;
  flex-wrap: wrap;
}

.auto-split-preview strong {
  color: #00a896;
  font-weight: 600;
}

.preview-divider {
  color: #d0d5dd;
}

/* 暗色主题 - 自动分卷 */
:root[data-theme='dark'] .auto-split-info {
  background: rgba(94, 234, 212, 0.08);
  border-color: rgba(94, 234, 212, 0.15);
  color: #d1d5db;
}

:root[data-theme='dark'] .auto-split-info .el-icon {
  color: #5eead4;
}

:root[data-theme='dark'] .auto-split-info strong {
  color: #5eead4;
}

:root[data-theme='dark'] .auto-split-label {
  color: #9ca3af;
}

:root[data-theme='dark'] .auto-split-suffix {
  color: #6b7280;
}

:root[data-theme='dark'] .auto-split-preview {
  background: rgba(94, 234, 212, 0.06);
  color: #9ca3af;
}

:root[data-theme='dark'] .auto-split-preview strong {
  color: #5eead4;
}

:root[data-theme='dark'] .preview-divider {
  color: #4b5563;
}

.fullscreen-editor-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.fullscreen-editor-container :deep(.el-textarea__inner) {
  font-size: 15px;
  line-height: 1.6;
  min-height: 400px;
}

.fullscreen-editor-footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

.fullscreen-editor-footer .char-count {
  font-size: 13px;
  color: #6b7280;
}

.hidden-input {
  display: none;
}

/* ========== 导入章节弹窗 ========== */
.import-chapter-content {
  min-height: 60px;
}

/* 上传区域 */
.import-chapter-upload {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  border: 2px dashed var(--ds-border-default, #d9d9d9);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.25s ease;
  background: rgba(16, 185, 129, 0.02);
}

.import-chapter-upload:hover {
  border-color: #10b981;
  background: rgba(16, 185, 129, 0.06);
}

.upload-icon-wrap {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #10b981, #059669);
  border-radius: 14px;
  margin-bottom: 14px;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.25);
}

.upload-main-icon {
  font-size: 28px;
  color: #fff;
}

.upload-text {
  font-size: 15px;
  font-weight: 600;
  color: var(--ds-text-primary, #333);
  margin: 0 0 6px;
}

.upload-hint {
  font-size: 12.5px;
  color: var(--ds-text-tertiary, #999);
  margin: 0;
}

/* 文件信息栏 */
.import-chapter-file-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: rgba(16, 185, 129, 0.06);
  border-radius: 8px;
  margin-bottom: 12px;
}

.file-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  color: var(--ds-text-primary, #262626);
}

.import-chapter-stats {
  font-size: 13px;
  color: var(--ds-text-secondary, #666);
  padding: 4px 2px 10px;
}

.import-chapter-select-all {
  margin-bottom: 10px;
}

.import-chapter-list {
  max-height: 340px;
  overflow-y: auto;
  padding-right: 4px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.import-chapter-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 14px;
  border: 1.5px solid var(--ds-border-default, #e8e8e8);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #fff;
}

.import-chapter-item:hover {
  border-color: #10b981;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.1);
}

.import-chapter-item.selected {
  border-color: #10b981;
  background: rgba(16, 185, 129, 0.04);
}

.import-chapter-info {
  flex: 1;
  min-width: 0;
}

.import-chapter-title {
  font-weight: 600;
  font-size: 13px;
  color: var(--ds-text-primary, #262626);
  margin-bottom: 3px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.import-chapter-preview {
  font-size: 11.5px;
  color: var(--ds-text-tertiary, #999);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.import-chapter-len {
  flex-shrink: 0;
  font-size: 11px;
  color: var(--ds-text-tertiary, #bbb);
  white-space: nowrap;
  margin-top: 2px;
}

/* 暗色主题 */
:root[data-theme='dark'] .import-chapter-upload {
  border-color: rgba(255, 255, 255, 0.12);
  background: rgba(52, 211, 153, 0.03);
}

:root[data-theme='dark'] .import-chapter-upload:hover {
  border-color: #34d399;
  background: rgba(52, 211, 153, 0.06);
}

:root[data-theme='dark'] .upload-text {
  color: rgba(255, 255, 255, 0.88);
}

:root[data-theme='dark'] .upload-hint {
  color: rgba(255, 255, 255, 0.38);
}

:root[data-theme='dark'] .import-chapter-file-bar {
  background: rgba(52, 211, 153, 0.06);
}

:root[data-theme='dark'] .file-name {
  color: rgba(255, 255, 255, 0.88);
}

:root[data-theme='dark'] .import-chapter-stats {
  color: rgba(255, 255, 255, 0.45);
}

:root[data-theme='dark'] .import-chapter-item {
  border-color: rgba(255, 255, 255, 0.08);
  background: transparent;
}

:root[data-theme='dark'] .import-chapter-item:hover {
  border-color: #34d399;
  box-shadow: 0 2px 8px rgba(52, 211, 153, 0.1);
}

:root[data-theme='dark'] .import-chapter-item.selected {
  border-color: #34d399;
  background: rgba(52, 211, 153, 0.06);
}

:root[data-theme='dark'] .import-chapter-title {
  color: rgba(255, 255, 255, 0.88);
}

:root[data-theme='dark'] .import-chapter-preview {
  color: rgba(255, 255, 255, 0.38);
}

:root[data-theme='dark'] .import-chapter-len {
  color: rgba(255, 255, 255, 0.25);
}

:root[data-theme='dark'] .model-select-trigger {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.12);
}

:root[data-theme='dark'] .model-select-trigger:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.2);
}

:root[data-theme='dark'] .creative-model-trigger {
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(255, 255, 255, 0.1);
}

:root[data-theme='dark'] .creative-model-trigger:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.16);
}

/* ================================
   角色库弹窗 — 遮罩 + 容器 + 动画
   ================================ */

/* 遮罩层 — 撑满全屏，点击背景关闭 */
.character-library-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  backdrop-filter: blur(2px);
}

/* 弹窗容器 — 白底大卡片 */
.character-library-overlay .modal-container {
  width: 100%;
  height: 100%;
  max-width: 1440px;
  max-height: 85vh;
  background: #fff;
  border-radius: 16px;
  box-shadow:
    0 25px 80px rgba(0, 0, 0, 0.35),
    0 0 0 1px rgba(255, 255, 255, 0.08) inset;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  position: relative;
}

/* 右上角关闭按钮 */
.modal-close-btn {
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 2100;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: rgba(0, 0, 0, 0.06);
  color: #666;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  transition: all 0.2s ease;
}

.modal-close-btn:hover {
  background: rgba(0, 0, 0, 0.12);
  color: #333;
  transform: rotate(90deg);
}

/* 缩放 + 淡入过渡动画 */
.modal-scale-enter-active {
  transition: opacity 0.2s ease-out, transform 0.25s cubic-bezier(0.22, 1, 0.36, 1);
}
.modal-scale-leave-active {
  transition: opacity 0.15s ease-in, transform 0.15s ease-in;
}

.modal-scale-enter-from {
  opacity: 0;
  transform: scale(0.95);
}
.modal-scale-leave-to {
  opacity: 0;
  transform: scale(0.98);
}

/* 弹窗内部微调：角色库的 header-bar 上方圆角裁剪 */
.character-library-overlay .header-bar {
  border-radius: 16px 16px 0 0;
}

/* 暗色主题适配 — CSS 变量注入，弹窗内所有 el-* 组件自动暗色 */
:root[data-theme='dark'] .character-library-overlay {
  background: rgba(0, 0, 0, 0.6);

  /* Element Plus 基础色板 */
  --el-bg-color: #1e293b;
  --el-bg-color-overlay: #1e293b;
  --el-text-color-primary: #e5e7eb;
  --el-text-color-regular: #d1d5db;
  --el-text-color-secondary: #94a3b8;
  --el-text-color-placeholder: #64748b;
  --el-border-color: #334155;
  --el-border-color-light: #334155;
  --el-border-color-lighter: #2d3a4e;
  --el-border-color-extra-light: #273445;
  --el-fill-color: #334155;
  --el-fill-color-light: #2d3a4e;
  --el-fill-color-lighter: #273445;
  --el-fill-color-blank: #1e293b;
  --el-color-white: #1e293b;
  --el-color-black: #f3f4f6;
  --el-color-info-light-9: #273445;
  --el-color-info-light-8: #2d3a4e;
  --el-color-info-light-7: #334155;
  --el-color-info-light-3: #64748b;
  --el-color-primary: #5eead4;
  --el-color-primary-light-3: #2dd4bf;
  --el-color-primary-light-5: #14b8a6;
  --el-color-primary-light-7: #0d9488;
  --el-color-primary-light-8: #0f766e;
  --el-color-primary-light-9: #115e59;
  --el-color-primary-dark-2: #99f6e4;
  --el-color-danger: #f87171;
  --el-color-danger-light-3: #ef4444;
  --el-color-danger-light-5: #dc2626;
  --el-color-danger-light-7: #b91c1c;
  --el-color-danger-light-8: #991b1b;
  --el-color-danger-light-9: #7f1d1d;
  --el-color-success: #34d399;
  --el-color-warning: #fbbf24;
  --el-color-info: #94a3b8;
  --el-color-error: #f87171;
  --el-box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);

  /* Button 变量 */
  --el-button-bg-color: #334155;
  --el-button-border-color: #475569;
  --el-button-text-color: #e5e7eb;
  --el-button-hover-bg-color: #475569;
  --el-button-hover-border-color: #5eead4;
  --el-button-hover-text-color: #5eead4;
  --el-button-active-bg-color: #2d3a4e;
  --el-button-active-border-color: #5eead4;
  --el-button-active-text-color: #5eead4;
  --el-button-disabled-bg-color: #273445;
  --el-button-disabled-border-color: #334155;
  --el-button-disabled-text-color: #64748b;

  /* Input / Textarea 变量 */
  --el-input-bg-color: #1e293b;
  --el-input-border-color: #334155;
  --el-input-text-color: #e5e7eb;
  --el-input-placeholder-color: #64748b;
  --el-input-hover-border-color: #5eead4;
  --el-input-focus-border-color: #5eead4;
  --el-input-focus-shadow: 0 0 0 2px rgba(94, 234, 212, 0.15);

  /* Select 变量 */
  --el-select-input-focus-border-color: #5eead4;
  --el-popper-bg-color: #1e293b;
  --el-popper-border-color: #334155;

  /* Dialog 变量 */
  --el-dialog-bg-color: #1e293b;
  --el-dialog-border-color: #334155;
  --el-dialog-title-font-color: #e5e7eb;
  --el-dialog-box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);

  /* Message 变量 */
  --el-message-bg-color: #1e293b;
  --el-message-border-color: #334155;
  --el-message-text-color: #e5e7eb;
  --el-message-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);

  /* MessageBox 变量 */
  --el-messagebox-bg-color: #1e293b;
  --el-messagebox-border-color: #334155;
  --el-messagebox-title-text-color: #e5e7eb;
  --el-messagebox-content-text-color: #d1d5db;
  --el-messagebox-box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);

  /* Notification 变量 */
  --el-notification-bg-color: #1e293b;
  --el-notification-border-color: #334155;
  --el-notification-text-color: #e5e7eb;
  --el-notification-box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);

  /* Tooltip / Popover 变量 */
  --el-popover-bg-color: #1e293b;
  --el-popover-border-color: #334155;
  --el-popover-text-color: #e5e7eb;
  --el-popover-box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);

  /* Tag 变量 */
  --el-tag-bg-color: #273445;
  --el-tag-border-color: #334155;
  --el-tag-text-color: #e5e7eb;
  --el-tag-hover-color: #5eead4;

  /* Table 变量 */
  --el-table-bg-color: #1e293b;
  --el-table-tr-bg-color: #1e293b;
  --el-table-border-color: #273445;
  --el-table-text-color: #e5e7eb;
  --el-table-header-text-color: #94a3b8;
  --el-table-row-hover-bg-color: #273445;
  --el-table-current-row-bg-color: #1a2332;

  /* Checkbox 变量 */
  --el-checkbox-bg-color: #1e293b;
  --el-checkbox-border-color: #475569;
  --el-checkbox-text-color: #e5e7eb;
  --el-checkbox-checked-bg-color: #5eead4;
  --el-checkbox-checked-border-color: #5eead4;
  --el-checkbox-checked-text-color: #5eead4;
  --el-checkbox-input-border-color-hover: #5eead4;

  /* Radio 变量 */
  --el-radio-bg-color: #1e293b;
  --el-radio-border-color: #475569;
  --el-radio-text-color: #e5e7eb;
  --el-radio-checked-text-color: #5eead4;
  --el-radio-input-border-color-hover: #5eead4;

  /* Switch 变量 */
  --el-switch-border-color: #475569;
  --el-switch-bg-color: #334155;

  /* Slider 变量 */
  --el-slider-runway-bg-color: #334155;
  --el-slider-stop-bg-color: #475569;
  --el-slider-bar-bg-color: #5eead4;
  --el-slider-button-bg-color: #5eead4;
  --el-slider-button-border-color: #5eead4;

  /* Tabs 变量 */
  --el-tabs-header-border-color: #273445;
  --el-tab-item-text-color: #94a3b8;
  --el-tab-item-active-text-color: #5eead4;
  --el-tab-item-hover-text-color: #5eead4;
}

:root[data-theme='dark'] .character-library-overlay .modal-container {
  background: #1e293b;
  box-shadow:
    0 25px 80px rgba(0, 0, 0, 0.5),
    0 0 0 1px rgba(255, 255, 255, 0.04) inset;
}

:root[data-theme='dark'] .modal-close-btn {
  background: rgba(255, 255, 255, 0.08);
  color: #aaa;
}
:root[data-theme='dark'] .modal-close-btn:hover {
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
}
</style>
