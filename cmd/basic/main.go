package main

import (
	"time"

	"github.com/darkit/process"
	"github.com/darkit/slog"
)

func main() {
	manager := process.NewManager()

	// Create a long-running process
	proc, err := manager.NewProcess(
		process.WithName("test-process"),
		process.WithCommand("sleep"),
		process.WithArgs([]string{"infinity"}),
		process.WithAutoReStart(process.AutoReStartTrue),
		process.WithStdoutLog("logs/test.log", "50MB"),
	)
	if err != nil {
		slog.Fatal(err.Error())
	}

	// Start the process
	proc.Start(true)

	// Wait for some time
	time.Sleep(10 * time.Second)

	// Stop the process
	proc.Stop(true)
}
