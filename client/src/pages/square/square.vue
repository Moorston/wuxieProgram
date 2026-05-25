<template>
  <view class="square-page">
    <view class="search-bar">
      <input class="search-input" placeholder="搜索视频" />
    </view>

    <view class="tabs">
      <text
        v-for="tab in tabs"
        :key="tab.id"
        class="tab-item"
        :class="{ active: currentTab === tab.id }"
        @tap="currentTab = tab.id"
      >{{ tab.name }}</text>
    </view>

    <view class="waterfall">
      <view class="column left">
        <view v-for="(item, index) in leftList" :key="item.id" class="card" @tap="goDetail(item.id)">
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
        <view v-for="(item, index) in rightList" :key="item.id" class="card" @tap="goDetail(item.id)">
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
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getCheckinList, toggleLike } from '../../api'

const currentTab = ref(0)
const tabs = [
  { id: 0, name: '广场' },
  { id: 1, name: '考核组' },
]

const list = ref<any[]>([])

const leftList = computed(() => list.value.filter((_, i) => i % 2 === 0))
const rightList = computed(() => list.value.filter((_, i) => i % 2 === 1))

onMounted(async () => {
  try {
    const res: any = await getCheckinList(1, 20)
    list.value = res.list || []
  } catch (e) {}
})

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
}
.search-input {
  font-size: 28rpx;
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
</style>
