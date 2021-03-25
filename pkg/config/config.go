package config

import (
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/auth/oauthserver"

	config2 "github.com/flyteorg/flyteadmin/pkg/auth/config"
	"github.com/flyteorg/flytestdlib/config"
)

const SectionKey = "server"

//go:generate pflags ServerConfig --default-var=defaultServerConfig

type ServerConfig struct {
	HTTPPort             int                     `json:"httpPort" pflag:",On which http port to serve admin"`
	GrpcPort             int                     `json:"grpcPort" pflag:",On which grpc port to serve admin"`
	GrpcServerReflection bool                    `json:"grpcServerReflection" pflag:",Enable GRPC Server Reflection"`
	KubeConfig           string                  `json:"kube-config" pflag:",Path to kubernetes client config file."`
	Master               string                  `json:"master" pflag:",The address of the Kubernetes API server."`
	Security             ServerSecurityOptions   `json:"security"`
	ThirdPartyConfig     ThirdPartyConfigOptions `json:"thirdPartyConfig"`
}

type ServerSecurityOptions struct {
	Secure      bool                      `json:"secure"`
	Ssl         SslOptions                `json:"ssl"`
	UseAuth     bool                      `json:"useAuth"`
	OpenID      config2.OpenIDOptions     `json:"openid"`
	OAuth2      oauthserver.OAuth2Options `json:"oauth2"`
	AuditAccess bool                      `json:"auditAccess"`

	// These options are here to allow deployments where the Flyte UI (Console) is served from a different domain/port.
	// Note that CORS only applies to Admin's API endpoints. The health check endpoint for instance is unaffected.
	// Please obviously evaluate security concerns before turning this on.
	AllowCors bool `json:"allowCors"`
	// Defines origins which are allowed to make CORS requests. This list should _not_ contain "*", as that
	// will make CORS header responses incompatible with the `withCredentials=true` setting.
	AllowedOrigins []string `json:"allowedOrigins"`
	// These are the Access-Control-Request-Headers that the server will respond to.
	// By default, the server will allow Accept, Accept-Language, Content-Language, and Content-Type.
	// User this setting to add any additional headers which are needed
	AllowedHeaders []string `json:"allowedHeaders"`
}

type SslOptions struct {
	CertificateFile string `json:"certificateFile"`
	KeyFile         string `json:"keyFile"`
}

var defaultServerConfig = &ServerConfig{
	Security: ServerSecurityOptions{
		OpenID: config2.OpenIDOptions{
			// Please see the comments in this struct's definition for more information
			HTTPAuthorizationHeader: "flyte-authorization",
			GrpcAuthorizationHeader: "flyte-authorization",
			// Default claims that should be supported by any OIdC server. Refer to https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
			// for a complete list.
			Scopes: []string{
				"openid",
				"profile",
			},
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
