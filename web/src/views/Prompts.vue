<template>
  <div class="prompts-container">
    <!-- 顶栏：标题 + 全局搜索 -->
    <div class="top-bar">
      <div class="top-title">
        <h2>提示词管理</h2>
        <span class="top-sub">主区只看内容，所有操作交给快捷中心</span>
      </div>
      <div class="top-search">
        <el-input v-model="searchKeyword" placeholder="搜索提示词名称或内容…" clearable :prefix-icon="Search" />
      </div>
    </div>

    <!-- 统计磁贴：前两个统计「卡包个数」，独立卡片统计张数，最后一个统计卡片总数 -->
    <div class="stat-row">
      <div class="stat-tile" @click="activeMainTab = 'pack'">
        <div class="stat-ic ic-purple"><el-icon><FolderOpened /></el-icon></div>
        <div class="stat-body"><b>{{ totalNormalPacks }}</b><span>常规卡包</span></div>
      </div>
      <div class="stat-tile" @click="activeMainTab = 'bookAnalysis'">
        <div class="stat-ic ic-pink"><el-icon><Reading /></el-icon></div>
        <div class="stat-body"><b>{{ totalBookAnalysisPacks }}</b><span>拆书卡包</span></div>
      </div>
      <div class="stat-tile" @click="activeMainTab = 'standalone'">
        <div class="stat-ic ic-green"><el-icon><Document /></el-icon></div>
        <div class="stat-body"><b>{{ combinedUncategorizedPrompts.length }}</b><span>独立卡片</span></div>
      </div>
      <div class="stat-tile">
        <div class="stat-ic ic-orange"><el-icon><Collection /></el-icon></div>
        <div class="stat-body"><b>{{ totalPromptCards }}</b><span>总卡片</span></div>
      </div>
    </div>

    <!-- Tab 分段 -->
    <div class="chip-row">
      <button class="chip" :class="{ on: activeMainTab === 'all' }" @click="activeMainTab = 'all'">
        全部<span class="chip-n">{{ totalPromptCards }}</span>
      </button>
      <button class="chip" :class="{ on: activeMainTab === 'pack' }" @click="activeMainTab = 'pack'">
        常规卡包<span class="chip-n">{{ totalNormalPacks }}</span>
      </button>
      <button class="chip" :class="{ on: activeMainTab === 'bookAnalysis' }" @click="activeMainTab = 'bookAnalysis'">
        拆书卡包<span class="chip-n">{{ totalBookAnalysisPacks }}</span>
      </button>
      <button class="chip" :class="{ on: activeMainTab === 'standalone' }" @click="activeMainTab = 'standalone'">
        独立卡片<span class="chip-n">{{ combinedUncategorizedPrompts.length }}</span>
      </button>
    </div>

    <!-- 内容区 -->
    <div class="content-area">
      <!-- ═══════ 全部 ═══════ -->
      <template v-if="activeMainTab === 'all'">
        <div
          class="cards-grid"
          :class="{ 'drag-over': isDragging }"
          @dragover.prevent="onDragOver"
          @drop.prevent="onDrop('未分类', $event)"
        >
          <div
            v-for="category in filteredCategoryList"
            :key="'p-' + category.name"
            class="card card-pack"
            :class="{ 'drag-over': isDragging && draggedPrompt?.category !== category.name }"
            @dragover.prevent="onDragOver"
            @drop.prevent.stop="onDrop(category.name, $event)"
          >
            <div class="card-main" @click="goToPackDetail(category.name)">
              <div class="card-ic ic-blue"><el-icon><FolderOpened /></el-icon></div>
              <div class="card-info">
                <div class="card-name">{{ category.name }}</div>
                <div class="card-desc">{{ category.prompts.length }} 张卡片</div>
              </div>
            </div>
            <div class="card-foot">
              <span class="card-tag tg-pack">卡包</span>
              <span class="card-foot-time">{{ category.prompts.length }} 张</span>
              <div class="card-ops" @click.stop>
                <el-dropdown trigger="click" @command="(cmd: string) => handlePackAction(cmd, category.name, category.prompts.length)">
                  <button class="card-more"><el-icon><MoreFilled /></el-icon></button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="add">
                        <el-icon><Plus /></el-icon>添加卡片
                      </el-dropdown-item>
                      <el-dropdown-item v-if="category.prompts.length > 0 && !DEFAULT_CATEGORIES.includes(category.name)" command="batchDelete">
                        <el-icon><Delete /></el-icon>批量删除
                      </el-dropdown-item>
                      <el-dropdown-item v-if="!DEFAULT_CATEGORIES.includes(category.name)" command="deletePack" divided>
                        <el-icon style="color: #f56c6c;"><Delete /></el-icon>
                        <span style="color: #f56c6c;">删除卡包</span>
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
            </div>
          </div>

          <div
            v-for="category in filteredBookAnalysisCategoryList"
            :key="'b-' + category.name"
            class="card card-pack card-book"
            :class="{ 'drag-over': isDragging && draggedPrompt?.category !== category.name }"
            @dragover.prevent="onDragOver"
            @drop.prevent.stop="onBookAnalysisDrop(category.name, $event)"
          >
            <div class="card-main" @click="goToPackDetail(category.name)">
              <div class="card-ic ic-pink"><el-icon><Reading /></el-icon></div>
              <div class="card-info">
                <div class="card-name">{{ category.name.replace(BOOK_ANALYSIS_PREFIX, '') }}</div>
                <div class="card-desc">拆书卡包 · {{ category.prompts.length }} 张卡片</div>
              </div>
            </div>
            <div class="card-foot">
              <span class="card-tag tg-book">拆书</span>
              <span class="card-foot-time">{{ category.prompts.length }} 张</span>
              <div class="card-ops" @click.stop>
                <el-dropdown trigger="click" @command="(cmd: string) => handleBookAnalysisPackAction(cmd, category.name, category.prompts.length)">
                  <button class="card-more"><el-icon><MoreFilled /></el-icon></button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="add">
                        <el-icon><Plus /></el-icon>添加卡片
                      </el-dropdown-item>
                      <el-dropdown-item command="deletePack" divided>
                        <el-icon style="color: #f56c6c;"><Delete /></el-icon>
                        <span style="color: #f56c6c;">删除卡包</span>
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
            </div>
          </div>

          <div
            v-for="prompt in filteredStandalonePrompts"
            :key="prompt.id"
            class="card card-std"
            :class="{ dragging: isDragging && draggedPrompt?.id === prompt.id, encrypted: prompt.card_type === 'encrypted' }"
            draggable="true"
            @dragstart="onDragStart(prompt, $event)"
            @dragend="onDragEnd"
            @click="handleEdit(prompt)"
          >
            <div class="card-main">
              <div class="card-ic ic-green"><el-icon><Document /></el-icon></div>
              <div class="card-info">
                <div class="card-name">
                  <el-icon class="card-grip"><Rank /></el-icon>
                  <span class="card-name-text">{{ prompt.name }}</span>
                  <el-icon v-if="prompt.card_type === 'encrypted'" class="std-lock"><Lock /></el-icon>
                </div>
                <div class="card-desc std-preview">
                  <template v-if="prompt.card_type === 'encrypted'">
                    <el-icon><Lock /></el-icon> 内容已加密
                  </template>
                  <template v-else>
                    {{ prompt.content.slice(0, 40) }}{{ prompt.content.length > 40 ? '...' : '' }}
                  </template>
                </div>
              </div>
            </div>
            <div class="card-foot">
              <span class="card-tag" :class="prompt.category === '拆书-未分类' ? 'tg-book' : 'tg-std'">{{ prompt.category === '拆书-未分类' ? '拆书' : '独立' }}</span>
              <span class="card-foot-time">{{ formatDate(prompt.created_at) }}</span>
              <div class="std-actions" @click.stop>
                <el-tooltip content="编辑" placement="top" :show-after="300">
                  <button class="std-act" @click="handleEdit(prompt)"><el-icon><Edit /></el-icon></button>
                </el-tooltip>
                <el-tooltip content="预览" placement="top" :show-after="300">
                  <button class="std-act" @click="handlePreview(prompt)"><el-icon><View /></el-icon></button>
                </el-tooltip>
                <el-tooltip content="删除" placement="top" :show-after="300">
                  <button class="std-act danger" @click="handleDelete(prompt)"><el-icon><Delete /></el-icon></button>
                </el-tooltip>
              </div>
            </div>
          </div>
        </div>
        <div
          v-if="filteredCategoryList.length === 0 && filteredBookAnalysisCategoryList.length === 0 && filteredStandalonePrompts.length === 0"
          class="no-search-match"
        >
          <el-empty :description="searchKeyword ? '没有匹配的内容' : '暂无内容'" />
        </div>
      </template>

      <!-- ═══════ 卡包 ═══════ -->
      <template v-else-if="activeMainTab === 'pack'">
        <div class="cards-grid">
          <div
            v-for="category in filteredCategoryList"
            :key="'p-' + category.name"
            class="card card-pack"
            :class="{ 'drag-over': isDragging && draggedPrompt?.category !== category.name }"
            @dragover.prevent="onDragOver"
            @drop.prevent.stop="onDrop(category.name, $event)"
          >
            <div class="card-main" @click="goToPackDetail(category.name)">
              <div class="card-ic ic-blue"><el-icon><FolderOpened /></el-icon></div>
              <div class="card-info">
                <div class="card-name">{{ category.name }}</div>
                <div class="card-desc">{{ category.prompts.length }} 张卡片</div>
              </div>
            </div>
            <div class="card-foot">
              <span class="card-tag tg-pack">卡包</span>
              <span class="card-foot-time">{{ category.prompts.length }} 张</span>
              <div class="card-ops" @click.stop>
                <el-dropdown trigger="click" @command="(cmd: string) => handlePackAction(cmd, category.name, category.prompts.length)">
                  <button class="card-more"><el-icon><MoreFilled /></el-icon></button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="add">
                        <el-icon><Plus /></el-icon>添加卡片
                      </el-dropdown-item>
                      <el-dropdown-item v-if="category.prompts.length > 0 && !DEFAULT_CATEGORIES.includes(category.name)" command="batchDelete">
                        <el-icon><Delete /></el-icon>批量删除
                      </el-dropdown-item>
                      <el-dropdown-item v-if="!DEFAULT_CATEGORIES.includes(category.name)" command="deletePack" divided>
                        <el-icon style="color: #f56c6c;"><Delete /></el-icon>
                        <span style="color: #f56c6c;">删除卡包</span>
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
            </div>
          </div>
        </div>
        <div v-if="filteredCategoryList.length === 0" class="no-search-match">
          <el-empty :description="searchKeyword ? '没有匹配的卡包' : '暂无卡包'" />
        </div>
      </template>

      <!-- ═══════ 拆书 ═══════ -->
      <template v-else-if="activeMainTab === 'bookAnalysis'">
        <div class="cards-grid">
          <div
            v-for="category in filteredBookAnalysisCategoryList"
            :key="'b-' + category.name"
            class="card card-pack card-book"
            :class="{ 'drag-over': isDragging && draggedPrompt?.category !== category.name }"
            @dragover.prevent="onDragOver"
            @drop.prevent.stop="onBookAnalysisDrop(category.name, $event)"
          >
            <div class="card-main" @click="goToPackDetail(category.name)">
              <div class="card-ic ic-pink"><el-icon><Reading /></el-icon></div>
              <div class="card-info">
                <div class="card-name">{{ category.name.replace(BOOK_ANALYSIS_PREFIX, '') }}</div>
                <div class="card-desc">拆书卡包 · {{ category.prompts.length }} 张卡片</div>
              </div>
            </div>
            <div class="card-foot">
              <span class="card-tag tg-book">拆书</span>
              <span class="card-foot-time">{{ category.prompts.length }} 张</span>
              <div class="card-ops" @click.stop>
                <el-dropdown trigger="click" @command="(cmd: string) => handleBookAnalysisPackAction(cmd, category.name, category.prompts.length)">
                  <button class="card-more"><el-icon><MoreFilled /></el-icon></button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="add">
                        <el-icon><Plus /></el-icon>添加卡片
                      </el-dropdown-item>
                      <el-dropdown-item command="deletePack" divided>
                        <el-icon style="color: #f56c6c;"><Delete /></el-icon>
                        <span style="color: #f56c6c;">删除卡包</span>
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
            </div>
          </div>
        </div>
        <div
          v-if="filteredBookAnalysisCategoryList.length === 0 && bookAnalysisCategoryList.length === 0"
          class="empty-standalone-hint"
          @click="handleCreateBookAnalysisPrompt()"
        >
          <el-icon><Plus /></el-icon>
          <span>暂无拆书提示词，点击创建或从「快捷中心 → 导入」拆书卡包</span>
        </div>
        <div v-else-if="filteredBookAnalysisCategoryList.length === 0" class="no-search-match">
          <el-empty description="没有匹配的拆书卡包" />
        </div>
      </template>

      <!-- ═══════ 独立卡片 ═══════ -->
      <template v-else>
        <div class="standalone-toolbar" v-if="combinedUncategorizedPrompts.length > 0">
          <span class="std-toolbar-title">独立卡片 <em>（含拆书-未分类）</em></span>
          <el-tooltip content="批量删除" placement="top" :show-after="300">
            <button class="std-batch" @click="openStandaloneBatchDeleteDialog"><el-icon><Delete /></el-icon></button>
          </el-tooltip>
        </div>
        <div
          v-if="combinedUncategorizedPrompts.length > 0"
          class="cards-grid std-area"
          :class="{ 'drag-over': isDragging }"
          @dragover.prevent="onDragOver"
          @drop.prevent="onDrop('未分类', $event)"
        >
          <div v-if="filteredStandalonePrompts.length === 0" class="no-search-match">
            <el-empty description="没有匹配的独立卡片" />
          </div>
          <div
            v-for="prompt in filteredStandalonePrompts"
            :key="prompt.id"
            class="card card-std"
            :class="{ dragging: isDragging && draggedPrompt?.id === prompt.id, encrypted: prompt.card_type === 'encrypted' }"
            draggable="true"
            @dragstart="onDragStart(prompt, $event)"
            @dragend="onDragEnd"
            @click="handleEdit(prompt)"
          >
            <div class="card-main">
              <div class="card-ic ic-green"><el-icon><Document /></el-icon></div>
              <div class="card-info">
                <div class="card-name">
                  <el-icon class="card-grip"><Rank /></el-icon>
                  <span class="card-name-text">{{ prompt.name }}</span>
                  <el-icon v-if="prompt.card_type === 'encrypted'" class="std-lock"><Lock /></el-icon>
                </div>
                <div class="card-desc std-preview">
                  <template v-if="prompt.card_type === 'encrypted'">
                    <el-icon><Lock /></el-icon> 内容已加密
                  </template>
                  <template v-else>
                    {{ prompt.content.slice(0, 40) }}{{ prompt.content.length > 40 ? '...' : '' }}
                  </template>
                </div>
              </div>
            </div>
            <div class="card-foot">
              <span class="card-tag" :class="prompt.category === '拆书-未分类' ? 'tg-book' : 'tg-std'">{{ prompt.category === '拆书-未分类' ? '拆书' : '独立' }}</span>
              <span class="card-foot-time">{{ formatDate(prompt.created_at) }}</span>
              <div class="std-actions" @click.stop>
                <el-tooltip content="编辑" placement="top" :show-after="300">
                  <button class="std-act" @click="handleEdit(prompt)"><el-icon><Edit /></el-icon></button>
                </el-tooltip>
                <el-tooltip content="预览" placement="top" :show-after="300">
                  <button class="std-act" @click="handlePreview(prompt)"><el-icon><View /></el-icon></button>
                </el-tooltip>
                <el-tooltip content="删除" placement="top" :show-after="300">
                  <button class="std-act danger" @click="handleDelete(prompt)"><el-icon><Delete /></el-icon></button>
                </el-tooltip>
              </div>
            </div>
          </div>
        </div>
        <div
          v-else
          class="empty-standalone-hint"
          :class="{ 'drag-over': isDragging }"
          @click="handleCreate"
          @dragover.prevent="onDragOver"
          @drop.prevent="onDrop('未分类', $event)"
        >
          <el-icon><Plus /></el-icon>
          <span>暂无独立卡片，点击创建或拖拽卡片到此处</span>
        </div>
      </template>
    </div>

    <!-- FAB：打开快捷中心 -->
    <button class="fab" @click="sheetVisible = !sheetVisible" title="快捷中心">
      <el-icon :size="24"><Plus /></el-icon>
    </button>

    <!-- 底部快捷中心：新建 / 导入 / 导出 统一入口（含拆书） -->
    <transition name="sheetup">
      <div v-if="sheetVisible" class="sheet-mask" @click.self="sheetVisible = false">
        <div class="sheet">
          <div class="sheet-grip"></div>
          <div class="sheet-head">
            <b>快捷中心</b>
            <span>新建 · 导入 · 导出（含拆书，统一入口）</span>
          </div>
          <div class="sheet-section">
            <div class="sheet-label">✦ 新建</div>
            <div class="sheet-row" @click="sheetVisible = false; handleCreate()">
              <div class="sheet-ic ic-green"><el-icon><Plus /></el-icon></div>
              <div class="sheet-info"><b>创建提示词</b><span>普通 / 加密卡片</span></div>
              <span class="sheet-go">›</span>
            </div>
            <div class="sheet-row" @click="sheetVisible = false; openCategoryDialog()">
              <div class="sheet-ic ic-purple"><el-icon><FolderAdd /></el-icon></div>
              <div class="sheet-info"><b>创建卡包</b><span>普通卡包</span></div>
              <span class="sheet-go">›</span>
            </div>
            <div class="sheet-row" @click="sheetVisible = false; openBookAnalysisCategoryDialog()">
              <div class="sheet-ic ic-pink"><el-icon><Reading /></el-icon></div>
              <div class="sheet-info"><b>创建拆书卡包</b><span>自动加「拆书-」前缀</span></div>
              <span class="sheet-go">›</span>
            </div>
            <div class="sheet-row" @click="sheetVisible = false; openMergeDialog()">
              <div class="sheet-ic ic-blue"><el-icon><CopyDocument /></el-icon></div>
              <div class="sheet-info"><b>卡包合并</b><span>多卡包合并 / 主包吸收</span></div>
              <span class="sheet-go">›</span>
            </div>
          </div>
          <div class="sheet-sep"></div>
          <div class="sheet-section">
            <div class="sheet-label">⤵ 导入</div>
            <div class="sheet-row" @click="sheetVisible = false; openImportAllDialog()">
              <div class="sheet-ic ic-red"><el-icon><Upload /></el-icon></div>
              <div class="sheet-info"><b>恢复整体备份</b><span>覆盖全部数据 · 不可恢复</span></div>
              <span class="sheet-go">›</span>
            </div>
            <div class="sheet-row" @click="sheetVisible = false; openImportPackDialog()">
              <div class="sheet-ic ic-blue"><el-icon><FolderOpened /></el-icon></div>
              <div class="sheet-info"><b>导入卡包</b><span>自动识别普通 / 拆书卡包文件</span></div>
              <span class="sheet-go">›</span>
            </div>
            <div class="sheet-row" @click="sheetVisible = false; openImportStandaloneDialog()">
              <div class="sheet-ic ic-blue"><el-icon><Document /></el-icon></div>
              <div class="sheet-info"><b>导入独立卡片</b><span>含拆书-未分类卡片</span></div>
              <span class="sheet-go">›</span>
            </div>
            <div class="sheet-row" @click="sheetVisible = false; openImportConvertDialog()">
              <div class="sheet-ic ic-blue"><el-icon><Switch /></el-icon></div>
              <div class="sheet-info"><b>txt / md 转换导入</b><span>批量拖入，可合并为卡包</span></div>
              <span class="sheet-go">›</span>
            </div>
          </div>
          <div class="sheet-sep"></div>
          <div class="sheet-section">
            <div class="sheet-label">⤴ 导出</div>
            <div class="sheet-row" @click="sheetVisible = false; executeExportAll()">
              <div class="sheet-ic ic-amber"><el-icon><Download /></el-icon></div>
              <div class="sheet-info"><b>整体备份导出</b><span>卡包 + 拆书 + 独立卡片</span></div>
              <span class="sheet-go">›</span>
            </div>
            <div class="sheet-row" @click="sheetVisible = false; openExportPackDialog()">
              <div class="sheet-ic ic-amber"><el-icon><FolderOpened /></el-icon></div>
              <div class="sheet-info"><b>导出卡包</b><span>勾选普通 / 拆书卡包</span></div>
              <span class="sheet-go">›</span>
            </div>
            <div class="sheet-row" @click="sheetVisible = false; openExportStandaloneDialog()">
              <div class="sheet-ic ic-amber"><el-icon><Document /></el-icon></div>
              <div class="sheet-info"><b>导出独立卡片</b><span>含拆书-未分类</span></div>
              <span class="sheet-go">›</span>
            </div>
          </div>
          <div class="sheet-sep"></div>
          <div class="sheet-section">
            <div class="sheet-label">✦ 管理</div>
            <div class="sheet-row" @click="sheetVisible = false; openDedupDialog()">
              <div class="sheet-ic ic-red"><el-icon><Search /></el-icon></div>
              <div class="sheet-info"><b>提示词查重</b><span>清理同名重复卡片</span></div>
              <span class="sheet-go">›</span>
            </div>
          </div>
        </div>
      </div>
    </transition>

    <PromptIODialogs
      ref="ioDialogsRef"
      :prompts="prompts"
      :combined-uncategorized-prompts="combinedUncategorizedPrompts"
      :all-exportable-packs="allExportablePacks"
      :category-list="categoryList"
      :book-analysis-category-list="bookAnalysisCategoryList"
      :custom-categories="customCategories"
      :refresh="fetchPrompts"
      @add-category="addIOCustomCategory"
      @set-categories="setIOCustomCategories"
    />

    <DedupMergeDialogs
      ref="dedupMergeRef"
      :prompts="prompts"
      :all-exportable-packs="allExportablePacks"
      :custom-categories="customCategories"
      :refresh="fetchPrompts"
      @add-category="addDedupCustomCategory"
      @remove-categories="removeDedupCustomCategories"
    />

    <!-- 创建/编辑卡包对话框 -->
    <el-dialog
      v-model="categoryDialogVisible"
      title="创建卡包"
      width="400px"
      :close-on-click-modal="false"
      append-to-body
    >
      <el-form :model="categoryForm" label-width="80px">
        <el-form-item label="卡包名称" required>
          <el-input v-model="categoryForm.name" placeholder="请输入卡包名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="categoryDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateCategory">确定</el-button>
      </template>
    </el-dialog>

    <!-- 创建拆书卡包对话框 -->
    <el-dialog
      v-model="bookAnalysisCategoryDialogVisible"
      title="创建拆书卡包"
      width="400px"
      :close-on-click-modal="false"
      append-to-body
    >
      <el-form :model="bookAnalysisCategoryForm" label-width="80px">
        <el-form-item label="卡包名称" required>
          <el-input v-model="bookAnalysisCategoryForm.name" placeholder="请输入卡包名称（自动加拆书-前缀）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="bookAnalysisCategoryDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateBookAnalysisCategory">确定</el-button>
      </template>
    </el-dialog>


    <BatchDeleteDialogs
      ref="batchDeleteDialogsRef"
      :category-list="categoryList"
      :combined-uncategorized-prompts="combinedUncategorizedPrompts"
      @deleted="fetchPrompts"
    />

    <!-- 创建/编辑提示词对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑提示词' : '创建提示词'"
      width="1200px"
      top="4vh"
      :close-on-click-modal="false"
      destroy-on-close
      class="pack-edit-dialog"
    >
      <!-- 顶部操作栏 -->
      <div class="pack-edit-opbar">
        <span class="pack-edit-opbar-title">基本信息</span>
        <div class="pack-edit-opbar-actions">
          <el-button size="small" text @click="handlePasteImport">
            <el-icon><DocumentAdd /></el-icon>
            粘贴导入
          </el-button>
          <el-button size="small" text @click="openPreviewTest" :disabled="!formData.content">
            <el-icon><View /></el-icon>
            预览测试
          </el-button>
          <el-button size="small" text @click="genParamsVisible = true">
            <el-icon><Setting /></el-icon>
            调整生成参数
          </el-button>
          <el-button size="small" text @click="usageGuideVisible = true">
            <el-icon><QuestionFilled /></el-icon>
            使用说明
          </el-button>
        </div>
      </div>
      <div class="pack-edit-layout">
        <!-- 左侧：表单字段 (2/3) -->
        <div class="pack-edit-left">
          <el-form :model="formData" label-position="top" class="pack-edit-form">
            <el-form-item label="适用作品">
              <el-radio-group v-model="formData.workType" size="small">
                <el-radio-button value="novel">小说</el-radio-button>
                <el-radio-button value="script">剧本</el-radio-button>
              </el-radio-group>
            </el-form-item>

            <el-form-item label="名称" required>
              <el-input
                v-model="formData.name"
                placeholder="请输入名称"
                maxlength="100"
                show-word-limit
              />
              <div class="pack-edit-field-tip">
                <el-icon><InfoFilled /></el-icon>
                起一个有特点、有吸引力的名字，让用户一眼就能看出提示词的特色
              </div>
            </el-form-item>

            <el-form-item label="分类" required>
              <el-select
                v-model="formData.category"
                placeholder="请选择卡包"
                style="width: 100%"
              >
                <el-option
                  v-for="cat in categoryOptions"
                  :key="cat"
                  :label="cat"
                  :value="cat"
                />
              </el-select>
            </el-form-item>

            <el-form-item label="作者">
              <el-input v-model="formData.creator_name" placeholder="请输入作者名，非必填" />
            </el-form-item>

            <el-form-item label="版本">
              <el-input v-model="formData.version" placeholder="如 v1.0、v2.1.0，非必填" />
            </el-form-item>
            <el-form-item label="卡片类型">
              <div class="pack-edit-card-type">
                <el-radio-group v-model="formData.card_type" @change="handleCardTypeChange" class="pack-edit-pills">
                  <el-radio-button value="normal">
                    <el-icon><Document /></el-icon>
                    普通卡片
                  </el-radio-button>
                  <el-radio-button value="encrypted">
                    <el-icon><Lock /></el-icon>
                    加密卡片
                  </el-radio-button>
                </el-radio-group>
                <div v-if="formData.card_type === 'encrypted'" class="pack-edit-card-type-actions">
                  <el-button
                    v-if="!formData.password"
                    type="warning"
                    size="small"
                    @click="openSetPasswordDialog"
                  >
                    <el-icon><Lock /></el-icon>
                    设置密码
                  </el-button>
                  <el-button
                    v-else
                    type="success"
                    size="small"
                    @click="openSetPasswordDialog"
                  >
                    <el-icon><Lock /></el-icon>
                    修改密码
                  </el-button>
                  <el-button
                    v-if="formData.password"
                    type="danger"
                    size="small"
                    link
                    @click="removeCardPassword"
                  >
                    移除密码
                  </el-button>
                </div>
                <div v-if="formData.card_type === 'encrypted' && formData.password" class="pack-edit-card-type-status">
                  <el-tag type="warning" size="small">
                    <el-icon><Lock /></el-icon>
                    已加密
                  </el-tag>
                </div>
                <div v-if="formData.card_type === 'encrypted' && !formData.password" class="pack-edit-card-type-status">
                  <el-tag type="info" size="small">未设置密码（保存前需设置密码）</el-tag>
                </div>
              </div>
            </el-form-item>
            <el-form-item label="使用模式">
              <div class="pack-edit-usage-mode">
                <el-radio-group v-model="formData.usageMode" @change="handleUsageModeChange" class="pack-edit-pills">
                  <el-radio-button value="private">
                    <el-icon><Lock /></el-icon>
                    自用
                  </el-radio-button>
                  <el-radio-button value="public">
                    <el-icon><Share /></el-icon>
                    公用
                  </el-radio-button>
                </el-radio-group>
                <el-alert
                  v-if="formData.usageMode === 'public'"
                  type="warning"
                  :closable="false"
                  show-icon
                  title="保存后需审核"
                  description="当前为公用模式，保存后需审核通过才可见。"
                  class="pack-edit-usage-alert"
                />
                <div class="pack-edit-content-public">
                  <span class="pack-edit-content-public-label">提示词正文内容是否公开</span>
                  <el-switch v-model="formData.contentPublic" size="small" />
                  <span class="pack-edit-content-public-hint">{{ formData.contentPublic ? '内容公开' : '内容私密' }}</span>
                </div>
              </div>
            </el-form-item>
            <el-form-item label="简介">
              <SplitRichTextEditor
                v-model="formData.description"
                placeholder="用于介绍这个提示词，非必填"
                class="description-rich-editor"
              />
            </el-form-item>
            <el-form-item label="内容" required>
              <div class="pack-edit-content-toolbar">
                <el-popover
                  v-model:visible="insertFieldPopoverVisible"
                  trigger="click"
                  placement="bottom-start"
                  :width="300"
                  popper-class="pack-edit-field-popover"
                >
                  <template #reference>
                    <el-button size="small" @click.stop>
                      <el-icon><MagicStick /></el-icon>
                      插入动态字段
                    </el-button>
                  </template>
                  <div class="field-insert-panel">
                    <div class="field-insert-tip">选择字段后会按 <code>${字段名}</code> 插入到当前光标位置</div>
                    <div class="field-insert-section">常用字段</div>
                    <div class="field-insert-grid">
                      <button
                        v-for="f in COMMON_FIELDS"
                        :key="f"
                        class="field-insert-item"
                        type="button"
                        @click="insertFieldToContent(f)"
                      >
                        <span class="field-insert-label">{{ f }}</span>
                        <code class="field-insert-code">${{ '{' }}{{ f }}{{ '}' }}</code>
                      </button>
                    </div>
                    <div class="field-insert-section">自定义字段</div>
                    <div class="field-insert-custom">
                      <el-input
                        v-model="customFieldName"
                        size="small"
                        placeholder="输入字段名称，例如期望字数"
                        @keyup.enter="insertCustomField"
                      />
                      <el-button size="small" type="primary" @click="insertCustomField">插入</el-button>
                    </div>
                    <div class="field-insert-hint">插入后可配合“提取字段”或右侧字段配置继续完善</div>
                  </div>
                </el-popover>
                <el-button size="small" @click="enlargeEditVisible = true">
                  <el-icon><FullScreen /></el-icon>
                  放大编辑
                </el-button>
              </div>
              <div class="pack-edit-content-warn">
                <el-icon><WarningFilled /></el-icon>
                <span>注意，某些动态字段使用者在使用的时候，可能输入内容比较长，几百几千字都有可能，这种字段占位不适合将字段放在一段话中间，建议单独成段。请结合预览测试尝试输入内容查看实际效果，保持提示词内容结构清晰。</span>
              </div>
              <el-input
                v-model="formData.content"
                type="textarea"
                :rows="12"
                maxlength="15000"
                show-word-limit
                ref="contentTextareaRef"
                placeholder="请输入提示词内容（作为AI的system层指令），使用 ${字段名} 引用变量"
                class="pack-edit-content-input"
              />
              <div v-if="unusedFields.length > 0" class="pack-edit-unused-fields">
                <el-icon><WarningFilled /></el-icon>
                <span>配置了未在模板中使用的字段：{{ unusedFields.join('、') }}，请在模板中使用 <code>${字段名}</code> 或删除这些字段配置</span>
              </div>
            </el-form-item>

            <div class="pack-edit-section">
              <div class="pack-edit-section-head">
                <span class="pack-edit-section-title">标签</span>
                <span class="pack-edit-section-tip">用于卡包内多维度筛选</span>
              </div>
              <div class="pack-edit-tags">
                <el-tag
                  v-for="(tag, i) in formSubcategories"
                  :key="i"
                  size="small"
                  closable
                  class="pack-edit-tag"
                  @close="removeFormSubcategory(i)"
                >
                  {{ tag }}
                </el-tag>
                <el-input
                  v-model="newSubcategory"
                  size="small"
                  placeholder="输入后回车添加"
                  class="pack-edit-tag-input"
                  @keyup.enter="addFormSubcategory"
                />
                <el-button size="small" type="primary" plain @click="addFormSubcategory">
                  <el-icon><Plus /></el-icon>
                </el-button>
              </div>
            </div>
          </el-form>
        </div>

        <!-- 右侧：字段配置 (1/3) -->
        <div class="pack-edit-right">
          <div class="pack-edit-right-head">
            <div class="pack-edit-right-title-wrap">
              <span class="pack-edit-right-title">字段配置</span>
              <span class="pack-edit-right-count">{{ fieldsConfig.length }} 个字段</span>
            </div>
            <div class="pack-edit-right-actions">
              <el-button size="small" plain @click="extractFieldsFromContent(formData.content)" title="从正文提取所有 ${字段名}">
                <el-icon><MagicStick /></el-icon>
                提取字段
              </el-button>
              <el-button type="primary" size="small" @click="addField">
                <el-icon><Plus /></el-icon>
                添加字段
              </el-button>
            </div>
          </div>
          <div class="pack-edit-right-tip">动态字段 <code>${字段名}</code></div>
          <div class="pack-edit-right-body">
            <!-- 空状态 -->
            <div v-if="fieldsConfig.length === 0" class="pack-edit-empty">
              <el-icon class="pack-edit-empty-icon"><Tickets /></el-icon>
              <p>暂无字段配置</p>
              <el-button type="primary" size="small" plain @click="addField">添加第一个字段</el-button>
            </div>
            <div v-else class="pack-edit-fields">
              <div
                v-for="(field, index) in fieldsConfig"
                :key="index"
                class="pack-edit-field"
                :class="{ expanded: fieldExpanded === index }"
              >
                <!-- 卡片头：点击展开/收起 -->
                <div class="pack-edit-field-head" @click="toggleFieldExpand(index)">
                  <span class="pack-edit-field-index">#{{ index + 1 }}</span>
                  <span class="pack-edit-field-name-text">{{ field.name || '未命名字段' }}</span>
                  <el-tag size="small" effect="plain" class="pack-edit-field-type-tag">{{ fieldTypeLabel(field.type) }}</el-tag>
                  <el-icon class="pack-edit-field-arrow" :class="{ rotated: fieldExpanded === index }"><ArrowDown /></el-icon>
                  <el-button
                    size="small"
                    type="danger"
                    link
                    class="pack-edit-field-head-delete"
                    title="删除字段"
                    @click.stop="removeField(index)"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <!-- 展开内容 -->
                <div v-show="fieldExpanded === index" class="pack-edit-field-body">
                  <div class="pack-edit-field-row">
                    <span class="pack-edit-field-row-label">字段名称 <em>*</em></span>
                    <el-input
                      v-model="field.name"
                      size="small"
                      placeholder="请输入字段名称"
                      @change="updateFieldName(index, field.name)"
                    />
                  </div>
                  <div class="pack-edit-field-row">
                    <span class="pack-edit-field-row-label">模板引用</span>
                    <div class="pack-edit-field-ref-row">
                      <code class="pack-edit-field-ref-code">在模板中使用 ${{ '{' }}{{ field.name }}{{ '}' }} 引用此字段</code>
                      <el-button size="small" link title="复制占位符" @click="copyFieldName(field.name)">
                        <el-icon><CopyDocument /></el-icon>
                      </el-button>
                    </div>
                  </div>
                  <div class="pack-edit-field-row">
                    <span class="pack-edit-field-row-label">字段类型</span>
                    <el-select v-model="field.type" size="small" style="width: 100%">
                      <el-option label="单行文本" value="text" />
                      <el-option label="多行文本" value="textarea" />
                      <el-option label="下拉选择" value="select" />
                    </el-select>
                  </div>
                  <div class="pack-edit-field-row">
                    <span class="pack-edit-field-row-label">字段描述</span>
                    <el-input
                      v-model="field.description"
                      size="small"
                      placeholder="请输入字段描述"
                    />
                  </div>
                  <div class="pack-edit-field-row">
                    <span class="pack-edit-field-row-label">默认值</span>
                    <el-input
                      v-model="field.defaultValue"
                      size="small"
                      placeholder="请输入默认值"
                    />
                  </div>
                  <div class="pack-edit-field-row">
                    <el-checkbox v-model="field.required">必填</el-checkbox>
                  </div>
                  <div v-if="field.type === 'select'" class="pack-edit-field-options">
                    <el-input
                      v-model="field.optionsText"
                      type="textarea"
                      :rows="2"
                      size="small"
                      placeholder="选项（每行一个）"
                      @change="updateFieldOptions(index)"
                    />
                    <el-input
                      v-model="field.optionLabelsText"
                      type="textarea"
                      :rows="2"
                      size="small"
                      placeholder="展示名（可选）"
                      @change="updateFieldOptionLabels(index)"
                    />
                  </div>
                  <div class="pack-edit-field-actions">
                    <div class="pack-edit-field-order">
                      <el-button size="small" link :disabled="index === 0" @click="moveField(index, index - 1)" title="上移">
                        <el-icon><ArrowUp /></el-icon>
                      </el-button>
                      <el-button size="small" link :disabled="index === fieldsConfig.length - 1" @click="moveField(index, index + 1)" title="下移">
                        <el-icon><ArrowDown /></el-icon>
                      </el-button>
                      <el-button size="small" type="danger" link @click="removeField(index)" title="删除">
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </div>
                    <div class="pack-edit-field-actions-right">
                      <el-button size="small" @click="toggleFieldExpand(index)">取消</el-button>
                      <el-button size="small" type="primary" @click="saveFieldExpand(index)">预览并保存</el-button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button plain @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitAndPreview">预览并保存</el-button>
      </template>
    </el-dialog>

    <!-- 放大编辑提示词内容弹窗 -->
    <el-dialog
      v-model="enlargeEditVisible"
      title="编辑提示词内容"
      width="800px"
      top="8vh"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div class="pack-enlarge-edit-toolbar">
        <el-popover
          v-model:visible="enlargeInsertPopoverVisible"
          trigger="click"
          placement="bottom-start"
          :width="300"
          popper-class="pack-edit-field-popover"
        >
          <template #reference>
            <el-button size="small" @click.stop>
              <el-icon><MagicStick /></el-icon>
              插入动态字段
            </el-button>
          </template>
          <div class="field-insert-panel">
            <div class="field-insert-tip">选择字段后会按 <code>${字段名}</code> 插入到当前光标位置</div>
            <div class="field-insert-section">常用字段</div>
            <div class="field-insert-grid">
              <button
                v-for="f in COMMON_FIELDS"
                :key="f"
                class="field-insert-item"
                type="button"
                @click="insertFieldToContent(f, enlargeTextareaRef)"
              >
                <span class="field-insert-label">{{ f }}</span>
                <code class="field-insert-code">${{ '{' }}{{ f }}{{ '}' }}</code>
              </button>
            </div>
            <div class="field-insert-section">自定义字段</div>
            <div class="field-insert-custom">
              <el-input
                v-model="customFieldName"
                size="small"
                placeholder="输入字段名称，例如期望字数"
                @keyup.enter="insertCustomField(enlargeTextareaRef)"
              />
              <el-button size="small" type="primary" @click="insertCustomField(enlargeTextareaRef)">插入</el-button>
            </div>
            <div class="field-insert-hint">插入后可配合“提取字段”或右侧字段配置继续完善</div>
          </div>
        </el-popover>
        <span class="pack-enlarge-edit-hint">使用 ${字段名} 标记可变字段，插入后会自动生成字段配置</span>
      </div>
      <el-input
        v-model="formData.content"
        type="textarea"
        :rows="20"
        ref="enlargeTextareaRef"
        placeholder="请输入提示词内容（作为AI的system层指令），使用 ${字段名} 引用变量"
      />
      <template #footer>
        <el-button @click="enlargeEditVisible = false">取消</el-button>
        <el-button type="primary" @click="enlargeEditVisible = false">确定</el-button>
      </template>
    </el-dialog>

    <!-- 提示词预览测试弹窗 -->
    <el-dialog
      v-model="previewTestVisible"
      title="提示词预览测试"
      width="860px"
      top="6vh"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div class="preview-test-tip">
        • 左侧表单：用户使用该提示词时看到的填写界面<br />
        • 右侧内容：橙色为未填写占位符，绿色为已替换内容
      </div>
      <div class="preview-test-layout">
        <div class="preview-test-left">
          <h4>表单填写 <span>{{ previewFields.length }} 个字段</span></h4>
          <div v-if="previewFields.length === 0" class="preview-test-empty">暂无字段，填写表单查看效果</div>
          <div v-for="pf in previewFields" :key="pf.name" class="preview-test-field">
            <label>{{ pf.label || pf.name }}</label>
            <el-input
              v-if="pf.type !== 'textarea'"
              v-model="pf.value"
              size="small"
              :placeholder="`请输入${pf.label || pf.name}`"
              maxlength="2000"
              show-word-limit
            />
            <el-input
              v-else
              v-model="pf.value"
              type="textarea"
              :rows="2"
              size="small"
              :placeholder="`请输入${pf.label || pf.name}`"
              maxlength="2000"
              show-word-limit
            />
            <el-select v-if="pf.type === 'select'" v-model="pf.value" size="small" style="width: 100%">
              <el-option v-for="opt in pf.options" :key="opt" :label="opt" :value="opt" />
            </el-select>
          </div>
        </div>
        <div class="preview-test-right">
          <div class="preview-test-right-head">
            <h4>预览效果</h4>
            <span class="preview-test-status" :class="unfilledFields.length > 0 ? 'partial' : 'complete'">
              {{ unfilledFields.length > 0 ? '部分未填写' : '全部已填写' }}
            </span>
            <el-button size="small" @click="copyPreviewResult">
              <el-icon><CopyDocument /></el-icon>
              复制
            </el-button>
          </div>
          <div class="preview-test-content">
            <span v-for="(seg, i) in previewSegments" :key="i" :class="seg.type">
              {{ seg.text }}
            </span>
          </div>
          <div v-if="unfilledFields.length > 0" class="preview-test-unfilled">
            以下字段尚未填写：
            <el-tag v-for="name in unfilledFields" :key="name" size="small" type="warning" class="preview-test-unfilled-tag">{{ name }}</el-tag>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="previewTestVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 生成参数设置弹窗 -->
    <el-dialog
      v-model="genParamsVisible"
      title="生成参数设置"
      width="480px"
      top="15vh"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div class="gen-params-form">
        <div class="gen-params-row">
          <span class="gen-params-label">温度 (Temperature)</span>
          <el-slider v-model="genParams.temperature" :min="0" :max="2" :step="0.1" />
          <span class="gen-params-value">{{ genParams.temperature }}</span>
        </div>
        <div class="gen-params-row">
          <span class="gen-params-label">最大Token数</span>
          <el-slider v-model="genParams.maxTokens" :min="0" :max="100000" :step="1000" />
          <span class="gen-params-value">{{ genParams.maxTokens }}</span>
        </div>
        <div class="gen-params-row">
          <span class="gen-params-label">Top-K</span>
          <el-slider v-model="genParams.topK" :min="0" :max="100" :step="1" />
          <span class="gen-params-value">{{ genParams.topK }}</span>
        </div>
        <div class="gen-params-row">
          <span class="gen-params-label">Top-P (核采样)</span>
          <el-slider v-model="genParams.topP" :min="0" :max="1" :step="0.05" />
          <span class="gen-params-value">{{ genParams.topP }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="genParamsVisible = false">取消</el-button>
        <el-button type="primary" @click="saveGenParams">保存设置</el-button>
      </template>
    </el-dialog>

    <!-- 使用说明弹窗（引导教程） -->
    <el-dialog
      v-model="usageGuideVisible"
      title="创建提示词使用说明"
      width="640px"
      top="10vh"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div class="usage-guide">
        <div class="usage-guide-step">
          <span class="usage-guide-num">1</span>
          <div>
            <h4>写模板，留空位</h4>
            <p>在提示词正文中，使用 <code>${字段名}</code> 标记需要用户填写的变量位置。</p>
            <p class="usage-guide-example">✅ 正确示范（独立成段）：<br /><code>请根据以下背景创作一个吸引人的开头：<br />【故事背景】 ${故事背景}<br />【创作要求】 - 节奏紧凑，吸引读者读完</code></p>
            <p class="usage-guide-warn">❌ 错误示范：长段落嵌在中间，背景长达千字时句子将被撑破。</p>
          </div>
        </div>
        <div class="usage-guide-step">
          <span class="usage-guide-num">2</span>
          <div>
            <h4>配置字段类型</h4>
            <p>系统会自动解析提取出的占位符，请为每个"空位"设置详细信息（类型、说明、默认值、必填），决定用户看到的表单。</p>
          </div>
        </div>
        <div class="usage-guide-step">
          <span class="usage-guide-num">3</span>
          <div>
            <h4>一键生成表单与最终效果</h4>
            <p>用户浏览提示词时，仅需填写由变量拼装出的表单，即可生成发送给 AI 的最终指令。点击「预览测试」可实时查看效果。</p>
            <p class="usage-guide-warn">⚠️ 模板中使用的 <code>${变量名}</code> 数量和名称，必须与字段配置里的卡片完全一致！</p>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="usageGuideVisible = false">我明白了</el-button>
      </template>
    </el-dialog>

    <PreviewDialog ref="previewDialogRef" />

    <!-- 密码验证对话框 -->
    <el-dialog
      v-model="passwordDialogVisible"
      title="密码验证"
      width="400px"
      :close-on-click-modal="false"
      class="password-dialog"
      append-to-body
    >
      <div class="password-dialog-content">
        <div class="password-dialog-icon">
          <el-icon :size="48" color="#e6a23c"><Lock /></el-icon>
        </div>
        <p class="password-dialog-hint">该卡片已加密，请输入密码{{ passwordAction === 'edit' ? '编辑' : passwordAction === 'preview' ? '预览' : '删除' }}</p>
        <el-input
          v-model="passwordInput"
          type="password"
          placeholder="请输入密码"
          show-password
          @keyup.enter="handlePasswordDialogConfirm"
        />
      </div>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handlePasswordDialogConfirm">确认</el-button>
      </template>
    </el-dialog>

    <!-- 设置密码对话框 -->
    <el-dialog
      v-model="setPasswordDialogVisible"
      title="设置卡片密码"
      width="400px"
      :close-on-click-modal="false"
      class="password-dialog"
      append-to-body
    >
      <div class="password-dialog-content">
        <div class="password-dialog-icon">
          <el-icon :size="48" color="#e6a23c"><Lock /></el-icon>
        </div>
        <p class="password-dialog-hint">设置密码后，该卡片将变为加密卡片，每次修改都需要输入密码</p>
        <el-input
          v-model="setPasswordInput"
          type="password"
          placeholder="请输入密码（至少4位）"
          show-password
          class="password-dialog-input"
        />
        <el-input
          v-model="setConfirmPasswordInput"
          type="password"
          placeholder="请再次输入密码"
          show-password
          @keyup.enter="handleSetPasswordConfirm"
        />
      </div>
      <template #footer>
        <el-button @click="setPasswordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSetPasswordConfirm">确认设置</el-button>
      </template>
    </el-dialog>
  </div>
</template>


<script setup lang="ts">
import { ref, onMounted, watch, computed, nextTick } from 'vue'
import { usePromptFields } from '@/composables/usePromptFields'
import type { FieldConfig } from '@/composables/usePromptFields'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, ArrowUp, ArrowDown, Delete, CopyDocument, FolderAdd, FolderOpened, Edit, Document, Rank, Download, Upload, UploadFilled, View, Switch, Lock, Reading, MoreFilled, Search, MagicStick, FullScreen, DocumentAdd, Setting, QuestionFilled, InfoFilled, WarningFilled, Tickets } from '@element-plus/icons-vue'
import { promptAPI } from '@/api'
import type { Prompt } from '@/types'
import SplitRichTextEditor from '@/components/SplitRichTextEditor.vue'
import PreviewDialog from '@/components/PreviewDialog.vue'
import BatchDeleteDialogs from '@/components/BatchDeleteDialogs.vue'
import DedupMergeDialogs from '@/components/DedupMergeDialogs.vue'
import PromptIODialogs from '@/components/PromptIODialogs.vue'
import { encodeExportData, collectInsightsByKey } from '@/utils/promptTransfer'
import { buildPromptKey, formatDate } from '@/utils/promptDedupMerge'
import type { DedupMode } from '@/utils/promptDedupMerge'
import { BOOK_ANALYSIS_PREFIX, DEFAULT_CATEGORIES } from '@/constants/promptCategories'
import { hashPassword, verifyPassword, encryptContent, decryptContent } from '@/utils/promptCrypto'

