<template>
  <div class="creative-page">
    <div class="page-header">
      <div class="header-content">
        <h1>抽卡区</h1>
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
        <div class="card-watermark" aria-hidden="true">
          <el-icon :size="44"><component :is="getIconComponent(generator.icon)" /></el-icon>
        </div>
        <div class="card-content-bottom">
          <div class="card-header">
            <div class="card-title">{{ generator.name }}</div>
            <el-tag v-if="generator.isNew" size="small" type="danger" class="new-tag">NEW</el-tag>
          </div>
          <div class="card-arrow-box" aria-hidden="true">
            <el-icon :size="14"><ArrowRight /></el-icon>
          </div>
        </div>
      </div>
    </div>

    <!-- 生成器弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      width="1200px"
      align-center
      destroy-on-close
      class="generator-dialog"
      modal-class="generator-dialog-modal"
      :show-close="false"
      append-to-body
    >
      <!-- 自定义标题栏 -->
      <template #header>
        <div class="gen-dialog-header">
          <div class="gen-dialog-title">
            <el-icon class="gen-title-icon"><MagicStick /></el-icon>
            <span>{{ currentGenerator?.name || '创意生成' }}</span>
          </div>
          <div class="gen-dialog-header-right">
            <el-button
              type="primary"
              link
              size="small"
              class="gen-close-btn"
              @click="dialogVisible = false"
            >
              <el-icon :size="18"><Close /></el-icon>
            </el-button>
          </div>
        </div>
      </template>

      <div class="dialog-content">
        <!-- 左侧：提示词模板展示区 -->
        <div class="prompt-template-area">
          <div class="template-header">
            <el-select
              v-model="selectedGeneratorCardPacks"
              multiple
              size="small"
              placeholder="选择卡包"
              class="card-pack-select"
              :max-collapse-tags="1"
              @change="onCardPackSelectionChange"
            >
              <el-option
                v-for="pack in availableCardPacks"
                :key="pack"
                :label="pack"
                :value="pack"
              />
            </el-select>
          </div>
          <div class="template-list">
            <div
              v-for="prompt in pinnedPrompts"
              :key="prompt.id"
              class="template-item"
              :class="{ active: prompt.enabled }"
            >
              <el-switch
                v-model="prompt.enabled"
                size="small"
                class="template-item-switch"
                @click.stop
              />
              <span class="template-item-name" @click="togglePinnedPrompt(prompt.id)">{{ prompt.name }}</span>
              <el-icon class="template-item-remove" @click.stop="removePinnedPrompt(prompt.id)"><Close /></el-icon>
            </div>
            <div v-if="pinnedPrompts.length === 0" class="empty-template">
              <el-icon :size="32"><DocumentRemove /></el-icon>
              <p>点击"详情"按钮选择提示词模板</p>
            </div>
          </div>
          <div class="template-footer">
            <el-button
              type="success"
              size="large"
              class="more-prompt-btn"
              @click="promptSelectDialogVisible = true"
            >
              更多提示词
            </el-button>
          </div>
        </div>

        <!-- 右侧：表单区 -->
        <div class="form-area">
          <!-- AI模型 -->
          <div class="form-block">
            <div class="form-block-label">
              <span>AI模型</span>
              <el-button
                type="primary"
                link
                size="small"
                class="form-block-action"
                @click="showFixedPrompt"
              >
                <el-icon><View /></el-icon>
                <span>查看完整提示词</span>
              </el-button>
            </div>
            <el-select
              v-model="selectedModel"
              placeholder="思考者"
              size="large"
              class="full-width"
              popper-class="form-area-select-popper"
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

          <!-- 提示词模板选择 -->
          <div class="form-block">
            <div class="form-block-label">提示词</div>
            <div class="prompt-tab-group">
              <el-button
                :type="promptTabMode === 'custom' ? 'success' : 'default'"
                round
                size="default"
                @click="promptTabMode = promptTabMode === 'custom' ? 'template' : 'custom'"
              >
                自定义
              </el-button>
            </div>
            <!-- 模板选择模式 -->
            <el-select
              v-if="promptTabMode !== 'custom'"
              v-model="selectedTemplateId"
              clearable
              placeholder="请选择提示词模板"
              size="large"
              class="full-width prompt-template-select"
            >
              <el-option
                v-for="prompt in pinnedPrompts"
                :key="prompt.id"
                :label="prompt.name"
                :value="prompt.id"
              />
            </el-select>
            <!-- 自定义模式 -->
            <div v-else class="custom-prompt-container">
              <el-input
                v-model="customPromptContent"
                type="textarea"
                :autosize="{ minRows: 4, maxRows: 10 }"
                placeholder="请输入自定义提示词内容..."
                maxlength="2000"
                show-word-limit
                resize="none"
              />
            </div>
            <div v-if="currentTemplateDescription && promptTabMode !== 'custom'" class="form-tip">
              <span class="form-tip-text">使用方法：</span>
              <span class="form-tip-content">{{ currentTemplateDescription }}</span>
              <a class="form-tip-link form-tip-link-strong" @click="showFixedPrompt">查看介绍</a>
            </div>
          </div>

          <!-- 任务参数：动态字段 -->
          <template v-if="hasAnyFields">
            <div
              v-for="field in allFields"
              :key="field.name"
              class="form-block"
            >
              <div class="form-block-label">
                {{ field.label }}
                <span
                  v-if="field.required !== false"
                  class="required-mark"
                  title="必填"
                >*</span>
              </div>
              <div v-if="field.description" class="field-description">{{ field.description }}</div>

              <div v-if="field.type === 'text' || field.type === 'textarea'" class="input-container">
                <el-input
                  v-if="field.type === 'text'"
                  v-model="fieldValues[field.name]"
                  type="textarea"
                  :autosize="{ minRows: 1, maxRows: 6 }"
                  :placeholder="`请输入${field.label}`"
                  maxlength="500"
                  show-word-limit
                  resize="none"
                />

                <el-input
                  v-else-if="field.type === 'textarea'"
                  v-model="fieldValues[field.name]"
                  type="textarea"
                  :autosize="{ minRows: 2, maxRows: 8 }"
                  :placeholder="`请输入${field.label}`"
                  maxlength="500"
                  show-word-limit
                  resize="none"
                />

                <el-button
                  class="fullscreen-btn"
                  circle
                  size="small"
                  @click="openFieldFullscreenEditor(field.name, field.label)"
                  title="全屏编辑"
                >
                  <el-icon><FullScreen /></el-icon>
                </el-button>
              </div>

              <el-select
                v-else-if="field.type === 'select'"
                v-model="fieldValues[field.name]"
                :placeholder="`请选择${field.label}`"
                class="full-width"
                size="large"
              >
                <el-option
                  v-for="(option, optionIndex) in field.options"
                  :key="option"
                  :label="field.optionLabels?.[optionIndex] || option"
                  :value="option"
                />
              </el-select>
            </div>
          </template>

          <!-- 补充信息 -->
          <div class="form-block">
            <div class="form-block-label">补充信息</div>
            <el-input
              v-model="additionalInfo"
              type="textarea"
              :rows="3"
              maxlength="500"
              show-word-limit
              placeholder=""
              resize="none"
            />
          </div>

          <!-- 关联作品 -->
          <div class="form-block">
            <div class="form-block-label">关联作品(可选)</div>
            <el-select
              v-model="selectedBookId"
              clearable
              filterable
              placeholder="选择一个作品以关联该作品的章节或备忘录"
              class="full-width"
              size="large"
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
          <div v-if="selectedBookId" class="form-block">
            <div class="form-block-label">关联章节</div>
            <el-select
              v-model="selectedChapterIds"
              multiple
              collapse-tags
              collapse-tags-tooltip
              placeholder="选择要关联的章节"
              class="full-width"
              size="large"
            >
              <el-option
                v-for="chapter in availableChapters"
                :key="chapter.id"
                :label="chapter.title"
                :value="chapter.id"
              />
            </el-select>
          </div>

          <!-- 关联备忘录 -->
          <div v-if="selectedBookId" class="form-block">
            <div class="form-block-label">关联备忘录</div>
            <el-select
              v-model="selectedMemoId"
              clearable
              filterable
              placeholder="选择一个备忘录"
              class="full-width"
              size="large"
            >
              <el-option
                v-for="memo in memos"
                :key="memo.id"
                :label="memo.title || memo.content?.slice(0, 30) + '...'"
                :value="memo.id"
              />
            </el-select>
          </div>

          <!-- 关联角色卡 -->
          <div v-if="selectedBookId" class="form-block">
            <div class="form-block-label">关联角色卡</div>
            <el-select
              v-model="selectedCharacterIds"
              multiple
              collapse-tags
              collapse-tags-tooltip
              placeholder="选择要关联的角色卡"
              class="full-width"
              size="large"
            >
              <el-option
                v-for="character in availableCharacters"
                :key="character.id"
                :label="character.name"
                :value="character.id"
              />
            </el-select>
          </div>

          <!-- 关联内容（已关联列表） -->
          <div v-if="hasRelatedContext" class="form-block">
            <div class="form-block-label">已关联内容</div>
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

      <!-- 自定义底部 -->
      <template #footer>
        <div class="gen-dialog-footer">
          <div class="gen-footer-disclaimer">以上内容均由AI生成,仅供参考和借鉴</div>
          <div class="gen-footer-actions">
            <el-button
              v-if="!generating"
              type="success"
              size="large"
              round
              :disabled="!canGenerate"
              class="gen-generate-btn"
              @click="generateContent"
            >
              <el-icon><Promotion /></el-icon>
              <span>生成</span>
            </el-button>
            <el-button
              v-else
              type="danger"
              size="large"
              round
              class="gen-generate-btn"
              @click="stopGenerate"
            >
              <el-icon><Close /></el-icon>
              <span>停止生成</span>
            </el-button>
          </div>
        </div>
      </template>
    </el-dialog>

    <!-- 动态字段全屏编辑弹窗 -->
    <el-dialog
      v-model="fieldFullscreenDialogVisible"
      :title="`全屏编辑 - ${fieldFullscreenFieldLabel}`"
      width="80%"
      top="5vh"
      align-center
      append-to-body
      destroy-on-close
      class="field-fullscreen-dialog"
    >
      <div class="fullscreen-editor-container">
        <el-input
          v-model="fieldFullscreenContent"
          type="textarea"
          :rows="20"
          :placeholder="`请输入${fieldFullscreenFieldLabel}`"
          resize="vertical"
          maxlength="10000"
          show-word-limit
        />
        <div class="fullscreen-editor-footer">
          <span class="char-count">{{ fieldFullscreenContent.length }} / 10000</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="cancelFieldFullscreenEditor">取消</el-button>
        <el-button type="primary" @click="saveFieldFullscreenEditor">确认</el-button>
      </template>
    </el-dialog>

    <!-- 提示词选择弹窗 -->
    <el-dialog
      v-model="promptSelectDialogVisible"
      width="800px"
      destroy-on-close
      class="prompt-select-dialog"
      modal-class="prompt-select-dialog-modal"
      top="5vh"
      append-to-body
      :show-close="false"
    >
      <!-- 自定义标题栏 -->
      <template #header>
        <div class="ps-dialog-header">
          <div class="ps-dialog-title">
            <el-icon class="ps-title-icon"><CollectionTag /></el-icon>
            <span>{{ promptSelectDialogTitle }}</span>
          </div>
          <el-button
            type="primary"
            link
            size="small"
            class="ps-close-btn"
            @click="promptSelectDialogVisible = false"
          >
            <el-icon :size="18"><Close /></el-icon>
          </el-button>
        </div>
      </template>
      <div class="prompt-select-content">
        <!-- 第一步：顶部标签导航 -->
        <div class="tab-navigation">
          <div
            v-for="tab in promptTabs"
            :key="tab.key"
            class="tab-item"
            :class="{ active: activePromptTab === tab.key }"
            @click="activePromptTab = tab.key"
          >
            {{ tab.label }}
          </div>
        </div>

        <!-- 搜索、筛选、排序区域 - 仅在"全部"和"已固定"标签页显示 -->
        <template v-if="activePromptTab === 'all' || activePromptTab === 'pinned'">
          <!-- 第二步：重新设计的搜索区域 -->
          <div class="search-section">
            <div class="search-box">
              <el-input
                v-model="promptSearchKeyword"
                placeholder="搜索提示词..."
                clearable
                class="search-input-field"
                @keyup.enter="handleSearchPrompts"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
              </el-input>
              <button class="search-btn" @click="handleSearchPrompts">
                <el-icon><Search /></el-icon>
                搜索
              </button>
            </div>
          </div>

          <!-- 第三步：优化分类筛选 - 下拉选择器 + 标签云 -->
          <div class="filter-section">
            <div class="filter-header" @click="toggleCategoryDropdown">
              <el-icon><Filter /></el-icon>
              <span>分类筛选</span>
              <el-icon class="filter-arrow" :class="{ expanded: showCategoryDropdown }"><ArrowDown /></el-icon>
            </div>
            
            <div v-if="showCategoryDropdown" class="category-dropdown">
              <div
                v-for="category in promptCategories"
                :key="category"
                class="dropdown-item"
                :class="{ 
                  selected: selectedPromptCategory === category || (category === '全部' && selectedPromptCategory === 'all')
                }"
                @click="selectCategory(category)"
              >
                {{ category }}
              </div>
            </div>

            <div v-if="selectedPromptCategory !== 'all' && selectedPromptCategory !== '全部'" class="selected-tags">
              <span class="tag-label">当前选择：</span>
              <el-tag 
                closable 
                type="success" 
                size="small"
                @close="clearCategorySelection"
              >
                {{ selectedPromptCategory === 'all' ? '全部' : selectedPromptCategory }}
              </el-tag>
            </div>
          </div>

          <!-- 第四步：排序选项栏 -->
          <div class="sort-section">
            <div class="sort-label">
              <el-icon><Sort /></el-icon>
              <span>排序方式</span>
            </div>
            <div class="sort-options">
              <button
                v-for="option in sortOptions"
                :key="option.key"
                class="sort-btn"
                :class="{ active: activeSortOption === option.key }"
                @click="activeSortOption = option.key"
              >
                <el-icon v-if="option.icon"><component :is="option.icon" /></el-icon>
                {{ option.label }}
              </button>
            </div>
          </div>
        </template>

        <!-- 第五步+第六步：根据标签显示不同内容 -->
        
        <!-- 全部标签页：显示提示词列表 -->
        <template v-if="activePromptTab === 'all'">
          <div class="prompt-select-list">
            <div
              v-for="prompt in sortedAndFilteredPrompts"
              :key="prompt.id"
              class="prompt-card-item"
              :class="{ selected: isPromptPinned(prompt.id) }"
            >
              <div class="card-main" @click="togglePromptPin(prompt)">
                <div class="card-top-row">
                  <div class="author-avatar" :style="{ background: getAvatarColor(prompt.author || prompt.name) }">
                    {{ (prompt.author || '匿').charAt(0) }}
                  </div>
                  <div class="card-info">
                    <div class="card-title-row">
                      <span class="card-title">{{ prompt.name }}</span>
                    </div>
                    <div class="card-meta-row">
                      <span class="author-name">{{ prompt.author || '匿名用户' }}</span>
                      <el-tag
                        v-if="prompt.is_featured"
                        size="small"
                        class="card-featured-tag"
                      >
                        精选
                      </el-tag>
                      <span class="author-stats">
                        <span class="stat-item">
                          <el-icon class="fire-icon"><Sunny /></el-icon>
                          {{ formatUsageCount(prompt.usage_count) }}
                        </span>
                      </span>
                      <span class="card-date">{{ formatPromptTime(prompt.created_at) }}</span>
                    </div>
                    <div class="card-tags">
                      <el-tag size="small" class="card-category-tag">{{ prompt.category }}</el-tag>
                      <el-tag size="small" class="card-version-tag">{{ prompt.version || 'v1.0' }}</el-tag>
                    </div>
                  </div>
                </div>
                
                <div class="card-content">
                  <p class="card-desc">{{ prompt.description || '1' }}</p>
                </div>
              </div>

              <div class="card-actions" @click.stop>
                <button
                  v-if="isPromptPinned(prompt.id)"
                  class="action-btn action-btn--remove"
                  @click="removePinnedPrompt(prompt.id)"
                >
                  <el-icon><Close /></el-icon>
                  删除
                </button>
                <button
                  v-else
                  class="action-btn action-btn--add"
                  @click="togglePromptPin(prompt)"
                >
                  <el-icon><Plus /></el-icon>
                  添加
                </button>
              </div>
            </div>
            
            <div v-if="sortedAndFilteredPrompts.length === 0" class="empty-prompts">
              <el-icon :size="48"><Document /></el-icon>
              <p>暂无匹配的提示词</p>
              <p class="empty-hint">尝试调整筛选条件或搜索关键词</p>
            </div>
          </div>
        </template>
        
        <!-- 已固定标签页：显示已固定的提示词 -->
        <template v-else-if="activePromptTab === 'pinned'">
          <div class="prompt-select-list">
            <div
              v-for="prompt in pinnedPrompts"
              :key="prompt.id"
              class="prompt-card-item selected"
            >
              <div class="card-main">
                <div class="card-top-row">
                  <div class="author-avatar" :style="{ background: getAvatarColor(prompt.author || prompt.name) }">
                    {{ (prompt.author || '匿').charAt(0) }}
                  </div>
                  <div class="card-info">
                    <div class="card-title-row">
                      <span class="card-title">{{ prompt.name }}</span>
                    </div>
                    <div class="card-meta-row">
                      <span class="author-name">{{ prompt.author || '匿名用户' }}</span>
                      <el-tag
                        v-if="prompt.is_featured"
                        size="small"
                        class="card-featured-tag"
                      >
                        精选
                      </el-tag>
                      <span class="author-stats">
                        <span class="stat-item">
                          <el-icon class="fire-icon"><Sunny /></el-icon>
                          {{ formatUsageCount(prompt.usage_count) }}
                        </span>
                      </span>
                      <span class="card-date">{{ formatPromptTime(prompt.created_at) }}</span>
                    </div>
                    <div class="card-tags">
                      <el-tag size="small" class="card-category-tag">{{ prompt.category }}</el-tag>
                      <el-tag size="small" class="card-version-tag">{{ prompt.version || 'v1.0' }}</el-tag>
                    </div>
                  </div>
                </div>
                
                <div class="card-content">
                  <p class="card-desc">{{ prompt.description || '1' }}</p>
                </div>
              </div>

              <div class="card-actions">
                <el-switch
                  v-model="prompt.enabled"
                  size="small"
                  active-text="启用"
                  inactive-text="禁用"
                  @change="savePinnedPromptsToStorage"
                />
                <button
                  class="action-btn action-btn--remove"
                  @click="removePinnedPrompt(prompt.id)"
                  style="margin-left: 8px;"
                >
                  <el-icon><Close /></el-icon>
                  移除
                </button>
              </div>
            </div>
            
            <div v-if="pinnedPrompts.length === 0" class="empty-prompts">
              <el-icon :size="48"><CollectionTag /></el-icon>
              <p>暂无固定的提示词</p>
              <p class="empty-hint">在"全部"标签页中添加提示词</p>
            </div>
          </div>
        </template>
        
        <!-- 已创建标签页：创建新提示词 -->
        <template v-else-if="activePromptTab === 'created'">
          <div class="create-prompt-section">
            <div class="create-prompt-form">
              <div class="form-row">
                <div class="form-group">
                  <label class="form-label">提示词名称</label>
                  <el-input
                    v-model="newPromptForm.name"
                    placeholder="输入提示词名称..."
                    clearable
                  />
                </div>
                
                <div class="form-group">
                  <label class="form-label">分类</label>
                  <el-select
                    v-model="newPromptForm.category"
                    placeholder="选择分类"
                    style="width: 100%"
                    allow-create
                    filterable
                  >
                    <el-option
                      v-for="cat in promptCategories.filter(c => c !== '全部')"
                      :key="cat"
                      :label="cat"
                      :value="cat"
                    />
                  </el-select>
                </div>
              </div>
              
              <div class="form-row">
                <div class="form-group">
                  <label class="form-label">作者</label>
                  <el-input
                    v-model="newPromptForm.author"
                    placeholder="输入作者名称..."
                    clearable
                  />
                </div>
                
                <div class="form-group">
                  <label class="form-label">版本号</label>
                  <el-input
                    v-model="newPromptForm.version"
                    placeholder="如: v1.0"
                    clearable
                  />
                </div>
              </div>
              
              <div class="form-group">
                <label class="form-label">提示词内容</label>
                <el-input
                  v-model="newPromptForm.content"
                  type="textarea"
                  :rows="8"
                  placeholder="输入提示词内容，可使用 {{变量名}} 格式定义变量..."
                />
              </div>
              
              <div class="form-group">
                <label class="form-label">变量字段（可选）</label>
                <div class="fields-config">
                  <div
                    v-for="(field, index) in newPromptForm.fields"
                    :key="index"
                    class="field-item"
                  >
                    <el-input v-model="field.name" placeholder="变量名" style="width: 100px" />
                    <el-input v-model="field.label" placeholder="显示名称" style="flex: 1" />
                    <el-select v-model="field.type" style="width: 100px">
                      <el-option label="文本" value="text" />
                      <el-option label="多行" value="textarea" />
                      <el-option label="选择" value="select" />
                    </el-select>
                    <el-button type="danger" link @click="newPromptForm.fields.splice(index, 1)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                  <el-button type="primary" link @click="addNewPromptField">
                    <el-icon><Plus /></el-icon>
                    添加字段
                  </el-button>
                </div>
              </div>
              
              <div class="form-actions">
                <el-button @click="resetNewPromptForm">重置</el-button>
                <el-button type="primary" @click="saveNewPrompt" :disabled="!canSaveNewPrompt">
                  <el-icon><Check /></el-icon>
                  保存提示词
                </el-button>
              </div>
            </div>
          </div>
        </template>
      </div>
    </el-dialog>

    <!-- 预览弹窗 -->
    <el-dialog
      v-model="previewDialogVisible"
      title="生成结果预览"
      width="50vw"
      destroy-on-close
      class="preview-dialog"
      modal-class="preview-dialog-modal"
      append-to-body
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
          <el-button
            v-if="!generating && previewContent"
            type="warning"
            link
            @click="openFollowUp"
          >
            <el-icon><ChatDotRound /></el-icon>
            追问
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

    <!-- 追问对话弹窗 -->
    <el-dialog
      v-model="followUpDialogVisible"
      width="800px"
      class="followup-dialog"
      modal-class="followup-dialog-modal"
      destroy-on-close
      append-to-body
    >
      <template #header>
        <div class="followup-dialog-header">
          <div class="header-title">
            <el-icon><ChatDotRound /></el-icon>
            <span>追问对话</span>
          </div>
          <el-select
            v-model="followUpSelectedModelId"
            placeholder="选择模型"
            class="model-select"
            size="small"
          >
            <el-option
              v-for="model in models"
              :key="model.id"
              :label="model.name"
              :value="model.id"
            >
              <div class="model-option-content">
                <span>{{ model.name }}</span>
                <el-tag v-if="model.provider_name" size="small" type="info">{{ model.provider_name }}</el-tag>
              </div>
            </el-option>
          </el-select>
        </div>
      </template>
      <div class="followup-dialog-content">
        <div class="followup-messages" ref="followUpMessagesContainer">
          <div v-if="followUpMessages.length > 2 && !followUpHistoryExpanded" class="history-collapsed-hint">
            <el-button
              text
              type="primary"
              @click="followUpHistoryExpanded = true"
              size="small"
              class="expand-btn"
            >
              <el-icon><ArrowDown /></el-icon>
              展开 {{ followUpMessages.length - 2 }} 条历史消息
            </el-button>
          </div>

          <template v-if="followUpHistoryExpanded">
            <transition-group name="message-slide">
              <div
                v-for="(msg, msgIndex) in followUpMessages"
                :key="msgIndex"
                :class="['followup-message-bubble', msg.role, { 'history-message': msgIndex < followUpMessages.length - 2 }]"
              >
                <div class="followup-message-sender">
                  <el-icon class="sender-icon">
                    <User v-if="msg.role === 'user'" />
                    <ChatDotRound v-if="msg.role === 'assistant'" />
                    <Setting v-if="msg.role === 'system'" />
                  </el-icon>
                  <span class="sender-name">
                    {{ msg.role === 'user' ? '用户' : msg.role === 'assistant' ? 'AI' : '系统' }}
                  </span>
                </div>
                <div class="followup-message-content">
                  <MarkdownRenderer :content="msg.content" />
                </div>
              </div>
            </transition-group>
          </template>

          <div v-if="followUpGenerating" class="followup-message-bubble assistant">
            <div class="followup-message-sender">
              <el-icon class="sender-icon"><ChatDotRound /></el-icon>
              <span class="sender-name">AI</span>
            </div>
            <div class="followup-message-content">
              <div class="typing-indicator">
                <span></span>
                <span></span>
                <span></span>
              </div>
            </div>
          </div>
        </div>

        <div class="followup-input-area">
          <el-input
            v-model="followUpUserInput"
            type="textarea"
            :rows="2"
            placeholder="输入您的问题，按 Ctrl+Enter 发送..."
            @keydown.enter.ctrl="sendFollowUpMessage"
            :disabled="followUpGenerating"
            resize="none"
            class="followup-input"
          />
          <div class="followup-input-actions">
            <el-tooltip content="发送 (Ctrl+Enter)" placement="top">
              <el-button
                type="primary"
                @click="sendFollowUpMessage"
                :loading="followUpGenerating"
                :disabled="!followUpUserInput.trim()"
                :icon="ChatDotRound"
                circle
                class="send-btn"
              />
            </el-tooltip>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 系统提示词修改弹窗 -->
    <el-dialog
      v-model="promptEditDialogVisible"
      title="修改系统提示词"
      width="600px"
      destroy-on-close
      modal-class="prompt-edit-dialog-modal"
      append-to-body
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
      append-to-body
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
      width="960px"
      destroy-on-close
      class="generator-manager-dialog"
      modal-class="generator-manager-dialog-modal"
      append-to-body
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
              v-for="(gen, index) in allGenerators"
              :key="gen.id"
              class="generator-item"
              :class="{ active: selectedGenerator?.id === gen.id }"
              @click="selectGenerator(gen)"
            >
              <div class="generator-item-icon">
                <el-icon :size="18">
                  <component :is="getIconComponent(gen.icon)" />
                </el-icon>
              </div>
              <div class="generator-item-info">
                <span class="generator-item-name">
                  {{ gen.name }}
                  <el-tag v-if="gen.isDefault" size="small" type="info" style="margin-left: 4px;">默认</el-tag>
                </span>
                <span class="generator-item-desc">{{ gen.description?.slice(0, 30) }}{{ gen.description?.length > 30 ? '...' : '' }}</span>
              </div>
              <div class="generator-item-actions" @click.stop>
                <el-button
                  type="primary"
                  link
                  size="small"
                  :disabled="index === 0"
                  @click="moveGeneratorUp(index)"
                >
                  <el-icon><ArrowUp /></el-icon>
                </el-button>
                <el-button
                  type="primary"
                  link
                  size="small"
                  :disabled="index === allGenerators.length - 1"
                  @click="moveGeneratorDown(index)"
                >
                  <el-icon><ArrowDown /></el-icon>
                </el-button>
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
              <div class="detail-label collapsible-label" @click="iconPickerExpanded = !iconPickerExpanded">
                <span>选择图标</span>
                <el-icon class="collapse-arrow" :class="{ expanded: iconPickerExpanded }"><ArrowDown /></el-icon>
              </div>
              <div v-show="iconPickerExpanded" class="icon-picker-grid">
                <div
                  v-for="icon in availableIcons"
                  :key="icon.name"
                  class="icon-picker-item"
                  :class="{ selected: selectedGenerator.icon === icon.name }"
                  @click="selectedGenerator.icon = icon.name"
                >
                  <el-icon :size="20">
                    <component :is="icon.component" />
                  </el-icon>
                  <span class="icon-picker-name">{{ icon.name }}</span>
                </div>
              </div>
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
      title="抽卡区历史记录"
      width="900px"
      destroy-on-close
      class="history-dialog"
      modal-class="history-dialog-modal"
      append-to-body
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
      append-to-body
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
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowRight,
  ArrowDown,
  ArrowUp,
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
  ChatDotRound,
  MagicStick,
  Reading,
  Tickets,
  FullScreen,
  CollectionTag,
  OfficeBuilding,
  Star,
  Sunny,
  Moon,
  Grid,
  Picture,
  Van,
  Brush,
  TrophyBase,
  Medal,
  DataLine,
  TrendCharts,
  Promotion,
  Clock,
  Timer,
  Switch,
  Key,
  Coin,
  Histogram,
  PieChart,
  DocumentRemove,
  InfoFilled
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
  if (models.value.length === 0) return ''
  const defaultModel = models.value.find(model => model.is_default === 1 || model.is_default === true)
  return defaultModel?.id ?? ''
}

