package auth

import (
	"context"
	"github.com/flyteorg/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flytestdlib/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var interceptor grpc.UnaryServerInterceptor
var defaultInterceptor grpc.UnaryServerInterceptor

type interceptorProvider struct {}

func (i *interceptorProvider) Register(newInterceptor grpc.UnaryServerInterceptor) {
	logger.Warnf(context.Background(), "** registered interceptor [%+v]", interceptor)
	interceptor = newInterceptor
}

func (i *interceptorProvider) RegisterDefault(newInterceptor grpc.UnaryServerInterceptor) {
	logger.Warnf(context.Background(), "** registered default interceptor [%+v]", interceptor)
	defaultInterceptor = newInterceptor
}

func (i *interceptorProvider) Get() grpc.UnaryServerInterceptor {
	logger.Warnf(context.TODO(), "**getting interceptor, interceptor [%+v] and default [%+v]", interceptor, defaultInterceptor)
	if interceptor != nil {
		return interceptor
	}
	return defaultInterceptor
}

func CustomAuthorization(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	return func() (resp interface{}, err error) {
		return GetInterceptorProvider().Get()(ctx, req, info, handler)
	}()
}

func BlanketAuthorization(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
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
