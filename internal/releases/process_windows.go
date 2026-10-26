//go:build windows

package releases

import (
	"golang.org/x/sys/windows"
	"os/exec"
	"syscall"
)

func processAlive(pid int) bool {
	h, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return err != windows.ERROR_INVALID_PARAMETER
	}
	defer windows.CloseHandle(h)
	state, err := windows.WaitForSingleObject(h, 0)
	return err != nil || state != windows.WAIT_OBJECT_0
}
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000200 | 0x00000008, HideWindow: true}
}
