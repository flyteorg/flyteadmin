package util

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/manager/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteadmin/pkg/repositories/transformers"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"
	"strings"
	"time"
	"unicode"
	"github.com/lyft/flyteadmin/pkg/repositories"
	repositoryInterfaces "github.com/lyft/flyteadmin/pkg/repositories/interfaces"
)

const maxNodeIDLength = 63
var defaultRetryStrategy = core.RetryStrategy{
	Retries:3,
}
var defaultTimeout = ptypes.DurationProto(24 * time.Hour)

func generateNodeNameFromTask(taskName string) string {
	if len(taskName) >= maxNodeIDLength {
		taskName = taskName[:maxNodeIDLength]
	}
	nodeNameBuilder := strings.Builder{}
	for _, i := range taskName {
		if i == '-' || unicode.IsLetter(i) || unicode.IsNumber(i) {
			nodeNameBuilder.WriteRune(i)
		}
	}
	return nodeNameBuilder.String()
}

func getBinding(literal *core.Literal) *core.BindingData{
	if literal.GetScalar() != nil {
		return &core.BindingData{
			Value:                &core.BindingData_Scalar{
				Scalar: literal.GetScalar(),
			},
		}
	} else if literal.GetCollection() != nil {
		bindings := make([]*core.BindingData, len(literal.GetCollection().Literals))
		for idx, subLiteral := range literal.GetCollection().Literals {
			bindings[idx] = getBinding(subLiteral)
		}
		return &core.BindingData{
			Value:                &core.BindingData_Collection{
				Collection: &core.BindingDataCollection{
					Bindings:             bindings,
				},
			},
		}
	} else if literal.GetMap() != nil {
		bindings := make(map[string]*core.BindingData)
		for key, subLiteral := range literal.GetMap().Literals{
			bindings[key] = getBinding(subLiteral)
		}
		return &core.BindingData{
			Value:                &core.BindingData_Map{
				Map: &core.BindingDataMap{
					Bindings:             bindings,
				},
			},
		}
	}
	return nil
}

func generateBindingsFromOutputs(outputs core.VariableMap) []*core.Binding{
	bindings := make([]*core.Binding, len(outputs.Variables))
	for key, _ := range outputs.Variables {
		binding := &core.Binding{
			Var: key,
		}

		bindings = append(bindings, binding)
	}
	return bindings
}

func generateBindingsFromInputs(inputTemplate core.VariableMap, inputs *core.LiteralMap) ([]*core.Binding, error){
	bindings := make([]*core.Binding, len(inputTemplate.Variables))
	for key, val := range inputTemplate.Variables{
		binding := &core.Binding{
			Var: key,
		}
		var bindingData core.BindingData
		if val.Type.GetSimple() != core.SimpleType_NONE {
			bindingData = core.BindingData{
				Value: &core.BindingData_Scalar{
					Scalar: inputs.Literals[key].GetScalar(),
				},
			}
		} else if val.Type.GetSchema() != nil {
			bindingData = core.BindingData{
				Value: &core.BindingData_Scalar{
					Scalar: &core.Scalar{
						Value:                &core.Scalar_Schema{
							Schema:inputs.Literals[key].GetScalar().GetSchema(),
						},
					},
				},
			}
		} else if val.Type.GetCollectionType() != nil {
			collectionBindings := make([]*core.BindingData, len(inputs.Literals[key].GetCollection().GetLiterals()))
			for idx, literal := range inputs.Literals[key].GetCollection().GetLiterals(){
				collectionBindings[idx] = getBinding(literal)

			}
			bindingData = core.BindingData{
				Value: &core.BindingData_Collection{
					Collection: &core.BindingDataCollection{
						Bindings:             collectionBindings,
					},
				},
			}
		} else if val.Type.GetMapValueType() != nil {
			bindingDataMap := make(map[string]*core.BindingData)
			for k, v := range inputs.Literals[key].GetMap().Literals {
				bindingDataMap[k] = getBinding(v)
			}

			bindingData = core.BindingData{
				Value: &core.BindingData_Map{
					Map: &core.BindingDataMap{
						Bindings:             bindingDataMap,
					},
				},
			}
		} else if val.Type.GetBlob() != nil {
			bindingData = core.BindingData{
				Value: &core.BindingData_Scalar{
					Scalar: &core.Scalar{
						Value:                &core.Scalar_Blob{
							Blob:inputs.Literals[key].GetScalar().GetBlob(),
						},
					},
				},
			}
		} else {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "Unrecognized value type [%+v]", val.GetType())
		}
		binding.Binding = &bindingData
		bindings = append(bindings, binding)
	}
	return bindings, nil
}


