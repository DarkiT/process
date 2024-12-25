package process

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/darkit/process/proclog"
	"github.com/darkit/process/signals"
	"github.com/darkit/process/utils"
	"github.com/darkit/slog"
)

type Process struct {
	Manager *Manager
	option  ProcOptions
	cmd     *exec.Cmd

	startTime  time.Time
	stopTime   time.Time
	state      State
	retryTimes *int32

	stopCtx    context.Context
	stopCancel context.CancelFunc

	stdin     io.WriteCloser
	stdoutLog proclog.Logger
	stderrLog proclog.Logger

	mu         sync.RWMutex
	inStart    bool
	stopByUser bool
}

// NewProcess 创建进程对象
func NewProcess(opts ...ProcOption) *Process {
	options := NewProcOptions()
	options.Environment.Sets(utils.Map())
	dir, _ := os.Getwd()
	options.Directory = dir
	for _, opt := range opts {
		opt(&options)
	}
	return NewProcessByOptions(options)
}

// NewProcessByOptions 通过详细配置，创建进程对象
func NewProcessByOptions(options ProcOptions) *Process {
	var t time.Time
	proc := &Process{
		Manager:    nil,
		option:     options,
		startTime:  t,
		stopTime:   t,
		state:      Stopped,
		inStart:    false,
		stopByUser: false,
		retryTimes: new(int32),
	}
	return proc
}

// NewProcessCmd 按命令启动
func NewProcessCmd(cmd string, environment map[string]string) *Process {
	return NewProcess(
		WithCommand(getShell()),
		WithArgs(append([]string{getShellOption()}, parseCommand(cmd)...)),
		WithEnvironment(environment),
	)
}

// Start 启动进程，wait表示阻塞等待进程启动成功
func (p *Process) Start(wait bool) {
	slog.Infof("尝试启动程序[%s]", p.option.Name)

	p.mu.Lock()
	if p.inStart {
		slog.Infof("不用重复启动该进程[%s],因为该进程已经启动！", p.option.Name)
		p.mu.Unlock()
		return
	}
	p.inStart = true
	p.stopByUser = false
	p.mu.Unlock()

	var runCond *sync.Cond
	if wait {
		runCond = sync.NewCond(&sync.Mutex{})
		runCond.L.Lock()
	}

	go func() {
		for {
			p.run(func() {
				if wait {
					runCond.L.Lock()
					runCond.Signal()
					runCond.L.Unlock()
				}
			})

			if time.Now().Unix()-p.startTime.Unix() < 2 {
				time.Sleep(3 * time.Second)
			}

			p.mu.RLock()
			stopByUser := p.stopByUser
			p.mu.RUnlock()

			if stopByUser {
				slog.Infof("用户主动结束了该程序[%s]，不要再次启动", p.option.Name)
				break
			}

			if !p.isAutoRestart() {
				slog.Infof("不用自动重启进程[%s],因为该进程设置了不需要自动重启", p.option.Name)
				break
			}
			slog.Infof("因为该进程设置了自动重启,自动重启进程[%s]", p.option.Name)
		}
		p.mu.Lock()
		p.inStart = false
		p.mu.Unlock()
	}()

	if wait {
		runCond.Wait()
		runCond.L.Unlock()
	}
}

