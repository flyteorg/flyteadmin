// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	http "net/http"

	config "github.com/flyteorg/flyteadmin/auth/config"

	interfaces "github.com/flyteorg/flyteadmin/auth/interfaces"

	mock "github.com/stretchr/testify/mock"

	oauth2 "golang.org/x/oauth2"

	oidc "github.com/coreos/go-oidc/v3/oidc"

	service "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	url "net/url"
)

// AuthenticationContext is an autogenerated mock type for the AuthenticationContext type
type AuthenticationContext struct {
	mock.Mock
}

type AuthenticationContext_AuthMetadataService struct {
	*mock.Call
}

func (_m AuthenticationContext_AuthMetadataService) Return(_a0 service.AuthMetadataServiceServer) *AuthenticationContext_AuthMetadataService {
	return &AuthenticationContext_AuthMetadataService{Call: _m.Call.Return(_a0)}
}

func (_m *AuthenticationContext) OnAuthMetadataService() *AuthenticationContext_AuthMetadataService {
	c_call := _m.On("AuthMetadataService")
	return &AuthenticationContext_AuthMetadataService{Call: c_call}
}

func (_m *AuthenticationContext) OnAuthMetadataServiceMatch(matchers ...interface{}) *AuthenticationContext_AuthMetadataService {
	c_call := _m.On("AuthMetadataService", matchers...)
	return &AuthenticationContext_AuthMetadataService{Call: c_call}
}

// AuthMetadataService provides a mock function with given fields:
func (_m *AuthenticationContext) AuthMetadataService() service.AuthMetadataServiceServer {
	ret := _m.Called()

	var r0 service.AuthMetadataServiceServer
	if rf, ok := ret.Get(0).(func() service.AuthMetadataServiceServer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(service.AuthMetadataServiceServer)
		}
	}

	return r0
}

type AuthenticationContext_CookieManager struct {
	*mock.Call
}

func (_m AuthenticationContext_CookieManager) Return(_a0 interfaces.CookieHandler) *AuthenticationContext_CookieManager {
	return &AuthenticationContext_CookieManager{Call: _m.Call.Return(_a0)}
}

func (_m *AuthenticationContext) OnCookieManager() *AuthenticationContext_CookieManager {
	c_call := _m.On("CookieManager")
	return &AuthenticationContext_CookieManager{Call: c_call}
}

func (_m *AuthenticationContext) OnCookieManagerMatch(matchers ...interface{}) *AuthenticationContext_CookieManager {
	c_call := _m.On("CookieManager", matchers...)
	return &AuthenticationContext_CookieManager{Call: c_call}
}

// CookieManager provides a mock function with given fields:
func (_m *AuthenticationContext) CookieManager() interfaces.CookieHandler {
	ret := _m.Called()

	var r0 interfaces.CookieHandler
	if rf, ok := ret.Get(0).(func() interfaces.CookieHandler); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.CookieHandler)
		}
	}

	return r0
}

type AuthenticationContext_GetHTTPClient struct {
	*mock.Call
}

func (_m AuthenticationContext_GetHTTPClient) Return(_a0 *http.Client) *AuthenticationContext_GetHTTPClient {
	return &AuthenticationContext_GetHTTPClient{Call: _m.Call.Return(_a0)}
}

func (_m *AuthenticationContext) OnGetHTTPClient() *AuthenticationContext_GetHTTPClient {
	c_call := _m.On("GetHTTPClient")
	return &AuthenticationContext_GetHTTPClient{Call: c_call}
}

func (_m *AuthenticationContext) OnGetHTTPClientMatch(matchers ...interface{}) *AuthenticationContext_GetHTTPClient {
	c_call := _m.On("GetHTTPClient", matchers...)
	return &AuthenticationContext_GetHTTPClient{Call: c_call}
}

// GetHTTPClient provides a mock function with given fields:
func (_m *AuthenticationContext) GetHTTPClient() *http.Client {
	ret := _m.Called()

	var r0 *http.Client
	if rf, ok := ret.Get(0).(func() *http.Client); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Client)
		}
	}

	return r0
}

type AuthenticationContext_GetOAuth2MetadataURL struct {
	*mock.Call
}

func (_m AuthenticationContext_GetOAuth2MetadataURL) Return(_a0 *url.URL) *AuthenticationContext_GetOAuth2MetadataURL {
	return &AuthenticationContext_GetOAuth2MetadataURL{Call: _m.Call.Return(_a0)}
}

func (_m *AuthenticationContext) OnGetOAuth2MetadataURL() *AuthenticationContext_GetOAuth2MetadataURL {
	c_call := _m.On("GetOAuth2MetadataURL")
	return &AuthenticationContext_GetOAuth2MetadataURL{Call: c_call}
}

