// detail.go
package main

import (
	"fmt"
	"strings"
)

// DetailPanel renders connection info in the right panel.
type DetailPanel struct {
	styles  Styles
	conn    *Connection // nil if no connection selected
	group   *Group      // the group the connection belongs to
	focused bool
	width   int
	height  int
}

// NewDetailPanel constructs a DetailPanel with the given styles.
func NewDetailPanel(styles Styles) DetailPanel {
	return DetailPanel{styles: styles}
}

// View renders the detail panel.
func (d DetailPanel) View() string {
	if d.conn == nil {
		if d.group != nil {
			return d.viewGroup()
		}
		return d.styles.Muted.Render("\n  Select a connection")
	}
	return d.viewConnection()
}

func (d DetailPanel) viewGroup() string {
	name := d.group.Name
	if name == "" {
		name = "Ungrouped"
	}
	lines := []string{
		d.styles.Title.Render(name),
		"",
		d.styles.Muted.Render(fmt.Sprintf("%d connection(s)", len(d.group.Connections))),
		"",
		d.styles.Muted.Render("Press enter on a connection to connect"),
		d.styles.Muted.Render("Press n to add a connection to this group"),
	}
	return d.styles.DetailPanel.Width(d.width).Render(strings.Join(lines, "\n"))
}

func (d DetailPanel) viewConnection() string {
	c := d.conn
	var lines []string

	lines = append(lines, d.styles.Title.Render(c.Host))
	lines = append(lines, "")

	row := func(label, value string) string {
		if value == "" {
			value = d.styles.Muted.Render("—")
		}
		return d.styles.Label.Render(label) + d.styles.Value.Render(value)
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
		lines = append(lines, d.styles.Muted.Render("Extra:"))
		lines = append(lines, d.styles.Muted.Render(c.Extra))
	}

	lines = append(lines, "")
	lines = append(lines, d.styles.Muted.Render("enter:connect  e:edit  d:delete"))

	return d.styles.DetailPanel.Width(d.width).Height(d.height).Render(strings.Join(lines, "\n"))
}
