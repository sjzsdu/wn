package openai

type CompletionRequestBody struct {
    Model          string    `json:"model"`
    Messages       []Message `json:"messages"`
    MaxTokens      int       `json:"max_tokens"`
    Stream         bool      `json:"stream"`
    Tools         []Tool    `json:"tools,omitempty"`
    ResponseFormat ResponseFormat `json:"response_format,omitempty"`
}

type Message struct {
    Role       string     `json:"role"`
    Content    string     `json:"content"`
    ToolCallID string     `json:"tool_call_id,omitempty"`
    Name       string     `json:"name,omitempty"`
    ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
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

type ToolCall struct {
    ID       string       `json:"id"`
    Type     string       `json:"type"`
    Function CallFunction `json:"function"`
}

type CallFunction struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments"`
}

type ResponseFormat struct {
    Type string `json:"type,omitempty"`
}

type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}

type Choice struct {
    FinishReason string  `json:"finish_reason"`
    Index        int     `json:"index"`
    Message      Message `json:"message"`
}

type OpenAIResponse struct {
    ID      string   `json:"id"`
    Object  string   `json:"object"`
    Created int      `json:"created"`
    Model   string   `json:"model"`
    Choices []Choice `json:"choices"`
    Usage   Usage    `json:"usage"`
}