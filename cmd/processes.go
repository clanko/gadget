package cmd

import (
	"syscall"
)

func KillPid(pid int) {
	pgid, err := syscall.Getpgid(pid)
	if err == nil {
		err := syscall.Kill(-pgid, 9)

		if err != nil {
			PrintfDanger(err.Error())
		}
	}
}
