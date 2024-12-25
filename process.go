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
	Manager *Manager  // 进程管理对象
	option  Options   // 进程配置
	cmd     *exec.Cmd // 进程对象

	startTime   time.Time // 启动时间
	stopTime    time.Time // 停止时间
	state       State     // 进程的当前状态
	inStart     bool      // 正在启动的时候，该值为true
	stopByUser  bool      // 用户主动关闭的时候，该值为true
	retryTimes  *int32    // 启动的次数
	lastModTime time.Time // 文件最后修改时间

	lock          sync.RWMutex
	stdin         io.WriteCloser
	stdoutLog     proclog.Logger
	stderrLog     proclog.Logger
	monitorCancel context.CancelFunc
}

// NewProcess 创建进程对象
func NewProcess(opts ...WithOption) *Process {
	options := NewOptions()
	options.Environment.Sets(utils.Map())
	dir, _ := os.Getwd()
	options.Directory = dir
	for _, opt := range opts {
		opt(&options)
	}
	return NewProcessByOptions(options)
}

// NewProcessByOptions 通过详细配置，创建进程对象
func NewProcessByOptions(options Options) *Process {
	var t time.Time
	proc := &Process{
		Manager:    nil,
		cmd:        nil,
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
		WithArgs(append([]string{getShellOption()}, parseCommand(cmd)...)...),
		WithEnvironment(environment),
	)
}

// Start 启动进程，wait表示阻塞等待进程启动成功
func (that *Process) Start(wait bool) {
	slog.Info("尝试启动程序[%s]", that.option.Name)

	that.lock.Lock()
	if that.inStart {
		slog.Info("不能重复启动该进程[%s],因为该进程已经启动！", that.option.Name)
		that.lock.Unlock()
		return
	}
	that.inStart = true
	that.stopByUser = false
	that.lock.Unlock()

	var runCond *sync.Cond
	if wait {
		runCond = sync.NewCond(&sync.Mutex{})
		runCond.L.Lock()
	}

	go func() {
		for {
			that.run(func() {
				if wait {
					runCond.L.Lock()
					runCond.Signal()
					runCond.L.Unlock()
				}
			})

			// 如果上一次进程启动失败，并且启动时间少于2秒，则需要暂停一会，避免死循环，耗干资源
			if time.Now().Unix()-that.startTime.Unix() < 2 {
				time.Sleep(3 * time.Second)
			}
			if that.stopByUser {
				slog.Info("用户主动结束了该程序[%s]，不要再次启动", that.option.Name)
				break
			}
			// 判断进程是否需要自动重启
			if !that.isAutoRestart() {
				slog.Info("不要自动重启进程[%s],因为该进程设置了不需要自动重启", that.option.Name)
				break
			}
			slog.Info("因为该进程设置了自动重启,自动重启进程[%s],", that.option.Name)
		}
		that.lock.Lock()
		that.inStart = false
		that.lock.Unlock()
	}()

	if wait {
		runCond.Wait()
		runCond.L.Unlock()
	}
}

