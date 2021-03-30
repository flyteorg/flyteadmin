package oauthserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-test/deep"
	"github.com/ory/fosite"

	"github.com/stretchr/testify/assert"
)

func TestMarshalAuthorizeRequest(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8088/oauth2/authorize?client_id=my-client&redirect_uri=http%3A%2F%2Flocalhost%3A3846%2Fcallback&response_type=code&scope=photos+openid+offline&state=some-random-state-foobar&nonce=some-random-nonce&code_challenge=p0v_UR0KrXl4--BpxM2BQa7qIW5k3k4WauBhjmkVQw8&code_challenge_method=S256", nil)
	assert.NoError(t, err)

	ctx := context.Background()
	ar, err := oauth2Provider.NewAuthorizeRequest(ctx, req)
	assert.NoError(t, err)

	raw, err := json.Marshal(ar)
	assert.NoError(t, err)

	newAr := &fosite.AuthorizeRequest{}
	err = json.Unmarshal(raw, newAr)
	assert.NoError(t, err)

	if diff := deep.Equal(ar, newAr); diff != nil {
		t.Errorf("Inject() Diff = %v\r\n", diff)
	}
}
