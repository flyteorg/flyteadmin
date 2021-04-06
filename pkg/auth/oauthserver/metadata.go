package oauthserver

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/flyteorg/flyteadmin/pkg/auth"

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

		err = key.Set(KeyMetadataPublicCert, &publicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to write public key. Error: %w", err)
		}

		s.Add(key)
	}

	return s, nil
}

// GetMetadataEndpoint serves auth.OAuth2MetadataEndpoint requests with an RFC Compliant json object.
// ref: https://tools.ietf.org/html/rfc8414
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
			GrantTypesSupported: supportedGrantTypes,
			ScopesSupported:     []string{auth.ScopeAll},
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

// GetMetadataRedirect redirects auth.OAuth2MetadataEndpoint requests to the external Authorization Server configured
func GetMetadataRedirect(authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		baseURL := authCtx.Options().AppAuth.ExternalAuthServer.BaseURL
		http.Redirect(rw, request, baseURL.ResolveReference(oauth2MetadataEndpoint).String(), http.StatusTemporaryRedirect)
	}
}