// Stop 主动停止进程
func (that *Process) Stop(wait bool) {
	that.lock.RLock()
	defer that.lock.RUnlock()

	if that.monitorCancel != nil {
		that.monitorCancel()
	}

	that.lock.Lock()
	that.stopByUser = true
	isRunning := that.isRunning()
	that.lock.Unlock()

	if !isRunning {
		slog.Info("程序[%s]未运行", that.GetName())
		return
	}
	slog.Info("正在停止程序[%s]", that.GetName())

	// 获取程序的正常退出信号
	sigs := that.option.StopSignal
	// 发送信号后的等待秒数
	waitSecond := time.Duration(that.option.StopWaitSecs) * time.Second
	// 发送强制结束信号后的等待秒数
	killWaitSecond := time.Duration(that.option.KillWaitSecs) * time.Second
	// 是否同时停止进程组
	stopAsGroup := that.option.StopAsGroup
	// 是否强制杀死进程组
	killAsGroup := that.option.KillAsGroup
	if stopAsGroup && !killAsGroup {
		slog.Error("不能够同时设置 stopAsGroup=true 和 killAsGroup=false")
	}
	var stopped int32 = 0

	// 添加 context 控制
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(that.option.StopWaitSecs)*time.Second)
	defer cancel()

	stopChan := make(chan struct{})
	go func() {
		defer close(stopChan)
		for i := 0; i < len(sigs) && atomic.LoadInt32(&stopped) == 0; i++ {
			// 获取需要发送的信号
			sig := signals.ToSignal(sigs[i])
			slog.Info("发送结束进程信号[%s]给进程[%s]", that.GetName(), sigs[i])
			// 发送结束进程信号给程序
			_ = that.Signal(sig, stopAsGroup)
			endTime := time.Now().Add(waitSecond)
			// 等待指定的时候后，判断当前进程是否还在存
			for endTime.After(time.Now()) {
				// 如果进程不存在了
				if that.state != Starting && that.state != Running && that.state != Stopping {
					atomic.StoreInt32(&stopped, 1)
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
		// 如果发送了设置的信号后，进程还未停止，则需要强制结束该进程
		if atomic.LoadInt32(&stopped) == 0 {
			slog.Info("强制结束程序[%s]", that.GetName())
			_ = that.Signal(syscall.SIGKILL, killAsGroup)
			killEndTime := time.Now().Add(killWaitSecond)
			for killEndTime.After(time.Now()) {
				// 如果进程结束成功
				if that.state != Starting && that.state != Running && that.state != Stopping {
					atomic.StoreInt32(&stopped, 1)
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			// 无论如何，发送了强杀信号后，默认认为它强杀成功
			atomic.StoreInt32(&stopped, 1)
		}
	}()

	_exit := func() {
		select {
		case <-ctx.Done():
			slog.Warn("停止进程[%s]超时", that.GetName())
		case <-stopChan:
			slog.Info("进程[%s]已停止", that.GetName())
		}
	}
	// 如果阻塞等待进程结束
	if wait {
		_exit()
		return
	}
	go _exit()
}

// 启动进程
func (that *Process) run(finishCb func()) {
	that.lock.Lock()
	defer that.lock.Unlock()

	// 判断进程是否正在运行
	if that.isRunning() {
		slog.Info("不能启动进程[%s],因为它正在运行中...", that.option.Name)
		finishCb()
		return
	}

	that.startTime = time.Now()
	atomic.StoreInt32(that.retryTimes, 0)
	// 指定启动多少秒后没有异常退出，则表示启动成功
	startSecs := that.option.StartSecs
	// 进程重启间隔秒数，默认是0，表示不间隔
	restartPause := that.option.RestartPause

	var once sync.Once
	finishCbWrapper := func() {
		once.Do(finishCb)
	}

	// 进程被用户结束
	for !that.stopByUser {

		// 如果进程启动失败，需要重试，则需要判断配置，重试启动是否需要间隔制定时间
		if restartPause > 0 && atomic.LoadInt32(that.retryTimes) != 0 {
			that.lock.Lock()
			slog.Info("不能立刻重启程序[%s],需要等待%d秒", that.option.Name, restartPause)
			time.Sleep(time.Duration(restartPause) * time.Second)
			that.lock.Unlock()
		}
		// 程序指定结束时间，如果在该时间内未退出，则表示进程启动成功
		endTime := time.Now().Add(time.Duration(startSecs) * time.Second)
		// 更新进程状态
		that.changeStateTo(Starting)
		// 启动次数+1
		atomic.AddInt32(that.retryTimes, 1)
		// 创建启动命令行
		err := that.createProgramCommand()
		if err != nil {
			that.failToStartProgram(fmt.Errorf("不能创建进程,error: %v", err), finishCbWrapper)
			break
		}
		// 启动程序
		err = that.cmd.Start()
		if err != nil {
			// 重试次数已经大于设置中的最大重试次数
			if atomic.LoadInt32(that.retryTimes) >= int32(that.option.StartRetries) {
				that.failToStartProgram(fmt.Errorf("重试次数已经达到最大重试次数 error: %v", err), finishCbWrapper)
				break
			} else {
				// 启动失败，再次重试
				slog.WithValue("error", err).Info("程序[%s]启动失败,再次重试", that.option.Name)
				that.changeStateTo(Backoff)
				continue
			}
		}
		// 设置标准输出日志的pid
		if that.stdoutLog != nil {
			that.stdoutLog.SetPid(that.Pid())
		}
		// 设置标准错误输出日志的pid
		if that.stderrLog != nil {
			that.stderrLog.SetPid(that.Pid())
		}
		monitorExited := int32(0)
		programExited := int32(0)
		// 如果未设置启动监视时长，则表示cmd.start成功就算该程序启动成功
		if startSecs <= 0 {
			slog.Info("程序[%s]启动成功", that.option.Name)
			that.changeStateTo(Running)
			go finishCbWrapper()
		} else {
			// 如果设置了启动监视时长，则表示需要程序启动了，稳定运行指定秒数后才算启动成功
			go func() {
				that.monitorProgramIsRunning(endTime, &monitorExited, &programExited)
				finishCbWrapper()
			}()
		}
		if that.stopTime.IsZero() {
			slog.Debugf("正在尝试启动[%s]请稍后...", that.option.Name)
		} else {
			slog.Debugf("进程正在运行[%s]等待退出", that.option.Name)
		}
		that.lock.Unlock()
		that.waitForExit(int64(startSecs))
		// 修改程序退出码
		atomic.StoreInt32(&programExited, 1)
		// 等待监控协程退出
		for atomic.LoadInt32(&monitorExited) == 0 {
			time.Sleep(time.Duration(10) * time.Millisecond)
		}
		that.lock.Lock()

		// 如果程序的运行状态为 Running，则更改它的状态
		if that.state == Running {
			that.changeStateTo(Exited)
			slog.Info("程序[%s]已经结束", that.option.Name)
			break
		} else {
			that.changeStateTo(Backoff)
		}
		// 如果重试次数已经超过了设置的最大重试次数
		if atomic.LoadInt32(that.retryTimes) >= int32(that.option.StartRetries) {
			that.failToStartProgram(fmt.Errorf("不能启动程序[%s],因为已经超出了它的最大重试值:%d", that.option.Name, that.option.StartRetries), finishCbWrapper)
			break
		}
	}
}

// 创建程序的cmd对象
func (that *Process) createProgramCommand() (err error) {
	// 创建命令对象
	that.cmd, err = that.option.CreateCommand()
	if err != nil {
		return err
	}
	// 设置程序运行时用户
	if that.setUser() != nil {
		return fmt.Errorf("设置程序运行时用户[%s]失败", that.option.User)
	}

	// 设置程序重启变化监控
	if err = that.setProgramRestartChangeMonitor(that.cmd.Args[0]); err != nil {
		slog.Error("设置程序重启监控失败: %v", err)
	}

	// 父进程退出，则它生成的子进程也全部退出
	that.sysProcAttrSetPGid(that.cmd.SysProcAttr)
	// 是否需要当前进程打开的句柄传给子进程
	if len(that.option.ExtraFiles) > 0 {
		that.cmd.ExtraFiles = that.option.ExtraFiles
	}
	// 设置进程运行的环境变量
	that.setEnv()
	// 设置程序的dir
	that.setDir()
	// 设置程序的运行日志存放未知
	that.setLog()
	// 程序的标准输入
	that.stdin, _ = that.cmd.StdinPipe()

	return nil
}

// 判断进程是否在运行
func (that *Process) isRunning() bool {
	if that.cmd != nil && that.cmd.Process != nil {
		if runtime.GOOS == "windows" {
			proc, err := os.FindProcess(that.cmd.Process.Pid)
			return proc != nil && err == nil
		}
		return that.cmd.Process.Signal(syscall.Signal(0)) == nil
	}
	return false
}

// IsAutoStart 在进程管理器启动的时候也自动启动
func (that *Process) IsAutoStart() bool {
	return that.option.AutoStart
}

// 设置进程运行的环境变量
func (that *Process) setEnv() {
	if that.option.Environment.Size() > 0 {
		_ = utils.SetMap(that.option.Environment.Map())
	}
	that.cmd.Env = utils.All()
}

// 设置进程的运行目录
func (that *Process) setDir() {
	dir := that.option.Directory
	if dir != "" {
		that.cmd.Dir = dir
	}
}

// 设置进程的运行日志存放文件
func (that *Process) setLog() {
	that.stdoutLog = that.createStdoutLogger()
	that.cmd.Stdout = that.stdoutLog
	if that.option.RedirectStderr {
		that.stderrLog = that.stdoutLog
	} else {
		that.stderrLog = that.createStderrLogger()
	}
	that.cmd.Stderr = that.stderrLog
}

// 设置程序启动失败状态
func (that *Process) failToStartProgram(err error, finishCb func()) {
	slog.WithValue("error", err).Error("程序[%s]启动失败", that.option.Name)
	that.changeStateTo(Fatal)
	finishCb()
}

// 监控进程是否正在运行中
func (that *Process) monitorProgramIsRunning(endTime time.Time, monitorExited *int32, programExited *int32) {
	// 未到超时时间
	for time.Now().Before(endTime) && atomic.LoadInt32(programExited) == 0 {
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
	atomic.StoreInt32(monitorExited, 1)

	that.lock.Lock()
	defer that.lock.Unlock()
	// 进程在此期间未退出
	if atomic.LoadInt32(programExited) == 0 && that.state == Starting {
		slog.Info("进程[%s]启动成功", that.option.Name)
		that.changeStateTo(Running)
	}
}

// 判断进程是否需要自动重启
func (that *Process) isAutoRestart() bool {
	autoRestart := that.option.AutoReStart

	if autoRestart == AutoReStartFalse {
		return false
	} else if autoRestart == AutoReStartTrue {
		return true
	} else {
		that.lock.RLock()
		defer that.lock.RUnlock()
		if that.cmd != nil && that.cmd.ProcessState != nil {
			exitCode, err := that.getExitCode()
			// 如果自动重启设置为unexpected，则表示，在配置中已明确的退出code不需要重启，
			// 不在预设的配置中的退出code则需要重启
			return err == nil && !that.inExitCodes(exitCode)
		}
	}
	return false
}

// 阻塞等待进程运行结束
func (that *Process) waitForExit(_ int64) {
	_ = that.cmd.Wait()
	if that.cmd.ProcessState != nil {
		slog.Info("程序[%s]已经运行结束，退出码为:%v", that.option.Name, that.cmd.ProcessState)
	} else {
		slog.Info("程序[%s]已经运行结束", that.option.Name)
	}
	that.lock.Lock()
	defer that.lock.Unlock()
	that.stopTime = time.Now()
	if that.stdoutLog != nil {
		_ = that.stdoutLog.Close()
	}
	if that.stderrLog != nil {
		_ = that.stderrLog.Close()
	}
}

// Clone 进程
func (that *Process) Clone() (*Process, error) {
	that.lock.Lock()
	defer that.lock.Unlock()

	var t time.Time
	proc := &Process{
		Manager:    that.Manager,
		option:     that.option,
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
func (that *Process) changeStateTo(procState State) {
	that.state = procState
}

// 监控程序重启变化
func (that *Process) setProgramRestartChangeMonitor(programPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(programPath); err != nil {
		return nil // 如果文件不存在，直接返回
	}

	// 获取程序文件的最后修改时间
	fileInfo, err := os.Stat(programPath)
	if err != nil {
		return err
	}

	// 存储最后修改时间
	that.lastModTime = fileInfo.ModTime()

	ctx, cancel := context.WithCancel(context.Background())
	that.monitorCancel = cancel

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 检查文件是否被修改
				currentInfo, err := os.Stat(programPath)
				if err != nil {
					continue
				}

				// 如果文件被修改且进程正在运行，则重启进程
				if currentInfo.ModTime() != that.lastModTime && that.state == Running {
					slog.Info("检测到程序文件[%s]发生变化，准备重启进程", programPath)
					that.lastModTime = currentInfo.ModTime()
					that.Stop(true)  // 停止当前进程
					that.Start(true) // 重启进程
				}
			}
		}
	}()

	return nil
}
