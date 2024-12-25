//go:build linux

package process

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

// TestReapZombie 测试僵尸进程回收功能
func TestReapZombie(t *testing.T) {
	// 启动僵尸进程回收器
	ReapZombie()
	time.Sleep(100 * time.Millisecond) // 等待回收器初始化

	// 创建一个会产生僵尸进程的测试
	t.Run("Reap Single Zombie", func(t *testing.T) {
		// 创建一个子进程，让它立即退出
		cmd := exec.Command("sh", "-c", "exit 0")
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true, // 设置进程组
		}

		err := cmd.Start()
		if err != nil {
			t.Fatalf("启动测试进程失败: %v", err)
		}

		// 等待一段时间让子进程变成僵尸进程
		time.Sleep(100 * time.Millisecond)
	})

	// 测试多个僵尸进程的回收
	t.Run("Reap Multiple Zombies", func(t *testing.T) {
		// 创建多个会立即退出的子进程
		for i := 0; i < 5; i++ {
			cmd := exec.Command("sh", "-c", "exit 0")
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Setpgid: true,
			}

			err := cmd.Start()
			if err != nil {
				t.Fatalf("启动测试进程 %d 失败: %v", i, err)
			}
		}

		// 等待一段时间让子进程变成僵尸进程并被回收
		time.Sleep(100 * time.Millisecond)
	})
}

// TestReapWithConfig 测试使用自定义配置的僵尸进程回收
func TestReapWithConfig(t *testing.T) {
	config := Config{
		Pid:              -1,   // 回收所有子进程
		Options:          0,    // 默认选项
		DisablePid1Check: true, // 禁用 PID 1 检查以便测试
	}

	// 启动带自定义配置的回收器
	Start(config)
	time.Sleep(100 * time.Millisecond)

	// 创建测试进程
	cmd := exec.Command("sh", "-c", "exit 0")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	err := cmd.Start()
	if err != nil {
		t.Fatalf("启动测试进程失败: %v", err)
	}

	// 等待进程退出和回收
	time.Sleep(100 * time.Millisecond)
}

// TestPid1Check 测试 PID 1 检查功能
func TestPid1Check(t *testing.T) {
	if os.Getpid() == 1 {
		t.Skip("此测试不能在 PID 1 进程中运行")
	}

	config := Config{
		Pid:              -1,
		Options:          0,
		DisablePid1Check: false, // 启用 PID 1 检查
	}

	// 启动回收器，应该因为不是 PID 1 而不会实际启动
	Start(config)
	time.Sleep(100 * time.Millisecond)

	// 创建测试进程
	cmd := exec.Command("sh", "-c", "exit 0")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	err := cmd.Start()
	if err != nil {
		t.Fatalf("启动测试进程失败: %v", err)
	}

	// 等待一段时间，进程应该变成僵尸进程（因为回收器未实际启动）
	time.Sleep(100 * time.Millisecond)
}

// TestSignalHandler 测试信号处理功能
func TestSignalHandler(t *testing.T) {
	notifications := make(chan os.Signal, 1)

	// 启动信号处理器
	go sigChildHandler(notifications)
	time.Sleep(100 * time.Millisecond)

	// 创建一个会立即退出的子进程
	cmd := exec.Command("sh", "-c", "exit 0")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	err := cmd.Start()
	if err != nil {
		t.Fatalf("启动测试进程失败: %v", err)
	}

	// 等待接收 SIGCHLD 信号
	select {
	case sig := <-notifications:
		if sig != syscall.SIGCHLD {
			t.Errorf("收到意外的信号: %v", sig)
		}
	case <-time.After(time.Second):
		t.Error("等待 SIGCHLD 信号超时")
	}
}
