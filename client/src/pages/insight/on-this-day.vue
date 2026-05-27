<template>
  <view class="on-this-day-page">
    <view class="header">
      <text class="header-title">历史今日</text>
      <text class="header-date">{{ todayStr }}</text>
    </view>

    <view v-if="insights.length > 0">
      <view v-for="item in insights" :key="item.id" class="insight-card" @tap="goDetail(item.id)">
        <text class="card-year">{{ new Date(item.created_at).getFullYear() }}年</text>
        <text class="card-mood">{{ moodIcon[item.mood] }}</text>
        <text class="card-content">{{ item.content }}</text>
        <view v-if="item.tags && item.tags.length > 0" class="card-tags">
          <text v-for="tag in item.tags" :key="tag" class="card-tag">{{ tag }}</text>
        </view>
      </view>
    </view>

    <view v-else class="empty">
      <text class="empty-icon">📖</text>
      <text class="empty-text">历史今日没有感悟记录</text>
      <text class="empty-hint">继续坚持记录，明年今天回来看</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getOnThisDay } from '../../api'

const insights = ref<any[]>([])

const todayStr = computed(() => {
  const d = new Date()
  return `${d.getMonth() + 1}月${d.getDate()}日`
})

const moodIcon: Record<string, string> = {
  breakthrough: '🔥',
  good: '😊',
  normal: '😐',
  confused: '🤔',
  low: '😔',
}

onMounted(async () => {
  try {
    const res: any = await getOnThisDay()
    insights.value = res || []
  } catch (e) {}
})

function goDetail(id: string) {
  uni.navigateTo({ url: `/pages/insight/detail?id=${id}` })
}
</script>

<style scoped>
.on-this-day-page {
  padding: 20rpx;
}
.header {
  background: linear-gradient(135deg, #1cbbb4, #0081ff);
  border-radius: 16rpx;
  padding: 40rpx;
  color: #fff;
  text-align: center;
  margin-bottom: 20rpx;
}
.header-title {
  font-size: 36rpx;
  font-weight: bold;
  display: block;
}
.header-date {
  font-size: 28rpx;
  opacity: 0.9;
  margin-top: 8rpx;
  display: block;
}
.insight-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.card-year {
  font-size: 24rpx;
  color: #999;
  display: block;
  margin-bottom: 8rpx;
}
.card-mood {
  font-size: 36rpx;
  margin-bottom: 8rpx;
  display: block;
}
.card-content {
  font-size: 28rpx;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.card-tags {
  display: flex;
  gap: 8rpx;
  margin-top: 12rpx;
}
.card-tag {
  font-size: 22rpx;
  background: #f0f0f0;
  padding: 4rpx 12rpx;
  border-radius: 12rpx;
  color: #666;
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
  font-size: 30rpx;
  color: #333;
  display: block;
}
.empty-hint {
  font-size: 26rpx;
  color: #999;
  margin-top: 12rpx;
  display: block;
}
</style>
