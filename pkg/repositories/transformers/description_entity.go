package transformers

import (
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// CreateDescriptionEntityModel Transforms a TaskCreateRequest to a Description entity model
func CreateDescriptionEntityModel(
	request admin.DescriptionEntityCreateRequest,
	digest []byte) (models.DescriptionEntity, error) {
	return models.DescriptionEntity{
		DescriptionEntityKey: models.DescriptionEntityKey{
			Project: request.Id.Project,
			Domain:  request.Id.Domain,
			Name:    request.Id.Name,
			Version: request.Id.Version,
		},
		Digest:           digest,
		ShortDescription: request.DescriptionEntity.ShortDescription,
	}, nil
}

func FromDescriptionEntityModel(descriptionEntityModel models.DescriptionEntity) admin.DescriptionEntity {
	return admin.DescriptionEntity{
		ShortDescription: descriptionEntityModel.ShortDescription,
	}
}
