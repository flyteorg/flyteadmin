package oauthserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/auth"
	"github.com/flyteorg/flyteadmin/pkg/auth/config"

	"github.com/ory/fosite"

	"github.com/flyteorg/flyteadmin/pkg/auth/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	requestedScopePrefix = "f."
	accessTokenScope     = "access_token"
	refreshTokenScope    = "offline"
)

func getAuthEndpoint(authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authEndpoint(authCtx, writer, request)
	}
}

func getAuthCallbackEndpoint(authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authCallbackEndpoint(authCtx, writer, request)
	}
}

func getIssuer(cfg *config.Config, req *http.Request) string {
	if configIssuer := cfg.AppAuth.SelfAuthServer.Issuer; len(configIssuer) > 0 {
		return configIssuer
	}

	u, err := getRequestBaseUrl(req)
	if err != nil {
		// Should never happen
		logger.Error(context.Background(), err)
		return ""
	}

	return u.String()
}

func getRequestBaseUrl(r *http.Request) (*url.URL, error) {
	scheme := "http://"
	if r.TLS != nil {
		scheme = "https://"
	}

	return url.Parse(scheme + r.Host)
}

func authCallbackEndpoint(authCtx interfaces.AuthenticationContext, rw http.ResponseWriter, req *http.Request) {
	issuer := getIssuer(authCtx.Options(), req)

	// This context will be passed to all methods.
	ctx := req.Context()
	oauth2Provider := authCtx.OAuth2Provider()

	// Get the user's identity
	identityContext, err := auth.IdentityContextFromRequest(ctx, req, authCtx)
	if err != nil {
		logger.Infof(ctx, "Failed to acquire user identity from request: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, fosite.NewAuthorizeRequest(), err)
		return
	}

	// Get latest user's info either from identity or by making a UserInfo() call to the original
	userInfo, err := auth.QueryUserInfo(ctx, identityContext, req, authCtx)
	if err != nil {
		err = fmt.Errorf("failed to query user info. Error: %w", err)
		http.Error(rw, err.Error(), http.StatusUnauthorized)
		return
	}

	arUrl, err := authCtx.CookieManager().RetrieveAuthCodeRequest(ctx, req)
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, fosite.NewAuthorizeRequest(), err)
		return
	}

	arReq, err := http.NewRequest(http.MethodGet, arUrl, nil)
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, fosite.NewAuthorizeRequest(), err)
		return
	}

	ar, err := oauth2Provider.NewAuthorizeRequest(ctx, arReq)
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, ar, err)
		return
	}

	// TODO: Ideally this is where we show users a consent form.

	// let's see what scopes the user gave consent to
	for _, scope := range req.PostForm["scopes"] {
		ar.GrantScope(scope)
	}

	// Now that the user is authorized, we set up a session:
	mySessionData := oauth2Provider.NewJWTSessionToken(userInfo.Subject(), userInfo, ar.GetClient().GetID(), issuer, issuer)

	mySessionData.JWTClaims.ExpiresAt = time.Now().Add(time.Hour * 24)
	mySessionData.SetExpiresAt(fosite.AuthorizeCode, time.Now().Add(time.Hour*24))
	mySessionData.SetExpiresAt(fosite.AccessToken, time.Now().Add(time.Hour*24))

	// It's also wise to check the requested scopes, e.g.:
	// if ar.GetRequestedScopes().Has("admin") {
	//     http.Error(rw, "you're not allowed to do that", http.StatusForbidden)
	//     return
	// }

	// Now we need to get a response. This is the place where the AuthorizeEndpointHandlers kick in and start processing the request.
	// NewAuthorizeResponse is capable of running multiple response type handlers which in turn enables this library
	// to support open id connect.
	response, err := oauth2Provider.NewAuthorizeResponse(ctx, ar, mySessionData)

	// Catch any errors, e.g.:
	// * unknown client
	// * invalid redirect
	// * ...
	if err != nil {
		log.Printf("Error occurred in NewAuthorizeResponse: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, ar, err)
		return
	}

	// Last but not least, send the response!
	oauth2Provider.WriteAuthorizeResponse(rw, ar, response)
}

func authEndpoint(authCtx interfaces.AuthenticationContext, rw http.ResponseWriter, req *http.Request) {
	// This context will be passed to all methods.
	ctx := req.Context()

	oauth2Provider := authCtx.OAuth2Provider()

	// Let's create an AuthorizeRequest object!
	// It will analyze the request and extract important information like scopes, response type and others.
	ar, err := oauth2Provider.NewAuthorizeRequest(ctx, req)
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, ar, err)
		return
	}

	err = authCtx.CookieManager().SetAuthCodeCookie(ctx, rw, req.URL.String())
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, ar, err)
		return
	}

	redirectUrl := fmt.Sprintf("/login?redirect_url=%v", authorizeCallbackRelativeUrl.String())
	http.Redirect(rw, req, redirectUrl, http.StatusTemporaryRedirect)
}
