package config

type OAuthOptions struct {
	ClientId         string `json:"clientId"`
	ClientSecretFile string `json:"clientSecretFile"`

	// This should be the base url of the authorization server that you are trying to hit. With Okta for instance, it
	// will look something like https://company.okta.com/oauth2/abcdef123456789/
	BaseUrl          string `json:"baseUrl"`

	// These two config elements currently need the entire path, including the already specified baseUrl
	// TODO: Refactor to allow use of discovery (see https://tools.ietf.org/id/draft-ietf-oauth-discovery-08.html)
	//       Also refactor to use relative paths when discovery is not available
	AuthorizeUrl     string `json:"authorizeUrl"`
	TokenUrl         string `json:"tokenUrl"`

	// This is the callback URL that will be sent to the IDP authorize endpoint. It is likely that your IDP application
	// needs to have this URL whitelisted before using.
	CallbackUrl      string `json:"callbackUrl"`
	Claims           Claims `json:"claims"`

	// This is the relative path of the user info endpoint, if there is one, for the given IDP. This will be appended to
	// the base URL of the IDP. This is used to support the /me endpoint that Admin will serve when running with authentication
	// See https://developer.okta.com/docs/reference/api/oidc/#userinfo as an example.
	IdpUserInfoEndpoint string `json:"idpUserInfoEndpoint"`

	// These should point to files that contain base64 encoded secrets. See the TestSecureCookieLifecycle() unit test.
	// See https://github.com/gorilla/securecookie#examples for more information
	CookieHashKeyFile  string `json:"cookieHashKeyFile"`
	CookieBlockKeyFile string `json:"cookieBlockKeyFile"`

	// This is where the user will be redirected to at the end of the flow, but you should not use it. Instead,
	// the initial /login handler should be called with a redirect_url parameter, which will get saved to a cookie.
	// This setting will only be used when that cookie is missing.
	// See the login handler code for more comments.
	RedirectUrl string `json:"redirectUrl"`

	// These settings are for non-SSL authentication modes, where Envoy is handling SSL termination
	// This is not yet used, but this is the HTTP variant of the setting below.
	HttpAuthorizationHeader string `json:"httpAuthorizationHeader"`

	// In order to support deployments of this Admin service where Envoy is terminating SSL connections, the metadata
	// header name cannot be "authorization", which is the standard metadata name. Envoy has special handling for that
	// name. Instead, there is a gRPC interceptor, GetAuthenticationCustomMetadataInterceptor, that will translate
	// incoming metadata headers with this config setting's name, into that standard header
	GrpcAuthorizationHeader string `json:"grpcAuthorizationHeader"`
}

type Claims struct {
	Audience string `json:"aud"`
	Issuer   string `json:"iss"`
}
