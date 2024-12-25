package process

import (
	"testing"
	"time"
)

// TestNewManager 测试创建新的进程管理器
func TestNewManager(t *testing.T) {
	// 创建一个新的管理器实例
	manager := NewManager()
	if manager == nil {
		t.Error("NewManager() 返回了空值")
	}
}

// TestNewProcess 测试创建新进程
func TestNewProcess(t *testing.T) {
	manager := NewManager()

	// 测试场景1：创建一个有效的进程
	proc1, err := manager.NewProcess(
		WithName("test-proc"),
		WithCommand("echo"),
		WithArgs([]string{"hello"}),
	)
	if err != nil {
		t.Errorf("创建进程失败: %v", err)
	}
	if proc1.GetName() != "test-proc" {
		t.Errorf("进程名称错误，期望 'test-proc'，得到 '%s'", proc1.GetName())
	}

	// 测试场景2：创建同名进程（应该返回错误）
	_, err = manager.NewProcess(WithName("test-proc"))
	if err == nil {
		t.Error("创建同名进程应该返回错误")
	}
}

// TestManagerProcessOperations 测试进程管理相关操作
func TestManagerProcessOperations(t *testing.T) {
	manager := NewManager()

	// 创建测试进程
	_, _ = manager.NewProcess(WithName("test-proc"))

	// 测试查找进程
	found := manager.Find("test-proc")
	if found == nil {
		t.Error("Find() 未能找到已创建的进程")
	}

	// 测试移除进程
	removed := manager.Remove("test-proc")
	if removed == nil {
		t.Error("Remove() 未能移除进程")
	}

	// 确认进程已被移除
	notFound := manager.Find("test-proc")
	if notFound != nil {
		t.Error("进程移除后仍能被找到")
	}
}

// TestStartStopProcess 测试启动和停止进程
func TestStartStopProcess(t *testing.T) {
	manager := NewManager()

	// 创建一个快速执行完的测试进程（使用 "echo" 命令替代 "sleep"）
	_, err := manager.NewProcess(
		WithName("test-proc"),
		WithCommand("echo"),
		WithArgs([]string{"hello"}),
	)
	if err != nil {
		t.Errorf("创建进程失败: %v", err)
	}

	// 设置测试超时
	done := make(chan bool)
	go func() {
		// 测试启动进程
		success, err := manager.StartProcess("test-proc", true) // 使用 wait=true 等待进程完成
		if !success || err != nil {
			t.Errorf("启动进程失败: %v", err)
		}
		done <- true
	}()

	// 添加超时控制
	select {
	case <-done:
		// 测试成功完成
	case <-time.After(5 * time.Second):
		t.Fatal("测试超时")
	}
}

// TestGetProcessInfo 测试获取进程信息
func TestGetProcessInfo(t *testing.T) {
	manager := NewManager()

	// 创建测试进程
	manager.NewProcess(
		WithName("info-test"),
		WithCommand("echo"),
	)

	// 测试获取单个进程信息
	info, err := manager.GetProcessInfo("info-test")
	if err != nil {
		t.Errorf("获取进程信息失败: %v", err)
	}
	if info.Name != "info-test" {
		t.Errorf("进程信息名称错误，期望 'info-test'，得到 '%s'", info.Name)
	}

	// 测试获取所有进程信息
	allInfos, err := manager.GetAllProcessInfo()
	if err != nil {
		t.Errorf("获取所有进程信息失败: %v", err)
	}
	if len(allInfos) != 1 {
		t.Errorf("进程数量错误，期望 1，得到 %d", len(allInfos))
	}
}

// TestGracefulReload 测试平滑重启功能
func TestGracefulReload(t *testing.T) {
	manager := NewManager()

	// 创建一个快速执行完的测试进程
	_, err := manager.NewProcess(
		WithName("reload-test"),
		WithCommand("echo"),
		WithArgs([]string{"hello"}),
	)
	if err != nil {
		t.Errorf("创建进程失败: %v", err)
	}

	// 设置测试超时
	done := make(chan bool)
	go func() {
		// 测试平滑重启
		success, err := manager.GracefulReload("reload-test", true)
		if !success || err != nil {
			t.Errorf("平滑重启失败: %v", err)
		}
		done <- true
	}()

	// 添加超时控制
	select {
	case <-done:
		// 测试成功完成
	case <-time.After(5 * time.Second):
		t.Fatal("测试超时")
	}

	// 验证进程是否仍然存在
	proc := manager.Find("reload-test")
	if proc == nil {
		t.Error("重启后进程不存在")
	}
}

// TestClear 测试清除所有进程
func TestClear(t *testing.T) {
	manager := NewManager()

	// 创建多个测试进程
	manager.NewProcess(WithName("proc1"))
	manager.NewProcess(WithName("proc2"))
	manager.NewProcess(WithName("proc3"))

	// 清除所有进程
	manager.Clear()

	// 验证是否所有进程都被清除
	info, _ := manager.GetAllProcessInfo()
	if len(info) != 0 {
		t.Errorf("Clear() 后仍有 %d 个进程存在", len(info))
	}
}
