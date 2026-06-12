//go:build !windows

package ui

import "golang.org/x/term"

func supportsANSITerminal(fd int) bool {
	return term.IsTerminal(fd)
}
