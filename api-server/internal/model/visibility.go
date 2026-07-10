package model

// Visibility 可见性类型
type Visibility int

const (
	VisibilityPublic  Visibility = 0 // 公开
	VisibilityGroup   Visibility = 1 // 仅团组可见
	VisibilityPrivate Visibility = 2 // 仅自己可见
)

// IsValid 检查可见性值是否有效
func (v Visibility) IsValid() bool {
	return v >= VisibilityPublic && v <= VisibilityPrivate
}

// String 返回可见性描述
func (v Visibility) String() string {
	switch v {
	case VisibilityPublic:
		return "public"
	case VisibilityGroup:
		return "group"
	case VisibilityPrivate:
		return "private"
	default:
		return "unknown"
	}
}