const dialogVisible = ref(false)
const isEdit = ref(false)
const router = useRouter()
const route = useRoute()
const goToPackDetail = (categoryName: string) => {
  router.push({ name: 'PackDetail', params: { name: encodeURIComponent(categoryName) } })
}
const prompts = ref<Prompt[]>([])
const formData = ref({
  id: 0,
  name: '',
  description: '',
  content: '',
  category: '默认',
  order_num: 0,
  card_type: 'normal' as 'normal' | 'encrypted',
  password: null as string | null,
  creator_name: '',
  version: '',
  // 扩展 metadata 相关（存 Prompt.metadata）
  workType: 'novel' as 'novel' | 'script',
  usageMode: 'private' as 'private' | 'public',
  contentPublic: false
})

// 生成参数（存 metadata.gen_params）
const genParams = ref({
  temperature: 1,
  maxTokens: 50000,
  topK: 40,
  topP: 0.95
})

// ============ 扩展功能状态 ============
const previewTestVisible = ref(false)
const genParamsVisible = ref(false)
const usageGuideVisible = ref(false)
const fieldExpanded = ref<number | null>(0) // 默认展开第一个字段卡片

// 预览测试：字段表单
const previewFields = ref<Array<{
  name: string
  label: string
  type: 'text' | 'textarea' | 'select'
  options: string[]
  value: string
}>>([])


