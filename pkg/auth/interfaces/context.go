package interfaces

import (
	"context"
	"crypto/rsa"
	"net/http"
	"net/url"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/ory/fosite"
	fositeOAuth2 "github.com/ory/fosite/handler/oauth2"

	"github.com/coreos/go-oidc"
	"github.com/flyteorg/flyteadmin/pkg/auth/config"
	"golang.org/x/oauth2"
)

//go:generate mockery -name=AuthenticationContext -case=underscore

type HandlerRegisterer interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

type OAuth2Provider interface {
	fosite.OAuth2Provider
	NewJWTSessionToken(subject string, userInfoClaims interface{}, appID, issuer, audience string) *fositeOAuth2.JWTSession
	ValidateAccessToken(ctx context.Context, tokenStr string) (IdentityContext, error)
	PublicKeys() []rsa.PublicKey
	KeySet() jwk.Set
}

// This interface is a convenience wrapper object that holds all the utilities necessary to run Flyte Admin behind authentication
// It is constructed at the root server layer, and passed around to the various auth handlers and utility functions/objects.
type AuthenticationContext interface {
	OAuth2Provider() OAuth2Provider
	OAuth2ClientConfig() *oauth2.Config
	OidcProvider() *oidc.Provider
	CookieManager() CookieHandler
	Options() *config.Config
	GetOAuth2MetadataURL() *url.URL
	GetOIdCMetadataURL() *url.URL
	GetHTTPClient() *http.Client
}

type IdentityContext interface {
	UserID() string
	AppID() string
	UserInfo() UserInfo
	IsEmpty() bool
	WithContext(ctx context.Context) context.Context
	AuthenticatedAt() time.Time
	Scopes() sets.String
}

type UserInfo interface {
	Subject() string
	Name() string
	PreferredUsername() string
	GivenName() string
	FamilyName() string
	Email() string
	Picture() string

	MarshalToJSON() ([]byte, error)
}
