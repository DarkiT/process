package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/darkit/process"
)

func main() {
	manager := process.NewManager()

	// Create multiple processes
	processes := []struct {
		name    string
		command string
		args    []string
	}{
		{"web", "nginx", []string{"-g", "daemon off;"}},
		{"app", "./myapp", []string{"--port", "8080"}},
		{"worker", "./worker", []string{"--queue", "default"}},
	}

	for _, p := range processes {
		proc, err := manager.NewProcess(
			process.WithName(p.name),
			process.WithCommand(p.command),
			process.WithArgs(p.args...),
			process.WithAutoReStart(process.AutoReStartTrue),
			process.WithStdoutLog("logs/"+p.name+".log", "50MB"),
			process.WithStderrLog("logs/"+p.name+".err", "50MB"),
		)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to create process %s: %v", p.name, err))
		}

		proc.Start(true)
	}

	// Wait for interruption
	time.Sleep(24 * time.Hour)
}
