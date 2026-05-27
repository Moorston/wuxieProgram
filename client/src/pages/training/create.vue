<template>
  <view class="create-page">
    <view class="form-section">
      <view class="form-item">
        <text class="label">计划名称</text>
        <input v-model="form.title" placeholder="如：初级长拳28天计划" class="input" />
      </view>
      <view class="form-item">
        <text class="label">计划描述</text>
        <textarea v-model="form.description" placeholder="简要描述训练目标..." class="textarea" maxlength="200" />
      </view>
      <view class="form-row">
        <view class="form-item half">
          <text class="label">开始日期</text>
          <picker mode="date" :value="form.startDate" @change="onStartDateChange">
            <text class="picker-text">{{ form.startDate || '选择日期' }}</text>
          </picker>
        </view>
        <view class="form-item half">
          <text class="label">结束日期</text>
          <picker mode="date" :value="form.endDate" @change="onEndDateChange">
            <text class="picker-text">{{ form.endDate || '选择日期' }}</text>
          </picker>
        </view>
      </view>
    </view>

    <view class="days-section">
      <view class="section-header">
        <text class="section-title">每日训练安排</text>
        <text class="add-btn" @tap="addDay">+ 添加一天</text>
      </view>

      <view v-for="(day, dayIndex) in form.days" :key="dayIndex" class="day-card">
        <view class="day-header">
          <text class="day-label">第 {{ dayIndex + 1 }} 天</text>
          <text class="remove-btn" @tap="removeDay(dayIndex)">删除</text>
        </view>

        <view v-for="(task, taskIndex) in day.tasks" :key="taskIndex" class="task-item">
          <input v-model="task.title" placeholder="训练项目名称" class="task-input" />
          <view class="task-row">
            <picker :value="taskTypeIndex(task.type)" :range="taskTypeOptions" @change="(e: any) => onTaskTypeChange(dayIndex, taskIndex, e.detail.value)">
              <text class="task-picker">{{ task.typeLabel }}</text>
            </picker>
            <input v-model="task.reps" placeholder="组数次数" class="task-input-sm" />
            <input v-model.number="task.duration" type="number" placeholder="分钟" class="task-input-num" />
          </view>
          <input v-model="task.note" placeholder="备注(可选)" class="task-input" />
          <text class="remove-task" @tap="removeTask(dayIndex, taskIndex)">删除项目</text>
        </view>

        <text class="add-task-btn" @tap="addTask(dayIndex)">+ 添加训练项目</text>
      </view>
    </view>

    <button class="submit-btn" :disabled="!canSubmit || submitting" @tap="onSubmit">
      {{ submitting ? '创建中...' : '创建计划' }}
    </button>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { createTrainingPlan } from '../../api'

const taskTypes = [
  { value: 'basic', label: '基本功' },
  { value: 'taolu', label: '套路' },
  { value: 'sanda', label: '散打' },
  { value: 'qigong', label: '气功' },
]
const taskTypeOptions = taskTypes.map(t => t.label)

const form = ref({
  title: '',
  description: '',
  startDate: '',
  endDate: '',
  days: [] as any[],
})

const submitting = ref(false)

const canSubmit = computed(() => form.value.title && form.value.startDate && form.value.endDate && form.value.days.length > 0)

function taskTypeIndex(type: string) {
  return taskTypes.findIndex(t => t.value === type)
}

function onStartDateChange(e: any) {
  form.value.startDate = e.detail.value
}

function onEndDateChange(e: any) {
  form.value.endDate = e.detail.value
}

function addDay() {
  form.value.days.push({
    day: form.value.days.length + 1,
    tasks: [{ title: '', type: 'basic', typeLabel: '基本功', duration: 0, reps: '', note: '', status: 0 }],
  })
}

function removeDay(index: number) {
  form.value.days.splice(index, 1)
}

function addTask(dayIndex: number) {
  form.value.days[dayIndex].tasks.push({
    title: '',
    type: 'basic',
    typeLabel: '基本功',
    duration: 0,
    reps: '',
    note: '',
    status: 0,
  })
}

