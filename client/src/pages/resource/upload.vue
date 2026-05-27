<template>
  <view class="upload-page">
    <view class="file-area" @tap="chooseFile">
      <view v-if="!fileInfo" class="placeholder">
        <text class="icon">+</text>
        <text class="tip">点击选择文件</text>
        <text class="hint">支持视频/图片/文档</text>
      </view>
      <view v-else class="file-preview">
        <text class="file-icon">{{ typeIcon[form.type] }}</text>
        <text class="file-name">{{ fileInfo.name }}</text>
        <text class="file-size">{{ formatSize(fileInfo.size) }}</text>
        <text class="file-remove" @tap.stop="fileInfo = null; form.file_url = ''">✕</text>
      </view>
    </view>

    <view class="form-section">
      <view class="form-item">
        <text class="label">标题</text>
        <input v-model="form.title" class="input" placeholder="输入资料标题" />
      </view>
      <view class="form-item">
        <text class="label">描述</text>
        <textarea v-model="form.description" class="textarea" placeholder="简要描述..." maxlength="200" />
      </view>
      <view class="form-row">
        <view class="form-item half">
          <text class="label">分类</text>
          <picker :value="categoryIndex" :range="categoryOptions" @change="categoryIndex = $event.detail.value; form.category = categoryOptions[categoryIndex]">
            <text class="picker-text">{{ form.category || '选择分类' }}</text>
          </picker>
        </view>
        <view class="form-item half">
          <text class="label">难度</text>
          <picker :value="difficultyIndex" :range="difficultyOptions" @change="difficultyIndex = $event.detail.value; form.difficulty = difficultyOptions[difficultyIndex]">
            <text class="picker-text">{{ form.difficulty || '选择难度' }}</text>
          </picker>
        </view>
      </view>
      <view class="form-item">
        <text class="label">标签</text>
        <view class="tag-list">
          <view v-for="(tag, i) in form.tags" :key="i" class="tag-item">
            <text>{{ tag }}</text>
            <text class="tag-remove" @tap="form.tags.splice(i, 1)">✕</text>
          </view>
          <text class="tag-add" @tap="showAddTag = true">+ 标签</text>
        </view>
      </view>
      <view class="form-item">
        <text class="label">分享范围</text>
        <view class="scope-options">
          <text class="scope-option" :class="{ active: form.share_scope === 'private' }" @tap="form.share_scope = 'private'">🔒 仅自己</text>
          <text class="scope-option" :class="{ active: form.share_scope === 'group' }" @tap="form.share_scope = 'group'">👥 考核组</text>
          <text class="scope-option" :class="{ active: form.share_scope === 'public' }" @tap="form.share_scope = 'public'">🌐 公开</text>
        </view>
      </view>
    </view>

    <view v-if="uploading" class="progress-bar">
      <view class="progress-fill" :style="{ width: progress + '%' }"></view>
      <text class="progress-text">{{ progress }}%</text>
    </view>

    <button class="submit-btn" :disabled="!canSubmit || uploading" @tap="onSubmit">
      {{ uploading ? '上传中...' : '上传资料' }}
    </button>

    <view v-if="showAddTag" class="tag-modal" @tap.self="showAddTag = false">
      <view class="modal-content">
        <text class="modal-title">添加标签</text>
        <input v-model="newTag" class="tag-input" placeholder="输入标签" />
        <view class="modal-btns">
          <text class="modal-btn cancel" @tap="showAddTag = false">取消</text>
          <text class="modal-btn confirm" @tap="addTag">确定</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { getResourcePresign, resourceUploadCallback } from '../../api'
import { mediaRequest } from '../../utils/request'

const typeIcon: Record<string, string> = { video: '🎬', image: '🖼', document: '📄' }
const categoryOptions = ['基本功', '套路', '散打', '太极', '理论']
const difficultyOptions = ['初级', '中级', '高级']

const categoryIndex = ref(0)
const difficultyIndex = ref(0)
const fileInfo = ref<any>(null)
const uploading = ref(false)
const progress = ref(0)
const showAddTag = ref(false)
const newTag = ref('')

const form = reactive({
  title: '',
  description: '',
  type: 'video',
  category: '',
  difficulty: '',
  tags: [] as string[],
  share_scope: 'private',
  file_url: '',
  file_size: 0,
})

const canSubmit = computed(() => form.title && form.file_url)

function chooseFile() {
  uni.chooseVideo({
    sourceType: ['camera', 'album'],
    compressed: true,
    success: (res) => {
      fileInfo.value = { name: res.tempFilePath.split('/').pop(), size: res.size, path: res.tempFilePath }
      form.type = 'video'
      form.file_size = res.size
    },
  })
}

function addTag() {
  if (newTag.value.trim() && !form.tags.includes(newTag.value.trim())) {
    form.tags.push(newTag.value.trim())
  }
  newTag.value = ''
  showAddTag.value = false
}

