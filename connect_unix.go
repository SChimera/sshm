// connect_unix.go
//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Connect replaces the current process with an SSH session to host.
// The TUI exits cleanly and the terminal is handed to SSH.
func Connect(host string) error {
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found in PATH: %w", err)
	}
	return syscall.Exec(sshPath, []string{"ssh", host}, os.Environ())
}
