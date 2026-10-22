<template>
  <div class="experience-page">
    <div class="page">
      <!-- 横幅吧头：头像/吧名/数据压图 + 管理、发帖按钮 -->
      <header class="bar-banner">
        <div class="bb-bottom">
          <div class="bb-av">{{ boardConf.short }}</div>
          <div class="bb-name">
            <h1>{{ boardConf.name }}</h1>
            <p>{{ boardConf.desc }} · 收录 {{ countAll }} 篇 · 回复 {{ totalReplies }} · 精华 {{ countFeatured }}</p>
          </div>
          <div class="bb-actions">
            <button class="bb-btn ghost" type="button" @click="manageVisible = true">管理</button>
            <el-dropdown trigger="click" @command="handleCreateCommand">
              <button class="bb-btn primary" type="button">＋ 发帖</button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="manual">
                    <el-icon><EditPen /></el-icon>
                    手动创建
                  </el-dropdown-item>
                  <el-dropdown-item command="pdf_import">
                    <el-icon><UploadFilled /></el-icon>
                    PDF 导入创建
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </header>

      <!-- 标签页行：下划线分区 tabs + 排序 + 搜索 -->
      <nav class="tabs-row">
        <div class="tabs">
          <button
            v-for="t in sectionTabs"
            :key="String(t.key)"
            class="tab"
            :class="{ on: activeSection === t.key }"
            @click="setSectionFilter(t.key)"
          >
            {{ t.label }}<span class="tn">{{ t.count }}</span>
          </button>
        </div>
        <div class="tr-right">
          <div class="sorts">
            <button class="sort" :class="{ on: sortBy === 'replies' }" @click="sortBy = 'replies'">最多回复</button>
            <button class="sort" :class="{ on: sortBy === 'likes' }" @click="sortBy = 'likes'">最多点赞</button>
            <button class="sort" :class="{ on: sortBy === 'time' }" @click="sortBy = 'time'">最新</button>
          </div>
          <el-input v-model="searchQuery" placeholder="搜索标题、摘要或作者…" clearable class="tr-search">
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
      </nav>

      <div class="cols">
        <!-- 左侧帖子流 -->
        <main v-loading="loading" class="feed">
          <!-- 置顶区：贴吧那种独立区块 —— 蓝框「置顶」小标签 + 纯标题一行一条。
               ⚠ 这些是真帖子（is_pinned=1），点一行进详情看正文；
               主列表会把它们排除掉，避免同一篇出现两次（贴吧也是这个口径）。 -->
          <div v-if="pinnedShares.length > 0" class="pin-box">
            <div
              v-for="item in pinnedShares"
              :key="item.id"
              class="pin-row"
              role="link"
              tabindex="0"
              :title="item.title"
              @click="openDetail(item)"
              @keydown.enter="openDetail(item)"
            >
              <span class="pin-tag">置顶</span>
              <span class="pin-text">{{ item.title || '未命名' }}</span>
            </div>
          </div>

          <template v-if="experienceShares.length > 0">
            <template v-if="paginatedExperienceShares.length > 0">
              <article
                v-for="item in paginatedExperienceShares"
                :key="item.id"
                class="card post"
                :class="{ 'is-selected': selectedIds.includes(item.id) }"
                @click="handleCardClick(item)"
              >
                <span v-if="selectionMode" class="p-check" @click.stop="handleCardClick(item)">
                  <el-icon v-if="selectedIds.includes(item.id)"><Check /></el-icon>
                </span>
                <div class="p-head">
                  <div class="p-av" :style="{ background: avatarGradient(item) }">
                    {{ (item.author_name || '星')[0] }}
                  </div>
                  <div class="p-who">
                    <div class="p-name">{{ item.author_name || '星芒用户' }}</div>
                    <div class="p-date">{{ formatListDate(item.created_at) }}</div>
                  </div>
                  <el-dropdown
                    v-if="sections.length > 0"
                    trigger="click"
                    @command="(sid: number | null) => moveToSection(item, sid)"
                  >
                    <span class="p-sec" :class="{ empty: !item.section_id }" @click.stop>
                      {{ sectionName(item.section_id) || '未归类' }}
                      <el-icon class="sec-caret"><ArrowDown /></el-icon>
                    </span>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item :command="null">未归类</el-dropdown-item>
                        <el-dropdown-item v-for="s in sections" :key="s.id" :command="s.id">
                          {{ s.name }}
                        </el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </div>
                <div class="p-title">
                  <span v-if="item.is_featured === 1" class="badge bd-feat">精华</span>
                  <span class="badge" :class="item.create_type === 'pdf_import' ? 'bd-pdf' : 'bd-manual'">
                    {{ item.create_type === 'pdf_import' ? 'PDF' : '心得' }}
                  </span>
                  <span class="p-title-text">{{ item.title || '未命名' }}</span>
                </div>
                <div v-if="postExcerpt(item)" class="p-content">
                  {{ postExcerpt(item) }}<a
                    v-if="excerptTruncated(item)"
                    class="p-full"
                    @click.stop="openDetail(item)"
                  >...全文</a>
                </div>
                <div v-if="hasUsableCover(item)" class="p-media single">
                  <img
                    class="ph"
                    :src="coverSrc(item) || ''"
                    :alt="item.title"
                    loading="lazy"
                    @click.stop="openDetail(item)"
                    @error="markCoverBroken(item.id)"
                  />
                </div>
                <div class="p-acts" @click.stop>
                  <button class="pa" type="button" @click="sharePost(item)">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 12v7a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-7" /><path d="M16 6l-4-4-4 4" /><path d="M12 2v13" /></svg>分享
                  </button>
                  <button class="pa" type="button" @click="openDetail(item)">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a8 8 0 0 1-8 8H4l2-3a8 8 0 1 1 15-5z" /></svg>{{ replyCount(item.id) }}
                  </button>
                  <button
                    class="pa pa-like"
                    :class="{ liked: isLiked(item.id) }"
                    type="button"
                    :title="isLiked(item.id) ? '取消赞' : '点赞'"
                    @click="toggleLike(item)"
                  >
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M7 10v12H4V10h3zm3 12h7a3 3 0 0 0 3-3l-1-6a2 2 0 0 0-2-2h-4l1-4a2 2 0 0 0-2-2l-4 8v9z" /></svg>{{ likeTotal(item) }}
                  </button>
                  <span class="pa-manage">
                    <button class="pa" type="button" @click.stop="toggleManage(item.id)">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3" /><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" /></svg>管理
                    </button>
                    <span class="pa-pop" :class="{ open: openManageId === item.id }">
                      <button class="po pin" :class="{ on: item.is_pinned === 1 }" @click.stop="togglePinned(item); openManageId = null">
                        {{ item.is_pinned === 1 ? '⌃ 取消置顶' : '⌃ 置顶' }}
                      </button>
                      <button class="po feat" :class="{ on: item.is_featured === 1 }" @click.stop="toggleFeatured(item); openManageId = null">
                        {{ item.is_featured === 1 ? '★ 取消精华' : '★ 加精' }}
                      </button>
                      <button class="po" @click.stop="openEdit(item); openManageId = null">编辑</button>
                      <button class="po" @click.stop="handleExportSingle(item); openManageId = null">导出</button>
                      <button class="po danger" @click.stop="handleDelete(item); openManageId = null">删除</button>
                    </span>
                  </span>
                </div>
              </article>
            </template>
            <div v-else class="t-filtered-empty">
              <el-icon class="fe-ic"><Collection /></el-icon>
              <p class="fe-title">没有匹配的经验卡片</p>
              <p class="fe-sub">换个关键词或分区试试</p>
              <el-button size="small" @click="resetFilter">清除筛选</el-button>
            </div>
          </template>

          <el-empty v-else description="还没有经验卡片">
            <template #image>
              <div class="empty-illustration">
                <el-icon><Collection /></el-icon>
              </div>
            </template>
            <el-button type="primary" @click="openCreateDialog('manual')">手动创建第一张卡片</el-button>
            <el-button style="margin-left: 12px" @click="openCreateDialog('pdf_import')">通过 PDF 导入</el-button>
          </el-empty>

          <!-- 底栏：分页 / 批量操作 -->
          <footer class="list-foot">
            <div class="stats-left">
              <template v-if="selectionMode">
                <el-checkbox
                  :model-value="selectAll"
                  :indeterminate="isIndeterminate"
                  @change="handleSelectAllChange"
                >
                  全选
                </el-checkbox>
                <span class="selection-count">已选 {{ selectedIds.length }} 项</span>
              </template>
              <span v-else class="stat-title">{{ statTitle }}</span>
            </div>
            <div class="stats-right">
              <template v-if="selectionMode">
                <el-popover v-model:visible="batchMoveVisible" :width="260" trigger="click" placement="top-end">
                  <template #reference>
                    <el-button size="small" :disabled="selectedIds.length === 0">移动到分区</el-button>
                  </template>
                  <div class="batch-move">
                    <div class="bm-title">把已选 {{ selectedIds.length }} 篇移动到</div>
                    <el-select v-model="batchTargetSection" size="small" placeholder="选择分区" clearable>
                      <el-option label="未归类" :value="null" />
                      <el-option v-for="s in sections" :key="s.id" :label="s.name" :value="s.id" />
                    </el-select>
                    <div class="bm-ops">
                      <el-button size="small" @click="batchMoveVisible = false">取消</el-button>
                      <el-button size="small" type="primary" @click="submitBatchMove">确定</el-button>
                    </div>
                  </div>
                </el-popover>
                <el-button size="small" type="primary" :disabled="selectedIds.length === 0" @click="handleExportSelected">
                  <el-icon><Download /></el-icon>
                  导出选中
                </el-button>
                <el-button size="small" @click="cancelSelection">取消选择</el-button>
              </template>
              <template v-else>
                <el-pagination
                  v-if="totalPages > 1"
                  small
                  layout="prev, pager, next"
                  :total="totalItems"
                  :page-size="pageSize"
                  v-model:current-page="currentPage"
                  @current-change="handlePageChange"
                />
                <span class="stats-hint">共 {{ totalItems }} 篇 · 每页 {{ pageSize }} 篇</span>
              </template>
            </div>
          </footer>
        </main>

        <!-- 右栏单容器：吧信息 + 相关推荐 共同滚动，管理在横幅「管理」弹窗 -->
        <aside class="rail">
          <section class="card side-all">
            <div class="side-scroll">
              <div class="bar-info">
                <div class="sec-head"><h3>吧信息</h3></div>
                <div class="bi-row">
                  <div class="bi-av">{{ boardConf.short }}</div>
                  <div>
                    <div class="bi-name">{{ boardConf.name }}</div>
                    <div class="bi-sub">本地数据 · 支持导出帖子内容</div>
                  </div>
                </div>
                <div class="bi-stats">
                  <div><b>{{ countAll }}</b><span>帖子</span></div>
                  <div><b>{{ totalReplies }}</b><span>回复</span></div>
                  <div><b>{{ countFeatured }}</b><span>精华</span></div>
                  <div><b>{{ countPdf }}</b><span>PDF</span></div>
                </div>
                <p class="bi-desc">
                  沉淀案例方法、实操心得和可复用的经验卡片。精华帖由吧主人工挑选，自定义分区上限
                  {{ SECTION_LIMIT }} 个。
                </p>
              </div>

              <div v-if="!relBlockHidden && relatedShares.length" class="rel">
                <div class="rel-head">
                  <h3>相关推荐</h3>
                  <button class="rel-close" title="收起推荐" @click="relBlockHidden = true">✕</button>
                </div>
                <a v-for="item in relatedShares" :key="item.id" class="rl-item" @click.stop="openDetail(item)">
                  <div class="rl-t">{{ item.title || '未命名' }}</div>
                  <div v-if="postExcerpt(item)" class="rl-x">{{ postExcerpt(item) }}</div>
                  <div v-if="hasUsableCover(item)" class="rl-img">
                    <img :src="coverSrc(item) || ''" :alt="item.title" loading="lazy" @error="markCoverBroken(item.id)" />
                  </div>
                  <div class="rl-foot">
                    <span class="rl-chip">
                      <i :style="{ background: avatarGradient(item) }"></i>
                      <span>{{ sectionName(item.section_id) || '未归类' }}</span>
                    </span>
                    <span class="rl-nums">{{ replyCount(item.id) }}评论 {{ likeTotal(item) }}点赞</span>
                  </div>
                </a>
              </div>
            </div>
          </section>
        </aside>
      </div>
    </div>

    <!-- 管理弹窗（横幅「管理」按钮唤起） -->
    <el-dialog v-model="manageVisible" title="管理" width="min(420px, 92vw)" top="14vh" append-to-body>
      <div class="mo-body">
        <button class="mg" type="button" @click="handleImport">
          📥 导入 JSON
        </button>
        <button class="mg" type="button" @click="handleExport">
          📤 导出全部<span class="cnt">{{ countAll }} 篇</span>
        </button>
        <button class="mg" type="button" @click="manageVisible = false; toggleSelectionMode()">
          ☑️ 批量选择 / 移动分区
        </button>
        <button class="mg" type="button" @click="manageVisible = false; openSectionManager()">
          🏷️ 管理分区<span class="cnt">{{ sections.length }}/{{ SECTION_LIMIT }}</span>
        </button>
        <p class="mo-tip">管理操作仅作用于本地数据，导出前请先备份。</p>
      </div>
    </el-dialog>

    <!-- 沉浸式发帖页（方案 C+D：整页 + 两步；手动创建/编辑专用） -->
    <teleport to="body">
      <div v-if="composerVisible" class="composer">
        <header class="cp-topbar">
          <button class="cp-back" type="button" @click="closeComposer">← 返回</button>
          <div class="cp-title">{{ isEdit ? '编辑帖子' : '发帖' }} · {{ boardConf.name }}</div>
          <button class="cp-publish" type="button" :disabled="submitting" @click="handleSubmit">
            {{ submitting ? '提交中…' : (isEdit ? '保存修改' : '发布') }}
          </button>
        </header>

        <div class="cp-steps">
          <span class="cp-step" :class="{ on: composerStep === 1 }">① 基本信息</span>
          <span class="cp-step-line"></span>
          <span class="cp-step" :class="{ on: composerStep === 2 }">② 补充信息（可选）</span>
        </div>

        <main class="cp-body">
          <!-- 第一步：贴吧三件套 -->
          <template v-if="composerStep === 1">
            <div class="cp-row">
              <span class="cp-label">发布到吧</span>
              <span class="cp-bar"><i class="cp-bar-av">{{ boardConf.short }}</i>{{ boardConf.name }}</span>
            </div>
            <div class="cp-row">
              <span class="cp-label">版块分区</span>
              <div class="sec-chips">
                <button type="button" class="sec-chip" :class="{ on: !formData.section_id }" @click="formData.section_id = null">未归类</button>
                <button
                  v-for="s in sections"
                  :key="s.id"
                  type="button"
                  class="sec-chip"
                  :class="{ on: formData.section_id === s.id }"
                  @click="formData.section_id = formData.section_id === s.id ? null : s.id"
                >{{ s.name }}</button>
              </div>
            </div>
            <el-input
              v-model="formData.title"
              maxlength="120"
              show-word-limit
              placeholder="必填，请输入帖子标题（5-31 个字）"
              class="cp-title-input"
            />
            <TipTapEditor
              v-model="formData.content"
              placeholder="请输入正文…"
              class="cp-editor"
            />
          </template>

          <!-- 第二步：补充信息（可选） -->
          <template v-else>
            <div class="cp-row top">
              <span class="cp-label">简介</span>
              <el-input
                v-model="formData.summary"
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
                <div v-if="currentCoverDisplay" class="cover-preview-card">
                  <img :src="currentCoverDisplay.url" alt="封面预览" class="cover-preview-image" />
                  <div class="cover-preview-actions">
                    <el-button type="primary" plain size="small" @click="openCoverImagePicker">替换图片</el-button>
                    <el-button type="danger" plain size="small" @click="removeCoverImage">移除</el-button>
                  </div>
                </div>
                <div v-else class="cover-upload-area" @click="openCoverImagePicker">
                  <el-icon class="cover-upload-icon"><Plus /></el-icon>
                  <span class="cover-upload-text">点击上传封面图</span>
                  <span class="cover-upload-hint">支持 jpg、png、gif 格式，建议尺寸 16:9</span>
                </div>
                <div class="cover-url-input">
                  <span class="cover-url-label">或输入图片URL：</span>
                  <el-input v-model="formData.cover_url" placeholder="https://example.com/cover.jpg" clearable @input="handleCoverUrlInput" />
                </div>
              </div>
            </div>
            <div class="cp-row">
              <span class="cp-label">内容声明</span>
              <div class="declare-row">
                <el-checkbox v-model="formData.is_featured">设为精华</el-checkbox>
                <span class="declare-hint">精华与分区互不冲突，精华帖会在「精华」入口下出现</span>
              </div>
            </div>
            <div class="cp-row">
              <span class="cp-label"></span>
              <el-button text type="primary" @click="openPreviewDialog">预览发布效果</el-button>
            </div>
          </template>
        </main>

        <footer class="cp-footer">
          <template v-if="composerStep === 1">
            <button class="cp-skip" type="button" :disabled="submitting" @click="handleSubmit">直接发布</button>
            <button class="cp-next" type="button" @click="composerNext">下一步：补充信息（可选）→</button>
          </template>
          <template v-else>
            <button class="cp-skip" type="button" @click="composerPrev">← 上一步</button>
            <button class="cp-next" type="button" :disabled="submitting" @click="handleSubmit">
              {{ submitting ? '提交中…' : (isEdit ? '保存修改' : '发布') }}
            </button>
          </template>
        </footer>
      </div>
    </teleport>

    <!-- 创建 / 编辑弹窗（贴吧发帖风；PDF 导入仍走这里） -->
    <el-dialog
      v-model="dialogVisible"
      width="min(1000px, 94vw)"
      top="3vh"
      :close-on-click-modal="false"
      append-to-body
      @closed="resetFormState"
    >
      <template #header>
        <div class="tb-dlg-tabs">
          <span class="tb-tab on">{{ isEdit ? '编辑帖子' : '发帖' }}</span>
        </div>
      </template>
      <div class="dialog-body">
        <div class="tb-post-to">
          <span class="tb-row-label">发布到吧</span>
          <span class="tb-bar-chip"><i class="tb-bar-av">{{ boardConf.short }}</i>{{ boardConf.name }}</span>
        </div>

        <el-alert
          v-if="isPdfImportMode"
          title="导入后系统只会自动生成标题和简介，正文不再强行解析，而是直接保留 PDF 原始预览效果。"
          type="info"
          :closable="false"
          show-icon
          class="section-alert"
        />

        <div v-if="isPdfImportMode" class="import-panel">
          <div class="import-info">
            <div class="import-title">PDF 导入</div>
            <div class="import-desc">
              上传一个 PDF 后，系统会自动回填标题和简介，正文区域则直接展示原始 PDF 预览，避免版式被错误解析。
            </div>
          </div>
          <div class="import-actions">
            <el-button type="primary" :loading="importingPdf" @click="openPdfPicker('import')">
              {{ currentPdfDisplay ? '重新导入 PDF' : '选择 PDF 并生成初稿' }}
            </el-button>
            <span v-if="currentPdfDisplay" class="import-file-name">
              {{ currentPdfDisplay.fileName }}
            </span>
          </div>
        </div>

        <el-form label-width="96px" class="experience-form">
          <el-form-item label="版块分区">
            <div class="sec-chips">
              <button
                type="button"
                class="sec-chip"
                :class="{ on: !formData.section_id }"
                @click="formData.section_id = null"
              >未归类</button>
              <button
                v-for="s in sections"
                :key="s.id"
                type="button"
                class="sec-chip"
                :class="{ on: formData.section_id === s.id }"
                @click="formData.section_id = formData.section_id === s.id ? null : s.id"
              >{{ s.name }}</button>
            </div>
          </el-form-item>

          <el-form-item label="标题" required>
            <el-input
              v-model="formData.title"
              maxlength="120"
              show-word-limit
              placeholder="必填，请输入帖子标题（5-31 个字）"
            />
          </el-form-item>

          <el-form-item label="简介">
            <el-input
              v-model="formData.summary"
              type="textarea"
              :rows="4"
              maxlength="200"
              show-word-limit
              placeholder="选填，建议一句话概括核心经验，列表页会优先展示这里的内容"
            />
          </el-form-item>

          <el-form-item v-if="!isPdfImportMode" label="封面图">
            <div class="cover-field">
              <div v-if="currentCoverDisplay" class="cover-preview-card">
                <img :src="currentCoverDisplay.url" alt="封面预览" class="cover-preview-image" />
                <div class="cover-preview-actions">
                  <el-button type="primary" plain size="small" @click="openCoverImagePicker">
                    替换图片
                  </el-button>
                  <el-button type="danger" plain size="small" @click="removeCoverImage">
                    移除
                  </el-button>
                </div>
              </div>
              <div v-else class="cover-upload-area" @click="openCoverImagePicker">
                <el-icon class="cover-upload-icon"><Plus /></el-icon>
                <span class="cover-upload-text">点击上传封面图</span>
                <span class="cover-upload-hint">支持 jpg、png、gif 格式，建议尺寸 16:9</span>
              </div>
              <div class="cover-url-input">
                <span class="cover-url-label">或输入图片URL：</span>
                <el-input
                  v-model="formData.cover_url"
                  placeholder="https://example.com/cover.jpg"
                  clearable
                  @input="handleCoverUrlInput"
                />
              </div>
            </div>
          </el-form-item>

          <el-form-item v-if="!isPdfImportMode" label="正文" required>
            <TipTapEditor
              v-model="formData.content"
              placeholder="请输入正文内容"
              class="content-editor"
            />
          </el-form-item>

          <el-form-item v-else label="PDF 页面">
            <div class="pdf-content-preview">
              <div class="pdf-content-preview__info">
                <strong>PDF 页面图片展示</strong>
                <span>PDF 有多少页，这里就展示多少页图片，正文不再作为必填项。</span>
              </div>
              <div v-if="currentPdfDisplay" class="pdf-content-preview__frame">
                <PdfPageGallery :src="currentPdfDisplay.url" :scale="1.1" compact />
              </div>
              <div v-else class="pdf-content-preview__empty">
                请先上传 PDF 文件，系统会按页展示图片内容。
              </div>
            </div>
          </el-form-item>

          <el-form-item label="内容声明">
            <div class="declare-row">
              <el-checkbox v-model="formData.is_featured">设为精华</el-checkbox>
              <span class="declare-hint">精华与分区互不冲突，精华帖会在「精华」入口下出现</span>
            </div>
          </el-form-item>
        </el-form>
      </div>

      <template #footer>
        <div class="form-footer">
          <div class="form-footer-left">
            <el-button @click="openPreviewDialog">预览效果</el-button>
            <el-button v-if="!currentPdfDisplay" @click="openPdfPicker(isPdfImportMode ? 'import' : 'attachment')">
              选择 PDF
            </el-button>
          </div>
          <div class="form-footer-right">
            <el-button text @click="dialogVisible = false">取消</el-button>
            <el-button type="primary" round class="tb-publish" :loading="submitting" @click="handleSubmit">
              {{ isEdit ? '保存修改' : '发布' }}
            </el-button>
          </div>
        </div>
      </template>
    </el-dialog>

    <!-- 分区管理弹窗 -->
    <el-dialog
      v-model="sectionMgrVisible"
      title="分区管理"
      width="min(560px, 92vw)"
      top="8vh"
      append-to-body
    >
      <div class="sec-mgr">
        <div class="sec-add-row">
          <el-input
            v-model="sectionDraft"
            maxlength="12"
            placeholder="新分区名称（最多 12 字）"
            :disabled="!canAddSection"
            @keyup.enter="createSection"
          />
          <el-button
            type="primary"
            :disabled="!canAddSection || !sectionDraft.trim()"
            @click="createSection"
          >
            新增分区
          </el-button>
        </div>
        <div class="sec-tip">
          自定义分区上限 {{ SECTION_LIMIT }} 个，当前 {{ sections.length }} 个。「精华」与「全部」是内置入口，不占用名额。
        </div>
        <div class="sec-list">
          <div v-for="(s, idx) in sections" :key="s.id" class="sec-row">
            <span class="sec-name">{{ s.name }}</span>
            <span class="sec-count">{{ countInSection(s.id) }} 篇</span>
            <span class="sec-ops">
              <el-button link size="small" :disabled="idx === 0" @click="moveSectionOrder(idx, -1)">上移</el-button>
              <el-button link size="small" :disabled="idx === sections.length - 1" @click="moveSectionOrder(idx, 1)">下移</el-button>
              <el-button link size="small" @click="renameSection(s)">重命名</el-button>
              <el-button link type="danger" size="small" @click="removeSection(s)">删除</el-button>
            </span>
          </div>
          <div v-if="sections.length === 0" class="sec-empty">还没有自定义分区，先新增一个吧</div>
        </div>
      </div>
    </el-dialog>

    <!-- 效果预览弹窗 -->
    <el-dialog
      v-model="previewDialogVisible"
      title="效果预览"
      width="800px"
      top="5vh"
      destroy-on-close
      append-to-body
      class="preview-dialog"
    >
      <div class="preview-content">
        <div class="preview-card">
          <div v-if="previewData.cover_url" class="preview-cover">
            <img :src="previewData.cover_url" alt="封面" />
          </div>
          <div class="preview-body">
            <h2 class="preview-title">{{ previewData.title || '未填写标题' }}</h2>
            <p v-if="previewData.summary" class="preview-summary">{{ previewData.summary }}</p>
            <div class="preview-main">
              <div v-if="previewData.content" class="preview-html" v-html="previewData.content"></div>
              <p v-else class="preview-empty">暂无正文内容</p>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- PDF 在线预览弹窗 -->
    <el-dialog
      v-model="pdfPreviewVisible"
      title="PDF 在线预览"
      width="min(1100px, 95vw)"
      top="3vh"
      destroy-on-close
      append-to-body
    >
      <div class="pdf-preview-shell">
        <div class="pdf-preview-toolbar">
          <div class="pdf-preview-file">
            <el-icon><Document /></el-icon>
            <span>{{ pdfPreviewFileName }}</span>
          </div>
          <div class="pdf-preview-actions">
            <el-button v-if="pdfPreviewUrl" @click="openPdfInNewTab">新窗口打开</el-button>
            <el-button v-if="pdfPreviewUrl" type="primary" @click="downloadPdf">下载</el-button>
          </div>
        </div>
        <div class="pdf-preview-body">
          <PdfPageGallery v-if="pdfPreviewUrl" :src="pdfPreviewUrl" :scale="1.45" />
        </div>
      </div>
    </el-dialog>

    <input
      ref="pdfFileInputRef"
      type="file"
      accept=".pdf,application/pdf"
      style="display: none"
      @change="handlePdfFileChange"
    />

    <input
      ref="coverImageInputRef"
      type="file"
      accept="image/jpeg,image/png,image/gif,image/webp"
      style="display: none"
      @change="handleCoverImageChange"
    />

    <input
      ref="importFileInputRef"
      type="file"
      accept=".json,application/json"
      style="display: none"
      @change="handleImportFileChange"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { experienceShareAPI, experienceSectionAPI } from '@/api'
