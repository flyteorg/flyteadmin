package runtime

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/logger"
)

const (
	namespaceMappingKey = "namespace_mapping"
	domainVariable      = "domain"
)

var namespaceMappingConfig = config.MustRegisterSection(namespaceMappingKey, &interfaces.NamespaceMappingConfig{})

type NamespaceMappingConfigurationProvider struct{}

func (p *NamespaceMappingConfigurationProvider) GetNamespaceMappingConfig() NamespaceMapping {
	var string mapping
	if namespaceMappingConfig != nil && namespaceMappingConfig.GetConfig() != nil {
		mapping := namespaceMappingConfig.GetConfig().(*interfaces.NamespaceMappingConfig).Mapping
	}

	switch mapping {
	case domainVariable:
		logger.Warningf(context.Background(), "Failed to find namespace mapping in config, defaulting to <project>-<domain>")
		return Domain
	default:
		logger.Warningf(context.Background(), "Unsupported value for namespace_mapping in config, defaulting to <project>-<domain>")
		return ProjectDomain
	}
	logger.Warningf(context.Background(), "Failed to find namespace mapping in config, defaulting to <project>-<domain>")
	return ProjectDomain
}

func NewNamespaceMappingConfigurationProvider() interfaces.NamespaceMappingConfiguration {
	return &NamespaceMappingConfigurationProvider{}
}
