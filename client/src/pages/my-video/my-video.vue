<template>
  <view class="my-video-page">
    <view class="timeline">
      <view v-for="item in list" :key="item.id" class="timeline-item">
        <view class="time-dot"></view>
        <view class="video-card" @tap="goDetail(item.id)">
          <image class="cover" :src="item.cover_url" mode="aspectFill" />
          <view class="card-info">
            <text class="card-desc">{{ item.description }}</text>
            <text class="card-time">{{ formatTime(item.created_at) }}</text>
          </view>
        </view>
      </view>
    </view>

    <view v-if="loading" class="loading-tip">
      <text>加载中...</text>
    </view>
    <view v-else-if="noMore && list.length > 0" class="loading-tip">
      <text>没有更多了</text>
    </view>

    <view v-if="list.length === 0 && !loading" class="empty">
      <text>暂无打卡记录</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { getMyCheckins } from '../../api'

const list = ref<any[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const pageSize = 10

const noMore = computed(() => list.value.length >= total.value)

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
    const res: any = await getMyCheckins(1, pageSize)
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
    const res: any = await getMyCheckins(1, pageSize)
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
    const res: any = await getMyCheckins(nextPage, pageSize)
    list.value.push(...(res.list || []))
    total.value = res.total || 0
    page.value = nextPage
  } catch (e) {} finally {
    loading.value = false
  }
}

function formatTime(ts: string) {
  if (!ts) return ''
  const d = new Date(ts)
  return `${d.getMonth() + 1}-${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}

function goDetail(id: string) {
  uni.navigateTo({ url: `/pages/video-detail/video-detail?id=${id}` })
}
</script>

<style scoped>
.my-video-page {
  padding: 20rpx;
}
.timeline {
  padding-left: 30rpx;
  border-left: 4rpx solid #1cbbb4;
}
.timeline-item {
  position: relative;
  margin-bottom: 30rpx;
  padding-left: 30rpx;
}
.time-dot {
  position: absolute;
  left: -38rpx;
  top: 20rpx;
  width: 16rpx;
  height: 16rpx;
  background: #1cbbb4;
  border-radius: 50%;
}
.video-card {
  background: #fff;
  border-radius: 12rpx;
  overflow: hidden;
}
.cover {
  width: 100%;
  height: 300rpx;
}
.card-info {
  padding: 16rpx;
}
.card-desc {
  font-size: 28rpx;
  display: block;
}
.card-time {
  font-size: 22rpx;
  color: #999;
  margin-top: 8rpx;
  display: block;
}
.empty {
  text-align: center;
  padding: 100rpx;
  color: #999;
}
.loading-tip {
  text-align: center;
  padding: 30rpx;
  color: #999;
  font-size: 24rpx;
}
</style>
