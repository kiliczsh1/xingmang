﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿<template>
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

    <!-- 统计磁贴 -->
    <div class="stat-row">
      <div class="stat-tile" @click="activeMainTab = 'pack'">
        <div class="stat-ic ic-purple"><el-icon><FolderOpened /></el-icon></div>
        <div class="stat-body"><b>{{ totalPackCards }}</b><span>卡包</span></div>
      </div>
      <div class="stat-tile" @click="activeMainTab = 'bookAnalysis'">
        <div class="stat-ic ic-pink"><el-icon><Reading /></el-icon></div>
        <div class="stat-body"><b>{{ totalBookAnalysisCards }}</b><span>拆书卡包</span></div>
      </div>
      <div class="stat-tile" @click="activeMainTab = 'standalone'">
        <div class="stat-ic ic-green"><el-icon><Document /></el-icon></div>
        <div class="stat-body"><b>{{ combinedUncategorizedPrompts.length }}</b><span>独立卡片</span></div>
      </div>
      <div class="stat-tile">
        <div class="stat-ic ic-orange"><el-icon><Collection /></el-icon></div>
        <div class="stat-body"><b>{{ totalPackCards + totalBookAnalysisCards + combinedUncategorizedPrompts.length }}</b><span>总提示词</span></div>
      </div>
    </div>

    <!-- Tab 分段 -->
    <div class="chip-row">
      <button class="chip" :class="{ on: activeMainTab === 'all' }" @click="activeMainTab = 'all'">
        全部<span class="chip-n">{{ totalPackCards + totalBookAnalysisCards + combinedUncategorizedPrompts.length }}</span>
      </button>
      <button class="chip" :class="{ on: activeMainTab === 'pack' }" @click="activeMainTab = 'pack'">
        卡包<span class="chip-n">{{ totalPackCards }}</span>
      </button>
      <button class="chip" :class="{ on: activeMainTab === 'bookAnalysis' }" @click="activeMainTab = 'bookAnalysis'">
        拆书<span class="chip-n">{{ totalBookAnalysisCards }}</span>
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
            @drop.prevent="onDrop(category.name, $event)"
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
                      <el-dropdown-item v-if="category.prompts.length > 0" command="preview">
                        <el-icon><View /></el-icon>预览卡包
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
            @drop.prevent="onBookAnalysisDrop(category.name, $event)"
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
                      <el-dropdown-item v-if="category.prompts.length > 0" command="preview">
                        <el-icon><View /></el-icon>预览卡包
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
            @drop.prevent="onDrop(category.name, $event)"
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
                      <el-dropdown-item v-if="category.prompts.length > 0" command="preview">
                        <el-icon><View /></el-icon>预览卡包
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
            @drop.prevent="onBookAnalysisDrop(category.name, $event)"
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
                      <el-dropdown-item v-if="category.prompts.length > 0" command="preview">
                        <el-icon><View /></el-icon>预览卡包
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

    <!-- 提示词查重弹窗 -->
    <el-dialog
      v-model="dedupDialogVisible"
      title="提示词查重"
      width="800px"
      top="6vh"
      :close-on-click-modal="false"
      destroy-on-close
      class="dedup-dialog"
    >
      <div class="dedup-topbar">
        <div class="dedup-mode">
          <span class="dedup-mode-label">判重口径</span>
          <el-radio-group v-model="dedupMode" size="small" @change="rebuildDedupGroups">
            <el-radio-button value="titleCategory">标题 + 分类</el-radio-button>
            <el-radio-button value="title">仅标题</el-radio-button>
          </el-radio-group>
        </div>
        <div v-if="dedupGroups.length > 0" class="dedup-stats">
          共 <b>{{ dedupGroupCount }}</b> 组重复，涉及 <b>{{ dedupTotalCount }}</b> 张卡片
        </div>
      </div>

      <div v-if="dedupLoading" class="dedup-loading">正在扫描重复提示词…</div>
      <el-empty v-else-if="dedupGroups.length === 0" description="未发现重复提示词" :image-size="90" />
      <div v-else class="dedup-groups">
        <div v-for="group in dedupGroups" :key="group.key" class="dedup-group">
          <div class="dedup-group-head">
            <span class="dedup-group-title" :title="group.name">{{ group.name }}</span>
            <span v-if="group.category !== undefined" class="dedup-group-cat">{{ group.category }}</span>
            <span class="dedup-group-count">×{{ group.items.length }}</span>
            <span class="dedup-group-spacer"></span>
            <el-button size="small" text @click="setDedupKeep(group.key, 'oldest')">保留最早</el-button>
            <el-button size="small" text @click="setDedupKeep(group.key, 'newest')">保留最新</el-button>
          </div>
          <div class="dedup-group-items">
            <div
              v-for="item in group.items"
              :key="item.id"
              class="dedup-item"
              :class="{ selected: dedupSelected.has(item.id) }"
            >
              <el-checkbox
                :model-value="dedupSelected.has(item.id)"
                :disabled="item.card_type === 'encrypted'"
                @change="toggleDedupItem(item.id, $event)"
              />
              <div class="dedup-item-body">
                <div class="dedup-item-line1">
                  <el-tag v-if="item.card_type === 'encrypted'" size="small" type="warning" class="dedup-lock-tag">
                    <el-icon><Lock /></el-icon>
                    需解锁
                  </el-tag>
                  <span class="dedup-item-name">{{ item.name }}</span>
                  <span class="dedup-item-cat">{{ item.category }}</span>
                  <span class="dedup-item-date">{{ formatDate(item.created_at) }}</span>
                </div>
                <div class="dedup-item-content">{{ getDedupContentPreview(item) }}</div>
              </div>
              <el-tag v-if="isDedupKeepItem(group, item)" size="small" type="success" effect="light" class="dedup-keep-tag">保留</el-tag>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <span v-if="dedupSelectedCount > 0" class="dedup-footer-tip">
          将删除 {{ dedupSelectedCount }} 张重复卡片（加密卡片需先解锁，不参与一键删除）
        </span>
        <el-button @click="dedupDialogVisible = false">关闭</el-button>
        <el-button
          type="danger"
          :loading="dedupDeleting"
          :disabled="dedupSelectedCount === 0"
          @click="handleDedupDelete"
        >
          删除选中{{ dedupSelectedCount > 0 ? ` ${dedupSelectedCount} 张` : '' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 卡包合并弹窗 -->
    <el-dialog
      v-model="mergeDialogVisible"
      title="卡包合并"
      width="760px"
      top="6vh"
      :close-on-click-modal="false"
      destroy-on-close
      class="merge-dialog"
    >
      <div class="merge-mode-bar">
        <el-radio-group v-model="mergeMode" size="small">
          <el-radio-button value="new">合并为新卡包</el-radio-button>
          <el-radio-button value="absorb">主包吸收副包</el-radio-button>
        </el-radio-group>
        <span class="merge-mode-desc">
          {{ mergeMode === 'new' ? '多个自定义卡包合并为一个新卡包（需自定义命名）' : '选择主卡包与副卡包，副包卡片并入主包' }}
        </span>
      </div>

      <el-empty v-if="mergeablePacks.length === 0" description="暂无可合并的自定义卡包" :image-size="80" />

      <div v-else class="merge-body">
        <!-- 模式1：合并为新卡包 -->
        <template v-if="mergeMode === 'new'">
          <div class="merge-section">
            <div class="merge-section-title">新卡包名称</div>
            <el-input
              v-model="mergeNewName"
              placeholder="请输入新卡包名称"
              maxlength="40"
              clearable
            />
            <div class="merge-section-tip">合并后将创建该名称的新卡包，源卡包自动移除</div>
          </div>
          <div class="merge-section">
            <div class="merge-section-title">选择源卡包 <span class="merge-section-count">{{ mergeSelectedSources.length }} / {{ mergeablePacks.length }}</span></div>
            <div class="merge-pack-list">
              <el-checkbox-group v-model="mergeSelectedSources">
                <div v-for="pack in mergeablePacks" :key="pack.name" class="merge-pack-item">
                  <el-checkbox :value="pack.name">
                    <div class="merge-pack-info">
                      <el-icon class="merge-pack-ic"><FolderOpened /></el-icon>
                      <span class="merge-pack-name">{{ pack.name.replace(BOOK_ANALYSIS_PREFIX, '') }}</span>
                      <el-tag v-if="pack.isBookAnalysis" size="small" type="info" effect="plain" class="merge-pack-tag">拆书</el-tag>
                      <span class="merge-pack-count">{{ pack.prompts.length }} 张</span>
                    </div>
                  </el-checkbox>
                </div>
              </el-checkbox-group>
            </div>
          </div>
          <div class="merge-preview">
            将合并 <b>{{ mergeSelectedSources.length }}</b> 个卡包，共 <b>{{ mergePreviewTotal }}</b> 张卡片
          </div>
        </template>

        <!-- 模式2：主包吸收副包 -->
        <template v-else>
          <div class="merge-section">
            <div class="merge-section-title">主卡包 <span class="merge-section-tip-inline">默认卡包也可作主包</span></div>
            <div class="merge-pack-list merge-pack-list--radio">
              <el-radio-group v-model="mergeMainName">
                <div v-for="pack in mergeMainOptions" :key="pack.name" class="merge-pack-item">
                  <el-radio :value="pack.name">
                    <div class="merge-pack-info">
                      <el-icon class="merge-pack-ic"><FolderOpened /></el-icon>
                      <span class="merge-pack-name">{{ pack.name.replace(BOOK_ANALYSIS_PREFIX, '') }}</span>
                      <el-tag v-if="pack.isBookAnalysis" size="small" type="info" effect="plain" class="merge-pack-tag">拆书</el-tag>
                      <el-tag v-if="isDefaultPack(pack.name)" size="small" type="warning" effect="light" class="merge-pack-tag">默认</el-tag>
                      <span class="merge-pack-count">{{ pack.prompts.length }} 张</span>
                    </div>
                  </el-radio>
                </div>
              </el-radio-group>
            </div>
          </div>
          <div class="merge-section">
            <div class="merge-section-title">选择副卡包 <span class="merge-section-count">{{ mergeSelectedSubs.length }} / {{ mergeAvailableSubs.length }}</span></div>
            <div v-if="mergeAvailableSubs.length === 0" class="merge-empty-subs">请先选择主卡包，或无更多自定义卡包可选</div>
            <div v-else class="merge-pack-list">
              <el-checkbox-group v-model="mergeSelectedSubs">
                <div v-for="pack in mergeAvailableSubs" :key="pack.name" class="merge-pack-item">
                  <el-checkbox :value="pack.name">
                    <div class="merge-pack-info">
                      <el-icon class="merge-pack-ic"><FolderOpened /></el-icon>
                      <span class="merge-pack-name">{{ pack.name.replace(BOOK_ANALYSIS_PREFIX, '') }}</span>
                      <el-tag v-if="pack.isBookAnalysis" size="small" type="info" effect="plain" class="merge-pack-tag">拆书</el-tag>
                      <span class="merge-pack-count">{{ pack.prompts.length }} 张</span>
                    </div>
                  </el-checkbox>
                </div>
              </el-checkbox-group>
            </div>
          </div>
          <div class="merge-preview" v-if="mergeMainName">
            将 {{ mergeSelectedSubs.length }} 个副卡包并入「{{ mergeMainName.replace(BOOK_ANALYSIS_PREFIX, '') }}」，共移动 <b>{{ mergePreviewTotal }}</b> 张卡片
          </div>
          <div class="merge-preview" v-else>请选择主卡包</div>
        </template>

        <div class="merge-note">
          <el-icon><InfoFilled /></el-icon>
          <span>合并不去重，如产生同名重复卡片可使用「提示词查重」工具清理</span>
        </div>
      </div>

      <template #footer>
        <el-button @click="mergeDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="mergeLoading"
          :disabled="!canMerge"
          @click="handleMergeExecute"
        >
          开始合并
        </el-button>
      </template>
    </el-dialog>

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

    <!-- 导出独立卡片弹窗 -->
    <el-dialog
      v-model="exportStandaloneDialogVisible"
      title="导出独立卡片"
      width="800px"
      :close-on-click-modal="false"
      class="export-dialog"
      append-to-body
    >
      <div class="export-standalone-content">
        <div class="export-hint">
          <el-alert
            title="请选择要导出的独立卡片"
            type="info"
            :closable="false"
            show-icon
          >
            勾选需要导出的卡片，然后点击导出按钮。导出的文件为JSON格式。
          </el-alert>
        </div>
        <div class="export-select-all">
          <el-checkbox
            v-model="selectAllStandalone"
            @change="handleSelectAllStandalone"
          >
            全选 ({{ selectedStandaloneCards.length }}/{{ combinedUncategorizedPrompts.length }})
          </el-checkbox>
        </div>
        <div v-if="combinedUncategorizedPrompts.length > 0" class="export-cards-grid">
          <div
            v-for="prompt in combinedUncategorizedPrompts"
            :key="prompt.id"
            class="export-card-item"
            :class="{ selected: selectedStandaloneCards.includes(prompt.id) }"
            @click="toggleStandaloneCard(prompt.id)"
          >
            <el-checkbox
              :model-value="selectedStandaloneCards.includes(prompt.id)"
              @click.stop
              @change="toggleStandaloneCard(prompt.id)"
            />
            <div class="export-card-content">
              <div class="export-card-name">
                {{ prompt.name }}
                <el-tag v-if="prompt.category === '拆书-未分类'" size="small" type="warning" style="margin-left: 6px;">拆书</el-tag>
              </div>
              <div class="export-card-preview">
                <template v-if="prompt.card_type === 'encrypted'">
                  <el-icon><Lock /></el-icon> 已加密
                </template>
                <template v-else>
                  {{ prompt.content.slice(0, 30) }}{{ prompt.content.length > 30 ? '...' : '' }}
                </template>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="export-empty">
          <el-empty description="暂无独立卡片可导出" />
        </div>
      </div>
      <template #footer>
        <el-button @click="exportStandaloneDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :disabled="selectedStandaloneCards.length === 0"
          @click="executeExportStandalone"
        >
          导出 ({{ selectedStandaloneCards.length }})
        </el-button>
      </template>
    </el-dialog>

    <!-- 导出卡包弹窗 -->
    <el-dialog
      v-model="exportPackDialogVisible"
      title="导出卡包"
      width="600px"
      :close-on-click-modal="false"
      class="export-dialog"
      append-to-body
    >
      <div class="export-pack-content">
        <div class="export-hint">
          <el-alert
            title="请选择要导出的卡包"
            type="info"
            :closable="false"
            show-icon
          >
            选择需要导出的卡包，导出的文件将包含卡包内的所有卡片。拆书卡包与普通卡包统一在此选择。
          </el-alert>
        </div>
        <div v-if="allExportablePacks.length > 0" class="export-pack-list">
          <div
            v-for="category in allExportablePacks"
            :key="category.name"
            class="export-pack-item"
            :class="{ selected: selectedPack === category.name }"
            @click="selectedPack = category.name"
          >
            <el-radio :model-value="selectedPack" :label="category.name">
              <div class="pack-item-info">
                <el-icon><FolderOpened /></el-icon>
                <span class="pack-item-name">{{ category.isBookAnalysis ? category.name.replace(BOOK_ANALYSIS_PREFIX, '') : category.name }}</span>
                <el-tag v-if="category.isBookAnalysis" size="small" type="warning">拆书</el-tag>
                <el-tag size="small" type="info">{{ category.prompts.length }} 张卡片</el-tag>
              </div>
            </el-radio>
          </div>
        </div>
        <div v-else class="export-empty">
          <el-empty description="暂无卡包可导出" />
        </div>
      </div>
      <template #footer>
        <el-button @click="exportPackDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :disabled="!selectedPack"
          @click="executeExportPack"
        >
          导出卡包
        </el-button>
      </template>
    </el-dialog>

    <!-- 导入独立卡片弹窗 -->
    <el-dialog
      v-model="importStandaloneDialogVisible"
      title="导入独立卡片"
      width="800px"
      :close-on-click-modal="false"
      class="import-dialog"
      append-to-body
    >
      <div class="import-content">
        <div class="import-upload">
          <el-upload
            ref="standaloneUploadRef"
            drag
            accept=".json"
            :auto-upload="false"
            :show-file-list="false"
            :on-change="handleStandaloneFileChange"
          >
            <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
            <div class="el-upload__text">
              将文件拖到此处，或<em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">只能上传 JSON 格式的提示词文件</div>
            </template>
          </el-upload>
        </div>
        <div v-if="importStandalonePreview.length > 0" class="import-preview">
          <div class="import-preview-header">
            <span>预览导入内容 ({{ importStandalonePreview.length }} 张卡片)</span>
            <el-checkbox
              v-model="selectAllImportStandalone"
              @change="handleSelectAllImportStandalone"
            >
              全选
            </el-checkbox>
          </div>
          <div class="import-cards-grid">
            <div
              v-for="(prompt, index) in importStandalonePreview"
              :key="index"
              class="import-card-item"
              :class="{ selected: selectedImportCards.includes(index) }"
              @click="toggleImportCard(index)"
            >
              <el-checkbox
                :model-value="selectedImportCards.includes(index)"
                @click.stop
                @change="toggleImportCard(index)"
              />
              <div class="import-card-content">
                <div class="import-card-name">{{ prompt.name }}</div>
                <div class="import-card-preview">
                  <template v-if="prompt.card_type === 'encrypted'">
                    <el-icon><Lock /></el-icon> 已加密
                  </template>
                  <template v-else>
                    {{ prompt.content.slice(0, 30) }}{{ prompt.content.length > 30 ? '...' : '' }}
                  </template>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="importStandaloneDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :disabled="selectedImportCards.length === 0"
          @click="executeImportStandalone"
        >
          导入 ({{ selectedImportCards.length }})
        </el-button>
      </template>
    </el-dialog>

    <!-- 导入卡包弹窗 -->
    <el-dialog
      v-model="importPackDialogVisible"
      title="导入卡包"
      width="600px"
      :close-on-click-modal="false"
      class="import-dialog"
      append-to-body
    >
      <div class="import-content">
        <div class="import-upload">
          <el-upload
            ref="packUploadRef"
            drag
            accept=".json"
            :auto-upload="false"
            :show-file-list="false"
            :on-change="handlePackFileChange"
          >
            <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
            <div class="el-upload__text">
              将文件拖到此处，或<em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">只能上传 JSON 格式的卡包文件（普通 / 拆书自动识别）</div>
            </template>
          </el-upload>
        </div>
        <div v-if="importPackPreview" class="import-pack-preview">
          <div class="import-pack-info">
            <el-icon><FolderOpened /></el-icon>
            <span class="import-pack-name">{{ importPackPreview.packName }}</span>
            <el-tag v-if="importPackPreview.isBookAnalysis" size="small" type="warning">拆书</el-tag>
            <el-tag size="small" type="info">{{ importPackPreview.prompts.length }} 张卡片</el-tag>
          </div>
          <div class="import-pack-prompts">
            <div v-for="(prompt, index) in importPackPreview.prompts" :key="index" class="import-pack-prompt-item">
              <el-icon><Document /></el-icon>
              <span>{{ prompt.name }}</span>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="importPackDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :disabled="!importPackPreview"
          @click="executeImportPack"
        >
          导入卡包
        </el-button>
      </template>
    </el-dialog>

    <!-- 批量删除对话框 -->
    <el-dialog
      v-model="batchDeleteDialogVisible"
      title="批量删除提示词"
      width="800px"
      :close-on-click-modal="false"
      class="batch-delete-dialog"
      append-to-body
    >
      <div class="batch-delete-content">
        <div class="batch-delete-hint">
          <el-alert
            :title="'正在删除卡包' + currentCategory + '中的提示词'"
            type="warning"
            :closable="false"
            show-icon
          >
            勾选需要删除的卡片，然后点击删除按钮。此操作不可恢复。
          </el-alert>
        </div>
        <div class="batch-delete-select-all">
          <el-checkbox
            v-model="selectAllBatchDelete"
            @change="handleSelectAllBatchDelete"
          >
            全选 ({{ selectedBatchDeleteCards.length }}/{{ currentCategoryPrompts.length }})
          </el-checkbox>
        </div>
        <div v-if="currentCategoryPrompts.length > 0" class="batch-delete-cards-grid">
          <div
            v-for="prompt in currentCategoryPrompts"
            :key="prompt.id"
            class="batch-delete-card-item"
            :class="{ selected: selectedBatchDeleteCards.includes(prompt.id) }"
            @click="toggleBatchDeleteCard(prompt.id)"
          >
            <el-checkbox
              :model-value="selectedBatchDeleteCards.includes(prompt.id)"
              @click.stop
              @change="toggleBatchDeleteCard(prompt.id)"
            />
            <div class="batch-delete-card-content">
              <div class="batch-delete-card-name">{{ prompt.name }}</div>
              <div class="batch-delete-card-preview">
                <template v-if="prompt.card_type === 'encrypted'">
                  <el-icon><Lock /></el-icon> 已加密
                </template>
                <template v-else>
                  {{ prompt.content.slice(0, 30) }}{{ prompt.content.length > 30 ? '...' : '' }}
                </template>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="batch-delete-empty">
          <el-empty description="该卡包暂无提示词" />
        </div>
      </div>
      <template #footer>
        <el-button @click="batchDeleteDialogVisible = false">取消</el-button>
        <el-button
          type="danger"
          :disabled="selectedBatchDeleteCards.length === 0"
          @click="executeBatchDelete"
        >
          删除 ({{ selectedBatchDeleteCards.length }})
        </el-button>
      </template>
    </el-dialog>

    <!-- 移动卡片到其他卡包对话框 -->
    <el-dialog
      v-model="movePromptDialogVisible"
      title="移动到其他卡包"
      width="500px"
      :close-on-click-modal="false"
      append-to-body
    >
      <div class="move-prompt-content">
        <div class="move-prompt-hint">
          <el-alert
            title="选择目标卡包"
            type="info"
            :closable="false"
            show-icon
          >
            将卡片"{{ movePromptData?.name }}"移动到以下卡包：
          </el-alert>
        </div>
        <div class="move-prompt-options">
          <el-radio-group v-model="moveTargetCategory" class="move-prompt-radio-group">
            <el-radio
              v-for="category in categoryList"
              :key="category.name"
              :label="category.name"
              :disabled="category.name === movePromptData?.category"
              class="move-prompt-radio-item"
            >
              <div class="move-prompt-category-info">
                <el-icon><FolderOpened /></el-icon>
                <span>{{ category.name }}</span>
                <el-tag size="small" type="info">{{ category.prompts.length }} 张</el-tag>
              </div>
            </el-radio>
          </el-radio-group>
        </div>
      </div>
      <template #footer>
        <el-button @click="movePromptDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :disabled="!moveTargetCategory || moveTargetCategory === movePromptData?.category"
          @click="executeMovePrompt"
        >
          移动
        </el-button>
      </template>
    </el-dialog>

    <!-- 独立卡片批量删除对话框 -->
    <el-dialog
      v-model="standaloneBatchDeleteDialogVisible"
      title="批量删除独立卡片"
      width="800px"
      :close-on-click-modal="false"
      class="batch-delete-dialog"
      append-to-body
    >
      <div class="batch-delete-content">
        <div class="batch-delete-hint">
          <el-alert
            title="正在删除独立卡片"
            type="warning"
            :closable="false"
            show-icon
          >
            勾选需要删除的独立卡片，然后点击删除按钮。此操作不可恢复。
          </el-alert>
        </div>
        <div class="batch-delete-select-all">
          <el-checkbox
            v-model="selectAllStandaloneBatchDelete"
            @change="handleSelectAllStandaloneBatchDelete"
          >
            全选 ({{ selectedStandaloneBatchDeleteCards.length }}/{{ combinedUncategorizedPrompts.length }})
          </el-checkbox>
        </div>
        <div v-if="combinedUncategorizedPrompts.length > 0" class="batch-delete-cards-grid">
          <div
            v-for="prompt in combinedUncategorizedPrompts"
            :key="prompt.id"
            class="batch-delete-card-item"
            :class="{ selected: selectedStandaloneBatchDeleteCards.includes(prompt.id) }"
            @click="toggleStandaloneBatchDeleteCard(prompt.id)"
          >
            <el-checkbox
              :model-value="selectedStandaloneBatchDeleteCards.includes(prompt.id)"
              @click.stop
              @change="toggleStandaloneBatchDeleteCard(prompt.id)"
            />
            <div class="batch-delete-card-content">
              <div class="batch-delete-card-name">
                {{ prompt.name }}
                <el-tag v-if="prompt.category === '拆书-未分类'" size="small" type="warning" style="margin-left: 6px;">拆书</el-tag>
              </div>
              <div class="batch-delete-card-preview">
                <template v-if="prompt.card_type === 'encrypted'">
                  <el-icon><Lock /></el-icon> 已加密
                </template>
                <template v-else>
                  {{ prompt.content.slice(0, 30) }}{{ prompt.content.length > 30 ? '...' : '' }}
                </template>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="batch-delete-empty">
          <el-empty description="暂无独立卡片" />
        </div>
      </div>
      <template #footer>
        <el-button @click="standaloneBatchDeleteDialogVisible = false">取消</el-button>
        <el-button
          type="danger"
          :disabled="selectedStandaloneBatchDeleteCards.length === 0"
          @click="executeStandaloneBatchDelete"
        >
          删除 ({{ selectedStandaloneBatchDeleteCards.length }})
        </el-button>
      </template>
    </el-dialog>

    <!-- 导入并转换txt/md弹窗 -->
    <el-dialog
      v-model="importConvertDialogVisible"
      title="导入并转换 txt/md 文件"
      width="800px"
      :close-on-click-modal="false"
      class="import-dialog"
      append-to-body
    >
      <div class="import-content">
        <div class="import-upload">
          <el-upload
            ref="convertUploadRef"
            drag
            multiple
            accept=".txt,.md"
            :auto-upload="false"
            :show-file-list="false"
            :on-change="handleConvertFilesChange"
          >
            <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
            <div class="el-upload__text">
              将文件拖到此处，或<em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">支持批量上传 txt 或 md 格式文件</div>
            </template>
          </el-upload>
        </div>
        <div class="import-options">
          <el-switch v-model="mergeAsPack" active-text="合并为卡包" inactive-text="" />
          <el-input
            v-if="mergeAsPack"
            v-model="mergePackName"
            placeholder="请输入卡包名称"
            style="margin-top: 10px;"
          />
        </div>
        <div v-if="importConvertPreview.length > 0" class="import-preview">
          <div class="import-preview-header">
            <span>预览导入内容 ({{ importConvertPreview.length }} 个文件)</span>
            <el-checkbox
              v-model="selectAllImportConvert"
              @change="handleSelectAllImportConvert"
            >
              全选
            </el-checkbox>
          </div>
          <div class="import-cards-grid">
            <div
              v-for="(file, index) in importConvertPreview"
              :key="index"
              class="import-card-item"
              :class="{ selected: selectedImportConvertCards.includes(index) }"
              @click="toggleImportConvertCard(index)"
            >
              <el-checkbox
                :model-value="selectedImportConvertCards.includes(index)"
                @click.stop
                @change="toggleImportConvertCard(index)"
              />
              <div class="import-card-content">
                <div class="import-card-name">{{ file.name }}</div>
                <div class="import-card-preview">{{ file.content.slice(0, 50) }}{{ file.content.length > 50 ? '...' : '' }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="importConvertDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :disabled="selectedImportConvertCards.length === 0"
          @click="executeImportConvert"
        >
          导入 ({{ selectedImportConvertCards.length }})
        </el-button>
      </template>
    </el-dialog>

    <!-- 整体备份导入对话框 -->
    <el-dialog
      v-model="importAllDialogVisible"
      title="整体恢复（全部导入）"
      width="600px"
      :close-on-click-modal="false"
      class="import-all-dialog"
      append-to-body
    >
      <div class="import-all-content">
        <div class="import-all-upload">
          <el-upload
            ref="importAllUploadRef"
            drag
            accept=".json"
            :auto-upload="false"
            :show-file-list="false"
            :on-change="handleImportAllFileChange"
          >
            <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
            <div class="el-upload__text">
              将备份文件拖到此处，或<em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">只能上传整体备份的 JSON 格式文件</div>
            </template>
          </el-upload>
        </div>
        <div v-if="importAllPreview" class="import-all-preview">
          <el-alert
            title="警告"
            type="warning"
            :closable="false"
            show-icon
            style="margin-bottom: 16px;"
          >
            导入后将覆盖所有现有数据，包括独立卡片和卡包。此操作不可恢复！
          </el-alert>
          <div class="import-all-info">
            <div class="import-all-item">
              <el-icon><Document /></el-icon>
              <span>独立卡片：<strong>{{ importAllPreview.standaloneCount }}</strong> 张</span>
            </div>
            <div class="import-all-item">
              <el-icon><FolderOpened /></el-icon>
              <span>卡包：<strong>{{ importAllPreview.packCount }}</strong> 个</span>
            </div>
            <div class="import-all-item">
              <el-icon><Collection /></el-icon>
              <span>总提示词数：<strong>{{ importAllPreview.totalPrompts }}</strong> 条</span>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="importAllDialogVisible = false">取消</el-button>
        <el-button
          type="danger"
          :disabled="!importAllPreview"
          @click="executeImportAll"
        >
          导入并覆盖
        </el-button>
      </template>
    </el-dialog>

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
      <!-- 顶部操作栏（对标灵犀） -->
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
            <!-- 空状态（对齐灵犀） -->
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

    <!-- 提示词预览测试弹窗（对标灵犀） -->
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

    <!-- 生成参数设置弹窗（对标灵犀） -->
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

    <!-- 使用说明弹窗（对标灵犀引导教程） -->
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

    <!-- 牌桌预览 -->
    <Teleport to="body">
      <transition name="pack-fan-stage">
        <div v-if="packPreviewVisible" class="pack-fan-stage" @click.self="closePackPreview">
          <div class="pack-fan-board">
            <div v-if="packPreviewPrompts.length > 0" class="pack-fan-row">
              <button
                v-for="prompt in packPreviewPrompts"
                :key="prompt.id"
                type="button"
                class="pack-fan-card"
                @click="handlePackPreviewItemClick(prompt)"
              >
                <div class="pack-fan-card-head">
                  <el-icon class="pack-fan-card-head-icon"><Document /></el-icon>
                  <span class="pack-fan-card-head-name">{{ prompt.name }}</span>
                  <el-icon v-if="prompt.card_type === 'encrypted'" class="pack-fan-card-head-lock"><Lock /></el-icon>
                </div>
                <div class="pack-fan-card-body">
                  <el-tag size="small" :type="getTagType(prompt.category)" class="pack-fan-card-tag">{{ prompt.category }}</el-tag>
                  <div class="pack-fan-card-content" :class="{ 'is-empty': !prompt.description }">
                    <template v-if="prompt.card_type === 'encrypted'">
                      <span class="pack-fan-card-content-encrypted">
                        <el-icon><Lock /></el-icon> 内容已加密
                      </span>
                    </template>
                    <template v-else>
                      {{ prompt.description || '暂无简介' }}
                    </template>
                  </div>
                </div>
                <div class="pack-fan-card-foot">
                  <span class="pack-fan-card-foot-time">{{ formatDate(prompt.created_at) }}</span>
                  <div class="pack-fan-card-actions" @click.stop>
                    <el-tooltip content="导出卡片" placement="top" :show-after="300">
                      <el-button class="pack-fan-action-btn" size="small" circle @click="handleExportSinglePrompt(prompt)">
                        <el-icon><Download /></el-icon>
                      </el-button>
                    </el-tooltip>
                    <el-tooltip content="移动到其他卡包" placement="top" :show-after="300">
                      <el-button class="pack-fan-action-btn" size="small" circle @click="openMovePromptDialog(prompt)">
                        <el-icon><FolderRemove /></el-icon>
                      </el-button>
                    </el-tooltip>
                    <el-tooltip content="删除" placement="top" :show-after="300">
                      <el-button class="pack-fan-action-btn pack-fan-action-btn--danger" size="small" circle @click="handlePackPreviewDelete(prompt)">
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </el-tooltip>
                  </div>
                </div>
              </button>
            </div>
            <el-empty v-else description="该卡包暂无可预览的卡片" />
          </div>
        </div>
      </transition>
    </Teleport>

    <!-- 预览弹窗 -->
    <el-dialog
      v-model="previewDialogVisible"
      title="生成结果预览"
      width="50vw"
      destroy-on-close
      class="preview-dialog"
      append-to-body
    >
      <div class="preview-dialog-content">
        <div class="preview-dialog-header">
          <el-button type="primary" link @click="copyPreviewContent">
            <el-icon><CopyDocument /></el-icon>
            复制
          </el-button>
        </div>
        <el-scrollbar class="preview-scrollbar">
          <MarkdownRenderer :content="previewContent" />
        </el-scrollbar>
      </div>
    </el-dialog>

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
import { ref, onMounted, onBeforeUnmount, watch, computed, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, ArrowUp, ArrowDown, Delete, CopyDocument, FolderAdd, FolderOpened, FolderRemove, Edit, Document, Rank, Download, Upload, UploadFilled, View, Switch, Lock, Reading, MoreFilled, Search, MagicStick, FullScreen, DocumentAdd, Setting, QuestionFilled, InfoFilled, WarningFilled, Tickets } from '@element-plus/icons-vue'
import { promptAPI } from '@/api'
import type { Prompt } from '@/types'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import SplitRichTextEditor from '@/components/SplitRichTextEditor.vue'

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
  // 灵犀式 metadata 相关（存 Prompt.metadata）
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

// ============ 灵犀式新功能状态 ============
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

// 预览测试：正文分段（未填=占位符橙色，已填=替换绿色）
const previewSegments = computed(() => {
  const content = formData.value.content || ''
  const regex = /\$\{([^}]+)\}/g
  const segments: Array<{ type: 'text' | 'unfilled' | 'filled'; text: string }> = []
  let lastIndex = 0
  let match
  while ((match = regex.exec(content)) !== null) {
    if (match.index > lastIndex) {
      segments.push({ type: 'text', text: content.slice(lastIndex, match.index) })
    }
    const fieldName = match[1].trim()
    const field = previewFields.value.find(f => f.name === fieldName)
    const value = field?.value?.trim() || ''
    if (value) {
      segments.push({ type: 'filled', text: value })
    } else {
      segments.push({ type: 'unfilled', text: `\${${fieldName}}` })
    }
    lastIndex = match.index + match[0].length
  }
  if (lastIndex < content.length) {
    segments.push({ type: 'text', text: content.slice(lastIndex) })
  }
  return segments
})

const unfilledFields = computed(() => {
  const content = formData.value.content || ''
  const regex = /\$\{([^}]+)\}/g
  const names = new Set<string>()
  let match
  while ((match = regex.exec(content)) !== null) {
    if (match && match[1]) names.add(match[1].trim())
  }
  return Array.from(names).filter(name => {
    const field = previewFields.value.find(f => f.name === name)
    return !field?.value?.trim()
  })
})

// 配置了但未在模板正文中引用的字段（对齐灵犀提示）
const unusedFields = computed(() => {
  const content = formData.value.content || ''
  const regex = /\$\{([^}]+)\}/g
  const used = new Set<string>()
  let match
  while ((match = regex.exec(content)) !== null) {
    if (match && match[1]) used.add(match[1].trim())
  }
  return fieldsConfig.value.filter(f => f.name && !used.has(f.name)).map(f => f.name)
})

