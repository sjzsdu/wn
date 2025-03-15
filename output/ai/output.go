package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sjzsdu/wn/llm"
)

func Output(output string, messages []llm.Message) error {
	if output == "" || len(messages) == 0 {
		return nil
	}

	outputExt := strings.ToLower(filepath.Ext(output))
	var content []byte
	var err error

	// 格式化对话内容
	var conversations []Conversation
	for _, msg := range messages {
		if msg.Role == "user" || msg.Role == "assistant" {
			conversations = append(conversations, Conversation{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	switch outputExt {
	case ".pdf":
		content, err = toPDF(conversations)
	case ".md":
		content, err = toMarkdown(conversations)
	case ".xml":
		content, err = toXML(conversations)
	default:
		content, err = toText(conversations)
	}

	if err != nil {
		return fmt.Errorf("pack content error: %v", err)
	}

	// 确保输出目录存在
	if dir := filepath.Dir(output); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create output directory error: %v", err)
		}
	}

	// 写入文件
	if err := os.WriteFile(output, content, 0644); err != nil {
		return fmt.Errorf("write file error: %v", err)
	}

	return nil
}

type Conversation struct {
	Role    string
	Content string
}

func toPDF(conversations []Conversation) ([]byte, error) {
	// TODO: 实现 PDF 格式输出
	return nil, fmt.Errorf("PDF format not implemented yet")
}

func toMarkdown(conversations []Conversation) ([]byte, error) {
	var sb strings.Builder
	sb.WriteString("# AI 对话记录\n\n")

	for _, conv := range conversations {
		sb.WriteString(fmt.Sprintf("## %s\n\n",
			strings.Title(conv.Role)))
		sb.WriteString(conv.Content)
		sb.WriteString("\n\n---\n\n")
	}

	return []byte(sb.String()), nil
}

func toXML(conversations []Conversation) ([]byte, error) {
	var sb strings.Builder
	sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	sb.WriteString("<conversations>\n")

	for _, conv := range conversations {
		sb.WriteString(fmt.Sprintf("  <message role=\"%s\">\n",
			conv.Role))
		sb.WriteString(fmt.Sprintf("    <content><![CDATA[%s]]></content>\n", conv.Content))
		sb.WriteString("  </message>\n")
	}

	sb.WriteString("</conversations>")
	return []byte(sb.String()), nil
}

func toText(conversations []Conversation) ([]byte, error) {
	var sb strings.Builder

	for _, conv := range conversations {
		sb.WriteString(fmt.Sprintf("%s:\n",
			strings.Title(conv.Role)))
		sb.WriteString(conv.Content)
		sb.WriteString("\n\n")
	}

	return []byte(sb.String()), nil
}
