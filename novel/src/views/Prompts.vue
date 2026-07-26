﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿<template>
  <div class="prompts-container">
    <div class="header">
      <div class="header-main">
        <h2>提示词管理</h2>
      </div>
      <div class="header-actions">
        <el-dropdown trigger="click" @command="handleExportCommand">
          <el-button class="btn-export" plain>
            <el-icon><Download /></el-icon>
            导出
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="exportAll">
                <el-icon><FolderOpened /></el-icon>
                整体备份（全部导出）
              </el-dropdown-item>
              <el-dropdown-item divided command="exportStandalone">
                <el-icon><Document /></el-icon>
                导出独立卡片
              </el-dropdown-item>
              <el-dropdown-item command="exportPack">
                <el-icon><FolderOpened /></el-icon>
                导出卡包
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-dropdown trigger="click" @command="handleImportCommand">
          <el-button class="btn-import" plain>
            <el-icon><Upload /></el-icon>
            导入
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="importAll">
                <el-icon><FolderOpened /></el-icon>
                整体恢复（全部导入）
              </el-dropdown-item>
              <el-dropdown-item divided command="importStandalone">
                <el-icon><Document /></el-icon>
                导入独立卡片
              </el-dropdown-item>
              <el-dropdown-item command="importPack">
                <el-icon><FolderOpened /></el-icon>
                导入卡包
              </el-dropdown-item>
              <el-dropdown-item command="importConvert">
                <el-icon><Switch /></el-icon>
                导入并转换 txt/md
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button class="btn-create-pack" @click="openCategoryDialog">
          <el-icon><FolderAdd /></el-icon>
          创建卡包
        </el-button>
        <el-button class="btn-create-prompt" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          创建提示词
        </el-button>
      </div>
    </div>

    <!-- 顶端 Tab 切换：卡包 / 独立卡片 -->
    <div class="main-tabs-wrapper">
      <el-tabs v-model="activeMainTab" class="main-tabs">
        <el-tab-pane name="pack">
          <template #label>
            <span class="main-tab-label">
              <el-icon><FolderOpened /></el-icon>
              <span>卡包</span>
              <span class="main-tab-count">{{ totalPackCards }}</span>
            </span>
          </template>
          <div class="tab-toolbar">
            <div class="pack-view-toggle">
              <button
                class="pack-view-btn"
                :class="{ active: packViewMode === 'grid' }"
                @click="packViewMode = 'grid'"
                title="网格视图"
              >
                <el-icon><Grid /></el-icon>
              </button>
              <button
                class="pack-view-btn"
                :class="{ active: packViewMode === 'list' }"
                @click="packViewMode = 'list'"
                title="列表视图"
              >
                <el-icon><List /></el-icon>
              </button>
            </div>
          </div>
          <div class="packs-wrapper" :class="packViewMode">
            <div
              v-for="category in categoryList"
              :key="category.name"
              class="pack-info-card"
              :class="{ 'drag-over': isDragging && draggedPrompt?.category !== category.name }"
              @dragover.prevent="onDragOver"
              @drop.prevent="onDrop(category.name, $event)"
            >
              <div class="pack-info-clickable" @click="goToPackDetail(category.name)">
                <div class="pack-info-icon">
                  <el-icon><FolderOpened /></el-icon>
                </div>
                <div class="pack-info-body">
                  <div class="pack-info-name">{{ category.name }}</div>
                  <div class="pack-info-meta">
                    <span class="pack-info-count">{{ category.prompts.length }} 张卡片</span>
                  </div>
                </div>
              </div>
              <div class="pack-info-spacer"></div>
              <div class="pack-info-actions" @click.stop>
                <el-dropdown trigger="click" @command="(cmd: string) => handlePackAction(cmd, category.name, category.prompts.length)">
                  <el-button class="pack-action-btn" size="small" circle>
                    <el-icon><MoreFilled /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="add">
                        <el-icon><Plus /></el-icon>
                        添加卡片
                      </el-dropdown-item>
                      <el-dropdown-item v-if="category.prompts.length > 0" command="preview">
                        <el-icon><View /></el-icon>
                        预览卡包
                      </el-dropdown-item>
                      <el-dropdown-item v-if="category.prompts.length > 0 && !DEFAULT_CATEGORIES.includes(category.name)" command="batchDelete">
                        <el-icon><Delete /></el-icon>
                        批量删除
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
        </el-tab-pane>
        <el-tab-pane name="standalone">
          <template #label>
            <span class="main-tab-label">
              <el-icon><Document /></el-icon>
              <span>独立卡片</span>
              <span class="main-tab-count">{{ uncategorizedPrompts.length + bookAnalysisUncategorizedPrompts.length }}</span>
            </span>
          </template>
          <div class="tab-toolbar">
            <div class="section-actions" v-if="combinedUncategorizedPrompts.length > 0">
              <el-tooltip content="批量删除" placement="top" :show-after="300">
                <el-button class="section-action-btn" size="small" circle @click="openStandaloneBatchDeleteDialog">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </el-tooltip>
            </div>
          </div>
          <div
            v-if="combinedUncategorizedPrompts.length > 0"
            class="standalone-cards-area"
            :class="[cardSize, { 'drag-over': isDragging }]"
            @dragover.prevent="onDragOver"
            @drop.prevent="onDrop('未分类', $event)"
          >
            <div class="standalone-cards-grid">
              <div
                v-for="prompt in combinedUncategorizedPrompts"
                :key="prompt.id"
                class="standalone-card"
                :class="[cardSize, { dragging: isDragging && draggedPrompt?.id === prompt.id }, { encrypted: prompt.card_type === 'encrypted' }]"
                draggable="true"
                @dragstart="onDragStart(prompt, $event)"
                @dragend="onDragEnd"
                @click="handleEdit(prompt)"
              >
                <div class="standalone-card-header">
                  <el-icon class="card-drag-handle"><Rank /></el-icon>
                  <span class="standalone-card-name">{{ prompt.name }}</span>
                  <el-icon v-if="prompt.card_type === 'encrypted'" class="standalone-card-lock"><Lock /></el-icon>
                </div>
                <div class="standalone-card-content">
                  <el-tag v-if="prompt.category === '拆书-未分类'" size="small" type="warning" class="standalone-card-tag">
                    拆书
                  </el-tag>
                  <el-tag v-else size="small" :type="getTagType(prompt.category)" class="standalone-card-tag">
                    {{ prompt.category }}
                  </el-tag>
                  <div class="standalone-card-preview">
                    <template v-if="prompt.card_type === 'encrypted'">
                      <el-icon><Lock /></el-icon> 内容已加密
                    </template>
                    <template v-else>
                      {{ prompt.content.slice(0, 40) }}{{ prompt.content.length > 40 ? '...' : '' }}
                    </template>
                  </div>
                </div>
                <div class="standalone-card-footer">
                  <span class="standalone-card-time">{{ formatDate(prompt.created_at) }}</span>
                  <div class="standalone-card-actions" @click.stop>
                    <el-button type="primary" link size="small" @click="handleEdit(prompt)">
                      <el-icon><Edit /></el-icon>
                    </el-button>
                    <el-button type="primary" link size="small" @click="handlePreview(prompt)">
                      <el-icon><View /></el-icon>
                    </el-button>
                    <el-button type="danger" link size="small" @click="handleDelete(prompt)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
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
        </el-tab-pane>
        <el-tab-pane name="bookAnalysis">
          <template #label>
            <span class="main-tab-label">
              <el-icon><Reading /></el-icon>
              <span>拆书提示词</span>
              <span class="main-tab-count">{{ totalBookAnalysisCards }}</span>
            </span>
          </template>
          <div class="tab-toolbar">
            <div class="pack-view-toggle">
              <button
                class="pack-view-btn"
                :class="{ active: packViewMode === 'grid' }"
                @click="packViewMode = 'grid'"
                title="网格视图"
              >
                <el-icon><Grid /></el-icon>
              </button>
              <button
                class="pack-view-btn"
                :class="{ active: packViewMode === 'list' }"
                @click="packViewMode = 'list'"
                title="列表视图"
              >
                <el-icon><List /></el-icon>
              </button>
            </div>
            <div class="tab-toolbar-right">
              <el-dropdown trigger="click" @command="handleBookAnalysisExportCommand">
                <el-button class="btn-export" size="small">
                  <el-icon><Download /></el-icon>
                  导出
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="fullBackup">完整备份（全部导出）</el-dropdown-item>
                    <el-dropdown-item command="pack">导出拆书卡包</el-dropdown-item>
                    <el-dropdown-item command="standalone">导出拆书独立卡片</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-dropdown trigger="click" @command="handleBookAnalysisImportCommand">
                <el-button class="btn-import" size="small">
                  <el-icon><Upload /></el-icon>
                  导入
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="pack">导入拆书卡包</el-dropdown-item>
                    <el-dropdown-item command="standalone">导入拆书独立卡片</el-dropdown-item>
                    <el-dropdown-item command="legacyBackup">导入旧版完整备份</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-button class="btn-create-pack" size="small" @click="openBookAnalysisCategoryDialog">
                <el-icon><FolderAdd /></el-icon>
                创建拆书卡包
              </el-button>
              <el-button class="btn-create-prompt" size="small" @click="handleCreateBookAnalysisPrompt()">
                <el-icon><Plus /></el-icon>
                创建拆书提示词
              </el-button>
            </div>
          </div>
          <!-- 拆书卡包列表 -->
          <div class="packs-wrapper" :class="packViewMode">
            <div
              v-for="category in bookAnalysisCategoryList"
              :key="category.name"
              class="pack-info-card"
              :class="{ 'drag-over': isDragging && draggedPrompt?.category !== category.name }"
              @dragover.prevent="onDragOver"
              @drop.prevent="onBookAnalysisDrop(category.name, $event)"
            >
              <div class="pack-info-clickable" @click="goToPackDetail(category.name)">
                <div class="pack-info-icon">
                  <el-icon><Reading /></el-icon>
                </div>
                <div class="pack-info-body">
                  <div class="pack-info-name">{{ category.name.replace(BOOK_ANALYSIS_PREFIX, '') }}</div>
                  <div class="pack-info-meta">
                    <span class="pack-info-count">{{ category.prompts.length }} 张卡片</span>
                  </div>
                </div>
              </div>
              <div class="pack-info-spacer"></div>
              <div class="pack-info-actions" @click.stop>
                <el-dropdown trigger="click" @command="(cmd: string) => handleBookAnalysisPackAction(cmd, category.name, category.prompts.length)">
                  <el-button class="pack-action-btn" size="small" circle>
                    <el-icon><MoreFilled /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="add">
                        <el-icon><Plus /></el-icon>
                        添加卡片
                      </el-dropdown-item>
                      <el-dropdown-item v-if="category.prompts.length > 0" command="preview">
                        <el-icon><View /></el-icon>
                        预览卡包
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
            v-if="bookAnalysisCategoryList.length === 0"
            class="empty-standalone-hint"
            @click="handleCreateBookAnalysisPrompt()"
          >
            <el-icon><Plus /></el-icon>
            <span>暂无拆书提示词，点击创建或导入拆书卡包</span>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

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

    <!-- 导出拆书卡包弹窗 -->
    <el-dialog
      v-model="exportBookAnalysisPackDialogVisible"
      title="导出拆书卡包"
      width="600px"
      :close-on-click-modal="false"
      class="export-dialog"
      append-to-body
    >
      <div class="export-dialog-header">
        <h4>请选择要导出的拆书卡包</h4>
        <p>选择需要导出的拆书卡包，导出的文件将包含卡包内的所有卡片。</p>
      </div>
      <div class="export-dialog-content">
        <el-checkbox-group v-model="selectedExportBookAnalysisPacks">
          <div v-for="category in bookAnalysisCategoryList" :key="category.name" class="export-checkbox-item">
            <el-checkbox :label="category.name">
              <div class="export-item-info">
                <span class="export-item-name">{{ category.name.replace(BOOK_ANALYSIS_PREFIX, '') }}</span>
                <span class="export-item-count">{{ category.prompts.length }} 张卡片</span>
              </div>
            </el-checkbox>
          </div>
        </el-checkbox-group>
        <el-empty v-if="bookAnalysisCategoryList.length === 0" description="暂无拆书卡包可导出" />
      </div>
      <template #footer>
        <el-button @click="exportBookAnalysisPackDialogVisible = false">取消</el-button>
        <el-button type="primary" :disabled="selectedExportBookAnalysisPacks.length === 0" @click="executeExportBookAnalysisPack">
          导出卡包
        </el-button>
      </template>
    </el-dialog>

    <!-- 导出拆书独立卡片弹窗 -->
    <el-dialog
      v-model="exportBookAnalysisStandaloneDialogVisible"
      title="导出拆书独立卡片"
      width="800px"
      :close-on-click-modal="false"
      class="export-dialog"
      append-to-body
    >
      <div class="export-dialog-header">
        <h4>请选择要导出的拆书独立卡片</h4>
        <p>勾选需要导出的卡片，然后点击导出按钮。导出的文件为JSON格式。</p>
      </div>
      <div class="export-dialog-content">
        <el-checkbox-group v-model="selectedExportBookAnalysisCards">
          <div v-for="prompt in bookAnalysisUncategorizedPrompts" :key="prompt.id" class="export-checkbox-item">
            <el-checkbox :label="prompt.id">
              <div class="export-item-info">
                <span class="export-item-name">{{ prompt.name }}</span>
                <span class="export-item-preview">{{ prompt.content.slice(0, 30) }}...</span>
              </div>
            </el-checkbox>
          </div>
        </el-checkbox-group>
        <el-empty v-if="bookAnalysisUncategorizedPrompts.length === 0" description="暂无拆书独立卡片可导出" />
      </div>
      <template #footer>
        <el-button @click="exportBookAnalysisStandaloneDialogVisible = false">取消</el-button>
        <el-button type="primary" :disabled="selectedExportBookAnalysisCards.length === 0" @click="executeExportBookAnalysisStandalone">
          导出 ({{ selectedExportBookAnalysisCards.length }})
        </el-button>
      </template>
    </el-dialog>

    <!-- 导入拆书卡包弹窗 -->
    <el-dialog
      v-model="importBookAnalysisPackDialogVisible"
      title="导入拆书卡包"
      width="600px"
      :close-on-click-modal="false"
      class="import-dialog"
      append-to-body
    >
      <div class="import-dialog-header">
        <h4>导入拆书卡包文件</h4>
        <p>导入的卡包将自动添加"拆书-"前缀，方便在拆书库中使用。</p>
      </div>
      <div class="import-dialog-content">
        <el-upload
          drag
          action="#"
          accept=".json"
          :auto-upload="false"
          :show-file-list="false"
          :on-change="handleImportBookAnalysisPackFileChange"
        >
          <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
          <div class="el-upload__text">将卡包文件拖到此处，或<em>点击上传</em></div>
          <template #tip>
            <div class="el-upload__tip">仅支持 .json 格式的导出文件</div>
          </template>
        </el-upload>
        <div v-if="importBookAnalysisPackPreview" class="import-preview">
          <div class="import-preview-header">
            <span>预览导入内容</span>
            <span class="import-preview-count">{{ importBookAnalysisPackPreview.prompts.length }} 张卡片</span>
          </div>
          <div class="import-preview-pack-name">
            卡包名称：<strong>{{ importBookAnalysisPackPreview.packName }}</strong>
            <span class="import-preview-prefix-hint">→ 导入后变为：{{ BOOK_ANALYSIS_PREFIX }}{{ importBookAnalysisPackPreview.packName }}</span>
          </div>
          <div class="import-preview-cards">
            <div v-for="(card, index) in importBookAnalysisPackPreview.prompts.slice(0, 5)" :key="index" class="import-preview-card">
              <span class="import-preview-card-name">{{ card.name }}</span>
              <span class="import-preview-card-content">{{ card.content.slice(0, 40) }}...</span>
            </div>
            <div v-if="importBookAnalysisPackPreview.prompts.length > 5" class="import-preview-more">
              还有 {{ importBookAnalysisPackPreview.prompts.length - 5 }} 张卡片...
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="importBookAnalysisPackDialogVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!importBookAnalysisPackPreview" @click="executeImportBookAnalysisPack">
          导入卡包
        </el-button>
      </template>
    </el-dialog>

    <!-- 导入拆书独立卡片弹窗 -->
    <el-dialog
      v-model="importBookAnalysisStandaloneDialogVisible"
      title="导入拆书独立卡片"
      width="800px"
      :close-on-click-modal="false"
      class="import-dialog"
      append-to-body
    >
      <div class="import-dialog-header">
        <h4>导入拆书独立卡片文件</h4>
        <p>导入的卡片将归类为"拆书-未分类"，方便在拆书库中使用。</p>
      </div>
      <div class="import-dialog-content">
        <el-upload
          drag
          action="#"
          accept=".json"
          :auto-upload="false"
          :show-file-list="false"
          :on-change="handleImportBookAnalysisStandaloneFileChange"
        >
          <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
          <div class="el-upload__text">将卡片文件拖到此处，或<em>点击上传</em></div>
          <template #tip>
            <div class="el-upload__tip">仅支持 .json 格式的导出文件</div>
          </template>
        </el-upload>
        <div v-if="importBookAnalysisStandalonePreview.length > 0" class="import-preview">
          <div class="import-preview-header">
            <span>预览导入内容 ({{ importBookAnalysisStandalonePreview.length }} 张卡片)</span>
            <el-checkbox v-model="selectAllImportBookAnalysisCards" @change="handleSelectAllImportBookAnalysisCards">全选</el-checkbox>
          </div>
          <el-checkbox-group v-model="selectedImportBookAnalysisCards" class="import-preview-cards-list">
            <div v-for="card in importBookAnalysisStandalonePreview" :key="card.name" class="import-preview-card-item">
              <el-checkbox :label="card.name">
                <div class="import-preview-card-info">
                  <span class="import-preview-card-name">{{ card.name }}</span>
                  <span class="import-preview-card-content">{{ card.content.slice(0, 50) }}...</span>
                </div>
              </el-checkbox>
            </div>
          </el-checkbox-group>
        </div>
      </div>
      <template #footer>
        <el-button @click="importBookAnalysisStandaloneDialogVisible = false">取消</el-button>
        <el-button type="primary" :disabled="selectedImportBookAnalysisCards.length === 0" @click="executeImportBookAnalysisStandalone">
          导入 ({{ selectedImportBookAnalysisCards.length }})
        </el-button>
      </template>
    </el-dialog>

    <!-- 导入旧版完整备份弹窗 -->
    <el-dialog
      v-model="importBookAnalysisLegacyDialogVisible"
      title="导入旧版拆书库完整备份"
      width="600px"
      :close-on-click-modal="false"
      class="import-dialog"
      append-to-body
    >
      <div class="import-dialog-header">
        <h4>导入旧版拆书库的完整备份文件</h4>
        <p>支持导入旧版拆书库导出的完整备份JSON文件，导入后自动转换为拆书卡包/独立卡片。</p>
      </div>
      <div class="import-dialog-content">
        <el-upload
          drag
          action="#"
          accept=".json"
          :auto-upload="false"
          :show-file-list="false"
          :on-change="handleImportBookAnalysisLegacyFileChange"
        >
          <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
          <div class="el-upload__text">将备份文件拖到此处，或<em>点击上传</em></div>
          <template #tip>
            <div class="el-upload__tip">支持旧版拆书库导出的完整备份JSON文件</div>
          </template>
        </el-upload>
        <div v-if="importBookAnalysisLegacyPreview" class="import-preview">
          <div class="import-preview-header">
            <span>预览导入内容</span>
          </div>
          <div class="import-preview-stats">
            <div class="import-preview-stat-item">
              <el-icon><FolderOpened /></el-icon>
              <span>卡包数量：{{ importBookAnalysisLegacyPreview.packCount }}</span>
            </div>
            <div class="import-preview-stat-item">
              <el-icon><Document /></el-icon>
              <span>提示词总数：{{ importBookAnalysisLegacyPreview.totalPrompts }}</span>
            </div>
          </div>
          <div class="import-preview-categories">
            <div v-for="cat in importBookAnalysisLegacyPreview.categories.slice(0, 5)" :key="cat" class="import-preview-category-tag">
              {{ cat }}
            </div>
            <div v-if="importBookAnalysisLegacyPreview.categories.length > 5" class="import-preview-more-categories">
              还有 {{ importBookAnalysisLegacyPreview.categories.length - 5 }} 个分类...
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="importBookAnalysisLegacyDialogVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!importBookAnalysisLegacyPreview" @click="executeImportBookAnalysisLegacy">
          导入并转换
        </el-button>
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
            选择需要导出的卡包，导出的文件将包含卡包内的所有卡片。
          </el-alert>
        </div>
        <div v-if="categoryList.length > 0" class="export-pack-list">
          <div
            v-for="category in categoryList"
            :key="category.name"
            class="export-pack-item"
            :class="{ selected: selectedPack === category.name }"
            @click="selectedPack = category.name"
          >
            <el-radio :model-value="selectedPack" :label="category.name">
              <div class="pack-item-info">
                <el-icon><FolderOpened /></el-icon>
                <span class="pack-item-name">{{ category.name }}</span>
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
              <div class="el-upload__tip">只能上传 JSON 格式的卡包文件</div>
            </template>
          </el-upload>
        </div>
        <div v-if="importPackPreview" class="import-pack-preview">
          <div class="import-pack-info">
            <el-icon><FolderOpened /></el-icon>
            <span class="import-pack-name">{{ importPackPreview.packName }}</span>
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
      width="900px"
      top="6vh"
      :close-on-click-modal="false"
      destroy-on-close
      class="pack-edit-dialog"
    >
      <div class="pack-edit-layout">
        <!-- 左侧：表单字段 (2/3) -->
        <div class="pack-edit-left">
          <el-form :model="formData" label-position="top" class="pack-edit-form">
            <el-form-item label="名称" required>
              <el-input
                v-model="formData.name"
                placeholder="请输入名称"
                maxlength="100"
                show-word-limit
              />
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
                <el-radio-group v-model="formData.card_type" @change="handleCardTypeChange">
                  <el-radio value="normal">
                    <el-icon><Document /></el-icon>
                    普通卡片
                  </el-radio>
                  <el-radio value="encrypted">
                    <el-icon><Lock /></el-icon>
                    加密卡片
                  </el-radio>
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
            <el-form-item label="简介">
              <SplitRichTextEditor
                v-model="formData.description"
                placeholder="用于介绍这个提示词，非必填"
                class="description-rich-editor"
              />
            </el-form-item>
            <el-form-item label="内容" required>
              <el-input
                v-model="formData.content"
                type="textarea"
                :rows="12"
                placeholder="请输入提示词内容（作为AI的system层指令），使用 ${字段名} 引用变量"
                class="pack-edit-content-input"
              />
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
            <span class="pack-edit-right-title">字段配置</span>
            <el-button type="primary" size="small" plain @click="addField">
              <el-icon><Plus /></el-icon>
              添加
            </el-button>
          </div>
          <div class="pack-edit-right-tip">对应内容里的 <code>${字段名}</code></div>
          <div class="pack-edit-right-body">
            <div v-if="fieldsConfig.length === 0" class="pack-edit-empty">暂无字段</div>
            <div v-else class="pack-edit-fields">
              <div v-for="(field, index) in fieldsConfig" :key="index" class="pack-edit-field">
                <div class="pack-edit-field-head">
                  <el-tag :type="field.required ? 'danger' : 'info'" size="small" effect="light">
                    {{ field.required ? '必填' : '选填' }}
                  </el-tag>
                  <el-input
                    v-model="field.name"
                    size="small"
                    class="pack-edit-field-name"
                    placeholder="字段名"
                    @change="updateFieldName(index, field.name)"
                  />
                  <el-button size="small" link @click="copyFieldName(field.name)" title="复制">
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                  <el-switch
                    v-model="field.required"
                    size="small"
                    inline-prompt
                    active-text="必"
                    inactive-text="选"
                    style="--el-switch-on-color: #f56c6c; --el-switch-off-color: #c0c4cc;"
                  />
                  <el-button size="small" type="danger" link @click="removeField(index)" title="删除">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <div class="pack-edit-field-body">
                  <el-input v-model="field.label" size="small" placeholder="显示名" />
                  <el-select v-model="field.type" size="small" placeholder="类型" style="width: 100%">
                    <el-option label="单行文本" value="text" />
                    <el-option label="多行文本" value="textarea" />
                    <el-option label="下拉选择" value="select" />
                  </el-select>
                  <el-input
                    v-model="field.description"
                    size="small"
                    placeholder="说明（可选）"
                  />
                </div>
                <div class="pack-edit-field-order">
                  <el-button size="small" link :disabled="index === 0" @click="moveField(index, index - 1)" title="上移">
                    <el-icon><ArrowUp /></el-icon>
                  </el-button>
                  <el-button size="small" link :disabled="index === fieldsConfig.length - 1" @click="moveField(index, index + 1)" title="下移">
                    <el-icon><ArrowDown /></el-icon>
                  </el-button>
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
              </div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

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
import { ref, onMounted, onBeforeUnmount, watch, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, ArrowUp, ArrowDown, Delete, CopyDocument, FolderAdd, FolderOpened, FolderRemove, Edit, Document, Rank, Check, Close, Download, Upload, UploadFilled, View, Switch, InfoFilled, Lock, Grid, List, Setting, Reading, MoreFilled } from '@element-plus/icons-vue'
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
  version: ''
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
const importPackPreview = ref<{ packName: string; prompts: Prompt[] } | null>(null)

