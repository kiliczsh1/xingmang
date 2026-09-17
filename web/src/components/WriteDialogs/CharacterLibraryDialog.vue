<template>
  <el-dialog
    :model-value="visible"
    fullscreen
    append-to-body
    destroy-on-close
    class="xm-char-library xm-dialog"
    @update:model-value="emit('update:visible', $event)"
  >
    <template #header>
      <div class="xm-cl-header">
        <span class="xm-cl-title">角色卡</span>
        <span class="xm-cl-link">固定到资料栏</span>
        <span class="xm-cl-link">教程</span>
        <el-button size="small" type="warning" @click="createDraft">新建</el-button>
        <el-button size="small" @click="importDraft">导入</el-button>
      </div>
    </template>

    <div class="xm-cl-body">
      <!-- 左：角色文件夹 -->
      <aside class="xm-cl-folders">
        <div class="xm-cl-side-title">角色文件夹</div>
        <div
          class="xm-cl-folder-item"
          :class="{ 'is-active': activeFolder === '' }"
          @click="activeFolder = ''"
        >全部角色 <span class="xm-cl-folder-count">{{ characters.length }}</span></div>
        <div
          v-for="f in folderList"
          :key="f"
          class="xm-cl-folder-item"
          :class="{ 'is-active': activeFolder === f }"
          @click="activeFolder = f"
        >{{ f }} <span class="xm-cl-folder-count">{{ characters.filter(c => c.folder === f).length }}</span></div>
        <el-button size="small" class="xm-cl-new-folder" @click="createFolder">新建文件夹</el-button>
        <el-input v-model="keyword" size="small" placeholder="搜索名称..." clearable class="xm-cl-search" />
      </aside>

      <!-- 中：角色列表 -->
      <section class="xm-cl-list">
        <div class="xm-cl-list-toolbar">
          <el-button size="small" type="warning" plain @click="emit('notify', '「AI生成角色」功能开发中')">AI生成</el-button>
          <el-button size="small" type="warning" plain @click="emit('notify', '「智能识别」功能开发中')">智能识别</el-button>
          <el-button size="small" @click="emit('notify', '「批量管理」功能开发中')">批量</el-button>
        </div>
        <div class="xm-cl-list-scroll">
          <div v-if="filteredCharacters.length === 0" class="xm-cl-empty">暂无角色</div>
          <div
            v-for="c in filteredCharacters"
            :key="c.id"
            class="xm-cl-card"
            :class="{ 'is-active': current?.id === c.id }"
            @click="selectCharacter(c)"
          >
            <div class="xm-cl-avatar">{{ c.name?.slice(0, 1) || '角' }}</div>
            <div class="xm-cl-card-main">
              <span class="xm-cl-card-name">{{ c.name }}</span>
              <span class="xm-cl-card-sub">{{ genderLabel(c.gender) }}{{ c.folder ? ' · ' + c.folder : '' }}</span>
            </div>
          </div>
        </div>
      </section>

      <!-- 右：角色详情 -->
      <section class="xm-cl-detail">
        <template v-if="current">
          <div class="xm-cl-detail-head">
            <span class="xm-cl-detail-title">角色详情</span>
            <span class="xm-cl-detail-id">ID: {{ current.id || 'char_new' }}</span>
            <el-button size="small" text @click="current = null">关闭详情</el-button>
          </div>
          <div class="xm-cl-detail-form">
            <div class="xm-cl-field">
              <label class="xm-cl-label">所属文件夹</label>
              <el-select v-model="form.folder" placeholder="未归入文件夹" clearable>
                <el-option label="未归入文件夹" value="" />
                <el-option v-for="f in folderList" :key="f" :label="f" :value="f" />
              </el-select>
            </div>
            <div class="xm-cl-field">
              <label class="xm-cl-label"><span class="xm-cl-req">*</span>姓名</label>
              <el-input v-model="form.name" maxlength="50" show-word-limit placeholder="角色名称示例一" />
            </div>
            <div class="xm-cl-field">
              <label class="xm-cl-label"><span class="xm-cl-req">*</span>性别</label>
              <el-select v-model="form.gender">
                <el-option label="男" value="male" />
                <el-option label="女" value="female" />
                <el-option label="未知" value="unknown" />
                <el-option label="无性别" value="none" />
              </el-select>
            </div>
            <div class="xm-cl-field">
              <label class="xm-cl-label"><span class="xm-cl-req">*</span>角色性格</label>
              <p class="xm-cl-hint">前 100 字免费关联，超出部分仅在使用本卡生成时计入关联字数</p>
              <el-input v-model="form.personality" type="textarea" :rows="4" placeholder="请输入角色的性格特点、喜好等..." />
            </div>
            <div class="xm-cl-field">
              <label class="xm-cl-label"><span class="xm-cl-req">*</span>角色设定与背景</label>
              <p class="xm-cl-hint">前 500 字免费关联，超出部分仅在使用本卡生成时计入关联字数</p>
              <el-input v-model="form.info" type="textarea" :rows="8" placeholder="请输入角色的详细经历、背景故事等内容..." />
            </div>
            <div class="xm-cl-field">
              <label class="xm-cl-label">外貌</label>
              <p class="xm-cl-hint">支持上传图片识别角色外貌特征 · 前 300 字免费关联</p>
              <el-input v-model="appearance" type="textarea" :rows="3" placeholder="描述角色的外貌特征..." />
            </div>
          </div>
          <div class="xm-cl-detail-footer">
            <el-button type="warning" @click="saveCharacter">保存角色</el-button>
            <el-button type="danger" plain @click="removeCharacter">删除角色</el-button>
            <el-button @click="exportCharacter">导出角色</el-button>
            <el-button @click="emit('notify', '「提及章节」功能开发中')">提及章节</el-button>
          </div>
        </template>
        <div v-else class="xm-cl-empty xm-cl-empty--detail">从左侧选择角色，或点击「新建」创建角色卡</div>
      </section>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { characterAPI } from '@/api'
