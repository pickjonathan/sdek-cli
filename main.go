package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/pickjonathan/sdek-cli/cmd"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	// Set version in cmd package
	cmd.SetVersion(version)

	// Set up panic recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %v\n", r)
			fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", debug.Stack())
			os.Exit(3)
		}
	}()

	// Execute root command
	if err := cmd.Execute(); err != nil {
		// Cobra already prints the error, just exit with appropriate code
		os.Exit(1)
	}
}
