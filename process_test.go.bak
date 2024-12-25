package process

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// 创建测试用的临时脚本文件
func createTestScript(t *testing.T, content string) string {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "test.sh")

	err := os.WriteFile(scriptPath, []byte(content), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	return scriptPath
}

func TestProcessBasicOperations(t *testing.T) {
	// 创建一个简单的测试脚本
	scriptContent := `#!/bin/sh
count=0
while [ $count -lt 5 ]; do
    echo "Running $count"
    sleep 1
    count=$((count + 1))
done`

	scriptPath := createTestScript(t, scriptContent)

	// 创建进程管理器
	manager := NewManager()

	// 创建进程
	proc, err := manager.NewProcess(
		WithCommand(scriptPath),
		WithName("test-process"),
		WithAutoStart(false),
	)
	if err != nil {
		t.Fatalf("Failed to create process: %v", err)
	}

	// 测试进程启动
	proc.Start(true)
	time.Sleep(time.Second) // 给进程一些启动时间

	if proc.GetState() != Running {
		t.Errorf("Expected process state to be Running, got %v", proc.GetState())
	}

	// 测试获取进程信息
	info := proc.GetProcessInfo()
	if info.Name != "test-process" {
		t.Errorf("Expected process name to be test-process, got %s", info.Name)
	}

	if info.Pid == 0 {
		t.Error("Expected non-zero PID")
	}

	// 测试停止进程
	proc.Stop(true)
	time.Sleep(time.Second) // 给进程一些停止时间

	if proc.GetState() != Exited && proc.GetState() != Stopped {
		t.Errorf("Expected process state to be Exited or Stopped, got %v", proc.GetState())
	}
}

func TestProcessAutoRestart(t *testing.T) {
	scriptContent := `#!/bin/sh
echo "Starting"
sleep 1
exit 1` // 让进程快速退出

	scriptPath := createTestScript(t, scriptContent)

	manager := NewManager()

	proc, err := manager.NewProcess(
		WithCommand(scriptPath),
		WithName("test-restart"),
		WithAutoReStart(AutoReStartTrue),
		WithStartRetries(2),
	)
	if err != nil {
		t.Fatalf("Failed to create process: %v", err)
	}

	proc.Start(true)
	time.Sleep(3 * time.Second) // 给足够时间让进程重启

	// 检查重启次数
	if *proc.retryTimes < 1 {
		t.Error("Process should have attempted to restart")
	}
}

func TestProcessManager(t *testing.T) {
	manager := NewManager()

	// 测试创建多个进程
	proc1, _ := manager.NewProcess(WithName("test1"), WithCommand("sleep"), WithArgs("1"))
	_, _ = manager.NewProcess(WithName("test2"), WithCommand("sleep"), WithArgs("1"))

	// 测试查找进程
	found := manager.Find("test1")
	if found != proc1 {
		t.Error("Failed to find process test1")
	}

	// 测试获取所有进程信息
	infos, err := manager.GetAllProcessInfo()
	if err != nil {
		t.Errorf("Failed to get all process info: %v", err)
	}

	if len(infos) != 2 {
		t.Errorf("Expected 2 processes, got %d", len(infos))
	}

	// 测试移除进程
	removed := manager.Remove("test1")
	if removed != proc1 {
		t.Error("Failed to remove process test1")
	}

	if manager.Find("test1") != nil {
		t.Error("Process test1 should have been removed")
	}
}
