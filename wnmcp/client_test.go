package wnmcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sjzsdu/wn/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewClient(t *testing.T) {
	mockConn := new(MockMCPClient)
	project := &project.Project{}

	// Setup mock expectations for Initialize call
	expectedResult := &mcp.InitializeResult{
		ProtocolVersion: "1.0",
		Capabilities: mcp.ServerCapabilities{},
		ServerInfo: mcp.Implementation{
			Name:    "test-server",
			Version: "1.0",
		},
	}
	mockConn.On("Initialize", mock.Anything, mock.Anything).Return(expectedResult, nil)

	client := NewClient(mockConn, project)
	assert.NotNil(t, client)
	assert.Equal(t, project, client.project)
	assert.Equal(t, mockConn, client.conn)

	// Verify mock expectations
	mockConn.AssertExpectations(t)
}

func TestClient_Initialize(t *testing.T) {
	mockConn := new(MockMCPClient)
	project := &project.Project{}

	// Setup initial mock expectations for NewClient's internal Initialize call
	initResult := &mcp.InitializeResult{
		ProtocolVersion: "1.0",
		Capabilities: mcp.ServerCapabilities{},
		ServerInfo: mcp.Implementation{
			Name:    "test-server",
			Version: "1.0",
		},
	}
	mockConn.On("Initialize", mock.Anything, mock.Anything).Return(initResult, nil)

	// Create client (this will trigger the internal Initialize call)
	client := NewClient(mockConn, project)

	// Setup mock expectations for the actual test call
	request := NewInitializeRequest()
	expectedResult := &mcp.InitializeResult{
		ProtocolVersion: "1.0",
		Capabilities: mcp.ServerCapabilities{},
		ServerInfo: mcp.Implementation{
			Name:    "test-server",
			Version: "1.0",
		},
	}
	mockConn.On("Initialize", mock.Anything, request).Return(expectedResult, nil)

	// Execute test
	result, err := client.Initialize(context.Background(), request)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	
	// Verify all mock expectations
	mockConn.AssertExpectations(t)
}

func TestClient_WithHook(t *testing.T) {
	mockConn := new(MockMCPClient)
	project := &project.Project{}
	mockHook := new(MockHook)

	// Setup mock expectations for Initialize call
	expectedResult := &mcp.InitializeResult{
		ProtocolVersion: "1.0",
		Capabilities: mcp.ServerCapabilities{},
		ServerInfo: mcp.Implementation{
			Name:    "test-server",
			Version: "1.0",
		},
	}
	mockConn.On("Initialize", mock.Anything, mock.Anything).Return(expectedResult, nil)

	// Setup hook expectations
	mockHook.On("BeforeRequest", mock.Anything, "Initialize", mock.Anything).Return()
	mockHook.On("AfterRequest", mock.Anything, "Initialize", mock.Anything, mock.Anything).Return()

	client := NewClient(mockConn, project, WithHook(mockHook))
	assert.Equal(t, mockHook, client.hook)

	// Verify all mock expectations
	mockConn.AssertExpectations(t)
	mockHook.AssertExpectations(t)
}

// ... 为其他方法编写类似的测试用例 ...

type MockHook struct {
	mock.Mock
}

func (m *MockHook) BeforeRequest(ctx context.Context, method string, args interface{}) {
	m.Called(ctx, method, args)
}

func (m *MockHook) AfterRequest(ctx context.Context, method string, response interface{}, err error) {
	m.Called(ctx, method, response, err)
}

func (m *MockHook) OnNotification(notification mcp.JSONRPCNotification) {
	m.Called(notification)
}
