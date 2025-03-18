package ai

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sjzsdu/wn/llm"
	"github.com/stretchr/testify/assert"
)

func TestOutput(t *testing.T) {
	tempDir := t.TempDir()
	testMessages := []llm.Message{
		{Role: "user", Content: "你好"},
		{Role: "assistant", Content: "你好！有什么我可以帮你的吗？"},
		{Role: "user", Content: "再见"},
		{Role: "assistant", Content: "再见！祝你有愉快的一天！"},
	}

	tests := []struct {
		name        string
		output      string
		messages    []llm.Message
		wantErr     bool
		checkOutput func(t *testing.T, content []byte)
	}{
		{
			name:     "空输出路径",
			output:   "",
			messages: testMessages,
			wantErr:  false,
		},
		{
			name:     "空消息列表",
			output:   filepath.Join(tempDir, "empty.txt"),
			messages: nil,
			wantErr:  false,
		},
		{
			name:     "文本格式输出",
			output:   filepath.Join(tempDir, "test.txt"),
			messages: testMessages,
			wantErr:  false,
			checkOutput: func(t *testing.T, content []byte) {
				assert.Contains(t, string(content), "User:")
				assert.Contains(t, string(content), "Assistant:")
				assert.Contains(t, string(content), "你好")
				assert.Contains(t, string(content), "再见")
			},
		},
		{
			name:     "Markdown格式输出",
			output:   filepath.Join(tempDir, "test.md"),
			messages: testMessages,
			wantErr:  false,
			checkOutput: func(t *testing.T, content []byte) {
				assert.Contains(t, string(content), "# AI 对话记录")
				assert.Contains(t, string(content), "## User")
				assert.Contains(t, string(content), "## Assistant")
				assert.Contains(t, string(content), "---")
			},
		},
		{
			name:     "XML格式输出",
			output:   filepath.Join(tempDir, "test.xml"),
			messages: testMessages,
			wantErr:  false,
			checkOutput: func(t *testing.T, content []byte) {
				assert.Contains(t, string(content), "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
				assert.Contains(t, string(content), "<conversations>")
				assert.Contains(t, string(content), "<message role=\"user\">")
				assert.Contains(t, string(content), "<message role=\"assistant\">")
				assert.Contains(t, string(content), "]]></content>")
			},
		},
		{
			name:     "PDF格式输出",
			output:   filepath.Join(tempDir, "test.pdf"),
			messages: testMessages,
			wantErr:  true, // 因为 PDF 功能还未实现
		},
		{
			name:     "创建嵌套目录",
			output:   filepath.Join(tempDir, "nested", "dir", "test.txt"),
			messages: testMessages,
			wantErr:  false,
			checkOutput: func(t *testing.T, content []byte) {
				assert.Contains(t, string(content), "User:")
				assert.Contains(t, string(content), "Assistant:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Output(tt.output, tt.messages)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			if tt.output != "" && tt.checkOutput != nil {
				content, err := os.ReadFile(tt.output)
				assert.NoError(t, err)
				tt.checkOutput(t, content)
			}
		})
	}
}

func TestConversionFunctions(t *testing.T) {
	conversations := []Conversation{
		{Role: "user", Content: "测试内容1"},
		{Role: "assistant", Content: "测试回复1"},
	}

	t.Run("toText", func(t *testing.T) {
		content, err := toText(conversations)
		assert.NoError(t, err)
		text := string(content)
		assert.Contains(t, text, "User:")
		assert.Contains(t, text, "测试内容1")
		assert.Contains(t, text, "Assistant:")
		assert.Contains(t, text, "测试回复1")
	})

	t.Run("toMarkdown", func(t *testing.T) {
		content, err := toMarkdown(conversations)
		assert.NoError(t, err)
		md := string(content)
		assert.Contains(t, md, "# AI 对话记录")
		assert.Contains(t, md, "## User")
		assert.Contains(t, md, "## Assistant")
		assert.Contains(t, md, "---")
	})

	t.Run("toXML", func(t *testing.T) {
		content, err := toXML(conversations)
		assert.NoError(t, err)
		xml := string(content)
		assert.Contains(t, xml, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
		assert.Contains(t, xml, "<conversations>")
		assert.Contains(t, xml, "<message role=\"user\">")
		assert.Contains(t, xml, "<![CDATA[测试内容1]]>")
	})

	t.Run("toPDF", func(t *testing.T) {
		_, err := toPDF(conversations)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PDF format not implemented yet")
	})
}