import type { Character } from '@/types'

defineOptions({ name: 'CharacterLibraryDialog' })

const props = defineProps<{
  visible: boolean
  bookId: number
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'changed'): void
  (e: 'notify', msg: string): void
}>()

const characters = ref<Character[]>([])
const folderList = ref<string[]>([])
const activeFolder = ref('')
const keyword = ref('')
const current = ref<Character | null>(null)
const form = ref({
  name: '',
  gender: 'unknown',
  personality: '',
  info: '',
  folder: ''
})
const appearance = ref('')
const isNewDraft = ref(false)

const filteredCharacters = computed(() =>
  characters.value
    .filter(c => !activeFolder.value || c.folder === activeFolder.value)
    .filter(c => !keyword.value.trim() || (c.name || '').includes(keyword.value.trim()))
)

const genderLabel = (g: string) => ({ male: '男', female: '女', unknown: '未知', none: '无性别' }[g] || '未知')

watch(() => props.visible, async (val) => {
  if (!val) return
  current.value = null
  keyword.value = ''
  activeFolder.value = ''
  await refresh()
})

const refresh = async () => {
  const res = await characterAPI.getByBook(props.bookId)
  characters.value = res.success && res.data ? res.data : []
  const fr = await characterAPI.getFolders(props.bookId)
  folderList.value = fr.success && fr.data ? fr.data : []
}

const selectCharacter = (c: Character) => {
  current.value = { ...c }
  form.value = {
    name: c.name || '',
    gender: c.gender || 'unknown',
    personality: c.personality || '',
    info: c.info || '',
    folder: c.folder || ''
  }
  appearance.value = ''
  isNewDraft.value = false
}

const createDraft = () => {
  current.value = {
    id: 0,
    book_id: props.bookId,
    name: '',
    gender: 'unknown',
    personality: '',
    info: '',
    folder: '',
    folders: [],
    created_at: '',
    updated_at: ''
  }
  form.value = { name: '', gender: 'unknown', personality: '', info: '', folder: '' }
  isNewDraft.value = true
}

const saveCharacter = async () => {
  if (!current.value) return
  if (!form.value.name.trim()) {
    ElMessage.warning('请输入角色姓名')
    return
  }
  try {
    if (isNewDraft.value) {
      const res = await characterAPI.create({
        book_id: props.bookId,
        name: form.value.name,
        gender: form.value.gender,
        personality: form.value.personality,
        info: form.value.info,
        folder: form.value.folder
      })
      if (res.success && res.data) {
        current.value = { ...res.data }
        isNewDraft.value = false
      }
      ElMessage.success('角色已创建')
    } else {
      await characterAPI.update(current.value.id, {
        name: form.value.name,
        gender: form.value.gender,
        personality: form.value.personality,
        info: form.value.info,
        folder: form.value.folder
      })
      ElMessage.success('角色已保存')
    }
    await refresh()
    emit('changed')
  } catch {
    ElMessage.error('保存失败')
  }
}

const removeCharacter = async () => {
  if (!current.value || isNewDraft.value) {
    current.value = null
    return
  }
  try {
    await ElMessageBox.confirm(`确定删除角色「${current.value.name}」？`, '删除角色', { type: 'warning' })
  } catch {
    return
  }
  const res = await characterAPI.delete(current.value.id)
  if (res.success) {
    ElMessage.success('角色已删除')
    current.value = null
    await refresh()
    emit('changed')
  }
}

