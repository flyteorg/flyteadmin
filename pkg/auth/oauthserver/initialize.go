package oauthserver

import (
	"crypto/rsa"
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/auth"
	"github.com/flyteorg/flyteadmin/pkg/auth/config"

	"github.com/ory/fosite"

	"github.com/flyteorg/flyteadmin/pkg/auth/interfaces"

	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/token/jwt"
)

var (
	oauth2MetadataEndpoint = mustParseURL("/" + auth.OAuth2MetadataEndpoint)
)

func RegisterHandlers(handler interfaces.HandlerRegisterer, authCtx interfaces.AuthenticationContext) {
	if authCtx.OAuth2Provider() != nil {
		// Set up oauthserver endpoints. You could also use gorilla/mux or any other router.
		handler.HandleFunc(authorizeRelativeUrl.String(), getAuthEndpoint(authCtx))
		handler.HandleFunc(authorizeCallbackRelativeUrl.String(), getAuthCallbackEndpoint(authCtx))
		handler.HandleFunc(tokenRelativeUrl.String(), getTokenEndpointHandler(authCtx))

		// The metadata endpoint is an RFC-defined constant, but we need a leading / for the handler to pattern match correctly.
		handler.HandleFunc(fmt.Sprintf("/%s", auth.OAuth2MetadataEndpoint), GetMetadataEndpoint(authCtx))
		handler.HandleFunc(jsonWebKeysUrl.String(), GetJSONWebKeysEndpoint(authCtx))
	} else {
		// The metadata endpoint is an RFC-defined constant, but we need a leading / for the handler to pattern match correctly.
		handler.HandleFunc(fmt.Sprintf("/%s", auth.OAuth2MetadataEndpoint), GetMetadataRedirect(authCtx))
	}

	// TODO: Support token revocation and introspection
	// revoke tokens
	//handler.HandleFunc("/oauthserver/revoke", revokeEndpoint)
	// introspect tokens
	//handler.HandleFunc("/oauthserver/introspect", introspectionEndpoint)
}

func composeOAuth2Provider(authCfg config.AuthorizationServer, config *compose.Config, storage interface{}, secret []byte, key *rsa.PrivateKey) fosite.OAuth2Provider {
	commonStrategy := &compose.CommonStrategy{
		CoreStrategy:               NewStatelessCodeProvider(authCfg, compose.NewOAuth2JWTStrategy(key, compose.NewOAuth2HMACStrategy(config, secret, nil))),
		OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(config, key),
		JWTStrategy: &jwt.RS256JWTStrategy{
			PrivateKey: key,
		},
	}

	return compose.Compose(
		config,
		storage,
		commonStrategy,
		nil,

		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2ClientCredentialsGrantFactory,
		compose.OAuth2RefreshTokenGrantFactory,

		compose.OAuth2StatelessJWTIntrospectionFactory,
		//compose.OAuth2TokenRevocationFactory,

		compose.OAuth2PKCEFactory,
	)
}
