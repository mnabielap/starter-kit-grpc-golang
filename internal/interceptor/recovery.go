package interceptor

import (
	"context"
	"runtime/debug"

	"starter-kit-grpc-golang/pkg/logger"

	"google.golang.org/grpc"
)

func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Log.Error("Panic recovered", 
					"error", r, 
					"stack", string(debug.Stack()),
					"method", info.FullMethod,
				)
			}
		}()

		return handler(ctx, req)
	}
}