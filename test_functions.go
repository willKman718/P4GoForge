package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

func coloredOutput(color, text string) string {
	return color + text + colorReset
}

func writeToFile(fname, contents string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(f, contents)
	if err != nil {
		_ = f.Close()
		return err
	}
	err = f.Close()
	return err
}

func (p4t *P4Test) ensureDirectories() {
	for _, d := range []string{p4t.serverRoot, p4t.brokerRoot, p4t.clientRoot} {
		err := os.MkdirAll(d, 0777)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create %s: %v", d, err)
		}
	}
}
func getFunctionName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "unknown"
	}

	fn := runtime.FuncForPC(pc)
	fullFuncName := fn.Name()

	// Split the full function name by the dot and return the last part
	parts := strings.Split(fullFuncName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1] // Last part is the actual function name
	}
	return "unknown"
}

type P4ConfigType string

const (
	DefaultP4Config P4ConfigType = "default"
	BrokerP4Config  P4ConfigType = "broker"
	CustomP4Config  P4ConfigType = "custom"
)

// Function to set different Perforce configurations
func setP4Config(configType P4ConfigType, p4Test *P4Test, newPath string) {
	switch configType {
	case DefaultP4Config:
		os.Setenv("PATH", newPath)
		os.Setenv("P4CONFIG", ".p4config")
		os.Setenv("P4ROOT", p4Test.serverRoot)
		os.Setenv("P4LOG", p4Test.serverLog)
		os.Setenv("P4USER", "perforce")
		os.Setenv("P4PORT", p4Test.port)
		os.Setenv("P4PASSWD", p4Test.p4Passwd)
		os.Setenv("P4TICKETS", p4Test.serverRoot)
		os.Setenv("P4TRUST", p4Test.serverRoot)
		// ... other default settings ...
	case BrokerP4Config:
		os.Setenv("P4PORT", p4Test.bport)

		// ... settings specific to broker configuration ...
	case CustomP4Config:
		os.Setenv("P4USER", "noodles")
		os.Setenv("P4PASSWD", "noodles")
		// ... custom settings ...
	}
	// Note: Add more cases as needed for different configurations
}
