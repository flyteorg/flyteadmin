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
	"time"
)

func TestClient(t *testing.T) {
	ctx := context.Background()
	endpoint := "localhost:8088"

	var opts []grpc.DialOption

	creds, err := credentials.NewClientTLSFromFile("/Users/ytong/temp/server.pem", ":8088")
	assert.NoError(t, err)
	opts = append(opts, grpc.WithTransportCredentials(creds))

	token := oauth2.Token{
		AccessToken: "eyJraWQiOiItY2FQXzgyX1o0cVVDMnUtWkRPS2pPYVVIa2RkaWN3YUJOMGJjMTZIb3ZFIiwiYWxnIjoiUlMyNTYifQ.eyJ2ZXIiOjEsImp0aSI6IkFULnFmdnZqeTF4cGhlSWJFbkZhYmVmQ1VKRHdIRUdseV9iSy1GV2NTd3RNV3ciLCJpc3MiOiJodHRwczovL2x5ZnQub2t0YS5jb20vb2F1dGgyL2RlZmF1bHQiLCJhdWQiOiJhcGk6Ly9kZWZhdWx0IiwiaWF0IjoxNTY4NzQ2NzgwLCJleHAiOjE1Njg3NTAzODAsImNpZCI6IjBvYWFpMm5ucWIyOVB2TXZIMXQ3IiwidWlkIjoiMDB1MXAyMjB2NmRtc0N3Y28xdDciLCJzY3AiOlsib3BlbmlkIl0sInN1YiI6Inl0b25nQGx5ZnQuY29tIn0.gDVqi23x4q3-lXVnPWue5o7FcMv9Xeaw2MWP9Qh7XQYqonupLMjf9FdTgbNqYdWebU0eUOsjw9Y7N6ku0WEuDOwQEzROBPdrX2XoZd5JF1JQMP35cIuH0vFL9zt30qCyV26B1DCerMW_LKk9utcwOmvl1ij6wUGttDVC1XIh4i0IG2VaKhTaOBTYvdxJb_sVuVKjfXYqW04tPRM9yPErEtK3Mt1yRNgT3mLIxYWB5uNHR4NVBOwYYMZ1zVlDZnGcA2hHKoh2Blju0sRk0NVq8E5zlZCFjVTUO6u3A89_yiDc9WL5FOlI2EQAsO1H1nYlx4tOLGZCmOrCeCXAnc-gmQ",
		Expiry:      time.Unix(1568418660, 0),
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
