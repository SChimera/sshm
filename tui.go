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
	ModeRenameGroupPrompt
)

// App is the root Bubbletea model.
type App struct {
	styles     Styles
	cfg        ParsedConfig
	configPath string
	keys       KeyMap

	listPanel   ListPanel
	detailPanel DetailPanel
	editForm    EditForm
	searchInput textinput.Model
	groupInput  textinput.Model

	mode          Mode
	focusLeft     bool // true = left panel focused, false = right
	width         int
	height        int
	statusMsg     string
	err           error
	renameOldName string // group name captured when entering rename mode
}

// NewApp constructs the App from a loaded config.
func NewApp(cfg ParsedConfig, configPath string, styles Styles) App {
	lp := NewListPanel(cfg.Groups, styles)
	lp.focused = true

	si := textinput.New()
	si.Placeholder = "search..."
	si.Width = 20

	gi := textinput.New()
	gi.Placeholder = "group name"
	gi.Width = 20

	return App{
		styles:      styles,
		cfg:         cfg,
		configPath:  configPath,
		keys:        DefaultKeyMap(),
		listPanel:   lp,
		detailPanel: NewDetailPanel(styles),
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
	case ModeRenameGroupPrompt:
		return a.handleRenameGroupKey(msg)
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
		a.editForm = NewEditForm(conn, groupName, false, a.styles)
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
		// Clear any active filter so the newly-created connection is visible
		// regardless of whether its Host matches the current filter.
		if a.listPanel.filter != "" {
			a.listPanel.filter = ""
			a.searchInput.SetValue("")
			a.restoreCursor(&Group{Name: groupName}, nil)
		}
		a.editForm = NewEditForm(nil, groupName, true, a.styles)
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

	case msg.String() == "r":
		group := a.listPanel.SelectedGroup()
		if group == nil || group.Name == "" {
			// Can't rename Ungrouped or when nothing is selected.
			return a, nil
		}
		a.renameOldName = group.Name
		a.groupInput.SetValue(group.Name)
		a.groupInput.CursorEnd()
		a.groupInput.Focus()
		a.mode = ModeRenameGroupPrompt
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
		a.searchInput.SetValue(a.listPanel.filter)
		a.searchInput.CursorEnd()
		a.searchInput.Focus()
		a.mode = ModeSearching
		return a, textinput.Blink

	case msg.String() == "esc":
		if a.listPanel.filter != "" {
			// Capture the current selection so we can restore it after the filter clears.
			selGroup := a.listPanel.SelectedGroup()
			selConn := a.listPanel.SelectedConnection()
			a.listPanel.filter = ""
			a.searchInput.SetValue("")
			a.restoreCursor(selGroup, selConn)
			a.syncDetail()
		}
		return a, nil
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
	switch msg.String() {
	case "esc":
		// Exit search mode; keep the filter active so results stay visible.
		a.mode = ModeNormal
		a.syncDetail()
		return a, nil
	case "up":
		a.listPanel.MoveUp()
		a.syncDetail()
		return a, nil
	case "down":
		a.listPanel.MoveDown()
		a.syncDetail()
		return a, nil
	case "enter":
		conn := a.listPanel.SelectedConnection()
		if conn == nil {
			return a, nil
		}
		host := conn.Host
		return a, func() tea.Msg {
			if err := Connect(host); err != nil {
				return errMsg{err}
			}
			return nil
		}
	}
	var cmd tea.Cmd
	a.searchInput, cmd = a.searchInput.Update(msg)
	a.listPanel.filter = a.searchInput.Value()
	a.listPanel.ClampCursor()
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

func (a App) handleRenameGroupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		a.mode = ModeNormal
		a.renameOldName = ""
		return a, nil
	}
	if msg.String() == "enter" {
		name := strings.TrimSpace(a.groupInput.Value())
		old := a.renameOldName
		a.mode = ModeNormal
		a.renameOldName = ""
		if name == "" || name == old {
			return a, nil
		}
		for _, g := range a.cfg.Groups {
			if g.Name == name {
				a.statusMsg = fmt.Sprintf("Group %q already exists", name)
				return a, nil
			}
		}
		for gi := range a.cfg.Groups {
			if a.cfg.Groups[gi].Name == old {
				a.cfg.Groups[gi].Name = name
				break
			}
		}
		// Carry collapse state across the rename.
		if a.listPanel.collapsed[old] {
			a.listPanel.collapsed[name] = true
			delete(a.listPanel.collapsed, old)
		}
		a.listPanel.groups = a.cfg.Groups
		a.persist()
		return a, nil
	}
	var cmd tea.Cmd
	a.groupInput, cmd = a.groupInput.Update(msg)
	return a, cmd
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
			// Clear any active filter so the new (empty) group is visible.
			a.listPanel.filter = ""
			a.searchInput.SetValue("")
			a.listPanel.cursor = Cursor{GroupIdx: len(a.cfg.Groups) - 1, ConnIdx: -1}
			a.persist()
		}
		a.mode = ModeNormal
		return a, nil
	}
	var cmd tea.Cmd
	a.groupInput, cmd = a.groupInput.Update(msg)
	return a, cmd
}

