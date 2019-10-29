package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lyft/flytestdlib/errors"
	"github.com/lyft/flytestdlib/logger"
	"io/ioutil"
	"net/http"
)

const (
	ErrIdpClient errors.ErrorCode = "IDP_REQUEST_FAILED"
)

// Refactor this to be a more complete Okta client and place behind an IDP client interface if usage grows beyond
// just one endpoint and one response object, or integration is needed with another IDP
/*
This is a sample response object returned from Okta
	{
	  "sub": "abc123",
	  "name": "John Smith",
	  "locale": "US",
	  "preferred_username": "jsmith123@company.com",
	  "given_name": "John",
	  "family_name": "Smith",
	  "zoneinfo": "America/Los_Angeles",
	  "updated_at": 1568750854
	}
*/
type OktaUserInfoResponse struct {
	Sub               string `json:"sub"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
}

func postToIdp(ctx context.Context, client *http.Client, userInfoUrl, accessToken string) (OktaUserInfoResponse, error) {
	fmt.Println("+===============")
	fmt.Printf("%s %s %v\n", userInfoUrl, accessToken, client)
	fmt.Println("+===============")

	request, err := http.NewRequest(http.MethodPost, userInfoUrl, nil)
	if err != nil {
		logger.Errorf(ctx, "Error creating user info request to IDP %s", err)
		return OktaUserInfoResponse{}, errors.Wrapf(ErrIdpClient, err, "Error creating user info request to IDP")
	}
	request.Header.Set(DefaultAuthorizationHeader, fmt.Sprintf("%s %s", BearerScheme, accessToken))
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		logger.Errorf(ctx, "Error while requesting user info from IDP %s", err)
		return OktaUserInfoResponse{}, errors.Wrapf(ErrIdpClient, err, "Error while requesting user info from IDP")
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			logger.Errorf(ctx, "Error closing response body %s", err)
		}
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil  {
		logger.Errorf(ctx, "Error reading user info response error %s", response.StatusCode, err)
		return OktaUserInfoResponse{}, errors.Wrapf(ErrIdpClient, err,"Error reading user info response")
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		logger.Errorf(ctx, "Bad response code from IDP %d", response.StatusCode)
		return OktaUserInfoResponse{}, errors.Errorf(ErrIdpClient,
			"Error reading user info response, code %d body %v", response.StatusCode, body)
	}

	responseObject := &OktaUserInfoResponse{}
	err = json.Unmarshal(body, responseObject)
	if err != nil {
		return OktaUserInfoResponse{}, errors.Wrapf(ErrIdpClient, err, "Could not unmarshal IDP response")
	}

	return *responseObject, nil
}
