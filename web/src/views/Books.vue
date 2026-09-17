<template>
  <div class="books-container">
    <!-- ===== 仪表盘头部 ===== -->
    <header class="dash-header">
      <div class="header-actions">
        <el-button @click="triggerBookImport()" class="dash-btn">
          <el-icon><Upload /></el-icon>
          导入作品
        </el-button>
        <el-button type="primary" @click="handleCreate" class="dash-btn">
          <el-icon><Plus /></el-icon>
          创建作品
        </el-button>
      </div>
    </header>

    <!-- ===== 筛选与排序 ===== -->
    <div class="filters">
      <div class="seg">
        <span :class="{ active: filterKey === 'all' }" @click="setFilter('all')">全部</span>
        <span :class="{ active: filterKey === 'draft' }" @click="setFilter('draft')">创作中</span>
        <span :class="{ active: filterKey === 'published' }" @click="setFilter('published')">已发布</span>
        <span :class="{ active: filterKey === 'completed' }" @click="setFilter('completed')">已完结</span>
      </div>
      <span class="sort" @click="sortDesc = !sortDesc">排序：<b>{{ sortDesc ? '最近更新 ↓' : '最早更新 ↑' }}</b></span>
    </div>

    <!-- ===== 作品卡 ===== -->
    <div class="books-grid" v-if="filteredBooks.length">
      <div v-for="book in filteredBooks" :key="book.id" class="book-card" @click="handleWrite(book.id)">
        <div class="card-top">
          <span class="book-title">{{ book.title }}</span>
          <div class="chips">
            <span v-if="book.category" class="chip cat">{{ book.category }}</span>
            <span class="chip" :class="statusChipClass(book.status)">{{ getStatusText(book.status) }}</span>
          </div>
        </div>

        <div class="card-body">
          <div class="book-cover" @click.stop="handlePreviewBookCover(book)">
            <img v-if="book.cover" :src="book.cover" class="cover-image" alt="封面" />
            <div v-else class="cover-placeholder">
              <el-icon class="cover-icon"><Notebook /></el-icon>
            </div>
            <div class="cover-tag">书测</div>
          </div>
          <div class="info">
            <div class="words">
              <template v-if="book.word_count">
                已写 <b>{{ formatWordCount(book.word_count) }}</b> / 目标 {{ formatWordCount(targetWordsOf(book)) }} 字
                <template v-if="book.chapter_count"> · {{ book.chapter_count }} 章</template>
              </template>
              <template v-else-if="book.chapter_count">
                {{ book.chapter_count }} 章 · 目标 {{ formatWordCount(targetWordsOf(book)) }} 字
              </template>
              <template v-else>
                目标 {{ formatWordCount(targetWordsOf(book)) }} 字，还未开写
              </template>
            </div>
            <div class="prog">
              <i :style="{ width: bookProgress(book) + '%' }"></i>
            </div>
            <div class="upd">
              最近更新 {{ formatDateShort(book.updated_at || book.created_at) }} {{ formatTime(book.updated_at || book.created_at) }}
            </div>
            <div class="desc" v-if="book.description">{{ book.description }}</div>
          </div>
        </div>

        <div class="card-bottom" @click.stop>
          <span class="foot-tag">进度 {{ bookProgress(book) }}%</span>
          <div class="action-buttons">
            <el-tooltip content="封面" placement="top">
              <button class="action-btn cover" @click.stop="handleOpenCoverDialog(book)">
                <el-icon><Picture /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip content="下载" placement="top">
              <button class="action-btn download" @click="handleDownload(book)">
                <el-icon><Download /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip content="确认" placement="top">
              <button class="action-btn confirm" @click="handleEdit(book)">
                <el-icon><Check /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip content="信息" placement="top">
              <button class="action-btn info" @click.stop="handleShowInfo(book)">
                <el-icon><InfoFilled /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip content="删除" placement="top">
              <button class="action-btn delete" @click="handleDelete(book)">
                <el-icon><Delete /></el-icon>
              </button>
            </el-tooltip>
          </div>
        </div>
      </div>
    </div>

    <!-- ===== 空状态 ===== -->
    <div class="empty-state" v-else>
      <el-icon><Notebook /></el-icon>
      <p v-if="bookStore.books.length">该分类下暂无作品，换个筛选试试</p>
      <p v-else>还没有作品，点击右上角「创建作品」开始你的第一部小说</p>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑作品' : '创建新作品'"
      width="420px"
      class="simple-book-dialog"
      append-to-body
    >
      <div class="simple-form">
        <div class="form-group">
          <label class="form-label required">作品名称</label>
          <el-input v-model="formData.title" placeholder="输入作品名称" />
        </div>
        <div class="form-group">
          <label class="form-label">作品简介</label>
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="4"
            placeholder="作品简介（可选）"
            resize="none"
          />
        </div>
      </div>
      <template #footer>
        <div class="simple-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">
            {{ isEdit ? '保存' : '创建' }}
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 导入文件弹窗 -->
    <el-dialog
      v-model="importDialogVisible"
      title="导入作品"
      width="500px"
      :close-on-click-modal="false"
      append-to-body
    >
      <div
        class="import-drop-zone"
        :class="{ 'is-dragover': isDragOver }"
        @dragenter.prevent="isDragOver = true"
        @dragleave.prevent="isDragOver = false"
        @dragover.prevent
        @drop.prevent="handleFileDrop"
        @click="triggerFileSelect"
      >
        <div class="import-drop-content">
          <el-icon class="import-icon"><Upload /></el-icon>
          <p class="import-text">拖拽文件到此处，或点击选择文件</p>
          <p class="import-hint">支持 TXT 格式，可同时选择多个文件</p>
        </div>
      </div>

      <div v-if="selectedFiles.length > 0" class="selected-files">
        <div class="files-header">
          <span>已选择 {{ selectedFiles.length }} 个文件</span>
          <el-button type="danger" link @click="clearSelectedFiles">清空</el-button>
        </div>
        <div class="files-list">
          <div v-for="(file, index) in selectedFiles" :key="index" class="file-item">
            <el-icon><Document /></el-icon>
            <span class="file-name">{{ file.name }}</span>
            <span class="file-size">{{ formatFileSize(file.size) }}</span>
            <el-icon class="file-remove" @click="removeFile(index)"><Close /></el-icon>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="importDialogVisible = false">取消</el-button>
          <el-button
            type="primary"
            :disabled="selectedFiles.length === 0"
            @click="confirmImport"
          >
            确认导入
          </el-button>
        </div>
      </template>
    </el-dialog>

    <input
      ref="importFileInput"
      class="book-import-input"
      type="file"
      accept=".txt,text/plain"
      multiple
      @change="handleImportFileSelect"
    />

    <el-dialog
      v-model="coverDialogVisible"
      width="900px"
      :close-on-click-modal="false"
      append-to-body
      class="cover-dialog"
    >
      <template #header>
        <div class="cover-dialog-header">
          <span class="cover-dialog-title">设置封面 · {{ selectedBookForCover?.title || '' }}</span>
        </div>
      </template>
      <div class="cover-blend">
        <!-- ===== 左：实时预览 + 前后对比 ===== -->
        <div class="blend-left">
          <span class="blend-label">实 时 预 览</span>

          <div
            class="blend-stage"
            :class="{ dragover: isCoverDragOver }"
            @dragenter.prevent="isCoverDragOver = true"
            @dragleave.prevent="isCoverDragOver = false"
            @dragover.prevent
            @drop.prevent="handleCoverDrop"
          >
            <div v-if="isGeneratingCover" class="blend-gen">
              <el-icon class="spin-icon"><Loading /></el-icon>
              <p>正在生成封面...</p>
              <el-button type="danger" size="small" @click="stopCoverGenerate">
                <el-icon><Close /></el-icon>停止
              </el-button>
            </div>
            <template v-else-if="coverPreview">
              <img :src="coverPreview" class="blend-img" alt="封面预览" />
              <button type="button" class="blend-zoom" title="放大预览" @click="showCoverPreview">⤢</button>
            </template>
            <div v-else class="blend-ph" @click="triggerCoverSelect">
              <span class="ic">📖</span>
              <b>拖入图片 · 或 AI 生成</b>
              <small>JPG / PNG，建议 3:4</small>
            </div>
          </div>

          <!-- 前后对比（拖动中线） -->
          <div
            v-show="compareMode && coverPreview"
            ref="cmpRef"
            class="blend-cmp"
          >
            <img class="cmp-back" :src="selectedBookForCover?.cover || ''" alt="原封面" />
            <img class="cmp-front" :src="coverPreview" alt="新封面" :style="{ clipPath: 'inset(0 0 0 ' + cmpPos + '%)' }" />
            <div class="cmp-handle" :style="{ left: cmpPos + '%' }" @mousedown.prevent="cmpStart" @touchstart.prevent="cmpStart"></div>
            <span class="lbl old">原封面</span>
            <span class="lbl new">新封面</span>
          </div>
          <div v-if="compareMode && !coverPreview" class="blend-cmp-empty">请先上传或生成新封面</div>

          <div class="view-seg">
            <button type="button" :class="{ on: !compareMode }" @click="setCompare(false)">预览</button>
            <button type="button" :class="{ on: compareMode }" @click="setCompare(true)">⇔ 对比原封面</button>
          </div>
        </div>

        <!-- ===== 右：配置面板 ===== -->
        <div class="blend-right">
          <div class="seg">
            <button type="button" :class="{ on: coverActiveTab === 'upload' }" @click="coverActiveTab = 'upload'">上传封面</button>
            <button type="button" :class="{ on: coverActiveTab === 'ai-generate' }" @click="coverActiveTab = 'ai-generate'">AI 生成</button>
          </div>

          <!-- 上传面板（含历史封面画廊） -->
          <div v-show="coverActiveTab === 'upload'" class="blend-form">
            <div class="field">
              <label>选择图片</label>
              <button type="button" class="ctrl" @click="triggerCoverSelect">🖼 从电脑选择 JPG / PNG 图片</button>
            </div>
            <div class="field">
              <label>格式建议</label>
              <div class="format-chips"><span>比例 3:4</span><span>≤ 20MB</span><span>JPG / PNG</span></div>
            </div>
            <div class="field">
              <label>历史封面 <span class="hist-count">{{ coverHistory.length }} 张</span></label>
              <div class="hist-grid">
                <button
                  v-for="(h, i) in coverHistory"
                  :key="h.src"
                  type="button"
                  class="hist-tile"
                  :class="{ sel: h.src === coverPreview }"
                  @click="useHistoryCover(h)"
                >
                  <img :src="h.src" alt="" />
                  <span class="del" title="从历史移除" @click.stop="removeCoverHistory(i)">✕</span>
                  <span class="ck">✓</span>
                </button>
              </div>
              <div v-if="coverHistory.length === 0" class="hist-empty">暂无历史封面 — 上传或 AI 生成的封面会自动保存在这里，随时点选复用</div>
            </div>
            <div class="field">
              <label>快捷方式</label>
              <div class="upload-tip"><b>直接把图片拖到左侧预览区</b>松手即替换封面</div>
            </div>
          </div>

          <!-- AI 面板（原版生成参数区块） -->
          <div v-show="coverActiveTab === 'ai-generate'" class="blend-form">
            <div class="form-top">
              <el-button
                v-if="generationForm.selectedPromptId"
                type="primary"
                link
                size="small"
                @click="showFullPrompt"
              >
                <el-icon><View /></el-icon>查看完整提示词
              </el-button>
            </div>
            <div class="field">
              <label>小说名称</label>
              <el-input :model-value="generationForm.name" disabled />
            </div>
            <div class="field">
              <label>生成模型 <span class="req">*</span></label>
              <el-select
                v-model="generationForm.selectedModelId"
                placeholder="选择图片生成模型"
                :loading="modelsLoading"
                class="ctrl-select"
              >
                <el-option
                  v-for="m in coverModels"
                  :key="m.id"
                  :label="`${m.provider_name ? m.provider_name + ' - ' : ''}${m.name}`"
                  :value="m.id"
                />
              </el-select>
            </div>
            <div class="field">
              <label>提示词 <span class="req">*</span></label>
              <el-select
                v-model="generationForm.selectedPromptId"
                placeholder="选择封面生成提示词"
                clearable
                :loading="promptsLoading"
                @change="handlePromptSelect"
                class="ctrl-select"
              >
                <el-option
                  v-for="p in coverPrompts"
                  :key="p.id"
                  :label="p.name"
                  :value="p.id"
                />
              </el-select>
            </div>
            <div class="field">
              <label>参考图 <span class="opt-hint">（可选，风格对齐用）</span></label>
              <div class="reference-image-actions">
                <el-button @click="triggerReferenceImageSelect">
                  <el-icon><Upload /></el-icon>上传参考图
                </el-button>
                <el-button @click="openRankReferenceDialog">
                  <el-icon><Picture /></el-icon>排行榜选择
                </el-button>
              </div>
              <span class="reference-image-hint">支持 JPG、PNG、WEBP、GIF，单张最大 20MB</span>
              <div v-if="referenceImage" class="reference-image-preview">
                <img :src="referenceImage.previewUrl" alt="参考图预览" class="reference-image-thumb" />
                <div class="reference-image-meta">
                  <span class="reference-image-name">{{ referenceImage.name }}</span>
                  <span class="reference-image-size">{{ formatFileSize(referenceImage.size) }}</span>
                </div>
                <el-button type="danger" link @click="clearReferenceImage">移除</el-button>
              </div>
            </div>
            <template v-if="coverPromptFields.length > 0">
              <div v-for="field in coverPromptFields" :key="field.name" class="field">
                <label>
                  <span class="field-label-text">{{ field.label }}</span>
                  <el-tag
                    :type="field.required !== false ? 'danger' : 'info'"
                    size="small"
                    class="field-required-tag"
                  >
                    {{ field.required !== false ? '必选' : '选填' }}
                  </el-tag>
                </label>
                <el-input
                  v-if="field.type === 'text'"
                  v-model="coverFieldValues[field.name]"
                  :placeholder="`请输入${field.label}`"
                />
                <el-input
                  v-else-if="field.type === 'textarea'"
                  v-model="coverFieldValues[field.name]"
                  type="textarea"
                  :rows="3"
                  :placeholder="`请输入${field.label}`"
                />
                <el-select
                  v-else-if="field.type === 'select'"
                  v-model="coverFieldValues[field.name]"
                  :placeholder="`请选择${field.label}`"
                  class="cover-field-select"
                  popper-class="cover-field-select-popper"
                  :teleported="false"
                >
                  <el-option
                    v-for="(opt, optionIndex) in field.options"
                    :key="opt"
                    :label="field.optionLabels?.[optionIndex] || opt"
                    :value="opt"
                  />
                </el-select>
              </div>
            </template>
            <button
              type="button"
              class="gen-btn"
              :disabled="!generationForm.selectedModelId || isGeneratingCover"
              @click="handleGenerateCover"
            >
              <el-icon v-if="isGeneratingCover" class="spin"><Loading /></el-icon>
              {{ isGeneratingCover ? '正在生成...' : '✨ 开始生成' }}
            </button>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button
            v-if="coverPreview"
            @click="downloadCurrentCover"
          >
            下载封面
          </el-button>
          <el-button
            v-if="coverPreview"
            @click="showCoverPreview"
          >
            <el-icon><ZoomIn /></el-icon>
            放大预览
          </el-button>
          <el-button @click="coverDialogVisible = false">取消</el-button>
          <el-button
            type="primary"
            :disabled="!coverPreview"
            @click="handleSaveCover"
          >
            保存封面
          </el-button>
        </div>
      </template>
    </el-dialog>

    <input
      ref="coverFileInput"
      class="cover-file-input"
      type="file"
      accept="image/jpeg,image/png,image/jpg"
      @change="handleCoverFileSelect"
    />

    <input
      ref="referenceImageInput"
      class="cover-file-input"
      type="file"
      accept="image/jpeg,image/png,image/jpg,image/webp,image/gif"
      @change="handleReferenceImageSelect"
    />

    <el-dialog
      v-model="rankReferenceDialogVisible"
      title="排行榜参考图"
      width="880px"
      append-to-body
      class="rank-reference-dialog"
    >
      <div class="rank-reference-layout">
        <div class="rank-reference-groups">
          <button
            v-for="group in rankReferenceGroups"
            :key="group.key"
            type="button"
            class="rank-reference-group-btn"
            :class="{ active: activeRankGroupKey === group.key }"
            @click="selectRankGroup(group.key)"
          >
            {{ group.label }}
          </button>
        </div>
        <div class="rank-reference-content">
          <div v-if="activeRankGroup" class="rank-reference-categories">
            <button
              v-for="category in activeRankGroup.categories"
              :key="category.id"
              type="button"
              class="rank-reference-category-btn"
              :class="{ active: activeRankCategoryPath === category.path }"
              @click="selectRankCategory(category.path)"
            >
              {{ category.label }}
            </button>
          </div>
          <div v-if="rankReferenceLoading" class="rank-reference-state">正在加载排行榜封面...</div>
          <div v-else-if="rankReferenceBooks.length === 0" class="rank-reference-state">暂无可用封面</div>
          <div v-else class="rank-reference-grid">
            <button
              v-for="book in rankReferenceBooks"
              :key="`${activeRankCategoryPath}_${book.rank}_${book.coverUrl}`"
              type="button"
              class="rank-reference-card"
              @click="useRankReferenceBook(book)"
            >
              <img :src="book.coverUrl" :alt="book.title || '排行榜封面'" class="rank-reference-cover" />
              <div class="rank-reference-book-meta">
                <span class="rank-reference-rank">#{{ String(book.rank).padStart(2, '0') }}</span>
                <span class="rank-reference-title">{{ book.title || '未命名作品' }}</span>
                <span class="rank-reference-author">{{ book.author || '未知作者' }}</span>
              </div>
            </button>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 作品信息查看弹窗 · 报刊编辑风格 -->
    <el-dialog
      v-model="infoDialogVisible"
      width="460px"
      :align-center="true"
      :close-on-click-modal="true"
      append-to-body
      class="book-info-dialog"
      destroy-on-close
    >
      <!-- 报头 Header -->
      <template #header>
        <div class="press-header">
          <div class="press-header-top">
            <span>EST. 2026 · 文学编辑部</span>
            <span class="press-issue">№ {{ String(infoBook?.id || 0).padStart(3, '0') }}</span>
          </div>
          <div class="press-header-title">作品信息<span class="press-accent"> · DAILY</span></div>
        </div>
      </template>

      <div v-if="infoBook" class="press-body">
        <!-- 顶部：封面 + 标题输入 -->
        <div class="press-top-row">
          <div
            class="press-cover"
            @click="infoBook.cover && handlePreviewBookCover(infoBook)"
          >
            <img v-if="infoBook.cover" :src="infoBook.cover" class="press-cover-img" alt="封面" />
            <span v-else class="press-cover-text">{{ (infoBook.title || '书')[0] }}</span>
          </div>
          <div class="press-title-block">
            <label class="press-field-label">HEADLINE · 作品标题</label>
            <input
              class="press-title-input"
              v-model="infoEditTitle"
              placeholder="作品名称"
            />
          </div>
        </div>

        <!-- Metadata 版权栏 -->
        <div class="press-meta-grid">
          <div class="press-meta-item">
            <span class="press-k">分类</span>
            <span class="press-v">{{ infoBook.category || '未分类' }}</span>
          </div>
          <div class="press-meta-item">
            <span class="press-k">状态</span>
            <span class="press-v">{{ getStatusText(infoBook.status) }}</span>
          </div>
          <div class="press-meta-item">
            <span class="press-k">可见性</span>
            <span class="press-v">{{ infoBook.is_public ? '公开' : '私密' }}</span>
          </div>
          <div class="press-meta-item">
            <span class="press-k">作者</span>
            <span class="press-v">{{ infoBook.author || '未填写' }}</span>
          </div>
          <div class="press-meta-item">
            <span class="press-k">ID</span>
            <span class="press-v">#{{ infoBook.id }}</span>
          </div>
          <div class="press-meta-item">
            <span class="press-k">章节</span>
            <span class="press-v">{{ infoBook.chapter_count || 0 }}</span>
          </div>
          <div class="press-meta-item">
            <span class="press-k">创建</span>
            <span class="press-v">{{ formatDateFull(infoBook.created_at) }}</span>
          </div>
          <div class="press-meta-item">
            <span class="press-k">更新</span>
            <span class="press-v">{{ formatDateFull(infoBook.updated_at) }}</span>
          </div>
        </div>

        <!-- 作品简介（报纸正文） -->
        <div class="press-section">
          <div class="press-section-label"><span>SYNOPSIS · 作品简介</span></div>
          <textarea
            class="press-textarea"
            v-model="infoEditDesc"
            placeholder="作品简介（可选）"
            rows="4"
          ></textarea>
        </div>

        <!-- 目标字数 + 进度条（油墨条） -->
        <div class="press-progress-section">
          <div class="press-progress-head">
            <span class="press-progress-label">PROGRESS · 字数进度</span>
            <span class="press-pct">{{ bookProgress(infoBook) }}%</span>
          </div>
          <div class="press-progress-track">
            <div
              class="press-progress-fill"
              :style="{ width: bookProgress(infoBook) + '%' }"
            ></div>
          </div>
          <div class="press-word-counts">
            <span><b>{{ formatWordCount(infoBook.word_count || 0) }}</b> 字已写 / {{ formatWordCount(targetWordsOf(infoBook)) }} 字目标</span>
            <span>{{ infoBook.chapter_count || 0 }} 章</span>
          </div>
          <div class="press-target-edit">
            <span>TARGET ─</span>
            <input
              class="press-target-input"
              v-model="targetWordsInput"
              type="number"
              placeholder="目标字数"
              @keyup.enter="saveTargetWords"
            />
            <span>字</span>
            <button class="press-target-save" @click="saveTargetWords">设定</button>
          </div>
        </div>

        <!-- 标签 -->
        <div class="press-section">
          <div class="press-section-label"><span>TAGS · 标签</span></div>
          <div class="press-tags-row">
            <span
              v-for="tag in infoBook.tags || []"
              :key="tag"
              class="press-tag"
            >{{ tag }}</span>
            <span v-if="!infoBook.tags || infoBook.tags.length === 0" class="press-empty">暂无标签</span>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="press-footer">
          <span class="press-footer-note">— Editorial Dept. —</span>
          <button class="press-save-btn" @click="saveInfoEdit">保存</button>
        </div>
      </template>
    </el-dialog>

    <!-- 封面放大预览 -->
    <Teleport to="body">
      <div
        v-if="coverPreviewViewerVisible"
        class="cover-zoom-mask"
        @click="closeCoverZoom"
      >
        <div class="cover-zoom-stage" @click.stop>
          <img
            :src="bookCoverPreviewUrl"
            class="cover-zoom-image"
            alt="封面放大"
          />
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Clock, View, EditPen, Delete, FolderOpened, Lock, Check, Upload, Loading, Notebook, Picture, Download, InfoFilled, Document, Close, ZoomIn } from '@element-plus/icons-vue'
import { useBookStore } from '@/stores/book'
import { promptAPI, aiAPI, chapterAPI, configAPI } from '@/api'
import type { Book, Prompt, Chapter, ApiModel } from '@/types'
import { nextTick } from 'vue'

