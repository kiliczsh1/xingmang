﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿﻿<template>
  <div class="prompts-container" :style="promptSkinCssVars">
    <div class="header">
      <div class="header-main">
        <h2>提示词管理</h2>
        <div class="skin-toolbar">
          <span class="skin-toolbar-label">模板 / 样式</span>
          <el-select
            v-model="selectedPromptSkinId"
            size="small"
            class="skin-select"
            placeholder="选择皮肤"
          >
            <el-option-group label="内置皮肤">
              <el-option
                v-for="skin in builtInPromptSkins"
                :key="skin.id"
                :label="skin.name"
                :value="skin.id"
              />
            </el-option-group>
            <el-option-group v-if="customPromptSkins.length > 0" label="导入皮肤">
              <el-option
                v-for="skin in customPromptSkins"
                :key="skin.id"
                :label="skin.name"
                :value="skin.id"
              />
            </el-option-group>
          </el-select>
          <el-dropdown trigger="click" @command="handlePromptSkinCommand">
            <el-button size="small" class="skin-config-btn">
              皮肤配置
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="exportCurrent">
                  <el-icon><Download /></el-icon>
                  导出当前皮肤
                </el-dropdown-item>
                <el-dropdown-item command="exportAll">
                  <el-icon><FolderOpened /></el-icon>
                  导出全部皮肤
                </el-dropdown-item>
                <el-dropdown-item divided command="import">
                  <el-icon><Upload /></el-icon>
                  导入皮肤配置
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <input
            ref="promptSkinInputRef"
            type="file"
            accept=".json"
            class="hidden-input"
            @change="handlePromptSkinFileChange"
          />
        </div>
      </div>
      <div class="header-actions">
        <el-button-group class="size-selector">
          <el-button 
            :type="cardSize === 'small' ? 'primary' : ''" 
            size="small"
            @click="cardSize = 'small'"
          >
            小
          </el-button>
          <el-button 
            :type="cardSize === 'medium' ? 'primary' : ''" 
            size="small"
            @click="cardSize = 'medium'"
          >
            中
          </el-button>
          <el-button 
            :type="cardSize === 'large' ? 'primary' : ''" 
            size="small"
            @click="cardSize = 'large'"
          >
            大
          </el-button>
        </el-button-group>
        <el-dropdown trigger="click" @command="handleExportCommand">
          <el-button type="warning">
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
          <el-button type="info">
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
        <el-button type="success" @click="openCategoryDialog">
          <el-icon><FolderAdd /></el-icon>
          创建卡包
        </el-button>
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          创建提示词
        </el-button>
      </div>
    </div>

    <!-- 卡包区域 - 在上 -->
    <div class="section-title">
      <el-icon><FolderOpened /></el-icon>
      <span>卡包</span>
    </div>
    <div class="packs-wrapper">
      <div
        v-for="category in categoryList"
        :key="category.name"
        class="card-pack-vertical"
        :class="[cardSize, { 'drag-over': isDragging && draggedPrompt?.category !== category.name }]"
        @dragover.prevent="onDragOver"
        @drop.prevent="onDrop(category.name, $event)"
      >
        <!-- 卡片列表区域 -->
        <div 
          class="cards-scroll-area"
        >
          <div
            v-for="prompt in category.prompts"
            :key="prompt.id"
            class="pack-card"
            :class="[cardSize, { dragging: isDragging && draggedPrompt?.id === prompt.id }, { encrypted: prompt.card_type === 'encrypted' }]"
            draggable="true"
            @dragstart="onDragStart(prompt, $event)"
            @dragend="onDragEnd"
            @click="handleEdit(prompt)"
          >
            <div class="pack-card-header">
              <el-icon class="pack-card-drag-handle"><Rank /></el-icon>
              <span class="pack-card-name">{{ prompt.name }}</span>
              <el-icon v-if="prompt.card_type === 'encrypted'" class="pack-card-lock"><Lock /></el-icon>
            </div>
            <div class="pack-card-content">
              <el-tag size="small" :type="getTagType(prompt.category)" class="pack-card-tag">
                {{ prompt.category }}
              </el-tag>
              <div class="pack-card-preview">
                <template v-if="prompt.card_type === 'encrypted'">
                  <el-icon><Lock /></el-icon> 内容已加密
                </template>
                <template v-else>
                  {{ prompt.content.slice(0, 40) }}{{ prompt.content.length > 40 ? '...' : '' }}
                </template>
              </div>
            </div>
            <div class="pack-card-footer">
              <span class="pack-card-time">{{ formatDate(prompt.created_at) }}</span>
              <div class="pack-card-actions" @click.stop>
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
          
          <!-- 空卡包提示 -->
          <div v-if="category.prompts.length === 0" class="empty-pack-hint" @click="handleCreateInCategory(category.name)">
            <el-icon><Plus /></el-icon>
            <span>点击添加卡片</span>
          </div>
        </div>

        <!-- 卡包底部操作 -->
        <div class="pack-footer-vertical">
          <div class="pack-footer-info">
            <el-icon class="pack-footer-icon"><FolderOpened /></el-icon>
            <span class="pack-footer-name">{{ category.name }}</span>
            <span class="pack-footer-count">{{ category.prompts.length }}</span>
          </div>
          <div class="pack-footer-actions">
            <el-button
              type="primary"
              link
              size="small"
              @click="handleCreateInCategory(category.name)"
            >
              <el-icon><Plus /></el-icon>
              添加
            </el-button>
            <el-tooltip content="预览卡包内容" placement="top" :show-after="150">
              <el-button
                v-if="category.prompts.length > 0"
                type="primary"
                size="small"
                circle
                @click="openPackPreviewDialog(category.name)"
              >
                <el-icon><View /></el-icon>
              </el-button>
            </el-tooltip>
            <el-tooltip content="批量删除" placement="top" :show-after="150">
              <el-button
                v-if="category.prompts.length > 0 && !DEFAULT_CATEGORIES.includes(category.name)"
                type="danger"
                size="small"
                circle
                @click="openBatchDeleteDialog(category.name)"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </el-tooltip>
            <el-button
              v-if="!DEFAULT_CATEGORIES.includes(category.name)"
              type="danger"
              link
              size="small"
              @click="handleDeleteCategory(category.name)"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 独立卡片区域 - 在下 -->
    <div class="section-title" style="margin-top: 32px;">
      <el-icon><Document /></el-icon>
      <span>独立卡片</span>
      <span class="section-count">{{ uncategorizedPrompts.length }}</span>
    </div>
    <div 
      v-if="uncategorizedPrompts.length > 0"
      class="standalone-cards-area"
      :class="[cardSize, { 'drag-over': isDragging }]"
      @dragover.prevent="onDragOver"
      @drop.prevent="onDrop('未分类', $event)"
    >
      <div class="standalone-cards-grid">
        <div
          v-for="prompt in uncategorizedPrompts"
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
            <el-tag size="small" :type="getTagType(prompt.category)" class="standalone-card-tag">
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
    <div v-else class="empty-standalone-hint" 
      :class="{ 'drag-over': isDragging }"
      @click="handleCreate"
      @dragover.prevent="onDragOver"
      @drop.prevent="onDrop('未分类', $event)"
    >
      <el-icon><Plus /></el-icon>
      <span>暂无独立卡片，点击创建或拖拽卡片到此处</span>
    </div>

    <!-- 创建/编辑卡包对话框 -->
    <el-dialog
      v-model="categoryDialogVisible"
      title="创建卡包"
      width="400px"
      :close-on-click-modal="false"
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

    <!-- 导出独立卡片弹窗 -->
    <el-dialog
      v-model="exportStandaloneDialogVisible"
      title="导出独立卡片"
      width="800px"
      :close-on-click-modal="false"
      class="export-dialog"
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
            全选 ({{ selectedStandaloneCards.length }}/{{ uncategorizedPrompts.length }})
          </el-checkbox>
        </div>
        <div v-if="uncategorizedPrompts.length > 0" class="export-cards-grid">
          <div
            v-for="prompt in uncategorizedPrompts"
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
              <div class="export-card-name">{{ prompt.name }}</div>
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

    <!-- 导入并转换txt/md弹窗 -->
    <el-dialog
      v-model="importConvertDialogVisible"
      title="导入并转换 txt/md 文件"
      width="800px"
      :close-on-click-modal="false"
      class="import-dialog"
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
      :width="dialogWidthMap[dialogSize]"
      top="6vh"
      append-to-body
      :close-on-click-modal="false"
      class="prompt-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <span>{{ isEdit ? '编辑提示词' : '创建提示词' }}</span>
          <div class="size-controls">
            <el-button-group>
              <el-button 
                :type="dialogSize === 'small' ? 'primary' : ''" 
                size="small"
                @click="dialogSize = 'small'"
              >
                小
              </el-button>
              <el-button 
                :type="dialogSize === 'medium' ? 'primary' : ''" 
                size="small"
                @click="dialogSize = 'medium'"
              >
                中
              </el-button>
              <el-button 
                :type="dialogSize === 'large' ? 'primary' : ''" 
                size="small"
                @click="dialogSize = 'large'"
              >
                大
              </el-button>
            </el-button-group>
          </div>
        </div>
      </template>
      <div class="dialog-content" :style="{ height: dialogHeightMap[dialogSize] }">
        <!-- 上方：基本信息和提示词内容 -->
        <div class="left-panel">
          <el-form :model="formData" label-width="80px">
            <el-form-item label="名称" required>
              <el-input v-model="formData.name" placeholder="请输入名称" />
            </el-form-item>
            <el-form-item label="分类名称" required>
              <el-select
                v-model="formData.category"
                placeholder="请选择卡包"
                class="field-input full-width-input"
              >
                <el-option
                  v-for="category in categoryOptions"
                  :key="category"
                  :label="category"
                  :value="category"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="卡片类型">
              <div class="card-type-selector">
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
                <div v-if="formData.card_type === 'encrypted'" class="card-type-actions">
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
                <div v-if="formData.card_type === 'encrypted' && formData.password" class="card-type-status">
                  <el-tag type="warning" size="small">
                    <el-icon><Lock /></el-icon>
                    已加密
                  </el-tag>
                </div>
                <div v-if="formData.card_type === 'encrypted' && !formData.password" class="card-type-status">
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
            <el-form-item required>
              <template #label>
                <div class="form-item-label-with-guide">
                  <span>内容</span>
                  <el-popover
                    placement="right-start"
                    :width="360"
                    trigger="click"
                    popper-class="prompt-guide-popper"
                  >
                    <template #reference>
                      <el-button size="small" text class="prompt-guide-trigger">
                        <el-icon><InfoFilled /></el-icon>
                        模板说明
                      </el-button>
                    </template>
                    <div class="prompt-guide-popover">
                      <div class="prompt-guide-title">把说明写进模板，把变量留给字段</div>
                      <p class="prompt-guide-intro">
                        正文里直接写完整提示词，把需要用户补充的内容写成 <code>${字段名称}</code>。
                      </p>
                      <div class="prompt-guide-example">
                        <span class="prompt-guide-example-label">示例</span>
                        <p>帮我生成 <code>${数量}</code> 个 <code>${小说类型}</code> 类型、适合 <code>${平台}</code> 平台的小说书名。</p>
                      </div>
                      <div class="prompt-guide-tips">
                        <div class="prompt-guide-tip">
                          <span class="prompt-guide-tip-index">1</span>
                          <span>先按正常语气写完整句子，再把可变内容替换成字段。</span>
                        </div>
                        <div class="prompt-guide-tip">
                          <span class="prompt-guide-tip-index">2</span>
                          <span>长文本内容建议单独成段，避免和短字段混在一句里。</span>
                        </div>
                        <div class="prompt-guide-tip">
                          <span class="prompt-guide-tip-index">3</span>
                          <span>字段名称尽量直接，比如“数量”“平台”“风格要求”。</span>
                        </div>
                      </div>
                    </div>
                  </el-popover>
                </div>
              </template>
              <el-input
                v-model="formData.content"
                type="textarea"
                :rows="15"
                placeholder="请输入提示词内容（作为AI的system层指令）"
              />
            </el-form-item>
            <div class="panel-header prompt-panel-header">
              <label class="form-label">字段配置</label>
              <el-button type="primary" size="small" @click="addField">
                <el-icon><Plus /></el-icon>
                添加字段
              </el-button>
            </div>
            <div class="field-config-section">
              <el-alert
                title="字段配置说明"
                type="info"
                :closable="false"
                show-icon
              >
                <template #default>
                  <div class="format-description">
                    <p><strong>⚙️ 配置字段信息</strong></p>
                    <p>为每个 <code>${字段名称}</code> 设置详细信息，决定用户看到什么样的表单</p>
                  </div>
                </template>
              </el-alert>
              
              <div v-if="fieldsConfig.length === 0" class="empty-fields">
                <p class="empty-hint">暂无字段配置，点击右上角"添加字段"开始配置</p>
              </div>
              
              <div v-else class="fields-list">
                <div v-for="(field, index) in fieldsConfig" :key="index" class="field-item">
                  <div class="field-header">
                    <div class="field-name-container">
                      <el-tag 
                        :type="field.required ? 'danger' : 'info'" 
                        size="small"
                        class="required-tag"
                      >
                        {{ field.required ? '必选' : '选填' }}
                      </el-tag>
                      <el-input
                        v-model="field.name"
                        class="field-name-input"
                        placeholder="字段名称"
                        @change="updateFieldName(index, field.name)"
                      />
                      <el-button
                        type="primary"
                        link
                        size="small"
                        @click="copyFieldName(field.name)"
                        title="复制占位符"
                      >
                        <el-icon><CopyDocument /></el-icon>
                      </el-button>
                    </div>
                    <div class="field-actions">
                      <el-switch
                        v-model="field.required"
                        active-text="必选"
                        inactive-text="选填"
                        :active-action-icon="Check"
                        :inactive-action-icon="Close"
                        inline-prompt
                        style="--el-switch-on-color: #f56c6c; --el-switch-off-color: #909399;"
                      />
                      <el-button 
                        type="primary" 
                        link 
                        size="small" 
                        @click="moveField(index, index - 1)"
                        :disabled="index === 0"
                        title="上移"
                      >
                        <el-icon><ArrowUp /></el-icon>
                      </el-button>
                      <el-button 
                        type="primary" 
                        link 
                        size="small" 
                        @click="moveField(index, index + 1)"
                        :disabled="index === fieldsConfig.length - 1"
                        title="下移"
                      >
                        <el-icon><ArrowDown /></el-icon>
                      </el-button>
                      <el-button type="danger" link size="small" @click="removeField(index)" title="删除">
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </div>
                  </div>
                  <div class="field-config-form">
                    <el-input
                      v-model="field.label"
                      placeholder="字段显示名称"
                      class="field-input"
                    />
                    <el-select
                      v-model="field.type"
                      placeholder="字段类型"
                      class="field-input"
                    >
                      <el-option label="单行文本" value="text" />
                      <el-option label="多行文本" value="textarea" />
                      <el-option label="下拉选择" value="select" />
                    </el-select>
                    <div v-if="field.type === 'select'" class="field-options">
                      <label>选项（每行一个）：</label>
                      <el-input
                        v-model="field.optionsText"
                        type="textarea"
                        :rows="3"
                        placeholder="请输入选项，每行一个"
                        @change="updateFieldOptions(index)"
                        class="field-input"
                      />
                    </div>
                    <el-input
                      v-model="field.description"
                      type="textarea"
                      :rows="2"
                      placeholder="字段说明"
                      class="field-input"
                    />
                  </div>
                </div>
              </div>
            </div>
            <el-form-item label="标签" class="prompt-tags-form-item">
              <div class="subcategory-form">
                <div class="subcategory-tags-row">
                  <el-tag
                    v-for="(subcat, index) in formSubcategories"
                    :key="index"
                    closable
                    size="small"
                    @close="removeFormSubcategory(index)"
                    class="subcategory-tag"
                  >
                    {{ subcat }}
                  </el-tag>
                </div>
                <div class="subcategory-input-row">
                  <el-input
                    v-model="newSubcategory"
                    placeholder="输入标签后按回车添加"
                    size="small"
                    class="subcategory-input"
                    @keyup.enter="addFormSubcategory"
                  />
                  <el-button size="small" type="primary" @click="addFormSubcategory">
                    <el-icon><Plus /></el-icon>
                  </el-button>
                  <el-button size="small" type="success" @click="saveSubcategoriesOnly">
                    保存
                  </el-button>
                </div>
              </div>
            </el-form-item>
          </el-form>
        </div>
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

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
              <span class="pack-fan-card-inner">
                <span class="pack-fan-card-face pack-fan-card-back">
                  <span class="pack-fan-card-back-core">
                    <span class="pack-fan-card-back-title">{{ packPreviewCategory }}</span>
                    <span class="pack-fan-card-back-mark">Prompt Deck</span>
                  </span>
                </span>
                <span class="pack-fan-card-face pack-fan-card-front">
                  <span class="pack-fan-card-front-header">
                    <el-icon class="pack-fan-card-front-handle"><Rank /></el-icon>
                    <span class="pack-fan-card-front-name">{{ prompt.name }}</span>
                    <el-icon v-if="prompt.card_type === 'encrypted'" class="pack-fan-card-front-lock"><Lock /></el-icon>
                  </span>
                  <span class="pack-fan-card-front-content">
                    <el-tag size="small" :type="getTagType(prompt.category)" class="pack-fan-card-front-tag">{{ prompt.category }}</el-tag>
                    <span class="pack-fan-card-front-preview">
                      <template v-if="prompt.card_type === 'encrypted'">
                        <el-icon><Lock /></el-icon> 内容已加密
                      </template>
                      <template v-else>
                        {{ prompt.content.slice(0, 60) }}{{ prompt.content.length > 60 ? '...' : '' }}
                      </template>
                    </span>
                  </span>
                  <span class="pack-fan-card-front-footer">
                    <span class="pack-fan-card-front-time">{{ formatDate(prompt.created_at) }}</span>
                  </span>
                </span>
              </span>
            </button>
          </div>
          <el-empty v-else description="该卡包暂无可预览的卡片" />
        </div>
      </div>
    </transition>

    <!-- 预览弹窗 -->
    <el-dialog
      v-model="previewDialogVisible"
      title="生成结果预览"
      width="50vw"
      destroy-on-close
      class="preview-dialog"
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, ArrowUp, ArrowDown, Delete, CopyDocument, FolderAdd, FolderOpened, Edit, Document, Rank, Check, Close, Download, Upload, UploadFilled, View, Switch, InfoFilled, Lock } from '@element-plus/icons-vue'
import { promptAPI } from '@/api'
import type { Prompt } from '@/types'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import SplitRichTextEditor from '@/components/SplitRichTextEditor.vue'

