package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/sjzsdu/wn/data"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/llm"
)

// Chatter 定义了项目聊天器的接口
type Chatter interface {
	NodeVisitor
	ChatWithLLM() error
}

// BaseChatter 提供了基本的聊天功能
type BaseChatter struct {
	project *Project
	cache   *data.CacheManager
	llm     llm.Provider
}

// NewBaseChatter 创建一个基本聊天器
func NewBaseChatter(p *Project) *BaseChatter {
	return &BaseChatter{
		project: p,
		cache:   data.GetDefaultCacheManager(),
		llm:     llm.GetDefaultProvider(),
	}
}

// VisitDirectory 实现通用的目录访问逻辑
func (b *BaseChatter) VisitDirectory(node *Node, path string, level int) error {
	path = b.project.GetAbsolutePath(path)
	hash, err := node.CalculateHash()
	if err != nil {
		return err
	}

	// 尝试从缓存获取
	content, found, err := b.cache.FindContent(path, hash)
	if err != nil {
		return err
	}
	if found {
		node.SetLLMResponse(content)
		return nil
	}

	// 缓存未命中，调用 LLM
	messages := PrepareDirectoryMessage(path)

	childrenResponses, err := node.GetChildrenResponses()
	if err != nil {
		// 如果是空内容错误，设置特殊响应并返回
		node.LLMResponse = NewNotProgramResponse()
		return nil
	}

	messages = append(messages, llm.Message{
		Role:    "user",
		Content: childrenResponses,
	})
	resContent, err := b.ensureValidJSONResponse(context.Background(), messages)
	if err != nil {
		fmt.Printf("Directory:%s, %s", path, resContent)
		return err
	}
	node.SetLLMResponse(resContent)
	b.cache.SetRecord(path, hash, resContent).Close()
	return nil
}

// VisitFile 实现通用的文件访问逻辑
func (b *BaseChatter) VisitFile(node *Node, path string, level int) error {
	path = b.project.GetAbsolutePath(path)
	hash, err := node.CalculateHash()
	if err != nil {
		return err
	}
	// 尝试从缓存获取
	content, found, err := b.cache.FindContent(path, hash)
	if err != nil {
		return err
	}
	if found {
		node.SetLLMResponse(content)
		return nil
	}

	if !helper.IsProgramFile(path) {
		node.LLMResponse = NewNotProgramResponse()
		return nil
	}

	// 缓存未命中，调用 LLM
	messages := PrepareFileMessage(path)
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: string(node.Content),
	})
	resContent, err := b.ensureValidJSONResponse(context.Background(), messages)
	if err != nil {
		fmt.Printf("File:%s, %s", path, resContent)
		return err
	}
	node.SetLLMResponse(resContent)
	// 保存到缓存
	b.cache.SetRecord(path, hash, resContent).Close()
	return nil
}

// ensureValidJSONResponse 确保获取到有效的JSON响应
func (b *BaseChatter) ensureValidJSONResponse(ctx context.Context, messages []llm.Message) (string, error) {
	req := llm.CompletionRequest{
		Messages: messages,
	}

	resp, err := b.llm.Complete(ctx, req)
	if err != nil {
		return "", err
	}

	// 处理返回内容中的 JSON 代码块
	content := resp.Content
	if strings.Contains(content, "```json") && strings.Contains(content, "```") {
		// 提取 ```json 和 ``` 之间的内容
		start := strings.Index(content, "```json") + 7
		end := strings.LastIndex(content, "```")
		if start < end {
			content = strings.TrimSpace(content[start:end])
		}
	}
	if _, err := NewLLMResponse(content); err != nil {
		return "", fmt.Errorf("无效的 JSON 响应: %v, %s", err, content)
	}

	return content, nil
}

func (b *BaseChatter) ChatWithLLM() error {
	totalNodes := b.project.GetTotalNodes()
	progress := helper.NewProgress("处理项目文件", totalNodes)
	progress.Show()
	defer func() {
		fmt.Println()
	}()

	visitor := &progressVisitor{
		BaseChatter: b,
		progress:    progress,
	}

	traverser := NewTreeTraverser(b.project)
	return traverser.SetTraverseOrder(PostOrder).TraverseTree(visitor)
}

// progressVisitor 包装了 BaseChatter，添加了进度更新功能
type progressVisitor struct {
	*BaseChatter
	progress *helper.Progress
}

func (p *progressVisitor) VisitDirectory(node *Node, path string, level int) error {
	err := p.BaseChatter.VisitDirectory(node, path, level)
	p.progress.Increment()
	return err
}

func (p *progressVisitor) VisitFile(node *Node, path string, level int) error {
	err := p.BaseChatter.VisitFile(node, path, level)
	p.progress.Increment()
	return err
}
