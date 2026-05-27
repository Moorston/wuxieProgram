<template>
  <view class="detail-page" v-if="plan">
    <view class="plan-header">
      <text class="plan-title">{{ plan.title }}</text>
      <text class="plan-status" :class="'status-' + plan.status">
        {{ statusMap[plan.status] }}
      </text>
    </view>
    <text class="plan-desc">{{ plan.description }}</text>
    <view class="plan-dates">
      <text>{{ plan.start_date?.slice(0, 10) }} ~ {{ plan.end_date?.slice(0, 10) }}</text>
    </view>

    <view class="progress-section">
      <view class="progress-bar">
        <view class="progress-fill" :style="{ width: plan.stats?.completion_rate + '%' }"></view>
      </view>
      <text class="progress-text">完成 {{ plan.stats?.completed || 0 }}/{{ plan.stats?.total_tasks || 0 }} 项 ({{ (plan.stats?.completion_rate || 0).toFixed(0) }}%)</text>
    </view>

    <view class="days-list">
      <view v-for="(day, dayIndex) in plan.days" :key="dayIndex" class="day-card">
        <view class="day-header" @tap="toggleDay(dayIndex)">
          <text class="day-label">第 {{ dayIndex + 1 }} 天</text>
          <text class="day-date">{{ day.date?.slice(0, 10) }}</text>
          <text class="arrow" :class="{ expanded: expandedDays.includes(dayIndex) }">▼</text>
        </view>

        <view v-if="expandedDays.includes(dayIndex)" class="day-tasks">
          <view v-for="(task, taskIndex) in day.tasks" :key="taskIndex" class="task-item">
            <view class="task-info">
              <text class="task-title">{{ task.title }}</text>
              <text class="task-meta">{{ typeMap[task.type] }} · {{ task.reps }} · {{ task.duration }}分钟</text>
              <text v-if="task.note" class="task-note">{{ task.note }}</text>
            </view>
            <view class="task-actions">
              <text v-if="task.status === 0" class="task-status pending" @tap="updateTask(dayIndex, taskIndex, 1)">完成</text>
              <text v-if="task.status === 0" class="task-status skip" @tap="updateTask(dayIndex, taskIndex, 2)">跳过</text>
              <text v-if="task.status === 1" class="task-status done">已完成</text>
              <text v-if="task.status === 2" class="task-status skipped">已跳过</text>
            </view>
          </view>
        </view>
      </view>
    </view>

    <view class="actions">
      <button class="action-btn report" @tap="goReport">查看报告</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getTrainingPlan, updateTaskStatus } from '../../api'

const planId = ref('')
const plan = ref<any>(null)
const expandedDays = ref<number[]>([])

const statusMap: Record<number, string> = {
  0: '草稿',
  1: '进行中',
  2: '已完成',
  3: '已终止',
}

const typeMap: Record<string, string> = {
  basic: '基本功',
  taolu: '套路',
  sanda: '散打',
  qigong: '气功',
}

onLoad((options) => {
  planId.value = options?.id || ''
  loadPlan()
})

async function loadPlan() {
  try {
    plan.value = await getTrainingPlan(planId.value)
  } catch (e) {}
}

function toggleDay(index: number) {
  const i = expandedDays.value.indexOf(index)
  if (i >= 0) {
    expandedDays.value.splice(i, 1)
  } else {
    expandedDays.value.push(index)
  }
}

async function updateTask(dayIndex: number, taskIndex: number, status: number) {
  try {
    await updateTaskStatus(planId.value, dayIndex, taskIndex, status)
    await loadPlan()
    uni.showToast({ title: status === 1 ? '已完成' : '已跳过', icon: 'success' })
  } catch (e) {}
}

function goReport() {
  uni.navigateTo({ url: `/pages/training/report?id=${planId.value}` })
}
</script>

<style scoped>
.detail-page {
  padding: 20rpx;
  padding-bottom: 120rpx;
}
.plan-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
}
.plan-title {
  font-size: 36rpx;
  font-weight: bold;
  flex: 1;
}
.plan-status {
  font-size: 22rpx;
  padding: 4rpx 16rpx;
  border-radius: 20rpx;
}
.status-0 { background: #f0f0f0; color: #999; }
.status-1 { background: #e8f5e9; color: #4caf50; }
.status-2 { background: #e3f2fd; color: #2196f3; }
.status-3 { background: #fce4ec; color: #e91e63; }
.plan-desc {
  font-size: 28rpx;
  color: #666;
  background: #fff;
  border-radius: 0 0 16rpx 16rpx;
  padding: 0 24rpx 24rpx;
  display: block;
}
.plan-dates {
  font-size: 24rpx;
  color: #999;
  margin-top: 12rpx;
}
.progress-section {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-top: 16rpx;
}
.progress-bar {
  height: 12rpx;
  background: #f0f0f0;
  border-radius: 6rpx;
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  background: #1cbbb4;
  border-radius: 6rpx;
  transition: width 0.3s;
}
.progress-text {
  font-size: 24rpx;
  color: #666;
  margin-top: 12rpx;
  display: block;
}
.days-list {
  margin-top: 16rpx;
}
.day-card {
  background: #fff;
  border-radius: 16rpx;
  margin-bottom: 12rpx;
  overflow: hidden;
}
.day-header {
  display: flex;
  align-items: center;
  padding: 24rpx;
}
.day-label {
  font-size: 30rpx;
  font-weight: bold;
  flex: 1;
}
.day-date {
  font-size: 24rpx;
  color: #999;
  margin-right: 12rpx;
}
.arrow {
  font-size: 24rpx;
  color: #ccc;
  transition: transform 0.3s;
}
.arrow.expanded {
  transform: rotate(180deg);
}
.day-tasks {
  padding: 0 24rpx 24rpx;
  border-top: 1rpx solid #f0f0f0;
}
.task-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16rpx 0;
  border-bottom: 1rpx solid #f8f8f8;
}
.task-info {
  flex: 1;
}
.task-title {
  font-size: 28rpx;
  display: block;
}
.task-meta {
  font-size: 22rpx;
  color: #999;
  margin-top: 4rpx;
  display: block;
}
.task-note {
  font-size: 22rpx;
  color: #666;
  margin-top: 4rpx;
  display: block;
}
.task-actions {
  display: flex;
  gap: 12rpx;
}
.task-status {
  font-size: 22rpx;
  padding: 6rpx 16rpx;
  border-radius: 20rpx;
}
.task-status.pending {
  background: #e3f2fd;
  color: #2196f3;
}
.task-status.skip {
  background: #f0f0f0;
  color: #999;
}
.task-status.done {
  background: #e8f5e9;
  color: #4caf50;
}
.task-status.skipped {
  background: #f0f0f0;
  color: #999;
}
.actions {
  margin-top: 20rpx;
}
.action-btn {
  background: #1cbbb4;
  color: #fff;
  border-radius: 40rpx;
  height: 80rpx;
  line-height: 80rpx;
  font-size: 30rpx;
}
</style>
