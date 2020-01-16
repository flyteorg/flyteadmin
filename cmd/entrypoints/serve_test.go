// fdsaxbuild integration

package entrypoints

// This is an integration test because the token will show up as expired, you will need a live token

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)
//
//func TestClient(t *testing.T) {
//	ctx := context.Background()
//	endpoint := "localhost:8088"
//
//	var opts []grpc.DialOption
//
//	creds, err := credentials.NewClientTLSFromFile("/path/to/server.pem", ":8088")
//	assert.NoError(t, err)
//	opts = append(opts, grpc.WithTransportCredentials(creds))
//
//	token := oauth2.Token{
//		AccessToken: "j.w.t",
//	}
//	tokenRpcCredentials := oauth.NewOauthAccess(&token)
//	tokenDialOption := grpc.WithPerRPCCredentials(tokenRpcCredentials)
//	opts = append(opts, tokenDialOption)
//
//	conn, err := grpc.Dial(endpoint, opts...)
//	if err != nil {
//		fmt.Printf("Dial error %v\n", err)
//	}
//	assert.NoError(t, err)
//	client := service.NewAdminServiceClient(conn)
//	resp, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
//	if err != nil {
//		fmt.Printf("Error %v\n", err)
//	}
//	assert.NoError(t, err)
//	fmt.Printf("Response: %v\n", resp)
//}

func GetEncodedCreds() string {
	/*
	def get_basic_authorization_header(client_id, client_secret):
	    """
	    This function transforms the client id and the client secret into a header that conforms with http basic auth.
	    It joins the id and the secret with a : then base64 encodes it, then adds the appropriate text.
	    :param Text client_id:
	    :param Text client_secret:
	    :rtype: Text
	    """
	    concated = "{}:{}".format(client_id, client_secret)
	    return "Basic {}".format(_base64.b64encode(concated.encode(_utf_8)).decode(_utf_8))

	    headers = {
	        'Authorization': authorization_header,
	        'Cache-Control': 'no-cache',
	        'Accept': 'application/json',
	        'Content-Type': 'application/x-www-form-urlencoded'
	    }
	    body = {
	        'grant_type': 'client_credentials',
	    }

	*/
	clientId := "0oacwencpbysS05G01t7"
	clientSecret := "fill_in_password_here"
	concatenated := fmt.Sprintf("%s:%s", clientId, clientSecret)
	encodedString := base64.StdEncoding.EncodeToString([]byte(concatenated))

	return encodedString
}

func TestHttpClient(t *testing.T) {
	oktaEndpoint := "https://lyft.okta.com/oauth2/ausc5wmjw96cRKvTd1t7/v1/token"

	credentials := GetEncodedCreds()
	authHeader := fmt.Sprintf("Basic %s", credentials)

	// These have to be here in order for Okta to respond correctly.
	var urlEncodedBody = []byte("grant_type=client_credentials&scope=svc")

	req, err := http.NewRequest("POST", oktaEndpoint, bytes.NewBuffer(urlEncodedBody))
	assert.NoError(t, err)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	fmt.Println(err)
	assert.NoError(t, err)

	accessToken, ok := data["access_token"].(string)
	assert.True(t, ok)
	fmt.Printf("token %s\n", accessToken)

	// Make Flyte request
	req2, err := http.NewRequest("GET", "https://flyte-staging.lyft.net/api/v1/projects", nil)
	assert.NoError(t, err)
	req2.Header.Add("flyte-authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp2, err := client.Do(req2)
	assert.NoError(t, err)

	f, err := ioutil.ReadAll(resp2.Body)
	assert.NoError(t, err)
	fmt.Println(string(f))
}

