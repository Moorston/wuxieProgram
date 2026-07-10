<template>
  <view class="competition-page">
    <view class="section-title">🏆 赛事活动</view>

    <view v-if="loading" class="empty-tip">加载中...</view>

    <view v-else-if="competitions.length === 0" class="empty-tip">暂无赛事活动</view>

    <view v-else class="comp-list">
      <view class="comp-card" v-for="comp in competitions" :key="comp.id" @click="goDetail(comp.id)">
        <view class="comp-header">
          <text class="comp-title">{{ comp.title }}</text>
          <text class="comp-status" :class="getStatusClass(comp)">
            {{ getStatusText(comp) }}
          </text>
        </view>
        <text class="comp-desc" v-if="comp.description">{{ comp.description }}</text>
        <view class="comp-meta">
          <text class="comp-date">📅 {{ formatDate(comp.start_date) }} - {{ formatDate(comp.end_date) }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getCompetitions } from '../api/competition'

const competitions = ref<any[]>([])
const loading = ref(true)

function getStatusClass(comp: any) {
  const now = Date.now()
  const start = new Date(comp.start_date).getTime()
  const end = new Date(comp.end_date).getTime()
  if (comp.status === 0) return 'status-draft'
  if (now >= start && now <= end) return 'status-active'
  return 'status-ended'
}

function getStatusText(comp: any) {
  const now = Date.now()
  const start = new Date(comp.start_date).getTime()
  const end = new Date(comp.end_date).getTime()
  if (comp.status === 0) return '草稿'
  if (now < start) return '未开始'
  if (now >= start && now <= end) return '进行中'
  return '已结束'
}

function formatDate(date: string) {
  if (!date) return ''
  return date.slice(0, 10)
}

function goDetail(id: string) {
  uni.navigateTo({ url: `/pages/competition/detail?id=${id}` })
}

onMounted(async () => {
  try {
    const res: any = await getCompetitions()
    competitions.value = res?.list || []
  } catch (e) {
    console.error('load competitions failed:', e)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.competition-page { padding: 24rpx; }
.section-title { font-size: 32rpx; font-weight: 600; margin-bottom: 24rpx; }
.comp-list { display: flex; flex-direction: column; gap: 16rpx; }
.comp-card { background: #fff; border-radius: 16rpx; padding: 24rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.06); }
.comp-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8rpx; }
.comp-title { font-size: 32rpx; font-weight: 600; }
.comp-status { font-size: 22rpx; padding: 4rpx 12rpx; border-radius: 12rpx; }
.status-active { background: #dcfce7; color: #16a34a; }
.status-ended { background: #f1f5f9; color: #64748b; }
.status-draft { background: #fef3c7; color: #d97706; }
.comp-desc { font-size: 26rpx; color: #666; margin-bottom: 8rpx; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.comp-meta { font-size: 24rpx; color: #999; }
.comp-date { display: block; }
.empty-tip { text-align: center; color: #999; padding: 60rpx; font-size: 28rpx; }
</style>
