package auth

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/coreos/go-oidc"
	"github.com/flyteorg/flyteadmin/pkg/auth/config"
	"github.com/flyteorg/flyteadmin/pkg/auth/interfaces"
	"github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"golang.org/x/oauth2"
)

const (
	IdpConnectionTimeout = 10 * time.Second

	ErrauthCtx    errors.ErrorCode = "AUTH_CONTEXT_SETUP_FAILED"
	ErrConfigFileRead errors.ErrorCode = "CONFIG_OPTION_FILE_READ_FAILED"
)

// Please see the comment on the corresponding AuthenticationContext for more information.
type Context struct {
	oauth2Client   *oauth2.Config
	cookieManager  interfaces.CookieHandler
	oidcProvider   *oidc.Provider
	options        *config.Config
	oauth2Provider interfaces.OAuth2Provider

	userInfoURL       *url.URL
	oauth2MetadataURL *url.URL
	oidcMetadataURL   *url.URL
	httpClient        *http.Client
}

func (c Context) OAuth2Provider() interfaces.OAuth2Provider {
	return c.oauth2Provider
}

func (c Context) OAuth2ClientConfig() *oauth2.Config {
	return c.oauth2Client
}

func (c Context) OidcProvider() *oidc.Provider {
	return c.oidcProvider
}

func (c Context) CookieManager() interfaces.CookieHandler {
	return c.cookieManager
}

func (c Context) Options() *config.Config {
	return c.options
}

func (c Context) GetUserInfoURL() *url.URL {
	return c.userInfoURL
}

func (c Context) GetHTTPClient() *http.Client {
	return c.httpClient
}

func (c Context) GetOAuth2MetadataURL() *url.URL {
	return c.oauth2MetadataURL
}

func (c Context) GetOIdCMetadataURL() *url.URL {
	return c.oidcMetadataURL
}

func NewAuthenticationContext(ctx context.Context, sm core.SecretManager, oauth2Provider interfaces.OAuth2Provider,
	options *config.Config) (Context, error) {

	// Construct the cookie manager object.
	hashKeyBase64, err := sm.Get(ctx, SecretCookieHashKey)
	if err != nil {
		return Context{}, errors.Wrapf(ErrConfigFileRead, err, "Could not read hash key file")
	}

	blockKeyBase64, err := sm.Get(ctx, SecretCookieBlockKey)
	if err != nil {
		return Context{}, errors.Wrapf(ErrConfigFileRead, err, "Could not read hash key file")
	}

	cookieManager, err := NewCookieManager(ctx, hashKeyBase64, blockKeyBase64)
	if err != nil {
		logger.Errorf(ctx, "Error creating cookie manager %s", err)
		return Context{}, errors.Wrapf(ErrauthCtx, err, "Error creating cookie manager")
	}

	// Construct an http client for interacting with the IDP if necessary.
	httpClient := &http.Client{
		Timeout: IdpConnectionTimeout,
	}

	// Construct an oidc Provider, which needs its own http Client.
	oidcCtx := oidc.ClientContext(ctx, httpClient)
	baseURL := options.UserAuth.OpenID.BaseURL.String()
	provider, err := oidc.NewProvider(oidcCtx, baseURL)
	if err != nil {
		return Context{}, errors.Wrapf(ErrauthCtx, err, "Error creating oidc provider w/ issuer [%v]", baseURL)
	}

	// Construct the golang OAuth2 library's own internal configuration object from this package's config
	oauth2Config, err := GetOAuth2ClientConfig(options.UserAuth.OpenID, provider.Endpoint())
	if err != nil {
		return Context{}, errors.Wrapf(ErrauthCtx, err, "Error creating OAuth2 library configuration")
	}

	logger.Infof(ctx, "Base IDP URL is %s", options.UserAuth.OpenID.BaseURL)

	oauth2MetadataURL, err := url.Parse(OAuth2MetadataEndpoint)
	if err != nil {
		logger.Errorf(ctx, "Error parsing oauth2 metadata URL %s", err)
		return Context{}, errors.Wrapf(ErrauthCtx, err, "Error parsing metadata URL")
	}

	logger.Infof(ctx, "Metadata endpoint is %s", oauth2MetadataURL)

	oidcMetadataURL, err := url.Parse(OIdCMetadataEndpoint)
	if err != nil {
		logger.Errorf(ctx, "Error parsing oidc metadata URL %s", err)
		return Context{}, errors.Wrapf(ErrauthCtx, err, "Error parsing metadata URL")
	}

	logger.Infof(ctx, "Metadata endpoint is %s", oidcMetadataURL)

	return Context{
		options:           options,
		oidcMetadataURL:   oidcMetadataURL,
		oauth2MetadataURL: oauth2MetadataURL,
		oauth2Client:      &oauth2Config,
		oidcProvider:      provider,
		httpClient:        httpClient,
		cookieManager:     cookieManager,
		oauth2Provider:    oauth2Provider,
	}, nil
}

// This creates a oauth2 library config object, with values from the Flyte Admin config
func GetOAuth2ClientConfig(options config.OpenIDOptions, endpoint oauth2.Endpoint) (oauth2.Config, error) {
	secretBytes, err := ioutil.ReadFile(options.ClientSecretFile)
	if err != nil {
		return oauth2.Config{}, err
	}

	secret := strings.TrimSuffix(string(secretBytes), "\n")
	return oauth2.Config{
		RedirectURL:  options.CallbackURL.String(),
		ClientID:     options.ClientID,
		ClientSecret: secret,
		Scopes:       options.Scopes,
		Endpoint:     endpoint,
	}, nil
}
