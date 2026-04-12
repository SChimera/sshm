// theme_unix.go
//go:build !windows

package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/term"
)

// QueryTerminalTheme queries the terminal's ANSI colour palette via OSC 4.
// It reads the 8 slots that map to sshm's Theme fields.
// Returns false if the terminal does not respond within 100ms or doesn't support OSC 4.
// Must be called before Bubbletea takes over the terminal.
func QueryTerminalTheme() (Theme, bool) {
	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return Theme{}, false
	}
	defer f.Close()

	oldState, err := term.MakeRaw(int(f.Fd()))
	if err != nil {
		return Theme{}, false
	}
	defer term.Restore(int(f.Fd()), oldState)

	bySlot := make(map[int]string)
	// Base slots — if any fail, terminal doesn't support OSC 4
	baseSlots := []int{0, 1, 4, 7, 8, 14}
	for _, n := range baseSlots {
		hex, ok := queryOSC4Slot(f, n)
		if !ok {
			return Theme{}, false
		}
		bySlot[n] = hex
	}
	// Extension slots 18/19 (base16-shell) — optional, fall back if unsupported
	for _, n := range []int{18, 19} {
		if hex, ok := queryOSC4Slot(f, n); ok {
			bySlot[n] = hex
		}
	}

	surface := bySlot[18]
	if surface == "" {
		surface = bySlot[0] // fall back to Background
	}
	border := bySlot[19]
	if border == "" {
		border = bySlot[8] // fall back to Muted
	}

	return Theme{
		Background: bySlot[0],
		Danger:     bySlot[1],
		Primary:    bySlot[4],
		Foreground: bySlot[7],
		Muted:      bySlot[8],
		Accent:     bySlot[14],
		Surface:    surface,
		Border:     border,
	}, true
}

// osc4Re matches OSC 4 colour responses in both 2-digit and 4-digit hex formats.
var osc4Re = regexp.MustCompile(`\033\]4;\d+;rgb:([0-9a-fA-F]{2})[0-9a-fA-F]{0,2}/([0-9a-fA-F]{2})[0-9a-fA-F]{0,2}/([0-9a-fA-F]{2})[0-9a-fA-F]{0,2}`)

func parseOSC4(resp string) (string, bool) {
	m := osc4Re.FindStringSubmatch(resp)
	if m == nil {
		return "", false
	}
	return "#" + strings.ToLower(m[1]+m[2]+m[3]), true
}

func queryOSC4Slot(f *os.File, n int) (string, bool) {
	fmt.Fprintf(f, "\033]4;%d;?\033\\", n)

	ch := make(chan string, 1)
	go func() {
		buf := make([]byte, 128)
		nr, _ := f.Read(buf)
		if nr > 0 {
			ch <- string(buf[:nr])
		} else {
			ch <- ""
		}
	}()

	select {
	case resp := <-ch:
		if resp == "" {
			return "", false
		}
		return parseOSC4(resp)
	case <-time.After(100 * time.Millisecond):
		return "", false
	}
}
