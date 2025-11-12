package main

import (
	"fmt"
	"os"
	"os/exec"
)

// runCommand executes a command and returns an error if it fails
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s %v - %w", name, args, err)
	}

	return nil
}
