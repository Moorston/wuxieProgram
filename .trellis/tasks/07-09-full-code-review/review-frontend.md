# 武俱打卡项目前端代码审查报告

审查日期：2026-07-09
审查范围：client/src/ 下全部 31 个前端文件
审查方式：人工审查 + 静态分析

---

## [CRITICAL] 问题列表

### 1. 头像上传发送本地临时路径给后端 —— 完全无法工作
- **文件**: `client/src/pages/mine/mine.vue:126-128`
- **描述**: `changeAvatar()` 调用 `uni.chooseImage` 后将 `res.tempFilePaths[0]`（本地临时文件路径）直接赋给 `editForm.avatar`，然后 `onSaveProfile()` 直接将这个临时路径通过 `updateProfile({ avatar: editForm.avatar })` 发送给后端。服务端无法访问客户端的临时文件路径，头像上传永远不会成功。
- **建议**: 应先调用 `uni.uploadFile` 将图片上传到媒体服务器，拿到 CDN URL 后再提交给后端。

### 2. 硬编码后端服务地址 —— 无法部署到生产
- **文件**: `client/src/utils/request.ts:1,54`
- **描述**: `BASE_URL = 'http://localhost:8080'` 和 `MEDIA_URL = 'http://localhost:8081'` 完全硬编码为 localhost，生产环境部署时必须修改代码才能切换到正式域名。同时使用 HTTP 而非 HTTPS，存在中间人攻击风险。
- **建议**: 通过 uni-app 的 `process.env.VITE_API_BASE_URL` 环境变量或 `manifest.json` 的配置项动态获取，且生产环境必须使用 HTTPS。

### 3. Token 无过期刷新机制
- **文件**: `client/src/store/user.ts:17-21`
- **描述**: `setToken()` 直接存储原始 token 到 Storage。`request.ts:37-39` 中 401 时只清空 token 跳转登录，没有任何 token 自动刷新逻辑。用户必须重新登录才能继续使用，体验差。
- **建议**: 增加 refresh_token 机制，拦截 401 响应时自动调用刷新接口，刷新失败再跳转登录。

### 4. 整个前端几乎 100% 使用 `any` 类型 —— 完全丧失类型安全
- **文件**: 全部 31 个文件的 API 调用、Store 声明、组件数据
- **描述**: `request<T>` 的泛型 T 从未被使用 —— 所有 API 调用写成 `const res: any = await xxx()`；所有 Store state 声明为 `ref<any>(null)` / `ref<any[]>([])`；所有组件数据声明为 `ref<any>(null)`。TypeScript 形同虚设。
- **建议**: 为每个 API 响应定义 TypeScript interface（如 `interface CheckinItem { id: string; description: string; ... }`），`request<T>` 调用时传入具体类型，彻底消灭 `any`。

### 5. 视频上传的预设签名请求 URL 未编码参数
- **文件**: `client/src/pages/checkin/checkin.vue:65`
- **描述**: `/media/upload/presign?checkin_id=${checkin.id}&ext=mp4` —— `checkin.id` 直接拼接到 URL 中，未做 `encodeURIComponent`。虽然 id 来自后端相对安全，但不符合安全编码规范。
- **文件**: `client/src/api/index.ts:202-204` `getResourcePresign(ext: string)` 中 `${ext}` 也未编码。
- **建议**: 对所有 URL 查询参数使用 `encodeURIComponent`。

### 6. 微信登录使用已弃用的 `uni.getUserProfile`
- **文件**: `client/src/pages/login/login.vue:26-32`
- **描述**: 微信官方自 2022 年起逐步弃用 `wx.getUserProfile` 接口（`uni.getUserProfile` 对应），新版小程序已无法获取用户头像昵称，登录后会得到空数据。
- **建议**: 改用 `uni.getUserInfo` + 头像昵称填写能力，或通过 `open-type="getUserInfo"` 按钮获取。

---

## [HIGH] 问题列表