interface PromptSkinPreset {
  id: string
  name: string
  description: string
  builtIn?: boolean
  tokens?: Record<string, string>
}

const PROMPT_SKIN_STORAGE_KEY = 'prompt_pack_custom_skins'
const PROMPT_SKIN_SELECTED_KEY = 'prompt_pack_selected_skin'

const promptSkinDefaultTokens: Record<string, string> = {
  'prompt-pack-shell-top': '#edf5f4',
  'prompt-pack-shell-mid': '#e5efee',
  'prompt-pack-shell-bottom': '#dbe8e7',
  'prompt-pack-shell-radius': '20px',
  'prompt-pack-shell-border': 'rgba(74, 126, 123, 0.28)',
  'prompt-pack-shell-shadow': '0 14px 30px rgba(25, 70, 68, 0.12), inset 0 1px 0 rgba(255, 255, 255, 0.38)',
  'prompt-pack-shell-hover-border': 'rgba(40, 111, 108, 0.44)',
  'prompt-pack-shell-hover-shadow': '0 20px 40px rgba(18, 84, 81, 0.16), inset 0 1px 0 rgba(255, 255, 255, 0.46)',
  'prompt-pack-shell-overlay': 'linear-gradient(180deg, rgba(255, 255, 255, 0.22), transparent 34%), linear-gradient(120deg, transparent 0%, rgba(255, 255, 255, 0.07) 48%, transparent 70%)',
  'prompt-pack-drag-bg': 'radial-gradient(circle at top right, rgba(98, 176, 170, 0.22), transparent 38%), radial-gradient(circle at left top, rgba(140, 201, 193, 0.16), transparent 44%), linear-gradient(180deg, #e7f1f0 0%, #ddeceb 100%)',
  'prompt-pack-drag-border': '#3d8e89',
  'prompt-pack-drag-shadow': '0 22px 44px rgba(22, 96, 93, 0.18), inset 0 0 0 1px rgba(123, 177, 171, 0.18)',
  'prompt-pack-header-top': '#c3ddd8',
  'prompt-pack-header-mid': '#afd2cd',
  'prompt-pack-header-bottom': '#98c2bc',
  'prompt-pack-header-border': 'rgba(31, 89, 86, 0.18)',
  'prompt-pack-header-text': '#154d4b',
  'prompt-pack-muted-text': '#607d79',
  'prompt-pack-badge-bg': 'rgba(239, 246, 245, 0.94)',
  'prompt-pack-badge-border': 'rgba(83, 131, 127, 0.18)',
  'prompt-pack-badge-text': '#276b67',
  'prompt-pack-header-align': 'flex-start',
  'prompt-pack-header-padding': '14px 12px 12px 12px',
  'prompt-pack-header-overlay': 'radial-gradient(circle at top, rgba(255, 255, 255, 0.16), transparent 42%), linear-gradient(180deg, rgba(255, 255, 255, 0.10), transparent 64%), linear-gradient(115deg, transparent 0%, rgba(255, 255, 255, 0.09) 48%, transparent 72%)',
  'prompt-pack-header-left-align': 'flex-start',
  'prompt-pack-header-left-gap': '6px',
  'prompt-pack-header-left-padding-top': '12px',
  'prompt-pack-header-actions-gap': '8px',
  'prompt-pack-header-actions-align': 'flex-start',
  'prompt-pack-icon-color': 'var(--pack-header-text)',
  'prompt-pack-icon-shadow': 'drop-shadow(0 4px 10px rgba(26, 88, 85, 0.14))',
  'prompt-pack-name-color': 'var(--pack-header-text)',
  'prompt-pack-name-align': 'left',
  'prompt-pack-name-shadow': '0 1px 6px rgba(255, 255, 255, 0.16)',
  'prompt-pack-count-shadow': '0 6px 12px rgba(22, 77, 75, 0.08)',
  'prompt-pack-action-bg': 'rgba(244, 249, 248, 0.58)',
  'prompt-pack-action-border': 'rgba(210, 228, 225, 0.88)',
  'prompt-pack-action-shadow': '0 8px 18px rgba(19, 91, 87, 0.10)',
  'prompt-pack-action-hover-bg': 'rgba(249, 252, 251, 0.96)',
  'prompt-pack-action-hover-border': 'rgba(227, 238, 236, 0.96)',
  'prompt-pack-action-primary': '#1f6f69',
  'prompt-pack-action-primary-hover': '#145b56',
  'prompt-pack-action-danger': '#b45366',
  'prompt-pack-action-danger-hover': '#9f324f',
  'prompt-pack-scroll-bg': 'linear-gradient(180deg, rgba(227, 238, 236, 0.82) 0%, rgba(239, 245, 244, 0.96) 100%)',
  'prompt-pack-card-bg': 'linear-gradient(180deg, rgba(246, 250, 249, 0.98) 0%, #eaf2f1 100%)',
  'prompt-pack-card-radius': '14px',
  'prompt-pack-card-border': '1px solid rgba(39, 107, 103, 0.14)',
  'prompt-pack-card-shadow': '0 8px 18px rgba(21, 77, 75, 0.08), inset 0 1px 0 rgba(255, 255, 255, 0.72)',
  'prompt-pack-card-hover-border': 'rgba(24, 99, 95, 0.28)',
  'prompt-pack-card-hover-shadow': '0 14px 26px rgba(18, 84, 81, 0.12), inset 0 1px 0 rgba(255, 255, 255, 0.82)',
  'prompt-pack-card-drag-opacity': '0.56',
  'prompt-pack-card-drag-scale': '0.97',
  'prompt-pack-card-header-bg': 'linear-gradient(135deg, rgba(133, 184, 178, 0.2) 0%, rgba(245, 249, 248, 0.94) 100%)',
  'prompt-pack-card-header-border': 'rgba(39, 107, 103, 0.10)',
  'prompt-pack-card-handle-color': '#0d9488',
  'prompt-pack-card-name-color': '#134e4a',
  'prompt-pack-card-content-gap': '6px',
  'prompt-pack-card-preview-color': '#5f7f7b',
  'prompt-pack-card-footer-bg': 'linear-gradient(180deg, rgba(241, 247, 246, 0.96) 0%, rgba(233, 241, 240, 0.98) 100%)',
  'prompt-pack-card-footer-border': 'rgba(39, 107, 103, 0.10)',
  'prompt-pack-card-time-color': 'var(--pack-muted-text)',
  'prompt-pack-empty-border': '1px dashed rgba(94, 234, 212, 0.32)',
  'prompt-pack-empty-radius': '14px',
  'prompt-pack-empty-bg': 'linear-gradient(180deg, rgba(236, 254, 255, 0.88) 0%, rgba(255, 255, 255, 0.96) 100%)',
  'prompt-pack-empty-text': '#7cb9b3',
  'prompt-pack-empty-hover-border': 'rgba(45, 212, 191, 0.4)',
  'prompt-pack-empty-hover-bg': 'linear-gradient(180deg, #f0fdfa 0%, #ffffff 100%)',
  'prompt-pack-empty-hover-text': '#14b8a6',
  'prompt-pack-footer-bg': 'linear-gradient(180deg, rgba(246, 255, 255, 0.98) 0%, rgba(239, 250, 250, 1) 100%)',
  'prompt-pack-footer-border': 'rgba(94, 234, 212, 0.2)',
  'prompt-pack-footer-button-bg': 'rgba(255, 255, 255, 0.62)',
  'prompt-pack-footer-button-shadow': '0 8px 18px rgba(56, 189, 248, 0.08)',
  'prompt-pack-footer-primary-color': '#22c7b8',
  'prompt-pack-footer-primary-border': 'rgba(94, 234, 212, 0.34)',
  'prompt-pack-footer-primary-hover-bg': 'rgba(236, 254, 255, 0.92)',
  'prompt-pack-footer-primary-hover-border': 'rgba(94, 234, 212, 0.5)',
  'prompt-pack-footer-primary-hover-color': '#0f766e',
  'prompt-pack-footer-danger-color': '#fb8fa3',
  'prompt-pack-footer-danger-border': 'rgba(251, 113, 133, 0.2)',
  'prompt-pack-footer-danger-hover-bg': 'rgba(255, 241, 242, 0.95)',
  'prompt-pack-footer-danger-hover-border': 'rgba(251, 113, 133, 0.34)',
  'prompt-pack-footer-danger-hover-color': '#f43f5e'
}