const router = useRouter()
const bookStore = useBookStore()

// ===== 仪表盘：筛选 / 排序 =====
const filterKey = ref<'all' | 'draft' | 'published' | 'completed'>('all')
const sortDesc = ref(true)

const setFilter = (key: 'all' | 'draft' | 'published' | 'completed') => {
  filterKey.value = key
}

const filteredBooks = computed(() => {
  let list = bookStore.books
  if (filterKey.value !== 'all') {
    list = list.filter(b => (b.status || 'draft') === filterKey.value)
  }
  const sorted = [...list]
  sorted.sort((a, b) => {
    const ta = new Date(a.updated_at || a.created_at).getTime()
    const tb = new Date(b.updated_at || b.created_at).getTime()
    return sortDesc.value ? tb - ta : ta - tb
  })
  return sorted
})

const statusChipClass = (status?: string) => {
  if (status === 'published') return 'on'
  if (status === 'completed') return 'fin'
  return ''
}

const DEFAULT_TARGET_WORDS = 100000

const targetWordsOf = (book: Book) => book.target_words || DEFAULT_TARGET_WORDS

const bookProgress = (book: Book) => {
  const wc = book.word_count || 0
  const tw = targetWordsOf(book)
  return Math.min(100, Math.round((wc / tw) * 100))
}

const formatWordCount = (n: number) => {
  if (n >= 10000) return (n / 10000).toFixed(1) + ' 万'
  return n.toLocaleString()
}