### 7. 分页加载逻辑在 10+ 个页面完全重复 —— 严重 DRY 违反
- **文件**: 以下文件均有完全相同的 `loadData` / `refreshData` / `loadMore` 三件套模式：
  - `client/src/pages/index/index.vue:104-142`
  - `client/src/pages/square/square.vue:119-166`
  - `client/src/pages/rank/rank.vue:75-120`
  - `client/src/pages/my-video/my-video.vue:55-93`
  - `client/src/pages/training/list.vue:88-130`
  - `client/src/pages/insight/list.vue:99-137`
  - `client/src/pages/insight/public.vue:69-107`
  - `client/src/pages/notification/list.vue:107-145`
  - `client/src/pages/resource/list.vue:112-143`
  - `client/src/pages/resource/favorites.vue:31-49`
- **描述**: 每个页面重复实现 `page/total/loading/pageSize` 状态变量、`loadData` + `refreshData` + `loadMore` 三个函数、以及 `noMore` computed。约 40 行完全相同的模板代码，仅调用的 API 函数名不同。一旦需要修改分页逻辑（如增加错误重试），需逐个文件修改。
- **建议**: 抽取为 `usePagination` composable（`composables/usePagination.ts`），封装分页状态和方法，参数接受 API 函数即可。

### 8. 所有 `catch (e) {}` 吞噬异常 —— 调试灾难
- **文件**: 全部 31 个文件的几乎所有 try/catch
- **描述**: 绝大多数 try/catch 的 catch 块为空 `catch (e) {}`，没有任何 `console.error`、无上报、无用户提示。请求失败时用户和开发者都无法得知原因。
- **建议**: 至少在 catch 块中加入 `console.error('[page:function]', e)` 以便调试，并可根据需要展示用户友好的错误提示。

### 9. `formatTime` 函数在 4 个页面重复实现
- **文件**: 
  - `client/src/pages/my-video/my-video.vue:95-99`
  - `client/src/pages/insight/list.vue:166-174`
  - `client/src/pages/insight/public.vue:113-121`
  - `client/src/pages/notification/list.vue:178-189`
- **描述**: 4 个页面各自实现了日期格式化函数，逻辑相似但细节不同（有的只格式化到月-日，有的支持"刚刚/分钟前"等相对时间）。任何格式化需求变更需改 4 个地方。
- **建议**: 抽取为 `formatDate` 工具函数放在 `utils/format.ts` 中统一管理和复用。

### 10. `formatSize` 函数在 5 个页面重复实现
- **文件**: 
  - `client/src/pages/resource/list.vue:150-156`
  - `client/src/pages/resource/upload.vue:133-139`
  - `client/src/pages/resource/detail.vue:102-108`
  - `client/src/pages/resource/favorites.vue:57-61`
  - `client/src/pages/resource/stats.vue:41-47`
- **描述**: 完全相同的文件大小格式化逻辑在各资源页面重复。
- **建议**: 抽取到 `utils/format.ts`。

### 11. 类型字典常量（`typeMap`/`typeLabel`/`typeIcon`/`moodIcon`）在多个页面重复定义
- **文件**: 
  - `typeMap`（training 类型）在 `training/detail.vue:69-74`、`training/today.vue:46-51`、`training/template-detail.vue:49-54`、`training/report.vue:58-63`、`training/create.vue:67-72` 共 5 处
  - `typeIcon`/`typeLabel`（资源类型）在 `resource/list.vue:93-94`、`resource/upload.vue:87`、`resource/detail.vue:53`、`resource/favorites.vue:25`、`resource/stats.vue:34-35` 共 5 处
  - `moodIcon`（心情图标）在 `insight/list.vue:72-78`、`insight/detail.vue:41-47`、`insight/on-this-day.vue:38-44`、`insight/public.vue:48-54` 共 4 处
  - `statusMap`（训练状态）在 `training/list.vue:60-65`、`training/detail.vue:62-67` 共 2 处
