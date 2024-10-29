package process

import (
	"testing"
	"time"
)

func TestProcessLifecycle(t *testing.T) {
	manager := NewManager()

	proc, err := manager.NewProcessCmd("echo test", nil)
	if err != nil {
		t.Fatalf("Failed to create process: %v", err)
	}

	// Test start
	proc.Start(true)
	if !proc.isRunning() {
		t.Error("Process should be running")
	}

	// Test stop
	proc.Stop(true)
	time.Sleep(100 * time.Millisecond)
	if proc.isRunning() {
		t.Error("Process should not be running")
	}
}

func TestProcessAutoRestart(t *testing.T) {
	manager := NewManager()

	proc, err := manager.NewProcess(
		ProcCommand("sleep"),
		ProcArgs("1"),
		ProcAutoReStart(AutoReStartTrue),
	)
	if err != nil {
		t.Fatalf("Failed to create process: %v", err)
	}

	proc.Start(true)
	time.Sleep(2 * time.Second)

	if !proc.isRunning() {
		t.Error("Process should have restarted")
	}

	proc.Stop(true)
}
