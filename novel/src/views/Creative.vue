<template>
  <div class="creative-page">
    <div class="page-header">
      <div class="header-content">
        <h1>创意区</h1>
        <p class="subtitle">选择生成器，开启你的创作之旅</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" class="create-generator-btn" @click="openGeneratorManager">
          <el-icon><Setting /></el-icon>
          生成器管理
        </el-button>
        <el-button type="primary" class="history-btn" @click="showHistoryDialog = true">
          <el-icon><Document /></el-icon>
          历史记录
        </el-button>
      </div>
    </div>

    <div class="generators-grid">
      <div
        v-for="generator in generators"
        :key="generator.id"
        class="generator-card"
        @click="openGenerator(generator)"
      >
        <div class="card-texture"></div>
        <div class="card-content-left">
          <div class="card-header">
            <div class="card-title">{{ generator.name }}</div>
            <el-tag v-if="generator.isNew" size="small" type="danger" class="new-tag">NEW</el-tag>
          </div>
          <div class="card-desc">{{ generator.description || '暂无描述' }}</div>
        </div>
        <div class="card-icon-wrapper">
          <div class="card-circle-decoration" :class="`deco-${generator.icon}`"></div>
        </div>
      </div>
    </div>

    <!-- 生成器弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="currentGenerator?.name"
      width="88vw"
      align-center
      destroy-on-close
      class="generator-dialog"
      modal-class="generator-dialog-modal"
      :show-close="true"
    >
      <div class="dialog-content">
        <!-- 左侧：固定提示词展示区 -->
        <div class="pinned-prompts-area">
          <div class="pinned-header">
            <el-icon><Document /></el-icon>
            <span>固定提示词</span>
            <el-button
              type="primary"
              size="small"
              @click="promptSelectDialogVisible = true"
            >
              <el-icon><Plus /></el-icon>
              选择提示词
            </el-button>
          </div>
          <div class="pinned-list">
            <div
              v-for="prompt in pinnedPrompts"
              :key="prompt.id"
              class="pinned-item"
              :class="{ disabled: !prompt.enabled }"
            >
              <div class="pinned-item-header">
                <el-switch
                  v-model="prompt.enabled"
                  size="small"
                  @click.stop
                />
                <span class="pinned-item-name">{{ prompt.name }}</span>
                <el-tag size="small" type="info">{{ prompt.category }}</el-tag>
                <el-icon class="remove-icon" @click="removePinnedPrompt(prompt.id)"><Close /></el-icon>
              </div>
              <div class="pinned-item-desc"></div>
            </div>
            <div v-if="pinnedPrompts.length === 0" class="empty-pinned">
              <el-icon :size="32"><Document /></el-icon>
              <p>点击右上角按钮选择提示词</p>
            </div>
          </div>
        </div>

        <!-- 右侧：功能区 -->
        <div class="function-area">
          <!-- 固定提示词编辑 -->
          <div class="selection-section">
            <div class="section-header">
              <el-icon><Setting /></el-icon>
              <span>发送给AI的提示词</span>
              <div class="fixed-prompt-actions">
                <el-button
                  type="primary"
                  link
                  size="small"
                  @click="saveFixedPrompt"
                  :disabled="!hasFixedPromptChanged"
                >
                  <el-icon><Check /></el-icon>
                  保存修改
                </el-button>
                <el-button
                  type="primary"
                  link
                  size="small"
                  @click="showFixedPrompt"
                >
                  <el-icon><View /></el-icon>
                  查看完整提示词
                </el-button>
              </div>
            </div>
            <div v-if="false" class="fixed-prompt-preview">
              <el-input
                v-model="editableFixedPrompt"
                type="textarea"
                :rows="4"
                placeholder="选择提示词后将显示完整提示词，您可以直接在此编辑..."
              />
            </div>
            <div v-if="hasFixedPromptChanged" class="fixed-prompt-hint">
              <el-tag size="small" type="warning">已修改</el-tag>
              <span>提示词已修改，记得点击保存</span>
            </div>
          </div>

          <!-- 选择模型 -->
          <div class="selection-section">
            <div class="section-header">
              <el-icon><Cpu /></el-icon>
              <span>选择模型</span>
            </div>
            <el-select
              v-model="selectedModel"
              placeholder="请选择AI模型"
              class="full-width"
            >
              <el-option
                v-for="model in models"
                :key="model.id"
                :label="model.name"
                :value="model.id"
              >
                <div class="model-option">
                  <span class="model-name">{{ model.name }}</span>
                  <el-tag v-if="model.provider_name" size="small" type="info">{{ model.provider_name }}</el-tag>
                </div>
              </el-option>
            </el-select>
          </div>

          <!-- 任务参数 -->
          <div v-if="hasAnyFields" class="selection-section">
            <div class="section-header">
              <el-icon><EditPen /></el-icon>
              <span>任务参数</span>
            </div>
            
            <!-- 如果有字段配置，显示动态生成的输入框 -->
            <div class="fields-input-container">
              <div
                v-for="field in allFields"
                :key="field.name"
                class="field-input-item"
              >
                <label class="field-label">
                  {{ field.label }}
                  <el-tag 
                    :type="field.required !== false ? 'danger' : 'info'" 
                    size="small"
                    class="required-tag"
                  >
                    {{ field.required !== false ? '必选' : '选填' }}
                  </el-tag>
                </label>
                <div v-if="field.description" class="field-description">{{ field.description }}</div>
                
                <!-- 单行文本 -->
                <el-input
                  v-if="field.type === 'text'"
                  v-model="fieldValues[field.name]"
                  :placeholder="`请输入${field.label}`"
                />
                
                <!-- 多行文本 -->
                <el-input
                  v-else-if="field.type === 'textarea'"
                  v-model="fieldValues[field.name]"
                  type="textarea"
                  :rows="4"
                  :placeholder="`请输入${field.label}`"
                />
                
                <!-- 下拉选择 -->
                <el-select
                  v-else-if="field.type === 'select'"
                  v-model="fieldValues[field.name]"
                  :placeholder="`请选择${field.label}`"
                  class="full-width"
                >
                  <el-option
                    v-for="option in field.options"
                    :key="option"
                    :label="option"
                    :value="option"
                  />
                </el-select>
              </div>
            </div>
          </div>

          <!-- 补充信息 -->
          <div class="selection-section">
            <div class="section-header">
              <el-icon><EditPen /></el-icon>
              <span>补充信息</span>
              <span class="section-meta">{{ additionalInfoLength }}</span>
            </div>
            <el-input
              v-model="additionalInfo"
              type="textarea"
              :rows="4"
              maxlength="500"
              show-word-limit
              placeholder="输入本次生成的额外要求、风格偏好或限制条件..."
              resize="none"
            />
          </div>

          <!-- 关联作品 -->
          <div class="selection-section">
            <div class="section-header">
              <el-icon><Document /></el-icon>
              <span>关联作品（可选）</span>
            </div>
            <el-select
              v-model="selectedBookId"
              clearable
              filterable
              placeholder="直接读取作品名称"
              class="full-width"
            >
              <el-option
                v-for="book in books"
                :key="book.id"
                :label="book.title"
                :value="book.id"
              />
            </el-select>
          </div>

          <!-- 关联章节 -->
          <div class="selection-section">
            <div class="section-header">
              <el-icon><Document /></el-icon>
              <span>关联章节（可选）</span>
            </div>
            <el-select
              v-model="selectedChapterId"
              clearable
              filterable
              :loading="loadingBookContext"
              :disabled="!selectedBookId"
              placeholder="先选择作品，再选择作品内章节"
              class="full-width"
            >
              <el-option
                v-for="chapter in availableChapters"
                :key="chapter.id"
                :label="chapter.title"
                :value="chapter.id"
              />
            </el-select>
            <div class="context-tip">
              可以选择关联章节以提供上下文参考，不选择也可以直接生成内容。
            </div>
          </div>

          <!-- 关联备忘录 -->
          <div class="selection-section">
            <div class="section-header">
              <el-icon><Document /></el-icon>
              <span>关联备忘录（可选）</span>
            </div>
            <el-select
              v-model="selectedMemoId"
              clearable
              filterable
              placeholder="选择备忘录内容作为参考"
              class="full-width"
            >
              <el-option
                v-for="memo in memos"
                :key="memo.id"
                :label="memo.title"
                :value="memo.id"
              />
            </el-select>
            <div class="context-tip">
              可以选择关联备忘录中的内容以提供上下文参考，不选择也可以直接生成内容。
            </div>
          </div>

          <!-- 关联角色卡 -->
          <div class="selection-section">
            <div class="section-header">
              <el-icon><User /></el-icon>
              <span>关联角色卡（可选）</span>
            </div>
            <el-select
              v-model="selectedCharacterIds"
              multiple
              collapse-tags
              collapse-tags-tooltip
              clearable
              filterable
              :loading="loadingBookContext"
              :disabled="!selectedBookId"
              placeholder="先选择作品，再选择角色卡"
              class="full-width"
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

          <div v-if="hasRelatedContext" class="selection-section">
            <div class="section-header">
              <el-icon><View /></el-icon>
              <span>已关联内容</span>
            </div>
            <div class="context-preview-list">
              <div
                v-for="item in relatedContextItems"
                :key="`${item.type}-${item.id}`"
                class="context-preview-item"
              >
                <el-tag size="small" type="info">{{ getRelatedTypeLabel(item.type) }}</el-tag>
                <span>{{ item.title }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button
            v-if="!generating"
            type="primary"
            size="large"
            :disabled="!canGenerate"
            @click="generateContent"
          >
            <el-icon><Lightning /></el-icon>
            <span>开始生成</span>
          </el-button>
          <el-button
            v-else
            type="danger"
            size="large"
            @click="stopGenerate"
          >
            <el-icon><Close /></el-icon>
            <span>停止生成</span>
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 提示词选择弹窗 -->
    <el-dialog
      v-model="promptSelectDialogVisible"
      title="选择提示词"
      width="600px"
      destroy-on-close
      class="prompt-select-dialog"
      modal-class="prompt-select-dialog-modal"
      center
    >
      <div class="prompt-select-content">
        <el-input
          v-model="promptSearchKeyword"
          placeholder="搜索提示词..."
          prefix-icon="Search"
          clearable
          class="search-input"
        />
        <div class="category-tabs" ref="categoryTabsRef">
          <div
            class="category-tabs-wrapper"
            @mousedown="startDrag"
            @mousemove="onDrag"
            @mouseleave="endDrag"
            @mouseup="endDrag"
          >
            <div
              v-for="category in promptCategories"
              :key="category"
              class="category-tab"
              :class="{ active: selectedPromptCategory === category || (category === '全部' && selectedPromptCategory === 'all') }"
              @click="selectedPromptCategory = category === '全部' ? 'all' : category"
            >
              {{ category }}
            </div>
          </div>
        </div>
        <div class="prompt-select-list">
          <div
            v-for="prompt in filteredPrompts"
            :key="prompt.id"
            class="prompt-select-item"
            :class="{ selected: isPromptPinned(prompt.id) }"
            @click="togglePromptPin(prompt)"
          >
            <el-checkbox
              :model-value="isPromptPinned(prompt.id)"
              @click.stop
              @change="togglePromptPin(prompt)"
            />
            <div class="prompt-select-info">
              <div class="prompt-select-name">{{ prompt.name }}</div>
              <div class="prompt-select-desc">{{ prompt.content?.slice(0, 80) }}{{ prompt.content?.length > 80 ? '...' : '' }}</div>
            </div>
            <el-tag size="small" type="info">{{ prompt.category }}</el-tag>
          </div>
          <div v-if="filteredPrompts.length === 0" class="empty-prompts">
            <el-icon :size="32"><Document /></el-icon>
            <p>暂无匹配的提示词</p>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="promptSelectDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 预览弹窗 -->
    <el-dialog
      v-model="previewDialogVisible"
      title="生成结果预览"
      width="50vw"
      destroy-on-close
      class="preview-dialog"
      modal-class="preview-dialog-modal"
    >
      <div class="preview-dialog-content">
        <div class="preview-dialog-header">
          <el-button
            v-if="generating"
            type="danger"
            link
            @click="stopGenerate"
          >
            <el-icon><Close /></el-icon>
            停止
          </el-button>
          <el-button
            v-if="!generating && !isEditingPreview"
            type="primary"
            link
            @click="isEditingPreview = true"
          >
            <el-icon><EditPen /></el-icon>
            编辑
          </el-button>
          <el-button
            v-if="isEditingPreview"
            type="success"
            link
            @click="isEditingPreview = false"
          >
            <el-icon><Check /></el-icon>
            完成
          </el-button>
          <el-button type="primary" link @click="copyPreviewContent">
            <el-icon><CopyDocument /></el-icon>
            复制
          </el-button>
          <el-button
            v-if="!generating && previewContent"
            type="success"
            link
            @click="importToMemo"
          >
            <el-icon><Document /></el-icon>
            一键导入
          </el-button>
        </div>
        <el-scrollbar class="preview-scrollbar">
          <el-input
            v-if="isEditingPreview"
            v-model="previewContent"
            type="textarea"
            :autosize="{ minRows: 20 }"
            placeholder="编辑内容..."
          />
          <MarkdownRenderer v-else :content="previewContent" />
        </el-scrollbar>
      </div>
    </el-dialog>

    <!-- 系统提示词修改弹窗 -->
    <el-dialog
      v-model="promptEditDialogVisible"
      title="修改系统提示词"
      width="600px"
      destroy-on-close
      modal-class="prompt-edit-dialog-modal"
    >
      <div class="prompt-edit-content">
        <div class="form-item">
          <label class="form-label">提示词名称</label>
          <el-input
            v-model="editingPrompt.name"
            placeholder="请输入提示词名称"
            class="full-width"
          />
        </div>
        <div class="form-item">
          <label class="form-label">提示词内容</label>
          <div class="prompt-format-hint">
            <el-alert
              title="提示词格式说明"
              type="info"
              :closable="false"
              show-icon
            >
              <template #default>
                <div class="format-description">
                  <p><strong>当前使用模板式格式（方案3）：</strong></p>
                  <p>角色：[生成器名称]</p>
                  <p>[此处为您编辑的内容]</p>
                  <p>具体要求：[选择的提示词内容]</p>
                  <p class="format-tip">💡 提示：请在此输入核心角色设定和基础能力描述</p>
                </div>
              </template>
            </el-alert>
            
            <el-alert
              title="模板功能说明"
              type="warning"
              :closable="false"
              show-icon
              style="margin-top: 12px;"
            >
              <template #default>
                <div class="format-description">
                  <p><strong>📝 写模板，留空位</strong></p>
                  <p>在提示词正文中，用 <code>${字段名称}</code> 标记需要用户填写的地方</p>
                  <p class="format-tip">示例：生成书名提示词</p>
                  <p class="format-example">帮我生成${数量}个${小说类型}类型${平台}平台的小说书名</p>
                  <p class="format-tip">📋 要求：</p>
                  <ul style="margin: 8px 0; padding-left: 20px;">
                    <li>合理放置字段位置</li>
                    <li>长文本字段建议独立成段</li>
                    <li>字段名称要清晰明确</li>
                  </ul>
                </div>
              </template>
            </el-alert>
          </div>
          <el-input
            v-model="editingPrompt.content"
            type="textarea"
            :rows="8"
            placeholder="请输入系统提示词内容...&#10;例如：你是一个专业的角色生成器助手，擅长创建丰富立体的虚构角色。你会根据用户的需求，生成包含性格、背景、外貌等详细信息的角色设定。"
            resize="none"
            class="full-width"
          />
        </div>
        <div class="form-item">
          <label class="form-label">提示词分类</label>
          <el-input
            v-model="editingPrompt.category"
            placeholder="请输入提示词分类"
            class="full-width"
          />
        </div>
        
        <!-- 字段配置 -->
        <div class="form-item">
          <label class="form-label">字段配置</label>
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
              <el-button type="primary" @click="addField">添加字段</el-button>
              <p class="empty-hint">暂无字段配置，点击添加字段开始配置</p>
            </div>
            
            <div v-else class="fields-list">
              <div v-for="(field, index) in fieldsConfig" :key="index" class="field-item">
                <div class="field-header">
                  <span class="field-name">${{ field.name }}</span>
                  <div class="field-actions">
                    <el-button 
                      type="primary" 
                      link 
                      size="small" 
                      @click="moveField(index, index - 1)"
                      :disabled="index === 0"
                    >
                      上移
                    </el-button>
                    <el-button 
                      type="primary" 
                      link 
                      size="small" 
                      @click="moveField(index, index + 1)"
                      :disabled="index === fieldsConfig.length - 1"
                    >
                      下移
                    </el-button>
                    <el-button type="primary" link size="small" @click="addField">添加字段</el-button>
                    <el-button type="danger" link size="small" @click="removeField(index)">删除</el-button>
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
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="promptEditDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="savePrompt">保存修改</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 固定提示词查看弹窗 -->
    <el-dialog
      v-model="fixedPromptDialogVisible"
      title="发送给AI的完整提示词"
      width="600px"
      destroy-on-close
      modal-class="fixed-prompt-dialog-modal"
    >
      <div class="fixed-prompt-dialog-content">
        <el-input
          v-model="fixedPromptContent"
          type="textarea"
          :rows="12"
          readonly
          placeholder="在此编辑完整提示词..."
          class="full-width"
        />
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="fixedPromptDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="copyFixedPrompt">复制</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 生成器管理弹窗 -->
    <el-dialog
      v-model="generatorManagerVisible"
      title="生成器管理"
      width="900px"
      destroy-on-close
      class="generator-manager-dialog"
      modal-class="generator-manager-dialog-modal"
    >
      <div class="generator-manager-content">
        <!-- 左侧：生成器列表 -->
        <div class="generator-list-panel">
          <div class="panel-header">
            <span class="panel-title">生成器列表</span>
            <el-button type="primary" size="small" @click="createNewGenerator">
              <el-icon><Plus /></el-icon>
              新建
            </el-button>
          </div>
          <div class="generator-list">
            <div
              v-for="gen in allGenerators"
              :key="gen.id"
              class="generator-item"
              :class="{ active: selectedGenerator?.id === gen.id }"
              @click="selectGenerator(gen)"
            >
              <div class="generator-item-icon">
                <el-icon><Lightning /></el-icon>
              </div>
              <div class="generator-item-info">
                <span class="generator-item-name">
                  {{ gen.name }}
                  <el-tag v-if="gen.isDefault" size="small" type="info" style="margin-left: 4px;">默认</el-tag>
                </span>
                <span class="generator-item-desc">{{ gen.description?.slice(0, 30) }}{{ gen.description?.length > 30 ? '...' : '' }}</span>
              </div>
              <div class="generator-item-actions" @click.stop>
                <el-button type="danger" link size="small" @click="deleteGenerator(gen)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>
            <div v-if="allGenerators.length === 0" class="empty-list">
              <el-icon :size="32"><Document /></el-icon>
              <p>暂无生成器</p>
            </div>
          </div>
        </div>

        <!-- 右侧：生成器详情 -->
        <div class="generator-detail-panel">
          <div v-if="selectedGenerator" class="detail-content">
            <div class="detail-section">
              <label class="detail-label">生成器名称</label>
              <el-input v-model="selectedGenerator.name" placeholder="请输入生成器名称" />
            </div>
            <div class="detail-section">
              <label class="detail-label">功能描述</label>
              <el-input
                v-model="selectedGenerator.description"
                type="textarea"
                :rows="3"
                placeholder="请输入功能描述"
              />
            </div>
            <div class="detail-section">
              <label class="detail-label">核心提示词</label>
              <el-input
                v-model="selectedGenerator.core_prompt"
                type="textarea"
                :rows="8"
                placeholder="请输入核心提示词，作为AI的system层指令..."
              />
            </div>
            <div class="detail-section">
              <label class="detail-label">备注</label>
              <el-input
                v-model="selectedGenerator.remark"
                type="textarea"
                :rows="2"
                placeholder="备注信息（可选）"
              />
            </div>
            <div class="detail-actions">
              <el-button @click="cancelEditGenerator">取消</el-button>
              <el-button type="primary" @click="saveGenerator">保存</el-button>
            </div>
          </div>
          <div v-else class="empty-detail">
            <el-icon :size="48"><Document /></el-icon>
            <p>选择或创建一个生成器</p>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 历史记录对话框 -->
    <el-dialog
      v-model="showHistoryDialog"
      title="创意区历史记录"
      width="900px"
      destroy-on-close
      class="history-dialog"
      modal-class="history-dialog-modal"
    >
      <div class="history-dialog-content">
        <div class="history-header">
          <div class="header-left">
            <el-icon class="header-icon"><Document /></el-icon>
            <span class="history-count">共 {{ historyRecords.length }} 条记录</span>
          </div>
          <div class="history-actions">
            <el-button 
              type="danger" 
              size="small" 
              @click="clearHistory"
              :disabled="historyRecords.length === 0"
              class="clear-all-btn"
            >
              <el-icon><Delete /></el-icon>
              清空历史
            </el-button>
          </div>
        </div>
        <div class="history-list">
          <div
            v-for="(record, index) in historyRecords"
            :key="record.id"
            class="history-item"
            @click="viewHistoryDetail(record)"
          >
            <div class="history-item-card">
              <div class="history-item-header">
                <div class="generator-info">
                  <div class="generator-icon-wrapper">
                    <el-icon class="generator-icon" :size="28">
                      <component :is="record.generatorIcon || 'Lightning'" />
                    </el-icon>
                  </div>
                  <div class="generator-details">
                    <span class="generator-name">{{ record.generatorName }}</span>
                    <span class="generator-time">{{ formatTimestamp(record.timestamp) }}</span>
                  </div>
                </div>
                <div class="history-item-tags">
                  <el-tag size="small" effect="plain" class="prompt-count-tag">
                    <el-icon><Document /></el-icon>
                    {{ record.selectedPrompts.length }} 个提示词
                  </el-tag>
                </div>
              </div>
              <div class="history-item-actions">
                <el-button 
                  size="small" 
                  type="primary" 
                  @click.stop="viewHistoryDetail(record)"
                  class="view-detail-btn"
                >
                  <el-icon><View /></el-icon>
                  查看详情
                </el-button>
                <el-button 
                  size="small" 
                  type="danger" 
                  @click.stop="deleteHistoryRecord(index)"
                  class="delete-btn"
                >
                  <el-icon><Delete /></el-icon>
                  删除
                </el-button>
              </div>
            </div>
          </div>
          <div v-if="historyRecords.length === 0" class="empty-history">
            <div class="empty-state-icon">
              <el-icon :size="64"><Document /></el-icon>
            </div>
            <p class="empty-title">暂无历史记录</p>
            <p class="empty-hint">开始第一次生成后，历史记录会显示在这里</p>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showHistoryDialog = false" class="cancel-btn">取消</el-button>
          <el-button type="primary" @click="showHistoryDialog = false" class="confirm-btn">关闭</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 历史记录详情对话框 -->
    <el-dialog
      v-model="historyDialogVisible"
      :title="selectedHistoryRecord?.generatorName || '对话详情'"
      width="900px"
      class="history-detail-dialog"
      modal-class="history-detail-dialog-modal"
      destroy-on-close
      align-center
    >
      <div class="history-detail-content" v-if="selectedHistoryRecord">
        <div class="conversation-messages">
          <template v-for="(msg, msgIndex) in selectedHistoryRecord.messages" :key="msgIndex">
            <div
              v-if="msg.role === 'user' || msg.role === 'assistant' || msg.role === 'system'"
              :class="['message-bubble', msg.role]"
            >
              <div class="message-sender">
                <el-icon class="sender-icon">
                  <User v-if="msg.role === 'user'" />
                  <ChatDotRound v-if="msg.role === 'assistant'" />
                  <Setting v-if="msg.role === 'system'" />
                </el-icon>
                <span class="sender-name">
                  {{ msg.role === 'user' ? '用户' : msg.role === 'assistant' ? 'AI' : '提示词' }}
                </span>
                <span class="message-time">{{ formatTimestamp(msg.timestamp) }}</span>
              </div>
              <div class="message-content">
                <MarkdownRenderer :content="msg.content" />
              </div>
            </div>
          </template>
        </div>
      </div>
      <template #footer>
        <el-button @click="historyDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowRight,
  Document,
  Cpu,
  EditPen,
  View,
  CopyDocument,
  Setting,
  Check,
  Plus,
  Delete,
  Lightning,
  Close,
  User,
  ChatDotRound
} from '@element-plus/icons-vue'
import { promptAPI, configAPI, generatorAPI, bookAPI, chapterAPI, memoAPI, characterAPI } from '@/api'
import type { Generator, Book, Chapter, Memo, Character, RelatedContent } from '@/types'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