const currentSessionPassword = ref('')

const passwordDialogVisible = ref(false)
const passwordDialogMode = ref<'verify' | 'set' | 'change'>('verify')
const passwordInput = ref('')
const passwordConfirmInput = ref('')
const pendingEditPrompt = ref<Prompt | null>(null)
const passwordAction = ref<'edit' | 'delete' | 'preview'>('edit')

const setPasswordDialogVisible = ref(false)
const setPasswordInput = ref('')
const setConfirmPasswordInput = ref('')

// 小分类相关状态
const formSubcategories = ref<string[]>([])
const newSubcategory = ref('')

// 卡包相关状态
const categoryDialogVisible = ref(false)
const categoryForm = ref({
  name: ''
})


const activeMainTab = ref<'all' | 'pack' | 'standalone' | 'bookAnalysis'>('pack')
const sheetVisible = ref(false)

// 导入 / 导出弹窗已拆分到 PromptIODialogs 组件（打开薄壳，调用点不变）
const ioDialogsRef = ref<InstanceType<typeof PromptIODialogs> | null>(null)
const openImportAllDialog = () => ioDialogsRef.value?.openImportAllDialog()
const openExportStandaloneDialog = () => ioDialogsRef.value?.openExportStandaloneDialog()
const openExportPackDialog = () => ioDialogsRef.value?.openExportPackDialog()
const openImportStandaloneDialog = () => ioDialogsRef.value?.openImportStandaloneDialog()
const openImportPackDialog = () => ioDialogsRef.value?.openImportPackDialog()
const openImportConvertDialog = () => ioDialogsRef.value?.openImportConvertDialog()

