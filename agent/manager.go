package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
)

const (
	systemAgentsDir = "assets/agents"
	AGENT_EXT       = ".md"
)

var (
	systemAgents = make(map[string]string) // 系统级别的agents
	userAgents   = make(map[string]string) // 用户级别的agents
)

func init() {
	systemDir, userDir := getAgentDirs()

	// 创建必要的目录
	if err := os.MkdirAll(systemDir, 0755); err != nil {
		fmt.Printf("Warning - Failed to create system dir: %v\n", err)
	}
	if err := os.MkdirAll(userDir, 0755); err != nil {
		fmt.Printf("Warning - Failed to create user dir: %v\n", err)
	}

	// 初始化系统agents
	loadAgentsFromDir(systemDir, systemAgents)
	// 初始化用户agents
	loadAgentsFromDir(userDir, userAgents)
}

func loadAgentsFromDir(dir string, agents map[string]string) {
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
				agents[name] = string(content)
			} else {
				fmt.Printf("Warning - Failed to read file %s: %v\n", f.Name(), err)
			}
		}
	}
}

// ListAgents 列出所有代理
func ListAgents() {
	systemDir, userDir := getAgentDirs()

	var output strings.Builder
	output.WriteString(lang.T("System Agents:"))
	if agents := listAgentsInDir(systemDir); len(agents) > 0 {
		output.WriteString("\n" + strings.Join(agents, "\n"))
	}
	output.WriteString("\n" + lang.T("User Agents:"))
	if agents := listAgentsInDir(userDir); len(agents) > 0 {
		output.WriteString("\n" + strings.Join(agents, "\n"))
	}
	output.WriteString("\n")

	fmt.Print(output.String())
}

func listAgentsInDir(dir string) []string {
	var agents []string
	files, err := os.ReadDir(dir)
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
	if _, exists := systemAgents[name]; exists {
		fmt.Println(lang.T("Cannot delete system agent"))
		return
	}
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

// ShowAgentContent 显示代理内容
func ShowAgentContent(name string) {
	if content, exists := userAgents[name]; exists {
		fmt.Println(content)
		return
	}
	if content, exists := systemAgents[name]; exists {
		fmt.Println(content)
		return
	}
	fmt.Printf(lang.T("Agent not found: %s\n"), name)
}

func getAgentDirs() (string, string) {
	userDir := helper.GetPath("agents")

	// 获取项目根目录下的 assets/agents 目录
	projectDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Warning - Failed to get current directory: %v\n", err)
		projectDir = filepath.Dir(os.Args[0])
	}

	systemDir := filepath.Join(projectDir, systemAgentsDir)

	return systemDir, userDir
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
