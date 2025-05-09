package qwen

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/llm/providers/base"
	"github.com/sjzsdu/wn/share"
)

const (
	name            = "qwen"
	baseAPIEndpoint = "https://dashscope.aliyuncs.com/api/v1"
	CompletionPath  = "/services/aigc/text-generation/generation"
	modelsPath      = "/models"
)

type Provider struct {
	base.Provider
	StreamHandler StreamHandler
}

func New(options map[string]interface{}) (llm.Provider, error) {
	// 配置验证和设置
	apiKey, ok := options["WN_QWEN_APIKEY"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("qwen: WN_QWEN_APIKEY is required")
	}

	config := base.RequestConfig{
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + apiKey,
		},
		Timeout: 30,
		RetryConfig: &base.RetryConfig{
			MaxRetries:  3,
			RetryDelay:  1,
			RetryPolicy: base.RetryPolicyLinear,
		},
	}

	p := &Provider{
		Provider: *base.NewProvider(
			name,
			apiKey,
			baseAPIEndpoint,
			"qwen-turbo",
			config,
		),
	}

	if endpoint, ok := options["WN_QWEN_ENDPOINT"].(string); ok && endpoint != "" {
		p.APIEndpoint = endpoint
	}
	if model, ok := options["WN_QWEN_MODEL"].(string); ok {
		p.Model = model
	}

	return p, nil
}

func (p *Provider) PrepareRequest(req llm.CompletionRequest, stream bool) ([]byte, error) {
	reqBody := p.Provider.PrepareRequest(req)
	reqBodyStruct := p.HandleRequestBody(req, reqBody).(*QwenRequest)
	// 修正：stream 参数应该在 Parameters 结构中设置
	reqBodyStruct.Parameters.Stream = stream
	reqBodyStruct.Model = p.Model // 添加模型信息

	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Request Body", reqBodyStruct)
	}
	return json.Marshal(reqBodyStruct)
}

func (p *Provider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	jsonBody, err := p.PrepareRequest(req, false)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := p.DoPost(ctx, CompletionPath, jsonBody)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("请求失败，状态码：%d，响应：%s", resp.StatusCode(), resp.String())
	}
	bodyBytes := resp.Body()
	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Raw Response", string(bodyBytes))
	}
	return p.ParseResponse(bodyBytes)
}

func (p *Provider) ParseResponse(bodyBytes []byte) (*llm.CompletionResponse, error) {
	var qwenResp QwenResponse
	if err := json.Unmarshal(bodyBytes, &qwenResp); err != nil {
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Unmarshal Error:", err)
		}
		return nil, fmt.Errorf("解析响应失败: %w, 原始响应: %s", err, string(bodyBytes))
	}

	if qwenResp.Output == nil {
		return nil, fmt.Errorf("no output in response")
	}

	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Qwen Response:", qwenResp)
	}

	response := llm.CompletionResponse{
		Content:      qwenResp.Output.Text,
		FinishReason: qwenResp.Output.FinishReason,
		Usage: llm.Usage{
			PromptTokens:     qwenResp.Usage.InputTokens,
			CompletionTokens: qwenResp.Usage.OutputTokens,
			TotalTokens:      qwenResp.Usage.InputTokens + qwenResp.Usage.OutputTokens,
		},
	}

	if qwenResp.Output.ToolCalls != nil && len(qwenResp.Output.ToolCalls) > 0 {
		toolCalls := make([]llm.ToolCall, len(qwenResp.Output.ToolCalls))
		for i, tc := range qwenResp.Output.ToolCalls {
			toolCalls[i] = llm.ToolCall{
				ID:        tc.ID,
				Type:      tc.Type,
				Function:  tc.Function.Name,
				Arguments: helper.StringToMap(tc.Function.Arguments),
			}
		}
		response.Content = ""
		response.ToolCalls = toolCalls
	}

	return &response, nil
}

