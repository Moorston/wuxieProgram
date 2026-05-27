<template>
  <view class="resource-page">
    <view class="search-bar">
      <input class="search-input" placeholder="搜索资料" v-model="keyword" confirm-type="search" @confirm="onSearch" />
      <text v-if="keyword" class="search-clear" @tap="keyword = ''; loadData()">✕</text>
    </view>

    <view class="filter-bar">
      <picker :value="typeIndex" :range="typeOptions" @change="onTypeChange">
        <text class="filter-btn">{{ currentType || '全部类型' }}</text>
      </picker>
      <picker :value="categoryIndex" :range="categoryOptions" @change="onCategoryChange">
        <text class="filter-btn">{{ currentCategory || '全部分类' }}</text>
      </picker>
      <picker :value="sortIndex" :range="sortOptions" @change="onSortChange">
        <text class="filter-btn">{{ sortOptions[sortIndex] }}</text>
      </picker>
      <text class="filter-btn" @tap="viewMode = viewMode === 'waterfall' ? 'list' : 'waterfall'">
        {{ viewMode === 'waterfall' ? '📋' : '🖼' }}
      </text>
    </view>

    <view v-if="viewMode === 'waterfall'" class="waterfall">
      <view class="column left">
        <view v-for="item in leftList" :key="item.id" class="card" @tap="goDetail(item.id)">
          <image v-if="item.cover_url" class="card-cover" :src="item.cover_url" mode="widthFix" />
          <view v-else class="card-cover-placeholder">{{ typeIcon[item.type] }}</view>
          <view class="card-info">
            <text class="card-title">{{ item.title }}</text>
            <text class="card-meta">{{ typeLabel[item.type] }} · {{ formatSize(item.file_size) }}</text>
          </view>
        </view>
      </view>
      <view class="column right">
        <view v-for="item in rightList" :key="item.id" class="card" @tap="goDetail(item.id)">
          <image v-if="item.cover_url" class="card-cover" :src="item.cover_url" mode="widthFix" />
          <view v-else class="card-cover-placeholder">{{ typeIcon[item.type] }}</view>
          <view class="card-info">
            <text class="card-title">{{ item.title }}</text>
            <text class="card-meta">{{ typeLabel[item.type] }} · {{ formatSize(item.file_size) }}</text>
          </view>
        </view>
      </view>
    </view>

    <view v-else class="list-view">
      <view v-for="item in list" :key="item.id" class="list-item" @tap="goDetail(item.id)">
        <text class="list-icon">{{ typeIcon[item.type] }}</text>
        <view class="list-info">
          <text class="list-title">{{ item.title }}</text>
          <text class="list-meta">{{ typeLabel[item.type] }} · {{ formatSize(item.file_size) }} · {{ item.created_at?.slice(0, 10) }}</text>
        </view>
        <text v-if="item.is_favorite" class="fav-icon">★</text>
      </view>
    </view>

    <view v-if="loading" class="loading-tip"><text>加载中...</text></view>
    <view v-else-if="noMore && list.length > 0" class="loading-tip"><text>没有更多了</text></view>
    <view v-else-if="list.length === 0 && !loading" class="empty"><text>暂无资料</text></view>

    <view class="fab" @tap="goUpload"><text class="fab-icon">+</text></view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { listResources } from '../../api'

const typeOptions = ['全部', '视频', '图片', '文档']
const typeValues = ['', 'video', 'image', 'document']
const categoryOptions = ['全部', '基本功', '套路', '散打', '太极', '理论']
const sortOptions = ['按时间', '按名称', '按大小']

const typeIndex = ref(0)
const categoryIndex = ref(0)
const sortIndex = ref(0)
const currentType = ref('')
const currentCategory = ref('')
const keyword = ref('')
const viewMode = ref<'waterfall' | 'list'>('waterfall')