- **建议**: 抽取到 `constants/` 目录共享。

### 12. 未使用的 `onMounted` 导入
- **文件**: 
  - `client/src/pages/video-detail/video-detail.vue:42` —— 使用 `onLoad` 代替
  - `client/src/pages/group/detail.vue:34` —— 使用 `onLoad` 代替
  - `client/src/pages/training/detail.vue:54` —— 使用 `onLoad` 代替
  - `client/src/pages/training/template-detail.vue:40` —— 使用 `onLoad` 代替
  - `client/src/pages/insight/create.vue:69` —— 使用 `onLoad` 代替
  - `client/src/pages/insight/detail.vue:34` —— 使用 `onLoad` 代替
  - `client/src/pages/resource/detail.vue:44` —— 使用 `onLoad` 代替
- **描述**: 以上文件导入了 `onMounted` 但从未使用（仅使用 `onLoad`）。不会导致运行错误，但影响编译后包的体积，混淆代码审查。
- **建议**: 删除未使用的 import 语句。

### 13. 图片列表未使用 `lazy-load` 属性 —— 性能浪费
- **文件**: 全部页面中涉及图片列表渲染的 `<image>` 标签
- **描述**: 所有 `<image>` 标签均未添加 `lazy-load` 属性，列表滚动时未显示的图片也会发起网络请求，浪费流量和内存。
- **建议**: 在列表页的 `<image>` 标签上增加 `lazy-load` 属性以启用 uni-app 的原生图片懒加载。

### 14. `resource/upload.vue` 仅支持视频上传，但声称支持图片/文档
- **文件**: `client/src/pages/resource/upload.vue:113-123`
- **描述**: `chooseFile()` 函数仅调用 `uni.chooseVideo`，但页面文案显示"支持视频/图片/文档"。用户无法选择图片或文档文件。
- **建议**: 根据 `form.type` 动态选择 `uni.chooseVideo` / `uni.chooseImage` / `uni.chooseFile`。

### 15. `resource/upload.vue` 上传回调时 `checkin_id` 传空字符串
- **文件**: `client/src/pages/resource/upload.vue:163`
- **描述**: `mediaRequest` 回调中 `data: { checkin_id: '', object_name: ..., bucket: ... }` 传了空字符串 `checkin_id`，如果后端收到空字符串可能覆盖原有关联或触发不必要的查询。
- **建议**: 不应发送空的 `checkin_id` 字段，应使用 `undefined` 或省略该字段。

---

## [MEDIUM] 问题列表

### 16. 瀑布流布局按奇偶索引分割，非真正瀑布流
- **文件**: 
  - `client/src/pages/square/square.vue:97-98`
  - `client/src/pages/resource/list.vue:90-91`
- **描述**: `leftList = list.filter((_, i) => i % 2 === 0)` 按索引奇偶平分两列，未考虑卡片高度。如果第一列某卡片因内容多而高，第二列对应位置可能堆积大量空白。真正的瀑布流应该按当前列高度决定放置。
- **建议**: 计算每列的累计高度，将新卡片插入总高度较小的列。

### 17. `v-for` 使用索引作 `key` —— 导致 DOM 复用错误
- **文件**: 
  - `client/src/pages/training/create.vue:34,40` —— 天数列表和任务列表的 `:key="dayIndex"` 和 `:key="taskIndex"`
  - `client/src/pages/insight/list.vue:24` —— 图片预览 `:key="i"`
  - `client/src/pages/resource/upload.vue:69` —— 标签列表 `:key="i"`
- **描述**: 在可增删的动态列表中使用数组索引作为 `key`，Vue 的虚拟 DOM diff 会错误复用元素，导致删除第一个元素时后续所有元素的状态混乱（input 值错位、picker 选中值错位）。
- **建议**: 对动态增删的列表使用唯一 ID 或 `Symbol()` 作为 `key`，而非索引。

