package auth

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

//func TestParse(t *testing.T) {
//	jwtVerifier := JwtVerifier{
//		Url: "https://lyft.okta.com/oauth2/default/v1/keys",
//	}
//	tokenStr := "eyJraWQiOiItY2FQXzgyX1o0cVVDMnUtWkRPS2pPYVVIa2RkaWN3YUJOMGJjMTZIb3ZFIiwiYWxnIjoiUlMyNTYifQ.eyJ2ZXIiOjEsImp0aSI6IkFULndRZDFab0dxS3RJa3JGY3pmTlRkd1JnZmZZM0VDSjU1VmNPZDVjQmxEMncuQ05mcGRMVlNoem4wekdWZEp0T1kyL080TXcrM0pRc1BxSGFBSms5bkxQRT0iLCJpc3MiOiJodHRwczovL2x5ZnQub2t0YS5jb20vb2F1dGgyL2RlZmF1bHQiLCJhdWQiOiJhcGk6Ly9kZWZhdWx0IiwiaWF0IjoxNTY5MjgyODYzLCJleHAiOjE1NjkyODY0NjMsImNpZCI6IjBvYWJzNTJna3VYNFZ6WXRKMXQ3IiwidWlkIjoiMDB1MXAyMjB2NmRtc0N3Y28xdDciLCJzY3AiOlsib2ZmbGluZV9hY2Nlc3MiLCJvcGVuaWQiXSwic3ViIjoieXRvbmdAbHlmdC5jb20ifQ.pVSnEBcATN6mE9PEa6Ht7Sd-EjzM3Cm6yxR-_pMWzo9DlmVFcV_T8WsNq2pdboblfuZXa-53a0aajrSXbWa-HkHOiE0g0RZaVBZdkp3Wtqj7Xnh9ROhvjLu8V42NcJBpvZlksjAepdoexaVHiIxcxM7lJf49ZiIrvE7L2ybuKYUE5iaZoM47jfizxUcRV-r_GriE0UtUoJRwioGKTm9bpPCA4odM7et42D1zd8bFs32so1cMJcSIHmvDIO5vZMYER29z1mZf-EKr9GrKSrUdjxDncnMspNt1BwwaI-hEf63csYOgxtmcgyTOjNpnIYgi0RB8G8UjDzn7qYyWbg4hhA"
//	_, err := jwt.Parse(tokenStr, jwtVerifier.GetKey)
//	assert.NotNil(t, err)
//	validationErr, ok := err.(*jwt.ValidationError)
//	assert.True(t, ok)
//	assert.Equal(t, jwt.ValidationErrorExpired, validationErr.Errors)
//}

