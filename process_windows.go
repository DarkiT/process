//go:build windows

package process

import "syscall"

func (p *Process) sysProcAttrSetPGid(_ *syscall.SysProcAttr) {
}
