<template>
  <div :class="['character-library', { 'modal-mode': isModal }]">
    <div class="header-bar">
      <div class="header-left">
        <button type="button" class="back-home-btn" @click="handleClose">
          <span class="back-arrow">{{ isModal ? '关闭' : '返回' }}</span>
          <span class="back-divider"></span>
          <span class="page-title">角色库</span>
        </button>
      </div>
      <div class="header-center">
        <div class="action-buttons-row">
          <el-button @click="createNewCharacter">
            <el-icon><Plus /></el-icon>
            新建角色
          </el-button>
          <el-button @click="showImportDialog = true; fetchAllBooks()">
            <el-icon><Upload /></el-icon>
            他书导入
          </el-button>
          <el-button @click="handleBatchOperation">
            <el-icon><DocumentCopy /></el-icon>
            批量操作
          </el-button>
          <div class="search-wrapper">
            <el-button class="search-btn">
              <el-icon><Search /></el-icon>
              搜索角色
            </el-button>
            <div class="search-expand">
              <el-input
                v-model="searchKeyword"
                placeholder="搜索角色名称、性格、信息..."
                clearable
                @clear="searchKeyword = ''"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
              </el-input>
            </div>
          </div>
        </div>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="showRecognizeDialog = true">
          <el-icon><MagicStick /></el-icon>
          智能识别
        </el-button>
      </div>
    </div>

    <div class="main-content">
      <div class="left-panel">
        <div class="folder-filter">
          <div
            v-for="folder in folders"
            :key="folder"
            :class="['folder-item', { active: currentFolder === folder }]"
            @click="currentFolder = folder"
          >
            {{ folder }}
          </div>
          <div class="add-folder-btn" @click="showAddFolderDialog = true">
            <el-icon><Plus /></el-icon>
          </div>
        </div>
        <div class="character-list">
          <div
            v-for="character in filteredCharacters"
            :key="character.id"
            :class="['character-card', { active: currentCharacter?.id === character.id }]"
            @click="selectCharacter(character)"
          >
            <div class="character-name">{{ character.name }}</div>
            <div class="character-actions">
              <el-icon @click.stop="editCharacter(character)"><Edit /></el-icon>
              <el-icon @click.stop="viewCharacter(character)"><View /></el-icon>
            </div>
          </div>
        </div>
      </div>

      <div class="right-panel">
        <div v-if="currentCharacter" class="edit-form">
          <div class="form-item">
            <label>所属文件夹</label>
            <el-select v-model="currentCharacter.folder" placeholder="选择文件夹">
              <el-option
                v-for="folder in folders"
                :key="folder"
                :label="folder"
                :value="folder"
              />
            </el-select>
          </div>

          <div class="form-item">
            <label>角色名称 <span class="required">*</span></label>
            <el-input
              v-model="currentCharacter.name"
              placeholder="请输入角色名称"
              maxlength="15"
              show-word-limit
            />
          </div>

          <div class="form-item">
            <label>性别</label>
            <div class="gender-options">
              <div
                :class="['gender-option', { active: currentCharacter.gender === 'male' }]"
                @click="currentCharacter.gender = 'male'"
              >
                男
              </div>
              <div
                :class="['gender-option', { active: currentCharacter.gender === 'female' }]"
                @click="currentCharacter.gender = 'female'"
              >
                女
              </div>
              <div
                :class="['gender-option', { active: currentCharacter.gender === 'unknown' }]"
                @click="currentCharacter.gender = 'unknown'"
              >
                未知
              </div>
              <div
                :class="['gender-option', { active: currentCharacter.gender === 'none' }]"
                @click="currentCharacter.gender = 'none'"
              >
                无
              </div>
            </div>
          </div>

          <div class="form-item">
            <label>角色性格 <span class="required">*</span></label>
            <el-input
              v-model="currentCharacter.personality"
              type="textarea"
              :rows="3"
              placeholder="请输入角色性格"
              maxlength="100"
              show-word-limit
            />
          </div>

          <div class="form-item">
            <label>角色信息 <span class="required">*</span></label>
            <div class="info-tip">请根据剧情同步更新，仅填写剧情用得到的信息</div>
            <el-input
              v-model="currentCharacter.info"
              type="textarea"
              :rows="6"
              placeholder="请输入角色信息"
              maxlength="500"
              show-word-limit
            />
          </div>

          <div class="action-buttons">
            <el-button v-if="isModal" type="success" @click="insertCharacterToEditor(currentCharacter!)">
              <el-icon><Position /></el-icon>
              插入正文
            </el-button>
            <el-button @click="showMentionedChapters">提及章节</el-button>
            <el-button @click="exportCharacter">导出</el-button>
            <el-button type="danger" @click="deleteCharacter">删除</el-button>
            <el-button type="primary" @click="saveCharacter">保存</el-button>
          </div>
        </div>
        <div v-else class="empty-state">
          <el-empty description="请选择或创建角色" />
        </div>
      </div>
    </div>

    <el-dialog v-model="showAddFolderDialog" title="添加文件夹" width="400px">
      <el-input v-model="newFolderName" placeholder="请输入文件夹名称" />
      <template #footer>
        <el-button @click="showAddFolderDialog = false">取消</el-button>
        <el-button type="primary" @click="addFolder">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog 
      v-model="showBatchDialog" 
      title="批量操作" 
      width="900px"
      custom-class="batch-dialog"
    >
      <div class="batch-content">
        <div class="batch-tabs">
          <div
            :class="['batch-tab', { active: batchOperation === 'delete' }]"
            @click="batchOperation = 'delete'"
          >
            批量删除
          </div>
          <div
            :class="['batch-tab', { active: batchOperation === 'import' }]"
            @click="batchOperation = 'import'"
          >
            导入他书角色
          </div>
        </div>

        <div v-if="batchOperation === 'delete'" class="batch-panel">
          <div class="batch-header">
            <span class="batch-hint">从当前角色库中选择要删除的角色</span>
            <el-checkbox 
              :model-value="selectedBatchCharacters.length === filteredCharacters.length && filteredCharacters.length > 0"
              @change="handleSelectAllBatch"
            >
              全选
            </el-checkbox>
          </div>
          <div class="batch-list">
            <div
              v-for="character in filteredCharacters"
              :key="character.id"
              :class="['batch-item', { selected: selectedBatchCharacters.includes(character.id) }]"
            >
              <el-checkbox 
                :model-value="selectedBatchCharacters.includes(character.id)"
                @change="(val: boolean) => {
                  if (val) {
                    selectedBatchCharacters.push(character.id)
                  } else {
                    selectedBatchCharacters = selectedBatchCharacters.filter(id => id !== character.id)
                  }
                }"
              >
                <div class="batch-character-info">
                  <div class="batch-character-name">{{ character.name }}</div>
                  <div class="batch-character-detail">{{ character.personality }}</div>
                </div>
              </el-checkbox>
            </div>
          </div>
          <div class="batch-footer">
            <span class="batch-selected-count">已选择 {{ selectedBatchCharacters.length }} 个角色</span>
            <el-button type="danger" @click="handleBatchDelete" :disabled="selectedBatchCharacters.length === 0">
              删除选中的角色
            </el-button>
          </div>
        </div>

        <div v-else class="batch-panel">
          <div class="import-book-select">
            <el-select 
              v-model="selectedImportBook" 
              placeholder="选择要导入的书籍"
              style="width: 100%"
            >
              <el-option
                v-for="book in allBooks"
                :key="book.id"
                :label="book.title"
                :value="book.id"
              />
            </el-select>
          </div>
          <div v-if="importCharacters.length > 0" class="batch-header">
            <span class="batch-hint">从 {{ allBooks.find(b => b.id === selectedImportBook)?.title || '选中书籍' }} 中选择要导入的角色</span>
            <el-checkbox 
              :model-value="selectedImportCharacters.length === importCharacters.length && importCharacters.length > 0"
              @change="(val: boolean) => {
                if (val) {
                  selectedImportCharacters = importCharacters.map(c => c.id)
                } else {
                  selectedImportCharacters = []
                }
              }"
            >
              全选
            </el-checkbox>
          </div>
          <div v-if="importCharacters.length > 0" class="batch-list">
            <div
              v-for="character in importCharacters"
              :key="character.id"
              :class="['batch-item', { selected: selectedImportCharacters.includes(character.id) }]"
            >
              <el-checkbox 
                :model-value="selectedImportCharacters.includes(character.id)"
                @change="(val: boolean) => {
                  if (val) {
                    selectedImportCharacters.push(character.id)
                  } else {
                    selectedImportCharacters = selectedImportCharacters.filter(id => id !== character.id)
                  }
                }"
              >
                <div class="batch-character-info">
                  <div class="batch-character-name">{{ character.name }}</div>
                  <div class="batch-character-detail">{{ character.personality }}</div>
                </div>
              </el-checkbox>
            </div>
          </div>
          <el-empty v-else description="请先选择书籍" />
          <div v-if="importCharacters.length > 0" class="batch-footer">
            <span class="batch-selected-count">已选择 {{ selectedImportCharacters.length }} 个角色</span>
            <el-button type="primary" @click="handleBatchImport" :disabled="selectedImportCharacters.length === 0">
              导入选中的角色
            </el-button>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showBatchDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="showRecognizeDialog"
      title="智能识别"
      width="680px"
      :close-on-click-modal="false"
      custom-class="recognize-simple-dialog"
    >
      <!-- 顶部图标工具栏 -->
      <div class="recognize-toolbar">
        <!-- 提示词设置 -->
        <el-tooltip content="自定义识别提示词" placement="top">
          <el-button
            class="toolbar-btn"
            :class="{ active: showPromptEditor }"
            @click="showPromptEditor = !showPromptEditor"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
            </svg>
          </el-button>
        </el-tooltip>

        <!-- 模型选择 -->
        <el-tooltip content="选择 AI 模型" placement="top">
          <el-popover placement="bottom" trigger="click" :width="220">
            <template #reference>
              <el-button class="toolbar-btn">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="3"/>
                  <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
                </svg>
              </el-button>
            </template>
            <div class="popover-body">
              <div class="popover-label">AI 模型</div>
              <el-select v-model="selectedModelId" placeholder="选择模型" style="width:100%">
                <el-option v-for="m in models" :key="m.id" :label="m.name" :value="m.id" />
              </el-select>
            </div>
          </el-popover>
        </el-tooltip>

        <!-- 文字上限 -->
        <el-tooltip content="输入文字上限" placement="top">
          <el-popover placement="bottom" trigger="click" :width="280">
            <template #reference>
              <el-button class="toolbar-btn">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="4 7 4 4 20 4 20 7"/>
                  <line x1="9" y1="20" x2="15" y2="20"/>
                  <line x1="12" y1="4" x2="12" y2="20"/>
                </svg>
                <span class="toolbar-badge">{{ recognizeMaxLength }}</span>
              </el-button>
            </template>
            <div class="popover-body">
              <div class="popover-label">输入文字上限：{{ recognizeMaxLength }}</div>
              <el-slider v-model="recognizeMaxLength" :min="500" :max="10000" :step="500" />
            </div>
          </el-popover>
        </el-tooltip>

        <!-- 分隔线 -->
        <span class="toolbar-divider" />

        <!-- 恢复默认 -->
        <el-tooltip content="恢复默认提示词" placement="top">
          <el-button class="toolbar-btn" @click="resetRecognizePrompt">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="23 4 23 10 17 10"/>
              <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
            </svg>
          </el-button>
        </el-tooltip>
      </div>

      <!-- 展开的提示词编辑器 -->
      <transition name="prompt-slide">
        <div v-if="showPromptEditor" class="prompt-editor">
          <div class="prompt-editor-header">
            <span>自定义识别提示词</span>
            <span class="prompt-editor-hint">修改后自动保存到本地</span>
          </div>
          <el-input
            v-model="recognizePrompt"
            type="textarea"
            :rows="5"
            resize="vertical"
            placeholder="可自定义角色识别提示词，建议保留 JSON 输出要求"
          />
        </div>
      </transition>

      <!-- 主输入区：大输入框 -->
      <div class="recognize-main">
        <el-input
          v-model="recognizeText"
          type="textarea"
          :rows="12"
          resize="vertical"
          placeholder="把角色简介、片段设定或人物名单粘贴到这里&#10;尽量带上性格、外貌、身份或关系描述，可一次识别多个角色"
          :maxlength="recognizeMaxLength"
          show-word-limit
          class="recognize-textarea"
        />
      </div>

      <template #footer>
        <span class="footer-hint">提示词会保存到本地，下次沿用</span>
        <div class="footer-actions">
          <el-button @click="showRecognizeDialog = false">取消</el-button>
          <el-button type="primary" size="large" @click="startRecognize" :loading="recognizing">
            开始识别
          </el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="showImportDialog" title="他书导入" width="600px">
      <div class="import-content">
        <div class="book-select">
          <el-select v-model="selectedImportBook" placeholder="选择书籍">
            <el-option
              v-for="book in allBooks"
              :key="book.id"
              :label="book.title"
              :value="book.id"
            />
          </el-select>
        </div>
        <div v-if="importCharacters.length > 0" class="import-list">
          <el-checkbox-group v-model="selectedImportCharacters">
            <div v-for="char in importCharacters" :key="char.id" class="import-item">
              <el-checkbox :label="char.id">{{ char.name }}</el-checkbox>
            </div>
          </el-checkbox-group>
        </div>
      </div>
      <template #footer>
        <el-button @click="showImportDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmImport">导入</el-button>
      </template>
    </el-dialog>

    <el-dialog 
      v-model="showRecognizeResultDialog" 
      title="识别结果" 
      :width="dialogWidth" 
      center
      custom-class="recognize-dialog"
    >
      <div v-if="recognizedCharacters.length > 0" class="recognize-result">
        <div class="result-header">
          <div class="result-summary">
            <span class="result-kicker">AI识别完成</span>
            <div class="result-count-row">
              <span class="result-count">已识别 {{ recognizedCharacters.length }} 个角色</span>
              <span class="result-selected-count">已选择 {{ selectedRecognizedCount }} 个</span>
            </div>
          </div>
          <el-checkbox v-model="selectAllRecognized" @change="handleSelectAllRecognized">全选</el-checkbox>
        </div>
        <div class="result-list">
          <div
            v-for="(char, index) in recognizedCharacters"
            :key="index"
            :class="['result-item', { selected: char.selected }]"
          >
            <el-checkbox v-model="char.selected" :label="index" class="result-check">
              <div class="character-preview simple-preview">
                <div class="preview-name">{{ char.name || '未命名角色' }}</div>
              </div>
            </el-checkbox>
          </div>
        </div>
      </div>
      <el-empty v-else description="未识别到角色" />
      <template #footer>
        <el-button @click="showRecognizeResultDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmCreateRecognizedCharacters" :loading="creatingCharacters">
          创建选中的角色 ({{ selectedRecognizedCount }})
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useCharacterLibrary } from '@/composables/useCharacterLibrary'
import { Position } from '@element-plus/icons-vue'
import type { Character } from '@/types'

