package util

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"google.golang.org/grpc/codes"

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
func parseQuantityNoError(ctx context.Context, value string) *resource.Quantity {
	if value == "" {
		return nil
	}
	q, err := resource.ParseQuantity(value)
	if err != nil {
		logger.Infof(ctx, "Failed to value [%s] with err: %v", value, err)
		return nil
	}

	return &q
}

func ConvertTaskResourceSetToCoreResources(resources runtimeInterfaces.TaskResourceSet) []*core.Resources_ResourceEntry {
	var resourceEntries = make([]*core.Resources_ResourceEntry, 0)

	if resources.CPU != nil {
		resourceEntries = append(resourceEntries, &core.Resources_ResourceEntry{
			Name:  core.Resources_CPU,
			Value: resources.CPU.String(),
		})
	}

	if resources.Memory != nil {
		resourceEntries = append(resourceEntries, &core.Resources_ResourceEntry{
			Name:  core.Resources_MEMORY,
			Value: resources.Memory.String(),
		})
	}

	if resources.EphemeralStorage != nil {
		resourceEntries = append(resourceEntries, &core.Resources_ResourceEntry{
			Name:  core.Resources_EPHEMERAL_STORAGE,
			Value: resources.EphemeralStorage.String(),
		})
	}

	if resources.GPU != nil {
		resourceEntries = append(resourceEntries, &core.Resources_ResourceEntry{
			Name:  core.Resources_GPU,
			Value: resources.GPU.String(),
		})
	}
	return resourceEntries
}

// GetTaskResourcesAndCoalesce takes a set of Flyte IDL ResourceEntry's and fills in missing fields from the coalesce
func GetTaskResourcesAndCoalesce(ctx context.Context,
	resourceEntries []*core.Resources_ResourceEntry, coalesce runtimeInterfaces.TaskResourceSet) runtimeInterfaces.TaskResourceSet {

	result := runtimeInterfaces.TaskResourceSet{}
	for _, entry := range resourceEntries {
		q := parseQuantityNoError(ctx, entry.Value)
		switch entry.Name {
		case core.Resources_CPU:
			if q == nil && coalesce.CPU != nil {
				result.CPU = coalesce.CPU
			}
		case core.Resources_MEMORY:
			if q == nil && coalesce.Memory != nil {
				result.Memory = coalesce.Memory
			}
		case core.Resources_EPHEMERAL_STORAGE:
			if q == nil && coalesce.EphemeralStorage != nil {
				result.EphemeralStorage = coalesce.EphemeralStorage
			}
		case core.Resources_GPU:
			if q == nil && coalesce.GPU != nil {
				result.GPU = coalesce.GPU
			}
		}
	}

	return result
}

// getTaskResourcesAsSet converts a list of flyteidl `ResourceEntry` messages into a singular `TaskResourceSet`.
func getTaskResourcesAsSet(ctx context.Context,
	resourceEntries []*core.Resources_ResourceEntry) runtimeInterfaces.TaskResourceSet {

	result := runtimeInterfaces.TaskResourceSet{}
	for _, entry := range resourceEntries {
		switch entry.Name {
		case core.Resources_CPU:
			result.CPU = parseQuantityNoError(ctx, entry.Value)
		case core.Resources_MEMORY:
			result.Memory = parseQuantityNoError(ctx, entry.Value)
		case core.Resources_EPHEMERAL_STORAGE:
			result.EphemeralStorage = parseQuantityNoError(ctx, entry.Value)
		case core.Resources_GPU:
			result.GPU = parseQuantityNoError(ctx, entry.Value)
		}
	}

	return result
}

// fromAdminProtoTaskResourceSpec parses the flyteidl `TaskResourceSpec` message into a `TaskResourceSet`.
func fromAdminProtoTaskResourceSpec(ctx context.Context, spec *admin.TaskResourceSpec) runtimeInterfaces.TaskResourceSet {
	result := runtimeInterfaces.TaskResourceSet{}
	if len(spec.Cpu) > 0 {
		result.CPU = parseQuantityNoError(ctx, spec.Cpu)
	}

	if len(spec.Memory) > 0 {
		result.Memory = parseQuantityNoError(ctx, spec.Memory)
	}

	if len(spec.Storage) > 0 {
		result.Storage = parseQuantityNoError(ctx, spec.Storage)
	}

	if len(spec.EphemeralStorage) > 0 {
		result.EphemeralStorage = parseQuantityNoError(ctx, spec.EphemeralStorage)
	}

	if len(spec.Gpu) > 0 {
		result.GPU = parseQuantityNoError(ctx, spec.Gpu)
	}

	return result
}