func CreateDefaultObjectsForSingleTaskExecution(
	ctx context.Context, request admin.ExecutionCreateRequest, db repositories.RepositoryInterface,
	workflowManager interfaces.WorkflowInterface, namedEntityManager interfaces.NamedEntityInterface) (*models.Task, error) {
	taskIdentifier := request.Spec.LaunchPlan

	// 1) verify the reference task exists.
	taskModel, err := db.TaskRepo().Get(ctx, repositoryInterfaces.GetResourceInput{
		Project: taskIdentifier.Project,
		Domain: taskIdentifier.Domain,
		Name: taskIdentifier.Name,
		Version: taskIdentifier.Version,
	})
	if err != nil {
		return nil, err
	}
	task, err := transformers.FromTaskModel(taskModel)
	if err != nil {
		return nil, err
	}

	// 2) See if a corresponding skeleton workflow exists and if not, create one on the fly.
	_, err = db.WorkflowRepo().Get(ctx, repositoryInterfaces.GetResourceInput{
		Project: taskIdentifier.Project,
		Domain: taskIdentifier.Domain,
		Name: taskIdentifier.Name,
		Version: taskIdentifier.Version,
	})
	if err != nil {
		if ferr, ok := err.(errors.FlyteAdminError); !ok || ferr.Code() != codes.NotFound{
			return nil, err
		}

		generatedInputs, err := generateBindingsFromInputs(*task.Closure.CompiledTask.Template.Interface.Inputs, request.Inputs)
		if err != nil {
			return nil, err
		}
		// If we got this far, there is no existing workflow. Create a skeleton one now.
		workflowSpec := admin.WorkflowSpec{
			Template:             &core.WorkflowTemplate{
				Id:                   &core.Identifier{
					ResourceType: core.ResourceType_WORKFLOW,
					Project: taskIdentifier.Project,
					Domain: taskIdentifier.Domain,
					Name: taskIdentifier.Name,
					Version: taskIdentifier.Version,
				},
				Interface: task.Closure.CompiledTask.Template.Interface,
				Nodes:                []*core.Node{
					{
						Id: generateNodeNameFromTask(taskIdentifier.Name),
						Metadata: &core.NodeMetadata{
							Name: generateNodeNameFromTask(taskIdentifier.Name),
							Retries: &defaultRetryStrategy,
							Timeout: defaultTimeout,
						},
						Inputs: generatedInputs,
						Target: &core.Node_TaskNode{
							TaskNode: &core.TaskNode{
								Reference:            &core.TaskNode_ReferenceId{
									ReferenceId: taskIdentifier,
								},
							},
						},
					},
				},

				Outputs: generateBindingsFromOutputs(*task.Closure.CompiledTask.Template.Interface.Outputs),
			},
		}

		_, err = workflowManager.CreateWorkflow(ctx, admin.WorkflowCreateRequest{
			Id:                   workflowSpec.Template.Id,
			Spec:                 &workflowSpec,
		})
		if err != nil {
			return nil, err
		}
		// Now, set the newly created skeleton workflow to 'ARCHIVED'.
		_, err = namedEntityManager.UpdateNamedEntity(ctx, admin.NamedEntityUpdateRequest{
			ResourceType: core.ResourceType_WORKFLOW,
			Id: &admin.NamedEntityIdentifier{
				Project: taskIdentifier.Project,
				Domain: taskIdentifier.Domain,
				Name: taskIdentifier.Name,
			},
			Metadata: &admin.NamedEntityMetadata{State: admin.NamedEntityState_NAMED_ENTITY_ARCHIVED},
		})
		if err != nil {
			return nil, err
		}
	}

	// 3. Create a default launch plan (if necessary)
	// TODO(katrina)
	return &taskModel, nil
}