const props = defineProps<{
  bookId?: number
  isModal?: boolean
}>()

const emit = defineEmits<{
  close: []
  selectCharacter: [character: Character]
}>();

const {
  currentFolder, folders, characters, currentCharacter,
  showAddFolderDialog, newFolderName,
  showRecognizeDialog, recognizeText, recognizePrompt,
  recognizeMaxLength, recognizing, selectedModelId, models,
  showImportDialog, allBooks, selectedImportBook,
  importCharacters, selectedImportCharacters,
  showBatchDialog, selectedBatchCharacters, batchOperation,
  searchKeyword,
  showRecognizeResultDialog, recognizedCharacters,
  selectAllRecognized, creatingCharacters,
  selectedRecognizedCount, dialogWidth, filteredCharacters,
  fetchCharacters, selectCharacter, editCharacter, viewCharacter,
  createNewCharacter, saveCharacter, deleteCharacter,
  addFolder, handleClose,
  showMentionedChapters, exportCharacter, aiGenerateCharacter,
  startRecognize, confirmCreateRecognizedCharacters,
  fetchAllBooks, fetchImportCharacters, confirmImport,
  handleBatchOperation, handleSelectAllBatch, handleBatchDelete,
  handleBatchImport, resetRecognizePrompt,
  handleSelectAllRecognized, insertCharacterToEditor,
  normalizeGender, getGenderText, getEffectiveRecognizePrompt,
} = useCharacterLibrary({
  bookId: props.bookId,
  onClose: () => emit('close'),
  onSelectCharacter: (character: Character) => emit('selectCharacter', character),
})

