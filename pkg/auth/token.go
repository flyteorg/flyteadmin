package auth

import (
	"context"
	"strings"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/auth/interfaces"

	"github.com/coreos/go-oidc"
	"github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"golang.org/x/oauth2"
)

const (
	ErrRefreshingToken errors.ErrorCode = "TOKEN_REFRESH_FAILURE"
	ErrTokenExpired    errors.ErrorCode = "JWT_EXPIRED"
	ErrJwtValidation   errors.ErrorCode = "JWT_VERIFICATION_FAILED"
)

// Refresh a JWT
func GetRefreshedToken(ctx context.Context, oauth *oauth2.Config, accessToken, refreshToken string) (*oauth2.Token, error) {
	logger.Debugf(ctx, "Attempting to refresh token")
	originalToken := oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-1 * time.Minute), // force expired by setting to the past
	}

	tokenSource := oauth.TokenSource(ctx, &originalToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		logger.Errorf(ctx, "Error refreshing token %s", err)
		return nil, errors.Wrapf(ErrRefreshingToken, err, "Error refreshing token")
	}

	return newToken, nil
}

func ParseIDTokenAndValidate(ctx context.Context, clientID, rawIDToken string, provider *oidc.Provider) (*oidc.IDToken, error) {
	var verifier = provider.Verifier(&oidc.Config{
		ClientID: clientID,
	})

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		logger.Debugf(ctx, "JWT parsing with claims failed %s", err)
		flyteErr := errors.Wrapf(ErrJwtValidation, err, "jwt parse with claims failed")
		// TODO: Contribute an errors package to the go-oidc library for proper error handling
		if strings.Contains(err.Error(), "token is expired") {
			return idToken, errors.Wrapf(ErrTokenExpired, flyteErr, "token is expired")
		}

		return idToken, flyteErr
	}

	return idToken, nil
}

// This function attempts to extract a token from the context, and will then call the validation function, passing up
// any errors.
func GRPCGetIdentityFromAccessToken(ctx context.Context, authCtx interfaces.AuthenticationContext) (
	interfaces.IdentityContext, error) {

	tokenStr, err := grpcauth.AuthFromMD(ctx, BearerScheme)
	if err != nil {
		logger.Debugf(ctx, "Could not retrieve bearer token from metadata %v", err)
		return nil, errors.Wrapf(ErrJwtValidation, err, "Could not retrieve bearer token from metadata")
	}

	if tokenStr == "" {
		logger.Debugf(ctx, "Found Bearer scheme but token was blank")
		return nil, errors.Errorf(ErrJwtValidation, "%v token is blank", IDTokenScheme)
	}

	return authCtx.OAuth2Provider().ValidateAccessToken(ctx, tokenStr)
}

// This function attempts to extract a token from the context, and will then call the validation function, passing up
// any errors.
func GRPCGetIdentityFromIDToken(ctx context.Context, clientID string, provider *oidc.Provider) (
	interfaces.IdentityContext, error) {

	tokenStr, err := grpcauth.AuthFromMD(ctx, IDTokenScheme)
	if err != nil {
		logger.Debugf(ctx, "Could not retrieve bearer token from metadata %v", err)
		return nil, errors.Wrapf(ErrJwtValidation, err, "Could not retrieve bearer token from metadata")
	}

	if tokenStr == "" {
		logger.Debugf(ctx, "Found Bearer scheme but token was blank")
		return nil, errors.Errorf(ErrJwtValidation, "%v token is blank", IDTokenScheme)
	}

	return IdentityContextFromIDTokenToken(ctx, tokenStr, clientID, provider)
}

func IdentityContextFromIDTokenToken(ctx context.Context, tokenStr, clientID string, provider *oidc.Provider) (
	interfaces.IdentityContext, error) {
	idToken, err := ParseIDTokenAndValidate(ctx, clientID, tokenStr, provider)
	if err != nil {
		return nil, err
	}

	return NewIdentityContext(idToken.Audience[0], idToken.Subject, "", idToken.IssuedAt, UserInfoResponse{}), nil
}
