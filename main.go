package main

import (
	"fmt"
	"os"

	"github.com/streed/ml-notes/cmd"
)

// Version is set via ldflags during build
var Version = "dev"

func main() {
	// Set version for the cmd package
	cmd.Version = Version
	
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
