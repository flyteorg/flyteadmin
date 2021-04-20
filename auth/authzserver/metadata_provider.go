package authzserver

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/flyteorg/flyteadmin/auth"

	authConfig "github.com/flyteorg/flyteadmin/auth/config"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
)

type OAuth2MetadataProvider struct {
	cfg *authConfig.Config
}

func (s OAuth2MetadataProvider) OAuth2Metadata(ctx context.Context, r *service.OAuth2MetadataRequest) (*service.OAuth2MetadataResponse, error) {
	switch s.cfg.AppAuth.AuthServerType {
	case authConfig.AuthorizationServerTypeSelf:
		u := auth.GetPublicURL(ctx, s.cfg.Secure, &s.cfg.HTTPPublicURI.URL)
		doc := &service.OAuth2MetadataResponse{
			Issuer:                        GetIssuer(ctx, nil, s.cfg),
			AuthorizationEndpoint:         u.ResolveReference(authorizeRelativeURL).String(),
			TokenEndpoint:                 u.ResolveReference(tokenRelativeURL).String(),
			JwksUri:                       u.ResolveReference(jsonWebKeysURL).String(),
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

		return doc, nil
	default:
		var externalMetadataURL *url.URL
		if len(s.cfg.AppAuth.ExternalAuthServer.BaseURL.String()) > 0 {
			externalMetadataURL = s.cfg.AppAuth.ExternalAuthServer.BaseURL.ResolveReference(oauth2MetadataEndpoint)
		} else {
			externalMetadataURL = s.cfg.UserAuth.OpenID.BaseURL.ResolveReference(oauth2MetadataEndpoint)
		}

		response, err := http.Get(externalMetadataURL.String())
		if err != nil {
			return nil, err
		}

		raw, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		resp := &service.OAuth2MetadataResponse{}
		err = unmarshalResp(response, raw, resp)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}

func (s OAuth2MetadataProvider) FlyteClient(context.Context, *service.FlyteClientRequest) (*service.FlyteClientResponse, error) {
	return &service.FlyteClientResponse{
		ClientId:                 s.cfg.AppAuth.ThirdParty.FlyteClientConfig.ClientID,
		RedirectUri:              s.cfg.AppAuth.ThirdParty.FlyteClientConfig.RedirectURI,
		Scopes:                   s.cfg.AppAuth.ThirdParty.FlyteClientConfig.Scopes,
		AuthorizationMetadataKey: s.cfg.GrpcAuthorizationHeader,
	}, nil
}

func NewService(config *authConfig.Config) OAuth2MetadataProvider {
	return OAuth2MetadataProvider{
		cfg: config,
	}
}
