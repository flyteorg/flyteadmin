package repositories

import (
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/repositories/config"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
)

type RepoConfig int32

const (
	POSTGRES RepoConfig = 0
)

var RepositoryConfigurationName = map[int32]string{
	0: "POSTGRES",
}

// The RepositoryInterface indicates the methods that each Repository must support.
// A Repository indicates a Database which is collection of Tables/models.
// The goal is allow databases to be Plugged in easily.
type RepositoryInterface interface {
	TaskRepo() interfaces.TaskRepoInterface
	WorkflowRepo() interfaces.WorkflowRepoInterface
	LaunchPlanRepo() interfaces.LaunchPlanRepoInterface
	ExecutionRepo() interfaces.ExecutionRepoInterface
	ExecutionEventRepo() interfaces.ExecutionEventRepoInterface
	ProjectRepo() interfaces.ProjectRepoInterface
	ResourceRepo() interfaces.ResourceRepoInterface
	NodeExecutionRepo() interfaces.NodeExecutionRepoInterface
	NodeExecutionEventRepo() interfaces.NodeExecutionEventRepoInterface
	TaskExecutionRepo() interfaces.TaskExecutionRepoInterface
	NamedEntityRepo() interfaces.NamedEntityRepoInterface
	SchedulableEntityRepo() interfaces.SchedulableEntityRepoInterface
	ScheduleEntitiesSnapshotRepo() interfaces.ScheduleEntitiesSnapShotRepoInterface
}

type SchedulerRepoInterface interface {
	SchedulableEntityRepo() interfaces.SchedulableEntityRepoInterface
	ScheduleEntitiesSnapshotRepo() interfaces.ScheduleEntitiesSnapShotRepoInterface
}

func GetRepository(repoType RepoConfig, dbConfig config.DbConfig, scope promutils.Scope) RepositoryInterface {
	switch repoType {
	case POSTGRES:
		postgresScope := scope.NewSubScope("postgres")
		db := config.OpenDbConnection(config.NewPostgresConfigProvider(dbConfig, postgresScope))
		return NewPostgresRepo(
			db,
			errors.NewPostgresErrorTransformer(postgresScope.NewSubScope("errors")),
			postgresScope.NewSubScope("repositories"))
	default:
		panic(fmt.Sprintf("Invalid repoType %v", repoType))
	}
}
