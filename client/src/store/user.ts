import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
  const token = ref('')
  const userInfo = ref<any>(null)
  const isLoggedIn = ref(false)

  function loadToken() {
    const saved = uni.getStorageSync('token')
    if (saved) {
      token.value = saved
      isLoggedIn.value = true
    }
  }

  function setToken(t: string) {
    token.value = t
    isLoggedIn.value = true
    uni.setStorageSync('token', t)
  }

  function setUser(user: any) {
    userInfo.value = user
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    isLoggedIn.value = false
    uni.removeStorageSync('token')
  }

  return { token, userInfo, isLoggedIn, loadToken, setToken, setUser, logout }
})
