# Faaah — Project Specification

> A self-contained Go CLI that hooks into your shell and plays a sound every time a command fails.

---

## 1. Problem Statement

When working in a terminal, failed commands often go unnoticed — especially during long builds, test runs, or multi-step pipelines. `faaah` solves this by playing an audible "faaah" sound every time a shell command exits with a non-zero status code.

The user installs the binary once, runs `faaah install`, and from that point on, every failed command triggers the sound automatically. No manual setup, no external audio players, no config files.

---

## 2. Project Metadata

| Field | Value |
|---|---|
| **Go module** | `github.com/MarcusXavierr/faaah` |
| **Go version** | `1.24.0` |
| **Binary name** | `faaah` |
| **Project root** | `/home/marcus/Projects/faah` |
| **Sound file** | `assets/faaah.mp3` (47KB, embedded at compile time) |
| **Release** | GoReleaser (future — not in scope for this spec) |
| **Development approach** | TDD (Red → Green → Refactor) |

---

## 3. CLI Interface

The binary exposes exactly 4 subcommands. No flags, no config files, no environment variables.

```
faaah install     # Detect shell, add trap hook to config file (idempotent)
faaah uninstall   # Remove trap hook from config file
faaah play        # Play the embedded faaah.mp3 sound
faaah status      # Show whether the hook is installed and for which shell
```

### 3.1 `faaah install`

1. Detects the user's shell by reading `$SHELL`
2. Determines the config file path (`~/.bashrc` for bash, `~/.zshrc` for zsh)
3. Resolves the absolute path of the `faaah` binary itself (via `os.Executable()`)
4. Checks if the hook markers already exist in the config file
5. If not present: appends the hook block to the end of the config file
6. If already present: prints a message and exits successfully (idempotent)

**Success output:**
```
✓ faaah hook installed in ~/.zshrc
  Restart your shell or run: source ~/.zshrc
```

**Already installed output:**
```
✓ faaah hook is already installed in ~/.zshrc
```

### 3.2 `faaah uninstall`

1. Detects the user's shell (same as install)
2. Reads the config file
3. Removes everything between the start and end markers (inclusive)
4. Writes the file back
5. If no markers found: prints a message and exits successfully

**Success output:**
```
✓ faaah hook removed from ~/.zshrc
  Restart your shell or run: source ~/.zshrc
```

**Not installed output:**
```
✓ faaah hook is not installed in ~/.zshrc (nothing to do)
```

### 3.3 `faaah play`

1. Decodes the embedded mp3 bytes using `go-mp3`
2. Opens an audio context using `oto`
3. Plays the sound to completion
4. Exits

This is the command that the shell hook calls on every command failure. It must:
- Start playing as fast as possible (low latency)
- Not block the shell longer than the sound duration
- Exit cleanly after playback

### 3.4 `faaah status`

1. Detects the user's shell
2. Checks if the hook markers exist in the config file
3. Prints the result

**Installed output:**
```
✓ faaah hook is installed in ~/.zshrc
```

**Not installed output:**
```
✗ faaah hook is not installed in ~/.zshrc
```

---

## 4. Shell Hook Details

### 4.1 Hook Block Format

The hook is wrapped in sentinel markers to enable idempotent install and clean uninstall:

```bash
# >>> faaah hook >>>
trap '/absolute/path/to/faaah play' ERR
# <<< faaah hook <<<
```

**Key details:**
- The `trap` line uses the **absolute path** to the `faaah` binary, not a relative path or bare command name. This ensures it works even if `faaah` is not in `$PATH`.
- The absolute path is resolved at install time via `os.Executable()` + `filepath.EvalSymlinks()`.
- The markers are unique strings unlikely to appear naturally in any shell config.

### 4.2 How `trap ERR` Works

- `trap '...' ERR` runs the given command whenever **any** simple command exits with a non-zero status
- Commands inside `if`, `while`, or after `||`/`&&` do **not** trigger the trap (these are "expected" failures)
- This works identically in both bash and zsh
- The trap does **not** interfere with `set -e` or `pipefail`

