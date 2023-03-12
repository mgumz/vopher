//go:build !windows

package ui

import (
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

func terminalSize(t *os.File) (cols, rows int, err error) {

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

func cursorNUp(t *os.File, n int) (err error) {
	_, _ = t.WriteString("\x1b[")
	_, _ = t.WriteString(strconv.Itoa(n))
	_, _ = t.WriteString("A")
	return
}
