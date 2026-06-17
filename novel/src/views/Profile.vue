<template>
  <div class="profile-container">
    <h2 class="page-title">个人中心</h2>
    
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon prompts">
            <el-icon><ChatDotSquare /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ overview?.promptCount || 0 }}</div>
            <div class="stat-label">提示词数量</div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon generators">
            <el-icon><Magic /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ overview?.generatorCount || 0 }}</div>
            <div class="stat-label">生成器应用</div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon books">
            <el-icon><Reading /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ overview?.bookCount || 0 }}</div>
            <div class="stat-label">书本数量</div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon usage">
            <el-icon><TrendCharts /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ overview?.totalUsageCount || 0 }}</div>
            <div class="stat-label">总调用次数</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="16">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <div class="flex items-center">
                <el-button
                  type="text"
                  @click="toggleDailyStats"
                  class="toggle-btn"
                >
                  <el-icon :class="{ 'rotate-90': showDailyStats }"><ArrowRight /></el-icon>
                </el-button>
                <span>每日模型使用统计</span>
              </div>
              <el-date-picker
                v-model="dateRange"
                type="daterange"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                format="YYYY-MM-DD"
                value-format="YYYY-MM-DD"
                @change="loadDailyStats"
                style="width: 280px"
                :disabled="!showDailyStats"
              />
            </div>
          </template>
          <div v-if="showDailyStats" class="daily-stats" v-loading="dailyLoading">
            <el-table :data="dailyStats" style="width: 100%" max-height="400">
              <el-table-column prop="date" label="日期" width="120" />
              <el-table-column prop="model_name" label="模型" min-width="150" />
              <el-table-column prop="provider_name" label="服务商" width="120" />
              <el-table-column prop="usage_count" label="调用次数" width="100" align="center">
                <template #default="{ row }">
                  <el-tag type="primary">{{ row.usage_count }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="total_tokens" label="Token用量" width="120" align="center">
                <template #default="{ row }">
                  <el-tag type="success">{{ formatTokens(row.total_tokens) }}</el-tag>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="dailyStats.length === 0 && !dailyLoading" description="暂无使用记录" />
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="8">
        <el-card class="chart-card">
          <template #header>
            <span>模型使用排行</span>
          </template>
          <div class="model-ranking" v-loading="modelLoading">
            <div 
              v-for="(item, index) in modelStats" 
              :key="item.model_id" 
              class="ranking-item"
            >
              <div class="ranking-index" :class="getRankClass(index)">
                {{ index + 1 }}
              </div>
              <div class="ranking-info">
                <div class="ranking-name">{{ item.model_name }}</div>
                <div class="ranking-provider">{{ item.provider_name }}</div>
              </div>
              <div class="ranking-count">
                <span class="count">{{ item.total_usage }}</span>
                <span class="unit">次</span>
              </div>
            </div>
            <el-empty v-if="modelStats.length === 0 && !modelLoading" description="暂无数据" :image-size="80" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card class="chart-card">
          <template #header>
            <div class="flex items-center">
              <el-button
                type="text"
                @click="toggleMonthlyStats"
                class="toggle-btn"
              >
                <el-icon :class="{ 'rotate-90': showMonthlyStats }"><ArrowRight /></el-icon>
              </el-button>
              <span>月度使用趋势</span>
            </div>
          </template>
          <div v-if="showMonthlyStats" class="monthly-stats" v-loading="monthlyLoading">
            <el-table :data="monthlyStats" style="width: 100%">
              <el-table-column prop="month" label="月份" width="120">
                <template #default="{ row }">
                  {{ row.month }}月
                </template>
              </el-table-column>
              <el-table-column prop="total_usage" label="调用次数" width="150" align="center">
                <template #default="{ row }">
                  <el-tag type="primary" size="large">{{ row.total_usage }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="total_tokens" label="Token用量" width="150" align="center">
                <template #default="{ row }">
                  <el-tag type="success" size="large">{{ formatTokens(row.total_tokens) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="model_count" label="使用模型数" width="150" align="center">
                <template #default="{ row }">
                  <el-tag type="warning" size="large">{{ row.model_count }} 种</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="使用占比" min-width="200">
                <template #default="{ row }">
                  <el-progress 
                    :percentage="getPercentage(row.total_usage)" 
                    :stroke-width="20"
                    :color="getProgressColor(row.total_usage)"
                  />
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="monthlyStats.length === 0 && !monthlyLoading" description="暂无月度数据" />
          </div>
        </el-card>
      </el-col>
    </el-row>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { statsAPI } from '@/api'
import type { UsageOverview, DailyUsage, ModelStats, MonthlyStats } from '@/types'

const overview = ref<UsageOverview | null>(null)
const dailyStats = ref<DailyUsage[]>([])
const modelStats = ref<ModelStats[]>([])
const monthlyStats = ref<MonthlyStats[]>([])

const dailyLoading = ref(false)
const modelLoading = ref(false)
const monthlyLoading = ref(false)

const showDailyStats = ref(false)
const showMonthlyStats = ref(false)

const dateRange = ref<[string, string] | null>(null)

const maxMonthlyUsage = computed(() => {
  if (monthlyStats.value.length === 0) return 1
  return Math.max(...monthlyStats.value.map(m => m.total_usage))
})

const loadOverview = async () => {
  try {
    const res = await statsAPI.getOverview()
    if (res.success && res.data) {
      overview.value = res.data
    }
  } catch (error) {
    console.error('加载概览失败:', error)
  }
}

const loadDailyStats = async () => {
  dailyLoading.value = true
  try {
    let startDate: string | undefined
    let endDate: string | undefined
    
    if (dateRange.value) {
      startDate = dateRange.value[0]
      endDate = dateRange.value[1]
    }
    
    const res = await statsAPI.getDaily(startDate, endDate)
    if (res.success && res.data) {
      dailyStats.value = res.data
    }
  } catch (error) {
    console.error('加载每日统计失败:', error)
  } finally {
    dailyLoading.value = false
  }
}

const loadModelStats = async () => {
  modelLoading.value = true
  try {
    const res = await statsAPI.getModelStats()
    if (res.success && res.data) {
      modelStats.value = res.data
    }
  } catch (error) {
    console.error('加载模型统计失败:', error)
  } finally {
    modelLoading.value = false
  }
}

const loadMonthlyStats = async () => {
  monthlyLoading.value = true
  try {
    const res = await statsAPI.getMonthly()
    if (res.success && res.data) {
      monthlyStats.value = res.data
    }
  } catch (error) {
    console.error('加载月度统计失败:', error)
  } finally {
    monthlyLoading.value = false
  }
}

const formatTokens = (tokens: number) => {
  if (tokens >= 1000000) {
    return (tokens / 1000000).toFixed(2) + 'M'
  } else if (tokens >= 1000) {
    return (tokens / 1000).toFixed(1) + 'K'
  }
  return tokens.toString()
}

const getRankClass = (index: number) => {
  if (index === 0) return 'first'
  if (index === 1) return 'second'
  if (index === 2) return 'third'
  return ''
}

const getPercentage = (usage: number) => {
  if (maxMonthlyUsage.value === 0) return 0
  return Math.round((usage / maxMonthlyUsage.value) * 100)
}

const getProgressColor = (usage: number) => {
  const percentage = getPercentage(usage)
  if (percentage >= 80) return '#67c23a'
  if (percentage >= 50) return '#409eff'
  if (percentage >= 30) return '#e6a23c'
  return '#909399'
}

const toggleDailyStats = async () => {
  showDailyStats.value = !showDailyStats.value
  if (showDailyStats.value) {
    await loadDailyStats()
  }
}

const toggleMonthlyStats = async () => {
  showMonthlyStats.value = !showMonthlyStats.value
  if (showMonthlyStats.value) {
    await loadMonthlyStats()
  }
}

onMounted(() => {
  loadOverview()
  loadModelStats()
})
</script>

<style scoped>
.profile-container {
  padding: 20px;
}

.page-title {
  margin-bottom: 20px;
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.stat-card {
  display: flex;
  align-items: center;
  padding: 20px;
  border-radius: 12px;
  transition: transform 0.3s;
}

.stat-card:hover {
  transform: translateY(-5px);
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
}

.stat-icon .el-icon {
  font-size: 28px;
  color: #fff;
}

.stat-icon.prompts {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.generators {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.books {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-icon.usage {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
  line-height: 1.2;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

.chart-card {
  border-radius: 12px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.flex {
  display: flex;
}

.items-center {
  align-items: center;
}

.toggle-btn {
  margin-right: 8px;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.rotate-90 {
  transform: rotate(90deg);
  transition: transform 0.3s;
}

.toggle-btn .el-icon {
  font-size: 16px;
  transition: transform 0.3s;
}

.model-ranking {
  max-height: 400px;
  overflow-y: auto;
}

.ranking-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border-bottom: 1px solid #f0f0f0;
  transition: background-color 0.3s;
}

.ranking-item:hover {
  background-color: #f5f7fa;
}

.ranking-item:last-child {
  border-bottom: none;
}

.ranking-index {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #e4e7ed;
  color: #909399;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
  margin-right: 12px;
}

.ranking-index.first {
  background: linear-gradient(135deg, #ffd700 0%, #ffec8b 100%);
  color: #fff;
}

.ranking-index.second {
  background: linear-gradient(135deg, #c0c0c0 0%, #e8e8e8 100%);
  color: #fff;
}

.ranking-index.third {
  background: linear-gradient(135deg, #cd7f32 0%, #daa06d 100%);
  color: #fff;
}

.ranking-info {
  flex: 1;
}

.ranking-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.ranking-provider {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

.ranking-count {
  text-align: right;
}

.ranking-count .count {
  font-size: 18px;
  font-weight: 600;
  color: #409eff;
}

.ranking-count .unit {
  font-size: 12px;
  color: #909399;
  margin-left: 2px;
}

/* ========== 暗色主题适配 ========== */
:root[data-theme='dark'] .page-title {
  color: #e5e7eb;
}

:root[data-theme='dark'] .stat-card {
  background: #1e293b;
  border-color: #334155;
}

:root[data-theme='dark'] .stat-value {
  color: #e5e7eb;
}

:root[data-theme='dark'] .stat-label {
  color: #9ca3af;
}

:root[data-theme='dark'] .chart-card {
  background: #1e293b;
  border-color: #334155;
}

:root[data-theme='dark'] .ranking-item {
  border-bottom-color: #334155;
}

:root[data-theme='dark'] .ranking-item:hover {
  background-color: #334155;
}

:root[data-theme='dark'] .ranking-index {
  background: #475569;
  color: #9ca3af;
}

:root[data-theme='dark'] .ranking-name {
  color: #e5e7eb;
}

:root[data-theme='dark'] .ranking-provider {
  color: #9ca3af;
}

:root[data-theme='dark'] .ranking-count .unit {
  color: #9ca3af;
}

:root[data-theme='dark'] :deep(.el-card__header) {
  border-bottom-color: #334155;
  background: #1e293b;
}

:root[data-theme='dark'] :deep(.el-card__body) {
  background: #1e293b;
}

:root[data-theme='dark'] :deep(.el-table) {
  background: #1e293b;
  color: #e5e7eb;
}

:root[data-theme='dark'] :deep(.el-table th) {
  background: #334155;
  color: #e5e7eb;
  border-color: #475569;
}

:root[data-theme='dark'] :deep(.el-table td) {
  border-color: #475569;
}

:root[data-theme='dark'] :deep(.el-table--striped .el-table__body tr.el-table__row--striped td) {
  background: #253347;
}

:root[data-theme='dark'] :deep(.el-table__body tr:hover > td) {
  background: #334155;
}

:root[data-theme='dark'] :deep(.el-empty) {
  background: transparent;
}

:root[data-theme='dark'] :deep(.el-date-picker) {
  background: #1e293b;
}

:root[data-theme='dark'] :deep(.el-progress__text) {
  color: #e5e7eb;
}
</style>
