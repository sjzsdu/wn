package helper

import (
	"math/rand"
	"regexp"
	"strings"
)

// StringSliceContains 检查切片中是否包含指定的字符串
func StringSliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 添加新的辅助函数来清理 ANSI 转义序列
// 修改 stripAnsiCodes 函数，确保正确处理 git diff 输出
func StripAnsiCodes(s string) string {
	// 处理 git diff 常见的颜色代码和格式控制符
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*[mGKHF]`)
	return strings.TrimSpace(ansi.ReplaceAllString(s, ""))
}

// randomString 生成指定长度的随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
