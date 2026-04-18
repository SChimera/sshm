// list.go
package main

import (
	"fmt"
	"strings"
)

// Cursor identifies the selected item in the list.
// ConnIdx == -1 means the group header itself is selected.
type Cursor struct {
	GroupIdx int
	ConnIdx  int
}

// ListPanel is the left-panel Bubbletea sub-model.
type ListPanel struct {
	styles    Styles
	groups    []Group
	cursor    Cursor
	collapsed map[string]bool // group name → collapsed? ("" for Ungrouped)
	filter    string          // empty = no filter
	width     int
	height    int
	focused   bool
}

func NewListPanel(groups []Group, styles Styles) ListPanel {
	return ListPanel{
		styles:    styles,
		groups:    groups,
		cursor:    Cursor{GroupIdx: 0, ConnIdx: -1},
		collapsed: make(map[string]bool),
	}
}

// SelectedConnection returns the currently selected Connection, or nil if a
// group header is selected.
func (l ListPanel) SelectedConnection() *Connection {
	if l.cursor.ConnIdx < 0 {
		return nil
	}
	g := l.visibleGroup(l.cursor.GroupIdx)
	if g == nil || l.cursor.ConnIdx >= len(g.Connections) {
		return nil
	}
	c := g.Connections[l.cursor.ConnIdx]
	return &c
}

// SelectedGroup returns the group at the current cursor position.
func (l ListPanel) SelectedGroup() *Group {
	g := l.visibleGroup(l.cursor.GroupIdx)
	return g
}

// ClampCursor keeps the cursor within the current visible-groups bounds.
// Called after the filter or underlying groups change.
func (l *ListPanel) ClampCursor() {
	vg := l.visibleGroups()
	if len(vg) == 0 {
		l.cursor = Cursor{GroupIdx: 0, ConnIdx: -1}
		return
	}
	if l.cursor.GroupIdx >= len(vg) {
		l.cursor.GroupIdx = len(vg) - 1
	}
	if l.cursor.GroupIdx < 0 {
		l.cursor.GroupIdx = 0
	}
	g := vg[l.cursor.GroupIdx]
	if l.cursor.ConnIdx >= len(g.Connections) {
		l.cursor.ConnIdx = len(g.Connections) - 1
	}
	if l.cursor.ConnIdx < -1 {
		l.cursor.ConnIdx = -1
	}
}

// visibleGroups returns groups that match the current filter.
func (l *ListPanel) visibleGroups() []Group {
	if l.filter == "" {
		return l.groups
	}
	var out []Group
	query := strings.ToLower(l.filter)
	for _, g := range l.groups {
		var matched []Connection
		for _, c := range g.Connections {
			if strings.Contains(strings.ToLower(c.Host), query) ||
				strings.Contains(strings.ToLower(c.HostName), query) {
				matched = append(matched, c)
			}
		}
		if len(matched) > 0 {
			out = append(out, Group{Name: g.Name, Connections: matched})
		}
	}
	return out
}

func (l *ListPanel) visibleGroup(idx int) *Group {
	vg := l.visibleGroups()
	if idx < 0 || idx >= len(vg) {
		return nil
	}
	g := vg[idx]
	return &g
}

// MoveUp moves the cursor up one item.
func (l *ListPanel) MoveUp() {
	vg := l.visibleGroups()
	if len(vg) == 0 {
		return
	}
	if l.cursor.ConnIdx > 0 {
		l.cursor.ConnIdx--
		return
	}
	if l.cursor.ConnIdx == 0 {
		l.cursor.ConnIdx = -1 // move to group header
		return
	}
	// cursor is on group header — move to previous group
	if l.cursor.GroupIdx > 0 {
		l.cursor.GroupIdx--
		g := vg[l.cursor.GroupIdx]
		if !l.collapsed[g.Name] && len(g.Connections) > 0 {
			l.cursor.ConnIdx = len(g.Connections) - 1
		} else {
			l.cursor.ConnIdx = -1
		}
	}
}

// MoveDown moves the cursor down one item.
func (l *ListPanel) MoveDown() {
	vg := l.visibleGroups()
	if len(vg) == 0 {
		return
	}
	g := vg[l.cursor.GroupIdx]
	if l.cursor.ConnIdx == -1 && !l.collapsed[g.Name] && len(g.Connections) > 0 {
		l.cursor.ConnIdx = 0
		return
	}
	if l.cursor.ConnIdx >= 0 && l.cursor.ConnIdx < len(g.Connections)-1 {
		l.cursor.ConnIdx++
		return
	}
	// move to next group
	if l.cursor.GroupIdx < len(vg)-1 {
		l.cursor.GroupIdx++
		l.cursor.ConnIdx = -1
	}
}

// ToggleCollapse collapses or expands the current group.
func (l *ListPanel) ToggleCollapse() {
	g := l.visibleGroup(l.cursor.GroupIdx)
	if g == nil {
		return
	}
	l.collapsed[g.Name] = !l.collapsed[g.Name]
	if l.collapsed[g.Name] {
		l.cursor.ConnIdx = -1
	}
}

// View renders the list panel.
func (l ListPanel) View() string {
	vg := l.visibleGroups()
	var lines []string

	for gi, g := range vg {
		groupName := g.Name
		if groupName == "" {
			groupName = "Ungrouped"
		}
		arrow := "▼"
		if l.collapsed[g.Name] {
			arrow = "▶"
		}

		groupLine := fmt.Sprintf("%s %s", arrow, groupName)
		if gi == l.cursor.GroupIdx && l.cursor.ConnIdx == -1 && l.focused {
			lines = append(lines, l.styles.SelectedItem.Render(groupLine))
		} else {
			lines = append(lines, l.styles.GroupName.Render(groupLine))
		}

		if l.collapsed[g.Name] {
			continue
		}

		for ci, c := range g.Connections {
			label := "  " + c.Host
			if gi == l.cursor.GroupIdx && ci == l.cursor.ConnIdx {
				if l.focused {
					lines = append(lines, l.styles.SelectedItem.Width(l.width-2).Render(label))
				} else {
					lines = append(lines, l.styles.SelectedItemUnfocused.
						Width(l.width-2).
						Render(label))
				}
			} else {
				lines = append(lines, l.styles.NormalItem.Render(label))
			}
		}
	}

	if len(lines) == 0 {
		lines = append(lines, l.styles.Muted.Render("  No connections"))
	}

	content := strings.Join(lines, "\n")
	return l.styles.ListPanel.Width(l.width).Height(l.height).Render(content)
}
