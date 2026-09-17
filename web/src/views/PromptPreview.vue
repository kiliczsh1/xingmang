<template>
  <div class="prompt-preview-container">
    <div class="header">
      <h2>提示词预览</h2>
      <div class="search-box">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索提示词"
          clearable
          prefix-icon="Search"
          class="search-input"
        />
      </div>
    </div>

    <!-- 分类选择 -->
    <div class="category-selector">
      <span class="category-title">分类：</span>
      <div class="category-buttons-wrapper">
        <div class="category-buttons">
          <el-button
            v-for="category in ['all', ...categoryList.map(c => c.name)]"
            :key="category"
            :type="selectedCategory === category ? 'primary' : ''"
            size="small"
            @click="selectedCategory = category"
          >
            {{ category === 'all' ? '全部' : category }}
          </el-button>
        </div>
      </div>
    </div>

    <!-- 排序 Tab -->
    <div class="sort-tab-bar">
      <div class="sort-tabs">
        <span
          v-for="tab in sortTabs"
          :key="tab.value"
          class="sort-tab-item"
          :class="{ active: selectedSort === tab.value }"
          @click="selectedSort = tab.value"
        >
          {{ tab.label }}
        </span>
      </div>
    </div>

    <!-- 提示词列表：翻转卡片（正面轻量信息 → 背面就地编辑） -->
    <div class="prompts-list">
      <div v-if="filteredPrompts.length === 0" class="empty-hint">
        <el-empty description="暂无提示词" />
      </div>
      <div v-else class="prompts-grid">
        <div
          v-for="prompt in filteredPrompts"
          :key="prompt.id"
          class="card-wrap"
        >
          <div
            class="prompt-card"
            :class="{ flip: isFlipped(prompt.id) }"
            @click="toggleFlip(prompt.id)"
          >
          <!-- ===== 正面：轻量识别信息 ===== -->
          <div class="face front">
            <div class="f-top">
              <span class="card-category-badge" :style="{ borderColor: getCategoryColor(prompt.category), color: getCategoryColor(prompt.category) }">
                {{ getCategoryLabel(prompt.category) }}
              </span>
              <h3 class="card-title" :title="prompt.name">{{ prompt.name }}</h3>
              <span v-if="prompt.card_type === 'encrypted'" class="lock-ico" title="私人提示词">🔒</span>
            </div>
            <div class="card-author-row">
              <div class="author-avatar" :style="{ background: getAvatarColor(prompt.creator_name || prompt.name) }">
                {{ (prompt.creator_name || '匿').charAt(0) }}
              </div>
              <span class="author-name" :title="prompt.creator_name">{{ prompt.creator_name || '匿名用户' }}</span>
              <span class="author-stats">
                <span class="stat-item">
                  <svg class="stat-icon fire-icon" viewBox="0 0 24 24" fill="currentColor"><path d="M12 23c-4.97 0-9-3.58-9-8 0-2.52 1.18-4.78 3-6.26V6c0-.55.45-1 1-1s1 .45 1 1v1.24c1.06-.5 2.25-.79 3.5-.79h.5c.28 0 .5.22.5.5v2c0 .28-.22.5-.5.5H12c-2.76 0-5 2.24-5 5s2.24 5 5 5 5-2.24 5-5c0-1.13-.38-2.17-1.01-3.02C15.62 11.55 14 10.45 14 9c0-.55.45-1 1-1s1 .45 1 1c0 .83.67 1.5 1.5 1.5S19 9.83 19 9c0-2.76-2.24-5-5-5-.34 0-.67.03-1 .09V3c0-.55.45-1 1-1s1 .45 1 1v1.74c1.82 1.48 3 3.74 3 6.26 0 4.42-4.03 8-9 8z"/></svg>
                  {{ prompt.use_count ?? prompt.order_num ?? 0 }}
                </span>
              </span>
              <span class="card-date">{{ formatDate(prompt.created_at) }}</span>
            </div>
            <div class="f-hint">点击翻转 · 背面可就地编辑</div>
          </div>

          <!-- ===== 背面：就地编辑 ===== -->
          <div class="face back">
            <div class="b-head">
              <span class="b-label">就地编辑</span>
              <button class="b-flipback" @click.stop="toggleFlip(prompt.id)">翻回 ⟳</button>
            </div>
            <div class="b-body">
              <!-- 名称内联编辑 -->
              <template v-if="inlineEdit.id === prompt.id && inlineEdit.field === 'name'">
                <input
                  class="ie-input"
                  :id="`ie-name-${prompt.id}`"
                  v-model="inlineEdit.draft"
                  placeholder="提示词名称"
                  @blur="commitInlineEdit"
                  @keydown.enter.prevent="commitInlineEdit"
                  @keydown.esc.prevent="cancelInlineEdit"
                />
              </template>
              <div v-else class="b-title" title="双击编辑名称" @dblclick="startInlineEdit(prompt, 'name')">{{ prompt.name }}</div>

              <!-- 标签：点击删除 / ＋添加 -->
              <div class="b-tags">
                <span
                  v-for="tag in (prompt.subcategories || [])"
                  :key="tag"
                  class="b-tag"
                  title="点击删除"
                  @click.stop="removeTagInline(prompt, tag)"
                >{{ tag }}<span class="x">✕</span></span>
                <template v-if="tagAddId === prompt.id">
                  <input
                    class="b-tag-input"
                    v-model="tagAddDraft"
                    placeholder="回车确认"
                    @blur="commitTagAdd(prompt)"
                    @keydown.enter.prevent="commitTagAdd(prompt)"
                    @keydown.esc.prevent="tagAddId = null"
                  />
                </template>
                <button v-else class="b-add" @click.stop="startTagAdd(prompt)">＋ 标签</button>
              </div>

              <!-- 简介内联编辑 -->
              <template v-if="inlineEdit.id === prompt.id && inlineEdit.field === 'description'">
                <textarea
                  class="ie-input ie-textarea"
                  :id="`ie-description-${prompt.id}`"
                  v-model="inlineEdit.draft"
                  rows="4"
                  placeholder="提示词简介"
                  @blur="commitInlineEdit"
                  @keydown.esc.prevent="cancelInlineEdit"
                ></textarea>
              </template>
              <p v-else class="b-desc" title="双击编辑简介" @dblclick="startInlineEdit(prompt, 'description')">
                {{ getPromptDescriptionPreview(prompt) || '暂无简介，双击添加…' }}
              </p>

              <div class="b-actions">
                <button class="ba primary" @click.stop="handlePreview(prompt)">预览全文</button>
                <button class="ba" @click.stop="openSubcategoryEditor(prompt)">管理标签</button>
              </div>
            </div>
          </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 预览弹窗：IDE 工作台（Demo 05） -->
    <el-dialog
      v-model="previewDialogVisible"
      width="78vw"
      top="5vh"
      :show-close="false"
      destroy-on-close
      class="preview-dialog ide-preview"
      append-to-body
    >
      <div v-if="selectedPreviewPrompt" class="ide-preview-body">
        <!-- 顶部标签栏 -->
        <div class="ide-tabs">
          <div class="ide-crumb">
            <span>提示词库</span><span class="sep">/</span>
            <span>{{ selectedPreviewPrompt.category }}</span><span class="sep">/</span>
            <span class="cur">{{ selectedPreviewPrompt.name }}</span>
          </div>
          <div class="ide-filetab"><span class="dot"></span><span class="name">{{ selectedPreviewPrompt.name }}</span><span class="lang">.prompt</span></div>
          <div class="ide-spacer"></div>
          <button class="ide-icon copy" title="复制全文" @click="copyPreviewContent"><el-icon><CopyDocument /></el-icon></button>
          <template v-if="selectedPreviewPrompt.card_type === 'encrypted' && selectedPreviewPrompt.password">
            <el-popover
              :visible="editPasswordPopoverVisible"
              placement="bottom"
              :width="280"
              trigger="click"
            >
              <template #reference>
                <button class="ide-icon" title="编辑" @click="editPasswordPopoverVisible = true"><el-icon><Edit /></el-icon></button>
              </template>
              <div class="edit-password-popover">
                <p class="popover-hint">请输入密码编辑</p>
                <el-input
                  v-model="editPasswordInput"
                  type="password"
                  placeholder="请输入密码"
                  show-password
                  @keyup.enter="verifyEditPassword"
                />
                <div class="popover-actions">
                  <el-button size="small" @click="editPasswordPopoverVisible = false">取消</el-button>
                  <el-button type="primary" size="small" @click="verifyEditPassword">确认</el-button>
                </div>
              </div>
            </el-popover>
          </template>
          <button v-else class="ide-icon" title="编辑" @click="editPreviewPrompt"><el-icon><Edit /></el-icon></button>
          <button class="ide-icon" title="关闭" @click="previewDialogVisible = false"><el-icon><Close /></el-icon></button>
        </div>

        <!-- 状态栏 -->
        <div class="ide-status">
          <span><b>作者</b> {{ selectedPreviewPrompt.creator_name || '匿名用户' }}</span>
          <span v-if="selectedPreviewPrompt.version"><b>版本</b> {{ selectedPreviewPrompt.version }}</span>
          <span><b>使用</b> {{ selectedPreviewPrompt.use_count ?? selectedPreviewPrompt.order_num ?? 0 }} 次</span>
          <span><b>更新</b> {{ formatDate(selectedPreviewPrompt.created_at) }}</span>
          <span
            v-for="(tag, idx) in (selectedPreviewPrompt.subcategories || [])"
            :key="idx"
            class="tag"
            :style="{ background: getTagColor(tag).bg, color: getTagColor(tag).text, borderColor: getTagColor(tag).border }"
          >{{ tag }}</span>
          <span class="ok"><i></i>已就绪</span>
        </div>

        <!-- 主体：左大纲/字段 + 右内容 -->
        <div class="ide-main">
          <aside class="ide-outline">
            <div class="ol-tabs">
              <button class="ol-tab" :class="{ on: previewOutlineTab === 'outline' }" @click="previewOutlineTab = 'outline'">大纲</button>
              <button class="ol-tab" :class="{ on: previewOutlineTab === 'fields' }" @click="previewOutlineTab = 'fields'">字段 <span class="fc">{{ previewFieldList.length }}</span></button>
            </div>
            <div v-show="previewOutlineTab === 'outline'" class="ol-panel ol-outline">
              <div
                v-for="item in previewOutlineItems"
                :key="item.key"
                class="ol-item"
                :class="{ on: previewOutlineActive === item.key }"
                @click="scrollPreviewTo(item.key, item.line)"
              >
                <svg class="ic" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 6h16M4 12h16M4 18h16"/></svg>{{ item.label }}<span class="ln">L{{ item.line + 1 }}</span>
              </div>
              <div v-if="previewOutlineItems.length === 0" class="ol-empty">暂无大纲</div>
            </div>
            <div v-show="previewOutlineTab === 'fields'" class="ol-panel ol-fields">
              <template v-if="previewFieldList.length">
                <div
                  v-for="(f, i) in previewFieldList"
                  :key="i"
                  class="ol-field"
                  :class="{ on: previewActiveField === f.name }"
                  @click="scrollToField(f.name)"
                >
                  <div class="f-head"><span class="f-name" :title="f.name">{{ f.label || f.name }}</span><span v-if="f.required" class="req">必选</span></div>
                  <div class="f-desc">{{ f.description || '暂无说明' }}</div>
                  <span class="f-ph">{{ '${' + f.name + '}' }}</span>
                </div>
              </template>
              <div v-else class="ol-empty">暂无字段说明</div>
            </div>
          </aside>
          <div class="ide-content">
            <div class="ide-file-head">
              <span class="fname">{{ selectedPreviewPrompt.name }}.prompt</span>
              <span class="lang-chip">markdown</span>
              <span class="size">{{ previewLineCount }} 行</span>
            </div>
            <el-scrollbar class="ide-scroll">
              <div class="ide-code-panel">
                <div
                  v-for="(l, i) in previewCodeLines"
                  :key="i"
                  :id="'preview-line-' + i"
                  class="code-line"
                  :class="{ on: previewActiveLine === i }"
                >
                  <span class="code-num">{{ i + 1 }}</span>
                  <div class="code-txt" v-html="highlightCodeLine(l.text, l.isComment)"></div>
                </div>
                <div v-if="selectedPreviewPrompt.card_type === 'encrypted' && selectedPreviewPrompt.password" class="ide-encrypted">
                  <el-icon :size="30" color="#e6a23c"><Lock /></el-icon>
                  <p class="encrypted-title">私人提示词</p>
                  <span class="encrypted-sub">请输入密码后查看或编辑内容</span>
                </div>
              </div>
            </el-scrollbar>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 小分类编辑弹窗 -->
    <el-dialog
      v-model="subcategoryDialogVisible"
      title=""
      width="380px"
      destroy-on-close
      class="subcategory-dialog"
      append-to-body
    >
      <div class="subcategory-editor">
        <div class="subcategory-header">
          <span class="subcategory-title">编辑标签</span>
          <span class="subcategory-count">{{ editingSubcategories.length }} 个标签</span>
        </div>
        <div class="subcategory-list">
          <transition-group name="tag-list">
            <div
              v-for="(_, index) in editingSubcategories"
              :key="index"
              class="subcategory-item"
            >
              <el-input
                v-model="editingSubcategories[index]"
                size="small"
                placeholder="输入标签名称"
                class="tag-input"
              />
              <el-button
                type="danger"
                link
                size="small"
                @click="removeSubcategory(index)"
                class="delete-btn"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </transition-group>
          <div v-if="editingSubcategories.length === 0" class="empty-tags">
            暂无标签，点击下方添加
          </div>
        </div>
        <el-button
          type="primary"
          size="small"
          @click="addSubcategory"
          class="add-btn"
          plain
        >
          <el-icon><Plus /></el-icon>
          添加标签
        </el-button>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="subcategoryDialogVisible = false" class="cancel-btn">取消</el-button>
          <el-button type="primary" @click="saveSubcategories" class="save-btn">保存</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 编辑提示词弹窗 -->
    <el-dialog
      v-model="editDialogVisible"
      title="编辑提示词"
      :width="editDialogWidthMap[editDialogSize]"
      top="6vh"
      append-to-body
      :close-on-click-modal="false"
      class="edit-prompt-dialog"
    >
      <template #header>
        <div class="edit-dialog-header">
          <span>编辑提示词</span>
          <div class="size-controls">
            <el-button-group>
              <el-button
                :type="editDialogSize === 'small' ? 'primary' : ''"
                size="small"
                @click="editDialogSize = 'small'"
              >
                小
              </el-button>
              <el-button
                :type="editDialogSize === 'medium' ? 'primary' : ''"
                size="small"
                @click="editDialogSize = 'medium'"
              >
                中
              </el-button>
              <el-button
                :type="editDialogSize === 'large' ? 'primary' : ''"
                size="small"
                @click="editDialogSize = 'large'"
              >
                大
              </el-button>
            </el-button-group>
          </div>
        </div>
      </template>
      <div class="edit-dialog-content" :style="{ height: editDialogHeightMap[editDialogSize] }">
        <div class="edit-left-panel">
          <el-form :model="editFormData" label-width="80px">
            <el-form-item label="名称" required>
              <el-input v-model="editFormData.name" placeholder="请输入名称" />
            </el-form-item>
            <el-form-item label="分类名称" required>
              <el-select
                v-model="editFormData.category"
                placeholder="请选择卡包"
                class="field-input full-width-input"
              >
                <el-option
                  v-for="category in editCategoryOptions"
                  :key="category"
                  :label="category"
                  :value="category"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="作者名">
              <el-input v-model="editFormData.creator_name" placeholder="请输入作者名，非必填" />
            </el-form-item>
            <el-form-item label="版本号">
              <el-input v-model="editFormData.version" placeholder="如 v1.0、v2.1.0，非必填" />
            </el-form-item>
            <el-form-item label="简介">
              <SplitRichTextEditor
                v-model="editFormData.description"
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
                          <span>字段名称尽量直接，比如"数量""平台""风格要求"。</span>
                        </div>
                      </div>
                    </div>
                  </el-popover>
                </div>
              </template>
              <el-input
                v-model="editFormData.content"
                type="textarea"
                :rows="15"
                placeholder="请输入提示词内容（作为AI的system层指令）"
              />
            </el-form-item>
            <div class="panel-header edit-panel-header">
              <label class="form-label">字段配置</label>
              <el-button type="primary" size="small" @click="addEditField">
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

              <div v-if="editFieldsConfig.length === 0" class="empty-fields">
                <p class="empty-hint">暂无字段配置，点击右上角"添加字段"开始配置</p>
              </div>

              <div v-else class="fields-list">
                <div v-for="(field, index) in editFieldsConfig" :key="index" class="field-item">
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
                        @change="updateEditFieldName(index, field.name)"
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
                        inline-prompt
                        style="--el-switch-on-color: #f56c6c; --el-switch-off-color: #909399;"
                      />
                      <el-button
                        type="primary"
                        link
                        size="small"
                        @click="moveEditField(index, index - 1)"
                        :disabled="index === 0"
                        title="上移"
                      >
                        <el-icon><ArrowUp /></el-icon>
                      </el-button>
                      <el-button
                        type="primary"
                        link
                        size="small"
                        @click="moveEditField(index, index + 1)"
                        :disabled="index === editFieldsConfig.length - 1"
                        title="下移"
                      >
                        <el-icon><ArrowDown /></el-icon>
                      </el-button>
                      <el-button type="danger" link size="small" @click="removeEditField(index)" title="删除">
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
                        @change="updateEditFieldOptions(index)"
                        class="field-input"
                      />
                      <label>选项展示名称（每行一个，可选）：</label>
                      <el-input
                        v-model="field.optionLabelsText"
                        type="textarea"
                        :rows="3"
                        placeholder="请输入展示名称，每行一个；留空时默认使用选项值"
                        @change="updateEditFieldOptionLabels(index)"
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
          </el-form>
        </div>
      </div>
      <template #footer>
        <div class="edit-dialog-footer">
          <el-button @click="editDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleEditSubmit">保存</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 保存条：即时反馈 + 撤销 -->
    <transition name="sb">
      <div v-if="saveBarVisible" class="save-bar">
        <span class="sb-msg">{{ saveBarMsg }}</span>
        <button class="sb-btn ghost" @click="undoLastChange">撤销</button>
        <button class="sb-btn" @click="saveBarVisible = false">知道了</button>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { CopyDocument, Edit, Plus, Delete, Close, ArrowUp, ArrowDown, Check, InfoFilled, Lock } from '@element-plus/icons-vue'
