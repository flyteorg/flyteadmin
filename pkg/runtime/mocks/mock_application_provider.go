package mocks

import (
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/database"
)

type MockApplicationProvider struct {
	dbConfig             database.DbConfig
	topLevelConfig       interfaces.ApplicationConfig
	schedulerConfig      interfaces.SchedulerConfig
	remoteDataConfig     interfaces.RemoteDataConfig
	notificationsConfig  interfaces.NotificationsConfig
	domainsConfig        interfaces.DomainsConfig
	externalEventsConfig interfaces.ExternalEventsConfig
	cloudEventConfig     interfaces.CloudEventsConfig
}

func (p *MockApplicationProvider) GetDbConfig() *database.DbConfig {
	return &p.dbConfig
}

func (p *MockApplicationProvider) SetDbConfig(dbConfig database.DbConfig) {
	p.dbConfig = dbConfig
}

func (p *MockApplicationProvider) GetTopLevelConfig() *interfaces.ApplicationConfig {
	return &p.topLevelConfig
}

func (p *MockApplicationProvider) SetTopLevelConfig(topLevelConfig interfaces.ApplicationConfig) {
	p.topLevelConfig = topLevelConfig
}

func (p *MockApplicationProvider) GetSchedulerConfig() *interfaces.SchedulerConfig {
	return &p.schedulerConfig
}

func (p *MockApplicationProvider) SetSchedulerConfig(schedulerConfig interfaces.SchedulerConfig) {
	p.schedulerConfig = schedulerConfig
}

func (p *MockApplicationProvider) GetRemoteDataConfig() *interfaces.RemoteDataConfig {
	return &p.remoteDataConfig
}

func (p *MockApplicationProvider) SetRemoteDataConfig(remoteDataConfig interfaces.RemoteDataConfig) {
	p.remoteDataConfig = remoteDataConfig
}

func (p *MockApplicationProvider) GetNotificationsConfig() *interfaces.NotificationsConfig {
	return &p.notificationsConfig
}

func (p *MockApplicationProvider) SetNotificationsConfig(notificationsConfig interfaces.NotificationsConfig) {
	p.notificationsConfig = notificationsConfig
}

func (p *MockApplicationProvider) GetDomainsConfig() *interfaces.DomainsConfig {
	return &p.domainsConfig
}

func (p *MockApplicationProvider) SetDomainsConfig(domainsConfig interfaces.DomainsConfig) {
	p.domainsConfig = domainsConfig
}

func (p *MockApplicationProvider) SetExternalEventsConfig(externalEventsConfig interfaces.ExternalEventsConfig) {
	p.externalEventsConfig = externalEventsConfig
}

func (p *MockApplicationProvider) GetExternalEventsConfig() *interfaces.ExternalEventsConfig {
	return &p.externalEventsConfig
}

func (p *MockApplicationProvider) SetCloudEventsConfig(cloudEventConfig interfaces.CloudEventsConfig) {
	p.cloudEventConfig = cloudEventConfig
}

func (p *MockApplicationProvider) GetCloudEventsConfig() *interfaces.CloudEventsConfig {
	return &p.cloudEventConfig
}

func (p *MockApplicationProvider) GetAsWorkflowExecutionAttribute() admin.WorkflowExecutionConfig {
	a := p.GetTopLevelConfig()

	wec := admin.WorkflowExecutionConfig{
		MaxParallelism: a.GetMaxParallelism(),
		OverwriteCache: a.GetOverwriteCache(),
		Interruptible:  a.GetInterruptible(),
	}

	// For the others, we only add the field when the field is set in the config.
	if a.GetSecurityContext().RunAs.GetK8SServiceAccount() != "" || a.GetSecurityContext().RunAs.GetIamRole() != "" {
		wec.SecurityContext = a.GetSecurityContext()
	}
	if a.GetRawOutputDataConfig().OutputLocationPrefix != "" {
		wec.RawOutputDataConfig = a.GetRawOutputDataConfig()
	}
	if len(a.GetLabels().Values) > 0 {
		wec.Labels = a.GetLabels()
	}
	if len(a.GetAnnotations().Values) > 0 {
		wec.Annotations = a.GetAnnotations()
	}

	return wec
}