// 组件里导入产生的卡包名变更：改回视图的 customCategories 并落 localStorage
const addIOCustomCategory = (name: string) => {
  customCategories.value.push(name)
  saveCategoriesToStorage(customCategories.value)
}
const setIOCustomCategories = (names: string[]) => {
  customCategories.value = names
  saveCategoriesToStorage(names)
}

// 从localStorage加载卡包列表
const loadCategoriesFromStorage = (): string[] => {
  const stored = localStorage.getItem('prompt_categories')
  return stored ? JSON.parse(stored) : []
}

// 保存卡包列表到localStorage
const saveCategoriesToStorage = (categories: string[]) => {
  localStorage.setItem('prompt_categories', JSON.stringify(categories))
}


const customCategories = ref<string[]>(loadCategoriesFromStorage())

// 拖拽相关状态
const draggedPrompt = ref<Prompt | null>(null)
const isDragging = ref(false)

// ===== 卡片归属的唯一判定标准 =====
// 卡包集合 = categoryList（普通卡包）+ bookAnalysisCategoryList（拆书卡包）。
// categoryList 会为「除 未分类 / 拆书-* 之外的任何分类」生成卡包分组，
// bookAnalysisCategoryList 收纳除「拆书-未分类」外的所有拆书分类。
// 因此：只要 category 不是以下两个值，该卡片就一定已被某个卡包收纳，绝不能重复计入独立卡片。
const STANDALONE_CATEGORY = '未分类'
const STANDALONE_BOOK_CATEGORY = '拆书-未分类'

// 普通独立卡片（不含拆书-未分类，拆书侧单独统计）
const uncategorizedPrompts = computed(() => {
  return prompts.value.filter(prompt => {
    if (prompt.name === '__category_placeholder__') return false
    const category = prompt.category || ''
    // 拆书分类（含 拆书-未分类）由拆书侧负责
    if (category.startsWith('拆书-')) return false
    return category === STANDALONE_CATEGORY
  })
})

// 按分类分组的提示词列表（排除拆书分类，拆书分类在独立Tab中展示）
const categoryList = computed(() => {
  const groups: Record<string, Prompt[]> = {}
  
  // 先确保所有默认卡包都存在
  DEFAULT_CATEGORIES.forEach(cat => {
    if (!groups[cat]) {
      groups[cat] = []
    }
  })
  
  // 再确保所有自定义卡包都存在（排除拆书分类）
  customCategories.value.forEach(cat => {
    if (cat.startsWith('拆书-')) return
    if (!groups[cat]) {
      groups[cat] = []
    }
  })
  
  // 添加提示词到对应分组（排除未分类和拆书分类的提示词）
  prompts.value.forEach(prompt => {
    const category = prompt.category || '默认'
    // 跳过未分类的提示词
    if (category === '未分类') return
    // 跳过拆书分类的提示词
    if (category.startsWith('拆书-')) return
    
    if (!groups[category]) {
      groups[category] = []
    }
    // 过滤掉占位符
    if (prompt.name !== '__category_placeholder__') {
      groups[category].push(prompt)
    }
  })
  
  // 确保默认分类存在
  if (!groups['默认']) {
    groups['默认'] = []
  }
  
  return Object.entries(groups).map(([name, prompts]) => ({
    name,
    prompts: prompts.sort((a, b) => (b.created_at || '').localeCompare(a.created_at || ''))
  }))
})

// ⚠ 注意区分两个量：卡包「数量」和卡包「内的卡片数」。
// 统计栏与 Tab 角标显示的是卡包数量；totalPackCards/totalBookAnalysisCards 是卡片数，仅用于汇总总卡片数。
const totalPackCards = computed(() => {
  return categoryList.value.reduce((sum, cat) => sum + cat.prompts.length, 0)
})

// 常规卡包数量（categoryList 里的每个分组就是一个卡包，含空卡包，与「常规卡包」Tab 显示的卡包数一致）
const totalNormalPacks = computed(() => categoryList.value.length)

// ===== 拆书提示词相关 =====

// 拆书卡包列表（独立计算，不依赖 categoryList）
const bookAnalysisCategoryList = computed(() => {
  const groups: Record<string, Prompt[]> = {}

  // 收集所有拆书分类
  const bookCategories = new Set<string>()
  customCategories.value.forEach(cat => {
    if (cat.startsWith(BOOK_ANALYSIS_PREFIX)) {
      bookCategories.add(cat)
    }
  })
  prompts.value.forEach(prompt => {
    const category = prompt.category || ''
    if (category.startsWith(BOOK_ANALYSIS_PREFIX) && category !== '拆书-未分类') {
      bookCategories.add(category)
    }
  })

  // 初始化分组
  bookCategories.forEach(cat => {
    groups[cat] = []
  })

  // 添加提示词到对应分组
  prompts.value.forEach(prompt => {
    const category = prompt.category || ''
    if (category.startsWith(BOOK_ANALYSIS_PREFIX) && category !== '拆书-未分类') {
      if (!groups[category]) {
        groups[category] = []
      }
      if (prompt.name !== '__category_placeholder__') {
        groups[category].push(prompt)
      }
    }
  })

  // 转换为数组并排序
  return Object.entries(groups).map(([name, prompts]) => ({
    name,
    prompts: prompts.sort((a, b) => (b.created_at || '').localeCompare(a.created_at || ''))
  })).sort((a, b) => a.name.localeCompare(b.name))
})

// 拆书独立卡片：category 为 "拆书-未分类"，即未被任何拆书卡包收纳
const bookAnalysisUncategorizedPrompts = computed(() => {
  return prompts.value.filter(prompt => {
    if (prompt.name === '__category_placeholder__') return false
    const category = prompt.category || ''
    return category === STANDALONE_BOOK_CATEGORY
  })
})

// 合并的独立卡片（普通独立卡片 + 拆书独立卡片）
const combinedUncategorizedPrompts = computed(() => {
  return [...uncategorizedPrompts.value, ...bookAnalysisUncategorizedPrompts.value]
})

// ===== 全局搜索：跨 卡包/独立卡片/拆书 三个 Tab 过滤 =====
const searchKeyword = ref('')

const filterByKeyword = (name: string, prompts: { name?: string; content?: string }[]) => {
  const kw = searchKeyword.value.trim().toLowerCase()
  if (!kw) return true
  return name.toLowerCase().includes(kw) ||
    prompts.some(p => (p.name || '').toLowerCase().includes(kw) || (p.content || '').toLowerCase().includes(kw))
}

const filteredCategoryList = computed(() => {
  return categoryList.value.filter(cat => filterByKeyword(cat.name, cat.prompts))
})

const filteredBookAnalysisCategoryList = computed(() => {
  return bookAnalysisCategoryList.value.filter(cat => filterByKeyword(cat.name, cat.prompts))
})

const filteredStandalonePrompts = computed(() => {
  const kw = searchKeyword.value.trim().toLowerCase()
  if (!kw) return combinedUncategorizedPrompts.value
  return combinedUncategorizedPrompts.value.filter(p =>
    (p.name || '').toLowerCase().includes(kw) || (p.content || '').toLowerCase().includes(kw)
  )
})

// 导出卡包弹窗的合并列表（普通卡包 + 拆书卡包），用于统一导出
const allExportablePacks = computed(() => {
  const normal = categoryList.value
    .filter(cat => cat.name !== '未分类')
    .map(cat => ({ name: cat.name, prompts: cat.prompts, isBookAnalysis: false }))
  const book = bookAnalysisCategoryList.value.map(cat => ({
    name: cat.name,
    prompts: cat.prompts,
    isBookAnalysis: true
  }))
  return [...normal, ...book]
})

// 拆书卡包总卡片数（不含拆书独立卡片，已并入独立卡片Tab）
const totalBookAnalysisCards = computed(() => {
  return bookAnalysisCategoryList.value.reduce((sum, cat) => sum + cat.prompts.length, 0)
})

// 拆书卡包数量（与「拆书卡包」Tab 显示的卡包数一致）
const totalBookAnalysisPacks = computed(() => bookAnalysisCategoryList.value.length)

// 全部卡片数 = 常规卡包内卡片 + 拆书卡包内卡片 + 独立卡片
// 三者互斥且完备：卡包分组排除了「未分类」「拆书-未分类」，而独立卡片只取这两个分类，
// 因此这里相加就是真实总数，不会把已入包的卡片重复计入独立卡片。
const totalPromptCards = computed(() => {
  return totalPackCards.value + totalBookAnalysisCards.value + combinedUncategorizedPrompts.value.length
})

// 拆书分类选项（用于创建/编辑时的分类下拉）
const bookAnalysisCategoryOptions = computed(() => {
  return bookAnalysisCategoryList.value.map(cat => cat.name).concat(['拆书-未分类'])
})

// 拆书Tab下创建卡包
const bookAnalysisCategoryDialogVisible = ref(false)
const bookAnalysisCategoryForm = ref({ name: '' })

const openBookAnalysisCategoryDialog = () => {
  bookAnalysisCategoryForm.value.name = ''
  bookAnalysisCategoryDialogVisible.value = true
}

const handleCreateBookAnalysisCategory = () => {
  const rawName = bookAnalysisCategoryForm.value.name.trim()
  if (!rawName) {
    ElMessage.warning('请输入卡包名称')
    return
  }
  // 自动加前缀
  const fullName = rawName.startsWith(BOOK_ANALYSIS_PREFIX) ? rawName : BOOK_ANALYSIS_PREFIX + rawName

  if (DEFAULT_CATEGORIES.includes(fullName)) {
    ElMessage.warning('该名称为默认卡包，不可使用')
    return
  }

  const exists = customCategories.value.includes(fullName) ||
                 categoryList.value.some(c => c.name === fullName)
  if (exists) {
    ElMessage.warning('该卡包名称已存在')
    return
  }

  customCategories.value.push(fullName)
  saveCategoriesToStorage(customCategories.value)

  ElMessage.success('拆书卡包创建成功')
  bookAnalysisCategoryDialogVisible.value = false
}

// 拆书Tab下创建提示词
const handleCreateBookAnalysisPrompt = (category?: string) => {
  isEdit.value = false
  currentSessionPassword.value = ''
  formData.value = {
    id: 0,
    name: '',
    description: '',
    content: '',
    category: category || (bookAnalysisCategoryOptions.value[0] || '拆书-默认'),
    order_num: 0,
    card_type: 'normal',
    password: null,
    creator_name: '',
    version: '',
    workType: 'novel',
    usageMode: 'private',
    contentPublic: false
  }
  genParams.value = { temperature: 1, maxTokens: 50000, topK: 40, topP: 0.95 }
  fieldsConfig.value = []
  formSubcategories.value = []
  newSubcategory.value = ''
  fieldExpanded.value = 0
  dialogVisible.value = true
}

// 处理拆书卡包操作下拉菜单命令
const handleBookAnalysisPackAction = (command: string, categoryName: string, promptsCount: number) => {
  switch (command) {
    case 'add':
      handleCreateBookAnalysisPrompt(categoryName)
      break
    case 'deletePack':
      handleDeleteBookAnalysisCategory(categoryName)
      break
  }
}

// 拆书Tab下删除卡包
const handleDeleteBookAnalysisCategory = async (categoryName: string) => {
  try {
    await ElMessageBox.confirm(
      `确定删除卡包"${categoryName}"吗？该卡包下的所有提示词将被移动到"拆书-未分类"。`,
      '提示',
      { type: 'warning' }
    )

    const promptsToUpdate = prompts.value.filter(p => p.category === categoryName)
    for (const prompt of promptsToUpdate) {
      await promptAPI.update(prompt.id, { ...prompt, category: '拆书-未分类' })
    }

    const index = customCategories.value.indexOf(categoryName)
    if (index > -1) {
      customCategories.value.splice(index, 1)
      saveCategoriesToStorage(customCategories.value)
    }

    await fetchPrompts()
    ElMessage.success('卡包删除成功')
  } catch {
    // 取消删除
  }
}

// 拆书Tab下拖拽放置
const onBookAnalysisDrop = async (targetCategory: string, event: DragEvent) => {
  event.preventDefault()
  if (!draggedPrompt.value) return

  const prompt = draggedPrompt.value
  const oldCategory = prompt.category || '默认'
  if (targetCategory === oldCategory) return

  try {
    await promptAPI.update(prompt.id, { ...prompt, category: targetCategory })
    await fetchPrompts()
    ElMessage.success(`已移动到"${targetCategory}"`)
  } catch (error) {
    ElMessage.error('移动失败')
  }

  draggedPrompt.value = null
  isDragging.value = false
}

const categoryOptions = computed(() => {
  const names = new Set<string>(DEFAULT_CATEGORIES)

  if (formData.value.category) {
    names.add(formData.value.category)
  }

  customCategories.value.forEach(name => {
    if (name && name !== '未分类') {
      names.add(name)
    }
  })

  categoryList.value.forEach(category => {
    if (category.name && category.name !== '未分类') {
      names.add(category.name)
    }
  })

  return Array.from(names)
})

