package claude

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		options map[string]interface{}
		wantErr bool
	}{
		{
			name: "基本配置",
			options: map[string]interface{}{
				"WN_CLAUDE_APIKEY": "test-key",
			},
			wantErr: false,
		},
		{
			name: "缺少 API Key",
			options: map[string]interface{}{
				"WN_CLAUDE_MODEL": "claude-2",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, p)
		})
	}
}
