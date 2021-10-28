package entrypoints

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/flyteorg/flyteidl/clients/go/admin"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

const (
	healthCheckSuccess = "Health check passed, Flyteadmin is up and running"
	healthCheckError   = "Health check failed with status %v"
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

				healthCheckResponse, err := clientSet.HealthServiceClient().Check(ctx,
					&grpc_health_v1.HealthCheckRequest{Service: "flyteadmin"})
				if err != nil {
					return err
				}
				if healthCheckResponse.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
					logger.Errorf(ctx, healthCheckError, healthCheckResponse.GetStatus())
					return errors.New(fmt.Sprintf(healthCheckError, healthCheckResponse.GetStatus()))
				}
				logger.Infof(ctx, "Health check response is %v", healthCheckResponse)
				return nil
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
