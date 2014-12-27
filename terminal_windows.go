// +build windows

package main

// note1: a lot of this file contains insight from termbox-go. termbox-go
// clears the screens on termbox.Init() and this is something i don't like
// to have for some of vopher's ui-modes. that is why a very limited portion
// of termbox-go's windows-code is duplicated here
//
// note2: termbox-go uses a special 'out' file, opened with
// syscall.Open("CONOUT$",...). why? i don't know (yet).
// termbox-go used syscall.STD_INPUT_HANDLE before and then
// switched to CONOUT$ (see [1]). this is strange as
// golang fills syscall.Stdout by calling GetStdHandle(STD_OUTPUT_HANDLE)
// which then should return "CONOUT$" by itself, see [2].
//
// note3: without the fmt.Println() statements we get an exception.
// whyever. TODO for another time.
//
// [1]: https://github.com/nsf/termbox-go/commit/5a8306fed5a06766bae36aaac383174748187ccf
// [2]: http://msdn.microsoft.com/en-us/library/windows/desktop/ms683231(v=vs.85).aspx

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32                   = syscall.NewLazyDLL("kernel32.dll")
	GetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	SetConsoleCursorPosition   = kernel32.NewProc("SetConsoleCursorPosition")
)

type _COORD struct {
	x, y int16
}

func (c _COORD) uintptr() uintptr {
	return uintptr(*(*int32)(unsafe.Pointer(&c)))
}

type _SMALL_RECT struct {
	left, top, right, bottom int16
}

type _CONSOLE_SCREEN_BUFFER_INFO struct {
	size           _COORD
	cursorPosition _COORD
	srWindow       _SMALL_RECT
	maxWindowSize  _COORD
}

func _GetConsoleScreenBufferInfo(t *os.File, buffer_info *_CONSOLE_SCREEN_BUFFER_INFO) error {

	ret, _, errno := syscall.Syscall(GetConsoleScreenBufferInfo.Addr(),
		2,
		uintptr(t.Fd()),
		uintptr(unsafe.Pointer(buffer_info)),
		0)

	if ret == 0 || errno != 0 {
		// TODO: find out why removing the next line leads to
		//    Exception 0xc0000005 0x8 0x400058 0x400058
		//    PC=0x400058
		fmt.Println(errno.Error())
		return errno
	}

	return nil
}

func TerminalSize(t *os.File) (int, int, error) {

	bi := _CONSOLE_SCREEN_BUFFER_INFO{}
	if err := _GetConsoleScreenBufferInfo(t, &bi); err != nil {
		fmt.Println(err)
		return -1, -1, err
	}
	return int(bi.size.x), int(bi.size.y), nil
}

func CursorNUp(t *os.File, n int) (err error) {

	bi := _CONSOLE_SCREEN_BUFFER_INFO{}
	if err := _GetConsoleScreenBufferInfo(t, &bi); err != nil {
		fmt.Println(err)
		return err
	}

	cp := bi.cursorPosition
	cp.y = cp.y - int16(n)

	ret, _, errno := syscall.Syscall(SetConsoleCursorPosition.Addr(),
		2,
		uintptr(t.Fd()),
		cp.uintptr(),
		0)

	if ret == 0 || errno != 0 {
		fmt.Println(errno)
		return errno
	}

	return
}
