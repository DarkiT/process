package main

import (
	"os"
	"time"

	"github.com/darkit/process"
	"github.com/darkit/slog"
)

func init() {
	slog.SetLevelDebug()
	slog.NewLogger(os.Stdout, false, true)
}

func main() {
	manager := process.NewManager()

	// Create a long-running process
	proc, err := manager.NewProcess(
		process.WithName("test-timeout"),
		process.WithCommand("sleep"),
		process.WithArgs("infinity"),
		process.WithAutoReStart(process.AutoReStartFalse),
		// process.WithStdoutLog("logs/test.log", "50MB"),
	)
	if err != nil {
		slog.Fatal(err.Error())
	}

	// Start the process
	proc.Start(false)

	// Wait for some time
	time.Sleep(30 * time.Second)

	// Stop the process
	proc.Stop(false)
}
