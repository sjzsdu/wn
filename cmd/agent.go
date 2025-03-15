package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sjzsdu/wn/agent"
	"github.com/sjzsdu/wn/lang"
	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: lang.T("Agent management"),
	Long:  lang.T("Agent management"),
	Run:   runAgent,
}

var (
	listAgent   bool
	createAgent string // 改为统一的创建/更新操作
	deleteAgent string
	showAgent   string
	content     string
	contentFile string
)

func init() {
	agentCmd.Flags().BoolVar(&listAgent, "list", false, lang.T("List all agents"))
	agentCmd.Flags().StringVar(&createAgent, "create", "", lang.T("Create or update an agent")) // 更新描述
	agentCmd.Flags().StringVar(&deleteAgent, "delete", "", lang.T("Delete an agent"))
	agentCmd.Flags().StringVar(&showAgent, "show", "", lang.T("Show agent content"))
	agentCmd.Flags().StringVar(&content, "content", "", lang.T("Agent content"))
	agentCmd.Flags().StringVar(&contentFile, "file", "", lang.T("Read content from file"))
	rootCmd.AddCommand(agentCmd)
}

func runAgent(cmd *cobra.Command, args []string) {
	if listAgent {
		agent.ListAgents()
		return
	}

	// 获取内容的函数
	getContent := func() string {
		// 如果指定了文件，从文件读取
		if contentFile != "" {
			data, err := os.ReadFile(contentFile)
			if err != nil {
				fmt.Printf(lang.T("Failed to read file: %v\n"), err)
				return ""
			}
			return string(data)
		}

		// 如果指定了内容，直接使用
		if content != "" {
			return content
		}

		// 从标准输入读取，使用空行作为结束标记
		fmt.Println(lang.T("Please input content, input an empty line to finish:"))
		var lines []string
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				break
			}
			lines = append(lines, line)
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf(lang.T("Failed to read input: %v\n"), err)
			return ""
		}
		return strings.Join(lines, "\n")
	}

	if createAgent != "" {
		agent.SaveAgent(createAgent, getContent()) // 使用新的统一方法
		return
	}

	if deleteAgent != "" {
		agent.DeleteExistingAgent(deleteAgent)
		return
	}

	if showAgent != "" {
		content := agent.ShowAgentContent(showAgent)
		fmt.Printf(content)
		return
	}

	cmd.Help()
}
