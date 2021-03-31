package auth

const (
	// OAuth2 Parameters
	CsrfFormKey                   = "state"
	AuthorizationResponseCodeType = "code"
	RefreshToken                  = "refresh_token"
	DefaultAuthorizationHeader    = "authorization"
	BearerScheme                  = "Bearer"
	IDTokenScheme                 = "IDToken"

	// https://tools.ietf.org/html/rfc8414
	// This should be defined without a leading slash. If there is one, the url library's ResolveReference will make it a root path
	OAuth2MetadataEndpoint = ".well-known/oauth-authorization-server"

	// https://openid.net/specs/openid-connect-discovery-1_0.html
	// This should be defined without a leading slash. If there is one, the url library's ResolveReference will make it a root path
	OIdCMetadataEndpoint = ".well-known/openid-configuration"

	ContextKeyIdentityContext = "identity_context"
)
