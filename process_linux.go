//go:build linux

package process

import "syscall"

func (p *Process) sysProcAttrSetPGid(s *syscall.SysProcAttr) {
	s.Setpgid = true
	s.Pdeathsig = syscall.SIGKILL
}
