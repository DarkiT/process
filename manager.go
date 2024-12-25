package process

import (
	"errors"
	"fmt"
	"sync"

	"github.com/darkit/slog"
)

type Manager struct {
	processes sync.Map
	wg        sync.WaitGroup
}

// NewManager 创建进程管理器
func NewManager() *Manager {
	return &Manager{}
}

// NewProcess 创建新的进程实例
// opts: 配置对象
func (m *Manager) NewProcess(opts ...WithOption) (*Process, error) {
	options := NewOptions(opts...)

	if len(options.Name) == 0 {
		options.Name = options.Command
	}

	if _, exists := m.processes.Load(options.Name); exists {
		return nil, fmt.Errorf("进程[%s]已存在", options.Name)
	}

	proc := &Process{
		Manager:    m,
		option:     options,
		state:      Stopped,
		retryTimes: new(int32),
	}

	m.processes.Store(options.Name, proc)
	slog.Infof("创建进程: %s", proc.GetName())

	return proc, nil
}

// NewProcessByOptions 创建进程
// opts: 配置对象
func (m *Manager) NewProcessByOptions(opts Options) (*Process, error) {
	if _, exists := m.processes.Load(opts.Name); exists {
		return nil, fmt.Errorf("进程[%s]已存在", opts.Name)
	}

	proc := NewProcessByOptions(opts)
	proc.Manager = m
	m.processes.Store(opts.Name, proc)

	return proc, nil
}

// NewProcessByProcess 创建进程
// proc: Process对象
func (m *Manager) NewProcessByProcess(proc *Process) (*Process, error) {
	if _, found := m.processes.Load(proc.GetName()); found {
		return nil, errors.New(fmt.Sprintf("进程[%s]已存在", proc.GetName()))
	}
	proc.Manager = m
	m.processes.Store(proc.GetName(), proc)
	slog.Infof("创建进程: %s", proc.GetName())
	return proc, nil
}

// NewProcessCmd 创建进程
// cmd: 可执行文件路径及参数
// environment: 环境变量
func (m *Manager) NewProcessCmd(cmd string, environment map[string]string) (*Process, error) {
	p := NewProcessCmd(cmd, environment)
	if _, exists := m.processes.Load(p.GetName()); exists {
		return nil, fmt.Errorf("进程[%s]已存在", p.GetName())
	}
	p.Manager = m
	m.processes.Store(p.GetName(), p)
	return p, nil
}

// Add 添加进程到Manager
func (m *Manager) Add(name string, proc *Process) {
	m.processes.Store(name, proc)
	slog.Infof("添加进程: %s", name)
}

// Remove 从Manager移除进程
func (m *Manager) Remove(name string) *Process {
	if value, ok := m.processes.LoadAndDelete(name); ok {
		slog.Infof("移除进程: %s", name)
		return value.(*Process)
	}
	return nil
}

// Clear 清除进程
func (m *Manager) Clear() {
	m.processes.Range(func(key, value interface{}) bool {
		m.processes.Delete(key)
		return true
	})
}

// ForEachProcess 迭代进程列表
func (m *Manager) ForEachProcess(procFunc func(p *Process)) {
	m.processes.Range(func(_, value interface{}) bool {
		procFunc(value.(*Process))
		return true
	})
}

// StopAllProcesses 关闭所有进程
func (m *Manager) StopAllProcesses() {
	var wg sync.WaitGroup
	m.processes.Range(func(_, value interface{}) bool {
		proc := value.(*Process)
		wg.Add(1)
		go func() {
			defer wg.Done()
			proc.Stop(true)
		}()
		return true
	})
	wg.Wait()
}

// Find 根据进程名查询进程
func (m *Manager) Find(name string) *Process {
	if value, ok := m.processes.Load(name); ok {
		return value.(*Process)
	}
	return nil
}

// GetAllProcessInfo 获取所有进程列表
func (m *Manager) GetAllProcessInfo() ([]*Info, error) {
	var infos []*Info
	m.processes.Range(func(_, value interface{}) bool {
		proc := value.(*Process)
		infos = append(infos, proc.GetProcessInfo())
		return true
	})
	return infos, nil
}

// GetProcessInfo 获取所有进程信息
func (m *Manager) GetProcessInfo(name string) (*Info, error) {
	proc := m.Find(name)
	if proc == nil {
		return nil, fmt.Errorf("没有找到进程[%s]", name)
	}
	return proc.GetProcessInfo(), nil
}

// StartProcess 启动指定进程
func (m *Manager) StartProcess(name string, wait bool) (bool, error) {
	slog.Infof("启动进程[%s]", name)
	proc := m.Find(name)
	if proc == nil {
		return false, fmt.Errorf("没有找到要启动的进程[%s]", name)
	}
	proc.Start(wait)
	return true, nil
}

// StopProcess 停止指定进程
func (m *Manager) StopProcess(name string, wait bool) (bool, error) {
	slog.Infof("结束进程[%s]", name)
	proc := m.Find(name)
	if proc == nil {
		return false, fmt.Errorf("没有找到要结束的进程[%s]", name)
	}
	proc.Stop(wait)
	return true, nil
}

// GracefulReload 停止指定进程
func (m *Manager) GracefulReload(name string, wait bool) (bool, error) {
	slog.Infof("平滑重启进程[%s]", name)
	proc := m.Find(name)
	if proc == nil {
		return false, fmt.Errorf("没有找到要重启的进程[%s]", name)
	}
	procClone, err := proc.Clone()
	if err != nil {
		return false, err
	}
	procClone.Start(wait)
	proc.Stop(wait)
	m.processes.Store(name, procClone)
	return true, nil
}
