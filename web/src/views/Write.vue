<template>
  <div class="write-container">
    <!-- ========== 标签工作区骨架 · 菜单栏 ========== -->
    <div class="tw-menu">
      <button type="button" class="tw-back" @click="goBackToBooks" title="返回书架">
        <el-icon><ArrowLeft /></el-icon>
      </button>
      <div class="tw-book-chip">
        <span class="tw-book-name">{{ currentBook?.title || '创作' }}</span>
        <span class="tw-book-mode">{{ currentMemo ? '备忘录' : '正文写作' }}</span>
      </div>
      <!-- 标签栏（内联，可横向滚动） -->
      <div class="tw-tabbar">
        <el-dropdown
          v-for="tab in openTabs"
          :key="tab.key"
          trigger="contextmenu"
          @command="(cmd: string) => handleTabCommand(cmd, tab)"
        >
          <div
            class="tw-tab"
            :class="{ active: tab.key === activeTabKey }"
            @click="activateTab(tab)"
          >
            <span class="tw-tab-title">{{ tabTitle(tab) }}</span>
            <span v-if="dirtyTabs.has(tab.key)" class="tw-tab-dirty" title="未保存"></span>
            <span class="tw-tab-close" @click.stop="requestCloseTab(tab)">×</span>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="save">保存</el-dropdown-item>
              <el-dropdown-item command="saveAll">保存全部</el-dropdown-item>
              <el-dropdown-item divided command="close">关闭</el-dropdown-item>
              <el-dropdown-item command="closeOthers">关闭其他标签</el-dropdown-item>
              <el-dropdown-item command="closeRight">关闭右侧标签</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <button type="button" class="tw-tab-new" @click="createBlankTab" title="新建标签">
          <el-icon><Plus /></el-icon>
        </button>
      </div>
      <el-button
        class="tw-tool"
        :class="{ 'tw-tool-active': activeRightPanel === 'chat' }"
        @click="toggleRightPanel('chat')"
        title="AI 对话"
      >
        <el-icon><ChatDotSquare /></el-icon>
      </el-button>
      <el-button
        class="tw-tool tw-tool-accent"
        :class="{ 'tw-tool-active': activeRightPanel === 'creative' }"
        @click="toggleRightPanel('creative')"
        title="AI 写作"
      >
        <el-icon><Lightning /></el-icon>
      </el-button>
      <el-button class="tw-tool" @click="openCharacterLibrary" title="角色库">
        <el-icon><User /></el-icon>
      </el-button>
      <el-button class="tw-tool" @click="goToBookAnalysis" title="拆书库">
        <el-icon><Notebook /></el-icon>
      </el-button>
      <el-button class="tw-tool" @click="toggleGraphPanel" title="知识图谱">
        <el-icon><Connection /></el-icon>
      </el-button>
      <el-button class="tw-tool" @click="showGlobalMemoDialog = true" title="全局备忘录">
        <el-icon><EditPen /></el-icon>
      </el-button>
      <el-button
        class="tw-tool"
        :class="{ 'tw-tool-active': phonePreviewVisible }"
        @click="phonePreviewVisible = !phonePreviewVisible"
        title="手机端阅读预览"
      >
        <el-icon><Iphone /></el-icon>
      </el-button>
      <el-button class="tw-tool" @click="termLibraryDialogVisible = true" title="词条库">
        <el-icon><Collection /></el-icon>
      </el-button>
      <el-button class="tw-tool" @click="showHistoryCenter = true" title="历史记录">
        <el-icon><Document /></el-icon>
      </el-button>
      <el-button class="tw-tool" @click="showSettingsDrawer = true" title="设置">
        <el-icon><Setting /></el-icon>
      </el-button>
    </div>

    <!-- 三栏布局 -->
    <div class="content-wrapper">
      <!-- 左侧：目录 -->
      <div class="left-panel" :class="{ 'tw-explorer-collapsed': explorerCollapsed }" :style="{ width: (explorerCollapsed ? 0 : leftPanelWidth) + 'px' }">
        <div class="tw-explorer-head">
          <span class="tw-explorer-title">{{ explorerTitle }}</span>
        </div>
        <el-tabs v-model="catalogType" class="tw-explorer-tabs">
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
                  @click="handleCreateFolder(null)"
                  class="icon-btn"
                  title="新建分卷"
                >
                  <el-icon><FolderAdd /></el-icon>
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
                  @click="refreshCatalog"
                  class="icon-btn"
                  title="刷新列表"
                >
                  <el-icon><RefreshRight /></el-icon>
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
                  @chapter-summary="saveChapterSummary"
                  @chapter-generate="onChapterGenerateSummary"
                  @drag-file-start="handleChapterDragStart"
                  @drag-file-end="handleChapterDragEnd"
                  @drag-over-folder="handleFolderDragOver"
                  @drag-leave-folder="handleFolderDragLeave"
                  @drop-folder="handleFolderDrop"
                />
              </div>
            </div>
          </el-tab-pane>

          <!-- 大纲：章节概要聚合（纯信息条目） -->
          <el-tab-pane label="大纲" name="outline">
            <div class="xm-info-list">
              <template v-if="outlineGroups.length">
                <template v-for="group in outlineGroups" :key="group.key">
                  <div class="xm-info-group">{{ group.title }}</div>
                  <div
                    v-for="ch in group.chapters"
                    :key="ch.id"
                    class="xm-info-item"
                    :class="{ 'is-active': ch.id === currentChapter?.id }"
                    @click="selectChapter(ch)"
                  >
                    <div class="xm-info-item__title">{{ ch.title }}</div>
                    <div class="xm-info-item__desc" :class="{ 'is-empty': !ch.summary }">{{ ch.summary || '暂无概要' }}</div>
                  </div>
                </template>
              </template>
              <div v-else class="xm-info-empty">暂无章节</div>
            </div>
          </el-tab-pane>

          <!-- 角色：角色信息条目（点击展开详情） -->
          <el-tab-pane label="角色" name="characters">
            <div class="xm-info-list">
              <template v-if="availableCharacters.length">
                <div
                  v-for="c in availableCharacters"
                  :key="c.id"
                  class="xm-char-item"
                  :class="{ 'is-expanded': expandedCharacterId === c.id }"
                  @click="expandedCharacterId = expandedCharacterId === c.id ? null : c.id"
                >
                  <div class="xm-char-item__head">
                    <span class="xm-char-item__avatar">{{ c.name.slice(0, 1) }}</span>
                    <div class="xm-char-item__meta">
                      <div class="xm-char-item__name">{{ c.name }}</div>
                      <div class="xm-char-item__sub">{{ genderLabel(c.gender) }}<template v-if="c.folder"> · {{ c.folder }}</template></div>
                    </div>
                    <el-icon class="xm-char-item__caret"><ArrowDown /></el-icon>
                  </div>
                  <div v-if="expandedCharacterId === c.id" class="xm-char-item__body">
                    <template v-if="c.personality">
                      <div class="xm-char-item__label">性格</div>
                      <div class="xm-char-item__text">{{ c.personality }}</div>
                    </template>
                    <template v-if="c.info">
                      <div class="xm-char-item__label">设定</div>
                      <div class="xm-char-item__text">{{ c.info }}</div>
                    </template>
                  </div>
                </div>
              </template>
              <div v-else class="xm-info-empty">
                暂无角色
                <el-button size="small" link type="primary" @click.stop="charLibraryDialogVisible = true">打开角色库</el-button>
              </div>
            </div>
          </el-tab-pane>

        </el-tabs>
      </div>

      <!-- 目录与编辑器之间的拖拽调宽手柄 -->
      <div class="tw-resize tw-resize-between" :class="{ 'tw-resize-between--collapsed': explorerCollapsed }" title="拖拽调整目录宽度" @mousedown="onResizeStart('left', $event)">
        <button
          type="button"
          class="tw-resize-collapse"
          :title="explorerCollapsed ? '展开目录' : '收起目录'"
          @mousedown.stop
          @click.stop="explorerCollapsed = !explorerCollapsed"
        >
          <el-icon>
            <component :is="explorerCollapsed ? 'DArrowRight' : 'DArrowLeft'" />
          </el-icon>
        </button>
      </div>

      <!-- 中间：编辑器 -->
      <div class="center-panel">
        <div class="center-editor">
        <div v-if="currentChapter" class="editor-container">
          <div class="editor-sheet" ref="editorSheetRef" @mouseup="handleEditorSelect" @keyup="handleEditorSelect">
            <!-- AI 工具栏（核心固定 + 收藏置顶：常用工具可钉到工具栏，收藏存 localStorage） -->
            <div class="xm-ai-toolbar">
              <el-button size="small" type="warning" plain class="xm-ai-btn" @click="toggleRightPanel('creative')">AI写作</el-button>
              <el-button
                v-for="tool in coreTools"
                :key="tool.id"
                size="small"
                type="warning"
                plain
                class="xm-ai-btn"
                @click="onAiToolClick(tool.id)"
              >{{ tool.label }}</el-button>
              <el-button
                v-for="item in pinnedItems"
                :key="'pin-' + item.id"
                size="small"
                type="warning"
                plain
                class="xm-ai-btn xm-ai-btn--pinned"
                @click="item.open()"
              >{{ item.label }}<span class="xm-ai-unpin" title="取消收藏" @click.stop="toggleAiPin(item.id)">✕</span></el-button>
              <el-dropdown trigger="click" :hide-on-click="false">
                <el-button size="small" type="warning" plain class="xm-ai-btn xm-ai-btn--add" title="收藏常用工具到工具栏">＋</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item v-for="item in aiFavoriteItems" :key="item.id" @click="toggleAiPin(item.id)">
                      {{ item.label }}
                      <span class="xm-pin-state" :class="{ 'is-on': aiPinnedIds.includes(item.id) }">{{ aiPinnedIds.includes(item.id) ? '已收藏 ✕' : '收藏 ＋' }}</span>
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
            <!-- 图标工具行（与编辑器功能栏合并同类项，单行；格式化类按钮已迁移至选区悬浮工具条） -->
            <div class="xm-icon-row">
              <div class="xm-icon-row-left">
                <el-tooltip content="粗体" placement="bottom">
                  <button type="button" class="xm-icon-btn xm-icon-btn--glyph xm-icon-btn--bold" @click="formatInline('bold')">B</button>
                </el-tooltip>
                <span class="xm-icon-sep"></span>
                <el-tooltip content="撤回" placement="bottom">
                  <button type="button" class="xm-icon-btn" @click="editorUndo"><el-icon><RefreshLeft /></el-icon></button>
                </el-tooltip>
                <el-tooltip content="撤销撤回" placement="bottom">
                  <button type="button" class="xm-icon-btn" @click="editorRedo"><el-icon><RefreshRight /></el-icon></button>
                </el-tooltip>
                <el-tooltip content="复制" placement="bottom">
                  <button type="button" class="xm-icon-btn" @click="editorCopy"><el-icon><CopyDocument /></el-icon></button>
                </el-tooltip>
                <el-tooltip content="全选" placement="bottom">
                  <button type="button" class="xm-icon-btn" @click="editorSelectAll"><el-icon><Select /></el-icon></button>
                </el-tooltip>
                <el-tooltip content="替换" placement="bottom">
                  <button type="button" class="xm-icon-btn" @click="openFindDialog('replace')"><el-icon><Switch /></el-icon></button>
                </el-tooltip>
                <span class="xm-icon-sep"></span>
                <el-tooltip content="角色库" placement="bottom">
                  <button type="button" class="xm-icon-btn" @click="charLibraryDialogVisible = true"><el-icon><User /></el-icon></button>
                </el-tooltip>
                <el-tooltip content="高频词透镜" placement="bottom">
                  <button type="button" class="xm-icon-btn" :class="{ 'xm-icon-btn--active': wordFrequencyVisible }" @click="wordFrequencyVisible = true"><el-icon><View /></el-icon></button>
                </el-tooltip>
                <el-tooltip content="统一中文标点" placement="bottom">
                  <button type="button" class="xm-icon-btn" @click="applyPunctuation"><el-icon><MagicStick /></el-icon></button>
                </el-tooltip>
                <el-tooltip content="智能排版" placement="bottom">
                  <button type="button" class="xm-icon-btn" @click="applySmartLayout"><el-icon><SetUp /></el-icon></button>
                </el-tooltip>
              </div>
              <div class="xm-icon-row-right">
                <el-tooltip content="预览" placement="bottom">
                  <button type="button" class="xm-icon-btn" :class="{ 'xm-icon-btn--active': previewMode }" @click="togglePreview"><el-icon><Document /></el-icon></button>
                </el-tooltip>
                <el-tooltip content="全屏编辑" placement="bottom">
                  <button type="button" class="xm-icon-btn" @click="toggleEditorFullscreen"><el-icon><FullScreen /></el-icon></button>
                </el-tooltip>
              </div>
            </div>
            <!-- 标题行 -->
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
              <el-tooltip content="固定章节标题" placement="top">
                <button type="button" class="xm-icon-btn" :class="{ 'xm-icon-btn--active': pinChapterTitle }" @click="pinChapterTitle = !pinChapterTitle">
                  <el-icon><Pushpin /></el-icon>
                </button>
              </el-tooltip>
            </div>
            <SplitRichTextEditor
              ref="chapterEditorRef"
              v-model="currentChapter.content"
              placeholder="开始创作..."
              class="chapter-content"
              :font-family="fontFamily"
              :font-size="fontSize"
              :text-indent="textIndent"
              :paragraph-spacing="paragraphSpacing"
              :paragraph-line="paragraphLine"
              :role-marks="roleMarkNames"
              :term-marks="termMarkNames"
              @blur="saveChapter"
              @mark-click="onMarkClick"
              @open-find="openFindDialog('find')"
              @open-replace="openFindDialog('replace')"
            />
            <!-- 阅读预览覆盖层 -->
            <div v-if="previewMode" class="xm-preview-overlay">
              <div class="xm-preview-bar">
                <span>预览模式（只读）</span>
                <button type="button" class="xm-preview-close" @click="previewMode = false">关闭预览</button>
              </div>
              <div class="xm-preview-body" v-html="previewHtml"></div>
            </div>
            <!-- 固定章节标题：滚动浮现的迷你标题条 -->
            <div v-if="pinChapterTitle && titleChipVisible" class="xm-title-chip">
              {{ currentChapter?.title }}
            </div>
            <!-- 编辑器悬浮导航（对齐星月写作） -->
            <div class="xm-float-nav">
              <el-tooltip content="回到编辑器顶部" placement="left">
                <button type="button" class="xm-float-btn" @click="scrollEditorTo('top')"><el-icon><Top /></el-icon></button>
              </el-tooltip>
              <el-tooltip content="前往编辑器底部" placement="left">
                <button type="button" class="xm-float-btn" @click="scrollEditorTo('bottom')"><el-icon><Bottom /></el-icon></button>
              </el-tooltip>
              <el-tooltip :content="ttsActive ? '停止朗读' : '朗读章节内容'" placement="left">
                <button type="button" class="xm-float-btn xm-float-btn--listen" :class="{ 'xm-icon-btn--active': ttsActive }" @click="toggleTts"><el-icon><Headset /></el-icon></button>
              </el-tooltip>
              <el-tooltip content="一键生成章节配图" placement="left">
                <button type="button" class="xm-float-btn xm-float-btn--draw" :class="{ 'xm-icon-btn--active': illustrationVisible }" @click="openIllustration(false)"><el-icon><Picture /></el-icon></button>
              </el-tooltip>
            </div>
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
              :style="{ fontSize: fontSize + 'px', fontFamily: fontFamily }"
              @blur="saveMemo"
              @select="handleTextSelect"
              @mouseup="handleMouseUp"
              @keyup="handleKeyUp"
            />
          </div>
        </div>
        <div v-else class="empty-state">
          <el-empty description="标签为空，请在左侧选择章节或备忘录" />
        </div>
        </div>
      </div>

      <!-- 右侧单栏（切换式）：AI 对话 / AI 写作 共用一栏 -->
      <div class="tw-right-dock" v-if="activeRightPanel" :style="{ width: rightPanelWidth + 'px' }">
        <div class="tw-resize tw-resize-edge" title="拖拽调整宽度" @mousedown="onResizeStart('right', $event)"></div>
        <ChatPanel
          v-if="activeRightPanel === 'chat'"
          :book-id="bookId"
          :panel-width="rightPanelWidth"
          @close="closeRightPanel"
        />
        <AiSidePanel
          v-else-if="activeRightPanel === 'ai-tool'"
          ref="aiSidePanelRef"
          :api-configs="apiConfigs"
          :available-characters="availableCharacters"
          :terms="termMarkNames"
          @apply="onAiSideApply"
          @export="onAiSideExport"
          @finish="onAiSideFinish"
          @close="activeRightPanel = null"
        />
        <Creative2Panel
          v-else-if="activeRightPanel === 'creative'"
          :book-id="bookId"
          :api-configs="apiConfigs"
          :prompts="prompts"
          :chapters-list="chaptersList"
          :memos="memos"
          :available-characters="availableCharacters"
          :has-chapter="!!currentChapter"
          @apply-result="handleApplyCreative2Result"
          @close="closeRightPanel"
        />
      </div>

      <!-- 右侧栏折叠按钮：与左侧同款圆钮，垂直居中贴于右侧栏左缘（右侧栏整体 v-if 隐藏，折叠即收起） -->
      <button
        v-if="activeRightPanel"
        type="button"
        class="tw-resize-collapse tw-resize-collapse--right"
        :style="{ right: (rightPanelWidth + 1) + 'px' }"
        title="收起右侧栏"
        @click="activeRightPanel = null"
      >
        <el-icon><DArrowRight /></el-icon>
      </button>
    </div>


    <!-- 知识图谱弹窗 -->
    <el-dialog
      v-model="showGraphPanel"
      title="知识图谱"
      width="1200px"
      top="0"
      append-to-body
      destroy-on-close
      class="knowledge-graph-dialog xm-dialog"
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

    <!-- 世界书管理 -->
    <WorldBookManager
      v-model:visible="showRelateDialog"
      :book-id="bookId"
      :book-title="currentBook?.title"
    />

    <!-- 全局备忘录弹窗 -->
    <GlobalMemoDialog
      v-model="showGlobalMemoDialog"
      :book-id="bookId"
      :memos="memos"
      @changed="fetchMemos"
      @jump-to-memo="handleJumpToMemo"
    />

    <!-- 重命名对话框 -->
    <RenameDialog
      v-model:visible="showRenameDialog"
      :conversation="renamingConversation"
      :current-title="renamingConversation?.title || ''"
      @renamed="handleConversationRenamed"
    />

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

    <!-- 历史记录中心（复用主页组件） -->
    <HistoryCenter v-model="showHistoryCenter" append-to-body />

    <ConversationListDialog
      v-model:visible="conversationListVisible"
      :conversations="conversations"
      :current-conversation-id="currentConversation?.id"
      @select="selectConversation"
      @rename="renameConversation"
      @delete="deleteConversation"
    />

    <!-- 导入章节弹窗 -->
    <ImportChapterDialog
      v-model:visible="importChapterDialogVisible"
      :book-id="bookId"
      @imported="fetchChapters"
    />

    <!-- AI 续写/章纲弹窗（对齐星月） -->
    <AiContinueDialog
      v-model:visible="continueDialogVisible"
      :tool="continueTool"
      :chapters-list="sortedChapters"
      :memos="memos"
      :available-characters="availableCharacters"
      :terms="termMarkNames"
      :prompts="prompts"
      @start="onContinueStart"
      @show-last="showAiLastResult"
    />

    <!-- AI审稿 / AI去痕（生成器式弹窗：左模板 + 右表单，对齐 Creative.vue） -->
    <AiReviewDialog
      v-model:visible="reviewDialogVisible"
      :tool="reviewTool"
      :source-text="reviewSourceText"
      :source-title="reviewSourceTitle"
      :current-chapter-id="currentChapter?.id ?? null"
      :chapters-list="sortedChapters"
      :characters="availableCharacters"
      @finish="onAiSideFinish"
      @apply="onAiSideApply"
    />

    <!-- 插画生成 -->
    <IllustrationDialog
      v-model:visible="illustrationVisible"
      :api-configs="apiConfigs"
      :initial-prompt="illustrationPrompt"
      @insert="onIllustrationInsert"
    />

    <!-- 敏感词检测 -->
    <SensitiveWordPanel
      v-model:visible="sensitiveVisible"
      :content="currentChapter?.content || ''"
      @locate="onSensitiveLocate"
    />

    <!-- 角色库弹窗（对齐星月角色卡） -->
    <CharacterLibraryDialog
      v-model:visible="charLibraryDialogVisible"
      :book-id="bookId"
      @changed="fetchCharacters"
    />

    <!-- 词条库弹窗（对齐星月词条卡，管理本书世界书词条） -->
    <TermLibraryDialog
      v-model:visible="termLibraryDialogVisible"
      :book-id="bookId"
    />

    <!-- 高频词透镜 -->
    <WordFrequencyPanel
      v-model:visible="wordFrequencyVisible"
      :content="currentChapter?.content || ''"
      @locate="onWordFreqLocate"
    />

    <!-- 选区悬浮工具条（对齐星月） -->
    <SelectionToolbar
      :visible="selectionBarVisible"
      :x="selectionBarX"
      :y="selectionBarY"
      :count="selectionCount"
      @bold="applySelectionFormat('bold')"
      @highlight="applySelectionFormat('highlight')"
      @clear="applySelectionFormat('clear')"
      @format="onSelectionFormat"
      @format-block="onSelectionFormatBlock"
      @tool="onSelectionTool"
    />

    <!-- 手机端阅读预览（对齐星月：可拖拽 + 左键阅读菜单） -->
    <PhonePreview
      :visible="phonePreviewVisible"
      :title="currentChapter?.title || ''"
      :content="currentChapter?.content || ''"
      :chapter-index="sortedChapters.findIndex(c => c.id === currentChapter?.id)"
      :total-chapters="sortedChapters.length"
      :get-scroll-ratio="getEditorScrollRatio"
      :scroll-to-ratio="scrollEditorToRatio"
      @nav="onPhoneNav"
      @close="phonePreviewVisible = false"
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

    <!-- ========== 关闭未保存确认 ========== -->
    <el-dialog v-model="closeConfirmVisible" title="章节尚未保存" width="400px" align-center append-to-body class="xm-dialog">
      <p class="tw-confirm-text">
        「{{ pendingCloseTab ? tabTitle(pendingCloseTab) : '' }}」有未保存的修改，关闭将丢失这些改动。
      </p>
      <template #footer>
        <el-button @click="closeConfirmVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmDiscardTab">放弃修改</el-button>
        <el-button type="primary" @click="confirmSaveAndCloseTab">保存并关闭</el-button>
      </template>
    </el-dialog>

    <!-- ========== 设置抽屉 ========== -->
    <WriteSettingsDrawer
      v-model:visible="showSettingsDrawer"
      :font-family="fontFamily"
      :font-size="fontSize"
      :text-indent="textIndent"
      :paragraph-spacing="paragraphSpacing"
      :paragraph-line="paragraphLine"
      :role-mark-enabled="roleMarkEnabled"
      :term-mark-enabled="termMarkEnabled"
      :theme-mode="themeMode"
      @font-family-change="(v: string) => { fontFamily = v; saveFontFamily() }"
      @font-size-change="(v: number) => { fontSize = v; saveFontSize() }"
      @text-indent-change="(v: boolean) => textIndent = v"
      @paragraph-spacing-change="(v: number) => paragraphSpacing = v"
      @paragraph-line-change="(v: boolean) => paragraphLine = v"
      @role-mark-change="(v: boolean) => roleMarkEnabled = v"
      @term-mark-change="(v: boolean) => termMarkEnabled = v"
      @theme-mode-change="(v: 'light' | 'dark' | 'system') => themeMode = v"
    />

    <!-- ========== 角色/词条资料卡 ========== -->
    <MarkCardDialog v-model:visible="markCardVisible" :card="markCard" />

    <!-- ========== 查找 / 替换（极简居中浮条） ========== -->
    <FindReplaceDialog
      v-model:visible="findDialogVisible"
      :mode="findMode"
      :find-text="findText"
      :replace-text="replaceText"
      :case-sensitive="findCaseSensitive"
      :match-count="matchCount"
      :match-index="matchIndex"
      @update:find-text="findText = $event"
      @update:replace-text="replaceText = $event"
      @update:case-sensitive="findCaseSensitive = $event"
      @find-next="findNext"
      @find-prev="findPrev"
      @replace-one="replaceOne"
      @replace-all="replaceAll"
      @update:visible="(v: boolean) => { findDialogVisible = v; if (!v) closeFindDialog() }"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, computed, watch, defineAsyncComponent, onUnmounted } from 'vue'
