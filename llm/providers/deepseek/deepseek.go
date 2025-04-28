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
	reqBodyStruct := p.HandleRequestBody(req, reqBody).(*CompletionRequestBody)
	reqBodyStruct.Stream = stream

	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Request Body", reqBodyStruct)
	}

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
	request, _ := helper.MapToStruct[CompletionRequestBody](reqBody)
	request.Tools = p.handleTools(req.Tools)
	request.ResponseFormat = ResponseFormat{
		Type: req.ResponseFormat,
	}
	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Resolve Tools:", request.Tools)
	}

	return request
}

// handleStream 处理流式响应
func (p *Provider) handleStream(body io.Reader, handler llm.StreamHandler) error {
	reader := bufio.NewReader(body)
	var fullContent strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// 流式响应结束时返回完整数据
				handler(llm.StreamResponse{
					Content:      fullContent.String(),
					FinishReason: "done",
					Done:         true,
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
			})
			break
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

func (p *Provider) ParseResponse(body io.Reader) (llm.CompletionResponse, error) {
	var deepseekResp DeepseekResponse

	if err := json.NewDecoder(body).Decode(&deepseekResp); err != nil {
		return llm.CompletionResponse{}, fmt.Errorf("decode response: %w", err)
	}

	if len(deepseekResp.Choices) == 0 {
		return llm.CompletionResponse{}, fmt.Errorf("no choices in response")
	}

	return llm.CompletionResponse{
		Content:      deepseekResp.Choices[0].Message.Content,
		FinishReason: deepseekResp.Choices[0].FinishReason,
		Usage: llm.Usage{
			PromptTokens:     deepseekResp.Usage.PromptTokens,
			CompletionTokens: deepseekResp.Usage.CompletionTokens,
			TotalTokens:      deepseekResp.Usage.TotalTokens,
		},
	}, nil
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
