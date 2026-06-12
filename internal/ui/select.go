package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

type key int

const (
	keyUnknown key = iota
	keyEnter
	keyUp
	keyDown
	keyCtrlC
)

func readKey() (key, error) {
	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return keyUnknown, err
	}
	if n == 0 {
		return keyUnknown, io.EOF
	}

	switch {
	case buf[0] == 13, buf[0] == 10:
		return keyEnter, nil
	case buf[0] == 27 && n >= 3 && buf[1] == 91:
		switch buf[2] {
		case 65:
			return keyUp, nil
		case 66:
			return keyDown, nil
		}
	case buf[0] == 3:
		return keyCtrlC, nil
	}

	return keyUnknown, nil
}

func Select(r io.Reader, w io.Writer, title string, options []string) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no options to choose from")
	}

	fd := int(os.Stdin.Fd())
	if supportsANSITerminal(fd) {
		return selectInteractive(w, fd, title, options)
	}
	return selectFallback(r, w, title, options)
}

func selectInteractive(w io.Writer, fd int, title string, options []string) (string, error) {
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", fmt.Errorf("failed to set raw terminal: %w", err)
	}
	defer term.Restore(fd, oldState)

	selected := 0
	lines := 1 + len(options)

	writeList(w, title, options, selected)

	for {
		k, err := readKey()
		if err != nil {
			return "", err
		}

		switch k {
		case keyUp:
			if selected > 0 {
				selected--
				redraw(w, title, options, selected, lines)
			}
		case keyDown:
			if selected < len(options)-1 {
				selected++
				redraw(w, title, options, selected, lines)
			}
		case keyEnter:
			clearBlock(w, lines)
			fmt.Fprintf(w, "%s %s\r\n", Bold(title), Cyan(options[selected]))
			return options[selected], nil
		case keyCtrlC:
			clearBlock(w, lines)
			return "", fmt.Errorf("cancelled")
		}
	}
}

func clearBlock(w io.Writer, lines int) {
	fmt.Fprintf(w, "\033[%dA\033[J", lines)
}

func redraw(w io.Writer, title string, options []string, selected int, lines int) {
	fmt.Fprintf(w, "\033[%dA\033[J", lines)
	writeList(w, title, options, selected)
}

func writeList(w io.Writer, title string, options []string, selected int) {
	fmt.Fprintf(w, "%s\r\n", Bold(title))
	for i, opt := range options {
		if i == selected {
			fmt.Fprintf(w, "\033[36m  > %s\033[0m\r\n", opt)
		} else {
			fmt.Fprintf(w, "    %s\r\n", opt)
		}
	}
}

func selectFallback(r io.Reader, w io.Writer, title string, options []string) (string, error) {
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}

	for {
		fmt.Fprintf(w, "%s\r\n", title)
		for i, opt := range options {
			fmt.Fprintf(w, "  %d) %s\r\n", i+1, opt)
		}
		fmt.Fprintf(w, "Enter number [1]: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}
		input = strings.TrimSpace(input)

		if input == "" {
			return options[0], nil
		}

		n, err := strconv.Atoi(input)
		if err != nil || n < 1 || n > len(options) {
			fmt.Fprintf(w, "Enter a number between 1 and %d.\r\n", len(options))
			continue
		}
		return options[n-1], nil
	}
}
