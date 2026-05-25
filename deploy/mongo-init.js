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

print('Database initialized successfully');
