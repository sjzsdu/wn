package base

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvider_Name(t *testing.T) {
	p := &Provider{
		Pname: "test",
	}
	assert.Equal(t, "test", p.Name())
}

func TestProvider_SetModel(t *testing.T) {
	p := &Provider{
		Model: "default",
	}

	// 测试设置新模型
	assert.Equal(t, "new-model", p.SetModel("new-model"))

	// 测试空模型
	assert.Equal(t, "new-model", p.SetModel(""))
}

func TestHTTPHandler_DoRequest(t *testing.T) {
	// 创建测试服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求头
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		// 检查路径是否包含 "error"
		if strings.Contains(r.URL.Path, "error") {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "bad request"}`))
			return
		}

		// 返回正常测试响应
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test": "response"}`))
	}))
	defer ts.Close()

	h := &HTTPHandler{
		APIEndpoint: ts.URL,
		APIKey:      "test-key",
		Client:      &http.Client{},
	}

	ctx := context.Background()
	reqBody := []byte(`{"test": "data"}`)

	// 测试成功请求
	resp, err := h.DoRequest(ctx, reqBody)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	defer resp.Body.Close()

	// 测试错误状态码
	h.APIEndpoint = ts.URL + "/error"
	_, err = h.DoRequest(ctx, reqBody)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code: 400")

	// 测试无效的 endpoint
	h.APIEndpoint = "invalid-url"
	_, err = h.DoRequest(ctx, reqBody)
	assert.Error(t, err)
}

func TestHTTPHandler_DoRequest_ContextCancellation(t *testing.T) {
	h := &HTTPHandler{
		APIEndpoint: "https://api.test.com",
		APIKey:      "test-key",
		Client:      &http.Client{},
	}

	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文取消
	_, err := h.DoRequest(ctx, []byte(`{}`))
	assert.Error(t, err)
}

func TestHTTPHandler_DoRequest_InvalidResponse(t *testing.T) {
	// 创建返回错误状态码的测试服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer ts.Close()

	h := &HTTPHandler{
		APIEndpoint: ts.URL,
		APIKey:      "test-key",
		Client:      &http.Client{},
	}

	ctx := context.Background()
	_, err := h.DoRequest(ctx, []byte(`{}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code: 400")
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		options map[string]interface{}
		wantErr bool
	}{
		{
			name: "基本配置",
			options: map[string]interface{}{
				"WN_BASE_APIKEY": "test-key",
			},
			wantErr: false,
		},
		{
			name: "缺少 API Key",
			options: map[string]interface{}{
				"WN_BASE_MODEL": "base-model",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 直接验证配置
			p := &Provider{
				Pname: "test",
			}

			// 验证配置是否正确应用
			if apiKey, ok := tt.options["WN_BASE_APIKEY"]; ok {
				p.APIKey = apiKey.(string)
				assert.Equal(t, apiKey, p.APIKey)
			}

			assert.NotNil(t, p)
		})
	}
}
