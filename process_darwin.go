//go:build darwin

package process

import "syscall"

func (p *Process) sysProcAttrSetPGid(*syscall.SysProcAttr) {
}
