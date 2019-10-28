package config

import (
	"fmt"
	config2 "github.com/lyft/flyteadmin/pkg/auth/config"
	"github.com/lyft/flytestdlib/config"
)

const SectionKey = "server"

//go:generate pflags ServerConfig --default-var=defaultServerConfig

type ServerConfig struct {
	HTTPPort   int                   `json:"httpPort" pflag:",On which http port to serve admin"`
	GrpcPort   int                   `json:"grpcPort" pflag:",On which grpc port to serve admin"`
	KubeConfig string                `json:"kube-config" pflag:",Path to kubernetes client config file."`
	Master     string                `json:"master" pflag:",The address of the Kubernetes API server."`
	Security   ServerSecurityOptions `json:"security"`
}

type ServerSecurityOptions struct {
	Secure  bool                 `json:"secure"`
	Ssl     SslOptions           `json:"ssl"`
	UseAuth bool                 `json:"useAuth"`
	Oauth   config2.OAuthOptions `json:"oauth"`
}

type SslOptions struct {
	CertificateFile string `json:"certificateFile"`
	KeyFile         string `json:"keyFile"`
}

var defaultServerConfig = &ServerConfig{
	Security: ServerSecurityOptions{
		Oauth: config2.OAuthOptions{
			// Please see the comments in this struct's definition for more information
			HttpAuthorizationHeader: "placeholder",
			GrpcAuthorizationHeader: "flyte-authorization",
		},
	},
}
var serverConfig = config.MustRegisterSection(SectionKey, defaultServerConfig)

func GetConfig() *ServerConfig {
	return serverConfig.GetConfig().(*ServerConfig)
}

func SetConfig(s *ServerConfig) {
	if err := serverConfig.SetConfig(s); err != nil {
		panic(err)
	}
}

func (s ServerConfig) GetHostAddress() string {
	return fmt.Sprintf(":%d", s.HTTPPort)
}

func (s ServerConfig) GetGrpcHostAddress() string {
	return fmt.Sprintf(":%d", s.GrpcPort)
}

func init() {
	SetConfig(&ServerConfig{})
}