const EXPORT_MAGIC = 'XNP1'
const encodeExportData = (data: string): string => {
  const base64 = btoa(unescape(encodeURIComponent(data)))
  const shifted = base64.split('').map(c => String.fromCharCode(c.charCodeAt(0) + 3)).join('')
  return EXPORT_MAGIC + shifted
}

const decodeExportData = (data: string): string => {
  if (!data.startsWith(EXPORT_MAGIC)) {
    return data
  }
  const shifted = data.slice(EXPORT_MAGIC.length).split('').map(c => String.fromCharCode(c.charCodeAt(0) - 3)).join('')
  return decodeURIComponent(escape(atob(shifted)))
}

const hashPassword = async (password: string): Promise<string> => {
  const encoder = new TextEncoder()
  const data = encoder.encode(password + '__xingnovel_salt__')
  const hashBuffer = await crypto.subtle.digest('SHA-256', data)
  const hashArray = Array.from(new Uint8Array(hashBuffer))
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
}

const verifyPassword = async (inputPassword: string, storedHash: string): Promise<boolean> => {
  const inputHash = await hashPassword(inputPassword)
  return inputHash === storedHash
}

const bufferToBase64 = (buffer: ArrayBuffer): string => {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  bytes.forEach(b => binary += String.fromCharCode(b))
  return btoa(binary)
}