// 拖拽开始
const onDragStart = (prompt: Prompt, event: DragEvent) => {
  draggedPrompt.value = prompt
  isDragging.value = true
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', prompt.id.toString())
  }
}

// 拖拽结束
const onDragEnd = () => {
  draggedPrompt.value = null
  isDragging.value = false
  // 移除所有拖拽高亮样式
  document.querySelectorAll('.drag-over').forEach(el => el.classList.remove('drag-over'))
}

// 拖拽经过
const onDragOver = (event: DragEvent) => {
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move'
  }
}

// 放置
const onDrop = async (targetCategory: string, event: DragEvent) => {
  event.preventDefault()
  
  if (!draggedPrompt.value) return
  
  const prompt = draggedPrompt.value
  const oldCategory = prompt.category || '默认'
  
  // 如果目标分类和当前分类相同，不做处理
  if (targetCategory === oldCategory) return
  
  // 更新提示词的分类
  try {
    await promptAPI.update(prompt.id, { ...prompt, category: targetCategory })
    await fetchPrompts()
    ElMessage.success(`已移动到"${targetCategory}"`)
  } catch (error) {
    ElMessage.error('移动失败')
  }
  
  draggedPrompt.value = null
  isDragging.value = false
}


// 打开创建卡包对话框
const openCategoryDialog = () => {
  categoryForm.value.name = ''
  categoryDialogVisible.value = true
}

// 创建卡包
const handleCreateCategory = () => {
  if (!categoryForm.value.name.trim()) {
    ElMessage.warning('请输入卡包名称')
    return
  }
  
  const categoryName = categoryForm.value.name.trim()
  
  if (DEFAULT_CATEGORIES.includes(categoryName)) {
    ElMessage.warning('该名称为默认卡包，不可使用')
    return
  }
  
  // 检查是否已存在
  const exists = customCategories.value.includes(categoryName) || 
                 categoryList.value.some(c => c.name === categoryName)
  if (exists) {
    ElMessage.warning('该卡包名称已存在')
    return
  }
  
  // 添加到自定义卡包列表并保存
  customCategories.value.push(categoryName)
  saveCategoriesToStorage(customCategories.value)
  
  ElMessage.success('卡包创建成功')
  categoryDialogVisible.value = false
}

// 处理卡包操作下拉菜单命令
const handlePackAction = (command: string, categoryName: string, promptsCount: number) => {
  switch (command) {
    case 'add':
      handleCreateInCategory(categoryName)
      break
    case 'batchDelete':
      if (promptsCount > 0 && !DEFAULT_CATEGORIES.includes(categoryName)) {
        openBatchDeleteDialog(categoryName)
      }
      break
    case 'deletePack':
      if (!DEFAULT_CATEGORIES.includes(categoryName)) {
        handleDeleteCategory(categoryName)
      }
      break
  }
}

