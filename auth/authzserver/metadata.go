package authzserver

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/flyteorg/flyteadmin/auth"
	"github.com/flyteorg/flyteadmin/auth/config"
	"net/http"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/auth/interfaces"
)

var (
	tokenRelativeUrl             = config.MustParseURL("/oauth2/token")
	authorizeRelativeUrl         = config.MustParseURL("/oauth2/authorize")
	authorizeCallbackRelativeUrl = config.MustParseURL("/oauth2/authorize_callback")
	jsonWebKeysUrl               = config.MustParseURL("/oauth2/jwks")
	oauth2MetadataEndpoint       = config.MustParseURL("/" + auth.OAuth2MetadataEndpoint)
)

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

func newJSONWebKeySet(publicKeys []rsa.PublicKey) (jwk.Set, error) {
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

		s.Add(key)
	}

	return s, nil
}
