//go:build windows

package ui

import (
	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

func supportsANSITerminal(fd int) bool {
	if !term.IsTerminal(fd) {
		return false
	}
	handle := windows.Handle(fd)
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return false
	}
	if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING != 0 {
		return true
	}
	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	if err := windows.SetConsoleMode(handle, mode); err != nil {
		return false
	}
	return true
}
