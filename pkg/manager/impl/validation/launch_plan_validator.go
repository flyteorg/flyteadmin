package validation

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler/validators"
	"github.com/iancoleman/orderedmap"
	"google.golang.org/grpc/codes"
)

func ValidateLaunchPlan(ctx context.Context,
	request admin.LaunchPlanCreateRequest, db repositories.RepositoryInterface,
	config runtimeInterfaces.ApplicationConfiguration, workflowInterface *core.TypedInterface) error {
	if err := ValidateIdentifier(request.Id, common.LaunchPlan); err != nil {
		return err
	}
	if err := ValidateProjectAndDomain(ctx, db, config, request.Id.Project, request.Id.Domain); err != nil {
		return err
	}
	if request.Spec == nil {
		return shared.GetMissingArgumentError(shared.Spec)
	}

	if err := ValidateIdentifier(request.Spec.WorkflowId, common.Workflow); err != nil {
		return err
	}

	if err := validateLiteralMap(request.Spec.FixedInputs, shared.FixedInputs); err != nil {
		return err
	}
	if err := validateParameterMap(request.Spec.DefaultInputs, shared.DefaultInputs); err != nil {
		return err
	}
	expectedInputs, err := checkAndFetchExpectedInputForLaunchPlan(workflowInterface.GetInputs(), request.Spec.FixedInputs, request.Spec.DefaultInputs)
	if err != nil {
		return err
	}
	if err := validateSchedule(request, expectedInputs); err != nil {
		return err
	}
	// Augment default inputs with the unbound workflow inputs.
	request.Spec.DefaultInputs = expectedInputs
	// TODO: Remove redundant validation that occurs with launch plan and the validate method for the message.
	// Ensure the notification types are validated.
	if err := request.Validate(); err != nil {
		return err
	}
	return nil
}

func validateSchedule(request admin.LaunchPlanCreateRequest, expectedInputs *core.ParameterMap) error {
	schedule := request.GetSpec().GetEntityMetadata().GetSchedule()
	if schedule.GetCronExpression() != "" || schedule.GetRate() != nil {
		for _, e := range expectedInputs.Parameters {
			if e.GetParameter().GetRequired() && e.GetName() != schedule.GetKickoffTimeInputArg() {
				return errors.NewFlyteAdminErrorf(
					codes.InvalidArgument,
					"Cannot create a launch plan with a schedule if there is an unbound required input. [%v] is required", e.GetName())
			}
		}
		if schedule.GetKickoffTimeInputArg() != "" {
			if param, ok := parameterMapEntriesToMap(expectedInputs.Parameters)[schedule.GetKickoffTimeInputArg()]; !ok {
				return errors.NewFlyteAdminErrorf(
					codes.InvalidArgument,
					"Cannot create a schedule with a KickoffTimeInputArg that does not point to a free input. [%v] is not free or does not exist.", schedule.GetKickoffTimeInputArg())
			} else if param.GetVar().GetType().GetSimple() != core.SimpleType_DATETIME {
				return errors.NewFlyteAdminErrorf(
					codes.InvalidArgument,
					"KickoffTimeInputArg must reference a datetime input. [%v] is a [%v]", schedule.GetKickoffTimeInputArg(), param.GetVar().GetType())
			}
		}
	}
	return nil
}

func checkAndFetchExpectedInputForLaunchPlan(
	workflowVariableMap *core.VariableMap, fixedInputs *core.LiteralMap, defaultInputs *core.ParameterMap) (*core.ParameterMap, error) {
	expectedInputMap := orderedmap.New()
	var workflowExpectedInputMap map[string]*core.Variable
	defaultInputMap := orderedmap.New()
	var fixedInputMap map[string]*core.Literal
	if defaultInputs != nil && len(defaultInputs.GetParameters()) > 0 {
		for _, e := range defaultInputs.GetParameters() {
			defaultInputMap.Set(e.GetName(), e.GetParameter())
		}
	}

	if fixedInputs != nil && len(fixedInputs.GetLiterals()) > 0 {
		fixedInputMap = fixedInputs.GetLiterals()
	}

	// If there are no inputs that the workflow requires, there should be none at launch plan as well
	if workflowVariableMap == nil || len(workflowVariableMap.Variables) == 0 {
		if len(defaultInputMap.Keys()) > 0 {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"invalid launch plan default inputs, expected none but found %d", len(defaultInputMap.Keys()))
		}
		if len(fixedInputMap) > 0 {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"invalid launch plan fixed inputs, expected none but found %d", len(fixedInputMap))
		}
		return &core.ParameterMap{
			Parameters: parameterOrderedMapToList(expectedInputMap),
		}, nil
	}

	workflowExpectedInputMap = variableMapEntriesToMap(workflowVariableMap.Variables)
	for _, k := range defaultInputMap.Keys() {
		value, ok := workflowExpectedInputMap[k]
		if !ok {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "unexpected default_input %s", k)
		} else {
			v, _ := defaultInputMap.Get(k)
			if !validators.AreTypesCastable(v.(*core.Parameter).GetVar().GetType(), value.GetType()) {
				return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"invalid default_input wrong type %s, expected %v, got %v instead",
					k, v.(*core.Parameter).GetVar().GetType().String(), value.GetType().String())
			}
		}
	}

	for name, fixedInput := range fixedInputMap {
		value, ok := workflowExpectedInputMap[name]
		if !ok {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "unexpected fixed_input %s", name)
		}
		inputType := validators.LiteralTypeForLiteral(fixedInput)
		if !validators.AreTypesCastable(inputType, value.GetType()) {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"invalid fixed_input wrong type %s, expected %v, got %v instead", name, value.GetType(), inputType)
		}
	}

	for name, workflowExpectedInput := range workflowExpectedInputMap {
		if value, ok := defaultInputMap.Get(name); ok {
			// If the launch plan has a default value - then use this value
			expectedInputMap.Set(name, value)
		} else if _, ok = fixedInputMap[name]; !ok {
			// If there is no mention of the input in LaunchPlan, then copy from the workflow
			expectedInputMap.Set(name, &core.Parameter{
				Var: &core.Variable{
					Type:        workflowExpectedInput.GetType(),
					Description: workflowExpectedInput.GetDescription(),
				},
				Behavior: &core.Parameter_Required{
					Required: true,
				},
			})
		}
	}
	return &core.ParameterMap{
		Parameters: parameterOrderedMapToList(expectedInputMap),
	}, nil
}

func parameterMapEntriesToMap(entries []*core.ParameterMapEntry) (parameterMap map[string]*core.Parameter) {
	parameterMap = make(map[string]*core.Parameter, len(entries))
	for _, e := range entries {
		parameterMap[e.GetName()] = e.GetParameter()
	}
	return
}

func variableMapEntriesToMap(entries []*core.VariableMapEntry) (variableMap map[string]*core.Variable) {
	variableMap = make(map[string]*core.Variable, len(entries))
	for _, v := range entries {
		variableMap[v.GetName()] = v.GetVar()
	}
	return
}

func parameterOrderedMapToList(orderedMap *orderedmap.OrderedMap) []*core.ParameterMapEntry {
	l := make([]*core.ParameterMapEntry, len(orderedMap.Keys()))
	for i, k := range orderedMap.Keys() {
		v, _ := orderedMap.Get(k)
		l[i] = &core.ParameterMapEntry{
			Name:      k,
			Parameter: v.(*core.Parameter),
		}
	}
	return l
}
