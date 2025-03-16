package claude

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/llm/providers/base"
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
			APIEndpoint: defaultAPIEndpoint,
			Client:      &http.Client{},
			Models:      []string{"claude-2", "claude-instant-1"},
			Model:       "claude-2",
			Pname:       name,
		},
	}

	p.Provider.SetParser(p)

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