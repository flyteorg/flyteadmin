package util

import (
	"context"
	"strings"
	"time"
	"unicode"

	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/manager/impl/validation"
	"github.com/lyft/flyteadmin/pkg/manager/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories"
	repositoryInterfaces "github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/logger"
	"google.golang.org/grpc/codes"
)

const maxNodeIDLength = 63

var defaultRetryStrategy = core.RetryStrategy{
	Retries: 3,
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

func getBinding(literal *core.Literal) *core.BindingData {
	if literal.GetScalar() != nil {
		return &core.BindingData{
			Value: &core.BindingData_Scalar{
				Scalar: literal.GetScalar(),
			},
		}
	} else if literal.GetCollection() != nil {
		bindings := make([]*core.BindingData, len(literal.GetCollection().Literals))
		for idx, subLiteral := range literal.GetCollection().Literals {
			bindings[idx] = getBinding(subLiteral)
		}
		return &core.BindingData{
			Value: &core.BindingData_Collection{
				Collection: &core.BindingDataCollection{
					Bindings: bindings,
				},
			},
		}
	} else if literal.GetMap() != nil {
		bindings := make(map[string]*core.BindingData)
		for key, subLiteral := range literal.GetMap().Literals {
			bindings[key] = getBinding(subLiteral)
		}
		return &core.BindingData{
			Value: &core.BindingData_Map{
				Map: &core.BindingDataMap{
					Bindings: bindings,
				},
			},
		}
	}
	return nil
}

func generateBindingsFromOutputs(outputs core.VariableMap, nodeID string) []*core.Binding {
	bindings := make([]*core.Binding, 0, len(outputs.Variables))
	for key := range outputs.Variables {
		binding := &core.Binding{
			Var: key,
			Binding: &core.BindingData{
				Value: &core.BindingData_Promise{
					Promise: &core.OutputReference{
						NodeId: nodeID,
						Var:    key,
					},
				},
			},
		}

		bindings = append(bindings, binding)
	}
	logger.Warningf(context.TODO(), "Generated outputs: [%+v]", bindings)
	return bindings
}

func generateBindingsFromInputs(inputTemplate core.VariableMap, inputs core.LiteralMap) ([]*core.Binding, error) {
	logger.Warningf(context.TODO(), "generating inputs from [%+v]", inputTemplate)
	bindings := make([]*core.Binding, 0, len(inputTemplate.Variables))
	for key, val := range inputTemplate.Variables {
		binding := &core.Binding{
			Var: key,
		}
		var bindingData core.BindingData
		if val.Type.GetSimple() != core.SimpleType_NONE {
			if inputs.Literals[key] != nil {
				bindingData = core.BindingData{
					Value: &core.BindingData_Scalar{
						Scalar: inputs.Literals[key].GetScalar(),
					},
				}
			}

		} else if val.Type.GetSchema() != nil {
			if inputs.Literals[key] != nil && inputs.Literals[key].GetScalar() != nil {
				bindingData = core.BindingData{
					Value: &core.BindingData_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Schema{
								Schema: inputs.Literals[key].GetScalar().GetSchema(),
							},
						},
					},
				}
			}
		} else if val.Type.GetCollectionType() != nil {
			if inputs.Literals[key] != nil && inputs.Literals[key].GetCollection() != nil &&
				inputs.Literals[key].GetCollection().GetLiterals() != nil {
				collectionBindings := make([]*core.BindingData, len(inputs.Literals[key].GetCollection().GetLiterals()))
				for idx, literal := range inputs.Literals[key].GetCollection().GetLiterals() {
					collectionBindings[idx] = getBinding(literal)

				}
				bindingData = core.BindingData{
					Value: &core.BindingData_Collection{
						Collection: &core.BindingDataCollection{
							Bindings: collectionBindings,
						},
					},
				}
			}
		} else if val.Type.GetMapValueType() != nil {
			if inputs.Literals[key] != nil && inputs.Literals[key].GetMap() != nil &&
				inputs.Literals[key].GetMap().Literals != nil {
				bindingDataMap := make(map[string]*core.BindingData)
				for k, v := range inputs.Literals[key].GetMap().Literals {
					bindingDataMap[k] = getBinding(v)
				}

				bindingData = core.BindingData{
					Value: &core.BindingData_Map{
						Map: &core.BindingDataMap{
							Bindings: bindingDataMap,
						},
					},
				}
			}
		} else if val.Type.GetBlob() != nil {
			if inputs.Literals[key] != nil && inputs.Literals[key].GetScalar() != nil {
				bindingData = core.BindingData{
					Value: &core.BindingData_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Blob{
								Blob: inputs.Literals[key].GetScalar().GetBlob(),
							},
						},
					},
				}
			}
		} else {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "Unrecognized value type [%+v]", val.GetType())
		}
		binding.Binding = &bindingData
		bindings = append(bindings, binding)
	}
	logger.Debugf(context.TODO(), "generated inputs [%+v]", bindings)
	return bindings, nil
}