import SplitRichTextEditor from '@/components/SplitRichTextEditor.vue'
import WriteCatalogTree from '@/components/WriteCatalogTree.vue'
import VolumeDialog from '@/components/WriteDialogs/VolumeDialog.vue'
import MoveChapterDialog from '@/components/WriteDialogs/MoveChapterDialog.vue'
import HistoryCenter from '@/components/HistoryCenter.vue'
import ImportChapterDialog from '@/components/WriteDialogs/ImportChapterDialog.vue'
import RenameDialog from '@/components/WriteDialogs/RenameDialog.vue'
import ConversationListDialog from '@/components/WriteDialogs/ConversationListDialog.vue'
import WorldBookManager from '@/components/WriteDialogs/WorldBookManager.vue'
import GlobalMemoDialog from '@/components/WriteDialogs/GlobalMemoDialog.vue'
import Creative2Panel from '@/components/WriteDialogs/Creative2Panel.vue'
import WriteSettingsDrawer from '@/components/WriteSettingsDrawer.vue'
import MarkCardDialog, { type MarkCardData } from '@/components/MarkCardDialog.vue'
import FindReplaceDialog from '@/components/FindReplaceDialog.vue'
import AiSidePanel from '@/components/WriteDialogs/AiSidePanel.vue'
import AiContinueDialog from '@/components/WriteDialogs/AiContinueDialog.vue'
import AiReviewDialog from '@/components/WriteDialogs/AiReviewDialog.vue'
import SelectionToolbar from '@/components/WriteDialogs/SelectionToolbar.vue'
import PhonePreview from '@/components/WriteDialogs/PhonePreview.vue'
import CharacterLibraryDialog from '@/components/WriteDialogs/CharacterLibraryDialog.vue'
import TermLibraryDialog from '@/components/WriteDialogs/TermLibraryDialog.vue'
import WordFrequencyPanel from '@/components/WriteDialogs/WordFrequencyPanel.vue'
import IllustrationDialog from '@/components/WriteDialogs/IllustrationDialog.vue'
import SensitiveWordPanel from '@/components/WriteDialogs/SensitiveWordPanel.vue'
import { AI_TOOLS, getAiTool } from '@/constants/aiTools'
import type { AiToolMeta } from '@/constants/aiTools'
import { useWorldBookStore } from '@/stores/worldbook'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBookStore } from '@/stores/book'
import { chapterAPI, memoAPI, promptAPI, configAPI, conversationAPI, volumeAPI, characterAPI } from '@/api'
import type { Chapter, Memo, Prompt, ApiConfig, ChatMessage, Volume, Character, KnowledgeGraphData } from '@/types'
import { useWriteRegex } from '@/composables/useWriteRegex'
import { getContentLength } from '@/utils/format'
import KnowledgeGraph from '@/components/KnowledgeGraph.vue'
import ChatPanel from '@/components/ChatPanel/index.vue'
import '@/styles/write-theme.css'
import {
  ChatDotSquare, User, Notebook,
  Plus, Folder,
  ArrowLeft, EditPen,
  Search,
  Document, Close, Lightning, Connection, Upload, Setting, RefreshRight,
  FullScreen, Iphone, Collection
} from '@element-plus/icons-vue'

// 延迟加载角色库组件
const CharacterLibraryModal = defineAsyncComponent(() => import('@/views/CharacterLibrary.vue'))

const route = useRoute()
const router = useRouter()
const bookStore = useBookStore()

const bookId = parseInt(route.params.bookId as string) || 0
const currentBook = computed(() => bookStore.currentBook)

// 目录相关
const catalogType = ref<'chapters' | 'outline' | 'characters'>('chapters')
// 左栏顶部标题随 tab 联动
const explorerTitle = computed(() => catalogType.value === 'outline' ? '大纲' : catalogType.value === 'characters' ? '角色' : '章节')
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

// 大纲 tab：按卷分组的章节概要条目（未分卷归入「未分卷」组）
const outlineGroups = computed(() => {
  const groups: Array<{ key: string; title: string; chapters: Chapter[] }> = []
  const grouped = new Set<number>()
  for (const folder of sortedFolders.value) {
    const chapters = sortedChapters.value.filter(c => c.volume_id === folder.id)
    if (!chapters.length) continue
    chapters.forEach(c => grouped.add(c.id))
    groups.push({ key: `v-${folder.id}`, title: folder.title, chapters })
  }
  const loose = sortedChapters.value.filter(c => !grouped.has(c.id))
  if (loose.length) groups.push({ key: 'loose', title: '未分卷', chapters: loose })
  return groups
})

// 角色 tab：手风琴展开的角色 id
const expandedCharacterId = ref<number | null>(null)
const genderLabel = (g: Character['gender']) => ({ male: '男', female: '女', unknown: '未知', none: '无' }[g] || '未知')

// API配置
const apiConfigs = ref<ApiConfig[]>([])
const selectedConfigId = ref<number>()

// 提示词
const prompts = ref<Prompt[]>([])
const availableCharacters = ref<Character[]>([])

// 对话相关
const conversations = ref<any[]>([])
const currentConversation = ref<any>(null)
// 右侧单栏（切换式）：chat / creative 共用一栏，点胶囊切换（outline 已移至左侧标签页）
type RightPanelKey = 'chat' | 'creative' | 'ai-tool'
const activeRightPanel = ref<RightPanelKey | null>(null)
const showGraphPanel = ref(false)
const showGlobalMemoDialog = ref(false)
const showHistoryCenter = ref(false)


const goBackToBooks = () => {
  router.push('/books')
}

const {
  writeDisplayRegexEnabled,
  writeStreamGuardEnabled,
  writeCopyUsesFiltered,
  writeRegexRules,
  computeWriteAssistantDisplayContent,
  loadWriteRegexSettings,
  saveWriteRegexSettings
} = useWriteRegex()

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

const chatMessages = ref<ChatMessage[]>([])

