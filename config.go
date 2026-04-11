// config.go
package main

import (
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
