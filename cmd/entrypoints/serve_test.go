// +build integration

package entrypoints
// This is an integration test because the token will show up as expired, you will need a live token

import (
	"context"
	"fmt"
	"github.com/grpc/grpc-go/credentials/oauth"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"testing"
)

func TestClient(t *testing.T) {
	ctx := context.Background()
	endpoint := "localhost:8088"

	var opts []grpc.DialOption

	creds, err := credentials.NewClientTLSFromFile("/Users/ytong/temp/server.pem", ":8088")
	assert.NoError(t, err)
	opts = append(opts, grpc.WithTransportCredentials(creds))

	token := oauth2.Token{
		AccessToken: "eyJraWQiOiItY2FQXzgyX1o0cVVDMnUtWkRPS2pPYVVIa2RkaWN3YUJOMGJjMTZIb3ZFIiwiYWxnIjoiUlMyNTYifQ.eyJ2ZXIiOjEsImp0aSI6IkFULmVlSjVhU2x4M0VvdUNiYVM5ZEtBeVF6YXUwYUVEY2NxTFhvY3lNbmFleHcuNHNXSnFWOXNFQy80cXRrQ00vTy9tT2hEYUsvSFc4T0JORG1rT2hxVDlIOD0iLCJpc3MiOiJodHRwczovL2x5ZnQub2t0YS5jb20vb2F1dGgyL2RlZmF1bHQiLCJhdWQiOiJhcGk6Ly9kZWZhdWx0IiwiaWF0IjoxNTcxNzgxMjYzLCJleHAiOjE1NzE3ODQ4NjMsImNpZCI6IjBvYWJzNTJna3VYNFZ6WXRKMXQ3IiwidWlkIjoiMDB1MXAyMjB2NmRtc0N3Y28xdDciLCJzY3AiOlsib2ZmbGluZV9hY2Nlc3MiLCJvcGVuaWQiXSwic3ViIjoieXRvbmdAbHlmdC5jb20ifQ.UdUTprvrjiOWZ3l7J2rxgeSOXhvhtPc3kD06NV0YQAfpoJKI-MStvZHfOwvFmpgTbrsNrxqpHCFE-cawJU1FlWP0YfXIoeHN94PZCr0YRGNKQwwCBLXr1GtcViM_crO9EBvH3Nl5cXA8sQzjNrCjs3KbF30eAH2ZSl0sgSHC6d4hbMEKhyIFmmxHnnj0HlbE3Tk_VUmUPC5b3LrKhse3mEtnOBCBmUGEIZxLPDImVt0PxWApXSyaUsOOagkQWR0qR_4HbgOQmRizA-ctNI50yhNkR1qY_UN22uNJqEWhe7vT13R3LQLg5-uTMkWAKy9KGKVk1StpvhIAIqQkNcDv9g",
	}
	tokenRpcCredentials := oauth.NewOauthAccess(&token)
	tokenDialOption := grpc.WithPerRPCCredentials(tokenRpcCredentials)
	opts = append(opts, tokenDialOption)

	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		fmt.Printf("Dial error %v\n", err)
	}
	assert.NoError(t, err)
	client := service.NewAdminServiceClient(conn)
	resp, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
	assert.NoError(t, err)
	fmt.Printf("Response: %v\n", resp)
}
