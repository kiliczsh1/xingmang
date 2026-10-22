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

export function useCharacterLibrary(options: {
  bookId?: number
  onClose?: () => void
  onSelectCharacter?: (character: Character) => void
}) {
  const route = useRoute()
  const router = useRouter()

  const bookId = options.bookId ?? parseInt(route.params.bookId as string)

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
        models.value = res.data.filter(m => m.enabled !== 0)
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
    if (options.onClose) {
      options.onClose()
    } else {
      router.push(`/write/${bookId}`)
    }
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

  const insertCharacterToEditor = (character: Character) => {
    if (options.onSelectCharacter) {
      options.onSelectCharacter(character)
    }
    handleClose()
  }

  onMounted(async () => {
    await fetchCharacters()
    await fetchModels()
  })

  return {
    // refs
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

    // computed
    selectedRecognizedCount, dialogWidth, filteredCharacters,

    // methods
    fetchCharacters, selectCharacter, editCharacter, viewCharacter,
    createNewCharacter, saveCharacter, deleteCharacter,
    addFolder, handleClose,
    showMentionedChapters, exportCharacter, aiGenerateCharacter,
    startRecognize, confirmCreateRecognizedCharacters,
    fetchAllBooks, fetchImportCharacters, confirmImport,
    handleBatchOperation, handleSelectAllBatch, handleBatchDelete,
    handleBatchImport, resetRecognizePrompt,
    handleSelectAllRecognized, insertCharacterToEditor,

    // utility functions
    normalizeGender, getGenderText, getEffectiveRecognizePrompt,
  }
}
