package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/llm/providers/base"
	"github.com/sjzsdu/wn/share"
)

const (
	name               = "claude"
	defaultAPIEndpoint = "https://api.anthropic.com/v1/messages"
)

type Provider struct {
	base.Provider
}

func New(options map[string]interface{}) (llm.Provider, error) {
	p := &Provider{
		Provider: base.Provider{
			Models:    []string{"claude-2", "claude-instant-1"},
			Model:     "claude-2",
			Pname:     name,
			MaxTokens: share.MAX_TOKENS,
			HTTPHandler: base.HTTPHandler{
				APIEndpoint: defaultAPIEndpoint,
				Client:     &http.Client{},
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
	if models, ok := options["WN_CLAUDE_MODELS"].([]string); ok && len(models) > 0 {
		p.Models = models
	}
	if model, ok := options["WN_CLAUDE_MODEL"].(string); ok {
		p.Model = model
	}

	return p, nil
}

// Complete 实现完整的请求处理
func (p *Provider) Complete(ctx context.Context, req llm.CompletionRequest) (llm.CompletionResponse, error) {
	if req.Model == "" {
		req.Model = p.Model
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return llm.CompletionResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := p.DoRequest(ctx, reqBody)
	if err != nil {
		return llm.CompletionResponse{}, err
	}
	defer resp.Body.Close()

	return p.ParseResponse(resp.Body)
}

// CompleteStream 实现流式请求处理
func (p *Provider) CompleteStream(ctx context.Context, req llm.CompletionRequest, handler llm.StreamHandler) error {
	if req.Model == "" {
		req.Model = p.Model
	}

	reqBody := map[string]interface{}{
		"model":      req.Model,
		"messages":   req.Messages,
		"max_tokens": req.MaxTokens,
		"stream":     true,
	}

	if reqBody["max_tokens"] == 0 {
		reqBody["max_tokens"] = p.MaxTokens
	}

	jsonBody, err := json.Marshal(reqBody)
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

// handleStream 处理流式响应
func (p *Provider) handleStream(body io.Reader, handler llm.StreamHandler) error {
	reader := bufio.NewReader(body)
	var fullContent strings.Builder

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
				Done:        true,
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

func init() {
	llm.Register(name, New)
}
