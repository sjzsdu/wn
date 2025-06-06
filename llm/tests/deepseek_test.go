package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sjzsdu/wn/llm"
	"github.com/sjzsdu/wn/llm/providers/deepseek"
	"github.com/stretchr/testify/assert"
)

func getDeepseekProvider() (llm.Provider, error) {
	if os.Getenv("WN_DEEPSEEK_APIKEY") == "" {
		return nil, fmt.Errorf("请设置 WN_DEEPSEEK_APIKEY 环境变量")
	}
	provider, err := deepseek.New(map[string]interface{}{
		"WN_DEEPSEEK_APIKEY": os.Getenv("WN_DEEPSEEK_APIKEY"),
	})
	if err != nil {
		return nil, err
	}
	return provider, nil
}

func TestDeepseekBasicChat(t *testing.T) {
	provider, err := getDeepseekProvider()
	if err != nil {
		t.Skip("WN_DEEPSEEK_APIKEY not set, skipping test")
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
		Model:          "deepseek-chat",
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

func TestDeepseekStreamChat(t *testing.T) {
	provider, err := getDeepseekProvider()
	if err != nil {
		t.Skip("WN_DEEPSEEK_APIKEY not set, skipping test")
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
		Model:          "deepseek-chat",
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

func TestDeepseekProviderName(t *testing.T) {
	provider, err := getDeepseekProvider()
	if err != nil {
		t.Skip("WN_DEEPSEEK_APIKEY not set, skipping test")
		return
	}
	assert.NotNil(t, provider)

	// 测试 GetName 方法
	name := provider.GetName()
	assert.Equal(t, "deepseek", name)
}

func TestDeepseekAvailableModels(t *testing.T) {
	provider, err := getDeepseekProvider()
	if err != nil {
		t.Skip("WN_DEEPSEEK_APIKEY not set, skipping test")
		return
	}
	assert.NotNil(t, provider)

	// 测试 AvailableModels 方法
	models := provider.AvailableModels()
	assert.NotEmpty(t, models)
	t.Logf("支持的模型列表: %v", models)
}

func TestDeepseekModelSetGet(t *testing.T) {
	provider, err := getDeepseekProvider()
	if err != nil {
		t.Skip("WN_DEEPSEEK_APIKEY not set, skipping test")
		return
	}
	assert.NotNil(t, provider)

	// 测试 SetModel 和 GetModel 方法
	testModel := "deepseek-chat"
	setModel := provider.SetModel(testModel)
	assert.Equal(t, testModel, setModel)

	getModel := provider.GetModel()
	assert.Equal(t, testModel, getModel)
}
