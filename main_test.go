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

// TODO Auto find checkpoint and journal
var p4payloadDir = "/home/will/P4GoForge/p4payload"
var checkpointPrefix = "zapload-payload"

var checkPointfile = "zaplock-payload.ckp.8"

const startDir = "/home/will/P4GoForge/tmp"
const binDir = "/home/will/P4GoForge/bin"
const serverPort = "1999"
const brokerPort = "1666"

type P4Test struct {
	startDir         string
	binDir           string
	p4payloadDir     string
	checkpointPrefix string
	checkPointfile   string
	p4d              string
	p4               string
	p4broker         string
	port             string
	bport            string
	testRoot         string
	serverRoot       string
	serverPort       string
	brokerPort       string
	serverLog        string
	brokerRoot       string
	clientRoot       string
	p4Passwd         string
	p4dProcess       *os.Process
	p4brokerProcess  *os.Process
}

func MakeP4Test(startDir string) *P4Test {
	os.Setenv("P4CONFIG", ".p4config")
	p4t := &P4Test{
		startDir:     startDir,
		binDir:       binDir,
		p4payloadDir: p4payloadDir,

		// Do not initialize serverRoot, brokerRoot, and clientRoot here
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
	//p4t.port = "localhost:" + serverPort

	//p4t.bport = "localhost:" + brokerPort
	p4t.bport = fmt.Sprintf("rsh:%s -c %s -q", p4t.p4broker, p4t.brokerRoot+"/p4broker.cfg")
	fmt.Println(coloredOutput(colorRed, p4t.bport))
	p4t.port = fmt.Sprintf("rsh:%s -r %s -L %s", p4t.p4d, p4t.serverRoot, p4t.serverLog)
	os.Chdir(p4t.clientRoot)

	return p4t
}

// setupTestEnv initializes the Perforce server environment for testing.
func setupTestEnv(t *testing.T) *P4Test {
	// Create the P4Test instance first
	p4Test := MakeP4Test(startDir)

	// Now set the environment variables using p4Test
	originalPath := os.Getenv("PATH")
	newPath := fmt.Sprintf("%s:%s", originalPath, p4Test.binDir)
	os.Setenv("PATH", newPath)

	//os.Setenv("P4CONFIG", ".p4config")
	os.Setenv("P4ROOT", p4Test.serverRoot)
	os.Setenv("P4LOG", p4Test.serverLog)
	os.Setenv("P4PORT", p4Test.port)
	os.Setenv("P4PASSWD", p4Test.p4Passwd)
	//os.Setenv("P4PORT", "1")

	t.Log("Test environment set up successfully")

	if err := startP4dDaemon(p4Test); err != nil {
		t.Fatalf("Failed to start p4d: %v", err)
	}
	t.Log("p4d started successfully")

	rshp4dCommand := fmt.Sprintf(`"rsh:%s -r %s -L %s -d"`, p4Test.p4d, p4Test.serverRoot, p4Test.serverLog)
	rshp4brokerCommand := fmt.Sprintf(`"rsh:%s -c %s -d"`, p4Test.p4broker, p4Test.brokerRoot+"/p4broker.cfg")

	// BROKER STUFF

	brokerConfig := BrokerConfig{
		TargetPort: rshp4dCommand, // Updated to use rsh

		//ListenPort:  brokerPort,
		ListenPort: rshp4brokerCommand,

		Directory:   binDir,
		Logfile:     p4Test.brokerRoot + "/p4broker.log",
		AdminName:   "Helix Core Admins",
		AdminPhone:  "999/911",
		AdminEmail:  "helix-core-admins@example.com",
		ServerID:    "brokerSvrID",
		ZaplockPath: binDir + "/zaplock",
	}
	configPath := p4Test.brokerRoot + "/p4broker.cfg"

	if err := generateBrokerConfig(brokerConfig, configPath); err != nil {
		t.Fatalf("Failed to generate broker config: %v", err)
	}
	if err := startP4broker(p4Test); err != nil {
		t.Fatalf("Failed to start p4broker: %v", err)
	}
	t.Log("p4broker started successfully")

	return p4Test
}

func startP4dDaemon(p4t *P4Test) error {
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

	// Start p4d using rsh
	p4dCmd := exec.Command(filepath.Join(p4t.binDir, "p4d"), "-r", p4t.serverRoot, "-L", p4t.serverLog, "-vserver=3 -d")
	if err := p4dCmd.Start(); err != nil {
		return fmt.Errorf("failed to start p4d: %v", err)
	}

	p4t.p4dProcess = p4dCmd.Process
	return nil
}

func startP4broker(p4t *P4Test) error {
	// Construct the command
	rshCommand := fmt.Sprintf("%s -c %s -d", p4t.p4broker, p4t.brokerRoot+"/p4broker.cfg")

	// Split the command into executable and arguments
	cmdParts := strings.Fields(rshCommand)
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)

	// Capture output (optional, for debugging)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start p4broker with rsh: %v", err)
	}

	// Wait and capture output
	err := cmd.Wait()
	stdoutStr, stderrStr := stdoutBuf.String(), stderrBuf.String()
	fmt.Printf("p4broker STDOUT:\n%s\n", stdoutStr)
	fmt.Printf("p4broker STDERR:\n%s\n", stderrStr)

	// Error handling
	if err != nil {
		return fmt.Errorf("p4broker command failed: %v", err)
	}

	// Store the process for later cleanup
	p4t.p4brokerProcess = cmd.Process
	return nil
}

