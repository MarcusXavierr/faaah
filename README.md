# Faaah

> A self-contained Go CLI that hooks into your shell and plays a sound every time a command fails.

https://github.com/user-attachments/assets/b9ff7cb1-72bf-4076-a561-e0e3ecedd9a3

When working in a terminal, failed commands often go unnoticed — especially during long builds, test runs, or multi-step pipelines. `faaah` solves this by playing an audible "faaah" sound every time a shell command exits with a non-zero status code.

You install the binary once, run `faaah install`, and from that point on, every failed command triggers the sound automatically. No manual setup, no external audio players, and no config files.

## Installation

1. Build or install the `faaah` binary (requires Go 1.24+ if building from source):
   ```bash
   git clone https://github.com/MarcusXavierr/faaah.git
   cd faaah
   go build -o faaah ./cmd/faaah/
   
   # Move the binary to a directory in your $PATH
   sudo mv faaah /usr/local/bin/
   ```

2. Install the shell hook:
   ```bash
   faaah install
   ```

3. Restart your shell or reload your config:
   ```bash
   source ~/.zshrc  # or ~/.bashrc if using bash
   ```

## Usage

The Faaah binary exposes 4 simple subcommands with no flags or environment variables required:

- `faaah install` — Detects your shell (`$SHELL`), finds your config file, and appends the trap hook. It uses absolute paths so it continues to work even if you move things around, and it's perfectly safe to run multiple times (idempotent).
- `faaah uninstall` — Cleanly removes the trap hook from your shell config file.
- `faaah play` — Decodes and plays the embedded mp3 Faaah sound. Faaah runs very fast and low latency.
- `faaah status` — Shows whether the Faaah hook is currently installed and for which shell.

## Supported Shells

Faaah currently supports:
- **Zsh** (`~/.zshrc`)
- **Bash** (`~/.bashrc`)

## How It Works

When Faaah installs itself, it adds a sentinel block to the bottom of your shell configuration file using `trap`:

```bash
# >>> faaah hook >>>
trap 'if [ -n "$ZSH_VERSION" ] && [ "$ZSH_EVAL_CONTEXT" != "toplevel:trap" ] && [ "$ZSH_EVAL_CONTEXT" != "cmdarg:trap" ]; then :; else (/absolute/path/to/faaah play >/dev/null 2>&1 &); fi' ERR
# <<< faaah hook <<<
```

The Faaah binary is 100% self-contained. At compile time, Faaah embeds a 47KB sound file directly into the binary utilizing Go's `//go:embed` directive. Faaah is not dependent on `afplay`, `aplay`, or any other media packages installed on your system.

## Development

This project was built following a strict phase-by-phase Test-Driven Development (TDD) methodology.

- **Go version:** 1.24.0
- **Audio Output:** Uses `github.com/hajimehoshi/oto/v2` which communicates directly with ALSA on Linux.
- **MP3 Decoding:** Uses `github.com/hajimehoshi/go-mp3` for pure-Go decoding without CGo.
- **CLI Framework:** Built utilizing `github.com/spf13/cobra`.

### Local Setup & Testing

Make sure you are running Go 1.24, and then you can build and test the project:

```bash
# Run tests
go test ./...

# Run tests with race detection
go test -race ./...

# Build locally
go build -o faaah ./cmd/faaah/
```