const chatMessagesRef = ref<HTMLElement>()
const selectedTextLength = ref(0)
const fontSize = ref(16)
const fontFamily = ref('Microsoft YaHei')

/* ============ 设置抽屉 ============ */
const showSettingsDrawer = ref(false)
const textIndent = ref(false)
const paragraphSpacing = ref(0)
const paragraphLine = ref(false)
const roleMarkEnabled = ref(false)
const termMarkEnabled = ref(false)
// 与 MainLayout 共用全局主题持久化（main-theme-mode，兼容旧 key main-theme）
const getInitialThemeMode = (): 'light' | 'dark' | 'system' => {
  const saved = localStorage.getItem('main-theme-mode')
  if (saved === 'light' || saved === 'dark' || saved === 'system') return saved
  const legacy = localStorage.getItem('main-theme')
  if (legacy === 'light' || legacy === 'dark') return legacy
  return 'light'
}
const themeMode = ref<'light' | 'dark' | 'system'>(getInitialThemeMode())
const worldBookStore = useWorldBookStore()

const SETTINGS_KEY = 'write-settings'
const loadWriteSettings = () => {
  try {
    const raw = localStorage.getItem(SETTINGS_KEY)
    if (!raw) return
    const s = JSON.parse(raw)
    if (typeof s.textIndent === 'boolean') textIndent.value = s.textIndent
    if (typeof s.paragraphSpacing === 'number') paragraphSpacing.value = s.paragraphSpacing
    if (typeof s.paragraphLine === 'boolean') paragraphLine.value = s.paragraphLine
    if (typeof s.roleMarkEnabled === 'boolean') roleMarkEnabled.value = s.roleMarkEnabled
    if (typeof s.termMarkEnabled === 'boolean') termMarkEnabled.value = s.termMarkEnabled
  } catch {}
}
const saveWriteSettings = () => {
  try {
    localStorage.setItem(SETTINGS_KEY, JSON.stringify({
      textIndent: textIndent.value,
      paragraphSpacing: paragraphSpacing.value,
      paragraphLine: paragraphLine.value,
      roleMarkEnabled: roleMarkEnabled.value,
      termMarkEnabled: termMarkEnabled.value,
    }))
  } catch {}
}
watch(
  [textIndent, paragraphSpacing, paragraphLine, roleMarkEnabled, termMarkEnabled],
  () => saveWriteSettings()
)

/* 主题：浅色 / 深色 / 跟随系统 */
const systemDark = ref(false)
const applyThemeMode = () => {
  const dark = themeMode.value === 'dark' || (themeMode.value === 'system' && systemDark.value)
  const root = document.documentElement
  if (dark) root.setAttribute('data-theme', 'dark')
  else root.removeAttribute('data-theme')
  localStorage.setItem('main-theme-mode', themeMode.value)
  localStorage.setItem('main-theme', dark ? 'dark' : 'light')
  window.dispatchEvent(new CustomEvent('xingmang:theme-changed', { detail: { theme: dark ? 'dark' : 'light' } }))
}
const onSystemThemeChange = (e: MediaQueryListEvent) => {
  systemDark.value = e.matches
  if (themeMode.value === 'system') applyThemeMode()
}
let systemThemeQuery: MediaQueryList | null = null
const initThemeMode = () => {
  systemThemeQuery = window.matchMedia('(prefers-color-scheme: dark)')
  systemDark.value = systemThemeQuery.matches
  systemThemeQuery.addEventListener('change', onSystemThemeChange)
  applyThemeMode()
}
watch(themeMode, () => applyThemeMode())

/* 正文标记数据：角色名 / 词条名 */
const roleMarkNames = computed(() => {
  if (!roleMarkEnabled.value) return []
  return availableCharacters.value.map(c => c.name).filter(Boolean)
})
const termMarkNames = computed(() => {
  if (!termMarkEnabled.value) return []
  const wb = worldBookStore.getBookWorldBook(String(bookId))
  const entries = wb?.entries || []
  const names = new Set<string>()
  entries.forEach(e => {
    if (e.disable) return
    if (e.comment?.trim()) names.add(e.comment.trim())
    ;(e.key || []).forEach(k => { if (k?.trim()) names.add(k.trim()) })
    ;(e.keysecondary || []).forEach(k => { if (k?.trim()) names.add(k.trim()) })
    ;(e.keyword || '').split(/[,，]/).forEach(k => { if (k?.trim()) names.add(k.trim()) })
    ;(e.keywords || '').split(/[,，]/).forEach(k => { if (k?.trim()) names.add(k.trim()) })
  })
  return Array.from(names)
})

/* 资料卡弹窗 */
const markCardVisible = ref(false)
const markCard = ref<MarkCardData | null>(null)
const onMarkClick = (payload: { type: 'role' | 'term'; name: string }) => {
  if (payload.type === 'role') {
    const c = availableCharacters.value.find(ch => ch.name === payload.name)
    markCard.value = { type: 'role', name: payload.name, data: c || null }
  } else {
    const wb = worldBookStore.getBookWorldBook(String(bookId))
    const entry = wb?.entries.find(e =>
      e.comment === payload.name ||
      (e.key || []).includes(payload.name) ||
      (e.keysecondary || []).includes(payload.name) ||
      (e.keyword || '').split(/[,，]/).includes(payload.name) ||
      (e.keywords || '').split(/[,，]/).includes(payload.name)
    )
    markCard.value = { type: 'term', name: payload.name, data: entry || null }
  }
  markCardVisible.value = true
}

/* ============ 查找 / 替换（记事本式） ============ */
const findDialogVisible = ref(false)
const findMode = ref<'find' | 'replace'>('find')
const findText = ref('')
const replaceText = ref('')
const findCaseSensitive = ref(false)
const matchCount = ref(0)
const matchIndex = ref(0)
// 当前匹配在全文中的绝对偏移（Vditor focus 会重渲染导致节点引用失效，
// 因此用偏移而非节点引用定位上一个/下一个）
let lastAbsStart = -1

const openFindDialog = (mode: 'find' | 'replace' = 'find') => {
  findMode.value = mode
  findDialogVisible.value = true
  if (findText.value) {
    nextTick(() => { findNext() })
  }
}

const closeFindDialog = () => {
  findDialogVisible.value = false
  clearFindSelection()
}

const findRegex = () => {
  const q = findText.value
  if (!q) return null
  const flags = findCaseSensitive.value ? 'g' : 'gi'
  try {
    return new RegExp(q.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), flags)
  } catch {
    return null
  }
}

/* ---- Vditor（章节） ---- */
const collectVditorMatches = (): { node: Text; start: number; end: number; absStart: number }[] => {
  const ed = chapterEditorRef.value
  const vd = ed?.vditor?.()
  // Vditor 实例的 wysiwyg DOM 在 vd.vditor.wysiwyg.element（内部 IVditor）
  const root = vd?.vditor?.wysiwyg?.element
  if (!root) return []
  const re = findRegex()
  if (!re) return []
  // absStart：匹配在整篇文本（按文本节点顺序拼接）中的绝对偏移，
  // 用于 findNext/findPrev 定位，不依赖节点引用（Vditor focus 会重渲染导致引用失效）
  const matches: { node: Text; start: number; end: number; absStart: number }[] = []
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT)
  const nodes: Text[] = []
  while (walker.nextNode()) nodes.push(walker.currentNode as Text)
  let absOffset = 0
  for (const node of nodes) {
    const text = node.nodeValue || ''
    if (text) {
      re.lastIndex = 0
      let m: RegExpExecArray | null
      while ((m = re.exec(text))) {
        matches.push({ node, start: m.index, end: m.index + m[0].length, absStart: absOffset + m.index })
      }
    }
    absOffset += text.length
  }
  return matches
}

const selectVditorMatch = (m: { node: Text; start: number; end: number; absStart: number }) => {
  const ed = chapterEditorRef.value
  const vd = ed?.vditor?.()
  const root = vd?.vditor?.wysiwyg?.element
  if (!root) return
  const range = document.createRange()
  range.setStart(m.node, m.start)
  range.setEnd(m.node, m.end)
  const sel = window.getSelection()
  sel?.removeAllRanges()
  sel?.addRange(range)
  lastAbsStart = m.absStart
  // 定位到可视区域
  const rect = range.getBoundingClientRect()
  if (rect) {
    const container = root.closest('.vditor-content') || root
    const cRect = container.getBoundingClientRect()
    if (rect.top < cRect.top || rect.bottom > cRect.bottom) {
      const el = document.createElement('span')
      el.style.position = 'absolute'
      el.style.top = '0'
      el.style.left = '0'
      el.style.width = '1px'
      el.style.height = '1px'
      range.insertNode(el)
      el.scrollIntoView({ block: 'center' })
      el.remove()
    }
  }
  // 查找浮条打开时焦点应留在查找输入框：抢焦点会打断输入法（拼音）组合
  if (!findDialogVisible.value) ed?.focusEditor()
}

const findNext = () => {
  const q = findText.value
  if (!q) return
  if (currentChapter.value) {
    const matches = collectVditorMatches()
    matchCount.value = matches.length
    if (!matches.length) {
      matchIndex.value = 0
      clearFindSelection()
      return
    }
    let idx = 0
    if (lastAbsStart >= 0) {
      // 从上次匹配之后找下一个（按绝对偏移，不依赖节点引用）
      const cur = matches.findIndex(m => m.absStart > lastAbsStart)
      idx = cur >= 0 ? cur : 0
    }
    matchIndex.value = idx
    selectVditorMatch(matches[idx])
  } else if (currentMemo.value) {
    findNextTextarea()
  }
}

const findPrev = () => {
  const q = findText.value
  if (!q) return
  if (currentChapter.value) {
    const matches = collectVditorMatches()
    matchCount.value = matches.length
    if (!matches.length) {
      matchIndex.value = 0
      clearFindSelection()
      return
    }
    let idx = matches.length - 1
    if (lastAbsStart >= 0) {
      // 从上次匹配之前找上一个
      let cur = -1
      for (let i = matches.length - 1; i >= 0; i--) {
        if (matches[i].absStart < lastAbsStart) { cur = i; break }
      }
      idx = cur >= 0 ? cur : matches.length - 1
    }
    matchIndex.value = idx
    selectVditorMatch(matches[idx])
  } else if (currentMemo.value) {
    findPrevTextarea()
  }
}

const clearFindSelection = () => {
  lastAbsStart = -1
  const sel = window.getSelection()
  if (sel) sel.removeAllRanges()
}

const replaceOne = () => {
  const q = findText.value
  if (!q) return
  if (currentChapter.value) {
    const ed = chapterEditorRef.value
    const vd = ed?.vditor?.()
    if (!vd) return
    const matches = collectVditorMatches()
    matchCount.value = matches.length
    if (!matches.length) return
    // 定位当前匹配（优先上次选中的）
    let idx = 0
    if (lastAbsStart >= 0) {
      const cur = matches.findIndex(m => m.absStart >= lastAbsStart)
      idx = cur >= 0 ? cur : 0
    }
    const m = matches[idx]
    const range = document.createRange()
    range.setStart(m.node, m.start)
    range.setEnd(m.node, m.end)
    const sel = window.getSelection()
    sel?.removeAllRanges()
    sel?.addRange(range)
    vd.updateValue(replaceText.value)
    lastAbsStart = -1
    nextTick(() => {
      const ms = collectVditorMatches()
      matchCount.value = ms.length
      if (ms.length) {
        matchIndex.value = Math.min(idx, ms.length - 1)
        selectVditorMatch(ms[matchIndex.value])
      } else {
        matchIndex.value = 0
      }
    })
  } else if (currentMemo.value) {
    replaceOneTextarea()
  }
}

const replaceAll = () => {
  const q = findText.value
  if (!q) return
  if (currentChapter.value) {
    const ed = chapterEditorRef.value
    const vd = ed?.vditor?.()
    if (!vd) return
    const re = findRegex()
    if (!re) return
    const md = vd.getValue() || ''
    const count = (md.match(re) || []).length
    if (!count) return
    re.lastIndex = 0
    const next = md.replace(re, () => replaceText.value)
    vd.setValue(next)
    lastAbsStart = -1
    matchCount.value = 0
    matchIndex.value = 0
    currentChapter.value.content = next
    ElMessage.success(`已替换 ${count} 处`)
  } else if (currentMemo.value) {
    replaceAllTextarea()
  }
}

/* ---- 备忘录 textarea ---- */
const findTextareaMatches = (): { start: number; end: number }[] => {
  const text = currentMemo.value?.content || ''
  const re = findRegex()
  if (!re) return []
  const matches: { start: number; end: number }[] = []
  re.lastIndex = 0
  let m: RegExpExecArray | null
  while ((m = re.exec(text))) {
    matches.push({ start: m.index, end: m.index + m[0].length })
  }
  return matches
}

const memoTextareaEl = (): HTMLTextAreaElement | null => {
  const ta = memoEditorRef.value as any
  if (!ta) return null
  return ta.$el?.querySelector?.('textarea') || ta.$el || null
}

const findNextTextarea = () => {
  const el = memoTextareaEl()
  if (!el) return
  const matches = findTextareaMatches()
  matchCount.value = matches.length
  if (!matches.length) {
    matchIndex.value = 0
    return
  }
  const caret = el.selectionStart ?? 0
  let idx = matches.findIndex(m => m.start > caret)
  if (idx < 0) idx = 0
  matchIndex.value = idx
  if (!findDialogVisible.value) el.focus() // 查找浮条打开时不抢焦点，避免打断输入法
  el.setSelectionRange(matches[idx].start, matches[idx].end)
  el.scrollTop = Math.max(0, el.scrollTop - 40)
}

const findPrevTextarea = () => {
  const el = memoTextareaEl()
  if (!el) return
  const matches = findTextareaMatches()
  matchCount.value = matches.length
  if (!matches.length) {
    matchIndex.value = 0
    return
  }
  const caret = el.selectionStart ?? 0
  let idx = -1
  for (let i = matches.length - 1; i >= 0; i--) {
    if (matches[i].start < caret) { idx = i; break }
  }
  if (idx < 0) idx = matches.length - 1
  matchIndex.value = idx
  if (!findDialogVisible.value) el.focus() // 查找浮条打开时不抢焦点，避免打断输入法
  el.setSelectionRange(matches[idx].start, matches[idx].end)
  el.scrollTop = Math.max(0, el.scrollTop - 40)
}

const replaceOneTextarea = () => {
  const el = memoTextareaEl()
  if (!el) return
  const matches = findTextareaMatches()
  matchCount.value = matches.length
  if (!matches.length) return
  const caret = el.selectionStart ?? 0
  let idx = matches.findIndex(m => m.start === caret && m.end === (el.selectionEnd ?? 0))
  if (idx < 0) idx = matches.findIndex(m => m.start >= caret)
  if (idx < 0) idx = 0
  const m = matches[idx]
  const text = currentMemo.value!.content
  const next = text.slice(0, m.start) + replaceText.value + text.slice(m.end)
  currentMemo.value!.content = next
  if (!findDialogVisible.value) el.focus() // 查找浮条打开时不抢焦点，避免打断输入法
  const pos = m.start + replaceText.value.length
  el.setSelectionRange(pos, pos)
  nextTick(() => {
    const ms = findTextareaMatches()
    matchCount.value = ms.length
    if (ms.length) {
      matchIndex.value = Math.min(idx, ms.length - 1)
      const nm = ms[matchIndex.value]
      el.setSelectionRange(nm.start, nm.end)
    } else {
      matchIndex.value = 0
    }
  })
}

