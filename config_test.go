// config_test.go
package main

import (
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
