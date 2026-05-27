<template>
  <view class="stats-page" v-if="stats">
    <view class="quota-card">
      <text class="card-title">存储空间</text>
      <view class="quota-bar">
        <view class="quota-fill" :style="{ width: Math.min(stats.usage_percent, 100) + '%' }"></view>
      </view>
      <text class="quota-text">{{ formatSize(stats.total_size) }} / {{ formatSize(stats.quota) }}</text>
      <text class="quota-percent">{{ stats.usage_percent.toFixed(1) }}% 已使用</text>
    </view>

    <view class="type-card">
      <text class="card-title">文件类型分布</text>
      <view v-for="(size, type) in stats.type_stats" :key="type" class="type-item">
        <view class="type-info">
          <text class="type-icon">{{ typeIcon[type] }}</text>
          <text class="type-name">{{ typeLabel[type] }}</text>
          <text class="type-count">{{ stats.type_counts[type] || 0 }}个</text>
        </view>
        <view class="type-bar">
          <view class="type-fill" :style="{ width: (stats.total_size > 0 ? size / stats.total_size * 100 : 0) + '%' }"></view>
        </view>
        <text class="type-size">{{ formatSize(size) }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getResourceStats } from '../../api'

const stats = ref<any>(null)
const typeIcon: Record<string, string> = { video: '🎬', image: '🖼', document: '📄' }
const typeLabel: Record<string, string> = { video: '视频', image: '图片', document: '文档' }

onMounted(async () => {
  try { stats.value = await getResourceStats() } catch (e) {}
})

function formatSize(bytes: number) {
  if (!bytes) return '0B'
  if (bytes < 1024) return bytes + 'B'
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + 'KB'
  if (bytes < 1073741824) return (bytes / 1048576).toFixed(1) + 'MB'
  return (bytes / 1073741824).toFixed(1) + 'GB'
}
</script>

<style scoped>
.stats-page { padding: 20rpx; }
.quota-card, .type-card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; }
.card-title { font-size: 30rpx; font-weight: bold; margin-bottom: 16rpx; display: block; }
.quota-bar { height: 16rpx; background: #f0f0f0; border-radius: 8rpx; overflow: hidden; }
.quota-fill { height: 100%; background: linear-gradient(90deg, #1cbbb4, #0081ff); border-radius: 8rpx; transition: width 0.3s; }
.quota-text { font-size: 28rpx; margin-top: 12rpx; display: block; }
.quota-percent { font-size: 24rpx; color: #999; display: block; }
.type-item { margin-bottom: 16rpx; }
.type-info { display: flex; align-items: center; gap: 12rpx; margin-bottom: 8rpx; }
.type-icon { font-size: 32rpx; }
.type-name { flex: 1; font-size: 26rpx; }
.type-count { font-size: 24rpx; color: #999; }
.type-bar { height: 8rpx; background: #f0f0f0; border-radius: 4rpx; overflow: hidden; }
.type-fill { height: 100%; background: #1cbbb4; border-radius: 4rpx; }
.type-size { font-size: 22rpx; color: #999; margin-top: 4rpx; display: block; }
</style>
