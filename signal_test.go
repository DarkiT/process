package process

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/darkit/process/signals"
)

// TestInvalidSignalHandling 测试无效信号处理
func TestInvalidSignalHandling(t *testing.T) {
	proc := NewProcess(
		WithName("test-invalid-signal"),
		WithCommand("sleep"),
		WithArgs("1"),
		WithAutoReStart(AutoReStartFalse),
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
		WithAutoReStart(AutoReStartFalse),
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
		WithAutoReStart(AutoReStartFalse),
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
		WithAutoReStart(AutoReStartFalse),
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
