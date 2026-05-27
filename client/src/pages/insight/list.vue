<template>
  <view class="insight-list-page">
    <view class="filter-bar">
      <picker :value="moodIndex" :range="moodOptions" @change="onMoodChange">
        <text class="filter-btn">{{ currentMood || '全部心情' }}</text>
      </picker>
      <text class="filter-btn" @tap="goTags">标签管理</text>
      <text class="filter-btn" @tap="goMoodStats">心情统计</text>
    </view>

    <view v-if="currentTag" class="tag-filter">
      <text class="tag-label">标签: {{ currentTag }}</text>
      <text class="tag-clear" @tap="clearTag">✕</text>
    </view>

    <view class="timeline">
      <view v-for="item in list" :key="item.id" class="insight-card" @tap="goDetail(item.id)">
        <view class="card-header">
          <text class="card-mood">{{ moodIcon[item.mood] }}</text>
          <text class="card-time">{{ formatTime(item.created_at) }}</text>
        </view>
        <text class="card-content">{{ item.content }}</text>
        <view v-if="item.images && item.images.length > 0" class="card-images">
          <image v-for="(img, i) in item.images.slice(0, 3)" :key="i" class="card-img" :src="img" mode="aspectFill" />
          <text v-if="item.images.length > 3" class="img-more">+{{ item.images.length - 3 }}</text>
        </view>
        <view class="card-footer">
          <view v-if="item.tags && item.tags.length > 0" class="card-tags">
            <text v-for="tag in item.tags.slice(0, 3)" :key="tag" class="card-tag">{{ tag }}</text>
          </view>
          <text v-if="item.visibility === 'public'" class="public-badge">公开</text>
        </view>
      </view>
    </view>

    <view v-if="loading" class="loading-tip">
      <text>加载中...</text>
    </view>
    <view v-else-if="noMore && list.length > 0" class="loading-tip">
      <text>没有更多了</text>
    </view>
    <view v-else-if="list.length === 0 && !loading" class="empty">
      <text>暂无感悟笔记</text>
    </view>

    <view class="fab" @tap="goCreate">
      <text class="fab-icon">+</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { listInsights } from '../../api'

const moodOptions = ['全部', '突破', '满意', '一般', '困惑', '低落']
const moodValues = ['', 'breakthrough', 'good', 'normal', 'confused', 'low']
const moodIndex = ref(0)
const currentMood = ref('')
const currentTag = ref('')

const list = ref<any[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const pageSize = 20

const noMore = computed(() => list.value.length >= total.value)

const moodIcon: Record<string, string> = {
  breakthrough: '🔥',
  good: '😊',
  normal: '😐',
  confused: '🤔',
  low: '😔',
}

onLoad((options) => {
  if (options?.tag) {
    currentTag.value = options.tag
  }
})

onMounted(() => {
  loadData()
})

onPullDownRefresh(async () => {
  await refreshData()
  uni.stopPullDownRefresh()
})

onReachBottom(() => {
  loadMore()
})

async function loadData() {
  if (loading.value) return
  loading.value = true
  try {
    const res: any = await listInsights(1, pageSize, currentTag.value || undefined, currentMood.value || undefined)
    list.value = res.list || []
    total.value = res.total || 0
    page.value = 1
  } catch (e) {} finally {
    loading.value = false
  }
}

async function refreshData() {
  page.value = 1
  loading.value = true
  try {
    const res: any = await listInsights(1, pageSize, currentTag.value || undefined, currentMood.value || undefined)
    list.value = res.list || []
    total.value = res.total || 0
    page.value = 1
  } catch (e) {} finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loading.value || noMore.value) return
  loading.value = true
  try {
    const nextPage = page.value + 1
    const res: any = await listInsights(nextPage, pageSize, currentTag.value || undefined, currentMood.value || undefined)
    list.value.push(...(res.list || []))
    total.value = res.total || 0
    page.value = nextPage
  } catch (e) {} finally {
    loading.value = false
  }
}

function onMoodChange(e: any) {
  moodIndex.value = e.detail.value
  currentMood.value = moodValues[moodIndex.value]
  loadData()
}

function clearTag() {
  currentTag.value = ''
  loadData()
}

function goDetail(id: string) {
  uni.navigateTo({ url: `/pages/insight/detail?id=${id}` })
}

function goCreate() {
  uni.navigateTo({ url: '/pages/insight/create' })
}

function goTags() {
  uni.navigateTo({ url: '/pages/insight/tags' })
}

function goMoodStats() {
  uni.navigateTo({ url: '/pages/insight/mood' })
}

function formatTime(ts: string) {
  if (!ts) return ''
  const d = new Date(ts)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  if (diff < 86400000) return '今天'
  if (diff < 172800000) return '昨天'
  return `${d.getMonth() + 1}月${d.getDate()}日`
}
</script>

<style scoped>
.insight-list-page {
  padding: 20rpx;
  padding-bottom: 120rpx;
}
.filter-bar {
  display: flex;
  gap: 16rpx;
  margin-bottom: 20rpx;
}
.filter-btn {
  background: #fff;
  border-radius: 24rpx;
  padding: 12rpx 24rpx;
  font-size: 26rpx;
  color: #666;
}
.tag-filter {
  display: flex;
  align-items: center;
  background: #e3f2fd;
  border-radius: 20rpx;
  padding: 8rpx 16rpx;
  margin-bottom: 16rpx;
  align-self: flex-start;
}
.tag-label {
  font-size: 24rpx;
  color: #2196f3;
}
.tag-clear {
  font-size: 24rpx;
  color: #2196f3;
  margin-left: 8rpx;
}
.timeline {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}
.insight-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
}
.card-header {
  display: flex;
  align-items: center;
  margin-bottom: 12rpx;
}
.card-mood {
  font-size: 36rpx;
  margin-right: 12rpx;
}
.card-time {
  font-size: 24rpx;
  color: #999;
}
.card-content {
  font-size: 28rpx;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.card-images {
  display: flex;
  gap: 8rpx;
  margin-top: 12rpx;
}
.card-img {
  width: 160rpx;
  height: 160rpx;
  border-radius: 8rpx;
}
.img-more {
  width: 160rpx;
  height: 160rpx;
  border-radius: 8rpx;
  background: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24rpx;
  color: #999;
}
.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12rpx;
}
.card-tags {
  display: flex;
  gap: 8rpx;
  flex-wrap: wrap;
}
.card-tag {
  font-size: 22rpx;
  background: #f0f0f0;
  padding: 4rpx 12rpx;
  border-radius: 12rpx;
  color: #666;
}
.public-badge {
  font-size: 22rpx;
  background: #e8f5e9;
  color: #4caf50;
  padding: 4rpx 12rpx;
  border-radius: 12rpx;
}
.loading-tip {
  text-align: center;
  padding: 30rpx;
  color: #999;
  font-size: 24rpx;
}
.empty {
  text-align: center;
  padding: 100rpx;
  color: #999;
}
.fab {
  position: fixed;
  right: 40rpx;
  bottom: 100rpx;
  width: 100rpx;
  height: 100rpx;
  background: #1cbbb4;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 16rpx rgba(28, 187, 180, 0.4);
}
.fab-icon {
  font-size: 48rpx;
  color: #fff;
}
</style>
