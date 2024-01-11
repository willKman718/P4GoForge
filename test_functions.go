package main

import (
	"fmt"
	"os"
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
