# wand

A magical CLI tool to connect to your servers via RDP and SSH with a flick of your wand! 🪄

## Features

- Instantly connect to RDP or SSH hosts defined in your `config.json`.
- Supports jump hosts for SSH.
- Easily manage multiple environments.
- Works seamlessly with zsh (and other shells).

## Installation

### Quick Install (Recommended)

Run the following in your terminal:

```zsh
curl -fsSL https://raw.githubusercontent.com/Joeri-Abbo/wand/stable/install.sh | bash
```

This will:
- Install Go (if needed, macOS/Homebrew only)
- Clone the repo to `~/wand`
- Build the binary
- Add it to your PATH in `~/.zshrc`

Restart your terminal or run `source ~/.zshrc` to use the `wand` command.

---

#### Manual Install

1. **Clone or Download** this repository.
2. **Add the wand tool to your PATH** (if not already):
   ```zsh
   export PATH="$PATH:/path/to/wand"
   ```
   (Replace `/path/to/wand` with the actual directory.)
3. **(Optional) Add to your zshrc** for convenience:
   ```zsh
   echo 'export PATH="$PATH:/path/to/wand"' >> ~/.zshrc
   source ~/.zshrc
   ```

## Configuration

Edit the `example.json` file to define your environments and hosts. Example:

```json
[
  {
    "servers": [
      {
        "connection": "rdp",
        "name": "winmachine001",
        "user": "jabbo",
        "host": "12.345.678.90"
      },
      {
        "connection": "ssh",
        "name": "linuxjump001",
        "host": "12.345.678.90",
        "user": "jabbo",
        "identityFile": ["~/.ssh/id_rsa"]
      },
      {
        "name": "secretvpn001",
        "uses_jumpHost": true,
        "host": "12.345.678.90",
        "jumpHost": ["linuxjump001"]
      }
    ]
  }
]
```

## Usage

### List available commands

- `wand` — interactive mode (pick group and machine)
- `wand [group]` — pick a machine from the group
- `wand [group] [machine]` — connect directly
- `wand edit` — edit your config file in your `$EDITOR`

### Example

```zsh
# Interactive (pick group and machine)
wand

# Pick machine from group servers
wand servers

# Connect directly to a machine
wand servers winmachine001

# Edit your config
wand edit
```

### Add a new host

Edit your config file and add your new host under the appropriate environment.

## Advanced

- **Jump Hosts:**
  - If a host uses a jump host, specify `uses_jumpHost` and `jumpHost` in the config.
- **Multiple Environments:**
  - Organize hosts by environment.
