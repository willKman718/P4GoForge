// P4GoForge
/*
  _____  _  _    _____       ______
 |  __ \| || |  / ____|     |  ____|
 | |__) | || |_| |  __  ___ | |__ ___  _ __ __ _  ___
 |  ___/|__   _| | |_ |/ _ \|  __/ _ \| '__/ _` |/ _ \
 | |       | | | |__| | (_) | | | (_) | | | (_| |  __/
 |_|       |_|  \_____|\___/|_|  \___/|_|  \__, |\___|
                                            __/ |
                                           |___/

*/

// TODO Clean up these ugly ugly variables
// TODO better enviroment variable handling
// TOOD fix pathings

/*
How to send a string of p4 commands

	commands := []string{
		"info",
		"configure show allservers",
		"groups",
		"users -a",
	}

output, err := P4Commands(p4Test, commands, DefaultP4Config)  <-- DefaultP4Config or BrokerP4Config (See test_functions.go for more details)
*/
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	// Other necessary imports...
)

var loadBroker = true
var BrokerDebug = false
var loadCheckPoint = false

var startDir string
var binDir string
var p4payloadDir string
var checkPointfile string
var newPath string
var originalEnvVars map[string]string
var wd string

func init() {
	fmt.Println(coloredOutput(colorBlue, "THIS IS SPARTA"))
	// Initialize global variables here
	var err error

	wd, err = os.Getwd()

	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		os.Exit(1)
	}
	fmt.Printf("wd = %s\n", wd)
	// Save the current environment variables
	originalEnvVars = make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			originalEnvVars[pair[0]] = pair[1]
		}
	}
	/*
		fmt.Println(coloredOutput(colorBlue, "THIS IS SPARTA"))
		// Save the current environment variables
		originalEnvVars = make(map[string]string)
		for _, env := range os.Environ() {
			pair := strings.SplitN(env, "=", 2)
			if len(pair) == 2 {
				originalEnvVars[pair[0]] = pair[1]
			}
		}
		// Initialize global variables here
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current working directory:", err)
			os.Exit(1)
		}

		startDir = filepath.Join(wd, "tmp")
		binDir = filepath.Join(wd, "bin")
		p4payloadDir = filepath.Join(wd, "p4payload")
		checkPointfile = "zaplock-payload.ckp.8"

		originalPath := os.Getenv("PATH")
		newPath := fmt.Sprintf("%s:%s", originalPath, binDir)
		os.Setenv("PATH", newPath)

		if err := checkBinaries(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	*/
}

type P4Test struct {
	startDir           string
	binDir             string
	p4payloadDir       string
	checkPointfile     string
	p4d                string
	p4                 string
	p4broker           string
	port               string
	bport              string
	testRoot           string
	serverRoot         string
	serverPort         string
	brokerPort         string
	serverLog          string
	brokerRoot         string
	clientRoot         string
	p4Passwd           string
	rshp4dCommand      string
	rshp4brokerCommand string
	p4dProcess         *os.Process
	p4brokerProcess    *os.Process
}

