package main

import (
	"fmt"
	"os"

	"github.com/liaoran123/sfsDb/cmd/sfsdb/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