const replaceAllTextarea = () => {
  const re = findRegex()
  if (!re) return
  const text = currentMemo.value?.content || ''
  const count = (text.match(re) || []).length
  if (!count) return
  re.lastIndex = 0
  currentMemo.value!.content = text.replace(re, () => replaceText.value)
  matchCount.value = 0
  matchIndex.value = 0
  ElMessage.success(`已替换 ${count} 处`)
}

watch(findText, () => {
  if (!findDialogVisible.value) return
  matchIndex.value = 0
  lastAbsStart = -1
  if (findText.value) {
    nextTick(() => { findNext() })
  } else {
    matchCount.value = 0
    clearFindSelection()
  }
})

// 面板宽度调节
const leftPanelWidth = ref(220)
const rightPanelWidth = ref(420)
const explorerCollapsed = ref(false)

/* ============ 侧栏拖拽调宽 ============ */
type ResizeTarget = 'left' | 'right'
interface ResizeCtx {
  target: ResizeTarget
  edge: 'left' | 'right'
  startX: number
  startWidth: number
  min: number
  max: number
}
const resizing = ref<ResizeCtx | null>(null)

function onResizeStart(target: ResizeTarget, e: MouseEvent) {
  const base: Record<ResizeTarget, number> = {
    left: leftPanelWidth.value,
    right: rightPanelWidth.value,
  }
  const range: Record<ResizeTarget, [number, number]> = {
    left: [160, 520],
    right: [260, 760],
  }
  resizing.value = {
    target,
    edge: target === 'left' ? 'right' : 'left',
    startX: e.clientX,
    startWidth: base[target],
    min: range[target][0],
    max: range[target][1],
  }
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
  window.addEventListener('mousemove', onResizeMove)
  window.addEventListener('mouseup', onResizeEnd)
}
function onResizeMove(e: MouseEvent) {
  const r = resizing.value
  if (!r) return
  const delta = r.edge === 'right' ? (e.clientX - r.startX) : (r.startX - e.clientX)
  const w = Math.min(r.max, Math.max(r.min, r.startWidth + delta))
  if (r.target === 'left') {
    leftPanelWidth.value = w
    if (explorerCollapsed.value) explorerCollapsed.value = false
  } else if (r.target === 'right') {
    rightPanelWidth.value = w
  }
}
function onResizeEnd() {
  resizing.value = null
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
  window.removeEventListener('mousemove', onResizeMove)
  window.removeEventListener('mouseup', onResizeEnd)
}

/* ============ 标签工作区（Tab Workspace） ============ */
interface TabItem {
  key: string
  kind: 'chapter' | 'memo' | 'blank'
  id: number
}
const openTabs = ref<TabItem[]>([])
const activeTabKey = ref('')
const dirtyTabs = ref(new Set<string>())
const lastSyncedContent = ref('')

const tabKeyOf = (kind: 'chapter' | 'memo', id: number) => `${kind}:${id}`
const tabObject = (tab: TabItem): Chapter | Memo | null => {
  if (tab.kind === 'chapter') return chaptersList.value.find(c => c.id === tab.id) ?? null
  if (tab.kind === 'memo') return memos.value.find(m => m.id === tab.id) ?? null
  return null
}
const tabTitle = (tab: TabItem) => {
  if (tab.kind === 'blank') return '新标签'
  const obj = tabObject(tab)
  return obj ? obj.title : tab.key
}

/* 新建标签页：只创建一个空标签，不创建章节。
   未选中任何章节时，该标签保持为空；在左侧选中章节后才会显示内容。 */
let blankTabSeq = 0
function createBlankTab() {
  const key = `blank:${++blankTabSeq}`
  openTabs.value.push({ key, kind: 'blank', id: 0 })
  activeTabKey.value = key
  currentChapter.value = null
  currentMemo.value = null
  lastSyncedContent.value = ''
  dirtyTabs.value.delete(key)
}

function ensureTab(kind: 'chapter' | 'memo', id: number) {
  const key = tabKeyOf(kind, id)
  if (!openTabs.value.some(t => t.key === key)) {
    openTabs.value.push({ key, kind, id })
  }
  activeTabKey.value = key
  dirtyTabs.value.delete(key)
}

/* 选中章节/备忘录：若当前激活的是空标签，则用所选内容填充该标签；
   否则按原有逻辑打开（或激活）对应的标签。 */
function openDocumentInTab(kind: 'chapter' | 'memo', id: number) {
  const key = tabKeyOf(kind, id)
  const activeIdx = openTabs.value.findIndex(t => t.key === activeTabKey.value)
  const activeIsBlank = activeIdx >= 0 && openTabs.value[activeIdx]?.kind === 'blank'
  if (activeIsBlank) {
    if (openTabs.value.some(t => t.key === key)) {
      openTabs.value.splice(activeIdx, 1) // 该内容已在其他标签打开，移除空标签并切换过去
    } else {
      openTabs.value[activeIdx] = { key, kind, id }
    }
  } else {
    ensureTab(kind, id)
    return
  }
  activeTabKey.value = key
  dirtyTabs.value.delete(key)
}

function activateTab(tab: TabItem) {
  if (tab.kind === 'blank') {
    currentChapter.value = null
    currentMemo.value = null
    lastSyncedContent.value = ''
    activeTabKey.value = tab.key
    return
  }
  const obj = tabObject(tab)
  if (!obj) return
  if (tab.kind === 'chapter') selectChapter(obj as Chapter)
  else selectMemo(obj as Memo)
}

const closeConfirmVisible = ref(false)
const pendingCloseTab = ref<TabItem | null>(null)

function requestCloseTab(tab: TabItem) {
  if (!dirtyTabs.value.has(tab.key)) {
    closeTab(tab)
    return
  }
  pendingCloseTab.value = tab
  closeConfirmVisible.value = true
}

function closeTab(tab: TabItem) {
  const idx = openTabs.value.findIndex(t => t.key === tab.key)
  if (idx < 0) return
  openTabs.value.splice(idx, 1)
  dirtyTabs.value.delete(tab.key)
  if (activeTabKey.value === tab.key) {
    const next = openTabs.value[Math.max(0, idx - 1)] || openTabs.value[0]
    if (next) {
      activateTab(next)
    } else {
      currentChapter.value = null
      currentMemo.value = null
      activeTabKey.value = ''
    }
  }
}

function confirmDiscardTab() {
  const tab = pendingCloseTab.value
  closeConfirmVisible.value = false
  pendingCloseTab.value = null
  if (tab) closeTab(tab)
}
function confirmSaveAndCloseTab() {
  const tab = pendingCloseTab.value
  closeConfirmVisible.value = false
  pendingCloseTab.value = null
  if (!tab) return
  if (tab.kind === 'chapter') {
    const c = tabObject(tab) as Chapter | null
    if (c) {
      saveChapterNow(c).then(() => closeTab(tab))
    }
  } else {
    const m = tabObject(tab) as Memo | null
    if (m) {
      saveMemoNow(m).then(() => closeTab(tab))
    }
  }
}

async function handleTabCommand(cmd: string, tab: TabItem) {
  switch (cmd) {
    case 'save':
      if (tab.kind === 'chapter') {
        const c = tabObject(tab) as Chapter | null
        if (c) await saveChapterNow(c)
      } else {
        const m = tabObject(tab) as Memo | null
        if (m) await saveMemoNow(m)
      }
      dirtyTabs.value.delete(tab.key)
      break
    case 'saveAll':
      await saveAllTabs()
      break
    case 'close':
      requestCloseTab(tab)
      break
    case 'closeOthers':
      openTabs.value.filter(t => t.key !== tab.key).forEach(t => requestCloseTab(t))
      break
    case 'closeRight': {
      const idx = openTabs.value.findIndex(t => t.key === tab.key)
      openTabs.value.slice(idx + 1).forEach(t => requestCloseTab(t))
      break
    }
  }
}
async function saveAllTabs() {
  for (const tab of openTabs.value) {
    if (!dirtyTabs.value.has(tab.key)) continue
    if (tab.kind === 'chapter') {
      const c = tabObject(tab) as Chapter | null
      if (c) await saveChapterNow(c)
    } else {
      const m = tabObject(tab) as Memo | null
      if (m) await saveMemoNow(m)
    }
    dirtyTabs.value.delete(tab.key)
  }
}

/* 脏标记追踪：编辑器通过 v-model 更新内容（无 input 事件），
   用「激活标签时的内容快照」对比判断是否改动。
   快照在 selectChapter / selectMemo / 新建时同步设置（见下方），
   不依赖 watcher 顺序，因此切换标签不会误标脏。 */
watch(
  () => currentChapter.value?.content,
  (v) => {
    if (v == null) return
    if (activeTabKey.value && v !== lastSyncedContent.value) {
      dirtyTabs.value.add(activeTabKey.value)
      lastSyncedContent.value = v
      scheduleAutoSave()
    }
  }
)
watch(
  () => currentMemo.value?.content,
  (v) => {
    if (v == null) return
    if (activeTabKey.value && v !== lastSyncedContent.value) {
      dirtyTabs.value.add(activeTabKey.value)
      lastSyncedContent.value = v
      scheduleAutoSave()
    }
  }
)

/* 自动保存（3 秒防抖） */
let autoSaveTimer: ReturnType<typeof setTimeout> | null = null
function scheduleAutoSave() {
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
  autoSaveTimer = setTimeout(() => {
    const key = activeTabKey.value
    if (!key || !dirtyTabs.value.has(key)) return
    if (currentChapter.value) {
      saveChapter().then(() => dirtyTabs.value.delete(key))
    } else if (currentMemo.value) {
      saveMemo().then(() => dirtyTabs.value.delete(key))
    }
  }, 3000)
}

/* 快捷键：⌘S 保存 / ⌘W 关闭 / ⌘F 查找 / ⌘H 替换 / Esc 关闭浮层 */
function handleGlobalKeydown(e: KeyboardEvent) {
  // 输入法组合期间的按键交给输入法处理（如 Esc 取消候选词），不触发全局快捷键
  if (e.isComposing) return
  const mod = e.metaKey || e.ctrlKey
  if (mod && e.key.toLowerCase() === 'f') { e.preventDefault(); openFindDialog('find') }
  else if (mod && e.key.toLowerCase() === 'h') { e.preventDefault(); openFindDialog('replace') }
  else if (mod && e.key.toLowerCase() === 's') { e.preventDefault(); if (currentChapter.value) saveChapter(); else if (currentMemo.value) saveMemo() }
  else if (mod && e.key.toLowerCase() === 'w') {
    e.preventDefault()
    const tab = openTabs.value.find(t => t.key === activeTabKey.value)
    if (tab) requestCloseTab(tab)
  }
  else if (e.key === 'Escape' && findDialogVisible.value) {
    closeFindDialog()
  }
}

const chapterEditorRef = ref()
const memoEditorRef = ref()
const knowledgeGraphRef = ref()
const editorSheetRef = ref<HTMLElement | null>(null)
const pinChapterTitle = ref(false)

// ===== AI 工具栏（核心固定 + 收藏置顶，收藏持久化到 localStorage） =====
const aiToolList = AI_TOOLS.filter(t => !t.hidden)
const CORE_TOOL_IDS = ['ai-continue', 'ai-polish', 'ai-review']
const PIN_STORAGE_KEY = 'xm-ai-toolbar-pins'
const PIN_LIMIT = 5
const coreTools = computed(() => aiToolList.filter(t => CORE_TOOL_IDS.includes(t.id)))

/* 收藏池：全部工具（含隐藏工具）+ 非工具功能（敏感词检测），收藏与「＋」菜单共用 */
interface AiFavoriteItem { id: string; label: string; open: () => void }
const aiFavoriteItems = computed<AiFavoriteItem[]>(() => {
  const items: AiFavoriteItem[] = aiToolList
    .filter(t => !CORE_TOOL_IDS.includes(t.id))
    .map(t => ({ id: t.id, label: t.label, open: () => onAiToolClick(t.id) }))
  // 隐藏工具（文风转换 / 对话生成）：走全章素材的面板入口
  for (const t of AI_TOOLS.filter(t => t.hidden)) {
    items.push({ id: t.id, label: t.label, open: () => openHiddenAiTool(t.id) })
  }
  items.push({ id: 'ai-sensitive', label: '敏感词检测', open: () => { sensitiveVisible.value = true } })
  return items
})
const isFavoriteId = (id: string) => aiFavoriteItems.value.some(i => i.id === id)
const loadPinnedIds = (): string[] => {
  try {
    const list = JSON.parse(localStorage.getItem(PIN_STORAGE_KEY) || '[]')
    return Array.isArray(list)
      ? list.filter((id): id is string => typeof id === 'string').slice(0, PIN_LIMIT)
      : []
  } catch {
    return []
  }
}
const aiPinnedIds = ref<string[]>(loadPinnedIds())
// 过滤掉已不存在的收藏 id（工具调整后仍安全）
const pinnedItems = computed(() => aiPinnedIds.value.map(id => aiFavoriteItems.value.find(i => i.id === id)).filter((i): i is AiFavoriteItem => !!i))
const toggleAiPin = (id: string) => {
  if (!isFavoriteId(id)) return
  if (aiPinnedIds.value.includes(id)) {
    aiPinnedIds.value = aiPinnedIds.value.filter(x => x !== id)
  } else {
    if (aiPinnedIds.value.length >= PIN_LIMIT) {
      ElMessage.info(`收藏上限 ${PIN_LIMIT} 个，先取消一个再收藏`)
      return
    }
    aiPinnedIds.value = [...aiPinnedIds.value, id]
  }
  try {
    localStorage.setItem(PIN_STORAGE_KEY, JSON.stringify(aiPinnedIds.value))
  } catch { /* 容量满则忽略，仅本次会话生效 */ }
}
const aiSidePanelRef = ref<InstanceType<typeof AiSidePanel> | null>(null)
const continueDialogVisible = ref(false)
const continueTool = ref<AiToolMeta | null>(null)
// AI审稿 / AI去痕：生成器式弹窗（左模板 + 右表单）
const reviewDialogVisible = ref(false)
const reviewTool = ref<AiToolMeta | null>(null)
const reviewSourceText = ref('')
const reviewSourceTitle = ref('')
const charLibraryDialogVisible = ref(false)
const termLibraryDialogVisible = ref(false)

// ===== 其余编辑器功能（高频词透镜 / 听书 / 一键生成概要） =====
const wordFrequencyVisible = ref(false)

const onWordFreqLocate = (word: string) => {
  findText.value = word
  openFindDialog('find')
}

// ===== 插画生成 / 敏感词检测 =====
const illustrationVisible = ref(false)
const illustrationPrompt = ref('')
const sensitiveVisible = ref(false)

