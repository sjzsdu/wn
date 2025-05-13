package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/llm/providers/qwen"
	"github.com/stretchr/testify/assert"
)

func getQwenProvider() (llm.Provider, error) {
	if os.Getenv("WN_QWEN_APIKEY") == "" {
		return nil, fmt.Errorf("请设置 WN_QWEN_APIKEY 环境变量")
	}
	provider, err := qwen.New(map[string]interface{}{
		"WN_QWEN_APIKEY": os.Getenv("WN_QWEN_APIKEY"),
	})
	if err != nil {
		return nil, err
	}
	return provider, nil
}

func TestQwenBasicChat(t *testing.T) {
	provider, err := getQwenProvider()
	if err != nil {
		t.Skip("WN_QWEN_APIKEY not set, skipping test")
		return
	}
	assert.NotNil(t, provider)

	// 创建测试请求
	req := llm.CompletionRequest{
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: "你是什么模型？",
			},
		},
		Model:          "qwen-turbo",
		ResponseFormat: "text",
	}

	// 执行请求
	resp, err := provider.Complete(context.Background(), req)
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.NotEmpty(t, resp.Content)
		// 打印响应内容
		t.Logf("模型响应: %s", resp.Content)
	}
}

func TestQwenStreamChat(t *testing.T) {
	provider, err := getQwenProvider()
	if err != nil {
		t.Skip("WN_QWEN_APIKEY not set, skipping test")
		return
	}
	assert.NotNil(t, provider)

	// 创建测试请求
	req := llm.CompletionRequest{
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: "你是什么模型？",
			},
		},
		Model:          "qwen-turbo",
		ResponseFormat: "text",
	}

	// 创建一个切片来存储响应
	var responses []string

	// 修改流式聊天测试的处理器
	streamHandler := func(resp llm.StreamResponse) {
		// 只要不是完成信号就记录响应
		if resp.Content != "" {
			responses = append(responses, resp.Content)
			t.Logf("流式响应: %s", resp.Content)
		}
	}

	// 执行流式请求
	err = provider.CompleteStream(context.Background(), req, streamHandler)
	assert.NoError(t, err)
	assert.NotEmpty(t, responses)
}

func TestQwenProviderName(t *testing.T) {
	provider, err := getQwenProvider()
	if err != nil {
		t.Skip("WN_QWEN_APIKEY not set, skipping test")
		return
	}
	assert.NotNil(t, provider)

	// 测试 GetName 方法
	name := provider.GetName()
	assert.Equal(t, "qwen", name)
}

func TestQwenAvailableModels(t *testing.T) {
	provider, err := getQwenProvider()
	if err != nil {
		t.Skip("WN_QWEN_APIKEY not set, skipping test")
		return
	}
	assert.NotNil(t, provider)

	// 测试 AvailableModels 方法
	models := provider.AvailableModels()
	assert.NotEmpty(t, models)
	t.Logf("支持的模型列表: %v", models)
}

func TestQwenModelSetGet(t *testing.T) {
	provider, err := getQwenProvider()
	if err != nil {
		t.Skip("WN_QWEN_APIKEY not set, skipping test")
		return
	}
	assert.NotNil(t, provider)

	// 测试 SetModel 和 GetModel 方法
	testModel := "qwen-turbo"
	setModel := provider.SetModel(testModel)
	assert.Equal(t, testModel, setModel)

	getModel := provider.GetModel()
	assert.Equal(t, testModel, getModel)
}
