package auth

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/lyft/flyteadmin/pkg/config"
	"github.com/lyft/flytestdlib/logger"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"net/http"
	"strings"
)

// Move to config file
func GetOauth2Config(options config.OauthOptions) (oauth2.Config, error) {
	secretBytes, err := ioutil.ReadFile(options.ClientSecretFile)
	if err != nil {
		return oauth2.Config{}, err
	}
	secret := strings.TrimSuffix(string(secretBytes), "\n")
	return oauth2.Config{
		RedirectURL:  options.RedirectUrl,
		ClientID:     options.ClientId,
		ClientSecret: secret,
		// Offline access needs to be specified in order to return a refresh token in the exchange.
		// TODO: Second parameter is IDP specific - move to config. Also handle case where a refresh token is not allowed
		Scopes: []string{OidcScope, OfflineAccessType},
		Endpoint: oauth2.Endpoint{
			AuthURL:  options.AuthorizeUrl,
			TokenURL: options.TokenUrl,
		},
	}, nil
}

var AllowedChars = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

// Look for access token and refresh token, if both are present and the access token is expired, then attempt to
// refresh. Otherwise do nothing and proceed to the next handler. If successfully refreshed, proceed to the landing page.
func RefreshTokensIfExists(ctx context.Context, oauth oauth2.Config, jwtVerifier JwtVerifier, cookieManager CookieManager,
		handlerFunc http.HandlerFunc) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {
		// Since we only do one thing if there are no errors anywhere along the chain, we can save code by just
		// using one variable and checking for errors at the end.
		var err error
		accessToken, refreshToken, err := cookieManager.RetrieveTokenValues(ctx, request)
		if err == nil && accessToken != "" && refreshToken != "" {
			jwtToken, e := jwt.Parse(accessToken, jwtVerifier.GetKey)
			err = e
			if validationErr, ok := err.(*jwt.ValidationError); ok &&
				validationErr.Errors == jwt.ValidationErrorExpired && jwtToken != nil {
				logger.Debugf(ctx, "Expired access token found, attempting to refresh")
				newToken, e := GetRefreshedToken(ctx, oauth, *jwtToken, refreshToken)
				err = e
				if err == nil {
					logger.Debugf(ctx, "Access token refreshed. Saving new tokens into cookies.")
					err = cookieManager.SetTokenCookies(ctx, writer, newToken)

					if err == nil {
						http.Redirect(writer, request, "/api/v1/projects", http.StatusTemporaryRedirect)
						return
					}
				}
			}
		}

		if err != nil {
			logger.Errorf(ctx, "Error in refresh token handler %s", err)
			handlerFunc(writer, request)
			return
		}
	}
}

func GetLoginHandler(ctx context.Context, oauth oauth2.Config) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		csrfCookie := NewCsrfCookie()
		csrfToken := csrfCookie.Value
		http.SetCookie(writer, &csrfCookie)

		state := HashCsrfState(csrfToken)
		logger.Debugf(ctx, "Setting CSRF state cookie to %s and state to %s\n", csrfToken, state)
		url := oauth.AuthCodeURL(state)
		fmt.Println(url)

		http.Redirect(writer, request, url, http.StatusTemporaryRedirect)
	}
}

func GetCallbackHandler(ctx context.Context, oauth oauth2.Config, manager CookieManager) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		logger.Infof(ctx, "In callback: **")
		authorizationCode := request.FormValue(AuthorizationResponseCodeType)
		fmt.Printf("Auth code: %s\n", authorizationCode)

		validCsrf, err := VerifyCsrfCookie(ctx, request)
		if !validCsrf {
			logger.Infof(ctx, "Invalid CSRF token cookie, error [%s]", err)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		// TODO: The second parameter is IDP specific but seems to be convention, make configurable anyways.
		// The second parameter is necessary to get the initial refresh token
		offlineAccessParam := oauth2.SetAuthURLParam(RefreshToken, OfflineAccessType)

		token, err := oauth.Exchange(ctx, authorizationCode, offlineAccessParam)
		if err != nil {
			logger.Errorf(ctx, "Error when exchanging code %s", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}
		fmt.Printf("Token.Access: %s\nToken.Refresh: %s\nToken.Type: %s\nToken.Expiry: %v\n",
			token.AccessToken, token.RefreshToken, token.TokenType, token.Expiry.Unix())

		err = manager.SetTokenCookies(ctx, writer, token)
		if err != nil {
			logger.Errorf(ctx, "Error setting encrypted JWT cookie %s", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		http.Redirect(writer, request, "/api/v1/projects", http.StatusTemporaryRedirect)
	}
}

func AuthenticationLoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Invoke 'handler' to use your gRPC server implementation and get
	// the response.
	fmt.Printf("gRPC server info in logging interceptor email %s method %s\n", ctx.Value("email"), info.FullMethod)
	return handler(ctx, req)
}

// This function will only look for a token from the request metadata, verify it, and extract the user email if valid.
// Unless there is an error, it will not return an unauthorized status. That is up to subsequent functions to decide,
// based on configuration.  We don't want to require authentication for all endpoints.
func GetAuthenticationInterceptor(oauth config.OauthOptions) func(context.Context) (context.Context, error) {
	// TODO: Use library after library has been written
	jwtVerifier := JwtVerifier{
		Url: oauth.JwksUrl,
	}
	return func(ctx context.Context) (context.Context, error) {
		logger.Debugf(ctx,"Running authentication gRPC interceptor")
		// Check if auth is enabled at all
		// Selectively apply authentication based on configuration to only certain endpoints
		tokenStr, err := grpcauth.AuthFromMD(ctx, "bearer")
		if err != nil {
			logger.Debugf(ctx,"Could not retrieve bearer token from metadata %v", err)
			return ctx, nil
		}

		// Currently auth is optional
		if tokenStr == "" {
			logger.Debugf(ctx,"Bearer token is empty, skipping parsing")
			return ctx, nil
		}

		token, err := jwt.Parse(tokenStr, jwtVerifier.GetKey)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "could not parse token string into object: %s %s", tokenStr, err)
		}

		claims := token.Claims.(jwt.MapClaims)
		claims.VerifyAudience("api://default", true)
		claims.VerifyIssuer(oauth.Issuer, true)

		if userEmail, ok := claims["sub"]; ok {
			if userEmail.(string) != "" {
				logger.Debugf(ctx, "Authenticated request for %s", userEmail.(string))
				newCtx := WithUserEmail(context.WithValue(ctx, "tokenInfo", tokenStr), userEmail.(string))
				return newCtx, nil
			}
		}

		return nil, status.Errorf(codes.Unauthenticated, "no email or empty email found")
	}
}

func WithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, "email", email)
}

type HttpRequestToMetadataAnnotator func(ctx context.Context, request *http.Request) metadata.MD

// This is effectively middleware for the grpc gateway, it allows us to modify the translation between HTTP request
// and gRPC request.
func GetHttpRequestCookieToMetadataHandler(cookieManager CookieManager) HttpRequestToMetadataAnnotator {
	return func(ctx context.Context, request *http.Request) metadata.MD {
		// TODO: Improve error handling
		accessToken, _, _ := cookieManager.RetrieveTokenValues(ctx, request)
		if accessToken == "" {
			logger.Infof(ctx, "Could not find access token cookie while requesting %s", request.RequestURI)
			return nil
		}
		fmt.Printf("HTTP Annotation: Decoded access token %s\n", accessToken)
		return metadata.MD{
			"authorization": []string{fmt.Sprintf("Bearer %s", accessToken)},
		}
	}
}