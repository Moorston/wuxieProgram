<template>
  <view class="plan-list-page">
    <view class="tabs">
      <text
        v-for="tab in tabs"
        :key="tab.value"
        class="tab-item"
        :class="{ active: currentTab === tab.value }"
        @tap="onTabChange(tab.value)"
      >{{ tab.label }}</text>
    </view>

    <view class="plan-list">
      <view v-for="plan in plans" :key="plan.id" class="plan-card" @tap="goDetail(plan.id)">
        <view class="plan-header">
          <text class="plan-title">{{ plan.title }}</text>
          <text class="plan-status" :class="'status-' + plan.status">
            {{ getStatusLabel(plan.status) }}
          </text>
        </view>
        <text class="plan-desc">{{ plan.description || '暂无描述' }}</text>
        <view class="plan-info">
          <text>{{ plan.start_date?.slice(0, 10) }} ~ {{ plan.end_date?.slice(0, 10) }}</text>
        </view>
        <view class="plan-progress">
          <view class="progress-bar">
            <view class="progress-fill" :style="{ width: plan.stats?.completion_rate + '%' }"></view>
          </view>
          <text class="progress-text">{{ plan.stats?.completed || 0 }}/{{ plan.stats?.total_tasks || 0 }} ({{ (plan.stats?.completion_rate || 0).toFixed(0) }}%)</text>
        </view>
      </view>
    </view>

    <view v-if="loading" class="loading-tip">
      <text>加载中...</text>
    </view>
    <view v-else-if="plans.length === 0" class="empty">
      <text>暂无训练计划</text>
    </view>

    <view class="fab" @tap="goCreate">
      <text class="fab-icon">+</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { listTrainingPlans } from '../../api'

const currentTab = ref('')
const tabs = [
  { value: '', label: '全部' },
  { value: '1', label: '进行中' },
  { value: '2', label: '已完成' },
  { value: '0', label: '草稿' },
]

const statusMap: Record<number, string> = {
  0: '草稿',
  1: '进行中',
  2: '已完成',
  3: '已终止',
}

const getStatusLabel = (status: number) => statusMap[status] || '未知状态'

const plans = ref<any[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const pageSize = 10

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

async function loadData(resetPage = true) {
  if (loading.value) return
  loading.value = true
  try {
    const currentPage = resetPage ? 1 : page.value
    const res: any = await listTrainingPlans(currentPage, pageSize, currentTab.value || undefined)
    if (resetPage) {
      plans.value = res.list || []
    } else {
      plans.value.push(...(res.list || []))
    }
    total.value = res.total || 0
    if (resetPage) {
      page.value = 1
    }
  } catch (e) {
    uni.showToast({ title: resetPage ? '加载失败，请重试' : '加载更多失败', icon: 'none' })
    console.error('Failed to load training plans:', e)
  } finally {
    loading.value = false
  }
}

async function refreshData() {
  await loadData(true)
}

async function loadMore() {
  if (loading.value || plans.value.length >= total.value) return
  const nextPage = page.value + 1
  loading.value = true
  try {
    const res: any = await listTrainingPlans(nextPage, pageSize, currentTab.value || undefined)
    plans.value.push(...(res.list || []))
    total.value = res.total || 0
    page.value = nextPage
  } catch (e) {
    uni.showToast({ title: '加载更多失败', icon: 'none' })
    console.error('Failed to load more training plans:', e)
  } finally {
    loading.value = false
  }
}

function onTabChange(value: string) {
  currentTab.value = value
  loadData()
}

function goDetail(id: string) {
  uni.navigateTo({ url: `/pages/training/detail?id=${id}` })
}

function goCreate() {
  uni.navigateTo({ url: '/pages/training/create' })
}
</script>

<style scoped>
.plan-list-page {
  padding: 20rpx;
  padding-bottom: 120rpx;
}
.tabs {
  display: flex;
  gap: 20rpx;
  margin-bottom: 20rpx;
  background: #fff;
  border-radius: 16rpx;
  padding: 16rpx;
}
.tab-item {
  flex: 1;
  text-align: center;
  font-size: 26rpx;
  color: #999;
  padding: 12rpx 0;
  border-radius: 24rpx;
}
.tab-item.active {
  background: #1cbbb4;
  color: #fff;
}
.plan-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}
.plan-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.plan-title {
  font-size: 32rpx;
  font-weight: bold;
  flex: 1;
}
.plan-status {
  font-size: 22rpx;
  padding: 4rpx 16rpx;
  border-radius: 20rpx;
}
.status-0 { background: #f0f0f0; color: #999; }
.status-1 { background: #e8f5e9; color: #4caf50; }
.status-2 { background: #e3f2fd; color: #2196f3; }
.status-3 { background: #fce4ec; color: #e91e63; }
.plan-desc {
  font-size: 26rpx;
  color: #666;
  margin-top: 12rpx;
  display: block;
}
.plan-info {
  font-size: 24rpx;
  color: #999;
  margin-top: 8rpx;
}
.plan-progress {
  margin-top: 16rpx;
}
.progress-bar {
  height: 8rpx;
  background: #f0f0f0;
  border-radius: 4rpx;
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  background: #1cbbb4;
  border-radius: 4rpx;
  transition: width 0.3s;
}
.progress-text {
  font-size: 22rpx;
  color: #999;
  margin-top: 8rpx;
  display: block;
}
.loading-tip {
  text-align: center;
  padding: 30rpx;
  color: #999;
  font-size: 24rpx;
}
.empty {
  text-align: center;
  padding: 100rpx;
  color: #999;
}
.fab {
  position: fixed;
  right: 40rpx;
  bottom: 100rpx;
  width: 100rpx;
  height: 100rpx;
  background: #1cbbb4;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 16rpx rgba(28, 187, 180, 0.4);
}
.fab-icon {
  font-size: 48rpx;
  color: #fff;
}
</style>
