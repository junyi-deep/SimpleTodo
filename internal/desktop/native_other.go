//go:build !darwin && !windows

package desktop

import (
	"fmt"
	"unsafe"
)

func SetLevel(p unsafe.Pointer, mode string) error {
	return fmt.Errorf("UNSUPPORTED: 当前平台不支持桌面层级")
}
func Pointer() (int, int) { return 0, 0 }
func Supported() bool     { return false }

func MouseDown() bool { return false }
