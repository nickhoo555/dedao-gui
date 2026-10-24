// Package tooling resolves external tools consistently for desktop processes,
// whose environment can differ from the user's interactive shell.
package tooling

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Resolve honours an explicit setting, then PATH, then macOS package-manager
// locations. An invalid explicit override is an error, not a silent fallback.
func Resolve(configured, name string) (string, error) {
	var dirs []string
	if runtime.GOOS == "darwin" {
		dirs = []string{"/opt/homebrew/bin", "/usr/local/bin"}
	}
	return find(configured, name, dirs)
}

func find(configured, name string, fallbackDirs []string) (string, error) {
	if strings.TrimSpace(configured) != "" {
		path, err := exec.LookPath(configured)
		if err != nil {
			return "", fmt.Errorf("%s 路径无效或不可执行，请在系统设置中检查路径 %q: %w", name, configured, err)
		}
		return path, nil
	}
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}
	for _, dir := range fallbackDirs {
		if path, err := exec.LookPath(filepath.Join(dir, name)); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("未找到 %s，请先安装，并在系统设置中填写其可执行文件完整路径: %w", name, exec.ErrNotFound)
}
