package auth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/coreos/go-oidc"
)

func NewMockOIdCProvider() (*oidc.Provider, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var issuer string
	hf := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, strings.ReplaceAll(`{
				"issuer": "https://dev-14186422.okta.com",
				"authorization_endpoint": "https://example.com/auth",
				"token_endpoint": "https://example.com/token",
				"jwks_uri": "https://example.com/keys",
				"id_token_signing_alg_values_supported": ["RS256"]
			}`, "ISSUER", issuer))
	}

	s := httptest.NewServer(http.HandlerFunc(hf))
	defer s.Close()

	issuer = s.URL
	return oidc.NewProvider(ctx, issuer)
}
