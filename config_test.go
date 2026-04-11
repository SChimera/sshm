// config_test.go
package main

import (
	"strings"
	"testing"
)

func TestParseEmpty(t *testing.T) {
	cfg := ParseConfig("")
	if len(cfg.Groups) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(cfg.Groups))
	}
}

func TestParseUngrouped(t *testing.T) {
	input := `Host web
    HostName 192.168.1.1
    User ubuntu
    Port 22
`
	cfg := ParseConfig(input)
	if len(cfg.Groups) != 1 {
		t.Fatalf("expected 1 group (ungrouped), got %d", len(cfg.Groups))
	}
	g := cfg.Groups[0]
	if g.Name != "" {
		t.Fatalf("expected ungrouped (Name==%q), got %q", "", g.Name)
	}
	if len(g.Connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(g.Connections))
	}
	c := g.Connections[0]
	if c.Host != "web" {
		t.Errorf("expected Host=web, got %q", c.Host)
	}
	if c.HostName != "192.168.1.1" {
		t.Errorf("expected HostName=192.168.1.1, got %q", c.HostName)
	}
	if c.User != "ubuntu" {
		t.Errorf("expected User=ubuntu, got %q", c.User)
	}
	if c.Port != "22" {
		t.Errorf("expected Port=22, got %q", c.Port)
	}
}

func TestParseGroups(t *testing.T) {
	input := `# Group: Production
Host web-prod
    HostName 10.0.0.1
    User ubuntu

Host db-prod
    HostName 10.0.0.2

# Group: Dev
Host dev-box
    HostName 10.0.0.3
`
	cfg := ParseConfig(input)
	if len(cfg.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(cfg.Groups))
	}
	if cfg.Groups[0].Name != "Production" {
		t.Errorf("expected Production, got %q", cfg.Groups[0].Name)
	}
	if len(cfg.Groups[0].Connections) != 2 {
		t.Errorf("expected 2 connections in Production, got %d", len(cfg.Groups[0].Connections))
	}
	if cfg.Groups[1].Name != "Dev" {
		t.Errorf("expected Dev, got %q", cfg.Groups[1].Name)
	}
	if len(cfg.Groups[1].Connections) != 1 {
		t.Errorf("expected 1 connection in Dev, got %d", len(cfg.Groups[1].Connections))
	}
}

func TestParsePreamble(t *testing.T) {
	input := `# my ssh config
# managed by sshm

Host web
    HostName 10.0.0.1
`
	cfg := ParseConfig(input)
	if cfg.Preamble == "" {
		t.Fatal("expected non-empty preamble")
	}
}

func TestParseWildcardSkipped(t *testing.T) {
	input := `Host *
    ServerAliveInterval 60

Host web
    HostName 10.0.0.1
`
	cfg := ParseConfig(input)
	if len(cfg.RawBlocks) != 1 {
		t.Fatalf("expected 1 raw block, got %d", len(cfg.RawBlocks))
	}
	if len(cfg.Groups) != 1 || len(cfg.Groups[0].Connections) != 1 {
		t.Fatalf("expected 1 ungrouped connection, got groups=%v", cfg.Groups)
	}
}

func TestParseExtendedFields(t *testing.T) {
	input := `Host jump
    HostName bastion.example.com
    User ec2-user
    IdentityFile ~/.ssh/bastion.pem
    ProxyJump none
    ForwardAgent yes
    LocalForward 8080 localhost:8080
    RemoteForward 9090 localhost:9090
`
	cfg := ParseConfig(input)
	c := cfg.Groups[0].Connections[0]
	if c.IdentityFile != "~/.ssh/bastion.pem" {
		t.Errorf("IdentityFile: got %q", c.IdentityFile)
	}
	if c.ProxyJump != "none" {
		t.Errorf("ProxyJump: got %q", c.ProxyJump)
	}
	if c.ForwardAgent != "yes" {
		t.Errorf("ForwardAgent: got %q", c.ForwardAgent)
	}
	if c.LocalForward != "8080 localhost:8080" {
		t.Errorf("LocalForward: got %q", c.LocalForward)
	}
	if c.RemoteForward != "9090 localhost:9090" {
		t.Errorf("RemoteForward: got %q", c.RemoteForward)
	}
}

