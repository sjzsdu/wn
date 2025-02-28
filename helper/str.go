package helper

// StringSliceContains 检查切片中是否包含指定的字符串
func StringSliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
