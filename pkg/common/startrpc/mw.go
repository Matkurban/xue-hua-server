package startrpc

import (
	"context"

	"google.golang.org/grpc"
	"server/v3/pkg/authverify"
)

func grpcServerIMAdminUserID(imAdminUserID []string) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx = authverify.WithIMAdminUserIDs(ctx, imAdminUserID)
		return handler(ctx, req)
	})
}
