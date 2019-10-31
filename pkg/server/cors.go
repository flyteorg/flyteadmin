package server

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
)

const (
	// At time of writing, the version argument is required to be 1.
	GrpcGatewayPatternVersion = 1
)

// This is a function used for CORS support. It produces a Pattern object that can be attached to the grpc-gateway
// ServeMux object. It should match any and all URLs. The two op codes say, push the entire path to the stack 'OpPushM',
// and then ignore the result, 'OpNop'.
func GetGlobPattern() runtime.Pattern {
	return runtime.MustPattern(runtime.NewPattern(GrpcGatewayPatternVersion, []int{int(utilities.OpPushM), int(utilities.OpNop)},
		[]string{}, ""))
}
