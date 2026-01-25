package main

import (
	"os"

	"github.com/mistakeknot/autarch/internal/coldwine/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		// cobra already prints; just exit non-zero
		os.Exit(1)
	}
}
