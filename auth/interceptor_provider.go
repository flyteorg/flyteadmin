package auth

import (
	"context"
	"github.com/flyteorg/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flytestdlib/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var interceptor grpc.UnaryServerInterceptor = blanketAuthorization

type interceptorProvider struct {}

func (i *interceptorProvider) Register(newInterceptor grpc.UnaryServerInterceptor) {
	logger.Warnf(context.Background(), "** registered interceptor [%+v]", interceptor)
	interceptor = newInterceptor
}

func (i *interceptorProvider) Get() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		logger.Warnf(context.Background(), "** returning interceptor [%+v]", interceptor)
		return interceptor(ctx, req, info, handler)
	}
}

func customAuthorization(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	return GetInterceptorProvider().Get()(ctx, req, info, handler)
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
	return &interceptorProvider{}
}

var authInterceptorProvider = NewInterceptorProvider()

func GetInterceptorProvider() interfaces.InterceptorProvider {
	return authInterceptorProvider
}
