package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func Input(r io.Reader, w io.Writer, message, defaultVal string) (string, error) {
	if supportsANSITerminal(int(os.Stdin.Fd())) {
		return inputInteractive(w, message, defaultVal)
	}
	return inputFallback(r, w, message, defaultVal)
}

func inputInteractive(w io.Writer, message, defaultVal string) (string, error) {
	prompt := message
	if defaultVal != "" {
		prompt = fmt.Sprintf("%s [%s]", message, Dim(defaultVal))
	}
	fmt.Fprintf(w, "%s ", Bold(prompt))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

func inputFallback(r io.Reader, w io.Writer, message, defaultVal string) (string, error) {
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}

	prompt := message
	if defaultVal != "" {
		prompt = fmt.Sprintf("%s [%s]", message, defaultVal)
	}
	fmt.Fprintf(w, "%s ", prompt)

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

func Confirm(r io.Reader, w io.Writer, message string) (bool, error) {
	if supportsANSITerminal(int(os.Stdin.Fd())) {
		return confirmInteractive(w, message)
	}
	return confirmFallback(r, w, message)
}

func confirmInteractive(w io.Writer, message string) (bool, error) {
	fmt.Fprintf(w, "%s %s ", Bold(message), Dim("[y/N]"))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	input = strings.TrimSpace(strings.ToLower(input))

	return input == "y" || input == "yes", nil
}

func confirmFallback(r io.Reader, w io.Writer, message string) (bool, error) {
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}

	fmt.Fprintf(w, "%s [y/N]: ", message)

	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	input = strings.TrimSpace(strings.ToLower(input))

	return input == "y" || input == "yes", nil
}