func TestOidc(t *testing.T) {
	//tokenStr := "eyJraWQiOiItY2FQXzgyX1o0cVVDMnUtWkRPS2pPYVVIa2RkaWN3YUJOMGJjMTZIb3ZFIiwiYWxnIjoiUlMyNTYifQ.eyJ2ZXIiOjEsImp0aSI6IkFULmRObmh1WTZPMm80QU1sOGNmYVRTWkNqQl9jNW1hYUNZUFZVdzRDZzJOTGcuWlFxMG5Tb0ZDTWl6RUlRUTNlL3gyaXoyY1EwNDdJemgrTDE5UTQ2Q3U4OD0iLCJpc3MiOiJodHRwczovL2x5ZnQub2t0YS5jb20vb2F1dGgyL2RlZmF1bHQiLCJhdWQiOiJhcGk6Ly9kZWZhdWx0IiwiaWF0IjoxNTcwNDAxNTU3LCJleHAiOjE1NzA0MDUxNTcsImNpZCI6IjBvYWJzNTJna3VYNFZ6WXRKMXQ3IiwidWlkIjoiMDB1MXAyMjB2NmRtc0N3Y28xdDciLCJzY3AiOlsib3BlbmlkIiwib2ZmbGluZV9hY2Nlc3MiXSwic3ViIjoieXRvbmdAbHlmdC5jb20ifQ.iTAwHhWZgdu-k6am8QH85Kt2l5duEiKuex5d2l0uAhppvVBKmyL1n2r55xxea5Q9AtQo_dFtUnceqmi6XVlksppZ__NAAG8v6vHzrUpin1G2XW4Ycv2_HMNSCAZCVQ_JCGIX9INxTb1K43sknZo0eMpEaf4Do24MtkJxQEZpCrPyt_qsMVxyRJBIsUcW3wzd8mMyvn5rX-EWIvDuaQYwrW2egCdw6I2IzWjZ923F9S-neHKkuf3E4fVnp8WUYoWs3BRR9LzZQPwDCES10sixLb7v6khopP4bW2L5yooccDp0Sdied08aFw63Uu4vzFRwIu_vEzEhSaDY7KOgccJjcA"
	tokenStr := "eyJraWQiOiItY2FQXzgyX1o0cVVDMnUtWkRPS2pPYVVIa2RkaWN3YUJOMGJjMTZIb3ZFIiwiYWxnIjoiUlMyNTYifQ.eyJ2ZXIiOjEsImp0aSI6IkFULnA1WUFlbzBodkhVODdBWGRqTVBDNXRrUG9fdHJBMnk3Z2hLQUMxQ1ktZWcucFNRZGJTR3RUK3hUV1Jza05ZaVd1L2FPSmEvUjd4OTRlbmpWb3dGL2c2QT0iLCJpc3MiOiJodHRwczovL2x5ZnQub2t0YS5jb20vb2F1dGgyL2RlZmF1bHQiLCJhdWQiOiJhcGk6Ly9kZWZhdWx0IiwiaWF0IjoxNTcwNDA2NjU4LCJleHAiOjE1NzA0MTAyNTgsImNpZCI6IjBvYWJzNTJna3VYNFZ6WXRKMXQ3IiwidWlkIjoiMDB1MXAyMjB2NmRtc0N3Y28xdDciLCJzY3AiOlsib2ZmbGluZV9hY2Nlc3MiLCJvcGVuaWQiXSwic3ViIjoieXRvbmdAbHlmdC5jb20ifQ.VJtWuP4vZN50SMJ6VT4kU7u6YfthgTTSs4KPXj80KpXA5VLt9mXVmlGOoVEVUvJqmgZuh5xqHU0epJw7fgJCH7fBvqVaxmfWleaCJKeZ2lo9Ly5kE_RbxYGrhDu285FOmJzZd_p8zJ1KJRNy5zXbkMCOuRXSqFQol2INXcy7auyxFe4_pN6dMpsz5mbKghBsPPsyRmWBnpuWMIz8HmVZXy6CVh4ik2vSIFVdn9R6ulysSSxwgfZYNHi8NUn_Wbf4lk1VDKNlatgKpwDD4vRYzzE5vu0JdVNKt_TgKc7_VR4ZxUrf82lztHSNmAsnGg5hcCHGWKQKHtcprsI3OSYZDA"

	myClient := &http.Client{}
	ctx := oidc.ClientContext(context.Background(), myClient)

	//keySet := oidc.NewRemoteKeySet(ctx, "https://lyft.okta.com/oauth2/default/v1/keys")

	provider, err := oidc.NewProvider(ctx, "https://lyft.okta.com/oauth2/default")
	if err != nil {
		panic(err)
	}
	var verifier = provider.Verifier(&oidc.Config{ClientID: "api://default"})

	x, err := verifier.Verify(context.Background(), tokenStr)
	assert.NoError(t, err)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

// Until the go-oidc library uses typed errors, we are left to check expiration with a string match.
// This test ensures that.
func TestExpiredToken(t *testing.T) {
	expiredToken := "eyJraWQiOiItY2FQXzgyX1o0cVVDMnUtWkRPS2pPYVVIa2RkaWN3YUJOMGJjMTZIb3ZFIiwiYWxnIjoiUlMyNTYifQ.eyJ2ZXIiOjEsImp0aSI6IkFULmRObmh1WTZPMm80QU1sOGNmYVRTWkNqQl9jNW1hYUNZUFZVdzRDZzJOTGcuWlFxMG5Tb0ZDTWl6RUlRUTNlL3gyaXoyY1EwNDdJemgrTDE5UTQ2Q3U4OD0iLCJpc3MiOiJodHRwczovL2x5ZnQub2t0YS5jb20vb2F1dGgyL2RlZmF1bHQiLCJhdWQiOiJhcGk6Ly9kZWZhdWx0IiwiaWF0IjoxNTcwNDAxNTU3LCJleHAiOjE1NzA0MDUxNTcsImNpZCI6IjBvYWJzNTJna3VYNFZ6WXRKMXQ3IiwidWlkIjoiMDB1MXAyMjB2NmRtc0N3Y28xdDciLCJzY3AiOlsib3BlbmlkIiwib2ZmbGluZV9hY2Nlc3MiXSwic3ViIjoieXRvbmdAbHlmdC5jb20ifQ.iTAwHhWZgdu-k6am8QH85Kt2l5duEiKuex5d2l0uAhppvVBKmyL1n2r55xxea5Q9AtQo_dFtUnceqmi6XVlksppZ__NAAG8v6vHzrUpin1G2XW4Ycv2_HMNSCAZCVQ_JCGIX9INxTb1K43sknZo0eMpEaf4Do24MtkJxQEZpCrPyt_qsMVxyRJBIsUcW3wzd8mMyvn5rX-EWIvDuaQYwrW2egCdw6I2IzWjZ923F9S-neHKkuf3E4fVnp8WUYoWs3BRR9LzZQPwDCES10sixLb7v6khopP4bW2L5yooccDp0Sdied08aFw63Uu4vzFRwIu_vEzEhSaDY7KOgccJjcA"

	myClient := &http.Client{}
	ctx := oidc.ClientContext(context.Background(), myClient)
	provider, err := oidc.NewProvider(ctx, "https://lyft.okta.com/oauth2/default")
	if err != nil {
		panic(err)
	}
	var verifier = provider.Verifier(&oidc.Config{ClientID: "api://default"})

	x, err := verifier.Verify(context.Background(), expiredToken)
	assert.Nil(t, x)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "token is expired"))
}
