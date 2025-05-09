package base

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

func NewHTTPHandler(apiKey, endpoint string, config RequestConfig) *HTTPHandler {
	client := resty.New()

	// 基础配置
	client.SetTimeout(time.Duration(config.Timeout) * time.Second)
	client.SetHeaders(config.Headers)

	// 重试配置
	if config.RetryConfig != nil {
		client.SetRetryCount(config.RetryConfig.MaxRetries)
		client.SetRetryWaitTime(time.Duration(config.RetryConfig.RetryDelay) * time.Second)

		// 根据重试策略设置退避时间
		if config.RetryConfig.RetryPolicy == RetryPolicyExponential {
			client.SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
				return time.Duration(math.Pow(2, float64(client.RetryCount))) * time.Second, nil
			})
		}

		// 设置重试条件
		client.AddRetryCondition(func(r *resty.Response, err error) bool {
			return err != nil || r.StatusCode() >= 500
		})
	}

	// 设置请求超时和连接池
	client.SetTransport(&http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	})

	return &HTTPHandler{
		APIKey:      apiKey,
		APIEndpoint: endpoint,
		Client:      client,
		Config:      config,
	}
}

func (h *HTTPHandler) DoGet(ctx context.Context, path string, params map[string]string) (*resty.Response, error) {
	return h.Client.R().
		SetContext(ctx).
		SetQueryParams(params).
		Get(h.APIEndpoint + path)
}

func (h *HTTPHandler) DoPost(ctx context.Context, path string, body interface{}) (*resty.Response, error) {
	return h.Client.R().
		SetContext(ctx).
		SetBody(body).
		Post(h.APIEndpoint + path)
}

func (h *HTTPHandler) DoStream(ctx context.Context, path string, body interface{}) (*resty.Response, error) {
	return h.Client.R().
		SetContext(ctx).
		SetBody(body).
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetHeader("Cache-Control", "no-cache").
		Post(h.APIEndpoint + path)
}

func (h *HTTPHandler) HandleJSONResponse(resp *resty.Response, v interface{}) error {
	if resp == nil {
		return fmt.Errorf("响应对象为空")
	}

	// 检查参数有效性
	if v == nil {
		return fmt.Errorf("目标对象为空")
	}

	// 检查状态码并提供更详细的错误信息
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("HTTP请求失败: 状态码=%d, 错误=%s", resp.StatusCode(), resp.Error())
	}

	// 检查响应体
	body := resp.Body()
	if len(body) == 0 {
		return fmt.Errorf("响应体为空")
	}

	// 使用标准库的 json 包进行解析，提供更好的错误处理
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("JSON解析失败: %w", err)
	}

	return nil
}

func (h *HTTPHandler) HandleStreamResponse(resp *resty.Response, handler StreamHandler) error {
	if resp == nil {
		return fmt.Errorf("响应对象为空")
	}

	reader := resp.RawResponse.Body
	if reader == nil {
		return fmt.Errorf("响应体为空")
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	// 设置更大的缓冲区，避免长消息被截断
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		select {
		case <-resp.Request.Context().Done():
			return fmt.Errorf("请求被取消: %w", resp.Request.Context().Err())
		default:
			data := scanner.Bytes()
			if len(data) > 0 {
				if err := handler.HandleStream(data); err != nil {
					return fmt.Errorf("处理流式数据失败: %w", err)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		if err == io.EOF {
			return nil
		}
		if err.Error() == "http: read on closed response body" {
			return nil
		}
		return fmt.Errorf("读取流式响应失败: %w", err)
	}

	return nil
}
