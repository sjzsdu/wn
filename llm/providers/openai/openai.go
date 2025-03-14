package openai

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
	defaultAPIEndpoint = "https://api.openai.com/v1/chat/completions"
)

// Provider 实现OpenAI的大模型提供商
type Provider struct {
	apiKey      string
	apiEndpoint string
	client      *http.Client
	models      []string
	model       string
}

// 确保Provider实现了llm.Provider接口
var _ llm.Provider = (*Provider)(nil)

// New 创建一个新的OpenAI提供商
func New(options map[string]interface{}) (llm.Provider, error) {
	p := &Provider{
		apiEndpoint: defaultAPIEndpoint,
		client:      &http.Client{},
		models:      []string{"gpt-3", "gpt-3.5-turbo", "gpt-4"},
		model:       "gpt-3",
	}

	// 从 options 中获取配置
	apiKey, ok := options["WN_OPENAI_APIKEY"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("openai: WN_OPENAI_APIKEY is required")
	}
	p.apiKey = apiKey

	if endpoint, ok := options["WN_OPENAI_ENDPOINT"].(string); ok && endpoint != "" {
		p.apiEndpoint = endpoint
	}
	if models, ok := options["WN_OPENAI_MODELS"].([]string); ok && len(models) > 0 {
		p.models = models
	}

	if model, ok := options["WN_OPENAI_MODEL"].(string); ok {
		p.model = model
	}

	return p, nil
}

// Name 返回提供商名称
func (p *Provider) Name() string {
	return "openai"
}

// AvailableModels 返回支持的模型列表
func (p *Provider) AvailableModels() []string {
	return p.models
}

func (p *Provider) SetModel(model string) string {
	if model == "" {
		return p.model
	}
	p.model = model
	return p.model
}

// Complete 发送请求到OpenAI并获取回复
func (p *Provider) Complete(ctx context.Context, req llm.CompletionRequest) (llm.CompletionResponse, error) {
	// 转换为OpenAI特定的请求格式
	openAIReq := struct {
		Model     string        `json:"model"`
		Messages  []llm.Message `json:"messages"`
		MaxTokens int           `json:"max_tokens,omitempty"`
	}{
		Model:     req.Model,
		Messages:  req.Messages,
		MaxTokens: req.MaxTokens,
	}

	if openAIReq.Model == "" {
		openAIReq.Model = "gpt-3.5-turbo" // 默认模型
	}

	reqBody, err := json.Marshal(openAIReq)
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

	var openAIResp struct {
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

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return llm.CompletionResponse{}, fmt.Errorf("decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return llm.CompletionResponse{}, fmt.Errorf("no choices in response")
	}

	return llm.CompletionResponse{
		Content:      openAIResp.Choices[0].Message.Content,
		FinishReason: openAIResp.Choices[0].FinishReason,
		Usage: llm.Usage{
			PromptTokens:     openAIResp.Usage.PromptTokens,
			CompletionTokens: openAIResp.Usage.CompletionTokens,
			TotalTokens:      openAIResp.Usage.TotalTokens,
		},
	}, nil
}

func init() {
	// 注册OpenAI提供商
	llm.Register("openai", New)
}
