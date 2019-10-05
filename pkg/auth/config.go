package auth

import (
	"golang.org/x/oauth2"
	"io/ioutil"
	"strings"
)


type OAuthOptions struct {
	ClientId         string `json:"clientId"`
	ClientSecretFile string `json:"clientSecretFile"`
	JwksUrl          string `json:"jwksUrl"`
	AuthorizeUrl     string `json:"authorizeUrl"`
	TokenUrl         string `json:"tokenUrl"`
	CallbackUrl      string `json:"callbackUrl"`
	Claims           Claims `json:"claims"`

	// These should point to files that contain base64 encoded secrets. See the TestSecureCookieLifecycle() unit test.
	// See https://github.com/gorilla/securecookie#examples for more information
	CookieHashKeyFile  string `json:"cookieHashKeyFile"`
	CookieBlockKeyFile string `json:"cookieBlockKeyFile"`

	RedirectUrl string `json:"redirectUrl"`
}

type Claims struct {
	Audience string `json:"aud"`
	Issuer   string `json:"iss"`
}

// This creates a oauth2 library config object, with values from the Flyte Admin config
func GetOauth2Config(options OAuthOptions) (oauth2.Config, error) {
	secretBytes, err := ioutil.ReadFile(options.ClientSecretFile)
	if err != nil {
		return oauth2.Config{}, err
	}
	secret := strings.TrimSuffix(string(secretBytes), "\n")
	return oauth2.Config{
		RedirectURL:  options.CallbackUrl,
		ClientID:     options.ClientId,
		ClientSecret: secret,
		// Offline access needs to be specified in order to return a refresh token in the exchange.
		// TODO: Second parameter is IDP specific - move to config. Also handle case where a refresh token is not allowed
		Scopes: []string{OidcScope, OfflineAccessType},
		Endpoint: oauth2.Endpoint{
			AuthURL:  options.AuthorizeUrl,
			TokenURL: options.TokenUrl,
		},
	}, nil
}