function removeTask(dayIndex: number, taskIndex: number) {
  form.value.days[dayIndex].tasks.splice(taskIndex, 1)
}

function onTaskTypeChange(dayIndex: number, taskIndex: number, typeIndex: number) {
  const t = taskTypes[typeIndex]
  form.value.days[dayIndex].tasks[taskIndex].type = t.value
  form.value.days[dayIndex].tasks[taskIndex].typeLabel = t.label
}

async function onSubmit() {
  if (!canSubmit.value) return
  submitting.value = true
  try {
    const days = form.value.days.map((d: any, i: number) => ({
      day: i + 1,
      date: new Date(new Date(form.value.startDate).getTime() + i * 86400000).toISOString(),
      tasks: d.tasks.map((t: any) => ({
        title: t.title,
        type: t.type,
        duration: t.duration || 0,
        reps: t.reps,
        note: t.note,
        status: 0,
      })),
    }))

    await createTrainingPlan({
      title: form.value.title,
      description: form.value.description,
      start_date: form.value.startDate,
      end_date: form.value.endDate,
      days,
    })

    uni.showToast({ title: '创建成功', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 1500)
  } catch (e) {
    uni.showToast({ title: '创建失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.create-page {
  padding: 20rpx;
  padding-bottom: 40rpx;
}
.form-section {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}
.form-item {
  margin-bottom: 20rpx;
}
.form-item.half {
  flex: 1;
}
.form-row {
  display: flex;
  gap: 20rpx;
}
.label {
  font-size: 28rpx;
  font-weight: bold;
  margin-bottom: 12rpx;
  display: block;
}
.input {
  border: 1rpx solid #e0e0e0;
  border-radius: 12rpx;
  padding: 16rpx;
  font-size: 28rpx;
}
.textarea {
  border: 1rpx solid #e0e0e0;
  border-radius: 12rpx;
  padding: 16rpx;
  font-size: 28rpx;
  height: 120rpx;
}
.picker-text {
  border: 1rpx solid #e0e0e0;
  border-radius: 12rpx;
  padding: 16rpx;
  font-size: 28rpx;
  display: block;
  color: #333;
}
.days-section {
  margin-bottom: 20rpx;
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}
.section-title {
  font-size: 30rpx;
  font-weight: bold;
}
.add-btn {
  font-size: 26rpx;
  color: #1cbbb4;
}
.day-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.day-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}
.day-label {
  font-size: 28rpx;
  font-weight: bold;
}
.remove-btn {
  font-size: 24rpx;
  color: #e54d42;
}
.task-item {
  background: #f8f8f8;
  border-radius: 12rpx;
  padding: 16rpx;
  margin-bottom: 12rpx;
}
.task-input {
  border: 1rpx solid #e0e0e0;
  border-radius: 8rpx;
  padding: 12rpx;
  font-size: 26rpx;
  margin-bottom: 8rpx;
  background: #fff;
}
.task-row {
  display: flex;
  gap: 12rpx;
  margin-bottom: 8rpx;
}
.task-picker {
  border: 1rpx solid #e0e0e0;
  border-radius: 8rpx;
  padding: 12rpx;
  font-size: 26rpx;
  background: #fff;
}
.task-input-sm {
  flex: 1;
  border: 1rpx solid #e0e0e0;
  border-radius: 8rpx;
  padding: 12rpx;
  font-size: 26rpx;
  background: #fff;
}
.task-input-num {
  width: 120rpx;
  border: 1rpx solid #e0e0e0;
  border-radius: 8rpx;
  padding: 12rpx;
  font-size: 26rpx;
  background: #fff;
}
.remove-task {
  font-size: 22rpx;
  color: #e54d42;
}
.add-task-btn {
  font-size: 26rpx;
  color: #1cbbb4;
  display: block;
  margin-top: 8rpx;
}
.submit-btn {
  margin-top: 20rpx;
  background: #1cbbb4;
  color: #fff;
  border-radius: 40rpx;
  height: 88rpx;
  line-height: 88rpx;
  font-size: 32rpx;
}
.submit-btn[disabled] {
  opacity: 0.5;
}
</style>
