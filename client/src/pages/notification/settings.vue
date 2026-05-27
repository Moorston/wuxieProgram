<template>
  <view class="settings-page">
    <view class="section">
      <text class="section-title">通知类型</text>
      <view class="setting-item">
        <text class="setting-label">点赞通知</text>
        <switch :checked="settings.like_notify" @change="onToggle('like_notify', $event)" color="#1cbbb4" />
      </view>
      <view class="setting-item">
        <text class="setting-label">评论通知</text>
        <switch :checked="settings.comment_notify" @change="onToggle('comment_notify', $event)" color="#1cbbb4" />
      </view>
      <view class="setting-item">
        <text class="setting-label">组通知</text>
        <switch :checked="settings.group_notify" @change="onToggle('group_notify', $event)" color="#1cbbb4" />
      </view>
    </view>

    <view class="section">
      <text class="section-title">训练提醒</text>
      <view class="setting-item">
        <text class="setting-label">每日训练提醒</text>
        <switch :checked="settings.plan_remind" @change="onToggle('plan_remind', $event)" color="#1cbbb4" />
      </view>
      <view class="setting-item" v-if="settings.plan_remind">
        <text class="setting-label">提醒时间</text>
        <picker mode="time" :value="settings.plan_remind_time" @change="onTimeChange">
          <text class="time-text">{{ settings.plan_remind_time || '20:00' }}</text>
        </picker>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getNotificationSettings, updateNotificationSettings } from '../../api'

const settings = ref<any>({
  like_notify: true,
  comment_notify: true,
  plan_remind: true,
  plan_remind_time: '20:00',
  group_notify: true,
})

onMounted(async () => {
  try {
    const res: any = await getNotificationSettings()
    settings.value = res
  } catch (e) {}
})

async function onToggle(field: string, e: any) {
  settings.value[field] = e.detail.value
  await saveSettings({ [field]: e.detail.value })
}

async function onTimeChange(e: any) {
  settings.value.plan_remind_time = e.detail.value
  await saveSettings({ plan_remind_time: e.detail.value })
}

async function saveSettings(data: any) {
  try {
    await updateNotificationSettings(data)
    uni.showToast({ title: '已保存', icon: 'success' })
  } catch (e) {
    uni.showToast({ title: '保存失败', icon: 'none' })
  }
}
</script>

<style scoped>
.settings-page {
  padding: 20rpx;
}
.section {
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
.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}
.setting-item:last-child {
  border-bottom: none;
}
.setting-label {
  font-size: 28rpx;
}
.time-text {
  font-size: 28rpx;
  color: #1cbbb4;
}
</style>
