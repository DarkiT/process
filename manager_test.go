package process

import (
	"testing"
	"time"
)

func TestManagerProcessManagement(t *testing.T) {
	manager := NewManager()

	// Test process creation
	proc, err := manager.NewProcessCmd("echo test", nil)
	if err != nil {
		t.Fatalf("Failed to create process: %v", err)
	}

	// Test process lookup
	if found := manager.Find(proc.GetName()); found == nil {
		t.Error("Process should be found in manager")
	}

	// Test process removal
	manager.Remove(proc.GetName())
	if found := manager.Find(proc.GetName()); found != nil {
		t.Error("Process should not be found after removal")
	}
}

func TestManagerStopAll(t *testing.T) {
	manager := NewManager()

	// Create multiple processes
	for i := 0; i < 3; i++ {
		_, err := manager.NewProcessCmd("sleep 5", nil)
		if err != nil {
			t.Fatalf("Failed to create process: %v", err)
		}
	}

	// Start all processes
	manager.ForEachProcess(func(p *Process) {
		p.Start(false)
	})

	time.Sleep(100 * time.Millisecond)

	// Stop all processes
	manager.StopAllProcesses()

	time.Sleep(100 * time.Millisecond)

	// Verify all processes are stopped
	manager.ForEachProcess(func(p *Process) {
		if p.isRunning() {
			t.Error("Process should be stopped")
		}
	})
}