func TestWriteEmpty(t *testing.T) {
	cfg := ParsedConfig{}
	out := WriteConfig(cfg)
	if out != "" {
		t.Errorf("expected empty output, got %q", out)
	}
}

func TestWriteGrouped(t *testing.T) {
	cfg := ParsedConfig{
		Groups: []Group{
			{
				Name: "Production",
				Connections: []Connection{
					{Host: "web", HostName: "10.0.0.1", User: "ubuntu", Port: "22"},
				},
			},
		},
	}
	out := WriteConfig(cfg)
	if !strings.Contains(out, "# Group: Production") {
		t.Errorf("missing group marker, got:\n%s", out)
	}
	if !strings.Contains(out, "Host web") {
		t.Errorf("missing Host line, got:\n%s", out)
	}
	if !strings.Contains(out, "HostName 10.0.0.1") {
		t.Errorf("missing HostName, got:\n%s", out)
	}
}

func TestWriteUngroupedLast(t *testing.T) {
	cfg := ParsedConfig{
		Groups: []Group{
			{Name: "Prod", Connections: []Connection{{Host: "prod", HostName: "1.1.1.1"}}},
			{Name: "", Connections: []Connection{{Host: "lone", HostName: "2.2.2.2"}}},
		},
	}
	out := WriteConfig(cfg)
	prodIdx := strings.Index(out, "# Group: Prod")
	loneIdx := strings.Index(out, "Host lone")
	if prodIdx == -1 || loneIdx == -1 {
		t.Fatalf("missing expected content:\n%s", out)
	}
	if loneIdx < prodIdx {
		t.Errorf("ungrouped connection 'lone' should appear after named group 'Prod'")
	}
}

func TestRoundTrip(t *testing.T) {
	input := `# Group: Production
Host web-prod
    HostName 10.0.0.1
    User ubuntu
    Port 22

Host db-prod
    HostName 10.0.0.2
    User postgres

# Group: Dev
Host dev-box
    HostName 10.0.0.3
    User admin
`
	cfg := ParseConfig(input)
	out := WriteConfig(cfg)
	cfg2 := ParseConfig(out)

	if len(cfg.Groups) != len(cfg2.Groups) {
		t.Fatalf("group count mismatch: %d vs %d", len(cfg.Groups), len(cfg2.Groups))
	}
	for i := range cfg.Groups {
		if cfg.Groups[i].Name != cfg2.Groups[i].Name {
			t.Errorf("group %d name mismatch: %q vs %q", i, cfg.Groups[i].Name, cfg2.Groups[i].Name)
		}
		if len(cfg.Groups[i].Connections) != len(cfg2.Groups[i].Connections) {
			t.Errorf("group %d connection count mismatch", i)
		}
		for j := range cfg.Groups[i].Connections {
			if j >= len(cfg2.Groups[i].Connections) {
				break
			}
			got := cfg2.Groups[i].Connections[j]
			want := cfg.Groups[i].Connections[j]
			if got.Host != want.Host || got.HostName != want.HostName || got.User != want.User || got.Port != want.Port {
				t.Errorf("group %d conn %d: got %+v, want %+v", i, j, got, want)
			}
		}
	}
}

func TestWritePreamble(t *testing.T) {
	cfg := ParsedConfig{
		Preamble: "# managed by sshm",
		Groups:   []Group{{Name: "Dev", Connections: []Connection{{Host: "box", HostName: "1.1.1.1"}}}},
	}
	out := WriteConfig(cfg)
	if !strings.HasPrefix(out, "# managed by sshm") {
		t.Errorf("preamble should be first, got:\n%s", out)
	}
}

func TestWriteRawBlocksLast(t *testing.T) {
	cfg := ParsedConfig{
		RawBlocks: []string{"Host *\n    ServerAliveInterval 60"},
		Groups:    []Group{{Name: "Dev", Connections: []Connection{{Host: "box", HostName: "1.1.1.1"}}}},
	}
	out := WriteConfig(cfg)
	boxIdx := strings.Index(out, "Host box")
	wildcardIdx := strings.Index(out, "Host *")
	if wildcardIdx < boxIdx {
		t.Errorf("wildcard block should appear after managed connections")
	}
}
