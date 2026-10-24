package tooling

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func toolFixture(t *testing.T, dir string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX executable fixture")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "ffmpeg")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestResolutionPrecedence(t *testing.T) {
	root := t.TempDir()
	configured := toolFixture(t, filepath.Join(root, "configured tool"))
	fromPATH := toolFixture(t, filepath.Join(root, "path"))
	fromFallback := toolFixture(t, filepath.Join(root, "homebrew"))
	t.Setenv("PATH", filepath.Dir(fromPATH))
	fallbacks := []string{filepath.Dir(fromFallback)}
	for _, tc := range []struct{ configured, want string }{
		{configured, configured}, {"", fromPATH},
	} {
		got, err := find(tc.configured, "ffmpeg", fallbacks)
		if err != nil || got != tc.want {
			t.Fatalf("resolve(%q) = %q, %v; want %q", tc.configured, got, err, tc.want)
		}
	}
	t.Setenv("PATH", t.TempDir())
	got, err := find("", "ffmpeg", fallbacks)
	if err != nil || got != fromFallback {
		t.Fatalf("Finder-like PATH must use fallback: %q %v", got, err)
	}
}

func TestInvalidOverrideIsNotSilentlyIgnored(t *testing.T) {
	tool := toolFixture(t, t.TempDir())
	t.Setenv("PATH", filepath.Dir(tool))
	missing := filepath.Join(t.TempDir(), "missing-ffmpeg")
	if _, err := find(missing, "ffmpeg", []string{filepath.Dir(tool)}); err == nil || !strings.Contains(err.Error(), "系统设置") {
		t.Fatalf("invalid explicit override must report actionable error: %v", err)
	}
}

func TestMissingToolHasActionableError(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, err := find("", "ffmpeg", nil)
	if !errors.Is(err, exec.ErrNotFound) || !strings.Contains(err.Error(), "系统设置") || !strings.Contains(err.Error(), "安装") {
		t.Fatalf("missing tool must explain installation/configuration: %v", err)
	}
}
