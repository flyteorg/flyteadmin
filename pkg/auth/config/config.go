package config

import (
	"time"

	"github.com/ory/fosite"

	"github.com/flyteorg/flytestdlib/config"
)

//go:generate pflags Config --default-var=defaultConfig
//go:generate enumer --type=AuthorizationServerType --trimprefix=AuthorizationServerType -json

const (
	SecretOIdCClientSecret = "oidc_client_secret"

	SecretCookieHashKey  = "cookie_hash_key"
	SecretCookieBlockKey = "cookie_block_key"

	// Base64 encoded secret of exactly 32 bytes
	SecretClaimSymmetricKey = "claim_symmetric_key"

	// PrivateKey is used to sign JWT tokens. The default strategy uses RS256 (RSA Signature with SHA-256)
	SecretTokenSigningRSAKey = "token_rsa_key.pem"
	// PrivateKey that was used to sign old JWT tokens. The default strategy uses RS256 (RSA Signature with SHA-256)
	// This is used to support key rotation. When present, it'll only be used to validate incoming tokens. New tokens
	// will not be issued using this key.
	SecretOldTokenSigningRSAKey = "token_rsa_key_old.pem"
)

// AuthorizationServerType defines the type of Authorization Server to use.
type AuthorizationServerType int

const (
	// AuthorizationServerTypeSelf determines that FlyteAdmin should act as the authorization server to serve
	// OAuth2 token requests
	AuthorizationServerTypeSelf AuthorizationServerType = iota

	// AuthorizationServerTypeExternal determines that FlyteAdmin should rely on an external authorization server (e.g.
	// Okta) to serve OAuth2 token requests
	AuthorizationServerTypeExternal
)

