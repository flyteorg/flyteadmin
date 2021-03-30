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

	SecretCookieHashKey  = "cookie_hash_key"
	SecretCookieBlockKey = "cookie_block_key"

	// Base64 encoded secret of exactly 32 bytes
	SecretTokenHash = "token_hash_key"

	// PrivateKey is used to sign JWT tokens. The default strategy uses RS256 (RSA Signature with SHA-256)
	SecretTokenSigningRSAKey = "token_rsa_key.pem"
	// PrivateKey that was used to sign old JWT tokens. The default strategy uses RS256 (RSA Signature with SHA-256)
	// This is used to support key rotation. When present, it'll only be used to validate incoming tokens. New tokens
	// will not be issued using this key.
	SecretOldTokenSigningRSAKey = "token_rsa_key_old.pem"

	ContextKeyIdentityContext = "identity_context"
)
