package authzserver

import (
	"context"
	"crypto/rsa"
	"github.com/flyteorg/flyteadmin/pkg/auth/config"
	"github.com/flyteorg/flyteadmin/pkg/auth/interfaces"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/ory/fosite"
	fositeOAuth2 "github.com/ory/fosite/handler/oauth2"
	"reflect"
	"testing"
)

func TestNewProvider(t *testing.T) {
	type args struct {
		ctx      context.Context
		cfg      config.AuthorizationServer
		audience string
		sm       core.SecretManager
	}
	tests := []struct {
		name    string
		args    args
		want    Provider
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewProvider(tt.args.ctx, tt.args.cfg, tt.args.audience, tt.args.sm)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProvider() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_KeySet(t *testing.T) {
	type fields struct {
		OAuth2Provider fosite.OAuth2Provider
		audience       string
		cfg            config.AuthorizationServer
		publicKey      []rsa.PublicKey
		keySet         jwk.Set
	}
	tests := []struct {
		name   string
		fields fields
		want   jwk.Set
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Provider{
				OAuth2Provider: tt.fields.OAuth2Provider,
				audience:       tt.fields.audience,
				cfg:            tt.fields.cfg,
				publicKey:      tt.fields.publicKey,
				keySet:         tt.fields.keySet,
			}
			if got := p.KeySet(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeySet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_NewJWTSessionToken(t *testing.T) {
	type fields struct {
		OAuth2Provider fosite.OAuth2Provider
		audience       string
		cfg            config.AuthorizationServer
		publicKey      []rsa.PublicKey
		keySet         jwk.Set
	}
	type args struct {
		subject        string
		userInfoClaims interface{}
		appID          string
		issuer         string
		audience       string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *fositeOAuth2.JWTSession
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Provider{
				OAuth2Provider: tt.fields.OAuth2Provider,
				audience:       tt.fields.audience,
				cfg:            tt.fields.cfg,
				publicKey:      tt.fields.publicKey,
				keySet:         tt.fields.keySet,
			}
			if got := p.NewJWTSessionToken(tt.args.subject, tt.args.userInfoClaims, tt.args.appID, tt.args.issuer, tt.args.audience); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewJWTSessionToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_PublicKeys(t *testing.T) {
	type fields struct {
		OAuth2Provider fosite.OAuth2Provider
		audience       string
		cfg            config.AuthorizationServer
		publicKey      []rsa.PublicKey
		keySet         jwk.Set
	}
	tests := []struct {
		name   string
		fields fields
		want   []rsa.PublicKey
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Provider{
				OAuth2Provider: tt.fields.OAuth2Provider,
				audience:       tt.fields.audience,
				cfg:            tt.fields.cfg,
				publicKey:      tt.fields.publicKey,
				keySet:         tt.fields.keySet,
			}
			if got := p.PublicKeys(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PublicKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_ValidateAccessToken(t *testing.T) {
	type fields struct {
		OAuth2Provider fosite.OAuth2Provider
		audience       string
		cfg            config.AuthorizationServer
		publicKey      []rsa.PublicKey
		keySet         jwk.Set
	}
	type args struct {
		ctx      context.Context
		tokenStr string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interfaces.IdentityContext
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Provider{
				OAuth2Provider: tt.fields.OAuth2Provider,
				audience:       tt.fields.audience,
				cfg:            tt.fields.cfg,
				publicKey:      tt.fields.publicKey,
				keySet:         tt.fields.keySet,
			}
			got, err := p.ValidateAccessToken(tt.args.ctx, tt.args.tokenStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateAccessToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}
