package openai

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
	name            = "openai"
	baseAPIEndpoint = "https://api.openai.com/v1"
	CompletionPath  = "/chat/completions"
	modelsPath      = "/models"
)

type Provider struct {
	base.Provider
	StreamHandler StreamHandler
}

func New(options map[string]interface{}) (llm.Provider, error) {
	apiKey, ok := options["WN_OPENAI_APIKEY"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("openai: WN_OPENAI_APIKEY is required")
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
			"gpt-3.5-turbo",
			config,
		),
	}

	if endpoint, ok := options["WN_OPENAI_ENDPOINT"].(string); ok && endpoint != "" {
		p.APIEndpoint = endpoint
	}
	if model, ok := options["WN_OPENAI_MODEL"].(string); ok {
		p.Model = model
	}

	return p, nil
}

// Complete 实现完整的请求处理
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
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Raw Response", resp.String())
		}
		return nil, fmt.Errorf("请求失败，状态码：%d，响应：%s", resp.StatusCode(), resp.String())
	}
	bodyBytes := resp.Body()
	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Raw Response", string(bodyBytes))
	}

	return p.ParseResponse(bodyBytes)
}

// CompleteStream 实现流式请求处理
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

// HandleStream 处理流式响应
func (p *Provider) HandleStream(bytes []byte) error {
	line := strings.TrimSpace(string(bytes))
	helper.PrintWithLabel("[DEBUG] Stream Response", line)
	if line == "" || line == "data: [DONE]" || !strings.HasPrefix(line, "data: ") {
		return nil
	}
	data := strings.TrimPrefix(line, "data: ")

	return p.StreamHandler.AddContent([]byte(data))
}

// PrepareRequest 准备请求
func (p *Provider) PrepareRequest(req llm.CompletionRequest, stream bool) ([]byte, error) {
	// 创建请求体结构
	request := &CompletionRequestBody{
		Model:     p.Model,
		MaxTokens: 4096,
		Stream:    stream,
		Messages:  make([]Message, len(req.Messages)),
	}

	if req.MaxTokens > 0 {
		request.MaxTokens = req.MaxTokens
	}

	// 处理消息
	for i, msg := range req.Messages {
		message := Message{
			Role:       msg.Role,
			Content:    msg.Content,
			Name:       msg.Name,
			ToolCallID: msg.ToolCallId,
		}

		// 处理工具调用
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
			message.ToolCalls = toolCalls
		}

		request.Messages[i] = message
	}

	// 处理工具
	if req.Tools != nil {
		tools := make([]Tool, 0, len(req.Tools))
		for _, t := range req.Tools {
			// 确保参数符合 OpenAI 的要求
			parameters := map[string]interface{}{
				"type":       "object",
				"properties": t.InputSchema.Properties,
			}

			// 如果有必需参数，添加到 schema 中
			if len(t.InputSchema.Required) > 0 {
				parameters["required"] = t.InputSchema.Required
			}

			tool := Tool{
				Type: "function",
				Function: Function{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  parameters,
				},
			}
			tools = append(tools, tool)
		}
		request.Tools = tools
	}

	// 处理响应格式
	if req.ResponseFormat != "" {
		request.ResponseFormat = ResponseFormat{
			Type: req.ResponseFormat,
		}
	}

	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Request Body", request)
	}

	return json.Marshal(request)
}

func (p *Provider) ParseResponse(bodyBytes []byte) (*llm.CompletionResponse, error) {
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(bodyBytes, &openAIResp); err != nil {
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Unmarshal Error:", err)
			helper.PrintWithLabel("[DEBUG] Raw Response", string(bodyBytes))
		}
		return nil, fmt.Errorf("解析响应失败: %w, 原始响应: %s", err, string(bodyBytes))
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := openAIResp.Choices[0]
	resp := llm.CompletionResponse{
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
		Usage: llm.Usage{
			PromptTokens:     openAIResp.Usage.PromptTokens,
			CompletionTokens: openAIResp.Usage.CompletionTokens,
			TotalTokens:      openAIResp.Usage.TotalTokens,
		},
	}

	// 处理工具调用
	if len(choice.Message.ToolCalls) > 0 {
		toolCalls := make([]llm.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			toolCalls[i] = llm.ToolCall{
				ID:        tc.ID,
				Type:      tc.Type,
				Function:  tc.Function.Name,
				Arguments: helper.StringToMap(tc.Function.Arguments),
			}
		}
		resp.Content = ""
		resp.ToolCalls = toolCalls
	}

	return &resp, nil
}

func (p *Provider) ParseStreamResponse(data string) (content string, finishReason string, err error) {
	var streamResp struct {
		Choices []struct {
			Delta struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"delta"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}

	if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
		return "", "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(streamResp.Choices) == 0 {
		return "", "", nil
	}

	choice := streamResp.Choices[0]

	// 如果是工具调用，返回空内容和工具调用的完成原因
	if choice.FinishReason == "tool_calls" {
		return "", choice.FinishReason, nil
	}

	return choice.Delta.Content, choice.FinishReason, nil
}

func (p *Provider) AvailableModels() []string {
	resp, err := p.DoGet(context.Background(), "/models", nil)
	if err != nil {
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Get Models Error:", err)
		}
		return []string{}
	}

	var response struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	bodyBytes := resp.Body()
	helper.PrintWithLabel("[DEBUG] Get Models Response", string(bodyBytes))
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Models Response Error", err)
			helper.PrintWithLabel("[DEBUG] Raw Response", string(bodyBytes))
		}
		return []string{}
	}

	models := make([]string, 0)
	for _, model := range response.Data {
		if strings.HasPrefix(model.ID, "gpt") {
			models = append(models, model.ID)
		}
	}
	return models
}

func init() {
	llm.Register(name, New)
}
