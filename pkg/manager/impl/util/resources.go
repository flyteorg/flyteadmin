package util

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	workflowengineInterfaces "github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/proto"
	"k8s.io/apimachinery/pkg/api/resource"
)

// parseQuantityNoError parses the k8s defined resource quantity gracefully masking errors.
func parseQuantityNoError(ctx context.Context, ownerID, name, value string) *resource.Quantity {
	if value == "" {
		return nil
	}
	q, err := resource.ParseQuantity(value)
	if err != nil {
		logger.Infof(ctx, "Failed to parse owner's [%s] resource [%s]'s value [%s] with err: %v", ownerID, name, value, err)
		return nil
	}

	return &q
}

// getTaskResourcesAsSet converts a list of flyteidl `ResourceEntry` messages into a singular `TaskResourceSet`.
func getTaskResourcesAsSet(ctx context.Context, identifier *core.Identifier,
	resourceEntries []*core.Resources_ResourceEntry, resourceName string) runtimeInterfaces.TaskResourceSet {

	result := runtimeInterfaces.TaskResourceSet{}
	for _, entry := range resourceEntries {
		switch entry.Name {
		case core.Resources_CPU:
			result.CPU = parseQuantityNoError(ctx, identifier.String(), fmt.Sprintf("%v.cpu", resourceName), entry.Value)
		case core.Resources_MEMORY:
			result.Memory = parseQuantityNoError(ctx, identifier.String(), fmt.Sprintf("%v.memory", resourceName), entry.Value)
		case core.Resources_EPHEMERAL_STORAGE:
			result.EphemeralStorage = parseQuantityNoError(ctx, identifier.String(),
				fmt.Sprintf("%v.ephemeral storage", resourceName), entry.Value)
		case core.Resources_GPU:
			result.GPU = parseQuantityNoError(ctx, identifier.String(), "gpu", entry.Value)
		}
	}

	return result
}

// GetCompleteTaskResourceRequirements parses the resource requests and limits from the `TaskTemplate` Container.
func GetCompleteTaskResourceRequirements(ctx context.Context, identifier *core.Identifier, task *core.CompiledTask) workflowengineInterfaces.TaskResources {
	return workflowengineInterfaces.TaskResources{
		Defaults: getTaskResourcesAsSet(ctx, identifier, task.GetTemplate().GetContainer().Resources.Requests, "requests"),
		Limits:   getTaskResourcesAsSet(ctx, identifier, task.GetTemplate().GetContainer().Resources.Limits, "limits"),
	}
}

// fromAdminProtoTaskResourceSpec parses the flyteidl `TaskResourceSpec` message into a `TaskResourceSet`.
func fromAdminProtoTaskResourceSpec(ctx context.Context, spec *admin.TaskResourceSpec) runtimeInterfaces.TaskResourceSet {
	result := runtimeInterfaces.TaskResourceSet{}
	if len(spec.Cpu) > 0 {
		result.CPU = parseQuantityNoError(ctx, "project", "cpu", spec.Cpu)
	}

	if len(spec.Memory) > 0 {
		result.Memory = parseQuantityNoError(ctx, "project", "memory", spec.Memory)
	}

	if len(spec.Storage) > 0 {
		result.Storage = parseQuantityNoError(ctx, "project", "storage", spec.Storage)
	}

	if len(spec.EphemeralStorage) > 0 {
		result.EphemeralStorage = parseQuantityNoError(ctx, "project", "ephemeral storage", spec.EphemeralStorage)
	}

	if len(spec.Gpu) > 0 {
		result.GPU = parseQuantityNoError(ctx, "project", "gpu", spec.Gpu)
	}

	return result
}

// GetTaskResources returns the most specific default and limit task resources for the specified id. This first checks
// if there is a matchable resource(s) defined, and uses the highest priority one, otherwise it falls back to using the
// flyteadmin default configured values.
func GetTaskResources(ctx context.Context, id *core.Identifier, resourceManager interfaces.ResourceInterface,
	taskResourceConfig runtimeInterfaces.TaskResourceConfiguration) workflowengineInterfaces.TaskResources {

	request := interfaces.ResourceRequest{
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	}
	if id != nil && len(id.Project) > 0 {
		request.Project = id.Project
	}
	if id != nil && len(id.Domain) > 0 {
		request.Domain = id.Domain
	}
	if id != nil && id.ResourceType == core.ResourceType_WORKFLOW && len(id.Name) > 0 {
		request.Workflow = id.Name
	}

	// change to getresourceslist
	resrc, err := resourceManager.GetResources(ctx, request)
	if err != nil {
		logger.Warningf(ctx, "Failed to fetch override values when assigning task resource default values for [%+v]: %v",
			id, err)
	}

	// gather non-nil task resource attributes
	// MergeDownTaskResources()

	var taskResourceAttributes = workflowengineInterfaces.TaskResources{}
	if resrc != nil && resrc.Attributes != nil && resrc.Attributes.GetTaskResourceAttributes() != nil {
		logger.Debugf(ctx, "Assigning task requested resources for [%+v] with resources [%+v]", id, resrc.Attributes.GetTaskResourceAttributes())
		taskResourceAttributes.Defaults = fromAdminProtoTaskResourceSpec(ctx, resrc.Attributes.GetTaskResourceAttributes().Defaults)
		taskResourceAttributes.Limits = fromAdminProtoTaskResourceSpec(ctx, resrc.Attributes.GetTaskResourceAttributes().Limits)
	} else {
		logger.Debugf(ctx, "Assigning default task requested resources for [%+v]", id)
		taskResourceAttributes = workflowengineInterfaces.TaskResources{
			Defaults: taskResourceConfig.GetDefaults(),
			Limits:   taskResourceConfig.GetLimits(),
		}
	}

	return taskResourceAttributes
}