// Stop 主动停止进程
func (p *Process) Stop(wait bool) {
	p.mu.Lock()
	p.stopByUser = true
	isRunning := p.isRunning()
	p.mu.Unlock()

	if !isRunning {
		slog.Infof("程序[%s]未运行", p.GetName())
		return
	}
	slog.Infof("正在停止程序[%s]", p.GetName())

	sigs := p.option.StopSignal
	waitSecond := time.Duration(p.option.StopWaitSecs) * time.Second
	killWaitSecond := time.Duration(p.option.KillWaitSecs) * time.Second
	stopAsGroup := p.option.StopAsGroup
	killAsGroup := p.option.KillAsGroup

	if stopAsGroup && !killAsGroup {
		slog.Error("不能够同时设置 stopAsGroup=true 和 killAsGroup=false")
	}

	var stopped int32 = 0

	go func() {
		for i := 0; i < len(sigs) && atomic.LoadInt32(&stopped) == 0; i++ {
			sig := signals.ToSignal(sigs[i])
			slog.Infof("发送结束进程信号[%s]给进程[%s]", p.GetName(), sigs[i])
			_ = p.Signal(sig, stopAsGroup)

			endTime := time.Now().Add(waitSecond)
			for endTime.After(time.Now()) {
				p.mu.RLock()
				state := p.state
				p.mu.RUnlock()

				if state != Starting && state != Running && state != Stopping {
					atomic.StoreInt32(&stopped, 1)
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
		}

		if atomic.LoadInt32(&stopped) == 0 {
			slog.Infof("强制结束程序[%s]", p.GetName())
			_ = p.Signal(syscall.SIGKILL, killAsGroup)

			killEndTime := time.Now().Add(killWaitSecond)
			for killEndTime.After(time.Now()) {
				p.mu.RLock()
				state := p.state
				p.mu.RUnlock()

				if state != Starting && state != Running && state != Stopping {
					atomic.StoreInt32(&stopped, 1)
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			atomic.StoreInt32(&stopped, 1)
		}
	}()

	if wait {
		for atomic.LoadInt32(&stopped) == 0 {
			time.Sleep(1 * time.Second)
		}
	}
}

// 启动进程
func (p *Process) run(finishCb func()) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning() {
		slog.Infof("不能启动进程[%s],因为它正在运行中...", p.option.Name)
		finishCb()
		return
	}

	p.startTime = time.Now()
	atomic.StoreInt32(p.retryTimes, 0)
	startSecs := p.option.StartSecs
	restartPause := p.option.RestartPause

	var once sync.Once
	finishCbWrapper := func() {
		once.Do(finishCb)
	}

	for !p.stopByUser {
		if restartPause > 0 && atomic.LoadInt32(p.retryTimes) != 0 {
			p.mu.Lock()
			slog.Infof("不能立刻重启程序[%s],需要等待%d秒", p.option.Name, restartPause)
			time.Sleep(time.Duration(restartPause) * time.Second)
			p.mu.Unlock()
		}

		endTime := time.Now().Add(time.Duration(startSecs) * time.Second)
		p.changeStateTo(Starting)
		atomic.AddInt32(p.retryTimes, 1)

		err := p.createProgramCommand()
		if err != nil {
			p.failToStartProgram(fmt.Sprintf("不能创建进程,err:%v", err), finishCbWrapper)
			break
		}

		err = p.cmd.Start()
		if err != nil {
			if atomic.LoadInt32(p.retryTimes) >= int32(p.option.StartRetries) {
				p.failToStartProgram(fmt.Sprintf("error:%v", err), finishCbWrapper)
				break
			} else {
				slog.Infof("程序[%s]启动失败,再次重试,error:%v", p.option.Name, err)
				p.changeStateTo(Backoff)
				continue
			}
		}

		if p.stdoutLog != nil {
			p.stdoutLog.SetPid(p.cmd.Process.Pid)
		}
		if p.stderrLog != nil {
			p.stderrLog.SetPid(p.cmd.Process.Pid)
		}

		monitorExited := int32(0)
		programExited := int32(0)

		if startSecs <= 0 {
			slog.Infof("程序[%s]启动成功", p.option.Name)
			p.changeStateTo(Running)
			go finishCbWrapper()
		} else {
			go func() {
				p.monitorProgramIsRunning(endTime, &monitorExited, &programExited)
				finishCbWrapper()
			}()
		}

		if p.stopTime.IsZero() {
			slog.Debugf("正在尝试启动[%s]请稍后...", p.option.Name)
		} else {
			slog.Debugf("进程正在运行[%s]等待退出", p.option.Name)
		}

		p.mu.Unlock()
		p.waitForExit(int64(startSecs))
		atomic.StoreInt32(&programExited, 1)

		for atomic.LoadInt32(&monitorExited) == 0 {
			time.Sleep(time.Duration(10) * time.Millisecond)
		}
		p.mu.Lock()

		if p.state == Running {
			p.changeStateTo(Exited)
			slog.Infof("程序[%s]已经结束", p.option.Name)
			break
		} else {
			p.changeStateTo(Backoff)
		}

		if atomic.LoadInt32(p.retryTimes) >= int32(p.option.StartRetries) {
			p.failToStartProgram(fmt.Sprintf("不能启动程序[%s],因为已经超出了它的最大重试值:%d", p.option.Name, p.option.StartRetries), finishCbWrapper)
			break
		}
	}
}

// 创建程序的cmd对象
func (p *Process) createProgramCommand() error {
	var err error
	p.cmd, err = p.option.CreateCommand()
	if err != nil {
		return err
	}

	if err = p.setUser(); err != nil {
		return fmt.Errorf("设置程序运行时用户[%s]失败", p.option.User)
	}

	p.sysProcAttrSetPGid(p.cmd.SysProcAttr)

	if len(p.option.ExtraFiles) > 0 {
		p.cmd.ExtraFiles = p.option.ExtraFiles
	}

	p.setEnv()
	p.setDir()
	p.setLog()

	p.stdin, _ = p.cmd.StdinPipe()
	return nil
}

// 判断进程是否在运行
func (p *Process) isRunning() bool {
	if p.cmd != nil && p.cmd.Process != nil {
		if runtime.GOOS == "windows" {
			proc, err := os.FindProcess(p.cmd.Process.Pid)
			return proc != nil && err == nil
		}
		return p.cmd.Process.Signal(syscall.Signal(0)) == nil
	}
	return false
}

// 在supervisord启动的时候也自动启动
func (p *Process) isAutoStart() bool {
	return p.option.AutoStart
}

// 设置进程运行的环境变量
func (p *Process) setEnv() {
	if p.option.Environment.Size() > 0 {
		_ = utils.SetMap(p.option.Environment.Map())
	}
	p.cmd.Env = utils.All()
}

// 设置进程的运行目录
func (p *Process) setDir() {
	dir := p.option.Directory
	if dir != "" {
		p.cmd.Dir = dir
	}
}

// 设置进程的运行日志存放文件
func (p *Process) setLog() {
	p.stdoutLog = p.createStdoutLogger()
	p.cmd.Stdout = p.stdoutLog

	if p.option.RedirectStderr {
		p.stderrLog = p.stdoutLog
	} else {
		p.stderrLog = p.createStderrLogger()
	}
	p.cmd.Stderr = p.stderrLog
}

// 设置程序启动失败状态
func (p *Process) failToStartProgram(reason string, finishCb func()) {
	slog.Error("程序[%s]启动失败，失败原因：%s ", p.option.Name, reason)
	p.changeStateTo(Fatal)
	finishCb()
}

// 监控进程是否正在运行中
func (p *Process) monitorProgramIsRunning(endTime time.Time, monitorExited *int32, programExited *int32) {
	for time.Now().Before(endTime) && atomic.LoadInt32(programExited) == 0 {
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
	atomic.StoreInt32(monitorExited, 1)

	p.mu.Lock()
	defer p.mu.Unlock()

	if atomic.LoadInt32(programExited) == 0 && p.state == Starting {
		slog.Infof("进程[%s]启动成功", p.option.Name)
		p.changeStateTo(Running)
	}
}

// 判断进程是否需要自动重启
func (p *Process) isAutoRestart() bool {
	autoRestart := p.option.AutoReStart

	if autoRestart == AutoReStartFalse {
		return false
	} else if autoRestart == AutoReStartTrue {
		return true
	} else {
		p.mu.RLock()
		defer p.mu.RUnlock()

		if p.cmd != nil && p.cmd.ProcessState != nil {
			exitCode, err := p.getExitCode()
			return err == nil && !p.inExitCodes(exitCode)
		}
	}
	return false
}

// 阻塞等待进程运行结束
func (p *Process) waitForExit(startSecs int64) {
	_ = p.cmd.Wait()

	if p.cmd.ProcessState != nil {
		slog.Infof("程序[%s]已经运行结束，退出码为:%v", p.option.Name, p.cmd.ProcessState)
	} else {
		slog.Infof("程序[%s]已经运行结束", p.option.Name)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.stopTime = time.Now()

	if p.stdoutLog != nil {
		_ = p.stdoutLog.Close()
	}
	if p.stderrLog != nil {
		_ = p.stderrLog.Close()
	}
}

// Clone 进程
func (p *Process) Clone() (*Process, error) {
	var t time.Time
	proc := &Process{
		Manager:    p.Manager,
		option:     p.option,
		startTime:  t,
		stopTime:   t,
		state:      Stopped,
		inStart:    false,
		stopByUser: false,
		retryTimes: new(int32),
	}

	err := proc.createProgramCommand()
	if err != nil {
		return nil, err
	}
	return proc, nil
}

// 更改进程的运行状态
func (p *Process) changeStateTo(procState State) {
	p.state = procState
}
