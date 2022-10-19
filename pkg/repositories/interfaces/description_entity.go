package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

type GetDescriptionEntityInput struct {
	ResourceType core.ResourceType
	Project      string
	Domain       string
	Name         string
	Version      string
}

type DescriptionEntityCollectionOutput struct {
	Entities []models.DescriptionEntity
}

// DescriptionEntityRepoInterface Defines the interface for interacting with Description models.
type DescriptionEntityRepoInterface interface {
	// Create Inserts a DescriptionEntity model into the database store.
	Create(ctx context.Context, input models.DescriptionEntity) (uint, error)
	// Get Returns a matching DescriptionEntity if it exists.
	Get(ctx context.Context, input GetDescriptionEntityInput) (models.DescriptionEntity, error)

	List(ctx context.Context, input ListResourceInput) (DescriptionEntityCollectionOutput, error)
}
