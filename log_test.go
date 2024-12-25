package process

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	logger := newDefaultLogger()

	// 测试各种日志级别
	// 这些调用不应该panic
	logger.Infof("test info message")
	logger.Debugf("test debug message")
	logger.Warnf("test warn message")
	logger.Errorf("test error message")

	// 测试格式化
	logger.Infof("test %s %d", "formatted", 123)
}

func TestProcessLogger(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建测试进程
	manager := NewManager()
	proc, err := manager.NewProcess(
		WithName("test-logger"),
		WithCommand("echo"),
		WithArgs("test"),
		WithStdoutLog(
			filepath.Join(tmpDir, "stdout.log"),
			"1MB",
			2,
		),
		WithStderrLog(
			filepath.Join(tmpDir, "stderr.log"),
			"1MB",
			2,
		),
	)
	if err != nil {
		t.Fatalf("Failed to create process: %v", err)
	}

	// 测试日志文件路径
	stdoutLogfile := proc.GetStdoutLogfile()
	if stdoutLogfile == "" {
		t.Error("Expected non-empty stdout logfile path")
	}

	stderrLogfile := proc.GetStderrLogfile()
	if stderrLogfile == "" {
		t.Error("Expected non-empty stderr logfile path")
	}

	// 测试创建日志记录器
	stdoutLogger := proc.createStdoutLogger()
	if stdoutLogger == nil {
		t.Error("Failed to create stdout logger")
	}

	stderrLogger := proc.createStderrLogger()
	if stderrLogger == nil {
		t.Error("Failed to create stderr logger")
	}

	// 启动进程并验证日志文件创建
	proc.Start(true)

	if _, err := os.Stat(stdoutLogfile); os.IsNotExist(err) {
		t.Error("Stdout log file was not created")
	}

	proc.Stop(true)
}
