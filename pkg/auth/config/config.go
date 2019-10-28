package config

type OAuthOptions struct {
	ClientId         string `json:"clientId"`
	ClientSecretFile string `json:"clientSecretFile"`
	AuthorizeUrl     string `json:"authorizeUrl"`
	TokenUrl         string `json:"tokenUrl"`
	CallbackUrl      string `json:"callbackUrl"`
	Claims           Claims `json:"claims"`

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
