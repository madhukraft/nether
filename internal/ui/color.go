package ui

import (
	"os"

	"golang.org/x/term"
)

var colorEnabled bool

func init() {
	_, noColor := os.LookupEnv("NO_COLOR")
	colorEnabled = term.IsTerminal(int(os.Stdout.Fd())) && !noColor
}

func Colorize(code, text string) string {
	if !colorEnabled {
		return text
	}
	return "\033[" + code + "m" + text + "\033[0m"
}

func Green(text string) string  { return Colorize("32", text) }
func Yellow(text string) string { return Colorize("33", text) }
func Red(text string) string    { return Colorize("31", text) }
func Blue(text string) string   { return Colorize("34", text) }
func Cyan(text string) string   { return Colorize("36", text) }
func Bold(text string) string   { return Colorize("1", text) }
func Dim(text string) string    { return Colorize("2", text) }