import { promptAPI } from '@/api'
import type { Prompt } from '@/types'
import SplitRichTextEditor from '@/components/SplitRichTextEditor.vue'

// 提示词列表
const prompts = ref<Prompt[]>([])

// 选中的分类
const selectedCategory = ref('all')

// 搜索关键词
const searchKeyword = ref('')

// 排序选项
const selectedSort = ref('latest')
const sortTabs = [
  { label: '最新', value: 'latest' },
  { label: '最热', value: 'hottest' },
  { label: '名称A-Z', value: 'name_asc' },
]

// 预览弹窗相关状态
const previewDialogVisible = ref(false)
const selectedPreviewPrompt = ref<Prompt | null>(null)
const editPasswordPopoverVisible = ref(false)
const editPasswordInput = ref('')

// ── IDE 预览：大纲 / 字段页签 ──
const previewOutlineTab = ref<'outline' | 'fields'>('outline')
const previewOutlineActive = ref('')
const previewActiveLine = ref(-1)
const previewActiveField = ref('')
// 将简介 + 内容拆成带行号的代码行
const previewCodeLines = computed(() => {
  const p = selectedPreviewPrompt.value
  if (!p) return []
  const lines: { text: string; isComment: boolean }[] = []
  lines.push({ text: '# 简介', isComment: true })
  const desc = stripHtml(p.description || '').trim()
  if (desc) desc.split('\n').forEach(l => lines.push({ text: l, isComment: false }))
  else lines.push({ text: '（暂无简介）', isComment: false })
  const isEnc = p.card_type === 'encrypted' && p.password
  if (!isEnc) {
    lines.push({ text: '', isComment: false })
    lines.push({ text: '# 提示词内容', isComment: true })
    const content = (p.content || '').trim()
    if (content) content.split('\n').forEach(l => lines.push({ text: l, isComment: false }))
    else lines.push({ text: '（暂无内容）', isComment: false })
  }
  return lines
})
// 大纲：从 # / ## 标题自动生成
const previewOutlineItems = computed(() => {
  const items: { key: string; label: string; line: number }[] = []
  previewCodeLines.value.forEach((l, i) => {
    const m = l.text.match(/^#{1,4}\s+(.+)/)
    if (m) items.push({ key: 'h-' + i, label: m[1].trim(), line: i })
  })
  return items
})
const previewFieldList = computed(() => {
  const p = selectedPreviewPrompt.value
  if (!p) return []
  if (p.fields && p.fields.length > 0) {
    return p.fields.map(f => ({
      name: f.name,
      label: f.label || f.name,
      description: f.description || '',
      required: f.required !== undefined ? f.required : true
    }))
  }
  return extractFieldNames(p.content).map(name => ({ name, label: name, description: '', required: true }))
})
const previewLineCount = computed(() => previewCodeLines.value.length)
function scrollPreviewTo(key: string, line: number) {
  previewOutlineActive.value = key
  previewActiveLine.value = line
  nextTick(() => {
    const el = document.getElementById('preview-line-' + line)
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  })
}
function scrollToField(name: string) {
  previewActiveField.value = name
  const line = previewCodeLines.value.findIndex(l => l.text.includes('${' + name + '}'))
  if (line >= 0) {
    previewActiveLine.value = line
    nextTick(() => {
      const el = document.getElementById('preview-line-' + line)
      if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    })
  } else {
    previewActiveLine.value = -1
  }
}
// 代码行高亮：标题→注释灰、${字段}→紫色、行首序号→绿色（先转义防 XSS）
function highlightCodeLine(text: string, isComment: boolean): string {
  const esc = String(text)
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    .replace(/\*\*/g, '')
  if (isComment) return `<span class="c">${esc || ' '}</span>`
  const numbered = esc.replace(/^(\d+[.)]\s*)/, '<span class="s">$1</span>')
  return numbered.replace(/\$\{([^}]+)\}/g, '<span class="h">${$1}</span>')
}

