package process

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/darkit/process/utils"
)

const (
	AutoReStartUnexpected AutoReStart = iota
	AutoReStartTrue       AutoReStart = iota // 1
	AutoReStartFalse      AutoReStart = iota // 0
)

// AutoReStart 定义自动重启类型
type AutoReStart uint8

// Options 进程配置选项
type Options struct {
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

// WithOption 定义选项函数类型
type WithOption func(*Options)

// WithName 设置进程名称
func WithName(opt string) WithOption {
	return func(options *Options) {
		options.Name = opt
	}
}

// WithCommand 启动命令
func WithCommand(opt string) WithOption {
	return func(options *Options) {
		options.Command = opt
	}
}

// WithArgs 启动参数
func WithArgs(opt ...string) WithOption {
	return func(options *Options) {
		options.Args = opt
	}
}

// WithAutoStart 启动的时候自动该进程启动
func WithAutoStart(opt bool) WithOption {
	return func(options *Options) {
		options.AutoStart = opt
	}
}

// WithDirectory 进程运行目录
func WithDirectory(opt string) WithOption {
	return func(options *Options) {
		options.Directory = opt
	}
}

// WithStartSecs 指定启动多少秒后没有异常退出，则表示启动成功
// // 未设置该值，则表示cmd.Start方法调用为出错，则表示启动成功，
// // 设置了该值，则表示程序启动后需稳定运行指定的秒数后才算启动成功
func WithStartSecs(opt int) WithOption {
	return func(options *Options) {
		options.StartSecs = opt
	}
}

// WithAutoReStart 程序退出后自动重启,可选值：[unexpected,true,false]，默认为unexpected，表示进程意外杀死后才重启
func WithAutoReStart(opt AutoReStart) WithOption {
	return func(options *Options) {
		options.AutoReStart = opt
	}
}

// WithExitCodes 进程退出的code值列表，该列表中的值表示已知
func WithExitCodes(opt ...int) WithOption {
	return func(options *Options) {
		options.ExitCodes = opt
	}
}

// WithStartRetries 启动失败自动重试次数，默认是3
func WithStartRetries(opt int) WithOption {
	return func(options *Options) {
		options.StartRetries = opt
	}
}

// WithRestartPause 进程重启间隔秒数，默认是0，表示不间隔
func WithRestartPause(opt int) WithOption {
	return func(options *Options) {
		options.RestartPause = opt
	}
}

// WithUser 用哪个用户启动进程，默认是父进程的所属用户
func WithUser(opt string) WithOption {
	return func(options *Options) {
		options.User = opt
	}
}

// WithPriority 进程启动优先级，默认999，值小的优先启动
func WithPriority(opt int) WithOption {
	return func(options *Options) {
		options.Priority = opt
	}
}

// WithStopAsGroup 默认为false,进程被杀死时，是否向这个进程组发送stop信号，包括子进程
func WithStopAsGroup(opt bool) WithOption {
	return func(options *Options) {
		options.StopAsGroup = opt
	}
}

// WithKillAsGroup 默认为false，向进程组发送kill信号，包括子进程
func WithKillAsGroup(opt bool) WithOption {
	return func(options *Options) {
		options.KillAsGroup = opt
	}
}

// WithStopSignal 结束进程发送的信号列表
func WithStopSignal(opt ...string) WithOption {
	return func(options *Options) {
		options.StopSignal = opt
	}
}

// WithStopWaitSecs 发送结束进程的信号后等待的秒数
func WithStopWaitSecs(opt int) WithOption {
	return func(options *Options) {
		options.StopWaitSecs = opt
	}
}

// WithKillWaitSecs 强杀进程等待秒数
func WithKillWaitSecs(opt int) WithOption {
	return func(options *Options) {
		options.KillWaitSecs = opt
	}
}

// WithSetEnvironment 环境变量
func WithSetEnvironment(key, val string) WithOption {
	return func(options *Options) {
		options.Environment.Set(key, val)
	}
}

func WithEnvironment(opt map[string]string) WithOption {
	return func(options *Options) {
		options.Environment.Sets(opt)
	}
}

// WithRestartWhenBinaryChanged 当进程的二进制文件有修改，是否需要重启
func WithRestartWhenBinaryChanged(opt bool) WithOption {
	return func(options *Options) {
		options.RestartWhenBinaryChanged = opt
	}
}

// WithExtraFiles 设置打开的文件句柄列表
func WithExtraFiles(opt []*os.File) WithOption {
	return func(options *Options) {
		options.ExtraFiles = opt
	}
}

// WithSetExtend 扩展参数
func WithSetExtend(key, val interface{}) WithOption {
	return func(options *Options) {
		options.Extend.Set(key, val)
	}
}

// WithStdoutLog 设置stdoutlog的存放配置
func WithStdoutLog(file string, maxBytes string, backups ...int) WithOption {
	return func(options *Options) {
		options.StdoutLogfile = file
		options.StdoutLogFileMaxBytes = utils.GetBytes(maxBytes, 50*1024*1024)
		options.StdoutLogFileBackups = 10
		if len(backups) > 0 {
			options.StdoutLogFileBackups = backups[0]
		}
	}
}

// WithStderrLog 设置stderrlog的存放配置
func WithStderrLog(file string, maxBytes string, backups ...int) WithOption {
	return func(options *Options) {
		options.StderrLogfile = file
		options.StderrLogFileMaxBytes = utils.GetBytes(maxBytes, 50*1024*1024)
		options.StderrLogFileBackups = 10
		if len(backups) > 0 {
			options.StderrLogFileBackups = backups[0]
		}
	}
}

// WithRedirectStderr 错误输出是否与标准输入一起
func WithRedirectStderr(opt bool) WithOption {
	return func(options *Options) {
		options.RedirectStderr = opt
	}
}

// NewOptions 创建进程启动配置
func NewOptions(opts ...WithOption) Options {
	proc := Options{
		AutoStart:                true,
		StartSecs:                1,
		AutoReStart:              AutoReStartTrue,
		StartRetries:             3,
		RestartPause:             0,
		StopWaitSecs:             15,
		KillWaitSecs:             2,
		Priority:                 999,
		StopAsGroup:              false,
		KillAsGroup:              false,
		RestartWhenBinaryChanged: false,
		Extend:                   utils.NewAnyAnyMap(),
		Environment:              utils.NewStrStrMap(),
		StdoutLogFileMaxBytes:    50 * 1024 * 1024,
		StdoutLogFileBackups:     10,
		RedirectStderr:           false,
		StderrLogFileMaxBytes:    50 * 1024 * 1024,
		StderrLogFileBackups:     10,
		// User:                     "root",
		// StdoutLogfile:            "",
		// StderrLogfile:            "",
	}
	for _, opt := range opts {
		opt(&proc)
	}
	return proc
}

// CreateCommand 根据就配置生成cmd对象
func (that Options) CreateCommand() (*exec.Cmd, error) {
	if len(that.Name) <= 0 {
		that.Name = that.Command
	}
	cmd := exec.Command(that.Command)
	if len(that.Args) > 0 {
		cmd.Args = append([]string{that.Command}, that.Args...)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	return cmd, nil
}
