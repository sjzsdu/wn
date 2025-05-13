package deepseek

import (
	"context"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/llm/providers/base"
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
				"WN_DEEPSEEK_APIKEY": "test-key",
			},
			wantErr: false,
		},
		{
			name: "缺少 API Key",
			options: map[string]interface{}{
				"WN_DEEPSEEK_MODEL": "deepseek-chat",
			},
			wantErr: true,
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

func TestHandleTools(t *testing.T) {
	p := &Provider{}
	tests := []struct {
		name     string
		tools    []mcp.Tool
		expected []Tool
	}{
		{
			name: "单路径工具",
			tools: []mcp.Tool{
				{
					Name:        "path_tool",
					Description: "测试单路径工具",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"path": map[string]interface{}{
								"type": "string",
							},
						},
						Required: []string{"path"},
					},
				},
			},
			expected: []Tool{
				{
					Type: "function",
					Function: Function{
						Name:        "path_tool",
						Description: "测试单路径工具",
						Parameters: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"path": map[string]interface{}{
									"type": "string",
								},
							},
							"required": []string{"path"},
						},
					},
				},
			},
		},
		{
			name: "多路径工具",
			tools: []mcp.Tool{
				{
					Name:        "paths_tool",
					Description: "测试多路径工具",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"paths": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
						},
						Required: []string{"paths"},
					},
				},
			},
			expected: []Tool{
				{
					Type: "function",
					Function: Function{
						Name:        "paths_tool",
						Description: "测试多路径工具",
						Parameters: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"paths": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"type": "string",
									},
								},
							},
							"required": []string{"paths"},
						},
					},
				},
			},
		},
		{
			name: "编辑工具",
			tools: []mcp.Tool{
				{
					Name:        "edit_tool",
					Description: "测试编辑工具",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"dryRun": map[string]interface{}{
								"type":        "boolean",
								"default":     false,
								"description": "Preview changes using git-style diff format",
							},
							"edits": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"oldText": map[string]interface{}{
											"type":        "string",
											"description": "Text to search for - must match exactly",
										},
										"newText": map[string]interface{}{
											"type":        "string",
											"description": "Text to replace with",
										},
									},
									"required":             []string{"oldText", "newText"},
									"additionalProperties": false,
								},
							},
							"path": map[string]interface{}{
								"type": "string",
							},
						},
						Required: []string{"path", "edits"},
					},
				},
			},
			expected: []Tool{
				{
					Type: "function",
					Function: Function{
						Name:        "edit_tool",
						Description: "测试编辑工具",
						Parameters: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"dryRun": map[string]interface{}{
									"type":        "boolean",
									"default":     false,
									"description": "Preview changes using git-style diff format",
								},
								"edits": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"oldText": map[string]interface{}{
												"type":        "string",
												"description": "Text to search for - must match exactly",
											},
											"newText": map[string]interface{}{
												"type":        "string",
												"description": "Text to replace with",
											},
										},
										"required":             []string{"oldText", "newText"},
										"additionalProperties": false,
									},
								},
								"path": map[string]interface{}{
									"type": "string",
								},
							},
							"required": []string{"path", "edits"},
						},
					},
				},
			},
		},
		{
			name:     "空工具列表",
			tools:    []mcp.Tool{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.handleTools(tt.tools)
			assert.Equal(t, tt.expected, result, "工具转换结果不匹配")
		})
	}
}

// 添加一个模拟的 HTTP 处理器
type mockHTTPHandler struct {
	Response []byte
}

func (h *mockHTTPHandler) DoPost(ctx context.Context, path string, body []byte) (*resty.Response, error) {
	// 创建一个新的 resty 客户端
	client := resty.New()

	// 创建一个模拟的响应
	resp := &resty.Response{
		Request: client.R(),
	}

	// 设置响应内容
	resp.SetBody(h.Response)
	return resp, nil
}

func TestParseResponse(t *testing.T) {
	// 创建 Provider 实例
	p := &Provider{
		Provider: *base.NewProvider(
			"deepseek",
			"test-key",
			"https://test.endpoint",
			"deepseek-chat",
			base.RequestConfig{},
		),
	}

	tests := []struct {
		name     string
		response string
		want     *llm.CompletionResponse
		wantErr  bool
	}{
		{
			name: "普通响应",
			response: `{
				"id": "31d0f92e-5cac-43eb-b98a-982a6dbb1dd1",
				"object": "chat.completion",
				"created": 1745843257,
				"model": "deepseek-chat",
				"choices": [
					{
						"index": 0,
						"message": {
							"role": "assistant",
							"content": "测试响应"
						},
						"logprobs": null,
						"finish_reason": "stop"
					}
				],
				"usage": {
					"prompt_tokens": 1621,
					"completion_tokens": 56,
					"total_tokens": 1677,
					"prompt_tokens_details": {
						"cached_tokens": 1600
					},
					"prompt_cache_hit_tokens": 1600,
					"prompt_cache_miss_tokens": 21
				},
				"system_fingerprint": "fp_8802369eaa_prod0425fp8"
			}`,
			want: &llm.CompletionResponse{
				Content:      "测试响应",
				FinishReason: "stop",
				Usage: llm.Usage{
					PromptTokens:     1621,
					CompletionTokens: 56,
					TotalTokens:      1677,
				},
			},
			wantErr: false,
		},
		{
			name:     "无效的 JSON",
			response: "invalid json",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "空的 choices",
			response: `{"choices": []}`,
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.ParseResponse([]byte(tt.response))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Content, got.Content)
			assert.Equal(t, tt.want.FinishReason, got.FinishReason)
			assert.Equal(t, tt.want.Usage, got.Usage)
		})
	}
}
