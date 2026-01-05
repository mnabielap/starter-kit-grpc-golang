package interceptor

import (
	"context"
	"strings"

	"starter-kit-grpc-golang/config"
	"starter-kit-grpc-golang/pkg/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	RoleKey   contextKey = "role"
)

// AuthInterceptor creates a unary server interceptor for JWT validation
func AuthInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 1. Define Public Methods (Skip Auth)
		// Format: /<package>.<Service>/<Method>
		publicMethods := map[string]bool{
			"/v1.AuthService/Login":                 true,
			"/v1.AuthService/Register":              true,
			"/v1.AuthService/RefreshToken":          true,
			"/v1.AuthService/ForgotPassword":        true,
			"/v1.AuthService/ResetPassword":         true,
			"/v1.AuthService/VerifyEmail":           true,
			"/v1.HealthService/HealthCheck":         true,
		}

		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// 2. Extract Token from Metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		authHeader := md["authorization"]
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}

		// Expected format: "Bearer <token>"
		tokenParts := strings.Split(authHeader[0], " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return nil, status.Error(codes.Unauthenticated, "invalid token format")
		}

		tokenString := tokenParts[1]

		// 3. Validate Token
		claims, err := utils.ValidateToken(tokenString, cfg.JWT.Secret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		if claims.Type != "access" {
			return nil, status.Error(codes.Unauthenticated, "invalid token type")
		}

		// 4. Inject Claims into Context
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)

		return handler(ctx, req)
	}
}