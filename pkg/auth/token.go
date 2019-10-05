package auth

import (
	"context"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lyft/flyteadmin/pkg/config"
	"github.com/lyft/flytestdlib/errors"
	"github.com/lyft/flytestdlib/logger"
	"golang.org/x/oauth2"
	"time"
)

const (
	ErrJwtVerification errors.ErrorCode = "JWKS_FAILURE"
)

type JwtVerifier struct {
	Url string
}

func (j *JwtVerifier) GetKey(token *jwt.Token) (interface{}, error) {
	set, err := jwk.FetchHTTP(j.Url)
	if err != nil {
		return nil, err
	}

	keyID, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.Errorf(ErrJwtVerification, "JWT header is missing a kid")
	}

	if key := set.LookupKeyID(keyID); len(key) == 1 {
		return key[0].Materialize()
	}

	return nil, errors.Errorf(ErrJwtVerification, "unable to find key")
}

const (
	ErrRefreshingToken errors.ErrorCode = "TOKEN_REFRESH_FAILURE"
	ErrTokenExpired                     = "JWT_EXPIRED"
	ErrJwtValidation                    = "JWT_VERIFICATION_FAILED"
	ErrNilJwt                           = "NIL_JWT_ERROR"
	ErrEmptySub                         = "SUB_EMPTY"
)

func getExpiration(claims jwt.MapClaims) *time.Time {
	switch exp := claims["exp"].(type) {
	case float64:
		t := time.Unix(int64(exp), 0)
		return &t
	case json.Number:
		v, _ := exp.Int64()
		t := time.Unix(v, 0)
		return &t
	default:
		return nil
	}
}

// Refreshes a JWT.
// NB: The input token and the output token are from two different libraries
func GetRefreshedToken(ctx context.Context, oauth oauth2.Config, token jwt.Token, refreshToken string) (*oauth2.Token, error) {
	logger.Debugf(ctx, "Attempting to refresh token")
	claims := token.Claims.(jwt.MapClaims)
	expiration := getExpiration(claims)
	var originalToken oauth2.Token
	if expiration == nil {
		logger.Infof(ctx, "No expiration claim found")
		originalToken = oauth2.Token{
			AccessToken:  token.Raw,
			RefreshToken: refreshToken,
			Expiry:       time.Now().Add(-1 * time.Minute), // force expired by setting to the past
		}
	} else {
		logger.Infof(ctx, "Refreshing access token with expiration %s", expiration)
		originalToken = oauth2.Token{
			AccessToken:  token.Raw,
			RefreshToken: refreshToken,
			Expiry:       *expiration,
		}
	}

	tokenSource := oauth.TokenSource(ctx, &originalToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		logger.Errorf(ctx, "Error refreshing token %s", err)
		return nil, errors.Wrapf(ErrRefreshingToken, err, "Error refreshing token")
	}

	return newToken, nil
}

func constructMapClaims(claims config.Claims) jwt.MapClaims {
	// TODO: iterate
	return jwt.MapClaims{
		"aud": claims.Audience,
		"iss": claims.Issuer,
	}
}

func ParseAndValidate(ctx context.Context, options config.OauthOptions, accessToken string,
	keyFunc jwt.Keyfunc) (*jwt.Token, error) {

	t, err := jwt.ParseWithClaims(accessToken, constructMapClaims(options.Claims), keyFunc)
	if err != nil {
		logger.Debugf(ctx, "JWT parsing with claims failed %s", err)
		flyteErr := errors.Wrapf(ErrJwtValidation, err, "jwt parse with claims failed")
		if validationErr, ok := err.(*jwt.ValidationError); ok &&
			validationErr.Errors == jwt.ValidationErrorExpired && t != nil {
			return t, errors.Wrapf(ErrTokenExpired, flyteErr, "token is expired")
		}
		return t, flyteErr
	}

	return t, nil
}

func GetSubClaim(t *jwt.Token) (string, error) {
	claims := t.Claims.(jwt.MapClaims)
	if userEmail, ok := claims["sub"]; ok {
		if userEmail.(string) != "" {
			return userEmail.(string), nil
		} else {
			return "", errors.Errorf(ErrEmptySub, "Email not found in sub claim")
		}
	}
	return "", errors.Errorf(ErrEmptySub, "No sub claim found")
}
