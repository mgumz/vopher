// + build !windows
package main

import (
	"os"
	"syscall"
	"unsafe"
)

func TerminalSize(t *os.File) (cols, rows int, err error) {

	var ws = struct {
		row, col       uint16
		xpixel, ypixel uint16
	}{}

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		t.Fd(),
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&ws)),
	)

	if errno != 0 {
		return -1, -1, syscall.Errno(errno)
	}

	return int(ws.col), int(ws.row), nil
}