func makeP4Test(startDir string) *P4Test {
	fmt.Println(coloredOutput(colorBlue, "makeP4Test"))
	/*
		// Save the current environment variables
		originalEnvVars = make(map[string]string)
		for _, env := range os.Environ() {
			pair := strings.SplitN(env, "=", 2)
			if len(pair) == 2 {
				originalEnvVars[pair[0]] = pair[1]
			}
		}
	*/
	startDir = filepath.Join(wd, "tmp")
	binDir = filepath.Join(wd, "bin")
	p4payloadDir = filepath.Join(wd, "p4payload")
	checkPointfile = "zaplock-payload.ckp.8"

	originalPath := os.Getenv("PATH")
	newPath := fmt.Sprintf("%s:%s", originalPath, binDir)
	os.Setenv("PATH", newPath)

	if err := checkBinaries(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Setenv("P4CONFIG", ".p4config")
	p4t := &P4Test{
		startDir:     startDir,
		binDir:       binDir,
		p4payloadDir: p4payloadDir,
	}

	// Now p4t is fully defined, initialize the other fields

	p4t.testRoot = filepath.Join(p4t.startDir, "testroot")
	p4t.serverRoot = filepath.Join(p4t.testRoot, "server")
	p4t.serverLog = filepath.Join(p4t.testRoot, "server-log/log")
	p4t.brokerRoot = filepath.Join(p4t.testRoot, "broker")
	p4t.clientRoot = filepath.Join(p4t.testRoot, "client")
	p4t.checkPointfile = filepath.Join(p4t.p4payloadDir, "/", checkPointfile)
	p4t.p4Passwd = "perforce"
	p4t.ensureDirectories()
	p4t.p4d = filepath.Join(p4t.binDir, "p4d")
	p4t.p4 = filepath.Join(p4t.binDir, "p4")
	p4t.p4broker = filepath.Join(p4t.binDir, "p4broker")
	p4t.bport = fmt.Sprintf("rsh:%s -c %s -q", p4t.p4broker, p4t.brokerRoot+"/p4broker.cfg")
	p4t.port = fmt.Sprintf("rsh:%s -r %s -L %s", p4t.p4d, p4t.serverRoot, p4t.serverLog)
	//os.Chdir(p4t.clientRoot)

	return p4t
}

// setupTestEnv initializes the Perforce server environment for testing.
func setupTestEnv(t *testing.T) *P4Test {
	fmt.Println(coloredOutput(colorBlue, "setupTestEnv"))
	// Create the P4Test instance first
	p4Test := makeP4Test(startDir)

	// Now set the environment variables using p4Test

	setP4Config(DefaultP4Config, p4Test)
	t.Log("Test environment set up successfully")

	if err := startP4dDaemon(p4Test); err != nil {
		t.Fatalf("Failed to start p4d: %v", err)
	}
	t.Log("p4d started successfully")

	p4Test.rshp4dCommand = fmt.Sprintf(`"rsh:%s -r %s -L %s -d -q"`, p4Test.p4d, p4Test.serverRoot, p4Test.serverLog)

	if loadBroker {
		if err := setupP4Broker(p4Test); err != nil {
			t.Fatalf("Broker setup failed: %v", err)
		}
		t.Log("p4broker started successfully")
	}
	os.Setenv("P4USER", "perforce")
	////// CREATE USERS
	// Create Users

	users := []string{"user1", "user2", "user3", "user4", "user5"}
	for _, user := range users {
		if err := createUser(p4Test, user); err != nil {
			t.Fatalf("Error creating user %s: %v", user, err)
		}
		if err := createWorkspace(p4Test, user); err != nil {
			t.Fatalf("Error creating user %s: %v", user, err)
		}
	}
	// Create the group and add specified users to it.
	// Define the users to be added to the AuthorizedUser-ZapLock group
	authorizedUsers := []string{"user1", "user3", "user5"}

	// Create the group and add specified users to it
	if err := createGroup(p4Test, "AuthorizedUser-ZapLock", authorizedUsers); err != nil {
		t.Fatalf("Error creating group AuthorizedUser-ZapLock: %v", err)
	}

	//TODO move this probaly
	type FileDetails struct {
		FileType string
		UserName string
	}
	filesToCreate := map[string]FileDetails{
		"eLOCKtextfile1.txt":   {FileType: "text+lmx", UserName: "user1"},
		"mLOCKbinaryfile1.bin": {FileType: "binary", UserName: "user1"},
		"mLOCKtextfile2.txt":   {FileType: "text", UserName: "user5"},
		"eLOCKtextfile3.txt":   {FileType: "text+l", UserName: "user5"},
		"eLOCKbinaryfile2.bin": {FileType: "binary+klm", UserName: "user5"},
	}
	////////////////////user2 is not in the group
	/*
		filesToCreate := map[string]string{
			"textfile1.txt":   "text+lmx",
			"binaryfile1.bin": "binary",
			"textfile2.txt":   "text",
			"textfile3.txt":   "text+l",
			"binaryfile2.bin": "binary+klm",
		}
	*/
	/*
		if err := createDepotFiles(p4Test, "user1", filesToCreate); err != nil {
			t.Fatalf("Error creating and adding files to depot: %v", err)
		}*/
	for fileName, details := range filesToCreate {
		if err := createDepotFiles(p4Test, details.UserName, map[string]string{fileName: details.FileType}); err != nil {
			t.Fatalf("Error creating and adding file %s to depot by user %s: %v", fileName, details.UserName, err)
		}
	}

	// Sync workspaces and create changelists for each user
	for _, user := range users {
		if err := syncUserWorkspace(p4Test, user); err != nil {
			t.Fatalf("Error syncing workspace for user %s: %v", user, err)
		}

		changelistNumber, err := createChangelist(p4Test, user)
		if err != nil {
			t.Fatalf("Error creating changelist for user %s: %v", user, err)
		}

		// Log or do something with the changelist number
		fmt.Printf("Changelist for user %s created: %s\n", user, changelistNumber)

		// Check out files associated with this user to their changelist
		for fileName, details := range filesToCreate {
			if details.UserName == user {
				fmt.Printf("Checking out file: %s, User: %s, Changelist: %s\n", fileName, user, changelistNumber)
				// Uncomment below line to actually execute the function after dry run
				if err := checkoutFilesToChangelist(p4Test, user, fileName, changelistNumber); err != nil {
					t.Fatalf("Error checking out file %s to changelist %s by user %s: %v", fileName, changelistNumber, user, err)
				}
			}
		}
	}

	/*filesToLock := []string{"textfile1.txt", "textfile2.txt"}

	usersToLockFiles := []string{"user2", "user4"}
	for _, user := range usersToLockFiles {
		if err := checkoutAndLockFiles(p4Test, user, filesToLock); err != nil {
			t.Fatalf("Error checking out and locking files for user %s: %v", user, err)
		}
	}
	*/
	/*
		errUser2 := checkoutAndLockFiles(p4Test, "user2", []string{"textfile1.txt", "textfile2.txt"})
		if errUser2 != nil {
			t.Fatalf("Error checking out and locking files for user2: %v", errUser2)
		}

		errUser4 := checkoutAndLockFiles(p4Test, "user4", []string{"textfile1.txt", "textfile2.txt"})
		if errUser4 != nil {
			t.Fatalf("Error checking out and locking files for user4: %v", errUser4)
		}
	*/
	return p4Test

}

func startP4dDaemon(p4t *P4Test) error {
	fmt.Println(coloredOutput(colorBlue, "startP4dDaemon"))
	if loadCheckPoint {
		// Load checkpoint
		output, err := P4dCommand(true, "-r", p4t.serverRoot, "-L", p4t.serverLog, "-jr", p4t.checkPointfile)
		fmt.Printf("Load checkpoint output: %s\n", output)
		if err != nil {
			return fmt.Errorf("failed to load checkpoint: %v", err)
		}
		fmt.Println("Checkpoint loaded successfully.")

		// Update the database
		output, err = P4dCommand(true, "-r", p4t.serverRoot, "-L", p4t.serverLog, "-xu")
		fmt.Printf("Update database output: %s\n", output)
		if err != nil {
			return fmt.Errorf("failed to update db: %v", err)
		}
		fmt.Println("Database updated successfully.")
	}
	// Start p4d using rsh
	p4dCmd := exec.Command(filepath.Join(p4t.binDir, "p4d"), "-r", p4t.serverRoot, "-L", p4t.serverLog, "-vserver=3 -d -q")
	if err := p4dCmd.Start(); err != nil {
		return fmt.Errorf("failed to start p4d: %v", err)
	}

	p4t.p4dProcess = p4dCmd.Process
	return nil
}

func startP4broker(p4t *P4Test) error {
	fmt.Println(coloredOutput(colorBlue, "startP4broker"))
	// Construct the command
	rshCommand := fmt.Sprintf("%s -c %s -d", p4t.p4broker, p4t.brokerRoot+"/p4broker.cfg")

	// Split the command into executable and arguments
	cmdParts := strings.Fields(rshCommand)
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)

	// Declare the buffers outside the if statement
	var stdoutBuf, stderrBuf bytes.Buffer

	if BrokerDebug {
		// Assign the buffers for capturing output
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start p4broker with rsh: %v", err)
	}

	// Wait for the command to finish if debugging is enabled
	if BrokerDebug {
		if err := cmd.Wait(); err != nil {
			// Capture and print output if there is an error
			stdoutStr, stderrStr := stdoutBuf.String(), stderrBuf.String()
			fmt.Printf("p4broker STDOUT:\n%s\n", stdoutStr)
			fmt.Printf("p4broker STDERR:\n%s\n", stderrStr)
			return fmt.Errorf("p4broker command failed: %v", err)
		}
	}

	// Store the process for later cleanup
	p4t.p4brokerProcess = cmd.Process
	return nil
}