import type { ExperienceShare, ExperienceSection, BoardKind } from '@/types'
import { BOARD_CONF, normalizeBoard, tutorialBoardPath, tutorialPostPath } from '@/constants/boards'
import TipTapEditor from '@/components/TipTapEditor.vue'
import PdfPageGallery from '@/components/PdfPageGallery.vue'

type CreateMode = 'manual' | 'pdf_import'
type PdfPickerMode = 'attachment' | 'import'
/** 分区入口：'featured' 精华 / 'all' 全部 / 'unfiled' 未归类 / number 自定义分区 id */
type SectionKey = 'featured' | 'all' | 'unfiled' | number
/** 列表排序口径：最多回复 / 最多点赞 / 最新 */
type SortKey = 'replies' | 'likes' | 'time'
/** 自定义分区上限（不含内置的「精华」「全部」） */
const SECTION_LIMIT = 15

type PdfUploadPayload = {
  name: string
  size: number
  data_base64: string
}

type LocalPdfAttachment = PdfUploadPayload & {
  objectUrl: string
}

type CoverImagePayload = {
  name: string
  size: number
  data_base64: string
}

type LocalCoverImage = CoverImagePayload & {
  objectUrl: string
}

type FormState = {
  id: number
  title: string
  summary: string
  cover_url: string
  content: string
  content_render_mode: 'markdown' | 'html'
  create_type: CreateMode
  author_name: string
  status: string
  pdf_parse_status: string
  pdf_parse_result: string
  source_file_name: string
  section_id: number | null
  is_featured: boolean
  existing_pdf_file_url: string
  existing_pdf_file_name: string
  existing_pdf_file_size: number
  remove_pdf: boolean
  remove_cover: boolean
}

