// 路由守卫 — 检查登录状态
// 在 App.vue 的 onLaunch 中调用

const PUBLIC_PAGES = [
  'pages/login/login',
  'pages/index/index',
  'pages/square/square',
  'pages/rank/rank',
]

export function checkAuth(to: string): boolean {
  const token = uni.getStorageSync('token')
  if (token) return true

  // 未登录时检查目标页面是否为公开页面
  const isPublic = PUBLIC_PAGES.some(p => to.startsWith(p))
  if (isPublic) return true

  // 非公开页面需要登录，跳转登录页
  uni.reLaunch({ url: '/pages/login/login' })
  return false
}