func (p *Provider) CompleteStream(ctx context.Context, req llm.CompletionRequest, handler llm.StreamHandler) error {
    p.StreamHandler = NewStreamHandler(handler)
    jsonBody, err := p.PrepareRequest(req, true)
    if err != nil {
        return fmt.Errorf("准备请求失败: %w", err)
    }

    resp, err := p.DoPost(ctx, CompletionPath, jsonBody)
    if err != nil {
        return fmt.Errorf("发送请求失败: %w", err)
    }

    if resp.StatusCode() != 200 {
        return fmt.Errorf("请求失败，状态码：%d，响应：%s", resp.StatusCode(), resp.String())
    }

    return p.HandleJSONResponse(resp, p)
}

func (p *Provider) HandleStream(bytes []byte) error {
	line := strings.TrimSpace(string(bytes))
	if line == "" || line == "data: [DONE]" || !strings.HasPrefix(line, "data: ") {
		return nil
	}
	data := strings.TrimPrefix(line, "data: ")

	p.StreamHandler.AddContent([]byte(data))
	return nil
}

func (p *Provider) HandleRequestBody(req llm.CompletionRequest, reqBody map[string]interface{}) interface{} {
    request, _ := helper.MapToStruct[QwenRequest](reqBody)
    
    // 设置基本参数
    request.Model = p.Model
    request.Input.Messages = p.handleMessages(req.Messages)
    
    // 设置默认参数
    if request.Parameters.MaxTokens == 0 {
        request.Parameters.MaxTokens = req.MaxTokens
    }
    
    // 处理工具调用
    if req.Tools != nil {
        request.Input.Tools = p.handleTools(req.Tools)
    }

    return request
}

func (p *Provider) handleTools(tools []mcp.Tool) []Tool {
	if len(tools) == 0 {
		return nil
	}

	result := make([]Tool, 0, len(tools))
	for _, t := range tools {
		qwenTool := Tool{
			Type: "function",
			Function: Function{
				Name:        t.Name,
				Description: t.Description,
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": t.InputSchema.Properties,
				},
			},
		}

		if len(t.InputSchema.Required) > 0 {
			qwenTool.Function.Parameters["required"] = t.InputSchema.Required
		}

		result = append(result, qwenTool)
	}

	return result
}

func (p *Provider) handleMessages(messages []llm.Message) []Message {
	result := make([]Message, 0, len(messages))
	for _, msg := range messages {
		// 转换工具调用
		var toolCalls []ToolCall
		if msg.ToolCalls != nil {
			toolCalls = make([]ToolCall, 0)
			for _, tc := range msg.ToolCalls {
				toolCalls = append(toolCalls, ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: CallFunction{
						Name:      tc.Function,
						Arguments: helper.ToJSONString(tc.Arguments),
					},
				})
			}
		}

		result = append(result, Message{
			Role:       msg.Role,
			Content:    msg.Content,
			Name:       msg.Name,
			ToolCallId: msg.ToolCallId,
			ToolCalls:  toolCalls,
		})
	}
	return result
}

func init() {
	llm.Register(name, New)
}

// AvailableModels 通过API获取支持的模型列表
func (p *Provider) AvailableModels() []string {
    // 修正：使用正确的 API 路径
    resp, err := p.DoGet(context.Background(), "/models", nil)
    if err != nil {
        if share.GetDebug() {
            helper.PrintWithLabel("[DEBUG] Get Models Error:", err)
        }
        return []string{"qwen-turbo", "qwen-plus", "qwen-max", "qwen-max-longcontext"}
    }

	var response struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Models Response Error", err)
			helper.PrintWithLabel("[DEBUG] Raw Response", string(resp.Body()))
		}
		// 解析失败时返回基础模型列表
		return []string{"qwen-turbo", "qwen-plus", "qwen-max", "qwen-max-longcontext"}
	}

	models := make([]string, 0)
	for _, model := range response.Models {
		if strings.HasPrefix(model.Name, "qwen") {
			models = append(models, model.Name)
		}
	}

	// 如果没有获取到任何模型，返回基础模型列表
	if len(models) == 0 {
		return []string{"qwen-turbo", "qwen-plus", "qwen-max", "qwen-max-longcontext"}
	}

	return models
}
