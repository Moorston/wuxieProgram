<template>
  <view class="mine-page">
    <view class="profile-card" @tap="showEditModal = true">
      <image class="avatar" :src="userInfo?.avatar || '/static/default-avatar.png'" />
      <view class="profile-info">
        <text class="nickname">{{ userInfo?.nickname || '未登录' }}</text>
        <text class="stats">积分 {{ userInfo?.score || 0 }} · 打卡 {{ userInfo?.check_days || 0 }} 天</text>
      </view>
      <text class="edit-arrow">✎</text>
    </view>

    <view class="menu-list">
      <view class="menu-item" @tap="goTo('/pages/insight/list')">
        <text>感悟笔记</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/insight/on-this-day')">
        <text>历史今日</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/insight/public')">
        <text>公开感悟</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/training/list')">
        <text>训练计划</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/training/today')">
        <text>今日任务</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/training/template')">
        <text>训练模板</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/resource/list')">
        <text>个人资料库</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/resource/favorites')">
        <text>我的收藏</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/resource/stats')">
        <text>存储统计</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/my-video/my-video')">
        <text>我的视频</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/group/group')">
        <text>考核组</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/notification/settings')">
        <text>通知设置</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="onLogout">
        <text>退出登录</text>
        <text class="arrow">></text>
      </view>
    </view>

    <view v-if="showEditModal" class="edit-modal" @tap.self="showEditModal = false">
      <view class="modal-content">
        <text class="modal-title">编辑资料</text>

        <view class="form-item">
          <text class="label">头像</text>
          <view class="avatar-upload" @tap="changeAvatar">
            <image class="upload-avatar" :src="editForm.avatar || '/static/default-avatar.png'" />
            <text class="upload-hint">点击更换</text>
          </view>
        </view>

        <view class="form-item">
          <text class="label">昵称</text>
          <input v-model="editForm.nickname" class="input" placeholder="输入新昵称" maxlength="20" />
        </view>

        <view class="modal-btns">
          <text class="btn cancel" @tap="showEditModal = false">取消</text>
          <text class="btn confirm" @tap="onSaveProfile">保存</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { getProfile, updateProfile } from '../../api'
import { useUserStore } from '../../store/user'

const userStore = useUserStore()
const userInfo = ref<any>(null)
const showEditModal = ref(false)
const editForm = reactive({
  nickname: '',
  avatar: '',
})

onMounted(async () => {
  if (!userStore.isLoggedIn) {
    uni.navigateTo({ url: '/pages/login/login' })
    return
  }
  try {
    userInfo.value = await getProfile()
    editForm.nickname = userInfo.value?.nickname || ''
    editForm.avatar = userInfo.value?.avatar || ''
  } catch (e) {}
})

function goTo(url: string) {
  uni.navigateTo({ url })
}

function changeAvatar() {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    success: (res) => {
      editForm.avatar = res.tempFilePaths[0]
    },
  })
}

async function onSaveProfile() {
  if (!editForm.nickname.trim()) {
    uni.showToast({ title: '昵称不能为空', icon: 'none' })
    return
  }
  try {
    await updateProfile({
      nickname: editForm.nickname.trim(),
      avatar: editForm.avatar,
    })
    userInfo.value = await getProfile()
    showEditModal.value = false
    uni.showToast({ title: '保存成功', icon: 'success' })
  } catch (e) {
    uni.showToast({ title: '保存失败', icon: 'none' })
  }
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
.profile-info {
  flex: 1;
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
.edit-arrow {
  font-size: 36rpx;
  opacity: 0.8;
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
.edit-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}
.modal-content {
  background: #fff;
  border-radius: 16rpx;
  padding: 32rpx;
  width: 650rpx;
}
.modal-title {
  font-size: 32rpx;
  font-weight: bold;
  display: block;
  margin-bottom: 24rpx;
}
.form-item {
  margin-bottom: 20rpx;
}
.label {
  font-size: 26rpx;
  color: #666;
  margin-bottom: 12rpx;
  display: block;
}
.avatar-upload {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.upload-avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
}
.upload-hint {
  font-size: 22rpx;
  color: #1cbbb4;
  margin-top: 8rpx;
}
.input {
  border: 1rpx solid #e0e0e0;
  border-radius: 12rpx;
  padding: 16rpx;
  font-size: 28rpx;
}
.modal-btns {
  display: flex;
  gap: 16rpx;
  margin-top: 24rpx;
}
.btn {
  flex: 1;
  text-align: center;
  padding: 16rpx;
  border-radius: 12rpx;
  font-size: 28rpx;
}
.btn.cancel {
  background: #f0f0f0;
  color: #666;
}
.btn.confirm {
  background: #1cbbb4;
  color: #fff;
}
</style>
