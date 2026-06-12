package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/madhukraft/nether/internal/server"
	"github.com/madhukraft/nether/internal/ui"
	"github.com/spf13/cobra"
)

var serverType string
var serverVersion string
var ramFlag string
var minRAMFlag string
var portFlag int
var dirFlag string
var javaFlag int
var targetOSFlag string
var targetArchFlag string
var agreeEULA bool

func promptEula(in *bufio.Reader, out io.Writer) (bool, error) {
	return ui.Confirm(in, out, "Accept the Minecraft EULA (https://aka.ms/MinecraftEULA)?")
}

func parseRAM(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("RAM amount cannot be empty")
	}

	normalized := strings.ToLower(input)
	normalized = strings.TrimSuffix(normalized, "b")
	normalized = strings.TrimSuffix(normalized, "i")
	normalized = strings.TrimSpace(normalized)

	if len(normalized) == 0 {
		return "", fmt.Errorf("invalid RAM format")
	}

	suffix := normalized[len(normalized)-1:]
	numStr := normalized[:len(normalized)-1]

	if suffix == "g" {
		val, err := strconv.Atoi(strings.TrimSpace(numStr))
		if err != nil || val <= 0 {
			return "", fmt.Errorf("invalid numeric value: %s", numStr)
		}
		return fmt.Sprintf("%dM", val*1024), nil
	}

	if suffix == "m" {
		val, err := strconv.Atoi(strings.TrimSpace(numStr))
		if err != nil || val <= 0 {
			return "", fmt.Errorf("invalid numeric value: %s", numStr)
		}
		return fmt.Sprintf("%dM", val), nil
	}

	return "", fmt.Errorf("missing unit (must specify G or M, e.g. 2G or 2048M)")
}

func promptRAM(in *bufio.Reader, out io.Writer) (string, error) {
	for {
		input, err := ui.Input(in, out, "RAM to allocate (e.g. 2G or 2048M)", "")
		if err != nil {
			return "", err
		}
		ram, err := parseRAM(input)
		if err == nil {
			return ram, nil
		}
		fmt.Fprintf(out, "error: %v. Please try again.\n", err)
	}
}

func promptPort(in *bufio.Reader, out io.Writer) (int, error) {
	for {
		input, err := ui.Input(in, out, "Server port (press Enter for 25565)", "25565")
		if err != nil {
			return 0, err
		}
		input = strings.TrimSpace(input)
		if input == "" {
			return 25565, nil
		}
		port, err := strconv.Atoi(input)
		if err != nil {
			fmt.Fprintf(out, "error: invalid port. Please try again.\n")
			continue
		}
		if port < 1 || port > 65535 {
			fmt.Fprintf(out, "error: port must be between 1 and 65535. Please try again.\n")
			continue
		}
		return port, nil
	}
}

func promptVersion(in *bufio.Reader, out io.Writer, serverType string) (string, error) {
	for {
		input, err := ui.Input(in, out, "Minecraft version (e.g. 1.21.1)", "")
		if err != nil {
			return "", err
		}
		if input == "" {
			fmt.Fprintln(out, "Version cannot be empty.")
			continue
		}
		if err := server.ValidateVersion(serverType, input); err != nil {
			fmt.Fprintf(out, "error: %v. Try again.\n", err)
			continue
		}
		return input, nil
	}
}

func promptServerType(in *bufio.Reader, out io.Writer) (string, error) {
	types := []string{"paper", "vanilla", "fabric", "neoforge", "forge"}
	return ui.Select(in, out, "Select server type:", types)
}

