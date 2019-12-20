package rpc

import (
	"context"
	"google.golang.org/grpc"
)

func AuditLogInterceptor(
	ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
	resp interface{}, err error) {

}