### 18. 排行榜页 API 响应结构不确定性
- **文件**: `client/src/pages/rank/rank.vue:80,96,113`
- **描述**: `const res: any = await getRankList(...)` 后使用 `res.list || res || []` 的 fallback，表明开发者不确定 API 返回 `{ list: items }` 还是直接返回数组。这种不确定性应该通过明确类型定义解决。
- **建议**: 定义 `RankListResponse` interface 明确响应结构，移除 `|| res` 这种模糊回退。

### 19. 评论列表不支持分页
- **文件**: `client/src/pages/video-detail/video-detail.vue:70-73`
- **描述**: `getComments` API 支持 `page` 和 `pageSize` 参数，但该页面只加载第一页评论，没有"加载更多"功能。
- **建议**: 增加 `loadMoreComments()` 函数和"加载更多"按钮。

### 20. 硬编码的训练模板列表页大小为 50
- **文件**: `client/src/pages/training/template.vue:65`
- **描述**: `listTemplates(1, 50, category, style)` 硬编码 `pageSize=50` 且没有分页加载能力。如果模板数量增长超过 50 条，用户看不到更多。
- **建议**: 实现分页加载或在模板数量超过 50 时显示"查看更多"。

### 21. `uni.request` 未设置超时时间
- **文件**: `client/src/utils/request.ts:28-50`
- **描述**: `uni.request` 调用时未传入 `timeout` 参数，网络异常时请求会一直挂起直到系统超时（通常 60 秒以上），用户体验差。
- **建议**: 增加合理超时（如 10 秒），超时时自动 reject 并提示用户。

### 22. 空 catch 块导致部分 UI 状态异常
- **文件**: `client/src/pages/index/index.vue:82` 等
- **描述**: `getProfile()` 失败时 `userInfo.value` 保持 `null`，页面会展示"未登录"，但用户实际已登录（token 有效）。如果只是网络波动，用户看到"未登录"会产生困惑。
- **建议**: 区分"未登录"和"网络异常"两种状态，网络异常时保留上次缓存的 `userInfo`。

### 23. `client/src/api/index.ts` 中所有 API 函数无 TypeScript 类型定义
- **文件**: `client/src/api/index.ts:1-254`
- **描述**: 所有 API 函数返回 `Promise<T>`（实际上是 `Promise<any>`），没有任何参数类型或返回类型定义。IDE 无代码补全，重构时无编译检查。
- **建议**: 全部加上显式 interface：`interface LoginResponse { token: string; user: UserProfile }` 等。

### 24. 通知列表删除功能缺失
- **文件**: `client/src/pages/notification/list.vue`
- **描述**: `api/index.ts` 中定义了 `deleteNotification(id)` 函数，但 `notification/list.vue` 的模板中没有删除按钮，该 API 从未被调用。
- **建议**: 在通知项上增加左滑删除功能，或长按弹出删除选项。

### 25. `mine.vue` 使用 `onShow` 但未考虑退出登录后的导航
- **文件**: `client/src/pages/mine/mine.vue:107-117`
- **描述**: `onShow` 中检测 `if (!userStore.isLoggedIn)` 时跳转到登录页，这是一个无限循环隐患：如果用户从登录页退回，`onShow` 再次触发再次跳转。但 `navigateTo` 不会重复压栈，所以不算 CRITICAL，但逻辑上应使用 switchTab 或 reLaunch。
- **建议**: 使用 `uni.reLaunch({ url: '/pages/login/login' })` 替代 `uni.navigateTo`。

### 26. `checkin.vue` 视频上传的 `ext` 硬编码为 `mp4`
- **文件**: `client/src/pages/checkin/checkin.vue:65`
- **描述**: `/media/upload/presign?checkin_id=${checkin.id}&ext=mp4` 硬编码 `mp4`，如果用户选择 `mov` 或 `avi` 文件可能会失败。
- **建议**: 从文件路径中动态获取扩展名（如 `videoPath.value.split('.').pop()`），与 `resource/upload.vue:146` 的做法一致。

