<template>
  <view class="tags-page">
    <view v-if="tags.length > 0" class="tag-list">
      <view v-for="tag in tags" :key="tag.id" class="tag-card" @tap="goTagInsights(tag.tag)">
        <text class="tag-name">#{{ tag.tag }}</text>
        <text class="tag-count">{{ tag.count }}篇</text>
      </view>
    </view>
    <view v-else class="empty">
      <text>暂无标签</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getInsightTags } from '../../api'

const tags = ref<any[]>([])

onMounted(async () => {
  try {
    const res: any = await getInsightTags()
    tags.value = res || []
  } catch (e) {}
})

function goTagInsights(tag: string) {
  uni.navigateTo({ url: `/pages/insight/list?tag=${encodeURIComponent(tag)}` })
}
</script>

<style scoped>
.tags-page {
  padding: 20rpx;
}
.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}
.tag-card {
  background: #fff;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.tag-name {
  font-size: 28rpx;
  color: #1cbbb4;
}
.tag-count {
  font-size: 22rpx;
  color: #999;
}
.empty {
  text-align: center;
  padding: 100rpx;
  color: #999;
}
</style>
