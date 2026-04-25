<template>
  <div class="character-library">
    <div class="header-bar">
      <div class="header-left">
        <button type="button" class="back-home-btn" @click="handleClose">
          <span class="back-arrow">返回</span>
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
      width="800px"
      :close-on-click-modal="false"
      custom-class="recognize-panel-dialog"
    >
      <div class="recognize-panel-content">
        <div class="recognize-summary">
          <span class="summary-tag">角色库快捷入口</span>
          <p>把角色简介、片段设定或人物名单粘贴进来，一次识别多个角色，再批量加入角色库。</p>
        </div>

        <div class="recognize-settings">
          <div class="setting-item setting-slider">
            <span class="setting-label">输入文字上限</span>
            <el-slider
              v-model="recognizeMaxLength"
              :min="500"
              :max="10000"
              :step="500"
              show-input
              :show-input-controls="false"
            />
          </div>
          <div class="setting-item setting-select">
            <span class="setting-label">AI 模型</span>
            <el-select
              v-model="selectedModelId"
              placeholder="选择模型"
              size="small"
            >
              <el-option
                v-for="model in models"
                :key="model.id"
                :label="model.name"
                :value="model.id"
              />
            </el-select>
          </div>
        </div>

        <div class="recognize-workbench">
          <div class="recognize-card">
            <div class="recognize-card-head">
              <div>
                <h3>待识别文本</h3>
                <p>支持一次输入多个角色，尽量带上性格、外貌、身份或关系描述。</p>
              </div>
            </div>
            <el-input
              v-model="recognizeText"
              type="textarea"
              :rows="7"
              resize="vertical"
              placeholder="请输入待识别角色简介，可一次识别多个角色"
              :maxlength="recognizeMaxLength"
              show-word-limit
            />
          </div>

          <div class="recognize-card">
            <div class="recognize-card-head">
              <div>
                <h3>识别提示词</h3>
                <p>这里可以人工自定义识别规则。建议保留 JSON 输出要求，避免结果无法解析。</p>
              </div>
              <el-button link type="primary" @click="resetRecognizePrompt">恢复默认</el-button>
            </div>
            <el-input
              v-model="recognizePrompt"
              type="textarea"
              :rows="10"
              resize="vertical"
              placeholder="可自定义角色识别提示词"
            />
          </div>
        </div>

        <div class="recognize-actions-bar">
          <span class="recognize-hint">当前提示词会自动保存到本地，下次进入角色库会继续沿用。</span>
          <el-button type="primary" size="large" @click="startRecognize" :loading="recognizing">
            开始识别
          </el-button>
        </div>
      </div>
      <template #footer>
        <el-button @click="showRecognizeDialog = false">关闭</el-button>
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
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { aiAPI, bookAPI, characterAPI, configAPI } from '@/api'
import type { ApiModel, Book, Character } from '@/types'

const DEFAULT_RECOGNIZE_PROMPT = `你是一个专业的角色识别助手。请从给定的文本中识别出所有角色，并为每个角色提取以下信息：
1. 角色名称
2. 性别（male / female / unknown / none）
3. 角色性格
4. 角色信息（仅填写剧情中明确给出的信息）

请只返回 JSON，格式如下：
{
  "characters": [
    {
      "name": "角色名称",
      "gender": "male|female|unknown|none",
      "personality": "角色性格描述",
      "info": "角色信息"
    }
  ]
}

注意：
- 只返回 JSON，不要包含额外说明
- 如果无法确定性别，请使用 "unknown"
- 性格和信息保持简洁
- 如果文本中没有明确角色，请返回空数组`

const RECOGNIZE_PROMPT_STORAGE_KEY = 'character-library-recognize-prompt'

const loadRecognizePrompt = () => {
  if (typeof window === 'undefined') {
    return DEFAULT_RECOGNIZE_PROMPT
  }

  const storedPrompt = window.localStorage.getItem(RECOGNIZE_PROMPT_STORAGE_KEY)
  return storedPrompt?.trim() ? storedPrompt : DEFAULT_RECOGNIZE_PROMPT
}

const route = useRoute()
const router = useRouter()

const bookId = parseInt(route.params.bookId as string)

