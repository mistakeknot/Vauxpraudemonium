package main

import (
	"context"
	"log"

	"github.com/mistakeknot/autarch/internal/gurgeh/cli"
	"github.com/mistakeknot/autarch/pkg/intermute"
)

func main() {
	if stop, err := intermute.RegisterTool(context.Background(), "gurgeh"); err != nil {
		log.Printf("intermute registration failed: %v", err)
	} else if stop != nil {
		defer stop()
	}
	_ = cli.Execute()
}
