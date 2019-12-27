package interfaces

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

type ProjectAttributesRepoInterface interface {
	// Inserts or updates an existing ProjectDomainAttributes model into the database store.
	CreateOrUpdate(ctx context.Context, input models.ProjectAttributes) error
	// Returns a matching project when it exists.
	Get(ctx context.Context, project, resource string) (models.ProjectAttributes, error)
}
