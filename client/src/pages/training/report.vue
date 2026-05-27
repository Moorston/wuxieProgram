<template>
  <view class="report-page" v-if="report">
    <view class="report-header">
      <text class="report-title">{{ report.plan?.title }}</text>
      <text class="report-period">{{ report.plan?.start_date?.slice(0, 10) }} ~ {{ report.plan?.end_date?.slice(0, 10) }}</text>
    </view>

    <view class="stats-grid">
      <view class="stat-card">
        <text class="stat-value">{{ report.completed }}</text>
        <text class="stat-label">已完成</text>
      </view>
      <view class="stat-card">
        <text class="stat-value">{{ report.skipped }}</text>
        <text class="stat-label">已跳过</text>
      </view>
      <view class="stat-card">
        <text class="stat-value">{{ report.total_tasks }}</text>
        <text class="stat-label">总任务</text>
      </view>
      <view class="stat-card">
        <text class="stat-value">{{ (report.completion_rate || 0).toFixed(0) }}%</text>
        <text class="stat-label">完成率</text>
      </view>
    </view>

    <view class="progress-section">
      <text class="section-title">完成进度</text>
      <view class="progress-bar">
        <view class="progress-fill" :style="{ width: report.completion_rate + '%' }"></view>
      </view>
      <text class="progress-detail">已进行 {{ report.days_passed }} 天 / 共 {{ report.total_days }} 天</text>
    </view>

    <view class="type-section">
      <text class="section-title">训练类型分布</text>
      <view v-for="(count, type) in report.type_stats" :key="type" class="type-item">
        <view class="type-info">
          <text class="type-name">{{ typeMap[type as string] || type }}</text>
          <text class="type-count">{{ report.type_completed[type as string] || 0 }}/{{ count }}</text>
        </view>
        <view class="type-bar">
          <view class="type-fill" :style="{ width: (count > 0 ? (report.type_completed[type as string] || 0) / count * 100 : 0) + '%' }"></view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getTrainingReport } from '../../api'

const reportId = ref('')
const report = ref<any>(null)

const typeMap: Record<string, string> = {
  basic: '基本功',
  taolu: '套路',
  sanda: '散打',
  qigong: '气功',
}

onLoad((options) => {
  reportId.value = options?.id || ''
  loadReport()
})

async function loadReport() {
  try {
    report.value = await getTrainingReport(reportId.value)
  } catch (e) {}
}
</script>

<style scoped>
.report-page {
  padding: 20rpx;
}
.report-header {
  background: linear-gradient(135deg, #1cbbb4, #0081ff);
  border-radius: 16rpx;
  padding: 40rpx;
  color: #fff;
  text-align: center;
  margin-bottom: 20rpx;
}
.report-title {
  font-size: 36rpx;
  font-weight: bold;
  display: block;
}
.report-period {
  font-size: 26rpx;
  opacity: 0.9;
  margin-top: 8rpx;
  display: block;
}
.stats-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16rpx;
  margin-bottom: 20rpx;
}
.stat-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  text-align: center;
}
.stat-value {
  font-size: 48rpx;
  font-weight: bold;
  color: #1cbbb4;
  display: block;
}
.stat-label {
  font-size: 24rpx;
  color: #999;
  margin-top: 8rpx;
  display: block;
}
.progress-section {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}
.section-title {
  font-size: 30rpx;
  font-weight: bold;
  margin-bottom: 16rpx;
  display: block;
}
.progress-bar {
  height: 16rpx;
  background: #f0f0f0;
  border-radius: 8rpx;
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #1cbbb4, #0081ff);
  border-radius: 8rpx;
  transition: width 0.3s;
}
.progress-detail {
  font-size: 24rpx;
  color: #999;
  margin-top: 12rpx;
  display: block;
}
.type-section {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
}
.type-item {
  margin-bottom: 16rpx;
}
.type-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8rpx;
}
.type-name {
  font-size: 26rpx;
}
.type-count {
  font-size: 24rpx;
  color: #999;
}
.type-bar {
  height: 8rpx;
  background: #f0f0f0;
  border-radius: 4rpx;
  overflow: hidden;
}
.type-fill {
  height: 100%;
  background: #1cbbb4;
  border-radius: 4rpx;
  transition: width 0.3s;
}
</style>
