package main

import (
	"bufio"
	"fmt"
	"os"
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

func createUser(p4Test *P4Test, username string) error {
	fmt.Println("Creating user:", username)

	// Define the path for the temporary user specification file
	tempUserSpecFile := filepath.Join(p4Test.clientRoot, "temp_user_spec.txt")

	// Construct the command to execute 'p4 user -o' and redirect output to tempUserSpecFile
	cmdString := fmt.Sprintf("p4 user -o %s > %s", username, tempUserSpecFile)
	cmd := exec.Command("bash", "-c", cmdString)

	// Execute the command
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to execute 'p4 user -o': %v", err)
	}

	// Read contents of the tempUserSpecFile
	userSpecData, err := os.ReadFile(tempUserSpecFile)
	if err != nil {
		return fmt.Errorf("failed to read user spec file: %v", err)
	}

	// Modify the user specification as needed
	userSpec := string(userSpecData)
	//userSpec = strings.Replace(userSpec, "Email:", "Email: email@generated_test", 1)
	//userSpec = strings.Replace(userSpec, "User:", "User: "+username, -1)

	// Rewrite the modified user specification back to the temporary file
	if err := os.WriteFile(tempUserSpecFile, []byte(userSpec), 0644); err != nil {
		return err
	}

	// Create a command to execute 'p4 user -f -i' using the modified tempUserSpecFile
	cmdString = fmt.Sprintf("p4 user -f -i < %s", tempUserSpecFile)
	cmd = exec.Command("bash", "-c", cmdString)

	// Execute the command
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create user with 'p4 user -f -i': %v", err)
	}

	// Clean up temporary file
	if err := os.Remove(tempUserSpecFile); err != nil {
		return err
	}
	fmt.Println("User created:", username)
	return nil
}

func createWorkspace(p4Test *P4Test, username string) error {
	userDir := filepath.Join(p4Test.clientRoot, username)
	workspaceName := username + "_ws"
	tempSpecFile := filepath.Join(p4Test.clientRoot, "temp_workspace_spec.txt")

	// 1. Create user directory
	//if err := os.MkdirAll(userDir, os.ModePerm); err != nil {
	if err := os.MkdirAll(userDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create user directory: %v", err)
	}
	//os.Chdir(userDir)

	// 2. Construct the command to execute 'p4 workspace -o' and redirect output to tempSpecFile
	//cmdString := fmt.Sprintf("p4 -u %s workspace -o %s > %s", username, workspaceName, tempSpecFile)
	cmdString := fmt.Sprintf("p4 workspace -o %s > %s", workspaceName, tempSpecFile)

	cmd := exec.Command("bash", "-c", cmdString)

	// Execute the command
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to execute 'p4 workspace -o': %v", err)
	}

	// 3. Read contents of the tempSpecFile
	wsSpecData, err := os.ReadFile(tempSpecFile)
	if err != nil {
		return fmt.Errorf("failed to read workspace spec file: %v", err)
	}

	// 4. Modify the workspace specification as needed

	// 4. Modify the workspace specification as needed
	var modifiedLines []string
	scanner := bufio.NewScanner(strings.NewReader(string(wsSpecData)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Root:") {
			modifiedLines = append(modifiedLines, "Root: "+userDir)
		} else if strings.HasPrefix(line, "Owner:") {
			modifiedLines = append(modifiedLines, "Owner: "+username)
		} else {
			modifiedLines = append(modifiedLines, line)
		}
	}
	if scanner.Err() != nil {
		return fmt.Errorf("error reading workspace spec: %v", scanner.Err())
	}

	modifiedSpec := strings.Join(modifiedLines, "\n")

	// Rewrite the modified workspace specification back to the temporary file
	if err := os.WriteFile(tempSpecFile, []byte(modifiedSpec), 0644); err != nil {
		return err
	}

	// 5. Create a command to execute 'p4 workspace -i' using the modified tempSpecFile
	cmdString = fmt.Sprintf("p4 workspace -i < %s", tempSpecFile)
	cmd = exec.Command("bash", "-c", cmdString)

	// Execute the command
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create workspace with 'p4 workspace -i': %v", err)
	}

	// 6. Clean up temporary file
	if err := os.Remove(tempSpecFile); err != nil {
		return err
	}

	//os.Chdir(p4t.clientRoot)
	fmt.Println("Workspace created:", workspaceName)
	return nil
}

