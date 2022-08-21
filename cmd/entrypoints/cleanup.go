package entrypoints

import (
	"context"
	"errors"
	"github.com/flyteorg/flyteadmin/pkg/server"
	"github.com/spf13/cobra"
	_ "gorm.io/driver/postgres" // Required to import database driver.
	"strconv"
)

var parentCleanupCmd = &cobra.Command{
	Use:   "migrate",
	Short: "This command cleans up the Database records in the Flyte admin database. Please choose a subcommand.",
}

var cleanupCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will run the cleanup for the Flyte admin database",
	Args:  cobra.ExactValidArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		retention, err := strconv.Atoi(args[0])
		if err != nil {
			return errors.New("argument needs to be an integer value")
		}
		return server.Cleanup(ctx, retention)
	},
}

func init() {
	RootCmd.AddCommand(parentCleanupCmd)
	parentMigrateCmd.AddCommand(cleanupCmd)
}
