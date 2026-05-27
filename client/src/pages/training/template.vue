<template>
  <view class="template-page">
    <view class="filter-bar">
      <picker :value="styleIndex" :range="styleOptions" @change="onStyleChange">
        <text class="filter-btn">{{ currentStyle || '全部拳种' }}</text>
      </picker>
      <picker :value="categoryIndex" :range="categoryOptions" @change="onCategoryChange">
        <text class="filter-btn">{{ currentCategory || '全部难度' }}</text>
      </picker>
    </view>

    <view class="template-list">
      <view v-for="t in templates" :key="t.id" class="template-card" @tap="goDetail(t.id)">
        <view class="template-header">
          <text class="template-name">{{ t.name }}</text>
          <text class="template-style">{{ t.style }}</text>
        </view>
        <text class="template-desc">{{ t.description }}</text>
        <view class="template-meta">
          <text>{{ t.duration_days }}天</text>
          <text>{{ t.category }}</text>
          <text>{{ t.usage_count }}人使用</text>
        </view>
      </view>
    </view>

    <view v-if="loading" class="loading-tip">
      <text>加载中...</text>
    </view>
    <view v-else-if="templates.length === 0" class="empty">
      <text>暂无训练模板</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onPullDownRefresh } from '@dcloudio/uni-app'
import { listTemplates } from '../../api'

const styleOptions = ['全部', '长拳', '太极', '南拳', '散打', '基本功']
const categoryOptions = ['全部', '初级', '中级', '高级']
const styleIndex = ref(0)
const categoryIndex = ref(0)
const currentStyle = ref('')
const currentCategory = ref('')

const templates = ref<any[]>([])
const loading = ref(false)

onMounted(() => {
  loadData()
})

onPullDownRefresh(async () => {
  await loadData()
  uni.stopPullDownRefresh()
})

async function loadData() {
  loading.value = true
  try {
    const style = styleIndex.value > 0 ? styleOptions[styleIndex.value] : ''
    const category = categoryIndex.value > 0 ? categoryOptions[categoryIndex.value] : ''
    const res: any = await listTemplates(1, 50, category, style)
    templates.value = res.list || []
  } catch (e) {} finally {
    loading.value = false
  }
}

function onStyleChange(e: any) {
  styleIndex.value = e.detail.value
  currentStyle.value = styleIndex.value > 0 ? styleOptions[styleIndex.value] : ''
  loadData()
}

function onCategoryChange(e: any) {
  categoryIndex.value = e.detail.value
  currentCategory.value = categoryIndex.value > 0 ? categoryOptions[categoryIndex.value] : ''
  loadData()
}

function goDetail(id: string) {
  uni.navigateTo({ url: `/pages/training/template-detail?id=${id}` })
}
</script>

<style scoped>
.template-page {
  padding: 20rpx;
}
.filter-bar {
  display: flex;
  gap: 16rpx;
  margin-bottom: 20rpx;
}
.filter-btn {
  background: #fff;
  border-radius: 24rpx;
  padding: 12rpx 24rpx;
  font-size: 26rpx;
  color: #666;
}
.template-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.template-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.template-name {
  font-size: 32rpx;
  font-weight: bold;
  flex: 1;
}
.template-style {
  font-size: 22rpx;
  background: #e3f2fd;
  color: #2196f3;
  padding: 4rpx 16rpx;
  border-radius: 20rpx;
}
.template-desc {
  font-size: 26rpx;
  color: #666;
  margin-top: 12rpx;
  display: block;
}
.template-meta {
  display: flex;
  gap: 24rpx;
  margin-top: 12rpx;
  font-size: 22rpx;
  color: #999;
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
</style>
