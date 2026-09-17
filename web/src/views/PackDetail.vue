<template>
  <div class="pack-detail-container">
    <div class="pack-detail-header">
      <div class="pack-detail-title-block">
        <button class="back-btn" @click="goBack" title="返回">
          <el-icon><ArrowLeft /></el-icon>
        </button>
        <h2 class="pack-detail-title">{{ packName }}</h2>
        <span class="pack-detail-count">{{ filteredPrompts.length }} 张卡片</span>
      </div>
      <div class="pack-detail-search">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索提示词..."
          clearable
          prefix-icon="Search"
          class="pack-search-input"
        />
        <el-button type="primary" @click="openCreateInCurrentPack">
          <el-icon><Plus /></el-icon>
          新增提示词
        </el-button>
      </div>
    </div>

    <!-- 排序 Tab -->
    <div class="pack-sort-tab-bar">
      <div class="pack-sort-tabs">
        <span
          v-for="tab in sortTabs"
          :key="tab.value"
          class="pack-sort-tab-item"
          :class="{ active: selectedSort === tab.value }"
          @click="selectedSort = tab.value"
        >
          <el-icon v-if="tab.icon" class="pack-sort-tab-icon">
            <component :is="tab.icon" />
          </el-icon>
          {{ tab.label }}
        </span>
      </div>
    </div>

    <!-- 提示词卡片列表 -->
    <div v-if="filteredPrompts.length === 0" class="pack-empty-hint">
      <el-empty :description="searchKeyword ? '没有匹配的提示词' : '该卡包暂无可预览的卡片'" />
    </div>
    <div v-else class="pack-prompts-grid">
      <div
        v-for="prompt in filteredPrompts"
        :key="prompt.id"
        class="pack-prompt-card"
        :class="{ encrypted: prompt.card_type === 'encrypted' }"
        @click="handleEdit(prompt)"
      >
        <span
          class="pack-card-category-badge"
          :style="{ borderColor: getCategoryColor(prompt.category), color: getCategoryColor(prompt.category) }"
        >
          {{ getCategoryLabel(prompt.category) }}
        </span>

        <div class="pack-card-title-row">
          <h3 class="pack-card-title" :title="prompt.name">
            <el-icon v-if="prompt.card_type === 'encrypted'" class="pack-card-lock"><Lock /></el-icon>
            {{ prompt.name }}
          </h3>
        </div>

        <div class="pack-card-author-row">
          <div class="pack-author-avatar" :style="{ background: getAvatarColor(prompt.creator_name || prompt.name) }">
            {{ (prompt.creator_name || '匿').charAt(0) }}
          </div>
          <span class="pack-author-name" :title="prompt.creator_name">
            {{ prompt.creator_name || '匿名用户' }}
          </span>
          <span class="pack-card-stats">
            <el-icon class="pack-stat-icon"><Star /></el-icon>
            {{ prompt.use_count ?? prompt.order_num ?? 0 }}
          </span>
          <span class="pack-card-date">{{ formatDate(prompt.created_at) }}</span>
        </div>

        <div class="pack-card-body-preview" :title="getPromptDescriptionPreview(prompt)">
          {{ getPromptDescriptionPreview(prompt) || '暂无简介' }}
        </div>

        <div v-if="(prompt.subcategories || []).length > 0" class="pack-card-footer-tags">
          <span
            v-for="(subcat, index) in (prompt.subcategories || []).slice(0, 3)"
            :key="index"
            class="pack-subcategory-tag"
            :style="{
              background: getTagColor(subcat).bg,
              color: getTagColor(subcat).text,
              borderColor: getTagColor(subcat).border
            }"
          >
            {{ subcat }}
          </span>
          <span
            v-if="(prompt.subcategories || []).length > 3"
            class="pack-more-tags"
          >+{{ (prompt.subcategories || []).length - 3 }}</span>
        </div>
        <div v-else class="pack-card-footer-tags pack-card-footer-tags--empty"></div>
      </div>
    </div>

    <!-- 编辑提示词弹窗 -->
    <el-dialog
      v-model="editDialogVisible"
      :title="isCreating ? '新增提示词' : '编辑提示词'"
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
                    <div class="field-insert-hint">插入后可配合"提取字段"或右侧字段配置继续完善</div>
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
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editSaving" @click="handleSubmitAndPreview">预览并保存</el-button>
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
            <div class="field-insert-hint">插入后可配合"提取字段"或右侧字段配置继续完善</div>
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

    <!-- 解锁密码弹窗 -->
    <el-dialog
      v-model="unlockPasswordDialogVisible"
      title="输入密码以解锁卡片"
      width="400px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form :model="unlockPasswordForm" label-position="top">
        <el-form-item label="密码" required>
          <el-input
            v-model="unlockPasswordForm.password"
            type="password"
            placeholder="请输入卡片密码"
            show-password
            @keyup.enter="unlockAndEdit"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="unlockPasswordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="unlockAndEdit">解锁并编辑</el-button>
      </template>
    </el-dialog>

    <!-- 设置密码弹窗 -->
    <el-dialog
      v-model="passwordDialogVisible"
      title="设置卡片密码"
      width="400px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form :model="passwordForm" label-position="top">
        <el-form-item label="密码" required>
          <el-input
            v-model="passwordForm.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="确认密码" required>
          <el-input
            v-model="passwordForm.confirmPassword"
            type="password"
            placeholder="请再次输入密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="setPassword">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Plus, Search, Star, Lock, CopyDocument, ArrowUp, ArrowDown, Delete, Document, DocumentAdd, View, Setting, QuestionFilled, MagicStick, FullScreen, WarningFilled, InfoFilled, Share, Tickets } from '@element-plus/icons-vue'

