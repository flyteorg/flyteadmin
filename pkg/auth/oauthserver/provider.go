package oauthserver

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/ory/x/jwtx"

	"github.com/flyteorg/flyteadmin/pkg/auth/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/auth"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"

	jwtgo "github.com/dgrijalva/jwt-go"
	fositeOAuth2 "github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/token/jwt"

	"github.com/flyteorg/flyteadmin/pkg/auth/config"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/storage"
)

const (
	ClientIDClaim = "client_id"
	UserIDClaim   = "user_info"
)

type Provider struct {
	fosite.OAuth2Provider
	cfg       config.OAuth2Options
	publicKey []rsa.PublicKey
}

func (p Provider) PublicKeys() []rsa.PublicKey {
	return p.publicKey
}

// A session is passed from the `/auth` to the `/token` endpoint. You probably want to store data like: "Who made the request",
// "What organization does that person belong to" and so on.
// For our use case, the session will meet the requirements imposed by JWT access tokens, HMAC access tokens and OpenID Connect
// ID Tokens plus a custom field

// NewJWTSessionToken is a helper function for creating a new session. This may look like a lot of code but since we are
// setting up multiple strategies it is a bit longer.
// Usually, you could do:
//
//  session = new(fosite.DefaultSession)
func (p Provider) NewJWTSessionToken(subject string, userInfoClaims interface{}, appID, issuer, audience string) *fositeOAuth2.JWTSession {
	return &fositeOAuth2.JWTSession{
		JWTClaims: &jwt.JWTClaims{
			Audience:  []string{audience},
			Issuer:    issuer,
			Subject:   subject,
			ExpiresAt: time.Now().Add(p.cfg.AccessTokenLifespan.Duration),
			IssuedAt:  time.Now(),
			Extra: map[string]interface{}{
				ClientIDClaim: appID,
				UserIDClaim:   userInfoClaims,
			},
		},
		JWTHeader: &jwt.Headers{
			Extra: make(map[string]interface{}),
		},
	}
}

func (p Provider) ValidateAccessToken(_ context.Context, tokenStr string) (interfaces.IdentityContext, error) {
	// Parse and validate the token.
	parsedToken, err := jwtgo.Parse(tokenStr, func(t *jwtgo.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtgo.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return &p.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("parsed token is invalid")
	}

	claimsRaw := parsedToken.Claims.(jwtgo.MapClaims)
	claims := jwtx.ParseMapStringInterfaceClaims(claimsRaw)
	//claims := parsedToken.Claims.(jwtgo.MapClaims)
	if len(claims.Audience) != 1 {
		return nil, fmt.Errorf("expected exactly one granted audience. found [%v]", len(claims.Audience))
	}

	// TODO: Add lifespan check
	// TODO: Add audience validation

	userInfoRaw := claimsRaw[UserIDClaim].(map[string]interface{})
	raw, err := json.Marshal(userInfoRaw)
	if err != nil {
		return nil, err
	}

	userInfo := auth.UserInfoResponse{}
	if err = json.Unmarshal(raw, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info claim into UserInfo type. Error: %w", err)
	}

	return auth.NewIdentityContext(claims.Audience[0], claims.Subject, claimsRaw[ClientIDClaim].(string),
		claims.IssuedAt, userInfo), nil
}

func toClientIface(clients map[string]*fosite.DefaultClient) map[string]fosite.Client {
	res := make(map[string]fosite.Client, len(clients))
	for clientID, client := range clients {
		res[clientID] = client
	}

	return res
}

// Creates a new OAuth2 Provider that is able to do OAuth 2-legged and 3-legged flows.
// It'll lookup auth.SecretTokenHash and auth.SecretTokenSigningRSAKey secrets from the secret manager to use to sign
// and generate hashes for tokens. The RSA Private key is expected to be in PEM format with the public key embedded.
// Use auth.GetInitSecretsCommand() to generate new valid secrets that will be accepted by this provider.
// The auth.SecretTokenHash must be a 32-bytes long key in Base64Encoding.
func NewProvider(ctx context.Context, cfg *config.Config, sm core.SecretManager) (Provider, error) {
	// fosite requires four parameters for the server to get up and running:
	// 1. config - for any enforcement you may desire, you can do this using `compose.Config`. You like PKCE, enforce it!
	// 2. store - no auth service is generally useful unless it can remember clients and users.
	//    fosite is incredibly composable, and the store parameter enables you to build and BYODb (Bring Your Own Database)
	// 3. secret - required for code, access and refresh token generation.
	// 4. privateKey - required for id/jwt token generation.

	// Check the api documentation of `compose.Config` for further configuration options.
	composeConfig := &compose.Config{
		AccessTokenLifespan: cfg.AppAuth.AccessTokenLifespan.Duration,
	}

	// Build an in-memory store with static clients defined in Config. This gives us the potential to move the clients
	// storage into DB and allow registration of new clients to users.
	store := &storage.MemoryStore{
		IDSessions:             make(map[string]fosite.Requester),
		Clients:                toClientIface(cfg.AppAuth.StaticClients),
		AuthorizeCodes:         map[string]storage.StoreAuthorizeCode{},
		AccessTokens:           map[string]fosite.Requester{},
		RefreshTokens:          map[string]storage.StoreRefreshToken{},
		PKCES:                  map[string]fosite.Requester{},
		AccessTokenRequestIDs:  map[string]string{},
		RefreshTokenRequestIDs: map[string]string{},
		IssuerPublicKeys:       map[string]storage.IssuerPublicKeys{},
	}

	// This secret is used to sign authorize codes, access and refresh tokens.
	// It has to be 32-bytes long for HMAC signing. This requirement can be configured via `compose.Config` above.
	// If you require this to key to be stable, for example, when running multiple fosite servers, you can generate the
	// 32byte random key as above and push it out to a base64 encoded string.
	// This can then be injected and decoded as the `var secret []byte` on server start.
	tokenHashBase64, err := sm.Get(ctx, cfg.AppAuth.TokenHashKeySecretName)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to read secretTokenHash file. Error: %w", err)
	}

	secret, err := base64.RawStdEncoding.DecodeString(tokenHashBase64)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to decode token hash using base64 encoding. Error: %w", err)
	}

	// privateKey is used to sign JWT tokens. The default strategy uses RS256 (RSA Signature with SHA-256)
	privateKeyPEM, err := sm.Get(ctx, cfg.AppAuth.TokenSigningRSAKeySecretName)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to read token signing RSA Key. Error: %w", err)
	}

	block, _ := pem.Decode([]byte(privateKeyPEM))
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to parse PKCS1PrivateKey. Error: %w", err)
	}

	// Build a fosite instance with all OAuth2 and OpenID Connect handlers enabled, plugging in our configurations as specified above.
	oauth2Provider := composeOAuth2Provider(composeConfig, store, secret, privateKey)

	publicKeys := []rsa.PublicKey{privateKey.PublicKey}

	// Try to load old key to validate tokens using it to support key rotation.
	privateKeyPEM, err = sm.Get(ctx, cfg.AppAuth.OldTokenSigningRSAKeySecretName)
	if err == nil {
		block, _ = pem.Decode([]byte(privateKeyPEM))
		oldPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return Provider{}, fmt.Errorf("failed to parse PKCS1PrivateKey. Error: %w", err)
		}

		publicKeys = append(publicKeys, oldPrivateKey.PublicKey)
	}

	return Provider{
		OAuth2Provider: oauth2Provider,
		publicKey:      publicKeys,
	}, nil
}