func teardownTestEnv(t *testing.T, p4t *P4Test, originalPath string) {
	// Since p4d is started with rsh and tied to the test process,
	// it should terminate automatically with the test.
	// If there is a need to explicitly stop it, additional logic will be needed.

	// Kill the p4broker process
	if p4t.p4brokerProcess != nil {
		if err := p4t.p4brokerProcess.Kill(); err != nil {
			t.Logf("Failed to kill p4broker process: %v", err)
		} else {
			t.Log("p4broker process killed successfully")
		}
	}

	// Restore the original PATH
	os.Setenv("PATH", originalPath)

	// Remove the test directory
	rmCmd := fmt.Sprintf("rm -rf %s*", p4t.testRoot)
	//rmCmd := fmt.Sprintf("echo WOOF WOOF")
	if err := exec.Command("bash", "-c", rmCmd).Run(); err != nil {
		t.Logf("Failed to remove test directory: %v", err)
	} else {
		fmt.Printf("Test environment cleaned up successfully\n")
	}
}
func withBrokerEnvVar(name, tempValue string, testFunc func()) {
	originalValue := os.Getenv(name)
	defer os.Setenv(name, originalValue)
	os.Setenv(name, tempValue)
	testFunc()
}
func TestP4VersionCommand(t *testing.T) {
	fmt.Println(coloredOutput(colorPurple, "TestP4VersionCommand"))
	originalPath := os.Getenv("PATH")
	p4Test := setupTestEnv(t)
	defer teardownTestEnv(t, p4Test, originalPath)

	output, err := P4Command(true, false, "-V")
	if err != nil {
		t.Fatalf("P4Command failed: %v", err)
	}

	fmt.Println(coloredOutput(colorGreen, "Output of p4 -V:\n"+output))
	fmt.Println("")
}

func TestP4SetCommand(t *testing.T) {
	fmt.Println(coloredOutput(colorPurple, "TestP4SetCommand"))
	originalPath := os.Getenv("PATH")
	p4Test := setupTestEnv(t)
	defer teardownTestEnv(t, p4Test, originalPath)
	output, err := P4Command(true, false, "set")
	if err != nil {
		t.Fatalf("P4Command failed: %v", err)
	}
	fmt.Println(coloredOutput(colorCyan, "Output of p4 set:\n"+output))
}

func TestP4InfoCommand(t *testing.T) {
	fmt.Println(coloredOutput(colorPurple, "TestP4InfoCommand"))
	originalPath := os.Getenv("PATH")
	p4Test := setupTestEnv(t)
	defer teardownTestEnv(t, p4Test, originalPath)
	output, err := P4Command(true, false, "info")
	if err != nil {
		t.Fatalf("P4Command failed: %v", err)
	}
	fmt.Println(coloredOutput(colorYellow, "Output of p4 info:\n"+output))
}

func TestP4DepotsCommand(t *testing.T) {
	fmt.Println(coloredOutput(colorPurple, "TestP4DepotsCommand"))
	originalPath := os.Getenv("PATH")
	p4Test := setupTestEnv(t)
	defer teardownTestEnv(t, p4Test, originalPath)
	output, err := P4Command(true, false, "depots")
	if err != nil {
		t.Fatalf("P4Command failed: %v", err)
	}
	fmt.Println(coloredOutput(colorYellow, "Output of p4 depots:\n"+output))
}

func TestP4InfoCommandBroker(t *testing.T) {
	fmt.Println(coloredOutput(colorPurple, "TestP4InfoCommandBroker"))

	originalPath := os.Getenv("PATH")
	p4Test := setupTestEnv(t)
	defer teardownTestEnv(t, p4Test, originalPath)
	// Temporarily change P4PORT to broker's port for this test
	originalPort := os.Getenv("P4PORT")
	defer os.Setenv("P4PORT", originalPort)

	//os.Setenv("P4PORT", brokerPort)
	os.Setenv("P4PORT", p4Test.bport)
	formattedString := fmt.Sprintf("P4PORT=%v", p4Test.bport)
	fmt.Println(coloredOutput(colorBlue, formattedString))

	output, err := P4Command(true, false, "info")
	if err != nil {
		t.Fatalf("P4Command failed: %v", err)
	}

	fmt.Println(coloredOutput(colorBlue, "Output of p4 info (Broker):\n"+output))
}

func TestP4ZapLockCommandBroker(t *testing.T) {
	originalPath := os.Getenv("PATH")
	p4Test := setupTestEnv(t)
	defer teardownTestEnv(t, p4Test, originalPath)
	// Temporarily change P4PORT to broker's port for this test
	originalPort := os.Getenv("P4PORT")
	defer os.Setenv("P4PORT", originalPort)

	os.Setenv("P4PORT", p4Test.bport)
	formattedString := fmt.Sprintf("P4PORT=%v", p4Test.bport)
	fmt.Println(coloredOutput(colorBlue, formattedString))

	output, err := P4Command(true, false, "zaplock", "-h")
	if err != nil {
		t.Fatalf("P4Command failed: %v", err)
	}

	fmt.Println(coloredOutput(colorBlue, "Output of p4 info (Broker):\n"+output))
}
