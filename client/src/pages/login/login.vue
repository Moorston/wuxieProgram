<template>
  <view class="login-page">
    <view class="logo-area">
      <image class="logo" src="/static/logo.png" />
      <text class="app-name">武协打卡</text>
    </view>

    <button class="login-btn" @tap="onLogin">微信登录</button>
  </view>
</template>

<script setup lang="ts">
import { useUserStore } from '../../store/user'
import { wxLogin } from '../../api'

const userStore = useUserStore()

async function onLogin() {
  try {
    // 获取微信code
    const loginRes: any = await new Promise((resolve, reject) => {
      uni.login({ success: resolve, fail: reject })
    })

    // 获取用户信息
    let userInfo: any = null
    try {
      const profileRes: any = await new Promise((resolve, reject) => {
        uni.getUserProfile({ desc: '用于完善用户资料', success: resolve, fail: reject })
      })
      userInfo = profileRes.userInfo
    } catch (e) {}

    // 调用后端登录
    const res: any = await wxLogin(loginRes.code, userInfo)

    userStore.setToken(res.token)
    userStore.setUser(res.user)

    uni.switchTab({ url: '/pages/index/index' })
  } catch (e) {
    uni.showToast({ title: '登录失败', icon: 'none' })
  }
}
</script>

<style scoped>
.login-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  padding: 60rpx;
}
.logo-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 100rpx;
}
.logo {
  width: 160rpx;
  height: 160rpx;
}
.app-name {
  font-size: 40rpx;
  font-weight: bold;
  margin-top: 24rpx;
  color: #1cbbb4;
}
.login-btn {
  width: 80%;
  background: #1cbbb4;
  color: #fff;
  border-radius: 40rpx;
  height: 88rpx;
  line-height: 88rpx;
  font-size: 32rpx;
}
</style>
