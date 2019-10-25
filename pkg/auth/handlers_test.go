package auth

import (
	"context"
	"fmt"
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
	dummyOAuth2Config := oauth2.Config{
		ClientID: "abc",
		Scopes: []string{"openid", "other"},
	}
	mockAuthCtx := mocks.AuthenticationContext{}
	mockAuthCtx.On("OAuth2Config").Return(&dummyOAuth2Config)

	handler := GetHttpRequestCookieToMetadataHandler(&mockAuthCtx)
	fmt.Println(handler(ctx, nil))
}