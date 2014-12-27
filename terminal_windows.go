// +build windows
package main

// TODO: make this work somehow on Windows as well, maybe thru
// http://msdn.microsoft.com/en-us/library/ms683172.aspx

import "os"

func TerminalSize(t *os.File) (int, int, error) {
	return -1, -1, nil
}
