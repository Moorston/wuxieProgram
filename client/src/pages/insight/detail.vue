<template>
  <view class="detail-page" v-if="insight">
    <view class="detail-header">
      <text class="detail-mood">{{ moodIcon[insight.mood] }}</text>
      <text class="detail-time">{{ formatFullTime(insight.created_at) }}</text>
    </view>

    <view class="detail-content">
      <text class="content-text">{{ insight.content }}</text>
    </view>

    <view v-if="insight.images && insight.images.length > 0" class="detail-images">
      <image v-for="(img, i) in insight.images" :key="i" class="detail-img" :src="img" mode="widthFix" @tap="previewImage(i)" />
    </view>

    <view v-if="insight.tags && insight.tags.length > 0" class="detail-tags">
      <text v-for="tag in insight.tags" :key="tag" class="detail-tag" @tap="goTagInsights(tag)">{{ tag }}</text>
    </view>

    <view class="detail-meta">
      <text v-if="insight.visibility === 'public'" class="public-badge">🌐 公开</text>
      <text v-if="insight.checkin_id" class="link-badge" @tap="goCheckin">📹 关联打卡</text>
      <text v-if="insight.plan_id" class="link-badge" @tap="goPlan">📋 关联计划</text>
    </view>

    <view class="detail-actions">
      <text class="action-btn" @tap="onEdit">编辑</text>
      <text class="action-btn delete" @tap="onDelete">删除</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getInsight, deleteInsight } from '../../api'

const insightId = ref('')
const insight = ref<any>(null)

const moodIcon: Record<string, string> = {
  breakthrough: '🔥',
  good: '😊',
  normal: '😐',
  confused: '🤔',
  low: '😔',
}

onLoad((options) => {
  insightId.value = options?.id || ''
  loadInsight()
})

async function loadInsight() {
  try {
    insight.value = await getInsight(insightId.value)
  } catch (e) {}
}

function formatFullTime(ts: string) {
  if (!ts) return ''
  const d = new Date(ts)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

function previewImage(index: number) {
  uni.previewImage({
    urls: insight.value.images,
    current: insight.value.images[index],
  })
}

function goTagInsights(tag: string) {
  uni.navigateTo({ url: `/pages/insight/list?tag=${encodeURIComponent(tag)}` })
}

function goCheckin() {
  uni.navigateTo({ url: `/pages/video-detail/video-detail?id=${insight.value.checkin_id}` })
}

function goPlan() {
  uni.navigateTo({ url: `/pages/training/detail?id=${insight.value.plan_id}` })
}

function onEdit() {
  uni.navigateTo({ url: `/pages/insight/create?id=${insightId.value}` })
}

async function onDelete() {
  uni.showModal({
    title: '确认删除',
    content: '确定要删除这条感悟吗？',
    success: async (res) => {
      if (res.confirm) {
        try {
          await deleteInsight(insightId.value)
          uni.showToast({ title: '已删除', icon: 'success' })
          setTimeout(() => uni.navigateBack(), 1500)
        } catch (e) {
          uni.showToast({ title: '删除失败', icon: 'none' })
        }
      }
    },
  })
}
</script>

<style scoped>
.detail-page {
  padding: 20rpx;
}
.detail-header {
  display: flex;
  align-items: center;
  margin-bottom: 20rpx;
}
.detail-mood {
  font-size: 48rpx;
  margin-right: 16rpx;
}
.detail-time {
  font-size: 26rpx;
  color: #999;
}
.detail-content {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.content-text {
  font-size: 30rpx;
  line-height: 1.8;
}
.detail-images {
  background: #fff;
  border-radius: 16rpx;
  padding: 16rpx;
  margin-bottom: 16rpx;
}
.detail-img {
  width: 100%;
  border-radius: 8rpx;
  margin-bottom: 8rpx;
}
.detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-bottom: 16rpx;
}
.detail-tag {
  background: #f0f0f0;
  padding: 8rpx 20rpx;
  border-radius: 20rpx;
  font-size: 24rpx;
  color: #666;
}
.detail-meta {
  display: flex;
  gap: 12rpx;
  margin-bottom: 16rpx;
}
.public-badge, .link-badge {
  font-size: 24rpx;
  padding: 8rpx 16rpx;
  border-radius: 20rpx;
}
.public-badge {
  background: #e8f5e9;
  color: #4caf50;
}
.link-badge {
  background: #e3f2fd;
  color: #2196f3;
}
.detail-actions {
  display: flex;
  gap: 16rpx;
}
.action-btn {
  flex: 1;
  text-align: center;
  padding: 20rpx;
  background: #fff;
  border-radius: 12rpx;
  font-size: 28rpx;
}
.action-btn.delete {
  color: #e54d42;
}
</style>
