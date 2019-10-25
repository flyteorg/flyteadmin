package interfaces

import (
	"context"
	"golang.org/x/oauth2"
	"net/http"
)

type CookieHandler interface {
	RetrieveTokenValues(ctx context.Context, request *http.Request) (accessToken string, refreshToken string, err error)
	SetTokenCookies(ctx context.Context, writer http.ResponseWriter, token *oauth2.Token) error
}