// 删除卡包
const handleDeleteCategory = async (categoryName: string) => {
  if (DEFAULT_CATEGORIES.includes(categoryName)) {
    ElMessage.warning('默认卡包不可删除')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定删除卡包"${categoryName}"吗？该卡包下的所有提示词将被移动到"默认"分类。`,
      '提示',
      { type: 'warning' }
    )
    
    // 将该分类下的所有提示词移动到默认分类
    const promptsToUpdate = prompts.value.filter(p => p.category === categoryName)
    for (const prompt of promptsToUpdate) {
      await promptAPI.update(prompt.id, { ...prompt, category: '默认' })
    }
    
    // 从自定义卡包列表中移除
    const index = customCategories.value.indexOf(categoryName)
    if (index > -1) {
      customCategories.value.splice(index, 1)
      saveCategoriesToStorage(customCategories.value)
    }
    
    await fetchPrompts()
    ElMessage.success('卡包删除成功')
  } catch (error) {
    // 取消删除
  }
}

// 在指定分类中创建提示词
const handleCreateInCategory = (categoryName: string) => {
  isEdit.value = false
  formData.value = {
    id: 0,
    name: '',
    description: '',
    content: '',
    category: categoryName,
    order_num: 0,
    card_type: 'normal',
    password: null,
    creator_name: '',
    version: '',
    workType: 'novel',
    usageMode: 'private',
    contentPublic: false
  }
  genParams.value = { temperature: 1, maxTokens: 50000, topK: 40, topP: 0.95 }
  fieldsConfig.value = []
  // 重置小分类
  formSubcategories.value = []
  newSubcategory.value = ''
  dialogVisible.value = true
}

// 字段配置相关状态
const fieldsConfig = ref<FieldConfig[]>([])

// ============ 插入动态字段 ============
// 常用字段预置
const COMMON_FIELDS = ['期望字数', '参考文风', '角色设定', '背景设定', '剧情方向', '输出格式']

const insertFieldPopoverVisible = ref(false)
const enlargeInsertPopoverVisible = ref(false)
const customFieldName = ref('')
const enlargeEditVisible = ref(false)
const contentTextareaRef = ref<{ $el: HTMLDivElement } | null>(null)
const enlargeTextareaRef = ref<{ $el: HTMLDivElement } | null>(null)

// 动态字段与内容编辑逻辑（状态仍由本视图持有，供模板绑定；实现见 usePromptFields）
const {
  fieldTypeLabel,
  toggleFieldExpand,
  saveFieldExpand,
  insertFieldToContent,
  insertCustomField,
  handlePasteImport,
  openPreviewTest,
  copyPreviewResult,
  addFormSubcategory,
  removeFormSubcategory,
  extractFieldsFromContent,
  addField,
  removeField,
  validateFieldsConfig,
  previewSegments,
  unfilledFields,
  unusedFields,
} = usePromptFields({
  formData,
  fieldsConfig,
  fieldExpanded,
  previewFields,
  previewTestVisible,
  insertFieldPopoverVisible,
  enlargeInsertPopoverVisible,
  customFieldName,
  contentTextareaRef,
  formSubcategories,
  newSubcategory,
})

// ============ 顶部操作栏功能 ============

// 保存并打开预览测试（"预览并保存"）
const handleSubmitAndPreview = async () => {
  await handleSubmit()
  if (!dialogVisible.value) {
    // 保存成功（弹窗已关闭）后打开预览测试
    openPreviewTest()
  }
}

// 保存生成参数
const saveGenParams = () => {
  genParamsVisible.value = false
  ElMessage.success('生成参数已保存')
}

// 使用模式切换
const handleUsageModeChange = () => {
  // 公用模式默认公开内容
  if (formData.value.usageMode === 'public' && !formData.value.contentPublic) {
    formData.value.contentPublic = true
  }
}

// 预览弹窗（已拆分到 PreviewDialog 组件）
const previewDialogRef = ref<InstanceType<typeof PreviewDialog> | null>(null)

// 预览提示词
const handlePreview = (prompt: Prompt) => {
  if (prompt.card_type === 'encrypted') {
    if (!prompt.password) {
      ElMessage.warning('此加密卡片需要先设置密码，请编辑卡片设置密码')
      return
    }
    pendingEditPrompt.value = prompt
    passwordAction.value = 'preview'
    passwordDialogMode.value = 'verify'
    passwordInput.value = ''
    passwordConfirmInput.value = ''
    passwordDialogVisible.value = true
    return
  }
  previewDialogRef.value?.open(prompt.content)
}

onMounted(async () => {
  await fetchPrompts()
  // 自动迁移旧的拆书库 localStorage 数据
  await migrateOldBookAnalysisData()
  // 处理路由参数，自动切换到拆书Tab
  if (route.query.tab === 'bookAnalysis') {
    activeMainTab.value = 'bookAnalysis'
  }
})

// 自动迁移旧的拆书库 localStorage 数据到数据库
const migrateOldBookAnalysisData = async () => {
  const OLD_PROMPTS_KEY = 'book-analysis-prompts_prompts'
  const OLD_CATEGORIES_KEY = 'book-analysis-prompts_categories'
  const MIGRATION_FLAG_KEY = 'book-analysis-migration-done'

  // 检查是否已迁移
  if (localStorage.getItem(MIGRATION_FLAG_KEY) === 'true') {
    return
  }

  // 检查是否有旧数据
  const oldPromptsData = localStorage.getItem(OLD_PROMPTS_KEY)
  if (!oldPromptsData) {
    // 没有旧数据，标记为已迁移
    localStorage.setItem(MIGRATION_FLAG_KEY, 'true')
    return
  }

  try {
    const oldPrompts = JSON.parse(oldPromptsData)
    if (!Array.isArray(oldPrompts) || oldPrompts.length === 0) {
      localStorage.setItem(MIGRATION_FLAG_KEY, 'true')
      localStorage.removeItem(OLD_PROMPTS_KEY)
      localStorage.removeItem(OLD_CATEGORIES_KEY)
      return
    }

    // 过滤有效提示词
    const validPrompts = oldPrompts.filter((p: any) => p && typeof p.name === 'string' && typeof p.content === 'string')

    if (validPrompts.length === 0) {
      localStorage.setItem(MIGRATION_FLAG_KEY, 'true')
      localStorage.removeItem(OLD_PROMPTS_KEY)
      localStorage.removeItem(OLD_CATEGORIES_KEY)
      return
    }

    // 显示迁移提示
    ElMessage.info(`正在迁移 ${validPrompts.length} 条拆书库提示词...`)

    let successCount = 0
    let failCount = 0
    const migratedCategories = new Set<string>()

    for (const prompt of validPrompts) {
      try {
        // 处理分类名：自动加"拆书-"前缀
        let category = prompt.category || '未分类'
        if (category === '未分类') {
          category = '拆书-未分类'
        } else if (!category.startsWith('拆书-')) {
          category = '拆书-' + category
        }

        migratedCategories.add(category)

        const res = await promptAPI.create({
          name: prompt.name,
          content: prompt.content,
          category: category,
          order_num: prompt.order_num || 0,
          created_at: prompt.created_at || new Date().toISOString(),
          fields: prompt.fields || [],
          card_type: prompt.card_type || 'normal',
          password: prompt.password || null,
          description: prompt.description || '',
          creator_name: prompt.creator_name || '',
          version: prompt.version || ''
        })

        if (res.success) {
          successCount++
        } else {
          failCount++
        }
      } catch (error) {
        failCount++
      }
    }

    // 创建迁移后的分类
    for (const cat of migratedCategories) {
      if (cat !== '拆书-未分类' && !customCategories.value.includes(cat)) {
        customCategories.value.push(cat)
      }
    }
    saveCategoriesToStorage(customCategories.value)

    // 清除旧数据
    localStorage.removeItem(OLD_PROMPTS_KEY)
    localStorage.removeItem(OLD_CATEGORIES_KEY)
    localStorage.setItem(MIGRATION_FLAG_KEY, 'true')

    // 刷新数据
    await fetchPrompts()

    if (failCount === 0) {
      ElMessage.success(`迁移完成！已迁移 ${successCount} 条拆书库提示词到拆书提示词Tab`)
    } else {
      ElMessage.warning(`迁移完成：成功 ${successCount} 条，失败 ${failCount} 条`)
    }
  } catch (error) {
    console.error('迁移失败:', error)
    ElMessage.error('迁移失败，请手动导出导入')
  }
}

const fetchPrompts = async () => {
  const res = await promptAPI.getAll()
  if (res.success && res.data) {
    prompts.value = res.data
  }
}

const handleCreate = () => {
  isEdit.value = false
  currentSessionPassword.value = ''
  formData.value = {
    id: 0,
    name: '',
    description: '',
    content: '',
    category: categoryOptions.value[0] || '默认',
    order_num: 0,
    card_type: 'normal',
    password: null,
    creator_name: '',
    version: '',
    workType: 'novel',
    usageMode: 'private',
    contentPublic: false
  }
  genParams.value = { temperature: 1, maxTokens: 50000, topK: 40, topP: 0.95 }
  fieldsConfig.value = []
  formSubcategories.value = []
  newSubcategory.value = ''
  fieldExpanded.value = 0
  dialogVisible.value = true
}

const handleEdit = (prompt: Prompt) => {
  if (prompt.card_type === 'encrypted' && prompt.password) {
    pendingEditPrompt.value = prompt
    passwordAction.value = 'edit'
    passwordDialogMode.value = 'verify'
    passwordInput.value = ''
    passwordConfirmInput.value = ''
    passwordDialogVisible.value = true
    return
  }
  openEditDialog(prompt)
}

const openEditDialog = (prompt: Prompt) => {
  isEdit.value = true
  const meta = prompt.metadata || {}
  formData.value = {
    id: prompt.id,
    name: prompt.name,
    description: prompt.description || '',
    content: prompt.content,
    category: prompt.category,
    order_num: prompt.order_num,
    card_type: prompt.card_type || 'normal',
    password: prompt.password || null,
    creator_name: prompt.creator_name || '',
    version: prompt.version || '',
    workType: meta.work_type === 'script' ? 'script' : 'novel',
    usageMode: meta.usage_mode === 'public' ? 'public' : 'private',
    contentPublic: !!meta.content_public
  }
  // 生成参数
  const gp = meta.gen_params || {}
  genParams.value = {
    temperature: typeof gp.temperature === 'number' ? gp.temperature : 1,
    maxTokens: typeof gp.maxTokens === 'number' ? gp.maxTokens : 50000,
    topK: typeof gp.topK === 'number' ? gp.topK : 40,
    topP: typeof gp.topP === 'number' ? gp.topP : 0.95
  }
  // 重置字段配置
  fieldsConfig.value = []
  fieldExpanded.value = 0
  
  // 如果有已保存的字段配置，优先使用
  if (prompt.fields && prompt.fields.length > 0) {
    fieldsConfig.value = prompt.fields.map(field => ({
      name: field.name,
      label: field.label,
      type: field.type,
      options: field.options || [],
      optionLabels: field.optionLabels || [],
      optionsText: (field.options || []).join('\n'),
      optionLabelsText: (field.optionLabels || []).join('\n'),
      description: field.description || '',
      required: field.required !== undefined ? field.required : true,
      defaultValue: field.defaultValue || ''
    }))
  } else {
    // 否则从提示词内容中提取字段名称
    extractFieldsFromContent(prompt.content)
  }
  
  // 初始化小分类
  formSubcategories.value = [...(prompt.subcategories || [])]
  newSubcategory.value = ''
  
  dialogVisible.value = true
}

const handlePasswordDialogConfirm = async () => {
  if (passwordDialogMode.value === 'verify') {
    if (!pendingEditPrompt.value) return
    const isValid = await verifyPassword(passwordInput.value, pendingEditPrompt.value.password || '')
    if (isValid) {
      const promptToProcess = pendingEditPrompt.value
      const action = passwordAction.value
      const rawPassword = passwordInput.value
      passwordDialogVisible.value = false
      passwordInput.value = ''
      pendingEditPrompt.value = null
      if (action === 'edit') {
        currentSessionPassword.value = rawPassword
        const decrypted = await decryptContent(promptToProcess.content, rawPassword)
        if (decrypted !== null) {
          promptToProcess.content = decrypted
        }
        openEditDialog(promptToProcess)
      } else if (action === 'delete') {
        executeDelete(promptToProcess)
      } else if (action === 'preview') {
        const decrypted = await decryptContent(promptToProcess.content, rawPassword)
        previewDialogRef.value?.open(decrypted !== null ? decrypted : promptToProcess.content)
      }
    } else {
      ElMessage.error('密码错误')
    }
  }
}

const handleSetPasswordConfirm = async () => {
  if (!setPasswordInput.value.trim()) {
    ElMessage.warning('请输入密码')
    return
  }
  if (setPasswordInput.value !== setConfirmPasswordInput.value) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }
  if (setPasswordInput.value.trim().length < 4) {
    ElMessage.warning('密码长度至少4位')
    return
  }
  const hashed = await hashPassword(setPasswordInput.value.trim())
  formData.value.card_type = 'encrypted'
  formData.value.password = hashed
  currentSessionPassword.value = setPasswordInput.value.trim()
  setPasswordDialogVisible.value = false
  setPasswordInput.value = ''
  setConfirmPasswordInput.value = ''
  ElMessage.success('密码设置成功，卡片已切换为加密卡片')
}

const openSetPasswordDialog = () => {
  setPasswordInput.value = ''
  setConfirmPasswordInput.value = ''
  setPasswordDialogVisible.value = true
}

const removeCardPassword = () => {
  formData.value.card_type = 'normal'
  formData.value.password = null
  currentSessionPassword.value = ''
  ElMessage.success('已移除密码，卡片已切换为普通卡片')
}

const handleCardTypeChange = (val: 'normal' | 'encrypted') => {
  if (val === 'normal') {
    formData.value.password = null
    currentSessionPassword.value = ''
  }
}

const executeDelete = async (prompt: Prompt) => {
  try {
    await ElMessageBox.confirm(`确定删除提示词"${prompt.name}"吗？`, '提示', {
      type: 'warning'
    })
    const res = await promptAPI.delete(prompt.id)
    if (res.success) {
      ElMessage.success('删除成功')
      await fetchPrompts()
    }
  } catch (error) {
    // 取消删除
  }
}

// 调整字段顺序
const moveField = (fromIndex: number, toIndex: number) => {
  if (fromIndex < 0 || fromIndex >= fieldsConfig.value.length || toIndex < 0 || toIndex >= fieldsConfig.value.length) return
  
  const [movedField] = fieldsConfig.value.splice(fromIndex, 1)
  if (movedField) {
    fieldsConfig.value.splice(toIndex, 0, movedField)
  }
}

// 更新字段选项
const updateFieldOptions = (index: number) => {
  const field = fieldsConfig.value[index]
  if (field) {
    const optionsText = field.optionsText
    field.options = optionsText.split('\n').map((option: string) => option.trim()).filter(Boolean)
    field.optionLabels = field.optionLabels.slice(0, field.options.length)
    field.optionLabelsText = field.optionLabels.join('\n')
  }
}

const updateFieldOptionLabels = (index: number) => {
  const field = fieldsConfig.value[index]
  if (field) {
    const labels = field.optionLabelsText
      .split('\n')
      .map((label: string) => label.trim())
      .slice(0, field.options.length)

    while (labels.length > 0 && !labels[labels.length - 1]) {
      labels.pop()
    }

    field.optionLabels = labels
    field.optionLabelsText = labels.join('\n')
  }
}

// 更新字段名称，同步修改提示词内容中的占位符
const updateFieldName = (index: number, newName: string) => {
  if (!newName.trim()) {
    ElMessage.warning('字段名称不能为空')
    return
  }
  
  const field = fieldsConfig.value[index]
  if (!field) return
  
  // 获取旧字段名称
  const oldName = field.name
  
  // 检查新字段名称是否与其他字段重复
  const existingNames = fieldsConfig.value
    .filter((_, i) => i !== index)
    .map(f => f.name)
  
  if (existingNames.includes(newName)) {
    ElMessage.warning('字段名称不能重复')
    return
  }
  
  // 更新字段名称
  field.name = newName
  
  // 更新提示词内容中的占位符
  if (oldName && oldName !== newName) {
    formData.value.content = formData.value.content.replace(
      new RegExp(`\$\{${oldName}\}`, 'g'),
      `\$\{${newName}\}`
    )
  }
}

// 复制字段名称的占位符形式到剪贴板
const copyFieldName = (fieldName: string) => {
  if (!fieldName.trim()) {
    ElMessage.warning('字段名称不能为空')
    return
  }
  
  const placeholder = `\$\{${fieldName}\}`
  navigator.clipboard.writeText(placeholder)
  ElMessage.success('占位符已复制到剪贴板')
}

// 监听提示词内容变化，自动更新字段配置
watch(() => formData.value.content, (newContent) => {
  if (newContent) {
    // 提取新的字段列表
    const fieldRegex = /\$\{([^}]+)\}/g
    const newFields = new Set<string>()
    let match
    while ((match = fieldRegex.exec(newContent)) !== null) {
      if (match && match[1]) {
        newFields.add(match[1].trim())
      }
    }
    
    // 收集现有字段名称
    const existingFieldNames = new Set(fieldsConfig.value.map(field => field.name))
    
    // 添加新字段（只添加新出现的字段，不删除现有字段）
    newFields.forEach(fieldName => {
      if (!existingFieldNames.has(fieldName)) {
        fieldsConfig.value.push({
          name: fieldName,
          label: fieldName,
          type: 'text',
          options: [],
          optionLabels: [],
          optionsText: '',
          optionLabelsText: '',
          description: '',
          required: true,
          defaultValue: ''
        })
      }
    })
    // 自动展开最后一个新字段卡片（如有新字段）
    if (newFields.size > 0) {
      fieldExpanded.value = fieldsConfig.value.length - 1
    }
  } else {
    // 内容为空时清空字段配置
    fieldsConfig.value = []
  }
}, { deep: true })

const handleDelete = async (prompt: Prompt) => {
  if (prompt.card_type === 'encrypted' && prompt.password) {
    pendingEditPrompt.value = prompt
    passwordAction.value = 'delete'
    passwordDialogMode.value = 'verify'
    passwordInput.value = ''
    passwordConfirmInput.value = ''
    passwordDialogVisible.value = true
    return
  }
  executeDelete(prompt)
}

const handleSubmit = async () => {
  if (!formData.value.name || !formData.value.category || !formData.value.content) {
    ElMessage.warning('请填写完整信息')
    return
  }

  if (formData.value.card_type === 'encrypted' && !formData.value.password) {
    ElMessage.warning('加密卡片必须设置密码')
    return
  }

  // 验证字段配置
  const validationError = validateFieldsConfig()
  if (validationError) {
    ElMessage.error(validationError)
    return
  }

  try {
    let saveContent = formData.value.content
    if (formData.value.card_type === 'encrypted' && currentSessionPassword.value) {
      saveContent = await encryptContent(saveContent, currentSessionPassword.value)
    }

    const dataToSave = {
      ...formData.value,
      content: saveContent,
      fields: fieldsConfig.value.map(field => ({
        name: field.name,
        label: field.label,
        type: field.type,
        options: field.options,
        optionLabels: field.optionLabels,
        description: field.description,
        required: field.required,
        defaultValue: field.defaultValue || ''
      })),
      subcategories: formSubcategories.value,
      card_type: formData.value.card_type || 'normal',
      password: formData.value.password || null,
      // 扩展信息（存 metadata，不修改后端 schema）
      metadata: {
        work_type: formData.value.workType,
        usage_mode: formData.value.usageMode,
        content_public: formData.value.contentPublic,
        gen_params: {
          temperature: genParams.value.temperature,
          maxTokens: genParams.value.maxTokens,
          topK: genParams.value.topK,
          topP: genParams.value.topP
        }
      }
    }
    
    if (isEdit.value) {
      const res = await promptAPI.update(formData.value.id, dataToSave)
      if (res.success) {
        ElMessage.success('更新成功')
      }
    } else {
      const res = await promptAPI.create(dataToSave)
      if (res.success) {
        ElMessage.success('创建成功')
      }
    }
    dialogVisible.value = false
    currentSessionPassword.value = ''
    await fetchPrompts()
  } catch (error) {
  }
}

// 打开批量删除对话框（委托给 BatchDeleteDialogs 组件）
const batchDeleteDialogsRef = ref<InstanceType<typeof BatchDeleteDialogs> | null>(null)
const openBatchDeleteDialog = (categoryName: string) => {
  batchDeleteDialogsRef.value?.open(categoryName)
}

// 打开独立卡片批量删除对话框（委托给 BatchDeleteDialogs 组件）
const openStandaloneBatchDeleteDialog = () => {
  batchDeleteDialogsRef.value?.openStandalone()
}






// 查重 / 合并弹窗已拆分到 DedupMergeDialogs 组件（打开薄壳，调用点不变）
const dedupMergeRef = ref<InstanceType<typeof DedupMergeDialogs> | null>(null)
const openDedupDialog = () => dedupMergeRef.value?.openDedup()
const openMergeDialog = () => dedupMergeRef.value?.openMerge()

// 组件里合并产生的卡包名变更：改回视图的 customCategories 并落 localStorage
const addDedupCustomCategory = (name: string) => {
  customCategories.value.push(name)
  saveCategoriesToStorage(customCategories.value)
}
const removeDedupCustomCategories = (names: string[]) => {
  const s = new Set(names)
  customCategories.value = customCategories.value.filter(n => !s.has(n))
  saveCategoriesToStorage(customCategories.value)
}







// 执行整体导出
const executeExportAll = async () => {
  const insights = await collectInsightsByKey(prompts.value)

  // 收集所有独立卡片（包含普通独立卡片和拆书独立卡片）
  const standalone = combinedUncategorizedPrompts.value.map(p => ({
    name: p.name,
    content: p.content,
    category: p.category,
    order_num: p.order_num,
    created_at: p.created_at,
    fields: p.fields,
    subcategories: p.subcategories,
    card_type: p.card_type || 'normal',
    password: p.password || null,
  }))

  // 收集所有卡包（包含普通卡包和拆书卡包）
  const packs: Record<string, any[]> = {}
  categoryList.value.forEach(category => {
    if (category.name !== '未分类') {
      packs[category.name] = category.prompts.map(p => ({
        name: p.name,
        content: p.content,
        category: p.category,
        order_num: p.order_num,
        created_at: p.created_at,
        fields: p.fields,
        subcategories: p.subcategories,
        card_type: p.card_type || 'normal',
        password: p.password || null,
      }))
    }
  })
  
  // 收集拆书卡包
  bookAnalysisCategoryList.value.forEach(category => {
    packs[category.name] = category.prompts.map(p => ({
      name: p.name,
      content: p.content,
      category: p.category,
      order_num: p.order_num,
      created_at: p.created_at,
      fields: p.fields,
      subcategories: p.subcategories,
      card_type: p.card_type || 'normal',
      password: p.password || null,
    }))
  })
  
  const exportData = {
    type: 'full-backup',
    version: '1.0',
    exportTime: new Date().toISOString(),
    data: {
      standalone,
      packs,
      insights
    }
  }
  
  const blob = new Blob([encodeExportData(JSON.stringify(exportData))], { type: 'application/octet-stream' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `提示词整体备份_${new Date().toLocaleDateString('zh-CN').replace(/\//g, '-')}.json`
  a.click()
  URL.revokeObjectURL(url)
  
  ElMessage.success(`成功导出整体备份，包含 ${standalone.length} 张独立卡片和 ${Object.keys(packs).length} 个卡包`)
}















</script>

<style scoped>
.no-search-match {
  padding: 24px 0;
  text-align: center;
}

.prompts-container {
  background: #f5f7fa;
  padding: 28px 36px 40px;
  min-height: 100%;
  animation: fadeIn 0.4s ease;
}

/* 空独立卡片提示 */
.empty-standalone-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px;
  background: linear-gradient(180deg, #fff7e6 0%, #ffffff 100%);
  border-radius: 12px;
  border: 2px dashed #ffd591;
  color: #fa8c16;
  cursor: pointer;
  transition: all 0.3s ease;
  gap: 12px;
}

.empty-standalone-hint:hover {
  border-color: #fa8c16;
  border-style: solid;
  background: linear-gradient(180deg, #ffe7ba 0%, #ffffff 100%);
}

.empty-standalone-hint.drag-over {
  border-color: #52c41a;
  border-style: solid;
  background: linear-gradient(180deg, #d9f7be 0%, #ffffff 100%);
  color: #52c41a;
}

.empty-standalone-hint .el-icon {
  font-size: 32px;
}

/* 字段配置样式 - 玻璃拟态主题 */
.description-rich-editor {
  width: 100%;
}
.description-rich-editor :deep(.vditor) {
  border-radius: 8px;
  border-color: #d1d5db;
}
.description-rich-editor :deep(.vditor-toolbar) {
  border-radius: 8px 8px 0 0;
  background: #f8fafc;
}
.description-rich-editor :deep(.vditor-content) {
  border-radius: 0 0 8px 8px;
}
:root[data-theme='dark'] .description-rich-editor :deep(.vditor) {
  border-color: #333c56;
  background: #141a29;
}
:root[data-theme='dark'] .description-rich-editor :deep(.vditor-toolbar) {
  background: #1e2434;
  border-color: #333c56;
}
:root[data-theme='dark'] .description-rich-editor :deep(.vditor-content) {
  background: #141a29;
}
:root[data-theme='dark'] .description-rich-editor :deep(.vditor-ir pre) {
  color: #e5e7eb;
}


.password-dialog-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 10px 0;
}

.password-dialog-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: rgba(230, 162, 60, 0.1);
}

.password-dialog-hint {
  text-align: center;
  color: #606266;
  font-size: 14px;
  line-height: 1.6;
  margin: 0;
}

.password-dialog-content .el-input {
  width: 100%;
}

.password-dialog-input {
  margin-bottom: 12px;
}

/* --- 暗色主题 - 网格模式 --- */
:root[data-theme='dark'] .prompts-container {
  background: #0f172a;
}


:root[data-theme='dark'] .password-dialog-icon {
  background: rgba(251, 191, 36, 0.15);
}

:root[data-theme='dark'] .password-dialog-hint {
  color: #9ca3af;
}

/* ---- 编辑弹窗左右布局 ---- */
.pack-edit-layout {
  display: flex;
  gap: 20px;
  height: 70vh;
}
.pack-edit-left {
  flex: 2;
  min-width: 0;
  overflow-y: auto;
  padding-right: 4px;
}
.pack-edit-right {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 14px;
}
/* ============ 内容工具栏（插入动态字段/放大编辑） ============ */
.pack-edit-content-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.pack-edit-content-toolbar .el-button {
  font-size: 12px;
}

/* 插入动态字段面板（popover 内容） */
.field-insert-panel {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-insert-tip {
  font-size: 12px;
  color: #6b7280;
  line-height: 1.5;
}

.field-insert-tip code {
  background: #eef2f7;
  color: #4f46e5;
  padding: 0 4px;
  border-radius: 3px;
  font-size: 11px;
}

.field-insert-section {
  font-size: 12px;
  font-weight: 600;
  color: #9ca3af;
  margin-top: 2px;
}

.field-insert-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
}

.field-insert-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  padding: 7px 10px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #fff;
  cursor: pointer;
  text-align: left;
  transition: all 0.18s ease;
}

.field-insert-item:hover {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}

.field-insert-label {
  font-size: 13px;
  color: #1f2937;
  font-weight: 500;
}

.field-insert-code {
  font-size: 11px;
  color: #4f46e5;
  background: transparent;
}

.field-insert-custom {
  display: flex;
  gap: 6px;
  align-items: center;
}

.field-insert-custom .el-input {
  flex: 1;
}

.field-insert-hint {
  font-size: 11px;
  color: #9ca3af;
  border-top: 1px dashed #e5e7eb;
  padding-top: 6px;
}

/* 字段配置头部按钮组 */
.pack-edit-right-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

/* 放大编辑弹窗 */
.pack-enlarge-edit-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.pack-enlarge-edit-hint {
  font-size: 12px;
  color: #9ca3af;
}

/* 深色主题适配 */
:root[data-theme='dark'] .field-insert-item {
  background: #1e2434;
  border-color: #2e3650;
}

:root[data-theme='dark'] .field-insert-item:hover {
  border-color: var(--el-color-primary);
  background: color-mix(in srgb, var(--el-color-primary) 12%, transparent);
}

:root[data-theme='dark'] .field-insert-label {
  color: #e5e7eb;
}

:root[data-theme='dark'] .field-insert-tip {
  color: #9ca3af;
}

:root[data-theme='dark'] .field-insert-tip code,
:root[data-theme='dark'] .field-insert-code,
:root[data-theme='dark'] .pack-edit-field-reference code {
  color: #93c5fd;
}

:root[data-theme='dark'] .field-insert-tip code {
  background: #2c3350;
}

:root[data-theme='dark'] .field-insert-hint {
  color: #6b7280;
  border-color: #2e3650;
}

:root[data-theme='dark'] .pack-enlarge-edit-hint {
  color: #6b7280;
}

.pack-edit-right-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
  gap: 8px;
}
.pack-edit-right-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.pack-edit-right-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
  white-space: nowrap;
}
.pack-edit-right-count {
  font-size: 11px;
  color: #fff;
  background: #3b82f6;
  border-radius: 10px;
  padding: 1px 8px;
  white-space: nowrap;
}
.pack-edit-right-tip {
  font-size: 11px;
  color: #94a3b8;
  margin-bottom: 10px;
}
.pack-edit-right-tip code {
  background: #eef2f7;
  color: #4f46e5;
  padding: 0 4px;
  border-radius: 3px;
}
.pack-edit-right-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

