package process

import "github.com/darkit/process/proclog"

// 创建标准输出日志
func (that *Process) createStdoutLogger() proclog.Logger {
	logFile := that.GetStdoutLogfile()
	maxBytes := int64(that.option.StdoutLogFileMaxBytes)
	backups := that.option.StdoutLogFileBackups

	props := make(map[string]string)

	return proclog.NewLogger(that.GetName(), logFile, proclog.NewNullLocker(), maxBytes, backups, props)
}

// 创建标准错误日志
func (that *Process) createStderrLogger() proclog.Logger {
	logFile := that.GetStderrLogfile()
	maxBytes := int64(that.option.StderrLogFileMaxBytes)
	backups := that.option.StderrLogFileBackups

	props := make(map[string]string)

	return proclog.NewLogger(that.GetName(), logFile, proclog.NewNullLocker(), maxBytes, backups, props)
}
