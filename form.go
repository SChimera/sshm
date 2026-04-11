// form.go
package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// fieldDef describes one form field.
type fieldDef struct {
	label       string
	placeholder string
}

var formFields = []fieldDef{
	{"Host alias", "e.g. web-prod"},
	{"HostName", "hostname or IP"},
	{"User", "username"},
	{"Port", "22"},
	{"IdentityFile", "~/.ssh/id_ed25519"},
	{"ProxyJump", "bastion-host"},
	{"ForwardAgent", "yes / no"},
	{"LocalForward", "8080 localhost:8080"},
	{"RemoteForward", "9090 localhost:9090"},
	{"Extra", "raw directives"},
}

// EditForm is the right-panel form for adding or editing a connection.
type EditForm struct {
	inputs      []textinput.Model
	activeField int
	isNew       bool
	groupName   string // group to save into
	width       int
	height      int
}

// NewEditForm creates a form pre-filled from conn. Pass nil for a new connection.
func NewEditForm(conn *Connection, groupName string, isNew bool) EditForm {
	values := make([]string, len(formFields))
	if conn != nil {
		values[0] = conn.Host
		values[1] = conn.HostName
		values[2] = conn.User
		values[3] = conn.Port
		values[4] = conn.IdentityFile
		values[5] = conn.ProxyJump
		values[6] = conn.ForwardAgent
		values[7] = conn.LocalForward
		values[8] = conn.RemoteForward
		values[9] = conn.Extra
	}

	inputs := make([]textinput.Model, len(formFields))
	for i, f := range formFields {
		t := textinput.New()
		t.Placeholder = f.placeholder
		t.SetValue(values[i])
		t.Width = 30
		if i == 0 {
			t.Focus()
		}
		inputs[i] = t
	}

	return EditForm{
		inputs:      inputs,
		activeField: 0,
		isNew:       isNew,
		groupName:   groupName,
	}
}

// Update handles key events for the form.
func (f EditForm) Update(msg tea.Msg) (EditForm, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			f.inputs[f.activeField].Blur()
			f.activeField = (f.activeField + 1) % len(f.inputs)
			f.inputs[f.activeField].Focus()
			cmds = append(cmds, textinput.Blink)
		case "shift+tab":
			f.inputs[f.activeField].Blur()
			f.activeField = (f.activeField - 1 + len(f.inputs)) % len(f.inputs)
			f.inputs[f.activeField].Focus()
			cmds = append(cmds, textinput.Blink)
		default:
			var cmd tea.Cmd
			f.inputs[f.activeField], cmd = f.inputs[f.activeField].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return f, tea.Batch(cmds...)
}

// ToConnection extracts a Connection from the form values.
func (f EditForm) ToConnection() Connection {
	return Connection{
		Host:          strings.TrimSpace(f.inputs[0].Value()),
		HostName:      strings.TrimSpace(f.inputs[1].Value()),
		User:          strings.TrimSpace(f.inputs[2].Value()),
		Port:          strings.TrimSpace(f.inputs[3].Value()),
		IdentityFile:  strings.TrimSpace(f.inputs[4].Value()),
		ProxyJump:     strings.TrimSpace(f.inputs[5].Value()),
		ForwardAgent:  strings.TrimSpace(f.inputs[6].Value()),
		LocalForward:  strings.TrimSpace(f.inputs[7].Value()),
		RemoteForward: strings.TrimSpace(f.inputs[8].Value()),
		Extra:         strings.TrimSpace(f.inputs[9].Value()),
	}
}

// View renders the edit form.
func (f EditForm) View() string {
	var lines []string

	title := "Edit Connection"
	if f.isNew {
		title = "New Connection"
	}
	lines = append(lines, StyleAccent.Render(title))
	lines = append(lines, "")

	for i, fd := range formFields {
		label := StyleLabel.Render(fd.label)
		input := f.inputs[i].View()
		if i == f.activeField {
			input = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Render(input)
		}
		lines = append(lines, label)
		lines = append(lines, input)
		lines = append(lines, "")
	}

	lines = append(lines, StyleMuted.Render("tab:next  shift+tab:prev  ctrl+s:save  esc:cancel"))

	return StyleDetailPanel.Width(f.width).Height(f.height).Render(strings.Join(lines, "\n"))
}
