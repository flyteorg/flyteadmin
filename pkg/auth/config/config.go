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

	RedirectUrl string `json:"redirectUrl"`

	// These settings are for non-SSL authentication modes, where
	// TODO: Work out defaults handling better
	HttpAuthorizationHeader string `json:"httpAuthorizationHeader"`
	GrpcAuthorizationHeader string `json:"grpcAuthorizationHeader"`
}

type Claims struct {
	Audience string `json:"aud"`
	Issuer   string `json:"iss"`
}
