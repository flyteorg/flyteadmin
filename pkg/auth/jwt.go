package auth

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lyft/flytestdlib/logger"
	"golang.org/x/oauth2"
	"time"
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
		return nil, errors.New("JWT header is missing a kid")
	}

	if key := set.LookupKeyID(keyID); len(key) == 1 {
		return key[0].Materialize()
	}

	return nil, errors.New("unable to find key")
}

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

	claims := token.Claims.(jwt.MapClaims)
	expiration := getExpiration(claims)
	var originalToken oauth2.Token
	if expiration == nil {
		logger.Infof(ctx, "No expiration claim found")
		// force expired by setting to the past
		originalToken = oauth2.Token{
			AccessToken:  token.Raw,
			RefreshToken: refreshToken,
			Expiry:       time.Now().Add(-1 * time.Minute),
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
		return nil, err
	}

	return newToken, nil
}
