package process

import (
	"testing"
	"time"
)

func TestManagerOperations(t *testing.T) {
	manager := NewManager()

	// 测试创建进程
	proc1, err := manager.NewProcess(
		WithName("test1"),
		WithCommand("sleep"),
		WithArgs("1"),
		WithStdoutLog("/dev/null", "1MB"),
		WithStderrLog("/dev/null", "1MB"),
	)
	if err != nil {
		t.Fatalf("Failed to create process: %v", err)
	}

	// 测试重复创建同名进程
	_, err = manager.NewProcess(WithName("test1"))
	if err == nil {
		t.Error("Expected error when creating duplicate process")
	}

	// 测试通过Options创建进程
	opts := NewOptions(
		WithName("test2"),
		WithCommand("sleep"),
		WithArgs("1"),
		WithStdoutLog("/dev/null", "1MB"),
		WithStderrLog("/dev/null", "1MB"),
	)

	proc2, err := manager.NewProcessByOptions(opts)
	if err != nil {
		t.Fatalf("Failed to create process by options: %v", err)
	}

	// 测试进程查找
	found := manager.Find("test1")
	if found != proc1 {
		t.Error("Failed to find process test1")
	}

	// 测试进程列表
	infos, err := manager.GetAllProcessInfo()
	if err != nil {
		t.Fatalf("Failed to get process info: %v", err)
	}

	if len(infos) != 2 {
		t.Errorf("Expected 2 processes, got %d", len(infos))
	}

	// 测试批量操作
	count := 0
	manager.ForEachProcess(func(p *Process) {
		count++
	})

	if count != 2 {
		t.Errorf("Expected ForEachProcess to visit 2 processes, visited %d", count)
	}

	// 测试启动进程
	success, err := manager.StartProcess("test1", true)
	if !success || err != nil {
		t.Errorf("Failed to start process: %v", err)
	}

	// 测试停止进程
	success, err = manager.StopProcess("test1", true)
	if !success || err != nil {
		t.Errorf("Failed to stop process: %v", err)
	}

	// 测试平滑重启
	proc1.Start(true)
	time.Sleep(time.Second)

	success, err = manager.GracefulReload("test1", true)
	if !success || err != nil {
		t.Errorf("Failed to gracefully reload process: %v", err)
	}

	// 测试停止所有进程
	proc1.Start(true)
	proc2.Start(true)
	time.Sleep(time.Second)

	manager.StopAllProcesses()
	time.Sleep(time.Second)

	allStopped := true
	manager.ForEachProcess(func(p *Process) {
		if p.isRunning() {
			allStopped = false
		}
	})

	if !allStopped {
		t.Error("Not all processes were stopped")
	}

	// 测试清理
	manager.Clear()
	if manager.Find("test1") != nil || manager.Find("test2") != nil {
		t.Error("Processes were not cleared")
	}
}

func TestManagerErrorCases(t *testing.T) {
	manager := NewManager()

	// 测试启动不存在的进程
	_, err := manager.StartProcess("nonexistent", true)
	if err == nil {
		t.Error("Expected error when starting nonexistent process")
	}

	// 测试停止不存在的进程
	_, err = manager.StopProcess("nonexistent", true)
	if err == nil {
		t.Error("Expected error when stopping nonexistent process")
	}

	// 测试获取不存在进程的信息
	_, err = manager.GetProcessInfo("nonexistent")
	if err == nil {
		t.Error("Expected error when getting info for nonexistent process")
	}

	// 测试平滑重启不存在的进程
	_, err = manager.GracefulReload("nonexistent", true)
	if err == nil {
		t.Error("Expected error when reloading nonexistent process")
	}
}