// ── Demo16 风格：翻转卡片 ──
const flippedIds = ref<Set<number>>(new Set())
function toggleFlip(id: number) {
  const s = new Set(flippedIds.value)
  s.has(id) ? s.delete(id) : s.add(id)
  flippedIds.value = s
}
const isFlipped = (id: number) => flippedIds.value.has(id)

// ── 内联编辑（标题 / 简介）──
const inlineEdit = ref<{ id: number | null; field: '' | 'name' | 'description'; draft: string }>({ id: null, field: '', draft: '' })
async function startInlineEdit(prompt: Prompt, field: 'name' | 'description') {
  inlineEdit.value = { id: prompt.id, field, draft: String((prompt as any)[field] ?? '') }
  await nextTick()
  const el = document.getElementById(`ie-${field}-${prompt.id}`) as HTMLInputElement | null
  if (el) { el.focus(); el.select() }
}
function cancelInlineEdit() { inlineEdit.value = { id: null, field: '', draft: '' } }
async function commitInlineEdit() {
  const { id, field, draft } = inlineEdit.value
  cancelInlineEdit()
  if (id === null || !field) return
  const prompt = prompts.value.find(p => p.id === id)
  if (!prompt) return
  const val = draft.trim()
  if (!val || val === String((prompt as any)[field] ?? '')) return
  snapshotUndo(prompt, field, (prompt as any)[field])
  ;(prompt as any)[field] = val
  await persistPrompt(prompt)
}

// ── 标签内联：删除 / 添加 ──
const tagAddId = ref<number | null>(null)
const tagAddDraft = ref('')
async function startTagAdd(prompt: Prompt) {
  tagAddId.value = prompt.id
  tagAddDraft.value = ''
  await nextTick()
  const el = document.querySelector('.b-tag-input') as HTMLInputElement | null
  if (el) el.focus()
}
function commitTagAdd(prompt: Prompt) {
  const v = tagAddDraft.value.trim()
  tagAddId.value = null
  if (!v) return
  const old = [...(prompt.subcategories || [])]
  if (old.includes(v)) { ElMessage.warning('标签已存在'); return }
  snapshotUndo(prompt, 'subcategories', old)
  prompt.subcategories = [...old, v]
  persistPrompt(prompt)
}
function removeTagInline(prompt: Prompt, tag: string) {
  snapshotUndo(prompt, 'subcategories', [...(prompt.subcategories || [])])
  prompt.subcategories = (prompt.subcategories || []).filter(t => t !== tag)
  persistPrompt(prompt)
}

// ── 保存条 + 撤销 ──
const saveBarVisible = ref(false)
const saveBarMsg = ref('')
const undoSnapshots = ref<Array<{ prompt: Prompt; field: string; oldVal: any }>>([])
let undoHideTimer: ReturnType<typeof setTimeout> | null = null
function snapshotUndo(prompt: Prompt, field: string, oldVal: any) {
  undoSnapshots.value.unshift({ prompt, field, oldVal })
  if (undoSnapshots.value.length > 5) undoSnapshots.value.pop()
  saveBarMsg.value = '已保存修改'
  saveBarVisible.value = true
  if (undoHideTimer) clearTimeout(undoHideTimer)
  undoHideTimer = setTimeout(() => { saveBarVisible.value = false }, 5000)
}
function undoLastChange() {
  const s = undoSnapshots.value.shift()
  if (!s) return
  ;(s.prompt as any)[s.field] = s.oldVal
  persistPrompt(s.prompt)
  ElMessage.success('已撤销')
  if (!undoSnapshots.value.length) saveBarVisible.value = false
}

async function persistPrompt(prompt: Prompt) {
  try {
    await promptAPI.update(prompt.id, { ...prompt })
  } catch {
    ElMessage.error('保存失败，请重试')
  }
}

// 密码哈希函数
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
const decryptedEditContent = ref('')

const verifyEditPassword = async () => {
  if (!selectedPreviewPrompt.value) return
  const isValid = await verifyPassword(editPasswordInput.value, selectedPreviewPrompt.value.password || '')
  if (isValid) {
    currentSessionPassword.value = editPasswordInput.value
    const decrypted = await decryptContent(selectedPreviewPrompt.value.content, editPasswordInput.value)
    decryptedEditContent.value = decrypted !== null ? decrypted : selectedPreviewPrompt.value.content
    editPasswordPopoverVisible.value = false
    editPasswordInput.value = ''
    editPreviewPrompt()
  } else {
    ElMessage.error('密码错误')
  }
}

// 小分类编辑弹窗相关状态
const subcategoryDialogVisible = ref(false)
const editingPrompt = ref<Prompt | null>(null)
const editingSubcategories = ref<string[]>([])

