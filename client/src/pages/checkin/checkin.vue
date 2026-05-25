<template>
  <view class="checkin-page">
    <view class="upload-area" @tap="chooseVideo">
      <view v-if="!videoPath" class="placeholder">
        <text class="icon">+</text>
        <text class="tip">点击录制或选择视频</text>
      </view>
      <video v-else :src="videoPath" class="preview" controls />
    </view>

    <view class="form">
      <textarea
        v-model="description"
        class="desc-input"
        placeholder="写点什么记录今天的训练..."
        maxlength="200"
      />
      <view class="char-count">{{ description.length }}/200</view>
    </view>

    <view class="progress-bar" v-if="uploading">
      <view class="progress" :style="{ width: progress + '%' }"></view>
      <text class="progress-text">{{ progress }}%</text>
    </view>

    <button class="submit-btn" :disabled="!videoPath || uploading" @tap="onSubmit">
      {{ uploading ? '上传中...' : '提交打卡' }}
    </button>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { prepareCheckin } from '../../api'
import { mediaRequest } from '../../utils/request'

const videoPath = ref('')
const description = ref('')
const uploading = ref(false)
const progress = ref(0)

function chooseVideo() {
  uni.chooseVideo({
    sourceType: ['camera', 'album'],
    maxDuration: 60,
    compressed: true,
    success: (res) => {
      videoPath.value = res.tempFilePath
    },
  })
}

async function onSubmit() {
  if (!videoPath.value) return

  uploading.value = true
  progress.value = 0

  try {
    // 1. 准备打卡记录
    const checkin: any = await prepareCheckin(description.value)

    // 2. 获取预签名URL
    const presign: any = await mediaRequest({
      url: `/media/upload/presign?checkin_id=${checkin.id}&ext=mp4`,
    })

    // 3. 上传视频到MinIO
    await uploadFile(presign.upload_url, videoPath.value)

    // 4. 回调media-server
    await mediaRequest({
      url: '/media/upload/callback',
      method: 'POST',
      data: {
        checkin_id: checkin.id,
        object_name: presign.object_name,
        bucket: presign.bucket,
      },
    })

    uni.showToast({ title: '打卡成功，视频转码中', icon: 'success' })
    setTimeout(() => {
      uni.switchTab({ url: '/pages/index/index' })
    }, 1500)
  } catch (e) {
    uni.showToast({ title: '上传失败', icon: 'none' })
  } finally {
    uploading.value = false
  }
}

function uploadFile(url: string, filePath: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const uploadTask = uni.uploadFile({
      url,
      filePath,
      name: 'file',
      success: (res) => {
        if (res.statusCode === 200) {
          resolve()
        } else {
          reject(new Error('upload failed'))
        }
      },
      fail: reject,
    })

    uploadTask.onProgressUpdate((res) => {
      progress.value = res.progress
    })
  })
}
</script>

<style scoped>
.checkin-page {
  padding: 30rpx;
}
.upload-area {
  width: 100%;
  height: 400rpx;
  background: #fff;
  border-radius: 16rpx;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}
.placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  color: #ccc;
}
.icon {
  font-size: 80rpx;
}
.tip {
  font-size: 28rpx;
  margin-top: 16rpx;
}
.preview {
  width: 100%;
  height: 100%;
}
.form {
  margin-top: 30rpx;
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
}
.desc-input {
  width: 100%;
  height: 200rpx;
  font-size: 28rpx;
}
.char-count {
  text-align: right;
  font-size: 22rpx;
  color: #999;
  margin-top: 8rpx;
}
.progress-bar {
  margin-top: 20rpx;
  background: #f0f0f0;
  border-radius: 8rpx;
  height: 16rpx;
  position: relative;
}
.progress {
  height: 100%;
  background: #1cbbb4;
  border-radius: 8rpx;
  transition: width 0.3s;
}
.progress-text {
  position: absolute;
  right: 0;
  top: -40rpx;
  font-size: 24rpx;
  color: #666;
}
.submit-btn {
  margin-top: 40rpx;
  background: #1cbbb4;
  color: #fff;
  border-radius: 40rpx;
  height: 80rpx;
  line-height: 80rpx;
  font-size: 32rpx;
}
.submit-btn[disabled] {
  opacity: 0.5;
}
</style>