const dialogVisible = ref(false)
const isEdit = ref(false)
const generating = ref(false)
const selectedPromptId = ref<number>()
const prompts = ref<Prompt[]>([])
const importFileInput = ref<HTMLInputElement | null>(null)
const importingBookId = ref<number | null>(null)
const importDialogVisible = ref(false)
const selectedFiles = ref<File[]>([])
const isDragOver = ref(false)
const coverDialogVisible = ref(false)
const selectedBookForCover = ref<Book | null>(null)
const coverPreview = ref<string>('')
const coverFileInput = ref<HTMLInputElement | null>(null)
const isCoverDragOver = ref(false)
const referenceImageInput = ref<HTMLInputElement | null>(null)
const referenceImage = ref<{
  name: string
  size: number
  mimeType: string
  dataBase64: string
  previewUrl: string
} | null>(null)
const rankReferenceDialogVisible = ref(false)
const rankReferenceLoading = ref(false)
const rankReferenceGroups = ref<Array<{
  key: string
  label: string
  categories: Array<{
    id: string
    label: string
    path: string
    fullUrl: string
  }>
}>>([])
const activeRankGroupKey = ref('')
const activeRankCategoryPath = ref('')
const rankReferenceBooks = ref<Array<{
  rank: number
  title: string
  author: string
  coverUrl: string
  bookPath: string
}>>([])

// AI Cover Generation
const coverActiveTab = ref('upload')
const isGeneratingCover = ref(false)
const promptsLoading = ref(false)
const modelsLoading = ref(false)
const coverPrompts = ref<Prompt[]>([])
const coverModels = ref<ApiModel[]>([])
const generationForm = ref({
  name: '',
  selectedPromptId: undefined as number | undefined,
  selectedModelId: undefined as number | undefined
})
const generatedCoverUrl = ref('')

const coverAbortController = ref<AbortController | null>(null)

// 前后对比（09）
const compareMode = ref(false)
const cmpPos = ref(50)
const cmpDragging = ref(false)
const cmpRef = ref<HTMLElement | null>(null)

// 历史封面画廊（03，按作品持久化到 localStorage）
const coverHistory = ref<Array<{ src: string; name: string; ts: number }>>([])
const HISTORY_CAP = 8
const historyKey = computed(() => `cover_hist_${selectedBookForCover.value?.id || '0'}`)

const coverFieldValues = ref<Record<string, string>>({})
const coverPreviewViewerVisible = ref(false)
const bookCoverPreviewUrl = ref('')

type CoverPromptField = { name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; optionLabels?: string[]; description: string; required: boolean }

const coverPromptFields = ref<CoverPromptField[]>([])

const IMAGE_MODEL_PATTERN = /(image|imagen|dall[-_ ]?e|gpt-image|codex-gpt-image|flux|stable[-_ ]?diffusion|seedream|jimeng|midjourney|ideogram|recraft|hidream|playground|pixverse|gemini)/i
const MAX_REFERENCE_IMAGE_SIZE = 20 * 1024 * 1024
const SUPPORTED_REFERENCE_IMAGE_TYPES = /^(image\/(jpeg|png|jpg|webp|gif))$/i

const formData = ref({
  id: 0,
  title: '',
  description: '',
  author: '',
  category: '',
  tags: '',
  status: 'draft' as 'draft' | 'published' | 'completed',
  is_public: false
})

const categoryOptions = ['玄幻', '奇幻', '武侠', '仙侠', '都市', '历史', '军事', '游戏', '竞技', '科幻', '灵异', '同人', '轻小说', '其他']

const infoDialogVisible = ref(false)
const infoBook = ref<Book | null>(null)
const targetWordsInput = ref('')
const infoEditTitle = ref('')
const infoEditDesc = ref('')

type ParsedImportChapter = Pick<Chapter, 'title' | 'content'>

const CHAPTER_HEADING_REGEX = /^(?:\s| )*(第\s*[0-9零一二三四五六七八九十百千万两〇○O0-9①-⑳]+?\s*[章节卷回部篇集幕折][^\r\n]{0,30}|(?:序章|序言|前言|楔子|引子|终章|尾声|后记|番外(?:篇|合集)?|大结局)[^\r\n]{0,30}|chapter\s*[0-9ivxlcdm]+[^\r\n]{0,30})\s*$/gim

onMounted(async () => {
  await bookStore.fetchBooks()
  const res = await promptAPI.getAll()
  if (res.success && res.data) {
    prompts.value = res.data
  }
  // 前后对比：拖动中线（鼠标 + 触摸）
  document.addEventListener('mousemove', onCmpDocMove)
  document.addEventListener('mouseup', onCmpDocUp)
  document.addEventListener('touchmove', onCmpDocTouch, { passive: false })
  document.addEventListener('touchend', onCmpDocUp)
})

onBeforeUnmount(() => {
  document.removeEventListener('mousemove', onCmpDocMove)
  document.removeEventListener('mouseup', onCmpDocUp)
  document.removeEventListener('touchmove', onCmpDocTouch)
  document.removeEventListener('touchend', onCmpDocUp)
})

const handleCreate = () => {
  isEdit.value = false
  formData.value = {
    id: 0,
    title: '',
    description: '',
    author: '',
    category: '',
    tags: '',
    status: 'draft',
    is_public: false
  }
  dialogVisible.value = true
}

const handleEdit = (book: Book) => {
  isEdit.value = true
  formData.value = {
    id: book.id,
    title: book.title,
    description: book.description,
    author: book.author || '',
    category: book.category || '',
    tags: book.tags ? book.tags.join(',') : '',
    status: book.status || 'draft',
    is_public: book.is_public || false
  }
  dialogVisible.value = true
}

const handleWrite = (bookId: number) => {
  router.push(`/write/${bookId}`)
}

const triggerBookImport = () => {
  selectedFiles.value = []
  importDialogVisible.value = true
}

const triggerFileSelect = () => {
  if (importFileInput.value) {
    importFileInput.value.value = ''
    importFileInput.value.click()
  }
}

const formatFileSize = (size: number) => {
  if (size < 1024) return size + ' B'
  if (size < 1024 * 1024) return (size / 1024).toFixed(1) + ' KB'
  return (size / (1024 * 1024)).toFixed(1) + ' MB'
}

const handleFileDrop = (event: DragEvent) => {
  isDragOver.value = false
  const files = event.dataTransfer?.files
  if (files) {
    const txtFiles = Array.from(files).filter(file => /\.txt$/i.test(file.name))
    if (txtFiles.length === 0) {
      ElMessage.warning('请选择 TXT 格式的文件')
      return
    }
    selectedFiles.value.push(...txtFiles)
    ElMessage.success(`已添加 ${txtFiles.length} 个文件`)
  }
}

const handleImportFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  const files = target.files
  if (files && files.length > 0) {
    const txtFiles = Array.from(files).filter(file => /\.txt$/i.test(file.name))
    if (txtFiles.length === 0) {
      ElMessage.warning('请选择 TXT 格式的文件')
      return
    }
    selectedFiles.value.push(...txtFiles)
    ElMessage.success(`已添加 ${txtFiles.length} 个文件`)
  }
}

const removeFile = (index: number) => {
  selectedFiles.value.splice(index, 1)
}

const clearSelectedFiles = () => {
  selectedFiles.value = []
}

const confirmImport = async () => {
  if (selectedFiles.value.length === 0) return

  try {
    const fileNames = selectedFiles.value.map(f => f.name).join('、')
    await ElMessageBox.confirm(
      `确定要将 ${selectedFiles.value.length} 个文件（${fileNames}）作为新作品导入吗？`,
      '确认导入',
      {
        confirmButtonText: '开始导入',
        cancelButtonText: '取消',
        type: 'info'
      }
    )

    importDialogVisible.value = false
    await processImportFiles()
  } catch (error) {
    // 用户取消
  }
}

const processImportFiles = async () => {
  let totalChapters = 0
  let successCount = 0
  importingBookId.value = -1 // 使用一个特殊值表示正在导入新书

  try {
    for (const file of selectedFiles.value) {
      const text = await readTxtFile(file)
      if (!text) {
        ElMessage.warning(`文件 "${file.name}" 内容为空，已跳过`)
        continue
      }
      
      const baseTitle = file.name.replace(/\.txt$/i, '').trim()
      const newBook = await bookStore.createBook({ title: baseTitle, description: '由文件导入' })
      
      if (!newBook) {
        ElMessage.error(`为文件 "${file.name}" 创建作品失败`)
        continue
      }

      const parsedChapters = splitTxtIntoChapters(text, baseTitle)
      if (parsedChapters.length > 0) {
        const batchImport = await tryBatchImportChapters(newBook.id, parsedChapters)
        if (batchImport.fallback) {
          await createChaptersSequentially(newBook.id, parsedChapters, 0)
        }
        totalChapters += parsedChapters.length
      }

      successCount++
    }

    await bookStore.fetchBooks()
    ElMessage.success(`成功导入 ${successCount} 个文件，共创建 ${successCount} 个作品`)
  } catch (error) {
    if (error instanceof Error && error.message) {
      ElMessage.error(error.message)
    }
  } finally {
    importingBookId.value = null
    selectedFiles.value = []
  }
}

const detectTextEncoding = (bytes: Uint8Array) => {
  if (bytes.length >= 3 && bytes[0] === 0xef && bytes[1] === 0xbb && bytes[2] === 0xbf) {
    return 'utf-8'
  }
  if (bytes.length >= 2 && bytes[0] === 0xff && bytes[1] === 0xfe) {
    return 'utf-16le'
  }
  if (bytes.length >= 2 && bytes[0] === 0xfe && bytes[1] === 0xff) {
    return 'utf-16be'
  }
  return ''
}

const readTxtFile = async (file: File) => {
  const buffer = await file.arrayBuffer()
  const bytes = new Uint8Array(buffer)
  const preferredEncoding = detectTextEncoding(bytes)
  const encodings = [preferredEncoding, 'utf-8', 'gb18030', 'big5'].filter(Boolean)
  let bestText = ''
  let bestScore = Number.POSITIVE_INFINITY
  for (const encoding of encodings) {
    try {
      const text = new TextDecoder(encoding).decode(bytes)
      const invalidCharCount = (text.match(/\uFFFD/g) || []).length
      if (invalidCharCount < bestScore) {
        bestScore = invalidCharCount
        bestText = text
      }
      if (invalidCharCount === 0) {
        break
      }
    } catch (error) {
    }
  }
  return bestText
    .replace(/\r\n?/g, '\n')
    .replace(/\u0000/g, '')
    .replace(/^\uFEFF/, '')
    .trim()
}

const cleanImportTitle = (title: string, fallbackIndex: number) => {
  const normalizedTitle = title
    .replace(/^[\s ]+|[\s ]+$/g, '')
    .replace(/[：:]\s*$/, '')
    .replace(/\s+/g, ' ')
    .trim()
  return normalizedTitle || `第${fallbackIndex + 1}章`
}

const chunkPlainText = (text: string, baseTitle: string) => {
  const paragraphs = text
    .split(/\n{2,}/)
    .map(item => item.trim())
    .filter(Boolean)
  const chunks: ParsedImportChapter[] = []
  let currentChunk = ''
  for (const paragraph of paragraphs) {
    const nextChunk = currentChunk ? `${currentChunk}\n\n${paragraph}` : paragraph
    if (nextChunk.length > 6000 && currentChunk.length >= 1500) {
      chunks.push({ title: `第${chunks.length + 1}章`, content: currentChunk.trim() })
      currentChunk = paragraph
    } else {
      currentChunk = nextChunk
    }
  }
  if (currentChunk.trim()) {
    chunks.push({
      title: chunks.length === 0 ? `${baseTitle} 正文` : `第${chunks.length + 1}章`,
      content: currentChunk.trim()
    })
  }
  if (chunks.length > 1) {
    return chunks.map((chapter, index) => ({
      ...chapter,
      title: `第${index + 1}章`
    }))
  }
  return chunks
}

const splitTxtIntoChapters = (text: string, baseTitle: string) => {
  const normalizedText = text.replace(/\n{3,}/g, '\n\n').trim()
  if (!normalizedText) {
    return []
  }
  const matches = Array.from(normalizedText.matchAll(CHAPTER_HEADING_REGEX))
  if (matches.length === 0) {
    return chunkPlainText(normalizedText, baseTitle)
  }
  const chapters: ParsedImportChapter[] = []
  matches.forEach((match, index) => {
    const title = cleanImportTitle(match[0], index)
    const start = (match.index || 0) + match[0].length
    const end = index + 1 < matches.length ? (matches[index + 1].index || normalizedText.length) : normalizedText.length
    const content = normalizedText.slice(start, end).trim()
    if (content) {
      chapters.push({ title, content })
    }
  })
  if (chapters.length > 0) {
    return chapters
  }
  return chunkPlainText(normalizedText, baseTitle)
}

