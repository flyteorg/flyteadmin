package interfaces

import (
	"github.com/coreos/go-oidc"
	"github.com/lyft/flyteadmin/pkg/auth/config"
	"golang.org/x/oauth2"
)

//go:generate mockery -name=AuthenticationContext -case=underscore

type AuthenticationContext interface {
	OAuth2Config() *oauth2.Config
	Claims() config.Claims
	OidcProvider() *oidc.Provider
	CookieManager() CookieHandler
	Options() config.OAuthOptions
}