// 加载模型列表
const fetchModels = async () => {
  try {
    const res = await configAPI.getAll()
    if (res.success && res.data) {
      models.value = res.data.filter(m => m.enabled !== 0)
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
const promptTabMode = ref<'template' | 'custom'>('template')
const customPromptContent = ref('')
const selectedTemplateId = ref<number | null>(null)
const selectedBookId = ref<number | null>(null)
const selectedChapterIds = ref<number[]>([])
const selectedMemoId = ref<number | null>(null)
const selectedCharacterIds = ref<number[]>([])
const generating = ref(false)
const generatedContent = ref('')
const abortController = ref<AbortController | null>(null)
const isEditingPreview = ref(false)

type PromptItem = {
  id: number
  name: string
  category: string
  content: string
  description?: string
  author?: string
  version?: string
  fields?: Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; optionLabels?: string[]; description: string; required: boolean }>
  usage_count?: number
  created_at?: string
  is_featured?: boolean
}

const prompts = ref<PromptItem[]>([])

// 固定提示词相关状态
const pinnedPrompts = ref<Array<PromptItem & { enabled: boolean }>>([])
const promptSelectDialogVisible = ref(false)
const promptSearchKeyword = ref('')
const selectedPromptCategory = ref('all')
const selectedGeneratorCardPacks = ref<string[]>([])
const categoryTabsRef = ref<HTMLElement | null>(null)
const isDragging = ref(false)
const startX = ref(0)
const scrollLeft = ref(0)

// 新增：标签导航相关
const activePromptTab = ref('all') // all: 全部, pinned: 已固定, created: 已创建
const promptTabs = [
  { key: 'all', label: '全部' },
  { key: 'pinned', label: '已固定' },
  { key: 'created', label: '已创建' }
]

// 新增：创建提示词表单
const newPromptForm = ref({
  name: '',
  category: '',
  author: '',
  version: '',
  content: '',
  fields: [] as Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; optionLabels?: string[]; description: string; required: boolean }>
})

const canSaveNewPrompt = computed(() => {
  return newPromptForm.value.name.trim() !== '' && 
         newPromptForm.value.category !== '' && 
         newPromptForm.value.content.trim() !== ''
})

const addNewPromptField = () => {
  newPromptForm.value.fields.push({
    name: '',
    label: '',
    type: 'text',
    options: [],
    description: '',
    required: false
  })
}

const resetNewPromptForm = () => {
  newPromptForm.value = {
    name: '',
    category: '',
    author: '',
    version: '',
    content: '',
    fields: []
  }
}

const saveNewPrompt = () => {
  if (!canSaveNewPrompt.value) return
  
  const newPrompt: PromptItem = {
    id: Date.now(),
    name: newPromptForm.value.name,
    category: newPromptForm.value.category,
    description: newPromptForm.value.content.slice(0, 50),
    content: newPromptForm.value.content,
    author: newPromptForm.value.author || '匿名',
    version: newPromptForm.value.version || 'v1.0',
    fields: newPromptForm.value.fields.length > 0 ? newPromptForm.value.fields : undefined,
    created_at: new Date().toISOString(),
    usage_count: 0
  }
  
  prompts.value.unshift(newPrompt)
  savePromptsToStorage()
  resetNewPromptForm()
  ElMessage.success('提示词创建成功！')
}

const savePromptsToStorage = () => {
  if (currentGenerator.value) {
    const key = `prompts_${currentGenerator.value.id}`
    localStorage.setItem(key, JSON.stringify(prompts.value))
  }
}

// 新增：分类下拉选择器相关
const showCategoryDropdown = ref(false)

// 新增：搜索和分类筛选方法
const handleSearchPrompts = () => {
  // 搜索逻辑已在 computed 中实现
}

const toggleCategoryDropdown = () => {
  showCategoryDropdown.value = !showCategoryDropdown.value
}

const selectCategory = (category: string) => {
  selectedPromptCategory.value = category === '全部' ? 'all' : category
  showCategoryDropdown.value = false
}

const clearCategorySelection = () => {
  selectedPromptCategory.value = 'all'
}

// 第四步：排序功能相关
const activeSortOption = ref('latest') // latest: 最新, hot: 最热, monthly: 月榜, featured: 精选, search: 搜索
const sortOptions = [
  { key: 'hot', label: '最热', icon: 'TrendCharts' },
  { key: 'monthly', label: '月榜', icon: 'Calendar' },
  { key: 'featured', label: '精选', icon: 'Star' },
  { key: 'latest', label: '最新', icon: 'Clock' },
  { key: 'search', label: '搜索', icon: 'Search' }
]

// 第五步+第六步：格式化方法和计算属性
const formatPromptTime = (timeStr?: string) => {
  const date = timeStr ? new Date(timeStr) : new Date()
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const formatUsageCount = (count?: number) => {
  if (count === undefined || count === null) return '0'
  if (count < 1000) return count.toString()
  if (count < 1_000_000) return `${(count / 1000).toFixed(1)}K`
  return `${(count / 1_000_000).toFixed(1)}M`
}

const getAvatarColor = (name: string) => {
  const colors = [
    'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
    'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
    'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
    'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
    'linear-gradient(135deg, #a8edea 0%, #fed6e3 100%)',
    'linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%)',
    'linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%)',
    'linear-gradient(135deg, #a1c4fd 0%, #c2e9fb 100%)',
    'linear-gradient(135deg, #d299c2 0%, #fef9d7 100%)',
    'linear-gradient(135deg, #89f7fe 0%, #66a6ff 100%)',
    'linear-gradient(135deg, #cd9cf2 0%, #f6f3ff 100%)',
  ]
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
}

// 排序和过滤后的提示词列表
const sortedAndFilteredPrompts = computed(() => {
  let result = [...filteredPrompts.value]
  
  // 根据排序选项排序
  switch (activeSortOption.value) {
    case 'hot':
      // 按使用次数降序（假设有 usage_count 字段）
      result.sort((a, b) => (b.usage_count || 0) - (a.usage_count || 0))
      break
    case 'monthly':
      // 本月热门（简化实现：按时间倒序）
      const thisMonth = new Date()
      const monthAgo = new Date(thisMonth.getTime() - 30 * 24 * 60 * 60 * 1000)
      result.sort((a, b) => {
        const aInMonth = new Date(a.created_at || '') > monthAgo ? 1 : 0
        const bInMonth = new Date(b.created_at || '') > monthAgo ? 1 : 0
        return bInMonth - aInMonth || (b.usage_count || 0) - (a.usage_count || 0)
      })
      break
    case 'featured':
      // 精选（标记为 is_featured 的优先）
      result.sort((a, b) => (b.is_featured ? 1 : 0) - (a.is_featured ? 1 : 0))
      break
    case 'search':
      // 搜索相关度排序（如果有搜索关键词，按匹配度排序）
      if (promptSearchKeyword.value.trim()) {
        const keyword = promptSearchKeyword.value.toLowerCase()
        result.sort((a, b) => {
          const aScore = calculateRelevanceScore(a, keyword)
          const bScore = calculateRelevanceScore(b, keyword)
          return bScore - aScore
        })
      }
      break
    case 'latest':
    default:
      // 默认按时间倒序
      result.sort((a, b) => (b.created_at || '').localeCompare(a.created_at || ''))
      break
  }
  
  return result
})

// 计算搜索相关度得分
const calculateRelevanceScore = (prompt: PromptItem, keyword: string): number => {
  let score = 0
  
  // 名称完全匹配
  if (prompt.name.toLowerCase().includes(keyword)) score += 10
  
  // 分类匹配
  if (prompt.category && prompt.category.toLowerCase().includes(keyword)) score += 5
  
  // 内容匹配
  if (prompt.content && prompt.content.toLowerCase().includes(keyword)) score += 3
  
  // 名称开头匹配加分更多
  if (prompt.name.toLowerCase().startsWith(keyword)) score += 5
  
  return score
}

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
const currentConversationMessages = ref<HistoryMessage[]>([])

// 追问对话状态
const followUpDialogVisible = ref(false)
const followUpMessages = ref<HistoryMessage[]>([])
const followUpUserInput = ref('')
const followUpGenerating = ref(false)
const followUpMessagesContainer = ref<HTMLElement | null>(null)
const followUpHistoryExpanded = ref(false)
const followUpSelectedModelId = ref<number | string>('')

// 字段值存储
const fieldValues = ref<Record<string, string>>({})

// 动态字段全屏编辑相关状态
const fieldFullscreenDialogVisible = ref(false)
const fieldFullscreenContent = ref('')
const fieldFullscreenFieldName = ref('')
const fieldFullscreenFieldLabel = ref('')

const openFieldFullscreenEditor = (fieldName: string, fieldLabel: string) => {
  fieldFullscreenFieldName.value = fieldName
  fieldFullscreenFieldLabel.value = fieldLabel
  fieldFullscreenContent.value = fieldValues.value[fieldName] || ''
  fieldFullscreenDialogVisible.value = true
}

const saveFieldFullscreenEditor = () => {
  fieldValues.value[fieldFullscreenFieldName.value] = fieldFullscreenContent.value
  fieldFullscreenDialogVisible.value = false
  fieldFullscreenFieldName.value = ''
  fieldFullscreenFieldLabel.value = ''
  fieldFullscreenContent.value = ''
}

const cancelFieldFullscreenEditor = () => {
  fieldFullscreenDialogVisible.value = false
  fieldFullscreenFieldName.value = ''
  fieldFullscreenFieldLabel.value = ''
  fieldFullscreenContent.value = ''
}

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
  optionLabels?: string[];
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
const iconPickerExpanded = ref(false)

// 图标映射表：将图标名称映射到 Element Plus 图标组件
const iconMap: Record<string, any> = {
  MagicStick,
  Document,
  EditPen,
  Reading,
  Tickets,
  Lightning,
  Cpu,
  CollectionTag,
  User,
  OfficeBuilding,
  Star,
  Sunny,
  Moon,
  Picture,
  Timer,
  TrophyBase,
  Medal,
  Coin,
  DataLine,
  TrendCharts,
  Promotion,
  Clock,
  Brush,
  Setting,
  Grid
}

const getIconComponent = (iconName: string) => {
  return iconMap[iconName] || MagicStick
}

// 可选择的图标列表（用于图标选择器）
const availableIcons = computed(() => {
  return Object.keys(iconMap).map(name => ({
    name,
    component: iconMap[name]
  }))
})

// 计算属性
const selectedPromptInfo = computed(() => {
  return prompts.value.find(p => p.id === selectedPrompt.value)
})

const allFields = computed(() => {
  const fields: Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; optionLabels?: string[]; description: string; required: boolean }> = []
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

// 生成器与卡包的映射关系
const GENERATOR_TO_CATEGORY_MAP: Record<string, string> = {
  '脑洞生成器': '脑洞',
  '书名生成器': '书名',
  '简介生成器': '简介',
  '大纲生成器': '大纲',
  '细纲生成器': '细纲',
  '黄金开篇生成器': '黄金开篇',
  '金手指生成器': '金手指',
  '名字生成器': '名字',
  '人设生成器': '人设',
  '世界观生成器': '世界观'
}

const promptCategories = computed(() => {
  const categories = new Set<string>()
  
  // 如果当前打开了生成器，使用用户选择的卡包（支持1-2个）
  if (currentGenerator.value && selectedGeneratorCardPacks.value.length > 0) {
    const selectedPacks = selectedGeneratorCardPacks.value.filter(pack =>
      prompts.value.some(p => p.category === pack)
    )
    if (selectedPacks.length > 0) {
      return selectedPacks
    }
  }
  
  // 否则显示所有卡包
  prompts.value.forEach(p => {
    if (p.category) {
      categories.add(p.category)
    }
  })
  return ['全部', ...Array.from(categories)]
})

const filteredPrompts = computed(() => {
  let result = prompts.value
  
  // 如果当前打开了生成器且选择了卡包，只显示对应卡包的提示词
  if (currentGenerator.value && selectedGeneratorCardPacks.value.length > 0) {
    result = result.filter(p => selectedGeneratorCardPacks.value.includes(p.category))
  } else if (selectedPromptCategory.value !== 'all' && selectedPromptCategory.value !== '全部') {
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

const promptSelectDialogTitle = computed(() => {
  const base = currentGenerator.value?.name || '选择提示词'
  if (currentGenerator.value && selectedGeneratorCardPacks.value.length > 0) {
    const packNames = selectedGeneratorCardPacks.value.join('+')
    return `${base} · ${packNames}卡包`
  }
  return base
})

const additionalInfoLength = computed(() => additionalInfo.value.length)

const currentTemplateDescription = computed(() => {
  if (!selectedTemplateId.value) return ''
  const target = pinnedPrompts.value.find(p => p.id === selectedTemplateId.value)
  return target?.description || ''
})

const selectedBook = computed(() => {
  return books.value.find(book => book.id === selectedBookId.value) || null
})

const selectedChapters = computed(() => {
  return availableChapters.value.filter(chapter => selectedChapterIds.value.includes(chapter.id))
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

  selectedChapters.value.forEach(chapter => {
    items.push({
      type: 'chapter',
      id: chapter.id,
      title: chapter.title,
      content: chapter.content
    })
  })

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

  // 自定义模式：检查自定义提示词内容
  if (promptTabMode.value === 'custom') {
    const hasCustomPrompt = customPromptContent.value.trim() !== ''
    if (!hasCustomPrompt) return false
  } else {
    // 模板选择模式：检查固定提示词或已启用的模板
    const hasPromptFoundation = Boolean(editableFixedPrompt.value.trim()) || pinnedPrompts.value.some(p => p.enabled)
    if (!hasPromptFoundation) return false
  }

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

const togglePromptPin = (prompt: { id: number; name: string; category: string; content: string; fields?: Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; optionLabels?: string[]; description: string; required: boolean }> }) => {
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

const togglePinnedPrompt = (promptId: number) => {
  const target = pinnedPrompts.value.find(p => p.id === promptId)
  if (target) {
    target.enabled = !target.enabled
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

const getGeneratorCardPacksKey = (generatorId: number) => {
  return `generator_card_packs_${generatorId}`
}

const saveGeneratorCardPacksToStorage = () => {
  if (currentGenerator.value) {
    const key = getGeneratorCardPacksKey(currentGenerator.value.id)
    localStorage.setItem(key, JSON.stringify(selectedGeneratorCardPacks.value))
  }
}

const loadGeneratorCardPacksFromStorage = (generatorId: number): string[] => {
  const key = getGeneratorCardPacksKey(generatorId)
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

const onCardPackSelectionChange = (value: string[]) => {
  if (value.length > 2) {
    selectedGeneratorCardPacks.value = value.slice(0, 2)
    return
  }
  saveGeneratorCardPacksToStorage()
  selectedPromptCategory.value = 'all'
}

const availableCardPacks = computed(() => {
  const packs = new Set<string>()
  prompts.value.forEach(p => {
    if (p.category) {
      packs.add(p.category)
    }
  })
  return Array.from(packs)
})

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
      selectedChapterIds: selectedChapterIds.value,
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

  // 自定义模式：使用用户输入的自定义提示词
  if (promptTabMode.value === 'custom' && customPromptContent.value.trim()) {
    return [fixedContent, customPromptContent.value.trim()]
      .filter(Boolean)
      .join('\n\n')
  }

  // 模板选择模式：使用选中的模板
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
  selectedChapterIds.value = []
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

  const savedCardPacks = loadGeneratorCardPacksFromStorage(generator.id)
  if (savedCardPacks.length > 0) {
    selectedGeneratorCardPacks.value = savedCardPacks
  } else {
    const mappedCategory = GENERATOR_TO_CATEGORY_MAP[generator.name]
    selectedGeneratorCardPacks.value = mappedCategory ? [mappedCategory] : []
  }

  const savedFormState = loadFormStateFromStorage(generator.id)
  if (savedFormState) {
    const savedModelId = savedFormState.selectedModel
    selectedModel.value = (savedModelId && models.value.some(m => m.id === savedModelId)) ? savedModelId : getDefaultModelId()
    inputParams.value = savedFormState.inputParams || ''
    additionalInfo.value = savedFormState.additionalInfo || ''
    selectedBookId.value = savedFormState.selectedBookId
    selectedChapterIds.value = savedFormState.selectedChapterIds || []
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

watch(selectedGeneratorCardPacks, () => {
  saveGeneratorCardPacksToStorage()
}, { deep: true })

watch(editableFixedPrompt, (newVal, oldVal) => {
  if (newVal && newVal !== oldVal && !hasUserEditedFixedPrompt.value) {
    hasUserEditedFixedPrompt.value = true
  }
})

watch(selectedBookId, async (newBookId, oldBookId) => {
  if (newBookId === oldBookId) return

  selectedChapterIds.value = []
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
          title: item.title,
          content: item.content
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
      currentConversationMessages.value = historyRecord.messages
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

// 追问对话相关
const openFollowUp = () => {
  if (!currentConversationMessages.value.length) {
    ElMessage.warning('没有可追问的对话内容')
    return
  }
  followUpMessages.value = [...currentConversationMessages.value]
  followUpUserInput.value = ''
  followUpGenerating.value = false
  followUpHistoryExpanded.value = false
  followUpSelectedModelId.value = selectedModel.value
  followUpDialogVisible.value = true

  nextTick(() => {
    scrollFollowUpToBottom()
  })
}

const scrollFollowUpToBottom = () => {
  if (followUpMessagesContainer.value) {
    followUpMessagesContainer.value.scrollTop = followUpMessagesContainer.value.scrollHeight
  }
}

const sendFollowUpMessage = async () => {
  if (!followUpUserInput.value.trim() || followUpGenerating.value) return

  const userMessage: HistoryMessage = {
    role: 'user',
    content: followUpUserInput.value.trim(),
    timestamp: Date.now()
  }

  followUpMessages.value.push(userMessage)
  followUpHistoryExpanded.value = true
  const currentInput = followUpUserInput.value
  followUpUserInput.value = ''
  followUpGenerating.value = true

  nextTick(() => {
    scrollFollowUpToBottom()
  })

  try {
    const selectedModelData = models.value.find(m => m.id === followUpSelectedModelId.value)
    if (!selectedModelData) {
      ElMessage.error('请选择模型')
      followUpGenerating.value = false
      return
    }

    const validRoles = ['system', 'user', 'assistant']
    const messages = followUpMessages.value
      .filter(msg => validRoles.includes(msg.role))
      .map(msg => ({
        role: msg.role,
        content: msg.content
      }))

    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        messages,
        configId: selectedModelData.id
      })
    })

    if (!response.ok) {
      throw new Error('请求失败')
    }

    const reader = response.body?.getReader()
    const decoder = new TextDecoder()
    let assistantMessage = ''

    const assistantMsg: HistoryMessage = {
      role: 'assistant',
      content: '',
      timestamp: Date.now()
    }
    followUpMessages.value.push(assistantMsg)
    const reactiveAssistantMsg = followUpMessages.value[followUpMessages.value.length - 1]

    while (reader) {
      const { done, value } = await reader.read()
      if (done) break

      const chunk = decoder.decode(value)
      const lines = chunk.split('\n')

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.slice(6)
          if (data === '[DONE]') continue

          try {
            const parsed = JSON.parse(data)
            if (parsed.content) {
              assistantMessage += parsed.content
              reactiveAssistantMsg.content = assistantMessage
              nextTick(() => {
                scrollFollowUpToBottom()
              })
            } else if (parsed.error) {
              throw new Error(parsed.error)
            }
          } catch (e) {
            if (e instanceof SyntaxError) continue
            throw e
          }
        }
      }
    }

    ElMessage.success('对话完成')
  } catch (error: any) {
    ElMessage.error(error.message || '发送失败')
    followUpMessages.value.pop()
    followUpMessages.value.pop()
  } finally {
    followUpGenerating.value = false
  }
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
      const res = await generatorAPI.create({
        name: selectedGenerator.value.name,
        description: selectedGenerator.value.description,
        icon: selectedGenerator.value.icon || 'Lightning',
        core_prompt: selectedGenerator.value.core_prompt,
        remark: selectedGenerator.value.remark,
        order_num: selectedGenerator.value.order_num || 0
      })
      if (res.success) {
        const generatorName = selectedGenerator.value.name
        const generatorCorePrompt = selectedGenerator.value.core_prompt || ''
        
        try {
          const promptRes = await promptAPI.create({
            name: `${generatorName}默认提示词`,
            category: generatorName,
            content: generatorCorePrompt,
            description: `${generatorName}生成器的默认提示词模板`
          })
          if (promptRes?.data) {
            prompts.value.unshift({
              id: promptRes.data.id || Date.now(),
              name: `${generatorName}默认提示词`,
              category: generatorName,
              content: generatorCorePrompt,
              description: `${generatorName}生成器的默认提示词模板`,
              author: '系统',
              version: 'v1.0',
              created_at: new Date().toISOString(),
              usage_count: 0
            })
          }
        } catch (promptError) {
          console.warn('创建默认提示词失败:', promptError)
        }
        
        ElMessage.success('创建成功')
        await fetchGenerators()
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

const moveGeneratorUp = (index: number) => {
  if (index <= 0) return
  const generators = [...allGenerators.value]
  const item = generators[index]
  generators[index] = generators[index - 1]
  generators[index - 1] = item
  generators.forEach((g, i) => { g.order_num = i })
  allGenerators.value = generators
  saveGeneratorsOrder()
}

const moveGeneratorDown = (index: number) => {
  if (index >= allGenerators.value.length - 1) return
  const generators = [...allGenerators.value]
  const item = generators[index]
  generators[index] = generators[index + 1]
  generators[index + 1] = item
  generators.forEach((g, i) => { g.order_num = i })
  allGenerators.value = generators
  saveGeneratorsOrder()
}

const saveGeneratorsOrder = () => {
  const defaults = allGenerators.value.filter(g => g.isDefault)
  saveDefaultGenerators(defaults as GeneratorItem[])
}

const createDefaultGenerators = (): GeneratorItem[] =>
  defaultGeneratorConfigs.map((g, i) => ({ ...g, isDefault: true, order_num: i }))

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
      { id: 1, name: '角色设定生成', category: '角色', content: '', author: '系统', version: 'v1.0', description: '用于生成角色设定的提示词' },
      { id: 2, name: '剧情大纲生成', category: '剧情', content: '', author: '系统', version: 'v1.0', description: '用于生成剧情大纲的提示词' },
      { id: 3, name: '对话润色', category: '对话', content: '', author: '系统', version: 'v1.0', description: '用于润色对话内容的提示词' },
      { id: 4, name: '场景描写', category: '场景', content: '', author: '系统', version: 'v1.0', description: '用于生成场景描写的提示词' },
      { id: 5, name: '情节转折', category: '剧情', content: '', author: '系统', version: 'v1.0', description: '用于生成情节转折的提示词' }
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
  min-height: 100%;
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
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  width: 100%;
  position: relative;
  z-index: 1;
}

@media (max-width: 1200px) {
  .generators-grid {
    grid-template-columns: repeat(3, 1fr);
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

/* === 生成器卡片 - 皮肤令牌（参考 v2-card 风格） === */
.generators-grid {
  /* 表面 */
  --gc-card-bg: #fff;
  --gc-card-bg-soft: #fafbfc;
  --gc-card-border: rgb(229, 231, 235);
  --gc-card-border-hover: rgba(24, 160, 88, 0.32);
  --gc-card-shadow: 0 1px 2px -2px rgba(0, 0, 0, 0.06), 0 3px 6px 0 rgba(0, 0, 0, 0.04);
  --gc-card-shadow-hover: 0 4px 12px -2px rgba(0, 0, 0, 0.08), 0 8px 20px 0 rgba(0, 0, 0, 0.06);
  --gc-card-radius: 14px;
  --gc-card-padding-y: 18px;
  --gc-card-padding-x: 14px;
  --gc-card-gap: 12px;
  --gc-card-min-height: 132px;
  /* 文字 */
  --gc-title-color: rgb(31, 34, 37);
  --gc-meta-color: rgb(102, 102, 102);
  --gc-accent: #18a058;
  --gc-accent-soft: rgba(24, 160, 88, 0.12);
  /* NEW 标签 */
  --gc-new-bg-from: #ef4444;
  --gc-new-bg-to: #dc2626;
  --gc-new-text: #ffffff;
  --gc-new-shadow: 0 2px 8px rgba(239, 68, 68, 0.22);
  /* 水印 / 箭头 */
  --gc-watermark-color: rgba(24, 160, 88, 0.08);
  --gc-watermark-color-hover: rgba(24, 160, 88, 0.16);
  --gc-arrow-color: rgba(0, 0, 0, 0.32);
  --gc-arrow-color-hover: #fff;
  --gc-arrow-bg-hover: var(--gc-accent);
  /* 动画曲线 */
  --gc-bezier: cubic-bezier(0.4, 0, 0.2, 1);
}

.generator-card {
  background: var(--gc-card-bg);
  border-radius: var(--gc-card-radius);
  padding: var(--gc-card-padding-y) var(--gc-card-padding-x);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--gc-card-gap);
  min-height: var(--gc-card-min-height);
  cursor: pointer;
  transition: all 0.25s var(--gc-bezier);
  border: 1px solid var(--gc-card-border);
  box-shadow: var(--gc-card-shadow);
  position: relative;
  overflow: hidden;
}

.generator-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--gc-card-shadow-hover);
  border-color: var(--gc-card-border-hover);
}

.card-content-bottom {
  width: 100%;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  position: relative;
  z-index: 1;
  margin-top: auto;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  flex: 1;
}

.generator-card .card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--gc-title-color);
  text-align: center;
  line-height: 1.4;
  letter-spacing: -0.2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.new-tag {
  flex-shrink: 0;
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.6px;
  padding: 2px 6px;
  border-radius: 4px;
  background: linear-gradient(135deg, var(--gc-new-bg-from) 0%, var(--gc-new-bg-to) 100%);
  border: none;
  color: var(--gc-new-text);
  box-shadow: var(--gc-new-shadow);
}

.card-desc {
  font-size: 12px;
  color: var(--gc-meta-color);
  margin: 0;
  line-height: 1.5;
  text-align: center;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  box-orient: vertical;
  overflow: hidden;
  font-weight: 400;
  letter-spacing: 0.2px;
}

/* 角标水印（参考 v2-card 的 v2-watermark） */
.card-watermark {
  position: absolute;
  top: -8px;
  right: -8px;
  width: 88px;
  height: 88px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--gc-watermark-color);
  pointer-events: none;
  transition: color 0.25s var(--gc-bezier), transform 0.25s var(--gc-bezier);
}

.generator-card:hover .card-watermark {
  color: var(--gc-watermark-color-hover);
  transform: rotate(8deg) scale(1.05);
}

/* 右下角箭头（参考 v2-card 的 v2-arrow-box） */
.card-arrow-box {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--gc-arrow-color);
  background: transparent;
  transition: all 0.25s var(--gc-bezier);
}

.generator-card:hover .card-arrow-box {
  color: var(--gc-arrow-color-hover);
  background: var(--gc-arrow-bg-hover);
}

/* 弹窗样式 - 使用全局样式 */
.generator-dialog-modal .el-dialog {
  width: 1200px;
  max-width: 90vw;
  height: 95vh;
  max-height: 95vh;
  margin: 0 auto !important;
  position: relative;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  flex-direction: column;
  background: #ffffff;
  border: none;
  border-radius: 20px;
  overflow: hidden;
  box-shadow:
    0 1px 2px -2px rgba(0, 0, 0, 0.08),
    0 3px 6px 0 rgba(0, 0, 0, 0.06),
    0 5px 12px 4px rgba(0, 0, 0, 0.04);
  padding: 0;
}

.generator-dialog-modal .el-dialog__header {
  flex-shrink: 0;
  background: #ffffff;
  border-bottom: 1px solid rgb(240, 240, 240);
  padding: 0;
  position: relative;
}

.gen-dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  position: relative;
}

.gen-dialog-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgb(31, 34, 37);
  font-size: 16px;
  font-weight: 500;
  letter-spacing: 0;
}

.gen-title-icon {
  color: #18a058;
  font-size: 18px;
}

.gen-dialog-header-right {
  display: flex;
  align-items: center;
  gap: 4px;
}

.gen-tutorial-btn {
  color: #94a3b8 !important;
  font-size: 14px;
  font-weight: 400;
}

.gen-tutorial-btn:hover {
  color: #64748b !important;
}

.gen-close-btn {
  color: #94a3b8 !important;
  margin-left: 4px;
}

.gen-close-btn:hover {
  color: #64748b !important;
}

.generator-dialog-modal .el-dialog__title {
  color: #0f766e;
  font-size: 17px;
  font-weight: 700;
}

.generator-dialog-modal .el-dialog__body {
  flex: 1;
  overflow: hidden;
  background: transparent;
  padding: 0;
}

.generator-dialog-modal .el-dialog__footer {
  flex-shrink: 0;
  border-top: 1px solid rgb(240, 240, 240);
  padding: 12px 16px;
  background: #ffffff;
}

.gen-dialog-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.gen-footer-disclaimer {
  font-size: 13px;
  color: #6b8481;
  letter-spacing: 0.2px;
}

.gen-footer-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.gen-generate-btn {
  min-width: 110px;
  background: #18a058 !important;
  border-color: #18a058 !important;
  color: #ffffff !important;
  box-shadow: none;
  font-weight: 400;
  border-radius: 3px;
  transition: background-color 0.2s cubic-bezier(0.4, 0, 0.2, 1), border-color 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.gen-generate-btn:hover:not(.is-disabled) {
  background: #36ad6a !important;
  border-color: #36ad6a !important;
  color: #ffffff !important;
}

.gen-generate-btn:active:not(.is-disabled) {
  background: #0c7a43 !important;
  border-color: #0c7a43 !important;
  color: #ffffff !important;
}

.gen-generate-btn.is-disabled {
  background: #18a058 !important;
  border-color: #18a058 !important;
  color: #ffffff !important;
  opacity: 0.5;
  box-shadow: none;
}

.dialog-content {
  display: flex;
  height: 100%;
  min-height: 0;
}

/* 左侧模板区 */
.prompt-template-area {
  width: 36%;
  min-width: 320px;
  padding: 8px;
  border-right: 1px solid rgb(240, 240, 240);
  display: flex;
  flex-direction: column;
  background: #ffffff;
}

.template-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(8, 198, 190, 0.14);
  margin-bottom: 12px;
}

.template-detail-btn {
  background: #ffffff !important;
  border: 1px solid rgb(224, 224, 230) !important;
  color: rgb(51, 54, 57) !important;
  font-weight: 400;
  border-radius: 3px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.template-detail-btn:hover {
  background: #ffffff !important;
  border-color: #36ad6a !important;
  color: #36ad6a !important;
}

.card-pack-select {
  flex: 1;
  min-width: 0;
}

.template-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 4px 0 8px 0;
}

.template-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  background: #ffffff;
  border: 1px solid transparent;
  border-radius: 3px;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
}

.template-item:hover {
  background: rgb(250, 250, 252);
}

.template-item.active {
  background: rgba(24, 160, 88, 0.08);
}

.template-item-switch {
  flex-shrink: 0;
}

.template-item-name {
  flex: 1;
  font-size: 14px;
  font-weight: 400;
  color: rgb(51, 54, 57);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  cursor: pointer;
}

.template-item.active .template-item-name {
  color: #18a058;
  font-weight: 500;
}

.template-item-remove {
  color: #7f9a96;
  cursor: pointer;
  transition: color 0.2s;
  flex-shrink: 0;
  font-size: 14px;
}

.template-item-remove:hover {
  color: #f43f5e;
}

.empty-template {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: rgba(118, 124, 130, 1);
}

.empty-template p {
  margin-top: 12px;
  font-size: 14px;
  color: rgba(118, 124, 130, 1);
}

.template-footer {
  padding-top: 8px;
  border-top: 1px solid rgb(240, 240, 240);
  margin-top: 4px;
}

.more-prompt-btn {
  width: 100%;
  background: #18a058 !important;
  border-color: #18a058 !important;
  color: #ffffff !important;
  font-weight: 400;
  font-size: 14px;
  height: 34px;
  border-radius: 3px;
  transition: background-color 0.2s cubic-bezier(0.4, 0, 0.2, 1), border-color 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.more-prompt-btn:hover {
  background: #36ad6a !important;
  border-color: #36ad6a !important;
  color: #ffffff !important;
}

.more-prompt-btn:active {
  background: #0c7a43 !important;
  border-color: #0c7a43 !important;
  color: #ffffff !important;
}

/* 右侧表单区 */
.form-area {
  flex: 1;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  overflow-y: auto;
  background: #ffffff;
}

.form-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-block-label {
  font-size: 14px;
  font-weight: 400;
  color: rgb(31, 34, 37);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 0 0 6px 2px;
  line-height: 26px;
  height: 26px;
}

.form-block-action {
  font-weight: 400;
  font-size: 12px;
  color: #94a3b8 !important;
  background: transparent !important;
  border: none !important;
  padding: 2px 6px !important;
}

.form-block-action:hover {
  color: #64748b !important;
  background: transparent !important;
}

.form-block-label .required-tag {
  margin-left: 4px;
  height: 20px;
  line-height: 18px;
  font-size: 11px;
  padding: 0 6px;
}

.required-mark {
  color: #d03050;
  font-size: 14px;
  font-weight: 400;
  line-height: 1;
  margin-left: 2px;
}

.prompt-tab-group {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.prompt-tab-group .el-button {
  font-weight: 400;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid rgb(224, 224, 230);
  border-radius: 3px;
  background: #ffffff;
  color: rgb(51, 54, 57);
  height: 28px;
  padding: 0 14px;
  font-size: 14px;
}

.prompt-tab-group .el-button:hover {
  border-color: #36ad6a;
  color: #36ad6a;
}

.prompt-tab-group .el-button.is-active,
.prompt-tab-group .el-button--primary {
  background: #18a058 !important;
  border-color: #18a058 !important;
  color: #ffffff !important;
}

.prompt-tab-group .el-button--primary:hover {
  background: #36ad6a !important;
  border-color: #36ad6a !important;
  color: #ffffff !important;
}

.prompt-template-select {
  margin-top: 4px;
}

.custom-prompt-container {
  margin-top: 4px;
}

.custom-prompt-container :deep(.el-textarea__inner) {
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  border-radius: 8px;
  padding: 12px;
  font-size: 14px;
  line-height: 1.6;
}

.custom-prompt-container :deep(.el-textarea__inner:focus) {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.custom-prompt-container :deep(.el-input__count) {
  background: transparent;
  color: var(--text-secondary);
}

.form-area .el-input__wrapper {
  background: #ffffff;
  border-radius: 3px;
  box-shadow: 0 0 0 1px rgb(224, 224, 230) inset;
  transition: box-shadow 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  padding: 0 12px;
  min-height: 28px;
}

.form-area .el-input__wrapper:hover {
  box-shadow: 0 0 0 1px #36ad6a inset;
}

.form-area .el-input__wrapper.is-focus {
  box-shadow:
    0 0 0 1px #18a058 inset,
    0 0 0 2px rgba(24, 160, 88, 0.2);
}

.form-area .el-textarea__inner {
  background: #ffffff;
  border-radius: 3px;
  color: rgb(51, 54, 57);
  font-size: 14px;
  line-height: 1.6;
  padding: 4px 12px;
  resize: none;
  border: 1px solid rgb(224, 224, 230);
  transition: border-color 0.2s cubic-bezier(0.4, 0, 0.2, 1), box-shadow 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.form-area .el-textarea__inner:hover {
  border-color: #36ad6a;
}

.form-area .el-textarea__inner:focus {
  border-color: #18a058;
  box-shadow: 0 0 0 2px rgba(24, 160, 88, 0.2);
  outline: none;
}

.form-area .el-textarea__inner::placeholder {
  color: rgba(194, 194, 194, 1);
  font-weight: 400;
}

.form-area .el-select__wrapper {
  background: #ffffff !important;
  border-radius: 3px !important;
  box-shadow: 0 0 0 1px rgb(224, 224, 230) inset !important;
  transition: box-shadow 0.2s cubic-bezier(0.4, 0, 0.2, 1) !important;
  min-height: 28px;
  font-size: 14px;
  font-weight: 400;
}

.form-area .el-select__wrapper:hover {
  box-shadow: 0 0 0 1px #36ad6a inset !important;
}

.form-area .el-select__wrapper.is-focused {
  box-shadow:
    0 0 0 1px #18a058 inset,
    0 0 0 2px rgba(24, 160, 88, 0.2) !important;
}

.form-tip {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 13px;
  margin-top: 2px;
  color: #5f7f7b;
}

.form-tip-orange {
  color: #ea7c2c;
  font-weight: 500;
}

.form-tip-text {
  color: inherit;
}

.form-tip-content {
  color: inherit;
  flex: 1;
  min-width: 0;
}

.form-tip-link {
  color: #18a058;
  cursor: pointer;
  text-decoration: none;
  transition: color 0.2s;
}

.form-tip-link:hover {
  color: #36ad6a;
  text-decoration: underline;
}

.form-tip-link-strong {
  color: #18a058;
  font-weight: 500;
}

.field-description {
  font-size: 13px;
  color: rgba(118, 124, 130, 1);
  margin-top: -2px;
}

.input-container {
  position: relative;
  width: 100%;
}

.input-container .fullscreen-btn {
  position: absolute;
  right: 6px;
  top: 6px;
  z-index: 2;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(8, 198, 190, 0.3);
  color: #08c6be;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.06);
  transition: all 0.2s ease;
}

.input-container .fullscreen-btn:hover {
  background: #08c6be;
  color: #fff;
  border-color: #08c6be;
  transform: scale(1.05);
}

.fullscreen-editor-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.fullscreen-editor-footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  font-size: 12px;
  color: #7f9a96;
}

.fullscreen-editor-footer .char-count {
  color: #08c6be;
  font-weight: 500;
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

  .prompt-template-area,
  .form-area {
    width: 100%;
    min-width: 0;
  }

  .prompt-template-area {
    max-height: 32%;
    border-right: none;
    border-bottom: 1px solid rgba(8, 198, 190, 0.16);
  }

  .form-area {
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
  font-size: 13px;
  color: rgba(118, 124, 130, 1);
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
/* 固定提示词弹窗 - 内容 */
.fixed-prompt-dialog-content {
  padding: 10px 0;
}

.fixed-prompt-dialog-content :deep(.el-textarea__inner) {
  background-color: var(--fp-bg-soft);
  color: var(--fp-text-color);
  font-family: 'Courier New', monospace;
  line-height: var(--fp-line-height);
  border-color: var(--fp-border-color);
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

/* 生成器管理弹窗样式 - 现代极简风格 */
.generator-manager-dialog-modal .el-dialog {
  background: #ffffff;
  border-radius: 16px;
  overflow: hidden;
  box-shadow:
    0 0 0 1px rgba(0, 0, 0, 0.03),
    0 2px 8px rgba(0, 0, 0, 0.04),
    0 12px 32px rgba(0, 0, 0, 0.08),
    0 32px 64px rgba(0, 0, 0, 0.06);
}

.generator-manager-dialog-modal .el-dialog__header {
  background: #fafafa;
  border-bottom: 1px solid #f0f0f0;
  padding: 16px 24px;
}

.generator-manager-dialog-modal .el-dialog__title {
  color: #111827;
  font-weight: 600;
  font-size: 16px;
  letter-spacing: -0.01em;
}

.generator-manager-dialog-modal .el-dialog__headerbtn {
  top: 16px;
  right: 20px;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  transition: background 0.15s;
}

.generator-manager-dialog-modal .el-dialog__headerbtn:hover {
  background: rgba(0, 0, 0, 0.04);
}

.generator-manager-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #6b7280;
  font-size: 16px;
}

.generator-manager-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #111827;
}

.generator-manager-dialog-modal .el-dialog__body {
  padding: 0;
  background: #ffffff;
}

.generator-manager-content {
  display: flex;
  height: 520px;
}

/* 左侧列表面板 */
.generator-list-panel {
  width: 300px;
  border-right: 1px solid #f0f0f0;
  display: flex;
  flex-direction: column;
  background: #fafafa;
}

.generator-list-panel .panel-header {
  padding: 14px 16px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fafafa;
}

.panel-title {
  font-weight: 600;
  color: #111827;
  font-size: 13px;
  letter-spacing: -0.01em;
  text-transform: uppercase;
}

.generator-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.generator-list::-webkit-scrollbar {
  width: 4px;
}

.generator-list::-webkit-scrollbar-track {
  background: transparent;
}

.generator-list::-webkit-scrollbar-thumb {
  background: #e5e7eb;
  border-radius: 4px;
}

.generator-list::-webkit-scrollbar-thumb:hover {
  background: #d1d5db;
}

.generator-item {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
  margin-bottom: 2px;
  border: 1px solid transparent;
}

.generator-item:hover {
  background-color: #f3f4f6;
}

.generator-item.active {
  background-color: #ffffff;
  border-color: #e5e7eb;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}

.generator-item-icon {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background: linear-gradient(135deg, #14b8a6 0%, #0d9488 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  margin-right: 12px;
  flex-shrink: 0;
  box-shadow: 0 2px 6px rgba(20, 184, 166, 0.25);
}

.generator-item-icon .el-icon {
  color: #ffffff !important;
  opacity: 1 !important;
}

.generator-item-info {
  flex: 1;
  min-width: 0;
}

.generator-item-name {
  display: block;
  font-weight: 500;
  color: #111827;
  font-size: 13px;
  margin-bottom: 2px;
  letter-spacing: -0.01em;
}

.generator-item-desc {
  display: block;
  font-size: 12px;
  color: #6b7280;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.4;
}

.generator-item-actions {
  opacity: 1;
}

.empty-list {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: #9ca3af;
}

.empty-list .el-icon {
  color: #d1d5db;
}

.empty-list p {
  margin-top: 12px;
  font-size: 13px;
  color: #9ca3af;
}

.empty-list .hint {
  font-size: 12px;
  margin-top: 4px;
  color: #d1d5db;
}

/* 右侧详情面板 */
.generator-detail-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: #ffffff;
}

.detail-content {
  padding: 24px;
  flex: 1;
  overflow-y: auto;
}

.detail-content::-webkit-scrollbar {
  width: 4px;
}

.detail-content::-webkit-scrollbar-track {
  background: transparent;
}

.detail-content::-webkit-scrollbar-thumb {
  background: #e5e7eb;
  border-radius: 4px;
}

.detail-section {
  margin-bottom: 20px;
}

.detail-label {
  display: block;
  font-weight: 500;
  color: #374151;
  font-size: 13px;
  margin-bottom: 8px;
  letter-spacing: -0.01em;
}

.collapsible-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  user-select: none;
  margin-bottom: 0;
  transition: color 0.15s;
}

.collapsible-label:hover {
  color: #14b8a6;
}

.collapse-arrow {
  transition: transform 0.2s ease;
  font-size: 14px;
}

.collapse-arrow.expanded {
  transform: rotate(180deg);
}

.icon-picker-grid {
  margin-top: 8px;
}

.detail-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.empty-detail {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #d1d5db;
}

.empty-detail .el-icon {
  color: #e5e7eb;
}

.empty-detail p {
  margin-top: 12px;
  font-size: 13px;
  color: #9ca3af;
}

/* 图标选择器 */
.icon-picker-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 4px 0;
}

.icon-picker-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 60px;
  height: 60px;
  border-radius: 10px;
  border: 1.5px solid #e5e7eb;
  background: #ffffff;
  cursor: pointer;
  transition: all 0.15s;
  gap: 4px;
}

.icon-picker-item:hover {
  border-color: #14b8a6;
  background: rgba(20, 184, 166, 0.03);
  transform: translateY(-1px);
}

.icon-picker-item.selected {
  border-color: #14b8a6;
  background: rgba(20, 184, 166, 0.06);
  box-shadow: 0 0 0 2px rgba(20, 184, 166, 0.12);
}

.icon-picker-item .el-icon {
  color: #6b7280;
}

.icon-picker-item.selected .el-icon {
  color: #14b8a6;
}

.icon-picker-name {
  font-size: 10px;
  color: #9ca3af;
  max-width: 56px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.icon-picker-item.selected .icon-picker-name {
  color: #14b8a6;
  font-weight: 500;
}

/* 生成器列表排序按钮 */
.generator-item-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  opacity: 1;
}

.generator-item-actions .el-button {
  width: 28px;
  height: 28px;
  padding: 0;
  border-radius: 6px;
  transition: all 0.15s ease;
  background: transparent;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
}

.generator-item-actions .el-button .el-icon {
  font-size: 14px;
}

/* 上移/下移按钮 */
.generator-item-actions .el-button--primary.link {
  color: #9ca3af !important;
}

.generator-item-actions .el-button--primary.link:hover {
  color: #14b8a6 !important;
  background: rgba(20, 184, 166, 0.08);
}

.generator-item-actions .el-button--primary.link .el-icon {
  color: #9ca3af !important;
}

.generator-item-actions .el-button--primary.link:hover .el-icon {
  color: #14b8a6 !important;
}

/* 删除按钮 */
.generator-item-actions .el-button--danger.link {
  color: #9ca3af !important;
}

.generator-item-actions .el-button--danger.link:hover {
  color: #ef4444 !important;
  background: rgba(239, 68, 68, 0.08);
}

.generator-item-actions .el-button--danger.link .el-icon {
  color: #9ca3af !important;
}

.generator-item-actions .el-button--danger.link:hover .el-icon {
  color: #ef4444 !important;
}

/* 禁用状态 */
.generator-item-actions .el-button.is-disabled {
  opacity: 0.35;
  cursor: not-allowed;
  background: transparent;
  border-color: transparent;
}

.generator-item-actions .el-button.is-disabled .el-icon {
  color: #d1d5db !important;
}

.generator-item-actions .el-button.is-disabled:hover {
  background: transparent;
  border-color: transparent;
}

/* === 提示词选择弹窗 - 皮肤令牌（参考 Naive UI n-card 风格） === */
.prompt-select-dialog-modal {
  /* 表面 / 文字 */
  --ps-bg: #fff;
  --ps-bg-embedded: rgb(250, 250, 252);
  --ps-bg-soft: #fafafa;
  --ps-text-color: rgb(51, 54, 57);
  --ps-text-color-secondary: rgb(102, 102, 102);
  --ps-text-color-tertiary: rgba(0, 0, 0, 0.42);
  --ps-title-color: rgb(31, 34, 37);
  --ps-meta-color: rgb(102, 102, 102);
  /* 边框 / 分割线 */
  --ps-border-color: rgb(239, 239, 245);
  --ps-border-color-soft: rgba(0, 0, 0, 0.06);
  --ps-divider: #e5e7eb;
  /* 阴影 */
  --ps-shadow: 0 1px 2px -2px rgba(0, 0, 0, 0.08), 0 3px 6px 0 rgba(0, 0, 0, 0.06), 0 5px 12px 4px rgba(0, 0, 0, 0.04);
  --ps-shadow-card-hover: 0 2px 8px -2px rgba(0, 0, 0, 0.08), 0 4px 12px 0 rgba(0, 0, 0, 0.04);
  --ps-shadow-focus: 0 0 0 2px rgba(24, 160, 88, 0.16);
  /* 圆角 */
  --ps-radius-dialog: 20px;
  --ps-radius-card: 12px;
  --ps-radius-input: 8px;
  --ps-radius-chip: 999px;
  /* 主色 */
  --ps-accent: #18a058;
  --ps-accent-hover: #36ad6a;
  --ps-accent-soft-bg: rgba(24, 160, 88, 0.08);
  --ps-accent-soft-bg-hover: rgba(24, 160, 88, 0.14);
  --ps-accent-border: rgba(24, 160, 88, 0.24);
  /* 危险色 */
  --ps-danger: #f43f5e;
  --ps-danger-soft-bg: rgba(244, 63, 94, 0.06);
  --ps-danger-border: rgba(244, 63, 94, 0.32);
  /* 字号 / 间距 / 曲线 */
  --ps-pad-y: 12px;
  --ps-pad-x: 16px;
  --ps-pad-content: 20px;
  --ps-font-size: 14px;
  --ps-font-size-sm: 13px;
  --ps-font-size-xs: 12px;
  --ps-title-size: 16px;
  --ps-title-font-weight: 500;
  --ps-line-height: 1.6;
  --ps-bezier: cubic-bezier(0.4, 0, 0.2, 1);
  /* 卡片 */
  --ps-card-bg: #fff;
  --ps-card-bg-hover: #fff;
  --ps-card-border: rgb(239, 239, 245);
  --ps-card-border-hover: rgba(24, 160, 88, 0.32);
  /* 关闭按钮 */
  --ps-close-size: 22px;
  --ps-close-color: rgba(102, 102, 102, 1);
  --ps-close-hover-bg: rgba(0, 0, 0, 0.09);
  /* 滚动条 */
  --ps-scroll-thumb: rgba(0, 0, 0, 0.18);
  --ps-scroll-thumb-hover: rgba(0, 0, 0, 0.28);
  /* 标签 / 装饰 */
  --ps-tag-bg-soft: rgba(0, 0, 0, 0.04);
  --ps-tag-bg-featured: rgba(251, 146, 60, 0.12);
  --ps-tag-color-featured: #ea580c;
  --ps-tag-border-featured: rgba(251, 146, 60, 0.32);
  --ps-fire-color: #f97316;
  --ps-crown-color: #fbbf24;
}

/* 弹窗容器 */
.prompt-select-dialog-modal .el-dialog {
  width: 800px;
  max-width: 95vw;
  height: 85vh;
  max-height: 85vh;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  background: var(--ps-bg);
  border: 1px solid var(--ps-border-color);
  border-radius: var(--ps-radius-dialog);
  overflow: hidden;
  box-shadow: var(--ps-shadow);
  padding: 0;
  animation: psDialogEnter 0.3s var(--ps-bezier);
}

@keyframes psDialogEnter {
  from { opacity: 0; transform: scale(0.98); }
  to   { opacity: 1; transform: scale(1); }
}

/* 头部 */
.prompt-select-dialog-modal .el-dialog__header {
  background: var(--ps-bg);
  border-bottom: 1px solid var(--ps-border-color);
  padding: 0;
}

.ps-dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--ps-pad-y) var(--ps-pad-x);
}

.ps-dialog-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--ps-title-color);
  font-size: var(--ps-title-size);
  font-weight: var(--ps-title-font-weight);
}

