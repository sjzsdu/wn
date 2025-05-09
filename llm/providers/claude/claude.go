package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/llm/providers/base"
)

const (
	name               = "claude"
	baseAPIEndpoint    = "https://api.anthropic.com/v1"
	defaultAPIEndpoint = baseAPIEndpoint + "/messages"
	modelsAPIEndpoint  = baseAPIEndpoint + "/models"
)

type Provider struct {
	base.Provider
	StreamHandler StreamHandler
}

func New(options map[string]interface{}) (llm.Provider, error) {
	apiKey, ok := options["WN_CLAUDE_APIKEY"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("claude: WN_CLAUDE_APIKEY is required")
	}

	config := base.RequestConfig{
		Headers: map[string]string{
			"Content-Type":      "application/json",
			"Authorization":     "Bearer " + apiKey,
			"anthropic-version": "2023-06-01", // Claude 特有的请求头
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
			defaultAPIEndpoint,
			"claude-2",
			config,
		),
	}

	if endpoint, ok := options["WN_CLAUDE_ENDPOINT"].(string); ok && endpoint != "" {
		p.APIEndpoint = endpoint
	}
	if model, ok := options["WN_CLAUDE_MODEL"].(string); ok {
		p.Model = model
	}

	return p, nil
}

func (p *Provider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	reqBody := p.PrepareRequest(req)
	reqBodyStruct := p.HandleRequestBody(req, reqBody).(*CompletionRequestBody)
	reqBodyStruct.Stream = false

	jsonBody, err := json.Marshal(reqBodyStruct)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := p.DoPost(ctx, "", jsonBody)
	if err != nil {
		return nil, err
	}

	return p.ParseResponse(resp.RawBody())
}

func (p *Provider) CompleteStream(ctx context.Context, req llm.CompletionRequest, handler llm.StreamHandler) error {
	p.StreamHandler = NewStreamHandler(handler)
	reqBody := p.PrepareRequest(req)
	reqBodyStruct := p.HandleRequestBody(req, reqBody).(*CompletionRequestBody)
	reqBodyStruct.Stream = true

	jsonBody, err := json.Marshal(reqBodyStruct)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := p.DoPost(ctx, "", jsonBody)
	if err != nil {
		return err
	}

	return p.handleStream(resp.RawBody())
}

func (p *Provider) handleStream(body io.Reader) error {
	reader := bufio.NewReader(body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read response: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" || line == "data: [DONE]" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if err := p.StreamHandler.AddContent([]byte(data)); err != nil {
			return fmt.Errorf("handle stream data: %w", err)
		}
	}

	return nil
}

// 将响应结构体提取到类型定义中
func (p *Provider) ParseResponse(body io.Reader) (*llm.CompletionResponse, error) {
	var claudeResp StreamResponse
	if err := json.NewDecoder(body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	response := &llm.CompletionResponse{
		Content:      claudeResp.Content[0].Text,
		FinishReason: claudeResp.StopReason,
		Usage: llm.Usage{
			PromptTokens:     claudeResp.Usage.InputTokens,
			CompletionTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:      claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
	}

	// 处理工具调用
	if len(claudeResp.Content[0].ToolCalls) > 0 {
		toolCalls := make([]llm.ToolCall, len(claudeResp.Content[0].ToolCalls))
		for i, tc := range claudeResp.Content[0].ToolCalls {
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

	return response, nil
}

// 移除不再需要的 ParseStreamResponse 方法，使用 StreamHandler 处理
func (p *Provider) ParseStreamResponse(data string) (content string, finishReason string, err error) {
	var streamResp struct {
		Type    string `json:"type"`
		Content struct {
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
	}

	if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
		return "", "", fmt.Errorf("unmarshal response: %w", err)
	}

	if streamResp.Type == "content_block_delta" {
		return streamResp.Content.Text, "", nil
	} else if streamResp.Type == "message_delta" {
		return "", streamResp.StopReason, nil
	}

	return "", "", nil
}

func (p *Provider) HandleRequestBody(req llm.CompletionRequest, reqBody map[string]interface{}) interface{} {
	request, _ := helper.MapToStruct[CompletionRequestBody](reqBody)
	request.Tools = p.handleTools(req.Tools)
	request.Messages = p.handleMessages(req.Messages)
	return request
}

func (p *Provider) handleTools(tools []mcp.Tool) []Tool {
	if len(tools) == 0 {
		return nil
	}

	result := make([]Tool, 0, len(tools))
	for _, t := range tools {
		tool := Tool{
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
			tool.Function.Parameters["required"] = t.InputSchema.Required
		}

		result = append(result, tool)
	}

	return result
}

func (p *Provider) handleMessages(messages []llm.Message) []Message {
	result := make([]Message, len(messages))
	for i, m := range messages {
		msg := Message{
			Role:    m.Role,
			Content: m.Content,
		}

		if m.ToolCallId != "" {
			msg.ToolCallID = m.ToolCallId
			msg.Name = m.Name
		}

		if len(m.ToolCalls) > 0 {
			toolCalls := make([]ToolCall, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				toolCalls[j] = ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: CallFunction{
						Name:      tc.Function,
						Arguments: helper.ToJSONString(tc.Arguments),
					},
				}
			}
			msg.ToolCalls = toolCalls
		}

		result[i] = msg
	}
	return result
}

func init() {
	llm.Register(name, New)
}

func (p *Provider) AvailableModels() []string {
	resp, err := p.DoGet(context.Background(), "/models", nil)
	if err != nil {
		helper.PrintWithLabel("[ERROR] Failed to fetch models", err)
		return []string{}
	}

	var response struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.RawBody()).Decode(&response); err != nil {
		helper.PrintWithLabel("[ERROR] Failed to decode models response", err)
		return []string{}
	}

	if len(response.Models) == 0 {
		helper.PrintWithLabel("[WARN] No models returned from API", nil)
		return []string{}
	}

	models := make([]string, len(response.Models))
	for i, model := range response.Models {
		models[i] = model.Name
	}
	return models
}
