//go:build linux

package process

import (
	"os"
	"syscall"
	"testing"
	"time"
)

// TestLinuxSpecificSignals 测试 Linux 特定的信号
func TestLinuxSpecificSignals(t *testing.T) {
	manager := NewManager()
	proc, err := manager.NewProcess(
		WithName("test-linux-signals"),
		WithCommand("sleep"),
		WithArgs([]string{"30"}),
	)
	if err != nil {
		t.Fatalf("创建进程失败: %v", err)
	}

	proc.Start(true)
	time.Sleep(1 * time.Second)

	// 测试 Linux 特定信号
	signals := []os.Signal{
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	}

	for _, sig := range signals {
		t.Run(sig.String(), func(t *testing.T) {
			err := proc.Signal(sig, false)
			if err != nil {
				t.Errorf("发送信号 %v 失败: %v", sig, err)
			}
		})
	}

	// 清理
	proc.Signal(syscall.SIGTERM, true)
}
