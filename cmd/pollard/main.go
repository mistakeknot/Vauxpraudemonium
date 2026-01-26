package main

import (
	"context"
	"log"

	"github.com/mistakeknot/autarch/internal/pollard/cli"
	"github.com/mistakeknot/autarch/pkg/intermute"
)

func main() {
	if stop, err := intermute.RegisterTool(context.Background(), "pollard"); err != nil {
		log.Printf("intermute registration failed: %v", err)
	} else if stop != nil {
		defer stop()
	}
	cli.Execute()
}
