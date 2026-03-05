package shell_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/MarcusXavierr/faaah/internal/shell"
)

func TestDetectFromPath(t *testing.T) {
	tests := []struct {
		name           string
		shellPath      string
		homeDir        string
		expectedName   string
		expectedConfig string
		expectedError  error
	}{
		{
			name:           "zsh from /usr/bin",
			shellPath:      "/usr/bin/zsh",
			homeDir:        "/home/user",
			expectedName:   "zsh",
			expectedConfig: "/home/user/.zshrc",
			expectedError:  nil,
		},
		{
			name:           "zsh from /bin",
			shellPath:      "/bin/zsh",
			homeDir:        "/home/user",
			expectedName:   "zsh",
			expectedConfig: "/home/user/.zshrc",
			expectedError:  nil,
		},
		{
			name:           "bash from /usr/bin",
			shellPath:      "/usr/bin/bash",
			homeDir:        "/home/user",
			expectedName:   "bash",
			expectedConfig: "/home/user/.bashrc",
			expectedError:  nil,
		},
		{
			name:           "bash from /bin",
			shellPath:      "/bin/bash",
			homeDir:        "/home/user",
			expectedName:   "bash",
			expectedConfig: "/home/user/.bashrc",
			expectedError:  nil,
		},
		{
			name:           "bash from /usr/local/bin",
			shellPath:      "/usr/local/bin/bash",
			homeDir:        "/home/user",
			expectedName:   "bash",
			expectedConfig: "/home/user/.bashrc",
			expectedError:  nil,
		},
		{
			name:           "fish (unsupported)",
			shellPath:      "/usr/bin/fish",
			homeDir:        "/home/user",
			expectedName:   "",
			expectedConfig: "",
			expectedError:  shell.ErrUnsupportedShell,
		},
		{
			name:           "empty SHELL",
			shellPath:      "",
			homeDir:        "/home/user",
			expectedName:   "",
			expectedConfig: "",
			expectedError:  shell.ErrShellNotSet,
		},
		{
			name:           "different home dir",
			shellPath:      "/usr/bin/zsh",
			homeDir:        "/root",
			expectedName:   "zsh",
			expectedConfig: "/root/.zshrc",
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := shell.DetectFromPath(tt.shellPath, tt.homeDir)

			if (err != nil && tt.expectedError == nil) || (err == nil && tt.expectedError != nil) || (err != nil && tt.expectedError != nil && !errors.Is(err, tt.expectedError)) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}

			if info.Name != tt.expectedName {
				t.Errorf("expected name %q, got %q", tt.expectedName, info.Name)
			}
			if info.ConfigPath != tt.expectedConfig {
				t.Errorf("expected config path %q, got %q", tt.expectedConfig, info.ConfigPath)
			}
		})
	}
}

func TestDetect(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("SHELL", "/usr/bin/zsh")
	t.Setenv("HOME", tmpHome)

	info, err := shell.Detect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedName := "zsh"
	expectedConfig := filepath.Join(tmpHome, ".zshrc")

	if info.Name != expectedName {
		t.Errorf("expected name %q, got %q", expectedName, info.Name)
	}
	if info.ConfigPath != expectedConfig {
		t.Errorf("expected config path %q, got %q", expectedConfig, info.ConfigPath)
	}
}
