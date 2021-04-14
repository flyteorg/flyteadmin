package oauthserver

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/pkg/auth/interfaces"
)

var (
	tokenRelativeUrl             = mustParseURL("/oauth2/token")
	authorizeRelativeUrl         = mustParseURL("/oauth2/authorize")
	authorizeCallbackRelativeUrl = mustParseURL("/oauth2/authorize_callback")
	jsonWebKeysUrl               = mustParseURL("/oauth2/jwks")
)

func mustParseURL(rawURL string) *url.URL {
	res, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	return res
}

type DiscoveryDocument struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	JSONWebKeysUri                    string   `json:"jwks_uri"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`

	// optional
	GrantTypesSupported []string `json:"grant_types_supported"`
}

// GetJSONWebKeysEndpoint serves requests to the jwks endpoint.
// ref: https://tools.ietf.org/html/rfc7517
func GetJSONWebKeysEndpoint(authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		s := authCtx.OAuth2Provider().KeySet()
		raw, err := json.Marshal(s)
		if err != nil {
			http.Error(writer, fmt.Errorf("failed to write public key. Error: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		size, err := writer.Write(raw)
		if err != nil {
			logger.Errorf(context.Background(), "Wrote JSONWebKeys response size %d, err %s", size, err)
		}
	}
}

func getJSONWebKeys(publicKeys []rsa.PublicKey) (jwk.Set, error) {
	s := jwk.NewSet()
	for _, publicKey := range publicKeys {
		key, err := jwk.New(publicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to write public key. Error: %w", err)
		}

		err = jwk.AssignKeyID(key)
		if err != nil {
			return nil, fmt.Errorf("failed to write public key. Error: %w", err)
		}
		var localPublicKey rsa.PublicKey
		localPublicKey = publicKey
		err = key.Set(KeyMetadataPublicCert, &localPublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to write public key. Error: %w", err)
		}

		s.Add(key)
	}

	return s, nil
}
