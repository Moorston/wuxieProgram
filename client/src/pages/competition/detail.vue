<template>
  <view class="detail-page">
    <view v-if="loading" class="empty-tip">加载中...</view>

    <view v-else-if="!comp" class="empty-tip">赛事不存在</view>

    <view v-else>
      <!-- 赛事信息 -->
      <view class="comp-header">
        <text class="comp-title">{{ comp.title }}</text>
        <text class="comp-status" :class="statusClass">{{ statusText }}</text>
      </view>
      <text class="comp-desc" v-if="comp.description">{{ comp.description }}</text>
      <view class="comp-meta">
        <text>📅 {{ formatDate(comp.start_date) }} - {{ formatDate(comp.end_date) }}</text>
        <text v-if="comp.rules">📋 {{ comp.rules }}</text>
      </view>

      <!-- 提交参赛 -->
      <view v-if="canSubmit" class="submit-section">
        <button class="btn-submit" @click="showSubmit = !showSubmit">📹 提交参赛作品</button>
        <view v-if="showSubmit" class="submit-form">
          <input v-model="checkinId" placeholder="输入打卡ID" class="submit-input" />
          <button class="btn-confirm" @click="doSubmit">确认提交</button>
        </view>
      </view>

      <!-- 排行榜 -->
      <view class="section-title">🏆 排行榜</view>
      <view v-if="ranking.length === 0" class="empty-tip-small">暂无排名数据</view>
      <view v-else class="ranking-list">
        <view class="ranking-item" v-for="(r, i) in ranking" :key="i">
          <text class="rank-num" :class="{ 'rank-top': i < 3 }">{{ r.rank }}</text>
          <text class="rank-name">{{ r.user?.nickname || '匿名' }}</text>
          <text class="rank-score">{{ r.score }} 分</text>
        </view>
      </view>

      <!-- 参赛作品 -->
      <view class="section-title">📹 参赛作品 ({{ entries.length }})</view>
      <view v-if="entries.length === 0" class="empty-tip-small">暂无参赛作品</view>
      <view v-else class="entry-list">
        <view class="entry-item" v-for="entry in entries" :key="entry.id">
          <text class="entry-user">{{ entry.user?.nickname || '匿名' }}</text>
          <text class="entry-score" v-if="entry.status === 1">{{ entry.score }} 分</text>
          <text class="entry-pending" v-else>待评分</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getCompetitionDetail, getCompetitionEntries, getCompetitionRanking, submitCompetitionEntry } from '../../api/competition'

const comp = ref<any>(null)
const entries = ref<any[]>([])
const ranking = ref<any[]>([])
const loading = ref(true)
const showSubmit = ref(false)
const checkinId = ref('')

const compId = ref('')

const canSubmit = computed(() => {
  if (!comp.value) return false
  const now = Date.now()
  const start = new Date(comp.value.start_date).getTime()
  const end = new Date(comp.value.end_date).getTime()
  return comp.value.status === 1 && now >= start && now <= end
})

const statusClass = computed(() => {
  if (!comp.value) return ''
  const now = Date.now()
  const start = new Date(comp.value.start_date).getTime()
  const end = new Date(comp.value.end_date).getTime()
  if (comp.value.status === 0) return 'status-draft'
  if (now >= start && now <= end) return 'status-active'
  return 'status-ended'
})

const statusText = computed(() => {
  if (!comp.value) return ''
  const now = Date.now()
  const start = new Date(comp.value.start_date).getTime()
  const end = new Date(comp.value.end_date).getTime()
  if (comp.value.status === 0) return '草稿'
  if (now < start) return '未开始'
  if (now >= start && now <= end) return '进行中'
  return '已结束'
})

function formatDate(date: string) {
  if (!date) return ''
  return date.slice(0, 10)
}

async function doSubmit() {
  if (!checkinId.value) {
    uni.showToast({ title: '请输入打卡ID', icon: 'none' })
    return
  }
  try {
    await submitCompetitionEntry(compId.value, checkinId.value)
    uni.showToast({ title: '提交成功', icon: 'success' })
    showSubmit.value = false
    checkinId.value = ''
    loadData()
  } catch (e: any) {
    uni.showToast({ title: e.message || '提交失败', icon: 'none' })
  }
}

async function loadData() {
  try {
    const [detail, entryRes, rankRes]: any = await Promise.all([
      getCompetitionDetail(compId.value),
      getCompetitionEntries(compId.value),
      getCompetitionRanking(compId.value),
    ])
    comp.value = detail
    entries.value = entryRes?.list || []
    ranking.value = rankRes?.ranking || []
  } catch (e) {
    console.error('load competition failed:', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  const pages = getCurrentPages()
  const page = pages[pages.length - 1]
  const id = (page as any).options?.id || ''
  compId.value = id
  loadData()
})
</script>

<style scoped>
.detail-page { padding: 24rpx; }
.comp-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12rpx; }
.comp-title { font-size: 36rpx; font-weight: 700; }
.comp-status { font-size: 22rpx; padding: 4rpx 12rpx; border-radius: 12rpx; }
.status-active { background: #dcfce7; color: #16a34a; }
.status-ended { background: #f1f5f9; color: #64748b; }
.status-draft { background: #fef3c7; color: #d97706; }
.comp-desc { font-size: 28rpx; color: #666; margin-bottom: 12rpx; }
.comp-meta { font-size: 24rpx; color: #999; margin-bottom: 24rpx; display: flex; flex-direction: column; gap: 8rpx; }
.section-title { font-size: 30rpx; font-weight: 600; margin: 24rpx 0 12rpx; }
.submit-section { margin: 24rpx 0; }
.btn-submit { background: #1cbbb4; color: #fff; border: none; border-radius: 12rpx; padding: 20rpx; font-size: 28rpx; width: 100%; }
.submit-form { margin-top: 16rpx; display: flex; gap: 12rpx; }
.submit-input { flex: 1; padding: 16rpx; border: 1px solid #e2e8f0; border-radius: 8rpx; font-size: 26rpx; }
.btn-confirm { background: #3b82f6; color: #fff; border: none; border-radius: 8rpx; padding: 16rpx 24rpx; font-size: 26rpx; }
.ranking-list { background: #fff; border-radius: 12rpx; overflow: hidden; }
.ranking-item { display: flex; align-items: center; padding: 16rpx 20rpx; border-bottom: 1px solid #f1f5f9; }
.rank-num { width: 48rpx; font-weight: 700; color: #64748b; font-size: 28rpx; }
.rank-top { color: #f59e0b; }
.rank-name { flex: 1; font-size: 28rpx; }
.rank-score { font-size: 28rpx; font-weight: 600; color: #1cbbb4; }
.entry-list { background: #fff; border-radius: 12rpx; overflow: hidden; }
.entry-item { display: flex; align-items: center; padding: 16rpx 20rpx; border-bottom: 1px solid #f1f5f9; }
.entry-user { flex: 1; font-size: 28rpx; }
.entry-score { font-size: 28rpx; font-weight: 600; color: #1cbbb4; }
.entry-pending { font-size: 24rpx; color: #999; }
.empty-tip { text-align: center; color: #999; padding: 60rpx; font-size: 28rpx; }
.empty-tip-small { text-align: center; color: #999; padding: 24rpx; font-size: 24rpx; }
</style>
