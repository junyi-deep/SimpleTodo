//go:build windows

package localcalendar

import (
	"os/exec"
	"syscall"
)

func hideCalendarProcess(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
