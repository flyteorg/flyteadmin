package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/lyft/flytestdlib/logger"
	"math/rand"
	"net/http"
	_ "net/http/pprof" // Required to serve application.
	"time"
)

const accessTokenCookie = "my-jwt-cookie"
const refreshTokenCookie = "my-refresh-cookie"
const csrfStateCookie = "my-csrf-state"

// TODO: Replace with proper key management and remove from code - local testing keys only
var hashKey, _ = base64.RawStdEncoding.DecodeString("rekd8I2k9UTE+qK3+8hVoG9c5CEM7QuYoNfAszkPf1I")
var blockKey, _ = base64.RawStdEncoding.DecodeString("jLDuQkVQyps4QLJEZaGRHiZbFWe8V5oTxZ7bmNkL7Iw")

func HashCsrfState(csrf string) string {
	shaBytes := sha256.Sum256([]byte(csrf))
	hash := hex.EncodeToString(shaBytes[:])
	fmt.Printf("Hashed |%s| to |%s|\n", csrf, hash)
	return hash
}

func NewSecureCookie(cookieName, value string) (http.Cookie, error) {
	var s = securecookie.New(hashKey, blockKey)
	if encoded, err := s.Encode(cookieName, value); err == nil {
		return http.Cookie{
			Name:  cookieName,
			Value: encoded,
		}, nil
	} else {
		return http.Cookie{}, err
	}
}

func ReadSecureCookie(ctx context.Context, cookie http.Cookie) (string, error) {
	var s = securecookie.New(hashKey, blockKey)
	var value string
	if err := s.Decode(cookie.Name, cookie.Value, &value); err == nil {
		fmt.Printf("Decrypted %s to\n%s", cookie.Value, value)
		return value, nil
	} else {
		logger.Errorf(ctx, "Error reading secure cookie %s %s", cookie.Name, err)
		return "", err
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

func VerifyCsrfCookie(ctx context.Context, request *http.Request) (bool, error) {
	if request == nil {
		logger.Error(ctx, "Cannot verify csrf from an empty HTTP request")
		return false, errors.New("request was nil")
	}
	csrfState := request.FormValue(CsrfFormKey)
	if csrfState == "" {
		logger.Errorf(ctx, "Empty state in callback, %s", request.Form)
		return false, nil
	}
	csrfCookie, err := request.Cookie(csrfStateCookie)
	if csrfCookie == nil || err != nil {
		logger.Errorf(ctx, "Could not find csrf cookie")
		return false, nil
	}
	if HashCsrfState(csrfCookie.Value) != csrfState {
		logger.Errorf(ctx, "CSRF token does not match state %s, %s vs %s", csrfCookie.Value, HashCsrfState(csrfCookie.Value), csrfState)
		return false, nil
	}
	return true, nil
}

func RetrieveTokenValues(ctx context.Context, request *http.Request) (accessToken string, refreshToken string, err error) {
	if request == nil {
		err = errors.New("nil http request")
		return
	}
	jwtCookie, err := request.Cookie(accessTokenCookie)
	if err != nil || jwtCookie == nil {
		logger.Errorf(ctx, "Could not detect existing access token cookie")
		return
	}
	logger.Debugf(ctx, "Existing JWT cookie found")
	accessToken, err = ReadSecureCookie(ctx, *jwtCookie)
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
	refreshToken, err = ReadSecureCookie(ctx, *refreshCookie)
	if err != nil{
		logger.Errorf(ctx, "Error reading existing secure refresh cookie %s", err)
		return
	}
	return
}