### 4.3 Supported Shells

| Shell | Detection | Config File |
|---|---|---|
| `zsh` | `$SHELL` ends with `/zsh` | `~/.zshrc` |
| `bash` | `$SHELL` ends with `/bash` | `~/.bashrc` |
| Anything else | Error | — |

Detection uses `filepath.Base(os.Getenv("SHELL"))` to extract the shell name from the path, regardless of the prefix (e.g., `/usr/bin/zsh`, `/bin/zsh`, `/usr/local/bin/zsh` all resolve to `zsh`).

---

## 5. Project Structure

Following the [golang-standards/project-layout](https://github.com/golang-standards/project-layout):

```
faah/
├── assets/
│   ├── embed.go                 # go:embed directive, exports SoundFile []byte
│   └── faaah.mp3                # The sound file (embedded at compile time)
├── cmd/
│   └── faaah/
│       ├── main.go              # CLI entrypoint — initializes cobra root command
│       ├── root.go              # Root command definition and version info
│       ├── install.go           # "install" subcommand
│       ├── uninstall.go         # "uninstall" subcommand
│       ├── play.go              # "play" subcommand
│       └── status.go            # "status" subcommand
├── internal/
│   ├── hook/
│   │   ├── hook.go              # Install / Uninstall / Status for shell config files
│   │   └── hook_test.go         # Table-driven tests using temp files
│   ├── player/
│   │   ├── player.go            # Decode embedded mp3 and play via oto
│   │   └── player_test.go       # Tests mp3 decoding (not actual audio output)
│   └── shell/
│       ├── detect.go            # Detect shell type and config path from $SHELL
│       └── detect_test.go       # Table-driven tests with various $SHELL values
├── go.mod
├── go.sum
├── Makefile
└── project.spec.md              # This file
```

### 5.1 Why This Layout

- **`cmd/faaah/main.go`** — Per golang-standards, each binary gets its own directory under `cmd/`. The `main.go` only calls `Execute()` on the root cobra command. Each subcommand lives in its own file (`install.go`, `uninstall.go`, `play.go`, `status.go`) and registers itself via `init()`. Zero business logic — all subcommands delegate to internal packages.
- **`internal/`** — All business logic lives here. The Go compiler enforces that `internal/` packages cannot be imported by external projects. This is perfect since all our logic is private to the binary.
- **`internal/shell/`** — Shell detection is isolated so it can be tested independently with mocked `$SHELL` values.
- **`internal/hook/`** — File I/O for the hook is isolated so it can be tested with temp files, no risk of touching real config files.
- **`internal/player/`** — Audio playback is isolated. Takes `[]byte` as input (not the global embed), making it testable and decoupled from the `assets` package.
- **`assets/`** — Houses the mp3 and the `go:embed` directive. This is the only package that knows about the file system embedding.

---

## 6. Package Specifications

### 6.1 `assets` — Embedded Sound File

```go
// assets/embed.go
package assets

import _ "embed"

//go:embed faaah.mp3
var SoundFile []byte
```

This is the entire package. `SoundFile` is a `[]byte` containing the raw mp3 data, compiled into the binary. Other packages reference `assets.SoundFile` to get the audio data.

---

### 6.2 `internal/shell` — Shell Detection

**Types:**

```go
package shell

// ShellInfo holds information about the detected shell.
type ShellInfo struct {
    Name       string // "bash" or "zsh"
    ConfigPath string // absolute path, e.g. "/home/marcus/.zshrc"
}
```

**Functions:**

```go
// Detect reads $SHELL and returns the shell info.
// This is the public API used by cmd/faaah.
func Detect() (ShellInfo, error)

// DetectFromPath is the pure, testable core.
// It takes a shell path (e.g. "/usr/bin/zsh") and a home dir,
// and returns the shell info.
// Exported for testing purposes.
func DetectFromPath(shellPath string, homeDir string) (ShellInfo, error)
```

**Logic in `DetectFromPath`:**

1. If `shellPath` is empty → return error `"$SHELL is not set"`
2. Extract base name via `filepath.Base(shellPath)` → e.g. `"zsh"`, `"bash"`
3. Match against known shells:
   - `"zsh"` → config = `filepath.Join(homeDir, ".zshrc")`
   - `"bash"` → config = `filepath.Join(homeDir, ".bashrc")`
   - anything else → return error `"unsupported shell: <name>"`
4. Return `ShellInfo{Name: name, ConfigPath: configPath}`

**Logic in `Detect`:**

1. Read `os.Getenv("SHELL")` → pass to `DetectFromPath`
2. Read `os.UserHomeDir()` → pass to `DetectFromPath`

**Test cases (table-driven):**

| Test Name | `shellPath` | `homeDir` | Expected Name | Expected Config | Expected Error |
|---|---|---|---|---|---|
| `zsh from /usr/bin` | `/usr/bin/zsh` | `/home/user` | `zsh` | `/home/user/.zshrc` | nil |
| `zsh from /bin` | `/bin/zsh` | `/home/user` | `zsh` | `/home/user/.zshrc` | nil |
| `bash from /usr/bin` | `/usr/bin/bash` | `/home/user` | `bash` | `/home/user/.bashrc` | nil |
| `bash from /bin` | `/bin/bash` | `/home/user` | `bash` | `/home/user/.bashrc` | nil |
| `bash from /usr/local/bin` | `/usr/local/bin/bash` | `/home/user` | `bash` | `/home/user/.bashrc` | nil |
| `fish (unsupported)` | `/usr/bin/fish` | `/home/user` | — | — | `unsupported shell` |
| `empty SHELL` | `""` | `/home/user` | — | — | `$SHELL is not set` |
| `different home dir` | `/usr/bin/zsh` | `/root` | `zsh` | `/root/.zshrc` | nil |

---

### 6.3 `internal/hook` — Hook Management

**Constants:**

```go
package hook

const (
    MarkerStart = "# >>> faaah hook >>>"
    MarkerEnd   = "# <<< faaah hook <<<"
)
```

**Functions:**

```go
// Install appends the faaah hook to configPath.
// binaryPath is the absolute path to the faaah binary (used in the trap command).
// Idempotent: if hook markers already exist, returns nil without modifying the file.
func Install(configPath string, binaryPath string) error

// Uninstall removes the faaah hook block from configPath.
// If no markers are found, returns nil without modifying the file.
func Uninstall(configPath string) error

// Status checks whether the faaah hook markers exist in configPath.
// Returns true if both start and end markers are found.
func Status(configPath string) (bool, error)
```

**`Install` detailed logic:**

1. Read entire file content as string (if file doesn't exist, treat as empty string — this handles the case where `~/.bashrc` doesn't exist yet)
2. If content contains `MarkerStart` → return nil (already installed)
3. Construct hook block:
   ```
   \n# >>> faaah hook >>>\ntrap '<binaryPath> play' ERR\n# <<< faaah hook <<<\n
   ```
4. Append hook block to content
5. Write content back to file (create if doesn't exist, permissions `0644`)

**`Uninstall` detailed logic:**

1. Read entire file content as string
2. If content does NOT contain `MarkerStart` → return nil (nothing to remove)
3. Find index of `MarkerStart` line and `MarkerEnd` line
4. Remove all lines from `MarkerStart` line through `MarkerEnd` line (inclusive)
5. Also remove any trailing blank line left by the removal (keep file clean)
6. Write content back to file

**`Status` detailed logic:**

1. Read entire file content as string (if file doesn't exist, return `false, nil`)
2. Return `strings.Contains(content, MarkerStart)`, nil

**Test cases (all use `t.TempDir()` for isolated temp files):**

| Test Name | Initial File Content | Operation | Expected File Content | Expected Return |
|---|---|---|---|---|
| `Install on empty file` | `""` (new file) | `Install(path, "/usr/bin/faaah")` | Hook block only | `nil` |
| `Install on existing content` | `export FOO=bar\n` | `Install(path, "/usr/bin/faaah")` | Original + hook block | `nil` |
| `Install idempotent` | Already has hook block | `Install(path, "/usr/bin/faaah")` | Unchanged | `nil` |
| `Install creates parent dirs if needed` | File doesn't exist | `Install(path, "/usr/bin/faaah")` | Hook block only | `nil` |
| `Uninstall removes hook` | Content with hook in middle | `Uninstall(path)` | Content without hook | `nil` |
| `Uninstall hook at end` | Content with hook at end | `Uninstall(path)` | Content without hook | `nil` |
| `Uninstall no hook present` | `export FOO=bar\n` | `Uninstall(path)` | Unchanged | `nil` |
| `Uninstall empty file` | `""` | `Uninstall(path)` | Unchanged | `nil` |
| `Status when installed` | Content with hook | `Status(path)` | — | `true, nil` |
| `Status when not installed` | `export FOO=bar\n` | `Status(path)` | — | `false, nil` |
| `Status on non-existent file` | File doesn't exist | `Status(path)` | — | `false, nil` |

**Hook block format for assertions:**

When testing `Install`, the expected content appended should be exactly:
```
\n# >>> faaah hook >>>
trap '/usr/bin/faaah play' ERR
# <<< faaah hook <<<
```

Note the leading `\n` that ensures the hook doesn't end up on the same line as existing content.

---

### 6.4 `internal/player` — Audio Playback

**Functions:**

```go
package player

// Play decodes the given mp3 data and plays it through the default audio output.
// Blocks until playback is complete.
// Returns an error if the data cannot be decoded or if audio output fails.
func Play(soundData []byte) error
```

**`Play` detailed logic:**

1. Create a `bytes.Reader` from `soundData`
2. Create an mp3 decoder: `mp3.NewDecoder(reader)` — this returns a decoded PCM stream
3. Create an oto context with:
   - Sample rate: from `decoder.SampleRate()`
   - Channels: 2 (stereo — mp3 standard)
   - Bit depth: 2 bytes (16-bit — mp3 standard)
4. Create a player from the context
5. Write decoded PCM data to the player (or use `io.Copy`)
6. Wait for playback to finish
7. Close player and context

**Important oto/go-mp3 details:**

- `oto.NewContext` should only be created once per process. Since `faaah play` runs as a short-lived process (one invocation per failed command), this is fine — we create it, play, and exit.
- The oto context uses ALSA on Linux. No PulseAudio or PipeWire wrapper needed — ALSA works on virtually all Linux systems.
- `go-mp3` decodes to raw PCM (signed 16-bit, little-endian, stereo). This is exactly what oto expects.

**Test cases:**

| Test Name | Input | Expected |
|---|---|---|
| `Decode valid mp3` | The actual embedded `assets.SoundFile` bytes | `mp3.NewDecoder` returns no error, `decoder.SampleRate()` > 0, `decoder.Length()` > 0 |
| `Decode invalid data` | `[]byte("not an mp3")` | `mp3.NewDecoder` returns an error |
| `Decode empty data` | `[]byte{}` | Returns an error |

> **Note:** We do NOT test actual audio output in automated tests (requires audio hardware and would be flaky). We test that the mp3 decodes correctly. Actual playback is verified manually.

---

### 6.5 `cmd/faaah` — CLI Entrypoint (Cobra)

The CLI uses [cobra](https://github.com/spf13/cobra) for subcommand management. Each subcommand lives in its own file and registers itself in `init()`.

#### `cmd/faaah/main.go`

```go
package main

import "os"

func main() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

#### `cmd/faaah/root.go`

```go
package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
    Use:   "faaah",
    Short: "Play a sound every time a shell command fails",
    Long:  "faaah hooks into your shell (bash/zsh) and plays a sound\nevery time a command exits with a non-zero status code.",
}
```

#### `cmd/faaah/install.go`

```go
package main

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
    "github.com/MarcusXavierr/faaah/internal/hook"
    "github.com/MarcusXavierr/faaah/internal/shell"
)

var installCmd = &cobra.Command{
    Use:   "install",
    Short: "Install the faaah hook into your shell config",
    RunE: func(cmd *cobra.Command, args []string) error {
        info, err := shell.Detect()
        if err != nil {
            return err
        }

        exe, err := os.Executable()
        if err != nil {
            return fmt.Errorf("could not determine executable path: %w", err)
        }
        exe, err = filepath.EvalSymlinks(exe)
        if err != nil {
            return fmt.Errorf("could not resolve symlinks: %w", err)
        }

        if err := hook.Install(info.ConfigPath, exe); err != nil {
            return err
        }

        fmt.Printf("✓ faaah hook installed in %s\n", info.ConfigPath)
        fmt.Printf("  Restart your shell or run: source %s\n", info.ConfigPath)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(installCmd)
}
```

#### `cmd/faaah/uninstall.go`

```go
var uninstallCmd = &cobra.Command{
    Use:   "uninstall",
    Short: "Remove the faaah hook from your shell config",
    RunE: func(cmd *cobra.Command, args []string) error {
        info, err := shell.Detect()
        if err != nil {
            return err
        }
        if err := hook.Uninstall(info.ConfigPath); err != nil {
            return err
        }
        fmt.Printf("✓ faaah hook removed from %s\n", info.ConfigPath)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(uninstallCmd)
}
```

#### `cmd/faaah/play.go`

```go
var playCmd = &cobra.Command{
    Use:   "play",
    Short: "Play the faaah sound",
    RunE: func(cmd *cobra.Command, args []string) error {
        return player.Play(assets.SoundFile)
    },
}

func init() {
    rootCmd.AddCommand(playCmd)
}
```

#### `cmd/faaah/status.go`

```go
var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "Check if the faaah hook is installed",
    RunE: func(cmd *cobra.Command, args []string) error {
        info, err := shell.Detect()
        if err != nil {
            return err
        }
        installed, err := hook.Status(info.ConfigPath)
        if err != nil {
            return err
        }
        if installed {
            fmt.Printf("✓ faaah hook is installed in %s\n", info.ConfigPath)
        } else {
            fmt.Printf("✗ faaah hook is not installed in %s\n", info.ConfigPath)
        }
        return nil
    },
}

func init() {
    rootCmd.AddCommand(statusCmd)
}
```

**Cobra gives us for free:**
- Auto-generated `--help` for root and all subcommands
- Auto-generated usage/error messages for unknown subcommands
- Consistent error handling via `RunE` (returns error, cobra prints it)
- Future extensibility (flags, persistent flags, etc.) if needed

---

## 7. Dependencies

| Dependency | Version | Purpose | Why This One |
|---|---|---|---|
| `github.com/hajimehoshi/go-mp3` | latest | Pure Go mp3 decoder | No CGo, no external libs. Decodes to raw PCM. Well-maintained (same author as Ebitengine). |
| `github.com/hajimehoshi/oto/v2` | v2 | Low-level audio output | Cross-platform (Linux/macOS/Windows). Uses ALSA on Linux via syscalls — no CGo needed. Same author as go-mp3 so they're designed to work together. |
| `github.com/spf13/cobra` | latest | CLI framework | Industry standard for Go CLIs. Provides subcommand routing, auto-generated help, consistent error handling. Used by kubectl, Hugo, GitHub CLI, etc. |

**Standard library packages used (no install needed):**
- `os`, `filepath` — file I/O and path handling
- `strings`, `bytes` — string manipulation
- `fmt` — output formatting
- `embed` — compile-time file embedding
- `testing` — test framework
- `io` — stream interfaces

---

## 8. TDD Implementation Order

The project is built phase-by-phase, each following Red → Green → Refactor:

### Phase 1: Project Scaffolding (no tests — infrastructure only)

1. Create all directories: `cmd/faaah/`, `internal/shell/`, `internal/hook/`, `internal/player/`, `assets/`
2. Move `faaah.mp3` from project root to `assets/faaah.mp3`
3. Create `assets/embed.go` with the `go:embed` directive
4. Run `go get github.com/hajimehoshi/go-mp3`, `go get github.com/hajimehoshi/oto/v2`, and `go get github.com/spf13/cobra`
5. Create `cmd/faaah/main.go` and `cmd/faaah/root.go` with the cobra root command
6. Create a `Makefile` with `build`, `test`, and `clean` targets
7. Verify: `go build ./...` compiles without errors

### Phase 2: Shell Detection — TDD (`internal/shell`)

1. **RED:** Write `internal/shell/detect_test.go` with all test cases from section 6.2
2. **GREEN:** Implement `internal/shell/detect.go` to make all tests pass
3. **REFACTOR:** Clean up if needed
4. Verify: `go test -v ./internal/shell/`

### Phase 3: Hook Management — TDD (`internal/hook`)

1. **RED:** Write `internal/hook/hook_test.go` with all test cases from section 6.3
2. **GREEN:** Implement `internal/hook/hook.go` to make all tests pass
3. **REFACTOR:** Clean up if needed
4. Verify: `go test -v ./internal/hook/`

### Phase 4: Audio Playback — TDD (`internal/player`)

1. **RED:** Write `internal/player/player_test.go` with decode test cases from section 6.4
2. **GREEN:** Implement `internal/player/player.go` to make all tests pass
3. **REFACTOR:** Clean up if needed
4. Verify: `go test -v ./internal/player/`

### Phase 5: CLI Wiring (`cmd/faaah`)

1. Implement cobra subcommand files: `install.go`, `uninstall.go`, `play.go`, `status.go` as described in section 6.5
2. Run `go build -o faaah ./cmd/faaah/` to produce the binary
3. Verify: full test suite `go test ./...` passes

---

## 9. Verification Plan

### 9.1 Automated Tests

```bash
# Run all tests in all packages
go test ./...

# Run with verbose output (see individual test names)
go test -v ./...

# Run a single package
go test -v ./internal/shell/
go test -v ./internal/hook/
go test -v ./internal/player/

# Run with race detection
go test -race ./...
```

### 9.2 Manual End-to-End Verification

After all automated tests pass, run these manual tests in order:

**Step 1 — Build the binary:**
```bash
cd /home/marcus/Projects/faah
go build -o faaah ./cmd/faaah/
```

**Step 2 — Test `play` (audio works):**
```bash
./faaah play
# ✅ Expected: you hear the "faaah" sound from your speakers
```

**Step 3 — Test `status` before install:**
```bash
./faaah status
# ✅ Expected: "✗ faaah hook is not installed in ~/.zshrc"
```

**Step 4 — Test `install`:**
```bash
./faaah install
# ✅ Expected: "✓ faaah hook installed in ~/.zshrc"

# Verify the hook was added:
tail -5 ~/.zshrc
# ✅ Expected: you see the ">>> faaah hook >>>" block with the trap command
```

**Step 5 — Test idempotent install:**
```bash
./faaah install
# ✅ Expected: "✓ faaah hook is already installed in ~/.zshrc"

# Verify no duplicate:
grep -c "faaah hook" ~/.zshrc
# ✅ Expected: "2" (one start marker, one end marker)
```

**Step 6 — Test `status` after install:**
```bash
./faaah status
# ✅ Expected: "✓ faaah hook is installed in ~/.zshrc"
```

**Step 7 — Test the actual hook (the money test!):**
```bash
source ~/.zshrc    # reload config
false              # run a command that always fails
# ✅ Expected: you hear the "faaah" sound

ls /nonexistent    # another failing command
# ✅ Expected: you hear the "faaah" sound again

echo "hello"       # a passing command
# ✅ Expected: no sound

if false; then echo yes; fi   # failure inside a conditional
# ✅ Expected: no sound (trap doesn't fire for expected failures)
```

**Step 8 — Test `uninstall`:**
```bash
./faaah uninstall
# ✅ Expected: "✓ faaah hook removed from ~/.zshrc"

# Verify the hook was removed:
grep "faaah hook" ~/.zshrc
# ✅ Expected: no output (hook is gone)
```

**Step 9 — Test no-args and bad args:**
```bash
./faaah
# ✅ Expected: usage message, exit code 1

./faaah badcommand
# ✅ Expected: usage message, exit code 1
```
