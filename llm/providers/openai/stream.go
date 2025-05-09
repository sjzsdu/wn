package openai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/llm"
)

type StreamHandler struct {
	handler      llm.StreamHandler
	fullContent  strings.Builder
	toolCalls    []llm.ToolCall
	usage        llm.Usage
	currentTool  *llm.ToolCall
	argsBuilder  strings.Builder
}

func NewStreamHandler(handler llm.StreamHandler) StreamHandler {
	return StreamHandler{
		handler: handler,
	}
}

func (h *StreamHandler) AddContent(data []byte) error {
	var streamResp struct {
		Choices []struct {
			Delta struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"delta"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(data, &streamResp); err != nil {
		return fmt.Errorf("unmarshal stream response: %w", err)
	}

	if len(streamResp.Choices) == 0 {
		return nil
	}

	choice := streamResp.Choices[0]

	if choice.Delta.Content != "" {
		h.fullContent.WriteString(choice.Delta.Content)
		h.handler(llm.StreamResponse{
			Content: choice.Delta.Content,
			Done:    false,
		})
	}

	if len(choice.Delta.ToolCalls) > 0 {
		for _, tc := range choice.Delta.ToolCalls {
			if tc.ID != "" {
				if h.currentTool == nil || h.currentTool.ID != tc.ID {
					if h.currentTool != nil {
						h.currentTool.Arguments = helper.StringToMap(h.argsBuilder.String())
						h.toolCalls = append(h.toolCalls, *h.currentTool)
					}
					h.currentTool = &llm.ToolCall{
						ID:       tc.ID,
						Type:     tc.Type,
						Function: tc.Function.Name,
					}
					h.argsBuilder.Reset()
				}
			}
			if tc.Function.Arguments != "" {
				h.argsBuilder.WriteString(tc.Function.Arguments)
			}
		}
	}

	if choice.FinishReason != "" {
		if choice.FinishReason == "tool_calls" && h.currentTool != nil {
			h.currentTool.Arguments = helper.StringToMap(h.argsBuilder.String())
			h.toolCalls = append(h.toolCalls, *h.currentTool)
		}

		if streamResp.Usage.TotalTokens > 0 {
			h.usage = llm.Usage{
				PromptTokens:     streamResp.Usage.PromptTokens,
				CompletionTokens: streamResp.Usage.CompletionTokens,
				TotalTokens:      streamResp.Usage.TotalTokens,
			}
		}

		h.handler(llm.StreamResponse{
			Content:      h.fullContent.String(),
			FinishReason: choice.FinishReason,
			Done:         true,
			Response: &llm.CompletionResponse{
				Content:      h.fullContent.String(),
				FinishReason: choice.FinishReason,
				ToolCalls:    h.toolCalls,
				Usage:        h.usage,
			},
		})
	}

	return nil
}