const currentFolder = ref('全部')
const folders = ref<string[]>(['全部'])
const characters = ref<Character[]>([])
const currentCharacter = ref<Character | null>(null)
const showAddFolderDialog = ref(false)
const newFolderName = ref('')
const showRecognizeDialog = ref(false)
const recognizeText = ref('')
const recognizePrompt = ref(loadRecognizePrompt())
const recognizeMaxLength = ref(3000)
const recognizing = ref(false)

const showImportDialog = ref(false)
const allBooks = ref<Book[]>([])
const selectedImportBook = ref<number>()
const importCharacters = ref<Character[]>([])
const selectedImportCharacters = ref<number[]>([])

const showBatchDialog = ref(false)
const selectedBatchCharacters = ref<number[]>([])
const batchOperation = ref<'delete' | 'import'>('delete')
const searchKeyword = ref('')

const showRecognizeResultDialog = ref(false)
const recognizedCharacters = ref<any[]>([])
const selectAllRecognized = ref(false)
const creatingCharacters = ref(false)
const selectedModelId = ref<number | undefined>()
const models = ref<ApiModel[]>([])

const selectedRecognizedCount = computed(() => {
  return recognizedCharacters.value.filter(c => c.selected).length
})

const dialogWidth = computed(() => {
  const screenWidth = window.innerWidth
  const screenHeight = window.innerHeight
  const size = Math.min(screenWidth, screenHeight) * 0.5
  return `${size}px`
})

const getEffectiveRecognizePrompt = () => {
  return recognizePrompt.value.trim() || DEFAULT_RECOGNIZE_PROMPT
}

const resetRecognizePrompt = () => {
  recognizePrompt.value = DEFAULT_RECOGNIZE_PROMPT
  ElMessage.success('已恢复默认识别提示词')
}

const fetchModels = async () => {
  try {
    const res = await configAPI.getAll()
    if (res.success && res.data) {
      models.value = res.data
      // 选择默认模型
      const defaultModel = res.data.find(m => m.is_default)
      if (defaultModel) {
        selectedModelId.value = defaultModel.id
      }
    }
  } catch (error) {
    console.error('获取模型列表失败:', error)
  }
}

watch(selectedRecognizedCount, (newCount) => {
  if (newCount === 0) {
    selectAllRecognized.value = false
  } else if (newCount === recognizedCharacters.value.length) {
    selectAllRecognized.value = true
  }
})

watch(recognizePrompt, (value) => {
  if (typeof window === 'undefined') {
    return
  }
  window.localStorage.setItem(RECOGNIZE_PROMPT_STORAGE_KEY, value)
})

const filteredCharacters = computed(() => {
  let result = characters.value
  
  if (currentFolder.value !== '全部') {
    result = result.filter(c => c.folder === currentFolder.value)
  }
  
  if (searchKeyword.value.trim()) {
    const keyword = searchKeyword.value.trim().toLowerCase()
    result = result.filter(c => {
      return (
        c.name.toLowerCase().includes(keyword) ||
        (c.personality && c.personality.toLowerCase().includes(keyword)) ||
        (c.info && c.info.toLowerCase().includes(keyword))
      )
    })
  }
  
  return result
})

const fetchCharacters = async () => {
  try {
    const res = await characterAPI.getByBook(bookId)
    if (res.success && res.data) {
      characters.value = res.data
      const folderSet = new Set(['全部'])
      res.data.forEach(c => {
        if (c.folder) folderSet.add(c.folder)
      })
      folders.value = Array.from(folderSet)
    }
  } catch (error) {
    ElMessage.error('获取角色列表失败')
  }
}

const selectCharacter = (character: Character) => {
  currentCharacter.value = { ...character }
}

const editCharacter = (character: Character) => {
  currentCharacter.value = { ...character }
}

const viewCharacter = (character: Character) => {
  currentCharacter.value = { ...character }
}

