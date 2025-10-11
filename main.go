package main

import (
	"fmt"
	"os"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("sdek version %s (build date: %s)\n", version, buildDate)
		os.Exit(0)
	}

	fmt.Println("sdek CLI - Compliance Evidence Mapping Tool")
	fmt.Println("Version:", version)
	fmt.Println()
	fmt.Println("Implementation in progress...")
	fmt.Println("Run 'sdek --version' to see version information")

	os.Exit(0)
}
