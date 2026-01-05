package interceptor

import (
	"context"
	"time"

	"starter-kit-grpc-golang/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func LoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		duration := time.Since(start)
		code := status.Code(err)

		// Log Request
		// Level: Info for success, Error for failures
		logArgs := []interface{}{
			"method", info.FullMethod,
			"code", code.String(),
			"duration", duration.String(),
		}

		if err != nil {
			logArgs = append(logArgs, "error", err.Error())
			logger.Log.Error("gRPC Request Failed", logArgs...)
		} else {
			logger.Log.Info("gRPC Request Processed", logArgs...)
		}

		return resp, err
	}
}