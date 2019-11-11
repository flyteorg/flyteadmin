package auth

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCookieManager_SetTokenCookies(t *testing.T) {
	ctx := context.Background()
	// These were generated for unit testing only.
	hashKeyEncoded := "wG4pE1ccdw/pHZ2ml8wrD5VJkOtLPmBpWbKHmezWXktGaFbRoAhXidWs8OpbA3y7N8vyZhz1B1E37+tShWC7gA"
	blockKeyEncoded := "afyABVgGOvWJFxVyOvCWCupoTn6BkNl4SOHmahho16Q"

	manager, err := NewCookieManager(ctx, hashKeyEncoded, blockKeyEncoded)
	assert.NoError(t, err)

	token := oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
	}

	w := httptest.NewRecorder()
	err = manager.SetTokenCookies(ctx, w, &token)
	assert.NoError(t, err)
	fmt.Println(w.Header().Get("Set-Cookie"))
	c := w.Result().Cookies()
	assert.Equal(t, "flyte_jwt", c[0].Name)
	assert.Equal(t, "flyte_refresh", c[1].Name)
}

func TestCookieManager_RetrieveTokenValues(t *testing.T) {
	ctx := context.Background()
	// These were generated for unit testing only.
	hashKeyEncoded := "wG4pE1ccdw/pHZ2ml8wrD5VJkOtLPmBpWbKHmezWXktGaFbRoAhXidWs8OpbA3y7N8vyZhz1B1E37+tShWC7gA"
	blockKeyEncoded := "afyABVgGOvWJFxVyOvCWCupoTn6BkNl4SOHmahho16Q"

	manager, err := NewCookieManager(ctx, hashKeyEncoded, blockKeyEncoded)
	assert.NoError(t, err)

	token := oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
	}

	w := httptest.NewRecorder()
	err = manager.SetTokenCookies(ctx, w, &token)

	cookies := w.Result().Cookies()
	req, err := http.NewRequest("GET", "/api/v1/projects", nil)
	assert.NoError(t, err)
	req.AddCookie(cookies[0])
	req.AddCookie(cookies[1])

	access, refresh, err := manager.RetrieveTokenValues(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, "access", access)
	assert.Equal(t, "refresh", refresh)
}
