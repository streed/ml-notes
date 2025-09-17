package main

import (
	"fmt"
	"os"

	"github.com/streed/ml-notes/internal/cli"
)

// Version is set via ldflags during build
var Version = "dev"

func main() {
	// Set version for the CLI package
	cli.Version = Version

	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
