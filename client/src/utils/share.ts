/**
 * 微信分享工具
 * 在 UniApp 小程序中，分享通过页面的 onShareAppMessage/onShareTimeline 生命周期钩子实现
 * 使用方式：在页面中 import 并调用 useShare() 获取分享配置
 */

export interface ShareOptions {
  title: string
  path: string
  imageUrl?: string
}

/** 默认分享图片 */
const DEFAULT_SHARE_IMAGE = '/static/share-default.png'

/**
 * 生成打卡分享配置
 */
export function getCheckinShareOptions(checkinId: string, nickname: string, description: string): ShareOptions {
  const title = nickname
    ? `${nickname} 的武术训练打卡`
    : '武术训练打卡'
  const desc = description ? ` - ${description.slice(0, 20)}` : ''
  return {
    title: title + desc,
    path: `/pages/video-detail/video-detail?id=${checkinId}`,
    imageUrl: DEFAULT_SHARE_IMAGE,
  }
}

/**
 * 生成赛事分享配置
 */
export function getCompetitionShareOptions(competitionId: string, title: string): ShareOptions {
  return {
    title: `🏆 ${title} - 武俱打卡赛事`,
    path: `/pages/competition/detail?id=${competitionId}`,
    imageUrl: DEFAULT_SHARE_IMAGE,
  }
}

/**
 * 生成团组分享配置
 */
export function getGroupShareOptions(groupId: string, name: string): ShareOptions {
  return {
    title: `👥 ${name} - 武俱打卡团组`,
    path: `/pages/group/detail?id=${groupId}`,
    imageUrl: DEFAULT_SHARE_IMAGE,
  }
}

/**
 * 生成默认分享配置
 */
export function getDefaultShareOptions(): ShareOptions {
  return {
    title: '武俱打卡 - 让每一次训练都有迹可循',
    path: '/pages/index/index',
    imageUrl: DEFAULT_SHARE_IMAGE,
  }
}

/**
 * 通用分享处理函数（用于页面 onShareAppMessage）
 */
export function handleShareMessage(options?: ShareOptions) {
  return options || getDefaultShareOptions()
}

/**
 * 朋友圈分享处理函数（用于页面 onShareTimeline）
 */
export function handleShareTimeline(options?: ShareOptions) {
  const opts = options || getDefaultShareOptions()
  return {
    title: opts.title,
    query: `path=${encodeURIComponent(opts.path)}`,
    imageUrl: opts.imageUrl,
  }
}
