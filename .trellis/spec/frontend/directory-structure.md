# 前端目录结构规范

> 武俱打卡项目前端代码的组织方式和命名约定。

---

## 概览

前端使用 **uni-app (Vue3 + TypeScript)** 框架，一套代码编译到微信小程序、Android、iOS。采用**页面即路由**的组织方式，`pages/` 下每个子目录对应一个功能模块。

---

## 目录布局

```
client/
├── package.json                # 依赖管理 + scripts
├── tsconfig.json               # TypeScript 编译配置
├── vite.config.ts              # Vite 构建配置
└── src/
    ├── main.ts                 # 应用入口
    ├── App.vue                 # 根组件
    ├── manifest.json           # uni-app 应用配置（打包时使用）
    ├── pages.json              # 路由配置（30+ 页面 + 5 个 TabBar）
    │
    ├── api/
    │   └── index.ts            # 接口封装（80+ API 方法）
    │
    ├── store/
    │   ├── user.ts             # 用户状态（Pinia）
    │   └── checkin.ts          # 打卡列表状态（Pinia）
    │
    ├── utils/
    │   └── request.ts          # HTTP 请求封装（含 mediaRequest 鉴权）
    │
    ├── components/             # 可复用组件
    │   └── ...                 # 全局性可复用组件
    │
    ├── static/                 # 静态资源
    │   ├── tab/                # TabBar 图标
    │   │   ├── home.png
    │   │   ├── square.png
    │   │   ├── checkin.png
    │   │   ├── rank.png
    │   │   └── mine.png
    │   ├── logo.png
    │   └── avatar.png          # 默认头像
    │
    └── pages/                  # 页面（按功能模块组织）
        ├── index/              # 首页
        │   └── index.vue
        ├── login/              # 登录
        │   └── login.vue
        ├── square/             # 广场
        │   └── square.vue
        ├── checkin/            # 打卡
        │   └── checkin.vue
        ├── rank/               # 排行榜
        │   └── rank.vue
        ├── mine/               # 我的
        │   └── mine.vue
        ├── video-detail/       # 视频详情
        │   └── video-detail.vue
        ├── my-video/           # 我的视频
        │   └── my-video.vue
        ├── group/
        │   ├── group.vue       # 考核组列表
        │   └── detail.vue      # 考核组详情
        ├── training/
        │   ├── list.vue        # 训练计划列表
        │   ├── create.vue      # 创建计划
        │   ├── detail.vue      # 计划详情
        │   ├── template.vue    # 模板列表
        │   ├── template-detail.vue  # 模板详情
        │   ├── today.vue       # 今日任务
        │   └── report.vue      # 训练报告
        ├── insight/
        │   ├── list.vue        # 感悟时间线
        │   ├── create.vue      # 写感悟
        │   ├── detail.vue      # 感悟详情
        │   ├── tags.vue        # 标签管理
        │   ├── mood.vue        # 心情统计
        │   ├── on-this-day.vue # 历史今日
        │   └── public.vue      # 公开感悟广场
        ├── notification/
        │   ├── list.vue        # 通知列表
        │   └── settings.vue    # 通知设置
        └── resource/
            ├── list.vue        # 资料库（瀑布流/列表）
            ├── upload.vue      # 上传资料
            ├── detail.vue      # 资料详情
            ├── favorites.vue   # 收藏列表
            ├── share.vue       # 分享设置
            ├── tags.vue        # 资料标签
            └── stats.vue       # 存储统计
```

---

## 模块组织

### 页面即目录

- 每个功能模块一个子目录（`pages/training/`, `pages/insight/` 等）
- 每个页面一个 `.vue` 文件
- 大模块（>3 页面）进一步拆分子目录

### 全局资源

- `api/` — 所有接口调用，按模块组织函数
- `store/` — 全局状态（Pinia）
- `components/` — 可复用的页面组件
- `utils/` — 工具函数
- `static/` — 静态资源

### 页面结构与路由映射

路由配置在 `pages.json` 集中管理，每个页面条目包含 `path` 和 `style`：

```json
{
  "pages": [
    {"path": "pages/index/index", "style": {"navigationBarTitleText": "首页"}},
    {"path": "pages/login/login", "style": {"navigationBarTitleText": "登录"}}
  ],
  "tabBar": {
    "list": [
      {"pagePath": "pages/index/index", "text": "首页"},
      {"pagePath": "pages/square/square", "text": "广场"},
      {"pagePath": "pages/checkin/checkin", "text": "打卡"},
      {"pagePath": "pages/rank/rank", "text": "排位"},
      {"pagePath": "pages/mine/mine", "text": "我的"}
    ]
  }
}
```

---

## 命名约定

### 文件命名

- **Vue 页面文件**：小写 kebab-case（`video-detail.vue`, `on-this-day.vue`）
- **TypeScript 文件**：小写 kebab-case（`request.ts`, `user.ts`）
- **静态资源**：小写英文（`home.png`, `logo.png`）

### 目录命名

- **页面目录**：全小写英文（`training/`, `insight/`, `notification/`）
- **功能目录**：全小写英文（`api/`, `store/`, `utils/`）

### 组件命名

- **Vue 组件**：PascalCase（但 uni-app 文件本身用小写 kebab-case）
- **函数命名**：camelCase（`getCheckinList`, `toggleLike`）
- **类型/接口**：PascalCase（`ApiResponse<T>`, `RequestOptions`）

---

## 代码示例

### API 封装模式（api/index.ts）

每个 API 函数对应一个后端接口，统一使用 `request()`：

```typescript
// ✅ 正确的 API 封装
export function getCheckinDetail(id: string) {
  return request({ url: `/api/checkin/${id}` })
}

export function toggleLike(id: string) {
  return request({ url: `/api/checkin/${id}/like`, method: 'POST' })
}
```

### 页面组件模式

```vue
<template>
  <view class="container">
    <!-- uni-app 标签，非 HTML -->
    <uni-nav-bar title="页面标题" />
    <scroll-view @scrolltolower="loadMore">
      <view v-for="item in list" :key="item._id">
        <!-- 内容 -->
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getCheckinList } from '@/api'

const list = ref<any[]>([])
const loading = ref(false)

onMounted(() => { loadMore() })

async function loadMore() {
  if (loading.value) return
  loading.value = true
  try {
    const res = await getCheckinList(page.value, pageSize.value)
    list.value.push(...res.list)
  } finally {
    loading.value = false
  }
}
</script>
```

---

## 本项目现状

- ✅ 30+ 页面已按模块组织
- ✅ 80+ API 方法集中在 `api/index.ts`
- ✅ TabBar 5 页签配置清晰
- ✅ Pinia store 分离管理用户状态和打卡列表状态
- ❌ `components/` 目录几乎为空（缺少复用组件提取）
- ❌ 部分页面逻辑重复（如分页加载代码在每个列表页面重复出现）