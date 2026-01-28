package main

import (
	"fmt"
	"os"

	"github.com/mistakeknot/autarch/internal/signals/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
