package claude

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
	name               = "claude"
	baseAPIEndpoint    = "https://api.anthropic.com/v1"
	defaultAPIEndpoint = baseAPIEndpoint + "/messages"
	modelsAPIEndpoint  = baseAPIEndpoint + "/models"
)

type Provider struct {
	base.Provider
}

func New(options map[string]interface{}) (llm.Provider, error) {
	p := &Provider{
		Provider: base.Provider{
			Model:     "claude-2",
			Pname:     name,
			MaxTokens: share.MAX_TOKENS,
			HTTPHandler: base.HTTPHandler{
				APIEndpoint: defaultAPIEndpoint,
				Client:      &http.Client{},
			},
		},
	}

	apiKey, ok := options["WN_CLAUDE_APIKEY"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("claude: WN_CLAUDE_APIKEY is required")
	}
	p.APIKey = apiKey

	if endpoint, ok := options["WN_CLAUDE_ENDPOINT"].(string); ok && endpoint != "" {
		p.APIEndpoint = endpoint
	}
	if model, ok := options["WN_CLAUDE_MODEL"].(string); ok {
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
	var claudeResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(body).Decode(&claudeResp); err != nil {
		return llm.CompletionResponse{}, fmt.Errorf("decode response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return llm.CompletionResponse{}, fmt.Errorf("no content in response")
	}

	return llm.CompletionResponse{
		Content:      claudeResp.Content[0].Text,
		FinishReason: claudeResp.StopReason,
		Usage: llm.Usage{
			PromptTokens:     claudeResp.Usage.InputTokens,
			CompletionTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:      claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
	}, nil
}

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
    req, err := http.NewRequestWithContext(context.Background(), "GET", modelsAPIEndpoint, nil)
    if err != nil {
        helper.PrintWithLabel("[ERROR] Failed to create models request", err)
        return []string{}
    }
    
    // 添加必要的请求头
    req.Header.Set("anthropic-version", "2023-06-01")
    req.Header.Set("Authorization", "Bearer "+p.APIKey)
    
    // 使用 base.HTTPHandler 的 DoRequest 方法
    resp, err := p.DoRequest(context.Background(), nil)
    if err != nil {
        helper.PrintWithLabel("[ERROR] Failed to fetch models", err)
        return []string{}
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        helper.PrintWithLabel("[ERROR] API returned non-200 status code", fmt.Sprintf("status: %d, body: %s", resp.StatusCode, string(body)))
        return []string{}
    }
    
    var response struct {
        Models []struct {
            Name string `json:"name"`
        } `json:"models"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
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
