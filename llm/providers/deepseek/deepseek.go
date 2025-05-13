package deepseek

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/llm/providers/base"
	"github.com/sjzsdu/wn/share"
)

const (
	name            = "deepseek"
	baseAPIEndpoint = "https://api.deepseek.com/v1"
	CompletionPath  = "/chat/completions"
	modelsPath      = "/models"
)

type Provider struct {
	base.Provider
	StreamHandler StreamHandler
}

func New(options map[string]interface{}) (llm.Provider, error) {
	// 配置验证和设置
	apiKey, ok := options["WN_DEEPSEEK_APIKEY"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("deepseek: WN_DEEPSEEK_APIKEY is required")
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
			"deepseek-chat",
			config,
		),
	}

	if endpoint, ok := options["WN_DEEPSEEK_ENDPOINT"].(string); ok && endpoint != "" {
		p.APIEndpoint = endpoint
	}
	if model, ok := options["WN_DEEPSEEK_MODEL"].(string); ok {
		p.Model = model
	}

	return p, nil
}

// PrepareRequest 将通用的 CompletionRequest 转换为 DeepseekRequest
func (p *Provider) PrepareRequest(req llm.CompletionRequest, stream bool) ([]byte, error) {
	// 创建 DeepseekRequest
	request := &DeepseekRequest{
		Messages: make([]Message, len(req.Messages)),
		Model:    p.GetModel(),
		Stream:   stream,
	}

	// 复制基本字段
	request.MaxTokens = req.MaxTokens

	// 转换消息
	for i, msg := range req.Messages {
		request.Messages[i] = Message{
			Role:       msg.Role,
			Content:    msg.Content,
			Name:       msg.Name,
			ToolCallId: msg.ToolCallId,
		}

		// 如果有工具调用，也需要转换
		if msg.ToolCalls != nil {
			toolCalls := make([]ToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				toolCalls[j] = ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: CallFunction{
						Name:      tc.Function,
						Arguments: helper.ToJSONString(tc.Arguments),
					},
				}
			}
			request.Messages[i].ToolCalls = toolCalls
		}
	}

	// 设置响应格式
	if req.ResponseFormat != "" {
		request.ResponseFormat = ResponseFormat{
			Type: req.ResponseFormat,
		}
	}

	// 处理工具
	if req.Tools != nil {
		tools := make([]Tool, len(req.Tools))
		for i, t := range req.Tools {
			tools[i] = Tool{
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
		}
		request.Tools = tools
	}

	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Request Body", request)
	}

	return json.Marshal(request)
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

	var deepseekResp DeepseekResponse
	if err := json.Unmarshal(bodyBytes, &deepseekResp); err != nil {
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Unmarshal Error:", err)
		}
		return nil, fmt.Errorf("解析响应失败: %w, 原始响应: %s", err, string(bodyBytes))
	}

	if len(deepseekResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Deepseek Response:", deepseekResp)
	}

	choice := deepseekResp.Choices[0]
	response := llm.CompletionResponse{
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
		Usage: llm.Usage{
			PromptTokens:     deepseekResp.Usage.PromptTokens,
			CompletionTokens: deepseekResp.Usage.CompletionTokens,
			TotalTokens:      deepseekResp.Usage.TotalTokens,
		},
	}

	if choice.Message.ToolCalls != nil && len(choice.Message.ToolCalls) > 0 {
		toolCalls := make([]llm.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
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
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := p.DoStream(ctx, CompletionPath, jsonBody)
	if err != nil {
		return err
	}

	return p.HandleStreamResponse(resp, p)
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

func (p *Provider) ParseStreamResponse(data string) (content string, finishReason string, err error) {
	var streamResp StreamResponse

	if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
		return "", "", fmt.Errorf("unmarshal response: %w", err)
	}

	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Stream Response", streamResp)
	}

	if len(streamResp.Choices) == 0 {
		return "", "", fmt.Errorf("empty choices in response")
	}

	choice := streamResp.Choices[0]

	// 处理工具调用
	if choice.FinishReason == "tool_calls" && len(choice.Delta.ToolCalls) > 0 {
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Tool Calls", choice.Delta.ToolCalls)
		}
		return "", choice.FinishReason, nil
	}

	return choice.Delta.Content, choice.FinishReason, nil
}

// AvailableModels 通过API获取支持的模型列表
func (p *Provider) AvailableModels() []string {
	resp, err := p.DoGet(context.Background(), modelsPath, nil)
	if err != nil {
		return []string{}
	}

	var response struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	// 使用 Body() 而不是 RawBody()
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Models Response Error", err)
			helper.PrintWithLabel("[DEBUG] Raw Response", string(resp.Body()))
		}
		return []string{}
	}

	models := make([]string, 0, len(response.Data))
	for _, model := range response.Data {
		models = append(models, model.ID)
	}
	return models
}

func init() {
	llm.Register(name, New)
}
