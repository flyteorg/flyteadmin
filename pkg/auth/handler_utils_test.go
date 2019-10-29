package auth

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type RoundTripFunc func(request *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestHttpClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestPostToIdp(t *testing.T) {
	ctx := context.Background()
	//token := "j.w.t"
	token := "eyJraWQiOiJNbHdnRHRxczdUdmlCOWFZZzlfQ01yVWRYTjhVVUZGdEQxaFlQWXZFaHAwIiwiYWxnIjoiUlMyNTYifQ.eyJ2ZXIiOjEsImp0aSI6IkFULnJwMmtkd2N6T2xvMWFScXNxdXNIY1JHM2JMZ2JyYnV4TTUxOHRQZ0hMckkuMW9jTjVLbzNjb0JwWU1jcXl5aGppTCtHeGQvTFlLcGZLUktQM3EwNi9kND0iLCJpc3MiOiJodHRwczovL2x5ZnQub2t0YS5jb20vb2F1dGgyL2F1c2M1d21qdzk2Y1JLdlRkMXQ3IiwiYXVkIjoiaHR0cHM6Ly9mbHl0ZS5seWZ0Lm5ldCIsImlhdCI6MTU3MjM3NTI4NSwiZXhwIjoxNTcyMzc4ODg1LCJjaWQiOiIwb2FiczUyZ2t1WDRWell0SjF0NyIsInVpZCI6IjAwdTFwMjIwdjZkbXNDd2NvMXQ3Iiwic2NwIjpbInByb2ZpbGUiLCJvZmZsaW5lX2FjY2VzcyIsIm9wZW5pZCJdLCJzdWIiOiJ5dG9uZ0BseWZ0LmNvbSJ9.PAzpp8skNWYAP5YoLj9wtgAOFwDAaveW-SiOIIRLxw_zlvvqDcWRZscmzRqxqeFnSz5XClyUQAOnvuNLl1SXH8Rrd5hhYZB1aLeb0p0aVtbmbNxn8GV96PvashLfCQKhv5dTRyjEPunn9AsMWrnyVwUKwMg-4kh-5xmDVYiMvI_GhWnxHKWbB-Ar-Rv0D1W5wlXlWZwAN0j1f8_U7cTcaElMtsdCUVDAAI71_mUddSoBIVcMOhnNm1jTRjW8h6hUSB9wTqflfKuKn7xEFW5MeJ7NrgGSrRG6s_o3m0lZUVNBtnsgPnrn67a7ksaCeAsOhAZ3vEr3o1N368TuT071pg"

	// Use a real client and a real token to run a live test
	client := &http.Client{}
	//responseObj := &OktaUserInfoResponse{
	//	Sub:               "abc123",
	//	Name:              "John Smith",
	//	PreferredUsername: "jsmith@company.com",
	//	GivenName:         "John",
	//	FamilyName:        "Smith",
	//}
	//responseBytes, err := json.Marshal(responseObj)
	//assert.NoError(t, err)
	//client := NewTestHttpClient(func(request *http.Request) *http.Response {
	//	return &http.Response{
	//		StatusCode: 200,
	//		Body:       ioutil.NopCloser(bytes.NewReader(responseBytes)),
	//		Header:     make(http.Header),
	//	}
	//})

	obj, err := postToIdp(ctx, client, "https://lyft.okta.com/oauth2/ausc5wmjw96cRKvTd1t7/v1/userinfo", token)
	assert.NoError(t, err)
	fmt.Println(obj)
	//assert.Equal(t, responseObj.Name, obj.Name)
	//assert.Equal(t, responseObj.Sub, obj.Sub)
	//assert.Equal(t, responseObj.PreferredUsername, obj.PreferredUsername)
	//assert.Equal(t, responseObj.GivenName, obj.GivenName)
	//assert.Equal(t, responseObj.FamilyName, obj.FamilyName)
}
