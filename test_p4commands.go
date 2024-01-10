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
