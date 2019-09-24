// Handles translating gRPC request & response objects to and from repository model objects
package transformers

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"
)

// Transforms a TaskCreateRequest to a task model
func CreateTaskModel(
	request admin.TaskCreateRequest,
	taskClosure admin.TaskClosure,
	digest []byte) (models.Task, error) {
	closureBytes, err := proto.Marshal(&taskClosure)
	if err != nil {
		return models.Task{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize task closure")
	}
	return models.Task{
		TaskKey: models.TaskKey{
			Project: request.Id.Project,
			Domain:  request.Id.Domain,
			Name:    request.Id.Name,
			Version: request.Id.Version,
		},
		Closure: closureBytes,
		Digest:  digest,
	}, nil
}

func FromTaskModel(taskModel models.Task) (admin.Task, error) {
	taskClosure := &admin.TaskClosure{}
	err := proto.Unmarshal(taskModel.Closure, taskClosure)
	if err != nil {
		return admin.Task{}, errors.NewFlyteAdminError(codes.Internal, "failed to unmarshal clsoure")
	}
	createdAt, err := ptypes.TimestampProto(taskModel.CreatedAt)
	if err != nil {
		return admin.Task{}, errors.NewFlyteAdminErrorf(codes.Internal, "failed to serialize created at")
	}
	taskClosure.CreatedAt = createdAt
	id := core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      taskModel.Project,
		Domain:       taskModel.Domain,
		Name:         taskModel.Name,
		Version:      taskModel.Version,
	}
	return admin.Task{
		Id:      &id,
		Closure: taskClosure,
	}, nil
}

func FromTaskModels(taskModels []models.Task) ([]*admin.Task, error) {
	tasks := make([]*admin.Task, len(taskModels))
	for idx, taskModel := range taskModels {
		task, err := FromTaskModel(taskModel)
		if err != nil {
			return nil, err
		}
		tasks[idx] = &task
	}
	return tasks, nil
}

func FromTaskModelsToIdentifiers(taskModels []models.Task) []*admin.NamedEntityIdentifier {
	ids := make([]*admin.NamedEntityIdentifier, len(taskModels))
	for i, taskModel := range taskModels {
		ids[i] = &admin.NamedEntityIdentifier{
			Project: taskModel.Project,
			Domain:  taskModel.Domain,
			Name:    taskModel.Name,
		}
	}
	return ids
}
