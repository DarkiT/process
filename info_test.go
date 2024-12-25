package process

import (
	"strings"
	"testing"
	"time"
)

func TestProcessInfo(t *testing.T) {
	manager := NewManager()
	proc, err := manager.NewProcess(
		WithName("test-info"),
		WithCommand("sleep"),
		WithArgs("1"),
	)
	if err != nil {
		t.Fatalf("Failed to create process: %v", err)
	}

	// 测试基本信息获取
	if name := proc.GetName(); name != "test-info" {
		t.Errorf("Expected name test-info, got %s", name)
	}

	// 测试进程信息
	info := proc.GetProcessInfo()
	if info.Name != "test-info" {
		t.Errorf("Expected info name test-info, got %s", info.Name)
	}

	// 测试未启动状态的信息
	if info.Pid != 0 {
		t.Error("Expected PID 0 for non-running process")
	}

	if info.State != int(Stopped) {
		t.Errorf("Expected initial state Stopped, got %d", info.State)
	}

	// 启动进程并测试运行状态信息
	proc.Start(true)
	time.Sleep(100 * time.Millisecond)

	runningInfo := proc.GetProcessInfo()
	if runningInfo.Pid == 0 {
		t.Error("Expected non-zero PID for running process")
	}

	if runningInfo.State != int(Running) {
		t.Errorf("Expected state Running, got %d", runningInfo.State)
	}

	// 测试描述信息
	desc := proc.GetDescription()
	if !strings.Contains(desc, "pid") || !strings.Contains(desc, "uptime") {
		t.Errorf("Expected description to contain pid and uptime, got %s", desc)
	}

	// 测试时间相关函数
	startTime := proc.GetStartTime()
	if startTime.IsZero() {
		t.Error("Expected non-zero start time")
	}

	stopTime := proc.GetStopTime()
	if !stopTime.IsZero() {
		t.Error("Expected zero stop time for running process")
	}

	// 停止进程并测试状态变化
	proc.Stop(true)
	time.Sleep(100 * time.Millisecond)

	stoppedInfo := proc.GetProcessInfo()
	if stoppedInfo.Pid != 0 {
		t.Error("Expected zero PID for stopped process")
	}

	finalStopTime := proc.GetStopTime()
	if finalStopTime.IsZero() {
		t.Error("Expected non-zero stop time after stopping")
	}
}
