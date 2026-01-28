package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func Execute() error {
	return NewRoot().Execute()
}

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "signals",
		Short: "Signals broadcast server",
	}
	root.AddCommand(ServeCmd())
	root.SetOut(os.Stdout)
	return root
}
