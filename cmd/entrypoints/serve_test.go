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
		AccessToken: "eyJraWQiOiItY2FQXzgyX1o0cVVDMnUtWkRPS2pPYVVIa2RkaWN3YUJOMGJjMTZIb3ZFIiwiYWxnIjoiUlMyNTYifQ.eyJ2ZXIiOjEsImp0aSI6IkFULmJMVDdvbXc4ejRnUkxmOTFybFYybjRxRmh2YnVBbzl2bFhWQnZFMDZ2Zjgua3J1eVBmSjc4QnFRazNNWk5VUXVLNmZITVlRekk3OE1rSktrWHpkeDlRST0iLCJpc3MiOiJodHRwczovL2x5ZnQub2t0YS5jb20vb2F1dGgyL2RlZmF1bHQiLCJhdWQiOiJhcGk6Ly9kZWZhdWx0IiwiaWF0IjoxNTY5ODg4NjgxLCJleHAiOjE1Njk4OTIyODEsImNpZCI6IjBvYWJzNTJna3VYNFZ6WXRKMXQ3IiwidWlkIjoiMDB1MXAyMjB2NmRtc0N3Y28xdDciLCJzY3AiOlsib2ZmbGluZV9hY2Nlc3MiLCJvcGVuaWQiXSwic3ViIjoieXRvbmdAbHlmdC5jb20ifQ.JaBGPlGfrO8BCpliurpWAS9MB-DCXakOxx45RoyxGkvSZUbZlxqeGBK5Nuwregp-Qe3dB4j_kW4wYIiXqe6e5NyOz-DyIKFxxRS-ioU8LZ-OojvDyy8zWrGyVHJ2TtG3Ry2Nx5in5uXibqgMO2wf8XOJSEzz0BALyn4FqXD4jPmz0oxqTmIMCpZ4W4pnB8t0aP_IOejFuhmG5QOsZo0p2o62TgcKbjasbRGaKylADz0Cgerx9s1_nX_9gmkma1r3sieE_OGEDoZbSBeGIlWsmOZ07DTtQSGPU_Sf5rCrgL5M1fXRaNps-9AzxRW11r0Dkj1esQeq0ot9PVZn8kgH7g",
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
	fmt.Println(resp)
}
