package process

import "testing"

func TestStateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{Stopped, "Stopped"},
		{Starting, "Starting"},
		{Running, "Running"},
		{Backoff, "Backoff"},
		{Stopping, "Stopping"},
		{Exited, "Exited"},
		{Fatal, "Fatal"},
		{Unknown, "Unknown"},
		{State(999), "Unknown"}, // 测试未定义的状态
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("State.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
