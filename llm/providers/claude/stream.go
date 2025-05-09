package claude

import (
	"encoding/json"
	"strings"

	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/llm"
)

type StreamHandler struct {
	StreamHandler    llm.StreamHandler
	ContentBuilder   strings.Builder
	ArgumentsBuilder strings.Builder
	CurrentToolCall  *llm.ToolCall
	ToolCalls        []llm.ToolCall
	Usage            llm.Usage
}

func NewStreamHandler(handler llm.StreamHandler) StreamHandler {
	return StreamHandler{
		StreamHandler: handler,
	}
}

func (h *StreamHandler) AddContent(data []byte) error {
	var streamResp StreamResponse
	if err := json.Unmarshal(data, &streamResp); err != nil {
		return err
	}

	if len(streamResp.Content) > 0 {
		content := streamResp.Content[0].Text
		h.StreamHandler(llm.StreamResponse{
			Content:      content,
			Done:         false,
			FinishReason: streamResp.StopReason,
		})

		if content != "" {
			h.ContentBuilder.WriteString(content)
		}
	}

	// 处理工具调用
	if len(streamResp.Content) > 0 && len(streamResp.Content[0].ToolCalls) > 0 {
		for _, tc := range streamResp.Content[0].ToolCalls {
			if tc.ID != "" {
				// 新的工具调用开始
				if h.CurrentToolCall == nil || h.CurrentToolCall.ID != tc.ID {
					if h.CurrentToolCall != nil {
						// 完成前一个工具调用
						h.CurrentToolCall.Arguments = helper.StringToMap(h.ArgumentsBuilder.String())
						h.ToolCalls = append(h.ToolCalls, *h.CurrentToolCall)
						// 发送工具调用完成通知
						h.StreamHandler(llm.StreamResponse{
							Content: "",
							Done:    false,
						})
					}
					// 创建新的工具调用
					h.CurrentToolCall = &llm.ToolCall{
						ID:       tc.ID,
						Type:     tc.Type,
						Function: tc.Function.Name,
					}
					h.ArgumentsBuilder.Reset()
					// 发送新工具调用开始通知
					h.StreamHandler(llm.StreamResponse{
						Content: "",
						Done:    false,
					})
				}
				// 累积参数字符串
				if tc.Function.Arguments != "" {
					h.ArgumentsBuilder.WriteString(tc.Function.Arguments)
				}
			}
		}
	}

	// 处理结束
	if streamResp.StopReason != "" {
		// 处理最后一个工具调用
		if h.CurrentToolCall != nil {
			h.CurrentToolCall.Arguments = helper.StringToMap(h.ArgumentsBuilder.String())
			h.ToolCalls = append(h.ToolCalls, *h.CurrentToolCall)
			// 发送最后一个工具调用完成通知
			h.StreamHandler(llm.StreamResponse{
				Content: "",
				Done:    false,
			})
		}

		if streamResp.Usage.InputTokens > 0 {
			h.Usage = llm.Usage{
				PromptTokens:     streamResp.Usage.InputTokens,
				CompletionTokens: streamResp.Usage.OutputTokens,
				TotalTokens:      streamResp.Usage.InputTokens + streamResp.Usage.OutputTokens,
			}
		}

		h.StreamHandler(llm.StreamResponse{
			Content:      h.ContentBuilder.String(),
			FinishReason: streamResp.StopReason,
			Done:         true,
			Response: &llm.CompletionResponse{
				Content:      h.ContentBuilder.String(),
				FinishReason: streamResp.StopReason,
				ToolCalls:    h.ToolCalls,
				Usage:        h.Usage,
			},
		})
	}

	return nil
}
