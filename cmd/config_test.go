package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfig(t *testing.T) {
	// 设置临时测试目录
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// 设置测试环境变量
	os.Setenv("WN_LANG", "zh")
	os.Setenv("WN_DEEPSEEK_APIKEY", "test-key")

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "使用简短键获取语言",
			key:      "lang",
			expected: "zh",
		},
		{
			name:     "使用环境变量键获取语言",
			key:      "WN_LANG",
			expected: "zh",
		},
		{
			name:     "使用简短键获取API密钥",
			key:      "deepseek_apikey",
			expected: "test-key",
		},
		{
			name:     "使用环境变量键获取API密钥",
			key:      "WN_DEEPSEEK_APIKEY",
			expected: "test-key",
		},
		{
			name:     "获取不存在的配置",
			key:      "nonexistent",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConfig(tt.key); got != tt.expected {
				t.Errorf("GetConfig() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfigOperations(t *testing.T) {
	// 设置临时测试目录
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	tests := []struct {
		name     string
		configs  map[string]string
		wantErr  bool
		validate func(t *testing.T, configs map[string]string)
	}{
		{
			name: "基本配置保存和加载",
			configs: map[string]string{
				"WN_LANG":            "zh",
				"WN_DEEPSEEK_APIKEY": "test-key",
			},
			wantErr: false,
			validate: func(t *testing.T, configs map[string]string) {
				if len(configs) != 2 {
					t.Errorf("期望配置数量为 2，实际为 %d", len(configs))
				}
				// 使用 GetConfig 验证
				if v := GetConfig("lang"); v != "zh" {
					t.Errorf("lang 期望为 zh，实际为 %s", v)
				}
				if v := GetConfig("deepseek_apikey"); v != "test-key" {
					t.Errorf("deepseek_apikey 期望为 test-key，实际为 %s", v)
				}
			},
		},
		{
			name:    "空配置",
			configs: map[string]string{},
			wantErr: false,
			validate: func(t *testing.T, configs map[string]string) {
				if len(configs) != 0 {
					t.Errorf("期望配置为空，实际数量为 %d", len(configs))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置 configMap
			configMap = make(map[string]string)

			// 复制测试配置到 configMap
			for k, v := range tt.configs {
				configMap[k] = v
			}

			// 测试保存
			if err := saveConfig(); (err != nil) != tt.wantErr {
				t.Errorf("saveConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 清空 configMap
			configMap = make(map[string]string)

			// 测试加载
			if err := loadConfig(); (err != nil) != tt.wantErr {
				t.Errorf("loadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 验证结果
			tt.validate(t, configMap)

			// 验证文件内容
			if !tt.wantErr {
				content, err := os.ReadFile(filepath.Join(tmpDir, ".wn", "config"))
				if err != nil {
					t.Errorf("读取配置文件失败: %v", err)
				}
				if len(content) == 0 && len(tt.configs) > 0 {
					t.Error("配置文件不应为空")
				}
			}
		})
	}
}
