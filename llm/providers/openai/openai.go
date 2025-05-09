package openai

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
		return nil, fmt.Errorf("请求失败，状态码：%d，响应：%s", resp.StatusCode(), resp.String())
	}

	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Raw Response", resp.String())
	}

	return p.ParseResponse(resp.RawBody())
}

// CompleteStream 实现流式请求处理
func (p *Provider) CompleteStream(ctx context.Context, req llm.CompletionRequest, handler llm.StreamHandler) error {
	p.StreamHandler = NewStreamHandler(handler)
	jsonBody, err := p.PrepareRequest(req, true)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := p.DoPost(ctx, "", jsonBody)
	if err != nil {
		return err
	}

	return p.HandleJSONResponse(resp, p)
}

// HandleStream 处理流式响应
func (p *Provider) HandleStream(bytes []byte) error {
	line := strings.TrimSpace(string(bytes))
	if line == "" || line == "data: [DONE]" || !strings.HasPrefix(line, "data: ") {
		return nil
	}
	data := strings.TrimPrefix(line, "data: ")

	return p.StreamHandler.AddContent([]byte(data))
}

// PrepareRequest 准备请求
func (p *Provider) PrepareRequest(req llm.CompletionRequest, stream bool) ([]byte, error) {
	reqBody := p.Provider.PrepareRequest(req)
	reqBodyStruct := p.HandleRequestBody(req, reqBody).(*CompletionRequestBody)
	reqBodyStruct.Stream = stream

	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Request Body", reqBodyStruct)
	}
	return json.Marshal(reqBodyStruct)
}

// handleStream 处理流式响应
func (p *Provider) handleStream(body io.Reader, handler llm.StreamHandler) error {
	reader := bufio.NewReader(body)
	var fullContent strings.Builder
	var toolCalls []llm.ToolCall
	var usage llm.Usage

	var currentToolCall *llm.ToolCall
	var argumentsBuilder strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if currentToolCall != nil {
					currentToolCall.Arguments = helper.StringToMap(argumentsBuilder.String())
					toolCalls = append(toolCalls, *currentToolCall)
				}
				handler(llm.StreamResponse{
					Content:      fullContent.String(),
					FinishReason: "done",
					Done:         true,
					Response: &llm.CompletionResponse{
						Content:      fullContent.String(),
						FinishReason: "done",
						ToolCalls:    toolCalls,
						Usage:        usage,
					},
				})
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
		var streamResp struct {
			Choices []struct {
				Delta struct {
					Content   string     `json:"content"`
					ToolCalls []ToolCall `json:"tool_calls"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
			Usage struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			return fmt.Errorf("unmarshal stream response: %w", err)
		}

		if len(streamResp.Choices) > 0 {
			choice := streamResp.Choices[0]

			if choice.Delta.Content != "" {
				fullContent.WriteString(choice.Delta.Content)
				handler(llm.StreamResponse{
					Content: choice.Delta.Content,
					Done:    false,
				})
			}

			if len(choice.Delta.ToolCalls) > 0 {
				for _, tc := range choice.Delta.ToolCalls {
					if tc.ID != "" {
						if currentToolCall == nil || currentToolCall.ID != tc.ID {
							if currentToolCall != nil {
								currentToolCall.Arguments = helper.StringToMap(argumentsBuilder.String())
								toolCalls = append(toolCalls, *currentToolCall)
							}
							currentToolCall = &llm.ToolCall{
								ID:       tc.ID,
								Type:     tc.Type,
								Function: tc.Function.Name,
							}
							argumentsBuilder.Reset()
						}
					}
					if tc.Function.Arguments != "" {
						argumentsBuilder.WriteString(tc.Function.Arguments)
					}
				}
			}

			if choice.FinishReason != "" {
				if choice.FinishReason == "tool_calls" && currentToolCall != nil {
					currentToolCall.Arguments = helper.StringToMap(argumentsBuilder.String())
					toolCalls = append(toolCalls, *currentToolCall)
				}

				if streamResp.Usage.TotalTokens > 0 {
					usage = llm.Usage{
						PromptTokens:     streamResp.Usage.PromptTokens,
						CompletionTokens: streamResp.Usage.CompletionTokens,
						TotalTokens:      streamResp.Usage.TotalTokens,
					}
				}

				handler(llm.StreamResponse{
					Content:      fullContent.String(),
					FinishReason: choice.FinishReason,
					Done:         true,
					Response: &llm.CompletionResponse{
						Content:      fullContent.String(),
						FinishReason: choice.FinishReason,
						ToolCalls:    toolCalls,
						Usage:        usage,
					},
				})
				break
			}
		}
	}

	return nil
}

func (p *Provider) ParseResponse(body io.Reader) (*llm.CompletionResponse, error) {
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

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
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

func (p *Provider) HandleRequestBody(req llm.CompletionRequest, reqBody map[string]interface{}) interface{} {
	request, _ := helper.MapToStruct[CompletionRequestBody](reqBody)
	request.Tools = p.handleTools(req.Tools)
	request.Messages = p.handleMessages(req.Messages)
	request.ResponseFormat = ResponseFormat{
		Type: req.ResponseFormat,
	}
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

// AvailableModels 通过API获取支持的模型列表
func (p *Provider) AvailableModels() []string {
	resp, err := p.DoGet(context.Background(), "/models", nil)
	if err != nil {
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Get Models Error:", err)
		}
		return []string{"gpt-3.5-turbo", "gpt-4"}
	}

	var response struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	bodyBytes := resp.Body()
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Models Response Error", err)
			helper.PrintWithLabel("[DEBUG] Raw Response", string(bodyBytes))
		}
		return []string{"gpt-3.5-turbo", "gpt-4"}
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
