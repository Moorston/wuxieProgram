<template>
  <view class="square-page">
    <view class="search-bar">
      <input
        class="search-input"
        placeholder="搜索视频"
        v-model="keyword"
        confirm-type="search"
        @confirm="onSearch"
      />
      <text v-if="keyword" class="search-clear" @tap="onClearSearch">✕</text>
    </view>

    <view class="tabs">
      <text
        v-for="tab in tabs"
        :key="tab.id"
        class="tab-item"
        :class="{ active: currentTab === tab.id }"
        @tap="onTabChange(tab.id)"
      >{{ tab.name }}</text>
    </view>

    <view class="waterfall">
      <view class="column left">
        <view v-for="item in leftList" :key="item.id" class="card" @tap="goDetail(item.id)">
          <image class="cover" :src="item.cover_url" mode="widthFix" />
          <view class="card-info">
            <text class="card-desc">{{ item.description }}</text>
            <view class="card-bottom">
              <view class="author">
                <image class="author-avatar" :src="item.user?.avatar" />
                <text class="author-name">{{ item.user?.nickname }}</text>
              </view>
              <view class="like" @tap.stop="onLike(item)">
                <text :class="{ liked: item.is_liked }">❤</text>
                <text class="like-count">{{ item.like_count }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>
      <view class="column right">
        <view v-for="item in rightList" :key="item.id" class="card" @tap="goDetail(item.id)">
          <image class="cover" :src="item.cover_url" mode="widthFix" />
          <view class="card-info">
            <text class="card-desc">{{ item.description }}</text>
            <view class="card-bottom">
              <view class="author">
                <image class="author-avatar" :src="item.user?.avatar" />
                <text class="author-name">{{ item.user?.nickname }}</text>
              </view>
              <view class="like" @tap.stop="onLike(item)">
                <text :class="{ liked: item.is_liked }">❤</text>
                <text class="like-count">{{ item.like_count }}</text>
              </view>
            </view>
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
    <view v-else-if="list.length === 0 && !loading" class="loading-tip">
      <text>暂无内容</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { getCheckinList, searchCheckinList, toggleLike, getProfile } from '../../api'

const currentTab = ref(0)
const tabs = [
  { id: 0, name: '广场' },
  { id: 1, name: '考核组' },
]

const keyword = ref('')
const list = ref<any[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const pageSize = 10
const userGroupId = ref('')

const noMore = computed(() => list.value.length >= total.value)

const leftList = computed(() => list.value.filter((_, i) => i % 2 === 0))
const rightList = computed(() => list.value.filter((_, i) => i % 2 === 1))

onMounted(async () => {
  try {
    const profile: any = await getProfile()
    if (profile?.group_id) {
      userGroupId.value = profile.group_id
    }
  } catch (e) {}
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
    const groupId = currentTab.value === 1 && userGroupId.value ? userGroupId.value : undefined
    const res: any = keyword.value
      ? await searchCheckinList(keyword.value, 1, pageSize)
      : await getCheckinList(1, pageSize, groupId)
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
    const groupId = currentTab.value === 1 && userGroupId.value ? userGroupId.value : undefined
    const res: any = keyword.value
      ? await searchCheckinList(keyword.value, 1, pageSize)
      : await getCheckinList(1, pageSize, groupId)
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
    const groupId = currentTab.value === 1 && userGroupId.value ? userGroupId.value : undefined
    const res: any = keyword.value
      ? await searchCheckinList(keyword.value, nextPage, pageSize)
      : await getCheckinList(nextPage, pageSize, groupId)
    list.value.push(...(res.list || []))
    total.value = res.total || 0
    page.value = nextPage
  } catch (e) {} finally {
    loading.value = false
  }
}

function onSearch() {
  if (!keyword.value.trim()) return
  loadData()
}

function onClearSearch() {
  keyword.value = ''
  loadData()
}

function onTabChange(tabId: number) {
  currentTab.value = tabId
  loadData()
}

async function onLike(item: any) {
  try {
    const res: any = await toggleLike(item.id)
    item.is_liked = res.liked
    item.like_count += res.liked ? 1 : -1
  } catch (e) {}
}

function goDetail(id: string) {
  uni.navigateTo({ url: `/pages/video-detail/video-detail?id=${id}` })
}
</script>

<style scoped>
.square-page {
  padding: 20rpx;
}
.search-bar {
  background: #fff;
  border-radius: 32rpx;
  padding: 16rpx 24rpx;
  margin-bottom: 20rpx;
  display: flex;
  align-items: center;
}
.search-input {
  font-size: 28rpx;
  flex: 1;
}
.search-clear {
  font-size: 28rpx;
  color: #999;
  padding: 0 12rpx;
}
.tabs {
  display: flex;
  gap: 30rpx;
  margin-bottom: 20rpx;
}
.tab-item {
  font-size: 28rpx;
  color: #999;
  padding-bottom: 8rpx;
}
.tab-item.active {
  color: #1cbbb4;
  border-bottom: 4rpx solid #1cbbb4;
}
.waterfall {
  display: flex;
  gap: 16rpx;
}
.column {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}
.card {
  background: #fff;
  border-radius: 12rpx;
  overflow: hidden;
}
.cover {
  width: 100%;
}
.card-info {
  padding: 16rpx;
}
.card-desc {
  font-size: 26rpx;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.card-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12rpx;
}
.author {
  display: flex;
  align-items: center;
}
.author-avatar {
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  margin-right: 8rpx;
}
.author-name {
  font-size: 22rpx;
  color: #666;
}
.like {
  display: flex;
  align-items: center;
  font-size: 24rpx;
}
.liked {
  color: red;
}
.like-count {
  margin-left: 4rpx;
  font-size: 22rpx;
  color: #999;
}
.loading-tip {
  text-align: center;
  padding: 30rpx;
  color: #999;
  font-size: 24rpx;
}
</style>
