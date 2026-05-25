<template>
  <view class="mine-page">
    <view class="profile-card">
      <image class="avatar" :src="userInfo?.avatar || '/static/default-avatar.png'" />
      <view class="profile-info">
        <text class="nickname">{{ userInfo?.nickname || '未登录' }}</text>
        <text class="stats">积分 {{ userInfo?.score || 0 }} · 打卡 {{ userInfo?.check_days || 0 }} 天</text>
      </view>
    </view>

    <view class="menu-list">
      <view class="menu-item" @tap="goTo('/pages/my-video/my-video')">
        <text>我的视频</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/group/group')">
        <text>考核组</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="onLogout">
        <text>退出登录</text>
        <text class="arrow">></text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getProfile } from '../../api'
import { useUserStore } from '../../store/user'

const userStore = useUserStore()
const userInfo = ref<any>(null)

onMounted(async () => {
  if (!userStore.isLoggedIn) {
    uni.navigateTo({ url: '/pages/login/login' })
    return
  }
  try {
    userInfo.value = await getProfile()
  } catch (e) {}
})

function goTo(url: string) {
  uni.navigateTo({ url })
}

function onLogout() {
  userStore.logout()
  uni.navigateTo({ url: '/pages/login/login' })
}
</script>

<style scoped>
.mine-page {
  padding: 20rpx;
}
.profile-card {
  display: flex;
  align-items: center;
  padding: 40rpx 30rpx;
  background: linear-gradient(135deg, #1cbbb4, #0081ff);
  border-radius: 16rpx;
  color: #fff;
}
.avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  margin-right: 30rpx;
}
.nickname {
  font-size: 36rpx;
  font-weight: bold;
  display: block;
}
.stats {
  font-size: 24rpx;
  opacity: 0.8;
  margin-top: 8rpx;
  display: block;
}
.menu-list {
  margin-top: 30rpx;
  background: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}
.menu-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 30rpx;
  border-bottom: 1rpx solid #f0f0f0;
  font-size: 30rpx;
}
.arrow {
  color: #ccc;
}
</style>
