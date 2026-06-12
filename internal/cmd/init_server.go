package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/madhukraft/nether/internal/config"
	"github.com/madhukraft/nether/internal/server"
)

func ensureServerInitialized() *config.Config {
	cfg, err := config.Load()
	if err == nil {
		return cfg
	}

	reader := bufio.NewReader(os.Stdin)
	isServer := server.IsServerDirectory(".")

	if isServer {
		fmt.Println("This looks like a Minecraft server but no nether.toml was found.")
		fmt.Print("Initialize one? (y/N): ")
	} else {
		fmt.Println("No Minecraft server detected in the current directory.")
		fmt.Print("Run 'nether create' first, or are you sure this is a server directory? (y/N): ")
	}

	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}

	response = strings.TrimSpace(response)
	if len(response) > 0 {
		response = response[:1]
	}

	if response != "y" && response != "Y" {
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