const base64ToBuffer = (base64: string): ArrayBuffer => {
  const binary = atob(base64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes.buffer
}

const ENCRYPT_MARKER = 'XNC1:'

const encryptContent = async (plaintext: string, password: string): Promise<string> => {
  const enc = new TextEncoder()
  const salt = crypto.getRandomValues(new Uint8Array(16))
  const keyMaterial = await crypto.subtle.importKey('raw', enc.encode(password), 'PBKDF2', false, ['deriveKey'])
  const key = await crypto.subtle.deriveKey(
    { name: 'PBKDF2', salt, iterations: 100000, hash: 'SHA-256' },
    keyMaterial,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt']
  )
  const iv = crypto.getRandomValues(new Uint8Array(12))
  const ciphertext = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, key, enc.encode(plaintext))
  const combined = new Uint8Array(16 + 12 + ciphertext.byteLength)
  combined.set(salt, 0)
  combined.set(iv, 16)
  combined.set(new Uint8Array(ciphertext), 28)
  return ENCRYPT_MARKER + bufferToBase64(combined.buffer)
}

const decryptContent = async (wrapped: string, password: string): Promise<string | null> => {
  if (!wrapped.startsWith(ENCRYPT_MARKER)) {
    return wrapped
  }
  try {
    const enc = new TextEncoder()
    const dec = new TextDecoder()
    const combined = new Uint8Array(base64ToBuffer(wrapped.slice(ENCRYPT_MARKER.length)))
    const salt = combined.slice(0, 16)
    const iv = combined.slice(16, 28)
    const ciphertext = combined.slice(28)
    const keyMaterial = await crypto.subtle.importKey('raw', enc.encode(password), 'PBKDF2', false, ['deriveKey'])
    const key = await crypto.subtle.deriveKey(
      { name: 'PBKDF2', salt, iterations: 100000, hash: 'SHA-256' },
      keyMaterial,
      { name: 'AES-GCM', length: 256 },
      false,
      ['decrypt']
    )
    const plaintext = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, key, ciphertext)
    return dec.decode(plaintext)
  } catch {
    return null
  }
}

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

// 添加小分类到表单
const addFormSubcategory = () => {
  const value = newSubcategory.value.trim()
  if (value && !formSubcategories.value.includes(value)) {
    formSubcategories.value.push(value)
    newSubcategory.value = ''
  }
}

// 从表单中删除小分类
const removeFormSubcategory = (index: number) => {
  formSubcategories.value.splice(index, 1)
}

// 单独保存小分类（仅保存标签，不保存其他内容）
const saveSubcategoriesOnly = async () => {
  if (!isEdit.value || !formData.value.id) {
    ElMessage.warning('请先保存提示词基本信息后再保存标签')
    return
  }
  
  try {
    let contentToSave = formData.value.content
    if (formData.value.card_type === 'encrypted' && currentSessionPassword.value) {
      contentToSave = await encryptContent(contentToSave, currentSessionPassword.value)
    }
    await promptAPI.update(formData.value.id, {
      name: formData.value.name,
      description: formData.value.description,
      content: contentToSave,
      category: formData.value.category,
      order_num: formData.value.order_num,
      creator_name: formData.value.creator_name,
      version: formData.value.version,
      fields: fieldsConfig.value,
      subcategories: formSubcategories.value,
      card_type: formData.value.card_type || 'normal',
      password: formData.value.password || null
    })
    ElMessage.success('标签保存成功')
    await fetchPrompts()
  } catch (error) {
    ElMessage.error('标签保存失败')
  }
}

// 卡包相关状态
const categoryDialogVisible = ref(false)
const categoryForm = ref({
  name: ''
})

// 导出独立卡片相关状态
const exportStandaloneDialogVisible = ref(false)
const selectedStandaloneCards = ref<number[]>([])
const selectAllStandalone = ref(false)

// 导出卡包相关状态
const exportPackDialogVisible = ref(false)
const selectedPack = ref<string | null>(null)

// 导入独立卡片相关状态
const importStandaloneDialogVisible = ref(false)
const importStandalonePreview = ref<Prompt[]>([])
const selectedImportCards = ref<number[]>([])
const selectAllImportStandalone = ref(false)

// 导入卡包相关状态
const importPackDialogVisible = ref(false)
const importPackPreview = ref<{ packName: string; prompts: Prompt[]; isBookAnalysis: boolean } | null>(null)

// 批量删除相关状态
const batchDeleteDialogVisible = ref(false)
const currentCategory = ref<string>('')
const selectedBatchDeleteCards = ref<number[]>([])
const selectAllBatchDelete = ref(false)
const packPreviewVisible = ref(false)
const packPreviewCategory = ref('')

// 移动卡片相关状态
const movePromptDialogVisible = ref(false)
const movePromptData = ref<Prompt | null>(null)
const moveTargetCategory = ref('')

// 独立卡片批量删除相关状态
const standaloneBatchDeleteDialogVisible = ref(false)
const selectedStandaloneBatchDeleteCards = ref<number[]>([])
const selectAllStandaloneBatchDelete = ref(false)

// 导入转换 txt/md 相关状态
const importConvertDialogVisible = ref(false)
const importConvertPreview = ref<{ name: string; content: string }[]>([])
const selectedImportConvertCards = ref<number[]>([])
const selectAllImportConvert = ref(false)
const convertUploadRef = ref<any>()
const mergeAsPack = ref(false)
const mergePackName = ref('')

// 整体备份导入相关状态
const importAllDialogVisible = ref(false)
const importAllPreview = ref<{
  standaloneCount: number
  packCount: number
  totalPrompts: number
  data: any
} | null>(null)
const importAllUploadRef = ref<any>()

// 卡片尺寸状态
const cardSize = ref<'small' | 'medium' | 'large'>('medium')
const packViewMode = ref<'grid' | 'list'>('grid')
const activeMainTab = ref<'all' | 'pack' | 'standalone' | 'bookAnalysis'>('pack')
const sheetVisible = ref(false)

// 从localStorage加载卡包列表
const loadCategoriesFromStorage = (): string[] => {
  const stored = localStorage.getItem('prompt_categories')
  return stored ? JSON.parse(stored) : []
}

// 保存卡包列表到localStorage
const saveCategoriesToStorage = (categories: string[]) => {
  localStorage.setItem('prompt_categories', JSON.stringify(categories))
}

// 卡包列表（从localStorage加载）
const DEFAULT_CATEGORIES = ['默认', '写作要求', '写作风格', '续写', '脑洞', '书名', '简介', '大纲', '细纲', '黄金开篇', '金手指', '名字', '人设', '世界观', '卡包']

const customCategories = ref<string[]>(loadCategoriesFromStorage())

// 拖拽相关状态
const draggedPrompt = ref<Prompt | null>(null)
const isDragging = ref(false)

