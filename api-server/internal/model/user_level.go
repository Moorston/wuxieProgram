package model

// UserLevel 用户等级
type UserLevel struct {
	Level       int    `json:"level"`       // 等级数字
	Name        string `json:"name"`        // 等级名称
	Icon        string `json:"icon"`        // 等级图标
	MinDays     int    `json:"min_days"`    // 最低打卡天数
	Description string `json:"description"` // 描述
}

// 预定义等级
var UserLevels = []UserLevel{
	{Level: 0, Name: "新手", Icon: "🌱", MinDays: 0, Description: "刚开始武术之旅"},
	{Level: 1, Name: "武者", Icon: "🥋", MinDays: 7, Description: "坚持训练一周"},
	{Level: 2, Name: "武师", Icon: "⚔️", MinDays: 30, Description: "坚持训练一个月"},
	{Level: 3, Name: "武道", Icon: "🐉", MinDays: 90, Description: "坚持训练三个月"},
	{Level: 4, Name: "宗师", Icon: "👑", MinDays: 180, Description: "坚持训练半年"},
	{Level: 5, Name: "武圣", Icon: "🏆", MinDays: 365, Description: "坚持训练一年"},
}

// GetUserLevel 根据打卡天数获取用户等级
func GetUserLevel(checkDays int) UserLevel {
	result := UserLevels[0]
	for _, level := range UserLevels {
		if checkDays >= level.MinDays {
			result = level
		}
	}
	return result
}

// GetNextLevel 获取下一个等级（用于显示进度）
func GetNextLevel(checkDays int) *UserLevel {
	for _, level := range UserLevels {
		if checkDays < level.MinDays {
			return &level
		}
	}
	return nil // 已满级
}
