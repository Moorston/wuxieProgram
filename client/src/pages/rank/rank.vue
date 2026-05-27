<template>
  <view class="rank-page">
    <view class="period-tabs">
      <text
        v-for="p in periods"
        :key="p.value"
        class="period-tab"
        :class="{ active: currentPeriod === p.value }"
        @tap="changePeriod(p.value)"
      >{{ p.label }}</text>
    </view>

    <view class="rank-list">
      <view v-for="(item, index) in rankList" :key="item.user_id" class="rank-item">
        <view class="rank-num" :class="'rank-' + (index + 1)">
          {{ index < 3 ? ['🥇','🥈','🥉'][index] : index + 1 }}
        </view>
        <image class="rank-avatar" :src="item.user?.avatar || '/static/default-avatar.png'" />
        <view class="rank-info">
          <text class="rank-name">{{ item.user?.nickname || '未知' }}</text>
          <text class="rank-score">{{ item.score }} 积分</text>
        </view>
      </view>
    </view>

    <view v-if="loading" class="loading-tip">
      <text>加载中...</text>
    </view>
    <view v-else-if="noMore && rankList.length > 0" class="loading-tip">
      <text>没有更多了</text>
    </view>

    <view class="my-rank" v-if="myRank">
      <text>我的排名: 第{{ myRank.rank }}名</text>
      <text>积分: {{ myRank.score }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { getRankList, getMyRank } from '../../api'

const currentPeriod = ref('all')
const periods = [
  { value: 'day', label: '今日' },
  { value: 'week', label: '本周' },
  { value: 'all', label: '总榜' },
]

const rankList = ref<any[]>([])
const myRank = ref<any>(null)
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const pageSize = 20

const noMore = computed(() => rankList.value.length >= total.value)

onMounted(() => {
  loadRank()
})

onPullDownRefresh(async () => {
  await refreshRank()
  uni.stopPullDownRefresh()
})

onReachBottom(() => {
  loadMore()
})

async function loadRank() {
  if (loading.value) return
  loading.value = true
  try {
    const res: any = await getRankList(currentPeriod.value, 1, pageSize)
    rankList.value = res.list || res || []
    total.value = res.total || rankList.value.length
    page.value = 1
  } catch (e) {} finally {
    loading.value = false
  }
  try {
    myRank.value = await getMyRank(currentPeriod.value)
  } catch (e) {}
}

async function refreshRank() {
  page.value = 1
  loading.value = true
  try {
    const res: any = await getRankList(currentPeriod.value, 1, pageSize)
    rankList.value = res.list || res || []
    total.value = res.total || rankList.value.length
    page.value = 1
  } catch (e) {} finally {
    loading.value = false
  }
  try {
    myRank.value = await getMyRank(currentPeriod.value)
  } catch (e) {}
}

async function loadMore() {
  if (loading.value || noMore.value) return
  loading.value = true
  try {
    const nextPage = page.value + 1
    const res: any = await getRankList(currentPeriod.value, nextPage, pageSize)
    const items = res.list || res || []
    rankList.value.push(...items)
    total.value = res.total || rankList.value.length
    page.value = nextPage
  } catch (e) {} finally {
    loading.value = false
  }
}

function changePeriod(period: string) {
  currentPeriod.value = period
  loadRank()
}
</script>

<style scoped>
.rank-page {
  padding: 20rpx;
}
.period-tabs {
  display: flex;
  justify-content: center;
  gap: 40rpx;
  margin-bottom: 30rpx;
}
.period-tab {
  font-size: 28rpx;
  color: #999;
  padding: 12rpx 24rpx;
  border-radius: 24rpx;
}
.period-tab.active {
  background: #1cbbb4;
  color: #fff;
}
.rank-list {
  background: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}
.rank-item {
  display: flex;
  align-items: center;
  padding: 24rpx;
  border-bottom: 1rpx solid #f0f0f0;
}
.rank-num {
  width: 60rpx;
  font-size: 32rpx;
  text-align: center;
}
.rank-avatar {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  margin: 0 20rpx;
}
.rank-info {
  flex: 1;
}
.rank-name {
  font-size: 30rpx;
  display: block;
}
.rank-score {
  font-size: 24rpx;
  color: #999;
  margin-top: 4rpx;
  display: block;
}
.my-rank {
  margin-top: 30rpx;
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  display: flex;
  justify-content: space-between;
  font-size: 28rpx;
}
.loading-tip {
  text-align: center;
  padding: 30rpx;
  color: #999;
  font-size: 24rpx;
}
</style>
