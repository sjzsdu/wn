package base

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

// CompleteStream 实现流式输出
// StreamResponseParser 定义流式响应解析接口
type StreamResponseParser interface {
	ParseStreamResponse(data string) (content string, finishReason string, err error)
}

// CompleteStream 实现通用的流式输出处理
func (p *Provider) CompleteStream(ctx context.Context, req llm.CompletionRequest, handler llm.StreamHandler) error {
	if req.Model == "" {
		req.Model = p.Model
	}

	// 构建请求体
	reqBody := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"max_tokens":  req.MaxTokens,
		"stream":      true,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.APIEndpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
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
		
		// 使用具体实现的解析器解析响应
		if streamParser, ok := p.parser.(StreamResponseParser); ok {
			content, finishReason, err := streamParser.ParseStreamResponse(data)
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
		} else {
			// 如果没有实现流式解析接口，则回退到非流式模式
			resp, err := p.Complete(ctx, req)
			if err != nil {
				return err
			}
			handler(llm.StreamResponse{
				Content: resp.Content,
				Done:    true,
			})
			return nil
		}
	}

	return nil
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
