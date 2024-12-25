package process

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetShell(t *testing.T) {
	shell := getShell()

	if shell == "" {
		t.Error("getShell() returned empty string")
	}

	switch runtime.GOOS {
	case "windows":
		if !strings.Contains(strings.ToLower(shell), "cmd.exe") {
			t.Errorf("Expected cmd.exe on Windows, got %s", shell)
		}
	default:
		if !strings.Contains(shell, "sh") && !strings.Contains(shell, "bash") {
			t.Errorf("Expected sh or bash on Unix-like systems, got %s", shell)
		}
	}
}

func TestGetShellOption(t *testing.T) {
	opt := getShellOption()

	switch runtime.GOOS {
	case "windows":
		if opt != "/c" {
			t.Errorf("Expected /c on Windows, got %s", opt)
		}
	default:
		if opt != "-c" {
			t.Errorf("Expected -c on Unix-like systems, got %s", opt)
		}
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		expected []string
	}{
		{
			name:     "Simple command",
			cmd:      "ls -l",
			expected: []string{"ls", "-l"},
		},
		{
			name:     "Quoted arguments",
			cmd:      `echo "hello world"`,
			expected: []string{"echo", "hello world"},
		},
		{
			name:     "Multiple quotes",
			cmd:      `echo "first part" 'second part'`,
			expected: []string{"echo", "first part", "second part"},
		},
		{
			name:     "Escaped quotes",
			cmd:      `echo "hello \"world\""`,
			expected: []string{"echo", `hello "world"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCommand(tt.cmd)

			if runtime.GOOS == "windows" {
				// Windows 特殊处理
				if len(got) != len(tt.expected) {
					t.Errorf("parseCommand() got %v, want %v", got, tt.expected)
				}
			} else {
				// Unix-like系统应该返回单个元素的切片
				if len(got) != 1 || got[0] != tt.cmd {
					t.Errorf("parseCommand() got %v, want %v", got, []string{tt.cmd})
				}
			}
		})
	}
}
