package interfaces

import (
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/config"
)

type TierName = string

type QualityOfServiceSpec struct {
	QueueingBudget config.Duration `json:"queueingBudget"`
}

type QualityOfServiceConfig struct {
	TierExecutionValues map[TierName]QualityOfServiceSpec `json:"tierExecutionValues"`
	DefaultTiers        map[DomainName]TierName           `json:"defaultTiers"`
}

type QualityOfServiceConfiguration interface {
	GetTierExecutionValues() map[core.QualityOfService_Tier]core.QualityOfServiceSpec
	GetDefaultTiers() map[DomainName]core.QualityOfService_Tier
}
