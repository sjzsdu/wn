package agent

import (
	"strings"

	"github.com/sjzsdu/wn/config"
	"github.com/sjzsdu/wn/llm"
)

const DEFAULT_AGENT = "fullstack"

// 添加语言映射表
var languageMap = map[string]string{
	"zh":      "中文",
	"cn":      "中文",
	"zh-CN":   "中文",
	"en":      "英文",
	"english": "英文",
	"jp":      "日文",
	"ja":      "日文",
	"kr":      "韩文",
	"ko":      "韩文",
	"fr":      "法文",
	"de":      "德文",
	"es":      "西班牙文",
	"ru":      "俄文",
}

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
		if displayLang, ok := languageMap[strings.ToLower(lang)]; ok {
			lang = displayLang
		}
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: "你需要用" + lang + "语言回复。",
		})
	}

	return messages
}