func promptTargetOS(in *bufio.Reader, out io.Writer) (string, error) {
	return ui.Select(in, out, "What OS is your host running on?", []string{"linux", "macos", "windows"})
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Minecraft server in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		if !server.SetTarget(targetOSFlag, targetArchFlag) {
			fmt.Fprintf(os.Stderr, "error: invalid target OS %q (use linux, macos, or windows)\n", targetOSFlag)
			os.Exit(1)
		}

		reader := bufio.NewReader(os.Stdin)

		if !cmd.Flags().Changed("os") && server.IsRunningInDocker() {
			osChoice, err := promptTargetOS(reader, os.Stdout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			if !server.SetTarget(osChoice, targetArchFlag) {
				fmt.Fprintf(os.Stderr, "error: invalid target OS %q\n", osChoice)
				os.Exit(1)
			}
		}

		if !cmd.Flags().Changed("type") {
			t, err := promptServerType(reader, os.Stdout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading server type: %v\n", err)
				os.Exit(1)
			}
			serverType = t
		}

		if !cmd.Flags().Changed("version") {
			v, err := promptVersion(reader, os.Stdout, serverType)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading version: %v\n", err)
				os.Exit(1)
			}
			serverVersion = v
		}

		accepted := agreeEULA
		if !accepted {
			var err error
			accepted, err = promptEula(reader, os.Stdout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
				os.Exit(1)
			}
		}
		if !accepted {
			fmt.Println("EULA not accepted. Aborting.")
			os.Exit(1)
		}

		var maxRAM string
		if ramFlag != "" {
			r, err := parseRAM(ramFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: invalid RAM value %q: %v\n", ramFlag, err)
				os.Exit(1)
			}
			maxRAM = r
		} else {
			r, err := promptRAM(reader, os.Stdout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading RAM allocation: %v\n", err)
				os.Exit(1)
			}
			maxRAM = r
		}

		var minRAM string
		if minRAMFlag != "" {
			r, err := parseRAM(minRAMFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: invalid min RAM value %q: %v\n", minRAMFlag, err)
				os.Exit(1)
			}
			minRAM = r
		} else {
			minRAM = maxRAM
		}

		var port int
		if portFlag != 0 {
			port = portFlag
		} else {
			p, err := promptPort(reader, os.Stdout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading port: %v\n", err)
				os.Exit(1)
			}
			port = p
		}

		if dirFlag != "" {
			if dirFlag != "." {
				if err := os.MkdirAll(dirFlag, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "error creating directory: %v\n", err)
					os.Exit(1)
				}
				if err := os.Chdir(dirFlag); err != nil {
					fmt.Fprintf(os.Stderr, "error changing to directory: %v\n", err)
					os.Exit(1)
				}
			}
		} else {
			dir := fmt.Sprintf("%s-%s", serverType, serverVersion)
			for i := 2; ; i++ {
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					break
				}
				dir = fmt.Sprintf("%s-%s-%d", serverType, serverVersion, i)
			}
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "error creating directory: %v\n", err)
				os.Exit(1)
			}
			if err := os.Chdir(dir); err != nil {
				fmt.Fprintf(os.Stderr, "error changing to directory: %v\n", err)
				os.Exit(1)
			}
		}

		if _, err := os.Stat("nether.toml"); err == nil {
			fmt.Println("error: a nether server already exists in this directory")
			os.Exit(1)
		}

		javaVer := server.GetJavaVersionForMinecraft(serverVersion)
		if javaFlag != 0 {
			if javaFlag < 8 || javaFlag > 30 {
				fmt.Fprintf(os.Stderr, "error: Java version %d is outside reasonable range (8-30)\n", javaFlag)
				os.Exit(1)
			}
			javaVer = javaFlag
		}

		var installerFile string
		switch serverType {
		case "paper":
			if err := server.DownloadPaper(serverVersion); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		case "vanilla":
			if err := server.DownloadVanilla(serverVersion); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		case "fabric":
			if err := server.DownloadFabric(serverVersion); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		case "neoforge":
			fmt.Println("Fetching NeoForge version...")
			installer, err := server.DownloadNeoForge(serverVersion)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			installerFile = installer
		case "forge":
			fmt.Println("Fetching Forge version...")
			installer, err := server.DownloadForge(serverVersion)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			installerFile = installer
		default:
			fmt.Fprintf(os.Stderr, "error: unknown server type %q\n", serverType)
			os.Exit(1)
		}

		if err := server.DownloadJava(serverVersion, javaVer); err != nil {
			fmt.Fprintf(os.Stderr, "error setting up Java: %v\n", err)
			os.Exit(1)
		}

		if installerFile != "" {
			switch serverType {
			case "neoforge":
				if err := server.InstallNeoForge(installerFile); err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					os.Exit(1)
				}
			case "forge":
				if err := server.InstallForge(installerFile); err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					os.Exit(1)
				}
			}
		}

		if installerFile != "" {
			if err := server.WriteUserJVMArgs(minRAM, maxRAM); err != nil {
				fmt.Fprintf(os.Stderr, "error writing JVM args: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := server.WriteStartScript(minRAM, maxRAM); err != nil {
				fmt.Fprintf(os.Stderr, "error writing start script: %v\n", err)
				os.Exit(1)
			}
		}

		if err := server.WriteServerProperties(port); err != nil {
			fmt.Fprintf(os.Stderr, "error writing server.properties: %v\n", err)
			os.Exit(1)
		}

		if err := server.WriteEula(); err != nil {
			fmt.Fprintf(os.Stderr, "error writing eula.txt: %v\n", err)
			os.Exit(1)
		}

		if err := server.WriteConfig(serverType, serverVersion); err != nil {
			fmt.Fprintf(os.Stderr, "error writing nether.toml: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Done! Server ready.")
	},
}

func init() {
	createCmd.Flags().StringVarP(&serverType, "type", "t", "paper", "Server type (paper, vanilla, fabric, neoforge, forge)")
	createCmd.Flags().StringVarP(&serverVersion, "version", "v", "", "Minecraft version (e.g. 1.21.1)")
	createCmd.Flags().StringVar(&ramFlag, "ram", "", "Amount of RAM for the server (e.g. 2G or 2048M)")
	createCmd.Flags().StringVar(&minRAMFlag, "min-ram", "", "Minimum RAM for the server (defaults to same as max if not set)")
	createCmd.Flags().IntVar(&portFlag, "port", 0, "Server port (default 25565)")
	createCmd.Flags().StringVar(&dirFlag, "dir", "", "Directory to create the server in (use '.' for current directory)")
	createCmd.Flags().IntVar(&javaFlag, "java", 0, "Java major version (auto-detected if not set)")
	createCmd.Flags().StringVar(&targetOSFlag, "os", "", "Target OS for Java and scripts (linux, macos, windows)")
	createCmd.Flags().StringVar(&targetArchFlag, "arch", "", "Target architecture (amd64, arm64; defaults to host)")
	createCmd.Flags().BoolVar(&agreeEULA, "agree-to-eula", false, "Agree to the Minecraft EULA automatically")
	rootCmd.AddCommand(createCmd)
}
