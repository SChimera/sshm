// theme.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

// Theme holds the 8 semantic colour roles used by sshm.
// All values are CSS hex strings (e.g. "#7dd3fc").
type Theme struct {
	Background string // base00 — main background
	Surface    string // base01 — panel/input backgrounds, status bar
	Muted      string // base03 — hints, placeholders
	Foreground string // base05 — default text
	Primary    string // base0D — selected items, group names
	Accent     string // base0E — edit mode, titles
	Danger     string // base08 — delete confirmation
	Border     string // base02 — panel borders
}

// DefaultTheme returns the built-in dark blue palette.
func DefaultTheme() Theme {
	return Theme{
		Background: "#0f0f1a",
		Surface:    "#1e1e3a",
		Muted:      "#555555",
		Foreground: "#e0e0e0",
		Primary:    "#7dd3fc",
		Accent:     "#f59e0b",
		Danger:     "#f87171",
		Border:     "#2a2a4e",
	}
}

// mergeTheme fills empty fields in dst from src.
// A field is considered empty if it is the zero string "".
func mergeTheme(dst, src Theme) Theme {
	if dst.Background == "" {
		dst.Background = src.Background
	}
	if dst.Surface == "" {
		dst.Surface = src.Surface
	}
	if dst.Muted == "" {
		dst.Muted = src.Muted
	}
	if dst.Foreground == "" {
		dst.Foreground = src.Foreground
	}
	if dst.Primary == "" {
		dst.Primary = src.Primary
	}
	if dst.Accent == "" {
		dst.Accent = src.Accent
	}
	if dst.Danger == "" {
		dst.Danger = src.Danger
	}
	if dst.Border == "" {
		dst.Border = src.Border
	}
	return dst
}

// ThemeConfigPath returns the platform-specific path to the sshm config file.
// Unix:    ~/.config/sshm/config.toml
// Windows: %APPDATA%\sshm\config.toml
func ThemeConfigPath() (string, error) {
	var base string
	if runtime.GOOS == "windows" {
		base = os.Getenv("APPDATA")
		if base == "" {
			return "", fmt.Errorf("APPDATA not set")
		}
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "sshm", "config.toml"), nil
}

type tomlThemeFile struct {
	Theme struct {
		Background string `toml:"background"`
		Surface    string `toml:"surface"`
		Muted      string `toml:"muted"`
		Foreground string `toml:"foreground"`
		Primary    string `toml:"primary"`
		Accent     string `toml:"accent"`
		Danger     string `toml:"danger"`
		Border     string `toml:"border"`
	} `toml:"theme"`
}

// LoadConfigTheme reads the sshm config file at path and returns a partial Theme.
// Only fields explicitly set in the file are populated; unset fields are "".
// Returns an error if the file exists but cannot be parsed.
func LoadConfigTheme(path string) (Theme, error) {
	var f tomlThemeFile
	if _, err := toml.DecodeFile(path, &f); err != nil {
		return Theme{}, err
	}
	return Theme{
		Background: f.Theme.Background,
		Surface:    f.Theme.Surface,
		Muted:      f.Theme.Muted,
		Foreground: f.Theme.Foreground,
		Primary:    f.Theme.Primary,
		Accent:     f.Theme.Accent,
		Danger:     f.Theme.Danger,
		Border:     f.Theme.Border,
	}, nil
}

// ResolveTheme resolves the active theme using priority order:
// 1. Config file (~/.config/sshm/config.toml or %APPDATA%\sshm\config.toml)
// 2. Terminal OSC 4 colour query
// 3. Built-in default theme
//
// Must be called before tea.NewProgram starts.
func ResolveTheme() Theme {
	result := Theme{}

	// 1. Config file (partial overrides allowed)
	if path, err := ThemeConfigPath(); err == nil {
		if t, err := LoadConfigTheme(path); err == nil {
			result = mergeTheme(result, t)
		}
	}

	// 2. Terminal OSC 4 query (fills remaining empty fields)
	if t, ok := QueryTerminalTheme(); ok {
		result = mergeTheme(result, t)
	}

	// 3. Built-in default fills anything still empty
	return mergeTheme(result, DefaultTheme())
}
