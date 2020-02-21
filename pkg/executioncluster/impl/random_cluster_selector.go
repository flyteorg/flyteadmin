package impl

import (
	"context"
	"fmt"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/manager/impl/resources"
	managerInterfaces "github.com/lyft/flyteadmin/pkg/manager/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories"
	"github.com/lyft/flytestdlib/logger"
	"google.golang.org/grpc/codes"
	"hash/fnv"
	"math/rand"

	"github.com/lyft/flyteadmin/pkg/executioncluster"
	"github.com/lyft/flyteadmin/pkg/executioncluster/interfaces"

	"github.com/lyft/flytestdlib/random"

	runtime "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/promutils"
)

// Implementation of Random cluster selector
// Selects cluster based on weights and domains.
type RandomClusterSelector struct {
	labelWeightedRandomMap map[string]random.WeightedRandomList
	executionTargetMap     map[string]executioncluster.ExecutionTarget
	resourceManager        managerInterfaces.ResourceInterface
}

func getRandSource(seed string) (rand.Source, error) {
	h := fnv.New64a()
	_, err := h.Write([]byte(seed))
	if err != nil {
		return nil, err
	}
	hashedSeed := int64(h.Sum64())
	return rand.NewSource(hashedSeed), nil
}

func getExecutionTargetMap(scope promutils.Scope, executionTargetProvider interfaces.ExecutionTargetProvider, clusterConfig runtime.ClusterConfiguration) (map[string]executioncluster.ExecutionTarget, error) {
	executionTargetMap := make(map[string]executioncluster.ExecutionTarget)
	for _, cluster := range clusterConfig.GetClusterConfigs() {
		if _, ok := executionTargetMap[cluster.Name]; ok {
			return nil, fmt.Errorf("duplicate clusters for name %s", cluster.Name)
		}
		executionTarget, err := executionTargetProvider.GetExecutionTarget(scope, cluster)
		if err != nil {
			return nil, err
		}
		executionTargetMap[cluster.Name] = *executionTarget
	}
	return executionTargetMap, nil
}

func getLabeledWeightedRandomForCluster(ctx context.Context, scope promutils.Scope,
	clusterConfig runtime.ClusterConfiguration, executionTargetMap map[string]executioncluster.ExecutionTarget) (map[string]random.WeightedRandomList, error) {
	labeledWeightedRandomMap := make(map[string]random.WeightedRandomList)
	for label, clusterEntities := range clusterConfig.GetLabelClusterMap() {
		entries := make([]random.Entry, len(clusterEntities))
		for _, clusterEntity := range clusterEntities {
			cluster := executionTargetMap[clusterEntity.ID]
			// If cluster is not enabled, it is not eligible for selection
			if !cluster.Enabled {
				continue
			}
			targetEntry := random.Entry{
				Item:   cluster,
				Weight: clusterEntity.Weight,
			}
			entries = append(entries, targetEntry)
		}
		weightedRandomList, err := random.NewWeightedRandom(ctx, entries)
		if err != nil {
			return nil, err
		}
		labeledWeightedRandomMap[label] = weightedRandomList
	}
	return labeledWeightedRandomMap, nil
}

func (s RandomClusterSelector) GetAllValidTargets() []executioncluster.ExecutionTarget {
	v := make([]executioncluster.ExecutionTarget, 0)
	for _, value := range s.executionTargetMap {
		if value.Enabled {
			v = append(v, value)
		}
	}
	return v
}

func (s RandomClusterSelector) GetTarget(ctx context.Context, spec *executioncluster.ExecutionTargetSpec) (*executioncluster.ExecutionTarget, error) {
	if spec == nil || (spec.TargetID == "") {
		return nil, fmt.Errorf("invalid executionTargetSpec %v", spec)
	}
	if spec.TargetID != "" {
		if val, ok := s.executionTargetMap[spec.TargetID]; ok {
			return &val, nil
		}
		return nil, fmt.Errorf("invalid cluster target %s", spec.TargetID)
	}
	resource, err := s.resourceManager.GetResource(ctx, managerInterfaces.ResourceRequest{
		Project: spec.Project,
		Domain: spec.Domain,
		Workflow: spec.Workflow,
		LaunchPlan: spec.LaunchPlan,
	})
	if err != nil {
		if flyteAdminError, ok := err.(errors.FlyteAdminError); !ok || flyteAdminError.Code() != codes.NotFound {
			return nil, err
		}
		return nil, err
	}
	if resource != nil {
		resource.Attributes.g
	}
	if spec.ExecutionID != nil {
		// Change to resource manager.

		if weightedRandomList, ok := s.labelWeightedRandomMap[spec.ExecutionID.GetDomain()]; ok {
			executionName := spec.ExecutionID.GetName()
			if executionName != "" {
				randSrc, err := getRandSource(executionName)
				if err != nil {
					return nil, err
				}
				result, err := weightedRandomList.GetWithSeed(randSrc)
				if err != nil {
					return nil, err
				}
				execTarget := result.(executioncluster.ExecutionTarget)
				return &execTarget, nil
			}
			execTarget := weightedRandomList.Get().(executioncluster.ExecutionTarget)
			return &execTarget, nil
		}

		// Get random from the entire list
	}

	return nil, fmt.Errorf("invalid executionTargetSpec %v", *spec)

}

func NewRandomClusterSelector(scope promutils.Scope, clusterConfig runtime.ClusterConfiguration, executionTargetProvider interfaces.ExecutionTargetProvider, db repositories.RepositoryInterface) (interfaces.ClusterInterface, error) {
	executionTargetMap, err := getExecutionTargetMap(scope, executionTargetProvider, clusterConfig)
	if err != nil {
		return nil, err
	}
	labelWeightedRandomMap, err := getLabeledWeightedRandomForCluster(context.Background(), scope, clusterConfig, executionTargetMap)
	if err != nil {
		return nil, err
	}
	return &RandomClusterSelector{
		labelWeightedRandomMap: labelWeightedRandomMap,
		executionTargetMap:     executionTargetMap,
		resourceManager:        resources.NewResourceManager(db),
	}, nil
}
