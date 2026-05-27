<template>
  <view class="group-page">
    <view v-for="g in groups" :key="g.id" class="group-card" @tap="goDetail(g.id)">
      <view class="group-header">
        <text class="group-name">{{ g.name }}</text>
        <text class="group-count">{{ g.members?.length || 0 }} 人</text>
      </view>
      <text class="group-desc">{{ g.description }}</text>
      <view class="member-avatars">
        <image
          v-for="(m, i) in (g.members || []).slice(0, 5)"
          :key="m.id"
          class="member-avatar"
          :src="m.avatar"
          :style="{ marginLeft: i > 0 ? '-16rpx' : '0' }"
        />
        <text v-if="(g.members || []).length > 5" class="more">+{{ g.members.length - 5 }}</text>
      </view>
    </view>

    <view v-if="groups.length === 0" class="empty">
      <text>暂无考核组</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getGroupList } from '../../api'

const groups = ref<any[]>([])

onMounted(async () => {
  try {
    groups.value = (await getGroupList()) as any[]
  } catch (e) {}
})

function goDetail(id: string) {
  uni.navigateTo({ url: `/pages/group/detail?id=${id}` })
}
</script>

<style scoped>
.group-page {
  padding: 20rpx;
}
.group-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}
.group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.group-name {
  font-size: 32rpx;
  font-weight: bold;
}
.group-count {
  font-size: 24rpx;
  color: #999;
}
.group-desc {
  font-size: 26rpx;
  color: #666;
  margin-top: 12rpx;
  display: block;
}
.member-avatars {
  display: flex;
  align-items: center;
  margin-top: 16rpx;
}
.member-avatar {
  width: 48rpx;
  height: 48rpx;
  border-radius: 50%;
  border: 2rpx solid #fff;
}
.more {
  font-size: 22rpx;
  color: #999;
  margin-left: 8rpx;
}
.empty {
  text-align: center;
  padding: 100rpx;
  color: #999;
}
</style>
