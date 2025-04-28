package deepseek

type CompletionRequestBody struct {
	Messages         []Message      `json:"messages"`
	Model            string         `json:"model"`
	FrequencyPenalty float64        `json:"frequency_penalty,omitempty"`
	MaxTokens        int            `json:"max_tokens,omitempty"`
	PresencePenalty  float64        `json:"presence_penalty,omitempty"`
	ResponseFormat   ResponseFormat `json:"response_format,omitempty"`
	Stop             interface{}    `json:"stop,omitempty"` // 可以是 string 或 []string
	Stream           bool           `json:"stream,omitempty"`
	StreamOptions    interface{}    `json:"stream_options,omitempty"`
	Temperature      float64        `json:"temperature,omitempty"`
	TopP             float64        `json:"top_p,omitempty"`
	Tools            []Tool         `json:"tools,omitempty"`
	ToolChoice       interface{}    `json:"tool_choice,omitempty"` // 可以是 string 或 SpecificToolChoice
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseFormat struct {
	Type string `json:"type,omitempty"`
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

type SpecificToolChoice struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name string `json:"name"`
}

type Choice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	FinishReason string `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Delta struct {
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
    ID       string `json:"id"`
    Type     string `json:"type"`
    Function struct {
        Name      string `json:"name"`
        Arguments string `json:"arguments"`
    } `json:"function"`
}

type StreamChoice struct {
	Delta        Delta  `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

type DeepseekResponse struct {
	ID              string   `json:"id"`
	Choices         []Choice `json:"choices"`
	Created         int      `json:"created"`
	Model           string   `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Object          string   `json:"object"`
	Usage           Usage    `json:"usage"`
}

type StreamResponse struct {
	Choices []StreamChoice `json:"choices"`
}
