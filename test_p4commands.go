package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func P4Command(args ...string) (string, error) {

	var cmd *exec.Cmd

	cmd = exec.Command("p4", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("p4 command error: %v, output: %s", err, output)
	}

	return strings.TrimSpace(string(output)), nil
}

func P4Commands(p4Test *P4Test, commands []string, configType P4ConfigType) (string, error) {
	var totalOutput strings.Builder

	// Check if the broker should be loaded for this test
	if configType == BrokerP4Config && !loadBroker {
		return colorRed + "The broker is not enabled for this test." + colorReset, nil
	}

	// Set P4 configuration
	setP4Config(configType, p4Test)

	for _, cmdString := range commands {
		args := strings.Fields(cmdString) // Split the command string into arguments

		// Create the command
		cmd := exec.Command("p4", args...)

		// Log the command being executed
		totalOutput.WriteString(coloredOutput(colorGreen, "-- Running command:"+colorCyan+" p4 "+cmdString+"\n"))

		// Execute the command and get output
		output, err := cmd.CombinedOutput()
		if err != nil {
			return totalOutput.String(), fmt.Errorf("p4 command error: %v, output: %s", err, output)
		}

		// Append the command output to the total output
		formattedOutput := strings.TrimSpace(string(output))
		totalOutput.WriteString(formattedOutput + "\n\n") // Adding extra newline for separation
	}

	// Optionally reset any configuration changes here, if necessary

	return totalOutput.String(), nil
}

func P4dCommand(useFullPath bool, args ...string) (string, error) {
	var cmd *exec.Cmd
	if useFullPath {
		p4FullPath := filepath.Join(binDir, "p4d")
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
