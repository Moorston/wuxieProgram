// Barrel re-export — 按领域拆分后的统一导出
// 保持向后兼容：所有现有 import { xxx } from '../api/index' 不需要改

// 认证
export { wxLogin, wxLogout, refreshToken } from './auth'

// 用户
export { getProfile, updateProfile, getPrivacySettings, updatePrivacySettings } from './user'

// 打卡
export {
  prepareCheckin,
  getCheckinList,
  getCheckinDetail,
  searchCheckinList,
  getMyCheckins,
  deleteCheckin,
  toggleLike,
  addComment,
  getComments,
} from './checkin'

// 排行榜 + 团组
export { getRankList, getMyRank, getGroupList, getGroupDetail, generateGroupInviteCode, joinGroupByInviteCode, createGroupAnnouncement, getGroupAnnouncements, deleteGroupAnnouncement, removeGroupMember, leaveGroup, setGroupLeader, getRankTrend } from './rank'

// 训练计划
export {
  createTrainingPlan,
  getTrainingPlan,
  listTrainingPlans,
  updateTrainingPlan,
  deleteTrainingPlan,
  getTodayTasks,
  updateTaskStatus,
  getTrainingReport,
  listTemplates,
  getTemplate,
  applyTemplate,
} from './training'

// 通知
export {
  getNotificationList,
  getUnreadCount,
  markNotificationRead,
  markAllNotificationsRead,
  deleteNotification,
  getNotificationSettings,
  updateNotificationSettings,
} from './notification'

// 感悟笔记
export {
  createInsight,
  getInsight,
  listInsights,
  listPublicInsights,
  updateInsight,
  deleteInsight,
  getInsightTags,
  getMoodStats,
  getOnThisDay,
  likeInsight,
} from './insight'

// 资源库
export {
  getResourcePresign,
  resourceUploadCallback,
  createResource,
  listResources,
  getResource,
  updateResource,
  deleteResource,
  toggleResourceFavorite,
  listResourceFavorites,
  getResourceTags,
  getResourceStats,
} from './resource'

// 数据分析
export { getCheckinHeatmap, getCheckinTrend, getAnalyticsOverview } from './analytics'

// 社交
export { followUser, unfollowUser, getFollowing, getFollowers, getFeed, getUserProfile } from './social'

// 赛事
export { getCompetitions, getCompetitionDetail, submitCompetitionEntry, getCompetitionEntries, getCompetitionRanking, scoreEntry } from './competition'

// 徽章
export { getAllBadges, getMyBadges } from './badge'

// 打卡挑战
export { createChallenge, getChallenges, getChallengeDetail, joinChallenge, getChallengeRanking } from './challenge'
