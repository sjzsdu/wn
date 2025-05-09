package claude

type CompletionRequestBody struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
	Stream    bool      `json:"stream"`
	Tools     []Tool    `json:"tools,omitempty"` // 添加工具支持
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function CallFunction `json:"function"`
}

type CallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type StreamResponse struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Role         string    `json:"role"`
	Content      []Content `json:"content"`
	Model        string    `json:"model"`
	StopReason   string    `json:"stop_reason,omitempty"`
	StopSequence string    `json:"stop_sequence,omitempty"`
	Usage        Usage     `json:"usage,omitempty"`
}

type Content struct {
	Text      string     `json:"text"`
	Type      string     `json:"type"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
