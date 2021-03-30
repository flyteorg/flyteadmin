package oauthserver

import (
	"context"
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
	ScopesSupported                   []string `json:"scoped_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	JSONWebKeysUri                    string   `json:"jwks_uri"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`

	// optional
	GrantTypesSupported []string `json:"grant_types_supported"`
}

func GetJSONWebKeysEndpoint(authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		s := jwk.NewSet()
		for _, publicKey := range authCtx.OAuth2Provider().PublicKeys() {
			key, err := jwk.New(publicKey)
			if err != nil {
				http.Error(writer, fmt.Errorf("failed to write public key. Error: %w", err).Error(), http.StatusInternalServerError)
				return
			}

			err = jwk.AssignKeyID(key)
			if err != nil {
				http.Error(writer, fmt.Errorf("failed to write public key. Error: %w", err).Error(), http.StatusInternalServerError)
				return
			}

			s.Add(key)
		}

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

func GetMetadataEndpoint(authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		requestUrl, err := getRequestBaseUrl(request)
		if err != nil {
			http.Error(rw, "Error parsing request url", http.StatusBadRequest)
			return
		}

		doc := DiscoveryDocument{
			Issuer:                        getIssuer(authCtx.Options(), request),
			AuthorizationEndpoint:         requestUrl.ResolveReference(authorizeRelativeUrl).String(),
			TokenEndpoint:                 requestUrl.ResolveReference(tokenRelativeUrl).String(),
			JSONWebKeysUri:                requestUrl.ResolveReference(jsonWebKeysUrl).String(),
			CodeChallengeMethodsSupported: []string{"S256"},
			ResponseTypesSupported: []string{
				"code",
				"token",
				"code token",
			},
			GrantTypesSupported: []string{
				"authorization_code",
				"refresh_token",
				"client_credentials",
			},
			ScopesSupported: []string{"all"},
			TokenEndpointAuthMethodsSupported: []string{
				"client_secret_basic",
			},
		}

		raw, err := json.Marshal(doc)
		if err != nil {
			http.Error(rw, "Error marshaling Metadata Doc", http.StatusBadRequest)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		size, err := rw.Write(raw)
		if err != nil {
			logger.Errorf(context.TODO(), "Wrote user info response size %d, err %s", size, err)
		}
	}
}
