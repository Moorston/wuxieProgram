// 训练类型映射
export const TRAINING_TYPE_MAP: Record<string, string> = {
  basic: '基本功',
  taolu: '套路',
  sanda: '散打',
  qigong: '气功',
}

// 训练状态映射
export const TRAINING_STATUS_MAP: Record<string, string> = {
  active: '进行中',
  completed: '已完成',
  draft: '草稿',
  cancelled: '已取消',
}

// 任务状态映射
export const TASK_STATUS_MAP: Record<number, string> = {
  0: '待完成',
  1: '已完成',
  2: '已跳过',
}

// 资源类型映射
export const RESOURCE_TYPE_MAP: Record<string, string> = {
  video: '视频',
  image: '图片',
  document: '文档',
}

// 资源分类映射
export const RESOURCE_CATEGORY_MAP: Record<string, string> = {
  basic: '基本功',
  taolu: '套路',
  sanda: '散打',
  qigong: '气功',
  theory: '理论',
  other: '其他',
}

// 难度映射
export const DIFFICULTY_MAP: Record<string, string> = {
  beginner: '入门',
  intermediate: '进阶',
  advanced: '高级',
}

// 心情图标映射
export const MOOD_ICON_MAP: Record<string, string> = {
  breakthrough: '🔥',
  satisfied: '😊',
  neutral: '😐',
  confused: '🤔',
  down: '😔',
}

// 心情文字映射
export const MOOD_TEXT_MAP: Record<string, string> = {
  breakthrough: '突破',
  satisfied: '满意',
  neutral: '一般',
  confused: '困惑',
  down: '低落',
}

// 分享范围映射
export const SHARE_SCOPE_MAP: Record<string, string> = {
  private: '仅自己',
  group: '考核组',
  public: '公开',
}

// 分页默认值
export const PAGE_SIZE_DEFAULT = 20
export const PAGE_SIZE_MAX = 50

// 文本长度限制
export const MAX_DESCRIPTION_LEN = 200
export const MAX_COMMENT_LEN = 500
export const MAX_INSIGHT_LEN = 2000
export const MAX_SEARCH_KEYWORD = 50