package agent

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
)

const (
	AGENT_EXT = ".md"
)

//go:embed agents/*.md
var embeddedAgents embed.FS

var (
	systemAgents = make(map[string]string) // 系统级别的agents
	userAgents   = make(map[string]string) // 用户级别的agents
)

func init() {
	_, userDir := getAgentDirs()

	// 创建用户目录
	if err := os.MkdirAll(userDir, 0755); err != nil {
		fmt.Printf("Warning - Failed to create user dir: %v\n", err)
	}

	// 初始化系统agents（从embed文件系统读取）
	loadSystemAgents()
	// 初始化用户agents（从用户目录读取）
	loadUserAgents(userDir)
}

// 从embed文件系统加载系统agents
func loadSystemAgents() {
	entries, err := embeddedAgents.ReadDir("agents")
	if err != nil {
		fmt.Printf("Warning - Failed to read embedded agents: %v\n", err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), AGENT_EXT) {
			name := strings.TrimSuffix(entry.Name(), AGENT_EXT)
			content, err := embeddedAgents.ReadFile(filepath.Join("agents", entry.Name()))
			if err == nil {
				systemAgents[name] = string(content)
			}
		}
	}
}

// 从用户目录加载用户agents
func loadUserAgents(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Warning - Failed to read dir %s: %v\n", dir, err)
		}
		return
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), AGENT_EXT) {
			name := strings.TrimSuffix(f.Name(), AGENT_EXT)
			content, err := os.ReadFile(filepath.Join(dir, f.Name()))
			if err == nil {
				userAgents[name] = string(content)
			} else {
				fmt.Printf("Warning - Failed to read file %s: %v\n", f.Name(), err)
			}
		}
	}
}

// 可以删除原来的 loadAgentsFromDir 函数
// ListAgents 列出所有代理
func ListAgents() {
	var output strings.Builder
	output.WriteString(lang.T("System Agents:"))
	if agents := listSystemAgents(); len(agents) > 0 {
		output.WriteString("\n" + strings.Join(agents, "\n"))
	}
	output.WriteString("\n" + lang.T("User Agents:"))
	if agents := listUserAgents(); len(agents) > 0 {
		output.WriteString("\n" + strings.Join(agents, "\n"))
	}
	output.WriteString("\n")
	fmt.Print(output.String())
}

func listSystemAgents() []string {
	var agents []string
	entries, err := embeddedAgents.ReadDir("agents")
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), AGENT_EXT) {
				agents = append(agents, "- "+strings.TrimSuffix(entry.Name(), AGENT_EXT))
			}
		}
	}
	return agents
}

func listUserAgents() []string {
	var agents []string
	_, userDir := getAgentDirs()
	files, err := os.ReadDir(userDir)
	if err != nil {
		return agents
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), AGENT_EXT) {
			agents = append(agents, "- "+strings.TrimSuffix(f.Name(), AGENT_EXT))
		}
	}
	return agents
}

// CreateNewAgent 创建新代理
func CreateNewAgent(name, content string) {
	if _, exists := systemAgents[name]; exists {
		fmt.Println(lang.T("Agent already exists in system agents"))
		return
	}
	if _, exists := userAgents[name]; exists {
		fmt.Println(lang.T("Agent already exists in user agents"))
		return
	}

	if content == "" {
		stdinContent, err := os.ReadFile(os.Stdin.Name())
		if err != nil {
			fmt.Printf(lang.T("Failed to read input: %v\n"), err)
			return
		}
		content = string(stdinContent)
	}

	// 保存到内存
	userAgents[name] = content

	// 保存到文件
	_, userDir := getAgentDirs()
	os.MkdirAll(userDir, 0755)
	err := os.WriteFile(filepath.Join(userDir, name+AGENT_EXT), []byte(content), 0644)
	if err != nil {
		fmt.Printf(lang.T("Failed to create agent: %v\n"), err)
		delete(userAgents, name)
		return
	}
	fmt.Println(lang.T("Agent created successfully"))
}

// UpdateExistingAgent 更新现有代理
func UpdateExistingAgent(name, content string) {
	if _, exists := systemAgents[name]; exists {
		fmt.Println(lang.T("Cannot update system agent"))
		return
	}
	if _, exists := userAgents[name]; !exists {
		fmt.Println(lang.T("Agent not found"))
		return
	}

	if content == "" {
		stdinContent, err := os.ReadFile(os.Stdin.Name())
		if err != nil {
			fmt.Printf(lang.T("Failed to read input: %v\n"), err)
			return
		}
		content = string(stdinContent)
	}

	// 更新内存
	userAgents[name] = content

	// 更新文件
	_, userDir := getAgentDirs()
	err := os.WriteFile(filepath.Join(userDir, name+AGENT_EXT), []byte(content), 0644)
	if err != nil {
		fmt.Printf(lang.T("Failed to update agent: %v\n"), err)
		return
	}
	fmt.Println(lang.T("Agent updated successfully"))
}

// DeleteExistingAgent 删除现有代理
func DeleteExistingAgent(name string) {

	if _, exists := userAgents[name]; !exists {
		fmt.Println(lang.T("Agent not found"))
		return
	}

	// 从内存中删除
	delete(userAgents, name)

	// 删除文件
	_, userDir := getAgentDirs()
	err := os.Remove(filepath.Join(userDir, name+AGENT_EXT))
	if err != nil {
		fmt.Printf(lang.T("Failed to delete agent file: %v\n"), err)
		return
	}
	fmt.Println(lang.T("Agent deleted successfully"))
}

// GetAgentContent 获取代理内容，优先从用户目录获取
func GetAgentContent(name string) string {
	// 优先检查用户代理
	if content, exists := userAgents[name]; exists {
		return content
	}
	// 找不到再检查系统代理
	if content, exists := systemAgents[name]; exists {
		return content
	}
	return ""
}

// ShowAgentContent 显示代理内容
func ShowAgentContent(name string) string {
	content := GetAgentContent(name)
	if content == "" {
		fmt.Printf(lang.T("Agent not found: %s\n"), name)
	}
	return content
}

func getAgentDirs() (string, string) {
	userDir := helper.GetPath("agents")
	return "agents", userDir
}

// SaveAgent 创建或更新代理
func SaveAgent(name, content string) {
	if content == "" {
		stdinContent, err := os.ReadFile(os.Stdin.Name())
		if err != nil {
			fmt.Printf(lang.T("Failed to read input: %v\n"), err)
			return
		}
		content = string(stdinContent)
	}

	// 检查是否是系统代理
	if _, exists := systemAgents[name]; exists {
		fmt.Println(lang.T("Cannot modify system agent"))
		return
	}

	// 保存到内存
	userAgents[name] = content

	// 保存到文件
	_, userDir := getAgentDirs()
	os.MkdirAll(userDir, 0755)
	err := os.WriteFile(filepath.Join(userDir, name+AGENT_EXT), []byte(content), 0644)
	if err != nil {
		fmt.Printf(lang.T("Failed to save agent: %v\n"), err)
		delete(userAgents, name)
		return
	}

	if _, existed := userAgents[name]; existed {
		fmt.Println(lang.T("Agent updated successfully"))
	} else {
		fmt.Println(lang.T("Agent created successfully"))
	}
}

// 删除 CreateNewAgent 和 UpdateExistingAgent 函数
