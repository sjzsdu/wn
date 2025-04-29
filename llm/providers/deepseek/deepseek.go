package deepseek

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/llm/providers/base"
	"github.com/sjzsdu/wn/share"
)

const (
	name               = "deepseek"
	defaultAPIEndpoint = "https://api.deepseek.com/v1/chat/completions"
)

type Provider struct {
	base.Provider
}

func New(options map[string]interface{}) (llm.Provider, error) {
	p := &Provider{
		Provider: base.Provider{
			Models:    []string{"deepseek-chat", "deepseek-coder"},
			Model:     "deepseek-chat",
			Pname:     name,
			MaxTokens: share.MAX_TOKENS,
			HTTPHandler: base.HTTPHandler{
				APIEndpoint: defaultAPIEndpoint,
				Client:      &http.Client{},
			},
		},
	}

	// 配置验证和设置
	apiKey, ok := options["WN_DEEPSEEK_APIKEY"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("deepseek: WN_DEEPSEEK_APIKEY is required")
	}
	p.APIKey = apiKey

	if endpoint, ok := options["WN_DEEPSEEK_ENDPOINT"].(string); ok && endpoint != "" {
		p.APIEndpoint = endpoint
	}
	if models, ok := options["WN_DEEPSEEK_MODELS"].([]string); ok && len(models) > 0 {
		p.Models = models
	}
	if model, ok := options["WN_DEEPSEEK_MODEL"].(string); ok {
		p.Model = model
	}

	return p, nil
}

// Complete 实现完整的请求处理
func (p *Provider) prepareRequest(req llm.CompletionRequest, stream bool) ([]byte, error) {
	reqBody := p.Provider.CommonRequest(req)
	reqBodyStruct := p.HandleRequestBody(req, reqBody).(*DeepseekRequest)
	reqBodyStruct.Stream = stream

	// if share.GetDebug() {
	// 	helper.PrintWithLabel("[DEBUG] Request Body", reqBodyStruct)
	// }
	return json.Marshal(reqBodyStruct)
}

func (p *Provider) Complete(ctx context.Context, req llm.CompletionRequest) (llm.CompletionResponse, error) {
	jsonBody, err := p.prepareRequest(req, false)
	if err != nil {
		return llm.CompletionResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := p.DoRequest(ctx, jsonBody)
	if err != nil {
		return llm.CompletionResponse{}, err
	}
	defer resp.Body.Close()

	return p.ParseResponse(resp.Body)
}

func (p *Provider) CompleteStream(ctx context.Context, req llm.CompletionRequest, handler llm.StreamHandler) error {
	jsonBody, err := p.prepareRequest(req, true)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := p.DoRequest(ctx, jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return p.handleStream(resp.Body, handler)
}

func (p *Provider) HandleRequestBody(req llm.CompletionRequest, reqBody map[string]interface{}) interface{} {
	request, _ := helper.MapToStruct[DeepseekRequest](reqBody)
	request.Tools = p.handleTools(req.Tools)
	request.Messages = p.handleMessages(req.Messages)
	request.ResponseFormat = ResponseFormat{
		Type: req.ResponseFormat,
	}

	// if share.GetDebug() {
	// 	helper.PrintWithLabel("[DEBUG] Resolve Tools:", request.Tools)
	// }

	return request
}

// handleStream 处理流式响应
func (p *Provider) handleStream(body io.Reader, handler llm.StreamHandler) error {
	reader := bufio.NewReader(body)
	var fullContent strings.Builder
	var toolCalls []llm.ToolCall
	var usage llm.Usage

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// 流式响应结束时返回完整数据
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
		var streamResp StreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			return fmt.Errorf("unmarshal stream response: %w", err)
		}

		if len(streamResp.Choices) > 0 {
			choice := streamResp.Choices[0]

			// 处理内容
			if choice.Delta.Content != "" {
				fullContent.WriteString(choice.Delta.Content)
				handler(llm.StreamResponse{
					Content: choice.Delta.Content,
					Done:    false,
				})
			}

			// 处理工具调用
			if len(choice.Delta.ToolCalls) > 0 {
				for _, tc := range choice.Delta.ToolCalls {
					toolCalls = append(toolCalls, llm.ToolCall{
						ID:        tc.ID,
						Type:      tc.Type,
						Function:  tc.Function.Name,
						Arguments: helper.StringToMap(tc.Function.Arguments),
					})
				}
			}

			// 处理结束原因
			if choice.FinishReason != "" {
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

func (p *Provider) handleTools(tools []mcp.Tool) []Tool {
	if len(tools) == 0 {
		return nil
	}

	result := make([]Tool, 0, len(tools))
	for _, t := range tools {
		// 创建一个新的 deepseek Tool
		dsTool := Tool{
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
			dsTool.Function.Parameters["required"] = t.InputSchema.Required
		}

		result = append(result, dsTool)
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

func (p *Provider) ParseResponse(body io.Reader) (llm.CompletionResponse, error) {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return llm.CompletionResponse{}, fmt.Errorf("read response body: %w", err)
	}

	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Raw Response:", string(bodyBytes))
	}

	var deepseekResp DeepseekResponse
	if err := json.Unmarshal(bodyBytes, &deepseekResp); err != nil {
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Raw Response Body:", string(bodyBytes))
			helper.PrintWithLabel("[DEBUG] Unmarshal Error:", err)
		}
		return llm.CompletionResponse{}, fmt.Errorf("decode response: %w, raw response: %s", err, string(bodyBytes))
	}

	if len(deepseekResp.Choices) == 0 {
		return llm.CompletionResponse{}, fmt.Errorf("no choices in response")
	}

	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Deepseek Response:", deepseekResp)
	}

	choice := deepseekResp.Choices[0]
	resp := llm.CompletionResponse{
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
		Usage: llm.Usage{
			PromptTokens:     deepseekResp.Usage.PromptTokens,
			CompletionTokens: deepseekResp.Usage.CompletionTokens,
			TotalTokens:      deepseekResp.Usage.TotalTokens,
		},
	}

	// 如果是工具调用，则处理工具调用的响应
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
		resp.Content = ""
		resp.ToolCalls = toolCalls
	}

	return resp, nil
}

func (p *Provider) ParseStreamResponse(data string) (content string, finishReason string, err error) {
	var streamResp StreamResponse

	if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
		return "", "", fmt.Errorf("unmarshal response: %w", err)
	}

	// if share.GetDebug() {
	// 	helper.PrintWithLabel("[DEBUG] Stream Response:", streamResp)
	// }

	if len(streamResp.Choices) == 0 {
		return "", "", fmt.Errorf("empty choices in response")
	}

	// 新增对工具调用的处理
	if streamResp.Choices[0].FinishReason == "tool_calls" {
		// if share.GetDebug() {
		// 	helper.PrintWithLabel("[DEBUG] Tool Calls:", streamResp.Choices[0].Delta.ToolCalls)
		// }
		return "", streamResp.Choices[0].FinishReason, nil
	}

	return streamResp.Choices[0].Delta.Content, streamResp.Choices[0].FinishReason, nil
}

func init() {
	llm.Register(name, New)
}
