package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	MarkerStart = "# >>> faaah hook >>>"
	MarkerEnd   = "# <<< faaah hook <<<"
)

func Install(configPath string, binaryPath string) error {
	content, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	strContent := string(content)
	if strings.Contains(strContent, MarkerStart) {
		return nil
	}

	hookBlock := fmt.Sprintf("\n%s\ntrap 'if [ -n \"$ZSH_VERSION\" ] && [ \"$ZSH_EVAL_CONTEXT\" != \"toplevel:trap\" ] && [ \"$ZSH_EVAL_CONTEXT\" != \"cmdarg:trap\" ]; then :; else (%s play >/dev/null 2>&1 &); fi' ERR\n%s\n", MarkerStart, binaryPath, MarkerEnd)

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(configPath, []byte(strContent+hookBlock), 0644)
}

func Uninstall(configPath string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	strContent := string(content)
	if !strings.Contains(strContent, MarkerStart) {
		return nil
	}

	startIdx := strings.Index(strContent, MarkerStart)
	if startIdx == -1 {
		return nil
	}
	endIdx := strings.Index(strContent[startIdx:], MarkerEnd)
	if endIdx == -1 {
		return nil
	}
	endIdx += startIdx + len(MarkerEnd)

	if startIdx > 0 && strContent[startIdx-1] == '\n' {
		startIdx--
	}

	if endIdx < len(strContent) && strContent[endIdx] == '\n' {
		endIdx++
	}

	newContent := strContent[:startIdx] + strContent[endIdx:]

	return os.WriteFile(configPath, []byte(newContent), 0644)
}

func Status(configPath string) (bool, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return strings.Contains(string(content), MarkerStart) && strings.Contains(string(content), MarkerEnd), nil
}