// 默认生成器基础配置（用于初始化）
const DEFAULT_GENERATORS_VERSION = '2026-03-22-v2'

const defaultGeneratorConfigs = [
  {
    id: 1,
    name: '脑洞生成器',
    description: '围绕题材、设定、冲突和反转，快速产出高概念创意点子。',
    icon: 'MagicStick',
    core_prompt: '你是一个专业的脑洞生成器助手，擅长围绕题材、设定、冲突、爽点和反转产出新颖且可落地的小说创意。根据用户输入，直接输出有吸引力的创意方案、卖点和延展方向。不要解释你的思考过程，不要输出多余客套话，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 2,
    name: '书名生成器',
    description: '根据题材、受众和卖点生成有传播力、有网文感的书名。',
    icon: 'Document',
    core_prompt: '你是一个专业的书名生成器助手，擅长根据题材、关键词、目标读者和核心卖点生成有吸引力、易传播、符合网文风格的书名。可提供多个方向，但只输出书名结果及简短标签，不要解释推理过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 3,
    name: '简介生成器',
    description: '生成抓人眼球的作品简介、投放文案和一句话卖点。',
    icon: 'EditPen',
    core_prompt: '你是一个专业的简介生成器助手，擅长根据作品题材、人物关系、冲突主线和核心爽点，生成抓人眼球的小说简介。输出内容要强钩子、强冲突、强代入感，可包含一句话卖点或多版简介，但不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 4,
    name: '大纲生成器',
    description: '生成完整主线大纲，梳理故事结构、阶段目标和高潮节点。',
    icon: 'Reading',
    core_prompt: '你是一个专业的大纲生成器助手，擅长根据题材设定、角色关系和核心冲突生成完整的故事大纲。输出应包含故事主线、阶段推进、关键矛盾、高潮节点和结局方向，结构清晰、节奏明确。不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 5,
    name: '细纲生成器',
    description: '把大纲拆到章节或剧情段落，明确每段目标、冲突和悬念。',
    icon: 'Tickets',
    core_prompt: '你是一个专业的细纲生成器助手，擅长把已有题材或大纲进一步细化为章节级、情节级的执行方案。输出内容应明确每一段剧情的目标、冲突、推进、转折、情绪点和悬念钩子。不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 6,
    name: '黄金开篇生成器',
    description: '设计前三章或开篇剧情，强化钩子、冲突和读者留存。',
    icon: 'Lightning',
    core_prompt: '你是一个专业的黄金开篇生成器助手，擅长设计小说开篇的前三章或首个高能场景。输出要突出第一眼钩子、人物处境、核心冲突、情绪爆点和继续追读的悬念，确保开头强抓人。不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 7,
    name: '金手指生成器',
    description: '生成贴合题材的外挂、系统、能力或特殊资源设定。',
    icon: 'Cpu',
    core_prompt: '你是一个专业的金手指生成器助手，擅长根据题材、主角定位和成长路线设计独特且有升级空间的金手指设定。输出应包含能力机制、使用限制、成长方式、爽点来源和剧情联动性。不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 8,
    name: '名字生成器',
    description: '生成人名、地名、宗门名、势力名等，兼顾风格和辨识度。',
    icon: 'CollectionTag',
    core_prompt: '你是一个专业的名字生成器助手，擅长根据题材、时代感、地域风格和人物气质生成人名、地名、组织名、功法名等。输出名字时要注意辨识度、记忆点和风格统一，不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 9,
    name: '人设生成器',
    description: '生成立体角色设定，覆盖身份、性格、目标、秘密和成长线。',
    icon: 'User',
    core_prompt: '你是一个专业的人设生成器助手，擅长生成立体鲜明、适合剧情推进的角色设定。输出应包含身份背景、外在特征、性格标签、行为逻辑、欲望目标、弱点秘密和成长空间。不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  },
  {
    id: 10,
    name: '世界观生成器',
    description: '构建世界规则、势力格局、修炼体系和社会运行逻辑。',
    icon: 'OfficeBuilding',
    core_prompt: '你是一个专业的世界观生成器助手，擅长构建完整且自洽的小说世界观。输出应包含时代背景、核心规则、力量体系、地理版图、势力结构、社会秩序和潜在矛盾，保证可持续支撑长篇剧情。不要解释过程，请使用纯文本回复，不要使用markdown格式。'
  }
]

// 生成器类型
type GeneratorItem = {
  id: number
  name: string
  description: string
  icon: string
  core_prompt?: string
  remark?: string
  order_num?: number
  isDefault?: boolean
  isCustom?: boolean
  isNew?: boolean
}

// 所有生成器列表（从localStorage加载默认生成器的修改）
const allGenerators = ref<GeneratorItem[]>([])

// 自定义生成器列表（来自数据库）
const customGenerators = ref<Generator[]>([])

// 合并后的生成器列表（用于显示）
const generators = computed(() => {
  return allGenerators.value.map(g => ({
    ...g,
    isCustom: !g.isDefault
  }))
})

// 历史记录相关
const showHistoryDialog = ref(false)
const historyDialogVisible = ref(false)

interface HistoryMessage {
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: number
}

interface HistoryRecord {
  id: number
  generatorName: string
  generatorIcon?: string
  fixedPrompt: string
  selectedPrompts: Array<{ name: string; content: string }>
  inputParams: string
  additionalInfo: string
  fieldValues: Record<string, string>
  relatedContext: Array<{ type: RelatedContent['type']; title: string }>
  messages: HistoryMessage[]
  previewContent: string
  timestamp: number
}

const historyRecords = ref<HistoryRecord[]>([])
const selectedHistoryRecord = ref<HistoryRecord | null>(null)

type CreativeModelOption = {
  id: number
  name: string
  model: string
  provider_name?: string
  is_default?: number | boolean
}

// 模型列表
const models = ref<CreativeModelOption[]>([])
const books = ref<Book[]>([])
const memos = ref<Memo[]>([])
const availableChapters = ref<Chapter[]>([])
const availableCharacters = ref<Character[]>([])
const loadingBookContext = ref(false)

const getDefaultModelId = () => {
  const defaultModel = models.value.find(model => model.is_default === 1 || model.is_default === true)
  return defaultModel?.id ?? models.value[0]?.id ?? ''
}

// 加载模型列表
const fetchModels = async () => {
  try {
    const res = await configAPI.getAll()
    if (res.success && res.data) {
      models.value = res.data
      selectedModel.value = getDefaultModelId()
    }
  } catch (error) {
    // 使用默认数据
    models.value = [
      { id: 1, name: 'GPT-4', model: 'gpt-4' },
      { id: 2, name: 'GPT-3.5', model: 'gpt-3.5-turbo' },
      { id: 3, name: 'Claude', model: 'claude-3-opus-20240229' },
      { id: 4, name: 'Gemini', model: 'gemini-pro' }
    ]
    selectedModel.value = getDefaultModelId()
  }
}

// 状态
const dialogVisible = ref(false)
const currentGenerator = ref<GeneratorItem | null>(null)
const selectedPrompt = ref<number>()
const selectedModel = ref<number | string>('')
const inputParams = ref('')
const additionalInfo = ref('')
const selectedBookId = ref<number | null>(null)
const selectedChapterId = ref<number | null>(null)
const selectedMemoId = ref<number | null>(null)
const selectedCharacterIds = ref<number[]>([])
const generating = ref(false)
const generatedContent = ref('')
const abortController = ref<AbortController | null>(null)
const isEditingPreview = ref(false)
const prompts = ref<Array<{ id: number; name: string; category: string; content: string; fields?: Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; description: string; required: boolean }> }>>([])