func CreateOrGetWorkflowModel(
	ctx context.Context, request admin.ExecutionCreateRequest, db repositories.RepositoryInterface,
	workflowManager interfaces.WorkflowInterface, namedEntityManager interfaces.NamedEntityInterface, taskIdentifier *core.Identifier,
	task *admin.Task) (*models.Workflow, error) {
	workflowModel, err := db.WorkflowRepo().Get(ctx, repositoryInterfaces.GetResourceInput{
		Project: taskIdentifier.Project,
		Domain:  taskIdentifier.Domain,
		Name:    taskIdentifier.Name,
		Version: taskIdentifier.Version,
	})

	logger.Warningf(ctx, "TODO - debug: 1")
	if err != nil {
		if ferr, ok := err.(errors.FlyteAdminError); !ok || ferr.Code() != codes.NotFound {
			return nil, err
		}
		// If we got this far, there is no existing workflow. Create a skeleton one now.

		logger.Warningf(ctx, "TODO - debug: 2")
		var requestInputs = core.LiteralMap{
			Literals: make(map[string]*core.Literal),
		}
		if request.Inputs != nil {
			requestInputs = *request.Inputs
		}
		generatedInputs, err := generateBindingsFromInputs(*task.Closure.CompiledTask.Template.Interface.Inputs, requestInputs)
		if err != nil {
			logger.Debugf(ctx, "Failed to generate requestInputs from task input bindings: %v", err)
			return nil, err
		}
		logger.Warningf(ctx, "TODO - debug: 3")
		workflowSpec := admin.WorkflowSpec{
			Template: &core.WorkflowTemplate{
				Id: &core.Identifier{
					ResourceType: core.ResourceType_WORKFLOW,
					Project:      taskIdentifier.Project,
					Domain:       taskIdentifier.Domain,
					Name:         taskIdentifier.Name,
					Version:      taskIdentifier.Version,
				},
				Interface: task.Closure.CompiledTask.Template.Interface,
				Nodes: []*core.Node{
					{
						Id: generateNodeNameFromTask(taskIdentifier.Name),
						Metadata: &core.NodeMetadata{
							Name:    generateNodeNameFromTask(taskIdentifier.Name),
							Retries: &defaultRetryStrategy,
							Timeout: defaultTimeout,
						},
						Inputs: generatedInputs,
						Target: &core.Node_TaskNode{
							TaskNode: &core.TaskNode{
								Reference: &core.TaskNode_ReferenceId{
									ReferenceId: taskIdentifier,
								},
							},
						},
					},
				},

				Outputs: generateBindingsFromOutputs(*task.Closure.CompiledTask.Template.Interface.Outputs, generateNodeNameFromTask(taskIdentifier.Name)),
			},
		}

		logger.Warningf(ctx, "TODO - creating workflow with spec: %+v", workflowSpec)
		_, err = workflowManager.CreateWorkflow(ctx, admin.WorkflowCreateRequest{
			Id:   workflowSpec.Template.Id,
			Spec: &workflowSpec,
		})
		if err != nil {
			logger.Debugf(ctx, "Failed to create skeleton workflow: %v", err)
			return nil, err
		}
		// Now, set the newly created skeleton workflow to 'ARCHIVED'.
		logger.Warningf(ctx, "TODO - debug: 5")
		_, err = namedEntityManager.UpdateNamedEntity(ctx, admin.NamedEntityUpdateRequest{
			ResourceType: core.ResourceType_WORKFLOW,
			Id: &admin.NamedEntityIdentifier{
				Project: taskIdentifier.Project,
				Domain:  taskIdentifier.Domain,
				Name:    taskIdentifier.Name,
			},
			Metadata: &admin.NamedEntityMetadata{State: admin.NamedEntityState_NAMED_ENTITY_ARCHIVED},
		})
		logger.Warningf(ctx, "TODO - debug: 6")
		if err != nil {
			logger.Debug(ctx, "Failed to set skeleton workflow state to archived: %v", err)
			return nil, err
		}
		workflowModel, err = db.WorkflowRepo().Get(ctx, repositoryInterfaces.GetResourceInput{
			Project: taskIdentifier.Project,
			Domain:  taskIdentifier.Domain,
			Name:    taskIdentifier.Name,
			Version: taskIdentifier.Version,
		})
		logger.Warningf(ctx, "TODO - debug: 7")
		if err != nil {
			// This is unexpected - at this point we've successfully just created the skeleton workflow.
			logger.Warningf(ctx, "Failed to fetch newly created workflow model from db store: %v", err)
			return nil, err
		}
	}

	return &workflowModel, nil
}

