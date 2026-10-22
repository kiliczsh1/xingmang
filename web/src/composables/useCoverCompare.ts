// 封面前后对比滑块（09）：拖动分割线看新旧封面
// 2026-09-21 从 views/Books.vue 抽出（第 2 期瘦身）；共享状态注入，不由本模块持有

import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { Ref } from 'vue'
import type { Book } from '@/types'

export const useCoverCompare = (coverPreview: Ref<string>, selectedBookForCover: Ref<Book | null>) => {
  // 前后对比（09）
  const compareMode = ref(false)

  const cmpPos = ref(50)

  const cmpDragging = ref(false)

  const cmpRef = ref<HTMLElement | null>(null)

  // ===== 前后对比（09）=====
  const setCompare = (v: boolean) => {
    if (v) {
      if (!coverPreview.value) {
        ElMessage.warning('请先上传或生成封面，再进行对比')
        return
      }
      if (!selectedBookForCover.value?.cover) {
        ElMessage.warning('本书暂无原封面可对比')
        return
      }
    }
    compareMode.value = v
  }

  const setCmpPos = (e: MouseEvent | TouchEvent) => {
    const el = cmpRef.value
    if (!el) return
    const r = el.getBoundingClientRect()
    if (!r.width) return
    const x = 'touches' in e ? e.touches[0].clientX : e.clientX
    cmpPos.value = Math.max(3, Math.min(97, ((x - r.left) / r.width) * 100))
  }

  const cmpStart = (e: MouseEvent | TouchEvent) => {
    cmpDragging.value = true
    setCmpPos(e)
  }

  const onCmpDocMove = (e: MouseEvent) => {
    if (cmpDragging.value) setCmpPos(e)
  }

  const onCmpDocUp = () => {
    cmpDragging.value = false
  }

  const onCmpDocTouch = (e: TouchEvent) => {
    if (cmpDragging.value) {
      e.preventDefault()
      setCmpPos(e)
    }
  }

  return { compareMode, cmpPos, cmpDragging, cmpRef, setCompare, setCmpPos, cmpStart, onCmpDocMove, onCmpDocUp, onCmpDocTouch }
}
