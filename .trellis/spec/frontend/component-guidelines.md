# 组件规范

> 武俱打卡项目前端组件的构建方式和使用约定。

---

## 概览

本项目使用 **Vue3 `<script setup lang="ts">`** 组合式 API 构建组件。基于 uni-app 框架，页面组件是主要单元，全局性可复用组件较少。组件树以**页面组件**为根，内部使用 `uni-*` 内置组件构建 UI。

---

## 组件结构

### 页面组件模式

```vue
<template>
  <view class="container">
    <uni-nav-bar title="页面标题" />
    <scroll-view @scrolltolower="onLoadMore">
      <view v-for="item in dataList" :key="item._id" @click="goDetail(item._id)">
        <!-- 内容 -->
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { getList } from '@/api'

const dataList = ref<any[]>([])
const page = ref(1)
const loading = ref(false)

onMounted(() => { fetchData() })

async function fetchData() {
  if (loading.value) return
  loading.value = true
  try {
    const res = await getList(page.value, 20)
    dataList.value.push(...res.list)
    page.value++
  } finally {
    loading.value = false
  }
}

function goDetail(id: string) {
  uni.navigateTo({ url: `/pages/detail/detail?id=${id}` })
}
</script>

<style scoped>
.container { padding: 20rpx; }
</style>
```

### 组件结构约定

| 区域 | 说明 |
|------|------|
| `<template>` | 模板，使用 uni-app 内置标签（`view`, `scroll-view`, `uni-*`） |
| `<script setup lang="ts">` | 组合式 API + TypeScript |
| `import` | 先 Vue 核心 API，再 uni-app API，再项目模块，再样式 |
| `ref/reactive` | 响应式数据定义 |
| `computed/watch` | 派生数据 |
| `onMounted/onLoad` | 生命周期 |
| 函数定义 | 事件处理 + 业务逻辑 |
| `<style scoped>` | 组件样式（rpx 单位） |

---

## Props 约定

- Props 使用 TypeScript `defineProps<>` 泛型定义类型
- 必填 props 不加 `?`，可选 props 加 `?`
- prop 使用 camelCase 命名

```vue
<script setup lang="ts">
interface Props {
  title: string           // 必填
  description?: string    // 可选
  count?: number          // 可选，默认值
}

const props = withDefaults(defineProps<Props>(), {
  description: '',
  count: 0,
})
</script>
```

---

## 事件处理

- 使用 `defineEmits` + TypeScript 类型标注
- 事件名使用 kebab-case 或 camelCase

```vue
<script setup lang="ts">
const emit = defineEmits<{
  'update': [id: string]
  'delete': [id: string]
}>()

function handleUpdate(id: string) {
  emit('update', id)
}
</script>
```

---

## 样式规范

### 单位

- **rpx**（responsive pixel）：所有尺寸使用 rpx，uni-app 自动适配不同屏幕
- **不使用** `px`, `rem`, `em` 作为尺寸单位

```css
/* ✅ 正确 */
.container { padding: 20rpx; }
.title { font-size: 32rpx; }
.avatar { width: 80rpx; height: 80rpx; }

/* ❌ 禁止 */
.container { padding: 20px; }
```

### scoped

- 所有组件样式使用 `<style scoped>`，避免样式污染
- 全局样式在 `App.vue` 中定义

---

## 禁用模式

### ❌ 全局组件注册

uni-app 中应避免全局注册非内置组件，保持按需引入。

### ❌ 直接操作 DOM

使用 uni-app 的 `ref` + 组件方法而非直接 DOM 操作：

```vue
<!-- ✅ 正确 -->
<uni-input ref="inputRef" />

<!-- ❌ 禁止 -->
<view id="myView" />
<script>
  document.getElementById('myView')  // uni-app 中没有 document
</script>
```

### ❌ HTML 原生标签

在 uni-app 中使用 Vue 标签替代 HTML 标签：

| 场景 | HTML 标签 | uni-app 标签 |
|------|-----------|-------------|
| 容器 | `<div>` | `<view>` |
| 文本 | `<span>` | `<text>` |
| 图片 | `<img>` | `<image>` |
| 输入 | `<input>` | `<uni-input>` 或 `<input>` |

---

## 本项目现状

- ✅ 所有页面使用 Vue3 `<script setup>` 组合式 API
- ✅ 页面按功能模块独立，文件结构清晰
- ❌ `components/` 目录几乎为空，应该提取的复用组件包括：
  - 分页加载组件（`PullRefreshList`）
  - 视频卡片组件（`VideoCard`）
  - 用户信息组件（`UserInfo`）
  - 瀑布流布局组件（`WaterfallLayout`）
- ❌ 很多页面存在相同的分页逻辑重复代码