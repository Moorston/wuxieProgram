<template>
  <view class="share-page">
    <view class="section">
      <text class="section-title">分享范围</text>
      <view class="scope-options">
        <text class="scope-option" :class="{ active: shareScope === 'private' }" @tap="shareScope = 'private'">🔒 仅自己</text>
        <text class="scope-option" :class="{ active: shareScope === 'group' }" @tap="shareScope = 'group'">👥 考核组</text>
        <text class="scope-option" :class="{ active: shareScope === 'public' }" @tap="shareScope = 'public'">🌐 公开</text>
      </view>
    </view>

    <view v-if="shareScope === 'group'" class="section">
      <text class="section-title">选择考核组</text>
      <view v-for="g in groups" :key="g.id" class="group-item" :class="{ active: selectedGroup === g.id }" @tap="selectedGroup = g.id">
        <text class="group-name">{{ g.name }}</text>
        <text class="group-count">{{ g.members?.length || 0 }}人</text>
      </view>
    </view>

    <button class="submit-btn" :disabled="!canSubmit || saving" @tap="onSave">
      {{ saving ? '保存中...' : '保存设置' }}
    </button>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { updateResource, getGroupList } from '../../api'

const resourceId = ref('')
const shareScope = ref('private')
const selectedGroup = ref('')
const groups = ref<any[]>([])
const saving = ref(false)

const canSubmit = computed(() => shareScope.value !== 'group' || selectedGroup.value)

onLoad((options) => { resourceId.value = options?.id || '' })

onMounted(async () => {
  try { groups.value = (await getGroupList()) as any[] } catch (e) {}
})

async function onSave() {
  if (!canSubmit.value) return
  saving.value = true
  try {
    await updateResource(resourceId.value, {
      share_scope: shareScope.value,
      group_id: selectedGroup.value,
    })
    uni.showToast({ title: '已保存', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 1500)
  } catch (e) {
    uni.showToast({ title: '保存失败', icon: 'none' })
  } finally { saving.value = false }
}
</script>

<style scoped>
.share-page { padding: 20rpx; }
.section { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; }
.section-title { font-size: 30rpx; font-weight: bold; margin-bottom: 16rpx; display: block; }
.scope-options { display: flex; gap: 12rpx; }
.scope-option { flex: 1; text-align: center; padding: 16rpx; border-radius: 12rpx; background: #f8f8f8; font-size: 26rpx; }
.scope-option.active { background: #e8f5e9; color: #4caf50; }
.group-item { display: flex; justify-content: space-between; padding: 16rpx; border-bottom: 1rpx solid #f0f0f0; }
.group-item.active { background: #e8f5e9; border-radius: 8rpx; }
.group-name { font-size: 28rpx; }
.group-count { font-size: 24rpx; color: #999; }
.submit-btn { margin-top: 20rpx; background: #1cbbb4; color: #fff; border-radius: 40rpx; height: 88rpx; line-height: 88rpx; font-size: 32rpx; }
.submit-btn[disabled] { opacity: 0.5; }
</style>
