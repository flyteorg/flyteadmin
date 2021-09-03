package entrypoints

import (
	"context"
	"github.com/avast/retry-go"
	adminClient "github.com/flyteorg/flyteidl/clients/go/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var preCheckRunCmd = &cobra.Command{
	Use:   "precheck",
	Short: "This command will check pre requirement for scheduler",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		err := retry.Do(
			func() error {
				clientSet, err := adminClient.ClientSetBuilder().WithConfig(adminClient.GetConfig(ctx)).Build(ctx)
				if err != nil {
					logger.Fatalf(ctx, "Flyte native scheduler failed to start due to %v", err)
					return err
				}
				version, err := clientSet.AdminClient().GetVersion(ctx, &admin.GetVersionRequest{})
				if err != nil {
					return err
				}
				if len(version.ControlPlaneVersion.Version) > 0 {
					return nil
				}
				return errors.New("something goes wrong")
			},
		)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(preCheckRunCmd)
}