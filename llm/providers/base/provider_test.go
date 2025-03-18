package base

import (
	"io"
	"testing"

	"github.com/sjzsdu/wn/llm"
	"github.com/stretchr/testify/assert"
)

type mockParser struct{}

func (m *mockParser) ParseResponse(body io.Reader) (llm.CompletionResponse, error) {
	return llm.CompletionResponse{
		Content: "test response",
		Usage: llm.Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}, nil
}

func TestProvider_SetParser(t *testing.T) {
	p := &Provider{}
	parser := &mockParser{}
	p.SetParser(parser)
	assert.Equal(t, parser, p.parser)
}

func TestProvider_Name(t *testing.T) {
	p := &Provider{Pname: "test"}
	assert.Equal(t, "test", p.Name())
}

func TestProvider_AvailableModels(t *testing.T) {
	models := []string{"model1", "model2"}
	p := &Provider{Models: models}
	assert.Equal(t, models, p.AvailableModels())
}

func TestProvider_SetModel(t *testing.T) {
	p := &Provider{Model: "default"}

	// 测试设置新模型
	assert.Equal(t, "new-model", p.SetModel("new-model"))

	// 测试空模型
	assert.Equal(t, "new-model", p.SetModel(""))
}
