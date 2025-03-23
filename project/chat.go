package project

import (
	"context"
	"fmt"

	"github.com/sjzsdu/wn/data"
	"github.com/sjzsdu/wn/helper"
	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/share"
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
	fmt.Println("visit directory:", path)

	// 缓存未命中，调用 LLM
	messages := PrepareDirectoryMessage(path)
	
	childrenResponses, err := node.GetChildrenResponses()
	if err != nil {
		// 如果是空内容错误，设置特殊响应并返回
		node.SetLLMResponse(fmt.Sprintf("目录分析跳过: %v", err))
		return nil
	}
	
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: "请分析这个目录结构：" + childrenResponses,
	})
	req := llm.CompletionRequest{
		Messages: messages,
	}

	resp, err := b.llm.Complete(context.Background(), req)
	if err != nil {
		return err
	}
	node.SetLLMResponse(resp.Content)
	b.cache.SetRecord(path, hash, resp.Content).Close()
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
		node.SetLLMResponse(share.NOT_PROGRAM_TIP)
		return nil
	}

	fmt.Println("visit file:", path)

	// 缓存未命中，调用 LLM
	messages := PrepareFileMessage(path)
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: string(node.Content),
	})
	req := llm.CompletionRequest{
		Messages: messages,
	}

	resp, err := b.llm.Complete(context.Background(), req)
	if err != nil {
		return err
	}
	node.SetLLMResponse(resp.Content)
	// 保存到缓存
	b.cache.SetRecord(path, hash, resp.Content).Close()
	return nil
}

func (b *BaseChatter) ChatWithLLM() error {
	traverser := NewTreeTraverser(b.project)
	return traverser.SetTraverseOrder(PostOrder).TraverseTree(b)
}
