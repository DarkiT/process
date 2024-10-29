package main

import (
	"time"

	"github.com/darkit/process"
	"github.com/darkit/slog"
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
			process.ProcName(p.name),
			process.ProcCommand(p.command),
			process.ProcArgs(p.args...),
			process.ProcAutoReStart(process.AutoReStartTrue),
			process.ProcStdoutLog("logs/"+p.name+".log", "50MB"),
			process.ProcStderrLog("logs/"+p.name+".err", "50MB"),
		)
		if err != nil {
			slog.Fatalf("Failed to create process %s: %v", p.name, err)
		}

		proc.Start(true)
	}

	// Wait for interruption
	time.Sleep(24 * time.Hour)
}
