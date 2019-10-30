package server

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
)

// This is a function used for CORS support. It produces a Pattern object that can be attached to the grpc-gateway
// ServeMux object. It should match any and all URLs. The pattern says, push the entire path to the stack, and then
// ignore the result.
func GetGlobPattern() runtime.Pattern {
	return runtime.MustPattern(runtime.NewPattern(1, []int{int(utilities.OpPushM), int(utilities.OpNop)},
		[]string{}, ""))
}