/* 字段卡片头部删除按钮 */
.pack-edit-field-head-delete {
  flex-shrink: 0;
  opacity: 0;
  transition: opacity 0.15s ease;
}
.pack-edit-field-head:hover .pack-edit-field-head-delete {
  opacity: 1;
}

/* 模板引用行（展开内容内） */
.pack-edit-field-ref-row {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 5px 8px;
  background: rgba(100, 116, 139, 0.08);
  border-radius: 6px;
}
.pack-edit-field-ref-code {
  flex: 1;
  min-width: 0;
  font-size: 11px;
  color: #4f46e5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 内容区警告横幅（橙色三角警告） */
.pack-edit-content-warn {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  padding: 8px 10px;
  margin-bottom: 8px;
  background: #fff7ed;
  border: 1px solid #fed7aa;
  border-radius: 8px;
  font-size: 12px;
  color: #c2410c;
  line-height: 1.6;
}
.pack-edit-content-warn .el-icon {
  flex-shrink: 0;
  margin-top: 2px;
  color: #ea580c;
}

/* 未使用字段提示 */
.pack-edit-unused-fields {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  padding: 8px 10px;
  margin-top: 8px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 8px;
  font-size: 12px;
  color: #92400e;
  line-height: 1.6;
}
.pack-edit-unused-fields .el-icon {
  flex-shrink: 0;
  margin-top: 2px;
  color: #d97706;
}
.pack-edit-unused-fields code {
  background: rgba(217, 119, 6, 0.12);
  color: #b45309;
  padding: 0 4px;
  border-radius: 3px;
}
.pack-edit-right-body {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}
.pack-edit-form :deep(.el-form-item) {
  margin-bottom: 18px;
}
.pack-edit-form :deep(.el-form-item__label) {
  font-size: 13px;
  color: #374151;
  font-weight: 600;
  padding-bottom: 6px;
  line-height: 1.5;
}
.pack-edit-content-input :deep(.el-textarea__inner) {
  padding: 12px 14px;
  font-size: 13px;
  line-height: 1.7;
  border-radius: 8px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}
.pack-edit-field-tip {
  display: flex;
  align-items: center;
  gap: 5px;
  margin-top: 6px;
  font-size: 12px;
  color: #94a3b8;
  line-height: 1.5;
}
.pack-edit-field-tip .el-icon {
  color: #eab308;
  flex-shrink: 0;
}
.pack-edit-section {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e5e7eb;
}
.pack-edit-section-head {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 8px;
}
.pack-edit-section-title {
  font-size: 13px;
  font-weight: 600;
  color: #1f2937;
}
.pack-edit-section-tip {
  font-size: 11px;
  color: #94a3b8;
}
.pack-edit-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}
.pack-edit-tag {
  margin: 0;
}
.pack-edit-tag-input {
  width: 120px;
}
.pack-edit-card-type {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.pack-edit-card-type-actions {
  display: flex;
  gap: 8px;
  margin-top: 4px;
}
.pack-edit-card-type-status {
  margin-top: 4px;
}
.pack-edit-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: #94a3b8;
  font-size: 13px;
  padding: 36px 0;
  text-align: center;
}
.pack-edit-empty-icon {
  font-size: 32px;
  color: #cbd5e1;
}
.pack-edit-empty p {
  margin: 0;
  font-size: 13px;
  color: #94a3b8;
}
.pack-edit-fields {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.pack-edit-field {
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
  transition: border-color 0.2s ease;
}
.pack-edit-field.expanded {
  border-color: var(--el-color-primary);
}
/* 卡片头（编号 + 名称 + 类型 + 箭头） */
.pack-edit-field-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 12px;
  cursor: pointer;
  user-select: none;
  transition: background 0.15s ease;
}
.pack-edit-field-head:hover {
  background: #f1f5f9;
}
.pack-edit-field-index {
  font-size: 12px;
  font-weight: 600;
  color: #94a3b8;
  flex-shrink: 0;
}
.pack-edit-field-name-text {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  font-weight: 500;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.pack-edit-field-type-tag {
  flex-shrink: 0;
}
.pack-edit-field-arrow {
  font-size: 13px;
  color: #94a3b8;
  flex-shrink: 0;
  transition: transform 0.25s ease;
}
.pack-edit-field-arrow.rotated {
  transform: rotate(180deg);
}
.pack-edit-field-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0 12px 12px;
  border-top: 1px dashed #e5e7eb;
  padding-top: 10px;
}
.pack-edit-field-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.pack-edit-field-row-label {
  font-size: 12px;
  color: #6b7280;
}
.pack-edit-field-row-label em {
  color: #f56c6c;
  font-style: normal;
}
.pack-edit-field-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-top: 4px;
}
.pack-edit-field-order {
  display: flex;
  gap: 2px;
}
.pack-edit-field-actions-right {
  display: flex;
  gap: 6px;
}
.pack-edit-field-options {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

/* ========== 顶部操作栏 ========== */
.pack-edit-opbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 0 0 14px;
  border-bottom: 1px solid rgba(100, 116, 139, 0.2);
  margin-bottom: 18px;
}
.pack-edit-opbar-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}
.pack-edit-opbar-title::before {
  content: '';
  width: 3px;
  height: 14px;
  border-radius: 2px;
  background: #3b82f6;
}
.pack-edit-opbar-actions {
  display: flex;
  align-items: center;
  gap: 2px;
}
.pack-edit-opbar-actions .el-button {
  font-size: 12px;
  color: #6b7280;
}
.pack-edit-opbar-actions .el-button:hover {
  color: #3b82f6;
  background: rgba(59, 130, 246, 0.1);
}

/* ========== 胶囊按钮组（适用作品/卡片类型/使用模式） ========== */
.pack-edit-pills.el-radio-group {
  display: inline-flex;
  gap: 0;
}
.pack-edit-pills :deep(.el-radio-button__inner) {
  padding: 7px 16px;
  border-color: #d1d5db;
  color: #4b5563;
  font-size: 13px;
  transition: all 0.2s ease;
}
.pack-edit-pills :deep(.el-radio-button:first-child .el-radio-button__inner) {
  border-radius: 20px 0 0 20px;
}
.pack-edit-pills :deep(.el-radio-button:last-child .el-radio-button__inner) {
  border-radius: 0 20px 20px 0;
}
.pack-edit-pills :deep(.el-radio-button__inner:hover) {
  color: #3b82f6;
}
.pack-edit-pills :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background: #3b82f6;
  border-color: #3b82f6;
  box-shadow: -1px 0 0 0 #3b82f6;
  color: #fff;
  font-weight: 500;
}
.pack-edit-pills :deep(.el-radio-button:last-child .el-radio-button__original-radio:checked + .el-radio-button__inner) {
  box-shadow: none;
}

/* ========== 使用模式 / 内容公开 ========== */
.pack-edit-usage-mode {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}
.pack-edit-usage-alert {
  width: 100%;
}
.pack-edit-content-public {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}
.pack-edit-content-public-label {
  font-size: 12px;
  color: #6b7280;
}
.pack-edit-content-public-hint {
  font-size: 12px;
  color: #94a3b8;
}

/* ========== 预览测试弹窗 ========== */
.preview-test-tip {
  font-size: 12px;
  color: #6b7280;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 6px;
  padding: 8px 10px;
  margin-bottom: 12px;
  line-height: 1.6;
}
.preview-test-layout {
  display: flex;
  gap: 14px;
  height: 52vh;
  min-height: 320px;
}
.preview-test-left {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
  border-right: 1px solid #e5e7eb;
  padding-right: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.preview-test-left h4 {
  margin: 0;
  font-size: 14px;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 6px;
}
.preview-test-left h4 span {
  font-size: 12px;
  color: #94a3b8;
  font-weight: normal;
}
.preview-test-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.preview-test-field label {
  font-size: 12px;
  color: #6b7280;
}
.preview-test-empty {
  color: #94a3b8;
  font-size: 12px;
  padding: 20px 0;
  text-align: center;
}
.preview-test-right {
  flex: 1.2;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.preview-test-right-head {
  display: flex;
  align-items: center;
  gap: 8px;
}
.preview-test-right-head h4 {
  margin: 0;
  font-size: 14px;
  color: #1f2937;
}
.preview-test-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 10px;
}
.preview-test-status.partial {
  color: #b45309;
  background: #fef3c7;
}
.preview-test-status.complete {
  color: #047857;
  background: #d1fae5;
}
.preview-test-content {
  flex: 1;
  overflow-y: auto;
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 12px;
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
  color: #1f2937;
}
.preview-test-content .unfilled {
  color: #b45309;
  background: #fef3c7;
  border-radius: 3px;
  padding: 0 2px;
  font-weight: 500;
}
.preview-test-content .filled {
  color: #047857;
  background: #d1fae5;
  border-radius: 3px;
  padding: 0 2px;
  font-weight: 500;
}
.preview-test-unfilled {
  font-size: 12px;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.preview-test-unfilled-tag {
  margin: 0;
}

/* ========== 生成参数弹窗 ========== */
.gen-params-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.gen-params-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.gen-params-label {
  width: 140px;
  flex-shrink: 0;
  font-size: 13px;
  color: #1f2937;
}
.gen-params-row .el-slider {
  flex: 1;
}
.gen-params-value {
  width: 50px;
  text-align: right;
  font-size: 13px;
  color: var(--el-color-primary);
  font-weight: 600;
}

/* ========== 使用说明弹窗 ========== */
.usage-guide {
  display: flex;
  flex-direction: column;
  gap: 18px;
  max-height: 60vh;
  overflow-y: auto;
}
.usage-guide-step {
  display: flex;
  gap: 12px;
}
.usage-guide-num {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--el-color-primary);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 2px;
}
.usage-guide-step h4 {
  margin: 0 0 6px;
  font-size: 14px;
  color: #1f2937;
}
.usage-guide-step p {
  margin: 0 0 6px;
  font-size: 13px;
  color: #6b7280;
  line-height: 1.7;
}
.usage-guide-step code {
  background: #eef2f7;
  color: #4f46e5;
  padding: 1px 5px;
  border-radius: 4px;
  font-size: 12px;
}
.usage-guide-example {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 6px;
  padding: 8px 10px;
}
.usage-guide-warn {
  color: #b45309 !important;
  background: #fefce8;
  border: 1px solid #fde68a;
  border-radius: 6px;
  padding: 8px 10px;
}

/* ========== 深色主题适配 ========== */
:root[data-theme='dark'] .pack-edit-opbar {
  border-color: rgba(100, 116, 139, 0.25);
}
:root[data-theme='dark'] .pack-edit-opbar-title {
  color: #e5e7eb;
}
:root[data-theme='dark'] .pack-edit-opbar-actions .el-button {
  color: #9ca3af;
}
:root[data-theme='dark'] .pack-edit-opbar-actions .el-button:hover {
  color: #60a5fa;
  background: rgba(59, 130, 246, 0.15);
}
:root[data-theme='dark'] .pack-edit-pills :deep(.el-radio-button__inner) {
  border-color: #3a4360;
  color: #9ca3af;
  background: transparent;
}
:root[data-theme='dark'] .pack-edit-pills :deep(.el-radio-button__inner:hover) {
  color: #60a5fa;
}
:root[data-theme='dark'] .pack-edit-pills :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background: #3b82f6;
  border-color: #3b82f6;
  color: #fff;
}
:root[data-theme='dark'] .pack-edit-field-tip {
  color: #6b7280;
}
:root[data-theme='dark'] .pack-edit-field-tip .el-icon {
  color: #eab308;
}
:root[data-theme='dark'] .pack-edit-form :deep(.el-form-item__label) {
  color: #cbd5e1;
}
:root[data-theme='dark'] .pack-edit-right-title {
  color: #e5e7eb;
}
:root[data-theme='dark'] .pack-edit-right-count {
  background: #3b82f6;
}
:root[data-theme='dark'] .pack-edit-right-tip code {
  background: #2c3350;
  color: #93c5fd;
}
:root[data-theme='dark'] .pack-edit-field-ref-row {
  background: rgba(100, 116, 139, 0.15);
}
:root[data-theme='dark'] .pack-edit-field-ref-code {
  color: #93c5fd;
}
:root[data-theme='dark'] .pack-edit-content-warn {
  background: rgba(234, 88, 12, 0.12);
  border-color: rgba(234, 88, 12, 0.35);
  color: #fdba74;
}
:root[data-theme='dark'] .pack-edit-content-warn .el-icon {
  color: #fb923c;
}
:root[data-theme='dark'] .pack-edit-unused-fields {
  background: rgba(217, 119, 6, 0.12);
  border-color: rgba(217, 119, 6, 0.35);
  color: #fcd34d;
}
:root[data-theme='dark'] .pack-edit-unused-fields .el-icon {
  color: #f59e0b;
}
:root[data-theme='dark'] .pack-edit-unused-fields code {
  background: rgba(217, 119, 6, 0.2);
  color: #fbbf24;
}
:root[data-theme='dark'] .pack-edit-field {
  background: #1e2434;
  border-color: #2e3650;
}
:root[data-theme='dark'] .pack-edit-right {
  background: #1a2030;
  border-color: #2e3650;
}
:root[data-theme='dark'] .pack-edit-empty-icon {
  color: #3a4360;
}
:root[data-theme='dark'] .pack-edit-field.expanded {
  border-color: var(--el-color-primary);
}
:root[data-theme='dark'] .pack-edit-field-head:hover {
  background: #222838;
}
:root[data-theme='dark'] .pack-edit-field-name-text {
  color: #e5e7eb;
}
:root[data-theme='dark'] .pack-edit-field-body {
  border-color: #2e3650;
}
:root[data-theme='dark'] .pack-edit-field-row-label {
  color: #9ca3af;
}
:root[data-theme='dark'] .pack-edit-content-public {
  background: #222838;
  border-color: #2e3650;
}
:root[data-theme='dark'] .pack-edit-content-public-label {
  color: #9ca3af;
}
:root[data-theme='dark'] .preview-test-tip {
  background: rgba(59, 130, 246, 0.12);
  border-color: #2e3650;
  color: #9ca3af;
}
:root[data-theme='dark'] .preview-test-left h4,
:root[data-theme='dark'] .preview-test-right-head h4 {
  color: #e5e7eb;
}
:root[data-theme='dark'] .preview-test-field label {
  color: #9ca3af;
}
:root[data-theme='dark'] .preview-test-content {
  background: #222838;
  border-color: #2e3650;
  color: #e5e7eb;
}
:root[data-theme='dark'] .preview-test-left {
  border-color: #2e3650;
}
:root[data-theme='dark'] .gen-params-label {
  color: #e5e7eb;
}
:root[data-theme='dark'] .usage-guide-step h4 {
  color: #e5e7eb;
}
:root[data-theme='dark'] .usage-guide-step p {
  color: #9ca3af;
}
:root[data-theme='dark'] .usage-guide-step code {
  background: #2c3350;
  color: #93c5fd;
}
:root[data-theme='dark'] .usage-guide-example {
  background: rgba(16, 185, 129, 0.1);
  border-color: #2e3650;
}
:root[data-theme='dark'] .usage-guide-warn {
  background: rgba(245, 158, 11, 0.1);
  border-color: #2e3650;
}

/* 布局：顶栏（标题+搜索）固定，统计/分段/卡片全部在下方内容区内滚动 */
.prompts-container {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: transparent;
  padding: 0 28px;
  animation: fadeIn 0.4s ease;
}

/* ---- 顶栏（固定，不参与滚动） ---- */
.top-bar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 14px 4px;
  background: transparent;
  border-bottom: 1px solid var(--topbar-border, #e6eaf0);
}
.top-title { flex-shrink: 0; }
.top-title h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 800;
  letter-spacing: -0.3px;
  color: #1f2937;
}
.top-sub { display: block; font-size: 11.5px; color: #9aa5b4; margin-top: 2px; }
.top-search { flex: 1; max-width: 380px; }
.top-search :deep(.el-input__wrapper) {
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 0 0 1.5px #eef0f3 inset;
  transition: box-shadow 0.2s ease;
}
.top-search :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1.5px #86efac inset, 0 0 0 4px rgba(34, 197, 94, 0.1);
}
.top-search :deep(.el-input__inner) { font-size: 13px; }

