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
		process.ProcName("test-process"),
		process.ProcCommand("sleep"),
		process.ProcArgs("infinity"),
		process.ProcAutoReStart(process.AutoReStartTrue),
		process.ProcStdoutLog("logs/test.log", "50MB"),
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