const showPromptEditor = ref(false)
</script>

<style scoped>
.character-library {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background:
    radial-gradient(circle at top left, rgba(255, 255, 255, 0.72), transparent 35%),
    linear-gradient(180deg, #eef9eb 0%, #e4f5df 100%);
  z-index: 2000;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* ========== 弹窗模式覆盖 ========== */
.character-library.modal-mode {
  position: relative;
  width: 100%;
  height: 100%;
  background: #fff;
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.character-library.modal-mode .header-bar {
  border-radius: 16px 16px 0 0;
  padding: 10px 16px;
  flex-wrap: nowrap;
  gap: 8px;
}

.character-library.modal-mode .back-home-btn {
  padding: 6px 12px;
  border-radius: 12px;
  gap: 8px;
}

.character-library.modal-mode .page-title {
  font-size: 15px;
}

.character-library.modal-mode .action-buttons-row {
  gap: 6px;
  flex-wrap: nowrap;
}

.character-library.modal-mode .action-buttons-row .el-button {
  height: 32px;
  font-size: 12px;
  padding: 0 12px;
  gap: 4px;
}

.character-library.modal-mode .search-wrapper .search-btn {
  display: none;
}

.character-library.modal-mode .search-expand {
  position: relative;
  left: auto;
  top: auto;
  transform: none;
  width: 180px !important;
  box-shadow: none;
}

.character-library.modal-mode .search-expand :deep(.el-input) {
  width: 180px;
}

.character-library.modal-mode .header-right .el-button {
  height: 32px;
  font-size: 12px;
  padding: 0 12px;
  gap: 4px;
}

/* 分栏比例调整：左 30% / 右 70% */
.character-library.modal-mode .left-panel {
  width: 30%;
  min-width: 240px;
  max-width: 320px;
}

.character-library.modal-mode .right-panel {
  flex: 1;
  min-width: 0;
  padding: 16px 20px;
}

/* 表单紧凑 */
.character-library.modal-mode .edit-form {
  max-width: 100%;
}

.character-library.modal-mode .form-item {
  margin-bottom: 12px;
}

.character-library.modal-mode .form-item label {
  margin-bottom: 4px;
  font-size: 13px;
}

.character-library.modal-mode .gender-option {
  padding: 4px 14px;
  font-size: 13px;
}

.character-library.modal-mode .action-buttons {
  margin-top: 16px;
  padding-top: 12px;
}

.character-library.modal-mode .action-buttons .el-button {
  height: 30px;
  font-size: 12px;
  padding: 0 14px;
}

.page-title {
  font-size: 18px;
  font-weight: 700;
  color: #21461d;
  letter-spacing: 0.04em;
}

.back-home-btn {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 12px 18px;
  border: 1px solid rgba(82, 196, 26, 0.18);
  border-radius: 18px;
  background:
    linear-gradient(145deg, rgba(255, 255, 255, 0.96) 0%, rgba(232, 245, 233, 0.92) 100%);
  box-shadow: 0 10px 24px rgba(82, 196, 26, 0.12);
  cursor: pointer;
  transition: all 0.25s ease;
}

.back-home-btn:hover {
  transform: translateY(-1px);
  border-color: rgba(82, 196, 26, 0.3);
  box-shadow: 0 14px 28px rgba(82, 196, 26, 0.16);
}

.back-home-btn:active {
  transform: translateY(0);
}

.back-arrow {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 700;
  color: #4f8f1d;
}

.back-arrow::before {
  content: '<';
  font-size: 16px;
  line-height: 1;
}

.back-divider {
  width: 1px;
  height: 18px;
  background: rgba(79, 143, 29, 0.18);
}

.folder-filter {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
  padding: 16px 14px 12px;
  background: rgba(232, 245, 233, 0.62);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid rgba(180, 220, 180, 0.4);
}

.left-panel .folder-filter {
  position: sticky;
  top: 0;
  z-index: 2;
}

.folder-item {
  padding: 8px 16px;
  border-radius: 999px;
  cursor: pointer;
  font-size: 14px;
  color: #606266;
  background: linear-gradient(135deg, #f5fbf4 0%, #e8f5e9 100%);
  transition: all 0.3s;
  white-space: nowrap;
}

.folder-item:hover {
  background: linear-gradient(135deg, #e8f5e9 0%, #d4edc4 100%);
  transform: translateY(-2px);
}

.folder-item.active {
  background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%);
  color: #fff;
  box-shadow: 0 2px 8px rgba(82, 196, 26, 0.25);
}

.add-folder-btn {
  width: 32px;
  height: 32px;
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  cursor: pointer;
  background: linear-gradient(135deg, #f5fbf4 0%, #e8f5e9 100%);
  transition: all 0.3s;
}

.add-folder-btn:hover {
  background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%);
  color: #fff;
  transform: scale(1.08);
  box-shadow: 0 2px 8px rgba(82, 196, 26, 0.25);
}

.main-content {
  flex: 1;
  min-height: 0;
  display: flex;
  overflow: hidden;
  margin-top: 18px;
}

.left-panel {
  width: 25%;
  min-width: 260px;
  min-height: 0;
  background: rgba(232, 245, 233, 0.84);
  backdrop-filter: blur(10px);
  border-right: 1px solid rgba(180, 220, 180, 0.4);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.character-list {
  flex: 1;
  min-height: 0;
  padding: 12px;
  overflow-y: auto;
}

.character-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  margin-bottom: 8px;
  background: linear-gradient(135deg, #f5fbf4 0%, #e8f5e9 100%);
  border-radius: 14px;
  cursor: pointer;
  transition: all 0.3s;
}

.character-card:hover {
  background: linear-gradient(135deg, #e8f5e9 0%, #d4edc4 100%);
  transform: translateX(4px);
}

.character-card.active {
  background: linear-gradient(135deg, #c8e6c9 0%, #a5d6a7 100%);
  border: 1px solid #81c784;
  box-shadow: 0 2px 8px rgba(82, 196, 26, 0.15);
}

.character-name {
  font-size: 14px;
  color: #333;
}

.character-actions {
  display: flex;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.3s;
}

.character-card:hover .character-actions {
  opacity: 1;
}

.character-actions .el-icon {
  cursor: pointer;
  color: #909399;
  transition: color 0.3s;
}

.character-actions .el-icon:hover {
  color: #52c41a;
}

.right-panel {
  flex: 1;
  min-width: 0;
  min-height: 0;
  background: rgba(232, 245, 233, 0.84);
  backdrop-filter: blur(10px);
  overflow-y: auto;
  padding: 20px;
}

.edit-form {
  max-width: 800px;
}

.form-item {
  margin-bottom: 20px;
}

.form-item label {
  display: block;
  font-size: 14px;
  color: #333;
  margin-bottom: 8px;
  font-weight: 500;
}

.required {
  color: #f56c6c;
}

.gender-options {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.gender-option {
  padding: 8px 20px;
  border-radius: 999px;
  cursor: pointer;
  font-size: 14px;
  color: #606266;
  background: linear-gradient(135deg, #f5fbf4 0%, #e8f5e9 100%);
  transition: all 0.3s;
}

.gender-option:hover {
  background: linear-gradient(135deg, #e8f5e9 0%, #d4edc4 100%);
  transform: translateY(-2px);
}

.gender-option.active {
  background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%);
  color: #fff;
  box-shadow: 0 2px 8px rgba(82, 196, 26, 0.25);
}

.info-tip {
  font-size: 12px;
  color: #909399;
  margin-bottom: 8px;
}

.action-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
  flex-wrap: wrap;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.import-content {
  min-height: 200px;
}

.book-select {
  margin-bottom: 20px;
}

.import-list {
  max-height: 300px;
  overflow-y: auto;
}

.import-item {
  padding: 8px 0;
}

.recognize-result {
  max-height: 500px;
  overflow-y: auto;
}

.result-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 4px 0 16px;
  border-bottom: 1px solid rgba(130, 210, 120, 0.2);
  margin-bottom: 16px;
}

.result-summary {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.result-kicker {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  padding: 5px 10px;
  border-radius: 999px;
  background: rgba(82, 196, 26, 0.12);
  color: #4f8f1d;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.result-count-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

.result-count {
  font-size: 18px;
  font-weight: 700;
  color: #21461d;
}

.result-selected-count {
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(33, 70, 29, 0.06);
  font-size: 13px;
  color: #688063;
}

.result-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 10px;
}

.result-item {
  border-radius: 14px;
  border: 1px solid rgba(130, 210, 120, 0.18);
  background: linear-gradient(145deg, #f8fcf7 0%, #edf8ea 100%);
  transition: all 0.25s ease;
}

.result-item:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(82, 196, 26, 0.1);
  border-color: rgba(82, 196, 26, 0.24);
}

.result-item.selected {
  border-color: rgba(82, 196, 26, 0.45);
  box-shadow: 0 8px 18px rgba(82, 196, 26, 0.14);
}

.result-check {
  width: 100%;
  padding: 12px 14px;
}

.result-check :deep(.el-checkbox__label) {
  display: block;
  width: 100%;
  padding-left: 12px;
}

.result-check :deep(.el-checkbox__input) {
  align-self: flex-start;
}

.character-preview {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.simple-preview {
  min-height: 24px;
  justify-content: center;
}

.preview-name {
  font-size: 15px;
  font-weight: 600;
  line-height: 1.4;
  color: #1e3f1b;
  word-break: break-word;
}

.recognize-dialog {
  max-height: 50vh;
}

.recognize-dialog .el-dialog__body {
  max-height: 40vh;
  overflow-y: auto;
}

@media (max-width: 1200px) {
  .header-bar {
    padding: 12px 16px;
    flex-wrap: wrap;
  }

  .header-center {
    order: 3;
    width: 100%;
    justify-content: flex-start;
    margin-top: 12px;
  }

  .action-buttons-row {
    flex-wrap: wrap;
  }

  .action-buttons-row .el-button {
    flex: 1;
    min-width: 120px;
  }
}

@media (max-width: 900px) {
  .character-library {
    position: relative;
    min-height: 100vh;
  }

  .folder-filter {
    height: auto;
  }

  .result-header {
    flex-direction: column;
    align-items: stretch;
  }

  .result-list {
    grid-template-columns: 1fr;
  }

  .main-content {
    flex-direction: column;
    overflow: auto;
  }

  .left-panel {
    width: 100%;
    min-width: 0;
    max-height: 35vh;
    border-right: 0;
    border-bottom: 1px solid rgba(180, 220, 180, 0.4);
  }
}

.header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: rgba(255, 255, 255, 0.6);
  border-bottom: 1px solid rgba(130, 210, 120, 0.2);
  gap: 20px;
}

.header-left {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.header-center {
  flex: 1;
  display: flex;
  justify-content: center;
  min-width: 0;
}

.header-right {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.action-buttons-row {
  display: flex;
  gap: 12px;
  flex-wrap: nowrap;
}

.action-buttons-row .el-button {
  height: 40px;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  white-space: nowrap;
}

.search-wrapper {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.search-wrapper .search-btn {
  height: 40px;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  white-space: nowrap;
  position: relative;
  z-index: 11;
}

.search-expand {
  position: absolute;
  left: 100%;
  margin-left: 8px;
  top: 50%;
  transform: translateY(-50%);
  width: 0;
  overflow: hidden;
  transition: width 0.3s ease;
  z-index: 10;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.search-wrapper:hover .search-expand {
  width: 300px;
}

.search-expand :deep(.el-input) {
  width: 100%;
}

.search-expand :deep(.el-input__wrapper) {
  border-radius: 8px;
  box-shadow: 0 0 0 1px rgba(130, 210, 120, 0.3) inset;
  transition: box-shadow 0.3s;
}

.search-expand :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px rgba(82, 196, 26, 0.5) inset;
}

.search-expand :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #52c41a inset, 0 0 0 3px rgba(82, 196, 26, 0.1);
}

.search-expand :deep(.el-input__inner) {
  padding-left: 35px;
}

@media (max-width: 1200px) {
  .search-wrapper:hover .search-expand {
    width: 250px;
  }
}

@media (max-width: 768px) {
  .search-wrapper {
    width: 100%;
  }
  
  .search-wrapper .search-btn {
    display: none;
  }
  
  .search-expand {
    position: relative;
    right: auto;
    top: auto;
    transform: none;
    width: 100% !important;
    box-shadow: none;
  }
}

.batch-dialog .el-dialog__body {
  padding: 0;
}

.batch-content {
  max-height: 70vh;
  overflow-y: auto;
}

.batch-tabs {
  display: flex;
  border-bottom: 1px solid rgba(130, 210, 120, 0.2);
  background: rgba(255, 255, 255, 0.8);
  position: sticky;
  top: 0;
  z-index: 1;
}

.batch-tab {
  flex: 1;
  padding: 16px 24px;
  text-align: center;
  cursor: pointer;
  font-size: 15px;
  font-weight: 500;
  color: #606266;
  transition: all 0.3s;
  border-bottom: 2px solid transparent;
}

.batch-tab:hover {
  color: #52c41a;
  background: rgba(82, 196, 26, 0.05);
}

.batch-tab.active {
  color: #52c41a;
  border-bottom-color: #52c41a;
  background: rgba(82, 196, 26, 0.08);
}

.batch-panel {
  padding: 20px;
}

.batch-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: rgba(232, 245, 233, 0.6);
  border-radius: 12px;
  margin-bottom: 16px;
}

.batch-hint {
  font-size: 14px;
  color: #4e6a4a;
}

.batch-list {
  max-height: 400px;
  overflow-y: auto;
  display: grid;
  gap: 10px;
}

.batch-item {
  padding: 14px 16px;
  border-radius: 12px;
  border: 1px solid rgba(130, 210, 120, 0.2);
  background: rgba(255, 255, 255, 0.8);
  transition: all 0.3s;
}

.batch-item:hover {
  background: rgba(232, 245, 233, 0.8);
  border-color: rgba(82, 196, 26, 0.3);
  transform: translateX(4px);
}

.batch-item.selected {
  background: rgba(82, 196, 26, 0.1);
  border-color: rgba(82, 196, 26, 0.5);
  box-shadow: 0 2px 8px rgba(82, 196, 26, 0.15);
}

.batch-item :deep(.el-checkbox) {
  width: 100%;
}

.batch-item :deep(.el-checkbox__label) {
  width: 100%;
  padding-left: 10px;
}

.batch-character-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.batch-character-name {
  font-size: 15px;
  font-weight: 600;
  color: #21461d;
}

.batch-character-detail {
  font-size: 13px;
  color: #688063;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.batch-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: rgba(255, 255, 255, 0.8);
  border-top: 1px solid rgba(130, 210, 120, 0.2);
  margin-top: 16px;
  border-radius: 12px;
}

.batch-selected-count {
  font-size: 14px;
  color: #606266;
}

.import-book-select {
  margin-bottom: 16px;
}

@media (max-width: 768px) {
  .batch-list {
    max-height: 300px;
  }
  
  .batch-footer {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }
  
  .batch-footer .el-button {
    width: 100%;
  }
}

/* 暗色主题适配 */
:root[data-theme='dark'] .character-library {
  background: linear-gradient(180deg, #0f172a 0%, #1e293b 100%);
}

:root[data-theme='dark'] .header-bar {
  background: rgba(30, 41, 59, 0.8);
  border-bottom-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .page-title {
  color: #f3f4f6;
}

:root[data-theme='dark'] .back-home-btn {
  border-color: rgba(71, 85, 105, 0.4);
  background: linear-gradient(145deg, rgba(30, 41, 59, 0.96) 0%, rgba(51, 65, 85, 0.92) 100%);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .back-home-btn:hover {
  border-color: rgba(94, 234, 212, 0.4);
  box-shadow: 0 14px 28px rgba(94, 234, 212, 0.15);
}

:root[data-theme='dark'] .back-arrow {
  color: #5eead4;
}

:root[data-theme='dark'] .back-divider {
  background: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .folder-filter {
  background: rgba(30, 41, 59, 0.6);
  border-bottom-color: rgba(71, 85, 105, 0.3);
}

:root[data-theme='dark'] .folder-item {
  color: #d1d5db;
  background: rgba(51, 65, 85, 0.6);
}

:root[data-theme='dark'] .folder-item:hover {
  background: rgba(71, 85, 105, 0.6);
}

:root[data-theme='dark'] .folder-item.active {
  background: #00c9a7;
  color: #fff;
}

:root[data-theme='dark'] .search-expand {
  background: #1e293b;
}

:root[data-theme='dark'] .left-panel {
  background: rgba(30, 41, 59, 0.8);
  border-right-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .character-card {
  background: rgba(51, 65, 85, 0.6);
}

:root[data-theme='dark'] .character-card:hover {
  background: rgba(71, 85, 105, 0.6);
}

:root[data-theme='dark'] .character-card.active {
  background: rgba(0, 201, 167, 0.2);
  border-color: rgba(94, 234, 212, 0.4);
  box-shadow: 0 2px 8px rgba(94, 234, 212, 0.15);
}

:root[data-theme='dark'] .character-name {
  color: #f3f4f6;
}

:root[data-theme='dark'] .character-actions .el-icon {
  color: #9ca3af;
}

:root[data-theme='dark'] .character-actions .el-icon:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .right-panel {
  background: rgba(30, 41, 59, 0.8);
}

:root[data-theme='dark'] .form-item label {
  color: #e5e7eb;
}

:root[data-theme='dark'] .form-item :deep(.el-input__wrapper),
:root[data-theme='dark'] .form-item :deep(.el-textarea__inner) {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(71, 85, 105, 0.4);
  box-shadow: none;
}

:root[data-theme='dark'] .form-item :deep(.el-input__inner),
:root[data-theme='dark'] .form-item :deep(.el-textarea__inner) {
  color: #f3f4f6;
}

:root[data-theme='dark'] .character-library.modal-mode {
  background: #1e293b;
}

/* ========================================
   智能识别 · 简洁版样式
   ======================================== */

.recognize-simple-dialog .el-dialog {
  border-radius: 20px;
  overflow: hidden;
  box-shadow:
    0 24px 48px rgba(0, 0, 0, 0.10),
    0 8px 16px rgba(0, 0, 0, 0.04);
}

.recognize-simple-dialog .el-dialog__header {
  padding: 18px 24px 12px;
  border-bottom: 1px solid rgba(130, 210, 120, 0.12);
}

.recognize-simple-dialog .el-dialog__title {
  font-size: 17px;
  font-weight: 700;
  color: #21461d;
}

.recognize-simple-dialog .el-dialog__body {
  padding: 12px 24px 8px;
}

.recognize-simple-dialog .el-dialog__footer {
  padding: 12px 24px 18px;
  border-top: 1px solid rgba(130, 210, 120, 0.12);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

/* ---- 顶部工具栏 ---- */
.recognize-toolbar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 0 10px;
  border-bottom: 1px solid rgba(130, 210, 120, 0.12);
  margin-bottom: 10px;
}

.toolbar-btn {
  width: 34px;
  height: 34px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  border: 1px solid transparent;
  background: transparent;
  color: #688063;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
}

.toolbar-btn:hover {
  background: rgba(82, 196, 26, 0.08);
  color: #52c41a;
  border-color: rgba(82, 196, 26, 0.15);
}

.toolbar-btn.active {
  background: rgba(82, 196, 26, 0.10);
  color: #52c41a;
  border-color: rgba(82, 196, 26, 0.20);
}

.toolbar-badge {
  position: absolute;
  top: -4px;
  right: -6px;
  font-size: 9px;
  font-weight: 700;
  background: #52c41a;
  color: #fff;
  padding: 0 5px;
  border-radius: 8px;
  line-height: 15px;
  min-width: 16px;
  text-align: center;
}

.toolbar-divider {
  width: 1px;
  height: 18px;
  background: rgba(130, 210, 120, 0.2);
  margin: 0 4px;
}

/* ---- Popover 内部 ---- */
.popover-body {
  padding: 4px 0;
}

.popover-label {
  font-size: 13px;
  font-weight: 600;
  color: #355430;
  margin-bottom: 10px;
}

/* ---- 提示词展开编辑器 ---- */
.prompt-editor {
  padding: 0 0 10px;
}

.prompt-editor-header {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  font-weight: 600;
  color: #355430;
  margin-bottom: 8px;
}

.prompt-editor-hint {
  font-weight: 400;
  font-size: 12px;
  color: #95a890;
}

/* ---- 提示词展开/收起动画 ---- */
.prompt-slide-enter-active {
  animation: prompt-in 0.25s cubic-bezier(0.22, 1, 0.36, 1);
}
.prompt-slide-leave-active {
  animation: prompt-in 0.2s cubic-bezier(0.22, 1, 0.36, 1) reverse;
}

@keyframes prompt-in {
  0%   { opacity: 0; max-height: 0; margin-bottom: 0; }
  100% { opacity: 1; max-height: 200px; margin-bottom: 0; }
}

/* ---- 主输入框 ---- */
.recognize-main {
  min-height: 0;
}

.recognize-textarea {
  width: 100%;
}

.recognize-textarea :deep(.el-textarea__inner) {
  min-height: 280px;
  border-radius: 14px;
  border: 1px solid rgba(130, 210, 120, 0.20);
  background: #fafcfa;
  padding: 16px 18px;
  font-size: 14px;
  line-height: 1.7;
  color: #333;
  transition:
    border-color 0.25s ease,
    box-shadow 0.25s ease;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.02);
  resize: vertical;
}

.recognize-textarea :deep(.el-textarea__inner):focus {
  border-color: #52c41a;
  box-shadow: 0 0 0 3px rgba(82, 196, 26, 0.08);
  background: #fff;
}

.recognize-textarea :deep(.el-textarea__inner::placeholder) {
  color: #b8c4b4;
  line-height: 1.7;
}

.recognize-textarea :deep(.el-input__count) {
  font-size: 12px;
  color: #95a890;
  background: transparent;
  bottom: 8px;
  right: 12px;
}

/* ---- 底部 ---- */
.footer-hint {
  font-size: 12px;
  color: #95a890;
}

.footer-actions {
  display: flex;
  gap: 10px;
}

.footer-actions .el-button--primary {
  padding: 0 28px;
  border-radius: 10px;
  font-weight: 600;
}

/* 暗色适配 */
:root[data-theme='dark'] .recognize-simple-dialog .el-dialog__title {
  color: #f3f4f6;
}

:root[data-theme='dark'] .toolbar-btn {
  color: #9ca3af;
}

:root[data-theme='dark'] .toolbar-btn:hover {
  color: #5eead4;
  background: rgba(94, 234, 212, 0.10);
  border-color: rgba(94, 234, 212, 0.15);
}

:root[data-theme='dark'] .toolbar-btn.active {
  color: #5eead4;
  background: rgba(94, 234, 212, 0.12);
  border-color: rgba(94, 234, 212, 0.20);
}

:root[data-theme='dark'] .toolbar-badge {
  background: #5eead4;
  color: #0f172a;
}

:root[data-theme='dark'] .toolbar-divider {
  background: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .popover-label {
  color: #d1d5db;
}

:root[data-theme='dark'] .prompt-editor-header {
  color: #d1d5db;
}

:root[data-theme='dark'] .prompt-editor-hint {
  color: #6b7280;
}

:root[data-theme='dark'] .recognize-textarea :deep(.el-textarea__inner) {
  background: rgba(30, 41, 59, 0.6);
  border-color: rgba(71, 85, 105, 0.4);
  color: #f3f4f6;
}

:root[data-theme='dark'] .recognize-textarea :deep(.el-textarea__inner):focus {
  border-color: #5eead4;
  box-shadow: 0 0 0 3px rgba(94, 234, 212, 0.10);
  background: rgba(30, 41, 59, 0.8);
}

:root[data-theme='dark'] .recognize-textarea :deep(.el-textarea__inner::placeholder) {
  color: #6b7280;
}

:root[data-theme='dark'] .recognize-textarea :deep(.el-input__count) {
  color: #6b7280;
}

:root[data-theme='dark'] .footer-hint {
  color: #6b7280;
}
</style>
