// +build !windows

package appos

import (
	"fmt"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"
)

type OsExecutor struct {
}

func (oe OsExecutor) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

func (oe OsExecutor) SetUser(userName, path string, cmd *exec.Cmd) error {
	if userName == "" {
		return nil
	}
	logrus.Debugf("will set user %s", userName)

	uid, gid, err := oe.parse(userName, path)
	if err != nil {
		return err
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Credential: &syscall.Credential{Uid: uid, Gid: gid},
	}

	return nil
}

func (oe OsExecutor) parse(userName, path string) (sysUserID, sysGroupID uint32, err error) {
	logrus.Debugf("parsing user '%s' to get uid and group id from OS", userName)
	u, err := user.Lookup(userName)
	if err != nil {
		err = fmt.Errorf("cannot locate user '%s': %w, check path '%s'", userName, err, path+".user")
		return
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		err = fmt.Errorf("non-numeric user ID '%s': %w, check path '%s'", u.Uid, err, path+".user")
		return
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		err = fmt.Errorf("non-numeric user group ID '%s': %w, check path '%s'", u.Gid, err, path+".user")
		return
	}

	return uint32(uid), uint32(gid), nil
}
