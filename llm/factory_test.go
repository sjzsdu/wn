package llm

import (
	"context"
	"testing"
)

// 模拟一个Provider实现
type mockProvider struct{}

func (m *mockProvider) Chat(messages []Message) (string, error) {
	return "mock response", nil
}

// 添加AvailableModels方法以完整实现Provider接口
func (m *mockProvider) AvailableModels() []string {
	return []string{"mock-model"}
}

// 添加Name方法以完整实现Provider接口
func (m *mockProvider) GetName() string {
	return "mock"
}

func (m *mockProvider) SetModel(model string) string {
	return "mock-model"
}

func (m *mockProvider) GetModel() string {
	return "mock-model"
}

// 修正Complete方法的签名以匹配Provider接口
func (m *mockProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	return &CompletionResponse{
		Content: "mock completion",
	}, nil
}

// 实现CompleteStream方法
func (m *mockProvider) CompleteStream(ctx context.Context, req CompletionRequest, handler StreamHandler) error {
	handler(StreamResponse{
		Content: "mock stream content",
		Done:    true,
	})
	return nil
}

// 添加HandleRequestBody方法以完整实现Provider接口
func (m *mockProvider) HandleRequestBody(req CompletionRequest, reqBody map[string]interface{}) interface{} {
	return reqBody
}

func TestRegisterAndCreateProvider(t *testing.T) {
	// 测试注册新的provider
	mockNew := func(options map[string]interface{}) (Provider, error) {
		return &mockProvider{}, nil
	}

	Register("mock", mockNew)

	// 测试Providers函数
	providers := Providers()
	found := false
	for _, name := range providers {
		if name == "mock" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'mock' in providers list")
	}

	// 测试创建provider
	provider, err := CreateProvider("mock", nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if provider == nil {
		t.Error("Expected provider not to be nil")
	}

	// 测试HandleRequestBody方法
	req := CompletionRequest{
		Messages: []Message{{Role: "user", Content: "test"}},
	}
	reqBody := map[string]interface{}{
		"messages": []Message{{Role: "user", Content: "test"}},
	}
	result := provider.HandleRequestBody(req, reqBody)
	if result == nil {
		t.Error("Expected HandleRequestBody result not to be nil")
	}

	// 测试创建不存在的provider
	_, err = CreateProvider("non-existent", nil)
	if err == nil {
		t.Error("Expected error when creating non-existent provider")
	}

	// 测试创建provider时name为空
	_, err = CreateProvider("", nil)
	if err == nil {
		t.Error("Expected error when creating provider with empty name")
	}
}

func TestRegisterNilProvider(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when registering nil provider")
		}
	}()
	Register("nil-provider", nil)
}

func TestRegisterDuplicateProvider(t *testing.T) {
	mockNew := func(options map[string]interface{}) (Provider, error) {
		return &mockProvider{}, nil
	}

	// 先注册一个provider
	Register("duplicate", mockNew)

	// 测试重复注册
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when registering duplicate provider")
		}
	}()
	Register("duplicate", mockNew)
}
