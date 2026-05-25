<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'
import { useUserStore } from './store/user'

onLaunch(() => {
  const userStore = useUserStore()
  userStore.loadToken()
})
</script>

<style>
page {
  background-color: #f8f8f8;
  font-size: 28rpx;
  color: #333;
}
</style>
