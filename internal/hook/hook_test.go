package hook_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MarcusXavierr/faaah/internal/hook"
)

const (
	testBinaryPath = "/usr/bin/faaah"
	expectedHook   = "\n# >>> faaah hook >>>\ntrap '/usr/bin/faaah play' ERR\n# <<< faaah hook <<<\n"
)

func TestInstall(t *testing.T) {
	tests := []struct {
		name          string
		initial       string
		createFile    bool
		subDir        bool
		expectedError bool
		expectedFile  string
	}{
		{
			name:          "Install on empty file",
			initial:       "",
			createFile:    true,
			expectedError: false,
			expectedFile:  expectedHook,
		},
		{
			name:          "Install on existing content",
			initial:       "export FOO=bar\n",
			createFile:    true,
			expectedError: false,
			expectedFile:  "export FOO=bar\n" + expectedHook,
		},
		{
			name:          "Install idempotent",
			initial:       "export FOO=bar\n" + expectedHook,
			createFile:    true,
			expectedError: false,
			expectedFile:  "export FOO=bar\n" + expectedHook,
		},
		{
			name:          "Install creates parent dirs if needed",
			initial:       "",
			createFile:    false,
			subDir:        true,
			expectedError: false,
			expectedFile:  expectedHook,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, ".zshrc")

			if tt.subDir {
				configPath = filepath.Join(tempDir, "subdir", ".zshrc")
			}

			if tt.createFile {
				err := os.WriteFile(configPath, []byte(tt.initial), 0644)
				if err != nil {
					t.Fatalf("Failed to create initial file: %v", err)
				}
			}

			err := hook.Install(configPath, testBinaryPath)
			if (err != nil) != tt.expectedError {
				t.Fatalf("Install() error = %v, wantErr %v", err, tt.expectedError)
			}

			content, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("Failed to read config file after Install: %v", err)
			}

			if string(content) != tt.expectedFile {
				t.Errorf("Install() modified file content mismatch:\nGot:\n%s\nWant:\n%s", string(content), tt.expectedFile)
			}
		})
	}
}

func TestUninstall(t *testing.T) {
	tests := []struct {
		name          string
		initial       string
		expectedError bool
		expectedFile  string
	}{
		{
			name:          "Uninstall removes hook",
			initial:       "export FOO=bar\n" + expectedHook + "export BAZ=qux\n",
			expectedError: false,
			expectedFile:  "export FOO=bar\nexport BAZ=qux\n",
		},
		{
			name:          "Uninstall hook at end",
			initial:       "export FOO=bar\n" + expectedHook,
			expectedError: false,
			expectedFile:  "export FOO=bar\n",
		},
		{
			name:          "Uninstall no hook present",
			initial:       "export FOO=bar\n",
			expectedError: false,
			expectedFile:  "export FOO=bar\n",
		},
		{
			name:          "Uninstall empty file",
			initial:       "",
			expectedError: false,
			expectedFile:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, ".zshrc")

			err := os.WriteFile(configPath, []byte(tt.initial), 0644)
			if err != nil {
				t.Fatalf("Failed to create initial file: %v", err)
			}

			err = hook.Uninstall(configPath)
			if (err != nil) != tt.expectedError {
				t.Fatalf("Uninstall() error = %v, wantErr %v", err, tt.expectedError)
			}

			content, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("Failed to read config file after Uninstall: %v", err)
			}

			if string(content) != tt.expectedFile {
				t.Errorf("Uninstall() modified file content mismatch:\nGot:\n%q\nWant:\n%q", string(content), tt.expectedFile)
			}
		})
	}
}

func TestStatus(t *testing.T) {
	tests := []struct {
		name           string
		initial        string
		createFile     bool
		expectedStatus bool
		expectedError  bool
	}{
		{
			name:           "Status when installed",
			initial:        "export FOO=bar\n" + expectedHook,
			createFile:     true,
			expectedStatus: true,
			expectedError:  false,
		},
		{
			name:           "Status when not installed",
			initial:        "export FOO=bar\n",
			createFile:     true,
			expectedStatus: false,
			expectedError:  false,
		},
		{
			name:           "Status on non-existent file",
			initial:        "",
			createFile:     false,
			expectedStatus: false,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, ".zshrc")

			if tt.createFile {
				err := os.WriteFile(configPath, []byte(tt.initial), 0644)
				if err != nil {
					t.Fatalf("Failed to create initial file: %v", err)
				}
			}

			installed, err := hook.Status(configPath)
			if (err != nil) != tt.expectedError {
				t.Fatalf("Status() error = %v, wantErr %v", err, tt.expectedError)
			}

			if installed != tt.expectedStatus {
				t.Errorf("Status() = %v, want %v", installed, tt.expectedStatus)
			}
		})
	}
}