// 固定提示词相关状态
const pinnedPrompts = ref<Array<{ id: number; name: string; category: string; content: string; enabled: boolean; fields?: Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; description: string; required: boolean }> }>>([])
const promptSelectDialogVisible = ref(false)
const promptSearchKeyword = ref('')
const selectedPromptCategory = ref('all')
const categoryTabsRef = ref<HTMLElement | null>(null)
const isDragging = ref(false)
const startX = ref(0)
const scrollLeft = ref(0)

const startDrag = (e: MouseEvent) => {
  if (!categoryTabsRef.value) return
  isDragging.value = true
  startX.value = e.pageX - categoryTabsRef.value.offsetLeft
  scrollLeft.value = categoryTabsRef.value.scrollLeft
}

const onDrag = (e: MouseEvent) => {
  if (!isDragging.value || !categoryTabsRef.value) return
  e.preventDefault()
  const x = e.pageX - categoryTabsRef.value.offsetLeft
  const walk = (x - startX.value) * 2
  categoryTabsRef.value.scrollLeft = scrollLeft.value - walk
}

const endDrag = () => {
  isDragging.value = false
}

const isLoadingPinnedPrompts = ref(false)

// 预览弹窗状态
const previewDialogVisible = ref(false)
const previewContent = ref('')

// 字段值存储
const fieldValues = ref<Record<string, string>>({})

// 提示词编辑相关状态
const promptEditDialogVisible = ref(false)
const editingPrompt = ref<{ id: number; name: string; content: string; category: string }>({
  id: 0,
  name: '',
  content: '',
  category: ''
})

// 字段配置相关状态
const fieldsConfig = ref<Array<{
  name: string;
  label: string;
  type: 'text' | 'textarea' | 'select';
  options: string[];
  optionsText: string;
  description: string;
}>>([])

// 固定提示词弹窗状态
const fixedPromptDialogVisible = ref(false)
const fixedPromptContent = ref('')

// 可编辑的固定提示词状态
const editableFixedPrompt = ref('')
const savedFixedPrompt = ref('')
const hasUserEditedFixedPrompt = ref(false)

// 生成器管理相关状态
const generatorManagerVisible = ref(false)
const selectedGenerator = ref<Partial<GeneratorItem> | null>(null)
const isNewGenerator = ref(false)

// 计算属性
const selectedPromptInfo = computed(() => {
  return prompts.value.find(p => p.id === selectedPrompt.value)
})

const allFields = computed(() => {
  const fields: Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; description: string; required: boolean }> = []
  pinnedPrompts.value.filter(p => p.enabled).forEach(prompt => {
    if (prompt.fields) {
      prompt.fields.forEach(field => {
        if (!fields.find(f => f.name === field.name)) {
          fields.push(field)
        }
      })
    }
  })
  return fields
})

const hasAnyFields = computed(() => {
  return allFields.value.length > 0
})

const promptCategories = computed(() => {
  const categories = new Set<string>()
  prompts.value.forEach(p => {
    if (p.category) {
      categories.add(p.category)
    }
  })
  return ['全部', ...Array.from(categories)]
})

const filteredPrompts = computed(() => {
  let result = prompts.value
  
  if (selectedPromptCategory.value !== 'all' && selectedPromptCategory.value !== '全部') {
    result = result.filter(p => p.category === selectedPromptCategory.value)
  }
  
  if (promptSearchKeyword.value) {
    const keyword = promptSearchKeyword.value.toLowerCase()
    result = result.filter(p => 
      p.name.toLowerCase().includes(keyword) || 
      p.category.toLowerCase().includes(keyword) ||
      p.content?.toLowerCase().includes(keyword)
    )
  }
  
  return result
})

const additionalInfoLength = computed(() => additionalInfo.value.length)

const selectedBook = computed(() => {
  return books.value.find(book => book.id === selectedBookId.value) || null
})

const selectedChapter = computed(() => {
  return availableChapters.value.find(chapter => chapter.id === selectedChapterId.value) || null
})

const selectedMemo = computed(() => {
  return memos.value.find(memo => memo.id === selectedMemoId.value) || null
})

const selectedCharacters = computed(() => {
  return availableCharacters.value.filter(character => selectedCharacterIds.value.includes(character.id))
})

const relatedContextItems = computed<RelatedContent[]>(() => {
  const items: RelatedContent[] = []

  if (selectedBook.value) {
    const bookLines = [
      `作品名称：${selectedBook.value.title}`,
      selectedBook.value.author ? `作者：${selectedBook.value.author}` : '',
      selectedBook.value.category ? `分类：${selectedBook.value.category}` : '',
      selectedBook.value.tags?.length ? `标签：${selectedBook.value.tags.join('、')}` : '',
      selectedBook.value.description ? `简介：${selectedBook.value.description}` : ''
    ].filter(Boolean)

    items.push({
      type: 'book',
      id: selectedBook.value.id,
      title: selectedBook.value.title,
      content: bookLines.join('\n')
    })
  }

  if (selectedChapter.value) {
    items.push({
      type: 'chapter',
      id: selectedChapter.value.id,
      title: selectedChapter.value.title,
      content: selectedChapter.value.content
    })
  }

  if (selectedMemo.value) {
    items.push({
      type: 'memo',
      id: selectedMemo.value.id,
      title: selectedMemo.value.title,
      content: selectedMemo.value.content
    })
  }

  selectedCharacters.value.forEach(character => {
    const characterLines = [
      character.gender ? `性别：${character.gender}` : '',
      character.folder ? `分组：${character.folder}` : '',
      character.personality ? `性格：${character.personality}` : '',
      character.info ? `设定：${character.info}` : ''
    ].filter(Boolean)

    items.push({
      type: 'character',
      id: character.id,
      title: character.name,
      content: characterLines.join('\n')
    })
  })

  return items
})

const hasRelatedContext = computed(() => relatedContextItems.value.length > 0)

const canGenerate = computed(() => {
  if (!selectedModel.value) return false

  const hasPromptFoundation = Boolean(editableFixedPrompt.value.trim()) || pinnedPrompts.value.some(p => p.enabled)
  if (!hasPromptFoundation) return false

  const hasAdditionalContent = additionalInfo.value.trim() !== '' || hasRelatedContext.value

  if (allFields.value.length > 0) {
    const requiredFieldsCompleted = allFields.value.every(field => {
      if (field.required === false) return true
      const value = fieldValues.value[field.name]
      return value && value.trim() !== ''
    })

    if (!requiredFieldsCompleted) return false

    const hasFieldContent = Object.values(fieldValues.value).some(value => value?.trim())
    return hasFieldContent || hasAdditionalContent
  }

  return inputParams.value.trim() !== '' || hasAdditionalContent
})

// 检查固定提示词是否已修改
const hasFixedPromptChanged = computed(() => {
  return editableFixedPrompt.value !== savedFixedPrompt.value && editableFixedPrompt.value !== ''
})

// 方法
const isPromptPinned = (promptId: number) => {
  return pinnedPrompts.value.some(p => p.id === promptId)
}

const togglePromptPin = (prompt: { id: number; name: string; category: string; content: string; fields?: Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; description: string; required: boolean }> }) => {
  const index = pinnedPrompts.value.findIndex(p => p.id === prompt.id)
  if (index > -1) {
    pinnedPrompts.value.splice(index, 1)
  } else {
    pinnedPrompts.value.push({ ...prompt, enabled: false })
  }
}

const removePinnedPrompt = (promptId: number) => {
  const index = pinnedPrompts.value.findIndex(p => p.id === promptId)
  if (index > -1) {
    pinnedPrompts.value.splice(index, 1)
  }
}

const copyPreviewContent = () => {
  navigator.clipboard.writeText(previewContent.value)
  ElMessage.success('已复制到剪贴板')
}

const importToMemo = async () => {
  if (!previewContent.value) {
    ElMessage.warning('没有可导入的内容')
    return
  }
  
  try {
    await ElMessageBox.prompt('请输入备忘录标题', '导入到备忘录', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPlaceholder: '例如：创意灵感记录',
      inputValidator: (value) => {
        if (!value || value.trim() === '') {
          return '标题不能为空'
        }
        return true
      }
    }).then(async ({ value }) => {
      const title = value.trim()
      const memoData: Partial<Memo> = {
        title: title,
        content: previewContent.value,
        category: '创意灵感',
        order_num: 0
      }
      
      const res = await memoAPI.create(memoData)
      if (res.success) {
        ElMessage.success('已成功导入到备忘录')
        // 刷新备忘录列表
        await fetchMemos()
      } else {
        ElMessage.error('导入失败，请稍后重试')
      }
    })
  } catch (error) {
    // 用户取消操作
  }
}

const getPinnedPromptsKey = (generatorId: number) => {
  return `generator_pinned_prompts_${generatorId}`
}

const savePinnedPromptsToStorage = () => {
  if (currentGenerator.value) {
    const key = getPinnedPromptsKey(currentGenerator.value.id)
    localStorage.setItem(key, JSON.stringify(pinnedPrompts.value))
  }
}

const loadPinnedPromptsFromStorage = (generatorId: number) => {
  const key = getPinnedPromptsKey(generatorId)
  const saved = localStorage.getItem(key)
  if (saved) {
    try {
      return JSON.parse(saved)
    } catch {
      return []
    }
  }
  return []
}

const getFormStateKey = (generatorId: number) => {
  return `generator_form_state_${generatorId}`
}

const saveFormStateToStorage = () => {
  if (currentGenerator.value) {
    const key = getFormStateKey(currentGenerator.value.id)
    const formState = {
      selectedModel: selectedModel.value,
      inputParams: inputParams.value,
      additionalInfo: additionalInfo.value,
      selectedBookId: selectedBookId.value,
      selectedChapterId: selectedChapterId.value,
      selectedMemoId: selectedMemoId.value,
      selectedCharacterIds: selectedCharacterIds.value,
      fieldValues: fieldValues.value,
      editableFixedPrompt: editableFixedPrompt.value,
      savedFixedPrompt: savedFixedPrompt.value,
      hasUserEditedFixedPrompt: hasUserEditedFixedPrompt.value
    }
    localStorage.setItem(key, JSON.stringify(formState))
  }
}

const loadFormStateFromStorage = (generatorId: number) => {
  const key = getFormStateKey(generatorId)
  const saved = localStorage.getItem(key)
  if (saved) {
    try {
      return JSON.parse(saved)
    } catch {
      return null
    }
  }
  return null
}

const syncGeneratorPromptState = (generator: Pick<GeneratorItem, 'id' | 'core_prompt'>) => {
  const latestPrompt = generator.core_prompt || ''
  const savedFormState = loadFormStateFromStorage(generator.id)

  if (savedFormState) {
    const key = getFormStateKey(generator.id)
    localStorage.setItem(key, JSON.stringify({
      ...savedFormState,
      editableFixedPrompt: latestPrompt,
      savedFixedPrompt: latestPrompt,
      hasUserEditedFixedPrompt: Boolean(latestPrompt)
    }))
  }

  if (currentGenerator.value?.id === generator.id) {
    currentGenerator.value = {
      ...currentGenerator.value,
      core_prompt: latestPrompt
    }
    editableFixedPrompt.value = latestPrompt
    savedFixedPrompt.value = latestPrompt
    hasUserEditedFixedPrompt.value = Boolean(latestPrompt)
    saveFormStateToStorage()
  }
}

const buildSelectedPromptContent = () => {
  let selectedContent = ''

  pinnedPrompts.value
    .filter(prompt => prompt.enabled)
    .forEach(prompt => {
      let content = prompt.content || ''
      if (prompt.fields) {
        prompt.fields.forEach(field => {
          const fieldValue = fieldValues.value[field.name] || ''
          content = content.replace(new RegExp(`\\$\\{${field.name}\\}`, 'g'), fieldValue)
        })
      }
      selectedContent += `${content}\n`
    })

  return selectedContent.trim()
}

const buildSystemPrompt = () => {
  const fixedContent = editableFixedPrompt.value.trim()
  const selectedContent = buildSelectedPromptContent()

  return [fixedContent, selectedContent ? `\n${selectedContent}` : '']
    .filter(Boolean)
    .join('\n\n')
}

const buildUserPrompt = () => {
  const generatorName = currentGenerator.value?.name || '创意生成'
  const sections: string[] = [``]

  if (additionalInfo.value.trim()) {
    sections.push(`补充信息：\n${additionalInfo.value.trim()}`)
  }

  if (selectedBook.value) {
    sections.push(`关联作品：${selectedBook.value.title}`)
  }

  return sections.join('\n\n').trim()
}

const buildFullPromptPreview = () => {
  const sections = [`${buildSystemPrompt() || ''}`]
  const userPrompt = buildUserPrompt()

  if (userPrompt) {
    sections.push(`${userPrompt}`)
  }

  if (relatedContextItems.value.length > 0) {
    const relatedText = relatedContextItems.value
      .map(item => `[${item.type}] ${item.title}\n${item.content}`)
      .join('\n\n')
    sections.push(`【关联参考内容】\n${relatedText}`)
  }

  return sections.join('\n\n')
}

const fetchBooks = async () => {
  try {
    const res = await bookAPI.getAll()
    books.value = res.success && res.data ? res.data : []
  } catch (error) {
    books.value = []
  }
}

const fetchMemos = async () => {
  try {
    const res = await memoAPI.getAll()
    memos.value = res.success && res.data ? res.data : []
  } catch (error) {
    memos.value = []
  }
}

const fetchBookContext = async (bookId: number) => {
  loadingBookContext.value = true

  try {
    const [chapterRes, characterRes] = await Promise.all([
      chapterAPI.getByBook(bookId),
      characterAPI.getByBook(bookId)
    ])

    availableChapters.value = chapterRes.success && chapterRes.data
      ? chapterRes.data.filter(chapter => chapter.type === 'chapter')
      : []
    availableCharacters.value = characterRes.success && characterRes.data ? characterRes.data : []
  } catch (error) {
    availableChapters.value = []
    availableCharacters.value = []
    ElMessage.warning('关联作品上下文加载失败，请稍后重试')
  } finally {
    loadingBookContext.value = false
  }
}

const resetRelatedSelections = () => {
  selectedBookId.value = null
  selectedChapterId.value = null
  selectedMemoId.value = null
  selectedCharacterIds.value = []
  availableChapters.value = []
  availableCharacters.value = []
}

const getRelatedTypeLabel = (type: RelatedContent['type']) => {
  switch (type) {
    case 'book':
      return '作品'
    case 'chapter':
      return '章节'
    case 'memo':
      return '备忘录'
    case 'character':
      return '角色卡'
    default:
      return type
  }
}

const openGenerator = (generator: GeneratorItem) => {
  currentGenerator.value = generator
  selectedPrompt.value = undefined
  generatedContent.value = ''
  promptSearchKeyword.value = ''
  selectedPromptCategory.value = 'all'
  resetRelatedSelections()

  isLoadingPinnedPrompts.value = true
  pinnedPrompts.value = loadPinnedPromptsFromStorage(generator.id)

  const savedFormState = loadFormStateFromStorage(generator.id)
  if (savedFormState) {
    selectedModel.value = savedFormState.selectedModel || getDefaultModelId()
    inputParams.value = savedFormState.inputParams || ''
    additionalInfo.value = savedFormState.additionalInfo || ''
    selectedBookId.value = savedFormState.selectedBookId
    selectedChapterId.value = savedFormState.selectedChapterId
    selectedMemoId.value = savedFormState.selectedMemoId
    selectedCharacterIds.value = savedFormState.selectedCharacterIds || []
    fieldValues.value = savedFormState.fieldValues || {}
    editableFixedPrompt.value = savedFormState.editableFixedPrompt || ''
    savedFixedPrompt.value = savedFormState.savedFixedPrompt || ''
    hasUserEditedFixedPrompt.value = savedFormState.hasUserEditedFixedPrompt || false

    if (selectedBookId.value) {
      fetchBookContext(selectedBookId.value)
    }
  } else {
    selectedModel.value = getDefaultModelId()
    inputParams.value = ''
    additionalInfo.value = ''
    fieldValues.value = {}

    if (generator.core_prompt) {
      editableFixedPrompt.value = generator.core_prompt
      savedFixedPrompt.value = generator.core_prompt
      hasUserEditedFixedPrompt.value = true
    } else {
      const defaultPrompt = `你是一个专业的${generator.name}助手，请根据用户的输入生成相关内容。`
      editableFixedPrompt.value = defaultPrompt
      savedFixedPrompt.value = defaultPrompt
      hasUserEditedFixedPrompt.value = false
    }
  }

  setTimeout(() => {
    isLoadingPinnedPrompts.value = false
    dialogVisible.value = true
  }, 0)
}