const DEFAULT_CATEGORIES = ['默认', '写作要求', '写作风格', '续写', '脑洞', '书名', '简介', '大纲', '细纲', '黄金开篇', '金手指', '名字', '人设', '世界观', '卡包']
import { promptAPI } from '@/api'
import type { Prompt } from '@/types'
import SplitRichTextEditor from '@/components/SplitRichTextEditor.vue'

const route = useRoute()
const router = useRouter()

const packName = computed(() => decodeURIComponent(String(route.params.name || '')))

const prompts = ref<Prompt[]>([])
const searchKeyword = ref('')
const selectedSort = ref('latest')

// 编辑弹窗相关
const editDialogVisible = ref(false)
const editSaving = ref(false)
const isCreating = ref(false)
const formData = ref({
  id: 0,
  name: '',
  description: '',
  content: '',
  category: '',
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

// 字段配置
type FieldConfig = {
  name: string
  label: string
  type: 'text' | 'textarea' | 'select'
  options: string[]
  optionLabels: string[]
  optionsText: string
  optionLabelsText: string
  description: string
  required: boolean
  defaultValue: string
}
const fieldsConfig = ref<FieldConfig[]>([])

// 标签
const formSubcategories = ref<string[]>([])
const newSubcategory = ref('')

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

// 分类选项
const categoryOptions = computed(() => {
  const names = new Set<string>(DEFAULT_CATEGORIES)
  if (formData.value.category) {
    names.add(formData.value.category)
  }
  prompts.value.forEach((p) => {
    if (p.category) names.add(p.category)
  })
  return Array.from(names)
})

// 加密/解密常量
const ENCRYPT_MARKER = 'XNC1:'
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
  bytes.forEach(b => (binary += String.fromCharCode(b)))
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

// 当前会话的明文密码（用于保存时加密正文，不落库）
const currentSessionPassword = ref('')

// 解锁弹窗
const unlockPasswordDialogVisible = ref(false)
const unlockPasswordForm = ref({ password: '' })
let unlockingPrompt: Prompt | null = null

// 密码弹窗
const passwordDialogVisible = ref(false)
const passwordForm = ref({
  password: '',
  confirmPassword: ''
})

const handleCardTypeChange = (value: string) => {
  if (value === 'normal') {
    formData.value.password = null
    currentSessionPassword.value = ''
  }
}

const openSetPasswordDialog = () => {
  passwordForm.value = {
    password: '',
    confirmPassword: ''
  }
  passwordDialogVisible.value = true
}

const removeCardPassword = () => {
  formData.value.card_type = 'normal'
  formData.value.password = null
  currentSessionPassword.value = ''
  ElMessage.success('已移除密码，卡片已切换为普通卡片')
}

const setPassword = async () => {
  if (!passwordForm.value.password.trim()) {
    ElMessage.warning('请输入密码')
    return
  }
  if (passwordForm.value.password !== passwordForm.value.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }
  const hashed = await hashPassword(passwordForm.value.password.trim())
  formData.value.card_type = 'encrypted'
  formData.value.password = hashed
  currentSessionPassword.value = passwordForm.value.password.trim()
  passwordDialogVisible.value = false
  ElMessage.success('密码设置成功，卡片已切换为加密卡片')
}

const sortTabs: Array<{ label: string; value: string; icon?: any }> = [
  { label: '最新', value: 'latest' },
  { label: '最热', value: 'hottest' },
  { label: '名称A-Z', value: 'name_asc' },
]

const fetchPrompts = async () => {
  const res = await promptAPI.getAll()
  if (res.success && res.data) {
    prompts.value = res.data
  }
}

const packPrompts = computed(() => {
  return prompts.value.filter((p) => (p.category || '默认') === packName.value)
})

const filteredPrompts = computed(() => {
  let result = packPrompts.value
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (keyword) {
    result = result.filter(
      (p) =>
        p.name.toLowerCase().includes(keyword) ||
        p.content.toLowerCase().includes(keyword) ||
        (p.description || '').toLowerCase().includes(keyword)
    )
  }
  const list = [...result]
  switch (selectedSort.value) {
    case 'latest':
      list.sort((a, b) => {
        const ta = a.created_at ? new Date(a.created_at).getTime() : 0
        const tb = b.created_at ? new Date(b.created_at).getTime() : 0
        return tb - ta
      })
      break
    case 'hottest':
      list.sort((a, b) => (b.order_num || 0) - (a.order_num || 0))
      break
    case 'name_asc':
      list.sort((a, b) => a.name.localeCompare(b.name, 'zh-CN'))
      break
  }
  return list
})

const formatDate = (dateStr: string | undefined) => {
  if (!dateStr) return '未知时间'
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

const stripHtml = (value: string) => {
  if (!value) return ''
  if (typeof window === 'undefined') return value.replace(/<[^>]+>/g, ' ')
  const temp = document.createElement('div')
  temp.innerHTML = value
  return temp.textContent || temp.innerText || ''
}

const getPromptDescriptionPreview = (prompt: Prompt) => {
  const description = stripHtml(prompt.description || '').trim()
  if (!description) return ''
  return `${description.slice(0, 100)}${description.length > 100 ? '...' : ''}`
}

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

const categoryColorMap: Record<string, string> = {}
const categoryLabelMap: Record<string, string> = {}
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
  '#a18cd1', '#fccb90', '#f5576c', '#764ba2', '#00f2fe',
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
const tagColorCache: Record<string, { bg: string; text: string; border: string }> = {}
const getTagColor = (tagText: string) => {
  if (!tagColorCache[tagText]) {
    let hash = 0
    for (let i = 0; i < tagText.length; i++) {
      hash = tagText.charCodeAt(i) + ((hash << 5) - hash)
    }
    tagColorCache[tagText] = tagColorPalette[Math.abs(hash) % tagColorPalette.length]
  }
  return tagColorCache[tagText]
}

const goBack = () => {
  router.push({ name: 'Prompts' })
}

const openCreateInCurrentPack = () => {
  isCreating.value = true
  formData.value = {
    id: 0,
    name: '',
    description: '',
    content: '',
    category: packName.value,
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
  currentSessionPassword.value = ''
  fieldExpanded.value = 0
  editDialogVisible.value = true
}

const handleEdit = (prompt: Prompt) => {
  if (prompt.card_type === 'encrypted' && prompt.content.startsWith(ENCRYPT_MARKER)) {
    unlockingPrompt = prompt
    unlockPasswordForm.value = { password: '' }
    unlockPasswordDialogVisible.value = true
    return
  }
  openEditDialog(prompt)
}

const unlockAndEdit = async () => {
  if (!unlockingPrompt) return
  const password = unlockPasswordForm.value.password.trim()
  if (!password) {
    ElMessage.warning('请输入密码')
    return
  }
  const isValid = await verifyPassword(password, unlockingPrompt.password || '')
  if (!isValid) {
    ElMessage.error('密码错误')
    return
  }
  const decrypted = await decryptContent(unlockingPrompt.content, password)
  if (decrypted === null) {
    ElMessage.error('解密失败，内容已损坏或密码不正确')
    return
  }
  unlockPasswordDialogVisible.value = false
  currentSessionPassword.value = password
  openEditDialog({ ...unlockingPrompt, content: decrypted })
}

const openEditDialog = (prompt: Prompt) => {
  isCreating.value = false
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
  // 字段配置
  fieldsConfig.value = []
  fieldExpanded.value = 0
  if (prompt.fields && prompt.fields.length > 0) {
    fieldsConfig.value = prompt.fields.map((field) => ({
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
    extractFieldsFromContent(prompt.content)
  }
  // 标签
  formSubcategories.value = [...(prompt.subcategories || [])]
  newSubcategory.value = ''

  editDialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formData.value.name.trim()) {
    ElMessage.warning('请输入名称')
    return
  }
  if (!formData.value.category.trim()) {
    ElMessage.warning('请输入卡包分类')
    return
  }
  if (!formData.value.content.trim()) {
    ElMessage.warning('请输入提示词内容')
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

  editSaving.value = true
  try {
    let saveContent = formData.value.content
    if (formData.value.card_type === 'encrypted' && currentSessionPassword.value) {
      saveContent = await encryptContent(saveContent, currentSessionPassword.value)
    }
    const payload = {
      name: formData.value.name,
      description: formData.value.description,
      content: saveContent,
      category: formData.value.category,
      order_num: formData.value.order_num,
      creator_name: formData.value.creator_name,
      version: formData.value.version,
      card_type: formData.value.card_type,
      password: formData.value.password,
      fields: fieldsConfig.value.map((field) => ({
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

    let res
    if (isCreating.value) {
      res = await promptAPI.create(payload)
    } else {
      res = await promptAPI.update(formData.value.id, payload)
    }

    if (res.success) {
      ElMessage.success(isCreating.value ? '创建成功' : '保存成功')
      editDialogVisible.value = false
      currentSessionPassword.value = ''
      await fetchPrompts()
    } else {
      ElMessage.error(res.message || (isCreating.value ? '创建失败' : '保存失败'))
    }
  } catch (e: any) {
    ElMessage.error(e?.message || (isCreating.value ? '创建失败' : '保存失败'))
  } finally {
    editSaving.value = false
  }
}

// 保存并打开预览测试（对齐灵犀"预览并保存"）
const handleSubmitAndPreview = async () => {
  await handleSubmit()
  if (!editDialogVisible.value) {
    // 保存成功（弹窗已关闭）后打开预览测试
    openPreviewTest()
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

const removeField = (index: number) => {
  fieldsConfig.value.splice(index, 1)
}

const moveField = (fromIndex: number, toIndex: number) => {
  if (fromIndex < 0 || fromIndex >= fieldsConfig.value.length) return
  if (toIndex < 0 || toIndex >= fieldsConfig.value.length) return
  const [moved] = fieldsConfig.value.splice(fromIndex, 1)
  if (moved) fieldsConfig.value.splice(toIndex, 0, moved)
}

const updateFieldName = (index: number, newName: string) => {
  if (!newName.trim()) {
    ElMessage.warning('字段名称不能为空')
    return
  }
  const field = fieldsConfig.value[index]
  if (!field) return
  const oldName = field.name
  const exists = fieldsConfig.value.some((f, i) => i !== index && f.name === newName)
  if (exists) {
    ElMessage.warning('字段名称不能重复')
    return
  }
  if (oldName && oldName !== newName) {
    formData.value.content = formData.value.content.replace(
      new RegExp(`\\$\\{${oldName}\\}`, 'g'),
      `$\{${newName}\}`
    )
  }
}

const copyFieldName = (fieldName: string) => {
  if (!fieldName.trim()) {
    ElMessage.warning('字段名称不能为空')
    return
  }
  const placeholder = `$\{${fieldName}\}`
  navigator.clipboard.writeText(placeholder).then(
    () => ElMessage.success('占位符已复制'),
    () => ElMessage.error('复制失败')
  )
}

const updateFieldOptions = (index: number) => {
  const field = fieldsConfig.value[index]
  if (!field) return
  field.options = field.optionsText
    .split('\n')
    .map((s) => s.trim())
    .filter(Boolean)
  field.optionLabels = field.optionLabels.slice(0, field.options.length)
  field.optionLabelsText = field.optionLabels.join('\n')
}

const updateFieldOptionLabels = (index: number) => {
  const field = fieldsConfig.value[index]
  if (!field) return
  const labels = field.optionLabelsText
    .split('\n')
    .map((s) => s.trim())
    .slice(0, field.options.length)
  while (labels.length > 0 && !labels[labels.length - 1]) labels.pop()
  field.optionLabels = labels
  field.optionLabelsText = labels.join('\n')
}

const addFormSubcategory = () => {
  const value = newSubcategory.value.trim()
  if (value && !formSubcategories.value.includes(value)) {
    formSubcategories.value.push(value)
    newSubcategory.value = ''
  }
}

const removeFormSubcategory = (index: number) => {
  formSubcategories.value.splice(index, 1)
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

watch(packName, () => {
  // 切换卡包时无需重新请求数据，但清空搜索关键字
  searchKeyword.value = ''
})

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

onMounted(() => {
  fetchPrompts()
})
</script>

<style scoped>
.pack-detail-container {
  padding: 24px 28px;
  background: #f5f7fa;
  min-height: 100%;
  box-sizing: border-box;
}

.pack-detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
  flex-wrap: wrap;
}

.pack-detail-title-block {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.back-btn {
  width: 36px;
  height: 36px;
  border: 1px solid #e2e8f0;
  background: #ffffff;
  color: #475569;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 16px;
}

.back-btn:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
  color: #1e293b;
}

.pack-detail-title {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pack-detail-count {
  font-size: 12px;
  font-weight: 500;
  color: #64748b;
  background: #e2e8f0;
  padding: 2px 10px;
  border-radius: 999px;
}

.pack-detail-search {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  justify-content: flex-end;
  min-width: 320px;
}

.pack-search-input {
  width: 280px;
}

/* 排序 Tab */
.pack-sort-tab-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #ffffff;
  border-radius: 14px;
  padding: 8px 14px;
  margin-bottom: 18px;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  border: 1px solid #eef0f3;
}

.pack-sort-tabs {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.pack-sort-tab-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 14px;
  border-radius: 999px;
  font-size: 13px;
  color: #475569;
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.pack-sort-tab-item:hover {
  background: #f1f5f9;
  color: #1e293b;
}

.pack-sort-tab-item.active {
  background: linear-gradient(135deg, #3b82f6 0%, #6366f1 100%);
  color: #ffffff;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.25);
}

.pack-sort-tab-icon {
  font-size: 14px;
}

/* 卡片网格 */
.pack-prompts-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

@media (max-width: 1280px) {
  .pack-prompts-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 960px) {
  .pack-prompts-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.pack-prompt-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px 18px 14px;
  background: #ffffff;
  border-radius: 14px;
  border: 1px solid #eef0f3;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  cursor: pointer;
  transition: all 0.22s cubic-bezier(0.4, 0, 0.2, 1);
  min-height: 168px;
}

.pack-prompt-card:hover {
  border-color: #c7d2fe;
  box-shadow: 0 10px 24px rgba(59, 130, 246, 0.12);
  transform: translateY(-2px);
}

.pack-prompt-card.encrypted {
  background: linear-gradient(180deg, #fffbeb 0%, #ffffff 100%);
}

.pack-card-category-badge {
  position: absolute;
  top: 12px;
  right: 12px;
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 6px;
  border: 1px solid transparent;
  background: #ffffff;
}

.pack-card-title-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding-right: 56px;
}

.pack-card-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 6px;
}

.pack-card-lock {
  color: #d97706;
  font-size: 14px;
}

.pack-card-author-row {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #64748b;
}

.pack-author-avatar {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  color: #ffffff;
  font-size: 12px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.pack-author-name {
  max-width: 90px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #475569;
}

.pack-card-stats {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  color: #f59e0b;
  font-weight: 500;
}

.pack-stat-icon {
  font-size: 12px;
}

.pack-card-date {
  margin-left: auto;
  color: #94a3b8;
  font-size: 12px;
}

.pack-card-body-preview {
  font-size: 13px;
  color: #475569;
  line-height: 1.55;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  min-height: 60px;
}

.pack-card-footer-tags {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  margin-top: auto;
}

.pack-card-footer-tags--empty {
  min-height: 22px;
}

/* 编辑弹窗布局 */
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
.description-rich-editor {
  width: 100%;
  height: 320px;
}
.description-rich-editor :deep(.split-rich-text-editor) {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  background: #ffffff;
  box-shadow: none;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}
.description-rich-editor :deep(.split-rich-text-editor:focus-within) {
  border-color: #d1d5db;
  box-shadow:
    0 1px 3px rgba(0, 0, 0, 0.04),
    0 4px 12px rgba(0, 0, 0, 0.05);
}
.description-rich-editor :deep(.toolbar) {
  background: #f8fafc;
  border-bottom: 1px solid #e5e7eb;
  padding: 8px 12px;
  gap: 4px;
}
.description-rich-editor :deep(.toolbar-button) {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  color: #374151;
  min-width: 28px;
  height: 24px;
  padding: 0 6px;
  border-radius: 6px;
  transition: all 0.15s ease;
}
.description-rich-editor :deep(.toolbar-button:hover:not(:disabled)) {
  border-color: #d1d5db;
  background: #f3f4f6;
  color: #1f2937;
  transform: none;
  box-shadow: none;
}
.description-rich-editor :deep(.toolbar-button:disabled) {
  opacity: 0.5;
}
.description-rich-editor :deep(.toolbar-icon) {
  width: 12px;
  height: 12px;
}
.description-rich-editor :deep(.toolbar-glyph) {
  font-size: 11px;
  font-weight: 500;
}
.description-rich-editor :deep(.editor-surface) {
  padding: 16px 20px;
  color: #1f2937;
  font-size: 14px;
  line-height: 1.7;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}
.description-rich-editor :deep(.editor-surface.is-empty::before) {
  color: #9ca3af;
  font-size: 14px;
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

.pack-subcategory-tag {
  font-size: 11px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 6px;
  border: 1px solid transparent;
  white-space: nowrap;
}

.pack-more-tags {
  font-size: 11px;
  color: #94a3b8;
}

.pack-empty-hint {
  background: #ffffff;
  border-radius: 14px;
  padding: 48px 0;
  border: 1px dashed #cbd5e1;
}

/* 暗色主题 */
:root[data-theme='dark'] .pack-detail-container {
  background: #0f172a;
}

:root[data-theme='dark'] .back-btn {
  background: #1e293b;
  border-color: #334155;
  color: #cbd5e1;
}

:root[data-theme='dark'] .back-btn:hover {
  background: #334155;
  border-color: #475569;
  color: #f1f5f9;
}

:root[data-theme='dark'] .pack-detail-title {
  color: #f1f5f9;
}

:root[data-theme='dark'] .pack-detail-count {
  background: #334155;
  color: #cbd5e1;
}

:root[data-theme='dark'] .pack-sort-tab-bar {
  background: #1e293b;
  border-color: #334155;
}

:root[data-theme='dark'] .pack-sort-tab-item {
  color: #cbd5e1;
}

:root[data-theme='dark'] .pack-sort-tab-item:hover {
  background: #334155;
  color: #f1f5f9;
}

:root[data-theme='dark'] .pack-prompt-card {
  background: #1e293b;
  border-color: #334155;
}

:root[data-theme='dark'] .pack-prompt-card:hover {
  border-color: #818cf8;
  box-shadow: 0 10px 24px rgba(99, 102, 241, 0.2);
}

:root[data-theme='dark'] .pack-prompt-card.encrypted {
  background: linear-gradient(180deg, rgba(180, 83, 9, 0.15) 0%, #1e293b 100%);
}

:root[data-theme='dark'] .pack-card-category-badge {
  background: #1e293b;
}

:root[data-theme='dark'] .pack-card-title {
  color: #f1f5f9;
}

:root[data-theme='dark'] .pack-author-name {
  color: #cbd5e1;
}

:root[data-theme='dark'] .pack-card-date {
  color: #64748b;
}

:root[data-theme='dark'] .pack-card-body-preview {
  color: #94a3b8;
}

:root[data-theme='dark'] .pack-empty-hint {
  background: #1e293b;
  border-color: #334155;
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
