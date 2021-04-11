package interfaces

import (
	"context"
	"crypto/rsa"
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

type OAuth2Provider interface {
	fosite.OAuth2Provider
	NewJWTSessionToken(subject string, userInfoClaims interface{}, appID, issuer, audience string) *fositeOAuth2.JWTSession
	ValidateAccessToken(ctx context.Context, tokenStr string) (IdentityContext, error)
	PublicKeys() []rsa.PublicKey
	KeySet() jwk.Set
}

type OAuth2MetadataProvider interface {
	OAuth2Metadata(context.Context, *service.OAuth2MetadataRequest) (*service.OAuth2MetadataResponse, error)
	FlyteClient(context.Context, *service.FlyteClientRequest) (*service.FlyteClientResponse, error)
	AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error)
}

type OIdCUserInfoProvider interface {
	UserInfo(context.Context, *service.UserInfoRequest) (*service.UserInfoResponse, error)
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
	AuthService() service.AuthServiceServer
}

type IdentityContext interface {
	UserID() string
	AppID() string
	UserInfo() *service.UserInfoResponse
	IsEmpty() bool
	WithContext(ctx context.Context) context.Context
	AuthenticatedAt() time.Time
	Scopes() sets.String
}
