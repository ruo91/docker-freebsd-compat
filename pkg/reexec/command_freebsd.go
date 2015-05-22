// +build freebsd

package reexec

import (
	"os/exec"
)

func Command(args ...string) *exec.Cmd {
	return &exec.Cmd{
		Path: Self(),
		Args: args,
		//SysProcAttr: &syscall.SysProcAttr{
		//	Pdeathsig: syscall.SIGTERM,
		//},
	}
}
