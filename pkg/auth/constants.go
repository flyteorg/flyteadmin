package auth

// OAuth2 Parameters
const CsrfFormKey = "state"
const AuthorizationResponseCodeType = "code"
const OidcScope = "openid"
const ProfileScope = "profile"
const RefreshToken = "refresh_token"
const DefaultAuthorizationHeader = "authorization"
const BearerScheme = "Bearer"

// https://tools.ietf.org/html/rfc8414
// This should be defined without a leading slash. If there is one, the url library's ResolveReference will make it a root path
const MetadataEndpoint = ".well-known/oauth-authorization-server"

// IDP specific
const OfflineAccessType = "offline_access"