const createNewCharacter = async () => {
  if (!bookId || isNaN(bookId)) {
    ElMessage.error('无效的书本ID，请从书本管理页面进入')
    return
  }
  
  try {
    const data = {
      book_id: bookId,
      name: '未命名角色',
      gender: 'unknown',
      personality: '',
      info: '',
      folder: currentFolder.value === '全部' ? '全部' : currentFolder.value,
      folders: []
    }
    const res = await characterAPI.create(data)
    if (res.success && res.data) {
      characters.value.unshift(res.data)
      currentCharacter.value = { ...res.data }
      ElMessage.success('创建成功')
    } else {
      ElMessage.error('创建角色失败: ' + (res.message || '未知错误'))
    }
  } catch (error: any) {
    console.error('Create character error:', error)
    ElMessage.error('创建角色失败: ' + (error.message || '未知错误'))
  }
}

const saveCharacter = async () => {
  if (!currentCharacter.value) return
  
  if (!currentCharacter.value.name.trim()) {
    ElMessage.warning('请输入角色名称')
    return
  }
  if (!currentCharacter.value.personality.trim()) {
    ElMessage.warning('请输入角色性格')
    return
  }
  if (!currentCharacter.value.info.trim()) {
    ElMessage.warning('请输入角色信息')
    return
  }

  try {
    const res = await characterAPI.update(currentCharacter.value.id, currentCharacter.value)
    if (res.success && res.data) {
      const index = characters.value.findIndex(c => c.id === currentCharacter.value!.id)
      if (index !== -1) {
        characters.value[index] = res.data
      }
      ElMessage.success('保存成功')
    }
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const deleteCharacter = async () => {
  if (!currentCharacter.value) return
  
  try {
    await ElMessageBox.confirm('确定要删除该角色吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const res = await characterAPI.delete(currentCharacter.value.id)
    if (res.success) {
      characters.value = characters.value.filter(c => c.id !== currentCharacter.value!.id)
      currentCharacter.value = null
      ElMessage.success('删除成功')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const addFolder = () => {
  if (!newFolderName.value.trim()) {
    ElMessage.warning('请输入文件夹名称')
    return
  }
  if (folders.value.includes(newFolderName.value.trim())) {
    ElMessage.warning('文件夹已存在')
    return
  }
  folders.value.push(newFolderName.value.trim())
  newFolderName.value = ''
  showAddFolderDialog.value = false
  ElMessage.success('添加成功')
}

const handleClose = () => {
  router.push(`/write/${bookId}`)
}

const showMentionedChapters = () => {
  ElMessage.info('提及章节功能开发中')
}

const exportCharacter = () => {
  if (!currentCharacter.value) return
  
  const content = `角色名称：${currentCharacter.value.name}
性别：${currentCharacter.value.gender === 'male' ? '男' : currentCharacter.value.gender === 'female' ? '女' : currentCharacter.value.gender === 'unknown' ? '未知' : '无'}
性格：${currentCharacter.value.personality}
信息：${currentCharacter.value.info}`
  
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${currentCharacter.value.name}.txt`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('导出成功')
}

const aiGenerateCharacter = () => {
  ElMessage.info('AI生成角色功能开发中')
}

const startRecognize = async () => {
  if (!recognizeText.value.trim()) {
    ElMessage.warning('请输入待识别的角色简介')
    return
  }
  
  if (models.value.length === 0) {
    ElMessage.warning('请先配置AI模型')
    return
  }
  
  recognizing.value = true
  try {
    const res = await aiAPI.recognizeCharacters({ 
      text: recognizeText.value,
      configId: selectedModelId.value,
      customPrompt: getEffectiveRecognizePrompt()
    })
    if (res.success && res.data && res.data.length > 0) {
      recognizedCharacters.value = res.data.map((char: any) => ({
        ...char,
        selected: true,
        gender: normalizeGender(char.gender)
      }))
      selectAllRecognized.value = true
      showRecognizeResultDialog.value = true
      ElMessage.success(`成功识别 ${res.data.length} 个角色`)
    } else {
      ElMessage.warning('未识别到角色，请尝试提供更详细的描述')
    }
  } catch (error) {
    console.error('识别失败:', error)
    ElMessage.error('识别失败，请检查AI配置或网络连接')
  } finally {
    recognizing.value = false
  }
}

const normalizeGender = (gender: string): 'male' | 'female' | 'unknown' | 'none' => {
  if (!gender) return 'unknown'
  const g = gender.toLowerCase()
  if (g === '男' || g === 'male') return 'male'
  if (g === '女' || g === 'female') return 'female'
  if (g === '无' || g === 'none') return 'none'
  return 'unknown'
}

const getGenderText = (gender: string): string => {
  const map: Record<string, string> = {
    'male': '男',
    'female': '女',
    'unknown': '未知',
    'none': '无'
  }
  return map[gender] || '未知'
}

const handleSelectAllRecognized = (val: boolean) => {
  recognizedCharacters.value.forEach(char => {
    char.selected = val
  })
}

const confirmCreateRecognizedCharacters = async () => {
  const selectedChars = recognizedCharacters.value.filter(c => c.selected)
  if (selectedChars.length === 0) {
    ElMessage.warning('请至少选择一个角色')
    return
  }
  
  creatingCharacters.value = true
  let successCount = 0
  let failCount = 0
  
  try {
    for (const char of selectedChars) {
      try {
        const data = {
          book_id: bookId,
          name: char.name || '未命名角色',
          gender: char.gender || 'unknown',
          personality: char.personality || '',
          info: char.info || '',
          folder: currentFolder.value === '全部' ? '全部' : currentFolder.value,
          folders: []
        }
        const res = await characterAPI.create(data)
        if (res.success && res.data) {
          successCount++
        } else {
          failCount++
        }
      } catch (error) {
        console.error('创建角色失败:', error)
        failCount++
      }
    }
    
    await fetchCharacters()
    
    if (failCount === 0) {
      ElMessage.success(`成功创建 ${successCount} 个角色`)
    } else {
      ElMessage.warning(`创建完成：成功 ${successCount} 个，失败 ${failCount} 个`)
    }
    
    showRecognizeResultDialog.value = false
    recognizeText.value = ''
  } catch (error) {
    ElMessage.error('批量创建角色失败')
  } finally {
    creatingCharacters.value = false
  }
}

const fetchAllBooks = async () => {
  try {
    const res = await bookAPI.getAll()
    if (res.success && res.data) {
      allBooks.value = res.data.filter(b => b.id !== bookId)
    }
  } catch (error) {
    console.error('获取书籍列表失败')
  }
}

const fetchImportCharacters = async () => {
  if (!selectedImportBook.value) return
  try {
    const res = await characterAPI.getByBook(selectedImportBook.value)
    if (res.success && res.data) {
      importCharacters.value = res.data
    }
  } catch (error) {
    ElMessage.error('获取角色列表失败')
  }
}

const confirmImport = async () => {
  if (selectedImportCharacters.value.length === 0) {
    ElMessage.warning('请选择要导入的角色')
    return
  }
  
  try {
    for (const charId of selectedImportCharacters.value) {
      const char = importCharacters.value.find(c => c.id === charId)
      if (char) {
        await characterAPI.create({
          book_id: bookId,
          name: char.name,
          gender: char.gender,
          personality: char.personality,
          info: char.info,
          folder: char.folder,
          folders: char.folders
        })
      }
    }
    await fetchCharacters()
    showImportDialog.value = false
    selectedImportCharacters.value = []
    ElMessage.success('导入成功')
  } catch (error) {
    ElMessage.error('导入失败')
  }
}

watch(selectedImportBook, () => {
  if (showBatchDialog.value && batchOperation.value === 'import') {
    fetchImportCharacters()
  }
})

const handleBatchOperation = () => {
  if (characters.value.length === 0) {
    ElMessage.warning('当前没有角色可操作')
    return
  }
  showBatchDialog.value = true
  selectedBatchCharacters.value = []
  batchOperation.value = 'delete'
  fetchAllBooks()
}

const handleSelectAllBatch = (val: boolean) => {
  if (val) {
    selectedBatchCharacters.value = filteredCharacters.value.map(c => c.id)
  } else {
    selectedBatchCharacters.value = []
  }
}

const handleBatchDelete = async () => {
  if (selectedBatchCharacters.value.length === 0) {
    ElMessage.warning('请至少选择一个角色')
    return
  }
  
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedBatchCharacters.value.length} 个角色吗？`, '批量删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    let successCount = 0
    let failCount = 0
    
    for (const charId of selectedBatchCharacters.value) {
      try {
        const res = await characterAPI.delete(charId)
        if (res.success) {
          successCount++
        } else {
          failCount++
        }
      } catch (error) {
        failCount++
      }
    }
    
    await fetchCharacters()
    currentCharacter.value = null
    showBatchDialog.value = false
    selectedBatchCharacters.value = []
    
    if (failCount === 0) {
      ElMessage.success(`成功删除 ${successCount} 个角色`)
    } else {
      ElMessage.warning(`删除完成：成功 ${successCount} 个，失败 ${failCount} 个`)
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

const handleBatchImport = async () => {
  if (selectedImportCharacters.value.length === 0) {
    ElMessage.warning('请选择要导入的角色')
    return
  }
  
  try {
    let successCount = 0
    let failCount = 0
    
    for (const charId of selectedImportCharacters.value) {
      const char = importCharacters.value.find(c => c.id === charId)
      if (char) {
        try {
          const res = await characterAPI.create({
            book_id: bookId,
            name: char.name,
            gender: char.gender,
            personality: char.personality,
            info: char.info,
            folder: char.folder,
            folders: char.folders
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
    }
    
    await fetchCharacters()
    showBatchDialog.value = false
    selectedImportCharacters.value = []
    
    if (failCount === 0) {
      ElMessage.success(`成功导入 ${successCount} 个角色`)
    } else {
      ElMessage.warning(`导入完成：成功 ${successCount} 个，失败 ${failCount} 个`)
    }
  } catch (error) {
    ElMessage.error('批量导入失败')
  }
}

onMounted(async () => {
  await fetchCharacters()
  await fetchModels()
})
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

.recognize-panel-dialog .el-dialog {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  margin: 0;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
}

.recognize-panel-dialog .el-dialog__body {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.recognize-panel-dialog .el-dialog__header {
  padding: 16px 24px;
  border-bottom: 1px solid rgba(130, 210, 120, 0.2);
}

.recognize-panel-dialog .el-dialog__footer {
  padding: 12px 24px;
  border-top: 1px solid rgba(130, 210, 120, 0.2);
}

.recognize-panel-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.recognize-panel-content .recognize-summary {
  text-align: center;
  margin-bottom: 8px;
}

.recognize-panel-content .summary-tag {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.72);
  color: #4f8f1d;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.04em;
  margin-bottom: 8px;
}

.recognize-panel-content .recognize-summary p {
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
  color: #4e6a4a;
}

.recognize-panel-content .recognize-settings {
  display: grid;
  grid-template-columns: minmax(280px, 1.2fr) minmax(220px, 0.8fr);
  gap: 18px;
}

.recognize-panel-content .setting-item {
  min-width: 0;
  padding: 16px 18px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.75);
  border: 1px solid rgba(130, 210, 120, 0.18);
}

.recognize-panel-content .setting-label {
  display: block;
  margin-bottom: 12px;
  font-size: 13px;
  font-weight: 600;
  color: #355430;
}

.recognize-panel-content .setting-slider :deep(.el-slider) {
  width: 100%;
}

.recognize-panel-content .setting-select :deep(.el-select) {
  width: 100%;
}

.recognize-panel-content .recognize-workbench {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(0, 0.85fr);
  gap: 18px;
}

.recognize-panel-content .recognize-card {
  min-width: 0;
  padding: 18px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.78);
  border: 1px solid rgba(130, 210, 120, 0.18);
}

.recognize-panel-content .recognize-card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
}

.recognize-panel-content .recognize-card-head h3 {
  margin: 0 0 6px;
  font-size: 16px;
  color: #21461d;
}

.recognize-panel-content .recognize-card-head p {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: #688063;
}

.recognize-panel-content .recognize-actions-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.recognize-panel-content .recognize-hint {
  font-size: 13px;
  line-height: 1.6;
  color: #688063;
}

@media (max-width: 900px) {
  .recognize-panel-content .recognize-settings,
  .recognize-panel-content .recognize-workbench {
    grid-template-columns: 1fr;
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
</style>