// 拆书导入导出相关状态
const exportBookAnalysisPackDialogVisible = ref(false)
const selectedExportBookAnalysisPacks = ref<string[]>([])
const exportBookAnalysisStandaloneDialogVisible = ref(false)
const selectedExportBookAnalysisCards = ref<number[]>([])
const importBookAnalysisPackDialogVisible = ref(false)
const importBookAnalysisPackPreview = ref<{ packName: string; prompts: Prompt[] } | null>(null)
const importBookAnalysisStandaloneDialogVisible = ref(false)
const importBookAnalysisStandalonePreview = ref<Prompt[]>([])
const selectedImportBookAnalysisCards = ref<string[]>([])
const selectAllImportBookAnalysisCards = ref(false)
const importBookAnalysisLegacyDialogVisible = ref(false)
const importBookAnalysisLegacyPreview = ref<{ packCount: number; totalPrompts: number; categories: string[]; prompts: any[] } | null>(null)

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
const activeMainTab = ref<'pack' | 'standalone' | 'bookAnalysis'>('pack')

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
    version: ''
  }
  fieldsConfig.value = []
  formSubcategories.value = []
  newSubcategory.value = ''
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
    version: ''
  }
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
}>>([])

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
    version: ''
  }
  fieldsConfig.value = []
  formSubcategories.value = []
  newSubcategory.value = ''
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
    version: prompt.version || ''
  }
  // 重置字段配置
  fieldsConfig.value = []
  
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
      required: field.required !== undefined ? field.required : true
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

