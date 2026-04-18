// keys.go
package main

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds all key bindings for the application.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Connect  key.Binding
	Edit     key.Binding
	New      key.Binding
	NewGroup key.Binding
	Delete   key.Binding
	Rename   key.Binding
	Search   key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Save     key.Binding
	Confirm  key.Binding
	Cancel   key.Binding
	Quit     key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Connect:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "connect")),
		Edit:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		New:      key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new connection")),
		NewGroup: key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "new group")),
		Delete:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		Rename:   key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename group")),
		Search:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch panel")),
		ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field")),
		Save:     key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
		Confirm:  key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "confirm")),
		Cancel:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		Quit:     key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
}
