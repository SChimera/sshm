// detail.go
package main

import (
	"fmt"
	"strings"
)

// DetailPanel renders connection info in the right panel.
type DetailPanel struct {
	conn    *Connection // nil if no connection selected
	group   *Group      // the group the connection belongs to
	focused bool
	width   int
	height  int
}

// View renders the detail panel.
func (d DetailPanel) View() string {
	if d.conn == nil {
		if d.group != nil {
			return d.viewGroup()
		}
		return StyleMuted.Render("\n  Select a connection")
	}
	return d.viewConnection()
}

func (d DetailPanel) viewGroup() string {
	name := d.group.Name
	if name == "" {
		name = "Ungrouped"
	}
	lines := []string{
		StyleTitle.Render(name),
		"",
		StyleMuted.Render(fmt.Sprintf("%d connection(s)", len(d.group.Connections))),
		"",
		StyleMuted.Render("Press enter on a connection to connect"),
		StyleMuted.Render("Press n to add a connection to this group"),
	}
	return StyleDetailPanel.Width(d.width).Render(strings.Join(lines, "\n"))
}

func (d DetailPanel) viewConnection() string {
	c := d.conn
	var lines []string

	lines = append(lines, StyleTitle.Render(c.Host))
	lines = append(lines, "")

	row := func(label, value string) string {
		if value == "" {
			value = StyleMuted.Render("—")
		}
		return StyleLabel.Render(label) + StyleValue.Render(value)
	}

	lines = append(lines, row("HostName", c.HostName))
	lines = append(lines, row("User", c.User))
	lines = append(lines, row("Port", c.Port))
	lines = append(lines, row("IdentityFile", c.IdentityFile))
	lines = append(lines, row("ProxyJump", c.ProxyJump))
	lines = append(lines, row("ForwardAgent", c.ForwardAgent))
	lines = append(lines, row("LocalForward", c.LocalForward))
	lines = append(lines, row("RemoteForward", c.RemoteForward))

	if c.Extra != "" {
		lines = append(lines, "")
		lines = append(lines, StyleMuted.Render("Extra:"))
		lines = append(lines, StyleMuted.Render(c.Extra))
	}

	lines = append(lines, "")
	lines = append(lines, StyleMuted.Render("enter:connect  e:edit  d:delete"))

	return StyleDetailPanel.Width(d.width).Height(d.height).Render(strings.Join(lines, "\n"))
}
