package qwen

type QwenRequest struct {
	Model   string        `json:"model"`
	Input   Input        `json:"input"`
	Parameters Parameters `json:"parameters,omitempty"`
}

type Input struct {
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
}

type Parameters struct {
	Stream           bool    `json:"stream,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
	TopP             float64 `json:"top_p,omitempty"`
	TopK             int     `json:"top_k,omitempty"`
	MaxTokens        int     `json:"max_tokens,omitempty"`
	Stop             []string `json:"stop,omitempty"`
	ResultFormat     string   `json:"result_format,omitempty"`
	IncrementalOutput bool    `json:"incremental_output,omitempty"`
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`
	ToolCallId string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Description string                 `json:"description,omitempty"`
	Name        string                 `json:"name"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

type CallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function CallFunction `json:"function"`
}

type QwenResponse struct {
	Output *Output `json:"output"`
	Usage  Usage   `json:"usage"`
}

type Output struct {
	Text         string     `json:"text"`
	FinishReason string     `json:"finish_reason"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}