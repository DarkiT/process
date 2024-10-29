package process

import (
	"fmt"
	"github.com/darkit/process/utils"
	"syscall"
	"time"
)

// Info 进程的运行状态
type Info struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Start         int    `json:"start"`
	Stop          int    `json:"stop"`
	Now           int    `json:"now"`
	State         int    `json:"state"`
	StateName     string `json:"statename"`
	SpawnErr      string `json:"spawnerr"`
	ExitStatus    int    `json:"exitstatus"`
	Logfile       string `json:"logfile"`
	StdoutLogfile string `json:"stdout_logfile"`
	StderrLogfile string `json:"stderr_logfile"`
	Pid           int    `json:"pid"`
}

// GetProcessInfo 获取进程的详情
func (p *Process) GetProcessInfo() *Info {
	return &Info{
		Name:          p.GetName(),
		Description:   p.GetDescription(),
		Start:         int(p.GetStartTime().Unix()),
		Stop:          int(p.GetStopTime().Unix()),
		Now:           int(time.Now().Unix()),
		State:         int(p.GetState()),
		StateName:     p.GetState().String(),
		SpawnErr:      "",
		ExitStatus:    p.GetExitStatus(),
		Logfile:       p.GetStdoutLogfile(),
		StdoutLogfile: p.GetStdoutLogfile(),
		StderrLogfile: p.GetStderrLogfile(),
		Pid:           p.Pid(),
	}
}

// 获取进程的退出code值
func (p *Process) getExitCode() (int, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.cmd.ProcessState == nil {
		return -1, fmt.Errorf("no exit code")
	}

	if status, ok := p.cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
		return status.ExitStatus(), nil
	}

	return -1, fmt.Errorf("no exit code")
}

// 进程的退出code值是否在设置中的codes列表中
func (p *Process) inExitCodes(exitCode int) bool {
	for _, code := range p.getExitCodes() {
		if code == exitCode {
			return true
		}
	}
	return false
}

// 获取配置的退出code值列表
func (p *Process) getExitCodes() []int {
	strExitCodes := p.option.ExitCodes
	if len(strExitCodes) > 0 {
		return strExitCodes
	}
	return []int{0, 2} // 默认的退出码
}

// GetName 获取进程名
func (p *Process) GetName() string {
	return p.option.Name
}

// GetDescription 获取进程描述
func (p *Process) GetDescription() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.state == Running {
		seconds := int(time.Now().Sub(p.startTime).Seconds())
		minutes := seconds / 60
		hours := minutes / 60
		days := hours / 24

		if days > 0 {
			return fmt.Sprintf("pid %d, uptime %d days, %d:%02d:%02d",
				p.cmd.Process.Pid, days, hours%24, minutes%60, seconds%60)
		}
		return fmt.Sprintf("pid %d, uptime %d:%02d:%02d", p.cmd.Process.Pid, hours%24, minutes%60, seconds%60)
	} else if p.state != Stopped {
		t := p.stopTime.Format("2006-01-02 15:04:05")
		if p.stopTime.IsZero() {
			t = "未知"
		}
		return fmt.Sprintf("进程[%s]状态: %s 前一次停止时间: %s", p.GetName(), p.state.String(), t)
	}
	return ""
}

// GetStartTime 获取进程启动时间
func (p *Process) GetStartTime() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.startTime
}

// GetStopTime 获取进程结束时间
func (p *Process) GetStopTime() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()

	switch p.state {
	case Starting, Running, Stopping:
		return time.Unix(0, 0)
	default:
		return p.stopTime
	}
}

// GetState 获取进程状态
func (p *Process) GetState() State {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// GetExitStatus 获取进程退出状态
func (p *Process) GetExitStatus() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.state == Exited || p.state == Backoff {
		if p.cmd.ProcessState == nil {
			return 0
		}
		status, ok := p.cmd.ProcessState.Sys().(syscall.WaitStatus)
		if ok {
			return status.ExitStatus()
		}
	}
	return 0
}

// GetStdoutLogfile 获取标准输出将要写入的日志文件
func (p *Process) GetStdoutLogfile() string {
	fileName := "/dev/null"
	if len(p.option.StdoutLogfile) > 0 {
		fileName = p.option.StdoutLogfile
	}
	return utils.RealPath(fileName)
}

// GetStderrLogfile 获取标准错误将要写入的日志文件
func (p *Process) GetStderrLogfile() string {
	fileName := "/dev/null"
	if len(p.option.StderrLogfile) > 0 {
		fileName = p.option.StderrLogfile
	}
	return utils.RealPath(fileName)
}

// Pid 获取进程pid，返回0表示进程未启动
func (p *Process) Pid() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.state == Stopped || p.state == Fatal || p.state == Unknown ||
		p.state == Exited || p.state == Backoff {
		return 0
	}

	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Pid
	}
	return 0
}
