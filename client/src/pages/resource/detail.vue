<template>
  <view class="detail-page" v-if="resource">
    <view class="media-area" v-if="resource.type === 'video'">
      <video :src="videoURL" class="video-player" controls />
    </view>
    <view class="media-area" v-else-if="resource.type === 'image'">
      <image class="image-view" :src="resource.file_url" mode="widthFix" @tap="previewImage" />
    </view>
    <view class="media-area doc" v-else>
      <text class="doc-icon">📄</text>
      <text class="doc-name">{{ resource.title }}</text>
    </view>

    <view class="info-section">
      <text class="title">{{ resource.title }}</text>
      <text class="desc">{{ resource.description || '暂无描述' }}</text>
      <view class="meta-row">
        <text class="meta-item">{{ typeLabel[resource.type] }}</text>
        <text class="meta-item">{{ resource.category }}</text>
        <text class="meta-item">{{ resource.difficulty }}</text>
        <text class="meta-item">{{ formatSize(resource.file_size) }}</text>
      </view>
      <view class="meta-row">
        <text class="meta-item">👁 {{ resource.view_count }}</text>
        <text class="meta-item">⬇ {{ resource.download_count }}</text>
        <text class="meta-item">{{ resource.created_at?.slice(0, 10) }}</text>
      </view>
      <view v-if="resource.tags && resource.tags.length > 0" class="tags-row">
        <text v-for="tag in resource.tags" :key="tag" class="tag">{{ tag }}</text>
      </view>
    </view>

    <view class="actions">
      <text class="action-btn" :class="{ active: resource.is_favorite }" @tap="onFavorite">
        {{ resource.is_favorite ? '★ 已收藏' : '☆ 收藏' }}
      </text>
      <text class="action-btn" @tap="onShare">分享到组</text>
      <text class="action-btn delete" @tap="onDelete">删除</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getResource, toggleResourceFavorite, deleteResource } from '../../api'
import { mediaRequest } from '../../utils/request'

const resourceId = ref('')
const resource = ref<any>(null)
const videoURL = ref('')

const typeLabel: Record<string, string> = { video: '视频', image: '图片', document: '文档' }

onLoad((options) => {
  resourceId.value = options?.id || ''
  loadResource()
})

async function loadResource() {
  try {
    resource.value = await getResource(resourceId.value)
    if (resource.value?.type === 'video' && resource.value?.file_url) {
      const urlRes: any = await mediaRequest({ url: `/media/url?object=${encodeURIComponent(resource.value.file_url)}&bucket=resource` })
      videoURL.value = urlRes.url
    }
  } catch (e) {}
}

function previewImage() {
  uni.previewImage({ urls: [resource.value.file_url] })
}

async function onFavorite() {
  try {
    const res: any = await toggleResourceFavorite(resourceId.value)
    resource.value.is_favorite = res.is_favorite
    uni.showToast({ title: res.is_favorite ? '已收藏' : '已取消收藏', icon: 'success' })
  } catch (e) {}
}

function onShare() {
  uni.navigateTo({ url: `/pages/resource/share?id=${resourceId.value}` })
}

async function onDelete() {
  uni.showModal({
    title: '确认删除',
    content: '确定要删除此资料吗？',
    success: async (res) => {
      if (res.confirm) {
        try {
          await deleteResource(resourceId.value)
          uni.showToast({ title: '已删除', icon: 'success' })
          setTimeout(() => uni.navigateBack(), 1500)
        } catch (e) {}
      }
    },
  })
}

function formatSize(bytes: number) {
  if (!bytes) return '0B'
  if (bytes < 1024) return bytes + 'B'
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + 'KB'
  if (bytes < 1073741824) return (bytes / 1048576).toFixed(1) + 'MB'
  return (bytes / 1073741824).toFixed(1) + 'GB'
}
</script>

<style scoped>
.detail-page { padding: 20rpx; }
.media-area { background: #000; border-radius: 16rpx; overflow: hidden; margin-bottom: 16rpx; }
.video-player { width: 100%; height: 420rpx; }
.image-view { width: 100%; }
.media-area.doc { height: 300rpx; display: flex; flex-direction: column; align-items: center; justify-content: center; background: #fff; }
.doc-icon { font-size: 80rpx; }
.doc-name { font-size: 28rpx; color: #666; margin-top: 12rpx; }
.info-section { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; }
.title { font-size: 32rpx; font-weight: bold; display: block; }
.desc { font-size: 26rpx; color: #666; margin-top: 8rpx; display: block; }
.meta-row { display: flex; gap: 16rpx; margin-top: 12rpx; flex-wrap: wrap; }
.meta-item { font-size: 24rpx; color: #999; }
.tags-row { display: flex; gap: 8rpx; margin-top: 12rpx; flex-wrap: wrap; }
.tag { background: #f0f0f0; padding: 4rpx 12rpx; border-radius: 12rpx; font-size: 22rpx; color: #666; }
.actions { display: flex; gap: 12rpx; }
.action-btn { flex: 1; text-align: center; padding: 18rpx; background: #fff; border-radius: 12rpx; font-size: 26rpx; }
.action-btn.active { color: #ff9800; }
.action-btn.delete { color: #e54d42; }
</style>
