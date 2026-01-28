package cli

import (
	"github.com/railzwaylabs/railzway-cloud/internal/app"
	"github.com/spf13/cobra"
)

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "migrate [up|down]",
		Short:     "Run database migrations",
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: []string{"up", "down"},
		RunE: func(cmd *cobra.Command, args []string) error {
			action := "up"
			if len(args) > 0 {
				action = args[0]
			}

			return app.RunMigrations(action)
		},
	}

	return cmd
}
