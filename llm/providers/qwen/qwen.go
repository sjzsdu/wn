package qwen

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
	name               = "qwen"
	defaultAPIEndpoint = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
)

type Provider struct {
	base.Provider
}

func New(options map[string]interface{}) (llm.Provider, error) {
	p := &Provider{
		Provider: base.Provider{
			Model:     "qwen-turbo",
			Pname:     name,
			MaxTokens: share.MAX_TOKENS,
			HTTPHandler: base.HTTPHandler{
				APIEndpoint: defaultAPIEndpoint,
				Client:      &http.Client{},
			},
		},
	}

	apiKey, ok := options["WN_QWEN_APIKEY"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("qwen: WN_QWEN_APIKEY is required")
	}
	p.APIKey = apiKey

	if endpoint, ok := options["WN_QWEN_ENDPOINT"].(string); ok && endpoint != "" {
		p.APIEndpoint = endpoint
	}
	if model, ok := options["WN_QWEN_MODEL"].(string); ok {
		p.Model = model
	}

	return p, nil
}

func (p *Provider) Complete(ctx context.Context, req llm.CompletionRequest) (llm.CompletionResponse, error) {
	reqBody := p.Provider.CommonRequest(req)
	reqBodyStruct := p.HandleRequestBody(req, reqBody).(*CompletionRequestBody)
	reqBodyStruct.Stream = false

	jsonBody, err := json.Marshal(reqBodyStruct)
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
	reqBody := p.Provider.CommonRequest(req)
	reqBodyStruct := p.HandleRequestBody(req, reqBody).(*CompletionRequestBody)
	reqBodyStruct.Stream = true

	jsonBody, err := json.Marshal(reqBodyStruct)
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

func (p *Provider) handleStream(body io.Reader, handler llm.StreamHandler) error {
	reader := bufio.NewReader(body)
	var fullContent strings.Builder
	var usage llm.Usage

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
		content, finishReason, err := p.ParseStreamResponse(data)
		if err != nil {
			return fmt.Errorf("parse stream response: %w", err)
		}

		if content != "" {
			fullContent.WriteString(content)
			handler(llm.StreamResponse{
				Content: content,
				Done:    false,
			})
		}

		if finishReason != "" {
			handler(llm.StreamResponse{
				Content:      fullContent.String(),
				FinishReason: finishReason,
				Done:         true,
				Response: &llm.CompletionResponse{
					Content:      fullContent.String(),
					FinishReason: finishReason,
					Usage:        usage,
				},
			})
			break
		}
	}

	return nil
}

func (p *Provider) ParseResponse(body io.Reader) (llm.CompletionResponse, error) {
	var qwenResp struct {
		Output struct {
			Text         string     `json:"text"`
			ToolCalls    []ToolCall `json:"tool_calls"`
			FinishReason string     `json:"finish_reason"`
		} `json:"output"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(body).Decode(&qwenResp); err != nil {
		return llm.CompletionResponse{}, fmt.Errorf("decode response: %w", err)
	}

	resp := llm.CompletionResponse{
		Content:      qwenResp.Output.Text,
		FinishReason: qwenResp.Output.FinishReason,
		Usage: llm.Usage{
			PromptTokens:     qwenResp.Usage.InputTokens,
			CompletionTokens: qwenResp.Usage.OutputTokens,
			TotalTokens:      qwenResp.Usage.InputTokens + qwenResp.Usage.OutputTokens,
		},
	}

	if len(qwenResp.Output.ToolCalls) > 0 {
		toolCalls := make([]llm.ToolCall, len(qwenResp.Output.ToolCalls))
		for i, tc := range qwenResp.Output.ToolCalls {
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
	var streamResp struct {
		Output struct {
			Text         string `json:"text"`
			FinishReason string `json:"finish_reason"`
		} `json:"output"`
	}

	if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
		return "", "", fmt.Errorf("unmarshal response: %w", err)
	}

	return streamResp.Output.Text, streamResp.Output.FinishReason, nil
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

// AvailableModels 通过API获取支持的模型列表
func (p *Provider) AvailableModels() []string {
	endpoint := "https://dashscope.aliyuncs.com/api/v1/models"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return []string{"qwen-turbo", "qwen-plus", "qwen-max"}
	}
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := p.Client.Do(req)
	if err != nil {
		return []string{"qwen-turbo", "qwen-plus", "qwen-max"}
	}
	defer resp.Body.Close()

	var response struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return []string{"qwen-turbo", "qwen-plus", "qwen-max"}
	}

	models := make([]string, 0)
	for _, model := range response.Models {
		if strings.HasPrefix(model.Name, "qwen") {
			models = append(models, model.Name)
		}
	}
	return models
}
