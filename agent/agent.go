package agent

import (
	"strings"

	"github.com/sjzsdu/wn/config"
	"github.com/sjzsdu/wn/llm"
)

const DEFAULT_AGENT = "fullstack"

// 添加语言映射表
var languageMap = map[string]string{
	"zh":      "你需要用中文语言回复。",
	"cn":      "你需要用中文语言回复。",
	"zh-CN":   "你需要用中文语言回复。",
	"en":      "Please respond in English.",
	"english": "Please respond in English.",
	"jp":      "日本語で返信してください。",
	"ja":      "日本語で返信してください。",
	"kr":      "한국어로 응답해 주세요.",
	"ko":      "한국어로 응답해 주세요.",
	"fr":      "Veuillez répondre en français.",
	"de":      "Bitte antworten Sie auf Deutsch.",
	"es":      "Por favor, responda en español.",
	"ru":      "Пожалуйста, ответьте на русском языке.",
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
		if promptText, ok := languageMap[strings.ToLower(lang)]; ok {
			messages = append(messages, llm.Message{
				Role:    "system",
				Content: promptText,
			})
		}
	}

	return messages
}