/* ---- 统计磁贴 ---- */
.stat-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
  margin: 8px 0 18px;
}
.stat-tile {
  display: flex;
  align-items: center;
  gap: 13px;
  background: #fff;
  border: 1px solid #eef0f3;
  border-radius: 14px;
  padding: 14px 16px;
  cursor: pointer;
  transition: all 0.2s ease;
}
.stat-tile:hover { transform: translateY(-2px); box-shadow: 0 8px 20px rgba(15, 23, 42, 0.07); }
.stat-ic {
  width: 38px;
  height: 38px;
  border-radius: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
  font-size: 18px;
}
.ic-purple { background: linear-gradient(135deg, #8b5cf6, #7c3aed); }
.ic-pink   { background: linear-gradient(135deg, #e879f9, #c026d3); }
.ic-green  { background: linear-gradient(135deg, #22c55e, #16a34a); }
.ic-orange { background: linear-gradient(135deg, #fa8c16, #ea580c); }
.ic-blue   { background: linear-gradient(135deg, #3b82f6, #2563eb); }
.ic-red    { background: linear-gradient(135deg, #ef4444, #dc2626); }
.ic-amber  { background: linear-gradient(135deg, #f59e0b, #d97706); }
.stat-body b { font-size: 20px; font-weight: 800; display: block; line-height: 1.1; color: #1f2937; }
.stat-body span { font-size: 11.5px; color: #94a3b8; }

.chip-row { display: flex; gap: 8px; margin-bottom: 18px; flex-wrap: wrap; }
.chip {
  padding: 8px 17px;
  border-radius: 999px;
  border: 1.5px solid #e8ecf2;
  background: #fff;
  font-size: 13px;
  font-weight: 700;
  color: #64748b;
  cursor: pointer;
  transition: all 0.16s ease;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  font-family: inherit;
}
.chip:hover { border-color: #c7d2fe; color: #4338ca; }
.chip.on {
  background: #6d28d9;
  border-color: #6d28d9;
  color: #fff;
  box-shadow: 0 3px 12px rgba(109, 40, 217, 0.3);
}
.chip-n {
  font-size: 10.5px;
  background: #f1f5f9;
  color: #94a3b8;
  border-radius: 999px;
  padding: 1px 8px;
  font-weight: 600;
}
.chip.on .chip-n { background: rgba(255, 255, 255, 0.25); color: #fff; }

/* ---- 内容区（唯一滚动区域） ---- */
.content-area {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 4px 4px 44px;
}
.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(232px, 1fr));
  gap: 14px;
  padding: 6px;
  border-radius: 16px;
  transition: background 0.2s ease;
}
.cards-grid.drag-over {
  background: rgba(250, 140, 22, 0.06);
  outline: 2px dashed #fbd38d;
  outline-offset: -2px;
}
.card {
  background: #fff;
  border: 1px solid #eef0f3;
  border-radius: 14px;
  padding: 15px 16px;
  transition: all 0.2s ease;
  cursor: pointer;
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 11px;
}
.card:hover {
  transform: translateY(-3px);
  box-shadow: 0 10px 26px rgba(15, 23, 42, 0.08);
  border-color: #d8e0ec;
}
.card.drag-over {
  border-color: #93c5fd;
  background: #f0f7ff;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.12);
}
.card-book { border-color: #fde4cd; }
.card-book:hover { border-color: #fb923c; }
.card-std { border-color: #fde9c4; }
.card-std:hover { border-color: #fa8c16; box-shadow: 0 8px 20px rgba(250, 140, 22, 0.14); }
.card-std.dragging { opacity: 0.45; transform: scale(0.96); }
.card-std.encrypted { border-color: rgba(230, 162, 60, 0.45); }

.card-main { display: flex; align-items: center; gap: 11px; }
.card-ic {
  width: 38px;
  height: 38px;
  border-radius: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
  font-size: 18px;
}
.card-info { flex: 1; min-width: 0; }
.card-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14.5px;
  font-weight: 700;
  color: #1f2937;
  min-width: 0;
}
.card-name-text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.card-grip { color: #fa8c16; font-size: 13px; flex-shrink: 0; cursor: grab; }
.std-lock { color: #e6a23c; font-size: 13px; flex-shrink: 0; }
.card-desc {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.std-preview {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  white-space: normal;
  line-height: 1.6;
  color: #8c8c8c;
}
.std-preview .el-icon { vertical-align: -2px; }

.card-foot {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-top: 10px;
  border-top: 1px dashed #f1f5f9;
  font-size: 11px;
  color: #94a3b8;
  margin-top: auto;
}
.card-tag { font-size: 10.5px; font-weight: 700; padding: 2px 9px; border-radius: 999px; flex-shrink: 0; }
.tg-pack { background: #eef2ff; color: #4338ca; }
.tg-book { background: rgba(230, 162, 60, 0.16); color: #b45309; }
.tg-std  { background: #f0fdf4; color: #16a34a; }
.card-foot-time { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.card-ops { flex-shrink: 0; }
.card-more {
  width: 26px;
  height: 26px;
  border: none;
  background: transparent;
  border-radius: 7px;
  color: #94a3b8;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}
.card-more:hover { background: #f1f5f9; color: #475569; }

.std-actions { display: flex; gap: 3px; flex-shrink: 0; }
.std-act {
  width: 25px;
  height: 25px;
  border: 1px solid transparent;
  background: rgba(15, 118, 110, 0.06);
  border-radius: 999px;
  color: #0f766e;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.18s ease;
}
.std-act .el-icon { font-size: 12px; }
.std-act:hover {
  color: #fff;
  background: linear-gradient(135deg, #0f766e, #2563eb);
  box-shadow: 0 6px 14px rgba(37, 99, 235, 0.22);
}
.std-act.danger { color: #c2415b; background: rgba(225, 29, 72, 0.06); }
.std-act.danger:hover {
  color: #fff;
  background: linear-gradient(135deg, #e11d48, #fb7185);
  box-shadow: 0 6px 14px rgba(225, 29, 72, 0.18);
}

.standalone-toolbar { display: flex; align-items: center; gap: 10px; margin: 4px 4px 12px; }
.std-toolbar-title { font-size: 12.5px; color: #64748b; font-weight: 600; }
.std-toolbar-title em { font-style: normal; color: #b45309; }
.std-batch {
  margin-left: auto;
  width: 30px;
  height: 30px;
  border: 1px solid #fecaca;
  background: #fef2f2;
  color: #dc2626;
  border-radius: 9px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}
.std-batch:hover { background: #fee2e2; }
.std-area { border-radius: 14px; }

/* ---- FAB ---- */
.fab {
  position: fixed;
  right: 26px;
  bottom: 26px;
  z-index: 80;
  width: 58px;
  height: 58px;
  border-radius: 50%;
  border: none;
  background: linear-gradient(135deg, #8b5cf6, #7c3aed);
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 10px 26px rgba(124, 58, 237, 0.4);
  transition: all 0.22s ease;
}
.fab:hover { transform: translateY(-3px) scale(1.05); box-shadow: 0 14px 34px rgba(124, 58, 237, 0.5); }
.fab:active { transform: scale(0.96); }

/* ---- 底部快捷中心 ---- */
.sheet-mask {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(5px);
  -webkit-backdrop-filter: blur(5px);
  z-index: 90;
  display: flex;
  align-items: flex-end;
}
.sheetup-enter-active, .sheetup-leave-active { transition: opacity 0.25s ease; }
.sheetup-enter-from, .sheetup-leave-to { opacity: 0; }
.sheet {
  width: 100%;
  max-width: 680px;
  margin: 0 auto;
  background: #fff;
  border-radius: 22px 22px 0 0;
  padding: 10px 0 26px;
  max-height: 78vh;
  overflow-y: auto;
  box-shadow: 0 -10px 40px rgba(0, 0, 0, 0.14);
  animation: sheetUp 0.28s cubic-bezier(0.4, 0, 0.2, 1);
}
@keyframes sheetUp { from { transform: translateY(100%); } }
.sheet-grip { width: 40px; height: 4px; border-radius: 99px; background: #e2e8f0; margin: 6px auto 12px; }
.sheet-head { display: flex; align-items: baseline; gap: 10px; padding: 2px 24px 12px; flex-wrap: wrap; }
.sheet-head b { font-size: 16px; color: #1f2937; }
.sheet-head span { font-size: 12px; color: #94a3b8; }
.sheet-section { padding: 4px 24px; }
.sheet-label { font-size: 11px; font-weight: 800; color: #94a3b8; letter-spacing: 1px; margin: 10px 0 4px; }
.sheet-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 11px 6px;
  cursor: pointer;
  border-radius: 12px;
  transition: background 0.15s ease;
}
.sheet-row:hover { background: #f8fafc; }
.sheet-ic {
  width: 38px;
  height: 38px;
  border-radius: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
  font-size: 17px;
}
.sheet-info { flex: 1; min-width: 0; }
.sheet-info b { font-size: 14px; color: #1f2937; display: block; }
.sheet-info span { font-size: 12px; color: #94a3b8; }
.sheet-go { color: #c3cad6; font-size: 18px; flex-shrink: 0; }
.sheet-sep { height: 1px; background: #eef0f3; margin: 8px 24px; }

:root[data-theme='dark'] .prompts-container { background: transparent; }
:root[data-theme='dark'] .top-bar { background: transparent; border-bottom-color: rgba(42, 50, 66, 0.4); }
:root[data-theme='dark'] .top-title h2 { color: #e6edf3; }
:root[data-theme='dark'] .top-sub { color: #5f6b7a; }
:root[data-theme='dark'] .top-search :deep(.el-input__wrapper) { background: rgba(22, 27, 34, 0.9); box-shadow: 0 0 0 1.5px rgba(42, 50, 66, 0.5) inset; }
:root[data-theme='dark'] .stat-tile { background: rgba(22, 27, 34, 0.85); border-color: rgba(42, 50, 66, 0.35); }
:root[data-theme='dark'] .stat-tile:hover { box-shadow: 0 8px 20px rgba(0, 0, 0, 0.3); }
:root[data-theme='dark'] .stat-body b { color: #e6edf3; }
:root[data-theme='dark'] .chip { background: rgba(255, 255, 255, 0.05); border-color: rgba(42, 50, 66, 0.5); color: #8b98a8; }
:root[data-theme='dark'] .chip:hover { border-color: rgba(165, 180, 252, 0.5); color: #c7d2fe; }
:root[data-theme='dark'] .chip.on { background: #6d28d9; border-color: #6d28d9; color: #fff; }
:root[data-theme='dark'] .chip-n { background: rgba(255, 255, 255, 0.08); color: #5f6b7a; }
:root[data-theme='dark'] .card { background: rgba(22, 27, 34, 0.9); border-color: rgba(42, 50, 66, 0.4); }
:root[data-theme='dark'] .card:hover { border-color: rgba(148, 163, 184, 0.4); box-shadow: 0 10px 26px rgba(0, 0, 0, 0.3); }
:root[data-theme='dark'] .card-name, :root[data-theme='dark'] .card-name-text { color: #e6edf3; }
:root[data-theme='dark'] .card-desc { color: #5f6b7a; }
:root[data-theme='dark'] .card-foot { border-top-color: rgba(42, 50, 66, 0.3); }
:root[data-theme='dark'] .card-book, :root[data-theme='dark'] .card-std { border-color: rgba(249, 115, 22, 0.25); }
:root[data-theme='dark'] .card-more { color: #5f6b7a; }
:root[data-theme='dark'] .card-more:hover { background: rgba(255, 255, 255, 0.08); color: #e6edf3; }
:root[data-theme='dark'] .sheet { background: #161b22; }
:root[data-theme='dark'] .sheet-head b { color: #e6edf3; }
:root[data-theme='dark'] .sheet-row:hover { background: rgba(255, 255, 255, 0.05); }
:root[data-theme='dark'] .sheet-info b { color: #e6edf3; }
:root[data-theme='dark'] .sheet-grip { background: #2a3242; }
:root[data-theme='dark'] .sheet-sep { background: rgba(42, 50, 66, 0.4); }
:root[data-theme='dark'] .std-toolbar-title { color: #8b98a8; }
:root[data-theme='dark'] .fab { box-shadow: 0 10px 26px rgba(0, 0, 0, 0.5); }

/* ---- 小屏适配 ---- */
@media (max-width: 768px) {
  .prompts-container { padding: 0 14px; }
  .top-bar { gap: 10px; padding: 12px 2px; }
  .top-sub { display: none; }
  .content-area { padding: 2px 2px 36px; }
}

</style>

<!-- ===================== 弹窗全局样式（el-dialog/el-select 等 Teleport 到 body 的元素需全局选择器） ===================== -->
<style>
/* 插入动态字段 popover 深色适配 */
:root[data-theme='dark'] .pack-edit-field-popover.el-popover {
  background: #1e2434;
  border-color: #2e3650;
  --el-popover-bg-color: #1e2434;
  --el-bg-color-overlay: #1e2434;
}

/* 创建/编辑提示词弹窗：阴影 + 圆角质感 */
.pack-edit-dialog.el-dialog {
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.35), 0 2px 8px rgba(0, 0, 0, 0.15);
}
.pack-edit-dialog .el-dialog__header {
  padding: 18px 24px 14px;
  margin: 0;
  border-bottom: 1px solid rgba(100, 116, 139, 0.15);
}
.pack-edit-dialog .el-dialog__title {
  font-size: 16px;
  font-weight: 600;
}
.pack-edit-dialog .el-dialog__body {
  padding: 18px 24px;
}
.pack-edit-dialog .el-dialog__footer {
  padding: 12px 24px 16px;
  border-top: 1px solid rgba(100, 116, 139, 0.15);
}

/* 深色主题：弹窗背景比页面提亮一层（层次感） */
:root[data-theme='dark'] .pack-edit-dialog.el-dialog {
  background: #1f2536;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.55), 0 2px 10px rgba(0, 0, 0, 0.3);
}
:root[data-theme='dark'] .pack-edit-dialog .el-dialog__header,
:root[data-theme='dark'] .pack-edit-dialog .el-dialog__footer {
  border-color: rgba(100, 116, 139, 0.2);
}

/* 深色主题：弹窗内输入框/下拉框统一深色 + 清晰边框（与背景对比） */
:root[data-theme='dark'] .pack-edit-dialog {
  --el-input-bg-color: #141a29;
  --el-input-border-color: #333c56;
  --el-input-hover-border-color: #4a5578;
  --el-input-focus-border-color: #3b82f6;
  --el-input-text-color: #e5e7eb;
  --el-input-placeholder-color: #64748b;
  --el-fill-color-blank: #141a29;
}
:root[data-theme='dark'] .pack-edit-dialog .el-input__wrapper,
:root[data-theme='dark'] .pack-edit-dialog .el-textarea__inner {
  background: #141a29;
  border-radius: 8px;
}

/* 深色主题：下拉选择面板 */
:root[data-theme='dark'] .pack-edit-dialog .el-select-dropdown,
:root[data-theme='dark'] .el-select-dropdown {
  background: #1e2434;
  border-color: #2e3650;
}
:root[data-theme='dark'] .pack-edit-dialog .el-select-dropdown__item,
:root[data-theme='dark'] .el-select-dropdown__item {
  color: #cbd5e1;
}
:root[data-theme='dark'] .pack-edit-dialog .el-select-dropdown__item.hover,
:root[data-theme='dark'] .el-select-dropdown__item:hover,
:root[data-theme='dark'] .el-select-dropdown__item.is-hovering {
  background: rgba(59, 130, 246, 0.12);
  color: #e5e7eb;
}
:root[data-theme='dark'] .pack-edit-dialog .el-select-dropdown__item.is-selected,
:root[data-theme='dark'] .el-select-dropdown__item.is-selected {
  color: #60a5fa;
}
:root[data-theme='dark'] .el-popper.is-light {
  --el-popper-border-color: #2e3650;
  background: #1e2434;
}
</style>
