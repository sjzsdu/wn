package deepseek

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/llm/providers/base"
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
			APIEndpoint: defaultAPIEndpoint,
			Client:      &http.Client{},
			Models:      []string{"deepseek-chat", "deepseek-coder"},
			Model:       "deepseek-chat",
			Pname:       name,
		},
	}

	// 设置响应解析器
	p.Provider.SetParser(p)

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

func (p *Provider) ParseResponse(body io.Reader) (llm.CompletionResponse, error) {
	var deepseekResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

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

func init() {
	llm.Register(name, New)
}
