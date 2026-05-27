<template>
  <view class="detail-page" v-if="group">
    <view class="group-header">
      <text class="group-name">{{ group.name }}</text>
      <text class="group-count">{{ group.members?.length || 0 }} 人</text>
    </view>
    <text class="group-desc">{{ group.description }}</text>

    <view class="leader-section" v-if="group.leader">
      <text class="section-title">组长</text>
      <view class="member-item">
        <image class="member-avatar" :src="group.leader.avatar || '/static/default-avatar.png'" />
        <view class="member-info">
          <text class="member-name">{{ group.leader.nickname }}</text>
          <text class="member-role">组长</text>
        </view>
      </view>
    </view>

    <view class="members-section">
      <text class="section-title">成员列表 ({{ group.members?.length || 0 }})</text>
      <view v-for="m in group.members" :key="m.id" class="member-item">
        <image class="member-avatar" :src="m.avatar || '/static/default-avatar.png'" />
        <view class="member-info">
          <text class="member-name">{{ m.nickname }}</text>
          <text class="member-score">积分 {{ m.score }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getGroupDetail } from '../../api'

const groupId = ref('')
const group = ref<any>(null)

onLoad((options) => {
  groupId.value = options?.id || ''
  loadGroup()
})

async function loadGroup() {
  try {
    group.value = await getGroupDetail(groupId.value)
  } catch (e) {}
}
</script>

<style scoped>
.detail-page {
  padding: 20rpx;
}
.group-header {
  background: linear-gradient(135deg, #1cbbb4, #0081ff);
  border-radius: 16rpx;
  padding: 40rpx;
  color: #fff;
}
.group-name {
  font-size: 36rpx;
  font-weight: bold;
  display: block;
}
.group-count {
  font-size: 26rpx;
  opacity: 0.9;
  margin-top: 8rpx;
  display: block;
}
.group-desc {
  font-size: 28rpx;
  color: #666;
  margin-top: 16rpx;
  display: block;
}
.leader-section, .members-section {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-top: 16rpx;
}
.section-title {
  font-size: 30rpx;
  font-weight: bold;
  margin-bottom: 16rpx;
  display: block;
}
.member-item {
  display: flex;
  align-items: center;
  padding: 16rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}
.member-item:last-child {
  border-bottom: none;
}
.member-avatar {
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  margin-right: 16rpx;
}
.member-info {
  flex: 1;
}
.member-name {
  font-size: 28rpx;
  display: block;
}
.member-role {
  font-size: 22rpx;
  color: #1cbbb4;
  margin-top: 4rpx;
  display: block;
}
.member-score {
  font-size: 22rpx;
  color: #999;
  margin-top: 4rpx;
  display: block;
}
</style>
