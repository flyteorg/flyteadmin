package auth

import "encoding/json"

/*
This struct represents what should be returned by an IDP according to the specification at
 https://openid.net/specs/openid-connect-core-1_0.html#UserInfoResponse

Keep in mind that not all fields are necessarily populated, and additional fields may be present as well. This is a sample
response object returned from Okta for instance
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
type UserInfoResponse struct {
	Sub                  string `json:"sub"`
	NameRaw              string `json:"name"`
	PreferredUsernameRaw string `json:"preferred_username"`
	GivenNameRaw         string `json:"given_name"`
	FamilyNameRaw        string `json:"family_name"`
	EmailRaw             string `json:"email"`
	PictureRaw           string `json:"picture"`
}

func (r UserInfoResponse) MarshalToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r UserInfoResponse) Subject() string {
	return r.Sub
}

func (r UserInfoResponse) Name() string {
	return r.NameRaw
}

func (r UserInfoResponse) PreferredUsername() string {
	return r.PreferredUsernameRaw
}

func (r UserInfoResponse) GivenName() string {
	return r.GivenNameRaw
}

func (r UserInfoResponse) FamilyName() string {
	return r.FamilyNameRaw
}

func (r UserInfoResponse) Email() string {
	return r.EmailRaw
}

func (r UserInfoResponse) Picture() string {
	return r.PictureRaw
}