.ps-title-icon { color: var(--ps-accent); font-size: 18px; }

.ps-close-btn {
  width: var(--ps-close-size);
  height: var(--ps-close-size);
  border-radius: 3px;
  color: var(--ps-close-color) !important;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s var(--ps-bezier);
}
.ps-close-btn:hover {
  background: var(--ps-close-hover-bg);
  color: var(--ps-close-color) !important;
}

/* Body */
.prompt-select-dialog-modal .el-dialog__body {
  padding: 0;
  background: var(--ps-bg);
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.prompt-select-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

/* 顶部标签导航 */
.tab-navigation {
  display: flex;
  gap: 24px;
  padding: 14px var(--ps-pad-x) 4px;
  background: transparent;
  border-bottom: 1px solid var(--ps-border-color);
  align-items: center;
}

.tab-item {
  position: relative;
  padding: 6px 4px 10px;
  font-size: var(--ps-font-size);
  font-weight: 500;
  color: var(--ps-text-color-secondary);
  cursor: pointer;
  transition: color 0.2s var(--ps-bezier);
  letter-spacing: 0.2px;
}
.tab-item:hover { color: var(--ps-title-color); }
.tab-item.active { color: var(--ps-accent); font-weight: 600; }
.tab-item.active::after {
  content: '';
  position: absolute;
  left: 50%;
  bottom: 0;
  transform: translateX(-50%);
  width: 24px;
  height: 2px;
  background: var(--ps-accent);
  border-radius: 2px;
}

/* 搜索区 */
.search-section {
  width: 100%;
  padding: 12px var(--ps-pad-x);
  background: transparent;
  border-bottom: 1px solid var(--ps-border-color);
}

.search-box { display: flex; gap: 10px; align-items: center; }
.search-input-field { flex: 1; }

.search-input-field .el-input__wrapper {
  background: var(--ps-bg-soft);
  border: 1px solid var(--ps-border-color-soft);
  box-shadow: none;
  border-radius: var(--ps-radius-input);
  transition: all 0.2s var(--ps-bezier);
  padding: 2px 12px;
  height: 36px;
}
.search-input-field .el-input__wrapper:hover {
  border-color: var(--ps-accent-border);
  background: var(--ps-bg);
}
.search-input-field .el-input__wrapper.is-focus {
  border-color: var(--ps-accent);
  background: var(--ps-bg);
  box-shadow: var(--ps-shadow-focus);
}
.search-input-field .el-input__inner {
  color: var(--ps-text-color);
  font-weight: 400;
  font-size: var(--ps-font-size-sm);
}
.search-input-field .el-input__inner::placeholder { color: var(--ps-text-color-tertiary); }

.search-btn {
  padding: 0 18px;
  height: 36px;
  background: var(--ps-accent);
  color: #fff;
  border: none;
  border-radius: var(--ps-radius-input);
  font-size: var(--ps-font-size-sm);
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 5px;
  transition: background 0.2s var(--ps-bezier);
}
.search-btn:hover { background: var(--ps-accent-hover); }

/* 分类筛选 */
.filter-section {
  background: transparent;
  border-radius: 0;
  overflow: hidden;
  border-bottom: 1px solid var(--ps-border-color);
}
.filter-section:hover { border-color: var(--ps-border-color); box-shadow: none; }

.filter-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px var(--ps-pad-x) 6px;
  cursor: pointer;
  transition: color 0.2s var(--ps-bezier);
  user-select: none;
  background: transparent;
  font-size: var(--ps-font-size-sm);
  font-weight: 500;
  color: var(--ps-title-color);
}
.filter-header:hover { color: var(--ps-accent); }
.filter-header span { flex: 1; font-size: var(--ps-font-size-sm); font-weight: 500; color: inherit; }
.filter-arrow { transition: transform 0.2s var(--ps-bezier); color: var(--ps-text-color-tertiary); font-size: 12px; }
.filter-arrow.expanded { transform: rotate(180deg); color: var(--ps-accent); }

