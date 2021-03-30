package oauthserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/auth/config"
	"golang.org/x/oauth2"

	oauth22 "github.com/ory/fosite/handler/oauth2"

	"github.com/flyteorg/flyteadmin/pkg/auth"

	"github.com/ory/fosite"

	"github.com/flyteorg/flyteadmin/pkg/auth/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
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

func getIssuer(cfg *config.Config, req *http.Request) string {
	if configIssuer := cfg.AppAuth.Issuer.Issuer; len(configIssuer) == 0 {
		return configIssuer
	}

	return "https://" + req.Host + "/"
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
	idTokenRaw, accessToken, _, err := authCtx.CookieManager().RetrieveTokenValues(ctx, req)
	if err != nil {
		http.Error(rw, "Error decoding identify token, try /login in again", http.StatusUnauthorized)
		return
	}

	// TODO: Move into a more generic GetUserIdentityContext() method...
	idToken, err := auth.ParseIDTokenAndValidate(ctx, authCtx.Options().UserAuth.OpenID.ClientID, idTokenRaw, authCtx.OidcProvider())
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, fosite.NewAuthorizeRequest(), err)
		return
	}

	originalToken := oauth2.Token{
		AccessToken: accessToken,
	}

	tokenSource := authCtx.OAuth2ClientConfig().TokenSource(ctx, &originalToken)

	// TODO: Investigate improving transparency of errors. The errors from this call may be just a local error, or may
	//       be an error from the HTTP request to the IDP. In the latter case, consider passing along the error code/msg.
	userInfo, err := authCtx.OidcProvider().UserInfo(ctx, tokenSource)
	if err != nil {
		logger.Errorf(ctx, "Error getting user info from IDP %s", err)
		http.Error(rw, "Error getting user info from IDP", http.StatusFailedDependency)
		return
	}

	resp := auth.UserInfoResponse{}
	err = userInfo.Claims(&resp)
	if err != nil {
		logger.Errorf(ctx, "Error getting user info from IDP %s", err)
		http.Error(rw, "Error getting user info from IDP", http.StatusFailedDependency)
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
	mySessionData := oauth2Provider.NewJWTSessionToken(idToken.Subject, resp, ar.GetClient().GetID(), issuer, issuer)

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
