# sshm

A terminal UI for managing SSH connections stored in `~/.ssh/config`.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)

## Features

- Browse connections organised into collapsible groups
- Add, edit, and delete connections with full SSH field support
- Launch SSH sessions directly from the TUI — the app hands the terminal over cleanly
- Reads and writes `~/.ssh/config` directly — no separate database
- Real-time search that filters as you type
- Cross-platform: Linux, macOS, and Windows

## Installation

```bash
go install github.com/SChimera/sshm@latest
```

Or build from source:

```bash
git clone https://github.com/SChimera/sshm
cd sshm
go build -o sshm .
```

## Usage

```bash
sshm
```

## Keybindings

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate |
| `Enter` | Connect to host / collapse group |
| `Tab` | Switch between panels |
| `e` | Edit selected connection |
| `n` | New connection |
| `g` | New group |
| `d` | Delete selected (with confirmation) |
| `/` | Search |
| `Esc` | Cancel / close |
| `q` | Quit |

## Config format

`sshm` uses `# Group: <name>` comment markers in `~/.ssh/config` to organise connections into groups. Everything else is standard SSH config syntax, so the file remains usable by other tools.

```
# Group: Production
Host web-prod
    HostName 192.168.1.10
    User ubuntu
    IdentityFile ~/.ssh/id_ed25519

# Group: Dev
Host dev-box
    HostName 10.0.0.5
    User dev
```

## License

MIT
