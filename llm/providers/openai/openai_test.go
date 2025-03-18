package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		options map[string]interface{}
		wantErr bool
	}{
		{
			name: "基本配置",
			options: map[string]interface{}{
				"WN_OPENAI_APIKEY": "test-key",
			},
			wantErr: false,
		},
		{
			name: "缺少 API Key",
			options: map[string]interface{}{
				"WN_OPENAI_MODEL": "gpt-4",
			},
			wantErr: true,
		},
		{
			name: "自定义配置",
			options: map[string]interface{}{
				"WN_OPENAI_APIKEY":   "test-key",
				"WN_OPENAI_ENDPOINT": "https://custom.endpoint",
				"WN_OPENAI_MODEL":    "gpt-4",
				"WN_OPENAI_MODELS":   []string{"gpt-3", "gpt-4"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, p)
		})
	}
}

func TestParseStreamResponse(t *testing.T) {
	p := &Provider{}
	tests := []struct {
		name        string
		input       string
		wantContent string
		wantFinish  string
		wantErr     bool
	}{
		{
			name:        "正常响应",
			input:       `{"choices":[{"delta":{"content":"Hello"},"finish_reason":null}]}`,
			wantContent: "Hello",
			wantFinish:  "",
			wantErr:     false,
		},
		{
			name:        "完成响应",
			input:       `{"choices":[{"delta":{"content":""},"finish_reason":"stop"}]}`,
			wantContent: "",
			wantFinish:  "stop",
			wantErr:     false,
		},
		{
			name:        "无效 JSON",
			input:       `invalid json`,
			wantContent: "",
			wantFinish:  "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, finish, err := p.ParseStreamResponse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantContent, content)
			assert.Equal(t, tt.wantFinish, finish)
		})
	}
}