func (_m *AuthenticationContext) OnGetOAuth2MetadataURLMatch(matchers ...interface{}) *AuthenticationContext_GetOAuth2MetadataURL {
	c_call := _m.On("GetOAuth2MetadataURL", matchers...)
	return &AuthenticationContext_GetOAuth2MetadataURL{Call: c_call}
}

// GetOAuth2MetadataURL provides a mock function with given fields:
func (_m *AuthenticationContext) GetOAuth2MetadataURL() *url.URL {
	ret := _m.Called()

	var r0 *url.URL
	if rf, ok := ret.Get(0).(func() *url.URL); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*url.URL)
		}
	}

	return r0
}

type AuthenticationContext_GetOIdCMetadataURL struct {
	*mock.Call
}

func (_m AuthenticationContext_GetOIdCMetadataURL) Return(_a0 *url.URL) *AuthenticationContext_GetOIdCMetadataURL {
	return &AuthenticationContext_GetOIdCMetadataURL{Call: _m.Call.Return(_a0)}
}

func (_m *AuthenticationContext) OnGetOIdCMetadataURL() *AuthenticationContext_GetOIdCMetadataURL {
	c_call := _m.On("GetOIdCMetadataURL")
	return &AuthenticationContext_GetOIdCMetadataURL{Call: c_call}
}

func (_m *AuthenticationContext) OnGetOIdCMetadataURLMatch(matchers ...interface{}) *AuthenticationContext_GetOIdCMetadataURL {
	c_call := _m.On("GetOIdCMetadataURL", matchers...)
	return &AuthenticationContext_GetOIdCMetadataURL{Call: c_call}
}

// GetOIdCMetadataURL provides a mock function with given fields:
func (_m *AuthenticationContext) GetOIdCMetadataURL() *url.URL {
	ret := _m.Called()

	var r0 *url.URL
	if rf, ok := ret.Get(0).(func() *url.URL); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*url.URL)
		}
	}

	return r0
}

type AuthenticationContext_IdentityService struct {
	*mock.Call
}

func (_m AuthenticationContext_IdentityService) Return(_a0 service.IdentityServiceServer) *AuthenticationContext_IdentityService {
	return &AuthenticationContext_IdentityService{Call: _m.Call.Return(_a0)}
}

func (_m *AuthenticationContext) OnIdentityService() *AuthenticationContext_IdentityService {
	c_call := _m.On("IdentityService")
	return &AuthenticationContext_IdentityService{Call: c_call}
}

func (_m *AuthenticationContext) OnIdentityServiceMatch(matchers ...interface{}) *AuthenticationContext_IdentityService {
	c_call := _m.On("IdentityService", matchers...)
	return &AuthenticationContext_IdentityService{Call: c_call}
}

// IdentityService provides a mock function with given fields:
func (_m *AuthenticationContext) IdentityService() service.IdentityServiceServer {
	ret := _m.Called()

	var r0 service.IdentityServiceServer
	if rf, ok := ret.Get(0).(func() service.IdentityServiceServer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(service.IdentityServiceServer)
		}
	}

	return r0
}

type AuthenticationContext_OAuth2ClientConfig struct {
	*mock.Call
}

func (_m AuthenticationContext_OAuth2ClientConfig) Return(_a0 *oauth2.Config) *AuthenticationContext_OAuth2ClientConfig {
	return &AuthenticationContext_OAuth2ClientConfig{Call: _m.Call.Return(_a0)}
}

func (_m *AuthenticationContext) OnOAuth2ClientConfig(requestURL *url.URL) *AuthenticationContext_OAuth2ClientConfig {
	c_call := _m.On("OAuth2ClientConfig", requestURL)
	return &AuthenticationContext_OAuth2ClientConfig{Call: c_call}
}

func (_m *AuthenticationContext) OnOAuth2ClientConfigMatch(matchers ...interface{}) *AuthenticationContext_OAuth2ClientConfig {
	c_call := _m.On("OAuth2ClientConfig", matchers...)
	return &AuthenticationContext_OAuth2ClientConfig{Call: c_call}
}

// OAuth2ClientConfig provides a mock function with given fields: requestURL
func (_m *AuthenticationContext) OAuth2ClientConfig(requestURL *url.URL) *oauth2.Config {
	ret := _m.Called(requestURL)

	var r0 *oauth2.Config
	if rf, ok := ret.Get(0).(func(*url.URL) *oauth2.Config); ok {
		r0 = rf(requestURL)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth2.Config)
		}
	}

	return r0
}

type AuthenticationContext_OAuth2Provider struct {
	*mock.Call
}

