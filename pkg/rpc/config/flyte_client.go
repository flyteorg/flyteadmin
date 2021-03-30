package config

import (
	"context"
	"encoding/json"
	"net/http"

	authConfig "github.com/flyteorg/flyteadmin/pkg/auth/config"

	"github.com/flyteorg/flyteadmin/pkg/config"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	clientID        = "client_id"
	redirectURI     = "redirect_uri"
	scopes          = "scopes"
	authMetadataKey = "authorization_metadata_key"
)

func HandleFlyteCliConfigFunc(ctx context.Context, cfg *config.ServerConfig, authConfig *authConfig.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		configValues := map[string]interface{}{
			clientID:        cfg.ThirdPartyConfig.FlyteClientConfig.ClientID,
			redirectURI:     cfg.ThirdPartyConfig.FlyteClientConfig.RedirectURI,
			scopes:          authConfig.UserAuth.OpenID.Scopes,
			authMetadataKey: authConfig.GrpcAuthorizationHeader,
		}

		configJSON, err := json.Marshal(configValues)
		if err != nil {
			logger.Infof(ctx, "Failed to marshal flyte_client config to JSON with err: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write(configJSON)
		if err != nil {
			logger.Warningf(ctx, "Failed to write config json [%+v] with err: %v", string(configJSON), err)
		}
	}
}
