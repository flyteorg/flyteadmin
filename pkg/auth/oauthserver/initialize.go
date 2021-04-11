package oauthserver

import (
	"crypto/rsa"
	"github.com/ory/fosite/handler/oauth2"

	"github.com/flyteorg/flyteadmin/pkg/auth"
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
		handler.HandleFunc(jsonWebKeysUrl.String(), GetJSONWebKeysEndpoint(authCtx))
	}

	// TODO: Support token revocation and introspection
	// revoke tokens
	//handler.HandleFunc("/oauthserver/revoke", revokeEndpoint)
	// introspect tokens
	//handler.HandleFunc("/oauthserver/introspect", introspectionEndpoint)
}

func composeOAuth2Provider(codeProvider oauth2.CoreStrategy, config *compose.Config, storage interface{}, key *rsa.PrivateKey) fosite.OAuth2Provider {
	commonStrategy := &compose.CommonStrategy{
		CoreStrategy:               codeProvider,
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
