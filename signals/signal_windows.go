//go:build windows

package signals

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// ToSignal 传入信号字符串，返回标准信号
func ToSignal(signalName string) os.Signal {
	if signalName == "HUP" {
		return syscall.SIGHUP
	} else if signalName == "INT" {
		return syscall.SIGINT
	} else if signalName == "QUIT" {
		return syscall.SIGQUIT
	} else if signalName == "KILL" {
		return syscall.SIGKILL
	} else if signalName == "USR1" {
		return syscall.SIGTERM
	} else if signalName == "USR2" {
		return syscall.SIGTERM
	} else {
		return syscall.SIGTERM
	}
}

// Kill 向指定的进程发送信号
// process: 进程对象
// sig: 信号
// sigChildren: windows 下会忽略这个参数
func Kill(process *os.Process, sig os.Signal, sigChildren bool) error {
	// Signal command can't kill children processes, call  taskkill command to kill them
	cmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", process.Pid))
	err := cmd.Start()
	if err == nil {
		return cmd.Wait()
	}
	// if fail to find taskkill, fallback to normal signal
	return process.Signal(sig)
}

// KillPid 向指定的Pid发送信号
// pid: 进程pid
// sig: 信号
// sigChildren: 如果为true，则信号会发送到该进程的子进程
func KillPid(pid int, sig os.Signal, sigChildren ...bool) error {
	// Signal command can't kill children processes, call  taskkill command to kill them
	cmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", pid))
	err := cmd.Start()
	if err == nil {
		return cmd.Wait()
	}
	return nil
}

// CheckPidExist 检查进程是否存在
func CheckPidExist(pid int) bool {
	// 尝试获取进程句柄
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// 在 Windows 上，FindProcess 总是返回非 nil 的进程和 nil 错误
	// 需要额外调用 Signal(0) 来验证进程是否真实存在
	return process.Signal(syscall.Signal(0)) == nil
}
