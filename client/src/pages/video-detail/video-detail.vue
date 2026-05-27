<template>
  <view class="detail-page">
    <video :src="videoURL" class="video-player" controls autoplay />

    <view class="info">
      <view class="author-row">
        <image class="avatar" :src="checkin?.user?.avatar" />
        <text class="author-name">{{ checkin?.user?.nickname }}</text>
      </view>
      <text class="description">{{ checkin?.description }}</text>
      <view class="stats">
        <text>{{ checkin?.like_count }} 赞</text>
        <text>{{ checkin?.comment_count }} 评论</text>
      </view>
    </view>

    <view class="actions">
      <button class="action-btn" @tap="onLike">
        {{ checkin?.is_liked ? '已赞' : '点赞' }}
      </button>
    </view>

    <view class="comments-section">
      <text class="section-title">评论</text>
      <view v-for="c in comments" :key="c.id" class="comment-item">
        <image class="comment-avatar" :src="c.user?.avatar" />
        <view class="comment-content">
          <text class="comment-name">{{ c.user?.nickname }}</text>
          <text class="comment-text">{{ c.content }}</text>
        </view>
      </view>

      <view class="comment-input">
        <input v-model="newComment" placeholder="写评论..." />
        <button class="send-btn" @tap="onComment">发送</button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getCheckinDetail, toggleLike, addComment, getComments } from '../../api'
import { mediaRequest } from '../../utils/request'

const checkinId = ref('')
const checkin = ref<any>(null)
const videoURL = ref('')
const comments = ref<any[]>([])
const newComment = ref('')

onLoad((options) => {
  checkinId.value = options?.id || ''
  loadData()
})

async function loadData() {
  try {
    checkin.value = await getCheckinDetail(checkinId.value)
    if (checkin.value?.video_url) {
      const urlRes: any = await mediaRequest({
        url: `/media/url?object=${encodeURIComponent(checkin.value.video_url)}`,
      })
      videoURL.value = urlRes.url
    }
  } catch (e) {}

  // 获取评论
  try {
    const res: any = await getComments(checkinId.value)
    comments.value = res.list || []
  } catch (e) {}
}

async function onLike() {
  try {
    const res: any = await toggleLike(checkinId.value)
    checkin.value.is_liked = res.liked
    checkin.value.like_count += res.liked ? 1 : -1
  } catch (e) {}
}

async function onComment() {
  if (!newComment.value.trim()) return
  try {
    const c = await addComment(checkinId.value, newComment.value)
    comments.value.unshift(c)
    checkin.value.comment_count++
    newComment.value = ''
  } catch (e) {}
}
</script>

<style scoped>
.detail-page {
  padding-bottom: 120rpx;
}
.video-player {
  width: 100%;
  height: 420rpx;
}
.info {
  padding: 24rpx;
  background: #fff;
}
.author-row {
  display: flex;
  align-items: center;
  margin-bottom: 16rpx;
}
.avatar {
  width: 60rpx;
  height: 60rpx;
  border-radius: 50%;
  margin-right: 16rpx;
}
.author-name {
  font-size: 28rpx;
  font-weight: bold;
}
.description {
  font-size: 30rpx;
  line-height: 1.6;
}
.stats {
  display: flex;
  gap: 30rpx;
  margin-top: 16rpx;
  font-size: 24rpx;
  color: #999;
}
.actions {
  padding: 16rpx 24rpx;
  background: #fff;
  margin-top: 16rpx;
}
.action-btn {
  font-size: 28rpx;
  background: #f0f0f0;
  border-radius: 32rpx;
}
.comments-section {
  padding: 24rpx;
  background: #fff;
  margin-top: 16rpx;
}
.section-title {
  font-size: 30rpx;
  font-weight: bold;
  margin-bottom: 20rpx;
  display: block;
}
.comment-item {
  display: flex;
  margin-bottom: 24rpx;
}
.comment-avatar {
  width: 48rpx;
  height: 48rpx;
  border-radius: 50%;
  margin-right: 16rpx;
}
.comment-name {
  font-size: 24rpx;
  color: #666;
  display: block;
}
.comment-text {
  font-size: 28rpx;
  margin-top: 4rpx;
}
.comment-input {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  padding: 16rpx 24rpx;
  background: #fff;
  border-top: 1rpx solid #f0f0f0;
}
.comment-input input {
  flex: 1;
  background: #f5f5f5;
  border-radius: 32rpx;
  padding: 16rpx 24rpx;
  font-size: 28rpx;
}
.send-btn {
  margin-left: 16rpx;
  background: #1cbbb4;
  color: #fff;
  font-size: 28rpx;
  border-radius: 32rpx;
  padding: 0 24rpx;
  height: 64rpx;
  line-height: 64rpx;
}
</style>