func teardownTestEnv(t *testing.T, p4t *P4Test) {
	fmt.Println(coloredOutput(colorRed, "teardownTestEnv"))
	fmt.Printf("Working Dir %s\n", wd)

	// Terminate p4d process
	if p4t.p4dProcess != nil {
		if err := p4t.p4dProcess.Kill(); err != nil {
			fmt.Printf("Failed to kill p4d process: %v\n", err)
		} else {
			fmt.Printf("p4d process terminated successfully\n")
		}
	}

	// Terminate p4broker process
	if p4t.p4brokerProcess != nil {
		if err := p4t.p4brokerProcess.Kill(); err != nil {
			fmt.Printf("Failed to kill p4broker process: %v\n", err)
		} else {
			fmt.Printf("p4broker process terminated successfully\n")
		}
	}

	// Clear environment variables and restore original ones
	os.Clearenv()
	for key, value := range originalEnvVars {
		if err := os.Setenv(key, value); err != nil {
			fmt.Printf("Failed to restore environment variable %s: %v\n", key, err)
		} else {
			//fmt.Printf("Restored environment variable %s successfully\n", key)
		}
	}

	// Remove the test directory
	rmCmd := fmt.Sprintf("rm -rvf %s", p4t.testRoot)
	//rmCmd := fmt.Sprintf("woof woof")
	if err := exec.Command("bash", "-c", rmCmd).Run(); err != nil {
		fmt.Printf("Failed to remove test directory: %v\n", err)
	} else {
		fmt.Printf("Test directory removed successfully\n")
	}

	// Check if test directory is empty
	isEmpty, err := isDirectoryEmpty(p4t.testRoot)
	if err != nil {
		fmt.Printf("checking if test directory is empty or exists: %v\n", err)
	} else if isEmpty {
		fmt.Printf("Test directory is empty\n")
	} else {
		fmt.Printf("Test directory is not empty\n")
	}
	fmt.Println(coloredOutput(colorRed, "teardownTestEnv complete"))
	// Log environment variables after teardown
	logEnvironmentVariables("Environment Variables After Teardown\n")
}