const list = ref<any[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const pageSize = 20

const noMore = computed(() => list.value.length >= total.value)
const leftList = computed(() => list.value.filter((_, i) => i % 2 === 0))
const rightList = computed(() => list.value.filter((_, i) => i % 2 === 1))

const typeIcon: Record<string, string> = { video: '🎬', image: '🖼', document: '📄' }
const typeLabel: Record<string, string> = { video: '视频', image: '图片', document: '文档' }

onMounted(() => loadData())

onPullDownRefresh(async () => { await refreshData(); uni.stopPullDownRefresh() })
onReachBottom(() => loadMore())

function buildParams() {
  return {
    type: currentType.value,
    category: currentCategory.value,
    keyword: keyword.value,
    sort: ['time', 'name', 'size'][sortIndex.value],
    page: 1,
    pageSize,
  }
}

async function loadData() {
  if (loading.value) return
  loading.value = true
  try {
    const res: any = await listResources(buildParams())
    list.value = res.list || []
    total.value = res.total || 0
    page.value = 1
  } catch (e) {} finally { loading.value = false }
}

async function refreshData() {
  page.value = 1
  loading.value = true
  try {
    const res: any = await listResources({ ...buildParams(), page: 1 })
    list.value = res.list || []
    total.value = res.total || 0
  } catch (e) {} finally { loading.value = false }
}

async function loadMore() {
  if (loading.value || noMore.value) return
  loading.value = true
  try {
    const nextPage = page.value + 1
    const res: any = await listResources({ ...buildParams(), page: nextPage })
    list.value.push(...(res.list || []))
    total.value = res.total || 0
    page.value = nextPage
  } catch (e) {} finally { loading.value = false }
}

function onSearch() { loadData() }
function onTypeChange(e: any) { typeIndex.value = e.detail.value; currentType.value = typeValues[typeIndex.value]; loadData() }
function onCategoryChange(e: any) { categoryIndex.value = e.detail.value; currentCategory.value = categoryIndex.value > 0 ? categoryOptions[categoryIndex.value] : ''; loadData() }
function onSortChange(e: any) { sortIndex.value = e.detail.value; loadData() }

function formatSize(bytes: number) {
  if (!bytes) return '0B'
  if (bytes < 1024) return bytes + 'B'
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + 'KB'
  if (bytes < 1073741824) return (bytes / 1048576).toFixed(1) + 'MB'
  return (bytes / 1073741824).toFixed(1) + 'GB'
}

function goDetail(id: string) { uni.navigateTo({ url: `/pages/resource/detail?id=${id}` }) }
function goUpload() { uni.navigateTo({ url: '/pages/resource/upload' }) }
</script>

<style scoped>
.resource-page { padding: 20rpx; padding-bottom: 120rpx; }
.search-bar { background: #fff; border-radius: 32rpx; padding: 16rpx 24rpx; margin-bottom: 16rpx; display: flex; align-items: center; }
.search-input { flex: 1; font-size: 28rpx; }
.search-clear { font-size: 28rpx; color: #999; padding: 0 12rpx; }
.filter-bar { display: flex; gap: 12rpx; margin-bottom: 16rpx; flex-wrap: wrap; }
.filter-btn { background: #fff; border-radius: 24rpx; padding: 10rpx 20rpx; font-size: 24rpx; color: #666; }
.waterfall { display: flex; gap: 12rpx; }
.column { flex: 1; display: flex; flex-direction: column; gap: 12rpx; }
.card { background: #fff; border-radius: 12rpx; overflow: hidden; }
.card-cover { width: 100%; }
.card-cover-placeholder { width: 100%; height: 200rpx; background: #f0f0f0; display: flex; align-items: center; justify-content: center; font-size: 48rpx; }
.card-info { padding: 12rpx; }
.card-title { font-size: 26rpx; display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.card-meta { font-size: 22rpx; color: #999; margin-top: 4rpx; display: block; }
.list-view { background: #fff; border-radius: 12rpx; }
.list-item { display: flex; align-items: center; padding: 20rpx; border-bottom: 1rpx solid #f0f0f0; }
.list-item:last-child { border-bottom: none; }
.list-icon { font-size: 40rpx; margin-right: 16rpx; }
.list-info { flex: 1; }
.list-title { font-size: 28rpx; display: block; }
.list-meta { font-size: 22rpx; color: #999; margin-top: 4rpx; display: block; }
.fav-icon { font-size: 32rpx; color: #ff9800; }
.loading-tip { text-align: center; padding: 30rpx; color: #999; font-size: 24rpx; }
.empty { text-align: center; padding: 100rpx; color: #999; }
.fab { position: fixed; right: 40rpx; bottom: 100rpx; width: 100rpx; height: 100rpx; background: #1cbbb4; border-radius: 50%; display: flex; align-items: center; justify-content: center; box-shadow: 0 4rpx 16rpx rgba(28,187,180,0.4); }
.fab-icon { font-size: 48rpx; color: #fff; }
</style>
