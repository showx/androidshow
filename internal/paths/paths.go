package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func OSName() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "macos"
	default:
		return "linux"
	}
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func ArchName() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "arm64":
		return "arm64"
	case "386":
		return "x86"
	default:
		return runtime.GOARCH
	}
}

func DataDir() string {
	if override := os.Getenv("ANDROIDSHOW_HOME"); override != "" {
		return expandHome(override)
	}
	switch OSName() {
	case "windows":
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = filepath.Join(homeDir(), "AppData", "Local")
		}
		return filepath.Join(base, "androidshow")
	case "macos":
		return filepath.Join(homeDir(), "Library", "Application Support", "androidshow")
	default:
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return filepath.Join(xdg, "androidshow")
		}
		return filepath.Join(homeDir(), ".local", "share", "androidshow")
	}
}

func ToolsDir() string {
	return filepath.Join(DataDir(), "tools")
}

func CacheDir() string {
	return filepath.Join(DataDir(), "cache")
}

func ToolHome(name string) string {
	return filepath.Join(ToolsDir(), name)
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}

func expandHome(p string) string {
	if p == "~" {
		return homeDir()
	}
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) {
		return filepath.Join(homeDir(), p[2:])
	}
	return p
}
