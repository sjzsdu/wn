package helper

import (
	"testing"
	"time"
)

func TestShowLoadingAnimation(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		{
			name:     "快速取消动画",
			duration: 100 * time.Millisecond,
		},
		{
			name:     "延迟取消动画",
			duration: 500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan bool)
			go ShowLoadingAnimation(done)

			time.Sleep(tt.duration)
			done <- true
		})
	}
}