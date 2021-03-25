package oauthserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	oauth22 "github.com/ory/fosite/handler/oauth2"

	"github.com/flyteorg/flyteadmin/pkg/auth"

	"github.com/ory/fosite"

	"github.com/flyteorg/flyteadmin/pkg/auth/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flytestdlib/logger"
)

func getAuthEndpoint(authContext interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authEndpoint(authContext, writer, request)
	}
}

type AuthContextSetter func(ctx context.Context) (context.Context, error)

func getAuthCallbackEndpoint(authContextSetter AuthContextSetter, authContext interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authCallbackEndpoint(authContextSetter, authContext, writer, request)
	}
}

// Returns the unique string which identifies the authenticated end user (if any).
func getUser(ctx context.Context) string {
	principalContextUser := ctx.Value(common.PrincipalContextKey)
	if principalContextUser != nil {
		return fmt.Sprintf("%v", principalContextUser)
	}

	return ""
}

type InCodeAuthorizeCodeProvider struct {
	oauth22.CoreStrategy
}

func (p InCodeAuthorizeCodeProvider) AuthorizeCodeSignature(token string) string {
	return token
}

func (p InCodeAuthorizeCodeProvider) GenerateAuthorizeCode(ctx context.Context, requester fosite.Requester) (token string, signature string, err error) {
	token, _, err = p.CoreStrategy.GenerateAccessToken(ctx, requester)
	return token, token, err
}

func (p InCodeAuthorizeCodeProvider) ValidateAuthorizeCode(ctx context.Context, requester fosite.Requester, token string) (err error) {
	return p.CoreStrategy.ValidateAccessToken(ctx, requester, token)
}

func authCallbackEndpoint(authContextSetter AuthContextSetter, authContext interfaces.AuthenticationContext, rw http.ResponseWriter, req *http.Request) {
	// This context will be passed to all methods.
	ctx := req.Context()
	idTokenRaw, _, _, err := authContext.CookieManager().RetrieveTokenValues(ctx, req)
	if err != nil {
		http.Error(rw, "Error decoding identify token, try /login in again", http.StatusUnauthorized)
		return
	}

	idToken, err := auth.ParseAndValidate(ctx, authContext.Claims(), idTokenRaw, authContext.OidcProvider())
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2.WriteAuthorizeError(rw, fosite.NewAuthorizeRequest(), err)
		return
	}

	arUrl, err := authContext.CookieManager().RetrieveAuthCodeRequest(ctx, req)
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2.WriteAuthorizeError(rw, fosite.NewAuthorizeRequest(), err)
		return
	}

	arReq, err := http.NewRequest(http.MethodGet, arUrl, nil)
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2.WriteAuthorizeError(rw, fosite.NewAuthorizeRequest(), err)
		return
	}

	ar, err := oauth2.NewAuthorizeRequest(ctx, arReq)
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2.WriteAuthorizeError(rw, ar, err)
		return
	}

	// let's see what scopes the user gave consent to
	for _, scope := range req.PostForm["scopes"] {
		ar.GrantScope(scope)
	}

	// Now that the user is authorized, we set up a session:
	mySessionData := newJWTSessionToken(idToken.Subject)

	// When using the HMACSHA strategy you must use something that implements the HMACSessionContainer.
	// It brings you the power of overriding the default values.
	//
	// mySessionData.HMACSession = &strategy.HMACSession{
	//	AccessTokenExpiry: time.Now().Add(time.Day),
	//	AuthorizeCodeExpiry: time.Now().Add(time.Day),
	// }
	//

	// If you're using the JWT strategy, there's currently no distinction between access token and authorize code claims.
	// Therefore, you both access token and authorize code will have the same "exp" claim. If this is something you
	// need let us know on github.
	//
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
	response, err := oauth2.NewAuthorizeResponse(ctx, ar, mySessionData)

	// Catch any errors, e.g.:
	// * unknown client
	// * invalid redirect
	// * ...
	if err != nil {
		log.Printf("Error occurred in NewAuthorizeResponse: %+v", err)
		oauth2.WriteAuthorizeError(rw, ar, err)
		return
	}

	// Last but not least, send the response!
	oauth2.WriteAuthorizeResponse(rw, ar, response)
}

func authEndpoint(authContext interfaces.AuthenticationContext, rw http.ResponseWriter, req *http.Request) {
	// This context will be passed to all methods.
	ctx := req.Context()

	// Let's create an AuthorizeRequest object!
	// It will analyze the request and extract important information like scopes, response type and others.
	ar, err := oauth2.NewAuthorizeRequest(ctx, req)
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2.WriteAuthorizeError(rw, ar, err)
		return
	}

	err = authContext.CookieManager().SetAuthCodeCookie(ctx, rw, req.URL.String())
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2.WriteAuthorizeError(rw, ar, err)
		return
	}

	redirectUrl := fmt.Sprintf("/login?redirect_url=/oauth2/authorize_callback")
	http.Redirect(rw, req, redirectUrl, http.StatusTemporaryRedirect)
}