func createGroup(p4Test *P4Test, groupName string, users []string) error {
	tempGroupSpecFile := filepath.Join(p4Test.clientRoot, "temp_group_spec.txt")

	// Generate the group specification and write it to a temporary file
	cmdString := fmt.Sprintf("p4 group -o %s > %s", groupName, tempGroupSpecFile)
	cmd := exec.Command("bash", "-c", cmdString)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to execute 'p4 group -o': %v", err)
	}

	// Read and modify the group specification
	groupSpecData, err := os.ReadFile(tempGroupSpecFile)
	if err != nil {
		return fmt.Errorf("failed to read group spec file: %v", err)
	}
	groupSpec := string(groupSpecData)
	var modifiedLines []string
	scanner := bufio.NewScanner(strings.NewReader(groupSpec))
	usersSectionFound := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Users:") {
			usersSectionFound = true
			modifiedLines = append(modifiedLines, "Users:")
			for _, user := range users {
				modifiedLines = append(modifiedLines, "\t"+user)
			}
		} else if !usersSectionFound || (usersSectionFound && !strings.HasPrefix(line, "\t")) {
			modifiedLines = append(modifiedLines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading group spec: %v", err)
	}
	modifiedGroupSpec := strings.Join(modifiedLines, "\n")

	// Write the modified group specification back to the temporary file
	if err := os.WriteFile(tempGroupSpecFile, []byte(modifiedGroupSpec), 0644); err != nil {
		return err
	}

	// Execute 'p4 group -i' using the modified file
	cmdString = fmt.Sprintf("p4 group -i < %s", tempGroupSpecFile)
	cmd = exec.Command("bash", "-c", cmdString)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create group with 'p4 group -i': %v", err)
	}

	// Clean up temporary file
	if err := os.Remove(tempGroupSpecFile); err != nil {
		return err
	}
	fmt.Println("Group created:", groupName)
	return nil
}

func createDepotFiles(p4Test *P4Test, username string, files map[string]string) error {
	workspaceName := username + "_ws"
	clientRoot := filepath.Join(p4Test.clientRoot, username)
	os.Setenv("P4USER", username)

	// Ensure the current workspace is set correctly
	cmd := exec.Command("p4", "set", "P4CLIENT="+workspaceName)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set P4CLIENT: %v", err)
	}

	// Create and add files with specified types
	for fileName, fileType := range files {
		filePath := filepath.Join(clientRoot, fileName)
		if err := os.WriteFile(filePath, []byte("Initial content"), 0777); err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}

		// Add file to Perforce with specified file type
		cmd = exec.Command("p4", "add", "-t", fileType, filePath)
		if _, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add file to Perforce: %v", err)
		}
	}

	// Submit the files
	submitDesc := "Initial submit of files with specific types"
	cmd = exec.Command("p4", "submit", "-d", submitDesc)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to submit files: %v", err)
	}

	fmt.Println("Files with specific types submitted to the depot")
	return nil
}
func checkoutAndLockFiles(p4Test *P4Test, username string, filesToLock []string) error {
	// Create a changelist for the user
	changelistNumber, err := createChangelist(p4Test, username)
	if err != nil {
		return fmt.Errorf("failed to create changelist for user %s: %v", username, err)
	}

	// Checkout and lock files
	for _, file := range filesToLock {
		if err := p4EditFile(file, changelistNumber); err != nil {
			return fmt.Errorf("failed to check out file %s for user %s: %v", file, username, err)
		}

		if err := p4LockFile(file); err != nil {
			return fmt.Errorf("failed to lock file %s for user %s: %v", file, username, err)
		}
	}

	return nil
}
func syncUserWorkspace(p4Test *P4Test, username string) error {
	workspaceName := username + "_ws"

	// Set the P4CLIENT environment variable to the user's workspace
	os.Setenv("P4CLIENT", workspaceName)
	//os.Setenv("P4USER", username)
	// Sync the workspace
	cmd := exec.Command("p4", "sync")
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to sync workspace for user %s: %v", username, err)
	}

	fmt.Printf("Workspace synced for user %s\n", username)
	return nil
}

