package process

import (
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/darkit/process/signals"
)

// TestProcessSignalHandling 测试进程信号处理
func TestProcessSignalHandling(t *testing.T) {
	// 创建一个测试进程管理器
	manager := NewManager()

	// 创建一个长期运行的测试进程
	proc, err := manager.NewProcess(
		WithName("test-signal"),
		WithCommand("sleep"),
		WithArgs("30"),
	)
	if err != nil {
		t.Fatalf("创建进程失败: %v", err)
	}

	// 启动进程
	proc.Start(true)

	// 等待进程完全启动
	time.Sleep(1 * time.Second)

	// 测试发送 SIGTERM 信号
	t.Run("SIGTERM Signal", func(t *testing.T) {
		err := proc.Signal(syscall.SIGTERM, false)
		if err != nil {
			t.Errorf("发送 SIGTERM 信号失败: %v", err)
		}
		time.Sleep(1 * time.Second)
		if proc.isRunning() {
			t.Error("进程应该在收到 SIGTERM 信号后停止")
		}
	})

	// 重新启动进程用于后续测试
	proc.Start(true)
	time.Sleep(1 * time.Second)

	// 测试发送 SIGKILL 信号
	t.Run("SIGKILL Signal", func(t *testing.T) {
		err := proc.Signal(syscall.SIGKILL, false)
		if err != nil {
			t.Errorf("发送 SIGKILL 信号失败: %v", err)
		}
		time.Sleep(1 * time.Second)
		if proc.isRunning() {
			t.Error("进程应该在收到 SIGKILL 信号后停止")
		}
	})

	// 测试进程组信号处理
	t.Run("Process Group Signal", func(t *testing.T) {
		proc, err := manager.NewProcess(
			WithName("test-group-signal"),
			WithCommand("bash"),
			WithArgs("-c", "sleep 30 & sleep 30"),
		)
		if err != nil {
			t.Fatalf("创建进程组测试进程失败: %v", err)
		}

		proc.Start(true)
		time.Sleep(1 * time.Second)

		// 发送信号到进程组
		err = proc.Signal(syscall.SIGTERM, true)
		if err != nil {
			t.Errorf("发送进程组信号失败: %v", err)
		}

		time.Sleep(1 * time.Second)
		if proc.isRunning() {
			t.Error("进程组应该在收到信号后停止")
		}
	})

	// 测试信号转换
	t.Run("Signal Conversion", func(t *testing.T) {
		testCases := []struct {
			signalStr string
			expected  os.Signal
		}{
			{"SIGTERM", syscall.SIGTERM},
			{"SIGKILL", syscall.SIGKILL},
			{"SIGINT", syscall.SIGINT},
			{"SIGHUP", syscall.SIGHUP},
		}

		for _, tc := range testCases {
			t.Run(tc.signalStr, func(t *testing.T) {
				if runtime.GOOS == "windows" {
					return
				}
				sig := signals.ToSignal(tc.signalStr)
				if sig != tc.expected {
					t.Errorf("信号转换错误: 期望 %v, 得到 %v", tc.expected, sig)
				}
			})
		}
	})

	// 清理
	manager.StopAllProcesses()
}

// TestInvalidSignalHandling 测试无效信号处理
func TestInvalidSignalHandling(t *testing.T) {
	// 创建 manager
	manager := NewManager()
	proc, _ := manager.NewProcess(
		WithName("test-invalid-signal"),
		WithCommand("sleep"),
		WithArgs("1"),
	)

	// 测试向未启动的进程发送信号
	err := proc.Signal(syscall.SIGTERM, false)
	if err == nil {
		t.Error("向未启动的进程发送信号应该返回错误")
	}

	// 测试无效的信号值
	proc.Start(true)
	time.Sleep(1 * time.Second)

	invalidSignal := signals.ToSignal("INVALID_SIGNAL")
	err = proc.Signal(invalidSignal, false)
	if err == nil {
		t.Error("发送无效信号应该返回错误")
	}
}

// TestSendSignals 测试发送多个信号
func TestSendSignals(t *testing.T) {
	manager := NewManager()
	proc, err := manager.NewProcess(
		WithName("test-multiple-signals"),
		WithCommand("sleep"),
		WithArgs("30"),
	)
	if err != nil {
		t.Fatalf("创建进程失败: %v", err)
	}

	// 启动进程
	proc.Start(true)
	time.Sleep(1 * time.Second)

	// 测试发送多个信号
	signals := []string{"SIGUSR1", "SIGUSR2", "SIGTERM"}
	proc.sendSignals(signals, false)

	time.Sleep(1 * time.Second)
	if proc.isRunning() {
		t.Error("进程应该在收到 SIGTERM 信号后停止")
	}
}