const exportCharacter = () => {
  if (!current.value) return
  const blob = new Blob([JSON.stringify(current.value, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `角色卡_${current.value.name || '未命名'}.json`
  a.click()
  URL.revokeObjectURL(url)
}

const importDraft = () => {
  emit('notify', '「导入角色」功能开发中')
}

const createFolder = async () => {
  try {
    const { value } = await ElMessageBox.prompt('请输入文件夹名称', '新建文件夹', {
      confirmButtonText: '创建',
      cancelButtonText: '取消',
      inputValidator: (v: string) => (v.trim() ? true : '名称不能为空')
    })
    if (!folderList.value.includes(value.trim())) folderList.value.push(value.trim())
  } catch {
    // 取消
  }
}
</script>

<style scoped>
.xm-cl-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.xm-cl-title {
  font-size: 17px;
  font-weight: 600;
  color: #1f2329;
}

.xm-cl-link {
  font-size: 12px;
  color: #8a8f99;
  cursor: pointer;
}

.xm-cl-link:hover {
  color: #4f7cf7;
}

.xm-cl-body {
  display: flex;
  gap: 14px;
  height: calc(100vh - 140px);
}

.xm-cl-folders {
  width: 190px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
  border-right: 1px solid #eeeeea;
  padding-right: 12px;
}

.xm-cl-side-title {
  font-size: 13px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 6px;
}

.xm-cl-folder-item {
  font-size: 12.5px;
  color: #5a6270;
  padding: 6px 8px;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
}

.xm-cl-folder-item:hover {
  background: #f3f5f7;
}

.xm-cl-folder-item.is-active {
  background: #f0f4ff;
  color: #4f6ef7;
}

.xm-cl-folder-count {
  color: #a3a8b3;
}

.xm-cl-new-folder {
  margin: 8px 0;
  align-self: flex-start;
}

.xm-cl-search {
  margin-top: auto;
}

.xm-cl-list {
  width: 260px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  border-right: 1px solid #eeeeea;
  padding-right: 12px;
}

.xm-cl-list-toolbar {
  display: flex;
  gap: 6px;
}

.xm-cl-list-toolbar .el-button {
  flex: 1;
  margin: 0;
}

.xm-cl-list-scroll {
  flex: 1;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.xm-cl-empty {
  font-size: 12px;
  color: #a3a8b3;
  text-align: center;
  padding: 40px 0;
}

.xm-cl-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border: 1px solid #eef0f2;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.12s ease;
}

.xm-cl-card:hover {
  border-color: #d8dcdf;
  background: #fafbfc;
}

.xm-cl-card.is-active {
  border-color: #4f6ef7;
  background: #f0f4ff;
}

.xm-cl-avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  background: linear-gradient(135deg, #6366f1, #a855f7);
  color: #fff;
  font-size: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.xm-cl-card-main {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.xm-cl-card-name {
  font-size: 13px;
  font-weight: 500;
  color: #1f2329;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.xm-cl-card-sub {
  font-size: 11px;
  color: #a3a8b3;
}

.xm-cl-detail {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.xm-cl-detail-head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.xm-cl-detail-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f2329;
}

.xm-cl-detail-id {
  font-size: 11px;
  color: #a3a8b3;
  flex: 1;
}

.xm-cl-detail-form {
  flex: 1;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-right: 4px;
}

.xm-cl-field {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.xm-cl-label {
  font-size: 12.5px;
  font-weight: 500;
  color: #374151;
}

.xm-cl-req {
  color: #e54d42;
  margin-right: 2px;
}

.xm-cl-hint {
  margin: 0;
  font-size: 11px;
  color: #a3a8b3;
}

.xm-cl-detail-footer {
  display: flex;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid #eeeeea;
}

.xm-cl-empty--detail {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

:root[data-theme='dark'] .xm-cl-title,
:root[data-theme='dark'] .xm-cl-detail-title,
:root[data-theme='dark'] .xm-cl-card-name,
:root[data-theme='dark'] .xm-cl-label {
  color: #e6edf3;
}

:root[data-theme='dark'] .xm-cl-folders,
:root[data-theme='dark'] .xm-cl-list,
:root[data-theme='dark'] .xm-cl-detail-footer {
  border-color: #3f4a5c;
}

:root[data-theme='dark'] .xm-cl-folder-item:hover,
:root[data-theme='dark'] .xm-cl-card:hover {
  background: rgba(255, 255, 255, 0.06);
}

:root[data-theme='dark'] .xm-cl-card {
  border-color: #3f4a5c;
}

:root[data-theme='dark'] .xm-cl-card.is-active {
  background: rgba(79, 110, 247, 0.15);
}

:root[data-theme='dark'] .xm-cl-hint,
:root[data-theme='dark'] .xm-cl-card-sub,
:root[data-theme='dark'] .xm-cl-folder-count,
:root[data-theme='dark'] .xm-cl-detail-id,
:root[data-theme='dark'] .xm-cl-empty {
  color: #6b7688;
}

:root[data-theme='dark'] .xm-cl-folder-item {
  color: #cbd5e1;
}
</style>
