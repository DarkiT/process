//go:build linux

package process

import (
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

func (p *Process) setUser() error {
	userName := p.option.User
	if len(userName) == 0 {
		return nil
	}

	pos := strings.Index(userName, ":")
	groupName := ""
	if pos != -1 {
		groupName = userName[pos+1:]
		userName = userName[0:pos]
	}

	u, err := user.Lookup(userName)
	if err != nil {
		return err
	}

	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return err
	}

	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil && groupName == "" {
		return err
	}

	if groupName != "" {
		g, err := user.LookupGroup(groupName)
		if err != nil {
			return err
		}
		gid, err = strconv.ParseUint(g.Gid, 10, 32)
		if err != nil {
			return err
		}
	}

	p.cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid:         uint32(uid),
		Gid:         uint32(gid),
		NoSetGroups: true,
	}
	return nil
}