func createChangelist(p4Test *P4Test, username string) (string, error) {
	// Set the P4USER environment variable
	os.Setenv("P4USER", username)

	// Generate a path for the temporary changelist specification file
	tempChangelistSpecFile := filepath.Join(p4Test.clientRoot, "temp_changelist_spec.txt")

	// Create a new changelist and write it to the temporary file
	cmdString := fmt.Sprintf("p4 change -o > %s", tempChangelistSpecFile)
	cmd := exec.Command("bash", "-c", cmdString)
	if _, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to output changelist form: %v", err)
	}

	// Read contents of the tempChangelistSpecFile
	clSpecData, err := os.ReadFile(tempChangelistSpecFile)
	if err != nil {
		return "", fmt.Errorf("failed to read changelist spec file: %v", err)
	}

	// Modify the changelist specification
	var modifiedLines []string
	scanner := bufio.NewScanner(strings.NewReader(string(clSpecData)))
	descriptionFound := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Description:") {
			descriptionFound = true
			modifiedLines = append(modifiedLines, "Description:")
			modifiedLines = append(modifiedLines, "\tChange created by "+username)
		} else if !descriptionFound || (descriptionFound && !strings.HasPrefix(line, "\t")) {
			modifiedLines = append(modifiedLines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading changelist spec: %v", err)
	}
	modifiedSpec := strings.Join(modifiedLines, "\n")

	// Rewrite the modified changelist specification back to the temporary file
	if err := os.WriteFile(tempChangelistSpecFile, []byte(modifiedSpec), 0644); err != nil {
		return "", err
	}

	// Submit the modified changelist
	cmdString = fmt.Sprintf("p4 change -i < %s", tempChangelistSpecFile)
	cmd = exec.Command("bash", "-c", cmdString)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to submit modified changelist: %v", err)
	}

	// Extract changelist number from the output
	changelistNumber, err := extractChangelistNumber(string(output))
	if err != nil {
		return "", err
	}

	// Clean up the temporary file
	if err := os.Remove(tempChangelistSpecFile); err != nil {
		return "", err
	}

	return changelistNumber, nil
}

// Helper function to extract changelist number
func extractChangelistNumber(output string) (string, error) {
	// Split the output by spaces
	parts := strings.Fields(output)

	// Look for the word "Change" and then extract the following number
	for i, part := range parts {
		if part == "Change" && i+1 < len(parts) {
			// Assuming the number is the next part
			return parts[i+1], nil
		}
	}

	// If no number found, return an error
	return "", fmt.Errorf("changelist number not found in output")
}

func p4EditFile(file string, changelistNumber string) error {
	cmd := exec.Command("p4", "edit", "-c", changelistNumber, file)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to edit file: %v", err)
	}
	return nil
}
func p4LockFile(file string) error {
	cmd := exec.Command("p4", "lock", file)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to lock file: %v", err)
	}
	return nil
}
func checkoutFilesToChangelist(p4Test *P4Test, username string, fileName string, changelistNumber string) error {
	// Ensure the P4CLIENT and P4USER environment variables are set
	workspaceName := username + "_ws"
	os.Setenv("P4CLIENT", workspaceName)
	os.Setenv("P4USER", username)
	clientRootFile := filepath.Join(p4Test.clientRoot, username, fileName)
	// Construct the command string
	cmdString := fmt.Sprintf("p4 edit -c %s %s", changelistNumber, clientRootFile)

	// Print the constructed command string
	fmt.Println("Executing command:", cmdString)

	// Execute the command
	cmd := exec.Command("bash", "-c", cmdString)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to check out file %s for user %s: %v", fileName, username, err)
	}
	// Lock the file
	lockCmdString := fmt.Sprintf("p4 lock %s", clientRootFile)
	fmt.Println("Executing lock command:", lockCmdString)
	cmd = exec.Command("bash", "-c", lockCmdString)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to lock file %s: %v", fileName, err)
	}
	return nil
}