const promptSkinTokenKeys = Object.keys(promptSkinDefaultTokens)

const builtInPromptSkins: PromptSkinPreset[] = [
  {
    id: 'modern-glass',
    name: '新拟态卡包',
    description: '当前使用的青瓷玻璃感卡包样式。',
    builtIn: true
  },
  {
    id: 'classic-green',
    name: '经典绿色',
    description: '恢复早期绿色卡包的旧版视觉。',
    builtIn: true,
    tokens: {
      'prompt-pack-shell-top': '#f6ffed',
      'prompt-pack-shell-mid': '#ffffff',
      'prompt-pack-shell-bottom': '#ffffff',
      'prompt-pack-shell-radius': '12px',
      'prompt-pack-shell-border': '#b7eb8f',
      'prompt-pack-shell-shadow': '0 4px 12px rgba(82, 196, 26, 0.15)',
      'prompt-pack-shell-hover-border': '#95de64',
      'prompt-pack-shell-hover-shadow': '0 8px 20px rgba(82, 196, 26, 0.18)',
      'prompt-pack-shell-overlay': 'linear-gradient(180deg, rgba(255, 255, 255, 0.15), transparent 70%)',
      'prompt-pack-drag-bg': 'linear-gradient(180deg, #d9f7be 0%, #ffffff 100%)',
      'prompt-pack-drag-border': '#52c41a',
      'prompt-pack-drag-shadow': '0 10px 24px rgba(82, 196, 26, 0.18)',
      'prompt-pack-header-top': '#95de64',
      'prompt-pack-header-mid': '#52c41a',
      'prompt-pack-header-bottom': '#389e0d',
      'prompt-pack-header-border': '#b7eb8f',
      'prompt-pack-header-text': '#ffffff',
      'prompt-pack-muted-text': '#bfbfbf',
      'prompt-pack-badge-bg': '#ffffff',
      'prompt-pack-badge-border': 'transparent',
      'prompt-pack-badge-text': '#52c41a',
      'prompt-pack-header-align': 'center',
      'prompt-pack-header-padding': '12px 8px',
      'prompt-pack-header-overlay': 'linear-gradient(180deg, rgba(255, 255, 255, 0.15), transparent 70%)',
      'prompt-pack-header-left-align': 'center',
      'prompt-pack-header-left-gap': '4px',
      'prompt-pack-header-left-padding-top': '0px',
      'prompt-pack-header-actions-gap': '6px',
      'prompt-pack-header-actions-align': 'center',
      'prompt-pack-icon-color': '#ffffff',
      'prompt-pack-icon-shadow': 'none',
      'prompt-pack-name-color': '#ffffff',
      'prompt-pack-name-align': 'center',
      'prompt-pack-name-shadow': 'none',
      'prompt-pack-count-shadow': 'none',
      'prompt-pack-action-bg': 'rgba(255, 255, 255, 0.9)',
      'prompt-pack-action-border': 'transparent',
      'prompt-pack-action-shadow': 'none',
      'prompt-pack-action-hover-bg': '#ffffff',
      'prompt-pack-action-hover-border': 'transparent',
      'prompt-pack-action-primary': '#409eff',
      'prompt-pack-action-primary-hover': '#66b1ff',
      'prompt-pack-action-danger': '#ff4d4f',
      'prompt-pack-action-danger-hover': '#ff7875',
      'prompt-pack-scroll-bg': 'linear-gradient(180deg, rgba(246, 255, 237, 0.85) 0%, rgba(255, 255, 255, 0.95) 100%)',
      'prompt-pack-card-bg': '#ffffff',
      'prompt-pack-card-radius': '8px',
      'prompt-pack-card-border': '1px solid #e8e8e8',
      'prompt-pack-card-shadow': '0 1px 4px rgba(0, 0, 0, 0.04)',
      'prompt-pack-card-hover-border': '#52c41a',
      'prompt-pack-card-hover-shadow': '0 2px 8px rgba(82, 196, 26, 0.15)',
      'prompt-pack-card-drag-opacity': '0.5',
      'prompt-pack-card-drag-scale': '0.95',
      'prompt-pack-card-header-bg': 'linear-gradient(135deg, #f6ffed 0%, #ffffff 100%)',
      'prompt-pack-card-header-border': '#f0f0f0',
      'prompt-pack-card-handle-color': '#52c41a',
      'prompt-pack-card-name-color': '#262626',
      'prompt-pack-card-content-gap': '4px',
      'prompt-pack-card-preview-color': '#8c8c8c',
      'prompt-pack-card-footer-bg': '#fafafa',
      'prompt-pack-card-footer-border': '#f0f0f0',
      'prompt-pack-card-time-color': '#bfbfbf',
      'prompt-pack-empty-border': '2px dashed #e8e8e8',
      'prompt-pack-empty-radius': '8px',
      'prompt-pack-empty-bg': 'transparent',
      'prompt-pack-empty-text': '#bfbfbf',
      'prompt-pack-empty-hover-border': '#52c41a',
      'prompt-pack-empty-hover-bg': '#f6ffed',
      'prompt-pack-empty-hover-text': '#52c41a',
      'prompt-pack-footer-bg': '#fafafa',
      'prompt-pack-footer-border': '#e8e8e8',
      'prompt-pack-footer-button-bg': 'transparent',
      'prompt-pack-footer-button-shadow': 'none',
      'prompt-pack-footer-primary-color': '#409eff',
      'prompt-pack-footer-primary-border': 'transparent',
      'prompt-pack-footer-primary-hover-bg': 'rgba(64, 158, 255, 0.08)',
      'prompt-pack-footer-primary-hover-border': 'transparent',
      'prompt-pack-footer-primary-hover-color': '#409eff',
      'prompt-pack-footer-danger-color': '#ff4d4f',
      'prompt-pack-footer-danger-border': 'transparent',
      'prompt-pack-footer-danger-hover-bg': 'rgba(255, 77, 79, 0.08)',
      'prompt-pack-footer-danger-hover-border': 'transparent',
      'prompt-pack-footer-danger-hover-color': '#ff4d4f'
    }
  }
]

