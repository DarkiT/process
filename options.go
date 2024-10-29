package process

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/darkit/process/utils"
)

type ProcOption func(*ProcOptions)

type AutoReStart int

const (
	AutoReStartUnexpected AutoReStart = iota
	AutoReStartTrue       AutoReStart = iota // 1
	AutoReStartFalse      AutoReStart = iota // 0
)

type ProcOptions struct {
	Name         string      // 进程名称
	Command      string      // 启动命令
	Args         []string    // 启动参数
	Directory    string      // 进程运行目录
	AutoStart    bool        // 启动的时候自动该进程启动
	StartSecs    int         // 启动10秒后没有异常退出，就表示进程正常启动了，默认为1秒
	AutoReStart  AutoReStart // 程序退出后自动重启,可选值：[unexpected,true,false]，默认为unexpected，表示进程意外杀死后才重启
	ExitCodes    []int       // 进程退出的code值
	StartRetries int         // 启动失败自动重试次数，默认是3
	RestartPause int         // 进程重启间隔秒数，默认是0，表示不间隔
	User         string      // 用哪个用户启动进程，默认是父进程的所属用户
	Priority     int         // 进程启动优先级，默认999，值小的优先启动

	StdoutLogfile         string // 日志文件，不存在时 supervisord 会自动创建日志文件）
	StdoutLogFileMaxBytes int    // stdout 日志文件大小，默认50MB
	StdoutLogFileBackups  int    // stdout 日志文件备份数，默认是10
	RedirectStderr        bool   // 把stderr重定向到stdout，默认false
	StderrLogfile         string // 日志文件，进程启动后的标准错误写入该文件
	StderrLogFileMaxBytes int    // stderr 日志文件大小，默认50MB
	StderrLogFileBackups  int    // stderr 日志文件备份数，默认是10

	StopAsGroup              bool             // 默认为false,进程被杀死时，是否向这个进程组发送stop信号，包括子进程
	KillAsGroup              bool             // 默认为false，向进程组发送kill信号，包括子进程
	StopSignal               []string         // 结束进程发送的信号
	StopWaitSecs             int              // 发送结束进程的信号后等待的秒数
	KillWaitSecs             int              // 强杀进程等待秒数
	Environment              *utils.StrStrMap // 环境变量
	RestartWhenBinaryChanged bool             // 当进程的二进制文件有修改，是否需要重启,默认false
	ExtraFiles               []*os.File       // 继承主进程已经打开的文件列表
	Extend                   *utils.AnyAnyMap // 扩展参数
}

// ProcName 设置进程名称
func ProcName(opt string) ProcOption {
	return func(options *ProcOptions) {
		options.Name = opt
	}
}

// ProcCommand 启动命令
func ProcCommand(opt string) ProcOption {
	return func(options *ProcOptions) {
		options.Command = opt
	}
}

// ProcArgs 启动参数
func ProcArgs(opt ...string) ProcOption {
	return func(options *ProcOptions) {
		options.Args = opt
	}
}

// ProcAutoStart 启动的时候自动该进程启动
func ProcAutoStart(opt bool) ProcOption {
	return func(options *ProcOptions) {
		options.AutoStart = opt
	}
}

// ProcDirectory 进程运行目录
func ProcDirectory(opt string) ProcOption {
	return func(options *ProcOptions) {
		options.Directory = opt
	}
}

// ProcStartSecs 指定启动多少秒后没有异常退出，则表示启动成功
// // 未设置该值，则表示cmd.Start方法调用为出错，则表示启动成功，
// // 设置了该值，则表示程序启动后需稳定运行指定的秒数后才算启动成功
func ProcStartSecs(opt int) ProcOption {
	return func(options *ProcOptions) {
		options.StartSecs = opt
	}
}

// ProcAutoReStart 程序退出后自动重启,可选值：[unexpected,true,false]，默认为unexpected，表示进程意外杀死后才重启
func ProcAutoReStart(opt AutoReStart) ProcOption {
	return func(options *ProcOptions) {
		options.AutoReStart = opt
	}
}

// ProcExitCodes 进程退出的code值列表，该列表中的值表示已知
func ProcExitCodes(opt ...int) ProcOption {
	return func(options *ProcOptions) {
		options.ExitCodes = opt
	}
}

// ProcStartRetries 启动失败自动重试次数，默认是3
func ProcStartRetries(opt int) ProcOption {
	return func(options *ProcOptions) {
		options.StartRetries = opt
	}
}

// ProcRestartPause 进程重启间隔秒数，默认是0，表示不间隔
func ProcRestartPause(opt int) ProcOption {
	return func(options *ProcOptions) {
		options.RestartPause = opt
	}
}

// ProcUser 用哪个用户启动进程，默认是父进程的所属用户
func ProcUser(opt string) ProcOption {
	return func(options *ProcOptions) {
		options.User = opt
	}
}

