<template>
  <view class="template-detail-page" v-if="template">
    <view class="template-header">
      <text class="template-name">{{ template.name }}</text>
      <view class="template-tags">
        <text class="tag">{{ template.style }}</text>
        <text class="tag">{{ template.category }}</text>
        <text class="tag">{{ template.duration_days }}天</text>
      </view>
    </view>
    <text class="template-desc">{{ template.description }}</text>
    <text class="template-meta">作者: {{ template.author }} · {{ template.usage_count }}人使用</text>

    <view class="days-preview">
      <text class="section-title">训练安排预览</text>
      <view v-for="(day, dayIndex) in template.days" :key="dayIndex" class="day-card">
        <text class="day-label">第 {{ dayIndex + 1 }} 天</text>
        <view v-for="(task, taskIndex) in day.tasks" :key="taskIndex" class="task-item">
          <text class="task-title">{{ task.title }}</text>
          <text class="task-meta">{{ typeMap[task.type] }} · {{ task.reps }} · {{ task.duration }}分钟</text>
        </view>
      </view>
    </view>

    <view class="apply-section">
      <view class="form-item">
        <text class="label">选择开始日期</text>
        <picker mode="date" :value="startDate" @change="onDateChange">
          <text class="picker-text">{{ startDate || '选择日期' }}</text>
        </picker>
      </view>
      <button class="apply-btn" :disabled="!startDate || applying" @tap="onApply">
        {{ applying ? '应用中...' : '一键应用创建计划' }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getTemplate, applyTemplate } from '../../api'

const templateId = ref('')
const template = ref<any>(null)
const startDate = ref('')
const applying = ref(false)

const typeMap: Record<string, string> = {
  basic: '基本功',
  taolu: '套路',
  sanda: '散打',
  qigong: '气功',
}

onLoad((options) => {
  templateId.value = options?.id || ''
  loadTemplate()
})

async function loadTemplate() {
  try {
    template.value = await getTemplate(templateId.value)
  } catch (e) {}
}

function onDateChange(e: any) {
  startDate.value = e.detail.value
}

async function onApply() {
  if (!startDate.value) return
  applying.value = true
  try {
    const plan = await applyTemplate(templateId.value, startDate.value)
    uni.showToast({ title: '创建成功', icon: 'success' })
    setTimeout(() => {
      uni.redirectTo({ url: `/pages/training/detail?id=${(plan as any).id}` })
    }, 1500)
  } catch (e) {
    uni.showToast({ title: '创建失败', icon: 'none' })
  } finally {
    applying.value = false
  }
}
</script>

<style scoped>
.template-detail-page {
  padding: 20rpx;
  padding-bottom: 40rpx;
}
.template-header {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
}
.template-name {
  font-size: 36rpx;
  font-weight: bold;
  display: block;
}
.template-tags {
  display: flex;
  gap: 12rpx;
  margin-top: 12rpx;
}
.tag {
  font-size: 22rpx;
  background: #f0f0f0;
  padding: 4rpx 16rpx;
  border-radius: 20rpx;
  color: #666;
}
.template-desc {
  font-size: 28rpx;
  color: #666;
  background: #fff;
  border-radius: 0 0 16rpx 16rpx;
  padding: 0 24rpx 24rpx;
  display: block;
}
.template-meta {
  font-size: 24rpx;
  color: #999;
  margin-top: 12rpx;
}
.days-preview {
  margin-top: 24rpx;
}
.section-title {
  font-size: 30rpx;
  font-weight: bold;
  margin-bottom: 16rpx;
  display: block;
}
.day-card {
  background: #fff;
  border-radius: 12rpx;
  padding: 20rpx;
  margin-bottom: 12rpx;
}
.day-label {
  font-size: 28rpx;
  font-weight: bold;
  margin-bottom: 12rpx;
  display: block;
}
.task-item {
  padding: 8rpx 0;
  border-bottom: 1rpx solid #f8f8f8;
}
.task-title {
  font-size: 26rpx;
  display: block;
}
.task-meta {
  font-size: 22rpx;
  color: #999;
  margin-top: 4rpx;
  display: block;
}
.apply-section {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-top: 24rpx;
}
.form-item {
  margin-bottom: 20rpx;
}
.label {
  font-size: 28rpx;
  font-weight: bold;
  margin-bottom: 12rpx;
  display: block;
}
.picker-text {
  border: 1rpx solid #e0e0e0;
  border-radius: 12rpx;
  padding: 16rpx;
  font-size: 28rpx;
  display: block;
}
.apply-btn {
  background: #1cbbb4;
  color: #fff;
  border-radius: 40rpx;
  height: 80rpx;
  line-height: 80rpx;
  font-size: 30rpx;
}
.apply-btn[disabled] {
  opacity: 0.5;
}
</style>