const slugifyPromptSkinName = (value: string) =>
  value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9\u4e00-\u9fa5]+/g, '-')
    .replace(/^-+|-+$/g, '') || 'custom-skin'

const createUniquePromptSkinId = (baseId: string, existingIds: Set<string>) => {
  let nextId = baseId
  let suffix = 2

  while (existingIds.has(nextId)) {
    nextId = `${baseId}-${suffix}`
    suffix += 1
  }

  return nextId
}

const normalizePromptSkin = (raw: any, existingIds: Set<string>): PromptSkinPreset | null => {
  if (!raw || typeof raw !== 'object') return null

  const name = String(raw.name || raw.label || '').trim()
  if (!name) return null

  const sourceTokens = raw.tokens && typeof raw.tokens === 'object'
    ? raw.tokens
    : raw.styles && typeof raw.styles === 'object'
      ? raw.styles
      : raw.variables && typeof raw.variables === 'object'
        ? raw.variables
        : {}

  const tokens: Record<string, string> = {}
  promptSkinTokenKeys.forEach((key) => {
    const value = sourceTokens[key]
    if (typeof value === 'string' && value.trim()) {
      tokens[key] = value.trim()
    }
  })

  const baseId = slugifyPromptSkinName(String(raw.id || name))
  const id = createUniquePromptSkinId(baseId, existingIds)
  existingIds.add(id)

  return {
    id,
    name,
    description: String(raw.description || '').trim() || '导入的模板/样式皮肤配置。',
    tokens
  }
}

const loadPromptCustomSkinsFromStorage = (): PromptSkinPreset[] => {
  const raw = localStorage.getItem(PROMPT_SKIN_STORAGE_KEY)
  if (!raw) return []

  try {
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []

    const existingIds = new Set(builtInPromptSkins.map((skin) => skin.id))
    return parsed
      .map((item) => normalizePromptSkin(item, existingIds))
      .filter((skin): skin is PromptSkinPreset => Boolean(skin))
  } catch {
    return []
  }
}

const savePromptCustomSkinsToStorage = (skins: PromptSkinPreset[]) => {
  localStorage.setItem(PROMPT_SKIN_STORAGE_KEY, JSON.stringify(skins))
}

const loadSelectedPromptSkinFromStorage = () => {
  return localStorage.getItem(PROMPT_SKIN_SELECTED_KEY) || builtInPromptSkins[0].id
}

