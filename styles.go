// styles.go
package main

import "github.com/charmbracelet/lipgloss"

var (
	// Colours
	colorPrimary  = lipgloss.Color("#7dd3fc") // light blue
	colorMuted    = lipgloss.Color("#555555")
	colorAccent   = lipgloss.Color("#f59e0b") // amber — used for edit mode
	colorSelected = lipgloss.Color("#1e2a4a")
	colorBorder   = lipgloss.Color("#2a2a4e")
	colorDanger   = lipgloss.Color("#f87171") // red — used for delete confirm

	// Panel styles
	StyleListPanel = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(colorBorder)

	StyleDetailPanel = lipgloss.NewStyle().
				PaddingLeft(1)

	// Text styles
	StyleGroupName = lipgloss.NewStyle().
			Foreground(colorPrimary)

	StyleSelectedItem = lipgloss.NewStyle().
				Background(colorSelected).
				Foreground(lipgloss.Color("#ffffff")).
				PaddingLeft(1)

	StyleNormalItem = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cccccc")).
			PaddingLeft(1)

	StyleMuted = lipgloss.NewStyle().
			Foreground(colorMuted)

	StyleLabel = lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(14)

	StyleValue = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e0e0e0"))

	StyleTitle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	StyleStatusBar = lipgloss.NewStyle().
			Background(lipgloss.Color("#1e1e3a")).
			Foreground(colorMuted).
			PaddingLeft(1).
			PaddingRight(1)

	StyleStatusEdit = lipgloss.NewStyle().
			Background(lipgloss.Color("#1e1e3a")).
			Foreground(colorAccent).
			PaddingLeft(1).
			PaddingRight(1)

	StyleConfirmPrompt = lipgloss.NewStyle().
				Foreground(colorDanger)

	StyleAccent = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
)