.category-dropdown {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 8px var(--ps-pad-x) 14px;
  border-top: 1px dashed var(--ps-border-color);
  max-height: 160px;
  overflow-y: auto;
  animation: dropdownSlide 0.2s var(--ps-bezier);
  background: var(--ps-bg-soft);
}

@keyframes dropdownSlide {
  from { opacity: 0; transform: translateY(-4px); }
  to   { opacity: 1; transform: translateY(0); }
}

.dropdown-item {
  padding: 5px 14px;
  font-size: var(--ps-font-size-sm);
  color: var(--ps-text-color-secondary);
  cursor: pointer;
  transition: all 0.2s var(--ps-bezier);
  border: 1px solid var(--ps-border-color);
  background: var(--ps-bg);
  border-radius: var(--ps-radius-chip);
  font-weight: 500;
  white-space: nowrap;
}
.dropdown-item:hover {
  background: var(--ps-accent-soft-bg);
  color: var(--ps-accent);
  border-color: var(--ps-accent-border);
}
.dropdown-item.selected {
  background: var(--ps-accent-soft-bg);
  color: var(--ps-accent);
  font-weight: 600;
  border-color: var(--ps-accent);
}

.selected-tags { display: flex; align-items: center; gap: 6px; padding: 6px var(--ps-pad-x); background: var(--ps-bg); }
.tag-label { font-size: 11px; color: var(--ps-text-color-tertiary); font-weight: 500; }

