package cli

import (
	"fmt"
	"os"

	"github.com/mistakeknot/autarch/pkg/signals"
	"github.com/spf13/cobra"
)

func ServeCmd() *cobra.Command {
	var addr string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve Signals WebSocket API (local-only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			srv := signals.NewServer(nil)
			fmt.Fprintf(cmd.OutOrStdout(), "Signals server listening on %s\n", addr)
			return srv.ListenAndServe(addr)
		},
	}
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:8092", "HTTP bind address")
	cmd.SetOut(os.Stdout)
	return cmd
}
