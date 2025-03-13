package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sjzsdu/wn/llm"
)

const (
	defaultAPIEndpoint = "https://api.deepseek.com/v1/chat/completions"
)

// Provider 实现DeepSeek的大模型提供商
type Provider struct {
	apiKey      string
	apiEndpoint string
	client      *http.Client
	models      []string
}

// 确保Provider实现了llm.Provider接口
var _ llm.Provider = (*Provider)(nil)

// New 创建一个新的DeepSeek提供商
func New(options map[string]interface{}) (llm.Provider, error) {
	p := &Provider{
		apiEndpoint: defaultAPIEndpoint,
		client:      &http.Client{},
		models:      []string{"deepseek-chat", "deepseek-coder"},
	}

	// 从 options 中获取配置
	apiKey, ok := options["WN_DEEPSEEK_APIKEY"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("deepseek: WN_DEEPSEEK_APIKEY is required")
	}
	p.apiKey = apiKey

	if endpoint, ok := options["WN_DEEPSEEK_ENDPOINT"].(string); ok && endpoint != "" {
		p.apiEndpoint = endpoint
	}
	if models, ok := options["WN_DEEPSEEK_MODELS"].([]string); ok && len(models) > 0 {
		p.models = models
	}

	return p, nil
}

// Name 返回提供商名称
func (p *Provider) Name() string {
	return "deepseek"
}

// AvailableModels 返回支持的模型列表
func (p *Provider) AvailableModels() []string {
	return p.models
}

// Complete 发送请求到DeepSeek并获取回复
func (p *Provider) Complete(ctx context.Context, req llm.CompletionRequest) (llm.CompletionResponse, error) {
	deepseekReq := struct {
		Model     string        `json:"model"`
		Messages  []llm.Message `json:"messages"`
		MaxTokens int           `json:"max_tokens,omitempty"`
	}{
		Model:     req.Model,
		Messages:  req.Messages,
		MaxTokens: req.MaxTokens,
	}

	if deepseekReq.Model == "" {
		deepseekReq.Model = "deepseek-chat" // 默认模型
	}

	reqBody, err := json.Marshal(deepseekReq)
	if err != nil {
		return llm.CompletionResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.apiEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		return llm.CompletionResponse{}, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return llm.CompletionResponse{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return llm.CompletionResponse{}, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

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

	if err := json.NewDecoder(resp.Body).Decode(&deepseekResp); err != nil {
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
	llm.Register("deepseek", New)
}