/* 排序选项栏 */
.sort-section {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px var(--ps-pad-x);
  background: transparent;
  border-bottom: 1px solid var(--ps-border-color);
}
.sort-label { display: none; }
.sort-options { display: flex; gap: 4px; flex-wrap: wrap; }

.sort-btn {
  padding: 4px 12px;
  background: transparent;
  border: none;
  color: var(--ps-text-color-secondary);
  font-size: var(--ps-font-size-sm);
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 3px;
  transition: color 0.2s var(--ps-bezier);
  position: relative;
}
.sort-btn:hover { color: var(--ps-title-color); }
.sort-btn.active { color: var(--ps-accent); font-weight: 600; }
.sort-btn.active::after {
  content: '';
  position: absolute;
  left: 12px; right: 12px; bottom: -8px;
  height: 2px;
  background: var(--ps-accent);
  border-radius: 2px;
}

/* 卡片列表 */
.prompt-select-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  padding: 12px var(--ps-pad-x);
  gap: 0;
  background: var(--ps-bg-embedded);
}

.prompt-select-list::-webkit-scrollbar { width: 6px; }
.prompt-select-list::-webkit-scrollbar-track { background: transparent; }
.prompt-select-list::-webkit-scrollbar-thumb { background: var(--ps-scroll-thumb); border-radius: 3px; }
.prompt-select-list::-webkit-scrollbar-thumb:hover { background: var(--ps-scroll-thumb-hover); }

/* 卡片 */
.prompt-card-item {
  display: flex;
  flex-direction: column;
  background: var(--ps-card-bg);
  border: 1px solid var(--ps-card-border);
  border-radius: var(--ps-radius-card);
  padding: 14px 16px;
  transition: all 0.2s var(--ps-bezier);
  margin-bottom: 12px;
  position: relative;
  gap: 10px;
}
.prompt-card-item:last-child { margin-bottom: 0; }
.prompt-card-item::before { display: none; }
.prompt-card-item:hover {
  border-color: var(--ps-card-border-hover);
  background: var(--ps-card-bg-hover);
  box-shadow: var(--ps-shadow-card-hover);
}
.prompt-card-item.selected {
  border-color: var(--ps-accent);
  background: var(--ps-accent-soft-bg);
  box-shadow: none;
}
.prompt-card-item.selected::before { display: none; }

.card-main {
  flex: 1;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}
.card-top-row { display: flex; align-items: center; gap: 10px; position: relative; }

.author-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  flex-shrink: 0;
  background: var(--ps-accent);
}

.card-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px; }
.card-title-row { display: flex; align-items: center; gap: 8px; min-width: 0; }

