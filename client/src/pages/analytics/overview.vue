<template>
  <view class="analytics-page">
    <view class="section-title">📊 数据概览</view>

    <!-- 概览卡片 -->
    <view class="stats-grid" v-if="overview">
      <view class="stat-card">
        <text class="stat-value">{{ overview.total_check_days }}</text>
        <text class="stat-label">累计打卡</text>
      </view>
      <view class="stat-card">
        <text class="stat-value">{{ overview.streak_days }}</text>
        <text class="stat-label">连续打卡</text>
      </view>
      <view class="stat-card">
        <text class="stat-value">{{ overview.week_checkins }}</text>
        <text class="stat-label">本周打卡</text>
      </view>
      <view class="stat-card">
        <text class="stat-value">{{ overview.total_score }}</text>
        <text class="stat-label">总积分</text>
      </view>
    </view>

    <!-- 热力图 -->
    <view class="section-title">🔥 打卡热力图</view>
    <view class="heatmap-container" v-if="heatmapDates.length > 0">
      <view class="heatmap-row" v-for="week in heatmapWeeks" :key="week[0]">
        <view
          class="heatmap-cell"
          v-for="day in week"
          :key="day.date"
          :class="getCellClass(day.count)"
          :title="day.date + ': ' + day.count + '次'"
        />
      </view>
    </view>
    <view v-else class="empty-tip">暂无打卡数据</view>

    <!-- 趋势 -->
    <view class="section-title">📈 打卡趋势（近30天）</view>
    <view class="trend-container" v-if="trend.length > 0">
      <view class="trend-bar" v-for="point in trend" :key="point.date">
        <view class="trend-fill" :style="{ height: getBarHeight(point.count) + 'px' }" />
        <text class="trend-label">{{ point.date.slice(5) }}</text>
      </view>
    </view>
    <view v-else class="empty-tip">暂无趋势数据</view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { getCheckinHeatmap, getCheckinTrend, getAnalyticsOverview } from '../api/analytics'

const overview = ref<any>(null)
const heatmap = ref<Record<string, number>>({})
const trend = ref<any[]>([])

const heatmapDates = computed(() => Object.keys(heatmap.value).sort())

const heatmapWeeks = computed(() => {
  const dates = heatmapDates.value
  if (dates.length === 0) return []

  const weeks: any[][] = []
  let currentWeek: any[] = []

  // 填充第一周的空白天
  const firstDate = new Date(dates[0])
  const firstDay = firstDate.getDay()
  for (let i = 0; i < firstDay; i++) {
    currentWeek.push({ date: '', count: 0 })
  }

  for (const date of dates) {
    currentWeek.push({ date, count: heatmap.value[date] || 0 })
    if (currentWeek.length === 7) {
      weeks.push(currentWeek)
      currentWeek = []
    }
  }

  if (currentWeek.length > 0) {
    weeks.push(currentWeek)
  }

  return weeks
})

function getCellClass(count: number) {
  if (count === 0) return 'heatmap-0'
  if (count === 1) return 'heatmap-1'
  if (count <= 3) return 'heatmap-2'
  return 'heatmap-3'
}

function getBarHeight(count: number) {
  const max = Math.max(...trend.value.map(t => t.count), 1)
  return Math.max(4, (count / max) * 80)
}

onMounted(async () => {
  try {
    const [overviewData, heatmapData, trendData] = await Promise.all([
      getAnalyticsOverview(),
      getCheckinHeatmap(6),
      getCheckinTrend(30),
    ])
    overview.value = overviewData
    heatmap.value = heatmapData as any || {}
    trend.value = trendData as any || []
  } catch (e) {
    console.error('load analytics failed:', e)
  }
})
</script>

<style scoped>
.analytics-page { padding: 24rpx; }
.section-title { font-size: 32rpx; font-weight: 600; margin: 32rpx 0 16rpx; }
.stats-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16rpx; }
.stat-card { background: #fff; border-radius: 16rpx; padding: 24rpx; text-align: center; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.06); }
.stat-value { font-size: 48rpx; font-weight: 700; color: #1cbbb4; display: block; }
.stat-label { font-size: 24rpx; color: #999; margin-top: 8rpx; display: block; }
.heatmap-container { display: flex; flex-wrap: wrap; gap: 4rpx; padding: 16rpx; background: #fff; border-radius: 16rpx; }
.heatmap-row { display: flex; gap: 4rpx; }
.heatmap-cell { width: 24rpx; height: 24rpx; border-radius: 4rpx; }
.heatmap-0 { background: #ebedf0; }
.heatmap-1 { background: #9be9a8; }
.heatmap-2 { background: #40c463; }
.heatmap-3 { background: #30a14e; }
.trend-container { display: flex; align-items: flex-end; gap: 4rpx; height: 120rpx; padding: 16rpx; background: #fff; border-radius: 16rpx; overflow-x: auto; }
.trend-bar { display: flex; flex-direction: column; align-items: center; min-width: 16rpx; }
.trend-fill { width: 12rpx; background: #1cbbb4; border-radius: 4rpx 4rpx 0 0; min-height: 4rpx; }
.trend-label { font-size: 16rpx; color: #999; margin-top: 4rpx; writing-mode: vertical-lr; }
.empty-tip { text-align: center; color: #999; padding: 40rpx; font-size: 28rpx; }
</style>
