package wnmcp

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sjzsdu/wn/share"
)

// MCPServerConfig 单个 MCP 服务器的配置
type MCPServerConfig struct {
	Disabled      bool     `json:"disabled"`
	Timeout       int      `json:"timeout"`
	Command       string   `json:"command"`
	Args          []string `json:"args"`
	Env           []string `json:"env"`
	TransportType string   `json:"transportType"`
	Url           string   `json:"url,omitempty"`
	AutoApprove   []string `json:"autoApprove,omitempty"`
}

// MCPConfig MCP 配置文件结构
type MCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// LoadMCPConfig 从指定目录加载 MCP 配置
func LoadMCPConfig(dir string) (*MCPConfig, error) {
	configPath := filepath.Join(dir, share.MCP_CONFIG_FILE)

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetServerConfig 获取指定服务器的配置
func (c *MCPConfig) GetServerConfig(name string) *MCPServerConfig {
	if c == nil {
		return nil
	}

	if config, exists := c.MCPServers[name]; exists && !config.Disabled {
		return &config
	}
	return nil
}
