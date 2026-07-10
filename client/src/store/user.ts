import { defineStore } from 'pinia'
import { ref } from 'vue'
import { wxLogout } from '../api/index'

export const useUserStore = defineStore('user', () => {
  const token = ref('')
  const refreshToken = ref('')
  const userInfo = ref<any>(null)
  const isLoggedIn = ref(false)

  function loadToken() {
    const saved = uni.getStorageSync('token')
    if (saved) {
      token.value = saved
      isLoggedIn.value = true
    }
    const savedRefresh = uni.getStorageSync('refresh_token')
    if (savedRefresh) {
      refreshToken.value = savedRefresh
    }
  }

  function setToken(t: string, rt?: string) {
    token.value = t
    isLoggedIn.value = true
    uni.setStorageSync('token', t)
    if (rt) {
      refreshToken.value = rt
      uni.setStorageSync('refresh_token', rt)
    }
  }

  function setUser(user: any) {
    userInfo.value = user
  }

  async function logout() {
    // 调用服务端注销（将 token 加入黑名单）
    try {
      await wxLogout()
    } catch {
      // 忽略错误，继续清除本地状态
    }
    token.value = ''
    refreshToken.value = ''
    userInfo.value = null
    isLoggedIn.value = false
    uni.removeStorageSync('token')
    uni.removeStorageSync('refresh_token')
  }

  return { token, refreshToken, userInfo, isLoggedIn, loadToken, setToken, setUser, logout }
})
