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
	LastContent      string // 用于跟踪上一次的内容
}

func NewStreamHandler(handler llm.StreamHandler) StreamHandler {
	return StreamHandler{
		StreamHandler: handler,
	}
}

func (h *StreamHandler) AddContent(data []byte) error {
	// 处理数据行，只取最后一个有效的JSON数据
	line := strings.TrimSpace(string(data))
	if !strings.HasPrefix(line, "data:") {
		return nil
	}

	jsonData := strings.TrimPrefix(line, "data:")
	jsonData = strings.TrimSpace(jsonData)

	if jsonData == "" || jsonData == "[DONE]" {
		return nil
	}

	var streamResp QwenResponse
	if err := json.Unmarshal([]byte(jsonData), &streamResp); err != nil {
		if share.GetDebug() {
			helper.PrintWithLabel("[DEBUG] Stream Response Parse Error", err)
		}
		return err
	}

	if share.GetDebug() {
		helper.PrintWithLabel("[DEBUG] Stream Response", streamResp)
	}

	if streamResp.Output != nil {
		// 计算增量内容
		currentText := strings.TrimSpace(streamResp.Output.Text)
		incrementalContent := ""

		if len(currentText) > len(h.LastContent) {
			if strings.HasPrefix(currentText, h.LastContent) {
				incrementalContent = strings.TrimSpace(currentText[len(h.LastContent):])
			} else {
				incrementalContent = currentText
			}
		}

		if incrementalContent != "" {
			h.StreamHandler(llm.StreamResponse{
				Content:      incrementalContent,
				Done:         false,
				FinishReason: streamResp.Output.FinishReason,
			})
			h.ContentBuilder.WriteString(incrementalContent)
			h.LastContent = currentText
		}

		// 处理工具调用
		if streamResp.Output.ToolCalls != nil && len(streamResp.Output.ToolCalls) > 0 {
			// 重置工具调用状态
			h.ToolCalls = nil
			h.CurrentToolCall = nil
			h.ArgumentsBuilder.Reset()

			for _, tc := range streamResp.Output.ToolCalls {
				toolCall := llm.ToolCall{
					ID:        tc.ID,
					Type:      tc.Type,
					Function:  tc.Function.Name,
					Arguments: helper.StringToMap(tc.Function.Arguments),
				}
				h.ToolCalls = append(h.ToolCalls, toolCall)
			}

			// 发送工具调用通知
			h.StreamHandler(llm.StreamResponse{
				Content: "",
				Done:    false,
			})
		}

		// 处理结束
		if streamResp.Output.FinishReason != "" && streamResp.Output.FinishReason != "null" {
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