const downloadPromptSkinFile = (filename: string, payload: unknown) => {
  const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

const dialogVisible = ref(false)
const isEdit = ref(false)
const prompts = ref<Prompt[]>([])
const formData = ref({
  id: 0,
  name: '',
  description: '',
  content: '',
  category: '默认',
  order_num: 0,
  card_type: 'normal' as 'normal' | 'encrypted',
  password: null as string | null
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
    await promptAPI.update(formData.value.id, {
      name: formData.value.name,
      description: formData.value.description,
      content: formData.value.content,
      category: formData.value.category,
      order_num: formData.value.order_num,
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

// 批量删除相关状态
const batchDeleteDialogVisible = ref(false)
const currentCategory = ref<string>('')
const selectedBatchDeleteCards = ref<number[]>([])
const selectAllBatchDelete = ref(false)
const packPreviewVisible = ref(false)
const packPreviewCategory = ref('')

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
const promptSkinInputRef = ref<HTMLInputElement | null>(null)
const customPromptSkins = ref<PromptSkinPreset[]>(loadPromptCustomSkinsFromStorage())
const selectedPromptSkinId = ref(loadSelectedPromptSkinFromStorage())

const allPromptSkins = computed(() => [...builtInPromptSkins, ...customPromptSkins.value])
const selectedPromptSkin = computed(() => {
  return allPromptSkins.value.find((skin) => skin.id === selectedPromptSkinId.value) ?? builtInPromptSkins[0]
})
const promptSkinCssVars = computed(() => {
  const mergedTokens = {
    ...promptSkinDefaultTokens,
    ...(selectedPromptSkin.value?.tokens || {})
  }

  return Object.fromEntries(
    Object.entries(mergedTokens).map(([key, value]) => [`--${key}`, value])
  )
})

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
const DEFAULT_CATEGORIES = ['默认', '写作要求', '写作风格']

const customCategories = ref<string[]>(loadCategoriesFromStorage())

// 拖拽相关状态
const draggedPrompt = ref<Prompt | null>(null)
const isDragging = ref(false)

// 未分类的提示词（不在任何自定义卡包中）
const uncategorizedPrompts = computed(() => {
  return prompts.value.filter(prompt => {
    const category = prompt.category || '默认'
    return category === '未分类' || (!customCategories.value.includes(category) && category !== '默认')
  })
})

// 按分类分组的提示词列表
const categoryList = computed(() => {
  const groups: Record<string, Prompt[]> = {}
  
  // 先确保所有默认卡包都存在
  DEFAULT_CATEGORIES.forEach(cat => {
    if (!groups[cat]) {
      groups[cat] = []
    }
  })
  
  // 再确保所有自定义卡包都存在
  customCategories.value.forEach(cat => {
    if (!groups[cat]) {
      groups[cat] = []
    }
  })
  
  // 添加提示词到对应分组（排除未分类的提示词）
  prompts.value.forEach(prompt => {
    const category = prompt.category || '默认'
    // 跳过未分类的提示词
    if (category === '未分类') return
    
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

watch(
  customPromptSkins,
  (skins) => {
    savePromptCustomSkinsToStorage(skins)
  },
  { deep: true }
)

watch(
  [allPromptSkins, selectedPromptSkinId],
  ([skins, selectedId]) => {
    const exists = skins.some((skin) => skin.id === selectedId)
    if (!exists) {
      selectedPromptSkinId.value = builtInPromptSkins[0].id
      return
    }

    localStorage.setItem(PROMPT_SKIN_SELECTED_KEY, selectedId)
  },
  { immediate: true }
)

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
    password: null
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
  optionsText: string;
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
})

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
  formData.value = {
    id: 0,
    name: '',
    description: '',
    content: '',
    category: categoryOptions.value[0] || '默认',
    order_num: 0,
    card_type: 'normal',
    password: null
  }
  // 重置字段配置
  fieldsConfig.value = []
  // 重置小分类
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
    password: prompt.password || null
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
      optionsText: (field.options || []).join('\n'),
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
      passwordDialogVisible.value = false
      pendingEditPrompt.value = null
      if (action === 'edit') {
        openEditDialog(promptToProcess)
      } else if (action === 'delete') {
        executeDelete(promptToProcess)
      } else if (action === 'preview') {
        previewContent.value = promptToProcess.content
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
  ElMessage.success('已移除密码，卡片已切换为普通卡片')
}

const handleCardTypeChange = (val: 'normal' | 'encrypted') => {
  if (val === 'normal') {
    formData.value.password = null
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
      optionsText: '',
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
    optionsText: '',
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
    field.options = optionsText.split('\n').filter((option: string) => option.trim())
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
          optionsText: '',
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
    // 准备要保存的数据，包括字段配置和小分类
    const dataToSave = {
      ...formData.value,
      fields: fieldsConfig.value.map(field => ({
        name: field.name,
        label: field.label,
        type: field.type,
        options: field.options,
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
    await fetchPrompts()
  } catch (error) {
    // 错误已在拦截器中处理
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

const handlePromptSkinCommand = (command: string) => {
  if (command === 'exportCurrent') {
    const skin = selectedPromptSkin.value
    downloadPromptSkinFile(
      `prompt-skin-${skin.id}.json`,
      {
        type: 'prompt-skin',
        version: 1,
        exportedAt: new Date().toISOString(),
        skin
      }
    )
    ElMessage.success(`已导出皮肤：${skin.name}`)
    return
  }

  if (command === 'exportAll') {
    downloadPromptSkinFile(
      'prompt-skin-collection.json',
      {
        type: 'prompt-skin-collection',
        version: 1,
        exportedAt: new Date().toISOString(),
        activeSkinId: selectedPromptSkinId.value,
        skins: allPromptSkins.value
      }
    )
    ElMessage.success('已导出全部皮肤配置')
    return
  }

  if (command === 'import') {
    promptSkinInputRef.value?.click()
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

const handlePromptSkinFileChange = async (event: Event) => {
  const input = event.target as HTMLInputElement | null
  const file = input?.files?.[0]
  if (!file) return

  try {
    const text = await file.text()
    const parsed = JSON.parse(text)
    const existingIds = new Set(allPromptSkins.value.map((skin) => skin.id))
    const importedSkins: PromptSkinPreset[] = []

    if (parsed?.type === 'prompt-skin-collection' && Array.isArray(parsed.skins)) {
      parsed.skins.forEach((item: any) => {
        const normalized = normalizePromptSkin(item, existingIds)
        if (normalized) {
          importedSkins.push(normalized)
        }
      })
    } else {
      const normalized = normalizePromptSkin(parsed?.skin ?? parsed, existingIds)
      if (normalized) {
        importedSkins.push(normalized)
      }
    }

    if (importedSkins.length === 0) {
      throw new Error('未找到可导入的皮肤配置')
    }

    customPromptSkins.value = [...customPromptSkins.value, ...importedSkins]
    selectedPromptSkinId.value = importedSkins[0].id
    ElMessage.success(`已导入 ${importedSkins.length} 套皮肤配置`)
  } catch (error: any) {
    ElMessage.error(`皮肤导入失败：${error?.message || '文件格式不正确'}`)
  } finally {
    if (input) {
      input.value = ''
    }
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
  packPreviewVisible.value = false
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
          fields: prompt.fields || [],
          subcategories: prompt.subcategories || [],
          card_type: prompt.card_type || 'normal'
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
              fields: prompt.fields || [],
              subcategories: prompt.subcategories || [],
              card_type: prompt.card_type || 'normal'
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
    selectedStandaloneCards.value = uncategorizedPrompts.value.map(p => p.id)
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
  selectAllStandalone.value = selectedStandaloneCards.value.length === uncategorizedPrompts.value.length
}

// 执行导出独立卡片
const executeExportStandalone = () => {
  const cardsToExport = uncategorizedPrompts.value.filter(p => selectedStandaloneCards.value.includes(p.id))
  const exportData = {
    type: 'standalone-cards',
    version: '1.0',
    exportTime: new Date().toISOString(),
    prompts: cardsToExport.map(p => ({
      name: p.name,
      content: p.content,
      category: p.category,
      order_num: p.order_num,
      fields: p.fields,
      card_type: p.card_type || 'normal'
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
      fields: p.fields,
      card_type: p.card_type || 'normal'
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
  // 收集所有独立卡片
  const standalone = uncategorizedPrompts.value.map(p => ({
    name: p.name,
    content: p.content,
    category: p.category,
    order_num: p.order_num,
    fields: p.fields,
    subcategories: p.subcategories,
    card_type: p.card_type || 'normal'
  }))
  
  // 收集所有卡包及其卡片
  const packs: Record<string, any[]> = {}
  categoryList.value.forEach(category => {
    if (category.name !== '未分类') {
      packs[category.name] = category.prompts.map(p => ({
        name: p.name,
        content: p.content,
        category: p.category,
        order_num: p.order_num,
        fields: p.fields,
        subcategories: p.subcategories,
        card_type: p.card_type || 'normal'
      }))
    }
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
        fields: card.fields,
        card_type: card.card_type || 'normal'
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
        fields: card.fields,
        card_type: card.card_type || 'normal'
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
  z-index: 10;
  pointer-events: auto;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pack-fan-board {
  display: grid;
  gap: 18px;
  justify-items: center;
  padding: 40px 40px 50px 120px;
  border-radius: 28px;
  background: transparent;
  box-shadow: none;
  backdrop-filter: none;
  pointer-events: auto;
  min-height: 320px;
}

.pack-fan-head {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.pack-fan-summary {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.pack-fan-summary-name {
  font-size: 18px;
  font-weight: 700;
  color: #184e4c;
}

.pack-fan-summary-count {
  padding: 4px 12px;
  border-radius: 999px;
  background: rgba(24, 119, 114, 0.1);
  color: #187772;
  font-size: 12px;
  font-weight: 700;
}

.pack-fan-close {
  border-color: rgba(24, 119, 114, 0.16);
  background: rgba(255, 255, 255, 0.86);
  color: #1b4d4a;
}

.pack-fan-row {
  width: 100%;
  display: flex;
  justify-content: flex-start;
  align-items: flex-end;
  gap: 0;
  overflow-x: auto;
  overflow-y: visible;
  padding: 50px 30px 10px 80px;
  perspective: 1600px;
  min-height: 320px;
}

.pack-fan-card {
  position: relative;
  flex: 0 0 156px;
  width: 156px;
  height: 220px;
  margin-left: -100px;
  isolation: isolate;
  padding: 0;
  border: none;
  background: transparent;
  cursor: pointer;
  transform: translateY(0) rotate(0deg);
  transform-origin: center bottom;
  transition: transform 420ms cubic-bezier(0.22, 1, 0.36, 1), filter 280ms ease;
  filter: drop-shadow(0 14px 22px rgba(15, 23, 42, 0.12));
}

.pack-fan-card:first-child {
  margin-left: 0;
}

.pack-fan-card::before {
  content: "";
  position: absolute;
  top: -10px;
  bottom: -10px;
  left: -12px;
  right: -12px;
  z-index: -1;
}

.pack-fan-card:first-child::before {
  left: 0;
  right: -12px;
}

.pack-fan-card:last-child::before {
  left: -12px;
  right: 0;
}

.pack-fan-card:hover {
  z-index: 12;
  transform: translateY(-28px) rotate(0deg) scale(1.03);
  filter: drop-shadow(0 24px 34px rgba(15, 23, 42, 0.2));
}

.pack-fan-row:hover .pack-fan-card:not(:hover) {
  filter: drop-shadow(0 10px 18px rgba(15, 23, 42, 0.08)) saturate(0.92);
}

.pack-fan-card-inner {
  position: relative;
  display: block;
  width: 100%;
  height: 100%;
  transform-style: preserve-3d;
  transition: transform 620ms cubic-bezier(0.2, 0.75, 0.25, 1);
}

.pack-fan-card:hover .pack-fan-card-inner {
  transform: rotateY(180deg);
}

.pack-fan-card-face {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  border-radius: 20px;
  overflow: hidden;
  backface-visibility: hidden;
  -webkit-backface-visibility: hidden;
}

.pack-fan-card-back {
  padding: 16px;
  border: 1px solid rgba(18, 74, 71, 0.14);
  background:
    linear-gradient(135deg, rgba(13, 148, 136, 0.94), rgba(17, 94, 89, 0.96)),
    repeating-linear-gradient(45deg, rgba(255, 255, 255, 0.08) 0, rgba(255, 255, 255, 0.08) 8px, transparent 8px, transparent 16px);
  box-shadow:
    0 12px 20px rgba(8, 47, 46, 0.12),
    inset 0 0 0 1px rgba(255, 255, 255, 0.22),
    inset 0 0 0 10px rgba(255, 255, 255, 0.06);
}

.pack-fan-card-back::before {
  content: "";
  position: absolute;
  inset: 12px;
  border-radius: 14px;
  border: 1px solid rgba(255, 255, 255, 0.26);
}

.pack-fan-card-back::after {
  content: "";
  position: absolute;
  inset: 28px;
  border-radius: 999px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.22), transparent 72%);
  mix-blend-mode: screen;
}

.pack-fan-card-back-core {
  position: relative;
  z-index: 1;
  display: grid;
  place-items: center;
  height: 100%;
  text-align: center;
  color: #effcfb;
}

.pack-fan-card-back-title {
  font-size: 18px;
  font-weight: 800;
  letter-spacing: 0.06em;
}

.pack-fan-card-back-mark {
  margin-top: 10px;
  font-size: 12px;
  letter-spacing: 0.24em;
  text-transform: uppercase;
  opacity: 0.76;
}

.pack-fan-card-front {
  padding: 12px;
  border: 1px solid rgba(39, 107, 103, 0.16);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.99) 0%, rgba(240, 248, 247, 0.98) 100%);
  box-shadow:
    0 18px 30px rgba(22, 77, 75, 0.14),
    0 1px 0 rgba(255, 255, 255, 0.9) inset;
  color: #163d3b;
  transform: rotateY(180deg);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-rendering: optimizeLegibility;
}

.pack-fan-card-front-header {
  display: flex;
  align-items: center;
  gap: 6px;
  background: linear-gradient(135deg, rgba(133, 184, 178, 0.2) 0%, rgba(245, 249, 248, 0.94) 100%);
  border-bottom: 1px solid rgba(39, 107, 103, 0.10);
  margin: -12px -12px 8px -12px;
  padding: 8px 12px;
  border-radius: 20px 20px 0 0;
}

.pack-fan-card-front-handle {
  color: #0d9488;
  font-size: 14px;
}

.pack-fan-card-front-name {
  font-size: 14px;
  font-weight: 700;
  color: #0f4c48;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: left;
  letter-spacing: 0.02em;
}

.pack-fan-card-front-lock {
  color: #e6a23c;
  font-size: 14px;
  flex-shrink: 0;
  margin-left: 4px;
}

.pack-fan-card-front-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
  overflow: hidden;
}

.pack-fan-card-front-tag {
  align-self: flex-start;
}

.pack-fan-card-front-preview {
  display: -webkit-box;
  color: #3d5c58;
  font-size: 12px;
  line-height: 1.6;
  text-align: left;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
  letter-spacing: 0.01em;
}

.pack-fan-card-front-footer {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  border-top: 1px solid rgba(39, 107, 103, 0.10);
  background: linear-gradient(180deg, rgba(241, 247, 246, 0.96) 0%, rgba(233, 241, 240, 0.98) 100%);
  margin: 8px -12px -12px -12px;
  padding: 6px 12px;
  border-radius: 0 0 20px 20px;
}

.pack-fan-card-front-time {
  font-size: 11px;
  color: #4a6b67;
  font-weight: 500;
}

.pack-fan-stage-enter-active,
.pack-fan-stage-leave-active {
  transition: opacity 220ms ease, transform 220ms ease;
}

.pack-fan-stage-enter-from,
.pack-fan-stage-leave-to {
  opacity: 0;
  transform: translate(-50%, calc(-50% + 18px));
}

@media (max-width: 900px) {
  .pack-fan-stage {
    width: calc(100vw - 28px);
  }

  .pack-fan-board {
    padding: 8px 4px 12px;
  }

  .pack-fan-row {
    justify-content: flex-start;
    gap: 0;
    padding: 14px 22px 8px;
  }

  .pack-fan-card {
    flex-basis: 138px;
    width: 138px;
    height: 200px;
    margin-left: -26px;
  }

  .pack-fan-card:hover {
    transform: translateY(-22px) rotate(0deg) scale(1.02);
  }
}

.prompts-container {
  background: var(--prompts-container-bg, #fff);
  border-radius: 16px;
  padding: 32px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
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
  align-items: flex-start;
  margin-bottom: 28px;
  padding-bottom: 20px;
  border-bottom: 2px solid #f0f2f5;
  gap: 16px;
}

.header-main {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.header h2 {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #52c41a 0%, #389e0d 100%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.skin-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.skin-toolbar-label {
  font-size: 12px;
  font-weight: 600;
  color: #607d79;
  white-space: nowrap;
}

.skin-select {
  width: 220px;
}

.skin-config-btn {
  border-radius: 999px;
}

.hidden-input {
  display: none;
}

.header-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

:deep(.el-button--primary) {
  background: linear-gradient(135deg, #52c41a 0%, #389e0d 100%);
  border: none;
}

:deep(.el-button--primary:hover) {
  background: linear-gradient(135deg, #73d13d 0%, #52c41a 100%);
}

:deep(.el-button--success) {
  background: linear-gradient(135deg, #722ed1 0%, #531dab 100%);
  border: none;
}

:deep(.el-button--success:hover) {
  background: linear-gradient(135deg, #9254de 0%, #722ed1 100%);
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

/* 卡包横向排列容器 */
.packs-wrapper {
  display: flex;
  flex-wrap: wrap;
  gap: 18px;
  align-items: flex-start;
}

/* 竖向卡包样式 - 基础 */
.card-pack-vertical {
  --pack-shell-top: var(--prompt-pack-shell-top, #edf5f4);
  --pack-shell-mid: var(--prompt-pack-shell-mid, #e5efee);
  --pack-shell-bottom: var(--prompt-pack-shell-bottom, #dbe8e7);
  --pack-shell-border: var(--prompt-pack-shell-border, rgba(74, 126, 123, 0.28));
  --pack-shell-hover-border: var(--prompt-pack-shell-hover-border, rgba(40, 111, 108, 0.44));
  --pack-header-top: var(--prompt-pack-header-top, #c3ddd8);
  --pack-header-mid: var(--prompt-pack-header-mid, #afd2cd);
  --pack-header-bottom: var(--prompt-pack-header-bottom, #98c2bc);
  --pack-header-border: var(--prompt-pack-header-border, rgba(31, 89, 86, 0.18));
  --pack-header-text: var(--prompt-pack-header-text, #154d4b);
  --pack-muted-text: var(--prompt-pack-muted-text, #607d79);
  --pack-badge-bg: var(--prompt-pack-badge-bg, rgba(239, 246, 245, 0.94));
  --pack-badge-border: var(--prompt-pack-badge-border, rgba(83, 131, 127, 0.18));
  --pack-badge-text: var(--prompt-pack-badge-text, #276b67);
  position: relative;
  background: linear-gradient(180deg, var(--pack-shell-top) 0%, var(--pack-shell-mid) 52%, var(--pack-shell-bottom) 100%);
  border-radius: var(--prompt-pack-shell-radius, 20px);
  border: 1px solid var(--pack-shell-border);
  box-shadow: var(--prompt-pack-shell-shadow, 0 14px 30px rgba(25, 70, 68, 0.12), inset 0 1px 0 rgba(255, 255, 255, 0.38));
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: all 0.3s ease;
  backdrop-filter: blur(8px);
}

.card-pack-vertical::before {
  content: '';
  position: absolute;
  inset: 0;
  background: var(--prompt-pack-shell-overlay, linear-gradient(180deg, rgba(255, 255, 255, 0.22), transparent 34%), linear-gradient(120deg, transparent 0%, rgba(255, 255, 255, 0.07) 48%, transparent 70%));
  pointer-events: none;
}

.card-pack-vertical:hover {
  transform: translateY(-4px);
  border-color: var(--pack-shell-hover-border);
  box-shadow: var(--prompt-pack-shell-hover-shadow, 0 20px 40px rgba(18, 84, 81, 0.16), inset 0 1px 0 rgba(255, 255, 255, 0.46));
}

/* 卡包尺寸 - 小 */
.card-pack-vertical.small {
  width: 160px;
  height: 200px;
}

/* 卡包尺寸 - 中 */
.card-pack-vertical.medium {
  width: 200px;
  height: 260px;
}

/* 卡包尺寸 - 大 */
.card-pack-vertical.large {
  width: 260px;
  height: 340px;
}

.card-pack-vertical.drag-over {
  border-color: var(--prompt-pack-drag-border, #3d8e89);
  background: var(--prompt-pack-drag-bg, linear-gradient(180deg, #e7f1f0 0%, #ddeceb 100%));
  transform: scale(1.02);
  box-shadow: var(--prompt-pack-drag-shadow, 0 22px 44px rgba(22, 96, 93, 0.18), inset 0 0 0 1px rgba(123, 177, 171, 0.18));
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

.standalone-card-actions :deep(.el-button),
.pack-card-actions :deep(.el-button) {
  min-width: 28px;
  height: 28px;
  padding: 0 8px;
  border-radius: 999px;
  border: 1px solid transparent;
  background: rgba(15, 118, 110, 0.06);
  transition: all 0.2s ease;
}

.standalone-card-actions :deep(.el-button.el-button--primary),
.pack-card-actions :deep(.el-button.el-button--primary) {
  color: #0f766e;
  border-color: rgba(15, 118, 110, 0.10);
}

.standalone-card-actions :deep(.el-button.el-button--primary:hover),
.pack-card-actions :deep(.el-button.el-button--primary:hover) {
  color: #fff;
  background: linear-gradient(135deg, #0f766e 0%, #14b8a6 100%);
  border-color: transparent;
  box-shadow: 0 8px 18px rgba(20, 184, 166, 0.22);
}

.standalone-card-actions :deep(.el-button.el-button--danger),
.pack-card-actions :deep(.el-button.el-button--danger) {
  color: #c2415b;
  background: rgba(225, 29, 72, 0.06);
  border-color: rgba(225, 29, 72, 0.10);
}

.standalone-card-actions :deep(.el-button.el-button--danger:hover),
.pack-card-actions :deep(.el-button.el-button--danger:hover) {
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

/* 卡包底部 */
.pack-footer-vertical {
  background:
    linear-gradient(135deg, var(--pack-header-top) 0%, var(--pack-header-mid) 52%, var(--pack-header-bottom) 100%);
  padding: 10px 12px;
  border-top: 1px solid var(--pack-header-border);
  display: flex;
  flex-direction: column;
  gap: 8px;
  position: relative;
}

.pack-footer-vertical::after {
  content: '';
  position: absolute;
  inset: 0;
  background: var(--prompt-pack-header-overlay, radial-gradient(circle at top, rgba(255, 255, 255, 0.16), transparent 42%), linear-gradient(180deg, rgba(255, 255, 255, 0.10), transparent 64%), linear-gradient(115deg, transparent 0%, rgba(255, 255, 255, 0.09) 48%, transparent 72%));
  pointer-events: none;
}

.pack-footer-info {
  display: flex;
  align-items: center;
  gap: 6px;
  position: relative;
  z-index: 1;
}

.pack-footer-icon {
  color: var(--pack-header-text);
  font-size: 18px;
}

.pack-footer-name {
  font-weight: 600;
  color: var(--pack-header-text);
  font-size: 13px;
  flex: 1;
}

.pack-footer-count {
  background: var(--pack-badge-bg);
  border: 1px solid var(--pack-badge-border);
  color: var(--pack-badge-text);
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
  min-width: 20px;
  text-align: center;
}

.pack-footer-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
  z-index: 1;
  flex-wrap: wrap;
}

.pack-footer-actions .el-button {
  background: var(--prompt-pack-action-bg, rgba(244, 249, 248, 0.58));
  border-color: var(--prompt-pack-action-border, rgba(210, 228, 225, 0.88));
  backdrop-filter: blur(8px);
  width: 30px;
  height: 30px;
  padding: 0;
  min-width: 0;
  box-shadow: var(--prompt-pack-action-shadow, 0 8px 18px rgba(19, 91, 87, 0.10));
}

.pack-footer-actions .el-button:hover {
  background: var(--prompt-pack-action-hover-bg, rgba(249, 252, 251, 0.96));
  border-color: var(--prompt-pack-action-hover-border, rgba(227, 238, 236, 0.96));
}

.pack-footer-actions .el-button.el-button--primary {
  color: var(--prompt-pack-action-primary, #1f6f69);
}

.pack-footer-actions .el-button.el-button--primary:hover {
  color: var(--prompt-pack-action-primary-hover, #145b56);
}

.pack-footer-actions .el-button.el-button--danger {
  color: var(--prompt-pack-action-danger, #c42727);
}

.pack-footer-actions .el-button.el-button--danger:hover {
  color: var(--prompt-pack-action-danger-hover, #a81e1e);
}



/* 卡片滚动区域 */
.cards-scroll-area {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: var(--prompt-pack-scroll-bg, linear-gradient(180deg, rgba(227, 238, 236, 0.82) 0%, rgba(239, 245, 244, 0.96) 100%));
}

/* 卡包内卡片 - 基础 */
.pack-card {
  background: var(--prompt-pack-card-bg, linear-gradient(180deg, rgba(246, 250, 249, 0.98) 0%, #eaf2f1 100%));
  border-radius: var(--prompt-pack-card-radius, 14px);
  border: var(--prompt-pack-card-border, 1px solid rgba(39, 107, 103, 0.14));
  cursor: grab;
  transition: all 0.2s ease;
  box-shadow: var(--prompt-pack-card-shadow, 0 8px 18px rgba(21, 77, 75, 0.08), inset 0 1px 0 rgba(255, 255, 255, 0.72));
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.pack-card:hover {
  transform: translateY(-2px);
  border-color: var(--prompt-pack-card-hover-border, rgba(24, 99, 95, 0.28));
  box-shadow: var(--prompt-pack-card-hover-shadow, 0 14px 26px rgba(18, 84, 81, 0.12), inset 0 1px 0 rgba(255, 255, 255, 0.82));
}

.pack-card:active {
  cursor: grabbing;
}

.pack-card.dragging {
  opacity: var(--prompt-pack-card-drag-opacity, 0.56);
  transform: scale(var(--prompt-pack-card-drag-scale, 0.97));
}

/* 卡包内卡片尺寸 - 小 */
.pack-card.small {
  min-height: 70px;
}

.pack-card.small .pack-card-header {
  padding: 6px 8px;
}

.pack-card.small .pack-card-drag-handle {
  font-size: 12px;
}

.pack-card.small .pack-card-name {
  font-size: 11px;
}

.pack-card.small .pack-card-content {
  padding: 0 8px;
}

.pack-card.small .pack-card-tag {
  transform: scale(0.75);
  transform-origin: left center;
}

.pack-card.small .pack-card-preview {
  font-size: 9px;
  -webkit-line-clamp: 1;
  line-clamp: 1;
}

.pack-card.small .pack-card-footer {
  padding: 4px 8px;
}

.pack-card.small .pack-card-time {
  font-size: 9px;
}

.pack-card.small .pack-card-actions :deep(.el-button) {
  padding: 1px 3px;
  font-size: 10px;
}

/* 卡包内卡片尺寸 - 中 */
.pack-card.medium {
  min-height: 90px;
}

.pack-card.medium .pack-card-header {
  padding: 8px 10px;
}

.pack-card.medium .pack-card-drag-handle {
  font-size: 14px;
}

.pack-card.medium .pack-card-name {
  font-size: 12px;
}

.pack-card.medium .pack-card-content {
  padding: 0 10px;
}

.pack-card.medium .pack-card-tag {
  transform: scale(0.85);
  transform-origin: left center;
}

.pack-card.medium .pack-card-preview {
  font-size: 10px;
  -webkit-line-clamp: 2;
  line-clamp: 2;
}

.pack-card.medium .pack-card-footer {
  padding: 6px 10px;
}

.pack-card.medium .pack-card-time {
  font-size: 10px;
}

.pack-card.medium .pack-card-actions :deep(.el-button) {
  padding: 2px 4px;
  font-size: 11px;
}

/* 卡包内卡片尺寸 - 大 */
.pack-card.large {
  min-height: 120px;
}

.pack-card.large .pack-card-header {
  padding: 10px 12px;
}

.pack-card.large .pack-card-drag-handle {
  font-size: 16px;
}

.pack-card.large .pack-card-name {
  font-size: 13px;
}

.pack-card.large .pack-card-content {
  padding: 0 12px;
}

.pack-card.large .pack-card-tag {
  transform: scale(0.9);
  transform-origin: left center;
}

.pack-card.large .pack-card-preview {
  font-size: 11px;
  -webkit-line-clamp: 2;
  line-clamp: 2;
}

.pack-card.large .pack-card-footer {
  padding: 8px 12px;
}

.pack-card.large .pack-card-time {
  font-size: 11px;
}

.pack-card.large .pack-card-actions :deep(.el-button) {
  padding: 3px 5px;
  font-size: 12px;
}

/* 卡包内卡片头部 */
.pack-card-header {
  display: flex;
  align-items: center;
  gap: 6px;
  background: var(--prompt-pack-card-header-bg, linear-gradient(135deg, rgba(133, 184, 178, 0.2) 0%, rgba(245, 249, 248, 0.94) 100%));
  border-bottom: 1px solid var(--prompt-pack-card-header-border, rgba(39, 107, 103, 0.10));
}

.pack-card-drag-handle {
  color: var(--prompt-pack-card-handle-color, #0d9488);
  cursor: grab;
}

.pack-card-name {
  font-weight: 600;
  color: var(--prompt-pack-card-name-color, #134e4a);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 卡包内卡片内容 */
.pack-card-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--prompt-pack-card-content-gap, 6px);
  overflow: hidden;
}

.pack-card-tag {
  align-self: flex-start;
}

.pack-card-preview {
  color: var(--prompt-pack-card-preview-color, #5f7f7b);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* 卡包内卡片底部 */
.pack-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid var(--prompt-pack-card-footer-border, rgba(39, 107, 103, 0.10));
  background: var(--prompt-pack-card-footer-bg, linear-gradient(180deg, rgba(241, 247, 246, 0.96) 0%, rgba(233, 241, 240, 0.98) 100%));
}

.pack-card-time {
  color: var(--prompt-pack-card-time-color, var(--pack-muted-text));
}

.pack-card-actions {
  display: flex;
  gap: 4px;
}

/* 空卡包提示 */
.empty-pack-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 30px 10px;
  color: var(--prompt-pack-empty-text, #7cb9b3);
  font-size: 12px;
  gap: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: var(--prompt-pack-empty-border, 1px dashed rgba(94, 234, 212, 0.32));
  border-radius: var(--prompt-pack-empty-radius, 14px);
  margin: 4px;
  background: var(--prompt-pack-empty-bg, linear-gradient(180deg, rgba(236, 254, 255, 0.88) 0%, rgba(255, 255, 255, 0.96) 100%));
}

.empty-pack-hint:hover {
  color: var(--prompt-pack-empty-hover-text, #14b8a6);
  border-color: var(--prompt-pack-empty-hover-border, rgba(45, 212, 191, 0.4));
  background: var(--prompt-pack-empty-hover-bg, linear-gradient(180deg, #f0fdfa 0%, #ffffff 100%));
}

/* 卡包底部操作 */
.pack-footer-vertical {
  padding: 12px;
  border-top: 1px solid var(--prompt-pack-footer-border, rgba(94, 234, 212, 0.2));
  display: flex;
  justify-content: center;
  gap: 8px;
  background: var(--prompt-pack-footer-bg, linear-gradient(180deg, rgba(246, 255, 255, 0.98) 0%, rgba(239, 250, 250, 1) 100%));
}

.pack-footer-vertical :deep(.el-button) {
  font-size: 12px;
  border-radius: 999px;
  padding-inline: 10px;
  border: 1px solid transparent;
  background: var(--prompt-pack-footer-button-bg, rgba(255, 255, 255, 0.62));
  backdrop-filter: blur(8px);
  box-shadow: var(--prompt-pack-footer-button-shadow, 0 8px 18px rgba(56, 189, 248, 0.08));
}

.pack-footer-vertical :deep(.el-button--primary) {
  color: var(--prompt-pack-footer-primary-color, #22c7b8);
  border-color: var(--prompt-pack-footer-primary-border, rgba(94, 234, 212, 0.34));
}

.pack-footer-vertical :deep(.el-button--primary:hover) {
  color: var(--prompt-pack-footer-primary-hover-color, #0f766e);
  background: var(--prompt-pack-footer-primary-hover-bg, rgba(236, 254, 255, 0.92));
  border-color: var(--prompt-pack-footer-primary-hover-border, rgba(94, 234, 212, 0.5));
}

.pack-footer-vertical :deep(.el-button--danger) {
  color: var(--prompt-pack-footer-danger-color, #fb8fa3);
  border-color: var(--prompt-pack-footer-danger-border, rgba(251, 113, 133, 0.2));
}

.pack-footer-vertical :deep(.el-button--danger:hover) {
  color: var(--prompt-pack-footer-danger-hover-color, #f43f5e);
  background: var(--prompt-pack-footer-danger-hover-bg, rgba(255, 241, 242, 0.95));
  border-color: var(--prompt-pack-footer-danger-hover-border, rgba(251, 113, 133, 0.34));
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
.format-description code {
  background-color: #f0f0f0;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
}

/* 字段配置样式 */
.field-config-section {
  margin-top: 12px;
}

.empty-fields {
  text-align: center;
  padding: 24px;
  background-color: #fafafa;
  border-radius: 8px;
  margin-top: 12px;
}

.empty-hint {
  margin-top: 12px;
  color: #909399;
  font-size: 14px;
}

.fields-list {
  margin-top: 12px;
}

.field-item {
  background-color: #fafafa;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
  border: 1px solid #e4e7ed;
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
  background: linear-gradient(135deg, #f0f0f0 0%, #d9d9d9 100%);
  border-color: #bfbfbf;
  color: #595959;
}

.field-name-input {
  width: 200px;
  font-weight: 600;
  font-family: 'Courier New', monospace;
}

.field-name {
  font-weight: 600;
  color: #303133;
  font-family: 'Courier New', monospace;
  background-color: #fffbe6;
  padding: 4px 8px;
  border-radius: 4px;
  border: 1px solid #ffe58f;
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

.pack-card.encrypted {
  border-color: rgba(230, 162, 60, 0.4) !important;
  box-shadow: 0 4px 12px rgba(230, 162, 60, 0.12), inset 0 1px 0 rgba(255, 255, 255, 0.72) !important;
}

.pack-card.encrypted:hover {
  border-color: rgba(230, 162, 60, 0.6) !important;
  box-shadow: 0 8px 20px rgba(230, 162, 60, 0.18), inset 0 1px 0 rgba(255, 255, 255, 0.82) !important;
}

.pack-card.encrypted .pack-card-header {
  background: linear-gradient(135deg, rgba(230, 162, 60, 0.15) 0%, rgba(245, 249, 248, 0.94) 100%) !important;
}

.pack-card-lock {
  color: #e6a23c;
  font-size: 14px;
  margin-left: auto;
  flex-shrink: 0;
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
</style>
