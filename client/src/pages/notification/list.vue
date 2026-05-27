<template>
  <view class="notif-page">
    <view class="notif-header">
      <text class="header-title">消息通知</text>
      <text class="mark-all" @tap="onMarkAllRead">全部已读</text>
    </view>

    <view v-if="notifications.length > 0">
      <view v-for="group in groupedNotifications" :key="group.label" class="notif-group">
        <text class="group-label">{{ group.label }}</text>
        <view v-for="item in group.items" :key="item.id" class="notif-item" :class="{ unread: !item.is_read }" @tap="onTapItem(item)">
          <view class="notif-avatar">
            <image v-if="item.sender?.avatar" class="avatar-img" :src="item.sender.avatar" />
            <view v-else class="avatar-icon" :class="'type-' + item.type">
              {{ typeIcon[item.type] }}
            </view>
          </view>
          <view class="notif-content">
            <text class="notif-title">{{ item.title }}</text>
            <text v-if="item.content" class="notif-text">{{ item.content }}</text>
            <text class="notif-time">{{ formatTime(item.created_at) }}</text>
          </view>
          <view v-if="!item.is_read" class="unread-dot"></view>
        </view>
      </view>
    </view>

    <view v-else-if="!loading" class="empty">
      <text class="empty-icon">🔔</text>
      <text class="empty-text">暂无通知</text>
    </view>

    <view v-if="loading" class="loading-tip">
      <text>加载中...</text>
    </view>

    <view v-else-if="noMore && notifications.length > 0" class="loading-tip">
      <text>没有更多了</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { getNotificationList, markNotificationRead, markAllNotificationsRead, deleteNotification } from '../../api'

const notifications = ref<any[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const pageSize = 20

const noMore = computed(() => notifications.value.length >= total.value)

const typeIcon: Record<string, string> = {
  like: '❤',
  comment: '💬',
  comment_reply: '↩',
  plan_remind: '📋',
  plan_complete: '🎉',
  group_notice: '👥',
  system: '📢',
}

const groupedNotifications = computed(() => {
  const groups: { label: string; items: any[] }[] = []
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const yesterday = new Date(today.getTime() - 86400000)

  let todayGroup: any[] = []
  let yesterdayGroup: any[] = []
  let earlierGroup: any[] = []

  for (const n of notifications.value) {
    const d = new Date(n.created_at)
    if (d >= today) {
      todayGroup.push(n)
    } else if (d >= yesterday) {
      yesterdayGroup.push(n)
    } else {
      earlierGroup.push(n)
    }
  }

  if (todayGroup.length > 0) groups.push({ label: '今天', items: todayGroup })
  if (yesterdayGroup.length > 0) groups.push({ label: '昨天', items: yesterdayGroup })
  if (earlierGroup.length > 0) groups.push({ label: '更早', items: earlierGroup })

  return groups
})

onMounted(() => {
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
    const res: any = await getNotificationList(1, pageSize)
    notifications.value = res.list || []
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
    const res: any = await getNotificationList(1, pageSize)
    notifications.value = res.list || []
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
    const res: any = await getNotificationList(nextPage, pageSize)
    notifications.value.push(...(res.list || []))
    total.value = res.total || 0
    page.value = nextPage
  } catch (e) {} finally {
    loading.value = false
  }
}

async function onTapItem(item: any) {
  if (!item.is_read) {
    try {
      await markNotificationRead(item.id)
      item.is_read = true
    } catch (e) {}
  }

  if (item.target_type === 'checkin' && item.target_id) {
    uni.navigateTo({ url: `/pages/video-detail/video-detail?id=${item.target_id}` })
  } else if (item.target_type === 'plan' && item.target_id) {
    if (item.type === 'plan_remind') {
      uni.navigateTo({ url: '/pages/training/today' })
    } else {
      uni.navigateTo({ url: `/pages/training/report?id=${item.target_id}` })
    }
  } else if (item.target_type === 'group' && item.target_id) {
    uni.navigateTo({ url: `/pages/group/detail?id=${item.target_id}` })
  }
}

async function onMarkAllRead() {
  try {
    await markAllNotificationsRead()
    for (const n of notifications.value) {
      n.is_read = true
    }
    uni.showToast({ title: '已全部标记已读', icon: 'success' })
  } catch (e) {}
}

function formatTime(ts: string) {
  if (!ts) return ''
  const d = new Date(ts)
  const now = new Date()
  const diff = now.getTime() - d.getTime()

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  if (diff < 172800000) return '昨天'
  return `${d.getMonth() + 1}-${d.getDate()}`
}
</script>

<style scoped>
.notif-page {
  padding: 20rpx;
}
.notif-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}
.header-title {
  font-size: 32rpx;
  font-weight: bold;
}
.mark-all {
  font-size: 26rpx;
  color: #1cbbb4;
}
.notif-group {
  margin-bottom: 20rpx;
}
.group-label {
  font-size: 24rpx;
  color: #999;
  margin-bottom: 12rpx;
  display: block;
}
.notif-item {
  display: flex;
  align-items: center;
  background: #fff;
  border-radius: 12rpx;
  padding: 20rpx;
  margin-bottom: 8rpx;
}
.notif-item.unread {
  background: #f0faf9;
}
.notif-avatar {
  margin-right: 16rpx;
}
.avatar-img {
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
}
.avatar-icon {
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
}
.type-like { background: #fce4ec; }
.type-comment { background: #e3f2fd; }
.type-comment_reply { background: #e3f2fd; }
.type-plan_remind { background: #fff3e0; }
.type-plan_complete { background: #e8f5e9; }
.type-group_notice { background: #f3e5f5; }
.type-system { background: #f5f5f5; }
.notif-content {
  flex: 1;
}
.notif-title {
  font-size: 28rpx;
  display: block;
}
.notif-text {
  font-size: 24rpx;
  color: #666;
  margin-top: 4rpx;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.notif-time {
  font-size: 22rpx;
  color: #999;
  margin-top: 4rpx;
  display: block;
}
.unread-dot {
  width: 16rpx;
  height: 16rpx;
  background: #e54d42;
  border-radius: 50%;
  margin-left: 12rpx;
}
.empty {
  text-align: center;
  padding: 100rpx;
}
.empty-icon {
  font-size: 80rpx;
  display: block;
  margin-bottom: 20rpx;
}
.empty-text {
  font-size: 30rpx;
  color: #999;
}
.loading-tip {
  text-align: center;
  padding: 30rpx;
  color: #999;
  font-size: 24rpx;
}
</style>