// 未分类的提示词（不在任何自定义卡包中，且不属于拆书分类）
const uncategorizedPrompts = computed(() => {
  return prompts.value.filter(prompt => {
    const category = prompt.category || '默认'
    // 排除拆书分类
    if (category.startsWith('拆书-')) return false
    return category === '未分类' || (!customCategories.value.includes(category) && category !== '默认')
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

const totalPackCards = computed(() => {
  return categoryList.value.reduce((sum, cat) => sum + cat.prompts.length, 0)
})

// ===== 拆书提示词相关 =====
const BOOK_ANALYSIS_PREFIX = '拆书-'

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

// 拆书独立卡片（未分类且属于拆书来源的，或 category 为 "拆书-未分类"）
const bookAnalysisUncategorizedPrompts = computed(() => {
  return prompts.value.filter(prompt => {
    const category = prompt.category || ''
    return category === '拆书-未分类'
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
    case 'preview':
      if (promptsCount > 0) {
        openPackPreviewDialog(categoryName)
      }
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

const getCategoryPromptsByName = (categoryName: string) => {
  if (!categoryName) return []
  const category = categoryList.value.find(cat => cat.name === categoryName)
  return category ? category.prompts : []
}

// 当前卡包的提示词列表（用于批量删除）
const currentCategoryPrompts = computed(() => {
  return getCategoryPromptsByName(currentCategory.value)
})

const packPreviewPrompts = computed(() => {
  return getCategoryPromptsByName(packPreviewCategory.value)
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

// 获取标签类型
const getTagType = (category: string): any => {
  const typeMap: Record<string, any> = {
    '角色': 'success',
    '剧情': 'warning',
    '对话': 'primary',
    '场景': 'info'
  }
  return typeMap[category] || 'info'
}

// 格式化日期
const formatDate = (dateStr: string | undefined) => {
  if (!dateStr) return '未知时间'
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
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
    case 'preview':
      if (promptsCount > 0) {
        openPackPreviewDialog(categoryName)
      }
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
const fieldsConfig = ref<Array<{
  name: string;
  label: string;
  type: 'text' | 'textarea' | 'select';
  options: string[];
  optionLabels: string[];
  optionsText: string;
  optionLabelsText: string;
  description: string;
  required: boolean;
  defaultValue: string;
}>>([])

// 字段类型标签映射
const fieldTypeLabel = (type: string): string => {
  const map: Record<string, string> = { text: '单行文本', textarea: '多行文本', select: '下拉选择' }
  return map[type] || type
}

// 字段卡片展开/收起
const toggleFieldExpand = (index: number) => {
  fieldExpanded.value = fieldExpanded.value === index ? null : index
}

// 预览并保存字段配置（收起卡片）
const saveFieldExpand = (index: number) => {
  const field = fieldsConfig.value[index]
  if (!field) return
  if (!field.name.trim()) {
    ElMessage.warning('字段名称不能为空')
    return
  }
  fieldExpanded.value = null
  ElMessage.success(`字段「${field.name}」已保存`)
}

// ============ 插入动态字段（对标灵犀） ============
// 常用字段预置
const COMMON_FIELDS = ['期望字数', '参考文风', '角色设定', '背景设定', '剧情方向', '输出格式']

const insertFieldPopoverVisible = ref(false)
const enlargeInsertPopoverVisible = ref(false)
const customFieldName = ref('')
const enlargeEditVisible = ref(false)
const contentTextareaRef = ref<{ $el: HTMLDivElement } | null>(null)
const enlargeTextareaRef = ref<{ $el: HTMLDivElement } | null>(null)

// 获取 textarea 原生元素（el-input ref 的 $el 是包裹 div，内部有 textarea）
function getTextareaEl(refItem: { $el: HTMLDivElement } | null): HTMLTextAreaElement | null {
  if (!refItem?.$el) return null
  return refItem.$el.querySelector('textarea')
}

// 在指定 textarea 的光标位置插入 ${字段名}
function insertFieldToContent(fieldName: string, targetRef?: { $el: HTMLDivElement } | null) {
  const name = (fieldName || '').trim()
  if (!name) return
  const placeholder = `\${${name}}`
  const ta = getTextareaEl(targetRef ?? contentTextareaRef.value) ?? getTextareaEl(contentTextareaRef.value)

  if (ta && ta.selectionStart !== undefined && ta.selectionStart !== null) {
    const start = ta.selectionStart
    const end = ta.selectionEnd
    const content = formData.value.content || ''
    formData.value.content = content.slice(0, start) + placeholder + content.slice(end)
    // 恢复焦点与光标位置（插入点之后）
    nextTick(() => {
      ta.focus()
      const pos = start + placeholder.length
      ta.setSelectionRange(pos, pos)
    })
  } else {
    formData.value.content = (formData.value.content || '') + placeholder
  }
  insertFieldPopoverVisible.value = false
  enlargeInsertPopoverVisible.value = false
  customFieldName.value = ''
}

// 插入自定义字段
function insertCustomField(targetRef?: { $el: HTMLDivElement } | null) {
  const name = customFieldName.value.trim()
  if (!name) {
    ElMessage.warning('请输入字段名称')
    return
  }
  if (/\$\{/.test(name) || /\}/.test(name)) {
    ElMessage.warning('字段名称不能包含 ${} 符号')
    return
  }
  insertFieldToContent(name, targetRef)
}

// ============ 顶部操作栏功能（对标灵犀） ============

// 粘贴导入：读取剪贴板文本填入正文
const handlePasteImport = async () => {
  try {
    const text = await navigator.clipboard.readText()
    if (!text || !text.trim()) {
      ElMessage.warning('剪贴板中没有可导入的文本')
      return
    }
    // 追加到正文末尾（保留已有内容）
    formData.value.content = formData.value.content
      ? formData.value.content + '\n' + text
      : text
    ElMessage.success('已从剪贴板导入')
  } catch (e) {
    ElMessage.error('读取剪贴板失败，请检查浏览器权限')
  }
}

// 打开预览测试：初始化字段表单
const openPreviewTest = () => {
  previewFields.value = fieldsConfig.value.map(field => ({
    name: field.name,
    label: field.label || field.name,
    type: field.type,
    options: field.options || [],
    value: field.defaultValue || ''
  }))
  previewTestVisible.value = true
}

// 复制预览结果（已填字段替换后的完整提示词）
const copyPreviewResult = async () => {
  const filled = previewFields.value.filter(f => f.value.trim()).map(f => ({
    name: f.name,
    value: f.value.trim()
  }))
  let result = formData.value.content || ''
  filled.forEach(f => {
    result = result.replace(new RegExp(`\\$\\{${f.name}\\}`, 'g'), f.value)
  })
  try {
    await navigator.clipboard.writeText(result)
    ElMessage.success('已复制')
  } catch (e) {
    ElMessage.error('复制失败')
  }
}

// 保存并打开预览测试（对齐灵犀"预览并保存"）
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

// 对话框大小状态
const dialogSize = ref<'small' | 'medium' | 'large'>('medium')

// 对话框宽度映射
const dialogWidthMap = {
  small: '900px',
  medium: '1200px',
  large: '1500px'
}

// 对话框高度映射
const dialogHeightMap = {
  small: '500px',
  medium: '600px',
  large: '700px'
}

// 预览弹窗相关状态
const previewDialogVisible = ref(false)
const previewContent = ref('')

// 复制预览内容到剪贴板
const copyPreviewContent = async () => {
  try {
    await navigator.clipboard.writeText(previewContent.value)
    ElMessage.success('内容已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

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
  previewContent.value = prompt.content
  previewDialogVisible.value = true
}

onMounted(async () => {
  window.addEventListener('keydown', handlePackPreviewKeydown)
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

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handlePackPreviewKeydown)
})

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
        previewContent.value = decrypted !== null ? decrypted : promptToProcess.content
        previewDialogVisible.value = true
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

// 从提示词内容中提取字段名称（幂等：只追加不存在的字段）
const extractFieldsFromContent = (content: string) => {
  if (!content) {
    ElMessage.info('正文为空，无可提取的字段')
    return
  }
  const fieldRegex = /\$\{([^}]+)\}/g
  const fields = new Set<string>()
  let match
  while ((match = fieldRegex.exec(content)) !== null) {
    if (match && match[1]) {
      fields.add(match[1].trim())
    }
  }
  
  if (fields.size === 0) {
    ElMessage.info('正文中没有找到 ${字段名} 占位符')
    return
  }
  
  // 收集现有字段名称，避免重复添加
  const existingNames = new Set(fieldsConfig.value.map(f => f.name))
  let addedCount = 0
  
  // 为每个新字段创建配置
  fields.forEach(fieldName => {
    if (!existingNames.has(fieldName)) {
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
      addedCount++
    }
  })
  
  if (addedCount > 0) {
    ElMessage.success(`已提取 ${addedCount} 个字段`)
  } else {
    ElMessage.info('所有字段均已配置')
  }
}

// 添加字段
const addField = () => {
  fieldsConfig.value.push({
    name: `字段${fieldsConfig.value.length + 1}`,
    label: `字段${fieldsConfig.value.length + 1}`,
    type: 'text',
    options: [],
    optionLabels: [],
    optionsText: '',
    optionLabelsText: '',
    description: '',
    required: true,
    defaultValue: ''
  })
  fieldExpanded.value = fieldsConfig.value.length - 1
}

// 删除字段
const removeField = (index: number) => {
  fieldsConfig.value.splice(index, 1)
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

// 验证字段配置
const validateFieldsConfig = () => {
  // 检查是否所有字段都有名称
  for (const field of fieldsConfig.value) {
    if (!field.name.trim()) {
      return '所有字段必须有名称'
    }
    
    // 检查下拉选择类型的字段是否有选项
    if (field.type === 'select' && field.options.length === 0) {
      return `字段"${field.name}"是下拉选择类型，必须添加选项`
    }
  }
  
  // 检查字段名称是否唯一
  const fieldNames = fieldsConfig.value.map(field => field.name.trim())
  const uniqueNames = new Set(fieldNames)
  if (fieldNames.length !== uniqueNames.size) {
    return '字段名称必须唯一'
  }
  
  return ''
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
      // 灵犀式扩展信息（存 metadata，不修改后端 schema）
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

// 导出命令处理
const handleExportCommand = (command: string) => {
  if (command === 'exportAll') {
    executeExportAll()
  } else if (command === 'exportStandalone') {
    openExportStandaloneDialog()
  } else if (command === 'exportPack') {
    openExportPackDialog()
  }
}

// 导入命令处理
const handleImportCommand = (command: string) => {
  if (command === 'importAll') {
    openImportAllDialog()
  } else if (command === 'importStandalone') {
    openImportStandaloneDialog()
  } else if (command === 'importPack') {
    openImportPackDialog()
  } else if (command === 'importConvert') {
    openImportConvertDialog()
  }
}

// 打开批量删除对话框
const openBatchDeleteDialog = (categoryName: string) => {
  currentCategory.value = categoryName
  selectedBatchDeleteCards.value = []
  selectAllBatchDelete.value = false
  batchDeleteDialogVisible.value = true
}

const openPackPreviewDialog = (categoryName: string) => {
  if (packPreviewVisible.value && packPreviewCategory.value === categoryName) {
    packPreviewVisible.value = false
    return
  }

  packPreviewCategory.value = categoryName
  packPreviewVisible.value = true
}

const handlePackPreviewItemClick = (prompt: Prompt) => {
  handleEdit(prompt)
}

const closePackPreview = () => {
  packPreviewVisible.value = false
}

const handlePackPreviewKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    packPreviewVisible.value = false
  }
}

const handlePackPreviewDelete = async (prompt: Prompt) => {
  if (prompt.card_type === 'encrypted' && prompt.password) {
    pendingEditPrompt.value = prompt
    passwordAction.value = 'delete'
    passwordDialogVisible.value = true
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定删除提示词"${prompt.name}"吗？`,
      '提示',
      { type: 'warning' }
    )
    await promptAPI.delete(prompt.id)
    await fetchPrompts()
    ElMessage.success('删除成功')
    if (packPreviewPrompts.value.length === 0) {
      packPreviewVisible.value = false
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleExportSinglePrompt = (prompt: Prompt) => {
  const exportData = {
    version: '1.0',
    type: 'single-prompt',
    exportTime: new Date().toISOString(),
    prompt: {
      id: prompt.id,
      name: prompt.name,
      content: prompt.content,
      description: prompt.description,
      category: prompt.category,
      created_at: prompt.created_at,
      updated_at: prompt.updated_at,
      card_type: prompt.card_type,
      fields: prompt.fields
    }
  }
  
  const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `prompt-${prompt.name.replace(/[^\u4e00-\u9fa5a-zA-Z0-9]/g, '_')}.json`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('导出成功')
}

const openMovePromptDialog = (prompt: Prompt) => {
  movePromptData.value = prompt
  moveTargetCategory.value = ''
  movePromptDialogVisible.value = true
}

const executeMovePrompt = async () => {
  if (!movePromptData.value || !moveTargetCategory.value) return
  
  const prompt = movePromptData.value
  const oldCategory = prompt.category || '默认'
  
  if (moveTargetCategory.value === oldCategory) {
    ElMessage.warning('卡片已在该卡包中')
    return
  }
  
  try {
    await promptAPI.update(prompt.id, { ...prompt, category: moveTargetCategory.value })
    await fetchPrompts()
    ElMessage.success(`已移动到"${moveTargetCategory.value}"`)
    movePromptDialogVisible.value = false
    if (packPreviewPrompts.value.length === 0) {
      packPreviewVisible.value = false
    }
  } catch (error) {
    ElMessage.error('移动失败')
  }
}

// 全选/取消全选批量删除卡片
const handleSelectAllBatchDelete = (val: boolean) => {
  if (val) {
    selectedBatchDeleteCards.value = currentCategoryPrompts.value.map(p => p.id)
  } else {
    selectedBatchDeleteCards.value = []
  }
}

// 切换单个批量删除卡片选择
const toggleBatchDeleteCard = (id: number) => {
  const index = selectedBatchDeleteCards.value.indexOf(id)
  if (index > -1) {
    selectedBatchDeleteCards.value.splice(index, 1)
  } else {
    selectedBatchDeleteCards.value.push(id)
  }
  selectAllBatchDelete.value = selectedBatchDeleteCards.value.length === currentCategoryPrompts.value.length
}

// 执行批量删除
const executeBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定删除选中的 ${selectedBatchDeleteCards.value.length} 张卡片吗？此操作不可恢复。`,
      '批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 批量删除选中的卡片
    const deletePromises = selectedBatchDeleteCards.value.map(id => promptAPI.delete(id))
    await Promise.all(deletePromises)
    
    ElMessage.success(`成功删除 ${selectedBatchDeleteCards.value.length} 张卡片`)
    batchDeleteDialogVisible.value = false
    await fetchPrompts()
  } catch (error) {
    // 取消删除
  }
}

// 打开独立卡片批量删除对话框
const openStandaloneBatchDeleteDialog = () => {
  selectedStandaloneBatchDeleteCards.value = []
  selectAllStandaloneBatchDelete.value = false
  standaloneBatchDeleteDialogVisible.value = true
}

// 全选/取消全选独立卡片批量删除
const handleSelectAllStandaloneBatchDelete = (val: boolean) => {
  if (val) {
    selectedStandaloneBatchDeleteCards.value = combinedUncategorizedPrompts.value.map(p => p.id)
  } else {
    selectedStandaloneBatchDeleteCards.value = []
  }
}

// 切换单个独立卡片批量删除选择
const toggleStandaloneBatchDeleteCard = (id: number) => {
  const index = selectedStandaloneBatchDeleteCards.value.indexOf(id)
  if (index > -1) {
    selectedStandaloneBatchDeleteCards.value.splice(index, 1)
  } else {
    selectedStandaloneBatchDeleteCards.value.push(id)
  }
  selectAllStandaloneBatchDelete.value = selectedStandaloneBatchDeleteCards.value.length === combinedUncategorizedPrompts.value.length
}

// 执行独立卡片批量删除
const executeStandaloneBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定删除选中的 ${selectedStandaloneBatchDeleteCards.value.length} 张独立卡片吗？此操作不可恢复。`,
      '批量删除独立卡片',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const deletePromises = selectedStandaloneBatchDeleteCards.value.map(id => promptAPI.delete(id))
    await Promise.all(deletePromises)
    
    ElMessage.success(`成功删除 ${selectedStandaloneBatchDeleteCards.value.length} 张独立卡片`)
    standaloneBatchDeleteDialogVisible.value = false
    await fetchPrompts()
  } catch (error) {
    // 取消删除
  }
}

// 打开整体导入对话框
const openImportAllDialog = () => {
  importAllPreview.value = null
  importAllDialogVisible.value = true
}

// 处理整体备份文件选择（统一入口：新版完整备份 + 旧版/拆书完整备份）
const handleImportAllFileChange = async (file: any) => {
  try {
    const text = await file.raw.text()
    const decodedText = decodeExportData(text)
    const backupData = JSON.parse(decodedText)

    let standalone: any[] = []
    let packs: Record<string, any[]> = {}
    let isBookAnalysis = false

    // 新版完整备份（普通 + 拆书统一包含在内）
    if (backupData.type === 'full-backup' && backupData.data) {
      standalone = backupData.data.standalone || []
      packs = backupData.data.packs || {}
    } else if (backupData.type === 'book-analysis-full-backup' && backupData.data) {
      // 旧版拆书完整备份
      standalone = backupData.data.standalone || []
      packs = backupData.data.packs || {}
      isBookAnalysis = true
    } else if (Array.isArray(backupData)) {
      // 旧版格式1：直接 prompts 数组
      standalone = backupData.filter((p: any) => p && p.name && p.content)
      isBookAnalysis = true
    } else if (backupData.prompts && Array.isArray(backupData.prompts)) {
      // 旧版格式2：{ prompts, categories }
      standalone = backupData.prompts.filter((p: any) => p && p.name && p.content)
      ;(backupData.categories || []).forEach((cat: string) => {
        if (cat && cat !== '未分类') packs[cat] = []
      })
      standalone.forEach((p: any) => {
        const cat = p.category
        if (cat && cat !== '未分类') {
          if (!packs[cat]) packs[cat] = []
          packs[cat].push(p)
        }
      })
      isBookAnalysis = true
    } else {
      throw new Error('无效的备份文件格式')
    }

    const standaloneCount = standalone.length
    const packCount = Object.keys(packs).length
    let totalPrompts = standaloneCount
    Object.values(packs).forEach((pack: any) => {
      totalPrompts += Array.isArray(pack) ? pack.length : 0
    })

    importAllPreview.value = {
      standaloneCount,
      packCount,
      totalPrompts,
      data: { standalone, packs, isBookAnalysis }
    }
  } catch (error: any) {
    ElMessage.error(`文件解析失败：${error.message}`)
    importAllPreview.value = null
  }
}

// 执行整体导入
const executeImportAll = async () => {
  if (!importAllPreview.value) return

  try {
    await ElMessageBox.confirm(
      '确定要导入备份数据吗？这将覆盖所有现有数据，包括独立卡片和卡包。此操作不可恢复！',
      '整体恢复',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'error'
      }
    )

    const { standalone, packs, isBookAnalysis } = importAllPreview.value.data

    // 先删除所有现有数据
    const deletePromises = prompts.value.map(p => promptAPI.delete(p.id))
    await Promise.all(deletePromises)

    // 防重：备份文件内同名（同分类）只保留第一条，避免恢复后仍有重复
    let skipCount = 0
    const importedKeys = new Set<string>()

    // 导入独立卡片
    if (standalone && standalone.length > 0) {
      for (const prompt of standalone) {
        let category = prompt.category || '未分类'
        if (isBookAnalysis) {
          if (category === '未分类') category = '拆书-未分类'
          else if (!category.startsWith(BOOK_ANALYSIS_PREFIX)) category = BOOK_ANALYSIS_PREFIX + category
        }
        const key = buildPromptKey(prompt.name, category)
        if (importedKeys.has(key)) {
          skipCount++
          continue
        }
        importedKeys.add(key)
        await promptAPI.create({
          name: prompt.name,
          content: prompt.content,
          category,
          order_num: prompt.order_num || 0,
          created_at: prompt.created_at,
          fields: prompt.fields || [],
          subcategories: prompt.subcategories || [],
          card_type: prompt.card_type || 'normal',
          password: prompt.password || null
        })
      }
    }

    // 导入卡包
    if (packs) {
      for (const [packName, promptsInPack] of Object.entries(packs)) {
        let fullPackName = packName
        if (isBookAnalysis && !fullPackName.startsWith(BOOK_ANALYSIS_PREFIX)) {
          fullPackName = BOOK_ANALYSIS_PREFIX + fullPackName
        }
        if (Array.isArray(promptsInPack)) {
          for (const prompt of promptsInPack) {
            const key = buildPromptKey(prompt.name, fullPackName)
            if (importedKeys.has(key)) {
              skipCount++
              continue
            }
            importedKeys.add(key)
            await promptAPI.create({
              name: prompt.name,
              content: prompt.content,
              category: fullPackName,
              order_num: prompt.order_num || 0,
              created_at: prompt.created_at,
              fields: prompt.fields || [],
              subcategories: prompt.subcategories || [],
              card_type: prompt.card_type || 'normal',
              password: prompt.password || null
            })
          }
        }
      }
    }

    // 更新自定义卡包列表
    if (packs) {
      customCategories.value = Object.keys(packs).map(p =>
        isBookAnalysis && !p.startsWith(BOOK_ANALYSIS_PREFIX) ? BOOK_ANALYSIS_PREFIX + p : p
      )
      saveCategoriesToStorage(customCategories.value)
    }

    ElMessage.success(skipCount > 0 ? `整体恢复成功（跳过重复 ${skipCount} 张）` : '整体恢复成功')
    importAllDialogVisible.value = false
    importAllPreview.value = null
    await fetchPrompts()
  } catch (error) {
    // 取消导入
  }
}

// ============ 提示词查重（快捷中心） ============
// 判重口径：titleCategory（标题+分类，默认）| title（仅标题）
type DedupMode = 'titleCategory' | 'title'
type DedupGroup = {
  key: string
  name: string
  category?: string
  items: Prompt[]
  keepId: number
}

const dedupDialogVisible = ref(false)
const dedupLoading = ref(false)
const dedupDeleting = ref(false)
const dedupMode = ref<DedupMode>('titleCategory')
const dedupGroups = ref<DedupGroup[]>([])
const dedupSelected = ref<Set<number>>(new Set())

const dedupGroupCount = computed(() => dedupGroups.value.length)
const dedupTotalCount = computed(() => dedupGroups.value.reduce((sum, g) => sum + g.items.length, 0))
const dedupSelectedCount = computed(() => dedupSelected.value.size)

// 判重键：标题（可选+分类）拼接，trim 后比较（导入防重与查重弹窗共用）
const buildPromptKey = (name: string, category: string | undefined, mode: DedupMode = 'titleCategory'): string => {
  const title = (name || '').trim()
  if (mode === 'title') return title
  return `${title}||${(category || '').trim()}`
}

// 从现有提示词列表构建判重键集合
const buildExistingKeySet = (list: Prompt[], mode: DedupMode = 'titleCategory'): Set<string> => {
  const keys = new Set<string>()
  list.forEach(p => {
    const key = buildPromptKey(p.name, p.category, mode)
    if (key) keys.add(key)
  })
  return keys
}

// 重建查重分组（组内按 id 升序，keepId 默认最早一条；默认勾选每组除保留项外的全部非加密卡片）
const rebuildDedupGroups = () => {
  const groupsMap = new Map<string, Prompt[]>()
  prompts.value.forEach(p => {
    const key = buildPromptKey(p.name, p.category, dedupMode.value)
    if (!key) return
    const list = groupsMap.get(key) || []
    list.push(p)
    groupsMap.set(key, list)
  })

  const groups: DedupGroup[] = []
  groupsMap.forEach((items, key) => {
    if (items.length < 2) return
    items.sort((a, b) => a.id - b.id)
    const [first] = items
    groups.push({
      key,
      name: first.name,
      category: dedupMode.value === 'title' ? undefined : first.category,
      items,
      keepId: first.id
    })
  })
  groups.sort((a, b) => b.items.length - a.items.length || a.name.localeCompare(b.name, 'zh-CN'))
  dedupGroups.value = groups

  const selected = new Set<number>()
  groups.forEach(g => {
    g.items.forEach(item => {
      if (item.id !== g.keepId && item.card_type !== 'encrypted') selected.add(item.id)
    })
  })
  dedupSelected.value = selected
}

const openDedupDialog = async () => {
  dedupDialogVisible.value = true
  dedupLoading.value = true
  try {
    // 重新拉取，保证数据最新
    await fetchPrompts()
    rebuildDedupGroups()
  } finally {
    dedupLoading.value = false
  }
}

const toggleDedupItem = (id: number, checked: boolean) => {
  const selected = new Set(dedupSelected.value)
  if (checked) selected.add(id)
  else selected.delete(id)
  dedupSelected.value = selected
}

// 每组保留最早/最新，其余非加密卡片勾选删除
const setDedupKeep = (groupKey: string, mode: 'oldest' | 'newest') => {
  const group = dedupGroups.value.find(g => g.key === groupKey)
  if (!group) return
  const sorted = [...group.items].sort((a, b) => (mode === 'oldest' ? a.id - b.id : b.id - a.id))
  group.keepId = sorted[0]?.id ?? group.keepId
  const selected = new Set(dedupSelected.value)
  group.items.forEach(item => {
    if (item.id === group.keepId || item.card_type === 'encrypted') selected.delete(item.id)
    else selected.add(item.id)
  })
  dedupSelected.value = selected
}

const isDedupKeepItem = (group: DedupGroup, item: Prompt) => item.id === group.keepId

const getDedupContentPreview = (item: Prompt) => {
  if (item.card_type === 'encrypted') return '内容已加密，需解锁后查看'
  const text = (item.content || '').replace(/\s+/g, ' ').trim()
  return text.length > 80 ? text.slice(0, 80) + '…' : text || '（无内容）'
}

const handleDedupDelete = async () => {
  if (dedupSelected.value.size === 0) return
  try {
    await ElMessageBox.confirm(
      `确定删除选中的 ${dedupSelected.value.size} 张重复提示词吗？此操作不可恢复！`,
      '删除重复提示词',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch (error) {
    return // 取消删除
  }

  dedupDeleting.value = true
  let successCount = 0
  let failCount = 0
  try {
    const ids = Array.from(dedupSelected.value)
    const results = await Promise.all(ids.map(id => promptAPI.delete(id)))
    results.forEach(res => (res.success ? successCount++ : failCount++))
    if (failCount === 0) {
      ElMessage.success(`已删除 ${successCount} 张重复提示词`)
    } else {
      ElMessage.warning(`删除完成：成功 ${successCount} 张，失败 ${failCount} 张`)
    }
    await fetchPrompts()
    rebuildDedupGroups()
    if (dedupGroups.value.length === 0) {
      dedupDialogVisible.value = false
    }
  } catch (error) {
    ElMessage.error('删除失败，请重试')
  } finally {
    dedupDeleting.value = false
  }
}

// ============ 卡包合并（快捷中心 · 新建分组） ============
type MergeMode = 'new' | 'absorb'

const mergeDialogVisible = ref(false)
const mergeLoading = ref(false)
const mergeMode = ref<MergeMode>('new')
const mergeNewName = ref('')
const mergeSelectedSources = ref<string[]>([]) // 模式1：选中的源卡包名
const mergeMainName = ref('') // 模式2：主卡包名
const mergeSelectedSubs = ref<string[]>([]) // 模式2：选中的副卡包名

// 判断是否默认卡包（DEFAULT_CATEGORIES 中的名称）
const isDefaultPack = (name: string): boolean => DEFAULT_CATEGORIES.includes(name)

// 可合并的自定义卡包（排除默认分类、卡片数>0，含拆书）
const mergeablePacks = computed(() =>
  allExportablePacks.value
    .filter(pack => !isDefaultPack(pack.name) && pack.prompts.length > 0)
    .map(pack => ({ name: pack.name, prompts: pack.prompts, isBookAnalysis: pack.isBookAnalysis }))
)

// 可作主卡包的选项（默认卡包也可作主包，卡片数>0）
const mergeMainOptions = computed(() =>
  allExportablePacks.value
    .filter(pack => pack.prompts.length > 0)
    .map(pack => ({ name: pack.name, prompts: pack.prompts, isBookAnalysis: pack.isBookAnalysis }))
)

// 模式2可选副卡包：mergeablePacks 排除已选主包
const mergeAvailableSubs = computed(() =>
  mergeablePacks.value.filter(pack => pack.name !== mergeMainName.value)
)

// 预览将移动的卡片总数
const mergePreviewTotal = computed(() => {
  if (mergeMode.value === 'new') {
    return mergeablePacks.value
      .filter(pack => mergeSelectedSources.value.includes(pack.name))
      .reduce((sum, pack) => sum + pack.prompts.length, 0)
  }
  return mergeAvailableSubs.value
    .filter(pack => mergeSelectedSubs.value.includes(pack.name))
    .reduce((sum, pack) => sum + pack.prompts.length, 0)
})

// 是否满足合并条件
const canMerge = computed(() => {
  if (mergeMode.value === 'new') {
    return mergeNewName.value.trim() !== '' && mergeSelectedSources.value.length >= 2
  }
  return mergeMainName.value !== '' && mergeSelectedSubs.value.length >= 1
})

const openMergeDialog = () => {
  mergeMode.value = 'new'
  mergeNewName.value = ''
  mergeSelectedSources.value = []
  mergeMainName.value = ''
  mergeSelectedSubs.value = []
  mergeDialogVisible.value = true
}

// 切换模式时清空选择
watch(mergeMode, () => {
  mergeNewName.value = ''
  mergeSelectedSources.value = []
  mergeMainName.value = ''
  mergeSelectedSubs.value = []
})

const handleMergeExecute = async () => {
  if (!canMerge.value) return

  // 确定目标卡包名与受影响卡包名
  let targetName = ''
  let sourceNames: string[] = []
  let totalCards = 0
  if (mergeMode.value === 'new') {
    targetName = mergeNewName.value.trim()
    sourceNames = [...mergeSelectedSources.value]
    totalCards = mergePreviewTotal.value
    // 校验新名
    if (!targetName) {
      ElMessage.warning('请输入新卡包名称')
      return
    }
    if (isDefaultPack(targetName)) {
      ElMessage.warning('新卡包名称不能与默认分类同名')
      return
    }
    const allNames = allExportablePacks.value.map(p => p.name).concat(customCategories.value)
    if (allNames.includes(targetName)) {
      ElMessage.warning('该卡包名称已存在，请更换')
      return
    }
  } else {
    targetName = mergeMainName.value
    sourceNames = [...mergeSelectedSubs.value]
    totalCards = mergePreviewTotal.value
    if (!targetName) {
      ElMessage.warning('请选择主卡包')
      return
    }
  }

  const targetDisplay = targetName.replace(BOOK_ANALYSIS_PREFIX, '')
  try {
    await ElMessageBox.confirm(
      `确定将 ${sourceNames.length} 个卡包合并到「${targetDisplay}」吗？将移动 ${totalCards} 张卡片，源卡包合并后自动移除。`,
      '卡包合并',
      { confirmButtonText: '开始合并', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return // 取消
  }

  mergeLoading.value = true
  let successCount = 0
  let failCount = 0
  try {
    // 收集需要移动的卡片（源卡包 → 目标卡包）
    const packsToMove = mergeMode.value === 'new'
      ? mergeablePacks.value.filter(p => mergeSelectedSources.value.includes(p.name))
      : mergeAvailableSubs.value.filter(p => mergeSelectedSubs.value.includes(p.name))

    for (const pack of packsToMove) {
      for (const prompt of pack.prompts) {
        try {
          const res = await promptAPI.update(prompt.id, { ...prompt, category: targetName })
          if (res.success) successCount++
          else failCount++
        } catch {
          failCount++
        }
      }
    }

    // 模式1：新卡包加入 customCategories（若非默认分类）
    if (mergeMode.value === 'new' && !isDefaultPack(targetName) && !customCategories.value.includes(targetName)) {
      customCategories.value.push(targetName)
      saveCategoriesToStorage(customCategories.value)
    }

    // 移除已合并的自定义源/副卡包（默认卡包不移除）
    const namesToRemove = new Set(sourceNames.filter(n => !isDefaultPack(n)))
    if (namesToRemove.size > 0) {
      customCategories.value = customCategories.value.filter(n => !namesToRemove.has(n))
      saveCategoriesToStorage(customCategories.value)
    }

    await fetchPrompts()

    if (failCount === 0) {
      ElMessage.success(`合并完成：已移动 ${successCount} 张卡片到「${targetDisplay}」`)
    } else {
      ElMessage.warning(`合并完成：成功 ${successCount} 张，失败 ${failCount} 张`)
    }
    mergeDialogVisible.value = false
  } catch (e) {
    ElMessage.error('合并失败，请重试')
  } finally {
    mergeLoading.value = false
  }
}

// 打开导出独立卡片弹窗
const openExportStandaloneDialog = () => {
  selectedStandaloneCards.value = []
  selectAllStandalone.value = false
  exportStandaloneDialogVisible.value = true
}

// 全选/取消全选独立卡片
const handleSelectAllStandalone = (val: boolean) => {
  if (val) {
    selectedStandaloneCards.value = combinedUncategorizedPrompts.value.map(p => p.id)
  } else {
    selectedStandaloneCards.value = []
  }
}

// 切换单个独立卡片选择
const toggleStandaloneCard = (id: number) => {
  const index = selectedStandaloneCards.value.indexOf(id)
  if (index > -1) {
    selectedStandaloneCards.value.splice(index, 1)
  } else {
    selectedStandaloneCards.value.push(id)
  }
  selectAllStandalone.value = selectedStandaloneCards.value.length === combinedUncategorizedPrompts.value.length
}

// 执行导出独立卡片
const executeExportStandalone = () => {
  const cardsToExport = combinedUncategorizedPrompts.value.filter(p => selectedStandaloneCards.value.includes(p.id))
  const exportData = {
    type: 'standalone-cards',
    version: '1.0',
    exportTime: new Date().toISOString(),
    prompts: cardsToExport.map(p => ({
      name: p.name,
      content: p.content,
      category: p.category,
      order_num: p.order_num,
      created_at: p.created_at,
      fields: p.fields,
      card_type: p.card_type || 'normal',
      password: p.password || null
    }))
  }
  
  const blob = new Blob([encodeExportData(JSON.stringify(exportData))], { type: 'application/octet-stream' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `独立卡片_${new Date().toLocaleDateString('zh-CN').replace(/\//g, '-')}.json`
  a.click()
  URL.revokeObjectURL(url)
  
  ElMessage.success(`成功导出 ${cardsToExport.length} 张独立卡片`)
  exportStandaloneDialogVisible.value = false
}

// 打开导出卡包弹窗
const openExportPackDialog = () => {
  selectedPack.value = null
  exportPackDialogVisible.value = true
}

// 执行导出卡包（统一入口：普通卡包与拆书卡包）
const executeExportPack = () => {
  if (!selectedPack.value) return

  // 优先匹配拆书卡包（拆书导出沿用"去前缀、可移植"格式）
  const bookCategory = bookAnalysisCategoryList.value.find(c => c.name === selectedPack.value)
  if (bookCategory) {
    const exportData = {
      type: 'book-analysis-packs',
      exportedAt: new Date().toISOString(),
      packs: [{
        packName: bookCategory.name.replace(BOOK_ANALYSIS_PREFIX, ''),
        prompts: bookCategory.prompts.map(p => ({
          name: p.name,
          content: p.content,
          category: bookCategory.name,
          order_num: p.order_num || 0,
          created_at: p.created_at,
          fields: p.fields || [],
          card_type: p.card_type || 'normal',
          password: p.password || null
        }))
      }]
    }
    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `拆书卡包导出_${new Date().toISOString().slice(0, 10)}.json`
    link.click()
    URL.revokeObjectURL(url)
    ElMessage.success(`成功导出拆书卡包"${bookCategory.name.replace(BOOK_ANALYSIS_PREFIX, '')}"，包含 ${bookCategory.prompts.length} 张卡片`)
    exportPackDialogVisible.value = false
    return
  }

  const category = categoryList.value.find(c => c.name === selectedPack.value)
  if (!category) return

  const exportData = {
    type: 'card-pack',
    version: '1.0',
    exportTime: new Date().toISOString(),
    packName: category.name,
    prompts: category.prompts.map(p => ({
      name: p.name,
      content: p.content,
      category: p.category,
      order_num: p.order_num,
      created_at: p.created_at,
      fields: p.fields,
      card_type: p.card_type || 'normal',
      password: p.password || null
    }))
  }

  const blob = new Blob([encodeExportData(JSON.stringify(exportData))], { type: 'application/octet-stream' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `卡包_${category.name}_${new Date().toLocaleDateString('zh-CN').replace(/\//g, '-')}.json`
  a.click()
  URL.revokeObjectURL(url)

  ElMessage.success(`成功导出卡包"${category.name}"，包含 ${category.prompts.length} 张卡片`)
  exportPackDialogVisible.value = false
}

// 执行整体导出
const executeExportAll = () => {
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
    password: p.password || null
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
        password: p.password || null
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
      password: p.password || null
    }))
  })
  
  const exportData = {
    type: 'full-backup',
    version: '1.0',
    exportTime: new Date().toISOString(),
    data: {
      standalone,
      packs
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

// 打开导入独立卡片弹窗
const openImportStandaloneDialog = () => {
  importStandalonePreview.value = []
  selectedImportCards.value = []
  selectAllImportStandalone.value = false
  importStandaloneDialogVisible.value = true
}

// 处理独立卡片文件选择
const handleStandaloneFileChange = (file: any) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const content = e.target?.result as string
      const decodedContent = decodeExportData(content)
      const data = JSON.parse(decodedContent)
      
      if (data.type === 'standalone-cards' && Array.isArray(data.prompts)) {
        importStandalonePreview.value = data.prompts
        selectedImportCards.value = data.prompts.map((_: any, index: number) => index)
        selectAllImportStandalone.value = true
        ElMessage.success(`成功读取 ${data.prompts.length} 张卡片`)
      } else if (data.type === 'single-prompt' && data.prompt) {
        importStandalonePreview.value = [data.prompt]
        selectedImportCards.value = [0]
        selectAllImportStandalone.value = true
        ElMessage.success(`成功读取 1 张卡片`)
      } else if (data.type === 'book-analysis-standalone' && Array.isArray(data.prompts)) {
        // 拆书独立卡片格式
        importStandalonePreview.value = data.prompts
        selectedImportCards.value = data.prompts.map((_: any, index: number) => index)
        selectAllImportStandalone.value = true
        ElMessage.success(`成功读取 ${data.prompts.length} 张拆书独立卡片`)
      } else if (Array.isArray(data)) {
        // 兼容：直接数组
        const list = data.filter((p: any) => p && p.name && p.content)
        importStandalonePreview.value = list
        selectedImportCards.value = list.map((_: any, index: number) => index)
        selectAllImportStandalone.value = true
        ElMessage.success(`成功读取 ${list.length} 张卡片`)
      } else if (data.type === 'card-pack' || data.type === 'book-analysis-packs') {
        ElMessage.warning('这是卡包文件，请使用"导入卡包"功能')
      } else {
        ElMessage.error('文件格式不正确')
      }
    } catch (error) {
      ElMessage.error('文件解析失败，请确保是有效的导出文件')
    }
  }
  reader.readAsText(file.raw)
}

// 全选/取消全选导入卡片
const handleSelectAllImportStandalone = (val: boolean) => {
  if (val) {
    selectedImportCards.value = importStandalonePreview.value.map((_, index) => index)
  } else {
    selectedImportCards.value = []
  }
}

// 切换单个导入卡片选择
const toggleImportCard = (index: number) => {
  const idx = selectedImportCards.value.indexOf(index)
  if (idx > -1) {
    selectedImportCards.value.splice(idx, 1)
  } else {
    selectedImportCards.value.push(index)
  }
  selectAllImportStandalone.value = selectedImportCards.value.length === importStandalonePreview.value.length
}

// 执行导入独立卡片
const executeImportStandalone = async () => {
  const cardsToImport = selectedImportCards.value
    .map(index => importStandalonePreview.value[index])
    .filter((card): card is Prompt => card !== undefined)
  
  let successCount = 0
  let failCount = 0
  let skipCount = 0

  // 防重：已存在的判重键（按目标分类规则计算）+ 文件内已导入的判重键
  const existingKeys = new Set<string>()
  prompts.value.forEach(p => {
    const cat = (p.category || '').startsWith(BOOK_ANALYSIS_PREFIX) ? p.category : '未分类'
    const key = buildPromptKey(p.name, cat)
    if (key) existingKeys.add(key)
  })
  const importedKeys = new Set<string>()

  for (const card of cardsToImport) {
    try {
      // 拆书卡片保留其拆书分类（拆书-未分类 等），普通卡片一律归入 未分类
      const category = (card.category || '').startsWith(BOOK_ANALYSIS_PREFIX)
        ? card.category
        : '未分类'
      const key = buildPromptKey(card.name, category)
      if (existingKeys.has(key) || importedKeys.has(key)) {
        skipCount++
        continue
      }
      importedKeys.add(key)
      const res = await promptAPI.create({
        name: card.name,
        content: card.content,
        category,
        order_num: card.order_num || 0,
        created_at: card.created_at,
        fields: card.fields,
        card_type: card.card_type || 'normal',
        password: card.password || null
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
  
  await fetchPrompts()
  
  if (failCount === 0) {
    ElMessage.success(
      skipCount > 0
        ? `成功导入 ${successCount} 张独立卡片，跳过重复 ${skipCount} 张`
        : `成功导入 ${successCount} 张独立卡片`
    )
  } else {
    ElMessage.warning(`导入完成：成功 ${successCount} 张，失败 ${failCount} 张${skipCount > 0 ? `，跳过重复 ${skipCount} 张` : ''}`)
  }
  
  importStandaloneDialogVisible.value = false
}

// 打开导入卡包弹窗
const openImportPackDialog = () => {
  importPackPreview.value = null
  importPackDialogVisible.value = true
}

// 处理卡包文件选择（统一入口：普通卡包 + 拆书卡包）
const handlePackFileChange = (file: any) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const content = e.target?.result as string
      const decodedContent = decodeExportData(content)
      const data = JSON.parse(decodedContent)

      // 普通卡包（编码格式）
      if (data.type === 'card-pack' && Array.isArray(data.prompts)) {
        importPackPreview.value = {
          packName: data.packName,
          prompts: data.prompts,
          isBookAnalysis: (data.packName || '').startsWith(BOOK_ANALYSIS_PREFIX)
        }
        ElMessage.success(`成功读取卡包"${data.packName}"，包含 ${data.prompts.length} 张卡片`)
      } else if (data.type === 'standalone-cards') {
        ElMessage.warning('这是独立卡片文件，请使用"导入独立卡片"功能')
      } else if (data.type === 'book-analysis-packs' && Array.isArray(data.packs) && data.packs.length > 0) {
        // 拆书多卡包格式：取第一个
        const first = data.packs[0]
        importPackPreview.value = {
          packName: first.packName,
          prompts: first.prompts || [],
          isBookAnalysis: true
        }
        ElMessage.success(`成功读取拆书卡包"${first.packName}"，包含 ${first.prompts?.length || 0} 张卡片`)
      } else if (data.packName && Array.isArray(data.prompts)) {
        // 拆书单卡包格式（无 type 标识）
        importPackPreview.value = {
          packName: data.packName,
          prompts: data.prompts,
          isBookAnalysis: true
        }
        ElMessage.success(`成功读取拆书卡包"${data.packName}"，包含 ${data.prompts.length} 张卡片`)
      } else {
        ElMessage.error('文件格式不正确')
      }
    } catch (error) {
      ElMessage.error('文件解析失败，请确保是有效的导出文件')
    }
  }
  reader.readAsText(file.raw)
}

// 执行导入卡包（统一入口：拆书卡包自动加前缀）
const executeImportPack = async () => {
  if (!importPackPreview.value) return

  const { packName, prompts: cardsToImport, isBookAnalysis } = importPackPreview.value

  // 拆书卡包自动补齐前缀，保证落入拆书 Tab
  const fullPackName = isBookAnalysis && !packName.startsWith(BOOK_ANALYSIS_PREFIX)
    ? BOOK_ANALYSIS_PREFIX + packName
    : packName

  // 检查卡包是否已存在，如果不存在则创建
  if (!customCategories.value.includes(fullPackName) && fullPackName !== '默认') {
    customCategories.value.push(fullPackName)
    saveCategoriesToStorage(customCategories.value)
  }

  let successCount = 0
  let failCount = 0
  let skipCount = 0

  // 防重：已存在的判重键（目标包名）+ 文件内已导入的判重键
  const existingKeys = buildExistingKeySet(prompts.value)
  const importedKeys = new Set<string>()

  for (const card of cardsToImport) {
    try {
      const key = buildPromptKey(card.name, fullPackName)
      if (existingKeys.has(key) || importedKeys.has(key)) {
        skipCount++
        continue
      }
      importedKeys.add(key)
      const res = await promptAPI.create({
        name: card.name,
        content: card.content,
        category: fullPackName,
        order_num: card.order_num || 0,
        created_at: card.created_at,
        fields: card.fields,
        card_type: card.card_type || 'normal',
        password: card.password || null
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

  await fetchPrompts()

  if (failCount === 0) {
    ElMessage.success(
      skipCount > 0
        ? `成功导入卡包"${fullPackName}"，包含 ${successCount} 张卡片，跳过重复 ${skipCount} 张`
        : `成功导入卡包"${fullPackName}"，包含 ${successCount} 张卡片`
    )
  } else {
    ElMessage.warning(`导入完成：成功 ${successCount} 张，失败 ${failCount} 张${skipCount > 0 ? `，跳过重复 ${skipCount} 张` : ''}`)
  }

  importPackDialogVisible.value = false
}

// 打开导入并转换txt/md弹窗
const openImportConvertDialog = () => {
  importConvertPreview.value = []
  selectedImportConvertCards.value = []
  selectAllImportConvert.value = false
  mergeAsPack.value = false
  mergePackName.value = ''
  convertUploadRef.value?.clearFiles()
  importConvertDialogVisible.value = true
}

// 读取文件内容
const readFileContent = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      resolve(e.target?.result as string)
    }
    reader.onerror = reject
    reader.readAsText(file)
  })
}

// 处理多个txt/md文件选择
const handleConvertFilesChange = (_uploadFile: any, uploadFiles: any[]) => {
  const files = uploadFiles?.map((item: any) => item.raw).filter((f: File) => f) || []
  
  if (files.length === 0) return
  
  const readPromises = files.map((f: File) => {
    const name = f.name.replace(/\.(txt|md)$/i, '')
    return readFileContent(f).then(content => ({ name, content }))
  })
  
  Promise.all(readPromises).then(items => {
    importConvertPreview.value = items
    selectedImportConvertCards.value = items.map((_, index) => index)
    selectAllImportConvert.value = true
    ElMessage.success(`成功读取 ${items.length} 个文件`)
  }).catch(() => {
    ElMessage.error('文件读取失败')
  })
}

// 全选/取消全选转换导入卡片
const handleSelectAllImportConvert = (val: boolean) => {
  if (val) {
    selectedImportConvertCards.value = importConvertPreview.value.map((_, index) => index)
  } else {
    selectedImportConvertCards.value = []
  }
}

// 切换单个转换导入卡片选择
const toggleImportConvertCard = (index: number) => {
  const idx = selectedImportConvertCards.value.indexOf(index)
  if (idx > -1) {
    selectedImportConvertCards.value.splice(idx, 1)
  } else {
    selectedImportConvertCards.value.push(index)
  }
  selectAllImportConvert.value = selectedImportConvertCards.value.length === importConvertPreview.value.length
}

// 执行导入转换txt/md
const executeImportConvert = async () => {
  const cardsToImport = selectedImportConvertCards.value
    .map(index => importConvertPreview.value[index])
    .filter((card): card is { name: string; content: string } => card !== undefined)
  
  let targetCategory = '未分类'
  if (mergeAsPack.value && mergePackName.value.trim()) {
    targetCategory = mergePackName.value.trim()
    if (!customCategories.value.includes(targetCategory)) {
      customCategories.value.push(targetCategory)
      saveCategoriesToStorage(customCategories.value)
    }
  }
  
  let successCount = 0
  let failCount = 0
  let skipCount = 0

  // 防重：已存在的判重键（目标分类）+ 文件内已导入的判重键
  const existingKeys = buildExistingKeySet(prompts.value)
  const importedKeys = new Set<string>()

  for (const card of cardsToImport) {
    try {
      const key = buildPromptKey(card.name, targetCategory)
      if (existingKeys.has(key) || importedKeys.has(key)) {
        skipCount++
        continue
      }
      importedKeys.add(key)
      const res = await promptAPI.create({
        name: card.name,
        content: card.content,
        category: targetCategory,
        order_num: 0,
        fields: []
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
  
  await fetchPrompts()
  
  if (mergeAsPack.value && mergePackName.value.trim()) {
    if (failCount === 0) {
      ElMessage.success(
        skipCount > 0
          ? `成功导入卡包"${targetCategory}"，包含 ${successCount} 张卡片，跳过重复 ${skipCount} 张`
          : `成功导入卡包"${targetCategory}"，包含 ${successCount} 张卡片`
      )
    } else {
      ElMessage.warning(`导入完成：成功 ${successCount} 张，失败 ${failCount} 张${skipCount > 0 ? `，跳过重复 ${skipCount} 张` : ''}`)
    }
  } else {
    if (failCount === 0) {
      ElMessage.success(
        skipCount > 0
          ? `成功导入 ${successCount} 张独立卡片，跳过重复 ${skipCount} 张`
          : `成功导入 ${successCount} 张独立卡片`
      )
    } else {
      ElMessage.warning(`导入完成：成功 ${successCount} 张，失败 ${failCount} 张${skipCount > 0 ? `，跳过重复 ${skipCount} 张` : ''}`)
    }
  }
  
  mergeAsPack.value = false
  mergePackName.value = ''
  convertUploadRef.value?.clearFiles()
  importConvertDialogVisible.value = false
}

</script>

<style scoped>
.search-bar {
  margin-bottom: 12px;
}
.search-bar .el-input {
  max-width: 360px;
}

.no-search-match {
  padding: 24px 0;
  text-align: center;
}

.preview-dialog-content {
  padding: 20px;
  max-height: 600px;
  display: flex;
  flex-direction: column;
}

.preview-dialog-header {
  margin-bottom: 15px;
  flex-shrink: 0;
}

.preview-scrollbar {
  flex-grow: 1;
  overflow-y: auto;
  max-height: 500px;
}

.pack-fan-stage {
  position: fixed;
  left: 0;
  top: 0;
  right: 0;
  bottom: 0;
  z-index: 2000;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
}

.pack-fan-board {
  max-width: 90vw;
  background: transparent;
}

.pack-fan-row {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 20px;
  padding: 24px;
  overflow-y: auto;
  max-height: 80vh;
  pointer-events: auto;
}

.pack-fan-card {
  position: relative;
  width: 220px;
  height: 240px;
  padding: 16px 18px;
  border: none;
  border-radius: 16px;
  background: #ffffff;
  box-shadow:
    0 4px 16px rgba(15, 23, 42, 0.06),
    0 1px 3px rgba(15, 23, 42, 0.04);
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 10px;
  transition: all 0.28s cubic-bezier(0.4, 0, 0.2, 1);
  text-align: left;
  outline: none;
  font-family: inherit;
}

.pack-fan-card:hover {
  transform: translateY(-6px);
  box-shadow:
    0 12px 32px rgba(15, 23, 42, 0.1),
    0 2px 6px rgba(15, 23, 42, 0.06);
  border-color: #e2e8f0;
}

.pack-fan-card:active {
  transform: translateY(-2px) scale(0.985);
}

/* ---- 卡片头部 ---- */
.pack-fan-card-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-bottom: 10px;
  border-bottom: 1px solid #f1f5f9;
}

.pack-fan-card-head-icon {
  color: #6366f1;
  font-size: 18px;
  flex-shrink: 0;
}

.pack-fan-card-head-name {
  font-size: 15px;
  font-weight: 600;
  color: #1e293b;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pack-fan-card-head-lock {
  color: #e6a23c;
  font-size: 14px;
  flex-shrink: 0;
}

/* ---- 卡片主体 ---- */
.pack-fan-card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow: hidden;
}

.pack-fan-card-tag {
  align-self: flex-start;
}

.pack-fan-card-content {
  font-size: 13px;
  line-height: 1.7;
  color: #475569;
  display: -webkit-box;
  -webkit-line-clamp: 5;
  line-clamp: 5;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-all;
}

.pack-fan-card-content.is-empty {
  color: #94a3b8;
  font-style: italic;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  -webkit-line-clamp: unset;
  line-clamp: unset;
}

.pack-fan-card-content-encrypted {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: #b45309;
  font-style: italic;
}

/* ---- 卡片底部 ---- */
.pack-fan-card-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 8px;
  border-top: 1px solid #f1f5f9;
}

.pack-fan-card-foot-time {
  font-size: 11px;
  color: #94a3b8;
}

.pack-fan-card-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.pack-fan-action-btn {
  width: 26px !important;
  height: 26px !important;
  padding: 0 !important;
  border: 1px solid #e2e8f0 !important;
  background: #fff !important;
  color: #64748b !important;
  transition: all 0.2s ease;
}

.pack-fan-action-btn:hover {
  background: #f1f5f9 !important;
  border-color: #cbd5e1 !important;
  color: #475569 !important;
}

.pack-fan-action-btn--danger:hover {
  background: #fef2f2 !important;
  border-color: #fecaca !important;
  color: #dc2626 !important;
}

.pack-fan-stage-enter-active,
.pack-fan-stage-leave-active {
  transition: opacity 240ms ease;
}

.pack-fan-stage-enter-from,
.pack-fan-stage-leave-to {
  opacity: 0;
}

@media (max-width: 768px) {
  .pack-fan-stage {
    padding: 16px;
  }

  .pack-fan-row {
    gap: 12px;
    padding: 12px;
  }

  .pack-fan-card {
    width: 100%;
    max-width: 320px;
  }
}

.prompts-container {
  background: #f5f7fa;
  padding: 28px 36px 40px;
  min-height: 100%;
  animation: fadeIn 0.4s ease;
}

.subcategory-form {
  width: 100%;
}

.subcategory-tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 8px;
  min-height: 24px;
}

.subcategory-input-row {
  display: flex;
  gap: 8px;
}

.subcategory-input-row .subcategory-input {
  flex-grow: 1;
}

.subcategory-tag {
  border-radius: 10px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  gap: 16px;
}

.header-main {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.header h2 {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: #1f2937;
}

.hidden-input {
  display: none;
}

.header-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: flex-end;
  align-items: center;
}

/* ========== 导出按钮 ========== */
.btn-export {
  --btn-color: #f59e0b;
  border: 1.5px solid #fde68a !important;
  background: #fffbeb !important;
  color: #b45309 !important;
  border-radius: 10px !important;
  font-weight: 500;
  transition: all 0.22s cubic-bezier(0.4, 0, 0.2, 1) !important;
}

.btn-export:hover {
  border-color: #f59e0b !important;
  background: #fffbeb !important;
  color: #92400e !important;
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.15);
}

/* ========== 导入按钮 ========== */
.btn-import {
  border: 1.5px solid #c7d2fe !important;
  background: #eef2ff !important;
  color: #4338ca !important;
  border-radius: 10px !important;
  font-weight: 500;
  transition: all 0.22s cubic-bezier(0.4, 0, 0.2, 1) !important;
}

.btn-import:hover {
  border-color: #818cf8 !important;
  background: #eef2ff !important;
  color: #3730a3 !important;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.15);
}

/* ========== 创建卡包按钮 ========== */
.btn-create-pack {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%) !important;
  border: none !important;
  color: #fff !important;
  border-radius: 10px !important;
  font-weight: 600;
  box-shadow: 0 2px 6px rgba(124, 58, 237, 0.3);
  transition: all 0.22s cubic-bezier(0.4, 0, 0.2, 1) !important;
}

.btn-create-pack:hover {
  background: linear-gradient(135deg, #a78bfa 0%, #8b5cf6 100%) !important;
  box-shadow: 0 4px 14px rgba(124, 58, 237, 0.4);
  transform: translateY(-1px);
}

/* ========== 创建提示词按钮 ========== */
.btn-create-prompt {
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%) !important;
  border: none !important;
  color: #fff !important;
  border-radius: 10px !important;
  font-weight: 600;
  box-shadow: 0 2px 6px rgba(22, 163, 74, 0.3);
  transition: all 0.22s cubic-bezier(0.4, 0, 0.2, 1) !important;
}

.btn-create-prompt:hover {
  background: linear-gradient(135deg, #3b82f6 0%, #22c55e 100%) !important;
  box-shadow: 0 4px 14px rgba(22, 163, 74, 0.4);
  transform: translateY(-1px);
}

/* 区块标题 */
.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  font-size: 18px;
  font-weight: 600;
  color: #262626;
}

.section-title .el-icon {
  font-size: 22px;
  color: #52c41a;
}

.section-count {
  background: #52c41a;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  padding: 2px 10px;
  border-radius: 10px;
  margin-left: 8px;
}

.section-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-action-btn {
  background: rgba(255, 77, 79, 0.1);
  border-color: transparent;
  color: #ff4d4f;
  transition: all 0.2s ease;
}

.section-action-btn:hover {
  background: rgba(255, 77, 79, 0.2);
  color: #ff7875;
}

/* 顶端 Tab 切换 */
.main-tabs-wrapper {
  margin-top: 4px;
}

.main-tabs {
  --el-tabs-header-height: 40px;
}

.main-tabs :deep(.el-tabs__header) {
  margin-bottom: 4px;
}

.main-tabs :deep(.el-tabs__content) {
  overflow: visible;
}

.main-tabs :deep(.el-tab-pane) {
  padding: 0;
}

.main-tabs :deep(.el-tabs__nav-wrap)::after {
  background-color: #e5e7eb;
}

.main-tabs :deep(.el-tabs__item) {
  font-size: 14px;
  font-weight: 500;
  color: #6b7280;
  padding: 0 14px;
  transition: color 0.2s ease;
}

.main-tabs :deep(.el-tabs__item.is-active) {
  color: #22c55e;
  font-weight: 600;
}

.main-tabs :deep(.el-tabs__active-bar) {
  background-color: #22c55e;
  height: 2px;
  border-radius: 1px;
}

.main-tabs :deep(.el-tabs__item:hover) {
  color: #22c55e;
}

.main-tab-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.main-tab-label .el-icon {
  font-size: 15px;
}

.main-tab-count {
  font-size: 12px;
  font-weight: 500;
  color: #9ca3af;
  margin-left: 2px;
  min-width: 22px;
  text-align: center;
}

.main-tabs :deep(.el-tabs__item.is-active) .main-tab-count {
  color: #22c55e;
}

.tab-toolbar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  margin-bottom: 12px;
}

/* 卡包横向排列容器 */
.packs-wrapper {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px;
  background: #f4f6f8;
  border-radius: 12px;
}

.packs-wrapper.grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px;
}

.packs-wrapper.list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* ========== 视图切换按钮 ========== */
.pack-view-toggle {
  display: inline-flex;
  margin-left: auto;
  background: #f1f5f9;
  border-radius: 8px;
  padding: 2px;
  gap: 1px;
  border: 1px solid #e2e8f0;
}

.pack-view-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 28px;
  border: none;
  background: transparent;
  border-radius: 6px;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 15px;
  outline: none;
}

.pack-view-btn:hover {
  color: #64748b;
  background: rgba(255, 255, 255, 0.5);
}

.pack-view-btn.active {
  background: #ffffff;
  color: #6366f1;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

/* ========== 卡包信息卡片 ========== */
.pack-info-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid #eef0f3;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
  transition: all 0.22s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: default;
  position: relative;
}

.pack-info-clickable {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
  cursor: pointer;
  border-radius: 8px;
  transition: background 0.2s ease;
  padding: 2px 4px;
  margin: -2px -4px;
}

.pack-info-clickable:hover {
  background: rgba(59, 130, 246, 0.05);
}

.pack-info-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 12px;
  pointer-events: none;
  transition: opacity 0.22s ease, box-shadow 0.22s ease;
  opacity: 0;
  box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.4);
}

.pack-info-card:hover {
  border-color: #e2e6ec;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.06);
  transform: translateY(-1px);
}

.pack-info-card:hover::before {
  opacity: 1;
}

.pack-info-card.drag-over {
  border-color: #93c5fd;
  background: #f0f7ff;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.12);
}

.pack-info-card.drag-over::before {
  opacity: 1;
  box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.55);
}

/* 网格模式 - 横向简洁布局 */
.packs-wrapper.grid .pack-info-card {
  flex-direction: row;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  text-align: left;
}

.packs-wrapper.grid .pack-info-icon {
  width: 38px;
  height: 38px;
  border-radius: 9px;
}

.packs-wrapper.grid .pack-info-icon .el-icon {
  font-size: 18px;
}

.packs-wrapper.grid .pack-info-body {
  align-items: flex-start;
  flex: 1;
  min-width: 0;
}

.packs-wrapper.grid .pack-info-name {
  font-size: 14px;
  max-width: 100%;
  font-weight: 600;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  width: 100%;
}

.packs-wrapper.grid .pack-info-count {
  font-size: 11px;
  color: #94a3b8;
  background: transparent;
  padding: 0;
  border-radius: 0;
  font-weight: 400;
}

.packs-wrapper.grid .pack-info-spacer {
  flex: 1;
}

.packs-wrapper.grid .pack-info-actions {
  display: flex;
  position: absolute;
  top: 8px;
  right: 8px;
}

.packs-wrapper.list .pack-info-card {
  flex-direction: row;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
}

.packs-wrapper.list .pack-info-spacer {
  display: block;
  flex: 1;
}

.packs-wrapper.list .pack-info-actions {
  display: flex;
}

/* 图标区 */
.pack-info-icon {
  width: 38px;
  height: 38px;
  border-radius: 9px;
  background: linear-gradient(135deg, #dbeafe 0%, #e0e7ff 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.pack-info-icon .el-icon {
  font-size: 18px;
  color: #3b82f6;
}

/* 信息主体 */
.pack-info-body {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.pack-info-name {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pack-info-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pack-info-count {
  font-size: 11px;
  color: #94a3b8;
  font-weight: 400;
}

.pack-info-spacer {
  flex: 1;
}

/* 操作按钮 */
.pack-info-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.pack-action-btn {
  width: 30px !important;
  height: 30px !important;
  padding: 0 !important;
  border: 1px solid transparent !important;
  background: transparent !important;
  color: #94a3b8 !important;
  transition: all 0.2s ease !important;
}

.pack-action-btn:hover {
  background: #f1f5f9 !important;
  border-color: #e2e8f0 !important;
  color: #475569 !important;
}

.pack-action-btn--danger:hover {
  background: #fef2f2 !important;
  border-color: #fecaca !important;
  color: #dc2626 !important;
}

/* 独立卡片区域 */
.standalone-cards-area {
  background: linear-gradient(180deg, #fff7e6 0%, #ffffff 100%);
  border-radius: 12px;
  border: 2px dashed #ffd591;
  padding: 20px;
  transition: all 0.3s ease;
}

.standalone-cards-area.drag-over {
  border-color: #fa8c16;
  border-style: solid;
  background: linear-gradient(180deg, #ffe7ba 0%, #ffffff 100%);
}

.standalone-cards-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

/* 独立卡片样式 - 基础 */
.standalone-card {
  background: #fff;
  border-radius: 12px;
  border: 2px solid #e8e8e8;
  cursor: grab;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.standalone-card:hover {
  border-color: #fa8c16;
  box-shadow: 0 8px 24px rgba(250, 140, 22, 0.2);
}

.standalone-card:active {
  cursor: grabbing;
}

.standalone-card.dragging {
  opacity: 0.5;
  transform: scale(0.95);
}

/* 独立卡片尺寸 - 小 */
.standalone-card.small {
  width: 160px;
  height: 200px;
}

.standalone-card.small .standalone-card-header {
  padding: 10px;
}

.standalone-card.small .standalone-card-name {
  font-size: 12px;
}

.standalone-card.small .standalone-card-content {
  padding: 0 10px;
}

.standalone-card.small .standalone-card-preview {
  font-size: 10px;
  -webkit-line-clamp: 2;
  line-clamp: 2;
}

.standalone-card.small .standalone-card-footer {
  padding: 8px 10px;
}

/* 独立卡片尺寸 - 中 */
.standalone-card.medium {
  width: 200px;
  height: 260px;
}

.standalone-card.medium .standalone-card-header {
  padding: 12px;
}

.standalone-card.medium .standalone-card-name {
  font-size: 14px;
}

.standalone-card.medium .standalone-card-content {
  padding: 0 12px;
}

.standalone-card.medium .standalone-card-preview {
  font-size: 11px;
  -webkit-line-clamp: 3;
  line-clamp: 3;
}

.standalone-card.medium .standalone-card-footer {
  padding: 10px 12px;
}

/* 独立卡片尺寸 - 大 */
.standalone-card.large {
  width: 260px;
  height: 340px;
}

.standalone-card.large .standalone-card-header {
  padding: 16px;
}

.standalone-card.large .standalone-card-name {
  font-size: 16px;
}

.standalone-card.large .standalone-card-content {
  padding: 0 16px;
}

.standalone-card.large .standalone-card-preview {
  font-size: 12px;
  -webkit-line-clamp: 4;
  line-clamp: 4;
}

.standalone-card.large .standalone-card-footer {
  padding: 12px 16px;
}

/* 独立卡片头部 */
.standalone-card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  background: linear-gradient(135deg, #fff7e6 0%, #ffffff 100%);
  border-bottom: 1px solid #f0f0f0;
}

.card-drag-handle {
  color: #fa8c16;
  cursor: grab;
}

.standalone-card-name {
  font-weight: 600;
  color: #262626;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 独立卡片内容区 */
.standalone-card-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow: hidden;
}

.standalone-card-tag {
  align-self: flex-start;
}

.standalone-card-preview {
  color: #8c8c8c;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* 独立卡片底部 */
.standalone-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid #f0f0f0;
  background: #fafafa;
}

.standalone-card-time {
  font-size: 11px;
  color: #bfbfbf;
}

.standalone-card-actions {
  display: flex;
  gap: 4px;
}

.standalone-card-actions :deep(.el-button) {
  min-width: 28px;
  height: 28px;
  padding: 0 8px;
  border-radius: 999px;
  border: 1px solid transparent;
  background: rgba(15, 118, 110, 0.06);
  transition: all 0.2s ease;
}

.standalone-card-actions :deep(.el-button.el-button--primary) {
  color: #0f766e;
  border-color: rgba(15, 118, 110, 0.10);
}

.standalone-card-actions :deep(.el-button.el-button--primary:hover) {
  color: #fff;
  background: linear-gradient(135deg, #0f766e 0%, #2563eb 100%);
  border-color: transparent;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.22);
}

.standalone-card-actions :deep(.el-button.el-button--danger) {
  color: #c2415b;
  background: rgba(225, 29, 72, 0.06);
  border-color: rgba(225, 29, 72, 0.10);
}

.standalone-card-actions :deep(.el-button.el-button--danger:hover) {
  color: #fff;
  background: linear-gradient(135deg, #e11d48 0%, #fb7185 100%);
  border-color: transparent;
  box-shadow: 0 8px 18px rgba(225, 29, 72, 0.18);
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

.form-item-label-with-guide {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.prompt-guide-trigger {
  padding: 0;
  height: auto;
  font-size: 12px;
  color: #409eff;
}

.prompt-guide-trigger .el-icon {
  margin-right: 4px;
}

.prompt-guide-popover {
  font-size: 13px;
  line-height: 1.7;
  color: #303133;
}

.prompt-guide-title {
  margin-bottom: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
}

.prompt-guide-intro {
  margin: 0;
  color: #4b5563;
}

.prompt-guide-example {
  margin-top: 12px;
  padding: 12px 14px;
  border-radius: 8px;
  background: #f6faff;
  border: 1px solid #d9ecff;
}

.prompt-guide-example-label {
  display: inline-block;
  margin-bottom: 6px;
  font-size: 12px;
  font-weight: 600;
  color: #409eff;
}

.prompt-guide-example p {
  margin: 0;
}

.prompt-guide-tips {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 12px;
}

.prompt-guide-tip {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  color: #4b5563;
}

.prompt-guide-tip-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 999px;
  background: #ecf5ff;
  color: #409eff;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}

.prompt-guide-popover code,
.field-config-banner code {
  background: linear-gradient(135deg, rgba(74, 126, 123, 0.12), rgba(98, 176, 170, 0.08));
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  color: #276b67;
  font-size: 12px;
}

/* 字段配置样式 - 玻璃拟态主题 */
.field-config-section {
  margin-top: 12px;
}

.field-config-banner {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  background: linear-gradient(135deg, #edf5f4 0%, #e5efee 50%, #dbe8e7 100%);
  border: 1px solid rgba(74, 126, 123, 0.18);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(25, 70, 68, 0.06), inset 0 1px 0 rgba(255, 255, 255, 0.5);
  position: relative;
  overflow: hidden;
}

.field-config-banner::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.6), transparent);
}

.field-config-banner-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #98c2bc 0%, #afd2cd 100%);
  border-radius: 10px;
  color: #154d4b;
  font-size: 18px;
  flex-shrink: 0;
  box-shadow: 0 2px 6px rgba(31, 89, 86, 0.15), inset 0 1px 0 rgba(255, 255, 255, 0.3);
}

.field-config-banner-body {
  flex: 1;
  min-width: 0;
}

.field-config-banner-title {
  font-weight: 600;
  font-size: 14px;
  color: #154d4b;
  margin-bottom: 3px;
}

.field-config-banner-desc {
  font-size: 13px;
  color: #607d79;
  line-height: 1.5;
}

.empty-fields {
  text-align: center;
  padding: 28px 24px;
  background: linear-gradient(135deg, #f7fbfa 0%, #edf5f4 100%);
  border: 1px dashed rgba(74, 126, 123, 0.22);
  border-radius: 12px;
  margin-top: 12px;
}

.empty-hint {
  margin-top: 12px;
  color: #607d79;
  font-size: 14px;
}

.fields-list {
  margin-top: 12px;
}

.field-item {
  background: linear-gradient(135deg, #f7fbfa 0%, #edf5f4 100%);
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
  border: 1px solid rgba(74, 126, 123, 0.15);
  box-shadow: 0 2px 8px rgba(25, 70, 68, 0.05), inset 0 1px 0 rgba(255, 255, 255, 0.5);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.field-item:hover {
  border-color: rgba(74, 126, 123, 0.3);
  box-shadow: 0 4px 14px rgba(25, 70, 68, 0.08), inset 0 1px 0 rgba(255, 255, 255, 0.5);
}

.field-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.field-name-container {
  display: flex;
  align-items: center;
  gap: 8px;
}

.required-tag {
  font-weight: 600;
  min-width: 40px;
  text-align: center;
}

.required-tag.el-tag--danger {
  background: linear-gradient(135deg, #fff1f0 0%, #ffccc7 100%);
  border-color: #ffa39e;
  color: #cf1322;
}

.required-tag.el-tag--info {
  background: linear-gradient(135deg, #e7f1f0 0%, #d0e6e2 100%);
  border-color: rgba(74, 126, 123, 0.25);
  color: #276b67;
}

.field-name-input {
  width: 200px;
  font-weight: 600;
  font-family: 'Courier New', monospace;
}

.field-name {
  font-weight: 600;
  color: #154d4b;
  font-family: 'Courier New', monospace;
  background: linear-gradient(135deg, #e7f1f0 0%, #ddeceb 100%);
  padding: 4px 8px;
  border-radius: 6px;
  border: 1px solid rgba(74, 126, 123, 0.18);
}

.field-actions {
  display: flex;
  gap: 8px;
}

.field-config-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.field-input {
  width: 100%;
}

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

.field-options {
  margin-top: 8px;
}

.field-options label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #303133;
}

.form-item {
  margin-bottom: 20px;
}

.form-label {
  display: block;
  font-weight: 600;
  color: #303133;
  font-size: 14px;
  margin-bottom: 8px;
}

/* 对话框左右布局样式 */
:deep(.prompt-dialog) {
  .el-dialog__header {
    padding: 16px 20px;
  }
  
  .el-dialog__body {
    padding: 20px;
  }
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.dialog-header span {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.size-controls {
  display: flex;
  align-items: center;
}

.dialog-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 600px;
  overflow-y: auto;
  padding-right: 4px;
}

.left-panel {
  flex: none;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: visible;
  padding-bottom: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.right-panel {
  flex: none;
  min-width: 0;
  display: flex;
  flex-direction: column;
  padding-top: 4px;
  overflow: visible;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 2px solid #f0f2f5;
}

.panel-header .form-label {
  margin-bottom: 0;
}

.field-config-section {
  flex: none;
  overflow: visible;
}

/* 导出弹窗样式 */
.export-dialog .export-standalone-content,
.export-dialog .export-pack-content {
  min-height: 300px;
}

.export-hint {
  margin-bottom: 16px;
}

.export-select-all {
  margin-bottom: 16px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 8px;
}

.export-cards-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
  padding: 4px;
}

.export-card-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px;
  background: #fff;
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.export-card-item:hover {
  border-color: #fa8c16;
  box-shadow: 0 2px 8px rgba(250, 140, 22, 0.15);
}

.export-card-item.selected {
  border-color: #52c41a;
  background: #f6ffed;
}

.export-card-content {
  flex: 1;
  min-width: 0;
}

.export-card-name {
  font-weight: 600;
  font-size: 13px;
  color: #262626;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.export-card-preview {
  font-size: 11px;
  color: #8c8c8c;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.export-empty {
  padding: 40px;
}

.export-pack-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 400px;
  overflow-y: auto;
}

.export-pack-item {
  padding: 12px 16px;
  background: #fff;
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.export-pack-item:hover {
  border-color: #52c41a;
  box-shadow: 0 2px 8px rgba(82, 196, 26, 0.15);
}

.export-pack-item.selected {
  border-color: #52c41a;
  background: #f6ffed;
}

.pack-item-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pack-item-info .el-icon {
  color: #52c41a;
  font-size: 18px;
}

.pack-item-name {
  font-weight: 600;
  color: #262626;
}

/* 批量删除弹窗样式 */
.batch-delete-content {
  max-height: 500px;
  overflow-y: auto;
}

.batch-delete-hint {
  margin-bottom: 16px;
}

.batch-delete-select-all {
  margin-bottom: 16px;
  padding: 8px 0;
  border-bottom: 1px solid #e8e8e8;
}

.batch-delete-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.batch-delete-card-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px;
  background: #fff;
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.batch-delete-card-item:hover {
  border-color: #ff4d4f;
  box-shadow: 0 2px 8px rgba(255, 77, 79, 0.15);
}

.batch-delete-card-item.selected {
  border-color: #ff4d4f;
  background: #fff1f0;
}

.batch-delete-card-content {
  flex: 1;
  min-width: 0;
}

.batch-delete-card-name {
  font-weight: 600;
  font-size: 13px;
  color: #262626;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.batch-delete-card-preview {
  font-size: 11px;
  color: #8c8c8c;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.batch-delete-empty {
  padding: 40px;
}

/* 移动卡片弹窗样式 */
.move-prompt-content {
  min-height: 200px;
}

.move-prompt-hint {
  margin-bottom: 20px;
}

.move-prompt-options {
  max-height: 300px;
  overflow-y: auto;
}

.move-prompt-radio-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.move-prompt-radio-item {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.move-prompt-radio-item:hover {
  border-color: #6366f1;
  background: #f8fafc;
}

.move-prompt-category-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.move-prompt-category-info .el-icon {
  color: #6366f1;
}

/* 导入弹窗样式 */
.import-dialog .import-content {
  min-height: 300px;
}

/* 整体备份导入弹窗样式 */
.import-all-content {
  min-height: 300px;
}

.import-all-upload {
  margin-bottom: 20px;
}

.import-all-preview {
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.import-all-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.import-all-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #fff;
  border-radius: 6px;
  font-size: 14px;
  color: #606266;
}

.import-all-item .el-icon {
  color: #409eff;
  font-size: 18px;
}

.import-all-item strong {
  color: #409eff;
  font-weight: 600;
}

.import-upload {
  margin-bottom: 20px;
}

.import-preview {
  margin-top: 20px;
}

.import-preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e8e8e8;
  font-weight: 600;
  color: #262626;
}

.import-cards-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  max-height: 300px;
  overflow-y: auto;
  padding: 4px;
}

.import-card-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px;
  background: #fff;
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.import-card-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.15);
}

.import-card-item.selected {
  border-color: #409eff;
  background: #ecf5ff;
}

.import-card-content {
  flex: 1;
  min-width: 0;
}

.import-card-name {
  font-weight: 600;
  font-size: 13px;
  color: #262626;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.import-card-preview {
  font-size: 11px;
  color: #8c8c8c;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.import-pack-preview {
  margin-top: 20px;
  padding: 16px;
  background: #f6ffed;
  border-radius: 8px;
  border: 2px solid #b7eb8f;
}

.import-pack-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #b7eb8f;
}

.import-pack-info .el-icon {
  color: #52c41a;
  font-size: 24px;
}

.import-pack-name {
  font-weight: 600;
  font-size: 16px;
  color: #262626;
}

.import-pack-prompts {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 200px;
  overflow-y: auto;
}

.import-pack-prompt-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #fff;
  border-radius: 6px;
  font-size: 13px;
  color: #595959;
}

.import-pack-prompt-item .el-icon {
  color: #52c41a;
}

.standalone-card.encrypted {
  border-color: rgba(230, 162, 60, 0.4) !important;
  box-shadow: 0 4px 12px rgba(230, 162, 60, 0.12) !important;
}

.standalone-card.encrypted:hover {
  border-color: rgba(230, 162, 60, 0.6) !important;
  box-shadow: 0 8px 20px rgba(230, 162, 60, 0.18) !important;
}

.standalone-card.encrypted .standalone-card-header {
  background: linear-gradient(135deg, rgba(230, 162, 60, 0.15) 0%, rgba(245, 249, 248, 0.94) 100%) !important;
}

.standalone-card-lock {
  color: #e6a23c;
  font-size: 14px;
  margin-left: auto;
  flex-shrink: 0;
}

.card-type-selector {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.card-type-selector .el-radio-group {
  display: flex;
  gap: 16px;
}

.card-type-selector .el-radio {
  display: flex;
  align-items: center;
  gap: 4px;
}

.card-type-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-left: 4px;
}

.card-type-status {
  padding-left: 4px;
}

.card-type-status .el-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
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

/* 暗色主题适配 - 卡包信息卡片 */
:root[data-theme='dark'] .packs-wrapper {
  background: rgba(13, 17, 23, 0.45);
}

:root[data-theme='dark'] .pack-info-clickable:hover {
  background: rgba(96, 165, 250, 0.08);
}

:root[data-theme='dark'] .pack-info-card {
  background: rgba(22, 27, 34, 0.85);
  border-color: rgba(42, 50, 66, 0.35);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

:root[data-theme='dark'] .pack-info-card:hover {
  border-color: rgba(148, 163, 184, 0.4);
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .pack-info-card.drag-over {
  border-color: #60a5fa;
  background: rgba(59, 130, 246, 0.1);
  box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.15);
}

:root[data-theme='dark'] .pack-info-icon {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.2) 0%, rgba(99, 102, 241, 0.2) 100%);
}

:root[data-theme='dark'] .pack-info-icon .el-icon {
  color: #93c5fd;
}

:root[data-theme='dark'] .pack-info-name {
  color: #eef5f0;
}

:root[data-theme='dark'] .pack-info-count {
  color: #5f6b7a;
  background: transparent;
}

:root[data-theme='dark'] .pack-action-btn {
  background: transparent !important;
  color: #5f6b7a !important;
}

:root[data-theme='dark'] .pack-action-btn:hover {
  background: rgba(42, 50, 66, 0.55) !important;
  border-color: rgba(148, 163, 184, 0.35) !important;
  color: #e6edf3 !important;
}

:root[data-theme='dark'] .pack-action-btn--danger:hover {
  background: rgba(220, 38, 38, 0.15) !important;
  border-color: rgba(248, 113, 113, 0.3) !important;
  color: #fca5a5 !important;
}

/* --- 暗色主题 - 视图切换按钮 --- */
:root[data-theme='dark'] .pack-view-toggle {
  background: rgba(28, 35, 51, 0.6);
  border-color: rgba(42, 50, 66, 0.5);
}

:root[data-theme='dark'] .pack-view-btn {
  color: #4d5768;
}

:root[data-theme='dark'] .pack-view-btn:hover {
  color: #5f6b7a;
  background: rgba(255, 255, 255, 0.06);
}

:root[data-theme='dark'] .pack-view-btn.active {
  background: rgba(99, 102, 241, 0.2);
  color: #a5b4fc;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

/* --- 暗色主题 - 网格模式 --- */
:root[data-theme='dark'] .packs-wrapper.grid .pack-info-actions {
  border-top-color: rgba(42, 50, 66, 0.3);
}

:root[data-theme='dark'] .standalone-cards-area {
  background: linear-gradient(180deg, rgba(124, 45, 18, 0.2) 0%, #161b22 100%);
  border-color: rgba(249, 115, 22, 0.4);
}

:root[data-theme='dark'] .standalone-cards-area.drag-over {
  border-color: rgba(249, 115, 22, 0.6);
  background: linear-gradient(180deg, rgba(124, 45, 18, 0.3) 0%, #161b22 100%);
}

:root[data-theme='dark'] .standalone-card {
  background: rgba(22, 27, 34, 0.95);
  border-color: rgba(42, 50, 66, 0.4);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

:root[data-theme='dark'] .standalone-card:hover {
  border-color: rgba(249, 115, 22, 0.5);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .standalone-card-header {
  background: linear-gradient(135deg, rgba(124, 45, 18, 0.2) 0%, rgba(22, 27, 34, 0.9) 100%);
  border-bottom-color: rgba(42, 50, 66, 0.3);
}

:root[data-theme='dark'] .card-drag-handle {
  color: #fbbf24;
}

:root[data-theme='dark'] .standalone-card-name {
  color: #e6edf3;
}

:root[data-theme='dark'] .standalone-card-preview {
  color: #9ca3af;
}

:root[data-theme='dark'] .standalone-card-footer {
  border-top-color: rgba(42, 50, 66, 0.3);
  background: rgba(22, 27, 34, 0.8);
}

:root[data-theme='dark'] .standalone-card-time {
  color: #4d5768;
}

:root[data-theme='dark'] .standalone-card-actions :deep(.el-button) {
  background: rgba(59, 130, 246, 0.1);
}

:root[data-theme='dark'] .standalone-card-actions :deep(.el-button.el-button--primary) {
  color: #93c5fd;
  border-color: rgba(59, 130, 246, 0.2);
}

:root[data-theme='dark'] .standalone-card-actions :deep(.el-button.el-button--primary:hover) {
  color: #fff;
  background: linear-gradient(135deg, #2563eb 0%, #22c55e 100%);
}

:root[data-theme='dark'] .standalone-card-actions :deep(.el-button.el-button--danger) {
  color: #f87171;
  background: rgba(248, 113, 113, 0.1);
  border-color: rgba(248, 113, 113, 0.2);
}

:root[data-theme='dark'] .standalone-card-actions :deep(.el-button.el-button--danger:hover) {
  color: #fff;
  background: linear-gradient(135deg, #e11d48 0%, #fb7185 100%);
}

:root[data-theme='dark'] .prompts-container {
  background: #0f172a;
}

:root[data-theme='dark'] .header {
  border-bottom-color: rgba(42, 50, 66, 0.35);
}

:root[data-theme='dark'] .header h2 {
  color: #eef5f0;
}

/* --- 暗色主题 - 导出/导入按钮 --- */
:root[data-theme='dark'] .btn-export {
  border-color: rgba(251, 191, 36, 0.3) !important;
  background: rgba(251, 191, 36, 0.08) !important;
  color: #fbbf24 !important;
}

:root[data-theme='dark'] .btn-export:hover {
  border-color: rgba(251, 191, 36, 0.5) !important;
  background: rgba(251, 191, 36, 0.12) !important;
  color: #fcd34d !important;
  box-shadow: 0 2px 12px rgba(251, 191, 36, 0.12);
}

:root[data-theme='dark'] .btn-import {
  border-color: rgba(129, 140, 248, 0.3) !important;
  background: rgba(129, 140, 248, 0.08) !important;
  color: #a5b4fc !important;
}

:root[data-theme='dark'] .btn-import:hover {
  border-color: rgba(129, 140, 248, 0.5) !important;
  background: rgba(129, 140, 248, 0.12) !important;
  color: #c7d2fe !important;
  box-shadow: 0 2px 12px rgba(129, 140, 248, 0.12);
}

/* --- 暗色主题 - 创建按钮 --- */
:root[data-theme='dark'] .btn-create-pack {
  background: linear-gradient(135deg, #7c3aed 0%, #6d28d9 100%) !important;
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.35);
}

:root[data-theme='dark'] .btn-create-pack:hover {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%) !important;
  box-shadow: 0 4px 16px rgba(124, 58, 237, 0.5);
}

:root[data-theme='dark'] .btn-create-prompt {
  background: linear-gradient(135deg, #16a34a 0%, #2563eb 100%) !important;
  box-shadow: 0 2px 8px rgba(22, 163, 74, 0.35);
}

:root[data-theme='dark'] .btn-create-prompt:hover {
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%) !important;
  box-shadow: 0 4px 16px rgba(22, 163, 74, 0.5);
}

:root[data-theme='dark'] .section-title {
  color: #e6edf3;
}

:root[data-theme='dark'] .section-title .el-icon {
  color: #34d399;
}

:root[data-theme='dark'] .section-count {
  background: #10b981;
}

:root[data-theme='dark'] .section-action-btn {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

:root[data-theme='dark'] .main-tabs :deep(.el-tabs__nav-wrap)::after {
  background-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .main-tabs :deep(.el-tabs__item) {
  color: #8b98a8;
}

:root[data-theme='dark'] .main-tabs :deep(.el-tabs__item:hover) {
  color: #34d399;
}

:root[data-theme='dark'] .main-tabs :deep(.el-tabs__item.is-active) {
  color: #34d399;
}

:root[data-theme='dark'] .main-tabs :deep(.el-tabs__active-bar) {
  background-color: #10b981;
}

:root[data-theme='dark'] .main-tab-count {
  background: rgba(16, 185, 129, 0.18);
  color: #34d399;
}

:root[data-theme='dark'] .main-tabs :deep(.el-tabs__item.is-active) .main-tab-count {
  background: rgba(16, 185, 129, 0.28);
  color: #6ee7b7;
}

:root[data-theme='dark'] .section-action-btn:hover {
  background: rgba(239, 68, 68, 0.25);
  color: #fca5a5;
}

:root[data-theme='dark'] .dialog-header span {
  color: #e6edf3;
}

:root[data-theme='dark'] .left-panel {
  border-bottom-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .panel-header {
  border-bottom-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .export-select-all {
  background: rgba(28, 35, 51, 0.6);
}

:root[data-theme='dark'] .export-card-item {
  background: rgba(22, 27, 34, 0.8);
  border-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .export-card-item:hover {
  border-color: rgba(249, 115, 22, 0.5);
}

:root[data-theme='dark'] .export-card-item.selected {
  border-color: #10b981;
  background: rgba(16, 185, 129, 0.15);
}

:root[data-theme='dark'] .export-card-name {
  color: #e6edf3;
}

:root[data-theme='dark'] .export-card-preview {
  color: #9ca3af;
}

:root[data-theme='dark'] .form-label {
  color: #e6edf3;
}

:root[data-theme='dark'] .empty-fields {
  background-color: rgba(28, 35, 51, 0.6);
}

:root[data-theme='dark'] .empty-hint {
  color: #9ca3af;
}

:root[data-theme='dark'] .field-item {
  background-color: rgba(28, 35, 51, 0.6);
}

:root[data-theme='dark'] .required-tag.el-tag--danger {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.2) 0%, rgba(239, 68, 68, 0.1) 100%);
  border-color: rgba(239, 68, 68, 0.4);
  color: #f87171;
}

:root[data-theme='dark'] .required-tag.el-tag--info {
  background: linear-gradient(135deg, rgba(107, 114, 128, 0.2) 0%, rgba(107, 114, 128, 0.1) 100%);
  border-color: rgba(107, 114, 128, 0.4);
  color: #9ca3af;
}

:root[data-theme='dark'] .field-name {
  color: #e6edf3;
  background-color: rgba(124, 45, 18, 0.3);
  border-color: rgba(251, 191, 36, 0.4);
}

:root[data-theme='dark'] .password-dialog-icon {
  background: rgba(251, 191, 36, 0.15);
}

:root[data-theme='dark'] .password-dialog-hint {
  color: #9ca3af;
}

:root[data-theme='dark'] .standalone-card.encrypted {
  border-color: rgba(251, 191, 36, 0.4) !important;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2) !important;
}

:root[data-theme='dark'] .standalone-card.encrypted:hover {
  border-color: rgba(251, 191, 36, 0.6) !important;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.3) !important;
}

:root[data-theme='dark'] .standalone-card.encrypted .standalone-card-header {
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.2) 0%, rgba(22, 27, 34, 0.9) 100%) !important;
}

:root[data-theme='dark'] .standalone-card-lock {
  color: #fbbf24;
}

/* --- 暗色主题 - 牌桌预览卡片 --- */
:root[data-theme='dark'] .pack-fan-stage {
  background: rgba(2, 6, 23, 0.7);
}

:root[data-theme='dark'] .pack-fan-card {
  background: rgba(22, 27, 34, 0.95);
  box-shadow:
    0 4px 16px rgba(0, 0, 0, 0.3),
    0 1px 3px rgba(0, 0, 0, 0.2);
}

:root[data-theme='dark'] .pack-fan-card:hover {
  box-shadow:
    0 12px 32px rgba(0, 0, 0, 0.4),
    0 2px 6px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .pack-fan-card-head {
  border-bottom-color: rgba(42, 50, 66, 0.3);
}

:root[data-theme='dark'] .pack-fan-card-head-icon {
  color: #a5b4fc;
}

:root[data-theme='dark'] .pack-fan-card-head-name {
  color: #eef5f0;
}

:root[data-theme='dark'] .pack-fan-card-head-lock {
  color: #fbbf24;
}

:root[data-theme='dark'] .pack-fan-card-content {
  color: #5f6b7a;
}

:root[data-theme='dark'] .pack-fan-card-content.is-empty {
  color: #4d5768;
}

:root[data-theme='dark'] .pack-fan-card-content-encrypted {
  color: #fbbf24;
}

:root[data-theme='dark'] .pack-fan-card-foot {
  border-top-color: rgba(42, 50, 66, 0.3);
}

:root[data-theme='dark'] .pack-fan-card-foot-time {
  color: #4d5768;
}

:root[data-theme='dark'] .pack-fan-action-btn {
  border-color: rgba(42, 50, 66, 0.4) !important;
  background: rgba(22, 27, 34, 0.8) !important;
  color: #5f6b7a !important;
}

:root[data-theme='dark'] .pack-fan-action-btn:hover {
  background: rgba(28, 35, 51, 0.8) !important;
  border-color: rgba(100, 116, 139, 0.5) !important;
  color: #e6edf3 !important;
}

:root[data-theme='dark'] .pack-fan-action-btn--danger:hover {
  background: rgba(127, 29, 29, 0.3) !important;
  border-color: rgba(239, 68, 68, 0.4) !important;
  color: #f87171 !important;
}

:root[data-theme='dark'] .move-prompt-radio-item {
  border-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .move-prompt-radio-item:hover {
  border-color: #818cf8;
  background: rgba(22, 27, 34, 0.5);
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
  border-color: #409eff;
  background: #ecf5ff;
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

/* 模板引用行（字段卡片内） */
.pack-edit-field-reference {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  padding: 5px 8px;
  border: 1px dashed #e5e7eb;
  border-radius: 6px;
  background: transparent;
  cursor: pointer;
  transition: all 0.18s ease;
}

.pack-edit-field-reference:hover {
  border-color: #409eff;
  background: #ecf5ff;
}

.pack-edit-field-reference span {
  font-size: 11px;
  color: #6b7280;
  flex-shrink: 0;
}

.pack-edit-field-reference code {
  font-size: 11px;
  color: #4f46e5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  border-color: #409eff;
  background: rgba(64, 158, 255, 0.12);
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

:root[data-theme='dark'] .pack-edit-field-reference {
  border-color: #2e3650;
}

:root[data-theme='dark'] .pack-edit-field-reference:hover {
  border-color: #409eff;
  background: rgba(64, 158, 255, 0.12);
}

:root[data-theme='dark'] .pack-edit-field-reference span {
  color: #9ca3af;
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

/* 内容区警告横幅（对齐灵犀：橙色三角警告） */
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

/* 未使用字段提示（对齐灵犀） */
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
  border-color: #409eff;
}
/* 卡片头（灵犀式：编号 + 名称 + 类型 + 箭头） */
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

/* ========== 顶部操作栏（对标灵犀） ========== */
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

/* ========== 使用模式 / 内容公开（对标灵犀） ========== */
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

/* ========== 预览测试弹窗（对标灵犀） ========== */
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
  color: #409eff;
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
  background: #409eff;
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
  border-color: #409eff;
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

/* ========== 拆书提示词Tab ========== */
.tab-toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.section-divider {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 20px 0 12px;
  padding: 8px 16px;
  font-size: 14px;
  font-weight: 500;
  color: #606266;
  border-bottom: 1px solid #e4e7ed;
}

.section-divider .section-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  font-size: 12px;
  color: #fff;
  background: #e6a23c;
  border-radius: 10px;
}

.import-preview-pack-name {
  margin: 12px 0;
  padding: 10px 12px;
  background: rgba(230, 162, 60, 0.1);
  border-radius: 6px;
  font-size: 14px;
}

.import-preview-prefix-hint {
  color: #e6a23c;
  font-size: 12px;
  margin-left: 8px;
}

.import-preview-cards-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 12px;
}

.import-preview-card-item {
  padding: 8px 12px;
  background: rgba(0, 0, 0, 0.02);
  border-radius: 6px;
}

.import-preview-card-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.import-preview-stats {
  display: flex;
  gap: 16px;
  margin: 12px 0;
  padding: 12px;
  background: rgba(16, 185, 129, 0.1);
  border-radius: 8px;
}

.import-preview-stat-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: #10b981;
}

.import-preview-categories {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.import-preview-category-tag {
  padding: 4px 12px;
  background: rgba(230, 162, 60, 0.15);
  border-radius: 4px;
  font-size: 13px;
  color: #e6a23c;
}

.import-preview-more-categories {
  padding: 4px 12px;
  font-size: 13px;
  color: #909399;
}

/* ═══════════════════ 浮动快捷中心 · 重构样式 ═══════════════════ */
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

/* ---- Tab 分段 ---- */
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

/* 独立卡片操作按钮 */
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

/* ---- 独立卡片视图工具栏 ---- */
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

/* ---- 暗色主题适配 ---- */
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

/* ========== 提示词查重弹窗 ========== */
.dedup-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}
.dedup-mode {
  display: flex;
  align-items: center;
  gap: 8px;
}
.dedup-mode-label {
  font-size: 13px;
  color: #475569;
}
.dedup-stats {
  font-size: 12px;
  color: #64748b;
}
.dedup-stats b {
  color: #ef4444;
}
.dedup-loading {
  padding: 40px 0;
  text-align: center;
  color: #94a3b8;
  font-size: 13px;
}
.dedup-groups {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 56vh;
  overflow-y: auto;
}
.dedup-group {
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  overflow: hidden;
  background: #f8fafc;
}
.dedup-group-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #ffffff;
  border-bottom: 1px solid #eef0f3;
}
.dedup-group-title {
  font-size: 13px;
  font-weight: 600;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 260px;
}
.dedup-group-cat {
  font-size: 11px;
  color: #4f46e5;
  background: #eef2ff;
  padding: 1px 8px;
  border-radius: 999px;
  white-space: nowrap;
}
.dedup-group-count {
  font-size: 11px;
  color: #fff;
  background: #ef4444;
  border-radius: 999px;
  padding: 1px 8px;
  white-space: nowrap;
}
.dedup-group-spacer {
  flex: 1;
}
.dedup-group-items {
  padding: 4px 0;
}
.dedup-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px dashed #eef0f3;
}
.dedup-item:last-child {
  border-bottom: none;
}
.dedup-item.selected {
  background: rgba(239, 68, 68, 0.04);
}
.dedup-item-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.dedup-item-line1 {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.dedup-lock-tag {
  margin: 0;
}
.dedup-item-name {
  font-size: 13px;
  font-weight: 500;
  color: #1f2937;
}
.dedup-item-cat {
  font-size: 11px;
  color: #94a3b8;
}
.dedup-item-date {
  margin-left: auto;
  font-size: 11px;
  color: #94a3b8;
}
.dedup-item-content {
  font-size: 12px;
  color: #64748b;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.dedup-keep-tag {
  margin: 0;
  flex-shrink: 0;
}
.dedup-footer-tip {
  font-size: 12px;
  color: #ef4444;
  margin-right: auto;
}

/* 深色主题适配 */
:root[data-theme='dark'] .dedup-group {
  background: #1e2434;
  border-color: #2e3650;
}
:root[data-theme='dark'] .dedup-group-head {
  background: #222838;
  border-color: #2e3650;
}
:root[data-theme='dark'] .dedup-group-title {
  color: #e5e7eb;
}
:root[data-theme='dark'] .dedup-item {
  border-color: #2e3650;
}
:root[data-theme='dark'] .dedup-item.selected {
  background: rgba(239, 68, 68, 0.1);
}
:root[data-theme='dark'] .dedup-item-name {
  color: #e5e7eb;
}
:root[data-theme='dark'] .dedup-mode-label {
  color: #9ca3af;
}

/* ========== 卡包合并弹窗 ========== */
.merge-mode-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 14px;
  margin-bottom: 16px;
  border-bottom: 1px solid #eef0f3;
  flex-wrap: wrap;
}
.merge-mode-desc {
  font-size: 12px;
  color: #94a3b8;
}
.merge-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.merge-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.merge-section-title {
  font-size: 13px;
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 8px;
}
.merge-section-tip {
  font-size: 11px;
  color: #94a3b8;
}
.merge-section-tip-inline {
  font-size: 11px;
  color: #94a3b8;
  font-weight: normal;
}
.merge-section-count {
  font-size: 11px;
  color: #94a3b8;
  font-weight: normal;
}
.merge-pack-list {
  max-height: 38vh;
  overflow-y: auto;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 4px;
  background: #f8fafc;
}
.merge-pack-list--radio {
  background: #ffffff;
}
.merge-pack-item {
  display: flex;
  align-items: center;
  padding: 4px 8px;
  border-radius: 8px;
  transition: background 0.15s ease;
}
.merge-pack-item:hover {
  background: #f1f5f9;
}
.merge-pack-item :deep(.el-checkbox) {
  width: 100%;
  height: auto;
  margin-right: 0;
}
.merge-pack-item :deep(.el-radio) {
  width: 100%;
  height: auto;
  margin-right: 0;
}
.merge-pack-item :deep(.el-checkbox__label),
.merge-pack-item :deep(.el-radio__label) {
  padding-left: 8px;
  width: 100%;
}
.merge-pack-info {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
}
.merge-pack-ic {
  color: #6366f1;
  flex-shrink: 0;
}
.merge-pack-name {
  font-size: 13px;
  font-weight: 500;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.merge-pack-tag {
  margin: 0;
  flex-shrink: 0;
}
.merge-pack-count {
  margin-left: auto;
  font-size: 11px;
  color: #94a3b8;
  white-space: nowrap;
}
.merge-preview {
  font-size: 13px;
  color: #475569;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 8px;
  padding: 10px 12px;
}
.merge-preview b {
  color: #3b82f6;
}
.merge-empty-subs {
  font-size: 12px;
  color: #94a3b8;
  text-align: center;
  padding: 16px 0;
  border: 1px dashed #e2e8f0;
  border-radius: 10px;
}
.merge-note {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #92400e;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 8px;
  padding: 8px 10px;
}
.merge-note .el-icon {
  color: #d97706;
  flex-shrink: 0;
}

/* 深色主题适配 */
:root[data-theme='dark'] .merge-mode-bar {
  border-color: #2e3650;
}
:root[data-theme='dark'] .merge-mode-desc {
  color: #6b7280;
}
:root[data-theme='dark'] .merge-section-title {
  color: #e5e7eb;
}
:root[data-theme='dark'] .merge-pack-list {
  background: #1a2030;
  border-color: #2e3650;
}
:root[data-theme='dark'] .merge-pack-list--radio {
  background: #1e2434;
}
:root[data-theme='dark'] .merge-pack-item:hover {
  background: #222838;
}
:root[data-theme='dark'] .merge-pack-ic {
  color: #818cf8;
}
:root[data-theme='dark'] .merge-pack-name {
  color: #e5e7eb;
}
:root[data-theme='dark'] .merge-preview {
  background: rgba(59, 130, 246, 0.12);
  border-color: #2e3650;
  color: #9ca3af;
}
:root[data-theme='dark'] .merge-preview b {
  color: #60a5fa;
}
:root[data-theme='dark'] .merge-empty-subs {
  color: #64748b;
  border-color: #2e3650;
}
:root[data-theme='dark'] .merge-note {
  background: rgba(217, 119, 6, 0.12);
  border-color: rgba(217, 119, 6, 0.35);
  color: #fcd34d;
}
:root[data-theme='dark'] .merge-note .el-icon {
  color: #f59e0b;
}

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
