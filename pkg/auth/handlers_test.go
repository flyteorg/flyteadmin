package auth

import (
	"context"
	"github.com/lyft/flyteadmin/pkg/auth/interfaces/mocks"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"net/http"
	"net/http/httptest"
	"strings"

	"testing"
)

//func TestGetCallbackHandler(t *testing.T) {
//	ctx := context.Background()
//	options
//	authCtx := NewAuthenticationContext(ctx, )
//}

func TestWithUserEmail(t *testing.T) {
	ctx := WithUserEmail(context.Background(), "abc")
	assert.Equal(t, "abc", ctx.Value("email"))
}

func TestGetLoginHandler(t *testing.T) {
	ctx := context.Background()
	dummyOAuth2Config := oauth2.Config{
		ClientID: "abc",
		Scopes: []string{"openid", "other"},
	}
	mockAuthCtx := mocks.AuthenticationContext{}
	mockAuthCtx.On("OAuth2Config").Return(&dummyOAuth2Config)
	handler := GetLoginHandler(ctx, &mockAuthCtx)
	req, err := http.NewRequest("GET", "/login", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, 307, w.Code)
	assert.True(t, strings.Contains(w.Header().Get("Location"), "response_type=code&scope=openid+other"))
	assert.True(t, strings.Contains(w.Header().Get("Set-Cookie"), "my-csrf-state="))
}

func TestGetHttpRequestCookieToMetadataHandler(t *testing.T) {
	ctx := context.Background()
	// These were generated for unit testing only.
	hashKeyEncoded := "wG4pE1ccdw/pHZ2ml8wrD5VJkOtLPmBpWbKHmezWXktGaFbRoAhXidWs8OpbA3y7N8vyZhz1B1E37+tShWC7gA"
	blockKeyEncoded := "afyABVgGOvWJFxVyOvCWCupoTn6BkNl4SOHmahho16Q"
	cookieManager, err := NewCookieManager(ctx, hashKeyEncoded, blockKeyEncoded)
	assert.NoError(t, err)
	mockAuthCtx := mocks.AuthenticationContext{}
	mockAuthCtx.On("CookieManager").Return(&cookieManager)
	handler := GetHttpRequestCookieToMetadataHandler(&mockAuthCtx)
	req, err := http.NewRequest("GET", "/api/v1/projects", nil)
	jwtCookie, err := NewSecureCookie(accessTokenCookie, "a.b.c", cookieManager.hashKey, cookieManager.blockKey)
	assert.NoError(t, err)
	req.AddCookie(&jwtCookie)
	assert.Equal(t, "Bearer a.b.c", handler(ctx, req)["authorization"][0])
}