const MAX_PDF_SIZE = 20 * 1024 * 1024

const SORT_LABELS: Record<SortKey, string> = {
  replies: '最多回复',
  likes: '最多点赞',
  time: '最新'
}

const COVER_GRADIENTS = [
  'linear-gradient(135deg,#667eea 0%,#764ba2 100%)',
  'linear-gradient(135deg,#08c6be 0%,#0ea5e9 100%)',
  'linear-gradient(135deg,#f59e0b 0%,#ef4444 100%)',
  'linear-gradient(135deg,#10b981 0%,#059669 100%)',
  'linear-gradient(135deg,#8b5cf6 0%,#ec4899 100%)',
  'linear-gradient(135deg,#0ea5e9 0%,#6366f1 100%)',
  'linear-gradient(135deg,#14b8a6 0%,#38bdf8 100%)',
  'linear-gradient(135deg,#f97316 0%,#f43f5e 100%)'
]

const route = useRoute()
const router = useRouter()

/**
 * 当前吧。由路由参数决定（/tutorial/board/:board）。
 *
 * ⚠ 同一个组件被两个吧复用，路由切换时组件**不会重新挂载**（只是 param 变了），
 * 所以所有跟吧相关的状态都必须靠 watch(currentBoard) 重置，见下方 watch。
 */
const currentBoard = computed<BoardKind>(() => {
  const raw = route.params.board
  return normalizeBoard(Array.isArray(raw) ? raw[0] : raw)
})
const boardConf = computed(() => BOARD_CONF[currentBoard.value])

const loading = ref(false)
const submitting = ref(false)
const importingPdf = ref(false)

const experienceShares = ref<ExperienceShare[]>([])
const dialogVisible = ref(false)
const pdfPreviewVisible = ref(false)
const previewDialogVisible = ref(false)
const previewData = ref({
  title: '',
  summary: '',
  content: '',
  cover_url: ''
})

/* ---- 工作台：分区 / 排序 / 搜索 / 分页 ---- */
const activeSection = ref<SectionKey>('all')
const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = 20
const gridRef = ref<HTMLElement | null>(null)

/* 排序口径：最多回复 / 最多点赞 / 最新（默认最新） */
const sortBy = ref<SortKey>('time')

/* 自定义分区 */
const sections = ref<ExperienceSection[]>([])
const sectionMgrVisible = ref(false)
const sectionDraft = ref('')
const canAddSection = computed(() => sections.value.length < SECTION_LIMIT)

const fetchSections = async () => {
  try {
    const res = await experienceSectionAPI.getAll({ board: currentBoard.value })
    if (res.success && res.data) {
      sections.value = res.data
    } else {
      sections.value = []
    }
  } catch {
    /* 后端未重启拿不到分区表时，列表仍要能正常浏览 */
    sections.value = []
  }
}

const byTimeDesc = (a: ExperienceShare, b: ExperienceShare) =>
  new Date(b.created_at || 0).getTime() - new Date(a.created_at || 0).getTime()

const filteredExperienceShares = computed(() => {
  // 先复制一份再筛选，否则无筛选直接 sort 会把原数组顺序改掉
  // 置顶帖单独在上面的置顶区展示，这里排除，避免同一篇出现两次
  let list = experienceShares.value.filter(i => i.is_pinned !== 1)
  const key = activeSection.value
  if (key === 'featured') {
    list = list.filter(i => i.is_featured === 1)
  } else if (key === 'unfiled') {
    list = list.filter(i => !i.section_id)
  } else if (typeof key === 'number') {
    list = list.filter(i => i.section_id === key)
  }

  const q = searchQuery.value.trim()
  if (q) {
    list = list.filter(i => {
      const hay = `${i.title || ''} ${i.summary || ''} ${i.author_name || ''}`.toLowerCase()
      return hay.includes(q.toLowerCase())
    })
  }

  if (sortBy.value === 'replies') {
    list.sort((a, b) => replyCount(b.id) - replyCount(a.id) || byTimeDesc(a, b))
  } else if (sortBy.value === 'likes') {
    list.sort((a, b) => likeTotal(b) - likeTotal(a) || byTimeDesc(a, b))
  } else {
    list.sort(byTimeDesc)
  }
  return list
})

/* ---- 分区管理 ---- */
/** 置顶区（置顶帖按时间倒序，与贴吧一致：新的在上） */
const pinnedShares = computed(() =>
  experienceShares.value
    .filter(i => i.is_pinned === 1)
    .sort(byTimeDesc)
)

const countInSection = (id: number) =>
  experienceShares.value.filter(i => i.section_id === id).length

const sectionTabs = computed(() => {
  const tabs: { key: SectionKey; label: string; count: number; builtin: boolean }[] = [
    { key: 'featured', label: '精华', count: countFeatured.value, builtin: true },
    { key: 'all', label: '全部', count: countAll.value, builtin: true }
  ]
  for (const s of sections.value) {
    tabs.push({ key: s.id, label: s.name, count: countInSection(s.id), builtin: false })
  }
  // 只有在已经有自定义分区、且确实存在未归类帖子时才挂这个入口，避免空入口占位
  if (sections.value.length > 0) {
    const n = experienceShares.value.filter(i => !i.section_id).length
    if (n > 0) tabs.push({ key: 'unfiled', label: '未归类', count: n, builtin: true })
  }
  return tabs
})

const setSectionFilter = (key: SectionKey) => {
  activeSection.value = key
}

const openSectionManager = () => {
  sectionDraft.value = ''
  sectionMgrVisible.value = true
}

const createSection = async () => {
  const name = sectionDraft.value.trim()
  if (!name) return
  if (!canAddSection.value) {
    ElMessage.warning(`自定义分区最多 ${SECTION_LIMIT} 个`)
    return
  }
  try {
    const res = await experienceSectionAPI.create(name, currentBoard.value)
    if (res.success) {
      sectionDraft.value = ''
      await fetchSections()
      ElMessage.success('分区已创建')
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '创建分区失败')
  }
}

const renameSection = async (s: ExperienceSection) => {
  try {
    const { value } = await ElMessageBox.prompt('分区名称（最多 12 个字）', '重命名分区', {
      inputValue: s.name,
      inputValidator: (v: string) => {
        const t = (v || '').trim()
        if (!t) return '名称不能为空'
        if (t.length > 12) return '最多 12 个字'
        return true
      }
    })
    const res = await experienceSectionAPI.rename(s.id, String(value).trim())
    if (res.success) {
      await fetchSections()
      ElMessage.success('已重命名')
    }
  } catch {
    /* 用户取消 */
  }
}

const removeSection = async (s: ExperienceSection) => {
  try {
    await ElMessageBox.confirm(
      `删除分区「${s.name}」后，该分区下的 ${countInSection(s.id)} 篇帖子不会被删除，会变成「未归类」。`,
      '删除分区',
      { type: 'warning' }
    )
  } catch {
    return
  }
  try {
    const res = await experienceSectionAPI.remove(s.id)
    if (res.success) {
      if (activeSection.value === s.id) activeSection.value = 'all'
      await fetchSections()
      ElMessage.success('分区已删除')
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '删除分区失败')
  }
}

/* 上移 / 下移：本地换序后把最终顺序整体提交 */
const moveSectionOrder = async (idx: number, dir: -1 | 1) => {
  const next = [...sections.value]
  const target = idx + dir
  if (target < 0 || target >= next.length) return
  ;[next[idx], next[target]] = [next[target], next[idx]]
  try {
    await experienceSectionAPI.reorder(next.map(s => s.id))
    sections.value = next.map((s, i) => ({ ...s, order_num: i + 1 }))
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '排序失败')
  }
}

/* ---- 单帖操作：置顶 / 加精 / 改分区 ---- */
const togglePinned = async (item: ExperienceShare) => {
  const next: 0 | 1 = item.is_pinned === 1 ? 0 : 1
  try {
    const res = await experienceShareAPI.setPinned(item.id, next)
    if (res.success) {
      item.is_pinned = next
      ElMessage.success(next === 1 ? '已置顶（每吧最多 3 条）' : '已取消置顶')
      // 置顶影响排序，重拉一次让列表顺序跟上
      await fetchExperienceShares()
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '操作失败')
  }
}

const toggleFeatured = async (item: ExperienceShare) => {
  const next: 0 | 1 = item.is_featured === 1 ? 0 : 1
  try {
    const res = await experienceShareAPI.setFeatured(item.id, next)
    if (res.success) {
      item.is_featured = next
      ElMessage.success(next === 1 ? '已设为精华' : '已取消精华')
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '操作失败')
  }
}

const moveToSection = async (item: ExperienceShare, sectionId: number | null) => {
  try {
    const res = await experienceShareAPI.setSection(item.id, sectionId)
    if (res.success) {
      item.section_id = sectionId
      ElMessage.success(sectionId ? '已移动分区' : '已移出到未归类')
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '操作失败')
  }
}

const batchMoveVisible = ref(false)
const batchTargetSection = ref<number | null>(null)

const submitBatchMove = async () => {
  if (selectedIds.value.length === 0) return
  try {
    const res = await experienceShareAPI.batchSetSection(selectedIds.value, batchTargetSection.value)
    if (res.success) {
      ElMessage.success(`已移动 ${res.data?.affected ?? selectedIds.value.length} 篇`)
      batchMoveVisible.value = false
      cancelSelection()
      await fetchExperienceShares()
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '批量移动失败')
  }
}

/* 回复数：与详情页 BBS 共用同一份 localStorage（key: bbs-replies-<id>） */
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

const replyCount = (id: number) => readBbsReplies(id).length

/* 楼层赞（localStorage）；主楼赞在后端 like_count，两边相加才是帖子总赞 */
const floorLikeCount = (id: number) =>
  readBbsReplies(id).reduce((sum, r) => sum + (Number(r?.likes) || 0), 0)

const likeTotal = (item: ExperienceShare) =>
  (Number(item.like_count) || 0) + floorLikeCount(item.id)

const formatListDate = (value?: string | null) => {
  if (!value) return ''
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

const paginatedExperienceShares = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredExperienceShares.value.slice(start, start + pageSize)
})

const totalItems = computed(() => filteredExperienceShares.value.length)
const totalPages = computed(() => Math.ceil(totalItems.value / pageSize))

