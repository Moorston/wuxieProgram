<template>
  <view class="index-page">
    <view class="header">
      <image class="avatar" :src="userInfo?.avatar || '/static/default-avatar.png'" />
      <view class="user-info">
        <text class="nickname">{{ userInfo?.nickname || '未登录' }}</text>
        <text class="score">积分: {{ userInfo?.score || 0 }}</text>
      </view>
    </view>

    <view class="week-checkin">
      <text class="section-title">本周打卡</text>
      <view class="week-days">
        <view
          v-for="(day, index) in weekDays"
          :key="index"
          class="day-item"
          :class="{ active: day.checked }"
        >
          <text>{{ day.label }}</text>
        </view>
      </view>
    </view>

    <view class="recent-videos">
      <text class="section-title">最新打卡</text>
      <view class="video-list">
        <view v-for="item in recentList" :key="item.id" class="video-card" @tap="goDetail(item.id)">
          <image class="cover" :src="item.cover_url" mode="aspectFill" />
          <view class="info">
            <text class="desc">{{ item.description }}</text>
            <text class="author">{{ item.user?.nickname }}</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getProfile, getCheckinList } from '../../api'

const userInfo = ref<any>(null)
const recentList = ref<any[]>([])

const weekDays = ref([
  { label: '一', checked: false },
  { label: '二', checked: false },
  { label: '三', checked: false },
  { label: '四', checked: false },
  { label: '五', checked: false },
  { label: '六', checked: false },
  { label: '日', checked: false },
])

onMounted(async () => {
  try {
    userInfo.value = await getProfile()
  } catch (e) {}

  try {
    const res: any = await getCheckinList(1, 5)
    recentList.value = res.list || []
  } catch (e) {}
})

function goDetail(id: string) {
  uni.navigateTo({ url: `/pages/video-detail/video-detail?id=${id}` })
}
</script>

<style scoped>
.index-page {
  padding: 20rpx;
}
.header {
  display: flex;
  align-items: center;
  padding: 30rpx;
  background: linear-gradient(135deg, #1cbbb4, #0081ff);
  border-radius: 16rpx;
  color: #fff;
}
.avatar {
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  margin-right: 24rpx;
}
.nickname {
  font-size: 36rpx;
  font-weight: bold;
}
.score {
  font-size: 24rpx;
  opacity: 0.8;
  margin-top: 8rpx;
}
.section-title {
  font-size: 32rpx;
  font-weight: bold;
  margin: 30rpx 0 20rpx;
  display: block;
}
.week-days {
  display: flex;
  justify-content: space-around;
  background: #fff;
  border-radius: 16rpx;
  padding: 20rpx;
}
.day-item {
  width: 60rpx;
  height: 60rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24rpx;
  background: #f0f0f0;
}
.day-item.active {
  background: #1cbbb4;
  color: #fff;
}
.video-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}
.video-card {
  width: calc(50% - 8rpx);
  background: #fff;
  border-radius: 12rpx;
  overflow: hidden;
}
.cover {
  width: 100%;
  height: 240rpx;
}
.info {
  padding: 16rpx;
}
.desc {
  font-size: 26rpx;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.author {
  font-size: 22rpx;
  color: #999;
  margin-top: 8rpx;
  display: block;
}
</style>
