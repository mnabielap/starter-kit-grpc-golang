package interceptor

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetUserIDFromContext extracts the user ID injected by AuthInterceptor
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", status.Error(codes.Unauthenticated, "user id not found in context")
	}
	return userID, nil
}

// AuthorizeAdmin ensures the user has 'admin' role
func AuthorizeAdmin(ctx context.Context) error {
	role, ok := ctx.Value(RoleKey).(string)
	if !ok {
		return status.Error(codes.Unauthenticated, "role not found in context")
	}

	if role != "admin" {
		return status.Error(codes.PermissionDenied, "forbidden: admins only")
	}
	return nil
}

// AuthorizeAdminOrSelf ensures user is admin OR matching the target ID
func AuthorizeAdminOrSelf(ctx context.Context, targetID string) error {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 1. Check if Self
	if userID == targetID {
		return nil
	}

	// 2. Check if Admin
	return AuthorizeAdmin(ctx)
}