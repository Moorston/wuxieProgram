package service

// extractTags 从interface{}中提取[]string标签列表
// 兼容JSON反序列化后的[]interface{}类型
func extractTags(v interface{}) []string {
	if v == nil {
		return nil
	}

	switch t := v.(type) {
	case []string:
		return t
	case []interface{}:
		tags := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok {
				tags = append(tags, s)
			}
		}
		return tags
	default:
		return nil
	}
}