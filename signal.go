package process

import (
	"fmt"
	"os"

	"github.com/darkit/process/signals"
	"github.com/darkit/slog"
)

// Signal 向进程发送信号
// sig: 要发送的信号
// sigChildren: 如果为true，则信号会发送到该进程的子进程
func (p *Process) Signal(sig os.Signal, sigChildren bool) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.sendSignal(sig, sigChildren)
}

// 发送多个信号到进程
// sig: 要发送的信号列表
// sigChildren: 如果为true，则信号会发送到该进程的子进程
func (p *Process) sendSignals(sigs []string, sigChildren bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, strSig := range sigs {
		sig := signals.ToSignal(strSig)
		err := p.sendSignal(sig, sigChildren)
		if err != nil {
			slog.Infof("向进程[%s]发送信号[%s]失败,err:%v", p.GetName(), strSig, err)
		}
	}
}

// sendSignal 向进程发送信号
// sig: 要发送的信号
// sigChildren: 如果为true，则信号会发送到该进程的子进程
func (p *Process) sendSignal(sig os.Signal, sigChildren bool) error {
	if p.cmd != nil && p.cmd.Process != nil {
		slog.Infof("发送信号[%s]到进程[%s]", sig, p.GetName())
		return signals.Kill(p.cmd.Process, sig, sigChildren)
	}
	return fmt.Errorf("进程[%s]没有启动", p.GetName())
}
