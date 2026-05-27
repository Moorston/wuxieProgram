<template>
  <view class="favorites-page">
    <view v-for="item in list" :key="item.id" class="fav-item" @tap="goDetail(item.id)">
      <text class="fav-icon">{{ typeIcon[item.type] }}</text>
      <view class="fav-info">
        <text class="fav-title">{{ item.title }}</text>
        <text class="fav-meta">{{ formatSize(item.file_size) }} · {{ item.created_at?.slice(0, 10) }}</text>
      </view>
      <text class="fav-star" @tap.stop="onUnfavorite(item)">★</text>
    </view>
    <view v-if="loading" class="loading-tip"><text>加载中...</text></view>
    <view v-else-if="list.length === 0" class="empty"><text>暂无收藏</text></view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { listResourceFavorites, toggleResourceFavorite } from '../../api'

const list = ref<any[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const typeIcon: Record<string, string> = { video: '🎬', image: '🖼', document: '📄' }

onMounted(() => loadData())
onPullDownRefresh(async () => { await loadData(); uni.stopPullDownRefresh() })
onReachBottom(() => loadMore())

async function loadData() {
  loading.value = true
  try {
    const res: any = await listResourceFavorites(1, 20)
    list.value = res.list || []
    total.value = res.total || 0
    page.value = 1
  } catch (e) {} finally { loading.value = false }
}

async function loadMore() {
  if (loading.value || list.value.length >= total.value) return
  loading.value = true
  try {
    const res: any = await listResourceFavorites(page.value + 1, 20)
    list.value.push(...(res.list || []))
    page.value++
  } catch (e) {} finally { loading.value = false }
}

async function onUnfavorite(item: any) {
  await toggleResourceFavorite(item.id)
  list.value = list.value.filter(r => r.id !== item.id)
}

function goDetail(id: string) { uni.navigateTo({ url: `/pages/resource/detail?id=${id}` }) }
function formatSize(bytes: number) {
  if (!bytes) return '0B'
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + 'KB'
  return (bytes / 1048576).toFixed(1) + 'MB'
}
</script>

<style scoped>
.favorites-page { padding: 20rpx; }
.fav-item { display: flex; align-items: center; background: #fff; border-radius: 12rpx; padding: 20rpx; margin-bottom: 12rpx; }
.fav-icon { font-size: 40rpx; margin-right: 16rpx; }
.fav-info { flex: 1; }
.fav-title { font-size: 28rpx; display: block; }
.fav-meta { font-size: 22rpx; color: #999; margin-top: 4rpx; display: block; }
.fav-star { font-size: 32rpx; color: #ff9800; }
.loading-tip { text-align: center; padding: 30rpx; color: #999; font-size: 24rpx; }
.empty { text-align: center; padding: 100rpx; color: #999; }
</style>