// 打开提示词编辑弹窗
const openPromptEditDialog = () => {
  const prompt = selectedPromptInfo.value
  if (prompt) {
    editingPrompt.value = {
      id: prompt.id,
      name: prompt.name,
      content: prompt.content,
      category: prompt.category
    }
    // 重置字段配置
    fieldsConfig.value = []
    // 从提示词内容中提取字段名称
    extractFieldsFromContent(prompt.content)
    promptEditDialogVisible.value = true
  }
}

// 监听提示词内容变化，自动更新字段配置
watch(() => editingPrompt.value.content, (newContent) => {
  if (newContent) {
    const fieldRegex = /\$\{([^}]+)\}/g
    const newFields = new Set<string>()
    let match
    while ((match = fieldRegex.exec(newContent)) !== null) {
      if (match && match[1]) {
        newFields.add(match[1].trim())
      }
    }
    
    const existingFieldNames = new Set(fieldsConfig.value.map(field => field.name))
    
    newFields.forEach(fieldName => {
      if (!existingFieldNames.has(fieldName)) {
        fieldsConfig.value.push({
          name: fieldName,
          label: fieldName,
          type: 'text',
          options: [],
          optionsText: '',
          description: ''
        })
      }
    })
    
    fieldsConfig.value = fieldsConfig.value.filter(field => newFields.has(field.name))
  } else {
    fieldsConfig.value = []
  }
})

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
      description: ''
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
    description: ''
  })
}

// 删除字段
const removeField = (index: number) => {
  fieldsConfig.value.splice(index, 1)
}

// 更新字段选项
const updateFieldOptions = (index: number) => {
  const field = fieldsConfig.value[index]
  if (field) {
    const optionsText = field.optionsText
    field.options = optionsText.split('\n').filter((option: string) => option.trim())
  }
}

// 显示固定提示词弹窗（方案 3 格式）
const showFixedPrompt = () => {
  fixedPromptContent.value = buildFullPromptPreview()
  fixedPromptDialogVisible.value = true
}

// 保存固定提示词
const saveFixedPrompt = () => {
  savedFixedPrompt.value = editableFixedPrompt.value
  hasUserEditedFixedPrompt.value = true
  ElMessage.success('固定提示词已保存')
}

watch(pinnedPrompts, (newVal, oldVal) => {
  if (!isLoadingPinnedPrompts.value) {
    savePinnedPromptsToStorage()
    if (newVal.length !== oldVal?.length) {
      fieldValues.value = {}
    }
  }
}, { deep: true })

watch(editableFixedPrompt, (newVal, oldVal) => {
  if (newVal && newVal !== oldVal && !hasUserEditedFixedPrompt.value) {
    hasUserEditedFixedPrompt.value = true
  }
})

watch(selectedBookId, async (newBookId, oldBookId) => {
  if (newBookId === oldBookId) return

  selectedChapterId.value = null
  selectedCharacterIds.value = []
  availableChapters.value = []
  availableCharacters.value = []

  if (newBookId) {
    await fetchBookContext(newBookId)
  }
})

watch(() => allFields.value.length, (newLen, oldLen) => {
  if (newLen !== oldLen && !isLoadingPinnedPrompts.value) {
    const newKeys = new Set(allFields.value.map(f => f.name))
    Object.keys(fieldValues.value).forEach(key => {
      if (!newKeys.has(key)) {
        delete fieldValues.value[key]
      }
    })
  }
})

watch(dialogVisible, (newVal) => {
  if (!newVal && currentGenerator.value) {
    saveFormStateToStorage()
  }
})

const generateContent = async () => {
  if (!canGenerate.value) return

  generating.value = true
  previewContent.value = ''
  previewDialogVisible.value = true
  isEditingPreview.value = false
  
  abortController.value = new AbortController()
  
  try {
    const selectedModelData = models.value.find(m => m.id === selectedModel.value)
    
    if (!selectedModelData) {
      ElMessage.error('请选择模型')
      generating.value = false
      return
    }
    
    const generatorName = currentGenerator.value?.name || ''
    const systemPrompt = buildSystemPrompt()
    const userPrompt = buildUserPrompt()
    const relatedContent = relatedContextItems.value
    
    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        messages: [
          { role: 'user', content: userPrompt }
        ],
        configId: selectedModelData.id,
        systemPrompts: systemPrompt ? [systemPrompt] : [],
        relatedContent
      }),
      signal: abortController.value.signal
    })
    
    if (!response.ok) {
      throw new Error('AI调用失败')
    }
    
    const reader = response.body?.getReader()
    const decoder = new TextDecoder()
    let aiResponse = ''
    
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
            
            if (json.error) {
              ElMessage.error(`AI调用失败: ${json.error}`)
              throw new Error(json.error)
            }
            
            if (json.content) {
              aiResponse += json.content
              previewContent.value = aiResponse
              generatedContent.value = aiResponse
            }
          } catch (e) {
            console.error('解析流式数据失败:', e)
          }
        }
      }
    }
    
    if (aiResponse) {
      ElMessage.success('生成成功！')
      
      const historyRecord: HistoryRecord = {
        id: Date.now(),
        generatorName: generatorName,
        generatorIcon: currentGenerator.value?.icon,
        fixedPrompt: editableFixedPrompt.value,
        selectedPrompts: pinnedPrompts.value.filter(p => p.enabled).map(p => ({
          name: p.name,
          content: p.content || ''
        })),
        inputParams: inputParams.value,
        additionalInfo: additionalInfo.value,
        fieldValues: { ...fieldValues.value },
        relatedContext: relatedContent.map(item => ({
          type: item.type,
          title: item.title
        })),
        messages: [
          {
            role: 'system',
            content: systemPrompt,
            timestamp: Date.now()
          },
          {
            role: 'user',
            content: userPrompt,
            timestamp: Date.now()
          },
          {
            role: 'assistant',
            content: aiResponse,
            timestamp: Date.now()
          }
        ],
        previewContent: aiResponse,
        timestamp: Date.now()
      }
      
      saveHistoryRecord(historyRecord)
    } else {
      ElMessage.error('生成内容为空')
    }
  } catch (error: any) {
    if (error.name === 'AbortError') {
      ElMessage.info('已停止生成')
    } else {
      console.error('生成内容失败:', error)
      ElMessage.error('生成失败，请重试')
    }
  } finally {
    generating.value = false
    abortController.value = null
  }
}

const stopGenerate = () => {
  if (abortController.value) {
    abortController.value.abort()
  }
}

const copyContent = () => {
  navigator.clipboard.writeText(generatedContent.value)
  ElMessage.success('已复制到剪贴板')
}

// 复制固定提示词
const copyFixedPrompt = () => {
  navigator.clipboard.writeText(fixedPromptContent.value)
  ElMessage.success('提示词已复制到剪贴板')
}

// 保存提示词修改
const savePrompt = async () => {
  // 验证字段配置
  const validationError = validateFieldsConfig()
  if (validationError) {
    ElMessage.error(validationError)
    return
  }
  
  try {
    const response = await promptAPI.update(editingPrompt.value.id, {
      name: editingPrompt.value.name,
      content: editingPrompt.value.content,
      category: editingPrompt.value.category
    })
    
    if (response.success) {
      // 更新本地提示词列表
      const index = prompts.value.findIndex(p => p.id === editingPrompt.value.id)
      if (index !== -1) {
        prompts.value[index] = {
          id: editingPrompt.value.id,
          name: editingPrompt.value.name,
          category: editingPrompt.value.category,
          content: editingPrompt.value.content
        }
      }
      
      ElMessage.success('提示词修改成功')
      promptEditDialogVisible.value = false
    }
  } catch (error) {
    console.error('保存提示词失败:', error)
    ElMessage.error('保存失败，请重试')
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

// 调整字段顺序
const moveField = (fromIndex: number, toIndex: number) => {
  if (fromIndex < 0 || fromIndex >= fieldsConfig.value.length || toIndex < 0 || toIndex >= fieldsConfig.value.length) return
  
  const [movedField] = fieldsConfig.value.splice(fromIndex, 1)
  if (movedField) {
    fieldsConfig.value.splice(toIndex, 0, movedField)
  }
}

// 生成器管理方法
const openGeneratorManager = () => {
  generatorManagerVisible.value = true
  selectedGenerator.value = null
  fetchGenerators()
}

const fetchGenerators = async () => {
  try {
    const res = await generatorAPI.getAll()
    if (res?.data) {
      customGenerators.value = res.data
    }
  } catch (error) {
    customGenerators.value = []
  }
}

const createNewGenerator = () => {
  selectedGenerator.value = {
    name: '',
    description: '',
    icon: 'Lightning',
    core_prompt: '',
    remark: '',
    order_num: 0,
    isDefault: false
  }
  isNewGenerator.value = true
}

const selectGenerator = (gen: GeneratorItem) => {
  selectedGenerator.value = { ...gen }
  isNewGenerator.value = false
}

const cancelEditGenerator = () => {
  selectedGenerator.value = null
  isNewGenerator.value = false
}

const saveGenerator = async () => {
  if (!selectedGenerator.value?.name?.trim()) {
    ElMessage.warning('请输入生成器名称')
    return
  }
  
  try {
    if (isNewGenerator.value) {
      // 创建新的自定义生成器
      const res = await generatorAPI.create({
        name: selectedGenerator.value.name,
        description: selectedGenerator.value.description,
        icon: selectedGenerator.value.icon || 'Lightning',
        core_prompt: selectedGenerator.value.core_prompt,
        remark: selectedGenerator.value.remark,
        order_num: selectedGenerator.value.order_num || 0
      })
      if (res.success) {
        ElMessage.success('创建成功')
        await fetchGenerators()
        // 更新本地列表
        const defaults = loadDefaultGenerators()
        allGenerators.value = [...defaults, ...customGenerators.value.map(g => ({ ...g, isDefault: false }))]
        selectedGenerator.value = null
        isNewGenerator.value = false
      }
    } else if (selectedGenerator.value.isDefault) {
      // 保存默认生成器的修改到localStorage
      const defaults = loadDefaultGenerators()
      const index = defaults.findIndex((g: GeneratorItem) => g.id === selectedGenerator.value!.id)
      if (index !== -1) {
        defaults[index] = { ...selectedGenerator.value } as GeneratorItem
        saveDefaultGenerators(defaults)
        syncGeneratorPromptState(defaults[index])
        // 更新本地列表
        allGenerators.value = [...defaults, ...customGenerators.value.map(g => ({ ...g, isDefault: false }))]
        ElMessage.success('保存成功')
      }
    } else {
      // 保存自定义生成器到数据库
      const res = await generatorAPI.update(selectedGenerator.value.id!, {
        name: selectedGenerator.value.name,
        description: selectedGenerator.value.description,
        icon: selectedGenerator.value.icon,
        core_prompt: selectedGenerator.value.core_prompt,
        remark: selectedGenerator.value.remark,
        order_num: selectedGenerator.value.order_num
      })
      if (res.success) {
        ElMessage.success('保存成功')
        await fetchGenerators()
        syncGeneratorPromptState({
          id: selectedGenerator.value.id!,
          core_prompt: selectedGenerator.value.core_prompt
        })
        // 更新本地列表
        const defaults = loadDefaultGenerators()
        allGenerators.value = [...defaults, ...customGenerators.value.map(g => ({ ...g, isDefault: false }))]
      }
    }
  } catch (error) {
    ElMessage.error('保存失败，请重试')
  }
}

const deleteGenerator = async (gen: GeneratorItem) => {
  try {
    await ElMessageBox.confirm('确定要删除这个生成器吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    if (gen.isDefault) {
      // 删除默认生成器：从localStorage中移除
      const defaults = loadDefaultGenerators()
      const index = defaults.findIndex((g: GeneratorItem) => g.id === gen.id)
      if (index !== -1) {
        defaults.splice(index, 1)
        saveDefaultGenerators(defaults)
        allGenerators.value = [...defaults, ...customGenerators.value.map(g => ({ ...g, isDefault: false }))]
        ElMessage.success('删除成功')
        if (selectedGenerator.value?.id === gen.id) {
          selectedGenerator.value = null
        }
      }
    } else {
      // 删除自定义生成器：从数据库中删除
      const res = await generatorAPI.delete(gen.id)
      if (res.success) {
        ElMessage.success('删除成功')
        await fetchGenerators()
        const defaults = loadDefaultGenerators()
        allGenerators.value = [...defaults, ...customGenerators.value.map(g => ({ ...g, isDefault: false }))]
        if (selectedGenerator.value?.id === gen.id) {
          selectedGenerator.value = null
        }
      }
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败，请重试')
    }
  }
}

const createDefaultGenerators = (): GeneratorItem[] =>
  defaultGeneratorConfigs.map(g => ({ ...g, isDefault: true }))

// 从localStorage加载默认生成器配置
const loadDefaultGenerators = () => {
  const stored = localStorage.getItem('default_generators')
  const version = localStorage.getItem('default_generators_version')

  if (stored && version === DEFAULT_GENERATORS_VERSION) {
    try {
      return JSON.parse(stored)
    } catch (error) {
      console.error('Failed to parse default generators:', error)
    }
  }

  const defaults = createDefaultGenerators()
  saveDefaultGenerators(defaults)
  return defaults
}

// 保存默认生成器配置到localStorage
const saveDefaultGenerators = (generators: GeneratorItem[]) => {
  localStorage.setItem('default_generators', JSON.stringify(generators))
  localStorage.setItem('default_generators_version', DEFAULT_GENERATORS_VERSION)
}

// 加载提示词列表
onMounted(async () => {
  // 加载默认生成器（从localStorage）
  const defaults = loadDefaultGenerators()
  
  // 加载自定义生成器（从数据库）
  await fetchGenerators()
  
  // 合并所有生成器
  allGenerators.value = [...defaults, ...customGenerators.value.map(g => ({ ...g, isDefault: false }))]
  
  try {
    const res = await promptAPI.getAll()
    if (res?.data) {
      prompts.value = res.data.map((p: { id: number; name: string; category: string; content: string; fields?: any[] }) => ({
        id: p.id,
        name: p.name,
        category: p.category || '通用',
        content: p.content,
        fields: p.fields
      }))
    }
  } catch (error) {
    // 使用默认数据
    prompts.value = [
      { id: 1, name: '角色设定生成', category: '角色', content: '' },
      { id: 2, name: '剧情大纲生成', category: '剧情', content: '' },
      { id: 3, name: '对话润色', category: '对话', content: '' },
      { id: 4, name: '场景描写', category: '场景', content: '' },
      { id: 5, name: '情节转折', category: '剧情', content: '' }
    ]
  }
  
  // 加载模型列表
  await fetchModels()
  await Promise.all([fetchBooks(), fetchMemos()])
  
  // 加载历史记录
  loadHistoryRecords()
})

// 历史记录相关函数
const loadHistoryRecords = () => {
  const stored = localStorage.getItem('creative-history')
  if (stored) {
    try {
      const records = JSON.parse(stored)
      historyRecords.value = records
    } catch (e) {
      console.error('Failed to parse history records:', e)
      historyRecords.value = []
    }
  }
}

const saveHistoryRecord = (record: HistoryRecord) => {
  const allRecords = JSON.parse(localStorage.getItem('creative-history') || '[]')
  allRecords.push(record)
  localStorage.setItem('creative-history', JSON.stringify(allRecords))
  loadHistoryRecords()
}

const deleteHistoryRecord = (index: number) => {
  historyRecords.value.splice(index, 1)
  const allRecords = JSON.parse(localStorage.getItem('creative-history') || '[]')
  allRecords.splice(index, 1)
  localStorage.setItem('creative-history', JSON.stringify(allRecords))
}

const clearHistory = () => {
  historyRecords.value = []
  localStorage.removeItem('creative-history')
  ElMessage.success('历史记录已清空')
}

const formatTimestamp = (timestamp: number) => {
  const date = new Date(timestamp)
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

const viewHistoryDetail = (record: HistoryRecord) => {
  selectedHistoryRecord.value = record
  historyDialogVisible.value = true
}
</script>

<style>
.creative-page {
  padding: 40px;
  background: #f8fafc;
  min-height: 100vh;
  position: relative;
}

.creative-page::before {
  content: '';
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background:
    radial-gradient(ellipse at 20% 50%, rgba(8, 198, 190, 0.06) 0%, transparent 50%),
    radial-gradient(ellipse at 80% 80%, rgba(5, 150, 145, 0.04) 0%, transparent 50%);
  pointer-events: none;
  z-index: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 48px;
  position: relative;
  z-index: 1;
}

.header-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.page-header h1 {
  font-size: 36px;
  font-weight: 800;
  color: #0f172a;
  margin: 0;
  letter-spacing: -1px;
}

.subtitle {
  color: #64748b;
  font-size: 15px;
  margin: 0;
  font-weight: 400;
  letter-spacing: 0.3px;
}

.generators-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 20px;
  width: 100%;
  position: relative;
  z-index: 1;
}

@media (max-width: 1400px) {
  .generators-grid {
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  }
}

@media (max-width: 1100px) {
  .generators-grid {
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  }
}

@media (max-width: 768px) {
  .generators-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .generators-grid {
    grid-template-columns: 1fr;
  }

  .creative-page {
    padding: 24px 16px;
  }
}

.generator-card {
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.03) 0%, rgba(255, 255, 255, 0.01) 100%);
  border-radius: 16px;
  padding: 20px 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  cursor: pointer;
  transition: all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94);
  border: 1px solid #e2e8f0;
  box-shadow:
    0 4px 24px rgba(148, 163, 184, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 1);
  position: relative;
  overflow: hidden;
}

