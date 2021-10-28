package entrypoints

import (
	"context"

	"github.com/flyteorg/flyteidl/clients/go/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

const (
	healthCheckSuccess = "Health check passed, Flyteadmin is up and running"
)

var preCheckRunCmd = &cobra.Command{
	Use:   "precheck",
	Short: "This command will check pre requirement for scheduler",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Do maximum of 30 retries on failures with constant backoff factor
		opts := wait.Backoff{Duration: 3000, Factor: 2.0, Steps: 30}
		err := retry.OnError(opts,
			func(err error) bool {
				logger.Errorf(ctx, "Attempt failed due to %v", err)
				return err != nil
			},
			func() error {
				clientSet, err := admin.ClientSetBuilder().WithConfig(admin.GetConfig(ctx)).Build(ctx)

				if err != nil {
					logger.Errorf(ctx, "Flyte native scheduler precheck failed due to %v\n", err)
					return err
				}

				_, err = clientSet.AuthMetadataClient().GetOAuth2Metadata(ctx, &service.OAuth2MetadataRequest{})
				return err
			},
		)

		if err != nil {
			return err
		}

		logger.Infof(ctx, healthCheckSuccess)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(preCheckRunCmd)
}