// restoreCursor re-points the list cursor at the given group/connection in the
// unfiltered list, falling back to a clamp if either lookup fails.
func (a *App) restoreCursor(group *Group, conn *Connection) {
	if group != nil {
		for gi, g := range a.cfg.Groups {
			if g.Name != group.Name {
				continue
			}
			a.listPanel.cursor = Cursor{GroupIdx: gi, ConnIdx: -1}
			if conn != nil {
				for ci, c := range g.Connections {
					if c.Host == conn.Host {
						a.listPanel.cursor.ConnIdx = ci
						break
					}
				}
			}
			a.listPanel.ClampCursor()
			return
		}
	}
	a.listPanel.ClampCursor()
}

// persist writes the config to disk and records any error in the status bar.
func (a *App) persist() {
	if err := Save(a.configPath, a.cfg); err != nil {
		a.statusMsg = fmt.Sprintf("save failed: %v", err)
	}
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
	a.persist()
}

// deleteSelectedConnection removes the currently selected connection.
// Uses group-name + host-alias lookup so it's safe when a search filter is active.
func (a *App) deleteSelectedConnection() {
	group := a.listPanel.SelectedGroup()
	conn := a.listPanel.SelectedConnection()
	if group == nil || conn == nil {
		return
	}
	for gi := range a.cfg.Groups {
		if a.cfg.Groups[gi].Name != group.Name {
			continue
		}
		for ci := range a.cfg.Groups[gi].Connections {
			if a.cfg.Groups[gi].Connections[ci].Host == conn.Host {
				a.cfg.Groups[gi].Connections = append(
					a.cfg.Groups[gi].Connections[:ci],
					a.cfg.Groups[gi].Connections[ci+1:]...)
				break
			}
		}
		break
	}
	a.listPanel.groups = a.cfg.Groups
	a.listPanel.ClampCursor()
	a.persist()
}

// deleteSelectedGroup moves all connections to Ungrouped and removes the group.
func (a *App) deleteSelectedGroup() {
	group := a.listPanel.SelectedGroup()
	if group == nil || group.Name == "" {
		return
	}
	targetName := group.Name
	var srcIdx = -1
	for gi := range a.cfg.Groups {
		if a.cfg.Groups[gi].Name == targetName {
			srcIdx = gi
			break
		}
	}
	if srcIdx == -1 {
		return
	}
	src := a.cfg.Groups[srcIdx]
	if len(src.Connections) > 0 {
		idx := ungroupedIndex(&a.cfg)
		if idx == srcIdx {
			return
		}
		a.cfg.Groups[idx].Connections = append(a.cfg.Groups[idx].Connections, src.Connections...)
	}
	a.cfg.Groups = append(a.cfg.Groups[:srcIdx], a.cfg.Groups[srcIdx+1:]...)
	delete(a.listPanel.collapsed, targetName)
	a.listPanel.groups = a.cfg.Groups
	a.listPanel.cursor.ConnIdx = -1
	a.listPanel.ClampCursor()
	a.persist()
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
	style := a.styles.StatusBar.Width(a.width)

	switch a.mode {
	case ModeConfirmDelete, ModeConfirmDeleteGroup:
		return a.styles.ConfirmPrompt.Width(a.width).Render(" " + a.statusMsg)
	case ModeEditing:
		return a.styles.StatusEdit.Width(a.width).Render(" — editing mode —")
	case ModeNewGroupPrompt:
		return style.Render(" New group: " + a.groupInput.View())
	case ModeRenameGroupPrompt:
		return style.Render(" Rename group: " + a.groupInput.View())
	case ModeSearching:
		return style.Render(" / " + a.searchInput.View())
	}

	if a.statusMsg != "" {
		return style.Render(" " + a.statusMsg)
	}
	if a.listPanel.filter != "" {
		return style.Render(fmt.Sprintf(" filter: %q  esc:clear  /:edit", a.listPanel.filter))
	}
	return style.Render(" tab:switch  n:new  g:group  r:rename  /:search  q:quit")
}

type errMsg struct{ err error }
