package entrypoints

import (
	"context"
	adminEntrypoint "github.com/flyteorg/flyteadmin/cmd/entrypoints"
	propellerEntrypoint "github.com/flyteorg/flytepropeller/cmd/controller/cmd"

	_ "github.com/golang/glog"
	"github.com/spf13/cobra"
	_ "gorm.io/driver/postgres" // Required to import database driver.

	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/k8s"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/hive"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/kfoperators/mpi"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/kfoperators/pytorch"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/pod"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/sagemaker"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/spark"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/webapi/athena"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/webapi/snowflake"
)

func startAdmin(ctx context.Context) error {
	return adminEntrypoint.Serve(ctx)
}

func startPropeller(ctx context.Context) error {
	return propellerEntrypoint.Run(ctx)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "This command will start the flyte native scheduler and periodically get new schedules from the db for scheduling",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		childCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func(ctx context.Context) {
			err := startAdmin(ctx)
			if err != nil {
				return
			}
		}(childCtx)

		go func(ctx context.Context) {
			err := startPropeller(ctx)
			if err != nil {
				return
			}
		}(childCtx)

		<-ctx.Done()
		return nil
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
