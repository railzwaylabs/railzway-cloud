package cli

import (
	"github.com/railzwaylabs/railzway-cloud/internal/app"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var migrateFirst bool

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if migrateFirst {
				if err := app.RunMigrations("up"); err != nil {
					return err
				}
			}

			app.RunServer()
			return nil
		},
	}

	cmd.Flags().BoolVar(&migrateFirst, "migrate", false, "Run database migrations before starting the server")

	return cmd
}