func buildp4dTestEnv(t *testing.T, p4t *P4Test) {
	fmt.Printf("Building p4d depots and suchs\n")
	// The needful commands
}
func TestP4OGCommands(t *testing.T) {
	logEnvironmentVariables("Environment Variables Before Setup")

	funcName := getFunctionName()
	fmt.Println(coloredOutput(colorPurple, funcName))

	p4Test := setupTestEnv(t)        // Setup test environment
	defer teardownTestEnv(t, p4Test) // Teardown test environment

	commands := []string{
		"info",
		"configure show allservers",
		"groups",
		"users -a",
		"clients",
		"depots",
		"files //...",
		"changes",
		"describe 1",
		//"fstat //...",
		"group -o AuthorizedUser-ZapLock",
	}

	output, err := P4Commands(p4Test, commands, DefaultP4Config)
	if err != nil {
		t.Fatalf("Error executing p4 commands: %v", err)
	}
	logEnvironmentVariables("Environment Variables During Setup")

	fmt.Println("Output of p4 commands:\n", output)
}

func TestP4OGBrokerCommands(t *testing.T) {
	funcName := getFunctionName()
	fmt.Println(coloredOutput(colorPurple, funcName))

	p4Test := setupTestEnv(t)        // Setup test environment
	defer teardownTestEnv(t, p4Test) // Teardown test environment

	commands := []string{
		"info",
		"configure show allservers",
		"groups",
		"users -a",
		"clients",
		"depots",
		"files //...",
		"changes",
		"fstat //...",
		"describe 1",
		//"group -o AuthorizedUser-ZapLock",
	}

	output, err := P4Commands(p4Test, commands, BrokerP4Config)
	if err != nil {
		t.Fatalf("Error executing p4 commands: %v", err)
	}

	fmt.Println("Output of p4 commands:\n", output)
}

