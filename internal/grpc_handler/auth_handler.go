package grpc_handler

import (
	"context"
	"time"

	pb "starter-kit-grpc-golang/api/gen/v1"
	"starter-kit-grpc-golang/internal/interceptor"
	"starter-kit-grpc-golang/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	service service.AuthService
}

func NewAuthHandler(s service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	user, accessToken, refreshToken, accessExp, refreshExp, err := h.service.Register(req.Name, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.AuthResponse{
		User:   convertUserToProto(user),
		Tokens: createTokenPair(accessToken, refreshToken, accessExp, refreshExp),
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	user, accessToken, refreshToken, accessExp, refreshExp, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &pb.AuthResponse{
		User:   convertUserToProto(user),
		Tokens: createTokenPair(accessToken, refreshToken, accessExp, refreshExp),
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := h.service.Logout(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.NotFound, "invalid token")
	}
	return &pb.LogoutResponse{Success: true}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.TokenPair, error) {
	accessToken, refreshToken, accessExp, refreshExp, err := h.service.RefreshAuth(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return createTokenPair(accessToken, refreshToken, accessExp, refreshExp), nil
}

func (h *AuthHandler) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.SuccessResponse, error) {
	err := h.service.ForgotPassword(req.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to process request")
	}
	// Return success regardless of whether email exists (Security)
	return &pb.SuccessResponse{Message: "If email exists, reset instructions have been sent"}, nil
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.SuccessResponse, error) {
	err := h.service.ResetPassword(req.Token, req.Password)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.SuccessResponse{Message: "Password reset successfully"}, nil
}

func (h *AuthHandler) SendVerificationEmail(ctx context.Context, req *pb.Empty) (*pb.SuccessResponse, error) {
	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.SendVerificationEmail(userID); err != nil {
		return nil, status.Error(codes.Internal, "failed to send email")
	}
	return &pb.SuccessResponse{Message: "Verification email sent"}, nil
}

func (h *AuthHandler) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.SuccessResponse, error) {
	if err := h.service.VerifyEmail(req.Token); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.SuccessResponse{Message: "Email verified successfully"}, nil
}

// Helper
func createTokenPair(access, refresh string, accessExp, refreshExp time.Time) *pb.TokenPair {
	return &pb.TokenPair{
		Access: &pb.TokenPair_TokenDetail{
			Token:   access,
			Expires: accessExp.Format(time.RFC3339),
		},
		Refresh: &pb.TokenPair_TokenDetail{
			Token:   refresh,
			Expires: refreshExp.Format(time.RFC3339),
		},
	}
}