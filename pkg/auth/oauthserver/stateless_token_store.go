package oauthserver

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/auth/config"

	"github.com/ory/fosite/handler/oauth2"
	oauth22 "github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/token/jwt"

	"github.com/ory/fosite"
	"github.com/ory/fosite/storage"
)

// StatelessTokenStore provides a ship on top of the MemoryStore to avoid storing tokens in memory (or elsewhere) but
// instead hydrates fosite.Request and sessions from the tokens themselves.
type StatelessTokenStore struct {
	*storage.MemoryStore
	jwt.JWTStrategy
}

func (s StatelessTokenStore) rehydrateSession(ctx context.Context, token string) (request *fosite.Request, err error) {
	t, err := s.JWTStrategy.Decode(ctx, token)
	if err != nil {
		return nil, err
	}

	ifaceRequest := oauth2.AccessTokenJWTToRequest(t)
	rawRequest, casted := ifaceRequest.(*fosite.Request)
	if !casted {
		return nil, fmt.Errorf("expected *fosite.Request. Found %v", reflect.TypeOf(ifaceRequest))
	}

	client, err := s.GetClient(ctx, rawRequest.GetClient().GetID())
	if err != nil {
		return nil, err
	}

	rawRequest.Client = client
	return rawRequest, nil
}

func (s StatelessTokenStore) InvalidateAuthorizeCodeSession(_ context.Context, _ string) (err error) {
	return nil
}

func (s StatelessTokenStore) GetAuthorizeCodeSession(ctx context.Context, code string, _ fosite.Session) (fosite.Requester, error) {
	request, err := s.rehydrateSession(ctx, code)
	if err != nil {
		return nil, err
	}

	if !request.RequestedScope.Has(accessTokenScope) {
		return nil, fmt.Errorf("authcode not found [%v]", code)
	}

	requestedScopes := request.RequestedScope
	request.RequestedScope = fosite.Arguments{}
	for _, requestedScope := range requestedScopes {
		if requestedScope != accessTokenScope {
			request.AppendRequestedScope(strings.TrimPrefix(requestedScope, requestedScopePrefix))
		}
	}

	return request, nil
}

func (s StatelessTokenStore) GetPKCERequestSession(ctx context.Context, signature string, _ fosite.Session) (fosite.Requester, error) {
	request, err := s.rehydrateSession(ctx, signature)
	if err != nil {
		return nil, err
	}

	if !request.RequestedScope.Has(accessTokenScope) {
		return nil, fmt.Errorf("PKCE request not found [%v]", signature)
	}

	requestedScopes := request.RequestedScope
	request.RequestedScope = fosite.Arguments{}
	for _, requestedScope := range requestedScopes {
		if requestedScope != accessTokenScope {
			request.AppendRequestedScope(strings.TrimPrefix(requestedScope, requestedScopePrefix))
		}
	}

	return request, nil
}

func (s StatelessTokenStore) GetRefreshTokenSession(ctx context.Context, signature string, _ fosite.Session) (request fosite.Requester, err error) {
	rawRequest, err := s.rehydrateSession(ctx, signature)
	if err != nil {
		return nil, err
	}

	requestedScopes := rawRequest.GrantedScope
	rawRequest.GrantedScope = fosite.Arguments{}
	rawRequest.RequestedScope = fosite.Arguments{}
	for _, scope := range requestedScopes {
		rawRequest.AppendRequestedScope(strings.TrimPrefix(scope, requestedScopePrefix))
		rawRequest.GrantScope(strings.TrimPrefix(scope, requestedScopePrefix))
	}

	return rawRequest, nil
}

func (s StatelessTokenStore) DeleteRefreshTokenSession(_ context.Context, _ string) (err error) {
	return nil
}

// StatelessCodeProvider offers a strategy that encodes authorization code and refresh tokens into JWT
// to avoid requiring storing these tokens on the server side. These tokens are usually short lived so storing them to a
// persistent store (e.g. DB) is not desired. A more suitable store would be an in-memory read-efficient store (e.g.
// Redis) however, that would add additional requirements on setting up flyteAdmin and hence why we are going with this
// strategy.
type StatelessCodeProvider struct {
	oauth22.CoreStrategy
	authorizationCodeLifespan time.Duration
	refreshTokenLifespan      time.Duration
}

func (p StatelessCodeProvider) AuthorizeCodeSignature(token string) string {
	return token
}

func (p StatelessCodeProvider) GenerateAuthorizeCode(ctx context.Context, requester fosite.Requester) (token string, signature string, err error) {
	rawRequest, casted := requester.(*fosite.AuthorizeRequest)
	if !casted {
		return "", "", fmt.Errorf("expected *fosite.AuthorizeRequest. Found [%v]", reflect.TypeOf(requester))
	}

	for _, requestedScope := range requester.GetRequestedScopes() {
		if requestedScope == refreshTokenScope {
			requester.GrantScope(refreshTokenScope)
		} else {
			requester.GrantScope(requestedScopePrefix + requestedScope)
		}
	}

	requester.GrantScope(accessTokenScope)
	rawRequest.GetSession().SetExpiresAt(fosite.AccessToken, time.Now().Add(p.authorizationCodeLifespan))
	token, _, err = p.CoreStrategy.GenerateAccessToken(ctx, requester)
	return token, token, err
}

func (p StatelessCodeProvider) ValidateAuthorizeCode(ctx context.Context, requester fosite.Requester, token string) (err error) {
	return p.CoreStrategy.ValidateAccessToken(ctx, requester, token)
}

func (p StatelessCodeProvider) RefreshTokenSignature(token string) string {
	return token
}

func (p StatelessCodeProvider) GenerateRefreshToken(ctx context.Context, requester fosite.Requester) (token string, signature string, err error) {
	rawRequest, casted := requester.(*fosite.AccessRequest)
	if !casted {
		return "", "", fmt.Errorf("expected *fosite.AccessRequest. Found [%v]", reflect.TypeOf(requester))
	}

	grantedScopes := requester.GetGrantedScopes()
	rawRequest.GrantedScope = fosite.Arguments{}

	for _, requestedScope := range grantedScopes {
		if requestedScope == refreshTokenScope || requestedScope == accessTokenScope {
			requester.GrantScope(requestedScope)
		} else if strings.HasPrefix(requestedScope, requestedScopePrefix) {
			requester.GrantScope(requestedScope)
		} else {
			requester.GrantScope(requestedScopePrefix + requestedScope)
		}
	}

	rawRequest.GetSession().SetExpiresAt(fosite.AccessToken, time.Now().Add(p.refreshTokenLifespan))

	token, _, err = p.CoreStrategy.GenerateAccessToken(ctx, requester)
	return token, token, err
}

func (p StatelessCodeProvider) ValidateRefreshToken(ctx context.Context, requester fosite.Requester, token string) (err error) {
	return p.CoreStrategy.ValidateAccessToken(ctx, requester, token)
}

func NewStatelessCodeProvider(cfg config.AuthorizationServer, strategy oauth22.CoreStrategy) StatelessCodeProvider {
	return StatelessCodeProvider{
		CoreStrategy:              strategy,
		authorizationCodeLifespan: cfg.AuthorizationCodeLifespan.Duration,
		refreshTokenLifespan:      cfg.RefreshTokenLifespan.Duration,
	}
}
