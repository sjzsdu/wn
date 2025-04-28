package deepseek

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletionRequestBody(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected DeepseekRequest
		wantErr  bool
	}{
		{
			name: "基本字段测试",
			input: `{
				"messages": [
					{
						"role": "system",
						"content": "你是一个AI助手"
					},
					{
						"role": "user",
						"content": "你好"
					}
				],
				"model": "deepseek-chat",
				"max_tokens": 8192
			}`,
			expected: DeepseekRequest{
				Messages: []Message{
					{
						Role:    "system",
						Content: "你是一个AI助手",
					},
					{
						Role:    "user",
						Content: "你好",
					},
				},
				Model:     "deepseek-chat",
				MaxTokens: 8192,
			},
			wantErr: false,
		},
		{
			name: "空消息测试",
			input: `{
				"messages": [],
				"model": "deepseek-chat",
				"max_tokens": 100
			}`,
			expected: DeepseekRequest{
				Messages:  []Message{},
				Model:     "deepseek-chat",
				MaxTokens: 100,
			},
			wantErr: false,
		},
		{
			name: "缺少必填字段测试",
			input: `{
				"max_tokens": 100
			}`,
			expected: DeepseekRequest{
				MaxTokens: 100,
			},
			wantErr: false, // Go的JSON解析不会验证必填字段
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got DeepseekRequest
			err := json.Unmarshal([]byte(tt.input), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Messages, got.Messages)
			assert.Equal(t, tt.expected.Model, got.Model)
			assert.Equal(t, tt.expected.MaxTokens, got.MaxTokens)
		})
	}
}
