// config.go
package main

import (
	"os"
	"path/filepath"
	"strings"
)

// ParseConfig parses the content of an ~/.ssh/config file into a ParsedConfig.
func ParseConfig(content string) ParsedConfig {
	var cfg ParsedConfig
	var currentGroup *Group // nil = ungrouped
	var currentConn *Connection
	var preambleLines []string
	var currentRaw []string // accumulates lines for a wildcard block
	inWildcard := false
	seenFirstBlock := false

	flushConn := func() {
		if currentConn == nil {
			return
		}
		if currentGroup != nil {
			currentGroup.Connections = append(currentGroup.Connections, *currentConn)
			// refresh pointer after append
			currentGroup = &cfg.Groups[len(cfg.Groups)-1]
		} else {
			// ungrouped: find or create the ungrouped group at end
			idx := ungroupedIndex(&cfg)
			cfg.Groups[idx].Connections = append(cfg.Groups[idx].Connections, *currentConn)
		}
		currentConn = nil
	}

	flushWildcard := func() {
		if !inWildcard || len(currentRaw) == 0 {
			return
		}
		cfg.RawBlocks = append(cfg.RawBlocks, strings.Join(currentRaw, "\n"))
		currentRaw = nil
		inWildcard = false
	}

	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Group marker
		if strings.HasPrefix(trimmed, "# Group:") {
			flushConn()
			flushWildcard()
			seenFirstBlock = true
			name := strings.TrimSpace(strings.TrimPrefix(trimmed, "# Group:"))
			cfg.Groups = append(cfg.Groups, Group{Name: name})
			currentGroup = &cfg.Groups[len(cfg.Groups)-1]
			continue
		}

		// Host line (must start at column 0)
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") &&
			strings.HasPrefix(strings.ToLower(trimmed), "host ") {
			flushConn()
			flushWildcard()
			seenFirstBlock = true
			alias := strings.TrimSpace(trimmed[5:]) // skip "host "

			// Wildcard host — preserve verbatim
			if strings.ContainsAny(alias, "*?") {
				inWildcard = true
				currentRaw = []string{line}
				continue
			}

			currentConn = &Connection{Host: alias}
			continue
		}

		// Continuation of a wildcard block
		if inWildcard {
			if trimmed == "" {
				flushWildcard()
			} else {
				currentRaw = append(currentRaw, line)
			}
			continue
		}

		// Directive inside a managed Host block
		if currentConn != nil {
			if trimmed != "" {
				applyDirective(trimmed, currentConn)
			}
			continue
		}

		// Preamble (before any Host or Group marker)
		if !seenFirstBlock {
			preambleLines = append(preambleLines, line)
		}
	}

	flushConn()
	flushWildcard()

	cfg.Preamble = strings.TrimRight(strings.Join(preambleLines, "\n"), "\n")
	return cfg
}

// ungroupedIndex returns the index of the Ungrouped group (Name==""),
// creating it if it doesn't exist.
func ungroupedIndex(cfg *ParsedConfig) int {
	for i, g := range cfg.Groups {
		if g.Name == "" {
			return i
		}
	}
	cfg.Groups = append(cfg.Groups, Group{Name: ""})
	return len(cfg.Groups) - 1
}

// applyDirective parses a single SSH config directive line and sets the
// corresponding field on conn. Unrecognised directives go into Extra.
func applyDirective(trimmed string, conn *Connection) {
	// Normalize "Key=Value" and "Key = Value" to "Key Value"
	if idx := strings.IndexAny(trimmed, "= "); idx > 0 {
		key := strings.TrimSpace(trimmed[:idx])
		value := strings.TrimSpace(trimmed[idx+1:])
		trimmed = key + " " + strings.TrimSpace(value)
	}
	parts := strings.SplitN(trimmed, " ", 2)
	if len(parts) != 2 {
		appendExtra(conn, trimmed)
		return
	}
	key, value := strings.ToLower(strings.TrimSpace(parts[0])), strings.TrimSpace(parts[1])
	switch key {
	case "hostname":
		conn.HostName = value
	case "user":
		conn.User = value
	case "port":
		conn.Port = value
	case "identityfile":
		conn.IdentityFile = value
	case "proxyjump":
		conn.ProxyJump = value
	case "forwardagent":
		conn.ForwardAgent = value
	case "localforward":
		conn.LocalForward = value
	case "remoteforward":
		conn.RemoteForward = value
	default:
		appendExtra(conn, "    "+trimmed)
	}
}

func appendExtra(conn *Connection, line string) {
	if conn.Extra != "" {
		conn.Extra += "\n"
	}
	conn.Extra += line
}

// WriteConfig serialises a ParsedConfig back to ~/.ssh/config format.
// Order: preamble, named groups, ungrouped connections, raw wildcard blocks.
func WriteConfig(cfg ParsedConfig) string {
	var sb strings.Builder

	if cfg.Preamble != "" {
		sb.WriteString(cfg.Preamble)
		sb.WriteString("\n\n")
	}

	// Named groups first
	for _, g := range cfg.Groups {
		if g.Name == "" {
			continue
		}
		sb.WriteString("# Group: ")
		sb.WriteString(g.Name)
		sb.WriteString("\n")
		for _, c := range g.Connections {
			writeConnection(&sb, c)
		}
	}

	// Ungrouped connections (Name == "")
	for _, g := range cfg.Groups {
		if g.Name != "" {
			continue
		}
		for _, c := range g.Connections {
			writeConnection(&sb, c)
		}
	}

	// Preserved wildcard blocks at end
	for _, block := range cfg.RawBlocks {
		sb.WriteString(block)
		sb.WriteString("\n\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

func writeConnection(sb *strings.Builder, c Connection) {
	sb.WriteString("Host ")
	sb.WriteString(c.Host)
	sb.WriteString("\n")
	writeField(sb, "HostName", c.HostName)
	writeField(sb, "User", c.User)
	writeField(sb, "Port", c.Port)
	writeField(sb, "IdentityFile", c.IdentityFile)
	writeField(sb, "ProxyJump", c.ProxyJump)
	writeField(sb, "ForwardAgent", c.ForwardAgent)
	writeField(sb, "LocalForward", c.LocalForward)
	writeField(sb, "RemoteForward", c.RemoteForward)
	if c.Extra != "" {
		sb.WriteString(c.Extra)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

func writeField(sb *strings.Builder, key, value string) {
	if value == "" {
		return
	}
	sb.WriteString("    ")
	sb.WriteString(key)
	sb.WriteString(" ")
	sb.WriteString(value)
	sb.WriteString("\n")
}

// ConfigPath returns the path to ~/.ssh/config on all platforms.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh", "config"), nil
}

// Load reads the SSH config file at path and returns a ParsedConfig.
// If the file does not exist, an empty ParsedConfig is returned without error.
func Load(path string) (ParsedConfig, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ParsedConfig{}, nil
	}
	if err != nil {
		return ParsedConfig{}, err
	}
	return ParseConfig(string(data)), nil
}

// Save writes cfg back to path, creating the file (and ~/.ssh directory) if needed.
// File permissions: 0600 (owner read/write only). Directory permissions: 0700.
func Save(path string, cfg ParsedConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	content := WriteConfig(cfg)
	if content != "" {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0600)
}
