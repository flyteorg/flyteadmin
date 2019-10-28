package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gorilla/securecookie"
	"github.com/lyft/flytestdlib/errors"
	"github.com/lyft/flytestdlib/logger"
	"math/rand"
	"net/http"
	"time"
)

const accessTokenCookie = "flyte_jwt"
const refreshTokenCookie = "flyte_refresh"
const csrfStateCookie = "flyte_csrf_state"

const (
	ErrSecureCookie errors.ErrorCode = "SECURE_COOKIE_ERROR"
)

var AllowedChars = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func HashCsrfState(csrf string) string {
	shaBytes := sha256.Sum256([]byte(csrf))
	hash := hex.EncodeToString(shaBytes[:])
	return hash
}

func NewSecureCookie(cookieName, value string, hashKey, blockKey []byte) (http.Cookie, error) {
	var s = securecookie.New(hashKey, blockKey)
	if encoded, err := s.Encode(cookieName, value); err == nil {
		return http.Cookie{
			Name:  cookieName,
			Value: encoded,
		}, nil
	} else {
		return http.Cookie{}, errors.Wrapf(ErrSecureCookie, err, "Error creating secure cookie")
	}
}

func ReadSecureCookie(ctx context.Context, cookie http.Cookie, hashKey, blockKey []byte) (string, error) {
	var s = securecookie.New(hashKey, blockKey)
	var value string
	if err := s.Decode(cookie.Name, cookie.Value, &value); err == nil {
		return value, nil
	} else {
		logger.Errorf(ctx, "Error reading secure cookie %s %s", cookie.Name, err)
		return "", errors.Wrapf(ErrSecureCookie, err, "Error reading secure cookie %s", cookie.Name)
	}
}

func NewCsrfToken(seed int64) string {
	rand.Seed(seed)
	csrfToken := [10]rune{}
	for i := 0; i < len(csrfToken); i++ {
		csrfToken[i] = AllowedChars[rand.Intn(len(AllowedChars))]
	}
	return string(csrfToken[:])
}

func NewCsrfCookie() http.Cookie {
	csrfStateToken := NewCsrfToken(time.Now().UnixNano())
	return http.Cookie{
		Name:     csrfStateCookie,
		Value:    csrfStateToken,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
	}
}

func VerifyCsrfCookie(ctx context.Context, request *http.Request) bool {
	csrfState := request.FormValue(CsrfFormKey)
	if csrfState == "" {
		logger.Errorf(ctx, "Empty state in callback, %s", request.Form)
		return false
	}
	csrfCookie, err := request.Cookie(csrfStateCookie)
	if csrfCookie == nil || err != nil {
		logger.Errorf(ctx, "Could not find csrf cookie")
		return false
	}
	if HashCsrfState(csrfCookie.Value) != csrfState {
		logger.Errorf(ctx, "CSRF token does not match state %s, %s vs %s", csrfCookie.Value, HashCsrfState(csrfCookie.Value), csrfState)
		return false
	}
	return true
}
