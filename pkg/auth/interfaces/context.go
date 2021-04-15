package interfaces

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/ory/fosite"
	fositeOAuth2 "github.com/ory/fosite/handler/oauth2"

	"github.com/coreos/go-oidc"
	"github.com/flyteorg/flyteadmin/pkg/auth/config"
	"golang.org/x/oauth2"
)

//go:generate mockery -all -case=underscore

type HandlerRegisterer interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

// OAuth2Provider represents an OAuth2 Provider that can be used to issue OAuth2 tokens.
type OAuth2Provider interface {
	fosite.OAuth2Provider
	OAuth2ResourceServer
	NewJWTSessionToken(subject, appID, issuer, audience string, userInfoClaims *service.UserInfoResponse) *fositeOAuth2.JWTSession
	KeySet() jwk.Set
}

// OAuth2ResourceServer represents a resource server that can be accessed through an access token.
type OAuth2ResourceServer interface {
	ValidateAccessToken(ctx context.Context, tokenStr string) (IdentityContext, error)
}

// OAuth2MetadataProvider represents a subset of the service.AuthServiceServer interface
type OAuth2MetadataProvider interface {
	OAuth2Metadata(context.Context, *service.OAuth2MetadataRequest) (*service.OAuth2MetadataResponse, error)
	FlyteClient(context.Context, *service.FlyteClientRequest) (*service.FlyteClientResponse, error)
}

// OIdCUserInfoProvider represents a subset of the service.AuthServiceServer interface
type OIdCUserInfoProvider interface {
	UserInfo(context.Context, *service.UserInfoRequest) (*service.UserInfoResponse, error)
}

// AuthenticationContext is a convenience wrapper object that holds all the utilities necessary to run Flyte Admin behind authentication
// It is constructed at the root server layer, and passed around to the various auth handlers and utility functions/objects.
type AuthenticationContext interface {
	OAuth2Provider() OAuth2Provider
	OAuth2ResourceServer() OAuth2ResourceServer
	OAuth2ClientConfig() *oauth2.Config
	OidcProvider() *oidc.Provider
	CookieManager() CookieHandler
	Options() *config.Config
	GetOAuth2MetadataURL() *url.URL
	GetOIdCMetadataURL() *url.URL
	GetHTTPClient() *http.Client
	AuthService() service.AuthServiceServer
}

// IdentityContext represents the authenticated identity and can be used to abstract the way the user/app authenticated
// to the platform.
type IdentityContext interface {
	UserID() string
	AppID() string
	UserInfo() *service.UserInfoResponse
	AuthenticatedAt() time.Time
	Scopes() sets.String

	IsEmpty() bool
	WithContext(ctx context.Context) context.Context
}
