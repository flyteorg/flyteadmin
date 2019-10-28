package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gorilla/securecookie"
	"github.com/lyft/flyteadmin/pkg/auth/interfaces"
	"github.com/lyft/flytestdlib/errors"
	"github.com/lyft/flytestdlib/logger"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

const (
	accessTokenCookie  = "flyte_jwt"
	refreshTokenCookie = "flyte_refresh"
	csrfStateCookie    = "flyte_csrf_state"
	redirectUrlCookie  = "flyte_redirect_location"
)

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

// This function takes in a string and returns a cookie that's used to keep track of where to send the user after
// the OAuth2 login flow is complete.
func NewRedirectCookie(ctx context.Context, redirectUrl string) *http.Cookie {
	urlObj, err := url.Parse(redirectUrl)
	if err != nil || urlObj == nil {
		logger.Errorf(ctx, "Error creating redirect cookie %s %s", urlObj, err)
		return nil
	}

	if urlObj.EscapedPath() == "" {
		logger.Errorf(ctx, "Error parsing URL, redirect %s resolved to empty string", redirectUrl)
		return nil
	}

	return &http.Cookie{
		Name:     redirectUrlCookie,
		Value:    urlObj.EscapedPath(),
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
	}
}

// At the end of the OAuth flow, the server needs to send the user somewhere. This should have been stored as a cookie
// during the initial /login call. If that cookie is missing from the request, it will default to the one configured
// in this package's Config object.
func getAuthFlowEndRedirect(ctx context.Context, authContext interfaces.AuthenticationContext, request *http.Request) string {
	cookie, err := request.Cookie(redirectUrlCookie)
	if err != nil {
		logger.Debugf(ctx, "Could not detect end-of-flow redirect url cookie")
		return authContext.Options().RedirectUrl
	}
	return cookie.Value
}
