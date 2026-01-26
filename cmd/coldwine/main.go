package main

import (
	"context"
	"log"
	"os"

	"github.com/mistakeknot/autarch/internal/coldwine/cli"
	"github.com/mistakeknot/autarch/pkg/intermute"
)

func main() {
	if stop, err := intermute.RegisterTool(context.Background(), "coldwine"); err != nil {
		log.Printf("intermute registration failed: %v", err)
	} else if stop != nil {
		defer stop()
	}
	if err := cli.Execute(); err != nil {
		// cobra already prints; just exit non-zero
		os.Exit(1)
	}
}
