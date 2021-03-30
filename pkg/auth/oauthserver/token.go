package oauthserver

import (
	"log"
	"net/http"

	"github.com/flyteorg/flyteadmin/pkg/auth/interfaces"
)

func getTokenEndpointHandler(authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		tokenEndpoint(authCtx, writer, request)
	}
}

func tokenEndpoint(authCtx interfaces.AuthenticationContext, rw http.ResponseWriter, req *http.Request) {
	// This context will be passed to all methods.
	ctx := req.Context()

	oauth2Provider := authCtx.OAuth2Provider()

	// Create an empty session object which will be passed to the request handlers
	mySessionData := oauth2Provider.NewJWTSessionToken("", nil, "", "", "")

	// This will create an access request object and iterate through the registered TokenEndpointHandlers to validate the request.
	accessRequest, err := oauth2Provider.NewAccessRequest(ctx, req, mySessionData)

	// Catch any errors, e.g.:
	// * unknown client
	// * invalid redirect
	// * ...
	if err != nil {
		log.Printf("Error occurred in NewAccessRequest: %+v", err)
		oauth2Provider.WriteAccessError(rw, accessRequest, err)
		return
	}

	// If this is a client_credentials grant, grant all requested scopes
	// NewAccessRequest validated that all requested scopes the client is allowed to perform
	// based on configured scope matching strategy.
	// If this is authorization_code, we should have consented the user for the requested scopes, so grant those too
	if accessRequest.GetGrantTypes().HasOneOf("client_credentials", "authorization_code") {
		for _, scope := range accessRequest.GetRequestedScopes() {
			accessRequest.GrantScope(scope)
		}

		issuer := getIssuer(authCtx.Options(), req)
		accessRequest.GrantAudience(issuer)
	}

	// Next we create a response for the access request. Again, we iterate through the TokenEndpointHandlers
	// and aggregate the result in response.
	response, err := oauth2Provider.NewAccessResponse(ctx, accessRequest)
	if err != nil {
		log.Printf("Error occurred in NewAccessResponse: %+v", err)
		oauth2Provider.WriteAccessError(rw, accessRequest, err)
		return
	}

	// All done, send the response.
	oauth2Provider.WriteAccessResponse(rw, accessRequest, response)

	// The client now has a valid access token
}