var (
	defaultConfig = &Config{
		// Please see the comments in this struct's definition for more information
		HTTPAuthorizationHeader: "flyte-authorization",
		GrpcAuthorizationHeader: "flyte-authorization",
		UserAuth: UserAuthConfig{
			CookieHashKeySecretName:  SecretCookieHashKey,
			CookieBlockKeySecretName: SecretCookieBlockKey,
			OpenID: OpenIDOptions{
				ClientSecretName: SecretOIdCClientSecret,
				// Default claims that should be supported by any OIdC server. Refer to https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
				// for a complete list.
				Scopes: []string{
					"openid",
					"profile",
				},
			},
		},
		AppAuth: OAuth2Options{
			AuthServerType: AuthorizationServerTypeSelf,
			SelfAuthServer: AuthorizationServer{
				AccessTokenLifespan:                   config.Duration{Duration: 30 * time.Minute},
				RefreshTokenLifespan:                  config.Duration{Duration: 60 * time.Minute},
				AuthorizationCodeLifespan:             config.Duration{Duration: 5 * time.Minute},
				ClaimSymmetricEncryptionKeySecretName: SecretClaimSymmetricKey,
				TokenSigningRSAKeySecretName:          SecretTokenSigningRSAKey,
				OldTokenSigningRSAKeySecretName:       SecretOldTokenSigningRSAKey,
				StaticClients: map[string]*fosite.DefaultClient{
					"flyte-cli": {
						ID:            "flyte-cli",
						RedirectURIs:  []string{"http://localhost:53593/callback", "http://localhost:12345/callback"},
						ResponseTypes: []string{"code", "token"},
						GrantTypes:    []string{"refresh_token", "authorization_code"},
						Scopes:        []string{"all", "offline", "access_token"},
						Public:        true,
					},
					"flytectl": {
						ID:            "flytectl",
						RedirectURIs:  []string{"http://localhost:53593/callback", "http://localhost:12345/callback"},
						ResponseTypes: []string{"code", "token"},
						GrantTypes:    []string{"refresh_token", "authorization_code"},
						Scopes:        []string{"all", "offline", "access_token"},
						Public:        true,
					},
					"flytepropeller": {
						ID:            "flytepropeller",
						Secret:        []byte(`$2a$10$IxMdI6d.LIRZPpSfEwNoeu4rY3FhDREsxFJXikcgdRRAStxUlsuEO`), // = "foobar"
						RedirectURIs:  []string{"http://localhost:3846/callback"},
						ResponseTypes: []string{"token"},
						GrantTypes:    []string{"refresh_token", "client_credentials"},
						Scopes:        []string{"all", "offline", "access_token"},
					},
				},
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

	HTTPPublicUri config.URL `json:"httpPublicUri" pflag:",The publicly accessible http endpoint. This is used to build absolute URLs for endpoints that are only exposed over http (e.g. /authorize and /token for OAuth2)."`

	UserAuth UserAuthConfig `json:"userAuth" pflag:",Defines Auth options for users."`
	AppAuth  OAuth2Options  `json:"appAuth" pflag:",Defines Auth options for apps. UserAuth must be enabled for AppAuth to work."`
}

type AuthorizationServer struct {
	Issuer                                string          `json:"issuer" pflag:",Defines the issuer to use when issuing and validating tokens. The default value is https://<requestUri.HostAndPort>/"`
	AccessTokenLifespan                   config.Duration `json:"accessTokenLifespan" pflag:",Defines the lifespan of issued access tokens."`
	RefreshTokenLifespan                  config.Duration `json:"refreshTokenLifespan" pflag:",Defines the lifespan of issued access tokens."`
	AuthorizationCodeLifespan             config.Duration `json:"authorizationCodeLifespan" pflag:",Defines the lifespan of issued access tokens."`
	ClaimSymmetricEncryptionKeySecretName string          `json:"claimSymmetricEncryptionKeySecretName" pflag:",OPTIONAL"`
	TokenSigningRSAKeySecretName          string          `json:"tokenSigningRSAKeySecretName" pflag:",OPTIONAL: Secret name to use to retrieve RSA Signing Key."`
	OldTokenSigningRSAKeySecretName       string          `json:"oldTokenSigningRSAKeySecretName" pflag:",OPTIONAL: Secret name to use to retrieve Old RSA Signing Key. This can be useful during key rotation to continue to accept older tokens."`

	StaticClients map[string]*fosite.DefaultClient `json:"staticClients" pflag:"-,Defines statically defined list of clients to allow."`
}

type ExternalAuthorizationServer struct {
	BaseURL config.URL `json:"baseUrl" pflag:",This should be the base url of the authorization server that you are trying to hit. With Okta for instance, it will look something like https://company.okta.com/oauth2/abcdef123456789/"`
}

type OAuth2Options struct {
	AuthServerType     AuthorizationServerType     `json:"authServerType" pflag:"-,Determines authorization server type to use. Additional config should be provided for the chosen AuthorizationServer"`
	SelfAuthServer     AuthorizationServer         `json:"selfAuthServer" pflag:",Authorization Server config to run as a service. Use this when using an IdP that does not offer a custom OAuth2 Authorization Server."`
	ExternalAuthServer ExternalAuthorizationServer `json:"externalAuthServer" pflag:",External Authorization Server config."`
	ThirdParty         ThirdPartyConfigOptions     `json:"thirdPartyConfig" pflag:",Defines settings to instruct flyte cli tools (and optionally others) on what config to use to setup their client."`
}

type UserAuthConfig struct {
	// This is where the user will be redirected to at the end of the flow, but you should not use it. Instead,
	// the initial /login handler should be called with a redirect_url parameter, which will get saved to a cookie.
	// This setting will only be used when that cookie is missing.
	// See the login handler code for more comments.
	RedirectURL config.URL `json:"redirectUrl"`

	OpenID OpenIDOptions `json:"openId" pflag:",OpenID Configuration for User Auth"`
	// Possibly add basicAuth & SAML/p support.

	CookieHashKeySecretName  string `json:"cookie_hash_key_secret_name" pflag:",OPTIONAL: Secret name to use for cookie hash key."`
	CookieBlockKeySecretName string `json:"cookie_block_key_secret_name" pflag:",OPTIONAL: Secret name to use for cookie block key."`
}

type OpenIDOptions struct {
	// The client ID for Admin in your IDP
	// See https://tools.ietf.org/html/rfc6749#section-2.2 for more information
	ClientID string `json:"clientId"`

	// The client secret used in the exchange of the authorization code for the token.
	// https://tools.ietf.org/html/rfc6749#section-2.3
	ClientSecretName string `json:"clientSecretName"`

	// Deprecated: Please use ClientSecretName instead.
	DeprecatedClientSecretFile string `json:"clientSecretFile"`

	// This should be the base url of the authorization server that you are trying to hit. With Okta for instance, it
	// will look something like https://company.okta.com/oauth2/abcdef123456789/
	BaseURL config.URL `json:"baseUrl"`

	// This is the callback URL that will be sent to the IDP authorize endpoint. It is likely that your IDP application
	// needs to have this URL whitelisted before using.
	CallbackURL config.URL `json:"callbackUrl"`

	// Provides a list of scopes to request from the IDP when authenticating. Default value requests claims that should
	// be supported by any OIdC server. Refer to https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims for
	// a complete list. Other providers might support additional scopes that you can define in a config.
	Scopes []string `json:"scopes"`
}

func GetConfig() *Config {
	return cfgSection.GetConfig().(*Config)
}
