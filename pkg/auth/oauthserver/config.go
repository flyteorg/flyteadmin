package oauthserver

type Issuer struct {
	Issuer string `json:"issuer"`
	// Cert & key
}

type OAuth2Options struct {
	Issuer Issuer `json:"issuer"`
}
