package auth

import (
	"context"
	"encoding/base64"
	"github.com/lyft/flytestdlib/errors"
	"github.com/lyft/flytestdlib/logger"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
)

type CookieHandler interface {
	RetrieveTokenValues(ctx context.Context, request *http.Request) (accessToken string, refreshToken string, err error)
}

type CookieManager struct {
	hashKey  []byte
	blockKey []byte
}

func NewCookieManager(ctx context.Context, hashKeyFile, blockKeyFile string) (CookieManager, error) {
	logger.Infof(ctx, "Instantiating cookie manager")
	hashKeyBytes, err := ioutil.ReadFile(hashKeyFile)
	if err != nil {
		return CookieManager{}, errors.Wrapf(ErrConfigFileRead, err, "Could not read hash key file")
	}
	blockKeyBytes, err := ioutil.ReadFile(blockKeyFile)
	if err != nil {
		return CookieManager{}, errors.Wrapf(ErrConfigFileRead, err, "Could not read block key file")
	}

	// Add error handling
	var hashKey, _ = base64.RawStdEncoding.DecodeString(string(hashKeyBytes))
	var blockKey, _ = base64.RawStdEncoding.DecodeString(string(blockKeyBytes))

	return CookieManager{
		hashKey:  hashKey,
		blockKey: blockKey,
	}, nil
}

// TODO: Separate refresh token from access token, remove named returns, and use stdlib errors.
func (c CookieManager) RetrieveTokenValues(ctx context.Context, request *http.Request) (accessToken string,
	refreshToken string, err error) {

	jwtCookie, err := request.Cookie(accessTokenCookie)
	if err != nil || jwtCookie == nil {
		logger.Errorf(ctx, "Could not detect existing access token cookie")
		return
	}
	logger.Debugf(ctx, "Existing JWT cookie found")
	accessToken, err = ReadSecureCookie(ctx, *jwtCookie, c.hashKey, c.blockKey)
	if err != nil {
		logger.Errorf(ctx, "Error reading existing secure JWT cookie %s", err)
		return
	}

	refreshCookie, err := request.Cookie(refreshTokenCookie)
	if err != nil || refreshCookie == nil {
		logger.Debugf(ctx, "Could not detect existing access token cookie")
		return
	}
	logger.Debugf(ctx, "Existing refresh cookie found")
	refreshToken, err = ReadSecureCookie(ctx, *refreshCookie, c.hashKey, c.blockKey)
	if err != nil {
		logger.Errorf(ctx, "Error reading existing secure refresh cookie %s", err)
		return
	}
	return
}

func (c CookieManager) SetTokenCookies(ctx context.Context, writer http.ResponseWriter, token *oauth2.Token) error {
	if token == nil {
		logger.Errorf(ctx, "Attempting to set cookies with nil token")
		return errors.Errorf(ErrTokenNil, "Attempting to set cookies with nil token")
	}

	jwtCookie, err := NewSecureCookie(accessTokenCookie, token.AccessToken, c.hashKey, c.blockKey)
	if err != nil {
		logger.Errorf(ctx, "Error generating encrypted JWT cookie %s", err)
		return err
	}
	http.SetCookie(writer, &jwtCookie)

	// Set the refresh cookie if there is a refresh token
	if token.RefreshToken != "" {
		refreshCookie, err := NewSecureCookie(refreshTokenCookie, token.RefreshToken, c.hashKey, c.blockKey)
		if err != nil {
			logger.Errorf(ctx, "Error generating encrypted refresh cookie %s", err)
			return err
		}
		http.SetCookie(writer, &refreshCookie)
	}
	return nil
}