.generator-card .card-texture {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image:
    radial-gradient(circle at 25% 25%, rgba(255, 255, 255, 0.15) 0%, transparent 50%),
    radial-gradient(circle at 75% 75%, rgba(148, 163, 184, 0.08) 0%, transparent 50%);
  opacity: 1;
  pointer-events: none;
  border-radius: 16px;
}

.generator-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(8, 198, 190, 0.4), transparent);
  opacity: 0;
  transition: opacity 0.4s;
}

.generator-card::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border-radius: 20px;
  background: radial-gradient(circle at 30% 50%, rgba(8, 198, 190, 0.06) 0%, transparent 60%);
  opacity: 0;
  transition: opacity 0.4s;
  pointer-events: none;
}

.generator-card:hover::before {
  opacity: 1;
}

.generator-card:hover::after {
  opacity: 1;
}

.generator-card:hover {
  transform: translateY(-6px);
  box-shadow:
    0 20px 48px rgba(148, 163, 184, 0.18),
    0 0 0 1px rgba(8, 198, 190, 0.25),
    inset 0 1px 0 rgba(255, 255, 255, 1);
  border-color: rgba(8, 198, 190, 0.35);
}

.card-content-left {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
  position: relative;
  z-index: 1;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.generator-card .card-title {
  font-size: 15px;
  font-weight: 600;
  color: #1e293b;
  text-align: left;
  line-height: 1.4;
  letter-spacing: -0.2px;
}

.new-tag {
  flex-shrink: 0;
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.6px;
  padding: 2px 6px;
  border-radius: 4px;
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  border: none;
  color: #ffffff;
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.25);
}

.card-desc {
  font-size: 12px;
  color: #64748b;
  margin: 0;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  line-clamp: 1;
  -webkit-box-orient: vertical;
  box-orient: vertical;
  overflow: hidden;
  font-weight: 400;
  letter-spacing: 0.2px;
}

.card-icon-wrapper {
  flex-shrink: 0;
  position: relative;
  z-index: 1;
  width: 100px;
  height: 100px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-circle-decoration {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  position: relative;
  overflow: hidden;
}

/* 脑洞生成器 - MagicStick - 星芒放射 */
.deco-MagicStick {
  background:
    radial-gradient(circle at 50% 50%, rgba(8, 198, 190, 0.12) 0%, transparent 40%),
    conic-gradient(
      from 0deg,
      rgba(8, 198, 190, 0.06) 0deg,
      transparent 15deg,
      rgba(8, 198, 190, 0.04) 30deg,
      transparent 45deg,
      rgba(8, 198, 190, 0.06) 60deg,
      transparent 75deg,
      rgba(8, 198, 190, 0.04) 90deg,
      transparent 105deg,
      rgba(8, 198, 190, 0.06) 120deg,
      transparent 135deg,
      rgba(8, 198, 190, 0.04) 150deg,
      transparent 165deg,
      rgba(8, 198, 190, 0.06) 180deg,
      transparent 195deg,
      rgba(8, 198, 190, 0.04) 210deg,
      transparent 225deg,
      rgba(8, 198, 190, 0.06) 240deg,
      transparent 255deg,
      rgba(8, 198, 190, 0.04) 270deg,
      transparent 285deg,
      rgba(8, 198, 190, 0.06) 300deg,
      transparent 315deg,
      rgba(8, 198, 190, 0.04) 330deg,
      transparent 345deg,
      rgba(8, 198, 190, 0.06) 360deg
    );
}

.deco-MagicStick::after {
  content: '';
  position: absolute;
  top: 30%;
  left: 30%;
  right: 30%;
  bottom: 30%;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(8, 198, 190, 0.15) 0%, transparent 70%);
}

/* 书名生成器 - Document - 书页纹理 */
.deco-Document {
  background:
    linear-gradient(180deg,
      rgba(8, 198, 190, 0.03) 0px,
      rgba(8, 198, 190, 0.03) 1px,
      transparent 1px,
      transparent 10px
    ),
    linear-gradient(90deg,
      rgba(8, 198, 190, 0.02) 0px,
      rgba(8, 198, 190, 0.02) 1px,
      transparent 1px,
      transparent 14px
    ),
    radial-gradient(circle at 30% 30%, rgba(8, 198, 190, 0.08) 0%, transparent 50%);
}

.deco-Document::after {
  content: '';
  position: absolute;
  top: 20%;
  left: 15%;
  right: 15%;
  bottom: 20%;
  border-left: 2px solid rgba(8, 198, 190, 0.08);
  border-radius: 0;
}

/* 简介生成器 - EditPen - 笔触纹理 */
.deco-EditPen {
  background:
    radial-gradient(circle at 25% 25%, rgba(8, 198, 190, 0.1) 0%, transparent 30%),
    radial-gradient(circle at 75% 60%, rgba(8, 198, 190, 0.08) 0%, transparent 25%),
    radial-gradient(circle at 50% 80%, rgba(8, 198, 190, 0.06) 0%, transparent 20%),
    linear-gradient(135deg, rgba(8, 198, 190, 0.04) 0%, transparent 50%);
}

.deco-EditPen::after {
  content: '';
  position: absolute;
  top: 25%;
  left: 20%;
  width: 60%;
  height: 2px;
  background: linear-gradient(90deg, rgba(8, 198, 190, 0.15), rgba(8, 198, 190, 0.05));
  border-radius: 1px;
  box-shadow:
    0 8px 0 rgba(8, 198, 190, 0.1),
    0 16px 0 rgba(8, 198, 190, 0.07),
    0 24px 0 rgba(8, 198, 190, 0.04);
}

/* 大纲生成器 - Reading - 层级结构纹理 */
.deco-Reading {
  background:
    linear-gradient(90deg,
      transparent 0%,
      rgba(8, 198, 190, 0.06) 20%,
      rgba(8, 198, 190, 0.06) 22%,
      transparent 22%,
      transparent 40%,
      rgba(8, 198, 190, 0.05) 40%,
      rgba(8, 198, 190, 0.05) 42%,
      transparent 42%,
      transparent 60%,
      rgba(8, 198, 190, 0.04) 60%,
      rgba(8, 198, 190, 0.04) 62%,
      transparent 62%,
      transparent 80%,
      rgba(8, 198, 190, 0.03) 80%,
      rgba(8, 198, 190, 0.03) 82%,
      transparent 82%
    ),
    radial-gradient(circle at 50% 50%, rgba(8, 198, 190, 0.06) 0%, transparent 60%);
}

.deco-Reading::after {
  content: '';
  position: absolute;
  top: 15%;
  left: 25%;
  right: 25%;
  bottom: 15%;
  border-radius: 4px;
  border: 1px solid rgba(8, 198, 190, 0.08);
}

/* 细纲生成器 - Tickets - 票据/网格纹理 */
.deco-Tickets {
  background:
    linear-gradient(rgba(8, 198, 190, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(8, 198, 190, 0.04) 1px, transparent 1px),
    radial-gradient(circle at 50% 50%, rgba(8, 198, 190, 0.06) 0%, transparent 50%);
  background-size: 10px 10px, 10px 10px, 100% 100%;
}

.deco-Tickets::after {
  content: '';
  position: absolute;
  top: 10%;
  left: 10%;
  right: 10%;
  bottom: 10%;
  border-radius: 50%;
  border: 1px dashed rgba(8, 198, 190, 0.1);
}

/* 黄金开篇生成器 - Lightning - 闪电纹理 */
.deco-Lightning {
  background:
    conic-gradient(
      from 45deg,
      transparent 0deg,
      rgba(8, 198, 190, 0.1) 10deg,
      transparent 20deg,
      transparent 90deg,
      rgba(8, 198, 190, 0.08) 100deg,
      transparent 110deg,
      transparent 180deg,
      rgba(8, 198, 190, 0.06) 190deg,
      transparent 200deg,
      transparent 270deg,
      rgba(8, 198, 190, 0.08) 280deg,
      transparent 290deg
    ),
    radial-gradient(circle at 50% 50%, rgba(8, 198, 190, 0.1) 0%, transparent 40%);
}

.deco-Lightning::after {
  content: '';
  position: absolute;
  top: 35%;
  left: 35%;
  right: 35%;
  bottom: 35%;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(8, 198, 190, 0.2) 0%, transparent 70%);
}

/* 金手指生成器 - Cpu - 电路板纹理 */
.deco-Cpu {
  background:
    linear-gradient(0deg, rgba(8, 198, 190, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(8, 198, 190, 0.05) 1px, transparent 1px),
    radial-gradient(circle at 50% 50%, rgba(8, 198, 190, 0.08) 0%, transparent 30%),
    radial-gradient(circle at 20% 20%, rgba(8, 198, 190, 0.04) 0%, transparent 15%),
    radial-gradient(circle at 80% 80%, rgba(8, 198, 190, 0.04) 0%, transparent 15%);
  background-size: 12px 12px, 12px 12px, 100% 100%, 100% 100%, 100% 100%;
}

.deco-Cpu::after {
  content: '';
  position: absolute;
  top: 25%;
  left: 25%;
  right: 25%;
  bottom: 25%;
  border-radius: 4px;
  border: 1px solid rgba(8, 198, 190, 0.1);
  background: rgba(8, 198, 190, 0.04);
}

/* 名字生成器 - CollectionTag - 印章纹理 */
.deco-CollectionTag {
  background:
    repeating-conic-gradient(
      from 0deg,
      rgba(8, 198, 190, 0.03) 0deg 5deg,
      transparent 5deg 10deg
    ),
    radial-gradient(circle at 50% 50%, rgba(8, 198, 190, 0.08) 0%, transparent 50%);
}

.deco-CollectionTag::after {
  content: '';
  position: absolute;
  top: 20%;
  left: 20%;
  right: 20%;
  bottom: 20%;
  border-radius: 50%;
  border: 2px solid rgba(8, 198, 190, 0.1);
}

/* 人设生成器 - User - 轮廓纹理 */
.deco-User {
  background:
    radial-gradient(ellipse 40% 30% at 50% 35%, rgba(8, 198, 190, 0.1) 0%, transparent 70%),
    radial-gradient(ellipse 50% 40% at 50% 75%, rgba(8, 198, 190, 0.06) 0%, transparent 70%),
    radial-gradient(circle at 50% 50%, rgba(8, 198, 190, 0.04) 0%, transparent 60%);
}

.deco-User::after {
  content: '';
  position: absolute;
  top: 10%;
  left: 10%;
  right: 10%;
  bottom: 10%;
  border-radius: 50%;
  border: 1px solid rgba(8, 198, 190, 0.06);
}

/* 世界观生成器 - OfficeBuilding - 建筑结构纹理 */
.deco-OfficeBuilding {
  background:
    linear-gradient(0deg,
      transparent 0%,
      transparent 30%,
      rgba(8, 198, 190, 0.06) 30%,
      rgba(8, 198, 190, 0.06) 32%,
      transparent 32%,
      transparent 50%,
      rgba(8, 198, 190, 0.05) 50%,
      rgba(8, 198, 190, 0.05) 52%,
      transparent 52%,
      transparent 70%,
      rgba(8, 198, 190, 0.04) 70%,
      rgba(8, 198, 190, 0.04) 72%,
      transparent 72%
    ),
    linear-gradient(90deg,
      transparent 0%,
      transparent 30%,
      rgba(8, 198, 190, 0.05) 30%,
      rgba(8, 198, 190, 0.05) 32%,
      transparent 32%,
      transparent 60%,
      rgba(8, 198, 190, 0.04) 60%,
      rgba(8, 198, 190, 0.04) 62%,
      transparent 62%
    ),
    radial-gradient(circle at 50% 50%, rgba(8, 198, 190, 0.06) 0%, transparent 50%);
}

.deco-OfficeBuilding::after {
  content: '';
  position: absolute;
  top: 15%;
  left: 15%;
  right: 15%;
  bottom: 15%;
  border-radius: 4px;
  border: 1px solid rgba(8, 198, 190, 0.08);
}

/* 弹窗样式 - 使用全局样式 */
.generator-dialog-modal {
}

.generator-dialog-modal .el-dialog {
  width: min(88vw, 1500px);
  height: 94vh;
  max-height: 94vh;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  background:
    radial-gradient(circle at top right, rgba(8, 198, 190, 0.12), transparent 28%),
    linear-gradient(180deg, rgba(248, 253, 253, 0.98) 0%, rgba(239, 248, 247, 0.98) 100%);
  border: 1px solid rgba(8, 198, 190, 0.2);
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(6, 95, 70, 0.15);
}

.generator-dialog-modal .el-dialog__header {
  flex-shrink: 0;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.9) 0%, rgba(242, 251, 250, 0.88) 100%);
  border-bottom: 1px solid rgba(8, 198, 190, 0.16);
  padding: 14px 20px;
  backdrop-filter: blur(12px);
}

.generator-dialog-modal .el-dialog__title {
  color: #0f766e;
  font-size: 17px;
  font-weight: 700;
}

.generator-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #5f7f7b;
}

.generator-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #0f766e;
}

.generator-dialog-modal .el-dialog__body {
  flex: 1;
  overflow: hidden;
  background: transparent;
  padding: 0;
}

.generator-dialog-modal .el-dialog__footer {
  flex-shrink: 0;
  border-top: 1px solid rgba(8, 198, 190, 0.14);
  padding: 12px 20px;
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(12px);
}

.dialog-content {
  display: flex;
  height: 100%;
  min-height: 0;
}

