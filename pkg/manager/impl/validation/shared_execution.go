package validation

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
)

func ValidateCluster(ctx context.Context, db repositories.RepositoryInterface, executionID *core.WorkflowExecutionIdentifier, cluster string) error {
	if len(cluster) == 0 || cluster == common.DefaultProducerID {
		return nil
	}
	workflowExecution, err := db.ExecutionRepo().Get(ctx, repoInterfaces.Identifier{
		Project: executionID.Project,
		Domain:  executionID.Domain,
		Name:    executionID.Name,
	})
	if err != nil {
		logger.Debugf(ctx, "Failed to find existing execution with id [%+v] with err: %v", executionID, err)
		return err
	}
	if workflowExecution.Cluster != cluster {
		errorMsg := fmt.Sprintf("Cluster/producer from event [%s] does not match existing workflow execution cluster: [%s]",
			workflowExecution.Cluster, cluster)
		return errors.NewIncompatibleClusterError(ctx, errorMsg, workflowExecution.Cluster)
	}
	return nil
}
