<template>
  <el-dialog
    :model-value="modelValue"
    title="编辑信息"
    width="560px"
    top="8vh"
    destroy-on-close
    @update:model-value="emit('update:modelValue', $event)"
    @open="initForm"
  >
    <el-form label-width="80px" class="meta-form">
      <el-form-item label="标题" required>
        <el-input
          v-model="form.title"
          maxlength="120"
          show-word-limit
          placeholder="请输入经验卡片标题"
        />
      </el-form-item>

      <el-form-item label="简介">
        <el-input
          v-model="form.summary"
          type="textarea"
          :rows="3"
          maxlength="200"
          show-word-limit
          placeholder="建议一句话概括核心经验，列表页会优先展示这里的内容"
        />
      </el-form-item>

      <el-form-item label="作者">
        <el-input
          v-model="form.author_name"
          maxlength="50"
          placeholder="请输入作者名称"
        />
      </el-form-item>

      <el-form-item label="版本号">
        <el-input
          v-model="form.version"
          maxlength="30"
          placeholder="如 v1.0.0"
        />
      </el-form-item>

      <el-form-item label="封面图">
        <div class="cover-field">
          <div v-if="form.cover_url" class="cover-preview">
            <img :src="form.cover_url" alt="封面预览" />
            <el-button type="danger" plain size="small" @click="form.cover_url = ''">
              移除
            </el-button>
          </div>
          <el-input
            v-model="form.cover_url"
            placeholder="https://example.com/cover.jpg"
            clearable
          />
        </div>
      </el-form-item>

      <el-form-item v-if="share?.pdf_file_url" label="PDF 附件">
        <div class="pdf-info">
          <el-icon><Document /></el-icon>
          <span>{{ share.pdf_file_name || 'PDF附件' }}</span>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Document } from '@element-plus/icons-vue'
import { experienceShareAPI } from '@/api'
import type { ExperienceShare } from '@/types'

const props = defineProps<{
  modelValue: boolean
  share: ExperienceShare | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'saved', data: ExperienceShare): void
}>()

const saving = ref(false)
const form = ref({
  title: '',
  summary: '',
  author_name: '',
  version: '',
  cover_url: ''
})

const initForm = () => {
  const s = props.share
  form.value = {
    title: s?.title || '',
    summary: s?.summary || '',
    author_name: s?.author_name || '',
    version: s?.version || '',
    cover_url: s?.cover_url || ''
  }
}

const handleSave = async () => {
  if (!props.share) return
  if (!form.value.title.trim()) {
    ElMessage.warning('标题不能为空')
    return
  }

  saving.value = true
  try {
    const s = props.share
    // 后端 update 是全量覆盖，必须带上所有字段，否则未传的字段会被清空
    const payload: any = {
      title: form.value.title.trim(),
      summary: form.value.summary.trim(),
      content: s.content || '',
      content_render_mode: 'html',
      cover_url: form.value.cover_url.trim(),
      create_type: s.create_type,
      author_id: s.author_id ?? null,
      author_name: form.value.author_name.trim(),
      version: form.value.version.trim(),
      status: s.status || 'published',
      pdf_file_url: s.pdf_file_url || null,
      pdf_file_name: s.pdf_file_name || null,
      pdf_file_size: s.pdf_file_size || 0,
      pdf_parse_status: s.pdf_parse_status || null,
      pdf_parse_result: s.pdf_parse_result || null,
      source_file_name: s.source_file_name || null
    }

    const res = await experienceShareAPI.update(s.id, payload)
    if (res.success && res.data) {
      ElMessage.success('已保存')
      emit('saved', res.data)
      emit('update:modelValue', false)
    }
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.meta-form :deep(.el-form-item__label) {
  font-size: 14px;
  color: #1f2225;
}

.cover-field {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.cover-preview {
  display: flex;
  align-items: flex-end;
  gap: 10px;
}

.cover-preview img {
  width: 100%;
  max-width: 320px;
  max-height: 160px;
  object-fit: cover;
  border-radius: 6px;
  border: 1px solid rgb(224, 224, 230);
}

.pdf-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #475569;
  font-size: 13px;
  min-width: 0;
}

.pdf-info span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
