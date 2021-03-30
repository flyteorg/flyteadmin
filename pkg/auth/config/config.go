package config

import (
	"time"

	"github.com/flyteorg/flytestdlib/config"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = &Config{
		// Please see the comments in this struct's definition for more information
		HTTPAuthorizationHeader: "flyte-authorization",
		GrpcAuthorizationHeader: "flyte-authorization",
		UserAuth: UserAuthConfig{
			OpenID: OpenIDOptions{
				// Default claims that should be supported by any OIdC server. Refer to https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
				// for a complete list.
				Scopes: []string{
					"openid",
					"profile",
				},
			},
		},
		AppAuth: OAuth2Options{
			Issuer: OAuth2Issuer{
				Issuer:              "https://localhost:8088",
				AccessTokenLifespan: config.Duration{Duration: 30 * time.Minute},
			},
		},
	}

	cfgSection = config.MustRegisterSection("auth", defaultConfig)
)

type Config struct {
	// These settings are for non-SSL authentication modes, where Envoy is handling SSL termination
	// This is not yet used, but this is the HTTP variant of the setting below.
	HTTPAuthorizationHeader string `json:"httpAuthorizationHeader"`

	// In order to support deployments of this Admin service where Envoy is terminating SSL connections, the metadata
	// header name cannot be "authorization", which is the standard metadata name. Envoy has special handling for that
	// name. Instead, there is a gRPC interceptor, GetAuthenticationCustomMetadataInterceptor, that will translate
	// incoming metadata headers with this config setting's name, into that standard header
	GrpcAuthorizationHeader string `json:"grpcAuthorizationHeader"`

	// To help ease migration, it was helpful to be able to only selectively enforce authentication.  The
	// dimension that made the most sense to cut by at time of writing is HTTP vs gRPC as the web UI mainly used HTTP
	// and the backend used mostly gRPC.  Cutting by individual endpoints is another option but it possibly falls more
	// into the realm of authorization rather than authentication.
	DisableForHTTP bool `json:"disableForHttp"`
	DisableForGrpc bool `json:"disableForGrpc"`

	UserAuth UserAuthConfig `json:"userAuth" pflag:",Defines Auth options for users."`
	AppAuth  OAuth2Options  `json:"appAuth" pflag:",Defines Auth options for apps. UserAuth must be enabled for AppAuth to work."`
}

type OAuth2Issuer struct {
	Issuer              string          `json:"issuer" pflag:",Defines the issuer to use when issuing and validating tokens. The default value is https://<requestUri.HostAndPort>/"`
	AccessTokenLifespan config.Duration `json:"accessTokenLifespan" pflag:",Defines the lifespan of issued access tokens."`
}

type OAuth2Options struct {
	Issuer OAuth2Issuer `json:"issuer"`
}

type UserAuthConfig struct {
	// This is where the user will be redirected to at the end of the flow, but you should not use it. Instead,
	// the initial /login handler should be called with a redirect_url parameter, which will get saved to a cookie.
	// This setting will only be used when that cookie is missing.
	// See the login handler code for more comments.
	RedirectURL config.URL `json:"redirectUrl"`

	OpenID OpenIDOptions `json:"openId" pflag:",OpenID Configuration for User Auth"`
	// Possibly add basicAuth & SAML/p support.
}

type OpenIDOptions struct {
	// The client ID for Admin in your IDP
	// See https://tools.ietf.org/html/rfc6749#section-2.2 for more information
	ClientID string `json:"clientId"`

	// The client secret used in the exchange of the authorization code for the token.
	// https://tools.ietf.org/html/rfc6749#section-2.3
	ClientSecretFile string `json:"clientSecretFile"`

	// This should be the base url of the authorization server that you are trying to hit. With Okta for instance, it
	// will look something like https://company.okta.com/oauth2/abcdef123456789/
	// TODO: Convert all the URLs in this config to the config.URL type
	BaseURL config.URL `json:"baseUrl"`

	// This is the callback URL that will be sent to the IDP authorize endpoint. It is likely that your IDP application
	// needs to have this URL whitelisted before using.
	CallbackURL config.URL `json:"callbackUrl"`

	// Provides a list of scopes to request from the IDP when authenticating. Default value requests claims that should
	// be supported by any OIdC server. Refer to https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims for
	// a complete list. Other providers might support additional scopes that you can define in a config.
	Scopes []string `json:"scopes"`

	// The list of audiences to allow when doing token validation. This should typically be the public endpoint that
	// Admin Service is accessed as (e.g. https://admin.mycompany.com)
	AcceptedAudiences []string `json:"aud"`
}

func GetConfig() *Config {
	return cfgSection.GetConfig().(*Config)
}