func (_m AuthenticationContext_OAuth2Provider) Return(_a0 interfaces.OAuth2Provider) *AuthenticationContext_OAuth2Provider {
	return &AuthenticationContext_OAuth2Provider{Call: _m.Call.Return(_a0)}
}

func (_m *AuthenticationContext) OnOAuth2Provider() *AuthenticationContext_OAuth2Provider {
	c_call := _m.On("OAuth2Provider")
	return &AuthenticationContext_OAuth2Provider{Call: c_call}
}

func (_m *AuthenticationContext) OnOAuth2ProviderMatch(matchers ...interface{}) *AuthenticationContext_OAuth2Provider {
	c_call := _m.On("OAuth2Provider", matchers...)
	return &AuthenticationContext_OAuth2Provider{Call: c_call}
}

// OAuth2Provider provides a mock function with given fields:
func (_m *AuthenticationContext) OAuth2Provider() interfaces.OAuth2Provider {
	ret := _m.Called()

	var r0 interfaces.OAuth2Provider
	if rf, ok := ret.Get(0).(func() interfaces.OAuth2Provider); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.OAuth2Provider)
		}
	}

	return r0
}

type AuthenticationContext_OAuth2ResourceServer struct {
	*mock.Call
}

func (_m AuthenticationContext_OAuth2ResourceServer) Return(_a0 interfaces.OAuth2ResourceServer) *AuthenticationContext_OAuth2ResourceServer {
	return &AuthenticationContext_OAuth2ResourceServer{Call: _m.Call.Return(_a0)}
}

func (_m *AuthenticationContext) OnOAuth2ResourceServer() *AuthenticationContext_OAuth2ResourceServer {
	c_call := _m.On("OAuth2ResourceServer")
	return &AuthenticationContext_OAuth2ResourceServer{Call: c_call}
}

func (_m *AuthenticationContext) OnOAuth2ResourceServerMatch(matchers ...interface{}) *AuthenticationContext_OAuth2ResourceServer {
	c_call := _m.On("OAuth2ResourceServer", matchers...)
	return &AuthenticationContext_OAuth2ResourceServer{Call: c_call}
}

// OAuth2ResourceServer provides a mock function with given fields:
func (_m *AuthenticationContext) OAuth2ResourceServer() interfaces.OAuth2ResourceServer {
	ret := _m.Called()

	var r0 interfaces.OAuth2ResourceServer
	if rf, ok := ret.Get(0).(func() interfaces.OAuth2ResourceServer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.OAuth2ResourceServer)
		}
	}

	return r0
}

type AuthenticationContext_OidcProvider struct {
	*mock.Call
}

func (_m AuthenticationContext_OidcProvider) Return(_a0 *oidc.Provider) *AuthenticationContext_OidcProvider {
	return &AuthenticationContext_OidcProvider{Call: _m.Call.Return(_a0)}
}

func (_m *AuthenticationContext) OnOidcProvider() *AuthenticationContext_OidcProvider {
	c_call := _m.On("OidcProvider")
	return &AuthenticationContext_OidcProvider{Call: c_call}
}

func (_m *AuthenticationContext) OnOidcProviderMatch(matchers ...interface{}) *AuthenticationContext_OidcProvider {
	c_call := _m.On("OidcProvider", matchers...)
	return &AuthenticationContext_OidcProvider{Call: c_call}
}

// OidcProvider provides a mock function with given fields:
func (_m *AuthenticationContext) OidcProvider() *oidc.Provider {
	ret := _m.Called()

	var r0 *oidc.Provider
	if rf, ok := ret.Get(0).(func() *oidc.Provider); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oidc.Provider)
		}
	}

	return r0
}

type AuthenticationContext_Options struct {
	*mock.Call
}

func (_m AuthenticationContext_Options) Return(_a0 *config.Config) *AuthenticationContext_Options {
	return &AuthenticationContext_Options{Call: _m.Call.Return(_a0)}
}

func (_m *AuthenticationContext) OnOptions() *AuthenticationContext_Options {
	c_call := _m.On("Options")
	return &AuthenticationContext_Options{Call: c_call}
}

func (_m *AuthenticationContext) OnOptionsMatch(matchers ...interface{}) *AuthenticationContext_Options {
	c_call := _m.On("Options", matchers...)
	return &AuthenticationContext_Options{Call: c_call}
}

// Options provides a mock function with given fields:
func (_m *AuthenticationContext) Options() *config.Config {
	ret := _m.Called()

	var r0 *config.Config
	if rf, ok := ret.Get(0).(func() *config.Config); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*config.Config)
		}
	}

	return r0
}