const countAll = computed(() => experienceShares.value.length)
const countFeatured = computed(() => experienceShares.value.filter(i => i.is_featured === 1).length)
const countManual = computed(() => experienceShares.value.filter(i => i.create_type === 'manual').length)
const countPdf = computed(() => experienceShares.value.filter(i => i.create_type === 'pdf_import').length)
const countAttach = computed(() => experienceShares.value.filter(i => !!i.pdf_file_url).length)

const totalReplies = computed(() =>
  experienceShares.value.reduce((sum, item) => sum + replyCount(item.id), 0)
)

const latestShares = computed(() =>
  [...experienceShares.value]
    .sort((a, b) => (b.created_at || '').localeCompare(a.created_at || ''))
    .slice(0, 5)
)

const sectionName = (id?: number | null) =>
  id ? sections.value.find(s => s.id === id)?.name || '' : ''

const statTitle = computed(() => {
  const tab = sectionTabs.value.find(t => t.key === activeSection.value)
  const base = activeSection.value === 'all' ? '全部帖子' : tab?.label || '全部帖子'
  const q = searchQuery.value.trim()
  return `${base} · ${totalItems.value} 项 · ${SORT_LABELS[sortBy.value]}` + (q ? ` · 搜索「${q}」` : '')
})

const coverGradient = (item: ExperienceShare) => {
  const idx = (item.id || 0) % COVER_GRADIENTS.length
  return COVER_GRADIENTS[idx]
}

const resetFilter = () => {
  activeSection.value = 'all'
  searchQuery.value = ''
}

watch([activeSection, searchQuery, sortBy], () => {
  currentPage.value = 1
})

/* ---- 选择模式 ---- */
const selectionMode = ref(false)
const selectedIds = ref<number[]>([])

const toggleSelectionMode = () => {
  if (selectionMode.value) {
    cancelSelection()
  } else {
    enterSelectionMode()
  }
}

const enterSelectionMode = () => {
  selectionMode.value = true
  selectedIds.value = []
}

const cancelSelection = () => {
  selectionMode.value = false
  selectedIds.value = []
}

const selectAll = computed(() => {
  const list = filteredExperienceShares.value
  return list.length > 0 && list.every(i => selectedIds.value.includes(i.id))
})

const isIndeterminate = computed(() => {
  const list = filteredExperienceShares.value
  const total = list.length
  const selected = list.filter(i => selectedIds.value.includes(i.id)).length
  return selected > 0 && selected < total
})

const handleSelectAllChange = (val: boolean) => {
  const list = filteredExperienceShares.value
  if (val) {
    list.forEach(i => {
      if (!selectedIds.value.includes(i.id)) selectedIds.value.push(i.id)
    })
  } else {
    const ids = new Set(list.map(i => i.id))
    selectedIds.value = selectedIds.value.filter(id => !ids.has(id))
  }
}

const handleSelectChange = (id: number, val: boolean) => {
  if (val) {
    if (!selectedIds.value.includes(id)) {
      selectedIds.value.push(id)
    }
  } else {
    selectedIds.value = selectedIds.value.filter(i => i !== id)
  }
}

const handleCardClick = (item: ExperienceShare) => {
  if (selectionMode.value) {
    const isSelected = selectedIds.value.includes(item.id)
    handleSelectChange(item.id, !isSelected)
  } else {
    openDetail(item)
  }
}

/* ---- 编辑状态 / 表单 ---- */
const isEdit = ref(false)
const pdfPreviewUrl = ref('')
const pdfPreviewFileName = ref('')
const pdfPreviewDownloadName = ref('')
const pdfFileInputRef = ref<HTMLInputElement | null>(null)
const coverImageInputRef = ref<HTMLInputElement | null>(null)
const pickerMode = ref<PdfPickerMode>('attachment')
const localPdfAttachment = ref<LocalPdfAttachment | null>(null)
const localCoverImage = ref<LocalCoverImage | null>(null)

const createEmptyForm = (mode: CreateMode = 'manual'): FormState => ({
  id: 0,
  title: '',
  summary: '',
  cover_url: '',
  content: '',
  content_render_mode: 'html',
  create_type: mode,
  author_name: '星芒用户',
  status: 'published',
  pdf_parse_status: '',
  pdf_parse_result: '',
  source_file_name: '',
  section_id: null,
  is_featured: false,
  existing_pdf_file_url: '',
  existing_pdf_file_name: '',
  existing_pdf_file_size: 0,
  remove_pdf: false,
  remove_cover: false
})

const formData = ref<FormState>(createEmptyForm())

const dialogTitle = computed(() => {
  if (isEdit.value) return '编辑经验卡片'
  return formData.value.create_type === 'pdf_import' ? 'PDF 导入创建' : '手动创建经验卡片'
})

const stripRichText = (value = '') =>
  value
    .replace(/<[^>]+>/g, ' ')
    .replace(/&nbsp;/gi, ' ')
    .replace(/&amp;/gi, '&')
    .replace(/&lt;/gi, '<')
    .replace(/&gt;/gi, '>')
    .replace(/&quot;/gi, '"')
    .replace(/&#39;/gi, "'")
    .replace(/\s+/g, ' ')
    .trim()

const hasMeaningfulContent = (value = '') => stripRichText(value).length > 0

const isPdfImportMode = computed(() => formData.value.create_type === 'pdf_import')

const currentPdfDisplay = computed(() => {
  if (localPdfAttachment.value) {
    return {
      fileName: localPdfAttachment.value.name,
      fileSize: localPdfAttachment.value.size,
      url: localPdfAttachment.value.objectUrl,
      downloadName: localPdfAttachment.value.name,
      isLocal: true
    }
  }

  if (!formData.value.remove_pdf && formData.value.existing_pdf_file_url) {
    return {
      fileName: formData.value.existing_pdf_file_name || 'PDF附件',
      fileSize: formData.value.existing_pdf_file_size,
      url: formData.value.existing_pdf_file_url,
      downloadName: formData.value.existing_pdf_file_name || '经验分享附件.pdf',
      isLocal: false
    }
  }

  return null
})

const currentCoverDisplay = computed(() => {
  if (localCoverImage.value) {
    return {
      url: localCoverImage.value.objectUrl,
      name: localCoverImage.value.name,
      size: localCoverImage.value.size,
      isLocal: true
    }
  }

  if (formData.value.cover_url) {
    return {
      url: formData.value.cover_url,
      name: '',
      size: 0,
      isLocal: false
    }
  }

  return null
})

const revokeLocalPdf = () => {
  if (localPdfAttachment.value?.objectUrl) {
    URL.revokeObjectURL(localPdfAttachment.value.objectUrl)
  }
  localPdfAttachment.value = null
}

const revokeLocalCoverImage = () => {
  if (localCoverImage.value?.objectUrl) {
    URL.revokeObjectURL(localCoverImage.value.objectUrl)
  }
  localCoverImage.value = null
}

const MAX_COVER_IMAGE_SIZE = 5 * 1024 * 1024

const openCoverImagePicker = () => {
  coverImageInputRef.value?.click()
}

const handleCoverImageChange = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]

  try {
    if (!file) return

    if (!file.type.startsWith('image/')) {
      throw new Error('请选择图片文件')
    }

    if (file.size > MAX_COVER_IMAGE_SIZE) {
      throw new Error('图片大小不能超过 5MB')
    }

    revokeLocalCoverImage()

    const data_base64 = await new Promise<string>((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => {
        const result = reader.result as string
        const base64 = result.split(',')[1]
        resolve(base64)
      }
      reader.onerror = () => reject(new Error('读取图片失败'))
      reader.readAsDataURL(file)
    })

    const objectUrl = URL.createObjectURL(file)

    localCoverImage.value = {
      name: file.name,
      size: file.size,
      data_base64,
      objectUrl
    }
    formData.value.cover_url = ''
    formData.value.remove_cover = false
  } catch (error: any) {
    ElMessage.error(error.message || '上传图片失败')
  } finally {
    target.value = ''
  }
}

const removeCoverImage = () => {
  revokeLocalCoverImage()
  if (formData.value.cover_url) {
    formData.value.remove_cover = true
  }
  formData.value.cover_url = ''
}

const handleCoverUrlInput = () => {
  if (formData.value.cover_url && localCoverImage.value) {
    revokeLocalCoverImage()
  }
}

const resetFormState = () => {
  revokeLocalPdf()
  revokeLocalCoverImage()
  formData.value = createEmptyForm()
  isEdit.value = false
  pickerMode.value = 'attachment'
}

const fetchExperienceShares = async () => {
  loading.value = true
  try {
    const res = await experienceShareAPI.getAll({ board: currentBoard.value })
    if (res.success && res.data) {
      experienceShares.value = res.data
      currentPage.value = 1
      syncLikedFromStorage()
    }
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  gridRef.value?.scrollTo({ top: 0, behavior: 'smooth' })
}

const openCreateDialog = (mode: CreateMode) => {
  resetFormState()
  formData.value = createEmptyForm(mode)
  /* 沉浸式发帖页只服务手动创建/编辑；PDF 导入维持原弹窗 */
  if (mode === 'manual') {
    composerVisible.value = true
    composerStep.value = 1
  } else {
    dialogVisible.value = true
  }
}

const handleCreateCommand = (command: CreateMode) => {
  openCreateDialog(command)
}

const openDetail = (item: ExperienceShare) => {
  router.push(tutorialPostPath(item.id))
}

const fillEditForm = (item: ExperienceShare) => {
  revokeLocalPdf()
  revokeLocalCoverImage()
  isEdit.value = true
  formData.value = {
    id: item.id,
    title: item.title || '',
    summary: item.summary || '',
    cover_url: item.cover_url || '',
    content: item.content || '',
    content_render_mode: item.content_render_mode === 'markdown' ? 'markdown' : 'html',
    create_type: item.create_type || 'manual',
    author_name: item.author_name || '星芒用户',
    status: item.status || 'published',
    pdf_parse_status: item.pdf_parse_status || '',
    pdf_parse_result: item.pdf_parse_result || '',
    source_file_name: item.source_file_name || '',
    section_id: item.section_id ?? null,
    is_featured: item.is_featured === 1,
    existing_pdf_file_url: item.pdf_file_url || '',
    existing_pdf_file_name: item.pdf_file_name || '',
    existing_pdf_file_size: Number(item.pdf_file_size || 0),
    remove_pdf: false,
    remove_cover: false
  }
  if ((formData.value.create_type || 'manual') === 'manual') {
    composerVisible.value = true
    composerStep.value = 1
  } else {
    dialogVisible.value = true
  }
}

const openEdit = (item: ExperienceShare) => {
  fillEditForm(item)
}

/* ---- 沉浸式发帖页（方案 C+D 合成：整页 + 两步流转） ---- */
const composerVisible = ref(false)
const composerStep = ref<1 | 2>(1)
const closeComposer = () => {
  composerVisible.value = false
  resetFormState()
}
const composerNext = () => {
  composerStep.value = 2
}
const composerPrev = () => {
  composerStep.value = 1
}

const syncEditFromQuery = async () => {
  const editId = Number(route.query.edit)
  if (!Number.isFinite(editId) || editId <= 0) return

  const existing = experienceShares.value.find(item => item.id === editId)

  if (existing) {
    fillEditForm(existing)
  } else {
    const res = await experienceShareAPI.getOne(editId)
    if (res.success && res.data) {
      fillEditForm(res.data)
    }
  }

  router.replace({ path: tutorialBoardPath(currentBoard.value) })
}

const validatePdfFile = (file: File) => {
  if (!/\.pdf$/i.test(file.name) && file.type !== 'application/pdf') {
    throw new Error('仅支持 PDF 格式文件')
  }

  if (file.size > MAX_PDF_SIZE) {
    throw new Error(`PDF 文件大小不能超过 ${Math.floor(MAX_PDF_SIZE / 1024 / 1024)}MB`)
  }
}

const fileToBase64 = (file: File) =>
  new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = typeof reader.result === 'string' ? reader.result : ''
      resolve(result.replace(/^data:application\/pdf;base64,/i, ''))
    }
    reader.onerror = () => reject(new Error('文件读取失败'))
    reader.readAsDataURL(file)
  })

const createLocalPdfAttachment = async (file: File) => {
  validatePdfFile(file)
  revokeLocalPdf()

  localPdfAttachment.value = {
    name: file.name,
    size: file.size,
    data_base64: await fileToBase64(file),
    objectUrl: URL.createObjectURL(file)
  }
  formData.value.remove_pdf = false
}

const openPdfPicker = (mode: PdfPickerMode) => {
  pickerMode.value = mode
  pdfFileInputRef.value?.click()
}

const applyImportedDraft = (draft: Partial<ExperienceShare>) => {
  formData.value.title = draft.title || formData.value.title
  formData.value.summary = draft.summary || ''
  formData.value.content = ''
  formData.value.create_type = 'pdf_import'
  formData.value.pdf_parse_status = draft.pdf_parse_status || ''
  formData.value.pdf_parse_result = draft.pdf_parse_result || ''
  formData.value.source_file_name = draft.source_file_name || localPdfAttachment.value?.name || ''
}

