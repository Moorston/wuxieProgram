<template>
  <view class="page-status">
    <!-- 加载状态 -->
    <view v-if="loading && !loaded" class="status-loading">
      <view class="loading-spinner" />
      <text class="status-text">{{ loadingText }}</text>
    </view>

    <!-- 空状态 -->
    <view v-else-if="!loading && empty" class="status-empty">
      <text class="status-icon">{{ emptyIcon }}</text>
      <text class="status-text">{{ emptyText }}</text>
      <button v-if="showRetry" class="retry-btn" @tap="$emit('retry')">重试</button>
    </view>

    <!-- 错误状态 -->
    <view v-else-if="error" class="status-error">
      <text class="status-icon">⚠️</text>
      <text class="status-text">{{ errorText }}</text>
      <button class="retry-btn" @tap="$emit('retry')">重试</button>
    </view>

    <!-- 默认插槽 -->
    <slot v-else />
  </view>
</template>

<script setup lang="ts">
defineProps({
  loading: { type: Boolean, default: false },
  loaded: { type: Boolean, default: false },
  empty: { type: Boolean, default: false },
  error: { type: Boolean, default: false },
  loadingText: { type: String, default: '加载中...' },
  emptyText: { type: String, default: '暂无数据' },
  emptyIcon: { type: String, default: '📭' },
  errorText: { type: String, default: '加载失败' },
  showRetry: { type: Boolean, default: true },
})

defineEmits(['retry'])
</script>

<style scoped>
.page-status {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 300rpx;
  padding: 40rpx;
}
.status-loading, .status-empty, .status-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16rpx;
}
.status-icon { font-size: 64rpx; }
.status-text { font-size: 28rpx; color: #999; }
.loading-spinner {
  width: 48rpx; height: 48rpx;
  border: 4rpx solid #e0e0e0;
  border-top-color: #1cbbb4;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.retry-btn {
  margin-top: 16rpx;
  padding: 12rpx 40rpx;
  background: #1cbbb4;
  color: #fff;
  border-radius: 8rpx;
  font-size: 26rpx;
  border: none;
}
</style>