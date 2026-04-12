// tui.go
package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Mode describes which UI state the app is in.
type Mode int

const (
	ModeNormal Mode = iota
	ModeEditing
	ModeSearching
	ModeConfirmDelete
	ModeConfirmDeleteGroup
	ModeNewGroupPrompt
)

// App is the root Bubbletea model.
type App struct {
	cfg        ParsedConfig
	configPath string
	keys       KeyMap

	listPanel   ListPanel
	detailPanel DetailPanel
	editForm    EditForm
	searchInput textinput.Model
	groupInput  textinput.Model

	mode      Mode
	focusLeft bool // true = left panel focused, false = right
	width     int
	height    int
	statusMsg string
	err       error
}

// NewApp constructs the App from a loaded config.
func NewApp(cfg ParsedConfig, configPath string) App {
	lp := NewListPanel(cfg.Groups, Styles{})
	lp.focused = true

	si := textinput.New()
	si.Placeholder = "search..."
	si.Width = 20

	gi := textinput.New()
	gi.Placeholder = "group name"
	gi.Width = 20

	return App{
		cfg:         cfg,
		configPath:  configPath,
		keys:        DefaultKeyMap(),
		listPanel:   lp,
		focusLeft:   true,
		searchInput: si,
		groupInput:  gi,
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updatePanelSizes()
		return a, nil

	case tea.KeyMsg:
		return a.handleKey(msg)

	case errMsg:
		a.statusMsg = fmt.Sprintf("error: %v", msg.err)
		return a, nil
	}

	return a, nil
}

// updatePanelSizes recalculates panel dimensions from terminal size.
func (a *App) updatePanelSizes() {
	statusH := 1
	contentH := a.height - statusH
	leftW := a.width / 3
	rightW := a.width - leftW - 1 // -1 for border

	a.listPanel.width = leftW
	a.listPanel.height = contentH
	a.detailPanel.width = rightW
	a.detailPanel.height = contentH
	a.editForm.width = rightW
	a.editForm.height = contentH
}

// handleKey routes key events based on the current mode.
func (a App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.mode {
	case ModeSearching:
		return a.handleSearchKey(msg)
	case ModeEditing:
		return a.handleEditKey(msg)
	case ModeConfirmDelete, ModeConfirmDeleteGroup:
		return a.handleConfirmKey(msg)
	case ModeNewGroupPrompt:
		return a.handleNewGroupKey(msg)
	default:
		return a.handleNormalKey(msg)
	}
}

func (a App) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "q":
		return a, tea.Quit

	case msg.String() == "tab":
		a.focusLeft = !a.focusLeft
		a.listPanel.focused = a.focusLeft
		a.detailPanel.focused = !a.focusLeft
		return a, nil

	case msg.String() == "up" || msg.String() == "k":
		if a.focusLeft {
			a.listPanel.MoveUp()
			a.syncDetail()
		}
		return a, nil

	case msg.String() == "down" || msg.String() == "j":
		if a.focusLeft {
			a.listPanel.MoveDown()
			a.syncDetail()
		}
		return a, nil

	case msg.String() == "enter":
		if a.focusLeft {
			conn := a.listPanel.SelectedConnection()
			if conn == nil {
				a.listPanel.ToggleCollapse()
				return a, nil
			}
			return a, func() tea.Msg {
				if err := Connect(conn.Host); err != nil {
					return errMsg{err}
				}
				return nil
			}
		}
		return a, nil

	case msg.String() == "e":
		conn := a.listPanel.SelectedConnection()
		if conn == nil {
			return a, nil
		}
		group := a.listPanel.SelectedGroup()
		groupName := ""
		if group != nil {
			groupName = group.Name
		}
		a.editForm = NewEditForm(conn, groupName, false)
		a.editForm.width = a.detailPanel.width
		a.editForm.height = a.detailPanel.height
		a.mode = ModeEditing
		a.focusLeft = false
		a.listPanel.focused = false
		return a, textinput.Blink

	case msg.String() == "n":
		group := a.listPanel.SelectedGroup()
		groupName := ""
		if group != nil {
			groupName = group.Name
		}
		a.editForm = NewEditForm(nil, groupName, true)
		a.editForm.width = a.detailPanel.width
		a.editForm.height = a.detailPanel.height
		a.mode = ModeEditing
		a.focusLeft = false
		a.listPanel.focused = false
		return a, textinput.Blink

	case msg.String() == "g":
		a.groupInput.SetValue("")
		a.groupInput.Focus()
		a.mode = ModeNewGroupPrompt
		return a, textinput.Blink

	case msg.String() == "d":
		conn := a.listPanel.SelectedConnection()
		if conn != nil {
			a.statusMsg = fmt.Sprintf("Delete %q? (y/N)", conn.Host)
			a.mode = ModeConfirmDelete
		} else {
			group := a.listPanel.SelectedGroup()
			if group != nil && group.Name != "" {
				a.statusMsg = fmt.Sprintf("Move connections in %q to Ungrouped? (y/N)", group.Name)
				a.mode = ModeConfirmDeleteGroup
			}
		}
		return a, nil

	case msg.String() == "/":
		a.searchInput.SetValue("")
		a.searchInput.Focus()
		a.mode = ModeSearching
		return a, textinput.Blink
	}

	return a, nil
}

