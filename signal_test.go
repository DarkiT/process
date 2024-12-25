package process

import (
	"os"
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

	// 测试信号转换功能
	t.Run("Signal Conversion", func(t *testing.T) {
		// 测试字符串到信号的转换
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
			sig := signals.ToSignal(tc.signalStr)
			if sig != tc.expected {
				t.Errorf("信号转换错误: 期望 %v, 得到 %v", tc.expected, sig)
			}
		}
	})

	// 清理
	manager.StopAllProcesses()
}

// TestInvalidSignalHandling 测试无效信号处理
func TestInvalidSignalHandling(t *testing.T) {
	proc := NewProcess(
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

// TestSignalWithChildren 测试子进程信号处理
func TestSignalWithChildren(t *testing.T) {
	manager := NewManager()
	proc, err := manager.NewProcess(
		WithName("test-children"),
		WithCommand("bash"),
		WithArgs("-c", "sleep 30 & sleep 30 & sleep 30"),
		WithAutoReStart(AutoReStartFalse),
	)
	if err != nil {
		t.Fatalf("创建进程失败: %v", err)
	}

	proc.Start(true)
	time.Sleep(1 * time.Second)

	// 测试仅向主进程发送信号
	t.Run("Signal Main Process Only", func(t *testing.T) {
		err := proc.Signal(syscall.SIGINT, false)
		if err != nil {
			t.Errorf("发送信号到主进程失败: %v", err)
		}
	})

	// 测试向所有子进程发送信号
	t.Run("Signal All Children", func(t *testing.T) {
		err := proc.Signal(syscall.SIGTERM, true)
		if err != nil {
			t.Errorf("发送信号到子进程失败: %v", err)
		}
		time.Sleep(1 * time.Second)
		if proc.isRunning() {
			t.Error("所有进程应该在收到 SIGTERM 信号后停止")
		}
	})
}