// MergeTaskResourceSpec merges two TaskResourceSpecs. No notion of quantity comparison, just whether or not fields
// are empty strings.
func MergeTaskResourceSpec(high, low *admin.TaskResourceSpec) *admin.TaskResourceSpec {
	if high == nil && low == nil {
		return nil
	} else if high == nil && low != nil {
		res := proto.Clone(low).(*admin.TaskResourceSpec)
		return res
	} else if high != nil && low == nil {
		res := proto.Clone(high).(*admin.TaskResourceSpec)
		return res
	}

	res := proto.Clone(high).(*admin.TaskResourceSpec)
	if res.GetCpu() == "" && low.GetCpu() != "" {
		res.Cpu = low.Cpu
	}
	if res.GetGpu() == "" && low.GetGpu() != "" {
		res.Gpu = low.Gpu
	}
	if res.GetMemory() == "" && low.GetMemory() != "" {
		res.Memory = low.Memory
	}
	if res.GetEphemeralStorage() == "" && low.GetEphemeralStorage() != "" {
		res.EphemeralStorage = low.EphemeralStorage
	}
	return res
}

// MergeTaskResourceAttributes will merge without error, taking non-empty strings from lower priority
// and filling in missing higher priority fields.
func MergeTaskResourceAttributes(high, low admin.TaskResourceAttributes) admin.TaskResourceAttributes {
	res := proto.Clone(&high).(*admin.TaskResourceAttributes)
	res.Defaults = MergeTaskResourceSpec(high.GetDefaults(), low.GetDefaults())
	res.DefaultLimits = MergeTaskResourceSpec(high.GetDefaultLimits(), low.GetDefaultLimits())
	res.Limits = MergeTaskResourceSpec(high.GetLimits(), low.GetLimits())
	return *res
}

func ConstrainTaskResourceSpec(spec admin.TaskResourceSpec, maxes admin.TaskResourceSpec) admin.TaskResourceSpec {
	res := proto.Clone(&spec).(*admin.TaskResourceSpec)
	if maxes.GetCpu() != "" && spec.GetCpu() != "" {
		maxCpu := resource.MustParse(maxes.GetCpu())
		specCpu := resource.MustParse(spec.GetCpu())
		if specCpu.Cmp(maxCpu) == 1 {
			res.Cpu = maxes.GetCpu()
		}
	}
	if maxes.GetGpu() != "" && spec.GetGpu() != "" {
		maxGpu := resource.MustParse(maxes.GetGpu())
		specGpu := resource.MustParse(spec.GetGpu())
		if specGpu.Cmp(maxGpu) == 1 {
			res.Gpu = maxes.GetGpu()
		}
	}
	if maxes.GetMemory() != "" && spec.GetMemory() != "" {
		maxMemory := resource.MustParse(maxes.GetMemory())
		specMemory := resource.MustParse(spec.GetMemory())
		if specMemory.Cmp(maxMemory) == 1 {
			res.Memory = maxes.GetMemory()
		}
	}
	if maxes.GetEphemeralStorage() != "" && spec.GetEphemeralStorage() != "" {
		maxEphemeralStorage := resource.MustParse(maxes.GetEphemeralStorage())
		specEphemeralStorage := resource.MustParse(spec.GetEphemeralStorage())
		if specEphemeralStorage.Cmp(maxEphemeralStorage) == 1 {
			res.EphemeralStorage = maxes.GetEphemeralStorage()
		}
	}
	return *res
}

func ConformLimits(attr admin.TaskResourceAttributes) admin.TaskResourceAttributes {
	maxes := admin.TaskResourceSpec{}
	if attr.GetLimits() != nil {
		maxes = *attr.GetLimits()
	}
	if attr.GetDefaultLimits() != nil {
		x := ConstrainTaskResourceSpec(*attr.GetDefaultLimits(), maxes)
		attr.DefaultLimits = &x
		maxes = *attr.GetDefaultLimits()
	}
	if attr.GetDefaults() != nil {
		x := ConstrainTaskResourceSpec(*attr.GetDefaultLimits(), maxes)
		attr.Defaults = &x
	}
	return attr
}

// MergeDownTaskResources does not today check that the defaults are below the limits when setting, therefore
// Go through the list from high to low priority, first merge, and then resolve inconsistencies around quantities.
//   - If set, must be limit >= default limit >= default
func MergeDownTaskResources(highToLowPriorityTaskResourceAttributes ...admin.TaskResourceAttributes) *admin.TaskResourceAttributes {
	// Merge each one down, checking each condition
	merged := admin.TaskResourceAttributes{}
	for _, attr := range highToLowPriorityTaskResourceAttributes {
		merged = MergeTaskResourceAttributes(merged, attr)
	}
	merged = ConformLimits(merged)
	return &merged
}
