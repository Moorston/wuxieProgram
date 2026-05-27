<template>
  <view class="today-page">
    <view v-if="taskGroups.length > 0">
      <view v-for="group in taskGroups" :key="group.plan_id" class="plan-group">
        <view class="plan-header">
          <text class="plan-title">{{ group.plan_title }}</text>
          <text class="day-label">第 {{ group.day?.day }} 天</text>
        </view>

        <view v-for="(task, taskIndex) in group.day?.tasks" :key="taskIndex" class="task-card">
          <view class="task-info">
            <text class="task-title">{{ task.title }}</text>
            <text class="task-meta">{{ typeMap[task.type] }} · {{ task.reps }} · {{ task.duration }}分钟</text>
            <text v-if="task.note" class="task-note">{{ task.note }}</text>
          </view>
          <view class="task-actions">
            <text v-if="task.status === 0" class="action-btn complete" @tap="completeTask(group.plan_id, group.day.day - 1, taskIndex)">完成</text>
            <text v-if="task.status === 0" class="action-btn skip" @tap="skipTask(group.plan_id, group.day.day - 1, taskIndex)">跳过</text>
            <text v-if="task.status === 1" class="action-btn done">已完成</text>
            <text v-if="task.status === 2" class="action-btn skipped">已跳过</text>
          </view>
        </view>
      </view>
    </view>

    <view v-else-if="!loading" class="empty">
      <text class="empty-icon">📋</text>
      <text class="empty-text">今天没有训练任务</text>
      <text class="empty-hint">去创建一个训练计划吧</text>
    </view>

    <view v-if="loading" class="loading-tip">
      <text>加载中...</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onPullDownRefresh } from '@dcloudio/uni-app'
import { getTodayTasks, updateTaskStatus } from '../../api'

const taskGroups = ref<any[]>([])
const loading = ref(false)

const typeMap: Record<string, string> = {
  basic: '基本功',
  taolu: '套路',
  sanda: '散打',
  qigong: '气功',
}

onMounted(() => {
  loadData()
})

onPullDownRefresh(async () => {
  await loadData()
  uni.stopPullDownRefresh()
})

async function loadData() {
  loading.value = true
  try {
    const res: any = await getTodayTasks()
    taskGroups.value = res || []
  } catch (e) {} finally {
    loading.value = false
  }
}

async function completeTask(planId: string, dayIndex: number, taskIndex: number) {
  try {
    await updateTaskStatus(planId, dayIndex, taskIndex, 1)
    await loadData()
    uni.showToast({ title: '已完成', icon: 'success' })
  } catch (e) {}
}

async function skipTask(planId: string, dayIndex: number, taskIndex: number) {
  try {
    await updateTaskStatus(planId, dayIndex, taskIndex, 2)
    await loadData()
    uni.showToast({ title: '已跳过', icon: 'none' })
  } catch (e) {}
}
</script>

<style scoped>
.today-page {
  padding: 20rpx;
}
.plan-group {
  background: #fff;
  border-radius: 16rpx;
  margin-bottom: 20rpx;
  overflow: hidden;
}
.plan-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx;
  background: linear-gradient(135deg, #1cbbb4, #0081ff);
  color: #fff;
}
.plan-title {
  font-size: 32rpx;
  font-weight: bold;
}
.day-label {
  font-size: 24rpx;
  opacity: 0.9;
}
.task-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx;
  border-bottom: 1rpx solid #f0f0f0;
}
.task-card:last-child {
  border-bottom: none;
}
.task-info {
  flex: 1;
}
.task-title {
  font-size: 30rpx;
  display: block;
}
.task-meta {
  font-size: 24rpx;
  color: #999;
  margin-top: 8rpx;
  display: block;
}
.task-note {
  font-size: 24rpx;
  color: #666;
  margin-top: 4rpx;
  display: block;
}
.task-actions {
  display: flex;
  gap: 12rpx;
}
.action-btn {
  font-size: 24rpx;
  padding: 8rpx 20rpx;
  border-radius: 24rpx;
}
.action-btn.complete {
  background: #e8f5e9;
  color: #4caf50;
}
.action-btn.skip {
  background: #f0f0f0;
  color: #999;
}
.action-btn.done {
  background: #e8f5e9;
  color: #4caf50;
}
.action-btn.skipped {
  background: #f0f0f0;
  color: #999;
}
.empty {
  text-align: center;
  padding: 100rpx;
}
.empty-icon {
  font-size: 80rpx;
  display: block;
  margin-bottom: 20rpx;
}
.empty-text {
  font-size: 32rpx;
  color: #333;
  display: block;
}
.empty-hint {
  font-size: 26rpx;
  color: #999;
  margin-top: 12rpx;
  display: block;
}
.loading-tip {
  text-align: center;
  padding: 30rpx;
  color: #999;
  font-size: 24rpx;
}
</style>
