package deepseek

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
	var streamResp StreamResponse
	if err := json.Unmarshal(data, &streamResp); err != nil {
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Stream Response Parse Error", err)
		}
		return err
	}
	if len(streamResp.Choices) > 0 {
		choice := streamResp.Choices[0]

		h.StreamHandler(llm.StreamResponse{
			Content:      choice.Delta.Content,
			Done:         false,
			FinishReason: choice.FinishReason,
		})

		if choice.Delta.Content != "" {
			h.ContentBuilder.WriteString(choice.Delta.Content)
		}

		if len(choice.Delta.ToolCalls) > 0 {
			for _, tc := range choice.Delta.ToolCalls {
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
		if choice.FinishReason != "" {
			// 处理最后一个工具调用
			if choice.FinishReason == "tool_calls" && h.CurrentToolCall != nil {
				h.CurrentToolCall.Arguments = helper.StringToMap(h.ArgumentsBuilder.String())
				h.ToolCalls = append(h.ToolCalls, *h.CurrentToolCall)
				// 发送最后一个工具调用完成通知
				h.StreamHandler(llm.StreamResponse{
					Content: "",
					Done:    false,
				})
			}

			if streamResp.Usage.TotalTokens > 0 {
				h.Usage = llm.Usage{
					PromptTokens:     streamResp.Usage.PromptTokens,
					CompletionTokens: streamResp.Usage.CompletionTokens,
					TotalTokens:      streamResp.Usage.TotalTokens,
				}
			}

			h.StreamHandler(llm.StreamResponse{
				Content:      h.ContentBuilder.String(),
				FinishReason: choice.FinishReason,
				Done:         true,
				Response: &llm.CompletionResponse{
					Content:      h.ContentBuilder.String(),
					FinishReason: choice.FinishReason,
					ToolCalls:    h.ToolCalls,
					Usage:        h.Usage,
				},
			})
		}
	}
	return nil
}