func (a App) handleEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+s" {
		conn := a.editForm.ToConnection()
		if conn.Host == "" {
			a.statusMsg = "Host alias is required"
			return a, nil
		}
		a.saveConnection(conn, a.editForm.groupName, a.editForm.isNew, a.editForm.originalHost)
		a.mode = ModeNormal
		a.focusLeft = true
		a.listPanel.focused = true
		a.syncDetail()
		return a, nil
	}
	if msg.String() == "esc" {
		a.mode = ModeNormal
		a.focusLeft = true
		a.listPanel.focused = true
		return a, nil
	}
	var cmd tea.Cmd
	a.editForm, cmd = a.editForm.Update(msg)
	return a, cmd
}

func (a App) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		a.mode = ModeNormal
		a.listPanel.filter = ""
		a.listPanel.groups = a.cfg.Groups
		a.syncDetail()
		return a, nil
	}
	var cmd tea.Cmd
	a.searchInput, cmd = a.searchInput.Update(msg)
	a.listPanel.filter = a.searchInput.Value()
	a.syncDetail()
	return a, cmd
}

func (a App) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		if a.mode == ModeConfirmDelete {
			a.deleteSelectedConnection()
		} else {
			a.deleteSelectedGroup()
		}
		a.mode = ModeNormal
		a.statusMsg = ""
		a.syncDetail()
	case "n", "N", "esc":
		a.mode = ModeNormal
		a.statusMsg = ""
	default:
		return a, nil
	}
	return a, nil
}

func (a App) handleNewGroupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		a.mode = ModeNormal
		return a, nil
	}
	if msg.String() == "enter" {
		name := strings.TrimSpace(a.groupInput.Value())
		if name != "" {
			// Check for duplicate group name
			duplicate := false
			for _, g := range a.cfg.Groups {
				if g.Name == name {
					duplicate = true
					break
				}
			}
			if duplicate {
				a.statusMsg = fmt.Sprintf("Group %q already exists", name)
				a.mode = ModeNormal
				return a, nil
			}
			a.cfg.Groups = append(a.cfg.Groups, Group{Name: name})
			a.listPanel.groups = a.cfg.Groups
			a.listPanel.cursor = Cursor{GroupIdx: len(a.cfg.Groups) - 1, ConnIdx: -1}
			_ = Save(a.configPath, a.cfg)
		}
		a.mode = ModeNormal
		return a, nil
	}
	var cmd tea.Cmd
	a.groupInput, cmd = a.groupInput.Update(msg)
	return a, cmd
}