const openIllustration = (useSelection: boolean) => {
  if (!currentChapter.value) {
    ElMessage.warning('请先打开一个章节')
    return
  }
  if (useSelection) {
    const sel = currentSelectionText.value.trim() || chapterEditorRef.value?.getSelection?.() || ''
    if (!sel.trim()) {
      ElMessage.warning('请先选中需要配图的文字')
      return
    }
    illustrationPrompt.value = `小说场景插画：${sel.trim().slice(0, 300)}`
  } else {
    const plain = getChapterPlainText()
      .replace(/^[#>\-*\s]+/gm, '')
      .replace(/\*/g, '')
      .trim()
    illustrationPrompt.value = plain
      ? `小说场景插画，取材自本章：${plain.slice(0, 300)}`
      : ''
  }
  illustrationVisible.value = true
}

const onIllustrationInsert = (markdown: string) => {
  const ed = chapterEditorRef.value
  if (!ed) return
  ed.insertPlainText(markdown)
  nextTick(() => { saveChapter() })
  ElMessage.success('插画已插入正文')
}

const onSensitiveLocate = (word: string) => {
  findText.value = word
  openFindDialog('find')
}

const openHiddenAiTool = (id: string) => {
  const tool = getAiTool(id)
  if (!tool) return
  if (!currentChapter.value) {
    ElMessage.warning('请先打开一个章节')
    return
  }
  openSidePanel(tool, {
    sourceText: getChapterPlainText(),
    sourceTitle: currentChapter.value.title
  })
}

// 听书（Web Speech TTS）
const ttsActive = ref(false)
const toggleTts = () => {
  const synth = window.speechSynthesis
  if (!synth) {
    ElMessage.error('当前环境不支持语音朗读')
    return
  }
  if (ttsActive.value) {
    synth.cancel()
    ttsActive.value = false
    return
  }
  if (!currentChapter.value) return
  const selected = chapterEditorRef.value?.getSelection?.() || ''
  const text = (selected.trim() || getChapterPlainText()).trim()
  if (!text) {
    ElMessage.info('暂无可朗读的内容')
    return
  }
  const utter = new SpeechSynthesisUtterance(text.slice(0, 3000))
  utter.lang = 'zh-CN'
  utter.rate = 1
  utter.onend = () => { ttsActive.value = false }
  utter.onerror = () => { ttsActive.value = false }
  synth.speak(utter)
  ttsActive.value = true
  ElMessage.info(selected.trim() ? '开始朗读选中内容' : '开始朗读本章（最多 3000 字）')
}

// 一键生成概要：复用 AI 面板管线，结果写回章节概要
const AI_SUMMARY_TOOL: AiToolMeta = {
  id: 'ai-summary',
  label: '一键生成概要',
  title: '一键生成概要',
  shell: 'side',
  requirementLabel: '概要要求',
  requirementPlaceholder: '',
  defaultRequirement: '总结本章的核心剧情、冲突推进与结果，300 字以内。',
  presets: [],
  startLabel: '生成',
  applyMode: 'copy',
  systemPrompt: '你是一位网文编辑。任务：阅读给定章节内容，输出本章概要：核心事件、冲突推进、出场角色、章末钩子，300 字以内。只输出概要本身。'
}

const onChapterGenerateSummary = (_chapter: Chapter) => {
  if (!currentChapter.value) return
  openSidePanel(AI_SUMMARY_TOOL, {
    sourceText: getChapterPlainText(),
    sourceTitle: currentChapter.value.title,
    autoStart: true
  })
}

// 选区跟踪（选区悬浮工具条）
const selectionBarVisible = ref(false)
const selectionBarX = ref(0)
const selectionBarY = ref(0)
const selectionCount = ref(0)
const currentSelectionText = ref('')

const getChapterPlainText = () => (currentChapter.value?.content || '')
  .replace(/<[^>]*>/g, '')
  .replace(/```[\s\S]*?```/g, '')
  .trim()

const handleEditorSelect = () => {
  const edSel = chapterEditorRef.value?.getSelection?.() || ''
  const winSel = window.getSelection()
  const winText = winSel?.toString() || ''
  const text = (edSel || winText).trim()
  if (!text) {
    selectionBarVisible.value = false
    return
  }
  currentSelectionText.value = edSel || winText
  selectionCount.value = [...currentSelectionText.value].length
  const rect = winSel && winSel.rangeCount > 0 ? winSel.getRangeAt(0).getBoundingClientRect() : null
  if (rect && rect.width > 0) {
    // 保存选区：悬浮条上的「标题」下拉菜单点击会转移焦点，执行格式化前需还原选区
    savedSelectionRange = winSel!.getRangeAt(0).cloneRange()
    selectionBarX.value = Math.min(Math.max(rect.x, 8), window.innerWidth - 740)
    selectionBarY.value = Math.max(rect.y - 44, 8)
    selectionBarVisible.value = true
  }
}

// 悬浮工具条格式化：优先用编辑器当前选区，丢失时还原打开工具条时保存的选区
let savedSelectionRange: Range | null = null
const ensureEditorSelection = () => {
  const sel = window.getSelection()
  if (sel && sel.rangeCount > 0 && !sel.isCollapsed) return
  if (savedSelectionRange) {
    sel?.removeAllRanges()
    sel?.addRange(savedSelectionRange)
  }
}

const onSelectionFormat = (kind: string) => {
  const root = getEditableRoot()
  if (!root) return
  root.focus()
  ensureEditorSelection()
  if (kind === 'inlineCode') { insertInlineCode(); return }
  document.execCommand(kind)
  nextTick(() => { saveChapter() })
}

const onSelectionFormatBlock = (tag: string) => {
  const root = getEditableRoot()
  if (!root) return
  root.focus()
  ensureEditorSelection()
  formatBlock(tag)
}

const hideSelectionBar = () => { selectionBarVisible.value = false }

const applySelectionFormat = (kind: 'bold' | 'highlight' | 'clear') => {
  const root = getEditableRoot()
  if (!root) return
  root.focus()
  if (kind === 'bold') document.execCommand('bold')
  else if (kind === 'highlight') document.execCommand('hiliteColor', false, '#fff176')
  else document.execCommand('removeFormat')
  nextTick(() => { saveChapter() })
}

const openSidePanel = (tool: AiToolMeta, opts: {
  sourceText?: string
  sourceTitle?: string
  chaptersContext?: string
  requirement?: string
  extra?: string
  configId?: number
  autoStart?: boolean
  multi?: Array<{ label: string; userMessage: string }>
  presetResult?: string
}) => {
  activeRightPanel.value = 'ai-tool'
  hideSelectionBar()
  nextTick(() => {
    aiSidePanelRef.value?.open({ tool, ...opts })
  })
}

const onAiToolClick = (id: string) => {
  const tool = getAiTool(id)
  if (!tool) return
  if (!currentChapter.value) {
    ElMessage.warning('请先打开一个章节')
    return
  }
  if (tool.shell === 'continue-dialog') {
    continueTool.value = tool
    continueDialogVisible.value = true
    return
  }
  if (tool.shell === 'review-dialog') {
    if (!currentChapter.value) {
      ElMessage.warning('请先打开一个章节')
      return
    }
    reviewTool.value = tool
    // 优先用选中文本作为素材，否则用本章全文（无字数上限）
    const sel = currentSelectionText.value.trim()
    reviewSourceText.value = sel || getChapterPlainText()
    reviewSourceTitle.value = sel ? '选中文本' : currentChapter.value.title
    reviewDialogVisible.value = true
    return
  }
  if (tool.requireSelection && !currentSelectionText.value.trim()) {
    ElMessage.warning(tool.selectionHint || '请先选中文本')
    return
  }
  openSidePanel(tool, {
    sourceText: tool.requireSelection ? currentSelectionText.value : getChapterPlainText(),
    sourceTitle: currentChapter.value.title
  })
}

// 选区悬浮工具条点击
const onSelectionTool = (rawId: string) => {
  selectionBarVisible.value = false
  if (rawId === 'ai-illustration') { openIllustration(true); return }
  onAiToolClick(rawId)
}

// 续写/章纲弹窗 → 开始生成
const onContinueStart = (payload: {
  requirement: string
  extra: string
  contextChapterIds: number[]
  memoIds: number[]
  characterNames: string[]
  termNames: string[]
}) => {
  const tool = continueTool.value
  if (!tool || !currentChapter.value) return
  const ctxParts: string[] = []
  const ctxChapters = payload.contextChapterIds
    .map(id => sortedChapters.value.find(c => c.id === id))
    .filter((c): c is Chapter => !!c)
  if (ctxChapters.length) {
    ctxParts.push(`【前文参考】\n${ctxChapters.map(c => `【章节：${c.title}】\n${(c.content || '').replace(/<[^>]*>/g, '').slice(0, 3000)}`).join('\n\n')}`)
  }
  const ctxMemos = payload.memoIds
    .map(id => memos.value.find(m => m.id === id))
    .filter((m): m is Memo => !!m)
  if (ctxMemos.length) {
    ctxParts.push(`【备忘录参考】\n${ctxMemos.map(m => `【${m.title}】\n${(m.content || '').slice(0, 2000)}`).join('\n\n')}`)
  }
  if (payload.characterNames.length) ctxParts.push(`【提及角色】\n${payload.characterNames.join('、')}`)
  if (payload.termNames.length) ctxParts.push(`【提及词条】\n${payload.termNames.join('、')}`)
  openSidePanel(tool, {
    sourceText: getChapterPlainText(),
    sourceTitle: currentChapter.value.title,
    chaptersContext: ctxParts.join('\n\n'),
    requirement: payload.requirement,
    extra: payload.extra,
    autoStart: true
  })
}

// 面板结果应用回编辑器
const onAiSideApply = ({ applyMode, text }: { applyMode: string; text: string }) => {
  const ed = chapterEditorRef.value
  if (!ed || !currentChapter.value) return
  if (applyMode === 'replace') {
    const sel = ed.getSelection?.() || ''
    if (sel.trim()) ed.insertContent(text)
    else ed.insertPlainText(text)
  } else if (applyMode === 'append') {
    const cur = ed.getValue() || ''
    ed.setValue(cur.replace(/\s*$/, '') + '\n\n' + text)
  } else if (applyMode === 'insert') {
    ed.insertPlainText(text)
  } else {
    return
  }
  nextTick(() => { saveChapter() })
  ElMessage.success('已应用到正文')
}

// 面板结果导出为新章节
const onAiSideExport = async ({ title, text }: { title: string; text: string }) => {
  if (!bookId) return
  try {
    const res = await chapterAPI.create({ book_id: bookId, title, content: text })
    if (res.success) {
      await fetchChapters()
      ElMessage.success(`已导出为新章节「${title}」`)
    }
  } catch {
    ElMessage.error('导出失败')
  }
}

// 每个工具最近一次生成结果（续写弹窗"查看上次结果"）
const aiLastResults = new Map<string, string>()
const onAiSideFinish = ({ toolId, text }: { toolId: string; text: string }) => {
  aiLastResults.set(toolId, text)
  // 一键生成概要：询问是否写回章节概要
  if (toolId === 'ai-summary' && currentChapter.value) {
    const chapter = currentChapter.value
    ElMessageBox.confirm('是否将本次结果保存为本章概要？', '一键生成概要', {
      confirmButtonText: '保存概要',
      cancelButtonText: '不保存',
      type: 'info'
    }).then(async () => {
      try {
        const res = await chapterAPI.update(chapter.id, {
          ...buildChapterUpdate(chapter),
          summary: text
        })
        if (res.success && res.data) {
          syncChapterState(res.data)
          ElMessage.success('概要已保存')
        }
      } catch {
        ElMessage.error('概要保存失败')
      }
    }).catch(() => {})
  }
}
const showAiLastResult = () => {
  const tool = continueTool.value
  if (!tool) return
  const last = aiLastResults.get(tool.id)
  if (!last) {
    ElMessage.info('暂无上次结果')
    return
  }
  openSidePanel(tool, { presetResult: last })
}

// 点击编辑器和悬浮条以外的区域时隐藏悬浮条
const onDocMouseDown = (e: MouseEvent) => {
  const t = e.target as HTMLElement
  if (!t || !t.closest) return
  if (t.closest('.xm-sel-toolbar') || t.closest('.editor-sheet')) return
  selectionBarVisible.value = false
}
onMounted(() => {
  document.addEventListener('mousedown', onDocMouseDown)
  // 滚动不冒泡但走捕获阶段，document 级捕获可覆盖条件渲染的编辑器
  document.addEventListener('scroll', onEditorScrollCapture, true)
})
onUnmounted(() => {
  document.removeEventListener('mousedown', onDocMouseDown)
  document.removeEventListener('scroll', onEditorScrollCapture, true)
  window.speechSynthesis?.cancel()
})

// 编辑器内容滚动容器（wysiwyg / ir / reset 依次回退）
const getEditorScrollEl = (): HTMLElement | null => {
  const sheet = editorSheetRef.value
  if (!sheet) return null
  const candidates = Array.from(
    sheet.querySelectorAll<HTMLElement>('.vditor-wysiwyg, .vditor-ir, .vditor-reset')
  )
  for (const el of candidates) {
    if (el.scrollHeight > el.clientHeight + 4) return el
  }
  return candidates[candidates.length - 1] || null
}

const scrollEditorTo = (pos: 'top' | 'bottom') => {
  const el = getEditorScrollEl()
  if (!el) return
  el.scrollTo({ top: pos === 'top' ? 0 : el.scrollHeight, behavior: 'smooth' })
}

// 编辑器可编辑根节点（wysiwyg 优先）
const getEditableRoot = (): HTMLElement | null => {
  const vd = chapterEditorRef.value?.vditor?.()
  return vd?.vditor?.wysiwyg?.element || vd?.vditor?.ir?.element || null
}

// ===== 合并单行：格式化 / 预览 / 全屏 =====
const formatInline = (cmd: string) => {
  const root = getEditableRoot()
  if (!root) return
  root.focus()
  document.execCommand(cmd)
  nextTick(() => { saveChapter() })
}

const formatBlock = (tag: string) => {
  const root = getEditableRoot()
  if (!root) return
  root.focus()
  const sel = window.getSelection()
  const anchor = sel?.anchorNode
  // 已是该标签则切回正文（toggle）
  const current = anchor ? (anchor.parentElement?.closest('h1, h2, h3')?.tagName?.toLowerCase() || '') : ''
  document.execCommand('formatBlock', false, current === tag ? 'p' : tag)
  nextTick(() => { saveChapter() })
}

const insertInlineCode = () => {
  const root = getEditableRoot()
  if (!root) return
  root.focus()
  const selected = window.getSelection()?.toString() || ''
  document.execCommand('insertHTML', false, `<code>${selected || '代码'}</code>`)
  nextTick(() => { saveChapter() })
}

const toggleEditorFullscreen = () => {
  const sheet = editorSheetRef.value
  if (!sheet) return
  if (document.fullscreenElement) {
    document.exitFullscreen()
  } else {
    sheet.requestFullscreen()
  }
}

const previewMode = ref(false)
const previewHtml = ref('')
// 固定章节标题：编辑器滚动后浮现迷你标题条
const titleChipVisible = ref(false)
const onEditorScrollCapture = () => {
  if (!pinChapterTitle.value) {
    titleChipVisible.value = false
    return
  }
  const el = getEditorScrollEl()
  titleChipVisible.value = !!el && el.scrollTop > 80
}
const togglePreview = () => {
  if (previewMode.value) {
    previewMode.value = false
    return
  }
  previewHtml.value = chapterEditorRef.value?.getHtml?.() || ''
  previewMode.value = true
}

const editorUndo = () => {
  const root = getEditableRoot()
  if (!root) return
  root.focus()
  document.execCommand('undo')
}

const editorRedo = () => {
  const root = getEditableRoot()
  if (!root) return
  root.focus()
  document.execCommand('redo')
}

const editorCopy = async () => {
  const ed = chapterEditorRef.value
  const selected = ed?.getSelection?.() || ''
  const text = selected || (ed?.getValue?.() || '').replace(/<[^>]*>/g, '')
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(selected ? '已复制选中内容' : '已复制全部内容')
  } catch {
    ElMessage.error('复制失败')
  }
}

const editorSelectAll = () => {
  const root = getEditableRoot()
  if (!root) return
  root.focus()
  document.execCommand('selectAll')
}

// 整篇替换（走编辑器事务，Ctrl+Z/撤回 可撤销）
const replaceWholeEditorText = (newText: string): boolean => {
  const root = getEditableRoot()
  if (!root || !newText) return false
  root.focus()
  document.execCommand('selectAll')
  document.execCommand('insertText', false, newText)
  return true
}

// ===== 统一中文标点（对齐星月：静默转换 + 可撤回） =====
const applyPunctuation = () => {
  const ed = chapterEditorRef.value
  if (!ed || !currentChapter.value) return
  const src = ed.getValue() || ''
  if (!src.trim()) {
    ElMessage.info('当前章节暂无内容')
    return
  }
  let count = 0
  const map: Record<string, string> = { ',': '，', ':': '：', ';': '；', '?': '？', '!': '！', '(': '（', ')': '）' }
  let quoteOpen = false
  const converted = src.split('\n').map((line: string) => {
    // 跳过 Markdown 结构行（标题/列表/引用/代码），避免破坏格式
    if (/^\s*(#{1,6}\s|[-*+]\s|>\s|\d+\.\s|```)/.test(line)) return line
    return line.replace(/[,:;?!()."']|(\.\.+)/g, (m: string, dots: string | undefined, offset: number, full: string) => {
      if (dots) { count++; return '……' }
      // 不转换数字间的小数点
      if (m === '.') {
        const prev = full[offset - 1] || ''
        const next = full[offset + 1] || ''
        if (/\d/.test(prev) && /\d/.test(next)) return m
        count++
        return '。'
      }
      if (m === '"' || m === "'") {
        count++
        const close = quoteOpen
        if (m === '"') quoteOpen = !quoteOpen
        return m === '"' ? (close ? '”' : '“') : (close ? '’' : '‘')
      }
      if (map[m]) { count++; return map[m] }
      return m
    })
  }).join('\n')
  if (count === 0) {
    ElMessage.success('已统一全文中文标点，共调整 0 处')
    return
  }
  if (replaceWholeEditorText(converted)) {
    nextTick(() => { saveChapter() })
    ElMessage.success(`已统一全文中文标点，共调整 ${count} 处，可撤回`)
  }
}

// ===== 智能排版（对齐星月：段首缩进 + 段间规范化） =====
const applySmartLayout = () => {
  const ed = chapterEditorRef.value
  if (!ed || !currentChapter.value) return
  const src = ed.getValue() || ''
  if (!src.trim()) {
    ElMessage.info('当前章节暂无内容')
    return
  }
  const paras = src
    .split(/\r?\n/)
    .map((l: string) => l.trim())
    .filter((l: string) => l.length > 0)
    .map((l: string) => (/^\s*　{2}/.test(l) ? l : '　　' + l))
  const converted = paras.join('\n\n')
  if (replaceWholeEditorText(converted)) {
    nextTick(() => { saveChapter() })
    ElMessage.success('智能排版完成')
  }
}

// ===== 手机端阅读预览（对齐星月） =====
const phonePreviewVisible = ref(false)
const getEditorScrollRatio = (): number | null => {
  const el = getEditorScrollEl()
  if (!el) return null
  const max = el.scrollHeight - el.clientHeight
  return max > 0 ? el.scrollTop / max : 0
}
const scrollEditorToRatio = (ratio: number) => {
  const el = getEditorScrollEl()
  if (!el) return
  el.scrollTop = ratio * (el.scrollHeight - el.clientHeight)
}
// 手机预览里的上一章/下一章
const onPhoneNav = (dir: number) => {
  const idx = sortedChapters.value.findIndex(c => c.id === currentChapter.value?.id)
  const target = sortedChapters.value[idx + dir]
  if (target) {
    selectChapter(target)
    ElMessage.success(dir < 0 ? '已切换到上一章' : '已切换到下一章')
  }
}

// 目录刷新（对齐星月「刷新列表」）
const refreshCatalog = async () => {
  await Promise.all([fetchChapters(), fetchVolumes()])
  ElMessage.success('列表已刷新')
}

// 关联内容
const showRelateDialog = ref(false)
const conversationListVisible = ref(false)

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
  // 加载写作设置（段落/横线/标记/主题）
  loadWriteSettings()
  initThemeMode()

  // 标签工作区快捷键
  window.addEventListener('keydown', handleGlobalKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalKeydown)
  window.removeEventListener('mousemove', onResizeMove)
  window.removeEventListener('mouseup', onResizeEnd)
  systemThemeQuery?.removeEventListener('change', onSystemThemeChange)
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
    ...buildChapterUpdate(chapter),
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

// 历史记录功能已迁移到 HistoryCenter 组件

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

/* 右侧单栏切换：点胶囊开对应面板，再点一次收起 */
const toggleRightPanel = async (key: RightPanelKey) => {
  if (activeRightPanel.value === key) {
    activeRightPanel.value = null
    return
  }
  showGraphPanel.value = false
  activeRightPanel.value = key
  if (key === 'chat') {
    if (!currentConversation.value && conversations.value.length === 0) {
      await createConversation()
    } else if (!currentConversation.value && conversations.value.length > 0) {
      await selectConversation(conversations.value[0])
    }
  }
}

const closeRightPanel = () => {
  activeRightPanel.value = null
}

// AI 写作面板「应用到当前章节」回调
const handleApplyCreative2Result = (content: string) => {
  if (!currentChapter.value) return
  if (currentChapter.value.content) {
    currentChapter.value.content += '\n\n' + content
  } else {
    currentChapter.value.content = content
  }
  ElMessage.success('已应用到当前章节')
  saveChapter()
}

const toggleGraphPanel = () => {
  showGraphPanel.value = !showGraphPanel.value
  if (showGraphPanel.value) activeRightPanel.value = null
}

const closeGraphPanel = () => {
  showGraphPanel.value = false
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

const handleTextSelect = (event: Event) => {
  const target = event.target as HTMLTextAreaElement
  const selectedText = target.value.substring(target.selectionStart, target.selectionEnd)
  selectedTextLength.value = selectedText.length > 0 ? getContentLength(selectedText) : 0
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

const handleCharacterSelected = (_character: Character) => {
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
  lastSyncedContent.value = chapter.content
  openDocumentInTab('chapter', chapter.id)
  // 记住当前章节位置
  try { localStorage.setItem(`novel_last_chapter_${bookId}`, String(chapter.id)) } catch {}
}

const selectMemo = (memo: Memo) => {
  currentMemo.value = memo
  currentChapter.value = null
  lastSyncedContent.value = memo.content
  openDocumentInTab('memo', memo.id)
  // 记住当前备忘录位置
  try { localStorage.setItem(`novel_last_memo_${bookId}`, String(memo.id)) } catch {}
}

// 全局备忘录弹窗「跳转编辑」回调
const handleJumpToMemo = (memo: Memo) => {
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
    lastSyncedContent.value = res.data.content
    openDocumentInTab('chapter', res.data.id)
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
    ...buildChapterUpdate(chapter),
    volume_id: targetFolderId || undefined
  } as any)

  if (res.success && res.data) {
    syncChapterState(res.data)
    moveChapterDialogVisible.value = false
    chapterToMove.value = null
    ElMessage.success('文件已移动')
  }
}

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

/* 构建章节更新载荷：必须带上 type / volume_id，
   否则后端 UpdateChapter 会把缺失字段覆盖为 '' / NULL（创建时写入的 type 会被清掉）。 */
const buildChapterUpdate = (chapter: Chapter) => ({
  title: chapter.title,
  content: chapter.content,
  summary: chapter.summary || '',
  order_num: chapter.order_num,
  type: chapter.type || 'chapter',
  volume_id: chapter.volume_id ?? undefined
})

const saveChapter = async () => {
  if (currentChapter.value) await saveChapterNow(currentChapter.value)
}

const saveChapterNow = async (chapter: Chapter) => {
  const res = await chapterAPI.update(chapter.id, buildChapterUpdate(chapter))

  if (res.success && res.data) {
    syncChapterState(res.data)
    dirtyTabs.value.delete(tabKeyOf('chapter', chapter.id))
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
      ...buildChapterUpdate(chapter),
      summary
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
  if (currentMemo.value) await saveMemoNow(currentMemo.value)
}

const saveMemoNow = async (memo: Memo) => {
  const res = await memoAPI.update(memo.id, {
    title: memo.title,
    content: memo.content
  })
  if (res?.success !== false) {
    dirtyTabs.value.delete(tabKeyOf('memo', memo.id))
  }
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

const scrollToBottom = () => {
  nextTick(() => {
    if (chatMessagesRef.value) {
      chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight
    }
  })
}


</script>

<style scoped>

.write-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #f0fdf9 0%, #e6fffa 100%);
  overflow: hidden;
  animation: fadeIn 0.4s ease;
}

/* ========== 主内容区域（工作台：圆角卡片悬浮于灰底） ========== */
.content-wrapper {
  flex: 1;
  display: flex;
  overflow: hidden;
  position: relative;
  padding: 10px;
}

/* ========== 左侧目录面板 - 圆角卡片 ========== */
.left-panel {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  min-height: 0;
  background: #ffffff;
  border: 1px solid #e4e2dd;
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

.left-panel > .el-tabs,
.left-panel :deep(.el-tabs) {
  background: transparent;
  border-radius: 0;
  overflow: hidden;
}

.catalog-header {
  padding: 10px 14px;
  border-bottom: 1px solid #ebedf0;
  background: transparent;
  border-radius: 0;
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
  background: #161b22;
  border-color: rgba(42, 50, 66, 0.4);
  box-shadow: none;
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
  background: #1f2633 !important;
}

.catalog-list::-webkit-scrollbar-thumb {
  background: #c0c4cc;
  border-radius: 3px;
}

:root[data-theme='dark'] .catalog-list::-webkit-scrollbar-thumb {
  background: #2d3646 !important;
}

:root[data-theme='dark'] .catalog-list::-webkit-scrollbar-thumb:hover {
  background: #3f4a5c !important;
}

/* ========== 中间编辑器面板（圆角卡片） ========== */
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
  padding: 0;
  animation: slideIn 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.editor-sheet {
  position: relative;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  border: 1px solid #e4e2dd;
  border-radius: 14px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
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

/* ── 星月对齐：AI 工具栏 / 图标工具行 / 悬浮导航 ── */
.xm-ai-toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  padding: 10px 16px 0;
  flex-shrink: 0;
}

.xm-ai-toolbar .xm-ai-btn {
  border-radius: 6px;
  font-weight: 500;
}

/* 收藏置顶：钉在工具栏上的工具（右上角 ✕ 取消收藏） */
.xm-ai-toolbar .xm-ai-btn--pinned {
  position: relative;
  padding-right: 22px !important;
}

.xm-ai-unpin {
  position: absolute;
  right: 6px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 10px;
  line-height: 1;
  color: var(--xm-ink-3, #9d9c96);
  cursor: pointer;
}

.xm-ai-unpin:hover {
  color: #e5484d;
}

/* 收藏入口按钮（虚线框的 ＋） */
.xm-ai-toolbar .xm-ai-btn--add {
  border-style: dashed !important;
  font-weight: 600;
  padding-left: 12px !important;
  padding-right: 12px !important;
}

/* 收藏菜单右侧的状态文案 */
.xm-pin-state {
  margin-left: 14px;
  font-size: 11px;
  color: var(--xm-ink-3, #9d9c96);
}

.xm-pin-state.is-on {
  color: var(--xm-accent, #2563eb);
}

.xm-icon-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 4px;
  padding: 6px 12px 0;
  flex-shrink: 0;
}

.xm-icon-row-left {
  display: flex;
  align-items: center;
  gap: 1px;
  flex-wrap: nowrap;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: none;
  min-width: 0;
}

.xm-icon-row-left::-webkit-scrollbar {
  display: none;
}

.xm-icon-sep {
  width: 1px;
  height: 14px;
  background: var(--xm-line, #e4e2dd);
  margin: 0 3px;
  flex-shrink: 0;
}

.xm-icon-btn--glyph {
  font-size: 13px;
  font-weight: 700;
  line-height: 1;
}

.xm-icon-btn--italic {
  font-style: italic;
  font-family: Georgia, serif;
}

.xm-icon-btn--strike {
  text-decoration: line-through;
}

.xm-icon-caret {
  font-size: 10px;
  margin-left: 1px;
}

.xm-preview-overlay {
  position: absolute;
  inset: 0;
  z-index: 8;
  background: var(--xm-surface, #fdfcf9);
  display: flex;
  flex-direction: column;
}

.xm-preview-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  font-size: 12.5px;
  color: var(--xm-ink-3, #8a8f99);
  border-bottom: 1px solid var(--xm-line, #eeeeea);
  flex-shrink: 0;
}

.xm-preview-close {
  border: 1px solid var(--xm-line-strong, #d6d2c8);
  background: var(--xm-surface, #ffffff);
  border-radius: 6px;
  font-size: 12px;
  color: var(--xm-ink-2, #5a6270);
  padding: 3px 10px;
  cursor: pointer;
  transition: color var(--xm-dur, 0.2s) var(--xm-ease, ease), border-color var(--xm-dur, 0.2s) var(--xm-ease, ease),
    background var(--xm-dur, 0.2s) var(--xm-ease, ease);
}

.xm-preview-close:hover {
  border-color: var(--xm-accent, #2563eb);
  color: var(--xm-accent, #2563eb);
  background: var(--xm-accent-soft, rgba(37, 99, 235, 0.1));
}

.xm-preview-body {
  flex: 1;
  overflow: auto;
  padding: 28px 56px;
  font-size: 15px;
  line-height: 2;
  color: var(--xm-ink, #2c3138);
}

.xm-preview-body h1,
.xm-preview-body h2,
.xm-preview-body h3 {
  margin: 1.2em 0 0.6em;
  color: var(--xm-ink, #1f2329);
}

.xm-preview-body pre {
  background: var(--xm-surface-2, #f3f4f6);
  border-radius: 8px;
  padding: 12px;
  overflow: auto;
}

.xm-preview-body code {
  background: var(--xm-surface-2, #f3f4f6);
  border-radius: 4px;
  padding: 1px 5px;
  font-size: 0.9em;
}

.xm-preview-body blockquote {
  margin: 0.8em 0;
  padding: 2px 14px;
  border-left: 3px solid var(--xm-line-strong, #d8dcdf);
  color: var(--xm-ink-2, #6b7280);
}

.xm-title-chip {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 9;
  padding: 6px 16px;
  font-size: 13px;
  font-weight: 600;
  color: var(--xm-ink, #1f2329);
  background: color-mix(in srgb, var(--xm-surface, #ffffff) 92%, transparent);
  backdrop-filter: blur(6px);
  border-bottom: 1px solid var(--xm-line, #eeeeea);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.xm-icon-row-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.xm-icon-btn {
  width: 26px;
  height: 26px;
  border: none;
  background: transparent;
  border-radius: 5px;
  color: var(--xm-ink-3, #8a8f99);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.15s var(--xm-ease, ease), color 0.15s var(--xm-ease, ease);
}

.xm-icon-btn:hover {
  background: var(--xm-surface-3, rgba(0, 0, 0, 0.06));
  color: var(--xm-ink, #374151);
}

.xm-icon-btn--active {
  background: var(--xm-accent-soft, rgba(37, 99, 235, 0.12));
  color: var(--xm-accent, #2563eb);
}

.xm-float-nav {
  position: absolute;
  right: 14px;
  bottom: 56px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  z-index: 5;
}

.xm-float-btn {
  width: 32px;
  height: 32px;
  border: 1px solid var(--xm-line, #e4e2dd);
  border-radius: 8px;
  background: var(--xm-surface, #ffffff);
  color: var(--xm-ink-2, #5a6270);
  box-shadow: var(--xm-shadow-pop, 0 2px 6px rgba(0, 0, 0, 0.08));
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s var(--xm-ease, ease);
}

.xm-float-btn:hover {
  color: var(--xm-accent, #0ca678);
  border-color: var(--xm-accent, rgba(0, 201, 167, 0.45));
  background: var(--xm-accent-soft, transparent);
}

.xm-float-btn--listen:hover {
  color: var(--xm-accent, #d97706);
  border-color: var(--xm-accent, rgba(217, 119, 6, 0.45));
}

.xm-float-btn--draw:hover {
  color: var(--xm-accent, #4f7cf7);
  border-color: var(--xm-accent, rgba(79, 124, 247, 0.45));
}

/* AI 工具栏橙色按钮（暗色下去掉白底） */
:root[data-theme='dark'] .xm-ai-toolbar .el-button--warning.is-plain {
  --el-button-bg-color: rgba(217, 119, 6, 0.14);
  --el-button-border-color: rgba(217, 119, 6, 0.45);
  --el-button-text-color: #fbbf24;
  --el-button-hover-bg-color: rgba(217, 119, 6, 0.24);
  --el-button-hover-border-color: rgba(217, 119, 6, 0.6);
  --el-button-hover-text-color: #fcd34d;
  --el-button-active-bg-color: rgba(217, 119, 6, 0.3);
  background: var(--el-button-bg-color);
  border-color: var(--el-button-border-color);
  color: var(--el-button-text-color);
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
  border-bottom-color: var(--xm-accent-soft, rgba(0, 201, 167, 0.3));
}

.chapter-title :deep(.el-input__wrapper):focus-within {
  border-bottom-color: var(--xm-accent, #00c9a7);
}

.chapter-title :deep(.el-input__inner) {
  color: var(--xm-ink, #1a1c1e);
  font-weight: 600;
  font-family: 'Georgia', 'Times New Roman', serif;
  letter-spacing: 0.02em;
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
  background: var(--xm-surface, #ffffff);
  border: 1px solid var(--xm-line, #e4e2dd);
  border-radius: 14px;
}

/* 暗色主题适配 - 正文板块 */
:root[data-theme='dark'] .center-panel {
  background: transparent;
}

:root[data-theme='dark'] .editor-sheet {
  background: #161b22;
  border-color: rgba(42, 50, 66, 0.4);
  box-shadow: none;
}

:root[data-theme='dark'] .editor-sheet:focus-within {
  border-color: rgba(59, 130, 246, 0.35);
  box-shadow:
    0 1px 3px rgba(0, 0, 0, 0.25),
    0 12px 32px rgba(59, 130, 246, 0.08);
}

/* 页面主容器与面板间隙背景（content-wrapper 的 10px padding 形成三栏卡片之间的"缝隙"） */
:root[data-theme='dark'] .write-container,
:root[data-theme='dark'] .content-wrapper {
  background: rgba(15, 29, 51, 0.55);
}

:root[data-theme='dark'] .word-count-inline {
  color: #93c5fd;
}

/* ========== 右侧单栏（切换式）：AI 对话 / 场景大纲 / AI 写作 共用一栏 ========== */
.tw-right-dock {
  flex: none;
  position: relative;
  z-index: 2;
  display: flex;
  margin-left: 10px;
  border: 1px solid var(--xm-line, #e4e2dd);
  border-radius: 14px;
  background: var(--xm-surface, #ffffff);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
  overflow: hidden;
}
.tw-right-dock > * {
  flex: 1;
  min-width: 0;
}
.tw-right-dock .ds-panel { height: 100%; }

body.knowledge-graph-open {
  overflow: hidden !important;
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

.fixed-height-section .view-intro-icon-btn {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
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

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
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
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 50px 20px;
  color: #999;
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

/* ========== Element Plus 组件覆盖 ========== */
.left-panel :deep(.el-tabs) {
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
  border: none;
  border-radius: 0;
  background: var(--xm-surface, #ffffff);
  box-shadow: none;
  overflow: hidden;
}

.left-panel :deep(.el-tabs__header) {
  display: flex;
  margin: 0;
  padding: 8px 10px 0;
  border-bottom: 1px solid var(--xm-line, #ebedf0);
  background: transparent;
}
.left-panel :deep(.el-tabs__nav-wrap::after) {
  display: none;
}
.left-panel :deep(.el-tabs__nav) {
  display: flex;
  gap: 4px;
}
.left-panel :deep(.el-tabs__item) {
  height: 30px;
  line-height: 30px;
  font-size: 12px;
  font-weight: 600;
  color: var(--xm-ink-3, #9d9c96);
  padding: 0 12px;
  border-radius: 7px;
  transition: all .15s;
}
.left-panel :deep(.el-tabs__item:hover) {
  color: var(--xm-accent, #1c5cab);
  background: var(--xm-surface-3, #f0efeb);
}
.left-panel :deep(.el-tabs__item.is-active) {
  color: var(--xm-accent, #1c5cab);
  background: var(--xm-accent-soft, #e7effa);
  font-weight: 600;
}
.left-panel :deep(.el-tabs__active-bar) {
  display: none;
}

/* ========== 左侧栏 · 大纲 / 角色 信息条目分栏 ========== */
.xm-info-list {
  height: 100%;
  overflow: auto;
  padding: 8px 10px 16px;
}

.xm-info-group {
  padding: 10px 4px 6px;
  font-size: 11.5px;
  font-weight: 600;
  letter-spacing: 0.04em;
  color: var(--xm-ink-3, #9d9c96);
}

.xm-info-item {
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  transition: background var(--xm-dur, 0.2s) var(--xm-ease, ease);
}

.xm-info-item:hover {
  background: var(--xm-surface-3, #f0efeb);
}

.xm-info-item.is-active {
  background: var(--xm-accent-soft, #e7effa);
}

.xm-info-item__title {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--xm-ink, #1e1e1c);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.xm-info-item.is-active .xm-info-item__title {
  color: var(--xm-accent, #1c5cab);
}

.xm-info-item__desc {
  margin-top: 3px;
  font-size: 11.5px;
  line-height: 1.55;
  color: var(--xm-ink-2, #5a6270);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.xm-info-item__desc.is-empty {
  color: var(--xm-ink-3, #9d9c96);
  font-style: italic;
}

.xm-info-empty {
  padding: 32px 12px;
  text-align: center;
  font-size: 12px;
  color: var(--xm-ink-3, #999);
}

.xm-char-item {
  border: 1px solid var(--xm-line, #eef0f2);
  border-radius: 8px;
  margin-bottom: 6px;
  background: var(--xm-surface, #fff);
  cursor: pointer;
  transition: border-color var(--xm-dur, 0.2s) var(--xm-ease, ease), background var(--xm-dur, 0.2s) var(--xm-ease, ease);
}

.xm-char-item:hover {
  border-color: var(--xm-line-strong, #c9c6bf);
  background: var(--xm-surface-2, #f7f6f2);
}

.xm-char-item__head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
}

.xm-char-item__avatar {
  width: 30px;
  height: 30px;
  flex: none;
  display: grid;
  place-items: center;
  border-radius: 50%;
  font-size: 13px;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, #6366f1, #a855f7);
}

.xm-char-item__meta {
  flex: 1;
  min-width: 0;
}

.xm-char-item__name {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--xm-ink, #1e1e1c);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.xm-char-item__sub {
  margin-top: 1px;
  font-size: 11px;
  color: var(--xm-ink-3, #9d9c96);
}

.xm-char-item__caret {
  flex: none;
  font-size: 12px;
  color: var(--xm-ink-3, #9d9c96);
  transition: transform var(--xm-dur, 0.2s) var(--xm-ease, ease);
}

.xm-char-item.is-expanded .xm-char-item__caret {
  transform: rotate(180deg);
}

.xm-char-item__body {
  padding: 2px 10px 10px 48px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.xm-char-item__label {
  font-size: 11px;
  font-weight: 600;
  color: var(--xm-accent, #1c5cab);
}

.xm-char-item__text {
  font-size: 11.5px;
  line-height: 1.6;
  color: var(--xm-ink-2, #5a6270);
  white-space: pre-wrap;
  word-break: break-word;
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
  color: var(--xm-ink-2, #6b9b97);
  border: none;
  transition: all 0.3s;
}

:deep(.el-tabs__item:hover) {
  color: var(--xm-accent, #00c9a7);
}

:deep(.el-tabs__item.is-active) {
  color: var(--xm-accent, #00c9a7);
  font-weight: 700;
}

:deep(.el-tabs__active-bar) {
  background: var(--xm-accent, #00c9a7);
  height: 3px;
  border-radius: 3px;
}

:deep(.el-tab-pane) {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* 覆盖所有按钮样式（精致扁平：实色强调，无渐变） */
:deep(.el-button--primary){
  background: var(--xm-accent, #2563eb) !important;
  border: none !important;
  color: var(--xm-on-accent, #fff) !important;
  font-weight: 600;
  transition: all 0.2s var(--xm-ease, cubic-bezier(0.4, 0, 0.2, 1));
}

:deep(.el-button--primary:hover){
  background: var(--xm-accent, #2563eb) !important;
  filter: brightness(0.92);
  transform: none;
  box-shadow: none;
}

:deep(.el-button--primary:active){
  background: var(--xm-accent, #2563eb) !important;
  filter: brightness(0.85);
  transform: translateY(1px);
}

/* 对话框样式（精致扁平，与 write-theme.css 令牌一致；in-place 弹窗走此覆盖） */
:deep(.el-dialog) {
  --el-input-bg-color: var(--xm-surface-2, #f7f6f2);
  --el-input-border-color: var(--xm-line, #e7e4dc);
  --el-input-text-color: var(--xm-ink, #2c3138);
  --el-input-focus-border-color: var(--xm-accent, #2563eb);

  border-radius: var(--xm-radius, 16px);
  border: 1px solid var(--xm-line, #e7e4dc);
  background: var(--xm-surface, #ffffff);
  overflow: hidden;
  box-shadow: var(--xm-shadow-dialog, 0 20px 60px rgba(15, 23, 42, 0.18));
}

:deep(.el-dialog__header) {
  padding: 18px 22px;
  margin: 0;
  border-bottom: 1px solid var(--xm-line, #e7e4dc);
  background: transparent;
}

:deep(.el-dialog__title) {
  color: var(--xm-ink, #2c3138);
  font-weight: 600;
  font-size: 16px;
}

:deep(.el-dialog__headerbtn .el-dialog__close) {
  color: var(--xm-ink-3, #8a8f99);
}

:deep(.el-dialog__headerbtn:hover .el-dialog__close) {
  color: var(--xm-ink, #2c3138);
}

:deep(.el-dialog__body) {
  padding: 20px 22px;
  color: var(--xm-ink, #2c3138);
}

:deep(.el-dialog__footer) {
  padding: 14px 22px;
  border-top: 1px solid var(--xm-line, #e7e4dc);
  background: transparent;
}

.footer-actions {
  display: flex;
  gap: 8px;
}

/* 输入框样式 */
:deep(.el-input__wrapper) {
  border-radius: 8px;
  border: 1px solid var(--xm-line, rgba(203, 213, 225, 0.85));
  background: var(--xm-surface-2, transparent);
  transition: border-color 0.2s var(--xm-ease, ease), box-shadow 0.2s var(--xm-ease, ease);
}

:deep(.el-input__wrapper:hover) {
  border-color: var(--xm-accent-soft, rgba(77, 134, 255, 0.3));
}

:deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 3px var(--xm-accent-soft, rgba(77, 134, 255, 0.08));
  border-color: var(--xm-accent, #4d86ff);
}

/* 复选框样式 */
:deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
  background-color: var(--xm-accent, #4d86ff);
  border-color: var(--xm-accent, #4d86ff);
}

:deep(.el-checkbox__inner) {
  border-radius: 4px;
  border: 1px solid var(--xm-line-strong, rgba(77, 134, 255, 0.3));
}

/* 徽章样式 */
:deep(.el-badge__content.is-fixed) {
  background: var(--xm-accent, #2563eb);
  border: 2px solid #fff;
}

/* 遮罩层 — 撑满全屏，点击背景关闭（磨砂遮罩与 write-theme 令牌一致） */
.character-library-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  background: var(--xm-veil, rgba(0, 0, 0, 0.45));
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
}

/* 弹窗容器 — 大卡片（面 + 发丝线 + 柔和阴影） */
.character-library-overlay .modal-container {
  width: 100%;
  height: 100%;
  max-width: 1440px;
  max-height: 85vh;
  background: var(--xm-surface, #fff);
  border: 1px solid var(--xm-line, rgba(255, 255, 255, 0.08));
  border-radius: 16px;
  box-shadow: var(--xm-shadow-dialog, 0 25px 80px rgba(0, 0, 0, 0.35));
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
  background: var(--xm-surface-3, rgba(0, 0, 0, 0.06));
  color: var(--xm-ink-2, #666);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  transition: all 0.2s var(--xm-ease, ease);
}

.modal-close-btn:hover {
  background: var(--xm-accent-soft, rgba(0, 0, 0, 0.12));
  color: var(--xm-accent, #333);
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
  --el-bg-color: #161b22;
  --el-bg-color-overlay: #161b22;
  --el-text-color-primary: #e6edf3;
  --el-text-color-regular: #c8d6cd;
  --el-text-color-secondary: #5f6b7a;
  --el-text-color-placeholder: #4d5768;
  --el-border-color: #21262d;
  --el-border-color-light: #21262d;
  --el-border-color-lighter: #26352e;
  --el-border-color-extra-light: #1f2736;
  --el-fill-color: #21262d;
  --el-fill-color-light: #26352e;
  --el-fill-color-lighter: #1f2736;
  --el-fill-color-blank: #161b22;
  --el-color-white: #161b22;
  --el-color-black: #e6edf3;
  --el-color-info-light-9: #1f2736;
  --el-color-info-light-8: #26352e;
  --el-color-info-light-7: #21262d;
  --el-color-info-light-3: #4d5768;
  --el-color-primary: #93c5fd;
  --el-color-primary-light-3: #3b82f6;
  --el-color-primary-light-5: #22c55e;
  --el-color-primary-light-7: #16a34a;
  --el-color-primary-light-8: #2563eb;
  --el-color-primary-light-9: #1d4ed8;
  --el-color-primary-dark-2: #93c5fd;
  --el-color-danger: #f87171;
  --el-color-danger-light-3: #ef4444;
  --el-color-danger-light-5: #dc2626;
  --el-color-danger-light-7: #b91c1c;
  --el-color-danger-light-8: #991b1b;
  --el-color-danger-light-9: #7f1d1d;
  --el-color-success: #34d399;
  --el-color-warning: #fbbf24;
  --el-color-info: #5f6b7a;
  --el-color-error: #f87171;
  --el-box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);

  /* Button 变量 */
  --el-button-bg-color: #21262d;
  --el-button-border-color: #3f4a5c;
  --el-button-text-color: #e6edf3;
  --el-button-hover-bg-color: #3f4a5c;
  --el-button-hover-border-color: #93c5fd;
  --el-button-hover-text-color: #93c5fd;
  --el-button-active-bg-color: #26352e;
  --el-button-active-border-color: #93c5fd;
  --el-button-active-text-color: #93c5fd;
  --el-button-disabled-bg-color: #1f2736;
  --el-button-disabled-border-color: #21262d;
  --el-button-disabled-text-color: #4d5768;

  /* Input / Textarea 变量 */
  --el-input-bg-color: #161b22;
  --el-input-border-color: #21262d;
  --el-input-text-color: #e6edf3;
  --el-input-placeholder-color: #4d5768;
  --el-input-hover-border-color: #93c5fd;
  --el-input-focus-border-color: #93c5fd;
  --el-input-focus-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15);

  /* Select 变量 */
  --el-select-input-focus-border-color: #93c5fd;
  --el-popper-bg-color: #161b22;
  --el-popper-border-color: #21262d;

  /* Dialog 变量 */
  --el-dialog-bg-color: #161b22;
  --el-dialog-border-color: #21262d;
  --el-dialog-title-font-color: #e6edf3;
  --el-dialog-box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);

  /* Message 变量 */
  --el-message-bg-color: #161b22;
  --el-message-border-color: #21262d;
  --el-message-text-color: #e6edf3;
  --el-message-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);

  /* MessageBox 变量 */
  --el-messagebox-bg-color: #161b22;
  --el-messagebox-border-color: #21262d;
  --el-messagebox-title-text-color: #e6edf3;
  --el-messagebox-content-text-color: #c8d6cd;
  --el-messagebox-box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);

  /* Notification 变量 */
  --el-notification-bg-color: #161b22;
  --el-notification-border-color: #21262d;
  --el-notification-text-color: #e6edf3;
  --el-notification-box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);

  /* Tooltip / Popover 变量 */
  --el-popover-bg-color: #161b22;
  --el-popover-border-color: #21262d;
  --el-popover-text-color: #e6edf3;
  --el-popover-box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);

  /* Tag 变量 */
  --el-tag-bg-color: #1f2736;
  --el-tag-border-color: #21262d;
  --el-tag-text-color: #e6edf3;
  --el-tag-hover-color: #93c5fd;

  /* Table 变量 */
  --el-table-bg-color: #161b22;
  --el-table-tr-bg-color: #161b22;
  --el-table-border-color: #1f2736;
  --el-table-text-color: #e6edf3;
  --el-table-header-text-color: #5f6b7a;
  --el-table-row-hover-bg-color: #1f2736;
  --el-table-current-row-bg-color: #1f2633;

  /* Checkbox 变量 */
  --el-checkbox-bg-color: #161b22;
  --el-checkbox-border-color: #3f4a5c;
  --el-checkbox-text-color: #e6edf3;
  --el-checkbox-checked-bg-color: #93c5fd;
  --el-checkbox-checked-border-color: #93c5fd;
  --el-checkbox-checked-text-color: #93c5fd;
  --el-checkbox-input-border-color-hover: #93c5fd;

  /* Radio 变量 */
  --el-radio-bg-color: #161b22;
  --el-radio-border-color: #3f4a5c;
  --el-radio-text-color: #e6edf3;
  --el-radio-checked-text-color: #93c5fd;
  --el-radio-input-border-color-hover: #93c5fd;

  /* Switch 变量 */
  --el-switch-border-color: #3f4a5c;
  --el-switch-bg-color: #21262d;

  /* Slider 变量 */
  --el-slider-runway-bg-color: #21262d;
  --el-slider-stop-bg-color: #3f4a5c;
  --el-slider-bar-bg-color: #93c5fd;
  --el-slider-button-bg-color: #93c5fd;
  --el-slider-button-border-color: #93c5fd;

  /* Tabs 变量 */
  --el-tabs-header-border-color: #1f2736;
  --el-tab-item-text-color: #5f6b7a;
  --el-tab-item-active-text-color: #93c5fd;
  --el-tab-item-hover-text-color: #93c5fd;
}

/* =========================================================
   标签工作区骨架（Tab Workspace）— 基于 Demo 08 方案落地
   ========================================================= */
.write-container {
  background: #f5f4f1;
}

/* ---------- 菜单栏 ---------- */
.tw-menu {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: #f5f4f1;
  border-bottom: 1px solid #e4e2dd;
  flex-shrink: 0;
  position: relative;
  z-index: 30;
}
.tw-back {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 9px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--xm-ink-2, #666560);
  cursor: pointer;
  background: transparent;
}
.tw-back:hover {
  background: var(--xm-surface-3, #eceae6);
  color: var(--xm-ink, #1e1e1c);
}
.tw-book-chip {
  display: flex;
  flex-direction: column;
  line-height: 1.15;
  min-width: 0;
  padding: 3px 12px;
  border-left: 1px solid #e4e2dd;
  margin-left: 4px;
}
.tw-book-name {
  font-size: 14px;
  font-weight: 600;
  color: #1e1e1c;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 200px;
}
.tw-book-mode {
  font-size: 11px;
  color: #9d9c96;
  letter-spacing: .05em;
}
.tw-spacer { flex: 1; }
.tw-tool {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px !important;
  height: 32px !important;
  padding: 0 !important;
  border: none !important;
  border-radius: 9px !important;
  background: transparent !important;
  color: var(--xm-ink-2, #666560) !important;
  margin-left: 4px !important;
  transition: all .15s var(--xm-ease, ease);
}
.tw-tool:hover {
  background: var(--xm-surface-3, #eceae6) !important;
  color: var(--xm-ink, #1e1e1c) !important;
}
.tw-tool-accent {
  background: transparent !important;
  color: var(--xm-accent, #1c5cab) !important;
}
.tw-tool-accent:hover {
  background: var(--xm-accent-soft, #d6e8fa) !important;
  color: var(--xm-accent, #164a8f) !important;
}
.tw-tool-active {
  background: var(--xm-accent-soft, #e7effa) !important;
  color: var(--xm-accent, #1c5cab) !important;
}
.tw-tool .el-icon { font-size: 15px; }

/* ---------- 顶栏 / 标签栏 · 暗色主题 ---------- */
:root[data-theme='dark'] .tw-menu {
  background: #161b22;
  border-bottom-color: rgba(42, 50, 66, 0.4);
}
:root[data-theme='dark'] .tw-book-chip { border-left-color: rgba(42, 50, 66, 0.4); }
:root[data-theme='dark'] .tw-book-name { color: #e6edf3; }
:root[data-theme='dark'] .tw-book-mode { color: #5f6b7a; }
/* ---------- 标签栏 ---------- */
.tw-tabbar {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 4px;
  overflow-x: auto;
  padding: 0 8px;
}
.tw-tabbar::-webkit-scrollbar { height: 0; }
.tw-tabbar .el-dropdown { display: flex; }
.tw-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 8px 5px 12px;
  font-size: 12px;
  color: var(--xm-ink-2, #666560);
  border: 1px solid var(--xm-line, #e4e2dd);
  border-radius: 999px;
  cursor: pointer;
  background: var(--xm-surface, #fff);
  max-width: 180px;
  white-space: nowrap;
  flex: none;
  user-select: none;
  transition: background .15s var(--xm-ease, ease), color .15s var(--xm-ease, ease), border-color .15s var(--xm-ease, ease);
  position: relative;
}
.tw-tab:hover { background: var(--xm-surface-3, #f0efeb); }
.tw-tab.active {
  background: var(--xm-accent-soft, #e7effa);
  border-color: color-mix(in srgb, var(--xm-accent, #1c5cab) 35%, transparent);
  color: var(--xm-accent, #1c5cab);
  font-weight: 600;
}
.tw-tab-title { overflow: hidden; text-overflow: ellipsis; }
.tw-tab-dirty {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #c98a00;
  flex: none;
  box-shadow: 0 0 0 0 rgba(201,138,0,.3);
  animation: twDirty 1.6s infinite;
}
@keyframes twDirty { 50% { box-shadow: 0 0 0 4px rgba(201,138,0,.12); } }
.tw-tab-close {
  opacity: 0;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  font-size: 11px;
  color: var(--xm-ink-3, #9d9c96);
  flex: none;
  cursor: pointer;
}
.tw-tab:hover .tw-tab-close, .tw-tab.active .tw-tab-close { opacity: 1; }
.tw-tab-close:hover { background: var(--xm-surface-3, #e4e2dd); color: var(--xm-ink, #1e1e1c); }
.tw-tab-new {
  width: 26px;
  height: 26px;
  border: 1px dashed var(--xm-line-strong, #d5d2cc);
  border-radius: 999px;
  display: grid;
  place-items: center;
  color: var(--xm-ink-3, #9d9c96);
  background: transparent;
  cursor: pointer;
  flex: none;
}
.tw-tab-new:hover { color: var(--xm-accent, #1c5cab); border-color: var(--xm-accent, #1c5cab); }

/* ---------- 目录（左）折叠头 ---------- */
.tw-explorer-head {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  font-size: 11.5px;
  color: #9d9c96;
  font-weight: 600;
  letter-spacing: .04em;
  border-bottom: 1px solid rgba(228, 226, 221, .6);
}
.tw-explorer-title { flex: 1; white-space: nowrap; overflow: hidden; }
/* 完全折叠：面板宽度为 0，边框/投影隐藏，不留缝 */
.tw-explorer-collapsed,
:root[data-theme='dark'] .tw-explorer-collapsed {
  border-color: transparent;
  box-shadow: none;
}

/* ---------- 中心编辑器（列内含迷你地图，与左列表/右面板构成三栏） ---------- */
.center-panel {
  flex: 1;
  flex-shrink: 1;
  min-width: 0;
  display: flex;
  flex-direction: row;
  overflow: hidden;
  background: transparent;
}
.center-panel > .center-editor {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

/* ---------- 右侧单栏（切换式）样式见上方 .tw-right-dock ---------- */

/* 拖拽调宽手柄 */
.tw-resize-between {
  flex: none;
  width: 10px;
  cursor: col-resize;
  position: relative;
  z-index: 15;
}
.tw-resize-between::after {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: -1px;
  width: 2px;
  background: transparent;
  transition: background .15s;
}
.tw-resize-between:hover::after {
  background: rgba(42, 120, 214, .35);
}
.tw-resize-collapse {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #e4e2dd;
  background: #fff;
  border-radius: 50%;
  color: #9d9c96;
  font-size: 11px;
  cursor: pointer;
  box-shadow: 0 1px 3px rgba(30, 30, 28, .12);
  transition: color .15s, border-color .15s;
}
.tw-resize-collapse:hover {
  color: #1c5cab;
  border-color: #1c5cab;
}
/* 折叠态：调宽手柄整体隐藏（宽度 0、不拦截点击），不留缝隙；
   展开按钮随容器贴在窗口左缘（按钮中心落于内边距中线，可点击展开） */
.tw-resize-between--collapsed {
  width: 0;
  pointer-events: none;
}
.tw-resize-between--collapsed::after {
  display: none;
}
.tw-resize-between--collapsed .tw-resize-collapse {
  pointer-events: auto;
}
.tw-resize-edge {
  position: absolute;
  left: -4px;
  top: 0;
  bottom: 0;
  width: 8px;
  cursor: col-resize;
  z-index: 60;
}
.tw-resize-edge::after {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: 3px;
  width: 2px;
  background: transparent;
  transition: background .15s;
}
.tw-resize-edge:hover::after {
  background: rgba(42, 120, 214, .35);
}
/* 右侧栏折叠按钮：避让 .tw-right-dock 的 overflow:hidden（按钮放其内部会被裁剪），
   作为内容区兄弟节点绝对定位，按钮中心对齐右侧栏左缘 */
.tw-resize-collapse--right {
  position: absolute;
  top: 50%;
  left: auto;
  transform: translateY(-50%);
  z-index: 61;
}

.tw-confirm-text { font-size: 13px; color: var(--xm-ink-2, #666560); line-height: 1.8; }

/* =========================================================
   编辑器正文留白：左右留白收窄、去掉 max-width 上限，
   充分利用编辑区宽度；文本对齐交由默认/继承决定。
   ========================================================= */
.editor-container {
  padding: 0;
}
.editor-header {
  padding-left: 24px;
  padding-right: 24px;
  border-bottom: 1px solid #f0efe9;
}
:root[data-theme='dark'] .editor-header {
  border-bottom-color: rgba(59, 130, 246, 0.08);
}
.chapter-title :deep(.el-input__inner) {
  text-align: left;
  font-size: 17px;
  font-weight: 600;
}
.chapter-content :deep(.vditor--wysiwyg .vditor-wysiwyg),
.chapter-content :deep(.vditor-wysiwyg) {
  /* vditor JS 会按 preview.maxWidth(默认800) 把正文居中成"纸面"，
     inline padding 通常约 198px；这里把 wrapper 留白归零，
     实际留白在下方 .vditor-reset 上用 !important 强制。 */
  padding: 0;
  max-width: none;
}
/* vditor 用 JS 给内容 pre.vditor-reset 写 inline padding（inline 优先级高于
   普通选择器），必须 !important 才能强制为与标题(24px)对齐的左右留白。 */
.chapter-content :deep(.vditor-wysiwyg pre.vditor-reset),
.chapter-content :deep(.vditor-wysiwyg .vditor-reset) {
  padding: 30px 24px 160px 24px !important;
}
/* 工具栏也随内容被 vditor 写了 inline padding-left，需一并强制对齐 */
.chapter-content :deep(.vditor-toolbar) {
  padding-left: 24px !important;
  padding-right: 24px;
}
/* 备忘录 textarea 同样收窄左右留白 */
.editor-container :deep(.el-textarea__inner) {
  padding-left: 24px;
  padding-right: 24px;
}

/* 移动端适配 */
@media (max-width: 760px) {
  .tw-tool { margin-left: 2px !important; }
}

/* ---------- 设置抽屉 · 暗色主题 ---------- */
:root[data-theme='dark'] .write-settings-drawer {
  --el-drawer-bg-color: #161b22;
  --el-drawer-title-color: #e6edf3;
  --el-drawer-border-color: #21262d;
}
</style>

<!-- ===================== 知识图谱弹窗（全局样式：el-dialog append-to-body Teleport 内容需用全局选择器匹配） ===================== -->
<style>
/* 注意：class="knowledge-graph-dialog" 透传到 dialog 根元素本身，须用同元素多类选择器 */
.knowledge-graph-dialog.el-dialog {
  margin: 0;
  position: fixed !important;
  top: 0 !important;
  left: 50%;
  transform: translateX(-50%);
  height: 100vh;
  max-height: 100vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.knowledge-graph-dialog .el-dialog__header {
  padding: 12px 20px;
  margin: 0;
  flex-shrink: 0;
}

.knowledge-graph-dialog .el-dialog__body {
  padding: 0;
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.knowledge-graph-dialog .knowledge-graph-panel {
  height: 100%;
  min-height: 0;
}
</style>
