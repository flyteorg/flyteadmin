package auth

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/lyft/flytestdlib/errors"
	"github.com/lyft/flytestdlib/logger"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net/http"
)

// Look for access token and refresh token, if both are present and the access token is expired, then attempt to
// refresh. Otherwise do nothing and proceed to the next handler. If successfully refreshed, proceed to the landing page.
func RefreshTokensIfExists(ctx context.Context, oauth oauth2.Config, options OAuthOptions, provider *oidc.Provider,
	cookieManager CookieManager, handlerFunc http.HandlerFunc) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {
		// Since we only do one thing if there are no errors anywhere along the chain, we can save code by just
		// using one variable and checking for errors at the end.
		var err error
		accessToken, refreshToken, err := cookieManager.RetrieveTokenValues(ctx, request)

		if err == nil && accessToken != "" && refreshToken != "" {
			_, e := ParseAndValidate(ctx, options, accessToken, provider)
			err = e
			if err != nil && errors.IsCausedBy(err, ErrTokenExpired) {
				logger.Debugf(ctx, "Expired access token found, attempting to refresh")
				newToken, e := GetRefreshedToken(ctx, oauth, accessToken, refreshToken)
				err = e
				if err == nil {
					logger.Debugf(ctx, "Access token refreshed. Saving new tokens into cookies.")
					err = cookieManager.SetTokenCookies(ctx, writer, newToken)
				}
			}
		}

		if err != nil {
			logger.Errorf(ctx, "Non-expiration error in refresh token handler %s, redirecting to login handler", err)
			handlerFunc(writer, request)
			return
		} else {
			http.Redirect(writer, request, options.RedirectUrl, http.StatusTemporaryRedirect)
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

func GetCallbackHandler(ctx context.Context, oauth oauth2.Config, options OAuthOptions, manager CookieManager) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		logger.Debugf(ctx, "Running callback handler...")
		authorizationCode := request.FormValue(AuthorizationResponseCodeType)

		validCsrf := VerifyCsrfCookie(ctx, request)
		if !validCsrf {
			logger.Infof(ctx, "Invalid CSRF token cookie")
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

		err = manager.SetTokenCookies(ctx, writer, token)
		if err != nil {
			logger.Errorf(ctx, "Error setting encrypted JWT cookie %s", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		http.Redirect(writer, request, options.RedirectUrl, http.StatusTemporaryRedirect)
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
func GetAuthenticationInterceptor(options OAuthOptions, provider *oidc.Provider) func(context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		logger.Debugf(ctx, "Running authentication gRPC interceptor")
		tokenStr, err := grpcauth.AuthFromMD(ctx, "bearer")
		if err != nil {
			logger.Debugf(ctx, "Could not retrieve bearer token from metadata %v", err)
			return ctx, nil
		}

		// Currently auth is optional...
		if tokenStr == "" {
			logger.Debugf(ctx, "Bearer token is empty, skipping parsing")
			return ctx, nil
		}

		// ...however, if there is a bearer token, but there are additional errors downstream, then we return an
		// authentication error.
		token, err := ParseAndValidate(ctx, options, tokenStr, provider)
		if err != nil {
			return ctx, status.Errorf(codes.Unauthenticated, "could not parse token string into object: %s %s", tokenStr, err)
		}
		if token == nil {
			return ctx, status.Errorf(codes.Unauthenticated, "Token was nil after parsing %s", tokenStr)
		} else if token.Subject == "" {
			return ctx, status.Errorf(codes.Unauthenticated, "no email or empty email found")
		} else {
			newCtx := WithUserEmail(context.WithValue(ctx, "bearer", tokenStr), token.Subject)
			return newCtx, nil
		}
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
