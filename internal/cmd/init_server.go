package cmd

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/madhukraft/nether/internal/config"
	"github.com/madhukraft/nether/internal/server"
	"github.com/madhukraft/nether/internal/ui"
)

func ensureServerInitialized() *config.Config {
	cfg, err := config.Load()
	if err == nil {
		return cfg
	}

	reader := bufio.NewReader(os.Stdin)
	isServer := server.IsServerDirectory(".")

	var msg string
	if isServer {
		msg = "This looks like a server but no nether.toml found. Initialize one?"
	} else {
		msg = "No server detected. Are you sure this is a server directory?"
	}

	ok, err := ui.Confirm(reader, os.Stdout, msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if !ok {
		os.Exit(0)
	}

	serverType, err := promptServerType(reader, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	version, err := promptVersion(reader, os.Stdout, serverType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	cfg = &config.Config{
		Type:     serverType,
		Version:  version,
		Created:  time.Now().UTC().Format(time.RFC3339),
		TargetOS: server.TargetOS(),
	}

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("nether.toml created. Continuing...")
	return cfg
}