// 编辑弹窗相关状态
const editDialogVisible = ref(false)
const editDialogSize = ref<'small' | 'medium' | 'large'>('medium')
const editDialogWidthMap = {
  small: '900px',
  medium: '1200px',
  large: '1500px'
}
const editDialogHeightMap = {
  small: '500px',
  medium: '600px',
  large: '700px'
}
const editFormData = ref({
  id: 0,
  name: '',
  description: '',
  content: '',
  category: '默认',
  order_num: 0,
  creator_name: '',
  version: ''
})
const editFieldsConfig = ref<Array<{
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
const editFormSubcategories = ref<string[]>([])
const DEFAULT_CATEGORIES = ['默认', '写作要求', '写作风格']

const editCategoryOptions = computed(() => {
  const names = new Set<string>(DEFAULT_CATEGORIES)
  if (editFormData.value.category) {
    names.add(editFormData.value.category)
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

const getPromptDescriptionPreview = (prompt: Prompt) => {
  const description = stripHtml(prompt.description || '').trim()
  if (!description) return ''
  return `${description.slice(0, 100)}${description.length > 100 ? '...' : ''}`
}

const categoryColorMap: Record<string, string> = {}
const categoryLabelMap: Record<string, string> = {}
const colorPalette = [
  'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
  'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
  'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
  'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
  'linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)',
  'linear-gradient(135deg, #fccb90 0%, #d57eeb 100%)',
  'linear-gradient(135deg, #e0c3fc 0%, #8ec5fc 100%)',
  'linear-gradient(135deg, #f5576c 0%, #ff8a5c 100%)',
  'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
]
let colorIndex = 0

const getCategoryColor = (category: string): string => {
  if (!categoryColorMap[category]) {
    categoryColorMap[category] = colorPalette[colorIndex % colorPalette.length]
    colorIndex++
  }
  return categoryColorMap[category]
}

const getCategoryLabel = (category: string): string => {
  if (!categoryLabelMap[category]) {
    categoryLabelMap[category] = (category || '').slice(0, 2)
  }
  return categoryLabelMap[category]
}

const avatarColors = [
  '#667eea', '#f093fb', '#4facfe', '#43e97b', '#fa709a',
  '#a18cd1', '#fccb90', '#f5576c', '#764ba2', '#00f2fe'
]
const avatarColorCache: Record<string, string> = {}

const getAvatarColor = (name: string): string => {
  if (!avatarColorCache[name]) {
    let hash = 0
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash)
    }
    avatarColorCache[name] = avatarColors[Math.abs(hash) % avatarColors.length]
  }
  return avatarColorCache[name]
}

const tagColorPalette = [
  { bg: '#fef3c7', text: '#d97706', border: '#fcd34d' },
  { bg: '#dbeafe', text: '#2563eb', border: '#93c5fd' },
  { bg: '#dcfce7', text: '#16a34a', border: '#93c5fd' },
  { bg: '#fce7f3', text: '#db2777', border: '#f9a8d4' },
  { bg: '#f3e8ff', text: '#9333ea', border: '#c4b5fd' },
  { bg: '#fff7ed', text: '#ea580c', border: '#fdba74' },
  { bg: '#ecfeff', text: '#0891b2', border: '#67e8f9' },
  { bg: '#fef2f2', text: '#dc2626', border: '#fca5a5' },
  { bg: '#f5f3ff', text: '#6d28d9', border: '#c4b5fd' },
  { bg: '#ecfdf5', text: '#059669', border: '#6ee7b7' },
]
const tagColorCache: Record<string, typeof tagColorPalette[0]> = {}

const getTagColor = (tagText: string): typeof tagColorPalette[0] => {
  if (!tagColorCache[tagText]) {
    let hash = 0
    for (let i = 0; i < tagText.length; i++) {
      hash = tagText.charCodeAt(i) + ((hash << 5) - hash)
    }
    tagColorCache[tagText] = tagColorPalette[Math.abs(hash) % tagColorPalette.length]
  }
  return tagColorCache[tagText]
}

const stripHtml = (value: string) => {
  if (!value) return ''
  if (typeof window === 'undefined') {
    return value.replace(/<[^>]+>/g, ' ')
  }
  const temp = document.createElement('div')
  temp.innerHTML = value
  return temp.textContent || temp.innerText || ''
}

// 打开小分类编辑器
const openSubcategoryEditor = (prompt: Prompt) => {
  editingPrompt.value = prompt
  editingSubcategories.value = [...(prompt.subcategories || [])]
  subcategoryDialogVisible.value = true
}

// 添加小分类
const addSubcategory = () => {
  editingSubcategories.value.push('')
}

// 删除小分类
const removeSubcategory = (index: number) => {
  editingSubcategories.value.splice(index, 1)
}

// 保存小分类
const saveSubcategories = async () => {
  if (!editingPrompt.value) return

  // 过滤掉空的小分类
  const filteredSubcategories = editingSubcategories.value.filter(subcat => subcat.trim())

  try {
    await promptAPI.update(editingPrompt.value.id, {
      ...editingPrompt.value,
      subcategories: filteredSubcategories
    })

    // 更新本地数据
    const promptIndex = prompts.value.findIndex(p => p.id === editingPrompt.value!.id)
    if (promptIndex !== -1) {
      const existingPrompt = prompts.value[promptIndex]
      if (existingPrompt) {
        prompts.value[promptIndex] = {
          ...existingPrompt,
          subcategories: filteredSubcategories
        }
      }
    }

    ElMessage.success('标签保存成功')
    subcategoryDialogVisible.value = false
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

// 从localStorage加载卡包列表
const loadCategoriesFromStorage = (): string[] => {
  const stored = localStorage.getItem('prompt_categories')
  return stored ? JSON.parse(stored) : []
}

// 卡包列表（从localStorage加载）
const customCategories = ref<string[]>(loadCategoriesFromStorage())

// 按分类分组的提示词列表
const categoryList = computed(() => {
  const groups: Record<string, Prompt[]> = {}

  // 先确保所有自定义卡包都存在
  customCategories.value.forEach(cat => {
    if (!groups[cat]) {
      groups[cat] = []
    }
  })

  // 添加提示词到对应分组
  prompts.value.forEach(prompt => {
    const category = prompt.category || '默认'
    if (!groups[category]) {
      groups[category] = []
    }
    groups[category].push(prompt)
  })

  // 确保默认分类存在
  if (!groups['默认']) {
    groups['默认'] = []
  }

  return Object.entries(groups).map(([name, prompts]) => ({
    name,
    prompts
  }))
})

// 过滤后的提示词列表
const filteredPrompts = computed(() => {
  let result = prompts.value

  // 按分类过滤
  if (selectedCategory.value !== 'all') {
    result = result.filter(prompt => prompt.category === selectedCategory.value)
  }

  // 按搜索关键词过滤
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(prompt =>
      prompt.name.toLowerCase().includes(keyword) ||
      prompt.content.toLowerCase().includes(keyword) ||
      (prompt.description || '').toLowerCase().includes(keyword)
    )
  }

  // 排序
  switch (selectedSort.value) {
    case 'latest':
      result = [...result].sort((a, b) => {
        const timeA = a.created_at ? new Date(a.created_at).getTime() : 0
        const timeB = b.created_at ? new Date(b.created_at).getTime() : 0
        return timeB - timeA
      })
      break
    case 'hottest':
      result = [...result].sort((a, b) => (b.order_num || 0) - (a.order_num || 0))
      break
    case 'name_asc':
      result = [...result].sort((a, b) => a.name.localeCompare(b.name, 'zh-CN'))
      break
  }

  return result
})

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

// 加载提示词
const fetchPrompts = async () => {
  const res = await promptAPI.getAll()
  if (res.success && res.data) {
    prompts.value = res.data
  }
}

// 预览提示词
const handlePreview = (prompt: Prompt) => {
  selectedPreviewPrompt.value = prompt
  previewDialogVisible.value = true
}

const editPreviewPrompt = () => {
  if (!selectedPreviewPrompt.value) return

  const prompt = selectedPreviewPrompt.value
  const displayContent = decryptedEditContent.value || prompt.content
  decryptedEditContent.value = ''
  editFormData.value = {
    id: prompt.id,
    name: prompt.name,
    description: prompt.description || '',
    content: displayContent,
    category: prompt.category,
    order_num: prompt.order_num,
    creator_name: prompt.creator_name || '',
    version: prompt.version || ''
  }

  editFieldsConfig.value = []

  if (prompt.fields && prompt.fields.length > 0) {
    editFieldsConfig.value = prompt.fields.map(field => ({
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
    const fieldNames = extractFieldNames(prompt.content)
    editFieldsConfig.value = fieldNames.map(name => ({
      name,
      label: name,
        type: 'text' as const,
        options: [],
        optionLabels: [],
        optionsText: '',
        optionLabelsText: '',
        description: '',
        required: true
      }))
  }

  editFormSubcategories.value = [...(prompt.subcategories || [])]
  previewDialogVisible.value = false
  editDialogVisible.value = true
}

const extractFieldNames = (content: string): string[] => {
  const regex = /\$\{([^}]+)\}/g
  const names: string[] = []
  let match
  while ((match = regex.exec(content)) !== null) {
    if (!names.includes(match[1])) {
      names.push(match[1])
    }
  }
  return names
}

const addEditField = () => {
  editFieldsConfig.value.push({
    name: '',
    label: '',
    type: 'text',
    options: [],
    optionLabels: [],
    optionsText: '',
    optionLabelsText: '',
    description: '',
    required: true
  })
}

const removeEditField = (index: number) => {
  editFieldsConfig.value.splice(index, 1)
}

const moveEditField = (fromIndex: number, toIndex: number) => {
  if (toIndex < 0 || toIndex >= editFieldsConfig.value.length) return
  const field = editFieldsConfig.value[fromIndex]
  editFieldsConfig.value.splice(fromIndex, 1)
  editFieldsConfig.value.splice(toIndex, 0, field)
}

const updateEditFieldOptions = (index: number) => {
  const field = editFieldsConfig.value[index]
  if (field) {
    field.options = field.optionsText.split('\n').map(opt => opt.trim()).filter(opt => opt)
    field.optionLabels = field.optionLabels.slice(0, field.options.length)
    field.optionLabelsText = field.optionLabels.join('\n')
  }
}

const updateEditFieldOptionLabels = (index: number) => {
  const field = editFieldsConfig.value[index]
  if (field) {
    const labels = field.optionLabelsText
      .split('\n')
      .map(opt => opt.trim())
      .slice(0, field.options.length)

    while (labels.length > 0 && !labels[labels.length - 1]) {
      labels.pop()
    }

    field.optionLabels = labels
    field.optionLabelsText = labels.join('\n')
  }
}

const updateEditFieldName = (index: number, newName: string) => {
  const oldName = editFieldsConfig.value[index]?.name
  if (oldName && newName && oldName !== newName) {
    editFormData.value.content = editFormData.value.content.replace(
      new RegExp(`\\$\\{${oldName}\\}`, 'g'),
      `\${${newName}}`
    )
  }
}

const copyFieldName = (name: string) => {
  navigator.clipboard.writeText(`\${${name}}`)
  ElMessage.success('已复制字段占位符')
}

const validateEditFieldsConfig = (): string | null => {
  for (const field of editFieldsConfig.value) {
    if (!field.name.trim()) {
      return '字段名称不能为空'
    }
    if (!field.label.trim()) {
      return `字段"${field.name}"的显示名称不能为空`
    }
    if (field.type === 'select' && field.options.length === 0) {
      return `字段"${field.name}"是下拉选择类型，请至少添加一个选项`
    }
  }
  return null
}

const handleEditSubmit = async () => {
  if (!editFormData.value.name || !editFormData.value.category || !editFormData.value.content) {
    ElMessage.warning('请填写完整信息')
    return
  }

  const validationError = validateEditFieldsConfig()
  if (validationError) {
    ElMessage.error(validationError)
    return
  }

  try {
    let saveContent = editFormData.value.content
    if (selectedPreviewPrompt.value?.card_type === 'encrypted' && currentSessionPassword.value) {
      saveContent = await encryptContent(saveContent, currentSessionPassword.value)
    }

    const dataToSave = {
      ...editFormData.value,
      content: saveContent,
      fields: editFieldsConfig.value.map(field => ({
        name: field.name,
        label: field.label,
        type: field.type,
        options: field.options,
        optionLabels: field.optionLabels,
        description: field.description,
        required: field.required
      })),
      subcategories: editFormSubcategories.value
    }

    const res = await promptAPI.update(editFormData.value.id, dataToSave)
    if (res.success) {
      currentSessionPassword.value = ''
      const promptIndex = prompts.value.findIndex(p => p.id === editFormData.value.id)
      if (promptIndex !== -1) {
        prompts.value[promptIndex] = {
          ...prompts.value[promptIndex],
          ...dataToSave
        } as Prompt
      }
      ElMessage.success('保存成功')
      editDialogVisible.value = false
    }
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

// 复制预览内容到剪贴板
const copyPreviewContent = async () => {
  try {
    if (!selectedPreviewPrompt.value) return
    const description = selectedPreviewPrompt.value.description?.trim()
    const textToCopy = [
      selectedPreviewPrompt.value.name,
      formatDate(selectedPreviewPrompt.value.created_at),
      description ? `简介\n${stripHtml(description)}` : '',
      `提示词内容\n${selectedPreviewPrompt.value.content}`
    ].filter(Boolean).join('\n\n')
    await navigator.clipboard.writeText(textToCopy)
    ElMessage.success('内容已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

onMounted(async () => {
  await fetchPrompts()
})
</script>

<style scoped>
/* ══════════ 整体容器：Demo16 薰衣草渐变 ══════════ */
.prompt-preview-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 24px;
  background: linear-gradient(180deg, #f4eefc 0%, #e7ddf8 40%, #f5effd 100%);
  overflow: hidden;
}

/* ══════════ 顶栏 ══════════ */
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  background: rgba(255, 255, 255, 0.82);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid #ece4f8;
  border-radius: 14px;
  padding: 20px 24px;
  box-shadow: 0 2px 10px rgba(124, 58, 237, 0.06);
  position: relative;
}
.header h2 {
  font-size: 22px;
  font-weight: 800;
  color: #1a1a2e;
  margin: 0;
  position: relative;
  z-index: 1;
}
.header h2::before {
  content: '';
  display: inline-block;
  width: 4px;
  height: 20px;
  background: linear-gradient(180deg, #7c3aed, #6d28d9);
  border-radius: 2px;
  margin-right: 10px;
  vertical-align: middle;
}
.search-box {
  width: 300px;
  position: relative;
  z-index: 1;
}
.search-input {
  width: 100%;
}
.search-input :deep(.el-input__wrapper) {
  border-radius: 20px;
  background: #f7f4fd;
  border: 1px solid #ece4f8;
  box-shadow: none;
  transition: all 0.25s;
}
.search-input :deep(.el-input__wrapper:hover),
.search-input :deep(.el-input__wrapper.is-focus) {
  border-color: #7c3aed;
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.12);
}
.search-input :deep(.el-input__inner) {
  color: #333;
}
.search-input :deep(.el-input__inner::placeholder) {
  color: #b0a6c6;
}

/* ══════════ 分类选择 ══════════ */
.category-selector {
  margin-bottom: 16px;
  background: rgba(255, 255, 255, 0.82);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  padding: 12px 20px;
  border-radius: 14px;
  box-shadow: 0 2px 10px rgba(124, 58, 237, 0.05);
  display: flex;
  align-items: center;
  gap: 12px;
}
.category-title {
  font-weight: 600;
  color: #555;
  white-space: nowrap;
  flex-shrink: 0;
  font-size: 14px;
}
.category-buttons-wrapper {
  flex: 1;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: thin;
  scrollbar-color: #ddd transparent;
}
.category-buttons-wrapper::-webkit-scrollbar {
  height: 4px;
}
.category-buttons-wrapper::-webkit-scrollbar-track {
  background: transparent;
}
.category-buttons-wrapper::-webkit-scrollbar-thumb {
  background: #ddd;
  border-radius: 2px;
}
.category-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: nowrap;
}
.category-buttons .el-button {
  border-radius: 20px;
  font-size: 12px;
  padding: 4px 16px;
  min-width: auto;
  background: #f7f4fd;
  border: 1px solid #ece4f8;
  color: #555;
  transition: all 0.25s;
}
.category-buttons .el-button:hover {
  background: #efe9fb;
  border-color: #d5c8f2;
  color: #333;
  transform: translateY(-1px);
}
.category-buttons .el-button--primary {
  background: linear-gradient(135deg, #7c3aed 0%, #6d28d9 100%);
  border-color: transparent;
  color: #fff;
  box-shadow: 0 2px 10px rgba(124, 58, 237, 0.35);
}
.category-buttons .el-button--primary:hover {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  box-shadow: 0 4px 16px rgba(124, 58, 237, 0.45);
}

/* ══════════ 排序 Tab ══════════ */
.sort-tab-bar {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
}
.sort-tabs {
  display: flex;
  gap: 4px;
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border-radius: 12px;
  padding: 4px;
  border: 1px solid rgba(255, 255, 255, 0.5);
}
.sort-tab-item {
  padding: 6px 18px;
  border-radius: 9px;
  font-size: 13px;
  font-weight: 500;
  color: #666;
  cursor: pointer;
  transition: all 0.25s;
  user-select: none;
  white-space: nowrap;
}
.sort-tab-item:hover {
  color: #333;
  background: rgba(124, 58, 237, 0.05);
}
.sort-tab-item.active {
  background: #fff;
  color: #7c3aed;
  font-weight: 700;
  box-shadow: 0 1px 5px rgba(124, 58, 237, 0.15);
}

/* ══════════ 列表滚动区 ══════════ */
.prompts-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding-right: 4px;
}
.prompts-list::-webkit-scrollbar {
  width: 6px;
}
.prompts-list::-webkit-scrollbar-track {
  background: transparent;
}
.prompts-list::-webkit-scrollbar-thumb {
  background: #d9ceef;
  border-radius: 3px;
}
.prompts-list::-webkit-scrollbar-thumb:hover {
  background: #bfb0e4;
}
.empty-hint {
  margin-top: 50px;
  text-align: center;
}

/* ══════════ 翻转卡片网格 ══════════ */
.prompts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 24px;
  width: 100%;
}
@media (max-width: 1400px) { .prompts-grid { grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); } }
@media (max-width: 1100px) { .prompts-grid { grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); } }
@media (max-width: 800px)  { .prompts-grid { grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); } }

/* 每张卡独立的透视原点：避免滚动容器内透视畸变（与 Demo 一致） */
.card-wrap {
  perspective: 1400px;
}

.prompt-card {
  height: 300px;
  position: relative;
  cursor: pointer;
  transform-style: preserve-3d;
  will-change: transform;
  transition: transform 0.6s cubic-bezier(0.4, 0.2, 0.2, 1);
}
.prompt-card.flip {
  transform: rotateY(180deg);
}
.face {
  position: absolute;
  inset: 0;
  backface-visibility: hidden;
  -webkit-backface-visibility: hidden;
  border-radius: 16px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: #fff;
  border: 1px solid rgba(236, 228, 248, 0.9);
  box-shadow: 0 4px 18px rgba(124, 58, 237, 0.1);
}
.face.front {
  padding: 16px 16px 14px;
  justify-content: flex-start;
  gap: 10px;
}
.face.back {
  transform: rotateY(180deg);
  border-color: #e4daf7;
}

/* ── 正面 ── */
.f-top {
  display: flex;
  align-items: center;
  gap: 8px;
}
.card-category-badge {
  flex-shrink: 0;
  font-size: 11px;
  padding: 3px 10px;
  border-radius: 6px;
  font-weight: 600;
  border: 1px solid;
  background: rgba(255, 255, 255, 0.5);
  white-space: nowrap;
}
.lock-ico {
  flex-shrink: 0;
  font-size: 14px;
}
.card-title {
  flex: 1;
  min-width: 0;
  font-size: 16px;
  font-weight: 800;
  color: #1a1a2e;
  margin: 0;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  word-break: break-word;
}
.card-author-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0;
}
.author-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 12px;
  font-weight: 700;
  flex-shrink: 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.12);
}
.author-name {
  font-size: 12px;
  color: #555;
  max-width: 72px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex-shrink: 0;
}
.author-stats {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: auto;
  flex-shrink: 0;
}
.stat-item {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 11px;
  color: #e8590c;
  font-weight: 600;
  background: rgba(232, 89, 12, 0.06);
  padding: 2px 8px;
  border-radius: 10px;
}
.stat-icon { width: 13px; height: 13px; }
.fire-icon { color: #e8590c; }
.card-date {
  font-size: 11px;
  color: #bbb;
  flex-shrink: 0;
}
.f-hint {
  margin-top: auto;
  font-size: 11px;
  color: #b0a2cc;
  display: flex;
  align-items: center;
  gap: 6px;
}
.f-hint::after { content: '⟳'; font-size: 14px; transition: transform 0.4s; }

/* ── 背面：就地编辑 ── */
.b-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f3eefb;
  background: linear-gradient(180deg, #faf7ff, #fff);
}
.b-label {
  font-size: 11px;
  letter-spacing: 0.2em;
  color: #a99cc9;
  font-weight: 700;
}
.b-flipback {
  border: none;
  background: #f3eefb;
  color: #7c3aed;
  border-radius: 8px;
  padding: 4px 10px;
  font-size: 11px;
  cursor: pointer;
  font-family: inherit;
  font-weight: 700;
  transition: all 0.15s;
}
.b-flipback:hover { background: #e9dff8; }
.b-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.b-body::-webkit-scrollbar { width: 5px; }
.b-body::-webkit-scrollbar-thumb { background: #e5ddf3; border-radius: 3px; }

.b-title {
  font-size: 15px;
  font-weight: 800;
  line-height: 1.4;
  color: #1a1a2e;
  border-radius: 8px;
  padding: 2px 6px;
  margin: 0 -6px;
  cursor: text;
}
.b-title:hover { background: #f0fdf8; }

.b-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}
.b-tag {
  font-size: 11px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 20px;
  cursor: pointer;
  background: #f0fdf8;
  color: #047857;
  border: 1px solid #b7e6d0;
  transition: all 0.18s;
}
.b-tag:hover { background: #dcf5e9; }
.b-tag .x { margin-left: 4px; opacity: 0.6; font-style: normal; }
.b-tag .x:hover { opacity: 1; }
.b-add {
  border: 1.5px dashed #bcdcd0;
  background: transparent;
  border-radius: 20px;
  padding: 3px 12px;
  font-size: 11px;
  color: #6aa48f;
  cursor: pointer;
  font-family: inherit;
  transition: all 0.18s;
}
.b-add:hover { border-color: #10b981; color: #047857; }

.b-desc {
  font-size: 12.5px;
  color: #5b5568;
  line-height: 1.75;
  border-radius: 8px;
  padding: 6px;
  margin: 0 -6px -6px;
  cursor: text;
}
.b-desc:hover { background: #f0fdf8; }

.b-actions {
  display: flex;
  gap: 8px;
  margin-top: auto;
  padding-top: 4px;
}
.ba {
  flex: 1;
  border: 1px solid #e5dcf5;
  background: #fff;
  border-radius: 10px;
  padding: 8px 10px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  font-family: inherit;
  color: #6b5b94;
  transition: all 0.18s;
}
.ba:hover { border-color: #c9b6ee; background: #faf7ff; }
.ba.primary {
  background: linear-gradient(135deg, #8b5cf6, #7c3aed);
  color: #fff;
  border-color: transparent;
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.3);
}
.ba.primary:hover { filter: brightness(1.06); box-shadow: 0 6px 16px rgba(124, 58, 237, 0.4); }

/* ── 内联编辑输入控件 ── */
.ie-input {
  width: 100%;
  border: 2px dashed #7c3aed;
  background: #faf7ff;
  border-radius: 10px;
  padding: 8px 12px;
  font-size: 13px;
  font-family: inherit;
  outline: none;
  color: #1a1a2e;
  transition: border-color 0.2s;
}
.ie-input:focus { border-style: solid; border-color: #7c3aed; box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.12); }
.ie-textarea {
  resize: vertical;
  min-height: 64px;
  line-height: 1.7;
}
.b-tag-input {
  width: 96px;
  border: 1.5px dashed #10b981;
  background: #f0fdf8;
  border-radius: 20px;
  padding: 3px 10px;
  font-size: 11px;
  outline: none;
  font-family: inherit;
  color: #1a1a2e;
}

/* ══════════ 保存条（即时反馈 + 撤销） ══════════ */
.save-bar {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  background: #1e1b2e;
  color: #fff;
  padding: 10px 22px;
  border-radius: 14px;
  font-size: 13px;
  display: flex;
  gap: 14px;
  align-items: center;
  box-shadow: 0 14px 40px rgba(30, 27, 46, 0.45);
  z-index: 300;
}
.sb-msg { font-weight: 600; }
.sb-btn {
  border: none;
  background: #34d399;
  color: #022c22;
  border-radius: 8px;
  padding: 6px 16px;
  font-weight: 700;
  cursor: pointer;
  font-family: inherit;
  font-size: 12.5px;
}
.sb-btn:hover { filter: brightness(1.08); }
.sb-btn.ghost { background: transparent; color: #a7f3d0; }
.sb-btn.ghost:hover { background: rgba(255, 255, 255, 0.08); }
.sb-enter-active, .sb-leave-active { transition: all 0.3s cubic-bezier(0.3, 1.4, 0.4, 1); }
.sb-enter-from, .sb-leave-to { opacity: 0; transform: translateX(-50%) translateY(16px); }

/* ══════════ IDE 工作台预览弹窗（Demo 05） ══════════
   容器尺寸/滚动规则在底部非 scoped 块中定义：
   el-dialog 经 append-to-body 挂载到 body，scoped 的
   :deep(.el-dialog) 无法命中「自身就是该类名」的 dialog 根元素。 */
.ide-preview-body { flex: 1; min-height: 0; display: flex; flex-direction: column; }

/* ── 顶部标签栏 ── */
.ide-tabs { display: flex; align-items: center; gap: 2px; background: #f4f6f9; border-bottom: 1px solid #e2e8f0; padding: 0 8px; flex-shrink: 0; }
.ide-crumb { display: flex; align-items: center; gap: 6px; font-size: 12.5px; color: #64748b; padding: 11px 10px; white-space: nowrap; min-width: 0; }
.ide-crumb .sep { opacity: 0.4; }
.ide-crumb .cur { color: #7c3aed; font-weight: 700; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ide-filetab { display: flex; align-items: center; gap: 8px; background: #fff; border: 1px solid #e2e8f0; border-bottom: none; border-radius: 9px 9px 0 0; padding: 9px 16px; font-size: 13px; font-weight: 600; color: #334155; margin-top: 6px; max-width: 300px; flex-shrink: 0; }
.ide-filetab .name { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ide-filetab .dot { width: 8px; height: 8px; border-radius: 2px; background: linear-gradient(135deg, #7c3aed, #a78bfa); flex-shrink: 0; }
.ide-filetab .lang { font-size: 11px; color: #94a3b8; font-weight: 500; flex-shrink: 0; }
.ide-spacer { flex: 1; }
.ide-icon { width: 34px; height: 34px; border: none; border-radius: 8px; background: transparent; color: #64748b; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s; flex-shrink: 0; font-size: 15px; }
.ide-icon:hover { background: #e2e8f0; color: #0f172a; }
.ide-icon.copy { color: #7c3aed; }
.ide-icon.copy:hover { background: #ede9fe; }

/* ── 状态栏 ── */
.ide-status { display: flex; align-items: center; gap: 14px; padding: 8px 16px; background: #fafbfc; border-bottom: 1px solid #e2e8f0; font-size: 12px; color: #64748b; flex-shrink: 0; flex-wrap: wrap; }
.ide-status b { color: #334155; font-weight: 600; }
.ide-status .tag { padding: 1px 8px; border-radius: 4px; font-size: 11px; border: 1px solid; }
.ide-status .ok { margin-left: auto; color: #16a34a; display: inline-flex; align-items: center; gap: 5px; }
.ide-status .ok i { width: 7px; height: 7px; border-radius: 50%; background: #22c55e; display: inline-block; }

/* ── 主体：左栏 + 右内容 ── */
.ide-main { flex: 1; min-height: 0; display: flex; }
.ide-outline { width: 190px; flex-shrink: 0; border-right: 1px solid #e2e8f0; padding: 14px 10px; overflow-y: auto; display: flex; flex-direction: column; gap: 10px; }
.ol-tabs { display: flex; gap: 3px; background: #f1f5f9; border-radius: 8px; padding: 3px; flex-shrink: 0; }
.ol-tab { flex: 1; border: none; background: transparent; padding: 5px 0; border-radius: 6px; font-size: 12px; font-weight: 700; color: #64748b; cursor: pointer; font-family: inherit; transition: all 0.15s; }
.ol-tab:hover { color: #0f172a; }
.ol-tab.on { background: #fff; color: #7c3aed; box-shadow: 0 1px 3px rgba(15, 23, 42, 0.08); }
.ol-tab .fc { font-size: 10px; background: #ede9fe; color: #7c3aed; border-radius: 999px; padding: 0 6px; margin-left: 3px; font-weight: 700; }
.ol-panel { display: flex; flex-direction: column; }
.ol-outline { gap: 2px; }
.ol-fields { gap: 6px; }
.ol-item { display: flex; align-items: center; gap: 7px; padding: 7px 8px; border-radius: 7px; font-size: 12.5px; color: #475569; cursor: pointer; transition: all 0.15s; }
.ol-item:hover { background: #f1f5f9; }
.ol-item.on { background: #ede9fe; color: #6d28d9; font-weight: 700; }
.ol-item .ic { width: 13px; height: 13px; flex-shrink: 0; opacity: 0.7; }
.ol-item .ln { margin-left: auto; font-size: 10.5px; color: #c0c8d4; }
.ol-field { padding: 8px 10px; border-radius: 8px; border: 1px solid #ece4f8; background: #fdfcff; cursor: pointer; transition: all 0.15s; }
.ol-field:hover { border-color: #c9b6ee; background: #faf7ff; }
.ol-field .f-head { display: flex; align-items: center; justify-content: space-between; gap: 6px; }
.ol-field .f-name { font-size: 12px; font-weight: 700; color: #1a1a2e; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ol-field .req { flex-shrink: 0; font-size: 10px; line-height: 1.4; padding: 2px 9px; border-radius: 999px; background: #fef2f2; color: #dc2626; font-weight: 500; letter-spacing: 0.03em; }
.ol-field .f-desc { font-size: 10.5px; color: #8b83a0; line-height: 1.55; margin-top: 3px; }
.ol-field .f-ph { display: inline-block; max-width: 100%; margin-top: 5px; font-family: "SFMono-Regular", Consolas, monospace; font-size: 10px; color: #7c3aed; background: #f0ecfb; border-radius: 4px; padding: 1px 6px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ol-empty { padding: 18px 8px; text-align: center; font-size: 12px; color: #b0a2cc; }

/* ── 右侧内容面板 ── */
.ide-content { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.ide-file-head { display: flex; align-items: center; gap: 10px; padding: 10px 18px; border-bottom: 1px solid #eef2f6; font-size: 12.5px; color: #64748b; flex-shrink: 0; }
.ide-file-head .fname { font-family: "SFMono-Regular", Consolas, monospace; font-weight: 600; color: #334155; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ide-file-head .lang-chip { font-size: 11px; color: #7c3aed; background: #f0ecfb; padding: 1px 8px; border-radius: 4px; flex-shrink: 0; }
.ide-file-head .size { margin-left: auto; flex-shrink: 0; }
.ide-scroll { flex: 1; min-height: 0; overflow-y: auto; }
.ide-code-panel { max-width: 860px; margin: 0 auto; padding: 16px 12px 40px; }
.code-line { display: flex; gap: 14px; padding: 1px 14px; border-radius: 7px; transition: background 0.12s; }
.code-line:hover { background: rgba(124, 58, 237, 0.05); }
.code-line.on { background: rgba(124, 58, 237, 0.09); }
.code-num { flex-shrink: 0; width: 26px; text-align: right; color: #c0c8d4; font-family: "SFMono-Regular", Consolas, monospace; font-size: 11.5px; line-height: 1.9; user-select: none; }
.code-txt { flex: 1; min-width: 0; color: #3d3a4a; font-family: "SFMono-Regular", Consolas, monospace; font-size: 12.5px; line-height: 1.9; word-break: break-word; white-space: pre-wrap; }
.code-txt .h { color: #7c3aed; font-weight: 600; }
.code-txt .c { color: #64748b; }
.code-txt .s { color: #16a34a; font-weight: 600; }
.ide-encrypted { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 60px 20px; text-align: center; gap: 4px; background: linear-gradient(135deg, rgba(230, 162, 60, 0.08) 0%, rgba(230, 162, 60, 0.04) 100%); border: 1px dashed #e6a23c; border-radius: 12px; }
.encrypted-title { font-size: 16px; font-weight: 600; color: #e6a23c; margin: 8px 0 0; }
.encrypted-sub { font-size: 12.5px; color: #b45309; }
.edit-password-popover { display: flex; flex-direction: column; gap: 12px; }
.popover-hint { font-size: 13px; color: #606266; margin: 0; }
.popover-actions { display: flex; justify-content: flex-end; gap: 8px; }

/* ══════════ 小分类编辑弹窗 ══════════ */
.subcategory-dialog :deep(.el-dialog__header) { display: none; }
.subcategory-dialog :deep(.el-dialog__body) { padding: 24px 28px 16px; }
.subcategory-dialog :deep(.el-dialog__footer) { padding: 0 28px 20px; }
.subcategory-dialog :deep(.el-dialog) { border-radius: 16px; border: 1px solid #ece4f8; }
.subcategory-editor { display: flex; flex-direction: column; gap: 16px; }
.subcategory-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(124, 58, 237, 0.1);
}
.subcategory-title { font-size: 16px; font-weight: 700; color: #1a1a2e; }
.subcategory-count { font-size: 12px; color: #8b83a0; background: #f4f0fb; padding: 4px 10px; border-radius: 12px; }
.subcategory-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 280px;
  overflow-y: auto;
  padding-right: 4px;
}
.subcategory-list::-webkit-scrollbar { width: 4px; }
.subcategory-list::-webkit-scrollbar-thumb { background: #ddd; border-radius: 2px; }
.subcategory-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: #f8f6fd;
  border-radius: 10px;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}
.subcategory-item:hover { background: #efe9fb; border-color: #e2d8f5; }
.tag-input { flex: 1; }
.tag-input :deep(.el-input__wrapper) {
  background: transparent;
  box-shadow: none;
  border: 1px solid transparent;
  border-radius: 8px;
  transition: all 0.2s;
}
.tag-input :deep(.el-input__wrapper:hover),
.tag-input :deep(.el-input__wrapper.is-focus) {
  border-color: #c9b6ee;
  background: #fff;
}
.delete-btn { opacity: 0.5; transition: all 0.2s; }
.subcategory-item:hover .delete-btn { opacity: 1; }
.empty-tags { text-align: center; color: #b0a6c6; font-size: 13px; padding: 24px 0; }
.add-btn {
  width: 100%;
  border-radius: 10px;
  border: 1px dashed #c9b6ee;
  background: transparent;
  color: #7c3aed;
  transition: all 0.2s;
}
.add-btn:hover { border-color: #7c3aed; background: rgba(124, 58, 237, 0.05); }
.dialog-footer { display: flex; justify-content: flex-end; gap: 10px; }
.cancel-btn { border-radius: 9px; padding: 8px 20px; }
.save-btn {
  border-radius: 9px;
  padding: 8px 20px;
  background: linear-gradient(135deg, #7c3aed, #6d28d9);
  border: none;
}
.save-btn:hover { background: linear-gradient(135deg, #8b5cf6, #7c3aed); }
.tag-list-enter-active, .tag-list-leave-active { transition: all 0.3s ease; }
.tag-list-enter-from, .tag-list-leave-to { opacity: 0; transform: translateX(-20px); }

/* ══════════ 编辑提示词弹窗 ══════════ */
.edit-prompt-dialog :deep(.el-dialog) { border-radius: 16px; border: 1px solid #ece4f8; }
.edit-prompt-dialog :deep(.el-dialog__body) { padding: 20px; overflow-y: auto; }
.edit-dialog-header { display: flex; justify-content: space-between; align-items: center; width: 100%; }
.edit-dialog-header span { font-size: 18px; font-weight: 700; color: #1a1a2e; }
.edit-dialog-content { overflow-y: auto; }
.edit-left-panel { padding-right: 10px; }
.edit-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 16px 0 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #eee;
}
.form-label { font-weight: 700; color: #4b445e; }
.field-config-section { margin-top: 8px; }
.empty-fields { padding: 20px; text-align: center; color: #909399; }
.fields-list { display: flex; flex-direction: column; gap: 12px; margin-top: 12px; }
.field-item {
  padding: 12px;
  border: 1px solid #ece4f8;
  border-radius: 10px;
  background: #faf8fe;
}
.field-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.field-name-container { display: flex; align-items: center; gap: 8px; }
.required-tag { font-size: 11px; }
.field-name-input { width: 150px; }
.field-actions { display: flex; align-items: center; gap: 4px; }
.field-config-form { display: flex; flex-direction: column; gap: 8px; }
.field-input { width: 100%; }
.full-width-input { width: 100%; }
.field-options { margin-top: 8px; }
.field-options label { font-size: 12px; color: #606266; margin-bottom: 4px; display: block; }
.form-item-label-with-guide { display: flex; align-items: center; gap: 8px; }
.prompt-guide-trigger { padding: 0 4px; color: #7c3aed; }
.prompt-guide-popover { padding: 8px; }
.prompt-guide-title { font-size: 14px; font-weight: 700; color: #1a1a2e; margin-bottom: 8px; }
.prompt-guide-intro { font-size: 13px; color: #606266; margin-bottom: 12px; }
.prompt-guide-intro code { background: #f4f0fb; padding: 2px 6px; border-radius: 4px; color: #7c3aed; }
.prompt-guide-example { background: #faf8fe; padding: 10px 12px; border-radius: 8px; margin-bottom: 12px; }
.prompt-guide-example-label { font-size: 12px; color: #909399; margin-bottom: 4px; display: block; }
.prompt-guide-example p { font-size: 13px; color: #303133; margin: 0; }
.prompt-guide-example code { background: #f0ecfb; padding: 1px 4px; border-radius: 3px; color: #7c3aed; }
.prompt-guide-tips { display: flex; flex-direction: column; gap: 6px; }
.prompt-guide-tip { display: flex; align-items: flex-start; gap: 8px; font-size: 12px; color: #606266; }
.prompt-guide-tip-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  background: #f0ecfb;
  color: #7c3aed;
  border-radius: 50%;
  font-size: 11px;
  flex-shrink: 0;
}
.edit-dialog-footer { display: flex; justify-content: flex-end; gap: 10px; }
.edit-dialog-footer .el-button--primary {
  background: linear-gradient(135deg, #7c3aed, #6d28d9);
  border-color: transparent;
}
.description-rich-editor { width: 100%; }

/* ══════════ 暗色主题 ══════════ */
:root[data-theme='dark'] .prompt-preview-container {
  background: linear-gradient(180deg, #14101f 0%, #1a132b 40%, #120d1d 100%);
}
:root[data-theme='dark'] .header {
  background: rgba(30, 24, 44, 0.88);
  border-color: rgba(255, 255, 255, 0.08);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}
:root[data-theme='dark'] .header h2 { color: #f4f0ff; }
:root[data-theme='dark'] .search-input :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.12);
}
:root[data-theme='dark'] .search-input :deep(.el-input__wrapper:hover),
:root[data-theme='dark'] .search-input :deep(.el-input__wrapper.is-focus) {
  border-color: #8b5cf6;
  box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.18);
}
:root[data-theme='dark'] .search-input :deep(.el-input__inner) { color: #f4f0ff; }
:root[data-theme='dark'] .search-input :deep(.el-input__inner::placeholder) { color: #6b6480; }
:root[data-theme='dark'] .category-selector {
  background: rgba(30, 24, 44, 0.88);
  border-color: rgba(255, 255, 255, 0.08);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}
:root[data-theme='dark'] .category-title { color: #c8c0de; }
:root[data-theme='dark'] .category-buttons .el-button {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.12);
  color: #c8c0de;
}
:root[data-theme='dark'] .category-buttons .el-button:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.18);
  color: #f4f0ff;
}
:root[data-theme='dark'] .category-buttons .el-button--primary {
  background: linear-gradient(135deg, #7c3aed, #6d28d9);
  border-color: transparent;
  color: #fff;
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.4);
}
:root[data-theme='dark'] .sort-tabs { background: rgba(30, 24, 44, 0.7); border-color: rgba(255, 255, 255, 0.08); }
:root[data-theme='dark'] .sort-tab-item { color: #8d839f; }
:root[data-theme='dark'] .sort-tab-item:hover { color: #c8c0de; background: rgba(255, 255, 255, 0.06); }
:root[data-theme='dark'] .sort-tab-item.active { background: rgba(45, 36, 66, 0.95); color: #a78bfa; box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3); }
:root[data-theme='dark'] .face {
  background: #1a1526;
  border-color: rgba(255, 255, 255, 0.09);
  box-shadow: 0 4px 18px rgba(0, 0, 0, 0.35);
}
:root[data-theme='dark'] .face.back { border-color: rgba(139, 92, 246, 0.3); }
:root[data-theme='dark'] .card-title { color: #f4f0ff; }
:root[data-theme='dark'] .b-title { color: #f4f0ff; }
:root[data-theme='dark'] .card-category-badge { background: rgba(255, 255, 255, 0.06); }
:root[data-theme='dark'] .author-name { color: #c8c0de; }
:root[data-theme='dark'] .card-date { color: #6b6480; }
:root[data-theme='dark'] .f-hint { color: #7d7495; }
:root[data-theme='dark'] .b-head { background: linear-gradient(180deg, #211a33, #1a1526); border-bottom-color: rgba(255, 255, 255, 0.07); }
:root[data-theme='dark'] .b-label { color: #7d7495; }
:root[data-theme='dark'] .b-flipback { background: rgba(139, 92, 246, 0.15); color: #a78bfa; }
:root[data-theme='dark'] .b-title:hover, :root[data-theme='dark'] .b-desc:hover { background: rgba(16, 185, 129, 0.08); }
:root[data-theme='dark'] .b-desc { color: #b9b2cc; }
:root[data-theme='dark'] .b-tag {
  background: rgba(16, 185, 129, 0.1);
  color: #6ee7b7;
  border-color: rgba(16, 185, 129, 0.3);
}
:root[data-theme='dark'] .b-tag:hover { background: rgba(16, 185, 129, 0.18); }
:root[data-theme='dark'] .b-add { color: #6ee7b7; border-color: rgba(16, 185, 129, 0.35); }
:root[data-theme='dark'] .b-add:hover { border-color: #10b981; color: #34d399; }
:root[data-theme='dark'] .ba {
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(255, 255, 255, 0.1);
  color: #c8c0de;
}
:root[data-theme='dark'] .ba:hover { background: rgba(255, 255, 255, 0.08); }
:root[data-theme='dark'] .ba.primary {
  background: linear-gradient(135deg, #7c3aed, #6d28d9);
  color: #fff;
  border-color: transparent;
}
:root[data-theme='dark'] .ie-input {
  background: rgba(139, 92, 246, 0.08);
  border-color: #8b5cf6;
  color: #f4f0ff;
}
:root[data-theme='dark'] .b-tag-input {
  background: rgba(16, 185, 129, 0.1);
  border-color: #10b981;
  color: #f4f0ff;
}
/* ── IDE 预览弹窗暗色 ── */
:root[data-theme='dark'] .ide-tabs { background: #161d2b; border-color: rgba(255, 255, 255, 0.08); }
:root[data-theme='dark'] .ide-crumb { color: #8ea0bd; }
:root[data-theme='dark'] .ide-crumb .cur { color: #a78bfa; }
:root[data-theme='dark'] .ide-filetab { background: #1a2233; border-color: rgba(255, 255, 255, 0.09); color: #dbe3f0; }
:root[data-theme='dark'] .ide-filetab .lang { color: #64748b; }
:root[data-theme='dark'] .ide-icon { color: #8ea0bd; }
:root[data-theme='dark'] .ide-icon:hover { background: rgba(255, 255, 255, 0.1); color: #fff; }
:root[data-theme='dark'] .ide-icon.copy { color: #a78bfa; }
:root[data-theme='dark'] .ide-icon.copy:hover { background: rgba(139, 92, 246, 0.18); }
:root[data-theme='dark'] .ide-status { background: #161d2b; border-color: rgba(255, 255, 255, 0.08); color: #8ea0bd; }
:root[data-theme='dark'] .ide-status b { color: #dbe3f0; }
:root[data-theme='dark'] .ide-status .ok { color: #4ade80; }
:root[data-theme='dark'] .ide-outline { border-color: rgba(255, 255, 255, 0.08); }
:root[data-theme='dark'] .ol-tabs { background: rgba(255, 255, 255, 0.05); }
:root[data-theme='dark'] .ol-tab { color: #8ea0bd; }
:root[data-theme='dark'] .ol-tab:hover { color: #fff; }
:root[data-theme='dark'] .ol-tab.on { background: #1a2233; color: #a78bfa; }
:root[data-theme='dark'] .ol-tab .fc { background: rgba(139, 92, 246, 0.2); color: #a78bfa; }
:root[data-theme='dark'] .ol-item { color: #a8b6cf; }
:root[data-theme='dark'] .ol-item:hover { background: rgba(255, 255, 255, 0.06); }
:root[data-theme='dark'] .ol-item.on { background: rgba(139, 92, 246, 0.18); color: #a78bfa; }
:root[data-theme='dark'] .ol-item .ln { color: #4a5a76; }
:root[data-theme='dark'] .ol-field { background: rgba(255, 255, 255, 0.04); border-color: rgba(255, 255, 255, 0.09); }
:root[data-theme='dark'] .ol-field:hover { background: rgba(255, 255, 255, 0.07); }
:root[data-theme='dark'] .ol-field .f-name { color: #f4f0ff; }
:root[data-theme='dark'] .ol-field .f-desc { color: #8d839f; }
:root[data-theme='dark'] .ol-field .f-ph { background: rgba(139, 92, 246, 0.14); color: #a78bfa; }
:root[data-theme='dark'] .ide-file-head { border-color: rgba(255, 255, 255, 0.07); color: #8ea0bd; }
:root[data-theme='dark'] .ide-file-head .fname { color: #dbe3f0; }
:root[data-theme='dark'] .ide-file-head .lang-chip { background: rgba(139, 92, 246, 0.14); color: #a78bfa; }
:root[data-theme='dark'] .code-line:hover { background: rgba(139, 92, 246, 0.1); }
:root[data-theme='dark'] .code-line.on { background: rgba(139, 92, 246, 0.16); }
:root[data-theme='dark'] .code-num { color: #4a5a76; }
:root[data-theme='dark'] .code-txt { color: #c9c4d6; }
:root[data-theme='dark'] .code-txt .h { color: #a78bfa; }
:root[data-theme='dark'] .code-txt .c { color: #64748b; }
:root[data-theme='dark'] .code-txt .s { color: #4ade80; }
:root[data-theme='dark'] .subcategory-dialog :deep(.el-dialog),
:root[data-theme='dark'] .edit-prompt-dialog :deep(.el-dialog),
:root[data-theme='dark'] .preview-dialog :deep(.el-dialog) {
  background: linear-gradient(180deg, #14101f 0%, #1a132b 40%, #120d1d 100%);
  border: 1px solid rgba(255, 255, 255, 0.08);
}
:root[data-theme='dark'] .subcategory-title { color: #f4f0ff; }
:root[data-theme='dark'] .subcategory-item { background: rgba(255, 255, 255, 0.04); }
:root[data-theme='dark'] .subcategory-item:hover { background: rgba(255, 255, 255, 0.07); }
:root[data-theme='dark'] .edit-dialog-header span { color: #f4f0ff; }
:root[data-theme='dark'] .form-label { color: #c8c0de; }
:root[data-theme='dark'] .field-item { background: rgba(255, 255, 255, 0.04); border-color: rgba(255, 255, 255, 0.09); }
:root[data-theme='dark'] .prompt-guide-title { color: #f4f0ff; }
:root[data-theme='dark'] .ide-encrypted {
  background: linear-gradient(135deg, rgba(230, 162, 60, 0.12) 0%, rgba(230, 162, 60, 0.06) 100%);
  border-color: rgba(230, 162, 60, 0.55);
}
</style>

<!-- 非 scoped：覆盖 append-to-body 的 el-dialog 容器尺寸与暗色背景 -->
<style>
.ide-preview.el-dialog {
  display: flex !important;
  flex-direction: column !important;
  height: 86vh !important;
  max-height: 86vh !important;
  margin-left: auto !important;
  margin-right: auto !important;
  border-radius: 14px;
  border: 1px solid #e2e8f0;
  overflow: hidden;
  padding: 0 !important;
  box-shadow: 0 24px 60px rgba(15, 23, 42, 0.18);
}
.ide-preview .el-overlay-dialog {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 0;
}
.ide-preview .el-dialog__header { display: none; }
.ide-preview .el-dialog__body {
  padding: 0;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
:root[data-theme='dark'] .preview-dialog.el-dialog {
  background: linear-gradient(180deg, #14101f 0%, #1a132b 40%, #120d1d 100%) !important;
  border: 1px solid rgba(255, 255, 255, 0.08);
}
:root[data-theme='dark'] .preview-dialog .el-dialog__body { background: transparent !important; }
:root[data-theme='dark'] .preview-dialog .el-scrollbar,
:root[data-theme='dark'] .preview-dialog .el-scrollbar__wrap,
:root[data-theme='dark'] .preview-dialog .el-scrollbar__view { background: transparent !important; }
</style>
