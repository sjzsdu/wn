package deepseek

type DeepseekRequest struct {
	Messages         []Message      `json:"messages"`
	Model            string         `json:"model"`
	FrequencyPenalty float64        `json:"frequency_penalty,omitempty"`
	MaxTokens        int            `json:"max_tokens,omitempty"`
	PresencePenalty  float64        `json:"presence_penalty,omitempty"`
	ResponseFormat   ResponseFormat `json:"response_format,omitempty"`
	Stop             interface{}    `json:"stop,omitempty"`
	Stream           bool           `json:"stream,omitempty"`
	StreamOptions    interface{}    `json:"stream_options,omitempty"`
	Temperature      float64        `json:"temperature,omitempty"`
	TopP             float64        `json:"top_p,omitempty"`
	Tools            []Tool         `json:"tools,omitempty"`
	ToolChoice       interface{}    `json:"tool_choice,omitempty"`
	LogitBias        map[string]int `json:"logit_bias,omitempty"`
}
type Message struct {
	Role             string     `json:"role"`
	Content          string     `json:"content"`
	Name             string     `json:"name,omitempty"`
	ReasoningContent string     `json:"reasoning_content,omitempty"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
	ToolCallId       string     `json:"tool_call_id,omitempty"`
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

type CallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments,omitempty"`
}

type ToolCall struct {
	Index    int          `json:"index"`
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function CallFunction `json:"function"`
}

type Delta struct {
	Role             string     `json:"role,omitempty"`
	Content          string     `json:"content"`
	ReasoningContent string     `json:"reasoning_content,omitempty"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
}

type TopLogProb struct {
	Token   string  `json:"token"`
	LogProb float64 `json:"logprob"`
	Bytes   []int   `json:"bytes"`
}

type ContentLogProb struct {
	Token       string       `json:"token"`
	LogProb     float64      `json:"logprob"`
	Bytes       []int        `json:"bytes"`
	TopLogProbs []TopLogProb `json:"top_logprobs"`
}

type LogProbs struct {
	Content []ContentLogProb `json:"content"`
}

type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

type Usage struct {
	CompletionTokens        int                     `json:"completion_tokens"`
	PromptTokens            int                     `json:"prompt_tokens"`
	PromptCacheHitTokens    int                     `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens   int                     `json:"prompt_cache_miss_tokens"`
	TotalTokens             int                     `json:"total_tokens"`
	CompletionTokensDetails CompletionTokensDetails `json:"completion_tokens_details"`
	PromptTokensDetails     map[string]interface{}  `json:"prompt_tokens_details"` // 新增字段
}

type Choice struct {
	FinishReason string    `json:"finish_reason"`
	Index        int       `json:"index"`
	Message      Message   `json:"message"`
	LogProbs     *LogProbs `json:"logprobs,omitempty"`
}

type StreamChoice struct {
	Delta        Delta  `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

type StreamResponse struct {
	Choices []StreamChoice `json:"choices"`
	Usage   Usage          `json:"usage,omitempty"`
}

type DeepseekResponse struct {
	ID                string   `json:"id"` // 改为大写 ID
	Object            string   `json:"object"`
	Created           int      `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
	SystemFingerprint string   `json:"system_fingerprint"`
}