// 从提示词内容中提取字段名称
const extractFieldsFromContent = (content: string) => {
  const fieldRegex = /\$\{([^}]+)\}/g
  const fields = new Set<string>()
  let match
  while ((match = fieldRegex.exec(content)) !== null) {
    if (match && match[1]) {
      fields.add(match[1].trim())
    }
  }
  
  // 为每个字段创建配置
  fields.forEach(fieldName => {
    fieldsConfig.value.push({
      name: fieldName,
      label: fieldName,
      type: 'text',
      options: [],
      optionLabels: [],
      optionsText: '',
      optionLabelsText: '',
      description: '',
      required: true
    })
  })
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
    required: true
  })
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
          required: true
        })
      }
    })
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
        required: field.required
      })),
      subcategories: formSubcategories.value,
      card_type: formData.value.card_type || 'normal',
      password: formData.value.password || null
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

// 处理整体备份文件选择
const handleImportAllFileChange = async (file: any) => {
  try {
    const text = await file.raw.text()
    const decodedText = decodeExportData(text)
    const backupData = JSON.parse(decodedText)
    
    // 验证备份数据格式
    if (!backupData.type || backupData.type !== 'full-backup' || !backupData.data) {
      throw new Error('无效的备份文件格式')
    }
    
    const { standalone, packs } = backupData.data
    
    // 计算预览数据
    const standaloneCount = standalone ? standalone.length : 0
    const packCount = packs ? Object.keys(packs).length : 0
    let totalPrompts = standaloneCount
    
    if (packs) {
      Object.values(packs).forEach((pack: any) => {
        totalPrompts += Array.isArray(pack) ? pack.length : 0
      })
    }
    
    importAllPreview.value = {
      standaloneCount,
      packCount,
      totalPrompts,
      data: backupData
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
    
    const { standalone, packs } = importAllPreview.value.data.data
    
    // 先删除所有现有数据
    const deletePromises = prompts.value.map(p => promptAPI.delete(p.id))
    await Promise.all(deletePromises)
    
    // 导入独立卡片
    if (standalone && standalone.length > 0) {
      for (const prompt of standalone) {
        await promptAPI.create({
          name: prompt.name,
          content: prompt.content,
          category: prompt.category || '未分类',
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
        if (Array.isArray(promptsInPack)) {
          for (const prompt of promptsInPack) {
            await promptAPI.create({
              name: prompt.name,
              content: prompt.content,
              category: packName,
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
      customCategories.value = Object.keys(packs)
      saveCategoriesToStorage(customCategories.value)
    }
    
    ElMessage.success('整体恢复成功')
    importAllDialogVisible.value = false
    importAllPreview.value = null
    await fetchPrompts()
  } catch (error) {
    // 取消导入
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

// 执行导出卡包
const executeExportPack = () => {
  if (!selectedPack.value) return
  
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
      } else if (data.type === 'card-pack') {
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
  
  for (const card of cardsToImport) {
    try {
      const res = await promptAPI.create({
        name: card.name,
        content: card.content,
        category: '未分类',
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
    ElMessage.success(`成功导入 ${successCount} 张独立卡片`)
  } else {
    ElMessage.warning(`导入完成：成功 ${successCount} 张，失败 ${failCount} 张`)
  }
  
  importStandaloneDialogVisible.value = false
}

// 打开导入卡包弹窗
const openImportPackDialog = () => {
  importPackPreview.value = null
  importPackDialogVisible.value = true
}

// 处理卡包文件选择
const handlePackFileChange = (file: any) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const content = e.target?.result as string
      const decodedContent = decodeExportData(content)
      const data = JSON.parse(decodedContent)
      
      if (data.type === 'card-pack' && Array.isArray(data.prompts)) {
        importPackPreview.value = {
          packName: data.packName,
          prompts: data.prompts
        }
        ElMessage.success(`成功读取卡包"${data.packName}"，包含 ${data.prompts.length} 张卡片`)
      } else if (data.type === 'standalone-cards') {
        ElMessage.warning('这是独立卡片文件，请使用"导入独立卡片"功能')
      } else {
        ElMessage.error('文件格式不正确')
      }
    } catch (error) {
      ElMessage.error('文件解析失败，请确保是有效的导出文件')
    }
  }
  reader.readAsText(file.raw)
}

// 执行导入卡包
const executeImportPack = async () => {
  if (!importPackPreview.value) return
  
  const { packName, prompts: cardsToImport } = importPackPreview.value
  
  // 检查卡包是否已存在，如果不存在则创建
  if (!customCategories.value.includes(packName) && packName !== '默认') {
    customCategories.value.push(packName)
    saveCategoriesToStorage(customCategories.value)
  }
  
  let successCount = 0
  let failCount = 0
  
  for (const card of cardsToImport) {
    try {
      const res = await promptAPI.create({
        name: card.name,
        content: card.content,
        category: packName,
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
    ElMessage.success(`成功导入卡包"${packName}"，包含 ${successCount} 张卡片`)
  } else {
    ElMessage.warning(`导入完成：成功 ${successCount} 张，失败 ${failCount} 张`)
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
  
  for (const card of cardsToImport) {
    try {
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
      ElMessage.success(`成功导入卡包"${targetCategory}"，包含 ${successCount} 张卡片`)
    } else {
      ElMessage.warning(`导入完成：成功 ${successCount} 张，失败 ${failCount} 张`)
    }
  } else {
    if (failCount === 0) {
      ElMessage.success(`成功导入 ${successCount} 张独立卡片`)
    } else {
      ElMessage.warning(`导入完成：成功 ${successCount} 张，失败 ${failCount} 张`)
    }
  }
  
  mergeAsPack.value = false
  mergePackName.value = ''
  convertUploadRef.value?.clearFiles()
  importConvertDialogVisible.value = false
}

// ===== 拆书导入导出函数 =====

// 拆书导出命令处理
const handleBookAnalysisExportCommand = (command: string) => {
  if (command === 'fullBackup') {
    executeExportBookAnalysisFullBackup()
  } else if (command === 'pack') {
    selectedExportBookAnalysisPacks.value = []
    exportBookAnalysisPackDialogVisible.value = true
  } else if (command === 'standalone') {
    selectedExportBookAnalysisCards.value = []
    exportBookAnalysisStandaloneDialogVisible.value = true
  }
}

// 拆书导入命令处理
const handleBookAnalysisImportCommand = (command: string) => {
  if (command === 'pack') {
    importBookAnalysisPackPreview.value = null
    importBookAnalysisPackDialogVisible.value = true
  } else if (command === 'standalone') {
    importBookAnalysisStandalonePreview.value = []
    selectedImportBookAnalysisCards.value = []
    importBookAnalysisStandaloneDialogVisible.value = true
  } else if (command === 'legacyBackup') {
    importBookAnalysisLegacyPreview.value = null
    importBookAnalysisLegacyDialogVisible.value = true
  }
}

// 执行导出拆书完整备份
const executeExportBookAnalysisFullBackup = () => {
  // 收集所有拆书提示词
  const allBookAnalysisPrompts = prompts.value.filter(p => p.category.startsWith(BOOK_ANALYSIS_PREFIX))
  
  if (allBookAnalysisPrompts.length === 0) {
    ElMessage.warning('没有拆书提示词可导出')
    return
  }

  // 按分类组织
  const packs: Record<string, any[]> = {}
  const standalone: any[] = []

  for (const prompt of allBookAnalysisPrompts) {
    const exportPrompt = {
      id: prompt.id,
      name: prompt.name,
      content: prompt.content,
      category: prompt.category.replace(BOOK_ANALYSIS_PREFIX, ''), // 导出时去掉前缀
      order_num: prompt.order_num || 0,
      created_at: prompt.created_at,
      fields: prompt.fields || [],
      card_type: prompt.card_type || 'normal',
      password: prompt.password || null,
      description: prompt.description || '',
      creator_name: prompt.creator_name || '',
      version: prompt.version || '',
      subcategories: prompt.subcategories || []
    }

    const cleanCategory = prompt.category.replace(BOOK_ANALYSIS_PREFIX, '')
    if (cleanCategory === '未分类') {
      standalone.push(exportPrompt)
    } else {
      if (!packs[cleanCategory]) {
        packs[cleanCategory] = []
      }
      packs[cleanCategory].push(exportPrompt)
    }
  }

  // 构建完整备份数据
  const backupData = {
    type: 'book-analysis-full-backup',
    version: '1.0',
    exportedAt: new Date().toISOString(),
    data: {
      packs,
      standalone
    },
    categories: Object.keys(packs)
  }

  // 导出文件
  const blob = new Blob([JSON.stringify(backupData, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `拆书提示词完整备份_${new Date().toISOString().slice(0, 10)}.json`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)

  const totalCount = allBookAnalysisPrompts.length
  const packCount = Object.keys(packs).length
  const standaloneCount = standalone.length
  
  ElMessage.success(`已导出完整备份：${packCount} 个卡包，${standaloneCount} 张独立卡片，共 ${totalCount} 条提示词`)
}

// 执行导出拆书卡包
const executeExportBookAnalysisPack = () => {
  if (selectedExportBookAnalysisPacks.value.length === 0) return

  const packsToExport = selectedExportBookAnalysisPacks.value.map(categoryName => {
    const category = bookAnalysisCategoryList.value.find(c => c.name === categoryName)
    return {
      packName: categoryName.replace(BOOK_ANALYSIS_PREFIX, ''),
      prompts: category ? category.prompts.map(p => ({
        id: p.id,
        name: p.name,
        content: p.content,
        category: categoryName,
        order_num: p.order_num || 0,
        created_at: p.created_at,
        fields: p.fields || [],
        card_type: p.card_type || 'normal',
        password: p.password || null
      })) : []
    }
  })

  const exportData = {
    type: 'book-analysis-packs',
    exportedAt: new Date().toISOString(),
    packs: packsToExport
  }

  const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `拆书卡包导出_${new Date().toISOString().slice(0, 10)}.json`
  link.click()
  URL.revokeObjectURL(url)

  ElMessage.success(`成功导出 ${packsToExport.length} 个拆书卡包`)
  exportBookAnalysisPackDialogVisible.value = false
}

// 执行导出拆书独立卡片
const executeExportBookAnalysisStandalone = () => {
  if (selectedExportBookAnalysisCards.value.length === 0) return

  const cardsToExport = selectedExportBookAnalysisCards.value.map(id => {
    const prompt = bookAnalysisUncategorizedPrompts.value.find(p => p.id === id)
    if (!prompt) return null
    return {
      id: prompt.id,
      name: prompt.name,
      content: prompt.content,
      category: '拆书-未分类',
      order_num: prompt.order_num || 0,
      created_at: prompt.created_at,
      fields: prompt.fields || [],
      card_type: prompt.card_type || 'normal',
      password: prompt.password || null
    }
  }).filter((card): card is NonNullable<typeof card> => card !== null)

  const exportData = {
    type: 'book-analysis-standalone',
    exportedAt: new Date().toISOString(),
    prompts: cardsToExport
  }

  const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `拆书独立卡片导出_${new Date().toISOString().slice(0, 10)}.json`
  link.click()
  URL.revokeObjectURL(url)

  ElMessage.success(`成功导出 ${cardsToExport.length} 张拆书独立卡片`)
  exportBookAnalysisStandaloneDialogVisible.value = false
}

// 处理拆书卡包文件选择
const handleImportBookAnalysisPackFileChange = (file: any) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const content = e.target?.result as string
      const data = JSON.parse(content)

      // 支持两种格式：单个卡包或多个卡包
      if (data.type === 'book-analysis-packs' && data.packs) {
        // 多个卡包格式，取第一个
        if (data.packs.length > 0) {
          importBookAnalysisPackPreview.value = {
            packName: data.packs[0].packName,
            prompts: data.packs[0].prompts
          }
        }
      } else if (data.packName && data.prompts) {
        // 单个卡包格式
        importBookAnalysisPackPreview.value = {
          packName: data.packName,
          prompts: data.prompts
        }
      } else {
        ElMessage.warning('文件格式不正确，请使用拆书卡包导出文件')
        importBookAnalysisPackPreview.value = null
      }
    } catch (error) {
      ElMessage.error('文件解析失败，请确保是有效的JSON文件')
      importBookAnalysisPackPreview.value = null
    }
  }
  reader.readAsText(file.raw)
}

// 执行导入拆书卡包
const executeImportBookAnalysisPack = async () => {
  if (!importBookAnalysisPackPreview.value) return

  const { packName, prompts: cardsToImport } = importBookAnalysisPackPreview.value
  // 自动加拆书前缀
  const fullPackName = packName.startsWith(BOOK_ANALYSIS_PREFIX) ? packName : BOOK_ANALYSIS_PREFIX + packName

  // 检查卡包是否已存在，如果不存在则创建
  if (!customCategories.value.includes(fullPackName)) {
    customCategories.value.push(fullPackName)
    saveCategoriesToStorage(customCategories.value)
  }

  let successCount = 0
  let failCount = 0

  for (const card of cardsToImport) {
    try {
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
    ElMessage.success(`成功导入拆书卡包"${fullPackName}"，包含 ${successCount} 张卡片`)
  } else {
    ElMessage.warning(`导入完成：成功 ${successCount} 张，失败 ${failCount} 张`)
  }

  importBookAnalysisPackDialogVisible.value = false
}

// 处理拆书独立卡片文件选择
const handleImportBookAnalysisStandaloneFileChange = (file: any) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const content = e.target?.result as string
      const data = JSON.parse(content)

      if (data.type === 'book-analysis-standalone' && data.prompts) {
        importBookAnalysisStandalonePreview.value = data.prompts
      } else if (Array.isArray(data)) {
        importBookAnalysisStandalonePreview.value = data.filter((p: any) => p && p.name && p.content)
      } else if (data.prompts && Array.isArray(data.prompts)) {
        importBookAnalysisStandalonePreview.value = data.prompts.filter((p: any) => p && p.name && p.content)
      } else {
        ElMessage.warning('文件格式不正确，请使用拆书独立卡片导出文件')
        importBookAnalysisStandalonePreview.value = []
      }

      selectedImportBookAnalysisCards.value = []
      selectAllImportBookAnalysisCards.value = false
    } catch (error) {
      ElMessage.error('文件解析失败，请确保是有效的JSON文件')
      importBookAnalysisStandalonePreview.value = []
    }
  }
  reader.readAsText(file.raw)
}

// 全选/取消全选拆书导入卡片
const handleSelectAllImportBookAnalysisCards = (val: boolean) => {
  if (val) {
    selectedImportBookAnalysisCards.value = importBookAnalysisStandalonePreview.value.map(card => card.name)
  } else {
    selectedImportBookAnalysisCards.value = []
  }
}

// 执行导入拆书独立卡片
const executeImportBookAnalysisStandalone = async () => {
  const cardsToImport = importBookAnalysisStandalonePreview.value
    .filter(card => selectedImportBookAnalysisCards.value.includes(card.name))

  if (cardsToImport.length === 0) return

  let successCount = 0
  let failCount = 0

  for (const card of cardsToImport) {
    try {
      const res = await promptAPI.create({
        name: card.name,
        content: card.content,
        category: '拆书-未分类',
        order_num: card.order_num || 0,
        created_at: card.created_at,
        fields: card.fields || [],
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
    ElMessage.success(`成功导入 ${successCount} 张拆书独立卡片`)
  } else {
    ElMessage.warning(`导入完成：成功 ${successCount} 张，失败 ${failCount} 张`)
  }

  importBookAnalysisStandaloneDialogVisible.value = false
}

// 处理旧版完整备份文件选择
const handleImportBookAnalysisLegacyFileChange = (file: any) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const content = e.target?.result as string
      let data: any
      
      try {
        data = JSON.parse(content)
      } catch (parseError) {
        console.error('JSON解析错误:', parseError)
        ElMessage.error('文件不是有效的JSON格式')
        importBookAnalysisLegacyPreview.value = null
        return
      }

      console.log('解析到的数据结构:', data)

      // 解析旧版格式
      let prompts: any[] = []
      let categories: string[] = []

      // 格式1: 直接是 prompts 数组
      if (Array.isArray(data)) {
        console.log('识别为格式1: 直接数组')
        prompts = data.filter((p: any) => p && p.name && p.content)
        categories = Array.from(new Set(prompts.map((p: any) => p.category || '未分类').filter((c: string) => c !== '未分类')))
      }
      // 格式2: { prompts: [...], categories: [...] }
      else if (data.prompts && Array.isArray(data.prompts)) {
        console.log('识别为格式2: { prompts, categories }')
        prompts = data.prompts.filter((p: any) => p && p.name && p.content)
        categories = data.categories || Array.from(new Set(prompts.map((p: any) => p.category || '未分类').filter((c: string) => c !== '未分类')))
      }
      // 格式3: { type: 'full-backup', data: { standalone: [...], packs: {...} } }
      else if (data.type === 'full-backup' && data.data) {
        console.log('识别为格式3: full-backup')
        const standalone = data.data.standalone || []
        const packs = data.data.packs || {}
        
        // 合并 standalone 和 packs
        prompts = [...standalone]
        for (const [packName, packPrompts] of Object.entries(packs)) {
          if (Array.isArray(packPrompts)) {
            prompts.push(...packPrompts.map((p: any) => ({ ...p, category: packName })))
          }
        }
        categories = Object.keys(packs).filter((c: string) => c !== '未分类')
      }
      // 格式4: 新版拆书完整备份 { type: 'book-analysis-full-backup', data: { packs, standalone } }
      else if (data.type === 'book-analysis-full-backup' && data.data) {
        console.log('识别为格式4: book-analysis-full-backup')
        const standalone = data.data.standalone || []
        const packs = data.data.packs || {}
        
        // 合并 standalone 和 packs
        prompts = [...standalone]
        for (const [packName, packPrompts] of Object.entries(packs)) {
          if (Array.isArray(packPrompts)) {
            prompts.push(...packPrompts.map((p: any) => ({ ...p, category: packName })))
          }
        }
        categories = Object.keys(packs).filter((c: string) => c !== '未分类')
      }
      // 格式5: localStorage 原始格式 { prompts: [...], categories: [...] }
      else if (data.prompts) {
        console.log('识别为格式5: localStorage格式')
        prompts = data.prompts.filter((p: any) => p && p.name && p.content)
        categories = data.categories || []
      }
      // 格式6: 尝试从任意结构中提取 prompts
      else if (data.data && typeof data.data === 'object') {
        console.log('尝试格式6: 从 data 中提取')
        // 尝试各种可能的结构
        const possiblePrompts = data.data.prompts || data.data.standalone || []
        const possiblePacks = data.data.packs || data.data.categories || {}
        
        if (Array.isArray(possiblePrompts)) {
          prompts = possiblePrompts.filter((p: any) => p && p.name && p.content)
        }
        
        if (typeof possiblePacks === 'object' && !Array.isArray(possiblePacks)) {
          for (const [packName, packPrompts] of Object.entries(possiblePacks)) {
            if (Array.isArray(packPrompts)) {
              prompts.push(...packPrompts.map((p: any) => ({ ...p, category: packName })))
              if (packName !== '未分类') {
                categories.push(packName)
              }
            }
          }
        }
      }
      else {
        console.error('无法识别的数据结构:', data)
        ElMessage.warning('无法识别的文件格式，请确保是旧版拆书库导出的备份文件')
        importBookAnalysisLegacyPreview.value = null
        return
      }

      console.log('提取到的提示词数量:', prompts.length)
      console.log('提取到的分类:', categories)

      if (prompts.length === 0) {
        ElMessage.warning('文件中没有有效的提示词数据')
        importBookAnalysisLegacyPreview.value = null
        return
      }

      // 构建预览数据
      importBookAnalysisLegacyPreview.value = {
        packCount: categories.length,
        totalPrompts: prompts.length,
        categories: categories.map((c: string) => c.startsWith('拆书-') ? c : `拆书-${c}`),
        prompts: prompts
      }
    } catch (error) {
      console.error('文件处理错误:', error)
      ElMessage.error('文件解析失败：' + (error instanceof Error ? error.message : '未知错误'))
      importBookAnalysisLegacyPreview.value = null
    }
  }
  reader.onerror = () => {
    ElMessage.error('文件读取失败')
    importBookAnalysisLegacyPreview.value = null
  }
  reader.readAsText(file.raw)
}

// 执行导入旧版完整备份
const executeImportBookAnalysisLegacy = async () => {
  if (!importBookAnalysisLegacyPreview.value) return

  const { prompts, categories } = importBookAnalysisLegacyPreview.value

  // 创建拆书分类
  for (const cat of categories) {
    if (cat !== '拆书-未分类' && !customCategories.value.includes(cat)) {
      customCategories.value.push(cat)
    }
  }
  saveCategoriesToStorage(customCategories.value)

  let successCount = 0
  let failCount = 0

  for (const prompt of prompts) {
    try {
      // 处理分类名：自动加"拆书-"前缀
      let category = prompt.category || '未分类'
      if (category === '未分类') {
        category = '拆书-未分类'
      } else if (!category.startsWith('拆书-')) {
        category = '拆书-' + category
      }

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
        version: prompt.version || '',
        subcategories: prompt.subcategories || []
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
    ElMessage.success(`成功导入旧版备份！共 ${successCount} 条提示词，已转换为拆书格式`)
  } else {
    ElMessage.warning(`导入完成：成功 ${successCount} 条，失败 ${failCount} 条`)
  }

  importBookAnalysisLegacyDialogVisible.value = false
}
</script>

<style scoped>
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
  background: linear-gradient(135deg, #4ade80 0%, #22c55e 100%) !important;
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
  background: linear-gradient(135deg, #0f766e 0%, #14b8a6 100%);
  border-color: transparent;
  box-shadow: 0 8px 18px rgba(20, 184, 166, 0.22);
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
  background: rgba(15, 23, 42, 0.45);
}

:root[data-theme='dark'] .pack-info-clickable:hover {
  background: rgba(96, 165, 250, 0.08);
}

:root[data-theme='dark'] .pack-info-card {
  background: rgba(30, 41, 59, 0.85);
  border-color: rgba(71, 85, 105, 0.35);
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
  color: #f1f5f9;
}

:root[data-theme='dark'] .pack-info-count {
  color: #94a3b8;
  background: transparent;
}

:root[data-theme='dark'] .pack-action-btn {
  background: transparent !important;
  color: #94a3b8 !important;
}

:root[data-theme='dark'] .pack-action-btn:hover {
  background: rgba(71, 85, 105, 0.55) !important;
  border-color: rgba(148, 163, 184, 0.35) !important;
  color: #e2e8f0 !important;
}

:root[data-theme='dark'] .pack-action-btn--danger:hover {
  background: rgba(220, 38, 38, 0.15) !important;
  border-color: rgba(248, 113, 113, 0.3) !important;
  color: #fca5a5 !important;
}

/* --- 暗色主题 - 视图切换按钮 --- */
:root[data-theme='dark'] .pack-view-toggle {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(71, 85, 105, 0.5);
}

:root[data-theme='dark'] .pack-view-btn {
  color: #64748b;
}

:root[data-theme='dark'] .pack-view-btn:hover {
  color: #94a3b8;
  background: rgba(255, 255, 255, 0.06);
}

:root[data-theme='dark'] .pack-view-btn.active {
  background: rgba(99, 102, 241, 0.2);
  color: #a5b4fc;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

/* --- 暗色主题 - 网格模式 --- */
:root[data-theme='dark'] .packs-wrapper.grid .pack-info-actions {
  border-top-color: rgba(71, 85, 105, 0.3);
}

:root[data-theme='dark'] .standalone-cards-area {
  background: linear-gradient(180deg, rgba(124, 45, 18, 0.2) 0%, #1e293b 100%);
  border-color: rgba(249, 115, 22, 0.4);
}

:root[data-theme='dark'] .standalone-cards-area.drag-over {
  border-color: rgba(249, 115, 22, 0.6);
  background: linear-gradient(180deg, rgba(124, 45, 18, 0.3) 0%, #1e293b 100%);
}

:root[data-theme='dark'] .standalone-card {
  background: rgba(30, 41, 59, 0.95);
  border-color: rgba(71, 85, 105, 0.4);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

:root[data-theme='dark'] .standalone-card:hover {
  border-color: rgba(249, 115, 22, 0.5);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .standalone-card-header {
  background: linear-gradient(135deg, rgba(124, 45, 18, 0.2) 0%, rgba(30, 41, 59, 0.9) 100%);
  border-bottom-color: rgba(71, 85, 105, 0.3);
}

:root[data-theme='dark'] .card-drag-handle {
  color: #fbbf24;
}

:root[data-theme='dark'] .standalone-card-name {
  color: #f3f4f6;
}

:root[data-theme='dark'] .standalone-card-preview {
  color: #9ca3af;
}

:root[data-theme='dark'] .standalone-card-footer {
  border-top-color: rgba(71, 85, 105, 0.3);
  background: rgba(30, 41, 59, 0.8);
}

:root[data-theme='dark'] .standalone-card-time {
  color: #6b7280;
}

:root[data-theme='dark'] .standalone-card-actions :deep(.el-button) {
  background: rgba(94, 234, 212, 0.1);
}

:root[data-theme='dark'] .standalone-card-actions :deep(.el-button.el-button--primary) {
  color: #5eead4;
  border-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .standalone-card-actions :deep(.el-button.el-button--primary:hover) {
  color: #fff;
  background: linear-gradient(135deg, #0f766e 0%, #14b8a6 100%);
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
  border-bottom-color: rgba(71, 85, 105, 0.35);
}

:root[data-theme='dark'] .header h2 {
  color: #f1f5f9;
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
  background: linear-gradient(135deg, #16a34a 0%, #15803d 100%) !important;
  box-shadow: 0 2px 8px rgba(22, 163, 74, 0.35);
}

:root[data-theme='dark'] .btn-create-prompt:hover {
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%) !important;
  box-shadow: 0 4px 16px rgba(22, 163, 74, 0.5);
}

:root[data-theme='dark'] .section-title {
  color: #f3f4f6;
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
  background-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .main-tabs :deep(.el-tabs__item) {
  color: #cbd5e1;
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
  color: #f3f4f6;
}

:root[data-theme='dark'] .left-panel {
  border-bottom-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .panel-header {
  border-bottom-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .export-select-all {
  background: rgba(51, 65, 85, 0.6);
}

:root[data-theme='dark'] .export-card-item {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .export-card-item:hover {
  border-color: rgba(249, 115, 22, 0.5);
}

:root[data-theme='dark'] .export-card-item.selected {
  border-color: #10b981;
  background: rgba(16, 185, 129, 0.15);
}

:root[data-theme='dark'] .export-card-name {
  color: #f3f4f6;
}

:root[data-theme='dark'] .export-card-preview {
  color: #9ca3af;
}

:root[data-theme='dark'] .form-label {
  color: #e5e7eb;
}

:root[data-theme='dark'] .empty-fields {
  background-color: rgba(51, 65, 85, 0.6);
}

:root[data-theme='dark'] .empty-hint {
  color: #9ca3af;
}

:root[data-theme='dark'] .field-item {
  background-color: rgba(51, 65, 85, 0.6);
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
  color: #f3f4f6;
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
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.2) 0%, rgba(30, 41, 59, 0.9) 100%) !important;
}

:root[data-theme='dark'] .standalone-card-lock {
  color: #fbbf24;
}

/* --- 暗色主题 - 牌桌预览卡片 --- */
:root[data-theme='dark'] .pack-fan-stage {
  background: rgba(2, 6, 23, 0.7);
}

:root[data-theme='dark'] .pack-fan-card {
  background: rgba(30, 41, 59, 0.95);
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
  border-bottom-color: rgba(71, 85, 105, 0.3);
}

:root[data-theme='dark'] .pack-fan-card-head-icon {
  color: #a5b4fc;
}

:root[data-theme='dark'] .pack-fan-card-head-name {
  color: #f1f5f9;
}

:root[data-theme='dark'] .pack-fan-card-head-lock {
  color: #fbbf24;
}

:root[data-theme='dark'] .pack-fan-card-content {
  color: #94a3b8;
}

:root[data-theme='dark'] .pack-fan-card-content.is-empty {
  color: #64748b;
}

:root[data-theme='dark'] .pack-fan-card-content-encrypted {
  color: #fbbf24;
}

:root[data-theme='dark'] .pack-fan-card-foot {
  border-top-color: rgba(71, 85, 105, 0.3);
}

:root[data-theme='dark'] .pack-fan-card-foot-time {
  color: #64748b;
}

:root[data-theme='dark'] .pack-fan-action-btn {
  border-color: rgba(71, 85, 105, 0.4) !important;
  background: rgba(30, 41, 59, 0.8) !important;
  color: #94a3b8 !important;
}

:root[data-theme='dark'] .pack-fan-action-btn:hover {
  background: rgba(51, 65, 85, 0.8) !important;
  border-color: rgba(100, 116, 139, 0.5) !important;
  color: #e2e8f0 !important;
}

:root[data-theme='dark'] .pack-fan-action-btn--danger:hover {
  background: rgba(127, 29, 29, 0.3) !important;
  border-color: rgba(239, 68, 68, 0.4) !important;
  color: #f87171 !important;
}

:root[data-theme='dark'] .move-prompt-radio-item {
  border-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .move-prompt-radio-item:hover {
  border-color: #818cf8;
  background: rgba(30, 41, 59, 0.5);
}

/* ---- 编辑弹窗左右布局 ---- */
.pack-edit-layout {
  display: flex;
  gap: 16px;
  height: 70vh;
}
.pack-edit-left {
  flex: 2;
  min-width: 0;
  overflow-y: auto;
  padding-right: 8px;
}
.pack-edit-right {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  border-left: 1px solid #e5e7eb;
  padding-left: 16px;
}
.pack-edit-right-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}
.pack-edit-right-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}
.pack-edit-right-tip {
  font-size: 11px;
  color: #94a3b8;
  margin-bottom: 8px;
}
.pack-edit-right-tip code {
  background: #eef2f7;
  color: #4f46e5;
  padding: 0 4px;
  border-radius: 3px;
}
.pack-edit-right-body {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}
.pack-edit-form :deep(.el-form-item) {
  margin-bottom: 12px;
}
.pack-edit-form :deep(.el-form-item__label) {
  font-size: 13px;
  color: #1f2937;
  font-weight: 600;
  padding-bottom: 3px;
  line-height: 1.4;
}
.pack-edit-content-input :deep(.el-textarea__inner) {
  padding: 10px 12px;
  font-size: 13px;
  line-height: 1.6;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
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
  color: #94a3b8;
  font-size: 13px;
  text-align: center;
  padding: 24px 0;
}
.pack-edit-fields {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.pack-edit-field {
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 10px;
}
.pack-edit-field-head {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
}
.pack-edit-field-name {
  flex: 1;
}
.pack-edit-field-body {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.pack-edit-field-order {
  display: flex;
  gap: 4px;
  margin-top: 6px;
}
.pack-edit-field-options {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 6px;
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

/* ========== 拆书导入导出对话框 ========== */
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
</style>
