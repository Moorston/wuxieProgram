<template>
  <view class="public-page">
    <view class="insight-list">
      <view v-for="item in list" :key="item.id" class="insight-card" @tap="goDetail(item.id)">
        <view class="card-user">
          <image class="user-avatar" :src="item.user?.avatar || '/static/default-avatar.png'" />
          <text class="user-name">{{ item.user?.nickname || '匿名' }}</text>
          <text class="card-mood">{{ moodIcon[item.mood] }}</text>
        </view>
        <text class="card-content">{{ item.content }}</text>
        <view v-if="item.images && item.images.length > 0" class="card-images">
          <image v-for="(img, i) in item.images.slice(0, 2)" :key="i" class="card-img" :src="img" mode="aspectFill" />
        </view>
        <view class="card-footer">
          <view v-if="item.tags && item.tags.length > 0" class="card-tags">
            <text v-for="tag in item.tags.slice(0, 3)" :key="tag" class="card-tag">{{ tag }}</text>
          </view>
          <text class="card-time">{{ formatTime(item.created_at) }}</text>
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
      <text>暂无公开感悟</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { listPublicInsights } from '../../api'

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
    const res: any = await listPublicInsights(1, pageSize)
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
    const res: any = await listPublicInsights(1, pageSize)
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
    const res: any = await listPublicInsights(nextPage, pageSize)
    list.value.push(...(res.list || []))
    total.value = res.total || 0
    page.value = nextPage
  } catch (e) {} finally {
    loading.value = false
  }
}

function goDetail(id: string) {
  uni.navigateTo({ url: `/pages/insight/detail?id=${id}` })
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
.public-page {
  padding: 20rpx;
}
.insight-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}
.insight-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
}
.card-user {
  display: flex;
  align-items: center;
  margin-bottom: 12rpx;
}
.user-avatar {
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  margin-right: 12rpx;
}
.user-name {
  font-size: 26rpx;
  flex: 1;
}
.card-mood {
  font-size: 32rpx;
}
.card-content {
  font-size: 28rpx;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.card-images {
  display: flex;
  gap: 8rpx;
  margin-top: 12rpx;
}
.card-img {
  width: 300rpx;
  height: 300rpx;
  border-radius: 8rpx;
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
}
.card-tag {
  font-size: 22rpx;
  background: #f0f0f0;
  padding: 4rpx 12rpx;
  border-radius: 12rpx;
  color: #666;
}
.card-time {
  font-size: 22rpx;
  color: #999;
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
</style>
