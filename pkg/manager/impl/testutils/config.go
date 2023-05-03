package testutils

import (
	"github.com/flyteorg/flyteadmin/pkg/common"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	"k8s.io/apimachinery/pkg/api/resource"
)

func GetApplicationConfigWithDefaultDomains() runtimeInterfaces.ApplicationConfiguration {
	config := runtimeMocks.MockApplicationProvider{}
	config.SetDomainsConfig(runtimeInterfaces.DomainsConfig{
		{
			ID:   "development",
			Name: "development",
		},
		{
			ID:   "staging",
			Name: "staging",
		},
		{
			ID:   "production",
			Name: "production",
		},
		{
			ID:   "domain",
			Name: "domain",
		},
	})
	config.SetRemoteDataConfig(runtimeInterfaces.RemoteDataConfig{
		Scheme: common.Local, SignedURL: runtimeInterfaces.SignedURL{
			Enabled: true,
		}})
	return &config
}

func GetPtr(quantity resource.Quantity) *resource.Quantity {
	return &quantity
}

func GetSampleTaskResourceConfiguration() runtimeInterfaces.TaskResourceConfiguration {
	resourceDefaults := runtimeInterfaces.TaskResourceSet{
		CPU:    GetPtr(resource.MustParse("200m")),
		Memory: GetPtr(resource.MustParse("200Gi")),
		GPU:    GetPtr(resource.MustParse("0")),
	}
	resourceLimits := runtimeInterfaces.TaskResourceSet{
		CPU:              GetPtr(resource.MustParse("300m")),
		Memory:           GetPtr(resource.MustParse("500Gi")),
		EphemeralStorage: GetPtr(resource.MustParse("10Gi")),
	}

	return runtimeMocks.NewMockTaskResourceConfiguration(resourceDefaults, resourceLimits)
}

func GetMockConfiguration() runtimeInterfaces.Configuration {
	appConfig := GetApplicationConfigWithDefaultDomains()
	taskResourceConfig := GetSampleTaskResourceConfiguration()
	return runtimeMocks.NewMockConfigurationProvider(appConfig, nil, nil, taskResourceConfig, nil, nil)
}