const importPdfDraft = async (file: File) => {
  importingPdf.value = true
  try {
    await createLocalPdfAttachment(file)
    const res = await experienceShareAPI.importPdf({
      pdf_file: {
        name: localPdfAttachment.value!.name,
        size: localPdfAttachment.value!.size,
        data_base64: localPdfAttachment.value!.data_base64
      }
    })

    if (res.success && res.data) {
      applyImportedDraft(res.data)
      ElMessage.success('PDF 导入成功，已自动生成标题和简介')
    }
  } finally {
    importingPdf.value = false
  }
}

const handlePdfFileChange = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]

  try {
    if (!file) return

    if (pickerMode.value === 'import') {
      await importPdfDraft(file)
    } else {
      await createLocalPdfAttachment(file)
      ElMessage.success('PDF 附件已添加')
    }
  } catch (error: any) {
    ElMessage.error(error.message || 'PDF 处理失败')
  } finally {
    target.value = ''
  }
}

const openPreviewDialog = () => {
  const coverUrl = localCoverImage.value?.objectUrl || formData.value.cover_url || ''
  previewData.value = {
    title: formData.value.title,
    summary: formData.value.summary,
    content: formData.value.content,
    cover_url: coverUrl
  }
  previewDialogVisible.value = true
}

const previewCurrentPdf = () => {
  if (!currentPdfDisplay.value) return
  openPdfPreview(currentPdfDisplay.value.url, currentPdfDisplay.value.fileName, currentPdfDisplay.value.downloadName)
}

const openPdfPreview = (url: string, fileName: string, downloadName = fileName) => {
  pdfPreviewUrl.value = url
  pdfPreviewFileName.value = fileName
  pdfPreviewDownloadName.value = downloadName
  pdfPreviewVisible.value = true
}

const openPdfInNewTab = () => {
  if (!pdfPreviewUrl.value) return
  window.open(pdfPreviewUrl.value, '_blank', 'noopener')
}

const downloadPdf = () => {
  if (!pdfPreviewUrl.value) return
  const link = document.createElement('a')
  link.href = pdfPreviewUrl.value
  link.download = pdfPreviewDownloadName.value || 'experience-share.pdf'
  link.target = '_blank'
  link.rel = 'noopener'
  link.click()
}

const removeCurrentPdf = () => {
  if (localPdfAttachment.value) {
    revokeLocalPdf()
  } else if (formData.value.existing_pdf_file_url) {
    formData.value.remove_pdf = true
  }

  if (isPdfImportMode.value) {
    formData.value.pdf_parse_status = ''
    formData.value.pdf_parse_result = ''
  }
}

const handleDelete = async (item: ExperienceShare) => {
  try {
    await ElMessageBox.confirm(`确定删除经验卡片“${item.title}”吗？`, '删除经验卡片', {
      type: 'warning'
    })

    const res = await experienceShareAPI.delete(item.id)
    if (res.success) {
      ElMessage.success('删除成功')
      await fetchExperienceShares()
    }
  } catch {
    // 用户取消时不处理
  }
}

const handleSubmit = async () => {
  if (!formData.value.title.trim()) {
    ElMessage.warning('标题不能为空')
    return
  }

  if (!isPdfImportMode.value && !hasMeaningfulContent(formData.value.content)) {
    ElMessage.warning('正文不能为空')
    return
  }

  if (isPdfImportMode.value && !currentPdfDisplay.value) {
    ElMessage.warning('PDF 导入创建需要保留导入的 PDF 附件')
    return
  }

  submitting.value = true
  try {
    const payload: any = {
      title: formData.value.title.trim(),
      summary: formData.value.summary.trim(),
      cover_url: formData.value.cover_url.trim(),
      content: isPdfImportMode.value ? '' : formData.value.content,
      content_render_mode: formData.value.content_render_mode,
      create_type: formData.value.create_type,
      author_name: formData.value.author_name,
      status: formData.value.status,
      // 新帖归属当前吧；编辑时不传 board（避免把老帖误搬到另一个吧）
      ...(isEdit.value ? {} : { board: currentBoard.value }),
      pdf_parse_status: formData.value.pdf_parse_status || null,
      pdf_parse_result: formData.value.pdf_parse_result || null,
      source_file_name: formData.value.source_file_name || null,
      section_id: formData.value.section_id ?? null,
      is_featured: formData.value.is_featured ? 1 : 0,
      remove_pdf: formData.value.remove_pdf,
      remove_cover: formData.value.remove_cover
    }

    if (localPdfAttachment.value) {
      payload.pdf_file = {
        name: localPdfAttachment.value.name,
        size: localPdfAttachment.value.size,
        data_base64: localPdfAttachment.value.data_base64
      }
    }

    if (localCoverImage.value) {
      payload.cover_image = {
        name: localCoverImage.value.name,
        size: localCoverImage.value.size,
        data_base64: localCoverImage.value.data_base64
      }
    }

    if (isEdit.value) {
      const res = await experienceShareAPI.update(formData.value.id, payload)
      if (res.success) {
        ElMessage.success('更新成功')
      }
    } else {
      const res = await experienceShareAPI.create(payload)
      if (res.success) {
        ElMessage.success('发布成功')
      }
    }

    dialogVisible.value = false
    composerVisible.value = false
    await fetchExperienceShares()
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await fetchSections()
  await fetchExperienceShares()
  await syncEditFromQuery()
  document.addEventListener('click', closeManageOnOutsideClick)
})

onBeforeUnmount(() => {
  revokeLocalPdf()
  revokeLocalCoverImage()
  document.removeEventListener('click', closeManageOnOutsideClick)
})

watch(
  () => route.query.edit,
  () => {
    syncEditFromQuery()
  }
)

/* 左栏切吧：同一路由只变 param，组件不会重新挂载 → 必须手动重置再重载。
   ⚠ 不重置 activeSection 的话，会带着「教程吧的某个分区 id」去查经验分享吧，列表直接空。 */
watch(currentBoard, async () => {
  activeSection.value = 'all'
  searchQuery.value = ''
  currentPage.value = 1
  sortBy.value = 'time'
  selectionMode.value = false
  selectedIds.value = []
  manageVisible.value = false
  composerVisible.value = false
  dialogVisible.value = false
  sectionMgrVisible.value = false
  brokenCovers.value = new Set()
  sections.value = []
  experienceShares.value = []
  await fetchSections()
  await fetchExperienceShares()
})

/* ---- 贴吧复刻版新增：管理弹窗 / 分享 / 相关推荐 / 摘要 ---- */
const manageVisible = ref(false)
const relBlockHidden = ref(false)

/* 封面加载失败的黑名单：历史数据里存在「本地磁盘路径」（D:\... 浏览器不可达）或已失效的 URL，
   一旦 <img> 报错就记下 id，模板直接不渲染图片区 —— 否则破图按原始尺寸撑爆卡片、把下方文字挤出错位。 */
const brokenCovers = ref<Set<number>>(new Set())
const markCoverBroken = (id: number) => {
  const next = new Set(brokenCovers.value)
  next.add(id)
  brokenCovers.value = next
}

/* 可用的封面地址：本地磁盘路径（C:\ D:\ /Users/…）浏览器取不到，直接判为不可用，
   不必等 <img> 报错（能少一次无效请求，也避免懒加载下面板闪一下）。
   外链（http/https）走同源代理：贴吧等图床有防盗链，直接加载必 403 破图。 */
const coverSrc = (item: ExperienceShare): string | null => {
  const url = (item.cover_url || '').trim()
  if (!url) return null
  // 站内绝对路径（后端落盘的封面）/ data / blob → 直接用
  if (/^(\/|data:|blob:)/i.test(url)) return url
  // 外链 → 交给后端代理（带上原 URL，由后端按目标站 Referer 抓取）
  if (/^https?:\/\//i.test(url)) return `/api/image-proxy?url=${encodeURIComponent(url)}`
  // 形如 C:\... / D:/... / \\server\share → Windows 本地盘路径，不可用
  if (/^[a-zA-Z]:[\\/]/.test(url) || url.startsWith('\\\\')) return null
  // 其余（如相对路径 图片.png）也当不可用，避免相对解析到站点根下乱找
  return null
}
const hasUsableCover = (item: ExperienceShare) =>
  !!coverSrc(item) && !brokenCovers.value.has(item.id)

/* 卡片「管理」菜单：**只能单击触发**（不再随鼠标悬停自动展开）。
   展开后：点同一个按钮收起、点任一菜单项执行并收起、点页面其它地方也收起。
   ⚠ 按钮与菜单项都是 @click.stop，所以文档级监听不会把「点自己」误判成「点外面」。 */
const openManageId = ref<number | null>(null)
const toggleManage = (id: number) => {
  openManageId.value = openManageId.value === id ? null : id
}

const closeManageOnOutsideClick = (e: MouseEvent) => {
  if (openManageId.value === null) return
  const target = e.target as HTMLElement | null
  if (target?.closest('.pa-manage')) return
  openManageId.value = null
}

/* 点赞：「我赞过」存本地（与详情页同一套 key），计数在后端 like_count */
const likedIds = ref<Set<number>>(new Set())
const LIKE_KEY = (id: number) => `bbs-main-like-${id}`
const isLiked = (id: number) => likedIds.value.has(id)
const syncLikedFromStorage = () => {
  const next = new Set<number>()
  experienceShares.value.forEach(item => {
    if (localStorage.getItem(LIKE_KEY(item.id)) === '1') next.add(item.id)
  })
  likedIds.value = next
}

const toggleLike = async (item: ExperienceShare) => {
  const nextLiked = !isLiked(item.id)
  const prev = Number(item.like_count) || 0
  /* 乐观更新：本地先动，失败再回滚 */
  item.like_count = Math.max(0, prev + (nextLiked ? 1 : -1))
  const set = new Set(likedIds.value)
  nextLiked ? set.add(item.id) : set.delete(item.id)
  likedIds.value = set
  try {
    const res = await experienceShareAPI.like(item.id, nextLiked ? 1 : -1)
    if (res.success) {
      item.like_count = res.data?.like_count ?? item.like_count
      localStorage.setItem(LIKE_KEY(item.id), nextLiked ? '1' : '0')
    } else {
      throw new Error('like failed')
    }
  } catch {
    item.like_count = prev
    const rollback = new Set(likedIds.value)
    nextLiked ? rollback.delete(item.id) : rollback.add(item.id)
    likedIds.value = rollback
    ElMessage.error('点赞失败，请确认后端已重启')
  }
}

/* 列表摘要：优先简介，没有简介再把正文剥掉标签后截断 */
const excerptRaw = (item: ExperienceShare) =>
  (item.summary || '').trim() || stripRichText(item.content || '')

const postExcerpt = (item: ExperienceShare) => {
  const raw = excerptRaw(item)
  return raw.length > 140 ? `${raw.slice(0, 140)}…` : raw
}

/* 「…全文」链接的显示条件：正文本身足够长（摘要可能是短简介） */
const excerptTruncated = (item: ExperienceShare) =>
  stripRichText(item.content || '').length > 140

/* 相关推荐：最新的 8 篇（单机吧没有平台推荐，用最近发布充当推荐位） */
const relatedShares = computed(() =>
  [...experienceShares.value]
    .sort((a, b) => (b.created_at || '').localeCompare(a.created_at || ''))
    .slice(0, 8)
)

/* 头像 / 推荐位色块：按 id 稳定取一组渐变，与封面渐变同源 */
const avatarGradient = (item: ExperienceShare) =>
  COVER_GRADIENTS[(item.id || 0) % COVER_GRADIENTS.length]

/* 分享 = 复制详情页链接。
   ⚠ router 是 createWebHistory（没有 #），原来拼的 /#/experience-shares/<id> 是死链。 */
const sharePost = async (item: ExperienceShare) => {
  const url = `${window.location.origin}${tutorialPostPath(item.id)}`
  try {
    await navigator.clipboard.writeText(url)
    ElMessage.success('链接已复制')
  } catch {
    ElMessage.info(url)
  }
}

/* ---- 导入 / 导出 ---- */
/* 下载 JSON：导出三个入口共用（单条 / 选中 / 全部） */
const downloadJson = (payload: unknown, filename: string) => {
  const jsonStr = JSON.stringify(payload, null, 2)
  const blob = new Blob([jsonStr], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

/* 单条 / 选中导出用同一套外壳，与后端 /export 的 export_version 2 对齐 */
const buildExportPayload = (cards: ExperienceShare[]) => ({
  export_version: 2,
  exported_at: new Date().toISOString().replace(/\.\d{3}Z$/, 'Z'),
  boards: ['tutorial', 'experience'],
  cards
})

const todayStamp = () => new Date().toISOString().slice(0, 10)

const handleExport = async () => {
  if (experienceShares.value.length === 0) {
    ElMessage.warning('没有可导出的经验卡片')
    return
  }

  try {
    // 走后端整包导出：两个吧全量，且字段完整（封面地址 / 分区 / 精华 / 点赞 / 创建时间都在）
    const payload = await experienceShareAPI.exportAll()
    downloadJson(payload, `经验帖子导出_${todayStamp()}.json`)
    ElMessage.success(`已导出 ${payload.cards?.length || 0} 篇（两个吧全量）`)
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.message || error.message || '导出失败')
  }
}

const handleExportSingle = (item: ExperienceShare) => {
  try {
    downloadJson(buildExportPayload([item]), `${(item.title || '经验卡片').replace(/[^\w一-龥]/g, '_').slice(0, 50) || '经验卡片'}_${todayStamp()}.json`)
    ElMessage.success('导出成功')
  } catch (error: any) {
    ElMessage.error(error.message || '导出失败')
  }
}

const handleExportSelected = () => {
  if (selectedIds.value.length === 0) {
    ElMessage.warning('请先选择要导出的经验卡片')
    return
  }

  try {
    const selectedItems = experienceShares.value.filter(item => selectedIds.value.includes(item.id))
    downloadJson(buildExportPayload(selectedItems), `经验卡片导出_${todayStamp()}.json`)
    ElMessage.success(`成功导出 ${selectedItems.length} 张经验卡片`)
    cancelSelection()
  } catch (error: any) {
    ElMessage.error(error.message || '导出失败')
  }
}

const importFileInputRef = ref<HTMLInputElement | null>(null)

const handleImport = () => {
  importFileInputRef.value?.click()
}

const handleImportFileChange = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]

  try {
    if (!file) return

    if (!file.name.endsWith('.json')) {
      throw new Error('仅支持 JSON 格式文件')
    }

    const text = await file.text()
    let parsed: unknown
    try {
      parsed = JSON.parse(text)
    } catch {
      throw new Error('文件不是合法的 JSON')
    }

    // 外壳形态交给后端判断（裸数组 / cards / shares / data 四种都认），
    // 前端不再用 Array.isArray 提前拦掉旧版产物。
    const res = await experienceShareAPI.importAll(parsed)
    const result = res.data
    if (!res.success || !result) {
      throw new Error(res.message || '导入失败')
    }

    if (result.failedCount > 0) {
      const detail = result.errors
        .slice(0, 3)
        .map(e => `第 ${e.index + 1} 条「${e.title || '无标题'}」：${e.reason}`)
        .join('；')
      ElMessage({
        type: 'warning',
        message: `成功导入 ${result.insertedCount} 篇，失败 ${result.failedCount} 篇。${detail}${result.failedCount > 3 ? ' …' : ''}`,
        duration: 8000
      })
    } else {
      ElMessage.success(`成功导入 ${result.insertedCount} 篇`)
    }
    await fetchExperienceShares()
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.message || error.message || '导入失败')
  } finally {
    target.value = ''
  }
}
</script>

