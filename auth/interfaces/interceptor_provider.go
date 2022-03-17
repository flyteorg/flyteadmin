package interfaces

import "google.golang.org/grpc"

// InterceptorProvider manages a singleton authorization grpc server interceptor to applied for all incoming requests.
type InterceptorProvider interface {
	// Register adds the singleton, custom authorization grpc server interceptor
	Register(interceptor grpc.UnaryServerInterceptor)
	// Get returns an authorization grpc server interceptor. If none has been registerd, default blanket authorization
	// interceptor will be provided which always grants permission.
	Get() grpc.UnaryServerInterceptor
}
