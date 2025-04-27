package deepseek

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sjzsdu/wn/llm"
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

func TestHandleRequestBody(t *testing.T) {
	p := &Provider{}
	tests := []struct {
		name     string
		req      llm.CompletionRequest
		reqBody  map[string]interface{}
		wantType string
	}{
		{
			name: "基本请求体转换",
			req: llm.CompletionRequest{
				Messages: []llm.Message{
					{Role: "user", Content: "hello"},
				},
				Model: "deepseek-chat",
			},
			reqBody: map[string]interface{}{
				"messages": []llm.Message{
					{Role: "user", Content: "hello"},
				},
				"model": "deepseek-chat",
			},
			wantType: "*deepseek.CompletionRequestBody",
		},
		{
			name: "带工具的请求体转换",
			req: llm.CompletionRequest{
				Messages: []llm.Message{
					{Role: "user", Content: "hello"},
				},
				Tools: []mcp.Tool{
					{
						Name:        "test_tool",
						Description: "test description",
						InputSchema: mcp.ToolInputSchema{
							Type: "object",
							Properties: map[string]interface{}{
								"test": map[string]interface{}{
									"type": "string",
								},
							},
							Required: []string{"test"},
						},
					},
				},
			},
			reqBody: map[string]interface{}{
				"messages": []llm.Message{
					{Role: "user", Content: "hello"},
				},
			},
			wantType: "*deepseek.CompletionRequestBody",
		},
		{
			name: "带单路径工具的请求体转换",
			req: llm.CompletionRequest{
				Messages: []llm.Message{
					{Role: "user", Content: "hello"},
				},
				Tools: []mcp.Tool{
					{
						Name:        "path_tool",
						Description: "test path tool",
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
			},
			reqBody: map[string]interface{}{
				"messages": []llm.Message{
					{Role: "user", Content: "hello"},
				},
			},
			wantType: "*deepseek.CompletionRequestBody",
		},
		{
			name: "带多路径工具的请求体转换",
			req: llm.CompletionRequest{
				Messages: []llm.Message{
					{Role: "user", Content: "hello"},
				},
				Tools: []mcp.Tool{
					{
						Name:        "paths_tool",
						Description: "test paths tool",
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
			},
			reqBody: map[string]interface{}{
				"messages": []llm.Message{
					{Role: "user", Content: "hello"},
				},
			},
			wantType: "*deepseek.CompletionRequestBody",
		},
		{
			name: "带编辑工具的请求体转换",
			req: llm.CompletionRequest{
				Messages: []llm.Message{
					{Role: "user", Content: "hello"},
				},
				Tools: []mcp.Tool{
					{
						Name:        "edit_tool",
						Description: "test edit tool",
						InputSchema: mcp.ToolInputSchema{
							Type: "object",
							Properties: map[string]interface{}{
								"path": map[string]interface{}{
									"type": "string",
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
										"required": []string{"oldText", "newText"},
									},
								},
								"dryRun": map[string]interface{}{
									"type":        "boolean",
									"default":     false,
									"description": "Preview changes using git-style diff format",
								},
							},
							Required: []string{"path", "edits"},
						},
					},
				},
			},
			reqBody: map[string]interface{}{
				"messages": []llm.Message{
					{Role: "user", Content: "hello"},
				},
			},
			wantType: "*deepseek.CompletionRequestBody",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.HandleRequestBody(tt.req, tt.reqBody)
			assert.NotNil(t, result)
			assert.IsType(t, &CompletionRequestBody{}, result)

			// 检查转换后的结构体
			reqBody := result.(*CompletionRequestBody)

			// 验证消息是否正确转换
			if len(tt.req.Messages) > 0 {
				assert.Equal(t, tt.req.Messages[0].Content, reqBody.Messages[0].Content)
				assert.Equal(t, tt.req.Messages[0].Role, reqBody.Messages[0].Role)
			}

			// 验证工具是否正确转换
			if len(tt.req.Tools) > 0 {
				assert.NotEmpty(t, reqBody.Tools)
				assert.Equal(t, "function", reqBody.Tools[0].Type)
				assert.Equal(t, tt.req.Tools[0].Name, reqBody.Tools[0].Function.Name)
				assert.Equal(t, tt.req.Tools[0].Description, reqBody.Tools[0].Function.Description)
			}
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