.card-tags {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
  margin-top: 6px;
  padding-left: 46px;
}
.card-category-tag {
  background: var(--ps-accent-soft-bg) !important;
  border-color: var(--ps-accent-border) !important;
  color: var(--ps-accent) !important;
  font-size: 10px;
  padding: 0 6px;
  height: 18px;
  line-height: 18px;
}
.card-version-tag {
  background: var(--ps-tag-bg-soft) !important;
  border-color: var(--ps-border-color-soft) !important;
  color: var(--ps-text-color-secondary) !important;
  font-size: 10px;
  padding: 0 6px;
  height: 18px;
  line-height: 18px;
}

.card-title {
  font-size: var(--ps-font-size);
  font-weight: 600;
  color: var(--ps-title-color);
  line-height: 1.4;
  flex: 1;
  min-width: 0;
  max-width: calc(100% - 90px);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-meta-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  padding-left: 46px;
}
.author-name { font-size: var(--ps-font-size-xs); color: var(--ps-meta-color); display: flex; align-items: center; gap: 4px; }
.author-stats { display: flex; align-items: center; gap: 8px; }
.stat-item { display: flex; align-items: center; gap: 3px; font-size: var(--ps-font-size-xs); color: var(--ps-meta-color); font-weight: 500; }
.stat-icon { width: 13px; height: 13px; }
.fire-icon { color: var(--ps-fire-color); }

.card-featured-tag {
  background: var(--ps-tag-bg-featured) !important;
  border-color: var(--ps-tag-border-featured) !important;
  color: var(--ps-tag-color-featured) !important;
  font-size: 10px !important;
  padding: 0 6px !important;
  height: 18px !important;
  line-height: 18px !important;
  font-weight: 600 !important;
}
.author-crown { color: var(--ps-crown-color); font-size: 12px; }
.card-date { font-size: var(--ps-font-size-xs); color: var(--ps-text-color-tertiary); }

.card-content { margin-top: 0; }
.card-desc {
  font-size: 12.5px;
  color: var(--ps-meta-color);
  line-height: var(--ps-line-height);
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  font-weight: 400;
  padding-left: 46px;
}

/* 卡片操作按钮 */
.card-actions {
  position: absolute;
  top: 14px;
  right: 14px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  background: transparent;
  border-left: none;
  min-width: 0;
}

.action-btn {
  padding: 4px 12px;
  border: 1px solid var(--ps-danger-border);
  border-radius: var(--ps-radius-chip);
  font-size: var(--ps-font-size-xs);
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 3px;
  transition: all 0.2s var(--ps-bezier);
  white-space: nowrap;
  background: var(--ps-danger-soft-bg);
  color: var(--ps-danger);
}
.action-btn:hover {
  background: var(--ps-danger);
  color: #fff;
  border-color: var(--ps-danger);
  transform: none;
  box-shadow: none;
}
.action-btn--add {
  border-color: var(--ps-accent-border);
  background: var(--ps-accent-soft-bg);
  color: var(--ps-accent);
}
.action-btn--add:hover {
  background: var(--ps-accent);
  color: #fff;
  border-color: var(--ps-accent);
  box-shadow: none;
}

/* 空状态 */
.empty-prompts {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--ps-text-color-tertiary);
  background: transparent;
  border-radius: var(--ps-radius-card);
  margin: 8px;
}
.empty-prompts .el-icon { color: var(--ps-text-color-tertiary); }
.empty-prompts p { margin-top: 12px; font-size: var(--ps-font-size-sm); color: var(--ps-meta-color); font-weight: 500; }
.empty-hint { margin-top: 6px !important; font-size: var(--ps-font-size-xs) !important; color: var(--ps-text-color-tertiary) !important; font-weight: 400 !important; }

/* 创建提示词区域 */
.create-prompt-section {
  padding: 14px;
  background: var(--ps-bg);
}
.create-prompt-form { max-width: 100%; margin: 0; }
.form-row { display: flex; gap: 12px; }
.form-row .form-group { flex: 1; margin-bottom: 14px; }
.form-group { margin-bottom: 14px; }
.form-label {
  display: block;
  font-size: var(--ps-font-size-xs);
  font-weight: 500;
  color: var(--ps-title-color);
  margin-bottom: 6px;
}
.form-group .el-input__wrapper,
.form-group .el-textarea__inner {
  background: var(--ps-bg-soft);
  border: 1px solid var(--ps-border-color-soft);
  border-radius: 6px;
  transition: all 0.15s var(--ps-bezier);
}
.form-group .el-input__wrapper:hover,
.form-group .el-textarea__inner:hover {
  border-color: var(--ps-accent-border);
  background: var(--ps-bg);
}
.form-group .el-input__wrapper.is-focus,
.form-group .el-textarea__inner:focus {
  border-color: var(--ps-accent);
  background: var(--ps-bg);
  box-shadow: var(--ps-shadow-focus);
}

.fields-config {
  background: var(--ps-bg-soft);
  border: 1px solid var(--ps-border-color-soft);
  border-radius: 6px;
  padding: 10px;
}
.field-item {
  display: flex;
  gap: 6px;
  align-items: center;
  margin-bottom: 6px;
  padding: 6px;
  background: var(--ps-bg);
  border-radius: 4px;
  border: 1px solid var(--ps-border-color-soft);
}
.field-item:last-child { margin-bottom: 0; }
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid var(--ps-border-color-soft);
  margin-top: 16px;
}
.form-actions .el-button {
  padding: 8px 16px;
  font-size: var(--ps-font-size-sm);
  border-radius: 5px;
}

/* 创建提示词区域 - 紧凑风格 */
.create-prompt-section {
  padding: 14px;
  background: #ffffff;
}

.create-prompt-form {
  max-width: 100%;
  margin: 0;
}

.form-row {
  display: flex;
  gap: 12px;
}

.form-row .form-group {
  flex: 1;
  margin-bottom: 14px;
}

.form-group {
  margin-bottom: 14px;
}

.form-label {
  display: block;
  font-size: 12px;
  font-weight: 500;
  color: #374151;
  margin-bottom: 6px;
}

.form-group .el-input__wrapper,
.form-group .el-textarea__inner {
  background: #fafafa;
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 6px;
  transition: all 0.15s ease;
}

.form-group .el-input__wrapper:hover,
.form-group .el-textarea__inner:hover {
  border-color: rgba(0, 0, 0, 0.15);
  background: #ffffff;
}

.form-group .el-input__wrapper.is-focus,
.form-group .el-textarea__inner:focus {
  border-color: #22c55e;
  background: #ffffff;
  box-shadow: 0 0 0 2px rgba(34, 197, 94, 0.1);
}

.fields-config {
  background: #fafafa;
  border: 1px solid rgba(0, 0, 0, 0.05);
  border-radius: 6px;
  padding: 10px;
}

.field-item {
  display: flex;
  gap: 6px;
  align-items: center;
  margin-bottom: 6px;
  padding: 6px;
  background: #ffffff;
  border-radius: 4px;
  border: 1px solid rgba(0, 0, 0, 0.04);
}

.field-item:last-child {
  margin-bottom: 0;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
  margin-top: 16px;
}

