package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func P4Command(useFullPath, useJSON bool, args ...string) (string, error) {

	// If useJSON is true, append JSON flags
	if useJSON {
		args = append([]string{"-ztag", "-Mj"}, args...)
	}

	var cmd *exec.Cmd
	if useFullPath {
		p4FullPath := filepath.Join("/home/will/P4GoForge/bin", "p4")
		cmd = exec.Command(p4FullPath, args...)
	} else {
		cmd = exec.Command("p4", args...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("p4 command error: %v, output: %s", err, output)
	}

	return strings.TrimSpace(string(output)), nil
}

func P4dCommand(useFullPath bool, args ...string) (string, error) {
	var cmd *exec.Cmd
	if useFullPath {
		p4FullPath := filepath.Join("/home/will/P4GoForge/bin", "p4d")
		cmd = exec.Command(p4FullPath, args...)
	} else {
		cmd = exec.Command("p4d", args...)
	}

	// Check if running in daemon mode
	isDaemon := false
	for _, arg := range args {
		if arg == "-d" {
			isDaemon = true
			break
		}
	}

	if isDaemon {
		// Start daemon without waiting
		if err := cmd.Start(); err != nil {
			return "", err
		}
		// Return process info or success message
		return fmt.Sprintf("p4d started as daemon with PID %d", cmd.Process.Pid), nil
	} else {
		// Execute command and wait for it to finish
		output, err := cmd.CombinedOutput()
		return strings.TrimSpace(string(output)), err
	}
}
func P4BrokerCommand(useFullPath, useJSON bool, args ...string) (string, error) {

	// If useJSON is true, append JSON flags
	if useJSON {
		args = append([]string{"-ztag", "-Mj"}, args...)
	}

	var cmd *exec.Cmd
	if useFullPath {
		p4FullPath := filepath.Join("/home/will/P4GoForge/bin", "p4")
		cmd = exec.Command(p4FullPath, args...)
	} else {
		cmd = exec.Command("p4", args...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("p4 command error: %v, output: %s", err, output)
	}

	return strings.TrimSpace(string(output)), nil
}
