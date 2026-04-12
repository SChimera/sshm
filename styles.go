// styles.go
package main

import "github.com/charmbracelet/lipgloss"

// Styles holds all lipgloss styles for the application, built from a Theme.
type Styles struct {
	ListPanel             lipgloss.Style
	DetailPanel           lipgloss.Style
	GroupName             lipgloss.Style
	SelectedItem          lipgloss.Style
	SelectedItemUnfocused lipgloss.Style
	NormalItem            lipgloss.Style
	Muted                 lipgloss.Style
	Label                 lipgloss.Style
	Value                 lipgloss.Style
	Title                 lipgloss.Style
	StatusBar             lipgloss.Style
	StatusEdit            lipgloss.Style
	ConfirmPrompt         lipgloss.Style
	Accent                lipgloss.Style
}

// BuildStyles constructs all lipgloss styles from a resolved Theme.
func BuildStyles(t Theme) Styles {
	primary := lipgloss.Color(t.Primary)
	muted   := lipgloss.Color(t.Muted)
	accent  := lipgloss.Color(t.Accent)
	bg      := lipgloss.Color(t.Background)
	surface := lipgloss.Color(t.Surface)
	border  := lipgloss.Color(t.Border)
	danger  := lipgloss.Color(t.Danger)
	fg      := lipgloss.Color(t.Foreground)

	return Styles{
		ListPanel: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(border),

		DetailPanel: lipgloss.NewStyle().PaddingLeft(1),

		GroupName: lipgloss.NewStyle().Foreground(primary),

		SelectedItem: lipgloss.NewStyle().
			Background(surface).
			Foreground(fg).
			PaddingLeft(1),

		SelectedItemUnfocused: lipgloss.NewStyle().
			Background(bg).
			Foreground(muted).
			PaddingLeft(1),

		NormalItem: lipgloss.NewStyle().
			Foreground(fg).
			PaddingLeft(1),

		Muted: lipgloss.NewStyle().Foreground(muted),

		Label: lipgloss.NewStyle().
			Foreground(muted).
			Width(14),

		Value: lipgloss.NewStyle().Foreground(fg),

		Title: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true),

		StatusBar: lipgloss.NewStyle().
			Background(surface).
			Foreground(muted).
			PaddingLeft(1).
			PaddingRight(1),

		StatusEdit: lipgloss.NewStyle().
			Background(surface).
			Foreground(accent).
			PaddingLeft(1).
			PaddingRight(1),

		ConfirmPrompt: lipgloss.NewStyle().Foreground(danger),

		Accent: lipgloss.NewStyle().Foreground(accent).Bold(true),
	}
}
