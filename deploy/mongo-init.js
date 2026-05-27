// MongoDB 初始化脚本
// 创建集合并建立索引

db = db.getSiblingDB('wuxie');

// 用户集合
db.createCollection('users');
db.users.createIndex({ "openid": 1 }, { unique: true });
db.users.createIndex({ "group_id": 1 });
db.users.createIndex({ "score": -1 });

// 打卡记录集合
db.createCollection('checkins');
db.checkins.createIndex({ "user_id": 1, "created_at": -1 });
db.checkins.createIndex({ "status": 1 });
db.checkins.createIndex({ "created_at": -1 });

// 评论集合
db.createCollection('comments');
db.comments.createIndex({ "checkin_id": 1, "created_at": -1 });
db.comments.createIndex({ "user_id": 1 });

// 点赞集合
db.createCollection('likes');
db.likes.createIndex({ "checkin_id": 1, "user_id": 1 }, { unique: true });
db.likes.createIndex({ "user_id": 1 });

// 考核组集合
db.createCollection('groups');
db.groups.createIndex({ "leader_id": 1 });

// 排行榜缓存集合
db.createCollection('rank_cache');
db.rank_cache.createIndex({ "period": 1, "rank": 1 });
db.rank_cache.createIndex({ "user_id": 1, "period": 1 });

// 训练计划集合
db.createCollection('training_plans');
db.training_plans.createIndex({ "user_id": 1, "status": 1 });
db.training_plans.createIndex({ "user_id": 1, "created_at": -1 });
db.training_plans.createIndex({ "group_id": 1 });
db.training_plans.createIndex({ "start_date": 1, "end_date": 1 });

// 训练模板集合
db.createCollection('training_templates');
db.training_templates.createIndex({ "style": 1 });
db.training_templates.createIndex({ "category": 1 });
db.training_templates.createIndex({ "usage_count": -1 });

// 通知集合
db.createCollection('notifications');
db.notifications.createIndex({ "user_id": 1, "created_at": -1 });
db.notifications.createIndex({ "user_id": 1, "is_read": 1 });
db.notifications.createIndex({ "target_type": 1, "target_id": 1 });

// 通知设置集合
db.createCollection('notification_settings');
db.notification_settings.createIndex({ "user_id": 1 }, { unique: true });

// 感悟笔记集合
db.createCollection('insights');
db.insights.createIndex({ "user_id": 1, "created_at": -1 });
db.insights.createIndex({ "user_id": 1, "tags": 1 });
db.insights.createIndex({ "user_id": 1, "mood": 1 });
db.insights.createIndex({ "visibility": 1, "created_at": -1 });

// 感悟标签集合
db.createCollection('insight_tags');
db.insight_tags.createIndex({ "user_id": 1, "tag": 1 }, { unique: true });
db.insight_tags.createIndex({ "user_id": 1, "count": -1 });

// 感悟点赞集合
db.createCollection('insight_likes');
db.insight_likes.createIndex({ "insight_id": 1, "user_id": 1 }, { unique: true });

// 个人资料库集合
db.createCollection('resources');
db.resources.createIndex({ "user_id": 1, "created_at": -1 });
db.resources.createIndex({ "user_id": 1, "type": 1 });
db.resources.createIndex({ "user_id": 1, "is_favorite": 1 });
db.resources.createIndex({ "share_scope": 1, "created_at": -1 });
db.resources.createIndex({ "tags": 1 });

// 资料标签集合
db.createCollection('resource_tags');
db.resource_tags.createIndex({ "user_id": 1, "tag": 1 }, { unique: true });
db.resource_tags.createIndex({ "user_id": 1, "count": -1 });

print('Database initialized successfully');
