package runtime

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/logger"
)

const namespaceMappingKey = "namespace_mapping"

var namespaceMappingConfig = config.MustRegisterSection(namespaceMappingKey, &interfaces.NamespaceMappingConfig{})

type NamespaceMappingConfigurationProvider struct{}

func (p *NamespaceMappingConfigurationProvider) GetNamespaceMappingConfig() string {
	if namespaceMappingConfig != nil && namespaceMappingConfig.GetConfig() != nil {
		return namespaceMappingConfig.GetConfig().(*interfaces.NamespaceMappingConfig).Mapping
	}
	logger.Warningf(context.Background(), "Failed to find namespace mapping in config, defaulting to <project>-<domain>")
	return ""
}
