package validation

import (
	"context"
	"errors"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"

    "github.com/golang/protobuf/proto"
)

func TestValidateSignalGetOrCreateRequest(t *testing.T) {
	ctx := context.TODO()

	t.Run("Happy", func(t *testing.T) {
		request := admin.SignalGetOrCreateRequest{
			Id: &core.SignalIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				SignalId: "signal",
			},
			Type: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_BOOLEAN,
				},
			},
		}
		assert.NoError(t, ValidateSignalGetOrCreateRequest(ctx, request))
	})

	t.Run("MissingSignalIdentifier", func(t *testing.T) {
		request := admin.SignalGetOrCreateRequest{
			Type: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_BOOLEAN,
				},
			},
		}
		assert.EqualError(t, ValidateSignalGetOrCreateRequest(ctx, request), "missing id")
	})

	t.Run("InvalidSignalIdentifier", func(t *testing.T) {
		request := admin.SignalGetOrCreateRequest{
			Id: &core.SignalIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
			},
			Type: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_BOOLEAN,
				},
			},
		}
		assert.EqualError(t, ValidateSignalGetOrCreateRequest(ctx, request), "missing signal_id")
	})

	t.Run("MissingExecutionIdentifier", func(t *testing.T) {
		request := admin.SignalGetOrCreateRequest{
			Id: &core.SignalIdentifier{
				SignalId: "signal",
			},
			Type: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_BOOLEAN,
				},
			},
		}
		assert.EqualError(t, ValidateSignalGetOrCreateRequest(ctx, request), "missing execution_id")
	})

	t.Run("InvalidExecutionIdentifier", func(t *testing.T) {
		request := admin.SignalGetOrCreateRequest{
			Id: &core.SignalIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Domain:  "domain",
					Name:    "name",
				},
				SignalId: "signal",
			},
			Type: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_BOOLEAN,
				},
			},
		}
		assert.EqualError(t, ValidateSignalGetOrCreateRequest(ctx, request), "missing project")
	})

	t.Run("MissingType", func(t *testing.T) {
		request := admin.SignalGetOrCreateRequest{
			Id: &core.SignalIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				SignalId: "signal",
			},
		}
		assert.EqualError(t, ValidateSignalGetOrCreateRequest(ctx, request), "missing type")
	})
}

func TestValidateSignalUpdateRequest(t *testing.T) {
	ctx := context.TODO()

	booleanType := &core.LiteralType{
		Type: &core.LiteralType_Simple{
			Simple: core.SimpleType_BOOLEAN,
		},
	}
	typeBytes, _ := proto.Marshal(booleanType)

	repo := repositoryMocks.NewMockRepository()
	repo.SignalRepo().(*repositoryMocks.MockSignalRepo).SetGetCallback(
		func(input models.SignalKey) (models.Signal, error) {
			return models.Signal{
				SignalKey: input,
				Type:      typeBytes,
			}, nil
		})

	t.Run("Happy", func(t *testing.T) {
		request := admin.SignalSetRequest{
			Id: &core.SignalIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				SignalId: "signal",
			},
			Value: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Boolean{
									Boolean: false,
								},
							},
						},
					},
				},
			},
		}
		assert.NoError(t, ValidateSignalSetRequest(ctx, repo, request))
	})

	t.Run("MissingValue", func(t *testing.T) {
		request := admin.SignalSetRequest{
			Id: &core.SignalIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				SignalId: "signal",
			},
		}
		assert.EqualError(t, ValidateSignalSetRequest(ctx, repo, request), "missing value")
	})

	t.Run("MissingSignal", func(t *testing.T) {
		repo := repositoryMocks.NewMockRepository()
		repo.SignalRepo().(*repositoryMocks.MockSignalRepo).SetGetCallback(
			func(input models.SignalKey) (models.Signal, error) {
				return models.Signal{}, errors.New("foo")
			})

		request := admin.SignalSetRequest{
			Id: &core.SignalIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				SignalId: "signal",
			},
			Value: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Boolean{
									Boolean: false,
								},
							},
						},
					},
				},
			},
		}
		assert.EqualError(t, ValidateSignalSetRequest(ctx, repo, request),
			"failed to validate that signal [{{project domain name} signal}] exists, err: [foo]")
	})

	t.Run("InvalidType", func(t *testing.T) {
		integerType := &core.LiteralType{
			Type: &core.LiteralType_Simple{
				Simple: core.SimpleType_INTEGER,
			},
		}
		typeBytes, _ := proto.Marshal(integerType)

		repo := repositoryMocks.NewMockRepository()
		repo.SignalRepo().(*repositoryMocks.MockSignalRepo).SetGetCallback(
			func(input models.SignalKey) (models.Signal, error) {
				return models.Signal{
					SignalKey: input,
					Type:      typeBytes,
				}, nil
			})

		request := admin.SignalSetRequest{
			Id: &core.SignalIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				SignalId: "signal",
			},
			Value: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Boolean{
									Boolean: false,
								},
							},
						},
					},
				},
			},
		}
		assert.EqualError(t, ValidateSignalSetRequest(ctx, repo, request),
			"requested signal value [scalar:<primitive:<boolean:false > > ] is not castable to existing signal type [[8 1]]")
	})
}
