package base

import "github.com/go-resty/resty/v2"

// RetryPolicy 定义重试策略类型
type RetryPolicy string

const (
	// RetryPolicyExponential 指数退避重试策略
	RetryPolicyExponential RetryPolicy = "exponential"
	// RetryPolicyLinear 线性重试策略
	RetryPolicyLinear RetryPolicy = "linear"
)

// RetryConfig 定义重试配置
type RetryConfig struct {
	MaxRetries  int
	RetryDelay  int
	RetryPolicy RetryPolicy
}

// RequestConfig 定义请求配置
type RequestConfig struct {
	Timeout     int
	RetryConfig *RetryConfig
	Headers     map[string]string
}

// ResponseHandler 定义响应处理器
type ResponseHandler interface {
	HandleResponse([]byte) error
}

// StreamHandler 定义流式响应处理器
type StreamHandler interface {
	HandleStream([]byte) error
}

// MiddlewareFunc 定义中间件函数类型
type MiddlewareFunc func(RequestConfig) RequestConfig

type HTTPHandler struct {
	APIKey      string
	APIEndpoint string
	Client      *resty.Client
	Config      RequestConfig
}

// Provider 提供基础的 LLM Provider 实现
type Provider struct {
	*HTTPHandler
	Model     string
	Name      string
	MaxTokens int
}