.form-actions .el-button {
  padding: 8px 16px;
  font-size: 13px;
  border-radius: 5px;
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

/* 追问对话弹窗 */
.followup-dialog :deep(.el-dialog) {
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
}

.followup-dialog :deep(.el-dialog__header) {
  padding: 16px 24px;
  border-bottom: 1px solid #e5e7eb;
  background: #f9fafb;
  border-radius: 16px 16px 0 0;
  margin-right: 0;
}

.followup-dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.followup-dialog-header .header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.followup-dialog-header .header-title .el-icon {
  color: #08c6be;
  font-size: 20px;
}

.followup-dialog-header .model-select {
  width: 180px;
}

.followup-dialog :deep(.el-dialog__body) {
  padding: 0;
  background: #ffffff;
}

.followup-dialog-content {
  display: flex;
  flex-direction: column;
  height: 550px;
}

.followup-messages {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  background: #f0f2f5;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.followup-messages .history-collapsed-hint {
  text-align: center;
  padding: 12px 0;
  margin-bottom: 12px;
  position: relative;
}

.followup-messages .history-collapsed-hint::before,
.followup-messages .history-collapsed-hint::after {
  content: '';
  position: absolute;
  top: 50%;
  width: 30%;
  height: 1px;
  background: linear-gradient(to right, transparent, #e5e7eb, transparent);
}

.followup-messages .history-collapsed-hint::before {
  left: 0;
}

.followup-messages .history-collapsed-hint::after {
  right: 0;
}

.followup-messages .expand-btn {
  font-size: 13px;
  color: #08c6be;
  padding: 8px 16px;
  border-radius: 20px;
  background: rgba(8, 198, 190, 0.08);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  z-index: 1;
}

.followup-messages .expand-btn:hover {
  color: #059691;
  background: rgba(8, 198, 190, 0.15);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(8, 198, 190, 0.2);
}

.followup-message-bubble {
  border-radius: 16px;
  padding: 16px 20px;
  max-width: 85%;
  animation: followupMessageSlideIn 0.3s ease;
}

.followup-message-bubble.user {
  background: #e3f2fd;
  border: 1px solid #bbdefb;
  margin-left: auto;
}

.followup-message-bubble.assistant {
  background: #ffffff;
  border: 1px solid #e5e7eb;
}

.followup-message-bubble.system {
  background: #f3e5f5;
  border: 1px solid #e1bee7;
  width: 100%;
  max-width: 100%;
}

.followup-message-bubble.history-message {
  opacity: 0.6;
  transition: all 0.3s ease;
}

.followup-message-bubble.history-message:hover {
  opacity: 1;
}

.followup-message-sender {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  font-size: 13px;
}

.followup-message-sender .sender-icon {
  font-size: 16px;
}

.followup-message-sender .sender-name {
  font-weight: 600;
  color: #6b7280;
}

.followup-message-content {
  color: #374151;
  line-height: 1.7;
  word-break: break-word;
}

.followup-input-area {
  padding: 16px 24px;
  background: #ffffff;
  border-top: 1px solid #e5e7eb;
  display: flex;
  align-items: flex-end;
  gap: 12px;
}

.followup-input {
  flex: 1;
}

.followup-input :deep(.el-textarea__inner) {
  padding: 12px 16px;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  background: #f9fafb;
  transition: all 0.3s ease;
  font-size: 14px;
}

.followup-input :deep(.el-textarea__inner:focus) {
  border-color: #08c6be;
  background: #ffffff;
  box-shadow: 0 2px 8px rgba(8, 198, 190, 0.1);
}

.followup-input-actions {
  display: flex;
  align-items: center;
}

.followup-input-actions .send-btn {
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  border-color: transparent;
  color: #ffffff;
}

.followup-input-actions .send-btn:hover {
  background: linear-gradient(135deg, #09b4ad 0%, #047a75 100%);
}

@keyframes followupMessageSlideIn {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.followup-messages .message-slide-enter-active {
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.followup-messages .message-slide-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.followup-messages .message-slide-enter-from {
  opacity: 0;
  transform: translateY(20px);
}

.followup-messages .message-slide-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}

.followup-messages .typing-indicator {
  display: flex;
  gap: 4px;
  padding: 8px 0;
}

.followup-messages .typing-indicator span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #08c6be;
  animation: followupTyping 1.4s infinite ease-in-out;
}

.followup-messages .typing-indicator span:nth-child(1) {
  animation-delay: -0.32s;
}

.followup-messages .typing-indicator span:nth-child(2) {
  animation-delay: -0.16s;
}

@keyframes followupTyping {
  0%, 80%, 100% {
    transform: scale(0.8);
    opacity: 0.5;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
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

.generator-dialog-modal .form-area::-webkit-scrollbar,
.generator-dialog-modal .template-list::-webkit-scrollbar {
  width: 6px;
}

.generator-dialog-modal .form-area::-webkit-scrollbar-track,
.generator-dialog-modal .template-list::-webkit-scrollbar-track {
  background: transparent;
}

.generator-dialog-modal .form-area::-webkit-scrollbar-thumb,
.generator-dialog-modal .template-list::-webkit-scrollbar-thumb {
  background: rgba(8, 198, 190, 0.2);
  border-radius: 3px;
}

.generator-dialog-modal .form-area::-webkit-scrollbar-thumb:hover,
.generator-dialog-modal .template-list::-webkit-scrollbar-thumb:hover {
  background: rgba(8, 198, 190, 0.35);
}

:root[data-theme='dark'] .form-area::-webkit-scrollbar-thumb,
:root[data-theme='dark'] .template-list::-webkit-scrollbar-thumb {
  background: rgba(94, 234, 212, 0.22);
}

:root[data-theme='dark'] .form-area::-webkit-scrollbar-thumb:hover,
:root[data-theme='dark'] .template-list::-webkit-scrollbar-thumb:hover {
  background: rgba(94, 234, 212, 0.4);
}

/* 生成器管理弹窗 - Element Plus 组件样式 */
.generator-manager-dialog-modal .el-input__wrapper {
  background: #fafafa;
  border: 1px solid #e5e7eb;
  box-shadow: none;
  border-radius: 8px;
  padding: 8px 12px;
  transition: all 0.15s;
}

.generator-manager-dialog-modal .el-input__wrapper:hover {
  border-color: #d1d5db;
}

.generator-manager-dialog-modal .el-input__inner {
  color: #111827;
  font-size: 14px;
}

.generator-manager-dialog-modal .el-input__inner::placeholder {
  color: #9ca3af;
}

.generator-manager-dialog-modal .el-textarea__inner {
  background: #fafafa;
  border: 1px solid #e5e7eb;
  color: #111827;
  border-radius: 8px;
  font-size: 14px;
  padding: 10px 12px;
  transition: all 0.15s;
}

.generator-manager-dialog-modal .el-textarea__inner:hover {
  border-color: #d1d5db;
}

.generator-manager-dialog-modal .el-textarea__inner::placeholder {
  color: #9ca3af;
}

.generator-manager-dialog-modal .el-textarea__inner:focus {
  border-color: #14b8a6;
  box-shadow: 0 0 0 3px rgba(20, 184, 166, 0.08);
}

.generator-manager-dialog-modal .el-input__wrapper:focus-within {
  border-color: #14b8a6;
  box-shadow: 0 0 0 3px rgba(20, 184, 166, 0.08);
  background: #ffffff;
}

.generator-manager-dialog-modal .el-select .el-input__wrapper {
  background: #fafafa;
  border: 1px solid #e5e7eb;
}

.generator-manager-dialog-modal .el-select-dropdown {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  padding: 6px;
}

.generator-manager-dialog-modal .el-select-dropdown__item {
  color: #374151;
  border-radius: 6px;
  padding: 8px 12px;
  font-size: 14px;
}

.generator-manager-dialog-modal .el-select-dropdown__item:hover {
  background: #f3f4f6;
}

.generator-manager-dialog-modal .el-select-dropdown__item.selected {
  background: rgba(20, 184, 166, 0.08);
  color: #0d9488;
  font-weight: 500;
}

.generator-manager-dialog-modal .el-switch.is-checked .el-switch__core {
  background-color: #14b8a6;
  border-color: #14b8a6;
}

.generator-manager-dialog-modal .el-button--primary {
  background: #14b8a6;
  border-color: #14b8a6;
  color: #ffffff;
  border-radius: 8px;
  font-weight: 500;
  font-size: 13px;
  transition: all 0.15s;
}

.generator-manager-dialog-modal .el-button--primary:hover {
  background: #0d9488;
  border-color: #0d9488;
}

.generator-manager-dialog-modal .el-button--primary.is-disabled {
  background: #e5e7eb;
  border-color: #e5e7eb;
  color: #9ca3af;
}

.generator-manager-dialog-modal .el-button--default {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  color: #374151;
  border-radius: 8px;
  font-weight: 500;
  font-size: 13px;
  transition: all 0.15s;
}

.generator-manager-dialog-modal .el-button--default:hover {
  background: #f9fafb;
  border-color: #d1d5db;
}

.generator-manager-dialog-modal .el-button--danger {
  background: #ef4444;
  border-color: #ef4444;
  color: #ffffff;
  border-radius: 8px;
  font-weight: 500;
}

.generator-manager-dialog-modal .el-button--danger:hover {
  background: #dc2626;
  border-color: #dc2626;
}

.generator-manager-dialog-modal .el-tag--info {
  background: rgba(20, 184, 166, 0.06);
  border-color: rgba(20, 184, 166, 0.15);
  color: #0d9488;
  border-radius: 6px;
  font-weight: 500;
}

.generator-manager-dialog-modal .el-icon {
  color: #6b7280;
}

.generator-manager-dialog-modal .el-icon:hover {
  color: #111827;
}

/* 提示词选择弹窗 */
.prompt-select-dialog-modal .el-input__wrapper {
  background: #fafafa;
  border-color: rgba(0, 0, 0, 0.08);
  box-shadow: none;
}

.prompt-select-dialog-modal .el-input__wrapper:hover {
  border-color: rgba(0, 0, 0, 0.15);
}

.prompt-select-dialog-modal .el-input__wrapper.is-focus {
  border-color: #08c6be;
  box-shadow: 0 0 0 2px rgba(8, 198, 190, 0.1);
}

.prompt-select-dialog-modal .el-input__inner {
  color: #134e4a;
}

.prompt-select-dialog-modal .el-input__inner::placeholder {
  color: #a3a3a3;
}

.prompt-select-dialog-modal .el-textarea__inner {
  background: #fafafa;
  border-color: rgba(0, 0, 0, 0.08);
  color: #134e4a;
}

.prompt-select-dialog-modal .el-textarea__inner:hover {
  border-color: rgba(0, 0, 0, 0.15);
}

.prompt-select-dialog-modal .el-textarea__inner:focus {
  border-color: #08c6be;
  box-shadow: 0 0 0 2px rgba(8, 198, 190, 0.1);
}

.prompt-select-dialog-modal .el-textarea__inner::placeholder {
  color: #a3a3a3;
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

/* === 固定提示词查看弹窗 - 皮肤令牌 === */
.fixed-prompt-dialog-modal {
  /* 表面 / 文字 */
  --fp-bg: #fff;
  --fp-bg-soft: #fafafa;
  --fp-text-color: rgb(31, 34, 37);
  --fp-text-color-muted: rgb(102, 102, 102);
  --fp-placeholder-color: #9ca3af;
  /* 边框 / 阴影 / 圆角 */
  --fp-border-color: rgb(229, 231, 235);
  --fp-border-color-hover: #d1d5db;
  --fp-shadow: 0 1px 2px -2px rgba(0, 0, 0, 0.08), 0 3px 6px 0 rgba(0, 0, 0, 0.06), 0 5px 12px 4px rgba(0, 0, 0, 0.04);
  --fp-radius: 14px;
  /* 主色（绿色，与功能语义一致：复制即执行） */
  --fp-accent: #22c55e;
  --fp-accent-hover: #16a34a;
  --fp-accent-text: #ffffff;
  /* 危险 / 次要按钮 */
  --fp-btn-default-bg: #fff;
  --fp-btn-default-border: rgb(229, 231, 235);
  --fp-btn-default-text: rgb(51, 54, 57);
  --fp-btn-default-hover-bg: #f9fafb;
  --fp-btn-default-hover-border: #d1d5db;
  /* 文本域 */
  --fp-textarea-bg: #fafafa;
  --fp-textarea-border: rgb(229, 231, 235);
  --fp-textarea-color: rgb(31, 34, 37);
  --fp-line-height: 1.6;
  /* 关闭 */
  --fp-close-color: rgb(102, 102, 102);
  --fp-close-hover-color: rgb(31, 34, 37);
}

.fixed-prompt-dialog-modal .el-dialog {
  background: var(--fp-bg);
  border: 1px solid var(--fp-border-color);
  border-radius: var(--fp-radius);
  box-shadow: var(--fp-shadow);
  overflow: hidden;
}

.fixed-prompt-dialog-modal .el-dialog__header {
  background: var(--fp-bg);
  border-bottom: 1px solid var(--fp-border-color);
  padding: 14px 20px;
}

.fixed-prompt-dialog-modal .el-dialog__title {
  color: var(--fp-text-color);
  font-weight: 500;
}

.fixed-prompt-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: var(--fp-close-color);
}
.fixed-prompt-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: var(--fp-close-hover-color);
}

.fixed-prompt-dialog-modal .el-dialog__body {
  background: var(--fp-bg);
  padding: 16px 20px;
}

.fixed-prompt-dialog-modal .el-dialog__footer {
  background: var(--fp-bg);
  border-top: 1px solid var(--fp-border-color);
  padding: 12px 20px;
}

.fixed-prompt-dialog-modal .el-input__wrapper,
.fixed-prompt-dialog-modal .el-textarea__inner {
  background: var(--fp-textarea-bg);
  border-color: var(--fp-textarea-border);
  color: var(--fp-textarea-color);
  box-shadow: none;
}

.fixed-prompt-dialog-modal .el-textarea__inner {
  font-family: 'Courier New', monospace;
  line-height: var(--fp-line-height);
}

.fixed-prompt-dialog-modal .el-input__inner {
  color: var(--fp-textarea-color);
}

.fixed-prompt-dialog-modal .el-input__inner::placeholder,
.fixed-prompt-dialog-modal .el-textarea__inner::placeholder {
  color: var(--fp-placeholder-color);
}

.fixed-prompt-dialog-modal .el-textarea__inner:focus,
.fixed-prompt-dialog-modal .el-input__wrapper:focus-within {
  border-color: var(--fp-accent);
}

.fixed-prompt-dialog-modal .el-button--primary {
  background: var(--fp-accent);
  border-color: var(--fp-accent);
  color: var(--fp-accent-text);
}
.fixed-prompt-dialog-modal .el-button--primary:hover {
  background: var(--fp-accent-hover);
  border-color: var(--fp-accent-hover);
}

.fixed-prompt-dialog-modal .el-button--default {
  background: var(--fp-btn-default-bg);
  border-color: var(--fp-btn-default-border);
  color: var(--fp-btn-default-text);
}
.fixed-prompt-dialog-modal .el-button--default:hover {
  background: var(--fp-btn-default-hover-bg);
  border-color: var(--fp-btn-default-hover-border);
  color: var(--fp-btn-default-text);
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

:root[data-theme='dark'] .input-container .fullscreen-btn {
  background: rgba(31, 41, 55, 0.9);
  border-color: rgba(94, 234, 212, 0.4);
  color: #5eead4;
}

:root[data-theme='dark'] .input-container .fullscreen-btn:hover {
  background: #5eead4;
  color: #0f172a;
  border-color: #5eead4;
}

:root[data-theme='dark'] .fullscreen-editor-footer {
  color: #94a3b8;
}

:root[data-theme='dark'] .fullscreen-editor-footer .char-count {
  color: #5eead4;
}

:root[data-theme='dark'] .context-tip {
  color: rgba(255, 255, 255, 0.5);
}

:root[data-theme='dark'] .fixed-prompt-hint {
  background: rgba(94, 234, 212, 0.12);
  color: #5eead4;
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
  background: linear-gradient(135deg, #2dd4bf 0%, #14b8a6 100%);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .pinned-header .el-button:hover {
  background: linear-gradient(135deg, #5eead4 0%, #2dd4bf 100%);
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

/* === 生成器卡片 - 深色皮肤令牌覆盖 === */
:root[data-theme='dark'] .generators-grid {
  --gc-card-bg: rgb(24, 24, 28);
  --gc-card-bg-soft: rgb(30, 30, 36);
  --gc-card-border: rgba(255, 255, 255, 0.08);
  --gc-card-border-hover: rgba(54, 173, 106, 0.4);
  --gc-card-shadow: 0 1px 2px -2px rgba(0, 0, 0, 0.5), 0 3px 6px 0 rgba(0, 0, 0, 0.4);
  --gc-card-shadow-hover: 0 4px 12px -2px rgba(0, 0, 0, 0.5), 0 8px 20px 0 rgba(0, 0, 0, 0.32);
  --gc-title-color: rgba(255, 255, 255, 0.92);
  --gc-meta-color: rgba(255, 255, 255, 0.52);
  --gc-accent: #36ad6a;
  --gc-accent-soft: rgba(54, 173, 106, 0.18);
  --gc-new-shadow: 0 2px 8px rgba(239, 68, 68, 0.35);
  --gc-watermark-color: rgba(255, 255, 255, 0.05);
  --gc-watermark-color-hover: rgba(54, 173, 106, 0.32);
  --gc-arrow-color: rgba(255, 255, 255, 0.42);
  --gc-arrow-color-hover: #0f172a;
  --gc-arrow-bg-hover: var(--gc-accent);
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
  background-color: rgba(94, 234, 212, 0.16);
  border-color: rgba(94, 234, 212, 0.45);
}

:root[data-theme='dark'] .generator-item-icon {
  background: linear-gradient(135deg, #2dd4bf 0%, #14b8a6 100%);
}

:root[data-theme='dark'] .generator-item-icon .el-icon {
  color: #ffffff !important;
  opacity: 1 !important;
}

:root[data-theme='dark'] .panel-title {
  color: #f1f5f9;
}

:root[data-theme='dark'] .generator-item-name {
  color: #f1f5f9;
}

:root[data-theme='dark'] .generator-item-desc {
  color: #94a3b8;
}

:root[data-theme='dark'] .generator-item-actions .el-button.is-disabled {
  opacity: 0.4;
  background: transparent;
  border-color: transparent;
}

:root[data-theme='dark'] .generator-item-actions .el-button.is-disabled .el-icon {
  color: #475569 !important;
}

:root[data-theme='dark'] .generator-item-actions .el-button.is-disabled:hover {
  background: transparent;
  border-color: transparent;
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
  background: #1e293b;
  box-shadow:
    0 1px 2px -2px rgba(0, 0, 0, 0.5),
    0 3px 6px 0 rgba(0, 0, 0, 0.4),
    0 5px 12px 4px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .generator-dialog-modal .el-dialog__header {
  background: #1e293b;
  border-bottom-color: rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .gen-dialog-title {
  color: #f3f4f6;
}

:root[data-theme='dark'] .gen-title-icon {
  color: #5eead4;
}

:root[data-theme='dark'] .gen-tutorial-btn {
  color: rgba(255, 255, 255, 0.5) !important;
}

:root[data-theme='dark'] .gen-tutorial-btn:hover {
  color: #5eead4 !important;
}

:root[data-theme='dark'] .gen-close-btn {
  color: rgba(255, 255, 255, 0.5) !important;
}

:root[data-theme='dark'] .gen-close-btn:hover {
  color: #5eead4 !important;
}

:root[data-theme='dark'] .generator-dialog-modal .el-dialog__title {
  color: #f3f4f6;
}

:root[data-theme='dark'] .generator-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: rgba(255, 255, 255, 0.6);
}

:root[data-theme='dark'] .generator-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .generator-dialog-modal .el-dialog__footer {
  background: #1e293b;
  border-top-color: rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .gen-footer-disclaimer {
  color: rgba(203, 213, 225, 0.85);
}

:root[data-theme='dark'] .prompt-template-area {
  background: rgba(15, 23, 42, 0.45);
  border-right-color: rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .template-header {
  border-bottom-color: rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .template-detail-btn {
  background: #1e293b !important;
  border-color: rgba(94, 234, 212, 0.22) !important;
  color: #cbd5e1 !important;
}

:root[data-theme='dark'] .template-detail-btn:hover {
  background: rgba(94, 234, 212, 0.12) !important;
  border-color: #5eead4 !important;
  color: #5eead4 !important;
}

:root[data-theme='dark'] .template-item {
  background: transparent;
  border-color: transparent;
}

:root[data-theme='dark'] .template-item:hover {
  background: rgba(94, 234, 212, 0.06);
}

:root[data-theme='dark'] .template-item.active {
  background: rgba(94, 234, 212, 0.16);
}

:root[data-theme='dark'] .template-item-name {
  color: #e2e8f0;
}

:root[data-theme='dark'] .template-item.active .template-item-name {
  color: #5eead4;
}

:root[data-theme='dark'] .template-item-remove {
  color: #94a3b8;
}

:root[data-theme='dark'] .template-item-remove:hover {
  color: #f87171;
}

:root[data-theme='dark'] .empty-template {
  color: #94a3b8;
}

:root[data-theme='dark'] .empty-template p {
  color: #94a3b8;
}

:root[data-theme='dark'] .template-footer {
  border-top-color: rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .more-prompt-btn {
  background: linear-gradient(135deg, #2dd4bf 0%, #14b8a6 100%) !important;
  border-color: transparent !important;
  color: #0f172a !important;
  box-shadow: 0 8px 20px rgba(45, 212, 191, 0.35);
}

:root[data-theme='dark'] .more-prompt-btn:hover {
  background: linear-gradient(135deg, #5eead4 0%, #2dd4bf 100%) !important;
  box-shadow: 0 10px 24px rgba(45, 212, 191, 0.45);
}

:root[data-theme='dark'] .form-area {
  background: #1e293b;
}

:root[data-theme='dark'] .form-block-label {
  color: #f3f4f6 !important;
}

:root[data-theme='dark'] .form-area .el-input__wrapper {
  background: rgba(15, 23, 42, 0.6);
  box-shadow: 0 0 0 1px rgba(94, 234, 212, 0.18) inset;
}

:root[data-theme='dark'] .form-area .el-input__wrapper:hover {
  box-shadow: 0 0 0 1px rgba(94, 234, 212, 0.32) inset;
}

:root[data-theme='dark'] .form-area .el-input__wrapper.is-focus {
  box-shadow:
    0 0 0 1px #5eead4 inset,
    0 0 0 2px rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .form-area .el-textarea__inner {
  background: rgba(15, 23, 42, 0.6);
  color: #e2e8f0;
  border-color: rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .form-area .el-textarea__inner:hover {
  border-color: rgba(94, 234, 212, 0.32);
}

:root[data-theme='dark'] .form-area .el-textarea__inner:focus {
  border-color: #5eead4;
  box-shadow: 0 0 0 2px rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .form-area .el-textarea__inner::placeholder {
  color: #64748b;
}

:root[data-theme='dark'] .form-area .el-select__wrapper {
  background: rgba(15, 23, 42, 0.6) !important;
  box-shadow: 0 0 0 1px rgba(94, 234, 212, 0.18) inset !important;
}

:root[data-theme='dark'] .form-area .el-select__wrapper:hover {
  box-shadow: 0 0 0 1px rgba(94, 234, 212, 0.32) inset !important;
}

:root[data-theme='dark'] .form-area .el-select__wrapper.is-focused {
  box-shadow:
    0 0 0 1px #5eead4 inset,
    0 0 0 2px rgba(94, 234, 212, 0.2) !important;
}

:root[data-theme='dark'] .form-block-action {
  color: #5eead4 !important;
  background: transparent !important;
}

:root[data-theme='dark'] .form-block-action:hover {
  color: #2dd4bf !important;
  background: transparent !important;
}

:root[data-theme='dark'] .required-mark {
  color: #f87171;
}

:root[data-theme='dark'] .form-tip {
  color: #94a3b8;
}

:root[data-theme='dark'] .form-tip-orange {
  color: #fbbf24;
}

:root[data-theme='dark'] .form-tip-link {
  color: #5eead4;
}

:root[data-theme='dark'] .form-tip-link:hover {
  color: #2dd4bf;
}

:root[data-theme='dark'] .form-tip-link-strong {
  color: #5eead4;
}

:root[data-theme='dark'] .field-description {
  color: #94a3b8;
}

:root[data-theme='dark'] .gen-generate-btn.is-disabled {
  background: #5eead4 !important;
  border-color: #5eead4 !important;
  color: #0f172a !important;
  opacity: 0.45;
}

:root[data-theme='dark'] .prompt-tab-group .el-button {
  background: rgba(15, 23, 42, 0.6);
  border-color: rgba(94, 234, 212, 0.18);
  color: #cbd5e1;
}

:root[data-theme='dark'] .prompt-tab-group .el-button:hover {
  border-color: #5eead4;
  color: #5eead4;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-dialog {
  background: #1e293b;
  border-radius: 16px;
  box-shadow:
    0 0 0 1px rgba(94, 234, 212, 0.12),
    0 4px 12px rgba(0, 0, 0, 0.3),
    0 16px 40px rgba(0, 0, 0, 0.4),
    0 40px 80px rgba(0, 0, 0, 0.2);
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-dialog__header {
  background: #1e293b;
  border-bottom-color: rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-dialog__title {
  color: #f3f4f6;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #94a3b8;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-dialog__body {
  background: #1e293b;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-input__wrapper {
  background: rgba(15, 23, 42, 0.6);
  border-color: rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .generator-dialog-modal .el-input__wrapper {
  background: rgba(15, 23, 42, 0.6);
  border-color: rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .generator-dialog-modal .el-input__wrapper:hover {
  border-color: rgba(94, 234, 212, 0.32);
}

:root[data-theme='dark'] .generator-dialog-modal .el-input__inner {
  color: #e2e8f0;
}

:root[data-theme='dark'] .generator-dialog-modal .el-input__inner::placeholder {
  color: #64748b;
}

:root[data-theme='dark'] .generator-dialog-modal .el-textarea__inner {
  background: rgba(15, 23, 42, 0.6);
  border-color: rgba(94, 234, 212, 0.18);
  color: #e2e8f0;
}

:root[data-theme='dark'] .generator-dialog-modal .el-textarea__inner:hover {
  border-color: rgba(94, 234, 212, 0.32);
}

:root[data-theme='dark'] .generator-dialog-modal .el-textarea__inner::placeholder {
  color: #64748b;
}

:root[data-theme='dark'] .generator-dialog-modal .el-textarea__inner:focus {
  border-color: #5eead4;
  box-shadow: 0 0 0 2px rgba(94, 234, 212, 0.12);
}

:root[data-theme='dark'] .generator-dialog-modal .el-input__wrapper:focus-within {
  border-color: #5eead4;
  box-shadow: 0 0 0 2px rgba(94, 234, 212, 0.12);
}

:root[data-theme='dark'] .generator-dialog-modal .el-select .el-input__wrapper {
  background: rgba(15, 23, 42, 0.6);
  border-color: rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .generator-dialog-modal .el-select-dropdown {
  background: #1e293b;
  border-color: rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .generator-dialog-modal .el-select-dropdown__item {
  color: #cbd5e1;
}

:root[data-theme='dark'] .generator-dialog-modal .el-select-dropdown__item:hover {
  background: rgba(94, 234, 212, 0.1);
}

:root[data-theme='dark'] .generator-dialog-modal .el-select-dropdown__item.selected {
  background: rgba(94, 234, 212, 0.18);
  color: #5eead4;
}

:root[data-theme='dark'] .generator-dialog-modal .el-tag--info {
  background: rgba(94, 234, 212, 0.12);
  border-color: rgba(94, 234, 212, 0.25);
  color: #5eead4;
}

:root[data-theme='dark'] .generator-dialog-modal .el-icon {
  color: #94a3b8;
}

:root[data-theme='dark'] .generator-dialog-modal .el-icon:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .generator-dialog-modal .el-button--default {
  background: rgba(30, 41, 59, 0.7);
  border-color: rgba(94, 234, 212, 0.22);
  color: #cbd5e1;
}

:root[data-theme='dark'] .generator-dialog-modal .el-button--default:hover {
  background: rgba(94, 234, 212, 0.12);
  border-color: rgba(94, 234, 212, 0.4);
  color: #5eead4;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-input__wrapper:hover {
  border-color: rgba(94, 234, 212, 0.32);
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-input__inner {
  color: #e2e8f0;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-input__inner::placeholder {
  color: #64748b;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-textarea__inner {
  background: rgba(15, 23, 42, 0.6);
  border-color: rgba(94, 234, 212, 0.18);
  color: #e2e8f0;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-textarea__inner:hover {
  border-color: rgba(94, 234, 212, 0.32);
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-textarea__inner::placeholder {
  color: #64748b;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-textarea__inner:focus {
  border-color: #5eead4;
  box-shadow: 0 0 0 3px rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-input__wrapper:focus-within {
  border-color: #5eead4;
  box-shadow: 0 0 0 3px rgba(94, 234, 212, 0.18);
  background: rgba(15, 23, 42, 0.6);
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-select .el-input__wrapper {
  background: rgba(15, 23, 42, 0.6);
  border-color: rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-select-dropdown {
  background: #1e293b;
  border-color: rgba(94, 234, 212, 0.18);
  border-radius: 10px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-select-dropdown__item {
  color: #cbd5e1;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-select-dropdown__item:hover {
  background: rgba(94, 234, 212, 0.08);
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-select-dropdown__item.selected {
  background: rgba(94, 234, 212, 0.18);
  color: #5eead4;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-switch.is-checked .el-switch__core {
  background-color: #5eead4;
  border-color: #5eead4;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-button--primary {
  background: #5eead4;
  border-color: #5eead4;
  color: #0f172a;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-button--primary:hover {
  background: #2dd4bf;
  border-color: #2dd4bf;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-button--primary.is-disabled {
  background: rgba(94, 234, 212, 0.18);
  border-color: rgba(94, 234, 212, 0.12);
  color: #64748b;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-button--default {
  background: rgba(30, 41, 59, 0.7);
  border: 1px solid rgba(94, 234, 212, 0.22);
  color: #cbd5e1;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-button--default:hover {
  background: rgba(94, 234, 212, 0.1);
  border-color: rgba(94, 234, 212, 0.4);
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-button--danger {
  background: #ef4444;
  border-color: #ef4444;
  color: #ffffff;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-button--danger:hover {
  background: #dc2626;
  border-color: #dc2626;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-tag--info {
  background: rgba(94, 234, 212, 0.12);
  border-color: rgba(94, 234, 212, 0.25);
  color: #5eead4;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-icon {
  color: #94a3b8;
}

:root[data-theme='dark'] .generator-manager-dialog-modal .el-icon:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .generator-detail-panel .detail-label {
  color: #cbd5e1;
}

:root[data-theme='dark'] .collapsible-label:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .generator-detail-panel .empty-detail,
:root[data-theme='dark'] .generator-detail-panel .empty-detail .el-icon {
  color: #64748b;
}

:root[data-theme='dark'] .generator-detail-panel .empty-detail p {
  color: #64748b;
}

:root[data-theme='dark'] .generator-list-panel {
  border-right-color: rgba(94, 234, 212, 0.18);
  background: #1e293b;
}

:root[data-theme='dark'] .generator-list-panel .panel-header {
  border-bottom-color: rgba(94, 234, 212, 0.18);
  background: #1e293b;
}

:root[data-theme='dark'] .generator-list-panel .panel-title {
  color: #cbd5e1;
}

:root[data-theme='dark'] .generator-list {
  padding: 6px;
}

:root[data-theme='dark'] .generator-list::-webkit-scrollbar-thumb {
  background: rgba(94, 234, 212, 0.25);
}

:root[data-theme='dark'] .generator-list::-webkit-scrollbar-thumb:hover {
  background: rgba(94, 234, 212, 0.4);
}

:root[data-theme='dark'] .generator-item:hover {
  background-color: rgba(94, 234, 212, 0.08);
}

:root[data-theme='dark'] .generator-item.active {
  background-color: rgba(94, 234, 212, 0.16);
  border-color: rgba(94, 234, 212, 0.3);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .generator-item-icon {
  background: linear-gradient(135deg, #2dd4bf 0%, #14b8a6 100%);
  box-shadow: 0 2px 8px rgba(45, 212, 191, 0.3);
}

:root[data-theme='dark'] .generator-item-icon .el-icon {
  color: #ffffff !important;
  opacity: 1 !important;
}

:root[data-theme='dark'] .generator-item-name {
  color: #f3f4f6;
}

:root[data-theme='dark'] .generator-item-desc {
  color: #94a3b8;
}

:root[data-theme='dark'] .generator-item-actions .el-button--primary.link {
  color: #94a3b8 !important;
}

:root[data-theme='dark'] .generator-item-actions .el-button--primary.link:hover {
  color: #5eead4 !important;
  background: rgba(94, 234, 212, 0.08);
}

:root[data-theme='dark'] .generator-item-actions .el-button--primary.link .el-icon {
  color: #94a3b8 !important;
}

:root[data-theme='dark'] .generator-item-actions .el-button--primary.link:hover .el-icon {
  color: #5eead4 !important;
}

:root[data-theme='dark'] .generator-item-actions .el-button--danger.link {
  color: #94a3b8 !important;
}

:root[data-theme='dark'] .generator-item-actions .el-button--danger.link:hover {
  color: #f87171 !important;
  background: rgba(248, 113, 113, 0.08);
}

:root[data-theme='dark'] .generator-item-actions .el-button--danger.link .el-icon {
  color: #94a3b8 !important;
}

:root[data-theme='dark'] .generator-item-actions .el-button--danger.link:hover .el-icon {
  color: #f87171 !important;
}

:root[data-theme='dark'] .generator-item-actions .el-button.is-disabled {
  opacity: 0.35;
  background: transparent;
  border-color: transparent;
}

:root[data-theme='dark'] .generator-item-actions .el-button.is-disabled .el-icon {
  color: #475569 !important;
}

:root[data-theme='dark'] .generator-item-actions .el-button.is-disabled:hover {
  background: transparent;
  border-color: transparent;
}

:root[data-theme='dark'] .empty-list {
  color: #64748b;
}

:root[data-theme='dark'] .empty-list .el-icon {
  color: #475569;
}

:root[data-theme='dark'] .empty-list p {
  color: #64748b;
}

:root[data-theme='dark'] .generator-detail-panel {
  background-color: #1e293b;
}

:root[data-theme='dark'] .detail-content::-webkit-scrollbar-thumb {
  background: rgba(94, 234, 212, 0.25);
}

:root[data-theme='dark'] .detail-content::-webkit-scrollbar-thumb:hover {
  background: rgba(94, 234, 212, 0.4);
}

:root[data-theme='dark'] .detail-actions {
  border-top-color: rgba(94, 234, 212, 0.18);
}

:root[data-theme='dark'] .icon-picker-item {
  border-color: rgba(94, 234, 212, 0.18);
  background: rgba(15, 23, 42, 0.6);
}

:root[data-theme='dark'] .icon-picker-item:hover {
  border-color: #5eead4;
  background: rgba(94, 234, 212, 0.08);
}

:root[data-theme='dark'] .icon-picker-item.selected {
  border-color: #5eead4;
  background: rgba(94, 234, 212, 0.12);
  box-shadow: 0 0 0 2px rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .icon-picker-item .el-icon {
  color: #94a3b8;
}

:root[data-theme='dark'] .icon-picker-item.selected .el-icon {
  color: #5eead4;
}

:root[data-theme='dark'] .icon-picker-name {
  color: #64748b;
}

:root[data-theme='dark'] .icon-picker-item.selected .icon-picker-name {
  color: #5eead4;
}

/* 提示词选择弹窗 - 深色皮肤令牌覆盖 */
:root[data-theme='dark'] .prompt-select-dialog-modal {
  --ps-bg: #1e293b;
  --ps-bg-embedded: #1e293b;
  --ps-bg-soft: #334155;
  --ps-text-color: rgba(255, 255, 255, 0.88);
  --ps-text-color-secondary: rgba(203, 213, 225, 0.7);
  --ps-text-color-tertiary: rgba(148, 163, 184, 0.7);
  --ps-title-color: #f3f4f6;
  --ps-meta-color: rgba(148, 163, 184, 0.8);
  --ps-border-color: rgba(94, 234, 212, 0.18);
  --ps-border-color-soft: rgba(94, 234, 212, 0.12);
  --ps-divider: rgba(94, 234, 212, 0.18);
  --ps-shadow: 0 1px 2px -2px rgba(0, 0, 0, 0.5), 0 3px 6px 0 rgba(0, 0, 0, 0.4), 0 5px 12px 4px rgba(0, 0, 0, 0.32);
  --ps-shadow-card-hover: 0 2px 8px -2px rgba(0, 0, 0, 0.5), 0 4px 12px 0 rgba(0, 0, 0, 0.32);
  --ps-shadow-focus: 0 0 0 2px rgba(94, 234, 212, 0.28);
  --ps-accent: #5eead4;
  --ps-accent-hover: #2dd4bf;
  --ps-accent-soft-bg: rgba(94, 234, 212, 0.14);
  --ps-accent-soft-bg-hover: rgba(94, 234, 212, 0.22);
  --ps-accent-border: rgba(94, 234, 212, 0.32);
  --ps-danger: #f87171;
  --ps-danger-soft-bg: rgba(248, 113, 113, 0.12);
  --ps-danger-border: rgba(248, 113, 113, 0.32);
  --ps-card-bg: #1e293b;
  --ps-card-bg-hover: #334155;
  --ps-card-border: rgba(94, 234, 212, 0.18);
  --ps-card-border-hover: rgba(94, 234, 212, 0.4);
  --ps-close-color: rgba(203, 213, 225, 0.7);
  --ps-close-hover-bg: rgba(94, 234, 212, 0.12);
  --ps-scroll-thumb: rgba(94, 234, 212, 0.25);
  --ps-scroll-thumb-hover: rgba(94, 234, 212, 0.4);
  --ps-tag-bg-soft: rgba(94, 234, 212, 0.12);
  --ps-tag-bg-featured: rgba(251, 146, 60, 0.18);
  --ps-tag-color-featured: #fdba74;
  --ps-tag-border-featured: rgba(251, 146, 60, 0.4);
  --ps-fire-color: #fb923c;
  --ps-crown-color: #fcd34d;
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

/* === 固定提示词查看弹窗 - 深色皮肤令牌覆盖 === */
:root[data-theme='dark'] .fixed-prompt-dialog-modal {
  --fp-bg: #1e293b;
  --fp-bg-soft: rgba(15, 23, 42, 0.6);
  --fp-text-color: #f3f4f6;
  --fp-text-color-muted: rgba(203, 213, 225, 0.7);
  --fp-placeholder-color: #64748b;
  --fp-border-color: rgba(94, 234, 212, 0.18);
  --fp-border-color-hover: rgba(94, 234, 212, 0.32);
  --fp-shadow: 0 1px 2px -2px rgba(0, 0, 0, 0.5), 0 3px 6px 0 rgba(0, 0, 0, 0.4), 0 5px 12px 4px rgba(0, 0, 0, 0.32);
  --fp-accent: #5eead4;
  --fp-accent-hover: #2dd4bf;
  --fp-accent-text: #0f172a;
  --fp-btn-default-bg: rgba(30, 41, 59, 0.7);
  --fp-btn-default-border: rgba(94, 234, 212, 0.22);
  --fp-btn-default-text: #cbd5e1;
  --fp-btn-default-hover-bg: rgba(94, 234, 212, 0.12);
  --fp-btn-default-hover-border: rgba(94, 234, 212, 0.4);
  --fp-textarea-bg: rgba(15, 23, 42, 0.6);
  --fp-textarea-border: rgba(94, 234, 212, 0.18);
  --fp-textarea-color: #e2e8f0;
  --fp-close-color: rgba(203, 213, 225, 0.7);
  --fp-close-hover-color: #5eead4;
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
  background: #2dd4bf;
  color: #0f172a;
}

:root[data-theme='dark'] .view-detail-btn:hover {
  background: #5eead4;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.35);
}

:root[data-theme='dark'] .delete-btn {
  background: rgba(244, 63, 94, 0.15);
}

:root[data-theme='dark'] .history-header {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .followup-dialog-modal .el-dialog {
  background: #1e293b;
  border: 1px solid rgba(94, 234, 212, 0.3);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
}

:root[data-theme='dark'] .followup-dialog-modal .el-dialog__header {
  background: #1e293b;
  border-bottom-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .followup-dialog-modal .el-dialog__title {
  color: #5eead4;
}

:root[data-theme='dark'] .followup-dialog-modal .el-dialog__headerbtn .el-dialog__close {
  color: #9ca3af;
}

:root[data-theme='dark'] .followup-dialog-modal .el-dialog__headerbtn .el-dialog__close:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .followup-dialog-modal .el-dialog__body {
  background: #1e293b;
}

:root[data-theme='dark'] .followup-dialog-header {
  color: #e2e8f0;
}

:root[data-theme='dark'] .followup-messages {
  background: #111827;
}

:root[data-theme='dark'] .followup-message-bubble.assistant {
  background: #1e293b;
  border-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .followup-message-bubble.user {
  background: rgba(14, 165, 233, 0.15);
  border-color: rgba(14, 165, 233, 0.3);
}

:root[data-theme='dark'] .followup-message-bubble.system {
  background: rgba(168, 85, 247, 0.12);
  border-color: rgba(168, 85, 247, 0.25);
}

:root[data-theme='dark'] .followup-message-content {
  color: #e2e8f0;
}

:root[data-theme='dark'] .followup-input-area {
  background: #1e293b;
  border-top-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .followup-input :deep(.el-textarea__inner) {
  background: #0f172a;
  border-color: rgba(94, 234, 212, 0.2);
  color: #e2e8f0;
}

:root[data-theme='dark'] .followup-input :deep(.el-textarea__inner:focus) {
  border-color: #5eead4;
  box-shadow: 0 2px 8px rgba(94, 234, 212, 0.1);
}

:root[data-theme='dark'] .followup-input :deep(.el-textarea__inner::placeholder) {
  color: #64748b;
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

/* ── 生成器弹窗：暗色主题可见性强化 ─────────────────── */
:root[data-theme='dark'] .form-block-label {
  color: rgba(255, 255, 255, 0.95) !important;
}

:root[data-theme='dark'] .form-block-action {
  color: #5eead4 !important;
  background: transparent !important;
}

:root[data-theme='dark'] .form-block-action:hover {
  color: #2dd4bf !important;
  background: transparent !important;
}

:root[data-theme='dark'] .context-preview-item {
  background: rgba(94, 234, 212, 0.08);
  border-color: rgba(94, 234, 212, 0.25);
  color: rgba(255, 255, 255, 0.9);
}

:root[data-theme='dark'] .gen-footer-disclaimer {
  color: rgba(255, 255, 255, 0.7) !important;
}

:root[data-theme='dark'] .generator-dialog-modal .el-select__placeholder,
:root[data-theme='dark'] .generator-dialog-modal .el-select__placeholder.is-transparent {
  color: #94a3b8;
}

:root[data-theme='dark'] .generator-dialog-modal .el-input__inner,
:root[data-theme='dark'] .generator-dialog-modal .el-select__selected-item,
:root[data-theme='dark'] .generator-dialog-modal .el-textarea__inner {
  color: rgba(255, 255, 255, 0.92);
  -webkit-text-fill-color: rgba(255, 255, 255, 0.92);
}

:root[data-theme='dark'] .form-area .el-input__inner,
:root[data-theme='dark'] .form-area .el-select__placeholder,
:root[data-theme='dark'] .form-area .el-select__selected-item,
:root[data-theme='dark'] .form-area .el-textarea__inner {
  color: rgba(255, 255, 255, 0.92);
  -webkit-text-fill-color: rgba(255, 255, 255, 0.92);
}

:root[data-theme='dark'] .form-area .el-input__inner::placeholder,
:root[data-theme='dark'] .form-area .el-textarea__inner::placeholder {
  color: #64748b;
  -webkit-text-fill-color: #64748b;
}

:root[data-theme='dark'] .model-name,
:root[data-theme='dark'] .model-option .model-name {
  color: rgba(255, 255, 255, 0.92);
}

/* ── Select 下拉弹出层：暗色主题补全（popper 全链路） ── */
:root[data-theme='dark'] .generator-dialog-modal .el-select__wrapper {
  background-color: rgba(15, 23, 42, 0.6);
  box-shadow: 0 0 0 1px rgba(94, 234, 212, 0.18) inset;
}

:root[data-theme='dark'] .generator-dialog-modal .el-select__wrapper:hover {
  box-shadow: 0 0 0 1px rgba(94, 234, 212, 0.32) inset;
}

:root[data-theme='dark'] .generator-dialog-modal .el-select__wrapper.is-focused {
  box-shadow: 0 0 0 1px #5eead4 inset;
}

:root[data-theme='dark'] .generator-dialog-modal .el-popper.is-light,
:root[data-theme='dark'] .generator-dialog-modal .el-popper.is-light.el-select__popper,
:root[data-theme='dark'] .generator-dialog-modal .el-select__popper {
  background: #1e293b !important;
  border: 1px solid rgba(94, 234, 212, 0.18) !important;
}

:root[data-theme='dark'] .generator-dialog-modal .el-select-dropdown__wrap {
  background: #1e293b;
}

:root[data-theme='dark'] .generator-dialog-modal .el-select-dropdown .el-scrollbar__view,
:root[data-theme='dark'] .generator-dialog-modal .el-select-dropdown .el-scrollbar__bar {
  background: transparent;
}
</style>