// TestSignalWithLock 测试信号发送时的锁机制
func TestSignalWithLock(t *testing.T) {
	manager := NewManager()
	proc, err := manager.NewProcess(
		WithName("test-signal-lock"),
		WithCommand("sleep"),
		WithArgs("30"),
	)
	if err != nil {
		t.Fatalf("创建进程失败: %v", err)
	}

	proc.Start(true)
	time.Sleep(1 * time.Second)

	// 并发发送信号测试锁机制
	done := make(chan bool)
	go func() {
		err := proc.Signal(syscall.SIGINT, false)
		if err != nil {
			t.Errorf("发送 SIGINT 信号失败: %v", err)
		}
		done <- true
	}()

	err = proc.Signal(syscall.SIGTERM, false)
	if err != nil {
		t.Errorf("发送 SIGTERM 信号失败: %v", err)
	}

	<-done
	time.Sleep(1 * time.Second)
	if proc.isRunning() {
		t.Error("进程应该在收到信号后停止")
	}
}

// TestSignalToStoppedProcess 测试向已停止的进程发送信号
func TestSignalToStoppedProcess(t *testing.T) {
	manager := NewManager()
	proc, err := manager.NewProcess(
		WithName("test-stopped-process"),
		WithCommand("sleep"),
		WithArgs("1"),
	)
	if err != nil {
		t.Fatalf("创建进程失败: %v", err)
	}

	// 启动并等待进程自然结束
	proc.Start(true)
	time.Sleep(2 * time.Second)

	// 向已停止的进程发送信号
	err = proc.Signal(syscall.SIGTERM, false)
	if err == nil {
		t.Error("向已停止的进程发送信号应该返回错误")
	}
}

// TestSignalWithChildrenGroup 测试子进程信号处理
func TestSignalWithChildrenGroup(t *testing.T) {
	// 创建一个会启动子进程的测试脚本
	scriptContent := `#!/bin/sh
	child_pid=""
	trap 'kill $child_pid; exit' TERM
	
	# 启动后台子进程
	sleep 100 &
	child_pid=$!
	
	# 等待信号
	while true; do
		sleep 1
	done`

	scriptPath := createTestScript(t, scriptContent)

	manager := NewManager()
	proc, err := manager.NewProcess(
		WithCommand(scriptPath),
		WithName("test-signal-children"),
		WithStopAsGroup(true),
		WithKillAsGroup(true),
	)
	if err != nil {
		t.Fatalf("Failed to create process: %v", err)
	}

	// 启动进程
	proc.Start(true)
	time.Sleep(2 * time.Second) // 等待进程和子进程启动

	// 获取进程组ID
	pgid := proc.cmd.Process.Pid

	// 发送信号到进程组
	err = proc.Signal(syscall.SIGTERM, true)
	if err != nil {
		t.Errorf("Failed to send signal to process group: %v", err)
	}

	time.Sleep(time.Second)

	// 验证进程组中的进程都已终止
	if proc.isRunning() {
		t.Error("Main process should not be running")
	}

	// 在Linux/Unix系统上验证进程组
	if os.Getenv("GITHUB_ACTIONS") != "true" { // 跳过CI环境
		if _, err := os.FindProcess(pgid); err == nil {
			t.Error("Process group should be terminated")
		}
	}
}

// 可以将信号转换测试完全分离为独立的测试函数
func TestSignalConversion(t *testing.T) {
	testCases := []struct {
		signalStr string
		expected  os.Signal
	}{
		{"SIGTERM", syscall.SIGTERM},
		{"SIGKILL", syscall.SIGKILL},
		{"SIGINT", syscall.SIGINT},
		{"SIGHUP", syscall.SIGHUP},
	}

	for _, tc := range testCases {
		t.Run(tc.signalStr, func(t *testing.T) {
			sig := signals.ToSignal(tc.signalStr)
			if sig != tc.expected {
				t.Errorf("信号转换错误: 期望 %v, 得到 %v", tc.expected, sig)
			}
		})
	}
}

func TestProcessSignal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	// 创建一个简单的测试脚本,该脚本会捕获信号
	scriptContent := `#!/bin/sh
trap 'echo "Received signal"' TERM
while true; do
    sleep 1
done`

	scriptPath := createTestScript(t, scriptContent)
	defer os.Remove(scriptPath) // 清理测试脚本

	manager := NewManager()
	proc, err := manager.NewProcess(
		WithCommand(scriptPath),
		WithName("test-signal"),
	)
	if err != nil {
		t.Fatalf("Failed to create process: %v", err)
	}

	// 启动进程
	proc.Start(true)
	time.Sleep(time.Second) // 等待进程完全启动

	// 测试发送单个信号
	err = proc.Signal(syscall.SIGTERM, false)
	if err != nil {
		t.Errorf("Failed to send SIGTERM: %v", err)
	}

	// 测试发送信号到未运行的进程
	proc.Stop(true)
	err = proc.Signal(syscall.SIGTERM, false)
	if err == nil {
		t.Error("Expected error when sending signal to stopped process")
	}

	// 测试发送多个信号
	proc.Start(true)
	time.Sleep(time.Second)

	proc.sendSignals([]string{"TERM", "KILL"}, false)
	time.Sleep(time.Second)

	if proc.isRunning() {
		t.Error("Process should not be running after SIGKILL")
	}
}
