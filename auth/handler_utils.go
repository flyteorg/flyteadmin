package auth

import (
	"context"
	"net/http"
	"net/url"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
)

const (
	metadataXForwardedHost = "x-forwarded-host"
	metadataAuthority      = ":authority"
)

//func NewMockOIdCProvider() (*oidc.Provider, error) {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	var issuer string
//	hf := func(w http.ResponseWriter, r *http.Request) {
//		if r.URL.Path == "/.well-known/openid-configuration" {
//			w.Header().Set("Content-Type", "application/json")
//			io.WriteString(w, strings.ReplaceAll(`{
//				"issuer": "ISSUER",
//				"authorization_endpoint": "https://example.com/auth",
//				"token_endpoint": "https://example.com/token",
//				"jwks_uri": "ISSUER/keys",
//				"id_token_signing_alg_values_supported": ["RS256"]
//			}`, "ISSUER", issuer))
//			return
//		} else if r.URL.Path == "/keys" {
//			w.Header().Set("Content-Type", "application/json")
//			io.WriteString(w, `{"keys":[{"kty":"RSA","alg":"RS256","kid":"Z6dmZ_TXhduw-jUBZ6uEEzvnh-jhNO0YhemB7qa_LOc","use":"sig","e":"AQAB","n":"jyMcudBiz7XqeDIvxfMlmG4fvAUU7cl3R4iSIv_ahHanCcVRvqcXOsIknwn7i4rOUjP6MlH45uIYsaj6MuLYgoaIbC-Z823Tu4asoC-rGbpZgf-bMcJLxtZVBNsSagr_M0n8xA1oogHRF1LGRiD93wNr2b9OkKVbWnyNdASk5_xui024nVzakm2-RAEyaC048nHfnjVBvwo4BdJVDgBEK03fbkBCyuaZyE1ZQF545MTbD4keCv58prSCmbDRJgRk48FzaFnQeYTho-pUxXxM9pvhMykeI62WZ7diDfIc9isOpv6ALFOHgKy7Ihhve6pLIylLRTnn2qhHFkGPtU3djQ"}]}`)
//			return
//		}
//
//		http.NotFound(w, r)
//		return
//
//	}
//
//	s := httptest.NewServer(http.HandlerFunc(hf))
//	defer s.Close()
//
//	issuer = s.URL
//	return oidc.NewProvider(ctx, issuer)
//}

func ReconstructRequestURL(req *http.Request) *url.URL {
	if req == nil {
		return nil
	}

	u, _ := url.ParseRequestURI(req.RequestURI)
	if u.IsAbs() {
		return u
	}

	scheme := "http://"
	if req.TLS != nil {
		scheme = "https://"
	}

	u, _ = url.Parse(scheme + req.Host)
	return u
}

func GetPublicURL(ctx context.Context, isSecure bool, httpURLCfg *url.URL) *url.URL {
	if httpURLCfg != nil && len(httpURLCfg.Host) > 0 {
		return httpURLCfg
	}

	md := metautils.ExtractIncoming(ctx)

	forwardedHost := md.Get(metadataXForwardedHost)
	if len(forwardedHost) == 0 {
		forwardedHost = md.Get(metadataAuthority)
	}

	if len(forwardedHost) == 0 {
		return httpURLCfg
	}

	if isSecure {
		forwardedHost = "https://" + forwardedHost
	} else {
		forwardedHost = "http://" + forwardedHost
	}

	u, _ := url.Parse(forwardedHost)
	return u
}
