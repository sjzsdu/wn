package qwen

import (
	"encoding/json"
	"strings"

	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/share"
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
	var streamResp QwenResponse
	if err := json.Unmarshal(data, &streamResp); err != nil {
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Stream Response Parse Error", err)
		}
		return err
	}
	helper.PrintWithLabel("[DEBUG] Stream AddContent", string(data), streamResp)

	if streamResp.Output != nil {
		h.StreamHandler(llm.StreamResponse{
			Content:      streamResp.Output.Text,
			Done:         false,
			FinishReason: streamResp.Output.FinishReason,
		})

		if streamResp.Output.Text != "" {
			h.ContentBuilder.WriteString(streamResp.Output.Text)
		}

		if streamResp.Output.ToolCalls != nil && len(streamResp.Output.ToolCalls) > 0 {
			for _, tc := range streamResp.Output.ToolCalls {
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
				}
				// 累积参数字符串
				if tc.Function.Arguments != "" {
					h.ArgumentsBuilder.WriteString(tc.Function.Arguments)
				}
			}
		}

		// 处理结束原因
		if streamResp.Output.FinishReason != "" {
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

			if streamResp.Usage.InputTokens > 0 || streamResp.Usage.OutputTokens > 0 {
				h.Usage = llm.Usage{
					PromptTokens:     streamResp.Usage.InputTokens,
					CompletionTokens: streamResp.Usage.OutputTokens,
					TotalTokens:      streamResp.Usage.InputTokens + streamResp.Usage.OutputTokens,
				}
			}

			h.StreamHandler(llm.StreamResponse{
				Content:      h.ContentBuilder.String(),
				FinishReason: streamResp.Output.FinishReason,
				Done:         true,
				Response: &llm.CompletionResponse{
					Content:      h.ContentBuilder.String(),
					FinishReason: streamResp.Output.FinishReason,
					ToolCalls:    h.ToolCalls,
					Usage:        h.Usage,
				},
			})
		}
	}
	return nil
}
