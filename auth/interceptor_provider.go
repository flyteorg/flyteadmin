package auth

import (
	"context"
	"github.com/flyteorg/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flytestdlib/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type interceptorProvider struct {
	interceptor grpc.UnaryServerInterceptor
}

func (i *interceptorProvider) Register(newInterceptor grpc.UnaryServerInterceptor) {
	logger.Warnf(context.Background(), "** registered interceptor [%+v]", i.interceptor)
	i.interceptor = newInterceptor
}

func (i *interceptorProvider) Get() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		logger.Warnf(context.Background(), "** returning interceptor [%+v]", i.interceptor)
		return i.interceptor(ctx, req, info, handler)
	}
}

func CustomAuthorization(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	logger.Warnf(ctx, "** getting intercept provider [%+v]", GetInterceptorProvider().Get())
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
	ip := interceptorProvider{}
	ip.Register(blanketAuthorization)
	return &ip
}

var authInterceptorProvider = NewInterceptorProvider()

func GetInterceptorProvider() interfaces.InterceptorProvider {
	return authInterceptorProvider
}
