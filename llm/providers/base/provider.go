package base

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sjzsdu/wn/llm"
)

// Provider 提供基础的 LLM Provider 实现
type Provider struct {
	APIKey      string
	APIEndpoint string
	Client      *http.Client
	Models      []string
	Model       string
	Pname       string
	parser      ResponseParser
}

// ResponseParser 定义响应解析接口
type ResponseParser interface {
	ParseResponse(body io.Reader) (llm.CompletionResponse, error)
}

func (p *Provider) SetParser(parser ResponseParser) {
	p.parser = parser
}

// Name 返回提供商名称
func (p *Provider) Name() string {
	return p.Pname
}

// Complete 发送请求到 LLM 并获取回复
func (p *Provider) Complete(ctx context.Context, req llm.CompletionRequest) (llm.CompletionResponse, error) {
	if req.Model == "" {
		req.Model = p.Model
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return llm.CompletionResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.APIEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		return llm.CompletionResponse{}, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return llm.CompletionResponse{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return llm.CompletionResponse{}, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	if p.parser == nil {
		return llm.CompletionResponse{}, fmt.Errorf("response parser not set")
	}
	return p.parser.ParseResponse(resp.Body)
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
