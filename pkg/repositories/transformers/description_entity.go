package transformers

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
)

// CreateDescriptionEntityModel Transforms a TaskCreateRequest to a Description entity model
func CreateDescriptionEntityModel(
	request admin.DescriptionEntityCreateRequest,
	digest []byte) (models.DescriptionEntity, error) {
	ctx := context.Background()

	// TODO: offload the LongDescription in to a separate file if value exceed 4KB, and update URI in LongDescription
	var longDescriptionBytes []byte
	var sourceCode models.SourceCode
	var err error

	if request.DescriptionEntity.LongDescription != nil {
		longDescriptionBytes, err = proto.Marshal(request.DescriptionEntity.LongDescription)
		if err != nil {
			logger.Errorf(ctx, "Failed to marshal LongDescription with error: %v", err)
			return models.DescriptionEntity{}, err
		}
	}

	if request.DescriptionEntity.LongDescription != nil {
		longDescriptionBytes, err = proto.Marshal(request.DescriptionEntity.LongDescription)
		if err != nil {
			logger.Errorf(ctx, "Failed to marshal LongDescription with error: %v", err)
			return models.DescriptionEntity{}, err
		}
	}
	if request.DescriptionEntity.SourceCode != nil {
		sourceCode = models.SourceCode{Link: request.DescriptionEntity.SourceCode.Link}
	}

	return models.DescriptionEntity{
		DescriptionEntityKey: models.DescriptionEntityKey{
			ResourceType: request.Id.ResourceType,
			Project:      request.Id.Project,
			Domain:       request.Id.Domain,
			Name:         request.Id.Name,
			Version:      request.Id.Version,
		},
		Digest:           digest,
		ShortDescription: request.DescriptionEntity.ShortDescription,
		LongDescription:  longDescriptionBytes,
		SourceCode:       sourceCode,
	}, nil
}

func FromDescriptionEntityModel(descriptionEntityModel models.DescriptionEntity) (*admin.DescriptionEntity, error) {

	longDescription := admin.LongDescription{}
	err := proto.Unmarshal(descriptionEntityModel.LongDescription, &longDescription)
	if err != nil {
		return nil, errors.NewFlyteAdminError(codes.Internal, "failed to unmarshal longDescription")
	}

	return &admin.DescriptionEntity{
		ShortDescription: descriptionEntityModel.ShortDescription,
		LongDescription:  &longDescription,
		SourceCode:       &admin.SourceCode{Link: descriptionEntityModel.Link},
	}, nil
}
