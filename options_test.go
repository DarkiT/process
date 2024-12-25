package process

import (
	"testing"
)

func TestNewOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     []WithOption
		expected Options
	}{
		{
			name: "Default options",
			opts: nil,
			expected: Options{
				AutoStart:    true,
				StartSecs:    1,
				AutoReStart:  AutoReStartTrue,
				StartRetries: 3,
				Priority:     999,
			},
		},
		{
			name: "Custom options",
			opts: []WithOption{
				WithName("test-process"),
				WithCommand("/bin/echo"),
				WithArgs("hello"),
				WithAutoStart(false),
				WithStartRetries(5),
				WithPriority(100),
			},
			expected: Options{
				Name:         "test-process",
				Command:      "/bin/echo",
				Args:         []string{"hello"},
				AutoStart:    false,
				StartSecs:    1,
				AutoReStart:  AutoReStartTrue,
				StartRetries: 5,
				Priority:     100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOptions(tt.opts...)

			if got.Name != tt.expected.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.expected.Name)
			}

			if got.Command != tt.expected.Command {
				t.Errorf("Command = %v, want %v", got.Command, tt.expected.Command)
			}

			if got.AutoStart != tt.expected.AutoStart {
				t.Errorf("AutoStart = %v, want %v", got.AutoStart, tt.expected.AutoStart)
			}

			if got.StartRetries != tt.expected.StartRetries {
				t.Errorf("StartRetries = %v, want %v", got.StartRetries, tt.expected.StartRetries)
			}

			if got.Priority != tt.expected.Priority {
				t.Errorf("Priority = %v, want %v", got.Priority, tt.expected.Priority)
			}
		})
	}
}
