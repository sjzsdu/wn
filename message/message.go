package message

import (
	"sync"

	"github.com/sjzsdu/wn/llm"
)

// Manager 管理聊天消息的结构体
type Manager struct {
	messages []llm.Message
	mutex    sync.RWMutex
}

// New 创建一个新的消息管理器
func New() *Manager {
	return &Manager{
		messages: make([]llm.Message, 0),
	}
}

// Append 添加一条新消息
func (m *Manager) Append(msg llm.Message) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = append(m.messages, msg)
}

// GetAll 获取所有消息
func (m *Manager) GetAll() []llm.Message {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	messages := make([]llm.Message, len(m.messages))
	copy(messages, m.messages)
	return messages
}

// Clear 清空所有消息
func (m *Manager) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = make([]llm.Message, 0)
}

// GetRecentMessages 获取最近的n条消息
func (m *Manager) GetRecentMessages(n int) []llm.Message {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if len(m.messages) <= n {
		messages := make([]llm.Message, len(m.messages))
		copy(messages, m.messages)
		return messages
	}

	start := len(m.messages) - n
	messages := make([]llm.Message, n)
	copy(messages, m.messages[start:])
	return messages
}