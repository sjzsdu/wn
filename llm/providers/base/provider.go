package base

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// HTTPHandler 处理基础的 HTTP 请求
type HTTPHandler struct {
	APIKey      string
	APIEndpoint string
	Client      *http.Client
}

// DoRequest 执行 HTTP 请求
func (h *HTTPHandler) DoRequest(ctx context.Context, reqBody []byte) (*http.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", h.APIEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+h.APIKey)

	resp, err := h.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// Provider 提供基础的 LLM Provider 实现
type Provider struct {
	HTTPHandler
	Models    []string
	Model     string
	Pname     string
	MaxTokens int
}

// Name 返回提供商名称
func (p *Provider) Name() string {
	return p.Pname
}

// AvailableModels 返回支持的模型列表
func (p *Provider) AvailableModels() []string {
	return p.Models
}

// SetModel 设置当前使用的模型
func (p *Provider) SetModel(model string) string {
	if model == "" {
		return p.Model
	}
	p.Model = model
	return p.Model
}
