// theme_windows.go
//go:build windows

package main

// QueryTerminalTheme on Windows always returns false.
// Windows users configure colours via config.toml.
func QueryTerminalTheme() (Theme, bool) {
	return Theme{}, false
}