// GetTaskResources returns a merged set of all the requests, and limits for the given request.
// This will combine all layers matched and merge missing resource types. That is, if CPU is set at the project level
// and memory is set at the project/domain/workflow level, this will return both.
// Admin default system wide configuration is also merged in.
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

	var attrs = make([]admin.TaskResourceAttributes, 0)

	// Get list of all task resources.
	resrc, err := resourceManager.GetResourcesList(ctx, request)
	if err != nil {
		ec, ok := err.(errors.FlyteAdminError)
		if ok && ec.Code() == codes.NotFound {
			logger.Debug(ctx, "HandleGetTaskResourceRequest did not find any task resources, falling back")
		} else {
			logger.Warningf(ctx, "Failed to fetch override values when assigning task resource default values for [%+v]: %v",
				id, err)
		}
	} else if resrc != nil {
		logger.Debugf(ctx, "GetTaskResources returned [%d] task resources, combining with config", len(resrc.AttributeList))
		for _, rr := range resrc.AttributeList {
			if rr.GetTaskResourceAttributes() != nil {
				attrs = append(attrs, *rr.GetTaskResourceAttributes())
			}
		}
	}

	attrs = append(attrs, taskResourceConfig.GetAsAttribute())
	responseAttributes := MergeDownTaskResources(attrs...)

	var taskResourceAttributes = workflowengineInterfaces.TaskResources{}

	if responseAttributes.GetLimits() != nil {
		taskResourceAttributes.Limits = fromAdminProtoTaskResourceSpec(ctx, responseAttributes.GetLimits())
	}
	if responseAttributes.GetDefaults() != nil {
		taskResourceAttributes.Defaults = fromAdminProtoTaskResourceSpec(ctx, responseAttributes.GetDefaults())
	}

	return taskResourceAttributes
}

// MergeTaskResourceSpec merges two TaskResourceSpecs. No notion of quantity comparison, just whether or not fields
// are empty strings.
func MergeTaskResourceSpec(high, low *admin.TaskResourceSpec) *admin.TaskResourceSpec {
	if high == nil && low == nil {
		return nil
	} else if high == nil && low != nil {
		// Return nil if all fields are empty strings. This is just done for this case, preserving the behavior that if
		// the higher priority thing is nil, an empty lower priority object shouldn't make it non-nil.
		if low.Cpu == "" && low.Gpu == "" && low.Memory == "" && low.EphemeralStorage == "" {
			return nil
		}
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
	res.Limits = MergeTaskResourceSpec(high.GetLimits(), low.GetLimits())
	return *res
}

// ConstrainTaskResourceSpec takes two TaskResourceSpecs and returns a new one, limiting the first argument, to the
// values of the second arg (maxes), for each resource type, if it exists in the maxes. This function parses the
// strings into resource.Quantity objects, and compares them using k8s Cmp.
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

func quantityToString(q *resource.Quantity) string {
	if q == nil {
		return ""
	}
	return q.String()
}

// ConstrainTaskResourceSet is the same as ConstrainTaskResourceSpec, but for TaskResourceSet.
// Converts to TaskResourceSpec, and then calls the limiting function for that, and convert back.
func ConstrainTaskResourceSet(ctx context.Context, spec runtimeInterfaces.TaskResourceSet, maxes runtimeInterfaces.TaskResourceSet) runtimeInterfaces.TaskResourceSet {

	specAsResourceSpec := admin.TaskResourceSpec{
		Cpu:              quantityToString(spec.CPU),
		Gpu:              quantityToString(spec.GPU),
		Memory:           quantityToString(spec.Memory),
		EphemeralStorage: quantityToString(spec.EphemeralStorage),
	}
	maxesAsResourceSpec := admin.TaskResourceSpec{
		Cpu:              quantityToString(spec.CPU),
		Gpu:              quantityToString(spec.GPU),
		Memory:           quantityToString(spec.Memory),
		EphemeralStorage: quantityToString(spec.EphemeralStorage),
	}

	r := ConstrainTaskResourceSpec(specAsResourceSpec, maxesAsResourceSpec)
	resourceSet := fromAdminProtoTaskResourceSpec(ctx, &r)
	return resourceSet
}

func ConformLimits(attr admin.TaskResourceAttributes) admin.TaskResourceAttributes {
	maxes := admin.TaskResourceSpec{}
	if attr.GetLimits() != nil {
		maxes = *attr.GetLimits()
	}
	if attr.GetDefaults() != nil {
		x := ConstrainTaskResourceSpec(*attr.GetDefaults(), maxes)
		attr.Defaults = &x
	}
	return attr
}

// MergeDownTaskResources does not today check that the defaults are below the limits when setting, therefore
// go through the list from high to low priority, first merge the various types, and then resolve inconsistencies
// around quantities.
//   - If set, must be limit >= default
func MergeDownTaskResources(highToLowPriorityTaskResourceAttributes ...admin.TaskResourceAttributes) *admin.TaskResourceAttributes {
	// Merge each one down, checking each condition
	merged := admin.TaskResourceAttributes{}
	for _, attr := range highToLowPriorityTaskResourceAttributes {
		merged = MergeTaskResourceAttributes(merged, attr)
	}
	merged = ConformLimits(merged)
	return &merged
}
