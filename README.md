# Nether

Minecraft server setup tool.

## Install

Download the latest binary for your OS from [GitHub Releases](https://github.com/madhukraft/nether/releases).

Or with Go:
```
go install github.com/madhukraft/nether/cmd/nether@latest
```

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

Server types: `paper`, `vanilla`, `fabric`, `neoforge`, `forge`

## How it works

Nether downloads the server jar and provisions its own bundled Java, fully isolated from your system. It then generates run scripts, `server.properties`, `eula.txt`, and JVM args in a self-contained directory. Nothing is installed globally and your system's Java installation is never touched.