func CreateOrGetLaunchPlan(ctx context.Context,
	db repositories.RepositoryInterface, config runtimeInterfaces.Configuration, identifier *core.Identifier,
	workflowInterface *core.TypedInterface, workflowID uint) (*admin.LaunchPlan, error) {
	var launchPlan *admin.LaunchPlan
	var err error
	launchPlan, err = GetLaunchPlan(ctx, db, *identifier)
	if err != nil {
		if ferr, ok := err.(errors.FlyteAdminError); !ok || ferr.Code() != codes.NotFound {
			return nil, err
		}

		// Create launch plan.
		generatedCreateLaunchPlanReq := admin.LaunchPlanCreateRequest{
			Id: &core.Identifier{
				ResourceType: core.ResourceType_LAUNCH_PLAN,
				Project:      identifier.Project,
				Domain:       identifier.Domain,
				Name:         identifier.Name,
				Version:      identifier.Version,
			},
			Spec: &admin.LaunchPlanSpec{
				WorkflowId: &core.Identifier{
					ResourceType: core.ResourceType_WORKFLOW,
					Project:      identifier.Project,
					Domain:       identifier.Domain,
					Name:         identifier.Name,
					Version:      identifier.Version,
				},
				EntityMetadata: &admin.LaunchPlanMetadata{},
				DefaultInputs:  &core.ParameterMap{},
				FixedInputs:    &core.LiteralMap{},
				Labels:         &admin.Labels{},
				Annotations:    &admin.Annotations{},
				Auth:           nil, // TODO: reconcile auth from CreateExecution request
			},
		}
		if err := validation.ValidateLaunchPlan(ctx, generatedCreateLaunchPlanReq, db, config.ApplicationConfiguration(), workflowInterface); err != nil {
			logger.Debugf(ctx, "could not create launch plan: %+v, request failed validation with err: %v", identifier, err)
			return nil, err
		}
		transformedLaunchPlan := transformers.CreateLaunchPlan(generatedCreateLaunchPlanReq, workflowInterface.Outputs)
		launchPlan = &transformedLaunchPlan
		launchPlanDigest, err := GetLaunchPlanDigest(ctx, launchPlan)
		if err != nil {
			logger.Errorf(ctx, "failed to compute launch plan digest for [%+v] with err: %v", launchPlan.Id, err)
			return nil, err
		}
		logger.Warningf(ctx, "TODO - debug: launch plan: %+v", launchPlan)
		launchPlanModel, err :=
			transformers.CreateLaunchPlanModel(*launchPlan, workflowID, launchPlanDigest, admin.LaunchPlanState_INACTIVE)
		if err != nil {
			logger.Errorf(ctx,
				"Failed to transform launch plan model [%+v], and workflow outputs [%+v] with err: %v",
				identifier, workflowInterface.Outputs, err)
			return nil, err
		}
		logger.Warningf(ctx, "TODO - debug: launch plan model: %+v", launchPlanModel)
		err = db.LaunchPlanRepo().Create(ctx, launchPlanModel) // Where not exists in case of transactions?
		if err != nil {
			logger.Errorf(ctx, "Failed to save launch plan model %+v with err: %v", identifier, err)
			return nil, err
		}
	}

	return launchPlan, nil
}
