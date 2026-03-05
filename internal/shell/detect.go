package shell

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrShellNotSet = errors.New("$SHELL is not set")
var ErrUnsupportedShell = errors.New("unsupported shell")

type ShellInfo struct {
	Name       string
	ConfigPath string
}

func Detect() (ShellInfo, error) {
	shellPath := os.Getenv("SHELL")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ShellInfo{}, fmt.Errorf("could not determine user home directory: %w", err)
	}
	return DetectFromPath(shellPath, homeDir)
}

func DetectFromPath(shellPath string, homeDir string) (ShellInfo, error) {
	if shellPath == "" {
		return ShellInfo{}, ErrShellNotSet
	}

	name := filepath.Base(shellPath)
	var configPath string

	switch name {
	case "zsh":
		configPath = filepath.Join(homeDir, ".zshrc")
	case "bash":
		configPath = filepath.Join(homeDir, ".bashrc")
	default:
		return ShellInfo{}, fmt.Errorf("%w: %s", ErrUnsupportedShell, name)
	}

	return ShellInfo{
		Name:       name,
		ConfigPath: configPath,
	}, nil
}
