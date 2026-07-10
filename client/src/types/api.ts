// API 响应类型

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 用户
export interface User {
  id: string
  nickname: string
  avatar: string
  gender: number
  province?: string
  city?: string
  group_id?: string
  score: number
  check_days: number
  created_at: string
  updated_at: string
}

// 打卡
export interface Checkin {
  id: string
  user_id: string
  video_url: string
  cover_url: string
  description: string
  duration: number
  score: number
  status: number
  like_count: number
  comment_count: number
  created_at: string
  updated_at: string
  user?: User
  is_liked: boolean
}

export interface CheckinListResponse {
  list: Checkin[]
  total: number
  page: number
  page_size: number
}

// 评论
export interface Comment {
  id: string
  checkin_id: string
  user_id: string
  content: string
  created_at: string
  user?: User
}

// 排行榜
export interface RankEntry {
  user_id: string
  score: number
  rank: number
}

// 团组
export interface Group {
  id: string
  name: string
  description: string
  member_count: number
  members?: User[]
}

// 训练计划
export interface TrainingPlan {
  id: string
  user_id: string
  group_id?: string
  title: string
  description: string
  days: TrainingDay[]
  status: number
  start_date: string
  end_date?: string
  created_at: string
}

export interface TrainingDay {
  title: string
  tasks: TrainingTask[]
}

export interface TrainingTask {
  name: string
  status: number
  checkin_id?: string
}

// 训练模板
export interface TrainingTemplate {
  id: string
  name: string
  description: string
  category: string
  style: string
  days: TrainingDay[]
  usage_count: number
}

// 通知
export interface Notification {
  id: string
  type: string
  title: string
  content: string
  target_type: string
  target_id: string
  is_read: boolean
  created_at: string
}

export interface NotificationSettings {
  like_enabled: boolean
  comment_enabled: boolean
  plan_remind_enabled: boolean
}

// 感悟笔记
export interface Insight {
  id: string
  user_id: string
  content: string
  images: string[]
  mood: string
  tags: string[]
  is_public: boolean
  like_count: number
  created_at: string
  user?: User
  is_liked: boolean
}

export interface MoodStats {
  [mood: string]: number
}

// 资源
export interface Resource {
  id: string
  user_id: string
  title: string
  type: string
  category: string
  difficulty: string
  url: string
  description: string
  tags: string[]
  file_size: number
  share_scope: string
  view_count: number
  download_count: number
  is_favorited: boolean
  created_at: string
}

export interface ResourceStats {
  total_count: number
  total_size: number
  category_count: Record<string, number>
}

// 分页通用
export interface PaginatedResponse<T> {
  list: T[]
  total: number
}

// 登录响应
export interface LoginResponse {
  token: string
  refresh_token: string
  user: User
}

// 刷新 Token 响应
export interface RefreshTokenResponse {
  token: string
  refresh_token: string
}