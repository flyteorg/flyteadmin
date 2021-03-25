package interfaces

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

type CookieHandler interface {
	RetrieveTokenValues(ctx context.Context, request *http.Request) (idToken, accessToken, refreshToken string, err error)
	SetTokenCookies(ctx context.Context, writer http.ResponseWriter, token *oauth2.Token) error
	SetAuthCodeCookie(ctx context.Context, writer http.ResponseWriter, authRequestUrl string) error
	RetrieveAuthCodeRequest(ctx context.Context, request *http.Request) (authRequestUrl string, err error)
	DeleteCookies(ctx context.Context, writer http.ResponseWriter)
}
