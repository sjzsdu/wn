package cmd

import (
	"context"
	"encoding/json" // 添加 JSON 包的导入
	"fmt"
	"strings"

	"github.com/sjzsdu/wn/aigc"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/lang"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/share"
	"github.com/spf13/cobra"
)

var blogCmd = &cobra.Command{
	Use:   "blog",
	Short: lang.T("Blog Writer"),
	Long:  lang.T("Blog Writer is a tool for writing blog"),
	Run:   runBlog,
}

var blogContent = ""
var blogMeta = ""

func init() {
	rootCmd.AddCommand(blogCmd)
}

// parseUpdateOperations 解析响应内容为更新操作数组
func parseUpdateOperations(resp string) ([]helper.UpdateOperation, error) {
	resp = strings.TrimSpace(resp)
	if resp == "" {
		return nil, nil
	}

	// 处理被 ```json 包裹的内容
	if strings.HasPrefix(resp, "```json") {
		endIndex := strings.LastIndex(resp, "```")
		if endIndex > 6 { // "```json" 长度为 7
			resp = strings.TrimSpace(resp[7:endIndex])
		}
	}

	// 尝试解析为数组
	var changes []helper.UpdateOperation
	err := json.Unmarshal([]byte(resp), &changes)
	if err == nil {
		return changes, nil
	}

	// 如果解析数组失败，尝试解析单个对象
	var singleChange helper.UpdateOperation
	err = json.Unmarshal([]byte(resp), &singleChange)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// 将单个对象转换为数组返回
	return []helper.UpdateOperation{singleChange}, nil
}

func runBlog(cmd *cobra.Command, args []string) {
	if output == "" {
		fmt.Println("the output paramter is required")
		return
	}

	fileContent, err := helper.GetFileContent(output)

	if err != nil {
		fmt.Printf("failed to read file: %v\n", err)
		return
	} else {

		blogContent, blogMeta = extractMarkdownInfo(fileContent)
		if share.GetDebug() {
			fmt.Printf("meta: %s \n", helper.SubString(blogMeta, 40))
			fmt.Printf("content: %s \n", helper.SubString(blogContent, 40))
		}
		helper.UpdatePreviewContent(blogContent)
	}

	// 启动预览服务器
	previewURL := helper.StartPreviewServer(share.SERVER_PORT)
	fmt.Printf("预览地址: %s\n", previewURL)

	// 创建聊天实例
	chat, err := aigc.NewChat(aigc.ChatOptions{
		UseAgent: "blog",
		Hooks: &aigc.Hooks{
			AfterResponse: func(ctx context.Context, req *llm.CompletionRequest, resp *llm.CompletionResponse) error {
				content := resp.Content
				changes, errParse := parseUpdateOperations(content)
				if errParse != nil {
					fmt.Printf("解析响应失败: %v\n", errParse)
					return errParse
				}

				if changes != nil {
					blogContent = helper.ApplyChanges(blogContent, changes)
				}
				helper.UpdatePreviewContent(blogContent)
				return nil
			},
			BeforeGetContext: func(ctx context.Context, agentMessages []llm.Message, historyMessages []llm.Message) []llm.Message {
				blogMessages := getBlogMessages(blogContent)
				messages := append(agentMessages, blogMessages...)

				if len(historyMessages) > 0 {
					lastMsg := historyMessages[len(historyMessages)-1]
					messages = append(messages, lastMsg)
				}

				return messages
			},
		},
		Request: llm.CompletionRequest{
			ResponseFormat: "json_object",
		},
	}, nil)
	if err != nil {
		fmt.Printf("failed to initialize chat: %v\n", err)
		return
	}

	// 启动交互式会话
	ctx := context.Background()
	err = chat.StartInteractiveSession(ctx, aigc.InteractiveOptions{
		Renderer: helper.GetDefaultRenderer(),
		Debug:    false,
	})
	if err != nil {
		fmt.Printf("chat session ended with error: %v\n", err)
	}
	helper.StopPreviewServer()

	if blogContent == "" {
		return
	}
	fmt.Println("\n正在生成博客元数据...")
	blogMeta := createBlogMeta()
	fmt.Println("元数据生成完成")

	fmt.Println("正在保存文件...")
	writeError := helper.WriteFileContent(output, blogMeta+blogContent)
	if writeError != nil {
		fmt.Printf("❌ 文件保存失败: %v\n", writeError)
	} else {
		fmt.Printf("✅ 文件已保存到: %s\n", output)
	}

	if writeError != nil {
		fmt.Printf("failed to write file: %v\n", writeError)
	}
}

func extractMarkdownInfo(str string) (string, string) {
	// 统一换行符
	str = strings.ReplaceAll(str, "\r\n", "\n")

	// 检查是否以 "---" 开头
	if !strings.HasPrefix(str, "---") {
		return str, ""
	}

	// 查找第二个 "---"
	index := strings.Index(str[3:], "---")
	if index == -1 {
		return str, ""
	}

	// 计算实际位置（加上前面跳过的3个字符）
	index += 3

	// 提取元数据和内容
	meta := str[:index+3] // 包含第二个 "---"
	content := strings.TrimSpace(str[index+3:])

	return content, meta
}

func getBlogMessages(content string) []llm.Message {
	if content == "" {
		return []llm.Message{}
	}
	return []llm.Message{
		{
			Role:    "system",
			Content: "【原文档】：\n" + content,
		},
		{
			Role:    "system",
			Content: "请根据下一条的修改意见，对原文档进行修改，修改后内容请直接输出，不要输出任何其他多余的内容。",
		},
	}
}

func createBlogMeta() string {
	chat, err := aigc.NewChat(aigc.ChatOptions{
		UseAgent: "blog-meta",
	}, nil)
	if err != nil {
		fmt.Printf("failed to initialize chat: %v\n", err)
		return metaString("")
	}
	content, err := chat.Complete(context.Background(), blogContent)
	if err != nil {
		fmt.Printf("failed to initialize chat: %v\n", err)
		return metaString("")
	}
	return metaString(content)
}

func metaString(content string) string {
	str := `---
{{content}}
---
`
	return strings.ReplaceAll(str, "{{content}}", content)
}