<style scoped>
.experience-page {
  --xp-accent-line: #bfdbfe;
  --xp-bg: #f4f6f8;
  --xp-panel: #ffffff;
  --xp-panel-2: #f9fafb;
  --xp-text: #1f2328;
  --xp-text-2: #57606a;
  --xp-text-3: #8b949e;
  --xp-border: #e5e7eb;
  --xp-border-2: #d0d7de;
  --xp-chip: #f1f3f5;
  --xp-radius: 12px;
  --xp-shadow: 0 1px 2px rgba(16, 24, 40, .05);
  --xp-banner-1: var(--accent, #2563eb);
  --xp-banner-2: #0ea5e9;
  --xp-banner-3: #14b8a6;
  --xp-floor: #fafbfc;
  background: var(--xp-bg);
  position: relative;
  min-height: 100%;
}

:root[data-theme='dark'] .experience-page {
  --xp-accent-line: #2c4a6b;
  --xp-bg: #16181d;
  --xp-panel: #1f2937;
  --xp-panel-2: #1a2230;
  --xp-text: #e6edf3;
  --xp-text-2: #9ca3af;
  --xp-text-3: #6e7681;
  --xp-border: #374151;
  --xp-border-2: #374151;
  --xp-chip: #232b3a;
  --xp-shadow: 0 1px 2px rgba(0, 0, 0, .35);
  --xp-banner-1: var(--accent, #2563eb);
  --xp-banner-2: #0c4a6e;
  --xp-banner-3: #134e4a;
  --xp-floor: #1a2230;
}

/* ---------- 页面骨架（内容列在 rail 右侧剩余空间内居中，两边留白） ---------- */
.page {
  margin: 0 auto 60px;
  padding: 16px 20px 0;
  max-width: 1220px;
}

/* ---------- 横幅吧头 ---------- */
.bar-banner {
  position: relative;
  height: 188px;
  border-radius: 16px;
  overflow: hidden;
  background:
    radial-gradient(560px 240px at 88% -30%, rgba(255, 255, 255, .38), transparent 62%),
    radial-gradient(420px 220px at 12% 120%, rgba(255, 255, 255, .16), transparent 60%),
    linear-gradient(112deg, var(--xp-banner-1) 0%, var(--xp-banner-2) 58%, var(--xp-banner-3) 100%);
  box-shadow: var(--xp-shadow);
}
/* 斜向装饰条（对应截图里的流光纹理） */
.bar-banner::before {
  content: '';
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(115deg,
    transparent 0 90px,
    rgba(255, 255, 255, .07) 90px 150px,
    transparent 150px 260px,
    rgba(255, 255, 255, .10) 260px 300px);
}
.bar-banner::after {
  content: '✦';
  position: absolute;
  right: 130px;
  top: -46px;
  font-size: 190px;
  color: rgba(255, 255, 255, .22);
  transform: rotate(18deg);
  text-shadow: 0 0 60px rgba(255, 255, 255, .35);
}
.bb-bottom {
  position: absolute;
  left: 0; right: 0; bottom: 0;
  display: flex;
  align-items: flex-end;
  gap: 14px;
  padding: 0 22px 16px;
}
.bb-av {
  width: 72px; height: 72px;
  border-radius: 18px;
  flex: none;
  border: 3px solid rgba(255, 255, 255, .9);
  background: linear-gradient(135deg, #60a5fa, #2563eb);
  color: #fff;
  font-size: 28px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 14px rgba(0, 0, 0, .18);
}
.bb-name { flex: 1; min-width: 0; color: #fff; padding-bottom: 4px; }
.bb-name h1 {
  margin: 0;
  font-size: 21px;
  font-weight: 600;
  letter-spacing: .5px;
  text-shadow: 0 1px 8px rgba(0, 0, 0, .25);
}
.bb-name p {
  margin: 3px 0 0;
  font-size: 12.5px;
  color: rgba(255, 255, 255, .88);
  text-shadow: 0 1px 6px rgba(0, 0, 0, .22);
}
.bb-actions { display: flex; gap: 10px; padding-bottom: 4px; }
.bb-btn {
  border: none;
  border-radius: 999px;
  padding: 8px 20px;
  font-size: 13.5px;
  font-family: inherit;
  cursor: pointer;
  transition: transform .12s ease, box-shadow .12s ease;
}
.bb-btn:active { transform: scale(.97); }
.bb-btn.ghost {
  background: rgba(255, 255, 255, .22);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, .55);
  backdrop-filter: blur(4px);
}
.bb-btn.ghost:hover { background: rgba(255, 255, 255, .32); }
.bb-btn.primary {
  background: #fff;
  color: var(--xp-banner-1);
  font-weight: 600;
  box-shadow: 0 2px 10px rgba(0, 0, 0, .16);
}
.bb-btn.primary:hover { box-shadow: 0 4px 14px rgba(0, 0, 0, .22); }
:root[data-theme='dark'] .bb-btn.primary { color: #1e3a8a; }

/* ---------- 标签页行 ---------- */
.tabs-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 4px 0;
}
.tabs {
  display: flex;
  gap: 4px;
  overflow-x: auto;
  scrollbar-width: none;
  min-width: 0;
}
.tabs::-webkit-scrollbar { display: none; }
.tab {
  position: relative;
  border: none;
  background: transparent;
  font-family: inherit;
  font-size: 14px;
  color: var(--xp-text-2);
  padding: 8px 13px 10px;
  cursor: pointer;
  white-space: nowrap;
  border-radius: 8px 8px 0 0;
}
.tab:hover { color: var(--xp-text); background: var(--xp-chip); }
.tab .tn {
  font-size: 11px;
  color: var(--xp-text-3);
  margin-left: 3px;
  font-variant-numeric: tabular-nums;
}
.tab.on { color: var(--accent); font-weight: 600; }
.tab.on .tn { color: var(--accent); }
.tab.on::after {
  content: '';
  position: absolute;
  left: 13px; right: 13px; bottom: 4px;
  height: 3px;
  border-radius: 3px;
  background: var(--accent);
}
.tr-right { margin-left: auto; display: flex; align-items: center; gap: 10px; flex: none; }
.sorts { display: flex; align-items: center; gap: 2px; }
.sort {
  border: none;
  background: transparent;
  font-family: inherit;
  font-size: 12.5px;
  color: var(--xp-text-3);
  padding: 5px 9px;
  border-radius: 6px;
  cursor: pointer;
  white-space: nowrap;
}
.sort:hover { background: var(--xp-chip); color: var(--xp-text-2); }
.sort.on { color: var(--accent); font-weight: 600; background: var(--accent-soft); }
.search {
  display: flex;
  align-items: center;
  gap: 6px;
  background: var(--xp-panel);
  border: 1px solid var(--xp-border-2);
  border-radius: 999px;
  padding: 6px 12px;
  width: 200px;
  color: var(--xp-text-3);
}
.search input {
  border: none;
  outline: none;
  background: transparent;
  font-family: inherit;
  font-size: 12.5px;
  color: var(--xp-text);
  width: 100%;
}
.search input::placeholder { color: var(--xp-text-3); }

/* ---------- 两栏 ---------- */
.cols {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 300px;
  gap: 18px;
  align-items: start;
  margin-top: 12px;
}
/* 帖子流：**没有容器、没有卡片外壳、没有分隔线**。
   帖子直接坐在页面底色上（贴吧就是这样），只有鼠标悬停 / 选中时才浮出一个块。
   ⚠ 不要在 .feed 上加 border/圆角/背景，一旦有容器就会和页面底色分层。 */
.feed {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
  background: transparent;
  border: none;
  border-radius: 0;
  box-shadow: none;
}

/* 帖子本体：平时无形状 */
.feed .post {
  border: none;
  border-radius: 10px;
  box-shadow: none;
  background: transparent;
  transition: background .14s ease;
}

/* 悬停 / 选中才出现形状 */
.feed .post:hover {
  background: var(--xp-panel);
}

/* 选中态：只要一层淡淡的底色（60% 不透明），不要描边、不要边框 —— 描边太"重" */
.feed .post.is-selected {
  background: color-mix(in srgb, var(--xp-panel) 60%, transparent);
}
.card {
  background: var(--xp-panel);
  border: 1px solid var(--xp-border);
  border-radius: var(--xp-radius);
  box-shadow: var(--xp-shadow);
}

/* ---------- 置顶区（贴吧式：独立区块 + 蓝框小标签 + 纯标题） ---------- */
/* 与下面的帖子列表同底色（透明，跟页面一致），中间只隔一条细线 */
.pin-box {
  padding: 4px 10px;
  border-bottom: 1px solid var(--xp-border);
}

.pin-row {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 38px;
  padding: 0 8px;
  border-radius: 6px;
  cursor: pointer;
  transition: background .12s ease;
}

.pin-row:hover { background: var(--xp-panel); }

.pin-row:focus-visible {
  outline: 2px solid var(--xp-accent-line);
  outline-offset: -2px;
}

/* 蓝框标签：贴吧那种 12px 小药丸，不要用黄色（黄色在本项目是「精华」） */
.pin-tag {
  flex: none;
  font-size: 12px;
  line-height: 16px;
  padding: 0 4px;
  color: #1e6fbf;
  background: #f0f7ff;
  border: 1px solid #b8d4f0;
  border-radius: 2px;
}

.pin-text {
  flex: 1;
  min-width: 0;
  font-size: 15px;
  line-height: 20px;
  color: var(--xp-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

:root[data-theme='dark'] .pin-tag {
  color: #93c5fd;
  background: #16273d;
  border-color: #2c4a6b;
}

:root[data-theme='dark'] .pin-row:hover { background: #232b3a; }

/* ---------- 帖子卡 ---------- */
.post { padding: 15px 16px 10px; cursor: pointer; transition: box-shadow .15s ease, border-color .15s ease; }
.post:hover { border-color: var(--xp-accent-line); box-shadow: 0 3px 14px rgba(16, 24, 40, .08); }
:root[data-theme='dark'] .post:hover { box-shadow: 0 3px 14px rgba(0, 0, 0, .4); }
.p-head { display: flex; align-items: center; gap: 10px; }
.p-av {
  width: 38px; height: 38px;
  border-radius: 50%;
  flex: none;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}
.av-1 { background: linear-gradient(135deg, #60a5fa, #2563eb); }
.av-2 { background: linear-gradient(135deg, #34d399, #059669); }
.av-3 { background: linear-gradient(135deg, #fbbf24, #d97706); }
.av-4 { background: linear-gradient(135deg, #a78bfa, #7c3aed); }
.av-5 { background: linear-gradient(135deg, #f472b6, #db2777); }
.p-who { flex: 1; min-width: 0; }
.p-name { font-size: 13px; font-weight: 600; color: var(--xp-text-2); }
.p-date { font-size: 11.5px; color: var(--xp-text-3); }
.p-sec {
  flex: none;
  font-size: 11.5px;
  color: var(--xp-text-2);
  background: var(--xp-chip);
  border: 1px solid var(--xp-border);
  border-radius: 999px;
  padding: 2px 10px;
  cursor: pointer;
}
.p-sec:hover { border-color: var(--accent); color: var(--accent); }

.p-title { margin: 9px 0 0; font-size: 15.5px; font-weight: 650; line-height: 1.45; display: flex; align-items: baseline; gap: 8px; flex-wrap: wrap; }
.badge {
  flex: none;
  font-size: 11px;
  font-weight: 600;
  border-radius: 5px;
  padding: 1px 7px;
  transform: translateY(-1px);
}
.bd-feat { color: #92400e; background: #fef3c7; border: 1px solid #fde68a; }
:root[data-theme='dark'] .bd-feat { color: #fcd34d; background: #2b2617; border-color: #5c4a1e; }
.bd-pdf { color: #1d4ed8; background: #dbeafe; border: 1px solid #bfdbfe; }
:root[data-theme='dark'] .bd-pdf { color: #93c5fd; background: #16273d; border-color: #2c4a6b; }
.bd-help { color: #047857; background: #d1fae5; border: 1px solid #a7f3d0; }
:root[data-theme='dark'] .bd-help { color: #6ee7b7; background: #12291f; border-color: #1f5c44; }

/* 正文：多行展示，长文收「…全文」 */
.p-content {
  margin: 6px 0 0;
  font-size: 13.5px;
  color: var(--xp-text-2);
  line-height: 1.75;
}
.p-full { color: var(--accent); text-decoration: none; cursor: pointer; white-space: nowrap; }
.p-full:hover { text-decoration: underline; }

/* 图片区：单图 = 一张大图；多图 = 三列方图网格，尾图可带 +N */
.p-media { margin-top: 10px; }
/* 列表里的配图是缩略图，不是大图：尺寸固定、object-fit: cover 裁切，
   铺满整行会让一屏只放得下一篇（贴吧也是小图）。 */
.p-media.single .ph {
  width: 220px;
  height: 165px;
  border-radius: 8px;
  overflow: hidden;
}
.p-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}
.p-grid .ph {
  position: relative;
  width: 100%;
  aspect-ratio: 1 / 1;
  border-radius: 12px;
  box-shadow: var(--xp-shadow);
  overflow: hidden;
}
.ph { display: block; background: var(--xp-chip); width: 100%; height: 100%; object-fit: cover; }
.p-more {
  position: absolute;
  right: 8px;
  bottom: 8px;
  background: rgba(0, 0, 0, .55);
  color: #fff;
  font-size: 12px;
  padding: 1px 9px;
  border-radius: 999px;
  backdrop-filter: blur(2px);
}

/* 操作栏：分享 / 评论 / 赞 / 管理 四项等距铺满。
   ⚠ 不要 border-top —— 帖子本来就是无痕的，中间横一条线又把形状勾出来了。 */
.p-acts {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 10px;
  padding-top: 4px;
  color: var(--xp-text-3);
}
.pa {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: none;
  background: transparent;
  font-family: inherit;
  font-size: 12.5px;
  color: var(--xp-text-3);
  padding: 6px 14px;
  border-radius: 7px;
  cursor: pointer;
}
.pa:hover { background: var(--xp-chip); color: var(--xp-text-2); }
.pa svg { width: 15px; height: 15px; }
/* 已赞：跟随皮肤的实心色（--accent 在亮色是深蓝、暗色是亮蓝，作文字色均达标） */
.pa-like.liked { color: var(--accent); font-weight: 600; }
.pa-like.liked svg { fill: currentColor; }
.pa-manage { position: relative; display: inline-flex; }
.pa-pop {
  position: absolute;
  right: 0;
  bottom: calc(100% + 8px);
  background: var(--xp-panel);
  border: 1px solid var(--xp-border-2);
  border-radius: 10px;
  box-shadow: 0 8px 24px rgba(16, 24, 40, .16);
  padding: 4px;
  display: none;
  min-width: 104px;
  z-index: 6;
  cursor: default;
}
:root[data-theme='dark'] .pa-pop { box-shadow: 0 8px 24px rgba(0, 0, 0, .5); }
/* 展开只由单击驱动：悬停不再展开（.pa-pop 默认 display:none，靠 .open 显示） */
.pa-pop.open { display: block; }
.po {
  display: block;
  width: 100%;
  border: none;
  background: transparent;
  font-family: inherit;
  font-size: 12px;
  color: var(--xp-text-2);
  padding: 6px 10px;
  border-radius: 7px;
  cursor: pointer;
  text-align: left;
  white-space: nowrap;
}
.po:hover { background: var(--xp-chip); }
.po.feat.on { color: #b45309; font-weight: 700; }
.po.pin.on { color: var(--xp-banner-1, #2563eb); font-weight: 700; }
:root[data-theme='dark'] .po.feat.on { color: #fcd34d; }
.po.danger:hover { color: #dc2626; }
:root[data-theme='dark'] .po.danger:hover { color: #f87171; }

/* ---------- 右栏 ---------- */
/* 单一独立容器：吧信息 + 相关推荐 两个分区共存、共同滚动；
   容器 sticky 且不超过视口高，页面怎么滚都能在容器里直接滑（管理已迁至顶部「管理」弹窗） */
.rail {
  position: sticky;
  top: 14px;
  max-height: calc(100vh - 28px);
}
.side-all {
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 28px);
  overflow: hidden;
}
.side-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-width: none; /* Firefox：隐藏滚动条，滚轮滚动保留 */
}
.side-scroll::-webkit-scrollbar { display: none; } /* Chrome/Edge：无痕 */
/* 分区通用头（滚动流内，跟随滚动） */
.sec-head { display: flex; align-items: center; justify-content: space-between; padding: 12px 14px 0; }
.sec-head h3 { margin: 0; padding: 0; font-size: 15px; font-weight: 600; color: var(--xp-text); }
.bar-info { padding-bottom: 14px; border-bottom: 1px solid var(--xp-border); }
.bi-row { display: flex; align-items: center; gap: 10px; padding: 10px 14px 0; }
.bi-av {
  width: 44px; height: 44px;
  border-radius: 12px;
  background: linear-gradient(135deg, #60a5fa, #2563eb);
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: none;
}
.bi-name { font-size: 14.5px; font-weight: 600; }
.bi-sub { font-size: 11.5px; color: var(--xp-text-3); }
.bi-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  padding: 12px 14px 0;
  gap: 6px;
}
.bi-stats div { text-align: center; background: var(--xp-panel-2); border-radius: 8px; padding: 7px 2px; }
.bi-stats b { display: block; font-size: 15px; font-variant-numeric: tabular-nums; }
.bi-stats span { font-size: 10.5px; color: var(--xp-text-3); }
.bi-desc { margin: 10px 14px 0; font-size: 12px; color: var(--xp-text-2); line-height: 1.7; }
.bi-meta { margin: 8px 14px 0; font-size: 11.5px; color: var(--xp-text-3); }

/* 右栏 · 相关推荐（对照贴吧「相关推荐」实拍：标题行 + 摘要 + 可选横图 + 来源chip + N评论 N点赞）
   作为滚动流中的一个分区，与吧信息共用 .side-scroll 滚动 */
.rel { padding-bottom: 2px; }
.rel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px 0;
}
.rel-head h3 { margin: 0; padding: 0; font-size: 15px; font-weight: 600; color: var(--xp-text); }
.rel-close {
  border: none;
  background: transparent;
  color: var(--xp-text-3);
  font-size: 14px;
  line-height: 1;
  padding: 3px 5px;
  border-radius: 6px;
  cursor: pointer;
}
.rel-close:hover { color: var(--xp-text-2); background: var(--xp-chip); }
.rl-item { display: block; padding: 10px 14px 9px; cursor: pointer; color: inherit; text-decoration: none; }
.rl-item + .rl-item { border-top: 1px solid var(--xp-border); }
.rl-item:hover { background: var(--xp-panel-2); }
.rl-t {
  font-size: 13px;
  font-weight: 550;
  color: var(--xp-text);
  line-height: 1.55;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.rl-x {
  margin-top: 3px;
  font-size: 12px;
  color: var(--xp-text-2);
  line-height: 1.65;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.rl-img {
  margin-top: 8px;
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: 8px;
  background: var(--xp-chip);
  box-shadow: var(--xp-shadow);
  overflow: hidden; /* 关键：无此则 img 按原始尺寸溢出，压住下一条卡片 */
}
.rl-img img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.rl-foot { margin-top: 7px; display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.rl-chip { display: inline-flex; align-items: center; gap: 5px; min-width: 0; font-size: 11px; color: var(--xp-text-3); }
.rl-chip i { width: 12px; height: 12px; border-radius: 4px; flex: none; }
.rl-chip span { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.rl-nums { flex: none; font-size: 11px; color: var(--xp-text-3); font-variant-numeric: tabular-nums; }

/* ---------- 管理弹窗（从右栏迁入，由横幅「管理」按钮唤起） ---------- */
.mask {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, .45);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 24px;
}
:root[data-theme='dark'] .mask { background: rgba(0, 0, 0, .62); }
.modal {
  width: min(420px, calc(100vw - 48px));
  background: var(--xp-panel);
  border: 1px solid var(--xp-border-2);
  border-radius: 14px;
  box-shadow: 0 18px 48px rgba(16, 24, 40, .22);
  overflow: hidden;
}
:root[data-theme='dark'] .modal { box-shadow: 0 18px 48px rgba(0, 0, 0, .6); }
.mo-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px 12px;
  border-bottom: 1px solid var(--xp-border);
}
.mo-head h3 { margin: 0; font-size: 15px; font-weight: 600; color: var(--xp-text); }
.mo-body { padding: 6px 0 10px; }
.mg {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  border: none;
  background: transparent;
  font-family: inherit;
  font-size: 13px;
  color: var(--xp-text-2);
  padding: 10px 16px;
  cursor: pointer;
  text-align: left;
}
.mg:hover { background: var(--accent-soft); color: var(--accent); }
.mg .cnt { margin-left: auto; font-size: 11px; color: var(--xp-text-3); font-style: normal; }
.mo-tip { margin: 0; padding: 4px 16px 0; font-size: 11.5px; color: var(--xp-text-3); }

/* ============ 以下沿用原实现：批量移动 / 表单分区 / 分区管理 / 空状态 / 底栏 / 弹窗 ============ */

/* ---------- 批量移动分区气泡 ---------- */
.batch-move {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.bm-title {
  font-size: 13px;
  color: var(--xp-text-2);
}

.bm-ops {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

/* ---------- 表单里的分区 / 精华 ---------- */
.sec-form-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.sec-select {
  width: 180px;
}

.sec-form-hint {
  font-size: 12px;
  color: var(--xp-text-3);
}

/* ---------- 分区管理弹窗 ---------- */
.sec-mgr {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.sec-add-row {
  display: flex;
  gap: 8px;
}

.sec-add-row :deep(.el-input) {
  flex: 1;
}

.sec-tip {
  font-size: 12px;
  line-height: 1.7;
  color: var(--xp-text-3);
}

.sec-list {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--xp-border);
  border-radius: 6px;
  overflow: hidden;
}

.sec-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--xp-border);
}

.sec-row:last-child {
  border-bottom: none;
}

.sec-row:hover {
  background: var(--xp-floor);
}

.sec-name {
  font-size: 13.5px;
  color: var(--xp-text);
  font-weight: 500;
}

.sec-count {
  font-size: 12px;
  color: var(--xp-text-3);
  font-variant-numeric: tabular-nums;
}

.sec-ops {
  margin-left: auto;
  display: flex;
  gap: 2px;
}

.sec-ops :deep(.el-button) {
  font-size: 12px;
}

.sec-empty {
  padding: 22px 12px;
  text-align: center;
  font-size: 13px;
  color: var(--xp-text-3);
}

/* 过滤后为空 */
.t-filtered-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 60px 0;
  color: var(--xp-text-3);
  text-align: center;
}

.t-filtered-empty .fe-ic {
  font-size: 34px;
  opacity: 0.55;
}

.t-filtered-empty .fe-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--xp-text-2);
}

.t-filtered-empty .fe-sub {
  font-size: 12px;
  margin-bottom: 6px;
}

/* ---------- 底栏：统计 + 分页 ---------- */
.list-foot {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 2px 2px 0;
  font-size: 12px;
  color: var(--xp-text-3);
  min-height: 40px;
}

.stats-left,
.stats-right {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.stat-title {
  color: var(--xp-text-2);
  font-weight: 600;
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.selection-count {
  color: var(--xp-text-2);
  font-size: 13px;
  font-weight: 600;
}

.stats-hint {
  color: var(--xp-text-3);
  font-size: 11px;
  white-space: nowrap;
}

.list-foot :deep(.el-pagination) {
  --el-pagination-button-bg-color: transparent;
  --el-pagination-hover-color: var(--accent);
  --el-pagination-button-color: var(--xp-text-2);
}

.list-foot :deep(.el-pager li.is-active) {
  background: var(--accent);
  color: #ffffff;
  border-radius: 4px;
}

.list-foot :deep(.el-pagination .btn-prev),
.list-foot :deep(.el-pagination .btn-next) {
  background: transparent;
}

.list-foot :deep(.el-checkbox__label) {
  font-size: 13px;
  font-weight: 500;
  color: var(--xp-text-2);
}

/* 空状态插图 */
.empty-illustration {
  width: 100px;
  height: 100px;
  border-radius: 24px;
  background: var(--accent-soft);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
}

.empty-illustration .el-icon {
  font-size: 40px;
  color: var(--accent);
  opacity: 0.7;
}

/* ============ 弹窗（沿用原实现） ============ */
.dialog-body {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.section-alert {
  margin-top: 4px;
}

.import-panel {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 18px 20px;
  border-radius: 6px;
  background: #f8fafc;
  border: 1px solid rgb(224, 224, 230);
}

.import-title {
  font-size: 16px;
  font-weight: 700;
  color: #7c4700;
}

.import-desc {
  margin-top: 6px;
  color: #8b5e13;
  font-size: 13px;
  line-height: 1.7;
}

.import-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.import-file-name {
  max-width: 320px;
  color: #7c4700;
  font-size: 13px;
  word-break: break-all;
}

.experience-form {
  margin-top: 4px;
}

.experience-form :deep(.el-form-item__label) {
  font-size: 14px;
  font-weight: 400;
  color: #1f2225;
  text-align: right;
  justify-content: flex-end;
}

.experience-form :deep(.el-input__wrapper) {
  border-radius: 8px;
  box-shadow: 0 0 0 1px #e5e7eb inset;
  background-color: #ffffff;
  transition: box-shadow 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.experience-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px #c0c4cc inset;
}

.experience-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px var(--accent-soft, rgba(37, 99, 235, 0.15)), 0 0 0 1px var(--accent, #2563eb) inset;
}

.experience-form :deep(.el-textarea__inner) {
  border-radius: 8px;
  box-shadow: 0 0 0 1px #e5e7eb inset;
  transition: box-shadow 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.experience-form :deep(.el-textarea__inner:hover) {
  box-shadow: 0 0 0 1px #c0c4cc inset;
}

.experience-form :deep(.el-textarea__inner:focus) {
  box-shadow: 0 0 0 2px var(--accent-soft, rgba(37, 99, 235, 0.15)), 0 0 0 1px var(--accent, #2563eb) inset;
}

.content-editor {
  width: 100%;
  max-height: 820px;
}

/* 正文编辑器跟随应用主题（组件自身用 prefers-color-scheme 跟系统，和应用内主题开关脱节，
   系统深色 + 应用亮色时会出现黑编辑器压白弹窗）。仅在本弹窗范围内覆盖，不影响其他页面。 */
.experience-form :deep(.tiptap-editor) {
  border-radius: 10px;
  border-color: #e5e7eb;
  background: #ffffff;
}
.experience-form :deep(.editor-toolbar) {
  background: #fafafa;
  border-bottom-color: #e5e7eb;
}
.experience-form :deep(.toolbar-btn) { color: #4b5563; }
.experience-form :deep(.toolbar-btn:hover:not(:disabled)) { background: #eef2ff; color: var(--accent, #2563eb); }
.experience-form :deep(.toolbar-btn.is-active) { background: var(--accent-soft, rgba(37, 99, 235, 0.1)); color: var(--accent, #2563eb); }
.experience-form :deep(.toolbar-divider) { background: #e5e7eb; }
.experience-form :deep(.heading-btn),
.experience-form :deep(.color-btn) { color: #4b5563; }
.experience-form :deep(.tiptap-editor:not(.is-ghost) .editor-content) { background: #ffffff; }
.experience-form :deep(.tiptap-editor:not(.is-ghost) .editor-content .tiptap p) { color: #1f2937; }
.experience-form :deep(.tiptap-editor:not(.is-ghost) .editor-content .tiptap h1),
.experience-form :deep(.tiptap-editor:not(.is-ghost) .editor-content .tiptap h2) { color: #111827; }

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

.cover-preview-actions {
  display: flex;
  justify-content: center;
  gap: 10px;
  padding: 12px;
  background: #f8fafc;
  border-top: 1px solid rgba(148, 163, 184, 0.12);
}

.cover-upload-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  max-width: 420px;
  height: 180px;
  border-radius: 6px;
  border: 2px dashed rgb(224, 224, 230);
  background: #fafafa;
  cursor: pointer;
  transition: border-color 0.2s cubic-bezier(0.4, 0, 0.2, 1), background 0.2s ease;
}

.cover-upload-area:hover {
  border-color: #36ad6a;
  background: rgba(24, 160, 88, 0.04);
}

.cover-upload-icon {
  font-size: 36px;
  color: #94a3b8;
}

.cover-upload-text {
  font-size: 14px;
  font-weight: 500;
  color: #475569;
}

.cover-upload-hint {
  font-size: 12px;
  color: #94a3b8;
}

.cover-url-input {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed rgba(148, 163, 184, 0.3);
}

.cover-url-label {
  font-size: 13px;
  color: #64748b;
}

.pdf-content-preview {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.pdf-content-preview__info {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 14px 16px;
  border-radius: 6px;
  color: #4b5b58;
  background: #f8fafc;
  border: 1px solid rgb(224, 224, 230);
}

.pdf-content-preview__frame {
  overflow: hidden;
  max-height: min(68vh, 760px);
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
  border-radius: 6px;
  border: 1px solid rgb(224, 224, 230);
  background: #ffffff;
}

.pdf-content-preview__empty {
  padding: 18px;
  border-radius: 6px;
  color: #64748b;
  background: #f8fafc;
  border: 1px dashed rgb(224, 224, 230);
}

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

.pdf-preview-file,
.pdf-preview-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.pdf-preview-file {
  color: #0f172a;
  font-weight: 600;
}

.pdf-preview-body {
  overflow-y: auto;
  max-height: calc(82vh - 74px);
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
  padding-right: 2px;
}

.form-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.form-footer-left {
  display: flex;
  gap: 10px;
}

.form-footer-right {
  display: flex;
  gap: 10px;
}

.preview-content {
  padding: 0;
}

.preview-card {
  border-radius: 6px;
  border: 1px solid rgb(224, 224, 230);
  background: #ffffff;
  overflow: hidden;
}

.preview-cover {
  width: 100%;
  max-height: 300px;
  overflow: hidden;
}

.preview-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preview-body {
  padding: 20px;
}

.preview-title {
  font-size: 20px;
  font-weight: 600;
  color: #1f2225;
  margin: 0 0 12px 0;
  line-height: 1.4;
}

.preview-summary {
  font-size: 14px;
  color: #64748b;
  margin: 0 0 16px 0;
  line-height: 1.6;
}

.preview-main {
  font-size: 14px;
  color: #334155;
  line-height: 1.8;
}

.preview-html,
.preview-markdown {
  min-height: 100px;
}

.preview-empty {
  color: #94a3b8;
  text-align: center;
  padding: 40px 0;
}

/* ---------- 贴吧发帖风弹窗（2026-10-10 复刻） ----------
   ⚠ 弹窗 append-to-body，拿不到 .experience-page 上的 --xp-* 变量，
   只能用字面量颜色 + 显式暗色覆盖（见文末暗色补充段）。 */
.tb-dlg-tabs {
  display: flex;
  align-items: center;
  gap: 22px;
}
.tb-tab {
  position: relative;
  font-size: 17px;
  font-weight: 600;
  color: #6b7280;
  padding: 4px 2px 10px;
  cursor: default;
}
.tb-tab.on { color: var(--accent, #2563eb); }
.tb-tab.on::after {
  content: '';
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  bottom: 0;
  width: 30px;
  height: 3px;
  border-radius: 2px;
  background: var(--accent, #2563eb);
}

.tb-post-to {
  display: flex;
  align-items: center;
  gap: 14px;
}
.tb-row-label {
  width: 96px;
  flex: none;
  font-size: 14px;
  color: #1f2225;
  text-align: right;
}
.tb-bar-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #374151;
}
.tb-bar-av {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: linear-gradient(135deg, #60a5fa, #2563eb);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  font-style: normal;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

/* 版块分区 chips（贴吧样式：灰底胶囊，选中亮蓝） */
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

/* 内容声明（设为精华）行 */
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

/* 发布按钮：贴吧蓝色胶囊 */
.tb-publish {
  padding-left: 34px;
  padding-right: 34px;
  font-weight: 600;
}

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

/* ============ 响应式 ============ */
@media (max-width: 1080px) {
  .cols { grid-template-columns: 1fr; }
  .rail { position: static; max-height: none; }
  .side-all { max-height: none; }
}

@media (max-width: 760px) {
  .tabs-row { flex-wrap: wrap; row-gap: 4px; }
  .tr-right { width: 100%; margin-left: 0; justify-content: space-between; }
  .tr-search { width: 200px; }
  .list-foot { flex-wrap: wrap; }
  .bb-name h1 { font-size: 18px; }
  .bb-av { width: 58px; height: 58px; font-size: 22px; }
}

/* ============ 暗色补充（新骨架里无 mockup 对应的规则） ============ */
:root[data-theme='dark'] .tr-search :deep(.el-input__wrapper) {
  background: #111827;
}

:root[data-theme='dark'] .sec-list,
:root[data-theme='dark'] .sec-row {
  border-color: #374151;
}

:root[data-theme='dark'] .badge.bd-manual {
  background: #16301f;
  color: #6ee7a8;
}

:root[data-theme='dark'] .post.is-selected {
  background: color-mix(in srgb, var(--xp-panel) 60%, transparent);
}

/* ---------- 贴吧发帖风弹窗 · 暗色覆盖 ---------- */
:root[data-theme='dark'] .tb-tab { color: #9ca3af; }
:root[data-theme='dark'] .tb-tab.on { color: var(--accent, #58a6ff); }
:root[data-theme='dark'] .experience-form :deep(.el-form-item__label) { color: #e5e7eb; }
:root[data-theme='dark'] .tb-row-label { color: #e5e7eb; }
:root[data-theme='dark'] .tb-bar-chip { color: #d1d5db; }
:root[data-theme='dark'] .sec-chip { background: #262d36; color: #d1d5db; }
:root[data-theme='dark'] .sec-chip:hover { background: #2d3540; }
:root[data-theme='dark'] .sec-chip.on {
  background: rgba(88, 166, 255, 0.16);
  color: var(--accent, #58a6ff);
}
:root[data-theme='dark'] .declare-hint { color: #6b7280; }
:root[data-theme='dark'] .experience-form :deep(.el-input__wrapper) {
  background-color: #111827;
  box-shadow: 0 0 0 1px #374151 inset;
}
:root[data-theme='dark'] .experience-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px #4b5563 inset;
}
:root[data-theme='dark'] .experience-form :deep(.el-textarea__inner) {
  background: #111827;
  box-shadow: 0 0 0 1px #374151 inset;
}
:root[data-theme='dark'] .experience-form :deep(.el-textarea__inner:hover) {
  box-shadow: 0 0 0 1px #4b5563 inset;
}
:root[data-theme='dark'] .import-panel { background: #111827; border-color: #374151; }
:root[data-theme='dark'] .import-title { color: #fcd34d; }
:root[data-theme='dark'] .import-desc { color: #d1d5db; }
:root[data-theme='dark'] .cover-preview-card { background: #111827; border-color: #374151; }
:root[data-theme='dark'] .cover-upload-area { background: #111827; border-color: #374151; }
:root[data-theme='dark'] .cover-upload-text { color: #d1d5db; }
:root[data-theme='dark'] .cover-upload-hint { color: #6b7280; }
:root[data-theme='dark'] .preview-title { color: #e5e7eb; }
:root[data-theme='dark'] .preview-main { color: #d1d5db; }

/* 暗色下编辑器跟随应用暗色（仍只在本弹窗范围） */
:root[data-theme='dark'] .experience-form :deep(.tiptap-editor) {
  background: #111827;
  border-color: #374151;
}
:root[data-theme='dark'] .experience-form :deep(.editor-toolbar) {
  background: #0d1420;
  border-bottom-color: #374151;
}
:root[data-theme='dark'] .experience-form :deep(.toolbar-btn) { color: #9ca3af; }
:root[data-theme='dark'] .experience-form :deep(.toolbar-btn:hover:not(:disabled)) { background: rgba(88, 166, 255, 0.12); color: var(--accent, #58a6ff); }
:root[data-theme='dark'] .experience-form :deep(.toolbar-btn.is-active) { background: rgba(88, 166, 255, 0.16); color: var(--accent, #58a6ff); }
:root[data-theme='dark'] .experience-form :deep(.toolbar-divider) { background: #374151; }
:root[data-theme='dark'] .experience-form :deep(.heading-btn),
:root[data-theme='dark'] .experience-form :deep(.color-btn) { color: #9ca3af; }
:root[data-theme='dark'] .experience-form :deep(.tiptap-editor:not(.is-ghost) .editor-content) { background: #111827; }
:root[data-theme='dark'] .experience-form :deep(.tiptap-editor:not(.is-ghost) .editor-content .tiptap p) { color: #e5e7eb; }
:root[data-theme='dark'] .experience-form :deep(.tiptap-editor:not(.is-ghost) .editor-content .tiptap h1),
:root[data-theme='dark'] .experience-form :deep(.tiptap-editor:not(.is-ghost) .editor-content .tiptap h2) { color: #f3f4f6; }

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
</style>