const createChaptersSequentially = async (
  bookId: number,
  chapters: ParsedImportChapter[],
  startOrder: number
) => {
  for (const [index, chapter] of chapters.entries()) {
    await chapterAPI.create({
      book_id: bookId,
      title: chapter.title,
      content: chapter.content,
      order_num: startOrder + index,
      type: 'chapter'
    })
  }
}

const showFullPrompt = () => {
  const selectedPrompt = coverPrompts.value.find(
    p => p.id === generationForm.value.selectedPromptId
  )
  if (!selectedPrompt) return

  // Build final prompt with variable replacement
  let finalPrompt = selectedPrompt.content || ''
  for (const field of coverPromptFields.value) {
    const value = coverFieldValues.value[field.name] || ''
    const dollarSyntax = new RegExp(`\\$\\{${field.name}\\}`, 'g')
    const curlySyntax = new RegExp(`\\{\\{${field.name}\\}\\}`, 'g')
    finalPrompt = finalPrompt.replace(dollarSyntax, value).replace(curlySyntax, value)
  }

  // Add system prompt with book name
  if (generationForm.value.name) {
    finalPrompt = `小说名称：${generationForm.value.name}\n\n${finalPrompt}`
  }
  
  ElMessageBox.alert(
    `<pre style="white-space: pre-wrap; word-break: break-word; max-height: 400px; overflow-y: auto; font-size: 13px; line-height: 1.6;">${finalPrompt}</pre>`,
    selectedPrompt.name,
    {
      dangerouslyUseHTMLString: true,
      confirmButtonText: '关闭',
      customClass: 'prompt-preview-dialog'
    }
  )
}

const tryBatchImportChapters = async (bookId: number, chapters: ParsedImportChapter[]) => {
  const response = await fetch('/api/chapters/import-book', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ bookId, chapters })
  })
  if (response.status === 404) {
    return { fallback: true as const }
  }
  const result = await response.json().catch(() => null)
  if (!response.ok) {
    throw new Error(result?.message || `Request failed with status code ${response.status}`)
  }
  return { fallback: false as const, result }
}

const handleDelete = async (book: Book) => {
  try {
    await ElMessageBox.confirm(`确定删除书本"${book.title}"吗？`, '提示', { type: 'warning' })
    await bookStore.deleteBook(book.id)
    ElMessage.success('删除成功')
  } catch (error) {
  }
}

const handleGenerateDescription = async () => {
  if (!formData.value.title) {
    ElMessage.warning('请先输入书名')
    return
  }
  try {
    generating.value = true
    const res = await aiAPI.generateDescription({
      title: formData.value.title,
      promptId: selectedPromptId.value
    })
    if (res.success && res.data) {
      formData.value.description = res.data
      ElMessage.success('生成成功')
    }
  } catch (error) {
  } finally {
    generating.value = false
  }
}

