<template>
  <view class="create-page">
    <view class="mood-section">
      <text class="section-label">今天的心情</text>
      <view class="mood-picker">
        <view v-for="m in moods" :key="m.value" class="mood-item" :class="{ active: form.mood === m.value }" @tap="form.mood = m.value">
          <text class="mood-icon">{{ m.icon }}</text>
          <text class="mood-name">{{ m.label }}</text>
        </view>
      </view>
    </view>

    <view class="content-section">
      <textarea v-model="form.content" class="content-input" placeholder="记录今天的训练感悟..." maxlength="2000" />
      <text class="char-count">{{ form.content.length }}/2000</text>
    </view>

    <view class="images-section">
      <view class="image-list">
        <view v-for="(img, i) in form.images" :key="i" class="image-item">
          <image class="preview-img" :src="img" mode="aspectFill" />
          <text class="remove-img" @tap="removeImage(i)">✕</text>
        </view>
        <view v-if="form.images.length < 9" class="add-image" @tap="chooseImage">
          <text class="add-icon">+</text>
        </view>
      </view>
    </view>

    <view class="tags-section">
      <text class="section-label">标签</text>
      <view class="tag-list">
        <view v-for="(tag, i) in form.tags" :key="i" class="tag-item">
          <text>{{ tag }}</text>
          <text class="tag-remove" @tap="removeTag(i)">✕</text>
        </view>
        <view class="tag-add" @tap="showAddTag = true">
          <text>+ 添加标签</text>
        </view>
      </view>
    </view>

    <view class="visibility-section">
      <text class="section-label">可见性</text>
      <view class="visibility-options">
        <text class="vis-option" :class="{ active: form.visibility === 'private' }" @tap="form.visibility = 'private'">🔒 仅自己</text>
        <text class="vis-option" :class="{ active: form.visibility === 'public' }" @tap="form.visibility = 'public'">🌐 公开到广场</text>
      </view>
    </view>

    <button class="submit-btn" :disabled="!form.content || submitting" @tap="onSubmit">
      {{ submitting ? '保存中...' : (isEdit ? '保存修改' : '发布感悟') }}
    </button>

    <view v-if="showAddTag" class="tag-modal" @tap.self="showAddTag = false">
      <view class="modal-content">
        <text class="modal-title">添加标签</text>
        <input v-model="newTag" class="tag-input" placeholder="输入标签名称" />
        <view class="modal-btns">
          <text class="modal-btn cancel" @tap="showAddTag = false">取消</text>
          <text class="modal-btn confirm" @tap="addTag">确定</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { createInsight, getInsight, updateInsight } from '../../api'

const isEdit = ref(false)
const editId = ref('')
const submitting = ref(false)
const showAddTag = ref(false)
const newTag = ref('')

const moods = [
  { value: 'breakthrough', icon: '🔥', label: '突破' },
  { value: 'good', icon: '😊', label: '满意' },
  { value: 'normal', icon: '😐', label: '一般' },
  { value: 'confused', icon: '🤔', label: '困惑' },
  { value: 'low', icon: '😔', label: '低落' },
]

const form = reactive({
  content: '',
  images: [] as string[],
  mood: 'good',
  tags: [] as string[],
  visibility: 'private',
})

onLoad((options) => {
  if (options?.id) {
    isEdit.value = true
    editId.value = options.id
    loadInsight()
  }
})

async function loadInsight() {
  try {
    const res: any = await getInsight(editId.value)
    form.content = res.content || ''
    form.images = res.images || []
    form.mood = res.mood || 'good'
    form.tags = res.tags || []
    form.visibility = res.visibility || 'private'
  } catch (e) {}
}

function chooseImage() {
  uni.chooseImage({
    count: 9 - form.images.length,
    sizeType: ['compressed'],
    success: (res) => {
      form.images.push(...res.tempFilePaths)
    },
  })
}

function removeImage(index: number) {
  form.images.splice(index, 1)
}

function addTag() {
  if (newTag.value.trim() && !form.tags.includes(newTag.value.trim())) {
    form.tags.push(newTag.value.trim())
  }
  newTag.value = ''
  showAddTag.value = false
}

