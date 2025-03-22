package project

import (
	"github.com/sjzsdu/wn/agent"
	"github.com/sjzsdu/wn/llm"
)

func PrepareFileMessage(file string) []llm.Message {
	return agent.GetAgentMessages("file")
}

func PrepareDirectoryMessage(path string) []llm.Message {
	return agent.GetAgentMessages("directory")
}
