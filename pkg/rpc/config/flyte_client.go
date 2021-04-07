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

func HandleFlyteCliConfigFunc(ctx context.Context, cfg *config.ServerConfig, authCfg *authConfig.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var configValues map[string]interface{}
		if !cfg.DeprecatedThirdPartyConfig.IsEmpty() {
			configValues = map[string]interface{}{
				clientID:        cfg.DeprecatedThirdPartyConfig.FlyteClientConfig.ClientID,
				redirectURI:     cfg.DeprecatedThirdPartyConfig.FlyteClientConfig.RedirectURI,
				scopes:          authCfg.UserAuth.OpenID.Scopes,
				authMetadataKey: authCfg.GrpcAuthorizationHeader,
			}
		} else {
			configValues = map[string]interface{}{
				clientID:        authCfg.AppAuth.ThirdParty.FlyteClientConfig.ClientID,
				redirectURI:     authCfg.AppAuth.ThirdParty.FlyteClientConfig.RedirectURI,
				scopes:          authCfg.AppAuth.ThirdParty.FlyteClientConfig.Scopes,
				authMetadataKey: authCfg.GrpcAuthorizationHeader,
			}
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
