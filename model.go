// model.go
package main

// Connection represents a single Host block in ~/.ssh/config.
// Fields map directly to SSH config directives.
// Extra holds any unrecognised directives verbatim.
type Connection struct {
	Host          string
	HostName      string
	User          string
	Port          string
	IdentityFile  string
	ProxyJump     string
	ForwardAgent  string
	LocalForward  string
	RemoteForward string
	Extra         string // unrecognised directives, preserved verbatim
}

// Group is a named collection of connections.
// Name == "" identifies the implicit Ungrouped group.
type Group struct {
	Name        string
	Connections []Connection
}

// ParsedConfig is the in-memory representation of ~/.ssh/config.
type ParsedConfig struct {
	Preamble  string   // raw text before the first Host or Group marker
	Groups    []Group  // named groups first, Ungrouped (Name=="") last
	RawBlocks []string // wildcard Host blocks (Host *), preserved verbatim
}
