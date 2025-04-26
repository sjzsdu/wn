package wnmcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadMCPConfig(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 测试用例
	tests := []struct {
		name        string
		setup       func() string
		expectError bool
		expectConfig *MCPConfig
	}{
		{
			name: "成功加载配置文件",
			setup: func() string {
				config := MCPConfig{
					MCPServers: map[string]MCPServerConfig{
						"server1": {
							Disabled: false,
							Timeout:  10,
						},
					},
				}
				data, _ := json.Marshal(config)
				configPath := filepath.Join(tempDir, "mcp.json")
				os.WriteFile(configPath, data, 0644)
				return configPath
			},
			expectError: false,
			expectConfig: &MCPConfig{
				MCPServers: map[string]MCPServerConfig{
					"server1": {
						Disabled: false,
						Timeout:  10,
					},
				},
			},
		},
		{
			name: "配置文件不存在",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent.json")
			},
			expectError: true,
		},
		{
			name: "无效的JSON格式",
			setup: func() string {
				configPath := filepath.Join(tempDir, "invalid.json")
				os.WriteFile(configPath, []byte("{invalid json}"), 0644)
				return configPath
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setup()
			config, err := LoadMCPConfig(tempDir, configPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				if tt.expectConfig != nil {
					assert.Equal(t, tt.expectConfig, config)
				} else {
					assert.Nil(t, config)
				}
			}
		})
	}
}

func TestGetServerConfig(t *testing.T) {
	// 测试用例
	tests := []struct {
		name         string
		config       *MCPConfig
		serverName   string
		expectConfig *MCPServerConfig
	}{
		{
			name: "获取存在的服务器配置",
			config: &MCPConfig{
				MCPServers: map[string]MCPServerConfig{
					"server1": {
						Disabled: false,
						Timeout:  10,
					},
				},
			},
			serverName: "server1",
			expectConfig: &MCPServerConfig{
				Disabled: false,
				Timeout:  10,
			},
		},
		{
			name: "获取不存在的服务器配置",
			config: &MCPConfig{
				MCPServers: map[string]MCPServerConfig{
					"server1": {
						Disabled: false,
						Timeout:  10,
					},
				},
			},
			serverName:   "server2",
			expectConfig: nil,
		},
		{
			name: "获取已禁用的服务器配置",
			config: &MCPConfig{
				MCPServers: map[string]MCPServerConfig{
					"server1": {
						Disabled: true,
						Timeout:  10,
					},
				},
			},
			serverName:   "server1",
			expectConfig: nil,
		},
		{
			name:         "空配置",
			config:       nil,
			serverName:   "server1",
			expectConfig: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config.GetServerConfig(tt.serverName)
			assert.Equal(t, tt.expectConfig, config)
		})
	}
}
