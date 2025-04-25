package deepseek

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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
		"tools":      req.Tools,
	}

	if reqBody["max_tokens"] == 0 {
		reqBody["max_tokens"] = p.MaxTokens
	}

	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Request Body:", reqBody)
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
				Done:         true,
			})
			break
		}
	}

	return nil
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

// ParseStreamResponse 实现流式响应解析
func (p *Provider) ParseStreamResponse(data string) (content string, finishReason string, err error) {
	var streamResp struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}

	if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
		return "", "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(streamResp.Choices) == 0 {
		return "", "", nil
	}

	return streamResp.Choices[0].Delta.Content, streamResp.Choices[0].FinishReason, nil
}

func init() {
	llm.Register(name, New)
}
