package agent

import (
	"github.com/sjzsdu/wn/config"
	"github.com/sjzsdu/wn/llm"
)

const DEFAULT_AGENT = "fullstack"

// GetAgentMessages 返回预设的 agent 系统消息
func GetAgentMessages(name string) []llm.Message {

	if name == "" {
		name = config.GetConfig("default_agent")
		if name == "" {
			name = DEFAULT_AGENT
		}
	}
	messages := make([]llm.Message, 0)
	content := ShowAgentContent(name)
	// 添加系统角色消息，定义 agent 的行为和能力
	messages = append(messages, llm.Message{
		Role:    "system",
		Content: content,
	})

	if config.GetConfig("lang") != "" && name != "translate" {
		lang := config.GetConfig("lang")
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: "无论用户使用什么语言提问，你必须始终使用" + lang + "语言回复。",
		})
	}

	return messages
}
