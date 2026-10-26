package main

import (
	"fmt"
	"os"

	"github.com/educates/educatesenv/pkg/cmd"
)

func main() {
	root := cmd.NewRootCommand()
	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintf(root.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
	}
}
