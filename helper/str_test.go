package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringSliceContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "空切片",
			slice:    []string{},
			item:     "test",
			expected: false,
		},
		{
			name:     "包含目标字符串",
			slice:    []string{"test1", "test2", "test3"},
			item:     "test2",
			expected: true,
		},
		{
			name:     "不包含目标字符串",
			slice:    []string{"test1", "test2", "test3"},
			item:     "test4",
			expected: false,
		},
		{
			name:     "空字符串测试",
			slice:    []string{"test1", "", "test3"},
			item:     "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringSliceContains(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}