package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

type GetDescriptionEntityInput struct {
	ResourceType core.ResourceType
	Project      string
	Domain       string
	Name         string
	Version      string
}

// Parameters for querying multiple resources.
type ListDescriptionEntityInput struct {
	ListResourceInput
	Project      string
	Domain       string
	ResourceType core.ResourceType
}

type DescriptionEntityCollectionOutput struct {
	Entities []models.DescriptionEntity
}

// DescriptionEntityRepoInterface Defines the interface for interacting with Description models.
type DescriptionEntityRepoInterface interface {
	// Create Inserts a DescriptionEntity model into the database store.
	Create(ctx context.Context, input models.DescriptionEntity) (uint, error)
	// Get Returns a matching DescriptionEntity if it exists.
	Get(ctx context.Context, input models.DescriptionEntityKey) (models.DescriptionEntity, error)
	// TODO: List DescriptionEntity
}
