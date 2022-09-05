package interfaces

import (
	"io/ioutil"

	"github.com/flyteorg/flytestdlib/config"

	"github.com/pkg/errors"
)

// KubeClientConfig contains the configuration used by flytepropeller to configure its internal Kubernetes Client.
type KubeClientConfig struct {
	// QPS indicates the maximum QPS to the master from this client.
	// If it's zero, the created RESTClient will use DefaultQPS: 5
	QPS float32 `json:"qps" pflag:"-,Max QPS to the master for requests to KubeAPI. 0 defaults to 5."`
	// Maximum burst for throttle.
	// If it's zero, the created RESTClient will use DefaultBurst: 10.
	Burst int `json:"burst" pflag:",Max burst rate for throttle. 0 defaults to 10"`
	// The maximum length of time to wait before giving up on a server request. A value of zero means no timeout.
	Timeout config.Duration `json:"timeout" pflag:",Max duration allowed for every request to KubeAPI before giving up. 0 implies no timeout."`
}

// Holds details about a cluster used for workflow execution.
type ClusterConfig struct {
	Name             string            `json:"name"`
	Endpoint         string            `json:"endpoint"`
	Auth             Auth              `json:"auth"`
	Enabled          bool              `json:"enabled"`
	KubeClientConfig *KubeClientConfig `json:"kube-client-config"`
}

type Auth struct {
	Type      string `json:"type"`
	TokenPath string `json:"tokenPath"`
	CertPath  string `json:"certPath"`
}

type ClusterEntity struct {
	ID     string  `json:"id"`
	Weight float32 `json:"weight"`
}

func (auth Auth) GetCA() ([]byte, error) {
	cert, err := ioutil.ReadFile(auth.CertPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read k8s CA cert from configured path")
	}
	return cert, nil
}

func (auth Auth) GetToken() (string, error) {
	token, err := ioutil.ReadFile(auth.TokenPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to read k8s bearer token from configured path")
	}
	return string(token), nil
}

type Clusters struct {
	ClusterConfigs  []ClusterConfig            `json:"clusterConfigs"`
	LabelClusterMap map[string][]ClusterEntity `json:"labelClusterMap"`
}

//go:generate mockery -name ClusterConfiguration -case=underscore -output=../mocks -case=underscore

// Provides values set in runtime configuration files.
// These files can be changed without requiring a full server restart.
type ClusterConfiguration interface {
	// Returns clusters defined in runtime configuration files.
	GetClusterConfigs() []ClusterConfig

	// Returns label cluster map for routing
	GetLabelClusterMap() map[string][]ClusterEntity
}
