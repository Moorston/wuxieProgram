/**
 * 分享 composable
 * 在页面中使用：
 *   const { shareOptions, setShare, onShareAppMessage, onShareTimeline } = useShare()
 *
 * 在 onMounted 中设置分享内容：
 *   setShare({ title: '...', path: '...' })
 *
 * 在页面中导出生命周期钩子：
 *   defineExpose({ onShareAppMessage, onShareTimeline })
 */

import { ref } from 'vue'
import { getDefaultShareOptions, handleShareMessage, handleShareTimeline } from '../utils/share'
import type { ShareOptions } from '../utils/share'

export function useShare() {
  const shareOptions = ref<ShareOptions>(getDefaultShareOptions())

  function setShare(options: ShareOptions) {
    shareOptions.value = options
  }

  function onShareAppMessage() {
    return handleShareMessage(shareOptions.value)
  }

  function onShareTimeline() {
    return handleShareTimeline(shareOptions.value)
  }

  return {
    shareOptions,
    setShare,
    onShareAppMessage,
    onShareTimeline,
  }
}
