package config

import (
	"fmt"
	"github.com/lyft/flytestdlib/config"
)

const SectionKey = "server"

//go:generate pflags ServerConfig

type ServerConfig struct {
	HTTPPort   int                   `json:"httpPort" pflag:",On which http port to serve admin"`
	GrpcPort   int                   `json:"grpcPort" pflag:",On which grpc port to serve admin"`
	KubeConfig string                `json:"kube-config" pflag:",Path to kubernetes client config file."`
	Master     string                `json:"master" pflag:",The address of the Kubernetes API server."`
	Security   ServerSecurityOptions `json:"security"`
}

type ServerSecurityOptions struct {
	Secure  bool         `json:"secure"`
	Ssl     SslOptions   `json:"ssl"`
	UseAuth bool         `json:"useAuth"`
	Oauth   OauthOptions `json:"oauth"`
}

type SslOptions struct {
	CertificateFile string `json:"certificateFile"`
	KeyFile         string `json:"keyFile"`
}

type OauthOptions struct {
	ClientId         string `json:"clientId"`
	ClientSecretFile string `json:"clientSecretFile"`
	JwksUrl          string `json:"jwksUrl"`
	Issuer           string `json:"issuer"`
	AuthorizeUrl     string `json:"authorizeUrl"`
	TokenUrl         string `json:"tokenUrl"`
	RedirectUrl      string `json:"redirectUrl"`
}

var serverConfig = config.MustRegisterSection(SectionKey, &ServerConfig{})

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
