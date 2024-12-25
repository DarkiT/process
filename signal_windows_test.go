//go:build windows

package process

import (
	"os"
	"syscall"
	"testing"
	"time"
)

// TestLinuxOnlySignals 测试 Windows 特定的信号处理
func TestLinuxOnlySignals(t *testing.T) {
	manager := NewManager()
	proc, err := manager.NewProcess(
		WithName("test-windows-signals"),
		WithCommand("timeout"),
		WithArgs("/t", "30"),
	)
	if err != nil {
		t.Fatalf("创建进程失败: %v", err)
	}

	proc.Start(true)
	time.Sleep(1 * time.Second)

	// 测试 Windows 支持的信号
	signals := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
	}

	for _, sig := range signals {
		t.Run(sig.String(), func(t *testing.T) {
			err := proc.Signal(sig, false)
			if err != nil {
				t.Errorf("发送信号 %v 失败: %v", sig, err)
			}
			// 重启进程用于下一个信号测试
			if !proc.isRunning() {
				proc.Start(true)
				time.Sleep(1 * time.Second)
			}
		})
	}

	// 清理
	proc.Signal(syscall.SIGTERM, true)
}
