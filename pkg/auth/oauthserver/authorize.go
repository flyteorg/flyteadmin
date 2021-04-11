package oauthserver

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/auth"
	"github.com/flyteorg/flyteadmin/pkg/auth/config"
	"github.com/gtank/cryptopasta"

	"github.com/ory/fosite"

	"github.com/flyteorg/flyteadmin/pkg/auth/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	requestedScopePrefix   = "f."
	accessTokenScope       = "access_token"
	refreshTokenScope      = "offline"
	codeChallengeFormParam = "code"
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

func getIssuer(cfg *config.Config) string {
	if configIssuer := cfg.AppAuth.SelfAuthServer.Issuer; len(configIssuer) > 0 {
		return configIssuer
	}

	return cfg.HTTPPublicUri.String()
}

func encryptString(plainTextCode string, blockKey [auth.SymmetricKeyLength]byte) (string, error) {
	cypher, err := cryptopasta.Encrypt([]byte(plainTextCode), &blockKey)
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(cypher), nil
}

func decryptString(encryptedEncoded string, blockKey [auth.SymmetricKeyLength]byte) (string, error) {
	cypher, err := base64.RawStdEncoding.DecodeString(encryptedEncoded)
	if err != nil {
		return "", err
	}

	raw, err := cryptopasta.Decrypt(cypher, &blockKey)
	return string(raw), err
}

func authCallbackEndpoint(authCtx interfaces.AuthenticationContext, rw http.ResponseWriter, req *http.Request) {
	issuer := getIssuer(authCtx.Options())

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
	mySessionData := oauth2Provider.NewJWTSessionToken(identityContext.UserID(), userInfo, ar.GetClient().GetID(), issuer, issuer)
	mySessionData.JWTClaims.ExpiresAt = time.Now().Add(time.Hour * 24)
	mySessionData.SetExpiresAt(fosite.AuthorizeCode, time.Now().Add(time.Hour*24))
	mySessionData.SetExpiresAt(fosite.AccessToken, time.Now().Add(time.Hour*24))

	// Now we need to get a response. This is the place where the AuthorizeEndpointHandlers kick in and start processing the request.
	// NewAuthorizeResponse is capable of running multiple response type handlers which in turn enables this library
	// to support open id connect.
	response, err := oauth2Provider.NewAuthorizeResponse(ctx, ar, mySessionData)
	if err != nil {
		log.Printf("Error occurred in NewAuthorizeResponse: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, ar, err)
		return
	}

	// Last but not least, send the response!
	oauth2Provider.WriteAuthorizeResponse(rw, ar, response)
}

// Get the /authorize endpoint handler that is supposed to be invoked in the browser for the user to log in and consent.
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