### 27. `manifest.json` 中微信小程序的 `appid` 为空
- **文件**: `client/src/manifest.json:9`
- **描述**: `"mp-weixin": { "appid": "" }`，未配置有效的微信小程序 AppID，无法在微信开发者工具和真机中运行。
- **建议**: 填入正确的微信小程序 AppID。

---

## [LOW] 问题列表

### 28. B 站等外部 emoji 直接硬编码在代码中
- **文件**: 多个文件的 template 和 script 中
- **描述**: emoji 符号（如 🔥 😊 🎬 🔔 ❤ 等）直接硬编码在模板字符串和 JavaScript 对象中，缺乏可维护性。如需更换图标或适配不同平台可能需逐一替换。
- **建议**: 统一使用 Unicode 码点或定义常量映射，方便后续维护。

### 29. `group.vue` 使用 `onMounted` 而非 `onShow`
- **文件**: `client/src/pages/group/group.vue:33`
- **描述**: 使用 `onMounted` 只会加载一次数据，从其他页面返回时不会刷新列表。其他列表页已使用 `onShow` 或支持下拉刷新来弥补，但 `group.vue` 仅依赖 `onMounted`。
- **建议**: 改为 `onShow` 确保每次进入页面刷新数据。

### 30. 资源标签管理页未实现跳转到标签筛选
- **文件**: `client/src/pages/resource/tags.vue`
- **描述**: `getResourceTags()` 获取标签列表，但点击标签没有跳转到对应的筛选资源列表功能（对比 `insight/tags.vue:28-30` 实现了 `goTagInsights` 导航）。
- **建议**: 为标签卡片添加点击事件，导航到 `resource/list` 并传入标签参数。

### 31. `any` 类型断言冗余样板代码
- **文件**: 全部页面
- **描述**: 每个 API 调用都要写 `const res: any = await ...` 和 `const x: any = await ...`，样板代码重复且无实际类型保护。
- **建议**: （同第 4 条）定义业务类型后移除全部 `: any` 断言。

### 32. CSS 中存在硬编码渐变重复
- **文件**: `index.vue:162`, `mine.vue:167`, `group/detail.vue:59`, `training/today.vue:103`, `insight/on-this-day.vue:63`, `resource/stats.vue:55`
- **描述**: 多处重复定义 `linear-gradient(135deg, #1cbbb4, #0081ff)`，颜色值变更时需多处修改。
- **建议**: 使用 CSS 变量或定义为公共 class。

---

## 分类汇总

| 严重级别 | 数量 | 主要分布 |
|----------|------|----------|
| CRITICAL | 6   | 类型安全、安全漏洞、无法部署 |
| HIGH     | 9   | 代码重复、空 catch、资源上传错误 |
| MEDIUM   | 11  | 瀑布流、v-for key、API 无类型、分页缺失 |
| LOW      | 5   | emoji 硬编码、onMounted 误用 |

## 核心改进建议（按优先级排序）

1. **立即修复**：头像上传发送 temp 路径（mine.vue）—— 该功能完全不工作
2. **立即修复**：硬编码 localhost 地址改为环境变量（request.ts）—— 影响生产部署和安全性
3. **短期**：为所有 API 定义 TypeScript interface，彻底消灭 `any`
4. **短期**：抽取 `usePagination` composable 消除 10+ 页面的重复分页代码
5. **短期**：所有 catch 块至少添加 `console.error`
6. **短期**：抽取共享常量（`typeMap`, `statusMap`, `moodIcon` 等）到 `constants/`
7. **中期**：Token 自动刷新机制
8. **中期**：将所有 `<image>` 添加 `lazy-load` 属性
9. **中期**：修复瀑布流布局，使用高度追踪算法替代奇偶分割
10. **长期**：建立完整的 uni-app 组件库和业务 composable 体系，从根本上消除代码重复