// ProcPriority 进程启动优先级，默认999，值小的优先启动
func ProcPriority(opt int) ProcOption {
	return func(options *ProcOptions) {
		options.Priority = opt
	}
}

// ProcStopAsGroup 默认为false,进程被杀死时，是否向这个进程组发送stop信号，包括子进程
func ProcStopAsGroup(opt bool) ProcOption {
	return func(options *ProcOptions) {
		options.StopAsGroup = opt
	}
}

// ProcKillAsGroup 默认为false，向进程组发送kill信号，包括子进程
func ProcKillAsGroup(opt bool) ProcOption {
	return func(options *ProcOptions) {
		options.KillAsGroup = opt
	}
}

// ProcStopSignal 结束进程发送的信号列表
func ProcStopSignal(opt ...string) ProcOption {
	return func(options *ProcOptions) {
		options.StopSignal = opt
	}
}

// ProcStopWaitSecs 发送结束进程的信号后等待的秒数
func ProcStopWaitSecs(opt int) ProcOption {
	return func(options *ProcOptions) {
		options.StopWaitSecs = opt
	}
}

// ProcKillWaitSecs 强杀进程等待秒数
func ProcKillWaitSecs(opt int) ProcOption {
	return func(options *ProcOptions) {
		options.KillWaitSecs = opt
	}
}

// ProcSetEnvironment 环境变量
func ProcSetEnvironment(key, val string) ProcOption {
	return func(options *ProcOptions) {
		options.Environment.Set(key, val)
	}
}

func ProcEnvironment(opt map[string]string) ProcOption {
	return func(options *ProcOptions) {
		options.Environment.Sets(opt)
	}
}

// ProcRestartWhenBinaryChanged 当进程的二进制文件有修改，是否需要重启
func ProcRestartWhenBinaryChanged(opt bool) ProcOption {
	return func(options *ProcOptions) {
		options.RestartWhenBinaryChanged = opt
	}
}

// ProcExtraFiles 设置打开的文件句柄列表
func ProcExtraFiles(opt []*os.File) ProcOption {
	return func(options *ProcOptions) {
		options.ExtraFiles = opt
	}
}

// ProcSetExtend 扩展参数
func ProcSetExtend(key, val interface{}) ProcOption {
	return func(options *ProcOptions) {
		options.Extend.Set(key, val)
	}
}

// ProcStdoutLog 设置stdoutlog的存放配置
func ProcStdoutLog(file string, maxBytes string, backups ...int) ProcOption {
	return func(options *ProcOptions) {
		options.StdoutLogfile = file
		options.StdoutLogFileMaxBytes = utils.GetBytes(maxBytes, 50*1024*1024)
		options.StdoutLogFileBackups = 10
		if len(backups) > 0 {
			options.StdoutLogFileBackups = backups[0]
		}
	}
}

// ProcStderrLog 设置stderrlog的存放配置
func ProcStderrLog(file string, maxBytes string, backups ...int) ProcOption {
	return func(options *ProcOptions) {
		options.StderrLogfile = file
		options.StderrLogFileMaxBytes = utils.GetBytes(maxBytes, 50*1024*1024)
		options.StderrLogFileBackups = 10
		if len(backups) > 0 {
			options.StderrLogFileBackups = backups[0]
		}
	}
}

// ProcRedirectStderr 错误输出是否与标准输入一起
func ProcRedirectStderr(opt bool) ProcOption {
	return func(options *ProcOptions) {
		options.RedirectStderr = opt
	}
}

// NewProcOptions 创建进程启动配置
func NewProcOptions(opts ...ProcOption) ProcOptions {
	proc := ProcOptions{
		AutoStart:    true,
		StartSecs:    1,
		AutoReStart:  AutoReStartTrue,
		StartRetries: 3,
		RestartPause: 0,
		StopWaitSecs: 10,
		KillWaitSecs: 2,
		//User:                     "root",
		Priority:                 999,
		StopAsGroup:              false,
		KillAsGroup:              false,
		RestartWhenBinaryChanged: false,
		Extend:                   utils.NewAnyAnyMap(),
		Environment:              utils.NewStrStrMap(),

		StdoutLogfile:         "",
		StdoutLogFileMaxBytes: 50 * 1024 * 1024,
		StdoutLogFileBackups:  10,
		RedirectStderr:        false,
		StderrLogfile:         "",
		StderrLogFileMaxBytes: 50 * 1024 * 1024,
		StderrLogFileBackups:  10,
	}
	for _, opt := range opts {
		opt(&proc)
	}
	return proc
}

// CreateCommand 根据就配置生成cmd对象
func (o ProcOptions) CreateCommand() (*exec.Cmd, error) {
	if len(o.Name) <= 0 {
		o.Name = o.Command
	}

	cmd := exec.Command(o.Command)
	if len(o.Args) > 0 {
		cmd.Args = append([]string{o.Command}, o.Args...)
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{}
	return cmd, nil
}
