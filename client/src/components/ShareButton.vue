<template>
  <button class="share-btn" open-type="share" :style="btnStyle">
    <slot>{{ text }}</slot>
  </button>
</template>

<script setup lang="ts">
defineProps({
  text: { type: String, default: '分享' },
  btnStyle: { type: Object, default: () => ({}) },
})
</script>

<style scoped>
.share-btn {
  background: none;
  border: none;
  padding: 0;
  margin: 0;
  color: inherit;
  font-size: inherit;
  line-height: inherit;
}
.share-btn::after {
  display: none;
}
</style>
