package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	jwtVerifier := JwtVerifier{
		Url: "https://lyft.okta.com/oauth2/default/v1/keys",
	}
	tokenStr := "eyJraWQiOiItY2FQXzgyX1o0cVVDMnUtWkRPS2pPYVVIa2RkaWN3YUJOMGJjMTZIb3ZFIiwiYWxnIjoiUlMyNTYifQ.eyJ2ZXIiOjEsImp0aSI6IkFULndRZDFab0dxS3RJa3JGY3pmTlRkd1JnZmZZM0VDSjU1VmNPZDVjQmxEMncuQ05mcGRMVlNoem4wekdWZEp0T1kyL080TXcrM0pRc1BxSGFBSms5bkxQRT0iLCJpc3MiOiJodHRwczovL2x5ZnQub2t0YS5jb20vb2F1dGgyL2RlZmF1bHQiLCJhdWQiOiJhcGk6Ly9kZWZhdWx0IiwiaWF0IjoxNTY5MjgyODYzLCJleHAiOjE1NjkyODY0NjMsImNpZCI6IjBvYWJzNTJna3VYNFZ6WXRKMXQ3IiwidWlkIjoiMDB1MXAyMjB2NmRtc0N3Y28xdDciLCJzY3AiOlsib2ZmbGluZV9hY2Nlc3MiLCJvcGVuaWQiXSwic3ViIjoieXRvbmdAbHlmdC5jb20ifQ.pVSnEBcATN6mE9PEa6Ht7Sd-EjzM3Cm6yxR-_pMWzo9DlmVFcV_T8WsNq2pdboblfuZXa-53a0aajrSXbWa-HkHOiE0g0RZaVBZdkp3Wtqj7Xnh9ROhvjLu8V42NcJBpvZlksjAepdoexaVHiIxcxM7lJf49ZiIrvE7L2ybuKYUE5iaZoM47jfizxUcRV-r_GriE0UtUoJRwioGKTm9bpPCA4odM7et42D1zd8bFs32so1cMJcSIHmvDIO5vZMYER29z1mZf-EKr9GrKSrUdjxDncnMspNt1BwwaI-hEf63csYOgxtmcgyTOjNpnIYgi0RB8G8UjDzn7qYyWbg4hhA"
	_, err := jwt.Parse(tokenStr, jwtVerifier.GetKey)
	assert.NotNil(t, err)
	validationErr, ok := err.(*jwt.ValidationError)
	assert.True(t, ok)
	assert.Equal(t, jwt.ValidationErrorExpired, validationErr.Errors)
}
