package openai

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
	name               = "openai"
	defaultAPIEndpoint = "https://api.openai.com/v1/chat/completions"
)

type Provider struct {
	base.Provider
}

func New(options map[string]interface{}) (llm.Provider, error) {
	p := &Provider{
		Provider: base.Provider{
			Models:    []string{"gpt-3", "gpt-3.5-turbo", "gpt-4"},
			Model:     "gpt-3.5-turbo",
			Pname:     name,
			MaxTokens: share.MAX_TOKENS,
			HTTPHandler: base.HTTPHandler{
				APIEndpoint: defaultAPIEndpoint,
				Client:      &http.Client{},
			},
		},
	}

	apiKey, ok := options["WN_OPENAI_APIKEY"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("openai: WN_OPENAI_APIKEY is required")
	}
	p.APIKey = apiKey

	if endpoint, ok := options["WN_OPENAI_ENDPOINT"].(string); ok && endpoint != "" {
		p.APIEndpoint = endpoint
	}
	if models, ok := options["WN_OPENAI_MODELS"].([]string); ok && len(models) > 0 {
		p.Models = models
	}
	if model, ok := options["WN_OPENAI_MODEL"].(string); ok {
		p.Model = model
	}

	return p, nil
}

// Complete 实现完整的请求处理
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

// CompleteStream 实现流式请求处理
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

func (p *Provider) ParseResponse(body io.Reader) (llm.CompletionResponse, error) {
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

	if err := json.NewDecoder(body).Decode(&openAIResp); err != nil {
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

func (p *Provider) HandleRequestBody(req llm.CompletionRequest, reqBody map[string]interface{}) interface{} {
	request, _ := helper.MapToStruct[CompletionRequestBody](reqBody)
	return request
}

type CompletionRequestBody struct {
	Model     string        `json:"model"`
	Messages  []llm.Message `json:"messages"`
	MaxTokens int           `json:"max_tokens"`
	Stream    bool          `json:"stream"`
}

func init() {
	llm.Register(name, New)
}
