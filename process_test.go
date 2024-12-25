package process

import (
	"os"
	"testing"
	"time"
)

// TestProcess 测试创建新进程对象
func TestProcess(t *testing.T) {
	// 创建一个基本的进程对象
	proc := NewProcess()
	if proc == nil {
		t.Error("NewProcess() 返回了空对象")
	}

	// 检查默认值设置
	if proc.state != Stopped {
		t.Errorf("期望初始状态为 Stopped，实际为 %v", proc.state)
	}

	// 检查工作目录是否正确设置
	dir, _ := os.Getwd()
	if proc.option.Directory != dir {
		t.Errorf("期望工作目录为 %s，实际为 %s", dir, proc.option.Directory)
	}
}

// TestNewProcessWithOptions 测试使用选项创建进程
func TestNewProcessWithOptions(t *testing.T) {
	// 创建带有自定义选项的进程
	proc := NewProcess(
		WithName("test-process"),
		WithCommand("echo"),
		WithArgs("hello"),
	)

	if proc.option.Name != "test-process" {
		t.Errorf("期望进程名称为 test-process，实际为 %s", proc.option.Name)
	}

	if proc.option.Command != "echo" {
		t.Errorf("期望命令为 echo，实际为 %s", proc.option.Command)
	}
}

// TestProcessStart 测试进程启动功能
func TestProcessStart(t *testing.T) {
	// 创建一个简单的测试进程（使用 echo 命令）
	proc := NewProcess(
		WithName("echo-test"),
		WithCommand("echo"),
		WithArgs("hello"),
		WithAutoStart(true),
	)

	// 启动进程并等待
	proc.Start(true)

	// 检查进程状态
	if proc.state != Exited && proc.state != Running {
		t.Errorf("期望进程状态为 Exited 或 Running，实际为 %v", proc.state)
	}
}

// TestProcessStop 测试进程停止功能
func TestProcessStop(t *testing.T) {
	// 创建一个长期运行的测试进程
	proc := NewProcess(
		WithName("sleep-test"),
		WithCommand("sleep"),
		WithArgs("5"),
	)

	// 启动进程
	proc.Start(true)

	// 停止进程
	proc.Stop(true)

	// 检查进程状态
	if proc.state != Stopped && proc.state != Exited {
		t.Errorf("期望进程状态为 Stopped 或 Exited，实际为 %v", proc.state)
	}
}

// TestProcessAutoRestart 测试自动重启功能
func TestProcessAutoRestart(t *testing.T) {
	// 创建一个会快速退出的进程，并设置自动重启
	proc := NewProcess(
		WithName("restart-test"),
		WithCommand("echo"),
		WithArgs("test"),
		WithAutoReStart(AutoReStartTrue),
	)

	// 启动进程
	proc.Start(true)

	// 等待一小段时间，让自动重启有机会发生
	time.Sleep(time.Second)

	// 检查重试次数
	if *proc.retryTimes == 0 {
		t.Error("进程应该至少重试一次")
	}
}

// TestProcessWithEnvironment 测试环境变量设置
func TestProcessWithEnvironment(t *testing.T) {
	// 创建带有自定义环境变量的进程
	env := map[string]string{
		"TEST_VAR": "test_value",
	}

	proc := NewProcess(
		WithName("env-test"),
		WithEnvironment(env),
	)

	if proc.option.Environment.Get("TEST_VAR") != "test_value" {
		t.Error("环境变量没有正确设置")
	}
}

// TestProcessClone 测试进程克隆功能
func TestProcessClone(t *testing.T) {
	// 创建原始进程
	original := NewProcess(
		WithName("original"),
		WithCommand("echo"),
		WithArgs("test"),
	)

	// 克隆进程
	cloned, err := original.Clone()
	if err != nil {
		t.Errorf("克隆进程失败: %v", err)
	}

	// 检查克隆的进程是否保持了相同的配置
	if cloned.option.Name != original.option.Name {
		t.Error("克隆的进程名称与原始进程不匹配")
	}

	if cloned.option.Command != original.option.Command {
		t.Error("克隆的进程命令与原始进程不匹配")
	}
}
