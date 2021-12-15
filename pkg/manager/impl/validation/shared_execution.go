package validation

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"google.golang.org/grpc/codes"
)

func ValidateCluster(ctx context.Context, db repositories.RepositoryInterface, executionID *core.WorkflowExecutionIdentifier, cluster string) error {
	workflowExecution, err := db.ExecutionRepo().Get(ctx, repoInterfaces.Identifier{
		Project: executionID.Project,
		Domain:  executionID.Domain,
		Name:    executionID.Name,
	})
	if err != nil {
		logger.Debugf(ctx, "Failed to find existing execution with id [%+v] with err: %v", executionID, err)
		return err
	}
	if !(workflowExecution.Cluster == cluster || cluster == common.DefaultProducerID) {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "Cluster/producer from event [%s] does not match existing workflow execution cluster: [%s]",
			workflowExecution.Cluster, cluster)
	}
	return nil
}