// saveConnection upserts a connection into the config and persists to disk.
func (a *App) saveConnection(conn Connection, groupName string, isNew bool, originalHost string) {
	if !isNew {
		// Update in-place, matching by originalHost to handle host alias renames
		for gi := range a.cfg.Groups {
			for ci := range a.cfg.Groups[gi].Connections {
				if a.cfg.Groups[gi].Connections[ci].Host == originalHost {
					a.cfg.Groups[gi].Connections[ci] = conn
					break
				}
			}
		}
	} else {
		// Append to the named group (or ungrouped)
		added := false
		for gi := range a.cfg.Groups {
			if a.cfg.Groups[gi].Name == groupName {
				a.cfg.Groups[gi].Connections = append(a.cfg.Groups[gi].Connections, conn)
				added = true
				break
			}
		}
		if !added {
			idx := ungroupedIndex(&a.cfg)
			a.cfg.Groups[idx].Connections = append(a.cfg.Groups[idx].Connections, conn)
		}
	}
	a.listPanel.groups = a.cfg.Groups
	_ = Save(a.configPath, a.cfg)
}

// deleteSelectedConnection removes the currently selected connection.
func (a *App) deleteSelectedConnection() {
	cur := a.listPanel.cursor
	if cur.GroupIdx >= len(a.cfg.Groups) || cur.ConnIdx < 0 {
		return
	}
	g := &a.cfg.Groups[cur.GroupIdx]
	g.Connections = append(g.Connections[:cur.ConnIdx], g.Connections[cur.ConnIdx+1:]...)
	if cur.ConnIdx >= len(g.Connections) && cur.ConnIdx > 0 {
		a.listPanel.cursor.ConnIdx--
	}
	a.listPanel.groups = a.cfg.Groups
	_ = Save(a.configPath, a.cfg)
}

// deleteSelectedGroup moves all connections to Ungrouped and removes the group.
func (a *App) deleteSelectedGroup() {
	cur := a.listPanel.cursor
	if cur.GroupIdx >= len(a.cfg.Groups) {
		return
	}
	g := a.cfg.Groups[cur.GroupIdx]
	if g.Name == "" {
		return // never delete Ungrouped
	}
	// Move connections to Ungrouped
	if len(g.Connections) > 0 {
		idx := ungroupedIndex(&a.cfg)
		if idx == cur.GroupIdx {
			// shouldn't happen since Name != ""
			return
		}
		a.cfg.Groups[idx].Connections = append(a.cfg.Groups[idx].Connections, g.Connections...)
	}
	a.cfg.Groups = append(a.cfg.Groups[:cur.GroupIdx], a.cfg.Groups[cur.GroupIdx+1:]...)
	if a.listPanel.cursor.GroupIdx >= len(a.cfg.Groups) {
		a.listPanel.cursor.GroupIdx = len(a.cfg.Groups) - 1
	}
	a.listPanel.cursor.ConnIdx = -1
	a.listPanel.groups = a.cfg.Groups
	_ = Save(a.configPath, a.cfg)
}

// syncDetail keeps detailPanel in sync with the list cursor.
// Safe to call from value receivers (modifies the local copy).
func (a *App) syncDetail() {
	a.detailPanel.conn = a.listPanel.SelectedConnection()
	a.detailPanel.group = a.listPanel.SelectedGroup()
}

// View renders the full application.
func (a App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	// Sync detail panel inline (value receiver safe — local copy only)
	a.detailPanel.conn = a.listPanel.SelectedConnection()
	a.detailPanel.group = a.listPanel.SelectedGroup()

	// Right panel content
	var rightContent string
	if a.mode == ModeEditing {
		rightContent = a.editForm.View()
	} else {
		rightContent = a.detailPanel.View()
	}

	leftContent := a.listPanel.View()

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent)

	// Status bar (doubles as search input bar in search mode)
	status := a.buildStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, panels, status)
}

func (a App) buildStatusBar() string {
	style := StyleStatusBar.Width(a.width)

	switch a.mode {
	case ModeConfirmDelete, ModeConfirmDeleteGroup:
		return StyleConfirmPrompt.Width(a.width).Render(" " + a.statusMsg)
	case ModeEditing:
		return StyleStatusEdit.Width(a.width).Render(" — editing mode —")
	case ModeNewGroupPrompt:
		return style.Render(" New group: " + a.groupInput.View())
	case ModeSearching:
		return style.Render(" / " + a.searchInput.View())
	}

	if a.statusMsg != "" {
		return style.Render(" " + a.statusMsg)
	}
	return style.Render(" tab:switch  n:new  g:group  /:search  q:quit")
}

type errMsg struct{ err error }
