package process

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/darkit/process/proclog"
)

// Logger 日志记录器接口
type Logger interface {
	Infof(format string, args ...any)
	Debugf(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

// defaultLogger 默认日志记录器
type defaultLogger struct {
	logger *slog.Logger
}

// newDefaultLogger 创建默认日志记录器
func newDefaultLogger() *defaultLogger {
	return &defaultLogger{
		logger: slog.Default(),
	}
}

func (l *defaultLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(sprintf(format, args...))
}

func (l *defaultLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug(sprintf(format, args...))
}

func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn(sprintf(format, args...))
}

func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error(sprintf(format, args...))
}

// sprintf 格式化字符串
func sprintf(format string, args ...interface{}) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}

// 创建标准输出日志
func (that *Process) createStdoutLogger() proclog.Logger {
	logFile := that.GetStdoutLogfile()
	maxBytes := int64(that.option.StdoutLogFileMaxBytes)
	backups := that.option.StdoutLogFileBackups

	props := map[string]string{
		"process": that.GetName(),
		"type":    "stdout",
		"pid":     strconv.Itoa(that.Pid()),
	}

	return proclog.NewLogger(that.GetName(), logFile, proclog.NewNullLocker(), maxBytes, backups, props)
}

// 创建标准错误日志
func (that *Process) createStderrLogger() proclog.Logger {
	logFile := that.GetStderrLogfile()
	maxBytes := int64(that.option.StderrLogFileMaxBytes)
	backups := that.option.StderrLogFileBackups

	props := map[string]string{
		"process": that.GetName(),
		"type":    "stderr",
		"pid":     strconv.Itoa(that.Pid()),
	}

	return proclog.NewLogger(that.GetName(), logFile, proclog.NewNullLocker(), maxBytes, backups, props)
}
