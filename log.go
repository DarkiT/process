package process

import (
	"strconv"

	"github.com/darkit/process/proclog"
)

func (p *Process) createStdoutLogger() proclog.Logger {
	logFile := p.GetStdoutLogfile()
	maxBytes := int64(p.option.StdoutLogFileMaxBytes)
	backups := p.option.StdoutLogFileBackups

	props := map[string]string{
		"process": p.GetName(),
		"type":    "stdout",
		"pid":     strconv.Itoa(p.Pid()),
	}

	return proclog.NewLogger(p.GetName(), logFile, proclog.NewNullLocker(), maxBytes, backups, props)
}

func (p *Process) createStderrLogger() proclog.Logger {
	logFile := p.GetStderrLogfile()
	maxBytes := int64(p.option.StderrLogFileMaxBytes)
	backups := p.option.StderrLogFileBackups

	props := map[string]string{
		"process": p.GetName(),
		"type":    "stderr",
		"pid":     strconv.Itoa(p.Pid()),
	}

	return proclog.NewLogger(p.GetName(), logFile, proclog.NewNullLocker(), maxBytes, backups, props)
}