function formatSize(bytes: number) {
  if (!bytes) return '0B'
  if (bytes < 1024) return bytes + 'B'
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + 'KB'
  if (bytes < 1073741824) return (bytes / 1048576).toFixed(1) + 'MB'
  return (bytes / 1073741824).toFixed(1) + 'GB'
}

async function onSubmit() {
  if (!canSubmit.value) return
  uploading.value = true
  progress.value = 0
  try {
    const ext = fileInfo.value.path.split('.').pop() || 'mp4'
    const presign: any = await getResourcePresign(ext)

    await new Promise<void>((resolve, reject) => {
      const task = uni.uploadFile({
        url: presign.upload_url,
        filePath: fileInfo.value.path,
        name: 'file',
        success: (res) => res.statusCode === 200 ? resolve() : reject(new Error('upload failed')),
        fail: reject,
      })
      task.onProgressUpdate((res) => { progress.value = res.progress })
    })

    await mediaRequest({
      url: '/media/upload/callback',
      method: 'POST',
      data: { checkin_id: '', object_name: presign.object_name, bucket: presign.bucket },
    })

    await resourceUploadCallback({
      object_name: presign.object_name,
      bucket: 'resource',
      file_size: form.file_size,
      title: form.title,
      cover_url: '',
      duration: 0,
    })

    uni.showToast({ title: '上传成功', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 1500)
  } catch (e) {
    uni.showToast({ title: '上传失败', icon: 'none' })
  } finally {
    uploading.value = false
  }
}
</script>

<style scoped>
.upload-page { padding: 20rpx; }
.file-area { background: #fff; border-radius: 16rpx; padding: 40rpx; margin-bottom: 16rpx; text-align: center; }
.placeholder { color: #ccc; }
.icon { font-size: 80rpx; display: block; }
.tip { font-size: 28rpx; margin-top: 12rpx; display: block; }
.hint { font-size: 22rpx; color: #999; margin-top: 8rpx; display: block; }
.file-preview { display: flex; align-items: center; gap: 16rpx; }
.file-icon { font-size: 48rpx; }
.file-name { flex: 1; font-size: 28rpx; text-align: left; }
.file-size { font-size: 24rpx; color: #999; }
.file-remove { font-size: 28rpx; color: #e54d42; }
.form-section { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; }
.form-item { margin-bottom: 16rpx; }
.form-item.half { flex: 1; }
.form-row { display: flex; gap: 16rpx; }
.label { font-size: 26rpx; font-weight: bold; margin-bottom: 8rpx; display: block; }
.input { border: 1rpx solid #e0e0e0; border-radius: 12rpx; padding: 14rpx; font-size: 28rpx; }
.textarea { border: 1rpx solid #e0e0e0; border-radius: 12rpx; padding: 14rpx; font-size: 28rpx; height: 120rpx; }
.picker-text { border: 1rpx solid #e0e0e0; border-radius: 12rpx; padding: 14rpx; font-size: 28rpx; display: block; }
.tag-list { display: flex; flex-wrap: wrap; gap: 12rpx; }
.tag-item { display: flex; align-items: center; background: #f0f0f0; padding: 6rpx 14rpx; border-radius: 20rpx; font-size: 24rpx; }
.tag-remove { margin-left: 6rpx; color: #999; }
.tag-add { padding: 6rpx 14rpx; border: 1rpx dashed #ccc; border-radius: 20rpx; font-size: 24rpx; color: #999; }
.scope-options { display: flex; gap: 12rpx; }
.scope-option { flex: 1; text-align: center; padding: 14rpx; border-radius: 12rpx; background: #f8f8f8; font-size: 24rpx; }
.scope-option.active { background: #e8f5e9; color: #4caf50; }
.progress-bar { margin-top: 16rpx; background: #f0f0f0; border-radius: 8rpx; height: 16rpx; position: relative; }
.progress-fill { height: 100%; background: #1cbbb4; border-radius: 8rpx; }
.progress-text { position: absolute; right: 0; top: -40rpx; font-size: 24rpx; color: #666; }
.submit-btn { margin-top: 16rpx; background: #1cbbb4; color: #fff; border-radius: 40rpx; height: 88rpx; line-height: 88rpx; font-size: 32rpx; }
.submit-btn[disabled] { opacity: 0.5; }
.tag-modal { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 999; }
.modal-content { background: #fff; border-radius: 16rpx; padding: 32rpx; width: 600rpx; }
.modal-title { font-size: 32rpx; font-weight: bold; display: block; margin-bottom: 24rpx; }
.tag-input { border: 1rpx solid #e0e0e0; border-radius: 12rpx; padding: 16rpx; font-size: 28rpx; margin-bottom: 24rpx; }
.modal-btns { display: flex; gap: 16rpx; }
.modal-btn { flex: 1; text-align: center; padding: 16rpx; border-radius: 12rpx; font-size: 28rpx; }
.modal-btn.cancel { background: #f0f0f0; color: #666; }
.modal-btn.confirm { background: #1cbbb4; color: #fff; }
</style>