.pinned-prompts-area {
  width: 34%;
  min-width: 300px;
  padding: 16px;
  border-right: 1px solid rgba(8, 198, 190, 0.14);
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, rgba(227, 246, 243, 0.76) 0%, rgba(240, 250, 248, 0.66) 100%);
}

.pinned-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(8, 198, 190, 0.16);
  font-weight: 600;
  color: #0f5d56;
  font-size: 14px;
}

.pinned-header .el-button {
  margin-left: auto;
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  border-color: transparent;
  color: #ffffff;
  box-shadow: 0 10px 24px rgba(8, 198, 190, 0.22);
}

.pinned-header .el-button:hover {
  background: linear-gradient(135deg, #09b4ad 0%, #047a75 100%);
  border-color: transparent;
  color: #ffffff;
}

.pinned-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.pinned-item {
  background: rgba(255, 255, 255, 0.76);
  border-radius: 6px;
  padding: 10px 12px;
  margin-bottom: 8px;
  border: 1px solid rgba(8, 198, 190, 0.1);
  transition: all 0.2s;
}

.pinned-item.disabled {
  background: rgba(246, 249, 249, 0.82);
  opacity: 0.72;
}

.pinned-item.disabled .pinned-item-name {
  color: #88a09d;
}

.pinned-item-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.pinned-item-name {
  font-weight: 600;
  color: #134e4a;
  font-size: 14px;
}

.pinned-item-desc {
  font-size: 12px;
  color: #5f7f7b;
  line-height: 1.5;
}

.remove-icon {
  margin-left: auto;
  color: #7f9a96;
  cursor: pointer;
  transition: color 0.2s;
}

.remove-icon:hover {
  color: #f87171;
}

.empty-pinned {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: #7f9a96;
}

.empty-pinned .el-icon {
  color: #7f9a96;
}

.empty-pinned p {
  margin-top: 12px;
  font-size: 14px;
  color: #6b8481;
}

.function-area {
  width: 66%;
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 18px;
  overflow-y: auto;
  background: linear-gradient(180deg, rgba(252, 255, 255, 0.7) 0%, rgba(244, 251, 250, 0.78) 100%);
}

@media (max-width: 1100px) {
  .generator-dialog-modal .el-dialog {
    width: 94vw;
    height: 94vh;
    max-height: 94vh;
  }

  .dialog-content {
    flex-direction: column;
  }

  .pinned-prompts-area,
  .function-area {
    width: 100%;
    min-width: 0;
  }

  .pinned-prompts-area {
    max-height: 32%;
    border-right: none;
    border-bottom: 1px solid rgba(8, 198, 190, 0.14);
  }

  .function-area {
    flex: 1;
  }
}

.selection-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  color: #134e4a;
  font-size: 14px;
  margin-bottom: 2px;
}

.section-meta {
  margin-left: auto;
  font-size: 12px;
  color: #6c8b88;
  font-weight: 500;
}

.fields-input-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.field-input-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.field-label {
  font-weight: 500;
  color: #164e49;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.required-tag {
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

.context-preview-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.context-preview-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  background: rgba(8, 198, 190, 0.08);
  border: 1px solid rgba(8, 198, 190, 0.16);
  color: #245c57;
  font-size: 13px;
}

.full-width {
  width: 100%;
}

.selected-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(74, 222, 128, 0.1);
  border-radius: 6px;
  font-size: 13px;
  color: #4ade80;
}

.generate-btn {
  width: 100%;
  margin-top: auto;
  height: 44px;
  font-size: 15px;
}

.generate-btn .el-icon {
  margin-right: 6px;
}

/* 选项样式 */
.prompt-option,
.model-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.prompt-name,
.model-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.prompt-type {
  font-size: 12px;
  color: #6b7280;
  margin-left: 8px;
}

/* 提示词编辑弹窗样式 */
.prompt-edit-content {
  padding: 10px 0;
}

.prompt-format-hint {
  margin-bottom: 12px;
}

.format-description {
  font-size: 13px;
  line-height: 1.6;
}

.format-description p {
  margin: 4px 0;
}

.format-tip {
  color: #4ade80;
  margin-top: 8px;
  font-weight: 500;
}

.format-example {
  background-color: rgba(74, 222, 128, 0.1);
  border: 1px solid #4ade80;
  padding: 8px 12px;
  border-radius: 4px;
  margin: 4px 0;
  font-family: 'Courier New', monospace;
  color: #ffffff;
}

.format-description code {
  background-color: #333333;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  color: #4ade80;
}

/* 字段配置样式 */
.field-config-section {
  margin-top: 12px;
}

.empty-fields {
  text-align: center;
  padding: 24px;
  background-color: #2a2a2a;
  border-radius: 8px;
  margin-top: 12px;
}

.empty-hint {
  margin-top: 12px;
  color: #6b7280;
  font-size: 14px;
}

.fields-list {
  margin-top: 12px;
}

.field-item {
  background-color: #2a2a2a;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
  border: 1px solid #333333;
}

.field-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.field-name {
  font-weight: 600;
  color: #4ade80;
  font-family: 'Courier New', monospace;
  background-color: rgba(74, 222, 128, 0.1);
  padding: 4px 8px;
  border-radius: 4px;
  border: 1px solid #4ade80;
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

.field-options {
  margin-top: 8px;
}

.field-options label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #ffffff;
}

/* 固定提示词预览样式 */
.fixed-prompt-preview {
  margin-top: 8px;
}

.fixed-prompt-preview :deep(.el-textarea__inner) {
  background-color: #2a2a2a;
  color: #ffffff;
  border-color: #4a4a4a;
}

.fixed-prompt-actions {
  display: flex;
  gap: 8px;
}

.fixed-prompt-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  padding: 6px 12px;
  background: rgba(74, 222, 128, 0.1);
  border-radius: 6px;
  font-size: 13px;
  color: #4ade80;
}

/* 固定提示词弹窗样式 */
.fixed-prompt-dialog-content {
  padding: 10px 0;
}

.fixed-prompt-dialog-content :deep(.el-textarea__inner) {
  background-color: #2a2a2a;
  color: #ffffff;
  font-family: 'Courier New', monospace;
  line-height: 1.6;
  border-color: #4a4a4a;
}

.form-item {
  margin-bottom: 20px;
}

.form-label {
  display: block;
  font-weight: 500;
  color: #ffffff;
  font-size: 14px;
  margin-bottom: 8px;
}

/* 页面头部按钮 */
.header-actions {
  display: flex;
  gap: 12px;
}

.create-generator-btn {
  background: #4ade80;
  border-color: #4ade80;
  color: #1a1a1a;
}

.create-generator-btn:hover {
  background: #22c55e;
  border-color: #22c55e;
}

.history-btn {
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  border-color: transparent;
  color: #ffffff;
  box-shadow: 0 12px 24px rgba(8, 198, 190, 0.18);
}

.history-btn:hover {
  background: linear-gradient(135deg, #09b4ad 0%, #047a75 100%);
  border-color: transparent;
  color: #ffffff;
}

/* 生成器管理弹窗样式 */
.generator-manager-dialog-modal .el-dialog {
  background: #1a1a1a;
}

.generator-manager-dialog-modal .el-dialog__header {
  background: #1a1a1a;
  border-bottom: 1px solid #333333;
}

.generator-manager-dialog-modal .el-dialog__title {
  color: #ffffff;
}

.generator-manager-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #9ca3af;
}

.generator-manager-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #ffffff;
}

.generator-manager-dialog-modal .el-dialog__body {
  padding: 0;
  background: #1a1a1a;
}

.generator-manager-content {
  display: flex;
  height: 500px;
}

.generator-list-panel {
  width: 280px;
  border-right: 1px solid #333333;
  display: flex;
  flex-direction: column;
  background: #1a1a1a;
}

.generator-list-panel .panel-header {
  padding: 16px;
  border-bottom: 1px solid #333333;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel-title {
  font-weight: 600;
  color: #ffffff;
  font-size: 15px;
}

.generator-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.generator-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 4px;
}

.generator-item:hover {
  background-color: #2a2a2a;
}

.generator-item.active {
  background-color: rgba(74, 222, 128, 0.15);
  border: 1px solid #4ade80;
}

.generator-item-icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: linear-gradient(135deg, #4ade80 0%, #22c55e 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #1a1a1a;
  margin-right: 12px;
  flex-shrink: 0;
}

.generator-item-info {
  flex: 1;
  min-width: 0;
}

.generator-item-name {
  display: block;
  font-weight: 500;
  color: #ffffff;
  font-size: 14px;
  margin-bottom: 2px;
}

.generator-item-desc {
  display: block;
  font-size: 12px;
  color: #9ca3af;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.generator-item-actions {
  opacity: 0;
  transition: opacity 0.2s;
}

.generator-item:hover .generator-item-actions {
  opacity: 1;
}

.empty-list {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: #4a4a4a;
}

.empty-list .el-icon {
  color: #4a4a4a;
}

.empty-list p {
  margin-top: 12px;
  font-size: 14px;
  color: #6b7280;
}

.empty-list .hint {
  font-size: 12px;
  margin-top: 4px;
  color: #6b7280;
}

.generator-detail-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: #1a1a1a;
}

.detail-content {
  padding: 24px;
  flex: 1;
  overflow-y: auto;
}

.detail-section {
  margin-bottom: 20px;
}

.detail-label {
  display: block;
  font-weight: 500;
  color: #ffffff;
  font-size: 14px;
  margin-bottom: 8px;
}

.detail-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #333333;
}

.empty-detail {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #4a4a4a;
}

.empty-detail .el-icon {
  color: #4a4a4a;
}

.empty-detail p {
  margin-top: 12px;
  font-size: 14px;
  color: #6b7280;
}

/* 提示词选择弹窗样式 */
.prompt-select-dialog-modal .el-dialog {
  background: #1a1a1a;
}

.prompt-select-dialog-modal .el-dialog__header {
  background: #1a1a1a;
  border-bottom: 1px solid #333333;
}

.prompt-select-dialog-modal .el-dialog__title {
  color: #ffffff;
}

.prompt-select-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #9ca3af;
}

.prompt-select-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #ffffff;
}

.prompt-select-dialog-modal .el-dialog__body {
  padding: 16px 20px;
  background: #1a1a1a;
}

.prompt-select-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.search-input {
  width: 100%;
}

.search-input .el-input__wrapper {
  background: #2a2a2a;
  border-color: #4a4a4a;
  box-shadow: none;
}

.search-input .el-input__inner {
  color: #ffffff;
}

.search-input .el-input__inner::placeholder {
  color: #6b7280;
}

.category-tabs {
  width: 100%;
  background: #2a2a2a;
  border: 1px solid #4a4a4a;
  border-radius: 8px;
  padding: 6px 10px;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: thin;
  scrollbar-color: #4a4a4a #2a2a2a;
}

.category-tabs::-webkit-scrollbar {
  height: 6px;
}

.category-tabs::-webkit-scrollbar-track {
  background: #2a2a2a;
  border-radius: 3px;
}

.category-tabs::-webkit-scrollbar-thumb {
  background: #4a4a4a;
  border-radius: 3px;
}

.category-tabs::-webkit-scrollbar-thumb:hover {
  background: #5a5a5a;
}

.category-tabs-wrapper {
  display: flex;
  gap: 4px;
  user-select: none;
  cursor: grab;
}

.category-tabs-wrapper:active {
  cursor: grabbing;
}

.category-tab {
  flex-shrink: 0;
  padding: 6px 14px;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: #9ca3af;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.category-tab:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #ffffff;
}

.category-tab.active {
  background: rgba(74, 222, 128, 0.2);
  color: #4ade80;
}

.prompt-select-list {
  max-height: 400px;
  overflow-y: auto;
}

.prompt-select-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
}

.prompt-select-item:hover {
  background-color: #2a2a2a;
}

.prompt-select-item.selected {
  background-color: rgba(74, 222, 128, 0.15);
  border-color: #4ade80;
}

.prompt-select-info {
  flex: 1;
  min-width: 0;
}

.prompt-select-name {
  font-weight: 600;
  color: #ffffff;
  font-size: 14px;
  margin-bottom: 4px;
}

.prompt-select-desc {
  font-size: 12px;
  color: #9ca3af;
  line-height: 1.5;
}

.empty-prompts {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 150px;
  color: #4a4a4a;
}

.empty-prompts .el-icon {
  color: #4a4a4a;
}

.empty-prompts p {
  margin-top: 12px;
  font-size: 14px;
  color: #6b7280;
}

/* 预览弹窗样式 */
.preview-dialog-modal .el-dialog {
  background: #fff;
}

.preview-dialog-modal .el-dialog__header {
  background: #fff;
  border-bottom: 1px solid #e5e7eb;
}

.preview-dialog-modal .el-dialog__title {
  color: #1f2937;
}

.preview-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #6b7280;
}

.preview-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #1f2937;
}

.preview-dialog-modal .el-dialog__body {
  padding: 0;
  background: #fff;
}

.preview-dialog-content {
  display: flex;
  flex-direction: column;
  height: 60vh;
}

.preview-dialog-header {
  display: flex;
  justify-content: flex-end;
  padding: 12px 16px;
  border-bottom: 1px solid #e5e7eb;
  background: #f9fafb;
}

.preview-scrollbar {
  flex: 1;
  padding: 20px;
  background: #fff;
}

/* 弹窗底部按钮 */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* Element Plus 浅色主题覆盖 - 生成器弹窗 */
.generator-dialog-modal .el-input__wrapper {
  background: rgba(255, 255, 255, 0.92);
  border-color: rgba(8, 198, 190, 0.14);
  box-shadow: none;
  border-radius: 6px;
}

.generator-dialog-modal .el-input__inner {
  color: #18453f;
}

.generator-dialog-modal .el-input__inner::placeholder {
  color: #88a09d;
}

.generator-dialog-modal .el-textarea__inner {
  background: rgba(255, 255, 255, 0.92);
  border-color: rgba(8, 198, 190, 0.14);
  color: #18453f;
  box-shadow: none;
  border-radius: 6px;
}

.generator-dialog-modal .el-textarea__inner::placeholder {
  color: #88a09d;
}

.generator-dialog-modal .el-textarea__inner:focus {
  border-color: #08c6be;
}

.generator-dialog-modal .el-input__wrapper:focus-within {
  border-color: #08c6be;
  box-shadow: 0 0 0 2px rgba(8, 198, 190, 0.1);
}

.generator-dialog-modal .el-select .el-input__wrapper {
  background: rgba(255, 255, 255, 0.92);
  border-color: rgba(8, 198, 190, 0.14);
  border-radius: 6px;
}

.generator-dialog-modal .el-select-dropdown {
  background: #ffffff;
  border-color: rgba(8, 198, 190, 0.16);
}

.generator-dialog-modal .el-select-dropdown__item {
  color: #18453f;
}

.generator-dialog-modal .el-select-dropdown__item:hover {
  background: rgba(8, 198, 190, 0.08);
}

.generator-dialog-modal .el-select-dropdown__item.selected {
  background: rgba(8, 198, 190, 0.14);
  color: #0f766e;
}

.generator-dialog-modal .el-switch.is-checked .el-switch__core {
  background-color: #08c6be;
  border-color: #08c6be;
}

.generator-dialog-modal .el-button--primary {
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  border-color: transparent;
  color: #ffffff;
  box-shadow: 0 8px 16px rgba(8, 198, 190, 0.15);
  border-radius: 6px;
}

