// main.go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	configPath, err := ConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not determine config path: %v\n", err)
		os.Exit(1)
	}

	cfg, err := Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read %s: %v\n", configPath, err)
		os.Exit(1)
	}

	theme, themeErr := ResolveTheme()
	styles := BuildStyles(theme)

	app := NewApp(cfg, configPath, styles)
	if themeErr != nil {
		app.statusMsg = fmt.Sprintf("theme: %v", themeErr)
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
