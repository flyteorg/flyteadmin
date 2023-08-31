package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"github.com/flyteorg/flyteadmin/auth/config"
)

func TestCookieManager_SetTokenCookies(t *testing.T) {
	ctx := context.Background()
	// These were generated for unit testing only.
	hashKeyEncoded := "wG4pE1ccdw/pHZ2ml8wrD5VJkOtLPmBpWbKHmezWXktGaFbRoAhXidWs8OpbA3y7N8vyZhz1B1E37+tShWC7gA" //nolint:goconst
	blockKeyEncoded := "afyABVgGOvWJFxVyOvCWCupoTn6BkNl4SOHmahho16Q"                                           //nolint:goconst
	cookieSetting := config.CookieSettings{
		SameSitePolicy: config.SameSiteDefaultMode,
		Domain:         "default",
	}
	manager, err := NewCookieManager(ctx, hashKeyEncoded, blockKeyEncoded, cookieSetting)
	assert.NoError(t, err)
	token := &oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
	}

	t.Run("set_token_cookies", func(t *testing.T) {
		token = token.WithExtra(map[string]interface{}{
			"id_token": "id token",
		})
		w := httptest.NewRecorder()
		_, err = http.NewRequest("GET", "/api/v1/projects", nil)
		assert.NoError(t, err)

		err = manager.SetTokenCookies(ctx, w, token)

		assert.NoError(t, err)
		fmt.Println(w.Header().Get("Set-Cookie"))
		c := w.Result().Cookies()
		assert.Equal(t, "flyte_at", c[0].Name)
		assert.Equal(t, "flyte_idt", c[1].Name)
		assert.Equal(t, "flyte_rt", c[2].Name)
	})

	t.Run("retrieve_token_values", func(t *testing.T) {
		token = token.WithExtra(map[string]interface{}{
			"id_token": "id token",
		})
		w := httptest.NewRecorder()
		_, err = http.NewRequest("GET", "/api/v1/projects", nil)
		assert.NoError(t, err)

		err = manager.SetTokenCookies(ctx, w, token)
		assert.NoError(t, err)

		cookies := w.Result().Cookies()
		req, err := http.NewRequest("GET", "/api/v1/projects", nil)
		assert.NoError(t, err)
		for _, c := range cookies {
			req.AddCookie(c)
		}

		idToken, access, refresh, err := manager.RetrieveTokenValues(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "id token", idToken)
		assert.Equal(t, "access", access)
		assert.Equal(t, "refresh", refresh)
	})

	t.Run("logout_access_cookie", func(t *testing.T) {
		cookie := manager.getLogoutAccessCookie()
		assert.True(t, time.Now().After(cookie.Expires))
	})

	t.Run("logout_refresh_cookie", func(t *testing.T) {
		cookie := manager.getLogoutRefreshCookie()
		assert.True(t, time.Now().After(cookie.Expires))
	})

	t.Run("delete_cookies", func(t *testing.T) {
		w := httptest.NewRecorder()

		manager.DeleteCookies(ctx, w)

		cookies := w.Result().Cookies()
		assert.Equal(t, 2, len(cookies))
		assert.True(t, time.Now().After(cookies[0].Expires))
		assert.True(t, time.Now().After(cookies[1].Expires))
	})

	t.Run("get_http_same_site_policy", func(t *testing.T) {
		manager.sameSitePolicy = config.SameSiteLaxMode
		assert.Equal(t, http.SameSiteLaxMode, manager.getHTTPSameSitePolicy())

		manager.sameSitePolicy = config.SameSiteStrictMode
		assert.Equal(t, http.SameSiteStrictMode, manager.getHTTPSameSitePolicy())

		manager.sameSitePolicy = config.SameSiteNoneMode
		assert.Equal(t, http.SameSiteNoneMode, manager.getHTTPSameSitePolicy())
	})
}