const handleSubmit = async () => {
  if (!formData.value.title) {
    ElMessage.warning('请输入书名')
    return
  }
  try {
    const submitData = {
      ...formData.value,
      tags: formData.value.tags
        ? formData.value.tags
            .split(',')
            .map(tag => tag.trim())
            .filter(Boolean)
        : []
    }
    if (isEdit.value) {
      await bookStore.updateBook(formData.value.id, submitData)
      ElMessage.success('更新成功')
    } else {
      await bookStore.createBook(submitData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
  } catch (error) {
  }
}

const getStatusText = (status?: string) => {
  const statusMap: Record<string, string> = {
    draft: '草稿',
    published: '已发布',
    completed: '已完结'
  }
  return statusMap[status || 'draft'] || '草稿'
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '未知'
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 7) return `${days}天前`
  if (days < 30) return `${Math.floor(days / 7)}周前`
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

const formatDateShort = (dateStr?: string) => {
  if (!dateStr) return '--'
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

const formatTime = (dateStr?: string) => {
  if (!dateStr) return '--'
  const date = new Date(dateStr)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

const handleDownload = async (book: Book) => {
  try {
    const res = await chapterAPI.getByBook(book.id)
    if (!res.success || !res.data) {
      ElMessage.warning('暂无内容可下载')
      return
    }

    const chapters = res.data.filter(ch => ch.type === 'chapter')
    if (chapters.length === 0) {
      ElMessage.warning('暂无章节可下载')
      return
    }

    let content = `${book.title}\n\n`
    chapters.forEach(chapter => {
      content += `${chapter.title}\n\n${chapter.content}\n\n`
    })

    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${book.title}.txt`
    link.click()
    URL.revokeObjectURL(url)

    ElMessage.success('下载成功')
  } catch (error) {
    ElMessage.error('下载失败')
  }
}

const handleOpenCoverDialog = async (book: Book) => {
  selectedBookForCover.value = book
  coverPreview.value = book.cover || ''
  generatedCoverUrl.value = ''
  coverActiveTab.value = 'upload'
  compareMode.value = false
  cmpPos.value = 50
  generationForm.value = {
    name: book.title,
    selectedPromptId: undefined,
    selectedModelId: undefined
  }
  clearReferenceImage()
  coverFieldValues.value = {}
  coverPromptFields.value = []

  // 历史封面：加载本作品历史；若为空则把当前封面作为最近一条
  loadCoverHistory()
  if (book.cover && coverHistory.value.length === 0) {
    addCoverHistory(book.cover, `${book.title} · 当前封面`)
  }

  promptsLoading.value = true
  modelsLoading.value = true
  try {
    const [promptRes, modelRes] = await Promise.all([
      promptAPI.getAll(),
      configAPI.getAll()
    ])

    if (promptRes.success && promptRes.data) {
      coverPrompts.value = promptRes.data.filter(
        p => p.category === '封面生成'
      )
    }

    if (modelRes.success && modelRes.data) {
      const enabled = modelRes.data.filter(m => m.enabled !== 0)
      coverModels.value = enabled.filter(m =>
        IMAGE_MODEL_PATTERN.test(m.model || m.name || '')
      )
      if (coverModels.value.length === 0) {
        coverModels.value = enabled
      }
      if (coverModels.value.length > 0) {
        const defaultModel = coverModels.value.find(m => m.is_default === 1)
        generationForm.value.selectedModelId = defaultModel
          ? defaultModel.id
          : coverModels.value[0].id
      }
    }
  } catch (error) {
    // fail silently
  } finally {
    promptsLoading.value = false
    modelsLoading.value = false
  }

  coverDialogVisible.value = true
}

const handlePromptSelect = (promptId: number | undefined) => {
  coverFieldValues.value = {}
  coverPromptFields.value = []
  if (!promptId) return

  const selected = coverPrompts.value.find(p => p.id === promptId)
  if (!selected) return

  const book = selectedBookForCover.value
  
  if (selected.fields && selected.fields.length > 0) {
    coverPromptFields.value = selected.fields as CoverPromptField[]
  } else {
    // Fallback to parsing placeholders from content
    const placeholderRegex = /\{\{([a-zA-Z0-9_]+)\}\}|\$\{([a-zA-Z0-9_]+)\}/g
    const placeholders = new Set<string>()
    let match
    while ((match = placeholderRegex.exec(selected.content || '')) !== null) {
      placeholders.add(match[1] || match[2])
    }
    
    coverPromptFields.value = Array.from(placeholders).map(name => ({
      name,
      label: name,
      type: 'text',
      options: [],
      description: '',
      required: true
    }))
  }

  // Pre-fill values
  coverPromptFields.value.forEach(field => {
    if (field.name === 'title' && generationForm.value.name) {
      coverFieldValues.value.title = generationForm.value.name
    }
    if (field.name === 'author' && book?.author) {
      coverFieldValues.value.author = book.author
    }
  })
}

const handleGenerateCover = async () => {
  if (!generationForm.value.selectedModelId) {
    ElMessage.warning('请选择生成模型')
    return
  }
  if (!generationForm.value.selectedPromptId) {
    ElMessage.warning('请选择提示词')
    return
  }
  if (isGeneratingCover.value) return

  const selectedPrompt = coverPrompts.value.find(
    p => p.id === generationForm.value.selectedPromptId
  )
  if (!selectedPrompt || !selectedPrompt.content) {
    ElMessage.warning('所选提示词内容为空')
    return
  }

  let finalPrompt = selectedPrompt.content
  for (const field of coverPromptFields.value) {
    const value = coverFieldValues.value[field.name] || ''
    const dollarSyntax = new RegExp(`\\$\\{${field.name}\\}`, 'g')
    const curlySyntax = new RegExp(`\\{\\{${field.name}\\}\\}`, 'g')
    finalPrompt = finalPrompt.replace(dollarSyntax, value).replace(curlySyntax, value)
  }

  // Add system prompt with book name
  if (generationForm.value.name) {
    finalPrompt = `小说名称：${generationForm.value.name}\n\n${finalPrompt}`
  }

  isGeneratingCover.value = true
  generatedCoverUrl.value = ''
  const abortController = new AbortController()
  coverAbortController.value = abortController

  try {
    const res = await aiAPI.generateImage({
      messages: [{ role: 'user', content: finalPrompt }],
      configId: generationForm.value.selectedModelId,
      size: '600x800',
      reference_images: referenceImage.value
        ? [{
            name: referenceImage.value.name,
            size: referenceImage.value.size,
            mime_type: referenceImage.value.mimeType,
            data_base64: referenceImage.value.dataBase64
          }]
        : undefined
    }, {
      signal: abortController.signal
    })
    if (res.success && res.data?.url) {
      generatedCoverUrl.value = res.data.url
      coverPreview.value = res.data.url
      addCoverHistory(res.data.url, 'AI 生成封面')
      ElMessage.success('封面生成成功，已更新预览')
    } else {
      ElMessage.error(res.message || '生成失败')
    }
  } catch (error: any) {
    if (abortController.signal.aborted || error.name === 'AbortError' || error.code === 'ERR_CANCELED' || error.message === 'canceled') {
      ElMessage.info('已停止生成')
      return
    }
    if (error.name !== 'AbortError' && error.message !== 'canceled') {
      ElMessage.error('生成请求失败')
    }
  } finally {
    if (coverAbortController.value === abortController) {
      coverAbortController.value = null
    }
    isGeneratingCover.value = false
  }
}

const stopCoverGenerate = () => {
  if (coverAbortController.value) {
    coverAbortController.value.abort()
    isGeneratingCover.value = false
  }
}

const triggerReferenceImageSelect = () => {
  if (referenceImageInput.value) {
    referenceImageInput.value.value = ''
    referenceImageInput.value.click()
  }
}

const clearReferenceImage = () => {
  if (referenceImage.value?.previewUrl) {
    URL.revokeObjectURL(referenceImage.value.previewUrl)
  }
  referenceImage.value = null
}

const setReferenceImageFromRemote = async (imageUrl: string, name: string) => {
  const response = await fetch(imageUrl)
  const blob = await response.blob()
  if (blob.size > MAX_REFERENCE_IMAGE_SIZE) {
    throw new Error('排行榜封面大小超过 20MB，无法作为参考图')
  }
  const mimeType = blob.type || 'image/jpeg'
  if (!SUPPORTED_REFERENCE_IMAGE_TYPES.test(mimeType)) {
    throw new Error('排行榜封面格式不受支持')
  }
  const dataBase64 = await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = typeof reader.result === 'string' ? reader.result : ''
      resolve(result.includes(',') ? result.split(',')[1] : result)
    }
    reader.onerror = () => reject(new Error('排行榜封面读取失败'))
    reader.readAsDataURL(blob)
  })
  clearReferenceImage()
  referenceImage.value = {
    name,
    size: blob.size,
    mimeType,
    dataBase64,
    previewUrl: imageUrl
  }
}

const loadRankReferenceGroups = async () => {
  const res = await aiAPI.getRankReferenceCategories()
  if (!res.success || !res.data) {
    throw new Error(res.message || '加载排行榜分类失败')
  }
  rankReferenceGroups.value = res.data
  if (!activeRankGroupKey.value && res.data.length > 0) {
    activeRankGroupKey.value = res.data[0].key
  }
}

const activeRankGroup = computed(() =>
  rankReferenceGroups.value.find(group => group.key === activeRankGroupKey.value) || null
)

const loadRankReferenceBooks = async (path: string) => {
  rankReferenceLoading.value = true
  try {
    const res = await aiAPI.getRankReferenceBooks(path)
    if (!res.success || !res.data) {
      throw new Error(res.message || '加载排行榜封面失败')
    }
    rankReferenceBooks.value = res.data
  } finally {
    rankReferenceLoading.value = false
  }
}

const selectRankGroup = async (groupKey: string) => {
  activeRankGroupKey.value = groupKey
  const group = rankReferenceGroups.value.find(item => item.key === groupKey)
  if (!group || group.categories.length === 0) {
    activeRankCategoryPath.value = ''
    rankReferenceBooks.value = []
    return
  }
  await selectRankCategory(group.categories[0].path)
}

const selectRankCategory = async (path: string) => {
  activeRankCategoryPath.value = path
  await loadRankReferenceBooks(path)
}

const openRankReferenceDialog = async () => {
  rankReferenceDialogVisible.value = true
  if (rankReferenceGroups.value.length === 0) {
    try {
      rankReferenceLoading.value = true
      await loadRankReferenceGroups()
      if (activeRankGroup.value?.categories?.[0]?.path) {
        await selectRankCategory(activeRankGroup.value.categories[0].path)
      }
    } catch (error: any) {
      ElMessage.error(error.message || '加载排行榜参考图失败')
    } finally {
      rankReferenceLoading.value = false
    }
    return
  }

  if (!activeRankCategoryPath.value && activeRankGroup.value?.categories?.[0]?.path) {
    try {
      await selectRankCategory(activeRankGroup.value.categories[0].path)
    } catch (error: any) {
      ElMessage.error(error.message || '加载排行榜参考图失败')
    }
  }
}

const useRankReferenceBook = async (book: { rank: number; title: string; author: string; coverUrl: string }) => {
  try {
    await setReferenceImageFromRemote(book.coverUrl, `${book.title || 'rank-cover'}_${book.rank}.jpg`)
    rankReferenceDialogVisible.value = false
    ElMessage.success('已选中排行榜封面作为参考图')
  } catch (error: any) {
    ElMessage.error(error.message || '设置排行榜参考图失败')
  }
}

const handleReferenceImageSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  if (!SUPPORTED_REFERENCE_IMAGE_TYPES.test(file.type)) {
    ElMessage.error('请上传 JPG、PNG、WEBP 或 GIF 格式的图片')
    target.value = ''
    return
  }
  if (file.size > MAX_REFERENCE_IMAGE_SIZE) {
    ElMessage.error('参考图大小不能超过 20MB')
    target.value = ''
    return
  }

  const reader = new FileReader()
  reader.onload = () => {
    const result = typeof reader.result === 'string' ? reader.result : ''
    const base64 = result.includes(',') ? result.split(',')[1] : result
    clearReferenceImage()
    referenceImage.value = {
      name: file.name,
      size: file.size,
      mimeType: file.type || 'image/png',
      dataBase64: base64,
      previewUrl: URL.createObjectURL(file)
    }
  }
  reader.onerror = () => {
    ElMessage.error('参考图读取失败')
  }
  reader.readAsDataURL(file)
}

const getCoverDownloadName = () => {
  const title = selectedBookForCover.value?.title?.trim() || 'cover'
  return `${title}_${Date.now()}.png`
}

const downloadCoverByUrl = async (imageUrl: string) => {
  try {
    const response = await fetch(imageUrl)
    const blob = await response.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = getCoverDownloadName()
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    ElMessage.success('下载成功')
  } catch (error) {
    ElMessage.error('下载失败')
  }
}

const downloadCurrentCover = async () => {
  if (!coverPreview.value) return
  await downloadCoverByUrl(coverPreview.value)
}

const triggerCoverSelect = () => {
  if (coverFileInput.value) {
    coverFileInput.value.value = ''
    coverFileInput.value.click()
  }
}

const handleCoverDrop = (event: DragEvent) => {
  isCoverDragOver.value = false
  const files = event.dataTransfer?.files
  if (files && files.length > 0) {
    handleCoverFile(files[0])
  }
}

const handleCoverFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    handleCoverFile(target.files[0])
  }
}

const handleCoverFile = (file: File) => {
  if (!file.type.match(/^image\/(jpeg|png|jpg)$/)) {
    ElMessage.error('请上传 JPG 或 PNG 格式的图片')
    return
  }
  const reader = new FileReader()
  reader.onload = (e) => {
    const src = e.target?.result as string
    coverPreview.value = src
    addCoverHistory(src, file.name)
  }
  reader.readAsDataURL(file)
}

// ===== 历史封面画廊（03，按作品持久化）=====
const loadCoverHistory = () => {
  try {
    coverHistory.value = JSON.parse(localStorage.getItem(historyKey.value) || '[]') || []
  } catch (e) {
    coverHistory.value = []
  }
}

const persistCoverHistory = () => {
  try {
    localStorage.setItem(historyKey.value, JSON.stringify(coverHistory.value))
  } catch (e) {
    // localStorage 超限时静默降级：仅本次会话生效
  }
}

const addCoverHistory = (src: string, name?: string) => {
  if (!src) return
  coverHistory.value = coverHistory.value.filter(h => h.src !== src)
  coverHistory.value.unshift({ src, name: name || '未命名', ts: Date.now() })
  if (coverHistory.value.length > HISTORY_CAP) coverHistory.value.length = HISTORY_CAP
  persistCoverHistory()
}

const useHistoryCover = (h: { src: string; name: string; ts: number }) => {
  // 最近使用在前：点选即移到画廊首位，不重复归档
  coverHistory.value = coverHistory.value.filter(x => x.src !== h.src)
  coverHistory.value.unshift(h)
  persistCoverHistory()
  coverPreview.value = h.src
}

const removeCoverHistory = (index: number) => {
  coverHistory.value.splice(index, 1)
  persistCoverHistory()
}

// ===== 前后对比（09）=====
const setCompare = (v: boolean) => {
  if (v) {
    if (!coverPreview.value) {
      ElMessage.warning('请先上传或生成封面，再进行对比')
      return
    }
    if (!selectedBookForCover.value?.cover) {
      ElMessage.warning('本书暂无原封面可对比')
      return
    }
  }
  compareMode.value = v
}

const setCmpPos = (e: MouseEvent | TouchEvent) => {
  const el = cmpRef.value
  if (!el) return
  const r = el.getBoundingClientRect()
  if (!r.width) return
  const x = 'touches' in e ? e.touches[0].clientX : e.clientX
  cmpPos.value = Math.max(3, Math.min(97, ((x - r.left) / r.width) * 100))
}

const cmpStart = (e: MouseEvent | TouchEvent) => {
  cmpDragging.value = true
  setCmpPos(e)
}

const onCmpDocMove = (e: MouseEvent) => {
  if (cmpDragging.value) setCmpPos(e)
}

const onCmpDocUp = () => {
  cmpDragging.value = false
}

const onCmpDocTouch = (e: TouchEvent) => {
  if (cmpDragging.value) {
    e.preventDefault()
    setCmpPos(e)
  }
}

const handleSaveCover = async () => {
  if (!selectedBookForCover.value || !coverPreview.value) return

  try {
    await bookStore.updateBook(selectedBookForCover.value.id, {
      cover: coverPreview.value
    })
    ElMessage.success('封面设置成功')
    coverDialogVisible.value = false
  } catch (error) {
    ElMessage.error('封面设置失败')
  }
}

const showCoverPreview = () => {
  if (!coverPreview.value) return
  bookCoverPreviewUrl.value = coverPreview.value
  coverPreviewViewerVisible.value = true
}

const handleShowInfo = (book: Book) => {
  infoBook.value = book
  infoEditTitle.value = book.title || ''
  infoEditDesc.value = book.description || ''
  targetWordsInput.value = book.target_words ? String(book.target_words) : ''
  infoDialogVisible.value = true
}

const saveInfoEdit = async () => {
  if (!infoBook.value) return
  const title = infoEditTitle.value.trim()
  if (!title) {
    ElMessage.warning('请输入作品名称')
    return
  }
  const updated = await bookStore.updateBook(infoBook.value.id, {
    title,
    description: infoEditDesc.value
  })
  if (updated) {
    infoBook.value = updated
    ElMessage.success('保存成功')
    infoDialogVisible.value = false
  } else {
    ElMessage.error('保存失败，请重试')
  }
}

const saveTargetWords = async () => {
  if (!infoBook.value) return
  const raw = targetWordsInput.value.trim()
  let tw: number | undefined
  if (raw) {
    tw = Math.floor(Number(raw))
    if (!isFinite(tw) || tw <= 0) {
      ElMessage.warning('请输入大于 0 的目标字数')
      return
    }
  }
  const updated = await bookStore.updateBook(infoBook.value.id, { target_words: tw ?? 0 })
  if (updated) {
    infoBook.value = updated
    targetWordsInput.value = updated.target_words ? String(updated.target_words) : ''
    ElMessage.success(
      tw
        ? `目标字数已设置为 ${formatWordCount(tw)}`
        : '已清除自定义目标，进度按 10 万字计算'
    )
  } else {
    ElMessage.error('保存失败，请重试')
  }
}

const formatDateFull = (dateStr?: string) => {
  if (!dateStr) return '未知'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const handlePreviewBookCover = (book: Book) => {
  if (!book.cover) return
  bookCoverPreviewUrl.value = book.cover
  coverPreviewViewerVisible.value = true
}

const closeCoverZoom = () => {
  coverPreviewViewerVisible.value = false
}
</script>

<style scoped>
.books-container {
  --c-bg: #f4f6fb;
  --c-card: #ffffff;
  --c-border: #e7eaf2;
  --c-text: #23283a;
  --c-sub: #6b7385;
  --c-muted: #9aa1b3;
  --c-line: #eef0f6;
  --c-accent: #6c8cff;
  --c-accent2: #9c6cff;
  background: var(--c-bg);
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 26px 32px 0;
  transition: background .3s;
}

/* ---------- 仪表盘头部 ---------- */
.dash-header {
  flex-shrink: 0;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  margin-bottom: 22px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.dash-btn.el-button {
  border-radius: 12px;
  font-weight: 600;
  border-color: var(--c-border);
  color: var(--c-sub);
  background: var(--c-card);
  box-shadow: 0 2px 8px rgba(30, 40, 60, .04);
  transition: all .18s;
}

.dash-btn.el-button:hover {
  border-color: var(--c-accent);
  color: var(--c-accent);
  box-shadow: 0 4px 14px rgba(108, 140, 255, .18);
}

.dash-btn.el-button--primary {
  background: linear-gradient(135deg, var(--c-accent), var(--c-accent2));
  border: none;
  color: #fff;
}

.dash-btn.el-button--primary:hover {
  box-shadow: 0 8px 22px rgba(108, 140, 255, .32);
  color: #fff;
}

/* ---------- 筛选与排序 ---------- */
.filters {
  flex-shrink: 0;
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  align-items: center;
  flex-wrap: wrap;
}

.seg {
  display: flex;
  gap: 4px;
  background: var(--c-line);
  padding: 4px;
  border-radius: 12px;
}

.seg span {
  padding: 7px 18px;
  border-radius: 9px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 600;
  color: var(--c-sub);
  transition: all .15s;
}

.seg span:hover { color: var(--c-text); }

.seg span.active {
  background: var(--c-card);
  color: var(--c-accent);
  box-shadow: 0 2px 8px rgba(60, 80, 130, .12);
}

.filters .sort {
  margin-left: auto;
  font-size: 12px;
  color: var(--c-muted);
  cursor: pointer;
  user-select: none;
}

.filters .sort b { color: var(--c-sub); font-weight: 600; }
.filters .sort:hover b { color: var(--c-accent); }

/* ---------- 空状态 ---------- */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 80px 0;
  color: var(--c-muted);
  background: var(--c-card);
  border: 1px dashed var(--c-border);
  border-radius: 16px;
}

.empty-state .el-icon { font-size: 48px; }
.empty-state p { font-size: 13px; }

/* ---------- 作品网格（唯一滚动区域） ---------- */
.books-grid {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 16px;
  align-content: start;
  padding: 2px 6px 44px;
}

.book-card {
  background: var(--c-card);
  border: 1px solid var(--c-border);
  border-radius: 16px;
  padding: 16px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 12px;
  transition: all .2s ease;
}

.book-card:hover {
  transform: translateY(-3px);
  border-color: #c9d4ff;
  box-shadow: 0 12px 30px rgba(60, 80, 130, .1);
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.book-title {
  flex: 1;
  min-width: 0;
  font-size: 16px;
  font-weight: 800;
  color: var(--c-text);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chips {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.chip {
  font-size: 10px;
  font-weight: 700;
  padding: 3px 9px;
  border-radius: 999px;
  background: var(--c-line);
  color: var(--c-muted);
  white-space: nowrap;
}

.chip.cat { background: #eef1fb; color: #6c8cff; }
.chip.on { background: #e2f7ee; color: #16a365; }
.chip.fin { background: #eef4ff; color: #4f8ff0; }

.card-body {
  display: flex;
  gap: 14px;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.book-cover {
  position: relative;
  width: 100px;
  aspect-ratio: 3 / 4;
  flex-shrink: 0;
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 6px 16px rgba(60, 80, 130, .16);
}

.cover-zoom-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  cursor: zoom-out;
}

.cover-zoom-stage {
  cursor: default;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cover-zoom-image {
  display: block;
  max-width: 80vw;
  max-height: 80vh;
  object-fit: contain;
  background: #1f2937;
  border-radius: 8px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.35);
}

.cover-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cover-icon {
  font-size: 28px;
  color: rgba(255, 255, 255, 0.8);
}

.cover-tag {
  position: absolute;
  bottom: 3px;
  right: 3px;
  background: rgba(0, 0, 0, 0.7);
  color: #ffffff;
  font-size: 9px;
  padding: 1px 4px;
  border-radius: 3px;
}

.info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 7px;
}

.info .words {
  font-size: 12px;
  color: var(--c-sub);
}

.info .words b {
  color: var(--c-text);
  font-weight: 700;
}

.info .prog {
  height: 5px;
  border-radius: 3px;
  background: var(--c-line);
  overflow: hidden;
}

.info .prog i {
  display: block;
  height: 100%;
  border-radius: 3px;
  background: linear-gradient(90deg, var(--c-accent), var(--c-accent2));
  transition: width .4s ease;
}

.info .upd {
  font-size: 11px;
  color: var(--c-muted);
}

.info .desc {
  font-size: 12px;
  color: var(--c-sub);
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid var(--c-line);
  padding-top: 10px;
  flex-shrink: 0;
}

.foot-tag {
  font-size: 11px;
  color: var(--c-muted);
}

.action-buttons {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.action-btn {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: 1px solid var(--c-border);
  background: var(--c-card);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all .15s ease;
  font-size: 13px;
  flex-shrink: 0;
  color: var(--c-muted);
}

.action-btn:hover {
  transform: translateY(-1px);
}

.action-btn.create:hover {
  background: var(--c-accent);
  border-color: var(--c-accent);
  color: #fff;
  box-shadow: 0 4px 12px rgba(108, 140, 255, .3);
}

.action-btn.cover:hover {
  background: #22d3ee;
  border-color: #22d3ee;
  color: #fff;
  box-shadow: 0 4px 12px rgba(34, 211, 238, .3);
}

.action-btn.download:hover {
  background: #34d399;
  border-color: #34d399;
  color: #fff;
  box-shadow: 0 4px 12px rgba(52, 211, 153, .3);
}

.action-btn.confirm:hover {
  background: #fbbf24;
  border-color: #fbbf24;
  color: #fff;
  box-shadow: 0 4px 12px rgba(251, 191, 36, .3);
}

.action-btn.info:hover {
  background: #60a5fa;
  border-color: #60a5fa;
  color: #fff;
  box-shadow: 0 4px 12px rgba(96, 165, 250, .3);
}

.action-btn.delete:hover {
  background: #ef5350;
  border-color: #ef5350;
  color: #fff;
  box-shadow: 0 4px 12px rgba(239, 83, 80, .3);
}

.action-btn.loading,
.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.book-import-input {
  display: none;
}

.import-drop-zone {
  border: 2px dashed #d9d9d9;
  border-radius: 8px;
  padding: 40px 20px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
  background: #fafafa;
}

.import-drop-zone:hover {
  border-color: #8b5cf6;
  background: #f5f3ff;
}

.import-drop-zone.is-dragover {
  border-color: #8b5cf6;
  background: #ede9fe;
}

.import-drop-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.import-icon {
  font-size: 48px;
  color: #8b5cf6;
}

.import-text {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin: 0;
}

.import-hint {
  font-size: 12px;
  color: #999;
  margin: 0;
}

.selected-files {
  margin-top: 20px;
}

.files-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.files-list {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-bottom: 1px solid #f0f0f0;
}

.file-item:last-child {
  border-bottom: none;
}

.file-item .el-icon {
  color: #8b5cf6;
}

.file-name {
  flex: 1;
  font-size: 13px;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  font-size: 12px;
  color: #999;
}

.file-remove {
  cursor: pointer;
  color: #999 !important;
  transition: color 0.2s;
}

.file-remove:hover {
  color: #ef4444 !important;
}

.dialog-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-section {
  background: #fafafa;
  border-radius: 8px;
  padding: 16px;
}

.section-header {
  margin-bottom: 16px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group.full {
  grid-column: 1 / -1;
}

.form-label {
  font-size: 13px;
  font-weight: 500;
  color: #666;
}

.form-label.required::after {
  content: '*';
  color: #ff4d4f;
  margin-left: 4px;
}

.description-editor {
  position: relative;
}

.char-count {
  position: absolute;
  bottom: 10px;
  right: 12px;
  font-size: 11px;
  color: #999;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

@media (max-width: 1200px) {
  .books-grid {
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  }
}

@media (max-width: 768px) {
  .books-container {
    padding: 20px 16px 0;
  }

  .dash-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 14px;
  }

  .stats {
    grid-template-columns: repeat(2, 1fr);
  }

  .books-grid {
    grid-template-columns: 1fr;
  }

  .card-body {
    flex-direction: row;
  }

  .book-cover {
    width: 92px;
  }

  .filters .sort {
    margin-left: 0;
  }

  .action-buttons {
    gap: 4px;
  }

  .action-btn {
    width: 26px;
    height: 26px;
    font-size: 12px;
  }
}

.cover-dialog-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 0;
}

.cover-drop-zone {
  width: 100%;
  aspect-ratio: 3 / 4;
  max-width: 280px;
  border: 2px dashed #dcdfe6;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s ease;
  overflow: hidden;
  background: #fafafa;
}

.cover-drop-zone:hover {
  border-color: #3b82f6;
  background: #f0f7ff;
}

.cover-drop-zone.is-dragover {
  border-color: #3b82f6;
  background: #e0edff;
  transform: scale(1.02);
}

.cover-drop-zone.has-image {
  border-style: solid;
  border-color: #dcdfe6;
  position: relative;
}

.cover-preview-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.cover-preview-zoom-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  border-radius: 50%;
  width: 32px;
  height: 32px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.cover-preview-zoom-btn:hover {
  background: rgba(0, 0, 0, 0.7);
}

.cover-drop-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #909399;
}

.cover-drop-icon {
  font-size: 48px;
  color: #c0c4cc;
}

.cover-drop-text {
  margin: 0;
  font-size: 14px;
  color: #606266;
}

.cover-drop-hint {
  margin: 0;
  font-size: 12px;
  color: #909399;
}

.cover-file-input {
  display: none;
}

/* 简约作品弹窗样式 */
.simple-book-dialog .el-dialog__body {
  padding: 20px 24px;
}

.simple-book-dialog .el-dialog__header {
  padding: 16px 24px;
  border-bottom: 1px solid #f0f0f0;
}

.simple-book-dialog .el-dialog__footer {
  padding: 12px 24px;
  border-top: 1px solid #f0f0f0;
}

.simple-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.simple-form .form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.simple-form .form-label {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
}

.simple-form .form-label.required::after {
  content: '*';
  color: #f56c6c;
  margin-left: 4px;
}

.simple-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* ========== 作品信息弹窗 · 报刊编辑风格 ========== */
.book-info-dialog :deep(.el-dialog) {
  width: 460px !important;
  max-width: calc(100vw - 32px);
  margin: 0 auto !important;
  border-radius: 2px;
  overflow: hidden;
  background-color: #f4f1ea;
  background-image:
    radial-gradient(circle at 25% 25%, rgba(0,0,0,0.015) 1px, transparent 1px),
    radial-gradient(circle at 75% 75%, rgba(0,0,0,0.015) 1px, transparent 1px);
  background-size: 4px 4px, 6px 6px;
  border: 1px solid #c8c4ba;
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.3);
  font-family: 'Lora', 'Noto Serif SC', Georgia, serif;
}

/* 报头 Header */
.book-info-dialog :deep(.el-dialog__header) {
  padding: 14px 22px 12px;
  border-bottom: 3px double #1a1a1a;
  margin-right: 0;
  position: relative;
}
.book-info-dialog :deep(.el-dialog__header)::after {
  content: '';
  position: absolute;
  bottom: -7px;
  left: 22px;
  right: 22px;
  height: 1px;
  background: #1a1a1a;
}
.book-info-dialog :deep(.el-dialog__headerbtn) {
  top: 16px;
  right: 18px;
  width: 28px;
  height: 28px;
  border: 1px solid #1a1a1a;
  border-radius: 2px;
  background: transparent;
  font-size: 16px;
  transition: all 0.15s;
}
.book-info-dialog :deep(.el-dialog__headerbtn:hover) {
  background: #c1272d;
  border-color: #c1272d;
}
.book-info-dialog :deep(.el-dialog__headerbtn .el-dialog__close) {
  color: #1a1a1a;
  font-family: 'Lora', serif;
}
.book-info-dialog :deep(.el-dialog__headerbtn:hover .el-dialog__close) {
  color: #f4f1ea;
}

/* Body */
.book-info-dialog :deep(.el-dialog__body) {
  padding: 16px 22px 18px;
  max-height: 60vh;
  overflow-y: auto;
}
.book-info-dialog :deep(.el-dialog__body)::-webkit-scrollbar {
  width: 8px;
}
.book-info-dialog :deep(.el-dialog__body)::-webkit-scrollbar-track {
  background: #ece7da;
}
.book-info-dialog :deep(.el-dialog__body)::-webkit-scrollbar-thumb {
  background: #6b6b6b;
  border: 2px solid #ece7da;
}

/* Footer */
.book-info-dialog :deep(.el-dialog__footer) {
  padding: 14px 22px;
  border-top: 3px double #1a1a1a;
  background: #ece7da;
}

.book-info-dialog :deep(.el-overlay-dialog) {
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ---- 报头内容 ---- */
.press-header {
  width: 100%;
}
.press-header-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 10px;
  letter-spacing: 0.15em;
  color: #6b6b6b;
  text-transform: uppercase;
  margin-bottom: 6px;
  font-family: 'Lora', serif;
}
.press-issue {
  color: #c1272d;
  font-weight: 600;
}
.press-header-title {
  font-family: 'Playfair Display', 'Noto Serif SC', serif;
  font-size: 26px;
  font-weight: 900;
  letter-spacing: 0.04em;
  color: #1a1a1a;
  line-height: 1;
}
.press-accent {
  color: #c1272d;
}

/* ---- Body 内容 ---- */
.press-body {
  color: #1a1a1a;
}

/* 顶部：封面 + 标题 */
.press-top-row {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  margin-bottom: 16px;
  padding-bottom: 14px;
  border-bottom: 1px solid #c8c4ba;
}

/* 封面：黑白处理 */
.press-cover {
  width: 64px;
  height: 80px;
  background: #1a1a1a;
  border: 1px solid #1a1a1a;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  cursor: pointer;
  position: relative;
  overflow: hidden;
}
.press-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  filter: grayscale(100%) contrast(1.1);
}
.press-cover-text {
  color: #f4f1ea;
  font-family: 'Playfair Display', 'Noto Serif SC', serif;
  font-size: 28px;
  font-weight: 900;
}

/* 标题输入框：头条标题 */
.press-title-block {
  flex: 1;
  min-width: 0;
}
.press-field-label {
  display: block;
  font-family: 'Lora', serif;
  font-size: 9px;
  font-weight: 600;
  letter-spacing: 0.2em;
  color: #6b6b6b;
  margin-bottom: 4px;
  text-transform: uppercase;
}
.press-title-input {
  width: 100%;
  background: transparent;
  border: none;
  border-bottom: 2px solid #1a1a1a;
  color: #1a1a1a;
  font-family: 'Playfair Display', 'Noto Serif SC', serif;
  font-size: 22px;
  font-weight: 900;
  line-height: 1.2;
  padding: 4px 0 6px;
  outline: none;
  transition: border-color 0.15s;
}
.press-title-input:focus {
  border-bottom-color: #c1272d;
}
.press-title-input::placeholder {
  color: #c8c4ba;
  font-weight: 400;
}

/* ---- Metadata 版权栏 ---- */
.press-meta-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0;
  margin-bottom: 16px;
  padding: 10px 0;
  border-bottom: 1px solid #c8c4ba;
  font-family: 'Lora', serif;
}
.press-meta-item {
  display: flex;
  align-items: baseline;
  gap: 6px;
  font-size: 11px;
  padding: 3px 0;
  border-bottom: 1px dotted #c8c4ba;
}
.press-meta-item:nth-last-child(-n+2) {
  border-bottom: none;
}
.press-meta-item:nth-child(odd) {
  padding-right: 10px;
}
.press-meta-item:nth-child(even) {
  padding-left: 10px;
  border-left: 1px dotted #c8c4ba;
}
.press-k {
  color: #6b6b6b;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-size: 9px;
  flex-shrink: 0;
}
.press-v {
  color: #1a1a1a;
  font-weight: 600;
  margin-left: auto;
  text-align: right;
}

/* ---- 章节标题 ---- */
.press-section {
  margin-top: 14px;
}
.press-section-label {
  font-family: 'Playfair Display', 'Noto Serif SC', serif;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.2em;
  color: #c1272d;
  text-transform: uppercase;
  padding-bottom: 4px;
  margin-bottom: 8px;
  border-bottom: 1px solid #1a1a1a;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.press-section-label::after {
  content: '◆';
  color: #1a1a1a;
  font-size: 8px;
}

/* ---- 简介正文 ---- */
.press-textarea {
  width: 100%;
  background: transparent;
  border: none;
  border-top: 1px solid transparent;
  color: #1a1a1a;
  font-family: 'Lora', 'Noto Serif SC', serif;
  font-size: 13px;
  line-height: 1.8;
  padding: 8px 0;
  outline: none;
  resize: vertical;
  min-height: 80px;
  text-align: justify;
  transition: border-color 0.15s;
}
.press-textarea:focus {
  border-top-color: #c1272d;
}
.press-textarea::placeholder {
  color: #c8c4ba;
  font-style: italic;
}

/* ---- 目标字数 + 进度条 ---- */
.press-progress-section {
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid #c8c4ba;
}
.press-progress-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 10px;
}
.press-progress-label {
  font-family: 'Playfair Display', 'Noto Serif SC', serif;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.2em;
  color: #1a1a1a;
  text-transform: uppercase;
}
.press-pct {
  font-family: 'Playfair Display', serif;
  font-size: 22px;
  font-weight: 900;
  color: #c1272d;
  line-height: 1;
}

/* 进度条：油墨条 */
.press-progress-track {
  height: 10px;
  background: #e6e1d6;
  border: 1px solid #1a1a1a;
  margin-bottom: 8px;
  position: relative;
  overflow: hidden;
}
.press-progress-fill {
  height: 100%;
  background: repeating-linear-gradient(
    90deg,
    #1a1a1a 0, #1a1a1a 2px,
    #2a2a2a 2px, #2a2a2a 4px
  );
  position: relative;
  transition: width 0.4s ease;
}
.press-progress-fill::after {
  content: '';
  position: absolute;
  right: 0;
  top: 0;
  width: 2px;
  height: 100%;
  background: #c1272d;
}

.press-word-counts {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-family: 'Lora', serif;
  font-size: 11px;
  color: #6b6b6b;
  margin-bottom: 10px;
  font-style: italic;
}
.press-word-counts b {
  color: #1a1a1a;
  font-weight: 700;
  font-style: normal;
}

/* 目标字数编辑 */
.press-target-edit {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-top: 8px;
  border-top: 1px dotted #c8c4ba;
  font-family: 'Lora', serif;
  font-size: 10px;
  letter-spacing: 0.1em;
  color: #6b6b6b;
  text-transform: uppercase;
}
.press-target-input {
  width: 100px;
  background: transparent;
  border: none;
  border-bottom: 1px solid #1a1a1a;
  color: #1a1a1a;
  font-family: 'Playfair Display', serif;
  font-size: 14px;
  font-weight: 700;
  padding: 2px 0;
  outline: none;
  text-align: right;
  transition: border-color 0.15s;
}
.press-target-input:focus {
  border-bottom-color: #c1272d;
}
.press-target-input::placeholder {
  color: #c8c4ba;
  font-weight: 400;
}
.press-target-save {
  background: transparent;
  border: 1px solid #1a1a1a;
  color: #1a1a1a;
  font-family: 'Lora', serif;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  padding: 2px 10px;
  cursor: pointer;
  border-radius: 2px;
  transition: all 0.15s;
}
.press-target-save:hover {
  background: #1a1a1a;
  color: #f4f1ea;
}

/* ---- 标签 ---- */
.press-tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 8px;
}
.press-tag {
  font-family: 'Playfair Display', 'Noto Serif SC', serif;
  font-size: 12px;
  font-weight: 700;
  color: #1a1a1a;
  letter-spacing: 0.04em;
}
.press-tag::before {
  content: '[';
  color: #c1272d;
  margin-right: 1px;
}
.press-tag::after {
  content: ']';
  color: #c1272d;
  margin-left: 1px;
}
.press-empty {
  color: #c8c4ba;
  font-size: 12px;
  font-style: italic;
  font-family: 'Lora', serif;
}

/* ---- Footer ---- */
.press-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}
.press-footer-note {
  font-family: 'Lora', serif;
  font-size: 10px;
  font-style: italic;
  color: #6b6b6b;
  letter-spacing: 0.05em;
}
.press-save-btn {
  background: #1a1a1a;
  color: #f4f1ea;
  border: 1px solid #1a1a1a;
  font-family: 'Playfair Display', 'Noto Serif SC', serif;
  font-weight: 700;
  font-size: 13px;
  letter-spacing: 0.25em;
  text-transform: uppercase;
  padding: 10px 28px;
  cursor: pointer;
  border-radius: 2px;
  transition: all 0.15s;
}
.press-save-btn::before {
  content: '▶ ';
  color: #c1272d;
  font-size: 9px;
}
.press-save-btn:hover {
  background: #c1272d;
  border-color: #c1272d;
}

/* ---- 响应式 ---- */
@media (max-width: 600px) {
  .book-info-dialog :deep(.el-dialog) {
    width: calc(100vw - 32px) !important;
  }
  .press-top-row {
    flex-direction: column;
    align-items: flex-start;
  }
  .press-meta-grid {
    grid-template-columns: 1fr;
  }
  .press-meta-item:nth-child(even) {
    padding-left: 0;
    border-left: none;
  }
}

/* ========== 暗色主题适配 ========== */
:root[data-theme='dark'] .books-container {
  --c-bg: #0d1117;
  --c-card: #161b22;
  --c-border: #21262d;
  --c-text: #e6edf3;
  --c-sub: #8b98a8;
  --c-muted: #4d5768;
  --c-line: #2a3042;
  --c-accent: #6c8cff;
  --c-accent2: #9c6cff;
}

:root[data-theme='dark'] .book-card:hover {
  border-color: #3b4a6b;
  box-shadow: 0 12px 30px rgba(0, 0, 0, .35);
}

:root[data-theme='dark'] .book-cover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, .4);
}

:root[data-theme='dark'] .stat:hover {
  box-shadow: 0 10px 26px rgba(0, 0, 0, .3);
}

:root[data-theme='dark'] .stat.active {
  box-shadow: 0 0 0 2px rgba(108, 140, 255, .25), 0 10px 26px rgba(0, 0, 0, .25);
}

:root[data-theme='dark'] .seg span.active {
  box-shadow: 0 2px 8px rgba(0, 0, 0, .4);
}

:root[data-theme='dark'] .empty-state {
  background: #161b22;
}

:root[data-theme='dark'] .chip.cat {
  background: #1e2a45;
  color: #8fa7d6;
}

:root[data-theme='dark'] .chip.on {
  background: #143f31;
  color: #4ade80;
}

:root[data-theme='dark'] .chip.fin {
  background: #1e3050;
  color: #8fb6f0;
}

:root[data-theme='dark'] .dash-btn.el-button {
  box-shadow: 0 2px 8px rgba(0, 0, 0, .2);
}

:root[data-theme='dark'] .cover-placeholder {
  background: linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%);
}

:root[data-theme='dark'] .cover-tag {
  background: rgba(0, 0, 0, 0.8);
}

:root[data-theme='dark'] .form-section {
  background: #161b22;
}

:root[data-theme='dark'] .section-title {
  color: #e6edf3;
}

:root[data-theme='dark'] .form-label {
  color: #9ca3af;
}

:root[data-theme='dark'] .dialog-content {
  background: transparent;
}

:root[data-theme='dark'] .cover-drop-zone {
  border-color: #3f4a5c;
  background: #161b22;
}

:root[data-theme='dark'] .cover-drop-zone:hover {
  border-color: #3b82f6;
  background: #1e3a5f;
}

:root[data-theme='dark'] .cover-drop-zone.is-dragover {
  border-color: #3b82f6;
  background: #1e3a5f;
}

:root[data-theme='dark'] .cover-drop-zone.has-image {
  border-color: #3f4a5c;
}

:root[data-theme='dark'] .cover-drop-placeholder {
  color: #4d5768;
}

:root[data-theme='dark'] .cover-drop-icon {
  color: #4b5563;
}

:root[data-theme='dark'] .cover-drop-text {
  color: #9ca3af;
}

:root[data-theme='dark'] .cover-drop-hint {
  color: #4d5768;
}

:root[data-theme='dark'] .simple-book-dialog .el-dialog__header {
  border-bottom-color: #21262d;
}

:root[data-theme='dark'] .simple-book-dialog .el-dialog__footer {
  border-top-color: #21262d;
}

:root[data-theme='dark'] .simple-form .form-label {
  color: #9ca3af;
}

.ai-cover-form {
  padding: 8px 12px;
}

.ai-cover-form :deep(.el-form-item) {
  margin-bottom: 12px;
}

.ai-cover-form :deep(.el-form-item__label) {
  line-height: 1.5;
  padding-bottom: 4px;
}

.reference-image-panel {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  width: 100%;
}

.reference-image-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.reference-image-hint {
  font-size: 12px;
  color: #909399;
}

.reference-image-preview {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ebeef5;
  border-radius: 10px;
  background: #fafafa;
}

.reference-image-thumb {
  width: 64px;
  height: 64px;
  object-fit: cover;
  border-radius: 8px;
  flex-shrink: 0;
}

.reference-image-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  flex: 1;
}

.reference-image-name {
  font-size: 13px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reference-image-size {
  font-size: 12px;
  color: #909399;
}

.rank-reference-layout {
  display: grid;
  grid-template-columns: 116px 1fr;
  gap: 10px;
  height: 100%;
  min-height: 0;
}

.rank-reference-groups {
  display: flex;
  flex-direction: column;
  gap: 6px;
  border-right: 1px solid #ebeef5;
  padding-right: 8px;
  min-height: 0;
  overflow-y: auto;
}

.rank-reference-group-btn,
.rank-reference-category-btn,
.rank-reference-card {
  border: none;
  background: transparent;
  cursor: pointer;
}

.rank-reference-group-btn {
  text-align: left;
  padding: 8px 10px;
  border-radius: 8px;
  color: #606266;
  font-size: 12px;
}

.rank-reference-group-btn.active {
  background: #ecf5ff;
  color: #409eff;
  font-weight: 600;
}

.rank-reference-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
}

.rank-reference-categories {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  flex-shrink: 0;
  max-height: 72px;
  overflow-y: auto;
  align-content: flex-start;
}

.rank-reference-category-btn {
  padding: 4px 10px;
  border-radius: 999px;
  background: #f5f7fa;
  color: #606266;
  font-size: 12px;
}

.rank-reference-category-btn.active {
  background: #409eff;
  color: #fff;
}

.rank-reference-state {
  flex: 1;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.rank-reference-grid-wrap {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding-right: 6px;
}

.rank-reference-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));
  gap: 10px;
  align-content: start;
}

.rank-reference-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px;
  border: 1px solid #ebeef5;
  border-radius: 12px;
  text-align: left;
  transition: border-color 0.2s ease, transform 0.2s ease, box-shadow 0.2s ease;
  min-height: 214px;
  height: 214px;
  overflow: hidden;
}

.rank-reference-card:hover {
  border-color: #409eff;
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(64, 158, 255, 0.12);
}

.rank-reference-cover {
  width: 100%;
  height: 124px;
  object-fit: cover;
  border-radius: 8px;
  background: #f5f7fa;
  flex-shrink: 0;
}

.rank-reference-book-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  flex: 1;
  overflow: hidden;
}

.rank-reference-rank {
  font-size: 11px;
  color: #f56c6c;
  font-weight: 600;
}

.rank-reference-title {
  font-size: 12px;
  color: #303133;
  line-height: 1.45;
  min-height: 34px;
  max-height: 34px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.rank-reference-author {
  font-size: 11px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rank-reference-dialog :deep(.el-dialog) {
  width: min(880px, calc(100vw - 32px)) !important;
  max-width: calc(100vw - 32px);
  height: min(360px, calc(100vh - 32px));
  max-height: calc(100vh - 32px);
  margin: 0 auto !important;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.rank-reference-dialog :deep(.el-dialog__header) {
  flex-shrink: 0;
}

.rank-reference-dialog :deep(.el-dialog__body) {
  flex: 1;
  min-height: 0;
  padding: 12px;
  overflow: hidden;
}

.rank-reference-dialog :deep(.el-overlay-dialog) {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  padding: 16px;
  box-sizing: border-box;
}

.field-label-text {
  margin-right: 8px;
}

.field-required-tag {
  margin-left: 4px;
  transform: scale(0.85);
}

.cover-dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.cover-dialog-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.generation-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.generate-btn {
  width: 100%;
}

.generate-btn .el-icon {
  margin-right: 6px;
}

.cover-tabs {
  margin: -10px 0;
}

.cover-tabs :deep(.el-tabs__header) {
  margin-bottom: 10px;
}

.cover-tabs :deep(.el-tabs__item) {
  height: 36px;
  line-height: 36px;
}

.cover-dialog :deep(.el-dialog) {
  width: 900px !important;
  max-width: 96vw;
  margin-top: 6vh !important;
}

.cover-dialog :deep(.el-dialog__body) {
  padding: 0;
  height: min(640px, 82vh);
  overflow: hidden;
}

.cover-field-select {
  width: 100%;
}

:deep(.cover-field-select-popper .el-select-dropdown__wrap),
:deep(.cover-field-select-popper .el-scrollbar__wrap) {
  max-height: 320px;
}

:deep(.cover-field-select-popper .el-scrollbar__bar.is-vertical) {
  opacity: 1;
}

.cover-preview-dialog .cover-preview-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.cover-preview-dialog .cover-preview-header {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.cover-preview-dialog .cover-preview-image-container {
  width: 100%;
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  border-radius: 8px;
  overflow: hidden;
}

.cover-preview-dialog .cover-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  color: #909399;
}

.cover-preview-dialog .cover-loading .loading-icon {
  font-size: 48px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.cover-preview-dialog .cover-preview-image {
  max-width: 100%;
  max-height: 500px;
  object-fit: contain;
}

.cover-preview-dialog .cover-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #c0c4cc;
}

.cover-preview-dialog .cover-empty .el-icon {
  font-size: 64px;
}

/* ================= 封面弹窗 · 综合重构（01 分屏 + 09 对比 + 03 历史画廊） ================= */
.cover-blend {
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  height: 100%;
}

.blend-left {
  background: linear-gradient(180deg, #eef1fb, #f8f6ff);
  border-right: 1px solid var(--c-line);
  padding: 20px 22px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  overflow-y: auto;
  min-height: 0;
}

.blend-label {
  font-size: 11px;
  color: var(--c-muted);
  font-weight: 600;
  letter-spacing: 1px;
  align-self: flex-start;
}

.blend-stage {
  width: 230px;
  aspect-ratio: 3 / 4;
  border-radius: 14px;
  overflow: hidden;
  position: relative;
  box-shadow: 0 12px 28px rgba(60, 80, 130, .18);
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: box-shadow .2s;
  flex-shrink: 0;
}

.blend-stage.dragover {
  outline: 3px dashed var(--c-accent);
  outline-offset: 4px;
  box-shadow: 0 0 0 8px rgba(108, 140, 255, .12);
}

.blend-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.blend-ph {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: var(--c-muted);
  text-align: center;
  padding: 0 16px;
  cursor: pointer;
}

.blend-ph .ic {
  width: 52px;
  height: 52px;
  border-radius: 16px;
  background: linear-gradient(135deg, var(--c-accent), var(--c-accent2));
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  box-shadow: 0 8px 18px rgba(108, 140, 255, .35);
}

.blend-ph b {
  font-size: 12.5px;
  color: var(--c-sub);
  font-weight: 600;
}

.blend-ph small {
  font-size: 11px;
}

.blend-zoom {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  border: none;
  background: rgba(0, 0, 0, .45);
  color: #fff;
  cursor: pointer;
  font-size: 13px;
  opacity: 0;
  transition: .2s;
}

.blend-stage:hover .blend-zoom {
  opacity: 1;
}

.blend-gen {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: var(--c-muted);
  font-size: 13px;
  padding: 0 16px;
  text-align: center;
}

.blend-gen .spin-icon {
  font-size: 42px;
  animation: spin 1s linear infinite;
}

/* 前后对比器（09） */
.blend-cmp {
  width: 230px;
  aspect-ratio: 3 / 4;
  border-radius: 14px;
  overflow: hidden;
  position: relative;
  box-shadow: 0 12px 28px rgba(60, 80, 130, .2);
  user-select: none;
  touch-action: none;
  background: #fff;
  flex-shrink: 0;
}

.blend-cmp img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  pointer-events: none;
}

.blend-cmp .cmp-back { z-index: 1; }
.blend-cmp .cmp-front { z-index: 2; }

.blend-cmp .cmp-handle {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 3px;
  background: #fff;
  transform: translateX(-50%);
  z-index: 3;
  cursor: ew-resize;
  box-shadow: 0 0 0 1px rgba(0, 0, 0, .15);
}

.blend-cmp .cmp-handle::after {
  content: "";
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 6px 18px rgba(0, 0, 0, .35);
}

.blend-cmp .cmp-handle::before {
  content: "⇔";
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 1;
  color: var(--c-accent);
  font-weight: 900;
  font-size: 13px;
  line-height: 36px;
}

.blend-cmp .lbl {
  position: absolute;
  top: 10px;
  font-size: 10px;
  font-weight: 800;
  color: #fff;
  background: rgba(0, 0, 0, .5);
  padding: 3px 9px;
  border-radius: 999px;
  z-index: 4;
  pointer-events: none;
}

.blend-cmp .lbl.old { left: 10px; }
.blend-cmp .lbl.new {
  right: 10px;
  background: linear-gradient(135deg, var(--c-accent), var(--c-accent2));
}

.blend-cmp-empty {
  width: 230px;
  aspect-ratio: 3 / 4;
  border: 2px dashed var(--c-border);
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--c-muted);
  font-size: 12px;
  background: #fff;
  text-align: center;
  padding: 0 16px;
  flex-shrink: 0;
}

.view-seg {
  display: flex;
  gap: 4px;
  background: rgba(233, 235, 246, .85);
  padding: 3px;
  border-radius: 10px;
  width: 230px;
  flex-shrink: 0;
}

.view-seg button {
  flex: 1;
  border: none;
  background: transparent;
  padding: 6px 0;
  border-radius: 8px;
  font-size: 11.5px;
  font-weight: 700;
  color: var(--c-sub);
  cursor: pointer;
  transition: .15s;
  font-family: inherit;
}

.view-seg button.on {
  background: #fff;
  color: var(--c-accent);
  box-shadow: 0 2px 6px rgba(60, 80, 130, .14);
}

/* 右侧面板 */
.blend-right {
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.blend-right .seg {
  display: flex;
  gap: 4px;
  background: var(--c-line);
  padding: 4px;
  border-radius: 12px;
  margin: 16px 20px 4px;
  flex-shrink: 0;
}

.blend-right .seg button {
  flex: 1;
  border: none;
  background: transparent;
  padding: 8px 0;
  border-radius: 9px;
  font-size: 12.5px;
  font-weight: 700;
  color: var(--c-sub);
  cursor: pointer;
  transition: .15s;
  font-family: inherit;
}

.blend-right .seg button.on {
  background: var(--c-card);
  color: var(--c-accent);
  box-shadow: 0 2px 8px rgba(60, 80, 130, .12);
}

.blend-form {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 12px 20px 18px;
  display: flex;
  flex-direction: column;
  gap: 13px;
}

.blend-form::-webkit-scrollbar {
  width: 5px;
}

.blend-form::-webkit-scrollbar-thumb {
  background: #d5d9e8;
  border-radius: 3px;
}

.blend-form .field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.blend-form .field > label {
  font-size: 12px;
  font-weight: 600;
  color: var(--c-sub);
  display: flex;
  align-items: center;
  gap: 6px;
  line-height: 1.6;
}

.blend-form .field .req { color: #ef5350; }
.blend-form .field .opt-hint { font-size: 10px; color: var(--c-muted); font-weight: 400; }
.blend-form .field .ctrl-select { width: 100%; }

.blend-form .ctrl {
  padding: 9px 12px;
  border: 1px solid var(--c-border);
  border-radius: 10px;
  font-size: 13px;
  font-family: inherit;
  color: var(--c-text);
  background: #fbfcfe;
  outline: none;
  transition: .15s;
  text-align: left;
  cursor: pointer;
}

.blend-form .ctrl:hover { border-color: var(--c-accent); }

.form-top {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  min-height: 20px;
  margin-bottom: -4px;
}

.format-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.format-chips span {
  font-size: 11px;
  color: var(--c-muted);
  background: var(--c-line);
  padding: 5px 11px;
  border-radius: 8px;
}

.hist-count {
  font-size: 10.5px;
  color: var(--c-muted);
  font-weight: 400;
  background: var(--c-line);
  padding: 1px 8px;
  border-radius: 999px;
}

.hist-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}

.hist-tile {
  position: relative;
  border: 2px solid var(--c-border);
  border-radius: 10px;
  overflow: hidden;
  cursor: pointer;
  padding: 0;
  background: var(--c-card);
  transition: .15s;
}

.hist-tile img {
  width: 100%;
  aspect-ratio: 3 / 4;
  object-fit: cover;
  display: block;
}

.hist-tile:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 14px rgba(60, 80, 130, .15);
}

.hist-tile.sel {
  border-color: var(--c-accent);
  box-shadow: 0 0 0 2px rgba(108, 140, 255, .22);
}

.hist-tile .del {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 18px;
  height: 18px;
  border-radius: 6px;
  border: none;
  background: rgba(0, 0, 0, .5);
  color: #fff;
  font-size: 9px;
  line-height: 18px;
  text-align: center;
  cursor: pointer;
  opacity: 0;
  transition: .15s;
}

.hist-tile:hover .del { opacity: 1; }
.hist-tile .del:hover { background: rgba(239, 83, 80, .9); }

.hist-tile .ck {
  position: absolute;
  left: 4px;
  bottom: 4px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--c-accent), var(--c-accent2));
  color: #fff;
  display: none;
  align-items: center;
  justify-content: center;
  font-size: 9px;
}

.hist-tile.sel .ck { display: flex; }

.hist-empty,
.upload-tip {
  font-size: 11.5px;
  color: var(--c-muted);
  line-height: 1.7;
  border: 1px dashed var(--c-border);
  border-radius: 10px;
  padding: 12px;
  text-align: center;
  background: #fbfcfe;
}

.upload-tip b {
  display: block;
  font-size: 13px;
  color: var(--c-sub);
  margin-bottom: 3px;
}

.gen-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  width: 100%;
  padding: 12px 0;
  border: none;
  border-radius: 12px;
  background: linear-gradient(135deg, var(--c-accent), var(--c-accent2));
  color: #fff;
  font-size: 13.5px;
  font-weight: 800;
  cursor: pointer;
  font-family: inherit;
  box-shadow: 0 8px 20px rgba(108, 140, 255, .35);
  transition: .18s;
}

.gen-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 12px 26px rgba(108, 140, 255, .45);
}

.gen-btn:disabled {
  opacity: .55;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.gen-btn .spin {
  animation: spin 1s linear infinite;
}

.field-required-tag {
  margin-left: 4px;
  transform: scale(.85);
}

/* 暗色适配 */
:root[data-theme='dark'] .blend-left {
  background: linear-gradient(180deg, #161b22, #1a2230);
}

:root[data-theme='dark'] .blend-stage,
:root[data-theme='dark'] .blend-cmp,
:root[data-theme='dark'] .blend-cmp-empty {
  background: #161b22;
}

:root[data-theme='dark'] .view-seg { background: rgba(255, 255, 255, .06); }
:root[data-theme='dark'] .view-seg button.on { background: #21262d; }
:root[data-theme='dark'] .hist-empty,
:root[data-theme='dark'] .upload-tip { background: #161b22; }
:root[data-theme='dark'] .reference-image-preview { background: #161b22; border-color: #21262d; }
</style>
