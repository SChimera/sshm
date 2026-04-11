// connect_windows.go
//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Connect runs SSH as a child process with stdio wired to the terminal.
// The TUI is suspended for the duration; the app exits when SSH ends.
func Connect(host string) error {
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found in PATH: %w", err)
	}
	cmd := exec.Command(sshPath, host)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
