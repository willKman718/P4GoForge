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
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	// Other necessary imports...
)

// var startDir = "/home/will/temp-p4goforge"
var startDir = "/home/will/P4GoForge/tmp"
var binDir = "/home/will/P4GoForge/bin"

type P4Test struct {
	startDir        string
	p4d             string
	p4              string
	p4broker        string
	port            string
	bport           string
	testRoot        string
	serverRoot      string
	serverPort      string
	brokerPort      string
	serverLog       string
	brokerRoot      string
	clientRoot      string
	binDir          string
	p4dProcess      *os.Process // Add this field to store the p4d process
	p4brokerProcess *os.Process // New field for p4broker process

}

func MakeP4Test(startDir string) *P4Test {
	os.Setenv("P4CONFIG", ".p4config")
	p4t := &P4Test{
		startDir:   startDir,
		binDir:     binDir,
		serverPort: "9666", // Assigning value
		brokerPort: "9777", // Assigning value
		// Do not initialize serverRoot, brokerRoot, and clientRoot here
	}

	// Now p4t is fully defined, initialize the other fields
	//const p4t.serverPort = "9666"
	//const p4t.brokerPort = "9777"
	p4t.testRoot = filepath.Join(p4t.startDir, "testroot")
	p4t.serverRoot = filepath.Join(p4t.testRoot, "server")
	p4t.serverLog = filepath.Join(p4t.testRoot, "server-log/log")
	p4t.brokerRoot = filepath.Join(p4t.testRoot, "broker")
	p4t.clientRoot = filepath.Join(p4t.testRoot, "client")

	p4t.ensureDirectories()
	p4t.p4d = filepath.Join(p4t.binDir, "p4d")
	p4t.p4 = filepath.Join(p4t.binDir, "p4")
	p4t.p4broker = filepath.Join(p4t.binDir, "p4broker")
	p4t.port = "localhost:" + p4t.serverPort
	p4t.bport = "localhost:" + p4t.brokerPort
	//p4t.port = fmt.Sprintf("rsh:%s -r \"%s\" -L log -vserver=3 -i", p4t.p4d, p4t.serverRoot)
	os.Chdir(p4t.clientRoot)
	p4config := filepath.Join(p4t.startDir, os.Getenv("P4CONFIG"))
	writeToFile(p4config, fmt.Sprintf("P4PORT=%s", p4t.port))
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

	t.Log("Test environment set up successfully")

	if err := startP4dDaemon(p4Test); err != nil {
		t.Fatalf("Failed to start p4d: %v", err)
	}
	t.Log("p4d started successfully")

	// BROKER STUFF
	brokerConfig := BrokerConfig{
		TargetPort:  p4Test.serverPort,
		ListenPort:  p4Test.brokerPort,
		Directory:   binDir,
		Logfile:     p4Test.brokerRoot + "/p4broker.log",
		AdminName:   "Helix Core Admins",
		AdminPhone:  "999/911",
		AdminEmail:  "helix-core-admins@example.com",
		ServerID:    "brokerSvrID",
		ZaplockPath: "/p4/common/bin/zaplock",
	}
	configPath := p4Test.brokerRoot + "/p4broker.log"
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
	p4dCmd := exec.Command(filepath.Join(p4t.binDir, "p4d"), "-r", p4t.serverRoot, "-p", p4t.port, "-L", p4t.serverLog, "-vserver=3 -d")
	//fmt.Printf("Running: %v\n", p4dCmd)
	fmt.Printf("Starting p4d: %v\n", p4dCmd)

	if err := p4dCmd.Start(); err != nil {
		return fmt.Errorf("failed to start p4d: %v", err)
	}
	p4t.p4dProcess = p4dCmd.Process
	return nil
}

func startP4broker(p4t *P4Test) error {
	brokerConfig := p4t.brokerRoot + "/broker.cfg"
	brokerPort := p4t.brokerPort

	p4brokerCmd := exec.Command(filepath.Join(p4t.binDir, "p4broker"), "-c", brokerConfig, "-d")
	// Uncomment the next line to see the command being run, for debugging
	fmt.Printf("Starting p4broker(%v): %v\n", brokerPort, p4brokerCmd)

	if err := p4brokerCmd.Start(); err != nil {
		return fmt.Errorf("failed to start p4broker: %v", err)
	}

	// Store the process for later cleanup
	p4t.p4brokerProcess = p4brokerCmd.Process

	return nil
}

func teardownTestEnv(t *testing.T, p4t *P4Test, originalPath string) {
	if p4t.p4dProcess != nil {
		if err := p4t.p4dProcess.Kill(); err != nil {
			t.Logf("Failed to kill p4d process: %v", err)
		}
	}
	// Restore the original PATH
	os.Setenv("PATH", originalPath)

	t.Log("Test environment cleaned up successfully")
}

func TestP4VersionCommand(t *testing.T) {
	//originalPath := os.Getenv("PATH")
	//p4Test := setupTestEnv(t)
	//defer teardownTestEnv(t, p4Test, originalPath)

	output, err := P4Command(true, false, "-V")
	if err != nil {
		t.Fatalf("P4Command failed: %v", err)
	}

	fmt.Println(coloredOutput(colorGreen, "Output of p4 -V:\n"+output))
	fmt.Println("")
}

func TestP4SetCommand(t *testing.T) {
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
	originalPath := os.Getenv("PATH")
	p4Test := setupTestEnv(t)
	defer teardownTestEnv(t, p4Test, originalPath)

	output, err := P4Command(true, false, "info")
	if err != nil {
		t.Fatalf("P4Command failed: %v", err)
	}

	fmt.Println(coloredOutput(colorYellow, "Output of p4 info:\n"+output))
}
func TestP4InfoCommandBroker(t *testing.T) {
	originalPath := os.Getenv("PATH")
	p4Test := setupTestEnv(t)
	defer teardownTestEnv(t, p4Test, originalPath)

	os.Setenv("P4PORT", p4Test.bport)
	output, err := P4Command(true, false, "info")
	if err != nil {
		t.Fatalf("P4Command failed: %v", err)
	}

	fmt.Println(coloredOutput(colorYellow, "Output of p4 info (Broker):\n"+output))
}
