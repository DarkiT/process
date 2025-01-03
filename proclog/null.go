package proclog

import (
	"errors"
	"fmt"
)

type NullLogger struct {
	// logEventEmitter LogEventEmitter
}

func NewNullLogger() *NullLogger {
	return &NullLogger{}
}

func (that *NullLogger) SetPid(pid int) {
	// NOTHING TO DO
}

func (that *NullLogger) Write(p []byte) (int, error) {
	return len(p), nil
}

func (that *NullLogger) Close() error {
	return nil
}

func (that *NullLogger) ReadLog(offset int64, length int64) (string, error) {
	return "", errors.New("NO_FILE")
}

func (that *NullLogger) ReadTailLog(offset int64, length int64) (string, int64, bool, error) {
	return "", 0, false, errors.New("NO_FILE")
}

func (that *NullLogger) ClearCurLogFile() error {
	return fmt.Errorf("NoLog")
}

func (that *NullLogger) ClearAllLogFile() error {
	return errors.New("NO_FILE")
}
