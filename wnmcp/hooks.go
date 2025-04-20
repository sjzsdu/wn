package wnmcp

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
)

// LogHook 实现了 Hook 接口，用于打印请求和响应日志
type LogHook struct {
	prefix string
}

// NewLogHook 创建一个新的日志钩子
func NewLogHook(prefix string) *LogHook {
	return &LogHook{
		prefix: prefix,
	}
}

func (h *LogHook) BeforeRequest(ctx context.Context, method string, args interface{}) {
	log.Printf("%s [%s] 开始请求: %+v", h.prefix, method, args)
}

func (h *LogHook) AfterRequest(ctx context.Context, method string, response interface{}, err error) {
	if err != nil {
		log.Printf("%s [%s] 请求失败: %v", h.prefix, method, err)
	} else {
		log.Printf("%s [%s] 请求成功: %+v", h.prefix, method, response)
	}
}

func (h *LogHook) OnNotification(notification mcp.JSONRPCNotification) {
	log.Printf("%s 收到通知: %+v", h.prefix, notification)
}

// CompositeHook 组合多个 Hook
type CompositeHook struct {
	hooks []Hook
}

// NewCompositeHook 创建一个新的组合钩子
func NewCompositeHook(hooks ...Hook) *CompositeHook {
	return &CompositeHook{
		hooks: hooks,
	}
}

func (h *CompositeHook) BeforeRequest(ctx context.Context, method string, args interface{}) {
	for _, hook := range h.hooks {
		hook.BeforeRequest(ctx, method, args)
	}
}

func (h *CompositeHook) AfterRequest(ctx context.Context, method string, response interface{}, err error) {
	for _, hook := range h.hooks {
		hook.AfterRequest(ctx, method, response, err)
	}
}

func (h *CompositeHook) OnNotification(notification mcp.JSONRPCNotification) {
	for _, hook := range h.hooks {
		hook.OnNotification(notification)
	}
}