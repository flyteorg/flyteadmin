package authzserver

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	config2 "github.com/flyteorg/flytestdlib/config"

	"github.com/flyteorg/flyteadmin/pkg/auth"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyteadmin/pkg/auth/interfaces/mocks"

	"github.com/flyteorg/flyteadmin/pkg/auth/config"

	"github.com/ory/fosite"

	"github.com/stretchr/testify/assert"
)

func TestAuthEndpoint(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		originalURL := "http://localhost:8088/oauth2/authorize?client_id=my-client&redirect_uri=http%3A%2F%2Flocalhost%3A3846%2Fcallback&response_type=code&scope=photos+openid+offline&state=some-random-state-foobar&nonce=some-random-nonce&code_challenge=p0v_UR0KrXl4--BpxM2BQa7qIW5k3k4WauBhjmkVQw8&code_challenge_method=S256"
		req := httptest.NewRequest(http.MethodGet, originalURL, nil)
		w := httptest.NewRecorder()

		authCtx := &mocks.AuthenticationContext{}
		oauth2Provider := &mocks.OAuth2Provider{}
		oauth2Provider.OnNewAuthorizeRequest(req.Context(), req).Return(fosite.NewAuthorizeRequest(), nil)
		authCtx.OnOAuth2Provider().Return(oauth2Provider)

		cookieManager := &mocks.CookieHandler{}
		cookieManager.OnSetAuthCodeCookie(req.Context(), w, originalURL).Return(nil)
		authCtx.OnCookieManager().Return(cookieManager)

		authEndpoint(authCtx, w, req)
		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	})

	t.Run("Fail to write cookie", func(t *testing.T) {
		originalURL := "http://localhost:8088/oauth2/authorize?client_id=my-client&redirect_uri=http%3A%2F%2Flocalhost%3A3846%2Fcallback&response_type=code&scope=photos+openid+offline&state=some-random-state-foobar&nonce=some-random-nonce&code_challenge=p0v_UR0KrXl4--BpxM2BQa7qIW5k3k4WauBhjmkVQw8&code_challenge_method=S256"
		req := httptest.NewRequest(http.MethodGet, originalURL, nil)
		w := httptest.NewRecorder()

		authCtx := &mocks.AuthenticationContext{}
		oauth2Provider := &mocks.OAuth2Provider{}
		requester := fosite.NewAuthorizeRequest()
		oauth2Provider.OnNewAuthorizeRequest(req.Context(), req).Return(requester, nil)
		oauth2Provider.On("WriteAuthorizeError", w, requester, mock.Anything).Run(func(args mock.Arguments) {
			rw := args.Get(0).(http.ResponseWriter)
			rw.WriteHeader(http.StatusForbidden)
		})
		authCtx.OnOAuth2Provider().Return(oauth2Provider)

		cookieManager := &mocks.CookieHandler{}
		cookieManager.OnSetAuthCodeCookie(req.Context(), w, originalURL).Return(fmt.Errorf("failure injection"))
		authCtx.OnCookieManager().Return(cookieManager)

		authEndpoint(authCtx, w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

const sampleIDToken = `eyJraWQiOiJ6emxaT1p0TmY3dzVJazlSaTRuWVNfaGEyNnhSdlk3aWVkblNWR3liMHZVIiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiIwMHVra2k0OHBzSDhMaWtZVjVkNiIsIm5hbWUiOiJIYXl0aGFtIEFidWVsZnV0dWgiLCJ2ZXIiOjEsImlzcyI6Imh0dHBzOi8vZGV2LTE0MTg2NDIyLm9rdGEuY29tIiwiYXVkIjoiMG9ha2toZXRlTmpDTUVSc3Q1ZDYiLCJpYXQiOjE2MTgyMzU3NTMsImV4cCI6MTYxODIzOTM1MywianRpIjoiSUQudXQ5d0JzeWs1dDNZTnp2RnY4bXdWend0OTVCM1hLSGhUUTdjc0p6MmpMNCIsImFtciI6WyJwd2QiXSwiaWRwIjoiMG9ha2ttMWNpMVVlUHBOVTA1ZDYiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJoYXl0aGFtQHVuaW9uLmFpIiwiYXV0aF90aW1lIjoxNjE4MjM1NzI2LCJhdF9oYXNoIjoiTVlhYlpSdVZsS2R6UGtIQ1B2VDQxZyJ9.I60TFUVJjegx8c1XWxtJUmmcGO3XL55kAjSGWNBYQdIn2xnqvboZs_2rb433dM1dLk2rBuesks9VlGttt_Qf4QSflQrFFUcv4Ky_A9Y_VYSFnDzoCOHah1bIl8uYrwIVXtdDM3S97dpENytGvN-o54d8gJcoK94gqk4L_5-NJuhwwkzAre9QE60ws8WkSxOYCzLKl-qA1QsuO3L_bdiQIWa8C7GSyg-yfIsVjtvlUCuwOquOnGaTY-u7NVzlRIJXI1dmKRtuwSYW2OwvhfkcEJDWsq5wpP6Ac6lKAmrV8lQ5yN_8JBqbMQg80jpXJZA7iuqZg1IUnjKvDm82MKjofA`

func TestAuthCallbackEndpoint(t *testing.T) {
	originalURL := "http://localhost:8088/oauth2/authorize?client_id=my-client&redirect_uri=http%3A%2F%2Flocalhost%3A3846%2Fcallback&response_type=code&scope=photos+openid+offline&state=some-random-state-foobar&nonce=some-random-nonce&code_challenge=p0v_UR0KrXl4--BpxM2BQa7qIW5k3k4WauBhjmkVQw8&code_challenge_method=S256"
	req := httptest.NewRequest(http.MethodGet, originalURL, nil)
	w := httptest.NewRecorder()

	authCtx := &mocks.AuthenticationContext{}

	oauth2Provider := &mocks.OAuth2Provider{}
	requester := fosite.NewAuthorizeRequest()
	oauth2Provider.OnNewAuthorizeRequest(req.Context(), req).Return(requester, nil)
	oauth2Provider.On("WriteAuthorizeError", w, requester, mock.Anything).Run(func(args mock.Arguments) {
		rw := args.Get(0).(http.ResponseWriter)
		rw.WriteHeader(http.StatusForbidden)
	})

	authCtx.OnOAuth2Provider().Return(oauth2Provider)

	cookieManager := &mocks.CookieHandler{}
	cookieManager.OnSetAuthCodeCookie(req.Context(), w, originalURL).Return(nil)
	cookieManager.OnRetrieveTokenValues(req.Context(), req).Return(sampleIDToken, "", "", nil)
	cookieManager.OnRetrieveUserInfo(req.Context(), req).Return(&service.UserInfoResponse{Subject: "abc"}, nil)
	authCtx.OnCookieManager().Return(cookieManager)

	authCtx.OnOptions().Return(&config.Config{})

	mockOidcProvider, err := auth.NewMockOIdCProvider()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	authCtx.OnOidcProvider().Return(mockOidcProvider)

	authCallbackEndpoint(authCtx, w, req)
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)

}

func TestGetIssuer(t *testing.T) {
	t.Run("SelfAuthServerIssuer wins", func(t *testing.T) {
		issuer := GetIssuer(&config.Config{
			AppAuth: config.OAuth2Options{
				SelfAuthServer: config.AuthorizationServer{
					Issuer: "my_issuer",
				},
			},
			HTTPPublicUri: config2.URL{URL: *mustParseURL("http://localhost/")},
		})

		assert.Equal(t, "my_issuer", issuer)
	})

	t.Run("Fallback to http public uri", func(t *testing.T) {
		issuer := GetIssuer(&config.Config{
			HTTPPublicUri: config2.URL{URL: *mustParseURL("http://localhost/")},
		})

		assert.Equal(t, "http://localhost/", issuer)
	})
}

func TestEncryptDecrypt(t *testing.T) {
	cookieHashKey := [auth.SymmetricKeyLength]byte{}
	_, err := rand.Read(cookieHashKey[:])
	assert.NoError(t, err)

	input := "hello world"
	encrypted, err := encryptString(input, cookieHashKey)
	assert.NoError(t, err)

	decrypted, err := decryptString(encrypted, cookieHashKey)
	assert.NoError(t, err)

	assert.Equal(t, input, decrypted)
	assert.NotEqual(t, input, encrypted)
}