.generator-dialog-modal .el-button--primary:hover {
  background: linear-gradient(135deg, #09b4ad 0%, #047a75 100%);
  border-color: transparent;
  color: #ffffff;
}

.generator-dialog-modal .el-button--primary.is-disabled {
  background: #d8e7e5;
  border-color: #d8e7e5;
  color: #89a3a0;
  box-shadow: none;
}

.generator-dialog-modal .el-button--default {
  background: rgba(255, 255, 255, 0.92);
  border-color: rgba(8, 198, 190, 0.16);
  color: #14534d;
}

.generator-dialog-modal .el-button--default:hover {
  background: rgba(8, 198, 190, 0.08);
  border-color: rgba(8, 198, 190, 0.3);
  color: #0f766e;
}

.generator-dialog-modal .el-button--danger {
  background: #f87171;
  border-color: #f87171;
  color: #ffffff;
}

.generator-dialog-modal .el-button--danger:hover {
  background: #ef4444;
  border-color: #ef4444;
}

.generator-dialog-modal .el-tag--info {
  background: rgba(8, 198, 190, 0.1);
  border-color: rgba(8, 198, 190, 0.18);
  color: #0f766e;
}

.generator-dialog-modal .el-icon {
  color: #69918c;
}

.generator-dialog-modal .el-icon:hover {
  color: #0f766e;
}

.generator-dialog-modal .function-area::-webkit-scrollbar,
.generator-dialog-modal .pinned-list::-webkit-scrollbar {
  width: 6px;
}

.generator-dialog-modal .function-area::-webkit-scrollbar-track,
.generator-dialog-modal .pinned-list::-webkit-scrollbar-track {
  background: transparent;
}

.generator-dialog-modal .function-area::-webkit-scrollbar-thumb,
.generator-dialog-modal .pinned-list::-webkit-scrollbar-thumb {
  background: rgba(8, 198, 190, 0.2);
  border-radius: 3px;
}

.generator-dialog-modal .function-area::-webkit-scrollbar-thumb:hover,
.generator-dialog-modal .pinned-list::-webkit-scrollbar-thumb:hover {
  background: rgba(8, 198, 190, 0.35);
}

/* 生成器管理弹窗 */
.generator-manager-dialog-modal .el-input__wrapper {
  background: #2a2a2a;
  border-color: #4a4a4a;
  box-shadow: none;
}

.generator-manager-dialog-modal .el-input__inner {
  color: #ffffff;
}

.generator-manager-dialog-modal .el-input__inner::placeholder {
  color: #6b7280;
}

.generator-manager-dialog-modal .el-textarea__inner {
  background: #2a2a2a;
  border-color: #4a4a4a;
  color: #ffffff;
}

.generator-manager-dialog-modal .el-textarea__inner::placeholder {
  color: #6b7280;
}

.generator-manager-dialog-modal .el-textarea__inner:focus {
  border-color: #4ade80;
}

.generator-manager-dialog-modal .el-input__wrapper:focus-within {
  border-color: #4ade80;
}

.generator-manager-dialog-modal .el-select .el-input__wrapper {
  background: #2a2a2a;
  border-color: #4a4a4a;
}

.generator-manager-dialog-modal .el-select-dropdown {
  background: #2a2a2a;
  border-color: #333333;
}

.generator-manager-dialog-modal .el-select-dropdown__item {
  color: #ffffff;
}

.generator-manager-dialog-modal .el-select-dropdown__item:hover {
  background: #333333;
}

.generator-manager-dialog-modal .el-select-dropdown__item.selected {
  background: rgba(74, 222, 128, 0.15);
  color: #4ade80;
}

.generator-manager-dialog-modal .el-switch.is-checked .el-switch__core {
  background-color: #4ade80;
  border-color: #4ade80;
}

.generator-manager-dialog-modal .el-button--primary {
  background: #4ade80;
  border-color: #4ade80;
  color: #1a1a1a;
}

.generator-manager-dialog-modal .el-button--primary:hover {
  background: #22c55e;
  border-color: #22c55e;
}

.generator-manager-dialog-modal .el-button--primary.is-disabled {
  background: #4a4a4a;
  border-color: #4a4a4a;
  color: #6b7280;
}

.generator-manager-dialog-modal .el-button--default {
  background: #2a2a2a;
  border-color: #4a4a4a;
  color: #ffffff;
}

.generator-manager-dialog-modal .el-button--default:hover {
  background: #333333;
  border-color: #6b7280;
}

.generator-manager-dialog-modal .el-button--danger {
  background: #f87171;
  border-color: #f87171;
  color: #1a1a1a;
}

.generator-manager-dialog-modal .el-button--danger:hover {
  background: #ef4444;
  border-color: #ef4444;
}

.generator-manager-dialog-modal .el-tag--info {
  background: rgba(107, 114, 128, 0.2);
  border-color: #6b7280;
  color: #9ca3af;
}

.generator-manager-dialog-modal .el-icon {
  color: #9ca3af;
}

.generator-manager-dialog-modal .el-icon:hover {
  color: #ffffff;
}

/* 提示词选择弹窗 */
.prompt-select-dialog-modal .el-input__wrapper {
  background: #2a2a2a;
  border-color: #4a4a4a;
  box-shadow: none;
}

.prompt-select-dialog-modal .el-input__inner {
  color: #ffffff;
}

.prompt-select-dialog-modal .el-input__inner::placeholder {
  color: #6b7280;
}

.prompt-select-dialog-modal .el-textarea__inner {
  background: #2a2a2a;
  border-color: #4a4a4a;
  color: #ffffff;
}

.prompt-select-dialog-modal .el-textarea__inner::placeholder {
  color: #6b7280;
}

.prompt-select-dialog-modal .el-button--primary {
  background: #4ade80;
  border-color: #4ade80;
  color: #1a1a1a;
}

.prompt-select-dialog-modal .el-button--primary:hover {
  background: #22c55e;
  border-color: #22c55e;
}

.prompt-select-dialog-modal .el-button--default {
  background: #2a2a2a;
  border-color: #4a4a4a;
  color: #ffffff;
}

.prompt-select-dialog-modal .el-button--default:hover {
  background: #333333;
  border-color: #6b7280;
}

.prompt-select-dialog-modal .el-tag--info {
  background: rgba(107, 114, 128, 0.2);
  border-color: #6b7280;
  color: #9ca3af;
}

.prompt-select-dialog-modal .el-icon {
  color: #9ca3af;
}

.prompt-select-dialog-modal .el-icon:hover {
  color: #ffffff;
}

.prompt-select-dialog-modal .el-checkbox__label {
  color: #ffffff;
}

/* 预览弹窗 */
.preview-dialog-modal .el-input__wrapper {
  background: #fff;
  border-color: #d1d5db;
  box-shadow: none;
}

.preview-dialog-modal .el-input__inner {
  color: #1f2937;
}

.preview-dialog-modal .el-input__inner::placeholder {
  color: #9ca3af;
}

.preview-dialog-modal .el-textarea__inner {
  background: #fff;
  border-color: #d1d5db;
  color: #1f2937;
}

.preview-dialog-modal .el-textarea__inner::placeholder {
  color: #9ca3af;
}

.preview-dialog-modal .el-button--primary {
  background: #4ade80;
  border-color: #4ade80;
  color: #1a1a1a;
}

.preview-dialog-modal .el-button--primary:hover {
  background: #22c55e;
  border-color: #22c55e;
}

.preview-dialog-modal .el-button--default {
  background: #2a2a2a;
  border-color: #4a4a4a;
  color: #ffffff;
}

.preview-dialog-modal .el-button--default:hover {
  background: #333333;
  border-color: #6b7280;
}

.preview-dialog-modal .el-icon {
  color: #9ca3af;
}

.preview-dialog-modal .el-icon:hover {
  color: #ffffff;
}

/* 系统提示词修改弹窗 */
.prompt-edit-dialog-modal .el-dialog {
  background: #1a1a1a;
}

.prompt-edit-dialog-modal .el-dialog__header {
  background: #1a1a1a;
  border-bottom: 1px solid #333333;
}

.prompt-edit-dialog-modal .el-dialog__title {
  color: #ffffff;
}

.prompt-edit-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #9ca3af;
}

.prompt-edit-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #ffffff;
}

.prompt-edit-dialog-modal .el-dialog__body {
  background: #1a1a1a;
}

.prompt-edit-dialog-modal .el-dialog__footer {
  background: #1a1a1a;
  border-top: 1px solid #333333;
}

.prompt-edit-dialog-modal .el-input__wrapper {
  background: #2a2a2a;
  border-color: #4a4a4a;
  box-shadow: none;
}

.prompt-edit-dialog-modal .el-input__inner {
  color: #ffffff;
}

.prompt-edit-dialog-modal .el-input__inner::placeholder {
  color: #6b7280;
}

.prompt-edit-dialog-modal .el-textarea__inner {
  background: #2a2a2a;
  border-color: #4a4a4a;
  color: #ffffff;
}

.prompt-edit-dialog-modal .el-textarea__inner::placeholder {
  color: #6b7280;
}

.prompt-edit-dialog-modal .el-textarea__inner:focus {
  border-color: #4ade80;
}

.prompt-edit-dialog-modal .el-input__wrapper:focus-within {
  border-color: #4ade80;
}

.prompt-edit-dialog-modal .el-button--primary {
  background: #4ade80;
  border-color: #4ade80;
  color: #1a1a1a;
}

.prompt-edit-dialog-modal .el-button--primary:hover {
  background: #22c55e;
  border-color: #22c55e;
}

.prompt-edit-dialog-modal .el-button--default {
  background: #2a2a2a;
  border-color: #4a4a4a;
  color: #ffffff;
}

.prompt-edit-dialog-modal .el-button--default:hover {
  background: #333333;
  border-color: #6b7280;
}

/* 固定提示词查看弹窗 */
.fixed-prompt-dialog-modal .el-dialog {
  background: #1a1a1a;
}

.fixed-prompt-dialog-modal .el-dialog__header {
  background: #1a1a1a;
  border-bottom: 1px solid #333333;
}

.fixed-prompt-dialog-modal .el-dialog__title {
  color: #ffffff;
}

.fixed-prompt-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #9ca3af;
}

.fixed-prompt-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #ffffff;
}

.fixed-prompt-dialog-modal .el-dialog__body {
  background: #1a1a1a;
}

.fixed-prompt-dialog-modal .el-dialog__footer {
  background: #1a1a1a;
  border-top: 1px solid #333333;
}

.fixed-prompt-dialog-modal .el-input__wrapper {
  background: #2a2a2a;
  border-color: #4a4a4a;
  box-shadow: none;
}

.fixed-prompt-dialog-modal .el-input__inner {
  color: #ffffff;
}

.fixed-prompt-dialog-modal .el-input__inner::placeholder {
  color: #6b7280;
}

.fixed-prompt-dialog-modal .el-textarea__inner {
  background: #2a2a2a;
  border-color: #4a4a4a;
  color: #ffffff;
}

.fixed-prompt-dialog-modal .el-textarea__inner::placeholder {
  color: #6b7280;
}

.fixed-prompt-dialog-modal .el-textarea__inner:focus {
  border-color: #4ade80;
}

.fixed-prompt-dialog-modal .el-input__wrapper:focus-within {
  border-color: #4ade80;
}

.fixed-prompt-dialog-modal .el-button--primary {
  background: #4ade80;
  border-color: #4ade80;
  color: #1a1a1a;
}

.fixed-prompt-dialog-modal .el-button--primary:hover {
  background: #22c55e;
  border-color: #22c55e;
}

.fixed-prompt-dialog-modal .el-button--default {
  background: #2a2a2a;
  border-color: #4a4a4a;
  color: #ffffff;
}

.fixed-prompt-dialog-modal .el-button--default:hover {
  background: #333333;
  border-color: #6b7280;
}

/* Alert组件深色主题 */
.el-alert--info {
  background: rgba(74, 222, 128, 0.1);
  border-color: #4ade80;
}

.el-alert--info .el-alert__title {
  color: #4ade80;
}

.el-alert--info .el-alert__description {
  color: #ffffff;
}

.el-alert--warning {
  background: rgba(251, 191, 36, 0.1);
  border-color: #fbbf24;
}

.el-alert--warning .el-alert__title {
  color: #fbbf24;
}

.el-alert--warning .el-alert__description {
  color: #ffffff;
}

.el-alert .el-alert__icon {
  color: inherit;
}

/* 历史记录对话框样式 */
.history-dialog-modal .el-overlay {
  background: rgba(0, 0, 0, 0.5);
}

.history-dialog-modal .el-dialog {
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
}

.history-dialog-modal .el-dialog__header {
  background: #ffffff;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
  border-radius: 0;
}

.history-dialog-modal .el-dialog__title {
  color: #1f2937;
  font-size: 18px;
  font-weight: 600;
}

.history-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #6b7280;
  font-size: 20px;
}

.history-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #1f2937;
}

.history-dialog-modal .el-dialog__body {
  padding: 24px;
  max-height: 70vh;
  overflow-y: auto;
  background: #ffffff;
}

.history-dialog-modal .el-dialog__footer {
  padding: 16px 24px;
  border-top: 1px solid #e5e7eb;
  background: #ffffff;
  border-radius: 0 0 16px 16px;
}

.history-dialog-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: #f9fafb;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-icon {
  color: #667eea;
  font-size: 24px;
}

.history-count {
  color: #1f2937;
  font-size: 16px;
  font-weight: 500;
}

.clear-all-btn {
  background: rgba(244, 63, 94, 0.1);
  border-color: #f43f5e;
  color: #f43f5e;
  transition: all 0.3s ease;
}

.clear-all-btn:hover:not(:disabled) {
  background: #f43f5e;
  color: #ffffff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(244, 63, 94, 0.2);
}

.clear-all-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.history-item {
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.history-item-card {
  background: #ffffff;
  border-radius: 12px;
  padding: 20px;
  border: 1px solid #e5e7eb;
  transition: all 0.3s ease;
  cursor: pointer;
  position: relative;
  overflow: hidden;
}

.history-item-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 4px;
  height: 100%;
  background: #667eea;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.history-item-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.15);
  border-color: #667eea;
}

.history-item-card:hover::before {
  opacity: 1;
}

.history-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.generator-info {
  display: flex;
  align-items: center;
  gap: 14px;
  flex: 1;
}

.generator-icon-wrapper {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  background: #f3f4f6;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #e5e7eb;
}

.generator-icon {
  color: #667eea;
}

.generator-details {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.generator-name {
  color: #1f2937;
  font-size: 16px;
  font-weight: 600;
}

.generator-time {
  color: #9ca3af;
  font-size: 13px;
}

.history-item-tags {
  display: flex;
  gap: 8px;
}

.prompt-count-tag {
  border-color: #e5e7eb;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: 500;
  background: #f9fafb;
}

.prompt-count-tag .el-icon {
  font-size: 14px;
}

.history-item-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid rgba(0, 0, 0, 0.06);
}

.view-detail-btn {
  background: #667eea;
  border: none;
  color: #ffffff;
  transition: all 0.3s ease;
}

.view-detail-btn:hover {
  background: #5568d3;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.delete-btn {
  background: rgba(244, 63, 94, 0.1);
  border-color: #f43f5e;
  color: #f43f5e;
  transition: all 0.3s ease;
}

.delete-btn:hover {
  background: #f43f5e;
  color: #ffffff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(244, 63, 94, 0.2);
}

