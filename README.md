# Nether

Minecraft server setup tool. Downloads the server jar and provisions its own bundled Java, fully isolated from your system.

## Install

**Linux / macOS:**
```sh
curl -fsSL https://raw.githubusercontent.com/madhukraft/nether/main/install.sh | sh
```

**Linux / macOS (Homebrew):**
```sh
brew install madhukraft/tap/nether
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/madhukraft/nether/main/install.ps1 | iex
```

**Windows (Scoop):**
```powershell
scoop bucket add madhukraft https://github.com/madhukraft/scoop-bucket
scoop install madhukraft/nether
```

Or download the binary for your OS from [GitHub Releases](https://github.com/madhukraft/nether/releases) (Linux, macOS, Windows, amd64 and arm64).

Or with Docker:
```
docker pull ghcr.io/madhukraft/nether
```

## Usage

Create a server interactively:
```
nether create
```

All flags at once:
```
nether create --type paper --version 1.21.1 --ram 4G --port 25565
```

With Docker (creates files in current directory):
```
docker run --rm -it -v .:/data ghcr.io/madhukraft/nether create
```

Server types: `paper`, `vanilla`, `fabric`, `neoforge`, `forge`

Manage mods from Modrinth:
```
nether mods search <query>              # search for mods
nether mods install <slug|url>          # install a mod with dependency resolution
nether mods remove <name>               # remove an installed mod
nether mods list                        # list installed mods
```

Install modpacks:
```
nether modpack install <slug|url>       # install a modpack with all dependencies
nether modpack list                     # list installed modpacks
```

## How it works

Nether downloads the server jar and provisions its own bundled Java, fully isolated from your system. It then generates run scripts, `server.properties`, `eula.txt`, and JVM args in a self-contained directory. Nothing is installed globally and your system's Java installation is never touched.
