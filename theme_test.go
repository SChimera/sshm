// theme_test.go
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultTheme(t *testing.T) {
	th := DefaultTheme()
	fields := map[string]string{
		"Background": th.Background,
		"Surface":    th.Surface,
		"Muted":      th.Muted,
		"Foreground": th.Foreground,
		"Primary":    th.Primary,
		"Accent":     th.Accent,
		"Danger":     th.Danger,
		"Border":     th.Border,
	}
	for name, val := range fields {
		if val == "" {
			t.Errorf("DefaultTheme.%s is empty", name)
		}
		if len(val) != 7 || val[0] != '#' {
			t.Errorf("DefaultTheme.%s = %q: want 7-char hex like #rrggbb", name, val)
		}
	}
}

func TestMergeTheme_DstWins(t *testing.T) {
	dst := Theme{Primary: "#ff0000"}
	src := Theme{Primary: "#0000ff", Background: "#000000"}
	got := mergeTheme(dst, src)
	if got.Primary != "#ff0000" {
		t.Errorf("Primary: got %q, dst value should win", got.Primary)
	}
	if got.Background != "#000000" {
		t.Errorf("Background: got %q, want %q (filled from src)", got.Background, "#000000")
	}
}

func TestMergeTheme_AllEmpty(t *testing.T) {
	dst := Theme{}
	src := DefaultTheme()
	got := mergeTheme(dst, src)
	if got != src {
		t.Errorf("empty dst should equal src after merge")
	}
}

func TestLoadConfigTheme_AllFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `[theme]
background = "#000001"
surface    = "#000002"
muted      = "#000003"
foreground = "#000004"
primary    = "#000005"
accent     = "#000006"
danger     = "#000007"
border     = "#000008"
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	th, err := LoadConfigTheme(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if th.Background != "#000001" {
		t.Errorf("Background: got %q want #000001", th.Background)
	}
	if th.Surface != "#000002" {
		t.Errorf("Surface: got %q want #000002", th.Surface)
	}
	if th.Muted != "#000003" {
		t.Errorf("Muted: got %q want #000003", th.Muted)
	}
	if th.Foreground != "#000004" {
		t.Errorf("Foreground: got %q want #000004", th.Foreground)
	}
	if th.Primary != "#000005" {
		t.Errorf("Primary: got %q want #000005", th.Primary)
	}
	if th.Accent != "#000006" {
		t.Errorf("Accent: got %q want #000006", th.Accent)
	}
	if th.Danger != "#000007" {
		t.Errorf("Danger: got %q want #000007", th.Danger)
	}
	if th.Border != "#000008" {
		t.Errorf("Border: got %q want #000008", th.Border)
	}
}

func TestLoadConfigTheme_PartialFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `[theme]
primary = "#ff0000"
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	th, err := LoadConfigTheme(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if th.Primary != "#ff0000" {
		t.Errorf("Primary: got %q want #ff0000", th.Primary)
	}
	if th.Background != "" {
		t.Errorf("Background should be empty when not set, got %q", th.Background)
	}
}

func TestLoadConfigTheme_MissingFile(t *testing.T) {
	_, err := LoadConfigTheme("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestResolveTheme_FallsBackToDefault(t *testing.T) {
	// With no config file reachable and OSC4 returning false (on Windows CI),
	// ResolveTheme should return the default theme values.
	th, err := ResolveTheme()
	if err != nil {
		t.Logf("ResolveTheme returned non-fatal config error: %v", err)
	}
	// All fields should be non-empty (filled by default at minimum)
	if th.Background == "" {
		t.Error("Background should not be empty after ResolveTheme")
	}
	// On this machine, OSC4 returns false (Windows), so config file determines result.
	// We can't assert exact values without controlling the env, but we can assert
	// that the result is fully populated.
	fields := []string{th.Background, th.Surface, th.Muted, th.Foreground,
		th.Primary, th.Accent, th.Danger, th.Border}
	for i, f := range fields {
		if f == "" {
			t.Errorf("field %d is empty after ResolveTheme", i)
		}
	}
}
