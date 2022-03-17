package auth

import (
	"context"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/auth/interfaces"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type interceptorProvider struct {
	interceptor grpc.UnaryServerInterceptor
}

func (i *interceptorProvider) Register(interceptor grpc.UnaryServerInterceptor) {
	logger.Warnf(context.Background(), "** registered interceptor [%+v]", interceptor)
	i.interceptor = interceptor
}

func (i *interceptorProvider) Get() grpc.UnaryServerInterceptor {
	logger.Warnf(context.Background(), "** returning interceptor [%+v]", i.interceptor)
	return i.interceptor
}

func blanketAuthorization(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
	resp interface{}, err error) {
	logger.Warnf(ctx, "** running blanket authorization")
	identityContext := IdentityContextFromContext(ctx)
	if identityContext.IsEmpty() {
		return handler(ctx, req)
	}

	if !identityContext.Scopes().Has(ScopeAll) {
		return nil, status.Errorf(codes.Unauthenticated, "authenticated user doesn't have required scope")
	}

	return handler(ctx, req)
}

func NewInterceptorProvider() interfaces.InterceptorProvider {
	return &interceptorProvider{
		interceptor: blanketAuthorization,
	}
}

var authInterceptorProvider = NewInterceptorProvider()

func GetInterceptorProvider() interfaces.InterceptorProvider {
	return authInterceptorProvider
}
