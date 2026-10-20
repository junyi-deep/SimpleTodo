//go:build windows

package desktop

import (
	"syscall"
	"unsafe"
)

var user32 = syscall.NewLazyDLL("user32.dll")

func SetLevel(p unsafe.Pointer, mode string) error {
	after := uintptr(0)
	if mode == "top" {
		after = ^uintptr(0)
	} else if mode == "desktop" {
		after = 1
	} else {
		after = ^uintptr(1)
	}
	r, _, err := user32.NewProc("SetWindowPos").Call(uintptr(p), after, 0, 0, 0, 0, 0x0001|0x0002|0x0010)
	if r == 0 {
		return err
	}
	return nil
}
func Pointer() (int, int) {
	var p struct{ X, Y int32 }
	user32.NewProc("GetCursorPos").Call(uintptr(unsafe.Pointer(&p)))
	return int(p.X), int(p.Y)
}
func Supported() bool { return true }

func MouseDown() bool { v, _, _ := user32.NewProc("GetAsyncKeyState").Call(1); return v&0x8000 != 0 }
