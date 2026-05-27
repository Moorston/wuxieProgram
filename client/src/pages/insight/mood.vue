<template>
  <view class="mood-page">
    <view class="stats-card">
      <text class="card-title">近30天心情分布</text>
      <view class="mood-bars">
        <view v-for="m in moodList" :key="m.value" class="mood-bar-item">
          <text class="bar-label">{{ m.icon }} {{ m.label }}</text>
          <view class="bar-track">
            <view class="bar-fill" :style="{ width: getPercent(m.value) + '%', background: m.color }"></view>
          </view>
          <text class="bar-count">{{ stats[m.value] || 0 }}</text>
        </view>
      </view>
    </view>

    <view class="summary-card">
      <text class="card-title">心情总结</text>
      <view class="summary-row">
        <view class="summary-item">
          <text class="summary-value">{{ total }}</text>
          <text class="summary-label">总感悟数</text>
        </view>
        <view class="summary-item">
          <text class="summary-value">{{ topMood.icon }}</text>
          <text class="summary-label">最常见心情</text>
        </view>
        <view class="summary-item">
          <text class="summary-value">{{ breakthroughCount }}</text>
          <text class="summary-label">突破次数</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getMoodStats } from '../../api'

const stats = ref<Record<string, number>>({})

const moodList = [
  { value: 'breakthrough', icon: '🔥', label: '突破', color: '#ff5722' },
  { value: 'good', icon: '😊', label: '满意', color: '#4caf50' },
  { value: 'normal', icon: '😐', label: '一般', color: '#9e9e9e' },
  { value: 'confused', icon: '🤔', label: '困惑', color: '#ff9800' },
  { value: 'low', icon: '😔', label: '低落', color: '#2196f3' },
]

const total = computed(() => Object.values(stats.value).reduce((a, b) => a + b, 0))
const breakthroughCount = computed(() => stats.value.breakthrough || 0)
const topMood = computed(() => {
  let max = 0
  let top = moodList[0]
  for (const m of moodList) {
    if ((stats.value[m.value] || 0) > max) {
      max = stats.value[m.value] || 0
      top = m
    }
  }
  return top
})

function getPercent(value: string) {
  if (total.value === 0) return 0
  return ((stats.value[value] || 0) / total.value * 100)
}

onMounted(async () => {
  try {
    const res: any = await getMoodStats(30)
    stats.value = res || {}
  } catch (e) {}
})
</script>

<style scoped>
.mood-page {
  padding: 20rpx;
}
.stats-card, .summary-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}
.card-title {
  font-size: 30rpx;
  font-weight: bold;
  margin-bottom: 20rpx;
  display: block;
}
.mood-bars {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}
.mood-bar-item {
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.bar-label {
  width: 120rpx;
  font-size: 26rpx;
}
.bar-track {
  flex: 1;
  height: 24rpx;
  background: #f0f0f0;
  border-radius: 12rpx;
  overflow: hidden;
}
.bar-fill {
  height: 100%;
  border-radius: 12rpx;
  transition: width 0.3s;
}
.bar-count {
  width: 60rpx;
  text-align: right;
  font-size: 26rpx;
  color: #666;
}
.summary-row {
  display: flex;
  gap: 16rpx;
}
.summary-item {
  flex: 1;
  text-align: center;
  padding: 16rpx;
  background: #f8f8f8;
  border-radius: 12rpx;
}
.summary-value {
  font-size: 40rpx;
  display: block;
}
.summary-label {
  font-size: 22rpx;
  color: #999;
  margin-top: 4rpx;
  display: block;
}
</style>