func TestP4OGZaplockHelpCommands(t *testing.T) {
	funcName := getFunctionName()
	fmt.Println(coloredOutput(colorPurple, funcName))

	p4Test := setupTestEnv(t)        // Setup test environment
	defer teardownTestEnv(t, p4Test) // Teardown test environment

	commands := []string{
		"zaplock -h",
	}

	output, err := P4Commands(p4Test, commands, BrokerP4Config)
	if err != nil {
		t.Fatalf("Error executing p4 commands: %v", err)
	}

	fmt.Println("Output of p4 commands:\n", output)
}

func TestP4OGZaplockCCommandsNONAUTH(t *testing.T) {
	funcName := getFunctionName()
	fmt.Println(coloredOutput(colorPurple, funcName))

	p4Test := setupTestEnv(t)        // Setup test environment
	defer teardownTestEnv(t, p4Test) // Teardown test environment

	commands := []string{
		"zaplock -c 10 -L -M -y",
	}
	os.Setenv("P4USER", "user2")
	output, err := P4Commands(p4Test, commands, BrokerP4Config)
	if err != nil {
		t.Fatalf("Error executing p4 commands: %v", err)
	}

	fmt.Println("Output of p4 commands:\n", output)
}
func TestP4OGZaplockCCommandsAUTH(t *testing.T) {
	funcName := getFunctionName()
	fmt.Println(coloredOutput(colorPurple, funcName))

	p4Test := setupTestEnv(t)        // Setup test environment
	defer teardownTestEnv(t, p4Test) // Teardown test environment

	commands := []string{
		"zaplock -c 10 -L -M -y",
	}
	os.Setenv("P4USER", "user3")
	output, err := P4Commands(p4Test, commands, BrokerP4Config)
	if err != nil {
		t.Fatalf("Error executing p4 commands: %v", err)
	}

	fmt.Println("Output of p4 commands:\n", output)
}
func TestP4OGZaplockCCommandsAUTHargL(t *testing.T) {
	funcName := getFunctionName()
	fmt.Println(coloredOutput(colorPurple, funcName))

	p4Test := setupTestEnv(t)        // Setup test environment
	defer teardownTestEnv(t, p4Test) // Teardown test environment

	commands := []string{
		"zaplock -c 10 -L -y",
	}
	os.Setenv("P4USER", "user3")
	output, err := P4Commands(p4Test, commands, BrokerP4Config)
	if err != nil {
		t.Fatalf("Error executing p4 commands: %v", err)
	}

	fmt.Println("Output of p4 commands:\n", output)
}
func TestP4OGZaplockCCommandsAUTHargM(t *testing.T) {
	funcName := getFunctionName()
	fmt.Println(coloredOutput(colorPurple, funcName))

	p4Test := setupTestEnv(t)        // Setup test environment
	defer teardownTestEnv(t, p4Test) // Teardown test environment

	commands := []string{
		"zaplock -c 10 -M -y",
	}
	os.Setenv("P4USER", "user3")
	output, err := P4Commands(p4Test, commands, BrokerP4Config)
	if err != nil {
		t.Fatalf("Error executing p4 commands: %v", err)
	}

	fmt.Println("Output of p4 commands:\n", output)
}