.empty-history {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.empty-state-icon {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  background: #f9fafb;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
  border: 2px solid #e5e7eb;
}

.empty-state-icon .el-icon {
  color: #667eea;
}

.empty-title {
  color: #1f2937;
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 8px;
}

.empty-hint {
  color: #9ca3af;
  font-size: 14px;
  line-height: 1.6;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.cancel-btn {
  background: #ffffff;
  border-color: #e5e7eb;
  color: #6b7280;
  transition: all 0.3s ease;
}

.cancel-btn:hover {
  background: #f9fafb;
  border-color: #d1d5db;
  transform: translateY(-2px);
}

.confirm-btn {
  background: #667eea;
  border: none;
  color: #ffffff;
  transition: all 0.3s ease;
}

.confirm-btn:hover {
  background: #5568d3;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

/* 历史记录详情对话框样式 */
.history-detail-dialog-modal .el-overlay {
  background: rgba(0, 0, 0, 0.5);
}

.history-detail-dialog-modal .el-dialog {
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
}

.history-detail-dialog-modal .el-dialog__header {
  background: #ffffff;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
  border-radius: 0;
}

.history-detail-dialog-modal .el-dialog__title {
  color: #1f2937;
  font-size: 18px;
  font-weight: 600;
}

.history-detail-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #6b7280;
  font-size: 20px;
}

.history-detail-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #1f2937;
}

.history-detail-dialog-modal .el-dialog__body {
  padding: 24px;
  max-height: 70vh;
  overflow-y: auto;
  background: #ffffff;
}

.history-detail-dialog-modal .el-dialog__footer {
  padding: 16px 24px;
  border-top: 1px solid #e5e7eb;
  background: #ffffff;
  border-radius: 0 0 16px 16px;
}

.history-detail-content {
  background: #ffffff;
}

.conversation-messages {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message-bubble {
  border-radius: 16px;
  padding: 16px 20px;
  max-width: 85%;
  animation: messageSlideIn 0.3s ease;
}

@keyframes messageSlideIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.message-bubble.user {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  align-self: flex-start;
  border-bottom-left-radius: 4px;
}

.message-bubble.assistant {
  background: #f3f4f6;
  border: 1px solid #e5e7eb;
  align-self: flex-end;
  border-bottom-right-radius: 4px;
}

.message-bubble.system {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  align-self: center;
  width: 100%;
  max-width: 100%;
}

.message-sender {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.sender-icon {
  font-size: 18px;
}

.message-bubble.user .sender-icon {
  color: #ffffff;
}

.message-bubble.assistant .sender-icon {
  color: #667eea;
}

.message-bubble.system .sender-icon {
  color: #667eea;
}

.sender-name {
  font-weight: 600;
  font-size: 14px;
}

.message-bubble.user .sender-name {
  color: #ffffff;
}

.message-bubble.assistant .sender-name {
  color: #1f2937;
}

.message-bubble.system .sender-name {
  color: #1f2937;
}

.message-time {
  color: #9ca3af;
  font-size: 12px;
}

.message-content {
  line-height: 1.8;
  font-size: 14px;
}

.message-bubble.user .message-content {
  color: #ffffff;
}

.message-bubble.assistant .message-content {
  color: #1f2937;
}

.message-bubble.system .message-content {
  color: #1f2937;
}

/* 暗色主题适配 */
:root[data-theme='dark'] .message-bubble.user {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.8) 0%, rgba(118, 75, 162, 0.8) 100%);
  border: 1px solid rgba(102, 126, 234, 0.3);
}

:root[data-theme='dark'] .message-bubble.assistant {
  background: rgba(30, 41, 59, 0.95);
  border: 1px solid rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .message-bubble.system {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(71, 85, 105, 0.3);
}

:root[data-theme='dark'] .message-bubble.assistant .sender-name {
  color: #f3f4f6;
}

:root[data-theme='dark'] .message-bubble.system .sender-name {
  color: #f3f4f6;
}

:root[data-theme='dark'] .message-bubble.assistant .message-content {
  color: #e5e7eb;
}

:root[data-theme='dark'] .message-bubble.system .message-content {
  color: #e5e7eb;
}

:root[data-theme='dark'] .page-header h1 {
  color: #f3f4f6;
}

:root[data-theme='dark'] .subtitle {
  color: #9ca3af;
}

:root[data-theme='dark'] .pinned-item {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(94, 234, 212, 0.2);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.2);
}

:root[data-theme='dark'] .pinned-item.disabled {
  background: rgba(30, 41, 59, 0.5);
}

:root[data-theme='dark'] .pinned-item.disabled .pinned-item-name {
  color: #6b7280;
}

:root[data-theme='dark'] .pinned-item-name {
  color: #f3f4f6;
}

:root[data-theme='dark'] .pinned-item-desc {
  color: #9ca3af;
}

:root[data-theme='dark'] .remove-icon {
  color: #9ca3af;
}

:root[data-theme='dark'] .remove-icon:hover {
  color: #f87171;
}

:root[data-theme='dark'] .empty-pinned {
  color: #6b7280;
}

:root[data-theme='dark'] .empty-pinned .el-icon {
  color: #6b7280;
}

:root[data-theme='dark'] .empty-pinned p {
  color: #6b7280;
}

:root[data-theme='dark'] .selection-section {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(94, 234, 212, 0.2);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.2);
}

:root[data-theme='dark'] .section-header {
  color: #f3f4f6;
}

:root[data-theme='dark'] .section-meta {
  color: #9ca3af;
}

:root[data-theme='dark'] .field-label {
  color: #e5e7eb;
}

:root[data-theme='dark'] .field-description {
  color: #9ca3af;
}

:root[data-theme='dark'] .context-tip {
  color: #9ca3af;
}

:root[data-theme='dark'] .fixed-prompt-hint {
  background: rgba(74, 222, 128, 0.15);
  color: #4ade80;
}

:root[data-theme='dark'] .pinned-prompts-area {
  border-right-color: rgba(94, 234, 212, 0.2);
  background: linear-gradient(180deg, rgba(30, 41, 59, 0.76) 0%, rgba(15, 23, 42, 0.66) 100%);
}

:root[data-theme='dark'] .pinned-header {
  border-bottom-color: rgba(94, 234, 212, 0.2);
  color: #5eead4;
}

:root[data-theme='dark'] .pinned-header .el-button {
  background: linear-gradient(135deg, #14b8a6 0%, #0d9488 100%);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .pinned-header .el-button:hover {
  background: linear-gradient(135deg, #0d9488 0%, #0f766e 100%);
}

:root[data-theme='dark'] .function-area {
  background: linear-gradient(180deg, rgba(30, 41, 59, 0.7) 0%, rgba(15, 23, 42, 0.78) 100%);
}

:root[data-theme='dark'] .creative-page {
  background: #0a0c10;
}

:root[data-theme='dark'] .creative-page::before {
  background:
    radial-gradient(ellipse at 20% 50%, rgba(8, 198, 190, 0.03) 0%, transparent 50%),
    radial-gradient(ellipse at 80% 80%, rgba(5, 150, 145, 0.02) 0%, transparent 50%);
}

:root[data-theme='dark'] .page-header h1 {
  color: #ffffff;
  text-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .subtitle {
  color: #64748b;
}

:root[data-theme='dark'] .generator-card {
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.04) 0%, rgba(255, 255, 255, 0.01) 100%);
  border-color: rgba(8, 198, 190, 0.08);
  box-shadow:
    0 4px 24px rgba(0, 0, 0, 0.5),
    inset 0 1px 0 rgba(255, 255, 255, 0.03);
}

:root[data-theme='dark'] .card-texture {
  background-image:
    radial-gradient(circle at 25% 25%, rgba(8, 198, 190, 0.08) 0%, transparent 50%),
    radial-gradient(circle at 75% 75%, rgba(5, 150, 145, 0.05) 0%, transparent 50%);
}

:root[data-theme='dark'] .generator-card .card-title {
  color: #f1f5f9;
}

:root[data-theme='dark'] .generator-card::before {
  background: linear-gradient(90deg, transparent, rgba(8, 198, 190, 0.25), transparent);
}

:root[data-theme='dark'] .generator-card::after {
  background: radial-gradient(circle at 30% 50%, rgba(8, 198, 190, 0.12) 0%, transparent 60%);
}

:root[data-theme='dark'] .generator-card:hover {
  box-shadow:
    0 20px 48px rgba(0, 0, 0, 0.6),
    0 0 0 1px rgba(8, 198, 190, 0.25),
    inset 0 1px 0 rgba(255, 255, 255, 0.06);
  border-color: rgba(8, 198, 190, 0.3);
}

:root[data-theme='dark'] .card-icon {
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  box-shadow:
    0 6px 20px rgba(8, 198, 190, 0.45),
    inset 0 1px 0 rgba(255, 255, 255, 0.18);
}

:root[data-theme='dark'] .generator-card:hover .card-icon {
  box-shadow:
    0 10px 28px rgba(8, 198, 190, 0.55),
    inset 0 1px 0 rgba(255, 255, 255, 0.25);
}

:root[data-theme='dark'] .card-desc {
  color: #94a3b8;
}

:root[data-theme='dark'] .new-tag {
  color: #ffffff;
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.35);
}

:root[data-theme='dark'] .generator-item:hover {
  background-color: rgba(51, 65, 85, 0.6);
}

:root[data-theme='dark'] .generator-item.active {
  background-color: rgba(74, 222, 128, 0.15);
  border-color: #4ade80;
}

:root[data-theme='dark'] .empty-list {
  color: #6b7280;
}

:root[data-theme='dark'] .empty-list .el-icon {
  color: #6b7280;
}

:root[data-theme='dark'] .empty-list p {
  color: #6b7280;
}

:root[data-theme='dark'] .empty-list .hint {
  color: #6b7280;
}

:root[data-theme='dark'] .empty-detail {
  color: #6b7280;
}

:root[data-theme='dark'] .empty-detail .el-icon {
  color: #6b7280;
}

:root[data-theme='dark'] .generator-dialog-modal .el-dialog {
  background:
    radial-gradient(circle at top right, rgba(94, 234, 212, 0.1), transparent 28%),
    linear-gradient(180deg, rgba(30, 41, 59, 0.98) 0%, rgba(15, 23, 42, 0.98) 100%);
  border-color: rgba(94, 234, 212, 0.3);
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.4);
}

:root[data-theme='dark'] .generator-dialog-modal .el-dialog__header {
  background: linear-gradient(180deg, rgba(30, 41, 59, 0.9) 0%, rgba(15, 23, 42, 0.88) 100%);
  border-bottom-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .generator-dialog-modal .el-dialog__title {
  color: #5eead4;
}

:root[data-theme='dark'] .generator-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #9ca3af;
}

:root[data-theme='dark'] .generator-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-dialog {
  background: #1e293b;
  border: 1px solid rgba(94, 234, 212, 0.3);
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-dialog__header {
  background: #1e293b;
  border-bottom-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-dialog__title {
  color: #5eead4;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #9ca3af;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-dialog__body {
  background: #1e293b;
}

:root[data-theme='dark'] .generator-list-panel {
  border-right-color: rgba(94, 234, 212, 0.2);
  background: #1e293b;
}

:root[data-theme='dark'] .generator-list-panel .panel-header {
  border-bottom-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .generator-detail-panel {
  background-color: #1e293b;
}

:root[data-theme='dark'] .detail-actions {
  border-top-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .prompt-select-dialog-modal .el-dialog {
  background: #1e293b;
  border: 1px solid rgba(94, 234, 212, 0.3);
}

:root[data-theme='dark'] .prompt-select-dialog-modal .el-dialog__header {
  background: #1e293b;
  border-bottom-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .prompt-select-dialog-modal .el-dialog__title {
  color: #5eead4;
}

:root[data-theme='dark'] .prompt-select-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #9ca3af;
}

:root[data-theme='dark'] .prompt-select-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .prompt-select-dialog-modal .el-dialog__body {
  background: #1e293b;
}

:root[data-theme='dark'] .preview-dialog-modal .el-dialog {
  background: #1e293b;
  border: 1px solid rgba(94, 234, 212, 0.3);
}

:root[data-theme='dark'] .preview-dialog-modal .el-dialog__header {
  background: #1e293b;
  border-bottom-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .preview-dialog-modal .el-dialog__title {
  color: #5eead4;
}

:root[data-theme='dark'] .preview-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #9ca3af;
}

:root[data-theme='dark'] .preview-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .preview-dialog-modal .el-dialog__body {
  background: #1e293b;
}

:root[data-theme='dark'] .preview-dialog-header {
  border-bottom-color: rgba(94, 234, 212, 0.2);
  background: rgba(30, 41, 59, 0.8);
}

:root[data-theme='dark'] .preview-scrollbar {
  background: #1e293b;
}

:root[data-theme='dark'] .prompt-edit-dialog-modal .el-dialog {
  background: #1e293b;
  border: 1px solid rgba(94, 234, 212, 0.3);
}

:root[data-theme='dark'] .prompt-edit-dialog-modal .el-dialog__header {
  background: #1e293b;
  border-bottom-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .prompt-edit-dialog-modal .el-dialog__title {
  color: #5eead4;
}

:root[data-theme='dark'] .prompt-edit-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #9ca3af;
}

:root[data-theme='dark'] .prompt-edit-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .prompt-edit-dialog-modal .el-dialog__body {
  background: #1e293b;
}

:root[data-theme='dark'] .prompt-edit-dialog-modal .el-dialog__footer {
  background: #1e293b;
  border-top-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .fixed-prompt-dialog-modal .el-dialog {
  background: #1e293b;
  border: 1px solid rgba(94, 234, 212, 0.3);
}

:root[data-theme='dark'] .fixed-prompt-dialog-modal .el-dialog__header {
  background: #1e293b;
  border-bottom-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .fixed-prompt-dialog-modal .el-dialog__title {
  color: #5eead4;
}

:root[data-theme='dark'] .fixed-prompt-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #9ca3af;
}

:root[data-theme='dark'] .fixed-prompt-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .fixed-prompt-dialog-modal .el-dialog__body {
  background: #1e293b;
}

:root[data-theme='dark'] .fixed-prompt-dialog-modal .el-dialog__footer {
  background: #1e293b;
  border-top-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .history-dialog-modal .el-dialog {
  background: #1e293b;
  border: 1px solid rgba(94, 234, 212, 0.3);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
}

:root[data-theme='dark'] .history-dialog-modal .el-dialog__header {
  background: #1e293b;
  border-bottom-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .history-dialog-modal .el-dialog__title {
  color: #5eead4;
}

:root[data-theme='dark'] .history-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #9ca3af;
}

:root[data-theme='dark'] .history-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .history-dialog-modal .el-dialog__body {
  background: #1e293b;
}

:root[data-theme='dark'] .history-dialog-modal .el-dialog__footer {
  border-top-color: rgba(94, 234, 212, 0.2);
  background: #1e293b;
}

:root[data-theme='dark'] .history-detail-dialog-modal .el-dialog {
  background: #1e293b;
  border: 1px solid rgba(94, 234, 212, 0.3);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
}

:root[data-theme='dark'] .history-detail-dialog-modal .el-dialog__header {
  background: #1e293b;
  border-bottom-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .history-detail-dialog-modal .el-dialog__title {
  color: #5eead4;
}

:root[data-theme='dark'] .history-detail-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #9ca3af;
}

:root[data-theme='dark'] .history-detail-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .history-detail-dialog-modal .el-dialog__body {
  background: #1e293b;
}

:root[data-theme='dark'] .history-detail-dialog-modal .el-dialog__footer {
  border-top-color: rgba(94, 234, 212, 0.2);
  background: #1e293b;
}

:root[data-theme='dark'] .history-detail-content {
  background: #1e293b;
}

:root[data-theme='dark'] .history-item-card {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .history-item-card::before {
  background: #5eead4;
}

:root[data-theme='dark'] .history-item-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
  border-color: #5eead4;
}

:root[data-theme='dark'] .generator-icon-wrapper {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(94, 234, 212, 0.3);
}

:root[data-theme='dark'] .generator-icon {
  color: #5eead4;
}

:root[data-theme='dark'] .generator-name {
  color: #f3f4f6;
}

:root[data-theme='dark'] .generator-time {
  color: #9ca3af;
}

:root[data-theme='dark'] .prompt-count-tag {
  border-color: rgba(94, 234, 212, 0.3);
  color: #9ca3af;
  background: rgba(51, 65, 85, 0.6);
}

:root[data-theme='dark'] .history-item-actions {
  border-top-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .view-detail-btn {
  background: #14b8a6;
}

:root[data-theme='dark'] .view-detail-btn:hover {
  background: #0d9488;
  box-shadow: 0 4px 12px rgba(20, 184, 166, 0.3);
}

:root[data-theme='dark'] .delete-btn {
  background: rgba(244, 63, 94, 0.15);
}

:root[data-theme='dark'] .history-header {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .header-icon {
  color: #5eead4;
}

:root[data-theme='dark'] .history-count {
  color: #f3f4f6;
}

:root[data-theme='dark'] .clear-all-btn {
  background: rgba(244, 63, 94, 0.15);
  border-color: #f43f5e;
  color: #f43f5e;
}

:root[data-theme='dark'] .clear-all-btn:hover:not(:disabled) {
  background: #f43f5e;
}
</style>