function removeTag(index: number) {
  form.tags.splice(index, 1)
}

async function onSubmit() {
  if (!form.content) return
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateInsight(editId.value, { ...form })
    } else {
      await createInsight({ ...form })
    }
    uni.showToast({ title: isEdit.value ? '修改成功' : '发布成功', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 1500)
  } catch (e) {
    uni.showToast({ title: '操作失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.create-page {
  padding: 20rpx;
}
.mood-section, .content-section, .images-section, .tags-section, .visibility-section {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.section-label {
  font-size: 28rpx;
  font-weight: bold;
  margin-bottom: 16rpx;
  display: block;
}
.mood-picker {
  display: flex;
  gap: 16rpx;
}
.mood-item {
  flex: 1;
  text-align: center;
  padding: 16rpx 0;
  border-radius: 12rpx;
  background: #f8f8f8;
}
.mood-item.active {
  background: #e8f5e9;
}
.mood-icon {
  font-size: 40rpx;
  display: block;
}
.mood-name {
  font-size: 22rpx;
  color: #666;
  margin-top: 4rpx;
}
.content-input {
  width: 100%;
  height: 300rpx;
  font-size: 28rpx;
  line-height: 1.6;
}
.char-count {
  font-size: 22rpx;
  color: #999;
  text-align: right;
  margin-top: 8rpx;
}
.image-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}
.image-item {
  position: relative;
  width: 200rpx;
  height: 200rpx;
}
.preview-img {
  width: 100%;
  height: 100%;
  border-radius: 8rpx;
}
.remove-img {
  position: absolute;
  top: -8rpx;
  right: -8rpx;
  width: 36rpx;
  height: 36rpx;
  background: #e54d42;
  color: #fff;
  border-radius: 50%;
  font-size: 20rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.add-image {
  width: 200rpx;
  height: 200rpx;
  border: 2rpx dashed #ccc;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.add-icon {
  font-size: 48rpx;
  color: #ccc;
}
.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}
.tag-item {
  display: flex;
  align-items: center;
  background: #f0f0f0;
  padding: 8rpx 16rpx;
  border-radius: 20rpx;
  font-size: 24rpx;
}
.tag-remove {
  margin-left: 8rpx;
  color: #999;
}
.tag-add {
  padding: 8rpx 16rpx;
  border: 1rpx dashed #ccc;
  border-radius: 20rpx;
  font-size: 24rpx;
  color: #999;
}
.visibility-options {
  display: flex;
  gap: 16rpx;
}
.vis-option {
  flex: 1;
  text-align: center;
  padding: 16rpx;
  border-radius: 12rpx;
  background: #f8f8f8;
  font-size: 26rpx;
}
.vis-option.active {
  background: #e8f5e9;
  color: #4caf50;
}
.submit-btn {
  margin-top: 16rpx;
  background: #1cbbb4;
  color: #fff;
  border-radius: 40rpx;
  height: 88rpx;
  line-height: 88rpx;
  font-size: 32rpx;
}
.submit-btn[disabled] {
  opacity: 0.5;
}
.tag-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}
.modal-content {
  background: #fff;
  border-radius: 16rpx;
  padding: 32rpx;
  width: 600rpx;
}
.modal-title {
  font-size: 32rpx;
  font-weight: bold;
  display: block;
  margin-bottom: 24rpx;
}
.tag-input {
  border: 1rpx solid #e0e0e0;
  border-radius: 12rpx;
  padding: 16rpx;
  font-size: 28rpx;
  margin-bottom: 24rpx;
}
.modal-btns {
  display: flex;
  gap: 16rpx;
}
.modal-btn {
  flex: 1;
  text-align: center;
  padding: 16rpx;
  border-radius: 12rpx;
  font-size: 28rpx;
}
.modal-btn.cancel {
  background: #f0f0f0;
  color: #666;
}
.modal-btn.confirm {
  background: #1cbbb4;
  color: #fff;
}
</style>
