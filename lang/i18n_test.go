package lang

import (
	"os"
	"path/filepath"
	"testing"
)

func resetEnv() {
	os.Unsetenv("WN_LANG")
	loc = nil
}

func TestI18n(t *testing.T) {
	// 设置测试用的语言文件路径
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	LocalePath = filepath.Join(pwd, "..", "lang", "locales")

	// 每个测试前重置环境
	resetEnv()

	// 测试默认语言（英文）
	if msg := T("test message"); msg != "test message" {
		t.Errorf("Expected original message, got %s", msg)
	}

	// 测试简体中文
	resetEnv()
	os.Setenv("WN_LANG", "zh-CN")
	if msg := T("Pack files"); msg != "打包文件" {
		t.Errorf("Expected '打包文件', got %s", msg)
	}

	// 测试繁体中文
	resetEnv()
	os.Setenv("WN_LANG", "zh-TW")
	if msg := T("Pack files"); msg != "打包文件" {
		t.Errorf("Expected '打包文件', got %s", msg)
	}

	// 测试不存在的语言
	resetEnv()
	os.Setenv("WN_LANG", "fr")
	if msg := T("Pack files"); msg != "Pack files" {
		t.Errorf("Expected original message, got %s", msg)
	}

	// 测试不存在的翻译键
	resetEnv()
	os.Setenv("WN_LANG", "zh-CN")
	if msg := T("non-existent key"); msg != "non-existent key" {
		t.Errorf("Expected original message, got %s", msg)
	}

	// 测试语言代码别名
	tests := []struct {
		lang     string
		message  string
		expected string
	}{
		{"zh", "Pack files", "打包文件"},
		{"cn", "Pack files", "打包文件"},
		{"tw", "Pack files", "打包文件"},
	}

	for _, test := range tests {
		resetEnv()
		os.Setenv("WN_LANG", test.lang)
		if msg := T(test.message); msg != test.expected {
			t.Errorf("Lang %s: Expected %s, got %s", test.lang, test.expected, msg)
		}
	}
}
