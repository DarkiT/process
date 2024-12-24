//go:build linux

package process

import (
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
	Pid              int
	Options          int
	DisablePid1Check bool
}

var debug bool = false

func EnableDebug() {
	debug = true
}

func DisableDebug() {
	debug = false
}

func sigChildHandler(notifications chan os.Signal) {
	sigs := make(chan os.Signal, 3)
	signal.Notify(sigs, syscall.SIGCHLD)

	for {
		sig := <-sigs
		select {
		case notifications <- sig:
		default:
		}
	}
}

func reapChildren(config Config) {
	notifications := make(chan os.Signal, 1)

	go sigChildHandler(notifications)

	pid := config.Pid
	opts := config.Options

	for {
		sig := <-notifications
		if debug {
			fmt.Printf(" - Received signal %v\n", sig)
		}
		for {
			var wstatus syscall.WaitStatus

			pid, err := syscall.Wait4(pid, &wstatus, opts, nil)
			for syscall.EINTR == err {
				pid, err = syscall.Wait4(pid, &wstatus, opts, nil)
			}

			if syscall.ECHILD == err {
				break
			}

			if debug {
				fmt.Printf(" - Grim reaper cleanup: pid=%d, wstatus=%+v\n", pid, wstatus)
			}
		}
	}
}

func ReapZombie() {
	go Reap()
}

func Reap() {
	Start(Config{
		Pid:              -1,
		Options:          0,
		DisablePid1Check: false,
	})
}

func Start(config Config) {
	if !config.DisablePid1Check {
		mypid := os.Getpid()
		if 1 != mypid {
			if debug {
				fmt.Printf(" - Grim reaper disabled, pid not 1\n")
			}
			return
		}
	}

	go reapChildren(config)
}
